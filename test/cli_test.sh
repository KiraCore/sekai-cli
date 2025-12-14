#!/bin/bash

# CLI Test Suite for sekai-cli
# Tests all implemented commands

CLI="./build/sekai-cli"
ADDRESS="kira1cw0wz6x9wy8wvw30q8qsxppgzqrr5qu846cut5"
VAL_ADDRESS="kiravaloper1cw0wz6x9wy8wvw30q8qsxppgzqrr5qu8zp4dc4"
KEY_NAME="genesis"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PASSED=0
FAILED=0
SKIPPED=0

run_test() {
    local name="$1"
    local cmd="$2"
    echo -n "Testing: $name... "
    if output=$(eval "$cmd" 2>&1); then
        echo -e "${GREEN}PASS${NC}"
        ((PASSED++))
    else
        echo -e "${RED}FAIL${NC}"
        echo "  Command: $cmd"
        echo "  Output: $output" | head -3
        ((FAILED++))
    fi
}

skip_test() {
    local name="$1"
    local reason="$2"
    echo -e "Testing: $name... ${YELLOW}SKIP${NC} ($reason)"
    ((SKIPPED++))
}

echo "============================================"
echo "SEKAI-CLI Full Test Suite"
echo "============================================"
echo ""

# Build first
echo "Building CLI..."
make docker-build 2>&1 | tail -1
echo ""

# =============================================
echo -e "${CYAN}=== BASIC COMMANDS ===${NC}"
# =============================================
run_test "version" "$CLI version"
run_test "help" "$CLI --help"

# =============================================
echo -e "\n${CYAN}=== STATUS ===${NC}"
# =============================================
run_test "status" "$CLI status"
run_test "status -o json" "$CLI -o json status"

# =============================================
echo -e "\n${CYAN}=== KEYS COMMANDS ===${NC}"
# =============================================
run_test "keys list" "$CLI keys list"
run_test "keys list -o json" "$CLI -o json keys list"
run_test "keys show" "$CLI keys show $KEY_NAME"
run_test "keys show -o json" "$CLI -o json keys show $KEY_NAME"
skip_test "keys add" "Would modify state"
skip_test "keys delete" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== QUERY AUTH ===${NC}"
# =============================================
run_test "query auth account" "$CLI query auth account $ADDRESS"
run_test "query auth accounts" "$CLI query auth accounts"
run_test "query auth module-accounts" "$CLI query auth module-accounts"
run_test "query auth params" "$CLI query auth params"
run_test "query auth address-by-acc-num" "$CLI query auth address-by-acc-num 0"

# =============================================
echo -e "\n${CYAN}=== QUERY BANK ===${NC}"
# =============================================
run_test "query bank balances" "$CLI query bank balances $ADDRESS"
run_test "query bank balances -o json" "$CLI -o json query bank balances $ADDRESS"
run_test "query bank total" "$CLI query bank total"

# =============================================
echo -e "\n${CYAN}=== QUERY TOKENS ===${NC}"
# =============================================
run_test "query tokens all-rates" "$CLI query tokens all-rates"
run_test "query tokens rate" "$CLI query tokens rate ukex"
run_test "query tokens rates-by-denom" "$CLI query tokens rates-by-denom ukex"
run_test "query tokens token-black-whites" "$CLI query tokens token-black-whites"

