# ADR-007: Canonical Market Lifecycle and Authoritative Reservations

**Status:** Accepted for local implementation
**Date:** 2026-07-21
**Upgrade:** `v1.6.0`
**Decision authority:** VirtEngine protocol continuation Task 84C

## Context

VirtEngine registers two mutable marketplace lifecycles:

1. `x/market` owns generated `Order -> Bid -> Lease` state in store `market`.
2. `x/market/types/marketplace` is registered through `x/marketplace` as module/store `mktplace` and owns `Offering -> Order -> Bid -> Allocation` state.

`x/resources` independently subtracts inventory for allocations, while `x/hpc` independently selects cluster capacity. The split permits a lease and an HPC job to claim the same provider capacity without a common reservation.

This ADR decision was recorded before Task 84C implementation changes. Its evidence was rechecked after implementation against base revision `99973c3af6c084a915e2913016d5d08899dcdaa6`, current working-tree imports, the checked-in mainnet genesis, generated descriptors, and current clients.

## Evidence matrix

| Dimension | `x/market` | `mktplace` | Consequence |
| --- | --- | --- | --- |
| Module/store | Module `market`; binary prefixes `0x1100` orders, `0x1200` bids, `0x1300` leases and reverse indexes | Module/store `mktplace`; JSON prefixes `0x01` offerings, `0x02` orders, `0x03` allocations, `0x04` bids, plus indexes | Stores are independent and require explicit reconciliation. |
| Wire and type URLs | Generated packages `virtengine.market.v1`, `v1beta5`, and `v2beta1`; deployed messages include `/virtengine.market.v1beta5.MsgCreateBid` and `MsgCreateLease` | Generated package `virtengine.marketplace.v1`; messages include `/virtengine.marketplace.v1.MsgAcceptBid` and allocation actions | Namespaces do not collide, but semantics overlap. Existing market type URLs must remain stable. |
| Genesis/deployed evidence | Checked-in mainnet genesis contains `market`; representative state is empty | The same genesis contains `mktplace`; representative state is empty | No production record-count claim is possible from source alone. Migration must still handle non-empty testnet/operator state. |
| Msg/query/gateway/events | Complete order/bid/lease Msg and Query contracts, generated gateway, typed lifecycle events, CLI/provider integrations | Generated Msg service exposes offering and allocation mutations; generated Query exposes only price and allocation lists; custom marketplace events are JSON records | `x/market` has the complete public financial lifecycle. |
| Portal/SDK/provider clients | The provider daemon directly imports `sdk/go/node/market/v1beta5` for order queries and `MsgCreateBid`; SDK `MarketClient` imports v1/v1beta5 market contracts; transaction loading retains `/virtengine.market.v1beta5.MsgCreateBid` | Current direct references are generated SDK/gateway surfaces and compatibility tests; repository portal `/marketplace/*` paths are provider HTTP product APIs, not direct on-chain `mktplace` Msg imports | Replacing `x/market` would break the active provider/client path. Generated exposure alone is not proof of direct product use. |
| Escrow/financial semantics | Bid deposit account, lease payment, withdrawal, close hooks, deployment group links, settlement lineage | No escrow keeper; accepting a bid creates an internal allocation only | Financial ownership already resides in `x/market`. |
| Deployment semantics | Deployment creates market orders; lease IDs are the manifest/provider-daemon contract | Internal allocations drive Waldur lifecycle helpers but are not the deployment lease authority | Preserve Waldur/supply behavior as adapters, not a second sale. |
| VEID/MFA/provider/encryption | Order VEID gating exists; market handler validates registered provider and audited attributes; provider registration has VEID/MFA checks; deployment encryption remains unchanged | Rich offering-level VEID/MFA/provider policy and encrypted offering/order fields | Preserve offering catalog and policy settings; move executable gating to canonical transitions. |
| Capacity | No authoritative reservation | No authoritative reservation | Neither wins this dimension; `x/resources` must own reservations. |
| Migration history | Pre-Task-84C consensus version 7 and existing v1beta4-to-v1beta5 store migration | Pre-Task-84C consensus version 1; introduced later as a product/Waldur lifecycle | Promoting `mktplace` would discard mature migration and financial history. Task 84C advances these to 8 and 2 respectively. |
| Rollback and total cost | Add companion reservation links while retaining all IDs/routes/events | Requires conversion of every market order/bid/lease, escrow payment, deployment manifest and all clients | Keeping `x/market` is materially lower risk and cost. |

History also shows the `mktplace` tree originated in the January 2026 patent/product implementation and later received portal lifecycle controls. That history supports preserving its richer supply catalog, but not making it the financial owner.

## Decision

