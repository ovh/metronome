package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"

	"github.com/ovh/metronome/src/metronome/core"
	"github.com/ovh/metronome/src/metronome/kafka"
)

// Task holds task attributes.
type Task struct {
	GUID      string                 `json:"guid" sql:"guid,pk"`
	ID        string                 `json:"id" sql:"id"`
	UserID    string                 `json:"user_id"`
	Name      string                 `json:"name"`
	Schedule  string                 `json:"schedule"`
	URN       string                 `json:"URN"`
	Payload   map[string]interface{} `json:"payload" sql:",notnull"`
	CreatedAt time.Time              `json:"created_at"`
}

// Tasks is a Task list
type Tasks []Task

// ToKafka serialize a Task to Kafka.
func (t *Task) ToKafka() *sarama.ProducerMessage {
	if len(t.GUID) == 0 {
		t.GUID = core.Sha256(t.UserID + t.ID)
	}

	pBytes, err := json.Marshal(t.Payload)
	if err != nil {
		log.WithError(err).Warn("Cannot marshall Task payload")
		pBytes = []byte("{}")
	}
	p := base64.StdEncoding.EncodeToString(pBytes)

	return &sarama.ProducerMessage{
		Topic: kafka.TopicTasks(),
		Key:   sarama.StringEncoder(t.GUID),
		Value: sarama.StringEncoder(fmt.Sprintf("%v %v %v %v %v %v %v", t.UserID, t.ID, t.Schedule, t.URN, url.QueryEscape(t.Name), t.CreatedAt.Unix(), p)),
	}
}

// FromKafka unserialize a Task from Kafka.
func (t *Task) FromKafka(msg *sarama.ConsumerMessage) error {
	key := string(msg.Key)
	segs := strings.Split(string(msg.Value), " ")
	if len(segs) != 7 {
		log.Infof("segments: %+v %+v", segs, len(segs))
		return fmt.Errorf("unprocessable task(%v) - bad segments", key)
	}

	name, err := url.QueryUnescape(segs[4])
	if err != nil {
		return fmt.Errorf("unprocessable task(%v) - bad name", key)
	}

	timestamp, err := strconv.Atoi(segs[5])
	if err != nil {
		return fmt.Errorf("unprocessable task(%v) - bad timestamp", key)
	}

	payload, err := base64.StdEncoding.DecodeString(segs[6])
	if err != nil {
		return fmt.Errorf("unprocessable task(%v) - Bad payload (not Base64)", key)
	}
	err = json.Unmarshal(payload, &t.Payload)
	if err != nil {
		return fmt.Errorf("unprocessable task(%v) - Bad payload (not map string-interface)", key)
	}

	t.GUID = key
	t.UserID = segs[0]
	t.ID = segs[1]
	t.Schedule = segs[2]
	t.URN = segs[3]
	t.Name = name
	t.CreatedAt = time.Unix(int64(timestamp), 0)

	return nil
}

// ToJSON serialize a Task as JSON.
func (t *Task) ToJSON() ([]byte, error) {
	out, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return out, nil
}
