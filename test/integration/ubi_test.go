// Package integration provides integration tests for the ubi module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/ubi"
)

// TestUBIRecords tests querying all UBI records.
func TestUBIRecords(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := ubi.New(client)
	result, err := mod.Records(ctx)
	requireNoError(t, err, "Failed to query UBI records")
	requireNotNil(t, result, "UBI records is nil")

	t.Logf("Found %d UBI records:", len(result.Records))
	for _, rec := range result.Records {
		t.Logf("  %s: amount=%s, period=%s, pool=%s", rec.Name, rec.Amount, rec.Period, rec.Pool)
	}
}

// TestUBIRecordByName tests querying a UBI record by name.
func TestUBIRecordByName(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := ubi.New(client)

	// First get all records
	records, err := mod.Records(ctx)
	requireNoError(t, err, "Failed to query UBI records")

	if len(records.Records) == 0 {
		t.Log("No UBI records found, skipping")
		return
	}

	// Query the first record
	recordName := records.Records[0].Name
	result, err := mod.RecordByName(ctx, recordName)
	requireNoError(t, err, "Failed to query UBI record by name")
	requireNotNil(t, result, "UBI record is nil")

	t.Logf("UBI record %s: amount=%s, period=%s", result.Name, result.Amount, result.Period)
}

// TestUBIProposalUpsert tests creating a proposal to upsert a UBI record.
// Note: Due to strict hardcap validation in the UBI module, this test only verifies
// that the SDK correctly constructs and submits the proposal.
func TestUBIProposalUpsert(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := ubi.New(client)

	// Get existing UBI record for reference
	records, err := mod.Records(ctx)
	requireNoError(t, err, "Failed to query existing UBI records")
	t.Logf("Found %d existing UBI records", len(records.Records))

	// Test that ProposalUpsertUBI can be called with correct parameters
	// The proposal may fail on-chain due to hardcap constraints, but SDK call should work
	testName := generateUniqueID("testubi")
	propOpts := &ubi.ProposalUpsertUBIOpts{
		Name:        testName,
		DistrStart:  0,
		DistrEnd:    0,
		Amount:      100,
		Period:      2592000,
		PoolName:    "ValidatorBasicRewardsPool",
		Title:       "Test UBI record",
		Description: "Integration test - verify upsert proposal submission",
	}

	resp, err := mod.ProposalUpsertUBI(ctx, TestKey, propOpts, nil)
	requireNoError(t, err, "Failed to submit upsert proposal")
	requireNotNil(t, resp, "Response should not be nil")

	t.Logf("ProposalUpsertUBI TX submitted: hash=%s, code=%d", resp.TxHash, resp.Code)

	// The TX submission should succeed (code may be non-zero due to hardcap)
	// We verify the SDK correctly built and submitted the transaction
	requireTrue(t, resp.TxHash != "", "TX hash should not be empty")
}

// TestUBIProposalRemove tests creating a proposal to remove a UBI record.
// Note: This test only verifies the SDK can submit the proposal correctly.
// It cannot create a new UBI record due to hardcap constraints.
func TestUBIProposalRemove(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	// Wait for previous TX to be confirmed
	waitForBlocks(t, 1)

	mod := ubi.New(client)

	// Test that ProposalRemoveUBI can be called with correct parameters
	// We use a non-existent name to verify the SDK call works
	// (The proposal will fail on-chain because the UBI doesn't exist, which is expected)
	testName := generateUniqueID("nonexistent_ubi")

	removeOpts := &ubi.ProposalRemoveUBIOpts{
		Name:        testName,
		Title:       "Remove non-existent UBI",
		Description: "Integration test - verify remove proposal submission",
	}

	resp, err := mod.ProposalRemoveUBI(ctx, TestKey, removeOpts, nil)
	// The call should succeed (TX submitted) even if proposal fails
	requireNoError(t, err, "Failed to submit remove proposal")
	requireNotNil(t, resp, "Response should not be nil")

	t.Logf("ProposalRemoveUBI TX submitted: hash=%s, code=%d", resp.TxHash, resp.Code)

	// Note: We don't wait for the proposal to pass because:
	// 1. Creating a new UBI would hit hardcap
	// 2. Removing existing UBI would break other tests
	// The test verifies the SDK correctly submits the proposal
}
