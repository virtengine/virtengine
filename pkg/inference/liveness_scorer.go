package inference

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"
)

// ============================================================================
// Liveness Scorer
// ============================================================================

// LivenessScorer integrates liveness detection results into the scoring pipeline.
// It interfaces with the Python liveness detection pipeline via gRPC sidecar
// or provides deterministic stub scores when unavailable.
type LivenessScorer struct {
	config LivenessScorerConfig
	client LivenessClient

	mu         sync.RWMutex
	isHealthy  bool
	checkCount uint64
	errorCount uint64
}

// LivenessScorerConfig contains configuration for liveness scoring
type LivenessScorerConfig struct {
	// SidecarAddress is the gRPC address of the liveness detection sidecar
	SidecarAddress string

	// Timeout for liveness check operations
	Timeout time.Duration

	// MinLivenessScore is the minimum score to pass liveness check
	MinLivenessScore float32

	// MinConfidence is the minimum confidence for valid checks
	MinConfidence float32

	// RequiredChallenges lists challenges that must pass
	RequiredChallenges []ChallengeType

	// OptionalChallenges lists challenges that contribute to score
	OptionalChallenges []ChallengeType

	// UseFallbackOnError returns stub values on errors
	UseFallbackOnError bool

	// StrictMode fails if any required challenge fails
	StrictMode bool

	// MaxRetries for transient failures
	MaxRetries int

	// RetryDelay between retry attempts
	RetryDelay time.Duration
}

// ChallengeType represents a liveness challenge type
type ChallengeType string

const (
	ChallengeBlink         ChallengeType = "blink"
	ChallengeSmile         ChallengeType = "smile"
	ChallengeHeadTurnLeft  ChallengeType = "head_turn_left"
	ChallengeHeadTurnRight ChallengeType = "head_turn_right"
	ChallengeHeadNod       ChallengeType = "head_nod"
	ChallengeRaiseEyebrows ChallengeType = "raise_eyebrows"
)

// DefaultLivenessScorerConfig returns sensible defaults
func DefaultLivenessScorerConfig() LivenessScorerConfig {
	return LivenessScorerConfig{
		SidecarAddress:   "localhost:50054",
		Timeout:          10 * time.Second,
		MinLivenessScore: 0.75,
		MinConfidence:    0.7,
		RequiredChallenges: []ChallengeType{
			ChallengeBlink,
		},
		OptionalChallenges: []ChallengeType{
			ChallengeSmile,
			ChallengeHeadTurnLeft,
		},
		UseFallbackOnError: true,
		StrictMode:         false,
		MaxRetries:         2,
		RetryDelay:         100 * time.Millisecond,
	}
}

// LivenessClient defines the interface for liveness detection backends
type LivenessClient interface {
	// CheckLiveness performs liveness detection on video frames
	CheckLiveness(ctx context.Context, videoData []byte, challenges []ChallengeType) (*LivenessResult, error)

	// CheckFrame performs quick liveness check on a single frame
	CheckFrame(ctx context.Context, imageData []byte) (*SingleFrameLivenessResult, error)

	// IsHealthy checks if the client is ready
	IsHealthy() bool

	// Close releases resources
	Close() error
}

// LivenessResult contains the complete result of liveness detection
type LivenessResult struct {
	// IsLive indicates if the subject passed liveness detection
	IsLive bool

	// Decision is the overall decision: "live", "spoof", or "uncertain"
	Decision string

	// LivenessScore is the overall liveness score (0.0-1.0)
	LivenessScore float32

	// Confidence in the liveness decision (0.0-1.0)
	Confidence float32

	// ActiveChallengeScore is the score from active challenge detection
	ActiveChallengeScore float32

	// PassiveAnalysisScore is the score from passive analysis
	PassiveAnalysisScore float32

	// SpoofDetectionScore is the inverse spoof score (1.0 = not spoof)
	SpoofDetectionScore float32

	// ChallengeResults contains per-challenge results
	ChallengeResults map[ChallengeType]*ChallengeResult

	// SpoofIndicators contains detected spoof indicators
	SpoofIndicators []SpoofIndicator

	// ModelVersion used for detection
	ModelVersion string

	// ModelHash for determinism verification
	ModelHash string

	// ResultHash for consensus verification
	ResultHash string

	// ProcessingTimeMs is the detection time in milliseconds
	ProcessingTimeMs int64

	// FrameCount is the number of frames analyzed
	FrameCount int

	// ValidFrameCount is the number of valid frames
	ValidFrameCount int

	// ReasonCodes provide explanations for the decision
	ReasonCodes []string
}

