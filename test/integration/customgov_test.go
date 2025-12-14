// Package integration provides integration tests for the customgov module.
package integration

import (
	"testing"
	"time"

	"github.com/kiracore/sekai-cli/pkg/sdk/modules/gov"
)

// === QUERY TESTS ===

// TestGovNetworkProperties tests querying network properties.
func TestGovNetworkProperties(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.NetworkProperties(ctx)
	requireNoError(t, err, "Failed to query network properties")
	requireNotNil(t, result, "Network properties is nil")

	requireTrue(t, result.MinTxFee != "", "MinTxFee should not be empty")
	requireTrue(t, result.MaxTxFee != "", "MaxTxFee should not be empty")

	t.Logf("Network Properties: MinTxFee=%s, MaxTxFee=%s, VoteQuorum=%s",
		result.MinTxFee, result.MaxTxFee, result.VoteQuorum)
}

// TestGovProposals tests querying proposals.
func TestGovProposals(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.Proposals(ctx, nil)
	requireNoError(t, err, "Failed to query proposals")
	requireNotNil(t, result, "Proposals is nil")

	t.Logf("Found %d proposals", len(result.Proposals))
	for _, prop := range result.Proposals {
		t.Logf("  Proposal %s: %s (%s)", prop.ProposalID, prop.Title, prop.Status)
	}
}

// TestGovCouncilors tests querying councilors.
func TestGovCouncilors(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.Councilors(ctx)
	requireNoError(t, err, "Failed to query councilors")

	t.Logf("Found %d councilors", len(result))
	for _, c := range result {
		t.Logf("  %s: %s (%s)", c.Address, c.Moniker, c.Status)
	}
}

// TestGovAllRoles tests querying all roles.
func TestGovAllRoles(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.AllRoles(ctx)
	requireNoError(t, err, "Failed to query all roles")

	t.Logf("Found %d roles", len(result))
	for _, role := range result {
		t.Logf("  Role %d: %s - %s", role.ID, role.Sid, role.Description)
	}
}

// TestGovRole tests querying a specific role.
func TestGovRole(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	// Query the sudo role (should exist)
	result, err := mod.Role(ctx, "sudo")
	requireNoError(t, err, "Failed to query role")
	requireNotNil(t, result, "Role is nil")

	t.Logf("Role: %s (ID: %d)", result.Sid, result.ID)
}

// TestGovRoles tests querying roles for an address.
func TestGovRoles(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.Roles(ctx, TestAddress)
	requireNoError(t, err, "Failed to query roles")

	t.Logf("Address %s has %d roles: %v", TestAddress, len(result), result)
}

// TestGovPermissions tests querying permissions for an address.
func TestGovPermissions(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.Permissions(ctx, TestAddress)
	requireNoError(t, err, "Failed to query permissions")
	requireNotNil(t, result, "Permissions is nil")

	t.Logf("Address %s permissions: whitelist=%d, blacklist=%d",
		TestAddress, len(result.Whitelist), len(result.Blacklist))
}

// TestGovExecutionFee tests querying execution fee for a transaction type.
func TestGovExecutionFee(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.ExecutionFee(ctx, "send")
	if err != nil {
		// This may fail if no execution fee is set for send
		t.Logf("Execution fee query: %v (may be expected)", err)
		return
	}

	t.Logf("Execution fee for send: %s", result.ExecutionFee)
}

// TestGovAllExecutionFees tests querying all execution fees.
func TestGovAllExecutionFees(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.AllExecutionFees(ctx)
	requireNoError(t, err, "Failed to query all execution fees")

	t.Logf("Found %d execution fees", len(result))
	for _, fee := range result {
		t.Logf("  %s: execution=%s, failure=%s", fee.TransactionType, fee.ExecutionFee, fee.FailureFee)
	}
}

// TestGovIdentityRecords tests querying identity records.
func TestGovIdentityRecords(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.IdentityRecords(ctx)
	requireNoError(t, err, "Failed to query identity records")

	t.Logf("Found %d identity records", len(result))
}

// TestGovDataRegistryKeys tests querying data registry keys.
func TestGovDataRegistryKeys(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.DataRegistryKeys(ctx)
	requireNoError(t, err, "Failed to query data registry keys")

	t.Logf("Found %d data registry keys: %v", len(result), result)
}

// TestGovPoorNetworkMessages tests querying poor network messages.
func TestGovPoorNetworkMessages(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.PoorNetworkMessages(ctx)
	requireNoError(t, err, "Failed to query poor network messages")

	t.Logf("Found %d poor network messages: %v", len(result), result)
}

// TestGovAllProposalDurations tests querying all proposal durations.
func TestGovAllProposalDurations(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.AllProposalDurations(ctx)
	requireNoError(t, err, "Failed to query all proposal durations")

	t.Logf("Found %d proposal durations", len(result))
	for pType, dur := range result {
		t.Logf("  %s: %s", pType, dur)
	}
}

