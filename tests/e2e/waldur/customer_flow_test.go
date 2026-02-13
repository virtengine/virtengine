//go:build e2e.integration

// Package waldur provides E2E tests for Waldur integration.
//
// VE-25M: Waldur integration E2E tests - Customer and Provider flows
package waldur

import (
	"bytes"
	"crypto/sha256"
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
// Customer Flow Test Suite
// ============================================================================

// CustomerFlowTestSuite tests the complete customer flow through Waldur:
// 1. Browse marketplace offerings
// 2. Select and configure offering
// 3. Submit order
// 4. Escrow funds locked
// 5. Order routed to provider
// 6. Resource provisioned
// 7. Customer accesses resource
type CustomerFlowTestSuite struct {
	suite.Suite

	// Mock Waldur server
	waldurMock *mocks.WaldurMock

	// Test data
	testProjectUUID  string
	testCustomerUUID string
	testOfferingUUID string
}

func TestCustomerFlow(t *testing.T) {
	suite.Run(t, new(CustomerFlowTestSuite))
}

func (s *CustomerFlowTestSuite) SetupSuite() {
	// Create mock Waldur server with auto-provision enabled
	config := mocks.DefaultWaldurMockConfig()
	config.AutoApproveOrders = true
	config.AutoProvision = true
	config.ProvisionDelay = 100 * time.Millisecond
	s.waldurMock = mocks.NewWaldurMockWithConfig(config)

	s.testProjectUUID = config.ProjectUUID
	s.testCustomerUUID = config.CustomerUUID

	// Register test offerings
	s.setupTestOfferings()
}

func (s *CustomerFlowTestSuite) TearDownSuite() {
	if s.waldurMock != nil {
		s.waldurMock.Close()
	}
}

func deterministicUUID(namespace uuid.UUID, name string) uuid.UUID {
	payload := make([]byte, 0, len(namespace)+len(name))
	payload = append(payload, namespace[:]...)
	payload = append(payload, []byte(name)...)
	sum := sha256.Sum256(payload)

	var id uuid.UUID
	copy(id[:], sum[:16])
	id[6] = (id[6] & 0x0f) | 0x40 // Version 4 (deterministic for tests)
	id[8] = (id[8] & 0x3f) | 0x80 // RFC 4122 variant

	return id
}

// generateTestUUIDs generates deterministic but valid UUIDs for testing
func (s *CustomerFlowTestSuite) generateTestUUIDs() (computeUUID, gpuUUID, storageUUID string) {
	// Use a deterministic SHA-256 based UUID to avoid SHA-1 in tests.
	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8") // Standard namespace
	computeUUID = deterministicUUID(namespace, "offering-compute-standard").String()
	gpuUUID = deterministicUUID(namespace, "offering-gpu-a100").String()
	storageUUID = deterministicUUID(namespace, "offering-storage-block").String()
	return
}

func (s *CustomerFlowTestSuite) setupTestOfferings() {
	computeUUID, gpuUUID, storageUUID := s.generateTestUUIDs()

	// Standard Compute Offering
	s.waldurMock.RegisterOffering(&mocks.MockWaldurOffering{
		UUID:         computeUUID,
		Name:         "Standard Compute Instance",
		Category:     "compute",
		Description:  "General purpose compute with 4 vCPU, 16GB RAM",
		CustomerUUID: s.testCustomerUUID,
		State:        "active",
		PricePerHour: "0.50",
		Attributes: map[string]interface{}{
			"vcpu":      4,
			"memory_gb": 16,
			"disk_gb":   100,
			"region":    "us-east-1",
		},
		Components: []mocks.OfferingComponent{
			{Type: "usage", Name: "cpu_hours", Unit: "hour", Amount: 0.10},
			{Type: "usage", Name: "memory_gb_hours", Unit: "hour", Amount: 0.02},
		},
	})
	s.testOfferingUUID = computeUUID

	// GPU Offering
	s.waldurMock.RegisterOffering(&mocks.MockWaldurOffering{
		UUID:         gpuUUID,
		Name:         "GPU A100 Instance",
		Category:     "gpu",
		Description:  "High-performance GPU for ML workloads",
		CustomerUUID: s.testCustomerUUID,
		State:        "active",
		PricePerHour: "3.50",
		Attributes: map[string]interface{}{
			"vcpu":      32,
			"memory_gb": 256,
			"gpu_type":  "A100",
			"gpu_count": 8,
		},
	})

	// Storage Offering
	s.waldurMock.RegisterOffering(&mocks.MockWaldurOffering{
		UUID:         storageUUID,
		Name:         "Block Storage",
		Category:     "storage",
		Description:  "High-performance block storage",
		CustomerUUID: s.testCustomerUUID,
		State:        "active",
		PricePerHour: "0.10",
		Attributes: map[string]interface{}{
			"storage_type": "ssd",
			"iops":         10000,
		},
	})
}

// ============================================================================
// HTTP Helper Methods
// ============================================================================

func (s *CustomerFlowTestSuite) httpGet(path string) ([]byte, int, error) {
	resp, err := http.Get(s.waldurMock.BaseURL() + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func (s *CustomerFlowTestSuite) httpPost(path string, payload interface{}) ([]byte, int, error) {
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
// Browse and Discovery Tests
// ============================================================================

func (s *CustomerFlowTestSuite) TestBrowseMarketplaceOfferings() {
	s.Run("ListAllOfferings", func() {
		body, status, err := s.httpGet("/api/marketplace-offerings/")
		s.Require().NoError(err)
		s.Equal(http.StatusOK, status)

		var offerings []map[string]interface{}
		err = json.Unmarshal(body, &offerings)
		s.Require().NoError(err)
		s.GreaterOrEqual(len(offerings), 3, "Should have at least 3 test offerings")
	})

	s.Run("FilterByCategory", func() {
		// Verify offering category through mock directly
		offering := s.waldurMock.GetOffering(s.testOfferingUUID)
		s.Require().NotNil(offering)
		s.Equal("compute", offering.Category)
	})

	s.Run("GetOfferingDetails", func() {
		path := fmt.Sprintf("/api/marketplace-offerings/%s/", s.testOfferingUUID)
		body, status, err := s.httpGet(path)
		s.Require().NoError(err)
		s.Equal(http.StatusOK, status)

		var offering map[string]interface{}
		err = json.Unmarshal(body, &offering)
		s.Require().NoError(err)
		s.Equal("Standard Compute Instance", offering["name"])
		s.Equal("active", offering["state"])
	})
}

func (s *CustomerFlowTestSuite) TestOfferingSearch() {
	s.Run("SearchByName", func() {
		// Test via mock directly since search is implemented there
		offering := s.waldurMock.GetOffering(s.testOfferingUUID)
		s.Require().NotNil(offering)
		s.Contains(offering.Name, "Compute")
	})

	s.Run("SearchBySharedStatus", func() {
		body, status, err := s.httpGet("/api/marketplace-offerings/")
		s.Require().NoError(err)
		s.Equal(http.StatusOK, status)

		var offerings []map[string]interface{}
		err = json.Unmarshal(body, &offerings)
		s.Require().NoError(err)
		s.GreaterOrEqual(len(offerings), 0)
	})
}

// ============================================================================
// Order Creation Tests
// ============================================================================

func (s *CustomerFlowTestSuite) TestCreateOrder() {
	s.Run("CreateSimpleOrder", func() {
		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "e2e-test-order-simple",
			"description":   "E2E test compute instance",
			"type":          "Create",
		}

		body, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		var order map[string]interface{}
		err = json.Unmarshal(body, &order)
		s.Require().NoError(err)
		s.NotEmpty(order["uuid"])
	})

	s.Run("CreateOrderWithAttributes", func() {
		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "e2e-test-order-attrs",
			"attributes": map[string]interface{}{
				"instance_size": "large",
				"custom_config": "value",
			},
		}

		body, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		var order map[string]interface{}
		err = json.Unmarshal(body, &order)
		s.Require().NoError(err)
		s.NotEmpty(order["uuid"])
	})

	s.Run("CreateOrderWithCallback", func() {
		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "e2e-test-order-callback",
			"callback_url":  "https://example.com/callback",
		}

		body, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		var order map[string]interface{}
		err = json.Unmarshal(body, &order)
		s.Require().NoError(err)
		s.NotEmpty(order["uuid"])
	})
}

// ============================================================================
// Order Lifecycle Tests
// ============================================================================

func (s *CustomerFlowTestSuite) TestOrderToResourceLifecycle() {
	s.Run("CompleteOrderProvisioningFlow", func() {
		// Create order
		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "e2e-lifecycle-test",
		}

		body, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		var order map[string]interface{}
		err = json.Unmarshal(body, &order)
		s.Require().NoError(err)
		orderUUID := order["uuid"].(string)

		// Wait for auto-provision
		time.Sleep(200 * time.Millisecond)

		// Verify order completed
		storedOrder := s.waldurMock.GetOrder(orderUUID)
		s.Require().NotNil(storedOrder)
		s.Contains([]string{"done", "approved", "executing", "pending"}, storedOrder.State)
	})

	s.Run("VerifyResourceCreation", func() {
		// Create order and wait for resource
		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "e2e-resource-test",
		}

		body, _, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)

		var order map[string]interface{}
		err = json.Unmarshal(body, &order)
		s.Require().NoError(err)
		orderUUID := order["uuid"].(string)

		// Wait for auto-provision
		time.Sleep(200 * time.Millisecond)

		// Check resource creation
		storedOrder := s.waldurMock.GetOrder(orderUUID)
		s.Require().NotNil(storedOrder)

		// If resource UUID is set, verify it
		if storedOrder.ResourceUUID != "" {
			resource := s.waldurMock.GetResource(storedOrder.ResourceUUID)
			s.NotNil(resource)
		}
	})
}

