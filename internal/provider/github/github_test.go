package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mustEncode(w http.ResponseWriter, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		panic(err)
	}
}

func setupMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /repos/owner/repo", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"stargazers_count":  1000,
			"forks_count":      200,
			"open_issues_count": 50,
		})
	})

	mux.HandleFunc("GET /repos/owner/repo/contributors", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Link", `<https://api.github.com/repos/owner/repo/contributors?per_page=1&page=42>; rel="last"`)
		mustEncode(w, []map[string]any{{"login": "user1"}})
	})

	mux.HandleFunc("GET /repos/owner/repo/releases", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Link", `<https://api.github.com/repos/owner/repo/releases?per_page=1&page=15>; rel="last"`)
		mustEncode(w, []map[string]any{{"tag_name": "v1.0"}})
	})

	mux.HandleFunc("GET /repos/owner/repo/commits", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, []map[string]any{
			{
				"sha": "abc123",
				"commit": map[string]any{
					"committer": map[string]any{
						"date": "2026-02-24T12:00:00Z",
					},
				},
			},
		})
	})

	mux.HandleFunc("GET /search/commits", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{"total_count": 120, "items": []any{}})
	})

	mux.HandleFunc("GET /search/issues", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{"total_count": 340, "items": []any{}})
	})

	return httptest.NewServer(mux)
}

func TestFetchSuccess(t *testing.T) {
	srv := setupMockServer(t)
	defer srv.Close()

	p := New("", srv.URL+"/")
	result, err := p.Fetch(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.Target != "github:owner/repo" {
		t.Errorf("target: got %q, want %q", result.Target, "github:owner/repo")
	}
	g := result.GitHub
	if g == nil {
		t.Fatal("expected GitHub metrics to be set")
	}
	if g.Stars != 1000 {
		t.Errorf("stars: got %d, want 1000", g.Stars)
	}
	if g.Forks != 200 {
		t.Errorf("forks: got %d, want 200", g.Forks)
	}
	if g.OpenIssues != 50 {
		t.Errorf("open_issues: got %d, want 50", g.OpenIssues)
	}
	if g.Contributors != 42 {
		t.Errorf("contributors: got %d, want 42", g.Contributors)
	}
	if g.ReleaseCount != 15 {
		t.Errorf("release_count: got %d, want 15", g.ReleaseCount)
	}
	if g.LastCommitDays < 0 {
		t.Errorf("last_commit_days: got %d, want >= 0", g.LastCommitDays)
	}
	if g.Commits30d != 120 {
		t.Errorf("commits_30d: got %d, want 120", g.Commits30d)
	}
	if g.IssuesClosed30d != 340 {
		t.Errorf("issues_closed_30d: got %d, want 340", g.IssuesClosed30d)
	}
}

func TestFetchNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/bad/repo", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		mustEncode(w, map[string]string{"message": "Not Found"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	p := New("", srv.URL+"/")
	result, err := p.Fetch(context.Background(), "bad/repo")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.Error == "" {
		t.Fatal("expected result.Error to be set for 404")
	}
	if result.GitHub != nil {
		t.Error("expected GitHub to be nil on error")
	}
}

func TestFetchInvalidIdentifier(t *testing.T) {
	p := New("", "http://unused/")
	result, err := p.Fetch(context.Background(), "noslash")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.Error == "" {
		t.Fatal("expected result.Error for invalid identifier")
	}
}

func TestScheme(t *testing.T) {
	p := New("", "")
	if p.Scheme() != "github" {
		t.Errorf("got %q, want %q", p.Scheme(), "github")
	}
}

func TestFetchMaliciousIdentifier(t *testing.T) {
	p := New("", "http://unused/")
	result, err := p.Fetch(context.Background(), "owner/repo is:public")
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.Error == "" {
		t.Fatal("expected result.Error for malicious identifier")
	}
}

func TestFetchNoLinkHeader(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/owner/small", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{
			"stargazers_count": 5,
			"forks_count":     1,
			"open_issues":     0,
		})
	})
	mux.HandleFunc("GET /repos/owner/small/contributors", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, []map[string]any{{"login": "user1"}, {"login": "user2"}})
	})
	mux.HandleFunc("GET /repos/owner/small/releases", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, []map[string]any{{"tag_name": "v0.1"}})
	})
	mux.HandleFunc("GET /repos/owner/small/commits", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, []map[string]any{
			{"sha": "abc", "commit": map[string]any{
				"committer": map[string]any{"date": "2026-02-25T00:00:00Z"},
			}},
		})
	})
	mux.HandleFunc("GET /search/commits", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{"total_count": 3, "items": []any{}})
	})
	mux.HandleFunc("GET /search/issues", func(w http.ResponseWriter, _ *http.Request) {
		mustEncode(w, map[string]any{"total_count": 1, "items": []any{}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	p := New("", srv.URL+"/")
	result, err := p.Fetch(context.Background(), "owner/small")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("unexpected result error: %s", result.Error)
	}
	if result.GitHub.Contributors != 2 {
		t.Errorf("contributors: got %d, want 2", result.GitHub.Contributors)
	}
	if result.GitHub.ReleaseCount != 1 {
		t.Errorf("releases: got %d, want 1", result.GitHub.ReleaseCount)
	}
}
