package types

// QueryServer is the server API for the MFA module's queries
type QueryServer interface {
	// GetMFAPolicy returns the MFA policy for an account
	GetMFAPolicy(address string) (*MFAPolicy, error)

	// GetFactorEnrollments returns all factor enrollments for an account
	GetFactorEnrollments(address string) ([]FactorEnrollment, error)

	// GetFactorEnrollment returns a specific factor enrollment
	GetFactorEnrollment(address string, factorType FactorType, factorID string) (*FactorEnrollment, error)

	// GetChallenge returns a challenge by ID
	GetChallenge(challengeID string) (*Challenge, error)

	// GetPendingChallenges returns pending challenges for an account
	GetPendingChallenges(address string) ([]Challenge, error)

	// GetAuthorizationSession returns an authorization session by ID
	GetAuthorizationSession(sessionID string) (*AuthorizationSession, error)

	// GetTrustedDevices returns trusted devices for an account
	GetTrustedDevices(address string) ([]TrustedDevice, error)

	// GetSensitiveTxConfig returns the configuration for a sensitive tx type
	GetSensitiveTxConfig(txType SensitiveTransactionType) (*SensitiveTxConfig, error)

	// GetAllSensitiveTxConfigs returns all sensitive tx configurations
	GetAllSensitiveTxConfigs() ([]SensitiveTxConfig, error)

	// GetParams returns the module parameters
	GetParams() Params

	// IsMFARequired checks if MFA is required for a transaction
	IsMFARequired(address string, txType SensitiveTransactionType) (bool, []FactorCombination, error)
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
