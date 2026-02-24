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

func TestResultNPMSuccess(t *testing.T) {
	r := provider.Result{
		Target: "npm:react",
		NPM: &provider.NPMMetrics{
			WeeklyDownloads:   25000000,
			LatestVersion:     "19.1.0",
			LastPublishDays:   15,
			DependenciesCount: 2,
			License:           "MIT",
		},
	}
	if r.Target != "npm:react" {
		t.Errorf("got target %q, want %q", r.Target, "npm:react")
	}
	if r.Error != "" {
		t.Errorf("got error %q, want empty", r.Error)
	}
	if r.GitHub != nil {
		t.Error("expected GitHub to be nil for npm result")
	}
	if r.NPM == nil {
		t.Fatal("expected NPM to be non-nil")
	}
	if r.NPM.WeeklyDownloads != 25000000 {
		t.Errorf("got weekly_downloads %d, want 25000000", r.NPM.WeeklyDownloads)
	}
	if r.NPM.LatestVersion != "19.1.0" {
		t.Errorf("got latest_version %q, want %q", r.NPM.LatestVersion, "19.1.0")
	}
	if r.NPM.License != "MIT" {
		t.Errorf("got license %q, want %q", r.NPM.License, "MIT")
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
