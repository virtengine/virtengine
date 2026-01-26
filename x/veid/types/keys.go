package types

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "veid"

	// StoreKey is the store key string for veid module
	StoreKey = ModuleName

	// RouterKey is the message route for veid module
	RouterKey = ModuleName

	// QuerierRoute is the querier route for veid module
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// PrefixIdentityRecord is the prefix for identity record storage
	// Key: PrefixIdentityRecord | address -> IdentityRecord
	PrefixIdentityRecord = []byte{0x01}

	// PrefixScope is the prefix for scope storage
	// Key: PrefixScope | address | scope_id -> IdentityScope
	PrefixScope = []byte{0x02}

	// PrefixScopeByType is the prefix for scope lookup by type
	// Key: PrefixScopeByType | address | scope_type -> []scope_id
	PrefixScopeByType = []byte{0x03}

	// PrefixVerificationHistory is the prefix for verification history
	// Key: PrefixVerificationHistory | address | timestamp -> VerificationEvent
	PrefixVerificationHistory = []byte{0x04}

	// PrefixParams is the prefix for module parameters
	PrefixParams = []byte{0x05}

	// PrefixApprovedClient is the prefix for approved client registry
	// Key: PrefixApprovedClient | client_id -> ApprovedClient
	PrefixApprovedClient = []byte{0x06}

	// PrefixSaltRegistry is the prefix for salt usage tracking
	// Key: PrefixSaltRegistry | salt_hash -> bool (used)
	PrefixSaltRegistry = []byte{0x07}

	// PrefixScore is the prefix for identity score storage
	// Key: PrefixScore | address -> IdentityScore
	PrefixScore = []byte{0x08}

	// PrefixScoreHistory is the prefix for score history storage
	// Key: PrefixScoreHistory | address | timestamp | block_height -> ScoreHistoryEntry
	PrefixScoreHistory = []byte{0x09}

	// PrefixIdentityWallet is the prefix for identity wallet storage
	// Key: PrefixIdentityWallet | address -> IdentityWallet
	PrefixIdentityWallet = []byte{0x0A}

	// PrefixWalletByID is the prefix for wallet lookup by wallet ID
	// Key: PrefixWalletByID | wallet_id -> address
	PrefixWalletByID = []byte{0x0B}

	// PrefixBorderlineFallback is the prefix for borderline fallback record storage
	// Key: PrefixBorderlineFallback | fallback_id -> BorderlineFallbackRecord
	PrefixBorderlineFallback = []byte{0x0C}

	// PrefixBorderlineFallbackByAccount is the prefix for fallback lookup by account
	// Key: PrefixBorderlineFallbackByAccount | address -> []fallback_id
	PrefixBorderlineFallbackByAccount = []byte{0x0D}

	// PrefixBorderlineParams is the prefix for borderline parameters
	// Key: PrefixBorderlineParams -> BorderlineParams
	PrefixBorderlineParams = []byte{0x0E}

	// PrefixPendingBorderlineFallback is the prefix for pending fallbacks queue
	// Key: PrefixPendingBorderlineFallback | expires_at | fallback_id -> bool
	PrefixPendingBorderlineFallback = []byte{0x0F}

	// PrefixEmbeddingEnvelope is the prefix for embedding envelope references
	// Key: PrefixEmbeddingEnvelope | envelope_id -> EmbeddingEnvelopeReference
	// SECURITY: Only stores hash references, NOT raw embeddings
	PrefixEmbeddingEnvelope = []byte{0x10}

	// PrefixEmbeddingEnvelopeByAccount is the prefix for envelope lookup by account
	// Key: PrefixEmbeddingEnvelopeByAccount | address | embedding_type -> []envelope_id
	PrefixEmbeddingEnvelopeByAccount = []byte{0x11}

	// PrefixDerivedFeatureRecord is the prefix for derived feature verification records
	// Key: PrefixDerivedFeatureRecord | record_id -> DerivedFeatureVerificationRecord
	PrefixDerivedFeatureRecord = []byte{0x12}

	// PrefixDerivedFeatureRecordByAccount is the prefix for record lookup by account
	// Key: PrefixDerivedFeatureRecordByAccount | address | block_height -> record_id
	PrefixDerivedFeatureRecordByAccount = []byte{0x15}

	// PrefixDataLifecycleRules is the prefix for data lifecycle rules
	// Key: PrefixDataLifecycleRules -> DataLifecycleRules
	PrefixDataLifecycleRules = []byte{0x16}

	// PrefixRetentionPolicy is the prefix for retention policies
	// Key: PrefixRetentionPolicy | policy_id -> RetentionPolicy
	PrefixRetentionPolicy = []byte{0x17}

	// PrefixExpiredArtifacts is the prefix for tracking expired artifacts to clean up
	// Key: PrefixExpiredArtifacts | expires_at | artifact_type | artifact_id -> bool
	PrefixExpiredArtifacts = []byte{0x18}

	// ============================================================================
	// Artifact Reference Keys (VE-218: Off-chain storage with on-chain references)
	// ============================================================================

	// PrefixArtifactReference is the prefix for identity artifact references
	// Key: PrefixArtifactReference | reference_id -> IdentityArtifactReference
	PrefixArtifactReference = []byte{0x19}

	// PrefixArtifactReferenceByAccount is the prefix for artifact lookup by account
	// Key: PrefixArtifactReferenceByAccount | address | artifact_type -> []reference_id
	PrefixArtifactReferenceByAccount = []byte{0x1A}

	// PrefixArtifactReferenceByContentHash is the prefix for artifact lookup by content hash
	// Key: PrefixArtifactReferenceByContentHash | content_hash -> reference_id
	PrefixArtifactReferenceByContentHash = []byte{0x1B}

	// PrefixChunkManifest is the prefix for chunk manifests
	// Key: PrefixChunkManifest | manifest_id -> ChunkManifestReference
	PrefixChunkManifest = []byte{0x1C}

	// PrefixPendingArtifactRetrieval is the prefix for pending artifact retrievals
	// Key: PrefixPendingArtifactRetrieval | request_id -> PendingArtifactRetrieval
	PrefixPendingArtifactRetrieval = []byte{0x1D}

	// ============================================================================
	// Pipeline Version Keys (VE-219: Deterministic verification runtime)
	// ============================================================================

	// PrefixPipelineVersion is the prefix for pipeline version storage
	// Key: PrefixPipelineVersion | version -> PipelineVersion
	PrefixPipelineVersion = []byte{0x1E}

	// PrefixActivePipelineVersion is the prefix for the active pipeline version
	// Key: PrefixActivePipelineVersion -> version string
	PrefixActivePipelineVersion = []byte{0x1F}

	// PrefixPipelineExecutionRecord is the prefix for pipeline execution records
	// Key: PrefixPipelineExecutionRecord | request_id -> PipelineExecutionRecord
	PrefixPipelineExecutionRecord = []byte{0x20}

	// PrefixPipelineExecutionByValidator is the prefix for execution records by validator
	// Key: PrefixPipelineExecutionByValidator | validator_address | request_id -> PipelineExecutionRecord
	PrefixPipelineExecutionByValidator = []byte{0x21}

	// PrefixModelManifest is the prefix for model manifest storage
	// Key: PrefixModelManifest | manifest_hash -> ModelManifest
	PrefixModelManifest = []byte{0x22}

	// PrefixPipelineConformanceResult is the prefix for conformance test results
	// Key: PrefixPipelineConformanceResult | test_id -> ConformanceTestResult
	PrefixPipelineConformanceResult = []byte{0x23}

	// ============================================================================
	// Scoring Model Keys (VE-220: Feature fusion scoring model)
	// ============================================================================

	// PrefixScoringModelVersion is the prefix for scoring model version storage
	// Key: PrefixScoringModelVersion | version -> ScoringModelVersion
	PrefixScoringModelVersion = []byte{0x24}

	// PrefixActiveScoringModel is the prefix for the active scoring model version
	// Key: PrefixActiveScoringModel -> version string
	PrefixActiveScoringModel = []byte{0x25}

	// PrefixScoringHistory is the prefix for scoring history entries
	// Key: PrefixScoringHistory | address | block_height -> ScoringHistoryEntry
	PrefixScoringHistory = []byte{0x26}

	// PrefixScoringVersionTransition is the prefix for version transition records
	// Key: PrefixScoringVersionTransition | address | block_height -> ScoreVersionTransition
	PrefixScoringVersionTransition = []byte{0x27}

	// PrefixEvidenceSummary is the prefix for evidence summary storage
	// Key: PrefixEvidenceSummary | address | block_height -> EvidenceSummary hash
	PrefixEvidenceSummary = []byte{0x28}

	// ============================================================================
	// SSO Verification Keys (VE-222)
	// ============================================================================

	// PrefixSSOLinkage is the prefix for SSO linkage storage
	// Key: PrefixSSOLinkage | linkage_id -> SSOLinkageMetadata
	PrefixSSOLinkage = []byte{0x29}

	// PrefixSSOLinkageByAccount is the prefix for SSO lookup by account
	// Key: PrefixSSOLinkageByAccount | address | provider -> linkage_id
	PrefixSSOLinkageByAccount = []byte{0x2A}

	// PrefixSSOChallenge is the prefix for SSO challenge storage
	// Key: PrefixSSOChallenge | challenge_id -> SSOVerificationChallenge
	PrefixSSOChallenge = []byte{0x2B}

	// ============================================================================
	// Domain Verification Keys (VE-223)
	// ============================================================================

	// PrefixDomainVerification is the prefix for domain verification storage
	// Key: PrefixDomainVerification | verification_id -> DomainVerificationRecord
	PrefixDomainVerification = []byte{0x2C}

	// PrefixDomainByAccount is the prefix for domain lookup by account
	// Key: PrefixDomainByAccount | address | domain_hash -> verification_id
	PrefixDomainByAccount = []byte{0x2D}

	// PrefixDomainByHash is the prefix for domain lookup by domain hash
	// Key: PrefixDomainByHash | domain_hash -> verification_id
	PrefixDomainByHash = []byte{0x2E}

	// PrefixDomainChallenge is the prefix for domain challenge storage
	// Key: PrefixDomainChallenge | challenge_id -> DomainVerificationChallenge
	PrefixDomainChallenge = []byte{0x2F}

	// ============================================================================
	// Email Verification Keys (VE-224)
	// ============================================================================

	// PrefixEmailVerification is the prefix for email verification storage
	// Key: PrefixEmailVerification | verification_id -> EmailVerificationRecord
	PrefixEmailVerification = []byte{0x30}

	// PrefixEmailByAccount is the prefix for email lookup by account
	// Key: PrefixEmailByAccount | address | email_hash -> verification_id
	PrefixEmailByAccount = []byte{0x31}

	// PrefixEmailChallenge is the prefix for email challenge storage
	// Key: PrefixEmailChallenge | challenge_id -> EmailVerificationChallenge
	PrefixEmailChallenge = []byte{0x32}

	// PrefixUsedNonce is the prefix for used nonce tracking (anti-replay)
	// Key: PrefixUsedNonce | nonce_hash -> UsedNonceRecord
	PrefixUsedNonce = []byte{0x33}

	// ============================================================================
	// Security Controls Keys (VE-225)
	// ============================================================================

	// PrefixTokenMapping is the prefix for token mapping storage
	// Key: PrefixTokenMapping | token -> TokenMapping
	PrefixTokenMapping = []byte{0x34}

	// PrefixTokenByInternal is the prefix for token lookup by internal reference
	// Key: PrefixTokenByInternal | internal_reference -> token
	PrefixTokenByInternal = []byte{0x35}

	// PrefixRetentionRule is the prefix for retention rule storage
	// Key: PrefixRetentionRule | rule_id -> RetentionRule
	PrefixRetentionRule = []byte{0x36}

	// PrefixRetentionEnforcement is the prefix for retention enforcement records
	// Key: PrefixRetentionEnforcement | enforcement_id -> RetentionEnforcementResult
	PrefixRetentionEnforcement = []byte{0x37}

	// PrefixSecurityAudit is the prefix for security audit events
	// Key: PrefixSecurityAudit | event_id -> SecurityAuditEvent
	PrefixSecurityAudit = []byte{0x38}

	// ============================================================================
	// Waldur Integration Keys (VE-226)
	// ============================================================================

	// PrefixWaldurLink is the prefix for Waldur link storage
	// Key: PrefixWaldurLink | link_id -> WaldurLinkRecord
	PrefixWaldurLink = []byte{0x39}

	// PrefixWaldurLinkByAccount is the prefix for Waldur link lookup by account
	// Key: PrefixWaldurLinkByAccount | address -> link_id
	PrefixWaldurLinkByAccount = []byte{0x3A}

	// PrefixWaldurLinkByUser is the prefix for Waldur link lookup by Waldur user
	// Key: PrefixWaldurLinkByUser | waldur_user_id -> link_id
	PrefixWaldurLinkByUser = []byte{0x3B}

	// PrefixWaldurRequest is the prefix for Waldur upload request storage
	// Key: PrefixWaldurRequest | request_id -> WaldurUploadRequest
	PrefixWaldurRequest = []byte{0x3C}

	// ============================================================================
	// Active Directory SSO Keys (VE-907)
	// ============================================================================

	// PrefixADSSOLinkage is the prefix for AD SSO linkage storage
	// Key: PrefixADSSOLinkage | linkage_id -> ADSSOLinkageMetadata
	PrefixADSSOLinkage = []byte{0x3D}

	// PrefixADSSOLinkageByAccount is the prefix for AD SSO lookup by account
	// Key: PrefixADSSOLinkageByAccount | address | auth_method -> linkage_id
	PrefixADSSOLinkageByAccount = []byte{0x3E}

	// PrefixADSSOLinkageByTenant is the prefix for AD SSO lookup by tenant
	// Key: PrefixADSSOLinkageByTenant | tenant_hash | subject_hash -> linkage_id
	PrefixADSSOLinkageByTenant = []byte{0x3F}

	// PrefixADSSOChallenge is the prefix for AD SSO challenge storage
	// Key: PrefixADSSOChallenge | challenge_id -> ADSSOChallenge
	PrefixADSSOChallenge = []byte{0x40}

	// PrefixADWalletBinding is the prefix for AD wallet binding storage
	// Key: PrefixADWalletBinding | binding_id -> ADWalletBinding
	PrefixADWalletBinding = []byte{0x41}

	// PrefixADWalletBindingByAddress is the prefix for wallet binding lookup by address
	// Key: PrefixADWalletBindingByAddress | wallet_address -> binding_id
	PrefixADWalletBindingByAddress = []byte{0x42}

	// ============================================================================
	// SMS Verification Keys (VE-910)
	// ============================================================================

	// PrefixSMSVerification is the prefix for SMS verification storage
	// Key: PrefixSMSVerification | verification_id -> SMSVerificationRecord
	// SECURITY: Only stores phone hashes, NEVER plaintext phone numbers
	PrefixSMSVerification = []byte{0x43}

	// PrefixSMSByAccount is the prefix for SMS lookup by account
	// Key: PrefixSMSByAccount | address | phone_hash -> verification_id
	PrefixSMSByAccount = []byte{0x44}

	// PrefixSMSByPhoneHash is the prefix for SMS lookup by phone hash
	// Key: PrefixSMSByPhoneHash | phone_hash -> verification_id
	PrefixSMSByPhoneHash = []byte{0x45}

	// PrefixSMSChallenge is the prefix for SMS OTP challenge storage
	// Key: PrefixSMSChallenge | challenge_id -> SMSOTPChallenge
	PrefixSMSChallenge = []byte{0x46}

	// PrefixSMSChallengeByAccount is the prefix for SMS challenge lookup by account
	// Key: PrefixSMSChallengeByAccount | address -> challenge_id
	PrefixSMSChallengeByAccount = []byte{0x47}

	// PrefixSMSRateLimit is the prefix for SMS rate limit state storage
	// Key: PrefixSMSRateLimit | entity_type | entity_hash -> SMSRateLimitState
	PrefixSMSRateLimit = []byte{0x48}

	// PrefixSMSGlobalRateLimit is the prefix for global SMS rate limit
	// Key: PrefixSMSGlobalRateLimit -> GlobalRateLimitState
	PrefixSMSGlobalRateLimit = []byte{0x49}

	// PrefixValidatorSMSGateway is the prefix for validator SMS gateway configuration
	// Key: PrefixValidatorSMSGateway | validator_address -> ValidatorSMSGateway
	PrefixValidatorSMSGateway = []byte{0x4A}

	// PrefixCarrierLookupCache is the prefix for carrier lookup cache
	// Key: PrefixCarrierLookupCache | phone_hash_ref -> CarrierLookupResult
	PrefixCarrierLookupCache = []byte{0x4B}

	// PrefixSMSDeliveryResult is the prefix for SMS delivery results
	// Key: PrefixSMSDeliveryResult | challenge_id -> SMSDeliveryResult
	PrefixSMSDeliveryResult = []byte{0x4C}
)

