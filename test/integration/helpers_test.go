// Package integration provides shared test utilities for integration tests.
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kiracore/sekai-cli/pkg/sdk"
	"github.com/kiracore/sekai-cli/pkg/sdk/client/docker"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/gov"
)

// Test configuration constants
const (
	TestContainer = "sekin-sekai-1"
	TestAddress   = "kira1cw0wz6x9wy8wvw30q8qsxppgzqrr5qu846cut5"
	TestKey       = "genesis"
	TestChainID   = "testnet-1"
	TestHome      = "/sekai"
	TestFees      = "100ukex"
)

// Vote options
const (
	VoteYes        = 1
	VoteNo         = 2
	VoteAbstain    = 3
	VoteNoWithVeto = 4
)

// getTestClient creates a new test client with standard configuration.
func getTestClient(t *testing.T) sdk.Client {
	t.Helper()
	client, err := docker.NewClient(TestContainer,
		docker.WithChainID(TestChainID),
		docker.WithKeyringBackend("test"),
		docker.WithHome(TestHome),
		docker.WithFees(TestFees),
	)
	if err != nil {
		t.Fatalf("Failed to create docker client: %v", err)
	}
	return client
}

// getTestContext creates a context with standard timeout.
func getTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 60*time.Second)
}

// getExtendedTestContext creates a context with extended timeout for proposal tests.
// Proposal tests need ~5 min voting + ~5 min enactment + buffer = ~15 min.
func getExtendedTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Minute)
}

// requireNoError fails the test if err is not nil.
func requireNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%s: %v", fmt.Sprint(msgAndArgs...), err)
		} else {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
}

// requireError fails the test if err is nil.
func requireError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected error but got nil: %s", fmt.Sprint(msgAndArgs...))
		} else {
			t.Fatal("Expected error but got nil")
		}
	}
}

// requireNotNil fails the test if obj is nil.
func requireNotNil(t *testing.T, obj interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if obj == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected non-nil but got nil: %s", fmt.Sprint(msgAndArgs...))
		} else {
			t.Fatal("Expected non-nil but got nil")
		}
	}
}

// requireTrue fails the test if condition is false.
func requireTrue(t *testing.T, condition bool, msgAndArgs ...interface{}) {
	t.Helper()
	if !condition {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Expected true but got false: %s", fmt.Sprint(msgAndArgs...))
		} else {
			t.Fatal("Expected true but got false")
		}
	}
}

// requireEqual fails the test if expected != actual.
func requireEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if expected != actual {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%s: expected %v, got %v", fmt.Sprint(msgAndArgs...), expected, actual)
		} else {
			t.Fatalf("Expected %v, got %v", expected, actual)
		}
	}
}

// generateUniqueID generates a unique identifier for test resources.
func generateUniqueID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// parseJSON parses JSON data into the given target.
func parseJSON(t *testing.T, data json.RawMessage, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
}

// TxResult represents a simplified transaction result.
type TxResult struct {
	Code      int    `json:"code"`
	TxHash    string `json:"txhash"`
	RawLog    string `json:"raw_log"`
	GasUsed   string `json:"gas_used"`
	GasWanted string `json:"gas_wanted"`
}

// isTxSuccess returns true if the transaction was successful.
func isTxSuccess(resp *sdk.TxResponse) bool {
	if resp == nil {
		return false
	}
	return resp.Code == 0
}

// requireTxSuccess verifies that a transaction was successful.
func requireTxSuccess(t *testing.T, resp *sdk.TxResponse, msgAndArgs ...interface{}) {
	t.Helper()
	if resp == nil {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%s: transaction response is nil", fmt.Sprint(msgAndArgs...))
		} else {
			t.Fatal("Transaction response is nil")
		}
	}
	if resp.Code != 0 {
		if len(msgAndArgs) > 0 {
			t.Fatalf("%s: transaction failed with code %d: %s", fmt.Sprint(msgAndArgs...), resp.Code, resp.RawLog)
		} else {
			t.Fatalf("Transaction failed with code %d: %s", resp.Code, resp.RawLog)
		}
	}
}

// getProposalIDFromLog extracts proposal ID from transaction log.
func getProposalIDFromLog(t *testing.T, resp *sdk.TxResponse) string {
	t.Helper()
	if resp == nil {
		t.Fatal("Cannot extract proposal ID from nil response")
	}

	// Try to find proposal_id in the raw log
	re := regexp.MustCompile(`"proposal_id":\s*"?(\d+)"?`)
	matches := re.FindStringSubmatch(resp.RawLog)
	if len(matches) > 1 {
		return matches[1]
	}

	// Try alternative patterns
	re2 := regexp.MustCompile(`proposal_id[^\d]*(\d+)`)
	matches = re2.FindStringSubmatch(resp.RawLog)
	if len(matches) > 1 {
		return matches[1]
	}

	// Raw log is empty (sync mode), return empty to signal fallback needed
	return ""
}

