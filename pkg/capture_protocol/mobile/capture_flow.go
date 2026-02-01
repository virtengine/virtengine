// Package mobile implements native mobile capture specifications for iOS and Android.
// VE-900/VE-4F: Capture flow - document + selfie + liveness workflow with compression
//
// This file implements the complete capture flow including document capture,
// selfie capture, and liveness detection with integrated compression.
package mobile

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"time"
)

// ============================================================================
// Capture Flow Configuration
// ============================================================================

// CaptureFlowConfig configures the complete capture flow
type CaptureFlowConfig struct {
	// Contract defines format and size requirements
	Contract CaptureContract `json:"contract"`

	// FlowType defines the capture flow type
	FlowType CaptureFlowType `json:"flow_type"`

	// Steps defines the capture steps required
	Steps []CaptureStep `json:"steps"`

	// Compression settings
	CompressionConfig CompressionConfig `json:"compression_config"`

	// Timeouts
	StepTimeoutSeconds   int `json:"step_timeout_seconds"`
	TotalTimeoutSeconds  int `json:"total_timeout_seconds"`

	// Retry settings
	MaxRetries       int `json:"max_retries"`
	RetryDelayMillis int `json:"retry_delay_millis"`

	// Callbacks enabled
	EnableQualityFeedback bool `json:"enable_quality_feedback"`
	EnableProgressUpdates bool `json:"enable_progress_updates"`
}

// CaptureFlowType represents the type of capture flow
type CaptureFlowType string

const (
	// FlowTypeIDVerification is full ID verification (document + selfie + liveness)
	FlowTypeIDVerification CaptureFlowType = "id_verification"

	// FlowTypeDocumentOnly is document-only verification
	FlowTypeDocumentOnly CaptureFlowType = "document_only"

	// FlowTypeSelfieOnly is selfie-only verification
	FlowTypeSelfieOnly CaptureFlowType = "selfie_only"

	// FlowTypeLivenessOnly is liveness-only verification
	FlowTypeLivenessOnly CaptureFlowType = "liveness_only"

	// FlowTypeReVerification is re-verification (just selfie + liveness)
	FlowTypeReVerification CaptureFlowType = "re_verification"
)

// CaptureStep represents a single step in the capture flow
type CaptureStep struct {
	// StepID is a unique identifier for this step
	StepID string `json:"step_id"`

	// StepType is the type of capture
	StepType CaptureStepType `json:"step_type"`

	// Required indicates if this step is mandatory
	Required bool `json:"required"`

	// Order is the sequence order
	Order int `json:"order"`

	// DocumentConfig for document steps
	DocumentConfig *DocumentStepConfig `json:"document_config,omitempty"`

	// SelfieConfig for selfie steps
	SelfieConfig *SelfieStepConfig `json:"selfie_config,omitempty"`

	// LivenessConfig for liveness steps
	LivenessConfig *LivenessStepConfig `json:"liveness_config,omitempty"`
}

// CaptureStepType represents types of capture steps
type CaptureStepType string

const (
	StepTypeDocumentFront CaptureStepType = "document_front"
	StepTypeDocumentBack  CaptureStepType = "document_back"
	StepTypeSelfie        CaptureStepType = "selfie"
	StepTypeLiveness      CaptureStepType = "liveness"
)

// DocumentStepConfig configures a document capture step
type DocumentStepConfig struct {
	DocumentType DocumentType `json:"document_type"`
	DocumentSide DocumentSide `json:"document_side"`
	AllowSkip    bool         `json:"allow_skip"` // e.g., passport has no back
}

// SelfieStepConfig configures a selfie capture step
type SelfieStepConfig struct {
	RequireFaceMatch bool `json:"require_face_match"` // Match against document
}

// LivenessStepConfig configures a liveness capture step
type LivenessStepConfig struct {
	Mode            LivenessMode            `json:"mode"`
	ChallengeTypes  []LivenessChallengeType `json:"challenge_types,omitempty"`
	MinChallenges   int                     `json:"min_challenges"`
	RequirePassive  bool                    `json:"require_passive"`
}

