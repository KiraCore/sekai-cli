// Package bridge provides bridge module functionality.
package bridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides bridge query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new bridge module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// GetCosmosEthereum queries change from Cosmos to Ethereum for an address.
func (m *Module) GetCosmosEthereum(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "bridge",
		Endpoint: "get_cosmos_ethereum",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query cosmos to ethereum: %w", err)
	}
	return resp.Data, nil
}

// GetEthereumCosmos queries change from Ethereum to Cosmos for an address.
func (m *Module) GetEthereumCosmos(ctx context.Context, address string) (json.RawMessage, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "bridge",
		Endpoint: "get_ethereum_cosmos",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query ethereum to cosmos: %w", err)
	}
	return resp.Data, nil
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

// ChangeCosmosEthereum creates a new change request from Cosmos to Ethereum.
// Args: [cosmosAddress] [ethereumAddress] [amount]
func (m *Module) ChangeCosmosEthereum(ctx context.Context, from, cosmosAddress, ethereumAddress, amount string, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "bridge",
		Action:           "change_cosmos_ethereum",
		Args:             []string{cosmosAddress, ethereumAddress, amount},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create change cosmos to ethereum: %w", err)
	}
	return resp, nil
}

// ChangeEthereumCosmos creates a new change request from Ethereum to Cosmos.
// Args: [cosmosAddress] [ethereumTxHash] [amount]
func (m *Module) ChangeEthereumCosmos(ctx context.Context, from, cosmosAddress, ethereumTxHash, amount string, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "bridge",
		Action:           "change_ethereum_cosmos",
		Args:             []string{cosmosAddress, ethereumTxHash, amount},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create change ethereum to cosmos: %w", err)
	}
	return resp, nil
}
