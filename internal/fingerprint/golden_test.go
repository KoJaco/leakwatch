package fingerprint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	pprof "github.com/google/pprof/profile"

	"github.com/KoJaco/leakwatch/internal/profile"
)

var goldenPatterns = []string{"channels", "nethttp", "grpc", "workers"}

type goldenCluster struct {
	ID       string          `json:"id"`
	Count    int             `json:"count"`
	Location goldenLocation  `json:"location"`
}

type goldenLocation struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type goldenClustersFile struct {
	Clusters []goldenCluster `json:"clusters"`
}

func TestGoldenFixtures(t *testing.T) {
	for _, pattern := range goldenPatterns {
		t.Run(pattern, func(t *testing.T) {
			prof := loadFixtureProfile(t, pattern)
			got, err := NewFingerprinter().Fingerprint(prof)
			if err != nil {
				t.Fatalf("Fingerprint: %v", err)
			}
			want := loadGoldenClusters(t, pattern)
			assertGoldenClusters(t, got, want)
		})
	}
}

func TestWriteGoldenFixtures(t *testing.T) {
	if os.Getenv("WRITE_GOLDEN") != "1" {
		t.Skip("set WRITE_GOLDEN=1 to regenerate golden files")
	}

	for _, pattern := range goldenPatterns {
		t.Run(pattern, func(t *testing.T) {
			prof := loadFixtureProfile(t, pattern)
			clusters, err := NewFingerprinter().Fingerprint(prof)
			if err != nil {
				t.Fatalf("Fingerprint: %v", err)
			}

			dir := filepath.Join("testdata", pattern)
			if err := writeGoldenClusters(dir, clusters); err != nil {
				t.Fatalf("write clusters.json: %v", err)
			}
		})
	}
}

func TestWriteSyntheticFixtures(t *testing.T) {
	if os.Getenv("WRITE_SYNTHETIC") != "1" {
		t.Skip("set WRITE_SYNTHETIC=1 to write synthetic leak.pb.gz fixtures")
	}

	writers := map[string]func(*testing.T) []byte{
		"channels": syntheticChannelsProfile,
		"nethttp":  syntheticNetHTTPProfile,
		"grpc":     syntheticGRPCProfile,
		"workers":  syntheticWorkersProfile,
	}

	for pattern, writer := range writers {
		t.Run(pattern, func(t *testing.T) {
			data := writer(t)
			path := filepath.Join("testdata", pattern, "leak.pb.gz")
			if err := os.WriteFile(path, data, 0o644); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
		})
	}
}

func loadFixtureProfile(t *testing.T, pattern string) profile.Profile {
	t.Helper()

	path := filepath.Join("testdata", pattern, "leak.pb.gz")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	prof, err := profile.NewParser().Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return prof
}

func loadGoldenClusters(t *testing.T, pattern string) goldenClustersFile {
	t.Helper()

	path := filepath.Join("testdata", pattern, "clusters.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read clusters.json: %v", err)
	}

	var want goldenClustersFile
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatalf("decode clusters.json: %v", err)
	}
	return want
}

func assertGoldenClusters(t *testing.T, got []LeakCluster, want goldenClustersFile) {
	t.Helper()

	if len(got) != len(want.Clusters) {
		t.Fatalf("cluster count: got %d, want %d", len(got), len(want.Clusters))
	}

	sort.Slice(got, func(i, j int) bool {
		return got[i].Fingerprint.ID < got[j].Fingerprint.ID
	})
	sort.Slice(want.Clusters, func(i, j int) bool {
		return want.Clusters[i].ID < want.Clusters[j].ID
	})

	for i := range want.Clusters {
		if got[i].Fingerprint.ID != want.Clusters[i].ID {
			t.Fatalf("cluster[%d] id: got %q, want %q", i, got[i].Fingerprint.ID, want.Clusters[i].ID)
		}
		if got[i].Count != want.Clusters[i].Count {
			t.Fatalf("cluster[%d] count: got %d, want %d", i, got[i].Count, want.Clusters[i].Count)
		}
		if got[i].Site.Location.Function != want.Clusters[i].Location.Function {
			t.Fatalf("cluster[%d] function: got %q, want %q", i, got[i].Site.Location.Function, want.Clusters[i].Location.Function)
		}
		if got[i].Site.Location.File != want.Clusters[i].Location.File {
			t.Fatalf("cluster[%d] file: got %q, want %q", i, got[i].Site.Location.File, want.Clusters[i].Location.File)
		}
		if got[i].Site.Location.Line != want.Clusters[i].Location.Line {
			t.Fatalf("cluster[%d] line: got %d, want %d", i, got[i].Site.Location.Line, want.Clusters[i].Location.Line)
		}
	}
}

