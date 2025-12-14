# Scenario Mapper: Command Integration Process

## Purpose
Step-by-step process for adding SDK commands to `pkg/scenarios/mapper.go`. Follow this for each of the 222 commands.

---

## FOR EACH COMMAND - Follow These Steps

### STEP 1: Identify Command Type

```
INPUT: module name, action name
OUTPUT: query OR transaction
```

**Rule: It's a QUERY if:**
- Action starts with: `query-`, `get-`, `list-`, `show-`, `all-`
- Action is one of: `balances`, `balance`, `account`, `accounts`, `params`, `status`, `info`
- Returns data without changing state

**Rule: It's a TRANSACTION if:**
- Action modifies blockchain state
- Requires a signer (`from` parameter)
- Examples: `send`, `delegate`, `create`, `update`, `delete`, `vote`, `claim`

---

### STEP 2: Find SDK Method Signature

```
INPUT: module name, action name
OUTPUT: method signature from pkg/sdk/modules/<module>/<module>.go
```

**Process:**
1. Open `pkg/sdk/modules/<module>/<module>.go`
2. Find method matching action name
3. Note: parameter names, parameter types, return type

**Example:**
```go
// From pkg/sdk/modules/bank/bank.go
func (m *Module) Send(ctx context.Context, from, to string, coins types.Coins, opts *SendOptions) (*sdk.TxResponse, error)
```

**Extract:**
- Params: `from` (string), `to` (string), `coins` (types.Coins), `opts` (*SendOptions)
- Return: `*sdk.TxResponse, error` → TRANSACTION
- Special: `coins` needs `types.ParseCoins()` conversion

---

### STEP 3: Identify Special Handling Needed

```
INPUT: parameter types from Step 2
OUTPUT: list of conversions needed
```

**Check each parameter type:**

| Parameter Type | Conversion Needed | Code Pattern |
|---------------|-------------------|--------------|
| `string` | None | `params["name"]` |
| `int` | `strconv.Atoi()` | `id, _ := strconv.Atoi(params["id"])` |
| `int64` | `strconv.ParseInt()` | `n, _ := strconv.ParseInt(params["num"], 10, 64)` |
| `bool` | `== "true"` | `enabled := params["enabled"] == "true"` |
| `types.Coins` | `types.ParseCoins()` | `coins, _ := types.ParseCoins(params["amount"])` |
| `address (kira1...)` | `resolveAddress()` | `addr, _ := m.resolveAddress(ctx, params["address"])` |
| `[]string` | `strings.Split()` | `items := strings.Split(params["items"], ",")` |
| `*Options` | Build from txOpts | `opts := m.buildXxxOptions(txOpts)` |

**Address Resolution Rule:**
- If param could be key name OR address, use `resolveAddress()`
- Applies to: `address`, `to`, `recipient`, `delegator`, `validator` params
- Does NOT apply to: `from` (signer) - SDK handles this internally

---

### STEP 4: Write the Case Block

```
INPUT: action name, params, conversions from Steps 2-3
OUTPUT: case block code
```

**Template for QUERY:**
```go
case "action-name":
    // 1. Extract required params
    param1 := params["param1"]
    if param1 == "" {
        return nil, nil, fmt.Errorf("module.action-name requires 'param1' parameter")
    }

    // 2. Apply conversions if needed
    // (address resolution, type conversion, etc.)

    // 3. Call SDK method
    result, err := m.moduleMod.MethodName(ctx, param1)

    // 4. Return: (result, nil, err) for queries
    return result, nil, err
```

**Template for TRANSACTION:**
```go
case "action-name":
    // 1. Extract required params
    from := params["from"]
    param1 := params["param1"]
    if from == "" || param1 == "" {
        return nil, nil, fmt.Errorf("module.action-name requires 'from' and 'param1' parameters")
    }

    // 2. Apply conversions if needed
    // (address resolution, coin parsing, etc.)

    // 3. Build options from txOpts
    opts := m.buildModuleOptions(txOpts)

    // 4. Call SDK method
    resp, err := m.moduleMod.MethodName(ctx, from, param1, opts)

    // 5. Return: (resp, resp, err) for transactions
    return resp, resp, err
```

---

### STEP 5: Add to Module Switch (if new module)

```
INPUT: module name
OUTPUT: case in Execute() switch + handler function
```

**If module not in Execute() switch, add:**
```go
case "modulename":
    return m.executeModuleName(ctx, action, params, txOpts)
```

**Create handler function:**
```go
func (m *ActionMapper) executeModuleName(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
    if m.moduleNameMod == nil {
        m.moduleNameMod = modulename.New(m.client)
    }

    switch action {
    // cases go here
    default:
        return nil, nil, fmt.Errorf("unknown modulename action: %s", action)
    }
}
```

**Add module field to ActionMapper struct:**
```go
type ActionMapper struct {
    // ...existing fields...
    moduleNameMod *modulename.Module
}
```

