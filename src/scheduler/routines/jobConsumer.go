package routines

import (
	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ovh/metronome/src/metronome/kafka"
	"github.com/ovh/metronome/src/metronome/models"
)

// NewJobComsumer return a new job consumer
func NewJobComsumer(offsets map[int32]int64, jobs chan models.Job) error {
	brokers := viper.GetStringSlice("kafka.brokers")

	config := kafka.NewConfig()
	config.ClientID = "metronome-scheduler"

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return err
	}

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return err
	}

	parts, err := client.Partitions(kafka.TopicJobs())
	if err != nil {
		return err
	}

	hwm := make(map[int32]int64)
	for p := range parts {
		i, err := client.GetOffset(kafka.TopicJobs(), int32(p), sarama.OffsetNewest)
		if err != nil {
			return err
		}

		hwm[int32(p)] = i
	}

	go func() {
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
						log.WithError(err).Warn("Could not retrieve the job from kafka message")
						continue
					}
					body, err := j.ToJSON()
					if err != nil {
						log.WithError(err).Warn("Could not deserialize the job")
						continue
					}
					log.Debugf("Job received: %s", string(body))
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
	}()

	return nil
}
