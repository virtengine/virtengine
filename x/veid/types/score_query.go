package types

import (
	"time"

	query "github.com/cosmos/cosmos-sdk/types/query"
)

// ============================================================================
// Score Query Request/Response Types
// ============================================================================

// QueryIdentityScoreRequest is the request for QueryIdentityScore
type QueryIdentityScoreRequest struct {
	// AccountAddress is the address to query the score for
	AccountAddress string `json:"account_address"`
}

// QueryIdentityScoreResponse is the response for QueryIdentityScore
type QueryIdentityScoreResponse struct {
	// Score is the identity score details
	Score *IdentityScore `json:"score,omitempty"`

	// Found indicates if a score was found for the account
	Found bool `json:"found"`
}

// QueryIdentityStatusRequest is the request for QueryIdentityStatus
type QueryIdentityStatusRequest struct {
	// AccountAddress is the address to query the status for
	AccountAddress string `json:"account_address"`
}

// QueryIdentityStatusResponse is the response for QueryIdentityStatus
type QueryIdentityStatusResponse struct {
	// AccountAddress is the queried address
	AccountAddress string `json:"account_address"`

	// Status is the current verification status
	Status AccountStatus `json:"status"`

	// Tier is the current identity tier (0-3)
	Tier int `json:"tier"`

	// TierName is the human-readable tier name
	TierName string `json:"tier_name"`

	// Score is the current score (if verified)
	Score uint32 `json:"score"`

	// ModelVersion is the ML model version used
	ModelVersion string `json:"model_version,omitempty"`

	// LastUpdatedAt is when the status was last updated
	LastUpdatedAt *time.Time `json:"last_updated_at,omitempty"`

	// Found indicates if the account exists
	Found bool `json:"found"`
}

