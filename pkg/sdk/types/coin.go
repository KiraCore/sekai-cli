package types

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Coin represents a token with denomination and amount.
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// NewCoin creates a new Coin.
func NewCoin(denom string, amount int64) Coin {
	return Coin{
		Denom:  denom,
		Amount: strconv.FormatInt(amount, 10),
	}
}

// NewCoinFromString creates a new Coin from string amount.
func NewCoinFromString(denom, amount string) Coin {
	return Coin{
		Denom:  denom,
		Amount: amount,
	}
}

// String returns the string representation of a coin (e.g., "100ukex").
func (c Coin) String() string {
	return c.Amount + c.Denom
}

// IsZero returns true if the amount is zero.
func (c Coin) IsZero() bool {
	return c.Amount == "" || c.Amount == "0"
}

// IsValid returns true if the coin has a valid denom and non-negative amount.
func (c Coin) IsValid() bool {
	if c.Denom == "" {
		return false
	}
	amount, err := strconv.ParseInt(c.Amount, 10, 64)
	if err != nil {
		return false
	}
	return amount >= 0
}

// AmountInt64 returns the amount as int64.
func (c Coin) AmountInt64() (int64, error) {
	return strconv.ParseInt(c.Amount, 10, 64)
}

// Coins represents a collection of Coin.
type Coins []Coin

// NewCoins creates a new Coins collection.
func NewCoins(coins ...Coin) Coins {
	result := make(Coins, 0, len(coins))
	for _, c := range coins {
		if !c.IsZero() {
			result = append(result, c)
		}
	}
	return result.Sort()
}

// String returns the string representation of coins (e.g., "100ukex,50uatom").
func (cs Coins) String() string {
	if len(cs) == 0 {
		return ""
	}
	parts := make([]string, len(cs))
	for i, c := range cs {
		parts[i] = c.String()
	}
	return strings.Join(parts, ",")
}

// Sort sorts coins by denom alphabetically.
func (cs Coins) Sort() Coins {
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].Denom < cs[j].Denom
	})
	return cs
}

// IsValid returns true if all coins are valid.
func (cs Coins) IsValid() bool {
	for _, c := range cs {
		if !c.IsValid() {
			return false
		}
	}
	return true
}

// IsZero returns true if all coins are zero.
func (cs Coins) IsZero() bool {
	for _, c := range cs {
		if !c.IsZero() {
			return false
		}
	}
	return true
}

// AmountOf returns the amount of a specific denom.
func (cs Coins) AmountOf(denom string) string {
	for _, c := range cs {
		if c.Denom == denom {
			return c.Amount
		}
	}
	return "0"
}

// GetCoin returns the coin for a specific denom.
func (cs Coins) GetCoin(denom string) (Coin, bool) {
	for _, c := range cs {
		if c.Denom == denom {
			return c, true
		}
	}
	return Coin{}, false
}

// ParseCoin parses a coin string like "100ukex" into a Coin.
func ParseCoin(s string) (Coin, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Coin{}, fmt.Errorf("empty coin string")
	}

	// Find where the amount ends and denom begins
	// Amount is digits, denom is the rest
	re := regexp.MustCompile(`^(\d+)([a-zA-Z][a-zA-Z0-9/]*)$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return Coin{}, fmt.Errorf("invalid coin format: %s", s)
	}

	return Coin{
		Amount: matches[1],
		Denom:  matches[2],
	}, nil
}

// ParseCoins parses a comma-separated list of coins.
func ParseCoins(s string) (Coins, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	parts := strings.Split(s, ",")
	coins := make(Coins, 0, len(parts))

	for _, part := range parts {
		coin, err := ParseCoin(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		coins = append(coins, coin)
	}

	return coins.Sort(), nil
}

// DecCoin represents a decimal coin for precise calculations.
type DecCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"` // Decimal string like "100.5"
}

// DecCoins is a collection of DecCoin.
type DecCoins []DecCoin

// String returns the string representation of decimal coins.
func (dcs DecCoins) String() string {
	if len(dcs) == 0 {
		return ""
	}
	parts := make([]string, len(dcs))
	for i, dc := range dcs {
		parts[i] = dc.Amount + dc.Denom
	}
	return strings.Join(parts, ",")
}
