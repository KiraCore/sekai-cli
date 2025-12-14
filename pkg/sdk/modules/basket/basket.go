// Package basket provides basket module functionality.
package basket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides basket query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new basket module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// TokenBaskets queries token baskets filtered by tokens list and derivatives_only flag.
func (m *Module) TokenBaskets(ctx context.Context, tokens string, derivativesOnly bool) (json.RawMessage, error) {
	derivFlag := "false"
	if derivativesOnly {
		derivFlag = "true"
	}
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "basket",
		Endpoint: "token-baskets",
		RawArgs:  []string{tokens, derivFlag},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query token baskets: %w", err)
	}
	return resp.Data, nil
}

// TokenBasketByID queries a token basket by ID.
func (m *Module) TokenBasketByID(ctx context.Context, id string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "basket",
		Endpoint: "token-basket-by-id",
		RawArgs:  []string{id},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query token basket by id: %w", err)
	}
	return resp.Data, nil
}

// TokenBasketByDenom queries a token basket by denom.
func (m *Module) TokenBasketByDenom(ctx context.Context, denom string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "basket",
		Endpoint: "token-basket-by-denom",
		RawArgs:  []string{denom},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query token basket by denom: %w", err)
	}
	return resp.Data, nil
}

// HistoricalMints queries historical mints.
func (m *Module) HistoricalMints(ctx context.Context, basketID string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "basket",
		Endpoint: "historical-mints",
		RawArgs:  []string{basketID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query historical mints: %w", err)
	}
	return resp.Data, nil
}

// HistoricalBurns queries historical burns.
func (m *Module) HistoricalBurns(ctx context.Context, basketID string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "basket",
		Endpoint: "historical-burns",
		RawArgs:  []string{basketID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query historical burns: %w", err)
	}
	return resp.Data, nil
}

// HistoricalSwaps queries historical swaps.
func (m *Module) HistoricalSwaps(ctx context.Context, basketID string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "basket",
		Endpoint: "historical-swaps",
		RawArgs:  []string{basketID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query historical swaps: %w", err)
	}
	return resp.Data, nil
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

// MintBasketTokens mints basket tokens.
func (m *Module) MintBasketTokens(ctx context.Context, from, basketID, depositCoins string, opts *TxOptions) (*sdk.TxResponse, error) {
	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "basket",
		Action:           "mint-basket-tokens",
		Args:             []string{basketID, depositCoins},
		Signer:           from,
		Flags:            buildTxFlags(opts),
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to mint basket tokens: %w", err)
	}
	return resp, nil
}

// BurnBasketTokens burns basket tokens.
func (m *Module) BurnBasketTokens(ctx context.Context, from, basketID, burnAmount string, opts *TxOptions) (*sdk.TxResponse, error) {
	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "basket",
		Action:           "burn-basket-tokens",
		Args:             []string{basketID, burnAmount},
		Signer:           from,
		Flags:            buildTxFlags(opts),
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to burn basket tokens: %w", err)
	}
	return resp, nil
}

// SwapBasketTokens swaps basket tokens.
func (m *Module) SwapBasketTokens(ctx context.Context, from, basketID, swapIn, swapOut string, opts *TxOptions) (*sdk.TxResponse, error) {
	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "basket",
		Action:           "swap-basket-tokens",
		Args:             []string{basketID, swapIn, swapOut},
		Signer:           from,
		Flags:            buildTxFlags(opts),
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to swap basket tokens: %w", err)
	}
	return resp, nil
}

// BasketClaimRewards claims rewards from a staking derivative basket.
func (m *Module) BasketClaimRewards(ctx context.Context, from, basketID string, opts *TxOptions) (*sdk.TxResponse, error) {
	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "basket",
		Action:           "basket-claim-rewards",
		Args:             []string{basketID},
		Signer:           from,
		Flags:            buildTxFlags(opts),
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to claim basket rewards: %w", err)
	}
	return resp, nil
}

// DisableBasketDeposits disables deposits for a basket.
func (m *Module) DisableBasketDeposits(ctx context.Context, from, basketID string, disabled bool, opts *TxOptions) (*sdk.TxResponse, error) {
	disabledStr := "false"
	if disabled {
		disabledStr = "true"
	}
	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "basket",
		Action:           "disable-basket-deposits",
		Args:             []string{basketID, disabledStr},
		Signer:           from,
		Flags:            buildTxFlags(opts),
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to disable basket deposits: %w", err)
	}
	return resp, nil
}

