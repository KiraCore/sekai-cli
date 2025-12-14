package scenarios

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// VariableStore holds variables for scenario execution.
// It supports nested access via dot notation (e.g., "step_output.field.subfield").
type VariableStore struct {
	vars map[string]interface{}
}

// NewVariableStore creates a new variable store.
func NewVariableStore() *VariableStore {
	return &VariableStore{
		vars: make(map[string]interface{}),
	}
}

// Set stores a variable value.
func (vs *VariableStore) Set(name string, value interface{}) {
	vs.vars[name] = value
}

// Get retrieves a variable value with support for nested access.
// Supports dot notation: "output.field.subfield"
func (vs *VariableStore) Get(name string) (interface{}, bool) {
	parts := strings.Split(name, ".")

	// Get the root variable
	current, exists := vs.vars[parts[0]]
	if !exists {
		return nil, false
	}

	// Navigate nested fields
	for _, part := range parts[1:] {
		switch v := current.(type) {
		case map[string]interface{}:
			current, exists = v[part]
			if !exists {
				return nil, false
			}
		case map[interface{}]interface{}:
			current, exists = v[part]
			if !exists {
				return nil, false
			}
		default:
			// Try to convert to map via JSON for structs
			data, err := json.Marshal(current)
			if err != nil {
				return nil, false
			}
			var m map[string]interface{}
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, false
			}
			current, exists = m[part]
			if !exists {
				return nil, false
			}
		}
	}

	return current, true
}

// GetString retrieves a variable value as a string.
func (vs *VariableStore) GetString(name string) (string, bool) {
	val, exists := vs.Get(name)
	if !exists {
		return "", false
	}
	return toString(val), true
}

// MergeFrom merges variables from a map into the store.
// Existing variables with the same name are overwritten.
func (vs *VariableStore) MergeFrom(vars map[string]interface{}) {
	for k, v := range vars {
		vs.vars[k] = v
	}
}

// MergeFromStringMap merges variables from a string map (CLI overrides).
func (vs *VariableStore) MergeFromStringMap(vars map[string]string) {
	for k, v := range vars {
		vs.vars[k] = v
	}
}

// All returns all variables as a map.
func (vs *VariableStore) All() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range vs.vars {
		result[k] = v
	}
	return result
}

// variablePattern matches {{ variable_name }} patterns.
var variablePattern = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_\.]*)\s*\}\}`)

// Interpolate replaces {{ variable }} placeholders in a string with their values.
func (vs *VariableStore) Interpolate(input string) (string, error) {
	var lastErr error

	result := variablePattern.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name from {{ name }}
		submatches := variablePattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}

		varName := submatches[1]
		value, exists := vs.GetString(varName)
		if !exists {
			lastErr = fmt.Errorf("undefined variable: %s", varName)
			return match
		}

		return value
	})

	return result, lastErr
}

// InterpolateParams interpolates all values in a params map.
func (vs *VariableStore) InterpolateParams(params map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	for k, v := range params {
		interpolated, err := vs.Interpolate(v)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate param '%s': %w", k, err)
		}
		result[k] = interpolated
	}

	return result, nil
}

// HasVariables checks if a string contains variable placeholders.
func HasVariables(s string) bool {
	return variablePattern.MatchString(s)
}

// ExtractVariables returns all variable names referenced in a string.
func ExtractVariables(s string) []string {
	matches := variablePattern.FindAllStringSubmatch(s, -1)
	vars := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) >= 2 {
			varName := match[1]
			if !seen[varName] {
				vars = append(vars, varName)
				seen[varName] = true
			}
		}
	}

	return vars
}

// toString converts any value to a string.
func toString(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	case []byte:
		return string(val)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case map[string]interface{}, []interface{}:
		// For complex types, marshal to JSON
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(data)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// ParseCLIVars parses CLI variable overrides in the format "key=value".
func ParseCLIVars(vars []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, v := range vars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid variable format '%s': expected 'key=value'", v)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("invalid variable format '%s': empty key", v)
		}

		result[key] = value
	}

	return result, nil
}
