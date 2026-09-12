//go:build integration

// HTTP handler leak generator for golden fixture recording.
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		done := make(chan struct{})
		go func() {
			<-done
		}()
		fmt.Fprintln(w, "ok")
	})

	go func() {
		fmt.Println("nethttp leak generator: app on :8080, pprof on :6060")
		_ = http.ListenAndServe(":6060", nil)
	}()

	fmt.Println("nethttp leak generator listening on :8080")
	_ = http.ListenAndServe(":8080", nil)
}
