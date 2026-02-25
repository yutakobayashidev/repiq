package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

func sampleResults() []provider.Result {
	return []provider.Result{
		{
			Target: "github:facebook/react",
			GitHub: &provider.GitHubMetrics{
				Stars:           215000,
				Forks:           45000,
				OpenIssues:      1000,
				Contributors:    1500,
				ReleaseCount:    200,
				LastCommitDays:  1,
				Commits30d:      120,
				IssuesClosed30d: 340,
			},
		},
		{
			Target: "npm:react",
			NPM: &provider.NPMMetrics{
				WeeklyDownloads:   25000000,
				MonthlyDownloads:  100000000,
				LatestVersion:     "19.1.0",
				LastPublishDays:   15,
				DependenciesCount: 2,
				License:           "MIT",
			},
		},
		{
			Target: "github:nonexistent/repo",
			Error:  "GitHub API: 404 Not Found",
		},
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	results := sampleResults()
	if err := JSON(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be valid JSON array.
	var parsed []json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(parsed) != 3 {
		t.Errorf("expected 3 items, got %d", len(parsed))
	}
	// Should end with newline.
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Error("JSON output should end with newline")
	}
}

func TestNDJSON(t *testing.T) {
	var buf bytes.Buffer
	results := sampleResults()
	if err := NDJSON(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	// Each line should be valid JSON.
	for i, line := range lines {
		var obj json.RawMessage
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestMarkdown(t *testing.T) {
	var buf bytes.Buffer
	results := sampleResults()
	if err := Markdown(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should contain table header.
	if !strings.Contains(output, "| target |") {
		t.Error("expected Markdown table header with '| target |'")
	}
	// Should contain separator line.
	if !strings.Contains(output, "|---") {
		t.Error("expected Markdown table separator")
	}
	// Should contain the repo name.
	if !strings.Contains(output, "facebook/react") {
		t.Error("expected facebook/react in output")
	}
	// Error row should show error.
	if !strings.Contains(output, "404 Not Found") {
		t.Error("expected error message in output")
	}
}

func TestMarkdownNPM(t *testing.T) {
	var buf bytes.Buffer
	results := []provider.Result{
		{
			Target: "npm:react",
			NPM: &provider.NPMMetrics{
				WeeklyDownloads:   25000000,
				MonthlyDownloads:  100000000,
				LatestVersion:     "19.1.0",
				LastPublishDays:   15,
				DependenciesCount: 2,
				License:           "MIT",
			},
		},
	}
	if err := Markdown(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "| weekly_downloads |") {
		t.Error("expected npm table header with '| weekly_downloads |'")
	}
	if !strings.Contains(output, "| monthly_downloads |") {
		t.Error("expected npm table header with '| monthly_downloads |'")
	}
	if !strings.Contains(output, "npm:react") {
		t.Error("expected npm:react in output")
	}
	if !strings.Contains(output, "25000000") {
		t.Error("expected weekly_downloads value in output")
	}
	if !strings.Contains(output, "100000000") {
		t.Error("expected monthly_downloads value in output")
	}
	if !strings.Contains(output, "19.1.0") {
		t.Error("expected latest_version in output")
	}
	if !strings.Contains(output, "MIT") {
		t.Error("expected license in output")
	}
}

func TestMarkdownPyPI(t *testing.T) {
	var buf bytes.Buffer
	results := []provider.Result{
		{
			Target: "pypi:requests",
			PyPI: &provider.PyPIMetrics{
				WeeklyDownloads:   5000000,
				MonthlyDownloads:  20000000,
				LatestVersion:     "2.31.0",
				LastPublishDays:   30,
				DependenciesCount: 5,
				License:           "Apache-2.0",
				RequiresPython:    ">=3.7",
			},
		},
	}
	if err := Markdown(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"| target |",
		"| weekly_downloads |",
		"| monthly_downloads |",
		"| latest_version |",
		"| last_publish_days |",
		"| dependencies_count |",
		"| license |",
		"| requires_python |",
		"| error |",
		"pypi:requests",
		"5000000",
		"20000000",
		"2.31.0",
		"30",
		"Apache-2.0",
		">=3.7",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in output, got:\n%s", want, output)
		}
	}
}

func TestMarkdownCrates(t *testing.T) {
	var buf bytes.Buffer
	results := []provider.Result{
		{
			Target: "crates:serde",
			Crates: &provider.CratesMetrics{
				Downloads:           80000000,
				RecentDownloads:     3000000,
				LatestVersion:       "1.0.197",
				LastPublishDays:     7,
				DependenciesCount:   2,
				License:             "MIT OR Apache-2.0",
				ReverseDependencies: 30000,
			},
		},
	}
	if err := Markdown(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"| target |",
		"| downloads |",
		"| recent_downloads |",
		"| latest_version |",
		"| last_publish_days |",
		"| dependencies_count |",
		"| license |",
		"| reverse_dependencies |",
		"| error |",
		"crates:serde",
		"80000000",
		"3000000",
		"1.0.197",
		"7",
		"MIT OR Apache-2.0",
		"30000",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in output, got:\n%s", want, output)
		}
	}
}

func TestMarkdownGo(t *testing.T) {
	var buf bytes.Buffer
	results := []provider.Result{
		{
			Target: "go:github.com/gin-gonic/gin",
			Go: &provider.GoMetrics{
				LatestVersion:     "v1.9.1",
				LastPublishDays:   60,
				DependenciesCount: 10,
				License:           "MIT",
			},
		},
	}
	if err := Markdown(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{
		"| target |",
		"| latest_version |",
		"| last_publish_days |",
		"| dependencies_count |",
		"| license |",
		"| error |",
		"go:github.com/gin-gonic/gin",
		"v1.9.1",
		"60",
		"10",
		"MIT",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in output, got:\n%s", want, output)
		}
	}
}

func TestMarkdownMixed(t *testing.T) {
	var buf bytes.Buffer
	results := []provider.Result{
		{
			Target: "github:facebook/react",
			GitHub: &provider.GitHubMetrics{
				Stars:           215000,
				Forks:           45000,
				OpenIssues:      1000,
				Contributors:    1500,
				ReleaseCount:    200,
				LastCommitDays:  1,
				Commits30d:      120,
				IssuesClosed30d: 340,
			},
		},
		{
			Target: "npm:react",
			NPM: &provider.NPMMetrics{
				WeeklyDownloads:   25000000,
				MonthlyDownloads:  100000000,
				LatestVersion:     "19.1.0",
				LastPublishDays:   15,
				DependenciesCount: 2,
				License:           "MIT",
			},
		},
		{
			Target: "pypi:requests",
			PyPI: &provider.PyPIMetrics{
				WeeklyDownloads:   5000000,
				MonthlyDownloads:  20000000,
				LatestVersion:     "2.31.0",
				LastPublishDays:   30,
				DependenciesCount: 5,
				License:           "Apache-2.0",
				RequiresPython:    ">=3.7",
			},
		},
		{
			Target: "crates:serde",
			Crates: &provider.CratesMetrics{
				Downloads:           80000000,
				RecentDownloads:     3000000,
				LatestVersion:       "1.0.197",
				LastPublishDays:     7,
				DependenciesCount:   2,
				License:             "MIT OR Apache-2.0",
				ReverseDependencies: 30000,
			},
		},
		{
			Target: "go:github.com/gin-gonic/gin",
			Go: &provider.GoMetrics{
				LatestVersion:     "v1.9.1",
				LastPublishDays:   60,
				DependenciesCount: 10,
				License:           "MIT",
			},
		},
		{
			Target: "github:nonexistent/repo",
			Error:  "GitHub API: 404 Not Found",
		},
	}
	if err := Markdown(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// All 5 provider tables should be present
	for _, want := range []string{
		"| stars |",                  // GitHub header
		"| weekly_downloads |",      // npm header (also appears in PyPI, but both should exist)
		"| monthly_downloads |",     // PyPI-specific header
		"| recent_downloads |",      // crates-specific header
		"| reverse_dependencies |",  // crates-specific header
		"| requires_python |",       // PyPI-specific header
		"facebook/react",            // GitHub data
		"npm:react",                 // npm data
		"pypi:requests",             // PyPI data
		"crates:serde",              // crates data
		"go:github.com/gin-gonic/gin", // Go data
		"404 Not Found",             // error data
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in output, got:\n%s", want, output)
		}
	}

	// Tables should be separated by blank lines
	tables := strings.Split(output, "\n\n")
	if len(tables) < 5 {
		t.Errorf("expected at least 5 tables separated by blank lines, got %d sections", len(tables))
	}
}

func TestJSONEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Errorf("expected empty array, got %q", buf.String())
	}
}

func TestNDJSONEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := NDJSON(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}
