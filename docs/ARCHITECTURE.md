# SEKAI-CLI Architecture

## Overview

SEKAI-CLI uses an **SDK-first architecture** where `pkg/sdk/` is a reusable library that can power CLI tools, web backends, bots, and other applications. The CLI is a thin wrapper around the SDK.

```
                    CONSUMERS
    ┌───────────────────┼───────────────────┐
    │                   │                   │
 sekai-cli          web-backend          bots (future)
    │                   │                   │
    └───────────────────┼───────────────────┘
                        │
              ┌─────────▼─────────┐
              │     pkg/sdk/      │  ← Core SDK (public, importable)
              │  ┌─────────────┐  │
              │  │ Client iface│  │
              │  └──────┬──────┘  │
              │         │         │
              │  ┌──────▼──────┐  │
              │  │   Modules   │  │
              │  │ bank, gov.. │  │
              │  └─────────────┘  │
              └─────────┬─────────┘
                        │
         ┌──────────────┼──────────────┐
         │              │              │
    DockerClient    RESTClient    gRPCClient (future)
         │              │              │
         └──────────────┼──────────────┘
                        │
                        ▼
               SEKAI Blockchain Node
            (RPC/gRPC/REST endpoints)
```

## Directory Structure

```
sekai-cli/
├── cmd/sekai-cli/
│   └── main.go                    # Entry point (~15 lines)
│
├── pkg/sdk/                       # ★ Core SDK (PUBLIC, importable)
│   ├── sdk.go                     # SEKAI struct, module accessors
│   ├── client.go                  # Client interface
│   ├── options.go                 # Functional options
│   ├── errors.go                  # SDK errors
│   │
│   ├── client/                    # Client implementations
│   │   ├── docker/
│   │   │   ├── docker.go          # DockerClient
│   │   │   ├── exec.go            # Command execution
│   │   │   └── parser.go          # Response parsing
│   │   ├── rest/                  # (Future: REST client)
│   │   └── mock/
│   │       └── mock.go            # MockClient for testing
│   │
│   ├── types/                     # Shared types
│   │   ├── coin.go                # Coin, Coins
│   │   ├── tx.go                  # TxResult, TxOptions
│   │   ├── address.go             # Address validation
│   │   └── query.go               # Pagination, QueryOptions
│   │
│   └── modules/                   # Business logic modules
│       ├── status/
│       ├── keys/
│       ├── bank/
│       └── gov/                   # (Future: more modules)
│
├── internal/                      # CLI-specific (PRIVATE)
│   ├── cli/                       # CLI framework
│   │   ├── command.go             # Command struct, execution
│   │   └── global.go              # Global/common flags
│   │
│   ├── app/                       # CLI ↔ SDK wiring
│   │   └── app.go                 # Application, command builders
│   │
│   ├── config/                    # Configuration management
│   │   └── config.go              # Config struct, loading
│   │
│   └── output/                    # Output formatters
│       └── formatter.go           # Text, JSON, YAML formatters
│
├── test/                          # Tests
│   ├── docker/
│   └── e2e/
│
└── docs/                          # Documentation
```

## Core Interfaces

### Client Interface (`pkg/sdk/client.go`)

The `Client` interface abstracts blockchain communication. It can be implemented by Docker exec, REST API, gRPC, etc.

```go
type Client interface {
    // Query executes a read operation
    Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error)

    // Tx executes a transaction (write operation)
    Tx(ctx context.Context, req *TxRequest) (*TxResponse, error)

    // Keys returns the keyring client
    Keys() KeysClient

    // Status returns node status
    Status(ctx context.Context) (*StatusResponse, error)

    // Close releases resources
    Close() error
}

type QueryRequest struct {
    Module   string            // e.g., "bank", "customgov"
    Endpoint string            // e.g., "balances", "proposal"
    Params   map[string]string
    RawArgs  []string
}

type TxRequest struct {
    Module           string
    Action           string
    Args             []string
    Flags            map[string]string
    Signer           string
    BroadcastMode    string
    SkipConfirmation bool
}
```

### SDK Entry Point (`pkg/sdk/sdk.go`)

```go
type SEKAI struct {
    client Client
}

func New(client Client) *SEKAI

func (s *SEKAI) Client() Client
func (s *SEKAI) Status(ctx) (*StatusResponse, error)
func (s *SEKAI) Keys() KeysClient
// Module accessors added as modules are implemented
```

### Module Pattern (`pkg/sdk/modules/bank/bank.go`)

```go
type Module struct {
    client sdk.Client
}

func New(client sdk.Client) *Module

func (m *Module) Balance(ctx, address, denom) (*types.Coin, error)
func (m *Module) Balances(ctx, address) (types.Coins, error)
func (m *Module) Send(ctx, from, to, amount, opts) (*TxResponse, error)
```

