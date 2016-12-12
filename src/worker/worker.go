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
	log "github.com/Sirupsen/logrus"

	"github.com/runabove/metronome/src/worker/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Panicf("%v", err)
	}
}
