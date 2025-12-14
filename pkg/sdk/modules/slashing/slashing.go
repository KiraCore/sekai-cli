// Package slashing provides customslashing module functionality.
package slashing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides slashing query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new slashing module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// SigningInfo represents validator signing info.
type SigningInfo struct {
	Address               string `json:"address"`
	StartHeight           string `json:"start_height"`
	IndexOffset           string `json:"index_offset"`
	JailedUntil           string `json:"jailed_until"`
	Tombstoned            bool   `json:"tombstoned"`
	MissedBlocksCounter   string `json:"missed_blocks_counter"`
	ProducedBlocksCounter string `json:"produced_blocks_counter"`
	MischanceConfidence   string `json:"mischance_confidence"`
	Mischance             string `json:"mischance"`
	LastPresentBlock      string `json:"last_present_block"`
}

// SigningInfoResponse contains the signing info response.
type SigningInfoResponse struct {
	ValSigningInfo SigningInfo `json:"val_signing_info"`
}

// SigningInfosResponse contains all signing infos.
type SigningInfosResponse struct {
	Info []SigningInfo `json:"info"`
}

// SigningInfo queries signing info for a validator.
func (m *Module) SigningInfo(ctx context.Context, consAddress string) (*SigningInfoResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customslashing",
		Endpoint: "signing-info",
		RawArgs:  []string{consAddress},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query signing info: %w", err)
	}

	// The API returns signing info directly (not wrapped in val_signing_info)
	var info SigningInfo
	if err := json.Unmarshal(resp.Data, &info); err != nil {
		// Try wrapped format as fallback
		var result SigningInfoResponse
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			return nil, fmt.Errorf("failed to parse signing info: %w", err)
		}
		return &result, nil
	}
	return &SigningInfoResponse{ValSigningInfo: info}, nil
}

// SigningInfos queries signing infos for all validators.
func (m *Module) SigningInfos(ctx context.Context) (*SigningInfosResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customslashing",
		Endpoint: "signing-infos",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query signing infos: %w", err)
	}

	var result SigningInfosResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse signing infos: %w", err)
	}
	return &result, nil
}

// StakingPoolsResponse contains staking pools response.
type StakingPoolsResponse struct {
	Pools json.RawMessage `json:"pools"`
}

// ActiveStakingPools queries active staking pools.
func (m *Module) ActiveStakingPools(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customslashing",
		Endpoint: "active-staking-pools",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query active staking pools: %w", err)
	}
	return resp.Data, nil
}

// InactiveStakingPools queries inactive staking pools.
func (m *Module) InactiveStakingPools(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customslashing",
		Endpoint: "inactive-staking-pools",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query inactive staking pools: %w", err)
	}
	return resp.Data, nil
}

// SlashedStakingPools queries slashed staking pools.
func (m *Module) SlashedStakingPools(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customslashing",
		Endpoint: "slashed-staking-pools",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query slashed staking pools: %w", err)
	}
	return resp.Data, nil
}

// SlashProposals queries slash proposals.
func (m *Module) SlashProposals(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customslashing",
		Endpoint: "slash-proposals",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query slash proposals: %w", err)
	}
	return resp.Data, nil
}
