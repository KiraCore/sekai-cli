package sdk

// ClientConfig holds common configuration for all client types.
type ClientConfig struct {
	// ChainID is the blockchain network identifier.
	ChainID string

	// KeyringBackend is the keyring backend type (test, file, os, kwallet, pass, memory).
	KeyringBackend string

	// Home is the configuration directory path.
	Home string

	// Node is the RPC endpoint URL (for REST/gRPC clients).
	Node string

	// Fees is the default transaction fee.
	Fees string

	// GasAdjustment is the gas adjustment factor.
	GasAdjustment float64

	// BroadcastMode is the default transaction broadcast mode.
	BroadcastMode string

	// Output is the default output format (json, text, yaml).
	Output string

	// Timeout is the request timeout in seconds.
	Timeout int
}

// DefaultClientConfig returns a ClientConfig with sensible defaults.
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		ChainID:        ChainIDLocalnet,
		KeyringBackend: "test",
		Home:           "/.sekaid",
		Node:           "tcp://localhost:26657",
		Fees:           DefaultFees,
		GasAdjustment:  DefaultGasAdjustment,
		BroadcastMode:  "sync",
		Output:         "json",
		Timeout:        30,
	}
}

// ClientOption is a function that configures a ClientConfig.
type ClientOption func(*ClientConfig)

// WithChainID sets the chain ID.
func WithChainID(chainID string) ClientOption {
	return func(c *ClientConfig) {
		c.ChainID = chainID
	}
}

// WithKeyringBackend sets the keyring backend.
func WithKeyringBackend(backend string) ClientOption {
	return func(c *ClientConfig) {
		c.KeyringBackend = backend
	}
}

// WithHome sets the home directory.
func WithHome(home string) ClientOption {
	return func(c *ClientConfig) {
		c.Home = home
	}
}

// WithNode sets the node RPC endpoint.
func WithNode(node string) ClientOption {
	return func(c *ClientConfig) {
		c.Node = node
	}
}

// WithFees sets the default fees.
func WithFees(fees string) ClientOption {
	return func(c *ClientConfig) {
		c.Fees = fees
	}
}

// WithGasAdjustment sets the gas adjustment factor.
func WithGasAdjustment(adj float64) ClientOption {
	return func(c *ClientConfig) {
		c.GasAdjustment = adj
	}
}

// WithBroadcastMode sets the default broadcast mode.
func WithBroadcastMode(mode string) ClientOption {
	return func(c *ClientConfig) {
		c.BroadcastMode = mode
	}
}

// WithOutput sets the default output format.
func WithOutput(output string) ClientOption {
	return func(c *ClientConfig) {
		c.Output = output
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(seconds int) ClientOption {
	return func(c *ClientConfig) {
		c.Timeout = seconds
	}
}

// ApplyOptions applies a list of options to a ClientConfig.
func ApplyOptions(cfg *ClientConfig, opts ...ClientOption) {
	for _, opt := range opts {
		opt(cfg)
	}
}
