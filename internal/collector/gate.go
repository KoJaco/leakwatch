package collector

import "context"

// Gate provides named sample gate helpers beyond AlwaysAllow and NeverAllow.

// GateIf returns a gate that permits sampling only when cond is true.
func GateIf(cond func(context.Context) bool) SampleGate {
	return func(ctx context.Context) (bool, error) {
		return cond(ctx), nil
	}
}
