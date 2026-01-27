// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-401: Provider Daemon bid engine and provider configuration watcher
package provider_daemon

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
)

// ErrBidEngineNotRunning is returned when the bid engine is not running
var ErrBidEngineNotRunning = errors.New("bid engine is not running")

// ErrBidRateLimited is returned when bidding is rate limited
var ErrBidRateLimited = errors.New("bid rate limited")

// ErrOrderNotMatchable is returned when an order doesn't match provider capabilities
var ErrOrderNotMatchable = errors.New("order does not match provider capabilities")

// ErrInvalidPrice is returned when a calculated price is invalid
var ErrInvalidPrice = errors.New("invalid bid price")

// BidEngineConfig configures the bid engine
type BidEngineConfig struct {
	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// MaxBidsPerMinute limits bidding rate
	MaxBidsPerMinute int `json:"max_bids_per_minute"`

	// MaxBidsPerHour limits hourly bidding rate
	MaxBidsPerHour int `json:"max_bids_per_hour"`

	// MaxConcurrentBids limits concurrent bid operations
	MaxConcurrentBids int `json:"max_concurrent_bids"`

	// BidRetryDelay is the delay between retry attempts
	BidRetryDelay time.Duration `json:"bid_retry_delay"`

	// MaxBidRetries is the maximum number of retry attempts
	MaxBidRetries int `json:"max_bid_retries"`

	// ConfigPollInterval is how often to poll for config updates
	ConfigPollInterval time.Duration `json:"config_poll_interval"`

	// OrderPollInterval is how often to poll for new orders
	OrderPollInterval time.Duration `json:"order_poll_interval"`
}

// DefaultBidEngineConfig returns the default bid engine configuration
func DefaultBidEngineConfig() BidEngineConfig {
	return BidEngineConfig{
		MaxBidsPerMinute:   10,
		MaxBidsPerHour:     100,
		MaxConcurrentBids:  5,
		BidRetryDelay:      time.Second * 5,
		MaxBidRetries:      3,
		ConfigPollInterval: time.Second * 30,
		OrderPollInterval:  time.Second * 5,
	}
}

// ProviderConfig represents the provider's on-chain configuration
type ProviderConfig struct {
	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// Pricing contains pricing configuration
	Pricing PricingConfig `json:"pricing"`

	// Capacity contains capacity configuration
	Capacity CapacityConfig `json:"capacity"`

	// SupportedOfferings lists supported offering types
	SupportedOfferings []string `json:"supported_offerings"`

	// Regions lists supported regions
	Regions []string `json:"regions"`

	// Attributes contains provider attributes
	Attributes map[string]string `json:"attributes"`

	// Active indicates if the provider is accepting orders
	Active bool `json:"active"`

	// LastUpdated is when the config was last updated
	LastUpdated time.Time `json:"last_updated"`

	// Version is the config version (for change detection)
	Version uint64 `json:"version"`
}

// PricingConfig contains provider pricing configuration
type PricingConfig struct {
	// CPUPricePerCore is the price per CPU core per hour
	CPUPricePerCore string `json:"cpu_price_per_core"`

	// MemoryPricePerGB is the price per GB memory per hour
	MemoryPricePerGB string `json:"memory_price_per_gb"`

	// StoragePricePerGB is the price per GB storage per hour
	StoragePricePerGB string `json:"storage_price_per_gb"`

	// NetworkPricePerGB is the price per GB network transfer
	NetworkPricePerGB string `json:"network_price_per_gb"`

	// GPUPricePerHour is the price per GPU per hour
	GPUPricePerHour string `json:"gpu_price_per_hour,omitempty"`

	// MinBidPrice is the minimum bid price
	MinBidPrice string `json:"min_bid_price"`

	// BidMarkupPercent is the markup percentage for bids
	BidMarkupPercent float64 `json:"bid_markup_percent"`

	// Currency is the currency for pricing
	Currency string `json:"currency"`
}

// CapacityConfig contains provider capacity configuration
type CapacityConfig struct {
	// TotalCPUCores is the total available CPU cores
	TotalCPUCores int64 `json:"total_cpu_cores"`

	// TotalMemoryGB is the total available memory in GB
	TotalMemoryGB int64 `json:"total_memory_gb"`

	// TotalStorageGB is the total available storage in GB
	TotalStorageGB int64 `json:"total_storage_gb"`

	// TotalGPUs is the total available GPUs
	TotalGPUs int64 `json:"total_gpus,omitempty"`

	// ReservedCPUCores is the reserved CPU cores (not for sale)
	ReservedCPUCores int64 `json:"reserved_cpu_cores"`

	// ReservedMemoryGB is the reserved memory (not for sale)
	ReservedMemoryGB int64 `json:"reserved_memory_gb"`

	// ReservedStorageGB is the reserved storage (not for sale)
	ReservedStorageGB int64 `json:"reserved_storage_gb"`
}

