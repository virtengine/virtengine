// Package billing provides tests for invoice generation and ledger management.
package billing

import (
	"bytes"
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// testAddress generates a valid test bech32 address from a seed number
func testAddress(seed int) string {
	var buffer bytes.Buffer
	buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6")
	buffer.WriteString(string(rune('0' + (seed/100)%10)))
	buffer.WriteString(string(rune('0' + (seed/10)%10)))
	buffer.WriteString(string(rune('0' + seed%10)))
	res, _ := sdk.AccAddressFromHexUnsafe(buffer.String())
	return res.String()
}

func TestInvoiceLedgerRecord_NewAndValidate(t *testing.T) {
	now := time.Now().UTC()
	blockHeight := int64(12345)

	invoice := createTestInvoice(now)
	artifactCID := "baf1234567890abcdef"

	record, err := NewInvoiceLedgerRecord(invoice, artifactCID, blockHeight, now)
	if err != nil {
		t.Fatalf("NewInvoiceLedgerRecord failed: %v", err)
	}

	if record.InvoiceID != invoice.InvoiceID {
		t.Errorf("expected InvoiceID %s, got %s", invoice.InvoiceID, record.InvoiceID)
	}

	if record.ContentHash == "" {
		t.Error("ContentHash should not be empty")
	}

	if len(record.ContentHash) != 64 {
		t.Errorf("ContentHash should be 64 hex characters, got %d", len(record.ContentHash))
	}

	if record.ArtifactCID != artifactCID {
		t.Errorf("expected ArtifactCID %s, got %s", artifactCID, record.ArtifactCID)
	}

	//nolint:gosec // G115: line items count is bounded by practical invoice limits
	if record.LineItemCount != uint32(len(invoice.LineItems)) {
		t.Errorf("expected LineItemCount %d, got %d", len(invoice.LineItems), record.LineItemCount)
	}

	if err := record.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestInvoiceLedgerRecord_VerifyContentHash(t *testing.T) {
	now := time.Now().UTC()
	blockHeight := int64(12345)

	invoice := createTestInvoice(now)
	artifactCID := "baf1234567890abcdef"

	record, err := NewInvoiceLedgerRecord(invoice, artifactCID, blockHeight, now)
	if err != nil {
		t.Fatalf("NewInvoiceLedgerRecord failed: %v", err)
	}

	// Verify with original invoice
	valid, err := record.VerifyContentHash(invoice)
	if err != nil {
		t.Fatalf("VerifyContentHash failed: %v", err)
	}
	if !valid {
		t.Error("ContentHash should be valid for original invoice")
	}

	// Modify invoice and verify again
	modifiedInvoice := createTestInvoice(now)
	modifiedInvoice.InvoiceNumber = "MODIFIED-001"

	valid, err = record.VerifyContentHash(modifiedInvoice)
	if err != nil {
		t.Fatalf("VerifyContentHash failed: %v", err)
	}
	if valid {
		t.Error("ContentHash should be invalid for modified invoice")
	}
}

func TestInvoiceLedgerEntry_NewAndValidate(t *testing.T) {
	now := time.Now().UTC()
	blockHeight := int64(12345)

	entry := NewInvoiceLedgerEntry(
		"entry-001",
		"invoice-001",
		LedgerEntryTypeCreated,
		InvoiceStatusDraft,
		InvoiceStatusDraft,
		sdk.NewCoins(),
		"invoice created",
		"virtengine1abc...",
		"txhash123",
		ZeroHash, // Previous entry hash (genesis)
		1,        // Sequence number
		blockHeight,
		now,
	)

	if entry.EntryID != "entry-001" {
		t.Errorf("expected EntryID entry-001, got %s", entry.EntryID)
	}

	if entry.EntryType != LedgerEntryTypeCreated {
		t.Errorf("expected EntryType %s, got %s", LedgerEntryTypeCreated, entry.EntryType)
	}

	if entry.SequenceNumber != 1 {
		t.Errorf("expected SequenceNumber 1, got %d", entry.SequenceNumber)
	}

	if entry.PreviousEntryHash != ZeroHash {
		t.Errorf("expected PreviousEntryHash %s, got %s", ZeroHash, entry.PreviousEntryHash)
	}

	if entry.EntryHash == "" {
		t.Error("expected EntryHash to be computed, got empty string")
	}

	if !entry.VerifyHash() {
		t.Error("hash verification failed")
	}

	if err := entry.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestInvoiceStatusTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     InvoiceStatus
		to       InvoiceStatus
		expected bool
	}{
		{"draft to pending", InvoiceStatusDraft, InvoiceStatusPending, true},
		{"draft to cancelled", InvoiceStatusDraft, InvoiceStatusCancelled, true},
		{"draft to paid", InvoiceStatusDraft, InvoiceStatusPaid, false},
		{"pending to paid", InvoiceStatusPending, InvoiceStatusPaid, true},
		{"pending to partially paid", InvoiceStatusPending, InvoiceStatusPartiallyPaid, true},
		{"pending to overdue", InvoiceStatusPending, InvoiceStatusOverdue, true},
		{"pending to disputed", InvoiceStatusPending, InvoiceStatusDisputed, true},
		{"partially paid to paid", InvoiceStatusPartiallyPaid, InvoiceStatusPaid, true},
		{"partially paid to disputed", InvoiceStatusPartiallyPaid, InvoiceStatusDisputed, true},
		{"overdue to paid", InvoiceStatusOverdue, InvoiceStatusPaid, true},
		{"disputed to pending", InvoiceStatusDisputed, InvoiceStatusPending, true},
		{"disputed to refunded", InvoiceStatusDisputed, InvoiceStatusRefunded, true},
		{"paid to refunded", InvoiceStatusPaid, InvoiceStatusRefunded, true},
		{"paid to cancelled", InvoiceStatusPaid, InvoiceStatusCancelled, false},
		{"cancelled to paid", InvoiceStatusCancelled, InvoiceStatusPaid, false},
		{"refunded to paid", InvoiceStatusRefunded, InvoiceStatusPaid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsValidTransition(%s, %s) = %v, want %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestGetValidNextStates(t *testing.T) {
	tests := []struct {
		status   InvoiceStatus
		minCount int
	}{
		{InvoiceStatusDraft, 2},
		{InvoiceStatusPending, 4},
		{InvoiceStatusPartiallyPaid, 3},
		{InvoiceStatusOverdue, 4},
		{InvoiceStatusDisputed, 4},
		{InvoiceStatusPaid, 1},
		{InvoiceStatusCancelled, 0},
		{InvoiceStatusRefunded, 0},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			states := GetValidNextStates(tt.status)
			if len(states) < tt.minCount {
				t.Errorf("GetValidNextStates(%s) returned %d states, want at least %d",
					tt.status, len(states), tt.minCount)
			}
		})
	}
}

func TestInvoiceStatusMachine_MarkIssued(t *testing.T) {
	now := time.Now().UTC()
	blockHeight := int64(12345)

	invoice := createTestInvoice(now)

	sm := NewInvoiceStatusMachine(invoice, blockHeight, now)
	err := sm.MarkIssued("provider-addr")
	if err != nil {
		t.Fatalf("MarkIssued failed: %v", err)
	}

	if invoice.Status != InvoiceStatusPending {
		t.Errorf("expected status %s, got %s", InvoiceStatusPending, invoice.Status)
	}

	entry := sm.GetLedgerEntry()
	if entry == nil {
		t.Fatal("ledger entry should not be nil")
	}

	if entry.EntryType != LedgerEntryTypeIssued {
		t.Errorf("expected entry type %s, got %s", LedgerEntryTypeIssued, entry.EntryType)
	}
}

func TestInvoiceStatusMachine_TransitionWithPayment(t *testing.T) {
	now := time.Now().UTC()
	blockHeight := int64(12345)

	invoice := createTestInvoice(now)
	_ = invoice.Finalize(now)

	sm := NewInvoiceStatusMachine(invoice, blockHeight, now)

	// Record full payment
	err := sm.TransitionWithPayment(invoice.Total, "customer-addr")
	if err != nil {
		t.Fatalf("TransitionWithPayment failed: %v", err)
	}

	if invoice.Status != InvoiceStatusPaid {
		t.Errorf("expected status %s, got %s", InvoiceStatusPaid, invoice.Status)
	}

	entry := sm.GetLedgerEntry()
	if entry == nil {
		t.Fatal("ledger entry should not be nil")
	}

	if entry.EntryType != LedgerEntryTypePayment {
		t.Errorf("expected entry type %s, got %s", LedgerEntryTypePayment, entry.EntryType)
	}
}

func TestMemoryArtifactStore(t *testing.T) {
	store := NewMemoryArtifactStore()
	ctx := context.Background()

	invoiceID := "invoice-001"
	content := []byte(`{"test": "data"}`)
	createdBy := "system"

	// Store artifact
	artifact, err := store.Store(ctx, invoiceID, ArtifactTypeJSON, content, createdBy)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if artifact.CID == "" {
		t.Error("CID should not be empty")
	}

	if artifact.InvoiceID != invoiceID {
		t.Errorf("expected InvoiceID %s, got %s", invoiceID, artifact.InvoiceID)
	}

	if artifact.Size != int64(len(content)) {
		t.Errorf("expected Size %d, got %d", len(content), artifact.Size)
	}

	// Get artifact
	retrievedContent, retrievedArtifact, err := store.Get(ctx, artifact.CID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrievedContent) != string(content) {
		t.Error("retrieved content does not match original")
	}

	if retrievedArtifact.CID != artifact.CID {
		t.Error("retrieved artifact CID does not match")
	}

	// Get by invoice
	artifacts, err := store.GetByInvoice(ctx, invoiceID)
	if err != nil {
		t.Fatalf("GetByInvoice failed: %v", err)
	}

	if len(artifacts) != 1 {
		t.Errorf("expected 1 artifact, got %d", len(artifacts))
	}

	// Verify
	valid, err := store.Verify(ctx, artifact.CID, content)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if !valid {
		t.Error("content verification should succeed")
	}

	// Verify with wrong content
	valid, err = store.Verify(ctx, artifact.CID, []byte("wrong content"))
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if valid {
		t.Error("content verification should fail for wrong content")
	}

	// Delete
	err = store.Delete(ctx, artifact.CID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, _, err = store.Get(ctx, artifact.CID)
	if err == nil {
		t.Error("Get should fail after delete")
	}
}

func TestInvoiceGenerator_GenerateInvoice(t *testing.T) {
	config := DefaultInvoiceGeneratorConfig()
	gen := NewInvoiceGenerator(config)

	now := time.Now().UTC()
	blockHeight := int64(12345)

	req := InvoiceGenerationRequest{
		EscrowID: "escrow-001",
		OrderID:  "order-001",
		LeaseID:  "lease-001",
		Provider: testAddress(100),
		Customer: testAddress(101),
		UsageInputs: []UsageInput{
			{
				UsageRecordID: "usage-001",
				UsageType:     UsageTypeCPU,
				Quantity:      sdkmath.LegacyNewDec(10),
				Unit:          "core-hour",
				UnitPrice:     sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDecWithPrec(100, 0)),
				Description:   "CPU usage for 10 core-hours",
				PeriodStart:   now.Add(-24 * time.Hour),
				PeriodEnd:     now,
			},
			{
				UsageRecordID: "usage-002",
				UsageType:     UsageTypeMemory,
				Quantity:      sdkmath.LegacyNewDec(32),
				Unit:          "gb-hour",
				UnitPrice:     sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDecWithPrec(50, 0)),
				Description:   "Memory usage for 32 GB-hours",
				PeriodStart:   now.Add(-24 * time.Hour),
				PeriodEnd:     now,
			},
		},
		BillingPeriod: BillingPeriod{
			StartTime:       now.Add(-24 * time.Hour),
			EndTime:         now,
			DurationSeconds: 86400,
			PeriodType:      BillingPeriodTypeDaily,
		},
		Currency: "uvirt",
	}

	invoice, err := gen.GenerateInvoice(req, blockHeight, now)
	if err != nil {
		t.Fatalf("GenerateInvoice failed: %v", err)
	}

	if invoice.InvoiceID == "" {
		t.Error("InvoiceID should not be empty")
	}

	if invoice.InvoiceNumber == "" {
		t.Error("InvoiceNumber should not be empty")
	}

	if len(invoice.LineItems) != 2 {
		t.Errorf("expected 2 line items, got %d", len(invoice.LineItems))
	}

	if invoice.Status != InvoiceStatusDraft {
		t.Errorf("expected status %s, got %s", InvoiceStatusDraft, invoice.Status)
	}

	// Verify totals
	expectedSubtotal := int64(10*100 + 32*50) // 1000 + 1600 = 2600
	if !invoice.Subtotal.IsAllGTE(sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(expectedSubtotal)))) {
		t.Errorf("subtotal mismatch, expected at least %d uvirt", expectedSubtotal)
	}

	if err := invoice.Validate(); err != nil {
		t.Errorf("generated invoice validation failed: %v", err)
	}
}

