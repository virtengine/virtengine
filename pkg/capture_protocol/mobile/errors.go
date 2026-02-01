package mobile

import (
	"fmt"
)

// ============================================================================
// Error Definitions for Mobile Capture Protocol
// ============================================================================

// Error codes for mobile capture
const (
	// Camera errors (2001-2099)
	ErrCodeCameraPermissionDenied = "CAMERA_PERMISSION_DENIED"
	ErrCodeCameraNotFound         = "CAMERA_NOT_FOUND"
	ErrCodeCameraNotReadable      = "CAMERA_NOT_READABLE"
	ErrCodeCameraOverconstrained  = "CAMERA_OVERCONSTRAINED"
	ErrCodeCameraSecurityError    = "CAMERA_SECURITY_ERROR"
	ErrCodeCameraSessionFailed    = "CAMERA_SESSION_FAILED"

	// Capture errors (2101-2199)
	ErrCodeCaptureTimeout          = "CAPTURE_TIMEOUT"
	ErrCodeCaptureCancelled        = "CAPTURE_CANCELLED"
	ErrCodeCaptureQualityFailed    = "CAPTURE_QUALITY_FAILED"
	ErrCodeCaptureProcessingFailed = "CAPTURE_PROCESSING_FAILED"
	ErrCodeCaptureNoDocument       = "CAPTURE_NO_DOCUMENT"
	ErrCodeCaptureNoFace           = "CAPTURE_NO_FACE"

	// Gallery prevention errors (2201-2299)
	ErrCodeGalleryUploadBlocked    = "GALLERY_UPLOAD_BLOCKED"
	ErrCodeOriginProofMissing      = "ORIGIN_PROOF_MISSING"
	ErrCodeOriginProofInvalid      = "ORIGIN_PROOF_INVALID"
	ErrCodeCaptureTimestampInvalid = "CAPTURE_TIMESTAMP_INVALID"
	ErrCodeCaptureSourceInvalid    = "CAPTURE_SOURCE_INVALID"

	// Liveness errors (2301-2399)
	ErrCodeLivenessTimeout         = "LIVENESS_TIMEOUT"
	ErrCodeLivenessChallengeFailed = "LIVENESS_CHALLENGE_FAILED"
	ErrCodeLivenessNotVerified     = "LIVENESS_NOT_VERIFIED"
	ErrCodeSpoofingDetected        = "SPOOFING_DETECTED"

	// Signing errors (2401-2499)
	ErrCodeClientSigningFailed = "CLIENT_SIGNING_FAILED"
	ErrCodeUserSigningFailed   = "USER_SIGNING_FAILED"
	ErrCodeKeyNotAvailable     = "KEY_NOT_AVAILABLE"
	ErrCodeHardwareKeyRequired = "HARDWARE_KEY_REQUIRED"
	ErrCodeBiometricRequired   = "BIOMETRIC_REQUIRED"

	// Device errors (2501-2599)
	ErrCodeDeviceFingerprintInvalid = "DEVICE_FINGERPRINT_INVALID"
	ErrCodePlatformNotSupported     = "PLATFORM_NOT_SUPPORTED"
	ErrCodeOSVersionTooOld          = "OS_VERSION_TOO_OLD"

	// Quality errors (2601-2699)
	ErrCodeQualityResolutionLow = "QUALITY_RESOLUTION_LOW"
	ErrCodeQualityTooDark       = "QUALITY_TOO_DARK"
	ErrCodeQualityTooBright     = "QUALITY_TOO_BRIGHT"
	ErrCodeQualityTooBlurry     = "QUALITY_TOO_BLURRY"
	ErrCodeQualityTooSkewed     = "QUALITY_TOO_SKEWED"
	ErrCodeQualityGlareDetected = "QUALITY_GLARE_DETECTED"

	// Contract validation errors (2701-2799)
	ErrCodeCaptureResolutionTooLow  = "CAPTURE_RESOLUTION_TOO_LOW"
	ErrCodeCaptureResolutionTooHigh = "CAPTURE_RESOLUTION_TOO_HIGH"
	ErrCodeCaptureFileTooLarge      = "CAPTURE_FILE_TOO_LARGE"
	ErrCodeCaptureFormatInvalid     = "CAPTURE_FORMAT_INVALID"
	ErrCodeCaptureCodecInvalid      = "CAPTURE_CODEC_INVALID"
	ErrCodeCaptureFPSTooLow         = "CAPTURE_FPS_TOO_LOW"
	ErrCodeCaptureDurationInvalid   = "CAPTURE_DURATION_INVALID"
)

// MobileError represents a mobile capture error
type MobileError struct {
	Code    string
	Message string
	Field   string
	Details map[string]interface{}
	Wrapped error
}

// Error implements the error interface
func (e *MobileError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Wrapped)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *MobileError) Unwrap() error {
	return e.Wrapped
}

// Wrap wraps another error
func (e *MobileError) Wrap(err error) *MobileError {
	return &MobileError{
		Code:    e.Code,
		Message: e.Message,
		Field:   e.Field,
		Details: e.Details,
		Wrapped: err,
	}
}

