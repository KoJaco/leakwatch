package profile

import "time"

// Profile is an immutable point-in-time snapshot of goroutine leak state.
// It answers "what did the runtime report at time T?". This is not whether something is growing.
type Profile struct {
	CapturedAt time.Time
	Goroutines []Goroutine
}

// Goroutine represents a single goroutine in a profile.
// No ID field in v1. Goroutine IDs are not stable across samples.
type Goroutine struct {
	Stack []Frame
}

// Frame is a single stack frame.
type Frame struct {
	Function string
	File     string
	Line     int
}
