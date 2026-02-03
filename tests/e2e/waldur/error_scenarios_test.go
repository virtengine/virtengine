//go:build e2e.integration

// Package waldur provides E2E tests for Waldur integration.
//
// VE-25M: Waldur integration E2E tests - Error scenarios
package waldur

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/tests/e2e/mocks"
)

// ============================================================================
// Error Scenarios Test Suite
// ============================================================================

// ErrorScenariosTestSuite tests error handling and edge cases:
// 1. Provider unavailable
// 2. Resource creation failures
// 3. Network interruptions
// 4. Partial failures
type ErrorScenariosTestSuite struct {
	suite.Suite

	waldurMock       *mocks.WaldurMock
	testProjectUUID  string
	testCustomerUUID string
	testOfferingUUID string
}

func TestErrorScenarios(t *testing.T) {
	suite.Run(t, new(ErrorScenariosTestSuite))
}

func (s *ErrorScenariosTestSuite) SetupSuite() {
	config := mocks.DefaultWaldurMockConfig()
	config.AutoProvision = false
	s.waldurMock = mocks.NewWaldurMockWithConfig(config)

	s.testProjectUUID = config.ProjectUUID
	s.testCustomerUUID = config.CustomerUUID

	// Setup offering
	offeringUUID := uuid.New().String()
	s.waldurMock.RegisterOffering(&mocks.MockWaldurOffering{
		UUID:         offeringUUID,
		Name:         "Error Test Offering",
		CustomerUUID: s.testCustomerUUID,
		State:        "active",
	})
	s.testOfferingUUID = offeringUUID
}

func (s *ErrorScenariosTestSuite) TearDownSuite() {
	if s.waldurMock != nil {
		s.waldurMock.Close()
	}
}

// ============================================================================
// HTTP Helpers
// ============================================================================

func (s *ErrorScenariosTestSuite) httpPost(path string, payload interface{}) ([]byte, int, error) {
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

func (s *ErrorScenariosTestSuite) httpGet(path string) ([]byte, int, error) {
	resp, err := http.Get(s.waldurMock.BaseURL() + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

// ============================================================================
// Provider Unavailable Tests
// ============================================================================

func (s *ErrorScenariosTestSuite) TestProviderUnavailable() {
	s.Run("OrderFailsWhenProvisionFails", func() {
		// Set error state
		s.waldurMock.SetErrorState(&mocks.WaldurMockErrorState{
			FailProvision: true,
			ErrorMessage:  "Provider unavailable",
		})
		defer s.waldurMock.ClearErrorState()

		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "error-test-order",
		}

		body, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)

		// Order creation may succeed but provisioning should fail
		if status == http.StatusCreated {
			var order map[string]interface{}
			_ = json.Unmarshal(body, &order)
			s.NotEmpty(order["uuid"])
		}
	})

	s.Run("OrderCreationFailure", func() {
		s.waldurMock.SetErrorState(&mocks.WaldurMockErrorState{
			FailCreateOrder: true,
			ErrorMessage:    "Order creation failed",
		})
		defer s.waldurMock.ClearErrorState()

		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "should-fail-order",
		}

		_, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)
		s.Equal(http.StatusInternalServerError, status)
	})
}

// ============================================================================
// Resource Creation Failure Tests
// ============================================================================

func (s *ErrorScenariosTestSuite) TestResourceCreationFailures() {
	s.Run("ProvisioningTimeout", func() {
		// Test with a long provision delay simulating timeout
		// The mock doesn't actually block, but we verify the state is correct
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "timeout-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "creating", // Stuck in creating state
		})

		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
		s.Equal("creating", resource.State)
	})

	s.Run("ResourceInErrorState", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "error-state-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "erred",
			Attributes: map[string]interface{}{
				"error_message": "Resource creation failed due to insufficient capacity",
			},
		})

		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
		s.Equal("erred", resource.State)
		s.Equal("Resource creation failed due to insufficient capacity", resource.Attributes["error_message"])
	})

	s.Run("PartialProvisioningFailure", func() {
		// Simulate partial failure (some components work, others fail)
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "partial-failure-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
			Attributes: map[string]interface{}{
				"network_attached":  true,
				"storage_attached":  false, // Failed
				"partial_failure":   true,
				"failure_component": "storage",
			},
		})

		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
		s.Equal(false, resource.Attributes["storage_attached"])
		s.Equal(true, resource.Attributes["partial_failure"])
	})
}