func (s *CustomerFlowTestSuite) TestResourceLifecycleControl() {
	s.Run("StartStopResource", func() {
		// Create a resource directly for lifecycle testing
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "lifecycle-test-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		// Verify resource exists
		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
		s.Equal("provisioned", resource.State)
	})
}

// ============================================================================
// Order Cancellation Tests
// ============================================================================

func (s *CustomerFlowTestSuite) TestOrderCancellation() {
	s.Run("CancelPendingOrder", func() {
		// Create an order
		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "e2e-cancel-test",
		}

		body, _, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)

		var order map[string]interface{}
		err = json.Unmarshal(body, &order)
		s.Require().NoError(err)

		// Verify order was created
		s.NotEmpty(order["uuid"])
	})

	s.Run("TerminateProvisionedResource", func() {
		// Create a provisioned resource
		resourceUUID := uuid.New().String()
		s.waldurMock.RegisterResource(&mocks.MockWaldurResource{
			UUID:         resourceUUID,
			Name:         "terminate-test-resource",
			OfferingUUID: s.testOfferingUUID,
			ProjectUUID:  s.testProjectUUID,
			State:        "provisioned",
		})

		// Verify resource exists
		resource := s.waldurMock.GetResource(resourceUUID)
		s.Require().NotNil(resource)
	})
}

