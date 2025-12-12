# SEKAI-CLI Implementation Plan

## Overview

Build a zero-dependency, highly modular CLI for SEKAI blockchain operations.

## Goals

1. **Zero external dependencies** - Only Go standard library
2. **100% sekaid coverage** - Support all ~150 commands
3. **Modular design** - Each module independent
4. **Easy to test** - Mock executor for unit tests
5. **Easy to extend** - Adding new modules is straightforward

## Phase 1: Foundation

### 1.1 CLI Framework (`internal/cli/`)

```
Files to create:
├── command.go      # Command struct and tree
├── parser.go       # Argument parsing
├── flags.go        # Flag handling
├── help.go         # Help text generation
├── context.go      # Execution context
└── cli_test.go     # Tests
```

**Key interfaces:**
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

type RunFunc func(ctx *Context) error
```

### 1.2 Executor (`internal/executor/`)

```
Files to create:
├── executor.go     # Interface definition
├── docker.go       # Docker implementation
├── result.go       # Result struct
├── mock.go         # Mock for testing
└── executor_test.go
```

**Key interfaces:**
```go
type Executor interface {
    Execute(args ...string) (*Result, error)
    ExecuteTx(args ...string) (*Result, error)
    ExecuteQuery(args ...string) (*Result, error)
    Container() string
    SetContainer(name string)
}
```

### 1.3 Configuration (`internal/config/`)

```
Files to create:
├── config.go       # Config struct and loading
├── defaults.go     # Default values
├── yaml.go         # Simple YAML parser (no external deps)
└── config_test.go
```

### 1.4 Output (`internal/output/`)

```
Files to create:
├── formatter.go    # Formatter interface
├── text.go         # Text output
├── json.go         # JSON output
├── yaml.go         # YAML output
└── output_test.go
```

### 1.5 Shared Types (`pkg/types/`)

```
Files to create:
├── address.go      # Address validation
├── coin.go         # Coin struct
├── tx.go           # TxResult struct
└── types_test.go
```

## Phase 2: Core Modules

### 2.1 Status Module

```
pkg/modules/status/
├── status.go       # Module + Commands()
├── query.go        # Status query
└── types.go        # NodeStatus struct
```

### 2.2 Keys Module

```
pkg/modules/keys/
├── keys.go         # Module + Commands()
├── add.go          # keys add
├── delete.go       # keys delete
├── list.go         # keys list
├── show.go         # keys show
├── export.go       # keys export
├── import.go       # keys import
├── types.go        # Key, KeyInfo structs
└── keys_test.go
```

### 2.3 Bank Module

```
pkg/modules/bank/
├── bank.go         # Module + Commands()
├── tx_send.go      # tx bank send
├── tx_multisend.go # tx bank multi-send
├── query_balance.go    # query bank balances
├── query_total.go      # query bank total
├── query_spendable.go  # query bank spendable
├── types.go        # Balance, Coin structs
└── bank_test.go
```

## Phase 3: Governance Module

```
pkg/modules/gov/
├── gov.go              # Module + Commands()
│
├── query_proposal.go   # query customgov proposal
├── query_proposals.go  # query customgov proposals
├── query_permissions.go # query customgov permissions
├── query_roles.go      # query customgov roles
├── query_network.go    # query customgov network-properties
├── query_councilors.go # query customgov councilors
├── query_votes.go      # query customgov votes
├── query_identity.go   # query customgov identity-*
│
├── tx_vote.go          # tx customgov proposal vote
├── tx_permission.go    # tx customgov permission *
├── tx_role.go          # tx customgov role *
├── tx_councilor.go     # tx customgov councilor *
├── tx_identity.go      # tx customgov *-identity-*
├── tx_proposal.go      # tx customgov proposal *
│
├── types.go            # Proposal, Permission, Role structs
└── gov_test.go
```

## Phase 4: Staking Modules

### 4.1 Staking Module

```
pkg/modules/staking/
├── staking.go          # Module + Commands()
├── tx_claim.go         # claim-validator-seat
├── tx_activate.go      # activate
├── tx_pause.go         # pause
├── tx_unpause.go       # unpause
├── query_validator.go  # query validator(s)
├── types.go
└── staking_test.go
```

### 4.2 MultiStaking Module

```
pkg/modules/multistaking/
├── multistaking.go     # Module + Commands()
├── tx_delegate.go      # delegate
├── tx_undelegate.go    # undelegate
├── tx_claim.go         # claim-rewards, claim-undelegation
├── tx_register.go      # register-delegator
├── tx_compound.go      # set-compound-info
├── tx_pool.go          # upsert-staking-pool
├── query.go            # queries
├── types.go
└── multistaking_test.go
```

## Phase 5: Token Modules

### 5.1 Tokens Module

```
pkg/modules/tokens/
├── tokens.go           # Module + Commands()
├── tx_rate.go          # upsert-rate
├── tx_proposal.go      # proposals
├── query_rate.go       # rate queries
├── query_blackwhite.go # black-white queries
├── types.go
└── tokens_test.go
```

### 5.2 Spending Module

```
pkg/modules/spending/
├── spending.go         # Module + Commands()
├── tx_create.go        # create-spending-pool
├── tx_deposit.go       # deposit-spending-pool
├── tx_claim.go         # claim-spending-pool
├── tx_register.go      # register-beneficiary
├── tx_proposal.go      # proposals
├── query.go            # queries
├── types.go
└── spending_test.go
```

### 5.3 UBI Module

```
pkg/modules/ubi/
├── ubi.go              # Module + Commands()
├── tx_proposal.go      # proposals
├── query.go            # queries
├── types.go
└── ubi_test.go
```

## Phase 6: Advanced Modules

### 6.1 Basket Module
### 6.2 Collectives Module
### 6.3 Custody Module
### 6.4 Bridge Module
### 6.5 Layer2 Module
### 6.6 Recovery Module
### 6.7 Upgrade Module

## Entry Point

```
cmd/sekai-cli/main.go
```

```go
package main

