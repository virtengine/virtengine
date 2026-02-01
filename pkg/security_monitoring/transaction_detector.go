package security_monitoring

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// TransactionDetectorConfig configures the transaction detector
type TransactionDetectorConfig struct {
	// Velocity thresholds
	MaxTxPerMinutePerAccount int     `json:"max_tx_per_minute_per_account"`
	MaxTxPerHourPerAccount   int     `json:"max_tx_per_hour_per_account"`
	MaxTxValueThreshold      float64 `json:"max_tx_value_threshold"`

	// Pattern detection
	RapidFireWindowSecs      int `json:"rapid_fire_window_secs"`
	RapidFireThreshold       int `json:"rapid_fire_threshold"`
	SplitTransactionWindow   int `json:"split_transaction_window_secs"`
	SplitTransactionThreshold int `json:"split_transaction_threshold"`

	// Anomaly detection
	EnableMLAnomalyDetection bool    `json:"enable_ml_anomaly_detection"`
	AnomalyScoreThreshold    float64 `json:"anomaly_score_threshold"`

	// Account behavior
	NewAccountCooldownMins int `json:"new_account_cooldown_mins"`
	HighRiskAccountTxLimit int `json:"high_risk_account_tx_limit"`
}

// DefaultTransactionDetectorConfig returns default configuration
func DefaultTransactionDetectorConfig() *TransactionDetectorConfig {
	return &TransactionDetectorConfig{
		MaxTxPerMinutePerAccount:  20,
		MaxTxPerHourPerAccount:    200,
		MaxTxValueThreshold:       1000000, // In smallest denomination

		RapidFireWindowSecs:       10,
		RapidFireThreshold:        5,
		SplitTransactionWindow:    60,
		SplitTransactionThreshold: 10,

		EnableMLAnomalyDetection: false,
		AnomalyScoreThreshold:    0.8,

		NewAccountCooldownMins: 60,
		HighRiskAccountTxLimit: 5,
	}
}

