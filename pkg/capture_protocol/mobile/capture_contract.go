// Package mobile implements native mobile capture specifications for iOS and Android.
// VE-900/VE-4F: Mobile capture contract - image/video format specifications
//
// This file defines the capture protocol contract including format requirements,
// size limits, metadata requirements, and device ID specifications.
package mobile

import (
	"time"
)

// ============================================================================
// Protocol Version and Constants
// ============================================================================

// CaptureProtocolVersion is the current mobile capture protocol version
const CaptureProtocolVersion = "1.0.0"

// ImageFormat represents supported image formats
type ImageFormat string

const (
	// ImageFormatJPEG is JPEG format (preferred for photos)
	ImageFormatJPEG ImageFormat = "jpeg"

	// ImageFormatHEIC is HEIC/HEIF format (iOS default, high compression)
	ImageFormatHEIC ImageFormat = "heic"

	// ImageFormatPNG is PNG format (lossless, larger files)
	ImageFormatPNG ImageFormat = "png"

	// ImageFormatWebP is WebP format (good compression, wide support)
	ImageFormatWebP ImageFormat = "webp"
)

// VideoFormat represents supported video formats
type VideoFormat string

const (
	// VideoFormatMP4 is MP4 format (H.264/H.265)
	VideoFormatMP4 VideoFormat = "mp4"

	// VideoFormatMOV is MOV format (iOS native)
	VideoFormatMOV VideoFormat = "mov"

	// VideoFormatWebM is WebM format (VP8/VP9)
	VideoFormatWebM VideoFormat = "webm"
)

// VideoCodec represents supported video codecs
type VideoCodec string

const (
	// VideoCodecH264 is H.264/AVC codec
	VideoCodecH264 VideoCodec = "h264"

	// VideoCodecH265 is H.265/HEVC codec
	VideoCodecH265 VideoCodec = "hevc"

	// VideoCodecVP9 is VP9 codec
	VideoCodecVP9 VideoCodec = "vp9"
)

// ============================================================================
// Capture Contract Specification
// ============================================================================

// CaptureContract defines the complete capture protocol contract
type CaptureContract struct {
	// Version is the contract version
	Version string `json:"version"`

	// Document capture specifications
	Document DocumentCaptureSpec `json:"document"`

	// Selfie capture specifications
	Selfie SelfieCaptureSpec `json:"selfie"`

	// Liveness video specifications
	Liveness LivenessCaptureSpec `json:"liveness"`

	// Device requirements
	Device DeviceRequirements `json:"device"`

	// Encryption requirements
	Encryption EncryptionRequirements `json:"encryption"`

	// Metadata requirements
	Metadata MetadataRequirements `json:"metadata"`

	// Timing constraints
	Timing TimingConstraints `json:"timing"`
}

// DefaultCaptureContract returns the default capture contract
func DefaultCaptureContract() CaptureContract {
	return CaptureContract{
		Version:    CaptureProtocolVersion,
		Document:   DefaultDocumentCaptureSpec(),
		Selfie:     DefaultSelfieCaptureSpec(),
		Liveness:   DefaultLivenessCaptureSpec(),
		Device:     DefaultDeviceRequirements(),
		Encryption: DefaultEncryptionRequirements(),
		Metadata:   DefaultMetadataRequirements(),
		Timing:     DefaultTimingConstraints(),
	}
}

// ============================================================================
// Document Capture Specification
// ============================================================================

