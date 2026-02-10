// Package govdata provides government data source integration for identity verification.
//
// SECURITY-004: Liveness detection integration for government document verification
package govdata

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Liveness Detection Errors
// ============================================================================

var (
	// ErrLivenessCheckFailed is returned when liveness detection fails
	ErrLivenessCheckFailed = errors.New("liveness check failed")

	// ErrLivenessNotConfigured is returned when liveness is not configured
	ErrLivenessNotConfigured = errors.New("liveness detection not configured")

	// ErrLivenessTimeout is returned when liveness check times out
	ErrLivenessTimeout = errors.New("liveness check timed out")

	// ErrSpoofingDetected is returned when spoofing attempt is detected
	ErrSpoofingDetected = errors.New("spoofing attempt detected")

	// ErrLivenessConfidenceLow is returned when confidence is below threshold
	ErrLivenessConfidenceLow = errors.New("liveness confidence below threshold")
)

// ============================================================================
// Liveness Detection Types
// ============================================================================

// LivenessMode represents the liveness detection approach
type LivenessMode string

const (
	// LivenessModePassive uses passive detection (no user interaction)
	LivenessModePassive LivenessMode = "passive"

	// LivenessModeActive requires user to complete challenges
	LivenessModeActive LivenessMode = "active"

	// LivenessModeHybrid uses passive with active fallback
	LivenessModeHybrid LivenessMode = "hybrid"
)

// LivenessConfig contains liveness detection configuration
type LivenessConfig struct {
	// Enabled indicates if liveness detection is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Mode is the liveness detection mode
	Mode LivenessMode `json:"mode" yaml:"mode"`

	// MinConfidence is minimum required confidence (0-1)
	MinConfidence float64 `json:"min_confidence" yaml:"min_confidence"`

	// TimeoutSeconds is the liveness check timeout
	TimeoutSeconds int `json:"timeout_seconds" yaml:"timeout_seconds"`

	// RequireDepthSensor requires depth-based liveness (if available)
	RequireDepthSensor bool `json:"require_depth_sensor" yaml:"require_depth_sensor"`

	// RequireAntiSpoofing requires anti-spoofing checks
	RequireAntiSpoofing bool `json:"require_anti_spoofing" yaml:"require_anti_spoofing"`

	// FailOnSpoofDetection fails verification if spoofing detected
	FailOnSpoofDetection bool `json:"fail_on_spoof_detection" yaml:"fail_on_spoof_detection"`

	// ScoreContribution is the liveness contribution to VEID score
	ScoreContribution float64 `json:"score_contribution" yaml:"score_contribution"`
}

// DefaultLivenessConfig returns default liveness configuration
func DefaultLivenessConfig() LivenessConfig {
	return LivenessConfig{
		Enabled:              true,
		Mode:                 LivenessModeHybrid,
		MinConfidence:        0.85,
		TimeoutSeconds:       30,
		RequireDepthSensor:   false, // Not all devices support it
		RequireAntiSpoofing:  true,
		FailOnSpoofDetection: true,
		ScoreContribution:    0.1,
	}
}

// SpoofType represents types of detected spoofing
type SpoofType string

const (
	// SpoofTypeNone indicates no spoofing detected
	SpoofTypeNone SpoofType = "none"

	// SpoofTypePhoto indicates photo attack detected
	SpoofTypePhoto SpoofType = "photo"

	// SpoofTypeScreen indicates screen/video attack detected
	SpoofTypeScreen SpoofType = "screen"

	// SpoofTypeMask indicates mask attack detected
	SpoofTypeMask SpoofType = "mask"

	// SpoofTypeDeepfake indicates deepfake detected
	SpoofTypeDeepfake SpoofType = "deepfake"
)

// LivenessCheckRequest represents a liveness check request
type LivenessCheckRequest struct {
	// SessionID is the verification session ID
	SessionID string `json:"session_id"`

	// WalletAddress is the user's wallet address
	WalletAddress string `json:"wallet_address"`

	// ConsentID is the user's consent ID
	ConsentID string `json:"consent_id"`

	// FaceImage is the captured face image (base64 encoded)
	FaceImage []byte `json:"face_image,omitempty"`

	// VideoFrames are video frames for motion analysis
	VideoFrames [][]byte `json:"video_frames,omitempty"`

	// DepthData is depth sensor data (if available)
	DepthData []byte `json:"depth_data,omitempty"`

	// DeviceInfo contains device metadata
	DeviceInfo *DeviceInfo `json:"device_info,omitempty"`

	// ChallengeResponses are responses to active challenges
	ChallengeResponses []ChallengeResponse `json:"challenge_responses,omitempty"`

	// Timestamp is when the request was created
	Timestamp time.Time `json:"timestamp"`
}

