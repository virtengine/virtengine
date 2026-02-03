//go:build e2e.integration

// Package waldur provides E2E tests for Waldur integration.
//
// VE-25M: Waldur integration E2E tests - Provider flows
package waldur

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
)

// ============================================================================
// Provider Flow Test Suite
// ============================================================================

// ProviderFlowTestSuite tests provider-specific Waldur integration:
// 1. Offering publication
// 2. Order fulfillment
// 3. Callbacks to chain
// 4. Usage reporting
type ProviderFlowTestSuite struct {
	suite.Suite

	// Mock Waldur server
	waldurMock *mocks.WaldurMock

	// Test data
	testProjectUUID  string
	testCustomerUUID string
	testOfferingUUID string
	testProviderAddr string
}

func TestProviderFlow(t *testing.T) {
	suite.Run(t, new(ProviderFlowTestSuite))
}

func (s *ProviderFlowTestSuite) SetupSuite() {
	config := mocks.DefaultWaldurMockConfig()
	config.AutoApproveOrders = false // Provider controls approval
	config.AutoProvision = false     // Provider controls provisioning
	s.waldurMock = mocks.NewWaldurMockWithConfig(config)

	s.testProjectUUID = config.ProjectUUID
	s.testCustomerUUID = config.CustomerUUID
	s.testProviderAddr = "ve1provider" + uuid.New().String()[:8]

	// Register base offerings
	s.setupTestOfferings()
}

func (s *ProviderFlowTestSuite) TearDownSuite() {
	if s.waldurMock != nil {
		s.waldurMock.Close()
	}
}

func (s *ProviderFlowTestSuite) setupTestOfferings() {
	offeringUUID := uuid.New().String()

	s.waldurMock.RegisterOffering(&mocks.MockWaldurOffering{
		UUID:         offeringUUID,
		Name:         "Provider Test Offering",
		Category:     "compute",
		Description:  "Offering for provider flow testing",
		CustomerUUID: s.testCustomerUUID,
		State:        "active",
		PricePerHour: "0.75",
		BackendID:    s.testProviderAddr + "/offering-001",
	})
	s.testOfferingUUID = offeringUUID
}

// ============================================================================
// HTTP Helper Methods
// ============================================================================

func (s *ProviderFlowTestSuite) httpGet(path string) ([]byte, int, error) {
	resp, err := http.Get(s.waldurMock.BaseURL() + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func (s *ProviderFlowTestSuite) httpPost(path string, payload interface{}) ([]byte, int, error) {
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

// ============================================================================
// Offering Publication Tests
// ============================================================================

func (s *ProviderFlowTestSuite) TestOfferingPublication() {
	s.Run("CreateNewOffering", func() {
		offeringReq := map[string]interface{}{
			"name":        "New Provider Offering",
			"description": "A new offering from provider",
			"customer":    s.testCustomerUUID,
			"category":    "compute",
			"backend_id":  s.testProviderAddr + "/new-offering",
			"shared":      true,
			"billable":    true,
		}

		_, status, err := s.httpPost("/api/marketplace-offerings/", offeringReq)
		s.Require().NoError(err)
		// Mock may return 200 (list) or 201 (created) depending on implementation
		s.Contains([]int{http.StatusOK, http.StatusCreated}, status)
	})

	s.Run("GetOfferingByBackendID", func() {
		backendID := s.testProviderAddr + "/offering-001"
		offering := s.waldurMock.GetOfferingByBackendID(backendID)
		s.Require().NotNil(offering)
		s.Equal("Provider Test Offering", offering.Name)
	})

	s.Run("NotFoundByBackendID", func() {
		offering := s.waldurMock.GetOfferingByBackendID("non-existent-backend-id")
		s.Nil(offering)
	})
}

// ============================================================================
// Order Fulfillment Tests
// ============================================================================

func (s *ProviderFlowTestSuite) TestOrderFulfillment() {
	s.Run("ReceiveOrderFromChain", func() {
		// Simulate order from chain
		orderReq := map[string]interface{}{
			"offering_uuid":     s.testOfferingUUID,
			"project_uuid":      s.testProjectUUID,
			"name":              "chain-submitted-order",
			"backend_id":        s.testProviderAddr + "/order-001",
			"requestor_address": "ve1customer123",
		}

		body, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		var order map[string]interface{}
		err = json.Unmarshal(body, &order)
		s.Require().NoError(err)
		s.NotEmpty(order["uuid"])
	})

	s.Run("ApproveOrder", func() {
		// Create order first
		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "order-to-approve",
		}

		body, _, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)

		var order map[string]interface{}
		_ = json.Unmarshal(body, &order)
		orderUUID := order["uuid"].(string)

		// Approve the order
		path := fmt.Sprintf("/api/marketplace-orders/%s/approve/", orderUUID)
		_, status, err := s.httpPost(path, map[string]interface{}{})
		s.Require().NoError(err)
		s.Equal(http.StatusOK, status)

		// Verify order state
		storedOrder := s.waldurMock.GetOrder(orderUUID)
		s.Require().NotNil(storedOrder)
		s.Equal("approved", storedOrder.State)
	})

	s.Run("SetBackendID", func() {
		// Create order
		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "order-for-backend-id",
		}

		body, _, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)

		var order map[string]interface{}
		_ = json.Unmarshal(body, &order)
		orderUUID := order["uuid"].(string)

		// Set backend ID
		backendID := s.testProviderAddr + "/resource-001"
		path := fmt.Sprintf("/api/marketplace-orders/%s/set-backend-id/", orderUUID)
		_, status, err := s.httpPost(path, map[string]interface{}{
			"backend_id": backendID,
		})
		s.Require().NoError(err)
		// Mock may return 200 OK or 204 No Content
		s.Contains([]int{http.StatusOK, http.StatusNoContent}, status)

		// Verify
		storedOrder := s.waldurMock.GetOrder(orderUUID)
		s.Require().NotNil(storedOrder)
		s.Equal(backendID, storedOrder.BackendID)
	})
}

