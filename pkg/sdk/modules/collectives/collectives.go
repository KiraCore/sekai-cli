// Package collectives provides collectives module functionality.
package collectives

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides collectives query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new collectives module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// Collectives queries all staking collectives.
func (m *Module) Collectives(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "collectives",
		Endpoint: "collectives",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query collectives: %w", err)
	}
	return resp.Data, nil
}

// Collective queries a collective by name.
func (m *Module) Collective(ctx context.Context, name string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "collectives",
		Endpoint: "collective",
		RawArgs:  []string{name},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query collective: %w", err)
	}
	return resp.Data, nil
}

// CollectivesByAccount queries collectives by account.
func (m *Module) CollectivesByAccount(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "collectives",
		Endpoint: "collectives-by-account",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query collectives by account: %w", err)
	}
	return resp.Data, nil
}

// CollectivesProposals queries proposals for collectives.
func (m *Module) CollectivesProposals(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "collectives",
		Endpoint: "collectives-proposals",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query collectives proposals: %w", err)
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

// CreateCollective creates a new collective.
func (m *Module) CreateCollective(ctx context.Context, from, name, description string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["collective-name"] = name
	if description != "" {
		flags["collective-description"] = description
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "collectives",
		Action:           "create-collective",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create collective: %w", err)
	}
	return resp, nil
}

// ContributeCollective contributes to a collective.
func (m *Module) ContributeCollective(ctx context.Context, from, name, bonds string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["collective-name"] = name
	flags["bonds"] = bonds

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "collectives",
		Action:           "contribute-collective",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to contribute to collective: %w", err)
	}
	return resp, nil
}

// DonateCollective sets lock and donation for bonds on the collective.
// donationLock is a boolean flag to lock contribution on the collective.
func (m *Module) DonateCollective(ctx context.Context, from, name string, locking uint64, donation string, donationLock bool, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	boolFlags := make(map[string]bool)
	flags["collective-name"] = name
	if locking > 0 {
		flags["locking"] = fmt.Sprintf("%d", locking)
	}
	if donation != "" {
		flags["donation"] = donation
	}
	if donationLock {
		boolFlags["donation-lock"] = true
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "collectives",
		Action:           "donate-collective",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		BoolFlags:        boolFlags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to donate to collective: %w", err)
	}
	return resp, nil
}

// WithdrawCollective withdraws from a collective.
func (m *Module) WithdrawCollective(ctx context.Context, from, name, bonds string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["collective-name"] = name
	flags["bonds"] = bonds

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "collectives",
		Action:           "withdraw-collective",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to withdraw from collective: %w", err)
	}
	return resp, nil
}

// ProposalCollectiveUpdateOpts contains options for proposal-collective-update.
type ProposalCollectiveUpdateOpts struct {
	Title                 string
	Description           string
	CollectiveName        string
	CollectiveDescription string
	CollectiveStatus      string
}

// ProposalCollectiveUpdate creates a proposal to update a collective.
func (m *Module) ProposalCollectiveUpdate(ctx context.Context, from string, propOpts *ProposalCollectiveUpdateOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.CollectiveName != "" {
			flags["collective-name"] = propOpts.CollectiveName
		}
		if propOpts.CollectiveDescription != "" {
			flags["collective-description"] = propOpts.CollectiveDescription
		}
		if propOpts.CollectiveStatus != "" {
			flags["collective-status"] = propOpts.CollectiveStatus
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "collectives",
		Action:           "proposal-collective-update",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal collective update: %w", err)
	}
	return resp, nil
}

// ProposalRemoveCollectiveOpts contains options for proposal-remove-collective.
type ProposalRemoveCollectiveOpts struct {
	Title          string
	Description    string
	CollectiveName string
}

// ProposalRemoveCollective creates a proposal to remove a collective.
func (m *Module) ProposalRemoveCollective(ctx context.Context, from string, propOpts *ProposalRemoveCollectiveOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.CollectiveName != "" {
			flags["collective-name"] = propOpts.CollectiveName
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "collectives",
		Action:           "proposal-remove-collective",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal remove collective: %w", err)
	}
	return resp, nil
}

// ProposalSendDonationOpts contains options for proposal-send-donation.
type ProposalSendDonationOpts struct {
	Title          string
	Description    string
	CollectiveName string
	Address        string
	Amounts        string
}

// ProposalSendDonation creates a proposal to send donation from a collective.
func (m *Module) ProposalSendDonation(ctx context.Context, from string, propOpts *ProposalSendDonationOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.CollectiveName != "" {
			flags["collective-name"] = propOpts.CollectiveName
		}
		if propOpts.Address != "" {
			flags["addr"] = propOpts.Address
		}
		if propOpts.Amounts != "" {
			flags["amounts"] = propOpts.Amounts
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "collectives",
		Action:           "proposal-send-donation",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal send donation: %w", err)
	}
	return resp, nil
}