// TransactionData represents a transaction for analysis
type TransactionData struct {
	TxHash          string                 `json:"tx_hash"`
	Sender          string                 `json:"sender"`
	Recipient       string                 `json:"recipient"`
	Amount          float64                `json:"amount"`
	Denom           string                 `json:"denom"`
	MsgType         string                 `json:"msg_type"`
	Timestamp       time.Time              `json:"timestamp"`
	BlockHeight     int64                  `json:"block_height"`
	GasUsed         int64                  `json:"gas_used"`
	Success         bool                   `json:"success"`
	Memo            string                 `json:"memo,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	AccountAge      time.Duration          `json:"account_age"`
	IsHighRiskAccount bool                 `json:"is_high_risk_account"`
}

// TransactionDetector detects suspicious transaction patterns
type TransactionDetector struct {
	config  *TransactionDetectorConfig
	logger  zerolog.Logger
	metrics *SecurityMetrics

	// State tracking
	accountTxHistory map[string][]txRecord
	recentTxHashes   map[string]time.Time // For replay detection
	mu               sync.RWMutex

	// Event channel
	eventChan chan<- *SecurityEvent
	ctx       context.Context
}

type txRecord struct {
	hash      string
	timestamp time.Time
	amount    float64
	recipient string
	msgType   string
}

// NewTransactionDetector creates a new transaction detector
func NewTransactionDetector(config *TransactionDetectorConfig, logger zerolog.Logger) *TransactionDetector {
	if config == nil {
		config = DefaultTransactionDetectorConfig()
	}

	return &TransactionDetector{
		config:           config,
		logger:           logger.With().Str("detector", "transaction").Logger(),
		metrics:          GetSecurityMetrics(),
		accountTxHistory: make(map[string][]txRecord),
		recentTxHashes:   make(map[string]time.Time),
	}
}

// Start starts the detector
func (d *TransactionDetector) Start(ctx context.Context, eventChan chan<- *SecurityEvent) {
	d.ctx = ctx
	d.eventChan = eventChan

	// Start cleanup goroutine
	go d.cleanup(ctx)
}

// Analyze analyzes a transaction for suspicious patterns
func (d *TransactionDetector) Analyze(tx *TransactionData) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Record the transaction
	record := txRecord{
		hash:      tx.TxHash,
		timestamp: tx.Timestamp,
		amount:    tx.Amount,
		recipient: tx.Recipient,
		msgType:   tx.MsgType,
	}

	if _, exists := d.accountTxHistory[tx.Sender]; !exists {
		d.accountTxHistory[tx.Sender] = make([]txRecord, 0)
	}
	d.accountTxHistory[tx.Sender] = append(d.accountTxHistory[tx.Sender], record)

	// Run all detection checks
	d.checkReplayAttack(tx)
	d.checkVelocityViolation(tx)
	d.checkRapidFirePattern(tx)
	d.checkSplitTransactionPattern(tx)
	d.checkValueAnomaly(tx)
	d.checkNewAccountAbuse(tx)
	d.checkHighRiskAccount(tx)
}

// checkReplayAttack checks for transaction replay attempts
func (d *TransactionDetector) checkReplayAttack(tx *TransactionData) {
	if prevTime, exists := d.recentTxHashes[tx.TxHash]; exists {
		d.metrics.TxReplayAttempts.Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        "tx_replay_attempt",
			Severity:    SeverityCritical,
			Timestamp:   tx.Timestamp,
			Source:      tx.Sender,
			Description: "Transaction replay attempt detected",
			Metadata: map[string]interface{}{
				"tx_hash":        tx.TxHash,
				"original_time":  prevTime,
				"replay_time":    tx.Timestamp,
				"sender":         tx.Sender,
			},
		})
	}
	d.recentTxHashes[tx.TxHash] = tx.Timestamp
}

// checkVelocityViolation checks for transaction rate violations
func (d *TransactionDetector) checkVelocityViolation(tx *TransactionData) {
	history := d.accountTxHistory[tx.Sender]

	// Count transactions in last minute
	minuteAgo := tx.Timestamp.Add(-1 * time.Minute)
	hourAgo := tx.Timestamp.Add(-1 * time.Hour)

	var txInMinute, txInHour int
	for _, rec := range history {
		if rec.timestamp.After(minuteAgo) {
			txInMinute++
		}
		if rec.timestamp.After(hourAgo) {
			txInHour++
		}
	}

	// Check minute threshold
	if txInMinute > d.config.MaxTxPerMinutePerAccount {
		d.metrics.TxAnomaliesDetected.WithLabelValues("velocity_minute", "high").Inc()
		d.metrics.TxVelocityRate.WithLabelValues("per_minute").Set(float64(txInMinute))
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        "tx_velocity_violation",
			Severity:    SeverityHigh,
			Timestamp:   tx.Timestamp,
			Source:      tx.Sender,
			Description: "Transaction velocity exceeded per-minute threshold",
			Metadata: map[string]interface{}{
				"sender":      tx.Sender,
				"tx_count":    txInMinute,
				"threshold":   d.config.MaxTxPerMinutePerAccount,
				"window":      "1m",
			},
		})
	}

	// Check hour threshold
	if txInHour > d.config.MaxTxPerHourPerAccount {
		d.metrics.TxAnomaliesDetected.WithLabelValues("velocity_hour", "medium").Inc()
		d.metrics.TxVelocityRate.WithLabelValues("per_hour").Set(float64(txInHour))
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        "tx_velocity_violation",
			Severity:    SeverityMedium,
			Timestamp:   tx.Timestamp,
			Source:      tx.Sender,
			Description: "Transaction velocity exceeded per-hour threshold",
			Metadata: map[string]interface{}{
				"sender":    tx.Sender,
				"tx_count":  txInHour,
				"threshold": d.config.MaxTxPerHourPerAccount,
				"window":    "1h",
			},
		})
	}
}

// checkRapidFirePattern checks for rapid-fire transactions
func (d *TransactionDetector) checkRapidFirePattern(tx *TransactionData) {
	history := d.accountTxHistory[tx.Sender]

	windowStart := tx.Timestamp.Add(-time.Duration(d.config.RapidFireWindowSecs) * time.Second)
	var count int
	for _, rec := range history {
		if rec.timestamp.After(windowStart) {
			count++
		}
	}

	if count >= d.config.RapidFireThreshold {
		d.metrics.TxSuspiciousPatterns.WithLabelValues("rapid_fire").Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        "tx_rapid_fire",
			Severity:    SeverityHigh,
			Timestamp:   tx.Timestamp,
			Source:      tx.Sender,
			Description: "Rapid-fire transaction pattern detected",
			Metadata: map[string]interface{}{
				"sender":      tx.Sender,
				"tx_count":    count,
				"window_secs": d.config.RapidFireWindowSecs,
				"threshold":   d.config.RapidFireThreshold,
			},
		})
	}
}

// checkSplitTransactionPattern checks for transaction splitting (structuring)
func (d *TransactionDetector) checkSplitTransactionPattern(tx *TransactionData) {
	history := d.accountTxHistory[tx.Sender]

	windowStart := tx.Timestamp.Add(-time.Duration(d.config.SplitTransactionWindow) * time.Second)

	// Group transactions by recipient
	recipientTxCounts := make(map[string]int)
	recipientTotalAmounts := make(map[string]float64)

	for _, rec := range history {
		if rec.timestamp.After(windowStart) {
			recipientTxCounts[rec.recipient]++
			recipientTotalAmounts[rec.recipient] += rec.amount
		}
	}

	// Check for split pattern to same recipient
	for recipient, count := range recipientTxCounts {
		if count >= d.config.SplitTransactionThreshold {
			totalAmount := recipientTotalAmounts[recipient]
			d.metrics.TxSuspiciousPatterns.WithLabelValues("split_transaction").Inc()
			d.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        "tx_split_pattern",
				Severity:    SeverityMedium,
				Timestamp:   tx.Timestamp,
				Source:      tx.Sender,
				Description: "Transaction splitting pattern detected",
				Metadata: map[string]interface{}{
					"sender":       tx.Sender,
					"recipient":    recipient,
					"tx_count":     count,
					"total_amount": totalAmount,
					"window_secs":  d.config.SplitTransactionWindow,
				},
			})
		}
	}
}

// checkValueAnomaly checks for unusual transaction values
func (d *TransactionDetector) checkValueAnomaly(tx *TransactionData) {
	// Check absolute threshold
	if tx.Amount > d.config.MaxTxValueThreshold {
		d.metrics.TxValueAnomalies.Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        "tx_value_anomaly",
			Severity:    SeverityMedium,
			Timestamp:   tx.Timestamp,
			Source:      tx.Sender,
			Description: "Transaction value exceeds threshold",
			Metadata: map[string]interface{}{
				"sender":    tx.Sender,
				"amount":    tx.Amount,
				"denom":     tx.Denom,
				"threshold": d.config.MaxTxValueThreshold,
			},
		})
	}

	// Statistical anomaly detection (simple Z-score)
	history := d.accountTxHistory[tx.Sender]
	if len(history) >= 10 {
		amounts := make([]float64, len(history))
		for i, rec := range history {
			amounts[i] = rec.amount
		}

		mean, stdDev := calculateStats(amounts)
		if stdDev > 0 {
			zScore := (tx.Amount - mean) / stdDev
			if math.Abs(zScore) > 3.0 {
				d.metrics.TxValueAnomalies.Inc()
				d.metrics.TxAnomaliesDetected.WithLabelValues("statistical", "medium").Inc()
				d.emitEvent(&SecurityEvent{
					ID:          generateEventID(),
					Type:        "tx_statistical_anomaly",
					Severity:    SeverityMedium,
					Timestamp:   tx.Timestamp,
					Source:      tx.Sender,
					Description: "Statistically anomalous transaction value",
					Metadata: map[string]interface{}{
						"sender":  tx.Sender,
						"amount":  tx.Amount,
						"mean":    mean,
						"std_dev": stdDev,
						"z_score": zScore,
					},
				})
			}
		}
	}
}

// checkNewAccountAbuse checks for suspicious activity from new accounts
func (d *TransactionDetector) checkNewAccountAbuse(tx *TransactionData) {
	cooldownDuration := time.Duration(d.config.NewAccountCooldownMins) * time.Minute

	if tx.AccountAge < cooldownDuration {
		history := d.accountTxHistory[tx.Sender]
		hourAgo := tx.Timestamp.Add(-1 * time.Hour)
		var txInHour int
		for _, rec := range history {
			if rec.timestamp.After(hourAgo) {
				txInHour++
			}
		}

		// New accounts get lower thresholds
		newAccountThreshold := d.config.MaxTxPerHourPerAccount / 4
		if txInHour > newAccountThreshold {
			d.metrics.TxSuspiciousPatterns.WithLabelValues("new_account_abuse").Inc()
			d.emitEvent(&SecurityEvent{
				ID:          generateEventID(),
				Type:        "tx_new_account_abuse",
				Severity:    SeverityHigh,
				Timestamp:   tx.Timestamp,
				Source:      tx.Sender,
				Description: "Suspicious activity from new account",
				Metadata: map[string]interface{}{
					"sender":       tx.Sender,
					"account_age":  tx.AccountAge.String(),
					"tx_count":     txInHour,
					"threshold":    newAccountThreshold,
					"cooldown":     cooldownDuration.String(),
				},
			})
		}
	}
}

// checkHighRiskAccount applies additional scrutiny to high-risk accounts
func (d *TransactionDetector) checkHighRiskAccount(tx *TransactionData) {
	if !tx.IsHighRiskAccount {
		return
	}

	history := d.accountTxHistory[tx.Sender]
	hourAgo := tx.Timestamp.Add(-1 * time.Hour)
	var txInHour int
	for _, rec := range history {
		if rec.timestamp.After(hourAgo) {
			txInHour++
		}
	}

	if txInHour > d.config.HighRiskAccountTxLimit {
		d.metrics.TxAnomaliesDetected.WithLabelValues("high_risk_account", "high").Inc()
		d.emitEvent(&SecurityEvent{
			ID:          generateEventID(),
			Type:        "tx_high_risk_account",
			Severity:    SeverityHigh,
			Timestamp:   tx.Timestamp,
			Source:      tx.Sender,
			Description: "High-risk account exceeded transaction limit",
			Metadata: map[string]interface{}{
				"sender":    tx.Sender,
				"tx_count":  txInHour,
				"limit":     d.config.HighRiskAccountTxLimit,
			},
		})
	}
}

// emitEvent sends an event to the security monitor
func (d *TransactionDetector) emitEvent(event *SecurityEvent) {
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
func (d *TransactionDetector) cleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.mu.Lock()
			cutoff := time.Now().Add(-2 * time.Hour)

			// Clean up account history
			for account, history := range d.accountTxHistory {
				newHistory := make([]txRecord, 0)
				for _, rec := range history {
					if rec.timestamp.After(cutoff) {
						newHistory = append(newHistory, rec)
					}
				}
				if len(newHistory) == 0 {
					delete(d.accountTxHistory, account)
				} else {
					d.accountTxHistory[account] = newHistory
				}
			}

			// Clean up tx hashes
			for hash, timestamp := range d.recentTxHashes {
				if timestamp.Before(cutoff) {
					delete(d.recentTxHashes, hash)
				}
			}

			d.mu.Unlock()
		}
	}
}

// Helper function to calculate mean and standard deviation
func calculateStats(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	stdDev = math.Sqrt(sumSquares / float64(len(values)))

	return mean, stdDev
}

