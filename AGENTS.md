# Agent Guidelines for repiq

## Commit Scopes

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
- `cache` - Cache layer

## External API Rules

Before using any external API endpoint in implementation:

1. Verify the endpoint exists with `curl` — don't assume based on naming conventions or docs from other APIs
2. Check the actual response shape with a real request before writing structs or mocks
3. Test mocks only prove the code works against themselves, not against the real API — always manual-test after implementation

Example mistake: deps.dev Go packages use `:requirements` (not `:dependencies`). The response format also differs completely from what the endpoint name implies.
