// Package rest provides a REST API client for SEKAI blockchain.
// It communicates with SEKAI nodes via INTERX gateway or direct REST endpoints.
package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Client implements the sdk.Client interface using REST API.
type Client struct {
	config     *Config
	httpClient *http.Client
	keys       *keysClient
}

// Config holds configuration for the REST client.
type Config struct {
	// BaseURL is the base URL for the API (e.g., "http://localhost:11000")
	BaseURL string

	// ChainID is the blockchain network identifier.
	ChainID string

	// Timeout is the request timeout.
	Timeout time.Duration

	// UseINTERX indicates whether to use INTERX API format.
	UseINTERX bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:   "http://localhost:11000",
		Timeout:   30 * time.Second,
		UseINTERX: true,
	}
}

// Option is a function that configures the REST client.
type Option func(*Config)

// WithBaseURL sets the base URL.
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = strings.TrimSuffix(url, "/")
	}
}

// WithChainID sets the chain ID.
func WithChainID(chainID string) Option {
	return func(c *Config) {
		c.ChainID = chainID
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithINTERX enables/disables INTERX API format.
func WithINTERX(useINTERX bool) Option {
	return func(c *Config) {
		c.UseINTERX = useINTERX
	}
}

// NewClient creates a new REST API client.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	cfg := DefaultConfig()
	cfg.BaseURL = strings.TrimSuffix(baseURL, "/")

	for _, opt := range opts {
		opt(cfg)
	}

	c := &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
	c.keys = &keysClient{client: c}

	return c, nil
}

// Query executes a query operation via REST API.
func (c *Client) Query(ctx context.Context, req *sdk.QueryRequest) (*sdk.QueryResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("query request is required")
	}

	// Build URL based on module and endpoint
	url := c.buildQueryURL(req)

	// Make HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &sdk.HTTPError{URL: url, Err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &sdk.HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
			URL:        url,
		}
	}

	// Check for INTERX error response
	if c.config.UseINTERX {
		var errResp struct {
			Error  string `json:"Error"`
			Status string `json:"Status"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Status == "NOK" {
			return nil, &sdk.QueryError{
				Module:   req.Module,
				Endpoint: req.Endpoint,
				Err:      fmt.Errorf("%s", errResp.Error),
			}
		}
	}

	return &sdk.QueryResponse{
		Data: body,
	}, nil
}

// Tx executes a transaction via REST API.
// Note: REST transactions require building, signing, and broadcasting separately.
// This implementation returns an error as transaction signing requires local key access.
func (c *Client) Tx(ctx context.Context, req *sdk.TxRequest) (*sdk.TxResponse, error) {
	return nil, sdk.ErrNotSupported
}

// Keys returns the keyring client.
// Note: Keys are stored locally, not accessible via REST.
func (c *Client) Keys() sdk.KeysClient {
	return c.keys
}

// Status returns the node status via REST API.
func (c *Client) Status(ctx context.Context) (*sdk.StatusResponse, error) {
	var url string
	if c.config.UseINTERX {
		url = c.config.BaseURL + "/api/status"
	} else {
		url = c.config.BaseURL + "/cosmos/base/tendermint/v1beta1/node_info"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &sdk.HTTPError{URL: url, Err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &sdk.HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
			URL:        url,
		}
	}

	// Parse INTERX status response
	if c.config.UseINTERX {
		return c.parseINTERXStatus(body)
	}

	// Parse standard Cosmos status response
	return c.parseCosmosStatus(body)
}

// Close releases resources (no-op for REST client).
func (c *Client) Close() error {
	return nil
}

// Config returns the client configuration.
func (c *Client) Config() *Config {
	return c.config
}

// buildQueryURL builds the URL for a query request.
func (c *Client) buildQueryURL(req *sdk.QueryRequest) string {
	var path string

	if c.config.UseINTERX {
		// INTERX API paths
		path = c.buildINTERXPath(req)
	} else {
		// Standard Cosmos REST paths
		path = c.buildCosmosPath(req)
	}

	url := c.config.BaseURL + path

	// Build query parameters
	var params []string

	// Add simple key-value params
	for k, v := range req.Params {
		if v != "" {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Add array params (e.g., tokens[]=ukex&tokens[]=lol)
	for k, values := range req.ArrayParams {
		for _, v := range values {
			if v != "" {
				params = append(params, fmt.Sprintf("%s[]=%s", k, v))
			}
		}
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	return url
}

// buildINTERXPath builds INTERX API path for a query.
func (c *Client) buildINTERXPath(req *sdk.QueryRequest) string {
	// INTERX uses /api/kira/ prefix for most endpoints
	switch req.Module {
	case "bank":
		switch req.Endpoint {
		case "balances":
			if len(req.RawArgs) > 0 {
				return "/api/kira/balances/" + req.RawArgs[0]
			}
			return "/api/kira/balances"
		case "total", "supply":
			return "/api/kira/supply"
		}
	case "customgov", "gov":
		switch req.Endpoint {
		case "proposals":
			return "/api/kira/gov/proposals"
		case "network-properties":
			return "/api/kira/gov/network_properties"
		}
	case "customstaking", "staking":
		switch req.Endpoint {
		case "validators":
			return "/api/valopers"
		}
	case "status":
		return "/api/status"
	}

	// Default: try to map directly
	return fmt.Sprintf("/api/kira/%s/%s", req.Module, req.Endpoint)
}

// buildCosmosPath builds standard Cosmos REST path for a query.
func (c *Client) buildCosmosPath(req *sdk.QueryRequest) string {
	switch req.Module {
	case "bank":
		switch req.Endpoint {
		case "balances":
			if len(req.RawArgs) > 0 {
				return "/cosmos/bank/v1beta1/balances/" + req.RawArgs[0]
			}
			return "/cosmos/bank/v1beta1/balances"
		case "total", "supply":
			return "/cosmos/bank/v1beta1/supply"
		}
	case "auth":
		switch req.Endpoint {
		case "account", "accounts":
			if len(req.RawArgs) > 0 {
				return "/cosmos/auth/v1beta1/accounts/" + req.RawArgs[0]
			}
			return "/cosmos/auth/v1beta1/accounts"
		}
	case "staking":
		switch req.Endpoint {
		case "validators":
			return "/cosmos/staking/v1beta1/validators"
		}
	}

	return fmt.Sprintf("/cosmos/%s/v1beta1/%s", req.Module, req.Endpoint)
}

// parseINTERXStatus parses INTERX status response.
func (c *Client) parseINTERXStatus(data []byte) (*sdk.StatusResponse, error) {
	var resp struct {
		InterxInfo struct {
			ChainID           string `json:"chain_id"`
			Version           string `json:"version"`
			SekaiVersion      string `json:"sekai_version"`
			LatestBlockHeight string `json:"latest_block_height"`
			CatchingUp        bool   `json:"catching_up"`
		} `json:"interx_info"`
		NodeInfo struct {
			Network string `json:"network"`
			Moniker string `json:"moniker"`
			Version string `json:"version"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHeight string `json:"latest_block_height"`
			LatestBlockTime   string `json:"latest_block_time"`
			CatchingUp        bool   `json:"catching_up"`
		} `json:"sync_info"`
		ValidatorInfo struct {
			Address     string `json:"address"`
			VotingPower string `json:"voting_power"`
		} `json:"validator_info"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	height := parseHeight(resp.SyncInfo.LatestBlockHeight)

	return &sdk.StatusResponse{
		NodeInfo: sdk.NodeInfo{
			Network: resp.NodeInfo.Network,
			Moniker: resp.NodeInfo.Moniker,
			Version: resp.NodeInfo.Version,
		},
		SyncInfo: sdk.SyncInfo{
			LatestBlockHeight: height,
			LatestBlockTime:   resp.SyncInfo.LatestBlockTime,
			CatchingUp:        resp.SyncInfo.CatchingUp,
		},
		ValidatorInfo: sdk.ValidatorInfo{
			Address: resp.ValidatorInfo.Address,
		},
	}, nil
}

// parseCosmosStatus parses standard Cosmos status response.
func (c *Client) parseCosmosStatus(data []byte) (*sdk.StatusResponse, error) {
	var resp struct {
		DefaultNodeInfo struct {
			Network string `json:"network"`
			Moniker string `json:"moniker"`
			Version string `json:"version"`
		} `json:"default_node_info"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	return &sdk.StatusResponse{
		NodeInfo: sdk.NodeInfo{
			Network: resp.DefaultNodeInfo.Network,
			Moniker: resp.DefaultNodeInfo.Moniker,
			Version: resp.DefaultNodeInfo.Version,
		},
	}, nil
}

