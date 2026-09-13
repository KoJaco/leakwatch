package scheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KoJaco/leakwatch/internal/collector"
)

func TestRun_continuesAfterSampleError(t *testing.T) {
	var calls atomic.Int32
	s := New(10*time.Millisecond, 0, collector.AlwaysAllow, Config{
		InitialBackoff: time.Millisecond,
		MaxBackoffCap:  time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := s.Run(ctx, func(context.Context) error {
		n := calls.Add(1)
		if n == 1 {
			return errors.New("transient failure")
		}
		return nil
	})
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() = %v, want context error", err)
	}
	if calls.Load() < 2 {
		t.Fatalf("calls = %d, want at least 2", calls.Load())
	}
}

func TestRun_gateErrorStops(t *testing.T) {
	s := New(time.Millisecond, 0, func(context.Context) (bool, error) {
		return false, errors.New("gate failure")
	}, Config{})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := s.Run(ctx, func(context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected gate error")
	}
}
