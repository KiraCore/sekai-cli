// Package bank provides bank module functionality for token transfers.
package bank

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
	"github.com/kiracore/sekai-cli/pkg/sdk/types"
)

// Module provides bank functionality.
type Module struct {
	client sdk.Client
}

// New creates a new bank module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// Balance queries the balance of an address for a specific denomination.
func (m *Module) Balance(ctx context.Context, address, denom string) (*types.Coin, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "bank",
		Endpoint: "balances",
		RawArgs:  []string{address},
		Params: map[string]string{
			"denom": denom,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query balance: %w", err)
	}

	// When querying with --denom, the response is a flat coin object
	var coin types.Coin
	if err := json.Unmarshal(resp.Data, &coin); err != nil {
		// Try wrapped format as fallback
		var result struct {
			Balance types.Coin `json:"balance"`
		}
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse balance: %w", err)
		}
		return &result.Balance, nil
	}

	return &coin, nil
}

// Balances queries all balances of an address.
func (m *Module) Balances(ctx context.Context, address string) (types.Coins, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "bank",
		Endpoint: "balances",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query balances: %w", err)
	}

	var result struct {
		Balances []types.Coin `json:"balances"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse balances: %w", err)
	}

	return types.Coins(result.Balances), nil
}

// SpendableBalances queries spendable balances of an address.
func (m *Module) SpendableBalances(ctx context.Context, address string) (types.Coins, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "bank",
		Endpoint: "spendable-balances",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query spendable balances: %w", err)
	}

	var result struct {
		Balances []types.Coin `json:"balances"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse spendable balances: %w", err)
	}

	return types.Coins(result.Balances), nil
}

// TotalSupply queries the total supply of all tokens.
func (m *Module) TotalSupply(ctx context.Context) (types.Coins, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "bank",
		Endpoint: "total",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query total supply: %w", err)
	}

	var result struct {
		Supply []types.Coin `json:"supply"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse total supply: %w", err)
	}

	return types.Coins(result.Supply), nil
}

// SupplyOf queries the supply of a specific denomination.
func (m *Module) SupplyOf(ctx context.Context, denom string) (*types.Coin, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "bank",
		Endpoint: "total",
		Params: map[string]string{
			"denom": denom,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query supply: %w", err)
	}

	var result struct {
		Amount types.Coin `json:"amount"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse supply: %w", err)
	}

	return &result.Amount, nil
}

// SendOptions configures a send transaction.
type SendOptions struct {
	// Fees is the transaction fee
	Fees string

	// Gas is the gas limit
	Gas string

	// GasAdjustment is the gas adjustment factor
	GasAdjustment float64

	// Memo is the transaction memo
	Memo string

	// BroadcastMode is the broadcast mode
	BroadcastMode string
}

// Send transfers tokens from one account to another.
func (m *Module) Send(ctx context.Context, from, to string, amount types.Coins, opts *SendOptions) (*sdk.TxResponse, error) {
	flags := make(map[string]string)

	if opts != nil {
		if opts.Fees != "" {
			flags["fees"] = opts.Fees
		}
		if opts.Gas != "" {
			flags["gas"] = opts.Gas
		}
		if opts.GasAdjustment > 0 {
			flags["gas-adjustment"] = fmt.Sprintf("%.2f", opts.GasAdjustment)
		}
		if opts.Memo != "" {
			flags["memo"] = opts.Memo
		}
		if opts.BroadcastMode != "" {
			flags["broadcast-mode"] = opts.BroadcastMode
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "bank",
		Action:           "send",
		Args:             []string{from, to, amount.String()},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send tokens: %w", err)
	}

	return resp, nil
}

// MultiSendOptions configures a multi-send transaction.
type MultiSendOptions struct {
	SendOptions
	// Split splits the amount equally between recipients instead of sending full amount to each
	Split bool
}

// MultiSend sends tokens from one account to multiple recipients.
// By default, sends the full amount to each address.
// Use Split option to split the amount equally between addresses.
func (m *Module) MultiSend(ctx context.Context, from string, toAddresses []string, amount types.Coins, opts *MultiSendOptions) (*sdk.TxResponse, error) {
	flags := make(map[string]string)
	boolFlags := make(map[string]bool)
	if opts != nil {
		if opts.Fees != "" {
			flags["fees"] = opts.Fees
		}
		if opts.Gas != "" {
			flags["gas"] = opts.Gas
		}
		if opts.GasAdjustment > 0 {
			flags["gas-adjustment"] = fmt.Sprintf("%.2f", opts.GasAdjustment)
		}
		if opts.Memo != "" {
			flags["memo"] = opts.Memo
		}
		if opts.BroadcastMode != "" {
			flags["broadcast-mode"] = opts.BroadcastMode
		}
		if opts.Split {
			boolFlags["split"] = true
		}
	}

	// Args: from, to1, to2, ..., amount
	args := []string{from}
	args = append(args, toAddresses...)
	args = append(args, amount.String())

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "bank",
		Action:           "multi-send",
		Args:             args,
		Signer:           from,
		Flags:            flags,
		BoolFlags:        boolFlags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to multi-send tokens: %w", err)
	}

	return resp, nil
}

// DenomMetadataResponse contains denom metadata query response.
type DenomMetadataResponse struct {
	Metadatas []DenomMetadata `json:"metadatas"`
}

// DenomMetadata contains metadata about a denomination.
type DenomMetadata struct {
	Description string      `json:"description"`
	DenomUnits  []DenomUnit `json:"denom_units"`
	Base        string      `json:"base"`
	Display     string      `json:"display"`
	Name        string      `json:"name"`
	Symbol      string      `json:"symbol"`
}

// DenomUnit represents a denomination unit.
type DenomUnit struct {
	Denom    string   `json:"denom"`
	Exponent uint32   `json:"exponent"`
	Aliases  []string `json:"aliases"`
}

// DenomMetadata queries metadata for a specific denomination or all if denom is empty.
func (m *Module) DenomMetadata(ctx context.Context, denom string) (*DenomMetadataResponse, error) {
	params := make(map[string]string)
	if denom != "" {
		params["denom"] = denom
	}

	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "bank",
		Endpoint: "denom-metadata",
		Params:   params,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query denom metadata: %w", err)
	}

	var result DenomMetadataResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse denom metadata: %w", err)
	}

	return &result, nil
}

// AllDenomsMetadata queries metadata for all denominations.
func (m *Module) AllDenomsMetadata(ctx context.Context) (*DenomMetadataResponse, error) {
	return m.DenomMetadata(ctx, "")
}

// SendEnabled represents a send enabled entry.
type SendEnabled struct {
	Denom   string `json:"denom"`
	Enabled bool   `json:"enabled"`
}

// SendEnabledResponse contains the response from send-enabled query.
type SendEnabledResponse struct {
	SendEnabled []SendEnabled `json:"send_enabled"`
}

// SendEnabled queries send enabled entries for specific denoms or all.
func (m *Module) SendEnabled(ctx context.Context, denoms ...string) (*SendEnabledResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "bank",
		Endpoint: "send-enabled",
		RawArgs:  denoms,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query send enabled: %w", err)
	}

	var result SendEnabledResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse send enabled: %w", err)
	}

	return &result, nil
}
