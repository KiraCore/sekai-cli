package cli

// GlobalFlags returns the common flags used across all commands.
func GlobalFlags() []Flag {
	return []Flag{
		{
			Name:    "help",
			Short:   "h",
			Usage:   "Show help for command",
			Default: "false",
		},
		{
			Name:    "config",
			Short:   "c",
			Usage:   "Path to config file",
			Default: "",
		},
		{
			Name:    "output",
			Short:   "o",
			Usage:   "Output format (text, json, yaml)",
			Default: "text",
		},
		{
			Name:  "verbose",
			Short: "v",
			Usage: "Enable verbose output",
		},
	}
}

// ConnectionFlags returns flags for blockchain connection settings.
func ConnectionFlags() []Flag {
	return []Flag{
		{
			Name:    "container",
			Usage:   "Docker container name",
			Default: "",
		},
		{
			Name:    "node",
			Usage:   "Node RPC endpoint",
			Default: "tcp://localhost:26657",
		},
		{
			Name:    "chain-id",
			Usage:   "Chain ID",
			Default: "",
		},
	}
}

// TxFlags returns flags for transaction commands.
func TxFlags() []Flag {
	return []Flag{
		{
			Name:  "from",
			Usage: "Name or address of key to sign with",
		},
		{
			Name:    "fees",
			Usage:   "Transaction fees",
			Default: "",
		},
		{
			Name:    "gas",
			Usage:   "Gas limit",
			Default: "",
		},
		{
			Name:    "gas-adjustment",
			Usage:   "Gas adjustment factor",
			Default: "1.3",
		},
		{
			Name:  "memo",
			Usage: "Transaction memo",
		},
		{
			Name:    "broadcast-mode",
			Usage:   "Broadcast mode (sync, async, block)",
			Default: "sync",
		},
		{
			Name:    "keyring-backend",
			Usage:   "Keyring backend (test, file, os)",
			Default: "test",
		},
		{
			Name:    "yes",
			Short:   "y",
			Usage:   "Skip confirmation prompts",
			Default: "false",
		},
	}
}

// QueryFlags returns flags for query commands.
func QueryFlags() []Flag {
	return []Flag{
		{
			Name:  "height",
			Usage: "Query at specific block height",
		},
	}
}

// PaginationFlags returns flags for paginated queries.
func PaginationFlags() []Flag {
	return []Flag{
		{
			Name:    "limit",
			Usage:   "Maximum number of results",
			Default: "100",
		},
		{
			Name:  "offset",
			Usage: "Number of results to skip",
		},
		{
			Name:  "page-key",
			Usage: "Pagination key",
		},
		{
			Name:  "count-total",
			Usage: "Include total count in response",
		},
	}
}

// AddGlobalFlags adds global flags to a command.
func AddGlobalFlags(cmd *Command) {
	for _, f := range GlobalFlags() {
		cmd.AddFlag(f)
	}
}

// AddConnectionFlags adds connection flags to a command.
func AddConnectionFlags(cmd *Command) {
	for _, f := range ConnectionFlags() {
		cmd.AddFlag(f)
	}
}

// AddTxFlags adds transaction flags to a command.
func AddTxFlags(cmd *Command) {
	for _, f := range TxFlags() {
		cmd.AddFlag(f)
	}
}

// AddQueryFlags adds query flags to a command.
func AddQueryFlags(cmd *Command) {
	for _, f := range QueryFlags() {
		cmd.AddFlag(f)
	}
}

// AddPaginationFlags adds pagination flags to a command.
func AddPaginationFlags(cmd *Command) {
	for _, f := range PaginationFlags() {
		cmd.AddFlag(f)
	}
}
