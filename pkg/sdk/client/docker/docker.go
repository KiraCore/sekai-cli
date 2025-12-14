// Package docker provides a Docker-based client for SEKAI blockchain.
// It executes sekaid commands via docker exec inside a running container.
package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Client implements the sdk.Client interface using Docker exec.
type Client struct {
	config *Config
	keys   *keysClient
}

// Config holds configuration for the Docker client.
type Config struct {
	// Container is the Docker container name or ID.
	Container string

	// SekaidPath is the path to the sekaid binary inside the container.
	SekaidPath string

	// ChainID is the blockchain network identifier.
	ChainID string

	// KeyringBackend is the keyring backend type.
	KeyringBackend string

	// Home is the sekaid home directory inside the container.
	Home string

	// Node is the RPC endpoint.
	Node string

	// Fees is the default transaction fee.
	Fees string

	// Gas is the default gas limit for transactions.
	Gas string

	// GasAdjustment is the gas adjustment factor.
	GasAdjustment float64

	// BroadcastMode is the default broadcast mode.
	BroadcastMode string

	// Output is the default output format.
	Output string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		SekaidPath:     "/sekaid",
		ChainID:        "localnet-1",
		KeyringBackend: "test",
		Home:           "/sekai",
		Node:           "tcp://localhost:26657",
		Fees:           "100ukex",
		Gas:            "200000",
		GasAdjustment:  1.3,
		BroadcastMode:  "sync",
		Output:         "json",
	}
}

// Option is a function that configures the Docker client.
type Option func(*Config)

// WithContainer sets the container name.
func WithContainer(container string) Option {
	return func(c *Config) {
		c.Container = container
	}
}

// WithSekaidPath sets the sekaid binary path.
func WithSekaidPath(path string) Option {
	return func(c *Config) {
		c.SekaidPath = path
	}
}

// WithChainID sets the chain ID.
func WithChainID(chainID string) Option {
	return func(c *Config) {
		c.ChainID = chainID
	}
}

// WithKeyringBackend sets the keyring backend.
func WithKeyringBackend(backend string) Option {
	return func(c *Config) {
		c.KeyringBackend = backend
	}
}

// WithHome sets the home directory.
func WithHome(home string) Option {
	return func(c *Config) {
		c.Home = home
	}
}

// WithNode sets the node RPC endpoint.
func WithNode(node string) Option {
	return func(c *Config) {
		c.Node = node
	}
}

// WithFees sets the default fees.
func WithFees(fees string) Option {
	return func(c *Config) {
		c.Fees = fees
	}
}

// WithGas sets the default gas limit.
func WithGas(gas string) Option {
	return func(c *Config) {
		c.Gas = gas
	}
}

// WithGasAdjustment sets the gas adjustment factor.
func WithGasAdjustment(adj float64) Option {
	return func(c *Config) {
		c.GasAdjustment = adj
	}
}

// WithBroadcastMode sets the broadcast mode.
func WithBroadcastMode(mode string) Option {
	return func(c *Config) {
		c.BroadcastMode = mode
	}
}

// WithOutput sets the output format.
func WithOutput(output string) Option {
	return func(c *Config) {
		c.Output = output
	}
}

// NewClient creates a new Docker-based client.
func NewClient(container string, opts ...Option) (*Client, error) {
	if container == "" {
		return nil, fmt.Errorf("container name is required")
	}

	cfg := DefaultConfig()
	cfg.Container = container

	for _, opt := range opts {
		opt(cfg)
	}

	c := &Client{
		config: cfg,
	}
	c.keys = &keysClient{client: c}

	return c, nil
}

// Query executes a query operation.
func (c *Client) Query(ctx context.Context, req *sdk.QueryRequest) (*sdk.QueryResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("query request is required")
	}

	// Build query command
	args := c.buildQueryArgs(req)

	// Execute command
	result, err := c.exec(ctx, args...)
	if err != nil {
		return nil, sdk.WrapQueryError(req.Module, req.Endpoint, err)
	}

	return &sdk.QueryResponse{
		Data: []byte(result.Stdout),
	}, nil
}

// Tx executes a transaction.
func (c *Client) Tx(ctx context.Context, req *sdk.TxRequest) (*sdk.TxResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("transaction request is required")
	}

	// Build transaction command
	args := c.buildTxArgs(req)

	// Execute command
	result, err := c.exec(ctx, args...)
	if err != nil {
		return nil, sdk.WrapTxError(req.Module, req.Action, err)
	}

	// Parse response
	var txResp sdk.TxResponse
	if err := json.Unmarshal([]byte(result.Stdout), &txResp); err != nil {
		// Try to extract error from raw output
		return nil, &sdk.TxError{
			Module: req.Module,
			Action: req.Action,
			RawLog: result.Stdout,
			Err:    fmt.Errorf("failed to parse response: %w", err),
		}
	}

	// Check for transaction error
	if txResp.Code != 0 {
		return &txResp, sdk.NewTxErrorFromResponse(req.Module, req.Action, &txResp)
	}

	return &txResp, nil
}

