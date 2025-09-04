package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var smearMethod string

func init() {
	cmd := &cobra.Command{
		Use:   "smear <date>",
		Short: "Simulate leap second smear adjustments across a 24h window",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing date (YYYY-MM-DD)")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			date := args[0]
			layout := "2006-01-02"
			day, err := time.Parse(layout, date)
			if err != nil {
				return fmt.Errorf("parse date: %w", err)
			}

			if smearMethod == "" {
				smearMethod = "google"
			}
			if smearMethod != "google" {
				return fmt.Errorf("unsupported smear method: %s", smearMethod)
			}
			secondsInDay := 86400.0
			fmt.Printf("Method: %s (total 1 leap second spread ±12h)\n", smearMethod)
			fmt.Println("Hour, SmearOffsetSeconds, AdjustedUTC")
			for h := 0; h < 24; h++ {
				secondsFromStart := float64(h * 3600)
				offset := (secondsFromStart/secondsInDay - 0.5) * 1.0 // range roughly -0.5..+0.5
				if offset < -0.5 {
					offset = -0.5
				}
				if offset > 0.5 {
					offset = 0.5
				}
				adjusted := day.Add(time.Duration(h) * time.Hour).Add(time.Duration(offset * float64(time.Second)))
				fmt.Printf("%02d, %.6f, %s\n", h, offset, adjusted.Format(time.RFC3339))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&smearMethod, "method", "google", "Smear method (google)")
	rootCmd.AddCommand(cmd)
}
