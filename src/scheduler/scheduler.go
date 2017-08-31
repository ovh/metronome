// Scheduler plan tasks executions.
//
// Usage
//
// 		scheduler  [flags]
// Flags:
//       --config string            config file to use
//       --help                     display help
//   -v, --verbose                  verbose output
package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/ovh/metronome/src/scheduler/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Panicf("%v", err)
	}
}
