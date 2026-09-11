package fingerprint

import (
	"strings"

	"github.com/KoJaco/leakwatch/internal/profile"
)

// ProjectStack reduces a stack to the application-relevant frames.
func ProjectStack(stack []profile.Frame) []profile.Frame {
	if len(stack) == 0 {
		return nil
	}

	out := make([]profile.Frame, 0, len(stack))
	for _, f := range stack {
		if isRuntimeFrame(f.Function) {
			continue
		}
		out = append(out, f)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isRuntimeFrame(fn string) bool {
	return fn == "runtime.goexit" ||
		strings.HasPrefix(fn, "runtime.") ||
		strings.HasPrefix(fn, "runtime/") ||
		strings.HasPrefix(fn, "internal/runtime/") ||
		strings.HasPrefix(fn, "internal/poll.") ||
		strings.HasPrefix(fn, "internal/sync.")
}

func leakLocation(projected []profile.Frame) Location {
	f := projected[0]
	return Location{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}
