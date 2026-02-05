package provider_daemon

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
)

type mockWaldurMarketplace struct {
	createErr error
	orders    map[string]*waldur.Order
}

const testWaldurProjectID = "proj-1"

func (m *mockWaldurMarketplace) CreateOrder(_ context.Context, req waldur.CreateOrderRequest) (*waldur.Order, error) {
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

func (m *mockWaldurMarketplace) ApproveOrderByProvider(_ context.Context, _ string) error {
	return nil
}

func (m *mockWaldurMarketplace) SetOrderBackendID(_ context.Context, _ string, _ string) error {
	return nil
}

func (m *mockWaldurMarketplace) GetOrder(_ context.Context, orderUUID string) (*waldur.Order, error) {
	if m.orders == nil {
		return nil, waldur.ErrNotFound
	}
	order, ok := m.orders[orderUUID]
	if !ok {
		return nil, waldur.ErrNotFound
	}
	return order, nil
}

func (m *mockWaldurMarketplace) GetOfferingByBackendID(_ context.Context, _ string) (*waldur.Offering, error) {
	return &waldur.Offering{UUID: "off-1"}, nil
}

func TestOrderRouter_RouteOrderSuccess(t *testing.T) {
	tmp := t.TempDir()
	cfg := DefaultOrderRoutingConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = testProvider1
	cfg.WaldurProjectID = testWaldurProjectID
	cfg.StateFile = filepath.Join(tmp, "state.json")

	mockClient := &mockWaldurMarketplace{}
	resolver := &MapOfferingResolver{Map: map[string]string{testProvider1 + "/1": "off-1"}, Marketplace: mockClient}

	router, err := NewOrderRouterWithClient(cfg, mockClient, resolver)
	if err != nil {
		t.Fatalf("router create error: %v", err)
	}

	req := OrderRoutingRequest{
		OrderID:         "cust1/1",
		OfferingID:      testProvider1 + "/1",
		ProviderAddress: testProvider1,
		CustomerAddress: "cust1",
		Quantity:        1,
		MaxBidPrice:     100,
	}

	if err := router.routeOrder(context.Background(), req); err != nil {
		t.Fatalf("route order error: %v", err)
	}

	state, err := router.store.Load()
	if err != nil {
		t.Fatalf("state load error: %v", err)
	}
	record := state.Get(req.OrderID)
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
	tmp := t.TempDir()
	cfg := DefaultOrderRoutingConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = testProvider1
	cfg.WaldurProjectID = testWaldurProjectID
	cfg.StateFile = filepath.Join(tmp, "state.json")

	mockClient := &mockWaldurMarketplace{createErr: waldur.ErrRateLimited}
	resolver := &MapOfferingResolver{Map: map[string]string{testProvider1 + "/1": "off-1"}, Marketplace: mockClient}
	router, err := NewOrderRouterWithClient(cfg, mockClient, resolver)
	if err != nil {
		t.Fatalf("router create error: %v", err)
	}

	req := OrderRoutingRequest{
		OrderID:         "cust1/2",
		OfferingID:      testProvider1 + "/1",
		ProviderAddress: testProvider1,
		CustomerAddress: "cust1",
	}

	if err := router.processTask(context.Background(), OrderRoutingTask{Request: req}); err == nil {
		t.Fatalf("expected error")
	}

	state, err := router.store.Load()
	if err != nil {
		t.Fatalf("state load error: %v", err)
	}
	record := state.Get(req.OrderID)
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
	tmp := t.TempDir()
	cfg := DefaultOrderRoutingConfig()
	cfg.Enabled = true
	cfg.ProviderAddress = testProvider1
	cfg.WaldurProjectID = testWaldurProjectID
	cfg.StateFile = filepath.Join(tmp, "state.json")
	cfg.MaxRetries = 0

	mockClient := &mockWaldurMarketplace{createErr: waldur.ErrUnauthorized}
	resolver := &MapOfferingResolver{Map: map[string]string{testProvider1 + "/1": "off-1"}, Marketplace: mockClient}
	router, err := NewOrderRouterWithClient(cfg, mockClient, resolver)
	if err != nil {
		t.Fatalf("router create error: %v", err)
	}

	req := OrderRoutingRequest{
		OrderID:         "cust1/3",
		OfferingID:      testProvider1 + "/1",
		ProviderAddress: testProvider1,
		CustomerAddress: "cust1",
	}

	if err := router.processTask(context.Background(), OrderRoutingTask{Request: req}); err == nil {
		t.Fatalf("expected error")
	}

	state, err := router.store.Load()
	if err != nil {
		t.Fatalf("state load error: %v", err)
	}
	record := state.Get(req.OrderID)
	if record == nil {
		t.Fatalf("record not found")
	}
	if !record.DeadLettered {
		t.Fatalf("expected dead letter")
	}
}
