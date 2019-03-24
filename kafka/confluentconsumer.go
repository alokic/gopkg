package kafka

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
)

const (
	offsetReset  = "earliest"
	commitPeriod = 5 * time.Second
	drainPeriod  = 2 * time.Second
	//minQueued    = 100
	//maxKbQueued  = 10 * 1024 //higher priority
)

type TopicPartition struct {
	Topic     string
	Partition int32
}

type confluentConsumer struct {
	name         string
	brokers      []string
	topics       []string
	autoCommit   bool
	config       kafka.ConfigMap
	consumer     *kafka.Consumer
	mu           sync.RWMutex
	maxOffsets   map[TopicPartition]kafka.Offset
	commiterDone chan struct{}
}

type ConsumerBuilder struct {
	c *confluentConsumer
}

func NewConfluentConsumerBuilder(name string) *ConsumerBuilder {
	return &ConsumerBuilder{
		c: &confluentConsumer{
			name:       name,
			autoCommit: true,
			mu:         sync.RWMutex{},
			maxOffsets: make(map[TopicPartition]kafka.Offset),
			config:     kafka.ConfigMap{}, //default empty
		},
	}
}

func (cb *ConsumerBuilder) SetBroker(hosts []string) {
	cb.c.brokers = hosts
}

func (cb *ConsumerBuilder) SetTopics(topics []string) {
	saneTopics := []string{}
	for _, t := range topics {
		if t != "" {
			saneTopics = append(saneTopics, t)
		}
	}
	cb.c.topics = saneTopics
}

func (cb *ConsumerBuilder) DisableAutoCommit() {
	cb.c.autoCommit = false
}

func (cb *ConsumerBuilder) SetConfig(cfg map[string]interface{}) {
	if cfg != nil {
		for k, v := range cfg {
			cb.c.config[k] = v
		}
	}
}

func (cb *ConsumerBuilder) Build() (KafkaConsumer, error) {

	c := cb.c

	if c.name == "" {
		return nil, errors.New("please set consumer name")
	}

	if len(c.topics) == 0 {
		return nil, errors.New("please set consumer topic(s)")
	}

	if len(c.brokers) == 0 {
		return nil, errors.New("please set the broker address")
	}

	hosts := strings.Join(c.brokers, ",")
	defConfig := kafka.ConfigMap{
		"bootstrap.servers":  hosts,
		"group.id":           cb.c.name,
		"enable.auto.commit": cb.c.autoCommit,
		//"queued.min.messages":        minQueued,
		//"queued.max.messages.kbytes": maxKbQueued,
		"auto.offset.reset": offsetReset,
	}

	//defConfig has high priority
	for k, v := range defConfig {
		cb.c.config[k] = v
	}
	kc, err := kafka.NewConsumer(&cb.c.config)

	if err != nil {
		return nil, err
	}

	err = kc.SubscribeTopics(c.topics, nil)
	if err != nil {
		return nil, err
	}
	c.consumer = kc

	return c, nil
}

//TODO call under sync.Once
func (c *confluentConsumer) Setup() error {
	if !c.autoCommit {
		c.commiterDone = make(chan struct{})
		go c.commiter()
	}

	return nil
}

//TODO make multiple calls to poll
//Poll timeout is in ms
func (c *confluentConsumer) Poll(timeout time.Duration) ([]Msg, error) {
	if c.consumer == nil {
		return nil, errors.New("attempt to poll on uninited consumer")
	}

	ev, err := c.consumer.ReadMessage(timeout)
	if err != nil {
		return nil, err
	}
	//ev is guaranteed to be non-nil now
	msg, err := decode(ev)
	if err != nil {
		//what do we do with corrupted message, right now we are dropping them.
		fmt.Printf("[ERROR] unable to decode message: %v, err: %v\n. Dropping.", ev, err)
		c.Commit([]interface{}{ev.TopicPartition})
		return nil, errors.Wrap(err, "unable to decode Msg")
	}

	return []Msg{*msg}, nil
}

func (c *confluentConsumer) Commit(offsets []interface{}) error {
	if c.consumer == nil {
		return errors.New("attempt to Commit on uninited consumer")
	}

	if !c.autoCommit {
		c.mu.Lock()
		defer c.mu.Unlock()

		kafkaOffsets, err := c.offsets(offsets)
		if err != nil {
			return err
		}
		for _, of := range kafkaOffsets {
			if of.Topic == nil {
				return errors.New(fmt.Sprintf("topic missing for offset: %v", of))
			}
			tp := TopicPartition{Topic: *of.Topic, Partition: of.Partition}
			old := c.maxOffsets[tp]
			tpOffset := max(old, of.Offset+1)
			c.maxOffsets[tp] = tpOffset
		}
	}
	return nil
}

//TODO call under sync.Once
func (c *confluentConsumer) Close() error {
	if c.consumer == nil {
		return errors.New("attempt to Close uninited consumer")
	}

	if !c.autoCommit {
		close(c.commiterDone)
		time.Sleep(drainPeriod)
	}
	return c.consumer.Close()
}

func (c *confluentConsumer) offsets(o []interface{}) ([]kafka.TopicPartition, error) {
	offsets := []kafka.TopicPartition{}

	for _, oi := range o {
		switch t := oi.(type) {
		case kafka.TopicPartition:
			offsets = append(offsets, t)
		default:
			return nil, errors.New("offset(s) is not []kafka.TopicPartition")
		}
	}
	return offsets, nil
}

func (c *confluentConsumer) commiter() {
	timer := time.NewTicker(commitPeriod)
	for {
		select {
		case <-c.commiterDone:
			c.commitOffsets()
			return
		case <-timer.C:
			c.commitOffsets()
		}
	}
}

func (c *confluentConsumer) commitOffsets() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.syncPartition()

	if len(c.maxOffsets) > 0 {
		offsets := []kafka.TopicPartition{}
		for tp, o := range c.maxOffsets {
			ktp := kafka.TopicPartition{}
			ktp.Topic = &tp.Topic
			ktp.Partition = tp.Partition
			ktp.Offset = o
			offsets = append(offsets, ktp)
		}
		success, err := c.consumer.CommitOffsets(offsets)
		if err != nil || len(success) == 0 {
			//a future successful commit would solve it. If we crash before that, then messages would be redelivered.
			fmt.Printf("error in commiting offset: %v, err: %v\n.", success, err)
		}
	}
}

func (c *confluentConsumer) syncPartition() {
	partitions, err := c.consumer.Assignment()

	if err == nil {
		kafkaPartitions := make(map[TopicPartition]struct{})
		for _, p := range partitions {
			kafkaPartitions[TopicPartition{Topic: *p.Topic, Partition: p.Partition}] = struct{}{}
		}

		for k := range c.maxOffsets {
			if _, ok := kafkaPartitions[k]; !ok {
				delete(c.maxOffsets, k)
			}
		}
	}
}

func max(a, b kafka.Offset) kafka.Offset {
	if a > b {
		return a
	}
	return b
}
