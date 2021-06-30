package influxdb

import (
	"context"
	"fmt"
	"strconv"
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
	tsLayout, tsRow string
	tags            []*flags.Tag
	fields          []*flags.Field
}

// NewWriter returns an Writer implementation from given parameters
func NewWriter(
	server, token, org, bucket, tsLayout, tsRow string,
	tags []*flags.Tag,
	fields []*flags.Field,
) Writer {
	cli := influxdb2.NewClient(server, token)
	api := cli.WriteAPIBlocking(org, bucket)
	return &writer{
		cli:      cli,
		api:      api,
		tsLayout: tsLayout,
		tsRow:    tsRow,
		tags:     tags,
		fields:   fields,
	}
}

// WriteRecords parses given rows and write appropriate points to InfluxDB instance
func (w *writer) WriteRecords(ctx context.Context, rows []map[string]interface{}) error {
	// Convert csv rows to InfluxDB points
	points, err := toPoints(rows, w.tsLayout, w.tsRow, w.tags, w.fields)
	if err != nil {
		return fmt.Errorf("Failed to convert CSV rows to points: %s", err)
	}

	// Write points to InfluxDB
	err = w.api.WritePoint(context.Background(), points...)
	if err != nil {
		return fmt.Errorf("Failed to write points to InfluxDB: %s", err)
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
	tsLayout, tsRow string,
	tags []*flags.Tag,
	fields []*flags.Field,
) ([]*write.Point, error) {
	if rows == nil {
		return nil, nil
	}
	res := make([]*write.Point, len(rows))
	for i, e := range rows {
		p, err := toPoint(e, tsLayout, tsRow, tags, fields)
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
	tsLayout, tsRow string,
	tags []*flags.Tag,
	fields []*flags.Field,
) (*write.Point, error) {
	t, err := time.Parse(tsLayout, fmt.Sprintf("%v", row[tsRow]))
	if err != nil {
		return nil, err
	}

	point := influxdb2.NewPointWithMeasurement("audience").SetTime(t)
	for _, e := range tags {
		val, ok := row[e.Row]
		if !ok {
			return nil, fmt.Errorf("Invalid row for tag: %s", e.Row)
		}
		point = point.AddTag(e.Tag, fmt.Sprintf("%v", val))
	}

	for _, e := range fields {
		val, ok := row[e.Row]
		if !ok {
			return nil, fmt.Errorf("Invalid row for field: %s", e.Row)
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
