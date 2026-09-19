# ADR-010: Unified Market Resolution and Waldur-Backed Supply

## Status

Accepted — implementation in progress (see "Implementation status")

## Date

2026-09-19

## Context

The original patent (`AU2024203136A1`) describes the VirtEngine Cloud Marketplace as the
primary acquisition surface, powered by Waldur, with a provider "bid engine" as automation.
Claim 11 is a listing/provisioning flow (providers list services, users request, resources
are determined, allocated, metered, and invoiced). Claim 13 offers the SLURM supercomputer
"from a cloud marketplace system". Nothing in the patent describes an on-chain order-book
matcher; the closest construct is the provider daemon's bid engine, which quotes prices for
existing orders.

ADR-007 subsequently established the canonical ownership split:

- `x/market` is the only mutable financial owner (`Order -> Bid -> Lease`, escrow).
- `x/resources` is the only mutable capacity/reservation owner.
- `mktplace` (`x/marketplace`) becomes a supply catalog and compatibility module; its
  order/bid/allocation writes are fenced behind `ErrLifecycleDeprecated`.
- `x/hpc` is a reservation consumer, not a parallel capacity ledger.

Three gaps remain after ADR-007:

1. **There is no automated resolution.** A buyer must manually name a bid
   (`MsgCreateLease{bid_id}` in `x/market/handler/server.go:207-306`, `MsgAcceptBid{bid_id}`
   in `x/market/types/marketplace/keeper/msg_server.go:236-286`). `EndBlock` is a no-op in
   both market modules. The "matching engine" is a narrative, not a component.
2. **The two acquisition styles are disconnected.** Fixed-price catalog listings
   (`x/marketplace` offerings) and provider bids (`x/market`) never share a book, a resolver,
   or a price comparison.
3. **Waldur is both over- and under-used.** It is the richest marketplace/metering/admin
   system, but the bridge is disabled by default (`--waldur-enabled=false`,
   `cmd/provider-daemon/main.go:429`), the ingest worker is implemented but not wired into
   any binary, and canonical writes through Waldur are fenced. Meanwhile the deterministic
   chain cannot host Waldur's non-deterministic Django/Celery/PostgreSQL runtime
   (ADR-005 consensus determinism).

The product direction is a **two-style system**:

- Automated: a user submits an order to the market and the engine resolves it.
- Manual: a user browses a unified catalog (including all public Waldur listings) and
  purchases directly.

Providers supply either by publishing listings (native on-chain and/or Waldur marketplace
offerings) or by bidding on open orders.

## Decision

### 1. One deterministic resolution engine, two supply styles

A **supply listing is a standing ask**; a **provider bid is a standing bid**. A single
resolver in `x/market` matches demand orders against both. This unifies the two styles
without introducing a second financial authority.

- Add `AcquisitionMode` to orders: `DIRECT` and `BID`.
- `DIRECT` orders resolve immediately (same block) against eligible active listings.
- `BID` orders accumulate bids and resolve at a deterministic `matching_deadline`.
- Resolution selects the best eligible candidate (listing ask or bid) using a
  deterministic, integer-only ranking.

### 2. Offering catalog becomes canonical supply, independent of the legacy allocation writer

`x/marketplace` remains the canonical supply catalog (offerings, plan/price components,
visibility, admin taxonomy, identity/MFA policy, `allow_bidding`/`min_bid`, backend type),
but its offering lifecycle writes are re-enabled as canonical. Only the legacy
`Order`/`Bid`/`Allocation` writer stays deprecated and routed to `x/market`; existing
records remain queryable.

### 3. Waldur is a deterministic mirror plus a signed off-chain adapter

Waldur's Django runtime is never consensus-critical. "Waldur in the chain" means porting
Waldur's marketplace **semantics/data model** into consensus Go, with Waldur as an
off-chain orchestration/metering/admin service:

- **Ingest (Waldur -> chain):** authorized relayers submit
  `MsgIngestWaldurOffering` carrying the offering snapshot and a signature from a Waldur
  key registered on-chain. Validators verify deterministically and upsert a
  `source=waldur` listing. The HTTP fetch happens off-chain; only signed results enter
  consensus.
- **Command (chain -> Waldur):** offering/order/lease events emit durable signed
  `WaldurCommand`s (create offering, create order, approve, set backend ID, lifecycle,
  usage), consumed off-chain by provider- or operator-run Waldur adapters. This reuses the
  existing `pkg/provider_daemon/waldur_bridge.go` and lifecycle command queue.
- **Metering:** Waldur usage is submitted as authenticated attestations and reconciled
  on-chain per ADR-006. Chain state is authoritative for settlement; Waldur is the
  measurement source.

