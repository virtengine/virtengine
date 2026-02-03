// Package mobile implements native mobile capture specifications for iOS and Android.
// VE-900: Mobile capture app - native camera integration
//
// This file defines the core interfaces that mobile SDKs must implement.
package mobile

import (
	"context"
	"time"
)

// ============================================================================
// Core Mobile Capture Library Interface
// ============================================================================

// MobileCaptureLibrary defines the interface for mobile capture SDKs
// This interface must be implemented by both iOS (Swift) and Android (Kotlin) SDKs
type MobileCaptureLibrary interface {
	// Initialize initializes the capture library
	Initialize(config MobileLibraryConfig) error

	// GetPlatformCapabilities returns device capabilities
	GetPlatformCapabilities() (*PlatformCapabilities, error)

	// CreateCaptureSession creates a new capture session
	CreateCaptureSession(config CaptureSessionConfig) (CaptureSessionInterface, error)

	// GetDeviceFingerprint generates device fingerprint
	GetDeviceFingerprint() (*DeviceFingerprint, error)

	// SetClientKeyProvider sets the approved client key provider
	SetClientKeyProvider(provider MobileClientKeyProvider) error

	// SetUserKeyProvider sets the user key provider
	SetUserKeyProvider(provider MobileUserKeyProvider) error

	// Shutdown gracefully shuts down the library
	Shutdown() error
}

// MobileLibraryConfig configures the mobile capture library
type MobileLibraryConfig struct {
	// ClientID is the approved client identifier
	ClientID string `json:"client_id"`

	// ClientVersion is the app version
	ClientVersion string `json:"client_version"`

	// Environment is production or staging
	Environment Environment `json:"environment"`

	// EnableDebugLogging enables debug logging (never in production)
	EnableDebugLogging bool `json:"enable_debug_logging"`

	// GalleryPrevention configures gallery upload prevention
	GalleryPrevention GalleryPreventionConfig `json:"gallery_prevention"`

	// DefaultDocumentQuality is default document quality config
	DefaultDocumentQuality QualityConfiguration `json:"default_document_quality"`

	// DefaultSelfieQuality is default selfie quality config
	DefaultSelfieQuality QualityConfiguration `json:"default_selfie_quality"`

	// DefaultLiveness is default liveness config
	DefaultLiveness LivenessConfiguration `json:"default_liveness"`

	// GuidanceMessages are localized guidance messages
	GuidanceMessages GuidanceMessages `json:"guidance_messages"`
}

// Environment represents the runtime environment
type Environment string

const (
	EnvironmentProduction  Environment = "production"
	EnvironmentStaging     Environment = "staging"
	EnvironmentDevelopment Environment = "development"
)

// DefaultMobileLibraryConfig returns default library configuration
func DefaultMobileLibraryConfig(clientID, clientVersion string) MobileLibraryConfig {
	return MobileLibraryConfig{
		ClientID:               clientID,
		ClientVersion:          clientVersion,
		Environment:            EnvironmentProduction,
		EnableDebugLogging:     false,
		GalleryPrevention:      DefaultGalleryPreventionConfig(),
		DefaultDocumentQuality: DefaultDocumentQualityConfig(),
		DefaultSelfieQuality:   DefaultSelfieQualityConfig(),
		DefaultLiveness:        DefaultLivenessConfig(),
		GuidanceMessages:       DefaultEnglishGuidance(),
	}
}

// ============================================================================
// Capture Session Interface
// ============================================================================

// CaptureSessionInterface defines a capture session
type CaptureSessionInterface interface {
	// GetSessionID returns the session ID
	GetSessionID() string

	// GetState returns current session state
	GetState() CaptureSessionState

	// Start starts the capture session
	Start(ctx context.Context) error

	// Capture triggers image capture
	Capture() error

	// Cancel cancels the capture session
	Cancel()

	// GetResult returns the capture result (after completion)
	GetResult() (*CaptureResult, error)

	// GetSignaturePackage returns the signature package (after completion)
	GetSignaturePackage() (*CaptureSignaturePackage, error)

	// SetQualityCallback sets the quality feedback callback
	SetQualityCallback(callback QualityFeedbackCallback)

	// SetLivenessCallback sets the liveness callback (selfie only)
	SetLivenessCallback(callback LivenessCallback)

	// SetEventCallback sets the general event callback
	SetEventCallback(callback CaptureEventCallback)
}

