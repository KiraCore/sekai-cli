# SEKAI Command Tree Reference

This document provides a comprehensive tree structure of all available commands and their flags for the `sekaid` daemon in the sekin-sekai-1 container.

## Main Command

```
sekaid [command]
```

### Global Flags
- `--home string` - Directory for config and data (default "/.sekaid")
- `--log_format string` - The logging format (json|plain) (default "plain")
- `--log_level string` - The logging level (trace|debug|info|warn|error|fatal|panic) (default "info")
- `--trace` - Print out full stack trace on errors
- `-h, --help` - Help for sekaid

## Complete Command Tree Structure

```
sekaid
├── add-genesis-account       # Add a genesis account to genesis.json
├── collect-gentxs            # Collect genesis txs and output a genesis.json file
├── config                    # Create or query an application CLI configuration file
├── debug                     # Tool for helping with debugging your application
│   ├── addr                  # Convert an address between hex and bech32
│   ├── prefixes              # List prefixes used for HRP in Bech32
│   ├── pubkey                # Decode a pubkey from proto JSON
│   ├── pubkey-raw            # Decode a ED25519 or secp256k1 pubkey
│   └── raw-bytes             # Convert raw bytes output to hex
├── export                    # Export state to JSON
├── export-metadata           # Get metadata for client interaction to the node
├── export-minimized-genesis  # Get minimized genesis from genesis with spaces
├── gentx                     # Generate a genesis tx carrying a self delegation
├── gentx-claim               # Adds validator into the genesis set
├── help                      # Help about any command
├── init                      # Initialize private validator, p2p, genesis, and application configuration files
├── keys                      # Manage your application's keys
│   ├── add                   # Add an encrypted private key
│   ├── delete                # Delete the given keys
│   ├── export                # Export private keys
│   ├── import                # Import private keys into the local keybase
│   ├── import-hex            # Import private keys into the local keybase
│   ├── list                  # List all keys
│   ├── list-key-types        # List all key types
│   ├── migrate               # Migrate keys from amino to proto serialization format
│   ├── mnemonic              # Compute the bip39 mnemonic for some input entropy
│   ├── parse                 # Parse address from hex to bech32 and vice versa
│   ├── rename                # Rename an existing key
│   └── show                  # Retrieve key information by name or address
├── new-genesis-from-exported # Get new genesis from exported app state json
├── query (q)                 # Querying subcommands
│   ├── account               # Query for account by address
│   ├── auth                  # Querying commands for the auth module
│   │   ├── account           # Query for account by address
│   │   ├── accounts          # Query all the accounts
│   │   ├── address-by-acc-num # Query for an address by account number
│   │   ├── module-account    # Query module account info by module name
│   │   ├── module-accounts   # Query all module accounts
│   │   └── params            # Query the current auth parameters
│   ├── bank                  # Querying commands for the bank module
│   │   ├── balances          # Query for account balances by address
│   │   ├── denom-metadata    # Query the client metadata for coin denominations
│   │   ├── send-enabled      # Query for send enabled entries
│   │   ├── spendable-balances # Query for account spendable balances by address
│   │   └── total             # Query the total supply of coins of the chain
│   ├── basket                # Query commands for the basket module
│   ├── block                 # Get verified data for the block at given height
│   ├── bridge                # Query commands for the bridge module
│   ├── collectives           # Query commands for the collectives module
│   ├── custody               # Query commands for the custody module
│   ├── customevidence        # Query for evidence by hash or for all submitted evidence
│   ├── customgov             # Query commands for the customgov module
│   │   ├── all-execution-fees                           # Query all execution fees
│   │   ├── all-identity-record-verify-requests          # Query all identity records verify requests
│   │   ├── all-proposal-durations                       # Query all proposal durations
│   │   ├── all-roles                                    # Query all registered roles
│   │   ├── blacklisted-permission-addresses             # Query all KIRA addresses by a specific blacklisted permission
│   │   ├── council-registry                             # Query governance registry
│   │   ├── councilors                                   # Query councilors
│   │   ├── custom-prefixes                              # Query custom prefixes
│   │   ├── data-registry                                # Query data registry by specific key
│   │   ├── data-registry-keys                           # Query all data registry keys
│   │   ├── execution-fee                                # Query execution fee by the type of transaction
│   │   ├── identity-record                              # Query identity record by id
│   │   ├── identity-record-verify-request               # Query identity record verify request by id
│   │   ├── identity-record-verify-requests-by-approver  # Query identity record verify request by approver
│   │   ├── identity-record-verify-requests-by-requester # Query identity records verify requests by requester
│   │   ├── identity-records                             # Query all identity records
│   │   ├── identity-records-by-addr                     # Query identity records by address
│   │   ├── network-properties                           # Query network properties
│   │   ├── non-councilors                               # Query all governance members that are NOT Councilors
│   │   ├── permissions                                  # Query permissions of an address
│   │   ├── poll-votes                                   # Get poll votes by id
│   │   ├── polls                                        # Get polls by address
│   │   ├── poor-network-messages                        # Query poor network messages
│   │   ├── proposal                                     # Query proposal details
│   │   ├── proposal-duration                            # Query a proposal duration
│   │   ├── proposals                                    # Query proposals with optional filters
│   │   ├── proposer-voters-count                        # Query proposer and voters count
│   │   ├── role                                         # Query role by sid or id
│   │   ├── roles                                        # Query roles assigned to an address
│   │   ├── vote                                         # Query details of a single vote
│   │   ├── voters                                       # Query voters of a proposal
│   │   ├── votes                                        # Query votes on a proposal
│   │   ├── whitelisted-permission-addresses             # Query all KIRA addresses by a specific whitelisted permission
│   │   └── whitelisted-role-addresses                   # Query all kira addresses by a specific whitelisted role
│   ├── customslashing        # Querying commands for the slashing module
│   ├── customstaking         # Querying commands for the staking module
│   │   ├── validator         # Query a validator
│   │   └── validators        # Query validators with filters
│   ├── distributor           # Query commands for the distributor module
│   ├── ethereum              # Query commands for the ethereum module
│   ├── layer2                # Query commands for the layer2 module
│   ├── multistaking          # Querying commands for the multistaking module
│   ├── params                # Querying commands for the params module
│   ├── recovery              # Querying commands for the recovery module
│   ├── spending              # Query commands for the spending module
│   ├── tendermint-validator-set # Get the full tendermint validator set at given height
│   ├── tokens                # Query commands for the tokens module
│   │   ├── all-rates         # Get all token rates
│   │   ├── rate              # Get the token rate by denom
│   │   ├── rates-by-denom    # Get token rates by denom
│   │   └── token-black-whites # Get token black whites
│   ├── tx                    # Query for a transaction by hash
│   ├── txs                   # Query for paginated transactions
│   ├── ubi                   # Query commands for the ubi module
│   └── upgrade               # Querying commands for the upgrade module
│       ├── current-plan      # Get the current plan
│       └── next-plan         # Get the next plan
├── rollback                  # Rollback cosmos-sdk and tendermint state by one height
├── rosetta                   # Spin up a rosetta server
├── start                     # Run the full node
├── status                    # Query remote node for status
├── tendermint (comet, cometbft) # Tendermint subcommands
│   ├── bootstrap-state       # Bootstrap CometBFT state at an arbitrary block height
│   ├── reset-state           # Remove all the data and WAL
│   ├── show-address          # Shows this node's tendermint validator consensus address
│   ├── show-node-id          # Show this node's ID
│   ├── show-validator        # Show this node's tendermint validator info
│   ├── unsafe-reset-all      # (unsafe) Remove all data and WAL, reset to genesis
│   └── version               # Print tendermint libraries' version
├── testnet                   # Initialize files for a Sekaid testnet
├── tx                        # Transactions subcommands
│   ├── bank                  # Bank transaction subcommands
│   │   ├── multi-send        # Send funds from one account to two or more accounts
│   │   └── send              # Send funds from one account to another
│   ├── basket                # Basket sub commands
│   │   ├── basket-claim-rewards             # Force staking derivative basket to claim rewards
│   │   ├── burn-basket-tokens               # Burn basket tokens
│   │   ├── disable-basket-deposits          # Emergency function to disable deposits
│   │   ├── disable-basket-swaps             # Emergency function to disable swaps
│   │   ├── disable-basket-withdraws         # Emergency function to disable withdraws
│   │   ├── mint-basket-tokens               # Mint basket tokens
│   │   ├── proposal-basket-withdraw-surplus # Create a proposal to withdraw surplus
│   │   ├── proposal-create-basket           # Create a proposal to create a basket
│   │   ├── proposal-edit-basket             # Create a proposal to edit a basket
│   │   └── swap-basket-tokens               # Swap basket tokens
│   ├── bridge                # Bridge sub commands
│   │   ├── change_cosmos_ethereum # Create new change request from Cosmos to Ethereum
│   │   └── change_ethereum_cosmos # Create new change request from Ethereum to Cosmos
│   ├── broadcast             # Broadcast transactions generated offline
│   ├── collectives           # Collectives sub commands
│   │   ├── contribute-collective      # Put bonds on collective
│   │   ├── create-collective          # Create collective
│   │   ├── donate-collective          # Set lock and donation for bonds
│   │   ├── proposal-collective-update # Create a proposal to update collective
│   │   ├── proposal-remove-collective # Create a proposal to withdraw collective
│   │   ├── proposal-send-donation     # Create a proposal to withdraw donation
│   │   └── withdraw-collective        # Withdraw tokens from collective
│   ├── custody               # Custody sub commands
│   ├── customevidence        # Evidence transaction subcommands
│   ├── customgov             # Custom gov sub commands
│   │   ├── cancel-identity-records-verify-request # Cancel identity records verification request
│   │   ├── councilor         # Councilor subcommands
│   │   │   ├── activate      # Activate councilor
│   │   │   ├── claim-seat    # Claim councilor seat
│   │   │   ├── pause         # Pause councilor
│   │   │   └── unpause       # Unpause councilor
│   │   ├── delete-identity-records                # Delete identity records
│   │   ├── handle-identity-records-verify-request # Approve or reject identity records verify request
│   │   ├── permission        # Permission management subcommands
│   │   │   ├── blacklist          # Assign permission to a kira account blacklist
│   │   │   ├── remove-blacklisted # Remove blacklisted permission from an address
│   │   │   ├── remove-whitelisted # Remove whitelisted permission from an address
│   │   │   └── whitelist          # Assign permission to a kira address whitelist
│   │   ├── poll              # Governance poll management subcommands
│   │   │   ├── create        # Create a poll
│   │   │   └── vote          # Vote a poll
│   │   ├── proposal          # Governance proposals management subcommands
│   │   │   ├── account       # Account proposals management subcommands
│   │   │   │   ├── assign-role                   # Create proposal to assign role
│   │   │   │   ├── blacklist-permission          # Create proposal to blacklist permission
│   │   │   │   ├── remove-blacklisted-permission # Create proposal to remove blacklisted permission
│   │   │   │   ├── remove-whitelisted-permission # Create proposal to remove whitelisted permission
│   │   │   │   ├── unassign-role                 # Create proposal to unassign role
│   │   │   │   └── whitelist-permission          # Create proposal to whitelist permission
│   │   │   ├── proposal-jail-councilor             # Create proposal to jail councilors
│   │   │   ├── proposal-reset-whole-councilor-rank # Create proposal to reset whole councilor rank
│   │   │   ├── proposal-set-execution-fees         # Create proposal to set execution fees
│   │   │   ├── role          # Role proposals management subcommands
│   │   │   │   ├── blacklist-permission          # Propose to blacklist permission for role
│   │   │   │   ├── create                        # Propose to create new role
│   │   │   │   ├── remove                        # Propose to remove role
│   │   │   │   ├── remove-blacklisted-permission # Propose to remove blacklisted permission
│   │   │   │   ├── remove-whitelisted-permission # Propose to remove whitelisted permission
│   │   │   │   └── whitelist-permission          # Propose to whitelist permission for role
│   │   │   ├── set-network-property                # Create proposal to set network property
│   │   │   ├── set-poor-network-msgs               # Create proposal to set poor network msgs
│   │   │   ├── set-proposal-durations-proposal     # Create proposal to set batch proposal durations
│   │   │   ├── upsert-data-registry                # Create proposal to upsert key in data registry
│   │   │   └── vote          # Vote a proposal
│   │   ├── register-identity-records              # Create an identity record
│   │   ├── request-identity-record-verify         # Request an identity verify record
│   │   ├── role              # Role management subcommands
│   │   │   ├── assign                        # Assign role to account
│   │   │   ├── blacklist-permission          # Blacklist permission for governance role
│   │   │   ├── create                        # Create new role
│   │   │   ├── remove-blacklisted-permission # Remove blacklisted permission from role
│   │   │   ├── remove-whitelisted-permission # Remove whitelisted permission from role
│   │   │   ├── unassign                      # Unassign role from account
│   │   │   └── whitelist-permission          # Whitelist permission to role
│   │   ├── set-execution-fee                      # Set execution fee
│   │   └── set-network-properties                 # Set network properties
│   ├── customslashing        # Slashing transaction subcommands
│   ├── customstaking         # Staking module subcommands
│   │   ├── claim-validator-seat # Claim validator seat to become a Validator
│   │   └── proposal          # Proposal subcommands
│   │       └── unjail-validator # Create proposal to unjail validator
│   ├── decode                # Decode a binary encoded transaction string
│   ├── distributor           # Distributor sub commands
│   ├── encode                # Encode transactions generated offline
│   ├── ethereum              # Ethereum sub commands
│   ├── layer2                # Layer2 sub commands
│   ├── multi-sign            # Generate multisig signatures for transactions
│   ├── multistaking          # Multistaking sub commands
│   │   ├── claim-matured-undelegations # Claim all matured undelegations
│   │   ├── claim-rewards               # Claim rewards from a pool
│   │   ├── claim-undelegation          # Claim matured undelegation
│   │   ├── delegate                    # Delegate to a pool
│   │   ├── register-delegator          # Register a delegator
│   │   ├── set-compound-info           # Set compound info
│   │   ├── undelegate                  # Start undelegation from a pool
│   │   └── upsert-staking-pool         # Upsert staking pool
│   ├── recovery              # Recovery transaction subcommands
│   ├── sign                  # Sign a transaction generated offline
│   ├── sign-batch            # Sign transaction batch files
│   ├── spending              # Spending sub commands
│   │   ├── claim-spending-pool                 # Claim spending pool
│   │   ├── create-spending-pool                # Create spending pool
│   │   ├── deposit-spending-pool               # Deposit spending pool
│   │   ├── proposal-spending-pool-distribution # Create proposal to distribute spending pool
│   │   ├── proposal-spending-pool-withdraw     # Create proposal to withdraw spending pool
│   │   ├── proposal-update-spending-pool       # Create proposal to update spending pool
│   │   └── register-spending-pool-beneficiary  # Register spending pool beneficiary
│   ├── tokens                # Tokens sub commands
│   │   ├── proposal-update-tokens-blackwhite # Create proposal to update whitelisted/blacklisted tokens
│   │   ├── proposal-upsert-rate              # Create proposal to upsert token rate
│   │   └── upsert-rate                       # Upsert token rate
│   ├── ubi                   # UBI sub commands
│   │   ├── proposal-remove-ubi # Create proposal to remove ubi
│   │   └── proposal-upsert-ubi # Create proposal to upsert ubi
│   ├── upgrade               # Upgrade transaction subcommands
│   │   ├── proposal-cancel-plan # Create proposal to cancel upgrade plan
│   │   └── proposal-set-plan    # Create proposal to set upgrade plan
│   └── validate-signatures   # Validate transactions signatures
├── val-address               # Get validator address from account address
├── valcons-address           # Get validator consensus address from account address
├── validate-genesis          # Validates the genesis file
└── version                   # Print the application binary version information
```

