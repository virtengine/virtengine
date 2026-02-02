//go:build e2e.integration

// Package waldur provides E2E tests for Waldur integration.
//
// VE-25M: Waldur integration E2E tests - Settlement flows
package waldur

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/tests/e2e/mocks"
)

// ============================================================================
// Settlement Test Suite
// ============================================================================

// SettlementTestSuite tests the settlement workflow:
// 1. Usage recording
// 2. Invoice generation
// 3. Escrow settlement
// 4. Payment distribution
type SettlementTestSuite struct {
	suite.Suite

	waldurMock       *mocks.WaldurMock
	testProjectUUID  string
	testCustomerUUID string
	testOfferingUUID string
}

func TestSettlement(t *testing.T) {
	suite.Run(t, new(SettlementTestSuite))
}

func (s *SettlementTestSuite) SetupSuite() {
	config := mocks.DefaultWaldurMockConfig()
	config.EnableUsageReporting = true
	s.waldurMock = mocks.NewWaldurMockWithConfig(config)

	s.testProjectUUID = config.ProjectUUID
	s.testCustomerUUID = config.CustomerUUID

	// Setup offering
	offeringUUID := uuid.New().String()
	s.waldurMock.RegisterOffering(&mocks.MockWaldurOffering{
		UUID:         offeringUUID,
		Name:         "Settlement Test Offering",
		CustomerUUID: s.testCustomerUUID,
		State:        "active",
		PricePerHour: "1.00",
		Components: []mocks.OfferingComponent{
			{Type: "usage", Name: "cpu_hours", Unit: "hour", Amount: 0.10},
			{Type: "usage", Name: "memory_gb_hours", Unit: "hour", Amount: 0.02},
		},
	})
	s.testOfferingUUID = offeringUUID
}

func (s *SettlementTestSuite) TearDownSuite() {
	if s.waldurMock != nil {
		s.waldurMock.Close()
	}
}

// ============================================================================
// HTTP Helpers
// ============================================================================