// getLatestProposalID queries proposals and returns the ID of the latest one.
func getLatestProposalID(t *testing.T, client sdk.Client) string {
	t.Helper()
	govMod := gov.New(client)
	ctx := context.Background()

	// Query all proposals
	proposals, err := govMod.Proposals(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to query proposals: %v", err)
	}

	if len(proposals.Proposals) == 0 {
		t.Fatal("No proposals found")
	}

	// Find the highest proposal ID
	var maxID int64
	var maxIDStr string
	for _, p := range proposals.Proposals {
		id, _ := strconv.ParseInt(p.ProposalID, 10, 64)
		if id > maxID {
			maxID = id
			maxIDStr = p.ProposalID
		}
	}

	return maxIDStr
}

// waitForNewProposal waits for a proposal with ID > previousID to appear.
func waitForNewProposal(t *testing.T, client sdk.Client, previousID int64, timeout time.Duration) string {
	t.Helper()
	govMod := gov.New(client)
	ctx := context.Background()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		proposals, err := govMod.Proposals(ctx, nil)
		if err != nil {
			t.Logf("Error querying proposals: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, p := range proposals.Proposals {
			id, _ := strconv.ParseInt(p.ProposalID, 10, 64)
			if id > previousID {
				t.Logf("Found new proposal %s (previous max was %d)", p.ProposalID, previousID)
				return p.ProposalID
			}
		}

		time.Sleep(2 * time.Second)
	}

	t.Fatalf("Timeout waiting for new proposal (previous max was %d)", previousID)
	return ""
}

// waitForProposal waits for a proposal to reach the target result.
// Note: The API uses "result" field not "status" for proposal state.
func waitForProposal(t *testing.T, client sdk.Client, proposalID string, targetResult string, timeout time.Duration) {
	t.Helper()
	govMod := gov.New(client)
	ctx := context.Background()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		proposal, err := govMod.Proposal(ctx, proposalID)
		if err != nil {
			t.Logf("Error querying proposal %s: %v", proposalID, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Use Result field (API returns "result" not "status")
		t.Logf("Proposal %s result: %s (target: %s)", proposalID, proposal.Result, targetResult)
		if strings.EqualFold(proposal.Result, targetResult) {
			return
		}

		// VOTE_RESULT_ENACTMENT means proposal passed and is being enacted
		if strings.EqualFold(targetResult, "VOTE_RESULT_PASSED") &&
			strings.EqualFold(proposal.Result, "VOTE_RESULT_ENACTMENT") {
			t.Logf("Proposal %s is in enactment phase (passed)", proposalID)
			return
		}

		// Check if proposal failed
		if strings.Contains(strings.ToLower(proposal.Result), "failed") ||
			strings.Contains(strings.ToLower(proposal.Result), "rejected") ||
			strings.Contains(strings.ToLower(proposal.Result), "quorum_not_reached") {
			t.Fatalf("Proposal %s reached terminal result: %s", proposalID, proposal.Result)
		}

		time.Sleep(3 * time.Second)
	}

	t.Fatalf("Timeout waiting for proposal %s to reach result %s", proposalID, targetResult)
}

// waitForProposalPassed waits for a proposal to pass.
func waitForProposalPassed(t *testing.T, client sdk.Client, proposalID string, timeout time.Duration) {
	t.Helper()
	// API returns "VOTE_RESULT_PASSED" when proposal passes
	waitForProposal(t, client, proposalID, "VOTE_RESULT_PASSED", timeout)
}

// waitForProposalEnacted waits for a proposal to be fully enacted and executed.
// This waits for the exec_result to be "executed successfully".
func waitForProposalEnacted(t *testing.T, client sdk.Client, proposalID string, timeout time.Duration) {
	t.Helper()
	govMod := gov.New(client)
	ctx := context.Background()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		proposal, err := govMod.Proposal(ctx, proposalID)
		if err != nil {
			t.Logf("Error querying proposal %s: %v", proposalID, err)
			time.Sleep(5 * time.Second)
			continue
		}

		t.Logf("Proposal %s result: %s, exec_result: %s", proposalID, proposal.Result, proposal.ExecResult)

		// Check for successful execution
		if strings.Contains(strings.ToLower(proposal.ExecResult), "executed successfully") {
			t.Logf("Proposal %s executed successfully", proposalID)
			return
		}

		// Check for failed execution
		if strings.Contains(strings.ToLower(proposal.ExecResult), "failed") {
			t.Fatalf("Proposal %s execution failed: %s", proposalID, proposal.ExecResult)
		}

		// Still waiting for enactment
		if strings.EqualFold(proposal.Result, "VOTE_RESULT_ENACTMENT") ||
			strings.EqualFold(proposal.Result, "VOTE_PENDING") {
			time.Sleep(10 * time.Second)
			continue
		}

		// If result is PASSED but no exec_result yet, wait a bit more
		if strings.EqualFold(proposal.Result, "VOTE_RESULT_PASSED") && proposal.ExecResult == "" {
			time.Sleep(10 * time.Second)
			continue
		}

		time.Sleep(10 * time.Second)
	}

	t.Fatalf("Timeout waiting for proposal %s to complete enactment", proposalID)
}

