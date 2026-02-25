package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

func escapeMarkdown(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

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
	var npmResults []provider.Result
	var pypiResults []provider.Result
	var cratesResults []provider.Result
	var goResults []provider.Result
	var errResults []provider.Result
	for _, r := range results {
		switch {
		case r.GitHub != nil:
			ghResults = append(ghResults, r)
		case r.NPM != nil:
			npmResults = append(npmResults, r)
		case r.PyPI != nil:
			pypiResults = append(pypiResults, r)
		case r.Crates != nil:
			cratesResults = append(cratesResults, r)
		case r.Go != nil:
			goResults = append(goResults, r)
		default:
			errResults = append(errResults, r)
		}
	}

	needSep := false

	if len(ghResults) > 0 {
		if _, err := fmt.Fprintln(w, "| target | stars | forks | open_issues | contributors | release_count | last_commit_days | commits_30d | issues_closed_30d | license | error |"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "|---|---|---|---|---|---|---|---|---|---|---|"); err != nil {
			return err
		}
		for _, r := range ghResults {
			g := r.GitHub
			if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
				escapeMarkdown(r.Target),
				strconv.Itoa(g.Stars),
				strconv.Itoa(g.Forks),
				strconv.Itoa(g.OpenIssues),
				strconv.Itoa(g.Contributors),
				strconv.Itoa(g.ReleaseCount),
				strconv.Itoa(g.LastCommitDays),
				strconv.Itoa(g.Commits30d),
				strconv.Itoa(g.IssuesClosed30d),
				escapeMarkdown(g.License),
				escapeMarkdown(r.Error),
			); err != nil {
				return err
			}
		}
		needSep = true
	}

	if len(npmResults) > 0 {
		if needSep {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "| target | weekly_downloads | monthly_downloads | latest_version | last_publish_days | dependencies_count | license | error |"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "|---|---|---|---|---|---|---|---|"); err != nil {
			return err
		}
		for _, r := range npmResults {
			n := r.NPM
			if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s |\n",
				escapeMarkdown(r.Target),
				strconv.Itoa(n.WeeklyDownloads),
				strconv.Itoa(n.MonthlyDownloads),
				escapeMarkdown(n.LatestVersion),
				strconv.Itoa(n.LastPublishDays),
				strconv.Itoa(n.DependenciesCount),
				escapeMarkdown(n.License),
				escapeMarkdown(r.Error),
			); err != nil {
				return err
			}
		}
		needSep = true
	}

	if len(pypiResults) > 0 {
		if needSep {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "| target | weekly_downloads | monthly_downloads | latest_version | last_publish_days | dependencies_count | license | requires_python | error |"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "|---|---|---|---|---|---|---|---|---|"); err != nil {
			return err
		}
		for _, r := range pypiResults {
			p := r.PyPI
			if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
				escapeMarkdown(r.Target),
				strconv.Itoa(p.WeeklyDownloads),
				strconv.Itoa(p.MonthlyDownloads),
				escapeMarkdown(p.LatestVersion),
				strconv.Itoa(p.LastPublishDays),
				strconv.Itoa(p.DependenciesCount),
				escapeMarkdown(p.License),
				escapeMarkdown(p.RequiresPython),
				escapeMarkdown(r.Error),
			); err != nil {
				return err
			}
		}
		needSep = true
	}

	if len(cratesResults) > 0 {
		if needSep {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "| target | downloads | recent_downloads | latest_version | last_publish_days | dependencies_count | license | reverse_dependencies | error |"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "|---|---|---|---|---|---|---|---|---|"); err != nil {
			return err
		}
		for _, r := range cratesResults {
			c := r.Crates
			if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
				escapeMarkdown(r.Target),
				strconv.Itoa(c.Downloads),
				strconv.Itoa(c.RecentDownloads),
				escapeMarkdown(c.LatestVersion),
				strconv.Itoa(c.LastPublishDays),
				strconv.Itoa(c.DependenciesCount),
				escapeMarkdown(c.License),
				strconv.Itoa(c.ReverseDependencies),
				escapeMarkdown(r.Error),
			); err != nil {
				return err
			}
		}
		needSep = true
	}

	if len(goResults) > 0 {
		if needSep {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "| target | latest_version | last_publish_days | dependencies_count | license | error |"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "|---|---|---|---|---|---|"); err != nil {
			return err
		}
		for _, r := range goResults {
			g := r.Go
			if _, err := fmt.Fprintf(w, "| %s | %s | %s | %s | %s | %s |\n",
				escapeMarkdown(r.Target),
				escapeMarkdown(g.LatestVersion),
				strconv.Itoa(g.LastPublishDays),
				strconv.Itoa(g.DependenciesCount),
				escapeMarkdown(g.License),
				escapeMarkdown(r.Error),
			); err != nil {
				return err
			}
		}
		needSep = true
	}

	if len(errResults) > 0 {
		if needSep {
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
			if _, err := fmt.Fprintf(w, "| %s | %s |\n", escapeMarkdown(r.Target), escapeMarkdown(r.Error)); err != nil {
				return err
			}
		}
	}

	return nil
}
