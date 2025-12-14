// Package spending provides spending module functionality.
package spending

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides spending query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new spending module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// PoolNames queries all pool names.
func (m *Module) PoolNames(ctx context.Context) (*PoolNamesResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "spending",
		Endpoint: "pool-names",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query pool names: %w", err)
	}

	var result PoolNamesResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pool names: %w", err)
	}
	return &result, nil
}

// PoolByName queries a pool by name.
func (m *Module) PoolByName(ctx context.Context, name string) (*PoolByNameResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "spending",
		Endpoint: "pool-by-name",
		RawArgs:  []string{name},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query pool by name: %w", err)
	}

	var result PoolByNameResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pool: %w", err)
	}
	return &result, nil
}

// PoolProposals queries proposals for a pool.
func (m *Module) PoolProposals(ctx context.Context, poolName string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "spending",
		Endpoint: "pool-proposals",
		RawArgs:  []string{poolName},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query pool proposals: %w", err)
	}

	return resp.Data, nil
}

// PoolsByAccount queries pools where an account can claim.
func (m *Module) PoolsByAccount(ctx context.Context, address string) (*PoolsByAccountResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "spending",
		Endpoint: "pools-by-account",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query pools by account: %w", err)
	}

	var result PoolsByAccountResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pools: %w", err)
	}
	return &result, nil
}

// TxOptions contains common transaction options.
type TxOptions struct {
	Fees          string
	Gas           string
	GasAdjustment float64
	Memo          string
	BroadcastMode string
}

// ClaimSpendingPool claims from a spending pool.
func (m *Module) ClaimSpendingPool(ctx context.Context, from, poolName string, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := make(map[string]string)
	flags["name"] = poolName
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
		Module:           "spending",
		Action:           "claim-spending-pool",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to claim spending pool: %w", err)
	}
	return resp, nil
}

