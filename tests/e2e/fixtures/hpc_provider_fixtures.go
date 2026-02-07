//go:build e2e.integration

// Package fixtures provides test fixtures for E2E tests.
//
// VE-8C: Test fixtures for HPC marketplace provider flow E2E tests.
package fixtures

import (
	"context"
	"fmt"
	"math"
	"os"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	pd "github.com/virtengine/virtengine/pkg/provider_daemon"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCProviderFixture contains all components needed for HPC provider E2E tests.
type HPCProviderFixture struct {
	t         *testing.T
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	cleanupFn []func()

	// Configuration
	Config HPCProviderFixtureConfig

	// Core components
	ProviderAddress string
	CustomerAddress string
	ClusterID       string

	// Scheduler mock
	Scheduler *MockScheduler

	// Settlement mock
	Settlement *MockSettlement

	// Waldur mock
	Waldur *MockWaldur

	// Usage reporter mock
	UsageReporter *MockUsageReporter

	// Generated test data
	Jobs        map[string]*hpctypes.HPCJob
	Orders      map[string]*MockOrder
	Allocations map[string]*MockAllocation
	Invoices    map[string]*MockInvoice
}

// HPCProviderFixtureConfig configures the test fixture.
type HPCProviderFixtureConfig struct {
	// ProviderAddress is the test provider address
	ProviderAddress string

	// CustomerAddress is the test customer address
	CustomerAddress string

	// ClusterID is the test cluster ID
	ClusterID string

	// NumOfferings is the number of offerings to create
	NumOfferings int

	// NumOrders is the number of orders to create
	NumOrders int

	// NumJobs is the number of jobs to create
	NumJobs int

	// EnableMockScheduler enables the mock scheduler
	EnableMockScheduler bool

	// EnableMockWaldur enables the mock Waldur client
	EnableMockWaldur bool

	// EnableMockSettlement enables the mock settlement pipeline
	EnableMockSettlement bool

	// JobTimeoutSeconds is the max job runtime
	JobTimeoutSeconds int64

	// CleanupOnTeardown removes all test data on teardown
	CleanupOnTeardown bool
}

// DefaultFixtureConfig returns a default fixture configuration.
func DefaultFixtureConfig() HPCProviderFixtureConfig {
	return HPCProviderFixtureConfig{
		ProviderAddress:      sdk.AccAddress([]byte("provider-fixture-1234")).String(),
		CustomerAddress:      sdk.AccAddress([]byte("customer-fixture-1234")).String(),
		ClusterID:            "e2e-fixture-cluster",
		NumOfferings:         3,
		NumOrders:            2,
		NumJobs:              5,
		EnableMockScheduler:  true,
		EnableMockWaldur:     true,
		EnableMockSettlement: true,
		JobTimeoutSeconds:    3600,
		CleanupOnTeardown:    true,
	}
}

// NewHPCProviderFixture creates a new test fixture.
func NewHPCProviderFixture(t *testing.T, config HPCProviderFixtureConfig) *HPCProviderFixture {
	ctx, cancel := context.WithCancel(context.Background())

	f := &HPCProviderFixture{
		t:               t,
		ctx:             ctx,
		cancel:          cancel,
		Config:          config,
		ProviderAddress: config.ProviderAddress,
		CustomerAddress: config.CustomerAddress,
		ClusterID:       config.ClusterID,
		Jobs:            make(map[string]*hpctypes.HPCJob),
		Orders:          make(map[string]*MockOrder),
		Allocations:     make(map[string]*MockAllocation),
		Invoices:        make(map[string]*MockInvoice),
	}

	// Initialize mocks
	if config.EnableMockScheduler {
		f.Scheduler = NewMockScheduler(config.ClusterID)
	}
	if config.EnableMockWaldur {
		f.Waldur = NewMockWaldur()
	}
	if config.EnableMockSettlement {
		f.Settlement = NewMockSettlement()
	}
	f.UsageReporter = NewMockUsageReporter()

	return f
}

// Setup initializes all fixture components.
func (f *HPCProviderFixture) Setup() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Start scheduler
	if f.Scheduler != nil {
		if err := f.Scheduler.Start(f.ctx); err != nil {
			return fmt.Errorf("start scheduler: %w", err)
		}
		f.addCleanup(func() { _ = f.Scheduler.Stop() })
	}

	// Register provider in Waldur
	if f.Waldur != nil {
		f.Waldur.RegisterProvider(f.ProviderAddress)
	}

	// Create default offerings
	f.createDefaultOfferings()

	return nil
}

