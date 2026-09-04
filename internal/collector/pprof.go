package collector

import (
	"context"

	"github.com/KoJaco/leakwatch/internal/profile"
)

// PProfCollector fetches and parses a goroutine leak profile from a pprof endpoint.
type PProfCollector struct {
	source profile.Source
	parser profile.Parser
}

// NewPProfCollector returns a Collector backed by the given source and parser.
func NewPProfCollector(source profile.Source, parser profile.Parser) *PProfCollector {
	return &PProfCollector{
		source: source,
		parser: parser,
	}
}

// Collect fetches raw profile bytes and parses them into a domain Profile.
func (c *PProfCollector) Collect(ctx context.Context) (profile.Profile, error) {
	data, err := c.source.Fetch(ctx)
	if err != nil {
		return profile.Profile{}, err
	}
	return c.parser.Parse(data)
}

// FixtureCollector returns a fixed profile — useful for tests.
type FixtureCollector struct {
	Profile profile.Profile
}

// Collect returns the configured fixture profile.
func (c *FixtureCollector) Collect(context.Context) (profile.Profile, error) {
	return c.Profile, nil
}
