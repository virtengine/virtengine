//go:build e2e.integration

// Package mocks provides enhanced mock implementations for Waldur E2E testing.
//
// VE-25M: Enhanced Waldur mock for comprehensive E2E testing
package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Extended Mock Capabilities
// ============================================================================

// WaldurProviderMock extends WaldurMock with provider-specific functionality
type WaldurProviderMock struct {
	*WaldurMock

	// Provider-specific data
	ProviderOfferings map[string]*MockProviderOffering
	SyncRecords       map[string]*MockSyncRecord
	UsageAggregates   map[string]*MockUsageAggregate

	// Callback tracking
	CallbackHistory []WaldurCallback

	// Provider configuration
	ProviderConfig ProviderMockConfig
}

// ProviderMockConfig extends mock configuration for provider scenarios
type ProviderMockConfig struct {
	WaldurMockConfig

	// Sync settings
	SyncEnabled    bool
	SyncInterval   time.Duration
	SyncRetryLimit int

	// Callback settings
	CallbackEnabled bool
	CallbackTimeout time.Duration

	// Metering settings
	MeteringEnabled  bool
	MeteringInterval time.Duration
}

// DefaultProviderMockConfig returns default provider mock configuration
func DefaultProviderMockConfig() ProviderMockConfig {
	return ProviderMockConfig{
		WaldurMockConfig: DefaultWaldurMockConfig(),
		SyncEnabled:      true,
		SyncInterval:     100 * time.Millisecond,
		SyncRetryLimit:   3,
		CallbackEnabled:  true,
		CallbackTimeout:  5 * time.Second,
		MeteringEnabled:  true,
		MeteringInterval: 100 * time.Millisecond,
	}
}

// MockProviderOffering represents a provider's offering with sync metadata
type MockProviderOffering struct {
	MockWaldurOffering

	// Sync metadata
	ChainOfferingID string    `json:"chain_offering_id"`
	SyncState       string    `json:"sync_state"`
	SyncVersion     int       `json:"sync_version"`
	LastSyncedAt    time.Time `json:"last_synced_at"`
	SyncChecksum    string    `json:"sync_checksum"`

	// Provider metadata
	ProviderAddress string                 `json:"provider_address"`
	ProviderConfig  map[string]interface{} `json:"provider_config"`

	// Additional fields for offerings
	Type     string `json:"type"`
	Shared   bool   `json:"shared"`
	Billable bool   `json:"billable"`
}

// MockSyncRecord tracks sync state between chain and Waldur
type MockSyncRecord struct {
	EntityType   string    `json:"entity_type"`
	EntityID     string    `json:"entity_id"`
	WaldurID     string    `json:"waldur_id"`
	State        string    `json:"state"`
	SyncVersion  int       `json:"sync_version"`
	ChainVersion int       `json:"chain_version"`
	LastSyncAt   time.Time `json:"last_sync_at"`
	LastError    string    `json:"last_error"`
	RetryCount   int       `json:"retry_count"`
}

