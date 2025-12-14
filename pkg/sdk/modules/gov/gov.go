// Package gov provides governance (customgov) module functionality.
package gov

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides governance query functionality.
type Module struct {
	client sdk.Client
}

// New creates a new gov module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// NetworkProperties queries network properties.
func (m *Module) NetworkProperties(ctx context.Context) (*NetworkProperties, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "network-properties",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query network properties: %w", err)
	}

	var result struct {
		Properties NetworkProperties `json:"properties"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		var props NetworkProperties
		if err := json.Unmarshal(resp.Data, &props); err != nil {
			return nil, fmt.Errorf("failed to parse network properties: %w", err)
		}
		return &props, nil
	}
	return &result.Properties, nil
}

// Proposals queries proposals with optional filters.
func (m *Module) Proposals(ctx context.Context, opts *ProposalQueryOpts) (*ProposalsResponse, error) {
	params := make(map[string]string)
	if opts != nil {
		if opts.Voter != "" {
			params["voter"] = opts.Voter
		}
		if opts.Status != "" {
			params["status"] = opts.Status
		}
	}

	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "proposals",
		Params:   params,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query proposals: %w", err)
	}

	var result ProposalsResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse proposals: %w", err)
	}
	return &result, nil
}

// Proposal queries a specific proposal by ID.
func (m *Module) Proposal(ctx context.Context, proposalID string) (*Proposal, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "proposal",
		RawArgs:  []string{proposalID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query proposal: %w", err)
	}

	// Try direct parse first (API returns flat proposal object)
	var prop Proposal
	if err := json.Unmarshal(resp.Data, &prop); err == nil && prop.ProposalID != "" {
		return &prop, nil
	}

	// Try wrapped format as fallback
	var result struct {
		Proposal Proposal `json:"proposal"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse proposal: %w", err)
	}
	return &result.Proposal, nil
}

// Votes queries votes on a proposal.
func (m *Module) Votes(ctx context.Context, proposalID string) ([]Vote, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "votes",
		RawArgs:  []string{proposalID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query votes: %w", err)
	}

	var result struct {
		Votes []Vote `json:"votes"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse votes: %w", err)
	}
	return result.Votes, nil
}

// Vote queries a specific vote.
func (m *Module) Vote(ctx context.Context, proposalID, voter string) (*Vote, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "vote",
		RawArgs:  []string{proposalID, voter},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query vote: %w", err)
	}

	var result struct {
		Vote Vote `json:"vote"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse vote: %w", err)
	}
	return &result.Vote, nil
}

// Voters queries voters of a proposal.
func (m *Module) Voters(ctx context.Context, proposalID string) ([]string, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "voters",
		RawArgs:  []string{proposalID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query voters: %w", err)
	}

	var result struct {
		Voters []string `json:"voters"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse voters: %w", err)
	}
	return result.Voters, nil
}

// Councilors queries all councilors.
func (m *Module) Councilors(ctx context.Context) ([]Councilor, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "councilors",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query councilors: %w", err)
	}

	var result struct {
		Councilors []Councilor `json:"councilors"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse councilors: %w", err)
	}
	return result.Councilors, nil
}

// Roles queries roles assigned to an address.
func (m *Module) Roles(ctx context.Context, address string) ([]string, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "roles",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}

	var result struct {
		RoleIds []string `json:"roleIds"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse roles: %w", err)
	}
	return result.RoleIds, nil
}

// AllRoles queries all registered roles.
func (m *Module) AllRoles(ctx context.Context) ([]Role, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "all-roles",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query all roles: %w", err)
	}

	var result struct {
		Roles []Role `json:"roles"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse roles: %w", err)
	}
	return result.Roles, nil
}

// Role queries a role by ID or SID.
func (m *Module) Role(ctx context.Context, identifier string) (*Role, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "role",
		RawArgs:  []string{identifier},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	// Try direct unmarshal first (sekaid returns unwrapped response)
	var role Role
	if err := json.Unmarshal(resp.Data, &role); err == nil && role.Sid != "" {
		return &role, nil
	}

	// Try wrapped response
	var result struct {
		Role Role `json:"role"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse role: %w", err)
	}
	return &result.Role, nil
}

// Permissions queries permissions of an address.
func (m *Module) Permissions(ctx context.Context, address string) (*PermissionsResponse, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "permissions",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}

	var result PermissionsResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse permissions: %w", err)
	}
	return &result, nil
}

// ExecutionFee queries execution fee by transaction type.
func (m *Module) ExecutionFee(ctx context.Context, txType string) (*ExecutionFee, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "execution-fee",
		RawArgs:  []string{txType},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query execution fee: %w", err)
	}

	var result struct {
		Fee ExecutionFee `json:"fee"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse execution fee: %w", err)
	}
	return &result.Fee, nil
}

