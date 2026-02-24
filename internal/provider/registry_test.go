package provider

import (
	"context"
	"testing"
)

type fakeProvider struct {
	scheme string
}

func (f *fakeProvider) Scheme() string { return f.scheme }
func (f *fakeProvider) Fetch(_ context.Context, _ string) (Result, error) {
	return Result{}, nil
}

func TestRegistryRegisterAndLookup(t *testing.T) {
	r := NewRegistry()
	p := &fakeProvider{scheme: "github"}
	r.Register(p)

	got, ok := r.Lookup("github")
	if !ok {
		t.Fatal("expected to find provider for scheme 'github'")
	}
	if got.Scheme() != "github" {
		t.Errorf("got scheme %q, want %q", got.Scheme(), "github")
	}
}

func TestRegistryLookupMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Lookup("unknown")
	if ok {
		t.Fatal("expected Lookup to return false for unknown scheme")
	}
}
