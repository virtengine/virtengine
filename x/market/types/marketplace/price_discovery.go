// Package marketplace provides types for the marketplace on-chain module.
//
// ECON-002: Marketplace Economics Optimization
// This file defines price discovery mechanisms and price oracle functionality.
package marketplace

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// PriceAggregationType defines how prices are aggregated
type PriceAggregationType uint8

const (
	// PriceAggregationTypeNone indicates no aggregation
	PriceAggregationTypeNone PriceAggregationType = 0

	// PriceAggregationTypeTWAP is Time-Weighted Average Price
	PriceAggregationTypeTWAP PriceAggregationType = 1

	// PriceAggregationTypeVWAP is Volume-Weighted Average Price
	PriceAggregationTypeVWAP PriceAggregationType = 2

	// PriceAggregationTypeEMA is Exponential Moving Average
	PriceAggregationTypeEMA PriceAggregationType = 3

	// PriceAggregationTypeMedian is median price
	PriceAggregationTypeMedian PriceAggregationType = 4

	// PriceAggregationTypeLastTrade is the last trade price
	PriceAggregationTypeLastTrade PriceAggregationType = 5
)

// PriceAggregationTypeNames maps types to human-readable names
var PriceAggregationTypeNames = map[PriceAggregationType]string{
	PriceAggregationTypeNone:      "none",
	PriceAggregationTypeTWAP:      "twap",
	PriceAggregationTypeVWAP:      "vwap",
	PriceAggregationTypeEMA:       "ema",
	PriceAggregationTypeMedian:    "median",
	PriceAggregationTypeLastTrade: "last_trade",
}