// WithDetails adds details to the error
func (e *MobileError) WithDetails(keyvals ...interface{}) *MobileError {
	details := make(map[string]interface{})
	for k, v := range e.Details {
		details[k] = v
	}
	for i := 0; i < len(keyvals)-1; i += 2 {
		if key, ok := keyvals[i].(string); ok {
			details[key] = keyvals[i+1]
		}
	}
	return &MobileError{
		Code:    e.Code,
		Message: e.Message,
		Field:   e.Field,
		Details: details,
		Wrapped: e.Wrapped,
	}
}

// WithField sets the field name
func (e *MobileError) WithField(field string) *MobileError {
	return &MobileError{
		Code:    e.Code,
		Message: e.Message,
		Field:   field,
		Details: e.Details,
		Wrapped: e.Wrapped,
	}
}

// ============================================================================
// Pre-defined Errors
// ============================================================================

// Camera errors
var (
	ErrCameraPermissionDenied = &MobileError{
		Code:    ErrCodeCameraPermissionDenied,
		Message: "Camera permission was denied",
	}
	ErrCameraNotFound = &MobileError{
		Code:    ErrCodeCameraNotFound,
		Message: "No camera found on device",
	}
	ErrCameraNotReadable = &MobileError{
		Code:    ErrCodeCameraNotReadable,
		Message: "Camera is not readable",
	}
	ErrCameraSessionFailed = &MobileError{
		Code:    ErrCodeCameraSessionFailed,
		Message: "Failed to create camera session",
	}
)

// Capture errors
var (
	ErrCaptureTimeout = &MobileError{
		Code:    ErrCodeCaptureTimeout,
		Message: "Capture timed out",
	}
	ErrCaptureCancelled = &MobileError{
		Code:    ErrCodeCaptureCancelled,
		Message: "Capture was cancelled",
	}
	ErrCaptureQualityFailed = &MobileError{
		Code:    ErrCodeCaptureQualityFailed,
		Message: "Image quality check failed",
	}
	ErrCaptureNoDocument = &MobileError{
		Code:    ErrCodeCaptureNoDocument,
		Message: "No document detected in frame",
	}
	ErrCaptureNoFace = &MobileError{
		Code:    ErrCodeCaptureNoFace,
		Message: "No face detected in frame",
	}
)

// Gallery prevention errors
var (
	ErrGalleryUploadBlocked = &MobileError{
		Code:    ErrCodeGalleryUploadBlocked,
		Message: "Gallery uploads are not allowed",
	}
	ErrOriginProofMissing = &MobileError{
		Code:    ErrCodeOriginProofMissing,
		Message: "Capture origin proof is missing",
	}
	ErrOriginProofInvalid = &MobileError{
		Code:    ErrCodeOriginProofInvalid,
		Message: "Capture origin proof is invalid",
	}
	ErrCaptureTimestampInvalid = &MobileError{
		Code:    ErrCodeCaptureTimestampInvalid,
		Message: "Capture timestamp is invalid or too old",
	}
	ErrCaptureSourceInvalid = &MobileError{
		Code:    ErrCodeCaptureSourceInvalid,
		Message: "Capture source could not be verified",
	}
)

// Liveness errors
var (
	ErrLivenessTimeout = &MobileError{
		Code:    ErrCodeLivenessTimeout,
		Message: "Liveness check timed out",
	}
	ErrLivenessChallengeFailed = &MobileError{
		Code:    ErrCodeLivenessChallengeFailed,
		Message: "Liveness challenge was not completed",
	}
	ErrLivenessNotVerified = &MobileError{
		Code:    ErrCodeLivenessNotVerified,
		Message: "Liveness could not be verified",
	}
	ErrSpoofingDetected = &MobileError{
		Code:    ErrCodeSpoofingDetected,
		Message: "Spoofing attempt detected",
	}
)

// Signing errors
var (
	ErrClientSigningFailed = &MobileError{
		Code:    ErrCodeClientSigningFailed,
		Message: "Failed to sign with client key",
	}
	ErrUserSigningFailed = &MobileError{
		Code:    ErrCodeUserSigningFailed,
		Message: "Failed to sign with user key",
	}
	ErrKeyNotAvailable = &MobileError{
		Code:    ErrCodeKeyNotAvailable,
		Message: "Signing key is not available",
	}
	ErrHardwareKeyRequired = &MobileError{
		Code:    ErrCodeHardwareKeyRequired,
		Message: "Hardware-backed key is required",
	}
	ErrBiometricRequired = &MobileError{
		Code:    ErrCodeBiometricRequired,
		Message: "Biometric authentication is required",
	}
)

// Device errors
var (
	ErrDeviceFingerprintInvalid = &MobileError{
		Code:    ErrCodeDeviceFingerprintInvalid,
		Message: "Device fingerprint is invalid",
	}
	ErrPlatformNotSupported = &MobileError{
		Code:    ErrCodePlatformNotSupported,
		Message: "Platform is not supported",
	}
	ErrOSVersionTooOld = &MobileError{
		Code:    ErrCodeOSVersionTooOld,
		Message: "Operating system version is too old",
	}
)

