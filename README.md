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

# Multiple repos in NDJSON
repiq --ndjson github:facebook/react github:vuejs/core

# Markdown table
repiq --markdown github:golang/go
```

## Supported Providers

| Provider | Target format           |
| -------- | ----------------------- |
| GitHub   | `github:<owner>/<repo>` |

## GitHub Metrics

| Metric              | Description                       |
| ------------------- | --------------------------------- |
| `stars`             | Star count                        |
| `forks`             | Fork count                        |
| `open_issues`       | Open issue count                  |
| `contributors`      | Number of contributors            |
| `release_count`     | Total releases                    |
| `last_commit_days`  | Days since last commit            |
| `commits_30d`       | Commits in the last 30 days       |
| `issues_closed_30d` | Issues closed in the last 30 days |

## Authentication

Token resolution follows this priority:

1. `gh auth token` (GitHub CLI)
2. `GITHUB_TOKEN` environment variable
3. Unauthenticated (lower rate limits)

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
