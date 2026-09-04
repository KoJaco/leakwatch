//go:build ignore

package main

import (
	"context"
	"time"

	"github.com/KoJaco/leakwatch"
)

func main() {
	watcher := leakwatch.New(
		leakwatch.WithInterval(5*time.Minute),
		leakwatch.WithJitter(30*time.Second),
	)

	watcher.Start(context.Background())
	_ = watcher.ServeDebug("/debug/leaks")

	select {}
}