// voteYesOnProposal votes YES on a proposal.
func voteYesOnProposal(t *testing.T, client sdk.Client, proposalID string) {
	t.Helper()
	govMod := gov.New(client)
	ctx, cancel := getTestContext()
	defer cancel()

	resp, err := govMod.VoteProposal(ctx, TestKey, proposalID, VoteYes, nil)
	requireNoError(t, err, "Failed to vote on proposal")
	requireTxSuccess(t, resp, "Vote transaction failed")
	t.Logf("Voted YES on proposal %s", proposalID)
}

// submitAndPassProposal submits a proposal, votes YES, and waits for it to pass.
func submitAndPassProposal(t *testing.T, client sdk.Client, proposalFunc func() (*sdk.TxResponse, error)) string {
	t.Helper()

	// Record current max proposal ID before submitting
	currentMaxID := getLatestProposalID(t, client)
	currentMaxIDInt, _ := strconv.ParseInt(currentMaxID, 10, 64)
	t.Logf("Current max proposal ID: %s", currentMaxID)

	// Submit proposal
	resp, err := proposalFunc()
	requireNoError(t, err, "Failed to submit proposal")
	if resp != nil {
		t.Logf("TX Response: Code=%d, TxHash=%s, RawLog=%s", resp.Code, resp.TxHash, resp.RawLog)
	}
	requireTxSuccess(t, resp, "Proposal submission failed")

	// Extract proposal ID from log
	proposalID := getProposalIDFromLog(t, resp)

	// If raw log was empty (sync mode), wait for new proposal to appear
	if proposalID == "" {
		t.Log("Raw log empty, waiting for new proposal to appear...")
		proposalID = waitForNewProposal(t, client, currentMaxIDInt, 30*time.Second)
	}

	t.Logf("Submitted proposal ID: %s", proposalID)

	// Vote YES
	voteYesOnProposal(t, client, proposalID)

	// Wait for proposal to pass (voting period is typically 5 minutes)
	waitForProposalPassed(t, client, proposalID, 6*time.Minute)

	// Wait for enactment to complete (enactment period is typically 5 more minutes)
	waitForProposalEnacted(t, client, proposalID, 6*time.Minute)

	return proposalID
}

// setFastProposalTiming sets fast proposal timing for testing.
// This creates governance proposals to update MIN_PROPOSAL_END_BLOCKS and
// MIN_PROPOSAL_ENACTMENT_BLOCKS. Since genesis has sudo permissions, proposals
// pass with a single YES vote.
// Note: This function is intentionally skipped as it requires existing fast timing
// to be useful. The network should be initialized with fast timing for tests.
func setFastProposalTiming(t *testing.T, client sdk.Client) {
	t.Helper()
	// Note: Setting proposal timing requires governance proposals, which creates
	// a chicken-and-egg problem. For integration tests, ensure the network is
	// initialized with fast proposal timing (MIN_PROPOSAL_END_BLOCKS=10,
	// MIN_PROPOSAL_ENACTMENT_BLOCKS=5) in genesis or via prior configuration.
	//
	// The set-network-properties command only supports:
	// --max_tx_fee, --min_custody_reward, --min_tx_fee, --min_validators
	//
	// For other properties like MIN_PROPOSAL_END_BLOCKS, use:
	// sekaid tx customgov proposal set-network-property MIN_PROPOSAL_END_BLOCKS 10 --from=genesis
	t.Log("Note: Fast proposal timing should be pre-configured in network genesis")
}

// parseInt64 parses a string to int64, returns 0 on error.
func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// parseUint64 parses a string to uint64, returns 0 on error.
func parseUint64(s string) uint64 {
	n, _ := strconv.ParseUint(s, 10, 64)
	return n
}

// contains checks if a slice contains an element.
func contains[T comparable](slice []T, element T) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// containsString checks if a string slice contains a string.
func containsString(slice []string, s string) bool {
	return contains(slice, s)
}

// waitForBlocks waits for a specified number of blocks.
func waitForBlocks(t *testing.T, blocks int) {
	t.Helper()
	// Approximate 5-6 seconds per block
	duration := time.Duration(blocks) * 6 * time.Second
	t.Logf("Waiting for ~%d blocks (%v)", blocks, duration)
	time.Sleep(duration)
}

// skipIfContainerNotRunning skips the test if the container is not running.
func skipIfContainerNotRunning(t *testing.T) {
	t.Helper()
	client, err := docker.NewClient(TestContainer)
	if err != nil {
		t.Skipf("Skipping test: container %s is not running: %v", TestContainer, err)
	}
	client.Close()
}