// DisableBasketWithdraws disables withdraws for a basket.
func (m *Module) DisableBasketWithdraws(ctx context.Context, from, basketID string, disabled bool, opts *TxOptions) (*sdk.TxResponse, error) {
	disabledStr := "false"
	if disabled {
		disabledStr = "true"
	}
	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "basket",
		Action:           "disable-basket-withdraws",
		Args:             []string{basketID, disabledStr},
		Signer:           from,
		Flags:            buildTxFlags(opts),
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to disable basket withdraws: %w", err)
	}
	return resp, nil
}

// DisableBasketSwaps disables swaps for a basket.
func (m *Module) DisableBasketSwaps(ctx context.Context, from, basketID string, disabled bool, opts *TxOptions) (*sdk.TxResponse, error) {
	disabledStr := "false"
	if disabled {
		disabledStr = "true"
	}
	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "basket",
		Action:           "disable-basket-swaps",
		Args:             []string{basketID, disabledStr},
		Signer:           from,
		Flags:            buildTxFlags(opts),
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to disable basket swaps: %w", err)
	}
	return resp, nil
}

// ProposalCreateBasketOpts contains options for creating a basket proposal.
type ProposalCreateBasketOpts struct {
	// Basket fields
	BasketSuffix      string
	BasketDescription string
	BasketTokens      string
	TokensCap         string
	// Proposal fields
	Title       string
	Description string
	// Mints
	MintsMin      string
	MintsMax      string
	MintsDisabled bool
	// Burns
	BurnsMin      string
	BurnsMax      string
	BurnsDisabled bool
	// Swaps
	SwapsMin       string
	SwapsMax       string
	SwapsDisabled  bool
	SwapFee        string
	SlippageFeeMin string
	// Limits
	LimitsPeriod uint64
}

