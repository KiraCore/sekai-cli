// Package integration provides integration tests for the tokens module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/tokens"
)

// TestTokensAllRates tests querying all token rates.
func TestTokensAllRates(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := tokens.New(client)
	result, err := mod.AllRates(ctx)
	requireNoError(t, err, "Failed to query all rates")
	requireNotNil(t, result, "All rates is nil")

	t.Logf("Found %d token rates:", len(result.Data))
	for _, rate := range result.Data {
		t.Logf("  %s: rate=%s, fee_enabled=%v", rate.Data.Denom, rate.Data.FeeRate, rate.Data.FeeEnabled)
	}
}

// TestTokensRate tests querying a specific token rate.
func TestTokensRate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := tokens.New(client)
	result, err := mod.Rate(ctx, "ukex")
	requireNoError(t, err, "Failed to query rate")
	requireNotNil(t, result, "Rate is nil")

	requireEqual(t, "ukex", result.Denom, "Denom mismatch")
	t.Logf("Rate for ukex: FeeRate=%s, FeeEnabled=%v", result.FeeRate, result.FeeEnabled)
}

// TestTokensRatesByDenom tests querying token rates by denom.
func TestTokensRatesByDenom(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := tokens.New(client)
	result, err := mod.RatesByDenom(ctx, "ukex")
	requireNoError(t, err, "Failed to query rates by denom")

	t.Logf("Found %d denoms in rates", len(result))
	for denom, rateData := range result {
		t.Logf("  %s: rate=%s, supply=%s", denom, rateData.Data.FeeRate, rateData.Supply.Amount)
	}
}

// TestTokensBlackWhites tests querying token blacklist and whitelist.
func TestTokensBlackWhites(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := tokens.New(client)
	result, err := mod.TokenBlackWhites(ctx)
	requireNoError(t, err, "Failed to query token black whites")
	requireNotNil(t, result, "Token black whites is nil")

	t.Logf("Whitelisted tokens: %v", result.Whitelisted)
	t.Logf("Blacklisted tokens: %v", result.Blacklisted)
}

// TestTokensProposalUpsertRate tests creating a proposal to upsert a token rate.
// This is an atomic test sequence that:
// 1. Submits a proposal to upsert a new token rate
// 2. Votes YES on the proposal
// 3. Waits for the proposal to pass
// 4. Verifies the new rate exists
func TestTokensProposalUpsertRate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getExtendedTestContext()
	defer cancel()

	// Set fast proposal timing for testing
	setFastProposalTiming(t, client)

	mod := tokens.New(client)
	testDenom := generateUniqueID("utest")
	testSymbol := generateUniqueID("TST") // Use unique symbol to avoid conflicts

	// Submit proposal to create a new token rate
	// Note: StakeCap defaults to 0.1 (10%) which can exceed 100% total cap
	// Set StakeCap to "0" to avoid stake cap errors
	propOpts := &tokens.ProposalUpsertRateOpts{
		Denom:       testDenom,
		Decimals:    6,
		FeeRate:     "1.0",
		FeePayments: true,
		Name:        "Test Token " + testDenom,
		Symbol:      testSymbol,
		StakeCap:    "0",
		StakeMin:    "0",
		Title:       "Add test token rate",
		Description: "Integration test - add new token rate",
	}

	proposalID := submitAndPassProposal(t, client, func() (*sdk.TxResponse, error) {
		return mod.ProposalUpsertRate(ctx, TestKey, propOpts, nil)
	})

	t.Logf("Proposal %s passed, verifying new rate exists", proposalID)

	// Verify the new rate exists
	rate, err := mod.Rate(ctx, testDenom)
	if err != nil {
		// Rate might not exist yet, wait a bit and retry
		waitForBlocks(t, 2)
		rate, err = mod.Rate(ctx, testDenom)
	}
	requireNoError(t, err, "Failed to query new rate")
	requireEqual(t, testDenom, rate.Denom, "New rate denom mismatch")

	t.Logf("Successfully created token rate for %s", testDenom)
}

// TestTokensUpsertRate tests direct upsert of a token rate (requires sudo permission).
func TestTokensUpsertRate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := tokens.New(client)
	testDenom := generateUniqueID("udirect")
	testSymbol := generateUniqueID("DIR")

	// Submit direct upsert rate (requires sudo permission - genesis has it)
	resp, err := mod.UpsertRate(ctx, TestKey, &tokens.UpsertRateOpts{
		Denom:    testDenom,
		Name:     "Direct Test Token " + testDenom,
		Symbol:   testSymbol,
		FeeRate:  "1.0",
		Decimals: 6,
		StakeCap: "0",
		StakeMin: "0",
	}, nil)
	requireNoError(t, err, "Failed to upsert rate")
	requireNotNil(t, resp, "Response is nil")

	t.Logf("UpsertRate TX hash: %s, code: %d", resp.TxHash, resp.Code)

	// Wait for TX to be processed
	waitForBlocks(t, 2)

	// Verify the new rate exists
	rate, err := mod.Rate(ctx, testDenom)
	requireNoError(t, err, "Failed to query new rate")
	requireEqual(t, testDenom, rate.Denom, "New rate denom mismatch")

	t.Logf("Successfully created token rate %s via direct upsert", testDenom)
}

// TestTokensProposalUpdateBlackWhite tests creating a proposal to update token blacklist/whitelist.
func TestTokensProposalUpdateBlackWhite(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getExtendedTestContext()
	defer cancel()

	// Set fast proposal timing for testing
	setFastProposalTiming(t, client)

	mod := tokens.New(client)

	// First, query current black/whites to know the state
	currentBW, err := mod.TokenBlackWhites(ctx)
	requireNoError(t, err, "Failed to query current black/whites")
	t.Logf("Current whitelist: %v, blacklist: %v", currentBW.Whitelisted, currentBW.Blacklisted)

	// Submit proposal to add a token to whitelist
	testToken := generateUniqueID("whitelist_token")
	propOpts := &tokens.ProposalUpdateTokensBlackWhiteOpts{
		IsBlacklist: false, // whitelist
		IsAdd:       true,  // add (not remove)
		Tokens:      []string{testToken},
		Title:       "Add token to whitelist",
		Description: "Integration test - add token to whitelist",
	}

	proposalID := submitAndPassProposal(t, client, func() (*sdk.TxResponse, error) {
		return mod.ProposalUpdateTokensBlackWhite(ctx, TestKey, propOpts, nil)
	})

	t.Logf("Proposal %s passed, verifying whitelist updated", proposalID)

	// Verify the token was added to whitelist
	newBW, err := mod.TokenBlackWhites(ctx)
	requireNoError(t, err, "Failed to query updated black/whites")

	// Check if testToken is in whitelist
	found := false
	for _, token := range newBW.Whitelisted {
		if token == testToken {
			found = true
			break
		}
	}
	requireEqual(t, true, found, "Token not found in whitelist after proposal passed")

	t.Logf("Successfully added %s to whitelist", testToken)
}
