package npm

import (
	"context"
	"encoding/json"
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

func setupMockServers(t *testing.T) (registry *httptest.Server, downloads *httptest.Server) {
	t.Helper()

	regMux := http.NewServeMux()

	// GET /react/latest
	regMux.HandleFunc("GET /react/latest", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"name":    "react",
			"version": "19.1.0",
			"license": "MIT",
			"dependencies": map[string]any{
				"loose-envify":    "^1.1.0",
				"object-assign":   "^4.1.1",
			},
		})
	})

	// GET /react (abbreviated metadata)
	modified15d := time.Now().Add(-15 * 24 * time.Hour).Format("2006-01-02T15:04:05.000Z")
	regMux.HandleFunc("GET /react", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.npm.install-v1+json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mustEncode(w, map[string]any{
			"modified": modified15d,
		})
	})

	dlMux := http.NewServeMux()

	// GET /downloads/point/last-week/react
	dlMux.HandleFunc("GET /downloads/point/last-week/react", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"downloads": 25000000,
			"package":   "react",
		})
	})

	registry = httptest.NewServer(regMux)
	downloads = httptest.NewServer(dlMux)
	t.Cleanup(func() {
		registry.Close()
		downloads.Close()
	})
	return
}

func TestScheme(t *testing.T) {
	p := New("", "")
	if p.Scheme() != "npm" {
		t.Errorf("got %q, want %q", p.Scheme(), "npm")
	}
}

func TestFetchSuccess(t *testing.T) {
	reg, dl := setupMockServers(t)
	p := New(reg.URL, dl.URL)

	result, err := p.Fetch(context.Background(), "react")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.Target != "npm:react" {
		t.Errorf("target: got %q, want %q", result.Target, "npm:react")
	}
	n := result.NPM
	if n == nil {
		t.Fatal("expected NPM metrics to be set")
	}
	if n.WeeklyDownloads != 25000000 {
		t.Errorf("weekly_downloads: got %d, want 25000000", n.WeeklyDownloads)
	}
	if n.LatestVersion != "19.1.0" {
		t.Errorf("latest_version: got %q, want %q", n.LatestVersion, "19.1.0")
	}
	if n.DependenciesCount != 2 {
		t.Errorf("dependencies_count: got %d, want 2", n.DependenciesCount)
	}
	if n.License != "MIT" {
		t.Errorf("license: got %q, want %q", n.License, "MIT")
	}
	if n.LastPublishDays < 14 || n.LastPublishDays > 16 {
		t.Errorf("last_publish_days: got %d, want ~15", n.LastPublishDays)
	}
}

func TestFetchNotFound(t *testing.T) {
	regMux := http.NewServeMux()
	regMux.HandleFunc("GET /nonexistent-pkg/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	regMux.HandleFunc("GET /nonexistent-pkg", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	reg := httptest.NewServer(regMux)
	defer reg.Close()

	dlMux := http.NewServeMux()
	dlMux.HandleFunc("GET /downloads/point/last-week/nonexistent-pkg", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	dl := httptest.NewServer(dlMux)
	defer dl.Close()

	p := New(reg.URL, dl.URL)
	result, err := p.Fetch(context.Background(), "nonexistent-pkg")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.Error == "" {
		t.Fatal("expected result.Error to be set for 404")
	}
	if result.NPM != nil {
		t.Error("expected NPM to be nil on full error")
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

func TestFetchScopedPackage(t *testing.T) {
	regMux := http.NewServeMux()
	regMux.HandleFunc("GET /@types/node/latest", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"name":         "@types/node",
			"version":      "22.13.0",
			"license":      "MIT",
			"dependencies": map[string]any{"undici-types": "~6.20.0"},
		})
	})
	regMux.HandleFunc("GET /@types/node", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.npm.install-v1+json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mustEncode(w, map[string]any{
			"modified": time.Now().Add(-5 * 24 * time.Hour).Format("2006-01-02T15:04:05.000Z"),
		})
	})
	reg := httptest.NewServer(regMux)
	defer reg.Close()

	dlMux := http.NewServeMux()
	// scoped package: url.PathEscape encodes "/" as %2F, which Go's ServeMux
	// doesn't decode for pattern matching. Use a wildcard to match.
	dlMux.HandleFunc("GET /downloads/point/last-week/{pkg...}", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{"downloads": 5000000})
	})
	dl := httptest.NewServer(dlMux)
	defer dl.Close()

	p := New(reg.URL, dl.URL)
	result, err := p.Fetch(context.Background(), "@types/node")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.NPM == nil {
		t.Fatal("expected NPM metrics to be set")
	}
	if result.NPM.LatestVersion != "22.13.0" {
		t.Errorf("latest_version: got %q, want %q", result.NPM.LatestVersion, "22.13.0")
	}
	if result.NPM.WeeklyDownloads != 5000000 {
		t.Errorf("weekly_downloads: got %d, want 5000000", result.NPM.WeeklyDownloads)
	}
}

