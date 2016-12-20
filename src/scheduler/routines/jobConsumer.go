package routines

import (
	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/runabove/metronome/src/metronome/kafka"
	"github.com/runabove/metronome/src/metronome/models"
)

// NewJobComsumer return a new job consumer
func NewJobComsumer(offsets map[int32]int64, jobs chan models.Job) {
	brokers := viper.GetStringSlice("kafka.brokers")

	config := kafka.NewConfig()
	config.ClientID = "metronome-scheduler"

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Error(err)
		return
	}

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		log.Error(err)
		return
	}

	parts, err := client.Partitions(kafka.TopicJobs())
	if err != nil {
		log.Error(err)
		return
	}

	hwm := make(map[int32]int64)
	for p := range parts {
		i, err := client.GetOffset(kafka.TopicJobs(), int32(p), sarama.OffsetNewest)
		if err != nil {
			log.Error(err)
			return
		}

		hwm[int32(p)] = i
	}

	for part, offset := range offsets {
		if (offset + 1) >= hwm[part] {
			// already to the end
			continue
		}

		pc, err := consumer.ConsumePartition(kafka.TopicJobs(), part, offset)
		if err != nil {
			log.Error(err)
			continue
		}

	consume:
		for {
			select {
			case msg := <-pc.Messages():
				var j models.Job
				if err := j.FromKafka(msg); err != nil {
					log.Error(err)
					continue
				}
				log.Debugf("Job received: %v", j.ToJSON())
				jobs <- j

				if (msg.Offset + 1) >= hwm[part] {
					break consume
				}
			}
		}

		if err := pc.Close(); err != nil {
			log.Error(err)
		}
	}

	close(jobs)
}
