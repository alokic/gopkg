package funcutil

import "time"

func Retry(f func() error, retries int, backoff time.Duration) error {
	var err error
	for i := 0; i <= retries; i++ {
		err = f()
		if err == nil {
			return nil
		}
		time.Sleep(backoff)
	}
	return err
}
