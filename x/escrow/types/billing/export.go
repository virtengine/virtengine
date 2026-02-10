// Package billing provides billing and invoice types for the escrow module.
package billing

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// ExportFormat defines the output format for exports
type ExportFormat uint8

const (
	// ExportFormatCSV exports data as CSV
	ExportFormatCSV ExportFormat = 0

	// ExportFormatJSON exports data as JSON
	ExportFormatJSON ExportFormat = 1

	// ExportFormatExcel exports data as Excel (XLSX)
	ExportFormatExcel ExportFormat = 2
)

// ExportFormatNames maps format to human-readable names
var ExportFormatNames = map[ExportFormat]string{
	ExportFormatCSV:   "csv",
	ExportFormatJSON:  "json",
	ExportFormatExcel: "excel",
}

// String returns string representation
func (f ExportFormat) String() string {
	if name, ok := ExportFormatNames[f]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", f)
}

// ContentType returns the MIME type for the export format
func (f ExportFormat) ContentType() string {
	switch f {
	case ExportFormatCSV:
		return "text/csv"
	case ExportFormatJSON:
		return "application/json"
	case ExportFormatExcel:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}

// FileExtension returns the file extension for the export format
func (f ExportFormat) FileExtension() string {
	switch f {
	case ExportFormatCSV:
		return ".csv"
	case ExportFormatJSON:
		return ".json"
	case ExportFormatExcel:
		return ".xlsx"
	default:
		return ".bin"
	}
}

// ExportType defines the type of export report
type ExportType uint8

const (
	// ExportTypeReconciliationReport exports reconciliation data
	ExportTypeReconciliationReport ExportType = 0

	// ExportTypeDisputeReport exports dispute data
	ExportTypeDisputeReport ExportType = 1

	// ExportTypeInvoiceSummary exports invoice summary data
	ExportTypeInvoiceSummary ExportType = 2

	// ExportTypeSettlementSummary exports settlement summary data
	ExportTypeSettlementSummary ExportType = 3

	// ExportTypeCorrectionReport exports correction/adjustment data
	ExportTypeCorrectionReport ExportType = 4

	// ExportTypePayoutReport exports payout data
	ExportTypePayoutReport ExportType = 5

	// ExportTypeAuditLog exports audit log data
	ExportTypeAuditLog ExportType = 6

	// ExportTypeComplianceReport exports compliance data
	ExportTypeComplianceReport ExportType = 7
)

// ExportTypeNames maps export type to human-readable names
var ExportTypeNames = map[ExportType]string{
	ExportTypeReconciliationReport: "reconciliation_report",
	ExportTypeDisputeReport:        "dispute_report",
	ExportTypeInvoiceSummary:       "invoice_summary",
	ExportTypeSettlementSummary:    "settlement_summary",
	ExportTypeCorrectionReport:     "correction_report",
	ExportTypePayoutReport:         "payout_report",
	ExportTypeAuditLog:             "audit_log",
	ExportTypeComplianceReport:     "compliance_report",
}

// String returns string representation
func (t ExportType) String() string {
	if name, ok := ExportTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// ExportStatus defines the status of an export request
type ExportStatus uint8

const (
	// ExportStatusPending is waiting to be processed
	ExportStatusPending ExportStatus = 0

	// ExportStatusInProgress is currently being generated
	ExportStatusInProgress ExportStatus = 1

	// ExportStatusCompleted has been successfully generated
	ExportStatusCompleted ExportStatus = 2

	// ExportStatusFailed failed to generate
	ExportStatusFailed ExportStatus = 3
)

// ExportStatusNames maps status to human-readable names
var ExportStatusNames = map[ExportStatus]string{
	ExportStatusPending:    "pending",
	ExportStatusInProgress: "in_progress",
	ExportStatusCompleted:  "completed",
	ExportStatusFailed:     "failed",
}

// String returns string representation
func (s ExportStatus) String() string {
	if name, ok := ExportStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// IsTerminal returns true if the status is final
func (s ExportStatus) IsTerminal() bool {
	return s == ExportStatusCompleted || s == ExportStatusFailed
}

// ExportFilter defines filters for export data selection
type ExportFilter struct {
	// StartTime is the start of the time range
	StartTime time.Time `json:"start_time"`

	// EndTime is the end of the time range
	EndTime time.Time `json:"end_time"`

	// Provider filters by provider address (optional)
	Provider string `json:"provider,omitempty"`

	// Customer filters by customer address (optional)
	Customer string `json:"customer,omitempty"`

	// InvoiceStatuses filters by invoice status (optional)
	InvoiceStatuses []InvoiceStatus `json:"invoice_statuses,omitempty"`

	// DisputeStatuses filters by dispute status (optional)
	DisputeStatuses []DisputeStatus `json:"dispute_statuses,omitempty"`

	// MinAmount filters by minimum amount (optional)
	MinAmount sdk.Coins `json:"min_amount,omitempty"`

	// MaxAmount filters by maximum amount (optional)
	MaxAmount sdk.Coins `json:"max_amount,omitempty"`
}

// Validate validates the export filter
func (f *ExportFilter) Validate() error {
	if f.StartTime.IsZero() {
		return fmt.Errorf("start_time is required")
	}

	if f.EndTime.IsZero() {
		return fmt.Errorf("end_time is required")
	}

	if f.EndTime.Before(f.StartTime) {
		return fmt.Errorf("end_time must be after start_time")
	}

	// Validate provider address if provided
	if f.Provider != "" {
		if _, err := sdk.AccAddressFromBech32(f.Provider); err != nil {
			return fmt.Errorf("invalid provider address: %w", err)
		}
	}

	// Validate customer address if provided
	if f.Customer != "" {
		if _, err := sdk.AccAddressFromBech32(f.Customer); err != nil {
			return fmt.Errorf("invalid customer address: %w", err)
		}
	}

	// Validate amount filters
	if !f.MinAmount.Empty() && !f.MinAmount.IsValid() {
		return fmt.Errorf("min_amount must be valid coins")
	}

	if !f.MaxAmount.Empty() && !f.MaxAmount.IsValid() {
		return fmt.Errorf("max_amount must be valid coins")
	}

	return nil
}

// ExportRequest represents a request to export billing data
type ExportRequest struct {
	// RequestID is the unique identifier for this export request
	RequestID string `json:"request_id"`

	// RequestedBy is the address that requested the export
	RequestedBy string `json:"requested_by"`

	// ExportType is the type of data to export
	ExportType ExportType `json:"export_type"`

	// Format is the output format
	Format ExportFormat `json:"format"`

	// Filter contains the filter criteria
	Filter ExportFilter `json:"filter"`

	// IncludeLineItems includes line item details in exports
	IncludeLineItems bool `json:"include_line_items"`

	// IncludeAuditTrail includes audit trail in exports
	IncludeAuditTrail bool `json:"include_audit_trail"`

	// RequestedAt is when the export was requested
	RequestedAt time.Time `json:"requested_at"`

	// CompletedAt is when the export was completed (if completed)
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Status is the current status of the export request
	Status ExportStatus `json:"status"`

	// OutputArtifactCID is the CID of the exported file (when completed)
	OutputArtifactCID string `json:"output_artifact_cid,omitempty"`

	// FileSize is the size of the exported file in bytes
	FileSize int64 `json:"file_size,omitempty"`

	// RecordCount is the number of records in the export
	RecordCount uint32 `json:"record_count,omitempty"`

	// ErrorMessage contains error details if the export failed
	ErrorMessage string `json:"error_message,omitempty"`
}

// NewExportRequest creates a new export request
func NewExportRequest(
	requestID string,
	requestedBy string,
	exportType ExportType,
	format ExportFormat,
	filter ExportFilter,
	includeLineItems bool,
	includeAuditTrail bool,
	now time.Time,
) *ExportRequest {
	return &ExportRequest{
		RequestID:         requestID,
		RequestedBy:       requestedBy,
		ExportType:        exportType,
		Format:            format,
		Filter:            filter,
		IncludeLineItems:  includeLineItems,
		IncludeAuditTrail: includeAuditTrail,
		RequestedAt:       now,
		Status:            ExportStatusPending,
	}
}

// Validate validates the export request
func (r *ExportRequest) Validate() error {
	if r.RequestID == "" {
		return fmt.Errorf("request_id is required")
	}

	if len(r.RequestID) > 64 {
		return fmt.Errorf("request_id exceeds maximum length of 64")
	}

	if r.RequestedBy == "" {
		return fmt.Errorf("requested_by is required")
	}

	if _, err := sdk.AccAddressFromBech32(r.RequestedBy); err != nil {
		return fmt.Errorf("invalid requested_by address: %w", err)
	}

	// Validate export type
	if _, ok := ExportTypeNames[r.ExportType]; !ok {
		return fmt.Errorf("invalid export_type: %d", r.ExportType)
	}

	// Validate format
	if _, ok := ExportFormatNames[r.Format]; !ok {
		return fmt.Errorf("invalid format: %d", r.Format)
	}

	// Validate filter
	if err := r.Filter.Validate(); err != nil {
		return fmt.Errorf("invalid filter: %w", err)
	}

	return nil
}

// MarkInProgress marks the export as in progress
func (r *ExportRequest) MarkInProgress() error {
	if r.Status != ExportStatusPending {
		return fmt.Errorf("can only start pending exports, current status: %s", r.Status)
	}
	r.Status = ExportStatusInProgress
	return nil
}

// MarkCompleted marks the export as completed
func (r *ExportRequest) MarkCompleted(artifactCID string, fileSize int64, recordCount uint32, now time.Time) error {
	if r.Status != ExportStatusInProgress {
		return fmt.Errorf("can only complete in-progress exports, current status: %s", r.Status)
	}

	if artifactCID == "" {
		return fmt.Errorf("artifact_cid is required for completed exports")
	}

	r.Status = ExportStatusCompleted
	r.OutputArtifactCID = artifactCID
	r.FileSize = fileSize
	r.RecordCount = recordCount
	r.CompletedAt = &now
	return nil
}

// MarkFailed marks the export as failed
func (r *ExportRequest) MarkFailed(errorMessage string, now time.Time) error {
	if r.Status.IsTerminal() {
		return fmt.Errorf("cannot fail terminal export, current status: %s", r.Status)
	}

	r.Status = ExportStatusFailed
	r.ErrorMessage = errorMessage
	r.CompletedAt = &now
	return nil
}

// ExportService defines the interface for the export service
type ExportService interface {
	// RequestExport creates a new export request
	RequestExport(ctx context.Context, req *ExportRequest) error

	// GetExportStatus retrieves the status of an export request
	GetExportStatus(ctx context.Context, requestID string) (*ExportRequest, error)

	// GenerateCSV generates a CSV export for the given type and filter
	GenerateCSV(ctx context.Context, exportType ExportType, filter ExportFilter) ([]byte, error)

	// GenerateJSON generates a JSON export for the given type and filter
	GenerateJSON(ctx context.Context, exportType ExportType, filter ExportFilter) ([]byte, error)

	// GetExportsByRequester retrieves all exports for a requester with pagination
	GetExportsByRequester(ctx context.Context, requester string, pagination *query.PageRequest) ([]*ExportRequest, *query.PageResponse, error)
}

// CSV Column Definitions for each export type

// ReconciliationReportCSVColumns defines CSV columns for reconciliation reports
var ReconciliationReportCSVColumns = []string{
	"report_id",
	"report_type",
	"period_start",
	"period_end",
	"provider",
	"customer",
	"status",
	"total_invoices",
	"total_invoice_amount",
	"total_settlements",
	"total_settlement_amount",
	"paid_invoices",
	"paid_amount",
	"outstanding_invoices",
	"outstanding_amount",
	"disputed_invoices",
	"disputed_amount",
	"discrepancy_count",
	"discrepancy_amount",
	"generated_at",
	"generated_by",
}

// DisputeReportCSVColumns defines CSV columns for dispute reports
var DisputeReportCSVColumns = []string{
	"window_id",
	"invoice_id",
	"status",
	"resolution",
	"disputed_by",
	"disputed_at",
	"dispute_reason",
	"resolved_by",
	"resolved_at",
	"resolution_details",
	"refund_amount",
	"window_start_time",
	"window_end_time",
	"escalation_count",
}

// InvoiceSummaryCSVColumns defines CSV columns for invoice summaries
var InvoiceSummaryCSVColumns = []string{
	"invoice_id",
	"invoice_number",
	"escrow_id",
	"order_id",
	"lease_id",
	"provider",
	"customer",
	"status",
	"billing_period_start",
	"billing_period_end",
	"billing_period_type",
	"subtotal",
	"discount_total",
	"tax_total",
	"total",
	"amount_paid",
	"amount_due",
	"currency",
	"due_date",
	"issued_at",
	"paid_at",
}

// SettlementSummaryCSVColumns defines CSV columns for settlement summaries
var SettlementSummaryCSVColumns = []string{
	"settlement_id",
	"escrow_id",
	"provider",
	"customer",
	"invoice_id",
	"settlement_amount",
	"currency",
	"settled_at",
	"block_height",
	"tx_hash",
}

// CorrectionReportCSVColumns defines CSV columns for correction reports
var CorrectionReportCSVColumns = []string{
	"correction_id",
	"invoice_id",
	"correction_type",
	"original_amount",
	"corrected_amount",
	"difference",
	"reason",
	"corrected_by",
	"corrected_at",
	"approved_by",
	"approved_at",
}

// PayoutReportCSVColumns defines CSV columns for payout reports
var PayoutReportCSVColumns = []string{
	"payout_id",
	"provider",
	"payout_amount",
	"currency",
	"invoice_count",
	"settlement_count",
	"payout_status",
	"initiated_at",
	"completed_at",
	"tx_hash",
}

// AuditLogCSVColumns defines CSV columns for audit logs
var AuditLogCSVColumns = []string{
	"log_id",
	"timestamp",
	"actor",
	"action",
	"resource_type",
	"resource_id",
	"old_value",
	"new_value",
	"ip_address",
	"user_agent",
	"block_height",
}

// ComplianceReportCSVColumns defines CSV columns for compliance reports
var ComplianceReportCSVColumns = []string{
	"report_id",
	"report_type",
	"period_start",
	"period_end",
	"jurisdiction",
	"tax_id",
	"total_revenue",
	"total_tax_collected",
	"total_refunds",
	"net_revenue",
	"transaction_count",
	"generated_at",
	"generated_by",
}

// GetCSVColumnsForExportType returns the CSV columns for the given export type
func GetCSVColumnsForExportType(exportType ExportType) []string {
	switch exportType {
	case ExportTypeReconciliationReport:
		return ReconciliationReportCSVColumns
	case ExportTypeDisputeReport:
		return DisputeReportCSVColumns
	case ExportTypeInvoiceSummary:
		return InvoiceSummaryCSVColumns
	case ExportTypeSettlementSummary:
		return SettlementSummaryCSVColumns
	case ExportTypeCorrectionReport:
		return CorrectionReportCSVColumns
	case ExportTypePayoutReport:
		return PayoutReportCSVColumns
	case ExportTypeAuditLog:
		return AuditLogCSVColumns
	case ExportTypeComplianceReport:
		return ComplianceReportCSVColumns
	default:
		return nil
	}
}

// Store key prefixes for export types
var (
	// ExportRequestPrefix is the prefix for export request storage
	ExportRequestPrefix = []byte{0xB0}

	// ExportRequestByRequesterPrefix indexes export requests by requester
	ExportRequestByRequesterPrefix = []byte{0xB1}
)

// BuildExportRequestKey builds the key for an export request
func BuildExportRequestKey(requestID string) []byte {
	return append(ExportRequestPrefix, []byte(requestID)...)
}

// ParseExportRequestKey parses an export request key
func ParseExportRequestKey(key []byte) (string, error) {
	if len(key) <= len(ExportRequestPrefix) {
		return "", fmt.Errorf("invalid export request key length")
	}
	return string(key[len(ExportRequestPrefix):]), nil
}

// BuildExportRequestByRequesterKey builds the index key for exports by requester
func BuildExportRequestByRequesterKey(requester string, requestID string) []byte {
	key := make([]byte, 0, len(ExportRequestByRequesterPrefix)+len(requester)+len(requestID)+1)
	key = append(key, ExportRequestByRequesterPrefix...)
	key = append(key, []byte(requester)...)
	key = append(key, byte('/'))
	return append(key, []byte(requestID)...)
}

// BuildExportRequestByRequesterPrefix builds the prefix for requester's exports
func BuildExportRequestByRequesterPrefix(requester string) []byte {
	key := make([]byte, 0, len(ExportRequestByRequesterPrefix)+len(requester)+1)
	key = append(key, ExportRequestByRequesterPrefix...)
	key = append(key, []byte(requester)...)
	return append(key, byte('/'))
}
