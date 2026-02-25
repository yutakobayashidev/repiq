package golang

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mustEncode(w http.ResponseWriter, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		panic(err)
	}
}

func setupMockServers(t *testing.T) (proxy *httptest.Server, depsdev *httptest.Server) {
	t.Helper()

	publishTime := time.Now().Add(-15 * 24 * time.Hour)

	proxyMux := http.NewServeMux()
	// GET /golang.org/x/text/@latest
	proxyMux.HandleFunc("GET /golang.org/x/text/@latest", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"Version": "v0.34.0",
			"Time":    publishTime.Format(time.RFC3339),
		})
	})

	depsdevMux := http.NewServeMux()
	// GET /v3alpha/systems/go/packages/golang.org%2Fx%2Ftext/versions/v0.34.0
	depsdevMux.HandleFunc("GET /v3alpha/systems/go/packages/golang.org%2Fx%2Ftext/versions/v0.34.0", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"licenses": []string{"BSD-3-Clause"},
		})
	})
	// GET /v3alpha/systems/go/packages/golang.org%2Fx%2Ftext/versions/v0.34.0:requirements
	depsdevMux.HandleFunc("GET /v3alpha/systems/go/packages/golang.org%2Fx%2Ftext/versions/v0.34.0:requirements", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"go": map[string]any{
				"directDependencies": []map[string]any{
					{"name": "golang.org/x/tools", "requirement": "v0.21.0"},
					{"name": "golang.org/x/mod", "requirement": "v0.17.0"},
				},
			},
		})
	})

	proxy = httptest.NewServer(proxyMux)
	depsdev = httptest.NewServer(depsdevMux)
	t.Cleanup(func() {
		proxy.Close()
		depsdev.Close()
	})
	return
}

func TestScheme(t *testing.T) {
	p := New("", "")
	if got := p.Scheme(); got != "go" {
		t.Errorf("Scheme() = %q, want %q", got, "go")
	}
}

func TestFetchSuccess(t *testing.T) {
	proxy, depsdev := setupMockServers(t)
	p := New(proxy.URL, depsdev.URL)

	result, err := p.Fetch(context.Background(), "golang.org/x/text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.Target != "go:golang.org/x/text" {
		t.Errorf("target: got %q, want %q", result.Target, "go:golang.org/x/text")
	}

	g := result.Go
	if g == nil {
		t.Fatal("expected Go metrics to be set")
	}
	if g.LatestVersion != "v0.34.0" {
		t.Errorf("latest_version: got %q, want %q", g.LatestVersion, "v0.34.0")
	}
	if g.LastPublishDays < 14 || g.LastPublishDays > 16 {
		t.Errorf("last_publish_days: got %d, want ~15", g.LastPublishDays)
	}
	if g.DependenciesCount != 2 {
		t.Errorf("dependencies_count: got %d, want 2", g.DependenciesCount)
	}
	if g.License != "BSD-3-Clause" {
		t.Errorf("license: got %q, want %q", g.License, "BSD-3-Clause")
	}
}

func TestFetchNotFound(t *testing.T) {
	proxyMux := http.NewServeMux()
	proxyMux.HandleFunc("GET /github.com/nonexistent/pkg/@latest", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	proxy := httptest.NewServer(proxyMux)
	defer proxy.Close()

	depsdev := httptest.NewServer(http.NewServeMux())
	defer depsdev.Close()

	p := New(proxy.URL, depsdev.URL)
	result, err := p.Fetch(context.Background(), "github.com/nonexistent/pkg")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.Error == "" {
		t.Fatal("expected result.Error to be set for 404")
	}
	if result.Go != nil {
		t.Error("expected Go to be nil on proxy error")
	}
}

func TestFetchEmptyIdentifier(t *testing.T) {
	p := New("http://unused", "http://unused")
	result, err := p.Fetch(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.Error == "" {
		t.Fatal("expected result.Error for empty identifier")
	}
}

func TestFetchDepsDevFailure(t *testing.T) {
	publishTime := time.Now().Add(-15 * 24 * time.Hour)

	proxyMux := http.NewServeMux()
	proxyMux.HandleFunc("GET /golang.org/x/text/@latest", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"Version": "v0.34.0",
			"Time":    publishTime.Format(time.RFC3339),
		})
	})
	proxy := httptest.NewServer(proxyMux)
	defer proxy.Close()

	// deps.dev returns 500 for everything
	depsdevMux := http.NewServeMux()
	depsdevMux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	depsdev := httptest.NewServer(depsdevMux)
	defer depsdev.Close()

	p := New(proxy.URL, depsdev.URL)
	result, err := p.Fetch(context.Background(), "golang.org/x/text")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	// Should have partial metrics
	g := result.Go
	if g == nil {
		t.Fatal("expected Go metrics for partial failure (proxy succeeded)")
	}
	if g.LatestVersion != "v0.34.0" {
		t.Errorf("latest_version: got %q, want %q", g.LatestVersion, "v0.34.0")
	}
	if g.LastPublishDays < 14 || g.LastPublishDays > 16 {
		t.Errorf("last_publish_days: got %d, want ~15", g.LastPublishDays)
	}
	if g.DependenciesCount != 0 {
		t.Errorf("dependencies_count: got %d, want 0 (deps.dev failed)", g.DependenciesCount)
	}
	if g.License != "" {
		t.Errorf("license: got %q, want %q (deps.dev failed)", g.License, "")
	}
	// Should also have error
	if result.Error == "" {
		t.Fatal("expected result.Error for deps.dev failure")
	}
}