## Detailed Command Flags Reference

### Key Management Commands

#### `keys add <name>`
**Flags:**
- `--account uint32` - Account number for HD derivation
- `--coin-type uint32` - Coin type number for HD derivation (default 118)
- `--dry-run` - Perform action, but don't add key to local keystore
- `--hd-path string` - Manual HD Path derivation (overrides BIP44 config)
- `--index uint32` - Address index number for HD derivation
- `-i, --interactive` - Interactively prompt user for BIP39 passphrase and mnemonic
- `--key-type string` - Key signing algorithm (default "secp256k1")
- `--ledger` - Store a local reference to a private key on a Ledger device
- `--multisig strings` - List of key names for public legacy multisig key
- `--multisig-threshold int` - K out of N required signatures (default 1)
- `--no-backup` - Don't print out seed phrase
- `--nosort` - Keys passed to --multisig are taken in the order supplied
- `--pubkey string` - Parse a public key in JSON format
- `--recover` - Provide seed phrase to recover existing key

#### `keys show [name_or_address]`
**Flags:**
- `-a, --address` - Output the address only
- `--bech string` - The Bech32 prefix encoding (acc|val|cons) (default "acc")
- `-d, --device` - Output the address in a ledger device
- `--multisig-threshold int` - K out of N required signatures (default 1)
- `-p, --pubkey` - Output the public key only

