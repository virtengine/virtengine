package mobile

import (
	"time"
)

// ============================================================================
// Quality Feedback Types (Real-time Quality Guidance)
// ============================================================================

// QualityConfiguration defines quality validation thresholds
type QualityConfiguration struct {
	// MinResolution is the minimum acceptable resolution
	MinResolution Resolution `json:"min_resolution"`

	// MinBrightness is minimum brightness (0-255)
	MinBrightness int `json:"min_brightness"`

	// MaxBrightness is maximum brightness (0-255)
	MaxBrightness int `json:"max_brightness"`

	// MaxBlurScore is maximum blur threshold (Laplacian variance)
	MaxBlurScore float64 `json:"max_blur_score"`

	// MaxSkewAngle is maximum skew angle in degrees
	MaxSkewAngle float64 `json:"max_skew_angle"`

	// MaxGlarePercentage is maximum glare area percentage (0-1)
	MaxGlarePercentage float64 `json:"max_glare_percentage"`

	// MinQualityScore is minimum overall quality score (0-100)
	MinQualityScore int `json:"min_quality_score"`

	// RequireDocumentDetection requires document in frame for documents
	RequireDocumentDetection bool `json:"require_document_detection"`

	// RequireFaceDetection requires face in frame for selfies
	RequireFaceDetection bool `json:"require_face_detection"`
}

// DefaultDocumentQualityConfig returns default quality config for documents
func DefaultDocumentQualityConfig() QualityConfiguration {
	return QualityConfiguration{
		MinResolution:            Resolution{Width: 1024, Height: 768},
		MinBrightness:            40,
		MaxBrightness:            220,
		MaxBlurScore:             100,
		MaxSkewAngle:             10,
		MaxGlarePercentage:       0.15,
		MinQualityScore:          70,
		RequireDocumentDetection: true,
		RequireFaceDetection:     false,
	}
}

// DefaultSelfieQualityConfig returns default quality config for selfies
func DefaultSelfieQualityConfig() QualityConfiguration {
	return QualityConfiguration{
		MinResolution:            Resolution{Width: 640, Height: 480},
		MinBrightness:            50,
		MaxBrightness:            200,
		MaxBlurScore:             80,
		MaxSkewAngle:             15,
		MaxGlarePercentage:       0.1,
		MinQualityScore:          75,
		RequireDocumentDetection: false,
		RequireFaceDetection:     true,
	}
}

// QualityResult represents complete quality validation results
type QualityResult struct {
	// Passed indicates if all quality checks passed
	Passed bool `json:"passed"`

	// OverallScore is the aggregate quality score (0-100)
	OverallScore int `json:"overall_score"`

	// Issues contains detected quality issues
	Issues []QualityIssue `json:"issues,omitempty"`

	// Checks contains individual check results
	Checks QualityChecks `json:"checks"`

	// AnalysisTimeMs is the time taken for analysis
	AnalysisTimeMs int64 `json:"analysis_time_ms"`

	// Timestamp is when analysis was performed
	Timestamp time.Time `json:"timestamp"`
}

// QualityChecks contains individual quality check results
type QualityChecks struct {
	Resolution ResolutionCheck `json:"resolution"`
	Brightness BrightnessCheck `json:"brightness"`
	Blur       BlurCheck       `json:"blur"`
	Skew       SkewCheck       `json:"skew"`
	Glare      GlareCheck      `json:"glare"`
	Noise      NoiseCheck      `json:"noise"`
	Focus      FocusCheck      `json:"focus"`
	Document   *DocumentCheck  `json:"document,omitempty"`
	Face       *FaceCheck      `json:"face,omitempty"`
}

// ResolutionCheck represents resolution validation
type ResolutionCheck struct {
	Passed     bool       `json:"passed"`
	Actual     Resolution `json:"actual"`
	Required   Resolution `json:"required"`
	Megapixels float64    `json:"megapixels"`
}

// BrightnessCheck represents brightness validation
type BrightnessCheck struct {
	Passed    bool    `json:"passed"`
	Value     float64 `json:"value"`      // Average brightness (0-255)
	MinLimit  int     `json:"min_limit"`  // Minimum threshold
	MaxLimit  int     `json:"max_limit"`  // Maximum threshold
	IsDark    bool    `json:"is_dark"`    // Too dark
	IsBright  bool    `json:"is_bright"`  // Too bright
	Histogram []int   `json:"histogram"`  // 8-bin brightness histogram
}

