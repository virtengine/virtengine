package types

// ============================================================================
// Assurance Vector Query Request/Response Types
// ============================================================================
//
// These are the local (hand-written) query types, matching the convention of
// score_query.go: the module keeps a local request/response shape per query and
// the SDK query server converts to the generated protobuf types. The conversion
// lives in keeper/sdk_query_server.go.
//
// Absence is explicit. A response with Vector == nil and Found == false means
// the account has never produced an assurance vector — that is "no assurance
// claim", not a vector of zeros.

// QueryAssuranceVectorRequest is the request for QueryAssuranceVector
type QueryAssuranceVectorRequest struct {
	// AccountAddress is the address to query the assurance vector for
	AccountAddress string `json:"account_address"`
}

// QueryAssuranceVectorResponse is the response for QueryAssuranceVector
type QueryAssuranceVectorResponse struct {
	// Vector is the account's current assurance vector, or nil when the account
	// has never produced one
	Vector *AssuranceVector `json:"vector,omitempty"`

	// Found indicates whether a vector exists for the account
	Found bool `json:"found"`

	// Recency describes how current the vector's evidence is, evaluated against
	// consensus time
	Recency *AssuranceVectorRecency `json:"recency,omitempty"`

	// CurrentEpoch is the epoch of the account's newest vector. A relying party
	// holding a vector at a lower epoch can detect that it was superseded.
	CurrentEpoch uint64 `json:"current_epoch"`
}

// QueryAssuranceVectorHistoryRequest is the request for
// QueryAssuranceVectorHistory
type QueryAssuranceVectorHistoryRequest struct {
	// AccountAddress is the address to query the vector history for
	AccountAddress string `json:"account_address"`
}

// QueryAssuranceVectorHistoryResponse is the response for
// QueryAssuranceVectorHistory
type QueryAssuranceVectorHistoryResponse struct {
	// Vectors holds the account's vectors, newest first. An account with no
	// vector returns an empty slice, not nil, so clients can range over it.
	Vectors []*AssuranceVector `json:"vectors"`
}

// ============================================================================
// Proto message shims
// ============================================================================
//
// The module's generated protobuf types use the gogo-style message interface.
// These satisfy the same interface for the local types so they can be used in
// the same generic machinery. String() intentionally returns "" to match the
// convention already used for the other query types in score_query.go.

func (m *QueryAssuranceVectorRequest) Reset()         { *m = QueryAssuranceVectorRequest{} }
func (m *QueryAssuranceVectorRequest) String() string { return "" }
func (*QueryAssuranceVectorRequest) ProtoMessage()    {}

func (m *QueryAssuranceVectorResponse) Reset()         { *m = QueryAssuranceVectorResponse{} }
func (m *QueryAssuranceVectorResponse) String() string { return "" }
func (*QueryAssuranceVectorResponse) ProtoMessage()    {}

func (m *QueryAssuranceVectorHistoryRequest) Reset()         { *m = QueryAssuranceVectorHistoryRequest{} }
func (m *QueryAssuranceVectorHistoryRequest) String() string { return "" }
func (*QueryAssuranceVectorHistoryRequest) ProtoMessage()    {}

func (m *QueryAssuranceVectorHistoryResponse) Reset()         { *m = QueryAssuranceVectorHistoryResponse{} }
func (m *QueryAssuranceVectorHistoryResponse) String() string { return "" }
func (*QueryAssuranceVectorHistoryResponse) ProtoMessage()    {}

func (m *AssuranceVectorRecency) Reset()         { *m = AssuranceVectorRecency{} }
func (m *AssuranceVectorRecency) String() string { return "" }
func (*AssuranceVectorRecency) ProtoMessage()    {}

func (m *AssuranceFactorEntry) Reset()         { *m = AssuranceFactorEntry{} }
func (m *AssuranceFactorEntry) String() string { return "" }
func (*AssuranceFactorEntry) ProtoMessage()    {}
