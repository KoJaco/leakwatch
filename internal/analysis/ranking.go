package analysis

import (
	"math"
	"sort"

	"github.com/KoJaco/leakwatch/internal/domain"
)

// RankBySeverity sorts leaks by severity for display and alerting.
func RankBySeverity(leaks []Leak, observations []domain.Observation) []Leak {
	if len(leaks) == 0 {
		return nil
	}
	out := make([]Leak, len(leaks))
	copy(out, leaks)
	sort.SliceStable(out, func(i, j int) bool {
		si := severityScore(out[i], observations)
		sj := severityScore(out[j], observations)
		if si != sj {
			return si > sj
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func severityScore(leak Leak, observations []domain.Observation) float64 {
	growth := GrowthFor(observations, leak.ID, DefaultGrowthWindow)
	rate := math.Max(growth.RatePerMinute, 0)
	return float64(leak.Count) * leak.Persistence.Seconds() * (1 + rate)
}
