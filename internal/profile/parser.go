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
	// ErrProfileTooLarge is returned when profile bytes or counts exceed limits.
	ErrProfileTooLarge = errors.New("profile: profile too large")
	// ErrWrongProfileType is returned when the profile is not goroutineleak.
	ErrWrongProfileType = errors.New("profile: expected goroutineleak sample type")
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
	URL             string
	Client          *http.Client // optional; defaults to client with DefaultHTTPTimeout
	MaxProfileBytes int64        // 0 = DefaultMaxProfileBytes
}

// SourceOption configures a PProfSource.
type SourceOption func(*PProfSource)

// WithSourceHTTPClient sets the HTTP client used for fetches.
func WithSourceHTTPClient(c *http.Client) SourceOption {
	return func(s *PProfSource) {
		s.Client = c
	}
}

// WithSourceMaxProfileBytes sets the maximum response body size.
func WithSourceMaxProfileBytes(n int64) SourceOption {
	return func(s *PProfSource) {
		s.MaxProfileBytes = n
	}
}

// NewPProfSource returns a Source that fetches from url.
func NewPProfSource(url string, opts ...SourceOption) *PProfSource {
	s := &PProfSource{URL: url}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *PProfSource) maxProfileBytes() int64 {
	if s.MaxProfileBytes > 0 {
		return s.MaxProfileBytes
	}
	return DefaultMaxProfileBytes
}

func (s *PProfSource) httpClient() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return &http.Client{Timeout: DefaultHTTPTimeout}
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("profile: pprof fetch %s: %s", s.URL, resp.Status)
	}

	max := s.maxProfileBytes()
	limited := io.LimitReader(resp.Body, max+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, fmt.Errorf("%w: response exceeds %d bytes", ErrProfileTooLarge, max)
	}
	if len(data) == 0 {
		return nil, ErrInvalidProfile
	}

	return data, nil
}

// DefaultParser parses goroutine leak profiles.
type DefaultParser struct {
	MaxSampleCount int // 0 = DefaultMaxSampleCount
	MaxTotalCount  int // 0 = DefaultMaxTotalCount
}

// NewParser returns the default profile parser.
func NewParser() *DefaultParser {
	return &DefaultParser{}
}

func (p *DefaultParser) maxSampleCount() int {
	if p.MaxSampleCount > 0 {
		return p.MaxSampleCount
	}
	return DefaultMaxSampleCount
}

func (p *DefaultParser) maxTotalCount() int {
	if p.MaxTotalCount > 0 {
		return p.MaxTotalCount
	}
	return DefaultMaxTotalCount
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

	if !isGoroutineLeakProfile(prof) {
		return Profile{}, ErrWrongProfileType
	}

	samples, err := samplesFromProfile(prof, p.maxSampleCount(), p.maxTotalCount())
	if err != nil {
		return Profile{}, err
	}

	return Profile{
		CapturedAt: capturedAt(prof),
		Samples:    samples,
	}, nil
}

func isGoroutineLeakProfile(prof *pprof.Profile) bool {
	for _, st := range prof.SampleType {
		if st != nil && st.Type == "goroutineleak" {
			return true
		}
	}
	return false
}

func capturedAt(prof *pprof.Profile) time.Time {
	if prof.TimeNanos != 0 {
		return time.Unix(0, prof.TimeNanos).UTC()
	}
	return time.Now().UTC()
}

func samplesFromProfile(prof *pprof.Profile, maxSampleCount, maxTotalCount int) ([]StackSample, error) {
	if len(prof.Sample) == 0 {
		return nil, nil
	}

	var samples []StackSample
	total := 0
	for _, sample := range prof.Sample {
		count := sampleCount(sample)
		if count == 0 {
			continue
		}
		if count > maxSampleCount {
			return nil, fmt.Errorf("%w: sample count %d exceeds limit %d", ErrProfileTooLarge, count, maxSampleCount)
		}

		total += count
		if total > maxTotalCount {
			return nil, fmt.Errorf("%w: total count %d exceeds limit %d", ErrProfileTooLarge, total, maxTotalCount)
		}

		stack := stackFromSample(sample)
		samples = append(samples, StackSample{
			Stack: stack,
			Count: count,
		})
	}

	return samples, nil
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