// QueryScoreHistoryRequest is the request for QueryScoreHistory
type QueryScoreHistoryRequest struct {
	// AccountAddress is the address to query history for
	AccountAddress string `json:"account_address"`

	// Pagination is the pagination options
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// QueryScoreHistoryResponse is the response for QueryScoreHistory
type QueryScoreHistoryResponse struct {
	// AccountAddress is the queried address
	AccountAddress string `json:"account_address"`

	// Entries are the score history entries
	Entries []ScoreHistoryEntry `json:"entries"`

	// Pagination is the pagination response
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// QueryRequiredScopesRequest is the request for QueryRequiredScopes
type QueryRequiredScopesRequest struct {
	// OfferingType is the type of offering to query requirements for
	OfferingType OfferingType `json:"offering_type"`
}

// QueryRequiredScopesResponse is the response for QueryRequiredScopes
type QueryRequiredScopesResponse struct {
	// Requirements are the required scopes for the offering type
	Requirements RequiredScopes `json:"requirements"`
}

// QueryAllRequiredScopesRequest is the request for all offering types
type QueryAllRequiredScopesRequest struct{}

// QueryAllRequiredScopesResponse is the response for all offering types
type QueryAllRequiredScopesResponse struct {
	// AllRequirements maps offering types to their requirements
	AllRequirements []RequiredScopes `json:"all_requirements"`
}

// QueryAccountsByScoreTierRequest is the request for QueryAccountsByScoreTier
type QueryAccountsByScoreTierRequest struct {
	// Tier is the tier to filter by (0-3)
	Tier int `json:"tier"`

	// MinScore is an optional minimum score filter
	MinScore *uint32 `json:"min_score,omitempty"`

	// MaxScore is an optional maximum score filter
	MaxScore *uint32 `json:"max_score,omitempty"`

	// Status is an optional status filter
	Status *AccountStatus `json:"status,omitempty"`

	// Pagination is the pagination options
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

// AccountScoreSummary is a summary of an account's score for listing
type AccountScoreSummary struct {
	// AccountAddress is the account address
	AccountAddress string `json:"account_address"`

	// Score is the current score
	Score uint32 `json:"score"`

	// Status is the verification status
	Status AccountStatus `json:"status"`

	// Tier is the computed tier
	Tier int `json:"tier"`

	// ModelVersion is the ML model version
	ModelVersion string `json:"model_version,omitempty"`

	// ComputedAt is when the score was computed
	ComputedAt time.Time `json:"computed_at"`
}

// QueryAccountsByScoreTierResponse is the response for QueryAccountsByScoreTier
type QueryAccountsByScoreTierResponse struct {
	// Accounts are the matching accounts
	Accounts []AccountScoreSummary `json:"accounts"`

	// Total is the total number of matching accounts
	Total uint64 `json:"total"`

	// Pagination is the pagination response
	Pagination *query.PageResponse `json:"pagination,omitempty"`
}

// QueryScoreThresholdCheckRequest checks if an account meets a score threshold
type QueryScoreThresholdCheckRequest struct {
	// AccountAddress is the address to check
	AccountAddress string `json:"account_address"`

	// Threshold is the minimum score threshold
	Threshold uint32 `json:"threshold"`

	// RequireVerified indicates if the account must be verified
	RequireVerified bool `json:"require_verified"`
}

// QueryScoreThresholdCheckResponse is the response for threshold check
type QueryScoreThresholdCheckResponse struct {
	// AccountAddress is the queried address
	AccountAddress string `json:"account_address"`

	// MeetsThreshold indicates if the account meets the threshold
	MeetsThreshold bool `json:"meets_threshold"`

	// CurrentScore is the account's current score
	CurrentScore uint32 `json:"current_score"`

	// CurrentStatus is the account's current status
	CurrentStatus AccountStatus `json:"current_status"`

	// Threshold is the threshold that was checked
	Threshold uint32 `json:"threshold"`

	// Found indicates if the account exists
	Found bool `json:"found"`
}

// QueryEligibilityRequest checks eligibility for an offering type
type QueryEligibilityRequest struct {
	// AccountAddress is the address to check eligibility for
	AccountAddress string `json:"account_address"`

	// OfferingType is the offering type to check eligibility for
	OfferingType OfferingType `json:"offering_type"`
}

// EligibilityResult contains detailed eligibility information
type EligibilityResult struct {
	// Eligible indicates if the account is eligible
	Eligible bool `json:"eligible"`

	// AccountAddress is the queried address
	AccountAddress string `json:"account_address"`

	// OfferingType is the offering type checked
	OfferingType OfferingType `json:"offering_type"`

	// CurrentScore is the account's current score
	CurrentScore uint32 `json:"current_score"`

	// RequiredScore is the minimum required score
	RequiredScore uint32 `json:"required_score"`

	// CurrentStatus is the account's current status
	CurrentStatus AccountStatus `json:"current_status"`

	// RequiresMFA indicates if MFA is required
	RequiresMFA bool `json:"requires_mfa"`

	// MissingScopes are scopes that need to be verified
	MissingScopes []ScopeType `json:"missing_scopes,omitempty"`

	// Reason is a human-readable eligibility explanation
	Reason string `json:"reason"`
}

// QueryEligibilityResponse is the response for eligibility check
type QueryEligibilityResponse struct {
	// Result is the eligibility result
	Result EligibilityResult `json:"result"`
}

// ============================================================================
// Score Statistics Types (for analytics)
// ============================================================================

// ScoreStatistics contains aggregate statistics about scores
type ScoreStatistics struct {
	// TotalAccounts is the total number of accounts with scores
	TotalAccounts uint64 `json:"total_accounts"`

	// TierCounts maps tier numbers to account counts
	TierCounts map[int]uint64 `json:"tier_counts"`

	// StatusCounts maps statuses to account counts
	StatusCounts map[AccountStatus]uint64 `json:"status_counts"`

	// AverageScore is the average score across all verified accounts
	AverageScore float64 `json:"average_score"`

	// MedianScore is the median score across all verified accounts
	MedianScore uint32 `json:"median_score"`

	// ComputedAt is when these statistics were computed
	ComputedAt time.Time `json:"computed_at"`
}

// QueryScoreStatisticsRequest is the request for score statistics
type QueryScoreStatisticsRequest struct{}

// QueryScoreStatisticsResponse is the response for score statistics
type QueryScoreStatisticsResponse struct {
	// Statistics contains the computed statistics
	Statistics ScoreStatistics `json:"statistics"`
}

// ============================================================================
// proto.Message interface implementations for Score Query types
// Required for Cosmos SDK gRPC router registration
// ============================================================================

func (m *QueryIdentityScoreRequest) Reset()         { *m = QueryIdentityScoreRequest{} }
func (m *QueryIdentityScoreRequest) String() string { return "" }
func (*QueryIdentityScoreRequest) ProtoMessage()    {}

func (m *QueryIdentityScoreResponse) Reset()         { *m = QueryIdentityScoreResponse{} }
func (m *QueryIdentityScoreResponse) String() string { return "" }
func (*QueryIdentityScoreResponse) ProtoMessage()    {}

func (m *QueryIdentityStatusRequest) Reset()         { *m = QueryIdentityStatusRequest{} }
func (m *QueryIdentityStatusRequest) String() string { return "" }
func (*QueryIdentityStatusRequest) ProtoMessage()    {}

func (m *QueryIdentityStatusResponse) Reset()         { *m = QueryIdentityStatusResponse{} }
func (m *QueryIdentityStatusResponse) String() string { return "" }
func (*QueryIdentityStatusResponse) ProtoMessage()    {}

func (m *QueryScoreHistoryRequest) Reset()         { *m = QueryScoreHistoryRequest{} }
func (m *QueryScoreHistoryRequest) String() string { return "" }
func (*QueryScoreHistoryRequest) ProtoMessage()    {}

func (m *QueryScoreHistoryResponse) Reset()         { *m = QueryScoreHistoryResponse{} }
func (m *QueryScoreHistoryResponse) String() string { return "" }
func (*QueryScoreHistoryResponse) ProtoMessage()    {}

func (m *QueryRequiredScopesRequest) Reset()         { *m = QueryRequiredScopesRequest{} }
func (m *QueryRequiredScopesRequest) String() string { return "" }
func (*QueryRequiredScopesRequest) ProtoMessage()    {}

func (m *QueryRequiredScopesResponse) Reset()         { *m = QueryRequiredScopesResponse{} }
func (m *QueryRequiredScopesResponse) String() string { return "" }
func (*QueryRequiredScopesResponse) ProtoMessage()    {}

func (m *QueryAllRequiredScopesRequest) Reset()         { *m = QueryAllRequiredScopesRequest{} }
func (m *QueryAllRequiredScopesRequest) String() string { return "" }
func (*QueryAllRequiredScopesRequest) ProtoMessage()    {}

func (m *QueryAllRequiredScopesResponse) Reset()         { *m = QueryAllRequiredScopesResponse{} }
func (m *QueryAllRequiredScopesResponse) String() string { return "" }
func (*QueryAllRequiredScopesResponse) ProtoMessage()    {}

func (m *QueryAccountsByScoreTierRequest) Reset()         { *m = QueryAccountsByScoreTierRequest{} }
func (m *QueryAccountsByScoreTierRequest) String() string { return "" }
func (*QueryAccountsByScoreTierRequest) ProtoMessage()    {}

func (m *QueryAccountsByScoreTierResponse) Reset()         { *m = QueryAccountsByScoreTierResponse{} }
func (m *QueryAccountsByScoreTierResponse) String() string { return "" }
func (*QueryAccountsByScoreTierResponse) ProtoMessage()    {}

func (m *QueryScoreThresholdCheckRequest) Reset()         { *m = QueryScoreThresholdCheckRequest{} }
func (m *QueryScoreThresholdCheckRequest) String() string { return "" }
func (*QueryScoreThresholdCheckRequest) ProtoMessage()    {}

func (m *QueryScoreThresholdCheckResponse) Reset()         { *m = QueryScoreThresholdCheckResponse{} }
func (m *QueryScoreThresholdCheckResponse) String() string { return "" }
func (*QueryScoreThresholdCheckResponse) ProtoMessage()    {}

func (m *QueryEligibilityRequest) Reset()         { *m = QueryEligibilityRequest{} }
func (m *QueryEligibilityRequest) String() string { return "" }
func (*QueryEligibilityRequest) ProtoMessage()    {}

func (m *QueryEligibilityResponse) Reset()         { *m = QueryEligibilityResponse{} }
func (m *QueryEligibilityResponse) String() string { return "" }
func (*QueryEligibilityResponse) ProtoMessage()    {}

func (m *QueryScoreStatisticsRequest) Reset()         { *m = QueryScoreStatisticsRequest{} }
func (m *QueryScoreStatisticsRequest) String() string { return "" }
func (*QueryScoreStatisticsRequest) ProtoMessage()    {}

func (m *QueryScoreStatisticsResponse) Reset()         { *m = QueryScoreStatisticsResponse{} }
func (m *QueryScoreStatisticsResponse) String() string { return "" }
func (*QueryScoreStatisticsResponse) ProtoMessage()    {}
