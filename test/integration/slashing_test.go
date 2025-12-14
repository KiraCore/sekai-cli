// Package integration provides integration tests for the slashing module.
package integration

import (
	"testing"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/slashing"
)

// TestSlashingSigningInfos tests querying all signing infos.
func TestSlashingSigningInfos(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := slashing.New(client)
	result, err := mod.SigningInfos(ctx)
	requireNoError(t, err, "Failed to query signing infos")
	requireNotNil(t, result, "Signing infos is nil")

	// There should be at least one validator signing info
	requireTrue(t, len(result.Info) > 0, "Should have at least one signing info")

	t.Logf("Found %d signing infos", len(result.Info))
	for _, info := range result.Info {
		t.Logf("  Validator: %s, MissedBlocks: %s, ProducedBlocks: %s",
			info.Address[:20]+"...", info.MissedBlocksCounter, info.ProducedBlocksCounter)
	}
}

// TestSlashingSigningInfo tests querying a specific validator's signing info.
func TestSlashingSigningInfo(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := slashing.New(client)

	// First get all signing infos to get a valid validator address
	allInfos, err := mod.SigningInfos(ctx)
	requireNoError(t, err, "Failed to query signing infos")
	requireTrue(t, len(allInfos.Info) > 0, "No signing infos found")

	// Query specific validator
	validatorConsAddr := allInfos.Info[0].Address
	result, err := mod.SigningInfo(ctx, validatorConsAddr)
	requireNoError(t, err, "Failed to query signing info")
	requireNotNil(t, result, "Signing info is nil")

	requireEqual(t, validatorConsAddr, result.ValSigningInfo.Address, "Address mismatch")
	t.Logf("Validator %s: Missed=%s, Produced=%s, Mischance=%s",
		result.ValSigningInfo.Address[:20]+"...",
		result.ValSigningInfo.MissedBlocksCounter,
		result.ValSigningInfo.ProducedBlocksCounter,
		result.ValSigningInfo.Mischance)
}

// TestSlashingActiveStakingPools tests querying active staking pools.
func TestSlashingActiveStakingPools(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := slashing.New(client)
	result, err := mod.ActiveStakingPools(ctx)
	requireNoError(t, err, "Failed to query active staking pools")

	// Result can be empty, just verify no error
	t.Logf("Active staking pools response: %s", string(result))
}

// TestSlashingInactiveStakingPools tests querying inactive staking pools.
func TestSlashingInactiveStakingPools(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := slashing.New(client)
	result, err := mod.InactiveStakingPools(ctx)
	requireNoError(t, err, "Failed to query inactive staking pools")

	// Result can be empty, just verify no error
	t.Logf("Inactive staking pools response: %s", string(result))
}

// TestSlashingSlashedStakingPools tests querying slashed staking pools.
func TestSlashingSlashedStakingPools(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := slashing.New(client)
	result, err := mod.SlashedStakingPools(ctx)
	requireNoError(t, err, "Failed to query slashed staking pools")

	// Result can be empty (no slashed pools), just verify no error
	t.Logf("Slashed staking pools response: %s", string(result))
}

// TestSlashingSlashProposals tests querying slash proposals.
func TestSlashingSlashProposals(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := slashing.New(client)
	result, err := mod.SlashProposals(ctx)
	requireNoError(t, err, "Failed to query slash proposals")

	// Result can be empty (no slash proposals), just verify no error
	t.Logf("Slash proposals response: %s", string(result))
}
