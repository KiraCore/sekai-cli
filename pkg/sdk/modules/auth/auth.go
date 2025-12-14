// Package auth provides auth module functionality for querying accounts.
package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides auth query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new auth module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// Account queries account info by address.
func (m *Module) Account(ctx context.Context, address string) (*AccountInfo, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "auth",
		Endpoint: "account",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query account: %w", err)
	}

	// Try direct unmarshal first (sekaid returns unwrapped response)
	var acc AccountInfo
	if err := json.Unmarshal(resp.Data, &acc); err == nil && acc.Address != "" {
		return &acc, nil
	}

	// Try wrapped response (some versions may wrap in "account" object)
	var result struct {
		Account AccountInfo `json:"account"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse account: %w", err)
	}
	if result.Account.Address != "" {
		return &result.Account, nil
	}

	return nil, fmt.Errorf("failed to parse account: empty response")
}

// Accounts queries all accounts with pagination.
func (m *Module) Accounts(ctx context.Context, pagination *sdk.Pagination) (*AccountsResponse, error) {
	params := make(map[string]string)
	if pagination != nil {
		if pagination.Limit > 0 {
			params["limit"] = fmt.Sprintf("%d", pagination.Limit)
		}
		if pagination.Offset > 0 {
			params["offset"] = fmt.Sprintf("%d", pagination.Offset)
		}
		if pagination.Key != "" {
			params["page-key"] = pagination.Key
		}
	}

	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "auth",
		Endpoint: "accounts",
		Params:   params,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}

	var result AccountsResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse accounts: %w", err)
	}
	return &result, nil
}

// ModuleAccount queries a module account by name.
func (m *Module) ModuleAccount(ctx context.Context, name string) (*ModuleAccountInfo, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "auth",
		Endpoint: "module-account",
		RawArgs:  []string{name},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query module account: %w", err)
	}

	var result struct {
		Account ModuleAccountInfo `json:"account"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse module account: %w", err)
	}
	return &result.Account, nil
}

// ModuleAccounts queries all module accounts.
func (m *Module) ModuleAccounts(ctx context.Context) ([]ModuleAccountInfo, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "auth",
		Endpoint: "module-accounts",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query module accounts: %w", err)
	}

	var result struct {
		Accounts []ModuleAccountInfo `json:"accounts"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse module accounts: %w", err)
	}
	return result.Accounts, nil
}

// Params queries auth module parameters.
func (m *Module) Params(ctx context.Context) (*AuthParams, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "auth",
		Endpoint: "params",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query params: %w", err)
	}

	// Try direct unmarshal first (sekaid returns unwrapped response)
	var params AuthParams
	if err := json.Unmarshal(resp.Data, &params); err == nil && params.MaxMemoCharacters != "" {
		return &params, nil
	}

	// Try wrapped response
	var result struct {
		Params AuthParams `json:"params"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse params: %w", err)
	}
	return &result.Params, nil
}

// AddressByAccNum queries an address by account number.
func (m *Module) AddressByAccNum(ctx context.Context, accNum string) (*AddressByAccNumResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "auth",
		Endpoint: "address-by-acc-num",
		RawArgs:  []string{accNum},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query address by acc num: %w", err)
	}

	var result AddressByAccNumResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse address: %w", err)
	}
	return &result, nil
}

// AddressByAccNumResponse represents the response for address-by-acc-num query.
type AddressByAccNumResponse struct {
	AccountAddress string `json:"account_address"`
}

// AccountInfo represents account information.
type AccountInfo struct {
	Type          string `json:"@type,omitempty"`
	Address       string `json:"address"`
	PubKey        any    `json:"pub_key,omitempty"`
	AccountNumber string `json:"account_number"`
	Sequence      string `json:"sequence"`
}

// AccountsResponse represents the accounts query response.
type AccountsResponse struct {
	Accounts   []AccountInfo       `json:"accounts"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

// PaginationResponse represents pagination info in response.
type PaginationResponse struct {
	NextKey string `json:"next_key,omitempty"`
	Total   string `json:"total,omitempty"`
}

// ModuleAccountInfo represents a module account.
type ModuleAccountInfo struct {
	Type        string   `json:"@type,omitempty"`
	BaseAccount any      `json:"base_account,omitempty"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions,omitempty"`
}

// AuthParams represents auth module parameters.
type AuthParams struct {
	MaxMemoCharacters      string `json:"max_memo_characters"`
	TxSigLimit             string `json:"tx_sig_limit"`
	TxSizeCostPerByte      string `json:"tx_size_cost_per_byte"`
	SigVerifyCostED25519   string `json:"sig_verify_cost_ed25519"`
	SigVerifyCostSecp256k1 string `json:"sig_verify_cost_secp256k1"`
}
