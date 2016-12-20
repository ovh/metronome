package kafka

import (
	"sync"

	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
)

var onceKafkaDefaults sync.Once

func setDefaults() {
	onceKafkaDefaults.Do(func() {
		viper.SetDefault("kafka.tls", false)
		viper.SetDefault("kafka.topics.tasks", "tasks")
		viper.SetDefault("kafka.topics.jobs", "jobs")
		viper.SetDefault("kafka.topics.states", "states")
		viper.SetDefault("kafka.groups.schedulers", "schedulers")
		viper.SetDefault("kafka.groups.aggregators", "aggregators")
		viper.SetDefault("kafka.groups.workers", "workers")
	})
}

// NewConfig returns a new state sarama config.
// Preset TLS and SASL config
func NewConfig() *sarama.Config {
	setDefaults()

	config := sarama.NewConfig()
	if viper.GetBool("kafka.tls") {
		config.Net.TLS.Enable = true
	}

	if viper.IsSet("kafka.sasl.user") || viper.IsSet("kafka.sasl.password") {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = viper.GetString("kafka.sasl.user")
		config.Net.SASL.Password = viper.GetString("kafka.sasl.password")
	}

	config.Version = sarama.V0_10_0_1

	return config
}

// TopicTasks kafka topic used for tasks
func TopicTasks() string {
	setDefaults()
	return viper.GetString("kafka.topics.tasks")
}

// TopicJobs kafka topic used for jobs
func TopicJobs() string {
	setDefaults()
	return viper.GetString("kafka.topics.jobs")
}

// TopicStates kafka topic used for states
func TopicStates() string {
	setDefaults()
	return viper.GetString("kafka.topics.states")
}

// GroupSchedulers kafka consumer group used for schedulers
func GroupSchedulers() string {
	setDefaults()
	return viper.GetString("kafka.groups.schedulers")
}

// GroupAggregators kafka consumer group used for aggregators
func GroupAggregators() string {
	setDefaults()
	return viper.GetString("kafka.groups.aggregators")
}

// GroupWorkers kafka consumer group used for workers
func GroupWorkers() string {
	setDefaults()
	return viper.GetString("kafka.groups.workers")
}
