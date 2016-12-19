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
