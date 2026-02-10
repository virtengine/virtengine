//go:build e2e.integration

// Package fixtures provides test fixtures for E2E tests.
//
// VE-15C: Test fixtures for provider E2E flow tests.
package fixtures

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
)

// =============================================================================
// Provider Offering Fixtures
// =============================================================================

// TestOffering represents a marketplace offering for E2E tests.
type TestOffering struct {
	OfferingID      string
	Name            string
	Description     string
	Category        string
	ProviderAddress string
	WaldurUUID      string

	// Pricing
	PricePerHour     sdkmath.LegacyDec
	Currency         string
	MinimumDeposit   sdk.Coin
	SettlementPeriod time.Duration

	// Resources
	CPUCores  int32
	MemoryGB  int32
	StorageGB int32
	GPUs      int32
	Region    string

	// Requirements
	MinVEIDScore         uint32
	RequireVerifiedEmail bool
	RequireMFA           bool

	// State
	Active    bool
	CreatedAt time.Time
}

// DefaultTestOffering returns a default test offering.
func DefaultTestOffering(providerAddr string) TestOffering {
	return TestOffering{
		OfferingID:       fmt.Sprintf("offering-e2e-%d", time.Now().UnixNano()%10000),
		Name:             "E2E Compute Standard",
		Description:      "Standard compute offering for E2E tests",
		Category:         "compute",
		ProviderAddress:  providerAddr,
		WaldurUUID:       uuid.NewString(),
		PricePerHour:     sdkmath.LegacyNewDec(10),
		Currency:         "uve",
		MinimumDeposit:   sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
		SettlementPeriod: time.Hour,
		CPUCores:         8,
		MemoryGB:         32,
		StorageGB:        100,
		GPUs:             0,
		Region:           "us-east",
		MinVEIDScore:     70,
		Active:           true,
		CreatedAt:        time.Now().UTC(),
	}
}

// ComputeSmallOffering returns a small compute offering.
func ComputeSmallOffering(providerAddr string) TestOffering {
	return TestOffering{
		OfferingID:       "offering-compute-small",
		Name:             "Compute Small",
		Description:      "Small compute instance with 4 cores",
		Category:         "compute",
		ProviderAddress:  providerAddr,
		WaldurUUID:       uuid.NewString(),
		PricePerHour:     sdkmath.LegacyNewDec(5),
		Currency:         "uve",
		MinimumDeposit:   sdk.NewCoin("uve", sdkmath.NewInt(500000)),
		SettlementPeriod: time.Hour,
		CPUCores:         4,
		MemoryGB:         16,
		StorageGB:        50,
		GPUs:             0,
		Region:           "us-east",
		MinVEIDScore:     50,
		Active:           true,
		CreatedAt:        time.Now().UTC(),
	}
}

// ComputeMediumOffering returns a medium compute offering.
func ComputeMediumOffering(providerAddr string) TestOffering {
	return TestOffering{
		OfferingID:       "offering-compute-medium",
		Name:             "Compute Medium",
		Description:      "Medium compute instance with 16 cores",
		Category:         "compute",
		ProviderAddress:  providerAddr,
		WaldurUUID:       uuid.NewString(),
		PricePerHour:     sdkmath.LegacyNewDec(20),
		Currency:         "uve",
		MinimumDeposit:   sdk.NewCoin("uve", sdkmath.NewInt(2000000)),
		SettlementPeriod: time.Hour,
		CPUCores:         16,
		MemoryGB:         64,
		StorageGB:        200,
		GPUs:             0,
		Region:           "us-east",
		MinVEIDScore:     70,
		Active:           true,
		CreatedAt:        time.Now().UTC(),
	}
}

// GPUOffering returns a GPU compute offering.
func GPUOffering(providerAddr string) TestOffering {
	return TestOffering{
		OfferingID:       "offering-gpu-standard",
		Name:             "GPU Standard",
		Description:      "GPU compute instance with 2 NVIDIA A100s",
		Category:         "gpu",
		ProviderAddress:  providerAddr,
		WaldurUUID:       uuid.NewString(),
		PricePerHour:     sdkmath.LegacyNewDec(100),
		Currency:         "uve",
		MinimumDeposit:   sdk.NewCoin("uve", sdkmath.NewInt(10000000)),
		SettlementPeriod: time.Hour,
		CPUCores:         32,
		MemoryGB:         128,
		StorageGB:        500,
		GPUs:             2,
		Region:           "us-west",
		MinVEIDScore:     80,
		RequireMFA:       true,
		Active:           true,
		CreatedAt:        time.Now().UTC(),
	}
}