1. **`x/market` is the only mutable financial marketplace owner.** Its existing order, bid, lease, escrow, deployment, type URL, route, event, and store contracts remain canonical.
2. **`x/resources` is the only mutable capacity/reservation owner.** A versioned `Reservation` aggregate links request, inventory, canonical order/bid/lease, HPC job, escrow, collateral, consumer, legacy source, and deterministic lifecycle times/heights.
3. **`mktplace` becomes a supply-catalog and compatibility module.** Offering, provider-policy, MFA configuration, event checkpoint, and pricing reads remain available. After `v1.6.0`, all independent order/bid/allocation and allocation-lifecycle writes return the stable `marketplace lifecycle writes deprecated; use x/market and x/resources` error. Existing records remain queryable and are reconciled into reservations or quarantined.
4. **HPC is a reservation consumer, not a parallel capacity ledger.** A standalone job owns one `hpc_job` reservation. A market-backed job reuses its active canonical lease reservation after provider, customer, lease, and capacity checks; market remains the release/quarantine owner for that shared reservation.
5. **Cross-module execution is atomic.** Lease creation and HPC reserve/activate paths run inside one Cosmos cached context. Failure commits no market, escrow, HPC, or resources partial state.

## Reservation transition table

| From | Operation | To | Capacity returned? |
| --- | --- | --- | --- |
| none | reserve | Pending | no |
| Pending | activate/link | Active | no |
| Active | consume | Consumed | no |
| Pending/Active/Consumed | release | Released | yes, once |
| Pending | expire | Expired | yes, once |
| Pending/Active/Consumed | quarantine | Quarantined | no; operator/dispute hold |
| Pending/Active/Consumed/Quarantined | slash | Slashed | yes, once |

Terminal states never reactivate. Exact retries with the same idempotency key and payload return the original reservation. A different payload under the same key fails. One reservation has at most one active consumer.

Capacity is committed integer state. For each inventory and dimension:

`available + nonterminal reservation capacity = declared total`.

Consumed means capacity consumed by executable work, not destroyed inventory. It therefore remains in the nonterminal reserved side until release/slash.

## Eligibility policy

Reservation requires an active, fresh inventory and a currently registered provider. Mandatory configured eligibility readers fail closed when unavailable. Benchmark, collateral, and attestation references are recorded in lineage only when their committed-state profiles are enabled; no synthetic score, stake, attestation, or capacity is invented. Provider suspension/heartbeat expiry quarantines active capacity and expires pending capacity.

## Compatibility window

- **Before activation:** both historical modules decode and query as today; operators run the preflight report.
- **At `v1.6.0`:** cross-module reconciliation runs before the non-owner write fence; ordered module migrations then rebuild/activate resources, market, HPC, and `mktplace` state. Historical genesis with activation flags absent/false remains replayable and is fenced only by this upgrade. A new chain initialized from the current binary's default genesis starts directly in the current protocol with both flags true; explicit post-upgrade exports also preserve true flags. This ordering prevents activation from hiding mutable legacy records from reconciliation.
- **For one minor-release window:** all existing `virtengine.market.*` routes/type URLs remain unchanged; `virtengine.marketplace.v1` supply/price reads and legacy allocation reads remain. Deprecated writes return the stable error rather than disappearing.
- **After the window:** removing deprecated Msg routes requires a separate ADR and major compatibility decision. Historical records remain exportable.

## Migration and quarantine policy

- Existing terminal market financial records are preserved; no capacity is synthesized for them.
- Existing resource allocations become separate deterministic `legacy/allocation/<allocation-id>` reservation aggregates while retaining their original allocation IDs. An active/pending allocation is linked only when one inventory matches, assigned equals required, and `available + all linked legacy assignments = total`; otherwise it is zero-capacity quarantined.
- Existing active/pending market leases and nonterminal HPC jobs without complete authoritative capacity evidence are relinked to deterministic zero-capacity quarantine records rather than guessed into active reservations.
- Legacy non-owner orders, open bids, and allocations are independently quarantined. Historical records remain queryable through the compatibility facade.
- Orphans, duplicate consumers, conflicting providers, missing capacity, inconsistent terminal states, and arithmetic violations are quarantined. Migration never guesses and never creates capacity.
- The migration emits deterministic counts for scanned, linked, terminal-preserved, and quarantined records. Rollback after activation is unsupported because old binaries permit duplicate owners and ignore reservation indexes.

## Alternatives rejected

### Promote `mktplace`

Rejected because it would require migration of market, escrow, deployment, provider-daemon, portal, SDK and historical lease state while recreating mature financial semantics.

### Keep both with explicit boundaries

Rejected because two mutable order/bid authorities still allow disagreement and cross-store partial failure. Offerings are retained as catalog data, not a second financial lifecycle.

### Put reservations in market or HPC

Rejected because both product paths consume the same physical inventory. Capacity conservation needs one product-neutral owner.

## Consequences

- Canonical market compatibility is additive.
- `x/resources` becomes consensus-critical and gains migrations, invariants, indexed expiration, stable events and generated lineage queries. Reservation IDs are deterministic hashes of the canonical request payload and are not caller-controlled idempotency keys.
- Marketplace/HPC mutations gain stricter fail-closed behavior at activation.
- External live-chain inventory, collateral policy certification, and real provider/HPC deployment evidence remain release gates and cannot be inferred from local fixtures.