//go:build integration

package e2e

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/KoJaco/leakwatch"
	"github.com/KoJaco/leakwatch/internal/analysis"
	"github.com/KoJaco/leakwatch/internal/profile"
)

const (
	profileReadyTimeout = 60 * time.Second
	leakAppearTimeout   = 30 * time.Second
	observationTimeout  = 15 * time.Second
	sampleInterval      = 500 * time.Millisecond
)

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func freePort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	return port
}

func pprofURL(hostPort int) string {
	return fmt.Sprintf("http://127.0.0.1:%d/debug/pprof/goroutineleak", hostPort)
}

func httpURL(hostPort int) string {
	return fmt.Sprintf("http://127.0.0.1:%d/", hostPort)
}

func grpcAddr(hostPort int) string {
	return fmt.Sprintf("127.0.0.1:%d", hostPort)
}

type generator struct {
	cmd      *exec.Cmd
	pprofURL string
	httpURL  string
	grpcAddr string
}

func startGenerator(t *testing.T, pkg string, extraEnv map[string]string) generator {
	t.Helper()

	pprofPort := freePort(t)
	env := map[string]string{
		"LEAKWATCH_PPROF_ADDR": fmt.Sprintf("127.0.0.1:%d", pprofPort),
	}
	gen := generator{pprofURL: pprofURL(pprofPort)}

	if pkg == "nethttp" {
		httpPort := freePort(t)
		env["LEAKWATCH_HTTP_ADDR"] = fmt.Sprintf("127.0.0.1:%d", httpPort)
		gen.httpURL = httpURL(httpPort)
	}
	if pkg == "grpc" {
		grpcPort := freePort(t)
		env["LEAKWATCH_GRPC_ADDR"] = fmt.Sprintf("127.0.0.1:%d", grpcPort)
		gen.grpcAddr = grpcAddr(grpcPort)
	}

	for k, v := range extraEnv {
		env[k] = v
	}

	root := repoRoot(t)
	cmd := exec.Command("go", "run", "-tags", "integration", filepath.Join(root, "test", "integration", pkg))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), flattenEnv(env)...)

	if err := cmd.Start(); err != nil {
		t.Fatalf("start generator %s: %v", pkg, err)
	}
	gen.cmd = cmd

	t.Cleanup(func() {
		if gen.cmd.Process != nil {
			_ = gen.cmd.Process.Kill()
		}
		_, _ = gen.cmd.Process.Wait()
	})

	waitForPProfReady(t, gen.pprofURL, 10*time.Second)
	return gen
}

func flattenEnv(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}

func waitForPProfReady(t *testing.T, url string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("pprof endpoint not ready: %s", url)
}

func waitForProfileGoroutines(t *testing.T, url string, timeout time.Duration) {
	t.Helper()

	parser := profile.NewParser()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		data, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil || resp.StatusCode != http.StatusOK {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		prof, err := parser.Parse(data)
		if err == nil && len(prof.Goroutines) > 0 {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("profile at %s did not contain goroutines within %s", url, timeout)
}

func newTestWatcher(t *testing.T, pprofURL string) *leakwatch.Watcher {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	w := leakwatch.New(
		leakwatch.WithPProfURL(pprofURL),
		leakwatch.WithInterval(sampleInterval),
		leakwatch.WithJitter(0),
	)
	w.Start(ctx)
	t.Cleanup(func() {
		w.Stop()
		cancel()
	})
	return w
}

func waitForLeaks(t *testing.T, w *leakwatch.Watcher, min int, timeout time.Duration) analysis.Snapshot {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		snap := w.Snapshot()
		if len(snap.Leaks) >= min {
			return snap
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("expected at least %d leak(s) within %s, got %d", min, timeout, len(w.Snapshot().Leaks))
	return analysis.Snapshot{}
}

func waitForObservationCount(t *testing.T, w *leakwatch.Watcher, n int, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if len(w.Observations()) >= n {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("expected at least %d observation(s) within %s, got %d", n, timeout, len(w.Observations()))
}

func triggerHTTP(t *testing.T, url string, n int) {
	t.Helper()

	for i := 0; i < n; i++ {
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("GET %s: %v", url, err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}
