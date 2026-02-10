package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "mfa"

	// StoreKey is the store key string for mfa module
	StoreKey = ModuleName

	// RouterKey is the message route for mfa module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for mfa module
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// PrefixMFAPolicy is the prefix for MFA policy storage
	// Key: PrefixMFAPolicy | address -> MFAPolicy
	PrefixMFAPolicy = []byte{0x01}

	// PrefixFactorEnrollment is the prefix for factor enrollment storage
	// Key: PrefixFactorEnrollment | address | factorType | factorID -> FactorEnrollment
	PrefixFactorEnrollment = []byte{0x02}

	// PrefixChallenge is the prefix for challenge storage
	// Key: PrefixChallenge | challengeID -> Challenge
	PrefixChallenge = []byte{0x03}

	// PrefixAccountChallenges is the prefix for account challenges index
	// Key: PrefixAccountChallenges | address | challengeID -> []byte{1}
	PrefixAccountChallenges = []byte{0x04}

	// PrefixTrustedDevice is the prefix for trusted device storage
	// Key: PrefixTrustedDevice | address | deviceFingerprint -> TrustedDevice
	PrefixTrustedDevice = []byte{0x05}

	// PrefixSensitiveTxConfig is the prefix for sensitive transaction configuration
	// Key: PrefixSensitiveTxConfig | txType -> SensitiveTxConfig
	PrefixSensitiveTxConfig = []byte{0x06}

	// PrefixAuthorizationSession is the prefix for authorization session storage
	// Key: PrefixAuthorizationSession | sessionID -> AuthorizationSession
	PrefixAuthorizationSession = []byte{0x07}

	// PrefixAccountSessions is the prefix for account sessions index
	// Key: PrefixAccountSessions | address | sessionID -> []byte{1}
	PrefixAccountSessions = []byte{0x08}

	// PrefixParams is the prefix for module parameters
	PrefixParams = []byte{0x09}

	// ============================================================================
	// Authorization Policy Keys (VE-221)
	// ============================================================================

	// PrefixAuthorizationPolicy is the prefix for authorization policy storage
	// Key: PrefixAuthorizationPolicy | address | policy_id -> AuthorizationPolicy
	PrefixAuthorizationPolicy = []byte{0x0A}

	// PrefixAuthorizationPolicyByAccount is the prefix for policy lookup by account
	// Key: PrefixAuthorizationPolicyByAccount | address -> []policy_id
	PrefixAuthorizationPolicyByAccount = []byte{0x0B}

	// PrefixAuthorizationAudit is the prefix for authorization audit events
	// Key: PrefixAuthorizationAudit | event_id -> AuthorizationAuditEvent
	PrefixAuthorizationAudit = []byte{0x0C}

	// PrefixAuthorizationAuditByAccount is the prefix for audit events by account
	// Key: PrefixAuthorizationAuditByAccount | address | timestamp -> event_id
	PrefixAuthorizationAuditByAccount = []byte{0x0D}
)

// MFAPolicyKey returns the store key for an MFA policy
func MFAPolicyKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixMFAPolicy)+len(address))
	key = append(key, PrefixMFAPolicy...)
	key = append(key, address...)
	return key
}

// FactorEnrollmentKey returns the store key for a factor enrollment
func FactorEnrollmentKey(address []byte, factorType FactorType, factorID string) []byte {
	factorIDBytes := []byte(factorID)
	key := make([]byte, 0, len(PrefixFactorEnrollment)+len(address)+1+len(factorIDBytes))
	key = append(key, PrefixFactorEnrollment...)
	key = append(key, address...)
	key = append(key, byte(factorType))
	key = append(key, factorIDBytes...)
	return key
}

// FactorEnrollmentPrefixKey returns the store key prefix for all factors of an address
func FactorEnrollmentPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixFactorEnrollment)+len(address))
	key = append(key, PrefixFactorEnrollment...)
	key = append(key, address...)
	return key
}

// FactorEnrollmentTypePrefixKey returns the store key prefix for a specific factor type of an address
func FactorEnrollmentTypePrefixKey(address []byte, factorType FactorType) []byte {
	key := make([]byte, 0, len(PrefixFactorEnrollment)+len(address)+1)
	key = append(key, PrefixFactorEnrollment...)
	key = append(key, address...)
	key = append(key, byte(factorType))
	return key
}

// ChallengeKey returns the store key for a challenge
func ChallengeKey(challengeID string) []byte {
	challengeIDBytes := []byte(challengeID)
	key := make([]byte, 0, len(PrefixChallenge)+len(challengeIDBytes))
	key = append(key, PrefixChallenge...)
	key = append(key, challengeIDBytes...)
	return key
}

// AccountChallengesKey returns the store key for an account challenge index entry
func AccountChallengesKey(address []byte, challengeID string) []byte {
	challengeIDBytes := []byte(challengeID)
	key := make([]byte, 0, len(PrefixAccountChallenges)+len(address)+len(challengeIDBytes))
	key = append(key, PrefixAccountChallenges...)
	key = append(key, address...)
	key = append(key, challengeIDBytes...)
	return key
}

// AccountChallengesPrefixKey returns the store key prefix for all challenges of an address
func AccountChallengesPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixAccountChallenges)+len(address))
	key = append(key, PrefixAccountChallenges...)
	key = append(key, address...)
	return key
}

// TrustedDeviceKey returns the store key for a trusted device
func TrustedDeviceKey(address []byte, deviceFingerprint string) []byte {
	fingerprintBytes := []byte(deviceFingerprint)
	key := make([]byte, 0, len(PrefixTrustedDevice)+len(address)+len(fingerprintBytes))
	key = append(key, PrefixTrustedDevice...)
	key = append(key, address...)
	key = append(key, fingerprintBytes...)
	return key
}

// TrustedDevicePrefixKey returns the store key prefix for all trusted devices of an address
func TrustedDevicePrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixTrustedDevice)+len(address))
	key = append(key, PrefixTrustedDevice...)
	key = append(key, address...)
	return key
}

// SensitiveTxConfigKey returns the store key for a sensitive transaction config
func SensitiveTxConfigKey(txType SensitiveTransactionType) []byte {
	key := make([]byte, 0, len(PrefixSensitiveTxConfig)+1)
	key = append(key, PrefixSensitiveTxConfig...)
	key = append(key, byte(txType))
	return key
}

// AuthorizationSessionKey returns the store key for an authorization session
func AuthorizationSessionKey(sessionID string) []byte {
	sessionIDBytes := []byte(sessionID)
	key := make([]byte, 0, len(PrefixAuthorizationSession)+len(sessionIDBytes))
	key = append(key, PrefixAuthorizationSession...)
	key = append(key, sessionIDBytes...)
	return key
}

// AccountSessionsKey returns the store key for an account session index entry
func AccountSessionsKey(address []byte, sessionID string) []byte {
	sessionIDBytes := []byte(sessionID)
	key := make([]byte, 0, len(PrefixAccountSessions)+len(address)+len(sessionIDBytes))
	key = append(key, PrefixAccountSessions...)
	key = append(key, address...)
	key = append(key, sessionIDBytes...)
	return key
}

// AccountSessionsPrefixKey returns the store key prefix for all sessions of an address
func AccountSessionsPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixAccountSessions)+len(address))
	key = append(key, PrefixAccountSessions...)
	key = append(key, address...)
	return key
}

// ParamsKey returns the store key for module parameters
func ParamsKey() []byte {
	return PrefixParams
}
