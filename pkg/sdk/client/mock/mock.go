// Package mock provides a mock client for testing SEKAI SDK.
// It allows setting up predefined responses for queries and transactions.
package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Client is a mock implementation of sdk.Client for testing.
type Client struct {
	mu sync.RWMutex

	// QueryResponses maps query keys to responses.
	// Key format: "module/endpoint" or "module/endpoint/param1=value1"
	QueryResponses map[string]*sdk.QueryResponse

	// QueryErrors maps query keys to errors.
	QueryErrors map[string]error

	// TxResponses maps transaction keys to responses.
	// Key format: "module/action" or "module/action/arg1/arg2"
	TxResponses map[string]*sdk.TxResponse

	// TxErrors maps transaction keys to errors.
	TxErrors map[string]error

	// StatusResponse is the response for Status() calls.
	StatusResponse *sdk.StatusResponse

	// StatusError is the error for Status() calls.
	StatusError error

	// Keys is the mock keys client.
	keys *keysClient

	// QueryCalls tracks all query calls made.
	QueryCalls []QueryCall

	// TxCalls tracks all transaction calls made.
	TxCalls []TxCall
}

// QueryCall records a query call for verification.
type QueryCall struct {
	Request *sdk.QueryRequest
	Key     string
}

// TxCall records a transaction call for verification.
type TxCall struct {
	Request *sdk.TxRequest
	Key     string
}

// NewClient creates a new mock client.
func NewClient() *Client {
	c := &Client{
		QueryResponses: make(map[string]*sdk.QueryResponse),
		QueryErrors:    make(map[string]error),
		TxResponses:    make(map[string]*sdk.TxResponse),
		TxErrors:       make(map[string]error),
		QueryCalls:     make([]QueryCall, 0),
		TxCalls:        make([]TxCall, 0),
	}
	c.keys = &keysClient{client: c}
	return c
}

// Query executes a mock query.
func (c *Client) Query(ctx context.Context, req *sdk.QueryRequest) (*sdk.QueryResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.queryKey(req)
	c.QueryCalls = append(c.QueryCalls, QueryCall{Request: req, Key: key})

	// Check for error first
	if err, ok := c.QueryErrors[key]; ok {
		return nil, err
	}

	// Try specific key first, then module/endpoint, then module only
	keys := []string{
		key,
		fmt.Sprintf("%s/%s", req.Module, req.Endpoint),
		req.Module,
	}

	for _, k := range keys {
		if resp, ok := c.QueryResponses[k]; ok {
			return resp, nil
		}
	}

	return nil, fmt.Errorf("no mock response for query: %s", key)
}

// Tx executes a mock transaction.
func (c *Client) Tx(ctx context.Context, req *sdk.TxRequest) (*sdk.TxResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.txKey(req)
	c.TxCalls = append(c.TxCalls, TxCall{Request: req, Key: key})

	// Check for error first
	if err, ok := c.TxErrors[key]; ok {
		return nil, err
	}

	// Try specific key first, then module/action, then module only
	keys := []string{
		key,
		fmt.Sprintf("%s/%s", req.Module, req.Action),
		req.Module,
	}

	for _, k := range keys {
		if resp, ok := c.TxResponses[k]; ok {
			return resp, nil
		}
	}

	// Return a default success response
	return &sdk.TxResponse{
		TxHash: "MOCK_TX_HASH_" + key,
		Code:   0,
	}, nil
}

// Keys returns the mock keys client.
func (c *Client) Keys() sdk.KeysClient {
	return c.keys
}

// Status returns the mock status.
func (c *Client) Status(ctx context.Context) (*sdk.StatusResponse, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.StatusError != nil {
		return nil, c.StatusError
	}

	if c.StatusResponse != nil {
		return c.StatusResponse, nil
	}

	// Default response
	return &sdk.StatusResponse{
		NodeInfo: sdk.NodeInfo{
			Network: "mock-network",
			Moniker: "mock-node",
			Version: "1.0.0",
		},
		SyncInfo: sdk.SyncInfo{
			LatestBlockHeight: 12345,
			CatchingUp:        false,
		},
	}, nil
}

// Close is a no-op for mock client.
func (c *Client) Close() error {
	return nil
}

// SetQueryResponse sets a mock response for a query.
func (c *Client) SetQueryResponse(module, endpoint string, data interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s", module, endpoint)
	c.QueryResponses[key] = &sdk.QueryResponse{
		Data: jsonData,
	}
	return nil
}

// SetQueryResponseRaw sets a raw mock response for a query.
func (c *Client) SetQueryResponseRaw(module, endpoint string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s/%s", module, endpoint)
	c.QueryResponses[key] = &sdk.QueryResponse{
		Data: data,
	}
}

// SetQueryError sets a mock error for a query.
func (c *Client) SetQueryError(module, endpoint string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s/%s", module, endpoint)
	c.QueryErrors[key] = err
}

// SetTxResponse sets a mock response for a transaction.
func (c *Client) SetTxResponse(module, action string, resp *sdk.TxResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s/%s", module, action)
	c.TxResponses[key] = resp
}

// SetTxError sets a mock error for a transaction.
func (c *Client) SetTxError(module, action string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s/%s", module, action)
	c.TxErrors[key] = err
}

// SetStatus sets the mock status response.
func (c *Client) SetStatus(resp *sdk.StatusResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.StatusResponse = resp
}