### Node Operation Commands

#### `start`
**All Available Flags:**
- **API Configuration:**
  - `--api.enable` - Define if the API server should be enabled
  - `--api.address string` - API server address (default "tcp://localhost:1317")
  - `--api.swagger` - Define if swagger documentation should be registered
  - `--api.enabled-unsafe-cors` - Define if CORS should be enabled (unsafe)
  - `--api.max-open-connections uint` - Max open connections (default 1000)
  - `--api.rpc-max-body-bytes uint` - Tendermint maximum request body (default 1000000)
  - `--api.rpc-read-timeout uint` - Tendermint RPC read timeout in seconds (default 10)
  - `--api.rpc-write-timeout uint` - Tendermint RPC write timeout in seconds

- **gRPC Configuration:**
  - `--grpc.enable` - Define if the gRPC server should be enabled (default true)
  - `--grpc.address string` - gRPC server address (default "localhost:9090")
  - `--grpc-web.enable` - Define if gRPC-Web server should be enabled (default true)
  - `--grpc-web.address string` - gRPC-Web server address (default "localhost:9091")
  - `--grpc-only` - Start in gRPC query only mode (no Tendermint)

- **P2P Network:**
  - `--p2p.laddr string` - Node listen address (default "tcp://0.0.0.0:26656")
  - `--p2p.external-address string` - IP:port address to advertise to peers
  - `--p2p.seeds string` - Comma-delimited ID@host:port seed nodes
  - `--p2p.persistent_peers string` - Comma-delimited ID@host:port persistent peers
  - `--p2p.pex` - Enable/disable Peer-Exchange (default true)
  - `--p2p.seed_mode` - Enable/disable seed mode
  - `--p2p.private_peer_ids string` - Comma-delimited private peer IDs
  - `--p2p.unconditional_peer_ids string` - Comma-delimited unconditional peer IDs
  - `--p2p.upnp` - Enable/disable UPNP port forwarding