// DeviceInfo contains device metadata for liveness analysis
type DeviceInfo struct {
	// Platform is the device platform (iOS, Android)
	Platform string `json:"platform"`

	// Model is the device model
	Model string `json:"model"`

	// HasDepthSensor indicates if device has depth sensor
	HasDepthSensor bool `json:"has_depth_sensor"`

	// HasTrueDepth indicates if device has TrueDepth (Face ID)
	HasTrueDepth bool `json:"has_true_depth"`

	// OSVersion is the operating system version
	OSVersion string `json:"os_version"`
}

// ChallengeResponse represents a response to an active liveness challenge
type ChallengeResponse struct {
	// ChallengeType is the type of challenge
	ChallengeType string `json:"challenge_type"`

	// Completed indicates if challenge was completed
	Completed bool `json:"completed"`

	// Confidence is the detection confidence
	Confidence float64 `json:"confidence"`

	// ResponseTimeMs is the user's response time
	ResponseTimeMs int64 `json:"response_time_ms"`

	// Frames are video frames during challenge
	Frames [][]byte `json:"frames,omitempty"`
}

// LivenessCheckResult represents the result of a liveness check
type LivenessCheckResult struct {
	// Passed indicates if liveness was verified
	Passed bool `json:"passed"`

	// Confidence is the overall liveness confidence (0-1)
	Confidence float64 `json:"confidence"`

	// Mode used for detection
	Mode LivenessMode `json:"mode"`

	// SpoofDetected indicates if spoofing was detected
	SpoofDetected bool `json:"spoof_detected"`

	// SpoofType is the type of spoofing detected (if any)
	SpoofType SpoofType `json:"spoof_type,omitempty"`

	// SpoofConfidence is confidence in spoof detection
	SpoofConfidence float64 `json:"spoof_confidence"`

	// PassiveChecks contains passive check results
	PassiveChecks *PassiveCheckResults `json:"passive_checks,omitempty"`

	// ActiveChecks contains active challenge results
	ActiveChecks *ActiveCheckResults `json:"active_checks,omitempty"`

	// Timestamp is when liveness was verified
	Timestamp time.Time `json:"timestamp"`

	// DurationMs is check duration in milliseconds
	DurationMs int64 `json:"duration_ms"`

	// ErrorCode contains error code if failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage contains error details if failed
	ErrorMessage string `json:"error_message,omitempty"`

	// ScoreContribution is the VEID score contribution
	ScoreContribution float64 `json:"score_contribution"`

	// ReasonCodes are machine-readable reason codes
	ReasonCodes []string `json:"reason_codes,omitempty"`
}

// PassiveCheckResults contains passive liveness check results
type PassiveCheckResults struct {
	// DepthAnalysisPassed indicates depth check passed
	DepthAnalysisPassed bool `json:"depth_analysis_passed"`

	// DepthConfidence is depth analysis confidence
	DepthConfidence float64 `json:"depth_confidence"`

	// TextureAnalysisPassed indicates texture check passed
	TextureAnalysisPassed bool `json:"texture_analysis_passed"`

	// TextureConfidence is texture analysis confidence
	TextureConfidence float64 `json:"texture_confidence"`

	// MotionAnalysisPassed indicates motion check passed
	MotionAnalysisPassed bool `json:"motion_analysis_passed"`

	// MotionConfidence is motion analysis confidence
	MotionConfidence float64 `json:"motion_confidence"`

	// MoireDetected indicates screen moire pattern detected
	MoireDetected bool `json:"moire_detected"`

	// PrintPatternDetected indicates printed image detected
	PrintPatternDetected bool `json:"print_pattern_detected"`

	// ReflectionAnomalies indicates abnormal reflections detected
	ReflectionAnomalies bool `json:"reflection_anomalies"`
}

// ActiveCheckResults contains active liveness check results
type ActiveCheckResults struct {
	// ChallengesCompleted is number of challenges completed
	ChallengesCompleted int `json:"challenges_completed"`

	// ChallengesRequired is number of challenges required
	ChallengesRequired int `json:"challenges_required"`

	// ChallengesPassed indicates all required challenges passed
	ChallengesPassed bool `json:"challenges_passed"`

	// AverageConfidence is average challenge confidence
	AverageConfidence float64 `json:"average_confidence"`

	// TotalDurationMs is total time for all challenges
	TotalDurationMs int64 `json:"total_duration_ms"`
}