// IdentityRecordKey returns the store key for an identity record
func IdentityRecordKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixIdentityRecord)+len(address))
	key = append(key, PrefixIdentityRecord...)
	key = append(key, address...)
	return key
}

// ScopeKey returns the store key for a specific scope
func ScopeKey(address []byte, scopeID string) []byte {
	scopeIDBytes := []byte(scopeID)
	key := make([]byte, 0, len(PrefixScope)+len(address)+1+len(scopeIDBytes))
	key = append(key, PrefixScope...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, scopeIDBytes...)
	return key
}

// ScopePrefixKey returns the prefix for all scopes of an address
func ScopePrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixScope)+len(address)+1)
	key = append(key, PrefixScope...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// ScopeByTypeKey returns the store key for scope lookup by type
func ScopeByTypeKey(address []byte, scopeType ScopeType) []byte {
	scopeTypeBytes := []byte(scopeType)
	key := make([]byte, 0, len(PrefixScopeByType)+len(address)+1+len(scopeTypeBytes))
	key = append(key, PrefixScopeByType...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, scopeTypeBytes...)
	return key
}

// VerificationHistoryKey returns the store key for verification history entry
func VerificationHistoryKey(address []byte, timestamp int64) []byte {
	key := make([]byte, 0, len(PrefixVerificationHistory)+len(address)+9)
	key = append(key, PrefixVerificationHistory...)
	key = append(key, address...)
	key = append(key, byte('/'))
	// Encode timestamp as big-endian for proper ordering
	key = append(key, encodeInt64(timestamp)...)
	return key
}

// VerificationHistoryPrefixKey returns the prefix for all verification history of an address
func VerificationHistoryPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixVerificationHistory)+len(address)+1)
	key = append(key, PrefixVerificationHistory...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// ParamsKey returns the store key for module parameters
func ParamsKey() []byte {
	return PrefixParams
}

// ApprovedClientKey returns the store key for an approved client
func ApprovedClientKey(clientID string) []byte {
	clientIDBytes := []byte(clientID)
	key := make([]byte, 0, len(PrefixApprovedClient)+len(clientIDBytes))
	key = append(key, PrefixApprovedClient...)
	key = append(key, clientIDBytes...)
	return key
}

// SaltRegistryKey returns the store key for salt usage tracking
func SaltRegistryKey(saltHash []byte) []byte {
	key := make([]byte, 0, len(PrefixSaltRegistry)+len(saltHash))
	key = append(key, PrefixSaltRegistry...)
	key = append(key, saltHash...)
	return key
}

// encodeInt64 encodes an int64 as big-endian bytes
func encodeInt64(n int64) []byte {
	b := make([]byte, 8)
	b[0] = byte(n >> 56)
	b[1] = byte(n >> 48)
	b[2] = byte(n >> 40)
	b[3] = byte(n >> 32)
	b[4] = byte(n >> 24)
	b[5] = byte(n >> 16)
	b[6] = byte(n >> 8)
	b[7] = byte(n)
	return b
}

// ScoreKey returns the store key for an identity score
func ScoreKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixScore)+len(address))
	key = append(key, PrefixScore...)
	key = append(key, address...)
	return key
}

