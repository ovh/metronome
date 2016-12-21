package consumers

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	saramaC "github.com/d33d33/sarama-cluster"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"

	"github.com/runabove/metronome/src/metronome/kafka"
	"github.com/runabove/metronome/src/metronome/models"
)

// JobConsumer consumed jobs messages from a Kafka topic and send them as HTTP POST request.
type JobConsumer struct {
	consumer *saramaC.Consumer
	producer sarama.SyncProducer
	// metrics
	jobCounter        *prometheus.CounterVec
	jobTime           *prometheus.HistogramVec
	jobSuccessCounter *prometheus.CounterVec
	jobFailureCounter *prometheus.CounterVec
}

// NewJobConsumer returns a new job consumer.
func NewJobConsumer() (*JobConsumer, error) {
	brokers := viper.GetStringSlice("kafka.brokers")

	config := saramaC.NewConfig()
	config.Config = *kafka.NewConfig()
	config.ClientID = "metronome-worker"
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Timeout = 1 * time.Second
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 3

	consumer, err := saramaC.NewConsumer(brokers, kafka.GroupWorkers(), []string{kafka.TopicJobs()}, config)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducer(brokers, &config.Config)
	if err != nil {
		return nil, err
	}

	jc := &JobConsumer{
		consumer: consumer,
		producer: producer,
	}

	// metrics
	jc.jobCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "metronome",
		Subsystem: "worker",
		Name:      "jobs",
		Help:      "Number of jobs processed.",
	},
		[]string{"partition"})
	prometheus.MustRegister(jc.jobCounter)
	jc.jobTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "metronome",
		Subsystem: "worker",
		Name:      "job_time",
		Help:      "Job processing time.",
	},
		[]string{"partition"})
	prometheus.MustRegister(jc.jobTime)
	jc.jobSuccessCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "metronome",
		Subsystem: "worker",
		Name:      "jobs_success",
		Help:      "Number of jobs success.",
	},
		[]string{"partition"})
	prometheus.MustRegister(jc.jobSuccessCounter)
	jc.jobFailureCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "metronome",
		Subsystem: "worker",
		Name:      "jobs_failure",
		Help:      "Number of jobs failure.",
	},
		[]string{"partition"})
	prometheus.MustRegister(jc.jobFailureCounter)

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
	jc.jobCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()
	var j models.Job
	if err := j.FromKafka(msg); err != nil {
		log.Error(err)
		return
	}

	start := time.Now()

	v := url.Values{}
	v.Set("time", strconv.FormatInt(j.At, 10))
	v.Set("epsilon", strconv.FormatInt(j.Epsilon, 10))
	v.Set("urn", j.URN)
	v.Set("at", strconv.FormatInt(time.Now().Unix(), 10))

	log.WithFields(log.Fields{
		"time":    j.At,
		"epsilon": j.Epsilon,
		"urn":     j.URN,
		"at":      start,
	}).Debug("POST")

	res, err := http.PostForm(j.URN, v)
	s := models.State{
		ID:       "",
		TaskGUID: j.GUID,
		UserID:   j.UserID,
		At:       j.At,
		DoneAt:   start.Unix(),
		Duration: time.Since(start).Nanoseconds() / 1000,
		URN:      j.URN,
		State:    models.Success,
	}

	jc.jobTime.WithLabelValues(strconv.Itoa(int(msg.Partition))).Observe(time.Since(start).Seconds()) // to seconds

	if err != nil {
		log.Warn(err)
		s.State = models.Failed
	} else if res.StatusCode < 200 || res.StatusCode >= 300 {
		s.State = models.Failed
	}
	if err == nil {
		res.Body.Close()
	}

	switch s.State {
	case models.Success:
		jc.jobSuccessCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()
	case models.Failed:
		jc.jobFailureCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()

	}

	if _, _, err := jc.producer.SendMessage(s.ToKafka()); err != nil {
		log.Errorf("FAILED to send message: %s\n", err)
	}
}
