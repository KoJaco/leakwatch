package analysis

import (
	"time"

	"github.com/KoJaco/leakwatch/internal/domain"
)

const (
	// DefaultGrowthWindow is the default lookback for growth and status classification.
	DefaultGrowthWindow = 30 * time.Minute

	growthEpsilon = 1e-9
)

// Growth holds derived growth rates over a specific window.
type Growth struct {
	RatePerMinute float64
	RatePerHour   float64
}

type leakPoint struct {
	at    time.Time
	count int
}

// GrowthFor computes growth over a specific window — not a single canonical rate.
func GrowthFor(observations []domain.Observation, leakID string, window time.Duration) Growth {
	if len(observations) == 0 || leakID == "" {
		return Growth{}
	}

	latest := observations[len(observations)-1].CapturedAt
	windowStart := latest.Add(-window)

	points := leakPointsInWindow(observations, leakID, windowStart)
	if len(points) < 2 {
		return Growth{}
	}

	first := points[0]
	last := points[len(points)-1]
	duration := last.at.Sub(first.at)
	if duration <= 0 {
		return Growth{}
	}

	delta := float64(last.count - first.count)
	return Growth{
		RatePerMinute: delta / duration.Minutes(),
		RatePerHour:   delta / duration.Hours(),
	}
}

func leakPointsInWindow(observations []domain.Observation, leakID string, windowStart time.Time) []leakPoint {
	var points []leakPoint
	for _, observation := range observations {
		if observation.CapturedAt.Before(windowStart) {
			continue
		}
		count, ok := clusterCount(observation, leakID)
		if !ok {
			continue
		}
		points = append(points, leakPoint{at: observation.CapturedAt, count: count})
	}
	return points
}

func clusterCount(observation domain.Observation, leakID string) (int, bool) {
	for _, c := range observation.Clusters {
		if c.Fingerprint.ID == leakID {
			return c.Count, true
		}
	}
	return 0, false
}
