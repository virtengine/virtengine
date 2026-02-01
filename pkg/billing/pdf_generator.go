// Copyright 2026 VirtEngine Authors. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the LICENSE file.

// Package billing provides PDF generation for invoice documents.
package billing

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// PDFConfig defines configuration for PDF generation
type PDFConfig struct {
	// CompanyName is the company name to display
	CompanyName string `json:"company_name"`

	// CompanyAddress is the company address
	CompanyAddress []string `json:"company_address"`

	// CompanyEmail is the contact email
	CompanyEmail string `json:"company_email"`

	// CompanyPhone is the contact phone
	CompanyPhone string `json:"company_phone"`

	// LogoPath is the path to the company logo (optional)
	LogoPath string `json:"logo_path,omitempty"`

	// PrimaryColor is the primary brand color (hex)
	PrimaryColor string `json:"primary_color"`

	// SecondaryColor is the secondary brand color (hex)
	SecondaryColor string `json:"secondary_color"`

	// FontFamily is the font family to use
	FontFamily string `json:"font_family"`

	// PageSize is the page size (A4, Letter, etc.)
	PageSize string `json:"page_size"`

	// IncludeTerms includes terms and conditions
	IncludeTerms bool `json:"include_terms"`

	// TermsText is the terms and conditions text
	TermsText string `json:"terms_text,omitempty"`

	// FooterText is the footer text
	FooterText string `json:"footer_text"`

	// ShowTaxBreakdown shows detailed tax breakdown
	ShowTaxBreakdown bool `json:"show_tax_breakdown"`

	// ShowPaymentInstructions shows payment instructions
	ShowPaymentInstructions bool `json:"show_payment_instructions"`

	// PaymentInstructions is the payment instructions text
	PaymentInstructions string `json:"payment_instructions,omitempty"`
}

// DefaultPDFConfig returns the default PDF configuration
func DefaultPDFConfig() PDFConfig {
	return PDFConfig{
		CompanyName:    "VirtEngine",
		CompanyAddress: []string{"VirtEngine Inc.", "Decentralized Cloud Infrastructure"},
		CompanyEmail:   "billing@virtengine.io",
		CompanyPhone:   "",
		PrimaryColor:   "#2563eb",
		SecondaryColor: "#1e40af",
		FontFamily:     "Helvetica",
		PageSize:       "A4",
		IncludeTerms:   true,
		TermsText: `Payment is due within the terms specified on this invoice. Late payments may incur additional fees.
For questions about this invoice, please contact billing@virtengine.io`,
		FooterText:              "Thank you for using VirtEngine - Decentralized Cloud Computing",
		ShowTaxBreakdown:        true,
		ShowPaymentInstructions: true,
		PaymentInstructions: `Payment can be made using VIRT tokens to the escrow address associated with your deployment.
Payments are automatically processed by the VirtEngine blockchain.`,
	}
}

// PDFGenerator generates PDF invoice documents
type PDFGenerator struct {
	config PDFConfig
}

// NewPDFGenerator creates a new PDF generator
func NewPDFGenerator(config PDFConfig) *PDFGenerator {
	return &PDFGenerator{config: config}
}

// GenerateInvoicePDF generates a PDF for the given invoice
// This generates a simple text-based PDF representation.
// For production use, this would integrate with a proper PDF library.
func (g *PDFGenerator) GenerateInvoicePDF(invoice *billing.Invoice) ([]byte, error) {
	if invoice == nil {
		return nil, fmt.Errorf("invoice is nil")
	}

	// Build the PDF content as a structured document
	doc := g.buildInvoiceDocument(invoice)

	// For now, generate a simple PDF representation
	// In production, this would use a library like go-pdf or similar
	return g.renderToPDF(doc)
}

// InvoicePDFDocument represents the structured PDF content
type InvoicePDFDocument struct {
	Header       PDFHeader         `json:"header"`
	InvoiceInfo  PDFInvoiceInfo    `json:"invoice_info"`
	Parties      PDFParties        `json:"parties"`
	LineItems    []PDFLineItem     `json:"line_items"`
	Summary      PDFSummary        `json:"summary"`
	TaxDetails   *PDFTaxDetails    `json:"tax_details,omitempty"`
	Notes        string            `json:"notes,omitempty"`
	Terms        string            `json:"terms,omitempty"`
	Footer       string            `json:"footer"`
	GeneratedAt  time.Time         `json:"generated_at"`
}

