package promexport

import (
	"github.com/KoJaco/leakwatch/internal/analysis"
	"github.com/KoJaco/leakwatch/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
)

// SnapshotProvider supplies the current leak snapshot and observation history for metric export.
type SnapshotProvider interface {
	Snapshot() analysis.Snapshot
	Observations() []domain.Observation
}

var (
	leakClustersDesc = prometheus.NewDesc(
		"goroutine_leak_clusters",
		"Number of active goroutine leak clusters.",
		nil, nil,
	)
	leakCountDesc = prometheus.NewDesc(
		"goroutine_leak_count",
		"Goroutine count per leak cluster.",
		[]string{"leak_id"}, nil,
	)
	leakFirstSeenDesc = prometheus.NewDesc(
		"goroutine_leak_first_seen_timestamp",
		"Unix timestamp when the leak cluster was first observed.",
		[]string{"leak_id"}, nil,
	)
	leakGrowthRateDesc = prometheus.NewDesc(
		"goroutine_leak_growth_rate",
		"Growth rate per minute for the leak cluster (derived at scrape time).",
		[]string{"leak_id"}, nil,
	)
)

// Collector exposes goroutine leak metrics to Prometheus.
type Collector struct {
	provider SnapshotProvider
}

// NewCollector returns a Prometheus collector backed by provider.
func NewCollector(provider SnapshotProvider) *Collector {
	return &Collector{provider: provider}
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- leakClustersDesc
	ch <- leakCountDesc
	ch <- leakFirstSeenDesc
	ch <- leakGrowthRateDesc
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	snap := c.provider.Snapshot()
	observations := c.provider.Observations()

	ch <- prometheus.MustNewConstMetric(
		leakClustersDesc,
		prometheus.GaugeValue,
		float64(len(snap.Leaks)),
	)

	for _, leak := range snap.Leaks {
		ch <- prometheus.MustNewConstMetric(
			leakCountDesc,
			prometheus.GaugeValue,
			float64(leak.Count),
			leak.ID,
		)
		if !leak.FirstSeen.IsZero() {
			ch <- prometheus.MustNewConstMetric(
				leakFirstSeenDesc,
				prometheus.GaugeValue,
				float64(leak.FirstSeen.Unix()),
				leak.ID,
			)
		}
		growth := analysis.GrowthFor(observations, leak.ID, analysis.DefaultGrowthWindow)
		ch <- prometheus.MustNewConstMetric(
			leakGrowthRateDesc,
			prometheus.GaugeValue,
			growth.RatePerMinute,
			leak.ID,
		)
	}
}
