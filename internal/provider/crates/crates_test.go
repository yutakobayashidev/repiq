package crates

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func mustEncode(w http.ResponseWriter, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		panic(err)
	}
}

func setupMockServer(t *testing.T) *httptest.Server {
	t.Helper()

	published15d := time.Now().Add(-15 * 24 * time.Hour).Format(time.RFC3339)

	mux := http.NewServeMux()

	// GET /api/v1/crates/serde — metadata
	mux.HandleFunc("GET /api/v1/crates/serde", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"crate": map[string]any{
				"downloads":          835433978,
				"recent_downloads":   116815564,
				"max_stable_version": "1.0.228",
				"newest_version":     "1.0.228",
			},
			"versions": []map[string]any{
				{
					"num":        "1.0.228",
					"license":    "MIT OR Apache-2.0",
					"created_at": published15d,
				},
				{
					"num":        "1.0.227",
					"license":    "MIT OR Apache-2.0",
					"created_at": "2024-01-01T00:00:00Z",
				},
			},
		})
	})

	// GET /api/v1/crates/serde/1.0.228/dependencies
	mux.HandleFunc("GET /api/v1/crates/serde/1.0.228/dependencies", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"dependencies": []map[string]any{
				{"crate_id": "serde_derive", "kind": "normal", "optional": false},
				{"crate_id": "serde_test", "kind": "dev", "optional": false},
				{"crate_id": "serde_json", "kind": "normal", "optional": false},
			},
		})
	})

	// GET /api/v1/crates/serde/reverse_dependencies
	mux.HandleFunc("GET /api/v1/crates/serde/reverse_dependencies", func(w http.ResponseWriter, r *http.Request) {
		mustEncode(w, map[string]any{
			"dependencies": []any{},
			"meta":         map[string]any{"total": 72719},
		})
	})

	return httptest.NewServer(mux)
}

func TestScheme(t *testing.T) {
	p := New("")
	if p.Scheme() != "crates" {
		t.Errorf("got %q, want %q", p.Scheme(), "crates")
	}
}

func TestFetchSuccess(t *testing.T) {
	srv := setupMockServer(t)
	defer srv.Close()

	p := New(srv.URL)
	result, err := p.Fetch(context.Background(), "serde")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.Target != "crates:serde" {
		t.Errorf("target: got %q, want %q", result.Target, "crates:serde")
	}
	c := result.Crates
	if c == nil {
		t.Fatal("expected Crates metrics to be set")
	}
	if c.Downloads != 835433978 {
		t.Errorf("downloads: got %d, want 835433978", c.Downloads)
	}
	if c.RecentDownloads != 116815564 {
		t.Errorf("recent_downloads: got %d, want 116815564", c.RecentDownloads)
	}
	if c.LatestVersion != "1.0.228" {
		t.Errorf("latest_version: got %q, want %q", c.LatestVersion, "1.0.228")
	}
	if c.LastPublishDays < 14 || c.LastPublishDays > 16 {
		t.Errorf("last_publish_days: got %d, want ~15", c.LastPublishDays)
	}
	if c.DependenciesCount != 2 {
		t.Errorf("dependencies_count: got %d, want 2 (normal kind only)", c.DependenciesCount)
	}
	if c.License != "MIT OR Apache-2.0" {
		t.Errorf("license: got %q, want %q", c.License, "MIT OR Apache-2.0")
	}
	if c.ReverseDependencies != 72719 {
		t.Errorf("reverse_dependencies: got %d, want 72719", c.ReverseDependencies)
	}
}

func TestFetchNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/crates/nonexistent-crate", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		mustEncode(w, map[string]any{"errors": []map[string]any{{"detail": "Not Found"}}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	p := New(srv.URL)
	result, err := p.Fetch(context.Background(), "nonexistent-crate")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.Error == "" {
		t.Fatal("expected result.Error to be set for 404")
	}
	if result.Crates != nil {
		t.Error("expected Crates to be nil on error")
	}
}

