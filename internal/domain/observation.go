package domain

import (
	"time"

	"github.com/KoJaco/leakwatch/internal/fingerprint"
)

// Observation is a timestamped set of cluster counts — raw historical data.
type Observation struct {
	CapturedAt time.Time
	Clusters   []fingerprint.LeakCluster
}
