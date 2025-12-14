// Package integration provides integration tests for the upgrade module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/upgrade"
)

// TestUpgradeCurrentPlan tests querying the current upgrade plan.
func TestUpgradeCurrentPlan(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := upgrade.New(client)
	result, err := mod.CurrentPlan(ctx)
	requireNoError(t, err, "Failed to query current plan")

	// Result can be nil if no upgrade is planned, which is normal
	if result == nil {
		t.Log("No current upgrade plan")
		return
	}

	t.Logf("Current upgrade plan: Name=%s, Height=%s", result.Name, result.Height)
}

// TestUpgradeNextPlan tests querying the next upgrade plan.
func TestUpgradeNextPlan(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := upgrade.New(client)
	result, err := mod.NextPlan(ctx)
	requireNoError(t, err, "Failed to query next plan")

	// Result can be nil if no upgrade is planned, which is normal
	if result == nil {
		t.Log("No next upgrade plan")
		return
	}

	t.Logf("Next upgrade plan: Name=%s, Height=%s", result.Name, result.Height)
}

// TestUpgradeProposalSetPlan tests creating a proposal to set an upgrade plan.
func TestUpgradeProposalSetPlan(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := upgrade.New(client)

	// Create a proposal to set an upgrade plan
	propOpts := &upgrade.ProposalSetPlanOpts{
		Name:        generateUniqueID("upgrade"),
		Title:       "Test upgrade plan proposal",
		Description: "Integration test - verify upgrade plan proposal submission",
		OldChainID:  "testnet-1",
		NewChainID:  "testnet-2",
	}

	resp, err := mod.ProposalSetPlan(ctx, TestKey, propOpts, nil)
	requireNoError(t, err, "Failed to submit set plan proposal")
	requireNotNil(t, resp, "Response should not be nil")

	t.Logf("ProposalSetPlan TX: hash=%s, code=%d", resp.TxHash, resp.Code)
	requireTrue(t, resp.TxHash != "", "TX hash should not be empty")
}

// TestUpgradeProposalCancelPlan tests creating a proposal to cancel an upgrade plan.
func TestUpgradeProposalCancelPlan(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	// Wait for any previous TX to be confirmed
	waitForBlocks(t, 1)

	mod := upgrade.New(client)

	// Create a proposal to cancel an upgrade plan
	// Note: This may fail if there's no active plan, but we verify TX submission works
	propOpts := &upgrade.ProposalCancelPlanOpts{
		Name:        "upgrade1", // Default plan name
		Title:       "Test cancel upgrade plan proposal",
		Description: "Integration test - verify cancel plan proposal submission",
	}

	resp, err := mod.ProposalCancelPlan(ctx, TestKey, propOpts, nil)
	requireNoError(t, err, "Failed to submit cancel plan proposal")
	requireNotNil(t, resp, "Response should not be nil")

	t.Logf("ProposalCancelPlan TX: hash=%s, code=%d", resp.TxHash, resp.Code)
	requireTrue(t, resp.TxHash != "", "TX hash should not be empty")
}
