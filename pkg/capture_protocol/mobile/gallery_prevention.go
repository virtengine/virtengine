package mobile

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"time"
)

// ============================================================================
// Gallery Upload Prevention Mechanism
// VE-900: Gallery uploads blocked - metadata validation
// ============================================================================

// GalleryPreventionConfig configures gallery upload prevention
type GalleryPreventionConfig struct {
	// StrictMode blocks all non-camera sources
	StrictMode bool `json:"strict_mode"`

	// MaxImageAge is maximum age for a captured image
	MaxImageAge time.Duration `json:"max_image_age"`

	// RequireTimestampBinding requires timestamp in capture proof
	RequireTimestampBinding bool `json:"require_timestamp_binding"`

	// ValidateEXIF enables EXIF metadata validation
	ValidateEXIF bool `json:"validate_exif"`

	// RejectEditedImages rejects images with editing software markers
	RejectEditedImages bool `json:"reject_edited_images"`

	// RequireRawCapture requires raw camera output (no processing)
	RequireRawCapture bool `json:"require_raw_capture"`
}

// DefaultGalleryPreventionConfig returns default gallery prevention config
func DefaultGalleryPreventionConfig() GalleryPreventionConfig {
	return GalleryPreventionConfig{
		StrictMode:              true,
		MaxImageAge:             30 * time.Second,
		RequireTimestampBinding: true,
		ValidateEXIF:            true,
		RejectEditedImages:      true,
		RequireRawCapture:       false,
	}
}

// ============================================================================
// Capture Origin Proof
// ============================================================================

// CaptureOriginProof proves image came from live camera capture
type CaptureOriginProof struct {
	// ProofVersion is the proof format version
	ProofVersion uint32 `json:"proof_version"`

	// CaptureTimestamp is when capture occurred (from device)
	CaptureTimestamp time.Time `json:"capture_timestamp"`

	// SystemTimestamp is system time at capture
	SystemTimestamp time.Time `json:"system_timestamp"`

	// MonotonicTimestamp is monotonic clock value at capture (nanoseconds)
	MonotonicTimestamp int64 `json:"monotonic_timestamp"`

	// CameraSessionID links to the camera session
	CameraSessionID string `json:"camera_session_id"`

	// FrameNumber is the frame number in camera session
	FrameNumber int64 `json:"frame_number"`

	// DeviceFingerprint is the device fingerprint
	DeviceFingerprint string `json:"device_fingerprint"`

	// HardwareBinding contains hardware-specific proof
	HardwareBinding *HardwareBindingProof `json:"hardware_binding,omitempty"`

	// ImageMetadata contains verified image metadata
	ImageMetadata CaptureImageMetadata `json:"image_metadata"`

	// ProofHash is SHA256(all proof fields)
	ProofHash []byte `json:"proof_hash"`
}

// HardwareBindingProof contains hardware-level capture proof
type HardwareBindingProof struct {
	// Platform is ios or android
	Platform Platform `json:"platform"`

	// IOSProof contains iOS-specific proof (AVCaptureSession data)
	IOSProof *IOSCaptureProof `json:"ios_proof,omitempty"`

	// AndroidProof contains Android-specific proof (Camera2 data)
	AndroidProof *AndroidCaptureProof `json:"android_proof,omitempty"`
}

// IOSCaptureProof contains iOS AVFoundation capture proof
type IOSCaptureProof struct {
	// SessionPreset is the AVCaptureSession preset used
	SessionPreset string `json:"session_preset"`

	// DeviceType is the AVCaptureDevice type
	DeviceType string `json:"device_type"`

	// DevicePosition is front/back camera position
	DevicePosition string `json:"device_position"`

	// LensPosition is the lens position at capture
	LensPosition float32 `json:"lens_position"`

	// ExposureDuration is exposure duration in seconds
	ExposureDuration float64 `json:"exposure_duration"`

	// ISO is the ISO value at capture
	ISO float32 `json:"iso"`

	// PhotoOutputUUID is the unique photo output identifier
	PhotoOutputUUID string `json:"photo_output_uuid"`

	// PhotoID is the photo's unique identifier
	PhotoID string `json:"photo_id"`

	// TimestampFromSensor is CMTime from sensor
	TimestampFromSensor int64 `json:"timestamp_from_sensor"`

	// SystemBootTime verifies continuity with boot time
	SystemBootTime int64 `json:"system_boot_time"`
}

