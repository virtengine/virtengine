// Package keeper contains tests for the Fraud module keeper.
//
// VE-912: Fraud reporting flow - open reporting (tenant reporters), de-duplication,
// spam control and the response/rebuttal path.
package keeper

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/fraud/types"
)

// mockMarketKeeper is a test double for the market keeper surface used to verify
// that a reporter is actually a party to the order they cite as standing.
type mockMarketKeeper struct {
	orders map[string]struct{ customer, provider string }
}

func newMockMarketKeeper() *mockMarketKeeper {
	return &mockMarketKeeper{orders: make(map[string]struct{ customer, provider string })}
}

func (m *mockMarketKeeper) SetOrder(orderID, customer, provider string) {
	m.orders[orderID] = struct{ customer, provider string }{customer: customer, provider: provider}
}

func (m *mockMarketKeeper) GetOrderByID(_ sdk.Context, orderID string) (interface{}, bool) {
	_, ok := m.orders[orderID]
	if !ok {
		return nil, false
	}
	return struct{}{}, true
}

func (m *mockMarketKeeper) GetOrderCustomer(_ sdk.Context, orderID string) string {
	if o, ok := m.orders[orderID]; ok {
		return o.customer
	}
	return ""
}

func (m *mockMarketKeeper) GetOrderProvider(_ sdk.Context, orderID string) string {
	if o, ok := m.orders[orderID]; ok {
		return o.provider
	}
	return ""
}

// newReport builds a distinct report for the given reporter.
func newReport(ctx sdk.Context, reporter, reported, description string, orders []string, noOrder bool) *types.FraudReport {
	report := types.NewFraudReport(
		"",
		reporter,
		reported,
		types.FraudCategoryFakeIdentity,
		description,
		createValidEvidence(),
		ctx.BlockHeight(),
		ctx.BlockTime(),
	)
	report.RelatedOrderIDs = orders
	report.NoOrderAvailable = noOrder
	return report
}

// TestKeeper_OpenReporting_TenantWithOrderLink proves a non-provider (tenant) can
// open a report when anchored to an order the reporter is a party to.
func TestKeeper_OpenReporting_TenantWithOrderLink(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	tenant := sdk.AccAddress("cosmos1tenant________").String()
	provider := sdk.AccAddress("cosmos1provider______").String()
	market := newMockMarketKeeper()
	market.SetOrder("order-1", tenant, provider)
	k.SetMarketKeeper(market)

	report := newReport(ctx, tenant, provider, testFraudDescription, []string{"order-1"}, false)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("tenant report linked to own order should be accepted, got: %v", err)
	}

	stored, found := k.GetFraudReport(ctx, report.ID)
	if !found {
		t.Fatal("accepted report was not persisted")
	}
	if stored.Reporter != tenant {
		t.Errorf("stored reporter = %q, want %q", stored.Reporter, tenant)
	}
}

// TestKeeper_OpenReporting_TenantNotPartyToOrder proves the order link is actually
// verified: citing someone else's order does not confer standing.
func TestKeeper_OpenReporting_TenantNotPartyToOrder(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	stranger := sdk.AccAddress("cosmos1stranger______").String()
	customer := sdk.AccAddress("cosmos1customer______").String()
	provider := sdk.AccAddress("cosmos1provider______").String()
	market := newMockMarketKeeper()
	market.SetOrder("order-1", customer, provider)
	k.SetMarketKeeper(market)

	report := newReport(ctx, stranger, provider, testFraudDescription, []string{"order-1"}, false)
	err := k.SubmitFraudReport(ctx, report)
	if err == nil {
		t.Fatal("report citing an order the reporter is not party to must be rejected")
	}
	if !errors.Is(err, types.ErrUnauthorizedReporter) {
		t.Errorf("error = %v, want ErrUnauthorizedReporter", err)
	}
}

// TestKeeper_OpenReporting_TenantWithoutStanding proves a non-provider with neither
// an order reference nor an explicit no-order basis is rejected.
func TestKeeper_OpenReporting_TenantWithoutStanding(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	tenant := sdk.AccAddress("cosmos1tenant________").String()
	reported := sdk.AccAddress("cosmos1reported______").String()

	report := newReport(ctx, tenant, reported, testFraudDescription, nil, false)
	err := k.SubmitFraudReport(ctx, report)
	if err == nil {
		t.Fatal("non-provider report with no standing must be rejected")
	}
	if !errors.Is(err, types.ErrMissingOrderReference) {
		t.Errorf("error = %v, want ErrMissingOrderReference", err)
	}
}

