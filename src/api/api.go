// Api agent expose a simple HTTP interface to manage tasks.
//
// The api is stateless making horizontal scaling a breeze.
//
// Usage
//
// 		api  [flags]
// Flags:
//   -l, --api.http.listen string   api listen addresse
//       --config string            config file to use
//       --help                     display help
//   -v, --verbose                  verbose output
package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/ovh/metronome/src/api/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.WithError(err).Error("Could not execute the api")
	}
}