// ScoreHistoryKey returns the store key for a score history entry
// Uses timestamp and block height for uniqueness and ordering
func ScoreHistoryKey(address []byte, timestamp int64, blockHeight int64) []byte {
	key := make([]byte, 0, len(PrefixScoreHistory)+len(address)+1+16)
	key = append(key, PrefixScoreHistory...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, encodeInt64(timestamp)...)
	key = append(key, encodeInt64(blockHeight)...)
	return key
}

// ScoreHistoryPrefixKey returns the prefix for all score history of an address
func ScoreHistoryPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixScoreHistory)+len(address)+1)
	key = append(key, PrefixScoreHistory...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// IdentityWalletKey returns the store key for an identity wallet
func IdentityWalletKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixIdentityWallet)+len(address))
	key = append(key, PrefixIdentityWallet...)
	key = append(key, address...)
	return key
}

// WalletByIDKey returns the store key for wallet lookup by wallet ID
func WalletByIDKey(walletID string) []byte {
	walletIDBytes := []byte(walletID)
	key := make([]byte, 0, len(PrefixWalletByID)+len(walletIDBytes))
	key = append(key, PrefixWalletByID...)
	key = append(key, walletIDBytes...)
	return key
}

// BorderlineFallbackKey returns the store key for a borderline fallback record
func BorderlineFallbackKey(fallbackID string) []byte {
	fallbackIDBytes := []byte(fallbackID)
	key := make([]byte, 0, len(PrefixBorderlineFallback)+len(fallbackIDBytes))
	key = append(key, PrefixBorderlineFallback...)
	key = append(key, fallbackIDBytes...)
	return key
}

