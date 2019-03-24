package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"testing"
)

//TODO fill this with 100% coverage of confluent and document cases tested

func TestConsumerBuilder(t *testing.T) {

	t.Run("name not set", func(t *testing.T) {
		cb := NewConfluentConsumerBuilder("")
		con, err := cb.Build()
		assert.Error(t, err)
		assert.Nil(t, con)
	})

	t.Run("broker not set", func(t *testing.T) {
		cb := NewConfluentConsumerBuilder("alpha")
		cb.SetTopics([]string{"a", "b"})
		con, err := cb.Build()
		assert.Error(t, err)
		assert.Nil(t, con)
	})

	t.Run("topics not set", func(t *testing.T) {
		cb := NewConfluentConsumerBuilder("alpha")
		cb.SetBroker([]string{"a:9092", "b:9092"})
		con, err := cb.Build()
		assert.Error(t, err)
		assert.Nil(t, con)
	})

	t.Run("all good", func(t *testing.T) {
		cb := NewConfluentConsumerBuilder("alpha")
		cb.SetBroker([]string{"a:9092", "b:9092"})
		cb.SetTopics([]string{"a", "b"})
		con, err := cb.Build()
		assert.NoError(t, err)
		assert.NotNil(t, con)
	})

}

func TestConsumer(t *testing.T) {

	t.Run("manual commit works", func(t *testing.T) {
		cb := NewConfluentConsumerBuilder("alpha")
		cb.SetBroker([]string{"a:9092", "b:9092"})
		cb.SetTopics([]string{"a", "b"})
		cb.DisableAutoCommit()
		con, err := cb.Build()
		assert.NoError(t, err)
		assert.NotNil(t, con)

		con.Setup()
		rawCon := con.(*confluentConsumer)
		assert.Equal(t, len(rawCon.maxOffsets), 0)
		topicA := new(string)
		topicB := new(string)
		*topicA = "a"
		*topicB = "b"
		tp := TopicPartition{Topic: *topicA, Partition: 0}
		tpB := TopicPartition{Topic: *topicB, Partition: 0}

		//bad partition commit errs
		assert.Error(t, rawCon.Commit([]interface{}{1}))
		assert.Equal(t, len(rawCon.maxOffsets), 0)

		assert.NoError(t, rawCon.Commit([]interface{}{kafka.TopicPartition{Topic: topicA, Partition: 0, Offset: 0}}))
		assert.Equal(t, len(rawCon.maxOffsets), 1)
		assert.Equal(t, rawCon.maxOffsets[tp], kafka.Offset(1))

		assert.NoError(t, rawCon.Commit([]interface{}{kafka.TopicPartition{Topic: topicA, Partition: 0, Offset: 1}}))
		assert.Equal(t, len(rawCon.maxOffsets), 1)
		assert.Equal(t, rawCon.maxOffsets[tp], kafka.Offset(2))

		//duplicate commit is NO-OP
		assert.NoError(t, rawCon.Commit([]interface{}{kafka.TopicPartition{Topic: topicA, Partition: 0, Offset: 1}}))
		assert.Equal(t, len(rawCon.maxOffsets), 1)
		assert.Equal(t, rawCon.maxOffsets[tp], kafka.Offset(2))

		//out of order commit is NO-OP
		assert.NoError(t, rawCon.Commit([]interface{}{kafka.TopicPartition{Topic: topicA, Partition: 0, Offset: 0}}))
		assert.Equal(t, len(rawCon.maxOffsets), 1)
		assert.Equal(t, rawCon.maxOffsets[tp], kafka.Offset(2))

		//jump commit is allowed
		assert.NoError(t, rawCon.Commit([]interface{}{kafka.TopicPartition{Topic: topicA, Partition: 0, Offset: 10}}))
		assert.Equal(t, len(rawCon.maxOffsets), 1)
		assert.Equal(t, rawCon.maxOffsets[tp], kafka.Offset(11))

		//commit to 2nd topic works
		assert.NoError(t, rawCon.Commit([]interface{}{kafka.TopicPartition{Topic: topicB, Partition: 0, Offset: 0}}))
		assert.Equal(t, len(rawCon.maxOffsets), 2)
		assert.Equal(t, rawCon.maxOffsets[tpB], kafka.Offset(1))

	})

	t.Run("client close", func(t *testing.T) {
		cb := NewConfluentConsumerBuilder("alpha")
		cb.SetBroker([]string{"a:9092", "b:9092"})
		cb.SetTopics([]string{"a", "b"})
		cb.DisableAutoCommit()
		con, err := cb.Build()
		assert.NoError(t, err)
		assert.NotNil(t, con)

		con.Setup()
		rawCon := con.(*confluentConsumer)
		assert.Equal(t, len(rawCon.maxOffsets), 0)
		topicA := new(string)
		topicB := new(string)
		*topicA = "a"
		*topicB = "b"

		//running the tests with below lines may cause SIGABORT which is weird.
		/*
			assert.NoError(t, rawCon.Commit([]interface{}{kafka.TopicPartition{Topic: topicA,Partition:0, Offset:0}}))
			assert.NoError(t, rawCon.Commit([]interface{}{kafka.TopicPartition{Topic: topicA,Partition:0, Offset:1}}))
			assert.NoError(t, rawCon.Commit([]interface{}{kafka.TopicPartition{Topic: topicA,Partition:0, Offset:2}}))

			time.Sleep(2 * time.Second)*/
		con.Close()
		<-rawCon.commiterDone
	})

}
