package main

import (
	"fmt"
	"os"
)

func runInspect(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: leakwatch inspect <pprof-url>")
		return 1
	}

	url := args[0]
	fmt.Fprintf(os.Stderr, "leakwatch inspect: not implemented (url: %s)\n", url)
	return 1
}
