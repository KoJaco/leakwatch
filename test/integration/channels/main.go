//go:build integration

// Channel leak generator placeholder.
package main

import (
	"fmt"
	"time"
)

func main() {
	leak := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			<-leak
		}()
	}
	fmt.Println("channels leak generator running")
	time.Sleep(time.Hour)
}
