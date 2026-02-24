package auth

import (
	"fmt"
	"testing"
)

type fakeCmd struct {
	output string
	err    error
}

func (f *fakeCmd) Run(name string, args ...string) (string, error) {
	return f.output, f.err
}

func TestResolveToken_GHAuthToken(t *testing.T) {
	r := &Resolver{
		Cmd:    &fakeCmd{output: "ghp_abc123\n"},
		Getenv: func(string) string { return "" },
	}
	tok := r.ResolveToken()
	if tok != "ghp_abc123" {
		t.Errorf("got %q, want %q", tok, "ghp_abc123")
	}
}

func TestResolveToken_EnvFallback(t *testing.T) {
	r := &Resolver{
		Cmd: &fakeCmd{err: fmt.Errorf("gh not found")},
		Getenv: func(key string) string {
			if key == "GITHUB_TOKEN" {
				return "env_token_456"
			}
			return ""
		},
	}
	tok := r.ResolveToken()
	if tok != "env_token_456" {
		t.Errorf("got %q, want %q", tok, "env_token_456")
	}
}

func TestResolveToken_Unauthenticated(t *testing.T) {
	r := &Resolver{
		Cmd:    &fakeCmd{err: fmt.Errorf("gh not found")},
		Getenv: func(string) string { return "" },
	}
	tok := r.ResolveToken()
	if tok != "" {
		t.Errorf("got %q, want empty string", tok)
	}
}

func TestResolveToken_GHPriority(t *testing.T) {
	r := &Resolver{
		Cmd: &fakeCmd{output: "ghp_from_gh\n"},
		Getenv: func(key string) string {
			if key == "GITHUB_TOKEN" {
				return "env_token"
			}
			return ""
		},
	}
	tok := r.ResolveToken()
	if tok != "ghp_from_gh" {
		t.Errorf("gh auth token should have priority; got %q", tok)
	}
}
