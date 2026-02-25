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

repiq fetches stars, downloads, commit activity, and more from GitHub, npm, PyPI, crates.io, and Go Modules. It returns Markdown tables by default (or JSON with `--json`) -- no opinions, no scores. The agent reasons. You decide.

## Install

```bash
go install github.com/yutakobayashidev/repiq/cmd/repiq@latest
```

Or with Nix:

```bash
nix run github:yutakobayashidev/repiq
```

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

## Quick Start

```bash
repiq github:facebook/react
```

```
## GitHub

| Target | Stars | Forks | Open Issues | Contributors | Releases | Last Commit (days) | Commits (30d) | Issues Closed (30d) |
|--------|-------|-------|-------------|--------------|----------|--------------------|---------------|---------------------|
| github:facebook/react | 234000 | 47000 | 1000 | 1700 | 220 | 0 | 80 | 150 |
```

## Use Cases

- **Library comparison** -- compare downloads, maintenance activity, and dependency count across candidates in one command
- **Maintenance check** -- `last_commit_days`, `commits_30d`, and `issues_closed_30d` reveal whether a project is abandoned, stable, or actively developed
- **Download trends** -- `weekly_downloads` vs `monthly_downloads` (npm, PyPI): `weekly * 4 / monthly > 1` means accelerating adoption
- **License compliance** -- SPDX license identifiers from GitHub, npm, PyPI, crates.io, and Go for quick enterprise checks
- **Supply chain risk** -- `dependencies_count` shows how much a package pulls in; `reverse_dependencies` (crates.io) shows ecosystem penetration
- **Cross-ecosystem evaluation** -- compare equivalent packages across languages (e.g. zod vs pydantic vs serde) using GitHub as the common axis

### Examples

**Compare web frameworks**

```bash
repiq github:expressjs/express github:fastify/fastify github:honojs/hono npm:express npm:fastify npm:hono
```

**Check if a repo is still maintained**

```bash
repiq github:some-org/some-lib
```

**Evaluate ORM options**

```bash
repiq github:prisma/prisma github:drizzle-team/drizzle-orm github:typeorm/typeorm npm:prisma npm:drizzle-orm npm:typeorm
```

**License check for enterprise adoption**

```bash
repiq github:facebook/react github:preactjs/preact github:sveltejs/svelte
```

**Cross-ecosystem comparison**

```bash
repiq npm:zod pypi:pydantic crates:serde
```

## Supported Providers

| Provider | Format | Example |
|----------|--------|---------|
| GitHub | `github:<owner>/<repo>` | `github:facebook/react` |
| npm | `npm:<package>` | `npm:react`, `npm:@types/node` |
| PyPI | `pypi:<package>` | `pypi:requests` |
| crates.io | `crates:<crate>` | `crates:serde` |
| Go Modules | `go:<module>` | `go:golang.org/x/text` |

Want to add a new provider? See [Adding a Provider](docs/adding-a-provider.md) and [Contributing Guide](CONTRIBUTING.md).

## Metrics

<details>
<summary><strong>GitHub</strong> (9 metrics)</summary>

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
| `license` | SPDX license identifier (e.g. MIT, Apache-2.0) |

</details>

<details>
<summary><strong>npm</strong> (6 metrics)</summary>

| Metric | Description |
|--------|-------------|
| `weekly_downloads` | Downloads in the last 7 days |
| `monthly_downloads` | Downloads in the last 30 days |
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
| _(none)_ | Markdown | Tables grouped by provider (default) |
| `--json` | JSON | Single JSON array |
| `--ndjson` | NDJSON | One JSON object per line |

## Authentication

GitHub provider only. Token is resolved automatically:

1. `gh auth token` (GitHub CLI)
2. `GITHUB_TOKEN` environment variable
3. Unauthenticated (60 req/hour)

Other providers require no authentication.

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