// CaptureEventCallback handles capture session events
type CaptureEventCallback interface {
	// OnSessionStart is called when session starts
	OnSessionStart(session CaptureSessionInterface)

	// OnSessionReady is called when ready for capture
	OnSessionReady(session CaptureSessionInterface)

	// OnCaptureStart is called when capture begins
	OnCaptureStart(session CaptureSessionInterface)

	// OnCaptureComplete is called when capture completes
	OnCaptureComplete(session CaptureSessionInterface, result *CaptureResult)

	// OnSessionComplete is called when session completes
	OnSessionComplete(session CaptureSessionInterface, pkg *CaptureSignaturePackage)

	// OnSessionError is called on error
	OnSessionError(session CaptureSessionInterface, err error)

	// OnSessionCancelled is called when cancelled
	OnSessionCancelled(session CaptureSessionInterface)
}

// ============================================================================
// Camera Control Interface
// ============================================================================

// CameraController defines camera control operations
type CameraController interface {
	// Initialize initializes the camera
	Initialize(config CameraConfiguration) error

	// Start starts the camera preview
	Start() error

	// Stop stops the camera
	Stop() error

	// SwitchCamera switches between front and back camera
	SwitchCamera() error

	// SetFlashMode sets the flash mode
	SetFlashMode(mode FlashMode) error

	// Focus triggers focus at a point
	Focus(x, y float64) error

	// SetZoom sets the zoom level
	SetZoom(level float64) error

	// CapturePhoto captures a photo
	CapturePhoto() ([]byte, *CaptureOriginProof, error)

	// GetCurrentFrame gets the current preview frame
	GetCurrentFrame() ([]byte, error)
}

// ============================================================================
// Image Processing Interface
// ============================================================================

// ImageProcessor defines image processing operations
type ImageProcessor interface {
	// StripMetadata removes EXIF/GPS data from image
	StripMetadata(imageData []byte) ([]byte, error)

	// ValidateQuality performs quality validation
	ValidateQuality(imageData []byte, config QualityConfiguration) (*QualityResult, error)

	// DetectDocument detects document in image
	DetectDocument(imageData []byte) (*DocumentCheck, error)

	// DetectFace detects face in image
	DetectFace(imageData []byte) (*FaceCheck, error)

	// GetImageMetadata extracts image metadata
	GetImageMetadata(imageData []byte) (*CaptureImageMetadata, error)
}

// ============================================================================
// Secure Storage Interface
// ============================================================================

// SecureStorage defines secure storage operations
type SecureStorage interface {
	// StoreKey stores a key securely
	StoreKey(keyID string, keyData []byte) error

	// RetrieveKey retrieves a stored key
	RetrieveKey(keyID string) ([]byte, error)

	// DeleteKey deletes a stored key
	DeleteKey(keyID string) error

	// IsHardwareBacked returns true if storage is hardware-backed
	IsHardwareBacked() bool

	// RequireUserPresence sets biometric requirement for key access
	RequireUserPresence(keyID string, required bool) error
}

// ============================================================================
// Encryption Interface
// ============================================================================

// CaptureEncryption defines encryption operations for captures
type CaptureEncryption interface {
	// EncryptPayload encrypts image data for upload
	EncryptPayload(imageData []byte, recipientKeys [][]byte) ([]byte, error)

	// ComputePayloadHash computes SHA256 hash of encrypted payload
	ComputePayloadHash(encryptedPayload []byte) []byte

	// GenerateSalt generates a cryptographic salt
	GenerateSalt() ([]byte, error)
}

// ============================================================================
// Workflow Interface
// ============================================================================

