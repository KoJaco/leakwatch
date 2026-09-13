package leakwatch

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/KoJaco/leakwatch/internal/collector"
	"github.com/KoJaco/leakwatch/internal/profile"
)

// Option configures a Watcher.
type Option func(*config)

type config struct {
	interval         time.Duration
	jitter           time.Duration
	sampleGate       collector.SampleGate
	pprofURL         string
	historyCapacity  int
	httpClient       *http.Client
	httpTimeout      time.Duration
	maxProfileBytes  int64
	logger           *slog.Logger
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

// WithHistoryCapacity sets the number of observations retained in memory.
func WithHistoryCapacity(n int) Option {
	return func(c *config) {
		c.historyCapacity = n
	}
}

// WithHTTPTimeout sets the timeout for the default pprof HTTP client.
// Ignored when WithHTTPClient is also set.
func WithHTTPTimeout(d time.Duration) Option {
	return func(c *config) {
		c.httpTimeout = d
	}
}

// WithMaxProfileBytes sets the maximum pprof response body size.
func WithMaxProfileBytes(n int64) Option {
	return func(c *config) {
		c.maxProfileBytes = n
	}
}

// WithHTTPClient sets the HTTP client used to fetch pprof profiles.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		c.httpClient = client
	}
}

// WithLogger sets the logger used for sample failures and watcher lifecycle events.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}

func defaultConfig() config {
	return config{
		interval:        5 * time.Minute,
		jitter:          30 * time.Second,
		sampleGate:      collector.AlwaysAllow,
		pprofURL:        "http://127.0.0.1:6060/debug/pprof/goroutineleak",
		historyCapacity: defaultHistoryCapacity,
		httpTimeout:     profile.DefaultHTTPTimeout,
		maxProfileBytes: profile.DefaultMaxProfileBytes,
		logger:          slog.Default(),
	}
}
