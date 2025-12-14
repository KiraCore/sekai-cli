package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/kiracore/sekai-cli/pkg/sdk"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/auth"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/bank"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/basket"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/bridge"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/collectives"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/custody"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/distributor"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/ethereum"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/evidence"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/gov"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/keys"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/layer2"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/multistaking"
	paramsmod "github.com/kiracore/sekai-cli/pkg/sdk/modules/params"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/recovery"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/slashing"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/spending"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/staking"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/status"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/tokens"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/ubi"
	"github.com/kiracore/sekai-cli/pkg/sdk/modules/upgrade"
	"github.com/kiracore/sekai-cli/pkg/sdk/types"
)

// ActionMapper maps module+action combinations to SDK method calls.
type ActionMapper struct {
	client sdk.Client

	// Module instances (lazily initialized)
	authMod        *auth.Module
	bankMod        *bank.Module
	basketMod      *basket.Module
	bridgeMod      *bridge.Module
	collectivesMod *collectives.Module
	custodyMod     *custody.Module
	distributorMod *distributor.Module
	ethereumMod    *ethereum.Module
	evidenceMod    *evidence.Module
	govMod         *gov.Module
	keysMod        *keys.Module
	layer2Mod      *layer2.Module
	multistakeMod  *multistaking.Module
	paramsMod      *paramsmod.Module
	recoveryMod    *recovery.Module
	slashingMod    *slashing.Module
	spendingMod    *spending.Module
	stakingMod     *staking.Module
	statusMod      *status.Module
	tokensMod      *tokens.Module
	ubiMod         *ubi.Module
	upgradeMod     *upgrade.Module
}

// NewActionMapper creates a new action mapper.
func NewActionMapper(client sdk.Client) *ActionMapper {
	return &ActionMapper{
		client: client,
	}
}

// resolveAddress resolves a key name to an address if needed.
// If the input looks like a bech32 address (starts with "kira"), it's returned as-is.
// Otherwise, it tries to resolve it as a key name.
func (m *ActionMapper) resolveAddress(ctx context.Context, nameOrAddress string) (string, error) {
	// If it looks like an address, return as-is
	if strings.HasPrefix(nameOrAddress, "kira") {
		return nameOrAddress, nil
	}

	// Try to resolve as key name
	if m.keysMod == nil {
		m.keysMod = keys.New(m.client)
	}

	keyInfo, err := m.keysMod.Show(ctx, nameOrAddress)
	if err != nil {
		return "", fmt.Errorf("failed to resolve '%s' as key name: %w", nameOrAddress, err)
	}

	return keyInfo.Address, nil
}

// Execute runs a module action with the given parameters.
// Returns the output data, transaction response (if applicable), and any error.
func (m *ActionMapper) Execute(ctx context.Context, module, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	module = strings.ToLower(module)
	action = strings.ToLower(action)

	switch module {
	case "keys":
		return m.executeKeys(ctx, action, params)
	case "bank":
		return m.executeBank(ctx, action, params, txOpts)
	case "auth":
		return m.executeAuth(ctx, action, params)
	case "gov", "customgov":
		return m.executeGov(ctx, action, params, txOpts)
	case "staking", "customstaking":
		return m.executeStaking(ctx, action, params, txOpts)
	case "multistaking":
		return m.executeMultistaking(ctx, action, params, txOpts)
	case "tokens":
		return m.executeTokens(ctx, action, params, txOpts)
	case "status":
		return m.executeStatus(ctx, action, params)
	case "custody":
		return m.executeCustody(ctx, action, params)
	case "params":
		return m.executeParams(ctx, action, params)
	case "ethereum":
		return m.executeEthereum(ctx, action, params)
	case "evidence", "customevidence":
		return m.executeEvidence(ctx, action, params)
	case "distributor":
		return m.executeDistributor(ctx, action, params)
	case "layer2":
		return m.executeLayer2(ctx, action, params)
	case "recovery":
		return m.executeRecovery(ctx, action, params)
	case "slashing", "customslashing":
		return m.executeSlashing(ctx, action, params)
	case "bridge":
		return m.executeBridge(ctx, action, params, txOpts)
	case "ubi":
		return m.executeUbi(ctx, action, params, txOpts)
	case "upgrade":
		return m.executeUpgrade(ctx, action, params, txOpts)
	case "spending":
		return m.executeSpending(ctx, action, params, txOpts)
	case "collectives":
		return m.executeCollectives(ctx, action, params, txOpts)
	case "basket":
		return m.executeBasket(ctx, action, params, txOpts)
	default:
		// For modules without specific mappings, use generic execution
		return m.executeGeneric(ctx, module, action, params, txOpts)
	}
}

// executeKeys handles keys module actions.
func (m *ActionMapper) executeKeys(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.keysMod == nil {
		m.keysMod = keys.New(m.client)
	}

	switch action {
	case "list":
		result, err := m.keysMod.List(ctx)
		return result, nil, err

	case "show":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("keys.show requires 'name' parameter")
		}
		result, err := m.keysMod.Show(ctx, name)
		return result, nil, err

	case "add":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("keys.add requires 'name' parameter")
		}
		result, err := m.keysMod.Add(ctx, name, nil)
		return result, nil, err

	case "delete":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("keys.delete requires 'name' parameter")
		}
		err := m.keysMod.Delete(ctx, name, true)
		return nil, nil, err

	case "export":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("keys.export requires 'name' parameter")
		}
		result, err := m.keysMod.Export(ctx, name)
		return result, nil, err

	case "import":
		name := params["name"]
		armor := params["armor"]
		if name == "" || armor == "" {
			return nil, nil, fmt.Errorf("keys.import requires 'name' and 'armor' parameters")
		}
		err := m.keysMod.Import(ctx, name, armor)
		return nil, nil, err

	case "rename":
		oldName := params["old_name"]
		newName := params["new_name"]
		if oldName == "" || newName == "" {
			return nil, nil, fmt.Errorf("keys.rename requires 'old_name' and 'new_name' parameters")
		}
		err := m.keysMod.Rename(ctx, oldName, newName)
		return nil, nil, err

	case "mnemonic":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("keys.mnemonic requires 'name' parameter")
		}
		result, err := m.keysMod.Mnemonic(ctx, name)
		return result, nil, err

	case "get-address", "getaddress", "address":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("keys.get-address requires 'name' parameter")
		}
		result, err := m.keysMod.GetAddress(ctx, name)
		return result, nil, err

	case "exists":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("keys.exists requires 'name' parameter")
		}
		result, err := m.keysMod.Exists(ctx, name)
		return result, nil, err

	case "create":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("keys.create requires 'name' parameter")
		}
		opts := &keys.CreateOptions{
			HDPath:    params["hd_path"],
			Algorithm: params["algorithm"],
		}
		if params["recover"] == "true" {
			opts.Recover = true
			opts.Mnemonic = params["mnemonic"]
		}
		if params["no_backup"] == "true" {
			opts.NoBackup = true
		}
		result, err := m.keysMod.Create(ctx, name, opts)
		return result, nil, err

	case "recover":
		name := params["name"]
		mnemonic := params["mnemonic"]
		if name == "" || mnemonic == "" {
			return nil, nil, fmt.Errorf("keys.recover requires 'name' and 'mnemonic' parameters")
		}
		result, err := m.keysMod.Recover(ctx, name, mnemonic)
		return result, nil, err

	case "import-hex", "importhex":
		name := params["name"]
		hexKey := params["hex_key"]
		keyType := params["key_type"]
		if name == "" || hexKey == "" {
			return nil, nil, fmt.Errorf("keys.import-hex requires 'name' and 'hex_key' parameters")
		}
		err := m.keysMod.ImportHex(ctx, name, hexKey, keyType)
		return nil, nil, err

	case "list-key-types", "listkeytypes":
		result, err := m.keysMod.ListKeyTypes(ctx)
		return result, nil, err

	case "migrate":
		err := m.keysMod.Migrate(ctx)
		return nil, nil, err

	case "parse":
		address := params["address"]
		if address == "" {
			return nil, nil, fmt.Errorf("keys.parse requires 'address' parameter")
		}
		result, err := m.keysMod.Parse(ctx, address)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown keys action: %s", action)
	}
}

