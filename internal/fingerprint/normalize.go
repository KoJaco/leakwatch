package fingerprint

import "github.com/KoJaco/leakwatch/internal/profile"

// NormalizeStack applies canonical normalization to a goroutine stack.
func NormalizeStack(stack []profile.Frame) []profile.Frame {
	_ = stack
	return nil
}
