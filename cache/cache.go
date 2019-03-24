package cache

import "context"

// Item :
type Item struct {
	Key string
	Val []byte
	TTL int64 // in seconds
}

// Cache :
type Cache interface {

	// Put : put a record. It can also update the existing key.. upsert operation
	//   Item: record data
	Put(context.Context, *Item) (interface{}, error)

	// Get : get a job based on ID
	//   string: record id
	Get(context.Context, string) (*Item, error)

	// Delete : get a job based on ID
	//   string: record id
	Delete(context.Context, string) error

	// MultiGet : Get multiple keys
	//   Maximum of maxCacheListSize is accepted
	MultiGet(context.Context, []string) ([]*Item, error)

	// MultiDelete : MultiDelete multiple keys
	//   Maximum of maxCacheListSize is accepted
	MultiDelete(context.Context, []string) error
}
