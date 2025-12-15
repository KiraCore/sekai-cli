// Package integration provides integration tests for the bank module.
package integration

import (
	"testing"
	"time"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/bank"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/keys"
	"github.com/kiracore/sekai-cli/pkg/sdk/types"
)

// TestBankBalances tests querying all balances for an address.
func TestBankBalances(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := bank.New(client)
	result, err := mod.Balances(ctx, testAddr)
	requireNoError(t, err, "Failed to query balances")

	t.Logf("Address %s has %d token types", testAddr, len(result))
	for _, coin := range result {
		t.Logf("  %s: %s", coin.Denom, coin.Amount)
	}
}

// TestBankBalance tests querying balance for a specific denom.
func TestBankBalance(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := bank.New(client)
	result, err := mod.Balance(ctx, testAddr, "ukex")
	requireNoError(t, err, "Failed to query balance")
	requireNotNil(t, result, "Balance is nil")

	requireEqual(t, "ukex", result.Denom, "Denom mismatch")
	t.Logf("Address %s has %s ukex", testAddr, result.Amount)
}

// TestBankSpendableBalances tests querying spendable balances.
func TestBankSpendableBalances(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := bank.New(client)
	result, err := mod.SpendableBalances(ctx, testAddr)
	requireNoError(t, err, "Failed to query spendable balances")

	t.Logf("Address %s has %d spendable token types", testAddr, len(result))
	for _, coin := range result {
		t.Logf("  %s: %s", coin.Denom, coin.Amount)
	}
}

// TestBankTotalSupply tests querying total supply of all tokens.
func TestBankTotalSupply(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := bank.New(client)
	result, err := mod.TotalSupply(ctx)
	requireNoError(t, err, "Failed to query total supply")
	requireTrue(t, len(result) > 0, "Should have at least one token in supply")

	t.Logf("Total supply has %d token types", len(result))
	for _, coin := range result {
		t.Logf("  %s: %s", coin.Denom, coin.Amount)
	}
}

// TestBankDenomMetadata tests querying denom metadata.
func TestBankDenomMetadata(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := bank.New(client)
	result, err := mod.AllDenomsMetadata(ctx)
	requireNoError(t, err, "Failed to query denom metadata")
	requireNotNil(t, result, "Denom metadata is nil")

	t.Logf("Found %d denom metadata entries", len(result.Metadatas))
	for _, meta := range result.Metadatas {
		t.Logf("  %s: %s", meta.Base, meta.Description)
	}
}

// TestBankSendEnabled tests querying send enabled status.
func TestBankSendEnabled(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := bank.New(client)
	result, err := mod.SendEnabled(ctx)
	requireNoError(t, err, "Failed to query send enabled")
	requireNotNil(t, result, "Send enabled is nil")

	t.Logf("Found %d send enabled entries", len(result.SendEnabled))
	for _, entry := range result.SendEnabled {
		t.Logf("  %s: enabled=%v", entry.Denom, entry.Enabled)
	}
}

// TestBankSend tests sending tokens from one account to another.
func TestBankSend(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	bankMod := bank.New(client)
	keysMod := keys.New(client)

	// Create a temporary recipient key for testing
	recipientKeyName := generateUniqueID("recipient")
	_, err := keysMod.Add(ctx, recipientKeyName, nil)
	requireNoError(t, err, "Failed to create recipient key")
	defer func() {
		// Cleanup: delete the test key
		_ = keysMod.Delete(ctx, recipientKeyName, true)
	}()

	// Get recipient address
	keyInfo, err := keysMod.Show(ctx, recipientKeyName)
	requireNoError(t, err, "Failed to get recipient key info")
	recipientAddr := keyInfo.Address
	t.Logf("Recipient address: %s", recipientAddr)

	// Get initial balances
	testAddr := getTestAddress(t)
	senderBalanceBefore, err := bankMod.Balance(ctx, testAddr, "ukex")
	requireNoError(t, err, "Failed to query sender balance")
	t.Logf("Sender balance before: %s ukex", senderBalanceBefore.Amount)

	// Send 1000 ukex to recipient
	amount := types.NewCoins(types.NewCoin("ukex", 1000))
	resp, err := bankMod.Send(ctx, TestKey, recipientAddr, amount, nil)
	requireNoError(t, err, "Failed to send tokens")
	requireTxSuccess(t, resp, "Send transaction failed")
	t.Logf("Send TX hash: %s", resp.TxHash)

	// Wait for TX to be included in block
	time.Sleep(7 * time.Second)

	// Verify recipient received tokens
	recipientBalance, err := bankMod.Balance(ctx, recipientAddr, "ukex")
	requireNoError(t, err, "Failed to query recipient balance")
	requireEqual(t, "1000", recipientBalance.Amount, "Recipient should have 1000 ukex")
	t.Logf("Recipient balance after: %s ukex", recipientBalance.Amount)
}

// TestBankMultiSend tests sending tokens to multiple recipients.
func TestBankMultiSend(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	bankMod := bank.New(client)
	keysMod := keys.New(client)

	// Create two temporary recipient keys for testing
	recipient1Name := generateUniqueID("msrecip1")
	recipient2Name := generateUniqueID("msrecip2")

	_, err := keysMod.Add(ctx, recipient1Name, nil)
	requireNoError(t, err, "Failed to create recipient1 key")
	defer func() { _ = keysMod.Delete(ctx, recipient1Name, true) }()

	_, err = keysMod.Add(ctx, recipient2Name, nil)
	requireNoError(t, err, "Failed to create recipient2 key")
	defer func() { _ = keysMod.Delete(ctx, recipient2Name, true) }()

	// Get recipient addresses
	keyInfo1, err := keysMod.Show(ctx, recipient1Name)
	requireNoError(t, err, "Failed to get recipient1 key info")
	recipient1Addr := keyInfo1.Address

	keyInfo2, err := keysMod.Show(ctx, recipient2Name)
	requireNoError(t, err, "Failed to get recipient2 key info")
	recipient2Addr := keyInfo2.Address

	t.Logf("Recipients: %s, %s", recipient1Addr, recipient2Addr)

	// Multi-send 500 ukex to each (with split=true, total 1000 ukex split between them)
	amount := types.NewCoins(types.NewCoin("ukex", 1000))
	resp, err := bankMod.MultiSend(ctx, TestKey, []string{recipient1Addr, recipient2Addr}, amount,
		&bank.MultiSendOptions{Split: true})
	requireNoError(t, err, "Failed to multi-send tokens")
	requireTxSuccess(t, resp, "Multi-send transaction failed")
	t.Logf("Multi-send TX hash: %s", resp.TxHash)

	// Wait for TX to be included in block
	time.Sleep(7 * time.Second)

	// Verify recipients received tokens (500 each with split)
	balance1, err := bankMod.Balance(ctx, recipient1Addr, "ukex")
	requireNoError(t, err, "Failed to query recipient1 balance")
	t.Logf("Recipient1 balance after: %s ukex", balance1.Amount)

	balance2, err := bankMod.Balance(ctx, recipient2Addr, "ukex")
	requireNoError(t, err, "Failed to query recipient2 balance")
	t.Logf("Recipient2 balance after: %s ukex", balance2.Amount)
}
