# Unified Market Resolution and Waldur-Backed Supply — Design Specification

> **Status:** Accepted — implementation in progress (see §13)  
> **Companion ADR:** `_docs/adr/ADR-010-unified-market-resolution-and-waldur-supply.md`  
> **Supersedes (partially):** the read-only-catalog assumption in ADR-007  
> **Last Updated:** 2026-09-19

This document specifies how VirtEngine unifies its two acquisition styles — automatic
order resolution and manual marketplace browsing — over one canonical supply catalog, and
how Waldur is integrated without compromising consensus determinism.

---

## 1. Problem statement

Today the protocol exposes two half-built acquisition paths:

| Path | Where | Behaviour |
| --- | --- | --- |
| Bidding | `x/market` (`Order -> Bid -> Lease`) | Provider bids; **buyer manually names a bid** to create a lease. No automation. |
| Listing | `x/marketplace` offerings + Waldur bridge | Rich catalog and Waldur lifecycle, but **order/bid/allocation writes are fenced** and the Waldur bridge is default-off. |

There is no component that compares a bid against a fixed-price listing, and no code path
that turns "I want 4 vCPU for ≤ X" into a lease without the buyer hand-picking a provider.

The goals are:

1. A **bidding engine** that can also resolve to Waldur marketplace resources.
2. Providers **supply by creating listings** (native and/or Waldur) as well as by bidding.
3. Users either **submit an order for automatic resolution** or **browse a unified catalog**
   of all public listings (native + ingested Waldur) and buy directly.
4. Waldur is leveraged for provisioning integration, metering, and administration, and its
   marketplace semantics are available **on validator nodes**.

---

## 2. Design principles

### 2.1 Control plane vs. data plane

| Plane | Runs where | Deterministic? | Owns |
| --- | --- | --- | --- |
| Control plane | Every validator (consensus Go modules) | Yes | Catalog, orders, bids, matching, escrow, settlement, capacity |
| Data/execution plane | Off-chain (provider daemons, Waldur Mastermind) | No | Provisioning, Kubernetes/SLURM/OpenStack/AWS/Azure execution, metering, admin UI |

The chain never performs HTTP, reads a wall clock other than `ctx.BlockTime()`, or runs a
database. All nondeterministic work is reduced to **signed inputs** that enter consensus as
deterministic transactions.

### 2.2 Waldur is a semantic reference, not an in-consensus runtime

Waldur's Django/Celery/PostgreSQL stack cannot run inside a validator (ADR-005). The "Waldur
in the chain" requirement is satisfied by porting Waldur's **marketplace model and state
machines** into Go on-chain modules, and treating a live Waldur instance as an off-chain
adapter that both feeds and consumes the chain. Section 8 defines exactly what runs where.

### 2.3 Listings are asks, bids are bids

The unifying abstraction:

- A **listing** (native or Waldur-mirrored) is a **standing ask**: a price and a capacity.
- A **bid** is a **standing bid**: a price for an order window.
- The resolver matches a demand **order** against the union of eligible asks and bids.

This makes "auto-resolve" and "browse-and-buy" two views of one order book.

---

## 3. Target module ownership

| Concern | Module | Change from ADR-007 |
| --- | --- | --- |
| Supply catalog: Offering, PriceComponent, visibility, admin taxonomy | `x/marketplace` | Offering writes become canonical; legacy order/bid/allocation writes stay fenced |
| Demand: Order, Bid, Lease, escrow, settlement lineage | `x/market` | Add offering-aware orders and the **resolution engine** |
| Capacity: Inventory, Reservation | `x/resources` | Unchanged |
| HPC: Cluster, HPCOffering, Job, accounting | `x/hpc` | Offerings mirrored into catalog; jobs remain reservation consumers |
| Execution / metering / Waldur | `pkg/provider_daemon`, `pkg/waldur` | Promoted from legacy path to canonical adapter |
| Admin / taxonomy mirror / catalog serving | `waldur-adapter` (off-chain, new) | Off-chain Waldur-compatible service |

`x/marketplace` continues to be a thin module shell over `x/market/types/marketplace`
(`virtengine/x/marketplace/module.go:19-21`). The `market` store stays binary-keyed;
the `mktplace` store keeps JSON prefixes.

