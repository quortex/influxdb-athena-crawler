package main

import (
	"context"
	"encoding/csv"
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
	"github.com/quortex/influxdb-athena-crawler/pkg/flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var opts flags.Options

func main() {
	start := time.Now()

	// Parse flags
	if err := flags.Parse(&opts); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse flags")
	}

	// Initialize logger
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.DurationFieldUnit = time.Second

	// Initialize context with defined tiemout
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	go func() {
		<-ctx.Done()
		log.Fatal().Msg("Timeout reached !")
	}()

	// Init AWS s3 client
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(opts.Region)},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to init session")
	}
	s3Cli := s3.New(sess)

	// List objects matching bucket / prefix
	res, err := s3Cli.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(opts.Bucket),
		Prefix: aws.String(opts.Prefix),
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
	cli := influxdb2.NewClient(opts.InfluxServer, opts.InfluxToken)
	defer cli.Close()
	api := cli.WriteAPIBlocking(opts.InfluxOrg, opts.InfluxBucket)
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
		Bucket: aws.String(opts.Bucket),
		Key:    o.Key,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("buckey", opts.Bucket).
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
	if opts.CleanObjects {
		_, err = s3Cli.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(opts.Bucket),
			Key:    o.Key,
		})
		if err != nil {
			log.Error().
				Err(err).
				Str("bucket", opts.Bucket).
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
	t, err := time.Parse(opts.TimestampLayout, fmt.Sprintf("%v", row[opts.TimestampRow]))
	if err != nil {
		return nil, err
	}

	point := influxdb2.NewPointWithMeasurement("audience").SetTime(t)
	for _, e := range opts.Tags {
		point = point.AddTag(e.Tag, fmt.Sprintf("%v", row[e.Row]))
	}

	for _, e := range opts.Fields {
		var fieldVal interface{}
		strField := fmt.Sprintf("%v", row[e.Row])
		switch e.FieldType {
		case flags.FieldTypeBool:
			fieldVal, err = strconv.ParseBool(strField)
		case flags.FieldTypeFloat:
			fieldVal, err = strconv.ParseFloat(strField, 64)
		case flags.FieldTypeInteger:
			fieldVal, err = strconv.Atoi(strField)
		case flags.FieldTypeString:
			fieldVal = strField
		}
		if err != nil {
			return nil, err
		}
		point = point.AddField(e.Field, fieldVal)
	}

	return point, nil
}
