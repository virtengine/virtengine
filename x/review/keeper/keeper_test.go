// Package keeper implements tests for the Review module keeper.
//
// VE-911: Provider public reviews - Comprehensive unit tests
package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/review/types"
)

// mockMarketKeeper is a mock implementation of MarketKeeper
type mockMarketKeeper struct {
	orders map[string]*mockOrder
}

type mockOrder struct {
	orderID         string
	customerAddress string
	providerAddress string
	completed       bool
	completedAt     time.Time
	orderHash       string
}

func newMockMarketKeeper() *mockMarketKeeper {
	return &mockMarketKeeper{
		orders: make(map[string]*mockOrder),
	}
}

func (m *mockMarketKeeper) AddCompletedOrder(orderID, customer, provider string) {
	m.orders[orderID] = &mockOrder{
		orderID:         orderID,
		customerAddress: customer,
		providerAddress: provider,
		completed:       true,
		completedAt:     time.Now().UTC().Add(-24 * time.Hour),
		orderHash:       "order-hash-" + orderID,
	}
}

func (m *mockMarketKeeper) AddIncompleteOrder(orderID, customer, provider string) {
	m.orders[orderID] = &mockOrder{
		orderID:         orderID,
		customerAddress: customer,
		providerAddress: provider,
		completed:       false,
		completedAt:     time.Time{},
		orderHash:       "order-hash-" + orderID,
	}
}

func (m *mockMarketKeeper) GetOrderByID(_ sdk.Context, orderID string) (interface{}, bool) {
	order, ok := m.orders[orderID]
	return order, ok
}

func (m *mockMarketKeeper) IsOrderCompleted(_ sdk.Context, orderID string) bool {
	order, ok := m.orders[orderID]
	if !ok {
		return false
	}
	return order.completed
}

func (m *mockMarketKeeper) GetOrderCustomer(_ sdk.Context, orderID string) string {
	order, ok := m.orders[orderID]
	if !ok {
		return ""
	}
	return order.customerAddress
}

func (m *mockMarketKeeper) GetOrderProvider(_ sdk.Context, orderID string) string {
	order, ok := m.orders[orderID]
	if !ok {
		return ""
	}
	return order.providerAddress
}

func (m *mockMarketKeeper) GetOrderCompletedAt(_ sdk.Context, orderID string) time.Time {
	order, ok := m.orders[orderID]
	if !ok {
		return time.Time{}
	}
	return order.completedAt
}

func (m *mockMarketKeeper) GetOrderHash(_ sdk.Context, orderID string) string {
	order, ok := m.orders[orderID]
	if !ok {
		return ""
	}
	return order.orderHash
}

// mockRolesKeeper is a mock implementation of RolesKeeper
type mockRolesKeeper struct {
	moderators map[string]bool
	admins     map[string]bool
}

func newMockRolesKeeper() *mockRolesKeeper {
	return &mockRolesKeeper{
		moderators: make(map[string]bool),
		admins:     make(map[string]bool),
	}
}

func (m *mockRolesKeeper) IsModerator(_ sdk.Context, addr sdk.AccAddress) bool {
	return m.moderators[addr.String()]
}

func (m *mockRolesKeeper) IsAdmin(_ sdk.Context, addr sdk.AccAddress) bool {
	return m.admins[addr.String()]
}

func (m *mockRolesKeeper) AddModerator(addr string) {
	m.moderators[addr] = true
}

func (m *mockRolesKeeper) AddAdmin(addr string) {
	m.admins[addr] = true
}

// setupKeeper creates a test keeper
func setupKeeper(t *testing.T) (Keeper, sdk.Context, *mockMarketKeeper, *mockRolesKeeper) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	mockMarket := newMockMarketKeeper()
	mockRoles := newMockRolesKeeper()

	keeper := NewKeeper(cdc, storeKey, mockMarket, mockRoles, "authority")

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	// Set default params
	_ = keeper.SetParams(ctx, types.DefaultParams())

	return keeper, ctx, mockMarket, mockRoles
}