// Teardown cleans up all fixture resources.
func (f *HPCProviderFixture) Teardown() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.cancel()

	if f.Config.CleanupOnTeardown {
		// Run cleanup functions in reverse order
		for i := len(f.cleanupFn) - 1; i >= 0; i-- {
			f.cleanupFn[i]()
		}
	}

	f.Jobs = make(map[string]*hpctypes.HPCJob)
	f.Orders = make(map[string]*MockOrder)
	f.Allocations = make(map[string]*MockAllocation)
	f.Invoices = make(map[string]*MockInvoice)
}

func (f *HPCProviderFixture) addCleanup(fn func()) {
	f.cleanupFn = append(f.cleanupFn, fn)
}

// CreateJob creates a test HPC job.
func (f *HPCProviderFixture) CreateJob(name string, cpuCores int32, memoryMB int64, gpus int32) *hpctypes.HPCJob {
	jobID := fmt.Sprintf("job-%s-%d", name, time.Now().UnixNano())
	memoryGB := memoryMB / 1024
	if memoryGB > math.MaxInt32 {
		memoryGB = math.MaxInt32
	}

	job := &hpctypes.HPCJob{
		JobID:           jobID,
		ClusterID:       f.ClusterID,
		OfferingID:      "hpc-compute-medium",
		ProviderAddress: f.ProviderAddress,
		CustomerAddress: f.CustomerAddress,
		State:           hpctypes.JobStatePending,
		QueueName:       "default",
		WorkloadSpec: hpctypes.JobWorkloadSpec{
			ContainerImage: "alpine:latest",
			Command:        fmt.Sprintf("echo 'Job %s running' && sleep 10", name),
		},
		Resources: hpctypes.JobResources{
			Nodes:           1,
			CPUCoresPerNode: cpuCores,
			MemoryGBPerNode: int32(memoryGB), // #nosec G115 -- bounded above
			GPUsPerNode:     gpus,
			StorageGB:       10,
		},
		MaxRuntimeSeconds: f.Config.JobTimeoutSeconds,
		CreatedAt:         time.Now(),
	}

	f.Jobs[jobID] = job
	return job
}

// SubmitJob submits a job to the mock scheduler.
func (f *HPCProviderFixture) SubmitJob(job *hpctypes.HPCJob) (*pd.HPCSchedulerJob, error) {
	if f.Scheduler == nil {
		return nil, fmt.Errorf("scheduler not initialized")
	}
	return f.Scheduler.SubmitJob(f.ctx, job)
}

// CreateOrder creates a test order.
func (f *HPCProviderFixture) CreateOrder(offeringID string, maxPrice string) *MockOrder {
	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())

	order := &MockOrder{
		OrderID:      orderID,
		CustomerAddr: f.CustomerAddress,
		OfferingID:   offeringID,
		MaxPrice:     maxPrice,
		Status:       "open",
		CreatedAt:    time.Now(),
	}

	f.Orders[orderID] = order

	if f.Waldur != nil {
		f.Waldur.CreateOrder(order)
	}

	return order
}

// PlaceBid places a bid on an order.
func (f *HPCProviderFixture) PlaceBid(orderID string, price string) (*MockBid, error) {
	if f.Waldur == nil {
		return nil, fmt.Errorf("waldur not initialized")
	}

	bid := &MockBid{
		BidID:        fmt.Sprintf("bid-%d", time.Now().UnixNano()),
		OrderID:      orderID,
		ProviderAddr: f.ProviderAddress,
		Price:        price,
		CreatedAt:    time.Now(),
	}

	f.Waldur.PlaceBid(bid)
	return bid, nil
}

// AcceptBid accepts a bid and creates an allocation.
func (f *HPCProviderFixture) AcceptBid(orderID, bidID string) (*MockAllocation, error) {
	if f.Waldur == nil {
		return nil, fmt.Errorf("waldur not initialized")
	}

	allocation, err := f.Waldur.AcceptBid(orderID, bidID)
	if err != nil {
		return nil, err
	}

	f.Allocations[allocation.AllocationID] = allocation
	return allocation, nil
}

