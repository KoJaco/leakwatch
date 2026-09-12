//go:build integration

// HTTP handler leak generator for golden fixture recording.
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/KoJaco/leakwatch/test/integration/envutil"
)

func main() {
	httpAddr := envutil.Or("LEAKWATCH_HTTP_ADDR", ":8080")
	pprofAddr := envutil.Or("LEAKWATCH_PPROF_ADDR", ":6060")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		done := make(chan struct{})
		go func() {
			<-done
		}()
		fmt.Fprintln(w, "ok")
	})

	go func() {
		fmt.Printf("nethttp leak generator: pprof on %s\n", pprofAddr)
		_ = http.ListenAndServe(pprofAddr, nil)
	}()

	fmt.Printf("nethttp leak generator listening on %s\n", httpAddr)
	_ = http.ListenAndServe(httpAddr, nil)
}
