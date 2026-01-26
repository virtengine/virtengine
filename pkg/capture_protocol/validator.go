package capture_protocol

import (
	"time"
)

// ProtocolValidator is the main validator for the Mobile Capture Protocol v1.
// It orchestrates salt validation, signature verification, and anti-replay checks.
type ProtocolValidator struct {
	// Salt validation
	saltValidator *SaltValidator

	// Signature validation
	signatureValidator *SignatureValidator

	// Replay protection
	replayProtector *ReplayProtector

	// Configuration
	config ValidationConfig

	// Time source (for testing)
	now func() time.Time
}

// ProtocolValidatorOption is a functional option for ProtocolValidator
type ProtocolValidatorOption func(*ProtocolValidator)

// WithValidationConfig sets the validation configuration
func WithValidationConfig(config ValidationConfig) ProtocolValidatorOption {
	return func(pv *ProtocolValidator) {
		pv.config = config
	}
}

// WithCustomSaltValidator sets a custom salt validator
func WithCustomSaltValidator(sv *SaltValidator) ProtocolValidatorOption {
	return func(pv *ProtocolValidator) {
		pv.saltValidator = sv
	}
}

// WithCustomSignatureValidator sets a custom signature validator
func WithCustomSignatureValidator(sv *SignatureValidator) ProtocolValidatorOption {
	return func(pv *ProtocolValidator) {
		pv.signatureValidator = sv
	}
}

// WithCustomReplayProtector sets a custom replay protector
func WithCustomReplayProtector(rp *ReplayProtector) ProtocolValidatorOption {
	return func(pv *ProtocolValidator) {
		pv.replayProtector = rp
	}
}

// NewProtocolValidator creates a new ProtocolValidator
func NewProtocolValidator(
	registry ApprovedClientRegistry,
	opts ...ProtocolValidatorOption,
) *ProtocolValidator {
	pv := &ProtocolValidator{
		config: DefaultValidationConfig(),
		now:    time.Now,
	}

	for _, opt := range opts {
		opt(pv)
	}

	// Initialize salt validator if not provided
	if pv.saltValidator == nil {
		pv.saltValidator = NewSaltValidator(
			WithMinSaltLength(pv.config.MinSaltLength),
			WithMaxSaltAge(pv.config.MaxSaltAge),
			WithReplayWindow(pv.config.ReplayWindow),
			WithMaxClockSkew(pv.config.MaxClockSkew),
		)
	}

	// Initialize signature validator if not provided
	if pv.signatureValidator == nil {
		pv.signatureValidator = NewSignatureValidator(
			registry,
			WithStrictMode(pv.config.RequireClientSignature && pv.config.RequireUserSignature),
		)
	}

	// Initialize replay protector if not provided
	if pv.replayProtector == nil {
		pv.replayProtector = NewReplayProtector(
			WithReplayProtectorWindow(pv.config.ReplayWindow),
		)
	}

	return pv
}

// ValidatePayload performs complete validation of a capture payload.
// Returns a ValidationResult with any errors.
func (pv *ProtocolValidator) ValidatePayload(payload CapturePayload, expectedAccount string) ValidationResult {
	result := ValidationResult{
		Valid:       true,
		ValidatedAt: pv.now(),
	}

	// 1. Validate protocol version
	if err := pv.validateVersion(payload); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, errorToValidationError(err))
		return result
	}

	// 2. Validate salt and binding
	if err := pv.validateSalt(payload); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, errorToValidationError(err))
		return result
	}

	// 3. Check for replay attack
	if err := pv.checkReplay(payload); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, errorToValidationError(err))
		return result
	}

	// 4. Validate client signature
	if pv.config.RequireClientSignature {
		if err := pv.validateClientSignature(payload); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, errorToValidationError(err))
			return result
		}
		result.ClientID = payload.ClientSignature.KeyID
	}

	// 5. Validate user signature
	if pv.config.RequireUserSignature {
		if err := pv.validateUserSignature(payload, expectedAccount); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, errorToValidationError(err))
			return result
		}
		result.UserAddress = payload.UserSignature.KeyID
	}

	// 6. Verify signature chain integrity
	if err := pv.verifySignatureChain(payload); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, errorToValidationError(err))
		return result
	}

	// 7. Validate metadata
	if err := pv.validateMetadata(payload); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, errorToValidationError(err))
		return result
	}

	// 8. Record salt as used (prevent future replay)
	if err := pv.recordSalt(payload); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, errorToValidationError(err))
		return result
	}

	return result
}

