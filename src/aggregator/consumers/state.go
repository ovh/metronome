package consumers

import (
	"strconv"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	saramaC "github.com/d33d33/sarama-cluster"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"

	"github.com/runabove/metronome/src/metronome/kafka"
	"github.com/runabove/metronome/src/metronome/models"
	"github.com/runabove/metronome/src/metronome/redis"
)

// StateConsumer consumed states messages from Kafka to maintain the state database.
type StateConsumer struct {
	consumer                  *saramaC.Consumer
	stateCounter              *prometheus.CounterVec
	stateUnprocessableCounter *prometheus.CounterVec
	stateProcessedCounter     *prometheus.CounterVec
	statePublishErrorCounter  *prometheus.CounterVec
}

// NewStateConsumer returns a new state consumer.
func NewStateConsumer() (*StateConsumer, error) {
	brokers := viper.GetStringSlice("kafka.brokers")

	config := saramaC.NewConfig()
	config.Config = *kafka.NewConfig()
	config.ClientID = "metronome-aggregator"
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := saramaC.NewConsumer(brokers, kafka.GroupAggregators(), []string{kafka.TopicStates()}, config)
	if err != nil {
		return nil, err
	}

	sc := &StateConsumer{
		consumer: consumer,
	}

	// metrics
	sc.stateCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "metronome",
		Subsystem: "aggregator",
		Name:      "states",
		Help:      "Number of states processed.",
	},
		[]string{"partition"})
	prometheus.MustRegister(sc.stateCounter)
	sc.stateUnprocessableCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "metronome",
		Subsystem: "aggregator",
		Name:      "states_unprocessable",
		Help:      "Number of unprocessable states.",
	},
		[]string{"partition"})
	prometheus.MustRegister(sc.stateUnprocessableCounter)
	sc.stateProcessedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "metronome",
		Subsystem: "aggregator",
		Name:      "states_processeed",
		Help:      "Number of processeed states.",
	},
		[]string{"partition"})
	prometheus.MustRegister(sc.stateProcessedCounter)
	sc.statePublishErrorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "metronome",
		Subsystem: "aggregator",
		Name:      "states_publish_error",
		Help:      "Number of states publish error.",
	},
		[]string{"partition"})
	prometheus.MustRegister(sc.statePublishErrorCounter)

	go func() {
		for {
			select {
			case msg, ok := <-consumer.Messages():
				if !ok { // shuting down
					return
				}
				sc.handleMsg(msg)
			}
		}
	}()

	return sc, nil
}

// Close the consumer.
func (sc *StateConsumer) Close() error {
	return sc.consumer.Close()
}

// Handle message from Kafka.
// Apply updates to the database.
func (sc *StateConsumer) handleMsg(msg *sarama.ConsumerMessage) {
	sc.stateCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()
	var s models.State
	if err := s.FromKafka(msg); err != nil {
		sc.stateUnprocessableCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()
		log.Error(err)
		return
	}

	log.Infof("UPDATE state: %s", s.TaskGUID)

	if err := redis.DB().HSet(s.UserID, s.TaskGUID, s.ToJSON()).Err(); err != nil {
		panic(err)
	}
	sc.stateProcessedCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()
	if err := redis.DB().PublishTopic(s.UserID, "state", s.ToJSON()).Err(); err != nil {
		sc.statePublishErrorCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()
		log.Error(err)
	}
}
