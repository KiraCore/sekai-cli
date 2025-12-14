// Package integration provides integration tests for the status module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/status"
)

// TestStatusNodeInfo tests querying node info.
func TestStatusNodeInfo(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := status.New(client)
	result, err := mod.NodeInfo(ctx)
	requireNoError(t, err, "Failed to query node info")
	requireNotNil(t, result, "Node info is nil")

	// Verify node info has expected fields
	requireTrue(t, result.Network != "", "Network should not be empty")
	requireEqual(t, TestChainID, result.Network, "Chain ID mismatch")

	t.Logf("Node: %s, Network: %s, Version: %s", result.Moniker, result.Network, result.Version)
}

// TestStatusSyncInfo tests querying sync info.
func TestStatusSyncInfo(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := status.New(client)
	result, err := mod.SyncInfo(ctx)
	requireNoError(t, err, "Failed to query sync info")
	requireNotNil(t, result, "Sync info is nil")

	// Verify sync info has expected fields
	requireTrue(t, result.LatestBlockHeight > 0, "Block height should be positive")
	requireTrue(t, result.LatestBlockTime != "", "Block time should not be empty")

	t.Logf("Latest block: %d, Time: %s, Syncing: %v", result.LatestBlockHeight, result.LatestBlockTime, result.CatchingUp)
}

// TestStatusValidatorInfo tests querying validator info.
func TestStatusValidatorInfo(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := status.New(client)
	result, err := mod.ValidatorInfo(ctx)
	requireNoError(t, err, "Failed to query validator info")
	requireNotNil(t, result, "Validator info is nil")

	// The test node should be a validator
	requireTrue(t, result.Address != "", "Validator address should not be empty")
	requireTrue(t, result.VotingPower >= 0, "Voting power should not be negative")

	t.Logf("Validator address: %s, Voting power: %d", result.Address, result.VotingPower)
}

// TestStatusChainID tests querying chain ID.
func TestStatusChainID(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := status.New(client)
	chainID, err := mod.ChainID(ctx)
	requireNoError(t, err, "Failed to query chain ID")
	requireEqual(t, TestChainID, chainID, "Chain ID mismatch")

	t.Logf("Chain ID: %s", chainID)
}

// TestStatusLatestBlockHeight tests querying latest block height.
func TestStatusLatestBlockHeight(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := status.New(client)
	height, err := mod.LatestBlockHeight(ctx)
	requireNoError(t, err, "Failed to query latest block height")
	requireTrue(t, height > 0, "Block height should be positive")

	t.Logf("Latest block height: %d", height)
}

// TestStatusIsSyncing tests querying sync status.
func TestStatusIsSyncing(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := status.New(client)
	syncing, err := mod.IsSyncing(ctx)
	requireNoError(t, err, "Failed to query sync status")

	t.Logf("Is syncing: %v", syncing)
}

// TestStatusNetworkProperties tests querying network properties.
func TestStatusNetworkProperties(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := status.New(client)
	props, err := mod.NetworkProperties(ctx)
	requireNoError(t, err, "Failed to query network properties")
	requireNotNil(t, props, "Network properties is nil")

	// Verify some expected fields are present
	requireTrue(t, props.MinTxFee != "", "MinTxFee should not be empty")
	requireTrue(t, props.MaxTxFee != "", "MaxTxFee should not be empty")
	requireTrue(t, props.VoteQuorum != "", "VoteQuorum should not be empty")

	t.Logf("Min TX Fee: %s, Max TX Fee: %s, Vote Quorum: %s", props.MinTxFee, props.MaxTxFee, props.VoteQuorum)
}

// TestStatusFull tests the full Status query.
func TestStatusFull(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := status.New(client)
	result, err := mod.Status(ctx)
	requireNoError(t, err, "Failed to query full status")
	requireNotNil(t, result, "Status is nil")

	// Verify all sections are present
	requireTrue(t, result.NodeInfo.Network != "", "NodeInfo.Network should not be empty")
	requireTrue(t, result.SyncInfo.LatestBlockHeight > 0, "SyncInfo.LatestBlockHeight should be positive")

	t.Logf("Full status: Network=%s, Height=%d, Catching up=%v",
		result.NodeInfo.Network,
		result.SyncInfo.LatestBlockHeight,
		result.SyncInfo.CatchingUp)
}
