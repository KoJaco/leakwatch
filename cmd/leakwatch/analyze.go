package main

import (
	"fmt"
	"os"

	"github.com/KoJaco/leakwatch/internal/pipeline"
)

func runAnalyze(args []string) int {
	positional, jsonOut, err := parseFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "usage: leakwatch analyze [-json] <profile.pb.gz>")
		return 1
	}

	prof, err := pipeline.ProfileFromFile(positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "leakwatch analyze: %v\n", err)
		return 1
	}

	snap, err := pipeline.SnapshotFromProfile(prof)
	if err != nil {
		fmt.Fprintf(os.Stderr, "leakwatch analyze: %v\n", err)
		return 1
	}

	if err := writeSnapshot(os.Stdout, snap, jsonOut); err != nil {
		fmt.Fprintf(os.Stderr, "leakwatch analyze: %v\n", err)
		return 1
	}
	return 0
}
