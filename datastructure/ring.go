package lib

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
)

//Ring is circular buffer. Not thread safe.
type Ring struct {
	size  int
	arr   []int64
	front int
	rear  int
	//mu sync.Mutex
}

//NewRing for getting new ring. Assumes that data is enqueued in monotonic order.
func NewRing(sz int) *Ring {
	if sz == 0 {
		return nil
	}
	return &Ring{
		size:  sz,
		front: 0,
		rear:  -1,
		arr:   make([]int64, sz),
	}
}

//String rep for ring.
func (r Ring) String() string {
	vals := "["
	if r.rear != -1 {
		for i := r.front; i != r.rear; i = (i + 1) % r.size {
			vals = vals + strconv.FormatInt(r.arr[i], 10) + ", "
		}
		vals = vals + strconv.FormatInt(r.arr[r.rear], 10)
	}
	vals = vals + "]"
	return fmt.Sprintf("\nRing {\n front: %v,\n rear: %v,\n size: %v,\n vals: %v\n}", r.front, r.rear, r.size, vals)
}

//Full checks if ring buffer is full.
func (r *Ring) Full() bool {
	if r.rear != -1 && ((r.front == 0 && r.rear == r.size-1) || ((r.rear+1)%r.size == r.front)) {
		return true
	}
	return false
}

//Empty check for ring.
func (r *Ring) Empty() bool {
	if r.rear == -1 {
		return true
	}
	return false
}

//Len returns len of ring.
func (r *Ring) Len() int {
	if r.rear == -1 {
		return 0
	}
	if r.rear >= r.front {
		return r.rear - r.front + 1
	}
	return r.size - r.front + r.rear + 1
}

//Enqueue queues to ring.
func (r *Ring) Enqueue(val int64) error {
	if r.Full() {
		return errors.New("queue full")
	}
	r.rear = (r.rear + 1) % r.size
	r.arr[r.rear] = val
	return nil
}

//Dequeue removes from ring.
func (r *Ring) Dequeue() (int64, error) {
	if r.Empty() {
		return 0, errors.New("queue empty")
	}
	val := r.arr[r.front]
	if r.front == r.rear {
		r.front = 0
		r.rear = -1
	} else {
		r.front = (r.front + 1) % r.size
	}
	return val, nil
}

//TrimTo trims ring. call in case of Full.
func (r *Ring) TrimTo(val int64) {
	if !r.Empty() {
		if val > r.arr[r.rear] {
			r.front = 0
			r.rear = -1
		} else if r.arr[r.front] < val {
			//TODO implement binary search
			i := r.front
			for ; i != r.rear && r.arr[i] < val; i = (i + 1) % r.size {
			}
			r.front = i
		}
	}
}
