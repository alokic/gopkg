package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Msg struct {
	Data   []byte
	offset interface{}
}

func (m *Msg) Offset() interface{} {
	return m.offset
}

//Should we check for data as JsonMessage??
func decode(m *kafka.Message) (*Msg, error) {
	msg := &Msg{}
	msg.Data = m.Value
	msg.offset = m.TopicPartition
	return msg, nil
}
