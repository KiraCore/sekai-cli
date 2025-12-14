// Package upgrade provides upgrade module functionality.
package upgrade

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides upgrade query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new upgrade module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// Plan represents an upgrade plan.
type Plan struct {
	Name                  string     `json:"name"`
	Height                string     `json:"height,omitempty"`
	Info                  string     `json:"info,omitempty"`
	UpgradeTime           string     `json:"upgrade_time,omitempty"`
	OldChainID            string     `json:"old_chain_id,omitempty"`
	NewChainID            string     `json:"new_chain_id,omitempty"`
	RollbackChecksum      string     `json:"rollback_checksum,omitempty"`
	MaxEnrollmentDuration string     `json:"max_enrollment_duration,omitempty"`
	InstateUpgrade        bool       `json:"instate_upgrade,omitempty"`
	RebootRequired        bool       `json:"reboot_required,omitempty"`
	SkipHandler           bool       `json:"skip_handler,omitempty"`
	Proposer              string     `json:"proposer,omitempty"`
	Resources             []Resource `json:"resources,omitempty"`
}

// Resource represents an upgrade resource.
type Resource struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

// CurrentPlan queries the current upgrade plan.
func (m *Module) CurrentPlan(ctx context.Context) (*Plan, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "upgrade",
		Endpoint: "current-plan",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query current plan: %w", err)
	}

	var result struct {
		Plan *Plan `json:"plan"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse current plan: %w", err)
	}
	return result.Plan, nil
}

// NextPlan queries the next upgrade plan.
func (m *Module) NextPlan(ctx context.Context) (*Plan, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "upgrade",
		Endpoint: "next-plan",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query next plan: %w", err)
	}

	var result struct {
		Plan *Plan `json:"plan"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse next plan: %w", err)
	}
	return result.Plan, nil
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

// ProposalSetPlanOpts contains options for creating an upgrade plan proposal.
type ProposalSetPlanOpts struct {
	Name                  string
	Title                 string
	Description           string
	MinUpgradeTime        int64
	OldChainID            string
	NewChainID            string
	MaxEnrollmentDuration int64
	InstateUpgrade        bool
	RebootRequired        bool
}

// ProposalSetPlan creates a proposal to set an upgrade plan.
func (m *Module) ProposalSetPlan(ctx context.Context, from string, propOpts *ProposalSetPlanOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
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
		if propOpts.MinUpgradeTime > 0 {
			flags["min-upgrade-time"] = fmt.Sprintf("%d", propOpts.MinUpgradeTime)
		}
		if propOpts.OldChainID != "" {
			flags["old-chain-id"] = propOpts.OldChainID
		}
		if propOpts.NewChainID != "" {
			flags["new-chain-id"] = propOpts.NewChainID
		}
		if propOpts.MaxEnrollmentDuration > 0 {
			flags["max-enrollment-duration"] = fmt.Sprintf("%d", propOpts.MaxEnrollmentDuration)
		}
		if propOpts.InstateUpgrade {
			flags["instate-upgrade"] = "true"
		}
		if propOpts.RebootRequired {
			flags["reboot-required"] = "true"
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "upgrade",
		Action:           "proposal-set-plan",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal set plan: %w", err)
	}
	return resp, nil
}

// ProposalCancelPlanOpts contains options for canceling an upgrade plan proposal.
type ProposalCancelPlanOpts struct {
	Name        string
	Title       string
	Description string
}

// ProposalCancelPlan creates a proposal to cancel an upgrade plan.
func (m *Module) ProposalCancelPlan(ctx context.Context, from string, propOpts *ProposalCancelPlanOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
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
		Module:           "upgrade",
		Action:           "proposal-cancel-plan",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal cancel plan: %w", err)
	}
	return resp, nil
}
