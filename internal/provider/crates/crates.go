package crates

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

var validCrateRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

const defaultBaseURL = "https://crates.io"

// userAgentTransport adds a User-Agent header to every outgoing request.
type userAgentTransport struct {
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", "repiq (https://github.com/yutakobayashidev/repiq)")
	return t.base.RoundTrip(req)
}

// Provider fetches metrics from the crates.io API.
type Provider struct {
	baseURL string
	client  *http.Client
}

// New creates a crates.io provider. Pass empty string for default base URL.
func New(baseURL string) *Provider {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Provider{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: &userAgentTransport{base: http.DefaultTransport},
		},
	}
}

func (p *Provider) Scheme() string { return "crates" }

func (p *Provider) Fetch(ctx context.Context, identifier string) (provider.Result, error) {
	target := "crates:" + identifier

	if identifier == "" || !validCrateRe.MatchString(identifier) {
		return provider.Result{
			Target: target,
			Error:  fmt.Sprintf("invalid crate name %q", identifier),
		}, nil
	}

	// Phase 1: fetch crate metadata
	meta, err := p.fetchMetadata(ctx, identifier)
	if err != nil {
		return provider.Result{
			Target: target,
			Error:  fmt.Sprintf("crates.io API: %s", err.Error()),
		}, nil
	}

	version := meta.maxStableVersion
	if version == "" {
		version = meta.newestVersion
	}

	metrics := &provider.CratesMetrics{
		Downloads:       meta.downloads,
		RecentDownloads: meta.recentDownloads,
		LatestVersion:   version,
	}

	// Find matching version for license and created_at
	for _, v := range meta.versions {
		if v.num == version {
			metrics.License = v.license
			t, err := time.Parse(time.RFC3339, v.createdAt)
			if err == nil {
				days := int(math.Floor(time.Since(t).Hours() / 24))
				if days < 0 {
					days = 0
				}
				metrics.LastPublishDays = days
			}
			break
		}
	}

	// Phase 2: parallel fetch deps + reverse deps
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errs []string

	wg.Add(2)

	go func() {
		defer wg.Done()
		count, err := p.fetchDependenciesCount(ctx, identifier, version)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("dependencies: %s", err.Error()))
			mu.Unlock()
			return
		}
		mu.Lock()
		metrics.DependenciesCount = count
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		count, err := p.fetchReverseDependenciesCount(ctx, identifier)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("reverse_dependencies: %s", err.Error()))
			mu.Unlock()
			return
		}
		mu.Lock()
		metrics.ReverseDependencies = count
		mu.Unlock()
	}()

	wg.Wait()

	result := provider.Result{
		Target: target,
		Crates: metrics,
	}
	if len(errs) > 0 {
		result.Error = strings.Join(errs, "; ")
	}
	return result, nil
}

// --- internal types for parsed metadata ---

type crateMetadata struct {
	downloads        int
	recentDownloads  int
	maxStableVersion string
	newestVersion    string
	versions         []versionEntry
}

type versionEntry struct {
	num       string
	license   string
	createdAt string
}

// --- HTTP helpers ---

func (p *Provider) fetchMetadata(ctx context.Context, crate string) (*crateMetadata, error) {
	u := fmt.Sprintf("%s/api/v1/crates/%s", p.baseURL, crate)
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
		return nil, fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var raw struct {
		Crate struct {
			Downloads        int     `json:"downloads"`
			RecentDownloads  int     `json:"recent_downloads"`
			MaxStableVersion *string `json:"max_stable_version"`
			NewestVersion    string  `json:"newest_version"`
		} `json:"crate"`
		Versions []struct {
			Num       string `json:"num"`
			License   string `json:"license"`
			CreatedAt string `json:"created_at"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	meta := &crateMetadata{
		downloads:       raw.Crate.Downloads,
		recentDownloads: raw.Crate.RecentDownloads,
		newestVersion:   raw.Crate.NewestVersion,
	}
	if raw.Crate.MaxStableVersion != nil {
		meta.maxStableVersion = *raw.Crate.MaxStableVersion
	}
	for _, v := range raw.Versions {
		meta.versions = append(meta.versions, versionEntry{
			num:       v.Num,
			license:   v.License,
			createdAt: v.CreatedAt,
		})
	}
	return meta, nil
}

func (p *Provider) fetchDependenciesCount(ctx context.Context, crate, version string) (int, error) {
	u := fmt.Sprintf("%s/api/v1/crates/%s/%s/dependencies", p.baseURL, crate, version)
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
		return 0, fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var deps struct {
		Dependencies []struct {
			CrateID  string `json:"crate_id"`
			Kind     string `json:"kind"`
			Optional bool   `json:"optional"`
		} `json:"dependencies"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&deps); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}

	count := 0
	for _, d := range deps.Dependencies {
		if d.Kind == "normal" {
			count++
		}
	}
	return count, nil
}

func (p *Provider) fetchReverseDependenciesCount(ctx context.Context, crate string) (int, error) {
	u := fmt.Sprintf("%s/api/v1/crates/%s/reverse_dependencies?per_page=1", p.baseURL, crate)
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
		return 0, fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var revDeps struct {
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&revDeps); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}
	return revDeps.Meta.Total, nil
}