func writeGoldenClusters(dir string, clusters []LeakCluster) error {
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Fingerprint.ID < clusters[j].Fingerprint.ID
	})

	out := goldenClustersFile{Clusters: make([]goldenCluster, len(clusters))}
	for i, c := range clusters {
		out.Clusters[i] = goldenCluster{
			ID:    c.Fingerprint.ID,
			Count: c.Count,
			Location: goldenLocation{
				Function: c.Site.Location.Function,
				File:     c.Site.Location.File,
				Line:     c.Site.Location.Line,
			},
		}
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(dir, "clusters.json"), data, 0o644)
}

func writeCountProfile(t *testing.T, count int64, frames []pprofFrame) []byte {
	t.Helper()

	prof := countProfile(count, frames)
	var buf bytes.Buffer
	if err := prof.Write(&buf); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	return buf.Bytes()
}

type pprofFrame struct {
	function string
	file     string
	line     int64
}

func countProfile(count int64, frames []pprofFrame) *pprof.Profile {
	functions := make([]*pprof.Function, len(frames))
	lines := make([]pprof.Line, len(frames))
	for i, f := range frames {
		id := uint64(i + 1)
		functions[i] = &pprof.Function{
			ID:       id,
			Name:     f.function,
			Filename: f.file,
		}
		lines[i] = pprof.Line{Function: functions[i], Line: f.line}
	}

	loc := &pprof.Location{ID: 1, Line: lines}
	return &pprof.Profile{
		TimeNanos:  time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC).UnixNano(),
		SampleType: []*pprof.ValueType{{Type: "goroutineleak", Unit: "count"}},
		Function:   functions,
		Location:   []*pprof.Location{loc},
		Sample: []*pprof.Sample{{
			Value:    []int64{count},
			Location: []*pprof.Location{loc},
		}},
	}
}

func syntheticChannelsProfile(t *testing.T) []byte {
	return writeCountProfile(t, 50, []pprofFrame{
		{function: "runtime.gopark", file: "runtime/proc.go", line: 100},
		{function: "runtime.chanrecv1", file: "runtime/chan.go", line: 50},
		{function: "main.main.func1", file: "/build/test/integration/channels/main.go", line: 15},
		{function: "main.main", file: "/build/test/integration/channels/main.go", line: 13},
	})
}

func syntheticNetHTTPProfile(t *testing.T) []byte {
	return writeCountProfile(t, 12, []pprofFrame{
		{function: "runtime.gopark", file: "runtime/proc.go", line: 100},
		{function: "runtime.chanrecv1", file: "runtime/chan.go", line: 50},
		{function: "main.main.func1.1", file: "/build/test/integration/nethttp/main.go", line: 16},
		{function: "main.main.func1", file: "/build/test/integration/nethttp/main.go", line: 15},
		{function: "net/http.HandlerFunc.ServeHTTP", file: "net/http/server.go", line: 100},
		{function: "main.main", file: "/build/test/integration/nethttp/main.go", line: 13},
	})
}

func syntheticWorkersProfile(t *testing.T) []byte {
	return writeCountProfile(t, 40, []pprofFrame{
		{function: "runtime.gopark", file: "runtime/proc.go", line: 100},
		{function: "runtime.chanrecv1", file: "runtime/chan.go", line: 50},
		{function: "main.workerLoop", file: "/build/test/integration/workers/main.go", line: 14},
		{function: "main.main", file: "/build/test/integration/workers/main.go", line: 19},
	})
}

func syntheticGRPCProfile(t *testing.T) []byte {
	return writeCountProfile(t, 8, []pprofFrame{
		{function: "runtime.gopark", file: "runtime/proc.go", line: 100},
		{function: "runtime.chanrecv1", file: "runtime/chan.go", line: 50},
		{function: "google.golang.org/grpc.(*Server).handleStream", file: "google.golang.org/grpc/server.go", line: 100},
		{function: "google.golang.org/grpc.(*Server).serveStreams", file: "google.golang.org/grpc/server.go", line: 200},
		{function: "main.startGRPCLeak", file: "/build/test/integration/grpc/main.go", line: 30},
		{function: "main.main", file: "/build/test/integration/grpc/main.go", line: 20},
	})
}
