package pypi

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

// upload15dAgo returns an RFC3339 timestamp approximately 15 days in the past.
func upload15dAgo() string {
	return time.Now().Add(-15 * 24 * time.Hour).Format(time.RFC3339)
}

// setupMockServers creates a PyPI JSON API server and a pypistats.org API server.
func setupMockServers(t *testing.T) (pypiSrv *httptest.Server, statsSrv *httptest.Server) {
	t.Helper()

	pypiMux := http.NewServeMux()
	pypiMux.HandleFunc("GET /pypi/requests/json", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"info": map[string]any{
				"version":         "2.32.5",
				"license":         "Apache-2.0",
				"requires_python": ">=3.9",
				"requires_dist": []string{
					"certifi>=2017.4.17",
					"charset-normalizer<4,>=2",
					"idna<4,>=2.5",
					"urllib3<3,>=1.21.1",
					`PySocks!=1.5.7,>=1.5.6; extra == "socks"`,
					`chardet<6,>=3.0.2; extra == "security"`,
				},
			},
			"releases": map[string]any{
				"2.32.5": []map[string]any{
					{"upload_time_iso_8601": upload15dAgo()},
				},
			},
		})
	})

	statsMux := http.NewServeMux()
	statsMux.HandleFunc("GET /api/packages/requests/recent", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"data": map[string]any{
				"last_day":   40574147,
				"last_week":  269467251,
				"last_month": 1081975948,
			},
			"package": "requests",
			"type":    "recent_downloads",
		})
	})

	pypiSrv = httptest.NewServer(pypiMux)
	statsSrv = httptest.NewServer(statsMux)
	t.Cleanup(func() {
		pypiSrv.Close()
		statsSrv.Close()
	})
	return
}

func TestScheme(t *testing.T) {
	p := New("", "")
	if got := p.Scheme(); got != "pypi" {
		t.Errorf("Scheme() = %q, want %q", got, "pypi")
	}
}

func TestFetchSuccess(t *testing.T) {
	pypiSrv, statsSrv := setupMockServers(t)
	p := New(pypiSrv.URL, statsSrv.URL)

	result, err := p.Fetch(context.Background(), "requests")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.Target != "pypi:requests" {
		t.Errorf("target: got %q, want %q", result.Target, "pypi:requests")
	}

	m := result.PyPI
	if m == nil {
		t.Fatal("expected PyPI metrics to be set")
	}
	if m.WeeklyDownloads != 269467251 {
		t.Errorf("weekly_downloads: got %d, want 269467251", m.WeeklyDownloads)
	}
	if m.MonthlyDownloads != 1081975948 {
		t.Errorf("monthly_downloads: got %d, want 1081975948", m.MonthlyDownloads)
	}
	if m.LatestVersion != "2.32.5" {
		t.Errorf("latest_version: got %q, want %q", m.LatestVersion, "2.32.5")
	}
	if m.LastPublishDays < 14 || m.LastPublishDays > 16 {
		t.Errorf("last_publish_days: got %d, want ~15", m.LastPublishDays)
	}
	// 4 non-extra deps out of 6 total
	if m.DependenciesCount != 4 {
		t.Errorf("dependencies_count: got %d, want 4", m.DependenciesCount)
	}
	if m.License != "Apache-2.0" {
		t.Errorf("license: got %q, want %q", m.License, "Apache-2.0")
	}
	if m.RequiresPython != ">=3.9" {
		t.Errorf("requires_python: got %q, want %q", m.RequiresPython, ">=3.9")
	}
}