// AvailableCPU returns available CPU cores
func (c CapacityConfig) AvailableCPU() int64 {
	return c.TotalCPUCores - c.ReservedCPUCores
}

// AvailableMemory returns available memory in GB
func (c CapacityConfig) AvailableMemory() int64 {
	return c.TotalMemoryGB - c.ReservedMemoryGB
}

// AvailableStorage returns available storage in GB
func (c CapacityConfig) AvailableStorage() int64 {
	return c.TotalStorageGB - c.ReservedStorageGB
}

// Order represents an order that can be bid on
type Order struct {
	// OrderID is the unique order identifier
	OrderID string `json:"order_id"`

	// CustomerAddress is the customer's address
	CustomerAddress string `json:"customer_address"`

	// OfferingType is the type of offering requested
	OfferingType string `json:"offering_type"`

	// Requirements contains resource requirements
	Requirements ResourceRequirements `json:"requirements"`

	// Region is the requested region (optional)
	Region string `json:"region,omitempty"`

	// MaxPrice is the maximum price the customer will pay
	MaxPrice string `json:"max_price"`

	// Currency is the price currency
	Currency string `json:"currency"`

	// CreatedAt is when the order was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the order expires
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// ResourceRequirements specifies required resources
type ResourceRequirements struct {
	// CPUCores is the required CPU cores
	CPUCores int64 `json:"cpu_cores"`

	// MemoryGB is the required memory in GB
	MemoryGB int64 `json:"memory_gb"`

	// StorageGB is the required storage in GB
	StorageGB int64 `json:"storage_gb"`

	// GPUs is the required GPUs (optional)
	GPUs int64 `json:"gpus,omitempty"`

	// GPUType is the required GPU type (optional)
	GPUType string `json:"gpu_type,omitempty"`
}

// Bid represents a bid on an order
type Bid struct {
	// BidID is the unique bid identifier
	BidID string `json:"bid_id"`

	// OrderID is the order being bid on
	OrderID string `json:"order_id"`

	// ProviderAddress is the provider placing the bid
	ProviderAddress string `json:"provider_address"`

	// Price is the bid price
	Price string `json:"price"`

	// Currency is the price currency
	Currency string `json:"currency"`

	// CreatedAt is when the bid was placed
	CreatedAt time.Time `json:"created_at"`

	// State is the bid state
	State string `json:"state"`
}

// ChainClient interface for interacting with the chain
type ChainClient interface {
	// GetProviderConfig retrieves the provider's on-chain configuration
	GetProviderConfig(ctx context.Context, address string) (*ProviderConfig, error)

	// GetOpenOrders retrieves open orders that match provider capabilities
	GetOpenOrders(ctx context.Context, offeringTypes []string, regions []string) ([]Order, error)

	// PlaceBid places a bid on an order
	PlaceBid(ctx context.Context, bid *Bid, signature *Signature) error

	// GetProviderBids retrieves bids placed by this provider
	GetProviderBids(ctx context.Context, address string) ([]Bid, error)
}

// RateLimiter tracks bid rate limits
type RateLimiter struct {
	maxPerMinute int
	maxPerHour   int
	minuteCounts []time.Time
	hourCounts   []time.Time
	mu           sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxPerMinute, maxPerHour int) *RateLimiter {
	return &RateLimiter{
		maxPerMinute: maxPerMinute,
		maxPerHour:   maxPerHour,
		minuteCounts: make([]time.Time, 0),
		hourCounts:   make([]time.Time, 0),
	}
}

// Allow checks if a bid is allowed and records it if so
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	minuteAgo := now.Add(-time.Minute)
	hourAgo := now.Add(-time.Hour)

	// Clean old entries
	newMinuteCounts := make([]time.Time, 0)
	for _, t := range r.minuteCounts {
		if t.After(minuteAgo) {
			newMinuteCounts = append(newMinuteCounts, t)
		}
	}
	r.minuteCounts = newMinuteCounts

	newHourCounts := make([]time.Time, 0)
	for _, t := range r.hourCounts {
		if t.After(hourAgo) {
			newHourCounts = append(newHourCounts, t)
		}
	}
	r.hourCounts = newHourCounts

	// Check limits
	if len(r.minuteCounts) >= r.maxPerMinute {
		return false
	}
	if len(r.hourCounts) >= r.maxPerHour {
		return false
	}

	// Record this attempt
	r.minuteCounts = append(r.minuteCounts, now)
	r.hourCounts = append(r.hourCounts, now)

	return true
}

