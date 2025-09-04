package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	convertDatetime string
	convertFrom     string
	convertTo       string
)

func init() {
	cmd := &cobra.Command{
		Use:   "convert <" + "datetime" + ">",
		Short: "Convert a datetime between IANA time zones",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing datetime argument (format: YYYY-MM-DD HH:MM)")
			}
			convertDatetime = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if convertFrom == "" || convertTo == "" {
				return errors.New("--from and --to are required")
			}
			srcLoc, err := time.LoadLocation(convertFrom)
			if err != nil {
				return fmt.Errorf("invalid from zone: %w", err)
			}
			dstLoc, err := time.LoadLocation(convertTo)
			if err != nil {
				return fmt.Errorf("invalid to zone: %w", err)
			}

			layout := "2006-01-02 15:04"
			altLayout := "2006-01-02T15:04"
			var t time.Time
			if tt, e := time.ParseInLocation(layout, convertDatetime, srcLoc); e == nil {
				t = tt
			} else if tt2, e2 := time.ParseInLocation(altLayout, convertDatetime, srcLoc); e2 == nil {
				t = tt2
			} else {
				return fmt.Errorf("parse datetime: %v / %v", e, e2)
			}

			converted := t.In(dstLoc)
			fmt.Printf("%s (%s) -> %s (%s)\n", t.Format(time.RFC3339), convertFrom, converted.Format(time.RFC3339), convertTo)
			return nil
		},
	}
	cmd.Flags().StringVar(&convertFrom, "from", "", "Source IANA timezone")
	cmd.Flags().StringVar(&convertTo, "to", "", "Target IANA timezone")
	rootCmd.AddCommand(cmd)
}
