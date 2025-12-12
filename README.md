# SEKAI-CLI

A zero-dependency, highly modular CLI for interacting with SEKAI blockchain nodes running in Docker containers.

## Design Principles

1. **Zero External Dependencies** - Uses only Go standard library
2. **Highly Modular** - Each module is completely independent
3. **Simple CLI** - No Cobra, no Viper, just clean Go code
4. **Pluggable Executor** - Easy to test, easy to extend
5. **Full Coverage** - Aims to cover 100% of sekaid commands

## Installation

```bash
go install github.com/kiracore/sekai-cli/cmd/sekai-cli@latest
```

## Quick Start

```bash
# Configure
sekai-cli config init

# Check status
sekai-cli status

# Query balance
sekai-cli query bank balance <address>

# Send tokens
sekai-cli tx bank send <to> <amount> --from alice
```

## Project Structure

```
sekai-cli/
├── cmd/sekai-cli/       # Entry point
├── internal/            # Private packages
│   ├── cli/             # CLI framework (no cobra)
│   ├── executor/        # Docker execution layer
│   ├── config/          # Configuration (no viper)
│   └── output/          # Output formatting
├── pkg/                 # Public, reusable modules
│   ├── modules/         # One package per sekaid module
│   └── types/           # Shared types
├── docs/                # Documentation
└── test/                # Tests
```

## Documentation

- [Architecture Plan](docs/ARCHITECTURE.md)
- [Command Coverage](docs/COVERAGE.md)
- [Sekaid Command Reference](docs/SEKAID_COMMANDS.md)
- [Implementation Plan](docs/PLAN.md)

## Development Status

See [COVERAGE.md](docs/COVERAGE.md) for current implementation status.

## License

[License TBD]
