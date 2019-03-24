package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/alokic/gopkg/cache"
	farm "github.com/alokic/gopkg/redisfarm"
	"github.com/alokic/gopkg/timeutils"
	"github.com/alokic/gopkg/typeutils"
)

var (
	cachePutScriptStr = `
		local keyTimestamp = 'KEYTIMESTAMP'
		local tms = redis.call("HGET", ARGV[1], keyTimestamp)
		local in_tms = ARGV[2] + 0
		local keyDeleted = 'KEYDELETED'

		if tms == false or in_tms > tonumber(tms) then
		    redis.call("HSET", ARGV[1], keyTimestamp, in_tms)
			redis.call("HSET", ARGV[1], ARGV[3], ARGV[4])
			redis.call("EXPIRE", ARGV[1], ARGV[5])
			redis.call("HDEL", ARGV[1], keyDeleted) -- delete 'keyDeleted' marker if deleted is put before Expiry
		end

		return tms
	`

	cacheGetScriptStr = `
		local tbl = {}

		if redis.call("EXISTS", ARGV[1]) == 0 then
			return tbl
		end

		local keyTTL = 'KEYTTL'
		local keyDeleted = 'KEYDELETED'
		local is_deleted = redis.call("HGET", ARGV[1], keyDeleted)

		if is_deleted == false then
		  redis.call("HSET", ARGV[1], keyTTL, redis.call("TTL", ARGV[1])) -- set ttl in map to return
			tbl = redis.call("HGETALL", ARGV[1])
			redis.call("HDEL", ARGV[1], keyTTL)  -- remove ttl from map as its not needed
		end

		return tbl
	`

	cacheDeleteScriptStr = `
		local tbl = {}

		if redis.call("EXISTS", ARGV[1]) == 0 then
			return tbl
		end

		local keyTTL = 'KEYTTL'
	    local keyTimestamp = 'KEYTIMESTAMP'
		local keyDeleted = 'KEYDELETED'
		local in_tms = ARGV[2] + 0

	    local tms = redis.call("HGET", ARGV[1], keyTimestamp)

		if tms == false or in_tms >= tonumber(tms) then
			redis.call("HSET", ARGV[1], keyDeleted, 1)

			redis.call("HSET", ARGV[1], keyTTL, redis.call("TTL", ARGV[1])) -- set ttl in map to return
			tbl = redis.call("HGETALL", ARGV[1])
			redis.call("HDEL", ARGV[1], keyTTL)  -- remove ttl from map as its not needed
		end

		return tbl
	`

	cacheMultiGetScriptStr = `
		local keyTTL = 'KEYTTL'
		local keyDeleted = 'KEYDELETED'

		local is_deleted
		local tbl = {}

    	for i = 1, #ARGV, 1 do
		  if redis.call("EXISTS", ARGV[i]) == 1 then
				if redis.call("HGET", ARGV[i], keyDeleted) == false then
					redis.call("HSET", ARGV[i], keyTTL, redis.call("TTL", ARGV[i])) -- set ttl in map to return
					tbl[#tbl+1] = redis.call("HGETALL", ARGV[i])
					redis.call("HDEL", ARGV[i], keyTTL)  -- remove ttl from map as its not needed
				end
			end
		end

		return tbl
	`
	cacheMultiDeleteScriptStr = `
		local keyTTL = 'KEYTTL'
		local keyTimestamp = 'KEYTIMESTAMP'
		local keyDeleted = 'KEYDELETED'
		local in_tms = ARGV[1] + 0

		local tms = 0
		local tbl = {}

		for i = 2, #ARGV, 1 do
			tms = redis.call("HGET", ARGV[i], keyTimestamp)
			if tms == false or in_tms >= tonumber(tms) then
				redis.call("HSET", ARGV[i], keyDeleted, 1)

				redis.call("HSET", ARGV[i], keyTTL, redis.call("TTL", ARGV[i])) -- set ttl in map to return
				tbl[#tbl+1] = redis.call("HGETALL", ARGV[i])
				redis.call("HDEL", ARGV[i], keyTTL)  -- remove ttl from map as its not needed
			end
		end

		return tbl
	`
	cachePutScript         *redis.Script
	cacheGetScript         *redis.Script
	cacheDeleteScript      *redis.Script
	cacheMultiGetScript    *redis.Script
	cacheMultiDeleteScript *redis.Script

	cacheMaxListSize = 25

	keyTimestamp   = "_tm_"
	keyTTL         = "_ttl_"
	keyDeleted     = "_del_"
	keyspacePrefix = "h"

	// only for test simulation
	timeLag = int64(0)
)

