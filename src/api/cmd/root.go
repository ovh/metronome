package cmd

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"

	"github.com/ovh/metronome/src/api/core"
	"github.com/ovh/metronome/src/api/routers"
	"github.com/ovh/metronome/src/metronome/metrics"
	"github.com/ovh/metronome/src/metronome/pg"
)

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringP("config", "", "", "config file to use")
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().String("pg.addr", "127.0.0.1:5432", "postgres address")
	RootCmd.PersistentFlags().String("pg.user", "metronome", "postgres user")
	RootCmd.PersistentFlags().String("pg.password", "metropass", "postgres password")
	RootCmd.PersistentFlags().String("pg.database", "metronome", "postgres database")

	RootCmd.Flags().StringP("api.http.listen", "l", "0.0.0.0:8080", "api listen addresse")
	RootCmd.Flags().StringSlice("kafka.brokers", []string{"localhost:9092"}, "kafka brokers address")
	RootCmd.Flags().String("redis.addr", "127.0.0.1:6379", "redis address")
	RootCmd.Flags().String("metrics.addr", "127.0.0.1:9100", "metrics address")

	if err := viper.BindPFlags(RootCmd.PersistentFlags()); err != nil {
		log.WithError(err).Error("Could not bind persitent flags")
	}

	if err := viper.BindPFlags(RootCmd.Flags()); err != nil {
		log.WithError(err).Error("Could not bind flags")
	}
}

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
	viper.SetEnvPrefix("mtrapi")
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

	// Load api config
	viper.SetConfigName("api")
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug("No api config file found")
		} else {
			log.Panicf("Fatal error in api config file: %v \n", err)
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

	// Required
	if !viper.IsSet("token.key") {
		log.Panic("'token.key' is required")
	}
}

// RootCmd launch the api agent.
var RootCmd = &cobra.Command{
	Use:   "metronome-api",
	Short: "Metronome api provide a rest api to manage metronome tasks",
	Long: `Metronome is a distributed and fault-tolerant event scheduler built with love by ovh teams and friends in Go.
Complete documentation is available at http://ovh.github.io/metronome`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Metronome API starting")

		n := negroni.New()

		// Log request
		logger := &negroni.Logger{
			ALogger: core.RequestLogger{
				LogType: "access",
				Level:   log.InfoLevel,
			},
		}
		logger.SetDateFormat(negroni.LoggerDefaultDateFormat)
		logger.SetFormat(negroni.LoggerDefaultFormat)
		n.Use(logger)

		// Handle handlers panic
		recovery := negroni.NewRecovery()
		recovery.Logger = core.RequestLogger{
			LogType: "recovery",
			Level:   log.ErrorLevel,
		}
		n.Use(recovery)

		// CORS support
		n.Use(cors.New(cors.Options{
			AllowedHeaders: []string{"Authorization", "Content-Type"},
			AllowedMethods: []string{"GET", "POST", "DELETE"},
		}))

		// Load routes
		router := routers.InitRoutes()
		n.UseHandler(router)

		server := &http.Server{
			Addr:    viper.GetString("api.http.listen"),
			Handler: n,
		}

		// Serve metrics
		metrics.Serve()

		go func() {
			log.Info("Metronome API started")
			log.Infof("Listen %s", viper.GetString("api.http.listen"))
			if err := server.ListenAndServe(); err != nil {
				log.WithError(err).Error("Could not start the server")
			}
		}()

		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)

		<-sigint

		if err := server.Close(); err != nil {
			log.WithError(err).Error("Could not stop gracefully the server")
		}

		database := pg.DB()
		if err := database.Close(); err != nil {
			log.WithError(err).Error("Could not stop gracefully close the connection to the database")
		}
	},
}
