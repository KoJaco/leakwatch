package pipeline

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	pprof "github.com/google/pprof/profile"

	"github.com/KoJaco/leakwatch/internal/analysis"
)

func TestSnapshotFromProfile(t *testing.T) {
	data := sampleProfileBytes(t, 3, "main.leak", "/build/main.go", 42)
	prof, err := ProfileFromFile(writeTempProfile(t, data))
	if err != nil {
		t.Fatalf("ProfileFromFile: %v", err)
	}

	snap, err := SnapshotFromProfile(prof)
	if err != nil {
		t.Fatalf("SnapshotFromProfile: %v", err)
	}
	if len(snap.Leaks) != 1 {
		t.Fatalf("len(leaks) = %d, want 1", len(snap.Leaks))
	}
	if snap.Leaks[0].Status != analysis.StatusNew {
		t.Fatalf("status = %q, want %q", snap.Leaks[0].Status, analysis.StatusNew)
	}
	if snap.Leaks[0].Count != 3 {
		t.Fatalf("count = %d, want 3", snap.Leaks[0].Count)
	}
}

func TestProfileFromURL(t *testing.T) {
	data := sampleProfileBytes(t, 2, "main.leak", "/build/main.go", 42)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	prof, err := ProfileFromURL(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("ProfileFromURL: %v", err)
	}
	if len(prof.Goroutines) != 2 {
		t.Fatalf("len(goroutines) = %d, want 2", len(prof.Goroutines))
	}
}

func sampleProfileBytes(t *testing.T, count int64, fn, file string, line int64) []byte {
	t.Helper()

	runtimeFn := &pprof.Function{ID: 1, Name: "runtime.gopark", Filename: "runtime/proc.go"}
	appFn := &pprof.Function{ID: 2, Name: fn, Filename: file}
	loc := &pprof.Location{
		ID: 1,
		Line: []pprof.Line{
			{Function: runtimeFn, Line: 100},
			{Function: appFn, Line: line},
		},
	}
	prof := &pprof.Profile{
		TimeNanos:  time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC).UnixNano(),
		SampleType: []*pprof.ValueType{{Type: "goroutineleak", Unit: "count"}},
		Function:   []*pprof.Function{runtimeFn, appFn},
		Location:   []*pprof.Location{loc},
		Sample: []*pprof.Sample{{
			Value:    []int64{count},
			Location: []*pprof.Location{loc},
		}},
	}

	var buf bytes.Buffer
	if err := prof.Write(&buf); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	return buf.Bytes()
}

func writeTempProfile(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "leak.pb")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	return path
}
