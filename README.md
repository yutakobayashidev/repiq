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

# Multiple targets in NDJSON
repiq --ndjson github:facebook/react github:vuejs/core

# Mixed providers in Markdown table
repiq --markdown github:facebook/react npm:react

# Markdown table
repiq --markdown github:golang/go
```

## Supported Providers

| Provider | Target format           |
| -------- | ----------------------- |
| GitHub   | `github:<owner>/<repo>` |
| npm      | `npm:<package>`         |

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

## Authentication

Token resolution follows this priority (GitHub provider only):

1. `gh auth token` (GitHub CLI)
2. `GITHUB_TOKEN` environment variable
3. Unauthenticated (lower rate limits)

npm provider requires no authentication.

## Output Formats

- `--json` (default) -- Single JSON object or array
- `--ndjson` -- Newline-delimited JSON, one object per line
- `--markdown` -- Markdown table

## Development

```bash
# Enter dev environment (requires Nix)
nix develop

# Run tests
go test ./...

# Lint
golangci-lint run ./...
```
