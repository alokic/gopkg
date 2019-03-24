package kafka

import (
	"encoding/json"
)

type KafkaProducer interface {
	Write([]json.RawMessage) error
	Close() error
}
