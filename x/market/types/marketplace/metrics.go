// Package marketplace provides types for the marketplace on-chain module.
//
// ECON-002: Marketplace Economics Optimization
// This file defines market efficiency metrics and analytics.
package marketplace

import (
	"fmt"
	"time"
)

// MetricType represents the type of market metric
type MetricType uint8

const (
	// MetricTypeNone indicates no metric type
	MetricTypeNone MetricType = 0

	// MetricTypeSpread tracks bid-ask spread
	MetricTypeSpread MetricType = 1

	// MetricTypeFillRate tracks order fill rate
	MetricTypeFillRate MetricType = 2

	// MetricTypeVolatility tracks price volatility
	MetricTypeVolatility MetricType = 3

	// MetricTypeDepth tracks market depth
	MetricTypeDepth MetricType = 4

	// MetricTypeLiquidity tracks liquidity ratio
	MetricTypeLiquidity MetricType = 5

	// MetricTypeVolume tracks trading volume
	MetricTypeVolume MetricType = 6

	// MetricTypeLatency tracks order latency
	MetricTypeLatency MetricType = 7

	// MetricTypeEfficiency tracks overall market efficiency
	MetricTypeEfficiency MetricType = 8
)

// trendStable is the constant for a stable trend indicator
const trendStable = "stable"

// MetricTypeNames maps metric types to human-readable names
var MetricTypeNames = map[MetricType]string{
	MetricTypeNone:       "none",
	MetricTypeSpread:     "spread",
	MetricTypeFillRate:   "fill_rate",
	MetricTypeVolatility: "volatility",
	MetricTypeDepth:      "depth",
	MetricTypeLiquidity:  "liquidity",
	MetricTypeVolume:     "volume",
	MetricTypeLatency:    "latency",
	MetricTypeEfficiency: "efficiency",
}

// String returns the string representation of a MetricType
func (t MetricType) String() string {
	if name, ok := MetricTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", t)
}

// MetricPeriod represents the time period for metrics
type MetricPeriod uint8

const (
	// MetricPeriodBlock is per-block metrics
	MetricPeriodBlock MetricPeriod = 0

	// MetricPeriodHour is hourly metrics
	MetricPeriodHour MetricPeriod = 1

	// MetricPeriodDay is daily metrics
	MetricPeriodDay MetricPeriod = 2

	// MetricPeriodWeek is weekly metrics
	MetricPeriodWeek MetricPeriod = 3

	// MetricPeriodMonth is monthly metrics
	MetricPeriodMonth MetricPeriod = 4
)

// MetricPeriodNames maps periods to human-readable names
var MetricPeriodNames = map[MetricPeriod]string{
	MetricPeriodBlock: "block",
	MetricPeriodHour:  "hour",
	MetricPeriodDay:   "day",
	MetricPeriodWeek:  "week",
	MetricPeriodMonth: "month",
}

// String returns the string representation of a MetricPeriod
func (p MetricPeriod) String() string {
	if name, ok := MetricPeriodNames[p]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", p)
}

// BlocksPerPeriod returns the approximate blocks per period (assuming 6s blocks)
func (p MetricPeriod) BlocksPerPeriod() int64 {
	switch p {
	case MetricPeriodBlock:
		return 1
	case MetricPeriodHour:
		return 600 // 3600 / 6
	case MetricPeriodDay:
		return 14400 // 86400 / 6
	case MetricPeriodWeek:
		return 100800 // 604800 / 6
	case MetricPeriodMonth:
		return 432000 // 2592000 / 6
	default:
		return 1
	}
}

