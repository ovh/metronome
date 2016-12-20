package routines

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	saramaC "github.com/d33d33/sarama-cluster"
	"github.com/spf13/viper"

	"github.com/runabove/metronome/src/metronome/kafka"
	"github.com/runabove/metronome/src/metronome/models"
)

// TaskConsumer handle the internal states of the consumer
type TaskConsumer struct {
	client   *saramaC.Client
	consumer *saramaC.Consumer
	drained  bool
	drainWg  sync.WaitGroup
	// group tasks by partition
	partitions map[int]chan models.Task
	tasks      chan chan models.Task
	hwm        map[int32]int64
}

// NewTaskComsumer return a new task consumer
func NewTaskComsumer() (*TaskConsumer, error) {
	brokers := viper.GetStringSlice("kafka.brokers")

	config := saramaC.NewConfig()
	config.Config = *kafka.NewConfig()
	config.ClientID = "metronome-scheduler"
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Group.Return.Notifications = true

	client, err := saramaC.NewClient(brokers, config)
	if err != nil {
		return nil, err
	}

	consumer, err := saramaC.NewConsumerFromClient(client, kafka.GroupSchedulers(), []string{kafka.TopicTasks()})
	if err != nil {
		return nil, err
	}

	tc := &TaskConsumer{
		client:     client,
		consumer:   consumer,
		partitions: make(map[int]chan models.Task),
		tasks:      make(chan chan models.Task),
	}

	tc.hwm = <-tc.highWaterMarks()
	offsets := make(map[int32]int64)
	messages := 0

	tc.drainWg.Add(1)

	// Progress display
	ticker := time.NewTicker(500 * time.Millisecond)

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
				if !tc.drained && tc.isDrained(tc.hwm, offsets) {
					ticker.Stop()
					tc.drained = true
					tc.drainWg.Done()
				}

			case notif := <-consumer.Notifications():
				log.Infof("Rebalance - claim %v, release %v", notif.Claimed[kafka.TopicTasks()], notif.Released[kafka.TopicTasks()])
				for p := range notif.Released[kafka.TopicTasks()] {
					if tc.partitions[p] != nil {
						close(tc.partitions[p])
					}
				}
				for p := range notif.Claimed[kafka.TopicTasks()] {
					tc.hwm = <-tc.highWaterMarks()
					if tc.drained {
						tc.drained = false
						tc.drainWg.Add(1)
					}

					tc.partitions[p] = make(chan models.Task)
					tc.tasks <- tc.partitions[p]
				}
			case <-ticker.C:
				log.WithField("count", messages).Debug("Loading tasks")
			}
		}
	}()

	return tc, nil
}

// Tasks return the incomming task channel
func (tc *TaskConsumer) Tasks() <-chan chan models.Task {
	return tc.tasks
}

// WaitForDrain wait for consumer to EOF partitions
func (tc *TaskConsumer) WaitForDrain() {
	tc.drainWg.Wait()
}

// Close the task consumer
func (tc *TaskConsumer) Close() (err error) {
	if e := tc.consumer.Close(); e != nil {
		err = e
	}
	if e := tc.client.Close(); e != nil {
		err = e
	}
	return
}

// Handle incomming messages
func (tc *TaskConsumer) handleMsg(msg *sarama.ConsumerMessage) {
	var t models.Task
	if err := t.FromKafka(msg); err != nil {
		log.Error(err)
		return
	}
	log.Debugf("Task received: %v partition %v", t.ToJSON(), msg.Partition)
	tc.partitions[int(msg.Partition)] <- t
}

// Retrieve highWaterMarks for each partition
func (tc *TaskConsumer) highWaterMarks() chan map[int32]int64 {
	resChan := make(chan map[int32]int64)

	go func() {
		for {
			parts, err := tc.client.Partitions(kafka.TopicTasks())
			if err != nil {
				log.Warn("Can't get topic. Retry")
				continue
			}

			res := make(map[int32]int64)
			for p := range parts {
				i, err := tc.client.GetOffset(kafka.TopicTasks(), int32(p), sarama.OffsetNewest)
				if err != nil {
					log.Panic(err)
				}

				res[int32(p)] = i
			}

			resChan <- res
			close(resChan)
			break
		}
	}()

	return resChan
}

// Check if consumer reach EOF on all the partitions
func (tc *TaskConsumer) isDrained(hwm, offsets map[int32]int64) bool {
	subs := tc.consumer.Subscriptions()[kafka.TopicTasks()]

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
