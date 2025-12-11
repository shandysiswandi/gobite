package pkgclock

import "time"

type Clocker interface {
	Now() time.Time
}

type TimeClocker struct{}

func New() *TimeClocker {
	return &TimeClocker{}
}

func (*TimeClocker) Now() time.Time {
	return time.Now()
}