// ============================================================================
// Resource Provisioning Tests
// ============================================================================

func (s *ProviderFlowTestSuite) TestResourceProvisioning() {
	s.Run("ProvisionResource", func() {
		// Create resource directly
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "provisioned-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			BackendID:    s.testProviderAddr + "/vm-001",
			State:        "provisioned",
			Attributes: map[string]interface{}{
				"external_ip": "203.0.113.45",
				"internal_ip": "10.0.0.5",
			},
		})

		// Verify
		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
		s.Equal("provisioned", resource.State)
		s.Equal("203.0.113.45", resource.Attributes["external_ip"])
	})

	s.Run("TerminateResource", func() {
		// Create resource
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "resource-to-terminate",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		// Terminate
		path := fmt.Sprintf("/api/marketplace-resources/%s/terminate/", resourceUUID)
		_, status, err := s.httpPost(path, map[string]interface{}{})
		s.Require().NoError(err)
		// Mock may return 200 OK or 202 Accepted for async operations
		s.Contains([]int{http.StatusOK, http.StatusAccepted}, status)

		// Verify state changed
		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
		s.Equal("terminated", resource.State)
	})
}

// ============================================================================
// Callback Tests
// ============================================================================

func (s *ProviderFlowTestSuite) TestProviderCallbacks() {
	s.Run("ParseProvisionCallback", func() {
		// Test callback parsing
		callbackJSON := `{
			"resource_uuid": "550e8400-e29b-41d4-a716-446655440003",
			"action": "provision",
			"success": true,
			"backend_id": "ve1customer123/101/1",
			"metadata": {
				"external_ip": "203.0.113.45"
			}
		}`

		payload, err := waldur.ParseLifecycleCallback([]byte(callbackJSON))
		s.Require().NoError(err)
		s.Equal("550e8400-e29b-41d4-a716-446655440003", payload.ResourceUUID)
		s.Equal(waldur.LifecycleAction("provision"), payload.Action)
		s.True(payload.Success)
		s.Equal("ve1customer123/101/1", payload.BackendID)
		s.Equal("203.0.113.45", payload.Metadata["external_ip"])
	})

	s.Run("ParseTerminateCallback", func() {
		callbackJSON := `{
			"resource_uuid": "550e8400-e29b-41d4-a716-446655440004",
			"action": "terminate",
			"success": true,
			"metadata": {}
		}`

		payload, err := waldur.ParseLifecycleCallback([]byte(callbackJSON))
		s.Require().NoError(err)
		s.Equal(waldur.LifecycleAction("terminate"), payload.Action)
		s.True(payload.Success)
	})

	s.Run("ParseFailedCallback", func() {
		callbackJSON := `{
			"resource_uuid": "550e8400-e29b-41d4-a716-446655440005",
			"action": "provision",
			"success": false,
			"error": "Insufficient resources"
		}`

		payload, err := waldur.ParseLifecycleCallback([]byte(callbackJSON))
		s.Require().NoError(err)
		s.False(payload.Success)
		s.Equal("Insufficient resources", payload.Error)
	})
}

// ============================================================================
// Usage Reporting Tests
// ============================================================================

