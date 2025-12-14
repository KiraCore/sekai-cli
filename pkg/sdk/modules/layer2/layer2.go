// Package layer2 provides layer2 module functionality.
package layer2

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides layer2 query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new layer2 module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// AllDapps queries all dapps.
func (m *Module) AllDapps(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "layer2",
		Endpoint: "all-dapps",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query all dapps: %w", err)
	}
	return resp.Data, nil
}

// ExecutionRegistrar queries execution registrar for a dapp.
func (m *Module) ExecutionRegistrar(ctx context.Context, dappName string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "layer2",
		Endpoint: "execution-registrar",
		RawArgs:  []string{dappName},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query execution registrar: %w", err)
	}
	return resp.Data, nil
}

// TransferDapps queries transfer dapps.
func (m *Module) TransferDapps(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "layer2",
		Endpoint: "transfer-dapps",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query transfer dapps: %w", err)
	}
	return resp.Data, nil
}