**Add import:**
```go
import "github.com/kiracore/sekai-cli/pkg/sdk/modules/modulename"
```

---

### STEP 6: Add Options Builder (if needed)

```
INPUT: module's TxOptions type
OUTPUT: builder function
```

**Template:**
```go
func (m *ActionMapper) buildModuleNameOptions(txOpts *StepTxOptions) *modulename.TxOptions {
    if txOpts == nil {
        return nil
    }
    return &modulename.TxOptions{
        Fees:          txOpts.Fees,
        Gas:           txOpts.Gas,
        Memo:          txOpts.Memo,
        BroadcastMode: txOpts.BroadcastMode,
    }
}
```

---

### STEP 7: Update Query Detection (if needed)

```
INPUT: new query action names
OUTPUT: update to parser.go queryActions list
```

**In `pkg/scenarios/parser.go`, add to queryActions:**
```go
var queryActions = map[string]bool{
    // existing...
    "new-query-action": true,
}
```

---

## Module Inventory (22 modules, 222 methods)

| Module | Queries | TXs | Total | Mapped | Methods |
|--------|---------|-----|-------|--------|---------|
| auth | 6 | 0 | 6 | ✅ 6/6 | Account, Accounts, ModuleAccount, ModuleAccounts, Params, AddressByAccNum |
| bank | 8 | 2 | 10 | ✅ 10/10 | Balance, Balances, SpendableBalances, TotalSupply, SupplyOf, DenomMetadata, AllDenomsMetadata, SendEnabled, **Send**, **MultiSend** |
| basket | 6 | 10 | 16 | ✅ 16/16 | TokenBaskets, TokenBasketByID, TokenBasketByDenom, HistoricalMints, HistoricalBurns, HistoricalSwaps, **MintBasketTokens**, **BurnBasketTokens**, **SwapBasketTokens**, **BasketClaimRewards**, **DisableBasketDeposits**, **DisableBasketWithdraws**, **DisableBasketSwaps**, **ProposalCreateBasket**, **ProposalEditBasket**, **ProposalWithdrawSurplus** |
| bridge | 2 | 2 | 4 | ✅ 4/4 | GetCosmosEthereum, GetEthereumCosmos, **ChangeCosmosEthereum**, **ChangeEthereumCosmos** |
| collectives | 4 | 7 | 11 | ✅ 11/11 | Collectives, Collective, CollectivesByAccount, CollectivesProposals, **CreateCollective**, **ContributeCollective**, **DonateCollective**, **WithdrawCollective**, **ProposalCollectiveUpdate**, **ProposalRemoveCollective**, **ProposalSendDonation** |
| custody | 5 | 0 | 5 | ✅ 5/5 | Get, Custodians, CustodiansPool, Whitelist, Limits |
| distributor | 5 | 0 | 5 | ✅ 5/5 | FeesTreasury, PeriodicSnapshot, SnapshotPeriod, SnapshotPeriodPerformance, YearStartSnapshot |
| ethereum | 1 | 0 | 1 | ✅ 1/1 | State |
| evidence | 2 | 0 | 2 | ✅ 2/2 | AllEvidence, Evidence |
| gov | 41 | 41 | 82 | ✅ 82/82 | NetworkProperties, Proposals, Proposal, AllRoles, Role, Roles, Permissions, Councilors, Votes, Vote, Voters, ExecutionFee, AllExecutionFees, IdentityRecord, IdentityRecords, IdentityRecordsByAddress, DataRegistry, DataRegistryKeys, Polls, PollVotes, PoorNetworkMessages, CustomPrefixes, AllProposalDurations, ProposalDuration, NonCouncilors, CouncilRegistry, WhitelistedPermissionAddresses, BlacklistedPermissionAddresses, WhitelistedRoleAddresses, ProposerVotersCount, AllIdentityRecordVerifyRequests, IdentityRecordVerifyRequest, IdentityRecordVerifyRequestsByApprover, IdentityRecordVerifyRequestsByRequester, **VoteProposal**, **CouncilorClaimSeat**, **CouncilorActivate**, **CouncilorPause**, **CouncilorUnpause**, **PermissionWhitelist**, **PermissionBlacklist**, **PermissionRemoveWhitelisted**, **PermissionRemoveBlacklisted**, **RoleCreate**, **RoleAssign**, **RoleUnassign**, **RoleWhitelistPermission**, **RoleBlacklistPermission**, **RoleRemoveWhitelistedPermission**, **RoleRemoveBlacklistedPermission**, **PollCreate**, **PollVote**, **SetNetworkProperties**, **SetExecutionFee**, **RegisterIdentityRecords**, **DeleteIdentityRecords**, **RequestIdentityRecordVerify**, **HandleIdentityRecordsVerifyRequest**, **CancelIdentityRecordsVerifyRequest**, **ProposalAssignRole**, **ProposalUnassignRole**, **ProposalWhitelistPermission**, **ProposalBlacklistPermission**, **ProposalRemoveWhitelistedPermission**, **ProposalRemoveBlacklistedPermission**, **ProposalCreateRole**, **ProposalRemoveRole**, **ProposalWhitelistRolePermission**, **ProposalBlacklistRolePermission**, **ProposalRemoveWhitelistedRolePermission**, **ProposalRemoveBlacklistedRolePermission**, **ProposalSetNetworkProperty**, **ProposalSetPoorNetworkMsgs**, **ProposalSetExecutionFees**, **ProposalUpsertDataRegistry**, **ProposalSetProposalDurations**, **ProposalJailCouncilor**, **ProposalResetWholeCouncilorRank**, **ProposalWhitelistAccountPermission**, **ProposalBlacklistAccountPermission**, **ProposalRemoveWhitelistedAccountPermission**, **ProposalRemoveBlacklistedAccountPermission** |
| keys | 16 | 0 | 16 | ✅ 16/16 | Add, Delete, List, Show, Export, Import, Rename, Mnemonic, GetAddress, Exists, Create, Recover, ImportHex, ListKeyTypes, Migrate, Parse |
| layer2 | 3 | 0 | 3 | ✅ 3/3 | AllDapps, ExecutionRegistrar, TransferDapps |
| multistaking | 5 | 8 | 13 | ✅ 13/13 | Pools, Undelegations, OutstandingRewards, CompoundInfo, StakingPoolDelegators, **Delegate**, **Undelegate**, **ClaimRewards**, **ClaimUndelegation**, **ClaimMaturedUndelegations**, **RegisterDelegator**, **SetCompoundInfo**, **UpsertStakingPool** |
| params | 1 | 0 | 1 | ✅ 1/1 | Subspace |
| recovery | 4 | 0 | 4 | ✅ 4/4 | RecoveryRecord, RecoveryToken, RRHolderRewards, RRHolders |
| slashing | 6 | 0 | 6 | ✅ 6/6 | SigningInfo, SigningInfos, ActiveStakingPools, InactiveStakingPools, SlashedStakingPools, SlashProposals |
| spending | 4 | 8 | 12 | ✅ 12/12 | PoolNames, PoolByName, PoolProposals, PoolsByAccount, **ClaimSpendingPool**, **DepositSpendingPool**, **CreateSpendingPool**, **RegisterSpendingPoolBeneficiary**, **ProposalUpdateSpendingPool**, **ProposalSpendingPoolDistribution**, **ProposalSpendingPoolWithdraw** |
| staking | 2 | 2 | 4 | ✅ 4/4 | Validators, Validator, **ClaimValidatorSeat**, **ProposalUnjailValidator** |
| status | 8 | 0 | 8 | ✅ 8/8 | Status, NodeInfo, SyncInfo, ValidatorInfo, ChainID, LatestBlockHeight, IsSyncing, NetworkProperties |
| tokens | 4 | 3 | 7 | ✅ 7/7 | AllRates, Rate, RatesByDenom, TokenBlackWhites, **UpsertRate**, **ProposalUpsertRate**, **ProposalUpdateTokensBlackWhite** |
| ubi | 2 | 2 | 4 | ✅ 4/4 | Records, RecordByName, **ProposalUpsertUBI**, **ProposalRemoveUBI** |
| upgrade | 2 | 2 | 4 | ✅ 4/4 | CurrentPlan, NextPlan, **ProposalSetPlan**, **ProposalCancelPlan** |
| **TOTAL** | **134** | **88** | **222** | **222/222** | |