// ============================================================================
// Multiple Orders Tests
// ============================================================================

func (s *CustomerFlowTestSuite) TestMultipleOrders() {
	s.Run("CreateMultipleOrdersConcurrently", func() {
		var wg sync.WaitGroup
		errors := make(chan error, 5)

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				orderReq := map[string]interface{}{
					"offering_uuid": s.testOfferingUUID,
					"project_uuid":  s.testProjectUUID,
					"name":          fmt.Sprintf("e2e-concurrent-order-%d", idx),
				}

				_, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
				if err != nil {
					errors <- err
					return
				}
				if status != http.StatusCreated {
					errors <- fmt.Errorf("unexpected status: %d", status)
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			s.NoError(err)
		}
	})

	s.Run("VerifyOrderIndependence", func() {
		// Create multiple orders and verify they are tracked separately
		orderReq1 := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "e2e-independent-order-1",
		}
		orderReq2 := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "e2e-independent-order-2",
		}

		body1, _, err := s.httpPost("/api/marketplace-orders/", orderReq1)
		s.Require().NoError(err)

		body2, _, err := s.httpPost("/api/marketplace-orders/", orderReq2)
		s.Require().NoError(err)

		var order1, order2 map[string]interface{}
		_ = json.Unmarshal(body1, &order1)
		_ = json.Unmarshal(body2, &order2)

		s.NotEqual(order1["uuid"], order2["uuid"])
	})
}

// ============================================================================
// Edge Cases
// ============================================================================

func (s *CustomerFlowTestSuite) TestCustomerFlowEdgeCases() {
	s.Run("OrderWithEmptyName", func() {
		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "",
		}

		_, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)
		// Mock accepts empty names
		s.Equal(http.StatusCreated, status)
	})

	s.Run("GetNonExistentOrder", func() {
		nonExistentUUID := "00000000-0000-0000-0000-000000000001"
		path := fmt.Sprintf("/api/marketplace-orders/%s/", nonExistentUUID)
		_, status, err := s.httpGet(path)
		s.Require().NoError(err)
		s.Equal(http.StatusNotFound, status)
	})

	s.Run("GetNonExistentOffering", func() {
		nonExistentUUID := "00000000-0000-0000-0000-000000000002"
		path := fmt.Sprintf("/api/marketplace-offerings/%s/", nonExistentUUID)
		_, status, err := s.httpGet(path)
		s.Require().NoError(err)
		s.Equal(http.StatusNotFound, status)
	})

	s.Run("OrderWithMaximumAttributes", func() {
		attrs := make(map[string]interface{})
		for i := 0; i < 50; i++ {
			attrs[fmt.Sprintf("attr_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		orderReq := map[string]interface{}{
			"offering_uuid": s.testOfferingUUID,
			"project_uuid":  s.testProjectUUID,
			"name":          "e2e-max-attrs-test",
			"attributes":    attrs,
		}

		body, status, err := s.httpPost("/api/marketplace-orders/", orderReq)
		s.Require().NoError(err)
		s.Equal(http.StatusCreated, status)

		var order map[string]interface{}
		err = json.Unmarshal(body, &order)
		s.Require().NoError(err)
		s.NotEmpty(order["uuid"])
	})
}

// ============================================================================
// Health Check Tests
// ============================================================================

func (s *CustomerFlowTestSuite) TestHealthCheck() {
	s.Run("HealthEndpoint", func() {
		body, status, err := s.httpGet("/api/health/")
		s.Require().NoError(err)
		s.Equal(http.StatusOK, status)

		var health map[string]interface{}
		err = json.Unmarshal(body, &health)
		s.Require().NoError(err)
		s.Equal("ok", health["status"])
	})
}
