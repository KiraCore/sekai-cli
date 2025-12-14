// Package ubi provides UBI module functionality.
package ubi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides UBI query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new UBI module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// UBIRecord represents a UBI record.
type UBIRecord struct {
	Name              string `json:"name"`
	DistributionStart string `json:"distribution_start"`
	DistributionEnd   string `json:"distribution_end"`
	DistributionLast  string `json:"distribution_last"`
	Amount            string `json:"amount"`
	Period            string `json:"period"`
	Pool              string `json:"pool"`
	Dynamic           bool   `json:"dynamic"`
}

// UBIRecordsResponse contains the UBI records query response.
type UBIRecordsResponse struct {
	Records []UBIRecord `json:"records"`
}

// Records queries all UBI records.
func (m *Module) Records(ctx context.Context) (*UBIRecordsResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "ubi",
		Endpoint: "ubi-records",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query ubi records: %w", err)
	}

	var result UBIRecordsResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ubi records: %w", err)
	}
	return &result, nil
}

// RecordByName queries a UBI record by name.
func (m *Module) RecordByName(ctx context.Context, name string) (*UBIRecord, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "ubi",
		Endpoint: "ubi-record-by-name",
		RawArgs:  []string{name},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query ubi record: %w", err)
	}

	var result struct {
		Record UBIRecord `json:"record"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ubi record: %w", err)
	}
	return &result.Record, nil
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

// ProposalUpsertUBIOpts contains options for creating a UBI upsert proposal.
type ProposalUpsertUBIOpts struct {
	Name        string
	DistrStart  uint64
	DistrEnd    uint64
	Amount      uint64
	Period      uint64
	PoolName    string
	Title       string
	Description string
}

// ProposalUpsertUBI creates a proposal to upsert a UBI record.
func (m *Module) ProposalUpsertUBI(ctx context.Context, from string, propOpts *ProposalUpsertUBIOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Name != "" {
			flags["name"] = propOpts.Name
		}
		if propOpts.DistrStart > 0 {
			flags["distr-start"] = fmt.Sprintf("%d", propOpts.DistrStart)
		}
		if propOpts.DistrEnd > 0 {
			flags["distr-end"] = fmt.Sprintf("%d", propOpts.DistrEnd)
		}
		if propOpts.Amount > 0 {
			flags["amount"] = fmt.Sprintf("%d", propOpts.Amount)
		}
		if propOpts.Period > 0 {
			flags["period"] = fmt.Sprintf("%d", propOpts.Period)
		}
		if propOpts.PoolName != "" {
			flags["pool-name"] = propOpts.PoolName
		}
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "ubi",
		Action:           "proposal-upsert-ubi",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal upsert ubi: %w", err)
	}
	return resp, nil
}

// ProposalRemoveUBIOpts contains options for creating a UBI remove proposal.
type ProposalRemoveUBIOpts struct {
	Name        string
	Title       string
	Description string
}

// ProposalRemoveUBI creates a proposal to remove a UBI record.
func (m *Module) ProposalRemoveUBI(ctx context.Context, from string, propOpts *ProposalRemoveUBIOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
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
		Module:           "ubi",
		Action:           "proposal-remove-ubi",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal remove ubi: %w", err)
	}
	return resp, nil
}
