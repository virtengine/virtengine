// Package govdata provides government data source integration for identity verification.
//
// SECURITY-004: Fraud detection hooks for government document verification
package govdata

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ============================================================================
// Fraud Detection Errors
// ============================================================================

var (
	// ErrFraudDetected is returned when fraud is detected
	ErrFraudDetected = errors.New("fraud detected during verification")

	// ErrFraudCheckFailed is returned when fraud check fails
	ErrFraudCheckFailed = errors.New("fraud check failed")

	// ErrFraudServiceUnavailable is returned when fraud service is unavailable
	ErrFraudServiceUnavailable = errors.New("fraud detection service unavailable")
)

// ============================================================================
// Fraud Detection Types
// ============================================================================

// FraudSignalType represents types of fraud signals
type FraudSignalType string

const (
	// FraudSignalDuplicateDocument indicates duplicate document detected
	FraudSignalDuplicateDocument FraudSignalType = "duplicate_document"

	// FraudSignalStolenIdentity indicates potential identity theft
	FraudSignalStolenIdentity FraudSignalType = "stolen_identity"

	// FraudSignalFakeDocument indicates fake/forged document detected
	FraudSignalFakeDocument FraudSignalType = "fake_document"

	// FraudSignalExpiredDocument indicates expired document being used
	FraudSignalExpiredDocument FraudSignalType = "expired_document"

	// FraudSignalRevokedDocument indicates revoked document being used
	FraudSignalRevokedDocument FraudSignalType = "revoked_document"

	// FraudSignalMismatchedData indicates data mismatch across sources
	FraudSignalMismatchedData FraudSignalType = "mismatched_data"

	// FraudSignalSuspiciousPattern indicates suspicious verification pattern
	FraudSignalSuspiciousPattern FraudSignalType = "suspicious_pattern"

	// FraudSignalSpoofingAttempt indicates spoofing/liveness failure
	FraudSignalSpoofingAttempt FraudSignalType = "spoofing_attempt"

	// FraudSignalVelocityAnomaly indicates abnormal verification velocity
	FraudSignalVelocityAnomaly FraudSignalType = "velocity_anomaly"

	// FraudSignalBlacklistedDocument indicates document on blocklist
	FraudSignalBlacklistedDocument FraudSignalType = "blacklisted_document"

	// FraudSignalSybilRisk indicates potential sybil attack
	FraudSignalSybilRisk FraudSignalType = "sybil_risk"
)

// FraudSeverity represents the severity of a fraud signal
type FraudSeverity string

const (
	// FraudSeverityLow indicates low-severity fraud signal
	FraudSeverityLow FraudSeverity = "low"

	// FraudSeverityMedium indicates medium-severity fraud signal
	FraudSeverityMedium FraudSeverity = "medium"

	// FraudSeverityHigh indicates high-severity fraud signal
	FraudSeverityHigh FraudSeverity = "high"

	// FraudSeverityCritical indicates critical fraud signal
	FraudSeverityCritical FraudSeverity = "critical"
)

// FraudAction represents actions taken on fraud detection
type FraudAction string

const (
	// FraudActionNone indicates no action (logging only)
	FraudActionNone FraudAction = "none"

	// FraudActionFlag indicates flagging for review
	FraudActionFlag FraudAction = "flag"

	// FraudActionBlock indicates blocking the verification
	FraudActionBlock FraudAction = "block"

	// FraudActionReportToAuthority indicates reporting to authorities
	FraudActionReportToAuthority FraudAction = "report_authority"

	// FraudActionSuspendAccount indicates account suspension
	FraudActionSuspendAccount FraudAction = "suspend_account"
)

// FraudConfig contains fraud detection configuration
type FraudConfig struct {
	// Enabled indicates if fraud detection is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// BlockOnHighSeverity blocks verification on high-severity fraud
	BlockOnHighSeverity bool `json:"block_on_high_severity" yaml:"block_on_high_severity"`

	// BlockOnCriticalSeverity blocks verification on critical fraud
	BlockOnCriticalSeverity bool `json:"block_on_critical_severity" yaml:"block_on_critical_severity"`

	// FlagThreshold is the fraud score threshold for flagging
	FlagThreshold float64 `json:"flag_threshold" yaml:"flag_threshold"`

	// BlockThreshold is the fraud score threshold for blocking
	BlockThreshold float64 `json:"block_threshold" yaml:"block_threshold"`

	// VelocityCheckWindow is the time window for velocity checks
	VelocityCheckWindow time.Duration `json:"velocity_check_window" yaml:"velocity_check_window"`

	// MaxVerificationsPerWindow is max verifications in velocity window
	MaxVerificationsPerWindow int `json:"max_verifications_per_window" yaml:"max_verifications_per_window"`

	// EnableDuplicateCheck enables duplicate document detection
	EnableDuplicateCheck bool `json:"enable_duplicate_check" yaml:"enable_duplicate_check"`

	// EnableBlacklistCheck enables document blacklist checking
	EnableBlacklistCheck bool `json:"enable_blacklist_check" yaml:"enable_blacklist_check"`

	// ReportCriticalToModule reports critical fraud to x/fraud module
	ReportCriticalToModule bool `json:"report_critical_to_module" yaml:"report_critical_to_module"`
}

