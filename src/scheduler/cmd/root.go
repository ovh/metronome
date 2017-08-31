package cmd

import (
	"os"
	"os/signal"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ovh/metronome/src/metronome/metrics"
	"github.com/ovh/metronome/src/scheduler/routines"
)

var cfgFile string
var verbose bool

// Scheduler init - define command line arguments
func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file to use")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	RootCmd.Flags().StringSlice("kafka.brokers", []string{"localhost:9092"}, "kafka brokers address")
	RootCmd.Flags().String("redis.addr", "127.0.0.1:6379", "redis address")
	RootCmd.Flags().String("metrics.addr", "127.0.0.1:9100", "metrics address")

	viper.BindPFlags(RootCmd.Flags())
}

// Load config - initialize defaults and read config
func initConfig() {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	// Bind environment variables
	viper.SetEnvPrefix("mtrsch")
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

		metrics.Serve()

		log.Info("Loading tasks")
		tc, err := routines.NewTaskComsumer()
		if err != nil {
			log.Fatal(err)
		}

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
					ts := routines.NewTaskScheduler(partition.Partition, partition.Tasks)

					tc.WaitForDrain()
					log.Infof("Scheduler tasks loaded %v", partition.Partition)
					if running {
						ts.Start()
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

		tc.Close()
		log.Infof("Consumer halted")
		schedulers.Wait()
	},
}
