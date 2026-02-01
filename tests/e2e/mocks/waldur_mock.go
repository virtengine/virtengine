//go:build e2e.integration

// Package mocks provides mock implementations for E2E testing.
//
// VE-15C: Waldur mock for E2E provider flow testing
package mocks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/google/uuid"
)

// WaldurMock provides a mock Waldur API for E2E testing.
// It simulates the provisioning lifecycle for resource ordering and management.
type WaldurMock struct {
	mu sync.RWMutex

	// Server is the mock HTTP server
	Server *httptest.Server

	// Resources tracks provisioned resources
	Resources map[string]*MockWaldurResource

	// Orders tracks marketplace orders
	Orders map[string]*MockWaldurOrder

	// Offerings tracks available offerings
	Offerings map[string]*MockWaldurOffering

	// UsageRecords tracks submitted usage records
	UsageRecords map[string][]*MockWaldurUsageRecord

	// Invoices tracks generated invoices
	Invoices map[string]*MockWaldurInvoice

	// Callbacks tracks callback requests
	Callbacks []CallbackRequest

	// Config holds mock configuration
	Config WaldurMockConfig

	// ErrorState allows injecting errors for testing
	ErrorState *WaldurMockErrorState
}

// WaldurMockConfig configures the mock behavior
type WaldurMockConfig struct {
	// ProvisionDelay simulates provisioning time
	ProvisionDelay time.Duration

	// TerminateDelay simulates termination time
	TerminateDelay time.Duration

	// AutoApproveOrders automatically approves orders
	AutoApproveOrders bool

	// AutoProvision automatically provisions resources
	AutoProvision bool

	// EnableUsageReporting enables usage record submission
	EnableUsageReporting bool

	// ProjectUUID is the default project UUID
	ProjectUUID string

	// CustomerUUID is the default customer UUID
	CustomerUUID string
}

// DefaultWaldurMockConfig returns default mock configuration
func DefaultWaldurMockConfig() WaldurMockConfig {
	return WaldurMockConfig{
		ProvisionDelay:       0,
		TerminateDelay:       0,
		AutoApproveOrders:    true,
		AutoProvision:        true,
		EnableUsageReporting: true,
		ProjectUUID:          "e2e-project-" + uuid.New().String()[:8],
		CustomerUUID:         "e2e-customer-" + uuid.New().String()[:8],
	}
}

// WaldurMockErrorState allows injecting errors for testing error paths
type WaldurMockErrorState struct {
	// FailCreateOrder fails order creation
	FailCreateOrder bool

	// FailApproveOrder fails order approval
	FailApproveOrder bool

	// FailProvision fails resource provisioning
	FailProvision bool

	// FailTerminate fails resource termination
	FailTerminate bool

	// FailUsageSubmit fails usage record submission
	FailUsageSubmit bool

	// ErrorMessage is the error message to return
	ErrorMessage string
}

// MockWaldurResource represents a provisioned resource
type MockWaldurResource struct {
	UUID           string                 `json:"uuid"`
	Name           string                 `json:"name"`
	OrderUUID      string                 `json:"order_uuid"`
	OfferingUUID   string                 `json:"offering_uuid"`
	ProjectUUID    string                 `json:"project_uuid"`
	BackendID      string                 `json:"backend_id"`
	State          string                 `json:"state"`
	Attributes     map[string]interface{} `json:"attributes"`
	CreatedAt      time.Time              `json:"created"`
	ProvisionedAt  *time.Time             `json:"provisioned,omitempty"`
	TerminatedAt   *time.Time             `json:"terminated,omitempty"`
	LastHeartbeat  *time.Time             `json:"last_heartbeat,omitempty"`
	UsageRecordIDs []string               `json:"usage_record_ids"`
}

// MockWaldurOrder represents a marketplace order
type MockWaldurOrder struct {
	UUID          string                 `json:"uuid"`
	OfferingUUID  string                 `json:"offering_uuid"`
	ProjectUUID   string                 `json:"project_uuid"`
	ResourceUUID  string                 `json:"resource_uuid,omitempty"`
	BackendID     string                 `json:"backend_id,omitempty"`
	State         string                 `json:"state"`
	Attributes    map[string]interface{} `json:"attributes"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	CallbackURL   string                 `json:"callback_url,omitempty"`
	CreatedAt     time.Time              `json:"created"`
	ApprovedAt    *time.Time             `json:"approved_at,omitempty"`
	ApprovedBy    string                 `json:"approved_by,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	RequestorAddr string                 `json:"requestor_address,omitempty"`
}

