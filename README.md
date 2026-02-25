# repiq

Fetch objective metrics for OSS repositories. Built for AI agents that need fast, non-reasoning data retrieval for library/repository selection.

No judgments, no recommendations, no scores -- just numbers.

## Install

```
go install github.com/yutakobayashidev/repiq/cmd/repiq@latest
```

## Usage

```bash
# Single repo (JSON output by default)
repiq github:facebook/react

# npm package
repiq npm:react

# Scoped npm package
repiq npm:@types/node

# Python package (PyPI)
repiq pypi:requests

# Rust crate (crates.io)
repiq crates:serde

# Go module
repiq go:golang.org/x/text

# Multiple targets in NDJSON
repiq --ndjson github:facebook/react github:vuejs/core

# Mixed providers in Markdown table
repiq --markdown github:facebook/react npm:react pypi:flask crates:tokio go:golang.org/x/net

# Markdown table
repiq --markdown github:golang/go
```

## Supported Providers

| Provider   | Target format              |
| ---------- | -------------------------- |
| GitHub     | `github:<owner>/<repo>`    |
| npm        | `npm:<package>`            |
| PyPI       | `pypi:<package>`           |
| crates.io  | `crates:<crate>`           |
| Go Modules | `go:<module>`              |

## GitHub Metrics

| Metric              | Description                       |
| ------------------- | --------------------------------- |
| `stars`             | Star count                        |
| `forks`             | Fork count                        |
| `open_issues`       | Open issue count (includes PRs)   |
| `contributors`      | Number of contributors            |
| `release_count`     | Total releases                    |
| `last_commit_days`  | Days since last commit            |
| `commits_30d`       | Commits in the last 30 days       |
| `issues_closed_30d` | Issues closed in the last 30 days |

## npm Metrics

| Metric               | Description                        |
| -------------------- | ---------------------------------- |
| `weekly_downloads`   | Downloads in the last 7 days       |
| `latest_version`     | Latest published version           |
| `last_publish_days`  | Days since last publish            |
| `dependencies_count` | Number of runtime dependencies     |
| `license`            | License identifier (e.g. MIT, ISC) |

## PyPI Metrics

| Metric               | Description                                 |
| -------------------- | ------------------------------------------- |
| `weekly_downloads`   | Downloads in the last 7 days                |
| `monthly_downloads`  | Downloads in the last 30 days               |
| `latest_version`     | Latest published version                    |
| `last_publish_days`  | Days since last publish                     |
| `dependencies_count` | Number of runtime dependencies              |
| `license`            | License identifier                          |
| `requires_python`    | Python version requirement (e.g. `>=3.9`)   |

## crates.io Metrics

| Metric                  | Description                              |
| ----------------------- | ---------------------------------------- |
| `downloads`             | Total all-time downloads                 |
| `recent_downloads`      | Downloads in the last 90 days            |
| `latest_version`        | Latest stable version                    |
| `last_publish_days`     | Days since last publish                  |
| `dependencies_count`    | Number of normal dependencies            |
| `license`               | SPDX license identifier                  |
| `reverse_dependencies`  | Number of crates that depend on this one |

## Go Modules Metrics

| Metric               | Description                            |
| -------------------- | -------------------------------------- |
| `latest_version`     | Latest version tag                     |
| `last_publish_days`  | Days since last publish                |
| `dependencies_count` | Number of direct dependencies          |
| `license`            | License identifier (via deps.dev)      |

> Go does not provide public download count APIs. Use GitHub metrics for popularity signals.

## Authentication

Token resolution follows this priority (GitHub provider only):

1. `gh auth token` (GitHub CLI)
2. `GITHUB_TOKEN` environment variable
3. Unauthenticated (lower rate limits)

Other providers (npm, PyPI, crates.io, Go Modules) require no authentication.

## Output Formats

- `--json` (default) -- Single JSON object or array
- `--ndjson` -- Newline-delimited JSON, one object per line
- `--markdown` -- Markdown table

## Agent Skills

repiq is available as an [Agent Skill](https://agentskills.io/) for AI coding agents (Claude Code, Cursor, Windsurf, etc.).

```bash
npx skills add github:yutakobayashidev/repiq
```

## Development

```bash
# Enter dev environment (requires Nix)
nix develop

# Run tests
go test ./...

# Lint
golangci-lint run ./...
```