// TestKeeper_OpenReporting_NoOrderAvailableBasis proves the escape hatch works and
// still requires justification evidence.
func TestKeeper_OpenReporting_NoOrderAvailableBasis(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	tenant := sdk.AccAddress("cosmos1tenant________").String()
	reported := sdk.AccAddress("cosmos1reported______").String()

	// With evidence: accepted.
	report := newReport(ctx, tenant, reported, testFraudDescription, nil, true)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("no-order-available report with justification evidence should be accepted, got: %v", err)
	}

	// Without evidence: rejected (the basis must be justified).
	reportNoEvidence := newReport(ctx, tenant, reported, testFraudDescription+" second", nil, true)
	reportNoEvidence.Evidence = nil
	if err := k.SubmitFraudReport(ctx, reportNoEvidence); err == nil {
		t.Fatal("no-order-available basis without justification evidence must be rejected")
	}
}

// TestKeeper_Dedup_IdenticalResubmissionRejected proves an identical repeat from the
// same reporter is rejected and does not create a second queue entry.
func TestKeeper_Dedup_IdenticalResubmissionRejected(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter______").String()
	reported := sdk.AccAddress("cosmos1reported______").String()
	mockProvider.SetProvider(reporter)

	first := newReport(ctx, reporter, reported, testFraudDescription, nil, false)
	if err := k.SubmitFraudReport(ctx, first); err != nil {
		t.Fatalf("first submission failed: %v", err)
	}

	second := newReport(ctx, reporter, reported, testFraudDescription, nil, false)
	err := k.SubmitFraudReport(ctx, second)
	if err == nil {
		t.Fatal("identical resubmission must be rejected")
	}
	if !errors.Is(err, types.ErrDuplicateReport) {
		t.Errorf("error = %v, want ErrDuplicateReport", err)
	}

	count := 0
	k.WithFraudReports(ctx, func(types.FraudReport) bool { count++; return false })
	if count != 1 {
		t.Errorf("stored reports = %d, want 1 (dedup must not persist the repeat)", count)
	}
}

// TestKeeper_RateLimit_TripsAfterWindow proves the per-reporter limit is enforced
// deterministically over a block-height window.
func TestKeeper_RateLimit_TripsAfterWindow(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter______").String()
	reported := sdk.AccAddress("cosmos1reported______").String()
	mockProvider.SetProvider(reporter)

	params := types.DefaultParams()
	params.MaxReportsPerWindow = 2
	params.ReportWindowBlocks = 1000
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("SetParams failed: %v", err)
	}

	for i := 0; i < 2; i++ {
		report := newReport(ctx, reporter, reported, fmt.Sprintf("%s limit-%d", testFraudDescription, i), nil, false)
		if err := k.SubmitFraudReport(ctx, report); err != nil {
			t.Fatalf("submission %d within limit failed: %v", i, err)
		}
	}

	over := newReport(ctx, reporter, reported, testFraudDescription+" limit-over", nil, false)
	err := k.SubmitFraudReport(ctx, over)
	if err == nil {
		t.Fatal("submission beyond MaxReportsPerWindow must be rejected")
	}
	if !errors.Is(err, types.ErrReporterRateLimited) {
		t.Errorf("error = %v, want ErrReporterRateLimited", err)
	}

	// A different reporter is unaffected by the first reporter's usage.
	other := sdk.AccAddress("cosmos1other_________").String()
	mockProvider.SetProvider(other)
	if err := k.SubmitFraudReport(ctx, newReport(ctx, other, reported, testFraudDescription, nil, false)); err != nil {
		t.Errorf("rate limit must be per-reporter, other reporter got: %v", err)
	}
}