// MockWaldurOffering represents a marketplace offering
type MockWaldurOffering struct {
	UUID            string                 `json:"uuid"`
	Name            string                 `json:"name"`
	Category        string                 `json:"category"`
	Description     string                 `json:"description"`
	BackendID       string                 `json:"backend_id,omitempty"`
	CustomerUUID    string                 `json:"customer_uuid"`
	State           string                 `json:"state"`
	PricePerHour    string                 `json:"price_per_hour"`
	Attributes      map[string]interface{} `json:"attributes"`
	Components      []OfferingComponent    `json:"components"`
	CreatedAt       time.Time              `json:"created"`
	ActiveOrders    int                    `json:"active_orders"`
	ActiveResources int                    `json:"active_resources"`
}

// OfferingComponent represents an offering pricing component
type OfferingComponent struct {
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Unit       string  `json:"unit"`
	Amount     float64 `json:"amount"`
	MeasuredIn string  `json:"measured_in"`
}

// MockWaldurUsageRecord represents a usage record
type MockWaldurUsageRecord struct {
	UUID         string                 `json:"uuid"`
	ResourceUUID string                 `json:"resource_uuid"`
	PeriodStart  time.Time              `json:"period_start"`
	PeriodEnd    time.Time              `json:"period_end"`
	Metrics      map[string]interface{} `json:"metrics"`
	IsFinal      bool                   `json:"is_final"`
	CreatedAt    time.Time              `json:"created"`
	BilledAmount string                 `json:"billed_amount,omitempty"`
}

// MockWaldurInvoice represents an invoice
type MockWaldurInvoice struct {
	UUID         string                      `json:"uuid"`
	ResourceUUID string                      `json:"resource_uuid"`
	CustomerUUID string                      `json:"customer_uuid"`
	LineItems    []MockWaldurInvoiceLineItem `json:"line_items"`
	TotalAmount  string                      `json:"total_amount"`
	Currency     string                      `json:"currency"`
	State        string                      `json:"state"`
	PeriodStart  time.Time                   `json:"period_start"`
	PeriodEnd    time.Time                   `json:"period_end"`
	CreatedAt    time.Time                   `json:"created"`
	PaidAt       *time.Time                  `json:"paid_at,omitempty"`
}

// MockWaldurInvoiceLineItem represents an invoice line item
type MockWaldurInvoiceLineItem struct {
	Name      string `json:"name"`
	Quantity  string `json:"quantity"`
	UnitPrice string `json:"unit_price"`
	Total     string `json:"total"`
	Unit      string `json:"unit"`
}

// CallbackRequest represents a captured callback request
type CallbackRequest struct {
	Method    string
	URL       string
	Body      []byte
	Headers   http.Header
	Timestamp time.Time
}

// NewWaldurMock creates a new Waldur mock with default configuration
func NewWaldurMock() *WaldurMock {
	return NewWaldurMockWithConfig(DefaultWaldurMockConfig())
}

// NewWaldurMockWithConfig creates a new Waldur mock with custom configuration
func NewWaldurMockWithConfig(config WaldurMockConfig) *WaldurMock {
	mock := &WaldurMock{
		Resources:    make(map[string]*MockWaldurResource),
		Orders:       make(map[string]*MockWaldurOrder),
		Offerings:    make(map[string]*MockWaldurOffering),
		UsageRecords: make(map[string][]*MockWaldurUsageRecord),
		Invoices:     make(map[string]*MockWaldurInvoice),
		Callbacks:    make([]CallbackRequest, 0),
		Config:       config,
	}

	// Create HTTP server
	mock.Server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))

	return mock
}

// Close shuts down the mock server
func (m *WaldurMock) Close() {
	if m.Server != nil {
		m.Server.Close()
	}
}

// BaseURL returns the mock server URL
func (m *WaldurMock) BaseURL() string {
	return m.Server.URL
}

// SetErrorState sets the error state for injecting failures
func (m *WaldurMock) SetErrorState(state *WaldurMockErrorState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorState = state
}

