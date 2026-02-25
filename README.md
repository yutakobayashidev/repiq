# repiq

[![CI](https://github.com/yutakobayashidev/repiq/actions/workflows/ci.yml/badge.svg)](https://github.com/yutakobayashidev/repiq/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/github/go-mod/go-version/yutakobayashidev/repiq)](go.mod)
[![DeepWiki](https://img.shields.io/badge/DeepWiki-yutakobayashidev%2Frepiq-blue.svg?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACwAAAAyCAYAAAAnWDnqAAAAAXNSR0IArs4c6QAAA05JREFUaEPtmUtyEzEQhtWTQyQLHNak2AB7ZnyXZMEjXMGeK/AIi+QuHrMnbChYY7MIh8g01fJoopFb0uhhEqqcbWTp06/uv1saEDv4O3n3dV60RfP947Mm9/SQc0ICFQgzfc4CYZoTPAswgSJCCUJUnAAoRHOAUOcATwbmVLWdGoH//PB8mnKqScAhsD0kYP3j/Yt5LPQe2KvcXmGvRHcDnpxfL2zOYJ1mFwrryWTz0advv1Ut4CJgf5uhDuDj5eUcAUoahrdY/56ebRWeraTjMt/00Sh3UDtjgHtQNHwcRGOC98BJEAEymycmYcWwOprTgcB6VZ5JK5TAJ+fXGLBm3FDAmn6oPPjR4rKCAoJCal2eAiQp2x0vxTPB3ALO2CRkwmDy5WohzBDwSEFKRwPbknEggCPB/imwrycgxX2NzoMCHhPkDwqYMr9tRcP5qNrMZHkVnOjRMWwLCcr8ohBVb1OMjxLwGCvjTikrsBOiA6fNyCrm8V1rP93iVPpwaE+gO0SsWmPiXB+jikdf6SizrT5qKasx5j8ABbHpFTx+vFXp9EnYQmLx02h1QTTrl6eDqxLnGjporxl3NL3agEvXdT0WmEost648sQOYAeJS9Q7bfUVoMGnjo4AZdUMQku50McDcMWcBPvr0SzbTAFDfvJqwLzgxwATnCgnp4wDl6Aa+Ax283gghmj+vj7feE2KBBRMW3FzOpLOADl0Isb5587h/U4gGvkt5v60Z1VLG8BhYjbzRwyQZemwAd6cCR5/XFWLYZRIMpX39AR0tjaGGiGzLVyhse5C9RKC6ai42ppWPKiBagOvaYk8lO7DajerabOZP46Lby5wKjw1HCRx7p9sVMOWGzb/vA1hwiWc6jm3MvQDTogQkiqIhJV0nBQBTU+3okKCFDy9WwferkHjtxib7t3xIUQtHxnIwtx4mpg26/HfwVNVDb4oI9RHmx5WGelRVlrtiw43zboCLaxv46AZeB3IlTkwouebTr1y2NjSpHz68WNFjHvupy3q8TFn3Hos2IAk4Ju5dCo8B3wP7VPr/FGaKiG+T+v+TQqIrOqMTL1VdWV1DdmcbO8KXBz6esmYWYKPwDL5b5FA1a0hwapHiom0r/cKaoqr+27/XcrS5UwSMbQAAAABJRU5ErkJggg==)](https://deepwiki.com/yutakobayashidev/repiq)

AI agents pick libraries by guessing. repiq gives them real data instead.

```
You: "Add an HTTP client library"
Agent: repiq npm:axios npm:ky npm:got → compares downloads, maintenance, dependencies → recommends with evidence
```

repiq fetches stars, downloads, commit activity, and more from GitHub, npm, PyPI, crates.io, and Go Modules. It returns raw JSON -- no opinions, no scores. The agent reasons. You decide.

## Install

```bash
go install github.com/yutakobayashidev/repiq/cmd/repiq@latest
```

Or with Nix:

```bash
nix run github:yutakobayashidev/repiq
```

## Quick Start

```bash
repiq github:facebook/react
```

```json
[
  {
    "target": "github:facebook/react",
    "github": {
      "stars": 234000,
      "forks": 47000,
      "open_issues": 1000,
      "contributors": 1700,
      "release_count": 220,
      "last_commit_days": 0,
      "commits_30d": 80,
      "issues_closed_30d": 150
    }
  }
]
```

## Use Cases

**Compare libraries for the same task**

```bash
repiq npm:express npm:fastify npm:hono
```

An agent can compare `weekly_downloads`, `last_commit_days`, and `dependencies_count` to recommend the best fit.

**Check if a repo is still maintained**

```bash
repiq github:some-org/some-lib
```

`last_commit_days: 400` and `issues_closed_30d: 0` tells the agent this project may be abandoned.

**Cross-ecosystem comparison**

```bash
repiq npm:zod pypi:pydantic crates:serde
```

Compare equivalent packages across languages using downloads, dependency count, and license.

## Supported Providers

| Provider | Format | Example |
|----------|--------|---------|
| GitHub | `github:<owner>/<repo>` | `github:facebook/react` |
| npm | `npm:<package>` | `npm:react`, `npm:@types/node` |
| PyPI | `pypi:<package>` | `pypi:requests` |
| crates.io | `crates:<crate>` | `crates:serde` |
| Go Modules | `go:<module>` | `go:golang.org/x/text` |

## Metrics

<details>
<summary><strong>GitHub</strong> (8 metrics)</summary>

| Metric | Description |
|--------|-------------|
| `stars` | Star count |
| `forks` | Fork count |
| `open_issues` | Open issue count (includes PRs) |
| `contributors` | Number of contributors |
| `release_count` | Total releases |
| `last_commit_days` | Days since last commit |
| `commits_30d` | Commits in the last 30 days |
| `issues_closed_30d` | Issues closed in the last 30 days |

</details>

<details>
<summary><strong>npm</strong> (5 metrics)</summary>

| Metric | Description |
|--------|-------------|
| `weekly_downloads` | Downloads in the last 7 days |
| `latest_version` | Latest published version |
| `last_publish_days` | Days since last publish |
| `dependencies_count` | Number of runtime dependencies |
| `license` | License identifier (e.g. MIT, ISC) |

</details>

<details>
<summary><strong>PyPI</strong> (7 metrics)</summary>

| Metric | Description |
|--------|-------------|
| `weekly_downloads` | Downloads in the last 7 days |
| `monthly_downloads` | Downloads in the last 30 days |
| `latest_version` | Latest published version |
| `last_publish_days` | Days since last publish |
| `dependencies_count` | Number of runtime dependencies |
| `license` | License identifier |
| `requires_python` | Python version requirement (e.g. `>=3.9`) |

</details>

<details>
<summary><strong>crates.io</strong> (7 metrics)</summary>

| Metric | Description |
|--------|-------------|
| `downloads` | Total all-time downloads |
| `recent_downloads` | Downloads in the last 90 days |
| `latest_version` | Latest stable version |
| `last_publish_days` | Days since last publish |
| `dependencies_count` | Number of normal dependencies |
| `license` | SPDX license identifier |
| `reverse_dependencies` | Number of crates that depend on this one |

</details>

<details>
<summary><strong>Go Modules</strong> (4 metrics)</summary>

| Metric | Description |
|--------|-------------|
| `latest_version` | Latest version tag |
| `last_publish_days` | Days since last publish |
| `dependencies_count` | Number of direct dependencies |
| `license` | License identifier (via deps.dev) |

> Go does not provide public download count APIs. Use GitHub metrics for popularity signals.

</details>

## Output Formats

| Flag | Format | Description |
|------|--------|-------------|
| `--json` | JSON | Single JSON array (default) |
| `--ndjson` | NDJSON | One JSON object per line |
| `--markdown` | Markdown | Tables grouped by provider |

## Authentication

GitHub provider only. Token is resolved automatically:

1. `gh auth token` (GitHub CLI)
2. `GITHUB_TOKEN` environment variable
3. Unauthenticated (60 req/hour)

Other providers require no authentication.

## Agent Skills

repiq is available as an [Agent Skill](https://agentskills.io/) for AI coding agents (Claude Code, Cursor, Windsurf, etc.).

```bash
npx skills add github:yutakobayashidev/repiq
```

With [agent-skills-nix](https://github.com/Kyure-A/agent-skills-nix):

```nix
sources.repiq = {
  github = { owner = "yutakobayashidev"; repo = "repiq"; };
};
skills.enable = [ "repiq" ];
targets.claude.enable = true;
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

## License

[MIT](LICENSE)
