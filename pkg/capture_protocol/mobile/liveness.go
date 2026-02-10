package mobile

import (
	"time"
)

// ============================================================================
// Liveness Detection Types
// VE-900: Liveness detection interface during facial capture
// ============================================================================

// LivenessConfiguration defines liveness detection settings
type LivenessConfiguration struct {
	// Enabled indicates if liveness detection is enabled
	Enabled bool `json:"enabled"`

	// Mode is the liveness detection mode
	Mode LivenessMode `json:"mode"`

	// ChallengeTypes specifies which challenges to use (for active mode)
	ChallengeTypes []LivenessChallengeType `json:"challenge_types,omitempty"`

	// MinChallenges is minimum number of challenges to complete (active mode)
	MinChallenges int `json:"min_challenges"`

	// ChallengeTimeoutSeconds is timeout per challenge
	ChallengeTimeoutSeconds int `json:"challenge_timeout_seconds"`

	// TotalTimeoutSeconds is total liveness check timeout
	TotalTimeoutSeconds int `json:"total_timeout_seconds"`

	// MinConfidence is minimum required confidence (0-1)
	MinConfidence float64 `json:"min_confidence"`

	// UseDepthSensor enables depth-based liveness (if available)
	UseDepthSensor bool `json:"use_depth_sensor"`

	// RequireTexture enables texture analysis for anti-spoofing
	RequireTexture bool `json:"require_texture"`
}

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

// LivenessChallengeType represents types of active challenges
type LivenessChallengeType string

const (
	// ChallengeBlink asks user to blink
	ChallengeBlink LivenessChallengeType = "blink"

	// ChallengeSmile asks user to smile
	ChallengeSmile LivenessChallengeType = "smile"

	// ChallengeTurnLeft asks user to turn head left
	ChallengeTurnLeft LivenessChallengeType = "turn_left"

	// ChallengeTurnRight asks user to turn head right
	ChallengeTurnRight LivenessChallengeType = "turn_right"

	// ChallengeNod asks user to nod
	ChallengeNod LivenessChallengeType = "nod"

	// ChallengeOpenMouth asks user to open mouth
	ChallengeOpenMouth LivenessChallengeType = "open_mouth"

	// ChallengeRaiseBrows asks user to raise eyebrows
	ChallengeRaiseBrows LivenessChallengeType = "raise_brows"
)

// DefaultLivenessConfig returns default liveness configuration
func DefaultLivenessConfig() LivenessConfiguration {
	return LivenessConfiguration{
		Enabled:                 true,
		Mode:                    LivenessModeHybrid,
		ChallengeTypes:          []LivenessChallengeType{ChallengeBlink, ChallengeSmile},
		MinChallenges:           1,
		ChallengeTimeoutSeconds: 10,
		TotalTimeoutSeconds:     30,
		MinConfidence:           0.85,
		UseDepthSensor:          true,
		RequireTexture:          true,
	}
}

// ============================================================================
// Liveness Detection Results
// ============================================================================

// LivenessResult represents complete liveness check results
type LivenessResult struct {
	// Passed indicates if liveness was verified
	Passed bool `json:"passed"`

	// Confidence is the overall liveness confidence (0-1)
	Confidence float64 `json:"confidence"`

	// Mode used for detection
	Mode LivenessMode `json:"mode"`

	// PassiveResult contains passive detection results
	PassiveResult *PassiveLivenessResult `json:"passive_result,omitempty"`

	// ActiveResult contains active challenge results
	ActiveResult *ActiveLivenessResult `json:"active_result,omitempty"`

	// AntiSpoofingResult contains anti-spoofing analysis
	AntiSpoofingResult AntiSpoofingResult `json:"anti_spoofing_result"`

	// DurationMs is total liveness check duration
	DurationMs int64 `json:"duration_ms"`

	// Timestamp is when liveness was verified
	Timestamp time.Time `json:"timestamp"`
}

