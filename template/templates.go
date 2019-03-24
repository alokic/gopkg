package template

import (
	"bytes"
)

// ApplyEnv apply env variables on a text file
func ApplyEnv(file string) ([]byte, error) {
	t, err := New(file)
	if err != nil {
		return nil, err
	}

	var b []byte
	buf := bytes.NewBuffer(b)
	err = t.Execute(buf, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