### 4. HPC is a catalog category and reservation consumer

HPC offerings are catalog entries with `category=hpc` and `backend_type in {slurm, moab,
ood}`. HPC jobs remain reservation consumers. A job may be acquired `DIRECT` or `BID` and
executed on the provider's scheduler backend; the existing market-backed job path
(`x/hpc/keeper/keeper.go:596-659`) is the mechanism, not a parallel market.

### 5. Module ownership after this ADR

| Concern | Owner | Change |
| --- | --- | --- |
| Supply catalog (offerings, plans, visibility, taxonomy) | `x/marketplace` | canonical writes re-enabled for offerings only |
| Orders, bids, leases, escrow, **resolution engine** | `x/market` | add offerings-aware orders + deterministic resolver |
| Capacity / reservations | `x/resources` | unchanged |
| HPC jobs / scheduler routing | `x/hpc` | unchanged; consumes reservations |
| Waldur / provider execution & metering | `pkg/provider_daemon`, `pkg/waldur` | promoted from legacy path to canonical adapter |

This ADR supersedes ADR-007 only where ADR-007 treats the offering catalog as read-only and
Waldur writes as the deprecated path; the financial/capacity ownership of ADR-007 is
unchanged.

## Consequences

- A buyer gets one consistent experience: submit an order for automatic resolution, or
  browse the catalog and buy directly.
- A provider can supply by listing (native or Waldur) or by bidding, with all three
  represented in the same resolver.
- Waldur remains fully usable for provisioning, metering, and administration while the
  chain stays deterministic and is the source of truth for commercial state.
- New consensus surface: `MsgIngestWaldurOffering`, matching deadline/expiry processing,
  and partial-fill accounting.
- New migration surface: legacy `mktplace` offerings gain `source`/plan fields; legacy
  order/bid/allocation records are quarantined per ADR-007.

## Alternatives rejected

### Run Waldur's Django runtime inside validators

Rejected: non-deterministic (PostgreSQL, Celery, wall-clock, outbound HTTP), heavyweight,
and upgrade-coupled to consensus. It would violate ADR-005 determinism and halt the chain
on any nondeterministic divergence.

### Keep manual bid acceptance as the only path

Rejected: does not satisfy the automated-resolution requirement and leaves fixed-price
listings and bids in separate marketplaces.

### Make `mktplace` the financial owner (ADR-007 alternative re-proposed)

Rejected for the same reasons ADR-007 rejected it: it would discard mature escrow,
deployment, and migration history and require converting all clients.

### Treat Waldur listings as fully off-chain, not mirrored

Rejected: the resolver and catalog must be deterministic for consensus; mirroring signed
snapshots is required for a single unified catalog.

## References

- `_docs/architecture/unified-market-waldur-design.md` (detailed specification)
- `_docs/adr/ADR-005-consensus-determinism.md`
- `_docs/adr/ADR-006-authenticated-metering.md`
- `_docs/adr/ADR-007-canonical-market-reservations.md`
- `_docs/architecture/waldur-market-mapping-spec.md`
- `VEPatentResponse/VE/patent_text_extract.txt`

## Implementation status

Landed (consensus Go, no protobuf changes required):

- Offering supply model: `Source`, `Visibility`, `Waldur`, `AcquisitionModes`,
  `BackendType`, `MeteringProfile`; selector/admission predicates.
  (`x/market/types/marketplace/offering_source.go`, `offering.go`)
- Order acquisition model: `AcquisitionMode`, `Selector`, `MatchingDeadline`,
  plus bid-window helpers. (`x/market/types/marketplace/order.go`)
- Deterministic resolution engine: ranking and selection with all-or-nothing and
  partial-fill policies. (`x/market/types/marketplace/resolution.go`)
- Keeper resolver + EndBlock wiring, gated by `Params.EnableAutoResolution`.
  (`keeper/resolution.go`, `x/marketplace/module.go`)
- Waldur source registry, signed offering ingest with replay protection, and the
  durable Waldur command queue. (`keeper/waldur.go`,
  `x/market/types/marketplace/waldur_command.go`)
- Unified catalog browse API. (`keeper/waldur.go` `UnifiedCatalog`)
- Tests for the engine, ingest, resolution, catalog, and commands.

Deferred (requires protobuf regeneration, which is Docker/WSL-gated in this
repository; the `.proto` contract is authored and validated with `buf lint`
and a full descriptor build):

- Generated Go for the new messages/queries, Msg/Query server method
  implementations, and converter updates.
- Wiring the provider-daemon Waldur ingest worker and offering publication
  service to submit the new signed ingestions and consume the command queue.
