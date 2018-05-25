package consumers

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	saramaC "github.com/bsm/sarama-cluster"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ovh/metronome/src/metronome/kafka"
	"github.com/ovh/metronome/src/metronome/models"
)

// JobConsumer consumed jobs messages from a Kafka topic and send them as HTTP POST request.
type JobConsumer struct {
	consumer *saramaC.Consumer
	producer sarama.SyncProducer
	wg       *sync.WaitGroup // Used to sync shut down
	// metrics
	jobCounter        *prometheus.CounterVec
	jobTime           *prometheus.HistogramVec
	jobSuccessCounter *prometheus.CounterVec
	jobFailureCounter *prometheus.CounterVec
	jobExpireCounter  *prometheus.CounterVec
	httpClient        *http.Client
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

	jc.httpClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   300 * time.Millisecond,
				KeepAlive: 1 * time.Minute,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 1024,
		},
	}

	// worker
	jc.wg = new(sync.WaitGroup)

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
	jc.jobExpireCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "metronome",
		Subsystem: "worker",
		Name:      "jobs_expire",
		Help:      "Number of expired jobs.",
	},
		[]string{"partition"})
	prometheus.MustRegister(jc.jobExpireCounter)

	// Spawning workers
	poolSize := viper.GetInt("worker.poolsize")

	log.Printf("Spawning %d goroutines...", poolSize)
	for i := 0; i < poolSize; i++ {
		jc.wg.Add(1)
		go jc.Worker(i + 1)
	}
	return jc, nil
}

// Close the consumer.
func (jc *JobConsumer) Close() error {
	err := jc.consumer.Close()
	jc.wg.Wait() // wait for all workers to shut down properly
	return err
}

// Worker is the main goroutine that is calling handleMsg
func (jc *JobConsumer) Worker(id int) {
	defer jc.wg.Done()
	for {
		select {
		case msg, ok := <-jc.consumer.Messages():
			if !ok { // shutting down
				log.Printf("Closing worker %d", id)
				return
			}
			if err := jc.handleMsg(msg); err != nil {
				log.
					WithError(err).
					WithFields(log.Fields{"id": id}).
					Error("Could not handle the message")
			}
		}
	}
}

// Handle message from Kafka.
// Forward them as http POST.
func (jc *JobConsumer) handleMsg(msg *sarama.ConsumerMessage) error {
	jc.jobCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()
	var j models.Job
	if err := j.FromKafka(msg); err != nil {
		return err
	}

	start := time.Now()

	log.WithFields(log.Fields{
		"time":    j.At,
		"epsilon": j.Epsilon,
		"urn":     j.URN,
		"at":      start,
	}).Debug("POST")

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

	if j.At < start.Unix()-j.Epsilon {
		s.State = models.Expired
	} else {
		url, err := url.Parse(s.URN)
		if err != nil {
			s.State = models.Failed
		} else {
			q := url.Query()
			q.Set("time", strconv.FormatInt(j.At, 10))
			q.Set("epsilon", strconv.FormatInt(j.Epsilon, 10))
			q.Set("at", strconv.FormatInt(time.Now().Unix(), 10))
			url.RawQuery = q.Encode()

			body, err := json.Marshal(j.Payload)
			if err != nil {
				log.WithError(err).Warn("Cannot marshall payload")
				s.State = models.Failed
			} else {
				res, err := jc.httpClient.Post(url.String(), "application/json", bytes.NewReader(body))
				if err != nil {
					log.WithError(err).Warn("Could not post form")
				} else {
					if _, err = io.Copy(ioutil.Discard, res.Body); err != nil {
						log.WithError(err).Warn("Failed to discard response body")
					}
					if err = res.Body.Close(); err != nil {
						log.WithError(err).Warn("Could not close the response body")
					}
				}

				if err != nil || res.StatusCode < 200 || res.StatusCode >= 300 {
					s.State = models.Failed
				}
			}
		}
	}

	jc.jobTime.WithLabelValues(strconv.Itoa(int(msg.Partition))).Observe(time.Since(start).Seconds()) // to seconds

	switch s.State {
	case models.Success:
		jc.jobSuccessCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()
	case models.Failed:
		jc.jobFailureCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()
	case models.Expired:
		jc.jobExpireCounter.WithLabelValues(strconv.Itoa(int(msg.Partition))).Inc()
	}

	if _, _, err := jc.producer.SendMessage(s.ToKafka()); err != nil {
		log.Errorf("FAILED to send message: %s\n", err)
		return err
	}
	return nil
}
