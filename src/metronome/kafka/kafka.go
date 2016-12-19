package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
)

// NewConfig returns a new state sarama config.
// Preset TLS and SASL config
func NewConfig() *sarama.Config {
	viper.SetDefault("kafka.tls", false)

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
	viper.SetDefault("kafka.topics.tasks", "tasks")
	return viper.GetString("kafka.topics.tasks")
}

// TopicJobs kafka topic used for jobs
func TopicJobs() string {
	viper.SetDefault("kafka.topics.jobs", "jobs")
	return viper.GetString("kafka.topics.jobs")
}

// TopicStates kafka topic used for states
func TopicStates() string {
	viper.SetDefault("kafka.topics.states", "states")
	return viper.GetString("kafka.topics.states")
}

// GroupSchedulers kafka consumer group used for schedulers
func GroupSchedulers() string {
	viper.SetDefault("kafka.groups.schedulers", "schedulers")
	return viper.GetString("kafka.groups.schedulers")
}

// GroupAggregators kafka consumer group used for aggregators
func GroupAggregators() string {
	viper.SetDefault("kafka.groups.aggregators", "aggregators")
	return viper.GetString("kafka.groups.aggregators")
}

// GroupWorkers kafka consumer group used for workers
func GroupWorkers() string {
	viper.SetDefault("kafka.groups.workers", "workers")
	return viper.GetString("kafka.groups.workers")
}
