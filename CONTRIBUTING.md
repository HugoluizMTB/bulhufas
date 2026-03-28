# Contributing to bulhufas

Thanks for your interest in contributing!

## Setup

```bash
# Prerequisites: Go 1.22+, Ollama
git clone https://github.com/HugoluizMTB/bulhufas.git
cd bulhufas
go mod download
make test
```

## Development Workflow

1. Fork and create a branch from `main`
2. Write your code
3. Add or update tests
4. Run `make test` and `make lint`
5. Open a PR

## Pull Requests

- Link to a related issue when possible
- Keep PRs small and focused
- Add tests for new functionality
- PR title should describe the problem, not the solution

## Code Style

- `gofmt` and `goimports` for formatting
- Tabs for indentation
- Exported types and functions need godoc comments
- Table-driven tests preferred
- Interfaces for external dependencies

## Testing

```bash
make test        # run all tests with race detection
make lint        # run golangci-lint
```

Test files go next to the code they test (`handler_test.go` next to `handler.go`). Use `testdata/` for fixtures.

## Reporting Bugs

Open a GitHub Issue with:
- Go version and OS
- Steps to reproduce
- Expected vs actual behavior

## Feature Requests

Open a Discussion first. Describe the use case, not just the solution.

## AI-Assisted Contributions

AI-assisted code is welcome. Please indicate AI usage in your PR description and review the output carefully before submitting.
