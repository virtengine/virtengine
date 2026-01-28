package types

import (
	"context"
)

// QueryServer is the server API for the MFA module's queries.
// All methods follow Cosmos SDK gRPC patterns with context.Context as first parameter.
type QueryServer interface {
	// GetMFAPolicy returns the MFA policy for an account
	GetMFAPolicy(context.Context, *QueryMFAPolicyRequest) (*QueryMFAPolicyResponse, error)

	// GetFactorEnrollments returns all factor enrollments for an account
	GetFactorEnrollments(context.Context, *QueryFactorEnrollmentsRequest) (*QueryFactorEnrollmentsResponse, error)

	// GetFactorEnrollment returns a specific factor enrollment
	GetFactorEnrollment(context.Context, *QueryFactorEnrollmentRequest) (*QueryFactorEnrollmentResponse, error)

	// GetChallenge returns a challenge by ID
	GetChallenge(context.Context, *QueryChallengeRequest) (*QueryChallengeResponse, error)

	// GetPendingChallenges returns pending challenges for an account
	GetPendingChallenges(context.Context, *QueryPendingChallengesRequest) (*QueryPendingChallengesResponse, error)

	// GetAuthorizationSession returns an authorization session by ID
	GetAuthorizationSession(context.Context, *QueryAuthorizationSessionRequest) (*QueryAuthorizationSessionResponse, error)

	// GetTrustedDevices returns trusted devices for an account
	GetTrustedDevices(context.Context, *QueryTrustedDevicesRequest) (*QueryTrustedDevicesResponse, error)

	// GetSensitiveTxConfig returns the configuration for a sensitive tx type
	GetSensitiveTxConfig(context.Context, *QuerySensitiveTxConfigRequest) (*QuerySensitiveTxConfigResponse, error)

	// GetAllSensitiveTxConfigs returns all sensitive tx configurations
	GetAllSensitiveTxConfigs(context.Context, *QueryAllSensitiveTxConfigsRequest) (*QueryAllSensitiveTxConfigsResponse, error)

	// GetParams returns the module parameters
	GetParams(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)

	// IsMFARequired checks if MFA is required for a transaction
	IsMFARequired(context.Context, *QueryMFARequiredRequest) (*QueryMFARequiredResponse, error)
}

// QueryMFAPolicyRequest is the request for QueryMFAPolicy
type QueryMFAPolicyRequest struct {
	Address string `json:"address"`
}

// QueryMFAPolicyResponse is the response for QueryMFAPolicy
type QueryMFAPolicyResponse struct {
	Policy *MFAPolicy `json:"policy"`
}

// QueryFactorEnrollmentsRequest is the request for QueryFactorEnrollments
type QueryFactorEnrollmentsRequest struct {
	Address string `json:"address"`
}

// QueryFactorEnrollmentsResponse is the response for QueryFactorEnrollments
type QueryFactorEnrollmentsResponse struct {
	Enrollments []FactorEnrollment `json:"enrollments"`
}

// QueryChallengeRequest is the request for QueryChallenge
type QueryChallengeRequest struct {
	ChallengeID string `json:"challenge_id"`
}

// QueryChallengeResponse is the response for QueryChallenge
type QueryChallengeResponse struct {
	Challenge *Challenge `json:"challenge"`
}

// QueryPendingChallengesRequest is the request for QueryPendingChallenges
type QueryPendingChallengesRequest struct {
	Address string `json:"address"`
}

// QueryPendingChallengesResponse is the response for QueryPendingChallenges
type QueryPendingChallengesResponse struct {
	Challenges []Challenge `json:"challenges"`
}

// QueryTrustedDevicesRequest is the request for QueryTrustedDevices
type QueryTrustedDevicesRequest struct {
	Address string `json:"address"`
}

// QueryTrustedDevicesResponse is the response for QueryTrustedDevices
type QueryTrustedDevicesResponse struct {
	Devices []TrustedDevice `json:"devices"`
}

// QuerySensitiveTxConfigRequest is the request for QuerySensitiveTxConfig
type QuerySensitiveTxConfigRequest struct {
	TransactionType SensitiveTransactionType `json:"transaction_type"`
}

// QuerySensitiveTxConfigResponse is the response for QuerySensitiveTxConfig
type QuerySensitiveTxConfigResponse struct {
	Config *SensitiveTxConfig `json:"config"`
}

// QueryMFARequiredRequest is the request for QueryMFARequired
type QueryMFARequiredRequest struct {
	Address         string                   `json:"address"`
	TransactionType SensitiveTransactionType `json:"transaction_type"`
}

// QueryMFARequiredResponse is the response for QueryMFARequired
type QueryMFARequiredResponse struct {
	Required           bool                `json:"required"`
	FactorCombinations []FactorCombination `json:"factor_combinations"`
	MinVEIDScore       uint32              `json:"min_veid_score"`
}

// QueryParamsRequest is the request for QueryParams
type QueryParamsRequest struct{}

// QueryParamsResponse is the response for QueryParams
type QueryParamsResponse struct {
	Params Params `json:"params"`
}

// QueryFactorEnrollmentRequest is the request for GetFactorEnrollment
type QueryFactorEnrollmentRequest struct {
	Address    string     `json:"address"`
	FactorType FactorType `json:"factor_type"`
	FactorID   string     `json:"factor_id"`
}

// QueryFactorEnrollmentResponse is the response for GetFactorEnrollment
type QueryFactorEnrollmentResponse struct {
	Enrollment *FactorEnrollment `json:"enrollment"`
}

// QueryAuthorizationSessionRequest is the request for GetAuthorizationSession
type QueryAuthorizationSessionRequest struct {
	SessionID string `json:"session_id"`
}

// QueryAuthorizationSessionResponse is the response for GetAuthorizationSession
type QueryAuthorizationSessionResponse struct {
	Session *AuthorizationSession `json:"session"`
}

// QueryAllSensitiveTxConfigsRequest is the request for GetAllSensitiveTxConfigs
type QueryAllSensitiveTxConfigsRequest struct{}

// QueryAllSensitiveTxConfigsResponse is the response for GetAllSensitiveTxConfigs
type QueryAllSensitiveTxConfigsResponse struct {
	Configs []SensitiveTxConfig `json:"configs"`
}
