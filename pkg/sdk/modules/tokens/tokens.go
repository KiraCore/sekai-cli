// Package tokens provides tokens module functionality.
package tokens

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides tokens query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new tokens module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// AllRates queries all token rates.
func (m *Module) AllRates(ctx context.Context) (*AllRatesResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "tokens",
		Endpoint: "all-rates",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query all rates: %w", err)
	}

	var result AllRatesResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse all rates: %w", err)
	}
	return &result, nil
}

// Rate queries a token rate by denom.
func (m *Module) Rate(ctx context.Context, denom string) (*TokenRate, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "tokens",
		Endpoint: "rate",
		RawArgs:  []string{denom},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query rate: %w", err)
	}

	// Try direct unmarshal
	var rate TokenRate
	if err := json.Unmarshal(resp.Data, &rate); err == nil && rate.Denom != "" {
		return &rate, nil
	}

	// Try wrapped response
	var result struct {
		Data TokenRate `json:"data"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse rate: %w", err)
	}
	return &result.Data, nil
}

// RatesByDenom queries token rates by denom.
// Returns a map of denom -> TokenRateWithSupply.
func (m *Module) RatesByDenom(ctx context.Context, denom string) (map[string]TokenRateWithSupply, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "tokens",
		Endpoint: "rates-by-denom",
		RawArgs:  []string{denom},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query rates by denom: %w", err)
	}

	// Response format: {"data": {"denom": {"data": {...}, "supply": {...}}}}
	var result struct {
		Data map[string]TokenRateWithSupply `json:"data"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse rates: %w", err)
	}
	return result.Data, nil
}

