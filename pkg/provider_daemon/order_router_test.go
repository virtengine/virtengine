package provider_daemon

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

type mockWaldurOrderClient struct {
	createErr error
	orders    map[string]*waldur.Order
}

func (m *mockWaldurOrderClient) CreateOrder(_ context.Context, req waldur.CreateOrderRequest) (*waldur.Order, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	order := &waldur.Order{
		UUID:         "ord-1",
		ResourceUUID: "res-1",
		State:        "executing",
		ProjectUUID:  req.ProjectUUID,
		CreatedAt:    time.Now(),
	}
	if m.orders == nil {
		m.orders = map[string]*waldur.Order{}
	}
	m.orders[order.UUID] = order
	return order, nil
}

func (m *mockWaldurOrderClient) ApproveOrderByProvider(_ context.Context, _ string) error {
	return nil
}

func (m *mockWaldurOrderClient) SetOrderBackendID(_ context.Context, _ string, _ string) error {
	return nil
}

func (m *mockWaldurOrderClient) GetOrder(_ context.Context, orderUUID string) (*waldur.Order, error) {
	if m.orders == nil {
		return nil, waldur.ErrNotFound
	}
	order, ok := m.orders[orderUUID]
	if !ok {
		return nil, waldur.ErrNotFound
	}
	return order, nil
}

func (m *mockWaldurOrderClient) GetOfferingByBackendID(_ context.Context, _ string) (*waldur.Offering, error) {
	return &waldur.Offering{UUID: "off-1"}, nil
}

func TestOrderRouter_ProcessOrderCreated(t *testing.T) {
	const (
		testBaseURL   = "http://waldur.local"
		testToken     = "token"
		testProjectID = "proj-1"
	)
	tmp := t.TempDir()
	cfg := DefaultOrderRouterConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = "provider1"
	cfg.WaldurBaseURL = testBaseURL
	cfg.WaldurToken = testToken
	cfg.WaldurProjectUUID = testProjectID
	cfg.StateFile = filepath.Join(tmp, "state.json")

	router, err := NewOrderRouterWithClient(cfg, &mockWaldurOrderClient{}, nil, nil)
	if err != nil {
		t.Fatalf("router create error: %v", err)
	}

	event := marketplace.OrderCreatedEvent{
		BaseMarketplaceEvent: marketplace.BaseMarketplaceEvent{Sequence: 1},
		OrderID:              "cust1/1",
		OfferingID:           "provider1/1",
		ProviderAddress:      "provider1",
		CustomerAddress:      "cust1",
		Quantity:             1,
		MaxBidPrice:          100,
	}

	if err := router.ProcessOrderCreated(context.Background(), event); err != nil {
		t.Fatalf("process order error: %v", err)
	}

	state, err := router.stateStore.Load()
	if err != nil {
		t.Fatalf("state load error: %v", err)
	}

	record := state.Records[event.OrderID]
	if record == nil {
		t.Fatalf("record not found")
	}
	if record.WaldurOrderUUID == "" {
		t.Fatalf("expected waldur order uuid")
	}
	if record.RetryCount != 0 {
		t.Fatalf("expected retry count 0, got %d", record.RetryCount)
	}
}

func TestOrderRouter_RetryableFailure(t *testing.T) {
	const (
		testBaseURL   = "http://waldur.local"
		testToken     = "token"
		testProjectID = "proj-1"
	)
	tmp := t.TempDir()
	cfg := DefaultOrderRouterConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = "provider1"
	cfg.WaldurBaseURL = testBaseURL
	cfg.WaldurToken = testToken
	cfg.WaldurProjectUUID = testProjectID
	cfg.StateFile = filepath.Join(tmp, "state.json")

	mockClient := &mockWaldurOrderClient{createErr: waldur.ErrRateLimited}
	router, err := NewOrderRouterWithClient(cfg, mockClient, nil, nil)
	if err != nil {
		t.Fatalf("router create error: %v", err)
	}

	event := marketplace.OrderCreatedEvent{
		BaseMarketplaceEvent: marketplace.BaseMarketplaceEvent{Sequence: 2},
		OrderID:              "cust1/2",
		OfferingID:           "provider1/1",
		ProviderAddress:      "provider1",
		CustomerAddress:      "cust1",
		Quantity:             1,
		MaxBidPrice:          100,
	}

	if err := router.ProcessOrderCreated(context.Background(), event); err == nil {
		t.Fatalf("expected error")
	}

	state, err := router.stateStore.Load()
	if err != nil {
		t.Fatalf("state load error: %v", err)
	}
	record := state.Records[event.OrderID]
	if record == nil {
		t.Fatalf("record not found")
	}
	if record.RetryCount == 0 {
		t.Fatalf("expected retry count > 0")
	}
	if record.DeadLettered {
		t.Fatalf("did not expect dead letter")
	}
}

func TestOrderRouter_FatalFailure(t *testing.T) {
	const (
		testBaseURL   = "http://waldur.local"
		testToken     = "token"
		testProjectID = "proj-1"
	)
	tmp := t.TempDir()
	cfg := DefaultOrderRouterConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = "provider1"
	cfg.WaldurBaseURL = testBaseURL
	cfg.WaldurToken = testToken
	cfg.WaldurProjectUUID = testProjectID
	cfg.StateFile = filepath.Join(tmp, "state.json")

	mockClient := &mockWaldurOrderClient{createErr: waldur.ErrUnauthorized}
	router, err := NewOrderRouterWithClient(cfg, mockClient, nil, nil)
	if err != nil {
		t.Fatalf("router create error: %v", err)
	}

	event := marketplace.OrderCreatedEvent{
		BaseMarketplaceEvent: marketplace.BaseMarketplaceEvent{Sequence: 3},
		OrderID:              "cust1/3",
		OfferingID:           "provider1/1",
		ProviderAddress:      "provider1",
		CustomerAddress:      "cust1",
		Quantity:             1,
		MaxBidPrice:          100,
	}

	err = router.ProcessOrderCreated(context.Background(), event)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, waldur.ErrUnauthorized) && !strings.Contains(err.Error(), "dead-lettered") {
		t.Fatalf("unexpected error: %v", err)
	}

	state, err := router.stateStore.Load()
	if err != nil {
		t.Fatalf("state load error: %v", err)
	}
	record := state.Records[event.OrderID]
	if record == nil {
		t.Fatalf("record not found")
	}
	if !record.DeadLettered {
		t.Fatalf("expected dead letter")
	}
}
