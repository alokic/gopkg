package kafka

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSaramaProducer(t *testing.T) {
	_, err := NewSaramaProducer("test", []string{})
	assert.Error(t, err)

	_, err = NewSaramaProducer("test", []string{"lalaland"})
	assert.Error(t, err)
}
