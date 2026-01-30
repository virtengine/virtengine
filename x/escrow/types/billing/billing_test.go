package billing

import (
	"bytes"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// testAddr generates a valid test bech32 address from a seed number
func testAddr(seed int) string {
	var buffer bytes.Buffer
	buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6")
	buffer.WriteString(string(rune('0' + (seed/100)%10)))
	buffer.WriteString(string(rune('0' + (seed/10)%10)))
	buffer.WriteString(string(rune('0' + seed%10)))
	res, _ := sdk.AccAddressFromHexUnsafe(buffer.String())
	return res.String()
}

func TestInvoiceValidation(t *testing.T) {
	now := time.Now()
	providerAddr := testAddr(100)
	customerAddr := testAddr(101)
	
	tests := []struct {
		name    string
		invoice Invoice
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid invoice",
			invoice: Invoice{
				InvoiceID:     "inv-001",
				InvoiceNumber: "VE-00000001",
				EscrowID:      "escrow-001",
				OrderID:       "order-001",
				LeaseID:       "lease-001",
				Provider:      providerAddr,
				Customer:      customerAddr,
				Status:        InvoiceStatusDraft,
				Currency:      "uvirt",
				BillingPeriod: BillingPeriod{
					StartTime: now.Add(-24 * time.Hour),
					EndTime:   now,
				},
				LineItems: []LineItem{
					{
						LineItemID:  "line-001",
						Description: "CPU Usage",
						UsageType:   UsageTypeCPU,
						Quantity:    sdkmath.LegacyNewDec(10),
						Unit:        "core-hour",
						UnitPrice:   sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDec(100)),
						Amount:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
					},
				},
				Subtotal:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				DiscountTotal: sdk.NewCoins(),
				TaxTotal:      sdk.NewCoins(),
				Total:         sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				AmountPaid:    sdk.NewCoins(),
				AmountDue:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
				DueDate:       now.Add(7 * 24 * time.Hour),
				IssuedAt:      now,
				BlockHeight:   100,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			wantErr: false,
		},
		{
			name: "missing invoice_id",
			invoice: Invoice{
				EscrowID: "escrow-001",
				Provider: providerAddr,
				Customer: customerAddr,
				Currency: "uvirt",
				Total:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
			wantErr: true,
			errMsg:  "invoice_id is required",
		},
		{
			name: "invalid provider address",
			invoice: Invoice{
				InvoiceID: "inv-001",
				EscrowID:  "escrow-001",
				Provider:  "invalid",
				Customer:  customerAddr,
				Currency:  "uvirt",
				Total:     sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
			wantErr: true,
			errMsg:  "invalid provider address",
		},
		{
			name: "invalid billing period",
			invoice: Invoice{
				InvoiceID: "inv-001",
				EscrowID:  "escrow-001",
				Provider:  providerAddr,
				Customer:  customerAddr,
				Currency:  "uvirt",
				BillingPeriod: BillingPeriod{
					StartTime: now,
					EndTime:   now.Add(-24 * time.Hour), // End before start
				},
				Total: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
			wantErr: true,
			errMsg:  "billing period end_time must be after start_time",
		},
		{
			name: "missing currency",
			invoice: Invoice{
				InvoiceID: "inv-001",
				EscrowID:  "escrow-001",
				Provider:  providerAddr,
				Customer:  customerAddr,
				BillingPeriod: BillingPeriod{
					StartTime: now.Add(-24 * time.Hour),
					EndTime:   now,
				},
				Total: sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
			},
			wantErr: true,
			errMsg:  "currency is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.invoice.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestInvoiceRecordPayment(t *testing.T) {
	now := time.Now()
	
	inv := &Invoice{
		Status:     InvoiceStatusPending,
		Total:      sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
		AmountPaid: sdk.NewCoins(),
		AmountDue:  sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
	}

	// Partial payment
	err := inv.RecordPayment(sdk.NewCoins(sdk.NewInt64Coin("uvirt", 500)), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inv.Status != InvoiceStatusPartiallyPaid {
		t.Errorf("expected status %v, got %v", InvoiceStatusPartiallyPaid, inv.Status)
	}

	if !inv.AmountPaid.Equal(sdk.NewCoins(sdk.NewInt64Coin("uvirt", 500))) {
		t.Errorf("expected amount_paid 500, got %v", inv.AmountPaid)
	}

	// Full payment
	err = inv.RecordPayment(sdk.NewCoins(sdk.NewInt64Coin("uvirt", 500)), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inv.Status != InvoiceStatusPaid {
		t.Errorf("expected status %v, got %v", InvoiceStatusPaid, inv.Status)
	}

	if inv.PaidAt == nil {
		t.Error("expected paid_at to be set")
	}
}

func TestRoundHalfEven(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1.5", 2},   // Round to even
		{"2.5", 2},   // Round to even
		{"3.5", 4},   // Round to even
		{"4.5", 4},   // Round to even
		{"1.4", 1},   // Round down
		{"1.6", 2},   // Round up
		{"0.5", 0},   // Round to even
		{"1.0", 1},   // Exact
		{"2.0", 2},   // Exact
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			dec := sdkmath.LegacyMustNewDecFromStr(tt.input)
			result := roundHalfEven(dec)
			if result.Int64() != tt.expected {
				t.Errorf("roundHalfEven(%s) = %d, want %d", tt.input, result.Int64(), tt.expected)
			}
		})
	}
}

func TestPricingTierEffectiveRate(t *testing.T) {
	tier := PricingTier{
		TierID:      "tier-1",
		Rate:        sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDec(100)),
		DiscountBps: 1000, // 10% discount
	}

	effectiveRate := tier.GetEffectiveRate()
	expected := sdkmath.LegacyNewDec(90) // 100 - 10%

	if !effectiveRate.Amount.Equal(expected) {
		t.Errorf("expected effective rate %s, got %s", expected, effectiveRate.Amount)
	}
}

func TestResourcePricingCalculate(t *testing.T) {
	rp := ResourcePricing{
		ResourceType:       UsageTypeCPU,
		BaseRate:           sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyNewDec(100)),
		Unit:               "core-hour",
		MinQuantity:        sdkmath.LegacyOneDec(),
		GranularitySeconds: 3600,
	}

	// Test basic calculation
	qty := sdkmath.LegacyNewDec(10)
	amount := rp.CalculateAmount(qty, RoundingModeHalfEven)

	if amount.Amount.Int64() != 1000 {
		t.Errorf("expected amount 1000, got %d", amount.Amount.Int64())
	}

	// Test minimum quantity
	smallQty := sdkmath.LegacyNewDecWithPrec(1, 1) // 0.1
	amount = rp.CalculateAmount(smallQty, RoundingModeHalfEven)

	if amount.Amount.Int64() != 100 { // min quantity is 1, so 1 * 100
		t.Errorf("expected amount 100 (minimum), got %d", amount.Amount.Int64())
	}
}

func TestDiscountPolicyValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		policy  DiscountPolicy
		wantErr bool
	}{
		{
			name: "valid percentage discount",
			policy: DiscountPolicy{
				PolicyID:      "disc-001",
				Name:          "10% Off",
				Type:          DiscountTypePercentage,
				PercentageBps: 1000, // 10%
				EffectiveFrom: now,
				IsActive:      true,
			},
			wantErr: false,
		},
		{
			name: "invalid percentage (over 100%)",
			policy: DiscountPolicy{
				PolicyID:      "disc-001",
				Name:          "Invalid",
				Type:          DiscountTypePercentage,
				PercentageBps: 15000, // 150%
				EffectiveFrom: now,
				IsActive:      true,
			},
			wantErr: true,
		},
		{
			name: "valid fixed discount",
			policy: DiscountPolicy{
				PolicyID:    "disc-002",
				Name:        "$10 Off",
				Type:        DiscountTypeFixed,
				FixedAmount: sdk.NewInt64Coin("uvirt", 10000000),
				IsActive:    true,
			},
			wantErr: false,
		},
		{
			name: "volume discount without thresholds",
			policy: DiscountPolicy{
				PolicyID: "disc-003",
				Name:     "Volume Discount",
				Type:     DiscountTypeVolume,
				IsActive: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCouponCodeRedemption(t *testing.T) {
	now := time.Now()

	coupon := &CouponCode{
		Code:             "SAVE10",
		DiscountPolicyID: "disc-001",
		MaxRedemptions:   100,
		PerCustomerLimit: 1,
		ValidFrom:        now.Add(-24 * time.Hour),
		ValidUntil:       now.Add(24 * time.Hour),
		IsActive:         true,
	}

	customer := "cosmos1customerxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	// First redemption should succeed
	err := coupon.CanRedeem(customer, now)
	if err != nil {
		t.Fatalf("expected first redemption to succeed: %v", err)
	}

	coupon.RecordRedemption(customer)

	// Second redemption should fail (per-customer limit)
	err = coupon.CanRedeem(customer, now)
	if err == nil {
		t.Error("expected second redemption to fail")
	}

	// Different customer should succeed
	err = coupon.CanRedeem("cosmos1anotherxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", now)
	if err != nil {
		t.Errorf("expected different customer to succeed: %v", err)
	}

	// Expired coupon should fail
	err = coupon.CanRedeem(customer, now.Add(48*time.Hour))
	if err == nil {
		t.Error("expected expired coupon to fail")
	}
}

func TestTaxJurisdictionValidation(t *testing.T) {
	tests := []struct {
		name    string
		juris   TaxJurisdiction
		wantErr bool
	}{
		{
			name: "valid jurisdiction",
			juris: TaxJurisdiction{
				JurisdictionID: "GB",
				CountryCode:    "GB",
				CountryName:    "United Kingdom",
				TaxRates: []TaxRate{
					{
						RateID:  "GB-VAT-20",
						TaxType: TaxTypeVAT,
						Name:    "VAT",
						RateBps: 2000,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid country code length",
			juris: TaxJurisdiction{
				JurisdictionID: "GBR",
				CountryCode:    "GBR", // Should be 2 chars
				CountryName:    "United Kingdom",
			},
			wantErr: true,
		},
		{
			name: "missing jurisdiction_id",
			juris: TaxJurisdiction{
				CountryCode: "GB",
				CountryName: "United Kingdom",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.juris.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestTaxCalculator(t *testing.T) {
	tc := NewTaxCalculator("US")
	
	// Add UK jurisdiction
	err := tc.AddJurisdiction(TaxJurisdiction{
		JurisdictionID: "GB",
		CountryCode:    "GB",
		CountryName:    "United Kingdom",
		TaxRates: []TaxRate{
			{
				RateID:  "GB-VAT-20",
				TaxType: TaxTypeVAT,
				Name:    "VAT Standard Rate",
				RateBps: 2000, // 20%
				ExemptCategories: []TaxExemptionCategory{
					TaxExemptionB2B,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to add jurisdiction: %v", err)
	}

	subtotal := sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000))
	now := time.Now()

	// Test B2C customer in UK
	details, err := tc.CalculateTax(subtotal, "GB", TaxExemptionNone, "GB", now)
	if err != nil {
		t.Fatalf("failed to calculate tax: %v", err)
	}

	if len(details.TaxLineItems) != 1 {
		t.Errorf("expected 1 tax line item, got %d", len(details.TaxLineItems))
	}

	expectedTax := int64(200) // 20% of 1000
	if !details.TotalTax.Equal(sdk.NewCoins(sdk.NewInt64Coin("uvirt", expectedTax))) {
		t.Errorf("expected total tax %d, got %v", expectedTax, details.TotalTax)
	}

	// Test B2B customer (exempt)
	details, err = tc.CalculateTax(subtotal, "GB", TaxExemptionB2B, "US", now)
	if err != nil {
		t.Fatalf("failed to calculate tax for B2B: %v", err)
	}

	if !details.IsReverseCharge {
		t.Error("expected reverse charge for cross-border B2B")
	}

	if !details.TotalTax.IsZero() {
		t.Errorf("expected zero tax for B2B reverse charge, got %v", details.TotalTax)
	}
}

func TestDisputeWindow(t *testing.T) {
	now := time.Now()

	window := NewDisputeWindow(
		"dispute-001",
		"inv-001",
		now,
		604800, // 7 days
	)

	// Window should be open
	if !window.IsOpen(now.Add(time.Hour)) {
		t.Error("expected window to be open")
	}

	// Initiate dispute
	err := window.InitiateDispute(
		"cosmos1customerxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"Incorrect charges",
		now.Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("failed to initiate dispute: %v", err)
	}

	if window.Status != DisputeStatusUnderReview {
		t.Errorf("expected status under_review, got %s", window.Status)
	}

	// Cannot initiate again
	err = window.InitiateDispute("cosmos1xxx", "reason", now)
	if err == nil {
		t.Error("expected error when initiating again")
	}

	// Resolve dispute
	err = window.Resolve(
		DisputeResolutionPartialRefund,
		"Partial refund for service issues",
		"cosmos1arbiterxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		sdk.NewCoins(sdk.NewInt64Coin("uvirt", 500)),
		now.Add(2*time.Hour),
	)
	if err != nil {
		t.Fatalf("failed to resolve dispute: %v", err)
	}

	if window.Status != DisputeStatusResolved {
		t.Errorf("expected status resolved, got %s", window.Status)
	}

	// Window should be closed after expiry
	if window.IsOpen(now.Add(8 * 24 * time.Hour)) {
		t.Error("expected window to be closed after expiry")
	}
}

func TestSettlementConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  SettlementConfig
		wantErr bool
	}{
		{
			name:    "default config is valid",
			config:  DefaultSettlementConfig(),
			wantErr: false,
		},
		{
			name: "invalid dispute window",
			config: SettlementConfig{
				DefaultDisputeWindowSeconds: -1,
			},
			wantErr: true,
		},
		{
			name: "max less than min",
			config: SettlementConfig{
				DefaultDisputeWindowSeconds: 86400,
				MinDisputeWindowSeconds:     172800,
				MaxDisputeWindowSeconds:     86400,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoyaltyPointsManagement(t *testing.T) {
	now := time.Now()

	loyalty := &CustomerLoyalty{
		Customer:          "cosmos1customerxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		ProgramID:         "program-001",
		TotalPointsEarned: 0,
		AvailablePoints:   0,
		JoinedAt:          now,
	}

	// Earn points
	loyalty.EarnPoints(1000, now)
	if loyalty.AvailablePoints != 1000 {
		t.Errorf("expected 1000 available points, got %d", loyalty.AvailablePoints)
	}

	// Redeem points
	err := loyalty.RedeemPoints(300, now)
	if err != nil {
		t.Fatalf("failed to redeem points: %v", err)
	}

	if loyalty.AvailablePoints != 700 {
		t.Errorf("expected 700 available points after redemption, got %d", loyalty.AvailablePoints)
	}

	if loyalty.RedeemedPoints != 300 {
		t.Errorf("expected 300 redeemed points, got %d", loyalty.RedeemedPoints)
	}

	// Cannot redeem more than available
	err = loyalty.RedeemPoints(1000, now)
	if err == nil {
		t.Error("expected error when redeeming more than available")
	}
}

func TestBuildKeys(t *testing.T) {
	// Test invoice key
	invoiceKey := BuildInvoiceKey("inv-001")
	if len(invoiceKey) != len(InvoicePrefix)+7 {
		t.Errorf("unexpected invoice key length: %d", len(invoiceKey))
	}

	invoiceID, err := ParseInvoiceKey(invoiceKey)
	if err != nil {
		t.Fatalf("failed to parse invoice key: %v", err)
	}
	if invoiceID != "inv-001" {
		t.Errorf("expected invoice ID 'inv-001', got '%s'", invoiceID)
	}

	// Test invoice by provider key
	providerKey := BuildInvoiceByProviderKey("cosmos1provider", "inv-001")
	if len(providerKey) == 0 {
		t.Error("expected non-empty provider key")
	}

	// Test next invoice number
	nextNum := NextInvoiceNumber(0, "VE")
	if nextNum != "VE-00000001" {
		t.Errorf("expected 'VE-00000001', got '%s'", nextNum)
	}

	nextNum = NextInvoiceNumber(999, "VE")
	if nextNum != "VE-00001000" {
		t.Errorf("expected 'VE-00001000', got '%s'", nextNum)
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
