# Task 31M: Invoice PDF Generation

**vibe-kanban ID:** `aa8b1cc0-81f6-4e5a-83eb-cb3eb4a91b17`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31M |
| **Title** | feat(billing): Invoice PDF generation |
| **Priority** | P2 |
| **Wave** | 4 |
| **Estimated LOC** | 2000 |
| **Duration** | 2 weeks |
| **Dependencies** | Escrow module, payment adapters |
| **Blocking** | None |

---

## Problem Statement

Enterprise customers and tax authorities require formal invoices for:
- Provider service payments
- Platform fees
- Settlement records
- Tax documentation

Currently, only on-chain records exist with no exportable invoice format.

### Current State Analysis

```
x/escrow/                       ✅ Settlement records exist
pkg/billing/                    ❌ Does not exist
pkg/pdf/                        ❌ No PDF generation
portal/invoices/                ❌ No invoice UI
```

---

## Acceptance Criteria

### AC-1: Invoice Generation
- [ ] PDF invoice template
- [ ] Line item breakdown
- [ ] Tax calculation support (VAT, GST)
- [ ] Multi-currency support
- [ ] Company branding customization

### AC-2: Invoice Data Model
- [ ] Invoice number generation (sequential)
- [ ] Invoice status tracking (draft, sent, paid)
- [ ] Link to escrow settlement records
- [ ] Customer billing details storage
- [ ] Tax ID validation

### AC-3: Invoice Delivery
- [ ] Email delivery with PDF attachment
- [ ] Portal download
- [ ] Bulk invoice generation
- [ ] Invoice history

### AC-4: Provider Invoices
- [ ] Provider earnings statements
- [ ] Payout invoices
- [ ] Commission breakdowns
- [ ] Tax withholding records

---

## Technical Requirements

### Invoice Types

```go
// pkg/billing/types.go

package billing

import (
    "time"
    
    "github.com/shopspring/decimal"
)

type Invoice struct {
    ID              string
    InvoiceNumber   string          // VE-2024-000001
    Type            InvoiceType
    Status          InvoiceStatus
    
    // Parties
    IssuerID        string          // VirtEngine or Provider address
    CustomerID      string          // Customer account address
    CustomerDetails CustomerDetails
    
    // Amounts
    Currency        string
    Subtotal        decimal.Decimal
    TaxAmount       decimal.Decimal
    TaxRate         decimal.Decimal
    TaxID           string
    Total           decimal.Decimal
    
    // Line items
    Items           []InvoiceItem
    
    // Dates
    InvoiceDate     time.Time
    DueDate         time.Time
    PaidDate        *time.Time
    
    // References
    SettlementIDs   []string        // Link to escrow settlements
    PaymentID       string          // Link to payment
    
    // Metadata
    Notes           string
    Terms           string
    PDFPath         string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type InvoiceType string

const (
    InvoiceTypeCustomer  InvoiceType = "customer"   // VE → Customer
    InvoiceTypeProvider  InvoiceType = "provider"   // Provider → VE (earnings)
    InvoiceTypePayout    InvoiceType = "payout"     // VE → Provider (payout)
)

type InvoiceStatus string

const (
    StatusDraft     InvoiceStatus = "draft"
    StatusSent      InvoiceStatus = "sent"
    StatusPaid      InvoiceStatus = "paid"
    StatusCancelled InvoiceStatus = "cancelled"
)

type InvoiceItem struct {
    Description string
    Quantity    decimal.Decimal
    UnitPrice   decimal.Decimal
    Amount      decimal.Decimal
    Period      *BillingPeriod
    LeaseID     string
    OrderID     string
}

type BillingPeriod struct {
    Start time.Time
    End   time.Time
}

type CustomerDetails struct {
    Name        string
    Email       string
    Address     Address
    TaxID       string
    CompanyName string
}

type Address struct {
    Line1      string
    Line2      string
    City       string
    State      string
    PostalCode string
    Country    string
}
```

### Invoice Service