// ClearErrorState clears all injected errors
func (m *WaldurMock) ClearErrorState() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorState = nil
}

// RegisterOffering registers a new offering
func (m *WaldurMock) RegisterOffering(offering *MockWaldurOffering) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if offering.UUID == "" {
		offering.UUID = uuid.New().String()
	}
	if offering.State == "" {
		offering.State = "active"
	}
	if offering.CreatedAt.IsZero() {
		offering.CreatedAt = time.Now().UTC()
	}
	m.Offerings[offering.UUID] = offering
}

// GetResource returns a resource by UUID
func (m *WaldurMock) GetResource(uuid string) *MockWaldurResource {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Resources[uuid]
}

// GetOrder returns an order by UUID
func (m *WaldurMock) GetOrder(uuid string) *MockWaldurOrder {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Orders[uuid]
}

// GetOffering returns an offering by UUID
func (m *WaldurMock) GetOffering(uuid string) *MockWaldurOffering {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Offerings[uuid]
}

// GetOfferingByBackendID returns an offering by backend ID
func (m *WaldurMock) GetOfferingByBackendID(backendID string) *MockWaldurOffering {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, o := range m.Offerings {
		if o.BackendID == backendID {
			return o
		}
	}
	return nil
}

// GetUsageRecords returns usage records for a resource
func (m *WaldurMock) GetUsageRecords(resourceUUID string) []*MockWaldurUsageRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.UsageRecords[resourceUUID]
}

// GetInvoice returns an invoice by UUID
func (m *WaldurMock) GetInvoice(uuid string) *MockWaldurInvoice {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Invoices[uuid]
}

// GetCallbacks returns all captured callbacks
func (m *WaldurMock) GetCallbacks() []CallbackRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]CallbackRequest, len(m.Callbacks))
	copy(result, m.Callbacks)
	return result
}

// CountActiveResources returns the count of active resources
func (m *WaldurMock) CountActiveResources() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, r := range m.Resources {
		if r.State == "provisioned" || r.State == "active" {
			count++
		}
	}
	return count
}

// CountCompletedOrders returns the count of completed orders
func (m *WaldurMock) CountCompletedOrders() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, o := range m.Orders {
		if o.State == "done" || o.State == "completed" {
			count++
		}
	}
	return count
}