// Reset resets the rate limiter
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.minuteCounts = make([]time.Time, 0)
	r.hourCounts = make([]time.Time, 0)
}

// BidEngine manages automatic bidding on orders
type BidEngine struct {
	config      BidEngineConfig
	provConfig  *ProviderConfig
	keyManager  *KeyManager
	chainClient ChainClient
	rateLimiter *RateLimiter

	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.RWMutex
	configMu sync.RWMutex

	// Channels for coordination
	orderChan  chan Order
	bidResults chan BidResult
}

// BidResult represents the result of a bid attempt
type BidResult struct {
	OrderID string
	BidID   string
	Success bool
	Error   error
}

// NewBidEngine creates a new bid engine
func NewBidEngine(
	config BidEngineConfig,
	keyManager *KeyManager,
	chainClient ChainClient,
) *BidEngine {
	return &BidEngine{
		config:      config,
		keyManager:  keyManager,
		chainClient: chainClient,
		rateLimiter: NewRateLimiter(config.MaxBidsPerMinute, config.MaxBidsPerHour),
		orderChan:   make(chan Order, 100),
		bidResults:  make(chan BidResult, 100),
	}
}

// Start starts the bid engine
func (be *BidEngine) Start(ctx context.Context) error {
	be.mu.Lock()
	if be.running {
		be.mu.Unlock()
		return nil
	}

	be.ctx, be.cancel = context.WithCancel(ctx)
	be.running = true
	be.mu.Unlock()

	// Load initial configuration
	if err := be.refreshConfig(); err != nil {
		return fmt.Errorf("failed to load provider config: %w", err)
	}

	// Start workers
	be.wg.Add(3)
	go be.configWatcher()
	go be.orderWatcher()
	go be.bidWorker()

	return nil
}

// Stop stops the bid engine
func (be *BidEngine) Stop() {
	be.mu.Lock()
	if !be.running {
		be.mu.Unlock()
		return
	}
	be.running = false
	be.mu.Unlock()

	if be.cancel != nil {
		be.cancel()
	}
	be.wg.Wait()
}

// IsRunning returns true if the bid engine is running
func (be *BidEngine) IsRunning() bool {
	be.mu.RLock()
	defer be.mu.RUnlock()
	return be.running
}

// GetConfig returns the current provider configuration
func (be *BidEngine) GetConfig() *ProviderConfig {
	be.configMu.RLock()
	defer be.configMu.RUnlock()
	return be.provConfig
}

// refreshConfig refreshes the provider configuration from chain
func (be *BidEngine) refreshConfig() error {
	config, err := be.chainClient.GetProviderConfig(be.ctx, be.config.ProviderAddress)
	if err != nil {
		return err
	}

	be.configMu.Lock()
	oldVersion := uint64(0)
	if be.provConfig != nil {
		oldVersion = be.provConfig.Version
	}

	if config.Version != oldVersion {
		be.provConfig = config
	}
	be.configMu.Unlock()

	return nil
}

// configWatcher watches for config updates
func (be *BidEngine) configWatcher() {
	defer be.wg.Done()

	ticker := time.NewTicker(be.config.ConfigPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-be.ctx.Done():
			return
		case <-ticker.C:
			if err := be.refreshConfig(); err != nil {
				// Log error but continue - config will be retried
				continue
			}
		}
	}
}

// orderWatcher watches for new orders
func (be *BidEngine) orderWatcher() {
	defer be.wg.Done()

	ticker := time.NewTicker(be.config.OrderPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-be.ctx.Done():
			return
		case <-ticker.C:
			if err := be.pollOrders(); err != nil {
				// Log error but continue
				continue
			}
		}
	}
}

// pollOrders polls for new orders and queues matchable ones
func (be *BidEngine) pollOrders() error {
	be.configMu.RLock()
	config := be.provConfig
	be.configMu.RUnlock()

	if config == nil || !config.Active {
		return nil
	}

	orders, err := be.chainClient.GetOpenOrders(be.ctx, config.SupportedOfferings, config.Regions)
	if err != nil {
		return err
	}

	for _, order := range orders {
		if be.matchOrder(order, config) {
			select {
			case be.orderChan <- order:
			default:
				// Channel full, skip this order
			}
		}
	}

	return nil
}