---

## 4. Canonical types

### 4.1 Offering (extend existing)

`x/market/types/marketplace/offering.go` already carries most of what is needed
(`PricingInfo`, `Prices []PriceComponent`, `AllowBidding`, `MinBid`, `Regions`, `Tags`,
`Specifications`, `MaxConcurrentOrders`). Add:

| Field | Type | Purpose |
| --- | --- | --- |
| `Source` | `OfferingSource` (`native`, `waldur`) | Discriminates supply origin |
| `Waldur` | `*WaldurOfferingRef` | `instance_id`, `offering_uuid`, `customer_uuid`, `backend_type`, `snapshot_hash`, `snapshot_height` |
| `Visibility` | `OfferingVisibility` (`public`, `unlisted`, `private`) | Controls catalog browsing vs. direct order |
| `AcquisitionModes` | `[]AcquisitionMode` | Which modes the listing supports (`direct`, `bid`) |
| `AdminPolicy` | `OfferingAdminPolicy` | Moderator/Genesis controls, pause reasons |
| `MeteringProfile` | `string` | Waldur component mapping (`cpu_hours`, `gpu_hours`, ...) |
| `BackendType` | `string` | `kubernetes`, `openstack`, `vmware`, `aws`, `azure`, `slurm`, `moab`, `ood` |

`x/marketplace` becomes the canonical writer for these fields.

### 4.2 Order (extend `x/market`)

Add to the canonical market order model:

| Field | Type | Purpose |
| --- | --- | --- |
| `OfferingID` | optional | Present for `DIRECT` orders bound to a listing |
| `AcquisitionMode` | `AcquisitionMode` | `direct` or `bid` |
| `Selector` | `OfferSelector` | For unbound `DIRECT` orders: `category`, `min_specs`, `regions`, `max_price` |
| `MatchingDeadline` | `*time.Time` | Bid window close; `nil` for `direct` |
| `ReservationID` | `string` | Shared with `x/resources` (already modelled in v1beta5) |

`DIRECT` with `OfferingID` = buy this listing. `DIRECT` with `Selector` = "resolve me the
best matching listing". `BID` = open a bid window.

### 4.3 Bid (extend)

Existing provider bid gains `OfferingID` (optional) so a provider can bid against a
specific listing or an open "intent" order. Supports `quote_kind` = `fixed` or `ask`.

### 4.4 OfferSelector (new)

A deterministic, bounded predicate:

```
category            string
regions             []string
min_specs           map[string]uint64   // e.g. vcpu>=4, memory_gb>=16
max_price           sdk.Coin
identity_required   bool
backend_allowlist   []string
```

No regular expressions, no unbounded scans; evaluation is O(offerings) per order and is
bounded by catalog size (see §11).

---

## 5. Message surface

### 5.1 Supply (providers)

| Message | Module | Purpose |
| --- | --- | --- |
| `MsgCreateOffering` | `x/marketplace` | Publish a native or Waldur-backed listing |
| `MsgUpdateOffering` / `MsgDeactivateOffering` | `x/marketplace` | Lifecycle |
| `MsgSetOfferingVisibility` | `x/marketplace` | `public`/`unlisted`/`private` |
| `MsgIngestWaldurOffering` | `x/marketplace` | Relayer upsert of a signed Waldur snapshot |
| `MsgCreateBid` | `x/market` | Provider quote for a `BID` order |
| `MsgWithdrawBid` | `x/market` | Withdraw an open bid |

### 5.2 Demand (users)

| Message | Module | Purpose |
| --- | --- | --- |
| `MsgCreateOrder` | `x/market` | `direct` (offering or selector) or `bid` |
| `MsgAcceptMatch` | `x/market` | Manual override of an engine result (optional policy) |
| `MsgCancelOrder` | `x/market` | Cancel an open order |
| `MsgTerminateLease` / `MsgResizeLease` / `MsgPauseLease` | `x/market` | Lifecycle (existing semantics) |
| `MsgSubmitJob` | `x/hpc` | HPC job, `direct` or market-backed |

### 5.3 Admin / relayers