// DefaultCaptureFlowConfig returns default capture flow configuration
func DefaultCaptureFlowConfig(flowType CaptureFlowType) CaptureFlowConfig {
	config := CaptureFlowConfig{
		Contract:              DefaultCaptureContract(),
		FlowType:              flowType,
		CompressionConfig:     DefaultCompressionConfig(),
		StepTimeoutSeconds:    120,
		TotalTimeoutSeconds:   600,
		MaxRetries:            3,
		RetryDelayMillis:      1000,
		EnableQualityFeedback: true,
		EnableProgressUpdates: true,
	}

	switch flowType {
	case FlowTypeIDVerification:
		config.Steps = []CaptureStep{
			{
				StepID:   "doc_front",
				StepType: StepTypeDocumentFront,
				Required: true,
				Order:    1,
				DocumentConfig: &DocumentStepConfig{
					DocumentType: DocumentTypeIDCard,
					DocumentSide: DocumentSideFront,
				},
			},
			{
				StepID:   "doc_back",
				StepType: StepTypeDocumentBack,
				Required: false, // May be skipped for passport
				Order:    2,
				DocumentConfig: &DocumentStepConfig{
					DocumentType: DocumentTypeIDCard,
					DocumentSide: DocumentSideBack,
					AllowSkip:    true,
				},
			},
			{
				StepID:   "selfie",
				StepType: StepTypeSelfie,
				Required: true,
				Order:    3,
				SelfieConfig: &SelfieStepConfig{
					RequireFaceMatch: true,
				},
			},
			{
				StepID:   "liveness",
				StepType: StepTypeLiveness,
				Required: true,
				Order:    4,
				LivenessConfig: &LivenessStepConfig{
					Mode:           LivenessModeHybrid,
					MinChallenges:  1,
					RequirePassive: true,
				},
			},
		}
	case FlowTypeDocumentOnly:
		config.Steps = []CaptureStep{
			{
				StepID:   "doc_front",
				StepType: StepTypeDocumentFront,
				Required: true,
				Order:    1,
				DocumentConfig: &DocumentStepConfig{
					DocumentType: DocumentTypeIDCard,
					DocumentSide: DocumentSideFront,
				},
			},
			{
				StepID:   "doc_back",
				StepType: StepTypeDocumentBack,
				Required: false,
				Order:    2,
				DocumentConfig: &DocumentStepConfig{
					DocumentType: DocumentTypeIDCard,
					DocumentSide: DocumentSideBack,
					AllowSkip:    true,
				},
			},
		}
	case FlowTypeSelfieOnly:
		config.Steps = []CaptureStep{
			{
				StepID:   "selfie",
				StepType: StepTypeSelfie,
				Required: true,
				Order:    1,
			},
		}
	case FlowTypeLivenessOnly:
		config.Steps = []CaptureStep{
			{
				StepID:   "liveness",
				StepType: StepTypeLiveness,
				Required: true,
				Order:    1,
				LivenessConfig: &LivenessStepConfig{
					Mode:          LivenessModeActive,
					MinChallenges: 2,
				},
			},
		}
	case FlowTypeReVerification:
		config.Steps = []CaptureStep{
			{
				StepID:   "selfie",
				StepType: StepTypeSelfie,
				Required: true,
				Order:    1,
				SelfieConfig: &SelfieStepConfig{
					RequireFaceMatch: false, // Match against stored
				},
			},
			{
				StepID:   "liveness",
				StepType: StepTypeLiveness,
				Required: true,
				Order:    2,
				LivenessConfig: &LivenessStepConfig{
					Mode:          LivenessModeHybrid,
					MinChallenges: 1,
				},
			},
		}
	}

	return config
}

// ============================================================================
// Compression Configuration
// ============================================================================

// CompressionConfig configures image/video compression
type CompressionConfig struct {
	// Image compression
	EnableImageCompression bool    `json:"enable_image_compression"`
	TargetImageSizeBytes   int64   `json:"target_image_size_bytes"`
	MinJPEGQuality         int     `json:"min_jpeg_quality"`
	MaxJPEGQuality         int     `json:"max_jpeg_quality"`
	MaxCompressionPasses   int     `json:"max_compression_passes"`

	// Video compression
	EnableVideoCompression bool  `json:"enable_video_compression"`
	TargetVideoBitrate     int   `json:"target_video_bitrate"` // kbps
	MaxVideoSizeBytes      int64 `json:"max_video_size_bytes"`

	// Metadata handling
	StripAllMetadata       bool `json:"strip_all_metadata"`
	PreserveOrientation    bool `json:"preserve_orientation"`
}

