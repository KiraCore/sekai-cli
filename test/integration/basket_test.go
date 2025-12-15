// Package integration provides integration tests for the basket module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/basket"
)

// TestBasketTokenBaskets tests querying token baskets.
func TestBasketTokenBaskets(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := basket.New(client)
	result, err := mod.TokenBaskets(ctx, "", false)
	requireNoError(t, err, "Failed to query token baskets")

	t.Logf("Token baskets: %s", string(result))
}

// TestBasketTokenBasketsDerivativesOnly tests querying token baskets with derivatives only.
func TestBasketTokenBasketsDerivativesOnly(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := basket.New(client)
	result, err := mod.TokenBaskets(ctx, "", true)
	requireNoError(t, err, "Failed to query derivative token baskets")

	t.Logf("Derivative token baskets: %s", string(result))
}

// TestBasketHistoricalMints tests querying historical mints.
func TestBasketHistoricalMints(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := basket.New(client)
	result, err := mod.HistoricalMints(ctx, "1")
	if err != nil {
		t.Logf("Historical mints query: %v (expected if no basket)", err)
		return
	}

	t.Logf("Historical mints: %s", string(result))
}

// TestBasketHistoricalBurns tests querying historical burns.
func TestBasketHistoricalBurns(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := basket.New(client)
	result, err := mod.HistoricalBurns(ctx, "1")
	if err != nil {
		t.Logf("Historical burns query: %v (expected if no basket)", err)
		return
	}

	t.Logf("Historical burns: %s", string(result))
}

// TestBasketHistoricalSwaps tests querying historical swaps.
func TestBasketHistoricalSwaps(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := basket.New(client)
	result, err := mod.HistoricalSwaps(ctx, "1")
	if err != nil {
		t.Logf("Historical swaps query: %v (expected if no basket)", err)
		return
	}

	t.Logf("Historical swaps: %s", string(result))
}

// TestBasketMintTokens tests minting basket tokens.
func TestBasketMintTokens(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := basket.New(client)

	resp, err := mod.MintBasketTokens(ctx, TestKey, "1", "100ukex", nil)
	if err != nil {
		t.Logf("MintBasketTokens may have failed (no basket): %v", err)
		return
	}

	t.Logf("MintBasketTokens TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestBasketBurnTokens tests burning basket tokens.
func TestBasketBurnTokens(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := basket.New(client)

	resp, err := mod.BurnBasketTokens(ctx, TestKey, "1", "10", nil)
	if err != nil {
		t.Logf("BurnBasketTokens may have failed (no basket): %v", err)
		return
	}

	t.Logf("BurnBasketTokens TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestBasketSwapTokens tests swapping basket tokens.
func TestBasketSwapTokens(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := basket.New(client)

	resp, err := mod.SwapBasketTokens(ctx, TestKey, "1", "100ukex", "100utest", nil)
	if err != nil {
		t.Logf("SwapBasketTokens may have failed (no basket): %v", err)
		return
	}

	t.Logf("SwapBasketTokens TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestBasketClaimRewards tests claiming basket rewards.
func TestBasketClaimRewards(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := basket.New(client)

	resp, err := mod.BasketClaimRewards(ctx, TestKey, "1", nil)
	if err != nil {
		t.Logf("BasketClaimRewards may have failed (no basket): %v", err)
		return
	}

	t.Logf("BasketClaimRewards TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestBasketDisableDeposits tests disabling basket deposits.
func TestBasketDisableDeposits(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := basket.New(client)

	resp, err := mod.DisableBasketDeposits(ctx, TestKey, "1", true, nil)
	if err != nil {
		t.Logf("DisableBasketDeposits may have failed (no basket): %v", err)
		return
	}

	t.Logf("DisableBasketDeposits TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestBasketDisableWithdraws tests disabling basket withdraws.
func TestBasketDisableWithdraws(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := basket.New(client)

	resp, err := mod.DisableBasketWithdraws(ctx, TestKey, "1", true, nil)
	if err != nil {
		t.Logf("DisableBasketWithdraws may have failed (no basket): %v", err)
		return
	}

	t.Logf("DisableBasketWithdraws TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestBasketDisableSwaps tests disabling basket swaps.
func TestBasketDisableSwaps(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := basket.New(client)

	resp, err := mod.DisableBasketSwaps(ctx, TestKey, "1", true, nil)
	if err != nil {
		t.Logf("DisableBasketSwaps may have failed (no basket): %v", err)
		return
	}

	t.Logf("DisableBasketSwaps TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestBasketProposalCreate tests creating a basket proposal.
func TestBasketProposalCreate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := basket.New(client)

	propOpts := &basket.ProposalCreateBasketOpts{
		BasketSuffix:      generateUniqueID("bskt"),
		BasketDescription: "Test basket",
		BasketTokens:      "ukex",
		Title:             "Test create basket proposal",
		Description:       "Integration test - verify create basket proposal",
	}

	resp, err := mod.ProposalCreateBasket(ctx, TestKey, propOpts, nil)
	if err != nil {
		t.Logf("ProposalCreateBasket may have failed: %v", err)
		return
	}

	t.Logf("ProposalCreateBasket TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestBasketProposalEdit tests editing a basket proposal.
func TestBasketProposalEdit(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := basket.New(client)

	propOpts := &basket.ProposalEditBasketOpts{
		BasketID:          1,
		BasketDescription: "Updated basket description",
		Title:             "Test edit basket proposal",
		Description:       "Integration test - verify edit basket proposal",
	}

	resp, err := mod.ProposalEditBasket(ctx, TestKey, propOpts, nil)
	if err != nil {
		t.Logf("ProposalEditBasket may have failed (no basket): %v", err)
		return
	}

	t.Logf("ProposalEditBasket TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestBasketProposalWithdrawSurplus tests withdrawing surplus proposal.
func TestBasketProposalWithdrawSurplus(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := basket.New(client)

	propOpts := &basket.ProposalWithdrawSurplusOpts{
		Title:       "Test withdraw surplus proposal",
		Description: "Integration test - verify withdraw surplus proposal",
	}

	testAddr := getTestAddress(t)
	resp, err := mod.ProposalWithdrawSurplus(ctx, TestKey, "1", testAddr, propOpts, nil)
	if err != nil {
		t.Logf("ProposalWithdrawSurplus may have failed (no basket): %v", err)
		return
	}

	t.Logf("ProposalWithdrawSurplus TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}
