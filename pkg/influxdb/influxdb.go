package influxdb

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/quortex/influxdb-athena-crawler/pkg/flags"
)

// Writer describes what an InfluxDB writer should do
type Writer interface {
	WriteRecords(ctx context.Context, rows []map[string]interface{}) error
	Close()
}

// writer is the Writer implementation
type writer struct {
	cli             influxdb2.Client
	api             api.WriteAPIBlocking
	measurement     string
	tsLayout, tsRow string
	tags            []*flags.Tag
	fields          []*flags.Field
}

// NewWriter returns an Writer implementation from given parameters
func NewWriter(
	server, token, org, bucket, measurement, tsLayout, tsRow string,
	tags []*flags.Tag,
	fields []*flags.Field,
) Writer {
	cli := influxdb2.NewClient(server, token)
	api := cli.WriteAPIBlocking(org, bucket)
	return &writer{
		cli:         cli,
		api:         api,
		measurement: measurement,
		tsLayout:    tsLayout,
		tsRow:       tsRow,
		tags:        tags,
		fields:      fields,
	}
}

// WriteRecords parses given rows and write appropriate points to InfluxDB instance
func (w *writer) WriteRecords(ctx context.Context, rows []map[string]interface{}) error {
	// Convert csv rows to InfluxDB points
	points, err := toPoints(rows, w.measurement, w.tsLayout, w.tsRow, w.tags, w.fields)
	if err != nil {
		return fmt.Errorf("failed to convert CSV rows to points: %s", err)
	}

	// No points to write, return immediately
	if len(points) == 0 {
		return nil
	}

	// Write points to InfluxDB
	err = w.api.WritePoint(context.Background(), points...)
	if err != nil {
		return fmt.Errorf("failed to write points to InfluxDB: %s", err)
	}

	return nil
}

// Close closes InfluxDB client
func (w *writer) Close() {
	w.cli.Close()
}

// toPoints converts rows slice to InfluxDB points slice
func toPoints(
	rows []map[string]interface{},
	measurement string,
	tsLayout, tsRow string,
	tags []*flags.Tag,
	fields []*flags.Field,
) ([]*write.Point, error) {
	if rows == nil {
		return nil, nil
	}
	res := make([]*write.Point, len(rows))
	for i, e := range rows {
		p, err := toPoint(e, measurement, tsLayout, tsRow, tags, fields)
		if err != nil {
			return nil, err
		}
		res[i] = p
	}
	return res, nil
}

// toPoints converts rows to InfluxDB points
func toPoint(
	row map[string]interface{},
	measurement string,
	tsLayout, tsRow string,
	tags []*flags.Tag,
	fields []*flags.Field,
) (*write.Point, error) {
	t, err := time.Parse(tsLayout, fmt.Sprintf("%v", row[tsRow]))
	if err != nil {
		return nil, err
	}

	if measurement == "" {
		return nil, fmt.Errorf("invalid measurement: measurement required")
	}

	point := influxdb2.NewPointWithMeasurement(measurement).SetTime(t)
	for _, e := range tags {
		val, ok := row[e.Row]
		if !ok {
			return nil, fmt.Errorf("invalid row for tag: %s", e.Row)
		}
		point = point.AddTag(e.Tag, fmt.Sprintf("%v", val))
	}

	for _, e := range fields {
		val, ok := row[e.Row]
		if !ok {
			return nil, fmt.Errorf("invalid row for field: %s", e.Row)
		}
		var fieldVal interface{}
		strField := fmt.Sprintf("%v", val)
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

// writers is a Writer implementation for multiple Writers
type writers []Writer

// NewWriters returns a Writers implementation from given parameters
func NewWriters(
	servers []string,
	token, org, bucket, measurement, tsLayout, tsRow string,
	tags []*flags.Tag,
	fields []*flags.Field,
) Writer {
	w := make(writers, len(servers))
	for i, server := range servers {
		w[i] = NewWriter(
			server,
			token,
			org,
			bucket,
			measurement,
			tsLayout,
			tsRow,
			tags,
			fields,
		)
	}

	return &w
}

// WriteRecords parses given rows and write appropriate points to InfluxDB instance
func (w *writers) WriteRecords(ctx context.Context, rows []map[string]interface{}) error {
	// Make waitgroup and channels to process
	// tasks asynchronously
	var wg sync.WaitGroup
	cDone := make(chan bool)
	cErr := make(chan error)
	wg.Add(len(*w))

	go func() {
		for _, item := range *w {
			writer := item
			go func() {
				defer wg.Done()
				if err := writer.WriteRecords(ctx, rows); err != nil {
					cErr <- err
				}
			}()
		}
		wg.Wait()
		close(cDone)
	}()

	// Wait until either WaitGroup is done or an error is received through the channel
	select {
	case <-cDone:
		break
	case err := <-cErr:
		return err
	}

	return nil
}

// Close closes InfluxDB client
func (w *writers) Close() {
	var wg sync.WaitGroup
	wg.Add(len(*w))

	go func() {
		for _, item := range *w {
			writer := item
			go func() {
				defer wg.Done()
				writer.Close()
			}()
		}
	}()

	wg.Wait()
}
