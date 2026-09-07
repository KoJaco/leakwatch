package profile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	pprof "github.com/google/pprof/profile"
)

var (
	// ErrEmptyURL is returned when a PProfSource has no URL configured.
	ErrEmptyURL = errors.New("profile: empty pprof URL")
	// ErrInvalidProfile is returned when profile bytes cannot be parsed.
	ErrInvalidProfile = errors.New("profile: invalid profile")
)


// Source fetches raw profile bytes from an arbitrary origin.
type Source interface {
	Fetch(ctx context.Context) ([]byte, error)
}

// Parser converts raw profile bytes into a domain Profile.
type Parser interface {
	Parse(data []byte) (Profile, error)
}

// PProfSource fetches a goroutine leak profile from a pprof HTTP endpoint.
type PProfSource struct {
	URL    string
	Client *http.Client // optional; defaults to http.DefaultClient
}

// NewPProfSource returns a Source that fetches from url.
func NewPProfSource(url string) *PProfSource {
	return &PProfSource{URL: url}
}

func (s *PProfSource) httpClient() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return http.DefaultClient
}

// Fetch retrieves raw profile bytes from the configured endpoint.
func (s *PProfSource) Fetch(ctx context.Context) ([]byte, error) {
	if s.URL == "" {
		return nil, ErrEmptyURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient().Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("profile: pprof fetch %s: %s", s.URL, resp.Status)
	}

	data, readErr := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); readErr == nil && closeErr != nil {
		return nil, closeErr
	}
	if readErr != nil {
		return nil, readErr
	}

	if len(data) == 0 {
		return nil, ErrInvalidProfile
	}

	return data, nil
}

// DefaultParser parses goroutine leak profiles.
type DefaultParser struct{}

// NewParser returns the default profile parser.
func NewParser() *DefaultParser {
	return &DefaultParser{}
}

// Parse converts raw bytes into a Profile.
func (p *DefaultParser) Parse(data []byte) (Profile, error) {
	if len(data) == 0 {
		return Profile{}, ErrInvalidProfile
	}

	prof, err := pprof.Parse(bytes.NewReader(data))
	if err != nil {
		return Profile{}, fmt.Errorf("%w: %v", ErrInvalidProfile, err)
	}

	return Profile{
		CapturedAt: capturedAt(prof),
		Goroutines: goroutinesFromProfile(prof),
	}, nil
}

func capturedAt(prof *pprof.Profile) time.Time {
	if prof.TimeNanos != 0 {
		return time.Unix(0, prof.TimeNanos).UTC()
	}
	return time.Now().UTC()
}


func goroutinesFromProfile(prof *pprof.Profile) []Goroutine {
	if len(prof.Sample) == 0 {
		return nil
	}

	var goroutines []Goroutine
	for _, sample := range prof.Sample {
		count := sampleCount(sample)
		if count == 0 {
			continue
		}

		stack := stackFromSample(sample)

		for i := 0; i < count; i++ {
			goroutines = append(goroutines, Goroutine{Stack: stack})
		}
	}

	return goroutines
}

func sampleCount(sample *pprof.Sample) int {
	if len(sample.Value) == 0 {
		return 0
	}

	if sample.Value[0] <= 0 {
		return 0
	}

	return int(sample.Value[0])
}


func stackFromSample(sample *pprof.Sample) []Frame {
	var stack []Frame
	for _, loc := range sample.Location {
		if loc == nil {
			continue
		}

		for _, line := range loc.Line {
			if line.Function == nil {
				continue
			}

			stack = append(stack, Frame{
				Function: line.Function.Name,
				File:     line.Function.Filename,
				Line:     int(line.Line),
			})
		}
	}

	return stack
}