// StorageOffering returns a storage offering.
func StorageOffering(providerAddr string) TestOffering {
	return TestOffering{
		OfferingID:       "offering-storage-block",
		Name:             "Block Storage",
		Description:      "High-performance block storage",
		Category:         "storage",
		ProviderAddress:  providerAddr,
		WaldurUUID:       uuid.NewString(),
		PricePerHour:     sdkmath.LegacyMustNewDecFromStr("0.5"),
		Currency:         "uve",
		MinimumDeposit:   sdk.NewCoin("uve", sdkmath.NewInt(100000)),
		SettlementPeriod: time.Hour * 24,
		CPUCores:         0,
		MemoryGB:         0,
		StorageGB:        1000,
		GPUs:             0,
		Region:           "us-east",
		MinVEIDScore:     50,
		Active:           true,
		CreatedAt:        time.Now().UTC(),
	}
}

// =============================================================================
// Provider Registration Fixtures
// =============================================================================

// TestProvider represents a provider configuration for E2E tests.
type TestProvider struct {
	Address      string
	Name         string
	Description  string
	Email        string
	Website      string
	Regions      []string
	Capabilities []string
	VEIDScore    uint32
	Active       bool
}

// DefaultTestProvider returns a default test provider.
func DefaultTestProvider(addr string) TestProvider {
	return TestProvider{
		Address:      addr,
		Name:         "E2E Test Provider",
		Description:  "Provider for E2E testing",
		Email:        "e2e-provider@virtengine.test",
		Website:      "https://e2e-provider.virtengine.test",
		Regions:      []string{"us-east", "us-west", "eu-west"},
		Capabilities: []string{"compute", "storage", "gpu"},
		VEIDScore:    85,
		Active:       true,
	}
}

// HighCapacityProvider returns a high-capacity provider.
func HighCapacityProvider(addr string) TestProvider {
	return TestProvider{
		Address:      addr,
		Name:         "High Capacity Provider",
		Description:  "Provider with large resource pool",
		Email:        "hpc@virtengine.test",
		Website:      "https://hpc.virtengine.test",
		Regions:      []string{"us-east", "us-west", "eu-west", "ap-south"},
		Capabilities: []string{"compute", "storage", "gpu", "hpc"},
		VEIDScore:    95,
		Active:       true,
	}
}

// =============================================================================
// Order and Bid Fixtures
// =============================================================================

// TestOrder represents a marketplace order for E2E tests.
type TestOrder struct {
	OrderID       string
	CustomerAddr  string
	OfferingID    string
	MaxPrice      sdkmath.LegacyDec
	Deposit       sdk.Coin
	Duration      time.Duration
	Status        string
	CreatedAt     time.Time
	ClosedAt      *time.Time
	AllocationID  string
	LeaseID       string
	ResourceSpecs ResourceSpecs
}

// ResourceSpecs defines requested resource specifications.
type ResourceSpecs struct {
	CPUCores  int32
	MemoryGB  int32
	StorageGB int32
	GPUs      int32
}

// DefaultTestOrder returns a default test order.
func DefaultTestOrder(customerAddr, offeringID string) TestOrder {
	return TestOrder{
		OrderID:      fmt.Sprintf("order-e2e-%d", time.Now().UnixNano()%10000),
		CustomerAddr: customerAddr,
		OfferingID:   offeringID,
		MaxPrice:     sdkmath.LegacyNewDec(15),
		Deposit:      sdk.NewCoin("uve", sdkmath.NewInt(5000000)),
		Duration:     time.Hour * 24,
		Status:       "open",
		CreatedAt:    time.Now().UTC(),
		ResourceSpecs: ResourceSpecs{
			CPUCores:  4,
			MemoryGB:  16,
			StorageGB: 50,
			GPUs:      0,
		},
	}
}

// TestBid represents a provider bid for E2E tests.
type TestBid struct {
	BidID        string
	OrderID      string
	ProviderAddr string
	Price        sdkmath.LegacyDec
	Deposit      sdk.Coin
	Status       string
	CreatedAt    time.Time
}

// DefaultTestBid returns a default test bid.
func DefaultTestBid(orderID, providerAddr string) TestBid {
	return TestBid{
		BidID:        fmt.Sprintf("bid-e2e-%d", time.Now().UnixNano()%10000),
		OrderID:      orderID,
		ProviderAddr: providerAddr,
		Price:        sdkmath.LegacyNewDec(12),
		Deposit:      sdk.NewCoin("uve", sdkmath.NewInt(1000000)),
		Status:       "open",
		CreatedAt:    time.Now().UTC(),
	}
}

// =============================================================================
// Allocation and Lease Fixtures
// =============================================================================

