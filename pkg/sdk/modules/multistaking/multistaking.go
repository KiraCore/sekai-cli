// Package multistaking provides multistaking module functionality.
package multistaking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides multistaking query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new multistaking module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// Pools queries all staking pools.
func (m *Module) Pools(ctx context.Context) (*PoolsResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "multistaking",
		Endpoint: "pools",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query pools: %w", err)
	}

	var result PoolsResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pools: %w", err)
	}
	return &result, nil
}

// Undelegations queries all undelegations for a delegator and validator.
func (m *Module) Undelegations(ctx context.Context, delegator, valAddr string) (*UndelegationsResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "multistaking",
		Endpoint: "undelegations",
		RawArgs:  []string{delegator, valAddr},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query undelegations: %w", err)
	}

	var result UndelegationsResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse undelegations: %w", err)
	}
	return &result, nil
}

// OutstandingRewards queries outstanding rewards for a delegator.
func (m *Module) OutstandingRewards(ctx context.Context, delegator string) (*OutstandingRewards, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "multistaking",
		Endpoint: "outstanding-rewards",
		RawArgs:  []string{delegator},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query outstanding rewards: %w", err)
	}

	var result OutstandingRewards
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse outstanding rewards: %w", err)
	}
	return &result, nil
}

// CompoundInfo queries compound information of a delegator.
func (m *Module) CompoundInfo(ctx context.Context, delegator string) (*CompoundInfo, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "multistaking",
		Endpoint: "compound-info",
		RawArgs:  []string{delegator},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query compound info: %w", err)
	}

	var result CompoundInfo
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse compound info: %w", err)
	}
	return &result, nil
}

// StakingPoolDelegators queries staking pool delegators for a validator.
func (m *Module) StakingPoolDelegators(ctx context.Context, validator string) ([]StakingPoolDelegator, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "multistaking",
		Endpoint: "staking-pool-delegators",
		RawArgs:  []string{validator},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query staking pool delegators: %w", err)
	}

	var result struct {
		Delegators []StakingPoolDelegator `json:"delegators"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse staking pool delegators: %w", err)
	}
	return result.Delegators, nil
}

// TxOptions contains common transaction options.
type TxOptions struct {
	Fees          string
	Gas           string
	GasAdjustment float64
	Memo          string
	BroadcastMode string
}

// Delegate delegates tokens to a validator pool.
func (m *Module) Delegate(ctx context.Context, from, validator, coins string, opts *TxOptions) (*sdk.TxResponse, error) {
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

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "multistaking",
		Action:           "delegate",
		Args:             []string{validator, coins},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delegate: %w", err)
	}
	return resp, nil
}

// Undelegate starts undelegation from a validator pool.
func (m *Module) Undelegate(ctx context.Context, from, validator, coins string, opts *TxOptions) (*sdk.TxResponse, error) {
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

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "multistaking",
		Action:           "undelegate",
		Args:             []string{validator, coins},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to undelegate: %w", err)
	}
	return resp, nil
}

// ClaimRewards claims rewards from a validator pool.
func (m *Module) ClaimRewards(ctx context.Context, from, validator string, opts *TxOptions) (*sdk.TxResponse, error) {
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

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "multistaking",
		Action:           "claim-rewards",
		Args:             []string{validator},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to claim rewards: %w", err)
	}
	return resp, nil
}

// ClaimUndelegation claims a matured undelegation.
func (m *Module) ClaimUndelegation(ctx context.Context, from string, undelegationID string, opts *TxOptions) (*sdk.TxResponse, error) {
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

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "multistaking",
		Action:           "claim-undelegation",
		Args:             []string{undelegationID},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to claim undelegation: %w", err)
	}
	return resp, nil
}

// ClaimMaturedUndelegations claims all matured undelegations.
func (m *Module) ClaimMaturedUndelegations(ctx context.Context, from string, opts *TxOptions) (*sdk.TxResponse, error) {
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

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "multistaking",
		Action:           "claim-matured-undelegations",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to claim matured undelegations: %w", err)
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

// RegisterDelegator registers an account as a delegator.
func (m *Module) RegisterDelegator(ctx context.Context, from string, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "multistaking",
		Action:           "register-delegator",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register delegator: %w", err)
	}
	return resp, nil
}

// SetCompoundInfo sets compound information for a delegator.
// allDenom indicates whether to apply to all denoms.
// specificDenoms is a comma-separated list of specific denoms.
func (m *Module) SetCompoundInfo(ctx context.Context, from string, allDenom bool, specificDenoms string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)

	allDenomStr := "false"
	if allDenom {
		allDenomStr = "true"
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "multistaking",
		Action:           "set-compound-info",
		Args:             []string{allDenomStr, specificDenoms},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set compound info: %w", err)
	}
	return resp, nil
}

// UpsertStakingPoolOpts contains options for upserting a staking pool.
type UpsertStakingPoolOpts struct {
	Enabled    bool
	Commission string
}

// UpsertStakingPool creates or updates a staking pool.
// validatorKey is the validator key for the pool.
func (m *Module) UpsertStakingPool(ctx context.Context, from, validatorKey string, poolOpts *UpsertStakingPoolOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	boolFlags := make(map[string]bool)
	if poolOpts != nil {
		if poolOpts.Enabled {
			boolFlags["enabled"] = true
		}
		if poolOpts.Commission != "" {
			flags["commission"] = poolOpts.Commission
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "multistaking",
		Action:           "upsert-staking-pool",
		Args:             []string{validatorKey},
		Signer:           from,
		Flags:            flags,
		BoolFlags:        boolFlags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert staking pool: %w", err)
	}
	return resp, nil
}