func TestModulePathEscape(t *testing.T) {
	publishTime := time.Now().Add(-10 * 24 * time.Hour)

	proxyMux := http.NewServeMux()
	// Azure → !azure in escaped path
	proxyMux.HandleFunc("GET /github.com/!azure/azure-sdk/@latest", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"Version": "v1.0.0",
			"Time":    publishTime.Format(time.RFC3339),
		})
	})
	proxy := httptest.NewServer(proxyMux)
	defer proxy.Close()

	depsdevMux := http.NewServeMux()
	depsdevMux.HandleFunc(fmt.Sprintf("GET /v3alpha/systems/go/packages/%s/versions/v1.0.0",
		"github.com%2FAzure%2Fazure-sdk"), func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"licenses": []string{"MIT"},
		})
	})
	depsdevMux.HandleFunc(fmt.Sprintf("GET /v3alpha/systems/go/packages/%s/versions/v1.0.0:requirements",
		"github.com%2FAzure%2Fazure-sdk"), func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"go": map[string]any{
				"directDependencies": []map[string]any{},
			},
		})
	})
	depsdev := httptest.NewServer(depsdevMux)
	defer depsdev.Close()

	p := New(proxy.URL, depsdev.URL)
	result, err := p.Fetch(context.Background(), "github.com/Azure/azure-sdk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	g := result.Go
	if g == nil {
		t.Fatal("expected Go metrics to be set")
	}
	if g.LatestVersion != "v1.0.0" {
		t.Errorf("latest_version: got %q, want %q", g.LatestVersion, "v1.0.0")
	}
	if g.License != "MIT" {
		t.Errorf("license: got %q, want %q", g.License, "MIT")
	}
}

func TestMajorVersionSuffix(t *testing.T) {
	publishTime := time.Now().Add(-5 * 24 * time.Hour)

	proxyMux := http.NewServeMux()
	proxyMux.HandleFunc("GET /github.com/owner/repo/v2/@latest", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"Version": "v2.3.0",
			"Time":    publishTime.Format(time.RFC3339),
		})
	})
	proxy := httptest.NewServer(proxyMux)
	defer proxy.Close()

	depsdevMux := http.NewServeMux()
	depsdevMux.HandleFunc(fmt.Sprintf("GET /v3alpha/systems/go/packages/%s/versions/v2.3.0",
		"github.com%2Fowner%2Frepo%2Fv2"), func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"licenses": []string{"Apache-2.0"},
		})
	})
	depsdevMux.HandleFunc(fmt.Sprintf("GET /v3alpha/systems/go/packages/%s/versions/v2.3.0:requirements",
		"github.com%2Fowner%2Frepo%2Fv2"), func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"go": map[string]any{
				"directDependencies": []map[string]any{
					{"name": "github.com/other/dep", "requirement": "v1.0.0"},
				},
			},
		})
	})
	depsdev := httptest.NewServer(depsdevMux)
	defer depsdev.Close()

	p := New(proxy.URL, depsdev.URL)
	result, err := p.Fetch(context.Background(), "github.com/owner/repo/v2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	g := result.Go
	if g == nil {
		t.Fatal("expected Go metrics to be set")
	}
	if g.LatestVersion != "v2.3.0" {
		t.Errorf("latest_version: got %q, want %q", g.LatestVersion, "v2.3.0")
	}
	if g.DependenciesCount != 1 {
		t.Errorf("dependencies_count: got %d, want 1", g.DependenciesCount)
	}
	if g.License != "Apache-2.0" {
		t.Errorf("license: got %q, want %q", g.License, "Apache-2.0")
	}
	if !strings.Contains(result.Target, "github.com/owner/repo/v2") {
		t.Errorf("target should contain module path, got %q", result.Target)
	}
}
