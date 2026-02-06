// Copyright 2024 The VirtEngine Authors
// This file is part of the VirtEngine library.

package keeper_test

import (
	"bytes"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/escrow/keeper"
	"github.com/virtengine/virtengine/x/escrow/types/billing"
)

// usageTestAddr generates a valid test bech32 address from a seed number
func usageTestAddr(seed int) string {
	var buffer bytes.Buffer
	buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6")
	buffer.WriteString(string(rune('0' + (seed/100)%10)))
	buffer.WriteString(string(rune('0' + (seed/10)%10)))
	buffer.WriteString(string(rune('0' + seed%10)))
	res, _ := sdk.AccAddressFromHexUnsafe(buffer.String())
	return res.String()
}

// UsagePipelineTestSuite tests the usage→invoice→settlement pipeline
type UsagePipelineTestSuite struct {
	suite.Suite
}

func TestUsagePipelineTestSuite(t *testing.T) {
	suite.Run(t, new(UsagePipelineTestSuite))
}

func (s *UsagePipelineTestSuite) SetupTest() {
	// Basic setup
}

// TestUsageReportValidation tests usage report validation logic
func (s *UsagePipelineTestSuite) TestUsageReportValidation() {
	providerAddr := usageTestAddr(100)
	customerAddr := usageTestAddr(101)
	now := time.Now()

	tests := []struct {
		name    string
		report  keeper.UsageReport
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid report",
			report: keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      "milli-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			report: keeper.UsageReport{
				Provider:    "",
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      "milli-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "invalid provider address",
			report: keeper.UsageReport{
				Provider:    "invalid-address",
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      "milli-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid provider address",
		},
		{
			name: "missing lease_id",
			report: keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      "milli-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			},
			wantErr: true,
			errMsg:  "lease_id is required",
		},
		{
			name: "missing customer",
			report: keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "lease-001",
				Customer:    "",
				EscrowID:    "escrow-001",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      "milli-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			},
			wantErr: true,
			errMsg:  "customer is required",
		},
		{
			name: "missing escrow_id",
			report: keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      "milli-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			},
			wantErr: true,
			errMsg:  "escrow_id is required",
		},
		{
			name: "empty resources",
			report: keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources:   []keeper.ResourceUsage{},
			},
			wantErr: true,
			errMsg:  "at least one resource usage entry is required",
		},
		{
			name: "period_end before period_start",
			report: keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now,
				PeriodEnd:   now.Add(-time.Hour),
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      "milli-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			},
			wantErr: true,
			errMsg:  "period_end must be after period_start",
		},
		{
			name: "equal period_start and period_end",
			report: keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now,
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      "milli-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			},
			wantErr: true,
			errMsg:  "period_end must be after period_start",
		},
		{
			name: "negative quantity",
			report: keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(-100),
						Unit:      "milli-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			},
			wantErr: true,
			errMsg:  "quantity cannot be negative",
		},
		{
			name: "missing unit",
			report: keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      "",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			},
			wantErr: true,
			errMsg:  "unit is required",
		},
		{
			name: "multiple resource types",
			report: keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      billing.UsageTypeCPU,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      "milli-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
					{
						Type:      billing.UsageTypeMemory,
						Quantity:  sdkmath.LegacyNewDec(512),
						Unit:      "byte-seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(5)),
					},
					{
						Type:      billing.UsageTypeGPU,
						Quantity:  sdkmath.LegacyNewDec(60),
						Unit:      "seconds",
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(100)),
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.report.Validate()
			if tt.wantErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.errMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestUsageSettlementResultFields tests the UsageSettlementResult type
func (s *UsagePipelineTestSuite) TestUsageSettlementResultFields() {
	result := keeper.UsageSettlementResult{
		UsageRecordIDs: []string{"usage-001", "usage-002"},
		InvoiceID:      "inv-001",
		SettlementID:   "stl-001",
		TotalAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 1000)),
		Status:         "settled",
	}

	s.Require().Len(result.UsageRecordIDs, 2)
	s.Require().Equal("inv-001", result.InvoiceID)
	s.Require().Equal("stl-001", result.SettlementID)
	s.Require().True(result.TotalAmount.AmountOf("uvirt").Equal(sdkmath.NewInt(1000)))
	s.Require().Equal("settled", result.Status)
}

// TestUsageSettlementResultPartialStatus tests partial settlement results
func (s *UsagePipelineTestSuite) TestUsageSettlementResultPartialStatus() {
	result := keeper.UsageSettlementResult{
		UsageRecordIDs: []string{"usage-001"},
		InvoiceID:      "inv-001",
		SettlementID:   "",
		TotalAmount:    sdk.NewCoins(sdk.NewInt64Coin("uvirt", 500)),
		Status:         "invoice_created",
	}

	s.Require().Equal("invoice_created", result.Status)
	s.Require().Empty(result.SettlementID)
}

