package main

import (
	"bytes"
	"context"
	"slices"
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

	unprocCsvs, procCsvs := filterBucketContent(*res, opts.Suffix, opts.ProcessedFlagSuffix)

	if len(procCsvs.Contents)+len(unprocCsvs.Contents) == 0 {
		log.Info().Msg("No objects matching bucket / prefix, processing done !")
		return
	}

	dwn := manager.NewDownloader(s3Cli)
	upl := manager.NewUploader(s3Cli)
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

	if len(unprocCsvs.Contents) > 0 {
		parallelApply(unprocCsvs, func(o types.Object) {
			if err := processObject(ctx, dwn, upl, influxWriter, o); err != nil {
				log.Error().Err(err).Msg("Processing error")
			}
		})
	}

	if opts.CleanObjects && len(procCsvs.Contents) > 0 {
		parallelApply(procCsvs, func(o types.Object) {
			if err := cleanObject(ctx, s3Cli, o); err != nil {
				log.Error().Err(err).Msg("Cleaning error")
			}
		})
	}

	log.Info().
		Dur("elapsed", time.Since(start)).
		Msg("Processing ended !")
}

func filterBucketContent(elems s3.ListObjectsOutput, suffix, processedFlagSuffix string) (unprocessed, processed s3.ListObjectsOutput) {
	// Rely on .processed files present on the bucket to detect which csv
	// have already been pushed to influx and which have yet to be processed
	csvFiles := []types.Object{}
	processedElems := []string{}

	if len(suffix) == 0 {
		return unprocessed, processed
	}
	for _, s := range elems.Contents {
		if strings.HasSuffix(*s.Key, suffix) {
			csvFiles = append(csvFiles, s)
		}
		if strings.HasSuffix(*s.Key, processedFlagSuffix) {
			processedElems = append(processedElems, strings.ReplaceAll(*s.Key, processedFlagSuffix, suffix))
		}
	}

	for _, o := range csvFiles {
		if !slices.Contains(processedElems, *o.Key) {
			unprocessed.Contents = append(unprocessed.Contents, o)
		} else {
			processed.Contents = append(processed.Contents, o)
		}
	}

	return unprocessed, processed
}

func parallelApply(list s3.ListObjectsOutput, fn func(o types.Object)) {
	var wg sync.WaitGroup
	wg.Add(len(list.Contents))
	for _, item := range list.Contents {
		o := item
		go func() {
			defer wg.Done()
			fn(o)
		}()
	}
	wg.Wait()
}

func processObject(
	ctx context.Context,
	s3Dwn *manager.Downloader,
	s3Upl *manager.Uploader,
	influxWriter influxdb.Writer,
	o types.Object,
) error {
	log.Info().
		Str("object", aws.ToString(o.Key)).
		Time("last modified", aws.ToTime(o.LastModified)).
		Int64("size", *o.Size).
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

	// Add .processed file to S3 bucket to avoid writing the same file to influx twice.
	markerFileName := strings.ReplaceAll(*o.Key, opts.Suffix, opts.ProcessedFlagSuffix)
	if _, err = s3Upl.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(opts.Bucket),
		Key:    &markerFileName,
		Body:   bytes.NewReader([]byte{0}),
	}); err != nil {
		log.Error().
			Err(err).
			Str("object", aws.ToString(o.Key)).
			Msg("Failed to create .processed file")
		return err
	}
	return nil
}

func cleanObject(
	ctx context.Context,
	s3Cli *s3.Client,
	o types.Object,
) error {
	if time.Since(aws.ToTime(o.LastModified)) > opts.S3MaxFileAge {
		// Delete object
		log.Info().
			Str("object", aws.ToString(o.Key)).
			Time("last modified", aws.ToTime(o.LastModified)).
			Int64("size", o.Size).
			Msg("Cleaning s3 object")

		_, err := s3Cli.DeleteObject(ctx, &s3.DeleteObjectInput{
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

		markerFileName := strings.ReplaceAll(*o.Key, opts.Suffix, opts.ProcessedFlagSuffix)
		_, err = s3Cli.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(opts.Bucket),
			Key:    &markerFileName,
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