// DocumentCaptureSpec defines document capture requirements
type DocumentCaptureSpec struct {
	// Image format requirements
	AcceptedFormats []ImageFormat `json:"accepted_formats"`
	PreferredFormat ImageFormat   `json:"preferred_format"`

	// Resolution requirements
	MinWidth       int `json:"min_width"`        // Minimum width in pixels
	MinHeight      int `json:"min_height"`       // Minimum height in pixels
	MaxWidth       int `json:"max_width"`        // Maximum width in pixels
	MaxHeight      int `json:"max_height"`       // Maximum height in pixels
	MinMegapixels  float64 `json:"min_megapixels"` // Minimum megapixels

	// File size limits
	MaxFileSizeBytes   int64 `json:"max_file_size_bytes"`
	MaxFileSizeAfterCompression int64 `json:"max_file_size_after_compression"`

	// Quality requirements
	MinQualityScore    int     `json:"min_quality_score"`     // 0-100
	MaxBlurScore       float64 `json:"max_blur_score"`        // Laplacian variance
	MaxSkewDegrees     float64 `json:"max_skew_degrees"`
	MinDocumentCoverage float64 `json:"min_document_coverage"` // % of frame
	MaxGlarePercentage  float64 `json:"max_glare_percentage"`  // % of document area

	// Compression settings
	JPEGQuality         int  `json:"jpeg_quality"`          // 0-100
	StripEXIF           bool `json:"strip_exif"`
	StripGPS            bool `json:"strip_gps"`
	AllowCompression    bool `json:"allow_compression"`
	TargetFileSizeBytes int64 `json:"target_file_size_bytes"`

	// Document type restrictions
	SupportedDocumentTypes []DocumentType `json:"supported_document_types"`
}

// DefaultDocumentCaptureSpec returns default document capture specifications
func DefaultDocumentCaptureSpec() DocumentCaptureSpec {
	return DocumentCaptureSpec{
		AcceptedFormats: []ImageFormat{ImageFormatJPEG, ImageFormatHEIC, ImageFormatPNG},
		PreferredFormat: ImageFormatJPEG,

		MinWidth:      1024,
		MinHeight:     768,
		MaxWidth:      4096,
		MaxHeight:     4096,
		MinMegapixels: 2.0,

		MaxFileSizeBytes:            20 * 1024 * 1024, // 20MB raw
		MaxFileSizeAfterCompression: 5 * 1024 * 1024,  // 5MB after compression

		MinQualityScore:     70,
		MaxBlurScore:        100,
		MaxSkewDegrees:      10,
		MinDocumentCoverage: 0.6, // 60% of frame
		MaxGlarePercentage:  0.15,

		JPEGQuality:         85,
		StripEXIF:           true,
		StripGPS:            true,
		AllowCompression:    true,
		TargetFileSizeBytes: 2 * 1024 * 1024, // 2MB target

		SupportedDocumentTypes: []DocumentType{
			DocumentTypeIDCard,
			DocumentTypePassport,
			DocumentTypeDriversLicense,
		},
	}
}

// ============================================================================
// Selfie Capture Specification
// ============================================================================

// SelfieCaptureSpec defines selfie capture requirements
type SelfieCaptureSpec struct {
	// Image format requirements
	AcceptedFormats []ImageFormat `json:"accepted_formats"`
	PreferredFormat ImageFormat   `json:"preferred_format"`

	// Resolution requirements
	MinWidth      int     `json:"min_width"`
	MinHeight     int     `json:"min_height"`
	MaxWidth      int     `json:"max_width"`
	MaxHeight     int     `json:"max_height"`
	MinMegapixels float64 `json:"min_megapixels"`

	// File size limits
	MaxFileSizeBytes            int64 `json:"max_file_size_bytes"`
	MaxFileSizeAfterCompression int64 `json:"max_file_size_after_compression"`

	// Face requirements
	MinFaceSizePercent float64 `json:"min_face_size_percent"` // % of frame
	MaxFaceSizePercent float64 `json:"max_face_size_percent"` // % of frame
	RequireCentered    bool    `json:"require_centered"`
	MaxFaceCount       int     `json:"max_face_count"` // 1 for identity

	// Quality requirements
	MinQualityScore int     `json:"min_quality_score"`
	MaxBlurScore    float64 `json:"max_blur_score"`

	// Compression settings
	JPEGQuality         int   `json:"jpeg_quality"`
	StripEXIF           bool  `json:"strip_exif"`
	StripGPS            bool  `json:"strip_gps"`
	TargetFileSizeBytes int64 `json:"target_file_size_bytes"`
}

