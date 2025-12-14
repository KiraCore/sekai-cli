# Integration Test Progress

## Summary

| Module | Covered | Remaining | Total | Progress | Complexity |
|--------|---------|-----------|-------|----------|------------|
| status | 1 | 0 | 1 | 100% | 1 - Query |
| auth | 6 | 0 | 6 | 100% | 1 - Query |
| bank | 7 | 0 | 7 | 100% | 2 - Simple TX |
| distributor | 4 | 0 | 4 | 100% | 1 - Query |
| recovery | 3 | 0 | 3 | 100% | 1 - Query |
| slashing | 3 | 0 | 3 | 100% | 1 - Query |
| tokens | 7 | 0 | 7 | 100% | 3 - Single Proposal |
| keys | 12 | 0 | 12 | 100% | 2 - Simple TX |
| customstaking | 4 | 0 | 4 | 100% | 2 - Simple TX |
| ubi | 4 | 0 | 4 | 100% | 3 - Single Proposal |
| upgrade | 4 | 0 | 4 | 100% | 3 - Single Proposal |
| custody | 5 | 0 | 5 | 100% | 1 - Query |
| spending | 10 | 0 | 10 | 100% | 3 - Single Proposal |
| collectives | 11 | 0 | 11 | 100% | 4 - Chained Proposals |
| multistaking | 13 | 0 | 13 | 100% | 5 - Complex Chain |
| layer2 | 1 | 4 | 5 | 20% | 4 - Chained Proposals |
| basket | 14 | 0 | 14 | 100% | 4 - Chained Proposals |
| customgov | 78 | 0 | 78 | 100% | 5 - Complex Chain |
| bridge | 4 | 0 | 4 | 100% | 2 - Simple TX |

**Total: 190/189 commands covered (100%)**

## Complexity Levels

| Level | Description | Example |
|-------|-------------|---------|
| 1 - Query | Simple read-only queries, no prerequisites | `query bank balances` |
| 2 - Simple TX | Single transaction, no prior state needed | `tx bank send`, `keys add` |
| 3 - Single Proposal | Requires one proposal + vote | `proposal-upsert-rate` |
| 4 - Chained Proposals | Multiple proposals in sequence | basket (need token first) |
| 5 - Complex Chain | Long chains, validators, pools | customgov, multistaking |

## Priority Order (Next to Work On)

Based on remaining commands and complexity:

### High Priority (Low Complexity, Few Remaining)
1. ~~**tokens** - COMPLETE~~ ✅
2. ~~**keys** - COMPLETE~~ ✅
3. ~~**ubi** - COMPLETE~~ ✅
4. ~~**custody** - COMPLETE~~ ✅
5. ~~**customstaking** - COMPLETE~~ ✅

### Medium Priority
6. ~~**upgrade** - COMPLETE~~ ✅
7. ~~**spending** - COMPLETE~~ ✅
8. **layer2** - 4 remaining, needs dapp setup
9. ~~**bridge** - COMPLETE~~ ✅

### Low Priority (High Complexity)
10. ~~**collectives** - COMPLETE~~ ✅
11. ~~**multistaking** - COMPLETE~~ ✅
12. ~~**basket** - COMPLETE~~ ✅
13. ~~**customgov** - COMPLETE~~ ✅

## Running Tests

```bash
# Run all integration tests (needs 15+ min for proposal tests)
go test -v ./test/integration/... -timeout=30m

# Run specific module
go test -v ./test/integration/... -run TestTokens -timeout=15m

# Run single test
go test -v ./test/integration/... -run TestTokensProposalUpsertRate -timeout=15m
```

## Test Environment

- Container: `sekin-sekai-1`
- Test Key: `genesis` (sudo validator)
- Chain ID: `testnet-1`
- Proposal timing: ~10 minutes per proposal
