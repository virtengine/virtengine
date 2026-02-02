package v1

import (
	"fmt"
	"time"
)

const (
	DefaultOracleOutlierThresholdBps   = uint32(150)
	DefaultSettlementEpochName         = "bme"
	DefaultCircuitBreakerWarnThreshold = uint32(9500)
	DefaultCircuitBreakerHaltThreshold = uint32(9000)
	DefaultMintSpreadBps               = uint32(25)
	DefaultSettleSpreadBps             = uint32(0)
	DefaultMinEpochBlocks              = 10
	DefaultEpochBlocksBackoff          = 10
)

var DefaultOracleTWAPWindow = time.Hour

func DefaultParams() Params {
	return Params{
		CircuitBreakerWarnThreshold: DefaultCircuitBreakerWarnThreshold,
		CircuitBreakerHaltThreshold: DefaultCircuitBreakerHaltThreshold,
		MintSpreadBps:               DefaultMintSpreadBps,
		SettleSpreadBps:             DefaultSettleSpreadBps,
		MinEpochBlocks:              DefaultMinEpochBlocks,
		EpochBlocksBackoff:          DefaultEpochBlocksBackoff,
	}
}

func (p Params) Validate() error {
	if p.MinEpochBlocks <= 0 {
		return fmt.Errorf("min_epoch_blocks must be positive")
	}
	if p.CircuitBreakerWarnThreshold > 10000 {
		return fmt.Errorf("circuit_breaker_warn_threshold cannot exceed 10000")
	}
	if p.CircuitBreakerHaltThreshold > 10000 {
		return fmt.Errorf("circuit_breaker_halt_threshold cannot exceed 10000")
	}
	if p.CircuitBreakerWarnThreshold <= p.CircuitBreakerHaltThreshold {
		return fmt.Errorf("warn threshold must be greater than halt threshold")
	}
	if p.MintSpreadBps > 1000 {
		return fmt.Errorf("mint_spread_bps cannot exceed 1000 (10%%)")
	}
	if p.SettleSpreadBps > 1000 {
		return fmt.Errorf("settle_spread_bps cannot exceed 1000 (10%%)")
	}
	return nil
}
