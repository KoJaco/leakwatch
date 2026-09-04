package fingerprint

import "github.com/KoJaco/leakwatch/internal/profile"

// ProjectStack reduces a stack to the application-relevant frames.
func ProjectStack(stack []profile.Frame) []profile.Frame {
	_ = stack
	return nil
}
