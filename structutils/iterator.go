package structutils

import (
	"errors"
	"reflect"
)

// Iterator struct
type Iterator struct {
	idx  int
	vb   reflect.Value
	tags []string
}

// Field defines structure of info returned for each struct field
type Field struct {
	Tags  []Tag
	Name  string
	Value interface{}
	Type  string
}

// Tag defines structure of associated tags with struct field
type Tag struct {
	Name  string
	Value string
}

// NewIterator is contructor for Iterator
func NewIterator(b interface{}, tags []string) (*Iterator, error) {
	// Check b.
	vb := reflect.ValueOf(b)
	if vb.Kind() == reflect.Ptr {
		// vb is a pointer, indirect it to get the
		// underlying value, and make sure it is a struct.
		vb = vb.Elem()
	}

	if vb.Kind() != reflect.Struct {
		return nil, errors.New("b is not a struct")
	}

	return &Iterator{vb: vb, tags: tags}, nil
}

// Next returns info of a Field in every iteration
func (s *Iterator) Next() *Field {
	for {
		if s.idx >= s.vb.NumField() {
			return nil
		}

		curridx := s.idx
		s.idx++

		fieldValue := s.vb.Field(curridx)
		if !fieldValue.CanInterface() {
			continue
		}

		fieldType := s.vb.Type().Field(curridx)

		return &Field{
			Name:  fieldType.Name,
			Tags:  s.extractTags(fieldType.Tag),
			Value: fieldValue.Interface(),
			Type:  fieldValue.Type().Name(),
		}
	}
}

func (s *Iterator) extractTags(tags reflect.StructTag) []Tag {
	var ts []Tag
	for _, v := range s.tags {
		ts = append(ts, Tag{Name: v, Value: tags.Get(v)})
	}
	return ts
}

// isZero return wether x is the is
// the zero-value of its underlying type.
func isZero(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
