package timeutil

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

func Convert(dt, from, to string) (time.Time, time.Time, error) {
	if from == "" || to == "" {
		return time.Time{}, time.Time{}, errors.New("from/to required")
	}
	srcLoc, err := time.LoadLocation(from)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	dstLoc, err := time.LoadLocation(to)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	layouts := []string{"2006-01-02 15:04", "2006-01-02T15:04"}
	var t time.Time
	var parseErr error
	for _, l := range layouts {
		if tt, e := time.ParseInLocation(l, dt, srcLoc); e == nil {
			t = tt
			parseErr = nil
			break
		} else {
			parseErr = e
		}
	}
	if parseErr != nil {
		return time.Time{}, time.Time{}, parseErr
	}
	return t, t.In(dstLoc), nil
}

type DSTTransitionType string

const (
	DSTNone    DSTTransitionType = "none"
	DSTGap     DSTTransitionType = "gap"
	DSTOverlap DSTTransitionType = "overlap"
)

type DSTInfo struct {
	Type        DSTTransitionType
	PreOffset   int
	Offset      int
	PostOffset  int
	Ambiguous   bool      // true if local representation could map to two UTC instants
	Nonexistent bool      // true if local time falls in a forward gap
	Parsed      time.Time // time returned by parser (may be adjusted)
}

func AnalyzeDST(dt, zone string) (DSTInfo, error) {
	info := DSTInfo{Type: DSTNone}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return info, err
	}
	layouts := []string{"2006-01-02 15:04", "2006-01-02T15:04"}
	var t time.Time
	var parseErr error
	for _, l := range layouts {
		if tt, e := time.ParseInLocation(l, dt, loc); e == nil {
			t = tt
			parseErr = nil
			break
		} else {
			parseErr = e
		}
	}
	if parseErr != nil {
		return info, parseErr
	}
	pre := t.Add(-30 * time.Minute)
	post := t.Add(30 * time.Minute)
	_, preOff := pre.Zone()
	_, off := t.Zone()
	_, postOff := post.Zone()
	info.PreOffset, info.Offset, info.PostOffset = preOff, off, postOff
	info.Parsed = t

	// Identify transition edges.
	if preOff == off && off == postOff { // stable
		return info, nil
	}
	// Forward jump (gap) usually: preOff != off and off == postOff OR preOff == off and off < postOff
	if (preOff != off && off == postOff && off > preOff) || (preOff == off && postOff > off) {
		info.Type = DSTGap
		// Heuristic for nonexistent: if hour in the missing block (commonly 02) and forward jump present.
		hour := t.Hour()
		if hour == 2 || hour == 3 { // typical spring-forward window hits 02:xx
			info.Nonexistent = true
		}
		return info, nil
	}
	// Backward jump (overlap) typically: preOff == off && off != postOff && postOff < off OR off != postOff && preOff == postOff && preOff > off
	if (preOff == off && postOff < off) || (preOff == postOff && preOff > off && off != postOff) || (preOff != off && off != postOff && preOff == postOff && preOff < off) {
		info.Type = DSTOverlap
		hour := t.Hour()
		if hour == 1 || hour == 2 { // fall-back repeated hour (commonly 02)
			info.Ambiguous = true
		}
		return info, nil
	}
	// Fallback: previous simple pattern from earlier version
	if preOff != off && off == postOff {
		info.Type = DSTGap
		return info, nil
	}
	if preOff != off && off != postOff && preOff == postOff {
		info.Type = DSTOverlap
		info.Ambiguous = true
		return info, nil
	}
	return info, nil
}

func CheckDSTTransition(dt, zone string) (DSTTransitionType, error) {
	info, err := AnalyzeDST(dt, zone)
	return info.Type, err
}

type LeapSecondEntries []struct {
	Date string `json:"date"`
}

func LoadLeapData(path string) (LeapSecondEntries, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var list LeapSecondEntries
	if err := json.NewDecoder(f).Decode(&list); err != nil {
		return nil, err
	}
	return list, nil
}

func (l LeapSecondEntries) IsLeapSecond(ts string) bool {
	for _, e := range l {
		if e.Date == ts {
			return true
		}
	}
	return false
}

type SmearPoint struct {
	Hour          int
	OffsetSeconds float64
	Adjusted      time.Time
}

func GoogleSmear(date string) ([]SmearPoint, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	secondsInDay := 86400.0
	points := make([]SmearPoint, 0, 24)
	for h := 0; h < 24; h++ {
		secondsFromStart := float64(h * 3600)
		offset := (secondsFromStart/secondsInDay - 0.5) * 1.0
		if offset < -0.5 {
			offset = -0.5
		}
		if offset > 0.5 {
			offset = 0.5
		}
		adj := day.Add(time.Duration(h) * time.Hour).Add(time.Duration(offset * float64(time.Second)))
		points = append(points, SmearPoint{Hour: h, OffsetSeconds: offset, Adjusted: adj})
	}
	return points, nil
}

var rfc3339Like = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:Z|[+-]\d{2}:?\d{2})?)`)

func ValidateLogLine(line string, prev time.Time) (issues []string, t time.Time) {
	match := rfc3339Like.FindString(line)
	if match == "" {
		return nil, prev
	}
	clean := match
	if len(clean) >= 24 && strings.HasSuffix(clean, "Z") == false && clean[22] != ':' { // normalize offset colon
		clean = clean[:22] + ":" + clean[22:]
	}
	tp, err := time.Parse(time.RFC3339, clean)
	if err != nil {
		return []string{"parse-error"}, prev
	}
	if !prev.IsZero() && tp.Before(prev.Add(-5*time.Minute)) {
		issues = append(issues, "time-regression")
	}
	if !strings.Contains(line, "TZ=") && !strings.Contains(line, "zone=") {
		issues = append(issues, "missing-timezone-context")
	}
	return issues, tp
}

func ValidateReader(r io.Reader) (issues int, lines int) {
	scanner := bufio.NewScanner(r)
	// Increase the max token size to handle very long log lines (up to 1MB each).
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	var prev time.Time
	for scanner.Scan() {
		lines++
		iss, t := ValidateLogLine(scanner.Text(), prev)
		if len(iss) > 0 {
			issues += len(iss)
		}
		if !t.IsZero() {
			prev = t
		}
	}
	return issues, lines
}