// DefaultCompressionConfig returns default compression settings
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		EnableImageCompression: true,
		TargetImageSizeBytes:   2 * 1024 * 1024, // 2MB
		MinJPEGQuality:         60,
		MaxJPEGQuality:         95,
		MaxCompressionPasses:   5,

		EnableVideoCompression: true,
		TargetVideoBitrate:     2000, // 2 Mbps
		MaxVideoSizeBytes:      20 * 1024 * 1024, // 20MB

		StripAllMetadata:    true,
		PreserveOrientation: true,
	}
}

// ============================================================================
// Capture Flow State Machine
// ============================================================================

// CaptureFlowState represents the current state of the capture flow
type CaptureFlowState string

const (
	FlowStateInitializing CaptureFlowState = "initializing"
	FlowStateReady        CaptureFlowState = "ready"
	FlowStateCapturing    CaptureFlowState = "capturing"
	FlowStateProcessing   CaptureFlowState = "processing"
	FlowStateCompressing  CaptureFlowState = "compressing"
	FlowStateUploading    CaptureFlowState = "uploading"
	FlowStateComplete     CaptureFlowState = "complete"
	FlowStateFailed       CaptureFlowState = "failed"
	FlowStateCancelled    CaptureFlowState = "cancelled"
)

// CaptureFlowProgress represents progress through the capture flow
type CaptureFlowProgress struct {
	// State is the current flow state
	State CaptureFlowState `json:"state"`

	// CurrentStep is the current step being executed
	CurrentStep *CaptureStep `json:"current_step,omitempty"`

	// CompletedSteps is the number of completed steps
	CompletedSteps int `json:"completed_steps"`

	// TotalSteps is the total number of steps
	TotalSteps int `json:"total_steps"`

	// ProgressPercent is the overall progress (0-100)
	ProgressPercent int `json:"progress_percent"`

	// StepProgressPercent is the current step progress (0-100)
	StepProgressPercent int `json:"step_progress_percent"`

	// Message is a user-facing status message
	Message string `json:"message"`

	// Error contains any error that occurred
	Error error `json:"error,omitempty"`

	// StartedAt is when the flow started
	StartedAt time.Time `json:"started_at"`

	// CurrentStepStartedAt is when the current step started
	CurrentStepStartedAt time.Time `json:"current_step_started_at,omitempty"`
}

// ============================================================================
// Capture Flow Result
// ============================================================================

// CaptureFlowResult represents the complete result of a capture flow
type CaptureFlowResult struct {
	// Success indicates if the flow completed successfully
	Success bool `json:"success"`

	// FlowID is the unique flow identifier
	FlowID string `json:"flow_id"`

	// FlowType is the type of flow executed
	FlowType CaptureFlowType `json:"flow_type"`

	// StepResults contains results for each step
	StepResults []StepResult `json:"step_results"`

	// CompressedPayloads contains compressed capture data
	CompressedPayloads []CompressedPayload `json:"compressed_payloads"`

	// TotalDuration is the total flow duration
	TotalDuration time.Duration `json:"total_duration"`

	// StartedAt is when the flow started
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when the flow completed
	CompletedAt time.Time `json:"completed_at"`

	// Error contains any error
	Error error `json:"error,omitempty"`

	// DeviceFingerprint is the device fingerprint
	DeviceFingerprint DeviceFingerprint `json:"device_fingerprint"`

	// Metadata contains flow metadata
	Metadata CaptureFlowMetadata `json:"metadata"`
}

// StepResult represents the result of a single capture step
type StepResult struct {
	// StepID is the step identifier
	StepID string `json:"step_id"`

	// StepType is the step type
	StepType CaptureStepType `json:"step_type"`

	// Success indicates if the step succeeded
	Success bool `json:"success"`

	// Skipped indicates if the step was skipped
	Skipped bool `json:"skipped"`

	// CaptureResult is the capture result (for image steps)
	CaptureResult *CaptureResult `json:"capture_result,omitempty"`

	// LivenessResult is the liveness result (for liveness steps)
	LivenessResult *LivenessResult `json:"liveness_result,omitempty"`

	// Duration is the step duration
	Duration time.Duration `json:"duration"`

	// Attempts is the number of attempts
	Attempts int `json:"attempts"`

	// Error contains any error
	Error error `json:"error,omitempty"`
}