func TestFetchEmptyIdentifier(t *testing.T) {
	p := New("http://unused")
	result, err := p.Fetch(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.Error == "" {
		t.Fatal("expected result.Error for empty identifier")
	}
}

func TestFetchPartialFailure(t *testing.T) {
	published := time.Now().Add(-10 * 24 * time.Hour).Format(time.RFC3339)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/crates/partial-crate", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"crate": map[string]any{
				"downloads":          1000,
				"recent_downloads":   200,
				"max_stable_version": "0.5.0",
				"newest_version":     "0.5.0",
			},
			"versions": []map[string]any{
				{
					"num":        "0.5.0",
					"license":    "MIT",
					"created_at": published,
				},
			},
		})
	})

	mux.HandleFunc("GET /api/v1/crates/partial-crate/0.5.0/dependencies", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	mux.HandleFunc("GET /api/v1/crates/partial-crate/reverse_dependencies", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	p := New(srv.URL)
	result, err := p.Fetch(context.Background(), "partial-crate")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	// Should have partial metrics
	if result.Crates == nil {
		t.Fatal("expected Crates metrics for partial failure")
	}
	if result.Crates.Downloads != 1000 {
		t.Errorf("downloads: got %d, want 1000", result.Crates.Downloads)
	}
	if result.Crates.LatestVersion != "0.5.0" {
		t.Errorf("latest_version: got %q, want %q", result.Crates.LatestVersion, "0.5.0")
	}
	if result.Crates.License != "MIT" {
		t.Errorf("license: got %q, want %q", result.Crates.License, "MIT")
	}

	// Should also have error
	if result.Error == "" {
		t.Fatal("expected result.Error for partial failure")
	}
}

func TestFetchPreReleaseOnly(t *testing.T) {
	published := time.Now().Add(-3 * 24 * time.Hour).Format(time.RFC3339)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/crates/prerelease-crate", func(w http.ResponseWriter, _ *http.Request) {
		// max_stable_version is null — should fall back to newest_version
		_, _ = w.Write([]byte(`{
			"crate": {
				"downloads": 500,
				"recent_downloads": 100,
				"max_stable_version": null,
				"newest_version": "2.0.0-beta.1"
			},
			"versions": [
				{
					"num": "2.0.0-beta.1",
					"license": "Apache-2.0",
					"created_at": "` + published + `"
				}
			]
		}`))
	})

	mux.HandleFunc("GET /api/v1/crates/prerelease-crate/2.0.0-beta.1/dependencies", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"dependencies": []map[string]any{},
		})
	})

	mux.HandleFunc("GET /api/v1/crates/prerelease-crate/reverse_dependencies", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"dependencies": []any{},
			"meta":         map[string]any{"total": 5},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	p := New(srv.URL)
	result, err := p.Fetch(context.Background(), "prerelease-crate")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.Crates == nil {
		t.Fatal("expected Crates metrics to be set")
	}
	if result.Crates.LatestVersion != "2.0.0-beta.1" {
		t.Errorf("latest_version: got %q, want %q", result.Crates.LatestVersion, "2.0.0-beta.1")
	}
	if result.Crates.ReverseDependencies != 5 {
		t.Errorf("reverse_dependencies: got %d, want 5", result.Crates.ReverseDependencies)
	}
}

func TestUserAgentHeader(t *testing.T) {
	var mu sync.Mutex
	var capturedUA string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/crates/test-ua", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		capturedUA = r.Header.Get("User-Agent")
		mu.Unlock()

		published := time.Now().Format(time.RFC3339)
		mustEncode(w, map[string]any{
			"crate": map[string]any{
				"downloads":          1,
				"recent_downloads":   1,
				"max_stable_version": "0.1.0",
				"newest_version":     "0.1.0",
			},
			"versions": []map[string]any{
				{"num": "0.1.0", "license": "MIT", "created_at": published},
			},
		})
	})
	mux.HandleFunc("GET /api/v1/crates/test-ua/0.1.0/dependencies", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{"dependencies": []any{}})
	})
	mux.HandleFunc("GET /api/v1/crates/test-ua/reverse_dependencies", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{"dependencies": []any{}, "meta": map[string]any{"total": 0}})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	p := New(srv.URL)
	_, err := p.Fetch(context.Background(), "test-ua")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	ua := capturedUA
	mu.Unlock()

	if !strings.Contains(ua, "repiq") {
		t.Errorf("User-Agent header: got %q, want it to contain %q", ua, "repiq")
	}
}