// Keys returns the keyring client.
func (c *Client) Keys() sdk.KeysClient {
	return c.keys
}

// Status returns the node status.
func (c *Client) Status(ctx context.Context) (*sdk.StatusResponse, error) {
	result, err := c.exec(ctx, "status", "--home", c.config.Home)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Parse the raw sekaid status response (PascalCase fields)
	var rawStatus struct {
		NodeInfo struct {
			Network string `json:"network"`
			Moniker string `json:"moniker"`
			Version string `json:"version"`
		} `json:"NodeInfo"`
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
			LatestBlockTime   string `json:"latest_block_time"`
			CatchingUp        bool   `json:"catching_up"`
		} `json:"SyncInfo"`
		ValidatorInfo struct {
			Address     string `json:"Address"`
			VotingPower string `json:"VotingPower"`
		} `json:"ValidatorInfo"`
	}

	if err := json.Unmarshal([]byte(result.Stdout), &rawStatus); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	// Convert to SDK response format
	var height int64
	fmt.Sscanf(rawStatus.SyncInfo.LatestBlockHeight, "%d", &height)

	var votingPower int64
	fmt.Sscanf(rawStatus.ValidatorInfo.VotingPower, "%d", &votingPower)

	return &sdk.StatusResponse{
		NodeInfo: sdk.NodeInfo{
			Network: rawStatus.NodeInfo.Network,
			Moniker: rawStatus.NodeInfo.Moniker,
			Version: rawStatus.NodeInfo.Version,
		},
		SyncInfo: sdk.SyncInfo{
			LatestBlockHeight: height,
			LatestBlockTime:   rawStatus.SyncInfo.LatestBlockTime,
			CatchingUp:        rawStatus.SyncInfo.CatchingUp,
		},
		ValidatorInfo: sdk.ValidatorInfo{
			Address:     rawStatus.ValidatorInfo.Address,
			VotingPower: votingPower,
		},
	}, nil
}

// Close releases resources (no-op for Docker client).
func (c *Client) Close() error {
	return nil
}

// Config returns the client configuration.
func (c *Client) Config() *Config {
	return c.config
}

// buildQueryArgs builds command arguments for a query.
func (c *Client) buildQueryArgs(req *sdk.QueryRequest) []string {
	args := []string{"query"}

	// Add module and endpoint
	if req.Module != "" {
		args = append(args, req.Module)
	}
	if req.Endpoint != "" {
		args = append(args, req.Endpoint)
	}

	// Add raw args if provided
	args = append(args, req.RawArgs...)

	// Add params as flags
	for key, value := range req.Params {
		if value != "" {
			args = append(args, "--"+key, value)
		}
	}

	// Add default flags
	args = append(args,
		"--output", "json",
		"--node", c.config.Node,
		"--home", c.config.Home,
	)

	if c.config.ChainID != "" {
		args = append(args, "--chain-id", c.config.ChainID)
	}

	return args
}

