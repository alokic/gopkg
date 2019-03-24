package funcutil

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var numFails int
var currFails int

func TestRetry(t *testing.T) {

	exec := func() error {
		t.Log(fmt.Sprintf("in func: curr: %v, total: %v", currFails, numFails))
		if currFails == numFails {
			return nil
		}
		currFails++
		return errors.New("failed")
	}

	tests := map[string]struct {
		numFailure int
		retries    int
		f          func() error
		backOff    time.Duration
	}{
		"success in 1 go": {
			numFailure: 0,
			retries:    0,
			f:          exec,
			backOff:    time.Second,
		},
		"failure with 0 retries": {
			numFailure: 1,
			retries:    0,
			f:          exec,
			backOff:    time.Second,
		},
		"failure with 2 retries": {
			numFailure: 3,
			retries:    2,
			f:          exec,
			backOff:    time.Second,
		},
		"success with 2 retries": {
			numFailure: 2,
			retries:    2,
			f:          exec,
			backOff:    time.Second,
		},
	}

	for tName, tInfo := range tests {
		t.Log("Running test", tName)
		currFails = 0
		numFails = tInfo.numFailure
		err := Retry(tInfo.f, tInfo.retries, tInfo.backOff)
		if tInfo.numFailure > tInfo.retries {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
