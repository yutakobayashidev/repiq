package provider

// Registry maps scheme names to providers.
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

// Register adds a provider to the registry, keyed by its scheme.
func (r *Registry) Register(p Provider) {
	r.providers[p.Scheme()] = p
}

// Lookup returns the provider for the given scheme.
func (r *Registry) Lookup(scheme string) (Provider, bool) {
	p, ok := r.providers[scheme]
	return p, ok
}