// Quality errors
var (
	ErrQualityResolutionLow = &MobileError{
		Code:    ErrCodeQualityResolutionLow,
		Message: "Image resolution is too low",
	}
	ErrQualityTooDark = &MobileError{
		Code:    ErrCodeQualityTooDark,
		Message: "Image is too dark",
	}
	ErrQualityTooBright = &MobileError{
		Code:    ErrCodeQualityTooBright,
		Message: "Image is too bright",
	}
	ErrQualityTooBlurry = &MobileError{
		Code:    ErrCodeQualityTooBlurry,
		Message: "Image is too blurry",
	}
	ErrQualityTooSkewed = &MobileError{
		Code:    ErrCodeQualityTooSkewed,
		Message: "Document is too skewed",
	}
	ErrQualityGlareDetected = &MobileError{
		Code:    ErrCodeQualityGlareDetected,
		Message: "Glare detected on image",
	}
)

// Contract validation errors
var (
	ErrCaptureResolutionTooLow = &MobileError{
		Code:    ErrCodeCaptureResolutionTooLow,
		Message: "Capture resolution is below minimum requirement",
	}
	ErrCaptureResolutionTooHigh = &MobileError{
		Code:    ErrCodeCaptureResolutionTooHigh,
		Message: "Capture resolution exceeds maximum allowed",
	}
	ErrCaptureFileTooLarge = &MobileError{
		Code:    ErrCodeCaptureFileTooLarge,
		Message: "Capture file size exceeds maximum allowed",
	}
	ErrCaptureFormatInvalid = &MobileError{
		Code:    ErrCodeCaptureFormatInvalid,
		Message: "Capture format is not supported",
	}
	ErrCaptureCodecInvalid = &MobileError{
		Code:    ErrCodeCaptureCodecInvalid,
		Message: "Video codec is not supported",
	}
	ErrCaptureFPSTooLow = &MobileError{
		Code:    ErrCodeCaptureFPSTooLow,
		Message: "Video frame rate is below minimum requirement",
	}
	ErrCaptureDurationInvalid = &MobileError{
		Code:    ErrCodeCaptureDurationInvalid,
		Message: "Video duration is outside allowed range",
	}
	ErrCaptureFrameRateTooLow = &MobileError{
		Code:    ErrCodeCaptureFPSTooLow,
		Message: "Video frame rate is below minimum requirement",
	}
	ErrCaptureDurationTooShort = &MobileError{
		Code:    ErrCodeCaptureDurationInvalid,
		Message: "Video duration is too short",
	}
	ErrCaptureDurationTooLong = &MobileError{
		Code:    ErrCodeCaptureDurationInvalid,
		Message: "Video duration is too long",
	}
)

// ============================================================================
// Error Type Checking
// ============================================================================

// IsCameraError returns true if error is camera-related
func IsCameraError(err error) bool {
	if e, ok := err.(*MobileError); ok {
		switch e.Code {
		case ErrCodeCameraPermissionDenied,
			ErrCodeCameraNotFound,
			ErrCodeCameraNotReadable,
			ErrCodeCameraOverconstrained,
			ErrCodeCameraSecurityError,
			ErrCodeCameraSessionFailed:
			return true
		}
	}
	return false
}

// IsGalleryError returns true if error is gallery-related
func IsGalleryError(err error) bool {
	if e, ok := err.(*MobileError); ok {
		switch e.Code {
		case ErrCodeGalleryUploadBlocked,
			ErrCodeOriginProofMissing,
			ErrCodeOriginProofInvalid,
			ErrCodeCaptureTimestampInvalid,
			ErrCodeCaptureSourceInvalid:
			return true
		}
	}
	return false
}

// IsLivenessError returns true if error is liveness-related
func IsLivenessError(err error) bool {
	if e, ok := err.(*MobileError); ok {
		switch e.Code {
		case ErrCodeLivenessTimeout,
			ErrCodeLivenessChallengeFailed,
			ErrCodeLivenessNotVerified,
			ErrCodeSpoofingDetected:
			return true
		}
	}
	return false
}

// IsQualityError returns true if error is quality-related
func IsQualityError(err error) bool {
	if e, ok := err.(*MobileError); ok {
		switch e.Code {
		case ErrCodeQualityResolutionLow,
			ErrCodeQualityTooDark,
			ErrCodeQualityTooBright,
			ErrCodeQualityTooBlurry,
			ErrCodeQualityTooSkewed,
			ErrCodeQualityGlareDetected:
			return true
		}
	}
	return false
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	if e, ok := err.(*MobileError); ok {
		switch e.Code {
		case ErrCodeCaptureQualityFailed,
			ErrCodeCaptureNoDocument,
			ErrCodeCaptureNoFace,
			ErrCodeLivenessChallengeFailed,
			ErrCodeQualityResolutionLow,
			ErrCodeQualityTooDark,
			ErrCodeQualityTooBright,
			ErrCodeQualityTooBlurry,
			ErrCodeQualityTooSkewed,
			ErrCodeQualityGlareDetected:
			return true
		}
	}
	return false
}

