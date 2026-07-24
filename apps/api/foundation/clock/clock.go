// Package clock provides a time boundary for deterministic tests and production.
package clock

import "time"

// Clock abstracts the current time so services and tests can control it.
type Clock interface {
	Now() time.Time
}

// Real returns UTC wall-clock time.
type Real struct{}

// Now returns the current time in UTC.
func (Real) Now() time.Time { return time.Now().UTC() }

// Fixed returns a fixed time for every call.
type Fixed struct{ T time.Time }

// Now returns the configured fixed time in UTC.
func (f Fixed) Now() time.Time { return f.T.UTC() }

// Advance adds d to the fixed time.
func (f *Fixed) Advance(d time.Duration) { f.T = f.T.Add(d) }
