package golang

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
	"unicode"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

var validModuleRe = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9.]*\.[a-zA-Z]{2,}/.+$`)

const (
	defaultProxyURL   = "https://proxy.golang.org"
	defaultDepsdevURL = "https://api.deps.dev"
)

// Provider fetches metrics from the Go Module Proxy and deps.dev.
type Provider struct {
	proxyURL   string
	depsdevURL string
	client     *http.Client
}

// New creates a Go modules provider. Pass empty strings for default URLs.
func New(proxyURL, depsdevURL string) *Provider {
	if proxyURL == "" {
		proxyURL = defaultProxyURL
	}
	if depsdevURL == "" {
		depsdevURL = defaultDepsdevURL
	}
	return &Provider{
		proxyURL:   strings.TrimRight(proxyURL, "/"),
		depsdevURL: strings.TrimRight(depsdevURL, "/"),
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Provider) Scheme() string { return "go" }

func (p *Provider) Fetch(ctx context.Context, identifier string) (provider.Result, error) {
	if identifier == "" || !validModuleRe.MatchString(identifier) {
		return provider.Result{
			Target: "go:" + identifier,
			Error:  fmt.Sprintf("invalid Go module path %q", identifier),
		}, nil
	}

	// Phase 1: Fetch latest version from Go Module Proxy.
	proxyInfo, err := p.fetchLatest(ctx, identifier)
	if err != nil {
		return provider.Result{
			Target: "go:" + identifier,
			Error:  fmt.Sprintf("proxy: %s", err.Error()),
		}, nil
	}

	days := int(math.Floor(time.Since(proxyInfo.Time).Hours() / 24))
	if days < 0 {
		days = 0
	}

	metrics := &provider.GoMetrics{
		LatestVersion:   proxyInfo.Version,
		LastPublishDays: days,
	}

	// Phase 2: Fetch license and dependencies from deps.dev (parallel).
	var mu sync.Mutex
	var wg sync.WaitGroup
	var errs []string

	wg.Add(2)

	go func() {
		defer wg.Done()
		license, err := p.fetchLicense(ctx, identifier, proxyInfo.Version)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Sprintf("license: %s", err.Error()))
			mu.Unlock()
			return
		}
		mu.Lock()
		metrics.License = license
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		count, err := p.fetchDependenciesCount(ctx, identifier, proxyInfo.Version)
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

	wg.Wait()

	result := provider.Result{
		Target: "go:" + identifier,
		Go:     metrics,
	}
	if len(errs) > 0 {
		result.Error = strings.Join(errs, "; ")
	}
	return result, nil
}

// proxyResponse represents the JSON response from the Go Module Proxy /@latest endpoint.
type proxyResponse struct {
	Version string    `json:"Version"`
	Time    time.Time `json:"Time"`
}

func (p *Provider) fetchLatest(ctx context.Context, module string) (*proxyResponse, error) {
	escaped := escapeModulePath(module)
	u := fmt.Sprintf("%s/%s/@latest", p.proxyURL, escaped)

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
		return nil, fmt.Errorf("go proxy: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var info proxyResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &info, nil
}

// versionInfoResponse represents deps.dev version info.
type versionInfoResponse struct {
	Licenses []string `json:"licenses"`
}

func (p *Provider) fetchLicense(ctx context.Context, module, version string) (string, error) {
	encoded := url.PathEscape(module)
	u := fmt.Sprintf("%s/v3alpha/systems/go/packages/%s/versions/%s", p.depsdevURL, encoded, url.PathEscape(version))

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return "", err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deps.dev version: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var info versionInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(info.Licenses) == 0 {
		return "", nil
	}
	return strings.Join(info.Licenses, " OR "), nil
}

// dependenciesResponse represents deps.dev dependencies response.
type dependenciesResponse struct {
	Nodes []struct {
		Relation string `json:"relation"`
	} `json:"nodes"`
}

func (p *Provider) fetchDependenciesCount(ctx context.Context, module, version string) (int, error) {
	encoded := url.PathEscape(module)
	u := fmt.Sprintf("%s/v3alpha/systems/go/packages/%s/versions/%s:dependencies", p.depsdevURL, encoded, url.PathEscape(version))

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
		return 0, fmt.Errorf("deps.dev dependencies: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var deps dependenciesResponse
	if err := json.NewDecoder(resp.Body).Decode(&deps); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}

	count := 0
	for _, n := range deps.Nodes {
		if n.Relation == "DIRECT" {
			count++
		}
	}
	return count, nil
}

// escapeModulePath escapes a Go module path for the Go Module Proxy.
// Uppercase letters are replaced with "!" followed by the lowercase letter.
// See: https://pkg.go.dev/golang.org/x/mod/module#EscapePath
func escapeModulePath(path string) string {
	var b strings.Builder
	for _, r := range path {
		if unicode.IsUpper(r) {
			b.WriteByte('!')
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
