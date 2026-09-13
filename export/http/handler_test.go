package httpexport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KoJaco/leakwatch/internal/analysis"
)

type stubProvider struct{}

func (stubProvider) Snapshot() analysis.Snapshot {
	return analysis.Snapshot{}
}

func TestRegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, "/debug/leaks", stubProvider{})

	req := httptest.NewRequest(http.MethodGet, "/debug/leaks", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRegisterRoutes_methodNotAllowed(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, "/debug/leaks", stubProvider{})

	req := httptest.NewRequest(http.MethodPost, "/debug/leaks", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestRegisterRoutes_middleware(t *testing.T) {
	mux := http.NewServeMux()
	called := false
	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			if r.Header.Get("X-Auth") != "ok" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	Register(mux, "/debug/leaks", stubProvider{}, WithMiddleware(auth))

	req := httptest.NewRequest(http.MethodGet, "/debug/leaks", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if !called {
		t.Fatal("middleware was not invoked")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	req = httptest.NewRequest(http.MethodGet, "/debug/leaks", nil)
	req.Header.Set("X-Auth", "ok")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
