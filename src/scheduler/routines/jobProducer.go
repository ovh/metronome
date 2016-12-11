package routines

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/d33d33/viper" // FIXME https://github.com/spf13/viper/pull/285

	"github.com/runabove/metronome/src/metronome/models"
)

// jobProducer handle the internal states of the producer
type jobProducer struct {
	producer sarama.AsyncProducer
	wg       sync.WaitGroup
	stopSig  chan struct{}
}

// Return a new job producer
// Read jobs to send from jobs channel
func NewJobProducer(jobs <-chan []models.Job) *jobProducer {
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

	jp := &jobProducer{
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
	go func() {
		jp.wg.Add(1)
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
	go func() {
		jp.wg.Add(1)
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
func (jp *jobProducer) Close() {
	jp.stopSig <- struct{}{}
	jp.producer.AsyncClose()
	jp.wg.Wait()
}
