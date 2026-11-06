package workflowconfig

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

// SendWindow limits automation emails to opening hours in the reader's time zone.
// End is exclusive; an End at or before Start closes the window the next day.
type SendWindow struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"` // HH:MM
	End     string `json:"end"`   // HH:MM
	Days    []int  `json:"days"`  // days the window opens, 0 = Sunday
}

// ParseSendWindow reads a stored window; empty or null means none was set.
func ParseSendWindow(raw []byte) (*SendWindow, error) {
	if s := strings.TrimSpace(string(raw)); s == "" || s == "null" {
		return nil, nil
	}
	var w SendWindow
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, fmt.Errorf("invalid send window")
	}
	return &w, w.Validate()
}

func (w SendWindow) Validate() error {
	if !w.Enabled {
		return nil
	}
	start, err1 := time.Parse("15:04", w.Start)
	end, err2 := time.Parse("15:04", w.End)
	if err1 != nil || err2 != nil || start.Equal(end) {
		return fmt.Errorf("send window needs different start and end times (HH:MM)")
	}
	if len(w.Days) == 0 || len(w.Days) > 7 {
		return fmt.Errorf("choose at least one sending day")
	}
	for _, d := range w.Days {
		if d < 0 || d > 6 {
			return fmt.Errorf("invalid sending day")
		}
	}
	return nil
}

// Next returns now when the window is open, else the next opening, in loc.
func (w SendWindow) Next(now time.Time, loc *time.Location) time.Time {
	if !w.Enabled || len(w.Days) == 0 {
		return now
	}
	start, _ := time.Parse("15:04", w.Start)
	end, _ := time.Parse("15:04", w.End)
	local := now.In(loc)
	// Start a day early so a window that crosses midnight is still found open.
	for i := -1; i <= 7; i++ {
		day := time.Date(local.Year(), local.Month(), local.Day()+i, 0, 0, 0, 0, loc)
		if !slices.Contains(w.Days, int(day.Weekday())) {
			continue
		}
		open := wallClock(day, start)
		closeDay := day
		if !end.After(start) {
			closeDay = day.AddDate(0, 0, 1)
		}
		shut := wallClock(closeDay, end)
		if now.Before(shut) {
			if now.Before(open) {
				return open
			}
			return now
		}
	}
	return now
}

// wallClock is clock's time on day; a time skipped by a DST change moves forward past the gap.
func wallClock(day, clock time.Time) time.Time {
	t := time.Date(day.Year(), day.Month(), day.Day(), clock.Hour(), clock.Minute(), 0, 0, day.Location())
	if t.Hour() != clock.Hour() || t.Minute() != clock.Minute() {
		_, before := t.Zone()
		_, after := t.Add(3 * time.Hour).Zone()
		t = t.Add(time.Duration(after-before) * time.Second)
	}
	return t
}

// ValidTimezone reports whether name is an IANA time zone.
func ValidTimezone(name string) bool {
	if name == "" || strings.EqualFold(name, "local") {
		return false
	}
	_, err := time.LoadLocation(name)
	return err == nil
}

// Location returns the first valid time zone among names, else UTC.
func Location(names ...string) *time.Location {
	for _, name := range names {
		if ValidTimezone(name) {
			loc, _ := time.LoadLocation(name)
			return loc
		}
	}
	return time.UTC
}

// CapPeriod is the rolling window a frequency cap counts emails over.
func CapPeriod(period string) time.Duration {
	if period == "week" {
		return 7 * 24 * time.Hour
	}
	return 24 * time.Hour
}
