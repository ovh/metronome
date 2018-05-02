package core

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ovh/metronome/src/metronome/kafka"
)

// Kafka handle Kafka connection.
// The producer use a WaitForAll strategy to perform message ack.
type Kafka struct {
	Producer sarama.SyncProducer
}

var k *Kafka
var once sync.Once

// GetKafka return the kafka instance.
func GetKafka() *Kafka {
	once.Do(func() {
		brokers := viper.GetStringSlice("kafka.brokers")

		config := kafka.NewConfig()
		config.ClientID = "metronome-api"
		config.Producer.RequiredAcks = sarama.WaitForAll
		config.Producer.Timeout = 1 * time.Second
		config.Producer.Compression = sarama.CompressionGZIP
		config.Producer.Flush.Frequency = 500 * time.Millisecond
		config.Producer.Partitioner = sarama.NewHashPartitioner
		config.Producer.Return.Successes = true
		config.Producer.Retry.Max = 3

		producer, err := sarama.NewSyncProducer(brokers, config)
		if err != nil {
			log.WithError(err).Fatal("Could not connect to kafka")
		}

		k = &Kafka{Producer: producer}
	})

	return k
}

// Close the producer.
func (k *Kafka) Close() error {
	return k.Producer.Close()
}
