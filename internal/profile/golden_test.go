package profile

import (
	"os"
	"path/filepath"
	"testing"
)

var goldenPatterns = []string{"channels", "nethttp", "grpc", "workers"}

func TestParseGoldenFixtures(t *testing.T) {
	for _, pattern := range goldenPatterns {
		t.Run(pattern, func(t *testing.T) {
			data, err := os.ReadFile(goldenFixturePath(pattern))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			prof, err := NewParser().Parse(data)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if prof.TotalCount() == 0 {
				t.Fatal("expected goroutines in golden fixture")
			}
			if prof.CapturedAt.IsZero() {
				t.Fatal("expected CapturedAt in golden fixture")
			}
		})
	}
}

func goldenFixturePath(pattern string) string {
	return filepath.Join("..", "fingerprint", "testdata", pattern, "leak.pb.gz")
}