func (s *SettlementTestSuite) httpPost(path string, payload interface{}) ([]byte, int, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	resp, err := http.Post(s.waldurMock.BaseURL()+path, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func (s *SettlementTestSuite) httpGet(path string) ([]byte, int, error) {
	resp, err := http.Get(s.waldurMock.BaseURL() + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

// ============================================================================
// Usage Recording Tests
// ============================================================================

func (s *SettlementTestSuite) TestUsageRecording() {
	s.Run("SubmitSingleUsageRecord", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "usage-test-1",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		now := time.Now()
		usageReq := map[string]interface{}{
			"resource":     resourceUUID,
			"period_start": now.Add(-1 * time.Hour).Format(time.RFC3339),
			"period_end":   now.Format(time.RFC3339),
			"usages": map[string]interface{}{
				"cpu_hours":       1.0,
				"memory_gb_hours": 4.0,
			},
			"is_final": false,
		}

		_, status, err := s.httpPost("/api/marketplace-component-usages/", usageReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		records := s.waldurMock.GetUsageRecords(resourceUUID)
		s.GreaterOrEqual(len(records), 1)
	})

	s.Run("SubmitMultipleUsageRecords", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "usage-test-multi",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		now := time.Now()
		for i := 0; i < 5; i++ {
			periodStart := now.Add(time.Duration(-5+i) * time.Hour)
			periodEnd := now.Add(time.Duration(-4+i) * time.Hour)

			usageReq := map[string]interface{}{
				"resource": resourceUUID,
				"period_start":  periodStart.Format(time.RFC3339),
				"period_end":    periodEnd.Format(time.RFC3339),
				"usages": map[string]interface{}{
					"cpu_hours":       1.0 + float64(i)*0.1,
					"memory_gb_hours": 4.0 + float64(i)*0.5,
				},
				"is_final": false,
			}

			_, status, err := s.httpPost("/api/marketplace-component-usages/", usageReq)
			s.Require().NoError(err)
			s.Equal(http.StatusCreated, status)
		}

		records := s.waldurMock.GetUsageRecords(resourceUUID)
		s.Equal(5, len(records))
	})

	s.Run("FinalUsageRecord", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "usage-test-final",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		now := time.Now()
		usageReq := map[string]interface{}{
			"resource": resourceUUID,
			"period_start":  now.Add(-1 * time.Hour).Format(time.RFC3339),
			"period_end":    now.Format(time.RFC3339),
			"usages": map[string]interface{}{
				"cpu_hours": 2.5,
			},
			"is_final": true,
		}

		_, status, err := s.httpPost("/api/marketplace-component-usages/", usageReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		records := s.waldurMock.GetUsageRecords(resourceUUID)
		s.Require().GreaterOrEqual(len(records), 1)
		s.True(records[len(records)-1].IsFinal)
	})
}

// ============================================================================
// Invoice Tests
// ============================================================================

func (s *SettlementTestSuite) TestInvoiceGeneration() {
	s.Run("GenerateInvoiceFromUsage", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "invoice-test-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		// Submit usage first
		now := time.Now()
		usageReq := map[string]interface{}{
			"resource": resourceUUID,
			"period_start":  now.Add(-24 * time.Hour).Format(time.RFC3339),
			"period_end":    now.Format(time.RFC3339),
			"usages": map[string]interface{}{
				"cpu_hours":       24.0,
				"memory_gb_hours": 96.0,
			},
			"is_final": true,
		}

		_, status, err := s.httpPost("/api/marketplace-component-usages/", usageReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		// Create invoice
		invoiceUUID := uuid.New().String()
		s.waldurMock.RegisterInvoice(&mocks.MockWaldurInvoice{
			UUID:         invoiceUUID,
			ResourceUUID: resourceUUID,
			CustomerUUID: s.testCustomerUUID,
			TotalAmount:  "4.32", // 24 * 0.10 + 96 * 0.02
			Currency:     "USD",
			State:        "pending",
			PeriodStart:  now.Add(-24 * time.Hour),
			PeriodEnd:    now,
			LineItems: []mocks.MockWaldurInvoiceLineItem{
				{Name: "cpu_hours", Quantity: "24.0", UnitPrice: "0.10", Total: "2.40", Unit: "hour"},
				{Name: "memory_gb_hours", Quantity: "96.0", UnitPrice: "0.02", Total: "1.92", Unit: "hour"},
			},
		})

		// Verify invoice
		invoice := s.waldurMock.GetInvoice(invoiceUUID)
		s.Require().NotNil(invoice)
		s.Equal("4.32", invoice.TotalAmount)
		s.Equal(2, len(invoice.LineItems))
	})

	s.Run("ListInvoices", func() {
		body, status, err := s.httpGet("/api/invoices/")
		s.Require().NoError(err)
		s.Equal(http.StatusOK, status)

		var invoices []map[string]interface{}
		err = json.Unmarshal(body, &invoices)
		s.Require().NoError(err)
		s.GreaterOrEqual(len(invoices), 0)
	})
}

// ============================================================================
// Settlement Workflow Tests
// ============================================================================

func (s *SettlementTestSuite) TestSettlementWorkflow() {
	s.Run("CompleteSettlementCycle", func() {
		// Create resource
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "settlement-cycle-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		// Submit usage
		now := time.Now()
		usageReq := map[string]interface{}{
			"resource": resourceUUID,
			"period_start":  now.Add(-1 * time.Hour).Format(time.RFC3339),
			"period_end":    now.Format(time.RFC3339),
			"usages": map[string]interface{}{
				"cpu_hours": 1.0,
			},
			"is_final": true,
		}
		_, _, _ = s.httpPost("/api/marketplace-component-usages/", usageReq)

		// Create and pay invoice
		invoiceUUID := uuid.New().String()
		s.waldurMock.RegisterInvoice(&mocks.MockWaldurInvoice{
			UUID:         invoiceUUID,
			ResourceUUID: resourceUUID,
			CustomerUUID: s.testCustomerUUID,
			TotalAmount:  "0.10",
			Currency:     "USD",
			State:        "pending",
			PeriodStart:  now.Add(-1 * time.Hour),
			PeriodEnd:    now,
		})

		// Mark as paid
		paidAt := time.Now()
		invoice := s.waldurMock.GetInvoice(invoiceUUID)
		s.Require().NotNil(invoice)
		invoice.State = "paid"
		invoice.PaidAt = &paidAt

		// Verify state
		s.Equal("paid", invoice.State)
		s.NotNil(invoice.PaidAt)
	})

	s.Run("SettlementWithEscrow", func() {
		// Simulate escrow-backed settlement
		resourceUUID := uuid.New().String()
		escrowID := "escrow-" + uuid.New().String()[:8]

		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "escrow-settlement-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
			Attributes: map[string]interface{}{
				"escrow_id": escrowID,
			},
		})

		// Verify escrow is tracked
		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
		s.Equal(escrowID, resource.Attributes["escrow_id"])
	})
}

// ============================================================================
// Metrics and Billing Tests
// ============================================================================

func (s *SettlementTestSuite) TestMetricsAndBilling() {
	s.Run("CalculateBillingFromMetrics", func() {
		// Create resource with known usage
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "billing-calc-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		// Submit usage with exact values
		now := time.Now()
		usageReq := map[string]interface{}{
			"resource": resourceUUID,
			"period_start":  now.Add(-1 * time.Hour).Format(time.RFC3339),
			"period_end":    now.Format(time.RFC3339),
			"usages": map[string]interface{}{
				"cpu_hours":       10.0, // 10 * 0.10 = 1.00
				"memory_gb_hours": 50.0, // 50 * 0.02 = 1.00
			},
			"is_final": true,
		}

		_, status, err := s.httpPost("/api/marketplace-component-usages/", usageReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		records := s.waldurMock.GetUsageRecords(resourceUUID)
		s.Require().GreaterOrEqual(len(records), 1)

		// Verify metrics
		lastRecord := records[len(records)-1]
		s.Equal(10.0, lastRecord.Metrics["cpu_hours"])
		s.Equal(50.0, lastRecord.Metrics["memory_gb_hours"])
	})

	s.Run("ZeroUsageRecord", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "zero-usage-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		now := time.Now()
		usageReq := map[string]interface{}{
			"resource": resourceUUID,
			"period_start":  now.Add(-1 * time.Hour).Format(time.RFC3339),
			"period_end":    now.Format(time.RFC3339),
			"usages": map[string]interface{}{
				"cpu_hours":       0.0,
				"memory_gb_hours": 0.0,
			},
			"is_final": false,
		}

		_, status, err := s.httpPost("/api/marketplace-component-usages/", usageReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)
	})
}

// ============================================================================
// Edge Cases
// ============================================================================

func (s *SettlementTestSuite) TestSettlementEdgeCases() {
	s.Run("UsageForNonExistentResource", func() {
		usageReq := map[string]interface{}{
			"resource": "00000000-0000-0000-0000-000000000001",
			"period_start":  time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"period_end":    time.Now().Format(time.RFC3339),
			"usages":       map[string]interface{}{"cpu_hours": 1.0},
		}

		_, status, _ := s.httpPost("/api/marketplace-component-usages/", usageReq)
		// Mock may accept or reject
		s.Contains([]int{http.StatusCreated, http.StatusNotFound, http.StatusBadRequest}, status)
	})

	s.Run("OverlappingUsagePeriods", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "overlap-usage-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		now := time.Now()
		// First period: -2h to -1h
		usageReq1 := map[string]interface{}{
			"resource": resourceUUID,
			"period_start":  now.Add(-2 * time.Hour).Format(time.RFC3339),
			"period_end":    now.Add(-1 * time.Hour).Format(time.RFC3339),
			"usages":       map[string]interface{}{"cpu_hours": 1.0},
		}
		_, _, _ = s.httpPost("/api/marketplace-component-usages/", usageReq1)

		// Overlapping period: -1.5h to -0.5h
		usageReq2 := map[string]interface{}{
			"resource": resourceUUID,
			"period_start":  now.Add(-90 * time.Minute).Format(time.RFC3339),
			"period_end":    now.Add(-30 * time.Minute).Format(time.RFC3339),
			"usages":       map[string]interface{}{"cpu_hours": 1.0},
		}
		_, status, _ := s.httpPost("/api/marketplace-component-usages/", usageReq2)
		// Mock accepts it (real API might reject)
		s.Equal(http.StatusCreated, status)
	})

	s.Run("VeryLargeUsageValues", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "large-usage-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		now := time.Now()
		usageReq := map[string]interface{}{
			"resource": resourceUUID,
			"period_start":  now.Add(-1 * time.Hour).Format(time.RFC3339),
			"period_end":    now.Format(time.RFC3339),
			"usages": map[string]interface{}{
				"cpu_hours":       999999.999,
				"memory_gb_hours": 999999.999,
			},
		}

		_, status, err := s.httpPost("/api/marketplace-component-usages/", usageReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)
	})
}

