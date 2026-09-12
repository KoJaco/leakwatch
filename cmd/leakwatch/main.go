package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "analyze":
		os.Exit(runAnalyze(os.Args[2:]))
	case "inspect":
		os.Exit(runInspect(os.Args[2:]))
	case "version":
		runVersion()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: leakwatch <command> [arguments]

Commands:
  analyze [-json] <profile.pb.gz>   Analyze an exported goroutine leak profile
  inspect [-json] <pprof-url>       Fetch and inspect a live pprof endpoint
  version                           Print version information
`)
}