// executeBank handles bank module actions.
func (m *ActionMapper) executeBank(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.bankMod == nil {
		m.bankMod = bank.New(m.client)
	}

	switch action {
	case "balances", "balance":
		addressParam := params["address"]
		if addressParam == "" {
			return nil, nil, fmt.Errorf("bank.balances requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addressParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.bankMod.Balances(ctx, address)
		return result, nil, err

	case "total", "total-supply":
		result, err := m.bankMod.TotalSupply(ctx)
		return result, nil, err

	case "send":
		from := params["from"]
		toParam := params["to"]
		amount := params["amount"]

		if from == "" || toParam == "" || amount == "" {
			return nil, nil, fmt.Errorf("bank.send requires 'from', 'to', and 'amount' parameters")
		}

		// Resolve 'to' address if it's a key name
		to, err := m.resolveAddress(ctx, toParam)
		if err != nil {
			return nil, nil, err
		}

		coins, err := types.ParseCoins(amount)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid amount '%s': %w", amount, err)
		}

		opts := m.buildSendOptions(txOpts)
		resp, err := m.bankMod.Send(ctx, from, to, coins, opts)
		return resp, resp, err

	case "spendable-balances", "spendablebalances":
		addressParam := params["address"]
		if addressParam == "" {
			return nil, nil, fmt.Errorf("bank.spendable-balances requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addressParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.bankMod.SpendableBalances(ctx, address)
		return result, nil, err

	case "supply-of", "supplyof":
		denom := params["denom"]
		if denom == "" {
			return nil, nil, fmt.Errorf("bank.supply-of requires 'denom' parameter")
		}
		result, err := m.bankMod.SupplyOf(ctx, denom)
		return result, nil, err

	case "multi-send", "multisend":
		from := params["from"]
		toAddrs := params["to_addresses"]
		amount := params["amount"]

		if from == "" || toAddrs == "" || amount == "" {
			return nil, nil, fmt.Errorf("bank.multi-send requires 'from', 'to_addresses', and 'amount' parameters")
		}

		// Parse comma-separated addresses
		addresses := strings.Split(toAddrs, ",")
		resolvedAddrs := make([]string, len(addresses))
		for i, addr := range addresses {
			resolved, err := m.resolveAddress(ctx, strings.TrimSpace(addr))
			if err != nil {
				return nil, nil, err
			}
			resolvedAddrs[i] = resolved
		}

		coins, err := types.ParseCoins(amount)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid amount '%s': %w", amount, err)
		}

		multiOpts := &bank.MultiSendOptions{
			SendOptions: bank.SendOptions{
				Fees:          txOpts.Fees,
				Gas:           txOpts.Gas,
				Memo:          txOpts.Memo,
				BroadcastMode: txOpts.BroadcastMode,
			},
			Split: params["split"] == "true",
		}
		resp, err := m.bankMod.MultiSend(ctx, from, resolvedAddrs, coins, multiOpts)
		return resp, resp, err

	case "denom-metadata", "denommetadata":
		denom := params["denom"]
		result, err := m.bankMod.DenomMetadata(ctx, denom)
		return result, nil, err

	case "all-denoms-metadata", "alldenomsmetadata":
		result, err := m.bankMod.AllDenomsMetadata(ctx)
		return result, nil, err

	case "send-enabled", "sendenabled":
		denoms := params["denoms"]
		var denomList []string
		if denoms != "" {
			denomList = strings.Split(denoms, ",")
		}
		result, err := m.bankMod.SendEnabled(ctx, denomList...)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown bank action: %s", action)
	}
}

// executeAuth handles auth module actions.
func (m *ActionMapper) executeAuth(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.authMod == nil {
		m.authMod = auth.New(m.client)
	}

	switch action {
	case "account":
		address := params["address"]
		if address == "" {
			return nil, nil, fmt.Errorf("auth.account requires 'address' parameter")
		}
		result, err := m.authMod.Account(ctx, address)
		return result, nil, err

	case "accounts":
		result, err := m.authMod.Accounts(ctx, nil)
		return result, nil, err

	case "module-accounts":
		result, err := m.authMod.ModuleAccounts(ctx)
		return result, nil, err

	case "params":
		result, err := m.authMod.Params(ctx)
		return result, nil, err

	case "module-account", "moduleaccount":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("auth.module-account requires 'name' parameter")
		}
		result, err := m.authMod.ModuleAccount(ctx, name)
		return result, nil, err

	case "address-by-acc-num", "addressbyaccnum":
		accNum := params["acc_num"]
		if accNum == "" {
			accNum = params["account_number"]
		}
		if accNum == "" {
			return nil, nil, fmt.Errorf("auth.address-by-acc-num requires 'acc_num' parameter")
		}
		result, err := m.authMod.AddressByAccNum(ctx, accNum)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown auth action: %s", action)
	}
}

// executeGov handles governance module actions.
func (m *ActionMapper) executeGov(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.govMod == nil {
		m.govMod = gov.New(m.client)
	}

	switch action {
	// Query actions
	case "network-properties":
		result, err := m.govMod.NetworkProperties(ctx)
		return result, nil, err

	case "proposals":
		result, err := m.govMod.Proposals(ctx, nil)
		return result, nil, err

	case "proposal":
		id := params["id"]
		if id == "" {
			return nil, nil, fmt.Errorf("gov.proposal requires 'id' parameter")
		}
		result, err := m.govMod.Proposal(ctx, id)
		return result, nil, err

	case "all-roles":
		result, err := m.govMod.AllRoles(ctx)
		return result, nil, err

	case "role":
		id := params["id"]
		if id == "" {
			id = params["sid"]
		}
		if id == "" {
			return nil, nil, fmt.Errorf("gov.role requires 'id' or 'sid' parameter")
		}
		result, err := m.govMod.Role(ctx, id)
		return result, nil, err

	case "roles":
		address := params["address"]
		if address == "" {
			return nil, nil, fmt.Errorf("gov.roles requires 'address' parameter")
		}
		result, err := m.govMod.Roles(ctx, address)
		return result, nil, err

	case "permissions":
		address := params["address"]
		if address == "" {
			return nil, nil, fmt.Errorf("gov.permissions requires 'address' parameter")
		}
		result, err := m.govMod.Permissions(ctx, address)
		return result, nil, err

	case "councilors":
		result, err := m.govMod.Councilors(ctx)
		return result, nil, err

	// Transaction actions
	case "vote", "proposal-vote":
		from := params["from"]
		proposalID := params["proposal_id"]
		optionStr := params["option"]

		if from == "" || proposalID == "" || optionStr == "" {
			return nil, nil, fmt.Errorf("gov.vote requires 'from', 'proposal_id', and 'option' parameters")
		}

		// Parse vote option (1=yes, 2=abstain, 3=no, 4=no_with_veto)
		voteOption := 1 // default to yes
		switch strings.ToLower(optionStr) {
		case "yes", "1":
			voteOption = 1
		case "abstain", "2":
			voteOption = 2
		case "no", "3":
			voteOption = 3
		case "no_with_veto", "noveto", "4":
			voteOption = 4
		}

		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.VoteProposal(ctx, from, proposalID, voteOption, opts)
		return resp, resp, err

	case "claim-seat", "councilor-claim-seat":
		from := params["from"]
		moniker := params["moniker"]
		if moniker == "" {
			moniker = params["from"]
		}

		if from == "" {
			return nil, nil, fmt.Errorf("gov.claim-seat requires 'from' parameter")
		}

		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.CouncilorClaimSeat(ctx, from, moniker, opts)
		return resp, resp, err

	case "role-create", "create-role":
		from := params["from"]
		sid := params["sid"]
		description := params["description"]

		if from == "" || sid == "" {
			return nil, nil, fmt.Errorf("gov.role-create requires 'from' and 'sid' parameters")
		}

		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.RoleCreate(ctx, from, sid, description, opts)
		return resp, resp, err

	case "role-assign", "assign-role":
		from := params["from"]
		address := params["address"]
		roleStr := params["role"]

		if from == "" || address == "" || roleStr == "" {
			return nil, nil, fmt.Errorf("gov.role-assign requires 'from', 'address', and 'role' parameters")
		}

		// Parse role ID as int
		roleID, err := strconv.Atoi(roleStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid role ID '%s': must be an integer", roleStr)
		}

		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.RoleAssign(ctx, from, address, roleID, opts)
		return resp, resp, err

	// Additional query actions
	case "votes":
		proposalID := params["proposal_id"]
		if proposalID == "" {
			proposalID = params["id"]
		}
		if proposalID == "" {
			return nil, nil, fmt.Errorf("gov.votes requires 'proposal_id' parameter")
		}
		result, err := m.govMod.Votes(ctx, proposalID)
		return result, nil, err

	case "voters":
		proposalID := params["proposal_id"]
		if proposalID == "" {
			proposalID = params["id"]
		}
		if proposalID == "" {
			return nil, nil, fmt.Errorf("gov.voters requires 'proposal_id' parameter")
		}
		result, err := m.govMod.Voters(ctx, proposalID)
		return result, nil, err

	case "execution-fee", "executionfee":
		txType := params["tx_type"]
		if txType == "" {
			return nil, nil, fmt.Errorf("gov.execution-fee requires 'tx_type' parameter")
		}
		result, err := m.govMod.ExecutionFee(ctx, txType)
		return result, nil, err

	case "all-execution-fees", "allexecutionfees":
		result, err := m.govMod.AllExecutionFees(ctx)
		return result, nil, err

	case "identity-record", "identityrecord":
		id := params["id"]
		if id == "" {
			return nil, nil, fmt.Errorf("gov.identity-record requires 'id' parameter")
		}
		result, err := m.govMod.IdentityRecord(ctx, id)
		return result, nil, err

	case "identity-records", "identityrecords":
		result, err := m.govMod.IdentityRecords(ctx)
		return result, nil, err

	case "identity-records-by-address", "identityrecordsbyaddress":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("gov.identity-records-by-address requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.govMod.IdentityRecordsByAddress(ctx, address)
		return result, nil, err

	case "data-registry", "dataregistry":
		key := params["key"]
		if key == "" {
			return nil, nil, fmt.Errorf("gov.data-registry requires 'key' parameter")
		}
		result, err := m.govMod.DataRegistry(ctx, key)
		return result, nil, err

	case "data-registry-keys", "dataregistrykeys":
		result, err := m.govMod.DataRegistryKeys(ctx)
		return result, nil, err

	case "polls":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("gov.polls requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.govMod.Polls(ctx, address)
		return result, nil, err

	case "poll-votes", "pollvotes":
		pollID := params["poll_id"]
		if pollID == "" {
			pollID = params["id"]
		}
		if pollID == "" {
			return nil, nil, fmt.Errorf("gov.poll-votes requires 'poll_id' parameter")
		}
		result, err := m.govMod.PollVotes(ctx, pollID)
		return result, nil, err

	case "poor-network-messages", "poornetworkmessages":
		result, err := m.govMod.PoorNetworkMessages(ctx)
		return result, nil, err

	case "custom-prefixes", "customprefixes":
		result, err := m.govMod.CustomPrefixes(ctx)
		return result, nil, err

	case "all-proposal-durations", "allproposaldurations":
		result, err := m.govMod.AllProposalDurations(ctx)
		return result, nil, err

	case "proposal-duration", "proposalduration":
		propType := params["proposal_type"]
		if propType == "" {
			propType = params["type"]
		}
		if propType == "" {
			return nil, nil, fmt.Errorf("gov.proposal-duration requires 'proposal_type' parameter")
		}
		result, err := m.govMod.ProposalDuration(ctx, propType)
		return result, nil, err

	case "non-councilors", "noncouncilors":
		result, err := m.govMod.NonCouncilors(ctx)
		return result, nil, err

	case "council-registry", "councilregistry":
		result, err := m.govMod.CouncilRegistry(ctx)
		return result, nil, err

	case "whitelisted-permission-addresses", "whitelistedpermissionaddresses":
		permission := params["permission"]
		if permission == "" {
			return nil, nil, fmt.Errorf("gov.whitelisted-permission-addresses requires 'permission' parameter")
		}
		result, err := m.govMod.WhitelistedPermissionAddresses(ctx, permission)
		return result, nil, err

	case "blacklisted-permission-addresses", "blacklistedpermissionaddresses":
		permission := params["permission"]
		if permission == "" {
			return nil, nil, fmt.Errorf("gov.blacklisted-permission-addresses requires 'permission' parameter")
		}
		result, err := m.govMod.BlacklistedPermissionAddresses(ctx, permission)
		return result, nil, err

	case "whitelisted-role-addresses", "whitelistedroleaddresses":
		role := params["role"]
		if role == "" {
			return nil, nil, fmt.Errorf("gov.whitelisted-role-addresses requires 'role' parameter")
		}
		result, err := m.govMod.WhitelistedRoleAddresses(ctx, role)
		return result, nil, err

	case "proposer-voters-count", "proposervoterscount":
		result, err := m.govMod.ProposerVotersCount(ctx)
		return result, nil, err

	case "all-identity-record-verify-requests", "allidentityrecordverifyrequests":
		result, err := m.govMod.AllIdentityRecordVerifyRequests(ctx)
		return result, nil, err

	case "identity-record-verify-request", "identityrecordverifyrequest":
		id := params["id"]
		if id == "" {
			return nil, nil, fmt.Errorf("gov.identity-record-verify-request requires 'id' parameter")
		}
		result, err := m.govMod.IdentityRecordVerifyRequest(ctx, id)
		return result, nil, err

	case "identity-record-verify-requests-by-approver", "identityrecordverifyrequestsbyapprover":
		approver := params["approver"]
		if approver == "" {
			return nil, nil, fmt.Errorf("gov.identity-record-verify-requests-by-approver requires 'approver' parameter")
		}
		result, err := m.govMod.IdentityRecordVerifyRequestsByApprover(ctx, approver)
		return result, nil, err

	case "identity-record-verify-requests-by-requester", "identityrecordverifyrequestsbyrequester":
		requester := params["requester"]
		if requester == "" {
			return nil, nil, fmt.Errorf("gov.identity-record-verify-requests-by-requester requires 'requester' parameter")
		}
		result, err := m.govMod.IdentityRecordVerifyRequestsByRequester(ctx, requester)
		return result, nil, err

	// Additional transaction actions
	case "councilor-activate", "counciloractivate":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("gov.councilor-activate requires 'from' parameter")
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.CouncilorActivate(ctx, from, opts)
		return resp, resp, err

	case "councilor-pause", "councilorpause":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("gov.councilor-pause requires 'from' parameter")
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.CouncilorPause(ctx, from, opts)
		return resp, resp, err

	case "councilor-unpause", "councilorunpause":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("gov.councilor-unpause requires 'from' parameter")
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.CouncilorUnpause(ctx, from, opts)
		return resp, resp, err

	case "permission-whitelist", "permissionwhitelist":
		from := params["from"]
		addrParam := params["address"]
		permStr := params["permission"]
		if from == "" || addrParam == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.permission-whitelist requires 'from', 'address', and 'permission' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.PermissionWhitelist(ctx, from, address, permission, opts)
		return resp, resp, err

	case "permission-blacklist", "permissionblacklist":
		from := params["from"]
		addrParam := params["address"]
		permStr := params["permission"]
		if from == "" || addrParam == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.permission-blacklist requires 'from', 'address', and 'permission' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.PermissionBlacklist(ctx, from, address, permission, opts)
		return resp, resp, err

	case "permission-remove-whitelisted", "permissionremovewhitelisted":
		from := params["from"]
		addrParam := params["address"]
		permStr := params["permission"]
		if from == "" || addrParam == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.permission-remove-whitelisted requires 'from', 'address', and 'permission' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.PermissionRemoveWhitelisted(ctx, from, address, permission, opts)
		return resp, resp, err

	case "permission-remove-blacklisted", "permissionremoveblacklisted":
		from := params["from"]
		addrParam := params["address"]
		permStr := params["permission"]
		if from == "" || addrParam == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.permission-remove-blacklisted requires 'from', 'address', and 'permission' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.PermissionRemoveBlacklisted(ctx, from, address, permission, opts)
		return resp, resp, err

	case "role-unassign", "unassign-role":
		from := params["from"]
		addrParam := params["address"]
		roleStr := params["role"]
		if from == "" || addrParam == "" || roleStr == "" {
			return nil, nil, fmt.Errorf("gov.role-unassign requires 'from', 'address', and 'role' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		roleID, err := strconv.Atoi(roleStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid role ID '%s': must be integer", roleStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.RoleUnassign(ctx, from, address, roleID, opts)
		return resp, resp, err

	case "role-whitelist-permission", "rolewhitelistpermission":
		from := params["from"]
		roleSID := params["role_sid"]
		permStr := params["permission"]
		if from == "" || roleSID == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.role-whitelist-permission requires 'from', 'role_sid', and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.RoleWhitelistPermission(ctx, from, roleSID, permission, opts)
		return resp, resp, err

	case "role-blacklist-permission", "roleblacklistpermission":
		from := params["from"]
		roleSID := params["role_sid"]
		permStr := params["permission"]
		if from == "" || roleSID == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.role-blacklist-permission requires 'from', 'role_sid', and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.RoleBlacklistPermission(ctx, from, roleSID, permission, opts)
		return resp, resp, err

	case "role-remove-whitelisted-permission", "roleremovewhitelistedpermission":
		from := params["from"]
		roleSID := params["role_sid"]
		permStr := params["permission"]
		if from == "" || roleSID == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.role-remove-whitelisted-permission requires 'from', 'role_sid', and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.RoleRemoveWhitelistedPermission(ctx, from, roleSID, permission, opts)
		return resp, resp, err

	case "role-remove-blacklisted-permission", "roleremoveblacklistedpermission":
		from := params["from"]
		roleSID := params["role_sid"]
		permStr := params["permission"]
		if from == "" || roleSID == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.role-remove-blacklisted-permission requires 'from', 'role_sid', and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.RoleRemoveBlacklistedPermission(ctx, from, roleSID, permission, opts)
		return resp, resp, err

	case "poll-create", "pollcreate":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("gov.poll-create requires 'from' parameter")
		}
		pollOpts := &gov.PollCreateOpts{
			Title:       params["title"],
			Description: params["description"],
			Reference:   params["reference"],
			Checksum:    params["checksum"],
			Roles:       params["roles"],
			PollType:    params["poll_type"],
			Options:     params["options"],
			Duration:    params["duration"],
		}
		if v := params["selection_count"]; v != "" {
			val, _ := strconv.Atoi(v)
			pollOpts.SelectionCount = val
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.PollCreate(ctx, from, pollOpts, opts)
		return resp, resp, err

	case "poll-vote", "pollvote":
		from := params["from"]
		pollID := params["poll_id"]
		options := params["options"]
		if from == "" || pollID == "" || options == "" {
			return nil, nil, fmt.Errorf("gov.poll-vote requires 'from', 'poll_id', and 'options' parameters")
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.PollVote(ctx, from, pollID, options, opts)
		return resp, resp, err

	case "set-network-properties", "setnetworkproperties":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("gov.set-network-properties requires 'from' parameter")
		}
		// Build properties map from params (excluding 'from')
		properties := make(map[string]string)
		for k, v := range params {
			if k != "from" && v != "" {
				properties[k] = v
			}
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.SetNetworkProperties(ctx, from, properties, opts)
		return resp, resp, err

	case "set-execution-fee", "setexecutionfee":
		from := params["from"]
		txType := params["tx_type"]
		if from == "" || txType == "" {
			return nil, nil, fmt.Errorf("gov.set-execution-fee requires 'from' and 'tx_type' parameters")
		}
		execFee, _ := strconv.ParseUint(params["execution_fee"], 10, 64)
		failFee, _ := strconv.ParseUint(params["failure_fee"], 10, 64)
		timeout, _ := strconv.ParseUint(params["timeout"], 10, 64)
		defaultParams, _ := strconv.ParseUint(params["default_params"], 10, 64)
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.SetExecutionFee(ctx, from, txType, execFee, failFee, timeout, defaultParams, opts)
		return resp, resp, err

	case "register-identity-records", "registeridentityrecords":
		from := params["from"]
		infosJSON := params["infos_json"]
		if from == "" || infosJSON == "" {
			return nil, nil, fmt.Errorf("gov.register-identity-records requires 'from' and 'infos_json' parameters")
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.RegisterIdentityRecords(ctx, from, infosJSON, opts)
		return resp, resp, err

	case "delete-identity-records", "deleteidentityrecords":
		from := params["from"]
		keys := params["keys"]
		if from == "" || keys == "" {
			return nil, nil, fmt.Errorf("gov.delete-identity-records requires 'from' and 'keys' parameters")
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.DeleteIdentityRecords(ctx, from, keys, opts)
		return resp, resp, err

	case "request-identity-record-verify", "requestidentityrecordverify":
		from := params["from"]
		verifier := params["verifier"]
		recordIDs := params["record_ids"]
		if from == "" || verifier == "" || recordIDs == "" {
			return nil, nil, fmt.Errorf("gov.request-identity-record-verify requires 'from', 'verifier', and 'record_ids' parameters")
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.RequestIdentityRecordVerify(ctx, from, verifier, recordIDs, params["verifier_tip"], opts)
		return resp, resp, err

	case "handle-identity-records-verify-request", "handleidentityrecordsverifyrequest":
		from := params["from"]
		requestID := params["request_id"]
		if from == "" || requestID == "" {
			return nil, nil, fmt.Errorf("gov.handle-identity-records-verify-request requires 'from' and 'request_id' parameters")
		}
		approve := params["approve"] == "true"
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.HandleIdentityRecordsVerifyRequest(ctx, from, requestID, approve, opts)
		return resp, resp, err

	case "cancel-identity-records-verify-request", "cancelidentityrecordsverifyrequest":
		from := params["from"]
		requestID := params["request_id"]
		if from == "" || requestID == "" {
			return nil, nil, fmt.Errorf("gov.cancel-identity-records-verify-request requires 'from' and 'request_id' parameters")
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.CancelIdentityRecordsVerifyRequest(ctx, from, requestID, opts)
		return resp, resp, err

	case "proposal-assign-role", "proposalassignrole":
		from := params["from"]
		addrParam := params["address"]
		role := params["role"]
		if from == "" || addrParam == "" || role == "" {
			return nil, nil, fmt.Errorf("gov.proposal-assign-role requires 'from', 'address', and 'role' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalAssignRole(ctx, from, address, role, params["title"], params["description"], opts)
		return resp, resp, err

	case "proposal-unassign-role", "proposalunassignrole":
		from := params["from"]
		addrParam := params["address"]
		role := params["role"]
		if from == "" || addrParam == "" || role == "" {
			return nil, nil, fmt.Errorf("gov.proposal-unassign-role requires 'from', 'address', and 'role' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalUnassignRole(ctx, from, address, role, params["title"], params["description"], opts)
		return resp, resp, err

	case "proposal-whitelist-permission", "proposalwhitelistpermission":
		from := params["from"]
		addrParam := params["address"]
		permStr := params["permission"]
		if from == "" || addrParam == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-whitelist-permission requires 'from', 'address', and 'permission' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalWhitelistPermission(ctx, from, address, permission, params["description"], opts)
		return resp, resp, err

	case "proposal-blacklist-permission", "proposalblacklistpermission":
		from := params["from"]
		addrParam := params["address"]
		permStr := params["permission"]
		if from == "" || addrParam == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-blacklist-permission requires 'from', 'address', and 'permission' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalBlacklistPermission(ctx, from, address, permission, params["description"], opts)
		return resp, resp, err

	case "proposal-remove-whitelisted-permission", "proposalremovewhitelistedpermission":
		from := params["from"]
		addrParam := params["address"]
		permStr := params["permission"]
		if from == "" || addrParam == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-remove-whitelisted-permission requires 'from', 'address', and 'permission' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalRemoveWhitelistedPermission(ctx, from, address, permission, params["description"], opts)
		return resp, resp, err

	case "proposal-remove-blacklisted-permission", "proposalremoveblacklistedpermission":
		from := params["from"]
		addrParam := params["address"]
		permStr := params["permission"]
		if from == "" || addrParam == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-remove-blacklisted-permission requires 'from', 'address', and 'permission' parameters")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalRemoveBlacklistedPermission(ctx, from, address, permission, params["description"], opts)
		return resp, resp, err

	case "proposal-create-role", "proposalcreaterole":
		from := params["from"]
		roleSID := params["role_sid"]
		roleDescription := params["role_description"]
		if from == "" || roleSID == "" {
			return nil, nil, fmt.Errorf("gov.proposal-create-role requires 'from' and 'role_sid' parameters")
		}
		propOpts := &gov.ProposalCreateRoleOpts{
			Title:       params["title"],
			Description: params["description"],
			Whitelist:   params["whitelist"],
			Blacklist:   params["blacklist"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalCreateRole(ctx, from, roleSID, roleDescription, propOpts, opts)
		return resp, resp, err

	case "proposal-remove-role", "proposalremoverole":
		from := params["from"]
		roleSID := params["role_sid"]
		if from == "" || roleSID == "" {
			return nil, nil, fmt.Errorf("gov.proposal-remove-role requires 'from' and 'role_sid' parameters")
		}
		propOpts := &gov.ProposalRoleOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalRemoveRole(ctx, from, roleSID, propOpts, opts)
		return resp, resp, err

	case "proposal-whitelist-role-permission", "proposalwhitelistrolepermission":
		from := params["from"]
		roleSID := params["role_sid"]
		permStr := params["permission"]
		if from == "" || roleSID == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-whitelist-role-permission requires 'from', 'role_sid', and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		propOpts := &gov.ProposalRoleOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalWhitelistRolePermission(ctx, from, roleSID, permission, propOpts, opts)
		return resp, resp, err

	case "proposal-blacklist-role-permission", "proposalblacklistrolepermission":
		from := params["from"]
		roleSID := params["role_sid"]
		permStr := params["permission"]
		if from == "" || roleSID == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-blacklist-role-permission requires 'from', 'role_sid', and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		propOpts := &gov.ProposalRoleOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalBlacklistRolePermission(ctx, from, roleSID, permission, propOpts, opts)
		return resp, resp, err

	case "proposal-remove-whitelisted-role-permission", "proposalremovewhitelistedrolepermission":
		from := params["from"]
		roleSID := params["role_sid"]
		permStr := params["permission"]
		if from == "" || roleSID == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-remove-whitelisted-role-permission requires 'from', 'role_sid', and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		propOpts := &gov.ProposalRoleOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalRemoveWhitelistedRolePermission(ctx, from, roleSID, permission, propOpts, opts)
		return resp, resp, err

	case "proposal-remove-blacklisted-role-permission", "proposalremoveblacklistedrolepermission":
		from := params["from"]
		roleSID := params["role_sid"]
		permStr := params["permission"]
		if from == "" || roleSID == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-remove-blacklisted-role-permission requires 'from', 'role_sid', and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		propOpts := &gov.ProposalRoleOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalRemoveBlacklistedRolePermission(ctx, from, roleSID, permission, propOpts, opts)
		return resp, resp, err

	case "proposal-set-network-property", "proposalsetnetworkproperty":
		from := params["from"]
		property := params["property"]
		value := params["value"]
		title := params["title"]
		if from == "" || property == "" || value == "" || title == "" {
			return nil, nil, fmt.Errorf("gov.proposal-set-network-property requires 'from', 'property', 'value', and 'title' parameters")
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalSetNetworkProperty(ctx, from, property, value, title, params["description"], opts)
		return resp, resp, err

	case "proposal-set-poor-network-msgs", "proposalsetpoornetworkmsgs":
		from := params["from"]
		messages := params["messages"]
		if from == "" || messages == "" {
			return nil, nil, fmt.Errorf("gov.proposal-set-poor-network-msgs requires 'from' and 'messages' parameters")
		}
		propOpts := &gov.ProposalOtherOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalSetPoorNetworkMsgs(ctx, from, messages, propOpts, opts)
		return resp, resp, err

	case "proposal-set-execution-fees", "proposalsetexecutionfees":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("gov.proposal-set-execution-fees requires 'from' parameter")
		}
		propOpts := &gov.ProposalSetExecutionFeesOpts{
			Title:         params["title"],
			Description:   params["description"],
			TxTypes:       params["tx_types"],
			ExecutionFees: params["execution_fees"],
			FailureFees:   params["failure_fees"],
			Timeouts:      params["timeouts"],
			DefaultParams: params["default_params"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalSetExecutionFees(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-upsert-data-registry", "proposalupsertdataregistry":
		from := params["from"]
		key := params["key"]
		hash := params["hash"]
		reference := params["reference"]
		encoding := params["encoding"]
		size := params["size"]
		if from == "" || key == "" {
			return nil, nil, fmt.Errorf("gov.proposal-upsert-data-registry requires 'from' and 'key' parameters")
		}
		propOpts := &gov.ProposalOtherOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalUpsertDataRegistry(ctx, from, key, hash, reference, encoding, size, propOpts, opts)
		return resp, resp, err

	case "proposal-set-proposal-durations", "proposalsetproposaldurations":
		from := params["from"]
		proposalTypes := params["proposal_types"]
		durations := params["durations"]
		if from == "" || proposalTypes == "" || durations == "" {
			return nil, nil, fmt.Errorf("gov.proposal-set-proposal-durations requires 'from', 'proposal_types', and 'durations' parameters")
		}
		propOpts := &gov.ProposalOtherOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalSetProposalDurations(ctx, from, proposalTypes, durations, propOpts, opts)
		return resp, resp, err

	case "proposal-jail-councilor", "proposaljailcouncilor":
		from := params["from"]
		councilors := params["councilors"]
		if from == "" || councilors == "" {
			return nil, nil, fmt.Errorf("gov.proposal-jail-councilor requires 'from' and 'councilors' parameters")
		}
		propOpts := &gov.ProposalOtherOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalJailCouncilor(ctx, from, councilors, propOpts, opts)
		return resp, resp, err

	case "proposal-reset-whole-councilor-rank", "proposalresetwholecouncilorrank":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("gov.proposal-reset-whole-councilor-rank requires 'from' parameter")
		}
		propOpts := &gov.ProposalOtherOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalResetWholeCouncilorRank(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-whitelist-account-permission", "proposalwhitelistaccountpermission":
		from := params["from"]
		permStr := params["permission"]
		if from == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-whitelist-account-permission requires 'from' and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		propOpts := &gov.ProposalAccountPermissionOpts{
			Title:       params["title"],
			Description: params["description"],
			Addr:        params["address"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalWhitelistAccountPermission(ctx, from, permission, propOpts, opts)
		return resp, resp, err

	case "proposal-blacklist-account-permission", "proposalblacklistaccountpermission":
		from := params["from"]
		permStr := params["permission"]
		if from == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-blacklist-account-permission requires 'from' and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		propOpts := &gov.ProposalAccountPermissionOpts{
			Title:       params["title"],
			Description: params["description"],
			Addr:        params["address"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalBlacklistAccountPermission(ctx, from, permission, propOpts, opts)
		return resp, resp, err

	case "proposal-remove-whitelisted-account-permission", "proposalremovewhitelistedaccountpermission":
		from := params["from"]
		permStr := params["permission"]
		if from == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-remove-whitelisted-account-permission requires 'from' and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		propOpts := &gov.ProposalAccountPermissionOpts{
			Title:       params["title"],
			Description: params["description"],
			Addr:        params["address"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalRemoveWhitelistedAccountPermission(ctx, from, permission, propOpts, opts)
		return resp, resp, err

	case "proposal-remove-blacklisted-account-permission", "proposalremoveblacklistedaccountpermission":
		from := params["from"]
		permStr := params["permission"]
		if from == "" || permStr == "" {
			return nil, nil, fmt.Errorf("gov.proposal-remove-blacklisted-account-permission requires 'from' and 'permission' parameters")
		}
		permission, err := strconv.Atoi(permStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid permission '%s': must be integer", permStr)
		}
		propOpts := &gov.ProposalAccountPermissionOpts{
			Title:       params["title"],
			Description: params["description"],
			Addr:        params["address"],
		}
		opts := m.buildGovTxOptions(txOpts)
		resp, err := m.govMod.ProposalRemoveBlacklistedAccountPermission(ctx, from, permission, propOpts, opts)
		return resp, resp, err

	default:
		return nil, nil, fmt.Errorf("unknown gov action: %s", action)
	}
}

// executeStaking handles staking module actions.
func (m *ActionMapper) executeStaking(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.stakingMod == nil {
		m.stakingMod = staking.New(m.client)
	}

	switch action {
	case "validators":
		opts := &staking.ValidatorQueryOpts{
			Address: params["address"],
			ValAddr: params["val_address"],
			Moniker: params["moniker"],
			Status:  params["status"],
		}
		result, err := m.stakingMod.Validators(ctx, opts)
		return result, nil, err

	case "validator":
		address := params["address"]
		valAddr := params["val_address"]
		moniker := params["moniker"]

		if address == "" && valAddr == "" && moniker == "" {
			return nil, nil, fmt.Errorf("staking.validator requires 'address', 'val_address', or 'moniker' parameter")
		}

		opts := &staking.ValidatorQueryOpts{
			Address: address,
			ValAddr: valAddr,
			Moniker: moniker,
		}
		result, err := m.stakingMod.Validator(ctx, opts)
		return result, nil, err

	case "claim-validator-seat", "claimvalidatorseat":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("staking.claim-validator-seat requires 'from' parameter")
		}

		seatOpts := &staking.ClaimValidatorSeatOpts{
			Moniker: params["moniker"],
			PubKey:  params["pubkey"],
		}
		opts := m.buildStakingTxOptions(txOpts)
		resp, err := m.stakingMod.ClaimValidatorSeat(ctx, from, seatOpts, opts)
		return resp, resp, err

	case "proposal-unjail-validator", "proposalunjailvalidator":
		from := params["from"]
		valAddr := params["val_address"]
		reference := params["reference"]
		if from == "" || valAddr == "" {
			return nil, nil, fmt.Errorf("staking.proposal-unjail-validator requires 'from' and 'val_address' parameters")
		}

		propOpts := &staking.ProposalUnjailValidatorOpts{
			Title:       params["title"],
			Description: params["description"],
		}
		opts := m.buildStakingTxOptions(txOpts)
		resp, err := m.stakingMod.ProposalUnjailValidator(ctx, from, valAddr, reference, propOpts, opts)
		return resp, resp, err

	default:
		return nil, nil, fmt.Errorf("unknown staking action: %s", action)
	}
}

// executeMultistaking handles multistaking module actions.
func (m *ActionMapper) executeMultistaking(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.multistakeMod == nil {
		m.multistakeMod = multistaking.New(m.client)
	}

	switch action {
	case "pools":
		result, err := m.multistakeMod.Pools(ctx)
		return result, nil, err

	case "delegate":
		from := params["from"]
		validator := params["validator"]
		coins := params["coins"]

		if from == "" || validator == "" || coins == "" {
			return nil, nil, fmt.Errorf("multistaking.delegate requires 'from', 'validator', and 'coins' parameters")
		}

		opts := m.buildMultistakingTxOptions(txOpts)
		resp, err := m.multistakeMod.Delegate(ctx, from, validator, coins, opts)
		return resp, resp, err

	case "undelegate":
		from := params["from"]
		validator := params["validator"]
		coins := params["coins"]

		if from == "" || validator == "" || coins == "" {
			return nil, nil, fmt.Errorf("multistaking.undelegate requires 'from', 'validator', and 'coins' parameters")
		}

		opts := m.buildMultistakingTxOptions(txOpts)
		resp, err := m.multistakeMod.Undelegate(ctx, from, validator, coins, opts)
		return resp, resp, err

	case "claim-rewards":
		from := params["from"]
		validator := params["validator"]

		if from == "" {
			return nil, nil, fmt.Errorf("multistaking.claim-rewards requires 'from' parameter")
		}

		opts := m.buildMultistakingTxOptions(txOpts)
		resp, err := m.multistakeMod.ClaimRewards(ctx, from, validator, opts)
		return resp, resp, err

	case "undelegations":
		delegatorParam := params["delegator"]
		validator := params["validator"]
		if delegatorParam == "" || validator == "" {
			return nil, nil, fmt.Errorf("multistaking.undelegations requires 'delegator' and 'validator' parameters")
		}
		delegator, err := m.resolveAddress(ctx, delegatorParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.multistakeMod.Undelegations(ctx, delegator, validator)
		return result, nil, err

	case "outstanding-rewards", "outstandingrewards":
		delegatorParam := params["delegator"]
		if delegatorParam == "" {
			return nil, nil, fmt.Errorf("multistaking.outstanding-rewards requires 'delegator' parameter")
		}
		delegator, err := m.resolveAddress(ctx, delegatorParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.multistakeMod.OutstandingRewards(ctx, delegator)
		return result, nil, err

	case "compound-info", "compoundinfo":
		delegatorParam := params["delegator"]
		if delegatorParam == "" {
			return nil, nil, fmt.Errorf("multistaking.compound-info requires 'delegator' parameter")
		}
		delegator, err := m.resolveAddress(ctx, delegatorParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.multistakeMod.CompoundInfo(ctx, delegator)
		return result, nil, err

	case "staking-pool-delegators", "stakingpooldelegators":
		validator := params["validator"]
		if validator == "" {
			return nil, nil, fmt.Errorf("multistaking.staking-pool-delegators requires 'validator' parameter")
		}
		result, err := m.multistakeMod.StakingPoolDelegators(ctx, validator)
		return result, nil, err

	case "claim-undelegation", "claimundelegation":
		from := params["from"]
		undelegationID := params["undelegation_id"]
		if from == "" || undelegationID == "" {
			return nil, nil, fmt.Errorf("multistaking.claim-undelegation requires 'from' and 'undelegation_id' parameters")
		}
		opts := m.buildMultistakingTxOptions(txOpts)
		resp, err := m.multistakeMod.ClaimUndelegation(ctx, from, undelegationID, opts)
		return resp, resp, err

	case "claim-matured-undelegations", "claimmaturedundelegations":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("multistaking.claim-matured-undelegations requires 'from' parameter")
		}
		opts := m.buildMultistakingTxOptions(txOpts)
		resp, err := m.multistakeMod.ClaimMaturedUndelegations(ctx, from, opts)
		return resp, resp, err

	case "register-delegator", "registerdelegator":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("multistaking.register-delegator requires 'from' parameter")
		}
		opts := m.buildMultistakingTxOptions(txOpts)
		resp, err := m.multistakeMod.RegisterDelegator(ctx, from, opts)
		return resp, resp, err

	case "set-compound-info", "setcompoundinfo":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("multistaking.set-compound-info requires 'from' parameter")
		}
		allDenom := params["all_denom"] == "true"
		specificDenoms := params["specific_denoms"]
		opts := m.buildMultistakingTxOptions(txOpts)
		resp, err := m.multistakeMod.SetCompoundInfo(ctx, from, allDenom, specificDenoms, opts)
		return resp, resp, err

	case "upsert-staking-pool", "upsertstakingpool":
		from := params["from"]
		validatorKey := params["validator_key"]
		if from == "" || validatorKey == "" {
			return nil, nil, fmt.Errorf("multistaking.upsert-staking-pool requires 'from' and 'validator_key' parameters")
		}
		poolOpts := &multistaking.UpsertStakingPoolOpts{
			Enabled:    params["enabled"] == "true",
			Commission: params["commission"],
		}
		opts := m.buildMultistakingTxOptions(txOpts)
		resp, err := m.multistakeMod.UpsertStakingPool(ctx, from, validatorKey, poolOpts, opts)
		return resp, resp, err

	default:
		return nil, nil, fmt.Errorf("unknown multistaking action: %s", action)
	}
}

// executeTokens handles tokens module actions.
func (m *ActionMapper) executeTokens(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.tokensMod == nil {
		m.tokensMod = tokens.New(m.client)
	}

	switch action {
	case "all-rates":
		result, err := m.tokensMod.AllRates(ctx)
		return result, nil, err

	case "rate":
		denom := params["denom"]
		if denom == "" {
			return nil, nil, fmt.Errorf("tokens.rate requires 'denom' parameter")
		}
		result, err := m.tokensMod.Rate(ctx, denom)
		return result, nil, err

	case "rates-by-denom", "ratesbydenom":
		denom := params["denom"]
		if denom == "" {
			return nil, nil, fmt.Errorf("tokens.rates-by-denom requires 'denom' parameter")
		}
		result, err := m.tokensMod.RatesByDenom(ctx, denom)
		return result, nil, err

	case "token-black-whites":
		result, err := m.tokensMod.TokenBlackWhites(ctx)
		return result, nil, err

	case "upsert-rate", "upsertrate":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("tokens.upsert-rate requires 'from' parameter")
		}

		rateOpts := &tokens.UpsertRateOpts{
			Denom:       params["denom"],
			FeeRate:     params["fee_rate"],
			FeePayments: params["fee_payments"] != "false",
			Name:        params["name"],
			Symbol:      params["symbol"],
			Description: params["description"],
			Icon:        params["icon"],
			Website:     params["website"],
			Social:      params["social"],
			StakeToken:  params["stake_token"] == "true",
			StakeCap:    params["stake_cap"],
			StakeMin:    params["stake_min"],
			Supply:      params["supply"],
			SupplyCap:   params["supply_cap"],
			NftHash:     params["nft_hash"],
			NftMetadata: params["nft_metadata"],
			Owner:       params["owner"],
			TokenRate:   params["token_rate"],
			TokenType:   params["token_type"],
			MintingFee:  params["minting_fee"],
		}
		if v := params["decimals"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 32)
			rateOpts.Decimals = uint32(val)
		}
		if params["owner_edit_disabled"] == "true" {
			rateOpts.OwnerEditDisabled = true
		}
		if params["invalidated"] == "true" {
			rateOpts.Invalidated = true
		}

		opts := m.buildTokensTxOptions(txOpts)
		resp, err := m.tokensMod.UpsertRate(ctx, from, rateOpts, opts)
		return resp, resp, err

	case "proposal-upsert-rate", "proposalupsertrate":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("tokens.proposal-upsert-rate requires 'from' parameter")
		}

		propOpts := &tokens.ProposalUpsertRateOpts{
			Denom:       params["denom"],
			FeeRate:     params["fee_rate"],
			FeePayments: params["fee_payments"] != "false",
			Name:        params["name"],
			Symbol:      params["symbol"],
			Icon:        params["icon"],
			Website:     params["website"],
			Social:      params["social"],
			StakeToken:  params["stake_token"] == "true",
			StakeCap:    params["stake_cap"],
			StakeMin:    params["stake_min"],
			Supply:      params["supply"],
			SupplyCap:   params["supply_cap"],
			NftHash:     params["nft_hash"],
			NftMetadata: params["nft_metadata"],
			Owner:       params["owner"],
			TokenRate:   params["token_rate"],
			TokenType:   params["token_type"],
			MintingFee:  params["minting_fee"],
			Title:       params["title"],
			Description: params["description"],
		}
		if v := params["decimals"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 32)
			propOpts.Decimals = uint32(val)
		}
		if params["owner_edit_disabled"] == "true" {
			propOpts.OwnerEditDisabled = true
		}
		if params["invalidated"] == "true" {
			propOpts.Invalidated = true
		}

		opts := m.buildTokensTxOptions(txOpts)
		resp, err := m.tokensMod.ProposalUpsertRate(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-update-tokens-black-white", "proposalupdatetokensblackwhite":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("tokens.proposal-update-tokens-black-white requires 'from' parameter")
		}

		propOpts := &tokens.ProposalUpdateTokensBlackWhiteOpts{
			IsBlacklist: params["is_blacklist"] == "true",
			IsAdd:       params["is_add"] == "true",
			Title:       params["title"],
			Description: params["description"],
		}
		if tokensStr := params["tokens"]; tokensStr != "" {
			propOpts.Tokens = strings.Split(tokensStr, ",")
		}

		opts := m.buildTokensTxOptions(txOpts)
		resp, err := m.tokensMod.ProposalUpdateTokensBlackWhite(ctx, from, propOpts, opts)
		return resp, resp, err

	default:
		return nil, nil, fmt.Errorf("unknown tokens action: %s", action)
	}
}

// executeStatus handles status module actions.
func (m *ActionMapper) executeStatus(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.statusMod == nil {
		m.statusMod = status.New(m.client)
	}

	switch action {
	case "status":
		result, err := m.statusMod.Status(ctx)
		return result, nil, err

	case "node-info", "nodeinfo":
		result, err := m.statusMod.NodeInfo(ctx)
		return result, nil, err

	case "sync-info", "syncinfo":
		result, err := m.statusMod.SyncInfo(ctx)
		return result, nil, err

	case "validator-info", "validatorinfo":
		result, err := m.statusMod.ValidatorInfo(ctx)
		return result, nil, err

	case "chain-id", "chainid":
		result, err := m.statusMod.ChainID(ctx)
		return result, nil, err

	case "latest-block-height", "latestblockheight", "height":
		result, err := m.statusMod.LatestBlockHeight(ctx)
		return result, nil, err

	case "is-syncing", "issyncing", "syncing":
		result, err := m.statusMod.IsSyncing(ctx)
		return result, nil, err

	case "network-properties", "networkproperties":
		result, err := m.statusMod.NetworkProperties(ctx)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown status action: %s", action)
	}
}

// executeCustody handles custody module actions.
func (m *ActionMapper) executeCustody(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.custodyMod == nil {
		m.custodyMod = custody.New(m.client)
	}

	switch action {
	case "get":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("custody.get requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.custodyMod.Get(ctx, address)
		return result, nil, err

	case "custodians":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("custody.custodians requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.custodyMod.Custodians(ctx, address)
		return result, nil, err

	case "custodians-pool", "custodianspool", "pool":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("custody.custodians-pool requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.custodyMod.CustodiansPool(ctx, address)
		return result, nil, err

	case "whitelist":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("custody.whitelist requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.custodyMod.Whitelist(ctx, address)
		return result, nil, err

	case "limits":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("custody.limits requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.custodyMod.Limits(ctx, address)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown custody action: %s", action)
	}
}

// executeParams handles params module actions.
func (m *ActionMapper) executeParams(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.paramsMod == nil {
		m.paramsMod = paramsmod.New(m.client)
	}

	switch action {
	case "subspace":
		subspace := params["subspace"]
		key := params["key"]
		if subspace == "" || key == "" {
			return nil, nil, fmt.Errorf("params.subspace requires 'subspace' and 'key' parameters")
		}
		result, err := m.paramsMod.Subspace(ctx, subspace, key)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown params action: %s", action)
	}
}

// executeEthereum handles ethereum module actions.
func (m *ActionMapper) executeEthereum(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.ethereumMod == nil {
		m.ethereumMod = ethereum.New(m.client)
	}

	switch action {
	case "state":
		result, err := m.ethereumMod.State(ctx)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown ethereum action: %s", action)
	}
}

// executeEvidence handles evidence module actions.
func (m *ActionMapper) executeEvidence(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.evidenceMod == nil {
		m.evidenceMod = evidence.New(m.client)
	}

	switch action {
	case "all", "all-evidence":
		result, err := m.evidenceMod.AllEvidence(ctx)
		return result, nil, err

	case "evidence", "get":
		hash := params["hash"]
		if hash == "" {
			return nil, nil, fmt.Errorf("evidence.evidence requires 'hash' parameter")
		}
		result, err := m.evidenceMod.Evidence(ctx, hash)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown evidence action: %s", action)
	}
}

// executeDistributor handles distributor module actions.
func (m *ActionMapper) executeDistributor(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.distributorMod == nil {
		m.distributorMod = distributor.New(m.client)
	}

	switch action {
	case "fees-treasury", "feestreasury":
		result, err := m.distributorMod.FeesTreasury(ctx)
		return result, nil, err

	case "periodic-snapshot", "periodicsnapshot":
		result, err := m.distributorMod.PeriodicSnapshot(ctx)
		return result, nil, err

	case "snapshot-period", "snapshotperiod":
		result, err := m.distributorMod.SnapshotPeriod(ctx)
		return result, nil, err

	case "snapshot-period-performance", "snapshotperiodperformance":
		validator := params["validator"]
		if validator == "" {
			return nil, nil, fmt.Errorf("distributor.snapshot-period-performance requires 'validator' parameter")
		}
		result, err := m.distributorMod.SnapshotPeriodPerformance(ctx, validator)
		return result, nil, err

	case "year-start-snapshot", "yearstartsnapshot":
		result, err := m.distributorMod.YearStartSnapshot(ctx)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown distributor action: %s", action)
	}
}

// executeLayer2 handles layer2 module actions.
func (m *ActionMapper) executeLayer2(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.layer2Mod == nil {
		m.layer2Mod = layer2.New(m.client)
	}

	switch action {
	case "all-dapps", "alldapps":
		result, err := m.layer2Mod.AllDapps(ctx)
		return result, nil, err

	case "execution-registrar", "executionregistrar":
		dappName := params["dapp_name"]
		if dappName == "" {
			dappName = params["dapp"]
		}
		if dappName == "" {
			return nil, nil, fmt.Errorf("layer2.execution-registrar requires 'dapp_name' parameter")
		}
		result, err := m.layer2Mod.ExecutionRegistrar(ctx, dappName)
		return result, nil, err

	case "transfer-dapps", "transferdapps":
		result, err := m.layer2Mod.TransferDapps(ctx)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown layer2 action: %s", action)
	}
}

// executeRecovery handles recovery module actions.
func (m *ActionMapper) executeRecovery(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.recoveryMod == nil {
		m.recoveryMod = recovery.New(m.client)
	}

	switch action {
	case "recovery-record", "recoveryrecord":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("recovery.recovery-record requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.recoveryMod.RecoveryRecord(ctx, address)
		return result, nil, err

	case "recovery-token", "recoverytoken":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("recovery.recovery-token requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.recoveryMod.RecoveryToken(ctx, address)
		return result, nil, err

	case "rr-holder-rewards", "rrholderrewards":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("recovery.rr-holder-rewards requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.recoveryMod.RRHolderRewards(ctx, address)
		return result, nil, err

	case "rr-holders", "rrholders":
		rrToken := params["rr_token"]
		if rrToken == "" {
			rrToken = params["token"]
		}
		if rrToken == "" {
			return nil, nil, fmt.Errorf("recovery.rr-holders requires 'rr_token' parameter")
		}
		result, err := m.recoveryMod.RRHolders(ctx, rrToken)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown recovery action: %s", action)
	}
}

// executeSlashing handles slashing module actions.
func (m *ActionMapper) executeSlashing(ctx context.Context, action string, params map[string]string) (interface{}, *sdk.TxResponse, error) {
	if m.slashingMod == nil {
		m.slashingMod = slashing.New(m.client)
	}

	switch action {
	case "signing-info", "signinginfo":
		consAddress := params["cons_address"]
		if consAddress == "" {
			consAddress = params["address"]
		}
		if consAddress == "" {
			return nil, nil, fmt.Errorf("slashing.signing-info requires 'cons_address' parameter")
		}
		result, err := m.slashingMod.SigningInfo(ctx, consAddress)
		return result, nil, err

	case "signing-infos", "signinginfos":
		result, err := m.slashingMod.SigningInfos(ctx)
		return result, nil, err

	case "active-staking-pools", "activestakingpools":
		result, err := m.slashingMod.ActiveStakingPools(ctx)
		return result, nil, err

	case "inactive-staking-pools", "inactivestakingpools":
		result, err := m.slashingMod.InactiveStakingPools(ctx)
		return result, nil, err

	case "slashed-staking-pools", "slashedstakingpools":
		result, err := m.slashingMod.SlashedStakingPools(ctx)
		return result, nil, err

	case "slash-proposals", "slashproposals":
		result, err := m.slashingMod.SlashProposals(ctx)
		return result, nil, err

	default:
		return nil, nil, fmt.Errorf("unknown slashing action: %s", action)
	}
}

// executeBridge handles bridge module actions.
func (m *ActionMapper) executeBridge(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.bridgeMod == nil {
		m.bridgeMod = bridge.New(m.client)
	}

	switch action {
	case "get-cosmos-ethereum", "getcosmosEthereum", "cosmos-ethereum":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("bridge.get-cosmos-ethereum requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.bridgeMod.GetCosmosEthereum(ctx, address)
		return result, nil, err

	case "get-ethereum-cosmos", "getethereumcosmos", "ethereum-cosmos":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("bridge.get-ethereum-cosmos requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.bridgeMod.GetEthereumCosmos(ctx, address)
		return result, nil, err

	case "change-cosmos-ethereum", "changecosmosEthereum":
		from := params["from"]
		cosmosAddress := params["cosmos_address"]
		ethereumAddress := params["ethereum_address"]
		amount := params["amount"]

		if from == "" || cosmosAddress == "" || ethereumAddress == "" || amount == "" {
			return nil, nil, fmt.Errorf("bridge.change-cosmos-ethereum requires 'from', 'cosmos_address', 'ethereum_address', and 'amount' parameters")
		}

		opts := m.buildBridgeTxOptions(txOpts)
		resp, err := m.bridgeMod.ChangeCosmosEthereum(ctx, from, cosmosAddress, ethereumAddress, amount, opts)
		return resp, resp, err

	case "change-ethereum-cosmos", "changeethereumcosmos":
		from := params["from"]
		cosmosAddress := params["cosmos_address"]
		ethereumTxHash := params["ethereum_tx_hash"]
		amount := params["amount"]

		if from == "" || cosmosAddress == "" || ethereumTxHash == "" || amount == "" {
			return nil, nil, fmt.Errorf("bridge.change-ethereum-cosmos requires 'from', 'cosmos_address', 'ethereum_tx_hash', and 'amount' parameters")
		}

		opts := m.buildBridgeTxOptions(txOpts)
		resp, err := m.bridgeMod.ChangeEthereumCosmos(ctx, from, cosmosAddress, ethereumTxHash, amount, opts)
		return resp, resp, err

	default:
		return nil, nil, fmt.Errorf("unknown bridge action: %s", action)
	}
}

// executeUbi handles ubi module actions.
func (m *ActionMapper) executeUbi(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.ubiMod == nil {
		m.ubiMod = ubi.New(m.client)
	}

	switch action {
	case "records", "ubi-records":
		result, err := m.ubiMod.Records(ctx)
		return result, nil, err

	case "record-by-name", "recordbyname", "record":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("ubi.record-by-name requires 'name' parameter")
		}
		result, err := m.ubiMod.RecordByName(ctx, name)
		return result, nil, err

	case "proposal-upsert-ubi", "proposalupsertubi":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("ubi.proposal-upsert-ubi requires 'from' parameter")
		}

		propOpts := &ubi.ProposalUpsertUBIOpts{
			Name:        params["name"],
			PoolName:    params["pool_name"],
			Title:       params["title"],
			Description: params["description"],
		}
		if v := params["distr_start"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			propOpts.DistrStart = val
		}
		if v := params["distr_end"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			propOpts.DistrEnd = val
		}
		if v := params["amount"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			propOpts.Amount = val
		}
		if v := params["period"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			propOpts.Period = val
		}

		opts := m.buildUbiTxOptions(txOpts)
		resp, err := m.ubiMod.ProposalUpsertUBI(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-remove-ubi", "proposalremoveubi":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("ubi.proposal-remove-ubi requires 'from' parameter")
		}

		propOpts := &ubi.ProposalRemoveUBIOpts{
			Name:        params["name"],
			Title:       params["title"],
			Description: params["description"],
		}

		opts := m.buildUbiTxOptions(txOpts)
		resp, err := m.ubiMod.ProposalRemoveUBI(ctx, from, propOpts, opts)
		return resp, resp, err

	default:
		return nil, nil, fmt.Errorf("unknown ubi action: %s", action)
	}
}

// executeUpgrade handles upgrade module actions.
func (m *ActionMapper) executeUpgrade(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.upgradeMod == nil {
		m.upgradeMod = upgrade.New(m.client)
	}

	switch action {
	case "current-plan", "currentplan":
		result, err := m.upgradeMod.CurrentPlan(ctx)
		return result, nil, err

	case "next-plan", "nextplan":
		result, err := m.upgradeMod.NextPlan(ctx)
		return result, nil, err

	case "proposal-set-plan", "proposalsetplan":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("upgrade.proposal-set-plan requires 'from' parameter")
		}

		propOpts := &upgrade.ProposalSetPlanOpts{
			Name:        params["name"],
			Title:       params["title"],
			Description: params["description"],
			OldChainID:  params["old_chain_id"],
			NewChainID:  params["new_chain_id"],
		}
		if v := params["min_upgrade_time"]; v != "" {
			val, _ := strconv.ParseInt(v, 10, 64)
			propOpts.MinUpgradeTime = val
		}
		if v := params["max_enrollment_duration"]; v != "" {
			val, _ := strconv.ParseInt(v, 10, 64)
			propOpts.MaxEnrollmentDuration = val
		}
		if params["instate_upgrade"] == "true" {
			propOpts.InstateUpgrade = true
		}
		if params["reboot_required"] == "true" {
			propOpts.RebootRequired = true
		}

		opts := m.buildUpgradeTxOptions(txOpts)
		resp, err := m.upgradeMod.ProposalSetPlan(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-cancel-plan", "proposalcancelplan":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("upgrade.proposal-cancel-plan requires 'from' parameter")
		}

		propOpts := &upgrade.ProposalCancelPlanOpts{
			Name:        params["name"],
			Title:       params["title"],
			Description: params["description"],
		}

		opts := m.buildUpgradeTxOptions(txOpts)
		resp, err := m.upgradeMod.ProposalCancelPlan(ctx, from, propOpts, opts)
		return resp, resp, err

	default:
		return nil, nil, fmt.Errorf("unknown upgrade action: %s", action)
	}
}

// executeSpending handles spending module actions.
func (m *ActionMapper) executeSpending(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.spendingMod == nil {
		m.spendingMod = spending.New(m.client)
	}

	switch action {
	case "pool-names", "poolnames":
		result, err := m.spendingMod.PoolNames(ctx)
		return result, nil, err

	case "pool-by-name", "poolbyname", "pool":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("spending.pool-by-name requires 'name' parameter")
		}
		result, err := m.spendingMod.PoolByName(ctx, name)
		return result, nil, err

	case "pool-proposals", "poolproposals":
		poolName := params["pool_name"]
		if poolName == "" {
			poolName = params["name"]
		}
		if poolName == "" {
			return nil, nil, fmt.Errorf("spending.pool-proposals requires 'pool_name' parameter")
		}
		result, err := m.spendingMod.PoolProposals(ctx, poolName)
		return result, nil, err

	case "pools-by-account", "poolsbyaccount":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("spending.pools-by-account requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.spendingMod.PoolsByAccount(ctx, address)
		return result, nil, err

	case "claim-spending-pool", "claimspendingpool", "claim":
		from := params["from"]
		poolName := params["pool_name"]
		if poolName == "" {
			poolName = params["name"]
		}
		if from == "" || poolName == "" {
			return nil, nil, fmt.Errorf("spending.claim-spending-pool requires 'from' and 'pool_name' parameters")
		}
		opts := m.buildSpendingTxOptions(txOpts)
		resp, err := m.spendingMod.ClaimSpendingPool(ctx, from, poolName, opts)
		return resp, resp, err

	case "deposit-spending-pool", "depositspendingpool", "deposit":
		from := params["from"]
		poolName := params["pool_name"]
		if poolName == "" {
			poolName = params["name"]
		}
		amount := params["amount"]
		if from == "" || poolName == "" || amount == "" {
			return nil, nil, fmt.Errorf("spending.deposit-spending-pool requires 'from', 'pool_name', and 'amount' parameters")
		}
		opts := m.buildSpendingTxOptions(txOpts)
		resp, err := m.spendingMod.DepositSpendingPool(ctx, from, poolName, amount, opts)
		return resp, resp, err

	case "create-spending-pool", "createspendingpool", "create":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("spending.create-spending-pool requires 'from' parameter")
		}

		poolOpts := &spending.CreateSpendingPoolOpts{
			Name:          params["name"],
			Rates:         params["rates"],
			VoteQuorum:    params["vote_quorum"],
			Owners:        params["owners"],
			Beneficiaries: params["beneficiaries"],
		}
		if v := params["claim_start"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			poolOpts.ClaimStart = val
		}
		if v := params["claim_end"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			poolOpts.ClaimEnd = val
		}
		if v := params["claim_expiry"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			poolOpts.ClaimExpiry = val
		}
		if v := params["vote_period"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			poolOpts.VotePeriod = val
		}
		if v := params["vote_enactment"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			poolOpts.VoteEnactment = val
		}
		if params["dynamic_rate"] == "true" {
			poolOpts.DynamicRate = true
		}
		if v := params["dynamic_rate_period"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			poolOpts.DynamicRatePeriod = val
		}

		opts := m.buildSpendingTxOptions(txOpts)
		resp, err := m.spendingMod.CreateSpendingPool(ctx, from, poolOpts, opts)
		return resp, resp, err

	case "register-spending-pool-beneficiary", "registerspendingpoolbeneficiary", "register-beneficiary":
		from := params["from"]
		poolName := params["pool_name"]
		if poolName == "" {
			poolName = params["name"]
		}
		if from == "" || poolName == "" {
			return nil, nil, fmt.Errorf("spending.register-spending-pool-beneficiary requires 'from' and 'pool_name' parameters")
		}
		opts := m.buildSpendingTxOptions(txOpts)
		resp, err := m.spendingMod.RegisterSpendingPoolBeneficiary(ctx, from, poolName, opts)
		return resp, resp, err

	case "proposal-update-spending-pool", "proposalupdatespendingpool":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("spending.proposal-update-spending-pool requires 'from' parameter")
		}

		propOpts := &spending.ProposalUpdateSpendingPoolOpts{
			Name:                      params["name"],
			Title:                     params["title"],
			Description:               params["description"],
			Rates:                     params["rates"],
			OwnerAccounts:             params["owner_accounts"],
			OwnerRoles:                params["owner_roles"],
			BeneficiaryAccounts:       params["beneficiary_accounts"],
			BeneficiaryRoles:          params["beneficiary_roles"],
			BeneficiaryAccountWeights: params["beneficiary_account_weights"],
			BeneficiaryRoleWeights:    params["beneficiary_role_weights"],
			DynamicRatePeriod:         params["dynamic_rate_period"],
		}
		if v := params["claim_start"]; v != "" {
			val, _ := strconv.Atoi(v)
			propOpts.ClaimStart = int32(val)
		}
		if v := params["claim_end"]; v != "" {
			val, _ := strconv.Atoi(v)
			propOpts.ClaimEnd = int32(val)
		}
		if v := params["vote_quorum"]; v != "" {
			val, _ := strconv.Atoi(v)
			propOpts.VoteQuorum = int32(val)
		}
		if v := params["vote_period"]; v != "" {
			val, _ := strconv.Atoi(v)
			propOpts.VotePeriod = int32(val)
		}
		if v := params["vote_enactment"]; v != "" {
			val, _ := strconv.Atoi(v)
			propOpts.VoteEnactment = int32(val)
		}
		if params["dynamic_rate"] == "true" {
			propOpts.DynamicRate = true
		}

		opts := m.buildSpendingTxOptions(txOpts)
		resp, err := m.spendingMod.ProposalUpdateSpendingPool(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-spending-pool-distribution", "proposalspendingpooldistribution":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("spending.proposal-spending-pool-distribution requires 'from' parameter")
		}

		propOpts := &spending.ProposalSpendingPoolDistributionOpts{
			Name:        params["name"],
			Title:       params["title"],
			Description: params["description"],
		}

		opts := m.buildSpendingTxOptions(txOpts)
		resp, err := m.spendingMod.ProposalSpendingPoolDistribution(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-spending-pool-withdraw", "proposalspendingpoolwithdraw":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("spending.proposal-spending-pool-withdraw requires 'from' parameter")
		}

		propOpts := &spending.ProposalSpendingPoolWithdrawOpts{
			Name:                params["name"],
			BeneficiaryAccounts: params["beneficiary_accounts"],
			Amount:              params["amount"],
			Title:               params["title"],
			Description:         params["description"],
		}

		opts := m.buildSpendingTxOptions(txOpts)
		resp, err := m.spendingMod.ProposalSpendingPoolWithdraw(ctx, from, propOpts, opts)
		return resp, resp, err

	default:
		return nil, nil, fmt.Errorf("unknown spending action: %s", action)
	}
}

// executeCollectives handles collectives module actions.
func (m *ActionMapper) executeCollectives(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.collectivesMod == nil {
		m.collectivesMod = collectives.New(m.client)
	}

	switch action {
	case "collectives", "all", "list":
		result, err := m.collectivesMod.Collectives(ctx)
		return result, nil, err

	case "collective", "get":
		name := params["name"]
		if name == "" {
			return nil, nil, fmt.Errorf("collectives.collective requires 'name' parameter")
		}
		result, err := m.collectivesMod.Collective(ctx, name)
		return result, nil, err

	case "collectives-by-account", "collectivesbyaccount", "by-account":
		addrParam := params["address"]
		if addrParam == "" {
			return nil, nil, fmt.Errorf("collectives.collectives-by-account requires 'address' parameter")
		}
		address, err := m.resolveAddress(ctx, addrParam)
		if err != nil {
			return nil, nil, err
		}
		result, err := m.collectivesMod.CollectivesByAccount(ctx, address)
		return result, nil, err

	case "collectives-proposals", "collectivesproposals", "proposals":
		result, err := m.collectivesMod.CollectivesProposals(ctx)
		return result, nil, err

	case "create-collective", "createcollective", "create":
		from := params["from"]
		name := params["name"]
		description := params["description"]
		if from == "" || name == "" {
			return nil, nil, fmt.Errorf("collectives.create-collective requires 'from' and 'name' parameters")
		}
		opts := m.buildCollectivesTxOptions(txOpts)
		resp, err := m.collectivesMod.CreateCollective(ctx, from, name, description, opts)
		return resp, resp, err

	case "contribute-collective", "contributecollective", "contribute":
		from := params["from"]
		name := params["name"]
		bonds := params["bonds"]
		if from == "" || name == "" || bonds == "" {
			return nil, nil, fmt.Errorf("collectives.contribute-collective requires 'from', 'name', and 'bonds' parameters")
		}
		opts := m.buildCollectivesTxOptions(txOpts)
		resp, err := m.collectivesMod.ContributeCollective(ctx, from, name, bonds, opts)
		return resp, resp, err

	case "donate-collective", "donatecollective", "donate":
		from := params["from"]
		name := params["name"]
		donation := params["donation"]
		if from == "" || name == "" {
			return nil, nil, fmt.Errorf("collectives.donate-collective requires 'from' and 'name' parameters")
		}
		var locking uint64
		if v := params["locking"]; v != "" {
			locking, _ = strconv.ParseUint(v, 10, 64)
		}
		donationLock := params["donation_lock"] == "true"
		opts := m.buildCollectivesTxOptions(txOpts)
		resp, err := m.collectivesMod.DonateCollective(ctx, from, name, locking, donation, donationLock, opts)
		return resp, resp, err

	case "withdraw-collective", "withdrawcollective", "withdraw":
		from := params["from"]
		name := params["name"]
		bonds := params["bonds"]
		if from == "" || name == "" || bonds == "" {
			return nil, nil, fmt.Errorf("collectives.withdraw-collective requires 'from', 'name', and 'bonds' parameters")
		}
		opts := m.buildCollectivesTxOptions(txOpts)
		resp, err := m.collectivesMod.WithdrawCollective(ctx, from, name, bonds, opts)
		return resp, resp, err

	case "proposal-collective-update", "proposalcollectiveupdate":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("collectives.proposal-collective-update requires 'from' parameter")
		}

		propOpts := &collectives.ProposalCollectiveUpdateOpts{
			Title:                 params["title"],
			Description:           params["description"],
			CollectiveName:        params["collective_name"],
			CollectiveDescription: params["collective_description"],
			CollectiveStatus:      params["collective_status"],
		}

		opts := m.buildCollectivesTxOptions(txOpts)
		resp, err := m.collectivesMod.ProposalCollectiveUpdate(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-remove-collective", "proposalremovecollective":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("collectives.proposal-remove-collective requires 'from' parameter")
		}

		propOpts := &collectives.ProposalRemoveCollectiveOpts{
			Title:          params["title"],
			Description:    params["description"],
			CollectiveName: params["collective_name"],
		}

		opts := m.buildCollectivesTxOptions(txOpts)
		resp, err := m.collectivesMod.ProposalRemoveCollective(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-send-donation", "proposalsenddonation":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("collectives.proposal-send-donation requires 'from' parameter")
		}

		propOpts := &collectives.ProposalSendDonationOpts{
			Title:          params["title"],
			Description:    params["description"],
			CollectiveName: params["collective_name"],
			Address:        params["address"],
			Amounts:        params["amounts"],
		}

		opts := m.buildCollectivesTxOptions(txOpts)
		resp, err := m.collectivesMod.ProposalSendDonation(ctx, from, propOpts, opts)
		return resp, resp, err

	default:
		return nil, nil, fmt.Errorf("unknown collectives action: %s", action)
	}
}

// executeBasket handles basket module actions.
func (m *ActionMapper) executeBasket(ctx context.Context, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	if m.basketMod == nil {
		m.basketMod = basket.New(m.client)
	}

	switch action {
	case "token-baskets", "tokenbaskets", "baskets":
		tokens := params["tokens"]
		derivativesOnly := params["derivatives_only"] == "true"
		result, err := m.basketMod.TokenBaskets(ctx, tokens, derivativesOnly)
		return result, nil, err

	case "token-basket-by-id", "tokenbasketbyid", "by-id":
		id := params["id"]
		if id == "" {
			return nil, nil, fmt.Errorf("basket.token-basket-by-id requires 'id' parameter")
		}
		result, err := m.basketMod.TokenBasketByID(ctx, id)
		return result, nil, err

	case "token-basket-by-denom", "tokenbasketbydenom", "by-denom":
		denom := params["denom"]
		if denom == "" {
			return nil, nil, fmt.Errorf("basket.token-basket-by-denom requires 'denom' parameter")
		}
		result, err := m.basketMod.TokenBasketByDenom(ctx, denom)
		return result, nil, err

	case "historical-mints", "historicalmints":
		basketID := params["basket_id"]
		if basketID == "" {
			basketID = params["id"]
		}
		if basketID == "" {
			return nil, nil, fmt.Errorf("basket.historical-mints requires 'basket_id' parameter")
		}
		result, err := m.basketMod.HistoricalMints(ctx, basketID)
		return result, nil, err

	case "historical-burns", "historicalburns":
		basketID := params["basket_id"]
		if basketID == "" {
			basketID = params["id"]
		}
		if basketID == "" {
			return nil, nil, fmt.Errorf("basket.historical-burns requires 'basket_id' parameter")
		}
		result, err := m.basketMod.HistoricalBurns(ctx, basketID)
		return result, nil, err

	case "historical-swaps", "historicalswaps":
		basketID := params["basket_id"]
		if basketID == "" {
			basketID = params["id"]
		}
		if basketID == "" {
			return nil, nil, fmt.Errorf("basket.historical-swaps requires 'basket_id' parameter")
		}
		result, err := m.basketMod.HistoricalSwaps(ctx, basketID)
		return result, nil, err

	case "mint-basket-tokens", "mintbaskettokens", "mint":
		from := params["from"]
		basketID := params["basket_id"]
		if basketID == "" {
			basketID = params["id"]
		}
		depositCoins := params["deposit_coins"]
		if depositCoins == "" {
			depositCoins = params["coins"]
		}
		if from == "" || basketID == "" || depositCoins == "" {
			return nil, nil, fmt.Errorf("basket.mint-basket-tokens requires 'from', 'basket_id', and 'deposit_coins' parameters")
		}
		opts := m.buildBasketTxOptions(txOpts)
		resp, err := m.basketMod.MintBasketTokens(ctx, from, basketID, depositCoins, opts)
		return resp, resp, err

	case "burn-basket-tokens", "burnbaskettokens", "burn":
		from := params["from"]
		basketID := params["basket_id"]
		if basketID == "" {
			basketID = params["id"]
		}
		burnAmount := params["burn_amount"]
		if burnAmount == "" {
			burnAmount = params["amount"]
		}
		if from == "" || basketID == "" || burnAmount == "" {
			return nil, nil, fmt.Errorf("basket.burn-basket-tokens requires 'from', 'basket_id', and 'burn_amount' parameters")
		}
		opts := m.buildBasketTxOptions(txOpts)
		resp, err := m.basketMod.BurnBasketTokens(ctx, from, basketID, burnAmount, opts)
		return resp, resp, err

	case "swap-basket-tokens", "swapbaskettokens", "swap":
		from := params["from"]
		basketID := params["basket_id"]
		if basketID == "" {
			basketID = params["id"]
		}
		swapIn := params["swap_in"]
		swapOut := params["swap_out"]
		if from == "" || basketID == "" || swapIn == "" || swapOut == "" {
			return nil, nil, fmt.Errorf("basket.swap-basket-tokens requires 'from', 'basket_id', 'swap_in', and 'swap_out' parameters")
		}
		opts := m.buildBasketTxOptions(txOpts)
		resp, err := m.basketMod.SwapBasketTokens(ctx, from, basketID, swapIn, swapOut, opts)
		return resp, resp, err

	case "basket-claim-rewards", "basketclaimrewards", "claim-rewards":
		from := params["from"]
		basketID := params["basket_id"]
		if basketID == "" {
			basketID = params["id"]
		}
		if from == "" || basketID == "" {
			return nil, nil, fmt.Errorf("basket.basket-claim-rewards requires 'from' and 'basket_id' parameters")
		}
		opts := m.buildBasketTxOptions(txOpts)
		resp, err := m.basketMod.BasketClaimRewards(ctx, from, basketID, opts)
		return resp, resp, err

	case "disable-basket-deposits", "disablebasketdeposits":
		from := params["from"]
		basketID := params["basket_id"]
		if basketID == "" {
			basketID = params["id"]
		}
		if from == "" || basketID == "" {
			return nil, nil, fmt.Errorf("basket.disable-basket-deposits requires 'from' and 'basket_id' parameters")
		}
		disabled := params["disabled"] == "true"
		opts := m.buildBasketTxOptions(txOpts)
		resp, err := m.basketMod.DisableBasketDeposits(ctx, from, basketID, disabled, opts)
		return resp, resp, err

	case "disable-basket-withdraws", "disablebasketwithraws":
		from := params["from"]
		basketID := params["basket_id"]
		if basketID == "" {
			basketID = params["id"]
		}
		if from == "" || basketID == "" {
			return nil, nil, fmt.Errorf("basket.disable-basket-withdraws requires 'from' and 'basket_id' parameters")
		}
		disabled := params["disabled"] == "true"
		opts := m.buildBasketTxOptions(txOpts)
		resp, err := m.basketMod.DisableBasketWithdraws(ctx, from, basketID, disabled, opts)
		return resp, resp, err

	case "disable-basket-swaps", "disablebasketswaps":
		from := params["from"]
		basketID := params["basket_id"]
		if basketID == "" {
			basketID = params["id"]
		}
		if from == "" || basketID == "" {
			return nil, nil, fmt.Errorf("basket.disable-basket-swaps requires 'from' and 'basket_id' parameters")
		}
		disabled := params["disabled"] == "true"
		opts := m.buildBasketTxOptions(txOpts)
		resp, err := m.basketMod.DisableBasketSwaps(ctx, from, basketID, disabled, opts)
		return resp, resp, err

	case "proposal-create-basket", "proposalcreatebasket":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("basket.proposal-create-basket requires 'from' parameter")
		}

		propOpts := &basket.ProposalCreateBasketOpts{
			BasketSuffix:      params["basket_suffix"],
			BasketDescription: params["basket_description"],
			BasketTokens:      params["basket_tokens"],
			TokensCap:         params["tokens_cap"],
			Title:             params["title"],
			Description:       params["description"],
			MintsMin:          params["mints_min"],
			MintsMax:          params["mints_max"],
			BurnsMin:          params["burns_min"],
			BurnsMax:          params["burns_max"],
			SwapsMin:          params["swaps_min"],
			SwapsMax:          params["swaps_max"],
			SwapFee:           params["swap_fee"],
			SlippageFeeMin:    params["slippage_fee_min"],
		}
		if params["mints_disabled"] == "true" {
			propOpts.MintsDisabled = true
		}
		if params["burns_disabled"] == "true" {
			propOpts.BurnsDisabled = true
		}
		if params["swaps_disabled"] == "true" {
			propOpts.SwapsDisabled = true
		}
		if v := params["limits_period"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			propOpts.LimitsPeriod = val
		}

		opts := m.buildBasketTxOptions(txOpts)
		resp, err := m.basketMod.ProposalCreateBasket(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-edit-basket", "proposaleditbasket":
		from := params["from"]
		if from == "" {
			return nil, nil, fmt.Errorf("basket.proposal-edit-basket requires 'from' parameter")
		}

		propOpts := &basket.ProposalEditBasketOpts{
			BasketSuffix:      params["basket_suffix"],
			BasketDescription: params["basket_description"],
			BasketTokens:      params["basket_tokens"],
			TokensCap:         params["tokens_cap"],
			Title:             params["title"],
			Description:       params["description"],
			MintsMin:          params["mints_min"],
			MintsMax:          params["mints_max"],
			BurnsMin:          params["burns_min"],
			BurnsMax:          params["burns_max"],
			SwapsMin:          params["swaps_min"],
			SwapsMax:          params["swaps_max"],
			SwapFee:           params["swap_fee"],
			SlippageFeeMin:    params["slippage_fee_min"],
		}
		if v := params["basket_id"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			propOpts.BasketID = val
		}
		if params["mints_disabled"] == "true" {
			propOpts.MintsDisabled = true
		}
		if params["burns_disabled"] == "true" {
			propOpts.BurnsDisabled = true
		}
		if params["swaps_disabled"] == "true" {
			propOpts.SwapsDisabled = true
		}
		if v := params["limits_period"]; v != "" {
			val, _ := strconv.ParseUint(v, 10, 64)
			propOpts.LimitsPeriod = val
		}

		opts := m.buildBasketTxOptions(txOpts)
		resp, err := m.basketMod.ProposalEditBasket(ctx, from, propOpts, opts)
		return resp, resp, err

	case "proposal-basket-withdraw-surplus", "proposalbasketwithdrawsurplus", "proposal-withdraw-surplus":
		from := params["from"]
		basketIDs := params["basket_ids"]
		withdrawTarget := params["withdraw_target"]
		if from == "" || basketIDs == "" || withdrawTarget == "" {
			return nil, nil, fmt.Errorf("basket.proposal-basket-withdraw-surplus requires 'from', 'basket_ids', and 'withdraw_target' parameters")
		}

		propOpts := &basket.ProposalWithdrawSurplusOpts{
			Title:       params["title"],
			Description: params["description"],
		}

		opts := m.buildBasketTxOptions(txOpts)
		resp, err := m.basketMod.ProposalWithdrawSurplus(ctx, from, basketIDs, withdrawTarget, propOpts, opts)
		return resp, resp, err

	default:
		return nil, nil, fmt.Errorf("unknown basket action: %s", action)
	}
}

// executeGeneric handles modules without specific mappings using raw SDK calls.
func (m *ActionMapper) executeGeneric(ctx context.Context, module, action string, params map[string]string, txOpts *StepTxOptions) (interface{}, *sdk.TxResponse, error) {
	// Determine if this is a query or transaction
	stepType := GetStepType(&Step{Module: module, Action: action})

	if stepType == StepTypeQuery {
		// Build query request
		req := &sdk.QueryRequest{
			Module:   module,
			Endpoint: action,
			Params:   params,
		}

		resp, err := m.client.Query(ctx, req)
		if err != nil {
			return nil, nil, err
		}

		// Try to parse JSON response
		var result interface{}
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			return string(resp.Data), nil, nil
		}
		return result, nil, nil
	}

	// Build transaction request
	args := make([]string, 0)
	flags := make(map[string]string)

	// Separate positional args from flags
	for k, v := range params {
		if strings.HasPrefix(k, "_arg") {
			args = append(args, v)
		} else {
			flags[k] = v
		}
	}

	// Add tx options as flags
	if txOpts != nil {
		if txOpts.Fees != "" {
			flags["fees"] = txOpts.Fees
		}
		if txOpts.Gas != "" {
			flags["gas"] = txOpts.Gas
		}
		if txOpts.Memo != "" {
			flags["memo"] = txOpts.Memo
		}
		if txOpts.BroadcastMode != "" {
			flags["broadcast-mode"] = txOpts.BroadcastMode
		}
	}

	signer := params["from"]
	if signer == "" {
		signer = params["signer"]
	}

	req := &sdk.TxRequest{
		Module:           module,
		Action:           action,
		Args:             args,
		Flags:            flags,
		Signer:           signer,
		SkipConfirmation: true,
	}

	resp, err := m.client.Tx(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return resp, resp, nil
}

// Helper functions to build module-specific options

func (m *ActionMapper) buildSendOptions(txOpts *StepTxOptions) *bank.SendOptions {
	if txOpts == nil {
		return nil
	}
	return &bank.SendOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}

func (m *ActionMapper) buildGovTxOptions(txOpts *StepTxOptions) *gov.TxOptions {
	if txOpts == nil {
		return nil
	}
	return &gov.TxOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}

func (m *ActionMapper) buildMultistakingTxOptions(txOpts *StepTxOptions) *multistaking.TxOptions {
	if txOpts == nil {
		return nil
	}
	return &multistaking.TxOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}

func (m *ActionMapper) buildBridgeTxOptions(txOpts *StepTxOptions) *bridge.TxOptions {
	if txOpts == nil {
		return nil
	}
	return &bridge.TxOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}

func (m *ActionMapper) buildUbiTxOptions(txOpts *StepTxOptions) *ubi.TxOptions {
	if txOpts == nil {
		return nil
	}
	return &ubi.TxOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}

func (m *ActionMapper) buildUpgradeTxOptions(txOpts *StepTxOptions) *upgrade.TxOptions {
	if txOpts == nil {
		return nil
	}
	return &upgrade.TxOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}

func (m *ActionMapper) buildSpendingTxOptions(txOpts *StepTxOptions) *spending.TxOptions {
	if txOpts == nil {
		return nil
	}
	return &spending.TxOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}

func (m *ActionMapper) buildCollectivesTxOptions(txOpts *StepTxOptions) *collectives.TxOptions {
	if txOpts == nil {
		return nil
	}
	return &collectives.TxOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}

func (m *ActionMapper) buildBasketTxOptions(txOpts *StepTxOptions) *basket.TxOptions {
	if txOpts == nil {
		return nil
	}
	return &basket.TxOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}

func (m *ActionMapper) buildStakingTxOptions(txOpts *StepTxOptions) *staking.TxOptions {
	if txOpts == nil {
		return nil
	}
	return &staking.TxOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}

func (m *ActionMapper) buildTokensTxOptions(txOpts *StepTxOptions) *tokens.TxOptions {
	if txOpts == nil {
		return nil
	}
	return &tokens.TxOptions{
		Fees:          txOpts.Fees,
		Gas:           txOpts.Gas,
		Memo:          txOpts.Memo,
		BroadcastMode: txOpts.BroadcastMode,
	}
}
