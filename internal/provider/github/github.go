package github

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	gogithub "github.com/google/go-github/v68/github"
	"github.com/yutakobayashidev/repiq/internal/provider"
)

var validNameRe = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Provider fetches metrics from the GitHub API.
type Provider struct {
	client *gogithub.Client
}

// New creates a GitHub provider. If token is non-empty, authenticated requests are used.
// baseURL overrides the API base URL (useful for testing); pass "" for default.
func New(token, baseURL string) *Provider {
	var httpClient *http.Client
	if token != "" {
		httpClient = &http.Client{
			Transport: &tokenTransport{token: token, base: http.DefaultTransport},
		}
	}
	client := gogithub.NewClient(httpClient)
	if baseURL != "" {
		client.BaseURL = mustParseURL(baseURL)
	}
	return &Provider{client: client}
}

func (p *Provider) Scheme() string { return "github" }

func (p *Provider) Fetch(ctx context.Context, identifier string) (provider.Result, error) {
	owner, repo, err := parseIdentifier(identifier)
	if err != nil {
		return provider.Result{Target: "github:" + identifier, Error: err.Error()}, nil
	}

	repoInfo, resp, err := p.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return provider.Result{
			Target: "github:" + identifier,
			Error:  fmt.Sprintf("GitHub API: %s", errorMessage(resp, err)),
		}, nil
	}

	var license string
	if lic := repoInfo.GetLicense(); lic != nil {
		license = lic.GetSPDXID()
	}

	metrics := &provider.GitHubMetrics{
		Stars:      repoInfo.GetStargazersCount(),
		Forks:      repoInfo.GetForksCount(),
		OpenIssues: repoInfo.GetOpenIssuesCount(),
		License:    license,
	}

	errs := p.fetchParallel(ctx, owner, repo, metrics)

	result := provider.Result{
		Target: "github:" + identifier,
		GitHub: metrics,
	}
	if len(errs) > 0 {
		result.Error = strings.Join(errs, "; ")
	}
	return result, nil
}

func parseIdentifier(identifier string) (owner, repo string, err error) {
	parts := strings.SplitN(identifier, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid identifier %q: expected owner/repo", identifier)
	}
	if !validNameRe.MatchString(parts[0]) || !validNameRe.MatchString(parts[1]) {
		return "", "", fmt.Errorf("invalid identifier %q: owner and repo must match [a-zA-Z0-9._-]+", identifier)
	}
	return parts[0], parts[1], nil
}

func (p *Provider) fetchParallel(ctx context.Context, owner, repo string, m *provider.GitHubMetrics) []string {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errs []string

	type job struct {
		fn func(context.Context) error
	}

	jobs := []job{
		{func(ctx context.Context) error {
			count, err := p.fetchContributorCount(ctx, owner, repo)
			if err != nil {
				return fmt.Errorf("contributors: %w", err)
			}
			mu.Lock()
			m.Contributors = count
			mu.Unlock()
			return nil
		}},
		{func(ctx context.Context) error {
			count, err := p.fetchReleaseCount(ctx, owner, repo)
			if err != nil {
				return fmt.Errorf("releases: %w", err)
			}
			mu.Lock()
			m.ReleaseCount = count
			mu.Unlock()
			return nil
		}},
		{func(ctx context.Context) error {
			days, err := p.fetchLastCommitDays(ctx, owner, repo)
			if err != nil {
				return fmt.Errorf("commits: %w", err)
			}
			mu.Lock()
			m.LastCommitDays = days
			mu.Unlock()
			return nil
		}},
		{func(ctx context.Context) error {
			count, err := p.fetchCommits30d(ctx, owner, repo)
			if err != nil {
				return fmt.Errorf("search commits: %w", err)
			}
			mu.Lock()
			m.Commits30d = count
			mu.Unlock()
			return nil
		}},
		{func(ctx context.Context) error {
			count, err := p.fetchIssuesClosed30d(ctx, owner, repo)
			if err != nil {
				return fmt.Errorf("search issues: %w", err)
			}
			mu.Lock()
			m.IssuesClosed30d = count
			mu.Unlock()
			return nil
		}},
	}

	wg.Add(len(jobs))
	for _, j := range jobs {
		go func(j job) {
			defer wg.Done()
			if err := j.fn(ctx); err != nil {
				mu.Lock()
				errs = append(errs, err.Error())
				mu.Unlock()
			}
		}(j)
	}
	wg.Wait()
	return errs
}

