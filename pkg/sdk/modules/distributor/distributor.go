// Package distributor provides distributor module functionality.
package distributor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides distributor query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new distributor module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// FeesTreasury queries the fees treasury.
func (m *Module) FeesTreasury(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "distributor",
		Endpoint: "fees-treasury",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query fees treasury: %w", err)
	}
	return resp.Data, nil
}

// PeriodicSnapshot queries the periodic snapshot.
func (m *Module) PeriodicSnapshot(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "distributor",
		Endpoint: "periodic-snapshot",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query periodic snapshot: %w", err)
	}
	return resp.Data, nil
}

// SnapshotPeriod queries the snapshot period.
func (m *Module) SnapshotPeriod(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "distributor",
		Endpoint: "snapshot-period",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshot period: %w", err)
	}
	return resp.Data, nil
}

// SnapshotPeriodPerformance queries snapshot period performance for a validator.
func (m *Module) SnapshotPeriodPerformance(ctx context.Context, validator string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "distributor",
		Endpoint: "snapshot-period-performance",
		RawArgs:  []string{validator},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshot period performance: %w", err)
	}
	return resp.Data, nil
}

// YearStartSnapshot queries the year start snapshot.
func (m *Module) YearStartSnapshot(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "distributor",
		Endpoint: "year-start-snapshot",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query year start snapshot: %w", err)
	}
	return resp.Data, nil
}
