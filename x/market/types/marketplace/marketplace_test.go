// Package marketplace provides tests for the marketplace on-chain module.
//
// VE-300 to VE-304: Marketplace on-chain module tests
package marketplace

import (
	"testing"
	"time"
)

// ============================================================================
// Offering Tests (VE-300)
// ============================================================================

func TestOfferingID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      OfferingID
		wantErr bool
	}{
		{
			name: "valid offering ID",
			id: OfferingID{
				ProviderAddress: "cosmos1abc123",
				Sequence:        1,
			},
			wantErr: false,
		},
		{
			name: "empty provider address",
			id: OfferingID{
				ProviderAddress: "",
				Sequence:        1,
			},
			wantErr: true,
		},
		{
			name: "zero sequence",
			id: OfferingID{
				ProviderAddress: "cosmos1abc123",
				Sequence:        0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("OfferingID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOfferingState_IsAcceptingOrders(t *testing.T) {
	tests := []struct {
		state    OfferingState
		expected bool
	}{
		{OfferingStateActive, true},
		{OfferingStatePaused, false},
		{OfferingStateSuspended, false},
		{OfferingStateDeprecated, false},
		{OfferingStateTerminated, false},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if got := tt.state.IsAcceptingOrders(); got != tt.expected {
				t.Errorf("OfferingState.IsAcceptingOrders() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOffering_Validate(t *testing.T) {
	validPricing := PricingInfo{
		Model:    PricingModelHourly,
		Currency: "uve",
	}

	tests := []struct {
		name     string
		offering Offering
		wantErr  bool
	}{
		{
			name: "valid offering",
			offering: Offering{
				ID: OfferingID{
					ProviderAddress: "cosmos1abc123",
					Sequence:        1,
				},
				State:               OfferingStateActive,
				Category:            OfferingCategoryCompute,
				Name:                "Test Offering",
				Pricing:             validPricing,
				IdentityRequirement: DefaultIdentityRequirement(),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			offering: Offering{
				ID: OfferingID{
					ProviderAddress: "cosmos1abc123",
					Sequence:        1,
				},
				State:    OfferingStateActive,
				Category: OfferingCategoryCompute,
				Name:     "",
				Pricing:  validPricing,
			},
			wantErr: true,
		},
		{
			name: "missing category",
			offering: Offering{
				ID: OfferingID{
					ProviderAddress: "cosmos1abc123",
					Sequence:        1,
				},
				State:   OfferingStateActive,
				Name:    "Test",
				Pricing: validPricing,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.offering.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Offering.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOffering_CanAcceptOrder(t *testing.T) {
	tests := []struct {
		name     string
		offering Offering
		wantErr  bool
	}{
		{
			name: "can accept order",
			offering: Offering{
				State:               OfferingStateActive,
				MaxConcurrentOrders: 10,
				ActiveOrderCount:    5,
			},
			wantErr: false,
		},
		{
			name: "offering paused",
			offering: Offering{
				State: OfferingStatePaused,
			},
			wantErr: true,
		},
		{
			name: "max orders reached",
			offering: Offering{
				State:               OfferingStateActive,
				MaxConcurrentOrders: 5,
				ActiveOrderCount:    5,
			},
			wantErr: true,
		},
		{
			name: "unlimited orders",
			offering: Offering{
				State:               OfferingStateActive,
				MaxConcurrentOrders: 0, // unlimited
				ActiveOrderCount:    1000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.offering.CanAcceptOrder()
			if (err != nil) != tt.wantErr {
				t.Errorf("Offering.CanAcceptOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// Order Tests (VE-300)
// ============================================================================

func TestOrderID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      OrderID
		wantErr bool
	}{
		{
			name: "valid order ID",
			id: OrderID{
				CustomerAddress: "cosmos1xyz789",
				Sequence:        1,
			},
			wantErr: false,
		},
		{
			name: "empty customer address",
			id: OrderID{
				CustomerAddress: "",
				Sequence:        1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("OrderID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrderState_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from     OrderState
		to       OrderState
		expected bool
	}{
		{OrderStatePendingPayment, OrderStateOpen, true},
		{OrderStatePendingPayment, OrderStateCancelled, true},
		{OrderStatePendingPayment, OrderStateActive, false},
		{OrderStateOpen, OrderStateMatched, true},
		{OrderStateOpen, OrderStateCancelled, true},
		{OrderStateOpen, OrderStateActive, false},
		{OrderStateMatched, OrderStateProvisioning, true},
		{OrderStateActive, OrderStatePendingTermination, true},
		{OrderStateTerminated, OrderStateActive, false},
	}

	for _, tt := range tests {
		t.Run(tt.from.String()+"->"+tt.to.String(), func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.expected {
				t.Errorf("OrderState.CanTransitionTo() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOrder_SetState(t *testing.T) {
	order := NewOrder(
		OrderID{CustomerAddress: "cosmos1abc", Sequence: 1},
		OfferingID{ProviderAddress: "cosmos1xyz", Sequence: 1},
		1000,
		1,
	)
	order.State = OrderStateOpen
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Valid transition
	err := order.SetStateAt(OrderStateMatched, "bid accepted", now)
	if err != nil {
		t.Errorf("Order.SetState() unexpected error: %v", err)
	}
	if order.State != OrderStateMatched {
		t.Errorf("Order.SetState() state = %v, want %v", order.State, OrderStateMatched)
	}
	if order.MatchedAt == nil {
		t.Error("Order.SetState() should set MatchedAt")
	}

	// Invalid transition
	err = order.SetStateAt(OrderStatePendingPayment, "invalid", now)
	if err == nil {
		t.Error("Order.SetState() expected error for invalid transition")
	}
}

func TestOrder_CanAcceptBid(t *testing.T) {
	now := time.Now()
	expired := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name    string
		order   Order
		wantErr bool
	}{
		{
			name: "open order can accept bid",
			order: Order{
				State:     OrderStateOpen,
				ExpiresAt: &future,
			},
			wantErr: false,
		},
		{
			name: "matched order cannot accept bid",
			order: Order{
				State: OrderStateMatched,
			},
			wantErr: true,
		},
		{
			name: "expired order cannot accept bid",
			order: Order{
				State:     OrderStateOpen,
				ExpiresAt: &expired,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.CanAcceptBid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Order.CanAcceptBid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// Allocation Tests (VE-300)
// ============================================================================

func TestAllocationState_IsTerminal(t *testing.T) {
	tests := []struct {
		state    AllocationState
		expected bool
	}{
		{AllocationStatePending, false},
		{AllocationStateActive, false},
		{AllocationStateTerminated, true},
		{AllocationStateFailed, true},
		{AllocationStateRejected, true},
	}

	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			if got := tt.state.IsTerminal(); got != tt.expected {
				t.Errorf("AllocationState.IsTerminal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAllocation_Validate(t *testing.T) {
	validAllocation := Allocation{
		ID: AllocationID{
			OrderID: OrderID{
				CustomerAddress: "cosmos1abc",
				Sequence:        1,
			},
			Sequence: 1,
		},
		OfferingID: OfferingID{
			ProviderAddress: "cosmos1xyz",
			Sequence:        1,
		},
		ProviderAddress: "cosmos1xyz",
		BidID: BidID{
			OrderID: OrderID{
				CustomerAddress: "cosmos1abc",
				Sequence:        1,
			},
			ProviderAddress: "cosmos1xyz",
			Sequence:        1,
		},
		State: AllocationStatePending,
	}

	err := validAllocation.Validate()
	if err != nil {
		t.Errorf("Allocation.Validate() unexpected error: %v", err)
	}

	// Test missing provider address
	invalidAllocation := validAllocation
	invalidAllocation.ProviderAddress = ""
	err = invalidAllocation.Validate()
	if err == nil {
		t.Error("Allocation.Validate() expected error for missing provider")
	}
}

// ============================================================================
// Identity Gating Tests (VE-301)
// ============================================================================

func TestIdentityRequirement_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     IdentityRequirement
		wantErr bool
	}{
		{
			name: "valid requirement",
			req: IdentityRequirement{
				MinScore: 50,
			},
			wantErr: false,
		},
		{
			name: "score exceeds 100",
			req: IdentityRequirement{
				MinScore: 101,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("IdentityRequirement.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIdentityRequirement_IsSatisfiedBy(t *testing.T) {
	tests := []struct {
		name           string
		req            IdentityRequirement
		score          uint32
		status         string
		emailVerified  bool
		domainVerified bool
		mfaEnabled     bool
		expected       bool
	}{
		{
			name: "all requirements met",
			req: IdentityRequirement{
				MinScore:             70,
				RequiredStatus:       "verified",
				RequireVerifiedEmail: true,
				RequireMFA:           true,
			},
			score:          80,
			status:         "verified",
			emailVerified:  true,
			domainVerified: false,
			mfaEnabled:     true,
			expected:       true,
		},
		{
			name: "score too low",
			req: IdentityRequirement{
				MinScore: 70,
			},
			score:    50,
			expected: false,
		},
		{
			name: "status mismatch",
			req: IdentityRequirement{
				RequiredStatus: "verified",
			},
			score:    100,
			status:   "pending",
			expected: false,
		},
		{
			name: "email not verified",
			req: IdentityRequirement{
				RequireVerifiedEmail: true,
			},
			score:         100,
			emailVerified: false,
			expected:      false,
		},
		{
			name: "MFA not enabled",
			req: IdentityRequirement{
				RequireMFA: true,
			},
			score:      100,
			mfaEnabled: false,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.req.IsSatisfiedBy(tt.score, tt.status, tt.emailVerified, tt.domainVerified, tt.mfaEnabled)
			if got != tt.expected {
				t.Errorf("IdentityRequirement.IsSatisfiedBy() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIdentityGatingChecker_Check(t *testing.T) {
	offering := &Offering{
		ID: OfferingID{
			ProviderAddress: "cosmos1xyz",
			Sequence:        1,
		},
		IdentityRequirement: IdentityRequirement{
			MinScore:             70,
			RequireVerifiedEmail: true,
		},
		RequireMFAForOrders: true,
	}

	tests := []struct {
		name         string
		customerInfo *CustomerIdentityInfo
		hasErrors    bool
		errorCount   int
	}{
		{
			name: "all checks pass",
			customerInfo: &CustomerIdentityInfo{
				Score:         80,
				Status:        "verified",
				EmailVerified: true,
				MFAEnabled:    true,
			},
			hasErrors:  false,
			errorCount: 0,
		},
		{
			name: "score too low",
			customerInfo: &CustomerIdentityInfo{
				Score:         50,
				Status:        "verified",
				EmailVerified: true,
				MFAEnabled:    true,
			},
			hasErrors:  true,
			errorCount: 1,
		},
		{
			name: "multiple failures",
			customerInfo: &CustomerIdentityInfo{
				Score:         50,
				Status:        "pending",
				EmailVerified: false,
				MFAEnabled:    false,
			},
			hasErrors:  true,
			errorCount: 3, // score, email, mfa
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewIdentityGatingChecker(offering, tt.customerInfo)
			result := checker.Check()

			if result.HasErrors() != tt.hasErrors {
				t.Errorf("IdentityGatingChecker.Check() HasErrors = %v, want %v", result.HasErrors(), tt.hasErrors)
			}
			if len(result.Reasons) != tt.errorCount {
				t.Errorf("IdentityGatingChecker.Check() error count = %d, want %d", len(result.Reasons), tt.errorCount)
			}
		})
	}
}

func TestValidateOrderCreation(t *testing.T) {
	offering := &Offering{
		ID: OfferingID{
			ProviderAddress: "cosmos1xyz",
			Sequence:        1,
		},
		IdentityRequirement: IdentityRequirement{
			MinScore: 50,
		},
	}

	validCustomer := &CustomerIdentityInfo{
		Score:  60,
		Status: "verified",
	}

	invalidCustomer := &CustomerIdentityInfo{
		Score:  30,
		Status: "pending",
	}

	// Valid case
	err := ValidateOrderCreation(offering, validCustomer, nil)
	if err != nil {
		t.Errorf("ValidateOrderCreation() unexpected error: %v", err)
	}

	// Invalid case
	err = ValidateOrderCreation(offering, invalidCustomer, nil)
	if err == nil {
		t.Error("ValidateOrderCreation() expected error for invalid customer")
	}
}

// ============================================================================
// MFA Gating Tests (VE-302)
// ============================================================================

func TestMarketplaceActionType_IsValid(t *testing.T) {
	tests := []struct {
		action   MarketplaceActionType
		expected bool
	}{
		{ActionPlaceOrder, true},
		{ActionWithdrawFunds, true},
		{ActionUnspecified, false},
		{MarketplaceActionType(100), false},
	}

	for _, tt := range tests {
		t.Run(tt.action.String(), func(t *testing.T) {
			if got := tt.action.IsValid(); got != tt.expected {
				t.Errorf("MarketplaceActionType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMFAGatingChecker_Check(t *testing.T) {
	checker := NewMFAGatingChecker()
	now := time.Now()
	recentMFA := now.Add(-30 * time.Minute)
	oldMFA := now.Add(-2 * time.Hour)

	tests := []struct {
		name      string
		ctx       *MFAGatingContext
		required  bool
		satisfied bool
	}{
		{
			name: "trusted device reduces requirement",
			ctx: &MFAGatingContext{
				ActionType:      ActionCancelOrder,
				IsTrustedDevice: true,
			},
			required:  false,
			satisfied: true,
		},
		{
			name: "recent MFA satisfies",
			ctx: &MFAGatingContext{
				ActionType:        ActionCancelOrder,
				IsTrustedDevice:   false,
				LastMFAVerifiedAt: &recentMFA,
			},
			required:  true,
			satisfied: true,
		},
		{
			name: "old MFA doesn't satisfy",
			ctx: &MFAGatingContext{
				ActionType:        ActionCancelOrder,
				IsTrustedDevice:   false,
				LastMFAVerifiedAt: &oldMFA,
			},
			required:  true,
			satisfied: false,
		},
		{
			name: "withdrawal always requires fresh MFA",
			ctx: &MFAGatingContext{
				ActionType:        ActionWithdrawFunds,
				IsTrustedDevice:   true, // trusted device doesn't help
				LastMFAVerifiedAt: &recentMFA,
			},
			required:  true,
			satisfied: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.Check(tt.ctx)
			if result.Required != tt.required {
				t.Errorf("MFAGatingChecker.Check() Required = %v, want %v", result.Required, tt.required)
			}
			if result.Satisfied != tt.satisfied {
				t.Errorf("MFAGatingChecker.Check() Satisfied = %v, want %v", result.Satisfied, tt.satisfied)
			}
		})
	}
}

func TestMFAAuditRecord(t *testing.T) {
	record := NewMFAAuditRecord(ActionPlaceOrder, "cosmos1abc", "challenge123")

	if record.ActionType != ActionPlaceOrder {
		t.Errorf("MFAAuditRecord.ActionType = %v, want %v", record.ActionType, ActionPlaceOrder)
	}
	if record.AttemptCount != 1 {
		t.Errorf("MFAAuditRecord.AttemptCount = %d, want 1", record.AttemptCount)
	}

	record.RecordSuccess([]string{"totp", "fido2"})
	if !record.Success {
		t.Error("MFAAuditRecord.RecordSuccess() should set Success to true")
	}
	if len(record.FactorTypesUsed) != 2 {
		t.Errorf("MFAAuditRecord.FactorTypesUsed length = %d, want 2", len(record.FactorTypesUsed))
	}

	record2 := NewMFAAuditRecord(ActionPlaceOrder, "cosmos1abc", "challenge456")
	record2.RecordFailure("invalid code")
	if record2.Success {
		t.Error("MFAAuditRecord.RecordFailure() should set Success to false")
	}
	if record2.FailureReason != "invalid code" {
		t.Errorf("MFAAuditRecord.FailureReason = %s, want 'invalid code'", record2.FailureReason)
	}
}

// ============================================================================
// Waldur Bridge Tests (VE-303)
// ============================================================================

func TestWaldurSyncRecord_NeedsSync(t *testing.T) {
	tests := []struct {
		name     string
		record   WaldurSyncRecord
		expected bool
	}{
		{
			name: "pending needs sync",
			record: WaldurSyncRecord{
				State: SyncStatePending,
			},
			expected: true,
		},
		{
			name: "out of sync needs sync",
			record: WaldurSyncRecord{
				State: SyncStateOutOfSync,
			},
			expected: true,
		},
		{
			name: "synced but version mismatch",
			record: WaldurSyncRecord{
				State:        SyncStateSynced,
				SyncVersion:  1,
				ChainVersion: 2,
			},
			expected: true,
		},
		{
			name: "fully synced",
			record: WaldurSyncRecord{
				State:        SyncStateSynced,
				SyncVersion:  2,
				ChainVersion: 2,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.record.NeedsSync(); got != tt.expected {
				t.Errorf("WaldurSyncRecord.NeedsSync() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWaldurCallback_Validate(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	validCallback := NewWaldurCallbackAt(
		ActionTypeProvision,
		"waldur123",
		SyncTypeAllocation,
		"allocation123",
		now,
	)
	validCallback.Signature = []byte("signature")

	err := validCallback.ValidateAt(now)
	if err != nil {
		t.Errorf("WaldurCallback.Validate() unexpected error: %v", err)
	}

	// Test expired callback
	expiredCallback := *validCallback
	expiredCallback.ExpiresAt = now.Add(-time.Hour)
	err = expiredCallback.ValidateAt(now)
	if err == nil {
		t.Error("WaldurCallback.Validate() expected error for expired callback")
	}

	// Test missing signature
	noSigCallback := *validCallback
	noSigCallback.Signature = nil
	err = noSigCallback.ValidateAt(now)
	if err == nil {
		t.Error("WaldurCallback.Validate() expected error for missing signature")
	}
}

func TestProcessedNonces(t *testing.T) {
	nonces := NewProcessedNonces(time.Hour)
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Initially not processed
	if nonces.IsProcessedAt("nonce1", now) {
		t.Error("ProcessedNonces.IsProcessed() should return false for new nonce")
	}

	// Mark as processed
	nonces.MarkProcessedAt("nonce1", now)
	if !nonces.IsProcessedAt("nonce1", now) {
		t.Error("ProcessedNonces.IsProcessed() should return true after marking")
	}

	// Test cleanup (manually expire)
	nonces.Nonces["expired_nonce"] = now.Add(-time.Hour)
	nonces.CleanupAt(now)
	if _, exists := nonces.Nonces["expired_nonce"]; exists {
		t.Error("ProcessedNonces.Cleanup() should remove expired nonces")
	}
}

func TestWaldurOfferingExport_FromOffering(t *testing.T) {
	offering := &Offering{
		ID: OfferingID{
			ProviderAddress: "cosmos1xyz",
			Sequence:        1,
		},
		Name:        "Test Offering",
		Description: "A test offering",
		Category:    OfferingCategoryCompute,
		State:       OfferingStateActive,
		Pricing: PricingInfo{
			Model:     PricingModelHourly,
			BasePrice: 1000,
			Currency:  "uve",
		},
		IdentityRequirement: IdentityRequirement{
			MinScore: 50,
		},
		RequireMFAForOrders: true,
		Regions:             []string{"us-east", "eu-west"},
		Tags:                []string{"compute", "gpu"},
		UpdatedAt:           time.Now(),
	}

	export := &WaldurOfferingExport{}
	export.FromOffering(offering)

	if export.ChainOfferingID != offering.ID.String() {
		t.Errorf("WaldurOfferingExport.ChainOfferingID = %s, want %s", export.ChainOfferingID, offering.ID.String())
	}
	if export.Name != offering.Name {
		t.Errorf("WaldurOfferingExport.Name = %s, want %s", export.Name, offering.Name)
	}
	if export.IdentityScoreRequired != 50 {
		t.Errorf("WaldurOfferingExport.IdentityScoreRequired = %d, want 50", export.IdentityScoreRequired)
	}
	if !export.MFARequired {
		t.Error("WaldurOfferingExport.MFARequired should be true")
	}
}

// ============================================================================
// Events Tests (VE-304)
// ============================================================================

func TestMarketplaceEventType_IsProviderDaemonEvent(t *testing.T) {
	daemonEvents := []MarketplaceEventType{
		EventOrderCreated,
		EventBidAccepted,
		EventAllocationCreated,
		EventProvisionRequested,
		EventTerminateRequested,
		EventUsageUpdateRequested,
	}

	nonDaemonEvents := []MarketplaceEventType{
		EventOfferingCreated,
		EventOfferingUpdated,
		EventOrderStateChanged,
	}

	for _, evt := range daemonEvents {
		if !evt.IsProviderDaemonEvent() {
			t.Errorf("MarketplaceEventType(%s).IsProviderDaemonEvent() = false, want true", evt)
		}
	}

	for _, evt := range nonDaemonEvents {
		if evt.IsProviderDaemonEvent() {
			t.Errorf("MarketplaceEventType(%s).IsProviderDaemonEvent() = true, want false", evt)
		}
	}
}

func TestNewOrderCreatedEvent(t *testing.T) {
	order := NewOrder(
		OrderID{CustomerAddress: "cosmos1abc", Sequence: 1},
		OfferingID{ProviderAddress: "cosmos1xyz", Sequence: 1},
		1000,
		1,
	)

	event := NewOrderCreatedEvent(order, 100, 1)

	if event.EventType != EventOrderCreated {
		t.Errorf("OrderCreatedEvent.EventType = %s, want %s", event.EventType, EventOrderCreated)
	}
	if event.BlockHeight != 100 {
		t.Errorf("OrderCreatedEvent.BlockHeight = %d, want 100", event.BlockHeight)
	}
	if event.Sequence != 1 {
		t.Errorf("OrderCreatedEvent.Sequence = %d, want 1", event.Sequence)
	}
	if event.OrderID != order.ID.String() {
		t.Errorf("OrderCreatedEvent.OrderID = %s, want %s", event.OrderID, order.ID.String())
	}
}

func TestEventCheckpoint(t *testing.T) {
	checkpoint := NewEventCheckpoint("subscriber1")

	if checkpoint.SubscriberID != "subscriber1" {
		t.Errorf("EventCheckpoint.SubscriberID = %s, want subscriber1", checkpoint.SubscriberID)
	}
	if checkpoint.LastSequence != 0 {
		t.Errorf("EventCheckpoint.LastSequence = %d, want 0", checkpoint.LastSequence)
	}

	// Update with event
	event := &OrderCreatedEvent{
		BaseMarketplaceEvent: BaseMarketplaceEvent{
			EventID:     "evt_123",
			BlockHeight: 100,
			Sequence:    5,
		},
	}
	checkpoint.Update(event)

	if checkpoint.LastSequence != 5 {
		t.Errorf("EventCheckpoint.LastSequence = %d, want 5", checkpoint.LastSequence)
	}
	if checkpoint.LastBlockHeight != 100 {
		t.Errorf("EventCheckpoint.LastBlockHeight = %d, want 100", checkpoint.LastBlockHeight)
	}
}

func TestEventBatch(t *testing.T) {
	batch := NewEventBatch()

	if batch.Size() != 0 {
		t.Errorf("EventBatch.Size() = %d, want 0", batch.Size())
	}

	event1 := &OrderCreatedEvent{
		BaseMarketplaceEvent: BaseMarketplaceEvent{
			Sequence:    1,
			BlockHeight: 100,
		},
	}
	event2 := &OrderCreatedEvent{
		BaseMarketplaceEvent: BaseMarketplaceEvent{
			Sequence:    2,
			BlockHeight: 101,
		},
	}

	batch.Add(event1)
	batch.Add(event2)

	if batch.Size() != 2 {
		t.Errorf("EventBatch.Size() = %d, want 2", batch.Size())
	}
	if batch.FromSequence != 1 {
		t.Errorf("EventBatch.FromSequence = %d, want 1", batch.FromSequence)
	}
	if batch.ToSequence != 2 {
		t.Errorf("EventBatch.ToSequence = %d, want 2", batch.ToSequence)
	}
}

func TestNewProviderDaemonSubscription(t *testing.T) {
	sub := NewProviderDaemonSubscription("cosmos1provider")

	if sub.ProviderAddress != "cosmos1provider" {
		t.Errorf("EventSubscription.ProviderAddress = %s, want cosmos1provider", sub.ProviderAddress)
	}
	if sub.FilterByProvider != "cosmos1provider" {
		t.Errorf("EventSubscription.FilterByProvider = %s, want cosmos1provider", sub.FilterByProvider)
	}
	if !sub.Active {
		t.Error("EventSubscription.Active should be true")
	}
	if len(sub.EventTypes) == 0 {
		t.Error("EventSubscription.EventTypes should not be empty")
	}
}

func TestNewUsageUpdateRequestedEvent(t *testing.T) {
	event := NewUsageUpdateRequestedEvent("alloc_123", "cosmos1provider", "periodic", 100, 5)

	if event.EventType != EventUsageUpdateRequested {
		t.Errorf("UsageUpdateRequestedEvent.EventType = %s, want %s", event.EventType, EventUsageUpdateRequested)
	}
	if event.BlockHeight != 100 {
		t.Errorf("UsageUpdateRequestedEvent.BlockHeight = %d, want 100", event.BlockHeight)
	}
	if event.Sequence != 5 {
		t.Errorf("UsageUpdateRequestedEvent.Sequence = %d, want 5", event.Sequence)
	}
	if event.AllocationID != "alloc_123" {
		t.Errorf("UsageUpdateRequestedEvent.AllocationID = %s, want alloc_123", event.AllocationID)
	}
	if event.ProviderAddress != "cosmos1provider" {
		t.Errorf("UsageUpdateRequestedEvent.ProviderAddress = %s, want cosmos1provider", event.ProviderAddress)
	}
	if event.RequestType != "periodic" {
		t.Errorf("UsageUpdateRequestedEvent.RequestType = %s, want periodic", event.RequestType)
	}
	if event.EventID == "" {
		t.Error("UsageUpdateRequestedEvent.EventID should not be empty")
	}
}

func TestNewProvisionRequestedEvent(t *testing.T) {
	allocation := NewAllocation(
		AllocationID{
			OrderID:  OrderID{CustomerAddress: "cosmos1abc", Sequence: 1},
			Sequence: 1,
		},
		OfferingID{ProviderAddress: "cosmos1xyz", Sequence: 1},
		"cosmos1xyz",
		BidID{
			OrderID:         OrderID{CustomerAddress: "cosmos1abc", Sequence: 1},
			ProviderAddress: "cosmos1xyz",
			Sequence:        1,
		},
		1000,
	)

	event := NewProvisionRequestedEvent(allocation, "encrypted_config_ref_123", 100, 5)

	if event.EventType != EventProvisionRequested {
		t.Errorf("ProvisionRequestedEvent.EventType = %s, want %s", event.EventType, EventProvisionRequested)
	}
	if event.BlockHeight != 100 {
		t.Errorf("ProvisionRequestedEvent.BlockHeight = %d, want 100", event.BlockHeight)
	}
	if event.Sequence != 5 {
		t.Errorf("ProvisionRequestedEvent.Sequence = %d, want 5", event.Sequence)
	}
	if event.AllocationID != allocation.ID.String() {
		t.Errorf("ProvisionRequestedEvent.AllocationID = %s, want %s", event.AllocationID, allocation.ID.String())
	}
	if event.ProviderAddress != "cosmos1xyz" {
		t.Errorf("ProvisionRequestedEvent.ProviderAddress = %s, want cosmos1xyz", event.ProviderAddress)
	}
	if event.EncryptedConfigRef != "encrypted_config_ref_123" {
		t.Errorf("ProvisionRequestedEvent.EncryptedConfigRef = %s, want encrypted_config_ref_123", event.EncryptedConfigRef)
	}
	if event.EventID == "" {
		t.Error("ProvisionRequestedEvent.EventID should not be empty")
	}
}

func TestNewTerminateRequestedEvent(t *testing.T) {
	event := NewTerminateRequestedEvent(
		"alloc_123",
		"order_456",
		"cosmos1provider",
		"cosmos1customer",
		"user requested termination",
		true,
		100,
		5,
	)

	if event.EventType != EventTerminateRequested {
		t.Errorf("TerminateRequestedEvent.EventType = %s, want %s", event.EventType, EventTerminateRequested)
	}
	if event.BlockHeight != 100 {
		t.Errorf("TerminateRequestedEvent.BlockHeight = %d, want 100", event.BlockHeight)
	}
	if event.Sequence != 5 {
		t.Errorf("TerminateRequestedEvent.Sequence = %d, want 5", event.Sequence)
	}
	if event.AllocationID != "alloc_123" {
		t.Errorf("TerminateRequestedEvent.AllocationID = %s, want alloc_123", event.AllocationID)
	}
	if event.OrderID != "order_456" {
		t.Errorf("TerminateRequestedEvent.OrderID = %s, want order_456", event.OrderID)
	}
	if event.ProviderAddress != "cosmos1provider" {
		t.Errorf("TerminateRequestedEvent.ProviderAddress = %s, want cosmos1provider", event.ProviderAddress)
	}
	if event.RequestedBy != "cosmos1customer" {
		t.Errorf("TerminateRequestedEvent.RequestedBy = %s, want cosmos1customer", event.RequestedBy)
	}
	if event.Reason != "user requested termination" {
		t.Errorf("TerminateRequestedEvent.Reason = %s, want 'user requested termination'", event.Reason)
	}
	if !event.Immediate {
		t.Error("TerminateRequestedEvent.Immediate should be true")
	}
	if event.EventID == "" {
		t.Error("TerminateRequestedEvent.EventID should not be empty")
	}
}

// ============================================================================
// Genesis Tests
// ============================================================================

func TestParams_Validate(t *testing.T) {
	validParams := DefaultParams()
	if err := validParams.Validate(); err != nil {
		t.Errorf("Params.Validate() unexpected error: %v", err)
	}

	invalidParams := Params{
		MaxOfferingsPerProvider: 0,
	}
	if err := invalidParams.Validate(); err == nil {
		t.Error("Params.Validate() expected error for zero MaxOfferingsPerProvider")
	}
}

func TestGenesisState_Validate(t *testing.T) {
	validGenesis := DefaultGenesisState()
	if err := validGenesis.Validate(); err != nil {
		t.Errorf("GenesisState.Validate() unexpected error: %v", err)
	}

	// Test with duplicate offering ID
	invalidGenesis := DefaultGenesisState()
	offering := Offering{
		ID: OfferingID{
			ProviderAddress: "cosmos1xyz",
			Sequence:        1,
		},
		State:    OfferingStateActive,
		Category: OfferingCategoryCompute,
		Name:     "Test",
		Pricing: PricingInfo{
			Model:    PricingModelHourly,
			Currency: "uve",
		},
	}
	invalidGenesis.Offerings = append(invalidGenesis.Offerings, offering, offering)
	if err := invalidGenesis.Validate(); err == nil {
		t.Error("GenesisState.Validate() expected error for duplicate offering")
	}
}