func (p *Provider) fetchContributorCount(ctx context.Context, owner, repo string) (int, error) {
	return p.countViaLinkHeader(ctx, owner, repo, "contributors")
}

func (p *Provider) fetchReleaseCount(ctx context.Context, owner, repo string) (int, error) {
	return p.countViaLinkHeader(ctx, owner, repo, "releases")
}

func (p *Provider) fetchLastCommitDays(ctx context.Context, owner, repo string) (int, error) {
	commits, _, err := p.client.Repositories.ListCommits(ctx, owner, repo, &gogithub.CommitsListOptions{
		ListOptions: gogithub.ListOptions{PerPage: 1},
	})
	if err != nil {
		return 0, err
	}
	if len(commits) == 0 {
		return 0, nil
	}
	date := commits[0].GetCommit().GetCommitter().GetDate()
	days := int(math.Floor(time.Since(date.Time).Hours() / 24))
	if days < 0 {
		days = 0
	}
	return days, nil
}

func (p *Provider) fetchCommits30d(ctx context.Context, owner, repo string) (int, error) {
	since := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	query := fmt.Sprintf("repo:%s/%s committer-date:>%s", owner, repo, since)
	result, _, err := p.client.Search.Commits(ctx, query, &gogithub.SearchOptions{
		ListOptions: gogithub.ListOptions{PerPage: 1},
	})
	if err != nil {
		return 0, err
	}
	return result.GetTotal(), nil
}

func (p *Provider) fetchIssuesClosed30d(ctx context.Context, owner, repo string) (int, error) {
	since := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	query := fmt.Sprintf("repo:%s/%s is:issue is:closed closed:>%s", owner, repo, since)
	result, _, err := p.client.Search.Issues(ctx, query, &gogithub.SearchOptions{
		ListOptions: gogithub.ListOptions{PerPage: 1},
	})
	if err != nil {
		return 0, err
	}
	return result.GetTotal(), nil
}

// countViaLinkHeader estimates count by requesting per_page=1 and reading LastPage from
// the Link header. If no pagination exists (small repos), it re-fetches with per_page=100
// and counts items directly. Note: the fallback is capped at 100 items; repos with
// >100 contributors/releases without pagination headers will report up to 100.
func (p *Provider) countViaLinkHeader(ctx context.Context, owner, repo, resource string) (int, error) {
	u := fmt.Sprintf("repos/%s/%s/%s", owner, repo, resource)
	req, err := p.client.NewRequest("GET", u, nil)
	if err != nil {
		return 0, err
	}
	q := req.URL.Query()
	q.Set("per_page", "1")
	req.URL.RawQuery = q.Encode()

	resp, err := p.client.Do(ctx, req, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.LastPage > 0 {
		return resp.LastPage, nil
	}

	switch resource {
	case "contributors":
		items, _, err := p.client.Repositories.ListContributors(ctx, owner, repo, &gogithub.ListContributorsOptions{
			ListOptions: gogithub.ListOptions{PerPage: 100},
		})
		if err != nil {
			return 0, err
		}
		return len(items), nil
	case "releases":
		items, _, err := p.client.Repositories.ListReleases(ctx, owner, repo, &gogithub.ListOptions{PerPage: 100})
		if err != nil {
			return 0, err
		}
		return len(items), nil
	}
	return 0, nil
}

func errorMessage(resp *gogithub.Response, err error) string {
	if resp != nil {
		return fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	return err.Error()
}

type tokenTransport struct {
	token string
	base  http.RoundTripper
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("invalid base URL %q: %v", rawURL, err))
	}
	return u
}
