package main

import "fmt"

var version = "dev"

func runVersion() {
	fmt.Println("leakwatch", version)
}
