package workflowconfig

import (
	"testing"
	"time"
)

func TestSendWindowNext(t *testing.T) {
	ny, _ := time.LoadLocation("America/New_York")
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	weekdays := SendWindow{Enabled: true, Start: "09:00", End: "17:00", Days: []int{1, 2, 3, 4, 5}}
	night := SendWindow{Enabled: true, Start: "22:00", End: "02:00", Days: []int{5}}
	at := func(loc *time.Location, s string) time.Time {
		v, err := time.ParseInLocation("2006-01-02 15:04", s, loc)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	cases := []struct {
		name   string
		window SendWindow
		loc    *time.Location
		now    time.Time
		want   time.Time
	}{
		{"inside", weekdays, ny, at(ny, "2026-11-04 10:30"), at(ny, "2026-11-04 10:30")},
		{"before opening", weekdays, ny, at(ny, "2026-11-04 07:00"), at(ny, "2026-11-04 09:00")},
		{"end is exclusive", weekdays, ny, at(ny, "2026-11-04 17:00"), at(ny, "2026-11-05 09:00")},
		{"friday evening waits for monday", weekdays, ny, at(ny, "2026-11-06 18:00"), at(ny, "2026-11-09 09:00")},
		{"saturday waits for monday", weekdays, ny, at(ny, "2026-11-07 12:00"), at(ny, "2026-11-09 09:00")},
		// 10:30 in New York is 00:30 Thursday in Tokyo: closed there until 09:00.
		{"reader time zone", weekdays, tokyo, at(ny, "2026-11-04 10:30"), at(tokyo, "2026-11-05 09:00")},
		// Clocks fall back on Sunday 1 November 2026; Monday 09:00 is EST, 14:00 UTC.
		{"after DST ends", weekdays, ny, at(ny, "2026-10-31 12:00"), time.Date(2026, 11, 2, 14, 0, 0, 0, time.UTC)},
		// Clocks spring forward on Sunday 8 March 2026; Monday 09:00 is EDT, 13:00 UTC.
		{"after DST starts", weekdays, ny, at(ny, "2026-03-07 12:00"), time.Date(2026, 3, 9, 13, 0, 0, 0, time.UTC)},
		{"overnight window open after midnight", night, ny, at(ny, "2026-11-07 01:00"), at(ny, "2026-11-07 01:00")},
		{"overnight window closed", night, ny, at(ny, "2026-11-07 03:00"), at(ny, "2026-11-13 22:00")},
		{"disabled", SendWindow{}, ny, at(ny, "2026-11-07 03:00"), at(ny, "2026-11-07 03:00")},
	}
	for _, c := range cases {
		if got := c.window.Next(c.now, c.loc); !got.Equal(c.want) {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}

	// 02:30 does not exist on 8 March 2026 in New York; the window opens at 03:30 EDT.
	skipped := SendWindow{Enabled: true, Start: "02:30", End: "05:00", Days: []int{0}}
	if got := skipped.Next(at(ny, "2026-03-08 00:00"), ny); !got.Equal(time.Date(2026, 3, 8, 7, 30, 0, 0, time.UTC)) {
		t.Errorf("DST gap: got %s", got.UTC())
	}
}

func TestSendRulesValidation(t *testing.T) {
	for _, bad := range []string{
		`{"enabled":true,"start":"9","end":"17:00","days":[1]}`,
		`{"enabled":true,"start":"09:00","end":"09:00","days":[1]}`,
		`{"enabled":true,"start":"09:00","end":"17:00","days":[]}`,
		`{"enabled":true,"start":"09:00","end":"17:00","days":[7]}`,
	} {
		if _, err := ParseSendWindow([]byte(bad)); err == nil {
			t.Errorf("accepted %s", bad)
		}
	}
	if w, err := ParseSendWindow([]byte("null")); w != nil || err != nil {
		t.Error("null window should mean none")
	}
	if ValidTimezone("Local") || ValidTimezone("Mars/Olympus") || !ValidTimezone("Europe/London") {
		t.Error("time zone validation")
	}
	if Location("", "Bad/Zone", "Asia/Kolkata").String() != "Asia/Kolkata" || Location().String() != "UTC" {
		t.Error("time zone fallback")
	}
}
