# Authoritative Resource Reservation State Machine

## Ownership

- `x/market` owns mutable financial `Order -> Bid -> Lease` state.
- `x/resources` owns mutable inventory and reservation state.
- `x/hpc` consumes an active reservation before queued/running execution. A market-backed job reuses its active canonical lease reservation; a standalone HPC job owns a dedicated `hpc_job` reservation.
- `mktplace` remains a supply/offering catalog and legacy read facade. Its independent lifecycle writes return the stable deprecation error after `v1.6.0`.

The full architectural evidence and compatibility decision are in ADR-007.

## Capacity units

All dimensions are signed 64-bit integers validated nonnegative before arithmetic:

| Dimension | Unit |
| --- | --- |
| CPU | existing protocol CPU unit; providers must publish one consistent unit per inventory profile |
| Memory | GiB for resources/HPC v1 contracts |
| Storage | GiB for resources/HPC v1 contracts |
| Network | Mbit/s |
| GPU | device count, with optional exact type |

Market deployment resource values are converted with checked integer multiplication by replica count. No float or wall clock is used.

## Transitions

| Current | Operation | Next | Inventory effect |
| --- | --- | --- | --- |
| none | `Reserve` | Pending | subtract capacity |
| Pending | `ActivateReservation` | Active | none |
| Active | `ConsumeReservation` | Consumed | none |
| Pending/Active/Consumed | `ReleaseReservation` | Released | restore once |
| Pending | indexed expiry | Expired | restore once |
| Pending/Active/Consumed | `QuarantineReservation` | Quarantined | retain |
| Pending/Active/Consumed/Quarantined | `SlashReservation` | Slashed | restore once and record slash evidence |

`Released`, `Expired`, and `Slashed` are final. Exact repeated terminal operations return the original record. A terminal reservation never reactivates.

## Idempotency and atomicity

`Reserve` binds an idempotency key to a SHA-256 digest of the complete request. Exact retry returns the original reservation. Same key/different payload returns `ErrReservationConflict`.

The reservation ID is derived from that complete request digest. The caller's idempotency key is an index, not a store-key namespace; this avoids delimiter/prefix collisions in consensus keys.

Canonical lease creation and HPC submission execute reservation, financial/job, and activation writes in one Cosmos cached context. Any error discards the cache. Cosmos serial transaction execution means two final-unit requests cannot both observe the original inventory state.

HPC must not reserve the same market capacity twice. A supplied market reservation is accepted only when it is active and its provider, requester/customer, canonical lease, and capacity all match the job. The market lease remains the release/quarantine owner for shared reservations; standalone HPC terminal states release their dedicated reservation.

## Eligibility

Fresh active inventory and registered-provider state are mandatory. Provider heartbeat expiry expires pending reservations and quarantines active/consumed reservations. Optional benchmark, attestation, and collateral profiles fail closed when requested because no certified profile reader is enabled in Task 84C.

## Expiration

Pending reservations have ordered big-endian time and optional height indexes. EndBlock processes at most 1,000 due entries per block. It does not scan the reservation store.

## Invariants

For every inventory and dimension:

`available + pending + active + consumed + quarantined = total`.

Additional executable checks cover:

- no negative or overflowing dimension;
- one active consumer per reservation;
- terminal finality;
- active lease has active reservation and matching lease lineage;
- queued/running HPC job has active reservation and matching provider/customer plus job-or-canonical-lease lineage;
- after canonical activation, every retained nonterminal non-owner record has a deterministic zero-capacity quarantine aggregate and all non-owner lifecycle mutation paths (including Waldur callbacks and usage requests) are rejected.

## Upgrade ordering

`v1.6.0` performs cross-module reconciliation before activating the `mktplace` and legacy-resources write fences. Module migrations then run in explicit dependency order: `resources -> market -> hpc -> mktplace`, followed by the remaining modules with `auth` last. Historical genesis with both flags absent/false remains valid for startup and pre-upgrade replay. Current-binary default genesis starts new chains with both flags true, and post-upgrade exports preserve true flags.

## Events

`resource_reservation_transition` exposes only bounded identifiers, provider, consumer type, and from/to states. It never emits encrypted inputs, manifests, workload commands, or data references.