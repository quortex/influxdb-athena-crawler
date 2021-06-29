package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const timeLayout = "2006-01-02T15:04:05.000Z"

var (
	fRegion       string
	fBucket       string
	fPrefix       string
	fCleanObjects bool
	fTimeout      time.Duration
	fInfluxServer string
	fInfluxToken  string
	fInfluxOrg    string
	fInfluxBucket string
)

func main() {
	start := time.Now()

	flag.StringVar(&fRegion, "region", "", "The AWS region.")
	flag.StringVar(&fBucket, "bucket", "", "The bucket to watch.")
	flag.StringVar(&fPrefix, "prefix", "", "The bucket prefix.")
	flag.BoolVar(&fCleanObjects, "clean-objects", false, "Whether to delete S3 objects after processing them.")
	flag.StringVar(&fInfluxServer, "influx-server", "", "The InfluxDB server address.")
	flag.StringVar(&fInfluxToken, "influx-token", "", "The InfluxDB token.")
	flag.StringVar(&fInfluxOrg, "influx-org", "", "The InfluxDB org to write to.")
	flag.StringVar(&fInfluxBucket, "influx-bucket", "", "The InfluxDB bucket write to.")
	flag.DurationVar(&fTimeout, "timeout", 2*time.Minute, "The program timeout.")
	flag.Parse()

	// Initialize context with defined tiemout
	ctx, cancel := context.WithTimeout(context.Background(), fTimeout)
	defer cancel()
	go func() {
		<-ctx.Done()
		log.Fatal().Msg("Timeout reached !")
	}()

	// Initialize logger
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Second

	// Check flags conformity
	if fRegion == "" {
		log.Fatal().Msg("Required flag: region")
	}
	if fBucket == "" {
		log.Fatal().Msg("Required flag: bucket")
	}
	if fInfluxServer == "" {
		log.Fatal().Msg("Required flag: influx-server")
	}
	if fInfluxToken == "" {
		log.Fatal().Msg("Required flag: influx-token")
	}
	if fInfluxOrg == "" {
		log.Fatal().Msg("Required flag: influx-org")
	}
	if fInfluxBucket == "" {
		log.Fatal().Msg("Required flag: influx-bucket")
	}

	// Init AWS s3 client
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(fRegion)},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to init session")
	}
	s3Cli := s3.New(sess)

	// List objects matching bucket / prefix
	res, err := s3Cli.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(fBucket),
		Prefix: aws.String(fPrefix),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to list objects")
	}
	if len(res.Contents) == 0 {
		log.Info().Msg("No objects matching bucket / prefix, processing done !")
	}

	// Make waitgroup and channels to process objects
	// tasks asynchronously
	var wg sync.WaitGroup
	cDone := make(chan bool)
	cErr := make(chan error)

	// Process each s3 object
	dwn := s3manager.NewDownloader(sess)
	cli := influxdb2.NewClient(fInfluxServer, fInfluxToken)
	defer cli.Close()
	api := cli.WriteAPIBlocking(fInfluxOrg, fInfluxBucket)
	for _, item := range res.Contents {
		o := *item
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := processObject(ctx, s3Cli, dwn, api, o); err != nil {
				cErr <- err
			}
		}()
	}

	// Wait until WaitGroup is done
	go func() {
		wg.Wait()
		close(cDone)
	}()

	// Wait until either WaitGroup is done or an error is received through the channel
	select {
	case <-cDone:
		log.Info().
			Dur("elapsed", time.Since(start)).
			Msg("Processing done with succes !")
		break
	case err := <-cErr:
		close(cErr)
		log.Error().Err(err).
			Dur("elapsed", time.Since(start)).
			Msg("Processing error")
	}
}

func processObject(
	ctx context.Context,
	s3Cli *s3.S3,
	s3Dwn *s3manager.Downloader,
	writeAPI api.WriteAPIBlocking,
	o s3.Object,
) error {
	log.Info().
		Str("object", aws.StringValue(o.Key)).
		Time("last modified", aws.TimeValue(o.LastModified)).
		Int64("size", aws.Int64Value(o.Size)).
		Msg("Processing s3 object")

	// Download object
	buf := aws.NewWriteAtBuffer([]byte{})
	_, err := s3Dwn.DownloadWithContext(ctx, buf, &s3.GetObjectInput{
		Bucket: aws.String(fBucket),
		Key:    o.Key,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("buckey", fBucket).
			Str("object", aws.StringValue(o.Key)).
			Msg("Failed to download object")
		return err
	}

	// Parse CSV to a map[string]interface{} slice
	res, err := parseCSV(string(buf.Bytes()))
	if err != nil {
		log.Error().
			Err(err).
			Str("object", aws.StringValue(o.Key)).
			Msg("Failed to parse CSV")
		return err
	}

	// Write records to InfluxDB
	if err = writeRecordsToInfluxDB(ctx, writeAPI, res); err != nil {
		log.Error().
			Err(err).
			Str("object", aws.StringValue(o.Key)).
			Msg("Failed to write records to InfluxDB")
		return err
	}

	// Delete object
	if fCleanObjects {
		_, err = s3Cli.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(fBucket),
			Key:    o.Key,
		})
		if err != nil {
			log.Error().
				Err(err).
				Str("bucket", fBucket).
				Str("object", aws.StringValue(o.Key)).
				Msg("Unable to delete object")
			return err
		}
	}

	return nil
}

func parseCSV(strCSV string) ([]map[string]interface{}, error) {
	// Read CSV object
	var header []string
	res := []map[string]interface{}{}

	reader := csv.NewReader(strings.NewReader(strCSV))
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// First line contains header fields
		if header == nil {
			header = line
			continue
		}

		// Other lines contains rows
		row := make(map[string]interface{}, len(header))
		for i, e := range line {
			row[header[i]] = e
		}
		res = append(res, row)
	}

	return res, nil
}

func writeRecordsToInfluxDB(ctx context.Context, writeAPI api.WriteAPIBlocking, rows []map[string]interface{}) error {
	// Convert csv rows to InfluxDB points
	points, err := toPoints(rows)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to convert CSV rows to points")
		return err
	}

	// Write points to InfluxDB
	err = writeAPI.WritePoint(context.Background(), points...)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to write points to InfluxDB")
		return err
	}

	return nil
}

func toPoints(rows []map[string]interface{}) ([]*write.Point, error) {
	res := make([]*write.Point, len(rows))
	for i, e := range rows {
		p, err := toPoint(e)
		if err != nil {
			return nil, err
		}
		res[i] = p
	}
	return res, nil
}

func toPoint(row map[string]interface{}) (*write.Point, error) {
	t, err := time.Parse(timeLayout, fmt.Sprintf("%sT%s.000Z", row["date"], row["time"]))
	if err != nil {
		return nil, err
	}

	aud, err := strconv.Atoi(fmt.Sprintf("%v", row["audience"]))
	if err != nil {
		return nil, err
	}

	return influxdb2.NewPointWithMeasurement("audience").
		SetTime(t).
		AddTag("publishing_point", fmt.Sprintf("%v", row["publishing_point"])).
		AddField("audience", aud), nil
}
