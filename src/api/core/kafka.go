package core

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/d33d33/viper" // FIXME https://github.com/spf13/viper/pull/285
)

type kafka struct {
	Producer sarama.SyncProducer
}

var k *kafka
var once sync.Once

func Kafka() *kafka {
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

		k = &kafka{
			Producer: producer,
		}
	})

	return k
}

func (k *kafka) Close() error {
	if err := k.Producer.Close(); err != nil {
		log.Error("Failed to shut down producer cleanly", err)
	}

	return nil
}
