package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version holds the application version. Overridden at build time with:
// go build -ldflags "-X github.com/nirakar24/timeguard/cmd.Version=v0.1.0"
var Version = "v0.0.0-dev"

// Commit and Date allow embedding build metadata (optional).
var (
	Commit = ""
	Date   = ""
)

func init() {
	vCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("timeguard %s", Version)
			if Commit != "" {
				fmt.Printf(" commit=%s", Commit)
			}
			if Date != "" {
				fmt.Printf(" date=%s", Date)
			}
			fmt.Printf(" go=%s os=%s arch=%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
		},
	}
	rootCmd.AddCommand(vCmd)
}
