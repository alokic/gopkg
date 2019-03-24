package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRing_NonZero(t *testing.T) {
	r := NewRing(3)
	assert.Equal(t, r.Empty(), true, "initially empty ring")
	assert.Equal(t, r.Full(), false, "initially non full ring")
	assert.Equal(t, r.Len(), 0, "initially len 0 ring")
	t.Log(r)

	assert.NoError(t, r.Enqueue(10), "enqueue on empty works")
	assert.Equal(t, r.Empty(), false, "ring of size 1")
	assert.Equal(t, r.Full(), false, "non full ring")
	assert.Equal(t, r.Len(), 1, "len 1 ring")
	assert.Equal(t, r.rear, 0, "rear is 0")
	assert.Equal(t, r.front, 0, "front is 0")
	t.Log(r)

	val, err := r.Dequeue()
	assert.NoError(t, err, "dequeue on empty works")
	assert.Equal(t, val, int64(10), "ring of size 1")
	assert.Equal(t, r.Full(), false, "non full ring")
	assert.Equal(t, r.Len(), 0, "len 0 ring")
	assert.Equal(t, r.rear, -1, "rear back to -1")
	assert.Equal(t, r.front, 0, "front back to 0")
	t.Log(r)

	assert.NoError(t, r.Enqueue(10), "enqueue on empty works")
	assert.NoError(t, r.Enqueue(11), "enqueue on empty works")
	assert.NoError(t, r.Enqueue(12), "enqueue on empty works")
	assert.Equal(t, r.Empty(), false, "ring of size 3")
	assert.Equal(t, r.Full(), true, "full ring")
	assert.Equal(t, r.Len(), 3, "len 3 ring")
	assert.Equal(t, r.rear, 2, "rear is 0")
	assert.Equal(t, r.front, 0, "front is 0")
	t.Log(r)

	//Enqueue errs on full
	assert.Error(t, r.Enqueue(13), "enqueue on full errs")
	t.Log(r)

	val, err = r.Dequeue()
	assert.NoError(t, err, "dequeue on full works")
	assert.Equal(t, val, int64(10), "ring val")
	assert.Equal(t, r.Full(), false, "non full ring")
	assert.Equal(t, r.Len(), 2, "len 2 ring")
	assert.Equal(t, r.rear, 2, "rear is 2")
	assert.Equal(t, r.front, 1, "front is 0")
	t.Log(r)

	//enqueue should work
	assert.NoError(t, r.Enqueue(13), "enqueue works")
	assert.Equal(t, r.Empty(), false, "ring of size 3")
	assert.Equal(t, r.Full(), true, "full ring")
	assert.Equal(t, r.Len(), 3, "len 3 ring")
	assert.Equal(t, r.rear, 0, "rear is 0")
	assert.Equal(t, r.front, 1, "front is 0")
	t.Log(r)

	assert.Error(t, r.Enqueue(14), "enqueue on full errs")
	t.Log(r)

	val, err = r.Dequeue()
	assert.NoError(t, err, "dequeue works")
	assert.Equal(t, val, int64(11), "ring val")
	val, err = r.Dequeue()
	assert.NoError(t, err, "dequeue works")
	assert.Equal(t, val, int64(12), "ring val")
	assert.Equal(t, r.rear, 0, "rear is 2")
	assert.Equal(t, r.front, 0, "front is 0")
	t.Log(r)

	val, err = r.Dequeue()
	assert.NoError(t, err, "dequeue works")
	assert.Equal(t, val, int64(13), "ring val")
	t.Log(r)

	assert.Equal(t, r.Full(), false, "non full ring")
	assert.Equal(t, r.Empty(), true, "non full ring")
	assert.Equal(t, r.Len(), 0, "len 2 ring")
	assert.Equal(t, r.rear, -1, "rear is reset")
	assert.Equal(t, r.front, 0, "front is reset")
	t.Log(r)

	val, err = r.Dequeue()
	assert.Error(t, err, "dequeue errs")
	t.Log(r)

	//Another round of tests from start
	r.Enqueue(1)
	r.Enqueue(2)
	r.Enqueue(3)
	assert.Equal(t, r.rear, 2, "rear is reset")
	assert.Equal(t, r.front, 0, "front is reset")
	t.Log(r)
	r.Dequeue()
	r.Dequeue()
	assert.Equal(t, r.rear, 2, "rear is reset")
	assert.Equal(t, r.front, 2, "front is reset")
	t.Log(r)
	r.Dequeue()
	assert.Equal(t, r.rear, -1, "rear is reset")
	assert.Equal(t, r.front, 0, "front is reset")
	t.Log(r)
}