// buildTxArgs builds command arguments for a transaction.
func (c *Client) buildTxArgs(req *sdk.TxRequest) []string {
	args := []string{"tx"}

	// Add module and action
	if req.Module != "" {
		args = append(args, req.Module)
	}
	if req.Action != "" {
		// Split action by spaces to support nested commands like "proposal unjail-validator"
		actionParts := strings.Split(req.Action, " ")
		args = append(args, actionParts...)
	}

	// Add positional args
	args = append(args, req.Args...)

	// Add signer
	if req.Signer != "" {
		args = append(args, "--from", req.Signer)
	}

	// Add request flags
	for key, value := range req.Flags {
		if value != "" {
			args = append(args, "--"+key, value)
		}
	}

	// Add boolean flags
	// For true: just --flag (or --flag=true)
	// For false: must use --flag=false to override defaults
	for key, enabled := range req.BoolFlags {
		if enabled {
			args = append(args, "--"+key)
		} else {
			args = append(args, "--"+key+"=false")
		}
	}

	// Add array flags (can be specified multiple times)
	for key, values := range req.ArrayFlags {
		for _, value := range values {
			args = append(args, "--"+key, value)
		}
	}

	// Add default flags
	args = append(args,
		"--output", "json",
		"--node", c.config.Node,
		"--keyring-backend", c.config.KeyringBackend,
	)

	if c.config.ChainID != "" {
		args = append(args, "--chain-id", c.config.ChainID)
	}

	if c.config.Home != "" {
		args = append(args, "--home", c.config.Home)
	}

	// Add fees if not already specified
	if _, ok := req.Flags["fees"]; !ok && c.config.Fees != "" {
		args = append(args, "--fees", c.config.Fees)
	}

	// Add gas adjustment if not already specified
	if _, ok := req.Flags["gas-adjustment"]; !ok && c.config.GasAdjustment > 0 {
		args = append(args, "--gas-adjustment", fmt.Sprintf("%.2f", c.config.GasAdjustment))
	}

	// Add gas if not already specified
	if _, ok := req.Flags["gas"]; !ok && c.config.Gas != "" {
		args = append(args, "--gas", c.config.Gas)
	}

	// Set broadcast mode
	mode := req.BroadcastMode
	if mode == "" {
		mode = c.config.BroadcastMode
	}
	if mode != "" {
		args = append(args, "--broadcast-mode", mode)
	}

	// Skip confirmation
	if req.SkipConfirmation {
		args = append(args, "--yes")
	} else {
		// Always skip confirmation in automated context
		args = append(args, "--yes")
	}

	return args
}

// ExecResult holds the result of a command execution.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// exec executes a sekaid command in the Docker container.
func (c *Client) exec(ctx context.Context, args ...string) (*ExecResult, error) {
	return execCommand(ctx, c.config.Container, c.config.SekaidPath, args...)
}

// RawExec executes a raw sekaid command and returns the output.
// This is useful for custom commands not covered by the standard interface.
func (c *Client) RawExec(ctx context.Context, args ...string) (string, error) {
	result, err := c.exec(ctx, args...)
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}

// RawExecWithFlags executes a sekaid command with common flags added.
func (c *Client) RawExecWithFlags(ctx context.Context, args ...string) (string, error) {
	// Add common flags
	args = append(args,
		"--output", "json",
		"--node", c.config.Node,
		"--keyring-backend", c.config.KeyringBackend,
	)

	if c.config.ChainID != "" {
		args = append(args, "--chain-id", c.config.ChainID)
	}

	if c.config.Home != "" {
		args = append(args, "--home", c.config.Home)
	}

	return c.RawExec(ctx, args...)
}

// keysClient implements sdk.KeysClient for Docker.
type keysClient struct {
	client *Client
}

func (k *keysClient) Add(ctx context.Context, name string, opts *sdk.KeyAddOptions) (*sdk.KeyInfo, error) {
	args := []string{"keys", "add", name}

	if opts != nil {
		if opts.Recover {
			args = append(args, "--recover")
		}
		if opts.HDPath != "" {
			args = append(args, "--hd-path", opts.HDPath)
		}
		if opts.Algorithm != "" {
			args = append(args, "--algo", opts.Algorithm)
		}
		if opts.NoBackup {
			args = append(args, "--no-backup")
		}
		if opts.Index > 0 {
			args = append(args, "--index", fmt.Sprintf("%d", opts.Index))
		}
	}

	args = append(args,
		"--keyring-backend", k.client.config.KeyringBackend,
		"--home", k.client.config.Home,
		"--output", "json",
	)

	// For recovery, we need to handle mnemonic input differently
	// This is a simplified implementation
	result, err := k.client.exec(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to add key: %w", err)
	}

	var keyInfo sdk.KeyInfo
	if err := json.Unmarshal([]byte(result.Stdout), &keyInfo); err != nil {
		// Try to parse as different format
		return nil, fmt.Errorf("failed to parse key info: %w", err)
	}

	return &keyInfo, nil
}

