package routines

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ovh/metronome/src/metronome/kafka"
	"github.com/ovh/metronome/src/metronome/models"
)

// JobProducer handle the internal states of the producer.
type JobProducer struct {
	producer     sarama.AsyncProducer
	wg           sync.WaitGroup
	stopSig      chan struct{}
	offsets      map[int32]int64
	offsetsMutex sync.RWMutex
}

// NewJobProducer return a new job producer.
// Read jobs to send from jobs channel.
func NewJobProducer(jobs <-chan []models.Job) (*JobProducer, error) {
	config := kafka.NewConfig()
	config.ClientID = "metronome-scheduler"
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Timeout = 1 * time.Second
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 300 * time.Millisecond
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3

	brokers := viper.GetStringSlice("kafka.brokers")

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	jp := &JobProducer{
		producer: producer,
		stopSig:  make(chan struct{}),
		offsets:  make(map[int32]int64),
	}

	go func() {
		for {
			select {
			case js, ok := <-jobs:
				if !ok {
					return
				}
				for _, j := range js {
					producer.Input() <- j.ToKafka()
				}

			case <-jp.stopSig:
				return
			}
		}
	}()

	// Success handling
	jp.wg.Add(1)
	go func() {
		for {
			select {
			case msg, ok := <-producer.Successes():
				if !ok {
					jp.wg.Done()
					return
				}
				jp.offsetsMutex.Lock()
				jp.offsets[msg.Partition] = msg.Offset
				jp.offsetsMutex.Unlock()
				log.Debugf("Msg send: %v", msg)
			}
		}
	}()

	// Failure handling
	jp.wg.Add(1)
	go func() {
		for {
			select {
			case err, ok := <-producer.Errors():
				if !ok {
					jp.wg.Done()
					return
				}
				log.Errorf("Failed to send message: %v", err)
			}
		}
	}()

	return jp, nil
}

// Close the job producer
func (jp *JobProducer) Close() {
	jp.stopSig <- struct{}{}
	jp.producer.AsyncClose()
	jp.wg.Wait()
}

// Indexes return the current write indexes by partition
func (jp *JobProducer) Indexes() map[int32]int64 {
	res := make(map[int32]int64)

	jp.offsetsMutex.RLock()
	defer jp.offsetsMutex.RUnlock()

	for k, v := range jp.offsets {
		res[k] = v
	}

	return res
}