// TestKeeper_RateLimit_StaleEntriesPruned proves entries older than the window are
// reclaimed instead of accumulating forever.
func TestKeeper_RateLimit_StaleEntriesPruned(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter______").String()
	reported := sdk.AccAddress("cosmos1reported______").String()
	mockProvider.SetProvider(reporter)

	params := types.DefaultParams()
	params.MaxReportsPerWindow = 5
	params.ReportWindowBlocks = 10
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("SetParams failed: %v", err)
	}

	// Seed an activity entry far outside the window.
	staleHeight := uint64(1)
	key := types.GetReporterActivityKey(reporter, int64(staleHeight), "stale-report")
	heightBz := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBz, staleHeight)
	store := ctx.KVStore(k.skey)
	store.Set(key, []byte("stale-report"))

	// A submission at the current height triggers the sweep.
	if err := k.SubmitFraudReport(ctx, newReport(ctx, reporter, reported, testFraudDescription, nil, false)); err != nil {
		t.Fatalf("submission failed: %v", err)
	}

	if store.Get(key) != nil {
		t.Error("activity entry older than the window should have been pruned")
	}
}

// TestKeeper_RateLimit_HighBitHeightKeyPruned proves a crafted activity key
// whose height bytes exceed math.MaxInt64 is reclaimed as stale instead of
// wrapping into a negative height or counting against the window.
func TestKeeper_RateLimit_HighBitHeightKeyPruned(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter______").String()

	params := types.DefaultParams()
	params.MaxReportsPerWindow = 5
	params.ReportWindowBlocks = 10
	if err := k.SetParams(ctx, params); err != nil {
		t.Fatalf("SetParams failed: %v", err)
	}

	// Plant a key with the high height bit set. GetReporterActivityKey cannot
	// produce it because block heights are non-negative, so it simulates a
	// crafted store entry.
	key := types.GetReporterActivityKey(reporter, 0, "crafted-report")
	prefix := types.GetReporterActivityPrefix(reporter)
	rest := key[len(prefix):]
	var bigHeight [8]byte
	binary.BigEndian.PutUint64(bigHeight[:], math.MaxUint64)
	copy(rest[:8], bigHeight[:])
	store := ctx.KVStore(k.skey)
	store.Set(key, []byte("crafted-report"))

	if err := k.checkReporterRateLimit(ctx, reporter); err != nil {
		t.Fatalf("crafted high-bit key must not trip the limiter: %v", err)
	}
	if store.Get(key) != nil {
		t.Error("high-bit activity key should have been pruned as stale")
	}
}

// TestKeeper_Response_BothSidesMayRespondAndOthersMayNot covers the rebuttal path.
func TestKeeper_Response_BothSidesMayRespondAndOthersMayNot(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter______").String()
	reported := sdk.AccAddress("cosmos1reported______").String()
	mockProvider.SetProvider(reporter)

	report := newReport(ctx, reporter, reported, testFraudDescription, nil, false)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	// The reported party files a rebuttal.
	rebuttal := types.NewFraudResponse("", report.ID, reported, types.FraudRespondentRoleReportedParty,
		createValidEvidence(), "sha256-of-rebuttal", ctx.BlockHeight(), ctx.BlockTime())
	if err := k.SubmitFraudResponse(ctx, rebuttal); err != nil {
		t.Fatalf("reported party should be able to respond, got: %v", err)
	}
	if rebuttal.Role != types.FraudRespondentRoleReportedParty {
		t.Errorf("derived role = %v, want reported party", rebuttal.Role)
	}

	// The reporter can also respond.
	followUp := types.NewFraudResponse("", report.ID, reporter, types.FraudRespondentRoleReporter,
		createValidEvidence(), "sha256-of-followup", ctx.BlockHeight(), ctx.BlockTime())
	if err := k.SubmitFraudResponse(ctx, followUp); err != nil {
		t.Fatalf("reporter should be able to respond, got: %v", err)
	}

	// A third party cannot.
	stranger := sdk.AccAddress("cosmos1stranger______").String()
	third := types.NewFraudResponse("", report.ID, stranger, types.FraudRespondentRoleReportedParty,
		createValidEvidence(), "sha256-of-third", ctx.BlockHeight(), ctx.BlockTime())
	err := k.SubmitFraudResponse(ctx, third)
	if err == nil {
		t.Fatal("an address that is neither party must not be able to respond")
	}
	if !errors.Is(err, types.ErrUnauthorizedRespondent) {
		t.Errorf("error = %v, want ErrUnauthorizedRespondent", err)
	}

	// The response count is surfaced on the report and in the queue entry.
	stored, found := k.GetFraudReport(ctx, report.ID)
	if !found {
		t.Fatal("report disappeared")
	}
	if stored.ResponseCount != 2 {
		t.Errorf("ResponseCount = %d, want 2", stored.ResponseCount)
	}
	if entry, ok := k.GetModeratorQueueEntry(ctx, report.ID); ok && entry.ResponseCount != 2 {
		t.Errorf("queue entry ResponseCount = %d, want 2", entry.ResponseCount)
	}
}

