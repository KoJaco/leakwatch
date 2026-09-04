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
