package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/quortex/influxdb-athena-crawler/pkg/csv"
	"github.com/quortex/influxdb-athena-crawler/pkg/flags"
	"github.com/quortex/influxdb-athena-crawler/pkg/influxdb"
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

	// Initialize context with defined timeout
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	go func() {
		<-ctx.Done()
		log.Fatal().Msg("Timeout reached !")
	}()

	// Init AWS s3 client
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(opts.Region))
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("unable to load SDK config")
	}
	s3Cli := s3.NewFromConfig(cfg)

	// List objects matching bucket / prefix
	res, err := s3Cli.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(opts.Bucket),
		Prefix: aws.String(opts.Prefix),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to list objects")
	}

	filtered_res := filterBucketContent(*res, opts.Suffix)

	if len(filtered_res.Contents) == 0 {
		log.Info().Msg("No objects matching bucket / prefix, processing done !")
	}

	dwn := manager.NewDownloader(s3Cli)
	influxWriter := influxdb.NewWriters(
		opts.InfluxServers,
		opts.InfluxToken,
		opts.InfluxOrg,
		opts.InfluxBucket,
		opts.Measurement,
		opts.TimestampLayout,
		opts.TimestampRow,
		opts.Tags,
		opts.Fields,
	)
	defer influxWriter.Close()

	// Make waitgroup and channels to process objects
	// tasks asynchronously
	var wg sync.WaitGroup
	cDone := make(chan bool)
	wg.Add(len(filtered_res.Contents))
	go func() {
		// Process each s3 object
		for _, item := range filtered_res.Contents {
			o := item
			go func() {
				defer wg.Done()
				if err := processObject(ctx, s3Cli, dwn, influxWriter, o); err != nil {
					log.Error().Err(err).
						Msg("Processing error")
				}
			}()
		}

		wg.Wait()
		close(cDone)
	}()

	<-cDone
	log.Info().
		Dur("elapsed", time.Since(start)).
		Msg("Processing ended !")
}

func filterBucketContent(elems s3.ListObjectsOutput, suffix string) (ret s3.ListObjectsOutput) {
	if len(suffix) == 0 {
		return elems
	}
	for _, s := range elems.Contents {
		if strings.HasSuffix(*s.Key, suffix) {
			ret.Contents = append(ret.Contents, s)
		}
	}
	return ret
}

func processObject(
	ctx context.Context,
	s3Cli *s3.Client,
	s3Dwn *manager.Downloader,
	influxWriter influxdb.Writer,
	o types.Object,
) error {
	log.Info().
		Str("object", aws.ToString(o.Key)).
		Time("last modified", aws.ToTime(o.LastModified)).
		Int64("size", o.Size).
		Msg("Processing s3 object")

	// Download object
	buf := manager.NewWriteAtBuffer([]byte{})
	_, err := s3Dwn.Download(ctx, buf, &s3.GetObjectInput{
		Bucket: aws.String(opts.Bucket),
		Key:    o.Key,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("bucket", opts.Bucket).
			Str("object", aws.ToString(o.Key)).
			Msg("Failed to download object")
		return err
	}

	// Parse CSV to a map[string]interface{} slice
	res, err := csv.ParseString(string(buf.Bytes()))
	if err != nil {
		log.Error().
			Err(err).
			Str("object", aws.ToString(o.Key)).
			Msg("Failed to parse CSV")
		return err
	}

	// Write records to InfluxDB
	if err = influxWriter.WriteRecords(ctx, res); err != nil {
		log.Error().
			Err(err).
			Str("object", aws.ToString(o.Key)).
			Msg("Failed to write records")
		return err
	}

	// Delete object
	if opts.CleanObjects {
		_, err = s3Cli.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(opts.Bucket),
			Key:    o.Key,
		})
		if err != nil {
			log.Error().
				Err(err).
				Str("bucket", opts.Bucket).
				Str("object", aws.ToString(o.Key)).
				Msg("Unable to delete object")
			return err
		}
	}

	return nil
}
