// Package sdk provides a Go SDK for interacting with SEKAI blockchain.
// It abstracts the communication layer (Docker exec, REST API, gRPC) behind
// a unified Client interface, enabling the same codebase to power CLI tools,
// web backends, bots, and other applications.
package sdk

import (
	"context"
)

// Client is the core abstraction for blockchain communication.
// It can be implemented by Docker exec, REST API, gRPC, etc.
type Client interface {
	// Query executes a read operation against the blockchain.
	Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error)

	// Tx executes a transaction (write operation) on the blockchain.
	Tx(ctx context.Context, req *TxRequest) (*TxResponse, error)

	// Keys returns the keyring client for key management operations.
	// Note: May return limited functionality for non-Docker clients.
	Keys() KeysClient

	// Status returns the node status.
	Status(ctx context.Context) (*StatusResponse, error)

	// Close releases any resources held by the client.
	Close() error
}

// QueryRequest represents a query to the blockchain.
type QueryRequest struct {
	// Module is the blockchain module to query (e.g., "bank", "customgov", "staking")
	Module string

	// Endpoint is the specific query endpoint (e.g., "balances", "proposal", "validators")
	Endpoint string

	// Params contains query parameters as key-value pairs
	Params map[string]string

	// ArrayParams contains query parameters that support multiple values (e.g., tokens[]=ukex&tokens[]=lol)
	ArrayParams map[string][]string

	// RawArgs contains raw arguments for direct command execution (Docker client)
	RawArgs []string
}

// QueryResponse represents the response from a query operation.
type QueryResponse struct {
	// Data contains the raw JSON response data
	Data []byte

	// Height is the block height at which the query was executed
	Height int64
}

// TxRequest represents a transaction to submit to the blockchain.
type TxRequest struct {
	// Module is the blockchain module (e.g., "bank", "customgov", "staking")
	Module string

	// Action is the transaction action (e.g., "send", "vote", "delegate")
	Action string

	// Args contains positional arguments for the transaction
	Args []string

	// Flags contains transaction flags (e.g., fees, gas, memo)
	Flags map[string]string

	// BoolFlags contains boolean flags that don't take values
	BoolFlags map[string]bool

	// ArrayFlags contains flags that can be specified multiple times (e.g., --tokens a --tokens b)
	ArrayFlags map[string][]string

	// Signer is the key name or address to sign the transaction
	Signer string

	// BroadcastMode specifies how to broadcast (sync, async, block)
	BroadcastMode string

	// SkipConfirmation skips the confirmation prompt (--yes flag)
	SkipConfirmation bool
}

// TxResponse represents the response from a transaction operation.
type TxResponse struct {
	// TxHash is the transaction hash
	TxHash string `json:"txhash"`

	// Code is the response code (0 = success)
	Code uint32 `json:"code"`

	// Height is the block height where the tx was included
	Height int64 `json:"height,string"`

	// GasUsed is the amount of gas used
	GasUsed int64 `json:"gas_used,string"`

	// GasWanted is the amount of gas requested
	GasWanted int64 `json:"gas_wanted,string"`

	// RawLog contains the raw log output
	RawLog string `json:"raw_log"`

	// Data contains any returned data
	Data string `json:"data,omitempty"`

	// Logs contains structured log entries
	Logs []TxLog `json:"logs,omitempty"`
}

// TxLog represents a single log entry in a transaction response.
type TxLog struct {
	// MsgIndex is the index of the message
	MsgIndex uint32 `json:"msg_index"`

	// Log is the log message
	Log string `json:"log"`

	// Events contains events emitted by this message
	Events []TxEvent `json:"events,omitempty"`
}

// TxEvent represents an event emitted during transaction execution.
type TxEvent struct {
	// Type is the event type
	Type string `json:"type"`

	// Attributes contains event attributes
	Attributes []TxEventAttribute `json:"attributes,omitempty"`
}

// TxEventAttribute represents a single attribute in an event.
type TxEventAttribute struct {
	// Key is the attribute key
	Key string `json:"key"`

	// Value is the attribute value
	Value string `json:"value"`
}

// StatusResponse contains node status information.
type StatusResponse struct {
	// NodeInfo contains information about the node
	NodeInfo NodeInfo `json:"node_info"`

	// SyncInfo contains sync status information
	SyncInfo SyncInfo `json:"sync_info"`

	// ValidatorInfo contains validator information (if applicable)
	ValidatorInfo ValidatorInfo `json:"validator_info,omitempty"`
}

