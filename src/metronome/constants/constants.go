package constants

import (
	"github.com/spf13/viper"
)

// KafkaTopicTasks kafka topic used for tasks
func KafkaTopicTasks() string {
	viper.SetDefault("kafka.topics.tasks", "tasks")
	return viper.GetString("kafka.topics.tasks")
}

// KafkaTopicJobs kafka topic used for jobs
func KafkaTopicJobs() string {
	viper.SetDefault("kafka.topics.jobs", "jobs")
	return viper.GetString("kafka.topics.jobs")
}

// KafkaTopicStates kafka topic used for states
func KafkaTopicStates() string {
	viper.SetDefault("kafka.topics.states", "states")
	return viper.GetString("kafka.topics.states")
}