// createTestReview creates a valid test review
func createTestReview(t *testing.T, reviewerAddr, providerAddr, orderID string, rating uint8) *types.Review {
	t.Helper()

	orderRef := types.OrderReference{
		OrderID:         orderID,
		CustomerAddress: reviewerAddr,
		ProviderAddress: providerAddr,
		CompletedAt:     time.Now().UTC().Add(-24 * time.Hour),
		OrderHash:       "test-order-hash",
	}

	review, err := types.NewReview(
		types.ReviewID{ProviderAddress: providerAddr, Sequence: 0}, // Sequence will be set by keeper
		reviewerAddr,
		providerAddr,
		orderRef,
		rating,
		"This is a test review with enough characters to pass validation.",
	)
	if err != nil {
		t.Fatalf("failed to create test review: %v", err)
	}

	return review
}

// Test: Submit a valid review
func TestSubmitReview_Success(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"
	orderID := "order-001"

	// Add completed order to mock
	mockMarket.AddCompletedOrder(orderID, reviewerAddr, providerAddr)

	review := createTestReview(t, reviewerAddr, providerAddr, orderID, 5)

	err := k.SubmitReview(ctx, review)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify review was stored
	stored, exists := k.GetReview(ctx, review.ID.String())
	if !exists {
		t.Fatal("expected review to exist")
	}

	if stored.Rating != 5 {
		t.Errorf("expected rating 5, got %d", stored.Rating)
	}

	if stored.ContentHash == "" {
		t.Error("expected content hash to be set")
	}

	// Verify content hash integrity
	if err := stored.VerifyContentHash(); err != nil {
		t.Errorf("content hash verification failed: %v", err)
	}
}

// Test: Rating must be between 1-5
func TestSubmitReview_InvalidRating(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"
	orderID := "order-002"

	mockMarket.AddCompletedOrder(orderID, reviewerAddr, providerAddr)

	testCases := []struct {
		name   string
		rating uint8
	}{
		{"rating 0", 0},
		{"rating 6", 6},
		{"rating 10", 10},
		{"rating 255", 255},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orderRef := types.OrderReference{
				OrderID:         orderID,
				CustomerAddress: reviewerAddr,
				ProviderAddress: providerAddr,
				CompletedAt:     time.Now().UTC().Add(-24 * time.Hour),
				OrderHash:       "test-hash",
			}

			review := &types.Review{
				ReviewerAddress: reviewerAddr,
				ProviderAddress: providerAddr,
				OrderRef:        orderRef,
				Rating:          tc.rating,
				Text:            "This is a test review with enough characters to pass validation.",
				State:           types.ReviewStateActive,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			}
			review.ID = types.ReviewID{ProviderAddress: providerAddr, Sequence: 1}
			review.ContentHash = review.ComputeContentHash()

			err := k.SubmitReview(ctx, review)
			if err == nil {
				t.Error("expected error for invalid rating")
			}
		})
	}
}

// Test: Valid ratings (1-5)
func TestSubmitReview_ValidRatings(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"

	for rating := uint8(1); rating <= 5; rating++ {
		orderID := "order-rating-" + string(rune('0'+rating))
		mockMarket.AddCompletedOrder(orderID, reviewerAddr, providerAddr)

		review := createTestReview(t, reviewerAddr, providerAddr, orderID, rating)

		err := k.SubmitReview(ctx, review)
		if err != nil {
			t.Errorf("expected no error for rating %d, got: %v", rating, err)
		}
	}
}

// Test: Order must be completed
func TestSubmitReview_OrderNotCompleted(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"
	orderID := "order-incomplete"

	// Add incomplete order
	mockMarket.AddIncompleteOrder(orderID, reviewerAddr, providerAddr)

	review := createTestReview(t, reviewerAddr, providerAddr, orderID, 4)

	err := k.SubmitReview(ctx, review)
	if err == nil {
		t.Error("expected error for incomplete order")
	}

	if err != nil && !types.ErrOrderNotCompleted.Is(err) {
		t.Errorf("expected ErrOrderNotCompleted, got: %v", err)
	}
}

// Test: Order must exist
func TestSubmitReview_OrderNotFound(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"
	orderID := "order-nonexistent"

	// Don't add order to mock

	review := createTestReview(t, reviewerAddr, providerAddr, orderID, 4)

	err := k.SubmitReview(ctx, review)
	if err == nil {
		t.Error("expected error for non-existent order")
	}

	if err != nil && !types.ErrOrderNotFound.Is(err) {
		t.Errorf("expected ErrOrderNotFound, got: %v", err)
	}
}

