package queue

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueue(t *testing.T) {
	q := New()
	assert.Equal(t, true, q.Empty(), "Empty queue")

	q.Push("alpha")
	q.Push("beta")

	assert.Equal(t, 2, q.Size(), "Size 2")
	assert.Equal(t, false, q.Empty(), "Non empty")
	assert.Equal(t, "alpha", q.Top(), "Top")
	assert.Equal(t, "alpha", q.Pop(), "Pop")
	assert.Equal(t, 1, q.Size(), "Pop")

	assert.Equal(t, "beta", q.Pop(), "Pop")
	assert.Equal(t, 0, q.Size(), "Size")
	assert.Equal(t, true, q.Empty(), "Empty")
}
