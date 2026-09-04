package analysis

import "github.com/KoJaco/leakwatch/internal/domain"

// Analyze is pure: observations in, leaks out. No hidden state.
func Analyze(observations []domain.Observation) []Leak {
	_ = observations
	return nil
}
