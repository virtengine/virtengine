//go:build e2e.integration

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/tests/e2e/fixtures"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
)

type waldurHarness struct {
	t           *testing.T
	mock        *mocks.WaldurMock
	client      *waldur.Client
	marketplace *waldur.MarketplaceClient
	lifecycle   *waldur.LifecycleClient
	usage       *waldur.UsageClient
	offering    fixtures.TestOffering
}

func newWaldurHarness(t *testing.T) *waldurHarness {
	t.Helper()

	mock := mocks.NewWaldurMock()
	t.Cleanup(func() {
		mock.Close()
	})

	cfg := waldur.DefaultConfig()
	cfg.BaseURL = mock.BaseURL()
	cfg.Token = "e2e-token"

	client, err := waldur.NewClient(cfg)
	require.NoError(t, err)

	marketplace := waldur.NewMarketplaceClient(client)
	lifecycle := waldur.NewLifecycleClient(marketplace)
	usage := waldur.NewUsageClient(marketplace)

	offering := fixtures.ComputeSmallOffering("provider-e2e")
	mock.RegisterOffering(&mocks.MockWaldurOffering{
		UUID:         offering.WaldurUUID,
		Name:         offering.Name,
		Category:     offering.Category,
		Description:  offering.Description,
		BackendID:    offering.OfferingID,
		CustomerUUID: mock.Config.CustomerUUID,
		State:        "active",
		PricePerHour: offering.PricePerHour.String(),
		Attributes: map[string]interface{}{
			"cpu_cores":  offering.CPUCores,
			"memory_gb":  offering.MemoryGB,
			"storage_gb": offering.StorageGB,
			"gpus":       offering.GPUs,
			"region":     offering.Region,
		},
		Components: []mocks.OfferingComponent{
			{Type: "cpu", Name: "CPU", Unit: "core-hour", Amount: 1.0},
			{Type: "memory", Name: "Memory", Unit: "gb-hour", Amount: 0.5},
			{Type: "storage", Name: "Storage", Unit: "gb-hour", Amount: 0.1},
		},
	})

	return &waldurHarness{
		t:           t,
		mock:        mock,
		client:      client,
		marketplace: marketplace,
		lifecycle:   lifecycle,
		usage:       usage,
		offering:    offering,
	}
}

func (h *waldurHarness) createOrder(ctx context.Context, name string, attrs map[string]interface{}) *waldur.Order {
	h.t.Helper()

	order, err := h.marketplace.CreateOrder(ctx, waldur.CreateOrderRequest{
		OfferingUUID: h.offering.WaldurUUID,
		ProjectUUID:  h.mock.Config.ProjectUUID,
		Name:         name,
		Description:  "e2e order",
		Attributes:   attrs,
	})
	require.NoError(h.t, err)

	return order
}

func (h *waldurHarness) waitForResource(orderUUID string) *mocks.MockWaldurResource {
	h.t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		order := h.mock.GetOrder(orderUUID)
		if order != nil && order.ResourceUUID != "" {
			resource := h.mock.GetResource(order.ResourceUUID)
			if resource != nil {
				return resource
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	h.t.Fatalf("resource not provisioned for order %s", orderUUID)
	return nil
}

func (h *waldurHarness) submitUsage(ctx context.Context, resourceUUID, backendID string) {
	h.t.Helper()

	now := time.Now().UTC()
	_, err := h.usage.SubmitUsageReport(ctx, &waldur.ResourceUsageReport{
		ResourceUUID: resourceUUID,
		PeriodStart:  now.Add(-1 * time.Hour),
		PeriodEnd:    now,
		BackendID:    backendID,
		Components: []waldur.ComponentUsage{
			{Type: "cpu", Amount: 1.0},
			{Type: "memory", Amount: 1.0},
		},
	})
	require.NoError(h.t, err)
}

func (h *waldurHarness) setMockError(state *mocks.WaldurMockErrorState) {
	h.t.Helper()
	h.mock.SetErrorState(state)
	h.t.Cleanup(func() {
		h.mock.ClearErrorState()
	})
}
