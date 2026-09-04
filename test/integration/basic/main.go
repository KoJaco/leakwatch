//go:build integration

// Intentionally leaks by blocking on a channel receive.
package main

import (
	"fmt"
	"time"
)

func main() {
	for i := 0; i < 10; i++ {
		ch := make(chan struct{})
		go func() {
			<-ch
		}()
	}
	fmt.Println("basic leak generator running")
	time.Sleep(time.Hour)
}
