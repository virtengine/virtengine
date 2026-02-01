package security_monitoring

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// FraudDetectorConfig configures the fraud detector
type FraudDetectorConfig struct {
	// Verification attempt thresholds
	MaxVerificationAttemptsPerAccount int `json:"max_verification_attempts_per_account"`
	MaxFailedVerificationsBeforeFlag  int `json:"max_failed_verifications_before_flag"`
	VerificationCooldownMins          int `json:"verification_cooldown_mins"`

	// Score anomaly detection
	MinExpectedScore             uint32  `json:"min_expected_score"`
	ScoreVarianceThreshold       float64 `json:"score_variance_threshold"`
	ConsecutiveLowScoreThreshold int     `json:"consecutive_low_score_threshold"`

	// Biometric thresholds
	BiometricMismatchThreshold float64 `json:"biometric_mismatch_threshold"`
	FaceSimilarityMinimum      float64 `json:"face_similarity_minimum"`

	// Document validation
	EnableDocumentForensics bool `json:"enable_document_forensics"`
	OCRConfidenceMinimum    float64 `json:"ocr_confidence_minimum"`

	// Replay detection
	ScopeHashWindowHours int `json:"scope_hash_window_hours"`

	// Behavioral analysis
	EnableBehavioralAnalysis bool `json:"enable_behavioral_analysis"`
}

// DefaultFraudDetectorConfig returns default configuration
func DefaultFraudDetectorConfig() *FraudDetectorConfig {
	return &FraudDetectorConfig{
		MaxVerificationAttemptsPerAccount: 5,
		MaxFailedVerificationsBeforeFlag:  3,
		VerificationCooldownMins:          60,

		MinExpectedScore:             60,
		ScoreVarianceThreshold:       0.3,
		ConsecutiveLowScoreThreshold: 3,

		BiometricMismatchThreshold: 0.5,
		FaceSimilarityMinimum:      0.8,

		EnableDocumentForensics: true,
		OCRConfidenceMinimum:    0.85,

		ScopeHashWindowHours: 24,

		EnableBehavioralAnalysis: true,
	}
}

