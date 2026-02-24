package provider_test

import (
	"context"
	"testing"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

func TestResultSuccess(t *testing.T) {
	r := provider.Result{
		Target: "github:facebook/react",
		GitHub: &provider.GitHubMetrics{
			Stars: 215000,
			Forks: 45000,
		},
	}
	if r.Target != "github:facebook/react" {
		t.Errorf("got target %q, want %q", r.Target, "github:facebook/react")
	}
	if r.Error != "" {
		t.Errorf("got error %q, want empty", r.Error)
	}
	if r.GitHub.Stars != 215000 {
		t.Errorf("got stars %d, want 215000", r.GitHub.Stars)
	}
}

func TestResultError(t *testing.T) {
	r := provider.Result{
		Target: "github:nonexistent/repo",
		Error:  "GitHub API: 404 Not Found",
	}
	if r.Error == "" {
		t.Error("expected error to be set")
	}
	if r.GitHub != nil {
		t.Error("expected GitHub to be nil on error")
	}
}

type stubProvider struct {
	result provider.Result
}

func (s *stubProvider) Scheme() string { return "stub" }

func (s *stubProvider) Fetch(_ context.Context, _ string) (provider.Result, error) {
	return s.result, nil
}

func TestProviderInterface(t *testing.T) {
	var p provider.Provider = &stubProvider{
		result: provider.Result{Target: "stub:test"},
	}
	r, err := p.Fetch(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Target != "stub:test" {
		t.Errorf("got target %q, want %q", r.Target, "stub:test")
	}
	if p.Scheme() != "stub" {
		t.Errorf("got scheme %q, want %q", p.Scheme(), "stub")
	}
}