| Message | Module | Purpose |
| --- | --- | --- |
| `MsgRegisterWaldurSource` | `x/marketplace` | Register a Waldur instance pubkey + metadata |
| `MsgUpsertCategory` / `MsgUpsertPlanTemplate` | `x/marketplace` | Admin taxonomy |
| `MsgModerateOffering` | `x/marketplace` | Pause/suspend/terminate with reason |
| `MsgWaldurCommandAck` | `x/market` | Adapter acknowledges a emitted command (idempotent) |

---

## 6. State machines

### 6.1 Offering

`Active`, `Paused`, `Suspended`, `Deprecated`, `Terminated` (existing, `offering.go:18-67`).
Add transition: any of `Suspended`/`Paused` -> `Active` only via moderator action; ingestion
may only set `Active`/`Paused`/`Terminated` and never override a moderator `Suspended`.

### 6.2 Order (extend `order.go:20-145`)

```
Open ──rslv(direct)──► Matched ─► Provisioning ─► Active ─► Suspended ─► PendingTermination ─► Terminated
  │                        ▲
  ├──rslv(bid@deadline)────┘
  ├─► Cancelled
  └─► Failed
```

- `DIRECT`: `Open -> Matched` in the same block as creation if a candidate is found.
- `BID`: stays `Open` until `MatchingDeadline`, then `Matched` or `Failed` (no candidate).

### 6.3 Bid

`Open`, `Accepted`, `Rejected`, `Withdrawn`, `Expired` (existing). Resolver sets exactly one
`Accepted`; all others `Rejected`, ties broken deterministically.

### 6.4 Reservation

Unchanged (ADR-007 transition table). The resolver reserves capacity atomically with match;
failure to reserve leaves the order `Open` (re-evaluated next block) or `Failed` at deadline.

---

## 7. Deterministic resolution engine

### 7.1 Location

New package: `x/market/keeper/resolution` (or `x/market/resolution`), invoked from
`EndBlock` and from `MsgCreateOrder` for immediate `DIRECT` matches. It reads the catalog via
an interface into `x/marketplace`, and capacity via `x/resources`.

```
type Catalog interface {
    ActiveOfferings(ctx sdk.Context) []marketplace.Offering
    Offering(ctx sdk.Context, id marketplace.OfferingID) (marketplace.Offering, bool)
}
type Capacity interface { // already available as ResourcesKeeper
    Reserve(ctx, ReservationRequest) (Reservation, error)
    ActivateReservation(ctx, id, link) error
}
```

### 7.2 Algorithm

```
EndBlock:
  for order in OpenOrders():
     if order.mode == DIRECT: tryResolve(order)
     else if order.mode == BID and now >= order.MatchingDeadline: tryResolve(order)
     else if order.ExpiresAt != nil and now >= order.ExpiresAt: order.Failed(NO_MATCH)

tryResolve(order):
  candidates = []
  for offering in Catalog.ActiveOfferings():
     if !offering.Admits(order) continue          // visibility, region, category, specs, identity
     ask = askFor(offering, order)                 // price via CalculateOfferingPrice
     if ask.price > order.MaxBidPrice continue
     candidates.append(ask)
  for bid in OpenBids(order):                      // BID mode only
     if bid.price <= order.MaxBidPrice: candidates.append(bid)
  candidate = rank(candidates)                     // deterministic
  if candidate == nil return
  if Capacity.Reserve(order, candidate) fails return   // retried next block
  createLease(order, candidate, reservation)
  emit WaldurCommand if candidate.source == waldur
```

### 7.3 Ranking (deterministic, integer-only)

Sort key, applied lexicographically (`<` wins):

```
(price,                      // lower total price
 -capacity_fit,              // better spec fit
 -provider_score,            // VEID + benchmark + reputation (x/benchmark)
  source_priority,           // policy: native before waldur, or configurable
  sequence,                  // offering seq or bid seq
  provider_address)          // final tie-break
```

All arithmetic uses `sdkmath.Int`/`uint64`; scores are already integer-based in
`x/benchmark`. No floating point, no map iteration order.

### 7.4 Partial fills

If `RequestedQuantity > 1` and policy `allow_partial_fill` is set, consume candidates in
rank order until filled, creating one lease per provider and recording a
`MatchGroupID`. Capacity is reserved per lease. Default is all-or-nothing.

