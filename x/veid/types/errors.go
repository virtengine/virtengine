package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the veid module
var (
	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1, "invalid address")

	// ErrInvalidScope is returned when a scope is malformed
	ErrInvalidScope = errorsmod.Register(ModuleName, 2, "invalid scope")

	// ErrInvalidScopeType is returned when a scope type is invalid
	ErrInvalidScopeType = errorsmod.Register(ModuleName, 3, "invalid scope type")

	// ErrInvalidScopeVersion is returned when a scope version is invalid
	ErrInvalidScopeVersion = errorsmod.Register(ModuleName, 4, "invalid scope version")

	// ErrInvalidPayload is returned when an encrypted payload is invalid
	ErrInvalidPayload = errorsmod.Register(ModuleName, 5, "invalid encrypted payload")

	// ErrInvalidSalt is returned when a salt is invalid
	ErrInvalidSalt = errorsmod.Register(ModuleName, 6, "invalid salt")

	// ErrSaltAlreadyUsed is returned when a salt has already been used
	ErrSaltAlreadyUsed = errorsmod.Register(ModuleName, 7, "salt already used")

	// ErrInvalidDeviceInfo is returned when device information is invalid
	ErrInvalidDeviceInfo = errorsmod.Register(ModuleName, 8, "invalid device info")

	// ErrInvalidClientID is returned when a client ID is invalid
	ErrInvalidClientID = errorsmod.Register(ModuleName, 9, "invalid client ID")

	// ErrClientNotApproved is returned when a client is not approved
	ErrClientNotApproved = errorsmod.Register(ModuleName, 10, "client not approved")

	// ErrInvalidClientSignature is returned when a client signature is invalid
	ErrInvalidClientSignature = errorsmod.Register(ModuleName, 11, "invalid client signature")

	// ErrInvalidUserSignature is returned when a user signature is invalid
	ErrInvalidUserSignature = errorsmod.Register(ModuleName, 12, "invalid user signature")

	// ErrInvalidPayloadHash is returned when a payload hash is invalid
	ErrInvalidPayloadHash = errorsmod.Register(ModuleName, 13, "invalid payload hash")

	// ErrInvalidVerificationStatus is returned when a verification status is invalid
	ErrInvalidVerificationStatus = errorsmod.Register(ModuleName, 14, "invalid verification status")

	// ErrInvalidVerificationEvent is returned when a verification event is invalid
	ErrInvalidVerificationEvent = errorsmod.Register(ModuleName, 15, "invalid verification event")

	// ErrInvalidScore is returned when an identity score is invalid
	ErrInvalidScore = errorsmod.Register(ModuleName, 16, "invalid score")

	// ErrInvalidTier is returned when an identity tier is invalid
	ErrInvalidTier = errorsmod.Register(ModuleName, 17, "invalid tier")

	// ErrInvalidIdentityRecord is returned when an identity record is invalid
	ErrInvalidIdentityRecord = errorsmod.Register(ModuleName, 18, "invalid identity record")

	// ErrInvalidWallet is returned when an identity wallet is invalid
	ErrInvalidWallet = errorsmod.Register(ModuleName, 19, "invalid identity wallet")

	// ErrScopeNotFound is returned when a scope is not found
	ErrScopeNotFound = errorsmod.Register(ModuleName, 20, "scope not found")

	// ErrIdentityRecordNotFound is returned when an identity record is not found
	ErrIdentityRecordNotFound = errorsmod.Register(ModuleName, 21, "identity record not found")

	// ErrScopeAlreadyExists is returned when a scope already exists
	ErrScopeAlreadyExists = errorsmod.Register(ModuleName, 22, "scope already exists")

	// ErrScopeRevoked is returned when trying to use a revoked scope
	ErrScopeRevoked = errorsmod.Register(ModuleName, 23, "scope has been revoked")

	// ErrScopeExpired is returned when trying to use an expired scope
	ErrScopeExpired = errorsmod.Register(ModuleName, 24, "scope has expired")

	// ErrUnauthorized is returned when the sender is not authorized
	ErrUnauthorized = errorsmod.Register(ModuleName, 25, "unauthorized")

	// ErrInvalidStatusTransition is returned when an invalid status transition is attempted
	ErrInvalidStatusTransition = errorsmod.Register(ModuleName, 26, "invalid status transition")

	// ErrIdentityLocked is returned when trying to modify a locked identity
	ErrIdentityLocked = errorsmod.Register(ModuleName, 27, "identity is locked")

	// ErrMaxScopesExceeded is returned when the maximum number of scopes is exceeded
	ErrMaxScopesExceeded = errorsmod.Register(ModuleName, 28, "maximum scopes exceeded")

	// ErrVerificationInProgress is returned when verification is already in progress
	ErrVerificationInProgress = errorsmod.Register(ModuleName, 29, "verification already in progress")

	// ErrValidatorOnly is returned when a non-validator attempts a validator-only action
	ErrValidatorOnly = errorsmod.Register(ModuleName, 30, "action restricted to validators")

	// ErrSignatureMismatch is returned when signatures don't match expected values
	ErrSignatureMismatch = errorsmod.Register(ModuleName, 31, "signature mismatch")

	// ErrInvalidParams is returned when module parameters are invalid
	ErrInvalidParams = errorsmod.Register(ModuleName, 32, "invalid parameters")

	// ErrInvalidVerificationRequest is returned when a verification request is invalid
	ErrInvalidVerificationRequest = errorsmod.Register(ModuleName, 33, "invalid verification request")

	// ErrInvalidVerificationResult is returned when a verification result is invalid
	ErrInvalidVerificationResult = errorsmod.Register(ModuleName, 34, "invalid verification result")

	// ErrVerificationRequestNotFound is returned when a verification request is not found
	ErrVerificationRequestNotFound = errorsmod.Register(ModuleName, 35, "verification request not found")

	// ErrDecryptionFailed is returned when scope decryption fails
	ErrDecryptionFailed = errorsmod.Register(ModuleName, 36, "scope decryption failed")

	// ErrMLInferenceFailed is returned when ML scoring fails
	ErrMLInferenceFailed = errorsmod.Register(ModuleName, 37, "ML inference failed")

	// ErrVerificationTimeout is returned when verification times out
	ErrVerificationTimeout = errorsmod.Register(ModuleName, 38, "verification timeout")

	// ErrMaxRetriesExceeded is returned when max retries are exceeded
	ErrMaxRetriesExceeded = errorsmod.Register(ModuleName, 39, "max retries exceeded")

	// ErrNotBlockProposer is returned when non-proposer attempts proposer-only action
	ErrNotBlockProposer = errorsmod.Register(ModuleName, 40, "not block proposer")

	// ErrValidatorKeyNotFound is returned when validator identity key is not found
	ErrValidatorKeyNotFound = errorsmod.Register(ModuleName, 41, "validator identity key not found")

	// ErrWalletNotFound is returned when an identity wallet is not found
	ErrWalletNotFound = errorsmod.Register(ModuleName, 42, "identity wallet not found")

	// ErrWalletAlreadyExists is returned when a wallet already exists for an account
	ErrWalletAlreadyExists = errorsmod.Register(ModuleName, 43, "identity wallet already exists")

	// ErrWalletInactive is returned when trying to use an inactive wallet
	ErrWalletInactive = errorsmod.Register(ModuleName, 44, "identity wallet is inactive")

	// ErrConsentNotGranted is returned when required consent is not granted
	ErrConsentNotGranted = errorsmod.Register(ModuleName, 45, "consent not granted")

	// ErrConsentExpired is returned when consent has expired
	ErrConsentExpired = errorsmod.Register(ModuleName, 46, "consent has expired")

	// ErrInvalidBindingSignature is returned when wallet binding signature is invalid
	ErrInvalidBindingSignature = errorsmod.Register(ModuleName, 47, "invalid wallet binding signature")

	// ErrScopeNotInWallet is returned when a scope is not found in the wallet
	ErrScopeNotInWallet = errorsmod.Register(ModuleName, 48, "scope not found in wallet")

	// ErrScopeAlreadyInWallet is returned when a scope is already in the wallet
	ErrScopeAlreadyInWallet = errorsmod.Register(ModuleName, 49, "scope already in wallet")

	// ErrDerivedFeaturesEmpty is returned when derived features are empty
	ErrDerivedFeaturesEmpty = errorsmod.Register(ModuleName, 50, "derived features are empty")

	// ErrVerificationMismatch is returned when consensus verification result mismatches
	ErrVerificationMismatch = errorsmod.Register(ModuleName, 51, "verification result mismatch")

	// ErrModelVersionMismatch is returned when ML model versions don't match
	ErrModelVersionMismatch = errorsmod.Register(ModuleName, 52, "model version mismatch")

	// ErrInputHashMismatch is returned when input hashes don't match
	ErrInputHashMismatch = errorsmod.Register(ModuleName, 53, "input hash mismatch")

	// ErrScoreToleranceExceeded is returned when score difference exceeds tolerance
	ErrScoreToleranceExceeded = errorsmod.Register(ModuleName, 54, "score tolerance exceeded")

	// ErrVoteExtensionInvalid is returned when a vote extension is invalid
	ErrVoteExtensionInvalid = errorsmod.Register(ModuleName, 55, "invalid vote extension")

	// ErrInvalidBorderlineFallback is returned when a borderline fallback is invalid
	ErrInvalidBorderlineFallback = errorsmod.Register(ModuleName, 56, "invalid borderline fallback")

	// ErrBorderlineFallbackNotFound is returned when a borderline fallback is not found
	ErrBorderlineFallbackNotFound = errorsmod.Register(ModuleName, 57, "borderline fallback not found")

	// ErrBorderlineFallbackExpired is returned when a borderline fallback has expired
	ErrBorderlineFallbackExpired = errorsmod.Register(ModuleName, 58, "borderline fallback expired")

	// ErrBorderlineFallbackAlreadyCompleted is returned when fallback is already completed
	ErrBorderlineFallbackAlreadyCompleted = errorsmod.Register(ModuleName, 59, "borderline fallback already completed")

	// ErrMFAChallengeNotSatisfied is returned when MFA challenge is not satisfied
	ErrMFAChallengeNotSatisfied = errorsmod.Register(ModuleName, 60, "MFA challenge not satisfied")

	// ErrNoEnrolledFactors is returned when account has no enrolled MFA factors
	ErrNoEnrolledFactors = errorsmod.Register(ModuleName, 61, "no enrolled MFA factors")

	// ErrBorderlineDisabled is returned when borderline fallback is disabled
	ErrBorderlineDisabled = errorsmod.Register(ModuleName, 62, "borderline fallback is disabled")

	// ============================================================================
	// Pipeline Version Errors (VE-219)
	// ============================================================================

	// ErrInvalidPipelineVersion is returned when a pipeline version is invalid
	ErrInvalidPipelineVersion = errorsmod.Register(ModuleName, 63, "invalid pipeline version")

	// ErrPipelineVersionNotFound is returned when a pipeline version is not found
	ErrPipelineVersionNotFound = errorsmod.Register(ModuleName, 64, "pipeline version not found")

	// ErrPipelineVersionAlreadyExists is returned when a pipeline version already exists
	ErrPipelineVersionAlreadyExists = errorsmod.Register(ModuleName, 65, "pipeline version already exists")

	// ErrNoPipelineVersionActive is returned when no active pipeline version exists
	ErrNoPipelineVersionActive = errorsmod.Register(ModuleName, 66, "no active pipeline version")

	// ErrPipelineVersionMismatch is returned when pipeline versions don't match for consensus
	ErrPipelineVersionMismatch = errorsmod.Register(ModuleName, 67, "pipeline version mismatch")

	// ErrModelManifestMismatch is returned when model manifests don't match
	ErrModelManifestMismatch = errorsmod.Register(ModuleName, 68, "model manifest mismatch")

	// ErrDeterminismViolation is returned when deterministic execution check fails
	ErrDeterminismViolation = errorsmod.Register(ModuleName, 69, "determinism violation detected")

	// ErrInvalidModelManifest is returned when a model manifest is invalid
	ErrInvalidModelManifest = errorsmod.Register(ModuleName, 70, "invalid model manifest")

	// ErrPipelineExecutionFailed is returned when pipeline execution fails
	ErrPipelineExecutionFailed = errorsmod.Register(ModuleName, 71, "pipeline execution failed")

	// ============================================================================
	// Scoring Model Errors (VE-220)
	// ============================================================================

	// ErrInvalidScoringModel is returned when a scoring model is invalid
	ErrInvalidScoringModel = errorsmod.Register(ModuleName, 72, "invalid scoring model")

	// ErrScoringModelNotFound is returned when a scoring model version is not found
	ErrScoringModelNotFound = errorsmod.Register(ModuleName, 73, "scoring model not found")

	// ErrScoringModelAlreadyExists is returned when a scoring model version already exists
	ErrScoringModelAlreadyExists = errorsmod.Register(ModuleName, 74, "scoring model already exists")

	// ErrInvalidScoringInput is returned when scoring input is invalid
	ErrInvalidScoringInput = errorsmod.Register(ModuleName, 75, "invalid scoring input")

	// ErrScoringFailed is returned when score computation fails
	ErrScoringFailed = errorsmod.Register(ModuleName, 76, "scoring computation failed")

	// ErrNoActiveScoringModel is returned when no active scoring model exists
	ErrNoActiveScoringModel = errorsmod.Register(ModuleName, 77, "no active scoring model")

	// ErrScoringModelVersionMismatch is returned when scoring model versions don't match
	ErrScoringModelVersionMismatch = errorsmod.Register(ModuleName, 78, "scoring model version mismatch")

	// ============================================================================
	// Verification Scope Errors (VE-222, VE-223, VE-224)
	// ============================================================================

	// ErrInvalidSSO is returned for SSO verification errors
	ErrInvalidSSO = errorsmod.Register(ModuleName, 79, "invalid SSO verification")

	// ErrSSONotFound is returned when SSO linkage is not found
	ErrSSONotFound = errorsmod.Register(ModuleName, 80, "SSO linkage not found")

	// ErrSSOExpired is returned when SSO linkage has expired
	ErrSSOExpired = errorsmod.Register(ModuleName, 81, "SSO linkage has expired")

	// ErrSSORevoked is returned when SSO linkage has been revoked
	ErrSSORevoked = errorsmod.Register(ModuleName, 82, "SSO linkage has been revoked")

	// ErrInvalidDomain is returned for domain verification errors
	ErrInvalidDomain = errorsmod.Register(ModuleName, 83, "invalid domain verification")

	// ErrDomainNotFound is returned when domain verification is not found
	ErrDomainNotFound = errorsmod.Register(ModuleName, 84, "domain verification not found")

	// ErrDomainExpired is returned when domain verification has expired
	ErrDomainExpired = errorsmod.Register(ModuleName, 85, "domain verification has expired")

	// ErrDomainRevoked is returned when domain verification has been revoked
	ErrDomainRevoked = errorsmod.Register(ModuleName, 86, "domain verification has been revoked")

	// ErrInvalidEmail is returned for email verification errors
	ErrInvalidEmail = errorsmod.Register(ModuleName, 87, "invalid email verification")

	// ErrEmailNotFound is returned when email verification is not found
	ErrEmailNotFound = errorsmod.Register(ModuleName, 88, "email verification not found")

	// ErrEmailExpired is returned when email verification has expired
	ErrEmailExpired = errorsmod.Register(ModuleName, 89, "email verification has expired")

	// ErrNonceAlreadyUsed is returned when a nonce has already been used (anti-replay)
	ErrNonceAlreadyUsed = errorsmod.Register(ModuleName, 90, "nonce already used")

	// ErrMaxAttemptsExceeded is returned when max verification attempts exceeded
	ErrMaxAttemptsExceeded = errorsmod.Register(ModuleName, 91, "maximum verification attempts exceeded")

	// ============================================================================
	// Security Controls Errors (VE-225)
	// ============================================================================

	// ErrInvalidToken is returned for tokenization errors
	ErrInvalidToken = errorsmod.Register(ModuleName, 92, "invalid token")

	// ErrTokenNotFound is returned when token mapping is not found
	ErrTokenNotFound = errorsmod.Register(ModuleName, 93, "token not found")

	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errorsmod.Register(ModuleName, 94, "token has expired")

	// ErrTokenRevoked is returned when token has been revoked
	ErrTokenRevoked = errorsmod.Register(ModuleName, 95, "token has been revoked")

	// ErrInvalidRetention is returned for retention policy errors
	ErrInvalidRetention = errorsmod.Register(ModuleName, 96, "invalid retention policy")

	// ErrRetentionNotFound is returned when retention policy is not found
	ErrRetentionNotFound = errorsmod.Register(ModuleName, 97, "retention policy not found")

	// ============================================================================
	// Waldur Integration Errors (VE-226)
	// ============================================================================

	// ErrInvalidWaldur is returned for Waldur integration errors
	ErrInvalidWaldur = errorsmod.Register(ModuleName, 98, "invalid Waldur integration")

	// ErrWaldurLinkNotFound is returned when Waldur link is not found
	ErrWaldurLinkNotFound = errorsmod.Register(ModuleName, 99, "Waldur link not found")

	// ErrWaldurLinkExpired is returned when Waldur request has expired
	ErrWaldurLinkExpired = errorsmod.Register(ModuleName, 100, "Waldur request has expired")

	// ErrWaldurSyncFailed is returned when Waldur sync fails
	ErrWaldurSyncFailed = errorsmod.Register(ModuleName, 101, "Waldur sync failed")

	// ============================================================================
	// Active Directory SSO Errors (VE-907)
	// ============================================================================

	// ErrInvalidADSSO is returned for AD SSO verification errors
	ErrInvalidADSSO = errorsmod.Register(ModuleName, 102, "invalid AD SSO verification")

	// ErrADSSONotFound is returned when AD SSO linkage is not found
	ErrADSSONotFound = errorsmod.Register(ModuleName, 103, "AD SSO linkage not found")

	// ErrADSSOExpired is returned when AD SSO linkage has expired
	ErrADSSOExpired = errorsmod.Register(ModuleName, 104, "AD SSO linkage has expired")

	// ErrADSSORevoked is returned when AD SSO linkage has been revoked
	ErrADSSORevoked = errorsmod.Register(ModuleName, 105, "AD SSO linkage has been revoked")

	// ErrADSSOChallengeExpired is returned when AD SSO challenge has expired
	ErrADSSOChallengeExpired = errorsmod.Register(ModuleName, 106, "AD SSO challenge has expired")

	// ErrADSSOChallengeNotFound is returned when AD SSO challenge is not found
	ErrADSSOChallengeNotFound = errorsmod.Register(ModuleName, 107, "AD SSO challenge not found")

	// ErrADWalletBindingFailed is returned when AD wallet binding fails
	ErrADWalletBindingFailed = errorsmod.Register(ModuleName, 108, "AD wallet binding failed")

	// ErrADWalletBindingNotFound is returned when AD wallet binding is not found
	ErrADWalletBindingNotFound = errorsmod.Register(ModuleName, 109, "AD wallet binding not found")

	// ErrADAuthMethodInvalid is returned when AD auth method is invalid
	ErrADAuthMethodInvalid = errorsmod.Register(ModuleName, 110, "invalid AD authentication method")

	// ErrADOIDCValidationFailed is returned when OIDC token validation fails
	ErrADOIDCValidationFailed = errorsmod.Register(ModuleName, 111, "OIDC token validation failed")

	// ErrADSAMLValidationFailed is returned when SAML assertion validation fails
	ErrADSAMLValidationFailed = errorsmod.Register(ModuleName, 112, "SAML assertion validation failed")

	// ErrADLDAPBindFailed is returned when LDAP bind operation fails
	ErrADLDAPBindFailed = errorsmod.Register(ModuleName, 113, "LDAP bind failed")

	// ============================================================================
	// SMS Verification Errors (VE-910)
	// ============================================================================

	// ErrInvalidPhone is returned for phone/SMS verification errors
	ErrInvalidPhone = errorsmod.Register(ModuleName, 114, "invalid phone verification")

	// ErrPhoneNotFound is returned when phone verification is not found
	ErrPhoneNotFound = errorsmod.Register(ModuleName, 115, "phone verification not found")

	// ErrPhoneExpired is returned when phone verification has expired
	ErrPhoneExpired = errorsmod.Register(ModuleName, 116, "phone verification has expired")

	// ErrPhoneRevoked is returned when phone verification has been revoked
	ErrPhoneRevoked = errorsmod.Register(ModuleName, 117, "phone verification has been revoked")

	// ErrSMSOTPExpired is returned when SMS OTP has expired
	ErrSMSOTPExpired = errorsmod.Register(ModuleName, 118, "SMS OTP has expired")

	// ErrSMSOTPInvalid is returned when SMS OTP is invalid
	ErrSMSOTPInvalid = errorsmod.Register(ModuleName, 119, "SMS OTP is invalid")

	// ErrSMSChallengeNotFound is returned when SMS challenge is not found
	ErrSMSChallengeNotFound = errorsmod.Register(ModuleName, 120, "SMS challenge not found")

	// ErrSMSDeliveryFailed is returned when SMS delivery fails
	ErrSMSDeliveryFailed = errorsmod.Register(ModuleName, 121, "SMS delivery failed")

	// ErrSMSResendLimitExceeded is returned when SMS resend limit is exceeded
	ErrSMSResendLimitExceeded = errorsmod.Register(ModuleName, 122, "SMS resend limit exceeded")

	// ErrVoIPNumberBlocked is returned when VoIP number is blocked
	ErrVoIPNumberBlocked = errorsmod.Register(ModuleName, 123, "VoIP numbers are not allowed")

	// ErrCarrierLookupFailed is returned when carrier lookup fails
	ErrCarrierLookupFailed = errorsmod.Register(ModuleName, 124, "carrier lookup failed")

	// ErrInvalidRateLimit is returned for rate limit configuration errors
	ErrInvalidRateLimit = errorsmod.Register(ModuleName, 125, "invalid rate limit configuration")

	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errorsmod.Register(ModuleName, 126, "rate limit exceeded")

	// ErrSMSGatewayNotConfigured is returned when validator SMS gateway is not configured
	ErrSMSGatewayNotConfigured = errorsmod.Register(ModuleName, 127, "SMS gateway not configured")

	// ErrSMSGatewayUnavailable is returned when SMS gateway is unavailable
	ErrSMSGatewayUnavailable = errorsmod.Register(ModuleName, 128, "SMS gateway unavailable")

	// ErrPhoneAlreadyVerified is returned when phone is already verified for account
	ErrPhoneAlreadyVerified = errorsmod.Register(ModuleName, 129, "phone number already verified")

	// ErrCountryNotAllowed is returned when phone country is not allowed
	ErrCountryNotAllowed = errorsmod.Register(ModuleName, 130, "phone number country not allowed")
)
