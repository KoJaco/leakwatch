package analysis

import "github.com/KoJaco/leakwatch/internal/domain"

// DetectRecurrence identifies leaks that disappear and reappear across observations.
func DetectRecurrence(observations []domain.Observation, leakID string) bool {
	if len(observations) == 0 || leakID == "" {
		return false
	}

	sawPresent := false
	sawGap := false

	for _, observation := range observations {
		present := hasLeak(observation, leakID)
		switch {
		case !sawPresent && present:
			sawPresent = true
		case sawPresent && !present:
			sawGap = true
		case sawPresent && sawGap && present:
			return true
		}
	}
	return false
}

func hasLeak(observation domain.Observation, leakID string) bool {
	_, ok := clusterCount(observation, leakID)
	return ok
}