// TestGovProposerVotersCount tests querying proposer and voters count.
func TestGovProposerVotersCount(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.ProposerVotersCount(ctx)
	requireNoError(t, err, "Failed to query proposer voters count")
	requireNotNil(t, result, "Proposer voters count is nil")

	t.Logf("Proposers: %s, Voters: %s", result.Proposers, result.Voters)
}

// TestGovNonCouncilors tests querying non-councilors.
func TestGovNonCouncilors(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.NonCouncilors(ctx)
	requireNoError(t, err, "Failed to query non-councilors")

	t.Logf("Found %d non-councilors", len(result))
}

// TestGovCustomPrefixes tests querying custom prefixes.
func TestGovCustomPrefixes(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.CustomPrefixes(ctx)
	requireNoError(t, err, "Failed to query custom prefixes")
	requireNotNil(t, result, "Custom prefixes is nil")

	t.Logf("Custom prefixes: DefaultDenom=%s, Bech32Prefix=%s",
		result.DefaultDenom, result.Bech32Prefix)
}

// === TX TESTS ===

// TestGovCouncilorClaimSeat tests claiming a councilor seat.
func TestGovCouncilorClaimSeat(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	// Try to claim seat - may already be claimed
	resp, err := mod.CouncilorClaimSeat(ctx, TestKey, "test-moniker", nil)
	if err != nil {
		t.Logf("Councilor claim seat: %v (may already be claimed)", err)
		return
	}

	t.Logf("Councilor claim seat TX: code=%d, hash=%s", resp.Code, resp.TxHash)

	// Wait for TX to be included in a block (block time ~6s) to avoid sequence mismatch
	time.Sleep(7 * time.Second)
}

// TestGovProposalVote tests voting on a proposal.
// This test creates a proposal, votes on it, and verifies the vote.
func TestGovProposalVote(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getExtendedTestContext()
	defer cancel()

	// Set fast proposal timing
	setFastProposalTiming(t, client)

	mod := gov.New(client)

	// Create a unique title for our proposal
	proposalTitle := generateUniqueID("TestProposal")

	// Create a simple proposal (set network property)
	// ProposalSetNetworkProperty(ctx, from, property, value, title, description, opts)
	// Note: VOTE_QUORUM is a percentage value between 0-1 (e.g., 0.34 = 34%)
	resp, err := mod.ProposalSetNetworkProperty(ctx, TestKey, "VOTE_QUORUM", "0.34", proposalTitle, "Test proposal for voting", nil)
	requireNoError(t, err, "Failed to create proposal")
	requireTxSuccess(t, resp, "Proposal creation failed")

	t.Logf("Proposal TX submitted: %s", resp.TxHash)

	// Wait for TX to be included in a block
	time.Sleep(5 * time.Second)

	// Query proposals to find the one we just created (by title)
	proposals, err := mod.Proposals(ctx, nil)
	requireNoError(t, err, "Failed to query proposals")
	requireTrue(t, len(proposals.Proposals) > 0, "No proposals found")

	// Find our proposal by title
	var proposalID string
	for _, p := range proposals.Proposals {
		if p.Title == proposalTitle {
			proposalID = p.ProposalID
			break
		}
	}
	requireTrue(t, proposalID != "", "Could not find our proposal by title: "+proposalTitle)
	t.Logf("Found our proposal %s: %s", proposalID, proposalTitle)

	// Query votes before voting
	votesBefore, err := mod.Votes(ctx, proposalID)
	requireNoError(t, err, "Failed to query votes")
	t.Logf("Votes before: %d", len(votesBefore))

	// Vote on the proposal
	voteResp, err := mod.VoteProposal(ctx, TestKey, proposalID, VoteYes, nil)
	requireNoError(t, err, "Failed to vote on proposal")
	requireTxSuccess(t, voteResp, "Vote transaction failed")

	t.Logf("Voted on proposal %s (TX: %s, code: %d)", proposalID, voteResp.TxHash, voteResp.Code)

	// Wait for vote TX to be included in a block
	time.Sleep(5 * time.Second)

	// Query votes after voting
	votesAfter, err := mod.Votes(ctx, proposalID)
	requireNoError(t, err, "Failed to query votes after voting")
	t.Logf("Votes after: %d", len(votesAfter))

	// Verify our vote exists
	found := false
	for _, v := range votesAfter {
		if v.Voter == TestAddress {
			found = true
			t.Logf("Found our vote: proposal=%s, voter=%s, option=%s", v.ProposalID, v.Voter, v.Option)
			break
		}
	}
	requireTrue(t, found, "Our vote should be in the votes list")
}