// TestKeeper_Response_TerminalReportRejected proves a settled report cannot be
// reopened by a late rebuttal.
func TestKeeper_Response_TerminalReportRejected(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter______").String()
	reported := sdk.AccAddress("cosmos1reported______").String()
	mockProvider.SetProvider(reporter)

	report := newReport(ctx, reporter, reported, testFraudDescription, nil, false)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	stored, _ := k.GetFraudReport(ctx, report.ID)
	stored.Status = types.FraudReportStatusRejected
	if err := k.SetFraudReport(ctx, stored); err != nil {
		t.Fatalf("SetFraudReport failed: %v", err)
	}

	late := types.NewFraudResponse("", report.ID, reported, types.FraudRespondentRoleReportedParty,
		createValidEvidence(), "sha256-late", ctx.BlockHeight(), ctx.BlockTime())
	err := k.SubmitFraudResponse(ctx, late)
	if err == nil {
		t.Fatal("a rejected report must not accept new responses")
	}
	if !errors.Is(err, types.ErrReportNotPending) {
		t.Errorf("error = %v, want ErrReportNotPending", err)
	}
}

// TestKeeper_GetFraudResponses_DeterministicOrder proves the query output order is
// stable and insertion-ordered, which consensus clients depend on.
func TestKeeper_GetFraudResponses_DeterministicOrder(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter______").String()
	reported := sdk.AccAddress("cosmos1reported______").String()
	mockProvider.SetProvider(reporter)

	report := newReport(ctx, reporter, reported, testFraudDescription, nil, false)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	for i := 0; i < 3; i++ {
		r := types.NewFraudResponse("", report.ID, reported, types.FraudRespondentRoleReportedParty,
			createValidEvidence(), fmt.Sprintf("sha256-%d", i), ctx.BlockHeight(), ctx.BlockTime())
		if err := k.SubmitFraudResponse(ctx, r); err != nil {
			t.Fatalf("response %d failed: %v", i, err)
		}
	}

	first := k.GetFraudResponses(ctx, report.ID)
	second := k.GetFraudResponses(ctx, report.ID)
	if len(first) != 3 {
		t.Fatalf("responses = %d, want 3", len(first))
	}
	for i := range first {
		if first[i].ID != second[i].ID {
			t.Fatalf("response order is not deterministic at %d: %q vs %q", i, first[i].ID, second[i].ID)
		}
	}
}

// TestKeeper_Response_UnauthorizedRespondentCountedOnce proves a rejected response
// does not mutate the report's public response counter.
func TestKeeper_Response_RejectedResponseDoesNotCount(t *testing.T) {
	k, ctx, _, mockProvider := setupKeeper(t)

	reporter := sdk.AccAddress("cosmos1reporter______").String()
	reported := sdk.AccAddress("cosmos1reported______").String()
	mockProvider.SetProvider(reporter)

	report := newReport(ctx, reporter, reported, testFraudDescription, nil, false)
	if err := k.SubmitFraudReport(ctx, report); err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	stranger := sdk.AccAddress("cosmos1stranger______").String()
	if err := k.SubmitFraudResponse(ctx, types.NewFraudResponse("", report.ID, stranger,
		types.FraudRespondentRoleReportedParty, createValidEvidence(), "sha256", ctx.BlockHeight(), ctx.BlockTime())); err == nil {
		t.Fatal("expected rejection")
	}

	stored, _ := k.GetFraudReport(ctx, report.ID)
	if stored.ResponseCount != 0 {
		t.Errorf("ResponseCount = %d after a rejected response, want 0", stored.ResponseCount)
	}
}
