package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

var validPkgRe = regexp.MustCompile(`^(@[a-zA-Z0-9][\w.-]*/)?[a-zA-Z0-9][\w.-]*$`)

const (
	defaultRegistryURL  = "https://registry.npmjs.org"
	defaultDownloadsURL = "https://api.npmjs.org"
)

// Provider fetches metrics from the npm registry.
type Provider struct {
	registryURL  string
	downloadsURL string
	client       *http.Client
}

// New creates an npm provider. Pass empty strings for default URLs.
func New(registryURL, downloadsURL string) *Provider {
	if registryURL == "" {
		registryURL = defaultRegistryURL
	}
	if downloadsURL == "" {
		downloadsURL = defaultDownloadsURL
	}
	return &Provider{
		registryURL:  strings.TrimRight(registryURL, "/"),
		downloadsURL: strings.TrimRight(downloadsURL, "/"),
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Provider) Scheme() string { return "npm" }

func (p *Provider) Fetch(ctx context.Context, identifier string) (provider.Result, error) {
	if identifier == "" || !validPkgRe.MatchString(identifier) {
		return provider.Result{
			Target: "npm:" + identifier,
			Error:  fmt.Sprintf("invalid npm package name %q", identifier),
		}, nil
	}

	metrics := &provider.NPMMetrics{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errs []string

	type job struct {
		name string
		fn   func(context.Context) error
	}

	jobs := []job{
		{"latest", func(ctx context.Context) error {
			latest, err := p.fetchLatest(ctx, identifier)
			if err != nil {
				return err
			}
			mu.Lock()
			metrics.LatestVersion = latest.Version
			metrics.DependenciesCount = len(latest.Dependencies)
			metrics.License = latest.License
			mu.Unlock()
			return nil
		}},
		{"modified", func(ctx context.Context) error {
			days, err := p.fetchLastPublishDays(ctx, identifier)
			if err != nil {
				return err
			}
			mu.Lock()
			metrics.LastPublishDays = days
			mu.Unlock()
			return nil
		}},
		{"downloads", func(ctx context.Context) error {
			count, err := p.fetchWeeklyDownloads(ctx, identifier)
			if err != nil {
				return err
			}
			mu.Lock()
			metrics.WeeklyDownloads = count
			mu.Unlock()
			return nil
		}},
		{"monthly_downloads", func(ctx context.Context) error {
			count, err := p.fetchMonthlyDownloads(ctx, identifier)
			if err != nil {
				return err
			}
			mu.Lock()
			metrics.MonthlyDownloads = count
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
				errs = append(errs, fmt.Sprintf("%s: %s", j.name, err.Error()))
				mu.Unlock()
			}
		}(j)
	}
	wg.Wait()

	result := provider.Result{
		Target: "npm:" + identifier,
	}

	if len(errs) == len(jobs) {
		result.Error = strings.Join(errs, "; ")
		return result, nil
	}

	result.NPM = metrics
	if len(errs) > 0 {
		result.Error = strings.Join(errs, "; ")
	}
	return result, nil
}

type latestResponse struct {
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
	License      string
	RawLicense   json.RawMessage `json:"license"`
}

func (p *Provider) fetchLatest(ctx context.Context, pkg string) (*latestResponse, error) {
	u := fmt.Sprintf("%s/%s/latest", p.registryURL, pkg)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("npm registry: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var latest latestResponse
	if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	latest.License = parseLicense(latest.RawLicense)
	return &latest, nil
}

func parseLicense(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	var obj struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj.Type
	}
	return ""
}

func (p *Provider) fetchLastPublishDays(ctx context.Context, pkg string) (int, error) {
	u := fmt.Sprintf("%s/%s", p.registryURL, pkg)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.npm.install-v1+json")

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("npm registry: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var meta struct {
		Modified string `json:"modified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}

	t, err := time.Parse(time.RFC3339Nano, meta.Modified)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05.000Z", meta.Modified)
		if err != nil {
			return 0, fmt.Errorf("parsing modified date %q: %w", meta.Modified, err)
		}
	}

	days := int(math.Floor(time.Since(t).Hours() / 24))
	if days < 0 {
		days = 0
	}
	return days, nil
}

func (p *Provider) fetchDownloads(ctx context.Context, pkg, period string) (int, error) {
	// For scoped packages (@scope/name), the npm downloads API requires
	// the package name as a single path token with the slash encoded as %2F
	// (e.g., @scope%2Fname). We use url.PathEscape on the full name to
	// achieve this, since PathEscape encodes "/" to %2F.
	encodedPkg := url.PathEscape(pkg)

	u := fmt.Sprintf("%s/downloads/point/%s/%s", p.downloadsURL, period, encodedPkg)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return 0, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("npm downloads API: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var dl struct {
		Downloads int `json:"downloads"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dl); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}
	return dl.Downloads, nil
}

func (p *Provider) fetchWeeklyDownloads(ctx context.Context, pkg string) (int, error) {
	return p.fetchDownloads(ctx, pkg, "last-week")
}

func (p *Provider) fetchMonthlyDownloads(ctx context.Context, pkg string) (int, error) {
	return p.fetchDownloads(ctx, pkg, "last-month")
}