// NodeInfo contains information about the blockchain node.
type NodeInfo struct {
	// Network is the chain ID
	Network string `json:"network"`

	// Moniker is the node's moniker/name
	Moniker string `json:"moniker"`

	// Version is the software version
	Version string `json:"version"`
}

// SyncInfo contains synchronization status.
type SyncInfo struct {
	// LatestBlockHeight is the latest block height
	LatestBlockHeight int64 `json:"latest_block_height"`

	// LatestBlockTime is the timestamp of the latest block
	LatestBlockTime string `json:"latest_block_time"`

	// CatchingUp indicates if the node is still syncing
	CatchingUp bool `json:"catching_up"`
}

// ValidatorInfo contains validator-specific information.
type ValidatorInfo struct {
	// Address is the validator's consensus address
	Address string `json:"address"`

	// VotingPower is the validator's voting power
	VotingPower int64 `json:"voting_power"`
}

// KeysClient handles key management operations.
// Note: Most implementations will only work with Docker client,
// as keys are stored locally on the node.
type KeysClient interface {
	// Add creates a new key with the given name.
	Add(ctx context.Context, name string, opts *KeyAddOptions) (*KeyInfo, error)

	// Delete removes a key by name.
	Delete(ctx context.Context, name string, force bool) error

	// List returns all keys in the keyring.
	List(ctx context.Context) ([]KeyInfo, error)

	// Show returns information about a specific key.
	Show(ctx context.Context, name string) (*KeyInfo, error)

	// Export exports a key as ASCII-armored string.
	Export(ctx context.Context, name string) (string, error)

	// Import imports a key from ASCII-armored string.
	Import(ctx context.Context, name, armor string) error

	// Rename renames a key.
	Rename(ctx context.Context, oldName, newName string) error

	// Mnemonic returns the mnemonic for a key (if available).
	Mnemonic(ctx context.Context, name string) (string, error)

	// ImportHex imports a hex-encoded private key.
	ImportHex(ctx context.Context, name, hexKey, keyType string) error

	// ListKeyTypes returns all supported key algorithms.
	ListKeyTypes(ctx context.Context) ([]string, error)

	// Migrate migrates keys from amino to protobuf format.
	Migrate(ctx context.Context) error

	// Parse converts address from hex to bech32 or vice versa.
	Parse(ctx context.Context, address string) (*ParsedAddress, error)
}

// ParsedAddress contains address conversion results.
type ParsedAddress struct {
	Human  string `json:"human"` // The human-readable prefix (e.g., "kira")
	Bytes  string `json:"bytes"` // The hex-encoded bytes
	Hex    string // Alias for Bytes (for convenience)
	Bech32 string // The original bech32 address (if input was bech32)
}

// Pagination holds pagination parameters for list queries.
type Pagination struct {
	// Key is the pagination key for cursor-based pagination
	Key string
	// Offset is the offset for offset-based pagination
	Offset uint64
	// Limit is the maximum number of items to return
	Limit uint64
	// CountTotal requests the total count
	CountTotal bool
	// Reverse reverses the order
	Reverse bool
}

// KeyAddOptions configures key creation.
type KeyAddOptions struct {
	// Recover indicates whether to recover from mnemonic
	Recover bool

	// Mnemonic is the BIP39 mnemonic (required if Recover is true)
	Mnemonic string

	// HDPath is the HD derivation path (default: m/44'/118'/0'/0/0)
	HDPath string

	// Algorithm is the key algorithm (default: secp256k1)
	Algorithm string

	// NoBackup skips mnemonic display
	NoBackup bool

	// Index is the account index for HD derivation
	Index uint32

	// Multisig configures multisig key generation
	Multisig []string

	// MultisigThreshold is the number of signatures required
	MultisigThreshold int
}

// KeyInfo contains information about a key.
type KeyInfo struct {
	// Name is the key name
	Name string `json:"name"`

	// Type is the key type (local, ledger, multi, offline)
	Type string `json:"type"`

	// Address is the account address
	Address string `json:"address"`

	// PubKey is the public key
	PubKey string `json:"pubkey"`

	// Mnemonic is the BIP39 mnemonic (only set on creation)
	Mnemonic string `json:"mnemonic,omitempty"`
}