```go
// pkg/billing/invoice_service.go

package billing

import (
    "context"
    "fmt"
    "time"
    
    "github.com/shopspring/decimal"
)

type InvoiceService struct {
    store       InvoiceStore
    pdfGen      PDFGenerator
    emailSender EmailSender
    escrowQuery EscrowQueryClient
}

func (s *InvoiceService) GenerateInvoice(ctx context.Context, req GenerateInvoiceRequest) (*Invoice, error) {
    // Generate invoice number
    invoiceNumber := s.generateInvoiceNumber(ctx, req.Type)
    
    // Get settlements if not provided
    var items []InvoiceItem
    if len(req.SettlementIDs) > 0 {
        items = s.getItemsFromSettlements(ctx, req.SettlementIDs)
    } else {
        items = req.Items
    }
    
    // Calculate totals
    subtotal := decimal.Zero
    for _, item := range items {
        subtotal = subtotal.Add(item.Amount)
    }
    
    taxRate := s.getTaxRate(req.CustomerDetails.Country)
    taxAmount := subtotal.Mul(taxRate)
    total := subtotal.Add(taxAmount)
    
    invoice := &Invoice{
        ID:              generateID(),
        InvoiceNumber:   invoiceNumber,
        Type:            req.Type,
        Status:          StatusDraft,
        IssuerID:        req.IssuerID,
        CustomerID:      req.CustomerID,
        CustomerDetails: req.CustomerDetails,
        Currency:        req.Currency,
        Subtotal:        subtotal,
        TaxAmount:       taxAmount,
        TaxRate:         taxRate,
        Total:           total,
        Items:           items,
        InvoiceDate:     time.Now(),
        DueDate:         time.Now().AddDate(0, 0, 30), // Net 30
        SettlementIDs:   req.SettlementIDs,
        Notes:           req.Notes,
        Terms:           s.getDefaultTerms(),
        CreatedAt:       time.Now(),
    }
    
    // Save invoice
    if err := s.store.Create(ctx, invoice); err != nil {
        return nil, err
    }
    
    return invoice, nil
}

func (s *InvoiceService) GeneratePDF(ctx context.Context, invoiceID string) ([]byte, error) {
    invoice, err := s.store.Get(ctx, invoiceID)
    if err != nil {
        return nil, err
    }
    
    // Get issuer details
    issuer := s.getIssuerDetails(invoice.IssuerID)
    
    pdfData := &PDFInvoiceData{
        Invoice: invoice,
        Issuer:  issuer,
        Logo:    s.getLogo(invoice.IssuerID),
    }
    
    pdf, err := s.pdfGen.Generate(pdfData)
    if err != nil {
        return nil, err
    }
    
    // Store PDF
    pdfPath := fmt.Sprintf("invoices/%s/%s.pdf", invoice.CustomerID, invoice.InvoiceNumber)
    if err := s.storePDF(ctx, pdfPath, pdf); err != nil {
        return nil, err
    }
    
    // Update invoice with PDF path
    invoice.PDFPath = pdfPath
    s.store.Update(ctx, invoice)
    
    return pdf, nil
}

func (s *InvoiceService) SendInvoice(ctx context.Context, invoiceID string) error {
    invoice, err := s.store.Get(ctx, invoiceID)
    if err != nil {
        return err
    }
    
    // Generate PDF if not exists
    if invoice.PDFPath == "" {
        if _, err := s.GeneratePDF(ctx, invoiceID); err != nil {
            return fmt.Errorf("generate PDF: %w", err)
        }
        invoice, _ = s.store.Get(ctx, invoiceID)
    }
    
    // Load PDF
    pdfData, err := s.loadPDF(ctx, invoice.PDFPath)
    if err != nil {
        return err
    }
    
    // Send email
    err = s.emailSender.Send(ctx, EmailRequest{
        To:      invoice.CustomerDetails.Email,
        Subject: fmt.Sprintf("Invoice %s from VirtEngine", invoice.InvoiceNumber),
        Body:    s.getEmailBody(invoice),
        Attachments: []Attachment{
            {
                Filename:    fmt.Sprintf("%s.pdf", invoice.InvoiceNumber),
                ContentType: "application/pdf",
                Data:        pdfData,
            },
        },
    })
    if err != nil {
        return err
    }
    
    // Update status
    invoice.Status = StatusSent
    return s.store.Update(ctx, invoice)
}

func (s *InvoiceService) generateInvoiceNumber(ctx context.Context, itype InvoiceType) string {
    year := time.Now().Year()
    seq := s.store.GetNextSequence(ctx, year)
    
    prefix := "VE"
    switch itype {
    case InvoiceTypeProvider:
        prefix = "PR"
    case InvoiceTypePayout:
        prefix = "PO"
    }
    
    return fmt.Sprintf("%s-%d-%06d", prefix, year, seq)
}
```