// Test: Reviewer must be order customer
func TestSubmitReview_UnauthorizedReviewer(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	realCustomer := "cosmos1realcustomer12345678901234567890123456"
	fakeReviewer := "cosmos1fakereviewer1234567890123456789012345"
	providerAddr := "cosmos1provider123456789012345678901234567890"
	orderID := "order-wrong-customer"

	// Add order with different customer
	mockMarket.AddCompletedOrder(orderID, realCustomer, providerAddr)

	// Try to review as different address
	review := createTestReview(t, fakeReviewer, providerAddr, orderID, 5)

	err := k.SubmitReview(ctx, review)
	if err == nil {
		t.Error("expected error for unauthorized reviewer")
	}

	if err != nil && !types.ErrUnauthorizedReviewer.Is(err) {
		t.Errorf("expected ErrUnauthorizedReviewer, got: %v", err)
	}
}

// Test: Duplicate review not allowed
func TestSubmitReview_DuplicateReview(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"
	orderID := "order-dup"

	mockMarket.AddCompletedOrder(orderID, reviewerAddr, providerAddr)

	// Submit first review
	review1 := createTestReview(t, reviewerAddr, providerAddr, orderID, 5)
	err := k.SubmitReview(ctx, review1)
	if err != nil {
		t.Fatalf("first review should succeed: %v", err)
	}

	// Try to submit second review for same order
	review2 := createTestReview(t, reviewerAddr, providerAddr, orderID, 3)
	err = k.SubmitReview(ctx, review2)
	if err == nil {
		t.Error("expected error for duplicate review")
	}

	if err != nil && !types.ErrDuplicateReview.Is(err) {
		t.Errorf("expected ErrDuplicateReview, got: %v", err)
	}
}

// Test: Content hash integrity
func TestReview_ContentHashIntegrity(t *testing.T) {
	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"

	orderRef := types.OrderReference{
		OrderID:         "order-hash-test",
		CustomerAddress: reviewerAddr,
		ProviderAddress: providerAddr,
		CompletedAt:     time.Now().UTC(),
		OrderHash:       "test-hash",
	}

	review, err := types.NewReview(
		types.ReviewID{ProviderAddress: providerAddr, Sequence: 1},
		reviewerAddr,
		providerAddr,
		orderRef,
		5,
		"This is a test review with enough characters to pass validation.",
	)
	if err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	// Verify hash is set
	if review.ContentHash == "" {
		t.Error("content hash should be set")
	}

	// Verify hash is valid
	if err := review.VerifyContentHash(); err != nil {
		t.Errorf("content hash verification should pass: %v", err)
	}

	// Modify content and verify hash fails
	originalHash := review.ContentHash
	review.Text = "Modified review text that is different from original"

	if err := review.VerifyContentHash(); err == nil {
		t.Error("content hash verification should fail after modification")
	}

	// Restore and verify
	review.ContentHash = originalHash
	review.Text = "This is a test review with enough characters to pass validation."
	if err := review.VerifyContentHash(); err != nil {
		t.Errorf("content hash verification should pass after restore: %v", err)
	}
}

// Test: Provider aggregation calculation
func TestProviderAggregation(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"

	// Submit multiple reviews
	ratings := []uint8{5, 4, 5, 3, 4}
	for i, rating := range ratings {
		orderID := "order-agg-" + string(rune('a'+i))
		mockMarket.AddCompletedOrder(orderID, reviewerAddr, providerAddr)

		review := createTestReview(t, reviewerAddr, providerAddr, orderID, rating)
		err := k.SubmitReview(ctx, review)
		if err != nil {
			t.Fatalf("failed to submit review %d: %v", i, err)
		}
	}

	// Check aggregation
	agg, exists := k.GetProviderAggregation(ctx, providerAddr)
	if !exists {
		t.Fatal("expected aggregation to exist")
	}

	if agg.TotalReviews != 5 {
		t.Errorf("expected 5 reviews, got %d", agg.TotalReviews)
	}

	// Expected average: (5+4+5+3+4)/5 = 21/5 = 4.20
	expectedAvg := 4.20
	actualAvg := agg.GetAverageRatingFloat()
	if actualAvg != expectedAvg {
		t.Errorf("expected average %.2f, got %.2f", expectedAvg, actualAvg)
	}

	// Check distribution
	if agg.Distribution.FiveStar != 2 {
		t.Errorf("expected 2 five-star reviews, got %d", agg.Distribution.FiveStar)
	}
	if agg.Distribution.FourStar != 2 {
		t.Errorf("expected 2 four-star reviews, got %d", agg.Distribution.FourStar)
	}
	if agg.Distribution.ThreeStar != 1 {
		t.Errorf("expected 1 three-star review, got %d", agg.Distribution.ThreeStar)
	}
}