// VEIDVerificationData represents VEID verification data for analysis
type VEIDVerificationData struct {
	RequestID         string                 `json:"request_id"`
	AccountAddress    string                 `json:"account_address"`
	Timestamp         time.Time              `json:"timestamp"`
	BlockHeight       int64                  `json:"block_height"`

	// Scores
	ProposerScore     uint32                 `json:"proposer_score"`
	ComputedScore     uint32                 `json:"computed_score"`
	ScoreDifference   int32                  `json:"score_difference"`
	Match             bool                   `json:"match"`

	// Biometric data (metadata only, never actual biometrics)
	FaceSimilarityScore float64              `json:"face_similarity_score,omitempty"`
	LivenessScore       float64              `json:"liveness_score,omitempty"`
	BiometricHash       string               `json:"biometric_hash,omitempty"`

	// Document data
	DocumentType      string                 `json:"document_type,omitempty"`
	OCRConfidence     float64                `json:"ocr_confidence,omitempty"`
	DocumentHash      string                 `json:"document_hash,omitempty"`
	DocumentCountry   string                 `json:"document_country,omitempty"`

	// Scope data
	ScopeHashes       []string               `json:"scope_hashes,omitempty"`
	ScopeCount        int                    `json:"scope_count,omitempty"`

	// Result
	Success           bool                   `json:"success"`
	FailureReason     string                 `json:"failure_reason,omitempty"`
	ReasonCodes       []string               `json:"reason_codes,omitempty"`

	// Source information
	ClientAppID       string                 `json:"client_app_id,omitempty"`
	SourceIP          string                 `json:"source_ip,omitempty"`

	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// FraudIndicatorType represents types of fraud indicators
type FraudIndicatorType string

const (
	FraudIndicatorDocumentTampering    FraudIndicatorType = "document_tampering"
	FraudIndicatorBiometricMismatch    FraudIndicatorType = "biometric_mismatch"
	FraudIndicatorReplayAttack         FraudIndicatorType = "replay_attack"
	FraudIndicatorScoreAnomaly         FraudIndicatorType = "score_anomaly"
	FraudIndicatorVelocityAbuse        FraudIndicatorType = "velocity_abuse"
	FraudIndicatorSyntheticIdentity    FraudIndicatorType = "synthetic_identity"
	FraudIndicatorDocumentForgery      FraudIndicatorType = "document_forgery"
	FraudIndicatorLivenessFailure      FraudIndicatorType = "liveness_failure"
	FraudIndicatorMultipleIdentities   FraudIndicatorType = "multiple_identities"
	FraudIndicatorSuspiciousBehavior   FraudIndicatorType = "suspicious_behavior"
)

// FraudDetector detects VEID fraud indicators
type FraudDetector struct {
	config  *FraudDetectorConfig
	logger  zerolog.Logger
	metrics *SecurityMetrics

	// State tracking
	accountVerificationHistory map[string][]verificationRecord
	seenScopeHashes            map[string]scopeHashRecord
	biometricHashHistory       map[string][]string // hash -> accounts that used it
	mu                         sync.RWMutex

	// Event channel
	eventChan chan<- *SecurityEvent
	ctx       context.Context
}

type verificationRecord struct {
	requestID    string
	timestamp    time.Time
	score        uint32
	success      bool
	biometricHash string
	documentHash string
}

type scopeHashRecord struct {
	hash          string
	accountAddress string
	firstSeenAt   time.Time
	lastSeenAt    time.Time
	seenCount     int
}

// NewFraudDetector creates a new fraud detector
func NewFraudDetector(config *FraudDetectorConfig, logger zerolog.Logger) *FraudDetector {
	if config == nil {
		config = DefaultFraudDetectorConfig()
	}

	return &FraudDetector{
		config:                     config,
		logger:                     logger.With().Str("detector", "fraud").Logger(),
		metrics:                    GetSecurityMetrics(),
		accountVerificationHistory: make(map[string][]verificationRecord),
		seenScopeHashes:            make(map[string]scopeHashRecord),
		biometricHashHistory:       make(map[string][]string),
	}
}

// Start starts the detector
func (d *FraudDetector) Start(ctx context.Context, eventChan chan<- *SecurityEvent) {
	d.ctx = ctx
	d.eventChan = eventChan

	// Start cleanup goroutine
	go d.cleanup(ctx)
}

// Analyze analyzes a VEID verification for fraud indicators
func (d *FraudDetector) Analyze(v *VEIDVerificationData) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Record the verification
	record := verificationRecord{
		requestID:     v.RequestID,
		timestamp:     v.Timestamp,
		score:         v.ComputedScore,
		success:       v.Success,
		biometricHash: v.BiometricHash,
		documentHash:  v.DocumentHash,
	}

	if _, exists := d.accountVerificationHistory[v.AccountAddress]; !exists {
		d.accountVerificationHistory[v.AccountAddress] = make([]verificationRecord, 0)
	}
	d.accountVerificationHistory[v.AccountAddress] = append(
		d.accountVerificationHistory[v.AccountAddress], record)

	// Run all fraud detection checks
	d.checkVerificationVelocity(v)
	d.checkReplayAttack(v)
	d.checkBiometricMismatch(v)
	d.checkDocumentFraud(v)
	d.checkScoreAnomalies(v)
	d.checkMultipleIdentities(v)
	d.checkLivenessFailure(v)
	d.checkFailedVerificationPattern(v)

	// Update verification metrics
	if !v.Success {
		d.metrics.VEIDVerificationFailures.WithLabelValues(v.FailureReason).Inc()
	}
}

// checkVerificationVelocity checks for excessive verification attempts
func (d *FraudDetector) checkVerificationVelocity(v *VEIDVerificationData) {
	history := d.accountVerificationHistory[v.AccountAddress]

	// Count recent attempts
	windowStart := v.Timestamp.Add(-time.Duration(d.config.VerificationCooldownMins) * time.Minute)
	var recentCount int
	for _, rec := range history {
		if rec.timestamp.After(windowStart) {
			recentCount++
		}
	}

	if recentCount > d.config.MaxVerificationAttemptsPerAccount {
		d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorVelocityAbuse), "high").Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(FraudIndicatorVelocityAbuse),
			Severity:    SeverityHigh,
			Timestamp:   v.Timestamp,
			Source:      v.AccountAddress,
			Description: "Excessive verification attempts detected",
			Metadata: map[string]interface{}{
				"account":         v.AccountAddress,
				"attempt_count":   recentCount,
				"threshold":       d.config.MaxVerificationAttemptsPerAccount,
				"window_mins":     d.config.VerificationCooldownMins,
				"request_id":      v.RequestID,
			},
		})
	}
}

