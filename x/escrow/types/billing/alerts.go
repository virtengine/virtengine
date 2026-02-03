// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
)

// AlertSeverity defines the severity level of an alert
type AlertSeverity uint8

const (
	// AlertSeverityInfo is an informational alert
	AlertSeverityInfo AlertSeverity = 0

	// AlertSeverityWarning is a warning alert
	AlertSeverityWarning AlertSeverity = 1

	// AlertSeverityCritical is a critical alert
	AlertSeverityCritical AlertSeverity = 2

	// AlertSeverityEmergency is an emergency alert requiring immediate action
	AlertSeverityEmergency AlertSeverity = 3
)

// AlertSeverityNames maps severity to names
var AlertSeverityNames = map[AlertSeverity]string{
	AlertSeverityInfo:      "info",
	AlertSeverityWarning:   "warning",
	AlertSeverityCritical:  "critical",
	AlertSeverityEmergency: "emergency",
}

// String returns string representation
func (s AlertSeverity) String() string {
	if name, ok := AlertSeverityNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid checks if the severity is valid
func (s AlertSeverity) IsValid() bool {
	_, ok := AlertSeverityNames[s]
	return ok
}

// AlertType defines the type of billing alert
type AlertType uint8

const (
	// AlertTypeVarianceThresholdExceeded is when billing variance exceeds threshold
	AlertTypeVarianceThresholdExceeded AlertType = 0

	// AlertTypeReconciliationFailed is when reconciliation fails
	AlertTypeReconciliationFailed AlertType = 1

	// AlertTypeDisputeVolumeHigh is when dispute volume is unusually high
	AlertTypeDisputeVolumeHigh AlertType = 2

	// AlertTypeOverdueInvoiceThreshold is when overdue invoices exceed threshold
	AlertTypeOverdueInvoiceThreshold AlertType = 3

	// AlertTypePayoutDelayDetected is when payout delays are detected
	AlertTypePayoutDelayDetected AlertType = 4

	// AlertTypeUnreconciledAmountHigh is when unreconciled amounts are high
	AlertTypeUnreconciledAmountHigh AlertType = 5

	// AlertTypeSystemAnomalyDetected is when system anomalies are detected
	AlertTypeSystemAnomalyDetected AlertType = 6
)

// AlertTypeNames maps alert types to names
var AlertTypeNames = map[AlertType]string{
	AlertTypeVarianceThresholdExceeded: "variance_threshold_exceeded",
	AlertTypeReconciliationFailed:      "reconciliation_failed",
	AlertTypeDisputeVolumeHigh:         "dispute_volume_high",
	AlertTypeOverdueInvoiceThreshold:   "overdue_invoice_threshold",
	AlertTypePayoutDelayDetected:       "payout_delay_detected",
	AlertTypeUnreconciledAmountHigh:    "unreconciled_amount_high",
	AlertTypeSystemAnomalyDetected:     "system_anomaly_detected",
}

// String returns string representation
func (t AlertType) String() string {
	if name, ok := AlertTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// IsValid checks if the alert type is valid
func (t AlertType) IsValid() bool {
	_, ok := AlertTypeNames[t]
	return ok
}

// AlertStatus defines the status of an alert
type AlertStatus uint8

const (
	// AlertStatusActive is an active alert
	AlertStatusActive AlertStatus = 0

	// AlertStatusAcknowledged is an acknowledged alert
	AlertStatusAcknowledged AlertStatus = 1

	// AlertStatusResolved is a resolved alert
	AlertStatusResolved AlertStatus = 2

	// AlertStatusSuppressed is a suppressed alert
	AlertStatusSuppressed AlertStatus = 3
)

// AlertStatusNames maps status to names
var AlertStatusNames = map[AlertStatus]string{
	AlertStatusActive:       "active",
	AlertStatusAcknowledged: "acknowledged",
	AlertStatusResolved:     "resolved",
	AlertStatusSuppressed:   "suppressed",
}

// String returns string representation
func (s AlertStatus) String() string {
	if name, ok := AlertStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsValid checks if the status is valid
func (s AlertStatus) IsValid() bool {
	_, ok := AlertStatusNames[s]
	return ok
}

// IsTerminal returns true if the status is final
func (s AlertStatus) IsTerminal() bool {
	return s == AlertStatusResolved || s == AlertStatusSuppressed
}

// ThresholdUnit defines the unit for threshold comparison
type ThresholdUnit uint8

const (
	// ThresholdUnitPercentage is a percentage-based threshold
	ThresholdUnitPercentage ThresholdUnit = 0

	// ThresholdUnitAbsoluteAmount is an absolute amount threshold
	ThresholdUnitAbsoluteAmount ThresholdUnit = 1

	// ThresholdUnitCount is a count-based threshold
	ThresholdUnitCount ThresholdUnit = 2
)

// ThresholdUnitNames maps units to names
var ThresholdUnitNames = map[ThresholdUnit]string{
	ThresholdUnitPercentage:     "percentage",
	ThresholdUnitAbsoluteAmount: "absolute_amount",
	ThresholdUnitCount:          "count",
}

// String returns string representation
func (u ThresholdUnit) String() string {
	if name, ok := ThresholdUnitNames[u]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", u)
}

// IsValid checks if the unit is valid
func (u ThresholdUnit) IsValid() bool {
	_, ok := ThresholdUnitNames[u]
	return ok
}

// ComparisonOperator defines comparison operators for thresholds
type ComparisonOperator uint8

const (
	// ComparisonOperatorGT is greater than
	ComparisonOperatorGT ComparisonOperator = 0

	// ComparisonOperatorGTE is greater than or equal
	ComparisonOperatorGTE ComparisonOperator = 1

	// ComparisonOperatorLT is less than
	ComparisonOperatorLT ComparisonOperator = 2

	// ComparisonOperatorLTE is less than or equal
	ComparisonOperatorLTE ComparisonOperator = 3

	// ComparisonOperatorEQ is equal
	ComparisonOperatorEQ ComparisonOperator = 4
)

// ComparisonOperatorNames maps operators to names
var ComparisonOperatorNames = map[ComparisonOperator]string{
	ComparisonOperatorGT:  "gt",
	ComparisonOperatorGTE: "gte",
	ComparisonOperatorLT:  "lt",
	ComparisonOperatorLTE: "lte",
	ComparisonOperatorEQ:  "eq",
}

// String returns string representation
func (o ComparisonOperator) String() string {
	if name, ok := ComparisonOperatorNames[o]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", o)
}

// IsValid checks if the operator is valid
func (o ComparisonOperator) IsValid() bool {
	_, ok := ComparisonOperatorNames[o]
	return ok
}

// Evaluate evaluates the comparison between actual and threshold values
func (o ComparisonOperator) Evaluate(actual, threshold sdkmath.LegacyDec) bool {
	switch o {
	case ComparisonOperatorGT:
		return actual.GT(threshold)
	case ComparisonOperatorGTE:
		return actual.GTE(threshold)
	case ComparisonOperatorLT:
		return actual.LT(threshold)
	case ComparisonOperatorLTE:
		return actual.LTE(threshold)
	case ComparisonOperatorEQ:
		return actual.Equal(threshold)
	default:
		return false
	}
}

// AlertThreshold defines a threshold for triggering alerts
type AlertThreshold struct {
	// ThresholdID is the unique identifier
	ThresholdID string `json:"threshold_id"`

	// Name is the human-readable name
	Name string `json:"name"`

	// Description describes the threshold
	Description string `json:"description"`

	// AlertType is the type of alert this threshold triggers
	AlertType AlertType `json:"alert_type"`

	// Severity is the severity level when triggered
	Severity AlertSeverity `json:"severity"`

	// ThresholdValue is the threshold value
	ThresholdValue sdkmath.LegacyDec `json:"threshold_value"`

	// ThresholdUnit is the unit of the threshold
	ThresholdUnit ThresholdUnit `json:"threshold_unit"`

	// ComparisonOperator is the comparison operator
	ComparisonOperator ComparisonOperator `json:"comparison_operator"`

	// EvaluationWindowSeconds is the time window for evaluation
	EvaluationWindowSeconds int64 `json:"evaluation_window_seconds"`

	// CooldownSeconds is the cooldown period between alerts
	CooldownSeconds int64 `json:"cooldown_seconds"`

	// IsEnabled indicates if the threshold is enabled
	IsEnabled bool `json:"is_enabled"`
}

// NewAlertThreshold creates a new alert threshold
func NewAlertThreshold(
	thresholdID string,
	name string,
	alertType AlertType,
	severity AlertSeverity,
	thresholdValue sdkmath.LegacyDec,
	thresholdUnit ThresholdUnit,
	operator ComparisonOperator,
) *AlertThreshold {
	return &AlertThreshold{
		ThresholdID:             thresholdID,
		Name:                    name,
		AlertType:               alertType,
		Severity:                severity,
		ThresholdValue:          thresholdValue,
		ThresholdUnit:           thresholdUnit,
		ComparisonOperator:      operator,
		EvaluationWindowSeconds: 3600, // 1 hour default
		CooldownSeconds:         1800, // 30 minutes default
		IsEnabled:               true,
	}
}

// Evaluate checks if the actual value triggers this threshold
func (t *AlertThreshold) Evaluate(actualValue sdkmath.LegacyDec) bool {
	if !t.IsEnabled {
		return false
	}
	return t.ComparisonOperator.Evaluate(actualValue, t.ThresholdValue)
}

// Validate validates the alert threshold
func (t *AlertThreshold) Validate() error {
	if t.ThresholdID == "" {
		return fmt.Errorf("threshold_id is required")
	}

	if len(t.ThresholdID) > 64 {
		return fmt.Errorf("threshold_id exceeds maximum length of 64")
	}

	if t.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(t.Name) > 128 {
		return fmt.Errorf("name exceeds maximum length of 128")
	}

	if !t.AlertType.IsValid() {
		return fmt.Errorf("invalid alert_type: %d", t.AlertType)
	}

	if !t.Severity.IsValid() {
		return fmt.Errorf("invalid severity: %d", t.Severity)
	}

	if !t.ThresholdUnit.IsValid() {
		return fmt.Errorf("invalid threshold_unit: %d", t.ThresholdUnit)
	}

	if !t.ComparisonOperator.IsValid() {
		return fmt.Errorf("invalid comparison_operator: %d", t.ComparisonOperator)
	}

	if t.ThresholdValue.IsNegative() {
		return fmt.Errorf("threshold_value cannot be negative")
	}

	if t.EvaluationWindowSeconds <= 0 {
		return fmt.Errorf("evaluation_window_seconds must be positive")
	}

	if t.CooldownSeconds < 0 {
		return fmt.Errorf("cooldown_seconds cannot be negative")
	}

	return nil
}

// AlertRule defines a rule for generating alerts
type AlertRule struct {
	// RuleID is the unique identifier
	RuleID string `json:"rule_id"`

	// Name is the human-readable name
	Name string `json:"name"`

	// Description describes the rule
	Description string `json:"description"`

	// Thresholds are the thresholds for this rule
	Thresholds []AlertThreshold `json:"thresholds"`

	// NotificationChannels are the channels to notify
	NotificationChannels []string `json:"notification_channels"`

	// EscalationPolicy is the escalation policy name
	EscalationPolicy string `json:"escalation_policy"`

	// CreatedAt is when the rule was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the rule was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// NewAlertRule creates a new alert rule
func NewAlertRule(
	ruleID string,
	name string,
	description string,
	now time.Time,
) *AlertRule {
	return &AlertRule{
		RuleID:               ruleID,
		Name:                 name,
		Description:          description,
		Thresholds:           make([]AlertThreshold, 0),
		NotificationChannels: make([]string, 0),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// AddThreshold adds a threshold to the rule
func (r *AlertRule) AddThreshold(threshold AlertThreshold) {
	r.Thresholds = append(r.Thresholds, threshold)
}

// AddNotificationChannel adds a notification channel
func (r *AlertRule) AddNotificationChannel(channel string) {
	r.NotificationChannels = append(r.NotificationChannels, channel)
}

// Validate validates the alert rule
func (r *AlertRule) Validate() error {
	if r.RuleID == "" {
		return fmt.Errorf("rule_id is required")
	}

	if len(r.RuleID) > 64 {
		return fmt.Errorf("rule_id exceeds maximum length of 64")
	}

	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(r.Name) > 128 {
		return fmt.Errorf("name exceeds maximum length of 128")
	}

	if len(r.Thresholds) == 0 {
		return fmt.Errorf("at least one threshold is required")
	}

	for i, threshold := range r.Thresholds {
		if err := threshold.Validate(); err != nil {
			return fmt.Errorf("thresholds[%d]: %w", i, err)
		}
	}

	if r.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}

	return nil
}

// GetEnabledThresholds returns only enabled thresholds
func (r *AlertRule) GetEnabledThresholds() []AlertThreshold {
	var enabled []AlertThreshold
	for _, t := range r.Thresholds {
		if t.IsEnabled {
			enabled = append(enabled, t)
		}
	}
	return enabled
}

// Alert represents a triggered alert instance
type Alert struct {
	// AlertID is the unique identifier
	AlertID string `json:"alert_id"`

	// RuleID is the rule that triggered this alert
	RuleID string `json:"rule_id"`

	// ThresholdID is the threshold that was exceeded
	ThresholdID string `json:"threshold_id"`

	// AlertType is the type of alert
	AlertType AlertType `json:"alert_type"`

	// Severity is the severity level
	Severity AlertSeverity `json:"severity"`

	// Status is the current status
	Status AlertStatus `json:"status"`

	// TriggerValue is the value that triggered the alert
	TriggerValue sdkmath.LegacyDec `json:"trigger_value"`

	// ThresholdValue is the threshold that was exceeded
	ThresholdValue sdkmath.LegacyDec `json:"threshold_value"`

	// Message is the alert message
	Message string `json:"message"`

	// EntityType is the type of entity (e.g., "invoice", "escrow", "provider")
	EntityType string `json:"entity_type"`

	// EntityID is the ID of the related entity
	EntityID string `json:"entity_id"`

	// TriggeredAt is when the alert was triggered
	TriggeredAt time.Time `json:"triggered_at"`

	// AcknowledgedAt is when the alert was acknowledged
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`

	// ResolvedAt is when the alert was resolved
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// AcknowledgedBy is who acknowledged the alert
	AcknowledgedBy string `json:"acknowledged_by,omitempty"`

	// ResolvedBy is who resolved the alert
	ResolvedBy string `json:"resolved_by,omitempty"`

	// Notes contains additional notes
	Notes string `json:"notes,omitempty"`
}

// NewAlert creates a new alert
func NewAlert(
	alertID string,
	ruleID string,
	thresholdID string,
	alertType AlertType,
	severity AlertSeverity,
	triggerValue sdkmath.LegacyDec,
	thresholdValue sdkmath.LegacyDec,
	message string,
	entityType string,
	entityID string,
	now time.Time,
) *Alert {
	return &Alert{
		AlertID:        alertID,
		RuleID:         ruleID,
		ThresholdID:    thresholdID,
		AlertType:      alertType,
		Severity:       severity,
		Status:         AlertStatusActive,
		TriggerValue:   triggerValue,
		ThresholdValue: thresholdValue,
		Message:        message,
		EntityType:     entityType,
		EntityID:       entityID,
		TriggeredAt:    now,
	}
}

// Acknowledge acknowledges the alert
func (a *Alert) Acknowledge(acknowledgedBy string, now time.Time) error {
	if a.Status.IsTerminal() {
		return fmt.Errorf("cannot acknowledge %s alert", a.Status)
	}

	a.Status = AlertStatusAcknowledged
	a.AcknowledgedAt = &now
	a.AcknowledgedBy = acknowledgedBy
	return nil
}

// Resolve resolves the alert
func (a *Alert) Resolve(resolvedBy string, notes string, now time.Time) error {
	if a.Status == AlertStatusResolved {
		return fmt.Errorf("alert already resolved")
	}

	a.Status = AlertStatusResolved
	a.ResolvedAt = &now
	a.ResolvedBy = resolvedBy
	if notes != "" {
		a.Notes = notes
	}
	return nil
}

// Suppress suppresses the alert
func (a *Alert) Suppress(notes string, now time.Time) error {
	if a.Status.IsTerminal() {
		return fmt.Errorf("cannot suppress %s alert", a.Status)
	}

	a.Status = AlertStatusSuppressed
	a.ResolvedAt = &now
	if notes != "" {
		a.Notes = notes
	}
	return nil
}

// Validate validates the alert
func (a *Alert) Validate() error {
	if a.AlertID == "" {
		return fmt.Errorf("alert_id is required")
	}

	if len(a.AlertID) > 64 {
		return fmt.Errorf("alert_id exceeds maximum length of 64")
	}

	if a.RuleID == "" {
		return fmt.Errorf("rule_id is required")
	}

	if a.ThresholdID == "" {
		return fmt.Errorf("threshold_id is required")
	}

	if !a.AlertType.IsValid() {
		return fmt.Errorf("invalid alert_type: %d", a.AlertType)
	}

	if !a.Severity.IsValid() {
		return fmt.Errorf("invalid severity: %d", a.Severity)
	}

	if !a.Status.IsValid() {
		return fmt.Errorf("invalid status: %d", a.Status)
	}

	if a.Message == "" {
		return fmt.Errorf("message is required")
	}

	if a.TriggeredAt.IsZero() {
		return fmt.Errorf("triggered_at is required")
	}

	return nil
}

// AlertConfig defines configuration for the alert system
type AlertConfig struct {
	// IsEnabled enables/disables the alert system
	IsEnabled bool `json:"is_enabled"`

	// DefaultCooldownSeconds is the default cooldown between alerts
	DefaultCooldownSeconds int64 `json:"default_cooldown_seconds"`

	// DefaultEvaluationWindowSeconds is the default evaluation window
	DefaultEvaluationWindowSeconds int64 `json:"default_evaluation_window_seconds"`

	// MaxActiveAlerts is the maximum number of active alerts
	MaxActiveAlerts uint32 `json:"max_active_alerts"`

	// RetentionDays is how long to retain resolved alerts
	RetentionDays uint32 `json:"retention_days"`

	// DefaultThresholds contains the default alert thresholds
	DefaultThresholds []AlertThreshold `json:"default_thresholds"`

	// NotificationChannels are the default notification channels
	NotificationChannels []string `json:"notification_channels"`
}

// DefaultAlertConfig returns default alert configuration with sensible production defaults
func DefaultAlertConfig() AlertConfig {
	return AlertConfig{
		IsEnabled:                      true,
		DefaultCooldownSeconds:         1800, // 30 minutes
		DefaultEvaluationWindowSeconds: 3600, // 1 hour
		MaxActiveAlerts:                1000,
		RetentionDays:                  90,
		DefaultThresholds:              DefaultAlertThresholds(),
		NotificationChannels:           []string{"webhook", "event"},
	}
}

// DefaultAlertThresholds returns default alert thresholds for production
func DefaultAlertThresholds() []AlertThreshold {
	return []AlertThreshold{
		{
			ThresholdID:             "variance-warning",
			Name:                    "Billing Variance Warning",
			Description:             "Warning when billing variance exceeds 5%",
			AlertType:               AlertTypeVarianceThresholdExceeded,
			Severity:                AlertSeverityWarning,
			ThresholdValue:          sdkmath.LegacyNewDecWithPrec(5, 2), // 0.05 = 5%
			ThresholdUnit:           ThresholdUnitPercentage,
			ComparisonOperator:      ComparisonOperatorGT,
			EvaluationWindowSeconds: 3600, // 1 hour
			CooldownSeconds:         1800, // 30 minutes
			IsEnabled:               true,
		},
		{
			ThresholdID:             "variance-critical",
			Name:                    "Billing Variance Critical",
			Description:             "Critical alert when billing variance exceeds 10%",
			AlertType:               AlertTypeVarianceThresholdExceeded,
			Severity:                AlertSeverityCritical,
			ThresholdValue:          sdkmath.LegacyNewDecWithPrec(10, 2), // 0.10 = 10%
			ThresholdUnit:           ThresholdUnitPercentage,
			ComparisonOperator:      ComparisonOperatorGT,
			EvaluationWindowSeconds: 3600, // 1 hour
			CooldownSeconds:         900,  // 15 minutes
			IsEnabled:               true,
		},
		{
			ThresholdID:             "reconciliation-failed",
			Name:                    "Reconciliation Failed",
			Description:             "Alert when reconciliation fails",
			AlertType:               AlertTypeReconciliationFailed,
			Severity:                AlertSeverityCritical,
			ThresholdValue:          sdkmath.LegacyOneDec(),
			ThresholdUnit:           ThresholdUnitCount,
			ComparisonOperator:      ComparisonOperatorGTE,
			EvaluationWindowSeconds: 3600, // 1 hour
			CooldownSeconds:         300,  // 5 minutes
			IsEnabled:               true,
		},
		{
			ThresholdID:             "dispute-volume-high",
			Name:                    "High Dispute Volume",
			Description:             "Warning when dispute rate exceeds 2%",
			AlertType:               AlertTypeDisputeVolumeHigh,
			Severity:                AlertSeverityWarning,
			ThresholdValue:          sdkmath.LegacyNewDecWithPrec(2, 2), // 0.02 = 2%
			ThresholdUnit:           ThresholdUnitPercentage,
			ComparisonOperator:      ComparisonOperatorGT,
			EvaluationWindowSeconds: 86400, // 24 hours
			CooldownSeconds:         3600,  // 1 hour
			IsEnabled:               true,
		},
		{
			ThresholdID:             "overdue-invoices-warning",
			Name:                    "Overdue Invoices Warning",
			Description:             "Warning when overdue invoices exceed 10",
			AlertType:               AlertTypeOverdueInvoiceThreshold,
			Severity:                AlertSeverityWarning,
			ThresholdValue:          sdkmath.LegacyNewDec(10),
			ThresholdUnit:           ThresholdUnitCount,
			ComparisonOperator:      ComparisonOperatorGT,
			EvaluationWindowSeconds: 86400, // 24 hours
			CooldownSeconds:         3600,  // 1 hour
			IsEnabled:               true,
		},
		{
			ThresholdID:             "overdue-invoices-critical",
			Name:                    "Overdue Invoices Critical",
			Description:             "Critical alert when overdue invoices exceed 50",
			AlertType:               AlertTypeOverdueInvoiceThreshold,
			Severity:                AlertSeverityCritical,
			ThresholdValue:          sdkmath.LegacyNewDec(50),
			ThresholdUnit:           ThresholdUnitCount,
			ComparisonOperator:      ComparisonOperatorGT,
			EvaluationWindowSeconds: 86400, // 24 hours
			CooldownSeconds:         1800,  // 30 minutes
			IsEnabled:               true,
		},
		{
			ThresholdID:             "payout-delay-warning",
			Name:                    "Payout Delay Warning",
			Description:             "Warning when payout is delayed more than 24 hours",
			AlertType:               AlertTypePayoutDelayDetected,
			Severity:                AlertSeverityWarning,
			ThresholdValue:          sdkmath.LegacyNewDec(86400), // 24 hours in seconds
			ThresholdUnit:           ThresholdUnitAbsoluteAmount,
			ComparisonOperator:      ComparisonOperatorGT,
			EvaluationWindowSeconds: 3600, // 1 hour
			CooldownSeconds:         3600, // 1 hour
			IsEnabled:               true,
		},
		{
			ThresholdID:             "payout-delay-critical",
			Name:                    "Payout Delay Critical",
			Description:             "Critical alert when payout is delayed more than 72 hours",
			AlertType:               AlertTypePayoutDelayDetected,
			Severity:                AlertSeverityCritical,
			ThresholdValue:          sdkmath.LegacyNewDec(259200), // 72 hours in seconds
			ThresholdUnit:           ThresholdUnitAbsoluteAmount,
			ComparisonOperator:      ComparisonOperatorGT,
			EvaluationWindowSeconds: 3600, // 1 hour
			CooldownSeconds:         1800, // 30 minutes
			IsEnabled:               true,
		},
		{
			ThresholdID:             "unreconciled-amount-warning",
			Name:                    "Unreconciled Amount Warning",
			Description:             "Warning when unreconciled amount exceeds 5% of total",
			AlertType:               AlertTypeUnreconciledAmountHigh,
			Severity:                AlertSeverityWarning,
			ThresholdValue:          sdkmath.LegacyNewDecWithPrec(5, 2), // 0.05 = 5%
			ThresholdUnit:           ThresholdUnitPercentage,
			ComparisonOperator:      ComparisonOperatorGT,
			EvaluationWindowSeconds: 86400, // 24 hours
			CooldownSeconds:         3600,  // 1 hour
			IsEnabled:               true,
		},
		{
			ThresholdID:             "unreconciled-amount-critical",
			Name:                    "Unreconciled Amount Critical",
			Description:             "Critical alert when unreconciled amount exceeds 10% of total",
			AlertType:               AlertTypeUnreconciledAmountHigh,
			Severity:                AlertSeverityCritical,
			ThresholdValue:          sdkmath.LegacyNewDecWithPrec(10, 2), // 0.10 = 10%
			ThresholdUnit:           ThresholdUnitPercentage,
			ComparisonOperator:      ComparisonOperatorGT,
			EvaluationWindowSeconds: 86400, // 24 hours
			CooldownSeconds:         1800,  // 30 minutes
			IsEnabled:               true,
		},
		{
			ThresholdID:             "system-anomaly-emergency",
			Name:                    "System Anomaly Emergency",
			Description:             "Emergency alert for system anomalies",
			AlertType:               AlertTypeSystemAnomalyDetected,
			Severity:                AlertSeverityEmergency,
			ThresholdValue:          sdkmath.LegacyOneDec(),
			ThresholdUnit:           ThresholdUnitCount,
			ComparisonOperator:      ComparisonOperatorGTE,
			EvaluationWindowSeconds: 300, // 5 minutes
			CooldownSeconds:         60,  // 1 minute
			IsEnabled:               true,
		},
	}
}

// Validate validates the alert configuration
func (c *AlertConfig) Validate() error {
	if c.DefaultCooldownSeconds < 0 {
		return fmt.Errorf("default_cooldown_seconds cannot be negative")
	}

	if c.DefaultEvaluationWindowSeconds <= 0 {
		return fmt.Errorf("default_evaluation_window_seconds must be positive")
	}

	if c.MaxActiveAlerts == 0 {
		return fmt.Errorf("max_active_alerts must be positive")
	}

	if c.RetentionDays == 0 {
		return fmt.Errorf("retention_days must be positive")
	}

	for i, threshold := range c.DefaultThresholds {
		if err := threshold.Validate(); err != nil {
			return fmt.Errorf("default_thresholds[%d]: %w", i, err)
		}
	}

	return nil
}

// GetThresholdByID returns a threshold by ID
func (c *AlertConfig) GetThresholdByID(thresholdID string) (*AlertThreshold, bool) {
	for i := range c.DefaultThresholds {
		if c.DefaultThresholds[i].ThresholdID == thresholdID {
			return &c.DefaultThresholds[i], true
		}
	}
	return nil, false
}

// GetThresholdsByType returns thresholds by alert type
func (c *AlertConfig) GetThresholdsByType(alertType AlertType) []AlertThreshold {
	var thresholds []AlertThreshold
	for _, t := range c.DefaultThresholds {
		if t.AlertType == alertType && t.IsEnabled {
			thresholds = append(thresholds, t)
		}
	}
	return thresholds
}

// GetThresholdsBySeverity returns thresholds by severity
func (c *AlertConfig) GetThresholdsBySeverity(severity AlertSeverity) []AlertThreshold {
	var thresholds []AlertThreshold
	for _, t := range c.DefaultThresholds {
		if t.Severity == severity && t.IsEnabled {
			thresholds = append(thresholds, t)
		}
	}
	return thresholds
}

// Store key prefixes for alert types
var (
	// AlertPrefix is the prefix for alert storage
	AlertPrefix = []byte{0xA0}

	// AlertRulePrefix is the prefix for alert rules
	AlertRulePrefix = []byte{0xA1}

	// AlertThresholdPrefix is the prefix for alert thresholds
	AlertThresholdPrefix = []byte{0xA2}

	// AlertConfigPrefix is the prefix for alert configuration
	AlertConfigPrefix = []byte{0xA3}

	// AlertByStatusPrefix indexes alerts by status
	AlertByStatusPrefix = []byte{0xA4}

	// AlertByTypePrefix indexes alerts by type
	AlertByTypePrefix = []byte{0xA5}

	// AlertBySeverityPrefix indexes alerts by severity
	AlertBySeverityPrefix = []byte{0xA6}

	// AlertByEntityPrefix indexes alerts by entity
	AlertByEntityPrefix = []byte{0xA7}
)

// BuildAlertKey builds the key for an alert
func BuildAlertKey(alertID string) []byte {
	return append(AlertPrefix, []byte(alertID)...)
}

// ParseAlertKey parses an alert key
func ParseAlertKey(key []byte) (string, error) {
	if len(key) <= len(AlertPrefix) {
		return "", fmt.Errorf("invalid alert key length")
	}
	return string(key[len(AlertPrefix):]), nil
}

// BuildAlertRuleKey builds the key for an alert rule
func BuildAlertRuleKey(ruleID string) []byte {
	return append(AlertRulePrefix, []byte(ruleID)...)
}

// ParseAlertRuleKey parses an alert rule key
func ParseAlertRuleKey(key []byte) (string, error) {
	if len(key) <= len(AlertRulePrefix) {
		return "", fmt.Errorf("invalid alert rule key length")
	}
	return string(key[len(AlertRulePrefix):]), nil
}

// BuildAlertThresholdKey builds the key for an alert threshold
func BuildAlertThresholdKey(thresholdID string) []byte {
	return append(AlertThresholdPrefix, []byte(thresholdID)...)
}

// ParseAlertThresholdKey parses an alert threshold key
func ParseAlertThresholdKey(key []byte) (string, error) {
	if len(key) <= len(AlertThresholdPrefix) {
		return "", fmt.Errorf("invalid alert threshold key length")
	}
	return string(key[len(AlertThresholdPrefix):]), nil
}

// BuildAlertByStatusKey builds the index key for alerts by status
func BuildAlertByStatusKey(status AlertStatus, alertID string) []byte {
	key := make([]byte, 0, len(AlertByStatusPrefix)+2+len(alertID))
	key = append(key, AlertByStatusPrefix...)
	key = append(key, byte(status))
	key = append(key, byte('/'))
	return append(key, []byte(alertID)...)
}

// BuildAlertByStatusPrefix builds the prefix for alerts by status
func BuildAlertByStatusPrefix(status AlertStatus) []byte {
	key := make([]byte, 0, len(AlertByStatusPrefix)+2)
	key = append(key, AlertByStatusPrefix...)
	key = append(key, byte(status))
	return append(key, byte('/'))
}

// BuildAlertByTypeKey builds the index key for alerts by type
func BuildAlertByTypeKey(alertType AlertType, alertID string) []byte {
	key := make([]byte, 0, len(AlertByTypePrefix)+2+len(alertID))
	key = append(key, AlertByTypePrefix...)
	key = append(key, byte(alertType))
	key = append(key, byte('/'))
	return append(key, []byte(alertID)...)
}

// BuildAlertByTypePrefix builds the prefix for alerts by type
func BuildAlertByTypePrefix(alertType AlertType) []byte {
	key := make([]byte, 0, len(AlertByTypePrefix)+2)
	key = append(key, AlertByTypePrefix...)
	key = append(key, byte(alertType))
	return append(key, byte('/'))
}

// BuildAlertBySeverityKey builds the index key for alerts by severity
func BuildAlertBySeverityKey(severity AlertSeverity, alertID string) []byte {
	key := make([]byte, 0, len(AlertBySeverityPrefix)+2+len(alertID))
	key = append(key, AlertBySeverityPrefix...)
	key = append(key, byte(severity))
	key = append(key, byte('/'))
	return append(key, []byte(alertID)...)
}

// BuildAlertBySeverityPrefix builds the prefix for alerts by severity
func BuildAlertBySeverityPrefix(severity AlertSeverity) []byte {
	key := make([]byte, 0, len(AlertBySeverityPrefix)+2)
	key = append(key, AlertBySeverityPrefix...)
	key = append(key, byte(severity))
	return append(key, byte('/'))
}

// BuildAlertByEntityKey builds the index key for alerts by entity
func BuildAlertByEntityKey(entityType string, entityID string, alertID string) []byte {
	key := make([]byte, 0, len(AlertByEntityPrefix)+len(entityType)+1+len(entityID)+1+len(alertID))
	key = append(key, AlertByEntityPrefix...)
	key = append(key, []byte(entityType)...)
	key = append(key, byte('/'))
	key = append(key, []byte(entityID)...)
	key = append(key, byte('/'))
	return append(key, []byte(alertID)...)
}

// BuildAlertByEntityPrefix builds the prefix for alerts by entity
func BuildAlertByEntityPrefix(entityType string, entityID string) []byte {
	key := make([]byte, 0, len(AlertByEntityPrefix)+len(entityType)+1+len(entityID)+1)
	key = append(key, AlertByEntityPrefix...)
	key = append(key, []byte(entityType)...)
	key = append(key, byte('/'))
	key = append(key, []byte(entityID)...)
	return append(key, byte('/'))
}
