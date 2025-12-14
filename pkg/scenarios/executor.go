package scenarios

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kiracore/sekai-cli/pkg/sdk"
)

// Executor runs scenarios against the blockchain.
type Executor struct {
	client sdk.Client
	opts   *ExecutorOptions
	output io.Writer
	vars   *VariableStore
	mapper *ActionMapper
}

// NewExecutor creates a new scenario executor.
func NewExecutor(client sdk.Client, opts *ExecutorOptions) *Executor {
	if opts == nil {
		opts = DefaultExecutorOptions()
	}

	return &Executor{
		client: client,
		opts:   opts,
		output: os.Stdout,
		vars:   NewVariableStore(),
		mapper: NewActionMapper(client),
	}
}

// SetOutput sets the output writer for progress messages.
func (e *Executor) SetOutput(w io.Writer) {
	e.output = w
}

// Execute runs a scenario and returns the results.
func (e *Executor) Execute(ctx context.Context, scenario *Scenario) (*ExecutionResult, error) {
	startTime := time.Now()

	result := &ExecutionResult{
		Scenario: scenario.Name,
		Steps:    make([]StepResult, 0, len(scenario.Steps)),
		Success:  true,
	}

	// Initialize variables from scenario defaults
	if scenario.Variables != nil {
		e.vars.MergeFrom(scenario.Variables)
	}

	// Apply CLI overrides
	if e.opts.Variables != nil {
		e.vars.MergeFromStringMap(e.opts.Variables)
	}

	e.logf("Starting scenario: %s\n", scenario.Name)
	if scenario.Description != "" {
		e.logf("Description: %s\n", scenario.Description)
	}
	e.logf("Steps: %d\n\n", len(scenario.Steps))

	// Execute each step sequentially
	for i, step := range scenario.Steps {
		stepNum := i + 1
		e.logf("[%d/%d] %s\n", stepNum, len(scenario.Steps), step.Name)

		stepResult := e.executeStep(ctx, &step)
		result.Steps = append(result.Steps, stepResult)

		if stepResult.Success {
			e.logf("  Status: OK")
			if stepResult.TxHash != "" {
				e.logf(" (tx: %s, height: %d)", truncateHash(stepResult.TxHash), stepResult.BlockHeight)
			}
			e.logf("\n")

			// Store output if specified
			if step.Output != "" && stepResult.Output != nil {
				e.vars.Set(step.Output, stepResult.Output)
				if e.opts.Verbose {
					e.logf("  Output stored in: %s\n", step.Output)
				}
			}
		} else {
			result.Success = false
			e.logf("  Status: FAILED\n")
			e.logf("  Error: %s\n", stepResult.Error)

			if !e.opts.ContinueOnError {
				result.Error = fmt.Sprintf("step '%s' failed: %s", step.Name, stepResult.Error)
				break
			}
		}
		e.logf("\n")
	}

	result.Duration = time.Since(startTime)

	// Print summary
	e.logf("=====================================\n")
	if result.Success {
		e.logf("Scenario completed successfully!\n")
	} else {
		e.logf("Scenario failed: %s\n", result.Error)
	}
	e.logf("Duration: %s\n", result.Duration.Round(time.Millisecond))

	return result, nil
}

// executeStep runs a single step and returns the result.
func (e *Executor) executeStep(ctx context.Context, step *Step) StepResult {
	startTime := time.Now()

	result := StepResult{
		Name:   step.Name,
		Module: step.Module,
		Action: step.Action,
	}

	// Handle dry-run mode
	if e.opts.DryRun {
		result.Success = true
		result.Skipped = true
		result.Duration = time.Since(startTime)
		e.logf("  [DRY-RUN] Would execute: %s.%s\n", step.Module, step.Action)
		if len(step.Params) > 0 {
			e.logf("  Params:\n")
			for k, v := range step.Params {
				interpolated, _ := e.vars.Interpolate(v)
				e.logf("    %s: %s\n", k, interpolated)
			}
		}
		return result
	}

	// Interpolate parameters
	params, err := e.vars.InterpolateParams(step.Params)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("parameter interpolation failed: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	if e.opts.Verbose && len(params) > 0 {
		e.logf("  Params:\n")
		for k, v := range params {
			e.logf("    %s: %s\n", k, v)
		}
	}

	// Determine step type
	stepType := GetStepType(step)

	// Execute the action
	output, txResp, err := e.mapper.Execute(ctx, step.Module, step.Action, params, step.TxOptions)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(startTime)
		return result
	}

	result.Output = output

	// For transactions, wait for confirmation and verify success
	if stepType == StepTypeTransaction && txResp != nil {
		result.TxHash = txResp.TxHash
		result.TxCode = txResp.Code
		result.BlockHeight = txResp.Height

		// Check if transaction was successful
		if txResp.Code != 0 {
			result.Success = false
			result.Error = fmt.Sprintf("transaction failed with code %d", txResp.Code)
			result.Duration = time.Since(startTime)
			return result
		}

		// If using async broadcast, wait for confirmation
		if step.TxOptions != nil && step.TxOptions.BroadcastMode == "async" {
			e.logf("  Waiting for TX confirmation...\n")
			confirmed, err := e.waitForTx(ctx, txResp.TxHash, step.TxOptions)
			if err != nil {
				result.Success = false
				result.Error = fmt.Sprintf("failed waiting for TX confirmation: %v", err)
				result.Duration = time.Since(startTime)
				return result
			}
			result.BlockHeight = confirmed.Height
			result.TxCode = confirmed.Code
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime)
	return result
}

// waitForTx polls for transaction confirmation.
func (e *Executor) waitForTx(ctx context.Context, txHash string, txOpts *StepTxOptions) (*sdk.TxResponse, error) {
	timeout := e.opts.TxWaitTimeout
	if txOpts != nil && txOpts.WaitTimeout > 0 {
		timeout = txOpts.WaitTimeout
	}

	pollInterval := e.opts.TxPollInterval

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Query transaction by hash
		resp, err := e.client.Query(ctx, &sdk.QueryRequest{
			Module:   "tx",
			Endpoint: txHash,
		})

		if err == nil && resp != nil {
			// Parse response to get tx result
			var txResp sdk.TxResponse
			if err := parseQueryResponse(resp.Data, &txResp); err == nil {
				if txResp.Height > 0 {
					return &txResp, nil
				}
			}
		}

		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("timeout waiting for transaction %s after %s", truncateHash(txHash), timeout)
}

// logf writes a formatted message to the output.
func (e *Executor) logf(format string, args ...interface{}) {
	fmt.Fprintf(e.output, format, args...)
}

// truncateHash returns a shortened version of a hash for display.
func truncateHash(hash string) string {
	if len(hash) <= 16 {
		return hash
	}
	return hash[:8] + "..." + hash[len(hash)-8:]
}

// parseQueryResponse parses JSON query response into a struct.
func parseQueryResponse(data []byte, v interface{}) error {
	// Import would be needed for json.Unmarshal
	// For now, this is a placeholder - the actual implementation
	// will use encoding/json
	return nil
}