// DefaultFraudConfig returns default fraud configuration
func DefaultFraudConfig() FraudConfig {
	return FraudConfig{
		Enabled:                   true,
		BlockOnHighSeverity:       false,
		BlockOnCriticalSeverity:   true,
		FlagThreshold:             0.5,
		BlockThreshold:            0.8,
		VelocityCheckWindow:       24 * time.Hour,
		MaxVerificationsPerWindow: 10,
		EnableDuplicateCheck:      true,
		EnableBlacklistCheck:      true,
		ReportCriticalToModule:    true,
	}
}

// FraudSignal represents a detected fraud signal
type FraudSignal struct {
	// Type is the type of fraud signal
	Type FraudSignalType `json:"type"`

	// Severity is the signal severity
	Severity FraudSeverity `json:"severity"`

	// Score is the signal score (0-1)
	Score float64 `json:"score"`

	// Description is a human-readable description
	Description string `json:"description"`

	// Evidence contains supporting evidence
	Evidence map[string]interface{} `json:"evidence,omitempty"`

	// DetectedAt is when the signal was detected
	DetectedAt time.Time `json:"detected_at"`

	// RecommendedAction is the recommended action
	RecommendedAction FraudAction `json:"recommended_action"`
}

// FraudCheckRequest represents a fraud check request
type FraudCheckRequest struct {
	// WalletAddress is the wallet being checked
	WalletAddress string `json:"wallet_address"`

	// DocumentType is the document type
	DocumentType DocumentType `json:"document_type"`

	// DocumentNumber is the document number (hashed)
	DocumentNumberHash string `json:"document_number_hash"`

	// Jurisdiction is the document jurisdiction
	Jurisdiction string `json:"jurisdiction"`

	// VerificationResult is the verification result to check
	VerificationResult *VerificationResponse `json:"verification_result,omitempty"`

	// LivenessResult is the liveness result to check
	LivenessResult *LivenessCheckResult `json:"liveness_result,omitempty"`

	// RequestID is the verification request ID
	RequestID string `json:"request_id"`

	// IPAddressHash is the hashed IP address
	IPAddressHash string `json:"ip_address_hash,omitempty"`

	// DeviceFingerprint is the device fingerprint
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`

	// Timestamp is when the check was requested
	Timestamp time.Time `json:"timestamp"`
}

// FraudCheckResult represents the result of a fraud check
type FraudCheckResult struct {
	// RequestID is the request ID
	RequestID string `json:"request_id"`

	// FraudScore is the overall fraud score (0-1, higher = more fraud risk)
	FraudScore float64 `json:"fraud_score"`

	// IsFraudulent indicates if fraud was detected
	IsFraudulent bool `json:"is_fraudulent"`

	// Signals contains detected fraud signals
	Signals []FraudSignal `json:"signals"`

	// Action is the action taken
	Action FraudAction `json:"action"`

	// Blocked indicates if verification was blocked
	Blocked bool `json:"blocked"`

	// ReportID is the fraud module report ID (if reported)
	ReportID string `json:"report_id,omitempty"`

	// Timestamp is when the check completed
	Timestamp time.Time `json:"timestamp"`

	// CheckDurationMs is the check duration
	CheckDurationMs int64 `json:"check_duration_ms"`
}

// ============================================================================
// Fraud Detector Interface
// ============================================================================

// FraudDetector defines the fraud detection interface
type FraudDetector interface {
	// CheckFraud performs a fraud check
	CheckFraud(ctx context.Context, req *FraudCheckRequest) (*FraudCheckResult, error)

	// ReportFraud reports fraud to the x/fraud module
	ReportFraud(ctx context.Context, result *FraudCheckResult, req *FraudCheckRequest) (string, error)

	// IsAvailable checks if the detector is available
	IsAvailable(ctx context.Context) bool

	// AddToBlacklist adds a document to the blacklist
	AddToBlacklist(ctx context.Context, documentNumberHash string, reason string) error

	// CheckBlacklist checks if a document is blacklisted
	CheckBlacklist(ctx context.Context, documentNumberHash string) (bool, error)
}

// FraudReporter interface for reporting to x/fraud module
type FraudReporter interface {
	// SubmitReport submits a fraud report to the blockchain module
	SubmitReport(ctx context.Context, report *FraudReport) (string, error)
}

// FraudReport represents a fraud report for the x/fraud module
type FraudReport struct {
	// Reporter is the reporting entity
	Reporter string `json:"reporter"`

	// ReportedParty is the wallet being reported
	ReportedParty string `json:"reported_party"`

	// Category is the fraud category
	Category string `json:"category"`

	// Description is the fraud description
	Description string `json:"description"`

	// Signals are the detected fraud signals
	Signals []FraudSignal `json:"signals"`

	// Evidence contains encrypted evidence
	Evidence []byte `json:"evidence,omitempty"`

	// Timestamp is when the report was created
	Timestamp time.Time `json:"timestamp"`
}

// ============================================================================
// Fraud Detection Implementation
// ============================================================================

// fraudHooks provides fraud detection integration
type fraudHooks struct {
	config            FraudConfig
	detector          FraudDetector
	reporter          FraudReporter
	velocityTracker   map[string]*velocityWindow // wallet -> velocity data
	documentBlacklist map[string]string          // hash -> reason
	mu                sync.RWMutex
}

type velocityWindow struct {
	windowStart time.Time
	count       int
}

// newFraudHooks creates a new fraud hooks instance
func newFraudHooks(config FraudConfig) *fraudHooks {
	return &fraudHooks{
		config:            config,
		velocityTracker:   make(map[string]*velocityWindow),
		documentBlacklist: make(map[string]string),
	}
}

// SetDetector sets the fraud detector implementation
func (f *fraudHooks) SetDetector(d FraudDetector) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.detector = d
}

// SetReporter sets the fraud reporter implementation
func (f *fraudHooks) SetReporter(r FraudReporter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reporter = r
}

// CheckFraud performs fraud detection on a verification
func (f *fraudHooks) CheckFraud(ctx context.Context, req *FraudCheckRequest) (*FraudCheckResult, error) {
	if !f.config.Enabled {
		return &FraudCheckResult{
			RequestID:  req.RequestID,
			FraudScore: 0,
			Timestamp:  time.Now(),
			Action:     FraudActionNone,
		}, nil
	}

	startTime := time.Now()
	result := &FraudCheckResult{
		RequestID: req.RequestID,
		Signals:   make([]FraudSignal, 0),
		Timestamp: startTime,
	}

	// Run fraud checks
	f.runVelocityCheck(req, result)
	f.runBlacklistCheck(req, result)
	f.runVerificationCheck(req, result)
	f.runLivenessCheck(req, result)

	// Use external detector if available
	f.mu.RLock()
	detector := f.detector
	f.mu.RUnlock()

	if detector != nil && detector.IsAvailable(ctx) {
		externalResult, err := detector.CheckFraud(ctx, req)
		if err == nil {
			result.Signals = append(result.Signals, externalResult.Signals...)
			if externalResult.FraudScore > result.FraudScore {
				result.FraudScore = externalResult.FraudScore
			}
		}
	}

	// Calculate overall fraud score
	result.FraudScore = f.calculateFraudScore(result.Signals)
	result.IsFraudulent = result.FraudScore >= f.config.FlagThreshold

	// Determine action
	result.Action = f.determineAction(result)
	result.Blocked = result.Action == FraudActionBlock

	// Report critical fraud
	if f.config.ReportCriticalToModule && f.hasCriticalSignal(result.Signals) {
		if reportID, err := f.reportToModule(ctx, result, req); err == nil {
			result.ReportID = reportID
		}
	}

	result.CheckDurationMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// runVelocityCheck checks for abnormal verification velocity
func (f *fraudHooks) runVelocityCheck(req *FraudCheckRequest, result *FraudCheckResult) {
	f.mu.Lock()
	defer f.mu.Unlock()

	window, exists := f.velocityTracker[req.WalletAddress]
	now := time.Now()

	if !exists || now.Sub(window.windowStart) > f.config.VelocityCheckWindow {
		// New window
		f.velocityTracker[req.WalletAddress] = &velocityWindow{
			windowStart: now,
			count:       1,
		}
		return
	}

	window.count++

	if window.count > f.config.MaxVerificationsPerWindow {
		result.Signals = append(result.Signals, FraudSignal{
			Type:              FraudSignalVelocityAnomaly,
			Severity:          FraudSeverityMedium,
			Score:             0.6,
			Description:       "Abnormal verification velocity detected",
			DetectedAt:        now,
			RecommendedAction: FraudActionFlag,
			Evidence: map[string]interface{}{
				"count":  window.count,
				"window": f.config.VelocityCheckWindow.String(),
				"limit":  f.config.MaxVerificationsPerWindow,
			},
		})
	}
}

// runBlacklistCheck checks document against blacklist
func (f *fraudHooks) runBlacklistCheck(req *FraudCheckRequest, result *FraudCheckResult) {
	if !f.config.EnableBlacklistCheck || req.DocumentNumberHash == "" {
		return
	}

	f.mu.RLock()
	reason, blacklisted := f.documentBlacklist[req.DocumentNumberHash]
	f.mu.RUnlock()

	if blacklisted {
		result.Signals = append(result.Signals, FraudSignal{
			Type:              FraudSignalBlacklistedDocument,
			Severity:          FraudSeverityCritical,
			Score:             1.0,
			Description:       "Document is on blocklist",
			DetectedAt:        time.Now(),
			RecommendedAction: FraudActionBlock,
			Evidence: map[string]interface{}{
				"reason": reason,
			},
		})
	}
}

// runVerificationCheck analyzes verification result for fraud signals
func (f *fraudHooks) runVerificationCheck(req *FraudCheckRequest, result *FraudCheckResult) {
	if req.VerificationResult == nil {
		return
	}

	vr := req.VerificationResult

	// Check for expired documents
	if vr.Status == VerificationStatusExpired {
		result.Signals = append(result.Signals, FraudSignal{
			Type:              FraudSignalExpiredDocument,
			Severity:          FraudSeverityMedium,
			Score:             0.5,
			Description:       "Expired document submitted for verification",
			DetectedAt:        time.Now(),
			RecommendedAction: FraudActionFlag,
		})
	}

	// Check for revoked documents
	if vr.Status == VerificationStatusRevoked {
		result.Signals = append(result.Signals, FraudSignal{
			Type:              FraudSignalRevokedDocument,
			Severity:          FraudSeverityHigh,
			Score:             0.8,
			Description:       "Revoked document submitted for verification",
			DetectedAt:        time.Now(),
			RecommendedAction: FraudActionBlock,
		})
	}

	// Check for low confidence (potential fake)
	if vr.Confidence < 0.5 && vr.Status.IsSuccess() {
		result.Signals = append(result.Signals, FraudSignal{
			Type:              FraudSignalFakeDocument,
			Severity:          FraudSeverityHigh,
			Score:             0.7,
			Description:       "Low confidence verification suggests potential fake document",
			DetectedAt:        time.Now(),
			RecommendedAction: FraudActionFlag,
			Evidence: map[string]interface{}{
				"confidence": vr.Confidence,
			},
		})
	}

	// Check for field mismatches
	mismatchCount := 0
	for _, field := range vr.FieldResults {
		if field.Match == FieldMatchNoMatch {
			mismatchCount++
		}
	}
	if mismatchCount > 2 {
		result.Signals = append(result.Signals, FraudSignal{
			Type:              FraudSignalMismatchedData,
			Severity:          FraudSeverityMedium,
			Score:             0.5,
			Description:       "Multiple field mismatches detected",
			DetectedAt:        time.Now(),
			RecommendedAction: FraudActionFlag,
			Evidence: map[string]interface{}{
				"mismatch_count": mismatchCount,
			},
		})
	}
}

// runLivenessCheck analyzes liveness result for fraud signals
func (f *fraudHooks) runLivenessCheck(req *FraudCheckRequest, result *FraudCheckResult) {
	if req.LivenessResult == nil {
		return
	}

	lr := req.LivenessResult

	// Check for spoofing
	if lr.SpoofDetected {
		severity := FraudSeverityHigh
		score := 0.8
		action := FraudActionBlock

		switch lr.SpoofType {
		case SpoofTypeDeepfake:
			severity = FraudSeverityCritical
			score = 1.0
		case SpoofTypeMask:
			severity = FraudSeverityHigh
			score = 0.9
		case SpoofTypePhoto, SpoofTypeScreen:
			severity = FraudSeverityMedium
			score = 0.7
			action = FraudActionFlag
		}

		result.Signals = append(result.Signals, FraudSignal{
			Type:              FraudSignalSpoofingAttempt,
			Severity:          severity,
			Score:             score,
			Description:       "Spoofing attempt detected: " + string(lr.SpoofType),
			DetectedAt:        time.Now(),
			RecommendedAction: action,
			Evidence: map[string]interface{}{
				"spoof_type":       lr.SpoofType,
				"spoof_confidence": lr.SpoofConfidence,
			},
		})
	}

	// Check for failed liveness with high-confidence fraud indicators
	if !lr.Passed && lr.Confidence < 0.3 {
		result.Signals = append(result.Signals, FraudSignal{
			Type:              FraudSignalSpoofingAttempt,
			Severity:          FraudSeverityMedium,
			Score:             0.6,
			Description:       "Failed liveness check with low confidence",
			DetectedAt:        time.Now(),
			RecommendedAction: FraudActionFlag,
		})
	}
}

// calculateFraudScore calculates overall fraud score from signals
func (f *fraudHooks) calculateFraudScore(signals []FraudSignal) float64 {
	if len(signals) == 0 {
		return 0
	}

	// Take the maximum signal score and add weighted contributions from others
	var maxScore float64
	var additionalScore float64

	for i, signal := range signals {
		if signal.Score > maxScore {
			maxScore = signal.Score
		}
		if i > 0 {
			// Add 10% of each additional signal
			additionalScore += signal.Score * 0.1
		}
	}

	total := maxScore + additionalScore
	if total > 1.0 {
		total = 1.0
	}

	return total
}

// determineAction determines the action based on fraud result
func (f *fraudHooks) determineAction(result *FraudCheckResult) FraudAction {
	// Check for critical signals
	for _, signal := range result.Signals {
		if signal.Severity == FraudSeverityCritical && f.config.BlockOnCriticalSeverity {
			return FraudActionBlock
		}
		if signal.Severity == FraudSeverityHigh && f.config.BlockOnHighSeverity {
			return FraudActionBlock
		}
	}

	// Check thresholds
	if result.FraudScore >= f.config.BlockThreshold {
		return FraudActionBlock
	}
	if result.FraudScore >= f.config.FlagThreshold {
		return FraudActionFlag
	}

	return FraudActionNone
}

// hasCriticalSignal checks if there are any critical signals
func (f *fraudHooks) hasCriticalSignal(signals []FraudSignal) bool {
	for _, signal := range signals {
		if signal.Severity == FraudSeverityCritical {
			return true
		}
	}
	return false
}

// reportToModule reports fraud to the x/fraud module
func (f *fraudHooks) reportToModule(ctx context.Context, result *FraudCheckResult, req *FraudCheckRequest) (string, error) {
	f.mu.RLock()
	reporter := f.reporter
	f.mu.RUnlock()

	if reporter == nil {
		return "", ErrFraudServiceUnavailable
	}

	report := &FraudReport{
		Reporter:      "govdata_service",
		ReportedParty: req.WalletAddress,
		Category:      "identity_fraud",
		Description:   "Automated fraud detection during government document verification",
		Signals:       result.Signals,
		Timestamp:     time.Now(),
	}

	return reporter.SubmitReport(ctx, report)
}

// AddToBlacklist adds a document to the blacklist
func (f *fraudHooks) AddToBlacklist(documentNumberHash string, reason string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.documentBlacklist[documentNumberHash] = reason
}

// ============================================================================
// Service Extension for Fraud Detection
// ============================================================================

// VerifyDocumentWithFraudCheck performs verification with fraud detection
func (s *service) VerifyDocumentWithFraudCheck(ctx context.Context, req *VerificationRequest, fraudReq *FraudCheckRequest) (*VerificationResponse, *FraudCheckResult, error) {
	// Perform document verification
	verifyResult, err := s.VerifyDocument(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	// Prepare fraud check request
	if fraudReq == nil {
		fraudReq = &FraudCheckRequest{
			WalletAddress: req.WalletAddress,
			DocumentType:  req.DocumentType,
			Jurisdiction:  req.Jurisdiction,
			RequestID:     req.RequestID,
			Timestamp:     time.Now(),
		}
	}
	fraudReq.VerificationResult = verifyResult

	// Perform fraud check using default implementation
	fraudHooks := newFraudHooks(DefaultFraudConfig())
	fraudResult, _ := fraudHooks.CheckFraud(ctx, fraudReq)

	// Add fraud info to verification warnings if flagged
	if fraudResult != nil && fraudResult.IsFraudulent {
		for _, signal := range fraudResult.Signals {
			verifyResult.Warnings = append(verifyResult.Warnings,
				"Fraud signal: "+signal.Description)
		}
	}

	return verifyResult, fraudResult, nil
}
