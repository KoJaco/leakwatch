package profile

import "time"

const (
	// DefaultHTTPTimeout is the default timeout for pprof HTTP fetches.
	DefaultHTTPTimeout = 30 * time.Second
	// DefaultMaxProfileBytes is the default maximum profile payload size.
	DefaultMaxProfileBytes = 32 << 20 // 32 MiB
	// DefaultMaxSampleCount is the maximum goroutines per pprof sample.
	DefaultMaxSampleCount = 1_000_000
	// DefaultMaxTotalCount is the maximum total goroutines across all samples.
	DefaultMaxTotalCount = 10_000_000
)

// Profile is an immutable point-in-time snapshot of goroutine leak state.
// It answers "what did the runtime report at time T?". This is not whether something is growing.
type Profile struct {
	CapturedAt time.Time
	Samples    []StackSample
}

// StackSample is a stack trace with an associated goroutine count.
type StackSample struct {
	Stack []Frame
	Count int
}

// TotalCount returns the sum of goroutine counts across all samples.
func (p Profile) TotalCount() int {
	total := 0
	for _, s := range p.Samples {
		total += s.Count
	}
	return total
}

// Frame is a single stack frame.
type Frame struct {
	Function string
	File     string
	Line     int
}
