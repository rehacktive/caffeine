package database

import (
	"reflect"
	"testing"
)

func TestMemDatabase_Upsert(t *testing.T) {
	type fields struct {
		namespaces map[string]namespace
	}
	type args struct {
		namespace string
		key       string
		value     []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *DbError
	}{
		{
			name: "case 1",
			fields: fields{
				namespaces: make(map[string]namespace),
			},
			args: args{
				namespace: "ns 1",
				key:       "first key",
				value:     []byte("true"),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &MemDatabase{
				namespaces: tt.fields.namespaces,
			}
			if got := mb.Upsert(tt.args.namespace, tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Upsert() = %v, want %v", got, tt.want)
			}
			if string(mb.namespaces[tt.args.namespace].data[tt.args.key]) != string(tt.args.value) {
				t.Errorf("Upsert() = %v, want %v", mb.namespaces[tt.args.namespace].data[tt.args.key], tt.args.value)
			}
		})
	}
}
