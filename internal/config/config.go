// Package config provides configuration management for sekai-cli.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds the CLI configuration.
type Config struct {
	// Container is the Docker container name.
	Container string `json:"container" yaml:"container"`

	// ChainID is the blockchain network identifier.
	ChainID string `json:"chain_id" yaml:"chain_id"`

	// Home is the sekaid home directory.
	Home string `json:"home" yaml:"home"`

	// Node is the RPC endpoint URL.
	Node string `json:"node" yaml:"node"`

	// KeyringBackend is the keyring backend type.
	KeyringBackend string `json:"keyring_backend" yaml:"keyring_backend"`

	// Fees is the default transaction fee.
	Fees string `json:"fees" yaml:"fees"`

	// Gas is the default gas limit for transactions.
	Gas string `json:"gas" yaml:"gas"`

	// GasAdjustment is the gas adjustment factor.
	GasAdjustment float64 `json:"gas_adjustment" yaml:"gas_adjustment"`

	// BroadcastMode is the default broadcast mode.
	BroadcastMode string `json:"broadcast_mode" yaml:"broadcast_mode"`

	// Output is the default output format.
	Output string `json:"output" yaml:"output"`

	// UseREST indicates whether to use REST API instead of Docker.
	UseREST bool `json:"use_rest" yaml:"use_rest"`

	// RESTURL is the REST API endpoint URL.
	RESTURL string `json:"rest_url" yaml:"rest_url"`

	// Verbose enables verbose output.
	Verbose bool `json:"verbose" yaml:"verbose"`

	// configPath is the path where config was loaded from.
	configPath string
}

// Default returns a Config with default values.
func Default() *Config {
	return &Config{
		Container:      "sekai-node",
		ChainID:        "localnet-1",
		Home:           "/sekai",
		Node:           "tcp://localhost:26657",
		KeyringBackend: "test",
		Fees:           "100ukex",
		Gas:            "200000",
		GasAdjustment:  1.3,
		BroadcastMode:  "sync",
		Output:         "text",
		UseREST:        false,
		RESTURL:        "http://localhost:1317",
		Verbose:        false,
	}
}

// Load loads configuration from the default locations.
// Priority: CLI flags > environment variables > config file > defaults
func Load() *Config {
	cfg := Default()

	// Try to load from config file
	configPath := findConfigFile()
	if configPath != "" {
		if err := cfg.LoadFromFile(configPath); err == nil {
			cfg.configPath = configPath
		}
	}

	// Override with environment variables
	cfg.loadFromEnv()

	return cfg
}

// LoadFromFile loads configuration from a file.
func (c *Config) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Try JSON first
	if err := json.Unmarshal(data, c); err != nil {
		// Try YAML-like format
		if err := c.parseYAML(string(data)); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	c.configPath = path
	return nil
}

// Save saves configuration to the specified file.
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	c.configPath = path
	return nil
}

// Path returns the path where config was loaded from.
func (c *Config) Path() string {
	return c.configPath
}

// loadFromEnv loads configuration from environment variables.
func (c *Config) loadFromEnv() {
	if v := os.Getenv("SEKAI_CONTAINER"); v != "" {
		c.Container = v
	}
	if v := os.Getenv("SEKAI_CHAIN_ID"); v != "" {
		c.ChainID = v
	}
	if v := os.Getenv("SEKAI_HOME"); v != "" {
		c.Home = v
	}
	if v := os.Getenv("SEKAI_NODE"); v != "" {
		c.Node = v
	}
	if v := os.Getenv("SEKAI_KEYRING_BACKEND"); v != "" {
		c.KeyringBackend = v
	}
	if v := os.Getenv("SEKAI_FEES"); v != "" {
		c.Fees = v
	}
	if v := os.Getenv("SEKAI_GAS"); v != "" {
		c.Gas = v
	}
	if v := os.Getenv("SEKAI_GAS_ADJUSTMENT"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.GasAdjustment = f
		}
	}
	if v := os.Getenv("SEKAI_BROADCAST_MODE"); v != "" {
		c.BroadcastMode = v
	}
	if v := os.Getenv("SEKAI_OUTPUT"); v != "" {
		c.Output = v
	}
	if v := os.Getenv("SEKAI_USE_REST"); v != "" {
		c.UseREST = v == "true" || v == "1"
	}
	if v := os.Getenv("SEKAI_REST_URL"); v != "" {
		c.RESTURL = v
	}
	if v := os.Getenv("SEKAI_VERBOSE"); v != "" {
		c.Verbose = v == "true" || v == "1"
	}
}

