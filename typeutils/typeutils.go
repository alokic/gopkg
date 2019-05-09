package typeutils

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cast"
)

var epochTime = time.Unix(0, 0).UTC()

func Blank(id interface{}) bool {
	switch reflect.TypeOf(id).Kind() {
	case reflect.TypeOf(time.Time{}).Kind():
		t := reflect.ValueOf(id).Interface().(time.Time)
		return (t.Sub(epochTime) == 0) || t.IsZero()
	default:
		return reflect.DeepEqual(reflect.ValueOf(id), reflect.Zero(reflect.TypeOf(id)))
	}
}

func Present(id interface{}) bool {
	return !Blank(id)
}

func ToStr(v interface{}) string {
	return cast.ToString(v)
}

func ToUint64(v interface{}) uint64 {
	return cast.ToUint64(v)
}

func ToInt64(v interface{}) int64 {
	return cast.ToInt64(v)
}

func ToFloat64(v interface{}) float64 {
	return cast.ToFloat64(v)
}
func ToInt(v interface{}) int {
	return cast.ToInt(v)
}

func ToId(v interface{}) uint64 {
	return ToUint64(v)
}

func ToUnixTime(v interface{}) int64 {
	switch v.(type) {
	case time.Time:
		return v.(time.Time).UnixNano() / int64(time.Millisecond)
	default:
		return int64(ToId(v))
	}
}

func ToBool(v interface{}) bool {
	return cast.ToBool(v)
}

// ToTime deprecated
func ToTime(v interface{}) int64 {
	return ToUnixTime(v)
}

func NormalizeNumber(d map[string]interface{}) map[string]interface{} {
	for k, v := range d {
		s := ""
		switch v.(type) {
		case json.Number:
			s = string(v.(json.Number))
		}

		if s != "" {
			match, _ := regexp.MatchString("\\.", s)
			if match {
				n, err := strconv.ParseFloat(s, 64)
				if err == nil {
					d[k] = n
				}
			} else {
				n, err := strconv.ParseUint(s, 10, 64)
				if err == nil {
					d[k] = n
				}
			}
		}

	}
	return d
}
