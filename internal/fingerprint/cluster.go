package fingerprint

import "github.com/KoJaco/leakwatch/internal/profile"

// ClusterGoroutines groups goroutines by normalized stack similarity.
func ClusterGoroutines(goroutines []profile.Goroutine) []LeakCluster {
	_ = goroutines
	return nil
}
