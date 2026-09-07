package profile

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pprof "github.com/google/pprof/profile"
)

// --- helpers ---

// writeProfile serializes a pprof.Profile the same way the runtime would
// (protobuf, optionally gzip — Write handles both).
func writeProfile(t *testing.T, prof *pprof.Profile) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := prof.Write(&buf); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	return buf.Bytes()
}

// sampleProfile builds a minimal goroutineleak-style count profile for tests.
func sampleProfile(count int64, fn string, file string, line int64) *pprof.Profile {
	fnObj := &pprof.Function{
		ID:       1,
		Name:     fn,
		Filename: file,
	}
	loc := &pprof.Location{
		ID: 1,
		Line: []pprof.Line{{
			Function: fnObj,
			Line:     line,
		}},
	}
	return &pprof.Profile{
		TimeNanos:  time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC).UnixNano(),
		SampleType: []*pprof.ValueType{{Type: "goroutineleak", Unit: "count"}},
		Function:   []*pprof.Function{fnObj},
		Location:   []*pprof.Location{loc},
		Sample: []*pprof.Sample{{
			Value:    []int64{count},
			Location: []*pprof.Location{loc},
		}},
	}
}

// --- Parse tests ---

func TestSampleCount(t *testing.T) {
	cases := []struct {
		name  string
		value []int64
		want  int
	}{
		{name: "nil value", value: nil, want: 0},
		{name: "empty value", value: []int64{}, want: 0},
		{name: "zero count", value: []int64{0}, want: 0},
		{name: "negative count", value: []int64{-5}, want: 0},
		{name: "positive count", value: []int64{3}, want: 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sampleCount(&pprof.Sample{Value: tc.value})
			if got != tc.want {
				t.Fatalf("sampleCount(%v): got %d, want %d", tc.value, got, tc.want)
			}
		})
	}
}

func TestParser_expandsSampleCount(t *testing.T) {
	// One sample says "3 goroutines with this stack" → 3 Goroutine entries.
	data := writeProfile(t, sampleProfile(3, "main.leak", "main.go", 42))

	got, err := NewParser().Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(got.Goroutines) != 3 {
		t.Fatalf("goroutine count: got %d, want 3", len(got.Goroutines))
	}
	for _, g := range got.Goroutines {
		if len(g.Stack) != 1 {
			t.Fatalf("stack depth: got %d, want 1", len(g.Stack))
		}
		if g.Stack[0].Function != "main.leak" {
			t.Fatalf("function: got %q, want main.leak", g.Stack[0].Function)
		}
		if g.Stack[0].File != "main.go" {
			t.Fatalf("file: got %q, want main.go", g.Stack[0].File)
		}
		if g.Stack[0].Line != 42 {
			t.Fatalf("line: got %d, want 42", g.Stack[0].Line)
		}
	}
}

func TestParser_emptyProfile(t *testing.T) {
	// Valid profile with no samples = no leaks reported.
	prof := &pprof.Profile{
		SampleType: []*pprof.ValueType{{Type: "goroutineleak", Unit: "count"}},
	}
	got, err := NewParser().Parse(writeProfile(t, prof))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(got.Goroutines) != 0 {
		t.Fatalf("goroutines: got %d, want 0", len(got.Goroutines))
	}
}

func TestParser_skipsZeroAndNegativeSampleCounts(t *testing.T) {
	inlinee := &pprof.Function{ID: 1, Name: "ok"}
	loc := &pprof.Location{
		ID: 1,
		Line: []pprof.Line{{
			Function: inlinee,
			Line:     1,
		}},
	}
	prof := &pprof.Profile{
		SampleType: []*pprof.ValueType{{Type: "goroutineleak", Unit: "count"}},
		Function:   []*pprof.Function{inlinee},
		Location:   []*pprof.Location{loc},
		Sample: []*pprof.Sample{
			{Value: []int64{0}},
			{Value: []int64{-1}},
			{Value: nil},
			{Value: []int64{2}, Location: []*pprof.Location{loc}},
		},
	}

	got := goroutinesFromProfile(prof)
	if len(got) != 2 {
		t.Fatalf("goroutines: got %d, want 2", len(got))
	}
}

func TestParser_capturedAtFromProfileTime(t *testing.T) {
	want := time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)
	data := writeProfile(t, sampleProfile(1, "main.leak", "main.go", 1))

	got, err := NewParser().Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !got.CapturedAt.Equal(want) {
		t.Fatalf("CapturedAt: got %v, want %v", got.CapturedAt, want)
	}
}

func TestParser_malformedInput(t *testing.T) {
	_, err := NewParser().Parse([]byte("not a pprof profile"))
	if !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("error: got %v, want ErrInvalidProfile", err)
	}
}

func TestParser_emptyInput(t *testing.T) {
	_, err := NewParser().Parse(nil)
	if !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("error: got %v, want ErrInvalidProfile", err)
	}
}

func TestParser_inlineFrames(t *testing.T) {
	inlinee := &pprof.Function{ID: 1, Name: "inlinee", Filename: "a.go"}
	caller := &pprof.Function{ID: 2, Name: "caller", Filename: "b.go"}
	loc := &pprof.Location{
		ID: 1,
		Line: []pprof.Line{
			{Function: inlinee, Line: 10},
			{Function: caller, Line: 20},
		},
	}
	prof := &pprof.Profile{
		SampleType: []*pprof.ValueType{{Type: "goroutineleak", Unit: "count"}},
		Function:   []*pprof.Function{inlinee, caller},
		Location:   []*pprof.Location{loc},
		Sample: []*pprof.Sample{{
			Value:    []int64{1},
			Location: []*pprof.Location{loc},
		}},
	}

	got, err := NewParser().Parse(writeProfile(t, prof))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(got.Goroutines) != 1 {
		t.Fatalf("goroutines: got %d, want 1", len(got.Goroutines))
	}
	stack := got.Goroutines[0].Stack
	if len(stack) != 2 {
		t.Fatalf("stack depth: got %d, want 2", len(stack))
	}
	if stack[0].Function != "inlinee" || stack[1].Function != "caller" {
		t.Fatalf("unexpected stack: %+v", stack)
	}
}

// --- Fetch tests ---

func TestPProfSource_Fetch_success(t *testing.T) {
	fixture := writeProfile(t, sampleProfile(1, "main.leak", "main.go", 1))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	src := &PProfSource{URL: srv.URL, Client: srv.Client()}
	data, err := src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !bytes.Equal(data, fixture) {
		t.Fatal("fetched bytes do not match fixture")
	}
}

func TestPProfSource_Fetch_emptyURL(t *testing.T) {
	_, err := NewPProfSource("").Fetch(context.Background())
	if !errors.Is(err, ErrEmptyURL) {
		t.Fatalf("error: got %v, want ErrEmptyURL", err)
	}
}

func TestPProfSource_Fetch_nonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	src := &PProfSource{URL: srv.URL, Client: srv.Client()}
	_, err := src.Fetch(context.Background())
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestPProfSource_Fetch_cancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	src := &PProfSource{URL: srv.URL, Client: srv.Client()}
	_, err := src.Fetch(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
