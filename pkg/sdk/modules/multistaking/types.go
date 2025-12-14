package multistaking

// StakingPool represents a staking pool.
type StakingPool struct {
	ID                 string `json:"id"`
	Validator          string `json:"validator,omitempty"`
	Enabled            bool   `json:"enabled"`
	Slashed            string `json:"slashed,omitempty"`
	TotalStakingTokens []Coin `json:"total_staking_tokens,omitempty"`
	TotalShareTokens   []Coin `json:"total_share_tokens,omitempty"`
	TotalRewards       []Coin `json:"total_rewards,omitempty"`
	Commission         string `json:"commission,omitempty"`
}

// Coin represents a token amount.
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// PoolsResponse contains the pools query response.
type PoolsResponse struct {
	Pools []StakingPool `json:"pools"`
}

// Undelegation represents an undelegation request.
type Undelegation struct {
	ID         string `json:"id"`
	Delegator  string `json:"delegator"`
	Pool       string `json:"pool,omitempty"`
	Amount     string `json:"amount"`
	Expiration string `json:"expiration,omitempty"`
}

// UndelegationsResponse contains the undelegations query response.
type UndelegationsResponse struct {
	Undelegations []Undelegation `json:"undelegations"`
}

// OutstandingRewards represents outstanding rewards.
type OutstandingRewards struct {
	Rewards []Reward `json:"rewards"`
}

// Reward represents a single reward entry.
type Reward struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// CompoundInfo represents compound information.
type CompoundInfo struct {
	AllowCompound bool `json:"all_compound"`
}

// StakingPoolDelegator represents a delegator in a staking pool.
type StakingPoolDelegator struct {
	Delegator string `json:"delegator"`
	Pool      string `json:"pool,omitempty"`
	Amount    string `json:"amount"`
}