// ProposalCreateBasket creates a proposal to create a basket.
func (m *Module) ProposalCreateBasket(ctx context.Context, from string, propOpts *ProposalCreateBasketOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	boolFlags := make(map[string]bool)

	// Set required defaults (sekaid requires these)
	flags["swap-fee"] = "0.01"
	flags["slippage-fee-min"] = "0.01"
	flags["tokens-cap"] = "1000000000000"
	flags["mints-min"] = "1"
	flags["mints-max"] = "1000000000000"
	flags["burns-min"] = "1"
	flags["burns-max"] = "1000000000000"
	flags["swaps-min"] = "1"
	flags["swaps-max"] = "1000000000000"
	flags["limits-period"] = "86400"

	if propOpts != nil {
		if propOpts.BasketSuffix != "" {
			flags["basket-suffix"] = propOpts.BasketSuffix
		}
		if propOpts.BasketDescription != "" {
			flags["basket-description"] = propOpts.BasketDescription
		}
		if propOpts.BasketTokens != "" {
			flags["basket-tokens"] = propOpts.BasketTokens
		}
		if propOpts.TokensCap != "" {
			flags["tokens-cap"] = propOpts.TokensCap
		}
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.MintsMin != "" {
			flags["mints-min"] = propOpts.MintsMin
		}
		if propOpts.MintsMax != "" {
			flags["mints-max"] = propOpts.MintsMax
		}
		if propOpts.MintsDisabled {
			boolFlags["mints-disabled"] = true
		}
		if propOpts.BurnsMin != "" {
			flags["burns-min"] = propOpts.BurnsMin
		}
		if propOpts.BurnsMax != "" {
			flags["burns-max"] = propOpts.BurnsMax
		}
		if propOpts.BurnsDisabled {
			boolFlags["burns-disabled"] = true
		}
		if propOpts.SwapsMin != "" {
			flags["swaps-min"] = propOpts.SwapsMin
		}
		if propOpts.SwapsMax != "" {
			flags["swaps-max"] = propOpts.SwapsMax
		}
		if propOpts.SwapsDisabled {
			boolFlags["swaps-disabled"] = true
		}
		if propOpts.SwapFee != "" {
			flags["swap-fee"] = propOpts.SwapFee
		}
		if propOpts.SlippageFeeMin != "" {
			flags["slippage-fee-min"] = propOpts.SlippageFeeMin
		}
		if propOpts.LimitsPeriod > 0 {
			flags["limits-period"] = fmt.Sprintf("%d", propOpts.LimitsPeriod)
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "basket",
		Action:           "proposal-create-basket",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		BoolFlags:        boolFlags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal create basket: %w", err)
	}
	return resp, nil
}

// ProposalEditBasketOpts contains options for editing a basket proposal.
type ProposalEditBasketOpts struct {
	// Basket fields
	BasketID          uint64
	BasketSuffix      string
	BasketDescription string
	BasketTokens      string
	TokensCap         string
	// Proposal fields
	Title       string
	Description string
	// Mints
	MintsMin      string
	MintsMax      string
	MintsDisabled bool
	// Burns
	BurnsMin      string
	BurnsMax      string
	BurnsDisabled bool
	// Swaps
	SwapsMin       string
	SwapsMax       string
	SwapsDisabled  bool
	SwapFee        string
	SlippageFeeMin string
	// Limits
	LimitsPeriod uint64
}

// ProposalEditBasket creates a proposal to edit a basket.
func (m *Module) ProposalEditBasket(ctx context.Context, from string, propOpts *ProposalEditBasketOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	boolFlags := make(map[string]bool)

	// Set required defaults (sekaid requires these)
	flags["swap-fee"] = "0.01"
	flags["slippage-fee-min"] = "0.01"
	flags["tokens-cap"] = "1000000000000"
	flags["mints-min"] = "1"
	flags["mints-max"] = "1000000000000"
	flags["burns-min"] = "1"
	flags["burns-max"] = "1000000000000"
	flags["swaps-min"] = "1"
	flags["swaps-max"] = "1000000000000"
	flags["limits-period"] = "86400"

	if propOpts != nil {
		if propOpts.BasketID > 0 {
			flags["basket-id"] = fmt.Sprintf("%d", propOpts.BasketID)
		}
		if propOpts.BasketSuffix != "" {
			flags["basket-suffix"] = propOpts.BasketSuffix
		}
		if propOpts.BasketDescription != "" {
			flags["basket-description"] = propOpts.BasketDescription
		}
		if propOpts.BasketTokens != "" {
			flags["basket-tokens"] = propOpts.BasketTokens
		}
		if propOpts.TokensCap != "" {
			flags["tokens-cap"] = propOpts.TokensCap
		}
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.MintsMin != "" {
			flags["mints-min"] = propOpts.MintsMin
		}
		if propOpts.MintsMax != "" {
			flags["mints-max"] = propOpts.MintsMax
		}
		if propOpts.MintsDisabled {
			boolFlags["mints-disabled"] = true
		}
		if propOpts.BurnsMin != "" {
			flags["burns-min"] = propOpts.BurnsMin
		}
		if propOpts.BurnsMax != "" {
			flags["burns-max"] = propOpts.BurnsMax
		}
		if propOpts.BurnsDisabled {
			boolFlags["burns-disabled"] = true
		}
		if propOpts.SwapsMin != "" {
			flags["swaps-min"] = propOpts.SwapsMin
		}
		if propOpts.SwapsMax != "" {
			flags["swaps-max"] = propOpts.SwapsMax
		}
		if propOpts.SwapsDisabled {
			boolFlags["swaps-disabled"] = true
		}
		if propOpts.SwapFee != "" {
			flags["swap-fee"] = propOpts.SwapFee
		}
		if propOpts.SlippageFeeMin != "" {
			flags["slippage-fee-min"] = propOpts.SlippageFeeMin
		}
		if propOpts.LimitsPeriod > 0 {
			flags["limits-period"] = fmt.Sprintf("%d", propOpts.LimitsPeriod)
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "basket",
		Action:           "proposal-edit-basket",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		BoolFlags:        boolFlags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal edit basket: %w", err)
	}
	return resp, nil
}

// ProposalWithdrawSurplusOpts contains options for withdrawing surplus proposal.
type ProposalWithdrawSurplusOpts struct {
	Title       string
	Description string
}

// ProposalWithdrawSurplus creates a proposal to withdraw surplus from a basket.
// basketIDs is a comma-separated list of basket IDs.
// withdrawTarget is the address to withdraw to.
func (m *Module) ProposalWithdrawSurplus(ctx context.Context, from, basketIDs, withdrawTarget string, propOpts *ProposalWithdrawSurplusOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "basket",
		Action:           "proposal-basket-withdraw-surplus",
		Args:             []string{basketIDs, withdrawTarget},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal withdraw surplus: %w", err)
	}
	return resp, nil
}
