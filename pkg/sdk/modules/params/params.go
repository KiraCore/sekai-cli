// Package params provides params module functionality.
package params

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides params query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new params module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// Subspace queries raw parameters by subspace and key.
func (m *Module) Subspace(ctx context.Context, subspace, key string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "params",
		Endpoint: "subspace",
		RawArgs:  []string{subspace, key},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query params subspace: %w", err)
	}
	return resp.Data, nil
}