func TestRing_TrimTo(t *testing.T) {
	tests := map[string]struct {
		r             Ring
		trim          int64
		expectedFront int
		expectedRear  int
	}{
		"trim on empty ring": {
			r:             Ring{size: 3, front: 0, rear: -1, arr: []int64{}},
			expectedFront: 0,
			expectedRear:  -1,
			trim:          1,
		},
		"trim on non full ring": {
			r:             Ring{size: 3, front: 0, rear: 1, arr: []int64{1, 5, 777777}},
			expectedFront: 1,
			expectedRear:  1,
			trim:          2,
		},
		"trim all on non full ring": {
			r:             Ring{size: 3, front: 0, rear: 1, arr: []int64{1, 5, 7777777}},
			expectedFront: 0,
			expectedRear:  -1,
			trim:          20,
		},
		"trim on full ring with val less than front": {
			r:             Ring{size: 3, front: 0, rear: 2, arr: []int64{1, 2, 3}},
			expectedFront: 0,
			expectedRear:  2,
			trim:          1,
		},
		"trim on full ring with val less than rear but > front": {
			r:             Ring{size: 3, front: 0, rear: 2, arr: []int64{1, 5, 30}},
			expectedFront: 1,
			expectedRear:  2,
			trim:          2,
		},
		"trim on full ring with val less than rear but > front: #2": {
			r:             Ring{size: 3, front: 0, rear: 2, arr: []int64{1, 5, 30}},
			expectedFront: 1,
			expectedRear:  2,
			trim:          5,
		},
		"trim on full ring with val less than rear but > front: #3": {
			r:             Ring{size: 3, front: 0, rear: 2, arr: []int64{1, 5, 30}},
			expectedFront: 2,
			expectedRear:  2,
			trim:          6,
		},
		"trim all": {
			r:             Ring{size: 3, front: 0, rear: 2, arr: []int64{1, 5, 30}},
			expectedFront: 0,
			expectedRear:  -1,
			trim:          60,
		},
		"trim on non full ring - pivoted": {
			r:             Ring{size: 3, front: 2, rear: 0, arr: []int64{5, 7777777, 1}},
			expectedFront: 0,
			expectedRear:  0,
			trim:          2,
		},
		"trim all on non full ring - pivoted": {
			r:             Ring{size: 3, front: 2, rear: 0, arr: []int64{5, 777777, 1}},
			expectedFront: 0,
			expectedRear:  -1,
			trim:          20,
		},
		"trim on full ring with val less than front - pivoted ring": {
			r:             Ring{size: 3, front: 2, rear: 1, arr: []int64{10, 20, 5}},
			expectedFront: 2,
			expectedRear:  1,
			trim:          1,
		},
		"trim on full ring with val less than rear but > front - pivoted ring": {
			r:             Ring{size: 3, front: 2, rear: 1, arr: []int64{5, 30, 1}},
			expectedFront: 0,
			expectedRear:  1,
			trim:          2,
		},
		"trim on full ring with val less than rear but > front - pivoted ring: #2": {
			r:             Ring{size: 3, front: 2, rear: 1, arr: []int64{5, 30, 1}},
			expectedFront: 0,
			expectedRear:  1,
			trim:          5,
		},
		"trim on full ring with val less than rear but > front - pivoted ring #3": {
			r:             Ring{size: 3, front: 2, rear: 1, arr: []int64{5, 30, 1}},
			expectedFront: 1,
			expectedRear:  1,
			trim:          6,
		},
		"trim all- pivoted ring": {
			r:             Ring{size: 3, front: 2, rear: 1, arr: []int64{5, 30, 1}},
			expectedFront: 0,
			expectedRear:  -1,
			trim:          60,
		},
	}
	for tName, tInfo := range tests {
		t.Log("Running test: ", tName)
		tInfo.r.TrimTo(tInfo.trim)
		assert.Equal(t, tInfo.expectedFront, tInfo.r.front)
		assert.Equal(t, tInfo.expectedRear, tInfo.r.rear)
	}
}