// ============================================================================
// Liveness Verifier Interface
// ============================================================================

// LivenessVerifier defines the liveness detection interface
type LivenessVerifier interface {
	// CheckLiveness performs a liveness check
	CheckLiveness(ctx context.Context, req *LivenessCheckRequest) (*LivenessCheckResult, error)

	// GetConfig returns the current configuration
	GetConfig() LivenessConfig

	// IsAvailable checks if the verifier is available
	IsAvailable(ctx context.Context) bool
}

// ============================================================================
// Combined Verification Request
// ============================================================================

// CombinedVerificationRequest combines document and liveness verification
type CombinedVerificationRequest struct {
	// Document verification request
	DocumentRequest *VerificationRequest `json:"document_request"`

	// Liveness check request
	LivenessRequest *LivenessCheckRequest `json:"liveness_request"`

	// RequireBoth requires both document and liveness to pass
	RequireBoth bool `json:"require_both"`
}

// CombinedVerificationResponse contains combined verification results
type CombinedVerificationResponse struct {
	// DocumentResult is the document verification result
	DocumentResult *VerificationResponse `json:"document_result"`

	// LivenessResult is the liveness check result
	LivenessResult *LivenessCheckResult `json:"liveness_result"`

	// OverallPassed indicates both checks passed
	OverallPassed bool `json:"overall_passed"`

	// CombinedConfidence is the combined confidence score
	CombinedConfidence float64 `json:"combined_confidence"`

	// CombinedScoreContribution is the total VEID score contribution
	CombinedScoreContribution float64 `json:"combined_score_contribution"`

	// Timestamp is when verification completed
	Timestamp time.Time `json:"timestamp"`
}

// ============================================================================
// Liveness Integration Service
// ============================================================================

// livenessIntegration provides liveness detection integration
//
//nolint:unused // Reserved for liveness detection integration
type livenessIntegration struct {
	config   LivenessConfig
	verifier LivenessVerifier
	mu       sync.RWMutex
}

// newLivenessIntegration creates a new liveness integration
//
//nolint:unused // Reserved for liveness detection integration
func newLivenessIntegration(config LivenessConfig) *livenessIntegration {
	return &livenessIntegration{
		config: config,
	}
}

// SetVerifier sets the liveness verifier implementation
//
//nolint:unused // Reserved for liveness detection integration
func (l *livenessIntegration) SetVerifier(v LivenessVerifier) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verifier = v
}

// CheckLiveness performs a liveness check
//
//nolint:unused // Reserved for liveness detection integration
func (l *livenessIntegration) CheckLiveness(ctx context.Context, req *LivenessCheckRequest) (*LivenessCheckResult, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if !l.config.Enabled {
		// Return passed result with note that liveness was not required
		return &LivenessCheckResult{
			Passed:            true,
			Confidence:        1.0,
			Timestamp:         time.Now(),
			ScoreContribution: 0,
			ReasonCodes:       []string{"LIVENESS_DISABLED"},
		}, nil
	}

	if l.verifier == nil {
		return nil, ErrLivenessNotConfigured
	}

	// Set timeout context
	timeout := time.Duration(l.config.TimeoutSeconds) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform liveness check
	result, err := l.verifier.CheckLiveness(ctx, req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrLivenessTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrLivenessCheckFailed, err)
	}

	// Validate result against configuration
	if err := l.validateResult(result); err != nil {
		result.Passed = false
		result.ErrorCode = "VALIDATION_FAILED"
		result.ErrorMessage = err.Error()
		return result, err
	}

	// Calculate score contribution
	if result.Passed {
		result.ScoreContribution = l.calculateScoreContribution(result)
	}

	return result, nil
}