- **RPC Configuration:**
  - `--rpc.laddr string` - RPC listen address (default "tcp://127.0.0.1:26657")
  - `--rpc.grpc_laddr string` - GRPC listen address (BroadcastTx only)
  - `--rpc.unsafe` - Enable unsafe rpc methods
  - `--rpc.pprof_laddr string` - pprof listen address

- **Consensus:**
  - `--consensus.create_empty_blocks` - Produce blocks when there are txs or AppHash changes (default true)
  - `--consensus.create_empty_blocks_interval string` - Interval between empty blocks (default "0s")
  - `--consensus.double_sign_check_height int` - Blocks to check for double signing

- **State Management:**
  - `--pruning string` - Pruning strategy (default|nothing|everything|custom) (default "default")
  - `--pruning-keep-recent uint` - Number of recent heights to keep (custom pruning)
  - `--pruning-interval uint` - Height interval at which pruned heights are removed (custom pruning)
  - `--min-retain-blocks uint` - Minimum block height offset during ABCI commit

- **Database:**
  - `--db_backend string` - Database backend: goleveldb|cleveldb|boltdb|rocksdb|badgerdb (default "goleveldb")
  - `--db_dir string` - Database directory (default "data")

- **Performance:**
  - `--iavl-disable-fastnode` - Disable fast node for IAVL tree
  - `--inter-block-cache` - Enable inter-block caching (default true)
  - `--inv-check-period uint` - Assert registered invariants every N blocks
  - `--cpu-profile string` - Enable CPU profiling and write to file
  - `--trace-store string` - Enable KVStore tracing to output file

