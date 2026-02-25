package cache

import (
	"context"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

var _ provider.Provider = (*Provider)(nil)

// Provider wraps a provider.Provider with disk caching.
type Provider struct {
	underlying provider.Provider
	store      *Store
	noCache    bool
}

// NewProvider creates a caching decorator around the given provider.
// When noCache is true, the cache is bypassed on reads but results are
// still written (so subsequent runs without --no-cache benefit).
func NewProvider(underlying provider.Provider, store *Store, noCache bool) *Provider {
	return &Provider{
		underlying: underlying,
		store:      store,
		noCache:    noCache,
	}
}

func (p *Provider) Scheme() string {
	return p.underlying.Scheme()
}

func (p *Provider) Fetch(ctx context.Context, identifier string) (provider.Result, error) {
	key := p.underlying.Scheme() + ":" + identifier

	if !p.noCache {
		if cached, ok := p.store.Get(key); ok {
			return cached, nil
		}
	}

	result, err := p.underlying.Fetch(ctx, identifier)
	if err != nil {
		return result, err
	}

	if result.Error == "" {
		_ = p.store.Set(key, result)
	}

	return result, nil
}
