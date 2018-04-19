package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
)

// NewConfig returns a new state sarama config.
// Preset TLS and SASL config
func NewConfig() *sarama.Config {
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
	return viper.GetString("kafka.topics.tasks")
}

// TopicJobs kafka topic used for jobs
func TopicJobs() string {
	return viper.GetString("kafka.topics.jobs")
}

// TopicStates kafka topic used for states
func TopicStates() string {
	return viper.GetString("kafka.topics.states")
}

// GroupSchedulers kafka consumer group used for schedulers
func GroupSchedulers() string {
	return viper.GetString("kafka.groups.schedulers")
}

// GroupAggregators kafka consumer group used for aggregators
func GroupAggregators() string {
	return viper.GetString("kafka.groups.aggregators")
}

// GroupWorkers kafka consumer group used for workers
func GroupWorkers() string {
	return viper.GetString("kafka.groups.workers")
}
