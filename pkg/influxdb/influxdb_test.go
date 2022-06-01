package influxdb

import (
	"reflect"
	"testing"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/quortex/influxdb-athena-crawler/pkg/flags"
)

func Test_toPoints(t *testing.T) {
	type args struct {
		rows        []map[string]interface{}
		measurement string
		tsLayout    string
		tsRow       string
		tags        []*flags.Tag
		fields      []*flags.Field
	}
	tests := []struct {
		name    string
		args    args
		want    []*write.Point
		wantErr bool
	}{
		{
			name: "Empty rows should return empty points",
			args: args{
				rows: nil,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Invalid timestamp should return an error",
			args: args{
				rows: []map[string]interface{}{
					{
						"timestamp": "foo",
					},
				},
				tsLayout: "2006-01-02T15:04:05.000Z",
				tsRow:    "timestamp",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid field type should return an error",
			args: args{
				rows: []map[string]interface{}{
					{
						"timestamp": "2021-06-30T13:06:18.000Z",
						"foo":       "bar",
					},
				},
				tsLayout: "2006-01-02T15:04:05.000Z",
				tsRow:    "timestamp",
				fields: []*flags.Field{
					{
						Row:       "foo",
						Field:     "foo",
						FieldType: flags.FieldTypeBool,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Measurement should be set",
			args: args{
				rows: []map[string]interface{}{
					{
						"timestamp": "2021-06-30T13:06:18.000Z",
					},
				},
				tsLayout: "2006-01-02T15:04:05.000Z",
				tsRow:    "timestamp",
				tags:     nil,
				fields:   nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Timestamp should be formatted correctly",
			args: args{
				rows: []map[string]interface{}{
					{
						"timestamp": "2021-06-30T13:06:18.000Z",
					},
				},
				measurement: "foo",
				tsLayout:    "2006-01-02T15:04:05.000Z",
				tsRow:       "timestamp",
				tags:        nil,
				fields:      nil,
			},
			want: []*write.Point{
				influxdb2.NewPointWithMeasurement("foo").
					SetTime(time.Date(2021, 6, 30, 13, 6, 18, 0, time.UTC)),
			},
			wantErr: false,
		},
		{
			name: "Missing row for tag or field should be ignored",
			args: args{
				rows: []map[string]interface{}{
					{
						"timestamp": "2022-06-30T13:06:18.000Z",
						"foo":       "bar",
					},
				},
				measurement: "foo",
				tsLayout:    "2006-01-02T15:04:05.000Z",
				tsRow:       "timestamp",
				tags: []*flags.Tag{
					{
						Row: "foo",
						Tag: "foo",
					},
					{
						Row: "bar",
						Tag: "bar",
					},
				},
				fields: []*flags.Field{
					{
						Row:       "foo",
						Field:     "foo",
						FieldType: flags.FieldTypeString,
					},
					{
						Row:       "baz",
						Field:     "baz",
						FieldType: flags.FieldTypeString,
					},
				},
			},
			want: []*write.Point{
				influxdb2.NewPointWithMeasurement("foo").
					SetTime(time.Date(2022, 6, 30, 13, 6, 18, 0, time.UTC)).
					AddTag("foo", "bar").
					AddField("foo", "bar"),
			},
			wantErr: false,
		},
		{
			name: "Complete valid rows should be converted to points correctly",
			args: args{
				rows: []map[string]interface{}{
					{
						"timestamp":   "2022-06-30T13:06:18.000Z",
						"tag1Row":     "row1tag1Value",
						"tag2Row":     "row1tag2Value",
						"fieldBool":   true,
						"fieldInt":    64,
						"fieldFloat":  12.76,
						"fieldString": "foo",
					},
					{
						"timestamp":   "2022-06-30T13:06:18.000Z",
						"tag1Row":     "row2tag1Value",
						"tag2Row":     "row2tag2Value",
						"fieldBool":   true,
						"fieldInt":    32,
						"fieldFloat":  3.4567,
						"fieldString": "bar",
					},
				},
				measurement: "foo",
				tsLayout:    "2006-01-02T15:04:05.000Z",
				tsRow:       "timestamp",
				tags: []*flags.Tag{
					{
						Row: "tag1Row",
						Tag: "tag1Foo",
					},
					{
						Row: "tag2Row",
						Tag: "tag2Foo",
					},
				},
				fields: []*flags.Field{
					{
						Row:       "fieldBool",
						Field:     "fieldBoolFoo",
						FieldType: flags.FieldTypeBool,
					},
					{
						Row:       "fieldInt",
						Field:     "fieldIntFoo",
						FieldType: flags.FieldTypeInteger,
					},
					{
						Row:       "fieldFloat",
						Field:     "fieldFloatFoo",
						FieldType: flags.FieldTypeFloat,
					},
					{
						Row:       "fieldString",
						Field:     "fieldStringFoo",
						FieldType: flags.FieldTypeString,
					},
				},
			},
			want: []*write.Point{
				influxdb2.NewPointWithMeasurement("foo").
					SetTime(time.Date(2022, 6, 30, 13, 6, 18, 0, time.UTC)).
					AddTag("tag1Foo", "row1tag1Value").
					AddTag("tag2Foo", "row1tag2Value").
					AddField("fieldBoolFoo", true).
					AddField("fieldIntFoo", 64).
					AddField("fieldFloatFoo", 12.76).
					AddField("fieldStringFoo", "foo"),
				influxdb2.NewPointWithMeasurement("foo").
					SetTime(time.Date(2022, 6, 30, 13, 6, 18, 0, time.UTC)).
					AddTag("tag1Foo", "row2tag1Value").
					AddTag("tag2Foo", "row2tag2Value").
					AddField("fieldBoolFoo", true).
					AddField("fieldIntFoo", 32).
					AddField("fieldFloatFoo", 3.4567).
					AddField("fieldStringFoo", "bar"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toPoints(tt.args.rows, tt.args.measurement, tt.args.tsLayout, tt.args.tsRow, tt.args.tags, tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("toPoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toPoints() = %v, want %v", got, tt.want)
			}
		})
	}
}
