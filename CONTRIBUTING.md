# Contributing to repiq

## Development Setup

```bash
# Requires Nix
nix develop

# Or with Go 1.24+
go test ./...
golangci-lint run ./...
```

## Making Changes

1. Fork the repository
2. Create a branch from `main`
3. Make your changes
4. Run tests and lint: `go test ./... && golangci-lint run ./...`
5. Commit following the [commit convention](#commit-messages)
6. Open a Pull Request

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) with **required scope**.

```
<type>(<scope>): <description>
```

**Types:** `feat`, `fix`, `docs`, `build`, `chore`, `ci`, `refactor`, `test`, `perf`, `style`, `revert`

**Scopes:** `cli`, `provider`, `github`, `npm`, `pypi`, `crates`, `go`, `nix`, `deps`

Examples:

```
feat(cli): add --csv output format
fix(github): handle rate limit errors gracefully
test(npm): add test for scoped package parsing
```

PR titles must follow the same convention (squash merges use the PR title).

## Adding a New Provider

See [docs/adding-a-provider.md](docs/adding-a-provider.md) for a step-by-step guide with code examples and a checklist.

## Code Guidelines

- No external dependencies unless strictly necessary (only `google/go-github` currently)
- All providers must implement the `Provider` interface
- Use `httptest` for API mocking in tests
- Validate identifiers to prevent SSRF/injection
- Errors go in `Result.Error`, not Go error returns (partial failure support)
- Always verify external API endpoints with `curl` before writing code against them

## Tests

Every provider must have tests for:

- Scheme name
- Successful fetch
- Not found (404)
- Empty/invalid identifier
- Partial failure (some APIs fail, others succeed)

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run a specific package
go test ./internal/provider/github/...
```

## Reporting Issues

- Use [GitHub Issues](https://github.com/yutakobayashidev/repiq/issues)
- Include the command you ran and the output
- Include `repiq --version` output