// ChallengeResult contains the result of a single challenge
type ChallengeResult struct {
	// ChallengeType is the type of challenge
	ChallengeType ChallengeType

	// Passed indicates if the challenge was passed
	Passed bool

	// Score is the challenge score (0.0-1.0)
	Score float32

	// Confidence in the challenge result (0.0-1.0)
	Confidence float32

	// DetectedAt is when the challenge was detected (frame index)
	DetectedAt int

	// Duration is how long the challenge was held (milliseconds)
	DurationMs int64
}

// SpoofIndicator represents a detected spoof attack indicator
type SpoofIndicator struct {
	// Type is the type of spoof attack detected
	Type SpoofType

	// Score is the detection score (0.0-1.0, higher = more likely spoof)
	Score float32

	// Confidence in the detection (0.0-1.0)
	Confidence float32

	// Description provides details about the indicator
	Description string
}

// SpoofType represents types of spoof attacks
type SpoofType string

const (
	SpoofTypePhotoPrint  SpoofType = "photo_print"
	SpoofTypePhotoScreen SpoofType = "photo_screen"
	SpoofTypeVideoReplay SpoofType = "video_replay"
	SpoofTypeMask2D      SpoofType = "mask_2d"
	SpoofTypeMask3D      SpoofType = "mask_3d"
	SpoofTypeDeepfake    SpoofType = "deepfake"
)

// SingleFrameLivenessResult contains quick liveness check from single frame
type SingleFrameLivenessResult struct {
	// LivenessScore from passive analysis
	LivenessScore float32

	// TextureScore from LBP texture analysis
	TextureScore float32

	// DepthScore from depth estimation
	DepthScore float32

	// ReflectionScore detecting reflections
	ReflectionScore float32

	// MoireScore detecting screen moire patterns
	MoireScore float32

	// ReasonCodes for the result
	ReasonCodes []string
}

// ============================================================================
// Liveness Reason Codes
// ============================================================================

const (
	LivenessReasonCodeSuccess             = "LIVENESS_CONFIRMED"
	LivenessReasonCodeHighConfidence      = "HIGH_CONFIDENCE_LIVE"
	LivenessReasonCodeSpoof               = "SPOOF_DETECTED"
	LivenessReasonCodeSpoofHighConf       = "SPOOF_HIGH_CONFIDENCE"
	LivenessReasonCodeUncertain           = "LIVENESS_UNCERTAIN"
	LivenessReasonCodeInsufficientFrames  = "INSUFFICIENT_FRAMES"
	LivenessReasonCodeChallengeFailed     = "CHALLENGE_FAILED"
	LivenessReasonCodeAllChallengesPassed = "ALL_CHALLENGES_PASSED"
	LivenessReasonCodePhotoSpoof          = "PHOTO_SPOOF_DETECTED"
	LivenessReasonCodeScreenSpoof         = "SCREEN_SPOOF_DETECTED"
	LivenessReasonCodeDeepfakeSpoof       = "DEEPFAKE_DETECTED"
	LivenessReasonCodeMaskSpoof           = "MASK_SPOOF_DETECTED"
	LivenessReasonCodeLowTextureVar       = "LOW_TEXTURE_VARIANCE"
	LivenessReasonCodeMoirePattern        = "MOIRE_PATTERN_DETECTED"
	LivenessReasonCodeExtractionError     = "LIVENESS_EXTRACTION_ERROR"
	LivenessReasonCodeSidecarUnavail      = "LIVENESS_SIDECAR_UNAVAILABLE"
	LivenessReasonCodeTimeout             = "LIVENESS_TIMEOUT"
)

// ============================================================================
// Constructor and Interface Implementation
// ============================================================================

// NewLivenessScorer creates a new liveness scorer
func NewLivenessScorer(config LivenessScorerConfig) *LivenessScorer {
	return &LivenessScorer{
		config:    config,
		isHealthy: false,
	}
}

