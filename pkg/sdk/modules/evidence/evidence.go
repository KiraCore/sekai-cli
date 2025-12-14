// Package evidence provides customevidence module functionality.
package evidence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides evidence query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new evidence module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// AllEvidence queries all submitted evidence.
func (m *Module) AllEvidence(ctx context.Context) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customevidence",
		Endpoint: "",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query evidence: %w", err)
	}
	return resp.Data, nil
}

// Evidence queries evidence by hash.
func (m *Module) Evidence(ctx context.Context, hash string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customevidence",
		Endpoint: "",
		RawArgs:  []string{hash},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query evidence: %w", err)
	}
	return resp.Data, nil
}