// checkReplayAttack checks for scope hash replay attacks
func (d *FraudDetector) checkReplayAttack(v *VEIDVerificationData) {
	for _, scopeHash := range v.ScopeHashes {
		if existing, found := d.seenScopeHashes[scopeHash]; found {
			// Same account resubmitting is suspicious but not necessarily an attack
			if existing.accountAddress != v.AccountAddress {
				d.metrics.VEIDReplayAttempts.Inc()
				d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorReplayAttack), "critical").Inc()
				d.emitEvent(&SecurityEvent{
					ID:          generateEventID(),
					Type:        string(FraudIndicatorReplayAttack),
					Severity:    SeverityCritical,
					Timestamp:   v.Timestamp,
					Source:      v.AccountAddress,
					Description: "Scope hash replay attack detected - same scope used by different account",
					Metadata: map[string]interface{}{
						"account":          v.AccountAddress,
						"original_account": existing.accountAddress,
						"scope_hash":       scopeHash[:16] + "...", // Truncate for logging
						"original_time":    existing.firstSeenAt,
						"request_id":       v.RequestID,
					},
				})
			} else if existing.seenCount > 3 {
				// Same account resubmitting multiple times
				d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorReplayAttack), "medium").Inc()
				d.emitEvent(&SecurityEvent{
					ID:          generateEventID(),
					Type:        string(FraudIndicatorReplayAttack),
					Severity:    SeverityMedium,
					Timestamp:   v.Timestamp,
					Source:      v.AccountAddress,
					Description: "Repeated scope submission detected",
					Metadata: map[string]interface{}{
						"account":     v.AccountAddress,
						"seen_count":  existing.seenCount,
						"request_id":  v.RequestID,
					},
				})
			}

			// Update record
			existing.lastSeenAt = v.Timestamp
			existing.seenCount++
			d.seenScopeHashes[scopeHash] = existing
		} else {
			// New scope hash
			d.seenScopeHashes[scopeHash] = scopeHashRecord{
				hash:           scopeHash,
				accountAddress: v.AccountAddress,
				firstSeenAt:    v.Timestamp,
				lastSeenAt:     v.Timestamp,
				seenCount:      1,
			}
		}
	}
}

// checkBiometricMismatch checks for biometric verification failures
func (d *FraudDetector) checkBiometricMismatch(v *VEIDVerificationData) {
	// Check face similarity score
	if v.FaceSimilarityScore > 0 && v.FaceSimilarityScore < d.config.FaceSimilarityMinimum {
		d.metrics.VEIDBiometricMismatches.Inc()
		d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorBiometricMismatch), "high").Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(FraudIndicatorBiometricMismatch),
			Severity:    SeverityHigh,
			Timestamp:   v.Timestamp,
			Source:      v.AccountAddress,
			Description: "Biometric face similarity below threshold",
			Metadata: map[string]interface{}{
				"account":          v.AccountAddress,
				"similarity_score": v.FaceSimilarityScore,
				"threshold":        d.config.FaceSimilarityMinimum,
				"request_id":       v.RequestID,
			},
		})
	}

	// Check for same biometric used by different accounts
	if v.BiometricHash != "" {
		accounts, exists := d.biometricHashHistory[v.BiometricHash]
		if exists {
			for _, existingAccount := range accounts {
				if existingAccount != v.AccountAddress {
					d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorMultipleIdentities), "critical").Inc()
					d.emitEvent(&SecurityEvent{
						ID:          generateEventID(),
						Type:        string(FraudIndicatorMultipleIdentities),
						Severity:    SeverityCritical,
						Timestamp:   v.Timestamp,
						Source:      v.AccountAddress,
						Description: "Same biometric used by multiple accounts",
						Metadata: map[string]interface{}{
							"account":          v.AccountAddress,
							"existing_account": existingAccount,
							"biometric_hash":   v.BiometricHash[:16] + "...",
							"request_id":       v.RequestID,
						},
					})
					break
				}
			}
		}

		// Add this account to the history
		d.biometricHashHistory[v.BiometricHash] = append(accounts, v.AccountAddress)
	}
}

