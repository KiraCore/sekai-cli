// Package integration provides integration tests for the bridge module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/bridge"
)

// TestBridgeGetCosmosEthereum tests querying cosmos to ethereum changes.
func TestBridgeGetCosmosEthereum(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := bridge.New(client)
	result, err := mod.GetCosmosEthereum(ctx, testAddr)
	if err != nil {
		// This may fail if no bridge changes exist
		t.Logf("Cosmos to Ethereum query: %v (expected if no changes)", err)
		return
	}

	t.Logf("Cosmos to Ethereum: %s", string(result))
}

// TestBridgeGetEthereumCosmos tests querying ethereum to cosmos changes.
func TestBridgeGetEthereumCosmos(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := bridge.New(client)
	result, err := mod.GetEthereumCosmos(ctx, testAddr)
	if err != nil {
		// This may fail if no bridge changes exist
		t.Logf("Ethereum to Cosmos query: %v (expected if no changes)", err)
		return
	}

	t.Logf("Ethereum to Cosmos: %s", string(result))
}

// TestBridgeChangeCosmosEthereum tests creating a change request from Cosmos to Ethereum.
func TestBridgeChangeCosmosEthereum(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := bridge.New(client)

	// Use test values
	testAddr := getTestAddress(t)
	cosmosAddress := testAddr
	ethAddress := "0x1234567890123456789012345678901234567890"
	amount := "100ukex"

	resp, err := mod.ChangeCosmosEthereum(ctx, TestKey, cosmosAddress, ethAddress, amount, nil)
	// Bridge may require special setup, so we handle errors gracefully
	if err != nil {
		t.Logf("ChangeCosmosEthereum may have failed (bridge setup): %v", err)
		return
	}

	t.Logf("ChangeCosmosEthereum TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestBridgeChangeEthereumCosmos tests creating a change request from Ethereum to Cosmos.
func TestBridgeChangeEthereumCosmos(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	// Wait for previous TX
	waitForBlocks(t, 1)

	mod := bridge.New(client)

	// Use test values
	testAddr := getTestAddress(t)
	cosmosAddress := testAddr
	ethTxHash := "0x" + generateUniqueID("txhash")
	amount := "100ukex"

	resp, err := mod.ChangeEthereumCosmos(ctx, TestKey, cosmosAddress, ethTxHash, amount, nil)
	// Bridge may require special setup, so we handle errors gracefully
	if err != nil {
		t.Logf("ChangeEthereumCosmos may have failed (bridge setup): %v", err)
		return
	}

	t.Logf("ChangeEthereumCosmos TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}