// AndroidCaptureProof contains Android Camera2 capture proof
type AndroidCaptureProof struct {
	// CameraID is the Camera2 camera identifier
	CameraID string `json:"camera_id"`

	// LensFacing is front/back lens facing
	LensFacing int `json:"lens_facing"`

	// SensorTimestamp is the sensor timestamp (nanoseconds)
	SensorTimestamp int64 `json:"sensor_timestamp"`

	// FrameNumber is the Camera2 frame number
	FrameNumber int64 `json:"frame_number"`

	// ExposureTimeNanos is exposure time in nanoseconds
	ExposureTimeNanos int64 `json:"exposure_time_nanos"`

	// SensitivityISO is the ISO sensitivity
	SensitivityISO int `json:"sensitivity_iso"`

	// FocusDistance is the focus distance
	FocusDistance float32 `json:"focus_distance"`

	// SessionID is the CameraCaptureSession identifier
	SessionID string `json:"session_id"`

	// BootTimeNanos is system boot time at capture
	BootTimeNanos int64 `json:"boot_time_nanos"`

	// RealtimeNanos is CLOCK_REALTIME at capture
	RealtimeNanos int64 `json:"realtime_nanos"`
}

// CaptureImageMetadata contains image metadata for validation
type CaptureImageMetadata struct {
	// Width is image width in pixels
	Width int `json:"width"`

	// Height is image height in pixels
	Height int `json:"height"`

	// Format is the image format (jpeg, heic, etc.)
	Format string `json:"format"`

	// ColorSpace is the color space
	ColorSpace string `json:"color_space"`

	// BitDepth is bits per pixel
	BitDepth int `json:"bit_depth"`

	// HasEXIF indicates if original had EXIF (should be stripped)
	HasEXIF bool `json:"has_exif"`

	// EXIFStripped confirms EXIF was stripped
	EXIFStripped bool `json:"exif_stripped"`

	// OriginalSizeBytes is original image size
	OriginalSizeBytes int64 `json:"original_size_bytes"`

	// StrippedSizeBytes is size after metadata stripping
	StrippedSizeBytes int64 `json:"stripped_size_bytes"`

	// ImageHash is SHA256 of the final image data
	ImageHash []byte `json:"image_hash"`
}

// ComputeProofHash computes the proof hash
func (p *CaptureOriginProof) ComputeProofHash() []byte {
	h := sha256.New()

	// Version
	vb := make([]byte, 4)
	binary.BigEndian.PutUint32(vb, p.ProofVersion)
	h.Write(vb)

	// Timestamps
	tb := make([]byte, 8)
	//nolint:gosec // G115: UnixNano timestamp safe for uint64
	binary.BigEndian.PutUint64(tb, uint64(p.CaptureTimestamp.UnixNano()))
	h.Write(tb)
	//nolint:gosec // G115: UnixNano timestamp safe for uint64
	binary.BigEndian.PutUint64(tb, uint64(p.SystemTimestamp.UnixNano()))
	h.Write(tb)
	//nolint:gosec // G115: MonotonicTimestamp is positive
	binary.BigEndian.PutUint64(tb, uint64(p.MonotonicTimestamp))
	h.Write(tb)

	// Session and frame
	h.Write([]byte(p.CameraSessionID))
	//nolint:gosec // G115: FrameNumber is positive int
	binary.BigEndian.PutUint64(tb, uint64(p.FrameNumber))
	h.Write(tb)

	// Device
	h.Write([]byte(p.DeviceFingerprint))

	// Image hash
	h.Write(p.ImageMetadata.ImageHash)

	return h.Sum(nil)
}

// Verify verifies the proof hash
func (p *CaptureOriginProof) Verify() bool {
	expected := p.ComputeProofHash()
	return bytes.Equal(p.ProofHash, expected)
}

