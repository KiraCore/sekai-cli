// Package staking provides customstaking module functionality.
package staking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides staking query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new staking module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// Validators queries validators with optional filters.
func (m *Module) Validators(ctx context.Context, opts *ValidatorQueryOpts) (*ValidatorsResponse, error) {
	params := make(map[string]string)
	if opts != nil {
		if opts.Address != "" {
			params["addr"] = opts.Address
		}
		if opts.ValAddr != "" {
			params["val-addr"] = opts.ValAddr
		}
		if opts.Moniker != "" {
			params["moniker"] = opts.Moniker
		}
		if opts.Status != "" {
			params["status"] = opts.Status
		}
		if opts.PubKey != "" {
			params["pubkey"] = opts.PubKey
		}
		if opts.Proposer != "" {
			params["proposer"] = opts.Proposer
		}
	}

	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customstaking",
		Endpoint: "validators",
		Params:   params,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query validators: %w", err)
	}

	var result ValidatorsResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse validators: %w", err)
	}
	return &result, nil
}

// Validator queries a single validator by address, val-address, or moniker.
func (m *Module) Validator(ctx context.Context, opts *ValidatorQueryOpts) (*Validator, error) {
	params := make(map[string]string)
	if opts != nil {
		if opts.Address != "" {
			params["addr"] = opts.Address
		}
		if opts.ValAddr != "" {
			params["val-addr"] = opts.ValAddr
		}
		if opts.Moniker != "" {
			params["moniker"] = opts.Moniker
		}
	}

	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customstaking",
		Endpoint: "validator",
		Params:   params,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query validator: %w", err)
	}

	// Try direct unmarshal first
	var val Validator
	if err := json.Unmarshal(resp.Data, &val); err == nil && (val.Status != "" || val.ValKey != "" || val.ValKeyAlt != "") {
		return &val, nil
	}

	// Try wrapped response
	var result struct {
		Validator Validator `json:"validator"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse validator: %w", err)
	}
	if result.Validator.Status != "" || result.Validator.ValKey != "" || result.Validator.ValKeyAlt != "" {
		return &result.Validator, nil
	}
	return &val, nil
}

// TxOptions contains common transaction options.
type TxOptions struct {
	Fees          string
	Gas           string
	GasAdjustment float64
	Memo          string
	BroadcastMode string
}

// ClaimValidatorSeatOpts contains options for claiming a validator seat.
type ClaimValidatorSeatOpts struct {
	Moniker string
	PubKey  string
}

// ClaimValidatorSeat claims a validator seat to become a validator.
func (m *Module) ClaimValidatorSeat(ctx context.Context, from string, seatOpts *ClaimValidatorSeatOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := make(map[string]string)
	if seatOpts != nil {
		if seatOpts.Moniker != "" {
			flags["moniker"] = seatOpts.Moniker
		}
		if seatOpts.PubKey != "" {
			flags["pubkey"] = seatOpts.PubKey
		}
	}
	if txOpts != nil {
		if txOpts.Fees != "" {
			flags["fees"] = txOpts.Fees
		}
		if txOpts.Gas != "" {
			flags["gas"] = txOpts.Gas
		}
		if txOpts.GasAdjustment > 0 {
			flags["gas-adjustment"] = fmt.Sprintf("%.2f", txOpts.GasAdjustment)
		}
		if txOpts.Memo != "" {
			flags["note"] = txOpts.Memo
		}
		if txOpts.BroadcastMode != "" {
			flags["broadcast-mode"] = txOpts.BroadcastMode
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customstaking",
		Action:           "claim-validator-seat",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to claim validator seat: %w", err)
	}
	return resp, nil
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

// ProposalUnjailValidatorOpts contains options for unjailing a validator proposal.
type ProposalUnjailValidatorOpts struct {
	Title       string
	Description string
}

// ProposalUnjailValidator creates a proposal to unjail a validator.
// valAddr is the validator address and reference is a reference string for the proposal.
func (m *Module) ProposalUnjailValidator(ctx context.Context, from, valAddr, reference string, propOpts *ProposalUnjailValidatorOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
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
		Module:           "customstaking",
		Action:           "proposal unjail-validator",
		Args:             []string{valAddr, reference},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal unjail validator: %w", err)
	}
	return resp, nil
}
