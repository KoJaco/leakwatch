package analysis

import (
	"testing"
	"time"

	"github.com/KoJaco/leakwatch/internal/domain"
)

func TestAnalyze_empty(t *testing.T) {
	if got := Analyze(nil); got != nil {
		t.Fatalf("Analyze(nil) = %v, want nil", got)
	}
}

func TestAnalyze_new(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 5)
	leaks := Analyze(observations)
	if len(leaks) != 1 {
		t.Fatalf("len(leaks) = %d, want 1", len(leaks))
	}
	if leaks[0].Status != StatusNew {
		t.Fatalf("status = %q, want %q", leaks[0].Status, StatusNew)
	}
	if leaks[0].Count != 5 {
		t.Fatalf("count = %d, want 5", leaks[0].Count)
	}
}

func TestAnalyze_growing(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 10, 20, 30)
	leaks := Analyze(observations)
	if len(leaks) != 1 {
		t.Fatalf("len(leaks) = %d, want 1", len(leaks))
	}
	if leaks[0].Status != StatusGrowing {
		t.Fatalf("status = %q, want %q", leaks[0].Status, StatusGrowing)
	}
}

func TestAnalyze_persistent(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 100, 100, 100)
	leaks := Analyze(observations)
	if len(leaks) != 1 {
		t.Fatalf("len(leaks) = %d, want 1", len(leaks))
	}
	if leaks[0].Status != StatusPersistent {
		t.Fatalf("status = %q, want %q", leaks[0].Status, StatusPersistent)
	}
}

func TestAnalyze_recurring(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 10, -1, 10)
	leaks := Analyze(observations)
	if len(leaks) != 1 {
		t.Fatalf("len(leaks) = %d, want 1", len(leaks))
	}
	if leaks[0].Status != StatusRecurring {
		t.Fatalf("status = %q, want %q", leaks[0].Status, StatusRecurring)
	}
}

func TestAnalyze_resolvedExcluded(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 10, 10, -1)
	leaks := Analyze(observations)
	if leaks != nil {
		t.Fatalf("Analyze() = %v, want nil", leaks)
	}
}

func TestAnalyze_firstSeenLastSeen(t *testing.T) {
	times := minutes(testBaseTime, 0, 5, 10)
	observations := []domain.Observation{
		obs(times[0], cluster("abc123", 10)),
		obs(times[1], cluster("abc123", 20)),
		obs(times[2], cluster("abc123", 30)),
	}
	leaks := Analyze(observations)
	if len(leaks) != 1 {
		t.Fatalf("len(leaks) = %d, want 1", len(leaks))
	}
	if !leaks[0].FirstSeen.Equal(times[0]) {
		t.Fatalf("FirstSeen = %v, want %v", leaks[0].FirstSeen, times[0])
	}
	if !leaks[0].LastSeen.Equal(times[2]) {
		t.Fatalf("LastSeen = %v, want %v", leaks[0].LastSeen, times[2])
	}
	if leaks[0].Persistence != 10*time.Minute {
		t.Fatalf("Persistence = %v, want 10m", leaks[0].Persistence)
	}
}

func TestAnalyze_multipleLeaks(t *testing.T) {
	times := minutes(testBaseTime, 0, 5, 10)
	observations := []domain.Observation{
		obs(times[0], cluster("aaa", 10), cluster("bbb", 5)),
		obs(times[1], cluster("aaa", 20), cluster("bbb", 5)),
		obs(times[2], cluster("aaa", 30), cluster("bbb", 5)),
	}
	leaks := Analyze(observations)
	if len(leaks) != 2 {
		t.Fatalf("len(leaks) = %d, want 2", len(leaks))
	}
	if leaks[0].ID != "aaa" {
		t.Fatalf("first leak ID = %q, want aaa", leaks[0].ID)
	}
	if leaks[0].Status != StatusGrowing {
		t.Fatalf("aaa status = %q, want %q", leaks[0].Status, StatusGrowing)
	}
	if leaks[1].ID != "bbb" {
		t.Fatalf("second leak ID = %q, want bbb", leaks[1].ID)
	}
	if leaks[1].Status != StatusPersistent {
		t.Fatalf("bbb status = %q, want %q", leaks[1].Status, StatusPersistent)
	}
}

func TestRankBySeverity_prefersHigherSeverity(t *testing.T) {
	times := minutes(testBaseTime, 0, 5, 10)
	observations := []domain.Observation{
		obs(times[0], cluster("aaa", 10), cluster("bbb", 5)),
		obs(times[1], cluster("aaa", 20), cluster("bbb", 5)),
		obs(times[2], cluster("aaa", 30), cluster("bbb", 5)),
	}
	leaks := Analyze(observations)
	if len(leaks) != 2 {
		t.Fatalf("len(leaks) = %d, want 2", len(leaks))
	}
	if leaks[0].ID != "aaa" {
		t.Fatalf("first ranked leak = %q, want aaa", leaks[0].ID)
	}
}
