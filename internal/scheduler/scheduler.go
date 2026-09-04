package scheduler

import (
	"context"
	"time"

	"github.com/KoJaco/leakwatch/internal/collector"
)

// Scheduler drives periodic sampling with interval, jitter, and an optional gate.
type Scheduler struct {
	Interval time.Duration
	Jitter   time.Duration
	Gate     collector.SampleGate
}

// New returns a Scheduler with the given configuration.
// Gate defaults to collector.AlwaysAllow when nil.
func New(interval, jitter time.Duration, gate collector.SampleGate) *Scheduler {
	if gate == nil {
		gate = collector.AlwaysAllow
	}
	return &Scheduler{
		Interval: interval,
		Jitter:   jitter,
		Gate:     gate,
	}
}

// Run ticks on interval, evaluates the gate, and calls fn when permitted.
func (s *Scheduler) Run(ctx context.Context, fn func(context.Context) error) error {
	if s.Interval <= 0 {
		return nil
	}

	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if s.Jitter > 0 {
				jitter := time.Duration(int64(s.Jitter) * int64(time.Now().UnixNano()%100) / 100)
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
				return err
			}
		}
	}
}
