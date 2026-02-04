# VE Portal On-Chain Gap Analysis (Portal Actions)

Date: 2026-02-04
Scope: Portal actions listed in VE-Portal request (offerings, bids, allocations, HPC jobs, VEID, MFA).

## Inventory: Messages (Tx)

| Portal action | Existing message | Module / proto | Status | Notes |
| --- | --- | --- | --- | --- |
| Provider creates offering | MsgCreateOffering | marketplace v1 `sdk/proto/node/virtengine/marketplace/v1/tx.proto` | ✅ Present | Marketplace offerings. Separate HPC offering messages also exist in `virtengine/hpc/v1` (for HPC module offerings). |
| Provider updates offering | MsgUpdateOffering | marketplace v1 `sdk/proto/node/virtengine/marketplace/v1/tx.proto` | ✅ Present | Marketplace offerings. |
| Provider deactivates offering | MsgDeactivateOffering | marketplace v1 `sdk/proto/node/virtengine/marketplace/v1/tx.proto` | ✅ Present | Marketplace offerings. |
| Customer accepts bid | MsgAcceptBid | marketplace v1 `sdk/proto/node/virtengine/marketplace/v1/tx.proto` | ✅ Present | Creates allocation. |
| Customer terminates allocation | MsgTerminateAllocation | marketplace v1 `sdk/proto/node/virtengine/marketplace/v1/tx.proto` | ✅ Present | Allocation lifecycle termination. |
| Customer resizes allocation | MsgResizeAllocation | marketplace v1 | ❌ Missing | No message or handler found. Stub added in proto (see below). |
| Customer pauses allocation | MsgPauseAllocation | marketplace v1 | ❌ Missing | No message or handler found. Stub added in proto (see below). |
| Customer submits HPC job | MsgSubmitJob | hpc v1 `sdk/proto/node/virtengine/hpc/v1/tx.proto` | ✅ Present (name mismatch) | Portal label is `MsgSubmitHPCJob`, but chain uses `MsgSubmitJob`. |
| Customer cancels HPC job | MsgCancelJob | hpc v1 `sdk/proto/node/virtengine/hpc/v1/tx.proto` | ✅ Present (name mismatch) | Portal label is `MsgCancelHPCJob`, but chain uses `MsgCancelJob`. |

## Inventory: Query Endpoints

| Portal query | Existing endpoint | Module / proto | Status | Notes |
| --- | --- | --- | --- | --- |
| QueryOfferings (with filters) | `Query/Offerings` | hpc v1 `sdk/proto/node/virtengine/hpc/v1/query.proto` | ✅ Present | Supports `active_only` + pagination. Additional filters (provider, region, resource type) are not in current request. |
| QueryBids (by order) | `Query/Bids` | market v1beta5 `sdk/proto/node/virtengine/market/v1beta5/query.proto` | ✅ Present | `BidFilters` supports owner + dseq/gseq/oseq + provider + state + bseq. Order-level filtering is possible via dseq/gseq/oseq. |
| QueryAllocations (by customer/provider) | N/A | marketplace v1 | ❌ Missing | Marketplace module does not expose allocation list queries. No gRPC query proto found. |
| QueryHPCJobs (by allocation) | N/A | hpc v1 | ❌ Missing | HPC job queries do not include allocation linkage; no query by allocation ID. |
| QueryVEIDStatus (by address) | `Query/IdentityStatus` | veid v1 `sdk/proto/node/virtengine/veid/v1/query.proto` | ✅ Present | `QueryIdentityStatusRequest.account_address`. |
| QueryMFAFactors (by address) | `Query/FactorEnrollments` | mfa v1 `sdk/proto/node/virtengine/mfa/v1/query.proto` | ✅ Present | Lists active enrollments per address. |

## Gaps Identified

1. Allocation lifecycle actions missing: `MsgResizeAllocation`, `MsgPauseAllocation`.
2. Allocation queries missing: no query to list allocations by customer/provider.
3. HPC job query by allocation missing: no allocation linkage in `QueryJobs*`.
4. Naming mismatch: portal labels `MsgSubmitHPCJob` / `MsgCancelHPCJob` vs chain `MsgSubmitJob` / `MsgCancelJob`.
5. Offerings filter depth: `QueryOfferings` is limited to `active_only` + pagination; portal may need richer filters (provider, region, resource type).

## Proto Stubs Added (Missing Messages)

- `MsgResizeAllocation` + `MsgResizeAllocationResponse` (marketplace v1)
- `MsgPauseAllocation` + `MsgPauseAllocationResponse` (marketplace v1)

Files updated:
- `sdk/proto/node/virtengine/marketplace/v1/tx.proto`

## Implementation Priority (Estimate)

P0 (Portal-blocking)
- Allocation actions (resize/pause): add keeper handlers, validation, state transitions, CLI/SDK wiring, events, and tests. Est: 5-8 dev days.
- Allocation query endpoints: add marketplace query service for allocations by customer/provider (pagination + filters) + index keys. Est: 4-6 dev days.

P1 (Portal data completeness)
- HPC jobs by allocation: add allocation linkage to job model + query + indexes + SDK update. Est: 4-7 dev days.
- Offerings filter expansion: extend QueryOfferingsRequest (provider, region, resource_type, price bounds). Est: 3-5 dev days.

P2 (DX consistency)
- Align portal naming with chain messages or add alias wrappers in portal SDK. Est: 1-2 dev days.

## Notes / Assumptions

- Marketplace module owns the allocation lifecycle (`MsgAcceptBid` creates allocation). Missing allocation actions should live in marketplace v1.
- HPC jobs currently link to offerings/clusters, not marketplace allocations.
- Proto stubs do not include keeper implementation or state transitions.
