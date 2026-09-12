package leakwatch

import (
	"context"
	stdhttp "net/http"
	"sync"
	"time"

	httpexport "github.com/KoJaco/leakwatch/export/http"
	promexport "github.com/KoJaco/leakwatch/export/prometheus"
	"github.com/KoJaco/leakwatch/internal/analysis"
	"github.com/KoJaco/leakwatch/internal/collector"
	"github.com/KoJaco/leakwatch/internal/domain"
	"github.com/KoJaco/leakwatch/internal/fingerprint"
	"github.com/KoJaco/leakwatch/internal/observer"
	"github.com/KoJaco/leakwatch/internal/profile"
	"github.com/KoJaco/leakwatch/internal/scheduler"
)

const defaultHistoryCapacity = 256

// Watcher samples goroutine leak profiles on a schedule and maintains leak state.
type Watcher struct {
	observer  *observer.Observer
	scheduler *scheduler.Scheduler

	mu     sync.Mutex
	cancel context.CancelFunc
}

// New creates a Watcher with the given options.
func New(opts ...Option) *Watcher {
	cfg := config{
		interval:   5 * time.Minute,
		jitter:     30 * time.Second,
		sampleGate: collector.AlwaysAllow,
		pprofURL:   "http://127.0.0.1:6060/debug/pprof/goroutineleak",
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	source := profile.NewPProfSource(cfg.pprofURL)
	parser := profile.NewParser()
	col := collector.NewPProfCollector(source, parser)
	fp := fingerprint.NewFingerprinter()
	hist := observer.NewHistory(defaultHistoryCapacity)
	obs := observer.New(col, fp, hist)
	sched := scheduler.New(cfg.interval, cfg.jitter, cfg.sampleGate)

	return &Watcher{
		observer:  obs,
		scheduler: sched,
	}
}

// Start begins periodic sampling in a background goroutine.
func (w *Watcher) Start(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cancel != nil {
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	go func() {
		_ = w.scheduler.Run(runCtx, w.observer.Sample)
	}()
}

// Stop cancels periodic sampling.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
}

// Snapshot returns the current export-ready leak state.
func (w *Watcher) Snapshot() analysis.Snapshot {
	return w.observer.Snapshot()
}

// Observations returns the retained observation history.
func (w *Watcher) Observations() []domain.Observation {
	return w.observer.Observations()
}

// ServeDebug registers leak debug handlers on http.DefaultServeMux at path.
func (w *Watcher) ServeDebug(path string) error {
	httpexport.Register(stdhttp.DefaultServeMux, path, w)
	return nil
}

var (
	_ httpexport.SnapshotProvider = (*Watcher)(nil)
	_ promexport.SnapshotProvider = (*Watcher)(nil)
)
