package lib

import (
	"github.com/alokic/gopkg/datastructure"
	"sync"
	"time"
)

const (
	noRateLimit = -1
	nano        = 1e9
)

//RateLimiter interface for ratelimiting.
type RateLimiter interface {
	SetRateLimit(n int)
	Allowed() bool
}

type rateLimiter struct {
	limit int //events per second
	ring  *lib.Ring
	mu    sync.RWMutex
}

//NewRateLimiter gives new ratelimiter instance.
func NewRateLimiter(limit int) RateLimiter {
	r := new(rateLimiter)
	r.setLimit(limit)
	return r
}

//SetRateLimit is for setting rate limit.
func (r *rateLimiter) SetRateLimit(limit int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.setLimit(limit)
}

func (r *rateLimiter) setLimit(limit int) {
	if r.limit != limit {
		r.limit = limit
		r.ring = nil

		if limit > 0 {
			r.ring = lib.NewRing(limit)
		} else if limit < 0 {
			r.limit = noRateLimit
		}
	}
}

//Allowed checks if its allowed.
func (r *rateLimiter) Allowed() bool {
	allowed := true

	if r.limit == 0 {
		allowed = false
	} else if r.limit != noRateLimit {
		curr := time.Now().UnixNano()
		if r.ring.Full() == true {
			r.ring.TrimTo(curr - nano)
		}
		if r.ring.Full() == true {
			allowed = false
		} else {
			r.ring.Enqueue(curr)
		}
	}
	return allowed
}