// PassiveLivenessResult contains passive detection details
type PassiveLivenessResult struct {
	// Passed indicates passive checks passed
	Passed bool `json:"passed"`

	// Confidence is passive detection confidence
	Confidence float64 `json:"confidence"`

	// DepthAnalysis contains depth-based results (if available)
	DepthAnalysis *DepthAnalysisResult `json:"depth_analysis,omitempty"`

	// TextureAnalysis contains texture-based results
	TextureAnalysis *TextureAnalysisResult `json:"texture_analysis,omitempty"`

	// MotionAnalysis contains motion-based results
	MotionAnalysis *MotionAnalysisResult `json:"motion_analysis,omitempty"`
}

// DepthAnalysisResult contains depth sensor analysis
type DepthAnalysisResult struct {
	// Available indicates depth sensor was available
	Available bool `json:"available"`

	// Passed indicates depth check passed
	Passed bool `json:"passed"`

	// Is3DFace indicates a 3D face was detected (not flat)
	Is3DFace bool `json:"is_3d_face"`

	// DepthVariance is the depth variation in face region
	DepthVariance float64 `json:"depth_variance"`

	// MinVarianceRequired is the minimum variance for 3D
	MinVarianceRequired float64 `json:"min_variance_required"`

	// Confidence is depth analysis confidence
	Confidence float64 `json:"confidence"`
}

// TextureAnalysisResult contains texture anti-spoofing analysis
type TextureAnalysisResult struct {
	// Passed indicates texture check passed
	Passed bool `json:"passed"`

	// IsNaturalSkin indicates natural skin texture detected
	IsNaturalSkin bool `json:"is_natural_skin"`

	// MoireDetected indicates screen moire pattern detected
	MoireDetected bool `json:"moire_detected"`

	// PrintPatternDetected indicates printed image pattern
	PrintPatternDetected bool `json:"print_pattern_detected"`

	// ReflectionDetected indicates abnormal reflections (screen/photo)
	ReflectionDetected bool `json:"reflection_detected"`

	// Confidence is texture analysis confidence
	Confidence float64 `json:"confidence"`
}

// MotionAnalysisResult contains motion-based liveness analysis
type MotionAnalysisResult struct {
	// Passed indicates motion analysis passed
	Passed bool `json:"passed"`

	// NaturalMovement indicates natural micro-movements detected
	NaturalMovement bool `json:"natural_movement"`

	// BlinkDetected indicates natural blink detected
	BlinkDetected bool `json:"blink_detected"`

	// MicroExpressions indicates micro-expressions detected
	MicroExpressions bool `json:"micro_expressions"`

	// Confidence is motion analysis confidence
	Confidence float64 `json:"confidence"`
}

// ActiveLivenessResult contains active challenge results
type ActiveLivenessResult struct {
	// Passed indicates active challenges passed
	Passed bool `json:"passed"`

	// Confidence is aggregate confidence
	Confidence float64 `json:"confidence"`

	// ChallengesCompleted is number of challenges completed
	ChallengesCompleted int `json:"challenges_completed"`

	// ChallengesRequired is number of challenges required
	ChallengesRequired int `json:"challenges_required"`

	// Challenges contains individual challenge results
	Challenges []ChallengeResult `json:"challenges"`

	// TotalDurationMs is total time for all challenges
	TotalDurationMs int64 `json:"total_duration_ms"`
}

// ChallengeResult represents a single challenge result
type ChallengeResult struct {
	// Type is the challenge type
	Type LivenessChallengeType `json:"type"`

	// Passed indicates challenge was completed
	Passed bool `json:"passed"`

	// Confidence is detection confidence
	Confidence float64 `json:"confidence"`

	// ResponseTimeMs is user response time
	ResponseTimeMs int64 `json:"response_time_ms"`

	// Attempts is number of attempts
	Attempts int `json:"attempts"`

	// TimedOut indicates challenge timed out
	TimedOut bool `json:"timed_out"`
}