// SetStatusError sets the mock status error.
func (c *Client) SetStatusError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.StatusError = err
}

// Reset clears all mock data and call history.
func (c *Client) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.QueryResponses = make(map[string]*sdk.QueryResponse)
	c.QueryErrors = make(map[string]error)
	c.TxResponses = make(map[string]*sdk.TxResponse)
	c.TxErrors = make(map[string]error)
	c.StatusResponse = nil
	c.StatusError = nil
	c.QueryCalls = make([]QueryCall, 0)
	c.TxCalls = make([]TxCall, 0)
	c.keys.reset()
}

// GetQueryCalls returns all query calls made.
func (c *Client) GetQueryCalls() []QueryCall {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]QueryCall{}, c.QueryCalls...)
}

// GetTxCalls returns all transaction calls made.
func (c *Client) GetTxCalls() []TxCall {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]TxCall{}, c.TxCalls...)
}

// queryKey generates a key for a query request.
func (c *Client) queryKey(req *sdk.QueryRequest) string {
	parts := []string{req.Module, req.Endpoint}
	parts = append(parts, req.RawArgs...)
	return strings.Join(parts, "/")
}

// txKey generates a key for a transaction request.
func (c *Client) txKey(req *sdk.TxRequest) string {
	parts := []string{req.Module, req.Action}
	parts = append(parts, req.Args...)
	return strings.Join(parts, "/")
}

// keysClient is a mock implementation of sdk.KeysClient.
type keysClient struct {
	client *Client
	mu     sync.RWMutex
	keys   map[string]*sdk.KeyInfo
}

func (k *keysClient) reset() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.keys = make(map[string]*sdk.KeyInfo)
}

func (k *keysClient) Add(ctx context.Context, name string, opts *sdk.KeyAddOptions) (*sdk.KeyInfo, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.keys == nil {
		k.keys = make(map[string]*sdk.KeyInfo)
	}

	if _, exists := k.keys[name]; exists {
		return nil, sdk.ErrKeyExists
	}

	info := &sdk.KeyInfo{
		Name:    name,
		Type:    "local",
		Address: fmt.Sprintf("kira1mock%s", name),
		PubKey:  fmt.Sprintf("kirapub1mock%s", name),
	}

	if opts != nil && !opts.NoBackup {
		info.Mnemonic = "mock mnemonic words for testing purposes only do not use in production"
	}

	k.keys[name] = info
	return info, nil
}

func (k *keysClient) Delete(ctx context.Context, name string, force bool) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.keys == nil || k.keys[name] == nil {
		return sdk.ErrKeyNotFound
	}

	delete(k.keys, name)
	return nil
}

func (k *keysClient) List(ctx context.Context) ([]sdk.KeyInfo, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	result := make([]sdk.KeyInfo, 0, len(k.keys))
	for _, info := range k.keys {
		result = append(result, *info)
	}
	return result, nil
}

func (k *keysClient) Show(ctx context.Context, name string) (*sdk.KeyInfo, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.keys == nil || k.keys[name] == nil {
		return nil, sdk.ErrKeyNotFound
	}

	info := *k.keys[name]
	return &info, nil
}

func (k *keysClient) Export(ctx context.Context, name string) (string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.keys == nil || k.keys[name] == nil {
		return "", sdk.ErrKeyNotFound
	}

	return fmt.Sprintf("mock_exported_key_%s", name), nil
}

func (k *keysClient) Import(ctx context.Context, name, armor string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.keys == nil {
		k.keys = make(map[string]*sdk.KeyInfo)
	}

	if _, exists := k.keys[name]; exists {
		return sdk.ErrKeyExists
	}

	k.keys[name] = &sdk.KeyInfo{
		Name:    name,
		Type:    "local",
		Address: fmt.Sprintf("kira1imported%s", name),
	}
	return nil
}

func (k *keysClient) Rename(ctx context.Context, oldName, newName string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.keys == nil || k.keys[oldName] == nil {
		return sdk.ErrKeyNotFound
	}

	if _, exists := k.keys[newName]; exists {
		return sdk.ErrKeyExists
	}

	info := k.keys[oldName]
	info.Name = newName
	k.keys[newName] = info
	delete(k.keys, oldName)
	return nil
}

func (k *keysClient) Mnemonic(ctx context.Context, name string) (string, error) {
	return "", fmt.Errorf("mnemonic retrieval not supported")
}

func (k *keysClient) ImportHex(ctx context.Context, name, hexKey, keyType string) error {
	return fmt.Errorf("import hex not supported in mock client")
}

func (k *keysClient) ListKeyTypes(ctx context.Context) ([]string, error) {
	return []string{"secp256k1", "ed25519"}, nil
}

func (k *keysClient) Migrate(ctx context.Context) error {
	return nil // No-op in mock
}

func (k *keysClient) Parse(ctx context.Context, address string) (*sdk.ParsedAddress, error) {
	return &sdk.ParsedAddress{
		Human:  "kira",
		Bytes:  "mock_hex_bytes",
		Hex:    "mock_hex_bytes",
		Bech32: address,
	}, nil
}

// AddKey is a helper to add a key directly for testing.
func (k *keysClient) AddKey(info *sdk.KeyInfo) {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.keys == nil {
		k.keys = make(map[string]*sdk.KeyInfo)
	}
	k.keys[info.Name] = info
}