// BlurCheck represents blur/sharpness validation
type BlurCheck struct {
	Passed           bool    `json:"passed"`
	LaplacianVariance float64 `json:"laplacian_variance"` // Higher = sharper
	Threshold        float64 `json:"threshold"`
	IsBlurry         bool    `json:"is_blurry"`
}

// SkewCheck represents document skew validation
type SkewCheck struct {
	Passed        bool    `json:"passed"`
	AngleDegrees  float64 `json:"angle_degrees"`   // Detected skew angle
	MaxAllowed    float64 `json:"max_allowed"`     // Maximum allowed angle
	IsSkewed      bool    `json:"is_skewed"`
	RotationNeeded float64 `json:"rotation_needed"` // Suggested rotation
}

// GlareCheck represents glare/reflection detection
type GlareCheck struct {
	Passed          bool      `json:"passed"`
	GlarePercentage float64   `json:"glare_percentage"` // % of image with glare
	MaxAllowed      float64   `json:"max_allowed"`
	GlareRegions    []Region  `json:"glare_regions,omitempty"` // Detected glare areas
}

// Region represents a rectangular region in an image
type Region struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// NoiseCheck represents image noise validation
type NoiseCheck struct {
	Passed     bool    `json:"passed"`
	NoiseLevel float64 `json:"noise_level"` // Noise estimate (0-1)
	MaxAllowed float64 `json:"max_allowed"`
	IsNoisy    bool    `json:"is_noisy"`
}

// FocusCheck represents focus/sharpness validation
type FocusCheck struct {
	Passed      bool    `json:"passed"`
	FocusScore  float64 `json:"focus_score"`   // Focus quality (0-100)
	MinRequired float64 `json:"min_required"`
	IsFocused   bool    `json:"is_focused"`
}

// DocumentCheck represents document detection validation
type DocumentCheck struct {
	Passed          bool      `json:"passed"`
	Detected        bool      `json:"detected"`
	Confidence      float64   `json:"confidence"`      // Detection confidence (0-1)
	BoundingBox     *Region   `json:"bounding_box,omitempty"`
	Corners         []Point   `json:"corners,omitempty"` // Four corners if detected
	CoveragePercent float64   `json:"coverage_percent"` // % of frame covered
	MinCoverage     float64   `json:"min_coverage"`     // Required coverage
}

// Point represents a 2D point
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// FaceCheck represents face detection validation
type FaceCheck struct {
	Passed           bool    `json:"passed"`
	Detected         bool    `json:"detected"`
	Confidence       float64 `json:"confidence"`
	BoundingBox      *Region `json:"bounding_box,omitempty"`
	FaceCount        int     `json:"face_count"`        // Number of faces detected
	IsCentered       bool    `json:"is_centered"`
	IsCorrectSize    bool    `json:"is_correct_size"`
	FaceSizePercent  float64 `json:"face_size_percent"` // % of frame
	MinFaceSizePercent float64 `json:"min_face_size_percent"`
	MaxFaceSizePercent float64 `json:"max_face_size_percent"`
}

// QualityIssue represents a detected quality problem
type QualityIssue struct {
	// Type is the issue type
	Type QualityIssueType `json:"type"`

	// Severity indicates issue severity
	Severity QualityIssueSeverity `json:"severity"`

	// Message is a user-friendly message
	Message string `json:"message"`

	// Suggestion is actionable guidance
	Suggestion string `json:"suggestion"`

	// Confidence is the detection confidence (0-1)
	Confidence float64 `json:"confidence"`

	// Region is the affected area (optional)
	Region *Region `json:"region,omitempty"`
}

// QualityIssueType represents types of quality issues
type QualityIssueType string

const (
	QualityIssueBlur          QualityIssueType = "blur"
	QualityIssueDark          QualityIssueType = "dark"
	QualityIssueBright        QualityIssueType = "bright"
	QualityIssueSkew          QualityIssueType = "skew"
	QualityIssueResolution    QualityIssueType = "resolution"
	QualityIssueGlare         QualityIssueType = "glare"
	QualityIssueNoise         QualityIssueType = "noise"
	QualityIssuePartial       QualityIssueType = "partial"
	QualityIssueReflection    QualityIssueType = "reflection"
	QualityIssueNoDocument    QualityIssueType = "no_document"
	QualityIssueNoFace        QualityIssueType = "no_face"
	QualityIssueFaceNotCentered QualityIssueType = "face_not_centered"
	QualityIssueFaceTooSmall  QualityIssueType = "face_too_small"
	QualityIssueFaceTooLarge  QualityIssueType = "face_too_large"
	QualityIssueMultipleFaces QualityIssueType = "multiple_faces"
)

