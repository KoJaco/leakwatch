package main

import (
	"fmt"
	"io"

	jsonexport "github.com/KoJaco/leakwatch/export/json"
	"github.com/KoJaco/leakwatch/internal/analysis"
	"github.com/KoJaco/leakwatch/internal/profile"
)

func parseFlags(args []string) (positional []string, jsonOut bool, err error) {
	for _, arg := range args {
		switch arg {
		case "-json":
			jsonOut = true
		default:
			positional = append(positional, arg)
		}
	}
	return positional, jsonOut, nil
}

func writeSnapshot(w io.Writer, snap analysis.Snapshot, jsonOut bool) error {
	if jsonOut {
		return jsonexport.Encode(w, snap)
	}
	return printSnapshot(w, snap)
}

func printSnapshot(w io.Writer, snap analysis.Snapshot) error {
	if _, err := fmt.Fprintf(w, "Captured: %s\n", snap.CapturedAt.UTC().Format(timeRFC3339)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Leaks: %d\n", len(snap.Leaks)); err != nil {
		return err
	}
	for i, leak := range snap.Leaks {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := printLeak(w, leak); err != nil {
			return err
		}
	}
	return nil
}

func printLeak(w io.Writer, leak analysis.Leak) error {
	if _, err := fmt.Fprintf(w, "ID: %s  Count: %d  Status: %s\n", leak.ID, leak.Count, leak.Status); err != nil {
		return err
	}
	loc := leak.Site.Location
	if _, err := fmt.Fprintf(w, "  Location: %s (%s:%d)\n", loc.Function, loc.File, loc.Line); err != nil {
		return err
	}
	if len(leak.Site.Stack) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "  Stack:"); err != nil {
		return err
	}
	for _, frame := range leak.Site.Stack {
		if err := printFrame(w, frame); err != nil {
			return err
		}
	}
	return nil
}

func printFrame(w io.Writer, frame profile.Frame) error {
	_, err := fmt.Fprintf(w, "    %s (%s:%d)\n", frame.Function, frame.File, frame.Line)
	return err
}

const timeRFC3339 = "2006-01-02T15:04:05Z"
