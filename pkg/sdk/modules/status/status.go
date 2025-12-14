// Package status provides node status functionality.
package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Module provides node status functionality.
type Module struct {
	client sdk.Client
}

// New creates a new status module.
func New(client sdk.Client) *Module {
	return &Module{client: client}
}

// Status returns the node status.
func (m *Module) Status(ctx context.Context) (*sdk.StatusResponse, error) {
	return m.client.Status(ctx)
}

// NodeInfo returns information about the node.
func (m *Module) NodeInfo(ctx context.Context) (*sdk.NodeInfo, error) {
	status, err := m.client.Status(ctx)
	if err != nil {
		return nil, err
	}
	return &status.NodeInfo, nil
}

// SyncInfo returns the node's sync status.
func (m *Module) SyncInfo(ctx context.Context) (*sdk.SyncInfo, error) {
	status, err := m.client.Status(ctx)
	if err != nil {
		return nil, err
	}
	return &status.SyncInfo, nil
}

// ValidatorInfo returns validator information (if node is a validator).
func (m *Module) ValidatorInfo(ctx context.Context) (*sdk.ValidatorInfo, error) {
	status, err := m.client.Status(ctx)
	if err != nil {
		return nil, err
	}
	return &status.ValidatorInfo, nil
}

// ChainID returns the chain ID.
func (m *Module) ChainID(ctx context.Context) (string, error) {
	status, err := m.client.Status(ctx)
	if err != nil {
		return "", err
	}
	return status.NodeInfo.Network, nil
}

// LatestBlockHeight returns the latest block height.
func (m *Module) LatestBlockHeight(ctx context.Context) (int64, error) {
	status, err := m.client.Status(ctx)
	if err != nil {
		return 0, err
	}
	return status.SyncInfo.LatestBlockHeight, nil
}

// IsSyncing returns true if the node is still syncing.
func (m *Module) IsSyncing(ctx context.Context) (bool, error) {
	status, err := m.client.Status(ctx)
	if err != nil {
		return false, err
	}
	return status.SyncInfo.CatchingUp, nil
}

// NetworkProperties queries the network properties.
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
		// Try direct unmarshal
		var props NetworkProperties
		if err := json.Unmarshal(resp.Data, &props); err != nil {
			return nil, fmt.Errorf("failed to parse network properties: %w", err)
		}
		return &props, nil
	}

	return &result.Properties, nil
}

// NetworkProperties contains network configuration.
type NetworkProperties struct {
	MinTxFee                        string `json:"min_tx_fee"`
	MaxTxFee                        string `json:"max_tx_fee"`
	VoteQuorum                      string `json:"vote_quorum"`
	MinimumProposalEndTime          string `json:"minimum_proposal_end_time"`
	ProposalEnactmentTime           string `json:"proposal_enactment_time"`
	MinProposalEndBlocks            string `json:"min_proposal_end_blocks"`
	MinProposalEnactmentBlocks      string `json:"min_proposal_enactment_blocks"`
	EnableForeignFeePayments        bool   `json:"enable_foreign_fee_payments"`
	MischanceRankDecreaseAmount     string `json:"mischance_rank_decrease_amount"`
	MaxMischance                    string `json:"max_mischance"`
	MischanceConfidence             string `json:"mischance_confidence"`
	InactiveRankDecreasePercent     string `json:"inactive_rank_decrease_percent"`
	MinValidators                   string `json:"min_validators"`
	PoorNetworkMaxBankSend          string `json:"poor_network_max_bank_send"`
	UnjailMaxTime                   string `json:"unjail_max_time"`
	EnableTokenWhitelist            bool   `json:"enable_token_whitelist"`
	EnableTokenBlacklist            bool   `json:"enable_token_blacklist"`
	MinIdentityApprovalTip          string `json:"min_identity_approval_tip"`
	UniqueIdentityKeys              string `json:"unique_identity_keys"`
	UbiHardcap                      string `json:"ubi_hardcap"`
	ValidatorsFeeShare              string `json:"validators_fee_share"`
	InflationRate                   string `json:"inflation_rate"`
	InflationPeriod                 string `json:"inflation_period"`
	UnstakingPeriod                 string `json:"unstaking_period"`
	MaxDelegators                   string `json:"max_delegators"`
	MinDelegationPushout            string `json:"min_delegation_pushout"`
	SlashingPeriod                  string `json:"slashing_period"`
	MaxJailedPercentage             string `json:"max_jailed_percentage"`
	MaxSlashingPercentage           string `json:"max_slashing_percentage"`
	MinCustodyReward                string `json:"min_custody_reward"`
	MaxCustodyBufferSize            string `json:"max_custody_buffer_size"`
	MaxCustodyTxSize                string `json:"max_custody_tx_size"`
	AbstentionRankDecreaseAmount    string `json:"abstention_rank_decrease_amount"`
	MaxAbstention                   string `json:"max_abstention"`
	MinCollectiveBond               string `json:"min_collective_bond"`
	MinCollectiveBondingTime        string `json:"min_collective_bonding_time"`
	MaxCollectiveOutputs            string `json:"max_collective_outputs"`
	MinCollectiveClaimPeriod        string `json:"min_collective_claim_period"`
	ValidatorRecoveryBond           string `json:"validator_recovery_bond"`
	MaxAnnualInflation              string `json:"max_annual_inflation"`
	MaxProposalTitleSize            string `json:"max_proposal_title_size"`
	MaxProposalDescriptionSize      string `json:"max_proposal_description_size"`
	MaxProposalPollOptionSize       string `json:"max_proposal_poll_option_size"`
	MaxProposalPollOptionCount      string `json:"max_proposal_poll_option_count"`
	MaxProposalReferenceSize        string `json:"max_proposal_reference_size"`
	MaxProposalChecksumSize         string `json:"max_proposal_checksum_size"`
	MinDappBond                     string `json:"min_dapp_bond"`
	MaxDappBond                     string `json:"max_dapp_bond"`
	DappBondDuration                string `json:"dapp_bond_duration"`
	DappVerifierBond                string `json:"dapp_verifier_bond"`
	DappAutoDenounceTime            string `json:"dapp_auto_denounce_time"`
	DappMischanceRankDecreaseAmount string `json:"dapp_mischance_rank_decrease_amount"`
	DappMaxMischance                string `json:"dapp_max_mischance"`
	DappInactiveRankDecreasePercent string `json:"dapp_inactive_rank_decrease_percent"`
	DappPoolSlippageDefault         string `json:"dapp_pool_slippage_default"`
	MintingFtFee                    string `json:"minting_ft_fee"`
	MintingNftFee                   string `json:"minting_nft_fee"`
	VetoThreshold                   string `json:"veto_threshold"`
	AutocompoundIntervalNumBlocks   string `json:"autocompound_interval_num_blocks"`
	DowntimeInactiveDuration        string `json:"downtime_inactive_duration"`
}
