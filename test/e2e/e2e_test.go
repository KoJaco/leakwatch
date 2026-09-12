//go:build integration

package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/KoJaco/leakwatch/internal/analysis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestBasicLeak(t *testing.T) {
	gen := startGenerator(t, "basic", nil)
	waitForProfileGoroutines(t, gen.pprofURL, profileReadyTimeout)

	w := newTestWatcher(t, gen.pprofURL)
	snap := waitForLeaks(t, w, 1, leakAppearTimeout)

	if snap.Leaks[0].Count < 10 {
		t.Fatalf("count = %d, want >= 10", snap.Leaks[0].Count)
	}
	if snap.Leaks[0].Status != analysis.StatusNew {
		t.Fatalf("status = %q, want %q", snap.Leaks[0].Status, analysis.StatusNew)
	}
}

func TestChannelsStableLeakID(t *testing.T) {
	gen := startGenerator(t, "channels", nil)
	waitForProfileGoroutines(t, gen.pprofURL, profileReadyTimeout)

	w := newTestWatcher(t, gen.pprofURL)
	first := waitForLeaks(t, w, 1, leakAppearTimeout)
	firstID := first.Leaks[0].ID

	waitForObservationCount(t, w, 2, observationTimeout)
	second := w.Snapshot()
	if len(second.Leaks) == 0 {
		t.Fatal("expected leaks in second snapshot")
	}
	if second.Leaks[0].ID != firstID {
		t.Fatalf("leak_id changed: first %q, second %q", firstID, second.Leaks[0].ID)
	}
	if second.Leaks[0].Status != analysis.StatusPersistent {
		t.Fatalf("status = %q, want %q", second.Leaks[0].Status, analysis.StatusPersistent)
	}
}

func TestNetHTTPLLeakAfterRequests(t *testing.T) {
	gen := startGenerator(t, "nethttp", nil)
	triggerHTTP(t, gen.httpURL, 3)
	waitForProfileGoroutines(t, gen.pprofURL, profileReadyTimeout)

	w := newTestWatcher(t, gen.pprofURL)
	first := waitForLeaks(t, w, 1, leakAppearTimeout)
	firstCount := first.Leaks[0].Count

	triggerHTTP(t, gen.httpURL, 5)
	waitForObservationCount(t, w, 2, observationTimeout)

	second := w.Snapshot()
	if len(second.Leaks) == 0 {
		t.Fatal("expected leaks after additional requests")
	}
	if second.Leaks[0].Count <= firstCount {
		t.Fatalf("count did not increase: first=%d second=%d", firstCount, second.Leaks[0].Count)
	}

	growth := analysis.GrowthFor(w.Observations(), second.Leaks[0].ID, analysis.DefaultGrowthWindow)
	if growth.RatePerMinute <= 0 {
		t.Fatalf("growth rate = %v, want > 0", growth.RatePerMinute)
	}
}

func TestGRPCLeak(t *testing.T) {
	gen := startGenerator(t, "grpc", nil)
	triggerGRPCStream(t, gen.grpcAddr, 3)
	waitForProfileGoroutines(t, gen.pprofURL, profileReadyTimeout)

	w := newTestWatcher(t, gen.pprofURL)
	snap := waitForLeaks(t, w, 1, leakAppearTimeout)

	foundGRPCFrame := false
	for _, frame := range snap.Leaks[0].Site.Stack {
		if strings.Contains(frame.Function, "grpc") {
			foundGRPCFrame = true
			break
		}
	}
	if !foundGRPCFrame {
		t.Fatalf("expected grpc frame in stack, got %v", snap.Leaks[0].Site.Stack)
	}
}

func triggerGRPCStream(t *testing.T, addr string, n int) {
	t.Helper()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	defer conn.Close()

	done := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			var out emptypb.Empty
			_ = conn.Invoke(ctx, "/leak.LeakService/Leak", &emptypb.Empty{}, &out)
		}()
	}

	deadline := time.Now().Add(2 * time.Second)
	for i := 0; i < n; i++ {
		select {
		case <-done:
		case <-time.After(time.Until(deadline)):
			t.Fatalf("timed out waiting for grpc leak RPCs to start")
		}
	}
	time.Sleep(200 * time.Millisecond)
}
