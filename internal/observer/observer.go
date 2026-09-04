package observer

import (
	"context"
	"time"

	"github.com/KoJaco/leakwatch/internal/analysis"
	"github.com/KoJaco/leakwatch/internal/collector"
	"github.com/KoJaco/leakwatch/internal/domain"
	"github.com/KoJaco/leakwatch/internal/fingerprint"
)

// Observer orchestrates one sample cycle: collect → fingerprint → record.
type Observer struct {
	collector     collector.Collector
	fingerprinter fingerprint.Fingerprinter
	history       *History
}

// New returns an Observer wired to the given dependencies.
func New(c collector.Collector, fp fingerprint.Fingerprinter, h *History) *Observer {
	return &Observer{
		collector:     c,
		fingerprinter: fp,
		history:       h,
	}
}

// Sample performs one full observation cycle.
func (o *Observer) Sample(ctx context.Context) error {
	prof, err := o.collector.Collect(ctx)
	if err != nil {
		return err
	}

	clusters, err := o.fingerprinter.Fingerprint(prof)
	if err != nil {
		return err
	}

	return o.history.Record(domain.Observation{
		CapturedAt: time.Now().UTC(),
		Clusters:   clusters,
	})
}

// Leaks returns the current interpreted state via pure analysis.
func (o *Observer) Leaks() []analysis.Leak {
	return analysis.Analyze(o.history.Observations())
}

// Snapshot returns the current export-ready state.
func (o *Observer) Snapshot() analysis.Snapshot {
	leaks := o.Leaks()
	capturedAt := time.Now().UTC()
	if obs := o.history.Observations(); len(obs) > 0 {
		capturedAt = obs[len(obs)-1].CapturedAt
	}
	return analysis.Snapshot{
		CapturedAt: capturedAt,
		Leaks:      leaks,
	}
}
