package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"

	"github.com/ovh/metronome/src/metronome/core"
	"github.com/ovh/metronome/src/metronome/kafka"
)

const (
	// Success task end success
	Success = iota
	// Failed task failed
	Failed
	// Expired task not performed within epsilon time frame
	Expired
)

// State is a state of a task execution.
type State struct {
	ID       string `json:"id"`
	TaskGUID string `json:"taskGUID"`
	UserID   string `json:"userID"`
	At       int64  `json:"at"`
	DoneAt   int64  `json:"doneAt"`
	Duration int64  `json:"duration"`
	URN      string `json:"URN"`
	State    int64  `json:"state"`
}

// States is a State array
type States []State

// ToKafka serialize a State to Kafka.
func (s *State) ToKafka() *sarama.ProducerMessage {
	if len(s.ID) == 0 {
		s.ID = core.Sha256(s.TaskGUID + strconv.FormatInt(s.At, 10))
	}
	return &sarama.ProducerMessage{
		Topic: kafka.TopicStates(),
		Key:   sarama.StringEncoder(s.ID),
		Value: sarama.StringEncoder(fmt.Sprintf("%v %v %v %v %v %v %v", s.TaskGUID, s.UserID, s.At, s.URN, s.DoneAt, s.Duration, s.State)),
	}
}

// FromKafka unserialize a State from Kafka.
func (s *State) FromKafka(msg *sarama.ConsumerMessage) error {
	key := string(msg.Key)
	segs := strings.Split(string(msg.Value), " ")
	if len(segs) != 7 {
		return fmt.Errorf("unprocessable state(%v) - bad segments", key)
	}

	at, err := strconv.ParseInt(segs[2], 0, 64)
	if err != nil {
		return fmt.Errorf("unprocessable state(%v) - bad at", key)
	}

	doneAt, err := strconv.ParseInt(segs[4], 0, 64)
	if err != nil {
		return fmt.Errorf("unprocessable state(%v) - bad done at", key)
	}

	duration, err := strconv.ParseInt(segs[5], 0, 64)
	if err != nil {
		return fmt.Errorf("unprocessable state(%v) - bad duration", key)
	}

	state, err := strconv.ParseInt(segs[6], 0, 64)
	if err != nil {
		return fmt.Errorf("unprocessable state(%v) - bad state", key)
	}

	s.ID = key
	s.TaskGUID = segs[0]
	s.UserID = segs[1]
	s.At = at
	s.DoneAt = doneAt
	s.Duration = duration
	s.URN = segs[3]
	s.State = state

	return nil
}

// ToJSON serialize a State as JSON.
func (s *State) ToJSON() ([]byte, error) {
	out, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// FromJSON unserialize a State from JSON.
func (s *State) FromJSON(in []byte) error {
	if len(in) == 0 {
		s.State = -1
		return nil
	}

	return json.Unmarshal(in, &s)
}
