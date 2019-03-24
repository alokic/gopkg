package redisfarm

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

// conn :
type conn struct {
	// conn pool for all conns except subscribers
	pool *redis.Pool

	// script object cache
	script map[string]*redis.Script

	// server url
	url string

	// script folder path
	scriptFolderPath string

	// script Map
	scriptMap map[string]ScriptHandler

	// maxIdle
	maxIdle int

	// maxActive
	maxActive int
}

// ScriptHandler : signature of script handler
type ScriptHandler func(...interface{}) (interface{}, error)

func createConn(s string, maxIdle, maxActive int) redis.Conn {
	r := &conn{
		url:       s,
		maxIdle:   maxIdle,
		maxActive: maxActive,
	}

	r.pool = newpool(r.url, maxIdle, maxActive)
	return r
}

// create conn pool
func newpool(url string, maxIdle, maxActive int) *redis.Pool {
	idleTimeout := 240 * time.Second

	return &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxIdle,
		Wait:        true,
		IdleTimeout: idleTimeout,

		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(url)
			if err != nil {
				log.Printf("Error: Subscriber conn can not be created: %s", err.Error())
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// Close closes the conn.
func (c *conn) Close() error {
	conn := c.pool.Get()
	defer conn.Close()

	return conn.Close()
}

// Err returns a non-nil value when the conn is not usable.
func (c *conn) Err() error {
	conn := c.pool.Get()
	defer conn.Close()

	return conn.Err()
}

// Do sends a command to the server and returns the received reply.
func (c *conn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	conn := c.pool.Get()
	defer conn.Close()

	return conn.Do(commandName, args...)
}

// Send writes the command to the client's output buffer.
func (c *conn) Send(commandName string, args ...interface{}) error {
	conn := c.pool.Get()
	defer conn.Close()

	return conn.Send(commandName, args...)
}

// Flush flushes the output buffer to the Redis server.
func (c *conn) Flush() error {
	conn := c.pool.Get()
	defer conn.Close()

	return conn.Flush()
}

// Receive receives a single reply from the Redis server
func (c *conn) Receive() (reply interface{}, err error) {
	conn := c.pool.Get()
	defer conn.Close()

	return conn.Receive()
}
