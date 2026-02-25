# repiq Output Reference

Detailed JSON output schema for each provider. See [SKILL.md](../SKILL.md) for usage instructions.

## Result Structure

Each target produces a `Result` object:

```json
{
  "target": "scheme:identifier",
  "github": { ... },
  "npm": { ... },
  "pypi": { ... },
  "crates": { ... },
  "go": { ... },
  "error": "error message if failed"
}
```

- Only the matching provider field is populated per result.
- `error` is present only when the fetch failed. Partial results may include both metrics and an error.

## GitHub Metrics

JSON key: `github`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Stars | int | `stars` | Total star count |
| Forks | int | `forks` | Total fork count |
| Open Issues | int | `open_issues` | Open issue count (includes pull requests per GitHub API) |
| Contributors | int | `contributors` | Number of contributors |
| Release Count | int | `release_count` | Total number of releases |
| Last Commit Days | int | `last_commit_days` | Days since the most recent commit |
| Commits 30d | int | `commits_30d` | Number of commits in the last 30 days |
| Issues Closed 30d | int | `issues_closed_30d` | Number of issues closed in the last 30 days |

## npm Metrics

JSON key: `npm`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Weekly Downloads | int | `weekly_downloads` | Downloads in the last 7 days |
| Latest Version | string | `latest_version` | Latest published version |
| Last Publish Days | int | `last_publish_days` | Days since last publish |
| Dependencies Count | int | `dependencies_count` | Number of runtime dependencies |
| License | string | `license` | SPDX license identifier (e.g. MIT, ISC) |

## PyPI Metrics

JSON key: `pypi`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Weekly Downloads | int | `weekly_downloads` | Downloads in the last 7 days |
| Monthly Downloads | int | `monthly_downloads` | Downloads in the last 30 days |
| Latest Version | string | `latest_version` | Latest published version |
| Last Publish Days | int | `last_publish_days` | Days since last publish |
| Dependencies Count | int | `dependencies_count` | Number of runtime dependencies |
| License | string | `license` | SPDX license identifier |
| Requires Python | string | `requires_python` | Minimum Python version (e.g. `>=3.9`) |

## crates.io Metrics

JSON key: `crates`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Downloads | int | `downloads` | Total all-time downloads |
| Recent Downloads | int | `recent_downloads` | Downloads in the last 90 days |
| Latest Version | string | `latest_version` | Latest stable version |
| Last Publish Days | int | `last_publish_days` | Days since last publish |
| Dependencies Count | int | `dependencies_count` | Number of normal dependencies |
| License | string | `license` | SPDX license identifier |
| Reverse Dependencies | int | `reverse_dependencies` | Number of crates that depend on this one |

## Go Modules Metrics

JSON key: `go`

| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| Latest Version | string | `latest_version` | Latest version tag |
| Last Publish Days | int | `last_publish_days` | Days since last publish |
| Dependencies Count | int | `dependencies_count` | Number of direct dependencies |
| License | string | `license` | SPDX license identifier (via deps.dev) |

> Go does not provide public download count APIs. Use GitHub metrics for popularity signals.

## Output Formats

### JSON (default)

Single JSON array. Each element is a Result object.

```bash
repiq github:facebook/react npm:react
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
  },
  {
    "target": "npm:react",
    "npm": {
      "weekly_downloads": 28000000,
      "latest_version": "19.1.0",
      "last_publish_days": 10,
      "dependencies_count": 0,
      "license": "MIT"
    }
  }
]
```

### NDJSON (`--ndjson`)

One JSON object per line, no array wrapper.

```bash
repiq --ndjson github:facebook/react npm:react
```

```
{"target":"github:facebook/react","github":{"stars":234000,...}}
{"target":"npm:react","npm":{"weekly_downloads":28000000,...}}
```

### Markdown (`--markdown`)

Provider-grouped tables.

```bash
repiq --markdown github:facebook/react npm:react
```

```markdown
| target | stars | forks | open_issues | contributors | release_count | last_commit_days | commits_30d | issues_closed_30d |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| github:facebook/react | 234000 | 47000 | 1000 | 1700 | 220 | 0 | 80 | 150 |

| target | weekly_downloads | latest_version | last_publish_days | dependencies_count | license |
| --- | --- | --- | --- | --- | --- |
| npm:react | 28000000 | 19.1.0 | 10 | 0 | MIT |
```

## Error Handling

When a target fails, the result contains only `target` and `error`:

```json
{
  "target": "github:nonexistent/repo",
  "error": "GitHub API: 404 Not Found"
}
```

- Partial failures are possible: successful targets return metrics even if other targets fail.
- The CLI exits with code 1 if any target has an error.
