package fingerprint

import (
	"testing"

	"github.com/KoJaco/leakwatch/internal/profile"
)

func channelLeakStack() []profile.Frame {
	return []profile.Frame{
		{Function: "runtime.gopark", File: "runtime/proc.go", Line: 100},
		{Function: "runtime.chanrecv1", File: "runtime/chan.go", Line: 50},
		{Function: "main.leak", File: "/build/main.go", Line: 42},
		{Function: "main.main", File: "/build/main.go", Line: 20},
	}
}

func TestNormalizeStack_usesBasename(t *testing.T) {
	got := NormalizeStack(channelLeakStack())
	if len(got) != 4 {
		t.Fatalf("frames: got %d, want 4", len(got))
	}
	if got[2].File != "main.go" {
		t.Fatalf("file: got %q, want main.go", got[2].File)
	}
}

func TestNormalizeStack_collapsesConsecutiveDuplicates(t *testing.T) {
	stack := []profile.Frame{
		{Function: "main.leak", File: "/a/main.go", Line: 1},
		{Function: "main.leak", File: "/a/main.go", Line: 1},
		{Function: "main.main", File: "/a/main.go", Line: 10},
	}
	got := NormalizeStack(stack)
	if len(got) != 2 {
		t.Fatalf("frames: got %d, want 2", len(got))
	}
}

func TestNormalizeStack_empty(t *testing.T) {
	if got := NormalizeStack(nil); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
	if got := NormalizeStack([]profile.Frame{{Function: ""}}); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestProjectStack_stripsRuntime(t *testing.T) {
	got := ProjectStack(NormalizeStack(channelLeakStack()))
	if len(got) != 2 {
		t.Fatalf("frames: got %d, want 2", len(got))
	}
	if got[0].Function != "main.leak" {
		t.Fatalf("first frame: got %q, want main.leak", got[0].Function)
	}
	if got[1].Function != "main.main" {
		t.Fatalf("second frame: got %q, want main.main", got[1].Function)
	}
}

func TestProjectStack_keepsStdlib(t *testing.T) {
	stack := []profile.Frame{
		{Function: "runtime.gopark", File: "runtime/proc.go", Line: 1},
		{Function: "net/http.HandlerFunc.ServeHTTP", File: "net/http/server.go", Line: 100},
		{Function: "main.main", File: "main.go", Line: 10},
	}
	got := ProjectStack(NormalizeStack(stack))
	if len(got) != 2 {
		t.Fatalf("frames: got %d, want 2", len(got))
	}
	if got[0].Function != "net/http.HandlerFunc.ServeHTTP" {
		t.Fatalf("first frame: got %q", got[0].Function)
	}
}

func TestProjectStack_runtimeOnly(t *testing.T) {
	stack := []profile.Frame{
		{Function: "runtime.gopark", File: "runtime/proc.go", Line: 1},
		{Function: "runtime.chanrecv1", File: "runtime/chan.go", Line: 2},
	}
	if got := ProjectStack(NormalizeStack(stack)); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestStackKey_deterministic(t *testing.T) {
	stack := ProjectStack(NormalizeStack(channelLeakStack()))
	k1 := stackKey(stack)
	k2 := stackKey(stack)
	if k1 != k2 {
		t.Fatalf("keys differ: %q vs %q", k1, k2)
	}

	stack[0].Line = 99
	k3 := stackKey(stack)
	if k1 == k3 {
		t.Fatal("expected different key after line change")
	}
}

func TestFingerprintID_stable(t *testing.T) {
	key := stackKey(ProjectStack(NormalizeStack(channelLeakStack())))
	id1 := fingerprintID(key)
	id2 := fingerprintID(key)
	if id1 != id2 {
		t.Fatalf("ids differ: %q vs %q", id1, id2)
	}
	if len(id1) != 16 {
		t.Fatalf("id length: got %d, want 16", len(id1))
	}
}

func TestClusterSamples_groupsByStack(t *testing.T) {
	stack := channelLeakStack()
	got := ClusterSamples([]profile.StackSample{
		{Stack: stack, Count: 1},
		{Stack: stack, Count: 1},
		{Stack: append([]profile.Frame(nil), stack...), Count: 1},
	})
	if len(got) != 1 {
		t.Fatalf("clusters: got %d, want 1", len(got))
	}
	if got[0].Count != 3 {
		t.Fatalf("count: got %d, want 3", got[0].Count)
	}
	if got[0].Site.Location.Function != "main.leak" {
		t.Fatalf("location: got %q, want main.leak", got[0].Site.Location.Function)
	}
	if got[0].Fingerprint.ID == "" || got[0].Fingerprint.Key == "" {
		t.Fatal("expected non-empty fingerprint")
	}
}

func TestClusterSamples_separateStacks(t *testing.T) {
	stackA := channelLeakStack()
	stackB := append([]profile.Frame(nil), channelLeakStack()...)
	stackB[2].Function = "main.otherLeak"

	got := ClusterSamples([]profile.StackSample{
		{Stack: stackA, Count: 1},
		{Stack: stackA, Count: 1},
		{Stack: stackB, Count: 1},
	})
	if len(got) != 2 {
		t.Fatalf("clusters: got %d, want 2", len(got))
	}
	if got[0].Count != 2 {
		t.Fatalf("top count: got %d, want 2", got[0].Count)
	}
	if got[1].Count != 1 {
		t.Fatalf("second count: got %d, want 1", got[1].Count)
	}
}

func TestClusterSamples_sortOrder(t *testing.T) {
	stackA := channelLeakStack()
	stackB := append([]profile.Frame(nil), channelLeakStack()...)
	stackB[2].Function = "main.otherLeak"

	got := ClusterSamples([]profile.StackSample{
		{Stack: stackB, Count: 1},
		{Stack: stackA, Count: 1},
		{Stack: stackA, Count: 1},
		{Stack: stackA, Count: 1},
	})
	if len(got) != 2 {
		t.Fatalf("clusters: got %d, want 2", len(got))
	}
	if got[0].Count != 3 {
		t.Fatalf("first cluster count: got %d, want 3", got[0].Count)
	}
	if got[1].Count != 1 {
		t.Fatalf("second cluster count: got %d, want 1", got[1].Count)
	}
}

func TestClusterSamples_empty(t *testing.T) {
	if got := ClusterSamples(nil); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestFingerprint_endToEnd(t *testing.T) {
	stack := channelLeakStack()
	prof := profile.Profile{
		Samples: []profile.StackSample{
			{Stack: stack, Count: 2},
		},
	}

	clusters, err := NewFingerprinter().Fingerprint(prof)
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}
	if len(clusters) != 1 {
		t.Fatalf("clusters: got %d, want 1", len(clusters))
	}
	if clusters[0].Count != 2 {
		t.Fatalf("count: got %d, want 2", clusters[0].Count)
	}
	if clusters[0].Site.Location.Function != "main.leak" {
		t.Fatalf("location: got %q, want main.leak", clusters[0].Site.Location.Function)
	}
	if len(clusters[0].Site.Stack) != 2 {
		t.Fatalf("stack depth: got %d, want 2", len(clusters[0].Site.Stack))
	}
}

func TestFingerprint_emptyProfile(t *testing.T) {
	clusters, err := NewFingerprinter().Fingerprint(profile.Profile{})
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}
	if clusters != nil {
		t.Fatalf("clusters: got %v, want nil", clusters)
	}
}