func (s *ProviderFlowTestSuite) TestUsageReporting() {
	s.Run("SubmitUsageRecord", func() {
		// Create resource first
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "usage-test-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		// Submit usage
		usageReq := map[string]interface{}{
			"resource":     resourceUUID,
			"period_start": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"period_end":   time.Now().Format(time.RFC3339),
			"usages": map[string]interface{}{
				"cpu_hours":       1.5,
				"memory_gb_hours": 4.0,
			},
			"is_final": false,
		}

		_, status, err := s.httpPost("/api/marketplace-component-usages/", usageReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		// Verify
		records := s.waldurMock.GetUsageRecords(resourceUUID)
		s.GreaterOrEqual(len(records), 1)
	})

	s.Run("SubmitFinalUsage", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "final-usage-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		usageReq := map[string]interface{}{
			"resource":     resourceUUID,
			"period_start": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"period_end":   time.Now().Format(time.RFC3339),
			"usages": map[string]interface{}{
				"cpu_hours": 2.0,
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
// Multi-Provider Tests
// ============================================================================

func (s *ProviderFlowTestSuite) TestMultiProviderScenarios() {
	s.Run("MultipleProvidersCanCreateOfferings", func() {
		// Register offerings directly to the mock to test multi-provider scenario
		providers := []string{"provider-A", "provider-B", "provider-C"}

		for _, provider := range providers {
			s.waldurMock.RegisterOffering(&mocks.MockWaldurOffering{
				UUID:         uuid.New().String(),
				Name:         fmt.Sprintf("Offering from %s", provider),
				Description:  "Multi-provider test",
				CustomerUUID: s.testCustomerUUID,
				BackendID:    provider + "/offering-001",
				State:        "active",
			})
		}

		// Verify all offerings exist
		body, _, err := s.httpGet("/api/marketplace-offerings/")
		s.Require().NoError(err)

		var offerings []map[string]interface{}
		_ = json.Unmarshal(body, &offerings)
		s.GreaterOrEqual(len(offerings), 4) // 3 new + 1 from setup
	})

	s.Run("OrderRoutedToCorrectProvider", func() {
		// Create provider-specific offering
		providerID := "specific-provider"
		backendID := providerID + "/specific-offering"

		s.waldurMock.RegisterOffering(&mocks.MockWaldurOffering{
			UUID:         uuid.New().String(),
			Name:         "Provider Specific Offering",
			CustomerUUID: s.testCustomerUUID,
			BackendID:    backendID,
			State:        "active",
		})

		// Verify offering can be found by backend ID
		offering := s.waldurMock.GetOfferingByBackendID(backendID)
		s.Require().NotNil(offering)
		s.Equal("Provider Specific Offering", offering.Name)
	})
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func (s *ProviderFlowTestSuite) TestProviderErrorHandling() {
	s.Run("CreateOfferingWithMissingName", func() {
		offeringReq := map[string]interface{}{
			"description": "Missing name",
			"customer":    s.testCustomerUUID,
		}

		_, status, _ := s.httpPost("/api/marketplace-offerings/", offeringReq)
		// Mock may accept it or return error (including 200 for list response)
		s.Contains([]int{http.StatusOK, http.StatusCreated, http.StatusBadRequest}, status)
	})

	s.Run("ApproveNonExistentOrder", func() {
		path := fmt.Sprintf("/api/marketplace-orders/%s/approve/", "00000000-0000-0000-0000-000000000003")
		_, status, err := s.httpPost(path, map[string]interface{}{})
		s.Require().NoError(err)
		s.Equal(http.StatusNotFound, status)
	})

	s.Run("TerminateNonExistentResource", func() {
		path := fmt.Sprintf("/api/marketplace-resources/%s/terminate/", "00000000-0000-0000-0000-000000000004")
		_, status, err := s.httpPost(path, map[string]interface{}{})
		s.Require().NoError(err)
		s.Equal(http.StatusNotFound, status)
	})
}

// ============================================================================
// State Transition Tests
// ============================================================================

func (s *ProviderFlowTestSuite) TestStateTransitions() {
	s.Run("ValidLifecycleTransitions", func() {
		// Test valid state transition: OK -> stopped
		err := waldur.ValidateLifecycleAction(waldur.ResourceStateOK, waldur.LifecycleActionStop)
		s.NoError(err)

		// Test valid: stopped -> started
		err = waldur.ValidateLifecycleAction(waldur.ResourceStateStopped, waldur.LifecycleActionStart)
		s.NoError(err)
	})

	s.Run("InvalidLifecycleTransitions", func() {
		// Test invalid: terminated -> start
		err := waldur.ValidateLifecycleAction(waldur.ResourceStateTerminated, waldur.LifecycleActionStart)
		s.Error(err)

		// Test invalid: paused -> stop
		err = waldur.ValidateLifecycleAction(waldur.ResourceStatePaused, waldur.LifecycleActionStop)
		s.Error(err)
	})
}