// DepositSpendingPool deposits into a spending pool.
func (m *Module) DepositSpendingPool(ctx context.Context, from, poolName, amount string, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := make(map[string]string)
	flags["name"] = poolName
	flags["amount"] = amount
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
		Module:           "spending",
		Action:           "deposit-spending-pool",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to deposit spending pool: %w", err)
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

// CreateSpendingPoolOpts contains options for creating a spending pool.
type CreateSpendingPoolOpts struct {
	Name              string
	ClaimStart        uint64
	ClaimEnd          uint64
	ClaimExpiry       uint64
	Rates             string
	VoteQuorum        string
	VotePeriod        uint64
	VoteEnactment     uint64
	Owners            string
	Beneficiaries     string
	DynamicRate       bool
	DynamicRatePeriod uint64
}

// CreateSpendingPool creates a new spending pool.
func (m *Module) CreateSpendingPool(ctx context.Context, from string, poolOpts *CreateSpendingPoolOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if poolOpts != nil {
		if poolOpts.Name != "" {
			flags["name"] = poolOpts.Name
		}
		if poolOpts.ClaimStart > 0 {
			flags["claim-start"] = fmt.Sprintf("%d", poolOpts.ClaimStart)
		}
		if poolOpts.ClaimEnd > 0 {
			flags["claim-end"] = fmt.Sprintf("%d", poolOpts.ClaimEnd)
		}
		if poolOpts.ClaimExpiry > 0 {
			flags["claim-expiry"] = fmt.Sprintf("%d", poolOpts.ClaimExpiry)
		}
		if poolOpts.Rates != "" {
			flags["rates"] = poolOpts.Rates
		}
		if poolOpts.VoteQuorum != "" {
			flags["vote-quorum"] = poolOpts.VoteQuorum
		}
		if poolOpts.VotePeriod > 0 {
			flags["vote-period"] = fmt.Sprintf("%d", poolOpts.VotePeriod)
		}
		if poolOpts.VoteEnactment > 0 {
			flags["vote-enactment"] = fmt.Sprintf("%d", poolOpts.VoteEnactment)
		}
		if poolOpts.Owners != "" {
			flags["owner-accounts"] = poolOpts.Owners
		}
		if poolOpts.Beneficiaries != "" {
			flags["beneficiary-accounts"] = poolOpts.Beneficiaries
		}
		if poolOpts.DynamicRate {
			flags["dynamic-rate"] = "true"
		}
		if poolOpts.DynamicRatePeriod > 0 {
			flags["dynamic-rate-period"] = fmt.Sprintf("%d", poolOpts.DynamicRatePeriod)
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "spending",
		Action:           "create-spending-pool",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create spending pool: %w", err)
	}
	return resp, nil
}

// RegisterSpendingPoolBeneficiary registers a beneficiary for a spending pool.
func (m *Module) RegisterSpendingPoolBeneficiary(ctx context.Context, from, poolName string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["name"] = poolName

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "spending",
		Action:           "register-spending-pool-beneficiary",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register spending pool beneficiary: %w", err)
	}
	return resp, nil
}

// ProposalUpdateSpendingPoolOpts contains options for updating a spending pool proposal.
type ProposalUpdateSpendingPoolOpts struct {
	Name                      string
	Title                     string
	Description               string
	ClaimStart                int32
	ClaimEnd                  int32
	Rates                     string
	VoteQuorum                int32
	VotePeriod                int32
	VoteEnactment             int32
	OwnerAccounts             string
	OwnerRoles                string
	BeneficiaryAccounts       string
	BeneficiaryRoles          string
	BeneficiaryAccountWeights string
	BeneficiaryRoleWeights    string
	DynamicRate               bool
	DynamicRatePeriod         string
}

// ProposalUpdateSpendingPool creates a proposal to update a spending pool.
func (m *Module) ProposalUpdateSpendingPool(ctx context.Context, from string, propOpts *ProposalUpdateSpendingPoolOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	boolFlags := make(map[string]bool)
	if propOpts != nil {
		if propOpts.Name != "" {
			flags["name"] = propOpts.Name
		}
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.ClaimStart != 0 {
			flags["claim-start"] = fmt.Sprintf("%d", propOpts.ClaimStart)
		}
		if propOpts.ClaimEnd != 0 {
			flags["claim-end"] = fmt.Sprintf("%d", propOpts.ClaimEnd)
		}
		if propOpts.Rates != "" {
			flags["rates"] = propOpts.Rates
		}
		if propOpts.VoteQuorum != 0 {
			flags["vote-quorum"] = fmt.Sprintf("%d", propOpts.VoteQuorum)
		}
		if propOpts.VotePeriod != 0 {
			flags["vote-period"] = fmt.Sprintf("%d", propOpts.VotePeriod)
		}
		if propOpts.VoteEnactment != 0 {
			flags["vote-enactment"] = fmt.Sprintf("%d", propOpts.VoteEnactment)
		}
		if propOpts.OwnerAccounts != "" {
			flags["owner-accounts"] = propOpts.OwnerAccounts
		}
		if propOpts.OwnerRoles != "" {
			flags["owner-roles"] = propOpts.OwnerRoles
		}
		if propOpts.BeneficiaryAccounts != "" {
			flags["beneficiary-accounts"] = propOpts.BeneficiaryAccounts
		}
		if propOpts.BeneficiaryRoles != "" {
			flags["beneficiary-roles"] = propOpts.BeneficiaryRoles
		}
		if propOpts.BeneficiaryAccountWeights != "" {
			flags["beneficiary-account-weights"] = propOpts.BeneficiaryAccountWeights
		}
		if propOpts.BeneficiaryRoleWeights != "" {
			flags["beneficiary-role-weights"] = propOpts.BeneficiaryRoleWeights
		}
		if propOpts.DynamicRate {
			boolFlags["dynamic-rate"] = true
		}
		if propOpts.DynamicRatePeriod != "" {
			flags["dynamic-rate-period"] = propOpts.DynamicRatePeriod
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "spending",
		Action:           "proposal-update-spending-pool",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		BoolFlags:        boolFlags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal update spending pool: %w", err)
	}
	return resp, nil
}

// ProposalSpendingPoolDistributionOpts contains options for pool distribution proposal.
type ProposalSpendingPoolDistributionOpts struct {
	Name        string
	Title       string
	Description string
}

// ProposalSpendingPoolDistribution creates a proposal to distribute from a spending pool.
func (m *Module) ProposalSpendingPoolDistribution(ctx context.Context, from string, propOpts *ProposalSpendingPoolDistributionOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Name != "" {
			flags["name"] = propOpts.Name
		}
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "spending",
		Action:           "proposal-spending-pool-distribution",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal spending pool distribution: %w", err)
	}
	return resp, nil
}

// ProposalSpendingPoolWithdrawOpts contains options for pool withdraw proposal.
type ProposalSpendingPoolWithdrawOpts struct {
	Name                string
	BeneficiaryAccounts string
	Amount              string
	Title               string
	Description         string
}

// ProposalSpendingPoolWithdraw creates a proposal to withdraw from a spending pool.
func (m *Module) ProposalSpendingPoolWithdraw(ctx context.Context, from string, propOpts *ProposalSpendingPoolWithdrawOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Name != "" {
			flags["name"] = propOpts.Name
		}
		if propOpts.BeneficiaryAccounts != "" {
			flags["beneficiary-accounts"] = propOpts.BeneficiaryAccounts
		}
		if propOpts.Amount != "" {
			flags["amount"] = propOpts.Amount
		}
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "spending",
		Action:           "proposal-spending-pool-withdraw",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal spending pool withdraw: %w", err)
	}
	return resp, nil
}
