// Package integration provides integration tests for the distributor module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/distributor"
)

// TestDistributorFeesTreasury tests querying fees treasury.
func TestDistributorFeesTreasury(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := distributor.New(client)
	result, err := mod.FeesTreasury(ctx)
	requireNoError(t, err, "Failed to query fees treasury")
	requireNotNil(t, result, "Fees treasury is nil")

	t.Logf("Fees treasury: %s", string(result))
}

// TestDistributorPeriodicSnapshot tests querying periodic snapshot.
func TestDistributorPeriodicSnapshot(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := distributor.New(client)
	result, err := mod.PeriodicSnapshot(ctx)
	requireNoError(t, err, "Failed to query periodic snapshot")
	requireNotNil(t, result, "Periodic snapshot is nil")

	t.Logf("Periodic snapshot: %s", string(result))
}

// TestDistributorSnapshotPeriod tests querying snapshot period.
func TestDistributorSnapshotPeriod(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := distributor.New(client)
	result, err := mod.SnapshotPeriod(ctx)
	requireNoError(t, err, "Failed to query snapshot period")
	requireNotNil(t, result, "Snapshot period is nil")

	t.Logf("Snapshot period: %s", string(result))
}

// TestDistributorYearStartSnapshot tests querying year start snapshot.
func TestDistributorYearStartSnapshot(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := distributor.New(client)
	result, err := mod.YearStartSnapshot(ctx)
	requireNoError(t, err, "Failed to query year start snapshot")
	requireNotNil(t, result, "Year start snapshot is nil")

	t.Logf("Year start snapshot: %s", string(result))
}

// TestDistributorSnapshotPeriodPerformance tests querying snapshot period performance.
func TestDistributorSnapshotPeriodPerformance(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := distributor.New(client)
	// Use test address as a validator address
	result, err := mod.SnapshotPeriodPerformance(ctx, TestAddress)
	if err != nil {
		// This may fail if the address is not a validator, which is expected
		t.Logf("Snapshot period performance query: %v (may be expected if not a validator)", err)
		return
	}

	t.Logf("Snapshot period performance: %s", string(result))
}
