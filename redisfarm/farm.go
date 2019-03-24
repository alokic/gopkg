package redisfarm

import (
	"errors"

	"github.com/garyburd/redigo/redis"
)

var (
	// ErrFarmNoCluster : "No cluster exists"
	ErrFarmNoCluster = errors.New("No cluster exists")
)

// Builder :
type Builder struct {
	f *Farm
}

// Farm :
type Farm struct {
	clusters []*Cluster
}

// NewBuilder :
func NewBuilder() *Builder {
	return &Builder{f: newFarm()}
}

// SetCluster :
func (f *Builder) SetCluster(clusters []*Cluster) *Builder {
	f.f.clusters = clusters
	return f
}

// Build :
func (f *Builder) Build() (*Farm, error) {
	return f.f.build()
}

func newFarm() *Farm {
	return &Farm{}
}

func (f *Farm) build() (*Farm, error) {
	if len(f.clusters) == 0 {
		return nil, ErrFarmNoCluster
	}
	return f, nil
}

// GetConn : implements consistent hash
// TODO
func (f *Farm) GetConn(keyspace string) redis.Conn {
	mc := &multiConn{}

	for _, v := range f.clusters {
		conn, err := v.GetConn(keyspace)

		// TODO - what to do when not enough conns available
		if err != nil {
			continue
		}

		mc.SetConn(conn)

	}

	return mc.Build()
}

// AllConn :
func (f *Farm) AllConn() redis.Conn {
	mc := &multiConn{}

	for _, v := range f.clusters {
		conns, err := v.AllConn()

		// TODO - what to do when not enough conns available
		if err != nil {
			continue
		}

		for _, c := range conns {
			mc.SetConn(c)
		}
	}

	return mc.Build()
}
