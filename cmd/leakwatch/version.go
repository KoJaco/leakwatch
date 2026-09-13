package main

import "fmt"

// version is overridden by GoReleaser ldflags on release builds.
var version = "dev"

func runVersion() {
	fmt.Println("leakwatch", version)
}
