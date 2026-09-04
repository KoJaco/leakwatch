package profile

import (
	"context"
	"errors"
)

// ErrNotImplemented is returned by stub implementations.
var ErrNotImplemented = errors.New("profile: not implemented")

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
	URL string
}

// NewPProfSource returns a Source that fetches from url.
func NewPProfSource(url string) *PProfSource {
	return &PProfSource{URL: url}
}

// Fetch retrieves raw profile bytes from the configured endpoint.
func (s *PProfSource) Fetch(ctx context.Context) ([]byte, error) {
	_ = ctx
	_ = s.URL
	return nil, ErrNotImplemented
}

// DefaultParser parses goroutine leak profiles.
type DefaultParser struct{}

// NewParser returns the default profile parser.
func NewParser() *DefaultParser {
	return &DefaultParser{}
}

// Parse converts raw bytes into a Profile.
func (p *DefaultParser) Parse(data []byte) (Profile, error) {
	_ = data
	return Profile{}, ErrNotImplemented
}