func TestFetchNotFound(t *testing.T) {
	pypiMux := http.NewServeMux()
	pypiMux.HandleFunc("GET /pypi/nonexistent-pkg/json", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	pypiSrv := httptest.NewServer(pypiMux)
	defer pypiSrv.Close()

	statsMux := http.NewServeMux()
	statsMux.HandleFunc("GET /api/packages/nonexistent-pkg/recent", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	statsSrv := httptest.NewServer(statsMux)
	defer statsSrv.Close()

	p := New(pypiSrv.URL, statsSrv.URL)
	result, err := p.Fetch(context.Background(), "nonexistent-pkg")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.Error == "" {
		t.Fatal("expected result.Error to be set for 404")
	}
	if result.PyPI != nil {
		t.Error("expected PyPI to be nil on full error")
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

func TestFetchPartialFailure(t *testing.T) {
	// PyPI JSON API succeeds
	pypiMux := http.NewServeMux()
	pypiMux.HandleFunc("GET /pypi/partial-pkg/json", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"info": map[string]any{
				"version":         "1.0.0",
				"license":         "MIT",
				"requires_python": ">=3.8",
				"requires_dist":   []string{"dep-a>=1.0"},
			},
			"releases": map[string]any{
				"1.0.0": []map[string]any{
					{"upload_time_iso_8601": time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339)},
				},
			},
		})
	})
	pypiSrv := httptest.NewServer(pypiMux)
	defer pypiSrv.Close()

	// Stats API returns 500
	statsMux := http.NewServeMux()
	statsMux.HandleFunc("GET /api/packages/partial-pkg/recent", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	statsSrv := httptest.NewServer(statsMux)
	defer statsSrv.Close()

	p := New(pypiSrv.URL, statsSrv.URL)
	result, err := p.Fetch(context.Background(), "partial-pkg")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	// Should have partial metrics
	if result.PyPI == nil {
		t.Fatal("expected PyPI metrics for partial failure")
	}
	if result.PyPI.LatestVersion != "1.0.0" {
		t.Errorf("latest_version: got %q, want %q", result.PyPI.LatestVersion, "1.0.0")
	}
	if result.PyPI.WeeklyDownloads != 0 {
		t.Errorf("weekly_downloads: got %d, want 0", result.PyPI.WeeklyDownloads)
	}
	if result.PyPI.MonthlyDownloads != 0 {
		t.Errorf("monthly_downloads: got %d, want 0", result.PyPI.MonthlyDownloads)
	}

	// Should also have error
	if result.Error == "" {
		t.Fatal("expected result.Error for partial failure")
	}
	if !strings.Contains(result.Error, "downloads") {
		t.Errorf("expected error to mention downloads, got %q", result.Error)
	}
}

func TestFetchNullRequiresDist(t *testing.T) {
	pypiMux := http.NewServeMux()
	pypiMux.HandleFunc("GET /pypi/null-deps/json", func(w http.ResponseWriter, _ *http.Request) {
		// requires_dist is null
		_, _ = w.Write([]byte(`{
			"info": {
				"version": "0.1.0",
				"license": "BSD",
				"requires_python": ">=3.7",
				"requires_dist": null
			},
			"releases": {
				"0.1.0": [{"upload_time_iso_8601": "` + time.Now().Add(-3*24*time.Hour).Format(time.RFC3339) + `"}]
			}
		}`))
	})
	pypiSrv := httptest.NewServer(pypiMux)
	defer pypiSrv.Close()

	statsMux := http.NewServeMux()
	statsMux.HandleFunc("GET /api/packages/null-deps/recent", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"data": map[string]any{"last_day": 10, "last_week": 100, "last_month": 500},
		})
	})
	statsSrv := httptest.NewServer(statsMux)
	defer statsSrv.Close()

	p := New(pypiSrv.URL, statsSrv.URL)
	result, err := p.Fetch(context.Background(), "null-deps")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.PyPI == nil {
		t.Fatal("expected PyPI metrics to be set")
	}
	if result.PyPI.DependenciesCount != 0 {
		t.Errorf("dependencies_count: got %d, want 0", result.PyPI.DependenciesCount)
	}
}

func TestFetchExtrasExcluded(t *testing.T) {
	pypiMux := http.NewServeMux()
	pypiMux.HandleFunc("GET /pypi/extras-pkg/json", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"info": map[string]any{
				"version":         "3.0.0",
				"license":         "MIT",
				"requires_python": ">=3.10",
				"requires_dist": []string{
					"core-dep>=1.0",
					"another-dep<2.0",
					`optional-a>=1.0; extra == "dev"`,
					`optional-b>=2.0; extra == "test"`,
					`optional-c>=3.0; extra == "docs"`,
				},
			},
			"releases": map[string]any{
				"3.0.0": []map[string]any{
					{"upload_time_iso_8601": time.Now().Add(-1 * 24 * time.Hour).Format(time.RFC3339)},
				},
			},
		})
	})
	pypiSrv := httptest.NewServer(pypiMux)
	defer pypiSrv.Close()

	statsMux := http.NewServeMux()
	statsMux.HandleFunc("GET /api/packages/extras-pkg/recent", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"data": map[string]any{"last_day": 100, "last_week": 1000, "last_month": 5000},
		})
	})
	statsSrv := httptest.NewServer(statsMux)
	defer statsSrv.Close()

	p := New(pypiSrv.URL, statsSrv.URL)
	result, err := p.Fetch(context.Background(), "extras-pkg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.PyPI == nil {
		t.Fatal("expected PyPI metrics to be set")
	}
	// Only 2 core deps, 3 extras excluded
	if result.PyPI.DependenciesCount != 2 {
		t.Errorf("dependencies_count: got %d, want 2", result.PyPI.DependenciesCount)
	}
}
