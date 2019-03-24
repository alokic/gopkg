package typeutils

import (
	"encoding/json"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cast"
)

func Blank(id interface{}) bool {

	// lets only do type assertion and not type conversio for basic types.
	// 3x slow.. type-converion 6x slow.. but ok
	switch id.(type) {
	case int:
		if id == int(0) {
			return true
		}
	case int32:
		if id == int32(0) {
			return true
		}
	case int64:
		if id == int64(0) {
			return true
		}

	case uint:
		if id == uint(0) {
			return true
		}
	case uint32:
		if id == uint32(0) {
			return true
		}
	case uint64:
		if id == uint64(0) {
			return true
		}

	case float32:
		if id == float32(0.0) {
			return true
		}
	case float64:
		if id == float64(0.0) {
			return true
		}

	case json.Number:
		n, _ := strconv.ParseFloat(string(id.(json.Number)), 64)
		if n == 0.0 {
			return true
		}

	case nil:
		return true

	case string:
		if id == "" {
			return true
		}

	case bool:
		if id == false {
			return true
		}
	}
	return false
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
