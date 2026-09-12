package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	pprof "github.com/google/pprof/profile"
)

func TestRunAnalyze_success(t *testing.T) {
	path := filepath.Join("testdata", "leak.pb")
	out := captureStdout(t, func() {
		if got := runAnalyze([]string{path}); got != 0 {
			t.Fatalf("runAnalyze() = %d, want 0", got)
		}
	})
	if !bytes.Contains(out, []byte("Status: new")) {
		t.Fatalf("stdout = %q, want status new", out)
	}
	if !bytes.Contains(out, []byte("Count: 3")) {
		t.Fatalf("stdout = %q, want count 3", out)
	}
}

func TestRunAnalyze_json(t *testing.T) {
	path := filepath.Join("testdata", "leak.pb")
	out := captureStdout(t, func() {
		if got := runAnalyze([]string{"-json", path}); got != 0 {
			t.Fatalf("runAnalyze() = %d, want 0", got)
		}
	})
	if !bytes.Contains(out, []byte(`"CapturedAt"`)) {
		t.Fatalf("stdout = %q, want CapturedAt JSON field", out)
	}
	if !bytes.Contains(out, []byte(`"Leaks"`)) {
		t.Fatalf("stdout = %q, want Leaks JSON field", out)
	}
}

func TestRunAnalyze_missingFile(t *testing.T) {
	if got := runAnalyze([]string{"testdata/does-not-exist.pb"}); got != 1 {
		t.Fatalf("runAnalyze() = %d, want 1", got)
	}
}

func TestRunAnalyze_usage(t *testing.T) {
	if got := runAnalyze(nil); got != 1 {
		t.Fatalf("runAnalyze() = %d, want 1", got)
	}
}

func TestWriteFixtureProfile(t *testing.T) {
	if os.Getenv("WRITE_FIXTURE") != "1" {
		t.Skip("set WRITE_FIXTURE=1 to regenerate testdata/leak.pb")
	}
	data := fixtureProfileBytes(t)
	path := filepath.Join("testdata", "leak.pb")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func fixtureProfileBytes(t *testing.T) []byte {
	t.Helper()

	runtimeFn := &pprof.Function{ID: 1, Name: "runtime.gopark", Filename: "runtime/proc.go"}
	chanFn := &pprof.Function{ID: 2, Name: "runtime.chanrecv1", Filename: "runtime/chan.go"}
	appFn := &pprof.Function{ID: 3, Name: "main.leak", Filename: "/build/main.go"}
	mainFn := &pprof.Function{ID: 4, Name: "main.main", Filename: "/build/main.go"}
	loc := &pprof.Location{
		ID: 1,
		Line: []pprof.Line{
			{Function: runtimeFn, Line: 100},
			{Function: chanFn, Line: 50},
			{Function: appFn, Line: 42},
			{Function: mainFn, Line: 20},
		},
	}
	prof := &pprof.Profile{
		TimeNanos:  time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC).UnixNano(),
		SampleType: []*pprof.ValueType{{Type: "goroutineleak", Unit: "count"}},
		Function:   []*pprof.Function{runtimeFn, chanFn, appFn, mainFn},
		Location:   []*pprof.Location{loc},
		Sample: []*pprof.Sample{{
			Value:    []int64{3},
			Location: []*pprof.Location{loc},
		}},
	}

	var buf bytes.Buffer
	if err := prof.Write(&buf); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	return buf.Bytes()
}

func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()
	return buf.Bytes()
}