### PDF Generator

```go
// pkg/billing/pdf_generator.go

package billing

import (
    "bytes"
    "fmt"
    
    "github.com/jung-kurt/gofpdf"
    "github.com/shopspring/decimal"
)

type PDFGenerator struct {
    fontDir string
}

type PDFInvoiceData struct {
    Invoice *Invoice
    Issuer  *IssuerDetails
    Logo    []byte
}

type IssuerDetails struct {
    Name        string
    Address     Address
    TaxID       string
    Email       string
    Phone       string
    BankDetails *BankDetails
}

type BankDetails struct {
    BankName      string
    AccountName   string
    AccountNumber string
    RoutingNumber string
    IBAN          string
    SWIFT         string
}

func (g *PDFGenerator) Generate(data *PDFInvoiceData) ([]byte, error) {
    pdf := gofpdf.New("P", "mm", "A4", g.fontDir)
    pdf.AddPage()
    
    // Header with logo
    if len(data.Logo) > 0 {
        pdf.RegisterImageOptionsReader("logo", gofpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(data.Logo))
        pdf.ImageOptions("logo", 15, 10, 40, 0, false, gofpdf.ImageOptions{}, 0, "")
    }
    
    // Invoice title
    pdf.SetFont("Arial", "B", 24)
    pdf.SetXY(120, 15)
    pdf.Cell(80, 10, "INVOICE")
    
    // Invoice details
    pdf.SetFont("Arial", "", 10)
    pdf.SetXY(120, 28)
    pdf.Cell(80, 5, fmt.Sprintf("Invoice Number: %s", data.Invoice.InvoiceNumber))
    pdf.SetXY(120, 33)
    pdf.Cell(80, 5, fmt.Sprintf("Date: %s", data.Invoice.InvoiceDate.Format("January 2, 2006")))
    pdf.SetXY(120, 38)
    pdf.Cell(80, 5, fmt.Sprintf("Due Date: %s", data.Invoice.DueDate.Format("January 2, 2006")))
    
    // Issuer details (From)
    pdf.SetFont("Arial", "B", 10)
    pdf.SetXY(15, 50)
    pdf.Cell(80, 5, "From:")
    pdf.SetFont("Arial", "", 10)
    pdf.SetXY(15, 56)
    pdf.MultiCell(80, 5, fmt.Sprintf("%s\n%s\n%s, %s %s\n%s\nTax ID: %s",
        data.Issuer.Name,
        data.Issuer.Address.Line1,
        data.Issuer.Address.City,
        data.Issuer.Address.State,
        data.Issuer.Address.PostalCode,
        data.Issuer.Address.Country,
        data.Issuer.TaxID,
    ), "", "L", false)
    
    // Customer details (To)
    pdf.SetFont("Arial", "B", 10)
    pdf.SetXY(120, 50)
    pdf.Cell(80, 5, "Bill To:")
    pdf.SetFont("Arial", "", 10)
    pdf.SetXY(120, 56)
    customer := data.Invoice.CustomerDetails
    pdf.MultiCell(80, 5, fmt.Sprintf("%s\n%s\n%s\n%s, %s %s\n%s\nTax ID: %s",
        customer.CompanyName,
        customer.Name,
        customer.Address.Line1,
        customer.Address.City,
        customer.Address.State,
        customer.Address.PostalCode,
        customer.Address.Country,
        customer.TaxID,
    ), "", "L", false)
    
    // Line items table
    y := 95.0
    pdf.SetFont("Arial", "B", 10)
    pdf.SetFillColor(240, 240, 240)
    pdf.SetXY(15, y)
    pdf.CellFormat(100, 8, "Description", "1", 0, "L", true, 0, "")
    pdf.CellFormat(25, 8, "Qty", "1", 0, "C", true, 0, "")
    pdf.CellFormat(30, 8, "Unit Price", "1", 0, "R", true, 0, "")
    pdf.CellFormat(30, 8, "Amount", "1", 1, "R", true, 0, "")
    
    pdf.SetFont("Arial", "", 10)
    for _, item := range data.Invoice.Items {
        y += 8
        pdf.SetXY(15, y)
        
        desc := item.Description
        if item.Period != nil {
            desc += fmt.Sprintf("\n(%s - %s)",
                item.Period.Start.Format("Jan 2"),
                item.Period.End.Format("Jan 2, 2006"))
        }
        
        pdf.CellFormat(100, 8, desc, "1", 0, "L", false, 0, "")
        pdf.CellFormat(25, 8, item.Quantity.String(), "1", 0, "C", false, 0, "")
        pdf.CellFormat(30, 8, formatCurrency(item.UnitPrice, data.Invoice.Currency), "1", 0, "R", false, 0, "")
        pdf.CellFormat(30, 8, formatCurrency(item.Amount, data.Invoice.Currency), "1", 1, "R", false, 0, "")
    }
    
    // Totals
    y += 16
    pdf.SetXY(120, y)
    pdf.CellFormat(45, 7, "Subtotal:", "0", 0, "R", false, 0, "")
    pdf.CellFormat(30, 7, formatCurrency(data.Invoice.Subtotal, data.Invoice.Currency), "0", 1, "R", false, 0, "")
    
    if !data.Invoice.TaxAmount.IsZero() {
        y += 7
        pdf.SetXY(120, y)
        pdf.CellFormat(45, 7, fmt.Sprintf("Tax (%s%%):", data.Invoice.TaxRate.Mul(decimal.NewFromInt(100)).String()), "0", 0, "R", false, 0, "")
        pdf.CellFormat(30, 7, formatCurrency(data.Invoice.TaxAmount, data.Invoice.Currency), "0", 1, "R", false, 0, "")
    }
    
    y += 7
    pdf.SetFont("Arial", "B", 12)
    pdf.SetXY(120, y)
    pdf.CellFormat(45, 8, "Total:", "T", 0, "R", false, 0, "")
    pdf.CellFormat(30, 8, formatCurrency(data.Invoice.Total, data.Invoice.Currency), "T", 1, "R", false, 0, "")
    
    // Status badge
    if data.Invoice.Status == StatusPaid {
        pdf.SetFont("Arial", "B", 20)
        pdf.SetTextColor(0, 128, 0)
        pdf.SetXY(15, y-10)
        pdf.Cell(50, 15, "PAID")
        pdf.SetTextColor(0, 0, 0)
    }
    
    // Payment details
    if data.Issuer.BankDetails != nil {
        y += 20
        pdf.SetFont("Arial", "B", 10)
        pdf.SetXY(15, y)
        pdf.Cell(80, 5, "Payment Details:")
        pdf.SetFont("Arial", "", 10)
        y += 6
        pdf.SetXY(15, y)
        pdf.MultiCell(80, 5, fmt.Sprintf("Bank: %s\nAccount: %s\nIBAN: %s\nSWIFT: %s",
            data.Issuer.BankDetails.BankName,
            data.Issuer.BankDetails.AccountNumber,
            data.Issuer.BankDetails.IBAN,
            data.Issuer.BankDetails.SWIFT,
        ), "", "L", false)
    }
    
    // Terms and notes
    if data.Invoice.Notes != "" || data.Invoice.Terms != "" {
        y = 250
        pdf.SetFont("Arial", "", 8)
        pdf.SetXY(15, y)
        pdf.MultiCell(180, 4, data.Invoice.Terms, "", "L", false)
        pdf.SetXY(15, y+20)
        pdf.MultiCell(180, 4, data.Invoice.Notes, "", "L", false)
    }
    
    // Output
    var buf bytes.Buffer
    if err := pdf.Output(&buf); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

func formatCurrency(amount decimal.Decimal, currency string) string {
    symbols := map[string]string{
        "USD": "$",
        "EUR": "€",
        "GBP": "£",
    }
    symbol := symbols[currency]
    if symbol == "" {
        symbol = currency + " "
    }
    return fmt.Sprintf("%s%s", symbol, amount.StringFixed(2))
}
```