# =============================================
echo -e "\n${CYAN}=== QUERY CUSTOMGOV ===${NC}"
# =============================================
run_test "query customgov network-properties" "$CLI query customgov network-properties"
run_test "query customgov proposals" "$CLI query customgov proposals"
run_test "query customgov proposal" "$CLI query customgov proposal 1"
run_test "query customgov councilors" "$CLI query customgov councilors"
run_test "query customgov all-roles" "$CLI query customgov all-roles"
run_test "query customgov role" "$CLI query customgov role 1"
run_test "query customgov roles" "$CLI query customgov roles $ADDRESS"
run_test "query customgov permissions" "$CLI query customgov permissions $ADDRESS"
run_test "query customgov all-execution-fees" "$CLI query customgov all-execution-fees"
run_test "query customgov identity-records" "$CLI query customgov identity-records"
run_test "query customgov identity-records-by-addr" "$CLI query customgov identity-records-by-addr $ADDRESS"
run_test "query customgov data-registry-keys" "$CLI query customgov data-registry-keys"
run_test "query customgov poor-network-messages" "$CLI query customgov poor-network-messages"
run_test "query customgov custom-prefixes" "$CLI query customgov custom-prefixes"
run_test "query customgov all-proposal-durations" "$CLI query customgov all-proposal-durations"
run_test "query customgov proposal-duration" "$CLI query customgov proposal-duration SetNetworkProperty"
run_test "query customgov non-councilors" "$CLI query customgov non-councilors"
run_test "query customgov whitelisted-permission-addresses" "$CLI query customgov whitelisted-permission-addresses 1"
run_test "query customgov blacklisted-permission-addresses" "$CLI query customgov blacklisted-permission-addresses 7"
run_test "query customgov proposer-voters-count" "$CLI query customgov proposer-voters-count"
run_test "query customgov all-identity-record-verify-requests" "$CLI query customgov all-identity-record-verify-requests"
run_test "query customgov polls" "$CLI query customgov polls $ADDRESS"

# =============================================
echo -e "\n${CYAN}=== QUERY CUSTOMSLASHING ===${NC}"
# =============================================
run_test "query customslashing signing-infos" "$CLI query customslashing signing-infos"
run_test "query customslashing active-staking-pools" "$CLI query customslashing active-staking-pools"
run_test "query customslashing inactive-staking-pools" "$CLI query customslashing inactive-staking-pools"
run_test "query customslashing slashed-staking-pools" "$CLI query customslashing slashed-staking-pools"
run_test "query customslashing slash-proposals" "$CLI query customslashing slash-proposals"

# =============================================
echo -e "\n${CYAN}=== QUERY CUSTOMSTAKING ===${NC}"
# =============================================
run_test "query customstaking validators" "$CLI query customstaking validators"
run_test "query customstaking validator" "$CLI query customstaking validator $VAL_ADDRESS"

# =============================================
echo -e "\n${CYAN}=== QUERY MULTISTAKING ===${NC}"
# =============================================
run_test "query multistaking pools" "$CLI query multistaking pools"

# =============================================
echo -e "\n${CYAN}=== QUERY SPENDING ===${NC}"
# =============================================
run_test "query spending pool-names" "$CLI query spending pool-names"

# =============================================
echo -e "\n${CYAN}=== QUERY UBI ===${NC}"
# =============================================
run_test "query ubi ubi-records" "$CLI query ubi ubi-records"

# =============================================
echo -e "\n${CYAN}=== QUERY UPGRADE ===${NC}"
# =============================================
run_test "query upgrade current-plan" "$CLI query upgrade current-plan"
run_test "query upgrade next-plan" "$CLI query upgrade next-plan"

# =============================================
echo -e "\n${CYAN}=== QUERY DISTRIBUTOR ===${NC}"
# =============================================
run_test "query distributor fees-treasury" "$CLI query distributor fees-treasury"
run_test "query distributor snapshot-period" "$CLI query distributor snapshot-period"

# =============================================
echo -e "\n${CYAN}=== QUERY BASKET ===${NC}"
# =============================================
run_test "query basket token-baskets" "$CLI query basket token-baskets"

# =============================================
echo -e "\n${CYAN}=== QUERY COLLECTIVES ===${NC}"
# =============================================
run_test "query collectives collectives" "$CLI query collectives collectives"

# =============================================
echo -e "\n${CYAN}=== QUERY CUSTODY ===${NC}"
# =============================================
run_test "query custody get" "$CLI query custody get $ADDRESS"
run_test "query custody custodians" "$CLI query custody custodians $ADDRESS"
run_test "query custody whitelist" "$CLI query custody whitelist $ADDRESS"
run_test "query custody limits" "$CLI query custody limits $ADDRESS"

# =============================================
echo -e "\n${CYAN}=== QUERY BRIDGE ===${NC}"
# =============================================
run_test "query bridge get_cosmos_ethereum" "$CLI query bridge get_cosmos_ethereum $ADDRESS"
run_test "query bridge get_ethereum_cosmos" "$CLI query bridge get_ethereum_cosmos $ADDRESS"

