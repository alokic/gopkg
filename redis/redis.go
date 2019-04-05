package redis

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

type poolConfig struct {
	maxIdleConns   int
	maxActiveConns int
	idleTimeout    int // seconds
}

//
type PoolOption func(*poolConfig)

// NewRedisPool returns a pool of redis connections.
func NewRedisPool(url string, opts ...interface{}) *redis.Pool {
	dialOpts, cfg := parsePoolOptions(opts)

	return &redis.Pool{
		MaxIdle:     cfg.maxIdleConns,
		MaxActive:   cfg.maxActiveConns,
		Wait:        true,
		IdleTimeout: time.Duration(cfg.idleTimeout) * time.Second,

		Dial: func() (redis.Conn, error) {
			return redis.DialURL(url, dialOpts...)
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// WithMaxIdleConns provides option to set maxIdleConns in connection pool.
func WithMaxIdleConns(count int) PoolOption {
	return func(cfg *poolConfig) {
		cfg.maxIdleConns = count
	}
}

// WithMaxActiveConns provides option to set maxActiveConns in connection pool.
func WithMaxActiveConns(count int) PoolOption {
	return func(cfg *poolConfig) {
		cfg.maxActiveConns = count
	}
}

// WithIdleTimeout provides option to set idleTimeout in connection pool.
func WithIdleTimeout(seconds int) PoolOption {
	return func(cfg *poolConfig) {
		cfg.idleTimeout = seconds
	}
}

func parsePoolOptions(options ...interface{}) ([]redis.DialOption, *poolConfig) {
	var dialOpts []redis.DialOption

	cfg := new(poolConfig)
	redifPoolDefaults(cfg)

	for _, opt := range options {
		switch o := opt.(type) {
		case PoolOption:
			o(cfg)
		case redis.DialOption:
			dialOpts = append(dialOpts, o)
		}
	}
	return dialOpts, cfg
}

func redifPoolDefaults(cfg *poolConfig) {
	cfg.maxIdleConns = 10
	cfg.maxActiveConns = 10
	cfg.idleTimeout = 240
}