// DefaultSelfieCaptureSpec returns default selfie capture specifications
func DefaultSelfieCaptureSpec() SelfieCaptureSpec {
	return SelfieCaptureSpec{
		AcceptedFormats: []ImageFormat{ImageFormatJPEG, ImageFormatHEIC},
		PreferredFormat: ImageFormatJPEG,

		MinWidth:      640,
		MinHeight:     480,
		MaxWidth:      2048,
		MaxHeight:     2048,
		MinMegapixels: 0.5,

		MaxFileSizeBytes:            10 * 1024 * 1024,
		MaxFileSizeAfterCompression: 2 * 1024 * 1024,

		MinFaceSizePercent: 0.15, // Face must be at least 15% of frame
		MaxFaceSizePercent: 0.70, // Face must be at most 70% of frame
		RequireCentered:    true,
		MaxFaceCount:       1,

		MinQualityScore: 75,
		MaxBlurScore:    80,

		JPEGQuality:         90,
		StripEXIF:           true,
		StripGPS:            true,
		TargetFileSizeBytes: 1 * 1024 * 1024,
	}
}

// ============================================================================
// Liveness Video Specification
// ============================================================================

// LivenessCaptureSpec defines liveness video capture requirements
type LivenessCaptureSpec struct {
	// Video format requirements
	AcceptedFormats []VideoFormat `json:"accepted_formats"`
	PreferredFormat VideoFormat   `json:"preferred_format"`
	AcceptedCodecs  []VideoCodec  `json:"accepted_codecs"`
	PreferredCodec  VideoCodec    `json:"preferred_codec"`

	// Resolution requirements
	MinWidth  int `json:"min_width"`
	MinHeight int `json:"min_height"`
	MaxWidth  int `json:"max_width"`
	MaxHeight int `json:"max_height"`

	// Frame rate requirements
	MinFrameRate int `json:"min_frame_rate"`
	MaxFrameRate int `json:"max_frame_rate"`

	// Duration requirements
	MinDurationSeconds float64 `json:"min_duration_seconds"`
	MaxDurationSeconds float64 `json:"max_duration_seconds"`

	// File size limits
	MaxFileSizeBytes int64 `json:"max_file_size_bytes"`

	// Quality requirements
	MinBitrate int `json:"min_bitrate"` // kbps
	MaxBitrate int `json:"max_bitrate"` // kbps

	// Liveness requirements
	RequireActiveChallenge bool `json:"require_active_challenge"`
	MinChallengesRequired  int  `json:"min_challenges_required"`
	MinConfidence          float64 `json:"min_confidence"`
}

// DefaultLivenessCaptureSpec returns default liveness video specifications
func DefaultLivenessCaptureSpec() LivenessCaptureSpec {
	return LivenessCaptureSpec{
		AcceptedFormats: []VideoFormat{VideoFormatMP4, VideoFormatMOV},
		PreferredFormat: VideoFormatMP4,
		AcceptedCodecs:  []VideoCodec{VideoCodecH264, VideoCodecH265},
		PreferredCodec:  VideoCodecH264,

		MinWidth:  640,
		MinHeight: 480,
		MaxWidth:  1920,
		MaxHeight: 1080,

		MinFrameRate: 24,
		MaxFrameRate: 60,

		MinDurationSeconds: 2.0,
		MaxDurationSeconds: 10.0,

		MaxFileSizeBytes: 50 * 1024 * 1024, // 50MB

		MinBitrate: 1000,  // 1 Mbps
		MaxBitrate: 10000, // 10 Mbps

		RequireActiveChallenge: false, // Passive by default
		MinChallengesRequired:  1,
		MinConfidence:          0.85,
	}
}

// ============================================================================
// Device Requirements
// ============================================================================