func TestInvoiceGenerator_GenerateHPCInvoice(t *testing.T) {
	config := DefaultInvoiceGeneratorConfig()
	gen := NewInvoiceGenerator(config)

	now := time.Now().UTC()
	blockHeight := int64(12345)

	hpcUsage := HPCUsageInput{
		JobID:           "job-001",
		CPUHours:        sdkmath.LegacyNewDec(100),
		GPUHours:        sdkmath.LegacyNewDec(10),
		MemoryGBHours:   sdkmath.LegacyNewDec(512),
		StorageGBMonths: sdkmath.LegacyNewDec(100),
		NetworkGB:       sdkmath.LegacyNewDec(50),
		PeriodStart:     now.Add(-24 * time.Hour),
		PeriodEnd:       now,
	}

	invoice, err := gen.GenerateHPCInvoice(
		"escrow-001",
		"order-001",
		"lease-001",
		testAddress(100),
		testAddress(101),
		hpcUsage,
		nil, // use default pricing
		blockHeight,
		now,
	)
	if err != nil {
		t.Fatalf("GenerateHPCInvoice failed: %v", err)
	}

	// Should have 5 line items (CPU, GPU, Memory, Storage, Network)
	if len(invoice.LineItems) != 5 {
		t.Errorf("expected 5 line items, got %d", len(invoice.LineItems))
	}

	// Verify job ID in metadata
	if invoice.Metadata["hpc_job_id"] != "job-001" {
		t.Error("hpc_job_id metadata should be set")
	}
}