// Test: Rating distribution
func TestRatingDistribution(t *testing.T) {
	dist := &types.RatingDistribution{}

	// Add ratings
	testCases := []struct {
		rating      uint8
		expectTotal uint64
	}{
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
		{5, 5},
		{5, 6}, // Second 5-star
		{1, 7}, // Second 1-star
	}

	for _, tc := range testCases {
		if err := dist.Add(tc.rating); err != nil {
			t.Errorf("failed to add rating %d: %v", tc.rating, err)
		}
		if dist.Total() != tc.expectTotal {
			t.Errorf("after adding %d, expected total %d, got %d", tc.rating, tc.expectTotal, dist.Total())
		}
	}

	// Verify distribution
	if dist.OneStar != 2 {
		t.Errorf("expected 2 one-star, got %d", dist.OneStar)
	}
	if dist.TwoStar != 1 {
		t.Errorf("expected 1 two-star, got %d", dist.TwoStar)
	}
	if dist.ThreeStar != 1 {
		t.Errorf("expected 1 three-star, got %d", dist.ThreeStar)
	}
	if dist.FourStar != 1 {
		t.Errorf("expected 1 four-star, got %d", dist.FourStar)
	}
	if dist.FiveStar != 2 {
		t.Errorf("expected 2 five-star, got %d", dist.FiveStar)
	}

	// Verify weighted sum: 1*2 + 2*1 + 3*1 + 4*1 + 5*2 = 2 + 2 + 3 + 4 + 10 = 21
	expectedSum := uint64(21)
	if dist.WeightedSum() != expectedSum {
		t.Errorf("expected weighted sum %d, got %d", expectedSum, dist.WeightedSum())
	}

	// Test remove
	if err := dist.Remove(5); err != nil {
		t.Errorf("failed to remove 5-star: %v", err)
	}
	if dist.FiveStar != 1 {
		t.Errorf("expected 1 five-star after remove, got %d", dist.FiveStar)
	}

	// Test invalid rating
	if err := dist.Add(6); err == nil {
		t.Error("expected error for invalid rating 6")
	}
	if err := dist.Add(0); err == nil {
		t.Error("expected error for invalid rating 0")
	}
}