// QualityIssueSeverity represents issue severity levels
type QualityIssueSeverity string

const (
	// SeverityWarning is a non-blocking issue
	SeverityWarning QualityIssueSeverity = "warning"

	// SeverityError is a blocking issue
	SeverityError QualityIssueSeverity = "error"
)

// ============================================================================
// Real-time Quality Feedback Interface
// ============================================================================

// QualityFeedback represents real-time quality feedback for UI
type QualityFeedback struct {
	// ReadyToCapture indicates image is ready for capture
	ReadyToCapture bool `json:"ready_to_capture"`

	// OverallScore is current quality score (0-100)
	OverallScore int `json:"overall_score"`

	// ActiveIssues are currently detected issues
	ActiveIssues []QualityIssue `json:"active_issues"`

	// Instruction is the current user guidance
	Instruction string `json:"instruction"`

	// InstructionType categorizes the instruction
	InstructionType InstructionType `json:"instruction_type"`

	// Progress indicates capture progress (0-100)
	Progress int `json:"progress"`

	// FrameAnalysisTimeMs is time taken for frame analysis
	FrameAnalysisTimeMs int64 `json:"frame_analysis_time_ms"`
}

// InstructionType categorizes guidance instructions
type InstructionType string

const (
	InstructionTypePositioning InstructionType = "positioning"
	InstructionTypeLighting    InstructionType = "lighting"
	InstructionTypeFocus       InstructionType = "focus"
	InstructionTypeStability   InstructionType = "stability"
	InstructionTypeReady       InstructionType = "ready"
	InstructionTypeCapturing   InstructionType = "capturing"
)

// QualityFeedbackCallback is the callback interface for quality updates
type QualityFeedbackCallback interface {
	// OnQualityUpdate is called when quality feedback changes
	OnQualityUpdate(feedback QualityFeedback)

	// OnQualityCheckComplete is called when final quality check completes
	OnQualityCheckComplete(result QualityResult)
}

// ============================================================================
// Localized Guidance Messages
// ============================================================================

// GuidanceMessages contains localized guidance messages
type GuidanceMessages struct {
	// PositionDocument - move document into frame
	PositionDocument string `json:"position_document"`

	// HoldSteady - keep device steady
	HoldSteady string `json:"hold_steady"`

	// MoveFurther - move device further away
	MoveFurther string `json:"move_further"`

	// MoveCloser - move device closer
	MoveCloser string `json:"move_closer"`

	// MoreLight - need more light
	MoreLight string `json:"more_light"`

	// LessLight - too much light/glare
	LessLight string `json:"less_light"`

	// ReduceGlare - adjust to reduce reflection
	ReduceGlare string `json:"reduce_glare"`

	// AlignDocument - straighten the document
	AlignDocument string `json:"align_document"`

	// PositionFace - position face in frame
	PositionFace string `json:"position_face"`

	// CenterFace - center face in frame
	CenterFace string `json:"center_face"`

	// Ready - ready to capture
	Ready string `json:"ready"`

	// Capturing - capture in progress
	Capturing string `json:"capturing"`

	// Processing - processing capture
	Processing string `json:"processing"`
}

// DefaultEnglishGuidance returns default English guidance messages
func DefaultEnglishGuidance() GuidanceMessages {
	return GuidanceMessages{
		PositionDocument: "Position the document within the frame",
		HoldSteady:       "Hold steady",
		MoveFurther:      "Move further away",
		MoveCloser:       "Move closer",
		MoreLight:        "Find better lighting",
		LessLight:        "Reduce light or glare",
		ReduceGlare:      "Tilt to reduce reflection",
		AlignDocument:    "Straighten the document",
		PositionFace:     "Position your face in the frame",
		CenterFace:       "Center your face",
		Ready:            "Perfect! Hold still...",
		Capturing:        "Capturing...",
		Processing:       "Processing...",
	}
}
