package provider

import (
	"fmt"
	"strings"
)

// Target represents a parsed <scheme>:<identifier> input.
type Target struct {
	Scheme     string
	Identifier string
}

// ParseTarget parses a string in the form "scheme:identifier".
func ParseTarget(s string) (Target, error) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return Target{}, fmt.Errorf("invalid target %q: missing ':'", s)
	}
	scheme := s[:idx]
	id := s[idx+1:]
	if scheme == "" {
		return Target{}, fmt.Errorf("invalid target %q: empty scheme", s)
	}
	if id == "" {
		return Target{}, fmt.Errorf("invalid target %q: empty identifier", s)
	}
	return Target{Scheme: scheme, Identifier: id}, nil
}
