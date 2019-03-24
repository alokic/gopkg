package v2

import (
	"errors"
	"reflect"
)

type structIterator struct {
	idx int
	vb  reflect.Value
	fi  *FieldInfo
}

type structField struct {
	dbtag string
	name  string
	value interface{}
}

func newStructIterator(b interface{}, fi *FieldInfo) (*structIterator, error) {
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

	return &structIterator{vb: vb, fi: fi}, nil
}

func (s *structIterator) next() *structField {
	for {
		if s.idx >= s.vb.NumField() {
			return nil
		}

		curridx := s.idx
		s.idx++

		field := s.vb.Field(curridx)

		if !field.CanInterface() {
			continue
		}

		tag := s.vb.Type().Field(curridx).Tag.Get(dbTagName)
		if tag == "" {
			continue
		}

		return &structField{
			name:  s.vb.Type().Field(curridx).Name,
			dbtag: s.fi.DBTagsArr[curridx],
			value: field.Interface(),
		}
	}
}

// isZero return wether x is the is
// the zero-value of its underlying type.
func isZero(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}
