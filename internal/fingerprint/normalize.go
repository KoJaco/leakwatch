package fingerprint

import (
	"path/filepath"

	"github.com/KoJaco/leakwatch/internal/profile"
)

// NormalizeStack applies canonical normalization to a goroutine stack.
func NormalizeStack(stack []profile.Frame) []profile.Frame {
	if len(stack) == 0 {
		return nil
	}

	out := make([]profile.Frame, 0, len(stack))
	for _, f := range stack {
		if f.Function == "" {
			continue
		}
		nf := profile.Frame{
			Function: f.Function,
			File:     filepath.Base(f.File),
			Line:     f.Line,
		}
		if len(out) > 0 && framesEqual(out[len(out)-1], nf) {
			continue
		}
		out = append(out, nf)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func framesEqual(a, b profile.Frame) bool {
	return a.Function == b.Function && a.File == b.File && a.Line == b.Line
}
