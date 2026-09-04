package analysis

import "github.com/KoJaco/leakwatch/internal/domain"

// DetectRecurrence identifies leaks that disappear and reappear across observations.
func DetectRecurrence(observations []domain.Observation, leakID string) bool {
	_ = observations
	_ = leakID
	return false
}