// parseHeight parses a height string to int64.
func parseHeight(s string) int64 {
	var height int64
	fmt.Sscanf(s, "%d", &height)
	return height
}

// Get makes a GET request and returns the response body.
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	url := c.config.BaseURL + path

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &sdk.HTTPError{URL: url, Err: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &sdk.HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
			URL:        url,
		}
	}

	return body, nil
}

// Post makes a POST request and returns the response body.
func (c *Client) Post(ctx context.Context, path string, data interface{}) ([]byte, error) {
	url := c.config.BaseURL + path

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &sdk.HTTPError{URL: url, Err: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &sdk.HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(respBody),
			URL:        url,
		}
	}

	return respBody, nil
}

// keysClient is a REST implementation of sdk.KeysClient.
// Most operations return ErrNotSupported as keys are local.
type keysClient struct {
	client *Client
}

func (k *keysClient) Add(ctx context.Context, name string, opts *sdk.KeyAddOptions) (*sdk.KeyInfo, error) {
	return nil, sdk.ErrNotSupported
}

func (k *keysClient) Delete(ctx context.Context, name string, force bool) error {
	return sdk.ErrNotSupported
}

func (k *keysClient) List(ctx context.Context) ([]sdk.KeyInfo, error) {
	return nil, sdk.ErrNotSupported
}

func (k *keysClient) Show(ctx context.Context, name string) (*sdk.KeyInfo, error) {
	return nil, sdk.ErrNotSupported
}

func (k *keysClient) Export(ctx context.Context, name string) (string, error) {
	return "", sdk.ErrNotSupported
}

func (k *keysClient) Import(ctx context.Context, name, armor string) error {
	return sdk.ErrNotSupported
}

func (k *keysClient) Rename(ctx context.Context, oldName, newName string) error {
	return sdk.ErrNotSupported
}

func (k *keysClient) Mnemonic(ctx context.Context, name string) (string, error) {
	return "", sdk.ErrNotSupported
}

func (k *keysClient) ImportHex(ctx context.Context, name, hexKey, keyType string) error {
	return sdk.ErrNotSupported
}

func (k *keysClient) ListKeyTypes(ctx context.Context) ([]string, error) {
	return nil, sdk.ErrNotSupported
}

func (k *keysClient) Migrate(ctx context.Context) error {
	return sdk.ErrNotSupported
}

func (k *keysClient) Parse(ctx context.Context, address string) (*sdk.ParsedAddress, error) {
	return nil, sdk.ErrNotSupported
}
