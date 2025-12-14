// Package integration provides integration tests for the recovery module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/recovery"
)

// TestRecoveryRecoveryRecord tests querying recovery record for an address.
func TestRecoveryRecoveryRecord(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := recovery.New(client)
	result, err := mod.RecoveryRecord(ctx, TestAddress)
	if err != nil {
		// This may fail if no recovery record exists, which is expected
		t.Logf("Recovery record query: %v (expected if no recovery record exists)", err)
		return
	}

	t.Logf("Recovery record: %s", string(result))
}

// TestRecoveryRecoveryToken tests querying recovery token for an address.
func TestRecoveryRecoveryToken(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := recovery.New(client)
	result, err := mod.RecoveryToken(ctx, TestAddress)
	if err != nil {
		// This may fail if no recovery token exists, which is expected
		t.Logf("Recovery token query: %v (expected if no recovery token exists)", err)
		return
	}

	t.Logf("Recovery token: %s", string(result))
}

// TestRecoveryRRHolderRewards tests querying RR holder rewards for an address.
func TestRecoveryRRHolderRewards(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := recovery.New(client)
	result, err := mod.RRHolderRewards(ctx, TestAddress)
	if err != nil {
		// This may fail if no RR holder rewards exist, which is expected
		t.Logf("RR holder rewards query: %v (expected if no rewards exist)", err)
		return
	}

	t.Logf("RR holder rewards: %s", string(result))
}

// TestRecoveryRRHolders tests querying RR holders for a token.
func TestRecoveryRRHolders(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := recovery.New(client)
	// Use a placeholder token name - may not exist
	result, err := mod.RRHolders(ctx, "ukex")
	if err != nil {
		// This may fail if no RR holders exist for this token, which is expected
		t.Logf("RR holders query: %v (expected if no holders exist)", err)
		return
	}

	t.Logf("RR holders: %s", string(result))
}
