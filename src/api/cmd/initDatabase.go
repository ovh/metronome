package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ovh/metronome/src/metronome/pg"
)

// init - define command line arguments.
func init() {
	RootCmd.AddCommand(initDatabaseCmd)
}

// initDatabaseCmd init the database.
var initDatabaseCmd = &cobra.Command{
	Use:   "init-database",
	Short: "Init the database schemas",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Initializing database schema")

		database := pg.DB()

		if _, err := database.Exec(string(pg.MustAsset("extensions.sql"))); err != nil {
			log.Errorf("Failed to setup extensions: %v", err)
		}
		if _, err := database.Exec(string(pg.MustAsset("users.sql"))); err != nil {
			log.Errorf("Failed to setup users table: %v", err)
		}
		if _, err := database.Exec(string(pg.MustAsset("tasks.sql"))); err != nil {
			log.Errorf("Failed to setup tasks table: %v", err)
		}
		if _, err := database.Exec(string(pg.MustAsset("tokens.sql"))); err != nil {
			log.Errorf("Failed to setup tokens table: %v", err)
		}

		log.Info("Done")
	},
}