// CaptureWorkflow defines a complete capture workflow
type CaptureWorkflow interface {
	// StartDocumentCapture starts document capture workflow
	StartDocumentCapture(ctx context.Context, params DocumentCaptureParams) (<-chan CaptureWorkflowEvent, error)

	// StartSelfieCapture starts selfie capture workflow
	StartSelfieCapture(ctx context.Context, params SelfieCaptureParams) (<-chan CaptureWorkflowEvent, error)

	// Cancel cancels the workflow
	Cancel()

	// GetResult returns the final result
	GetResult() (*CaptureWorkflowResult, error)
}

// DocumentCaptureParams contains document capture parameters
type DocumentCaptureParams struct {
	DocumentType DocumentType
	DocumentSide DocumentSide
	SessionID    string
	Timeout      time.Duration
}

// SelfieCaptureParams contains selfie capture parameters
type SelfieCaptureParams struct {
	EnableLiveness bool
	LivenessConfig *LivenessConfiguration
	SessionID      string
	Timeout        time.Duration
}

// CaptureWorkflowEvent represents a workflow event
type CaptureWorkflowEvent struct {
	Type      CaptureWorkflowEventType
	Timestamp time.Time
	Data      interface{}
}

// CaptureWorkflowEventType represents workflow event types
type CaptureWorkflowEventType string

const (
	WorkflowEventCameraReady       CaptureWorkflowEventType = "camera_ready"
	WorkflowEventQualityUpdate     CaptureWorkflowEventType = "quality_update"
	WorkflowEventReadyToCapture    CaptureWorkflowEventType = "ready_to_capture"
	WorkflowEventCaptureStarted    CaptureWorkflowEventType = "capture_started"
	WorkflowEventCaptureComplete   CaptureWorkflowEventType = "capture_complete"
	WorkflowEventLivenessStarted   CaptureWorkflowEventType = "liveness_started"
	WorkflowEventLivenessChallenge CaptureWorkflowEventType = "liveness_challenge"
	WorkflowEventLivenessComplete  CaptureWorkflowEventType = "liveness_complete"
	WorkflowEventSigningStarted    CaptureWorkflowEventType = "signing_started"
	WorkflowEventSigningComplete   CaptureWorkflowEventType = "signing_complete"
	WorkflowEventComplete          CaptureWorkflowEventType = "complete"
	WorkflowEventError             CaptureWorkflowEventType = "error"
	WorkflowEventCancelled         CaptureWorkflowEventType = "cancelled"
)

// CaptureWorkflowResult contains the complete workflow result
type CaptureWorkflowResult struct {
	// Success indicates if workflow completed successfully
	Success bool `json:"success"`

	// CaptureResult contains the capture result
	CaptureResult *CaptureResult `json:"capture_result,omitempty"`

	// SignaturePackage contains the signature package
	SignaturePackage *CaptureSignaturePackage `json:"signature_package,omitempty"`

	// EncryptedPayload is the encrypted image data
	EncryptedPayload []byte `json:"encrypted_payload,omitempty"`

	// Error contains any error
	Error error `json:"error,omitempty"`

	// Duration is the total workflow duration
	Duration time.Duration `json:"duration"`
}

// ============================================================================
// SDK Version Information
// ============================================================================

// SDKInfo contains SDK version information
type SDKInfo struct {
	// SDKVersion is the SDK version
	SDKVersion string `json:"sdk_version"`

	// ProtocolVersion is the supported protocol version
	ProtocolVersion uint32 `json:"protocol_version"`

	// Platform is the SDK platform
	Platform Platform `json:"platform"`

	// MinOSVersion is the minimum supported OS version
	MinOSVersion string `json:"min_os_version"`

	// BuildDate is when the SDK was built
	BuildDate string `json:"build_date"`

	// Features lists supported features
	Features []string `json:"features"`
}

// CurrentSDKInfo returns the current SDK info
func CurrentSDKInfo() SDKInfo {
	return SDKInfo{
		SDKVersion:      "1.0.0",
		ProtocolVersion: MobileProtocolVersion,
		Platform:        "", // Set by platform-specific implementation
		MinOSVersion:    "", // Set by platform-specific implementation
		BuildDate:       "2026-01-24",
		Features: []string{
			"document_capture",
			"selfie_capture",
			"liveness_detection",
			"gallery_prevention",
			"quality_feedback",
			"hardware_backed_signing",
			"origin_proof",
		},
	}
}
