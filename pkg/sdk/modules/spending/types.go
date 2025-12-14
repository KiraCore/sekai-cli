package spending

// SpendingPool represents a spending pool.
type SpendingPool struct {
	Name                    string            `json:"name"`
	ClaimStart              string            `json:"claim_start"`
	ClaimEnd                string            `json:"claim_end"`
	ClaimExpiry             string            `json:"claim_expiry"`
	Rates                   []Rate            `json:"rates"`
	VoteQuorum              string            `json:"vote_quorum"`
	VotePeriod              string            `json:"vote_period"`
	VoteEnactment           string            `json:"vote_enactment"`
	Owners                  PoolOwners        `json:"owners"`
	Beneficiaries           PoolBeneficiaries `json:"beneficiaries"`
	Balances                []Balance         `json:"balances"`
	DynamicRate             bool              `json:"dynamic_rate"`
	DynamicRatePeriod       string            `json:"dynamic_rate_period"`
	LastDynamicRateCalcTime string            `json:"last_dynamic_rate_calc_time"`
}

// Rate represents a rate entry.
type Rate struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// Balance represents a balance entry.
type Balance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// PoolOwners represents pool owners.
type PoolOwners struct {
	OwnerRoles    []string `json:"owner_roles"`
	OwnerAccounts []string `json:"owner_accounts"`
}

// PoolBeneficiaries represents pool beneficiaries.
type PoolBeneficiaries struct {
	Roles    []BeneficiaryRole    `json:"roles"`
	Accounts []BeneficiaryAccount `json:"accounts"`
}

// BeneficiaryRole represents a beneficiary role.
type BeneficiaryRole struct {
	Role   string `json:"role"`
	Weight string `json:"weight"`
}

// BeneficiaryAccount represents a beneficiary account.
type BeneficiaryAccount struct {
	Address string `json:"address"`
	Weight  string `json:"weight"`
}

// PoolNamesResponse contains pool names query response.
type PoolNamesResponse struct {
	Names []string `json:"names"`
}

// PoolByNameResponse contains pool by name query response.
type PoolByNameResponse struct {
	Pool SpendingPool `json:"pool"`
}

// PoolsByAccountResponse contains pools by account query response.
type PoolsByAccountResponse struct {
	Names []string `json:"names"`
}