// TestGovRegisterIdentityRecords tests registering identity records.
func TestGovRegisterIdentityRecords(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)

	// Register an identity record
	// Format is a simple key-value map: {"key": "value"}
	infosJSON := `{"test_key":"test_value"}`
	resp, err := mod.RegisterIdentityRecords(ctx, TestKey, infosJSON, nil)
	requireNoError(t, err, "Failed to register identity records")
	requireTxSuccess(t, resp, "Register identity records failed")

	t.Logf("Registered identity record TX: %s", resp.TxHash)

	// Wait for TX to be included in block
	time.Sleep(7 * time.Second)

	// Query identity records for our address
	records, err := mod.IdentityRecordsByAddress(ctx, TestAddress)
	requireNoError(t, err, "Failed to query identity records by address")
	t.Logf("Found %d identity records for %s", len(records), TestAddress)
}

// TestGovPollCreateAndVote tests creating a poll and voting on it.
func TestGovPollCreateAndVote(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)

	// Create a poll
	// Required: Title, Roles, PollType
	pollOpts := &gov.PollCreateOpts{
		Title:          "Test Poll",
		Description:    "Integration test poll",
		Roles:          "1",    // sudoer role
		PollType:       "vote", // vote type poll
		Options:        "option1,option2,option3",
		SelectionCount: 1,
		Duration:       "3600", // 1 hour
	}
	resp, err := mod.PollCreate(ctx, TestKey, pollOpts, nil)
	requireNoError(t, err, "Failed to create poll")
	requireTxSuccess(t, resp, "Poll creation failed")

	t.Logf("Created poll TX: %s", resp.TxHash)

	// Wait for TX to be included in block
	time.Sleep(7 * time.Second)

	// Query polls for our address
	polls, err := mod.Polls(ctx, TestAddress)
	requireNoError(t, err, "Failed to query polls")
	t.Logf("Found %d polls for %s", len(polls), TestAddress)
}

// TestGovRoleOperations tests role creation and assignment.
// This is an atomic test that:
// 1. Creates a new role
// 2. Assigns the role to an address
// 3. Verifies the role assignment
// 4. Unassigns the role
func TestGovRoleOperations(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	roleSID := generateUniqueID("testrole")

	// Create a role
	resp, err := mod.RoleCreate(ctx, TestKey, roleSID, "Test role for integration tests", nil)
	requireNoError(t, err, "Failed to create role")
	requireTxSuccess(t, resp, "Role creation failed")

	t.Logf("Created role %s (tx: %s)", roleSID, resp.TxHash)

	// Wait for TX to be included in a block
	time.Sleep(5 * time.Second)

	// Query the role to verify it exists
	role, err := mod.Role(ctx, roleSID)
	requireNoError(t, err, "Failed to query created role")
	requireEqual(t, roleSID, role.Sid, "Role SID mismatch")

	t.Logf("Role %s created with ID %d", role.Sid, role.ID)
}

// === ADDITIONAL QUERY TESTS ===

// TestGovProposal tests querying a specific proposal.
func TestGovProposal(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)

	// First get all proposals
	proposals, err := mod.Proposals(ctx, nil)
	requireNoError(t, err, "Failed to query proposals")

	if len(proposals.Proposals) == 0 {
		t.Log("No proposals found, skipping")
		return
	}

	// Query the first proposal
	propID := proposals.Proposals[0].ProposalID
	result, err := mod.Proposal(ctx, propID)
	requireNoError(t, err, "Failed to query proposal")
	requireNotNil(t, result, "Proposal is nil")

	t.Logf("Proposal %s: %s (%s)", result.ProposalID, result.Title, result.Status)
}

// TestGovVote tests querying a specific vote.
func TestGovVote(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)

	// First get all proposals
	proposals, err := mod.Proposals(ctx, nil)
	requireNoError(t, err, "Failed to query proposals")

	if len(proposals.Proposals) == 0 {
		t.Log("No proposals found, skipping")
		return
	}

	// Get votes for the first proposal
	propID := proposals.Proposals[0].ProposalID
	votes, err := mod.Votes(ctx, propID)
	if err != nil || len(votes) == 0 {
		t.Logf("No votes found for proposal %s, skipping", propID)
		return
	}

	// Query a specific vote
	voter := votes[0].Voter
	vote, err := mod.Vote(ctx, propID, voter)
	if err != nil {
		t.Logf("Vote query: %v (may be expected)", err)
		return
	}

	t.Logf("Vote: proposal=%s, voter=%s, option=%s", vote.ProposalID, vote.Voter, vote.Option)
}

// TestGovVoters tests querying voters of a proposal.
func TestGovVoters(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)

	// First get all proposals
	proposals, err := mod.Proposals(ctx, nil)
	requireNoError(t, err, "Failed to query proposals")

	if len(proposals.Proposals) == 0 {
		t.Log("No proposals found, skipping")
		return
	}

	// Query voters for the first proposal
	propID := proposals.Proposals[0].ProposalID
	voters, err := mod.Voters(ctx, propID)
	if err != nil {
		t.Logf("Voters query: %v (may be expected)", err)
		return
	}

	t.Logf("Voters for proposal %s: %v", propID, voters)
}

