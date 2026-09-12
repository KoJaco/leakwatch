//go:build integration

// Intentionally leaks by blocking on a channel receive.
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/KoJaco/leakwatch/test/integration/envutil"
)

func main() {
	for i := 0; i < 10; i++ {
		ch := make(chan struct{})
		go func() {
			<-ch
		}()
	}

	pprofAddr := envutil.Or("LEAKWATCH_PPROF_ADDR", ":6060")
	go func() {
		fmt.Printf("basic leak generator: pprof on %s\n", pprofAddr)
		_ = http.ListenAndServe(pprofAddr, nil)
	}()

	fmt.Println("basic leak generator running")
	time.Sleep(time.Hour)
}
