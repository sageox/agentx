# Contributing

We welcome contributions! This project has a slower release cadence, so PRs may take some time to review.

## Quick Start

```bash
git clone https://github.com/YOUR_USERNAME/agentx.git
cd agentx
make check   # fmt + lint + test
```

**Prerequisites:** Go 1.22+, [golangci-lint](https://golangci-lint.run), [gotestsum](https://github.com/gotestyourself/gotestsum)

## Submitting Changes

1. Fork and branch from `main`
2. Write tests for new functionality
3. Run `make check` before submitting
4. Open a pull request

## Adding a New Agent

1. Create the agent file in `agents/`
2. Implement the `Agent` interface
3. Register it in `setup/setup.go`
4. Add detection tests

## Reporting Issues

File a [GitHub issue](https://github.com/sageox/agentx/issues) with steps to reproduce.

For security vulnerabilities, email security@sageox.com instead.