### 7.5 Front-running mitigation

BID windows close at `MatchingDeadline`; resolution runs in `EndBlock` using only state
committed by then. Bids in the closing block are included regardless of transaction order
within the block, so ordering within the block cannot bias the winner beyond the deterministic
tie-breaks. For price-sensitive intents, orders may require commit-reveal bids in a future
revision (out of scope here).

---

## 8. Waldur integration protocol

### 8.1 What runs where ("Waldur on validator nodes")

| Component | Language | Location | Consensus-critical |
| --- | --- | --- | --- |
| `x/marketplace` marketplace engine (catalog, plans, visibility, taxonomy) | Go | Every validator | **Yes** |
| `x/market` resolution engine | Go | Every validator | **Yes** |
| `x/resources` reservations | Go | Every validator | **Yes** |
| `waldur-adapter` (bridge, ingest relayer, command executor, metering relay) | Go | Provider and/or operator nodes | No |
| Waldur Mastermind (off-chain orchestration, UI, metering, admin) | Python/Django | Operator/provider infrastructure | No |
| Portal catalog view (serves public listings incl. Waldur) | TS | Edge | No |

The phrase "integrate the Waldur codebase directly in the chain" is realised as **the
marketplace engine on validators**, i.e. a Go port of Waldur's catalogue/plan/order/usage
semantics. Waldur Mastermind itself is not embedded: it is the off-chain orchestration and
administration plane that syncs with the chain via signed messages.

### 8.2 Supply ingest (Waldur -> chain)

1. A registered Waldur instance publishes offerings in Waldur as usual (provider UI/API).
2. An authorized relayer (`waldur-adapter` ingest worker, `pkg/provider_daemon/waldur_ingest_worker.go`)
   fetches offerings, normalises them to `WaldurOfferingImport`, and computes a snapshot hash.
3. The relayer signs and submits `MsgIngestWaldurOffering{ snapshot, signature, instance_id }`.
4. Validators verify: signature against the on-chain registered Waldur pubkey, monotonic
   `snapshot_height`, idempotency key, and category/price normalisation.
5. The catalog upserts a `source=waldur` offering; `visibility` defaults from the snapshot but
   moderator `Suspended` is never overridden.

Trust model options (selectable per instance): single provider key for provider-owned
listings, `k`-of-`n` relayer quorum for shared/operator listings. Both are deterministic
signature checks.

### 8.3 Command egress (chain -> Waldur)

