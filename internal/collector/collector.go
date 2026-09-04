package collector

import (
	"context"

	"github.com/KoJaco/leakwatch/internal/profile"
)

// SampleGate decides whether a sample should proceed.
// Returns (false, nil) to skip, (true, nil) to proceed, (_, err) on failure.
type SampleGate func(context.Context) (bool, error)

// Collector fetches a goroutine leak profile safely and on schedule.
type Collector interface {
	Collect(context.Context) (profile.Profile, error)
}

// AlwaysAllow is a SampleGate that always permits sampling.
func AlwaysAllow(context.Context) (bool, error) {
	return true, nil
}

// NeverAllow is a SampleGate that always skips sampling.
func NeverAllow(context.Context) (bool, error) {
	return false, nil
}
