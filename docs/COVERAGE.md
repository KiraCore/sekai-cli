# SEKAID vs SEKAI-CLI Coverage Analysis

## Summary

| Status | Count | Percentage |
|--------|-------|------------|
| Implemented | 0 | 0% |
| In Progress | 0 | 0% |
| Not Started | ~150 | 100% |

## Module Coverage Matrix

```
MODULE          │ SEKAID COMMANDS                      │ SEKAI-CLI │ STATUS
────────────────┼──────────────────────────────────────┼───────────┼────────
KEYS            │ add, delete, list, show              │ ⬜ 0/11   │ TODO
                │ export, import, import-hex           │ ⬜        │
                │ rename, migrate, mnemonic, parse     │ ⬜        │
────────────────┼──────────────────────────────────────┼───────────┼────────
BANK            │ send, multi-send                     │ ⬜ 0/2    │ TODO
────────────────┼──────────────────────────────────────┼───────────┼────────
QUERY BANK      │ balances, total, spendable           │ ⬜ 0/5    │ TODO
                │ denom-metadata, send-enabled         │ ⬜        │
────────────────┼──────────────────────────────────────┼───────────┼────────
CUSTOMGOV       │ ~30 query commands                   │ ⬜ 0/30   │ TODO
                │ ~25 tx commands                      │ ⬜ 0/25   │ TODO
────────────────┼──────────────────────────────────────┼───────────┼────────
CUSTOMSTAKING   │ claim-validator, unjail-proposal     │ ⬜ 0/2    │ TODO
                │ query validator(s)                   │ ⬜        │
────────────────┼──────────────────────────────────────┼───────────┼────────
CUSTOMSLASHING  │ activate, pause, unpause             │ ⬜ 0/3    │ TODO
────────────────┼──────────────────────────────────────┼───────────┼────────
MULTISTAKING    │ delegate, undelegate, claim-rewards  │ ⬜ 0/8    │ TODO
                │ claim-undelegation(s), register      │ ⬜        │
                │ set-compound, upsert-pool            │ ⬜        │
────────────────┼──────────────────────────────────────┼───────────┼────────
TOKENS          │ query: rates, all-rates, black-white │ ⬜ 0/6    │ TODO
                │ tx: upsert-rate, proposals           │ ⬜        │
────────────────┼──────────────────────────────────────┼───────────┼────────
SPENDING        │ create, deposit, claim pools         │ ⬜ 0/7    │ TODO
                │ register-beneficiary, proposals      │ ⬜        │
────────────────┼──────────────────────────────────────┼───────────┼────────
UBI             │ proposal-upsert, proposal-remove     │ ⬜ 0/2+   │ TODO
                │ query commands                       │ ⬜        │
────────────────┼──────────────────────────────────────┼───────────┼────────
BASKET          │ mint, burn, swap, claim-rewards      │ ⬜ 0/10   │ TODO
                │ disable-*, proposals                 │ ⬜        │
────────────────┼──────────────────────────────────────┼───────────┼────────
COLLECTIVES     │ create, contribute, donate, withdraw │ ⬜ 0/7    │ TODO
                │ proposals                            │ ⬜        │
────────────────┼──────────────────────────────────────┼───────────┼────────
CUSTODY         │ all commands                         │ ⬜ 0/?    │ TODO
────────────────┼──────────────────────────────────────┼───────────┼────────
BRIDGE          │ cosmos-ethereum, ethereum-cosmos     │ ⬜ 0/2    │ TODO
────────────────┼──────────────────────────────────────┼───────────┼────────
LAYER2          │ all commands                         │ ⬜ 0/?    │ TODO
────────────────┼──────────────────────────────────────┼───────────┼────────
RECOVERY        │ all commands                         │ ⬜ 0/?    │ TODO
────────────────┼──────────────────────────────────────┼───────────┼────────
UPGRADE         │ current-plan, next-plan, proposals   │ ⬜ 0/4    │ TODO
────────────────┼──────────────────────────────────────┼───────────┼────────
STATUS          │ node status                          │ ⬜ 0/1    │ TODO
────────────────┼──────────────────────────────────────┼───────────┼────────
```

Legend:
- ⬜ Not implemented
- 🔄 In progress
- ✅ Complete

## Detailed Command Tracking

### Keys Module (`pkg/modules/keys/`)