func (k *keysClient) Delete(ctx context.Context, name string, force bool) error {
	args := []string{"keys", "delete", name, "--yes"}

	if force {
		args = append(args, "--force")
	}

	args = append(args,
		"--keyring-backend", k.client.config.KeyringBackend,
		"--home", k.client.config.Home,
	)

	_, err := k.client.exec(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	return nil
}

func (k *keysClient) List(ctx context.Context) ([]sdk.KeyInfo, error) {
	args := []string{"keys", "list",
		"--keyring-backend", k.client.config.KeyringBackend,
		"--home", k.client.config.Home,
		"--output", "json",
	}

	result, err := k.client.exec(ctx, args...)
	if err != nil {
		// Check if it's just "no keys found" message
		if strings.Contains(result.Stderr, "No records") || strings.Contains(result.Stdout, "No records") {
			return []sdk.KeyInfo{}, nil
		}
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	// Handle empty list or "No records" message
	stdout := strings.TrimSpace(result.Stdout)
	if stdout == "" || stdout == "[]" || stdout == "null" || strings.Contains(stdout, "No records") {
		return []sdk.KeyInfo{}, nil
	}

	var keys []sdk.KeyInfo
	if err := json.Unmarshal([]byte(stdout), &keys); err != nil {
		return nil, fmt.Errorf("failed to parse keys: %w", err)
	}

	return keys, nil
}

func (k *keysClient) Show(ctx context.Context, name string) (*sdk.KeyInfo, error) {
	args := []string{"keys", "show", name,
		"--keyring-backend", k.client.config.KeyringBackend,
		"--home", k.client.config.Home,
		"--output", "json",
	}

	result, err := k.client.exec(ctx, args...)
	if err != nil {
		// Check if the error indicates key not found
		errStr := err.Error()
		if strings.Contains(errStr, "is not a valid name or address") ||
			strings.Contains(errStr, "key not found") ||
			strings.Contains(errStr, "No key") {
			return nil, sdk.ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to show key: %w", err)
	}

	var keyInfo sdk.KeyInfo
	if err := json.Unmarshal([]byte(result.Stdout), &keyInfo); err != nil {
		return nil, fmt.Errorf("failed to parse key info: %w", err)
	}

	return &keyInfo, nil
}

func (k *keysClient) Export(ctx context.Context, name string) (string, error) {
	args := []string{"keys", "export", name,
		"--keyring-backend", k.client.config.KeyringBackend,
		"--home", k.client.config.Home,
		"--unarmored-hex", "--unsafe",
	}

	result, err := k.client.exec(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to export key: %w", err)
	}

	return strings.TrimSpace(result.Stdout), nil
}

func (k *keysClient) Import(ctx context.Context, name, armor string) error {
	// Key import typically requires file input
	// This is a simplified implementation
	return fmt.Errorf("import not implemented - requires file input")
}

func (k *keysClient) Rename(ctx context.Context, oldName, newName string) error {
	args := []string{"keys", "rename", oldName, newName, "--yes",
		"--keyring-backend", k.client.config.KeyringBackend,
		"--home", k.client.config.Home,
	}

	_, err := k.client.exec(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to rename key: %w", err)
	}

	return nil
}

func (k *keysClient) Mnemonic(ctx context.Context, name string) (string, error) {
	// Mnemonic is only available at creation time
	return "", fmt.Errorf("mnemonic retrieval not supported - only available during key creation")
}

func (k *keysClient) ImportHex(ctx context.Context, name, hexKey, keyType string) error {
	args := []string{"keys", "import-hex", name, hexKey,
		"--keyring-backend", k.client.config.KeyringBackend,
		"--home", k.client.config.Home,
	}
	if keyType != "" {
		args = append(args, "--key-type", keyType)
	}
	_, err := k.client.exec(ctx, args...)
	return err
}

func (k *keysClient) ListKeyTypes(ctx context.Context) ([]string, error) {
	args := []string{"keys", "list-key-types",
		"--keyring-backend", k.client.config.KeyringBackend,
		"--home", k.client.config.Home,
	}
	result, err := k.client.exec(ctx, args...)
	if err != nil {
		return nil, err
	}
	// Parse output (each line is a key type)
	var types []string
	for _, line := range strings.Split(strings.TrimSpace(result.Stdout), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			types = append(types, line)
		}
	}
	return types, nil
}

func (k *keysClient) Migrate(ctx context.Context) error {
	args := []string{"keys", "migrate",
		"--keyring-backend", k.client.config.KeyringBackend,
		"--home", k.client.config.Home,
	}
	_, err := k.client.exec(ctx, args...)
	return err
}

func (k *keysClient) Parse(ctx context.Context, address string) (*sdk.ParsedAddress, error) {
	args := []string{"keys", "parse", address, "--output", "json"}
	execResult, err := k.client.exec(ctx, args...)
	if err != nil {
		return nil, err
	}

	output := strings.TrimSpace(execResult.Stdout)

	// Parse JSON output: {"human":"kira","bytes":"C39EE168..."}
	var result sdk.ParsedAddress
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		// Try parsing text output
		for _, line := range strings.Split(output, "\n") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if key == "bytes" {
					result.Bytes = val
				} else if key == "human" {
					result.Human = val
				}
			}
		}
	}

	// Set convenience fields
	result.Hex = result.Bytes
	if strings.HasPrefix(address, result.Human) {
		result.Bech32 = address
	}

	return &result, nil
}
