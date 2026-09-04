package promexport

import (
	"testing"

	"github.com/KoJaco/leakwatch/internal/analysis"
	"github.com/prometheus/client_golang/prometheus"
)

type stubProvider struct{}

func (stubProvider) Snapshot() analysis.Snapshot {
	return analysis.Snapshot{}
}

func TestCollectorRegistersMetrics(t *testing.T) {
	c := NewCollector(stubProvider{})
	reg := prometheus.NewRegistry()
	if err := reg.Register(c); err != nil {
		t.Fatalf("register: %v", err)
	}
}