- **Other:**
  - `--abci string` - Specify abci transport (socket|grpc) (default "socket")
  - `--address string` - Listen address (default "tcp://0.0.0.0:26658")
  - `--block_sync` - Sync blockchain using blocksync algorithm (default true)
  - `--genesis_hash bytesHex` - Optional SHA-256 hash of genesis file
  - `--halt-height uint` - Block height at which to gracefully halt
  - `--halt-time uint` - Minimum block time to gracefully halt
  - `--mempool.max-txs int` - Sets MaxTx value for app-side mempool
  - `--minimum-gas-prices string` - Minimum gas prices to accept for transactions
  - `--moniker string` - Node name (default "sekai.local")
  - `--priv_validator_laddr string` - Socket address for external priv_validator
  - `--proxy_app string` - Proxy app address (default "tcp://127.0.0.1:26658")
  - `--state-sync.snapshot-interval uint` - State sync snapshot interval
  - `--state-sync.snapshot-keep-recent uint32` - State sync snapshots to keep (default 2)
  - `--transport string` - Transport protocol: socket, grpc (default "socket")
  - `--unsafe-skip-upgrades ints` - Skip upgrade heights to continue old binary
  - `--with-tendermint` - Run abci app embedded with tendermint (default true)
  - `--x-crisis-skip-assert-invariants` - Skip x/crisis invariants check on startup