// AllExecutionFees queries all execution fees.
func (m *Module) AllExecutionFees(ctx context.Context) ([]ExecutionFee, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "all-execution-fees",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query execution fees: %w", err)
	}

	var result struct {
		Fees []ExecutionFee `json:"fees"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse execution fees: %w", err)
	}
	return result.Fees, nil
}

// IdentityRecord queries identity record by ID.
func (m *Module) IdentityRecord(ctx context.Context, id string) (*IdentityRecord, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "identity-record",
		RawArgs:  []string{id},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query identity record: %w", err)
	}

	var result struct {
		Record IdentityRecord `json:"record"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse identity record: %w", err)
	}
	return &result.Record, nil
}

// IdentityRecords queries all identity records.
func (m *Module) IdentityRecords(ctx context.Context) ([]IdentityRecord, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "identity-records",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query identity records: %w", err)
	}

	var result struct {
		Records []IdentityRecord `json:"records"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse identity records: %w", err)
	}
	return result.Records, nil
}

// IdentityRecordsByAddress queries identity records by address.
func (m *Module) IdentityRecordsByAddress(ctx context.Context, address string) ([]IdentityRecord, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "identity-records-by-addr",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query identity records: %w", err)
	}

	var result struct {
		Records []IdentityRecord `json:"records"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse identity records: %w", err)
	}
	return result.Records, nil
}

// DataRegistry queries data registry by key.
func (m *Module) DataRegistry(ctx context.Context, key string) (*DataRegistryEntry, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "data-registry",
		RawArgs:  []string{key},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query data registry: %w", err)
	}

	var result DataRegistryEntry
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse data registry: %w", err)
	}
	return &result, nil
}

// DataRegistryKeys queries all data registry keys.
func (m *Module) DataRegistryKeys(ctx context.Context) ([]string, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "data-registry-keys",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query data registry keys: %w", err)
	}

	var result struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse data registry keys: %w", err)
	}
	return result.Keys, nil
}

// Polls queries polls by address.
func (m *Module) Polls(ctx context.Context, address string) ([]Poll, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "polls",
		RawArgs:  []string{address},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query polls: %w", err)
	}

	var result struct {
		Polls []Poll `json:"polls"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse polls: %w", err)
	}
	return result.Polls, nil
}

// PollVotes queries poll votes by ID.
func (m *Module) PollVotes(ctx context.Context, pollID string) ([]PollVote, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "poll-votes",
		RawArgs:  []string{pollID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query poll votes: %w", err)
	}

	var result struct {
		Votes []PollVote `json:"votes"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse poll votes: %w", err)
	}
	return result.Votes, nil
}

// PoorNetworkMessages queries poor network messages.
func (m *Module) PoorNetworkMessages(ctx context.Context) ([]string, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "poor-network-messages",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query poor network messages: %w", err)
	}

	var result struct {
		Messages []string `json:"messages"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse poor network messages: %w", err)
	}
	return result.Messages, nil
}

// CustomPrefixes queries custom prefixes.
func (m *Module) CustomPrefixes(ctx context.Context) (*CustomPrefixes, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "custom-prefixes",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query custom prefixes: %w", err)
	}

	var result CustomPrefixes
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse custom prefixes: %w", err)
	}
	return &result, nil
}

// AllProposalDurations queries all proposal durations.
func (m *Module) AllProposalDurations(ctx context.Context) (map[string]string, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "all-proposal-durations",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query proposal durations: %w", err)
	}

	var result struct {
		ProposalDurations map[string]string `json:"proposal_durations"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse proposal durations: %w", err)
	}
	return result.ProposalDurations, nil
}

// ProposalDuration queries a specific proposal duration.
func (m *Module) ProposalDuration(ctx context.Context, proposalType string) (string, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "proposal-duration",
		RawArgs:  []string{proposalType},
	})
	if err != nil {
		return "", fmt.Errorf("failed to query proposal duration: %w", err)
	}

	var result struct {
		Duration string `json:"duration"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return "", fmt.Errorf("failed to parse proposal duration: %w", err)
	}
	return result.Duration, nil
}

// NonCouncilors queries all governance members that are not councilors.
func (m *Module) NonCouncilors(ctx context.Context) ([]GovernanceMember, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "non-councilors",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query non-councilors: %w", err)
	}

	var result struct {
		NonCouncilors []GovernanceMember `json:"non_councilors"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse non-councilors: %w", err)
	}
	return result.NonCouncilors, nil
}

