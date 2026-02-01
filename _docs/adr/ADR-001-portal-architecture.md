# ADR-001: VE Portal Architecture Decision

## Status

**PROPOSED** - February 2026

## Context

VirtEngine needs a customer-facing portal for interacting with the decentralized marketplace. Currently:

1. **Waldur HomePort** is the existing UI for Waldur (provider backend)
2. **HomePort architecture**: Angular 15+, NgRx state, Bootstrap UI, direct Waldur REST API
3. **Chain-first requirement**: The blockchain should be the source of truth, not Waldur
4. **Pricing models differ**: Waldur uses fixed USD/unit pricing; Akash (our fork origin) uses competitive bidding

This ADR evaluates three approaches for the customer portal.

## Decision Drivers

1. **Decentralization**: Chain must be source of truth for offerings, orders, payments
2. **Development velocity**: Time to production-ready portal
3. **Maintenance burden**: Long-term cost of maintaining the solution
4. **User experience**: Modern, responsive, wallet-integrated interface
5. **Pricing flexibility**: Support both fixed pricing AND optional bidding/spot pricing

## Options Considered

### Option A: Fork and Modify Waldur HomePort

**Description**: Fork HomePort, add blockchain wallet integration, modify API calls to use chain.

**Pros**:
- Existing tested UI components
- Familiar to Waldur operators
- Feature-complete for resource management

**Cons**:
- **Angular vs React mismatch**: HomePort is Angular; we're standardizing on React
- **Deep Waldur coupling**: Every component assumes Waldur REST API structure
- **No wallet support**: Would need to add Cosmos wallet integration to Angular
- **Architectural mismatch**: HomePort assumes Waldur is source of truth
- **Dual maintenance**: Would need to track upstream HomePort changes

**Effort estimate**: 8-12 weeks to fork + 16-20 weeks to refactor for chain-first

### Option B: Hybrid - Embed HomePort for Provider Admin, Build Chain Portal

**Description**: Keep HomePort as-is for provider administration. Build separate customer portal that queries chain.

**Pros**:
- Providers keep familiar Waldur admin UI
- Customer portal is chain-native from day one
- Clean separation of concerns

**Cons**:
- Two separate UIs to maintain
- Provider admin actions still go through Waldur (not on-chain)
- Potential UX inconsistency between portals

**Effort estimate**: 6-8 weeks for customer portal + ongoing dual maintenance

### Option C: Build VE Portal from Scratch (React + Cosmos SDK) ✅ RECOMMENDED

**Description**: Build modern React portal that queries chain directly. Waldur remains provider-side only.

**Pros**:
- **Chain-first architecture**: All reads from chain, all writes are chain transactions
- **Modern stack**: React 18, TypeScript, Vite, TanStack Query, Tailwind
- **Native wallet integration**: @cosmos-kit/react for Keplr, Leap, WalletConnect
- **Single source of truth**: No API translation layer
- **Pricing flexibility**: Can implement both fixed pricing AND bidding
- **Mobile-ready**: React Native path for future mobile app

**Cons**:
- More upfront development (no existing components to reuse)
- Need to implement all UI components

**Effort estimate**: 10-14 weeks to MVP, feature parity with HomePort by week 20

## Architecture Analysis

### Why HomePort Doesn't Fit

```
┌─────────────────────────────────────────────────────────────┐
│                    CURRENT HOMEPORT                         │
├─────────────────────────────────────────────────────────────┤
│  Angular App                                                │
│     │                                                       │
│     ▼                                                       │
│  Waldur REST API ──────────────────▶ PostgreSQL            │
│     │                                                       │
│  (Waldur is source of truth)                               │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    VE PORTAL (DESIRED)                      │
├─────────────────────────────────────────────────────────────┤
│  React App                                                  │
│     │                                                       │
│     ├──▶ Chain (gRPC/REST) ──────────▶ CometBFT/Cosmos     │
│     │       (source of truth)                               │
│     │                                                       │
│     └──▶ Provider Daemon (usage metrics, logs)             │
│                │                                            │
│                ▼                                            │
│            Waldur (provider-side only)                     │
└─────────────────────────────────────────────────────────────┘
```

HomePort would require:
1. Replacing ALL API calls with chain queries
2. Adding transaction signing for every write operation
3. Implementing wallet connection (not trivial in Angular)
4. Removing Waldur-specific features (OIDC, SAML, internal billing)
5. Rewriting state management for blockchain async patterns

This is effectively a rewrite wearing HomePort's clothes.

### Pricing Model Decision

**Waldur's Model** (current):
```
Offering:
  - RAM: $10/GB/month
  - CPU: $5/vCPU/month
  - Storage: $0.10/GB/month
  
Order total = sum(units × price)
```