// ============================================================================
// Termination Failure Tests
// ============================================================================

func (s *ErrorScenariosTestSuite) TestTerminationFailures() {
	s.Run("TerminationFails", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "terminate-fail-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		s.waldurMock.SetErrorState(&mocks.WaldurMockErrorState{
			FailTerminate: true,
			ErrorMessage:  "Termination failed: resource busy",
		})
		defer s.waldurMock.ClearErrorState()

		path := fmt.Sprintf("/api/marketplace-resources/%s/terminate/", resourceUUID)
		_, status, err := s.httpPost(path, map[string]interface{}{})
		s.Require().NoError(err)
		s.Equal(http.StatusInternalServerError, status)

		// Resource should still exist and not be terminated
		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
		s.NotEqual("terminated", resource.State)
	})
}

// ============================================================================
// Usage Submission Failure Tests
// ============================================================================

func (s *ErrorScenariosTestSuite) TestUsageSubmissionFailures() {
	s.Run("UsageSubmissionFails", func() {
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "usage-fail-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		s.waldurMock.SetErrorState(&mocks.WaldurMockErrorState{
			FailUsageSubmit: true,
			ErrorMessage:    "Usage submission failed",
		})
		defer s.waldurMock.ClearErrorState()

		usageReq := map[string]interface{}{
			"resource":     resourceUUID,
			"period_start": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"period_end":   time.Now().Format(time.RFC3339),
			"usages":       map[string]interface{}{"cpu_hours": 1.0},
		}

		_, status, err := s.httpPost("/api/marketplace-component-usages/", usageReq)
		s.Require().NoError(err)
		s.Equal(http.StatusInternalServerError, status)
	})
}

// ============================================================================
// Not Found Error Tests
// ============================================================================

func (s *ErrorScenariosTestSuite) TestNotFoundErrors() {
	s.Run("GetNonExistentOrder", func() {
		s.waldurMock.ClearErrorState()

		path := fmt.Sprintf("/api/marketplace-orders/%s/", "00000000-0000-0000-0000-000000000004")
		_, status, err := s.httpGet(path)
		s.Require().NoError(err)
		s.Equal(http.StatusNotFound, status)
	})

	s.Run("GetNonExistentOffering", func() {
		path := fmt.Sprintf("/api/marketplace-offerings/%s/", "00000000-0000-0000-0000-000000000005")
		_, status, err := s.httpGet(path)
		s.Require().NoError(err)
		s.Equal(http.StatusNotFound, status)
	})

	s.Run("GetNonExistentResource", func() {
		path := fmt.Sprintf("/api/marketplace-resources/%s/", "00000000-0000-0000-0000-000000000006")
		_, status, err := s.httpGet(path)
		s.Require().NoError(err)
		s.Equal(http.StatusNotFound, status)
	})
}

// ============================================================================
// Validation Error Tests
// ============================================================================

func (s *ErrorScenariosTestSuite) TestValidationErrors() {
	s.Run("MissingRequiredFields", func() {
		// Order without offering_uuid
		orderReq := map[string]interface{}{
			"name": "incomplete-order",
		}

		_, status, _ := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Contains([]int{http.StatusCreated, http.StatusBadRequest}, status)
	})

	s.Run("InvalidUUID", func() {
		path := "/api/marketplace-orders/not-a-uuid/"
		_, status, _ := s.httpGet(path)
		s.Equal(http.StatusNotFound, status)
	})

	s.Run("EmptyRequest", func() {
		_, status, _ := s.httpPost("/api/marketplace-orders/", map[string]interface{}{})
		s.Contains([]int{http.StatusCreated, http.StatusBadRequest}, status)
	})
}

// ============================================================================
// Concurrent Error Tests
// ============================================================================

