package security_monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// CryptoAnomalyConfig configures the cryptographic anomaly detector
type CryptoAnomalyConfig struct {
	// Signature verification thresholds
	MaxSignatureFailuresPerHour    int `json:"max_signature_failures_per_hour"`
	MaxSignatureFailuresPerAccount int `json:"max_signature_failures_per_account"`

	// Key operation thresholds
	MaxKeyOperationsPerMinute int `json:"max_key_operations_per_minute"`
	KeyOperationCooldownSecs  int `json:"key_operation_cooldown_secs"`

	// Entropy detection
	EnableEntropyAnalysis bool    `json:"enable_entropy_analysis"`
	MinEntropyThreshold   float64 `json:"min_entropy_threshold"`

	// Algorithm restrictions
	AllowedAlgorithms     []string `json:"allowed_algorithms"`
	DeprecatedAlgorithms  []string `json:"deprecated_algorithms"`

	// Key reuse detection
	EnableKeyReuseDetection bool `json:"enable_key_reuse_detection"`
	KeyHashWindowHours      int  `json:"key_hash_window_hours"`
}

// DefaultCryptoAnomalyConfig returns default configuration
func DefaultCryptoAnomalyConfig() *CryptoAnomalyConfig {
	return &CryptoAnomalyConfig{
		MaxSignatureFailuresPerHour:    50,
		MaxSignatureFailuresPerAccount: 10,

		MaxKeyOperationsPerMinute: 20,
		KeyOperationCooldownSecs:  5,

		EnableEntropyAnalysis: true,
		MinEntropyThreshold:   3.5, // bits per byte, typical secure random is ~8

		AllowedAlgorithms: []string{
			"X25519-XSalsa20-Poly1305",
			"Ed25519",
			"secp256k1",
			"P-256",
		},
		DeprecatedAlgorithms: []string{
			"RSA-1024",
			"MD5",
			"SHA1",
			"DES",
			"3DES",
		},

		EnableKeyReuseDetection: true,
		KeyHashWindowHours:      168, // 7 days
	}
}