// TestAllocation represents a resource allocation for E2E tests.
type TestAllocation struct {
	AllocationID    string
	OrderID         string
	BidID           string
	ProviderAddr    string
	CustomerAddr    string
	Status          string
	WaldurOrderUUID string
	WaldurResUUID   string
	CreatedAt       time.Time
	ProvisionedAt   *time.Time
	TerminatedAt    *time.Time
}

// DefaultTestAllocation returns a default test allocation.
func DefaultTestAllocation(orderID, bidID, providerAddr, customerAddr string) TestAllocation {
	return TestAllocation{
		AllocationID: fmt.Sprintf("alloc-e2e-%d", time.Now().UnixNano()%10000),
		OrderID:      orderID,
		BidID:        bidID,
		ProviderAddr: providerAddr,
		CustomerAddr: customerAddr,
		Status:       "provisioning",
		CreatedAt:    time.Now().UTC(),
	}
}

// TestLease represents a lease for E2E tests.
type TestLease struct {
	LeaseID      string
	OrderID      string
	AllocationID string
	ProviderAddr string
	CustomerAddr string
	Price        sdkmath.LegacyDec
	Status       string
	StartTime    time.Time
	EndTime      *time.Time
}

// DefaultTestLease returns a default test lease.
func DefaultTestLease(orderID, allocationID, providerAddr, customerAddr string) TestLease {
	return TestLease{
		LeaseID:      fmt.Sprintf("lease-e2e-%d", time.Now().UnixNano()%10000),
		OrderID:      orderID,
		AllocationID: allocationID,
		ProviderAddr: providerAddr,
		CustomerAddr: customerAddr,
		Price:        sdkmath.LegacyNewDec(12),
		Status:       "active",
		StartTime:    time.Now().UTC(),
	}
}

// =============================================================================
// Usage and Settlement Fixtures
// =============================================================================

// TestUsageRecord represents a usage record for E2E tests.
type TestUsageRecord struct {
	RecordID       string
	AllocationID   string
	ProviderAddr   string
	CustomerAddr   string
	PeriodStart    time.Time
	PeriodEnd      time.Time
	IsFinal        bool
	Metrics        UsageMetrics
	BilledAmount   sdkmath.LegacyDec
	SettlementHash string
}

// UsageMetrics contains resource usage metrics.
type UsageMetrics struct {
	CPUMilliSeconds    int64
	MemoryByteSeconds  int64
	StorageByteSeconds int64
	NetworkBytesIn     int64
	NetworkBytesOut    int64
	GPUSeconds         int64
}

// DefaultTestUsageRecord returns a default usage record for 1 hour.
func DefaultTestUsageRecord(allocationID, providerAddr, customerAddr string) TestUsageRecord {
	periodStart := time.Now().Add(-time.Hour).UTC()
	periodEnd := time.Now().UTC()

	return TestUsageRecord{
		RecordID:     fmt.Sprintf("usage-e2e-%d", time.Now().UnixNano()%10000),
		AllocationID: allocationID,
		ProviderAddr: providerAddr,
		CustomerAddr: customerAddr,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		IsFinal:      false,
		Metrics: UsageMetrics{
			CPUMilliSeconds:    4 * 3600 * 1000,                // 4 cores for 1 hour
			MemoryByteSeconds:  16 * 1024 * 1024 * 1024 * 3600, // 16 GB for 1 hour
			StorageByteSeconds: 50 * 1024 * 1024 * 1024 * 3600, // 50 GB for 1 hour
			NetworkBytesIn:     100 * 1024 * 1024,              // 100 MB in
			NetworkBytesOut:    50 * 1024 * 1024,               // 50 MB out
			GPUSeconds:         0,
		},
		BilledAmount: sdkmath.LegacyNewDec(12),
	}
}

// FinalTestUsageRecord returns a final usage record (for termination).
func FinalTestUsageRecord(allocationID, providerAddr, customerAddr string, duration time.Duration) TestUsageRecord {
	periodStart := time.Now().Add(-duration).UTC()
	periodEnd := time.Now().UTC()
	hours := float64(duration) / float64(time.Hour)

	return TestUsageRecord{
		RecordID:     fmt.Sprintf("usage-final-e2e-%d", time.Now().UnixNano()%10000),
		AllocationID: allocationID,
		ProviderAddr: providerAddr,
		CustomerAddr: customerAddr,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		IsFinal:      true,
		Metrics: UsageMetrics{
			CPUMilliSeconds:    int64(4 * hours * 3600 * 1000),
			MemoryByteSeconds:  int64(16 * 1024 * 1024 * 1024 * hours * 3600),
			StorageByteSeconds: int64(50 * 1024 * 1024 * 1024 * hours * 3600),
			NetworkBytesIn:     int64(hours * 100 * 1024 * 1024),
			NetworkBytesOut:    int64(hours * 50 * 1024 * 1024),
			GPUSeconds:         0,
		},
		BilledAmount: sdkmath.LegacyNewDec(int64(12 * hours)),
	}
}