| Command | File | Status | Notes |
|---------|------|--------|-------|
| `keys add` | `add.go` | ⬜ | |
| `keys delete` | `delete.go` | ⬜ | |
| `keys list` | `list.go` | ⬜ | |
| `keys show` | `show.go` | ⬜ | |
| `keys export` | `export.go` | ⬜ | |
| `keys import` | `import.go` | ⬜ | |
| `keys import-hex` | `import.go` | ⬜ | |
| `keys rename` | `rename.go` | ⬜ | |
| `keys migrate` | `migrate.go` | ⬜ | |
| `keys mnemonic` | `mnemonic.go` | ⬜ | |
| `keys parse` | `parse.go` | ⬜ | |

### Bank Module (`pkg/modules/bank/`)

| Command | File | Status | Notes |
|---------|------|--------|-------|
| `tx bank send` | `tx_send.go` | ⬜ | |
| `tx bank multi-send` | `tx_multisend.go` | ⬜ | |
| `query bank balances` | `query_balance.go` | ⬜ | |
| `query bank total` | `query_total.go` | ⬜ | |
| `query bank spendable-balances` | `query_spendable.go` | ⬜ | |
| `query bank denom-metadata` | `query_metadata.go` | ⬜ | |
| `query bank send-enabled` | `query_send_enabled.go` | ⬜ | |

### Gov Module (`pkg/modules/gov/`)

| Command | File | Status | Notes |
|---------|------|--------|-------|
| `query customgov proposal` | `query_proposal.go` | ⬜ | |
| `query customgov proposals` | `query_proposals.go` | ⬜ | |
| `query customgov permissions` | `query_permissions.go` | ⬜ | |
| `query customgov roles` | `query_roles.go` | ⬜ | |
| `query customgov role` | `query_role.go` | ⬜ | |
| `query customgov network-properties` | `query_network.go` | ⬜ | |
| `query customgov councilors` | `query_councilors.go` | ⬜ | |
| `query customgov votes` | `query_votes.go` | ⬜ | |
| `query customgov identity-record` | `query_identity.go` | ⬜ | |
| ... | ... | ⬜ | ~20 more query commands |
| `tx customgov proposal vote` | `tx_vote.go` | ⬜ | |
| `tx customgov permission whitelist` | `tx_permission.go` | ⬜ | |
| `tx customgov permission blacklist` | `tx_permission.go` | ⬜ | |
| `tx customgov role create` | `tx_role.go` | ⬜ | |
| `tx customgov councilor claim-seat` | `tx_councilor.go` | ⬜ | |
| ... | ... | ⬜ | ~20 more tx commands |

### Staking Module (`pkg/modules/staking/`)

| Command | File | Status | Notes |
|---------|------|--------|-------|
| `tx customstaking claim-validator-seat` | `tx_claim.go` | ⬜ | |
| `tx customslashing activate` | `tx_activate.go` | ⬜ | |
| `tx customslashing pause` | `tx_pause.go` | ⬜ | |
| `tx customslashing unpause` | `tx_unpause.go` | ⬜ | |
| `query customstaking validator` | `query_validator.go` | ⬜ | |
| `query customstaking validators` | `query_validators.go` | ⬜ | |

### MultiStaking Module (`pkg/modules/multistaking/`)

| Command | File | Status | Notes |
|---------|------|--------|-------|
| `tx multistaking delegate` | `tx_delegate.go` | ⬜ | |
| `tx multistaking undelegate` | `tx_undelegate.go` | ⬜ | |
| `tx multistaking claim-rewards` | `tx_claim.go` | ⬜ | |
| `tx multistaking claim-undelegation` | `tx_claim.go` | ⬜ | |
| `tx multistaking claim-matured-undelegations` | `tx_claim.go` | ⬜ | |
| `tx multistaking register-delegator` | `tx_register.go` | ⬜ | |
| `tx multistaking set-compound-info` | `tx_compound.go` | ⬜ | |
| `tx multistaking upsert-staking-pool` | `tx_pool.go` | ⬜ | |

### Tokens Module (`pkg/modules/tokens/`)

| Command | File | Status | Notes |
|---------|------|--------|-------|
| `query tokens rate` | `query_rate.go` | ⬜ | |
| `query tokens all-rates` | `query_rates.go` | ⬜ | |
| `query tokens rates-by-denom` | `query_rates.go` | ⬜ | |
| `query tokens token-black-whites` | `query_blackwhite.go` | ⬜ | |
| `tx tokens upsert-rate` | `tx_rate.go` | ⬜ | |
| `tx tokens proposal-upsert-rate` | `tx_proposal.go` | ⬜ | |
| `tx tokens proposal-update-tokens-blackwhite` | `tx_proposal.go` | ⬜ | |

### Spending Module (`pkg/modules/spending/`)

