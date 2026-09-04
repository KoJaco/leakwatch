package analysis

import (
	"time"

	"github.com/KoJaco/leakwatch/internal/fingerprint"
)

// Status describes the current interpreted state of a leak cluster.
type Status string

const (
	StatusNew        Status = "new"
	StatusGrowing    Status = "growing"
	StatusPersistent Status = "persistent"
	StatusRecurring  Status = "recurring"
	StatusResolved   Status = "resolved"
)

// Leak is what leakwatch currently knows about a cluster — not a raw runtime artifact.
type Leak struct {
	ID          string
	Site        fingerprint.LeakSite
	Count       int
	FirstSeen   time.Time
	LastSeen    time.Time
	Persistence time.Duration
	Status      Status
}

// Snapshot is the export boundary — downstream consumers read this, never analyze.
type Snapshot struct {
	CapturedAt time.Time
	Leaks      []Leak
}
