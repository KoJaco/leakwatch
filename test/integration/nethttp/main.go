//go:build integration

// HTTP handler leak generator placeholder.
package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		done := make(chan struct{})
		go func() {
			<-done
		}()
		fmt.Fprintln(w, "ok")
	})
	fmt.Println("nethttp leak generator on :8080")
	_ = http.ListenAndServe(":8080", nil)
}