// AntiSpoofingResult contains anti-spoofing analysis
type AntiSpoofingResult struct {
	// Passed indicates anti-spoofing checks passed
	Passed bool `json:"passed"`

	// IsRealFace indicates face appears genuine
	IsRealFace bool `json:"is_real_face"`

	// SpoofType is detected spoof type (if any)
	SpoofType SpoofType `json:"spoof_type,omitempty"`

	// Confidence is anti-spoofing confidence
	Confidence float64 `json:"confidence"`

	// Checks contains individual anti-spoofing checks
	Checks AntiSpoofingChecks `json:"checks"`
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

// AntiSpoofingChecks contains individual anti-spoofing check results
type AntiSpoofingChecks struct {
	// PhotoAttack is photo attack detection
	PhotoAttack SpoofCheck `json:"photo_attack"`

	// ScreenAttack is screen/video attack detection
	ScreenAttack SpoofCheck `json:"screen_attack"`

	// MaskAttack is mask attack detection
	MaskAttack SpoofCheck `json:"mask_attack"`

	// DeepfakeAttack is deepfake detection
	DeepfakeAttack SpoofCheck `json:"deepfake_attack"`
}

// SpoofCheck represents a single spoof detection check
type SpoofCheck struct {
	// Detected indicates if spoof was detected
	Detected bool `json:"detected"`

	// Confidence is detection confidence
	Confidence float64 `json:"confidence"`

	// Indicators are specific detection indicators
	Indicators []string `json:"indicators,omitempty"`
}

// ============================================================================
// Liveness Detection Interface
// ============================================================================

// LivenessCallback is the callback interface for liveness detection
type LivenessCallback interface {
	// OnChallengeStart is called when a challenge begins
	OnChallengeStart(challenge LivenessChallengeType)

	// OnChallengeProgress is called with challenge progress
	OnChallengeProgress(challenge LivenessChallengeType, progress float64)

	// OnChallengeComplete is called when challenge completes
	OnChallengeComplete(result ChallengeResult)

	// OnLivenessComplete is called when liveness check completes
	OnLivenessComplete(result LivenessResult)

	// OnLivenessError is called on error
	OnLivenessError(err error)
}

// LivenessDetector defines the liveness detection interface
type LivenessDetector interface {
	// Configure configures the liveness detector
	Configure(config LivenessConfiguration) error

	// Start starts liveness detection
	Start(callback LivenessCallback) error

	// ProcessFrame processes a video frame for liveness
	ProcessFrame(frame []byte, timestamp time.Time) error

	// GetCurrentState returns current detection state
	GetCurrentState() LivenessState

	// Cancel cancels the liveness check
	Cancel()

	// GetResult returns the final result (after completion)
	GetResult() (*LivenessResult, error)
}

// LivenessState represents current liveness detection state
type LivenessState struct {
	// Phase is the current detection phase
	Phase LivenessPhase `json:"phase"`

	// CurrentChallenge is the active challenge (if any)
	CurrentChallenge *LivenessChallengeType `json:"current_challenge,omitempty"`

	// ChallengesCompleted is number completed
	ChallengesCompleted int `json:"challenges_completed"`

	// ChallengesRemaining is number remaining
	ChallengesRemaining int `json:"challenges_remaining"`

	// Progress is overall progress (0-100)
	Progress int `json:"progress"`

	// Instruction is current user instruction
	Instruction string `json:"instruction"`

	// TimeRemainingMs is time remaining for current phase
	TimeRemainingMs int64 `json:"time_remaining_ms"`
}

// LivenessPhase represents detection phases
type LivenessPhase string

const (
	LivenessPhaseInitializing    LivenessPhase = "initializing"
	LivenessPhasePassive         LivenessPhase = "passive"
	LivenessPhaseActiveChallenge LivenessPhase = "active_challenge"
	LivenessPhaseProcessing      LivenessPhase = "processing"
	LivenessPhaseComplete        LivenessPhase = "complete"
	LivenessPhaseFailed          LivenessPhase = "failed"
)
