package scheduler

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/KoJaco/leakwatch/internal/collector"
)

const (
	defaultInitialBackoff = time.Second
	defaultMaxBackoffCap    = 5 * time.Minute
)

// Config holds optional scheduler behaviour.
type Config struct {
	Logger        *slog.Logger
	OnSampleError func(error)
	InitialBackoff time.Duration
	MaxBackoffCap  time.Duration
}

// Scheduler drives periodic sampling with interval, jitter, and an optional gate.
type Scheduler struct {
	Interval time.Duration
	Jitter   time.Duration
	Gate     collector.SampleGate
	Config
}

// New returns a Scheduler with the given configuration.
// Gate defaults to collector.AlwaysAllow when nil.
func New(interval, jitter time.Duration, gate collector.SampleGate, cfg Config) *Scheduler {
	if gate == nil {
		gate = collector.AlwaysAllow
	}
	if cfg.InitialBackoff <= 0 {
		cfg.InitialBackoff = defaultInitialBackoff
	}
	if cfg.MaxBackoffCap <= 0 {
		cfg.MaxBackoffCap = defaultMaxBackoffCap
	}
	return &Scheduler{
		Interval: interval,
		Jitter:   jitter,
		Gate:     gate,
		Config:   cfg,
	}
}

func (s *Scheduler) maxBackoff() time.Duration {
	cap := s.MaxBackoffCap
	if s.Interval > 0 && s.Interval < cap {
		cap = s.Interval
	}
	return cap
}

// Run ticks on interval, evaluates the gate, and calls fn when permitted.
// Transient sample errors are logged and retried with exponential backoff.
func (s *Scheduler) Run(ctx context.Context, fn func(context.Context) error) error {
	if s.Interval <= 0 {
		return nil
	}

	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	backoff := s.InitialBackoff

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if s.Jitter > 0 {
				jitter := time.Duration(rand.Int64N(int64(s.Jitter)))
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(jitter):
				}
			}

			ok, err := s.Gate(ctx)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}

			if err := fn(ctx); err != nil {
				if s.OnSampleError != nil {
					s.OnSampleError(err)
				} else if s.Logger != nil {
					s.Logger.Warn("sample failed", "err", err)
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(backoff):
				}
				if next := backoff * 2; next > s.maxBackoff() {
					backoff = s.maxBackoff()
				} else {
					backoff = next
				}
				continue
			}

			backoff = s.InitialBackoff
		}
	}
}
