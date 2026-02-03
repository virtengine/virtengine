// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ReconciliationStatus defines the status of a reconciliation report
type ReconciliationStatus uint8

const (
	// ReconciliationStatusPending is a pending reconciliation
	ReconciliationStatusPending ReconciliationStatus = 0

	// ReconciliationStatusComplete is a completed reconciliation
	ReconciliationStatusComplete ReconciliationStatus = 1

	// ReconciliationStatusFailed is a failed reconciliation
	ReconciliationStatusFailed ReconciliationStatus = 2

	// ReconciliationStatusPartial is a partially complete reconciliation
	ReconciliationStatusPartial ReconciliationStatus = 3
)

// ReconciliationStatusNames maps status to names
var ReconciliationStatusNames = map[ReconciliationStatus]string{
	ReconciliationStatusPending:  "pending",
	ReconciliationStatusComplete: "complete",
	ReconciliationStatusFailed:   "failed",
	ReconciliationStatusPartial:  "partial",
}

// String returns string representation
func (s ReconciliationStatus) String() string {
	if name, ok := ReconciliationStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// ReconciliationReport represents a reconciliation report for invoices
type ReconciliationReport struct {
	// ReportID is the unique identifier
	ReportID string `json:"report_id"`

	// ReportType is the type of reconciliation
	ReportType ReconciliationReportType `json:"report_type"`

	// PeriodStart is the start of the reconciliation period
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the reconciliation period
	PeriodEnd time.Time `json:"period_end"`

	// Provider is the provider address (optional, for provider-specific reports)
	Provider string `json:"provider,omitempty"`

	// Customer is the customer address (optional, for customer-specific reports)
	Customer string `json:"customer,omitempty"`

	// Status is the reconciliation status
	Status ReconciliationStatus `json:"status"`

	// Summary contains the reconciliation summary
	Summary ReconciliationSummary `json:"summary"`

	// Discrepancies lists any discrepancies found
	Discrepancies []ReconciliationDiscrepancy `json:"discrepancies,omitempty"`

	// InvoiceIDs are the invoices included in this report
	InvoiceIDs []string `json:"invoice_ids"`

	// SettlementIDs are the settlements included in this report
	SettlementIDs []string `json:"settlement_ids"`

	// UsageRecordIDs are the usage records included in this report
	UsageRecordIDs []string `json:"usage_record_ids"`

	// GeneratedAt is when the report was generated
	GeneratedAt time.Time `json:"generated_at"`

	// GeneratedBy is who generated the report
	GeneratedBy string `json:"generated_by"`

	// BlockHeight is when the report was generated
	BlockHeight int64 `json:"block_height"`

	// Notes contains any additional notes
	Notes string `json:"notes,omitempty"`
}

// ReconciliationReportType defines types of reconciliation reports
type ReconciliationReportType uint8

const (
	// ReconciliationReportTypeDaily is a daily reconciliation
	ReconciliationReportTypeDaily ReconciliationReportType = 0

	// ReconciliationReportTypeWeekly is a weekly reconciliation
	ReconciliationReportTypeWeekly ReconciliationReportType = 1

	// ReconciliationReportTypeMonthly is a monthly reconciliation
	ReconciliationReportTypeMonthly ReconciliationReportType = 2

	// ReconciliationReportTypeOnDemand is an on-demand reconciliation
	ReconciliationReportTypeOnDemand ReconciliationReportType = 3

	// ReconciliationReportTypeSettlement is a settlement-based reconciliation
	ReconciliationReportTypeSettlement ReconciliationReportType = 4
)

// ReconciliationReportTypeNames maps types to names
var ReconciliationReportTypeNames = map[ReconciliationReportType]string{
	ReconciliationReportTypeDaily:      "daily",
	ReconciliationReportTypeWeekly:     "weekly",
	ReconciliationReportTypeMonthly:    "monthly",
	ReconciliationReportTypeOnDemand:   "on_demand",
	ReconciliationReportTypeSettlement: "settlement",
}

// String returns string representation
func (t ReconciliationReportType) String() string {
	if name, ok := ReconciliationReportTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// ReconciliationSummary contains summary statistics
type ReconciliationSummary struct {
	// TotalInvoices is the total number of invoices
	TotalInvoices uint32 `json:"total_invoices"`

	// TotalInvoiceAmount is the total invoiced amount
	TotalInvoiceAmount sdk.Coins `json:"total_invoice_amount"`

	// TotalSettlements is the total number of settlements
	TotalSettlements uint32 `json:"total_settlements"`

	// TotalSettlementAmount is the total settled amount
	TotalSettlementAmount sdk.Coins `json:"total_settlement_amount"`

	// TotalUsageRecords is the total number of usage records
	TotalUsageRecords uint32 `json:"total_usage_records"`

	// TotalUsageAmount is the total usage amount
	TotalUsageAmount sdk.Coins `json:"total_usage_amount"`

	// PaidInvoices is the number of paid invoices
	PaidInvoices uint32 `json:"paid_invoices"`

	// PaidAmount is the total paid amount
	PaidAmount sdk.Coins `json:"paid_amount"`

	// OutstandingInvoices is the number of outstanding invoices
	OutstandingInvoices uint32 `json:"outstanding_invoices"`

	// OutstandingAmount is the total outstanding amount
	OutstandingAmount sdk.Coins `json:"outstanding_amount"`

	// DisputedInvoices is the number of disputed invoices
	DisputedInvoices uint32 `json:"disputed_invoices"`

	// DisputedAmount is the total disputed amount
	DisputedAmount sdk.Coins `json:"disputed_amount"`

	// OverdueInvoices is the number of overdue invoices
	OverdueInvoices uint32 `json:"overdue_invoices"`

	// OverdueAmount is the total overdue amount
	OverdueAmount sdk.Coins `json:"overdue_amount"`

	// DiscrepancyCount is the number of discrepancies found
	DiscrepancyCount uint32 `json:"discrepancy_count"`

	// DiscrepancyAmount is the total discrepancy amount
	DiscrepancyAmount sdk.Coins `json:"discrepancy_amount"`
}

// ReconciliationDiscrepancy represents a discrepancy found during reconciliation
type ReconciliationDiscrepancy struct {
	// DiscrepancyID is the unique identifier
	DiscrepancyID string `json:"discrepancy_id"`

	// Type is the type of discrepancy
	Type DiscrepancyType `json:"type"`

	// InvoiceID is the related invoice (if applicable)
	InvoiceID string `json:"invoice_id,omitempty"`

	// SettlementID is the related settlement (if applicable)
	SettlementID string `json:"settlement_id,omitempty"`

	// UsageRecordID is the related usage record (if applicable)
	UsageRecordID string `json:"usage_record_id,omitempty"`

	// ExpectedAmount is the expected amount
	ExpectedAmount sdk.Coins `json:"expected_amount"`

	// ActualAmount is the actual amount
	ActualAmount sdk.Coins `json:"actual_amount"`

	// Difference is the difference amount
	Difference sdk.Coins `json:"difference"`

	// Description describes the discrepancy
	Description string `json:"description"`

	// Severity is the severity level
	Severity DiscrepancySeverity `json:"severity"`

	// Resolution is the recommended resolution
	Resolution string `json:"resolution,omitempty"`

	// ResolvedAt is when the discrepancy was resolved (if resolved)
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// ResolvedBy is who resolved the discrepancy
	ResolvedBy string `json:"resolved_by,omitempty"`
}

// DiscrepancyType defines types of discrepancies
type DiscrepancyType uint8

const (
	// DiscrepancyTypeAmountMismatch is an amount mismatch
	DiscrepancyTypeAmountMismatch DiscrepancyType = 0

	// DiscrepancyTypeMissingInvoice is a missing invoice
	DiscrepancyTypeMissingInvoice DiscrepancyType = 1

	// DiscrepancyTypeMissingSettlement is a missing settlement
	DiscrepancyTypeMissingSettlement DiscrepancyType = 2

	// DiscrepancyTypeMissingUsage is a missing usage record
	DiscrepancyTypeMissingUsage DiscrepancyType = 3

	// DiscrepancyTypeDuplicateInvoice is a duplicate invoice
	DiscrepancyTypeDuplicateInvoice DiscrepancyType = 4

	// DiscrepancyTypeStatusMismatch is a status mismatch
	DiscrepancyTypeStatusMismatch DiscrepancyType = 5

	// DiscrepancyTypeDateMismatch is a date mismatch
	DiscrepancyTypeDateMismatch DiscrepancyType = 6
)

// DiscrepancyTypeNames maps types to names
var DiscrepancyTypeNames = map[DiscrepancyType]string{
	DiscrepancyTypeAmountMismatch:    "amount_mismatch",
	DiscrepancyTypeMissingInvoice:    "missing_invoice",
	DiscrepancyTypeMissingSettlement: "missing_settlement",
	DiscrepancyTypeMissingUsage:      "missing_usage",
	DiscrepancyTypeDuplicateInvoice:  "duplicate_invoice",
	DiscrepancyTypeStatusMismatch:    "status_mismatch",
	DiscrepancyTypeDateMismatch:      "date_mismatch",
}

// String returns string representation
func (t DiscrepancyType) String() string {
	if name, ok := DiscrepancyTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// DiscrepancySeverity defines severity levels
type DiscrepancySeverity uint8

const (
	// DiscrepancySeverityLow is low severity
	DiscrepancySeverityLow DiscrepancySeverity = 0

	// DiscrepancySeverityMedium is medium severity
	DiscrepancySeverityMedium DiscrepancySeverity = 1

	// DiscrepancySeverityHigh is high severity
	DiscrepancySeverityHigh DiscrepancySeverity = 2

	// DiscrepancySeverityCritical is critical severity
	DiscrepancySeverityCritical DiscrepancySeverity = 3
)

// DiscrepancySeverityNames maps severity to names
var DiscrepancySeverityNames = map[DiscrepancySeverity]string{
	DiscrepancySeverityLow:      "low",
	DiscrepancySeverityMedium:   "medium",
	DiscrepancySeverityHigh:     "high",
	DiscrepancySeverityCritical: "critical",
}

// String returns string representation
func (s DiscrepancySeverity) String() string {
	if name, ok := DiscrepancySeverityNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// Validate validates the reconciliation report
func (r *ReconciliationReport) Validate() error {
	if r.ReportID == "" {
		return fmt.Errorf("report_id is required")
	}

	if r.PeriodEnd.Before(r.PeriodStart) {
		return fmt.Errorf("period_end must be after period_start")
	}

	return nil
}

// ReconciliationHookType defines types of reconciliation hooks
type ReconciliationHookType uint8

const (
	// ReconciliationHookPreGenerate runs before report generation
	ReconciliationHookPreGenerate ReconciliationHookType = 0

	// ReconciliationHookPostGenerate runs after report generation
	ReconciliationHookPostGenerate ReconciliationHookType = 1

	// ReconciliationHookOnDiscrepancy runs when discrepancy is found
	ReconciliationHookOnDiscrepancy ReconciliationHookType = 2

	// ReconciliationHookOnComplete runs when reconciliation completes
	ReconciliationHookOnComplete ReconciliationHookType = 3
)

// ReconciliationHookTypeNames maps types to names
var ReconciliationHookTypeNames = map[ReconciliationHookType]string{
	ReconciliationHookPreGenerate:   "pre_generate",
	ReconciliationHookPostGenerate:  "post_generate",
	ReconciliationHookOnDiscrepancy: "on_discrepancy",
	ReconciliationHookOnComplete:    "on_complete",
}

// String returns string representation
func (t ReconciliationHookType) String() string {
	if name, ok := ReconciliationHookTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// ReconciliationHookConfig defines a reconciliation hook
type ReconciliationHookConfig struct {
	// HookID is the unique identifier
	HookID string `json:"hook_id"`

	// HookType is the type of hook
	HookType ReconciliationHookType `json:"hook_type"`

	// Name is the hook name
	Name string `json:"name"`

	// Description describes the hook
	Description string `json:"description"`

	// Priority determines execution order
	Priority uint32 `json:"priority"`

	// IsEnabled indicates if hook is enabled
	IsEnabled bool `json:"is_enabled"`

	// Action is the action to perform
	Action string `json:"action"`

	// Parameters are hook-specific parameters
	Parameters map[string]string `json:"parameters,omitempty"`
}

// Validate validates the hook config
func (h *ReconciliationHookConfig) Validate() error {
	if h.HookID == "" {
		return fmt.Errorf("hook_id is required")
	}

	if h.Name == "" {
		return fmt.Errorf("name is required")
	}

	if h.Action == "" {
		return fmt.Errorf("action is required")
	}

	return nil
}

// ReconciliationHookResult records a hook execution result
type ReconciliationHookResult struct {
	// HookID is the hook that was executed
	HookID string `json:"hook_id"`

	// HookType is the type of hook
	HookType ReconciliationHookType `json:"hook_type"`

	// ReportID is the report this hook was executed for
	ReportID string `json:"report_id"`

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

// DefaultReconciliationHooks returns default reconciliation hooks
func DefaultReconciliationHooks() []ReconciliationHookConfig {
	return []ReconciliationHookConfig{
		{
			HookID:      "validate-invoice-totals",
			HookType:    ReconciliationHookPreGenerate,
			Name:        "Validate Invoice Totals",
			Description: "Validates that invoice totals match line items",
			Priority:    1,
			IsEnabled:   true,
			Action:      "validate_totals",
		},
		{
			HookID:      "check-settlement-coverage",
			HookType:    ReconciliationHookPreGenerate,
			Name:        "Check Settlement Coverage",
			Description: "Checks that all usage records have settlements",
			Priority:    2,
			IsEnabled:   true,
			Action:      "check_coverage",
		},
		{
			HookID:      "emit-reconciliation-event",
			HookType:    ReconciliationHookPostGenerate,
			Name:        "Emit Reconciliation Event",
			Description: "Emits blockchain event for reconciliation",
			Priority:    1,
			IsEnabled:   true,
			Action:      "emit_event",
		},
		{
			HookID:      "alert-on-discrepancy",
			HookType:    ReconciliationHookOnDiscrepancy,
			Name:        "Alert on Discrepancy",
			Description: "Alerts when discrepancy is found",
			Priority:    1,
			IsEnabled:   true,
			Action:      "alert",
		},
	}
}

// Store key prefixes for reconciliation types
var (
	// ReconciliationReportPrefix is the prefix for reconciliation reports
	ReconciliationReportPrefix = []byte{0x70}

	// ReconciliationReportByProviderPrefix indexes reports by provider
	ReconciliationReportByProviderPrefix = []byte{0x71}

	// ReconciliationReportByCustomerPrefix indexes reports by customer
	ReconciliationReportByCustomerPrefix = []byte{0x72}

	// ReconciliationDiscrepancyPrefix is the prefix for discrepancies
	ReconciliationDiscrepancyPrefix = []byte{0x73}

	// ReconciliationHookResultPrefix is the prefix for hook results
	ReconciliationHookResultPrefix = []byte{0x74}
)

// BuildReconciliationReportKey builds the key for a reconciliation report
func BuildReconciliationReportKey(reportID string) []byte {
	return append(ReconciliationReportPrefix, []byte(reportID)...)
}

// BuildReconciliationReportByProviderKey builds the index key
func BuildReconciliationReportByProviderKey(provider string, reportID string) []byte {
	key := make([]byte, 0, len(ReconciliationReportByProviderPrefix)+len(provider)+1+len(reportID))
	key = append(key, ReconciliationReportByProviderPrefix...)
	key = append(key, []byte(provider)...)
	key = append(key, byte('/'))
	return append(key, []byte(reportID)...)
}

// BuildReconciliationReportByCustomerKey builds the index key
func BuildReconciliationReportByCustomerKey(customer string, reportID string) []byte {
	key := make([]byte, 0, len(ReconciliationReportByCustomerPrefix)+len(customer)+1+len(reportID))
	key = append(key, ReconciliationReportByCustomerPrefix...)
	key = append(key, []byte(customer)...)
	key = append(key, byte('/'))
	return append(key, []byte(reportID)...)
}
