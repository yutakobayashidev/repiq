package cache

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

// mockProvider implements provider.Provider with a call counter.
type mockProvider struct {
	scheme string
	result provider.Result
	err    error
	calls  atomic.Int32
}

func (m *mockProvider) Scheme() string { return m.scheme }

func (m *mockProvider) Fetch(_ context.Context, identifier string) (provider.Result, error) {
	m.calls.Add(1)
	return m.result, m.err
}

func TestProviderCacheMiss(t *testing.T) {
	mock := &mockProvider{
		scheme: "github",
		result: provider.Result{
			Target: "github:facebook/react",
			GitHub: &provider.GitHubMetrics{Stars: 200000},
		},
	}
	store := NewStore(t.TempDir(), 24*time.Hour)
	p := NewProvider(mock, store, false)

	if p.Scheme() != "github" {
		t.Fatalf("Scheme = %q, want github", p.Scheme())
	}

	result, err := p.Fetch(context.Background(), "facebook/react")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if mock.calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", mock.calls.Load())
	}
	if result.GitHub.Stars != 200000 {
		t.Errorf("Stars = %d, want 200000", result.GitHub.Stars)
	}
}

func TestProviderCacheHit(t *testing.T) {
	mock := &mockProvider{
		scheme: "github",
		result: provider.Result{
			Target: "github:facebook/react",
			GitHub: &provider.GitHubMetrics{Stars: 200000},
		},
	}
	store := NewStore(t.TempDir(), 24*time.Hour)
	p := NewProvider(mock, store, false)

	// First fetch: cache miss → calls underlying
	_, _ = p.Fetch(context.Background(), "facebook/react")

	// Second fetch: cache hit → underlying NOT called
	result, err := p.Fetch(context.Background(), "facebook/react")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if mock.calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1 (cache hit should not call underlying)", mock.calls.Load())
	}
	if result.GitHub.Stars != 200000 {
		t.Errorf("Stars = %d, want 200000", result.GitHub.Stars)
	}
}

func TestProviderNoCache(t *testing.T) {
	mock := &mockProvider{
		scheme: "npm",
		result: provider.Result{
			Target: "npm:react",
			NPM:    &provider.NPMMetrics{WeeklyDownloads: 5000000},
		},
	}
	store := NewStore(t.TempDir(), 24*time.Hour)
	p := NewProvider(mock, store, true) // noCache = true

	_, _ = p.Fetch(context.Background(), "react")
	_, _ = p.Fetch(context.Background(), "react")

	if mock.calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2 (noCache should always call underlying)", mock.calls.Load())
	}
}

func TestProviderErrorNotCached(t *testing.T) {
	mock := &mockProvider{
		scheme: "github",
		result: provider.Result{
			Target: "github:nonexistent/repo",
			Error:  "GitHub API: 404 Not Found",
		},
	}
	store := NewStore(t.TempDir(), 24*time.Hour)
	p := NewProvider(mock, store, false)

	_, _ = p.Fetch(context.Background(), "nonexistent/repo")
	_, _ = p.Fetch(context.Background(), "nonexistent/repo")

	if mock.calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2 (error results should not be cached)", mock.calls.Load())
	}
}

func TestProviderPartialErrorNotCached(t *testing.T) {
	mock := &mockProvider{
		scheme: "github",
		result: provider.Result{
			Target: "github:partial/repo",
			GitHub: &provider.GitHubMetrics{Stars: 100},
			Error:  "contributors: unexpected status 500",
		},
	}
	store := NewStore(t.TempDir(), 24*time.Hour)
	p := NewProvider(mock, store, false)

	_, _ = p.Fetch(context.Background(), "partial/repo")
	_, _ = p.Fetch(context.Background(), "partial/repo")

	if mock.calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2 (partial error results should not be cached)", mock.calls.Load())
	}
}

func TestProviderTTLExpiry(t *testing.T) {
	mock := &mockProvider{
		scheme: "npm",
		result: provider.Result{
			Target: "npm:react",
			NPM:    &provider.NPMMetrics{WeeklyDownloads: 5000000},
		},
	}
	store := NewStore(t.TempDir(), 1*time.Millisecond)
	p := NewProvider(mock, store, false)

	_, _ = p.Fetch(context.Background(), "react")
	time.Sleep(5 * time.Millisecond)
	_, _ = p.Fetch(context.Background(), "react")

	if mock.calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2 (TTL expiry should re-fetch)", mock.calls.Load())
	}
}