// CacheBuilder :
type CacheBuilder struct {
	rs *redisCache
}

// redisCache :
type redisCache struct {
	farm     *farm.Farm
	prefix   string
	app      string
	keyspace string
	script   map[string]*redis.Script
}

// NewCacheBuilder :
func NewCacheBuilder() *CacheBuilder {
	return &CacheBuilder{rs: newCache()}
}

// SetFarm :
func (c *CacheBuilder) SetFarm(f *farm.Farm) *CacheBuilder {
	c.rs.farm = f
	return c
}

// SetApp :
func (c *CacheBuilder) SetApp(name string) *CacheBuilder {
	c.rs.app = name
	return c
}

// SetPrefix :
func (c *CacheBuilder) SetPrefix(name string) *CacheBuilder {
	c.rs.prefix = name
	return c
}

// Build :
func (c *CacheBuilder) Build() (cache.Cache, error) {
	return c.rs.build()
}

func newCache() *redisCache {
	return &redisCache{script: make(map[string]*redis.Script)}
}

func (c *redisCache) build() (cache.Cache, error) {
	c.keyspace = fmt.Sprintf("%s:%s:%s", keyspacePrefix, c.app, c.prefix)
	if c.prefix != "" {
		c.keyspace = fmt.Sprintf("%s:%s", c.keyspace, c.prefix)
	}

	err := c.loadScript()
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Put : put a item
func (c *redisCache) Put(ctx context.Context, r *cache.Item) (interface{}, error) {
	formattedKey := c.formatKey(r.Key)

	// get connection based on formattedKey
	conn := c.farm.GetConn(formattedKey)

	arr := []interface{}{formattedKey, unixTime(), r.Key, r.Val, r.TTL}

	reply, err := cachePutScript.Do(conn, arr...)

	// since its of type redis.Error, nil check fails. So always need to check for blank string
	if err.Error() != "" {
		return nil, errors.New(err.Error())
	}

	return reply, nil
}

// Get : get a item
func (c *redisCache) Get(ctx context.Context, key string) (*cache.Item, error) {
	formattedKey := c.formatKey(key)

	// get connection based on formattedKey
	conn := c.farm.GetConn(formattedKey)

	// splat the args..
	reply, err := cacheGetScript.Do(conn, formattedKey)

	// since its of type redis.Error, nil check fails. So always need to check for blank string
	if err.Error() != "" {
		return nil, errors.New(err.Error())
	}

	var items []*cache.Item

	for _, e := range reply.([]interface{}) {
		if len(e.([]interface{})) == 0 {
			continue
		}
		items = append(items, c.unmarshallItem(e.([]interface{})))
	}

	if len(items) > 0 {
		return items[0], nil
	}
	return nil, nil
}

// Delete : delete a item based on ID
func (c *redisCache) Delete(ctx context.Context, key string) error {
	formattedKey := c.formatKey(key)

	// get connection based on formattedKey
	conn := c.farm.GetConn(formattedKey)

	arr := []interface{}{formattedKey, unixTime()}
	// splat the args..

	_, err := cacheDeleteScript.Do(conn, arr...)

	// since its of type redis.Error, nil check fails. So always need to check for blank string
	if err.Error() != "" {
		return errors.New(err.Error())
	}

	return nil
}

// MultiGet : Get multiple keys
//   Maximum of cacheMaxListSize is accepted
func (c *redisCache) MultiGet(ctx context.Context, keys []string) ([]*cache.Item, error) {
	conn := c.farm.AllConn()

	if len(keys) > cacheMaxListSize {
		keys = keys[:cacheMaxListSize]
	}

	arr := []interface{}{}
	for _, v := range keys {
		arr = append(arr, c.formatKey(v))
	}

	// splat the args..
	reply, err := cacheMultiGetScript.Do(conn, arr...)

	// since its of type redis.Error, nil check fails. So always need to check for blank string
	if err.Error() != "" {
		return nil, errors.New(err.Error())
	}

	var items []*cache.Item
	for _, elements := range reply.([]interface{}) {
		for _, e := range elements.([]interface{}) {
			if len(e.([]interface{})) == 0 {
				continue
			}
			items = append(items, c.unmarshallItem(e.([]interface{})))
		}
	}

	return items, nil
}

// MultiDelete : Delete multiple keys
//   Maximum of cacheMaxListSize is accepted
func (c *redisCache) MultiDelete(ctx context.Context, keys []string) error {
	conn := c.farm.AllConn()

	if len(keys) > cacheMaxListSize {
		keys = keys[:cacheMaxListSize]
	}

	arr := []interface{}{unixTime()}
	for _, v := range keys {
		arr = append(arr, c.formatKey(v))
	}

	// splat the args..
	_, err := cacheMultiDeleteScript.Do(conn, arr...)

	// since its of type redis.Error, nil check fails. So always need to check for blank string
	if err.Error() != "" {
		return errors.New(err.Error())
	}
	return nil
}

// unmarshallItem :
//   arr: array, where key and value are ordered {key1, val1, key2, val2}
func (c *redisCache) unmarshallItem(arr []interface{}) *cache.Item {
	r := &cache.Item{}

	for i := 0; i < len(arr); {
		k := string(arr[i].([]byte))

		switch k {
		case keyTimestamp:
		case keyDeleted:
		case keyTTL:
			r.TTL = typeutils.ToInt64(arr[i+1].([]byte))
		default:
			r.Key = k
			r.Val = arr[i+1].([]byte)
		}
		i += 2
	}

	return r
}

func (c *redisCache) loadScript() error {
	cachePutScript = redis.NewScript(0, strings.
		NewReplacer("KEYTTL", keyTTL, "KEYTIMESTAMP", keyTimestamp, "KEYDELETED", keyDeleted).
		Replace(cachePutScriptStr))

	cacheGetScript = redis.NewScript(0, strings.
		NewReplacer("KEYDELETED", keyDeleted, "KEYTTL", keyTTL).
		Replace(cacheGetScriptStr))

	cacheDeleteScript = redis.NewScript(0, strings.
		NewReplacer("KEYTIMESTAMP", keyTimestamp, "KEYDELETED", keyDeleted, "KEYTTL", keyTTL).
		Replace(cacheDeleteScriptStr))

	cacheMultiGetScript = redis.NewScript(0, strings.
		NewReplacer("KEYDELETED", keyDeleted, "KEYTTL", keyTTL).
		Replace(cacheMultiGetScriptStr))

	cacheMultiDeleteScript = redis.NewScript(0, strings.
		NewReplacer("KEYDELETED", keyDeleted, "KEYTTL", keyTTL).
		Replace(cacheMultiDeleteScriptStr))

	return nil
}

func (c *redisCache) formatKey(key string) string {
	return fmt.Sprintf("%s:%s", c.keyspace, key)
}

// SetTimeDiffForTesting : clock adjustment is needed
// only for testinf
func SetTimeDiffForTesting(tms int64) {
	timeLag = tms
}

func unixTime() int64 {
	return timeutils.UnixTime() + timeLag
}