import (
    "os"

    "github.com/kiracore/sekai-cli/internal/cli"
    "github.com/kiracore/sekai-cli/internal/config"
    "github.com/kiracore/sekai-cli/internal/executor"
    "github.com/kiracore/sekai-cli/pkg/modules/bank"
    "github.com/kiracore/sekai-cli/pkg/modules/gov"
    "github.com/kiracore/sekai-cli/pkg/modules/keys"
    "github.com/kiracore/sekai-cli/pkg/modules/staking"
    "github.com/kiracore/sekai-cli/pkg/modules/status"
    // ... more modules
)

func main() {
    // Load config
    cfg := config.Load()

    // Create executor
    exec := executor.NewDocker(cfg)

    // Create root command
    root := &cli.Command{
        Name:  "sekai-cli",
        Short: "SEKAI blockchain CLI",
    }

    // Register modules
    root.AddCommands(
        status.New(exec).Commands(),
        keys.New(exec).Commands(),
        bank.New(exec).Commands(),
        gov.New(exec).Commands(),
        staking.New(exec).Commands(),
        // ... more modules
    )

    // Run
    if err := root.Execute(os.Args[1:]); err != nil {
        os.Exit(1)
    }
}
```

## Testing Strategy

### Unit Tests
- Each module has `*_test.go` files
- Use `MockExecutor` to simulate responses
- Test all edge cases

### Integration Tests
- Located in `test/e2e/`
- Require running Docker container
- Test real command execution

### CI/CD
- Run unit tests on every PR
- Run integration tests on merge to main
- Build and release binaries

## Milestones

### Milestone 1: Bootstrap
- [ ] CLI framework working
- [ ] Executor working
- [ ] Config working
- [ ] Status command working

### Milestone 2: Core
- [ ] Keys module complete
- [ ] Bank module complete
- [ ] Basic queries working

### Milestone 3: Governance
- [ ] Gov queries complete
- [ ] Gov transactions complete
- [ ] Voting working

### Milestone 4: Staking
- [ ] Staking module complete
- [ ] MultiStaking module complete

### Milestone 5: Tokens
- [ ] Tokens module complete
- [ ] Spending module complete
- [ ] UBI module complete

### Milestone 6: Advanced
- [ ] All remaining modules
- [ ] 100% coverage

## File Count Estimate

```
internal/cli/       ~6 files
internal/executor/  ~4 files
internal/config/    ~3 files
internal/output/    ~4 files
pkg/types/          ~4 files
pkg/modules/        ~15 modules × ~8 files = ~120 files
cmd/sekai-cli/      ~1 file
docs/               ~4 files
test/               ~10 files
─────────────────────────────
Total:              ~156 files
```

## Getting Started

1. Start with `internal/cli/` - the foundation
2. Then `internal/executor/` - to run commands
3. Then `pkg/modules/status/` - simplest module
4. Then `pkg/modules/keys/` - commonly used
5. Then `pkg/modules/bank/` - transactions
6. Continue with other modules...
