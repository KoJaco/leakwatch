package observer

import (
	"sync"
	"time"

	"github.com/KoJaco/leakwatch/internal/domain"
)

// History is a concrete in-memory ring buffer for observations.
type History struct {
	capacity int
	entries  []domain.Observation
	mu       sync.RWMutex
}

// NewHistory returns an in-memory history with the given capacity.
func NewHistory(capacity int) *History {
	return &History{
		capacity: capacity,
		entries:  make([]domain.Observation, 0, capacity),
	}
}

// Record appends an observation to history.
func (h *History) Record(obs domain.Observation) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.capacity > 0 && len(h.entries) >= h.capacity {
		h.entries = h.entries[1:]
	}
	h.entries = append(h.entries, obs)
	return nil
}

// Since returns observations captured at or after t.
func (h *History) Since(t time.Time) []domain.Observation {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var out []domain.Observation
	for _, obs := range h.entries {
		if !obs.CapturedAt.Before(t) {
			out = append(out, obs)
		}
	}
	return out
}

// Observations returns all retained observations.
func (h *History) Observations() []domain.Observation {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make([]domain.Observation, len(h.entries))
	copy(out, h.entries)
	return out
}
