package observer

import (
	"sync"
	"testing"
	"time"

	"github.com/KoJaco/leakwatch/internal/domain"
	"github.com/KoJaco/leakwatch/internal/fingerprint"
)

func TestHistoryRecord_ringBuffer(t *testing.T) {
	h := NewHistory(2)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	if err := h.Record(obsAt(base, "first")); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := h.Record(obsAt(base.Add(time.Minute), "second")); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := h.Record(obsAt(base.Add(2*time.Minute), "third")); err != nil {
		t.Fatalf("Record: %v", err)
	}

	got := h.Observations()
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Clusters[0].Fingerprint.ID != "second" {
		t.Fatalf("oldest retained ID = %q, want second", got[0].Clusters[0].Fingerprint.ID)
	}
	if got[1].Clusters[0].Fingerprint.ID != "third" {
		t.Fatalf("newest retained ID = %q, want third", got[1].Clusters[0].Fingerprint.ID)
	}
}

func TestHistoryRecord_since(t *testing.T) {
	h := NewHistory(4)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	for i, id := range []string{"a", "b", "c"} {
		if err := h.Record(obsAt(base.Add(time.Duration(i)*time.Minute), id)); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}

	since := h.Since(base.Add(time.Minute))
	if len(since) != 2 {
		t.Fatalf("len(Since) = %d, want 2", len(since))
	}
	if since[0].Clusters[0].Fingerprint.ID != "b" {
		t.Fatalf("first Since ID = %q, want b", since[0].Clusters[0].Fingerprint.ID)
	}
}

func TestHistoryRecord_observationsCopy(t *testing.T) {
	h := NewHistory(4)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if err := h.Record(obsAt(base, "a")); err != nil {
		t.Fatalf("Record: %v", err)
	}

	got := h.Observations()
	got[0].CapturedAt = base.Add(time.Hour)
	if h.Observations()[0].CapturedAt.Equal(base.Add(time.Hour)) {
		t.Fatal("mutating returned slice affected stored history")
	}
}

func TestHistory_concurrentRecordAndRead(t *testing.T) {
	h := NewHistory(128)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = h.Observations()
				_ = h.Since(base)
			}
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			if err := h.Record(obsAt(base.Add(time.Duration(i)*time.Second), "id")); err != nil {
				t.Errorf("Record: %v", err)
			}
		}
	}()

	wg.Wait()
}

func obsAt(at time.Time, id string) domain.Observation {
	return domain.Observation{
		CapturedAt: at,
		Clusters: []fingerprint.LeakCluster{
			{Fingerprint: fingerprint.Fingerprint{ID: id, Key: id}, Count: 1},
		},
	}
}
