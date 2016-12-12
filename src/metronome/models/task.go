package models

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/sarama"

	"github.com/runabove/metronome/src/api/core"
	"github.com/runabove/metronome/src/metronome/constants"
)

type Task struct {
	GUID      string    `json:"-",sql:"GUID,pk"`
	Id        string    `json:"id",sql:"-"`
	UserId    string    `json:"user_id"`
	Name      string    `json:"name"`
	Schedule  string    `json:"schedule"`
	URN       string    `json:"URN"`
	CreatedAt time.Time `json:"created_at"`
}
type Tasks []Task

func (t *Task) ToKafka() *sarama.ProducerMessage {
	if len(t.GUID) == 0 {
		t.GUID = core.Sha256(t.UserId + t.Id)
	}

	return &sarama.ProducerMessage{
		Topic: constants.KAFKA_TOPIC_TASKS,
		Key:   sarama.StringEncoder(t.GUID),
		Value: sarama.StringEncoder(fmt.Sprintf("%v %v %v %v %v %v", t.UserId, t.Id, t.Schedule, t.URN, url.QueryEscape(t.Name), t.CreatedAt.Unix())),
	}
}

func (t *Task) FromKafka(msg *sarama.ConsumerMessage) error {
	key := string(msg.Key)
	segs := strings.Split(string(msg.Value), " ")
	if len(segs) != 6 {
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

	t.GUID = key
	t.UserId = segs[0]
	t.Id = segs[1]
	t.Schedule = segs[2]
	t.URN = segs[3]
	t.Name = name
	t.CreatedAt = time.Unix(int64(timestamp), 0)

	return nil
}

func (t *Task) ToJSON() string {
	out, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}

	return string(out)
}
