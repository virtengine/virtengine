// Copyright 2026 VirtEngine Authors. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the LICENSE file.

// Package billing provides billing export services including PDF generation
// for invoice documents in the VirtEngine marketplace.
package billing

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// ExportService provides invoice export functionality
type ExportService struct {
	pdfGenerator *PDFGenerator
}

// NewExportService creates a new export service
func NewExportService() *ExportService {
	return &ExportService{
		pdfGenerator: NewPDFGenerator(DefaultPDFConfig()),
	}
}

// NewExportServiceWithConfig creates a new export service with custom PDF config
func NewExportServiceWithConfig(pdfConfig PDFConfig) *ExportService {
	return &ExportService{
		pdfGenerator: NewPDFGenerator(pdfConfig),
	}
}

// ExportInvoicePDF exports an invoice as a PDF document
func (s *ExportService) ExportInvoicePDF(ctx context.Context, invoice *billing.Invoice) ([]byte, error) {
	if invoice == nil {
		return nil, fmt.Errorf("invoice is nil")
	}

	if err := invoice.Validate(); err != nil {
		return nil, fmt.Errorf("invalid invoice: %w", err)
	}

	return s.pdfGenerator.GenerateInvoicePDF(invoice)
}

// ExportInvoiceJSON exports an invoice as a JSON document
func (s *ExportService) ExportInvoiceJSON(ctx context.Context, invoice *billing.Invoice) ([]byte, error) {
	if invoice == nil {
		return nil, fmt.Errorf("invoice is nil")
	}

	doc := InvoiceJSONExport{
		Version:       "1.0",
		SchemaVersion: "virtengine/invoice/v1",
		ExportedAt:    time.Now().UTC(),
		Invoice:       invoice,
	}

	return json.MarshalIndent(doc, "", "  ")
}

// ExportInvoiceCSV exports an invoice as a CSV document
func (s *ExportService) ExportInvoiceCSV(ctx context.Context, invoice *billing.Invoice) ([]byte, error) {
	if invoice == nil {
		return nil, fmt.Errorf("invoice is nil")
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"invoice_id", "invoice_number", "provider", "customer", "status",
		"line_item_id", "description", "usage_type", "quantity", "unit",
		"unit_price_denom", "unit_price_amount", "amount",
	}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write line items
	for _, item := range invoice.LineItems {
		row := []string{
			invoice.InvoiceID,
			invoice.InvoiceNumber,
			invoice.Provider,
			invoice.Customer,
			invoice.Status.String(),
			item.LineItemID,
			item.Description,
			item.UsageType.String(),
			item.Quantity.String(),
			item.Unit,
			item.UnitPrice.Denom,
			item.UnitPrice.Amount.String(),
			item.Amount.String(),
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportInvoiceSummaryCSV exports multiple invoices as a summary CSV
func (s *ExportService) ExportInvoiceSummaryCSV(ctx context.Context, invoices []*billing.InvoiceLedgerRecord) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	if err := writer.Write(billing.InvoiceSummaryCSVColumns); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, inv := range invoices {
		paidAt := ""
		if inv.PaidAt != nil {
			paidAt = inv.PaidAt.Format(time.RFC3339)
		}

		row := []string{
			inv.InvoiceID,
			inv.InvoiceNumber,
			inv.EscrowID,
			inv.OrderID,
			inv.LeaseID,
			inv.Provider,
			inv.Customer,
			inv.Status.String(),
			inv.BillingPeriodStart.Format(time.RFC3339),
			inv.BillingPeriodEnd.Format(time.RFC3339),
			"", // billing_period_type not in ledger record
			inv.Subtotal.String(),
			inv.DiscountTotal.String(),
			inv.TaxTotal.String(),
			inv.Total.String(),
			inv.AmountPaid.String(),
			inv.AmountDue.String(),
			inv.Currency,
			inv.DueDate.Format(time.RFC3339),
			inv.IssuedAt.Format(time.RFC3339),
			paidAt,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// ExportSettlementSummaryCSV exports settlement data as CSV
func (s *ExportService) ExportSettlementSummaryCSV(
	ctx context.Context,
	settlements []SettlementExportRecord,
) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	if err := writer.Write(billing.SettlementSummaryCSVColumns); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, s := range settlements {
		row := []string{
			s.SettlementID,
			s.EscrowID,
			s.Provider,
			s.Customer,
			s.InvoiceID,
			s.Amount.String(),
			s.Currency,
			s.SettledAt.Format(time.RFC3339),
			fmt.Sprintf("%d", s.BlockHeight),
			s.TxHash,
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// InvoiceJSONExport is the JSON export document structure
type InvoiceJSONExport struct {
	Version       string           `json:"version"`
	SchemaVersion string           `json:"schema_version"`
	ExportedAt    time.Time        `json:"exported_at"`
	Invoice       *billing.Invoice `json:"invoice"`
}

// SettlementExportRecord represents a settlement for export
type SettlementExportRecord struct {
	SettlementID string    `json:"settlement_id"`
	EscrowID     string    `json:"escrow_id"`
	Provider     string    `json:"provider"`
	Customer     string    `json:"customer"`
	InvoiceID    string    `json:"invoice_id"`
	Amount       sdk.Coins `json:"amount"`
	Currency     string    `json:"currency"`
	SettledAt    time.Time `json:"settled_at"`
	BlockHeight  int64     `json:"block_height"`
	TxHash       string    `json:"tx_hash"`
}

// BatchExportRequest defines a batch export request
type BatchExportRequest struct {
	// ExportType is the type of export
	ExportType billing.ExportType `json:"export_type"`

	// Format is the output format
	Format billing.ExportFormat `json:"format"`

	// Filter is the filter criteria
	Filter billing.ExportFilter `json:"filter"`

	// InvoiceIDs are specific invoices to export (optional)
	InvoiceIDs []string `json:"invoice_ids,omitempty"`
}

// BatchExportResult is the result of a batch export
type BatchExportResult struct {
	// Data is the exported data
	Data []byte `json:"data"`

	// RecordCount is the number of records exported
	RecordCount uint32 `json:"record_count"`

	// ContentType is the MIME type
	ContentType string `json:"content_type"`

	// FileExtension is the file extension
	FileExtension string `json:"file_extension"`

	// GeneratedAt is when the export was generated
	GeneratedAt time.Time `json:"generated_at"`
}

// ExportBatch performs a batch export based on the request
func (s *ExportService) ExportBatch(ctx context.Context, req BatchExportRequest) (*BatchExportResult, error) {
	if err := req.Filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	result := &BatchExportResult{
		ContentType:   req.Format.ContentType(),
		FileExtension: req.Format.FileExtension(),
		GeneratedAt:   time.Now().UTC(),
	}

	return result, nil
}

// GetSupportedFormats returns the list of supported export formats
func (s *ExportService) GetSupportedFormats() []billing.ExportFormat {
	return []billing.ExportFormat{
		billing.ExportFormatCSV,
		billing.ExportFormatJSON,
		ExportFormatPDF,
	}
}

// ExportFormatPDF is the PDF export format (extending the billing package formats)
const ExportFormatPDF billing.ExportFormat = 3

// PDFFormatName is the name of the PDF format
const PDFFormatName = "pdf"

// PDFContentType is the MIME type for PDF
const PDFContentType = "application/pdf"

// PDFFileExtension is the file extension for PDF
const PDFFileExtension = ".pdf"
