package flags

import (
	"reflect"
	"testing"
)

func TestTag_UnmarshalFlag(t *testing.T) {
	type fields struct {
		Tag string
		Row string
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
			name:   "Unmarshal invalid flag should return an error",
			fields: fields{},
			args: args{
				arg: "foo=%",
			},
			wantErr: true,
		},
		{
			name: "Unmarshal flag with key only should return a field with identical row / tag",
			fields: fields{
				Tag: "foo",
				Row: "foo",
			},
			args: args{
				arg: "foo",
			},
			wantErr: false,
		},
		{
			name: "Unmarshal flag with key only should return a field with specific tag",
			fields: fields{
				Tag: "foo",
				Row: "bar",
			},
			args: args{
				arg: "foo={row:bar}",
			},
			wantErr: false,
		},
		{
			name: "Unmarshal flag with additional args should work as well",
			fields: fields{
				Tag: "foo",
				Row: "bar",
			},
			args: args{
				arg: "foo={row:bar,fooArg:barArg}",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got := &Tag{}
		want := &Tag{
			Tag: tt.fields.Tag,
			Row: tt.fields.Row,
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
		Tag string
		Row string
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
				Tag: "foo",
			},
			want:    "foo",
			wantErr: false,
		},
		{
			name: "Marshal tag with row and tag should return a properly formatted tag",
			fields: fields{
				Tag: "foo",
				Row: "bar",
			},
			want:    "foo={row:bar}",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Tag{
				Tag: tt.fields.Tag,
				Row: tt.fields.Row,
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
		Field     string
		Row       string
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
				Field:     "foo",
				Row:       "foo",
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
				Field:     "foo",
				Row:       "bar",
				FieldType: FieldTypeFloat,
			},
			args: args{
				arg: "foo={type:float,row:bar}",
			},
			wantErr: false,
		},
		{
			name: "Unmarshal flag with with additional args should work as well",
			fields: fields{
				Field:     "foo",
				Row:       "bar",
				FieldType: FieldTypeString,
			},
			args: args{
				arg: "foo={type:string,row:bar,bar:baz}",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &Field{}
			want := &Field{
				Field:     tt.fields.Field,
				Row:       tt.fields.Row,
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
		Field     string
		Row       string
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
				Field: "foo",
			},
			want:    "foo",
			wantErr: false,
		},
		{
			name: "Marshal tag with row and type should return a properly formatted tag",
			fields: fields{
				Field:     "foo",
				FieldType: FieldTypeBool,
			},
			want:    "foo={type:bool}",
			wantErr: false,
		},
		{
			name: "Marshal tag with row, type and field should return a properly formatted tag",
			fields: fields{
				Field:     "foo",
				Row:       "bar",
				FieldType: FieldTypeInteger,
			},
			want:    "foo={row:bar,type:int}",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Field{
				Field:     tt.fields.Field,
				Row:       tt.fields.Row,
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
