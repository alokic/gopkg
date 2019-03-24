package timeutils

import (
	"testing"
)

func TestUnixTimeAtStartOfToday(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{
		{name: "TestStartOfTodaySuccess", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnixTimeAtStartOfToday(); got == 0 {
				t.Errorf("UnixTimeAtStartOfToday() = %v, want %v", got, tt.want)
			} else {
				t.Log(got)
			}

		})
	}
}
