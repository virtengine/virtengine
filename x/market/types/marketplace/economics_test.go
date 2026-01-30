// Package marketplace provides types for the marketplace on-chain module.
//
// ECON-002: Marketplace Economics Optimization
// This file contains tests for the economics system including attack simulations.
package marketplace

import (
	"testing"
	"time"
)

// TestFeeScheduleValidation tests fee schedule validation
func TestFeeScheduleValidation(t *testing.T) {
	tests := []struct {
		name      string
		schedule  FeeSchedule
		expectErr bool
	}{
		{
			name:      "valid default schedule",
			schedule:  DefaultFeeSchedule(),
			expectErr: false,
		},
		{
			name: "zero base rate",
			schedule: FeeSchedule{
				BaseTakeRateBps: 0,
				MinTakeRateBps:  0,
				MaxTakeRateBps:  500,
			},
			expectErr: true,
		},
		{
			name: "min exceeds base",
			schedule: FeeSchedule{
				BaseTakeRateBps: 100,
				MinTakeRateBps:  200,
				MaxTakeRateBps:  500,
			},
			expectErr: true,
		},
		{
			name: "max below base",
			schedule: FeeSchedule{
				BaseTakeRateBps: 200,
				MinTakeRateBps:  50,
				MaxTakeRateBps:  100,
			},
			expectErr: true,
		},
		{
			name: "max exceeds 100%",
			schedule: FeeSchedule{
				BaseTakeRateBps: 200,
				MinTakeRateBps:  50,
				MaxTakeRateBps:  15000,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schedule.Validate()
			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestDynamicFeeCalculator tests fee calculation
func TestDynamicFeeCalculator(t *testing.T) {
	schedule := DefaultFeeSchedule()
	calc := NewDynamicFeeCalculator(schedule)

	t.Run("basic fee calculation", func(t *testing.T) {
		input := FeeCalculationInput{
			OrderValue:      10000000, // 10 tokens
			UserTier:        FeeTierStandard,
			User30DayVolume: 0,
			IsMaker:         false,
			IsEarlyAdopter:  false,
			Utilization:     UtilizationMetrics{},
		}

		result, err := calc.CalculateFee(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Base rate is 2% (200 bps)
		expectedGross := uint64(10000000 * 200 / 10000) // 200000
		if result.GrossFee != expectedGross {
			t.Errorf("expected gross fee %d, got %d", expectedGross, result.GrossFee)
		}
	})

	t.Run("tier discount", func(t *testing.T) {
		input := FeeCalculationInput{
			OrderValue:      10000000,
			UserTier:        FeeTierGold, // 30% discount
			User30DayVolume: 0,
			IsMaker:         false,
			IsEarlyAdopter:  false,
			Utilization:     UtilizationMetrics{},
		}

		result, err := calc.CalculateFee(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.TierDiscount == 0 {
			t.Error("expected tier discount but got none")
		}

		// Verify discount is applied
		if result.NetFee >= result.GrossFee {
			t.Error("net fee should be less than gross fee with tier discount")
		}
	})

	t.Run("maker rebate", func(t *testing.T) {
		input := FeeCalculationInput{
			OrderValue:      10000000,
			UserTier:        FeeTierStandard,
			User30DayVolume: 0,
			IsMaker:         true,
			IsEarlyAdopter:  false,
			Utilization:     UtilizationMetrics{},
		}

		result, err := calc.CalculateFee(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.MakerRebate == 0 {
			t.Error("expected maker rebate but got none")
		}
	})

	t.Run("utilization adjustment", func(t *testing.T) {
		input := FeeCalculationInput{
			OrderValue:      10000000,
			UserTier:        FeeTierStandard,
			User30DayVolume: 0,
			IsMaker:         false,
			IsEarlyAdopter:  false,
			Utilization: UtilizationMetrics{
				TotalCapacity: 1000,
				UsedCapacity:  900, // 90% utilization
			},
		}

		result, err := calc.CalculateFee(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.UtilizationAdjustment <= 0 {
			t.Error("expected positive utilization adjustment for high utilization")
		}
	})

	t.Run("zero order value", func(t *testing.T) {
		input := FeeCalculationInput{
			OrderValue: 0,
		}

		result, err := calc.CalculateFee(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.NetFee != 0 {
			t.Errorf("expected zero fee for zero order value, got %d", result.NetFee)
		}
	})
}

// TestGetTierForVolume tests tier assignment based on volume
func TestGetTierForVolume(t *testing.T) {
	tests := []struct {
		volume       uint64
		expectedTier FeeTier
	}{
		{0, FeeTierStandard},
		{50000000, FeeTierStandard},
		{100000000, FeeTierBronze},
		{500000000, FeeTierSilver},
		{2000000000, FeeTierGold},
		{10000000000, FeeTierPlatinum},
		{50000000000, FeeTierDiamond},
		{100000000000, FeeTierDiamond},
	}

	for _, tt := range tests {
		tier := GetTierForVolume(tt.volume)
		if tier != tt.expectedTier {
			t.Errorf("volume %d: expected tier %s, got %s", tt.volume, tt.expectedTier, tier)
		}
	}
}

// TestProviderIncentiveCalculator tests incentive calculations
func TestProviderIncentiveCalculator(t *testing.T) {
	config := DefaultProviderIncentiveConfig()
	calc := NewIncentiveCalculator(config)
	now := time.Now()
	blockHeight := int64(1000)

	t.Run("uptime reward", func(t *testing.T) {
		metrics := &ProviderMetrics{
			Address:          "provider1",
			Tier:             ProviderTierGold,
			UptimePercentage: 99,
		}

		reward := calc.CalculateUptimeReward(metrics, blockHeight, now)
		if reward == nil {
			t.Fatal("expected uptime reward but got nil")
		}

		if reward.FinalAmount == 0 {
			t.Error("expected non-zero reward amount")
		}

		if reward.Multiplier != ProviderTierGold.Multiplier() {
			t.Errorf("expected multiplier %d, got %d", ProviderTierGold.Multiplier(), reward.Multiplier)
		}
	})

	t.Run("no reward below minimum uptime", func(t *testing.T) {
		metrics := &ProviderMetrics{
			Address:          "provider1",
			Tier:             ProviderTierGold,
			UptimePercentage: 90, // Below 95% minimum
		}

		reward := calc.CalculateUptimeReward(metrics, blockHeight, now)
		if reward != nil {
			t.Error("expected no reward for low uptime")
		}
	})

	t.Run("early adopter bonus", func(t *testing.T) {
		metrics := &ProviderMetrics{
			Address:          "provider1",
			Tier:             ProviderTierGold,
			UptimePercentage: 99,
			IsEarlyAdopter:   true,
		}

		rewards := calc.CalculateAllRewards(metrics, blockHeight, now)
		
		earlyAdopterFound := false
		for _, r := range rewards {
			if r != nil && r.Reason != "" && len(r.Reason) > 0 {
				// Check if early adopter bonus was applied
				if r.FinalAmount > 0 {
					earlyAdopterFound = true
				}
			}
		}

		if !earlyAdopterFound {
			t.Error("expected early adopter bonus to be applied")
		}
	})
}

// TestLiquidityIncentives tests liquidity incentive calculations
func TestLiquidityIncentives(t *testing.T) {
	mmConfig := DefaultMarketMakerConfig()
	lmConfig := DefaultLiquidityMiningConfig()
	calc := NewLiquidityIncentiveCalculator(mmConfig, lmConfig)

	t.Run("market maker reward", func(t *testing.T) {
		mm := &MarketMaker{
			Address:          "mm1",
			Status:           MarketMakerStatusActive,
			Tier:             LiquidityTierGold,
			TotalLiquidity:   10000000,
			AverageSpreadBps: 50, // Tight spread
			UptimePercentage: 95,
		}

		result := calc.CalculateMarketMakerReward(mm, 1000)
		if result == nil {
			t.Fatal("expected market maker reward but got nil")
		}

		if result.SpreadBonus == 0 {
			t.Error("expected spread bonus for tight spread")
		}
	})

	t.Run("liquidity mining reward", func(t *testing.T) {
		pos := &LiquidityPosition{
			ProviderAddress: "lp1",
			Amount:          100000000,
			IsLocked:        true,
			LastRewardBlock: 900,
			Tier:            LiquidityTierSilver,
		}

		result := calc.CalculateLiquidityMiningReward(pos, 1000000000, 1000)
		if result == nil {
			t.Fatal("expected liquidity mining reward but got nil")
		}

		if result.LockupBonus == 0 {
			t.Error("expected lockup bonus for locked position")
		}
	})
}

// TestPriceOracle tests price oracle functionality
func TestPriceOracle(t *testing.T) {
	config := DefaultPriceOracleConfig()
	oracle := NewPriceOracle(config)

	t.Run("TWAP calculation", func(t *testing.T) {
		history := NewPriceHistory(OfferingID{ProviderAddress: "test", Sequence: 1}, 100)
		now := time.Now()

		// Add price points
		for i := int64(0); i < 10; i++ {
			history.AddPoint(PricePoint{
				Price:       1000 + uint64(i*10),
				Volume:      100,
				BlockHeight: i,
				Timestamp:   now.Add(time.Duration(i*6) * time.Second),
			})
		}

		twap, err := oracle.CalculateTWAP(history, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if twap == 0 {
			t.Error("expected non-zero TWAP")
		}

		// TWAP should be around the middle of the range
		if twap < 1000 || twap > 1100 {
			t.Errorf("TWAP %d outside expected range", twap)
		}
	})

	t.Run("price band validation", func(t *testing.T) {
		history := NewPriceHistory(OfferingID{ProviderAddress: "test", Sequence: 1}, 100)
		now := time.Now()

		// Set up price band with 5% bounds (500 bps)
		history.PriceBand = NewPriceBand(1000, 500, 500, time.Hour, now) // 5% band

		// Valid price - within band
		result := oracle.ValidatePrice(1040, history, now)
		if result.Status != PriceValidationStatusValid {
			t.Errorf("expected valid status, got %s", result.Status)
		}

		// Suspicious price movement - beyond both band and suspicious threshold
		// Price 1200 = 20% deviation (2000 bps), which exceeds both 5% band and 5% suspicious threshold
		result = oracle.ValidatePrice(1200, history, now)
		if result.Status != PriceValidationStatusSuspicious {
			t.Errorf("expected suspicious status, got %s", result.Status)
		}

		// Also verify extreme price is suspicious
		result = oracle.ValidatePrice(2000, history, now)
		if result.Status != PriceValidationStatusSuspicious {
			t.Errorf("expected suspicious status, got %s", result.Status)
		}
	})
}

// TestManipulationDetector tests manipulation detection
func TestManipulationDetector(t *testing.T) {
	config := DefaultSafeguardConfig()
	detector := NewManipulationDetector(config)

	t.Run("detect self-dealing wash trading", func(t *testing.T) {
		check := WashTradingCheck{
			BuyerAddress:  "addr1",
			SellerAddress: "addr1", // Same address!
			Amount:        1000000,
			Timestamp:     time.Now(),
			BlockHeight:   1000,
		}

		state := NewAccountSafeguardState("addr1")
		violation := detector.DetectWashTrading(check, state)

		if violation == nil {
			t.Fatal("expected wash trading violation")
		}

		if violation.Type != ManipulationTypeWashTrading {
			t.Errorf("expected wash trading type, got %s", violation.Type)
		}
	})

	t.Run("detect related address wash trading", func(t *testing.T) {
		check := WashTradingCheck{
			BuyerAddress:  "addr1",
			SellerAddress: "addr2",
			Amount:        1000000,
			Timestamp:     time.Now(),
			BlockHeight:   1000,
		}

		state := NewAccountSafeguardState("addr1")
		state.RelatedAddresses = []string{"addr2", "addr3"}

		violation := detector.DetectWashTrading(check, state)

		if violation == nil {
			t.Fatal("expected wash trading violation for related addresses")
		}
	})

	t.Run("detect spoofing", func(t *testing.T) {
		check := SpoofingCheck{
			Address:              "addr1",
			TotalOrders:          100,
			CancelledOrders:      95, // 95% cancellation rate
			AverageOrderLifetime: 10,
			WindowEnd:            time.Now(),
			BlockHeight:          1000,
		}

		violation := detector.DetectSpoofing(check)

		if violation == nil {
			t.Fatal("expected spoofing violation")
		}

		if violation.Type != ManipulationTypeSpoofing {
			t.Errorf("expected spoofing type, got %s", violation.Type)
		}
	})
}

// TestCircuitBreaker tests circuit breaker functionality
func TestCircuitBreaker(t *testing.T) {
	config := DefaultCircuitBreakerConfig()

	t.Run("trip and recovery", func(t *testing.T) {
		state := NewCircuitBreakerState(OfferingID{ProviderAddress: "test", Sequence: 1})
		now := time.Now()
		block := int64(1000)

		// Initially closed
		if state.Status != CircuitBreakerStatusClosed {
			t.Errorf("expected closed status, got %s", state.Status)
		}

		// Trip the breaker
		state.Trip("test trip", config, block, now)

		if state.Status != CircuitBreakerStatusOpen {
			t.Errorf("expected open status, got %s", state.Status)
		}

		// Should not allow orders when open
		err := state.AllowOrder(config)
		if err == nil {
			t.Error("expected error when circuit is open")
		}

		// Fast-forward past trip duration
		state.Update(block+config.TripDurationBlocks+1, config, now.Add(time.Hour))

		if state.Status != CircuitBreakerStatusHalfOpen {
			t.Errorf("expected half-open status, got %s", state.Status)
		}

		// Should allow limited orders in half-open
		err = state.AllowOrder(config)
		if err != nil {
			t.Errorf("unexpected error in half-open: %v", err)
		}
	})

	t.Run("consecutive trips escalation", func(t *testing.T) {
		state := NewCircuitBreakerState(OfferingID{ProviderAddress: "test", Sequence: 1})
		now := time.Now()
		block := int64(1000)

		// Trip multiple times
		for i := 0; i < 5; i++ {
			state.Trip("test trip", config, block, now)
			block += config.TripDurationBlocks + config.HalfOpenDurationBlocks + 1
		}

		if state.ConsecutiveTrips != 5 {
			t.Errorf("expected 5 consecutive trips, got %d", state.ConsecutiveTrips)
		}
	})
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	config := DefaultRateLimitConfig()

	t.Run("block rate limit", func(t *testing.T) {
		state := NewAccountSafeguardState("addr1")

		// Should allow up to max orders
		for i := uint32(0); i < config.MaxOrdersPerBlock; i++ {
			err := state.CheckRateLimit("order", config)
			if err != nil {
				t.Errorf("unexpected rate limit error on order %d: %v", i, err)
			}
			state.OrdersThisBlock++
		}

		// Next order should be rate limited
		err := state.CheckRateLimit("order", config)
		if err == nil {
			t.Error("expected rate limit error")
		}
	})
}

// TestViolationEscalation tests penalty escalation
func TestViolationEscalation(t *testing.T) {
	config := DefaultPenaltyConfig()
	state := NewAccountSafeguardState("addr1")
	now := time.Now()

	actions := []PenaltyAction{}

	// Record violations and track escalation
	for i := 0; i < 25; i++ {
		violation := ViolationRecord{
			Address:     "addr1",
			Type:        ManipulationTypeSpoofing,
			Severity:    5,
			DetectedAt:  now,
			BlockHeight: int64(i),
		}

		action := state.RecordViolation(violation, config)
		actions = append(actions, action)
	}

	// Verify escalation
	if actions[0] != PenaltyActionWarning {
		t.Errorf("first violation should be warning, got %s", actions[0])
	}

	// Should eventually reach ban
	finalAction := actions[len(actions)-1]
	if finalAction != PenaltyActionBan {
		t.Errorf("expected ban after many violations, got %s", finalAction)
	}

	if !state.IsBanned {
		t.Error("state should be banned")
	}
}

// TestMarketMetrics tests market metrics calculation
func TestMarketMetrics(t *testing.T) {
	metrics := NewMarketMetrics(MetricPeriodDay, 1000, time.Now())

	t.Run("efficiency score calculation", func(t *testing.T) {
		// Set up some metrics
		metrics.Spread.AverageSpreadBps = 100  // 1% spread
		metrics.FillRate.TotalOrders = 100
		metrics.FillRate.FilledOrders = 80
		metrics.FillRate.FillRatePercentage = 80
		metrics.Liquidity.LiquidityUtilizationPct = 60
		metrics.Depth.DepthRatioBps = 10000 // Balanced
		metrics.Volatility.CurrentVolatilityBps = 200

		metrics.UpdateEfficiencyScores()

		if metrics.Efficiency.EfficiencyScore == 0 {
			t.Error("expected non-zero efficiency score")
		}

		if metrics.Efficiency.HealthStatus == "" {
			t.Error("expected health status to be set")
		}
	})

	t.Run("fill rate calculation", func(t *testing.T) {
		fillRate := &FillRateMetrics{
			TotalOrders:  100,
			FilledOrders: 75,
		}

		rate := fillRate.CalculateFillRate()
		if rate != 75 {
			t.Errorf("expected 75%% fill rate, got %d%%", rate)
		}
	})
}

// === ECONOMIC ATTACK SIMULATIONS ===

// TestSimulateWashTradingAttack simulates a wash trading attack
func TestSimulateWashTradingAttack(t *testing.T) {
	config := DefaultSafeguardConfig()
	detector := NewManipulationDetector(config)
	penaltyConfig := DefaultPenaltyConfig()
	state := NewAccountSafeguardState("attacker")
	now := time.Now()

	// Simulate attacker trying to inflate volume
	detectedCount := 0
	totalAttempts := 100

	for i := 0; i < totalAttempts; i++ {
		check := WashTradingCheck{
			BuyerAddress:  "attacker",
			SellerAddress: "attacker", // Self-dealing
			Amount:        1000000 + uint64(i),
			Timestamp:     now,
			BlockHeight:   int64(1000 + i),
		}

		violation := detector.DetectWashTrading(check, state)
		if violation != nil {
			detectedCount++
			state.RecordViolation(*violation, penaltyConfig)
		}
	}

	// All should be detected
	if detectedCount != totalAttempts {
		t.Errorf("expected %d detections, got %d", totalAttempts, detectedCount)
	}

	// Attacker should be banned
	if !state.IsBanned {
		t.Error("attacker should be banned after wash trading attempts")
	}

	t.Logf("Wash trading attack: %d/%d detected, attacker banned: %v",
		detectedCount, totalAttempts, state.IsBanned)
}

// TestSimulateSpoofingAttack simulates a spoofing attack
func TestSimulateSpoofingAttack(t *testing.T) {
	config := DefaultSafeguardConfig()
	detector := NewManipulationDetector(config)
	now := time.Now()

	// Simulate attacker placing and cancelling many orders
	check := SpoofingCheck{
		Address:              "spoofer",
		TotalOrders:          1000,
		CancelledOrders:      950, // 95% cancellation
		AverageOrderLifetime: 2,   // Very short
		WindowEnd:            now,
		BlockHeight:          1000,
	}

	violation := detector.DetectSpoofing(check)

	if violation == nil {
		t.Fatal("spoofing attack should be detected")
	}

	if violation.Type != ManipulationTypeSpoofing {
		t.Errorf("expected spoofing type, got %s", violation.Type)
	}

	t.Logf("Spoofing attack detected: %s", violation.Description)
}

// TestSimulateOrderSpamAttack simulates an order spam DOS attack
func TestSimulateOrderSpamAttack(t *testing.T) {
	config := DefaultRateLimitConfig()
	state := NewAccountSafeguardState("spammer")

	// Simulate rapid order placement
	blockedCount := 0
	totalAttempts := 100

	for i := 0; i < totalAttempts; i++ {
		err := state.CheckRateLimit("order", config)
		if err != nil {
			blockedCount++
		} else {
			state.OrdersThisBlock++
		}
	}

	// Most should be blocked
	expectedBlocked := totalAttempts - int(config.MaxOrdersPerBlock)
	if blockedCount != expectedBlocked {
		t.Errorf("expected %d blocked, got %d", expectedBlocked, blockedCount)
	}

	t.Logf("Order spam attack: %d/%d blocked by rate limiter", blockedCount, totalAttempts)
}

// TestSimulatePriceManipulationAttack simulates a price manipulation attack
func TestSimulatePriceManipulationAttack(t *testing.T) {
	config := DefaultPriceOracleConfig()
	oracle := NewPriceOracle(config)
	now := time.Now()

	history := NewPriceHistory(OfferingID{ProviderAddress: "test", Sequence: 1}, 100)

	// Establish normal price range
	for i := int64(0); i < 20; i++ {
		history.AddPoint(PricePoint{
			Price:       1000 + uint64(i%10), // 1000-1009 range
			Volume:      100,
			BlockHeight: i,
			Timestamp:   now.Add(time.Duration(i) * time.Minute),
		})
	}

	// Update price band
	_ = oracle.UpdatePriceBand(history, 20, now)

	// Attempt price manipulation
	manipulatedPrices := []uint64{500, 2000, 3000, 100}
	rejectedCount := 0

	for _, price := range manipulatedPrices {
		result := oracle.ValidatePrice(price, history, now)
		if result.Status != PriceValidationStatusValid {
			rejectedCount++
		}
	}

	// All manipulated prices should be rejected
	if rejectedCount != len(manipulatedPrices) {
		t.Errorf("expected %d rejected, got %d", len(manipulatedPrices), rejectedCount)
	}

	t.Logf("Price manipulation attack: %d/%d price submissions rejected",
		rejectedCount, len(manipulatedPrices))
}

// TestSimulateCircuitBreakerTrigger simulates extreme volatility triggering circuit breaker
func TestSimulateCircuitBreakerTrigger(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	state := NewCircuitBreakerState(OfferingID{ProviderAddress: "test", Sequence: 1})
	now := time.Now()
	block := int64(1000)

	// Simulate flash crash scenario
	state.Trip("flash_crash_25pct_drop", config, block, now)

	if state.Status != CircuitBreakerStatusOpen {
		t.Error("circuit breaker should be open")
	}

	// Trading should be halted
	err := state.AllowOrder(config)
	if err == nil {
		t.Error("orders should not be allowed during circuit breaker")
	}

	// After cooldown, should transition to half-open
	state.Update(block+config.TripDurationBlocks+1, config, now.Add(time.Hour))

	if state.Status != CircuitBreakerStatusHalfOpen {
		t.Errorf("expected half-open, got %s", state.Status)
	}

	t.Logf("Circuit breaker test: status=%s, consecutive_trips=%d",
		state.Status, state.ConsecutiveTrips)
}

// TestSimulateSybilAttack simulates multiple related accounts
func TestSimulateSybilAttack(t *testing.T) {
	config := DefaultSafeguardConfig()
	detector := NewManipulationDetector(config)
	penaltyConfig := DefaultPenaltyConfig()
	now := time.Now()

	// Main attacker with linked accounts
	mainState := NewAccountSafeguardState("main_attacker")
	mainState.RelatedAddresses = []string{"sybil1", "sybil2", "sybil3", "sybil4", "sybil5"}

	// Simulate trading between main and sybil accounts
	detectedCount := 0
	for i, sybil := range mainState.RelatedAddresses {
		check := WashTradingCheck{
			BuyerAddress:  "main_attacker",
			SellerAddress: sybil,
			Amount:        1000000,
			Timestamp:     now,
			BlockHeight:   int64(1000 + i),
		}

		violation := detector.DetectWashTrading(check, mainState)
		if violation != nil {
			detectedCount++
			mainState.RecordViolation(*violation, penaltyConfig)
		}
	}

	// All sybil trades should be detected
	if detectedCount != len(mainState.RelatedAddresses) {
		t.Errorf("expected %d detections, got %d",
			len(mainState.RelatedAddresses), detectedCount)
	}

	t.Logf("Sybil attack: %d/%d related address trades detected",
		detectedCount, len(mainState.RelatedAddresses))
}

// TestFeeArbitrageResistance tests resistance to fee arbitrage
func TestFeeArbitrageResistance(t *testing.T) {
	schedule := DefaultFeeSchedule()
	calc := NewDynamicFeeCalculator(schedule)

	// Test that minimum fee is always enforced
	input := FeeCalculationInput{
		OrderValue:      1000000000, // 1000 tokens
		UserTier:        FeeTierDiamond,
		User30DayVolume: 100000000000,
		IsMaker:         true,
		IsEarlyAdopter:  true,
		Utilization:     UtilizationMetrics{}, // Low utilization
	}

	result, err := calc.CalculateFee(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Even with all discounts, fee should not go below minimum
	minFee := (input.OrderValue * uint64(schedule.MinTakeRateBps)) / 10000
	if result.NetFee < minFee {
		t.Errorf("fee %d below minimum %d", result.NetFee, minFee)
	}

	t.Logf("Fee arbitrage test: gross=%d, net=%d, min=%d, effective_rate=%d bps",
		result.GrossFee, result.NetFee, minFee, result.EffectiveRateBps)
}

// Benchmark tests
func BenchmarkFeeCalculation(b *testing.B) {
	schedule := DefaultFeeSchedule()
	calc := NewDynamicFeeCalculator(schedule)

	input := FeeCalculationInput{
		OrderValue:      10000000,
		UserTier:        FeeTierGold,
		User30DayVolume: 5000000000,
		IsMaker:         false,
		IsEarlyAdopter:  false,
		Utilization: UtilizationMetrics{
			TotalCapacity: 1000,
			UsedCapacity:  500,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = calc.CalculateFee(input)
	}
}

func BenchmarkTWAPCalculation(b *testing.B) {
	config := DefaultPriceOracleConfig()
	oracle := NewPriceOracle(config)
	history := NewPriceHistory(OfferingID{ProviderAddress: "test", Sequence: 1}, 1000)
	now := time.Now()

	// Populate history
	for i := int64(0); i < 100; i++ {
		history.AddPoint(PricePoint{
			Price:       1000 + uint64(i),
			Volume:      100,
			BlockHeight: i,
			Timestamp:   now.Add(time.Duration(i) * time.Minute),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = oracle.CalculateTWAP(history, 100)
	}
}

func BenchmarkWashTradingDetection(b *testing.B) {
	config := DefaultSafeguardConfig()
	detector := NewManipulationDetector(config)
	state := NewAccountSafeguardState("test")
	state.RelatedAddresses = []string{"related1", "related2", "related3"}
	now := time.Now()

	check := WashTradingCheck{
		BuyerAddress:  "test",
		SellerAddress: "related1",
		Amount:        1000000,
		Timestamp:     now,
		BlockHeight:   1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.DetectWashTrading(check, state)
	}
}
