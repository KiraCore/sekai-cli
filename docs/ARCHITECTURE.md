# SEKAI-CLI Architecture

## Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          sekai-cli (Go Binary)                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  cmd/sekai-cli/main.go                                                      │
│         │                                                                   │
│         ▼                                                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     internal/cli/                                    │   │
│  │  Simple CLI framework - parses args, routes to modules               │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│         │                                                                   │
│         ▼                                                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     pkg/modules/                                     │   │
│  │  ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐ ┌───────┐       │   │
│  │  │ bank  │ │ keys  │ │  gov  │ │staking│ │tokens │ │  ...  │       │   │
│  │  └───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘ └───┬───┘       │   │
│  │      └─────────┴─────────┴─────────┴─────────┴─────────┘            │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                           ┌────────▼────────┐                              │
│                           │ internal/executor│                              │
│                           │  DockerExecutor  │                              │
│                           └────────┬────────┘                              │
│                                    │                                        │
└────────────────────────────────────┼────────────────────────────────────────┘
                                     │
                                     ▼
                    docker exec <container> /sekaid <args>
```

## Package Structure

### `cmd/sekai-cli/main.go`
Entry point. ~30 lines. Just initializes and runs.

### `internal/cli/`
Simple CLI framework without external dependencies.

```go
// command.go
type Command struct {
    Name        string
    Short       string
    Long        string
    Args        string           // "[required] <optional>"
    Flags       []Flag
    Run         RunFunc
    SubCommands []*Command
}

type Flag struct {
    Name     string
    Short    string
    Default  string
    Required bool
    Usage    string
}

type RunFunc func(ctx *Context) error

type Context struct {
    Args   []string
    Flags  map[string]string
    Stdin  io.Reader
    Stdout io.Writer
    Stderr io.Writer
}
```

### `internal/executor/`
Docker execution abstraction.

```go
// executor.go
type Executor interface {
    Execute(args ...string) (*Result, error)
    ExecuteTx(args ...string) (*Result, error)
    ExecuteQuery(args ...string) (*Result, error)
}

type Result struct {
    Stdout   string
    Stderr   string
    ExitCode int
}

// docker.go - Docker implementation
type DockerExecutor struct {
    Container      string
    ChainID        string
    Home           string
    KeyringBackend string
    Node           string
    Fees           string
    GasAdjustment  float64
}

// mock.go - Mock for testing
type MockExecutor struct {
    Responses map[string]*Result
}
```

### `internal/config/`
Simple YAML configuration without Viper.

```go
// config.go
type Config struct {
    Container      string  `yaml:"container"`
    ChainID        string  `yaml:"chain_id"`
    Home           string  `yaml:"home"`
    Node           string  `yaml:"node"`
    KeyringBackend string  `yaml:"keyring_backend"`
    Fees           string  `yaml:"fees"`
    GasAdjustment  float64 `yaml:"gas_adjustment"`
    Output         string  `yaml:"output"`
}

func Load(path string) (*Config, error)
func Save(path string, cfg *Config) error
func Default() *Config
```

### `internal/output/`
Output formatting.

```go
// formatter.go
type Formatter interface {
    Format(data interface{}) (string, error)
}

// Implementations
type TextFormatter struct{}
type JSONFormatter struct{}
type YAMLFormatter struct{}
```

### `pkg/modules/`
Each module is self-contained and independent.

```go
// Example: pkg/modules/bank/bank.go
package bank

type Module struct {
    exec executor.Executor
}

func New(exec executor.Executor) *Module {
    return &Module{exec: exec}
}

// Transaction methods
func (m *Module) Send(from, to, amount string, opts *SendOpts) (*types.TxResult, error)
func (m *Module) MultiSend(from string, outputs []Output, opts *SendOpts) (*types.TxResult, error)

// Query methods
func (m *Module) Balance(address string) (*Balance, error)
func (m *Module) Balances(address string) ([]Balance, error)
func (m *Module) TotalSupply() ([]Coin, error)

// CLI registration
func (m *Module) Commands() []*cli.Command
```

### `pkg/types/`
Shared types used across modules.

```go
// tx.go
type TxResult struct {
    TxHash    string
    Code      uint32
    Height    string
    GasUsed   string
    GasWanted string
    RawLog    string
}

// coin.go
type Coin struct {
    Denom  string
    Amount string
}

// address.go
func IsValidAddress(addr string) bool
func IsValidValAddress(addr string) bool
```

## Module Independence

Each module in `pkg/modules/` should:

1. **Only import from**:
   - Standard library
   - `internal/executor`
   - `internal/cli`
   - `pkg/types`

2. **Never import from**:
   - Other modules in `pkg/modules/`
   - External packages

3. **Provide**:
   - A `New(executor)` constructor
   - Methods for each sekaid command
   - A `Commands()` method for CLI registration

## Adding a New Module

1. Create directory: `pkg/modules/<name>/`
2. Create files:
   - `<name>.go` - Module struct and constructor
   - `tx_*.go` - Transaction commands
   - `query_*.go` - Query commands
   - `types.go` - Module-specific types
3. Register in `cmd/sekai-cli/main.go`

## Testing Strategy

1. **Unit Tests**: Use `MockExecutor` to test module logic
2. **Integration Tests**: Use real Docker container
3. **E2E Tests**: Full command-line testing

```go
// Example unit test
func TestBankSend(t *testing.T) {
    mock := &executor.MockExecutor{
        Responses: map[string]*executor.Result{
            "tx bank send": {Stdout: `{"txhash":"ABC123"}`, ExitCode: 0},
        },
    }

    m := bank.New(mock)
    result, err := m.Send("alice", "bob", "100ukex", nil)

    assert.NoError(t, err)
    assert.Equal(t, "ABC123", result.TxHash)
}
```
