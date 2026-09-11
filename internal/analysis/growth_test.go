package analysis

import (
	"math"
	"testing"
)

func TestGrowthFor_growing(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 10, 20, 30)
	growth := GrowthFor(observations, "abc123", DefaultGrowthWindow)
	if growth.RatePerMinute <= 0 {
		t.Fatalf("RatePerMinute = %v, want > 0", growth.RatePerMinute)
	}
	if growth.RatePerHour <= 0 {
		t.Fatalf("RatePerHour = %v, want > 0", growth.RatePerHour)
	}
}

func TestGrowthFor_persistent(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 100, 100, 100)
	growth := GrowthFor(observations, "abc123", DefaultGrowthWindow)
	if math.Abs(growth.RatePerMinute) > growthEpsilon {
		t.Fatalf("RatePerMinute = %v, want 0", growth.RatePerMinute)
	}
	if math.Abs(growth.RatePerHour) > growthEpsilon {
		t.Fatalf("RatePerHour = %v, want 0", growth.RatePerHour)
	}
}

func TestGrowthFor_singlePoint(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 10)
	growth := GrowthFor(observations, "abc123", DefaultGrowthWindow)
	if growth.RatePerMinute != 0 || growth.RatePerHour != 0 {
		t.Fatalf("growth = %+v, want zero rates", growth)
	}
}

func TestGrowthFor_absentInWindow(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 10, -1, -1)
	growth := GrowthFor(observations, "abc123", DefaultGrowthWindow)
	if growth.RatePerMinute != 0 || growth.RatePerHour != 0 {
		t.Fatalf("growth = %+v, want zero rates", growth)
	}
}

func TestGrowthFor_empty(t *testing.T) {
	growth := GrowthFor(nil, "abc123", DefaultGrowthWindow)
	if growth.RatePerMinute != 0 || growth.RatePerHour != 0 {
		t.Fatalf("growth = %+v, want zero rates", growth)
	}
}