// BorderlineFallbackByAccountKey returns the store key for fallback lookup by account
func BorderlineFallbackByAccountKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixBorderlineFallbackByAccount)+len(address))
	key = append(key, PrefixBorderlineFallbackByAccount...)
	key = append(key, address...)
	return key
}

// BorderlineParamsKey returns the store key for borderline parameters
func BorderlineParamsKey() []byte {
	return PrefixBorderlineParams
}

// PendingBorderlineFallbackKey returns the store key for a pending fallback in queue
func PendingBorderlineFallbackKey(expiresAt int64, fallbackID string) []byte {
	fallbackIDBytes := []byte(fallbackID)
	key := make([]byte, 0, len(PrefixPendingBorderlineFallback)+8+1+len(fallbackIDBytes))
	key = append(key, PrefixPendingBorderlineFallback...)
	key = append(key, encodeInt64(expiresAt)...)
	key = append(key, byte('/'))
	key = append(key, fallbackIDBytes...)
	return key
}

// PendingBorderlineFallbackPrefixKey returns the prefix for pending fallbacks
func PendingBorderlineFallbackPrefixKey() []byte {
	return PrefixPendingBorderlineFallback
}

// ============================================================================
// Embedding Envelope Keys (VE-217: Derived Feature Minimization)
// ============================================================================

