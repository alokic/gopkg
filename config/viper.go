package config

import (
	"bytes"
	"errors"
	"strings"
	"sync"

	"github.com/alokic/gopkg/template"
	"github.com/spf13/viper"
)

// Viper :
type Viper struct {
	fname []string
	ftype string
	mu    sync.RWMutex
	v     *viper.Viper
}

var (
	ErrInvalidFileName   = errors.New("invalid file name")
	ErrInvalidFileFormat = errors.New("invalid file format")
)

// NewViper :
//   fname: file name with absolute path // example:  "/a/b/c/d.yml", mandatory
func NewViper(fname ...string) (*Viper, error) {

	y := &Viper{fname: fname}

	var data [][]byte
	{
		for _, f := range fname {
			a := strings.Split(f, ".")
			if len(a) < 2 {
				return nil, ErrInvalidFileName
			}

			var d []byte
			var err error

			switch a[len(a)-1] {
			case "yaml", "yml":
				d, err = template.ApplyEnv(f)
			default:
				err = ErrInvalidFileFormat
			}
			if err != nil {
				return nil, err
			}
			data = append(data, d)
		}
	}

	y.ftype = "yaml"
	y.createViper(data)
	return y, nil
}

// Get :
func (c *Viper) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.v.Get(key)
}

// Set :
func (c *Viper) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.v.Set(key, value)
}

func (c *Viper) createViper(data [][]byte) {
	v := viper.New()
	c.v = v
	v.SetConfigType(c.ftype)
	v.AutomaticEnv()

	for _, d := range data {
		v.MergeConfig(bytes.NewBuffer(d))
	}
}
