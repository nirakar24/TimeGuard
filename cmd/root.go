package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "timeguard",
	Short: "TimeGuard: advanced timezone & temporal edge-case toolkit",
	Long:  `TimeGuard is a CLI to convert, validate, and simulate complex time behaviors (DST, leap seconds, smear).`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