### Transaction Common Flags

All transaction commands support these flags:
- `--chain-id string` - The network chain ID
- `--from string` - Name or address of private key with which to sign
- `--fees string` - Fees to pay along with transaction
- `--gas string` - Gas limit to set per-transaction
- `--gas-prices string` - Gas prices in decimal format
- `--gas-adjustment float` - Gas adjustment factor
- `--broadcast-mode string` - Transaction broadcasting mode (sync|async)
- `--dry-run` - Perform simulation without broadcasting
- `--generate-only` - Build unsigned transaction
- `--offline` - Offline mode
- `-y, --yes` - Skip tx broadcasting prompt confirmation
- `--keyring-backend string` - Select keyring's backend
- `--keyring-dir string` - The client Keyring directory
- `--node string` - <host>:<port> to tendermint rpc interface
- `--note string` - Note to add a description to the transaction
- `-o, --output string` - Output format (text|json)
- `--sign-mode string` - Choose sign mode (direct|amino-json|direct-aux)
- `--timeout-height uint` - Set a block timeout height
- `--ledger` - Use a connected Ledger device
- `-a, --account-number uint` - Account number of signing account
- `-s, --sequence uint` - Sequence number of signing account
- `--aux` - Generate aux signer data
- `--fee-granter string` - Fee granter grants fees for the transaction
- `--fee-payer string` - Fee payer pays fees for the transaction
- `--tip string` - Tip amount for target chain

