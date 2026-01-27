// Package mobile implements native mobile capture specifications for iOS and Android.
// VE-900: Mobile capture app - native camera integration
//
// This package defines types, interfaces, and protocol specifications for mobile
// capture libraries that integrate with the VirtEngine identity verification system.
//
// Features:
//   - Native camera capture types for iOS (AVFoundation) and Android (Camera2)
//   - Real-time quality feedback interfaces
//   - Liveness detection interface during facial capture
//   - Gallery upload prevention mechanism (metadata validation)
//   - Approved-client key signing integration
//
// See _docs/protocols/mobile-capture-protocol-v1.md for full specification.
package mobile

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Version constants
const (
	// MobileProtocolVersion is the current mobile protocol version
	MobileProtocolVersion uint32 = 1

	// MinSupportedIOSVersion is the minimum iOS version
	MinSupportedIOSVersion = "14.0"

	// MinSupportedAndroidAPI is the minimum Android API level
	MinSupportedAndroidAPI = 26 // Android 8.0 (Oreo)
)

// ============================================================================
// Platform Types
// ============================================================================

// Platform represents the mobile platform
type Platform string

const (
	// PlatformIOS represents iOS/iPadOS
	PlatformIOS Platform = "ios"

	// PlatformAndroid represents Android
	PlatformAndroid Platform = "android"
)

// PlatformCapabilities defines platform-specific capabilities
type PlatformCapabilities struct {
	// Platform type
	Platform Platform `json:"platform"`

	// OSVersion is the operating system version
	OSVersion string `json:"os_version"`

	// APILevel is the Android API level (0 for iOS)
	APILevel int `json:"api_level,omitempty"`

	// DeviceModel is the device model identifier
	DeviceModel string `json:"device_model"`

	// SupportsTrustedExecution indicates TEE/Secure Enclave support
	SupportsTrustedExecution bool `json:"supports_trusted_execution"`

	// SupportsHardwareKeystore indicates hardware-backed keystore
	SupportsHardwareKeystore bool `json:"supports_hardware_keystore"`

	// SupportsBiometrics indicates biometric authentication support
	SupportsBiometrics bool `json:"supports_biometrics"`

	// BiometricTypes lists available biometric types
	BiometricTypes []BiometricType `json:"biometric_types,omitempty"`

	// CameraCapabilities describes camera features
	CameraCapabilities CameraCapabilities `json:"camera_capabilities"`
}

// BiometricType represents a type of biometric authentication
type BiometricType string

const (
	// BiometricFaceID is Apple Face ID
	BiometricFaceID BiometricType = "face_id"

	// BiometricTouchID is Apple Touch ID
	BiometricTouchID BiometricType = "touch_id"

	// BiometricFingerprint is Android fingerprint
	BiometricFingerprint BiometricType = "fingerprint"

	// BiometricFace is Android face unlock
	BiometricFace BiometricType = "face"

	// BiometricIris is iris scanning
	BiometricIris BiometricType = "iris"
)

// ============================================================================
// Camera Types
// ============================================================================

// CameraCapabilities describes device camera features
type CameraCapabilities struct {
	// HasBackCamera indicates back camera availability
	HasBackCamera bool `json:"has_back_camera"`

	// HasFrontCamera indicates front camera availability
	HasFrontCamera bool `json:"has_front_camera"`

	// MaxBackResolution is the maximum back camera resolution
	MaxBackResolution Resolution `json:"max_back_resolution"`

	// MaxFrontResolution is the maximum front camera resolution
	MaxFrontResolution Resolution `json:"max_front_resolution"`

	// SupportsAutoFocus indicates auto-focus support
	SupportsAutoFocus bool `json:"supports_auto_focus"`

	// SupportsFlash indicates flash support
	SupportsFlash bool `json:"supports_flash"`

	// SupportsOpticalZoom indicates optical zoom support
	SupportsOpticalZoom bool `json:"supports_optical_zoom"`

	// SupportsDepthCapture indicates depth sensing support (for liveness)
	SupportsDepthCapture bool `json:"supports_depth_capture"`

	// SupportsRAWCapture indicates RAW image capture support
	SupportsRAWCapture bool `json:"supports_raw_capture"`
}

// Resolution represents image resolution
type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Megapixels returns the resolution in megapixels
func (r Resolution) Megapixels() float64 {
	return float64(r.Width*r.Height) / 1_000_000.0
}

// CameraPosition represents camera facing direction
type CameraPosition string

const (
	// CameraPositionBack is the rear/environment camera
	CameraPositionBack CameraPosition = "back"

	// CameraPositionFront is the front/selfie camera
	CameraPositionFront CameraPosition = "front"
)