On `LeaseCreated` for a `source=waldur` offering the chain emits a durable `WaldurCommand`
(create order, approve, set backend ID) into the existing lifecycle command queue
(`pkg/provider_daemon/lifecycle_command_queue.go`). The adapter executes it against Waldur
and returns `MsgWaldurCommandAck`. Lifecycle callbacks (`MsgWaldurCallback`) become
**reconciliation inputs** verified against the canonical lease, not the driver of state
(contrast with today's `x/market/types/marketplace/keeper/keeper.go:1027-1074`).

### 8.4 Metering and settlement

Waldur component usage (`cpu_hours`, `gpu_hours`, `ram_gb_hours`, `storage_gb_hours`,
`network_gb`; `pkg/waldur/usage.go`) is relayed as authenticated usage attestations
(ADR-006). The chain computes settlement deterministically and reconciles against
provider-reported usage via `pkg/provider_daemon/waldur_reconciler.go`.

### 8.5 Admin and taxonomy

Categories, plan templates, and moderator controls live on-chain in `x/marketplace`
(mirroring Waldur taxonomy) so that the resolver can filter deterministically. Waldur's
admin UI remains the human-friendly console; changes flow to the chain as signed
transactions.

### 8.6 Compatibility with ADR-007

- The financial fence on legacy `mktplace` `Order`/`Bid`/`Allocation` **remains**;
  `ErrLifecycleDeprecated` still routes those writes to `x/market`.
- Only the **Offering** lifecycle is promoted to canonical supply. This is the minimal change
  required for a unified catalog and does not reintroduce a second financial authority.

---

## 9. End-to-end flows

### 9.1 Automatic resolution, native listing

```
User: MsgCreateOrder{mode=direct, selector={category=compute, vcpu>=4, max_price=X}}
  └─ EndBlock resolver: pick cheapest eligible native listing, Reserve capacity
  └─ Lease created -> provider daemon provisions (K8s/OpenStack/...)
```

### 9.2 Automatic resolution, Waldur-backed listing

```
Provider: publishes offering in Waldur -> adapter ingests as source=waldur listing
User: MsgCreateOrder{mode=direct, offering_id=...}
  └─ Resolver matches -> Lease -> WaldurCommand(create order/approve/set backend id)
  └─ Adapter provisions in Waldur; usage relayed; settlement on-chain
```

### 9.3 Manual browse and buy

```
User: portal queries unified catalog (native + public Waldur) -> selects listing
User: MsgCreateOrder{mode=direct, offering_id=<chosen>}
  └─ same resolution path as 9.2
```

### 9.4 Bidding

```
User: MsgCreateOrder{mode=bid, selector=..., max_price=Y, matching_deadline=T}
Providers: MsgCreateBid{order, price<=Y}
  └─ EndBlock at T: rank bids + eligible asks, accept best, reject rest
```

### 9.5 HPC

```
User: MsgCreateOrder{mode=direct|bid, category=hpc, backend=slurm}
  └─ resolves to HPC offering; x/hpc job consumes the lease reservation
  └─ scheduler adapter (SLURM/MOAB/OOD) executes; accounting -> escrow
```

---

## 10. Migration from the current codebase

Phased, each phase independently shippable:

1. **Catalog canonicalisation.** Enable canonical offering writes in `x/marketplace`; add
   `Source`, `WaldurOfferingRef`, `Visibility`, `AcquisitionModes`, `BackendType`,
   `MeteringProfile`. Keep legacy order/bid/allocation fenced. Wire the existing but
   unconstructed `WaldurIngestWorker`/`OfferingPublicationService` into `waldur-adapter`
   and add `MsgIngestWaldurOffering`.
2. **Resolver (DIRECT).** Add `MsgCreateOrder`, `AcquisitionMode`, `OfferSelector`, and the
   `x/market/keeper/resolution` package; run immediate resolution in `MsgCreateOrder` and
   `EndBlock`. Reuse `Resources.Reserve`/`ActivateReservation` and existing lease/escrow
   primitives.
3. **Resolver (BID).** Add `MatchingDeadline`, batch `EndBlock` resolution, deterministic
   ranking, reject losers, optional partial fills.
4. **Waldur command egress and attestation.** Formalise `WaldurCommand`/ack, register Waldur
   source keys, promote `waldur_bridge` from legacy to canonical adapter, wire metering.
5. **Unified catalog surface.** Portal/SDK query merges native and public Waldur listings;
   manual browse-and-buy maps to `DIRECT`.
6. **Deprecation cleanup.** Retire remaining `mktplace` legacy reads; keep historical
   exports per ADR-007.

---

## 11. Risks and mitigations

| Risk | Mitigation |
| --- | --- |
| Resolver scan cost grows with catalog size | Maintain state/category indexes in `x/marketplace` (`keeper/offering_registry.go` already has provider/category/state prefixes); bound candidates per order |
| Ingest trust | On-chain registered Waldur pubkeys, monotonic snapshots, idempotency keys, optional relayer quorum |
| Price normalisation (Waldur decimal vs integer coins) | Normalise at ingest; store integer coin amounts only; reject on overflow |
| Capacity over-commit under partial fills | Atomic per-lease reservations; ADR-007 capacity-conservation invariant |
| Front-running at deadline | End-of-window batch resolution; commit-reveal as a future option |
| Backwards compatibility | Keep ADR-007 financial fence; additive messages; legacy records queryable |

## 12. Open questions

- Default supply priority when a native listing and a Waldur listing are price-equal.
- Whether relayer quorum is required or provider self-attestation suffices for provider-owned
  listings.
- Partial-fill policy granularity (per offering vs. per order).
- Whether `MsgAcceptMatch` manual override is enabled by default.

---

## Appendix A — Key code references

| Concern | Reference |
| --- | --- |
| Marketplace offering model | `x/market/types/marketplace/offering.go:316-535` |
| `allow_bidding` / `min_bid` | `x/market/types/marketplace/offering.go:342-346`, `keeper/keeper.go:399-406` |
| Order model | `x/market/types/marketplace/order.go:237-297`, transitions `:122-145` |
| Manual bid acceptance | `x/market/types/marketplace/keeper/keeper.go:596-668` |
| Canonical market lease + reservation | `x/market/handler/server.go:207-306` |
| Canonical lifecycle fence | `x/market/types/marketplace/keeper/keeper.go:180-186`, `:354-357` |
| Capacity reservations | `x/resources/keeper/reservation.go:41-225` |
| HPC market-backed job | `x/hpc/keeper/keeper.go:596-659` |
| Waldur bridge | `pkg/provider_daemon/waldur_bridge.go`, `pkg/waldur/marketplace.go` |
| Waldur ingest types | `x/market/types/marketplace/waldur_ingest.go:20-467` |
| Service types / specs | `x/market/types/marketplace/service_types.go` |
| Determinism | `_docs/adr/ADR-005-consensus-determinism.md` |
| Authenticated metering | `_docs/adr/ADR-006-authenticated-metering.md` |
| Canonical market/reservations | `_docs/adr/ADR-007-canonical-market-reservations.md` |
| Waldur mapping | `_docs/architecture/waldur-market-mapping-spec.md` |

---

## 13. Implementation status

All five phases are implemented (PR #899), tested, and wired end to end.

### Landed

| Area | Location |
| --- | --- |
| Supply model (source, visibility, Waldur ref, backends, acquisition modes, selector) | `x/market/types/marketplace/offering_source.go` |
| Offering/order extensions + validation | `x/market/types/marketplace/offering.go`, `order.go` |
| Deterministic ranking and selection | `x/market/types/marketplace/resolution.go` |
| Waldur command, source, and attestation types | `x/market/types/marketplace/waldur_command.go` |
| Waldur ingest mapping extensions | `x/market/types/marketplace/waldur_ingest.go` |
| Resolver keeper + capacity interface | `x/market/types/marketplace/keeper/resolution.go` |
| Source registry, signed ingest, command queue, unified catalog | `x/market/types/marketplace/keeper/waldur.go` |
| EndBlock resolution wiring (param-gated) | `x/marketplace/module.go` |
| Resolution parameters + ingest provider map | `x/market/types/marketplace/genesis.go` |
| Genesis persistence of registered Waldur sources | `x/market/types/marketplace/genesis.go`, `x/marketplace/genesis.go` |
| Protobuf contract for new messages/queries | `sdk/proto/node/virtengine/marketplace/v1/{tx,query,types}.proto` |
| Msg/Query handlers, codec registration, converters, canonical writers | `keeper/msg_server.go`, `keeper/query_server.go`, `keeper/convert.go`, `x/market/types/marketplace/msgs.go`, `codec.go` |
| Regenerated Go/TS/OpenAPI marketplace contracts | `sdk/go/node/marketplace/v1/*`, `sdk/ts/src/generated/**`, `api/openapi/virtengine-proto.swagger.json` |
| x/resources capacity adapter + app wiring | `app/types/marketplace_capacity.go`, `app/types/app.go` |
| Provider-daemon snapshot ingest + command poller + mutation kinds + CLI flags | `pkg/provider_daemon/waldur_snapshot_ingest.go`, `waldur_command_poller.go`, `provider_mutation.go`, `cmd/provider-daemon/main.go` |
| `v1.9.0` governance upgrade enabling auto-resolution | `upgrades/software/v1.9.0/`, `upgrades/types/types.go` |

Behaviour is disabled by default: `Params.EnableAutoResolution` is `false`, so
`EndBlock` is a no-op until the `v1.9.0` governance upgrade sets the parameter
under its preconditions (canonical reservations active, canonical fence active,
no live legacy records).

### Remaining (operator/CI steps, not code gaps)

1. Run the pinned Docker/WSL codegen in CI to confirm the locally generated
   contracts.
2. Deploy a Waldur-side (or operator-held) snapshot signer and register its key
   via `MsgRegisterWaldurSource`.
3. Submit the `v1.9.0` governance proposal after the mainnet readiness checks
   in `_docs/operations/`.