---

## Usage Examples

### Initialize a new node
```bash
docker exec sekin-sekai-1 /sekaid init my-node --chain-id testnet-1
```

### Create and manage keys
```bash
# Add a new key
docker exec sekin-sekai-1 /sekaid keys add mykey

# List all keys
docker exec sekin-sekai-1 /sekaid keys list

# Show key details
docker exec sekin-sekai-1 /sekaid keys show mykey
```

### Start the node with specific configuration
```bash
docker exec sekin-sekai-1 /sekaid start \
  --api.enable \
  --grpc.enable \
  --p2p.seeds "id1@host1:26656,id2@host2:26656" \
  --minimum-gas-prices "0.001ukex"
```

### Query operations
```bash
# Query account balance
docker exec sekin-sekai-1 /sekaid query bank balances <address>

# Query validator
docker exec sekin-sekai-1 /sekaid query customstaking validator <address>

# Query governance proposals
docker exec sekin-sekai-1 /sekaid query customgov proposals
```

### Transaction operations
```bash
# Send tokens
docker exec sekin-sekai-1 /sekaid tx bank send <from> <to> <amount> \
  --chain-id testnet-1 --from mykey

# Create governance proposal
docker exec sekin-sekai-1 /sekaid tx customgov proposal <type> <params> \
  --chain-id testnet-1 --from mykey

# Vote on proposal
docker exec sekin-sekai-1 /sekaid tx customgov proposal vote <proposal-id> <vote> \
  --chain-id testnet-1 --from mykey
```

### State management
```bash
# Export current state
docker exec sekin-sekai-1 /sekaid export --height -1 --output-document state.json

# Validate genesis
docker exec sekin-sekai-1 /sekaid validate-genesis
```

---

## Module-Specific Command Details

### CustomGov Module
The governance module provides extensive functionality for managing the network through proposals, roles, permissions, and identity records.

#### Proposal Types:
- Account management (assign/unassign roles, whitelist/blacklist permissions)
- Role management (create, remove, modify permissions)
- Network properties and execution fees
- Identity record verification
- Data registry updates

#### Key Subcommands:
- `proposal account` - Manage account-related proposals
- `proposal role` - Manage role-related proposals
- `councilor` - Councilor-specific operations
- `poll` - Create and vote on polls
- `permission` - Direct permission management
- `role` - Direct role management

### MultiStaking Module
Provides advanced staking functionality with pools and compound rewards.

#### Key Operations:
- Delegate to staking pools
- Claim rewards with optional compounding
- Manage undelegations
- Register as delegator
- Create and update staking pools

### Tokens Module
Manages token rates and blacklist/whitelist functionality.

#### Key Features:
- Token rate management
- Token blacklist/whitelist updates
- Rate proposals through governance

### Basket Module
Manages basket tokens with advanced DeFi features.

#### Key Operations:
- Mint and burn basket tokens
- Swap basket tokens
- Claim rewards from baskets
- Emergency disable functions
- Governance proposals for basket management

### Spending Module
Manages spending pools for fund distribution.

#### Key Features:
- Create and manage spending pools
- Register beneficiaries
- Claim from pools
- Governance-controlled distributions

---

## Notes

- All commands must be prefixed with `docker exec sekin-sekai-1` when running from outside the container
- Most commands support both JSON and text output formats via the `--output` flag
- Transaction commands typically require the `--chain-id` flag
- Query commands can specify a specific node with the `--node` flag
- The `--home` flag can be used to specify a custom configuration directory
- Use `--help` on any command or subcommand to see all available flags and options
- Commands with multiple levels of subcommands follow a hierarchical structure as shown in the tree
- Some modules (like customgov) have extensive sub-subcommand structures for complex operations