// Package gid is unique id generator based on twitter/snowflake and sonyflake
// 39 bits for time in units of 10 msec
//  8 bits for a sequence number
// 16 bits for a machine id
package gid

import (
	"fmt"
	"hash/crc32"
	"math/rand"
	"net"
	"sync"
	"time"
	//"bytes"
	//"encoding/binary"
)

const (
	nano = 1000 * 1000
)

const (
	WorkerIdBits = 10              // worker id
	MaxWorkerId  = -1 ^ (-1 << 10) // worker id mask
	SequenceBits = 12              // sequence
	MaxSequence  = -1 ^ (-1 << 12) //sequence mask
)

var (
	Since int64 = time.Date(2012, 1, 0, 0, 0, 0, 0, time.UTC).UnixNano() / nano
	gid   *GID
	mu    sync.Mutex
)

type GID struct {
	lastTimestamp uint64
	workerId      uint32
	sequence      uint32
}

func Get() (uint64, error) {
	mu.Lock()
	defer mu.Unlock()

	if gid == nil {
		gid, _ = defaultGID()
	}
	return gid.next()
}

func (gi *GID) uint64() uint64 {
	return (gi.lastTimestamp << (WorkerIdBits + SequenceBits)) |
		(uint64(gi.workerId) << SequenceBits) |
		(uint64(gi.sequence))
}

func (gi *GID) next() (uint64, error) {
	ts := timestamp()
	if ts == gi.lastTimestamp {
		gi.sequence = (gi.sequence + 1) & MaxSequence
		if gi.sequence == 0 {
			ts = tilNextMillis(ts)
		}
	} else {
		gi.sequence = 0
	}

	if ts < gi.lastTimestamp {
		return 0, fmt.Errorf("Invalid timestamp: %v - precedes %v", ts, gi)
	}
	gi.lastTimestamp = ts
	return gi.uint64(), nil
}

func defaultGID() (*GID, error) {
	return newGID(defaultWorkId())
}

func newGID(workerId uint32) (*GID, error) {
	if workerId < 0 || workerId > MaxWorkerId {
		return nil, fmt.Errorf("Worker id %v is invalid", workerId)
	}
	return &GID{workerId: workerId}, nil
}

func timestamp() uint64 {
	return uint64(time.Now().UnixNano()/nano - Since)
}

func tilNextMillis(ts uint64) uint64 {
	i := timestamp()
	for i <= ts {
		i = timestamp()
	}
	return i
}

func defaultWorkId() uint32 {
	var id uint32
	ift, err := net.Interfaces()
	if err != nil {
		rand.Seed(time.Now().UnixNano())
		id = rand.Uint32() % MaxWorkerId
	} else {
		h := crc32.NewIEEE()
		for _, value := range ift {
			h.Write(value.HardwareAddr)
		}
		//randomization
		var uid uint64
		h.Write([]byte(fmt.Sprint(&uid)))

		id = h.Sum32() % MaxWorkerId
	}
	return id & MaxWorkerId
}