// SpreadMetrics tracks bid-ask spread metrics
type SpreadMetrics struct {
	// CurrentSpreadBps is the current spread in basis points
	CurrentSpreadBps uint32 `json:"current_spread_bps"`

	// AverageSpreadBps is the average spread
	AverageSpreadBps uint32 `json:"average_spread_bps"`

	// MinSpreadBps is the minimum spread observed
	MinSpreadBps uint32 `json:"min_spread_bps"`

	// MaxSpreadBps is the maximum spread observed
	MaxSpreadBps uint32 `json:"max_spread_bps"`

	// SpreadVolatilityBps is the spread volatility
	SpreadVolatilityBps uint32 `json:"spread_volatility_bps"`

	// TightSpreadPercentage is % of time spread was below threshold
	TightSpreadPercentage uint32 `json:"tight_spread_percentage"`

	// TightSpreadThresholdBps is the threshold for tight spread
	TightSpreadThresholdBps uint32 `json:"tight_spread_threshold_bps"`

	// ObservationCount is the number of observations
	ObservationCount uint64 `json:"observation_count"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// FillRateMetrics tracks order fill rate metrics
type FillRateMetrics struct {
	// TotalOrders is total orders in the period
	TotalOrders uint64 `json:"total_orders"`

	// FilledOrders is fully filled orders
	FilledOrders uint64 `json:"filled_orders"`

	// PartiallyFilledOrders is partially filled orders
	PartiallyFilledOrders uint64 `json:"partially_filled_orders"`

	// CancelledOrders is cancelled orders
	CancelledOrders uint64 `json:"cancelled_orders"`

	// ExpiredOrders is expired orders
	ExpiredOrders uint64 `json:"expired_orders"`

	// FillRatePercentage is the fill rate (filled / total)
	FillRatePercentage uint32 `json:"fill_rate_percentage"`

	// AverageFillPercentage is average fill % across all orders
	AverageFillPercentage uint32 `json:"average_fill_percentage"`

	// AverageTimeToFillBlocks is average blocks to fill
	AverageTimeToFillBlocks int64 `json:"average_time_to_fill_blocks"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// CalculateFillRate calculates the fill rate percentage
func (m *FillRateMetrics) CalculateFillRate() uint32 {
	if m.TotalOrders == 0 {
		return 0
	}
	return uint32((m.FilledOrders * 100) / m.TotalOrders)
}

// MarketDepthMetrics tracks market depth metrics
type MarketDepthMetrics struct {
	// TotalBidVolume is total volume on bid side
	TotalBidVolume uint64 `json:"total_bid_volume"`

	// TotalAskVolume is total volume on ask side (offerings)
	TotalAskVolume uint64 `json:"total_ask_volume"`

	// BidCount is number of bids
	BidCount uint64 `json:"bid_count"`

	// AskCount is number of asks
	AskCount uint64 `json:"ask_count"`

	// DepthRatioBps is bid/ask ratio in basis points (10000 = balanced)
	DepthRatioBps uint32 `json:"depth_ratio_bps"`

	// Depth5PctBps is depth within 5% of best price
	Depth5PctBps uint32 `json:"depth_5pct_bps"`

	// Depth10PctBps is depth within 10% of best price
	Depth10PctBps uint32 `json:"depth_10pct_bps"`

	// BestBidPrice is the best bid price
	BestBidPrice uint64 `json:"best_bid_price"`

	// BestAskPrice is the best ask price
	BestAskPrice uint64 `json:"best_ask_price"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// CalculateDepthRatio calculates the bid/ask depth ratio
func (m *MarketDepthMetrics) CalculateDepthRatio() uint32 {
	if m.TotalAskVolume == 0 {
		if m.TotalBidVolume == 0 {
			return 10000 // Balanced at 0
		}
		return 20000 // All bids, no asks
	}
	return uint32((m.TotalBidVolume * 10000) / m.TotalAskVolume)
}

// LiquidityMetrics tracks liquidity metrics
type LiquidityMetrics struct {
	// TotalLiquidity is total liquidity in the market
	TotalLiquidity uint64 `json:"total_liquidity"`

	// ActiveLiquidity is actively available liquidity
	ActiveLiquidity uint64 `json:"active_liquidity"`

	// LockedLiquidity is locked liquidity
	LockedLiquidity uint64 `json:"locked_liquidity"`

	// LiquidityUtilizationPct is utilization percentage
	LiquidityUtilizationPct uint32 `json:"liquidity_utilization_pct"`

	// UniqueProviders is number of unique liquidity providers
	UniqueProviders uint64 `json:"unique_providers"`

	// ConcentrationHHI is Herfindahl-Hirschman Index for concentration
	ConcentrationHHI uint32 `json:"concentration_hhi"`

	// TopProviderSharePct is market share of top provider
	TopProviderSharePct uint32 `json:"top_provider_share_pct"`

	// Top5ProvidersSharePct is market share of top 5 providers
	Top5ProvidersSharePct uint32 `json:"top_5_providers_share_pct"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// VolumeMetrics tracks trading volume metrics
type VolumeMetrics struct {
	// TotalVolume is total trading volume
	TotalVolume uint64 `json:"total_volume"`

	// OrderVolume is volume from orders
	OrderVolume uint64 `json:"order_volume"`

	// BidVolume is volume from bids
	BidVolume uint64 `json:"bid_volume"`

	// TradeCount is number of trades
	TradeCount uint64 `json:"trade_count"`

	// AverageTradeSize is average trade size
	AverageTradeSize uint64 `json:"average_trade_size"`

	// LargestTrade is the largest trade
	LargestTrade uint64 `json:"largest_trade"`

	// SmallestTrade is the smallest trade
	SmallestTrade uint64 `json:"smallest_trade"`

	// VolumeChange24hPct is 24-hour volume change percentage
	VolumeChange24hPct int32 `json:"volume_change_24h_pct"`

	// VolumeMA7 is 7-day moving average volume
	VolumeMA7 uint64 `json:"volume_ma_7"`

	// VolumeMA30 is 30-day moving average volume
	VolumeMA30 uint64 `json:"volume_ma_30"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// VolatilityMetrics tracks price volatility metrics
type VolatilityMetrics struct {
	// CurrentVolatilityBps is current volatility in basis points
	CurrentVolatilityBps uint32 `json:"current_volatility_bps"`

	// HistoricalVolatilityBps is historical volatility
	HistoricalVolatilityBps uint32 `json:"historical_volatility_bps"`

	// VolatilityMA7 is 7-day volatility moving average
	VolatilityMA7 uint32 `json:"volatility_ma_7"`

	// VolatilityMA30 is 30-day volatility moving average
	VolatilityMA30 uint32 `json:"volatility_ma_30"`

	// MaxDrawdownBps is maximum drawdown in basis points
	MaxDrawdownBps uint32 `json:"max_drawdown_bps"`

	// MaxUptickBps is maximum uptick in basis points
	MaxUptickBps uint32 `json:"max_uptick_bps"`

	// VolatilityRegime indicates current volatility regime
	VolatilityRegime string `json:"volatility_regime"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// DetermineVolatilityRegime determines the volatility regime
func (m *VolatilityMetrics) DetermineVolatilityRegime() string {
	switch {
	case m.CurrentVolatilityBps < 100:
		return "very_low"
	case m.CurrentVolatilityBps < 300:
		return "low"
	case m.CurrentVolatilityBps < 500:
		return "normal"
	case m.CurrentVolatilityBps < 1000:
		return "elevated"
	case m.CurrentVolatilityBps < 2000:
		return "high"
	default:
		return "extreme"
	}
}

// EfficiencyMetrics tracks overall market efficiency
type EfficiencyMetrics struct {
	// EfficiencyScore is the overall efficiency score (0-100)
	EfficiencyScore uint32 `json:"efficiency_score"`

	// SpreadScore is the spread efficiency score (0-100)
	SpreadScore uint32 `json:"spread_score"`

	// FillRateScore is the fill rate efficiency score (0-100)
	FillRateScore uint32 `json:"fill_rate_score"`

	// LiquidityScore is the liquidity efficiency score (0-100)
	LiquidityScore uint32 `json:"liquidity_score"`

	// DepthScore is the depth efficiency score (0-100)
	DepthScore uint32 `json:"depth_score"`

	// VolatilityScore is the volatility efficiency score (0-100)
	VolatilityScore uint32 `json:"volatility_score"`

	// PriceDiscoveryScore is the price discovery efficiency (0-100)
	PriceDiscoveryScore uint32 `json:"price_discovery_score"`

	// HealthStatus indicates overall market health
	HealthStatus string `json:"health_status"`

	// Recommendations are efficiency improvement recommendations
	Recommendations []string `json:"recommendations,omitempty"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// CalculateOverallScore calculates the overall efficiency score
func (m *EfficiencyMetrics) CalculateOverallScore() uint32 {
	// Weighted average of all scores
	weights := map[string]uint32{
		"spread":          20,
		"fill_rate":       20,
		"liquidity":       20,
		"depth":           15,
		"volatility":      10,
		"price_discovery": 15,
	}

	total := m.SpreadScore*weights["spread"] +
		m.FillRateScore*weights["fill_rate"] +
		m.LiquidityScore*weights["liquidity"] +
		m.DepthScore*weights["depth"] +
		m.VolatilityScore*weights["volatility"] +
		m.PriceDiscoveryScore*weights["price_discovery"]

	totalWeight := uint32(0)
	for _, w := range weights {
		totalWeight += w
	}

	return total / totalWeight
}

// DetermineHealthStatus determines the market health status
func (m *EfficiencyMetrics) DetermineHealthStatus() string {
	score := m.CalculateOverallScore()
	switch {
	case score >= 90:
		return "excellent"
	case score >= 75:
		return "healthy"
	case score >= 60:
		return "fair"
	case score >= 40:
		return "stressed"
	case score >= 20:
		return "critical"
	default:
		return "failed"
	}
}

// GenerateRecommendations generates efficiency recommendations
func (m *EfficiencyMetrics) GenerateRecommendations() []string {
	recommendations := make([]string, 0)

	if m.SpreadScore < 50 {
		recommendations = append(recommendations, "increase_market_maker_incentives")
	}
	if m.FillRateScore < 50 {
		recommendations = append(recommendations, "review_order_matching_algorithm")
	}
	if m.LiquidityScore < 50 {
		recommendations = append(recommendations, "increase_liquidity_mining_rewards")
	}
	if m.DepthScore < 50 {
		recommendations = append(recommendations, "attract_more_providers")
	}
	if m.VolatilityScore < 50 {
		recommendations = append(recommendations, "review_circuit_breaker_thresholds")
	}
	if m.PriceDiscoveryScore < 50 {
		recommendations = append(recommendations, "improve_price_oracle_parameters")
	}

	return recommendations
}

// MarketMetrics aggregates all market metrics
type MarketMetrics struct {
	// OfferingID is the offering these metrics are for (empty for global)
	OfferingID *OfferingID `json:"offering_id,omitempty"`

	// Period is the metric period
	Period MetricPeriod `json:"period"`

	// PeriodStart is the start of the period
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the period
	PeriodEnd time.Time `json:"period_end"`

	// StartBlock is the start block
	StartBlock int64 `json:"start_block"`

	// EndBlock is the end block
	EndBlock int64 `json:"end_block"`

	// Spread contains spread metrics
	Spread SpreadMetrics `json:"spread"`

	// FillRate contains fill rate metrics
	FillRate FillRateMetrics `json:"fill_rate"`

	// Depth contains market depth metrics
	Depth MarketDepthMetrics `json:"depth"`

	// Liquidity contains liquidity metrics
	Liquidity LiquidityMetrics `json:"liquidity"`

	// Volume contains volume metrics
	Volume VolumeMetrics `json:"volume"`

	// Volatility contains volatility metrics
	Volatility VolatilityMetrics `json:"volatility"`

	// Efficiency contains efficiency metrics
	Efficiency EfficiencyMetrics `json:"efficiency"`

	// LastUpdated is when metrics were last updated
	LastUpdated time.Time `json:"last_updated"`
}

// NewMarketMetrics creates new market metrics for a period
func NewMarketMetrics(period MetricPeriod, startBlock int64, startTime time.Time) *MarketMetrics {
	endBlock := startBlock + period.BlocksPerPeriod()
	endTime := startTime.Add(time.Duration(period.BlocksPerPeriod()*6) * time.Second)

	return &MarketMetrics{
		Period:      period,
		PeriodStart: startTime,
		PeriodEnd:   endTime,
		StartBlock:  startBlock,
		EndBlock:    endBlock,
		LastUpdated: startTime,
	}
}

// UpdateEfficiencyScores updates all efficiency scores
func (m *MarketMetrics) UpdateEfficiencyScores() {
	// Calculate spread score (lower spread = higher score)
	if m.Spread.AverageSpreadBps == 0 {
		m.Efficiency.SpreadScore = 100
	} else if m.Spread.AverageSpreadBps >= 1000 {
		m.Efficiency.SpreadScore = 0
	} else {
		m.Efficiency.SpreadScore = 100 - (m.Spread.AverageSpreadBps / 10)
	}

	// Calculate fill rate score
	m.Efficiency.FillRateScore = m.FillRate.FillRatePercentage

	// Calculate liquidity score
	if m.Liquidity.LiquidityUtilizationPct >= 90 {
		m.Efficiency.LiquidityScore = 50 // Over-utilized
	} else if m.Liquidity.LiquidityUtilizationPct >= 50 {
		m.Efficiency.LiquidityScore = 100
	} else {
		m.Efficiency.LiquidityScore = m.Liquidity.LiquidityUtilizationPct * 2
	}

	// Calculate depth score (balanced depth = higher score)
	if m.Depth.DepthRatioBps >= 8000 && m.Depth.DepthRatioBps <= 12000 {
		m.Efficiency.DepthScore = 100
	} else if m.Depth.DepthRatioBps >= 5000 && m.Depth.DepthRatioBps <= 15000 {
		m.Efficiency.DepthScore = 75
	} else {
		m.Efficiency.DepthScore = 50
	}

	// Calculate volatility score (lower volatility = higher score)
	if m.Volatility.CurrentVolatilityBps >= 2000 {
		m.Efficiency.VolatilityScore = 0
	} else {
		m.Efficiency.VolatilityScore = 100 - (m.Volatility.CurrentVolatilityBps / 20)
	}

	// Calculate price discovery score (combine multiple factors)
	m.Efficiency.PriceDiscoveryScore = (m.Efficiency.SpreadScore + m.Efficiency.DepthScore) / 2

	// Calculate overall score
	m.Efficiency.EfficiencyScore = m.Efficiency.CalculateOverallScore()
	m.Efficiency.HealthStatus = m.Efficiency.DetermineHealthStatus()
	m.Efficiency.Recommendations = m.Efficiency.GenerateRecommendations()
	m.Efficiency.LastUpdated = m.LastUpdated
}

// MetricsConfig defines metrics collection configuration
type MetricsConfig struct {
	// Enabled indicates if metrics collection is enabled
	Enabled bool `json:"enabled"`

	// BlockMetricsEnabled enables per-block metrics
	BlockMetricsEnabled bool `json:"block_metrics_enabled"`

	// HourlyMetricsEnabled enables hourly metrics
	HourlyMetricsEnabled bool `json:"hourly_metrics_enabled"`

	// DailyMetricsEnabled enables daily metrics
	DailyMetricsEnabled bool `json:"daily_metrics_enabled"`

	// WeeklyMetricsEnabled enables weekly metrics
	WeeklyMetricsEnabled bool `json:"weekly_metrics_enabled"`

	// RetentionBlockMetrics is retention for block metrics in blocks
	RetentionBlockMetrics int64 `json:"retention_block_metrics"`

	// RetentionHourlyMetrics is retention for hourly metrics in hours
	RetentionHourlyMetrics int64 `json:"retention_hourly_metrics"`

	// RetentionDailyMetrics is retention for daily metrics in days
	RetentionDailyMetrics int64 `json:"retention_daily_metrics"`

	// TightSpreadThresholdBps is threshold for tight spread
	TightSpreadThresholdBps uint32 `json:"tight_spread_threshold_bps"`

	// HighVolumeThreshold is threshold for high volume flag
	HighVolumeThreshold uint64 `json:"high_volume_threshold"`

	// LowLiquidityThreshold is threshold for low liquidity warning
	LowLiquidityThreshold uint64 `json:"low_liquidity_threshold"`

	// EfficiencyAlertThreshold is threshold for efficiency alerts
	EfficiencyAlertThreshold uint32 `json:"efficiency_alert_threshold"`
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Enabled:                  true,
		BlockMetricsEnabled:      false, // Too granular by default
		HourlyMetricsEnabled:     true,
		DailyMetricsEnabled:      true,
		WeeklyMetricsEnabled:     true,
		RetentionBlockMetrics:    1000,       // ~1.6 hours
		RetentionHourlyMetrics:   168,        // 1 week
		RetentionDailyMetrics:    90,         // 3 months
		TightSpreadThresholdBps:  100,        // 1%
		HighVolumeThreshold:      1000000000, // 1000 tokens
		LowLiquidityThreshold:    100000000,  // 100 tokens
		EfficiencyAlertThreshold: 50,
	}
}

// Validate validates the metrics configuration
func (c *MetricsConfig) Validate() error {
	if c.RetentionBlockMetrics < 0 {
		return fmt.Errorf("retention_block_metrics cannot be negative")
	}
	if c.RetentionHourlyMetrics < 0 {
		return fmt.Errorf("retention_hourly_metrics cannot be negative")
	}
	if c.RetentionDailyMetrics < 0 {
		return fmt.Errorf("retention_daily_metrics cannot be negative")
	}
	if c.EfficiencyAlertThreshold > 100 {
		return fmt.Errorf("efficiency_alert_threshold cannot exceed 100")
	}
	return nil
}

// MarketMetricsParams holds all metrics parameters
type MarketMetricsParams struct {
	// Config is the metrics configuration
	Config MetricsConfig `json:"config"`

	// GlobalMetricsEnabled enables global (market-wide) metrics
	GlobalMetricsEnabled bool `json:"global_metrics_enabled"`

	// PerOfferingMetricsEnabled enables per-offering metrics
	PerOfferingMetricsEnabled bool `json:"per_offering_metrics_enabled"`

	// AlertsEnabled enables efficiency alerts
	AlertsEnabled bool `json:"alerts_enabled"`

	// AlertRecipientAddress receives alert notifications
	AlertRecipientAddress string `json:"alert_recipient_address,omitempty"`
}

// DefaultMarketMetricsParams returns default market metrics parameters
func DefaultMarketMetricsParams() MarketMetricsParams {
	return MarketMetricsParams{
		Config:                    DefaultMetricsConfig(),
		GlobalMetricsEnabled:      true,
		PerOfferingMetricsEnabled: true,
		AlertsEnabled:             true,
	}
}

// Validate validates the market metrics parameters
func (p *MarketMetricsParams) Validate() error {
	if err := p.Config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	return nil
}

// MetricSnapshot captures a point-in-time metric snapshot
type MetricSnapshot struct {
	// Type is the metric type
	Type MetricType `json:"type"`

	// Value is the metric value
	Value uint64 `json:"value"`

	// ValueBps is the value in basis points (if applicable)
	ValueBps uint32 `json:"value_bps,omitempty"`

	// BlockHeight is the block height
	BlockHeight int64 `json:"block_height"`

	// Timestamp is the snapshot timestamp
	Timestamp time.Time `json:"timestamp"`

	// OfferingID is the offering (if per-offering)
	OfferingID *OfferingID `json:"offering_id,omitempty"`

	// Metadata contains additional context
	Metadata map[string]string `json:"metadata,omitempty"`
}

// MetricAlert represents a metric-based alert
type MetricAlert struct {
	// Type is the metric type that triggered the alert
	Type MetricType `json:"type"`

	// Severity is the alert severity (1-10)
	Severity uint8 `json:"severity"`

	// Message is the alert message
	Message string `json:"message"`

	// CurrentValue is the current metric value
	CurrentValue uint64 `json:"current_value"`

	// ThresholdValue is the threshold that was breached
	ThresholdValue uint64 `json:"threshold_value"`

	// OfferingID is the affected offering (if applicable)
	OfferingID *OfferingID `json:"offering_id,omitempty"`

	// TriggeredAt is when the alert was triggered
	TriggeredAt time.Time `json:"triggered_at"`

	// BlockHeight is the block height when triggered
	BlockHeight int64 `json:"block_height"`

	// Acknowledged indicates if the alert was acknowledged
	Acknowledged bool `json:"acknowledged"`

	// AcknowledgedAt is when it was acknowledged
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`

	// Resolution is the resolution if resolved
	Resolution string `json:"resolution,omitempty"`
}

// MarketHealthDashboard aggregates health indicators
type MarketHealthDashboard struct {
	// OverallHealth is the overall market health (0-100)
	OverallHealth uint32 `json:"overall_health"`

	// Status is the current status
	Status string `json:"status"`

	// ActiveAlerts is the number of active alerts
	ActiveAlerts uint32 `json:"active_alerts"`

	// CriticalAlerts is the number of critical alerts
	CriticalAlerts uint32 `json:"critical_alerts"`

	// Metrics is the current metrics snapshot
	Metrics MarketMetrics `json:"metrics"`

	// Trends contains trend indicators
	Trends map[string]string `json:"trends"`

	// LastUpdated is when the dashboard was last updated
	LastUpdated time.Time `json:"last_updated"`
}

// NewMarketHealthDashboard creates a new market health dashboard
func NewMarketHealthDashboard() *MarketHealthDashboard {
	return &MarketHealthDashboard{
		OverallHealth: 100,
		Status:        "healthy",
		Trends:        make(map[string]string),
	}
}

// UpdateFromMetrics updates the dashboard from metrics
func (d *MarketHealthDashboard) UpdateFromMetrics(metrics MarketMetrics, now time.Time) {
	d.Metrics = metrics
	d.OverallHealth = metrics.Efficiency.EfficiencyScore
	d.Status = metrics.Efficiency.HealthStatus
	d.LastUpdated = now

	// Calculate trends (simplified)
	d.Trends["volume"] = trendStable
	d.Trends["spread"] = trendStable
	d.Trends["liquidity"] = trendStable
}
