package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var validateFile string
var validateJSON bool
var validateSampleLimit int

func init() {
	cmd := &cobra.Command{
		Use:   "validate-logs <file>",
		Short: "Validate log file timestamps for common temporal issues",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing file")
			}
			validateFile = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := os.Open(validateFile)
			if err != nil {
				return err
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			buf := make([]byte, 64*1024)
			scanner.Buffer(buf, 1024*1024)
			tsRe := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:Z|[+-]\d{2}:?\d{2})?)`)

			lineNum := 0
			var prev time.Time
			issuesTotal := 0
			linesWithTS := 0

			counts := map[string]int{"parse-error": 0, "time-regression": 0, "missing-timezone-context": 0}
			samples := map[string][]string{"parse-error": {}, "time-regression": {}, "missing-timezone-context": {}}

			addSample := func(category, line string) {
				if !validateJSON {
					return
				}
				if len(samples[category]) < validateSampleLimit {
					trim := line
					if len(trim) > 200 {
						trim = trim[:200] + "…"
					}
					samples[category] = append(samples[category], fmt.Sprintf("%d: %s", lineNum, trim))
				}
			}

			for scanner.Scan() {
				lineNum++
				line := scanner.Text()
				match := tsRe.FindString(line)
				if match == "" {
					continue
				}
				linesWithTS++
				clean := match
				if len(clean) == 20 { // YYYY-MM-DDTHH:MM:SSZ
				} else if strings.HasSuffix(clean, "Z") {
				} else if len(clean) >= 24 && (clean[22] != ':') {
					clean = clean[:22] + ":" + clean[22:]
				}
				t, err := time.Parse(time.RFC3339, clean)
				if err != nil {
					counts["parse-error"]++
					issuesTotal++
					addSample("parse-error", line)
					if !validateJSON {
						fmt.Printf("line %d: parse error: %v\n", lineNum, err)
					}
					continue
				}
				if !prev.IsZero() && t.Before(prev.Add(-5*time.Minute)) {
					counts["time-regression"]++
					issuesTotal++
					addSample("time-regression", line)
					if !validateJSON {
						fmt.Printf("line %d: time regression >5m (prev %s, current %s)\n", lineNum, prev.Format(time.RFC3339), t.Format(time.RFC3339))
					}
				}
				if !strings.Contains(line, "TZ=") && !strings.Contains(line, "zone=") {
					counts["missing-timezone-context"]++
					issuesTotal++
					addSample("missing-timezone-context", line)
					if !validateJSON {
						fmt.Printf("line %d: no explicit timezone context\n", lineNum)
					}
				}
				prev = t
			}
			if err := scanner.Err(); err != nil {
				return err
			}

			if validateJSON {
				payload := struct {
					File           string              `json:"file"`
					LinesProcessed int                 `json:"lines_processed"`
					LinesWithTS    int                 `json:"lines_with_timestamp"`
					Counts         map[string]int      `json:"counts"`
					Samples        map[string][]string `json:"samples"`
					IssuesTotal    int                 `json:"issues_total"`
				}{
					File:           validateFile,
					LinesProcessed: lineNum,
					LinesWithTS:    linesWithTS,
					Counts:         counts,
					Samples:        samples,
					IssuesTotal:    issuesTotal,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(payload)
			}
			fmt.Printf("Validation complete: %d issues\n", issuesTotal)
			return nil
		},
	}
	cmd.Flags().BoolVar(&validateJSON, "json", false, "Emit JSON summary instead of line-by-line output")
	cmd.Flags().IntVar(&validateSampleLimit, "samples", 5, "Number of sample lines to include per issue category in JSON output")
	rootCmd.AddCommand(cmd)
}
