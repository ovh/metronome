package cmd

import (
	log "github.com/sirupsen/logrus"
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
		assets := []string{
			"extensions.sql",
			"users.sql",
			"tasks.sql",
			"tokens.sql",
		}

		for _, asset := range assets {
			content, err := pg.Assets().MustString(asset)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"asset": asset,
				}).Error("Cannot found the asset")
				continue
			}

			if _, err := database.Exec(content); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"asset": asset,
				}).Error("Failed to setup table")
			}
		}

		log.Info("Done")
		if err := database.Close(); err != nil {
			log.WithError(err).Error("Could not gracefully close the connection to the database")
		}
	},
}
