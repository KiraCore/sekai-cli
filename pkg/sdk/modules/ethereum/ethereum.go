// Package ethereum provides ethereum module functionality.
package ethereum

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides ethereum query and tx functionality.
type Module struct {
	client sdk.Client
}

// New creates a new ethereum module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// State queries the ethereum module state.
func (m *Module) State(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "ethereum",
		Endpoint: "",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query ethereum state: %w", err)
	}
	return resp.Data, nil
}
