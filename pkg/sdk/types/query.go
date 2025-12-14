package types

// PageRequest is the request for paginated queries.
type PageRequest struct {
	// Key is the pagination key to start from
	Key []byte `json:"key,omitempty"`

	// Offset is the numeric offset to start from (mutually exclusive with Key)
	Offset uint64 `json:"offset,omitempty"`

	// Limit is the maximum number of results to return
	Limit uint64 `json:"limit,omitempty"`

	// CountTotal indicates whether to return total count
	CountTotal bool `json:"count_total,omitempty"`

	// Reverse indicates whether to iterate in reverse
	Reverse bool `json:"reverse,omitempty"`
}

// PageResponse is the response for paginated queries.
type PageResponse struct {
	// NextKey is the key for the next page
	NextKey []byte `json:"next_key,omitempty"`

	// Total is the total number of results (if CountTotal was set)
	Total uint64 `json:"total,omitempty"`
}

// DefaultPageRequest returns a default page request with reasonable defaults.
func DefaultPageRequest() *PageRequest {
	return &PageRequest{
		Limit:      100,
		CountTotal: false,
	}
}

// WithLimit sets the limit.
func (p *PageRequest) WithLimit(limit uint64) *PageRequest {
	p.Limit = limit
	return p
}

// WithOffset sets the offset.
func (p *PageRequest) WithOffset(offset uint64) *PageRequest {
	p.Offset = offset
	return p
}

// WithCountTotal sets whether to count total.
func (p *PageRequest) WithCountTotal(count bool) *PageRequest {
	p.CountTotal = count
	return p
}

// WithReverse sets whether to reverse iteration.
func (p *PageRequest) WithReverse(reverse bool) *PageRequest {
	p.Reverse = reverse
	return p
}

// ToParams converts PageRequest to query parameters.
func (p *PageRequest) ToParams() map[string]string {
	if p == nil {
		return nil
	}

	params := make(map[string]string)

	if len(p.Key) > 0 {
		params["pagination.key"] = string(p.Key)
	}
	if p.Offset > 0 {
		params["pagination.offset"] = formatUint(p.Offset)
	}
	if p.Limit > 0 {
		params["pagination.limit"] = formatUint(p.Limit)
	}
	if p.CountTotal {
		params["pagination.count_total"] = "true"
	}
	if p.Reverse {
		params["pagination.reverse"] = "true"
	}

	return params
}

// QueryOptions contains common query options.
type QueryOptions struct {
	// Height is the block height to query at (0 = latest)
	Height int64

	// Prove indicates whether to include merkle proofs
	Prove bool
}

// DefaultQueryOptions returns default query options.
func DefaultQueryOptions() *QueryOptions {
	return &QueryOptions{
		Height: 0,
		Prove:  false,
	}
}

// WithHeight sets the query height.
func (o *QueryOptions) WithHeight(height int64) *QueryOptions {
	o.Height = height
	return o
}

// WithProve sets whether to include proofs.
func (o *QueryOptions) WithProve(prove bool) *QueryOptions {
	o.Prove = prove
	return o
}

// ToParams converts QueryOptions to query parameters.
func (o *QueryOptions) ToParams() map[string]string {
	if o == nil {
		return nil
	}

	params := make(map[string]string)

	if o.Height > 0 {
		params["height"] = formatInt(o.Height)
	}
	if o.Prove {
		params["prove"] = "true"
	}

	return params
}

// formatInt formats an int64 to string.
func formatInt(i int64) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	var buf [20]byte
	j := len(buf)
	for i > 0 {
		j--
		buf[j] = byte(i%10) + '0'
		i /= 10
	}
	if neg {
		j--
		buf[j] = '-'
	}
	return string(buf[j:])
}
