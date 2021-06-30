package flags

import (
	"reflect"
	"testing"
)

func TestTag_UnmarshalFlag(t *testing.T) {
	type fields struct {
		Row string
		Tag string
	}
	type args struct {
		arg string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Unmarshal invalid flag should return an error",
			fields: fields{},
			args: args{
				arg: "foo=%",
			},
			wantErr: true,
		},
		{
			name: "Unmarshal flag with key only should return a field with identical row / tag",
			fields: fields{
				Row: "foo",
				Tag: "foo",
			},
			args: args{
				arg: "foo",
			},
			wantErr: false,
		},
		{
			name: "Unmarshal flag with key only should return a field with specific tag",
			fields: fields{
				Row: "foo",
				Tag: "bar",
			},
			args: args{
				arg: "foo={tag:bar}",
			},
			wantErr: false,
		},
		{
			name: "Unmarshal flag with additional args should work as well",
			fields: fields{
				Row: "foo",
				Tag: "bar",
			},
			args: args{
				arg: "foo={tag:bar,fooArg:barArg}",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got := &Tag{}
		want := &Tag{
			Row: tt.fields.Row,
			Tag: tt.fields.Tag,
		}
		if err := got.UnmarshalFlag(tt.args.arg); (err != nil) != tt.wantErr {
			t.Errorf("Tag.UnmarshalFlag() error = %v, wantErr %v", err, tt.wantErr)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Tag.UnmarshalFlag() = %v, want %v", got, want)
		}
	}
}

func TestTag_MarshalFlag(t *testing.T) {
	type fields struct {
		Row string
		Tag string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:    "Marshal empty tag should return empty string",
			fields:  fields{},
			want:    "",
			wantErr: false,
		},
		{
			name: "Marshal tag with row only should return a properly formatted tag",
			fields: fields{
				Row: "foo",
			},
			want:    "foo",
			wantErr: false,
		},
		{
			name: "Marshal tag with row and tag should return a properly formatted tag",
			fields: fields{
				Row: "foo",
				Tag: "bar",
			},
			want:    "foo={tag:bar}",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tag{
				Row: tt.fields.Row,
				Tag: tt.fields.Tag,
			}
			got, err := tr.MarshalFlag()
			if (err != nil) != tt.wantErr {
				t.Errorf("Tag.MarshalFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Tag.MarshalFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestField_UnmarshalFlag(t *testing.T) {
	type fields struct {
		Row       string
		Field     string
		FieldType FieldType
	}
	type args struct {
		arg string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Unmarshal invalid field should return an error",
			fields: fields{},
			args: args{
				arg: "foo=%",
			},
			wantErr: true,
		},
		{
			name:   "Unmarshal flag without field type should return an error",
			fields: fields{},
			args: args{
				arg: "foo",
			},
			wantErr: true,
		},
		{
			name:   "Unmarshal flag with invalid field type should return an error",
			fields: fields{},
			args: args{
				arg: "foo={type:bar}",
			},
			wantErr: true,
		},
		{
			name: "Unmarshal flag with valid field type should return a properly formatted field",
			fields: fields{
				Row:       "foo",
				Field:     "foo",
				FieldType: FieldTypeInteger,
			},
			args: args{
				arg: "foo={type:int}",
			},
			wantErr: false,
		},
		{
			name: "Unmarshal flag with valid field type and specific field name should return a properly formatted flag",
			fields: fields{
				Row:       "foo",
				Field:     "bar",
				FieldType: FieldTypeFloat,
			},
			args: args{
				arg: "foo={type:float,field:bar}",
			},
			wantErr: false,
		},
		{
			name: "Unmarshal flag with with additional args should work as well",
			fields: fields{
				Row:       "foo",
				Field:     "bar",
				FieldType: FieldTypeString,
			},
			args: args{
				arg: "foo={type:string,field:bar,bar:baz}",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &Field{}
			want := &Field{
				Row:       tt.fields.Row,
				Field:     tt.fields.Field,
				FieldType: tt.fields.FieldType,
			}
			if err := got.UnmarshalFlag(tt.args.arg); (err != nil) != tt.wantErr {
				t.Errorf("Field.UnmarshalFlag() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Field.UnmarshalFlag() = %v, want %v", got, want)
			}
		})
	}
}

func TestField_MarshalFlag(t *testing.T) {
	type fields struct {
		Row       string
		Field     string
		FieldType FieldType
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:    "Marshal empty field should return empty string",
			fields:  fields{},
			want:    "",
			wantErr: false,
		},
		{
			name: "Marshal field with row only should return a properly formatted field",
			fields: fields{
				Row: "foo",
			},
			want:    "foo",
			wantErr: false,
		},
		{
			name: "Marshal tag with row and type should return a properly formatted tag",
			fields: fields{
				Row:       "foo",
				FieldType: FieldTypeBool,
			},
			want:    "foo={type:bool}",
			wantErr: false,
		},
		{
			name: "Marshal tag with row, type and field should return a properly formatted tag",
			fields: fields{
				Row:       "foo",
				FieldType: FieldTypeInteger,
				Field:     "bar",
			},
			want:    "foo={type:int,field:bar}",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Field{
				Row:       tt.fields.Row,
				Field:     tt.fields.Field,
				FieldType: tt.fields.FieldType,
			}
			got, err := f.MarshalFlag()
			if (err != nil) != tt.wantErr {
				t.Errorf("Field.MarshalFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Field.MarshalFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}
