// Package cache provides auto-detected network configuration caching for sekai-cli.
// It stores container, network properties, and keys information to reduce
// the need for users to specify flags like --chain-id, --fees, --from repeatedly.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Cache holds cached network configuration.
type Cache struct {
	// Version is the cache format version for future compatibility.
	Version int `json:"version"`

	// LastSync is when the cache was last updated.
	LastSync time.Time `json:"last_sync"`

	// Container is the Docker container name.
	Container string `json:"container"`

	// Network contains cached network properties.
	Network NetworkCache `json:"network"`

	// Keys contains cached key information.
	Keys []KeyCache `json:"keys"`

	// DefaultKey is the default signing key name.
	DefaultKey string `json:"default_key"`

	// cachePath is the path where cache was loaded from.
	cachePath string
}

// NetworkCache contains cached network properties.
type NetworkCache struct {
	ChainID                  string `json:"chain_id"`
	Moniker                  string `json:"moniker"`
	MinTxFee                 string `json:"min_tx_fee"`
	MaxTxFee                 string `json:"max_tx_fee"`
	VoteQuorum               string `json:"vote_quorum"`
	MinimumProposalEndTime   string `json:"minimum_proposal_end_time"`
	ProposalEnactmentTime    string `json:"proposal_enactment_time"`
	EnableForeignFeePayments bool   `json:"enable_foreign_fee_payments"`
	MinValidators            string `json:"min_validators"`
	PoorNetworkMaxBankSend   string `json:"poor_network_max_bank_send"`
	UnjailMaxTime            string `json:"unjail_max_time"`
	UnstakingPeriod          string `json:"unstaking_period"`
	MaxDelegators            string `json:"max_delegators"`
}

// KeyCache contains cached key information.
type KeyCache struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Type    string `json:"type"`
}

// New creates a new empty cache.
func New() *Cache {
	return &Cache{
		Version:  1,
		LastSync: time.Now(),
		Keys:     []KeyCache{},
	}
}

// Load loads cache from the default location.
func Load() (*Cache, error) {
	path := DefaultCachePath()
	return LoadFromFile(path)
}

// LoadFromFile loads cache from a specific file.
func LoadFromFile(path string) (*Cache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("cache not found (run 'sekai-cli init' first)")
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	cache.cachePath = path
	return &cache, nil
}

// TryLoad attempts to load cache, returning nil if not found.
func TryLoad() *Cache {
	cache, err := Load()
	if err != nil {
		return nil
	}
	return cache
}

// Save saves cache to the default location.
func (c *Cache) Save() error {
	return c.SaveToFile(DefaultCachePath())
}

// SaveToFile saves cache to a specific file.
func (c *Cache) SaveToFile(path string) error {
	c.LastSync = time.Now()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	c.cachePath = path
	return nil
}

// DefaultCachePath returns the default cache file path.
// Uses XDG_CACHE_HOME (~/.cache/sekai-cli) on Linux.
func DefaultCachePath() string {
	cacheHome := os.Getenv("XDG_CACHE_HOME")
	if cacheHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cacheHome = filepath.Join(home, ".cache")
		}
	}
	if cacheHome != "" {
		return filepath.Join(cacheHome, "sekai-cli", "cache.json")
	}
	return "./sekai-cli-cache.json"
}

// Exists returns true if a cache file exists at the default location.
func Exists() bool {
	_, err := os.Stat(DefaultCachePath())
	return err == nil
}

// Clear removes the cache file.
func Clear() error {
	path := DefaultCachePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Already doesn't exist
	}
	return os.Remove(path)
}

// Path returns the path where cache was loaded from.
func (c *Cache) Path() string {
	if c.cachePath != "" {
		return c.cachePath
	}
	return DefaultCachePath()
}

// GetDefaultKey returns the default signing key name.
func (c *Cache) GetDefaultKey() string {
	return c.DefaultKey
}

// GetChainID returns the cached chain ID.
func (c *Cache) GetChainID() string {
	return c.Network.ChainID
}

// GetMinFee returns the cached minimum transaction fee.
func (c *Cache) GetMinFee() string {
	return c.Network.MinTxFee
}

// GetContainer returns the cached container name.
func (c *Cache) GetContainer() string {
	return c.Container
}

// SetDefaultKey sets the default signing key.
func (c *Cache) SetDefaultKey(name string) error {
	// Verify key exists
	for _, k := range c.Keys {
		if k.Name == name {
			c.DefaultKey = name
			return nil
		}
	}
	return fmt.Errorf("key '%s' not found in cache", name)
}

// GetKeyByName returns a key by name.
func (c *Cache) GetKeyByName(name string) *KeyCache {
	for _, k := range c.Keys {
		if k.Name == name {
			return &k
		}
	}
	return nil
}

// GetKeyByAddress returns a key by address.
func (c *Cache) GetKeyByAddress(address string) *KeyCache {
	for _, k := range c.Keys {
		if k.Address == address {
			return &k
		}
	}
	return nil
}

// KeyNames returns all key names.
func (c *Cache) KeyNames() []string {
	names := make([]string, len(c.Keys))
	for i, k := range c.Keys {
		names[i] = k.Name
	}
	return names
}

// Age returns how long ago the cache was last synced.
func (c *Cache) Age() time.Duration {
	return time.Since(c.LastSync)
}

// IsStale returns true if the cache is older than the given duration.
func (c *Cache) IsStale(maxAge time.Duration) bool {
	return c.Age() > maxAge
}

// FormatAge returns a human-readable age string.
func (c *Cache) FormatAge() string {
	age := c.Age()
	switch {
	case age < time.Minute:
		return "just now"
	case age < time.Hour:
		mins := int(age.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case age < 24*time.Hour:
		hours := int(age.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		days := int(age.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// ShortenAddress returns a shortened address for display.
func ShortenAddress(addr string) string {
	if len(addr) <= 20 {
		return addr
	}
	return addr[:10] + "..." + addr[len(addr)-6:]
}

// FormatKeys returns a formatted list of keys for display.
func (c *Cache) FormatKeys() string {
	if len(c.Keys) == 0 {
		return "  (none)"
	}

	var lines []string
	for _, k := range c.Keys {
		marker := ""
		if k.Name == c.DefaultKey {
			marker = " [default]"
		}
		lines = append(lines, fmt.Sprintf("  %s (%s)%s", k.Name, ShortenAddress(k.Address), marker))
	}
	return strings.Join(lines, "\n")
}

// Summary returns a formatted summary of the cache.
func (c *Cache) Summary() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Container:   %s\n", c.Container))
	b.WriteString(fmt.Sprintf("Chain ID:    %s\n", c.Network.ChainID))
	if c.Network.Moniker != "" {
		b.WriteString(fmt.Sprintf("Moniker:     %s\n", c.Network.Moniker))
	}
	b.WriteString(fmt.Sprintf("Min TX Fee:  %s\n", c.Network.MinTxFee))
	b.WriteString(fmt.Sprintf("Max TX Fee:  %s\n", c.Network.MaxTxFee))
	b.WriteString(fmt.Sprintf("Last Sync:   %s\n", c.FormatAge()))
	b.WriteString(fmt.Sprintf("\nKeys (%d):\n%s\n", len(c.Keys), c.FormatKeys()))
	return b.String()
}