// TestResourceUsageTypes tests all resource usage types are valid
func (s *UsagePipelineTestSuite) TestResourceUsageTypes() {
	providerAddr := usageTestAddr(100)
	customerAddr := usageTestAddr(101)
	now := time.Now()

	usageTypes := []billing.UsageType{
		billing.UsageTypeCPU,
		billing.UsageTypeMemory,
		billing.UsageTypeStorage,
		billing.UsageTypeNetwork,
		billing.UsageTypeGPU,
	}

	for _, ut := range usageTypes {
		s.Run(ut.String(), func() {
			report := keeper.UsageReport{
				Provider:    providerAddr,
				LeaseID:     "lease-001",
				Customer:    customerAddr,
				EscrowID:    "escrow-001",
				PeriodStart: now.Add(-time.Hour),
				PeriodEnd:   now,
				Resources: []keeper.ResourceUsage{
					{
						Type:      ut,
						Quantity:  sdkmath.LegacyNewDec(100),
						Unit:      billing.UnitForUsageType(ut),
						UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
					},
				},
			}
			s.Require().NoError(report.Validate())
		})
	}
}

// TestUsageReportMultipleResourcesCalculation tests that multiple resources
// in a report would produce correct total amounts
func (s *UsagePipelineTestSuite) TestUsageReportMultipleResourcesCalculation() {
	providerAddr := usageTestAddr(100)
	customerAddr := usageTestAddr(101)
	now := time.Now()

	report := keeper.UsageReport{
		Provider:    providerAddr,
		LeaseID:     "lease-001",
		Customer:    customerAddr,
		EscrowID:    "escrow-001",
		PeriodStart: now.Add(-time.Hour),
		PeriodEnd:   now,
		Resources: []keeper.ResourceUsage{
			{
				Type:      billing.UsageTypeCPU,
				Quantity:  sdkmath.LegacyNewDec(100), // 100 * 10 = 1000
				Unit:      "milli-seconds",
				UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
			},
			{
				Type:      billing.UsageTypeMemory,
				Quantity:  sdkmath.LegacyNewDec(200), // 200 * 5 = 1000
				Unit:      "byte-seconds",
				UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(5)),
			},
			{
				Type:      billing.UsageTypeGPU,
				Quantity:  sdkmath.LegacyNewDec(10), // 10 * 100 = 1000
				Unit:      "seconds",
				UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(100)),
			},
		},
	}

	s.Require().NoError(report.Validate())

	// Verify expected total: 1000 + 1000 + 1000 = 3000 uvirt
	expectedTotal := sdkmath.NewInt(3000)
	actualTotal := sdkmath.ZeroInt()
	for _, res := range report.Resources {
		amount := res.Quantity.Mul(res.UnitPrice.Amount).TruncateInt()
		actualTotal = actualTotal.Add(amount)
	}
	s.Require().True(expectedTotal.Equal(actualTotal), "expected %s, got %s", expectedTotal, actualTotal)
}

// TestUsageReportZeroQuantity tests that zero quantity resources are valid
func (s *UsagePipelineTestSuite) TestUsageReportZeroQuantity() {
	providerAddr := usageTestAddr(100)
	customerAddr := usageTestAddr(101)
	now := time.Now()

	report := keeper.UsageReport{
		Provider:    providerAddr,
		LeaseID:     "lease-001",
		Customer:    customerAddr,
		EscrowID:    "escrow-001",
		PeriodStart: now.Add(-time.Hour),
		PeriodEnd:   now,
		Resources: []keeper.ResourceUsage{
			{
				Type:      billing.UsageTypeCPU,
				Quantity:  sdkmath.LegacyZeroDec(),
				Unit:      "milli-seconds",
				UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
			},
		},
	}
	s.Require().NoError(report.Validate())
}

// TestInvalidCustomerAddress tests that invalid customer address is rejected
func (s *UsagePipelineTestSuite) TestInvalidCustomerAddress() {
	providerAddr := usageTestAddr(100)
	now := time.Now()

	report := keeper.UsageReport{
		Provider:    providerAddr,
		LeaseID:     "lease-001",
		Customer:    "invalid-customer",
		EscrowID:    "escrow-001",
		PeriodStart: now.Add(-time.Hour),
		PeriodEnd:   now,
		Resources: []keeper.ResourceUsage{
			{
				Type:      billing.UsageTypeCPU,
				Quantity:  sdkmath.LegacyNewDec(100),
				Unit:      "milli-seconds",
				UnitPrice: sdk.NewDecCoin("uvirt", sdkmath.NewInt(10)),
			},
		},
	}

	err := report.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "invalid customer address")
}
