package consumers

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	saramaC "github.com/d33d33/sarama-cluster"
	"github.com/d33d33/viper" // FIXME https://github.com/spf13/viper/pull/285

	"github.com/runabove/metronome/src/metronome/constants"
	"github.com/runabove/metronome/src/metronome/models"
)

// JobConsumer consumed jobs messages from a Kafka topic and send them as HTTP POST request.
type JobConsumer struct {
	consumer *saramaC.Consumer
}

// NewJobConsumer returns a new job consumer.
func NewJobConsumer() (*JobConsumer, error) {
	brokers := viper.GetStringSlice("kafka.brokers")

	config := saramaC.NewConfig()
	config.ClientID = "metronome-worker"

	consumer, err := saramaC.NewConsumer(brokers, "worker", []string{constants.KafkaTopicJobs}, config)
	if err != nil {
		return nil, err
	}

	jc := &JobConsumer{
		consumer: consumer,
	}

	go func() {
		for {
			select {
			case msg, ok := <-consumer.Messages():
				if !ok { // shuting down
					return
				}
				jc.handleMsg(msg)
			}
		}
	}()

	return jc, nil
}

// Close the consumer.
func (jc *JobConsumer) Close() error {
	return jc.consumer.Close()
}

// Handle message from Kafka.
// Forward them as http POST.
func (jc *JobConsumer) handleMsg(msg *sarama.ConsumerMessage) {
	var j models.Job
	if err := j.FromKafka(msg); err != nil {
		log.Error(err)
		return
	}

	v := url.Values{}
	v.Set("time", strconv.FormatInt(j.At, 10))
	v.Set("epsilon", strconv.FormatInt(j.Epsilon, 10))
	v.Set("urn", j.URN)
	v.Set("at", strconv.FormatInt(time.Now().Unix(), 10))

	log.WithFields(log.Fields{
		"time":    j.At,
		"epsilon": j.Epsilon,
		"urn":     j.URN,
		"at":      time.Now().Unix(),
	}).Debug("POST")

	http.PostForm(j.URN, v)
}