// PDFHeader contains header information
type PDFHeader struct {
	CompanyName    string   `json:"company_name"`
	CompanyAddress []string `json:"company_address"`
	CompanyEmail   string   `json:"company_email"`
	CompanyPhone   string   `json:"company_phone"`
	LogoPath       string   `json:"logo_path,omitempty"`
}

// PDFInvoiceInfo contains invoice metadata
type PDFInvoiceInfo struct {
	InvoiceNumber   string    `json:"invoice_number"`
	InvoiceID       string    `json:"invoice_id"`
	IssueDate       time.Time `json:"issue_date"`
	DueDate         time.Time `json:"due_date"`
	BillingPeriod   string    `json:"billing_period"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	EscrowID        string    `json:"escrow_id"`
	OrderID         string    `json:"order_id"`
}

// PDFParties contains party information
type PDFParties struct {
	Provider PDFParty `json:"provider"`
	Customer PDFParty `json:"customer"`
}

// PDFParty represents a party in the invoice
type PDFParty struct {
	Address string `json:"address"`
	Label   string `json:"label"`
}

// PDFLineItem represents a line item in the PDF
type PDFLineItem struct {
	Number      int    `json:"number"`
	Description string `json:"description"`
	UsageType   string `json:"usage_type"`
	Quantity    string `json:"quantity"`
	Unit        string `json:"unit"`
	UnitPrice   string `json:"unit_price"`
	Amount      string `json:"amount"`
}

// PDFSummary contains invoice totals
type PDFSummary struct {
	Subtotal      string          `json:"subtotal"`
	Discounts     []PDFDiscount   `json:"discounts,omitempty"`
	DiscountTotal string          `json:"discount_total,omitempty"`
	TaxTotal      string          `json:"tax_total,omitempty"`
	Total         string          `json:"total"`
	AmountPaid    string          `json:"amount_paid"`
	AmountDue     string          `json:"amount_due"`
}

// PDFDiscount represents a discount in the PDF
type PDFDiscount struct {
	Description string `json:"description"`
	Amount      string `json:"amount"`
}

// PDFTaxDetails contains tax breakdown
type PDFTaxDetails struct {
	CustomerTaxID string        `json:"customer_tax_id,omitempty"`
	ProviderTaxID string        `json:"provider_tax_id,omitempty"`
	Jurisdiction  string        `json:"jurisdiction,omitempty"`
	TaxLines      []PDFTaxLine  `json:"tax_lines,omitempty"`
}

// PDFTaxLine represents a tax line item
type PDFTaxLine struct {
	Description string `json:"description"`
	Rate        string `json:"rate"`
	Amount      string `json:"amount"`
}

// buildInvoiceDocument builds the structured document from an invoice
func (g *PDFGenerator) buildInvoiceDocument(invoice *billing.Invoice) *InvoicePDFDocument {
	doc := &InvoicePDFDocument{
		Header: PDFHeader{
			CompanyName:    g.config.CompanyName,
			CompanyAddress: g.config.CompanyAddress,
			CompanyEmail:   g.config.CompanyEmail,
			CompanyPhone:   g.config.CompanyPhone,
			LogoPath:       g.config.LogoPath,
		},
		InvoiceInfo: PDFInvoiceInfo{
			InvoiceNumber: invoice.InvoiceNumber,
			InvoiceID:     invoice.InvoiceID,
			IssueDate:     invoice.IssuedAt,
			DueDate:       invoice.DueDate,
			BillingPeriod: fmt.Sprintf("%s - %s",
				invoice.BillingPeriod.StartTime.Format("2006-01-02"),
				invoice.BillingPeriod.EndTime.Format("2006-01-02")),
			Currency: invoice.Currency,
			Status:   invoice.Status.String(),
			EscrowID: invoice.EscrowID,
			OrderID:  invoice.OrderID,
		},
		Parties: PDFParties{
			Provider: PDFParty{
				Address: invoice.Provider,
				Label:   "Provider",
			},
			Customer: PDFParty{
				Address: invoice.Customer,
				Label:   "Customer",
			},
		},
		Footer:      g.config.FooterText,
		GeneratedAt: time.Now().UTC(),
	}

	// Add line items
	for i, item := range invoice.LineItems {
		doc.LineItems = append(doc.LineItems, PDFLineItem{
			Number:      i + 1,
			Description: item.Description,
			UsageType:   item.UsageType.String(),
			Quantity:    item.Quantity.String(),
			Unit:        item.Unit,
			UnitPrice:   item.UnitPrice.String(),
			Amount:      item.Amount.String(),
		})
	}

	// Build summary
	doc.Summary = PDFSummary{
		Subtotal:   invoice.Subtotal.String(),
		Total:      invoice.Total.String(),
		AmountPaid: invoice.AmountPaid.String(),
		AmountDue:  invoice.AmountDue.String(),
	}

	// Add discounts
	if len(invoice.Discounts) > 0 {
		for _, d := range invoice.Discounts {
			doc.Summary.Discounts = append(doc.Summary.Discounts, PDFDiscount{
				Description: d.Description,
				Amount:      d.Amount.String(),
			})
		}
		doc.Summary.DiscountTotal = invoice.DiscountTotal.String()
	}

	// Add tax if present
	if invoice.TaxDetails != nil && g.config.ShowTaxBreakdown {
		doc.TaxDetails = &PDFTaxDetails{
			CustomerTaxID: invoice.TaxDetails.CustomerTaxID,
			ProviderTaxID: invoice.TaxDetails.ProviderTaxID,
			Jurisdiction:  invoice.TaxDetails.Jurisdiction.CountryCode,
		}

		for _, taxLine := range invoice.TaxDetails.TaxLineItems {
			doc.TaxDetails.TaxLines = append(doc.TaxDetails.TaxLines, PDFTaxLine{
				Description: taxLine.Name,
				Rate:        fmt.Sprintf("%.2f%%", float64(taxLine.RateBps)/100),
				Amount:      taxLine.TaxAmount.String(),
			})
		}
		doc.Summary.TaxTotal = invoice.TaxTotal.String()
	}

	// Add terms if configured
	if g.config.IncludeTerms {
		doc.Terms = g.config.TermsText
	}

	// Add payment instructions if configured
	if g.config.ShowPaymentInstructions {
		doc.Notes = g.config.PaymentInstructions
	}

	return doc
}

// renderToPDF renders the document to PDF format
// This is a simplified text-based PDF representation.
// For production, integrate with a proper PDF library like go-pdf, gofpdf, or pdfcpu.
func (g *PDFGenerator) renderToPDF(doc *InvoicePDFDocument) ([]byte, error) {
	var buf bytes.Buffer

	// Write PDF header
	buf.WriteString("%PDF-1.7\n")
	buf.WriteString("% VirtEngine Invoice PDF\n\n")

	// For a real implementation, this would use a proper PDF library.
	// Here we generate a simple text-based structure that represents the invoice data.
	// The actual PDF rendering would be handled by a library like:
	// - github.com/jung-kurt/gofpdf
	// - github.com/pdfcpu/pdfcpu
	// - github.com/signintech/gopdf

	// Write document content as structured text (simplified)
	content := g.buildTextContent(doc)
	
	// Object 1: Catalog
	buf.WriteString("1 0 obj\n")
	buf.WriteString("<< /Type /Catalog /Pages 2 0 R >>\n")
	buf.WriteString("endobj\n\n")

	// Object 2: Pages
	buf.WriteString("2 0 obj\n")
	buf.WriteString("<< /Type /Pages /Kids [3 0 R] /Count 1 >>\n")
	buf.WriteString("endobj\n\n")

	// Object 3: Page
	buf.WriteString("3 0 obj\n")
	buf.WriteString("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] ")
	buf.WriteString("/Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\n")
	buf.WriteString("endobj\n\n")

	// Object 4: Content stream
	stream := fmt.Sprintf("BT\n/F1 10 Tf\n50 742 Td\n(%s)Tj\nET", escapeForPDF(content))
	buf.WriteString("4 0 obj\n")
	buf.WriteString(fmt.Sprintf("<< /Length %d >>\n", len(stream)))
	buf.WriteString("stream\n")
	buf.WriteString(stream)
	buf.WriteString("\nendstream\n")
	buf.WriteString("endobj\n\n")

	// Object 5: Font
	buf.WriteString("5 0 obj\n")
	buf.WriteString("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\n")
	buf.WriteString("endobj\n\n")

	// xref table
	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	buf.WriteString("0 6\n")
	buf.WriteString("0000000000 65535 f \n")
	buf.WriteString("0000000009 00000 n \n")
	buf.WriteString("0000000058 00000 n \n")
	buf.WriteString("0000000115 00000 n \n")
	buf.WriteString("0000000266 00000 n \n")
	buf.WriteString("0000000400 00000 n \n")

	// Trailer
	buf.WriteString("trailer\n")
	buf.WriteString("<< /Size 6 /Root 1 0 R >>\n")
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefOffset))
	buf.WriteString("%%EOF\n")

	return buf.Bytes(), nil
}

// buildTextContent builds a text representation of the invoice
func (g *PDFGenerator) buildTextContent(doc *InvoicePDFDocument) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("%s\\n", doc.Header.CompanyName))
	for _, line := range doc.Header.CompanyAddress {
		sb.WriteString(fmt.Sprintf("%s\\n", line))
	}
	sb.WriteString("\\n")

	// Invoice Info
	sb.WriteString(fmt.Sprintf("INVOICE %s\\n", doc.InvoiceInfo.InvoiceNumber))
	sb.WriteString(fmt.Sprintf("Date: %s\\n", doc.InvoiceInfo.IssueDate.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("Due: %s\\n", doc.InvoiceInfo.DueDate.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("Status: %s\\n", doc.InvoiceInfo.Status))
	sb.WriteString("\\n")

	// Parties
	sb.WriteString(fmt.Sprintf("Provider: %s\\n", truncateAddress(doc.Parties.Provider.Address)))
	sb.WriteString(fmt.Sprintf("Customer: %s\\n", truncateAddress(doc.Parties.Customer.Address)))
	sb.WriteString("\\n")

	// Line Items
	sb.WriteString("Line Items:\\n")
	for _, item := range doc.LineItems {
		sb.WriteString(fmt.Sprintf("%d. %s - %s %s @ %s = %s\\n",
			item.Number, item.Description, item.Quantity, item.Unit,
			item.UnitPrice, item.Amount))
	}
	sb.WriteString("\\n")

	// Summary
	sb.WriteString(fmt.Sprintf("Subtotal: %s\\n", doc.Summary.Subtotal))
	if len(doc.Summary.Discounts) > 0 {
		sb.WriteString(fmt.Sprintf("Discounts: -%s\\n", doc.Summary.DiscountTotal))
	}
	if doc.Summary.TaxTotal != "" {
		sb.WriteString(fmt.Sprintf("Tax: %s\\n", doc.Summary.TaxTotal))
	}
	sb.WriteString(fmt.Sprintf("Total: %s\\n", doc.Summary.Total))
	sb.WriteString(fmt.Sprintf("Paid: %s\\n", doc.Summary.AmountPaid))
	sb.WriteString(fmt.Sprintf("Due: %s\\n", doc.Summary.AmountDue))

	return sb.String()
}

// escapeForPDF escapes special characters for PDF text
func escapeForPDF(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

// truncateAddress truncates a bech32 address for display
func truncateAddress(addr string) string {
	if len(addr) <= 20 {
		return addr
	}
	return addr[:10] + "..." + addr[len(addr)-8:]
}

// GenerateReceiptPDF generates a payment receipt PDF
func (g *PDFGenerator) GenerateReceiptPDF(
	invoice *billing.Invoice,
	paymentAmount string,
	paymentDate time.Time,
	txHash string,
) ([]byte, error) {
	// Receipt is a simplified version of the invoice PDF
	doc := g.buildInvoiceDocument(invoice)
	
	// Modify for receipt
	doc.InvoiceInfo.Status = "PAID"
	doc.Notes = fmt.Sprintf("Payment received: %s on %s\\nTransaction: %s",
		paymentAmount, paymentDate.Format("2006-01-02 15:04:05 UTC"), txHash)

	return g.renderToPDF(doc)
}
