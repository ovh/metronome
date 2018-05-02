package cmd

import (
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ovh/metronome/src/aggregator/consumers"
	"github.com/ovh/metronome/src/metronome/metrics"
)

// Aggregator init - define command line arguments.
func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringP("config", "", "", "config file to use")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	RootCmd.Flags().String("pg.addr", "127.0.0.1:5432", "postgres address")
	RootCmd.Flags().String("pg.user", "metronome", "postgres user")
	RootCmd.Flags().String("pg.password", "metropass", "postgres password")
	RootCmd.Flags().String("pg.database", "metronome", "postgres database")
	RootCmd.Flags().StringSlice("kafka.brokers", []string{"localhost:9092"}, "kafka brokers address")
	RootCmd.Flags().String("redis.addr", "127.0.0.1:6379", "redis address")
	RootCmd.Flags().String("metrics.addr", "127.0.0.1:9100", "metrics address")

	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		log.WithError(err).Error("Could not bind persistent flags")
	}

	if err := viper.BindPFlags(RootCmd.Flags()); err != nil {
		log.WithError(err).Error("Could not bind flags")
	}
}

// Load config - initialize defaults and read config.
func initConfig() {
	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}

	// Set defaults
	viper.SetDefault("metrics.addr", ":9100")
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("redis.pass", "")
	viper.SetDefault("kafka.tls", false)
	viper.SetDefault("kafka.topics.tasks", "tasks")
	viper.SetDefault("kafka.topics.jobs", "jobs")
	viper.SetDefault("kafka.topics.states", "states")
	viper.SetDefault("kafka.groups.schedulers", "schedulers")
	viper.SetDefault("kafka.groups.aggregators", "aggregators")
	viper.SetDefault("kafka.groups.workers", "workers")
	viper.SetDefault("worker.poolsize", 100)
	viper.SetDefault("token.ttl", 3600)
	viper.SetDefault("redis.pass", "")

	// Bind environment variables
	viper.SetEnvPrefix("mtragg")
	viper.AutomaticEnv()

	// Set config search path
	viper.AddConfigPath("/etc/metronome/")
	viper.AddConfigPath("$HOME/.metronome")
	viper.AddConfigPath(".")

	// Load default config
	viper.SetConfigName("default")
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug("No default config file found")
		} else {
			log.Panicf("Fatal error in default config file: %v \n", err)
		}
	}

	// Load aggregator config
	viper.SetConfigName("aggregator")
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug("No aggregator config file found")
		} else {
			log.Panicf("Fatal error in aggregator config file: %v \n", err)
		}
	}

	// Load user defined config
	cfgFile := viper.GetString("Config")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		err := viper.ReadInConfig()
		if err != nil {
			log.Panicf("Fatal error in config file: %v \n", err)
		}
	}
}

// RootCmd launch the aggregator agent.
var RootCmd = &cobra.Command{
	Use:   "metronome-aggregator",
	Short: "Metronome aggregator update task database",
	Long: `Metronome is a distributed and fault-tolerant event scheduler built with love by ovh teams and friends in Go.
Complete documentation is available at http://ovh.github.io/metronome`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Metronome Aggregator starting")

		metrics.Serve()

		tc, err := consumers.NewTaskConsumer()
		if err != nil {
			log.WithError(err).Fatal("Could not start the task consumer")
		}

		sc, err := consumers.NewStateConsumer()
		if err != nil {
			log.WithError(err).Fatal("Could not start the state consumer")
		}

		log.Info("Started")

		// Trap SIGINT to trigger a shutdown.
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		<-sigint

		log.Info("Shuting down")
		if err := sc.Close(); err != nil {
			log.WithError(err).Error("Could not stop gracefully the state consumer")
		}

		if err := tc.Close(); err != nil {
			log.WithError(err).Error("Could not stop gracefully the task consumer")
		}
	},
}