// EmbeddingEnvelopeKey returns the store key for an embedding envelope reference
func EmbeddingEnvelopeKey(envelopeID string) []byte {
	envelopeIDBytes := []byte(envelopeID)
	key := make([]byte, 0, len(PrefixEmbeddingEnvelope)+len(envelopeIDBytes))
	key = append(key, PrefixEmbeddingEnvelope...)
	key = append(key, envelopeIDBytes...)
	return key
}

// EmbeddingEnvelopeByAccountKey returns the key for envelope lookup by account and type
func EmbeddingEnvelopeByAccountKey(address []byte, embeddingType EmbeddingType) []byte {
	typeBytes := []byte(embeddingType)
	key := make([]byte, 0, len(PrefixEmbeddingEnvelopeByAccount)+len(address)+1+len(typeBytes))
	key = append(key, PrefixEmbeddingEnvelopeByAccount...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, typeBytes...)
	return key
}

// EmbeddingEnvelopeByAccountPrefixKey returns the prefix for all envelopes of an account
func EmbeddingEnvelopeByAccountPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixEmbeddingEnvelopeByAccount)+len(address)+1)
	key = append(key, PrefixEmbeddingEnvelopeByAccount...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// ============================================================================
// Derived Feature Verification Record Keys (VE-217)
// ============================================================================

// DerivedFeatureRecordKey returns the store key for a derived feature verification record
func DerivedFeatureRecordKey(recordID string) []byte {
	recordIDBytes := []byte(recordID)
	key := make([]byte, 0, len(PrefixDerivedFeatureRecord)+len(recordIDBytes))
	key = append(key, PrefixDerivedFeatureRecord...)
	key = append(key, recordIDBytes...)
	return key
}

// DerivedFeatureRecordByAccountKey returns the key for record lookup by account and block
func DerivedFeatureRecordByAccountKey(address []byte, blockHeight int64) []byte {
	key := make([]byte, 0, len(PrefixDerivedFeatureRecordByAccount)+len(address)+1+8)
	key = append(key, PrefixDerivedFeatureRecordByAccount...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, encodeInt64(blockHeight)...)
	return key
}

// DerivedFeatureRecordByAccountPrefixKey returns the prefix for all records of an account
func DerivedFeatureRecordByAccountPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixDerivedFeatureRecordByAccount)+len(address)+1)
	key = append(key, PrefixDerivedFeatureRecordByAccount...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// ============================================================================
// Data Lifecycle Keys (VE-217)
// ============================================================================

// DataLifecycleRulesKey returns the store key for data lifecycle rules
func DataLifecycleRulesKey() []byte {
	return PrefixDataLifecycleRules
}

// RetentionPolicyKey returns the store key for a retention policy
func RetentionPolicyKey(policyID string) []byte {
	policyIDBytes := []byte(policyID)
	key := make([]byte, 0, len(PrefixRetentionPolicy)+len(policyIDBytes))
	key = append(key, PrefixRetentionPolicy...)
	key = append(key, policyIDBytes...)
	return key
}

// ExpiredArtifactKey returns the key for an expired artifact to clean up
func ExpiredArtifactKey(expiresAt int64, artifactType string, artifactID string) []byte {
	typeBytes := []byte(artifactType)
	idBytes := []byte(artifactID)
	key := make([]byte, 0, len(PrefixExpiredArtifacts)+8+1+len(typeBytes)+1+len(idBytes))
	key = append(key, PrefixExpiredArtifacts...)
	key = append(key, encodeInt64(expiresAt)...)
	key = append(key, byte('/'))
	key = append(key, typeBytes...)
	key = append(key, byte('/'))
	key = append(key, idBytes...)
	return key
}

// ExpiredArtifactPrefixKey returns the prefix for expired artifacts
func ExpiredArtifactPrefixKey() []byte {
	return PrefixExpiredArtifacts
}

// ExpiredArtifactBeforeKey returns a prefix for artifacts expiring before a given time
func ExpiredArtifactBeforeKey(beforeTime int64) []byte {
	key := make([]byte, 0, len(PrefixExpiredArtifacts)+8)
	key = append(key, PrefixExpiredArtifacts...)
	key = append(key, encodeInt64(beforeTime)...)
	return key
}

// ============================================================================
// Artifact Reference Keys (VE-218: Off-chain storage with on-chain references)
// ============================================================================

// ArtifactReferenceKey returns the store key for an identity artifact reference
func ArtifactReferenceKey(referenceID string) []byte {
	referenceIDBytes := []byte(referenceID)
	key := make([]byte, 0, len(PrefixArtifactReference)+len(referenceIDBytes))
	key = append(key, PrefixArtifactReference...)
	key = append(key, referenceIDBytes...)
	return key
}

// ArtifactReferenceByAccountKey returns the key for artifact lookup by account and type
func ArtifactReferenceByAccountKey(address []byte, artifactType ArtifactType) []byte {
	typeBytes := []byte(artifactType)
	key := make([]byte, 0, len(PrefixArtifactReferenceByAccount)+len(address)+1+len(typeBytes))
	key = append(key, PrefixArtifactReferenceByAccount...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, typeBytes...)
	return key
}

// ArtifactReferenceByAccountPrefixKey returns the prefix for all artifacts of an account
func ArtifactReferenceByAccountPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixArtifactReferenceByAccount)+len(address)+1)
	key = append(key, PrefixArtifactReferenceByAccount...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// ArtifactReferenceByContentHashKey returns the key for artifact lookup by content hash
func ArtifactReferenceByContentHashKey(contentHash []byte) []byte {
	key := make([]byte, 0, len(PrefixArtifactReferenceByContentHash)+len(contentHash))
	key = append(key, PrefixArtifactReferenceByContentHash...)
	key = append(key, contentHash...)
	return key
}

// ChunkManifestKey returns the store key for a chunk manifest
func ChunkManifestKey(manifestID string) []byte {
	manifestIDBytes := []byte(manifestID)
	key := make([]byte, 0, len(PrefixChunkManifest)+len(manifestIDBytes))
	key = append(key, PrefixChunkManifest...)
	key = append(key, manifestIDBytes...)
	return key
}

// PendingArtifactRetrievalKey returns the store key for a pending artifact retrieval
func PendingArtifactRetrievalKey(requestID string) []byte {
	requestIDBytes := []byte(requestID)
	key := make([]byte, 0, len(PrefixPendingArtifactRetrieval)+len(requestIDBytes))
	key = append(key, PrefixPendingArtifactRetrieval...)
	key = append(key, requestIDBytes...)
	return key
}

// PendingArtifactRetrievalPrefixKey returns the prefix for all pending retrievals
func PendingArtifactRetrievalPrefixKey() []byte {
	return PrefixPendingArtifactRetrieval
}

// ============================================================================
// Pipeline Version Keys (VE-219: Deterministic verification runtime)
// ============================================================================

// PipelineVersionKey returns the store key for a pipeline version
func PipelineVersionKey(version string) []byte {
	versionBytes := []byte(version)
	key := make([]byte, 0, len(PrefixPipelineVersion)+len(versionBytes))
	key = append(key, PrefixPipelineVersion...)
	key = append(key, versionBytes...)
	return key
}

// PipelineVersionPrefixKey returns the prefix for all pipeline versions
func PipelineVersionPrefixKey() []byte {
	return PrefixPipelineVersion
}

// ActivePipelineVersionKey returns the store key for the active pipeline version
func ActivePipelineVersionKey() []byte {
	return PrefixActivePipelineVersion
}

// PipelineExecutionRecordKey returns the store key for a pipeline execution record
func PipelineExecutionRecordKey(requestID string) []byte {
	requestIDBytes := []byte(requestID)
	key := make([]byte, 0, len(PrefixPipelineExecutionRecord)+len(requestIDBytes))
	key = append(key, PrefixPipelineExecutionRecord...)
	key = append(key, requestIDBytes...)
	return key
}

// PipelineExecutionByValidatorKey returns the store key for execution lookup by validator
func PipelineExecutionByValidatorKey(validatorAddress []byte, requestID string) []byte {
	requestIDBytes := []byte(requestID)
	key := make([]byte, 0, len(PrefixPipelineExecutionByValidator)+len(validatorAddress)+1+len(requestIDBytes))
	key = append(key, PrefixPipelineExecutionByValidator...)
	key = append(key, validatorAddress...)
	key = append(key, byte('/'))
	key = append(key, requestIDBytes...)
	return key
}

// PipelineExecutionByValidatorPrefixKey returns the prefix for all executions by a validator
func PipelineExecutionByValidatorPrefixKey(validatorAddress []byte) []byte {
	key := make([]byte, 0, len(PrefixPipelineExecutionByValidator)+len(validatorAddress)+1)
	key = append(key, PrefixPipelineExecutionByValidator...)
	key = append(key, validatorAddress...)
	key = append(key, byte('/'))
	return key
}

// ModelManifestKey returns the store key for a model manifest
func ModelManifestKey(manifestHash string) []byte {
	hashBytes := []byte(manifestHash)
	key := make([]byte, 0, len(PrefixModelManifest)+len(hashBytes))
	key = append(key, PrefixModelManifest...)
	key = append(key, hashBytes...)
	return key
}

// PipelineConformanceResultKey returns the store key for a conformance test result
func PipelineConformanceResultKey(testID string) []byte {
	testIDBytes := []byte(testID)
	key := make([]byte, 0, len(PrefixPipelineConformanceResult)+len(testIDBytes))
	key = append(key, PrefixPipelineConformanceResult...)
	key = append(key, testIDBytes...)
	return key
}

// PipelineConformanceResultPrefixKey returns the prefix for all conformance test results
func PipelineConformanceResultPrefixKey() []byte {
	return PrefixPipelineConformanceResult
}

// ============================================================================
// Scoring Model Key Functions (VE-220)
// ============================================================================

// ScoringModelVersionKey returns the store key for a scoring model version
func ScoringModelVersionKey(version string) []byte {
	versionBytes := []byte(version)
	key := make([]byte, 0, len(PrefixScoringModelVersion)+len(versionBytes))
	key = append(key, PrefixScoringModelVersion...)
	key = append(key, versionBytes...)
	return key
}

// ScoringModelVersionPrefixKey returns the prefix for all scoring model versions
func ScoringModelVersionPrefixKey() []byte {
	return PrefixScoringModelVersion
}

// ActiveScoringModelKey returns the store key for the active scoring model
func ActiveScoringModelKey() []byte {
	return PrefixActiveScoringModel
}

// ScoringHistoryKey returns the store key for a scoring history entry
func ScoringHistoryKey(address []byte, blockHeight int64) []byte {
	heightBytes := make([]byte, 8)
	// Use big-endian for proper ordering
	heightBytes[0] = byte(blockHeight >> 56)
	heightBytes[1] = byte(blockHeight >> 48)
	heightBytes[2] = byte(blockHeight >> 40)
	heightBytes[3] = byte(blockHeight >> 32)
	heightBytes[4] = byte(blockHeight >> 24)
	heightBytes[5] = byte(blockHeight >> 16)
	heightBytes[6] = byte(blockHeight >> 8)
	heightBytes[7] = byte(blockHeight)

	key := make([]byte, 0, len(PrefixScoringHistory)+len(address)+1+8)
	key = append(key, PrefixScoringHistory...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, heightBytes...)
	return key
}

// ScoringHistoryPrefixKey returns the prefix for all scoring history for an address
func ScoringHistoryPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixScoringHistory)+len(address)+1)
	key = append(key, PrefixScoringHistory...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// ScoringVersionTransitionKey returns the store key for a version transition record
func ScoringVersionTransitionKey(address []byte, blockHeight int64) []byte {
	heightBytes := make([]byte, 8)
	heightBytes[0] = byte(blockHeight >> 56)
	heightBytes[1] = byte(blockHeight >> 48)
	heightBytes[2] = byte(blockHeight >> 40)
	heightBytes[3] = byte(blockHeight >> 32)
	heightBytes[4] = byte(blockHeight >> 24)
	heightBytes[5] = byte(blockHeight >> 16)
	heightBytes[6] = byte(blockHeight >> 8)
	heightBytes[7] = byte(blockHeight)

	key := make([]byte, 0, len(PrefixScoringVersionTransition)+len(address)+1+8)
	key = append(key, PrefixScoringVersionTransition...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, heightBytes...)
	return key
}

// ScoringVersionTransitionPrefixKey returns the prefix for all version transitions for an address
func ScoringVersionTransitionPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixScoringVersionTransition)+len(address)+1)
	key = append(key, PrefixScoringVersionTransition...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// ============================================================================
// Active Directory SSO Key Functions (VE-907)
// ============================================================================

// ADSSOLinkageKey returns the store key for an AD SSO linkage
func ADSSOLinkageKey(linkageID string) []byte {
	linkageIDBytes := []byte(linkageID)
	key := make([]byte, 0, len(PrefixADSSOLinkage)+len(linkageIDBytes))
	key = append(key, PrefixADSSOLinkage...)
	key = append(key, linkageIDBytes...)
	return key
}

// ADSSOLinkagePrefixKey returns the prefix for all AD SSO linkages
func ADSSOLinkagePrefixKey() []byte {
	return PrefixADSSOLinkage
}

// ADSSOLinkageByAccountKey returns the store key for AD SSO lookup by account and method
func ADSSOLinkageByAccountKey(address []byte, authMethod ADAuthMethod) []byte {
	methodBytes := []byte(authMethod)
	key := make([]byte, 0, len(PrefixADSSOLinkageByAccount)+len(address)+1+len(methodBytes))
	key = append(key, PrefixADSSOLinkageByAccount...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, methodBytes...)
	return key
}

// ADSSOLinkageByAccountPrefixKey returns the prefix for all AD SSO linkages for an address
func ADSSOLinkageByAccountPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixADSSOLinkageByAccount)+len(address)+1)
	key = append(key, PrefixADSSOLinkageByAccount...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// ADSSOLinkageByTenantKey returns the store key for AD SSO lookup by tenant and subject
func ADSSOLinkageByTenantKey(tenantHash, subjectHash string) []byte {
	tenantHashBytes := []byte(tenantHash)
	subjectHashBytes := []byte(subjectHash)
	key := make([]byte, 0, len(PrefixADSSOLinkageByTenant)+len(tenantHashBytes)+1+len(subjectHashBytes))
	key = append(key, PrefixADSSOLinkageByTenant...)
	key = append(key, tenantHashBytes...)
	key = append(key, byte('/'))
	key = append(key, subjectHashBytes...)
	return key
}

// ADSSOLinkageByTenantPrefixKey returns the prefix for all AD SSO linkages for a tenant
func ADSSOLinkageByTenantPrefixKey(tenantHash string) []byte {
	tenantHashBytes := []byte(tenantHash)
	key := make([]byte, 0, len(PrefixADSSOLinkageByTenant)+len(tenantHashBytes)+1)
	key = append(key, PrefixADSSOLinkageByTenant...)
	key = append(key, tenantHashBytes...)
	key = append(key, byte('/'))
	return key
}

// ADSSOChallengeKey returns the store key for an AD SSO challenge
func ADSSOChallengeKey(challengeID string) []byte {
	challengeIDBytes := []byte(challengeID)
	key := make([]byte, 0, len(PrefixADSSOChallenge)+len(challengeIDBytes))
	key = append(key, PrefixADSSOChallenge...)
	key = append(key, challengeIDBytes...)
	return key
}

// ADSSOChallengePrefixKey returns the prefix for all AD SSO challenges
func ADSSOChallengePrefixKey() []byte {
	return PrefixADSSOChallenge
}

// ADWalletBindingKey returns the store key for an AD wallet binding
func ADWalletBindingKey(bindingID string) []byte {
	bindingIDBytes := []byte(bindingID)
	key := make([]byte, 0, len(PrefixADWalletBinding)+len(bindingIDBytes))
	key = append(key, PrefixADWalletBinding...)
	key = append(key, bindingIDBytes...)
	return key
}

// ADWalletBindingPrefixKey returns the prefix for all AD wallet bindings
func ADWalletBindingPrefixKey() []byte {
	return PrefixADWalletBinding
}

// ADWalletBindingByAddressKey returns the store key for wallet binding lookup by address
func ADWalletBindingByAddressKey(walletAddress []byte) []byte {
	key := make([]byte, 0, len(PrefixADWalletBindingByAddress)+len(walletAddress))
	key = append(key, PrefixADWalletBindingByAddress...)
	key = append(key, walletAddress...)
	return key
}

// EvidenceSummaryKey returns the store key for an evidence summary
func EvidenceSummaryKey(address []byte, blockHeight int64) []byte {
	heightBytes := make([]byte, 8)
	heightBytes[0] = byte(blockHeight >> 56)
	heightBytes[1] = byte(blockHeight >> 48)
	heightBytes[2] = byte(blockHeight >> 40)
	heightBytes[3] = byte(blockHeight >> 32)
	heightBytes[4] = byte(blockHeight >> 24)
	heightBytes[5] = byte(blockHeight >> 16)
	heightBytes[6] = byte(blockHeight >> 8)
	heightBytes[7] = byte(blockHeight)

	key := make([]byte, 0, len(PrefixEvidenceSummary)+len(address)+1+8)
	key = append(key, PrefixEvidenceSummary...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, heightBytes...)
	return key
}

// EvidenceSummaryPrefixKey returns the prefix for all evidence summaries for an address
func EvidenceSummaryPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixEvidenceSummary)+len(address)+1)
	key = append(key, PrefixEvidenceSummary...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// ============================================================================
// SMS Verification Key Functions (VE-910)
// ============================================================================

// SMSVerificationKey returns the store key for an SMS verification record
func SMSVerificationKey(verificationID string) []byte {
	verificationIDBytes := []byte(verificationID)
	key := make([]byte, 0, len(PrefixSMSVerification)+len(verificationIDBytes))
	key = append(key, PrefixSMSVerification...)
	key = append(key, verificationIDBytes...)
	return key
}

// SMSVerificationPrefixKey returns the prefix for all SMS verification records
func SMSVerificationPrefixKey() []byte {
	return PrefixSMSVerification
}

// SMSByAccountKey returns the store key for SMS lookup by account and phone hash
func SMSByAccountKey(address []byte, phoneHash string) []byte {
	phoneHashBytes := []byte(phoneHash)
	key := make([]byte, 0, len(PrefixSMSByAccount)+len(address)+1+len(phoneHashBytes))
	key = append(key, PrefixSMSByAccount...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, phoneHashBytes...)
	return key
}

// SMSByAccountPrefixKey returns the prefix for all SMS verifications for an account
func SMSByAccountPrefixKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixSMSByAccount)+len(address)+1)
	key = append(key, PrefixSMSByAccount...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// SMSByPhoneHashKey returns the store key for SMS lookup by phone hash
func SMSByPhoneHashKey(phoneHash string) []byte {
	phoneHashBytes := []byte(phoneHash)
	key := make([]byte, 0, len(PrefixSMSByPhoneHash)+len(phoneHashBytes))
	key = append(key, PrefixSMSByPhoneHash...)
	key = append(key, phoneHashBytes...)
	return key
}

// SMSChallengeKey returns the store key for an SMS OTP challenge
func SMSChallengeKey(challengeID string) []byte {
	challengeIDBytes := []byte(challengeID)
	key := make([]byte, 0, len(PrefixSMSChallenge)+len(challengeIDBytes))
	key = append(key, PrefixSMSChallenge...)
	key = append(key, challengeIDBytes...)
	return key
}

// SMSChallengePrefixKey returns the prefix for all SMS challenges
func SMSChallengePrefixKey() []byte {
	return PrefixSMSChallenge
}

// SMSChallengeByAccountKey returns the store key for SMS challenge lookup by account
func SMSChallengeByAccountKey(address []byte) []byte {
	key := make([]byte, 0, len(PrefixSMSChallengeByAccount)+len(address))
	key = append(key, PrefixSMSChallengeByAccount...)
	key = append(key, address...)
	return key
}

// SMSRateLimitKey returns the store key for SMS rate limit state
func SMSRateLimitKey(entityType RateLimitType, entityHash string) []byte {
	entityTypeBytes := []byte(entityType)
	entityHashBytes := []byte(entityHash)
	key := make([]byte, 0, len(PrefixSMSRateLimit)+len(entityTypeBytes)+1+len(entityHashBytes))
	key = append(key, PrefixSMSRateLimit...)
	key = append(key, entityTypeBytes...)
	key = append(key, byte('/'))
	key = append(key, entityHashBytes...)
	return key
}

// SMSRateLimitPrefixKey returns the prefix for all rate limit states of a type
func SMSRateLimitPrefixKey(entityType RateLimitType) []byte {
	entityTypeBytes := []byte(entityType)
	key := make([]byte, 0, len(PrefixSMSRateLimit)+len(entityTypeBytes)+1)
	key = append(key, PrefixSMSRateLimit...)
	key = append(key, entityTypeBytes...)
	key = append(key, byte('/'))
	return key
}

// SMSGlobalRateLimitKey returns the store key for global SMS rate limit
func SMSGlobalRateLimitKey() []byte {
	return PrefixSMSGlobalRateLimit
}

// ValidatorSMSGatewayKey returns the store key for a validator SMS gateway
func ValidatorSMSGatewayKey(validatorAddress []byte) []byte {
	key := make([]byte, 0, len(PrefixValidatorSMSGateway)+len(validatorAddress))
	key = append(key, PrefixValidatorSMSGateway...)
	key = append(key, validatorAddress...)
	return key
}

// ValidatorSMSGatewayPrefixKey returns the prefix for all validator SMS gateways
func ValidatorSMSGatewayPrefixKey() []byte {
	return PrefixValidatorSMSGateway
}

// CarrierLookupCacheKey returns the store key for carrier lookup cache
func CarrierLookupCacheKey(phoneHashRef string) []byte {
	phoneHashRefBytes := []byte(phoneHashRef)
	key := make([]byte, 0, len(PrefixCarrierLookupCache)+len(phoneHashRefBytes))
	key = append(key, PrefixCarrierLookupCache...)
	key = append(key, phoneHashRefBytes...)
	return key
}

// SMSDeliveryResultKey returns the store key for SMS delivery result
func SMSDeliveryResultKey(challengeID string) []byte {
	challengeIDBytes := []byte(challengeID)
	key := make([]byte, 0, len(PrefixSMSDeliveryResult)+len(challengeIDBytes))
	key = append(key, PrefixSMSDeliveryResult...)
	key = append(key, challengeIDBytes...)
	return key
}