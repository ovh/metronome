package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version = "0.0.0"
	githash = "HEAD"
	date    = "1970-01-01T00:00:00Z UTC"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, arguments []string) {
		fmt.Printf("metronome-scheduler version %s %s\n", version, githash)
		fmt.Printf("metronome-scheduler build date %s\n", date)
		fmt.Printf("go version %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}
