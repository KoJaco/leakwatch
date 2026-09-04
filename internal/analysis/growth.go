package analysis

import (
	"time"

	"github.com/KoJaco/leakwatch/internal/domain"
)

// Growth holds derived growth rates over a specific window.
type Growth struct {
	RatePerMinute float64
	RatePerHour   float64
}

// GrowthFor computes growth over a specific window — not a single canonical rate.
func GrowthFor(observations []domain.Observation, leakID string, window time.Duration) Growth {
	_ = observations
	_ = leakID
	_ = window
	return Growth{}
}
