package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/runabove/metronome/src/scheduler/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Panicf("%v", err)
	}
}
