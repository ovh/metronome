package routines

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	saramaC "github.com/d33d33/sarama-cluster"
	"github.com/d33d33/viper" // FIXME https://github.com/spf13/viper/pull/285

	"github.com/runabove/metronome/src/metronome/constants"
	"github.com/runabove/metronome/src/metronome/models"
)

// taskConsumer handle the internal states of the consumer
type taskConsumer struct {
	client   *saramaC.Client
	consumer *saramaC.Consumer
	drained  bool
	drainMux sync.Mutex
	tasks    chan models.Task
}

// Return a new task consumer
func NewTaskComsumer() (*taskConsumer, error) {
	brokers := viper.GetStringSlice("kafka.brokers")

	config := saramaC.NewConfig()
	config.ClientID = "metronome-scheduler"
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := saramaC.NewClient(brokers, config)
	if err != nil {
		return nil, err
	}

	consumer, err := saramaC.NewConsumerFromClient(client, "schedulers", []string{constants.KAFKA_TOPIC_TASKS})
	if err != nil {
		return nil, err
	}

	tc := &taskConsumer{
		client:   client,
		consumer: consumer,
		tasks:    make(chan models.Task),
	}

	hwm := tc.highWaterMarks()
	offsets := make(map[int32]int64)
	messages := 0

	tc.drainMux.Lock()

	// Progress display
	ticker := time.NewTicker(500 * time.Millisecond)
	if log.GetLevel() == log.DebugLevel {
		go func() {
			for {
				select {
				case <-ticker.C:
					log.WithField("count", messages).Debug("Loading tasks")
				}
			}
		}()
	}

	// Drain timeout
	go func() {
		select {
		case <-time.After(700 * time.Millisecond):
			if !tc.drained {
				tc.drained = true
				ticker.Stop()
				tc.drainMux.Unlock()
			}
		}
	}()

	go func() {
		for {
			select {
			case msg, ok := <-consumer.Messages():
				if !ok { // shuting down
					return
				}
				messages++
				tc.handleMsg(msg)

				offsets[msg.Partition] = msg.Offset
				if !tc.drained && tc.isDrained(hwm, offsets) {
					tc.drained = true
					ticker.Stop()
					tc.drainMux.Unlock()
				}
			}
		}
	}()

	return tc, nil
}

// Incomming task channel
func (tc *taskConsumer) Tasks() <-chan models.Task {
	return tc.tasks
}

// Wait for consumer to EOF partitions
func (tc *taskConsumer) WaitForDrain() {
	if tc.drained {
		return
	}
	tc.drainMux.Lock()
	tc.drainMux.Unlock()
}

// Close the task consumer
func (tc *taskConsumer) Close() (err error) {
	if e := tc.consumer.Close(); e != nil {
		err = e
	}
	if e := tc.client.Close(); e != nil {
		err = e
	}
	return
}

// Handle incomming messages
func (tc *taskConsumer) handleMsg(msg *sarama.ConsumerMessage) {
	var t models.Task
	if err := t.FromKafka(msg); err != nil {
		log.Error(err)
		return
	}
	log.Debugf("Task received: %v", t.ToJSON())
	tc.tasks <- t
}

// Retrive highWaterMarks for each partition
func (tc *taskConsumer) highWaterMarks() map[int32]int64 {
	res := make(map[int32]int64)

	parts, err := tc.client.Partitions(constants.KAFKA_TOPIC_TASKS)
	if err != nil {
		log.Panic(err)
	}

	for p := range parts {
		i, err := tc.client.GetOffset(constants.KAFKA_TOPIC_TASKS, int32(p), sarama.OffsetNewest)
		if err != nil {
			log.Panic(err)
		}

		res[int32(p)] = i
	}

	return res
}

// Check if consumer reach EOF on all the partitions
func (tc *taskConsumer) isDrained(hwm, offsets map[int32]int64) bool {
	subs := tc.consumer.Subscriptions()[constants.KAFKA_TOPIC_TASKS]

	for partition := range subs {
		part := int32(partition)
		if _, ok := hwm[part]; !ok {
			log.Panicf("Missing HighWaterMarks for partition %v", part)
		}
		if hwm[part] == 0 {
			continue
		}
		// No message received for partiton
		if _, ok := offsets[part]; !ok {
			return false
		}
		// Check offset
		if (offsets[part] + 1) < hwm[part] {
			return false
		}
	}

	return true
}
