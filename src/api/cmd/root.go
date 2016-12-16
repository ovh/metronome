package cmd

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"

	"github.com/runabove/metronome/src/api/core"
	"github.com/runabove/metronome/src/api/routers"
)

var cfgFile string
var verbose bool

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file to use")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	RootCmd.Flags().StringP("api.http.listen", "l", "0.0.0.0:8080", "api listen addresse")
	RootCmd.Flags().String("pg.addr", "127.0.0.1:5432", "postgres address")
	RootCmd.Flags().String("pg.user", "metronome", "postgres user")
	RootCmd.Flags().String("pg.password", "metropass", "postgres password")
	RootCmd.Flags().String("pg.database", "metronome", "postgres database")
	RootCmd.Flags().StringSlice("kafka.brokers", []string{"localhost:9092"}, "kafka brokers address")
	RootCmd.Flags().String("redis.addr", "127.0.0.1:6379", "redis address")

	viper.BindPFlags(RootCmd.Flags())
}

func initConfig() {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	// Defaults
	viper.SetDefault("token.ttl", 3600)

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
Complete documentation is available at http://runabove.github.io/metronome`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Metronome API starting")

		n := negroni.New()

		// Log request
		logger := &negroni.Logger{
			core.RequestLogger{
				"access",
				log.InfoLevel,
			},
		}
		n.Use(logger)

		// Handle handlers panic
		recovery := negroni.NewRecovery()
		recovery.Logger = core.RequestLogger{
			"recovery",
			log.ErrorLevel,
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

		log.Info("Metronome API started")
		log.Infof("Listen %s", viper.GetString("api.http.listen"))

		log.Fatal(http.ListenAndServe(viper.GetString("api.http.listen"), n))
	},
}