| Command | File | Status | Notes |
|---------|------|--------|-------|
| `tx spending create-spending-pool` | `tx_create.go` | ⬜ | |
| `tx spending deposit-spending-pool` | `tx_deposit.go` | ⬜ | |
| `tx spending claim-spending-pool` | `tx_claim.go` | ⬜ | |
| `tx spending register-spending-pool-beneficiary` | `tx_register.go` | ⬜ | |
| `tx spending proposal-spending-pool-distribution` | `tx_proposal.go` | ⬜ | |
| `tx spending proposal-spending-pool-withdraw` | `tx_proposal.go` | ⬜ | |
| `tx spending proposal-update-spending-pool` | `tx_proposal.go` | ⬜ | |

### UBI Module (`pkg/modules/ubi/`)

| Command | File | Status | Notes |
|---------|------|--------|-------|
| `tx ubi proposal-upsert-ubi` | `tx_proposal.go` | ⬜ | |
| `tx ubi proposal-remove-ubi` | `tx_proposal.go` | ⬜ | |
| `query ubi ...` | `query.go` | ⬜ | |

### Basket Module (`pkg/modules/basket/`)

| Command | File | Status | Notes |
|---------|------|--------|-------|
| `tx basket mint-basket-tokens` | `tx_mint.go` | ⬜ | |
| `tx basket burn-basket-tokens` | `tx_burn.go` | ⬜ | |
| `tx basket swap-basket-tokens` | `tx_swap.go` | ⬜ | |
| `tx basket basket-claim-rewards` | `tx_claim.go` | ⬜ | |
| `tx basket disable-basket-deposits` | `tx_disable.go` | ⬜ | |
| `tx basket disable-basket-swaps` | `tx_disable.go` | ⬜ | |
| `tx basket disable-basket-withdraws` | `tx_disable.go` | ⬜ | |
| `tx basket proposal-create-basket` | `tx_proposal.go` | ⬜ | |
| `tx basket proposal-edit-basket` | `tx_proposal.go` | ⬜ | |
| `tx basket proposal-basket-withdraw-surplus` | `tx_proposal.go` | ⬜ | |

### Collectives Module (`pkg/modules/collectives/`)

| Command | File | Status | Notes |
|---------|------|--------|-------|
| `tx collectives create-collective` | `tx_create.go` | ⬜ | |
| `tx collectives contribute-collective` | `tx_contribute.go` | ⬜ | |
| `tx collectives donate-collective` | `tx_donate.go` | ⬜ | |
| `tx collectives withdraw-collective` | `tx_withdraw.go` | ⬜ | |
| `tx collectives proposal-collective-update` | `tx_proposal.go` | ⬜ | |
| `tx collectives proposal-remove-collective` | `tx_proposal.go` | ⬜ | |
| `tx collectives proposal-send-donation` | `tx_proposal.go` | ⬜ | |

### Other Modules

| Module | Status | Priority |
|--------|--------|----------|
| custody | ⬜ Not started | Medium |
| bridge | ⬜ Not started | Low |
| layer2 | ⬜ Not started | Low |
| recovery | ⬜ Not started | Low |
| upgrade | ⬜ Not started | Medium |

## Implementation Priority

### Phase 1: Core (HIGH PRIORITY)
1. `internal/cli` - CLI framework
2. `internal/executor` - Docker executor
3. `internal/config` - Configuration
4. `pkg/modules/keys` - Key management
5. `pkg/modules/bank` - Basic transactions
6. `status` - Node status

### Phase 2: Governance (HIGH PRIORITY)
1. `pkg/modules/gov` - Governance queries
2. `pkg/modules/gov` - Voting
3. `pkg/modules/gov` - Proposals
4. `pkg/modules/gov` - Permissions/Roles

### Phase 3: Staking (MEDIUM PRIORITY)
1. `pkg/modules/staking` - Validator operations
2. `pkg/modules/multistaking` - Delegation

### Phase 4: DeFi (MEDIUM PRIORITY)
1. `pkg/modules/tokens` - Token rates
2. `pkg/modules/spending` - Spending pools
3. `pkg/modules/ubi` - UBI

### Phase 5: Advanced (LOW PRIORITY)
1. `pkg/modules/basket` - Basket tokens
2. `pkg/modules/collectives` - Collectives
3. `pkg/modules/custody` - Custody
4. `pkg/modules/bridge` - Bridge
5. `pkg/modules/layer2` - Layer 2
6. `pkg/modules/recovery` - Recovery
7. `pkg/modules/upgrade` - Upgrades
