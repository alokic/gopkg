package redisfarm

import (
	"errors"
	"fmt"

	"github.com/garyburd/redigo/redis"
)

var (
	// ErrMultiConnBlank : ""
	ErrMultiConnBlank = errors.New("")
)

type multiConn struct {
	conns []redis.Conn
}

// newMultiConnBuilder :
func newMultiConnBuilder() (*multiConn, error) {
	return &multiConn{conns: make([]redis.Conn, 0)}, nil
}

// SetConn :
func (m *multiConn) SetConn(r redis.Conn) *multiConn {
	m.conns = append(m.conns, r)
	return m
}

// Build :
func (m *multiConn) Build() *multiConn {
	return m
}

// Close closes the connection.
func (m *multiConn) Close() error {
	err := ErrMultiConnBlank

	for _, c := range m.conns {
		e := c.Close()
		if e != nil {
			err = fmt.Errorf("%s, %s", e.Error(), err.Error())
		}
	}
	return redis.Error(err.Error())
}

// Err returns a non-nil value when the connection is not usable.
func (m *multiConn) Err() error {
	err := ErrMultiConnBlank

	for _, c := range m.conns {
		e := c.Err()
		if e != nil {
			err = fmt.Errorf("%s, %s", e.Error(), err.Error())
		}
	}
	return redis.Error(err.Error())
}

// Do sends a command to the server and returns the received reply.
func (m *multiConn) Do(commandName string, args ...interface{}) (interface{}, error) {
	replies := []interface{}{}
	err := ErrMultiConnBlank

	for _, c := range m.conns {
		r, e := c.Do(commandName, args...)
		if e != nil {
			err = fmt.Errorf("%s %s", e.Error(), err.Error())
		}
		if r != nil {
			replies = append(replies, r)
		}
	}

	return replies, redis.Error(err.Error())
}

// Send writes the command to the client's output buffer.
func (m *multiConn) Send(commandName string, args ...interface{}) error {
	err := ErrMultiConnBlank

	for _, c := range m.conns {
		e := c.Send(commandName, args...)
		if e != nil {
			err = fmt.Errorf("%s, %s", e.Error(), err.Error())
		}
	}
	return redis.Error(err.Error())
}

// Flush flushes the output buffer to the Redis server.
func (m *multiConn) Flush() error {
	err := ErrMultiConnBlank

	for _, c := range m.conns {
		e := c.Flush()
		if e != nil {
			err = fmt.Errorf("%s, %s", e.Error(), err.Error())
		}
	}
	return redis.Error(err.Error())
}

// Receive receives a single reply from the Redis server
func (m *multiConn) Receive() (reply interface{}, err error) {
	return nil, fmt.Errorf("Cluster: Receive is not supported")
}
