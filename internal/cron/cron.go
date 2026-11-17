// Package cron is a tiny 5-field cron matcher (min hour dom month dow) with
// standard Vixie-cron semantics: when both day-of-month and day-of-week are
// restricted, a time matches if EITHER matches. Day-of-week accepts 0-7
// (0 and 7 are Sunday). Macros @hourly @daily @weekly @monthly @yearly work.
package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Schedule struct {
	f       [5]map[int]bool
	domStar bool
	dowStar bool
}

var bounds = [5][2]int{{0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 7}}

var macros = map[string]string{
	"@yearly": "0 0 1 1 *", "@annually": "0 0 1 1 *", "@monthly": "0 0 1 * *",
	"@weekly": "0 0 * * 0", "@daily": "0 0 * * *", "@midnight": "0 0 * * *", "@hourly": "0 * * * *",
}

func Parse(spec string) (*Schedule, error) {
	spec = strings.TrimSpace(spec)
	if m, ok := macros[strings.ToLower(spec)]; ok {
		spec = m
	}
	parts := strings.Fields(spec)
	if len(parts) != 5 {
		return nil, fmt.Errorf("cron needs 5 fields, got %d", len(parts))
	}
	s := &Schedule{domStar: parts[2] == "*" || parts[2] == "?", dowStar: parts[4] == "*" || parts[4] == "?"}
	for i, p := range parts {
		set := map[int]bool{}
		for _, item := range strings.Split(p, ",") {
			step := 1
			if k := strings.Index(item, "/"); k >= 0 {
				n, err := strconv.Atoi(item[k+1:])
				if err != nil || n < 1 {
					return nil, fmt.Errorf("bad step in %q", item)
				}
				step, item = n, item[:k]
			}
			lo, hi := bounds[i][0], bounds[i][1]
			if i == 4 && (item == "*" || item == "?") {
				hi = 6 // don't double-count Sunday for "*"
			}
			if item != "*" && item != "?" {
				if r := strings.Split(item, "-"); len(r) == 2 {
					a, e1 := strconv.Atoi(r[0])
					b, e2 := strconv.Atoi(r[1])
					if e1 != nil || e2 != nil {
						return nil, fmt.Errorf("bad range %q", item)
					}
					lo, hi = a, b
				} else {
					a, err := strconv.Atoi(item)
					if err != nil {
						return nil, fmt.Errorf("bad value %q", item)
					}
					lo, hi = a, a
					if step > 1 {
						hi = bounds[i][1]
					}
				}
			}
			if lo < bounds[i][0] || hi > bounds[i][1] || lo > hi {
				return nil, fmt.Errorf("value out of range in %q", p)
			}
			for v := lo; v <= hi; v += step {
				if i == 4 && v == 7 {
					v2 := 0
					set[v2] = true
					continue
				}
				set[v] = true
			}
		}
		s.f[i] = set
	}
	return s, nil
}

func (s *Schedule) Match(t time.Time) bool {
	if !s.f[0][t.Minute()] || !s.f[1][t.Hour()] || !s.f[3][int(t.Month())] {
		return false
	}
	dom, dow := s.f[2][t.Day()], s.f[4][int(t.Weekday())]
	switch {
	case s.domStar && s.dowStar:
		return true
	case s.domStar:
		return dow
	case s.dowStar:
		return dom
	default:
		return dom || dow
	}
}

// Next returns up to n match times strictly after t (searches ~4 years).
func (s *Schedule) Next(t time.Time, n int) []time.Time {
	var out []time.Time
	cur := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location()).Add(time.Minute)
	limit := cur.Add(4 * 366 * 24 * time.Hour)
	for len(out) < n && cur.Before(limit) {
		if !s.f[3][int(cur.Month())] {
			cur = time.Date(cur.Year(), cur.Month()+1, 1, 0, 0, 0, 0, cur.Location())
			continue
		}
		if !s.f[1][cur.Hour()] {
			// time.Truncate works on absolute time and breaks in half-hour zones (IST)
			cur = time.Date(cur.Year(), cur.Month(), cur.Day(), cur.Hour(), 0, 0, 0, cur.Location()).Add(time.Hour)
			continue
		}
		if s.Match(cur) {
			out = append(out, cur)
		}
		cur = cur.Add(time.Minute)
	}
	return out
}
