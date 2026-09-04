package jsonexport

import (
	"encoding/json"
	"io"

	"github.com/KoJaco/leakwatch/internal/analysis"
)

// Encode writes snap as JSON to w.
func Encode(w io.Writer, snap analysis.Snapshot) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(snap)
}