// checkDocumentFraud checks for document fraud indicators
func (d *FraudDetector) checkDocumentFraud(v *VEIDVerificationData) {
	// Check OCR confidence
	if v.OCRConfidence > 0 && v.OCRConfidence < d.config.OCRConfidenceMinimum {
		d.metrics.VEIDDocumentForgery.Inc()
		d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorDocumentForgery), "high").Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(FraudIndicatorDocumentForgery),
			Severity:    SeverityHigh,
			Timestamp:   v.Timestamp,
			Source:      v.AccountAddress,
			Description: "Document OCR confidence below threshold - possible forgery",
			Metadata: map[string]interface{}{
				"account":        v.AccountAddress,
				"document_type":  v.DocumentType,
				"ocr_confidence": v.OCRConfidence,
				"threshold":      d.config.OCRConfidenceMinimum,
				"request_id":     v.RequestID,
			},
		})
	}

	// Check for tampering indicators in reason codes
	for _, code := range v.ReasonCodes {
		if code == "DOCUMENT_TAMPERED" || code == "DOCUMENT_MODIFIED" || code == "METADATA_MISMATCH" {
			d.metrics.VEIDTamperingAttempts.Inc()
			d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorDocumentTampering), "critical").Inc()
			d.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        string(FraudIndicatorDocumentTampering),
				Severity:    SeverityCritical,
				Timestamp:   v.Timestamp,
				Source:      v.AccountAddress,
				Description: "Document tampering detected",
				Metadata: map[string]interface{}{
					"account":       v.AccountAddress,
					"document_type": v.DocumentType,
					"reason_code":   code,
					"request_id":    v.RequestID,
				},
			})
			break
		}
	}
}

// checkScoreAnomalies checks for verification score anomalies
func (d *FraudDetector) checkScoreAnomalies(v *VEIDVerificationData) {
	// Check for low scores
	if v.ComputedScore < d.config.MinExpectedScore {
		d.metrics.VEIDScoreAnomalies.Inc()

		// Check for consecutive low scores
		history := d.accountVerificationHistory[v.AccountAddress]
		var consecutiveLow int
		for i := len(history) - 1; i >= 0 && consecutiveLow < d.config.ConsecutiveLowScoreThreshold; i-- {
			if history[i].score < d.config.MinExpectedScore {
				consecutiveLow++
			} else {
				break
			}
		}

		if consecutiveLow >= d.config.ConsecutiveLowScoreThreshold {
			d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorScoreAnomaly), "high").Inc()
			d.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        string(FraudIndicatorScoreAnomaly),
				Severity:    SeverityHigh,
				Timestamp:   v.Timestamp,
				Source:      v.AccountAddress,
				Description: "Consecutive low verification scores detected",
				Metadata: map[string]interface{}{
					"account":            v.AccountAddress,
					"current_score":      v.ComputedScore,
					"consecutive_count":  consecutiveLow,
					"threshold":          d.config.MinExpectedScore,
					"request_id":         v.RequestID,
				},
			})
		}
	}

	// Check for significant score difference between proposer and computed
	if !v.Match && v.ScoreDifference != 0 {
		variance := float64(abs32(v.ScoreDifference)) / float64(v.ProposerScore+1)
		if variance > d.config.ScoreVarianceThreshold {
			d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorScoreAnomaly), "medium").Inc()
			d.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        string(FraudIndicatorScoreAnomaly),
				Severity:    SeverityMedium,
				Timestamp:   v.Timestamp,
				Source:      v.AccountAddress,
				Description: "Significant score variance detected between validators",
				Metadata: map[string]interface{}{
					"account":          v.AccountAddress,
					"proposer_score":   v.ProposerScore,
					"computed_score":   v.ComputedScore,
					"score_difference": v.ScoreDifference,
					"variance":         variance,
					"request_id":       v.RequestID,
				},
			})
		}
	}
}