// CameraConfiguration defines camera capture settings
type CameraConfiguration struct {
	// Position is the camera to use
	Position CameraPosition `json:"position"`

	// TargetResolution is the desired capture resolution
	TargetResolution Resolution `json:"target_resolution"`

	// AutoFocusEnabled enables auto-focus
	AutoFocusEnabled bool `json:"auto_focus_enabled"`

	// FlashMode is the flash configuration
	FlashMode FlashMode `json:"flash_mode"`

	// ExposureMode is the exposure configuration
	ExposureMode ExposureMode `json:"exposure_mode"`

	// WhiteBalanceMode is the white balance configuration
	WhiteBalanceMode WhiteBalanceMode `json:"white_balance_mode"`

	// StabilizationEnabled enables optical/digital stabilization
	StabilizationEnabled bool `json:"stabilization_enabled"`

	// ContinuousCapture enables continuous frame capture for quality preview
	ContinuousCapture bool `json:"continuous_capture"`

	// FrameRate is the target frame rate for preview
	FrameRate int `json:"frame_rate"`
}

// FlashMode represents flash configuration
type FlashMode string

const (
	FlashModeOff  FlashMode = "off"
	FlashModeOn   FlashMode = "on"
	FlashModeAuto FlashMode = "auto"
)

// ExposureMode represents exposure configuration
type ExposureMode string

const (
	ExposureModeAuto       ExposureMode = "auto"
	ExposureModeLocked     ExposureMode = "locked"
	ExposureModeContinuous ExposureMode = "continuous"
)

// WhiteBalanceMode represents white balance configuration
type WhiteBalanceMode string

const (
	WhiteBalanceModeAuto       WhiteBalanceMode = "auto"
	WhiteBalanceModeLocked     WhiteBalanceMode = "locked"
	WhiteBalanceModeContinuous WhiteBalanceMode = "continuous"
)

// DefaultDocumentCameraConfig returns default camera config for document capture
func DefaultDocumentCameraConfig() CameraConfiguration {
	return CameraConfiguration{
		Position:             CameraPositionBack,
		TargetResolution:     Resolution{Width: 3840, Height: 2160}, // 4K
		AutoFocusEnabled:     true,
		FlashMode:            FlashModeAuto,
		ExposureMode:         ExposureModeAuto,
		WhiteBalanceMode:     WhiteBalanceModeAuto,
		StabilizationEnabled: true,
		ContinuousCapture:    true,
		FrameRate:            30,
	}
}

// DefaultSelfieCameraConfig returns default camera config for selfie capture
func DefaultSelfieCameraConfig() CameraConfiguration {
	return CameraConfiguration{
		Position:             CameraPositionFront,
		TargetResolution:     Resolution{Width: 1920, Height: 1080}, // 1080p
		AutoFocusEnabled:     true,
		FlashMode:            FlashModeOff, // No flash for selfie
		ExposureMode:         ExposureModeContinuous,
		WhiteBalanceMode:     WhiteBalanceModeContinuous,
		StabilizationEnabled: true,
		ContinuousCapture:    true,
		FrameRate:            30,
	}
}

// ============================================================================
// Capture Session Types
// ============================================================================

// CaptureSessionConfig defines capture session configuration
type CaptureSessionConfig struct {
	// SessionID is a unique session identifier
	SessionID string `json:"session_id"`

	// CaptureType is the type of capture to perform
	CaptureType CaptureType `json:"capture_type"`

	// DocumentType is required for document captures
	DocumentType DocumentType `json:"document_type,omitempty"`

	// DocumentSide is required for document captures
	DocumentSide DocumentSide `json:"document_side,omitempty"`

	// CameraConfig is the camera configuration
	CameraConfig CameraConfiguration `json:"camera_config"`

	// QualityConfig is the quality validation configuration
	QualityConfig QualityConfiguration `json:"quality_config"`

	// LivenessConfig is the liveness detection configuration (for selfie)
	LivenessConfig *LivenessConfiguration `json:"liveness_config,omitempty"`

	// TimeoutSeconds is the capture timeout
	TimeoutSeconds int `json:"timeout_seconds"`

	// MaxRetries is the maximum number of capture retries
	MaxRetries int `json:"max_retries"`
}

// CaptureType represents the type of capture
type CaptureType string

const (
	// CaptureTypeDocument is document capture
	CaptureTypeDocument CaptureType = "document"

	// CaptureTypeSelfie is selfie/facial capture
	CaptureTypeSelfie CaptureType = "selfie"
)

// DocumentType represents supported document types
type DocumentType string

const (
	DocumentTypeIDCard         DocumentType = "id_card"
	DocumentTypePassport       DocumentType = "passport"
	DocumentTypeDriversLicense DocumentType = "drivers_license"
)

// DocumentSide represents document side
type DocumentSide string

const (
	DocumentSideFront DocumentSide = "front"
	DocumentSideBack  DocumentSide = "back"
)

// CaptureSessionState represents session state
type CaptureSessionState string

const (
	SessionStateInitializing CaptureSessionState = "initializing"
	SessionStateReady        CaptureSessionState = "ready"
	SessionStateCapturing    CaptureSessionState = "capturing"
	SessionStateProcessing   CaptureSessionState = "processing"
	SessionStateComplete     CaptureSessionState = "complete"
	SessionStateFailed       CaptureSessionState = "failed"
	SessionStateCancelled    CaptureSessionState = "cancelled"
)

