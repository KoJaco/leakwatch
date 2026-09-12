package promexport

import (
	"strings"
	"testing"
	"time"

	"github.com/KoJaco/leakwatch/internal/analysis"
	"github.com/KoJaco/leakwatch/internal/domain"
	"github.com/KoJaco/leakwatch/internal/fingerprint"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type stubProvider struct{}

func (stubProvider) Snapshot() analysis.Snapshot {
	return analysis.Snapshot{}
}

func (stubProvider) Observations() []domain.Observation {
	return nil
}

func TestCollectorRegistersMetrics(t *testing.T) {
	c := NewCollector(stubProvider{})
	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("register: %v", err)
	}
}

type growthProvider struct {
	snap analysis.Snapshot
	obs  []domain.Observation
}

func (p growthProvider) Snapshot() analysis.Snapshot {
	return p.snap
}

func (p growthProvider) Observations() []domain.Observation {
	return p.obs
}

func TestCollectorGrowthRate(t *testing.T) {
	const leakID = "abc123"
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	observations := sequenceObs(leakID, base, 10, 20, 30)
	leaks := analysis.Analyze(observations)
	if len(leaks) != 1 {
		t.Fatalf("len(leaks) = %d, want 1", len(leaks))
	}

	p := growthProvider{
		snap: analysis.Snapshot{
			CapturedAt: observations[len(observations)-1].CapturedAt,
			Leaks:      leaks,
		},
		obs: observations,
	}
	c := NewCollector(p)

	expected := `
		# HELP goroutine_leak_growth_rate Growth rate per minute for the leak cluster (derived at scrape time).
		# TYPE goroutine_leak_growth_rate gauge
		goroutine_leak_growth_rate{leak_id="abc123"} 2
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "goroutine_leak_growth_rate"); err != nil {
		t.Fatal(err)
	}
}

func TestCollectorGrowthRate_zeroWithInsufficientHistory(t *testing.T) {
	const leakID = "abc123"
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	observations := sequenceObs(leakID, base, 10)
	leaks := analysis.Analyze(observations)
	if len(leaks) != 1 {
		t.Fatalf("len(leaks) = %d, want 1", len(leaks))
	}

	p := growthProvider{
		snap: analysis.Snapshot{
			CapturedAt: observations[0].CapturedAt,
			Leaks:      leaks,
		},
		obs: observations,
	}
	c := NewCollector(p)

	expected := `
		# HELP goroutine_leak_growth_rate Growth rate per minute for the leak cluster (derived at scrape time).
		# TYPE goroutine_leak_growth_rate gauge
		goroutine_leak_growth_rate{leak_id="abc123"} 0
	`
	if err := testutil.CollectAndCompare(c, strings.NewReader(expected), "goroutine_leak_growth_rate"); err != nil {
		t.Fatal(err)
	}
}

func cluster(id string, count int) fingerprint.LeakCluster {
	return fingerprint.LeakCluster{
		Fingerprint: fingerprint.Fingerprint{ID: id, Key: id},
		Count:       count,
	}
}

func obs(at time.Time, clusters ...fingerprint.LeakCluster) domain.Observation {
	return domain.Observation{
		CapturedAt: at,
		Clusters:   clusters,
	}
}

func sequenceObs(id string, base time.Time, counts ...int) []domain.Observation {
	out := make([]domain.Observation, len(counts))
	for i, count := range counts {
		out[i] = obs(base.Add(time.Duration(i*5)*time.Minute), cluster(id, count))
	}
	return out
}
