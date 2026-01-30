// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DisputeStatus defines the status of a dispute
type DisputeStatus uint8

const (
	// DisputeStatusOpen is an open dispute
	DisputeStatusOpen DisputeStatus = 0

	// DisputeStatusUnderReview is under review
	DisputeStatusUnderReview DisputeStatus = 1

	// DisputeStatusResolved is resolved
	DisputeStatusResolved DisputeStatus = 2

	// DisputeStatusEscalated is escalated
	DisputeStatusEscalated DisputeStatus = 3

	// DisputeStatusClosed is closed
	DisputeStatusClosed DisputeStatus = 4

	// DisputeStatusExpired is expired without resolution
	DisputeStatusExpired DisputeStatus = 5
)

// DisputeStatusNames maps status to names
var DisputeStatusNames = map[DisputeStatus]string{
	DisputeStatusOpen:        "open",
	DisputeStatusUnderReview: "under_review",
	DisputeStatusResolved:    "resolved",
	DisputeStatusEscalated:   "escalated",
	DisputeStatusClosed:      "closed",
	DisputeStatusExpired:     "expired",
}

// String returns string representation
func (s DisputeStatus) String() string {
	if name, ok := DisputeStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// DisputeResolutionType defines how a dispute was resolved
type DisputeResolutionType uint8

const (
	// DisputeResolutionNone means no resolution yet
	DisputeResolutionNone DisputeResolutionType = 0

	// DisputeResolutionProviderWin means provider wins
	DisputeResolutionProviderWin DisputeResolutionType = 1

	// DisputeResolutionCustomerWin means customer wins
	DisputeResolutionCustomerWin DisputeResolutionType = 2

	// DisputeResolutionPartialRefund means partial refund
	DisputeResolutionPartialRefund DisputeResolutionType = 3

	// DisputeResolutionMutualAgreement means mutual agreement
	DisputeResolutionMutualAgreement DisputeResolutionType = 4

	// DisputeResolutionArbitration means third-party arbitration
	DisputeResolutionArbitration DisputeResolutionType = 5
)

// DisputeResolutionTypeNames maps resolution types to names
var DisputeResolutionTypeNames = map[DisputeResolutionType]string{
	DisputeResolutionNone:            "none",
	DisputeResolutionProviderWin:     "provider_win",
	DisputeResolutionCustomerWin:     "customer_win",
	DisputeResolutionPartialRefund:   "partial_refund",
	DisputeResolutionMutualAgreement: "mutual_agreement",
	DisputeResolutionArbitration:     "arbitration",
}

// String returns string representation
func (t DisputeResolutionType) String() string {
	if name, ok := DisputeResolutionTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// DisputeWindow defines a dispute window configuration
type DisputeWindow struct {
	// WindowID is the unique identifier
	WindowID string `json:"window_id"`

	// InvoiceID is the invoice being disputed
	InvoiceID string `json:"invoice_id"`

	// WindowStartTime is when the dispute window opens
	WindowStartTime time.Time `json:"window_start_time"`

	// WindowEndTime is when the dispute window closes
	WindowEndTime time.Time `json:"window_end_time"`

	// WindowDurationSeconds is the window duration
	WindowDurationSeconds int64 `json:"window_duration_seconds"`

	// Status is the dispute status
	Status DisputeStatus `json:"status"`

	// DisputeReason is the reason for dispute
	DisputeReason string `json:"dispute_reason,omitempty"`

	// DisputedBy is who initiated the dispute
	DisputedBy string `json:"disputed_by,omitempty"`

	// DisputedAt is when the dispute was initiated
	DisputedAt *time.Time `json:"disputed_at,omitempty"`

	// Resolution is the resolution type
	Resolution DisputeResolutionType `json:"resolution"`

	// ResolutionDetails describes the resolution
	ResolutionDetails string `json:"resolution_details,omitempty"`

	// ResolvedBy is who resolved the dispute
	ResolvedBy string `json:"resolved_by,omitempty"`

	// ResolvedAt is when the dispute was resolved
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// RefundAmount is any refund amount
	RefundAmount sdk.Coins `json:"refund_amount,omitempty"`

	// EscalationPath is the escalation path taken
	EscalationPath []EscalationStep `json:"escalation_path,omitempty"`
}

// EscalationStep represents a step in the escalation path
type EscalationStep struct {
	// StepNumber is the step sequence
	StepNumber uint32 `json:"step_number"`

	// Description describes the escalation step
	Description string `json:"description"`

	// EscalatedTo is who was escalated to
	EscalatedTo string `json:"escalated_to"`

	// EscalatedAt is when escalation occurred
	EscalatedAt time.Time `json:"escalated_at"`

	// ResponseDeadline is deadline for response
	ResponseDeadline time.Time `json:"response_deadline"`

	// Response is the response received
	Response string `json:"response,omitempty"`

	// RespondedAt is when response was received
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// NewDisputeWindow creates a new dispute window
func NewDisputeWindow(
	windowID string,
	invoiceID string,
	startTime time.Time,
	durationSeconds int64,
) *DisputeWindow {
	return &DisputeWindow{
		WindowID:              windowID,
		InvoiceID:             invoiceID,
		WindowStartTime:       startTime,
		WindowEndTime:         startTime.Add(time.Duration(durationSeconds) * time.Second),
		WindowDurationSeconds: durationSeconds,
		Status:                DisputeStatusOpen,
		Resolution:            DisputeResolutionNone,
		EscalationPath:        make([]EscalationStep, 0),
	}
}

// IsOpen checks if the dispute window is currently open
func (w *DisputeWindow) IsOpen(now time.Time) bool {
	return now.After(w.WindowStartTime) && now.Before(w.WindowEndTime)
}

// TimeRemaining returns time remaining in the window
func (w *DisputeWindow) TimeRemaining(now time.Time) time.Duration {
	if now.After(w.WindowEndTime) {
		return 0
	}
	return w.WindowEndTime.Sub(now)
}

// InitiateDispute initiates a dispute
func (w *DisputeWindow) InitiateDispute(disputedBy string, reason string, now time.Time) error {
	if !w.IsOpen(now) {
		return fmt.Errorf("dispute window is not open")
	}

	if w.Status != DisputeStatusOpen {
		return fmt.Errorf("dispute already initiated, status: %s", w.Status)
	}

	w.Status = DisputeStatusUnderReview
	w.DisputedBy = disputedBy
	w.DisputeReason = reason
	w.DisputedAt = &now
	return nil
}

// Resolve resolves the dispute
func (w *DisputeWindow) Resolve(
	resolution DisputeResolutionType,
	details string,
	resolvedBy string,
	refundAmount sdk.Coins,
	now time.Time,
) error {
	if w.Status == DisputeStatusResolved || w.Status == DisputeStatusClosed {
		return fmt.Errorf("dispute already resolved")
	}

	w.Status = DisputeStatusResolved
	w.Resolution = resolution
	w.ResolutionDetails = details
	w.ResolvedBy = resolvedBy
	w.ResolvedAt = &now
	w.RefundAmount = refundAmount
	return nil
}

// Escalate escalates the dispute
func (w *DisputeWindow) Escalate(
	escalateTo string,
	description string,
	responseDeadline time.Time,
	now time.Time,
) error {
	if w.Status == DisputeStatusResolved || w.Status == DisputeStatusClosed {
		return fmt.Errorf("cannot escalate resolved dispute")
	}

	stepNumber := uint32(len(w.EscalationPath) + 1)
	w.EscalationPath = append(w.EscalationPath, EscalationStep{
		StepNumber:       stepNumber,
		Description:      description,
		EscalatedTo:      escalateTo,
		EscalatedAt:      now,
		ResponseDeadline: responseDeadline,
	})

	w.Status = DisputeStatusEscalated
	return nil
}

// Validate validates the dispute window
func (w *DisputeWindow) Validate() error {
	if w.WindowID == "" {
		return fmt.Errorf("window_id is required")
	}

	if w.InvoiceID == "" {
		return fmt.Errorf("invoice_id is required")
	}

	if w.WindowEndTime.Before(w.WindowStartTime) {
		return fmt.Errorf("window_end_time must be after window_start_time")
	}

	if w.WindowDurationSeconds <= 0 {
		return fmt.Errorf("window_duration_seconds must be positive")
	}

	return nil
}

// SettlementHookType defines types of settlement hooks
type SettlementHookType uint8

const (
	// SettlementHookPreValidation runs before validation
	SettlementHookPreValidation SettlementHookType = 0

	// SettlementHookPostValidation runs after validation
	SettlementHookPostValidation SettlementHookType = 1

	// SettlementHookPreSettlement runs before settlement
	SettlementHookPreSettlement SettlementHookType = 2

	// SettlementHookPostSettlement runs after settlement
	SettlementHookPostSettlement SettlementHookType = 3

	// SettlementHookOnDispute runs when dispute is initiated
	SettlementHookOnDispute SettlementHookType = 4

	// SettlementHookOnResolution runs when dispute is resolved
	SettlementHookOnResolution SettlementHookType = 5
)

// SettlementHookTypeNames maps types to names
var SettlementHookTypeNames = map[SettlementHookType]string{
	SettlementHookPreValidation:  "pre_validation",
	SettlementHookPostValidation: "post_validation",
	SettlementHookPreSettlement:  "pre_settlement",
	SettlementHookPostSettlement: "post_settlement",
	SettlementHookOnDispute:      "on_dispute",
	SettlementHookOnResolution:   "on_resolution",
}

// String returns string representation
func (t SettlementHookType) String() string {
	if name, ok := SettlementHookTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// SettlementHookConfig defines a settlement hook configuration
type SettlementHookConfig struct {
	// HookID is the unique identifier
	HookID string `json:"hook_id"`

	// HookType is the type of hook
	HookType SettlementHookType `json:"hook_type"`

	// Name is the hook name
	Name string `json:"name"`

	// Description describes the hook
	Description string `json:"description"`

	// Priority determines execution order (lower = first)
	Priority uint32 `json:"priority"`

	// IsEnabled indicates if hook is enabled
	IsEnabled bool `json:"is_enabled"`

	// Module is the module that handles this hook
	Module string `json:"module"`

	// Action is the action to perform
	Action string `json:"action"`

	// Parameters are hook-specific parameters
	Parameters map[string]string `json:"parameters,omitempty"`

	// FailurePolicy defines what to do on failure
	FailurePolicy HookFailurePolicy `json:"failure_policy"`
}

// HookFailurePolicy defines how to handle hook failures
type HookFailurePolicy uint8

const (
	// HookFailurePolicyAbort aborts the settlement on failure
	HookFailurePolicyAbort HookFailurePolicy = 0

	// HookFailurePolicyContinue continues despite failure
	HookFailurePolicyContinue HookFailurePolicy = 1

	// HookFailurePolicyRetry retries the hook
	HookFailurePolicyRetry HookFailurePolicy = 2

	// HookFailurePolicyEscalate escalates to dispute
	HookFailurePolicyEscalate HookFailurePolicy = 3
)

// HookFailurePolicyNames maps policies to names
var HookFailurePolicyNames = map[HookFailurePolicy]string{
	HookFailurePolicyAbort:    "abort",
	HookFailurePolicyContinue: "continue",
	HookFailurePolicyRetry:    "retry",
	HookFailurePolicyEscalate: "escalate",
}

// String returns string representation
func (p HookFailurePolicy) String() string {
	if name, ok := HookFailurePolicyNames[p]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", p)
}

// Validate validates the hook configuration
func (h *SettlementHookConfig) Validate() error {
	if h.HookID == "" {
		return fmt.Errorf("hook_id is required")
	}

	if h.Name == "" {
		return fmt.Errorf("name is required")
	}

	if h.Module == "" {
		return fmt.Errorf("module is required")
	}

	if h.Action == "" {
		return fmt.Errorf("action is required")
	}

	return nil
}

// SettlementHookResult records a hook execution result
type SettlementHookResult struct {
	// HookID is the hook that was executed
	HookID string `json:"hook_id"`

	// HookType is the type of hook
	HookType SettlementHookType `json:"hook_type"`

	// Success indicates if the hook succeeded
	Success bool `json:"success"`

	// Error is the error message if failed
	Error string `json:"error,omitempty"`

	// ExecutionTimeMs is execution time in milliseconds
	ExecutionTimeMs int64 `json:"execution_time_ms"`

	// ExecutedAt is when the hook was executed
	ExecutedAt time.Time `json:"executed_at"`

	// Output is any output from the hook
	Output map[string]string `json:"output,omitempty"`
}

// SettlementConfig defines settlement configuration
type SettlementConfig struct {
	// DefaultDisputeWindowSeconds is the default dispute window
	DefaultDisputeWindowSeconds int64 `json:"default_dispute_window_seconds"`

	// MinDisputeWindowSeconds is the minimum dispute window
	MinDisputeWindowSeconds int64 `json:"min_dispute_window_seconds"`

	// MaxDisputeWindowSeconds is the maximum dispute window
	MaxDisputeWindowSeconds int64 `json:"max_dispute_window_seconds"`

	// EscalationTimeoutSeconds is timeout for each escalation step
	EscalationTimeoutSeconds int64 `json:"escalation_timeout_seconds"`

	// MaxEscalationSteps is maximum escalation steps
	MaxEscalationSteps uint32 `json:"max_escalation_steps"`

	// AutoSettleAfterDisputeWindow enables auto-settlement
	AutoSettleAfterDisputeWindow bool `json:"auto_settle_after_dispute_window"`

	// RequireCustomerAcknowledgment requires customer ack for settlement
	RequireCustomerAcknowledgment bool `json:"require_customer_acknowledgment"`

	// Hooks are the settlement hooks
	Hooks []SettlementHookConfig `json:"hooks"`
}

// DefaultSettlementConfig returns default settlement configuration
func DefaultSettlementConfig() SettlementConfig {
	return SettlementConfig{
		DefaultDisputeWindowSeconds:   604800, // 7 days
		MinDisputeWindowSeconds:       86400,  // 1 day
		MaxDisputeWindowSeconds:       2592000, // 30 days
		EscalationTimeoutSeconds:      172800, // 2 days
		MaxEscalationSteps:            3,
		AutoSettleAfterDisputeWindow:  true,
		RequireCustomerAcknowledgment: false,
		Hooks:                         DefaultSettlementHooks(),
	}
}

// DefaultSettlementHooks returns default settlement hooks
func DefaultSettlementHooks() []SettlementHookConfig {
	return []SettlementHookConfig{
		{
			HookID:        "validate-usage-records",
			HookType:      SettlementHookPreValidation,
			Name:          "Validate Usage Records",
			Description:   "Validates that all usage records are properly signed and match escrow",
			Priority:      1,
			IsEnabled:     true,
			Module:        "settlement",
			Action:        "validate_usage",
			FailurePolicy: HookFailurePolicyAbort,
		},
		{
			HookID:        "verify-escrow-balance",
			HookType:      SettlementHookPreSettlement,
			Name:          "Verify Escrow Balance",
			Description:   "Verifies sufficient funds in escrow for settlement",
			Priority:      1,
			IsEnabled:     true,
			Module:        "escrow",
			Action:        "verify_balance",
			FailurePolicy: HookFailurePolicyAbort,
		},
		{
			HookID:        "emit-settlement-event",
			HookType:      SettlementHookPostSettlement,
			Name:          "Emit Settlement Event",
			Description:   "Emits blockchain event for settlement",
			Priority:      1,
			IsEnabled:     true,
			Module:        "settlement",
			Action:        "emit_event",
			FailurePolicy: HookFailurePolicyContinue,
		},
		{
			HookID:        "notify-dispute",
			HookType:      SettlementHookOnDispute,
			Name:          "Notify Dispute",
			Description:   "Notifies parties of dispute initiation",
			Priority:      1,
			IsEnabled:     true,
			Module:        "notification",
			Action:        "notify_dispute",
			FailurePolicy: HookFailurePolicyContinue,
		},
	}
}

// Validate validates the settlement configuration
func (c *SettlementConfig) Validate() error {
	if c.DefaultDisputeWindowSeconds <= 0 {
		return fmt.Errorf("default_dispute_window_seconds must be positive")
	}

	if c.MinDisputeWindowSeconds <= 0 {
		return fmt.Errorf("min_dispute_window_seconds must be positive")
	}

	if c.MaxDisputeWindowSeconds < c.MinDisputeWindowSeconds {
		return fmt.Errorf("max_dispute_window_seconds must be >= min_dispute_window_seconds")
	}

	if c.DefaultDisputeWindowSeconds < c.MinDisputeWindowSeconds ||
		c.DefaultDisputeWindowSeconds > c.MaxDisputeWindowSeconds {
		return fmt.Errorf("default_dispute_window_seconds must be within min/max bounds")
	}

	for i, hook := range c.Hooks {
		if err := hook.Validate(); err != nil {
			return fmt.Errorf("hooks[%d]: %w", i, err)
		}
	}

	return nil
}

// GetHooksForType returns hooks of a specific type, sorted by priority
func (c *SettlementConfig) GetHooksForType(hookType SettlementHookType) []SettlementHookConfig {
	var hooks []SettlementHookConfig
	for _, hook := range c.Hooks {
		if hook.HookType == hookType && hook.IsEnabled {
			hooks = append(hooks, hook)
		}
	}
	// Sort by priority (simple bubble sort for small lists)
	for i := 0; i < len(hooks)-1; i++ {
		for j := i + 1; j < len(hooks); j++ {
			if hooks[j].Priority < hooks[i].Priority {
				hooks[i], hooks[j] = hooks[j], hooks[i]
			}
		}
	}
	return hooks
}