// CaptureSession represents an active capture session
type CaptureSession struct {
	// Config is the session configuration
	Config CaptureSessionConfig `json:"config"`

	// State is the current session state
	State CaptureSessionState `json:"state"`

	// CreatedAt is when the session was created
	CreatedAt time.Time `json:"created_at"`

	// StartedAt is when capture started
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is when capture completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// AttemptCount is the number of capture attempts
	AttemptCount int `json:"attempt_count"`

	// LastQualityResult is the last quality check result
	LastQualityResult *QualityResult `json:"last_quality_result,omitempty"`

	// LastLivenessResult is the last liveness check result
	LastLivenessResult *LivenessResult `json:"last_liveness_result,omitempty"`
}

// ============================================================================
// Device Fingerprint Types
// ============================================================================

// DeviceFingerprint represents device identification data
type DeviceFingerprint struct {
	// FingerprintHash is the SHA256 hash of fingerprint components
	FingerprintHash string `json:"fingerprint_hash"`

	// Platform is the device platform
	Platform Platform `json:"platform"`

	// DeviceID is a platform-specific device identifier
	// iOS: identifierForVendor
	// Android: ANDROID_ID (scoped to app)
	DeviceID string `json:"device_id"`

	// AppInstanceID is the app installation instance ID
	AppInstanceID string `json:"app_instance_id"`

	// HardwareID is hardware-derived identifier (optional)
	HardwareID string `json:"hardware_id,omitempty"`

	// CreatedAt is when the fingerprint was generated
	CreatedAt time.Time `json:"created_at"`
}

// ComputeFingerprintHash computes the fingerprint hash
func (df *DeviceFingerprint) ComputeFingerprintHash() string {
	h := sha256.New()
	h.Write([]byte(df.Platform))
	h.Write([]byte(df.DeviceID))
	h.Write([]byte(df.AppInstanceID))
	h.Write([]byte(df.HardwareID))
	return hex.EncodeToString(h.Sum(nil))
}

// Validate validates the fingerprint data
func (df *DeviceFingerprint) Validate() error {
	if df.Platform == "" {
		return ErrDeviceFingerprintInvalid.WithDetails("field", "platform")
	}
	if df.DeviceID == "" {
		return ErrDeviceFingerprintInvalid.WithDetails("field", "device_id")
	}
	if df.AppInstanceID == "" {
		return ErrDeviceFingerprintInvalid.WithDetails("field", "app_instance_id")
	}
	return nil
}

// ============================================================================
// Capture Result Types
// ============================================================================

// CaptureResult represents a successful capture result
type CaptureResult struct {
	// SessionID is the capture session ID
	SessionID string `json:"session_id"`

	// CaptureType is the type of capture
	CaptureType CaptureType `json:"capture_type"`

	// ImageData is the captured image (JPEG, stripped of EXIF/GPS)
	ImageData []byte `json:"image_data"`

	// ImageHash is SHA256 of the image data
	ImageHash []byte `json:"image_hash"`

	// Resolution is the captured image resolution
	Resolution Resolution `json:"resolution"`

	// MimeType is the image MIME type
	MimeType string `json:"mime_type"`

	// CapturedAt is when the image was captured
	CapturedAt time.Time `json:"captured_at"`

	// QualityResult is the quality validation result
	QualityResult QualityResult `json:"quality_result"`

	// LivenessResult is the liveness check result (for selfie)
	LivenessResult *LivenessResult `json:"liveness_result,omitempty"`

	// DeviceFingerprint is the device fingerprint
	DeviceFingerprint DeviceFingerprint `json:"device_fingerprint"`

	// CaptureSource confirms the image source
	CaptureSource CaptureSource `json:"capture_source"`

	// GalleryUploadBlocked indicates if gallery was blocked
	GalleryUploadBlocked bool `json:"gallery_upload_blocked"`
}

// CaptureSource represents the verified source of a capture
type CaptureSource struct {
	// SourceType is the capture source type
	SourceType CaptureSourceType `json:"source_type"`

	// Verified indicates source was verified via platform APIs
	Verified bool `json:"verified"`

	// VerificationMethod is how source was verified
	VerificationMethod string `json:"verification_method"`

	// Timestamp is when source verification occurred
	Timestamp time.Time `json:"timestamp"`
}

// CaptureSourceType represents capture source types
type CaptureSourceType string

const (
	// CaptureSourceLiveCamera indicates live camera capture
	CaptureSourceLiveCamera CaptureSourceType = "live_camera"

	// CaptureSourceGallery indicates gallery upload (blocked)
	CaptureSourceGallery CaptureSourceType = "gallery"

	// CaptureSourceUnknown indicates unknown source (rejected)
	CaptureSourceUnknown CaptureSourceType = "unknown"
)
