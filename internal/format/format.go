package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

// JSON writes results as a JSON array.
func JSON(w io.Writer, results []provider.Result) error {
	if results == nil {
		results = []provider.Result{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

// NDJSON writes results as newline-delimited JSON (one object per line).
func NDJSON(w io.Writer, results []provider.Result) error {
	enc := json.NewEncoder(w)
	for _, r := range results {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}

// Markdown writes results as a Markdown table grouped by scheme.
func Markdown(w io.Writer, results []provider.Result) error {
	var ghResults []provider.Result
	var errResults []provider.Result
	for _, r := range results {
		if r.GitHub != nil {
			ghResults = append(ghResults, r)
		} else {
			errResults = append(errResults, r)
		}
	}

	if len(ghResults) > 0 {
		if _, err := fmt.Fprintln(w, "| target | stars | forks | open_issues | contributors | release_count | last_commit_days | commits_30d | issues_closed_30d | error |"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "|---|---|---|---|---|---|---|---|---|---|"); err != nil {
			return err
		}
		for _, r := range ghResults {
			g := r.GitHub
			if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
				r.Target,
				strconv.Itoa(g.Stars),
				strconv.Itoa(g.Forks),
				strconv.Itoa(g.OpenIssues),
				strconv.Itoa(g.Contributors),
				strconv.Itoa(g.ReleaseCount),
				strconv.Itoa(g.LastCommitDays),
				strconv.Itoa(g.Commits30d),
				strconv.Itoa(g.IssuesClosed30d),
				r.Error,
			); err != nil {
				return err
			}
		}
	}

	if len(errResults) > 0 {
		if len(ghResults) > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "| target | error |"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "|---|---|"); err != nil {
			return err
		}
		for _, r := range errResults {
			if _, err := fmt.Fprintf(w, "| %s | %s |\n", r.Target, r.Error); err != nil {
				return err
			}
		}
	}

	return nil
}
