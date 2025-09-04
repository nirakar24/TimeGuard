package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var nowZone string

func init() {
	cmd := &cobra.Command{
		Use:   "now",
		Short: "Show current time in a given timezone with DST status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if nowZone == "" {
				return errors.New("--zone required")
			}
			loc, err := time.LoadLocation(nowZone)
			if err != nil {
				return err
			}
			now := time.Now().In(loc)
			_, off := now.Zone()
			fmt.Printf("%s (UTC%+03d:%02d) DST=%v\n", now.Format(time.RFC3339), off/3600, (off%3600)/60, now.IsDST())
			return nil
		},
	}
	cmd.Flags().StringVar(&nowZone, "zone", "", "IANA timezone")
	rootCmd.AddCommand(cmd)
}
