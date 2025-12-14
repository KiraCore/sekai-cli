package types

// TxResult represents the result of a transaction.
// This is a convenience type that mirrors sdk.TxResponse for use in modules.
type TxResult struct {
	// TxHash is the transaction hash
	TxHash string `json:"txhash"`

	// Code is the response code (0 = success)
	Code uint32 `json:"code"`

	// Height is the block height where the tx was included
	Height int64 `json:"height"`

	// GasUsed is the amount of gas used
	GasUsed int64 `json:"gas_used"`

	// GasWanted is the amount of gas requested
	GasWanted int64 `json:"gas_wanted"`

	// RawLog contains the raw log output
	RawLog string `json:"raw_log"`

	// Data contains any returned data
	Data string `json:"data,omitempty"`
}

// IsSuccess returns true if the transaction succeeded (code == 0).
func (r TxResult) IsSuccess() bool {
	return r.Code == 0
}

// BroadcastMode specifies how a transaction should be broadcast.
type BroadcastMode string

const (
	// BroadcastModeSync waits for CheckTx to complete.
	BroadcastModeSync BroadcastMode = "sync"

	// BroadcastModeAsync returns immediately without waiting.
	BroadcastModeAsync BroadcastMode = "async"

	// BroadcastModeBlock waits for the tx to be included in a block.
	BroadcastModeBlock BroadcastMode = "block"
)

// String returns the string representation of the broadcast mode.
func (m BroadcastMode) String() string {
	return string(m)
}

// TxOptions contains common transaction options.
type TxOptions struct {
	// Fees is the transaction fee (e.g., "1000ukex")
	Fees string

	// Gas is the gas limit (e.g., "auto", "200000")
	Gas string

	// GasAdjustment is the gas adjustment factor (e.g., 1.2)
	GasAdjustment float64

	// GasPrices is the gas price (e.g., "0.025ukex")
	GasPrices string

	// Memo is the transaction memo
	Memo string

	// BroadcastMode specifies how to broadcast the transaction
	BroadcastMode BroadcastMode

	// TimeoutHeight is the block height after which the tx is invalid
	TimeoutHeight uint64

	// Sequence is the account sequence number (optional, auto-fetched if not set)
	Sequence uint64

	// AccountNumber is the account number (optional, auto-fetched if not set)
	AccountNumber uint64
}

// DefaultTxOptions returns default transaction options.
func DefaultTxOptions() *TxOptions {
	return &TxOptions{
		Gas:           "auto",
		GasAdjustment: 1.3,
		BroadcastMode: BroadcastModeSync,
	}
}

// WithFees sets the fees option.
func (o *TxOptions) WithFees(fees string) *TxOptions {
	o.Fees = fees
	return o
}

// WithGas sets the gas option.
func (o *TxOptions) WithGas(gas string) *TxOptions {
	o.Gas = gas
	return o
}

// WithGasAdjustment sets the gas adjustment option.
func (o *TxOptions) WithGasAdjustment(adj float64) *TxOptions {
	o.GasAdjustment = adj
	return o
}

// WithMemo sets the memo option.
func (o *TxOptions) WithMemo(memo string) *TxOptions {
	o.Memo = memo
	return o
}

// WithBroadcastMode sets the broadcast mode option.
func (o *TxOptions) WithBroadcastMode(mode BroadcastMode) *TxOptions {
	o.BroadcastMode = mode
	return o
}

// ToFlags converts TxOptions to a map of flag values.
func (o *TxOptions) ToFlags() map[string]string {
	if o == nil {
		return nil
	}

	flags := make(map[string]string)

	if o.Fees != "" {
		flags["fees"] = o.Fees
	}
	if o.Gas != "" {
		flags["gas"] = o.Gas
	}
	if o.GasAdjustment > 0 {
		flags["gas-adjustment"] = formatFloat(o.GasAdjustment)
	}
	if o.GasPrices != "" {
		flags["gas-prices"] = o.GasPrices
	}
	if o.Memo != "" {
		flags["memo"] = o.Memo
	}
	if o.BroadcastMode != "" {
		flags["broadcast-mode"] = string(o.BroadcastMode)
	}
	if o.TimeoutHeight > 0 {
		flags["timeout-height"] = formatUint(o.TimeoutHeight)
	}
	if o.Sequence > 0 {
		flags["sequence"] = formatUint(o.Sequence)
	}
	if o.AccountNumber > 0 {
		flags["account-number"] = formatUint(o.AccountNumber)
	}

	return flags
}

// formatFloat formats a float64 without trailing zeros.
func formatFloat(f float64) string {
	s := make([]byte, 0, 24)
	s = append(s, []byte{byte(int(f) + '0')}...)
	if f != float64(int(f)) {
		s = append(s, '.')
		frac := f - float64(int(f))
		for i := 0; i < 6 && frac > 0; i++ {
			frac *= 10
			s = append(s, byte(int(frac)+'0'))
			frac -= float64(int(frac))
		}
	}
	return string(s)
}

// formatUint formats a uint64 to string.
func formatUint(u uint64) string {
	if u == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for u > 0 {
		i--
		buf[i] = byte(u%10) + '0'
		u /= 10
	}
	return string(buf[i:])
}
