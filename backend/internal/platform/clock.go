package platform

import "time"

// Clock is a small abstraction over time.Now for testability.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// NewClock returns a Clock that delegates to time.Now (UTC).
func NewClock() Clock { return realClock{} }