// GenerateUsageRecord generates a usage record for a job.
func (f *HPCProviderFixture) GenerateUsageRecord(jobID string, metrics *pd.HPCSchedulerMetrics) *pd.HPCUsageRecord {
	record := &pd.HPCUsageRecord{
		RecordID:        fmt.Sprintf("usage-%d", time.Now().UnixNano()),
		JobID:           jobID,
		ClusterID:       f.ClusterID,
		ProviderAddress: f.ProviderAddress,
		CustomerAddress: f.CustomerAddress,
		PeriodStart:     time.Now().Add(-time.Hour),
		PeriodEnd:       time.Now(),
		Metrics:         metrics,
		IsFinal:         true,
		JobState:        pd.HPCJobStateCompleted,
		Timestamp:       time.Now(),
	}

	f.UsageReporter.RecordUsage(record)
	return record
}

// CreateInvoice creates an invoice for settlement testing.
func (f *HPCProviderFixture) CreateInvoice(orderID string, lineItems []MockLineItem, totalAmount string) *MockInvoice {
	invoice := &MockInvoice{
		InvoiceID:    fmt.Sprintf("invoice-%d", time.Now().UnixNano()),
		ProviderAddr: f.ProviderAddress,
		CustomerAddr: f.CustomerAddress,
		OrderID:      orderID,
		LineItems:    lineItems,
		TotalAmount:  totalAmount,
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	f.Invoices[invoice.InvoiceID] = invoice

	if f.Settlement != nil {
		f.Settlement.CreateInvoice(invoice)
	}

	return invoice
}

func (f *HPCProviderFixture) createDefaultOfferings() {
	if f.Waldur == nil {
		return
	}

	offerings := []MockOffering{
		{
			OfferingID:   "hpc-compute-small",
			Name:         "HPC Compute Small",
			Category:     "compute",
			CPUCores:     16,
			MemoryGB:     64,
			GPUs:         0,
			PricePerHour: "5.0",
			Active:       true,
		},
		{
			OfferingID:   "hpc-compute-medium",
			Name:         "HPC Compute Medium",
			Category:     "compute",
			CPUCores:     64,
			MemoryGB:     256,
			GPUs:         2,
			PricePerHour: "20.0",
			Active:       true,
		},
		{
			OfferingID:   "hpc-gpu-large",
			Name:         "HPC GPU Large",
			Category:     "gpu",
			CPUCores:     128,
			MemoryGB:     512,
			GPUs:         8,
			PricePerHour: "100.0",
			Active:       true,
		},
	}

	for _, o := range offerings[:min(f.Config.NumOfferings, len(offerings))] {
		f.Waldur.PublishOffering(o)
	}
}

// GetEnvOrDefault gets an environment variable or returns a default value.
func GetEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// =============================================================================
// Mock Components for Fixtures
// =============================================================================

// MockScheduler is a mock HPC scheduler for fixture tests.
type MockScheduler struct {
	clusterID   string
	running     bool
	jobs        map[string]*pd.HPCSchedulerJob
	metrics     map[string]*pd.HPCSchedulerMetrics
	maxCPU      int32
	maxMemoryMB int64
	maxGPUs     int32
	mu          sync.RWMutex
}

// NewMockScheduler creates a new mock scheduler.
func NewMockScheduler(clusterID string) *MockScheduler {
	return &MockScheduler{
		clusterID:   clusterID,
		jobs:        make(map[string]*pd.HPCSchedulerJob),
		metrics:     make(map[string]*pd.HPCSchedulerMetrics),
		maxCPU:      1000,
		maxMemoryMB: 1024 * 1024,
		maxGPUs:     100,
	}
}

func (m *MockScheduler) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = true
	return nil
}

func (m *MockScheduler) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	return nil
}

func (m *MockScheduler) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

func (m *MockScheduler) SubmitJob(ctx context.Context, job *hpctypes.HPCJob) (*pd.HPCSchedulerJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job.Resources.CPUCoresPerNode > m.maxCPU {
		return nil, fmt.Errorf("insufficient CPU: requested %d, available %d", job.Resources.CPUCoresPerNode, m.maxCPU)
	}
	if int64(job.Resources.MemoryGBPerNode)*1024 > m.maxMemoryMB {
		return nil, fmt.Errorf("insufficient memory: requested %d MB, available %d MB", job.Resources.MemoryGBPerNode*1024, m.maxMemoryMB)
	}

	schedulerJob := &pd.HPCSchedulerJob{
		VirtEngineJobID: job.JobID,
		SchedulerJobID:  fmt.Sprintf("slurm-%s", job.JobID),
		SchedulerType:   pd.HPCSchedulerTypeSLURM,
		State:           pd.HPCJobStatePending,
		SubmitTime:      time.Now(),
		OriginalJob:     job,
	}

	m.jobs[job.JobID] = schedulerJob
	m.metrics[job.JobID] = &pd.HPCSchedulerMetrics{}

	return schedulerJob, nil
}

