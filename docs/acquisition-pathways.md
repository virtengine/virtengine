# Acquisition Pathways

> Canonical, always-current version of this page lives on the docs site:
> <https://docs.virtengine.com/concepts/acquisition-pathways/>. This repository
> copy is a concise summary for readers working from the source tree.

VirtEngine answers two separate questions:

1. **How does demand meet supply?** — the *commercial path*.
2. **Where does the work run?** — the *fulfilment backend*.

Bidding is a commercial path, not a backend; Kubernetes and SLURM are
backends, not ways to buy.

## Two ways to acquire

| Style | How it works | Resolution |
| --- | --- | --- |
| **Browse and buy (direct)** | Choose a public listing and purchase it at its published price | Immediate |
| **Open an order (bidding)** | Publish requirements and a price cap; providers bid | At the bidding deadline |

Both styles resolve to the same on-chain objects — an **order**, a **lease**,
and a capacity **reservation** — so escrow, usage, and settlement behave
identically.

A listing is a **standing ask**; a provider bid is a **standing bid**. The
deterministic resolution engine ranks both with the same rule: price, then
capacity fit, then provider reputation, then source and sequence.

## Provider supply options

| Option | Description |
| --- | --- |
| Native listing | Fixed price and capacity published in the protocol catalogue |
| Waldur listing | Offering published in Waldur and mirrored on-chain as a signed snapshot |
| Positional bid | Priced quote against a specific open order |
| HPC offering | Priced queue or plan on a registered cluster |

## Fulfilment backends

| Backend | Typical listing |
| --- | --- |
| Kubernetes | Containers and applications |
| OpenStack / VMware / AWS / Azure | Virtual machines, volumes, networking |
| SLURM / MOAB / Open OnDemand | HPC jobs and batch workloads |

## Where Waldur fits

The chain owns the deterministic commercial state (catalogue, orders, bids,
leases, reservations, escrow). Waldur owns the off-chain operational plane
(offering discovery, provider onboarding, service accounting, admin UI) and
provider infrastructure execution. Waldur offerings enter consensus as signed,
replay-protected snapshots; Waldur-backed purchases emit durable commands
executed off-chain.

In short: **marketplace semantics run on validators as protocol code; the
Waldur service runs at the edge.**

## Authority model

| Concern | Owner |
| --- | --- |
| Catalogue, orders, bids, leases, reservations, escrow | Chain (deterministic) |
| Offering discovery, provider onboarding, admin UI, service accounting | Waldur (off-chain) |
| Provisioning execution (VM, cloud, HPC) | Provider daemon |
| Metering reconciliation | Provider daemon vs. Waldur |

## Related documentation

- `_docs/architecture/unified-market-waldur-design.md` — the design specification.
- `_docs/adr/ADR-010-unified-market-resolution-and-waldur-supply.md` — the decision record.
- `_docs/adr/ADR-007-canonical-market-reservations.md` — canonical market and capacity ownership.
- `_docs/architecture/waldur-market-mapping-spec.md` — field-level Waldur mapping.
- `_docs/hpc-marketplace-e2e-test-guide.md` — HPC marketplace end-to-end flow.
