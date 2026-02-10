// Copyright 2026 VirtEngine Authors. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the LICENSE file.

package billing

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// exportTestAddr generates a valid test bech32 address from a seed
func exportTestAddr(seed int) string {
	var buffer bytes.Buffer
	buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6")
	buffer.WriteString(string(rune('0' + (seed/100)%10)))
	buffer.WriteString(string(rune('0' + (seed/10)%10)))
	buffer.WriteString(string(rune('0' + seed%10)))
	res, _ := sdk.AccAddressFromHexUnsafe(buffer.String())
	return res.String()
}

type ExportTestSuite struct {
	suite.Suite
	svc *ExportService
}

func TestExportTestSuite(t *testing.T) {
	suite.Run(t, new(ExportTestSuite))
}

func (s *ExportTestSuite) SetupTest() {
	s.svc = NewExportService()
}

// testInvoice creates a valid test invoice
func (s *ExportTestSuite) testInvoice() *billing.Invoice {
	now := time.Now()
	providerAddr := exportTestAddr(100)
	customerAddr := exportTestAddr(101)

	return &billing.Invoice{
		InvoiceID:     "inv-001",
		InvoiceNumber: "VE-00000001",
		EscrowID:      "escrow-001",
		OrderID:       "order-001",
		LeaseID:       "lease-001",
		Provider:      providerAddr,
		Customer:      customerAddr,
		Status:        billing.InvoiceStatusPending,
		Currency:      "uvirt",
		BillingPeriod: billing.BillingPeriod{
			StartTime:       now.Add(-24 * time.Hour),
			EndTime:         now,
			DurationSeconds: 86400,
			PeriodType:      billing.BillingPeriodTypeMonthly,
		},
		LineItems: []billing.LineItem{
			{
				LineItemID:  "li-001",
				Description: "CPU usage for 100 milli-seconds",
				UsageType:   billing.UsageTypeCPU,
				Quantity:    sdkmath.LegacyNewDec(100),
				Unit:        "milli-seconds",
				UnitPrice:   sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
				Amount:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
		},
		Subtotal:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
		DiscountTotal: sdk.NewCoins(),
		TaxTotal:      sdk.NewCoins(),
		Total:         sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
		AmountPaid:    sdk.NewCoins(),
		AmountDue:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
		DueDate:       now.Add(30 * 24 * time.Hour),
		IssuedAt:      now,
		BlockHeight:   100,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// TestExportInvoiceJSON tests JSON export
func (s *ExportTestSuite) TestExportInvoiceJSON() {
	invoice := s.testInvoice()

	data, err := s.svc.ExportInvoiceJSON(context.Background(), invoice)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)

	// Parse and validate
	var doc InvoiceJSONExport
	err = json.Unmarshal(data, &doc)
	s.Require().NoError(err)

	s.Require().Equal("1.0", doc.Version)
	s.Require().Equal("virtengine/invoice/v1", doc.SchemaVersion)
	s.Require().Equal(invoice.InvoiceID, doc.Invoice.InvoiceID)
	s.Require().Equal(invoice.InvoiceNumber, doc.Invoice.InvoiceNumber)
	s.Require().Equal(invoice.Provider, doc.Invoice.Provider)
	s.Require().Equal(invoice.Customer, doc.Invoice.Customer)
}

// TestExportInvoiceCSV tests CSV export
func (s *ExportTestSuite) TestExportInvoiceCSV() {
	invoice := s.testInvoice()

	data, err := s.svc.ExportInvoiceCSV(context.Background(), invoice)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	s.Require().NoError(err)

	// Header + 1 line item
	s.Require().Len(records, 2)

	// Verify header
	s.Require().Equal("invoice_id", records[0][0])
	s.Require().Equal("invoice_number", records[0][1])

	// Verify data
	s.Require().Equal("inv-001", records[1][0])
	s.Require().Equal("VE-00000001", records[1][1])
}

// TestExportInvoicePDF tests PDF export
func (s *ExportTestSuite) TestExportInvoicePDF() {
	invoice := s.testInvoice()

	data, err := s.svc.ExportInvoicePDF(context.Background(), invoice)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)
	// PDF should contain a valid header
	s.Require().True(bytes.HasPrefix(data, []byte("%PDF")))
}

// TestExportNilInvoice tests that nil invoice returns error
func (s *ExportTestSuite) TestExportNilInvoice() {
	_, err := s.svc.ExportInvoiceJSON(context.Background(), nil)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invoice is nil")

	_, err = s.svc.ExportInvoiceCSV(context.Background(), nil)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invoice is nil")

	_, err = s.svc.ExportInvoicePDF(context.Background(), nil)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invoice is nil")
}

// TestExportServiceWithConfig tests custom config
func (s *ExportTestSuite) TestExportServiceWithConfig() {
	config := PDFConfig{
		CompanyName:    "Test Company",
		CompanyAddress: []string{"123 Test St"},
	}

	svc := NewExportServiceWithConfig(config)
	s.Require().NotNil(svc)

	invoice := s.testInvoice()
	data, err := svc.ExportInvoicePDF(context.Background(), invoice)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)
}