// ============================================================================
// Gallery Detection
// ============================================================================

// GalleryDetectionResult contains gallery upload detection results
type GalleryDetectionResult struct {
	// IsGalleryUpload indicates if image is from gallery
	IsGalleryUpload bool `json:"is_gallery_upload"`

	// IsLiveCapture indicates if image is from live camera
	IsLiveCapture bool `json:"is_live_capture"`

	// Confidence is detection confidence (0-1)
	Confidence float64 `json:"confidence"`

	// DetectionMethod is how detection was performed
	DetectionMethod GalleryDetectionMethod `json:"detection_method"`

	// Indicators contains specific indicators found
	Indicators []GalleryIndicator `json:"indicators"`

	// Blocked indicates if upload was blocked
	Blocked bool `json:"blocked"`

	// BlockReason is the reason for blocking
	BlockReason string `json:"block_reason,omitempty"`
}

// GalleryDetectionMethod represents detection methods
type GalleryDetectionMethod string

const (
	// DetectionMethodPlatformAPI uses platform-specific APIs
	DetectionMethodPlatformAPI GalleryDetectionMethod = "platform_api"

	// DetectionMethodTimestampAnalysis analyzes timestamps
	DetectionMethodTimestampAnalysis GalleryDetectionMethod = "timestamp_analysis"

	// DetectionMethodMetadataAnalysis analyzes image metadata
	DetectionMethodMetadataAnalysis GalleryDetectionMethod = "metadata_analysis"

	// DetectionMethodOriginProof uses cryptographic origin proof
	DetectionMethodOriginProof GalleryDetectionMethod = "origin_proof"

	// DetectionMethodMultiple uses multiple methods
	DetectionMethodMultiple GalleryDetectionMethod = "multiple"
)

// GalleryIndicator represents a gallery upload indicator
type GalleryIndicator struct {
	// Type is the indicator type
	Type GalleryIndicatorType `json:"type"`

	// Description explains the indicator
	Description string `json:"description"`

	// Severity is the indicator severity
	Severity GalleryIndicatorSeverity `json:"severity"`

	// Value is the detected value
	Value string `json:"value,omitempty"`

	// Expected is the expected value
	Expected string `json:"expected,omitempty"`
}

// GalleryIndicatorType represents indicator types
type GalleryIndicatorType string

const (
	// IndicatorNonCameraSource indicates non-camera image source
	IndicatorNonCameraSource GalleryIndicatorType = "non_camera_source"

	// IndicatorOldTimestamp indicates image is too old
	IndicatorOldTimestamp GalleryIndicatorType = "old_timestamp"

	// IndicatorFutureTimestamp indicates future timestamp
	IndicatorFutureTimestamp GalleryIndicatorType = "future_timestamp"

	// IndicatorEditingSoftware indicates editing software markers
	IndicatorEditingSoftware GalleryIndicatorType = "editing_software"

	// IndicatorScreenshot indicates screenshot markers
	IndicatorScreenshot GalleryIndicatorType = "screenshot"

	// IndicatorDownloaded indicates downloaded image markers
	IndicatorDownloaded GalleryIndicatorType = "downloaded"

	// IndicatorMissingCameraData indicates missing camera metadata
	IndicatorMissingCameraData GalleryIndicatorType = "missing_camera_data"

	// IndicatorInconsistentTimestamps indicates timestamp inconsistency
	IndicatorInconsistentTimestamps GalleryIndicatorType = "inconsistent_timestamps"

	// IndicatorNoOriginProof indicates missing origin proof
	IndicatorNoOriginProof GalleryIndicatorType = "no_origin_proof"

	// IndicatorInvalidOriginProof indicates invalid origin proof
	IndicatorInvalidOriginProof GalleryIndicatorType = "invalid_origin_proof"
)

// GalleryIndicatorSeverity represents indicator severity
type GalleryIndicatorSeverity string

