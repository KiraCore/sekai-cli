package sdk

import (
	"context"
	"sync"
)

// SEKAI is the main SDK entry point.
// It provides access to all blockchain modules through a unified interface.
type SEKAI struct {
	client Client

	// Module instances (lazily initialized)
	mu sync.RWMutex

	// Cached module instances
	bankModule   interface{}
	govModule    interface{}
	keysModule   interface{}
	statusModule interface{}

	// Additional modules will be added as they are implemented
	// stakingModule      interface{}
	// multistakingModule interface{}
	// tokensModule       interface{}
	// spendingModule     interface{}
	// ubiModule          interface{}
	// basketModule       interface{}
	// collectivesModule  interface{}
	// custodyModule      interface{}
	// bridgeModule       interface{}
	// layer2Module       interface{}
	// recoveryModule     interface{}
	// upgradeModule      interface{}
}

// New creates a new SEKAI SDK instance with the given client.
//
// Example usage with Docker client:
//
//	client, err := docker.NewClient("sekai-container")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	sekai := sdk.New(client)
//	balance, err := sekai.Bank().Balance(ctx, "kira1...", "ukex")
//
// Example usage with REST client:
//
//	client, err := rest.NewClient("http://localhost:1317")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	sekai := sdk.New(client)
//	status, err := sekai.Status(ctx)
func New(client Client) *SEKAI {
	return &SEKAI{
		client: client,
	}
}

// Client returns the underlying client.
// This can be used for advanced operations not covered by the SDK modules.
func (s *SEKAI) Client() Client {
	return s.client
}

// Close releases all resources held by the SDK.
func (s *SEKAI) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Status returns the node status.
// This is a convenience method that delegates to the underlying client.
func (s *SEKAI) Status(ctx context.Context) (*StatusResponse, error) {
	return s.client.Status(ctx)
}

// Keys returns the keys client for key management operations.
// Note: This may have limited functionality for non-Docker clients.
func (s *SEKAI) Keys() KeysClient {
	return s.client.Keys()
}

// Module accessor methods will be added here as modules are implemented.
// They follow the pattern:
//
// func (s *SEKAI) ModuleName() *modulename.Module {
//     s.mu.RLock()
//     if s.moduleNameModule != nil {
//         s.mu.RUnlock()
//         return s.moduleNameModule.(*modulename.Module)
//     }
//     s.mu.RUnlock()
//
//     s.mu.Lock()
//     defer s.mu.Unlock()
//     if s.moduleNameModule == nil {
//         s.moduleNameModule = modulename.New(s.client)
//     }
//     return s.moduleNameModule.(*modulename.Module)
// }

// Version returns the SDK version.
func Version() string {
	return "0.1.0"
}

// ChainID constants for common KIRA networks.
const (
	// ChainIDMainnet is the mainnet chain ID.
	ChainIDMainnet = "kira-1"

	// ChainIDTestnet is the testnet chain ID.
	ChainIDTestnet = "testnet-1"

	// ChainIDLocalnet is the default localnet chain ID.
	ChainIDLocalnet = "localnet-1"
)

// DefaultDenom is the default token denomination for KIRA.
const DefaultDenom = "ukex"

// DefaultFees is the default transaction fee.
const DefaultFees = "100ukex"

// DefaultGasAdjustment is the default gas adjustment factor.
const DefaultGasAdjustment = 1.3
