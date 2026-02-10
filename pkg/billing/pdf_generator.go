// Copyright 2026 VirtEngine Authors. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the LICENSE file.

// Package billing provides PDF generation for invoice documents.
package billing

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/jung-kurt/gofpdf"

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

// GenerateInvoicePDF generates a PDF for the given invoice.
func (g *PDFGenerator) GenerateInvoicePDF(invoice *billing.Invoice) ([]byte, error) {
	if invoice == nil {
		return nil, fmt.Errorf("invoice is nil")
	}

	// Build the PDF content as a structured document
	doc := g.buildInvoiceDocument(invoice)

	return g.renderToPDF(doc)
}

// InvoicePDFDocument represents the structured PDF content
type InvoicePDFDocument struct {
	Header      PDFHeader      `json:"header"`
	InvoiceInfo PDFInvoiceInfo `json:"invoice_info"`
	Parties     PDFParties     `json:"parties"`
	LineItems   []PDFLineItem  `json:"line_items"`
	Summary     PDFSummary     `json:"summary"`
	TaxDetails  *PDFTaxDetails `json:"tax_details,omitempty"`
	Notes       string         `json:"notes,omitempty"`
	Terms       string         `json:"terms,omitempty"`
	Footer      string         `json:"footer"`
	GeneratedAt time.Time      `json:"generated_at"`
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
	InvoiceNumber string    `json:"invoice_number"`
	InvoiceID     string    `json:"invoice_id"`
	IssueDate     time.Time `json:"issue_date"`
	DueDate       time.Time `json:"due_date"`
	BillingPeriod string    `json:"billing_period"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	EscrowID      string    `json:"escrow_id"`
	OrderID       string    `json:"order_id"`
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
	Subtotal      string        `json:"subtotal"`
	Discounts     []PDFDiscount `json:"discounts,omitempty"`
	DiscountTotal string        `json:"discount_total,omitempty"`
	TaxTotal      string        `json:"tax_total,omitempty"`
	Total         string        `json:"total"`
	AmountPaid    string        `json:"amount_paid"`
	AmountDue     string        `json:"amount_due"`
}

// PDFDiscount represents a discount in the PDF
type PDFDiscount struct {
	Description string `json:"description"`
	Amount      string `json:"amount"`
}

// PDFTaxDetails contains tax breakdown
type PDFTaxDetails struct {
	CustomerTaxID string       `json:"customer_tax_id,omitempty"`
	ProviderTaxID string       `json:"provider_tax_id,omitempty"`
	Jurisdiction  string       `json:"jurisdiction,omitempty"`
	TaxLines      []PDFTaxLine `json:"tax_lines,omitempty"`
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
			UnitPrice:   formatDecCoin(item.UnitPrice),
			Amount:      formatCoins(item.Amount),
		})
	}

	// Build summary
	doc.Summary = PDFSummary{
		Subtotal:   formatCoins(invoice.Subtotal),
		Total:      formatCoins(invoice.Total),
		AmountPaid: formatCoins(invoice.AmountPaid),
		AmountDue:  formatCoins(invoice.AmountDue),
	}

	// Add discounts
	if len(invoice.Discounts) > 0 {
		for _, d := range invoice.Discounts {
			doc.Summary.Discounts = append(doc.Summary.Discounts, PDFDiscount{
				Description: d.Description,
				Amount:      formatCoins(d.Amount),
			})
		}
		doc.Summary.DiscountTotal = formatCoins(invoice.DiscountTotal)
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
				Amount:      formatCoins(taxLine.TaxAmount),
			})
		}
		doc.Summary.TaxTotal = formatCoins(invoice.TaxTotal)
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

// renderToPDF renders the document to PDF format.
func (g *PDFGenerator) renderToPDF(doc *InvoicePDFDocument) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", g.pageSize(), "")
	pdf.SetMargins(15, 20, 15)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	fontFamily := sanitizeFontFamily(g.config.FontFamily)
	primary := parseHexColorOrDefault(g.config.PrimaryColor, 37, 99, 235)
	secondary := parseHexColorOrDefault(g.config.SecondaryColor, 30, 64, 175)

	// Header with optional logo
	headerX := 15.0
	headerY := 15.0
	if g.config.LogoPath != "" {
		if logoPath, ok := resolveLogoPath(g.config.LogoPath); ok {
			pdf.ImageOptions(logoPath, headerX, headerY, 28, 0, false, imageOptionsForPath(logoPath), 0, "")
			headerX = 50
		}
	}

	pdf.SetFont(fontFamily, "B", 14)
	pdf.SetXY(headerX, headerY)
	pdf.SetTextColor(primary.r, primary.g, primary.b)
	pdf.CellFormat(80, 6, doc.Header.CompanyName, "", 1, "L", false, 0, "")

	pdf.SetFont(fontFamily, "", 9)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetX(headerX)
	for _, line := range doc.Header.CompanyAddress {
		pdf.CellFormat(80, 4.5, line, "", 1, "L", false, 0, "")
		pdf.SetX(headerX)
	}
	if doc.Header.CompanyEmail != "" {
		pdf.CellFormat(80, 4.5, doc.Header.CompanyEmail, "", 1, "L", false, 0, "")
	}
	if doc.Header.CompanyPhone != "" {
		pdf.CellFormat(80, 4.5, doc.Header.CompanyPhone, "", 1, "L", false, 0, "")
	}

	// Invoice title and metadata
	rightColX := 120.0
	pdf.SetXY(rightColX, headerY)
	pdf.SetFont(fontFamily, "B", 20)
	pdf.SetTextColor(secondary.r, secondary.g, secondary.b)
	pdf.CellFormat(70, 10, "INVOICE", "", 1, "R", false, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont(fontFamily, "", 9)
	pdf.SetX(rightColX)
	pdf.CellFormat(70, 4.5, fmt.Sprintf("Invoice Number: %s", doc.InvoiceInfo.InvoiceNumber), "", 1, "R", false, 0, "")
	pdf.SetX(rightColX)
	pdf.CellFormat(70, 4.5, fmt.Sprintf("Issue Date: %s", doc.InvoiceInfo.IssueDate.Format("2006-01-02")), "", 1, "R", false, 0, "")
	pdf.SetX(rightColX)
	pdf.CellFormat(70, 4.5, fmt.Sprintf("Due Date: %s", doc.InvoiceInfo.DueDate.Format("2006-01-02")), "", 1, "R", false, 0, "")
	pdf.SetX(rightColX)
	pdf.CellFormat(70, 4.5, fmt.Sprintf("Status: %s", strings.ToUpper(doc.InvoiceInfo.Status)), "", 1, "R", false, 0, "")
	if doc.InvoiceInfo.EscrowID != "" {
		pdf.SetX(rightColX)
		pdf.CellFormat(70, 4.5, fmt.Sprintf("Escrow: %s", doc.InvoiceInfo.EscrowID), "", 1, "R", false, 0, "")
	}
	if doc.InvoiceInfo.OrderID != "" {
		pdf.SetX(rightColX)
		pdf.CellFormat(70, 4.5, fmt.Sprintf("Order: %s", doc.InvoiceInfo.OrderID), "", 1, "R", false, 0, "")
	}

	pdf.SetDrawColor(230, 230, 230)
	pdf.Line(15, 55, 195, 55)
	pdf.SetY(60)

	// Parties section
	pdf.SetFont(fontFamily, "B", 10)
	pdf.CellFormat(90, 5, "Bill To", "", 0, "L", false, 0, "")
	pdf.CellFormat(90, 5, "Provider", "", 1, "L", false, 0, "")

	pdf.SetFont(fontFamily, "", 9)
	startY := pdf.GetY()
	pdf.MultiCell(90, 4.5, doc.Parties.Customer.Address, "", "L", false)
	leftEndY := pdf.GetY()
	pdf.SetXY(105, startY)
	pdf.MultiCell(90, 4.5, doc.Parties.Provider.Address, "", "L", false)
	rightEndY := pdf.GetY()
	if rightEndY < leftEndY {
		rightEndY = leftEndY
	}
	pdf.SetY(rightEndY + 4)

	// Line items table header
	pdf.SetFont(fontFamily, "B", 9)
	pdf.SetFillColor(primary.r, primary.g, primary.b)
	pdf.SetTextColor(255, 255, 255)
	tableHeaders := []string{"Description", "Usage", "Qty", "Unit", "Unit Price", "Amount"}
	colWidths := []float64{60, 22, 20, 18, 30, 30}
	for i, header := range tableHeaders {
		pdf.CellFormat(colWidths[i], 7, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Line items table body
	pdf.SetFont(fontFamily, "", 9)
	pdf.SetTextColor(0, 0, 0)
	for _, item := range doc.LineItems {
		row := []string{
			item.Description,
			item.UsageType,
			item.Quantity,
			item.Unit,
			item.UnitPrice,
			item.Amount,
		}
		drawTableRow(pdf, colWidths, row)
	}

	pdf.Ln(2)

	// Summary section
	pdf.SetFont(fontFamily, "B", 10)
	summaryX := 110.0
	pdf.SetXY(summaryX, pdf.GetY())
	pdf.CellFormat(50, 6, "Subtotal", "", 0, "L", false, 0, "")
	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(35, 6, doc.Summary.Subtotal, "", 1, "R", false, 0, "")

	if doc.Summary.DiscountTotal != "" {
		pdf.SetX(summaryX)
		pdf.SetFont(fontFamily, "B", 10)
		pdf.CellFormat(50, 6, "Discounts", "", 0, "L", false, 0, "")
		pdf.SetFont(fontFamily, "", 10)
		pdf.CellFormat(35, 6, doc.Summary.DiscountTotal, "", 1, "R", false, 0, "")
	}

	if doc.Summary.TaxTotal != "" {
		pdf.SetX(summaryX)
		pdf.SetFont(fontFamily, "B", 10)
		pdf.CellFormat(50, 6, "Tax", "", 0, "L", false, 0, "")
		pdf.SetFont(fontFamily, "", 10)
		pdf.CellFormat(35, 6, doc.Summary.TaxTotal, "", 1, "R", false, 0, "")
	}

	pdf.SetX(summaryX)
	pdf.SetFont(fontFamily, "B", 11)
	pdf.CellFormat(50, 7, "Total", "T", 0, "L", false, 0, "")
	pdf.SetFont(fontFamily, "B", 11)
	pdf.CellFormat(35, 7, doc.Summary.Total, "T", 1, "R", false, 0, "")

	pdf.SetX(summaryX)
	pdf.SetFont(fontFamily, "", 9)
	pdf.CellFormat(50, 5, "Amount Paid", "", 0, "L", false, 0, "")
	pdf.CellFormat(35, 5, doc.Summary.AmountPaid, "", 1, "R", false, 0, "")
	pdf.SetX(summaryX)
	pdf.CellFormat(50, 5, "Amount Due", "", 0, "L", false, 0, "")
	pdf.CellFormat(35, 5, doc.Summary.AmountDue, "", 1, "R", false, 0, "")

	// Tax details
	if doc.TaxDetails != nil && len(doc.TaxDetails.TaxLines) > 0 {
		pdf.Ln(4)
		pdf.SetFont(fontFamily, "B", 10)
		pdf.CellFormat(0, 5, "Tax Details", "", 1, "L", false, 0, "")
		pdf.SetFont(fontFamily, "", 9)
		for _, tax := range doc.TaxDetails.TaxLines {
			line := fmt.Sprintf("%s (%s): %s", tax.Description, tax.Rate, tax.Amount)
			pdf.CellFormat(0, 4.5, line, "", 1, "L", false, 0, "")
		}
	}

	// Notes and terms
	if doc.Notes != "" {
		pdf.Ln(6)
		pdf.SetFont(fontFamily, "B", 9)
		pdf.CellFormat(0, 5, "Payment Instructions", "", 1, "L", false, 0, "")
		pdf.SetFont(fontFamily, "", 8)
		pdf.MultiCell(0, 4, doc.Notes, "", "L", false)
	}
	if doc.Terms != "" {
		pdf.Ln(4)
		pdf.SetFont(fontFamily, "B", 9)
		pdf.CellFormat(0, 5, "Terms", "", 1, "L", false, 0, "")
		pdf.SetFont(fontFamily, "", 8)
		pdf.MultiCell(0, 4, doc.Terms, "", "L", false)
	}

	// Footer
	if doc.Footer != "" {
		pdf.SetY(-15)
		pdf.SetFont(fontFamily, "", 8)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 10, doc.Footer, "", 0, "C", false, 0, "")
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	if pdf.Error() != nil {
		return nil, pdf.Error()
	}
	return buf.Bytes(), nil
}

func (g *PDFGenerator) pageSize() string {
	if g.config.PageSize == "" {
		return "A4"
	}
	return g.config.PageSize
}

type rgbColor struct {
	r int
	g int
	b int
}

func parseHexColorOrDefault(hex string, r, g, b int) rgbColor {
	if parsed, ok := parseHexColor(hex); ok {
		return parsed
	}
	return rgbColor{r: r, g: g, b: b}
}

func parseHexColor(hex string) (rgbColor, bool) {
	if hex == "" {
		return rgbColor{}, false
	}
	value := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(value) != 6 {
		return rgbColor{}, false
	}
	parsed, err := strconv.ParseInt(value, 16, 32)
	if err != nil {
		return rgbColor{}, false
	}
	return rgbColor{
		r: int((parsed >> 16) & 0xff),
		g: int((parsed >> 8) & 0xff),
		b: int(parsed & 0xff),
	}, true
}

func sanitizeFontFamily(font string) string {
	if font == "" {
		return "Helvetica"
	}
	normalized := strings.ToLower(strings.TrimSpace(font))
	allowed := map[string]string{
		"arial":        "Arial",
		"helvetica":    "Helvetica",
		"times":        "Times",
		"courier":      "Courier",
		"symbol":       "Symbol",
		"zapfdingbats": "ZapfDingbats",
	}
	if mapped, ok := allowed[normalized]; ok {
		return mapped
	}
	return "Helvetica"
}

func resolveLogoPath(path string) (string, bool) {
	clean := filepath.Clean(path)
	if _, err := os.Stat(clean); err == nil {
		return clean, true
	}
	return "", false
}

func imageOptionsForPath(path string) gofpdf.ImageOptions {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	return gofpdf.ImageOptions{ImageType: ext}
}

func formatCoins(coins sdk.Coins) string {
	if len(coins) == 0 {
		return "0"
	}
	parts := make([]string, 0, len(coins))
	for _, coin := range coins {
		parts = append(parts, fmt.Sprintf("%s %s", coin.Amount.String(), coin.Denom))
	}
	return strings.Join(parts, ", ")
}

func formatDecCoin(coin sdk.DecCoin) string {
	if coin.Denom == "" {
		return coin.String()
	}
	return fmt.Sprintf("%s %s", coin.Amount.String(), coin.Denom)
}

func drawTableRow(pdf *gofpdf.Fpdf, widths []float64, columns []string) {
	if len(widths) != len(columns) {
		return
	}
	lineHeight := 5.0
	maxLines := 1
	for i, col := range columns {
		lines := pdf.SplitLines([]byte(col), widths[i])
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	rowHeight := float64(maxLines) * lineHeight

	startX, startY := pdf.GetX(), pdf.GetY()
	x := startX
	for i, col := range columns {
		pdf.Rect(x, startY, widths[i], rowHeight, "D")
		pdf.SetXY(x, startY)
		pdf.MultiCell(widths[i], lineHeight, col, "", "L", false)
		x += widths[i]
		pdf.SetXY(x, startY)
	}
	pdf.SetXY(startX, startY+rowHeight)
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
