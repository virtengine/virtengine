package types

import (
	errorsmod "cosmossdk.io/errors"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

// Error codes for the veid module (range: 1000-1199)
// Errors 1000-1029 are defined in sdk/go/node/veid/v1/errors.go and aliased here
// to avoid duplicate registration. Errors 1030+ are defined only here.
var (
	// ============================================================================
	// Core Errors (aliased from SDK - codes 1000-1029)
	// ============================================================================

	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = veidv1.ErrInvalidAddress

	// ErrInvalidScope is returned when a scope is malformed
	ErrInvalidScope = veidv1.ErrInvalidScope

	// ErrInvalidScopeType is returned when a scope type is invalid
	ErrInvalidScopeType = veidv1.ErrInvalidScopeType

	// ErrInvalidScopeVersion is returned when a scope version is invalid
	ErrInvalidScopeVersion = veidv1.ErrInvalidScopeVersion

	// ErrInvalidPayload is returned when an encrypted payload is invalid
	ErrInvalidPayload = veidv1.ErrInvalidPayload

	// ErrInvalidSalt is returned when a salt is invalid
	ErrInvalidSalt = veidv1.ErrInvalidSalt

	// ErrSaltAlreadyUsed is returned when a salt has already been used
	ErrSaltAlreadyUsed = veidv1.ErrSaltAlreadyUsed

	// ErrInvalidDeviceInfo is returned when device information is invalid
	ErrInvalidDeviceInfo = veidv1.ErrInvalidDeviceInfo

	// ErrInvalidClientID is returned when a client ID is invalid
	ErrInvalidClientID = veidv1.ErrInvalidClientID

	// ErrClientNotApproved is returned when a client is not approved
	ErrClientNotApproved = veidv1.ErrClientNotApproved

	// ErrInvalidClientSignature is returned when a client signature is invalid
	ErrInvalidClientSignature = veidv1.ErrInvalidClientSignature

	// ErrInvalidUserSignature is returned when a user signature is invalid
	ErrInvalidUserSignature = veidv1.ErrInvalidUserSignature

	// ErrInvalidPayloadHash is returned when a payload hash is invalid
	ErrInvalidPayloadHash = veidv1.ErrInvalidPayloadHash

	// ErrInvalidVerificationStatus is returned when a verification status is invalid
	ErrInvalidVerificationStatus = veidv1.ErrInvalidVerificationStatus

	// ErrInvalidVerificationEvent is returned when a verification event is invalid
	ErrInvalidVerificationEvent = veidv1.ErrInvalidVerificationEvent

	// ErrInvalidScore is returned when an identity score is invalid
	ErrInvalidScore = veidv1.ErrInvalidScore

	// ErrInvalidTier is returned when an identity tier is invalid
	ErrInvalidTier = veidv1.ErrInvalidTier

	// ErrInvalidIdentityRecord is returned when an identity record is invalid
	ErrInvalidIdentityRecord = veidv1.ErrInvalidIdentityRecord

	// ErrInvalidWallet is returned when an identity wallet is invalid
	ErrInvalidWallet = veidv1.ErrInvalidWallet

	// ErrScopeNotFound is returned when a scope is not found
	ErrScopeNotFound = veidv1.ErrScopeNotFound

	// ErrIdentityRecordNotFound is returned when an identity record is not found
	ErrIdentityRecordNotFound = veidv1.ErrIdentityRecordNotFound

	// ErrScopeAlreadyExists is returned when a scope already exists
	ErrScopeAlreadyExists = veidv1.ErrScopeAlreadyExists

	// ErrScopeRevoked is returned when trying to use a revoked scope
	ErrScopeRevoked = veidv1.ErrScopeRevoked

	// ErrScopeExpired is returned when trying to use an expired scope
	ErrScopeExpired = veidv1.ErrScopeExpired

	// ErrUnauthorized is returned when the sender is not authorized
	ErrUnauthorized = veidv1.ErrUnauthorized

	// ErrInvalidStatusTransition is returned when an invalid status transition is attempted
	ErrInvalidStatusTransition = veidv1.ErrInvalidStatusTransition

	// ErrIdentityLocked is returned when trying to modify a locked identity
	ErrIdentityLocked = veidv1.ErrIdentityLocked

	// ErrMaxScopesExceeded is returned when the maximum number of scopes is exceeded
	ErrMaxScopesExceeded = veidv1.ErrMaxScopesExceeded

	// ErrVerificationInProgress is returned when verification is already in progress
	ErrVerificationInProgress = veidv1.ErrVerificationInProgress

	// ErrValidatorOnly is returned when a non-validator attempts a validator-only action
	ErrValidatorOnly = veidv1.ErrValidatorOnly

	// ============================================================================
	// Extended Errors (defined here - codes 1030+)
	// ============================================================================

	// ErrSignatureMismatch is returned when signatures don't match expected values
	ErrSignatureMismatch = errorsmod.Register(ModuleName, 1030, "signature mismatch")

	// ErrInvalidParams is returned when module parameters are invalid
	ErrInvalidParams = errorsmod.Register(ModuleName, 1031, "invalid parameters")

	// ErrInvalidVerificationRequest is returned when a verification request is invalid
	ErrInvalidVerificationRequest = errorsmod.Register(ModuleName, 1032, "invalid verification request")

	// ErrInvalidVerificationResult is returned when a verification result is invalid
	ErrInvalidVerificationResult = errorsmod.Register(ModuleName, 1033, "invalid verification result")

	// ErrVerificationRequestNotFound is returned when a verification request is not found
	ErrVerificationRequestNotFound = errorsmod.Register(ModuleName, 1034, "verification request not found")

	// ErrDecryptionFailed is returned when scope decryption fails
	ErrDecryptionFailed = errorsmod.Register(ModuleName, 1035, "scope decryption failed")

	// ErrMLInferenceFailed is returned when ML scoring fails
	ErrMLInferenceFailed = errorsmod.Register(ModuleName, 1036, "ML inference failed")

	// ErrVerificationTimeout is returned when verification times out
	ErrVerificationTimeout = errorsmod.Register(ModuleName, 1037, "verification timeout")

	// ErrMaxRetriesExceeded is returned when max retries are exceeded
	ErrMaxRetriesExceeded = errorsmod.Register(ModuleName, 1038, "max retries exceeded")

	// ErrNotBlockProposer is returned when non-proposer attempts proposer-only action
	ErrNotBlockProposer = errorsmod.Register(ModuleName, 1039, "not block proposer")

	// ErrValidatorKeyNotFound is returned when validator identity key is not found
	ErrValidatorKeyNotFound = errorsmod.Register(ModuleName, 1040, "validator identity key not found")

	// ErrWalletNotFound is returned when an identity wallet is not found
	ErrWalletNotFound = errorsmod.Register(ModuleName, 1041, "identity wallet not found")

	// ErrWalletAlreadyExists is returned when a wallet already exists for an account
	ErrWalletAlreadyExists = errorsmod.Register(ModuleName, 1042, "identity wallet already exists")

	// ErrWalletInactive is returned when trying to use an inactive wallet
	ErrWalletInactive = errorsmod.Register(ModuleName, 1043, "identity wallet is inactive")

	// ErrConsentNotGranted is returned when required consent is not granted
	ErrConsentNotGranted = errorsmod.Register(ModuleName, 1044, "consent not granted")

	// ErrConsentExpired is returned when consent has expired
	ErrConsentExpired = errorsmod.Register(ModuleName, 1045, "consent has expired")

	// ErrInvalidBindingSignature is returned when wallet binding signature is invalid
	ErrInvalidBindingSignature = errorsmod.Register(ModuleName, 1046, "invalid wallet binding signature")

	// ErrScopeNotInWallet is returned when a scope is not found in the wallet
	ErrScopeNotInWallet = errorsmod.Register(ModuleName, 1047, "scope not found in wallet")

	// ErrScopeAlreadyInWallet is returned when a scope is already in the wallet
	ErrScopeAlreadyInWallet = errorsmod.Register(ModuleName, 1048, "scope already in wallet")

	// ErrDerivedFeaturesEmpty is returned when derived features are empty
	ErrDerivedFeaturesEmpty = errorsmod.Register(ModuleName, 1049, "derived features are empty")

	// ErrVerificationMismatch is returned when consensus verification result mismatches
	ErrVerificationMismatch = errorsmod.Register(ModuleName, 1050, "verification result mismatch")

	// ErrModelVersionMismatch is returned when ML model versions don't match
	ErrModelVersionMismatch = errorsmod.Register(ModuleName, 1051, "model version mismatch")

	// ErrInputHashMismatch is returned when input hashes don't match
	ErrInputHashMismatch = errorsmod.Register(ModuleName, 1052, "input hash mismatch")

	// ErrScoreToleranceExceeded is returned when score difference exceeds tolerance
	ErrScoreToleranceExceeded = errorsmod.Register(ModuleName, 1053, "score tolerance exceeded")

	// ErrVoteExtensionInvalid is returned when a vote extension is invalid
	ErrVoteExtensionInvalid = errorsmod.Register(ModuleName, 1054, "invalid vote extension")

	// ErrInvalidBorderlineFallback is returned when a borderline fallback is invalid
	ErrInvalidBorderlineFallback = errorsmod.Register(ModuleName, 1055, "invalid borderline fallback")

	// ErrBorderlineFallbackNotFound is returned when a borderline fallback is not found
	ErrBorderlineFallbackNotFound = errorsmod.Register(ModuleName, 1056, "borderline fallback not found")

	// ErrBorderlineFallbackExpired is returned when a borderline fallback has expired
	ErrBorderlineFallbackExpired = errorsmod.Register(ModuleName, 1057, "borderline fallback expired")

	// ErrBorderlineFallbackAlreadyCompleted is returned when fallback is already completed
	ErrBorderlineFallbackAlreadyCompleted = errorsmod.Register(ModuleName, 1058, "borderline fallback already completed")

	// ErrMFAChallengeNotSatisfied is returned when MFA challenge is not satisfied
	ErrMFAChallengeNotSatisfied = errorsmod.Register(ModuleName, 1059, "MFA challenge not satisfied")

	// ErrNoEnrolledFactors is returned when account has no enrolled MFA factors
	ErrNoEnrolledFactors = errorsmod.Register(ModuleName, 1060, "no enrolled MFA factors")

	// ErrBorderlineDisabled is returned when borderline fallback is disabled
	ErrBorderlineDisabled = errorsmod.Register(ModuleName, 1061, "borderline fallback is disabled")

	// ============================================================================
	// Pipeline Version Errors (VE-219)
	// ============================================================================

	// ErrInvalidPipelineVersion is returned when a pipeline version is invalid
	ErrInvalidPipelineVersion = errorsmod.Register(ModuleName, 1062, "invalid pipeline version")

	// ErrPipelineVersionNotFound is returned when a pipeline version is not found
	ErrPipelineVersionNotFound = errorsmod.Register(ModuleName, 1063, "pipeline version not found")

	// ErrPipelineVersionAlreadyExists is returned when a pipeline version already exists
	ErrPipelineVersionAlreadyExists = errorsmod.Register(ModuleName, 1064, "pipeline version already exists")

	// ErrNoPipelineVersionActive is returned when no active pipeline version exists
	ErrNoPipelineVersionActive = errorsmod.Register(ModuleName, 1065, "no active pipeline version")

	// ErrPipelineVersionMismatch is returned when pipeline versions don't match for consensus
	ErrPipelineVersionMismatch = errorsmod.Register(ModuleName, 1066, "pipeline version mismatch")

	// ErrModelManifestMismatch is returned when model manifests don't match
	ErrModelManifestMismatch = errorsmod.Register(ModuleName, 1067, "model manifest mismatch")

	// ErrDeterminismViolation is returned when deterministic execution check fails
	ErrDeterminismViolation = errorsmod.Register(ModuleName, 1068, "determinism violation detected")

	// ErrInvalidModelManifest is returned when a model manifest is invalid
	ErrInvalidModelManifest = errorsmod.Register(ModuleName, 1069, "invalid model manifest")

	// ErrPipelineExecutionFailed is returned when pipeline execution fails
	ErrPipelineExecutionFailed = errorsmod.Register(ModuleName, 1070, "pipeline execution failed")

	// ============================================================================
	// Scoring Model Errors (VE-220)
	// ============================================================================

	// ErrInvalidScoringModel is returned when a scoring model is invalid
	ErrInvalidScoringModel = errorsmod.Register(ModuleName, 1071, "invalid scoring model")

	// ErrScoringModelNotFound is returned when a scoring model version is not found
	ErrScoringModelNotFound = errorsmod.Register(ModuleName, 1072, "scoring model not found")

	// ErrScoringModelAlreadyExists is returned when a scoring model version already exists
	ErrScoringModelAlreadyExists = errorsmod.Register(ModuleName, 1073, "scoring model already exists")

	// ErrInvalidScoringInput is returned when scoring input is invalid
	ErrInvalidScoringInput = errorsmod.Register(ModuleName, 1074, "invalid scoring input")

	// ErrScoringFailed is returned when score computation fails
	ErrScoringFailed = errorsmod.Register(ModuleName, 1075, "scoring computation failed")

	// ErrNoActiveScoringModel is returned when no active scoring model exists
	ErrNoActiveScoringModel = errorsmod.Register(ModuleName, 1076, "no active scoring model")

	// ErrScoringModelVersionMismatch is returned when scoring model versions don't match
	ErrScoringModelVersionMismatch = errorsmod.Register(ModuleName, 1077, "scoring model version mismatch")

	// ============================================================================
	// Verification Scope Errors (VE-222, VE-223, VE-224)
	// ============================================================================

	// ErrInvalidSSO is returned for SSO verification errors
	ErrInvalidSSO = errorsmod.Register(ModuleName, 1078, "invalid SSO verification")

	// ErrSSONotFound is returned when SSO linkage is not found
	ErrSSONotFound = errorsmod.Register(ModuleName, 1079, "SSO linkage not found")

	// ErrSSOExpired is returned when SSO linkage has expired
	ErrSSOExpired = errorsmod.Register(ModuleName, 1080, "SSO linkage has expired")

	// ErrSSORevoked is returned when SSO linkage has been revoked
	ErrSSORevoked = errorsmod.Register(ModuleName, 1081, "SSO linkage has been revoked")

	// ErrInvalidDomain is returned for domain verification errors
	ErrInvalidDomain = errorsmod.Register(ModuleName, 1082, "invalid domain verification")

	// ErrDomainNotFound is returned when domain verification is not found
	ErrDomainNotFound = errorsmod.Register(ModuleName, 1083, "domain verification not found")

	// ErrDomainExpired is returned when domain verification has expired
	ErrDomainExpired = errorsmod.Register(ModuleName, 1084, "domain verification has expired")

	// ErrDomainRevoked is returned when domain verification has been revoked
	ErrDomainRevoked = errorsmod.Register(ModuleName, 1085, "domain verification has been revoked")

	// ErrInvalidEmail is returned for email verification errors
	ErrInvalidEmail = errorsmod.Register(ModuleName, 1086, "invalid email verification")

	// ErrEmailNotFound is returned when email verification is not found
	ErrEmailNotFound = errorsmod.Register(ModuleName, 1087, "email verification not found")

	// ErrEmailExpired is returned when email verification has expired
	ErrEmailExpired = errorsmod.Register(ModuleName, 1088, "email verification has expired")

	// ErrNonceAlreadyUsed is returned when a nonce has already been used (anti-replay)
	ErrNonceAlreadyUsed = errorsmod.Register(ModuleName, 1089, "nonce already used")

	// ErrMaxAttemptsExceeded is returned when max verification attempts exceeded
	ErrMaxAttemptsExceeded = errorsmod.Register(ModuleName, 1090, "maximum verification attempts exceeded")

	// ============================================================================
	// Security Controls Errors (VE-225)
	// ============================================================================

	// ErrInvalidToken is returned for tokenization errors
	ErrInvalidToken = errorsmod.Register(ModuleName, 1091, "invalid token")

	// ErrTokenNotFound is returned when token mapping is not found
	ErrTokenNotFound = errorsmod.Register(ModuleName, 1092, "token not found")

	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errorsmod.Register(ModuleName, 1093, "token has expired")

	// ErrTokenRevoked is returned when token has been revoked
	ErrTokenRevoked = errorsmod.Register(ModuleName, 1094, "token has been revoked")

	// ErrInvalidRetention is returned for retention policy errors
	ErrInvalidRetention = errorsmod.Register(ModuleName, 1095, "invalid retention policy")

	// ErrRetentionNotFound is returned when retention policy is not found
	ErrRetentionNotFound = errorsmod.Register(ModuleName, 1096, "retention policy not found")

	// ============================================================================
	// Waldur Integration Errors (VE-226)
	// ============================================================================

	// ErrInvalidWaldur is returned for Waldur integration errors
	ErrInvalidWaldur = errorsmod.Register(ModuleName, 1097, "invalid Waldur integration")

	// ErrWaldurLinkNotFound is returned when Waldur link is not found
	ErrWaldurLinkNotFound = errorsmod.Register(ModuleName, 1098, "Waldur link not found")

	// ErrWaldurLinkExpired is returned when Waldur request has expired
	ErrWaldurLinkExpired = errorsmod.Register(ModuleName, 1099, "Waldur request has expired")

	// ErrWaldurSyncFailed is returned when Waldur sync fails
	ErrWaldurSyncFailed = errorsmod.Register(ModuleName, 1100, "Waldur sync failed")

	// ============================================================================
	// Active Directory SSO Errors (VE-907)
	// ============================================================================

	// ErrInvalidADSSO is returned for AD SSO verification errors
	ErrInvalidADSSO = errorsmod.Register(ModuleName, 1101, "invalid AD SSO verification")

	// ErrADSSONotFound is returned when AD SSO linkage is not found
	ErrADSSONotFound = errorsmod.Register(ModuleName, 1102, "AD SSO linkage not found")

	// ErrADSSOExpired is returned when AD SSO linkage has expired
	ErrADSSOExpired = errorsmod.Register(ModuleName, 1103, "AD SSO linkage has expired")

	// ErrADSSORevoked is returned when AD SSO linkage has been revoked
	ErrADSSORevoked = errorsmod.Register(ModuleName, 1104, "AD SSO linkage has been revoked")

	// ErrADSSOChallengeExpired is returned when AD SSO challenge has expired
	ErrADSSOChallengeExpired = errorsmod.Register(ModuleName, 1105, "AD SSO challenge has expired")

	// ErrADSSOChallengeNotFound is returned when AD SSO challenge is not found
	ErrADSSOChallengeNotFound = errorsmod.Register(ModuleName, 1106, "AD SSO challenge not found")

	// ErrADWalletBindingFailed is returned when AD wallet binding fails
	ErrADWalletBindingFailed = errorsmod.Register(ModuleName, 1107, "AD wallet binding failed")

	// ErrADWalletBindingNotFound is returned when AD wallet binding is not found
	ErrADWalletBindingNotFound = errorsmod.Register(ModuleName, 1108, "AD wallet binding not found")

	// ErrADAuthMethodInvalid is returned when AD auth method is invalid
	ErrADAuthMethodInvalid = errorsmod.Register(ModuleName, 1109, "invalid AD authentication method")

	// ErrADOIDCValidationFailed is returned when OIDC token validation fails
	ErrADOIDCValidationFailed = errorsmod.Register(ModuleName, 1110, "OIDC token validation failed")

	// ErrADSAMLValidationFailed is returned when SAML assertion validation fails
	ErrADSAMLValidationFailed = errorsmod.Register(ModuleName, 1111, "SAML assertion validation failed")

	// ErrADLDAPBindFailed is returned when LDAP bind operation fails
	ErrADLDAPBindFailed = errorsmod.Register(ModuleName, 1112, "LDAP bind failed")

	// ============================================================================
	// SMS Verification Errors (VE-910)
	// ============================================================================

	// ErrInvalidPhone is returned for phone/SMS verification errors
	ErrInvalidPhone = errorsmod.Register(ModuleName, 1113, "invalid phone verification")

	// ErrPhoneNotFound is returned when phone verification is not found
	ErrPhoneNotFound = errorsmod.Register(ModuleName, 1114, "phone verification not found")

	// ErrPhoneExpired is returned when phone verification has expired
	ErrPhoneExpired = errorsmod.Register(ModuleName, 1115, "phone verification has expired")

	// ErrPhoneRevoked is returned when phone verification has been revoked
	ErrPhoneRevoked = errorsmod.Register(ModuleName, 1116, "phone verification has been revoked")

	// ErrSMSOTPExpired is returned when SMS OTP has expired
	ErrSMSOTPExpired = errorsmod.Register(ModuleName, 1117, "SMS OTP has expired")

	// ErrSMSOTPInvalid is returned when SMS OTP is invalid
	ErrSMSOTPInvalid = errorsmod.Register(ModuleName, 1118, "SMS OTP is invalid")

	// ErrSMSChallengeNotFound is returned when SMS challenge is not found
	ErrSMSChallengeNotFound = errorsmod.Register(ModuleName, 1119, "SMS challenge not found")

	// ErrSMSDeliveryFailed is returned when SMS delivery fails
	ErrSMSDeliveryFailed = errorsmod.Register(ModuleName, 1120, "SMS delivery failed")

	// ErrSMSResendLimitExceeded is returned when SMS resend limit is exceeded
	ErrSMSResendLimitExceeded = errorsmod.Register(ModuleName, 1121, "SMS resend limit exceeded")

	// ErrVoIPNumberBlocked is returned when VoIP number is blocked
	ErrVoIPNumberBlocked = errorsmod.Register(ModuleName, 1122, "VoIP numbers are not allowed")

	// ErrCarrierLookupFailed is returned when carrier lookup fails
	ErrCarrierLookupFailed = errorsmod.Register(ModuleName, 1123, "carrier lookup failed")

	// ErrInvalidRateLimit is returned for rate limit configuration errors
	ErrInvalidRateLimit = errorsmod.Register(ModuleName, 1124, "invalid rate limit configuration")

	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errorsmod.Register(ModuleName, 1125, "rate limit exceeded")

	// ErrSMSGatewayNotConfigured is returned when validator SMS gateway is not configured
	ErrSMSGatewayNotConfigured = errorsmod.Register(ModuleName, 1126, "SMS gateway not configured")

	// ErrSMSGatewayUnavailable is returned when SMS gateway is unavailable
	ErrSMSGatewayUnavailable = errorsmod.Register(ModuleName, 1127, "SMS gateway unavailable")

	// ErrPhoneAlreadyVerified is returned when phone is already verified for account
	ErrPhoneAlreadyVerified = errorsmod.Register(ModuleName, 1128, "phone number already verified")

	// ErrCountryNotAllowed is returned when phone country is not allowed
	ErrCountryNotAllowed = errorsmod.Register(ModuleName, 1129, "phone number country not allowed")

	// ============================================================================
	// Delegation Errors (VE-3024)
	// ============================================================================

	// ErrInvalidDelegation is returned when a delegation is malformed or invalid
	ErrInvalidDelegation = errorsmod.Register(ModuleName, 1140, "invalid delegation")

	// ErrDelegationNotFound is returned when a delegation is not found
	ErrDelegationNotFound = errorsmod.Register(ModuleName, 1141, "delegation not found")

	// ErrDelegationExpired is returned when a delegation has expired
	ErrDelegationExpired = errorsmod.Register(ModuleName, 1142, "delegation has expired")

	// ErrDelegationRevoked is returned when a delegation has been revoked
	ErrDelegationRevoked = errorsmod.Register(ModuleName, 1143, "delegation has been revoked")

	// ErrDelegationExhausted is returned when a delegation has no uses remaining
	ErrDelegationExhausted = errorsmod.Register(ModuleName, 1144, "delegation has no uses remaining")

	// ErrDelegationAlreadyRevoked is returned when trying to revoke an already revoked delegation
	ErrDelegationAlreadyRevoked = errorsmod.Register(ModuleName, 1145, "delegation already revoked")

	// ErrDelegationPermissionDenied is returned when delegation lacks required permission
	ErrDelegationPermissionDenied = errorsmod.Register(ModuleName, 1146, "delegation does not grant required permission")

	// ErrDelegationUnauthorized is returned when operation is not authorized for delegation
	ErrDelegationUnauthorized = errorsmod.Register(ModuleName, 1147, "unauthorized delegation operation")

	// ErrDelegationAlreadyExists is returned when trying to create a duplicate delegation
	ErrDelegationAlreadyExists = errorsmod.Register(ModuleName, 1148, "delegation already exists")

	// ============================================================================
	// Verifiable Credential Errors (VE-3025)
	// ============================================================================

	// ErrInvalidCredential is returned when a verifiable credential is malformed or invalid
	ErrInvalidCredential = errorsmod.Register(ModuleName, 1150, "invalid verifiable credential")

	// ErrCredentialNotFound is returned when a credential is not found
	ErrCredentialNotFound = errorsmod.Register(ModuleName, 1151, "credential not found")

	// ErrCredentialExpired is returned when a credential has expired
	ErrCredentialExpired = errorsmod.Register(ModuleName, 1152, "credential has expired")

	// ErrCredentialRevoked is returned when a credential has been revoked
	ErrCredentialRevoked = errorsmod.Register(ModuleName, 1153, "credential has been revoked")

	// ErrCredentialAlreadyRevoked is returned when trying to revoke an already revoked credential
	ErrCredentialAlreadyRevoked = errorsmod.Register(ModuleName, 1154, "credential already revoked")

	// ErrInvalidProof is returned when a credential proof is invalid
	ErrInvalidProof = errorsmod.Register(ModuleName, 1155, "invalid credential proof")

	// ErrProofVerificationFailed is returned when proof verification fails
	ErrProofVerificationFailed = errorsmod.Register(ModuleName, 1156, "proof verification failed")

	// ErrCredentialUnauthorized is returned when operation is not authorized for credential
	ErrCredentialUnauthorized = errorsmod.Register(ModuleName, 1157, "unauthorized credential operation")

	// ErrCredentialAlreadyExists is returned when trying to create a duplicate credential
	ErrCredentialAlreadyExists = errorsmod.Register(ModuleName, 1158, "credential already exists")

	// ErrInvalidIssuer is returned when the credential issuer is invalid or unauthorized
	ErrInvalidIssuer = errorsmod.Register(ModuleName, 1159, "invalid or unauthorized issuer")

	// ============================================================================
	// Privacy-Preserving Proof Errors (VE-3029)
	// Error codes: 1170-1189
	// ============================================================================

	// ErrInvalidProofRequest is returned when a proof request is malformed or invalid
	ErrInvalidProofRequest = errorsmod.Register(ModuleName, 1170, "invalid proof request")

	// ErrProofRequestNotFound is returned when a proof request is not found
	ErrProofRequestNotFound = errorsmod.Register(ModuleName, 1171, "proof request not found")

	// ErrProofRequestExpired is returned when a proof request has expired
	ErrProofRequestExpired = errorsmod.Register(ModuleName, 1172, "proof request has expired")

	// ErrProofNotFound is returned when a proof is not found
	ErrProofNotFound = errorsmod.Register(ModuleName, 1173, "proof not found")

	// ErrProofExpired is returned when a proof has expired
	ErrProofExpired = errorsmod.Register(ModuleName, 1174, "proof has expired")

	// ErrInvalidClaimType is returned when a claim type is invalid
	ErrInvalidClaimType = errorsmod.Register(ModuleName, 1175, "invalid claim type")

	// ErrClaimNotAvailable is returned when a requested claim is not available for the subject
	ErrClaimNotAvailable = errorsmod.Register(ModuleName, 1176, "claim not available for subject")

	// ErrInvalidProofScheme is returned when a proof scheme is invalid or unsupported
	ErrInvalidProofScheme = errorsmod.Register(ModuleName, 1177, "invalid or unsupported proof scheme")

	// ErrProofGenerationFailed is returned when proof generation fails
	ErrProofGenerationFailed = errorsmod.Register(ModuleName, 1178, "proof generation failed")

	// ErrInvalidNonce is returned when a nonce is invalid or already used
	ErrInvalidNonce = errorsmod.Register(ModuleName, 1179, "invalid or reused nonce")

	// ErrCommitmentMismatch is returned when commitment verification fails
	ErrCommitmentMismatch = errorsmod.Register(ModuleName, 1180, "commitment mismatch")

	// ErrInsufficientVerificationLevel is returned when subject lacks required verification
	ErrInsufficientVerificationLevel = errorsmod.Register(ModuleName, 1181, "insufficient verification level for claim")

	// ============================================================================
	// Biometric Hash Errors (VE-3030)
	// Error codes: 1190-1199
	// ============================================================================

	// ErrInvalidBiometricTemplate is returned when a biometric template is invalid
	ErrInvalidBiometricTemplate = errorsmod.Register(ModuleName, 1190, "invalid biometric template")

	// ErrBiometricHashFailed is returned when biometric hashing fails
	ErrBiometricHashFailed = errorsmod.Register(ModuleName, 1191, "biometric hash operation failed")

	// ErrBiometricHashNotFound is returned when a biometric hash is not found
	ErrBiometricHashNotFound = errorsmod.Register(ModuleName, 1192, "biometric hash not found")

	// ErrBiometricMatchFailed is returned when biometric matching fails
	ErrBiometricMatchFailed = errorsmod.Register(ModuleName, 1193, "biometric match operation failed")

	// ErrBiometricVersionMismatch is returned when biometric hash versions are incompatible
	ErrBiometricVersionMismatch = errorsmod.Register(ModuleName, 1194, "biometric hash version mismatch")

	// ============================================================================
	// Geographic Restriction Errors (VE-3032)
	// Error codes: 1200-1219
	// ============================================================================

	// ErrGeoRestrictionInvalid is returned when a geographic restriction is invalid
	ErrGeoRestrictionInvalid = errorsmod.Register(ModuleName, 1200, "invalid geographic restriction")

	// ErrGeoPolicyNotFound is returned when a geo restriction policy is not found
	ErrGeoPolicyNotFound = errorsmod.Register(ModuleName, 1201, "geo restriction policy not found")

	// ErrGeoPolicyAlreadyExists is returned when a geo restriction policy already exists
	ErrGeoPolicyAlreadyExists = errorsmod.Register(ModuleName, 1202, "geo restriction policy already exists")

	// ErrGeoCheckFailed is returned when a geographic compliance check fails
	ErrGeoCheckFailed = errorsmod.Register(ModuleName, 1203, "geographic compliance check failed")

	// ErrGeoLocationBlocked is returned when a location is blocked by policy
	ErrGeoLocationBlocked = errorsmod.Register(ModuleName, 1204, "location blocked by geo restriction policy")

	// ErrGeoIPMismatch is returned when IP geolocation doesn't match document country
	ErrGeoIPMismatch = errorsmod.Register(ModuleName, 1205, "IP geolocation mismatch with document country")

	// ErrGeoCountryBlocked is returned when a specific country is blocked
	ErrGeoCountryBlocked = errorsmod.Register(ModuleName, 1206, "country is blocked by geo restriction policy")

	// ErrGeoRegionBlocked is returned when a specific region is blocked
	ErrGeoRegionBlocked = errorsmod.Register(ModuleName, 1207, "region is blocked by geo restriction policy")

	// ErrGeoRestrictionDisabled is returned when geo restriction system is disabled
	ErrGeoRestrictionDisabled = errorsmod.Register(ModuleName, 1208, "geographic restriction system is disabled")

	// ErrMaxGeoPoliciesExceeded is returned when max policies limit is exceeded
	ErrMaxGeoPoliciesExceeded = errorsmod.Register(ModuleName, 1209, "maximum geo restriction policies exceeded")

	// ErrGeoPolicyInactive is returned when trying to use an inactive policy
	ErrGeoPolicyInactive = errorsmod.Register(ModuleName, 1210, "geo restriction policy is not active")

	// ErrInvalidCountryCode is returned when an ISO 3166 country code is invalid
	ErrInvalidCountryCode = errorsmod.Register(ModuleName, 1211, "invalid ISO 3166-1 alpha-2 country code")

	// ErrInvalidRegionCode is returned when an ISO 3166-2 region code is invalid
	ErrInvalidRegionCode = errorsmod.Register(ModuleName, 1212, "invalid ISO 3166-2 region code")

	// ErrGeoLocationNotFound is returned when geo location data is not found for user
	ErrGeoLocationNotFound = errorsmod.Register(ModuleName, 1213, "geographic location not found for user")

	// ErrGeoParamsInvalid is returned when geo restriction parameters are invalid
	ErrGeoParamsInvalid = errorsmod.Register(ModuleName, 1214, "invalid geo restriction parameters")
)
