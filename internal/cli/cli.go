package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yutakobayashidev/repiq/internal/auth"
	"github.com/yutakobayashidev/repiq/internal/cache"
	"github.com/yutakobayashidev/repiq/internal/format"
	"github.com/yutakobayashidev/repiq/internal/provider"
	ghprovider "github.com/yutakobayashidev/repiq/internal/provider/github"
	npmprovider "github.com/yutakobayashidev/repiq/internal/provider/npm"
)

// Version is set at build time via ldflags.
var Version = "dev"

const timeout = 30 * time.Second

// Run executes the CLI with the given arguments.
func Run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("repiq", flag.ContinueOnError)
	fs.SetOutput(stderr)

	jsonFlag := fs.Bool("json", false, "output as JSON array (default)")
	ndjsonFlag := fs.Bool("ndjson", false, "output as newline-delimited JSON")
	markdownFlag := fs.Bool("markdown", false, "output as Markdown table")
	versionFlag := fs.Bool("version", false, "print version and exit")
	noCacheFlag := fs.Bool("no-cache", false, "bypass cache and always fetch from API")

	fs.Usage = func() {
		_, _ = io.WriteString(stderr, `Usage: repiq [flags] <scheme>:<identifier> [...]

Fetch objective metrics for OSS libraries and repositories.

Examples:
  repiq github:facebook/react
  repiq npm:react
  repiq --ndjson github:facebook/react npm:react
  repiq --markdown github:golang/go

Flags:
`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *versionFlag {
		_, _ = fmt.Fprintf(stdout, "repiq %s\n", Version)
		return nil
	}

	targets := fs.Args()
	if len(targets) == 0 {
		fs.Usage()
		return fmt.Errorf("no targets specified")
	}

	// Determine output format. When multiple flags are set, priority:
	// markdown > ndjson > json (default). This matches the spec's
	// "last specified wins" intent for wrapper scripts that set defaults.
	formatter := format.JSON
	if *ndjsonFlag {
		formatter = format.NDJSON
	}
	if *markdownFlag {
		formatter = format.Markdown
	}
	if *jsonFlag && !*ndjsonFlag && !*markdownFlag {
		formatter = format.JSON
	}

	// Set up providers.
	resolver := &auth.Resolver{
		Cmd:    auth.ExecRunner{},
		Getenv: os.Getenv,
	}
	token := resolver.ResolveToken()

	registry := provider.NewRegistry()

	ghProvider := provider.Provider(ghprovider.New(token, ""))
	npmProvider := provider.Provider(npmprovider.New("", ""))

	if cacheDir, err := os.UserCacheDir(); err == nil {
		store := cache.NewStore(filepath.Join(cacheDir, "repiq"), 24*time.Hour)
		ghProvider = cache.NewProvider(ghProvider, store, *noCacheFlag)
		npmProvider = cache.NewProvider(npmProvider, store, *noCacheFlag)
	}

	registry.Register(ghProvider)
	registry.Register(npmProvider)

	// Parse and validate all targets first.
	parsed := make([]provider.Target, len(targets))
	for i, raw := range targets {
		t, err := provider.ParseTarget(raw)
		if err != nil {
			return err
		}
		if _, ok := registry.Lookup(t.Scheme); !ok {
			return fmt.Errorf("unknown scheme %q in target %q", t.Scheme, raw)
		}
		parsed[i] = t
	}

	// Fetch all targets in parallel.
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	results := make([]provider.Result, len(parsed))
	var wg sync.WaitGroup
	wg.Add(len(parsed))
	for i, t := range parsed {
		go func(i int, t provider.Target) {
			defer wg.Done()
			p, _ := registry.Lookup(t.Scheme)
			result, err := p.Fetch(ctx, t.Identifier)
			if err != nil {
				results[i] = provider.Result{
					Target: t.Scheme + ":" + t.Identifier,
					Error:  err.Error(),
				}
				return
			}
			results[i] = result
		}(i, t)
	}
	wg.Wait()

	// Output results.
	if err := formatter(stdout, results); err != nil {
		return fmt.Errorf("formatting output: %w", err)
	}

	// Exit code: 1 if any result has an error.
	for _, r := range results {
		if r.Error != "" {
			return fmt.Errorf("one or more targets failed")
		}
	}
	return nil
}