// checkMultipleIdentities checks for synthetic identity indicators
func (d *FraudDetector) checkMultipleIdentities(v *VEIDVerificationData) {
	// This check is primarily handled in biometric check, but we can add
	// document-based detection here
	if v.DocumentHash != "" {
		// Check if same document used by different accounts
		for account, history := range d.accountVerificationHistory {
			if account != v.AccountAddress {
				for _, rec := range history {
					if rec.documentHash == v.DocumentHash {
						d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorSyntheticIdentity), "critical").Inc()
						d.emitEvent(&SecurityEvent{
							ID:          generateEventID(),
							Type:        string(FraudIndicatorSyntheticIdentity),
							Severity:    SeverityCritical,
							Timestamp:   v.Timestamp,
							Source:      v.AccountAddress,
							Description: "Same document used by multiple accounts - synthetic identity suspected",
							Metadata: map[string]interface{}{
								"account":          v.AccountAddress,
								"existing_account": account,
								"document_type":    v.DocumentType,
								"request_id":       v.RequestID,
							},
						})
						return
					}
				}
			}
		}
	}
}

// checkLivenessFailure checks for liveness detection failures
func (d *FraudDetector) checkLivenessFailure(v *VEIDVerificationData) {
	if v.LivenessScore > 0 && v.LivenessScore < 0.5 {
		d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorLivenessFailure), "high").Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(FraudIndicatorLivenessFailure),
			Severity:    SeverityHigh,
			Timestamp:   v.Timestamp,
			Source:      v.AccountAddress,
			Description: "Liveness detection failure - potential presentation attack",
			Metadata: map[string]interface{}{
				"account":        v.AccountAddress,
				"liveness_score": v.LivenessScore,
				"request_id":     v.RequestID,
			},
		})
	}
}

// checkFailedVerificationPattern checks for patterns in failed verifications
func (d *FraudDetector) checkFailedVerificationPattern(v *VEIDVerificationData) {
	history := d.accountVerificationHistory[v.AccountAddress]

	var recentFailures int
	cutoff := v.Timestamp.Add(-1 * time.Hour)
	for _, rec := range history {
		if rec.timestamp.After(cutoff) && !rec.success {
			recentFailures++
		}
	}

	if recentFailures >= d.config.MaxFailedVerificationsBeforeFlag {
		d.metrics.VEIDFraudIndicators.WithLabelValues(string(FraudIndicatorSuspiciousBehavior), "medium").Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(FraudIndicatorSuspiciousBehavior),
			Severity:    SeverityMedium,
			Timestamp:   v.Timestamp,
			Source:      v.AccountAddress,
			Description: "Multiple failed verification attempts",
			Metadata: map[string]interface{}{
				"account":       v.AccountAddress,
				"failure_count": recentFailures,
				"threshold":     d.config.MaxFailedVerificationsBeforeFlag,
				"window":        "1h",
				"request_id":    v.RequestID,
			},
		})
	}
}

// emitEvent sends an event to the security monitor
func (d *FraudDetector) emitEvent(event *SecurityEvent) {
	if d.eventChan == nil {
		return
	}

	select {
	case d.eventChan <- event:
	default:
		d.logger.Warn().Str("event_id", event.ID).Msg("event channel full, dropping event")
	}
}

// cleanup periodically cleans up old history
func (d *FraudDetector) cleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.mu.Lock()
			cutoff := time.Now().Add(-time.Duration(d.config.ScopeHashWindowHours) * time.Hour)

			// Clean up verification history
			for account, history := range d.accountVerificationHistory {
				newHistory := make([]verificationRecord, 0)
				for _, rec := range history {
					if rec.timestamp.After(cutoff) {
						newHistory = append(newHistory, rec)
					}
				}
				if len(newHistory) == 0 {
					delete(d.accountVerificationHistory, account)
				} else {
					d.accountVerificationHistory[account] = newHistory
				}
			}

			// Clean up scope hashes
			for hash, record := range d.seenScopeHashes {
				if record.lastSeenAt.Before(cutoff) {
					delete(d.seenScopeHashes, hash)
				}
			}

			d.mu.Unlock()
		}
	}
}

// Helper function for absolute value of int32
func abs32(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

// hashData creates a SHA256 hash of data for comparison
func hashData(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