// NewLivenessScorerWithClient creates a liveness scorer with a specific client
func NewLivenessScorerWithClient(config LivenessScorerConfig, client LivenessClient) *LivenessScorer {
	ls := NewLivenessScorer(config)
	ls.client = client
	ls.isHealthy = client != nil && client.IsHealthy()
	return ls
}

// CheckLiveness performs full liveness detection on video data
func (ls *LivenessScorer) CheckLiveness(ctx context.Context, videoData []byte) (*LivenessResult, error) {
	if len(videoData) == 0 {
		return ls.createFailureResult(LivenessReasonCodeInsufficientFrames, "empty video data"), nil
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, ls.config.Timeout)
	defer cancel()

	ls.mu.Lock()
	ls.checkCount++
	ls.mu.Unlock()

	// Combine required and optional challenges
	allChallenges := append(
		append([]ChallengeType{}, ls.config.RequiredChallenges...),
		ls.config.OptionalChallenges...,
	)

	// Try detection with retries
	var result *LivenessResult
	var lastErr error

	for attempt := 0; attempt <= ls.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ls.handleError(ctx.Err(), "timeout during retry")
			case <-time.After(ls.config.RetryDelay):
			}
		}

		result, lastErr = ls.doLivenessCheck(ctx, videoData, allChallenges)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return ls.handleError(lastErr, "liveness check failed after retries")
	}

	// Post-process and validate result
	ls.postProcessResult(result)

	return result, nil
}

