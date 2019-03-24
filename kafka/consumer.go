package kafka

import (
	"time"
)

type KafkaConsumer interface {
	Setup() error
	Poll(time.Duration) ([]Msg, error)
	Commit([]interface{}) error //pass commit handle e.g kafka.TopicPartition for confluent one
	Close() error
}
