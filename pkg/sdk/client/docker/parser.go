package docker

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kiracore/sekai-cli/pkg/sdk"
	"github.com/kiracore/sekai-cli/pkg/sdk/types"
)

// ParseTxResponse parses a transaction response from sekaid output.
func ParseTxResponse(output string) (*sdk.TxResponse, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("empty response")
	}

	var resp sdk.TxResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		// Try to extract txhash from text output
		if hash := extractTxHash(output); hash != "" {
			return &sdk.TxResponse{TxHash: hash}, nil
		}
		return nil, fmt.Errorf("failed to parse transaction response: %w", err)
	}

	return &resp, nil
}

// ParseStatusResponse parses a status response from sekaid output.
func ParseStatusResponse(output string) (*sdk.StatusResponse, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("empty response")
	}

	// The status response can be nested, try different structures
	var resp sdk.StatusResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		// Try nested structure
		var nested struct {
			Result sdk.StatusResponse `json:"result"`
		}
		if err2 := json.Unmarshal([]byte(output), &nested); err2 != nil {
			return nil, fmt.Errorf("failed to parse status response: %w", err)
		}
		return &nested.Result, nil
	}

	return &resp, nil
}

// ParseKeyInfo parses key information from sekaid output.
func ParseKeyInfo(output string) (*sdk.KeyInfo, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("empty response")
	}

	var info sdk.KeyInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		// Try to parse YAML-like format
		parsed, err := parseYAMLKeyInfo(output)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key info: %w", err)
		}
		return parsed, nil
	}

	return &info, nil
}

// ParseKeyInfoList parses a list of key information.
func ParseKeyInfoList(output string) ([]sdk.KeyInfo, error) {
	output = strings.TrimSpace(output)
	if output == "" || output == "[]" || output == "null" {
		return []sdk.KeyInfo{}, nil
	}

	var keys []sdk.KeyInfo
	if err := json.Unmarshal([]byte(output), &keys); err != nil {
		return nil, fmt.Errorf("failed to parse key list: %w", err)
	}

	return keys, nil
}

// ParseBalance parses balance information.
func ParseBalance(output string) (types.Coins, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	// Try to parse as JSON
	var resp struct {
		Balances []types.Coin `json:"balances"`
	}
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		// Try direct array
		var coins []types.Coin
		if err2 := json.Unmarshal([]byte(output), &coins); err2 != nil {
			return nil, fmt.Errorf("failed to parse balance: %w", err)
		}
		return types.Coins(coins), nil
	}

	return types.Coins(resp.Balances), nil
}

// ParseQueryResponse parses a generic query response.
func ParseQueryResponse[T any](output string) (*T, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("empty response")
	}

	var result T
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// extractTxHash attempts to extract a transaction hash from text output.
func extractTxHash(output string) string {
	// Look for common patterns
	patterns := []string{
		`txhash:\s*([A-Fa-f0-9]{64})`,
		`"txhash":\s*"([A-Fa-f0-9]{64})"`,
		`TxHash:\s*([A-Fa-f0-9]{64})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// parseYAMLKeyInfo parses key info from YAML-like text format.
func parseYAMLKeyInfo(output string) (*sdk.KeyInfo, error) {
	info := &sdk.KeyInfo{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "name":
			info.Name = value
		case "type":
			info.Type = value
		case "address":
			info.Address = value
		case "pubkey":
			info.PubKey = value
		case "mnemonic":
			info.Mnemonic = value
		}
	}

	if info.Name == "" && info.Address == "" {
		return nil, fmt.Errorf("could not parse key info from output")
	}

	return info, nil
}

// ParseErrorResponse attempts to parse an error from sekaid output.
func ParseErrorResponse(output string) (code int, message string) {
	// Try JSON error format
	var jsonErr struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		RawLog  string `json:"raw_log"`
	}
	if err := json.Unmarshal([]byte(output), &jsonErr); err == nil {
		if jsonErr.Message != "" {
			return jsonErr.Code, jsonErr.Message
		}
		if jsonErr.RawLog != "" {
			return jsonErr.Code, jsonErr.RawLog
		}
	}

	// Try to extract error from text
	patterns := []string{
		`error:\s*(.+)`,
		`Error:\s*(.+)`,
		`failed:\s*(.+)`,
		`code:\s*(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(output); len(matches) > 1 {
			if pattern == `code:\s*(\d+)` {
				code, _ = strconv.Atoi(matches[1])
			} else {
				message = matches[1]
			}
		}
	}

	if message == "" {
		message = output
	}

	return code, message
}

// ExtractHeight extracts block height from a response.
func ExtractHeight(output string) (int64, error) {
	patterns := []string{
		`"height":\s*"?(\d+)"?`,
		`height:\s*(\d+)`,
		`Height:\s*(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(output); len(matches) > 1 {
			return strconv.ParseInt(matches[1], 10, 64)
		}
	}

	return 0, fmt.Errorf("height not found in output")
}
