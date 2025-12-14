package sdk

import (
	"errors"
	"fmt"
)

// Common SDK errors.
var (
	// ErrNotConnected indicates the client is not connected.
	ErrNotConnected = errors.New("client not connected")

	// ErrTimeout indicates an operation timed out.
	ErrTimeout = errors.New("operation timed out")

	// ErrInvalidAddress indicates an invalid blockchain address.
	ErrInvalidAddress = errors.New("invalid address")

	// ErrInvalidAmount indicates an invalid coin amount.
	ErrInvalidAmount = errors.New("invalid amount")

	// ErrKeyNotFound indicates a key was not found in the keyring.
	ErrKeyNotFound = errors.New("key not found")

	// ErrKeyExists indicates a key already exists.
	ErrKeyExists = errors.New("key already exists")

	// ErrInsufficientFunds indicates insufficient funds for a transaction.
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrUnauthorized indicates the signer is not authorized.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrNotSupported indicates the operation is not supported by this client.
	ErrNotSupported = errors.New("operation not supported")

	// ErrInvalidResponse indicates an invalid response from the node.
	ErrInvalidResponse = errors.New("invalid response")

	// ErrTxFailed indicates a transaction failed.
	ErrTxFailed = errors.New("transaction failed")
)

// QueryError represents an error during a query operation.
type QueryError struct {
	// Module is the module that was queried
	Module string

	// Endpoint is the endpoint that was queried
	Endpoint string

	// Err is the underlying error
	Err error
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("query %s/%s failed: %v", e.Module, e.Endpoint, e.Err)
}

func (e *QueryError) Unwrap() error {
	return e.Err
}

// TxError represents an error during a transaction.
type TxError struct {
	// Module is the module for the transaction
	Module string

	// Action is the action that was attempted
	Action string

	// Code is the error code (if any)
	Code uint32

	// Err is the underlying error
	Err error

	// RawLog contains the raw error log from the node
	RawLog string
}

func (e *TxError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("tx %s/%s failed (code %d): %s", e.Module, e.Action, e.Code, e.RawLog)
	}
	return fmt.Sprintf("tx %s/%s failed: %v", e.Module, e.Action, e.Err)
}

func (e *TxError) Unwrap() error {
	return e.Err
}

// ExecutionError represents an error during command execution (Docker client).
type ExecutionError struct {
	// Command is the command that was executed
	Command string

	// Args are the command arguments
	Args []string

	// ExitCode is the exit code
	ExitCode int

	// Stderr contains the stderr output
	Stderr string

	// Err is the underlying error
	Err error
}

func (e *ExecutionError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("command failed (exit %d): %s", e.ExitCode, e.Stderr)
	}
	if e.Err != nil {
		return fmt.Sprintf("command execution failed: %v", e.Err)
	}
	return fmt.Sprintf("command failed with exit code %d", e.ExitCode)
}

func (e *ExecutionError) Unwrap() error {
	return e.Err
}

// HTTPError represents an error during HTTP communication (REST client).
type HTTPError struct {
	// StatusCode is the HTTP status code
	StatusCode int

	// Status is the HTTP status text
	Status string

	// Body contains the response body
	Body string

	// URL is the requested URL
	URL string

	// Err is the underlying error
	Err error
}

func (e *HTTPError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("HTTP %d %s: %s", e.StatusCode, e.Status, e.Body)
	}
	return fmt.Sprintf("HTTP request to %s failed: %v", e.URL, e.Err)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// WrapQueryError wraps an error as a QueryError.
func WrapQueryError(module, endpoint string, err error) error {
	if err == nil {
		return nil
	}
	return &QueryError{
		Module:   module,
		Endpoint: endpoint,
		Err:      err,
	}
}

// WrapTxError wraps an error as a TxError.
func WrapTxError(module, action string, err error) error {
	if err == nil {
		return nil
	}
	return &TxError{
		Module: module,
		Action: action,
		Err:    err,
	}
}

// NewTxErrorFromResponse creates a TxError from a TxResponse.
func NewTxErrorFromResponse(module, action string, resp *TxResponse) error {
	if resp.Code == 0 {
		return nil
	}
	return &TxError{
		Module: module,
		Action: action,
		Code:   resp.Code,
		RawLog: resp.RawLog,
		Err:    ErrTxFailed,
	}
}
