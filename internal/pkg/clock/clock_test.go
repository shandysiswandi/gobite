package clock

import (
	"testing"
	"time"
)

func TestTimeClockerNow(t *testing.T) {
	cases := []struct {
		name   string
		clock  Clocker
		before time.Duration
		after  time.Duration
	}{
		{
			name:   "now_between_bounds",
			clock:  New(),
			before: -time.Second,
			after:  time.Second,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now().Add(tc.before)
			now := tc.clock.Now()
			end := time.Now().Add(tc.after)

			if now.Before(start) || now.After(end) {
				t.Fatalf("now %v not within expected bounds %v-%v", now, start, end)
			}
		})
	}
}
