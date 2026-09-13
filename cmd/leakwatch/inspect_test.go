package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunInspect_success(t *testing.T) {
	data := fixtureProfileBytes(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	out := captureStdout(t, func() {
		if got := runInspect([]string{srv.URL}); got != 0 {
			t.Fatalf("runInspect() = %d, want 0", got)
		}
	})
	if !bytes.Contains(out, []byte("Status: new")) {
		t.Fatalf("stdout = %q, want status new", out)
	}
}

func TestRunInspect_rejectsRemoteWithoutFlag(t *testing.T) {
	if got := runInspect([]string{"http://example.com/debug/pprof/goroutineleak"}); got != 1 {
		t.Fatalf("runInspect() = %d, want 1", got)
	}
}

func TestRunInspect_allowsRemoteWithFlag(t *testing.T) {
	data := fixtureProfileBytes(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	out := captureStdout(t, func() {
		if got := runInspect([]string{"--allow-remote", srv.URL}); got != 0 {
			t.Fatalf("runInspect() = %d, want 0", got)
		}
	})
	if !bytes.Contains(out, []byte("Status: new")) {
		t.Fatalf("stdout = %q, want status new", out)
	}
}

func TestRunInspect_usage(t *testing.T) {
	if got := runInspect(nil); got != 1 {
		t.Fatalf("runInspect() = %d, want 1", got)
	}
}