// Test: Review text validation
func TestReviewTextValidation(t *testing.T) {
	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"

	orderRef := types.OrderReference{
		OrderID:         "order-text-test",
		CustomerAddress: reviewerAddr,
		ProviderAddress: providerAddr,
		CompletedAt:     time.Now().UTC(),
		OrderHash:       "test-hash",
	}

	testCases := []struct {
		name      string
		text      string
		expectErr bool
	}{
		{"too short", "Short", true},
		{"minimum length", "Exactly 10", false}, // 10 chars
		{"normal text", "This is a perfectly valid review text with enough characters.", false},
		{"empty text", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := types.NewReview(
				types.ReviewID{ProviderAddress: providerAddr, Sequence: 1},
				reviewerAddr,
				providerAddr,
				orderRef,
				5,
				tc.text,
			)

			if tc.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Test: Get reviews by provider
func TestGetReviewsByProvider(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	provider1 := "cosmos1provider1234567890123456789012345678a"
	provider2 := "cosmos1provider1234567890123456789012345678b"

	// Add reviews for provider1
	for i := 0; i < 3; i++ {
		orderID := "order-p1-" + string(rune('a'+i))
		mockMarket.AddCompletedOrder(orderID, reviewerAddr, provider1)
		review := createTestReview(t, reviewerAddr, provider1, orderID, 5)
		if err := k.SubmitReview(ctx, review); err != nil {
			t.Fatalf("failed to submit review: %v", err)
		}
	}

	// Add reviews for provider2
	for i := 0; i < 2; i++ {
		orderID := "order-p2-" + string(rune('a'+i))
		mockMarket.AddCompletedOrder(orderID, reviewerAddr, provider2)
		review := createTestReview(t, reviewerAddr, provider2, orderID, 4)
		if err := k.SubmitReview(ctx, review); err != nil {
			t.Fatalf("failed to submit review: %v", err)
		}
	}

	// Get reviews for provider1
	reviews1 := k.GetReviewsByProvider(ctx, provider1)
	if len(reviews1) != 3 {
		t.Errorf("expected 3 reviews for provider1, got %d", len(reviews1))
	}

	// Get reviews for provider2
	reviews2 := k.GetReviewsByProvider(ctx, provider2)
	if len(reviews2) != 2 {
		t.Errorf("expected 2 reviews for provider2, got %d", len(reviews2))
	}
}

// Test: Get review by order
func TestGetReviewByOrder(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"
	orderID := "order-lookup"

	mockMarket.AddCompletedOrder(orderID, reviewerAddr, providerAddr)

	review := createTestReview(t, reviewerAddr, providerAddr, orderID, 5)
	if err := k.SubmitReview(ctx, review); err != nil {
		t.Fatalf("failed to submit review: %v", err)
	}

	// Lookup by order ID
	found, exists := k.GetReviewByOrder(ctx, orderID)
	if !exists {
		t.Fatal("expected to find review by order")
	}

	if found.OrderRef.OrderID != orderID {
		t.Errorf("expected order ID %s, got %s", orderID, found.OrderRef.OrderID)
	}

	// Lookup non-existent order
	_, exists = k.GetReviewByOrder(ctx, "nonexistent-order")
	if exists {
		t.Error("should not find review for non-existent order")
	}
}

// Test: Delete review (moderator action)
func TestDeleteReview(t *testing.T) {
	k, ctx, mockMarket, mockRoles := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"
	moderatorAddr := "cosmos1moderator12345678901234567890123456"
	orderID := "order-delete"

	mockMarket.AddCompletedOrder(orderID, reviewerAddr, providerAddr)
	mockRoles.AddModerator(moderatorAddr)

	// Submit review
	review := createTestReview(t, reviewerAddr, providerAddr, orderID, 5)
	if err := k.SubmitReview(ctx, review); err != nil {
		t.Fatalf("failed to submit review: %v", err)
	}

	// Check aggregation before delete
	aggBefore, _ := k.GetProviderAggregation(ctx, providerAddr)
	if aggBefore.TotalReviews != 1 {
		t.Fatalf("expected 1 review before delete, got %d", aggBefore.TotalReviews)
	}

	// Delete review
	err := k.DeleteReview(ctx, review.ID.String(), moderatorAddr, "Policy violation")
	if err != nil {
		t.Fatalf("failed to delete review: %v", err)
	}

	// Verify review is marked as deleted
	deleted, exists := k.GetReview(ctx, review.ID.String())
	if !exists {
		t.Fatal("review should still exist but be marked deleted")
	}
	if deleted.State != types.ReviewStateDeleted {
		t.Errorf("expected deleted state, got %s", deleted.State)
	}
	if deleted.ModerationReason != "Policy violation" {
		t.Errorf("expected moderation reason, got %s", deleted.ModerationReason)
	}

	// Check aggregation after delete
	aggAfter, _ := k.GetProviderAggregation(ctx, providerAddr)
	if aggAfter.TotalReviews != 0 {
		t.Errorf("expected 0 reviews after delete, got %d", aggAfter.TotalReviews)
	}
}

// Test: Review state transitions
func TestReviewStateTransitions(t *testing.T) {
	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"

	orderRef := types.OrderReference{
		OrderID:         "order-state",
		CustomerAddress: reviewerAddr,
		ProviderAddress: providerAddr,
		CompletedAt:     time.Now().UTC(),
		OrderHash:       "test-hash",
	}

	review, _ := types.NewReview(
		types.ReviewID{ProviderAddress: providerAddr, Sequence: 1},
		reviewerAddr,
		providerAddr,
		orderRef,
		5,
		"This is a test review with enough characters to pass validation.",
	)

	// Initial state should be active
	if review.State != types.ReviewStateActive {
		t.Errorf("expected active state, got %s", review.State)
	}

	// Hide the review
	if err := review.Hide("moderator1", "Under investigation"); err != nil {
		t.Errorf("failed to hide review: %v", err)
	}
	if review.State != types.ReviewStateHidden {
		t.Errorf("expected hidden state, got %s", review.State)
	}

	// Restore the review
	if err := review.Restore(); err != nil {
		t.Errorf("failed to restore review: %v", err)
	}
	if review.State != types.ReviewStateActive {
		t.Errorf("expected active state after restore, got %s", review.State)
	}

	// Delete the review
	if err := review.Delete("moderator1", "Spam content"); err != nil {
		t.Errorf("failed to delete review: %v", err)
	}
	if review.State != types.ReviewStateDeleted {
		t.Errorf("expected deleted state, got %s", review.State)
	}

	// Cannot hide deleted review
	if err := review.Hide("moderator1", "Cannot hide"); err == nil {
		t.Error("should not be able to hide deleted review")
	}
}

// Test: Parameters validation
func TestParamsValidation(t *testing.T) {
	testCases := []struct {
		name      string
		params    types.Params
		expectErr bool
	}{
		{
			name:      "default params",
			params:    types.DefaultParams(),
			expectErr: false,
		},
		{
			name: "invalid min text length",
			params: types.Params{
				MinReviewTextLength:   0,
				MaxReviewTextLength:   2000,
				ReviewCooldownSeconds: 86400,
				MaxReviewsPerProvider: 1000,
				RequireCompletedOrder: true,
			},
			expectErr: true,
		},
		{
			name: "min greater than max",
			params: types.Params{
				MinReviewTextLength:   100,
				MaxReviewTextLength:   50,
				ReviewCooldownSeconds: 86400,
				MaxReviewsPerProvider: 1000,
				RequireCompletedOrder: true,
			},
			expectErr: true,
		},
		{
			name: "negative cooldown",
			params: types.Params{
				MinReviewTextLength:   10,
				MaxReviewTextLength:   2000,
				ReviewCooldownSeconds: -1,
				MaxReviewsPerProvider: 1000,
				RequireCompletedOrder: true,
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectErr && err == nil {
				t.Error("expected validation error")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// Test: Genesis state validation
func TestGenesisStateValidation(t *testing.T) {
	// Valid genesis state
	validGenesis := types.DefaultGenesisState()
	if err := validGenesis.Validate(); err != nil {
		t.Errorf("default genesis should be valid: %v", err)
	}

	// Invalid genesis with duplicate review IDs
	invalidGenesis := types.DefaultGenesisState()
	orderRef := types.OrderReference{
		OrderID:         "order-1",
		CustomerAddress: "cosmos1reviewer",
		ProviderAddress: "cosmos1provider",
		CompletedAt:     time.Now().UTC(),
		OrderHash:       "hash",
	}
	review1 := types.Review{
		ID:              types.ReviewID{ProviderAddress: "cosmos1provider", Sequence: 1},
		ReviewerAddress: "cosmos1reviewer",
		ProviderAddress: "cosmos1provider",
		OrderRef:        orderRef,
		Rating:          5,
		Text:            "Test review text with enough characters for validation.",
		State:           types.ReviewStateActive,
		ContentHash:     "hash1",
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	review2 := review1 // Same ID

	invalidGenesis.Reviews = []types.Review{review1, review2}
	if err := invalidGenesis.Validate(); err == nil {
		t.Error("expected error for duplicate review IDs")
	}
}

// Test: Provider aggregation helper methods
func TestProviderAggregationMethods(t *testing.T) {
	agg := types.NewProviderAggregation("cosmos1provider")

	// Add reviews
	ratings := []uint8{5, 4, 3, 4, 5}
	for _, r := range ratings {
		if err := agg.AddReview(r, time.Now().UTC()); err != nil {
			t.Fatalf("failed to add review: %v", err)
		}
	}

	// Check total
	if agg.TotalReviews != 5 {
		t.Errorf("expected 5 reviews, got %d", agg.TotalReviews)
	}

	// Check average: (5+4+3+4+5)/5 = 21/5 = 4.20
	if agg.GetAverageRatingFloat() != 4.20 {
		t.Errorf("expected 4.20 average, got %.2f", agg.GetAverageRatingFloat())
	}

	// Check display
	if agg.GetAverageRatingDisplay() != "4.20" {
		t.Errorf("expected '4.20' display, got %s", agg.GetAverageRatingDisplay())
	}

	// Remove a review
	if err := agg.RemoveReview(5); err != nil {
		t.Fatalf("failed to remove review: %v", err)
	}

	if agg.TotalReviews != 4 {
		t.Errorf("expected 4 reviews after remove, got %d", agg.TotalReviews)
	}

	// New average: (4+3+4+5)/4 = 16/4 = 4.00
	if agg.GetAverageRatingFloat() != 4.00 {
		t.Errorf("expected 4.00 average after remove, got %.2f", agg.GetAverageRatingFloat())
	}
}

// Test: Sequence generation
func TestSequenceGeneration(t *testing.T) {
	k, ctx, _, _ := setupKeeper(t)

	// First sequence should be 1
	seq1 := k.GetNextReviewSequence(ctx)
	if seq1 != 1 {
		t.Errorf("expected first sequence 1, got %d", seq1)
	}

	// Second sequence should be 2
	seq2 := k.GetNextReviewSequence(ctx)
	if seq2 != 2 {
		t.Errorf("expected second sequence 2, got %d", seq2)
	}

	// Third sequence should be 3
	seq3 := k.GetNextReviewSequence(ctx)
	if seq3 != 3 {
		t.Errorf("expected third sequence 3, got %d", seq3)
	}
}

// Test: Reviews slice helper methods
func TestReviewsSliceMethods(t *testing.T) {
	orderRef := types.OrderReference{
		OrderID:         "order-1",
		CustomerAddress: "cosmos1reviewer",
		ProviderAddress: "cosmos1provider",
		CompletedAt:     time.Now().UTC(),
		OrderHash:       "hash",
	}

	reviews := types.Reviews{
		{
			ID:              types.ReviewID{ProviderAddress: "cosmos1provider1", Sequence: 1},
			ReviewerAddress: "cosmos1reviewer1",
			ProviderAddress: "cosmos1provider1",
			OrderRef:        orderRef,
			Rating:          5,
			State:           types.ReviewStateActive,
		},
		{
			ID:              types.ReviewID{ProviderAddress: "cosmos1provider1", Sequence: 2},
			ReviewerAddress: "cosmos1reviewer2",
			ProviderAddress: "cosmos1provider1",
			OrderRef:        orderRef,
			Rating:          4,
			State:           types.ReviewStateHidden,
		},
		{
			ID:              types.ReviewID{ProviderAddress: "cosmos1provider2", Sequence: 1},
			ReviewerAddress: "cosmos1reviewer1",
			ProviderAddress: "cosmos1provider2",
			OrderRef:        orderRef,
			Rating:          3,
			State:           types.ReviewStateActive,
		},
	}

	// Test Active filter
	active := reviews.Active()
	if len(active) != 2 {
		t.Errorf("expected 2 active reviews, got %d", len(active))
	}

	// Test ByProvider filter
	provider1 := reviews.ByProvider("cosmos1provider1")
	if len(provider1) != 2 {
		t.Errorf("expected 2 reviews for provider1, got %d", len(provider1))
	}

	// Test ByReviewer filter
	reviewer1 := reviews.ByReviewer("cosmos1reviewer1")
	if len(reviewer1) != 2 {
		t.Errorf("expected 2 reviews by reviewer1, got %d", len(reviewer1))
	}

	// Test AverageRating (only active reviews)
	avg := reviews.AverageRating()
	// Active reviews: 5 and 3, average = 4.0
	if avg != 4.0 {
		t.Errorf("expected average 4.0, got %.2f", avg)
	}
}

// Test: Top providers by rating
func TestTopProvidersByRating(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"

	// Create providers with different ratings
	providers := []struct {
		addr    string
		ratings []uint8
	}{
		{"cosmos1provider111111111111111111111111111111", []uint8{5, 5, 5}},      // avg 5.00
		{"cosmos1provider222222222222222222222222222222", []uint8{4, 4, 4}},      // avg 4.00
		{"cosmos1provider333333333333333333333333333333", []uint8{3, 3, 3}},      // avg 3.00
		{"cosmos1provider444444444444444444444444444444", []uint8{5, 4, 5, 4}},   // avg 4.50
		{"cosmos1provider555555555555555555555555555555", []uint8{2, 2, 3}},      // avg 2.33
	}

	orderNum := 0
	for _, p := range providers {
		for _, rating := range p.ratings {
			orderID := "order-top-" + string(rune('a'+orderNum))
			orderNum++
			mockMarket.AddCompletedOrder(orderID, reviewerAddr, p.addr)
			review := createTestReview(t, reviewerAddr, p.addr, orderID, rating)
			if err := k.SubmitReview(ctx, review); err != nil {
				t.Fatalf("failed to submit review: %v", err)
			}
		}
	}

	// Get top 3 providers
	top3 := k.GetTopProviders(ctx, 3)
	if len(top3) != 3 {
		t.Fatalf("expected 3 top providers, got %d", len(top3))
	}

	// Verify order: should be provider1 (5.00), provider4 (4.50), provider2 (4.00)
	expectedOrder := []string{
		"cosmos1provider111111111111111111111111111111",
		"cosmos1provider444444444444444444444444444444",
		"cosmos1provider222222222222222222222222222222",
	}

	for i, expected := range expectedOrder {
		if top3[i].ProviderAddress != expected {
			t.Errorf("position %d: expected %s, got %s", i, expected, top3[i].ProviderAddress)
		}
	}
}

// Test: Message validation
func TestMsgSubmitReviewValidation(t *testing.T) {
	testCases := []struct {
		name      string
		msg       *types.MsgSubmitReview
		expectErr bool
	}{
		{
			name: "valid message",
			msg: types.NewMsgSubmitReview(
				"cosmos1reviewer123456789012345678901234567890",
				"order-123",
				"cosmos1provider123456789012345678901234567890",
				5,
				"This is a valid review with enough characters.",
			),
			expectErr: false,
		},
		{
			name: "empty reviewer",
			msg: types.NewMsgSubmitReview(
				"",
				"order-123",
				"cosmos1provider123456789012345678901234567890",
				5,
				"This is a valid review with enough characters.",
			),
			expectErr: true,
		},
		{
			name: "invalid rating",
			msg: types.NewMsgSubmitReview(
				"cosmos1reviewer123456789012345678901234567890",
				"order-123",
				"cosmos1provider123456789012345678901234567890",
				0,
				"This is a valid review with enough characters.",
			),
			expectErr: true,
		},
		{
			name: "text too short",
			msg: types.NewMsgSubmitReview(
				"cosmos1reviewer123456789012345678901234567890",
				"order-123",
				"cosmos1provider123456789012345678901234567890",
				5,
				"Short",
			),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectErr && err == nil {
				t.Error("expected validation error")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

// Test: Iterator functions
func TestIterators(t *testing.T) {
	k, ctx, mockMarket, _ := setupKeeper(t)

	reviewerAddr := "cosmos1reviewer123456789012345678901234567890"
	providerAddr := "cosmos1provider123456789012345678901234567890"

	// Add 5 reviews
	for i := 0; i < 5; i++ {
		orderID := "order-iter-" + string(rune('a'+i))
		mockMarket.AddCompletedOrder(orderID, reviewerAddr, providerAddr)
		review := createTestReview(t, reviewerAddr, providerAddr, orderID, uint8(i%5+1))
		if err := k.SubmitReview(ctx, review); err != nil {
			t.Fatalf("failed to submit review: %v", err)
		}
	}

	// Count reviews using iterator
	reviewCount := 0
	k.WithReviews(ctx, func(r types.Review) bool {
		reviewCount++
		return false
	})

	if reviewCount != 5 {
		t.Errorf("expected 5 reviews, got %d", reviewCount)
	}

	// Count aggregations using iterator
	aggCount := 0
	k.WithProviderAggregations(ctx, func(a types.ProviderAggregation) bool {
		aggCount++
		return false
	})

	if aggCount != 1 {
		t.Errorf("expected 1 aggregation, got %d", aggCount)
	}
}
