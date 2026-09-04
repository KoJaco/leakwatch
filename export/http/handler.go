package httpexport

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/KoJaco/leakwatch/internal/analysis"
	"github.com/KoJaco/leakwatch/internal/fingerprint"
	"github.com/KoJaco/leakwatch/internal/profile"
)

// SnapshotProvider supplies the current leak snapshot for HTTP export.
type SnapshotProvider interface {
	Snapshot() analysis.Snapshot
}

// Handler serves leak debug endpoints.
type Handler struct {
	Provider SnapshotProvider
	Prefix   string
}

// Register mounts leak debug routes on mux at prefix.
//
//   - GET {prefix}         — list all leaks
//   - GET {prefix}/{id}    — leak detail
func Register(mux *http.ServeMux, prefix string, provider SnapshotProvider) {
	h := &Handler{Provider: provider, Prefix: prefix}
	mux.Handle(prefix, h)
	mux.Handle(prefix+"/", h)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snap := h.Provider.Snapshot()

	rel := strings.TrimPrefix(r.URL.Path, h.Prefix)
	rel = strings.TrimPrefix(rel, "/")

	if rel == "" {
		writeJSON(w, leakListResponse{
			CapturedAt: snap.CapturedAt,
			Leaks:      toSummary(snap.Leaks),
		})
		return
	}

	for _, leak := range snap.Leaks {
		if leak.ID == rel {
			writeJSON(w, leakDetailResponse{
				LeakID:    leak.ID,
				Count:     leak.Count,
				FirstSeen: leak.FirstSeen,
				LastSeen:  leak.LastSeen,
				Status:    leak.Status,
				Location:  toLocationJSON(leak.Site.Location),
				Stack:     toStackJSON(leak.Site.Stack),
			})
			return
		}
	}

	http.NotFound(w, r)
}

type leakListResponse struct {
	CapturedAt time.Time     `json:"captured_at"`
	Leaks      []leakSummary `json:"leaks"`
}

type leakSummary struct {
	LeakID string          `json:"leak_id"`
	Count  int             `json:"count"`
	Status analysis.Status `json:"status"`
}

type leakDetailResponse struct {
	LeakID    string          `json:"leak_id"`
	Count     int             `json:"count"`
	FirstSeen time.Time       `json:"first_seen"`
	LastSeen  time.Time       `json:"last_seen"`
	Status    analysis.Status `json:"status"`
	Location  locationJSON    `json:"location"`
	Stack     []frameJSON     `json:"stack"`
}

type locationJSON struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type frameJSON struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

func toLocationJSON(loc fingerprint.Location) locationJSON {
	return locationJSON{
		Function: loc.Function,
		File:     loc.File,
		Line:     loc.Line,
	}
}

func toStackJSON(frames []profile.Frame) []frameJSON {
	out := make([]frameJSON, len(frames))
	for i, f := range frames {
		out[i] = frameJSON{
			Function: f.Function,
			File:     f.File,
			Line:     f.Line,
		}
	}
	return out
}

func toSummary(leaks []analysis.Leak) []leakSummary {
	out := make([]leakSummary, 0, len(leaks))
	for _, l := range leaks {
		out = append(out, leakSummary{
			LeakID: l.ID,
			Count:  l.Count,
			Status: l.Status,
		})
	}
	return out
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
