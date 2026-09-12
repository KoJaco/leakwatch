//go:build integration

package envutil

import "os"

// Or returns the environment variable value or fallback when unset.
func Or(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
