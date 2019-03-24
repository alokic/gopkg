package config

import (
	"errors"
	"fmt"
)

// Client :
type Client interface {
	Get(string) interface{}
	Set(string, interface{})
	SetNamespace(string)
}

var (
	defaultValue  = "default"
	defaultClient Client
	ErrNoFiles    = errors.New("no files given")
)

type client struct {
	currentNs string
	v         *Viper
}

func New(fname ...string) (Client, error) {
	if len(fname) == 0 {
		return nil, ErrNoFiles
	}

	v, err := NewViper(fname...)
	if err != nil {
		return nil, err
	}
	return &client{v: v}, nil
}

func (c *client) SetNamespace(currentNs string) {
	c.currentNs = currentNs
}

// Get a key. Searches key in following order:
// lowercase(<input key>) in env, <input key> in configMap,  <currentNsKey>.<input key> in configMap
func (c *client) Get(key string) interface{} {
	if c == nil {
		return ""
	}

	v := c.v.Get(key)
	if v != nil {
		return v
	}

	return c.v.Get(fmt.Sprintf("%s.%s", c.currentNs, key))
}

func (c *client) GetString(key string) string {
	if c == nil {
		return ""
	}

	return c.Get(key).(string)
}

func (c *client) GetStringArr(key string) []string {
	if c == nil {
		return []string{}
	}

	var arr []string
	for _, v := range c.Get(key).([]interface{}) {
		arr = append(arr, v.(string))
	}

	return arr
}

func (c *client) Set(key string, val interface{}) {
	if c == nil {
		return
	}

	c.v.Set(key, val)
}

func SetDefault(cl Client) {
	defaultClient = cl
}

func GetDefault() Client {
	return defaultClient
}

func Get(key string) interface{} {
	return defaultClient.Get(key)
}

func GetString(key string) string {
	c, ok := defaultClient.(*client)
	if !ok {
		return ""
	}
	return c.GetString(key)
}

func GetStringArr(key string) []string {
	c, ok := defaultClient.(*client)
	if !ok {
		return nil
	}
	return c.GetStringArr(key)
}

func Set(key string, val interface{}) {
	defaultClient.Set(key, val)
}
