package clock

import "time"

// Clocker abstracts time so callers can replace real time in tests.
type Clocker interface {
	Now() time.Time
}

// TimeClocker is the production clock implementation backed by time.Now.
type TimeClocker struct{}

// New returns a TimeClocker that reads the current system time.
func New() *TimeClocker {
	return &TimeClocker{}
}

// Now returns the current system time.
func (*TimeClocker) Now() time.Time {
	return time.Now()
}