func TestInvoiceGenerator_DeterministicInvoiceID(t *testing.T) {
	config := DefaultInvoiceGeneratorConfig()
	gen := NewInvoiceGenerator(config)

	now := time.Now().UTC()
	blockHeight := int64(12345)

	req := createTestRequest(now)

	// Generate invoice twice with same inputs
	invoice1, err := gen.GenerateInvoice(req, blockHeight, now)
	if err != nil {
		t.Fatalf("GenerateInvoice failed: %v", err)
	}

	gen2 := NewInvoiceGenerator(config)
	invoice2, err := gen2.GenerateInvoice(req, blockHeight, now)
	if err != nil {
		t.Fatalf("GenerateInvoice failed: %v", err)
	}

	// Invoice IDs should be the same (deterministic)
	if invoice1.InvoiceID != invoice2.InvoiceID {
		t.Errorf("invoice IDs should be deterministic: %s != %s",
			invoice1.InvoiceID, invoice2.InvoiceID)
	}
}

func TestReconciliationReport_Validate(t *testing.T) {
	now := time.Now().UTC()

	report := &ReconciliationReport{
		ReportID:    "report-001",
		ReportType:  ReconciliationReportTypeDaily,
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		Status:      ReconciliationStatusComplete,
		Summary: ReconciliationSummary{
			TotalInvoices:      10,
			TotalInvoiceAmount: sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000000))),
		},
		GeneratedAt: now,
		GeneratedBy: "system",
		BlockHeight: 12345,
	}

	if err := report.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	}

	// Test invalid: period end before start
	invalidReport := *report
	invalidReport.PeriodStart = now
	invalidReport.PeriodEnd = now.Add(-24 * time.Hour)

	if err := invalidReport.Validate(); err == nil {
		t.Error("should fail validation when period_end before period_start")
	}
}

