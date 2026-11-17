package cron

import (
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	at := func(s string) time.Time { v, _ := time.Parse("2006-01-02 15:04", s); return v }
	cases := []struct {
		spec string
		t    string
		want bool
	}{
		{"* * * * *", "2026-11-17 00:41", true},
		{"0 3 * * *", "2026-11-17 03:00", true},
		{"0 3 * * *", "2026-11-17 03:01", false},
		{"*/15 * * * *", "2026-11-17 10:45", true},
		{"*/15 * * * *", "2026-11-17 10:46", false},
		{"0 9-17 * * 1-5", "2026-11-17 12:00", true},  // Tuesday
		{"0 9-17 * * 1-5", "2026-11-15 12:00", false}, // Sunday
		{"0,30 * * * *", "2026-11-17 08:30", true},
		{"5/20 * * * *", "2026-11-17 08:25", true},
		// dom AND dow restricted -> OR semantics (1st of month OR Monday)
		{"0 0 1 * 1", "2026-11-16 00:00", true},  // Monday the 16th
		{"0 0 1 * 1", "2026-11-01 00:00", true},  // Sunday the 1st
		{"0 0 1 * 1", "2026-11-17 00:00", false}, // Tuesday the 17th
		{"0 0 * * 7", "2026-11-15 00:00", true},  // 7 = Sunday
		{"@daily", "2026-11-17 00:00", true},
	}
	for _, c := range cases {
		s, err := Parse(c.spec)
		if err != nil {
			t.Fatalf("%q: %v", c.spec, err)
		}
		if got := s.Match(at(c.t)); got != c.want {
			t.Errorf("%q at %s = %v, want %v", c.spec, c.t, got, c.want)
		}
	}
	for _, bad := range []string{"", "* * * *", "61 * * * *", "a * * * *", "*/0 * * * *", "5-1 * * * *"} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestNextHalfHourZone(t *testing.T) {
	ist := time.FixedZone("IST", 5*3600+1800)
	s, _ := Parse("0 3 * * *")
	got := s.Next(time.Date(2026, 11, 17, 9, 51, 0, 0, ist), 3)
	if len(got) != 3 || got[0].Hour() != 3 || got[0].Minute() != 0 || got[0].Day() != 18 {
		t.Fatalf("unexpected next runs: %v", got)
	}
	if got[2].Day() != 20 {
		t.Fatalf("third run: %v", got[2])
	}
}