// CouncilRegistry queries the governance council registry.
func (m *Module) CouncilRegistry(ctx context.Context) (*GovernanceMember, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "council-registry",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query council registry: %w", err)
	}

	var result GovernanceMember
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse council registry: %w", err)
	}
	return &result, nil
}

// WhitelistedPermissionAddresses queries addresses with a specific whitelisted permission.
func (m *Module) WhitelistedPermissionAddresses(ctx context.Context, permission string) ([]string, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "whitelisted-permission-addresses",
		RawArgs:  []string{permission},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query whitelisted permission addresses: %w", err)
	}

	var result struct {
		Addresses []string `json:"addresses"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse whitelisted permission addresses: %w", err)
	}
	return result.Addresses, nil
}

// BlacklistedPermissionAddresses queries addresses with a specific blacklisted permission.
func (m *Module) BlacklistedPermissionAddresses(ctx context.Context, permission string) ([]string, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "blacklisted-permission-addresses",
		RawArgs:  []string{permission},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query blacklisted permission addresses: %w", err)
	}

	var result struct {
		Addresses []string `json:"addresses"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse blacklisted permission addresses: %w", err)
	}
	return result.Addresses, nil
}

// WhitelistedRoleAddresses queries addresses with a specific whitelisted role.
func (m *Module) WhitelistedRoleAddresses(ctx context.Context, role string) ([]string, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "whitelisted-role-addresses",
		RawArgs:  []string{role},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query whitelisted role addresses: %w", err)
	}

	var result struct {
		Addresses []string `json:"addresses"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse whitelisted role addresses: %w", err)
	}
	return result.Addresses, nil
}

// ProposerVotersCount queries the count of proposers and voters.
func (m *Module) ProposerVotersCount(ctx context.Context) (*ProposerVotersCount, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "proposer-voters-count",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query proposer voters count: %w", err)
	}

	var result ProposerVotersCount
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse proposer voters count: %w", err)
	}
	return &result, nil
}

// AllIdentityRecordVerifyRequests queries all identity record verify requests.
func (m *Module) AllIdentityRecordVerifyRequests(ctx context.Context) ([]IdentityRecordVerifyRequest, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "all-identity-record-verify-requests",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query identity record verify requests: %w", err)
	}

	var result struct {
		Requests []IdentityRecordVerifyRequest `json:"verify_requests"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse identity record verify requests: %w", err)
	}
	return result.Requests, nil
}

// IdentityRecordVerifyRequest queries identity record verify request by ID.
func (m *Module) IdentityRecordVerifyRequest(ctx context.Context, id string) (*IdentityRecordVerifyRequest, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "identity-record-verify-request",
		RawArgs:  []string{id},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query identity record verify request: %w", err)
	}

	var result struct {
		Request IdentityRecordVerifyRequest `json:"verify_request"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse identity record verify request: %w", err)
	}
	return &result.Request, nil
}

// IdentityRecordVerifyRequestsByApprover queries identity record verify requests by approver.
func (m *Module) IdentityRecordVerifyRequestsByApprover(ctx context.Context, approver string) ([]IdentityRecordVerifyRequest, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "identity-record-verify-requests-by-approver",
		RawArgs:  []string{approver},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query identity record verify requests by approver: %w", err)
	}

	var result struct {
		Requests []IdentityRecordVerifyRequest `json:"verify_requests"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse identity record verify requests: %w", err)
	}
	return result.Requests, nil
}