// MockUsageAggregate aggregates usage records for billing
type MockUsageAggregate struct {
	ResourceUUID     string                 `json:"resource_uuid"`
	PeriodStart      time.Time              `json:"period_start"`
	PeriodEnd        time.Time              `json:"period_end"`
	TotalUsage       map[string]float64     `json:"total_usage"`
	RecordCount      int                    `json:"record_count"`
	BilledAmount     string                 `json:"billed_amount"`
	SettlementState  string                 `json:"settlement_state"`
	EscrowAllocation string                 `json:"escrow_allocation"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// WaldurCallback represents a callback from Waldur to the chain
type WaldurCallback struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	ResourceUUID   string                 `json:"resource_uuid"`
	OrderUUID      string                 `json:"order_uuid"`
	Action         string                 `json:"action"`
	State          string                 `json:"state"`
	Success        bool                   `json:"success"`
	Payload        map[string]interface{} `json:"payload"`
	Timestamp      time.Time              `json:"timestamp"`
	Signature      string                 `json:"signature"`
	IdempotencyKey string                 `json:"idempotency_key"`
}

// NewWaldurProviderMock creates an enhanced mock for provider testing
func NewWaldurProviderMock() *WaldurProviderMock {
	return NewWaldurProviderMockWithConfig(DefaultProviderMockConfig())
}

// NewWaldurProviderMockWithConfig creates provider mock with config
func NewWaldurProviderMockWithConfig(config ProviderMockConfig) *WaldurProviderMock {
	baseMock := NewWaldurMockWithConfig(config.WaldurMockConfig)

	return &WaldurProviderMock{
		WaldurMock:        baseMock,
		ProviderOfferings: make(map[string]*MockProviderOffering),
		SyncRecords:       make(map[string]*MockSyncRecord),
		UsageAggregates:   make(map[string]*MockUsageAggregate),
		CallbackHistory:   make([]WaldurCallback, 0),
		ProviderConfig:    config,
	}
}

// ============================================================================
// Provider Offering Management
// ============================================================================

// RegisterProviderOffering registers a provider offering with sync metadata
func (m *WaldurProviderMock) RegisterProviderOffering(offering *MockProviderOffering) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if offering.UUID == "" {
		offering.UUID = uuid.New().String()
	}
	if offering.State == "" {
		offering.State = "active"
	}
	if offering.SyncState == "" {
		offering.SyncState = "pending"
	}
	if offering.CreatedAt.IsZero() {
		offering.CreatedAt = time.Now().UTC()
	}

	m.ProviderOfferings[offering.UUID] = offering

	// Also register in base offerings
	m.Offerings[offering.UUID] = &offering.MockWaldurOffering
}

// GetProviderOffering retrieves a provider offering
func (m *WaldurProviderMock) GetProviderOffering(uuid string) *MockProviderOffering {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ProviderOfferings[uuid]
}

// GetProviderOfferingByChainID finds offering by chain ID
func (m *WaldurProviderMock) GetProviderOfferingByChainID(chainID string) *MockProviderOffering {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, o := range m.ProviderOfferings {
		if o.ChainOfferingID == chainID {
			return o
		}
	}
	return nil
}

// SyncOfferingFromChain simulates syncing an offering from chain
func (m *WaldurProviderMock) SyncOfferingFromChain(chainOfferingID string, data map[string]interface{}) (*MockProviderOffering, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already synced
	for _, o := range m.ProviderOfferings {
		if o.ChainOfferingID == chainOfferingID {
			// Update existing
			o.SyncVersion++
			o.LastSyncedAt = time.Now().UTC()
			o.SyncState = "synced"

			if name, ok := data["name"].(string); ok {
				o.Name = name
			}
			if desc, ok := data["description"].(string); ok {
				o.Description = desc
			}

			return o, nil
		}
	}

	// Create new
	offering := &MockProviderOffering{
		MockWaldurOffering: MockWaldurOffering{
			UUID:         uuid.New().String(),
			Name:         fmt.Sprintf("Chain Offering %s", chainOfferingID),
			State:        "active",
			CreatedAt:    time.Now().UTC(),
			CustomerUUID: m.Config.CustomerUUID,
		},
		ChainOfferingID: chainOfferingID,
		SyncState:       "synced",
		SyncVersion:     1,
		LastSyncedAt:    time.Now().UTC(),
	}

	if name, ok := data["name"].(string); ok {
		offering.Name = name
	}
	if desc, ok := data["description"].(string); ok {
		offering.Description = desc
	}
	if provider, ok := data["provider_address"].(string); ok {
		offering.ProviderAddress = provider
	}

	m.ProviderOfferings[offering.UUID] = offering
	m.Offerings[offering.UUID] = &offering.MockWaldurOffering

	return offering, nil
}

// ============================================================================
// Sync Record Management
// ============================================================================

// CreateSyncRecord creates a sync record
func (m *WaldurProviderMock) CreateSyncRecord(entityType, entityID, waldurID string) *MockSyncRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", entityType, entityID)
	record := &MockSyncRecord{
		EntityType:   entityType,
		EntityID:     entityID,
		WaldurID:     waldurID,
		State:        "pending",
		SyncVersion:  0,
		ChainVersion: 1,
		LastSyncAt:   time.Now().UTC(),
	}

	m.SyncRecords[key] = record
	return record
}

// GetSyncRecord retrieves a sync record
func (m *WaldurProviderMock) GetSyncRecord(entityType, entityID string) *MockSyncRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%s", entityType, entityID)
	return m.SyncRecords[key]
}

// UpdateSyncRecord updates sync record state
func (m *WaldurProviderMock) UpdateSyncRecord(entityType, entityID, state string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", entityType, entityID)
	record, ok := m.SyncRecords[key]
	if !ok {
		return fmt.Errorf("sync record not found: %s", key)
	}

	record.State = state
	record.SyncVersion++
	record.LastSyncAt = time.Now().UTC()

	return nil
}

// ============================================================================
// Usage Aggregation
// ============================================================================

// AggregateUsage aggregates usage records for a resource
func (m *WaldurProviderMock) AggregateUsage(resourceUUID string, periodStart, periodEnd time.Time) *MockUsageAggregate {
	m.mu.Lock()
	defer m.mu.Unlock()

	records := m.UsageRecords[resourceUUID]
	if len(records) == 0 {
		return nil
	}

	aggregate := &MockUsageAggregate{
		ResourceUUID:    resourceUUID,
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		TotalUsage:      make(map[string]float64),
		RecordCount:     0,
		SettlementState: "pending",
	}

	for _, record := range records {
		// Check if record falls within period
		if record.PeriodStart.After(periodEnd) || record.PeriodEnd.Before(periodStart) {
			continue
		}

		aggregate.RecordCount++
		for key, value := range record.Metrics {
			if v, ok := value.(float64); ok {
				aggregate.TotalUsage[key] += v
			}
		}
	}

	key := fmt.Sprintf("%s:%d:%d", resourceUUID, periodStart.Unix(), periodEnd.Unix())
	m.UsageAggregates[key] = aggregate

	return aggregate
}

// GetUsageAggregate retrieves usage aggregate
func (m *WaldurProviderMock) GetUsageAggregate(resourceUUID string, periodStart, periodEnd time.Time) *MockUsageAggregate {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key := fmt.Sprintf("%s:%d:%d", resourceUUID, periodStart.Unix(), periodEnd.Unix())
	return m.UsageAggregates[key]
}

// SettleUsage marks usage as settled
func (m *WaldurProviderMock) SettleUsage(resourceUUID string, periodStart, periodEnd time.Time, billedAmount, escrowAllocation string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%d:%d", resourceUUID, periodStart.Unix(), periodEnd.Unix())
	aggregate, ok := m.UsageAggregates[key]
	if !ok {
		return fmt.Errorf("usage aggregate not found")
	}

	aggregate.BilledAmount = billedAmount
	aggregate.EscrowAllocation = escrowAllocation
	aggregate.SettlementState = "settled"

	return nil
}

// ============================================================================
// Callback Management
// ============================================================================

// RecordCallback records a callback event
func (m *WaldurProviderMock) RecordCallback(callback WaldurCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if callback.ID == "" {
		callback.ID = uuid.New().String()
	}
	if callback.Timestamp.IsZero() {
		callback.Timestamp = time.Now().UTC()
	}

	m.CallbackHistory = append(m.CallbackHistory, callback)
}

// GetCallbackHistory retrieves callback history
func (m *WaldurProviderMock) GetCallbackHistory() []WaldurCallback {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]WaldurCallback, len(m.CallbackHistory))
	copy(result, m.CallbackHistory)
	return result
}

// GetCallbacksByResource retrieves callbacks for a resource
func (m *WaldurProviderMock) GetCallbacksByResource(resourceUUID string) []WaldurCallback {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []WaldurCallback
	for _, cb := range m.CallbackHistory {
		if cb.ResourceUUID == resourceUUID {
			result = append(result, cb)
		}
	}
	return result
}

// SimulateProvisionCallback simulates a provision completion callback
func (m *WaldurProviderMock) SimulateProvisionCallback(orderUUID string, success bool, backendID string) {
	m.mu.Lock()
	storedOrder := m.Orders[orderUUID]
	m.mu.Unlock()

	if storedOrder == nil {
		return
	}

	payload := map[string]interface{}{
		"backend_id": backendID,
	}
	if !success {
		payload["error"] = "Provisioning failed"
	}

	callback := WaldurCallback{
		Type:         "provision",
		OrderUUID:    orderUUID,
		ResourceUUID: storedOrder.ResourceUUID,
		Action:       "provision",
		State:        "OK",
		Success:      success,
		Payload:      payload,
		Timestamp:    time.Now().UTC(),
	}

	if !success {
		callback.State = "Erred"
	}

	m.RecordCallback(callback)
}

// SimulateTerminateCallback simulates a terminate completion callback
func (m *WaldurProviderMock) SimulateTerminateCallback(resourceUUID string, success bool) {
	callback := WaldurCallback{
		Type:         "terminate",
		ResourceUUID: resourceUUID,
		Action:       "terminate",
		State:        "Terminated",
		Success:      success,
		Payload:      map[string]interface{}{},
		Timestamp:    time.Now().UTC(),
	}

	if !success {
		callback.State = "Erred"
		callback.Payload["error"] = "Termination failed"
	}

	m.RecordCallback(callback)
}

// SimulateUsageCallback simulates a usage report callback
func (m *WaldurProviderMock) SimulateUsageCallback(resourceUUID string, periodStart, periodEnd time.Time, usage map[string]interface{}) {
	callback := WaldurCallback{
		Type:         "usage_report",
		ResourceUUID: resourceUUID,
		Action:       "usage_report",
		State:        "reported",
		Success:      true,
		Payload: map[string]interface{}{
			"period_start": periodStart.Format(time.RFC3339),
			"period_end":   periodEnd.Format(time.RFC3339),
			"usage":        usage,
		},
		Timestamp: time.Now().UTC(),
	}

	m.RecordCallback(callback)
}

// ============================================================================
// Statistics and Metrics
// ============================================================================

// MockStats contains mock statistics
type MockStats struct {
	TotalOrders         int
	CompletedOrders     int
	FailedOrders        int
	ActiveResources     int
	TerminatedResources int
	TotalUsageRecords   int
	TotalInvoices       int
	PaidInvoices        int
	TotalCallbacks      int
	SyncedOfferings     int
	PendingSyncRecords  int
}

// GetStats returns mock statistics
func (m *WaldurProviderMock) GetStats() MockStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := MockStats{
		TotalOrders:        len(m.Orders),
		TotalInvoices:      len(m.Invoices),
		TotalCallbacks:     len(m.CallbackHistory),
		SyncedOfferings:    0,
		PendingSyncRecords: 0,
	}

	for _, o := range m.Orders {
		switch o.State {
		case "done", "completed":
			stats.CompletedOrders++
		case "erred", "failed":
			stats.FailedOrders++
		}
	}

	for _, r := range m.Resources {
		switch r.State {
		case "provisioned", "active", "OK":
			stats.ActiveResources++
		case "terminated":
			stats.TerminatedResources++
		}
	}

	for _, records := range m.UsageRecords {
		stats.TotalUsageRecords += len(records)
	}

	for _, inv := range m.Invoices {
		if inv.State == "paid" {
			stats.PaidInvoices++
		}
	}

	for _, o := range m.ProviderOfferings {
		if o.SyncState == "synced" {
			stats.SyncedOfferings++
		}
	}

	for _, r := range m.SyncRecords {
		if r.State == "pending" {
			stats.PendingSyncRecords++
		}
	}

	return stats
}

// ============================================================================
// Mock Scenario Helpers
// ============================================================================

// ScenarioConfig defines a test scenario configuration
type ScenarioConfig struct {
	NumProviders   int
	NumOfferings   int
	NumCustomers   int
	NumOrders      int
	AutoProvision  bool
	IncludeErrors  bool
	UsageFrequency time.Duration
}

// SetupScenario sets up a complete test scenario
func (m *WaldurProviderMock) SetupScenario(config ScenarioConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create offerings
	for i := 0; i < config.NumOfferings; i++ {
		offering := &MockProviderOffering{
			MockWaldurOffering: MockWaldurOffering{
				UUID:         fmt.Sprintf("offering-scenario-%d", i),
				Name:         fmt.Sprintf("Scenario Offering %d", i),
				Category:     "compute",
				State:        "active",
				CustomerUUID: m.Config.CustomerUUID,
				PricePerHour: fmt.Sprintf("%.2f", 0.50+float64(i)*0.10),
				CreatedAt:    time.Now().UTC(),
			},
			ChainOfferingID: fmt.Sprintf("ve1scenario/offering-%d", i),
			SyncState:       "synced",
			SyncVersion:     1,
			LastSyncedAt:    time.Now().UTC(),
			ProviderAddress: fmt.Sprintf("ve1provider%d", i%config.NumProviders),
		}
		m.ProviderOfferings[offering.UUID] = offering
		m.Offerings[offering.UUID] = &offering.MockWaldurOffering
	}
}

// Reset clears all mock data
func (m *WaldurProviderMock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Resources = make(map[string]*MockWaldurResource)
	m.Orders = make(map[string]*MockWaldurOrder)
	m.Offerings = make(map[string]*MockWaldurOffering)
	m.UsageRecords = make(map[string][]*MockWaldurUsageRecord)
	m.Invoices = make(map[string]*MockWaldurInvoice)
	m.Callbacks = make([]CallbackRequest, 0)
	m.ProviderOfferings = make(map[string]*MockProviderOffering)
	m.SyncRecords = make(map[string]*MockSyncRecord)
	m.UsageAggregates = make(map[string]*MockUsageAggregate)
	m.CallbackHistory = make([]WaldurCallback, 0)
	m.ErrorState = nil
}

// ============================================================================
// HTTP Handler Extensions
// ============================================================================

// ExtendedWaldurMock adds additional HTTP handlers
type ExtendedWaldurMock struct {
	*WaldurProviderMock
	extMu sync.RWMutex
}

// NewExtendedWaldurMock creates an extended mock with additional endpoints
func NewExtendedWaldurMock() *ExtendedWaldurMock {
	providerMock := NewWaldurProviderMock()

	extended := &ExtendedWaldurMock{
		WaldurProviderMock: providerMock,
	}

	// Close the default server and create extended one
	if providerMock.Server != nil {
		providerMock.Server.Close()
	}
	providerMock.Server = extended.createExtendedServer()

	return extended
}

func (m *ExtendedWaldurMock) createExtendedServer() *httptest.Server {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/api/health/", m.handleHealth)

	// Marketplace endpoints
	mux.HandleFunc("/api/marketplace-orders/", m.handleOrders)
	mux.HandleFunc("/api/marketplace-offerings/", m.handleOfferings)
	mux.HandleFunc("/api/marketplace-public-offerings/", m.handlePublicOfferings)
	mux.HandleFunc("/api/marketplace-provider-offerings/", m.handleProviderOfferings)
	mux.HandleFunc("/api/marketplace-resources/", m.handleResources)
	mux.HandleFunc("/api/marketplace-component-usages/", m.handleUsages)

	// Stats endpoint
	mux.HandleFunc("/api/stats/", m.handleStats)

	// Create test server
	return httptest.NewServer(mux)
}

func (m *ExtendedWaldurMock) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"stats":     m.GetStats(),
	})
}

func (m *ExtendedWaldurMock) handleOrders(w http.ResponseWriter, r *http.Request) {
	m.WaldurMock.handleRequest(w, r)
}

func (m *ExtendedWaldurMock) handleOfferings(w http.ResponseWriter, r *http.Request) {
	m.WaldurMock.handleRequest(w, r)
}

func (m *ExtendedWaldurMock) handlePublicOfferings(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Handle backend_id filter
	if backendID := r.URL.Query().Get("backend_id"); backendID != "" {
		offering := m.GetOfferingByBackendID(backendID)
		if offering == nil {
			// Return empty array for not found
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]interface{}{})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]interface{}{offering})
		return
	}

	// Return all offerings
	offerings := make([]*MockWaldurOffering, 0, len(m.Offerings))
	for _, o := range m.Offerings {
		offerings = append(offerings, o)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(offerings)
}

func (m *ExtendedWaldurMock) handleProviderOfferings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		m.handleCreateProviderOffering(w, r)
	case http.MethodPatch:
		m.handleUpdateProviderOffering(w, r)
	default:
		m.handlePublicOfferings(w, r)
	}
}

func (m *ExtendedWaldurMock) handleCreateProviderOffering(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name, _ := req["name"].(string)
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	customerUUID, _ := req["customer"].(string)
	if customerUUID == "" {
		http.Error(w, "customer UUID is required", http.StatusBadRequest)
		return
	}

	offering := &MockProviderOffering{
		MockWaldurOffering: MockWaldurOffering{
			UUID:         uuid.New().String(),
			Name:         name,
			CustomerUUID: customerUUID,
			State:        "active",
			CreatedAt:    time.Now().UTC(),
		},
	}

	if desc, ok := req["description"].(string); ok {
		offering.Description = desc
	}
	if backendID, ok := req["backend_id"].(string); ok {
		offering.BackendID = backendID
		offering.ChainOfferingID = backendID
	}
	if shared, ok := req["shared"].(bool); ok {
		offering.Shared = shared
	}
	if billable, ok := req["billable"].(bool); ok {
		offering.Billable = billable
	}

	m.RegisterProviderOffering(offering)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"uuid":        offering.UUID,
		"name":        offering.Name,
		"description": offering.Description,
		"type":        offering.Type,
		"state":       offering.State,
		"shared":      offering.Shared,
		"billable":    offering.Billable,
	})
}

func (m *ExtendedWaldurMock) handleUpdateProviderOffering(w http.ResponseWriter, r *http.Request) {
	// Extract UUID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/marketplace-provider-offerings/")
	offeringUUID := strings.TrimSuffix(path, "/")

	offering := m.GetProviderOffering(offeringUUID)
	if offering == nil {
		http.NotFound(w, r)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if name, ok := req["name"].(string); ok && name != "" {
		offering.Name = name
	}
	if desc, ok := req["description"].(string); ok {
		offering.Description = desc
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"uuid":        offering.UUID,
		"name":        offering.Name,
		"description": offering.Description,
		"type":        offering.Type,
		"state":       offering.State,
		"shared":      offering.Shared,
		"billable":    offering.Billable,
	})
}

func (m *ExtendedWaldurMock) handleResources(w http.ResponseWriter, r *http.Request) {
	m.WaldurMock.handleRequest(w, r)
}

func (m *ExtendedWaldurMock) handleUsages(w http.ResponseWriter, r *http.Request) {
	m.WaldurMock.handleRequest(w, r)
}

func (m *ExtendedWaldurMock) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := m.GetStats()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}
