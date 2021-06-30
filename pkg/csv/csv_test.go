package csv

import (
	"reflect"
	"testing"
)

func Test_ParseString(t *testing.T) {
	type args struct {
		strCSV string
	}
	tests := []struct {
		name    string
		args    args
		want    []map[string]interface{}
		wantErr bool
	}{
		{
			name: "An invalid CSV should return an error",
			args: args{
				strCSV: `"timestamp","publishing_point","audience"
"2021-06-24T06:00:00.000Z","/foo_bar_00"
"2021-06-24T06:00:00.000Z","/foo_bar_01"
"2021-06-24T06:00:00.000Z","/foo_bar_02"
"2021-06-24T06:00:00.000Z","/foo_bar_03"
"2021-06-24T06:00:00.000Z","/foo_bar_04"
"2021-06-24T06:00:00.000Z","/foo_bar_05"
"2021-06-24T06:00:00.000Z","/foo_bar_06"`,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "A valid CSV should be parsed correctly",
			args: args{
				strCSV: `"timestamp","publishing_point","audience"
"2021-06-24T06:00:00.000Z","/foo_bar_00","6892"
"2021-06-24T06:00:00.000Z","/foo_bar_01","7945"
"2021-06-24T06:00:00.000Z","/foo_bar_02","12157"
"2021-06-24T06:00:00.000Z","/foo_bar_03","2017"
"2021-06-24T06:00:00.000Z","/foo_bar_04","988598"
"2021-06-24T06:00:00.000Z","/foo_bar_05","22976"
"2021-06-24T06:00:00.000Z","/foo_bar_06","6822"`,
			},
			want: []map[string]interface{}{
				{
					"timestamp":        "2021-06-24T06:00:00.000Z",
					"publishing_point": "/foo_bar_00",
					"audience":         "6892",
				},
				{
					"timestamp":        "2021-06-24T06:00:00.000Z",
					"publishing_point": "/foo_bar_01",
					"audience":         "7945",
				},
				{
					"timestamp":        "2021-06-24T06:00:00.000Z",
					"publishing_point": "/foo_bar_02",
					"audience":         "12157",
				},
				{
					"timestamp":        "2021-06-24T06:00:00.000Z",
					"publishing_point": "/foo_bar_03",
					"audience":         "2017",
				},
				{
					"timestamp":        "2021-06-24T06:00:00.000Z",
					"publishing_point": "/foo_bar_04",
					"audience":         "988598",
				},
				{
					"timestamp":        "2021-06-24T06:00:00.000Z",
					"publishing_point": "/foo_bar_05",
					"audience":         "22976",
				},
				{
					"timestamp":        "2021-06-24T06:00:00.000Z",
					"publishing_point": "/foo_bar_06",
					"audience":         "6822",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseString(tt.args.strCSV)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() = %v, want %v", got, tt.want)
			}
		})
	}
}
