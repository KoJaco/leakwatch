package analysis

import "testing"

func TestDetectRecurrence_presentGapPresent(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 10, -1, 10)
	if !DetectRecurrence(observations, "abc123") {
		t.Fatal("DetectRecurrence = false, want true")
	}
}

func TestDetectRecurrence_continuousPresent(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, 10, 10, 10)
	if DetectRecurrence(observations, "abc123") {
		t.Fatal("DetectRecurrence = true, want false")
	}
}

func TestDetectRecurrence_neverSeen(t *testing.T) {
	observations := sequenceObs("abc123", testBaseTime, -1, -1, -1)
	if DetectRecurrence(observations, "abc123") {
		t.Fatal("DetectRecurrence = true, want false")
	}
}

func TestDetectRecurrence_empty(t *testing.T) {
	if DetectRecurrence(nil, "abc123") {
		t.Fatal("DetectRecurrence = true, want false")
	}
}
