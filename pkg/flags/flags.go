package flags

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

// regexFlagMap is a regex used toi extract map types flags
var regexFlagMap = regexp.MustCompile(`^(\w*)(?:={(.*)})?$`)

// flagMap handles map types flags marshalling
type flagMap struct {
	k string
	v map[string]string
}

// unmarshalFlag converts flag args to flagMap
func unmarshalFlag(arg string) (*flagMap, error) {
	matchs := regexFlagMap.FindStringSubmatch(arg)
	if len(matchs) < 3 {
		return nil, fmt.Errorf("%q failed to parse", arg)
	}
	res := flagMap{k: matchs[1], v: make(map[string]string)}
	values := matchs[2]
	if values != "" {
		for _, e := range strings.Split(values, ",") {
			s := strings.Split(e, ":")
			if len(s) != 2 {
				return nil, fmt.Errorf("%q failed to parse", arg)
			}
			res.v[s[0]] = s[1]
		}
	}

	return &res, nil
}

// marshalFlag converts flagMap to flag args
func (m flagMap) marshalFlag() (string, error) {
	// Arbitrary sorting of the keys to always have the same marshalling
	keys := make([]string, 0, len(m.v))
	for k := range m.v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var args string
	argList := []string{}
	for _, k := range keys {
		argList = append(argList, fmt.Sprintf("%s:%s", k, m.v[k]))
	}
	if len(argList) > 0 {
		args = fmt.Sprintf("={%s}", strings.Join(argList, ","))
	}
	return fmt.Sprintf("%s%s", m.k, args), nil
}

// Tag describes an InfluxDB tag flag
type Tag struct {
	Tag, Row string
}

// UnmarshalFlag is the go-flags Value UnmarshalFlag implementation for Tag
func (t *Tag) UnmarshalFlag(arg string) error {
	fm, err := unmarshalFlag(arg)
	if err != nil {
		return fmt.Errorf("%q failed to parse", err)
	}
	t.Tag = fm.k
	t.Row = fm.v["row"]
	if t.Row == "" {
		t.Row = t.Tag
	}
	return nil
}

// MarshalFlag is the go-flags Value MarshalFlag implementation for Tag
func (t *Tag) MarshalFlag() (string, error) {
	m := flagMap{k: t.Tag, v: map[string]string{}}
	if t.Row != "" {
		m.v["row"] = string(t.Row)
	}

	return m.marshalFlag()
}

// FieldType describes an InfluxDB field type.
// Field values can be floats, integers, strings, or Booleans.
type FieldType string

// All InfluxDB field types
const (
	FieldTypeFloat   FieldType = "float"
	FieldTypeInteger FieldType = "int"
	FieldTypeString  FieldType = "string"
	FieldTypeBool    FieldType = "bool"
)

// isValid returns if the FieldType is a valid one
func (t FieldType) isValid() bool {
	switch t {
	case FieldTypeFloat, FieldTypeInteger, FieldTypeString, FieldTypeBool:
		return true
	}
	return false
}

// Field describes an InfluxDB field tag.
type Field struct {
	Field, Row string
	FieldType  FieldType
}

// UnmarshalFlag is the go-flags Value UnmarshalFlag implementation for Field
func (f *Field) UnmarshalFlag(arg string) error {
	fm, err := unmarshalFlag(arg)
	if err != nil {
		return fmt.Errorf("%q failed to parse", err)
	}

	field := fm.k
	row := fm.v["row"]
	if row == "" {
		row = field
	}

	fType := FieldType(fm.v["type"])
	if !fType.isValid() {
		return fmt.Errorf("%q invalid field type", arg)
	}

	f.Field = field
	f.Row = row
	f.FieldType = fType
	return nil
}

// MarshalFlag is the go-flags Value MarshalFlag implementation for Field
func (f *Field) MarshalFlag() (string, error) {
	m := flagMap{k: f.Field, v: map[string]string{}}
	if f.FieldType != "" {
		m.v["type"] = string(f.FieldType)
	}
	if f.Row != "" {
		m.v["row"] = f.Row
	}

	return m.marshalFlag()
}

// Options wraps all flags
type Options struct {
	Region              string        `long:"region" description:"The AWS region." required:"true"`
	Bucket              string        `long:"bucket" description:"The AWS bucket to watch." required:"true"`
	Prefix              string        `long:"prefix" description:"The bucket prefix."`
	Suffix              string        `long:"suffix" description:"Filename suffix to limit files read on the bucket."`
	ProcessedFlagSuffix string        `long:"processed-flag-suffix" description:"Filename suffix to mark csv files as processed on the bucket." default:"processed"`
	CleanObjects        bool          `long:"clean-objects" description:"Whether to delete S3 objects after processing them."`
	S3MaxFileAge        time.Duration `long:"s3-max-file-age" description:"When cleanup is activated, only trigger deletion if csv is at leas this old." default:"10m"`
	Timeout             time.Duration `long:"timeout" description:"The global timeout." default:"30s"`
	InfluxServers       []string      `long:"influx-server" description:"The InfluxDB servers addresses." required:"true"`
	InfluxToken         string        `long:"influx-token" description:"The InfluxDB token." required:"true"`
	InfluxOrg           string        `long:"influx-org" description:"The InfluxDB org to write to." required:"true"`
	InfluxBucket        string        `long:"influx-bucket" description:"The InfluxDB bucket write to." required:"true"`
	Measurement         string        `long:"measurement" description:"A measurement acts as a container for tags, fields, and timestamps. Use a measurement name that describes your data." required:"true"`
	TimestampRow        string        `long:"timestamp-row" description:"The timestamp row in CSV." default:"timestamp"`
	TimestampLayout     string        `long:"timestamp-layout" description:"The layout to parse timestamp." default:"2006-01-02T15:04:05.000Z"`
	Tags                []*Tag        `long:"tag" description:"Tags to add to InfluxDB point. Could be of the form --tag=foo if tag name matches CSV row or --tag='foo={row:bar}' to specify row."`
	Fields              []*Field      `long:"field" description:"Fields to add to InfluxDB point. Could be of the form --field='foo={type:int,row:bar}', if not specified, CSV row matches field name. Type can be float, int, string or bool."`
}

// Parse parses flags into give Option
func Parse(opts *Options) error {
	parser := flags.NewParser(opts, flags.Default)
	_, err := parser.Parse()
	return err
}