func (m *MockScheduler) GetJobStatus(ctx context.Context, jobID string) (*pd.HPCSchedulerJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if job, ok := m.jobs[jobID]; ok {
		return job, nil
	}
	return nil, fmt.Errorf("job not found: %s", jobID)
}

func (m *MockScheduler) GetJobAccounting(ctx context.Context, jobID string) (*pd.HPCSchedulerMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if metrics, ok := m.metrics[jobID]; ok {
		return metrics, nil
	}
	return nil, fmt.Errorf("job not found: %s", jobID)
}

func (m *MockScheduler) SetJobState(jobID string, state pd.HPCJobState) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job, ok := m.jobs[jobID]; ok {
		job.State = state
		if state.IsTerminal() {
			now := time.Now()
			job.EndTime = &now
		}
	}
}

func (m *MockScheduler) SetJobMetrics(jobID string, metrics *pd.HPCSchedulerMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[jobID] = metrics
}

// MockWaldur is a mock Waldur client for fixture tests.
type MockWaldur struct {
	providers   map[string]bool
	offerings   []MockOffering
	orders      map[string]*MockOrder
	bids        map[string][]*MockBid
	allocations map[string]*MockAllocation
	mu          sync.RWMutex
}

// MockOffering represents a marketplace offering.
type MockOffering struct {
	OfferingID   string
	Name         string
	Category     string
	CPUCores     int32
	MemoryGB     int32
	GPUs         int32
	PricePerHour string
	Active       bool
	WaldurUUID   string
}

// MockOrder represents a marketplace order.
type MockOrder struct {
	OrderID      string
	CustomerAddr string
	OfferingID   string
	MaxPrice     string
	Status       string
	CreatedAt    time.Time
}

// MockBid represents a provider bid.
type MockBid struct {
	BidID        string
	OrderID      string
	ProviderAddr string
	Price        string
	CreatedAt    time.Time
}

// MockAllocation represents a resource allocation.
type MockAllocation struct {
	AllocationID string
	OrderID      string
	BidID        string
	ProviderAddr string
	Status       string
	CreatedAt    time.Time
}

// NewMockWaldur creates a new mock Waldur client.
func NewMockWaldur() *MockWaldur {
	return &MockWaldur{
		providers:   make(map[string]bool),
		offerings:   make([]MockOffering, 0),
		orders:      make(map[string]*MockOrder),
		bids:        make(map[string][]*MockBid),
		allocations: make(map[string]*MockAllocation),
	}
}

func (m *MockWaldur) RegisterProvider(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[addr] = true
}

func (m *MockWaldur) IsProviderRegistered(addr string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.providers[addr]
}

func (m *MockWaldur) PublishOffering(offering MockOffering) {
	m.mu.Lock()
	defer m.mu.Unlock()
	offering.WaldurUUID = fmt.Sprintf("waldur-%s", offering.OfferingID)
	m.offerings = append(m.offerings, offering)
}

func (m *MockWaldur) GetOfferings() []MockOffering {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]MockOffering, len(m.offerings))
	copy(result, m.offerings)
	return result
}

func (m *MockWaldur) CreateOrder(order *MockOrder) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[order.OrderID] = order
}

func (m *MockWaldur) GetOrder(orderID string) *MockOrder {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.orders[orderID]
}

func (m *MockWaldur) PlaceBid(bid *MockBid) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bids[bid.OrderID] = append(m.bids[bid.OrderID], bid)
}

func (m *MockWaldur) GetBids(orderID string) []*MockBid {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bids[orderID]
}

func (m *MockWaldur) AcceptBid(orderID, bidID string) (*MockAllocation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	order, ok := m.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	var selectedBid *MockBid
	for _, bid := range m.bids[orderID] {
		if bid.BidID == bidID {
			selectedBid = bid
			break
		}
	}
	if selectedBid == nil {
		return nil, fmt.Errorf("bid not found: %s", bidID)
	}

	order.Status = "matched"

	allocation := &MockAllocation{
		AllocationID: fmt.Sprintf("alloc-%s", orderID),
		OrderID:      orderID,
		BidID:        bidID,
		ProviderAddr: selectedBid.ProviderAddr,
		Status:       "provisioned",
		CreatedAt:    time.Now(),
	}
	m.allocations[allocation.AllocationID] = allocation

	return allocation, nil
}