// TokenBlackWhites queries token blacklisted and whitelisted tokens.
func (m *Module) TokenBlackWhites(ctx context.Context) (*TokenBlackWhites, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "tokens",
		Endpoint: "token-black-whites",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query token black whites: %w", err)
	}

	// Response is wrapped in "data"
	var result struct {
		Data TokenBlackWhites `json:"data"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		// Try direct unmarshal
		var tbw TokenBlackWhites
		if err := json.Unmarshal(resp.Data, &tbw); err != nil {
			return nil, fmt.Errorf("failed to parse token black whites: %w", err)
		}
		return &tbw, nil
	}
	return &result.Data, nil
}

// TxOptions contains common transaction options.
type TxOptions struct {
	Fees          string
	Gas           string
	GasAdjustment float64
	Memo          string
	BroadcastMode string
}

func buildTxFlags(opts *TxOptions) map[string]string {
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
			flags["note"] = opts.Memo
		}
		if opts.BroadcastMode != "" {
			flags["broadcast-mode"] = opts.BroadcastMode
		}
	}
	return flags
}

// UpsertRateOpts contains options for upserting a token rate.
type UpsertRateOpts struct {
	Denom       string
	FeeRate     string
	FeePayments bool
	Decimals    uint32

	// Token metadata
	Name        string
	Symbol      string
	Description string
	Icon        string
	Website     string
	Social      string

	// Staking
	StakeToken bool
	StakeCap   string
	StakeMin   string

	// Supply
	Supply    string
	SupplyCap string

	// NFT
	NftHash     string
	NftMetadata string

	// Other
	Owner             string
	OwnerEditDisabled bool
	Invalidated       bool
	TokenRate         string
	TokenType         string
	MintingFee        string
}

// UpsertRate upserts a token rate.
func (m *Module) UpsertRate(ctx context.Context, from string, rateOpts *UpsertRateOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	boolFlags := make(map[string]bool)

	// Set required defaults
	flags["fee_rate"] = "1.0"
	flags["token_rate"] = "1.0"
	flags["stake_cap"] = "0"
	flags["stake_min"] = "0"
	flags["supply"] = "0"
	flags["supply_cap"] = "0"
	flags["minting_fee"] = "0"
	boolFlags["fee_payments"] = true

	if rateOpts != nil {
		if rateOpts.Denom != "" {
			flags["denom"] = rateOpts.Denom
		}
		if rateOpts.FeeRate != "" {
			flags["fee_rate"] = rateOpts.FeeRate
		}
		// FeePayments already set by default
		if rateOpts.Decimals > 0 {
			flags["decimals"] = fmt.Sprintf("%d", rateOpts.Decimals)
		}
		if rateOpts.Name != "" {
			flags["name"] = rateOpts.Name
		}
		if rateOpts.Symbol != "" {
			flags["symbol"] = rateOpts.Symbol
		}
		if rateOpts.Description != "" {
			flags["description"] = rateOpts.Description
		}
		if rateOpts.Icon != "" {
			flags["icon"] = rateOpts.Icon
		}
		if rateOpts.Website != "" {
			flags["website"] = rateOpts.Website
		}
		if rateOpts.Social != "" {
			flags["social"] = rateOpts.Social
		}
		if rateOpts.StakeToken {
			boolFlags["stake_token"] = true
		}
		if rateOpts.StakeCap != "" {
			flags["stake_cap"] = rateOpts.StakeCap
		}
		if rateOpts.StakeMin != "" {
			flags["stake_min"] = rateOpts.StakeMin
		}
		if rateOpts.Supply != "" {
			flags["supply"] = rateOpts.Supply
		}
		if rateOpts.SupplyCap != "" {
			flags["supply_cap"] = rateOpts.SupplyCap
		}
		if rateOpts.NftHash != "" {
			flags["nft_hash"] = rateOpts.NftHash
		}
		if rateOpts.NftMetadata != "" {
			flags["nft_metadata"] = rateOpts.NftMetadata
		}
		if rateOpts.Owner != "" {
			flags["owner"] = rateOpts.Owner
		}
		if rateOpts.OwnerEditDisabled {
			boolFlags["owner_edit_disabled"] = true
		}
		if rateOpts.Invalidated {
			boolFlags["invalidated"] = true
		}
		if rateOpts.TokenRate != "" {
			flags["token_rate"] = rateOpts.TokenRate
		}
		if rateOpts.TokenType != "" {
			flags["token_type"] = rateOpts.TokenType
		}
		if rateOpts.MintingFee != "" {
			flags["minting_fee"] = rateOpts.MintingFee
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "tokens",
		Action:           "upsert-rate",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		BoolFlags:        boolFlags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert rate: %w", err)
	}
	return resp, nil
}

// ProposalUpsertRateOpts contains options for a proposal to upsert token rate.
type ProposalUpsertRateOpts struct {
	// Basic token info
	Denom       string
	Decimals    uint32
	FeeRate     string
	FeePayments bool

	// Token metadata
	Name    string
	Symbol  string
	Icon    string
	Website string
	Social  string

	// Staking
	StakeToken bool
	StakeCap   string
	StakeMin   string

	// Supply
	Supply    string
	SupplyCap string

	// NFT
	NftHash     string
	NftMetadata string

	// Other
	Owner             string
	OwnerEditDisabled bool
	Invalidated       bool
	TokenRate         string
	TokenType         string
	MintingFee        string

	// Proposal fields
	Title       string
	Description string
}

// ProposalUpsertRate creates a proposal to upsert a token rate.
func (m *Module) ProposalUpsertRate(ctx context.Context, from string, propOpts *ProposalUpsertRateOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	boolFlags := make(map[string]bool)

	// Set required defaults (sekaid requires these even though help shows defaults)
	flags["fee_rate"] = "1.0"
	flags["token_rate"] = "1.0"
	flags["stake_cap"] = "0.1"
	flags["stake_min"] = "1"
	flags["supply"] = "0"
	flags["supply_cap"] = "0"
	flags["minting_fee"] = "0"
	boolFlags["fee_payments"] = true

	if propOpts != nil {
		if propOpts.Denom != "" {
			flags["denom"] = propOpts.Denom
		}
		if propOpts.Decimals > 0 {
			flags["decimals"] = fmt.Sprintf("%d", propOpts.Decimals)
		}
		if propOpts.FeeRate != "" {
			flags["fee_rate"] = propOpts.FeeRate
		}
		// FeePayments is already set to true by default above
		if propOpts.Name != "" {
			flags["name"] = propOpts.Name
		}
		if propOpts.Symbol != "" {
			flags["symbol"] = propOpts.Symbol
		}
		if propOpts.Icon != "" {
			flags["icon"] = propOpts.Icon
		}
		if propOpts.Website != "" {
			flags["website"] = propOpts.Website
		}
		if propOpts.Social != "" {
			flags["social"] = propOpts.Social
		}
		if propOpts.StakeToken {
			boolFlags["stake_token"] = true
		}
		if propOpts.StakeCap != "" {
			flags["stake_cap"] = propOpts.StakeCap
		}
		if propOpts.StakeMin != "" {
			flags["stake_min"] = propOpts.StakeMin
		}
		if propOpts.Supply != "" {
			flags["supply"] = propOpts.Supply
		}
		if propOpts.SupplyCap != "" {
			flags["supply_cap"] = propOpts.SupplyCap
		}
		if propOpts.NftHash != "" {
			flags["nft_hash"] = propOpts.NftHash
		}
		if propOpts.NftMetadata != "" {
			flags["nft_metadata"] = propOpts.NftMetadata
		}
		if propOpts.Owner != "" {
			flags["owner"] = propOpts.Owner
		}
		if propOpts.OwnerEditDisabled {
			boolFlags["owner_edit_disabled"] = true
		}
		if propOpts.Invalidated {
			boolFlags["invalidated"] = true
		}
		if propOpts.TokenRate != "" {
			flags["token_rate"] = propOpts.TokenRate
		}
		if propOpts.TokenType != "" {
			flags["token_type"] = propOpts.TokenType
		}
		if propOpts.MintingFee != "" {
			flags["minting_fee"] = propOpts.MintingFee
		}
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "tokens",
		Action:           "proposal-upsert-rate",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		BoolFlags:        boolFlags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal upsert rate: %w", err)
	}
	return resp, nil
}

// ProposalUpdateTokensBlackWhiteOpts contains options for updating token blacklist/whitelist.
type ProposalUpdateTokensBlackWhiteOpts struct {
	IsBlacklist bool
	IsAdd       bool
	Tokens      []string
	Title       string
	Description string
}

// ProposalUpdateTokensBlackWhite creates a proposal to update token blacklist/whitelist.
func (m *Module) ProposalUpdateTokensBlackWhite(ctx context.Context, from string, propOpts *ProposalUpdateTokensBlackWhiteOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	boolFlags := make(map[string]bool)
	arrayFlags := make(map[string][]string)
	if propOpts != nil {
		// Must explicitly set is_blacklist (defaults to true in sekaid)
		boolFlags["is_blacklist"] = propOpts.IsBlacklist
		// Must explicitly set is_add (defaults to true in sekaid)
		boolFlags["is_add"] = propOpts.IsAdd
		if len(propOpts.Tokens) > 0 {
			arrayFlags["tokens"] = propOpts.Tokens
		}
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "tokens",
		Action:           "proposal-update-tokens-blackwhite",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		BoolFlags:        boolFlags,
		ArrayFlags:       arrayFlags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal update tokens blackwhite: %w", err)
	}
	return resp, nil
}