// DeviceRequirements defines device requirements for capture
type DeviceRequirements struct {
	// Platform requirements
	MinIOSVersion      string `json:"min_ios_version"`
	MinAndroidAPI      int    `json:"min_android_api"`

	// Camera requirements
	RequireBackCamera  bool    `json:"require_back_camera"`
	RequireFrontCamera bool    `json:"require_front_camera"`
	MinCameraMP        float64 `json:"min_camera_mp"` // Megapixels

	// Security requirements
	RequireSecureEnclave    bool `json:"require_secure_enclave"`
	RequireHardwareKeystore bool `json:"require_hardware_keystore"`
	AllowRooted             bool `json:"allow_rooted"`
	AllowJailbroken         bool `json:"allow_jailbroken"`
	AllowEmulator           bool `json:"allow_emulator"`

	// Device attestation
	RequireDeviceAttestation bool   `json:"require_device_attestation"`
	AttestationProvider      string `json:"attestation_provider"` // "safetynet", "play_integrity", "devicecheck", "app_attest"
}

// DefaultDeviceRequirements returns default device requirements
func DefaultDeviceRequirements() DeviceRequirements {
	return DeviceRequirements{
		MinIOSVersion: "14.0",
		MinAndroidAPI: 26, // Android 8.0

		RequireBackCamera:  true,
		RequireFrontCamera: true,
		MinCameraMP:        2.0,

		RequireSecureEnclave:    false, // Not all devices support
		RequireHardwareKeystore: false, // Not all devices support
		AllowRooted:             false,
		AllowJailbroken:         false,
		AllowEmulator:           false,

		RequireDeviceAttestation: true,
		AttestationProvider:      "auto", // Auto-detect based on platform
	}
}

// ============================================================================
// Encryption Requirements
// ============================================================================

// EncryptionRequirements defines encryption requirements for captured data
type EncryptionRequirements struct {
	// Algorithm requirements
	Algorithm         string `json:"algorithm"`         // e.g., "X25519-XSalsa20-Poly1305"
	AlgorithmVersion  uint32 `json:"algorithm_version"`

	// Salt requirements
	MinSaltLength int `json:"min_salt_length"`
	MaxSaltLength int `json:"max_salt_length"`

	// Signature requirements
	RequireClientSignature bool   `json:"require_client_signature"`
	RequireUserSignature   bool   `json:"require_user_signature"`
	SignatureAlgorithm     string `json:"signature_algorithm"` // "ed25519" or "secp256k1"

	// Nonce requirements
	NonceLength int `json:"nonce_length"`
}

// DefaultEncryptionRequirements returns default encryption requirements
func DefaultEncryptionRequirements() EncryptionRequirements {
	return EncryptionRequirements{
		Algorithm:        "X25519-XSalsa20-Poly1305",
		AlgorithmVersion: 1,

		MinSaltLength: 32,
		MaxSaltLength: 64,

		RequireClientSignature: true,
		RequireUserSignature:   true,
		SignatureAlgorithm:     "ed25519",

		NonceLength: 24,
	}
}

// ============================================================================
// Metadata Requirements
// ============================================================================

// MetadataRequirements defines required metadata for captures
type MetadataRequirements struct {
	// Required fields
	RequireDeviceFingerprint bool `json:"require_device_fingerprint"`
	RequireSessionID         bool `json:"require_session_id"`
	RequireCaptureTimestamp  bool `json:"require_capture_timestamp"`
	RequireClientID          bool `json:"require_client_id"`
	RequireClientVersion     bool `json:"require_client_version"`

	// Optional fields
	IncludePlatformInfo bool `json:"include_platform_info"`
	IncludeQualityScore bool `json:"include_quality_score"`
	IncludeLivenessInfo bool `json:"include_liveness_info"`
	IncludeGeoHint      bool `json:"include_geo_hint"`

	// Privacy settings
	AllowPreciseLocation bool `json:"allow_precise_location"` // Should always be false
	AllowDeviceID        bool `json:"allow_device_id"`        // Hashed/scoped only
}

// DefaultMetadataRequirements returns default metadata requirements
func DefaultMetadataRequirements() MetadataRequirements {
	return MetadataRequirements{
		RequireDeviceFingerprint: true,
		RequireSessionID:         true,
		RequireCaptureTimestamp:  true,
		RequireClientID:          true,
		RequireClientVersion:     true,

		IncludePlatformInfo: true,
		IncludeQualityScore: true,
		IncludeLivenessInfo: true,
		IncludeGeoHint:      false, // Off by default for privacy

		AllowPreciseLocation: false, // Never allow precise location
		AllowDeviceID:        true,  // Hashed device ID only
	}
}