// CryptoOperationData represents cryptographic operation data
type CryptoOperationData struct {
	OperationID    string                 `json:"operation_id"`
	OperationType  string                 `json:"operation_type"` // sign, verify, encrypt, decrypt, keygen
	Algorithm      string                 `json:"algorithm"`
	KeyID          string                 `json:"key_id,omitempty"`
	KeyFingerprint string                 `json:"key_fingerprint,omitempty"`
	AccountAddress string                 `json:"account_address"`
	Timestamp      time.Time              `json:"timestamp"`
	Success        bool                   `json:"success"`
	FailureReason  string                 `json:"failure_reason,omitempty"`
	DataSize       int64                  `json:"data_size,omitempty"`
	EntropyScore   float64                `json:"entropy_score,omitempty"`
	SourceIP       string                 `json:"source_ip,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CryptoAnomalyType represents types of cryptographic anomalies
type CryptoAnomalyType string

const (
	CryptoAnomalySignatureFailure    CryptoAnomalyType = "signature_failure"
	CryptoAnomalyWeakEntropy         CryptoAnomalyType = "weak_entropy"
	CryptoAnomalyKeyMisuse           CryptoAnomalyType = "key_misuse"
	CryptoAnomalyDeprecatedAlgorithm CryptoAnomalyType = "deprecated_algorithm"
	CryptoAnomalyUnauthorizedKey     CryptoAnomalyType = "unauthorized_key"
	CryptoAnomalyKeyReuse            CryptoAnomalyType = "key_reuse"
	CryptoAnomalyRapidOperations     CryptoAnomalyType = "rapid_operations"
	CryptoAnomalyDecryptionFailure   CryptoAnomalyType = "decryption_failure"
	CryptoAnomalyInvalidNonce        CryptoAnomalyType = "invalid_nonce"
)

// CryptoAnomalyDetector detects anomalies in cryptographic operations
type CryptoAnomalyDetector struct {
	config  *CryptoAnomalyConfig
	logger  zerolog.Logger
	metrics *SecurityMetrics

	// State tracking
	accountOperations    map[string][]cryptoOpRecord
	globalSignatureFailures []signatureFailureRecord
	seenKeyHashes        map[string]keyHashRecord
	mu                   sync.RWMutex

	// Event channel
	eventChan chan<- *SecurityEvent
	ctx       context.Context
}

type cryptoOpRecord struct {
	operationID string
	opType      string
	algorithm   string
	timestamp   time.Time
	success     bool
	keyID       string
}

type signatureFailureRecord struct {
	accountAddress string
	timestamp      time.Time
	reason         string
	keyFingerprint string
}

type keyHashRecord struct {
	keyFingerprint string
	accountAddress string
	firstUsedAt    time.Time
	lastUsedAt     time.Time
	useCount       int
}

// NewCryptoAnomalyDetector creates a new crypto anomaly detector
func NewCryptoAnomalyDetector(config *CryptoAnomalyConfig, logger zerolog.Logger) *CryptoAnomalyDetector {
	if config == nil {
		config = DefaultCryptoAnomalyConfig()
	}

	return &CryptoAnomalyDetector{
		config:                  config,
		logger:                  logger.With().Str("detector", "crypto-anomaly").Logger(),
		metrics:                 GetSecurityMetrics(),
		accountOperations:       make(map[string][]cryptoOpRecord),
		globalSignatureFailures: make([]signatureFailureRecord, 0),
		seenKeyHashes:           make(map[string]keyHashRecord),
	}
}

// Start starts the detector
func (d *CryptoAnomalyDetector) Start(ctx context.Context, eventChan chan<- *SecurityEvent) {
	d.ctx = ctx
	d.eventChan = eventChan

	// Start cleanup goroutine
	go d.cleanup(ctx)
}

// Analyze analyzes a cryptographic operation for anomalies
func (d *CryptoAnomalyDetector) Analyze(op *CryptoOperationData) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Record the operation
	record := cryptoOpRecord{
		operationID: op.OperationID,
		opType:      op.OperationType,
		algorithm:   op.Algorithm,
		timestamp:   op.Timestamp,
		success:     op.Success,
		keyID:       op.KeyID,
	}

	if _, exists := d.accountOperations[op.AccountAddress]; !exists {
		d.accountOperations[op.AccountAddress] = make([]cryptoOpRecord, 0)
	}
	d.accountOperations[op.AccountAddress] = append(
		d.accountOperations[op.AccountAddress], record)

	// Run all checks
	d.checkOperationSuccess(op)
	d.checkAlgorithm(op)
	d.checkEntropy(op)
	d.checkOperationVelocity(op)
	d.checkKeyReuse(op)
}

// checkOperationSuccess checks for cryptographic operation failures
func (d *CryptoAnomalyDetector) checkOperationSuccess(op *CryptoOperationData) {
	if op.Success {
		return
	}

	// Record failure
	d.globalSignatureFailures = append(d.globalSignatureFailures, signatureFailureRecord{
		accountAddress: op.AccountAddress,
		timestamp:      op.Timestamp,
		reason:         op.FailureReason,
		keyFingerprint: op.KeyFingerprint,
	})

	// Update metrics
	d.metrics.CryptoOperationFailures.WithLabelValues(op.OperationType, op.FailureReason).Inc()

	if op.OperationType == "verify" || op.OperationType == "sign" {
		d.metrics.CryptoSignatureFailures.WithLabelValues(op.Algorithm, op.FailureReason).Inc()
	}

	// Check account-specific failures
	accountHistory := d.accountOperations[op.AccountAddress]
	var accountFailures int
	hourAgo := op.Timestamp.Add(-1 * time.Hour)
	for _, rec := range accountHistory {
		if !rec.success && rec.timestamp.After(hourAgo) {
			accountFailures++
		}
	}

	if accountFailures >= d.config.MaxSignatureFailuresPerAccount {
		d.metrics.CryptoKeyMisuse.WithLabelValues("signature", "repeated_failure").Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(CryptoAnomalySignatureFailure),
			Severity:    SeverityHigh,
			Timestamp:   op.Timestamp,
			Source:      op.AccountAddress,
			Description: "Multiple cryptographic operation failures for account",
			Metadata: map[string]interface{}{
				"account":        op.AccountAddress,
				"failure_count":  accountFailures,
				"threshold":      d.config.MaxSignatureFailuresPerAccount,
				"operation_type": op.OperationType,
				"last_reason":    op.FailureReason,
			},
		})
	}

	// Check global failures (potential attack)
	var globalFailures int
	for _, rec := range d.globalSignatureFailures {
		if rec.timestamp.After(hourAgo) {
			globalFailures++
		}
	}

	if globalFailures >= d.config.MaxSignatureFailuresPerHour {
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(CryptoAnomalySignatureFailure),
			Severity:    SeverityCritical,
			Timestamp:   op.Timestamp,
			Source:      "system",
			Description: "High rate of cryptographic failures across system - potential attack",
			Metadata: map[string]interface{}{
				"failure_count": globalFailures,
				"threshold":     d.config.MaxSignatureFailuresPerHour,
				"window":        "1h",
			},
		})
	}
}

// checkAlgorithm checks for deprecated or unauthorized algorithms
func (d *CryptoAnomalyDetector) checkAlgorithm(op *CryptoOperationData) {
	// Check for deprecated algorithms
	for _, deprecated := range d.config.DeprecatedAlgorithms {
		if op.Algorithm == deprecated {
			d.metrics.CryptoAlgorithmMisuse.Inc()
			d.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        string(CryptoAnomalyDeprecatedAlgorithm),
				Severity:    SeverityHigh,
				Timestamp:   op.Timestamp,
				Source:      op.AccountAddress,
				Description: "Deprecated cryptographic algorithm used",
				Metadata: map[string]interface{}{
					"account":        op.AccountAddress,
					"algorithm":      op.Algorithm,
					"operation_type": op.OperationType,
				},
			})
			return
		}
	}

	// Check for unauthorized algorithms (not in allowed list)
	if len(d.config.AllowedAlgorithms) > 0 {
		allowed := false
		for _, alg := range d.config.AllowedAlgorithms {
			if op.Algorithm == alg {
				allowed = true
				break
			}
		}
		if !allowed {
			d.metrics.CryptoAlgorithmMisuse.Inc()
			d.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        string(CryptoAnomalyDeprecatedAlgorithm),
				Severity:    SeverityMedium,
				Timestamp:   op.Timestamp,
				Source:      op.AccountAddress,
				Description: "Unauthorized cryptographic algorithm used",
				Metadata: map[string]interface{}{
					"account":           op.AccountAddress,
					"algorithm":         op.Algorithm,
					"allowed_algorithms": d.config.AllowedAlgorithms,
				},
			})
		}
	}
}

// checkEntropy checks for weak entropy in cryptographic operations
func (d *CryptoAnomalyDetector) checkEntropy(op *CryptoOperationData) {
	if !d.config.EnableEntropyAnalysis {
		return
	}

	// Only check for key generation and encryption operations
	if op.OperationType != "keygen" && op.OperationType != "encrypt" {
		return
	}

	if op.EntropyScore > 0 && op.EntropyScore < d.config.MinEntropyThreshold {
		d.metrics.CryptoWeakEntropy.Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(CryptoAnomalyWeakEntropy),
			Severity:    SeverityCritical,
			Timestamp:   op.Timestamp,
			Source:      op.AccountAddress,
			Description: "Weak entropy detected in cryptographic operation",
			Metadata: map[string]interface{}{
				"account":        op.AccountAddress,
				"operation_type": op.OperationType,
				"entropy_score":  op.EntropyScore,
				"threshold":      d.config.MinEntropyThreshold,
				"algorithm":      op.Algorithm,
			},
		})
	}
}

// checkOperationVelocity checks for rapid cryptographic operations
func (d *CryptoAnomalyDetector) checkOperationVelocity(op *CryptoOperationData) {
	history := d.accountOperations[op.AccountAddress]

	// Count operations in last minute
	minuteAgo := op.Timestamp.Add(-1 * time.Minute)
	var recentOps int
	for _, rec := range history {
		if rec.timestamp.After(minuteAgo) {
			recentOps++
		}
	}

	if recentOps > d.config.MaxKeyOperationsPerMinute {
		d.metrics.CryptoKeyMisuse.WithLabelValues("any", "rapid_operations").Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        string(CryptoAnomalyRapidOperations),
			Severity:    SeverityMedium,
			Timestamp:   op.Timestamp,
			Source:      op.AccountAddress,
			Description: "Rapid cryptographic operations detected",
			Metadata: map[string]interface{}{
				"account":        op.AccountAddress,
				"operation_count": recentOps,
				"threshold":      d.config.MaxKeyOperationsPerMinute,
				"window":         "1m",
			},
		})
	}
}

// checkKeyReuse checks for inappropriate key reuse
func (d *CryptoAnomalyDetector) checkKeyReuse(op *CryptoOperationData) {
	if !d.config.EnableKeyReuseDetection || op.KeyFingerprint == "" {
		return
	}

	if existing, found := d.seenKeyHashes[op.KeyFingerprint]; found {
		// Same account using key is fine
		if existing.accountAddress != op.AccountAddress {
			d.metrics.CryptoKeyMisuse.WithLabelValues("any", "cross_account_reuse").Inc()
			d.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        string(CryptoAnomalyKeyReuse),
				Severity:    SeverityCritical,
				Timestamp:   op.Timestamp,
				Source:      op.AccountAddress,
				Description: "Cryptographic key used by multiple accounts",
				Metadata: map[string]interface{}{
					"current_account":  op.AccountAddress,
					"original_account": existing.accountAddress,
					"key_fingerprint":  op.KeyFingerprint[:16] + "...",
					"first_used":       existing.firstUsedAt,
				},
			})
		}

		// Update record
		existing.lastUsedAt = op.Timestamp
		existing.useCount++
		d.seenKeyHashes[op.KeyFingerprint] = existing
	} else {
		// New key
		d.seenKeyHashes[op.KeyFingerprint] = keyHashRecord{
			keyFingerprint: op.KeyFingerprint,
			accountAddress: op.AccountAddress,
			firstUsedAt:    op.Timestamp,
			lastUsedAt:     op.Timestamp,
			useCount:       1,
		}
	}
}

// RecordSignatureVerification records a signature verification result
func (d *CryptoAnomalyDetector) RecordSignatureVerification(
	account string,
	keyFingerprint string,
	algorithm string,
	success bool,
	failureReason string,
) {
	d.Analyze(&CryptoOperationData{
		OperationID:    generateEventID(),
		OperationType:  "verify",
		Algorithm:      algorithm,
		KeyFingerprint: keyFingerprint,
		AccountAddress: account,
		Timestamp:      time.Now(),
		Success:        success,
		FailureReason:  failureReason,
	})
}

// RecordEncryptionOperation records an encryption/decryption operation
func (d *CryptoAnomalyDetector) RecordEncryptionOperation(
	account string,
	keyFingerprint string,
	algorithm string,
	opType string, // "encrypt" or "decrypt"
	success bool,
	failureReason string,
	entropyScore float64,
) {
	d.Analyze(&CryptoOperationData{
		OperationID:    generateEventID(),
		OperationType:  opType,
		Algorithm:      algorithm,
		KeyFingerprint: keyFingerprint,
		AccountAddress: account,
		Timestamp:      time.Now(),
		Success:        success,
		FailureReason:  failureReason,
		EntropyScore:   entropyScore,
	})
}

// emitEvent sends an event to the security monitor
func (d *CryptoAnomalyDetector) emitEvent(event *SecurityEvent) {
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
func (d *CryptoAnomalyDetector) cleanup(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.mu.Lock()
			cutoff := time.Now().Add(-2 * time.Hour)
			keyHashCutoff := time.Now().Add(-time.Duration(d.config.KeyHashWindowHours) * time.Hour)

			// Clean up account operations
			for account, history := range d.accountOperations {
				newHistory := make([]cryptoOpRecord, 0)
				for _, rec := range history {
					if rec.timestamp.After(cutoff) {
						newHistory = append(newHistory, rec)
					}
				}
				if len(newHistory) == 0 {
					delete(d.accountOperations, account)
				} else {
					d.accountOperations[account] = newHistory
				}
			}

			// Clean up global failures
			newFailures := make([]signatureFailureRecord, 0)
			for _, rec := range d.globalSignatureFailures {
				if rec.timestamp.After(cutoff) {
					newFailures = append(newFailures, rec)
				}
			}
			d.globalSignatureFailures = newFailures

			// Clean up key hashes
			for hash, rec := range d.seenKeyHashes {
				if rec.lastUsedAt.Before(keyHashCutoff) {
					delete(d.seenKeyHashes, hash)
				}
			}

			d.mu.Unlock()
		}
	}
}
