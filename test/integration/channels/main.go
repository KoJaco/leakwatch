//go:build integration

// Channel leak generator for golden fixture recording.
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/KoJaco/leakwatch/test/integration/envutil"
)

func main() {
	leak := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			<-leak
		}()
	}

	pprofAddr := envutil.Or("LEAKWATCH_PPROF_ADDR", ":6060")
	go func() {
		fmt.Printf("channels leak generator: pprof on %s\n", pprofAddr)
		_ = http.ListenAndServe(pprofAddr, nil)
	}()

	time.Sleep(time.Hour)
}