// ============================================================================
// Timing Constraints
// ============================================================================

// TimingConstraints defines timing requirements for the capture flow
type TimingConstraints struct {
	// Capture timing
	MaxCaptureAgeSeconds  int64 `json:"max_capture_age_seconds"`  // Max age of capture before upload
	MaxUploadDelaySeconds int64 `json:"max_upload_delay_seconds"` // Max delay between capture and upload

	// Salt timing
	SaltValiditySeconds   int64 `json:"salt_validity_seconds"`   // How long a salt is valid
	ReplayWindowSeconds   int64 `json:"replay_window_seconds"`   // Window for replay detection

	// Session timing
	SessionTimeoutSeconds int64 `json:"session_timeout_seconds"` // Capture session timeout
	MaxRetryDelaySeconds  int64 `json:"max_retry_delay_seconds"` // Max delay between retries

	// Clock skew tolerance
	MaxClockSkewSeconds int64 `json:"max_clock_skew_seconds"` // Allowed clock skew
}

// DefaultTimingConstraints returns default timing constraints
func DefaultTimingConstraints() TimingConstraints {
	return TimingConstraints{
		MaxCaptureAgeSeconds:  300,  // 5 minutes
		MaxUploadDelaySeconds: 600,  // 10 minutes
		SaltValiditySeconds:   300,  // 5 minutes
		ReplayWindowSeconds:   600,  // 10 minutes
		SessionTimeoutSeconds: 1800, // 30 minutes
		MaxRetryDelaySeconds:  60,   // 1 minute
		MaxClockSkewSeconds:   30,   // 30 seconds
	}
}

// MaxCaptureAge returns the max capture age as a duration
func (t TimingConstraints) MaxCaptureAge() time.Duration {
	return time.Duration(t.MaxCaptureAgeSeconds) * time.Second
}

// MaxUploadDelay returns the max upload delay as a duration
func (t TimingConstraints) MaxUploadDelay() time.Duration {
	return time.Duration(t.MaxUploadDelaySeconds) * time.Second
}

// SaltValidity returns the salt validity as a duration
func (t TimingConstraints) SaltValidity() time.Duration {
	return time.Duration(t.SaltValiditySeconds) * time.Second
}

// ReplayWindow returns the replay window as a duration
func (t TimingConstraints) ReplayWindow() time.Duration {
	return time.Duration(t.ReplayWindowSeconds) * time.Second
}

// SessionTimeout returns the session timeout as a duration
func (t TimingConstraints) SessionTimeout() time.Duration {
	return time.Duration(t.SessionTimeoutSeconds) * time.Second
}

// MaxRetryDelay returns the max retry delay as a duration
func (t TimingConstraints) MaxRetryDelay() time.Duration {
	return time.Duration(t.MaxRetryDelaySeconds) * time.Second
}

// MaxClockSkew returns the max clock skew as a duration
func (t TimingConstraints) MaxClockSkew() time.Duration {
	return time.Duration(t.MaxClockSkewSeconds) * time.Second
}

// ============================================================================
// Contract Validation
// ============================================================================

// ValidateImageSpec validates an image against document specifications
func (d *DocumentCaptureSpec) ValidateImageSpec(width, height int, format ImageFormat, fileSize int64) error {
	if width < d.MinWidth {
		return ErrCaptureResolutionTooLow.WithDetails(
			"min_width", d.MinWidth,
			"actual_width", width,
		)
	}
	if height < d.MinHeight {
		return ErrCaptureResolutionTooLow.WithDetails(
			"min_height", d.MinHeight,
			"actual_height", height,
		)
	}
	if width > d.MaxWidth || height > d.MaxHeight {
		return ErrCaptureResolutionTooHigh.WithDetails(
			"max_width", d.MaxWidth,
			"max_height", d.MaxHeight,
		)
	}

	megapixels := float64(width*height) / 1_000_000
	if megapixels < d.MinMegapixels {
		return ErrCaptureResolutionTooLow.WithDetails(
			"min_megapixels", d.MinMegapixels,
			"actual_megapixels", megapixels,
		)
	}

	if fileSize > d.MaxFileSizeBytes {
		return ErrCaptureFileTooLarge.WithDetails(
			"max_size", d.MaxFileSizeBytes,
			"actual_size", fileSize,
		)
	}

	formatValid := false
	for _, f := range d.AcceptedFormats {
		if f == format {
			formatValid = true
			break
		}
	}
	if !formatValid {
		return ErrCaptureFormatInvalid.WithDetails(
			"format", format,
			"accepted", d.AcceptedFormats,
		)
	}

	return nil
}

