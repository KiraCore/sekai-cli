package types

import (
	"regexp"
	"strings"
)

// Address prefixes for KIRA network.
const (
	// Bech32PrefixAccAddr is the prefix for account addresses.
	Bech32PrefixAccAddr = "kira"

	// Bech32PrefixAccPub is the prefix for account public keys.
	Bech32PrefixAccPub = "kirapub"

	// Bech32PrefixValAddr is the prefix for validator operator addresses.
	Bech32PrefixValAddr = "kiravaloper"

	// Bech32PrefixValPub is the prefix for validator operator public keys.
	Bech32PrefixValPub = "kiravaloperpub"

	// Bech32PrefixConsAddr is the prefix for consensus node addresses.
	Bech32PrefixConsAddr = "kiravalcons"

	// Bech32PrefixConsPub is the prefix for consensus node public keys.
	Bech32PrefixConsPub = "kiravalconspub"
)

// AccAddress represents an account address.
type AccAddress string

// ValAddress represents a validator operator address.
type ValAddress string

// ConsAddress represents a consensus node address.
type ConsAddress string

// String returns the string representation of the address.
func (a AccAddress) String() string {
	return string(a)
}

// String returns the string representation of the validator address.
func (a ValAddress) String() string {
	return string(a)
}

// String returns the string representation of the consensus address.
func (a ConsAddress) String() string {
	return string(a)
}

// IsValid checks if the account address is valid.
func (a AccAddress) IsValid() bool {
	return IsValidAddress(string(a))
}

// IsValid checks if the validator address is valid.
func (a ValAddress) IsValid() bool {
	return IsValidValAddress(string(a))
}

// IsValid checks if the consensus address is valid.
func (a ConsAddress) IsValid() bool {
	return IsValidConsAddress(string(a))
}

// Empty returns true if the address is empty.
func (a AccAddress) Empty() bool {
	return a == ""
}

// Empty returns true if the validator address is empty.
func (a ValAddress) Empty() bool {
	return a == ""
}

// Empty returns true if the consensus address is empty.
func (a ConsAddress) Empty() bool {
	return a == ""
}

// IsValidAddress validates a KIRA account address.
// A valid address has the format: kira1[39 alphanumeric chars]
func IsValidAddress(addr string) bool {
	if addr == "" {
		return false
	}
	// Bech32 addresses are case-insensitive but should be lowercase
	addr = strings.ToLower(addr)

	// Check prefix
	if !strings.HasPrefix(addr, Bech32PrefixAccAddr+"1") {
		return false
	}

	// Basic pattern check: kira1 + 38 alphanumeric chars (bech32)
	// Actual bech32 validation would require implementing the algorithm
	pattern := regexp.MustCompile(`^kira1[a-z0-9]{38}$`)
	return pattern.MatchString(addr)
}

// IsValidValAddress validates a KIRA validator operator address.
func IsValidValAddress(addr string) bool {
	if addr == "" {
		return false
	}
	addr = strings.ToLower(addr)

	if !strings.HasPrefix(addr, Bech32PrefixValAddr+"1") {
		return false
	}

	// kiravaloper1 + bech32 data
	pattern := regexp.MustCompile(`^kiravaloper1[a-z0-9]{38}$`)
	return pattern.MatchString(addr)
}

// IsValidConsAddress validates a KIRA consensus node address.
func IsValidConsAddress(addr string) bool {
	if addr == "" {
		return false
	}
	addr = strings.ToLower(addr)

	if !strings.HasPrefix(addr, Bech32PrefixConsAddr+"1") {
		return false
	}

	// kiravalcons1 + bech32 data
	pattern := regexp.MustCompile(`^kiravalcons1[a-z0-9]{38}$`)
	return pattern.MatchString(addr)
}

// GetAddressType returns the type of address based on its prefix.
func GetAddressType(addr string) string {
	addr = strings.ToLower(addr)
	switch {
	case strings.HasPrefix(addr, Bech32PrefixConsAddr+"1"):
		return "consensus"
	case strings.HasPrefix(addr, Bech32PrefixValAddr+"1"):
		return "validator"
	case strings.HasPrefix(addr, Bech32PrefixAccAddr+"1"):
		return "account"
	default:
		return "unknown"
	}
}

// MustAccAddressFromString creates an AccAddress from string, panics if invalid.
func MustAccAddressFromString(s string) AccAddress {
	if !IsValidAddress(s) {
		panic("invalid account address: " + s)
	}
	return AccAddress(s)
}

// MustValAddressFromString creates a ValAddress from string, panics if invalid.
func MustValAddressFromString(s string) ValAddress {
	if !IsValidValAddress(s) {
		panic("invalid validator address: " + s)
	}
	return ValAddress(s)
}

// MustConsAddressFromString creates a ConsAddress from string, panics if invalid.
func MustConsAddressFromString(s string) ConsAddress {
	if !IsValidConsAddress(s) {
		panic("invalid consensus address: " + s)
	}
	return ConsAddress(s)
}
