//go:build integration

// Worker pool leak generator for golden fixture recording.
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func workerLoop(ch <-chan struct{}) {
	<-ch
}

func main() {
	for i := 0; i < 40; i++ {
		ch := make(chan struct{})
		go workerLoop(ch)
	}

	go func() {
		fmt.Println("workers leak generator: pprof on :6060")
		_ = http.ListenAndServe(":6060", nil)
	}()

	time.Sleep(time.Hour)
}
