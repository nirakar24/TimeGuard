package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var dstZone string

func init() {
	cmd := &cobra.Command{
		Use:   "dst-check <datetime>",
		Short: "Check if a local datetime is during a DST transition (gap or overlap)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing datetime argument (format: YYYY-MM-DD HH:MM)")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dstZone == "" {
				return errors.New("--zone required")
			}
			loc, err := time.LoadLocation(dstZone)
			if err != nil {
				return fmt.Errorf("invalid zone: %w", err)
			}
			input := args[0]
			layout := "2006-01-02 15:04"
			altLayout := "2006-01-02T15:04"
			var t time.Time
			if tt, e := time.ParseInLocation(layout, input, loc); e == nil {
				t = tt
			} else if tt2, e2 := time.ParseInLocation(altLayout, input, loc); e2 == nil {
				t = tt2
			} else {
				return fmt.Errorf("parse datetime: %v / %v", e, e2)
			}

			pre := t.Add(-30 * time.Minute)
			post := t.Add(30 * time.Minute)
			_, preOff := pre.Zone()
			_, tOff := t.Zone()
			_, postOff := post.Zone()

			if preOff != tOff && tOff == postOff {
				fmt.Println("Gap transition: time is near a forward DST jump (some local times may not exist)")
			} else if preOff == tOff && tOff != postOff {
				fmt.Println("Forward transition just after the gap")
			} else if preOff != tOff && tOff != postOff && preOff == postOff {
				fmt.Println("Overlap: clocks rolled back; this local time is ambiguous")
			} else {
				fmt.Println("No DST transition within ±30m; time is stable")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dstZone, "zone", "", "IANA timezone")
	rootCmd.AddCommand(cmd)
}