// String returns the string representation of a PriceAggregationType
func (t PriceAggregationType) String() string {
	if name, ok := PriceAggregationTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// PriceValidationStatus represents the validation status of a price
type PriceValidationStatus uint8

const (
	// PriceValidationStatusUnknown indicates unknown validation status
	PriceValidationStatusUnknown PriceValidationStatus = 0

	// PriceValidationStatusValid indicates the price is valid
	PriceValidationStatusValid PriceValidationStatus = 1

	// PriceValidationStatusOutOfBand indicates price is outside acceptable band
	PriceValidationStatusOutOfBand PriceValidationStatus = 2

	// PriceValidationStatusStale indicates the price is stale
	PriceValidationStatusStale PriceValidationStatus = 3

	// PriceValidationStatusSuspicious indicates suspicious price movement
	PriceValidationStatusSuspicious PriceValidationStatus = 4
)

// PriceValidationStatusNames maps statuses to human-readable names
var PriceValidationStatusNames = map[PriceValidationStatus]string{
	PriceValidationStatusUnknown:    "unknown",
	PriceValidationStatusValid:      "valid",
	PriceValidationStatusOutOfBand:  "out_of_band",
	PriceValidationStatusStale:      "stale",
	PriceValidationStatusSuspicious: "suspicious",
}

// String returns the string representation of a PriceValidationStatus
func (s PriceValidationStatus) String() string {
	if name, ok := PriceValidationStatusNames[s]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// PricePoint represents a single price observation
type PricePoint struct {
	// Price is the observed price
	Price uint64 `json:"price"`

	// Volume is the volume at this price
	Volume uint64 `json:"volume"`

	// Timestamp is when the price was observed
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is the block height when observed
	BlockHeight int64 `json:"block_height"`

	// Source identifies the price source
	Source string `json:"source"`
}

// PriceBand defines acceptable price boundaries
type PriceBand struct {
	// ReferencePrice is the reference price for the band
	ReferencePrice uint64 `json:"reference_price"`

	// UpperBoundBps is the upper bound deviation in basis points
	UpperBoundBps uint32 `json:"upper_bound_bps"`

	// LowerBoundBps is the lower bound deviation in basis points
	LowerBoundBps uint32 `json:"lower_bound_bps"`

	// ValidUntil is when the band expires
	ValidUntil time.Time `json:"valid_until"`

	// LastUpdated is when the band was last updated
	LastUpdated time.Time `json:"last_updated"`
}

// NewPriceBand creates a new price band
func NewPriceBand(refPrice uint64, upperBps, lowerBps uint32, validity time.Duration, now time.Time) *PriceBand {
	return &PriceBand{
		ReferencePrice: refPrice,
		UpperBoundBps:  upperBps,
		LowerBoundBps:  lowerBps,
		ValidUntil:     now.Add(validity),
		LastUpdated:    now,
	}
}

// UpperBound returns the upper price bound
func (b *PriceBand) UpperBound() uint64 {
	return b.ReferencePrice + (b.ReferencePrice*uint64(b.UpperBoundBps))/10000
}

// LowerBound returns the lower price bound
func (b *PriceBand) LowerBound() uint64 {
	lowerDelta := (b.ReferencePrice * uint64(b.LowerBoundBps)) / 10000
	if lowerDelta >= b.ReferencePrice {
		return 0
	}
	return b.ReferencePrice - lowerDelta
}

// IsWithinBand checks if a price is within the band
func (b *PriceBand) IsWithinBand(price uint64) bool {
	return price >= b.LowerBound() && price <= b.UpperBound()
}

// IsValid checks if the band is still valid
func (b *PriceBand) IsValid(now time.Time) bool {
	return now.Before(b.ValidUntil)
}

// PriceOracleConfig defines price oracle configuration
type PriceOracleConfig struct {
	// Enabled indicates if the price oracle is enabled
	Enabled bool `json:"enabled"`

	// AggregationType is the default aggregation type
	AggregationType PriceAggregationType `json:"aggregation_type"`

	// TWAPWindowBlocks is the TWAP window in blocks
	TWAPWindowBlocks int64 `json:"twap_window_blocks"`

	// VWAPWindowBlocks is the VWAP window in blocks
	VWAPWindowBlocks int64 `json:"vwap_window_blocks"`

	// EMAAlpha is the EMA smoothing factor (1-100, higher = more recent)
	EMAAlpha uint32 `json:"ema_alpha"`

	// PriceBandUpperBps is the default upper price band
	PriceBandUpperBps uint32 `json:"price_band_upper_bps"`

	// PriceBandLowerBps is the default lower price band
	PriceBandLowerBps uint32 `json:"price_band_lower_bps"`

	// MaxPriceAgeDuration is the maximum age for valid prices
	MaxPriceAgeDuration time.Duration `json:"max_price_age_duration"`

	// MinObservationsForTWAP is minimum observations for TWAP calculation
	MinObservationsForTWAP uint32 `json:"min_observations_for_twap"`

	// MinVolumeForVWAP is minimum volume for VWAP calculation
	MinVolumeForVWAP uint64 `json:"min_volume_for_vwap"`

	// SuspiciousMovementThresholdBps is the threshold for suspicious movement
	SuspiciousMovementThresholdBps uint32 `json:"suspicious_movement_threshold_bps"`

	// PriceUpdateCooldownBlocks is cooldown between price updates
	PriceUpdateCooldownBlocks int64 `json:"price_update_cooldown_blocks"`
}

// DefaultPriceOracleConfig returns the default price oracle configuration
func DefaultPriceOracleConfig() PriceOracleConfig {
	return PriceOracleConfig{
		Enabled:                        true,
		AggregationType:                PriceAggregationTypeTWAP,
		TWAPWindowBlocks:               100, // ~10 minutes at 6s blocks
		VWAPWindowBlocks:               100,
		EMAAlpha:                       20,   // 20% weight to most recent
		PriceBandUpperBps:              1000, // 10% upper band
		PriceBandLowerBps:              1000, // 10% lower band
		MaxPriceAgeDuration:            time.Hour,
		MinObservationsForTWAP:         3,
		MinVolumeForVWAP:               1000000, // 1 token
		SuspiciousMovementThresholdBps: 500,     // 5% sudden movement
		PriceUpdateCooldownBlocks:      1,
	}
}

// Validate validates the price oracle configuration
func (c *PriceOracleConfig) Validate() error {
	if c.TWAPWindowBlocks <= 0 {
		return fmt.Errorf("twap_window_blocks must be positive")
	}
	if c.VWAPWindowBlocks <= 0 {
		return fmt.Errorf("vwap_window_blocks must be positive")
	}
	if c.EMAAlpha > 100 {
		return fmt.Errorf("ema_alpha cannot exceed 100")
	}
	if c.PriceBandUpperBps > 10000 {
		return fmt.Errorf("price_band_upper_bps cannot exceed 10000")
	}
	if c.PriceBandLowerBps > 10000 {
		return fmt.Errorf("price_band_lower_bps cannot exceed 10000")
	}
	return nil
}

// PriceHistory maintains price history for an offering
type PriceHistory struct {
	// OfferingID is the offering this history is for
	OfferingID OfferingID `json:"offering_id"`

	// Points are the price observation points
	Points []PricePoint `json:"points"`

	// MaxPoints is the maximum points to retain
	MaxPoints int `json:"max_points"`

	// CurrentTWAP is the current TWAP
	CurrentTWAP uint64 `json:"current_twap"`

	// CurrentVWAP is the current VWAP
	CurrentVWAP uint64 `json:"current_vwap"`

	// CurrentEMA is the current EMA
	CurrentEMA uint64 `json:"current_ema"`

	// LastPrice is the last observed price
	LastPrice uint64 `json:"last_price"`

	// LastVolume is the last observed volume
	LastVolume uint64 `json:"last_volume"`

	// TotalVolume is total volume in the window
	TotalVolume uint64 `json:"total_volume"`

	// HighPrice is the highest price in the window
	HighPrice uint64 `json:"high_price"`

	// LowPrice is the lowest price in the window
	LowPrice uint64 `json:"low_price"`

	// PriceBand is the current price band
	PriceBand *PriceBand `json:"price_band,omitempty"`

	// LastUpdated is when the history was last updated
	LastUpdated time.Time `json:"last_updated"`

	// LastUpdateBlock is the last update block
	LastUpdateBlock int64 `json:"last_update_block"`
}

// NewPriceHistory creates a new price history
func NewPriceHistory(offeringID OfferingID, maxPoints int) *PriceHistory {
	return &PriceHistory{
		OfferingID: offeringID,
		Points:     make([]PricePoint, 0, maxPoints),
		MaxPoints:  maxPoints,
	}
}

// AddPoint adds a new price point
func (h *PriceHistory) AddPoint(point PricePoint) {
	h.Points = append(h.Points, point)

	// Trim to max points
	if len(h.Points) > h.MaxPoints {
		h.Points = h.Points[len(h.Points)-h.MaxPoints:]
	}

	h.LastPrice = point.Price
	h.LastVolume = point.Volume
	h.LastUpdated = point.Timestamp
	h.LastUpdateBlock = point.BlockHeight

	// Update high/low
	if point.Price > h.HighPrice || h.HighPrice == 0 {
		h.HighPrice = point.Price
	}
	if point.Price < h.LowPrice || h.LowPrice == 0 {
		h.LowPrice = point.Price
	}
}

// GetPointsInWindow returns points within a block window
func (h *PriceHistory) GetPointsInWindow(currentBlock int64, windowBlocks int64) []PricePoint {
	startBlock := currentBlock - windowBlocks
	result := make([]PricePoint, 0)
	for _, p := range h.Points {
		if p.BlockHeight >= startBlock {
			result = append(result, p)
		}
	}
	return result
}

// PriceOracle calculates and validates prices
type PriceOracle struct {
	Config PriceOracleConfig `json:"config"`
}

// NewPriceOracle creates a new price oracle
func NewPriceOracle(config PriceOracleConfig) *PriceOracle {
	return &PriceOracle{Config: config}
}

// CalculateTWAP calculates Time-Weighted Average Price
func (o *PriceOracle) CalculateTWAP(history *PriceHistory, currentBlock int64) (uint64, error) {
	points := history.GetPointsInWindow(currentBlock, o.Config.TWAPWindowBlocks)

	if len(points) < int(o.Config.MinObservationsForTWAP) {
		return 0, fmt.Errorf("insufficient observations for TWAP: %d < %d", len(points), o.Config.MinObservationsForTWAP)
	}

	// Sort by block height
	sort.Slice(points, func(i, j int) bool {
		return points[i].BlockHeight < points[j].BlockHeight
	})

	var weightedSum, totalWeight uint64

	for i := 0; i < len(points)-1; i++ {
		timeDelta := uint64(points[i+1].BlockHeight - points[i].BlockHeight)
		weightedSum += points[i].Price * timeDelta
		totalWeight += timeDelta
	}

	// Add the last point with weight 1
	if len(points) > 0 {
		weightedSum += points[len(points)-1].Price
		totalWeight++
	}

	if totalWeight == 0 {
		return 0, fmt.Errorf("zero total weight in TWAP calculation")
	}

	return weightedSum / totalWeight, nil
}

// CalculateVWAP calculates Volume-Weighted Average Price
func (o *PriceOracle) CalculateVWAP(history *PriceHistory, currentBlock int64) (uint64, error) {
	points := history.GetPointsInWindow(currentBlock, o.Config.VWAPWindowBlocks)

	var totalVolumeWeightedPrice, totalVolume uint64
	for _, p := range points {
		totalVolumeWeightedPrice += p.Price * p.Volume
		totalVolume += p.Volume
	}

	if totalVolume < o.Config.MinVolumeForVWAP {
		return 0, fmt.Errorf("insufficient volume for VWAP: %d < %d", totalVolume, o.Config.MinVolumeForVWAP)
	}

	return totalVolumeWeightedPrice / totalVolume, nil
}

// CalculateEMA calculates Exponential Moving Average
func (o *PriceOracle) CalculateEMA(currentEMA, newPrice uint64) uint64 {
	if currentEMA == 0 {
		return newPrice
	}

	alpha := uint64(o.Config.EMAAlpha)
	// EMA = alpha * newPrice + (1 - alpha) * currentEMA
	// Using fixed point: alpha is 0-100, so divide by 100
	weightedNew := (newPrice * alpha) / 100
	weightedOld := (currentEMA * (100 - alpha)) / 100
	return weightedNew + weightedOld
}

// CalculateMedian calculates the median price
func (o *PriceOracle) CalculateMedian(history *PriceHistory, currentBlock int64) (uint64, error) {
	points := history.GetPointsInWindow(currentBlock, o.Config.TWAPWindowBlocks)

	if len(points) == 0 {
		return 0, fmt.Errorf("no price points for median calculation")
	}

	prices := make([]uint64, len(points))
	for i, p := range points {
		prices[i] = p.Price
	}

	sort.Slice(prices, func(i, j int) bool {
		return prices[i] < prices[j]
	})

	mid := len(prices) / 2
	if len(prices)%2 == 0 {
		return (prices[mid-1] + prices[mid]) / 2, nil
	}
	return prices[mid], nil
}

// GetAggregatedPrice returns the aggregated price using the configured method
func (o *PriceOracle) GetAggregatedPrice(history *PriceHistory, currentBlock int64) (uint64, PriceAggregationType, error) {
	if !o.Config.Enabled {
		return history.LastPrice, PriceAggregationTypeLastTrade, nil
	}

	switch o.Config.AggregationType {
	case PriceAggregationTypeTWAP:
		price, err := o.CalculateTWAP(history, currentBlock)
		if err != nil {
			// Fallback to last trade
			return history.LastPrice, PriceAggregationTypeLastTrade, nil
		}
		return price, PriceAggregationTypeTWAP, nil

	case PriceAggregationTypeVWAP:
		price, err := o.CalculateVWAP(history, currentBlock)
		if err != nil {
			// Fallback to TWAP
			price, err = o.CalculateTWAP(history, currentBlock)
			if err != nil {
				return history.LastPrice, PriceAggregationTypeLastTrade, nil
			}
			return price, PriceAggregationTypeTWAP, nil
		}
		return price, PriceAggregationTypeVWAP, nil

	case PriceAggregationTypeEMA:
		return history.CurrentEMA, PriceAggregationTypeEMA, nil

	case PriceAggregationTypeMedian:
		price, err := o.CalculateMedian(history, currentBlock)
		if err != nil {
			return history.LastPrice, PriceAggregationTypeLastTrade, nil
		}
		return price, PriceAggregationTypeMedian, nil

	default:
		return history.LastPrice, PriceAggregationTypeLastTrade, nil
	}
}

// PriceValidationResult contains the result of price validation
type PriceValidationResult struct {
	// Status is the validation status
	Status PriceValidationStatus `json:"status"`

	// ValidatedPrice is the validated/adjusted price
	ValidatedPrice uint64 `json:"validated_price"`

	// OriginalPrice is the original submitted price
	OriginalPrice uint64 `json:"original_price"`

	// ReferencePrice is the reference price used for validation
	ReferencePrice uint64 `json:"reference_price"`

	// DeviationBps is the deviation from reference in basis points
	DeviationBps int32 `json:"deviation_bps"`

	// Reason describes the validation result
	Reason string `json:"reason"`

	// ValidatedAt is when validation occurred
	ValidatedAt time.Time `json:"validated_at"`
}

// ValidatePrice validates a price against the current price band
func (o *PriceOracle) ValidatePrice(price uint64, history *PriceHistory, now time.Time) *PriceValidationResult {
	result := &PriceValidationResult{
		OriginalPrice: price,
		ValidatedAt:   now,
	}

	// Check if we have a valid price band
	if history.PriceBand == nil || !history.PriceBand.IsValid(now) {
		// No valid band - accept the price but mark as unknown
		result.Status = PriceValidationStatusValid
		result.ValidatedPrice = price
		result.Reason = "no_price_band"
		return result
	}

	band := history.PriceBand
	result.ReferencePrice = band.ReferencePrice

	// Calculate deviation
	var deviation int64
	if price > band.ReferencePrice {
		deviation = int64(price - band.ReferencePrice)
	} else {
		deviation = -int64(band.ReferencePrice - price)
	}
	result.DeviationBps = int32((deviation * 10000) / int64(band.ReferencePrice))

	// Check if within band
	if band.IsWithinBand(price) {
		result.Status = PriceValidationStatusValid
		result.ValidatedPrice = price
		result.Reason = "within_band"
		return result
	}

	// Price is outside band
	result.Status = PriceValidationStatusOutOfBand

	// Check for suspicious movement
	if abs(int64(result.DeviationBps)) > int64(o.Config.SuspiciousMovementThresholdBps) {
		result.Status = PriceValidationStatusSuspicious
		result.Reason = fmt.Sprintf("suspicious_movement_%dbps", result.DeviationBps)
	} else {
		result.Reason = fmt.Sprintf("out_of_band_%dbps", result.DeviationBps)
	}

	// Clamp to band boundaries
	if price > band.UpperBound() {
		result.ValidatedPrice = band.UpperBound()
	} else {
		result.ValidatedPrice = band.LowerBound()
	}

	return result
}

// UpdatePriceBand updates the price band based on current prices
func (o *PriceOracle) UpdatePriceBand(history *PriceHistory, currentBlock int64, now time.Time) error {
	// Get aggregated price as reference
	refPrice, _, err := o.GetAggregatedPrice(history, currentBlock)
	if err != nil {
		return fmt.Errorf("failed to get aggregated price: %w", err)
	}

	if refPrice == 0 {
		return fmt.Errorf("reference price is zero")
	}

	history.PriceBand = NewPriceBand(
		refPrice,
		o.Config.PriceBandUpperBps,
		o.Config.PriceBandLowerBps,
		o.Config.MaxPriceAgeDuration,
		now,
	)

	return nil
}

// PriceDiscoveryParams holds price discovery parameters
type PriceDiscoveryParams struct {
	// OracleConfig is the price oracle configuration
	OracleConfig PriceOracleConfig `json:"oracle_config"`

	// MaxHistoryPointsPerOffering is max price points per offering
	MaxHistoryPointsPerOffering int `json:"max_history_points_per_offering"`

	// PriceBandUpdateIntervalBlocks is how often to update bands
	PriceBandUpdateIntervalBlocks int64 `json:"price_band_update_interval_blocks"`

	// EnableCircuitBreaker enables circuit breaker on extreme moves
	EnableCircuitBreaker bool `json:"enable_circuit_breaker"`

	// CircuitBreakerThresholdBps is the threshold for circuit breaker
	CircuitBreakerThresholdBps uint32 `json:"circuit_breaker_threshold_bps"`

	// CircuitBreakerCooldownBlocks is cooldown after circuit breaker trips
	CircuitBreakerCooldownBlocks int64 `json:"circuit_breaker_cooldown_blocks"`
}

// DefaultPriceDiscoveryParams returns default price discovery parameters
func DefaultPriceDiscoveryParams() PriceDiscoveryParams {
	return PriceDiscoveryParams{
		OracleConfig:                  DefaultPriceOracleConfig(),
		MaxHistoryPointsPerOffering:   1000,
		PriceBandUpdateIntervalBlocks: 10,
		EnableCircuitBreaker:          true,
		CircuitBreakerThresholdBps:    2000, // 20% move
		CircuitBreakerCooldownBlocks:  100,  // ~10 minutes
	}
}

// Validate validates the price discovery parameters
func (p *PriceDiscoveryParams) Validate() error {
	if err := p.OracleConfig.Validate(); err != nil {
		return fmt.Errorf("invalid oracle config: %w", err)
	}
	if p.MaxHistoryPointsPerOffering <= 0 {
		return fmt.Errorf("max_history_points_per_offering must be positive")
	}
	if p.PriceBandUpdateIntervalBlocks <= 0 {
		return fmt.Errorf("price_band_update_interval_blocks must be positive")
	}
	if p.CircuitBreakerThresholdBps > 10000 {
		return fmt.Errorf("circuit_breaker_threshold_bps cannot exceed 10000")
	}
	return nil
}

// PriceStatistics holds price statistics for an offering
type PriceStatistics struct {
	// OfferingID is the offering ID
	OfferingID OfferingID `json:"offering_id"`

	// TWAP is the current TWAP
	TWAP uint64 `json:"twap"`

	// VWAP is the current VWAP
	VWAP uint64 `json:"vwap"`

	// EMA is the current EMA
	EMA uint64 `json:"ema"`

	// MedianPrice is the current median
	MedianPrice uint64 `json:"median_price"`

	// LastTradePrice is the last trade price
	LastTradePrice uint64 `json:"last_trade_price"`

	// High24h is the 24-hour high
	High24h uint64 `json:"high_24h"`

	// Low24h is the 24-hour low
	Low24h uint64 `json:"low_24h"`

	// Volume24h is the 24-hour volume
	Volume24h uint64 `json:"volume_24h"`

	// PriceChange24hBps is 24-hour price change in basis points
	PriceChange24hBps int32 `json:"price_change_24h_bps"`

	// Volatility is the price volatility (standard deviation as bps)
	Volatility uint32 `json:"volatility_bps"`

	// LastUpdated is when stats were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// CalculateVolatility calculates price volatility as standard deviation in basis points
func CalculateVolatility(prices []uint64) uint32 {
	if len(prices) < 2 {
		return 0
	}

	// Calculate mean
	var sum uint64
	for _, p := range prices {
		sum += p
	}
	mean := float64(sum) / float64(len(prices))

	// Calculate variance
	var variance float64
	for _, p := range prices {
		diff := float64(p) - mean
		variance += diff * diff
	}
	variance /= float64(len(prices))

	// Standard deviation as percentage of mean
	stdDev := math.Sqrt(variance)
	if mean == 0 {
		return 0
	}

	return uint32((stdDev / mean) * 10000)
}

// abs returns absolute value of an int64
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
