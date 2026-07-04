# IP River CLI

[![CI](https://github.com/ipriverdev/cli/actions/workflows/ci.yml/badge.svg)](https://github.com/ipriverdev/cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ipriverdev/cli)](https://github.com/ipriverdev/cli/releases/latest)
[![License](https://img.shields.io/github/license/ipriverdev/cli)](LICENSE)

Bring your IP River Portal experience to the command line.

## Installation

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/ipriverdev/cli/main/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/ipriverdev/cli/main/install.ps1 | iex
```

### Homebrew

```bash
brew install ipriverdev/tap/ipriver
```

### Binary download

Download the latest release for your platform from [GitHub Releases](https://github.com/ipriverdev/cli/releases).

## Usage

```bash
$ ipriver [command] [subcommand] {parameters}
```

The shorthand `ipr` is installed automatically alongside `ipriver`. Both commands are identical.

## Quick start

```bash
ipriver login

# Search for the available addresses
ipriver address postcode "SW1A 1AA"

# Search for available products
ipriver check --postcode "SW1A 1AA" --uprn 10033544614
```

## Commands

See [USAGE.md](USAGE.md) for full command reference with examples, flags, and JSON output.

## Global flags

| Flag | Description |
|------|-------------|
| `--format json` | Output as JSON |
| `--no-color` | Disable color output |
| `-v`, `--version` | Print version information |

## Contributing

### Prerequisites

- Go 1.26+
- GNU Make
- [golangci-lint](https://golangci-lint.run/) (for linting)

### Setup

```bash
git clone https://github.com/ipriverdev/cli.git
cd cli
make build
```

### Build & test

```bash
make build        # → bin/ipriver
make install      # → $GOPATH/bin/ipriver
make test         # run tests
make lint         # golangci-lint
make fmt          # format code (gofmt + goimports)
make cross        # cross-compile to dist/
```

### Workflow

1. Fork the repo and create a feature branch from `main`.
2. Make your changes
3. Run `make lint && make test` before pushing.
4. Open a pull request against `main`.

### Project structure

```
cmd/             CLI commands (cobra)
internal/
  api/           HTTP client
  app/           Global flags and state
  auth/          Authentication, token management
  config/        YAML config loading/saving
  ui/            Terminal output helpers
```

### Release

Tagged releases are built with [GoReleaser](https://goreleaser.com):

```bash
goreleaser release --clean
```
