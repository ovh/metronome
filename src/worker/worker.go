// Worker perform jobs.
//
// You can launch as much Worker agent as you want/need as they rely on Kafka partitons to share the workload.
//
// Usage
//
// 		worker  [flags]
// Flags:
//       --config string            config file to use
//       --help                     display help
//   -v, --verbose                  verbose output
package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/ovh/metronome/src/worker/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.WithError(err).Error("Could not execute the worker")
	}
}
