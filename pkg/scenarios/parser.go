package scenarios

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

// LoadFromFile loads a scenario from a YAML file.
func LoadFromFile(path string) (*Scenario, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open scenario file: %w", err)
	}
	defer f.Close()

	return Load(f)
}

// Load parses a scenario from a YAML reader.
func Load(r io.Reader) (*Scenario, error) {
	var scenario Scenario

	decoder := yaml.NewDecoder(r, yaml.Strict())

	if err := decoder.Decode(&scenario); err != nil {
		return nil, fmt.Errorf("failed to parse scenario YAML: %w", err)
	}

	if err := validate(&scenario); err != nil {
		return nil, err
	}

	return &scenario, nil
}

// LoadFromString parses a scenario from a YAML string.
func LoadFromString(content string) (*Scenario, error) {
	return Load(strings.NewReader(content))
}

// validate checks that a scenario has all required fields and valid structure.
func validate(s *Scenario) error {
	if s.Name == "" {
		return fmt.Errorf("scenario validation failed: 'name' is required")
	}

	if len(s.Steps) == 0 {
		return fmt.Errorf("scenario validation failed: at least one step is required")
	}

	outputs := make(map[string]bool)

	for i, step := range s.Steps {
		stepNum := i + 1

		if step.Name == "" {
			return fmt.Errorf("scenario validation failed: step %d missing 'name'", stepNum)
		}

		if step.Module == "" {
			return fmt.Errorf("scenario validation failed: step '%s' missing 'module'", step.Name)
		}

		if step.Action == "" {
			return fmt.Errorf("scenario validation failed: step '%s' missing 'action'", step.Name)
		}

		// Validate module name
		if !isValidModule(step.Module) {
			return fmt.Errorf("scenario validation failed: step '%s' has unknown module '%s'", step.Name, step.Module)
		}

		// Track output variables for duplicate detection
		if step.Output != "" {
			if outputs[step.Output] {
				return fmt.Errorf("scenario validation failed: duplicate output variable '%s' in step '%s'", step.Output, step.Name)
			}
			outputs[step.Output] = true
		}

		// Validate tx_options if present
		if step.TxOptions != nil {
			if step.TxOptions.BroadcastMode != "" {
				mode := strings.ToLower(step.TxOptions.BroadcastMode)
				if mode != "sync" && mode != "async" && mode != "block" {
					return fmt.Errorf("scenario validation failed: step '%s' has invalid broadcast_mode '%s' (must be sync, async, or block)", step.Name, step.TxOptions.BroadcastMode)
				}
			}
		}
	}

	return nil
}

// isValidModule checks if a module name is supported.
func isValidModule(module string) bool {
	validModules := map[string]bool{
		// Core modules
		"keys":   true,
		"bank":   true,
		"status": true,
		"auth":   true,

		// Governance
		"gov":        true,
		"customgov":  true,
		"permission": true,
		"role":       true,
		"councilor":  true,
		"poll":       true,
		"proposal":   true,

		// Staking
		"staking":       true,
		"customstaking": true,
		"multistaking":  true,

		// Economy
		"tokens":      true,
		"basket":      true,
		"spending":    true,
		"ubi":         true,
		"distributor": true,

		// Advanced
		"upgrade":     true,
		"slashing":    true,
		"collectives": true,
		"custody":     true,
		"bridge":      true,
		"layer2":      true,
		"recovery":    true,
	}

	return validModules[strings.ToLower(module)]
}

// GetStepType determines if a step is a query or transaction based on the action.
func GetStepType(step *Step) StepType {
	// Actions that are typically queries (read-only)
	queryActions := map[string]bool{
		// Common query patterns
		"list":        true,
		"show":        true,
		"get":         true,
		"query":       true,
		"balances":    true,
		"balance":     true,
		"total":       true,
		"validators":  true,
		"validator":   true,
		"proposals":   true,
		"proposal":    true,
		"votes":       true,
		"vote":        true,
		"roles":       true,
		"role":        true,
		"permissions": true,
		"pools":       true,
		"pool":        true,
		"records":     true,
		"record":      true,
		"status":      true,
		"params":      true,
		"account":     true,
		"accounts":    true,
		"all-rates":   true,
		"rate":        true,
	}

	action := strings.ToLower(step.Action)

	// Check if it's a known query action
	if queryActions[action] {
		return StepTypeQuery
	}

	// Check for query prefixes
	if strings.HasPrefix(action, "query-") ||
		strings.HasPrefix(action, "get-") ||
		strings.HasPrefix(action, "list-") ||
		strings.HasPrefix(action, "show-") {
		return StepTypeQuery
	}

	// Default to transaction for everything else
	return StepTypeTransaction
}

// String returns a human-readable representation of a scenario.
func (s *Scenario) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Scenario: %s\n", s.Name))
	if s.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", s.Description))
	}
	if len(s.Variables) > 0 {
		sb.WriteString("Variables:\n")
		for k, v := range s.Variables {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}
	sb.WriteString(fmt.Sprintf("Steps: %d\n", len(s.Steps)))
	for i, step := range s.Steps {
		sb.WriteString(fmt.Sprintf("  %d. [%s] %s.%s", i+1, GetStepType(&step), step.Module, step.Action))
		if step.Output != "" {
			sb.WriteString(fmt.Sprintf(" -> %s", step.Output))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