func (m *MockWaldur) GetAllocation(allocationID string) *MockAllocation {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.allocations[allocationID]
}

// MockSettlement is a mock settlement pipeline for fixture tests.
type MockSettlement struct {
	invoices map[string]*MockInvoice
	payouts  map[string]*MockPayout
	fees     map[string]*MockFee
	mu       sync.RWMutex
}

// MockInvoice represents a billing invoice.
type MockInvoice struct {
	InvoiceID    string
	ProviderAddr string
	CustomerAddr string
	OrderID      string
	LineItems    []MockLineItem
	TotalAmount  string
	Status       string
	CreatedAt    time.Time
	SettledAt    *time.Time
}

// MockLineItem represents an invoice line item.
type MockLineItem struct {
	ResourceType string
	Quantity     sdkmath.LegacyDec
	UnitPrice    string
	TotalCost    string
}

// MockPayout represents a provider payout.
type MockPayout struct {
	PayoutID  string
	InvoiceID string
	Provider  string
	Amount    string
	Status    string
}

// MockFee represents a platform fee.
type MockFee struct {
	FeeID     string
	InvoiceID string
	Amount    string
}

// NewMockSettlement creates a new mock settlement pipeline.
func NewMockSettlement() *MockSettlement {
	return &MockSettlement{
		invoices: make(map[string]*MockInvoice),
		payouts:  make(map[string]*MockPayout),
		fees:     make(map[string]*MockFee),
	}
}

func (m *MockSettlement) CreateInvoice(invoice *MockInvoice) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.invoices[invoice.InvoiceID] = invoice
}

func (m *MockSettlement) GetInvoice(invoiceID string) *MockInvoice {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.invoices[invoiceID]
}

func (m *MockSettlement) SettleInvoice(invoiceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	invoice, ok := m.invoices[invoiceID]
	if !ok {
		return fmt.Errorf("invoice not found: %s", invoiceID)
	}

	if invoice.Status == "disputed" {
		return fmt.Errorf("cannot settle disputed invoice")
	}

	now := time.Now()
	invoice.Status = "settled"
	invoice.SettledAt = &now

	totalAmount, err := sdkmath.LegacyNewDecFromStr(invoice.TotalAmount)
	if err != nil {
		return fmt.Errorf("parse invoice total amount: %w", err)
	}

	feeRate, err := sdkmath.LegacyNewDecFromStr("0.025")
	if err != nil {
		return fmt.Errorf("parse fee rate: %w", err)
	}

	feeAmount := totalAmount.Mul(feeRate)
	payoutAmount := totalAmount.Sub(feeAmount)

	// Create payout
	m.payouts[invoiceID] = &MockPayout{
		PayoutID:  fmt.Sprintf("payout-%s", invoiceID),
		InvoiceID: invoiceID,
		Provider:  invoice.ProviderAddr,
		Amount:    payoutAmount.String(),
		Status:    "completed",
	}

	// Create fee
	m.fees[invoiceID] = &MockFee{
		FeeID:     fmt.Sprintf("fee-%s", invoiceID),
		InvoiceID: invoiceID,
		Amount:    feeAmount.String(),
	}

	return nil
}

func (m *MockSettlement) DisputeInvoice(invoiceID, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	invoice, ok := m.invoices[invoiceID]
	if !ok {
		return fmt.Errorf("invoice not found: %s", invoiceID)
	}

	invoice.Status = "disputed"
	return nil
}

func (m *MockSettlement) GetPayout(invoiceID string) *MockPayout {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.payouts[invoiceID]
}

func (m *MockSettlement) GetFee(invoiceID string) *MockFee {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fees[invoiceID]
}

// MockUsageReporter is a mock usage reporter for fixture tests.
type MockUsageReporter struct {
	records map[string]*pd.HPCUsageRecord
	mu      sync.RWMutex
}

// NewMockUsageReporter creates a new mock usage reporter.
func NewMockUsageReporter() *MockUsageReporter {
	return &MockUsageReporter{
		records: make(map[string]*pd.HPCUsageRecord),
	}
}

func (m *MockUsageReporter) RecordUsage(record *pd.HPCUsageRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[record.RecordID] = record
}

func (m *MockUsageReporter) GetRecord(recordID string) *pd.HPCUsageRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.records[recordID]
}

func (m *MockUsageReporter) GetRecordsForJob(jobID string) []*pd.HPCUsageRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*pd.HPCUsageRecord
	for _, record := range m.records {
		if record.JobID == jobID {
			result = append(result, record)
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
