package pipeline

import (
	"context"
	"os"

	"github.com/KoJaco/leakwatch/internal/analysis"
	"github.com/KoJaco/leakwatch/internal/collector"
	"github.com/KoJaco/leakwatch/internal/domain"
	"github.com/KoJaco/leakwatch/internal/fingerprint"
	"github.com/KoJaco/leakwatch/internal/profile"
)

// ProfileFromFile reads path and parses it with the default parser.
func ProfileFromFile(path string) (profile.Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return profile.Profile{}, err
	}
	return profile.NewParser().Parse(data)
}

// ProfileFromURL fetches and parses a profile from a pprof HTTP endpoint.
func ProfileFromURL(ctx context.Context, url string) (profile.Profile, error) {
	col := collector.NewPProfCollector(profile.NewPProfSource(url), profile.NewParser())
	return col.Collect(ctx)
}

// SnapshotFromProfile fingerprints a profile and analyzes a single observation.
func SnapshotFromProfile(prof profile.Profile) (analysis.Snapshot, error) {
	clusters, err := fingerprint.NewFingerprinter().Fingerprint(prof)
	if err != nil {
		return analysis.Snapshot{}, err
	}

	obs := domain.Observation{
		CapturedAt: prof.CapturedAt,
		Clusters:   clusters,
	}
	leaks := analysis.Analyze([]domain.Observation{obs})
	return analysis.Snapshot{
		CapturedAt: prof.CapturedAt,
		Leaks:      leaks,
	}, nil
}
