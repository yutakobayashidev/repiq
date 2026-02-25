package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

func TestStoreGetMiss(t *testing.T) {
	store := NewStore(t.TempDir(), 24*time.Hour)
	_, ok := store.Get("github:facebook/react")
	if ok {
		t.Fatal("expected cache miss for empty store")
	}
}

func TestStoreSetAndGet(t *testing.T) {
	store := NewStore(t.TempDir(), 24*time.Hour)
	result := provider.Result{
		Target: "github:facebook/react",
		GitHub: &provider.GitHubMetrics{Stars: 200000, Forks: 40000},
	}

	if err := store.Set("github:facebook/react", result); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, ok := store.Get("github:facebook/react")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Target != result.Target {
		t.Errorf("Target = %q, want %q", got.Target, result.Target)
	}
	if got.GitHub == nil {
		t.Fatal("expected GitHub metrics")
	}
	if got.GitHub.Stars != 200000 {
		t.Errorf("Stars = %d, want 200000", got.GitHub.Stars)
	}
}

func TestStoreTTLExpiry(t *testing.T) {
	store := NewStore(t.TempDir(), 1*time.Millisecond)
	result := provider.Result{
		Target: "npm:react",
		NPM:    &provider.NPMMetrics{WeeklyDownloads: 5000000},
	}

	if err := store.Set("npm:react", result); err != nil {
		t.Fatalf("Set: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	_, ok := store.Get("npm:react")
	if ok {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestStoreCorruptedJSON(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir, 24*time.Hour)

	// Write a corrupted file at the expected path
	key := "github:corrupted/repo"
	path := store.path(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, ok := store.Get(key)
	if ok {
		t.Fatal("expected cache miss for corrupted JSON")
	}
}

func TestStoreDirectoryAutoCreation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "cache")
	store := NewStore(dir, 24*time.Hour)
	result := provider.Result{
		Target: "npm:react",
		NPM:    &provider.NPMMetrics{WeeklyDownloads: 100},
	}

	if err := store.Set("npm:react", result); err != nil {
		t.Fatalf("Set should create directory: %v", err)
	}

	got, ok := store.Get("npm:react")
	if !ok {
		t.Fatal("expected cache hit after auto-creation")
	}
	if got.NPM.WeeklyDownloads != 100 {
		t.Errorf("WeeklyDownloads = %d, want 100", got.NPM.WeeklyDownloads)
	}
}

func TestStoreVersionMismatch(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir, 24*time.Hour)

	result := provider.Result{
		Target: "npm:react",
		NPM:    &provider.NPMMetrics{WeeklyDownloads: 5000000},
	}
	if err := store.Set("npm:react", result); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Tamper with the cached file: set version to 0 (simulates old schema)
	key := "npm:react"
	path := store.path(key)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	tampered := []byte(`{"version":0,` + string(data[len(`{"version":2,`):]))
	if err := os.WriteFile(path, tampered, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, ok := store.Get(key)
	if ok {
		t.Fatal("expected cache miss for old schema version")
	}
}

func TestStoreCollisionFreeKeys(t *testing.T) {
	store := NewStore(t.TempDir(), 24*time.Hour)

	r1 := provider.Result{Target: "npm:@a/b_c", NPM: &provider.NPMMetrics{WeeklyDownloads: 1}}
	r2 := provider.Result{Target: "npm:@a_b/c", NPM: &provider.NPMMetrics{WeeklyDownloads: 2}}

	if err := store.Set("npm:@a/b_c", r1); err != nil {
		t.Fatalf("Set r1: %v", err)
	}
	if err := store.Set("npm:@a_b/c", r2); err != nil {
		t.Fatalf("Set r2: %v", err)
	}

	got1, ok := store.Get("npm:@a/b_c")
	if !ok {
		t.Fatal("expected hit for npm:@a/b_c")
	}
	got2, ok := store.Get("npm:@a_b/c")
	if !ok {
		t.Fatal("expected hit for npm:@a_b/c")
	}

	if got1.NPM.WeeklyDownloads != 1 {
		t.Errorf("npm:@a/b_c downloads = %d, want 1", got1.NPM.WeeklyDownloads)
	}
	if got2.NPM.WeeklyDownloads != 2 {
		t.Errorf("npm:@a_b/c downloads = %d, want 2", got2.NPM.WeeklyDownloads)
	}
}
