package provider

import "context"

// Provider fetches metrics for a given identifier.
type Provider interface {
	Scheme() string
	Fetch(ctx context.Context, identifier string) (Result, error)
}

// Result holds the output for a single target.
type Result struct {
	Target string         `json:"target"`
	GitHub *GitHubMetrics `json:"github,omitempty"`
	Error  string         `json:"error,omitempty"`
}

// GitHubMetrics holds GitHub-specific metrics.
type GitHubMetrics struct {
	Stars            int `json:"stars"`
	Forks            int `json:"forks"`
	OpenIssues       int `json:"open_issues"`
	Contributors     int `json:"contributors"`
	ReleaseCount     int `json:"release_count"`
	LastCommitDays   int `json:"last_commit_days"`
	Commits30d       int `json:"commits_30d"`
	IssuesClosed30d  int `json:"issues_closed_30d"`
}
