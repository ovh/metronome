package cmd

import (
	"os"
	"os/signal"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ovh/metronome/src/metronome/metrics"
	"github.com/ovh/metronome/src/scheduler/routines"
)

// Scheduler init - define command line arguments
func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringP("config", "", "", "config file to use")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	RootCmd.Flags().StringSlice("kafka.brokers", []string{"localhost:9092"}, "kafka brokers address")
	RootCmd.Flags().String("redis.addr", "127.0.0.1:6379", "redis address")
	RootCmd.Flags().String("metrics.addr", "127.0.0.1:9100", "metrics address")

	if err := viper.BindPFlags(RootCmd.Flags()); err != nil {
		log.WithError(err).Error("Could not bind the flags")
	}

	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		log.WithError(err).Error("Could not bind the persistent flags")
	}
}

// Load config - initialize defaults and read config
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
	viper.SetEnvPrefix("mtrsch")
	viper.SetEnvKeyReplacer(strings.NewReplacer("_", "."))
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

	// Load scheduler config
	viper.SetConfigName("scheduler")
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug("No scheduler config file found")
		} else {
			log.Panicf("Fatal error in scheduler config file: %v \n", err)
		}
	}

	// Load user defined config
	cfgFile := viper.GetString("config")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		err := viper.ReadInConfig()
		if err != nil {
			log.Panicf("Fatal error in config file: %v \n", err)
		}
	}
}

// RootCmd launch the scheduler agent.
var RootCmd = &cobra.Command{
	Use:   "metronome-scheduler",
	Short: "Metronome scheduler plan tasks executions",
	Long: `Metronome is a distributed and fault-tolerant event scheduler built with love by ovh teams and friends in Go.
Complete documentation is available at http://ovh.github.io/metronome`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Metronome Scheduler starting")

		log.Info("Loading tasks")
		tc, err := routines.NewTaskConsumer()
		if err != nil {
			log.WithError(err).Fatal("Could not start the task consumer")
		}

		metrics.Serve()

		// Trap SIGINT to trigger a shutdown.
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		var schedulers sync.WaitGroup

		running := true

	loop:
		for {
			select {
			case partition := <-tc.Partitons():
				schedulers.Add(1)
				go func() {
					log.Infof("Scheduler start %v", partition.Partition)
					ts, err := routines.NewTaskScheduler(partition.Partition, partition.Tasks)
					if err != nil {
						log.WithError(err).Error("Could not create a new task scheduler")
						return
					}

					tc.WaitForDrain()
					log.Infof("Scheduler tasks loaded %v", partition.Partition)
					if running {
						if err = ts.Start(); err != nil {
							log.WithError(err).Error("Could not start the task scheduler")
							return
						}
						log.Infof("Scheduler started %v", partition.Partition)
					}

					ts.Halted()
					log.Infof("Scheduler halted %v", partition.Partition)
					schedulers.Done()
				}()
			case <-sigint:
				log.Info("Shuting down")
				running = false
				break loop
			}
		}

		if err = tc.Close(); err != nil {
			log.WithError(err).Error("Could not stop gracefully the task consumer")
		}

		log.Infof("Consumer halted")
		schedulers.Wait()
	},
}
