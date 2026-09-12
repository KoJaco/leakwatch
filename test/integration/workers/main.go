//go:build integration

// Worker pool leak generator for golden fixture recording.
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/KoJaco/leakwatch/test/integration/envutil"
)

func workerLoop(ch <-chan struct{}) {
	<-ch
}

func main() {
	for i := 0; i < 40; i++ {
		ch := make(chan struct{})
		go workerLoop(ch)
	}

	pprofAddr := envutil.Or("LEAKWATCH_PPROF_ADDR", ":6060")
	go func() {
		fmt.Printf("workers leak generator: pprof on %s\n", pprofAddr)
		_ = http.ListenAndServe(pprofAddr, nil)
	}()

	time.Sleep(time.Hour)
}
