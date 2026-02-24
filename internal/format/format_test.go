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
	if !strings.Contains(output, "npm:react") {
		t.Error("expected npm:react in output")
	}
	if !strings.Contains(output, "25000000") {
		t.Error("expected weekly_downloads value in output")
	}
	if !strings.Contains(output, "19.1.0") {
		t.Error("expected latest_version in output")
	}
	if !strings.Contains(output, "MIT") {
		t.Error("expected license in output")
	}
}

func TestMarkdownMixed(t *testing.T) {
	var buf bytes.Buffer
	results := sampleResults()
	if err := Markdown(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	// Should have both GitHub and npm tables
	if !strings.Contains(output, "| stars |") {
		t.Error("expected GitHub table header")
	}
	if !strings.Contains(output, "| weekly_downloads |") {
		t.Error("expected npm table header")
	}
	if !strings.Contains(output, "facebook/react") {
		t.Error("expected GitHub result in output")
	}
	if !strings.Contains(output, "npm:react") {
		t.Error("expected npm result in output")
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
