package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var leapFile string

func init() {
	cmd := &cobra.Command{
		Use:   "leap-check <utc-second>",
		Short: "Check if a UTC timestamp is a leap second entry",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing UTC timestamp (e.g. 2016-12-31T23:59:60Z)")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := leapFile
			if path == "" {
				path = "internal/timeutil/leapdata.json"
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			var entries []struct {
				Date string `json:"date"`
			}
			if err := json.NewDecoder(f).Decode(&entries); err != nil {
				return err
			}
			needle := args[0]
			for _, e := range entries {
				if e.Date == needle {
					fmt.Printf("true: %s is a defined leap second\n", needle)
					return nil
				}
			}
			fmt.Printf("false: %s is not a leap second entry\n", needle)
			return nil
		},
	}
	cmd.Flags().StringVar(&leapFile, "data", "", "Path to leapdata.json")
	rootCmd.AddCommand(cmd)
}
