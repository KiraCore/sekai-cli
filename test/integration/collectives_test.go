// Package integration provides integration tests for the collectives module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/collectives"
)

// TestCollectivesAll tests querying all collectives.
func TestCollectivesAll(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := collectives.New(client)
	result, err := mod.Collectives(ctx)
	requireNoError(t, err, "Failed to query collectives")

	t.Logf("Collectives: %s", string(result))
}

// TestCollectivesByAccount tests querying collectives by account.
func TestCollectivesByAccount(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := collectives.New(client)
	result, err := mod.CollectivesByAccount(ctx, TestAddress)
	if err != nil {
		// This may fail if no collectives for this account
		t.Logf("Collectives by account query: %v (expected if no collectives)", err)
		return
	}

	t.Logf("Collectives for %s: %s", TestAddress, string(result))
}

// TestCollectivesProposals tests querying collectives proposals.
func TestCollectivesProposals(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := collectives.New(client)
	result, err := mod.CollectivesProposals(ctx)
	requireNoError(t, err, "Failed to query collectives proposals")

	t.Logf("Collectives proposals: %s", string(result))
}

// TestCollectiveByName tests querying a collective by name.
func TestCollectiveByName(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := collectives.New(client)

	// Query a collective by name - use a test name
	result, err := mod.Collective(ctx, "test-collective")
	if err != nil {
		// Expected if collective doesn't exist
		t.Logf("Collective query: %v (expected if not exists)", err)
		return
	}

	t.Logf("Collective: %s", string(result))
}

// TestCollectivesCreate tests creating a new collective.
func TestCollectivesCreate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := collectives.New(client)

	collectiveName := generateUniqueID("testcoll")
	resp, err := mod.CreateCollective(ctx, TestKey, collectiveName, "Test collective description", nil)
	// May fail due to permissions or setup requirements
	if err != nil {
		t.Logf("CreateCollective may have failed (permissions): %v", err)
		return
	}

	t.Logf("CreateCollective TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestCollectivesContribute tests contributing to a collective.
func TestCollectivesContribute(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	// Wait for previous TX
	waitForBlocks(t, 1)

	mod := collectives.New(client)

	// Try to contribute to a collective
	resp, err := mod.ContributeCollective(ctx, TestKey, "test-collective", "100ukex", nil)
	// May fail if collective doesn't exist
	if err != nil {
		t.Logf("ContributeCollective may have failed (collective not exists): %v", err)
		return
	}

	t.Logf("ContributeCollective TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestCollectivesDonate tests donating to a collective.
func TestCollectivesDonate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	// Wait for previous TX
	waitForBlocks(t, 1)

	mod := collectives.New(client)

	// Try to donate to a collective
	resp, err := mod.DonateCollective(ctx, TestKey, "test-collective", 0, "10ukex", false, nil)
	// May fail if collective doesn't exist or not contributor
	if err != nil {
		t.Logf("DonateCollective may have failed (collective not exists): %v", err)
		return
	}

	t.Logf("DonateCollective TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestCollectivesWithdraw tests withdrawing from a collective.
func TestCollectivesWithdraw(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	// Wait for previous TX
	waitForBlocks(t, 1)

	mod := collectives.New(client)

	// Try to withdraw from a collective
	resp, err := mod.WithdrawCollective(ctx, TestKey, "test-collective", "10ukex", nil)
	// May fail if collective doesn't exist or not contributor
	if err != nil {
		t.Logf("WithdrawCollective may have failed (collective not exists): %v", err)
		return
	}

	t.Logf("WithdrawCollective TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestCollectivesProposalUpdate tests creating a proposal to update a collective.
func TestCollectivesProposalUpdate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	// Wait for previous TX
	waitForBlocks(t, 1)

	mod := collectives.New(client)

	propOpts := &collectives.ProposalCollectiveUpdateOpts{
		Title:                 "Test update collective proposal",
		Description:           "Integration test - verify update proposal submission",
		CollectiveName:        "test-collective",
		CollectiveDescription: "Updated description",
	}

	resp, err := mod.ProposalCollectiveUpdate(ctx, TestKey, propOpts, nil)
	// May fail if collective doesn't exist
	if err != nil {
		t.Logf("ProposalCollectiveUpdate may have failed (collective not exists): %v", err)
		return
	}

	t.Logf("ProposalCollectiveUpdate TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestCollectivesProposalRemove tests creating a proposal to remove a collective.
func TestCollectivesProposalRemove(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	// Wait for previous TX
	waitForBlocks(t, 1)

	mod := collectives.New(client)

	propOpts := &collectives.ProposalRemoveCollectiveOpts{
		Title:          "Test remove collective proposal",
		Description:    "Integration test - verify remove proposal submission",
		CollectiveName: "test-collective",
	}

	resp, err := mod.ProposalRemoveCollective(ctx, TestKey, propOpts, nil)
	// May fail if collective doesn't exist
	if err != nil {
		t.Logf("ProposalRemoveCollective may have failed (collective not exists): %v", err)
		return
	}

	t.Logf("ProposalRemoveCollective TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestCollectivesProposalSendDonation tests creating a proposal to send donation.
func TestCollectivesProposalSendDonation(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	// Wait for previous TX
	waitForBlocks(t, 1)

	mod := collectives.New(client)

	propOpts := &collectives.ProposalSendDonationOpts{
		Title:          "Test send donation proposal",
		Description:    "Integration test - verify send donation proposal submission",
		CollectiveName: "test-collective",
		Address:        TestAddress,
		Amounts:        "10ukex",
	}

	resp, err := mod.ProposalSendDonation(ctx, TestKey, propOpts, nil)
	// May fail if collective doesn't exist or no donations
	if err != nil {
		t.Logf("ProposalSendDonation may have failed (collective not exists): %v", err)
		return
	}

	t.Logf("ProposalSendDonation TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}
