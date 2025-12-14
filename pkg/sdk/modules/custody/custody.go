// Package custody provides custody module functionality.
package custody

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides custody query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new custody module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// Get queries custody assigned to an address.
func (m *Module) Get(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "custody",
		Endpoint: "get",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query custody: %w", err)
	}
	return resp.Data, nil
}

// Custodians queries custody custodians for an address.
func (m *Module) Custodians(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "custody",
		Endpoint: "custodians",
		RawArgs:  []string{"get", address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query custodians: %w", err)
	}
	return resp.Data, nil
}

// CustodiansPool queries the custody pool for an address.
func (m *Module) CustodiansPool(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "custody",
		Endpoint: "custodians",
		RawArgs:  []string{"pool", address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query custody pool: %w", err)
	}
	return resp.Data, nil
}

// Whitelist queries custody whitelist for an address.
func (m *Module) Whitelist(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "custody",
		Endpoint: "whitelist",
		RawArgs:  []string{"get", address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query whitelist: %w", err)
	}
	return resp.Data, nil
}

// Limits queries custody limits for an address.
func (m *Module) Limits(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "custody",
		Endpoint: "limits",
		RawArgs:  []string{"get", address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query limits: %w", err)
	}
	return resp.Data, nil
}