// IdentityRecordVerifyRequestsByRequester queries identity record verify requests by requester.
func (m *Module) IdentityRecordVerifyRequestsByRequester(ctx context.Context, requester string) ([]IdentityRecordVerifyRequest, error) {
	resp, err := m.client.Query(ctx, &sdk.QueryRequest{
		Module:   "customgov",
		Endpoint: "identity-record-verify-requests-by-requester",
		RawArgs:  []string{requester},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query identity record verify requests by requester: %w", err)
	}

	var result struct {
		Requests []IdentityRecordVerifyRequest `json:"verify_requests"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse identity record verify requests: %w", err)
	}
	return result.Requests, nil
}

// TxOptions contains common transaction options.
type TxOptions struct {
	Fees          string
	Gas           string
	GasAdjustment float64
	Memo          string
	BroadcastMode string
}

// VoteProposal votes on a proposal.
func (m *Module) VoteProposal(ctx context.Context, from, proposalID string, voteOption int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal vote",
		Args:             []string{proposalID, fmt.Sprintf("%d", voteOption)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to vote on proposal: %w", err)
	}
	return resp, nil
}

func buildTxFlags(opts *TxOptions) map[string]string {
	flags := make(map[string]string)
	if opts != nil {
		if opts.Fees != "" {
			flags["fees"] = opts.Fees
		}
		if opts.Gas != "" {
			flags["gas"] = opts.Gas
		}
		if opts.GasAdjustment > 0 {
			flags["gas-adjustment"] = fmt.Sprintf("%.2f", opts.GasAdjustment)
		}
		if opts.Memo != "" {
			flags["note"] = opts.Memo
		}
		if opts.BroadcastMode != "" {
			flags["broadcast-mode"] = opts.BroadcastMode
		}
	}
	return flags
}

// CouncilorClaimSeat claims a councilor seat.
func (m *Module) CouncilorClaimSeat(ctx context.Context, from, moniker string, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)
	if moniker != "" {
		flags["moniker"] = moniker
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "councilor claim-seat",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to claim councilor seat: %w", err)
	}
	return resp, nil
}

// CouncilorActivate activates a councilor.
func (m *Module) CouncilorActivate(ctx context.Context, from string, opts *TxOptions) (*sdk.TxResponse, error) {
	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "councilor activate",
		Args:             []string{},
		Signer:           from,
		Flags:            buildTxFlags(opts),
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to activate councilor: %w", err)
	}
	return resp, nil
}

// CouncilorPause pauses a councilor.
func (m *Module) CouncilorPause(ctx context.Context, from string, opts *TxOptions) (*sdk.TxResponse, error) {
	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "councilor pause",
		Args:             []string{},
		Signer:           from,
		Flags:            buildTxFlags(opts),
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to pause councilor: %w", err)
	}
	return resp, nil
}

// CouncilorUnpause unpauses a councilor.
func (m *Module) CouncilorUnpause(ctx context.Context, from string, opts *TxOptions) (*sdk.TxResponse, error) {
	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "councilor unpause",
		Args:             []string{},
		Signer:           from,
		Flags:            buildTxFlags(opts),
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unpause councilor: %w", err)
	}
	return resp, nil
}

// PermissionWhitelist whitelists a permission for an address.
func (m *Module) PermissionWhitelist(ctx context.Context, from, address string, permission int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)
	flags["addr"] = address
	flags["permission"] = fmt.Sprintf("%d", permission)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "permission whitelist",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to whitelist permission: %w", err)
	}
	return resp, nil
}

// PermissionBlacklist blacklists a permission for an address.
func (m *Module) PermissionBlacklist(ctx context.Context, from, address string, permission int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)
	flags["addr"] = address
	flags["permission"] = fmt.Sprintf("%d", permission)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "permission blacklist",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to blacklist permission: %w", err)
	}
	return resp, nil
}

// PermissionRemoveWhitelisted removes a whitelisted permission from an address.
func (m *Module) PermissionRemoveWhitelisted(ctx context.Context, from, address string, permission int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)
	flags["addr"] = address
	flags["permission"] = fmt.Sprintf("%d", permission)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "permission remove-whitelisted",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove whitelisted permission: %w", err)
	}
	return resp, nil
}

// PermissionRemoveBlacklisted removes a blacklisted permission from an address.
func (m *Module) PermissionRemoveBlacklisted(ctx context.Context, from, address string, permission int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)
	flags["addr"] = address
	flags["permission"] = fmt.Sprintf("%d", permission)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "permission remove-blacklisted",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove blacklisted permission: %w", err)
	}
	return resp, nil
}

// RoleCreate creates a new role.
// Usage: sekaid tx customgov role create [role_sid] [role_description] [flags]
func (m *Module) RoleCreate(ctx context.Context, from, sid, description string, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "role create",
		Args:             []string{sid, description},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	return resp, nil
}

// RoleAssign assigns a role to an account.
// Usage: sekaid tx customgov role assign [role_id] --addr <address> [flags]
func (m *Module) RoleAssign(ctx context.Context, from, address string, roleID int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)
	flags["addr"] = address

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "role assign",
		Args:             []string{fmt.Sprintf("%d", roleID)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}
	return resp, nil
}

// RoleUnassign unassigns a role from an account.
// Usage: sekaid tx customgov role unassign [role_id] --addr <address> [flags]
func (m *Module) RoleUnassign(ctx context.Context, from, address string, roleID int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)
	flags["addr"] = address

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "role unassign",
		Args:             []string{fmt.Sprintf("%d", roleID)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unassign role: %w", err)
	}
	return resp, nil
}

// RoleWhitelistPermission whitelists a permission for a role.
func (m *Module) RoleWhitelistPermission(ctx context.Context, from, roleSID string, permission int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "role whitelist-permission",
		Args:             []string{roleSID, fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to whitelist permission for role: %w", err)
	}
	return resp, nil
}

// RoleBlacklistPermission blacklists a permission for a role.
func (m *Module) RoleBlacklistPermission(ctx context.Context, from, roleSID string, permission int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "role blacklist-permission",
		Args:             []string{roleSID, fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to blacklist permission for role: %w", err)
	}
	return resp, nil
}

// RoleRemoveWhitelistedPermission removes a whitelisted permission from a role.
func (m *Module) RoleRemoveWhitelistedPermission(ctx context.Context, from, roleSID string, permission int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "role remove-whitelisted-permission",
		Args:             []string{roleSID, fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove whitelisted permission from role: %w", err)
	}
	return resp, nil
}

// RoleRemoveBlacklistedPermission removes a blacklisted permission from a role.
func (m *Module) RoleRemoveBlacklistedPermission(ctx context.Context, from, roleSID string, permission int, opts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(opts)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "role remove-blacklisted-permission",
		Args:             []string{roleSID, fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove blacklisted permission from role: %w", err)
	}
	return resp, nil
}

// PollCreateOpts contains options for creating a poll.
// Required: Title, Roles, PollType
type PollCreateOpts struct {
	Title          string // Required: --title
	Description    string // Optional: --description
	Reference      string // Optional: --poll-reference
	Checksum       string // Optional: --poll-checksum
	Roles          string // Required: --poll-roles (comma-separated list)
	PollType       string // Required: --poll-type
	Options        string // Optional: --poll-options (comma-separated list)
	SelectionCount int    // Optional: --poll-choices (default 1)
	Duration       string // Optional: --poll-duration
}

// PollCreate creates a new poll.
// Usage: sekaid tx customgov poll create [flags]
// Flags: --title, --description, --poll-reference, --poll-checksum, --poll-roles,
// --poll-type, --poll-options, --poll-choices, --poll-duration
func (m *Module) PollCreate(ctx context.Context, from string, pollOpts *PollCreateOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if pollOpts != nil {
		if pollOpts.Title != "" {
			flags["title"] = pollOpts.Title
		}
		if pollOpts.Description != "" {
			flags["description"] = pollOpts.Description
		}
		if pollOpts.Reference != "" {
			flags["poll-reference"] = pollOpts.Reference
		}
		if pollOpts.Checksum != "" {
			flags["poll-checksum"] = pollOpts.Checksum
		}
		if pollOpts.Roles != "" {
			flags["poll-roles"] = pollOpts.Roles
		}
		if pollOpts.PollType != "" {
			flags["poll-type"] = pollOpts.PollType
		}
		if pollOpts.Options != "" {
			flags["poll-options"] = pollOpts.Options
		}
		if pollOpts.SelectionCount > 0 {
			flags["poll-choices"] = fmt.Sprintf("%d", pollOpts.SelectionCount)
		}
		if pollOpts.Duration != "" {
			flags["poll-duration"] = pollOpts.Duration
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "poll create",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create poll: %w", err)
	}
	return resp, nil
}

// PollVote votes on a poll.
func (m *Module) PollVote(ctx context.Context, from, pollID, options string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["poll-id"] = pollID
	flags["options"] = options

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "poll vote",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to vote on poll: %w", err)
	}
	return resp, nil
}

// SetNetworkProperties sets network properties directly (requires sudo permission).
// Usage: sekaid tx customgov set-network-properties [flags]
// Supported properties (passed as flags with underscores):
//   - max_tx_fee
//   - min_custody_reward
//   - min_tx_fee
//   - min_validators
//
// For other properties, use ProposalSetNetworkProperty to create a governance proposal.
func (m *Module) SetNetworkProperties(ctx context.Context, from string, properties map[string]string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	for k, v := range properties {
		flags[k] = v
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "set-network-properties",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set network properties: %w", err)
	}
	return resp, nil
}

// SetExecutionFee sets an execution fee.
func (m *Module) SetExecutionFee(ctx context.Context, from, txType string, executionFee, failureFee uint64, timeout uint64, defaultParams uint64, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["transaction_type"] = txType
	flags["execution_fee"] = fmt.Sprintf("%d", executionFee)
	flags["failure_fee"] = fmt.Sprintf("%d", failureFee)
	flags["timeout"] = fmt.Sprintf("%d", timeout)
	flags["default_parameters"] = fmt.Sprintf("%d", defaultParams)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "set-execution-fee",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set execution fee: %w", err)
	}
	return resp, nil
}

// RegisterIdentityRecords registers identity records.
func (m *Module) RegisterIdentityRecords(ctx context.Context, from string, infosJSON string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["infos-json"] = infosJSON

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "register-identity-records",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register identity records: %w", err)
	}
	return resp, nil
}

// DeleteIdentityRecords deletes identity records.
func (m *Module) DeleteIdentityRecords(ctx context.Context, from string, keys string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["keys"] = keys

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "delete-identity-records",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete identity records: %w", err)
	}
	return resp, nil
}

// RequestIdentityRecordVerify requests identity record verification.
func (m *Module) RequestIdentityRecordVerify(ctx context.Context, from, verifier, recordIDs, verifierTip string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["verifier"] = verifier
	flags["record-ids"] = recordIDs
	if verifierTip != "" {
		flags["verifier-tip"] = verifierTip
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "request-identity-record-verify",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to request identity record verification: %w", err)
	}
	return resp, nil
}

// HandleIdentityRecordsVerifyRequest handles (approves/rejects) identity records verify request.
func (m *Module) HandleIdentityRecordsVerifyRequest(ctx context.Context, from, requestID string, approve bool, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	boolFlags := make(map[string]bool)
	if approve {
		boolFlags["approve"] = true
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "handle-identity-records-verify-request",
		Args:             []string{requestID},
		Signer:           from,
		Flags:            flags,
		BoolFlags:        boolFlags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to handle identity records verify request: %w", err)
	}
	return resp, nil
}

// CancelIdentityRecordsVerifyRequest cancels identity records verify request.
func (m *Module) CancelIdentityRecordsVerifyRequest(ctx context.Context, from, requestID string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "cancel-identity-records-verify-request",
		Args:             []string{requestID},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to cancel identity records verify request: %w", err)
	}
	return resp, nil
}

// ProposalAssignRole creates a proposal to assign a role.
// Usage: sekaid tx customgov proposal account assign-role [role_identifier] --addr <address> --title <title> [flags]
func (m *Module) ProposalAssignRole(ctx context.Context, from, address, role, title, description string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["addr"] = address
	if title != "" {
		flags["title"] = title
	}
	if description != "" {
		flags["description"] = description
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal account assign-role",
		Args:             []string{role},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal assign role: %w", err)
	}
	return resp, nil
}

// ProposalUnassignRole creates a proposal to unassign a role.
// Usage: sekaid tx customgov proposal account unassign-role [role] --addr <address> --title <title> [flags]
func (m *Module) ProposalUnassignRole(ctx context.Context, from, address, role, title, description string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["addr"] = address
	if title != "" {
		flags["title"] = title
	}
	if description != "" {
		flags["description"] = description
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal account unassign-role",
		Args:             []string{role},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal unassign role: %w", err)
	}
	return resp, nil
}

// ProposalWhitelistPermission creates a proposal to whitelist a permission.
func (m *Module) ProposalWhitelistPermission(ctx context.Context, from, address string, permission int, description string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["addr"] = address
	flags["permission"] = fmt.Sprintf("%d", permission)
	if description != "" {
		flags["description"] = description
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal account whitelist-permission",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal whitelist permission: %w", err)
	}
	return resp, nil
}

// ProposalBlacklistPermission creates a proposal to blacklist a permission.
func (m *Module) ProposalBlacklistPermission(ctx context.Context, from, address string, permission int, description string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["addr"] = address
	flags["permission"] = fmt.Sprintf("%d", permission)
	if description != "" {
		flags["description"] = description
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal account blacklist-permission",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal blacklist permission: %w", err)
	}
	return resp, nil
}

// ProposalRemoveWhitelistedPermission creates a proposal to remove a whitelisted permission.
func (m *Module) ProposalRemoveWhitelistedPermission(ctx context.Context, from, address string, permission int, description string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["addr"] = address
	flags["permission"] = fmt.Sprintf("%d", permission)
	if description != "" {
		flags["description"] = description
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal account remove-whitelisted-permission",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal remove whitelisted permission: %w", err)
	}
	return resp, nil
}

// ProposalRemoveBlacklistedPermission creates a proposal to remove a blacklisted permission.
func (m *Module) ProposalRemoveBlacklistedPermission(ctx context.Context, from, address string, permission int, description string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["addr"] = address
	flags["permission"] = fmt.Sprintf("%d", permission)
	if description != "" {
		flags["description"] = description
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal account remove-blacklisted-permission",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal remove blacklisted permission: %w", err)
	}
	return resp, nil
}

// ProposalCreateRoleOpts contains options for creating a role proposal.
type ProposalCreateRoleOpts struct {
	Title       string
	Description string
	Whitelist   string
	Blacklist   string
}

// ProposalCreateRole creates a proposal to create a new role.
func (m *Module) ProposalCreateRole(ctx context.Context, from, roleSID, roleDescription string, propOpts *ProposalCreateRoleOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.Whitelist != "" {
			flags["whitelist"] = propOpts.Whitelist
		}
		if propOpts.Blacklist != "" {
			flags["blacklist"] = propOpts.Blacklist
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal role create",
		Args:             []string{roleSID, roleDescription},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal create role: %w", err)
	}
	return resp, nil
}

// ProposalRoleOpts contains common options for role proposals.
type ProposalRoleOpts struct {
	Title       string
	Description string
}

// ProposalRemoveRole creates a proposal to remove a role.
func (m *Module) ProposalRemoveRole(ctx context.Context, from, roleSID string, propOpts *ProposalRoleOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal role remove",
		Args:             []string{roleSID},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal remove role: %w", err)
	}
	return resp, nil
}

// ProposalWhitelistRolePermission creates a proposal to whitelist a permission for a role.
func (m *Module) ProposalWhitelistRolePermission(ctx context.Context, from, roleSID string, permission int, propOpts *ProposalRoleOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal role whitelist-permission",
		Args:             []string{roleSID, fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal whitelist role permission: %w", err)
	}
	return resp, nil
}

// ProposalBlacklistRolePermission creates a proposal to blacklist a permission for a role.
func (m *Module) ProposalBlacklistRolePermission(ctx context.Context, from, roleSID string, permission int, propOpts *ProposalRoleOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal role blacklist-permission",
		Args:             []string{roleSID, fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal blacklist role permission: %w", err)
	}
	return resp, nil
}

// ProposalRemoveWhitelistedRolePermission creates a proposal to remove a whitelisted permission from a role.
func (m *Module) ProposalRemoveWhitelistedRolePermission(ctx context.Context, from, roleSID string, permission int, propOpts *ProposalRoleOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal role remove-whitelisted-permission",
		Args:             []string{roleSID, fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal remove whitelisted role permission: %w", err)
	}
	return resp, nil
}

// ProposalRemoveBlacklistedRolePermission creates a proposal to remove a blacklisted permission from a role.
func (m *Module) ProposalRemoveBlacklistedRolePermission(ctx context.Context, from, roleSID string, permission int, propOpts *ProposalRoleOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal role remove-blacklisted-permission",
		Args:             []string{roleSID, fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal remove blacklisted role permission: %w", err)
	}
	return resp, nil
}

// ProposalSetNetworkProperty creates a proposal to set a network property.
// Usage: sekaid tx customgov proposal set-network-property <property> <value> [flags]
// Required flags: --title
// Optional flags: --description
// Available properties: MIN_TX_FEE, MAX_TX_FEE, VOTE_QUORUM, MINIMUM_PROPOSAL_END_TIME,
// PROPOSAL_ENACTMENT_TIME, MIN_PROPOSAL_END_BLOCKS, MIN_PROPOSAL_ENACTMENT_BLOCKS,
// ENABLE_FOREIGN_FEE_PAYMENTS, etc.
func (m *Module) ProposalSetNetworkProperty(ctx context.Context, from, property, value, title, description string, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	flags["title"] = title
	if description != "" {
		flags["description"] = description
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal set-network-property",
		Args:             []string{property, value},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal set network property: %w", err)
	}
	return resp, nil
}

// ProposalOtherOpts contains common options for other proposal commands.
type ProposalOtherOpts struct {
	Title       string
	Description string
}

// ProposalSetPoorNetworkMsgs creates a proposal to set poor network messages.
func (m *Module) ProposalSetPoorNetworkMsgs(ctx context.Context, from, messages string, propOpts *ProposalOtherOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal set-poor-network-msgs",
		Args:             []string{messages},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal set poor network msgs: %w", err)
	}
	return resp, nil
}

// ProposalSetExecutionFeesOpts contains options for proposal set execution fees.
type ProposalSetExecutionFeesOpts struct {
	Title         string
	Description   string
	TxTypes       string
	ExecutionFees string
	FailureFees   string
	Timeouts      string
	DefaultParams string
}

// ProposalSetExecutionFees creates a proposal to set execution fees.
func (m *Module) ProposalSetExecutionFees(ctx context.Context, from string, propOpts *ProposalSetExecutionFeesOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.TxTypes != "" {
			flags["tx-types"] = propOpts.TxTypes
		}
		if propOpts.ExecutionFees != "" {
			flags["execution-fees"] = propOpts.ExecutionFees
		}
		if propOpts.FailureFees != "" {
			flags["failure-fees"] = propOpts.FailureFees
		}
		if propOpts.Timeouts != "" {
			flags["timeouts"] = propOpts.Timeouts
		}
		if propOpts.DefaultParams != "" {
			flags["default-params"] = propOpts.DefaultParams
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal proposal-set-execution-fees",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal set execution fees: %w", err)
	}
	return resp, nil
}

// ProposalUpsertDataRegistry creates a proposal to upsert a data registry entry.
func (m *Module) ProposalUpsertDataRegistry(ctx context.Context, from, key, hash, reference, encoding, size string, propOpts *ProposalOtherOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal upsert-data-registry",
		Args:             []string{key, hash, reference, encoding, size},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal upsert data registry: %w", err)
	}
	return resp, nil
}

// ProposalSetProposalDurations creates a proposal to set proposal durations.
func (m *Module) ProposalSetProposalDurations(ctx context.Context, from, proposalTypes, durations string, propOpts *ProposalOtherOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal set-proposal-durations-proposal",
		Args:             []string{proposalTypes, durations},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal set proposal durations: %w", err)
	}
	return resp, nil
}

// ProposalJailCouncilor creates a proposal to jail councilors.
func (m *Module) ProposalJailCouncilor(ctx context.Context, from, councilors string, propOpts *ProposalOtherOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal proposal-jail-councilor",
		Args:             []string{councilors},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal jail councilor: %w", err)
	}
	return resp, nil
}

// ProposalResetWholeCouncilorRank creates a proposal to reset whole councilor rank.
func (m *Module) ProposalResetWholeCouncilorRank(ctx context.Context, from string, propOpts *ProposalOtherOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal proposal-reset-whole-councilor-rank",
		Args:             []string{},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal reset whole councilor rank: %w", err)
	}
	return resp, nil
}

// ProposalAccountPermissionOpts contains options for account permission proposals.
type ProposalAccountPermissionOpts struct {
	Title       string
	Description string
	Addr        string
}

// ProposalWhitelistAccountPermission creates a proposal to whitelist a permission for an account.
func (m *Module) ProposalWhitelistAccountPermission(ctx context.Context, from string, permission int, propOpts *ProposalAccountPermissionOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.Addr != "" {
			flags["addr"] = propOpts.Addr
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal account whitelist-permission",
		Args:             []string{fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal whitelist account permission: %w", err)
	}
	return resp, nil
}

// ProposalBlacklistAccountPermission creates a proposal to blacklist a permission for an account.
func (m *Module) ProposalBlacklistAccountPermission(ctx context.Context, from string, permission int, propOpts *ProposalAccountPermissionOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.Addr != "" {
			flags["addr"] = propOpts.Addr
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal account blacklist-permission",
		Args:             []string{fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal blacklist account permission: %w", err)
	}
	return resp, nil
}

// ProposalRemoveWhitelistedAccountPermission creates a proposal to remove a whitelisted permission from an account.
func (m *Module) ProposalRemoveWhitelistedAccountPermission(ctx context.Context, from string, permission int, propOpts *ProposalAccountPermissionOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.Addr != "" {
			flags["addr"] = propOpts.Addr
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal account remove-whitelisted-permission",
		Args:             []string{fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal remove whitelisted account permission: %w", err)
	}
	return resp, nil
}

// ProposalRemoveBlacklistedAccountPermission creates a proposal to remove a blacklisted permission from an account.
func (m *Module) ProposalRemoveBlacklistedAccountPermission(ctx context.Context, from string, permission int, propOpts *ProposalAccountPermissionOpts, txOpts *TxOptions) (*sdk.TxResponse, error) {
	flags := buildTxFlags(txOpts)
	if propOpts != nil {
		if propOpts.Title != "" {
			flags["title"] = propOpts.Title
		}
		if propOpts.Description != "" {
			flags["description"] = propOpts.Description
		}
		if propOpts.Addr != "" {
			flags["addr"] = propOpts.Addr
		}
	}

	resp, err := m.client.Tx(ctx, &sdk.TxRequest{
		Module:           "customgov",
		Action:           "proposal account remove-blacklisted-permission",
		Args:             []string{fmt.Sprintf("%d", permission)},
		Signer:           from,
		Flags:            flags,
		SkipConfirmation: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal remove blacklisted account permission: %w", err)
	}
	return resp, nil
}
