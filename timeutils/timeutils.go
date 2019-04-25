package timeutils

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
)

// UnixTime in milliseconds
func UnixTime() int64 {
	return int64(time.Nanosecond) * (time.Now().UnixNano() / int64(time.Millisecond))
}

// UnixTimeAtStartOfToday in milliseconds
func UnixTimeAtStartOfToday() int64 {
	n := time.Now()
	return int64(time.Nanosecond) * (time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC).UnixNano() / int64(time.Millisecond))
}

// DaysTillToday number of days till today
func DaysTillToday(unixtimestamp int64) time.Duration {
	tm := time.Unix(unixtimestamp, 0)
	return time.Now().Sub(tm)
}

// ToProto converts to proto3 time.
func ToProto(t time.Time) (*tspb.Timestamp, error) {
	return ptypes.TimestampProto(t)
}

// FromProto converts from proto3 time to time.Time.
func FromProto(t *tspb.Timestamp) (time.Time, error) {
	return ptypes.Timestamp(t)
}
