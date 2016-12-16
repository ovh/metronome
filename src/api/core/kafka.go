package core

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
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
	brokers := viper.GetStringSlice("kafka.brokers")

	once.Do(func() {
		config := sarama.NewConfig()
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
			panic(err)
		}

		k = &Kafka{
			Producer: producer,
		}
	})

	return k
}

// Close the producer.
func (k *Kafka) Close() error {
	if err := k.Producer.Close(); err != nil {
		log.Error("Failed to shut down producer cleanly", err)
	}

	return nil
}
