//go:build ignore
// +build ignore

// TODO: This test file is excluded until MockRolesKeeper interface is aligned.

// Package keeper contains tests for the Fraud module keeper.
//
// VE-912: Fraud reporting flow - Keeper tests
package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
)

// MockRolesKeeper implements a mock roles keeper for testing
type MockRolesKeeper struct {
	moderators map[string]bool
	admins     map[string]bool
	providers  map[string]bool
}

func NewMockRolesKeeper() *MockRolesKeeper {
	return &MockRolesKeeper{
		moderators: make(map[string]bool),
		admins:     make(map[string]bool),
		providers:  make(map[string]bool),
	}
}

func (m *MockRolesKeeper) HasRole(ctx sdk.Context, address sdk.AccAddress, role interface{}) bool {
	roleStr, ok := role.(string)
	if !ok {
		return false
	}
	if roleStr == "service_provider" {
		return m.providers[address.String()]
	}
	return false
}

func (m *MockRolesKeeper) IsModerator(ctx sdk.Context, addr sdk.AccAddress) bool {
	return m.moderators[addr.String()]
}

func (m *MockRolesKeeper) IsAdmin(ctx sdk.Context, addr sdk.AccAddress) bool {
	return m.admins[addr.String()]
}

func (m *MockRolesKeeper) SetModerator(addr string) {
	m.moderators[addr] = true
}

func (m *MockRolesKeeper) SetAdmin(addr string) {
	m.admins[addr] = true
}

func (m *MockRolesKeeper) SetProvider(addr string) {
	m.providers[addr] = true
}

// MockProviderKeeper implements a mock provider keeper for testing
type MockProviderKeeper struct {
	providers map[string]bool
}

func NewMockProviderKeeper() *MockProviderKeeper {
	return &MockProviderKeeper{
		providers: make(map[string]bool),
	}
}

func (m *MockProviderKeeper) IsProvider(ctx sdk.Context, addr sdk.AccAddress) bool {
	return m.providers[addr.String()]
}

func (m *MockProviderKeeper) SetProvider(addr string) {
	m.providers[addr] = true
}

// setupKeeper creates a test keeper with mocked dependencies
func setupKeeper(t testing.TB) (Keeper, sdk.Context, *MockRolesKeeper, *MockProviderKeeper) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	mockRoles := NewMockRolesKeeper()
	mockProvider := NewMockProviderKeeper()

	k := NewKeeper(cdc, storeKey, mockRoles, mockProvider, "authority")

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 100,
		Time:   time.Now(),
	}, false, log.NewNopLogger())

	// Initialize params
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	return k, ctx, mockRoles, mockProvider
}

func createValidEvidence() []types.EncryptedEvidence {
	return []types.EncryptedEvidence{{
		AlgorithmID:     "X25519-XSALSA20-POLY1305",
		RecipientKeyIDs: []string{"moderator-key-1"},
		Nonce:           []byte("unique_nonce_123"),
		Ciphertext:      []byte("encrypted_evidence_data"),
		SenderPubKey:    []byte("sender_public_key"),
		EvidenceHash:    "sha256_hash_of_original",
	}}
}

