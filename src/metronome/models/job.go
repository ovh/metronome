package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"

	"github.com/ovh/metronome/src/metronome/kafka"
)

// Job is a task execution.
type Job struct {
	GUID    string `json:"guid"`
	UserID  string `json:"user_id"`
	At      int64  `json:"at"`
	Epsilon int64  `json:"epsilon"`
	URN     string `json:"URN"`
}

// ToKafka serialize a Job to Kafka.
func (j *Job) ToKafka() *sarama.ProducerMessage {
	return &sarama.ProducerMessage{
		Topic: kafka.TopicJobs(),
		Key:   sarama.StringEncoder(j.GUID),
		Value: sarama.StringEncoder(fmt.Sprintf("%v %v %v %v %v", j.GUID, j.UserID, j.At, j.Epsilon, j.URN)),
	}
}

// FromKafka unserialize a Job from Kafka.
func (j *Job) FromKafka(msg *sarama.ConsumerMessage) error {
	key := string(msg.Key)
	segs := strings.Split(string(msg.Value), " ")
	if len(segs) != 5 {
		return fmt.Errorf("unprocessable job(%v) - bad segments", key)
	}

	timestamp, err := strconv.ParseInt(segs[2], 0, 64)
	if err != nil {
		return fmt.Errorf("unprocessable job(%v) - bad timestamp", key)
	}

	epsilon, err := strconv.ParseInt(segs[3], 0, 64)
	if err != nil {
		return fmt.Errorf("unprocessable job(%v) - bad epsilon", key)
	}

	j.GUID = key
	j.UserID = segs[1]
	j.At = timestamp
	j.Epsilon = epsilon
	j.URN = segs[4]

	return nil
}

// ToJSON serialize a Task as JSON.
func (j *Job) ToJSON() string {
	out, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}

	return string(out)
}