// matchOrder checks if an order matches provider capabilities
func (be *BidEngine) matchOrder(order Order, config *ProviderConfig) bool {
	// Check offering type
	supported := false
	for _, t := range config.SupportedOfferings {
		if t == order.OfferingType {
			supported = true
			break
		}
	}
	if !supported {
		return false
	}

	// Check region if specified
	if order.Region != "" {
		regionMatch := false
		for _, r := range config.Regions {
			if r == order.Region {
				regionMatch = true
				break
			}
		}
		if !regionMatch {
			return false
		}
	}

	// Check capacity
	capacity := config.Capacity
	if order.Requirements.CPUCores > capacity.AvailableCPU() {
		return false
	}
	if order.Requirements.MemoryGB > capacity.AvailableMemory() {
		return false
	}
	if order.Requirements.StorageGB > capacity.AvailableStorage() {
		return false
	}
	if order.Requirements.GPUs > capacity.TotalGPUs {
		return false
	}

	return true
}

// bidWorker processes orders and places bids
func (be *BidEngine) bidWorker() {
	defer be.wg.Done()

	for {
		select {
		case <-be.ctx.Done():
			return
		case order := <-be.orderChan:
			result := be.processBid(order)
			select {
			case be.bidResults <- result:
			default:
				// Results channel full, discard
			}
		}
	}
}

// processBid processes a single order and places a bid
func (be *BidEngine) processBid(order Order) BidResult {
	result := BidResult{
		OrderID: order.OrderID,
	}

	// Check rate limit
	if !be.rateLimiter.Allow() {
		result.Error = ErrBidRateLimited
		return result
	}

	// Get current config
	be.configMu.RLock()
	config := be.provConfig
	be.configMu.RUnlock()

	if config == nil {
		result.Error = errors.New("provider config not loaded")
		return result
	}

	// Calculate bid price
	price, err := be.calculateBidPrice(order, config)
	if err != nil {
		result.Error = err
		return result
	}

	// Create bid
	bid := &Bid{
		BidID:           generateBidID(order.OrderID, be.config.ProviderAddress),
		OrderID:         order.OrderID,
		ProviderAddress: be.config.ProviderAddress,
		Price:           price,
		Currency:        config.Pricing.Currency,
		CreatedAt:       time.Now().UTC(),
		State:           "open",
	}

	// Sign the bid
	bidData := fmt.Sprintf("%s:%s:%s:%d", bid.BidID, bid.OrderID, bid.Price, bid.CreatedAt.Unix())
	sig, err := be.keyManager.Sign([]byte(bidData))
	if err != nil {
		result.Error = fmt.Errorf("failed to sign bid: %w", err)
		return result
	}

	// Place the bid
	if err := be.chainClient.PlaceBid(be.ctx, bid, sig); err != nil {
		result.Error = fmt.Errorf("failed to place bid: %w", err)
		return result
	}

	result.BidID = bid.BidID
	result.Success = true
	return result
}

