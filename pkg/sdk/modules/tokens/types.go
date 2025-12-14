package tokens

// TokenRate represents a token rate configuration.
type TokenRate struct {
	Denom             string `json:"denom"`
	TokenType         string `json:"token_type,omitempty"`
	FeeRate           string `json:"fee_rate"`
	FeeEnabled        bool   `json:"fee_enabled"`
	Supply            string `json:"supply,omitempty"`
	SupplyCap         string `json:"supply_cap,omitempty"`
	StakeCap          string `json:"stake_cap,omitempty"`
	StakeMin          string `json:"stake_min,omitempty"`
	StakeEnabled      bool   `json:"stake_enabled"`
	Inactive          bool   `json:"inactive"`
	Symbol            string `json:"symbol,omitempty"`
	Name              string `json:"name,omitempty"`
	Icon              string `json:"icon,omitempty"`
	Decimals          int    `json:"decimals"`
	Description       string `json:"description,omitempty"`
	Website           string `json:"website,omitempty"`
	Social            string `json:"social,omitempty"`
	Holders           string `json:"holders,omitempty"`
	MintingFee        string `json:"minting_fee,omitempty"`
	Owner             string `json:"owner,omitempty"`
	OwnerEditDisabled bool   `json:"owner_edit_disabled"`
	NFTMetadata       string `json:"nft_metadata,omitempty"`
	NFTHash           string `json:"nft_hash,omitempty"`
}

// TokenRateWithSupply represents a token rate with supply info.
type TokenRateWithSupply struct {
	Data   TokenRate `json:"data"`
	Supply Coin      `json:"supply,omitempty"`
}

// Coin represents a token amount.
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// AllRatesResponse contains the all-rates query response.
type AllRatesResponse struct {
	Data []TokenRateWithSupply `json:"data"`
}

// TokenBlackWhites contains the whitelisted and blacklisted tokens.
type TokenBlackWhites struct {
	Whitelisted []string `json:"whitelisted,omitempty"`
	Blacklisted []string `json:"blacklisted,omitempty"`
}
