package misc

import "time"

var Clock ClockInterface

type ClockInterface interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
	Since(t time.Time) time.Duration
}

type realClock struct{}

func (realClock) Now() time.Time                         { return time.Now() }
func (realClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
func (realClock) Since(t time.Time) time.Duration        { return time.Since(t) }

func init() {
	Clock = &realClock{}
}
