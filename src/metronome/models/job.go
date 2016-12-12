package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"

	"github.com/runabove/metronome/src/metronome/constants"
)

// Job is a task execution.
type Job struct {
	GUID    string
	At      int64
	Epsilon int64
	URN     string
}

// ToKafka serialize a Job to Kafka.
func (j *Job) ToKafka() *sarama.ProducerMessage {
	return &sarama.ProducerMessage{
		Topic: constants.KafkaTopicJobs,
		Key:   sarama.StringEncoder(j.GUID),
		Value: sarama.StringEncoder(fmt.Sprintf("%v %v %v %v", j.GUID, j.At, j.Epsilon, j.URN)),
	}
}

// FromKafka unserialize a Job from Kafka.
func (j *Job) FromKafka(msg *sarama.ConsumerMessage) error {
	key := string(msg.Key)
	segs := strings.Split(string(msg.Value), " ")
	if len(segs) != 4 {
		return fmt.Errorf("unprocessable job(%v) - bad segments", key)
	}

	timestamp, err := strconv.ParseInt(segs[1], 0, 64)
	if err != nil {
		return fmt.Errorf("unprocessable job(%v) - bad timestamp", key)
	}

	epsilon, err := strconv.ParseInt(segs[2], 0, 64)
	if err != nil {
		return fmt.Errorf("unprocessable job(%v) - bad epsilon", key)
	}

	j.GUID = key
	j.At = timestamp
	j.Epsilon = epsilon
	j.URN = segs[3]

	return nil
}
