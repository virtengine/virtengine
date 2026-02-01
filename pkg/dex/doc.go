// Package dex provides DEX (Decentralized Exchange) integration for crypto-to-fiat
// conversions and liquidity operations in the VirtEngine marketplace.
//
// VE-905: DEX integration for crypto-to-fiat conversions
//
// This package implements:
//   - Multi-DEX protocol client interface (Uniswap, Osmosis, etc.)
//   - Price feed integration with oracle-style aggregation
//   - Swap execution helpers with slippage protection
//   - Liquidity pool queries and analytics
//   - Fiat off-ramp bridge interface for marketplace settlements
//
// Architecture:
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│                          DEX Service                                │
//	├─────────────────────────────────────────────────────────────────────┤
//	│  PriceFeed          │  SwapExecutor       │  OffRampBridge         │
//	│  - Multi-source     │  - Route finding    │  - Fiat partners       │
//	│  - TWAP/VWAP       │  - Slippage mgmt    │  - KYC integration     │
//	│  - Staleness check │  - MEV protection   │  - Settlement          │
//	├─────────────────────────────────────────────────────────────────────┤
//	│                        DEX Adapters                                 │
//	├─────────────┬─────────────┬─────────────┬─────────────┬────────────┤
//	│   Uniswap   │   Osmosis   │   Curve     │   1inch     │  Custom    │
//	└─────────────┴─────────────┴─────────────┴─────────────┴────────────┘
//
// Security Considerations:
//   - All price feeds use multi-source validation to prevent manipulation
//   - Swap execution includes circuit breakers for abnormal price movements
//   - Off-ramp bridges require KYC verification via VEID integration
//   - Private keys never pass through this package; signing is external
//
// Usage:
//
//	cfg := dex.DefaultConfig()
//	service, err := dex.NewService(cfg)
//	if err != nil {
//	    return err
//	}
//
//	// Get current price
//	price, err := service.GetPrice(ctx, "UVE", "USDC")
//
//	// Execute swap
//	quote, err := service.GetSwapQuote(ctx, dex.SwapRequest{...})
//	result, err := service.ExecuteSwap(ctx, quote, signedTx)
//
//	// Off-ramp to fiat
//	settlement, err := service.InitiateOffRamp(ctx, dex.OffRampRequest{...})
package dex