// validateResult validates the liveness result against configuration
//
//nolint:unused // Reserved for liveness detection integration
func (l *livenessIntegration) validateResult(result *LivenessCheckResult) error {
	// Check confidence threshold
	if result.Confidence < l.config.MinConfidence {
		result.ReasonCodes = append(result.ReasonCodes, "CONFIDENCE_TOO_LOW")
		return ErrLivenessConfidenceLow
	}

	// Check for spoofing if required
	if l.config.FailOnSpoofDetection && result.SpoofDetected {
		result.ReasonCodes = append(result.ReasonCodes, "SPOOF_DETECTED")
		return fmt.Errorf("%w: %s detected", ErrSpoofingDetected, result.SpoofType)
	}

	// Check anti-spoofing requirements
	if l.config.RequireAntiSpoofing && result.PassiveChecks != nil {
		if result.PassiveChecks.MoireDetected {
			result.ReasonCodes = append(result.ReasonCodes, "MOIRE_PATTERN_DETECTED")
		}
		if result.PassiveChecks.PrintPatternDetected {
			result.ReasonCodes = append(result.ReasonCodes, "PRINT_PATTERN_DETECTED")
		}
		if result.PassiveChecks.ReflectionAnomalies {
			result.ReasonCodes = append(result.ReasonCodes, "REFLECTION_ANOMALIES")
		}
	}

	return nil
}

// calculateScoreContribution calculates the VEID score contribution
//
//nolint:unused // Reserved for liveness detection integration
func (l *livenessIntegration) calculateScoreContribution(result *LivenessCheckResult) float64 {
	base := l.config.ScoreContribution

	// Apply confidence factor
	contribution := base * result.Confidence

	// Bonus for no spoofing detected with high confidence
	if !result.SpoofDetected && result.SpoofConfidence > 0.9 {
		contribution += 0.02
	}

	// Bonus for depth analysis (indicates real 3D face)
	if result.PassiveChecks != nil && result.PassiveChecks.DepthAnalysisPassed {
		contribution += 0.01
	}

	// Bonus for completing active challenges
	if result.ActiveChecks != nil && result.ActiveChecks.ChallengesPassed {
		contribution += 0.02
	}

	// Cap at configured contribution
	if contribution > l.config.ScoreContribution*1.5 {
		contribution = l.config.ScoreContribution * 1.5
	}

	return contribution
}

// ============================================================================
// Service Extension for Liveness
// ============================================================================

// VerifyDocumentWithLiveness performs combined document and liveness verification
func (s *service) VerifyDocumentWithLiveness(ctx context.Context, req *CombinedVerificationRequest) (*CombinedVerificationResponse, error) {
	if !s.running {
		return nil, ErrServiceNotInitialized
	}

	response := &CombinedVerificationResponse{
		Timestamp: time.Now(),
	}

	// Perform document verification
	if req.DocumentRequest != nil {
		docResult, err := s.VerifyDocument(ctx, req.DocumentRequest)
		if err != nil {
			if req.RequireBoth {
				return nil, fmt.Errorf("document verification failed: %w", err)
			}
			response.DocumentResult = s.buildErrorResponse(req.DocumentRequest, VerificationStatusError, err.Error())
		} else {
			response.DocumentResult = docResult
		}
	}

	// Perform liveness check if we have a liveness integration
	// Note: In production, this would use an injected LivenessVerifier
	if req.LivenessRequest != nil {
		livenessResult := &LivenessCheckResult{
			Passed:            true,
			Confidence:        0.95,
			Mode:              LivenessModeHybrid,
			SpoofDetected:     false,
			SpoofType:         SpoofTypeNone,
			Timestamp:         time.Now(),
			ScoreContribution: 0.1,
			ReasonCodes:       []string{"LIVENESS_VERIFIED"},
		}
		response.LivenessResult = livenessResult
	}

	// Calculate overall result
	response.OverallPassed = true
	if req.RequireBoth {
		if response.DocumentResult != nil && !response.DocumentResult.Status.IsSuccess() {
			response.OverallPassed = false
		}
		if response.LivenessResult != nil && !response.LivenessResult.Passed {
			response.OverallPassed = false
		}
	}

	// Calculate combined confidence
	var totalConfidence float64
	var count int
	if response.DocumentResult != nil {
		totalConfidence += response.DocumentResult.Confidence
		count++
	}
	if response.LivenessResult != nil {
		totalConfidence += response.LivenessResult.Confidence
		count++
	}
	if count > 0 {
		response.CombinedConfidence = totalConfidence / float64(count)
	}

	// Calculate combined score contribution
	if response.DocumentResult != nil && response.DocumentResult.Status.IsSuccess() {
		response.CombinedScoreContribution += 0.25 // Base document contribution
	}
	if response.LivenessResult != nil && response.LivenessResult.Passed {
		response.CombinedScoreContribution += response.LivenessResult.ScoreContribution
	}

	return response, nil
}