// TestSettlement represents a settlement for E2E tests.
type TestSettlement struct {
	SettlementID      string
	AllocationID      string
	ProviderAddr      string
	CustomerAddr      string
	UsageRecordIDs    []string
	TotalAmount       sdk.Coin
	ProviderPayout    sdk.Coin
	PlatformFee       sdk.Coin
	Status            string
	CreatedAt         time.Time
	SettledAt         *time.Time
	DisputedAt        *time.Time
	DisputeReason     string
	DisputeResolution string
}

// DefaultTestSettlement returns a default test settlement.
func DefaultTestSettlement(allocationID, providerAddr, customerAddr string, usageRecordIDs []string, totalAmount int64) TestSettlement {
	total := sdk.NewCoin("uve", sdkmath.NewInt(totalAmount))
	// 2.5% platform fee
	feeAmount := totalAmount * 25 / 1000
	fee := sdk.NewCoin("uve", sdkmath.NewInt(feeAmount))
	payout := sdk.NewCoin("uve", sdkmath.NewInt(totalAmount-feeAmount))

	return TestSettlement{
		SettlementID:   fmt.Sprintf("settlement-e2e-%d", time.Now().UnixNano()%10000),
		AllocationID:   allocationID,
		ProviderAddr:   providerAddr,
		CustomerAddr:   customerAddr,
		UsageRecordIDs: usageRecordIDs,
		TotalAmount:    total,
		ProviderPayout: payout,
		PlatformFee:    fee,
		Status:         "pending",
		CreatedAt:      time.Now().UTC(),
	}
}

// =============================================================================
// VEID Fixtures for Provider Registration
// =============================================================================

// ProviderVEIDFixture represents VEID data for provider registration.
type ProviderVEIDFixture struct {
	Address       string
	Score         uint32
	Tier          string
	Status        string
	ScopesCount   int
	EmailVerified bool
	MFAEnabled    bool
}

// ValidProviderVEID returns a valid VEID for provider registration (score >= 70).
func ValidProviderVEID(addr string) ProviderVEIDFixture {
	return ProviderVEIDFixture{
		Address:       addr,
		Score:         85,
		Tier:          "verified",
		Status:        "active",
		ScopesCount:   3,
		EmailVerified: true,
		MFAEnabled:    true,
	}
}

// InsufficientProviderVEID returns VEID with score < 70 (insufficient for provider).
func InsufficientProviderVEID(addr string) ProviderVEIDFixture {
	return ProviderVEIDFixture{
		Address:       addr,
		Score:         55,
		Tier:          "basic",
		Status:        "active",
		ScopesCount:   1,
		EmailVerified: true,
		MFAEnabled:    false,
	}
}

// =============================================================================
// Pricing Calculation Helpers
// =============================================================================

// CalculateBilledAmount calculates the billed amount for given metrics and rates.
func CalculateBilledAmount(metrics UsageMetrics, cpuRate, memRate, storageRate sdkmath.LegacyDec) sdkmath.LegacyDec {
	// Convert to hours
	cpuHours := sdkmath.LegacyNewDec(metrics.CPUMilliSeconds).Quo(sdkmath.LegacyNewDec(1000 * 3600))
	memGBHours := sdkmath.LegacyNewDec(metrics.MemoryByteSeconds).Quo(sdkmath.LegacyNewDec(1024 * 1024 * 1024 * 3600))
	storageGBHours := sdkmath.LegacyNewDec(metrics.StorageByteSeconds).Quo(sdkmath.LegacyNewDec(1024 * 1024 * 1024 * 3600))

	// Calculate costs
	cpuCost := cpuHours.Mul(cpuRate)
	memCost := memGBHours.Mul(memRate)
	storageCost := storageGBHours.Mul(storageRate)

	return cpuCost.Add(memCost).Add(storageCost)
}

// CalculatePlatformFee calculates the platform fee (default 2.5%).
func CalculatePlatformFee(amount sdkmath.LegacyDec, feePercent sdkmath.LegacyDec) sdkmath.LegacyDec {
	return amount.Mul(feePercent).Quo(sdkmath.LegacyNewDec(100))
}

// DefaultFeePercent returns the default platform fee percentage.
func DefaultFeePercent() sdkmath.LegacyDec {
	return sdkmath.LegacyMustNewDecFromStr("2.5")
}
