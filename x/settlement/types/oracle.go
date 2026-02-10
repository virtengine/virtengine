package types

import (
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
)

// OracleSourceType defines the oracle source type.
type OracleSourceType string

const (
	OracleSourceTypeCosmosOracle OracleSourceType = "cosmos_oracle"
	OracleSourceTypeBandIBC      OracleSourceType = "band_ibc"
	OracleSourceTypeChainlinkIBC OracleSourceType = "chainlink_ibc"
	OracleSourceTypeManual       OracleSourceType = "manual"
)

// CurrencyPair defines a base/quote currency pair.
type CurrencyPair struct {
	Base  string `json:"base"`
	Quote string `json:"quote"`
}

// String returns the pair formatted as BASE/QUOTE.
func (p CurrencyPair) String() string {
	return fmt.Sprintf("%s/%s", strings.ToUpper(p.Base), strings.ToUpper(p.Quote))
}

// Validate ensures the pair is valid.
func (p CurrencyPair) Validate() error {
	if strings.TrimSpace(p.Base) == "" || strings.TrimSpace(p.Quote) == "" {
		return ErrInvalidParams.Wrap("currency pair requires base and quote")
	}
	if strings.EqualFold(p.Base, p.Quote) {
		return ErrInvalidParams.Wrap("currency pair base and quote must differ")
	}
	return nil
}

// Price represents a price for a currency pair.
type Price struct {
	Base      string            `json:"base"`
	Quote     string            `json:"quote"`
	Rate      sdkmath.LegacyDec `json:"rate"`
	Timestamp time.Time         `json:"timestamp"`
	Source    string            `json:"source"`
}

// Pair returns the currency pair for the price.
func (p Price) Pair() CurrencyPair {
	return CurrencyPair{Base: p.Base, Quote: p.Quote}
}

// Validate ensures the price entry is valid.
func (p Price) Validate() error {
	if strings.TrimSpace(p.Base) == "" || strings.TrimSpace(p.Quote) == "" {
		return ErrInvalidParams.Wrap("price requires base and quote")
	}
	if !p.Rate.IsPositive() {
		return ErrInvalidParams.Wrap("price rate must be positive")
	}
	return nil
}

// PriceUpdate represents a streaming price update.
type PriceUpdate struct {
	Pair  CurrencyPair `json:"pair"`
	Price Price        `json:"price"`
	Err   string       `json:"err,omitempty"`
}

// OracleSourceConfig configures an oracle data source.
type OracleSourceConfig struct {
	ID       string           `json:"id"`
	Type     OracleSourceType `json:"type"`
	Enabled  bool             `json:"enabled"`
	Priority uint32           `json:"priority"`
}

// Validate ensures the oracle source config is valid.
func (c OracleSourceConfig) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return ErrInvalidParams.Wrap("oracle source id required")
	}
	switch c.Type {
	case OracleSourceTypeCosmosOracle, OracleSourceTypeBandIBC, OracleSourceTypeChainlinkIBC, OracleSourceTypeManual:
		return nil
	default:
		return ErrInvalidParams.Wrapf("unsupported oracle source type: %s", c.Type)
	}
}

// ManualPriceOverride defines a governance-set emergency price.
type ManualPriceOverride struct {
	Base      string            `json:"base"`
	Quote     string            `json:"quote"`
	Rate      sdkmath.LegacyDec `json:"rate"`
	UpdatedAt time.Time         `json:"updated_at"`
	ExpiresAt time.Time         `json:"expires_at"`
}

// IsExpired returns true when the override is expired at the provided time.
func (m ManualPriceOverride) IsExpired(at time.Time) bool {
	if m.ExpiresAt.IsZero() {
		return false
	}
	return at.After(m.ExpiresAt)
}

// Validate ensures the manual price override is valid.
func (m ManualPriceOverride) Validate() error {
	if strings.TrimSpace(m.Base) == "" || strings.TrimSpace(m.Quote) == "" {
		return ErrInvalidParams.Wrap("manual price requires base and quote")
	}
	if !m.Rate.IsPositive() {
		return ErrInvalidParams.Wrap("manual price rate must be positive")
	}
	if !m.ExpiresAt.IsZero() && !m.UpdatedAt.IsZero() && m.ExpiresAt.Before(m.UpdatedAt) {
		return ErrInvalidParams.Wrap("manual price expires before update time")
	}
	return nil
}

// PriceAlert captures anomalous price movements.
type PriceAlert struct {
	Base         string            `json:"base"`
	Quote        string            `json:"quote"`
	OldRate      sdkmath.LegacyDec `json:"old_rate"`
	NewRate      sdkmath.LegacyDec `json:"new_rate"`
	ChangePct    sdkmath.LegacyDec `json:"change_pct"`
	OccurredAt   time.Time         `json:"occurred_at"`
	WindowSec    uint64            `json:"window_sec"`
	ThresholdBps uint32            `json:"threshold_bps"`
	Source       string            `json:"source"`
}

// SettlementRateStatus captures the status of a fiat rate lock.
type SettlementRateStatus string

const (
	SettlementRateStatusLocked  SettlementRateStatus = "locked"
	SettlementRateStatusPending SettlementRateStatus = "pending"
	SettlementRateStatusFailed  SettlementRateStatus = "failed"
)

// LockedRate captures a locked oracle rate with spread applied.
type LockedRate struct {
	Base      string            `json:"base"`
	Quote     string            `json:"quote"`
	Source    string            `json:"source"`
	RawRate   sdkmath.LegacyDec `json:"raw_rate"`
	SpreadBps uint32            `json:"spread_bps"`
	FinalRate sdkmath.LegacyDec `json:"final_rate"`
	LockedAt  time.Time         `json:"locked_at"`
}

// SettlementRateLock stores the locked settlement rates for audit.
type SettlementRateLock struct {
	SettlementID string               `json:"settlement_id"`
	InvoiceID    string               `json:"invoice_id"`
	Status       SettlementRateStatus `json:"status"`
	Rates        []LockedRate         `json:"rates"`
	LockedAt     time.Time            `json:"locked_at"`
	Reason       string               `json:"reason,omitempty"`
}