// CompressedPayload represents a compressed capture payload
type CompressedPayload struct {
	// StepID is the source step
	StepID string `json:"step_id"`

	// OriginalSize is the original size in bytes
	OriginalSize int64 `json:"original_size"`

	// CompressedSize is the compressed size in bytes
	CompressedSize int64 `json:"compressed_size"`

	// CompressionRatio is the compression ratio
	CompressionRatio float64 `json:"compression_ratio"`

	// Format is the output format
	Format string `json:"format"`

	// Quality is the compression quality used
	Quality int `json:"quality"`

	// ImageHash is SHA256 of the compressed image
	ImageHash []byte `json:"image_hash"`

	// Data is the compressed image data
	Data []byte `json:"data"`

	// Metadata contains compression metadata
	Metadata CompressionMetadata `json:"metadata"`
}

// CompressionMetadata contains metadata about the compression
type CompressionMetadata struct {
	// OriginalWidth is the original image width
	OriginalWidth int `json:"original_width"`

	// OriginalHeight is the original image height
	OriginalHeight int `json:"original_height"`

	// CompressedWidth is the compressed width (if resized)
	CompressedWidth int `json:"compressed_width"`

	// CompressedHeight is the compressed height (if resized)
	CompressedHeight int `json:"compressed_height"`

	// CompressionPasses is the number of compression passes
	CompressionPasses int `json:"compression_passes"`

	// MetadataStripped indicates if metadata was stripped
	MetadataStripped bool `json:"metadata_stripped"`

	// CompressionDuration is how long compression took
	CompressionDuration time.Duration `json:"compression_duration"`
}

// CaptureFlowMetadata contains flow-level metadata
type CaptureFlowMetadata struct {
	// ClientID is the approved client ID
	ClientID string `json:"client_id"`

	// ClientVersion is the client version
	ClientVersion string `json:"client_version"`

	// SessionID is the capture session ID
	SessionID string `json:"session_id"`

	// Platform is the device platform
	Platform Platform `json:"platform"`

	// OSVersion is the OS version
	OSVersion string `json:"os_version"`

	// SDKVersion is the SDK version
	SDKVersion string `json:"sdk_version"`

	// Timestamp is when the flow started
	Timestamp time.Time `json:"timestamp"`

	// Locale is the device locale
	Locale string `json:"locale"`
}

// ============================================================================
// Image Compressor
// ============================================================================

// ImageCompressor handles image compression
type ImageCompressor struct {
	config CompressionConfig
}

// NewImageCompressor creates a new image compressor
func NewImageCompressor(config CompressionConfig) *ImageCompressor {
	return &ImageCompressor{config: config}
}

// CompressImage compresses an image to target size
func (c *ImageCompressor) CompressImage(
	ctx context.Context,
	imageData []byte,
	targetSize int64,
) (*CompressedPayload, error) {
	startTime := time.Now()

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Try compression at different quality levels
	quality := c.config.MaxJPEGQuality
	var compressedData []byte
	passes := 0

	for passes < c.config.MaxCompressionPasses {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		passes++

		// Encode with current quality
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, fmt.Errorf("failed to encode image: %w", err)
		}

		compressedData = buf.Bytes()

		// Check if size is acceptable
		if int64(len(compressedData)) <= targetSize {
			break
		}

		// Reduce quality for next pass
		quality -= 10
		if quality < c.config.MinJPEGQuality {
			quality = c.config.MinJPEGQuality
			break
		}
	}

	// Compute hash
	hash := sha256.Sum256(compressedData)

	compressionRatio := float64(len(imageData)) / float64(len(compressedData))

	return &CompressedPayload{
		OriginalSize:     int64(len(imageData)),
		CompressedSize:   int64(len(compressedData)),
		CompressionRatio: compressionRatio,
		Format:           format,
		Quality:          quality,
		ImageHash:        hash[:],
		Data:             compressedData,
		Metadata: CompressionMetadata{
			OriginalWidth:       originalWidth,
			OriginalHeight:      originalHeight,
			CompressedWidth:     originalWidth,
			CompressedHeight:    originalHeight,
			CompressionPasses:   passes,
			MetadataStripped:    c.config.StripAllMetadata,
			CompressionDuration: time.Since(startTime),
		},
	}, nil
}

// ============================================================================
// Capture Flow Executor
// ============================================================================