**Akash's Model** (bidding):
```
Order:
  - Requested: 4 CPU, 8GB RAM, 100GB storage
  - Budget: 100 AKT/month
  
Providers bid:
  - Provider A: 80 AKT/month
  - Provider B: 95 AKT/month
  - Provider C: 75 AKT/month ✓ (winner)
```

**Recommended VirtEngine Model** (hybrid):

```protobuf
message Offering {
  // Provider-set pricing (Waldur-style)
  repeated PriceComponent fixed_prices = 1;
  
  // Optional: Allow bidding for spot/preemptible
  bool allow_bidding = 2;
  int64 min_bid_amount = 3;  // Floor price
  
  // Payment terms
  string payment_denom = 4;  // "uve"
  
  // USD oracle reference for price stability
  string price_oracle = 5;   // e.g., "cosmos-oracle-v1"
}

message PriceComponent {
  string resource_type = 1;  // "cpu", "ram", "storage", "gpu"
  string unit = 2;           // "vcpu", "gb", "hour"
  cosmos.base.v1beta1.Coin price = 3;
  
  // Optional USD equivalent for display
  string usd_equivalent = 4;
}
```

**Benefits of hybrid approach**:
1. Providers set fixed prices (familiar Waldur workflow)
2. Optional bidding for spot pricing / price discovery
3. USD oracle integration for stable pricing display
4. Chain stores prices in UVE; UI shows USD equivalent

### USD → UVE Price Oracle

For live USD pricing, we need an oracle. Options:

1. **Existing Cosmos oracles**: Band Protocol, Pyth Network
2. **Custom oracle module**: `x/oracle` with validator-submitted prices
3. **External API + caching**: Fallback for early stage

Recommended: Start with external API (CoinGecko/CoinMarketCap), add proper oracle module for mainnet.

```go
// x/oracle/keeper/keeper.go
type IKeeper interface {
    GetExchangeRate(ctx sdk.Context, denom string, quote string) (sdk.Dec, error)
    SetExchangeRate(ctx sdk.Context, denom string, quote string, rate sdk.Dec) error
}
```

## Decision

**Option C: Build VE Portal from Scratch**

### Rationale

1. **Technical fit**: React + Cosmos SDK is the industry standard for chain UIs
2. **Clean architecture**: No legacy coupling to work around
3. **Pricing flexibility**: Can implement Waldur-style fixed pricing AND optional bidding
4. **Long-term maintainability**: Single codebase, single stack, chain-native
5. **Developer ecosystem**: More React/Cosmos developers than Angular/Waldur

### What We Keep from Waldur

Waldur remains valuable as **provider backend**:
- Provider-side resource management
- Infrastructure adapters (K8s, OpenStack, SLURM)
- Usage metering and reporting
- Provider's internal billing/accounting

The provider-daemon bridges Waldur → Chain:
```
Chain Order → provider-daemon → Waldur (provision) → provider-daemon → Chain (status update)
```

### Portal Feature Scope

| Feature | Phase 1 (MVP) | Phase 2 | Phase 3 |
|---------|---------------|---------|---------|
| Wallet connect | ✅ | | |
| Browse offerings | ✅ | | |
| Create orders | ✅ | | |
| View allocations | ✅ | | |
| Escrow/payments | ✅ | | |
| Provider dashboard | | ✅ | |
| VEID verification | | ✅ | |
| HPC job submission | | ✅ | |
| Bidding system | | | ✅ |
| Mobile app | | | ✅ |

## Consequences

### Positive

- Clean chain-first architecture
- Modern developer experience
- Pricing model flexibility
- Path to mobile app via React Native

### Negative

- More upfront development effort
- No UI components to reuse from HomePort
- Need to design all screens from scratch

### Risks

- **Risk**: Development takes longer than estimated
  **Mitigation**: MVP scope is minimal; iterate fast

- **Risk**: USD oracle complexity
  **Mitigation**: Start with external API; formalize oracle for mainnet

## Related Decisions

- ADR-002: Pricing Oracle Implementation (TBD)
- ADR-003: Bidding System Design (TBD)

## Task Revisions

Based on this ADR, revise the portal tasks:

### Remove/Defer
- 23E (Bid review) → Defer to Phase 3
- 23W (HomePort reuse analysis) → Superseded by this ADR

### Add
- Pricing oracle integration task
- USD display component
- Fixed pricing checkout flow

### Modify
- 23C, 23D: Focus on fixed pricing first, bidding later
- 23K: Add USD/UVE conversion display

## References

- [Waldur HomePort](https://github.com/waldur/waldur-homeport)
- [Cosmos Kit](https://cosmoskit.com/)
- [Akash Network Console](https://github.com/akash-network/console)
- [Band Protocol](https://bandprotocol.com/)
