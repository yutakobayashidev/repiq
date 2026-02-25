---
name: repiq
description: >-
  Fetch objective metrics for OSS repositories and packages.
  Use when evaluating, comparing, or selecting libraries and repositories.
  Supports GitHub repos, npm/PyPI/crates.io packages, and Go modules.
  Returns stars, downloads, contributors, release activity, and more as structured JSON.
  No judgments, no recommendations, no scores — just numbers.
license: MIT
compatibility: Requires repiq CLI installed in PATH. Install via `go install` or Nix flake.
metadata:
  author: yutakobayashidev
  version: "1.0"
---

# repiq

Fetch objective metrics for OSS repositories and packages. Built for AI agents that need fast, non-reasoning data retrieval for library and repository evaluation.

**Role separation:** repiq provides raw data. You (the agent) do the reasoning. The human makes the final decision.

## Quick Start

```bash
# Fetch GitHub repo metrics
repiq github:facebook/react

# Fetch npm package metrics
repiq npm:react

# Compare multiple targets across providers
repiq --json github:facebook/react npm:react pypi:flask crates:serde go:golang.org/x/net
```

Output is JSON by default. Each target returns a structured result with provider-specific metrics.

## Supported Schemes

| Scheme | Format | Example |
|--------|--------|---------|
| GitHub | `github:<owner>/<repo>` | `github:facebook/react` |
| npm | `npm:<package>` | `npm:react`, `npm:@types/node` |
| PyPI | `pypi:<package>` | `pypi:requests` |
| crates.io | `crates:<crate>` | `crates:serde` |
| Go Modules | `go:<module/path>` | `go:golang.org/x/net` |

Multiple targets can be passed in a single command. They are fetched in parallel.

## Flags

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON array (default) |
| `--ndjson` | Output as newline-delimited JSON (one object per line) |
| `--markdown` | Output as Markdown tables (grouped by provider) |
| `--no-cache` | Bypass 24-hour disk cache and always fetch from API |
| `--version` | Print version and exit |

## Key Metrics by Provider

**GitHub** (8 metrics): `stars`, `forks`, `open_issues`, `contributors`, `release_count`, `last_commit_days`, `commits_30d`, `issues_closed_30d`

**npm** (5 metrics): `weekly_downloads`, `latest_version`, `last_publish_days`, `dependencies_count`, `license`

**PyPI** (7 metrics): `weekly_downloads`, `monthly_downloads`, `latest_version`, `last_publish_days`, `dependencies_count`, `license`, `requires_python`

**crates.io** (7 metrics): `downloads`, `recent_downloads`, `latest_version`, `last_publish_days`, `dependencies_count`, `license`, `reverse_dependencies`

**Go Modules** (4 metrics): `latest_version`, `last_publish_days`, `dependencies_count`, `license`

For full field descriptions and types, see [references/REFERENCE.md](./references/REFERENCE.md).

## Use Cases

### Compare libraries for the same task

```bash
repiq --json github:expressjs/express github:fastify/fastify npm:express npm:fastify
```

Compare `stars`, `commits_30d`, `weekly_downloads`, and `last_publish_days` to assess activity and adoption.

### Evaluate repository health

```bash
repiq github:some-org/some-repo
```

Check `last_commit_days` (is it maintained?), `issues_closed_30d` (is the team responsive?), and `contributors` (bus factor).

### Cross-ecosystem package comparison

```bash
repiq npm:zod pypi:pydantic crates:serde
```

Compare equivalent packages across language ecosystems using `weekly_downloads`, `dependencies_count`, and `license`.

## Authentication

**GitHub only.** Token is resolved automatically:

1. `gh auth token` (GitHub CLI — recommended)
2. `GITHUB_TOKEN` environment variable
3. Unauthenticated (60 req/hour rate limit)

Authenticated requests get 5,000 req/hour. Other providers require no authentication.

## Installation

```bash
# Go install
go install github.com/yutakobayashidev/repiq/cmd/repiq@latest

# Or use Nix flake
nix run github:yutakobayashidev/repiq
```

Verify installation:

```bash
repiq --version
```

## Caching

Results are cached on disk for 24 hours at `$XDG_CACHE_HOME/repiq/` (typically `~/.cache/repiq/`).

Use `--no-cache` to bypass cache reads (writes still occur for subsequent runs).

## Error Handling

- Failed targets return `{"target": "...", "error": "..."}` — other targets still succeed.
- Exit code is 1 if any target fails, 0 if all succeed.
- Common errors: 404 (not found), rate limit exceeded, network timeout (30s).
