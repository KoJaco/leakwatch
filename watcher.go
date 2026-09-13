package leakwatch

import (
	"context"
	stdhttp "net/http"
	"sync"
	"sync/atomic"

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
	logger    interface {
		Warn(msg string, args ...any)
		Info(msg string, args ...any)
	}

	mu              sync.Mutex
	cancel          context.CancelFunc
	lastSampleError atomic.Pointer[sampleError]
}

type sampleError struct {
	err error
}

// New creates a Watcher with the given options.
func New(opts ...Option) *Watcher {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	sourceOpts := []profile.SourceOption{
		profile.WithSourceMaxProfileBytes(cfg.maxProfileBytes),
	}
	if cfg.httpClient != nil {
		sourceOpts = append(sourceOpts, profile.WithSourceHTTPClient(cfg.httpClient))
	} else if cfg.httpTimeout > 0 {
		sourceOpts = append(sourceOpts, profile.WithSourceHTTPClient(&stdhttp.Client{
			Timeout: cfg.httpTimeout,
		}))
	}

	source := profile.NewPProfSource(cfg.pprofURL, sourceOpts...)
	parser := profile.NewParser()
	col := collector.NewPProfCollector(source, parser)
	fp := fingerprint.NewFingerprinter()

	capacity := cfg.historyCapacity
	if capacity <= 0 {
		capacity = defaultHistoryCapacity
	}
	hist := observer.NewHistory(capacity)
	obs := observer.New(col, fp, hist)

	w := &Watcher{
		observer: obs,
		logger:   cfg.logger,
	}

	w.scheduler = scheduler.New(cfg.interval, cfg.jitter, cfg.sampleGate, scheduler.Config{
		Logger: cfg.logger,
		OnSampleError: func(err error) {
			w.lastSampleError.Store(&sampleError{err: err})
			w.logger.Warn("leakwatch sample failed", "err", err)
		},
	})

	w.logger.Info("leakwatch watcher created",
		"fingerprint_version", fingerprint.FingerprintVersion,
		"pprof_url", cfg.pprofURL,
	)

	return w
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

// LastSampleError returns the most recent sample failure, if any.
func (w *Watcher) LastSampleError() error {
	if v := w.lastSampleError.Load(); v != nil {
		return v.err
	}
	return nil
}

// ServeDebug registers leak debug handlers on http.DefaultServeMux at path.
// Pass httpexport.WithMiddleware to add auth or other protection before exposing
// stack traces on a shared HTTP server.
func (w *Watcher) ServeDebug(path string, opts ...httpexport.RegisterOption) error {
	httpexport.Register(stdhttp.DefaultServeMux, path, w, opts...)
	return nil
}

var (
	_ httpexport.SnapshotProvider = (*Watcher)(nil)
	_ promexport.SnapshotProvider = (*Watcher)(nil)
)