// parseYAML parses a simple YAML-like configuration format.
// This is a basic implementation that handles key: value pairs.
func (c *Config) parseYAML(data string) error {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		switch key {
		case "container":
			c.Container = value
		case "chain_id":
			c.ChainID = value
		case "home":
			c.Home = value
		case "node":
			c.Node = value
		case "keyring_backend":
			c.KeyringBackend = value
		case "fees":
			c.Fees = value
		case "gas":
			c.Gas = value
		case "gas_adjustment":
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				c.GasAdjustment = f
			}
		case "broadcast_mode":
			c.BroadcastMode = value
		case "output":
			c.Output = value
		case "use_rest":
			c.UseREST = value == "true"
		case "rest_url":
			c.RESTURL = value
		case "verbose":
			c.Verbose = value == "true"
		}
	}
	return nil
}

// findConfigFile looks for config file in standard locations.
// Follows XDG Base Directory Specification for Ubuntu/Linux.
func findConfigFile() string {
	// Check locations in order of priority
	locations := []string{
		"./sekai-cli.json",
		"./sekai-cli.yaml",
	}

	// XDG_CONFIG_HOME (default: ~/.config)
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			configHome = filepath.Join(home, ".config")
		}
	}
	if configHome != "" {
		locations = append(locations,
			filepath.Join(configHome, "sekai-cli", "config.json"),
			filepath.Join(configHome, "sekai-cli", "config.yaml"),
		)
	}

	// System-wide config
	locations = append(locations,
		"/etc/sekai-cli/config.json",
		"/etc/sekai-cli/config.yaml",
	)

	// Legacy locations (for backwards compatibility)
	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations,
			filepath.Join(home, ".sekai-cli", "config.json"),
			filepath.Join(home, ".sekai-cli", "config.yaml"),
		)
	}

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// DefaultConfigPath returns the default config file path.
// Uses XDG_CONFIG_HOME (~/.config/sekai-cli) on Linux.
func DefaultConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		if home, err := os.UserHomeDir(); err == nil {
			configHome = filepath.Join(home, ".config")
		}
	}
	if configHome != "" {
		return filepath.Join(configHome, "sekai-cli", "config.json")
	}
	return "./sekai-cli.json"
}

// Merge merges another config into this one (non-empty values override).
func (c *Config) Merge(other *Config) {
	if other == nil {
		return
	}
	if other.Container != "" {
		c.Container = other.Container
	}
	if other.ChainID != "" {
		c.ChainID = other.ChainID
	}
	if other.Home != "" {
		c.Home = other.Home
	}
	if other.Node != "" {
		c.Node = other.Node
	}
	if other.KeyringBackend != "" {
		c.KeyringBackend = other.KeyringBackend
	}
	if other.Fees != "" {
		c.Fees = other.Fees
	}
	if other.Gas != "" {
		c.Gas = other.Gas
	}
	if other.GasAdjustment != 0 {
		c.GasAdjustment = other.GasAdjustment
	}
	if other.BroadcastMode != "" {
		c.BroadcastMode = other.BroadcastMode
	}
	if other.Output != "" {
		c.Output = other.Output
	}
	if other.UseREST {
		c.UseREST = other.UseREST
	}
	if other.RESTURL != "" {
		c.RESTURL = other.RESTURL
	}
	if other.Verbose {
		c.Verbose = other.Verbose
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if !c.UseREST && c.Container == "" {
		return fmt.Errorf("container name is required when not using REST API")
	}
	if c.UseREST && c.RESTURL == "" {
		return fmt.Errorf("REST URL is required when using REST API")
	}
	if c.GasAdjustment < 1.0 {
		return fmt.Errorf("gas adjustment must be >= 1.0")
	}
	return nil
}