const (
	// SeverityInfo is informational
	SeverityInfo GalleryIndicatorSeverity = "info"

	// SeveritySuspicious is suspicious but not blocking
	SeveritySuspicious GalleryIndicatorSeverity = "suspicious"

	// SeverityBlocking causes upload to be blocked
	SeverityBlocking GalleryIndicatorSeverity = "blocking"
)

// ============================================================================
// Gallery Prevention Validator
// ============================================================================

// GalleryPreventionValidator validates capture origin
type GalleryPreventionValidator struct {
	config GalleryPreventionConfig
}

// NewGalleryPreventionValidator creates a new validator
func NewGalleryPreventionValidator(config GalleryPreventionConfig) *GalleryPreventionValidator {
	return &GalleryPreventionValidator{config: config}
}

// ValidateCaptureOrigin validates a capture's origin
func (v *GalleryPreventionValidator) ValidateCaptureOrigin(
	proof *CaptureOriginProof,
	imageData []byte,
) (*GalleryDetectionResult, error) {
	result := &GalleryDetectionResult{
		IsLiveCapture:   true,
		IsGalleryUpload: false,
		Confidence:      1.0,
		DetectionMethod: DetectionMethodMultiple,
		Indicators:      make([]GalleryIndicator, 0),
	}

	// Run validation steps
	if done := v.validateProofExists(proof, result); done {
		return result, nil
	}

	if done := v.validateProofHash(proof, result); done {
		return result, nil
	}

	if done := v.validateTimestamp(proof, result); done {
		return result, nil
	}

	if done := v.validateImageHash(proof, imageData, result); done {
		return result, nil
	}

	v.validateEXIF(proof, result)

	if proof != nil && proof.HardwareBinding != nil {
		if err := v.validateHardwareBinding(proof.HardwareBinding, result); err != nil {
			return result, err
		}
	}

	v.determineFinalResult(result)

	return result, nil
}

// validateProofExists checks if the proof exists
func (v *GalleryPreventionValidator) validateProofExists(proof *CaptureOriginProof, result *GalleryDetectionResult) bool {
	if proof == nil {
		result.IsLiveCapture = false
		result.IsGalleryUpload = true
		result.Confidence = 0.9
		result.Indicators = append(result.Indicators, GalleryIndicator{
			Type:        IndicatorNoOriginProof,
			Description: "No capture origin proof provided",
			Severity:    SeverityBlocking,
		})
		if v.config.StrictMode {
			result.Blocked = true
			result.BlockReason = "Missing capture origin proof"
			return true
		}
	}
	return false
}

// validateProofHash verifies the proof hash
func (v *GalleryPreventionValidator) validateProofHash(proof *CaptureOriginProof, result *GalleryDetectionResult) bool {
	if proof == nil {
		return false
	}
	if !proof.Verify() {
		result.IsLiveCapture = false
		result.IsGalleryUpload = true
		result.Confidence = 0.95
		result.Indicators = append(result.Indicators, GalleryIndicator{
			Type:        IndicatorInvalidOriginProof,
			Description: "Origin proof hash verification failed",
			Severity:    SeverityBlocking,
		})
		result.Blocked = true
		result.BlockReason = "Invalid capture origin proof"
		return true
	}
	return false
}

// validateTimestamp validates the capture timestamp
func (v *GalleryPreventionValidator) validateTimestamp(proof *CaptureOriginProof, result *GalleryDetectionResult) bool {
	if proof == nil || !v.config.RequireTimestampBinding {
		return false
	}

	age := time.Since(proof.CaptureTimestamp)
	if age > v.config.MaxImageAge {
		result.IsLiveCapture = false
		result.Confidence *= 0.5
		result.Indicators = append(result.Indicators, GalleryIndicator{
			Type:        IndicatorOldTimestamp,
			Description: "Capture timestamp is too old",
			Severity:    SeverityBlocking,
			Value:       age.String(),
			Expected:    v.config.MaxImageAge.String(),
		})
		if v.config.StrictMode {
			result.Blocked = true
			result.BlockReason = "Capture is too old"
			return true
		}
	}

	if proof.CaptureTimestamp.After(time.Now().Add(time.Minute)) {
		result.IsLiveCapture = false
		result.Confidence *= 0.3
		result.Indicators = append(result.Indicators, GalleryIndicator{
			Type:        IndicatorFutureTimestamp,
			Description: "Capture timestamp is in the future",
			Severity:    SeverityBlocking,
		})
		result.Blocked = true
		result.BlockReason = "Invalid future timestamp"
		return true
	}

	return false
}