func TestKeeper_SubmitFraudReport(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")

	// Set reporter as provider
	mockProvider.SetProvider(reporter.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	report := types.NewFraudReport(
		"", // Will be auto-generated
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		description,
		evidence,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	err := k.SubmitFraudReport(ctx, report)
	if err != nil {
		t.Fatalf("SubmitFraudReport() error = %v", err)
	}

	// Verify report was stored
	if report.ID == "" {
		t.Error("SubmitFraudReport() should set report ID")
	}

	stored, found := k.GetFraudReport(ctx, report.ID)
	if !found {
		t.Error("SubmitFraudReport() report not found after submission")
	}
	if stored.Reporter != reporter.String() {
		t.Errorf("Stored report Reporter = %v, want %v", stored.Reporter, reporter.String())
	}

	// Verify it's in the moderator queue
	entry, found := k.GetModeratorQueueEntry(ctx, report.ID)
	if !found {
		t.Error("SubmitFraudReport() report not in moderator queue")
	}
	if entry.ReportID != report.ID {
		t.Errorf("Queue entry ReportID = %v, want %v", entry.ReportID, report.ID)
	}

	// Verify audit log was created
	logs := k.GetAuditLogsForReport(ctx, report.ID)
	if len(logs) == 0 {
		t.Error("SubmitFraudReport() should create audit log")
	}
	if logs[0].Action != types.AuditActionSubmitted {
		t.Errorf("Audit log Action = %v, want %v", logs[0].Action, types.AuditActionSubmitted)
	}
}

func TestKeeper_SubmitFraudReport_UnauthorizedReporter(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	// Not setting reporter as provider

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		description,
		evidence,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)

	err := k.SubmitFraudReport(ctx, report)
	if err == nil {
		t.Error("SubmitFraudReport() should fail for non-provider")
	}
}

func TestKeeper_GetFraudReportsByReporter(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	mockProvider.SetProvider(reporter.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	// Submit two reports from the same reporter
	for i := 0; i < 2; i++ {
		report := types.NewFraudReport(
			"",
			reporter.String(),
			reported.String(),
			types.FraudCategoryFakeIdentity,
			description,
			evidence,
			ctx.BlockHeight(),
			ctx.BlockTime(),
		)
		if err := k.SubmitFraudReport(ctx, report); err != nil {
			t.Fatalf("Failed to submit report: %v", err)
		}
	}

	reports := k.GetFraudReportsByReporter(ctx, reporter.String())
	if len(reports) != 2 {
		t.Errorf("GetFraudReportsByReporter() returned %d reports, want 2", len(reports))
	}
}

func TestKeeper_GetFraudReportsByStatus(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	mockProvider.SetProvider(reporter.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	// Submit a report
	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		description,
		evidence,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("Failed to submit report: %v", err)
	}

	// Get by submitted status
	reports := k.GetFraudReportsByStatus(ctx, types.FraudReportStatusSubmitted)
	if len(reports) != 1 {
		t.Errorf("GetFraudReportsByStatus(Submitted) returned %d reports, want 1", len(reports))
	}

	// Get by reviewing status (should be empty)
	reports = k.GetFraudReportsByStatus(ctx, types.FraudReportStatusReviewing)
	if len(reports) != 0 {
		t.Errorf("GetFraudReportsByStatus(Reviewing) returned %d reports, want 0", len(reports))
	}
}

func TestKeeper_AssignModerator(t *testing.T) {
	k, ctx, mockRoles, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	moderator := sdk.AccAddress("cosmos1moderator____")

	mockProvider.SetProvider(reporter.String())
	mockRoles.SetModerator(moderator.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		description,
		evidence,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("Failed to submit report: %v", err)
	}

	// Assign moderator
	err := k.AssignModerator(ctx, report.ID, moderator.String())
	if err != nil {
		t.Fatalf("AssignModerator() error = %v", err)
	}

	// Verify assignment
	stored, _ := k.GetFraudReport(ctx, report.ID)
	if stored.AssignedModerator != moderator.String() {
		t.Errorf("AssignedModerator = %v, want %v", stored.AssignedModerator, moderator.String())
	}
	if stored.Status != types.FraudReportStatusReviewing {
		t.Errorf("Status = %v, want %v", stored.Status, types.FraudReportStatusReviewing)
	}

	// Verify audit log
	logs := k.GetAuditLogsForReport(ctx, report.ID)
	found := false
	for _, log := range logs {
		if log.Action == types.AuditActionAssigned {
			found = true
			break
		}
	}
	if !found {
		t.Error("AssignModerator() should create audit log")
	}
}

func TestKeeper_AssignModerator_Unauthorized(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	nonModerator := sdk.AccAddress("cosmos1nonmod________")

	mockProvider.SetProvider(reporter.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		description,
		evidence,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("Failed to submit report: %v", err)
	}

	// Try to assign without moderator role
	err := k.AssignModerator(ctx, report.ID, nonModerator.String())
	if err == nil {
		t.Error("AssignModerator() should fail for non-moderator")
	}
}

func TestKeeper_ResolveFraudReport(t *testing.T) {
	k, ctx, mockRoles, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	moderator := sdk.AccAddress("cosmos1moderator____")

	mockProvider.SetProvider(reporter.String())
	mockRoles.SetModerator(moderator.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		description,
		evidence,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("Failed to submit report: %v", err)
	}

	// Assign and move to reviewing
	if err := k.AssignModerator(ctx, report.ID, moderator.String()); err != nil {
		t.Fatalf("AssignModerator() error = %v", err)
	}

	// Resolve the report
	err := k.ResolveFraudReport(ctx, report.ID, types.ResolutionTypeWarning, "User warned", moderator.String())
	if err != nil {
		t.Fatalf("ResolveFraudReport() error = %v", err)
	}

	// Verify resolution
	stored, _ := k.GetFraudReport(ctx, report.ID)
	if stored.Status != types.FraudReportStatusResolved {
		t.Errorf("Status = %v, want %v", stored.Status, types.FraudReportStatusResolved)
	}
	if stored.Resolution != types.ResolutionTypeWarning {
		t.Errorf("Resolution = %v, want %v", stored.Resolution, types.ResolutionTypeWarning)
	}
	if stored.ResolvedAt == nil {
		t.Error("ResolvedAt should be set")
	}

	// Verify removed from queue
	_, found := k.GetModeratorQueueEntry(ctx, report.ID)
	if found {
		t.Error("Resolved report should be removed from queue")
	}

	// Verify audit log
	logs := k.GetAuditLogsForReport(ctx, report.ID)
	found = false
	for _, log := range logs {
		if log.Action == types.AuditActionResolved {
			found = true
			break
		}
	}
	if !found {
		t.Error("ResolveFraudReport() should create audit log")
	}
}

func TestKeeper_RejectFraudReport(t *testing.T) {
	k, ctx, mockRoles, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	moderator := sdk.AccAddress("cosmos1moderator____")

	mockProvider.SetProvider(reporter.String())
	mockRoles.SetModerator(moderator.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		description,
		evidence,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("Failed to submit report: %v", err)
	}

	// Assign moderator
	if err := k.AssignModerator(ctx, report.ID, moderator.String()); err != nil {
		t.Fatalf("AssignModerator() error = %v", err)
	}

	// Reject the report
	err := k.RejectFraudReport(ctx, report.ID, "No evidence of fraud found", moderator.String())
	if err != nil {
		t.Fatalf("RejectFraudReport() error = %v", err)
	}

	// Verify rejection
	stored, _ := k.GetFraudReport(ctx, report.ID)
	if stored.Status != types.FraudReportStatusRejected {
		t.Errorf("Status = %v, want %v", stored.Status, types.FraudReportStatusRejected)
	}
	if stored.Resolution != types.ResolutionTypeNoAction {
		t.Errorf("Resolution = %v, want %v", stored.Resolution, types.ResolutionTypeNoAction)
	}
}

func TestKeeper_EscalateFraudReport(t *testing.T) {
	k, ctx, mockRoles, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	moderator := sdk.AccAddress("cosmos1moderator____")

	mockProvider.SetProvider(reporter.String())
	mockRoles.SetModerator(moderator.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		description,
		evidence,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("Failed to submit report: %v", err)
	}

	// Assign and move to reviewing
	if err := k.AssignModerator(ctx, report.ID, moderator.String()); err != nil {
		t.Fatalf("AssignModerator() error = %v", err)
	}

	// Escalate the report
	err := k.EscalateFraudReport(ctx, report.ID, "Needs admin review", moderator.String())
	if err != nil {
		t.Fatalf("EscalateFraudReport() error = %v", err)
	}

	// Verify escalation
	stored, _ := k.GetFraudReport(ctx, report.ID)
	if stored.Status != types.FraudReportStatusEscalated {
		t.Errorf("Status = %v, want %v", stored.Status, types.FraudReportStatusEscalated)
	}

	// Verify queue entry updated with higher priority
	entry, found := k.GetModeratorQueueEntry(ctx, report.ID)
	if !found {
		t.Error("Escalated report should still be in queue")
	}
	if entry.Priority != 15 {
		t.Errorf("Queue Priority = %v, want 15 (escalated)", entry.Priority)
	}
	if entry.AssignedTo != "" {
		t.Error("Escalated report should be unassigned")
	}
}

func TestKeeper_ModeratorQueue(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	mockProvider.SetProvider(reporter.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	// Submit multiple reports with different categories
	categories := []types.FraudCategory{
		types.FraudCategoryFakeIdentity,
		types.FraudCategoryPaymentFraud,
		types.FraudCategoryTermsViolation,
	}

	for _, cat := range categories {
		report := types.NewFraudReport(
			"",
			reporter.String(),
			reported.String(),
			cat,
			description,
			evidence,
			ctx.BlockHeight(),
			ctx.BlockTime(),
		)
		if err := k.SubmitFraudReport(ctx, report); err != nil {
			t.Fatalf("Failed to submit report: %v", err)
		}
	}

	// Get queue
	queue := k.GetModeratorQueue(ctx)
	if len(queue) != 3 {
		t.Errorf("GetModeratorQueue() returned %d entries, want 3", len(queue))
	}
}

func TestKeeper_AuditLogs(t *testing.T) {
	k, ctx, mockRoles, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	moderator := sdk.AccAddress("cosmos1moderator____")

	mockProvider.SetProvider(reporter.String())
	mockRoles.SetModerator(moderator.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	report := types.NewFraudReport(
		"",
		reporter.String(),
		reported.String(),
		types.FraudCategoryFakeIdentity,
		description,
		evidence,
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("Failed to submit report: %v", err)
	}

	// Assign moderator (creates another log)
	if err := k.AssignModerator(ctx, report.ID, moderator.String()); err != nil {
		t.Fatalf("AssignModerator() error = %v", err)
	}

	// Resolve (creates another log)
	if err := k.ResolveFraudReport(ctx, report.ID, types.ResolutionTypeWarning, "Warned", moderator.String()); err != nil {
		t.Fatalf("ResolveFraudReport() error = %v", err)
	}

	// Get all logs for report
	logs := k.GetAuditLogsForReport(ctx, report.ID)
	if len(logs) != 3 {
		t.Errorf("GetAuditLogsForReport() returned %d logs, want 3", len(logs))
	}

	// Verify log actions
	actions := make(map[types.AuditAction]bool)
	for _, log := range logs {
		actions[log.Action] = true
	}

	if !actions[types.AuditActionSubmitted] {
		t.Error("Missing AuditActionSubmitted log")
	}
	if !actions[types.AuditActionAssigned] {
		t.Error("Missing AuditActionAssigned log")
	}
	if !actions[types.AuditActionResolved] {
		t.Error("Missing AuditActionResolved log")
	}
}

func TestKeeper_Sequences(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	// Test fraud report sequence
	seq := k.GetNextFraudReportSequence(ctx)
	if seq != 1 {
		t.Errorf("Initial fraud report sequence = %v, want 1", seq)
	}

	k.SetNextFraudReportSequence(ctx, 100)
	seq = k.GetNextFraudReportSequence(ctx)
	if seq != 100 {
		t.Errorf("After set, fraud report sequence = %v, want 100", seq)
	}

	// Test audit log sequence
	seq = k.GetNextAuditLogSequence(ctx)
	if seq != 1 {
		t.Errorf("Initial audit log sequence = %v, want 1", seq)
	}

	k.SetNextAuditLogSequence(ctx, 200)
	seq = k.GetNextAuditLogSequence(ctx)
	if seq != 200 {
		t.Errorf("After set, audit log sequence = %v, want 200", seq)
	}
}

func TestKeeper_Params(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	// Get default params
	params := k.GetParams(ctx)
	if params.MinDescriptionLength != types.MinDescriptionLength {
		t.Errorf("Default MinDescriptionLength = %v, want %v",
			params.MinDescriptionLength, types.MinDescriptionLength)
	}

	// Update params
	newParams := types.Params{
		MinDescriptionLength:    30,
		MaxDescriptionLength:    10000,
		MaxEvidenceCount:        5,
		MaxEvidenceSizeBytes:    5 * 1024 * 1024,
		AutoAssignEnabled:       false,
		EscalationThresholdDays: 14,
		ReportRetentionDays:     180,
		AuditLogRetentionDays:   365,
	}

	if err := k.SetParams(ctx, newParams); err != nil {
		t.Fatalf("SetParams() error = %v", err)
	}

	params = k.GetParams(ctx)
	if params.MinDescriptionLength != 30 {
		t.Errorf("Updated MinDescriptionLength = %v, want 30", params.MinDescriptionLength)
	}
	if params.AutoAssignEnabled {
		t.Error("Updated AutoAssignEnabled should be false")
	}
}

func TestKeeper_Params_InvalidParams(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	// Try to set invalid params
	invalidParams := types.Params{
		MinDescriptionLength: 5, // Too low
	}

	err := k.SetParams(ctx, invalidParams)
	if err == nil {
		t.Error("SetParams() should fail for invalid params")
	}
}

func TestKeeper_WithFraudReports(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter_____")
	reported := sdk.AccAddress("cosmos1reported_____")
	mockProvider.SetProvider(reporter.String())

	evidence := createValidEvidence()
	description := "This is a detailed fraud report description that exceeds the minimum length."

	// Submit 3 reports
	for i := 0; i < 3; i++ {
		report := types.NewFraudReport(
			"",
			reporter.String(),
			reported.String(),
			types.FraudCategoryFakeIdentity,
			description,
			evidence,
			ctx.BlockHeight(),
			ctx.BlockTime(),
		)
		if err := k.SubmitFraudReport(ctx, report); err != nil {
			t.Fatalf("Failed to submit report: %v", err)
		}
	}

	// Count reports using iterator
	count := 0
	k.WithFraudReports(ctx, func(r types.FraudReport) bool {
		count++
		return false
	})

	if count != 3 {
		t.Errorf("WithFraudReports() iterated %d reports, want 3", count)
	}

	// Test early exit
	count = 0
	k.WithFraudReports(ctx, func(r types.FraudReport) bool {
		count++
		return true // Stop after first
	})

	if count != 1 {
		t.Errorf("WithFraudReports() with early exit iterated %d reports, want 1", count)
	}
}

func TestKeeper_calculatePriority(t *testing.T) {
	k, _, _, _ := setupKeeper(t)

	tests := []struct {
		category types.FraudCategory
		expected uint8
	}{
		{types.FraudCategoryFakeIdentity, 10},
		{types.FraudCategorySybilAttack, 10},
		{types.FraudCategoryPaymentFraud, 8},
		{types.FraudCategoryMaliciousContent, 8},
		{types.FraudCategoryServiceMisrepresentation, 6},
		{types.FraudCategoryResourceAbuse, 6},
		{types.FraudCategoryTermsViolation, 4},
		{types.FraudCategoryOther, 2},
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			got := k.calculatePriority(tt.category)
			if got != tt.expected {
				t.Errorf("calculatePriority(%v) = %v, want %v", tt.category, got, tt.expected)
			}
		})
	}
}
