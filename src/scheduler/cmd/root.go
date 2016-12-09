package cmd

import (
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/d33d33/viper" // FIXME https://github.com/spf13/viper/pull/285
	"github.com/spf13/cobra"

	"github.com/runabove/metronome/src/scheduler/routines"
)

var cfgFile string
var Verbose bool

// Scheduler init - define command line arguments
func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file to use")
	RootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")

	viper.BindPFlags(RootCmd.Flags())
}

// Load config - initialize defaults and read config
func initConfig() {
	if Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// Defaults
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})

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

// Main scheduler command - launch task scheduling process
var RootCmd = &cobra.Command{
	Use:   "metronome-scheduler",
	Short: "Metronome scheduler plan tasks executions",
	Long: `Metronome is a distributed and fault-tolerant event scheduler built with love by ovh teams and friends in Go.
Complete documentation is available at http://runabove.github.io/metronome`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Metronome Scheduler starting")

		log.Info("Loading tasks")
		tc, err := routines.NewTaskComsumer()
		if err != nil {
			log.Fatal(err)
		}

		ts := routines.NewTaskScheduler(tc.Tasks())

		tc.WaitForDrain()
		log.Info("Tasks loaded")

		jp := routines.NewJobProducer(ts.Jobs())
		ts.Start()
		log.Info("Started")

		// Trap SIGINT to trigger a shutdown.
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		<-sigint
		log.Info("Shuting down")
		err = tc.Close()
		ts.Stop()
		jp.Close()

		if err != nil {
			log.Fatal(err)
		}
	},
}
