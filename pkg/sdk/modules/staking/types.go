package staking

// Validator represents a validator.
type Validator struct {
	Address   string           `json:"address,omitempty"`
	ValKey    string           `json:"valkey,omitempty"`
	ValKeyAlt string           `json:"val_key,omitempty"` // Alternative field name
	PubKey    any              `json:"pubkey,omitempty"`
	PubKeyAlt any              `json:"pub_key,omitempty"` // Alternative field name
	Proposer  string           `json:"proposer,omitempty"`
	Moniker   string           `json:"moniker,omitempty"`
	Status    string           `json:"status"`
	Rank      string           `json:"rank"`
	Streak    string           `json:"streak"`
	Mischance string           `json:"mischance,omitempty"`
	Identity  []IdentityRecord `json:"identity,omitempty"`
}

// GetValKey returns the validator key from either field.
func (v *Validator) GetValKey() string {
	if v.ValKey != "" {
		return v.ValKey
	}
	return v.ValKeyAlt
}

// IdentityRecord represents an identity record embedded in validator.
type IdentityRecord struct {
	ID        string   `json:"id"`
	Address   string   `json:"address"`
	Key       string   `json:"key"`
	Value     string   `json:"value"`
	Date      string   `json:"date"`
	Verifiers []string `json:"verifiers,omitempty"`
}

// ValidatorsResponse contains validators query response.
type ValidatorsResponse struct {
	Validators []Validator         `json:"validators"`
	Actors     []string            `json:"actors,omitempty"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

// PaginationResponse represents pagination info in response.
type PaginationResponse struct {
	NextKey string `json:"next_key,omitempty"`
	Total   string `json:"total,omitempty"`
}

// ValidatorQueryOpts contains options for querying validators.
type ValidatorQueryOpts struct {
	Address  string
	ValAddr  string
	Moniker  string
	Status   string
	PubKey   string
	Proposer string
}