**Progress: 222/222 (100%) - Complete**

**Legend:** Regular = Query, **Bold** = Transaction

---

## Common Patterns Reference

**Address that might be key name:**
```go
addrParam := params["address"]
address, err := m.resolveAddress(ctx, addrParam)
if err != nil {
    return nil, nil, err
}
```

**Coin amount parsing:**
```go
coins, err := types.ParseCoins(params["amount"])
if err != nil {
    return nil, nil, fmt.Errorf("invalid amount '%s': %w", params["amount"], err)
}
```

**Integer parameter:**
```go
idStr := params["id"]
id, err := strconv.Atoi(idStr)
if err != nil {
    return nil, nil, fmt.Errorf("invalid id '%s': must be integer", idStr)
}
```

**Optional parameter with default:**
```go
limit := params["limit"]
if limit == "" {
    limit = "100"
}
```

**Multiple required params validation:**
```go
from, to, amount := params["from"], params["to"], params["amount"]
if from == "" || to == "" || amount == "" {
    return nil, nil, fmt.Errorf("bank.send requires 'from', 'to', and 'amount' parameters")
}
```

---

## Execution Checklist Per Command

- [ ] Identified type (query/transaction)
- [ ] Found SDK method signature
- [ ] Listed parameter conversions needed
- [ ] Wrote case block with validation
- [ ] Added module handler (if new module)
- [ ] Added options builder (if needed)
- [ ] Updated queryActions (if query)
- [ ] Tested with dry-run
- [ ] Tested actual execution