// ValidatePayloadQuick performs a quick validation without recording the salt.
// Use this for pre-validation before committing.
func (pv *ProtocolValidator) ValidatePayloadQuick(payload CapturePayload, expectedAccount string) error {
	// 1. Validate protocol version
	if err := pv.validateVersion(payload); err != nil {
		return err
	}

	// 2. Validate salt (without recording)
	if err := pv.saltValidator.ValidateSalt(payload.SaltBinding); err != nil {
		return err
	}

	// 3. Validate signatures
	if pv.config.RequireClientSignature || pv.config.RequireUserSignature {
		if err := pv.signatureValidator.ValidateBothSignatures(payload, expectedAccount); err != nil {
			return err
		}
	}

	return nil
}

// validateVersion checks the protocol version
func (pv *ProtocolValidator) validateVersion(payload CapturePayload) error {
	if payload.Version == 0 {
		return ErrVersionMissing
	}
	if payload.Version > ProtocolVersion {
		return ErrVersionUnsupported.WithDetails(
			"supported", ProtocolVersion,
			"received", payload.Version,
		)
	}
	return nil
}

// validateSalt validates salt and binding
func (pv *ProtocolValidator) validateSalt(payload CapturePayload) error {
	// Validate salt bytes match binding
	if !constantTimeEqual(payload.Salt, payload.SaltBinding.Salt) {
		return ErrSaltMismatch
	}

	// Full salt validation
	return pv.saltValidator.ValidateSalt(payload.SaltBinding)
}

// checkReplay checks for replay attacks
func (pv *ProtocolValidator) checkReplay(payload CapturePayload) error {
	return pv.replayProtector.CheckNotReplayed(payload)
}

// validateClientSignature validates the client signature
func (pv *ProtocolValidator) validateClientSignature(payload CapturePayload) error {
	return pv.signatureValidator.ValidateClientSignature(payload)
}

// validateUserSignature validates the user signature
func (pv *ProtocolValidator) validateUserSignature(payload CapturePayload, expectedAccount string) error {
	return pv.signatureValidator.ValidateUserSignature(payload, expectedAccount)
}

// verifySignatureChain verifies the signature chain is intact
func (pv *ProtocolValidator) verifySignatureChain(payload CapturePayload) error {
	return pv.signatureValidator.VerifySignatureChain(payload)
}

// validateMetadata validates capture metadata
func (pv *ProtocolValidator) validateMetadata(payload CapturePayload) error {
	meta := payload.CaptureMetadata

	if meta.ClientID == "" {
		return ErrMetadataClientIDMissing
	}

	if meta.DeviceFingerprint == "" {
		return ErrMetadataDeviceFingerprintMissing
	}

	if meta.SessionID == "" {
		return ErrMetadataSessionIDMissing
	}

	// Verify metadata client ID matches signature client ID
	if meta.ClientID != payload.ClientSignature.KeyID {
		return ErrMetadataClientIDMismatch.WithDetails(
			"metadata_client_id", meta.ClientID,
			"signature_client_id", payload.ClientSignature.KeyID,
		)
	}

	// Verify metadata session ID matches binding session ID
	if meta.SessionID != payload.SaltBinding.SessionID {
		return ErrMetadataSessionIDMismatch.WithDetails(
			"metadata_session_id", meta.SessionID,
			"binding_session_id", payload.SaltBinding.SessionID,
		)
	}

	return nil
}

// recordSalt records the salt as used
func (pv *ProtocolValidator) recordSalt(payload CapturePayload) error {
	if err := pv.saltValidator.RecordUsedSalt(payload.Salt); err != nil {
		return err
	}
	return pv.replayProtector.RecordPayload(payload)
}

// ReplayProtector provides anti-replay protection for the capture protocol.
type ReplayProtector struct {
	// Salt cache tracks used salts
	saltCache *saltCache

	// Nonce cache tracks used nonces (if applicable)
	nonceCache *saltCache

	// Maximum replay window
	maxWindowSize time.Duration

	// Time source
	now func() time.Time
}