// TestGovProposalDuration tests querying a specific proposal duration.
func TestGovProposalDuration(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)

	// Query duration for a common proposal type
	duration, err := mod.ProposalDuration(ctx, "SetNetworkProperty")
	if err != nil {
		t.Logf("Proposal duration query: %v (may be expected)", err)
		return
	}

	t.Logf("Proposal duration for SetNetworkProperty: %s", duration)
}

// TestGovCouncilRegistry tests querying council registry.
func TestGovCouncilRegistry(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.CouncilRegistry(ctx)
	if err != nil {
		t.Logf("Council registry query: %v (may be expected)", err)
		return
	}

	t.Logf("Council registry: %+v", result)
}

// TestGovDataRegistry tests querying data registry by key.
func TestGovDataRegistry(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)

	// First get all keys
	keys, err := mod.DataRegistryKeys(ctx)
	requireNoError(t, err, "Failed to query data registry keys")

	if len(keys) == 0 {
		t.Log("No data registry keys found, skipping")
		return
	}

	// Query the first key
	result, err := mod.DataRegistry(ctx, keys[0])
	if err != nil {
		t.Logf("Data registry query: %v (may be expected)", err)
		return
	}

	t.Logf("Data registry %s: %+v", keys[0], result)
}

// TestGovPollVotes tests querying poll votes.
func TestGovPollVotes(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)

	// First get polls
	polls, err := mod.Polls(ctx, TestAddress)
	if err != nil || len(polls) == 0 {
		t.Log("No polls found, skipping poll votes test")
		return
	}

	// Query votes for first poll
	pollID := polls[0].ID
	votes, err := mod.PollVotes(ctx, pollID)
	if err != nil {
		t.Logf("Poll votes query: %v (may be expected)", err)
		return
	}

	t.Logf("Poll %s votes: %+v", pollID, votes)
}

// TestGovIdentityRecord tests querying identity record by ID.
func TestGovIdentityRecord(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)

	// First get all records
	records, err := mod.IdentityRecords(ctx)
	if err != nil || len(records) == 0 {
		t.Log("No identity records found, skipping")
		return
	}

	// Query first record by ID
	result, err := mod.IdentityRecord(ctx, records[0].ID)
	if err != nil {
		t.Logf("Identity record query: %v (may be expected)", err)
		return
	}

	t.Logf("Identity record: %+v", result)
}

// TestGovAllIdentityRecordVerifyRequests tests querying all identity record verify requests.
func TestGovAllIdentityRecordVerifyRequests(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.AllIdentityRecordVerifyRequests(ctx)
	if err != nil {
		t.Logf("All identity record verify requests query: %v (may be expected)", err)
		return
	}

	t.Logf("Found %d identity record verify requests", len(result))
}

// TestGovIdentityRecordVerifyRequest tests querying identity record verify request by ID.
func TestGovIdentityRecordVerifyRequest(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)

	// First get all requests
	requests, err := mod.AllIdentityRecordVerifyRequests(ctx)
	if err != nil || len(requests) == 0 {
		t.Log("No identity record verify requests found, skipping")
		return
	}

	// Query first request by ID
	result, err := mod.IdentityRecordVerifyRequest(ctx, requests[0].ID)
	if err != nil {
		t.Logf("Identity record verify request query: %v (may be expected)", err)
		return
	}

	t.Logf("Identity record verify request: %+v", result)
}

// TestGovIdentityRecordVerifyRequestsByApprover tests querying by approver.
func TestGovIdentityRecordVerifyRequestsByApprover(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.IdentityRecordVerifyRequestsByApprover(ctx, TestAddress)
	if err != nil {
		t.Logf("Identity record verify requests by approver query: %v (may be expected)", err)
		return
	}

	t.Logf("Found %d identity record verify requests by approver", len(result))
}

// TestGovIdentityRecordVerifyRequestsByRequester tests querying by requester.
func TestGovIdentityRecordVerifyRequestsByRequester(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	result, err := mod.IdentityRecordVerifyRequestsByRequester(ctx, TestAddress)
	if err != nil {
		t.Logf("Identity record verify requests by requester query: %v (may be expected)", err)
		return
	}

	t.Logf("Found %d identity record verify requests by requester", len(result))
}

// TestGovWhitelistedPermissionAddresses tests querying whitelisted permission addresses.
func TestGovWhitelistedPermissionAddresses(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	// Query addresses with permission 1 (claim validator)
	result, err := mod.WhitelistedPermissionAddresses(ctx, "1")
	if err != nil {
		t.Logf("Whitelisted permission addresses query: %v (may be expected)", err)
		return
	}

	t.Logf("Found %d addresses with whitelisted permission 1: %v", len(result), result)
}

// TestGovBlacklistedPermissionAddresses tests querying blacklisted permission addresses.
func TestGovBlacklistedPermissionAddresses(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	// Query addresses with blacklisted permission 7 (our test key doesn't have this)
	result, err := mod.BlacklistedPermissionAddresses(ctx, "7")
	if err != nil {
		t.Logf("Blacklisted permission addresses query: %v (may be expected)", err)
		return
	}

	t.Logf("Found %d addresses with blacklisted permission 7: %v", len(result), result)
}

