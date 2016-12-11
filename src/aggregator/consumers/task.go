package consumers

import (
	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	saramaC "github.com/d33d33/sarama-cluster"
	"github.com/d33d33/viper" // FIXME https://github.com/spf13/viper/pull/285

	"github.com/runabove/metronome/src/metronome/constants"
	"github.com/runabove/metronome/src/metronome/models"
	"github.com/runabove/metronome/src/metronome/pg"
)

type taskConsumer struct {
	consumer *saramaC.Consumer
}

func NewTaskConsumer() (*taskConsumer, error) {
	brokers := viper.GetStringSlice("kafka.brokers")

	config := saramaC.NewConfig()
	config.ClientID = "metronome-scheduler"
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := saramaC.NewConsumer(brokers, "aggregator", []string{constants.KAFKA_TOPIC_TASKS}, config)
	if err != nil {
		return nil, err
	}

	tc := &taskConsumer{
		consumer: consumer,
	}

	go func() {
		for {
			select {
			case msg, ok := <-consumer.Messages():
				if !ok { // shuting down
					return
				}
				tc.handleMsg(msg)
			}
		}
	}()

	return tc, nil
}

func (tc *taskConsumer) Close() error {
	return tc.consumer.Close()
}

func (tc *taskConsumer) handleMsg(msg *sarama.ConsumerMessage) {
	var t models.Task
	if err := t.FromKafka(msg); err != nil {
		log.Error(err)
		return
	}

	db := pg.DB()

	if t.Schedule == "" {
		log.Infof("DELETE task: %s", t.Guid)

		_, err := db.Model(&t).Delete()
		if err != nil {
			log.Errorf("%v %v", err) // TODO log for replay or not commit
		}
		return
	}

	_, err := db.Model(&t).OnConflict("(Guid) DO UPDATE").Set("name = ?name", "urn = ?urn", "schedule = ?schedule").Insert()
	if err != nil {
		log.Errorf("%v", err) // TODO log for replay or not commit
	}
}