// CheckSingleFrame performs quick liveness check on a single frame
func (ls *LivenessScorer) CheckSingleFrame(ctx context.Context, imageData []byte) (*SingleFrameLivenessResult, error) {
	if len(imageData) == 0 {
		return &SingleFrameLivenessResult{
			LivenessScore: 0.0,
			ReasonCodes:   []string{LivenessReasonCodeExtractionError},
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, ls.config.Timeout/2)
	defer cancel()

	if ls.client == nil || !ls.client.IsHealthy() {
		return ls.checkFrameFallback(imageData), nil
	}

	result, err := ls.client.CheckFrame(ctx, imageData)
	if err != nil {
		if ls.config.UseFallbackOnError {
			return ls.checkFrameFallback(imageData), nil
		}
		return nil, fmt.Errorf("single frame check failed: %w", err)
	}

	return result, nil
}

// doLivenessCheck performs the actual liveness detection
func (ls *LivenessScorer) doLivenessCheck(ctx context.Context, videoData []byte, challenges []ChallengeType) (*LivenessResult, error) {
	startTime := time.Now()

	if ls.client == nil || !ls.client.IsHealthy() {
		return ls.checkLivenessFallback(videoData, challenges, startTime), nil
	}

	result, err := ls.client.CheckLiveness(ctx, videoData, challenges)
	if err != nil {
		return nil, err
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// checkLivenessFallback generates deterministic stub liveness results
func (ls *LivenessScorer) checkLivenessFallback(videoData []byte, challenges []ChallengeType, startTime time.Time) *LivenessResult {
	// Generate deterministic scores from video hash
	hash := sha256.Sum256(videoData)

	// Generate challenge results
	challengeResults := make(map[ChallengeType]*ChallengeResult)
	allChallengesPassed := true

	for i, challenge := range challenges {
		byteIdx := i % len(hash)
		score := 0.7 + float32(hash[byteIdx]%30)/100.0
		passed := score >= 0.75

		if !passed && ls.isRequiredChallenge(challenge) {
			allChallengesPassed = false
		}

		challengeResults[challenge] = &ChallengeResult{
			ChallengeType: challenge,
			Passed:        passed,
			Score:         score,
			Confidence:    0.85,
			DetectedAt:    int(hash[byteIdx]) % 30,
			DurationMs:    200 + int64(hash[byteIdx])%300,
		}
	}

	// Compute component scores
	activeScore := ls.computeActiveScore(challengeResults)
	passiveScore := 0.75 + float32(hash[0]%25)/100.0
	spoofScore := 0.90 + float32(hash[1]%10)/100.0 // High = not spoof

	// Compute combined liveness score
	livenessScore := activeScore*0.4 + passiveScore*0.35 + spoofScore*0.25

	// Determine decision
	isLive := livenessScore >= ls.config.MinLivenessScore && allChallengesPassed
	decision := "live"
	if !isLive {
		if livenessScore < 0.5 {
			decision = "spoof"
		} else {
			decision = "uncertain"
		}
	}

	// Build reason codes
	reasonCodes := []string{LivenessReasonCodeSidecarUnavail}
	if isLive {
		reasonCodes = append(reasonCodes, LivenessReasonCodeSuccess)
		if livenessScore >= 0.9 {
			reasonCodes = append(reasonCodes, LivenessReasonCodeHighConfidence)
		}
	} else {
		reasonCodes = append(reasonCodes, LivenessReasonCodeUncertain)
	}
	if allChallengesPassed {
		reasonCodes = append(reasonCodes, LivenessReasonCodeAllChallengesPassed)
	}

	// Compute result hash
	resultHash := ls.computeResultHash(livenessScore, isLive, decision, reasonCodes)

	return &LivenessResult{
		IsLive:               isLive,
		Decision:             decision,
		LivenessScore:        livenessScore,
		Confidence:           0.85,
		ActiveChallengeScore: activeScore,
		PassiveAnalysisScore: passiveScore,
		SpoofDetectionScore:  spoofScore,
		ChallengeResults:     challengeResults,
		SpoofIndicators:      []SpoofIndicator{},
		ModelVersion:         "stub-v1.0.0",
		ModelHash:            hex.EncodeToString(hash[:16]),
		ResultHash:           resultHash,
		ProcessingTimeMs:     time.Since(startTime).Milliseconds(),
		FrameCount:           30,
		ValidFrameCount:      28,
		ReasonCodes:          reasonCodes,
	}
}

// checkFrameFallback generates deterministic stub single frame results
func (ls *LivenessScorer) checkFrameFallback(imageData []byte) *SingleFrameLivenessResult {
	hash := sha256.Sum256(imageData)

	return &SingleFrameLivenessResult{
		LivenessScore:   0.75 + float32(hash[0]%25)/100.0,
		TextureScore:    0.70 + float32(hash[1]%30)/100.0,
		DepthScore:      0.65 + float32(hash[2]%35)/100.0,
		ReflectionScore: 0.80 + float32(hash[3]%20)/100.0,
		MoireScore:      0.85 + float32(hash[4]%15)/100.0,
		ReasonCodes:     []string{LivenessReasonCodeSidecarUnavail},
	}
}

// computeActiveScore computes the overall active challenge score
func (ls *LivenessScorer) computeActiveScore(results map[ChallengeType]*ChallengeResult) float32 {
	if len(results) == 0 {
		return 0.5 // Neutral if no challenges
	}

	var totalScore float32
	var totalWeight float32

	for challenge, result := range results {
		weight := float32(1.0)
		if ls.isRequiredChallenge(challenge) {
			weight = 2.0 // Required challenges weighted higher
		}

		totalScore += result.Score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.5
	}

	return totalScore / totalWeight
}

// isRequiredChallenge checks if a challenge is required
func (ls *LivenessScorer) isRequiredChallenge(challenge ChallengeType) bool {
	for _, req := range ls.config.RequiredChallenges {
		if req == challenge {
			return true
		}
	}
	return false
}

// postProcessResult validates and enriches the liveness result
func (ls *LivenessScorer) postProcessResult(result *LivenessResult) {
	if result == nil {
		return
	}

	// Clamp score values
	result.LivenessScore = clampFloat32(result.LivenessScore, 0.0, 1.0)
	result.Confidence = clampFloat32(result.Confidence, 0.0, 1.0)
	result.ActiveChallengeScore = clampFloat32(result.ActiveChallengeScore, 0.0, 1.0)
	result.PassiveAnalysisScore = clampFloat32(result.PassiveAnalysisScore, 0.0, 1.0)
	result.SpoofDetectionScore = clampFloat32(result.SpoofDetectionScore, 0.0, 1.0)

	// Clamp challenge result scores
	for _, cr := range result.ChallengeResults {
		if cr != nil {
			cr.Score = clampFloat32(cr.Score, 0.0, 1.0)
			cr.Confidence = clampFloat32(cr.Confidence, 0.0, 1.0)
		}
	}

	// Add reason codes for spoof indicators
	for _, indicator := range result.SpoofIndicators {
		if indicator.Score > 0.5 {
			switch indicator.Type {
			case SpoofTypePhotoPrint, SpoofTypePhotoScreen:
				result.ReasonCodes = appendIfNotExists(result.ReasonCodes, LivenessReasonCodePhotoSpoof)
			case SpoofTypeVideoReplay:
				result.ReasonCodes = appendIfNotExists(result.ReasonCodes, LivenessReasonCodeScreenSpoof)
			case SpoofTypeDeepfake:
				result.ReasonCodes = appendIfNotExists(result.ReasonCodes, LivenessReasonCodeDeepfakeSpoof)
			case SpoofTypeMask2D, SpoofTypeMask3D:
				result.ReasonCodes = appendIfNotExists(result.ReasonCodes, LivenessReasonCodeMaskSpoof)
			}
		}
	}

	// Check for failed required challenges
	if ls.config.StrictMode {
		for _, reqChallenge := range ls.config.RequiredChallenges {
			if cr, exists := result.ChallengeResults[reqChallenge]; exists && !cr.Passed {
				result.ReasonCodes = appendIfNotExists(result.ReasonCodes, LivenessReasonCodeChallengeFailed)
				result.IsLive = false
				result.Decision = "uncertain"
			}
		}
	}

	// Compute result hash if not present
	if result.ResultHash == "" {
		result.ResultHash = ls.computeResultHash(
			result.LivenessScore,
			result.IsLive,
			result.Decision,
			result.ReasonCodes,
		)
	}
}

// computeResultHash computes a deterministic hash of the result
func (ls *LivenessScorer) computeResultHash(score float32, isLive bool, decision string, reasonCodes []string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%.4f|%v|%s|", score, isLive, decision)))
	for _, code := range reasonCodes {
		h.Write([]byte(code))
		h.Write([]byte(","))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// handleError handles liveness errors with optional fallback
func (ls *LivenessScorer) handleError(err error, msg string) (*LivenessResult, error) {
	ls.mu.Lock()
	ls.errorCount++
	ls.mu.Unlock()

	if ls.config.UseFallbackOnError {
		return ls.createFailureResult(LivenessReasonCodeExtractionError, msg), nil
	}
	return nil, fmt.Errorf("%s: %w", msg, err)
}

// createFailureResult creates a result indicating liveness check failure
//
//nolint:unparam // message kept for future logging or error details in result
func (ls *LivenessScorer) createFailureResult(reasonCode, _ string) *LivenessResult {
	return &LivenessResult{
		IsLive:           false,
		Decision:         "uncertain",
		LivenessScore:    0.0,
		Confidence:       0.0,
		ChallengeResults: make(map[ChallengeType]*ChallengeResult),
		SpoofIndicators:  []SpoofIndicator{},
		ModelVersion:     "unknown",
		ReasonCodes:      []string{reasonCode},
	}
}

// ============================================================================
// Feature Vector Conversion
// ============================================================================

// ToScoreInputContribution returns values that contribute to ScoreInputs
func (ls *LivenessScorer) ToScoreInputContribution(result *LivenessResult) (float32, []string) {
	if result == nil {
		return 0.0, []string{}
	}

	// Return liveness score and reason codes
	return result.LivenessScore, result.ReasonCodes
}

// GetChallengeScoreVector returns challenge scores as a feature vector
func (ls *LivenessScorer) GetChallengeScoreVector(result *LivenessResult) []float32 {
	// Fixed order for deterministic feature vector
	challengeOrder := []ChallengeType{
		ChallengeBlink,
		ChallengeSmile,
		ChallengeHeadTurnLeft,
		ChallengeHeadTurnRight,
		ChallengeHeadNod,
		ChallengeRaiseEyebrows,
	}

	scores := make([]float32, len(challengeOrder))

	if result == nil {
		return scores
	}

	for i, challenge := range challengeOrder {
		if cr, exists := result.ChallengeResults[challenge]; exists && cr != nil {
			scores[i] = cr.Score
		}
	}

	return scores
}

// GetComponentScores returns the three main component scores
func (ls *LivenessScorer) GetComponentScores(result *LivenessResult) (active, passive, spoof float32) {
	if result == nil {
		return 0.0, 0.0, 0.0
	}
	return result.ActiveChallengeScore, result.PassiveAnalysisScore, result.SpoofDetectionScore
}

// ============================================================================
// Validation Functions
// ============================================================================

// ValidateResult validates that a liveness result meets quality requirements
func (ls *LivenessScorer) ValidateResult(result *LivenessResult) []string {
	var issues []string

	if result == nil {
		issues = append(issues, "nil liveness result")
		return issues
	}

	// Check score ranges
	if result.LivenessScore < 0 || result.LivenessScore > 1 {
		issues = append(issues, fmt.Sprintf(
			"liveness score out of range [0,1]: %.4f",
			result.LivenessScore,
		))
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		issues = append(issues, fmt.Sprintf(
			"confidence out of range [0,1]: %.4f",
			result.Confidence,
		))
	}

	// Check for NaN values
	if math.IsNaN(float64(result.LivenessScore)) {
		issues = append(issues, "liveness score is NaN")
	}
	if math.IsNaN(float64(result.Confidence)) {
		issues = append(issues, "confidence is NaN")
	}

	// Check decision validity
	validDecisions := map[string]bool{"live": true, "spoof": true, "uncertain": true}
	if !validDecisions[result.Decision] {
		issues = append(issues, fmt.Sprintf(
			"invalid decision: %s",
			result.Decision,
		))
	}

	// Check challenge results
	for challenge, cr := range result.ChallengeResults {
		if cr == nil {
			continue
		}
		if cr.Score < 0 || cr.Score > 1 {
			issues = append(issues, fmt.Sprintf(
				"challenge '%s' score out of range [0,1]: %.4f",
				challenge, cr.Score,
			))
		}
	}

	return issues
}

// SanitizeResult ensures all values in a liveness result are valid
func (ls *LivenessScorer) SanitizeResult(result *LivenessResult) {
	if result == nil {
		return
	}

	// Sanitize main scores
	result.LivenessScore = sanitizeFloat32(result.LivenessScore, 0.0, 1.0)
	result.Confidence = sanitizeFloat32(result.Confidence, 0.0, 1.0)
	result.ActiveChallengeScore = sanitizeFloat32(result.ActiveChallengeScore, 0.0, 1.0)
	result.PassiveAnalysisScore = sanitizeFloat32(result.PassiveAnalysisScore, 0.0, 1.0)
	result.SpoofDetectionScore = sanitizeFloat32(result.SpoofDetectionScore, 0.0, 1.0)

	// Sanitize challenge results
	for _, cr := range result.ChallengeResults {
		if cr == nil {
			continue
		}
		cr.Score = sanitizeFloat32(cr.Score, 0.0, 1.0)
		cr.Confidence = sanitizeFloat32(cr.Confidence, 0.0, 1.0)
	}

	// Ensure valid decision
	validDecisions := map[string]bool{"live": true, "spoof": true, "uncertain": true}
	if !validDecisions[result.Decision] {
		result.Decision = "uncertain"
	}
}

// sanitizeFloat32 ensures a float32 is valid and in range
func sanitizeFloat32(v, min, max float32) float32 {
	if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
		return min
	}
	return clampFloat32(v, min, max)
}

// appendIfNotExists appends a value to a slice only if it doesn't exist
func appendIfNotExists(slice []string, value string) []string {
	for _, v := range slice {
		if v == value {
			return slice
		}
	}
	return append(slice, value)
}

// ============================================================================
// Statistics and Health
// ============================================================================

// IsHealthy returns whether the scorer is ready
func (ls *LivenessScorer) IsHealthy() bool {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	if ls.client != nil {
		return ls.client.IsHealthy()
	}
	// Stub mode is always healthy
	return true
}

// GetStats returns scorer statistics
func (ls *LivenessScorer) GetStats() LivenessScorerStats {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	return LivenessScorerStats{
		CheckCount: ls.checkCount,
		ErrorCount: ls.errorCount,
		IsHealthy:  ls.IsHealthy(),
		UsingStub:  ls.client == nil,
	}
}

// LivenessScorerStats contains scorer statistics
type LivenessScorerStats struct {
	CheckCount uint64
	ErrorCount uint64
	IsHealthy  bool
	UsingStub  bool
}

// Close releases resources
func (ls *LivenessScorer) Close() error {
	if ls.client != nil {
		return ls.client.Close()
	}
	return nil
}
