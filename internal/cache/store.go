package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/yutakobayashidev/repiq/internal/provider"
)

// entry is the on-disk JSON structure for a cache entry.
type entry struct {
	CachedAt time.Time       `json:"cached_at"`
	Result   provider.Result `json:"result"`
}

// Store is a file-based cache keyed by scheme:identifier.
type Store struct {
	dir string
	ttl time.Duration
}

// NewStore creates a Store that writes JSON files to dir with the given TTL.
func NewStore(dir string, ttl time.Duration) *Store {
	return &Store{dir: dir, ttl: ttl}
}

// path returns the file path for a given cache key using SHA-256 hash.
func (s *Store) path(key string) string {
	h := sha256.Sum256([]byte(key))
	return filepath.Join(s.dir, hex.EncodeToString(h[:])+".json")
}

// Get retrieves a cached Result. Returns (Result, false) on miss or TTL expiry.
func (s *Store) Get(key string) (provider.Result, bool) {
	data, err := os.ReadFile(s.path(key))
	if err != nil {
		return provider.Result{}, false
	}

	var e entry
	if err := json.Unmarshal(data, &e); err != nil {
		return provider.Result{}, false
	}

	if time.Since(e.CachedAt) > s.ttl {
		return provider.Result{}, false
	}

	return e.Result, true
}

// Set writes a Result to the cache. The directory is created lazily.
// Writes are atomic: data is written to a temp file, then renamed.
func (s *Store) Set(key string, result provider.Result) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}

	e := entry{
		CachedAt: time.Now(),
		Result:   result,
	}
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(s.dir, "*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, s.path(key))
}
