package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Error codes for the mfa module
var (
	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errorsmod.Register(ModuleName, 1, "invalid address")

	// ErrInvalidFactorType is returned when a factor type is invalid
	ErrInvalidFactorType = errorsmod.Register(ModuleName, 2, "invalid factor type")

	// ErrInvalidPolicy is returned when an MFA policy is invalid
	ErrInvalidPolicy = errorsmod.Register(ModuleName, 3, "invalid MFA policy")

	// ErrPolicyNotFound is returned when an MFA policy is not found
	ErrPolicyNotFound = errorsmod.Register(ModuleName, 4, "MFA policy not found")

	// ErrInvalidEnrollment is returned when a factor enrollment is invalid
	ErrInvalidEnrollment = errorsmod.Register(ModuleName, 5, "invalid factor enrollment")

	// ErrEnrollmentNotFound is returned when a factor enrollment is not found
	ErrEnrollmentNotFound = errorsmod.Register(ModuleName, 6, "factor enrollment not found")

	// ErrEnrollmentAlreadyExists is returned when a factor enrollment already exists
	ErrEnrollmentAlreadyExists = errorsmod.Register(ModuleName, 7, "factor enrollment already exists")

	// ErrInvalidChallenge is returned when a challenge is invalid
	ErrInvalidChallenge = errorsmod.Register(ModuleName, 8, "invalid challenge")

	// ErrChallengeNotFound is returned when a challenge is not found
	ErrChallengeNotFound = errorsmod.Register(ModuleName, 9, "challenge not found")

	// ErrChallengeExpired is returned when a challenge has expired
	ErrChallengeExpired = errorsmod.Register(ModuleName, 10, "challenge has expired")

	// ErrChallengeAlreadyUsed is returned when a challenge has already been used
	ErrChallengeAlreadyUsed = errorsmod.Register(ModuleName, 11, "challenge already used")

	// ErrMaxAttemptsExceeded is returned when max verification attempts exceeded
	ErrMaxAttemptsExceeded = errorsmod.Register(ModuleName, 12, "maximum verification attempts exceeded")

	// ErrInvalidChallengeResponse is returned when challenge response is invalid
	ErrInvalidChallengeResponse = errorsmod.Register(ModuleName, 13, "invalid challenge response")

	// ErrVerificationFailed is returned when verification fails
	ErrVerificationFailed = errorsmod.Register(ModuleName, 14, "verification failed")

	// ErrMFARequired is returned when MFA is required but not provided
	ErrMFARequired = errorsmod.Register(ModuleName, 15, "MFA verification required")

	// ErrInsufficientFactors is returned when not enough factors are verified
	ErrInsufficientFactors = errorsmod.Register(ModuleName, 16, "insufficient factors verified")

	// ErrSessionNotFound is returned when an authorization session is not found
	ErrSessionNotFound = errorsmod.Register(ModuleName, 17, "authorization session not found")

	// ErrSessionExpired is returned when an authorization session has expired
	ErrSessionExpired = errorsmod.Register(ModuleName, 18, "authorization session has expired")

	// ErrSessionAlreadyUsed is returned when a single-use session has been used
	ErrSessionAlreadyUsed = errorsmod.Register(ModuleName, 19, "authorization session already used")

	// ErrUnauthorized is returned when the sender is not authorized
	ErrUnauthorized = errorsmod.Register(ModuleName, 20, "unauthorized")

	// ErrInvalidSensitiveTxType is returned when a sensitive transaction type is invalid
	ErrInvalidSensitiveTxType = errorsmod.Register(ModuleName, 21, "invalid sensitive transaction type")

	// ErrInvalidSensitiveTxConfig is returned when a sensitive tx config is invalid
	ErrInvalidSensitiveTxConfig = errorsmod.Register(ModuleName, 22, "invalid sensitive transaction config")

	// ErrTrustedDeviceNotFound is returned when a trusted device is not found
	ErrTrustedDeviceNotFound = errorsmod.Register(ModuleName, 23, "trusted device not found")

	// ErrTrustedDeviceExpired is returned when a trusted device has expired
	ErrTrustedDeviceExpired = errorsmod.Register(ModuleName, 24, "trusted device has expired")

	// ErrMaxTrustedDevicesReached is returned when max trusted devices limit reached
	ErrMaxTrustedDevicesReached = errorsmod.Register(ModuleName, 25, "maximum trusted devices limit reached")

	// ErrChallengeCreationFailed is returned when challenge creation fails
	ErrChallengeCreationFailed = errorsmod.Register(ModuleName, 26, "challenge creation failed")

	// ErrMFADisabled is returned when MFA is disabled for the account
	ErrMFADisabled = errorsmod.Register(ModuleName, 27, "MFA is disabled for this account")

	// ErrFactorRevoked is returned when attempting to use a revoked factor
	ErrFactorRevoked = errorsmod.Register(ModuleName, 28, "factor has been revoked")

	// ErrFactorExpired is returned when attempting to use an expired factor
	ErrFactorExpired = errorsmod.Register(ModuleName, 29, "factor has expired")

	// ErrVEIDScoreInsufficient is returned when VEID score is below threshold
	ErrVEIDScoreInsufficient = errorsmod.Register(ModuleName, 30, "VEID score is below required threshold")

	// ErrDeviceMismatch is returned when device doesn't match session binding
	ErrDeviceMismatch = errorsmod.Register(ModuleName, 31, "device fingerprint does not match session")

	// ErrCooldownActive is returned when an operation is in cooldown period
	ErrCooldownActive = errorsmod.Register(ModuleName, 32, "operation is in cooldown period")

	// ErrInvalidProof is returned when MFA proof is invalid
	ErrInvalidProof = errorsmod.Register(ModuleName, 33, "invalid MFA proof")

	// ErrNoActiveFactors is returned when account has no active factors enrolled
	ErrNoActiveFactors = errorsmod.Register(ModuleName, 34, "no active factors enrolled")

	// ============================================================================
	// Hardware Key MFA Errors (VE-925)
	// ============================================================================

	// ErrInvalidCertificate is returned when a certificate is invalid
	ErrInvalidCertificate = errorsmod.Register(ModuleName, 35, "invalid certificate")

	// ErrCertificateExpired is returned when a certificate has expired
	ErrCertificateExpired = errorsmod.Register(ModuleName, 36, "certificate has expired")

	// ErrCertificateNotYetValid is returned when a certificate is not yet valid
	ErrCertificateNotYetValid = errorsmod.Register(ModuleName, 37, "certificate is not yet valid")

	// ErrCertificateRevoked is returned when a certificate has been revoked
	ErrCertificateRevoked = errorsmod.Register(ModuleName, 38, "certificate has been revoked")

	// ErrCertificateChainInvalid is returned when a certificate chain is invalid
	ErrCertificateChainInvalid = errorsmod.Register(ModuleName, 39, "certificate chain is invalid")

	// ErrRevocationCheckFailed is returned when revocation checking fails
	ErrRevocationCheckFailed = errorsmod.Register(ModuleName, 40, "certificate revocation check failed")

	// ErrSmartCardAuthFailed is returned when smart card authentication fails
	ErrSmartCardAuthFailed = errorsmod.Register(ModuleName, 41, "smart card authentication failed")

	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errorsmod.Register(ModuleName, 42, "invalid signature")

	// ErrKeyUsageNotAllowed is returned when key usage is not allowed
	ErrKeyUsageNotAllowed = errorsmod.Register(ModuleName, 43, "key usage not allowed for this operation")

	// ErrPINRequired is returned when PIN verification is required
	ErrPINRequired = errorsmod.Register(ModuleName, 44, "PIN verification required")

	// ErrSmartCardExpired is returned when a smart card has expired
	ErrSmartCardExpired = errorsmod.Register(ModuleName, 45, "smart card has expired")
)
