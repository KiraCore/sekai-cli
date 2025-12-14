package gov

// NetworkProperties contains network configuration.
type NetworkProperties struct {
	MinTxFee                    string `json:"min_tx_fee"`
	MaxTxFee                    string `json:"max_tx_fee"`
	VoteQuorum                  string `json:"vote_quorum"`
	MinimumProposalEndTime      string `json:"minimum_proposal_end_time"`
	ProposalEnactmentTime       string `json:"proposal_enactment_time"`
	EnableForeignFeePayments    bool   `json:"enable_foreign_fee_payments"`
	MischanceRankDecreaseAmount string `json:"mischance_rank_decrease_amount"`
	MaxMischance                string `json:"max_mischance"`
	InactiveRankDecreasePercent string `json:"inactive_rank_decrease_percent"`
	MinValidators               string `json:"min_validators"`
	PoorNetworkMaxBankSend      string `json:"poor_network_max_bank_send"`
	UnjailMaxTime               string `json:"unjail_max_time"`
	MinIdentityApprovalTip      string `json:"min_identity_approval_tip"`
	UniqueIdentityKeys          string `json:"unique_identity_keys"`
	UbiHardcap                  string `json:"ubi_hardcap"`
	ValidatorsFeeShare          string `json:"validators_fee_share"`
	InflationRate               string `json:"inflation_rate"`
	InflationPeriod             string `json:"inflation_period"`
	UnstakingPeriod             string `json:"unstaking_period"`
	MaxDelegators               string `json:"max_delegators"`
	MinDelegationPushout        string `json:"min_delegation_pushout"`
	SlashingPeriod              string `json:"slashing_period"`
	MaxJailedPercentage         string `json:"max_jailed_percentage"`
	MaxSlashingPercentage       string `json:"max_slashing_percentage"`
}

// ProposalQueryOpts contains options for querying proposals.
type ProposalQueryOpts struct {
	Voter  string
	Status string
}

// ProposalsResponse contains proposals query response.
type ProposalsResponse struct {
	Proposals []Proposal `json:"proposals"`
}

// Proposal represents a governance proposal.
type Proposal struct {
	ProposalID       string `json:"proposal_id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Content          any    `json:"content,omitempty"`
	SubmitTime       string `json:"submit_time"`
	VotingEndTime    string `json:"voting_end_time"`
	EnactmentEndTime string `json:"enactment_end_time"`
	Status           string `json:"status"`
	Result           string `json:"result"`
	ExecResult       string `json:"exec_result,omitempty"`
	MinVotingEndTime string `json:"min_voting_end_time,omitempty"`
}

// Vote represents a vote on a proposal.
type Vote struct {
	ProposalID string `json:"proposal_id"`
	Voter      string `json:"voter"`
	Option     string `json:"option"`
}

// Councilor represents a councilor.
type Councilor struct {
	Address string `json:"address"`
	Moniker string `json:"moniker,omitempty"`
	Status  string `json:"status"`
	Rank    string `json:"rank,omitempty"`
}

// Role represents a governance role.
type Role struct {
	ID          uint64           `json:"id"`
	Sid         string           `json:"sid"`
	Description string           `json:"description,omitempty"`
	Permissions *RolePermissions `json:"permissions,omitempty"`
}

// RolePermissions contains whitelisted and blacklisted permissions for a role.
type RolePermissions struct {
	Whitelist []uint64 `json:"whitelist,omitempty"`
	Blacklist []uint64 `json:"blacklist,omitempty"`
}

// PermissionsResponse contains permissions query response.
type PermissionsResponse struct {
	Whitelist []uint64 `json:"whitelist,omitempty"`
	Blacklist []uint64 `json:"blacklist,omitempty"`
}

// ExecutionFee represents an execution fee configuration.
type ExecutionFee struct {
	TransactionType   string `json:"transaction_type"`
	ExecutionFee      string `json:"execution_fee"`
	FailureFee        string `json:"failure_fee"`
	Timeout           string `json:"timeout"`
	DefaultParameters string `json:"default_parameters,omitempty"`
}

// IdentityRecord represents an identity record.
type IdentityRecord struct {
	ID        string            `json:"id"`
	Address   string            `json:"address"`
	Key       string            `json:"key"`
	Value     string            `json:"value"`
	Date      string            `json:"date"`
	Verifiers []string          `json:"verifiers,omitempty"`
	Infos     map[string]string `json:"infos,omitempty"`
}

// DataRegistryEntry represents a data registry entry.
type DataRegistryEntry struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Reference string `json:"reference,omitempty"`
	Encoding  string `json:"encoding,omitempty"`
	Size      string `json:"size,omitempty"`
}

// Poll represents a governance poll.
type Poll struct {
	ID            string   `json:"id"`
	Creator       string   `json:"creator"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Reference     string   `json:"reference,omitempty"`
	Checksum      string   `json:"checksum,omitempty"`
	Options       []string `json:"options"`
	VotingEndTime string   `json:"voting_end_time"`
	Roles         []uint64 `json:"roles,omitempty"`
}

// PollVote represents a vote on a poll.
type PollVote struct {
	PollID string `json:"poll_id"`
	Voter  string `json:"voter"`
	Option string `json:"option"`
}

// CustomPrefixes contains network custom prefixes.
type CustomPrefixes struct {
	DefaultDenom string `json:"default_denom"`
	Bech32Prefix string `json:"bech32_prefix"`
}

// ProposalDurations contains proposal duration by type.
type ProposalDurations struct {
	Durations map[string]string `json:"proposal_durations"`
}

// GovernanceMember represents a governance member (councilor or non-councilor).
type GovernanceMember struct {
	Address     string           `json:"address"`
	Roles       []string         `json:"roles,omitempty"`
	Status      string           `json:"status,omitempty"`
	Votes       []string         `json:"votes,omitempty"`
	Permissions *RolePermissions `json:"permissions,omitempty"`
	Skin        string           `json:"skin,omitempty"`
}

// ProposerVotersCount contains proposer and voter counts.
type ProposerVotersCount struct {
	Proposers string `json:"proposers"`
	Voters    string `json:"voters"`
}

// IdentityRecordVerifyRequest represents an identity verification request.
type IdentityRecordVerifyRequest struct {
	ID        string   `json:"id"`
	Address   string   `json:"address"`
	Verifier  string   `json:"verifier"`
	RecordIds []string `json:"record_ids,omitempty"`
	Tip       string   `json:"tip,omitempty"`
}
