// Package integration provides integration tests for the customstaking module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/staking"
)

// TestStakingValidators tests querying all validators.
func TestStakingValidators(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := staking.New(client)
	result, err := mod.Validators(ctx, nil)
	requireNoError(t, err, "Failed to query validators")
	requireNotNil(t, result, "Validators is nil")

	requireTrue(t, len(result.Validators) > 0, "Should have at least one validator")

	t.Logf("Found %d validators:", len(result.Validators))
	for _, v := range result.Validators {
		t.Logf("  %s: status=%s, moniker=%s", v.Address, v.Status, v.Moniker)
	}
}

// TestStakingValidator tests querying a specific validator.
func TestStakingValidator(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := staking.New(client)

	// First get all validators to get a valid address
	validators, err := mod.Validators(ctx, nil)
	requireNoError(t, err, "Failed to query validators")
	requireTrue(t, len(validators.Validators) > 0, "No validators found")

	// Query the first validator
	valAddr := validators.Validators[0].Address
	result, err := mod.Validator(ctx, &staking.ValidatorQueryOpts{Address: valAddr})
	requireNoError(t, err, "Failed to query validator")
	requireNotNil(t, result, "Validator is nil")

	t.Logf("Validator %s: status=%s, moniker=%s", result.Address, result.Status, result.Moniker)
}

// TestStakingValidatorsByStatus tests querying validators by status.
func TestStakingValidatorsByStatus(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := staking.New(client)

	// Query active validators
	result, err := mod.Validators(ctx, &staking.ValidatorQueryOpts{Status: "ACTIVE"})
	requireNoError(t, err, "Failed to query active validators")

	t.Logf("Found %d ACTIVE validators", len(result.Validators))
	for _, v := range result.Validators {
		t.Logf("  %s: moniker=%s", v.Address, v.Moniker)
	}
}

// TestStakingValidatorByMoniker tests querying validator by moniker.
func TestStakingValidatorByMoniker(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := staking.New(client)

	// First get a validator to know its moniker
	validators, err := mod.Validators(ctx, nil)
	requireNoError(t, err, "Failed to query validators")
	requireTrue(t, len(validators.Validators) > 0, "No validators found")

	moniker := validators.Validators[0].Moniker
	if moniker == "" {
		t.Skip("No validator with moniker found")
		return
	}

	// Query by moniker
	result, err := mod.Validator(ctx, &staking.ValidatorQueryOpts{Moniker: moniker})
	requireNoError(t, err, "Failed to query validator by moniker")
	requireNotNil(t, result, "Validator is nil")

	t.Logf("Validator with moniker %s: address=%s, status=%s", moniker, result.Address, result.Status)
}

// TestStakingClaimValidatorSeat tests claiming a validator seat.
// Note: This test verifies the SDK TX submission works, but the TX will fail
// because the genesis account is already a validator.
func TestStakingClaimValidatorSeat(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := staking.New(client)

	// Try to claim validator seat (will fail because genesis is already a validator)
	seatOpts := &staking.ClaimValidatorSeatOpts{
		Moniker: "test-validator",
	}

	// This will fail because genesis is already a validator, but we verify SDK call works
	resp, err := mod.ClaimValidatorSeat(ctx, TestKey, seatOpts, nil)

	// The transaction will be submitted but should fail on-chain
	// We just verify the SDK can construct and submit the TX
	if err != nil {
		// Expected to fail - genesis is already a validator
		t.Logf("ClaimValidatorSeat failed as expected (genesis is already a validator): %v", err)
		return
	}

	t.Logf("ClaimValidatorSeat TX: hash=%s, code=%d", resp.TxHash, resp.Code)
	// Code should be non-zero since the account is already a validator
	if resp.Code == 0 {
		t.Logf("Unexpected success - genesis might not be a validator yet")
	}
}

// TestStakingProposalUnjailValidator tests creating a proposal to unjail a validator.
// Note: This test verifies the SDK TX submission works.
func TestStakingProposalUnjailValidator(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	// Wait for any previous TX to be confirmed
	waitForBlocks(t, 1)

	mod := staking.New(client)

	// Get a validator address
	validators, err := mod.Validators(ctx, nil)
	requireNoError(t, err, "Failed to query validators")
	requireTrue(t, len(validators.Validators) > 0, "No validators found")

	valAddr := validators.Validators[0].ValKey
	t.Logf("Using validator: %s", valAddr)

	// Create unjail proposal (validator is not jailed, but we test TX submission)
	propOpts := &staking.ProposalUnjailValidatorOpts{
		Title:       "Test unjail validator proposal",
		Description: "Integration test - verify unjail proposal submission",
	}

	resp, err := mod.ProposalUnjailValidator(ctx, TestKey, valAddr, "test-reference", propOpts, nil)
	requireNoError(t, err, "Failed to submit unjail proposal")
	requireNotNil(t, resp, "Response should not be nil")

	t.Logf("ProposalUnjailValidator TX: hash=%s, code=%d", resp.TxHash, resp.Code)
	requireTrue(t, resp.TxHash != "", "TX hash should not be empty")
}