// handleRequest routes requests to appropriate handlers
func (m *WaldurMock) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Record the request
	m.recordRequest(r)

	// Route based on path
	switch {
	case r.URL.Path == "/api/health/":
		m.handleHealth(w, r)
	case r.URL.Path == "/api/marketplace-orders/":
		m.handleOrders(w, r)
	case matchPath(r.URL.Path, "/api/marketplace-orders/*/approve/"):
		m.handleApproveOrder(w, r)
	case matchPath(r.URL.Path, "/api/marketplace-orders/*/set-backend-id/"):
		m.handleSetBackendID(w, r)
	case matchPath(r.URL.Path, "/api/marketplace-orders/*"):
		m.handleOrderDetail(w, r)
	case r.URL.Path == "/api/marketplace-offerings/":
		m.handleOfferings(w, r)
	case matchPath(r.URL.Path, "/api/marketplace-offerings/*"):
		m.handleOfferingDetail(w, r)
	case r.URL.Path == "/api/marketplace-resources/":
		m.handleResources(w, r)
	case matchPath(r.URL.Path, "/api/marketplace-resources/*/terminate/"):
		m.handleTerminateResource(w, r)
	case matchPath(r.URL.Path, "/api/marketplace-resources/*"):
		m.handleResourceDetail(w, r)
	case matchPath(r.URL.Path, "/api/marketplace-component-usages/"):
		m.handleUsageSubmit(w, r)
	case r.URL.Path == "/api/invoices/":
		m.handleInvoices(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (m *WaldurMock) recordRequest(r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	body := make([]byte, 0)
	if r.Body != nil {
		body, _ = json.Marshal(r.Body)
	}

	m.Callbacks = append(m.Callbacks, CallbackRequest{
		Method:    r.Method,
		URL:       r.URL.String(),
		Body:      body,
		Headers:   r.Header.Clone(),
		Timestamp: time.Now().UTC(),
	})
}

func (m *WaldurMock) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *WaldurMock) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		m.handleListOrders(w, r)
	case http.MethodPost:
		m.handleCreateOrder(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (m *WaldurMock) handleListOrders(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	orders := make([]*MockWaldurOrder, 0, len(m.Orders))
	for _, o := range m.Orders {
		orders = append(orders, o)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(orders)
}

func (m *WaldurMock) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for injected errors
	if m.ErrorState != nil && m.ErrorState.FailCreateOrder {
		http.Error(w, m.ErrorState.ErrorMessage, http.StatusInternalServerError)
		return
	}

	var req struct {
		OfferingUUID   string                 `json:"offering"`
		ProjectUUID    string                 `json:"project"`
		Name           string                 `json:"name"`
		Description    string                 `json:"description"`
		Attributes     map[string]interface{} `json:"attributes"`
		CallbackURL    string                 `json:"callback_url"`
		RequestComment string                 `json:"request_comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order := &MockWaldurOrder{
		UUID:         uuid.New().String(),
		OfferingUUID: req.OfferingUUID,
		ProjectUUID:  req.ProjectUUID,
		Name:         req.Name,
		Description:  req.Description,
		Attributes:   req.Attributes,
		CallbackURL:  req.CallbackURL,
		State:        "pending",
		CreatedAt:    time.Now().UTC(),
	}

	m.Orders[order.UUID] = order

	// Auto-provision if configured
	if m.Config.AutoApproveOrders && m.Config.AutoProvision {
		go m.autoProvision(order.UUID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(order)
}

func (m *WaldurMock) autoProvision(orderUUID string) {
	if m.Config.ProvisionDelay > 0 {
		time.Sleep(m.Config.ProvisionDelay)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	order, ok := m.Orders[orderUUID]
	if !ok {
		return
	}

	// Create resource
	resource := &MockWaldurResource{
		UUID:         uuid.New().String(),
		Name:         order.Name,
		OrderUUID:    order.UUID,
		OfferingUUID: order.OfferingUUID,
		ProjectUUID:  order.ProjectUUID,
		State:        "provisioned",
		Attributes:   order.Attributes,
		CreatedAt:    time.Now().UTC(),
	}
	now := time.Now().UTC()
	resource.ProvisionedAt = &now

	m.Resources[resource.UUID] = resource

	// Update order
	order.ResourceUUID = resource.UUID
	order.State = "done"
	order.CompletedAt = &now
}

func (m *WaldurMock) handleApproveOrder(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ErrorState != nil && m.ErrorState.FailApproveOrder {
		http.Error(w, m.ErrorState.ErrorMessage, http.StatusInternalServerError)
		return
	}

	orderUUID := extractUUID(r.URL.Path, "/api/marketplace-orders/", "/approve/")
	order, ok := m.Orders[orderUUID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	now := time.Now().UTC()
	order.State = "approved"
	order.ApprovedAt = &now
	order.ApprovedBy = "e2e-provider"

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}

func (m *WaldurMock) handleSetBackendID(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	orderUUID := extractUUID(r.URL.Path, "/api/marketplace-orders/", "/set-backend-id/")
	order, ok := m.Orders[orderUUID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	var req struct {
		BackendID string `json:"backend_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order.BackendID = req.BackendID

	// Also update resource if exists
	if order.ResourceUUID != "" {
		if resource, ok := m.Resources[order.ResourceUUID]; ok {
			resource.BackendID = req.BackendID
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (m *WaldurMock) handleOrderDetail(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	orderUUID := extractUUID(r.URL.Path, "/api/marketplace-orders/", "")
	order, ok := m.Orders[orderUUID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(order)
}

func (m *WaldurMock) handleOfferings(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	offerings := make([]*MockWaldurOffering, 0, len(m.Offerings))
	for _, o := range m.Offerings {
		offerings = append(offerings, o)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(offerings)
}

func (m *WaldurMock) handleOfferingDetail(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	offeringUUID := extractUUID(r.URL.Path, "/api/marketplace-offerings/", "")
	offering, ok := m.Offerings[offeringUUID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(offering)
}

func (m *WaldurMock) handleResources(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resources := make([]*MockWaldurResource, 0, len(m.Resources))
	for _, res := range m.Resources {
		resources = append(resources, res)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resources)
}

func (m *WaldurMock) handleResourceDetail(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	resourceUUID := extractUUID(r.URL.Path, "/api/marketplace-resources/", "")
	resource, ok := m.Resources[resourceUUID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resource)
}

func (m *WaldurMock) handleTerminateResource(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ErrorState != nil && m.ErrorState.FailTerminate {
		http.Error(w, m.ErrorState.ErrorMessage, http.StatusInternalServerError)
		return
	}

	resourceUUID := extractUUID(r.URL.Path, "/api/marketplace-resources/", "/terminate/")
	resource, ok := m.Resources[resourceUUID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	now := time.Now().UTC()
	resource.State = "terminated"
	resource.TerminatedAt = &now

	w.WriteHeader(http.StatusAccepted)
}

func (m *WaldurMock) handleUsageSubmit(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ErrorState != nil && m.ErrorState.FailUsageSubmit {
		http.Error(w, m.ErrorState.ErrorMessage, http.StatusInternalServerError)
		return
	}

	if !m.Config.EnableUsageReporting {
		http.Error(w, "Usage reporting disabled", http.StatusForbidden)
		return
	}

	var req struct {
		ResourceUUID string                 `json:"resource"`
		PeriodStart  time.Time              `json:"period_start"`
		PeriodEnd    time.Time              `json:"period_end"`
		Usages       map[string]interface{} `json:"usages"`
		IsFinal      bool                   `json:"is_final"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record := &MockWaldurUsageRecord{
		UUID:         uuid.New().String(),
		ResourceUUID: req.ResourceUUID,
		PeriodStart:  req.PeriodStart,
		PeriodEnd:    req.PeriodEnd,
		Metrics:      req.Usages,
		IsFinal:      req.IsFinal,
		CreatedAt:    time.Now().UTC(),
	}

	m.UsageRecords[req.ResourceUUID] = append(m.UsageRecords[req.ResourceUUID], record)

	// Update resource
	if resource, ok := m.Resources[req.ResourceUUID]; ok {
		resource.UsageRecordIDs = append(resource.UsageRecordIDs, record.UUID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(record)
}

func (m *WaldurMock) handleInvoices(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	invoices := make([]*MockWaldurInvoice, 0, len(m.Invoices))
	for _, inv := range m.Invoices {
		invoices = append(invoices, inv)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(invoices)
}

// CreateInvoice creates an invoice for testing settlement
func (m *WaldurMock) CreateInvoice(ctx context.Context, resourceUUID string, periodStart, periodEnd time.Time, lineItems []MockWaldurInvoiceLineItem, total string) (*MockWaldurInvoice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	resource, ok := m.Resources[resourceUUID]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceUUID)
	}

	invoice := &MockWaldurInvoice{
		UUID:         uuid.New().String(),
		ResourceUUID: resourceUUID,
		CustomerUUID: m.Config.CustomerUUID,
		LineItems:    lineItems,
		TotalAmount:  total,
		Currency:     "uve",
		State:        "pending",
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
		CreatedAt:    time.Now().UTC(),
	}

	m.Invoices[invoice.UUID] = invoice

	// Update resource
	resource.UsageRecordIDs = append(resource.UsageRecordIDs, invoice.UUID)

	return invoice, nil
}

// PayInvoice marks an invoice as paid
func (m *WaldurMock) PayInvoice(ctx context.Context, invoiceUUID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	invoice, ok := m.Invoices[invoiceUUID]
	if !ok {
		return fmt.Errorf("invoice not found: %s", invoiceUUID)
	}

	now := time.Now().UTC()
	invoice.State = "paid"
	invoice.PaidAt = &now

	return nil
}

// Helper functions

func matchPath(path, pattern string) bool {
	// Simple pattern matching with * as wildcard
	parts := splitPath(path)
	patternParts := splitPath(pattern)

	if len(parts) != len(patternParts) {
		return false
	}

	for i, p := range patternParts {
		if p != "*" && p != parts[i] {
			return false
		}
	}
	return true
}

func splitPath(path string) []string {
	parts := make([]string, 0)
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func extractUUID(path, prefix, suffix string) string {
	path = path[len(prefix):]
	if suffix != "" {
		path = path[:len(path)-len(suffix)]
	}
	// Remove trailing slash if present
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}
