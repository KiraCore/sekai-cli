// Package recovery provides recovery module functionality.
package recovery

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides recovery query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new recovery module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// RecoveryRecord queries recovery information for an account.
func (m *Module) RecoveryRecord(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "recovery",
		Endpoint: "recovery-record",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query recovery record: %w", err)
	}
	return resp.Data, nil
}

// RecoveryToken queries recovery token information for an account.
func (m *Module) RecoveryToken(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "recovery",
		Endpoint: "recovery-token",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query recovery token: %w", err)
	}
	return resp.Data, nil
}

// RRHolderRewards queries RR holder rewards for an account.
func (m *Module) RRHolderRewards(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "recovery",
		Endpoint: "rr-holder-rewards",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query rr holder rewards: %w", err)
	}
	return resp.Data, nil
}

// RRHolders queries registered RR holders for a token.
func (m *Module) RRHolders(ctx context.Context, rrToken string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "recovery",
		Endpoint: "rr-holders",
		RawArgs:  []string{rrToken},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query rr holders: %w", err)
	}
	return resp.Data, nil
}