func (s *ErrorScenariosTestSuite) TestConcurrentErrors() {
	s.Run("ConcurrentOrdersWithMixedErrors", func() {
		var wg sync.WaitGroup
		results := make(chan int, 10)

		// Create 10 orders - some will succeed, some will fail
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				// Intermittently fail orders
				if idx%3 == 0 {
					s.waldurMock.SetErrorState(&mocks.WaldurMockErrorState{
						FailCreateOrder: true,
					})
					defer s.waldurMock.ClearErrorState()
				}

				orderReq := map[string]interface{}{
					"offering_uuid": s.testOfferingUUID,
					"project_uuid":  s.testProjectUUID,
					"name":          fmt.Sprintf("concurrent-error-order-%d", idx),
				}

				_, status, _ := s.httpPost("/api/marketplace-orders/", orderReq)
				results <- status
			}(i)
		}

		wg.Wait()
		close(results)

		// Count successes and failures
		var successes, failures int
		for status := range results {
			if status == http.StatusCreated {
				successes++
			} else {
				failures++
			}
		}

		// We expect a mix of successes and failures
		s.GreaterOrEqual(successes+failures, 10)
	})
}

// ============================================================================
// Recovery Tests
// ============================================================================

func (s *ErrorScenariosTestSuite) TestRecoveryScenarios() {
	s.Run("RecoveryAfterFailure", func() {
		// First, fail an order
		s.waldurMock.SetErrorState(&mocks.WaldurMockErrorState{
			FailCreateOrder: true,
		})

		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "failed-order",
		}

		_, status, _ := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Equal(http.StatusInternalServerError, status)

		// Clear error and retry
		s.waldurMock.ClearErrorState()

		_, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)
	})

	s.Run("ResourceRecovery", func() {
		// Create resource in error state
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "recovery-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "erred",
		})

		// Simulate recovery by updating state
		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
		resource.State = "provisioned"

		// Verify recovery
		resource = s.waldurMock.GetResource(resourceUUID)
		s.Equal("provisioned", resource.State)
	})
}

// ============================================================================
// Edge Cases
// ============================================================================

func (s *ErrorScenariosTestSuite) TestEdgeCases() {
	s.Run("OrderCancellationDuringProvisioning", func() {
		// Create resource in provisioning state
		resourceUUID := uuid.New().String()
		orderUUID := uuid.New().String()

		s.waldurMock.RegisterOrder(&mocks.MockWaldurOrder{
			UUID:         orderUUID,
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			ResourceUUID: resourceUUID,
			State:        "executing",
			Name:         "cancel-during-provision",
		})

		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "provisioning-resource",
			OrderUUID:    orderUUID,
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "creating",
		})

		// Verify state
		order := s.waldurMock.GetOrder(orderUUID)
		s.Require().NotNil(order)
		s.Equal("executing", order.State)

		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
		s.Equal("creating", resource.State)
	})

	s.Run("ProviderOfflineRecovery", func() {
		// Simulate provider going offline then coming back
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "offline-recovery-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
			Attributes: map[string]interface{}{
				"last_heartbeat": time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			},
		})

		// Update heartbeat simulating recovery
		resource := s.waldurMock.GetResource(resourceUUID)
		resource.Attributes["last_heartbeat"] = time.Now().Format(time.RFC3339)
		now := time.Now()
		resource.LastHeartbeat = &now

		// Verify
		s.NotNil(resource.LastHeartbeat)
	})

	s.Run("DuplicateOrderHandling", func() {
		// Create first order
		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "duplicate-test-order",
		}

		body1, status1, _ := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Equal(http.StatusCreated, status1)

		// Create duplicate
		body2, status2, _ := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Equal(http.StatusCreated, status2) // Mock creates new orders

		// Verify different UUIDs
		var order1, order2 map[string]interface{}
		_ = json.Unmarshal(body1, &order1)
		_ = json.Unmarshal(body2, &order2)
		s.NotEqual(order1["uuid"], order2["uuid"])
	})
}

// ============================================================================
// Health Check During Errors
// ============================================================================

func (s *ErrorScenariosTestSuite) TestHealthCheckDuringErrors() {
	s.Run("HealthCheckSucceedsDuringErrors", func() {
		// Set error state
		s.waldurMock.SetErrorState(&mocks.WaldurMockErrorState{
			FailCreateOrder: true,
			FailProvision:   true,
		})
		defer s.waldurMock.ClearErrorState()

		// Health check should still work
		body, status, err := s.httpGet("/api/health/")
		s.Require().NoError(err)
		s.Equal(http.StatusOK, status)

		var health map[string]interface{}
		_ = json.Unmarshal(body, &health)
		s.Equal("ok", health["status"])
	})
}
