# Agent Guidelines for repiq

## Commit Message Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/) with **required scope**.

Format: `<type>(<scope>): <description>`

### Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `build` - Build system, dependencies, Nix flake
- `chore` - Maintenance tasks
- `ci` - CI/CD configuration
- `refactor` - Code restructuring without behavior change
- `test` - Adding or updating tests
- `perf` - Performance improvements
- `style` - Code style (formatting, semicolons, etc.)
- `revert` - Reverting a previous commit

### Scope

Scope is **mandatory**. Use the most relevant area:

- `cli` - CLI entry point, argument parsing, output formatting
- `provider` - Provider interface, shared provider logic
- `github` - GitHub provider
- `npm` - npm provider
- `pypi` - PyPI provider
- `crates` - crates.io provider
- `go` - Go Modules provider
- `specs` - Spec documents under docs/specs/
- `nix` - Nix flake, dev environment
- `deps` - Dependency updates

### Examples

```
feat(cli): add --markdown output format
fix(github): handle rate limit errors gracefully
docs(specs): add core epic scope document
build(nix): add gitleaks pre-commit hook
test(provider): add unit tests for target parsing
refactor(cli): extract output formatter into separate package
```

### PR Titles

PR titles **must** follow the same convention. Squash merges use the PR title as the commit message.

```
docs(specs): add CLI + GitHub provider feature spec
feat(github): implement GitHub provider with metrics fetching
```

### Bad Examples

```
feat: add markdown output        # missing scope
update code                      # missing type and scope
fix(github): Fix bug.            # don't capitalize, don't end with period
Core: CLI + GitHub プロバイダー  # PR title not following convention
```

## External API Rules

Before using any external API endpoint in implementation:

1. Verify the endpoint exists with `curl` — don't assume based on naming conventions or docs from other APIs
2. Check the actual response shape with a real request before writing structs or mocks
3. Test mocks only prove the code works against themselves, not against the real API — always manual-test after implementation

Example mistake: deps.dev Go packages use `:requirements` (not `:dependencies`). The response format also differs completely from what the endpoint name implies.