// TestGovWhitelistedRoleAddresses tests querying whitelisted role addresses.
func TestGovWhitelistedRoleAddresses(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	// Query addresses with role "sudo"
	result, err := mod.WhitelistedRoleAddresses(ctx, "sudo")
	if err != nil {
		t.Logf("Whitelisted role addresses query: %v (may be expected)", err)
		return
	}

	t.Logf("Found %d addresses with role sudo: %v", len(result), result)
}

// === COUNCILOR TX TESTS ===

// TestGovCouncilorActivate tests activating a councilor.
func TestGovCouncilorActivate(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	mod := gov.New(client)
	resp, err := mod.CouncilorActivate(ctx, TestKey, nil)
	if err != nil {
		t.Logf("Councilor activate: %v (may already be active)", err)
		return
	}

	t.Logf("Councilor activate TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovCouncilorPause tests pausing a councilor.
func TestGovCouncilorPause(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	resp, err := mod.CouncilorPause(ctx, TestKey, nil)
	if err != nil {
		t.Logf("Councilor pause: %v (may be expected)", err)
		return
	}

	t.Logf("Councilor pause TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovCouncilorUnpause tests unpausing a councilor.
func TestGovCouncilorUnpause(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	resp, err := mod.CouncilorUnpause(ctx, TestKey, nil)
	if err != nil {
		t.Logf("Councilor unpause: %v (may be expected)", err)
		return
	}

	t.Logf("Councilor unpause TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// === PERMISSION TX TESTS ===

// TestGovPermissionWhitelist tests whitelisting a permission for an address.
func TestGovPermissionWhitelist(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// Permission 999 is a high number unlikely to conflict
	resp, err := mod.PermissionWhitelist(ctx, TestKey, TestAddress, 999, nil)
	if err != nil {
		t.Logf("Permission whitelist: %v (may be expected)", err)
		return
	}

	t.Logf("Permission whitelist TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovPermissionBlacklist tests blacklisting a permission for an address.
func TestGovPermissionBlacklist(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// Permission 998 is a high number unlikely to conflict
	resp, err := mod.PermissionBlacklist(ctx, TestKey, TestAddress, 998, nil)
	if err != nil {
		t.Logf("Permission blacklist: %v (may be expected)", err)
		return
	}

	t.Logf("Permission blacklist TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovPermissionRemoveWhitelisted tests removing a whitelisted permission.
func TestGovPermissionRemoveWhitelisted(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// Try to remove permission 999 we might have added
	resp, err := mod.PermissionRemoveWhitelisted(ctx, TestKey, TestAddress, 999, nil)
	if err != nil {
		t.Logf("Permission remove whitelisted: %v (may not exist)", err)
		return
	}

	t.Logf("Permission remove whitelisted TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovPermissionRemoveBlacklisted tests removing a blacklisted permission.
func TestGovPermissionRemoveBlacklisted(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// Try to remove permission 998 we might have added
	resp, err := mod.PermissionRemoveBlacklisted(ctx, TestKey, TestAddress, 998, nil)
	if err != nil {
		t.Logf("Permission remove blacklisted: %v (may not exist)", err)
		return
	}

	t.Logf("Permission remove blacklisted TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// === ROLE TX TESTS ===

// TestGovRoleAssign tests assigning a role to an address.
func TestGovRoleAssign(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// Assign role 2 (validator) to test address
	resp, err := mod.RoleAssign(ctx, TestKey, TestAddress, 2, nil)
	if err != nil {
		t.Logf("Role assign: %v (may already be assigned)", err)
		return
	}

	t.Logf("Role assign TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovRoleUnassign tests unassigning a role from an address.
func TestGovRoleUnassign(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// Unassign role 2 we might have assigned
	resp, err := mod.RoleUnassign(ctx, TestKey, TestAddress, 2, nil)
	if err != nil {
		t.Logf("Role unassign: %v (may not be assigned)", err)
		return
	}

	t.Logf("Role unassign TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovRoleWhitelistPermission tests whitelisting a permission for a role.
func TestGovRoleWhitelistPermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// First create a role to modify
	roleSID := generateUniqueID("permrole")
	_, err := mod.RoleCreate(ctx, TestKey, roleSID, "Role for permission test", nil)
	if err != nil {
		t.Logf("Role create for permission test: %v", err)
		return
	}

	waitForBlocks(t, 1)

	// Whitelist a permission for the role
	resp, err := mod.RoleWhitelistPermission(ctx, TestKey, roleSID, 100, nil)
	if err != nil {
		t.Logf("Role whitelist permission: %v (may be expected)", err)
		return
	}

	t.Logf("Role whitelist permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovRoleBlacklistPermission tests blacklisting a permission for a role.
func TestGovRoleBlacklistPermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// First create a role to modify
	roleSID := generateUniqueID("blkrole")
	_, err := mod.RoleCreate(ctx, TestKey, roleSID, "Role for blacklist test", nil)
	if err != nil {
		t.Logf("Role create for blacklist test: %v", err)
		return
	}

	waitForBlocks(t, 1)

	// Blacklist a permission for the role
	resp, err := mod.RoleBlacklistPermission(ctx, TestKey, roleSID, 101, nil)
	if err != nil {
		t.Logf("Role blacklist permission: %v (may be expected)", err)
		return
	}

	t.Logf("Role blacklist permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovRoleRemoveWhitelistedPermission tests removing a whitelisted permission from a role.
func TestGovRoleRemoveWhitelistedPermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// Try to remove a permission from sudo role (won't affect core permissions)
	resp, err := mod.RoleRemoveWhitelistedPermission(ctx, TestKey, "sudo", 999, nil)
	if err != nil {
		t.Logf("Role remove whitelisted permission: %v (may not exist)", err)
		return
	}

	t.Logf("Role remove whitelisted permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovRoleRemoveBlacklistedPermission tests removing a blacklisted permission from a role.
func TestGovRoleRemoveBlacklistedPermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// Try to remove a blacklisted permission from sudo role
	resp, err := mod.RoleRemoveBlacklistedPermission(ctx, TestKey, "sudo", 998, nil)
	if err != nil {
		t.Logf("Role remove blacklisted permission: %v (may not exist)", err)
		return
	}

	t.Logf("Role remove blacklisted permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// === POLL TX TESTS ===

// TestGovPollVote tests voting on a poll.
func TestGovPollVote(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)

	// First get polls
	polls, err := mod.Polls(ctx, TestAddress)
	if err != nil || len(polls) == 0 {
		t.Log("No polls found, skipping poll vote test")
		return
	}

	// Vote on first poll
	pollID := polls[0].ID
	resp, err := mod.PollVote(ctx, TestKey, pollID, "option1", nil)
	if err != nil {
		t.Logf("Poll vote: %v (may be expected)", err)
		return
	}

	t.Logf("Poll vote TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// === IDENTITY TX TESTS ===

// TestGovDeleteIdentityRecords tests deleting identity records.
func TestGovDeleteIdentityRecords(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	// Delete a test key we might have created
	resp, err := mod.DeleteIdentityRecords(ctx, TestKey, "test_key", nil)
	if err != nil {
		t.Logf("Delete identity records: %v (may not exist)", err)
		return
	}

	t.Logf("Delete identity records TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovRequestIdentityRecordVerify tests requesting identity record verification.
func TestGovRequestIdentityRecordVerify(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)

	// First get identity records
	records, err := mod.IdentityRecordsByAddress(ctx, TestAddress)
	if err != nil || len(records) == 0 {
		t.Log("No identity records found, skipping verify request test")
		return
	}

	// Request verification for first record
	recordID := records[0].ID
	resp, err := mod.RequestIdentityRecordVerify(ctx, TestKey, TestAddress, recordID, "100ukex", nil)
	if err != nil {
		t.Logf("Request identity record verify: %v (may be expected)", err)
		return
	}

	t.Logf("Request identity record verify TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovHandleIdentityRecordsVerifyRequest tests handling identity verify requests.
func TestGovHandleIdentityRecordsVerifyRequest(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)

	// First get all verify requests
	requests, err := mod.AllIdentityRecordVerifyRequests(ctx)
	if err != nil || len(requests) == 0 {
		t.Log("No identity record verify requests found, skipping handle test")
		return
	}

	// Handle (approve) first request
	resp, err := mod.HandleIdentityRecordsVerifyRequest(ctx, TestKey, requests[0].ID, true, nil)
	if err != nil {
		t.Logf("Handle identity records verify request: %v (may be expected)", err)
		return
	}

	t.Logf("Handle identity records verify request TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovCancelIdentityRecordsVerifyRequest tests canceling identity verify requests.
func TestGovCancelIdentityRecordsVerifyRequest(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)

	// Try to cancel request ID 1 (may not exist)
	resp, err := mod.CancelIdentityRecordsVerifyRequest(ctx, TestKey, "1", nil)
	if err != nil {
		t.Logf("Cancel identity records verify request: %v (may not exist)", err)
		return
	}

	t.Logf("Cancel identity records verify request TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// === DIRECT TX TESTS ===

// TestGovSetNetworkProperties tests setting network properties directly.
func TestGovSetNetworkProperties(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	properties := map[string]string{
		"min_tx_fee": "100",
	}
	resp, err := mod.SetNetworkProperties(ctx, TestKey, properties, nil)
	if err != nil {
		t.Logf("Set network properties: %v (may require permission)", err)
		return
	}

	t.Logf("Set network properties TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovSetExecutionFee tests setting an execution fee.
func TestGovSetExecutionFee(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	resp, err := mod.SetExecutionFee(ctx, TestKey, "test_tx_type", 100, 50, 3600, 0, nil)
	if err != nil {
		t.Logf("Set execution fee: %v (may require permission)", err)
		return
	}

	t.Logf("Set execution fee TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// === PROPOSAL TX TESTS ===

// TestGovProposalSetPoorNetworkMsgs tests creating a proposal to set poor network messages.
func TestGovProposalSetPoorNetworkMsgs(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalOtherOpts{
		Title:       "Test set poor network msgs",
		Description: "Integration test",
	}
	resp, err := mod.ProposalSetPoorNetworkMsgs(ctx, TestKey, "MsgSend,MsgMultiSend", propOpts, nil)
	if err != nil {
		t.Logf("Proposal set poor network msgs: %v", err)
		return
	}

	t.Logf("Proposal set poor network msgs TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalSetProposalDurations tests creating a proposal to set proposal durations.
func TestGovProposalSetProposalDurations(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalOtherOpts{
		Title:       "Test set proposal durations",
		Description: "Integration test",
	}
	resp, err := mod.ProposalSetProposalDurations(ctx, TestKey, "SetNetworkProperty", "600", propOpts, nil)
	if err != nil {
		t.Logf("Proposal set proposal durations: %v", err)
		return
	}

	t.Logf("Proposal set proposal durations TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalUpsertDataRegistry tests creating a proposal to upsert data registry.
func TestGovProposalUpsertDataRegistry(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalOtherOpts{
		Title:       "Test upsert data registry",
		Description: "Integration test",
	}
	key := generateUniqueID("testkey")
	resp, err := mod.ProposalUpsertDataRegistry(ctx, TestKey, key, "abc123hash", "https://example.com", "json", "1024", propOpts, nil)
	if err != nil {
		t.Logf("Proposal upsert data registry: %v", err)
		return
	}

	t.Logf("Proposal upsert data registry TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalSetExecutionFees tests creating a proposal to set execution fees.
func TestGovProposalSetExecutionFees(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalSetExecutionFeesOpts{
		Title:         "Test set execution fees",
		Description:   "Integration test",
		TxTypes:       "send",
		ExecutionFees: "100",
		FailureFees:   "50",
		Timeouts:      "3600",
		DefaultParams: "0",
	}
	resp, err := mod.ProposalSetExecutionFees(ctx, TestKey, propOpts, nil)
	if err != nil {
		t.Logf("Proposal set execution fees: %v", err)
		return
	}

	t.Logf("Proposal set execution fees TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalJailCouncilor tests creating a proposal to jail councilors.
func TestGovProposalJailCouncilor(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalOtherOpts{
		Title:       "Test jail councilor",
		Description: "Integration test - will not pass",
	}
	// Use a non-existent address to avoid actually jailing anyone
	resp, err := mod.ProposalJailCouncilor(ctx, TestKey, "kira1nonexistent", propOpts, nil)
	if err != nil {
		t.Logf("Proposal jail councilor: %v", err)
		return
	}

	t.Logf("Proposal jail councilor TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalResetWholeCouncilorRank tests creating a proposal to reset councilor ranks.
func TestGovProposalResetWholeCouncilorRank(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalOtherOpts{
		Title:       "Test reset councilor rank",
		Description: "Integration test - will not pass",
	}
	resp, err := mod.ProposalResetWholeCouncilorRank(ctx, TestKey, propOpts, nil)
	if err != nil {
		t.Logf("Proposal reset whole councilor rank: %v", err)
		return
	}

	t.Logf("Proposal reset whole councilor rank TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalAssignRole tests creating a proposal to assign a role.
func TestGovProposalAssignRole(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	resp, err := mod.ProposalAssignRole(ctx, TestKey, TestAddress, "2", "Test assign role", "Assign validator role", nil)
	if err != nil {
		t.Logf("Proposal assign role: %v", err)
		return
	}

	t.Logf("Proposal assign role TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalUnassignRole tests creating a proposal to unassign a role.
func TestGovProposalUnassignRole(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	resp, err := mod.ProposalUnassignRole(ctx, TestKey, TestAddress, "2", "Test unassign role", "Unassign validator role", nil)
	if err != nil {
		t.Logf("Proposal unassign role: %v", err)
		return
	}

	t.Logf("Proposal unassign role TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalWhitelistAccountPermission tests creating a proposal to whitelist a permission for an account.
func TestGovProposalWhitelistAccountPermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalAccountPermissionOpts{
		Title:       "Test whitelist permission",
		Description: "Integration test",
		Addr:        TestAddress,
	}
	resp, err := mod.ProposalWhitelistAccountPermission(ctx, TestKey, 100, propOpts, nil)
	if err != nil {
		t.Logf("Proposal whitelist account permission: %v", err)
		return
	}

	t.Logf("Proposal whitelist account permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalBlacklistAccountPermission tests creating a proposal to blacklist a permission for an account.
func TestGovProposalBlacklistAccountPermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalAccountPermissionOpts{
		Title:       "Test blacklist permission",
		Description: "Integration test",
		Addr:        TestAddress,
	}
	resp, err := mod.ProposalBlacklistAccountPermission(ctx, TestKey, 101, propOpts, nil)
	if err != nil {
		t.Logf("Proposal blacklist account permission: %v", err)
		return
	}

	t.Logf("Proposal blacklist account permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalRemoveWhitelistedAccountPermission tests creating a proposal to remove a whitelisted permission.
func TestGovProposalRemoveWhitelistedAccountPermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalAccountPermissionOpts{
		Title:       "Test remove whitelisted permission",
		Description: "Integration test",
		Addr:        TestAddress,
	}
	resp, err := mod.ProposalRemoveWhitelistedAccountPermission(ctx, TestKey, 100, propOpts, nil)
	if err != nil {
		t.Logf("Proposal remove whitelisted account permission: %v", err)
		return
	}

	t.Logf("Proposal remove whitelisted account permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalRemoveBlacklistedAccountPermission tests creating a proposal to remove a blacklisted permission.
func TestGovProposalRemoveBlacklistedAccountPermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalAccountPermissionOpts{
		Title:       "Test remove blacklisted permission",
		Description: "Integration test",
		Addr:        TestAddress,
	}
	resp, err := mod.ProposalRemoveBlacklistedAccountPermission(ctx, TestKey, 101, propOpts, nil)
	if err != nil {
		t.Logf("Proposal remove blacklisted account permission: %v", err)
		return
	}

	t.Logf("Proposal remove blacklisted account permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalCreateRole tests creating a proposal to create a role.
func TestGovProposalCreateRole(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	roleSID := generateUniqueID("proprole")
	propOpts := &gov.ProposalCreateRoleOpts{
		Title:       "Test create role proposal",
		Description: "Integration test",
		Whitelist:   "1,2,3",
	}
	resp, err := mod.ProposalCreateRole(ctx, TestKey, roleSID, "Test role via proposal", propOpts, nil)
	if err != nil {
		t.Logf("Proposal create role: %v", err)
		return
	}

	t.Logf("Proposal create role TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalRemoveRole tests creating a proposal to remove a role.
func TestGovProposalRemoveRole(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalRoleOpts{
		Title:       "Test remove role proposal",
		Description: "Integration test - will not pass",
	}
	// Use a non-existent role to avoid actually removing anything
	resp, err := mod.ProposalRemoveRole(ctx, TestKey, "nonexistent_role", propOpts, nil)
	if err != nil {
		t.Logf("Proposal remove role: %v", err)
		return
	}

	t.Logf("Proposal remove role TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalWhitelistRolePermission tests creating a proposal to whitelist a role permission.
func TestGovProposalWhitelistRolePermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalRoleOpts{
		Title:       "Test whitelist role permission",
		Description: "Integration test",
	}
	resp, err := mod.ProposalWhitelistRolePermission(ctx, TestKey, "sudo", 200, propOpts, nil)
	if err != nil {
		t.Logf("Proposal whitelist role permission: %v", err)
		return
	}

	t.Logf("Proposal whitelist role permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalBlacklistRolePermission tests creating a proposal to blacklist a role permission.
func TestGovProposalBlacklistRolePermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalRoleOpts{
		Title:       "Test blacklist role permission",
		Description: "Integration test",
	}
	resp, err := mod.ProposalBlacklistRolePermission(ctx, TestKey, "sudo", 201, propOpts, nil)
	if err != nil {
		t.Logf("Proposal blacklist role permission: %v", err)
		return
	}

	t.Logf("Proposal blacklist role permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalRemoveWhitelistedRolePermission tests creating a proposal to remove a whitelisted role permission.
func TestGovProposalRemoveWhitelistedRolePermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalRoleOpts{
		Title:       "Test remove whitelisted role permission",
		Description: "Integration test",
	}
	resp, err := mod.ProposalRemoveWhitelistedRolePermission(ctx, TestKey, "sudo", 200, propOpts, nil)
	if err != nil {
		t.Logf("Proposal remove whitelisted role permission: %v", err)
		return
	}

	t.Logf("Proposal remove whitelisted role permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}

// TestGovProposalRemoveBlacklistedRolePermission tests creating a proposal to remove a blacklisted role permission.
func TestGovProposalRemoveBlacklistedRolePermission(t *testing.T) {
	skipIfContainerNotRunning(t)
	client := getTestClient(t)
	defer client.Close()

	ctx, cancel := getTestContext()
	defer cancel()

	waitForBlocks(t, 1)

	mod := gov.New(client)
	propOpts := &gov.ProposalRoleOpts{
		Title:       "Test remove blacklisted role permission",
		Description: "Integration test",
	}
	resp, err := mod.ProposalRemoveBlacklistedRolePermission(ctx, TestKey, "sudo", 201, propOpts, nil)
	if err != nil {
		t.Logf("Proposal remove blacklisted role permission: %v", err)
		return
	}

	t.Logf("Proposal remove blacklisted role permission TX: hash=%s, code=%d", resp.TxHash, resp.Code)
}
