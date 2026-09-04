package main

import (
	"fmt"
	"os"
)

func runAnalyze(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: leakwatch analyze <profile.pb.gz>")
		return 1
	}

	path := args[0]
	fmt.Fprintf(os.Stderr, "leakwatch analyze: not implemented (profile: %s)\n", path)
	return 1
}