// TestExportSettlementCSV tests settlement summary CSV export
func (s *ExportTestSuite) TestExportSettlementCSV() {
	settlements := []SettlementExportRecord{
		{
			SettlementID: "stl-001",
			EscrowID:     "escrow-001",
			Provider:     "provider1",
			Customer:     "customer1",
			InvoiceID:    "inv-001",
			Amount:       sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			Currency:     "uvirt",
			SettledAt:    time.Now(),
			BlockHeight:  100,
			TxHash:       "abc123",
		},
	}

	data, err := s.svc.ExportSettlementSummaryCSV(context.Background(), settlements)
	s.Require().NoError(err)
	s.Require().NotEmpty(data)

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	s.Require().NoError(err)

	// Header + 1 row
	s.Require().Len(records, 2)
	s.Require().Equal("stl-001", records[1][0])
}

// TestGetSupportedFormats tests that supported formats include CSV, JSON, PDF
func (s *ExportTestSuite) TestGetSupportedFormats() {
	formats := s.svc.GetSupportedFormats()
	s.Require().Len(formats, 3)
	s.Require().Contains(formats, billing.ExportFormatCSV)
	s.Require().Contains(formats, billing.ExportFormatJSON)
	s.Require().Contains(formats, ExportFormatPDF)
}

// TestExportMultiLineItemCSV tests CSV export with multiple line items
func (s *ExportTestSuite) TestExportMultiLineItemCSV() {
	invoice := s.testInvoice()

	// Add more line items
	invoice.LineItems = append(invoice.LineItems,
		billing.LineItem{
			LineItemID:  "li-002",
			Description: "Memory usage",
			UsageType:   billing.UsageTypeMemory,
			Quantity:    sdkmath.LegacyNewDec(512),
			Unit:        "byte-seconds",
			UnitPrice:   sdk.NewDecCoin("uvirt", sdkmath.NewInt(5)),
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 2560)),
		},
		billing.LineItem{
			LineItemID:  "li-003",
			Description: "GPU usage",
			UsageType:   billing.UsageTypeGPU,
			Quantity:    sdkmath.LegacyNewDec(60),
			Unit:        "seconds",
			UnitPrice:   sdk.NewDecCoin("uvirt", sdkmath.NewInt(100)),
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 6000)),
		},
	)

	data, err := s.svc.ExportInvoiceCSV(context.Background(), invoice)
	s.Require().NoError(err)

	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	s.Require().NoError(err)

	// Header + 3 line items
	s.Require().Len(records, 4)
	s.Require().Equal("li-001", records[1][5])
	s.Require().Equal("li-002", records[2][5])
	s.Require().Equal("li-003", records[3][5])
}