func TestComputeInvoiceHash(t *testing.T) {
	now := time.Now().UTC()
	invoice := createTestInvoice(now)

	hash1, err := ComputeInvoiceHash(invoice)
	if err != nil {
		t.Fatalf("ComputeInvoiceHash failed: %v", err)
	}

	if hash1 == "" {
		t.Error("hash should not be empty")
	}

	if len(hash1) != 64 {
		t.Errorf("hash should be 64 hex characters, got %d", len(hash1))
	}

	// Same invoice should produce same hash
	hash2, err := ComputeInvoiceHash(invoice)
	if err != nil {
		t.Fatalf("ComputeInvoiceHash failed: %v", err)
	}

	if hash1 != hash2 {
		t.Error("same invoice should produce same hash")
	}

	// Modified invoice should produce different hash
	modifiedInvoice := createTestInvoice(now)
	modifiedInvoice.InvoiceNumber = "MODIFIED-001"

	hash3, err := ComputeInvoiceHash(modifiedInvoice)
	if err != nil {
		t.Fatalf("ComputeInvoiceHash failed: %v", err)
	}

	if hash1 == hash3 {
		t.Error("modified invoice should produce different hash")
	}
}

// Helper functions

func createTestInvoice(now time.Time) *Invoice {
	provider := testAddress(100)
	customer := testAddress(101)
	dueDate := now.Add(7 * 24 * time.Hour)
	invoice := NewInvoice(
		"invoice-001",
		"VE-INV-00000001",
		"escrow-001",
		"order-001",
		"lease-001",
		provider,
		customer,
		"uvirt",
		BillingPeriod{
			StartTime:       now.Add(-24 * time.Hour),
			EndTime:         now,
			DurationSeconds: 86400,
			PeriodType:      BillingPeriodTypeDaily,
		},
		dueDate,
		12345,
		now,
	)

	// Add line items
	invoice.AddLineItem(LineItem{
		LineItemID:     "li-1",
		Description:    "CPU Usage",
		UsageType:      UsageTypeCPU,
		Quantity:       sdkmath.LegacyNewDec(10),
		Unit:           "core-hour",
		UnitPrice:      sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDec(100)),
		Amount:         sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1000))),
		UsageRecordIDs: []string{"usage-001"},
	})

	invoice.AddLineItem(LineItem{
		LineItemID:     "li-2",
		Description:    "Memory Usage",
		UsageType:      UsageTypeMemory,
		Quantity:       sdkmath.LegacyNewDec(32),
		Unit:           "gb-hour",
		UnitPrice:      sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDec(50)),
		Amount:         sdk.NewCoins(sdk.NewCoin("uvirt", sdkmath.NewInt(1600))),
		UsageRecordIDs: []string{"usage-002"},
	})

	return invoice
}

func createTestRequest(now time.Time) InvoiceGenerationRequest {
	return InvoiceGenerationRequest{
		EscrowID: "escrow-001",
		OrderID:  "order-001",
		LeaseID:  "lease-001",
		Provider: testAddress(100),
		Customer: testAddress(101),
		UsageInputs: []UsageInput{
			{
				UsageRecordID: "usage-001",
				UsageType:     UsageTypeCPU,
				Quantity:      sdkmath.LegacyNewDec(10),
				Unit:          "core-hour",
				UnitPrice:     sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDecWithPrec(100, 0)),
				Description:   "CPU usage",
				PeriodStart:   now.Add(-24 * time.Hour),
				PeriodEnd:     now,
			},
		},
		BillingPeriod: BillingPeriod{
			StartTime:       now.Add(-24 * time.Hour),
			EndTime:         now,
			DurationSeconds: 86400,
			PeriodType:      BillingPeriodTypeDaily,
		},
		Currency: "uvirt",
	}
}