// CaptureFlowExecutor executes capture flows
type CaptureFlowExecutor struct {
	config            CaptureFlowConfig
	library           MobileCaptureLibrary
	compressor        *ImageCompressor
	clientKeyProvider MobileClientKeyProvider
	userKeyProvider   MobileUserKeyProvider
	progressChan      chan CaptureFlowProgress
}

// NewCaptureFlowExecutor creates a new capture flow executor
func NewCaptureFlowExecutor(
	config CaptureFlowConfig,
	library MobileCaptureLibrary,
	clientKeyProvider MobileClientKeyProvider,
	userKeyProvider MobileUserKeyProvider,
) *CaptureFlowExecutor {
	return &CaptureFlowExecutor{
		config:            config,
		library:           library,
		compressor:        NewImageCompressor(config.CompressionConfig),
		clientKeyProvider: clientKeyProvider,
		userKeyProvider:   userKeyProvider,
		progressChan:      make(chan CaptureFlowProgress, 10),
	}
}

// Execute executes the capture flow
func (e *CaptureFlowExecutor) Execute(ctx context.Context) (*CaptureFlowResult, error) {
	startTime := time.Now()

	// Generate flow ID
	flowID, err := generateFlowID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate flow ID: %w", err)
	}

	// Get device fingerprint
	fingerprint, err := e.library.GetDeviceFingerprint()
	if err != nil {
		return nil, fmt.Errorf("failed to get device fingerprint: %w", err)
	}

	result := &CaptureFlowResult{
		FlowID:            flowID,
		FlowType:          e.config.FlowType,
		StartedAt:         startTime,
		DeviceFingerprint: *fingerprint,
		StepResults:       make([]StepResult, 0, len(e.config.Steps)),
		CompressedPayloads: make([]CompressedPayload, 0),
		Metadata: CaptureFlowMetadata{
			ClientID:      e.clientKeyProvider.GetClientID(),
			ClientVersion: e.clientKeyProvider.GetClientVersion(),
			SessionID:     flowID,
			Timestamp:     startTime,
			SDKVersion:    CurrentSDKInfo().SDKVersion,
		},
	}

	// Create timeout context
	timeout := time.Duration(e.config.TotalTimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute each step
	for _, step := range e.config.Steps {
		select {
		case <-ctx.Done():
			result.Success = false
			result.Error = ctx.Err()
			result.CompletedAt = time.Now()
			result.TotalDuration = time.Since(startTime)
			return result, nil
		default:
		}

		stepResult, err := e.executeStep(ctx, step)
		if err != nil && step.Required {
			result.Success = false
			result.Error = err
			result.StepResults = append(result.StepResults, *stepResult)
			result.CompletedAt = time.Now()
			result.TotalDuration = time.Since(startTime)
			return result, nil
		}

		result.StepResults = append(result.StepResults, *stepResult)

		// Compress capture if successful
		if stepResult.Success && stepResult.CaptureResult != nil {
			compressed, err := e.compressor.CompressImage(
				ctx,
				stepResult.CaptureResult.ImageData,
				e.config.CompressionConfig.TargetImageSizeBytes,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to compress image: %w", err)
			}
			compressed.StepID = step.StepID
			result.CompressedPayloads = append(result.CompressedPayloads, *compressed)
		}
	}

	result.Success = true
	result.CompletedAt = time.Now()
	result.TotalDuration = time.Since(startTime)

	return result, nil
}

