//go:build e2e.integration

package waldur

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/tests/e2e/mocks"
)

type waldurHarness struct {
	t *testing.T

	mock        *mocks.WaldurMock
	client      *waldur.Client
	marketplace *waldur.MarketplaceClient
	lifecycle   *waldur.LifecycleClient
	usage       *waldur.UsageClient

	projectUUID  string
	customerUUID string
	providerAddr string
}

func newWaldurHarness(t *testing.T, configure func(*mocks.WaldurMockConfig)) *waldurHarness {
	t.Helper()

	cfg := mocks.DefaultWaldurMockConfig()
	if configure != nil {
		configure(&cfg)
	}

	mock := mocks.NewWaldurMockWithConfig(cfg)
	t.Cleanup(mock.Close)

	clientCfg := waldur.DefaultConfig()
	clientCfg.BaseURL = mock.BaseURL()
	clientCfg.Token = "test-token"
	clientCfg.MaxRetries = 0
	clientCfg.RetryWaitMin = 5 * time.Millisecond
	clientCfg.RetryWaitMax = 20 * time.Millisecond

	client, err := waldur.NewClient(clientCfg)
	if err != nil {
		t.Fatalf("new Waldur client: %v", err)
	}

	return &waldurHarness{
		t:            t,
		mock:         mock,
		client:       client,
		marketplace:  waldur.NewMarketplaceClient(client),
		lifecycle:    waldur.NewLifecycleClient(waldur.NewMarketplaceClient(client)),
		usage:        waldur.NewUsageClient(waldur.NewMarketplaceClient(client)),
		projectUUID:  cfg.ProjectUUID,
		customerUUID: cfg.CustomerUUID,
		providerAddr: "ve1provider" + uuid.NewString()[:8],
	}
}

func (h *waldurHarness) context() (context.Context, context.CancelFunc) {
	h.t.Helper()
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (h *waldurHarness) tempStateFile(name string) string {
	h.t.Helper()
	return filepath.Join(h.t.TempDir(), name)
}

func (h *waldurHarness) createOffering(name, backendID, offeringType string, components []waldur.PricingComponent) *waldur.Offering {
	h.t.Helper()

	ctx, cancel := h.context()
	defer cancel()

	offering, err := h.marketplace.CreateOffering(ctx, waldur.CreateOfferingRequest{
		Name:         name,
		Description:  name + " description",
		Type:         offeringType,
		CustomerUUID: h.customerUUID,
		Shared:       true,
		Billable:     true,
		BackendID:    backendID,
		Components:   components,
	})
	if err != nil {
		h.t.Fatalf("create offering %q: %v", name, err)
	}
	return offering
}

func (h *waldurHarness) createOrder(offeringUUID, name string, attributes map[string]interface{}) *waldur.Order {
	h.t.Helper()

	ctx, cancel := h.context()
	defer cancel()

	order, err := h.marketplace.CreateOrder(ctx, waldur.CreateOrderRequest{
		OfferingUUID:   offeringUUID,
		ProjectUUID:    h.projectUUID,
		Type:           "Create",
		Name:           name,
		Description:    "test order for " + name,
		RequestComment: "e2e integration test",
		Attributes:     attributes,
	})
	if err != nil {
		h.t.Fatalf("create order %q: %v", name, err)
	}
	return order
}

func (h *waldurHarness) registerResource(offeringUUID, orderUUID, backendID, state string, attributes map[string]interface{}) string {
	h.t.Helper()

	resourceUUID := uuid.NewString()
	h.mock.RegisterResource(&mocks.MockWaldurResource{
		UUID:         resourceUUID,
		Name:         fmt.Sprintf("resource-%s", resourceUUID[:8]),
		OrderUUID:    orderUUID,
		OfferingUUID: offeringUUID,
		ProjectUUID:  h.projectUUID,
		BackendID:    backendID,
		State:        state,
		Attributes:   attributes,
	})
	return resourceUUID
}

func (h *waldurHarness) waitFor(description string, timeout time.Duration, condition func() bool) {
	h.t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	h.t.Fatalf("timed out waiting for %s", description)
}
