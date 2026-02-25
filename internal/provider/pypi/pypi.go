package pypi

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

var validPkgRe = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9._-]*[a-zA-Z0-9])?$`)

const (
	defaultPyPIURL  = "https://pypi.org"
	defaultStatsURL = "https://pypistats.org"
)

// Provider fetches metrics from the PyPI registry and pypistats.org.
type Provider struct {
	pypiURL  string
	statsURL string
	client   *http.Client
}

// New creates a PyPI provider. Pass empty strings for default URLs.
func New(pypiURL, statsURL string) *Provider {
	if pypiURL == "" {
		pypiURL = defaultPyPIURL
	}
	if statsURL == "" {
		statsURL = defaultStatsURL
	}
	return &Provider{
		pypiURL:  strings.TrimRight(pypiURL, "/"),
		statsURL: strings.TrimRight(statsURL, "/"),
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Provider) Scheme() string { return "pypi" }

func (p *Provider) Fetch(ctx context.Context, identifier string) (provider.Result, error) {
	if identifier == "" || !validPkgRe.MatchString(identifier) {
		return provider.Result{
			Target: "pypi:" + identifier,
			Error:  fmt.Sprintf("invalid PyPI package name %q", identifier),
		}, nil
	}

	metrics := &provider.PyPIMetrics{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errs []string

	type job struct {
		name string
		fn   func(context.Context) error
	}

	jobs := []job{
		{"metadata", func(ctx context.Context) error {
			meta, err := p.fetchMetadata(ctx, identifier)
			if err != nil {
				return err
			}
			mu.Lock()
			metrics.LatestVersion = meta.Info.Version
			metrics.License = meta.Info.License
			metrics.RequiresPython = meta.Info.RequiresPython
			metrics.DependenciesCount = countNonExtraDeps(meta.Info.RequiresDist)
			metrics.LastPublishDays = meta.lastPublishDays()
			mu.Unlock()
			return nil
		}},
		{"downloads", func(ctx context.Context) error {
			dl, err := p.fetchDownloads(ctx, identifier)
			if err != nil {
				return err
			}
			mu.Lock()
			metrics.WeeklyDownloads = dl.Data.LastWeek
			metrics.MonthlyDownloads = dl.Data.LastMonth
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
		Target: "pypi:" + identifier,
	}

	if len(errs) == len(jobs) {
		result.Error = strings.Join(errs, "; ")
		return result, nil
	}

	result.PyPI = metrics
	if len(errs) > 0 {
		result.Error = strings.Join(errs, "; ")
	}
	return result, nil
}

// pypiResponse represents the PyPI JSON API response.
type pypiResponse struct {
	Info     pypiInfo                       `json:"info"`
	Releases map[string][]pypiReleaseFile   `json:"releases"`
}

type pypiInfo struct {
	Version        string   `json:"version"`
	License        string   `json:"license"`
	RequiresPython string   `json:"requires_python"`
	RequiresDist   []string `json:"requires_dist"`
}

type pypiReleaseFile struct {
	UploadTimeISO string `json:"upload_time_iso_8601"`
}

func (r *pypiResponse) lastPublishDays() int {
	files, ok := r.Releases[r.Info.Version]
	if !ok || len(files) == 0 {
		return 0
	}
	t, err := time.Parse(time.RFC3339, files[0].UploadTimeISO)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z", files[0].UploadTimeISO)
		if err != nil {
			return 0
		}
	}
	days := int(math.Floor(time.Since(t).Hours() / 24))
	if days < 0 {
		days = 0
	}
	return days
}

func (p *Provider) fetchMetadata(ctx context.Context, pkg string) (*pypiResponse, error) {
	u := fmt.Sprintf("%s/pypi/%s/json", p.pypiURL, pkg)
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
		return nil, fmt.Errorf("PyPI API: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var data pypiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &data, nil
}

// statsResponse represents the pypistats.org recent downloads response.
type statsResponse struct {
	Data statsData `json:"data"`
}

type statsData struct {
	LastDay   int `json:"last_day"`
	LastWeek  int `json:"last_week"`
	LastMonth int `json:"last_month"`
}

func (p *Provider) fetchDownloads(ctx context.Context, pkg string) (*statsResponse, error) {
	u := fmt.Sprintf("%s/api/packages/%s/recent", p.statsURL, pkg)
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
		return nil, fmt.Errorf("pypistats downloads API: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var data statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &data, nil
}

// countNonExtraDeps counts requires_dist entries that are NOT extras
// (i.e., entries that do NOT contain "; extra ==").
func countNonExtraDeps(deps []string) int {
	count := 0
	for _, d := range deps {
		if !strings.Contains(d, "; extra ==") {
			count++
		}
	}
	return count
}
