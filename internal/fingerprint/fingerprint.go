package fingerprint

import (
	"github.com/KoJaco/leakwatch/internal/profile"
)

// Location is a human-readable attribution point for a leak.
type Location struct {
	Function string
	File     string
	Line     int
}

// Fingerprint is stable machine identity — not human-readable attribution.
type Fingerprint struct {
	ID  string // e.g. "1f246299e3674466"
	Key string // canonical representation used to generate ID
}

// LeakSite holds human-readable explanation for a fingerprint.
type LeakSite struct {
	Location Location
	Stack    []profile.Frame
}

// LeakCluster groups goroutines sharing a fingerprint at a point in time.
type LeakCluster struct {
	Fingerprint Fingerprint
	Site        LeakSite
	Count       int
}

// Fingerprinter clusters a profile into leak clusters.
type Fingerprinter interface {
	Fingerprint(p profile.Profile) ([]LeakCluster, error)
}

// DefaultFingerprinter is the standard fingerprinting implementation.
type DefaultFingerprinter struct{}

// NewFingerprinter returns the default fingerprinter.
func NewFingerprinter() *DefaultFingerprinter {
	return &DefaultFingerprinter{}
}

// Fingerprint clusters goroutines in the profile.
func (f *DefaultFingerprinter) Fingerprint(p profile.Profile) ([]LeakCluster, error) {
	if len(p.Samples) == 0 {
		return nil, nil
	}
	return ClusterSamples(p.Samples), nil
}