// calculateBidPrice calculates the bid price for an order
func (be *BidEngine) calculateBidPrice(order Order, config *ProviderConfig) (string, error) {
	if config == nil {
		return "", ErrInvalidPrice
	}

	if config.Pricing.MinBidPrice == "" {
		return "", ErrInvalidPrice
	}

	cpuPrice, err := parsePriceDecRequired("cpu_price_per_core", config.Pricing.CPUPricePerCore, order.Requirements.CPUCores > 0)
	if err != nil {
		return "", err
	}
	memoryPrice, err := parsePriceDecRequired("memory_price_per_gb", config.Pricing.MemoryPricePerGB, order.Requirements.MemoryGB > 0)
	if err != nil {
		return "", err
	}
	storagePrice, err := parsePriceDecRequired("storage_price_per_gb", config.Pricing.StoragePricePerGB, order.Requirements.StorageGB > 0)
	if err != nil {
		return "", err
	}
	minBidPrice, err := parsePriceDec("min_bid_price", config.Pricing.MinBidPrice)
	if err != nil {
		return "", err
	}

	total := sdkmath.LegacyZeroDec()
	total = total.Add(cpuPrice.Mul(sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(order.Requirements.CPUCores))))
	total = total.Add(memoryPrice.Mul(sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(order.Requirements.MemoryGB))))
	total = total.Add(storagePrice.Mul(sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(order.Requirements.StorageGB))))

	if order.Requirements.GPUs > 0 {
		if config.Pricing.GPUPricePerHour == "" {
			return "", fmt.Errorf("%w: gpu_price_per_hour missing", ErrInvalidPrice)
		}
		gpuPrice, err := parsePriceDec("gpu_price_per_hour", config.Pricing.GPUPricePerHour)
		if err != nil {
			return "", err
		}
		total = total.Add(gpuPrice.Mul(sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(order.Requirements.GPUs))))
	}

	if config.Pricing.BidMarkupPercent < 0 {
		return "", fmt.Errorf("%w: bid_markup_percent cannot be negative", ErrInvalidPrice)
	}
	if config.Pricing.BidMarkupPercent > 0 {
		percentStr := strconv.FormatFloat(config.Pricing.BidMarkupPercent, 'f', -1, 64)
		percentDec, err := sdkmath.LegacyNewDecFromStr(percentStr)
		if err != nil {
			return "", fmt.Errorf("%w: invalid bid_markup_percent", ErrInvalidPrice)
		}
		markup := percentDec.Quo(sdkmath.LegacyNewDec(100))
		total = total.Mul(sdkmath.LegacyOneDec().Add(markup))
	}

	if total.LT(minBidPrice) {
		total = minBidPrice
	}

	if order.MaxPrice != "" {
		maxPrice, err := parsePriceDec("max_price", order.MaxPrice)
		if err != nil {
			return "", err
		}
		if total.GT(maxPrice) {
			return "", ErrInvalidPrice
		}
	}

	if total.IsNegative() || total.IsZero() {
		return "", ErrInvalidPrice
	}

	return total.String(), nil
}

func parsePriceDec(field, value string) (sdkmath.LegacyDec, error) {
	if value == "" {
		return sdkmath.LegacyDec{}, fmt.Errorf("%w: %s is empty", ErrInvalidPrice, field)
	}
	dec, err := sdkmath.LegacyNewDecFromStr(value)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("%w: %s invalid", ErrInvalidPrice, field)
	}
	if dec.IsNegative() {
		return sdkmath.LegacyDec{}, fmt.Errorf("%w: %s negative", ErrInvalidPrice, field)
	}
	return dec, nil
}

func parsePriceDecRequired(field, value string, required bool) (sdkmath.LegacyDec, error) {
	if !required && value == "" {
		return sdkmath.LegacyZeroDec(), nil
	}
	return parsePriceDec(field, value)
}

// generateBidID generates a unique bid ID
func generateBidID(orderID, providerAddress string) string {
	return fmt.Sprintf("bid-%s-%s-%d", orderID, providerAddress[:8], time.Now().UnixNano())
}

// GetBidResults returns the bid results channel for monitoring
func (be *BidEngine) GetBidResults() <-chan BidResult {
	return be.bidResults
}

// ManualBid allows manually placing a bid on a specific order
func (be *BidEngine) ManualBid(ctx context.Context, order Order) (*Bid, error) {
	if !be.IsRunning() {
		return nil, ErrBidEngineNotRunning
	}

	be.configMu.RLock()
	config := be.provConfig
	be.configMu.RUnlock()

	if config == nil {
		return nil, errors.New("provider config not loaded")
	}

	// Verify order matches
	if !be.matchOrder(order, config) {
		return nil, ErrOrderNotMatchable
	}

	// Check rate limit
	if !be.rateLimiter.Allow() {
		return nil, ErrBidRateLimited
	}

	// Calculate price
	price, err := be.calculateBidPrice(order, config)
	if err != nil {
		return nil, err
	}

	// Create and sign bid
	bid := &Bid{
		BidID:           generateBidID(order.OrderID, be.config.ProviderAddress),
		OrderID:         order.OrderID,
		ProviderAddress: be.config.ProviderAddress,
		Price:           price,
		Currency:        config.Pricing.Currency,
		CreatedAt:       time.Now().UTC(),
		State:           "open",
	}

	bidData := fmt.Sprintf("%s:%s:%s:%d", bid.BidID, bid.OrderID, bid.Price, bid.CreatedAt.Unix())
	sig, err := be.keyManager.Sign([]byte(bidData))
	if err != nil {
		return nil, fmt.Errorf("failed to sign bid: %w", err)
	}

	if err := be.chainClient.PlaceBid(ctx, bid, sig); err != nil {
		return nil, fmt.Errorf("failed to place bid: %w", err)
	}

	return bid, nil
}
