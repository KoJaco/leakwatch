package analysis

import (
	"math"
	"time"

	"github.com/KoJaco/leakwatch/internal/domain"
	"github.com/KoJaco/leakwatch/internal/fingerprint"
)

type leakSample struct {
	at    time.Time
	count int
	site  fingerprint.LeakSite
}

// Analyze is pure: observations in, leaks out. No hidden state.
func Analyze(observations []domain.Observation) []Leak {
	if len(observations) == 0 {
		return nil
	}

	latest := observations[len(observations)-1]
	if len(latest.Clusters) == 0 {
		return nil
	}

	history := indexLeakHistory(observations)
	leaks := make([]Leak, 0, len(latest.Clusters))
	for _, cluster := range latest.Clusters {
		id := cluster.Fingerprint.ID
		samples, ok := history[id]
		if !ok || len(samples) == 0 {
			continue
		}

		first := samples[0]
		last := samples[len(samples)-1]
		leaks = append(leaks, Leak{
			ID:          id,
			Site:        cluster.Site,
			Count:       cluster.Count,
			FirstSeen:   first.at,
			LastSeen:    last.at,
			Persistence: last.at.Sub(first.at),
			Status:      classifyStatus(observations, id, len(samples)),
		})
	}

	return RankBySeverity(leaks, observations)
}

func indexLeakHistory(observations []domain.Observation) map[string][]leakSample {
	history := make(map[string][]leakSample)
	for _, observation := range observations {
		for _, cluster := range observation.Clusters {
			id := cluster.Fingerprint.ID
			history[id] = append(history[id], leakSample{
				at:    observation.CapturedAt,
				count: cluster.Count,
				site:  cluster.Site,
			})
		}
	}
	return history
}

func classifyStatus(observations []domain.Observation, leakID string, presenceCount int) Status {
	if DetectRecurrence(observations, leakID) {
		return StatusRecurring
	}
	if presenceCount == 1 {
		return StatusNew
	}

	growth := GrowthFor(observations, leakID, DefaultGrowthWindow)
	if growth.RatePerMinute > growthEpsilon {
		return StatusGrowing
	}
	if presenceCount >= 2 && math.Abs(growth.RatePerMinute) <= growthEpsilon {
		return StatusPersistent
	}
	return StatusNew
}
