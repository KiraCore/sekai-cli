// Package scenarios provides Ansible-like playbook execution for sekai-cli.
// It allows users to define sequences of blockchain operations in YAML files
// and execute them with variable interpolation and transaction tracking.
package scenarios

import (
	"time"
)

// Scenario represents a playbook of steps to execute against the blockchain.
type Scenario struct {
	// Name is a short identifier for the scenario
	Name string `yaml:"name"`

	// Description provides details about what the scenario does
	Description string `yaml:"description,omitempty"`

	// Version of the scenario format (for future compatibility)
	Version string `yaml:"version,omitempty"`

	// Variables are default values that can be overridden via CLI
	Variables map[string]interface{} `yaml:"variables,omitempty"`

	// Steps are the ordered list of operations to execute
	Steps []Step `yaml:"steps"`
}

// Step represents a single operation in a scenario.
type Step struct {
	// Name is a human-readable description of this step
	Name string `yaml:"name"`

	// Module is the SDK module to use (e.g., "bank", "gov", "keys")
	Module string `yaml:"module"`

	// Action is the operation to perform (e.g., "send", "balances", "add")
	Action string `yaml:"action"`

	// Params are the parameters passed to the action
	// Values can contain {{ variable }} placeholders
	Params map[string]string `yaml:"params,omitempty"`

	// Output is the variable name to store the step's result
	// Can be referenced in later steps as {{ output_name.field }}
	Output string `yaml:"output,omitempty"`

	// TxOptions configures transaction-specific settings
	TxOptions *StepTxOptions `yaml:"tx_options,omitempty"`
}

// StepTxOptions configures transaction behavior for a step.
type StepTxOptions struct {
	// Fees to pay for the transaction
	Fees string `yaml:"fees,omitempty"`

	// Gas limit for the transaction
	Gas string `yaml:"gas,omitempty"`

	// Memo to include in the transaction
	Memo string `yaml:"memo,omitempty"`

	// BroadcastMode: sync, async, or block
	BroadcastMode string `yaml:"broadcast_mode,omitempty"`

	// WaitTimeout is how long to wait for TX confirmation (default: 60s)
	WaitTimeout time.Duration `yaml:"wait_timeout,omitempty"`
}

// ExecutionResult contains the complete result of running a scenario.
type ExecutionResult struct {
	// Scenario name that was executed
	Scenario string `json:"scenario"`

	// Success indicates if all steps completed successfully
	Success bool `json:"success"`

	// Steps contains results for each step
	Steps []StepResult `json:"steps"`

	// Duration is how long the entire scenario took
	Duration time.Duration `json:"duration"`

	// Error message if the scenario failed
	Error string `json:"error,omitempty"`
}

// StepResult contains the result of executing a single step.
type StepResult struct {
	// Name of the step
	Name string `json:"name"`

	// Module used
	Module string `json:"module"`

	// Action performed
	Action string `json:"action"`

	// Success indicates if this step completed successfully
	Success bool `json:"success"`

	// Output data from the step (can be query result or TX response)
	Output interface{} `json:"output,omitempty"`

	// TxHash if this was a transaction
	TxHash string `json:"tx_hash,omitempty"`

	// TxCode is the transaction result code (0 = success)
	TxCode uint32 `json:"tx_code,omitempty"`

	// BlockHeight where TX was included
	BlockHeight int64 `json:"block_height,omitempty"`

	// Duration of this step
	Duration time.Duration `json:"duration"`

	// Error message if the step failed
	Error string `json:"error,omitempty"`

	// Skipped indicates if step was skipped (e.g., dry-run mode)
	Skipped bool `json:"skipped,omitempty"`
}

// ExecutorOptions configures how scenarios are executed.
type ExecutorOptions struct {
	// DryRun shows what would happen without executing
	DryRun bool

	// Verbose enables detailed output
	Verbose bool

	// Variables to override scenario defaults
	Variables map[string]string

	// TxWaitTimeout is the default timeout for TX confirmation
	TxWaitTimeout time.Duration

	// TxPollInterval is how often to check for TX confirmation
	TxPollInterval time.Duration

	// ContinueOnError continues executing even if a step fails
	ContinueOnError bool
}

// DefaultExecutorOptions returns sensible defaults for scenario execution.
func DefaultExecutorOptions() *ExecutorOptions {
	return &ExecutorOptions{
		DryRun:          false,
		Verbose:         false,
		Variables:       make(map[string]string),
		TxWaitTimeout:   60 * time.Second,
		TxPollInterval:  2 * time.Second,
		ContinueOnError: false,
	}
}

// StepType indicates whether a step is a query or transaction.
type StepType int

const (
	// StepTypeQuery is a read-only operation
	StepTypeQuery StepType = iota

	// StepTypeTransaction is a state-changing operation
	StepTypeTransaction
)

// String returns the string representation of StepType.
func (s StepType) String() string {
	switch s {
	case StepTypeQuery:
		return "query"
	case StepTypeTransaction:
		return "transaction"
	default:
		return "unknown"
	}
}
