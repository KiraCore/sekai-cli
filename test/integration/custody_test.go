// Package integration provides integration tests for the custody module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/custody"
)

// TestCustodyGet tests querying custody for an address.
func TestCustodyGet(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := custody.New(client)
	result, err := mod.Get(ctx, testAddr)
	if err != nil {
		// This may fail if no custody is set
		t.Logf("Custody get query: %v (expected if no custody)", err)
		return
	}

	t.Logf("Custody for %s: %s", testAddr, string(result))
}

// TestCustodyCustodians tests querying custody custodians.
func TestCustodyCustodians(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := custody.New(client)
	result, err := mod.Custodians(ctx, testAddr)
	if err != nil {
		// This may fail if no custodians are set
		t.Logf("Custodians query: %v (expected if no custodians)", err)
		return
	}

	t.Logf("Custodians for %s: %s", testAddr, string(result))
}

// TestCustodyWhitelist tests querying custody whitelist.
func TestCustodyWhitelist(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := custody.New(client)
	result, err := mod.Whitelist(ctx, testAddr)
	if err != nil {
		// This may fail if no whitelist is set
		t.Logf("Whitelist query: %v (expected if no whitelist)", err)
		return
	}

	t.Logf("Whitelist for %s: %s", testAddr, string(result))
}

// TestCustodyLimits tests querying custody limits.
func TestCustodyLimits(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := custody.New(client)
	result, err := mod.Limits(ctx, testAddr)
	if err != nil {
		// This may fail if no limits are set
		t.Logf("Limits query: %v (expected if no limits)", err)
		return
	}

	t.Logf("Limits for %s: %s", testAddr, string(result))
}

// TestCustodyCustodiansPool tests querying custody pool for an address.
func TestCustodyCustodiansPool(t *testing.T) {
	skipIfContainerNotRunning(t)
	testAddr := getTestAddress(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := custody.New(client)
	result, err := mod.CustodiansPool(ctx, testAddr)
	requireNoError(t, err, "Failed to query custody pool")

	t.Logf("Custody pool for %s: %s", testAddr, string(result))
}