## Client Implementations

### Docker Client (`pkg/sdk/client/docker/`)

Executes commands via `docker exec <container> /sekaid <args>`.

**Usage:**
```go
client, err := docker.NewClient("sekai-node",
    docker.WithChainID("localnet-1"),
    docker.WithKeyringBackend("test"),
)
sekai := sdk.New(client)
```

### Mock Client (`pkg/sdk/client/mock/`)

For testing without Docker or network access.

**Usage:**
```go
mock := mock.NewClient()
mock.SetQueryResponse("bank", "balances", balanceData)
mock.SetTxResponse("bank", "send", &sdk.TxResponse{TxHash: "ABC123"})

sekai := sdk.New(mock)
```

### REST Client (Future: `pkg/sdk/client/rest/`)

Will communicate directly with SEKAI REST API (port 1317).

## CLI Framework (`internal/cli/`)

Lightweight CLI framework without external dependencies (no Cobra/Viper).

```go
type Command struct {
    Name        string
    Aliases     []string
    Short       string
    Long        string
    Args        []Arg
    Flags       []Flag
    Run         RunFunc
    SubCommands []*Command
}

type Context struct {
    Command *Command
    Args    []string
    Flags   map[string]string
    Stdin   io.Reader
    Stdout  io.Writer
    Stderr  io.Writer
}
```

## Module Independence

Each module in `pkg/sdk/modules/` should:

1. **Only import from**:
   - Standard library
   - `github.com/kiracore/sekai-cli/pkg/sdk` (Client interface)
   - `github.com/kiracore/sekai-cli/pkg/sdk/types` (shared types)

2. **Never import from**:
   - Other modules in `pkg/sdk/modules/`
   - External packages
   - `internal/` packages (CLI-specific)

3. **Provide**:
   - A `New(client)` constructor
   - Methods for each blockchain operation
   - Module-specific types if needed

## Adding a New Module

1. Create directory: `pkg/sdk/modules/<name>/`
2. Create main file: `<name>.go` with:
   - Module struct
   - `New(client)` constructor
   - Query and transaction methods
3. Optionally create `types.go` for module-specific types
4. Add CLI commands in `internal/app/app.go`

**Example:**
```go
// pkg/sdk/modules/staking/staking.go
package staking

type Module struct {
    client sdk.Client
}

func New(client sdk.Client) *Module {
    return &Module{client: client}
}

func (m *Module) Validators(ctx context.Context) ([]Validator, error) {
    resp, err := m.client.Query(ctx, &sdk.QueryRequest{
        Module:   "customstaking",
        Endpoint: "validators",
    })
    // ... parse response
}
```

## Testing Strategy

### Unit Tests
Use `mock.Client` to test module logic without Docker:

```go
func TestBankSend(t *testing.T) {
    mock := mock.NewClient()
    mock.SetTxResponse("bank", "send", &sdk.TxResponse{
        TxHash: "ABC123",
        Code:   0,
    })

    m := bank.New(mock)
    resp, err := m.Send(ctx, "alice", "bob", coins, nil)

    require.NoError(t, err)
    assert.Equal(t, "ABC123", resp.TxHash)
}
```

### Integration Tests
Use real Docker container:

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    client, _ := docker.NewClient("sekai-test")
    sekai := sdk.New(client)

    status, err := sekai.Status(ctx)
    require.NoError(t, err)
    assert.NotEmpty(t, status.NodeInfo.Network)
}
```

## Usage Examples

### CLI Usage

```bash
# Get node status
sekai-cli status

# List keys
sekai-cli keys list

# Query balance
sekai-cli bank balances kira1...

# Send tokens
sekai-cli bank send alice kira1... 100ukex --fees 100ukex
```

### SDK Usage (Go)

```go
package main

import (
    "context"
    "github.com/kiracore/sekai-cli/pkg/sdk"
    "github.com/kiracore/sekai-cli/pkg/sdk/client/docker"
    "github.com/kiracore/sekai-cli/pkg/sdk/modules/bank"
)

func main() {
    // Create client
    client, _ := docker.NewClient("sekai-node")
    defer client.Close()

    // Create SDK
    sekai := sdk.New(client)

    // Use modules
    bankMod := bank.New(client)
    balances, _ := bankMod.Balances(context.Background(), "kira1...")
}
```

## Design Decisions

1. **SDK Location**: `pkg/sdk/` (public, importable by external Go projects)
2. **CLI Location**: `internal/` (private, CLI-specific)
3. **Zero Dependencies**: Maintained (REST uses net/http, JSON uses encoding/json)
4. **Client Abstraction**: Single `Client` interface with `Query()` and `Tx()` methods
5. **Module Independence**: Modules only depend on SDK types and Client interface
6. **gRPC**: Deferred (requires google.golang.org/grpc dependency)
