//go:build integration

// Channel leak generator for golden fixture recording.
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	leak := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			<-leak
		}()
	}

	go func() {
		fmt.Println("channels leak generator: pprof on :6060")
		_ = http.ListenAndServe(":6060", nil)
	}()

	time.Sleep(time.Hour)
}
