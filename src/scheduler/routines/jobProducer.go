package routines

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/runabove/metronome/src/metronome/models"
)

// JobProducer handle the internal states of the producer.
type JobProducer struct {
	producer sarama.AsyncProducer
	wg       sync.WaitGroup
	stopSig  chan struct{}
}

// NewJobProducer return a new job producer.
// Read jobs to send from jobs channel.
func NewJobProducer(jobs <-chan []models.Job) *JobProducer {
	config := sarama.NewConfig()
	config.ClientID = "metronome-scheduler"
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Timeout = 1 * time.Second
	config.Producer.Compression = sarama.CompressionGZIP
	config.Producer.Flush.Frequency = 300 * time.Millisecond
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3

	brokers := viper.GetStringSlice("kafka.brokers")

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}

	jp := &JobProducer{
		producer: producer,
		stopSig:  make(chan struct{}),
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

	return jp
}

// Close the job producer
func (jp *JobProducer) Close() {
	jp.stopSig <- struct{}{}
	jp.producer.AsyncClose()
	jp.wg.Wait()
}