# =============================================
echo -e "\n${CYAN}=== QUERY LAYER2 ===${NC}"
# =============================================
run_test "query layer2 all-dapps" "$CLI query layer2 all-dapps"
run_test "query layer2 transfer-dapps" "$CLI query layer2 transfer-dapps"

# =============================================
echo -e "\n${CYAN}=== QUERY RECOVERY ===${NC}"
# =============================================
run_test "query recovery recovery-token" "$CLI query recovery recovery-token $ADDRESS"

# =============================================
echo -e "\n${CYAN}=== TX BANK ===${NC}"
# =============================================
skip_test "tx bank send" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX CUSTOMGOV COUNCILOR ===${NC}"
# =============================================
skip_test "tx customgov councilor claim-seat" "Would modify state"
skip_test "tx customgov councilor activate" "Would modify state"
skip_test "tx customgov councilor pause" "Would modify state"
skip_test "tx customgov councilor unpause" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX CUSTOMGOV PERMISSION ===${NC}"
# =============================================
skip_test "tx customgov permission whitelist" "Would modify state"
skip_test "tx customgov permission blacklist" "Would modify state"
skip_test "tx customgov permission remove-whitelisted" "Would modify state"
skip_test "tx customgov permission remove-blacklisted" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX CUSTOMGOV ROLE ===${NC}"
# =============================================
skip_test "tx customgov role create" "Would modify state"
skip_test "tx customgov role assign" "Would modify state"
skip_test "tx customgov role unassign" "Would modify state"
skip_test "tx customgov role whitelist-permission" "Would modify state"
skip_test "tx customgov role blacklist-permission" "Would modify state"
skip_test "tx customgov role remove-whitelisted-permission" "Would modify state"
skip_test "tx customgov role remove-blacklisted-permission" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX CUSTOMGOV POLL ===${NC}"
# =============================================
skip_test "tx customgov poll create" "Would modify state"
skip_test "tx customgov poll vote" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX CUSTOMGOV IDENTITY ===${NC}"
# =============================================
skip_test "tx customgov register-identity-records" "Would modify state"
skip_test "tx customgov delete-identity-records" "Would modify state"
skip_test "tx customgov request-identity-record-verify" "Would modify state"
skip_test "tx customgov handle-identity-records-verify-request" "Would modify state"
skip_test "tx customgov cancel-identity-records-verify-request" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX CUSTOMGOV SUDO ===${NC}"
# =============================================
skip_test "tx customgov set-network-properties" "Would modify state"
skip_test "tx customgov set-execution-fee" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX CUSTOMGOV PROPOSAL ===${NC}"
# =============================================
skip_test "tx customgov proposal vote" "Would modify state"
skip_test "tx customgov proposal assign-role" "Would modify state"
skip_test "tx customgov proposal unassign-role" "Would modify state"
skip_test "tx customgov proposal set-network-property" "Would modify state"
skip_test "tx customgov proposal set-poor-network-msgs" "Would modify state"
skip_test "tx customgov proposal set-proposal-durations" "Would modify state"
skip_test "tx customgov proposal upsert-data-registry" "Would modify state"
skip_test "tx customgov proposal set-execution-fees" "Would modify state"
skip_test "tx customgov proposal jail-councilor" "Would modify state"
skip_test "tx customgov proposal reset-whole-councilor-rank" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX CUSTOMGOV PROPOSAL ACCOUNT ===${NC}"
# =============================================
skip_test "tx customgov proposal account whitelist-permission" "Would modify state"
skip_test "tx customgov proposal account blacklist-permission" "Would modify state"
skip_test "tx customgov proposal account remove-whitelisted-permission" "Would modify state"
skip_test "tx customgov proposal account remove-blacklisted-permission" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX CUSTOMGOV PROPOSAL ROLE ===${NC}"
# =============================================
skip_test "tx customgov proposal role create" "Would modify state"
skip_test "tx customgov proposal role remove" "Would modify state"
skip_test "tx customgov proposal role whitelist-permission" "Would modify state"
skip_test "tx customgov proposal role blacklist-permission" "Would modify state"
skip_test "tx customgov proposal role remove-whitelisted-permission" "Would modify state"
skip_test "tx customgov proposal role remove-blacklisted-permission" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX CUSTOMSTAKING ===${NC}"
# =============================================
skip_test "tx customstaking claim-validator-seat" "Would modify state"
skip_test "tx customstaking proposal-unjail-validator" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX MULTISTAKING ===${NC}"
# =============================================
skip_test "tx multistaking delegate" "Would modify state"
skip_test "tx multistaking undelegate" "Would modify state"
skip_test "tx multistaking claim-rewards" "Would modify state"
skip_test "tx multistaking claim-undelegation" "Would modify state"
skip_test "tx multistaking claim-matured-undelegations" "Would modify state"
skip_test "tx multistaking register-delegator" "Would modify state"
skip_test "tx multistaking set-compound-info" "Would modify state"
skip_test "tx multistaking upsert-staking-pool" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX SPENDING ===${NC}"
# =============================================
skip_test "tx spending claim-spending-pool" "Would modify state"
skip_test "tx spending deposit-spending-pool" "Would modify state"
skip_test "tx spending create-spending-pool" "Would modify state"
skip_test "tx spending register-spending-pool-beneficiary" "Would modify state"
skip_test "tx spending proposal-spending-pool-distribution" "Would modify state"
skip_test "tx spending proposal-spending-pool-withdraw" "Would modify state"
skip_test "tx spending proposal-update-spending-pool" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX TOKENS ===${NC}"
# =============================================
skip_test "tx tokens upsert-rate" "Would modify state"
skip_test "tx tokens proposal-upsert-rate" "Would modify state"
skip_test "tx tokens proposal-update-tokens-blackwhite" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX UBI ===${NC}"
# =============================================
skip_test "tx ubi proposal-upsert-ubi" "Would modify state"
skip_test "tx ubi proposal-remove-ubi" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX UPGRADE ===${NC}"
# =============================================
skip_test "tx upgrade proposal-set-plan" "Would modify state"
skip_test "tx upgrade proposal-cancel-plan" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX BASKET ===${NC}"
# =============================================
skip_test "tx basket mint-basket-tokens" "Would modify state"
skip_test "tx basket burn-basket-tokens" "Would modify state"
skip_test "tx basket swap-basket-tokens" "Would modify state"
skip_test "tx basket basket-claim-rewards" "Would modify state"
skip_test "tx basket disable-basket-deposits" "Would modify state"
skip_test "tx basket disable-basket-withdraws" "Would modify state"
skip_test "tx basket disable-basket-swaps" "Would modify state"
skip_test "tx basket proposal-create-basket" "Would modify state"
skip_test "tx basket proposal-edit-basket" "Would modify state"
skip_test "tx basket proposal-basket-withdraw-surplus" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX COLLECTIVES ===${NC}"
# =============================================
skip_test "tx collectives create-collective" "Would modify state"
skip_test "tx collectives contribute-collective" "Would modify state"
skip_test "tx collectives withdraw-collective" "Would modify state"
skip_test "tx collectives donate-collective" "Would modify state"
skip_test "tx collectives proposal-collective-update" "Would modify state"
skip_test "tx collectives proposal-remove-collective" "Would modify state"
skip_test "tx collectives proposal-send-donation" "Would modify state"

# =============================================
echo -e "\n${CYAN}=== TX BRIDGE ===${NC}"
# =============================================
skip_test "tx bridge change-cosmos-ethereum" "Would modify state"
skip_test "tx bridge change-ethereum-cosmos" "Would modify state"

# =============================================
echo ""
echo "============================================"
echo "TEST SUMMARY"
echo "============================================"
echo -e "Passed:   ${GREEN}$PASSED${NC}"
echo -e "Failed:   ${RED}$FAILED${NC}"
echo -e "Skipped:  ${YELLOW}$SKIPPED${NC} (TX commands that would modify state)"
echo ""
TOTAL=$((PASSED + FAILED + SKIPPED))
echo "Total commands tested: $TOTAL"
echo ""

if [ $FAILED -gt 0 ]; then
    exit 1
fi
