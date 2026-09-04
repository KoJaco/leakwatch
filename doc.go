// Package leakwatch detects goroutine leaks using the Go 1.27+ runtime leak profile.
//
// The public API is intentionally small. Create a Watcher, start it, and optionally
// expose debug endpoints for human-readable leak attribution.
//
//	 watcher := leakwatch.New(leakwatch.WithInterval(5 * time.Minute))
//	 watcher.Start(context.Background())
package leakwatch
