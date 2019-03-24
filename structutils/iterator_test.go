package structutils_test

import (
	"reflect"
	"testing"

	"github.com/alokic/gopkg/structutils"
)

type testStruct struct {
	Env            string `json:"env" mandatory:"true" usage:"i am test env"`
	DBMaxOpenConns int    `json:"db_open_conns"`
}

func TestIterator_Next(t *testing.T) {
	iter, _ := structutils.NewIterator(&testStruct{Env: "test"}, []string{"json", "mandatory", "test"})

	tests := []struct {
		name string
		iter *structutils.Iterator
		want *structutils.Field
	}{
		{name: "TestNextSuccess", iter: iter, want: &structutils.Field{
			Tags: []structutils.Tag{
				structutils.Tag{Name: "json", Value: "env"},
				structutils.Tag{Name: "mandatory", Value: "true"},
				structutils.Tag{Name: "test", Value: ""},
			},
			Name:  "Env",
			Value: "test",
			Type:  "string",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.iter
			if got := s.Next(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Iterator.Next() = %v, want %v", got, tt.want)
			}
		})
	}
}
