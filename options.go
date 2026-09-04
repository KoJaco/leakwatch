package leakwatch

import (
	"time"

	"github.com/KoJaco/leakwatch/internal/collector"
)

// Option configures a Watcher.
type Option func(*config)

type config struct {
	interval  time.Duration
	jitter    time.Duration
	sampleGate collector.SampleGate
	pprofURL  string
}

// WithInterval sets the sampling interval.
func WithInterval(d time.Duration) Option {
	return func(c *config) {
		c.interval = d
	}
}

// WithJitter sets random jitter applied before each sample.
func WithJitter(d time.Duration) Option {
	return func(c *config) {
		c.jitter = d
	}
}

// WithSampleGate sets a gate that can skip individual samples.
func WithSampleGate(gate collector.SampleGate) Option {
	return func(c *config) {
		c.sampleGate = gate
	}
}

// WithPProfURL sets the pprof goroutine leak endpoint for the default collector.
func WithPProfURL(url string) Option {
	return func(c *config) {
		c.pprofURL = url
	}
}