// ReplayProtectorOption is a functional option for ReplayProtector
type ReplayProtectorOption func(*ReplayProtector)

// WithReplayProtectorWindow sets the replay protection window
func WithReplayProtectorWindow(window time.Duration) ReplayProtectorOption {
	return func(rp *ReplayProtector) {
		rp.maxWindowSize = window
	}
}

// NewReplayProtector creates a new ReplayProtector
func NewReplayProtector(opts ...ReplayProtectorOption) *ReplayProtector {
	rp := &ReplayProtector{
		maxWindowSize: DefaultReplayWindow,
		now:           time.Now,
	}

	for _, opt := range opts {
		opt(rp)
	}

	rp.saltCache = newSaltCache(100000, rp.maxWindowSize)
	rp.nonceCache = newSaltCache(100000, rp.maxWindowSize)

	return rp
}

// CheckNotReplayed checks if a payload is a replay attack
func (rp *ReplayProtector) CheckNotReplayed(payload CapturePayload) error {
	// Check salt not used before
	if err := rp.checkSaltNotUsed(payload.Salt); err != nil {
		return err
	}

	// Check timestamp within window
	if err := rp.checkTimestampWindow(payload); err != nil {
		return err
	}

	return nil
}

// checkSaltNotUsed verifies the salt hasn't been used before
func (rp *ReplayProtector) checkSaltNotUsed(salt []byte) error {
	hash := rp.computeHash(salt)
	if rp.saltCache.exists(hash) {
		return ErrSaltReplayed
	}
	return nil
}

// checkTimestampWindow verifies the payload timestamp is within the acceptable window
func (rp *ReplayProtector) checkTimestampWindow(payload CapturePayload) error {
	now := rp.now()
	payloadTime := payload.Timestamp

	age := now.Sub(payloadTime)

	// Check if too old
	if age > rp.maxWindowSize {
		return ErrPayloadExpired.WithDetails(
			"max_age", rp.maxWindowSize.String(),
			"actual_age", age.String(),
		)
	}

	// Check if from future (with small tolerance)
	if age < -DefaultMaxClockSkew {
		return ErrPayloadFromFuture.WithDetails(
			"max_skew", DefaultMaxClockSkew.String(),
			"time_ahead", (-age).String(),
		)
	}

	return nil
}

// RecordPayload records a payload to prevent future replay
func (rp *ReplayProtector) RecordPayload(payload CapturePayload) error {
	// Record salt
	saltHash := rp.computeHash(payload.Salt)
	rp.saltCache.add(saltHash)

	return nil
}

// computeHash computes a hash for cache lookup
func (rp *ReplayProtector) computeHash(data []byte) [32]byte {
	var hash [32]byte
	h := make([]byte, 32)
	copy(h, data)
	if len(data) >= 32 {
		copy(hash[:], data[:32])
	} else {
		copy(hash[:], h)
	}
	return hash
}

// Cleanup removes expired entries from all caches
func (rp *ReplayProtector) Cleanup() int {
	saltRemoved := rp.saltCache.cleanup()
	nonceRemoved := rp.nonceCache.cleanup()
	return saltRemoved + nonceRemoved
}

// errorToValidationError converts an error to a ValidationError
func errorToValidationError(err error) ValidationError {
	if protocolErr, ok := err.(*ProtocolError); ok {
		return ValidationError{
			Code:    protocolErr.Code,
			Message: protocolErr.Message,
			Field:   protocolErr.Field,
		}
	}
	return ValidationError{
		Code:    "UNKNOWN_ERROR",
		Message: err.Error(),
	}
}

// CreateCapturePayload creates a new CapturePayload with all required fields
func CreateCapturePayload(
	payloadHash []byte,
	salt []byte,
	deviceID string,
	sessionID string,
	clientSignature SignatureProof,
	userSignature SignatureProof,
	metadata CaptureMetadata,
) CapturePayload {
	now := time.Now()

	return CapturePayload{
		Version:         ProtocolVersion,
		PayloadHash:     payloadHash,
		Salt:            salt,
		SaltBinding:     CreateSaltBinding(salt, deviceID, sessionID, now.Unix()),
		ClientSignature: clientSignature,
		UserSignature:   userSignature,
		CaptureMetadata: metadata,
		Timestamp:       now,
	}
}
