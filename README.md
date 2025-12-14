# SEKAI-CLI

[![CI](https://github.com/kiracore/sekai-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/kiracore/sekai-cli/actions/workflows/ci.yml)
[![Release](https://github.com/kiracore/sekai-cli/actions/workflows/release.yml/badge.svg)](https://github.com/kiracore/sekai-cli/releases)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A zero-dependency, SDK-first CLI for interacting with SEKAI blockchain.

## Features

- **222 SDK Commands** - Full coverage of all SEKAI blockchain modules
- **Scenario Automation** - YAML-based playbooks for complex workflows
- **SDK-First Architecture** - Core logic in reusable `pkg/sdk/` library
- **Zero External Dependencies** - Uses only Go standard library
- **Multiple Clients** - Docker exec, REST API support
- **XDG Compliant** - Standard Linux config/cache paths

## Installation

### Debian/Ubuntu (.deb)

```bash
# Download latest release
curl -LO https://github.com/kiracore/sekai-cli/releases/latest/download/sekai-cli_VERSION_amd64.deb

# Install
sudo dpkg -i sekai-cli_*_amd64.deb
```

### Binary (Linux/macOS)

```bash
# Linux amd64
curl -LO https://github.com/kiracore/sekai-cli/releases/latest/download/sekai-cli-linux-amd64
chmod +x sekai-cli-linux-amd64
sudo mv sekai-cli-linux-amd64 /usr/local/bin/sekai-cli

# macOS arm64 (Apple Silicon)
curl -LO https://github.com/kiracore/sekai-cli/releases/latest/download/sekai-cli-darwin-arm64
chmod +x sekai-cli-darwin-arm64
sudo mv sekai-cli-darwin-arm64 /usr/local/bin/sekai-cli
```

### Build from Source

```bash
git clone https://github.com/kiracore/sekai-cli.git
cd sekai-cli
make docker-build
# Binary at ./build/sekai-cli
```

## Quick Start

```bash
# Initialize and detect network
sekai-cli init

# Check status
sekai-cli status

# List keys
sekai-cli keys list

# Query balance
sekai-cli bank balances kira1...

# Send tokens
sekai-cli bank send alice kira1... 100ukex --fees 100ukex
```

## Scenario Automation

Execute complex workflows with YAML playbooks:

```yaml
# transfer-and-delegate.yaml
name: Transfer and Delegate
steps:
  - name: send-tokens
    module: bank
    action: send
    params:
      from: genesis
      to: kira1abc...
      amount: 1000ukex

  - name: delegate-stake
    module: multistaking
    action: delegate
    params:
      from: genesis
      validator: kiravaloper1...
      amount: 500ukex
```

```bash
sekai-cli scenario run transfer-and-delegate.yaml
```

## Using the SDK

The SDK can be imported and used by other Go applications:

```go
package main

import (
    "context"
    "fmt"

    "github.com/kiracore/sekai-cli/pkg/sdk/client/docker"
    "github.com/kiracore/sekai-cli/pkg/sdk/modules/bank"
)

func main() {
    client, _ := docker.NewClient("sekai-node",
        docker.WithChainID("localnet-1"),
    )
    defer client.Close()

    bankMod := bank.New(client)
    balances, _ := bankMod.Balances(context.Background(), "kira1...")
    fmt.Println(balances)
}
```

## Configuration

Config files are stored following XDG Base Directory Specification:

| Path | Purpose |
|------|---------|
| `~/.config/sekai-cli/config.json` | User configuration |
| `~/.cache/sekai-cli/cache.json` | Network cache |
| `/etc/sekai-cli/config.json` | System-wide config |

Configure via flags, environment variables, or config file:

```bash
# Via flags
sekai-cli --container sekai-node --chain-id localnet-1 status

# Via environment
export SEKAI_CONTAINER=sekai-node
export SEKAI_CHAIN_ID=localnet-1
sekai-cli status

# Via config file
sekai-cli config init
```

## Shell Completion

Enable tab-completion for commands, subcommands, and flags.

### Bash

```bash
# System-wide (requires sudo)
sekai-cli completion bash | sudo tee /etc/bash_completion.d/sekai-cli

# Current session only
source <(sekai-cli completion bash)

# Add to ~/.bashrc for persistence
echo 'source <(sekai-cli completion bash)' >> ~/.bashrc
```

### Zsh

```bash
# Add to fpath
sekai-cli completion zsh > "${fpath[1]}/_sekai-cli"

# Or source directly
echo 'source <(sekai-cli completion zsh)' >> ~/.zshrc
```

### Fish

```bash
sekai-cli completion fish > ~/.config/fish/completions/sekai-cli.fish
```

## Modules

| Module | Commands | Description |
|--------|----------|-------------|
| auth | 6 | Account queries |
| bank | 10 | Token transfers |
| basket | 16 | Token baskets |
| bridge | 4 | Cross-chain bridge |
| collectives | 11 | Collectives management |
| custody | 5 | Custody queries |
| distributor | 5 | Fee distribution |
| gov | 82 | Governance (roles, proposals, voting) |
| keys | 16 | Key management |
| multistaking | 13 | Multi-asset staking |
| spending | 12 | Spending pools |
| staking | 4 | Validator staking |
| tokens | 7 | Token rates |
| upgrade | 4 | Network upgrades |
| **Total** | **222** | |

## Development

```bash
# Build
make docker-build

# Run tests
make docker-test

# Build .deb package
make deb

# Format code
make docker-fmt

# Lint
make docker-lint
```

### Branch Conventions

- `feature/*` - New features (auto-creates PR, bumps minor version)
- `bugfix/*` - Bug fixes (auto-creates PR, bumps patch version)
- `hotfix/*` - Urgent fixes (auto-creates PR, bumps patch version)
- `release/*` - Release preparation (auto-creates PR)
- `major/*` - Breaking changes (auto-creates PR, bumps major version)

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - SDK-first design
- [Sekaid Reference](docs/SEKAID_COMMANDS.md) - All sekaid commands
- [Mapper Process](docs/MAPPER_PROCESS.md) - Command integration guide

## License

Copyright 2025 KIRA Network. Licensed under Apache 2.0.
