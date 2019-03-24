package redisfarm

import (
	"errors"

	"github.com/garyburd/redigo/redis"
)

var (
	// ErrClusterNoConn : "No conn exists"
	ErrClusterNoConn = errors.New("No conn exists")

	defaultIdleConns = 6
)

// ConnSelectHandler : Handler to select shard
type ConnSelectHandler func(string) (int, error)

// ClusterBuilder :
type ClusterBuilder struct {
	c *Cluster
}

// Cluster :
type Cluster struct {
	conns     []redis.Conn
	maxIdle   int
	maxActive int
	CS        ConnSelectHandler
}

// NewClusterBuilder :
func NewClusterBuilder() *ClusterBuilder {
	return &ClusterBuilder{c: newCluster()}
}

// SetServers :
func (c *ClusterBuilder) SetServers(servers []string) *ClusterBuilder {
	maxIdle, maxActive := c.c.maxIdle, c.c.maxActive

	if maxIdle == 0 {
		maxIdle = defaultIdleConns
	}

	for _, v := range servers {
		c.c.conns = append(c.c.conns, createConn(v, maxIdle, maxActive))
	}
	return c
}

// SetMaxIdleConns :
func (c *ClusterBuilder) SetMaxIdleConns(n int) *ClusterBuilder {
	c.c.maxIdle = n
	return c
}

// SetMaxActiveConns :
func (c *ClusterBuilder) SetMaxActiveConns(n int) *ClusterBuilder {
	c.c.maxActive = n
	return c
}

// SetConnSelectHandler :
func (c *ClusterBuilder) SetConnSelectHandler(fn ConnSelectHandler) *ClusterBuilder {
	c.c.CS = fn
	return c
}

// Build :
func (c *ClusterBuilder) Build() (*Cluster, error) {
	return c.c.build()
}

func newCluster() *Cluster {
	return &Cluster{conns: make([]redis.Conn, 0), CS: defaultConnSelectHandler}
}

func (c *Cluster) build() (*Cluster, error) {
	if len(c.conns) == 0 {
		return nil, ErrClusterNoConn
	}
	return c, nil
}

// GetConn : implements consistent hash
// TODO
func (c *Cluster) GetConn(keyspace string) (redis.Conn, error) {
	idx, err := c.CS(keyspace)
	if err != nil {
		return nil, err
	}
	return c.conns[idx], nil
}

// AllConn : return all connection
func (c *Cluster) AllConn() ([]redis.Conn, error) {
	return c.conns, nil
}

func defaultConnSelectHandler(s string) (int, error) {
	return 0, nil
}
