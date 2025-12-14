// Package integration provides integration tests for the auth module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/auth"
)

// TestAuthAccount tests querying a single account by address.
func TestAuthAccount(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := auth.New(client)
	result, err := mod.Account(ctx, TestAddress)
	requireNoError(t, err, "Failed to query account")
	requireNotNil(t, result, "Account is nil")

	// Verify account has expected fields
	requireEqual(t, TestAddress, result.Address, "Address mismatch")
	requireTrue(t, result.AccountNumber != "", "Account number should not be empty")

	t.Logf("Account: %s, Number: %s, Sequence: %s", result.Address, result.AccountNumber, result.Sequence)
}

// TestAuthAccounts tests querying all accounts.
func TestAuthAccounts(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := auth.New(client)
	result, err := mod.Accounts(ctx, nil)
	requireNoError(t, err, "Failed to query accounts")
	requireNotNil(t, result, "Accounts response is nil")

	// There should be at least one account (genesis)
	requireTrue(t, len(result.Accounts) > 0, "Should have at least one account")

	t.Logf("Found %d accounts", len(result.Accounts))

	// Verify the test address is in the list
	found := false
	for _, acc := range result.Accounts {
		if acc.Address == TestAddress {
			found = true
			break
		}
	}
	requireTrue(t, found, "Test address should be in accounts list")
}

// TestAuthAddressByAccNum tests querying an address by account number.
func TestAuthAddressByAccNum(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := auth.New(client)

	// First get the account to know its account number
	acc, err := mod.Account(ctx, TestAddress)
	requireNoError(t, err, "Failed to query account")
	requireNotNil(t, acc, "Account is nil")

	// Now query by account number
	result, err := mod.AddressByAccNum(ctx, acc.AccountNumber)
	requireNoError(t, err, "Failed to query address by account number")
	requireNotNil(t, result, "Result is nil")

	requireEqual(t, TestAddress, result.AccountAddress, "Address mismatch")
	t.Logf("Account number %s maps to address %s", acc.AccountNumber, result.AccountAddress)
}

// TestAuthModuleAccount tests querying a module account by name.
func TestAuthModuleAccount(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := auth.New(client)

	// Query a known module account (fee_collector is standard in cosmos-sdk)
	result, err := mod.ModuleAccount(ctx, "fee_collector")
	requireNoError(t, err, "Failed to query module account")
	requireNotNil(t, result, "Module account is nil")

	requireEqual(t, "fee_collector", result.Name, "Module account name mismatch")
	t.Logf("Module account: %s", result.Name)
}

// TestAuthModuleAccounts tests querying all module accounts.
func TestAuthModuleAccounts(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := auth.New(client)
	result, err := mod.ModuleAccounts(ctx)
	requireNoError(t, err, "Failed to query module accounts")
	requireNotNil(t, result, "Module accounts is nil")

	// There should be multiple module accounts
	requireTrue(t, len(result) > 0, "Should have at least one module account")

	t.Logf("Found %d module accounts:", len(result))
	for _, acc := range result {
		t.Logf("  - %s", acc.Name)
	}

	// Verify fee_collector is in the list (standard module account)
	found := false
	for _, acc := range result {
		if acc.Name == "fee_collector" {
			found = true
			break
		}
	}
	requireTrue(t, found, "fee_collector should be in module accounts list")
}

// TestAuthParams tests querying auth module parameters.
func TestAuthParams(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := auth.New(client)
	result, err := mod.Params(ctx)
	requireNoError(t, err, "Failed to query auth params")
	requireNotNil(t, result, "Auth params is nil")

	// Verify params have expected fields
	requireTrue(t, result.MaxMemoCharacters != "", "MaxMemoCharacters should not be empty")
	requireTrue(t, result.TxSigLimit != "", "TxSigLimit should not be empty")
	requireTrue(t, result.TxSizeCostPerByte != "", "TxSizeCostPerByte should not be empty")

	t.Logf("Auth Params: MaxMemo=%s, TxSigLimit=%s, TxSizeCost=%s",
		result.MaxMemoCharacters, result.TxSigLimit, result.TxSizeCostPerByte)
}

// TestAuthAccountsPagination tests querying accounts with pagination.
func TestAuthAccountsPagination(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := auth.New(client)

	// Query with a limit
	result, err := mod.Accounts(ctx, &sdk.Pagination{Limit: 2})
	requireNoError(t, err, "Failed to query accounts with pagination")
	requireNotNil(t, result, "Accounts response is nil")

	// Should have at most 2 accounts
	requireTrue(t, len(result.Accounts) <= 2, "Should have at most 2 accounts with limit")

	t.Logf("Paginated query returned %d accounts", len(result.Accounts))
}
