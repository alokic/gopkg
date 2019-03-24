package redis

import (
	"errors"
	"os"

	"github.com/alokic/gopkg/cache"
	farm "github.com/alokic/gopkg/redisfarm"
)

var (
	ErrRedisUrlNotFound = errors.New("redis server url not found")
)

type Logger interface {
	Println(...interface{})
}

func New(redisServer, app, prefix string, logger Logger) cache.Cache {
	if redisServer == "" {
		logger.Println(ErrRedisUrlNotFound)
		os.Exit(1)
	}

	// build cluster
	cl, _ := farm.
		NewClusterBuilder().
		SetMaxIdleConns(6).
		SetMaxActiveConns(6).
		SetServers([]string{redisServer}).
		Build()

	// build farm
	f, _ := farm.
		NewBuilder().
		SetCluster([]*farm.Cluster{cl}).
		Build()

	// build cache
	c, _ := NewCacheBuilder().
		SetFarm(f).
		SetApp(app).
		SetPrefix(prefix).
		Build()

	return c
}