// validateImageHash validates the image hash matches the proof
func (v *GalleryPreventionValidator) validateImageHash(proof *CaptureOriginProof, imageData []byte, result *GalleryDetectionResult) bool {
	if proof == nil || imageData == nil {
		return false
	}

	computedHash := sha256.Sum256(imageData)
	if !bytes.Equal(computedHash[:], proof.ImageMetadata.ImageHash) {
		result.IsLiveCapture = false
		result.Indicators = append(result.Indicators, GalleryIndicator{
			Type:        IndicatorInvalidOriginProof,
			Description: "Image hash does not match origin proof",
			Severity:    SeverityBlocking,
		})
		result.Blocked = true
		result.BlockReason = "Image hash mismatch"
		return true
	}
	return false
}

// validateEXIF validates EXIF metadata handling
func (v *GalleryPreventionValidator) validateEXIF(proof *CaptureOriginProof, result *GalleryDetectionResult) {
	if proof == nil || !v.config.ValidateEXIF {
		return
	}

	if proof.ImageMetadata.HasEXIF && !proof.ImageMetadata.EXIFStripped {
		result.Indicators = append(result.Indicators, GalleryIndicator{
			Type:        IndicatorNonCameraSource,
			Description: "EXIF data was not properly stripped",
			Severity:    SeveritySuspicious,
		})
		result.Confidence *= 0.8
	}
}

// determineFinalResult determines the final validation result
func (v *GalleryPreventionValidator) determineFinalResult(result *GalleryDetectionResult) {
	result.IsGalleryUpload = !result.IsLiveCapture
	if !result.Blocked && result.Confidence < 0.5 {
		result.IsGalleryUpload = true
		result.IsLiveCapture = false
		if v.config.StrictMode {
			result.Blocked = true
			result.BlockReason = "Low confidence in live capture origin"
		}
	}
}

// validateHardwareBinding validates platform-specific hardware binding
func (v *GalleryPreventionValidator) validateHardwareBinding(
	binding *HardwareBindingProof,
	result *GalleryDetectionResult,
) error {
	switch binding.Platform {
	case PlatformIOS:
		if binding.IOSProof == nil {
			result.Indicators = append(result.Indicators, GalleryIndicator{
				Type:        IndicatorMissingCameraData,
				Description: "Missing iOS camera proof data",
				Severity:    SeveritySuspicious,
			})
			result.Confidence *= 0.7
		} else if binding.IOSProof.PhotoID == "" {
			// Validate iOS-specific fields
			result.Indicators = append(result.Indicators, GalleryIndicator{
				Type:        IndicatorMissingCameraData,
				Description: "Missing photo ID from AVFoundation",
				Severity:    SeveritySuspicious,
			})
			result.Confidence *= 0.9
		}

	case PlatformAndroid:
		if binding.AndroidProof == nil {
			result.Indicators = append(result.Indicators, GalleryIndicator{
				Type:        IndicatorMissingCameraData,
				Description: "Missing Android camera proof data",
				Severity:    SeveritySuspicious,
			})
			result.Confidence *= 0.7
		} else if binding.AndroidProof.SensorTimestamp == 0 {
			// Validate Android-specific fields
			result.Indicators = append(result.Indicators, GalleryIndicator{
				Type:        IndicatorMissingCameraData,
				Description: "Missing sensor timestamp from Camera2",
				Severity:    SeveritySuspicious,
			})
			result.Confidence *= 0.9
		}

	default:
		return errors.New("unsupported platform")
	}

	return nil
}

// IsBlocked returns true if the capture should be blocked
func (r *GalleryDetectionResult) IsBlocked() bool {
	return r.Blocked || r.IsGalleryUpload
}