### Portal Invoice Components

```tsx
// portal/src/app/billing/invoices/page.tsx

'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useVirtEngine } from '@virtengine/portal';
import { DataTable } from '@/components/ui/DataTable';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Download, Eye, Send } from 'lucide-react';

interface Invoice {
  id: string;
  invoiceNumber: string;
  status: 'draft' | 'sent' | 'paid' | 'cancelled';
  total: string;
  currency: string;
  invoiceDate: string;
  dueDate: string;
}

const columns = [
  {
    header: 'Invoice #',
    accessor: 'invoiceNumber',
  },
  {
    header: 'Date',
    accessor: 'invoiceDate',
    cell: (row: Invoice) => new Date(row.invoiceDate).toLocaleDateString(),
  },
  {
    header: 'Amount',
    accessor: 'total',
    cell: (row: Invoice) => `${getCurrencySymbol(row.currency)}${row.total}`,
  },
  {
    header: 'Status',
    accessor: 'status',
    cell: (row: Invoice) => (
      <Badge variant={getStatusVariant(row.status)}>
        {row.status.charAt(0).toUpperCase() + row.status.slice(1)}
      </Badge>
    ),
  },
  {
    header: 'Due Date',
    accessor: 'dueDate',
    cell: (row: Invoice) => {
      const due = new Date(row.dueDate);
      const isOverdue = row.status !== 'paid' && due < new Date();
      return (
        <span className={isOverdue ? 'text-red-500' : ''}>
          {due.toLocaleDateString()}
        </span>
      );
    },
  },
  {
    header: 'Actions',
    accessor: 'actions',
    cell: (row: Invoice) => (
      <div className="flex gap-2">
        <Button size="sm" variant="ghost" onClick={() => viewInvoice(row.id)}>
          <Eye className="h-4 w-4" />
        </Button>
        <Button size="sm" variant="ghost" onClick={() => downloadInvoice(row.id)}>
          <Download className="h-4 w-4" />
        </Button>
        {row.status === 'draft' && (
          <Button size="sm" variant="ghost" onClick={() => sendInvoice(row.id)}>
            <Send className="h-4 w-4" />
          </Button>
        )}
      </div>
    ),
  },
];

export default function InvoicesPage() {
  const { address } = useVirtEngine();
  const [page, setPage] = useState(1);

  const { data, isLoading } = useQuery({
    queryKey: ['invoices', address, page],
    queryFn: () => fetchInvoices(address, page),
  });

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">Invoices</h1>
      </div>

      <DataTable
        columns={columns}
        data={data?.invoices || []}
        isLoading={isLoading}
        pagination={{
          page,
          totalPages: data?.totalPages || 1,
          onPageChange: setPage,
        }}
      />
    </div>
  );
}

const downloadInvoice = async (id: string) => {
  const res = await fetch(`/api/billing/invoices/${id}/pdf`);
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `invoice-${id}.pdf`;
  a.click();
  URL.revokeObjectURL(url);
};
```

---

## Directory Structure

```
pkg/billing/
├── types.go              # Invoice types
├── invoice_service.go    # Main service
├── pdf_generator.go      # PDF generation
├── tax.go                # Tax calculation
├── templates/
│   └── invoice.html      # Email template
└── store/
    └── postgres.go       # Database storage

portal/src/app/billing/
├── page.tsx              # Billing overview
├── invoices/
│   ├── page.tsx          # Invoice list
│   └── [id]/
│       └── page.tsx      # Invoice detail
└── settings/
    └── page.tsx          # Billing settings
```

---

## Testing Requirements

### Unit Tests
- Invoice number generation
- Tax calculation
- PDF generation

### Integration Tests
- Full invoice lifecycle
- Email delivery
- PDF download

### Visual Tests
- PDF rendering in different locales
- Multi-currency formatting

---

## Security Considerations

1. **Access Control**: Only invoice owner can download
2. **PDF Storage**: Encrypted at rest
3. **Tax ID Validation**: Validate against VIES (EU)
4. **Audit Trail**: Log all invoice operations
