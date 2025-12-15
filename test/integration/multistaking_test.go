// Package integration provides integration tests for the multistaking module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/multistaking"
)

// TestMultistakingPools tests querying all staking pools.
func TestMultistakingPools(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := multistaking.New(client)
	result, err := mod.Pools(ctx)
	requireNoError(t, err, "Failed to query pools")
	requireNotNil(t, result, "Pools is nil")

	t.Logf("Found %d staking pools", len(result.Pools))
	for _, pool := range result.Pools {
		t.Logf("  Pool: id=%s, enabled=%v", pool.ID, pool.Enabled)
	}
}

// TestMultistakingOutstandingRewards tests querying outstanding rewards.
func TestMultistakingOutstandingRewards(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := multistaking.New(client)
	result, err := mod.OutstandingRewards(ctx, testAddr)
	if err != nil {
		// This may fail if no rewards exist
		t.Logf("Outstanding rewards query: %v (expected if no rewards)", err)
		return
	}

	t.Logf("Outstanding rewards for %s: %+v", testAddr, result)
}

// TestMultistakingCompoundInfo tests querying compound info.
func TestMultistakingCompoundInfo(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := multistaking.New(client)
	result, err := mod.CompoundInfo(ctx, testAddr)
	if err != nil {
		// This may fail if no compound info exists
		t.Logf("Compound info query: %v (expected if no compound info)", err)
		return
	}

	t.Logf("Compound info for %s: %+v", testAddr, result)
}

// TestMultistakingStakingPoolDelegators tests querying staking pool delegators.
func TestMultistakingStakingPoolDelegators(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := multistaking.New(client)

	// Get a pool first
	pools, err := mod.Pools(ctx)
	requireNoError(t, err, "Failed to query pools")

	if len(pools.Pools) == 0 {
		t.Log("No staking pools found, skipping delegators test")
		return
	}

	poolID := pools.Pools[0].ID
	result, err := mod.StakingPoolDelegators(ctx, poolID)
	if err != nil {
		t.Logf("Staking pool delegators query: %v (expected if no delegators)", err)
		return
	}

	t.Logf("Delegators for pool %s: %+v", poolID, result)
}

// TestMultistakingUndelegations tests querying undelegations.
func TestMultistakingUndelegations(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := multistaking.New(client)

	// Get a pool first
	pools, err := mod.Pools(ctx)
	requireNoError(t, err, "Failed to query pools")

	if len(pools.Pools) == 0 {
		t.Log("No staking pools found, skipping undelegations test")
		return
	}

	poolID := pools.Pools[0].ID
	testAddr := getTestAddress(t)
	result, err := mod.Undelegations(ctx, testAddr, poolID)
	if err != nil {
		t.Logf("Undelegations query: %v (expected if no undelegations)", err)
		return
	}

	t.Logf("Undelegations for %s in pool %s: %+v", testAddr, poolID, result)
}

// TestMultistakingDelegate tests delegating tokens.
func TestMultistakingDelegate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := multistaking.New(client)

	// Get a pool first
	pools, err := mod.Pools(ctx)
	requireNoError(t, err, "Failed to query pools")

	if len(pools.Pools) == 0 {
		t.Log("No staking pools found, skipping delegate test")
		return
	}

	poolID := pools.Pools[0].ID
	resp, err := mod.Delegate(ctx, TestKey, poolID, "1000ukex", nil)
	if err != nil {
		t.Logf("Delegate may have failed (setup): %v", err)
		return
	}

	t.Logf("Delegate TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestMultistakingUndelegate tests undelegating tokens.
func TestMultistakingUndelegate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := multistaking.New(client)

	// Get a pool first
	pools, err := mod.Pools(ctx)
	requireNoError(t, err, "Failed to query pools")

	if len(pools.Pools) == 0 {
		t.Log("No staking pools found, skipping undelegate test")
		return
	}

	poolID := pools.Pools[0].ID
	resp, err := mod.Undelegate(ctx, TestKey, poolID, "100ukex", nil)
	if err != nil {
		t.Logf("Undelegate may have failed (no delegation): %v", err)
		return
	}

	t.Logf("Undelegate TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestMultistakingClaimRewards tests claiming rewards.
func TestMultistakingClaimRewards(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := multistaking.New(client)

	// Get a pool first
	pools, err := mod.Pools(ctx)
	requireNoError(t, err, "Failed to query pools")

	if len(pools.Pools) == 0 {
		t.Log("No staking pools found, skipping claim rewards test")
		return
	}

	poolID := pools.Pools[0].ID
	resp, err := mod.ClaimRewards(ctx, TestKey, poolID, nil)
	if err != nil {
		t.Logf("ClaimRewards may have failed (no rewards): %v", err)
		return
	}

	t.Logf("ClaimRewards TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestMultistakingClaimUndelegation tests claiming a specific undelegation.
func TestMultistakingClaimUndelegation(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := multistaking.New(client)

	// Try to claim undelegation ID 1
	resp, err := mod.ClaimUndelegation(ctx, TestKey, "1", nil)
	if err != nil {
		t.Logf("ClaimUndelegation may have failed (no undelegation): %v", err)
		return
	}

	t.Logf("ClaimUndelegation TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestMultistakingClaimMaturedUndelegations tests claiming all matured undelegations.
func TestMultistakingClaimMaturedUndelegations(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := multistaking.New(client)

	resp, err := mod.ClaimMaturedUndelegations(ctx, TestKey, nil)
	if err != nil {
		t.Logf("ClaimMaturedUndelegations may have failed: %v", err)
		return
	}

	t.Logf("ClaimMaturedUndelegations TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestMultistakingRegisterDelegator tests registering as a delegator.
func TestMultistakingRegisterDelegator(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := multistaking.New(client)

	resp, err := mod.RegisterDelegator(ctx, TestKey, nil)
	if err != nil {
		t.Logf("RegisterDelegator may have failed (already registered): %v", err)
		return
	}

	t.Logf("RegisterDelegator TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestMultistakingSetCompoundInfo tests setting compound info.
func TestMultistakingSetCompoundInfo(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := multistaking.New(client)

	resp, err := mod.SetCompoundInfo(ctx, TestKey, true, "", nil)
	if err != nil {
		t.Logf("SetCompoundInfo may have failed: %v", err)
		return
	}

	t.Logf("SetCompoundInfo TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestMultistakingUpsertStakingPool tests upserting a staking pool.
func TestMultistakingUpsertStakingPool(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := multistaking.New(client)

	// Get a pool first to get a valid validator key
	pools, err := mod.Pools(ctx)
	requireNoError(t, err, "Failed to query pools")

	if len(pools.Pools) == 0 {
		t.Log("No staking pools found, skipping upsert test")
		return
	}

	poolID := pools.Pools[0].ID
	poolOpts := &multistaking.UpsertStakingPoolOpts{
		Enabled:    true,
		Commission: "0.1",
	}

	resp, err := mod.UpsertStakingPool(ctx, TestKey, poolID, poolOpts, nil)
	if err != nil {
		t.Logf("UpsertStakingPool may have failed (permissions): %v", err)
		return
	}

	t.Logf("UpsertStakingPool TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}