func TestFetchLicenseObject(t *testing.T) {
	regMux := http.NewServeMux()
	regMux.HandleFunc("GET /old-pkg/latest", func(w http.ResponseWriter, _ *http.Request) {
		// license as object (legacy format)
		_, _ = w.Write([]byte(`{"name":"old-pkg","version":"1.0.0","license":{"type":"ISC","url":"https://example.com"},"dependencies":{}}`))
	})
	regMux.HandleFunc("GET /old-pkg", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.npm.install-v1+json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mustEncode(w, map[string]any{"modified": time.Now().Add(-55 * 24 * time.Hour).Format("2006-01-02T15:04:05.000Z")})
	})
	reg := httptest.NewServer(regMux)
	defer reg.Close()

	dlMux := http.NewServeMux()
	dlMux.HandleFunc("GET /downloads/point/last-week/old-pkg", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{"downloads": 100})
	})
	dl := httptest.NewServer(dlMux)
	defer dl.Close()

	p := New(reg.URL, dl.URL)
	result, err := p.Fetch(context.Background(), "old-pkg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.NPM.License != "ISC" {
		t.Errorf("license: got %q, want %q", result.NPM.License, "ISC")
	}
}

func TestFetchNoDependencies(t *testing.T) {
	regMux := http.NewServeMux()
	regMux.HandleFunc("GET /no-deps/latest", func(w http.ResponseWriter, _ *http.Request) {
		// no dependencies field at all
		_, _ = w.Write([]byte(`{"name":"no-deps","version":"1.0.0","license":"MIT"}`))
	})
	regMux.HandleFunc("GET /no-deps", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.npm.install-v1+json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mustEncode(w, map[string]any{"modified": time.Now().Add(-24 * 24 * time.Hour).Format("2006-01-02T15:04:05.000Z")})
	})
	reg := httptest.NewServer(regMux)
	defer reg.Close()

	dlMux := http.NewServeMux()
	dlMux.HandleFunc("GET /downloads/point/last-week/no-deps", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{"downloads": 500})
	})
	dl := httptest.NewServer(dlMux)
	defer dl.Close()

	p := New(reg.URL, dl.URL)
	result, err := p.Fetch(context.Background(), "no-deps")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NPM.DependenciesCount != 0 {
		t.Errorf("dependencies_count: got %d, want 0", result.NPM.DependenciesCount)
	}
}

func TestFetchPartialFailure(t *testing.T) {
	regMux := http.NewServeMux()
	regMux.HandleFunc("GET /partial/latest", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"name":    "partial",
			"version": "2.0.0",
			"license": "Apache-2.0",
		})
	})
	regMux.HandleFunc("GET /partial", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.npm.install-v1+json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mustEncode(w, map[string]any{"modified": time.Now().Add(-10 * 24 * time.Hour).Format("2006-01-02T15:04:05.000Z")})
	})
	reg := httptest.NewServer(regMux)
	defer reg.Close()

	// downloads API returns 500
	dlMux := http.NewServeMux()
	dlMux.HandleFunc("GET /downloads/point/last-week/partial", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	dl := httptest.NewServer(dlMux)
	defer dl.Close()

	p := New(reg.URL, dl.URL)
	result, err := p.Fetch(context.Background(), "partial")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	// Should have partial metrics
	if result.NPM == nil {
		t.Fatal("expected NPM metrics for partial failure")
	}
	if result.NPM.LatestVersion != "2.0.0" {
		t.Errorf("latest_version: got %q, want %q", result.NPM.LatestVersion, "2.0.0")
	}
	// Should also have error
	if result.Error == "" {
		t.Fatal("expected result.Error for partial failure")
	}
	if !strings.Contains(result.Error, "downloads") {
		t.Errorf("expected error to mention downloads, got %q", result.Error)
	}
}
