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

type jobConsumer struct {
	consumer *saramaC.Consumer
}

func NewJobConsumer() (*jobConsumer, error) {
	brokers := viper.GetStringSlice("kafka.brokers")

	config := saramaC.NewConfig()
	config.ClientID = "metronome-worker"
	// config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := saramaC.NewConsumer(brokers, "worker", []string{constants.KAFKA_TOPIC_JOBS}, config)
	if err != nil {
		return nil, err
	}

	jc := &jobConsumer{
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

func (jc *jobConsumer) Close() error {
	return jc.consumer.Close()
}

func (jc *jobConsumer) handleMsg(msg *sarama.ConsumerMessage) {
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
