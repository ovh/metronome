package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"

	"github.com/ovh/metronome/src/metronome/kafka"
)

// Job is a task execution.
type Job struct {
	GUID    string                 `json:"guid"`
	UserID  string                 `json:"user_id"`
	At      int64                  `json:"at"`
	Epsilon int64                  `json:"epsilon"`
	URN     string                 `json:"URN"`
	Payload map[string]interface{} `json:"payload"`
}

// ToKafka serialize a Job to Kafka.
func (j *Job) ToKafka() *sarama.ProducerMessage {
	payloadBytes, err := json.Marshal(j.Payload)
	if err != nil {
		log.WithError(err).Warn("Cannot marshall job payload")
		payloadBytes = []byte("{}")
	}
	p := base64.StdEncoding.EncodeToString(payloadBytes)

	return &sarama.ProducerMessage{
		Topic: kafka.TopicJobs(),
		Key:   sarama.StringEncoder(j.GUID),
		Value: sarama.StringEncoder(fmt.Sprintf("%v %v %v %v %v %v", j.GUID, j.UserID, j.At, j.Epsilon, j.URN, p)),
	}
}

// FromKafka unserialize a Job from Kafka.
func (j *Job) FromKafka(msg *sarama.ConsumerMessage) error {
	key := string(msg.Key)
	segs := strings.Split(string(msg.Value), " ")
	if len(segs) != 6 {
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

	var payload []byte
	payload, err = base64.StdEncoding.DecodeString(segs[5])
	if err != nil {
		return fmt.Errorf("unprocessable job(%v) - bad payload (not base64)", key)
	}
	err = json.Unmarshal(payload, &j.Payload)
	if err != nil {
		return fmt.Errorf("unprocessable job(%v) - bad payload (not map string-interface)", key)
	}

	j.GUID = key
	j.UserID = segs[1]
	j.At = timestamp
	j.Epsilon = epsilon
	j.URN = segs[4]

	return nil
}

// ToJSON serialize a Task as JSON.
func (j *Job) ToJSON() ([]byte, error) {
	out, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return out, nil
}