// executeStep executes a single capture step
func (e *CaptureFlowExecutor) executeStep(ctx context.Context, step CaptureStep) (*StepResult, error) {
	startTime := time.Now()

	result := &StepResult{
		StepID:   step.StepID,
		StepType: step.StepType,
		Attempts: 0,
	}

	// Create step timeout context
	stepTimeout := time.Duration(e.config.StepTimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()

	var lastErr error

	for attempt := 0; attempt < e.config.MaxRetries; attempt++ {
		result.Attempts++

		select {
		case <-ctx.Done():
			result.Success = false
			result.Error = ctx.Err()
			result.Duration = time.Since(startTime)
			return result, ctx.Err()
		default:
		}

		switch step.StepType {
		case StepTypeDocumentFront, StepTypeDocumentBack:
			captureResult, err := e.captureDocument(ctx, step)
			if err != nil {
				lastErr = err
				time.Sleep(time.Duration(e.config.RetryDelayMillis) * time.Millisecond)
				continue
			}
			result.CaptureResult = captureResult
			result.Success = true

		case StepTypeSelfie:
			captureResult, err := e.captureSelfie(ctx, step)
			if err != nil {
				lastErr = err
				time.Sleep(time.Duration(e.config.RetryDelayMillis) * time.Millisecond)
				continue
			}
			result.CaptureResult = captureResult
			result.Success = true

		case StepTypeLiveness:
			livenessResult, err := e.captureLiveness(ctx, step)
			if err != nil {
				lastErr = err
				time.Sleep(time.Duration(e.config.RetryDelayMillis) * time.Millisecond)
				continue
			}
			result.LivenessResult = livenessResult
			result.Success = livenessResult.Passed
		}

		if result.Success {
			break
		}
	}

	if !result.Success {
		result.Error = lastErr
	}

	result.Duration = time.Since(startTime)
	return result, lastErr
}

// captureDocument captures a document image
func (e *CaptureFlowExecutor) captureDocument(ctx context.Context, step CaptureStep) (*CaptureResult, error) {
	if step.DocumentConfig == nil {
		return nil, fmt.Errorf("document config required for document step")
	}

	sessionConfig := CaptureSessionConfig{
		SessionID:     fmt.Sprintf("%s-%d", step.StepID, time.Now().UnixNano()),
		CaptureType:   CaptureTypeDocument,
		DocumentType:  step.DocumentConfig.DocumentType,
		DocumentSide:  step.DocumentConfig.DocumentSide,
		CameraConfig:  DefaultDocumentCameraConfig(),
		QualityConfig: DefaultDocumentQualityConfig(),
	}

	session, err := e.library.CreateCaptureSession(sessionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create capture session: %w", err)
	}

	if err := session.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start capture session: %w", err)
	}

	// Wait for capture
	if err := session.Capture(); err != nil {
		return nil, fmt.Errorf("failed to capture: %w", err)
	}

	return session.GetResult()
}

// captureSelfie captures a selfie image
func (e *CaptureFlowExecutor) captureSelfie(ctx context.Context, step CaptureStep) (*CaptureResult, error) {
	sessionConfig := CaptureSessionConfig{
		SessionID:     fmt.Sprintf("%s-%d", step.StepID, time.Now().UnixNano()),
		CaptureType:   CaptureTypeSelfie,
		CameraConfig:  DefaultSelfieCameraConfig(),
		QualityConfig: DefaultSelfieQualityConfig(),
	}

	session, err := e.library.CreateCaptureSession(sessionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create capture session: %w", err)
	}

	if err := session.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start capture session: %w", err)
	}

	if err := session.Capture(); err != nil {
		return nil, fmt.Errorf("failed to capture: %w", err)
	}

	return session.GetResult()
}

// captureLiveness performs liveness detection
func (e *CaptureFlowExecutor) captureLiveness(ctx context.Context, step CaptureStep) (*LivenessResult, error) {
	livenessConfig := DefaultLivenessConfig()
	if step.LivenessConfig != nil {
		livenessConfig.Mode = step.LivenessConfig.Mode
		if len(step.LivenessConfig.ChallengeTypes) > 0 {
			livenessConfig.ChallengeTypes = step.LivenessConfig.ChallengeTypes
		}
		livenessConfig.MinChallenges = step.LivenessConfig.MinChallenges
	}

	sessionConfig := CaptureSessionConfig{
		SessionID:      fmt.Sprintf("%s-%d", step.StepID, time.Now().UnixNano()),
		CaptureType:    CaptureTypeSelfie,
		CameraConfig:   DefaultSelfieCameraConfig(),
		QualityConfig:  DefaultSelfieQualityConfig(),
		LivenessConfig: &livenessConfig,
	}

	session, err := e.library.CreateCaptureSession(sessionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create capture session: %w", err)
	}

	if err := session.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start capture session: %w", err)
	}

	if err := session.Capture(); err != nil {
		return nil, fmt.Errorf("failed to capture: %w", err)
	}

	result, err := session.GetResult()
	if err != nil {
		return nil, err
	}

	return result.LivenessResult, nil
}

// Progress returns the progress channel
func (e *CaptureFlowExecutor) Progress() <-chan CaptureFlowProgress {
	return e.progressChan
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateFlowID generates a unique flow ID
func generateFlowID() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return fmt.Sprintf("flow-%x", b), nil
}

// generateSalt generates a cryptographic salt
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}