// ValidateSelfieSpec validates an image against selfie specifications
func (s *SelfieCaptureSpec) ValidateSelfieSpec(width, height int, format ImageFormat, fileSize int64) error {
	if width < s.MinWidth || height < s.MinHeight {
		return ErrCaptureResolutionTooLow.WithDetails(
			"min_width", s.MinWidth,
			"min_height", s.MinHeight,
		)
	}
	if width > s.MaxWidth || height > s.MaxHeight {
		return ErrCaptureResolutionTooHigh.WithDetails(
			"max_width", s.MaxWidth,
			"max_height", s.MaxHeight,
		)
	}

	if fileSize > s.MaxFileSizeBytes {
		return ErrCaptureFileTooLarge.WithDetails(
			"max_size", s.MaxFileSizeBytes,
			"actual_size", fileSize,
		)
	}

	formatValid := false
	for _, f := range s.AcceptedFormats {
		if f == format {
			formatValid = true
			break
		}
	}
	if !formatValid {
		return ErrCaptureFormatInvalid.WithDetails(
			"format", format,
			"accepted", s.AcceptedFormats,
		)
	}

	return nil
}

// ValidateVideoSpec validates a video against liveness specifications
func (l *LivenessCaptureSpec) ValidateVideoSpec(
	width, height int,
	format VideoFormat,
	codec VideoCodec,
	frameRate int,
	durationSeconds float64,
	fileSize int64,
) error {
	if width < l.MinWidth || height < l.MinHeight {
		return ErrCaptureResolutionTooLow.WithDetails(
			"min_width", l.MinWidth,
			"min_height", l.MinHeight,
		)
	}
	if width > l.MaxWidth || height > l.MaxHeight {
		return ErrCaptureResolutionTooHigh.WithDetails(
			"max_width", l.MaxWidth,
			"max_height", l.MaxHeight,
		)
	}

	if frameRate < l.MinFrameRate {
		return ErrCaptureFrameRateTooLow.WithDetails(
			"min_frame_rate", l.MinFrameRate,
			"actual", frameRate,
		)
	}

	if durationSeconds < l.MinDurationSeconds {
		return ErrCaptureDurationTooShort.WithDetails(
			"min_duration", l.MinDurationSeconds,
			"actual", durationSeconds,
		)
	}
	if durationSeconds > l.MaxDurationSeconds {
		return ErrCaptureDurationTooLong.WithDetails(
			"max_duration", l.MaxDurationSeconds,
			"actual", durationSeconds,
		)
	}

	if fileSize > l.MaxFileSizeBytes {
		return ErrCaptureFileTooLarge.WithDetails(
			"max_size", l.MaxFileSizeBytes,
			"actual_size", fileSize,
		)
	}

	formatValid := false
	for _, f := range l.AcceptedFormats {
		if f == format {
			formatValid = true
			break
		}
	}
	if !formatValid {
		return ErrCaptureFormatInvalid.WithDetails(
			"format", format,
			"accepted", l.AcceptedFormats,
		)
	}

	codecValid := false
	for _, c := range l.AcceptedCodecs {
		if c == codec {
			codecValid = true
			break
		}
	}
	if !codecValid {
		return ErrCaptureCodecInvalid.WithDetails(
			"codec", codec,
			"accepted", l.AcceptedCodecs,
		)
	}

	return nil
}

