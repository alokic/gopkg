package v2

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

var (
	ErrSQLAssertBytes  = errors.New("([]byte) type assertion failed")
	ErrSQLAssertSQLMap = errors.New("(Map) type assertion failed")
	ErrSQLAssertSQLArr = errors.New("(Arr) type assertion failed")
)

// Arr type to be serialized in db
type Arr []string

// Map type to be serialized in db
type Map map[string]string

func (a Map) Value() (driver.Value, error) {
	j, err := json.Marshal(a)
	return j, err
}

func (a *Map) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return ErrSQLAssertBytes
	}

	var i interface{}

	d := json.NewDecoder(bytes.NewReader(source))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return err
	}

	m := map[string]string{}

	switch t := i.(type) {
	case map[string]interface{}:
		for k, v := range t {
			m[k] = v.(string)
		}
	}
	*a = m
	return nil
}

func (a Arr) Value() (driver.Value, error) {
	j, err := json.Marshal(a)
	return j, err
}

func (a *Arr) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return ErrSQLAssertBytes
	}

	var i interface{}

	d := json.NewDecoder(bytes.NewReader(source))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return err
	}

	if i == nil {
		*a = []string{}
		return nil
	}

	m, ok := i.([]interface{})
	if !ok {
		return ErrSQLAssertSQLArr
	}

	arr := []string{}
	for _, v := range m {
		arr = append(arr, v.(string))
	}

	*a = arr
	return nil
}
