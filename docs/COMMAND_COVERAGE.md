# Command Coverage Tracking

This document tracks sekaid commands coverage in sekai-cli.

## Summary

| Category | sekaid | CLI | Mapper | Coverage |
|----------|--------|-----|--------|----------|
| Query commands | 98 | 98 | 98 | 100% |
| TX commands | 57 | 57 | 57 | 100% |
| Keys commands | 12 | 4 | 16 | CLI: 33%, Mapper: 100% |
| Node/Daemon commands | ~40 | 0 | 0 | N/A (not applicable) |
| **Total SDK commands** | **167** | **159** | **171** | **100%** |

- **CLI**: Direct command-line commands (`sekai-cli <module> <command>`)
- **Mapper**: Available via scenarios (`sekai-cli scenario run ...`)

## Not Implemented (By Design)

These sekaid commands are **not applicable** to sekai-cli (daemon/node operations):

### Node Operation Commands
- `start` - Start the node daemon
- `init` - Initialize node
- `gentx` - Generate genesis transaction
- `gentx-claim` - Claim validator in genesis
- `collect-gentxs` - Collect genesis transactions
- `add-genesis-account` - Add account to genesis
- `validate-genesis` - Validate genesis file
- `export` - Export state
- `export-metadata` - Export metadata
- `export-minimized-genesis` - Export minimized genesis
- `new-genesis-from-exported` - Create new genesis
- `bootstrap-state` - Bootstrap state
- `reset-state` - Reset state
- `rollback` - Rollback state
- `unsafe-reset-all` - Unsafe reset

### Tendermint/CometBFT Commands
- `tendermint` - Tendermint subcommands
- `show-node-id` - Show node ID
- `show-address` - Show address
- `show-validator` - Show validator info
- `tendermint-validator-set` - Query validator set

### Debug Commands
- `debug` - Debug subcommands
- `addr` - Address conversion
- `prefixes` - List HRP prefixes
- `pubkey` - Decode pubkey
- `pubkey-raw` - Decode raw pubkey
- `raw-bytes` - Convert raw bytes
- `val-address` - Get validator address
- `valcons-address` - Get validator consensus address

### Transaction Signing (Offline)
- `sign` - Sign transaction offline
- `sign-batch` - Sign batch offline
- `multi-sign` - Multi-signature
- `broadcast` - Broadcast signed tx
- `encode` - Encode transaction
- `decode` - Decode transaction
- `validate-signatures` - Validate signatures

### Other
- `config` - CLI configuration
- `help` - Help command
- `version` - Version info
- `testnet` - Setup testnet
- `rosetta` - Rosetta API

## Keys Commands Coverage

| sekaid | CLI | Mapper/Scenario | Status |
|--------|-----|-----------------|--------|
| `keys list` | ✅ | ✅ | Full |
| `keys show` | ✅ | ✅ | Full |
| `keys add` | ✅ | ✅ | Full |
| `keys delete` | ✅ | ✅ | Full |
| `keys export` | ❌ | ✅ | Scenario only |
| `keys import` | ❌ | ✅ | Scenario only |
| `keys import-hex` | ❌ | ✅ | Scenario only |
| `keys rename` | ❌ | ✅ | Scenario only |
| `keys mnemonic` | ❌ | ✅ | Scenario only |
| `keys migrate` | ❌ | ✅ | Scenario only |
| `keys parse` | ❌ | ✅ | Scenario only |
| `keys list-key-types` | ❌ | ✅ | Scenario only |
| `keys get-address` | ❌ | ✅ | Scenario only |
| `keys exists` | ❌ | ✅ | Scenario only |
| `keys create` | ❌ | ✅ | Scenario only |
| `keys recover` | ❌ | ✅ | Scenario only |

**Note**: All 16 keys commands are implemented in the mapper (usable via scenarios).
Only 4 are exposed as direct CLI commands.

## Query Commands - Full Coverage ✅

All query commands are implemented in sekai-cli.

## TX Commands - Full Coverage ✅

All transaction commands are implemented in sekai-cli.

## Notes

1. sekai-cli focuses on **SDK commands** (queries and transactions)
2. Node operation commands are handled by `sekaid` directly or via `scaller`
3. Keys commands have partial coverage - basic operations supported
4. Some commands exist in mapper with different names (aliases)
