# Waldur Action to Chain Message Mapping (Portal)

Date: 2026-02-07
Scope: Map every Waldur action and API endpoint used by the portal/provider-daemon to its on-chain equivalent.

## Source of truth and flow

- Chain is the source of truth for marketplace state.
- Waldur is provider-operated control plane. The provider-daemon bridges chain events to Waldur APIs and reports back to chain.
- Off-chain actions that mutate infrastructure must be reflected on-chain via `MsgWaldurCallback` or `MsgReportJobStatus`.

## Verified chain message definitions (by proto)

| Module | Message / Query | Proto path | Notes |
| --- | --- | --- | --- |
| marketplace v1 | MsgCreateOffering, MsgUpdateOffering, MsgDeactivateOffering | sdk/proto/node/virtengine/marketplace/v1/tx.proto | Provider offering lifecycle. |
| marketplace v1 | MsgAcceptBid, MsgTerminateAllocation, MsgResizeAllocation, MsgPauseAllocation | sdk/proto/node/virtengine/marketplace/v1/tx.proto | Allocation lifecycle. |
| marketplace v1 | MsgWaldurCallback | sdk/proto/node/virtengine/marketplace/v1/tx.proto | Provider-daemon callback into chain. |
| marketplace v1 | QueryAllocationsByCustomer, QueryAllocationsByProvider, QueryOfferingPrice | sdk/proto/node/virtengine/marketplace/v1/query.proto | Allocation visibility + pricing. |
| hpc v1 | MsgRegisterCluster, MsgUpdateCluster, MsgDeregisterCluster | sdk/proto/node/virtengine/hpc/v1/tx.proto | Provider HPC cluster lifecycle. |
| hpc v1 | MsgCreateOffering, MsgUpdateOffering | sdk/proto/node/virtengine/hpc/v1/tx.proto | HPC offering lifecycle. |
| hpc v1 | MsgSubmitJob, MsgCancelJob, MsgReportJobStatus | sdk/proto/node/virtengine/hpc/v1/tx.proto | HPC job lifecycle. |
| hpc v1 | QueryOfferings, QueryJobs, QueryJobsByCustomer, QueryJobsByProvider | sdk/proto/node/virtengine/hpc/v1/query.proto | HPC market queries. |
| settlement v1 | MsgRecordUsage, MsgAcknowledgeUsage, MsgSettleOrder | sdk/proto/node/virtengine/settlement/v1/tx.proto | Usage and settlement. |
| deployment v1beta5 | MsgCreateDeployment | sdk/proto/node/virtengine/deployment/v1beta5/deploymentmsg.proto | Akash-style order creation (see gaps). |

## Waldur actions to chain operations

Legend: Direction refers to the canonical flow in decentralized mode.

### Marketplace offerings and orders

| Waldur action | Waldur API endpoint (client) | Chain message / query | Direction | Notes |
| --- | --- | --- | --- | --- |
| List offerings | GET /api/marketplace-public-offerings/ | QueryOfferings (HPC only) | Portal -> Chain | Non-HPC marketplace offerings do not have a chain list query yet. HPC uses `virtengine.hpc.v1.Query/Offerings`. |
| Get offering details | GET /api/marketplace-public-offerings/{uuid}/ | QueryOffering (HPC) | Portal -> Chain | Non-HPC offerings lack a public chain query. |
| Create offering | Provider UI action | MsgCreateOffering (marketplace v1 or hpc v1) | Provider -> Chain | Providers register offerings on-chain, then map to Waldur via offering map JSON. |
| Update offering | Provider UI action | MsgUpdateOffering (marketplace v1 or hpc v1) | Provider -> Chain | Price/spec changes occur on-chain. |
| Deactivate offering | Provider UI action | MsgDeactivateOffering (marketplace v1) | Provider -> Chain | Waldur offering state changes are downstream of chain. |
| Create order | POST /api/marketplace-orders/ | No direct MsgCreateOrder | Portal -> Chain | Order creation is not defined in marketplace v1. Akash-style flow uses `MsgCreateDeployment` which creates a market order; confirm final flow. |
| Approve order | POST /api/marketplace-orders/{uuid}/approve_by_provider/ | MsgAcceptBid (marketplace v1) | Provider -> Chain | Accepting a bid creates allocation on-chain. |
| Set backend ID | POST /api/marketplace-orders/{uuid}/set_backend_id/ | MsgWaldurCallback | Provider -> Chain | Provider-daemon reports Waldur IDs to chain after provisioning. |
| List orders | GET /api/marketplace-orders/ | QueryBids / orders (market module) | Portal -> Chain | Marketplace order query endpoints are not defined in marketplace v1. |
| Get order | GET /api/marketplace-orders/{uuid}/ | QueryBids / order detail (market module) | Portal -> Chain | Portal should use chain orders; mapping depends on final market module flow. |

### Allocation and resource lifecycle

| Waldur action | Waldur API endpoint (client) | Chain message / query | Direction | Notes |
| --- | --- | --- | --- | --- |
| Provision resource | Provider daemon action | MsgWaldurCallback (callback_type=provision_complete) | Waldur -> Chain | Provider daemon sends signed callback after provisioning. |
| Update resource status | Provider daemon action | MsgWaldurCallback (callback_type=status_update) | Waldur -> Chain | Status updates must be signed and replay-protected. |
| Terminate resource | POST /api/marketplace-resources/{uuid}/terminate/ | MsgTerminateAllocation | Portal -> Chain | Chain request triggers provider termination; provider reports termination via callback. |
| Pause resource | POST /api/marketplace-resources/{uuid}/suspend/ | MsgPauseAllocation | Portal -> Chain | Waldur suspend maps to allocation pause; provider reports actual state via callback. |
| Resume resource | POST /api/marketplace-resources/{uuid}/resume/ | Missing (no MsgResumeAllocation) | Portal -> Chain | Gap: resume action not in marketplace v1. |
| Resize resource | POST /api/marketplace-resources/{uuid}/resize/ | MsgResizeAllocation | Portal -> Chain | Provider-daemon performs resize and reports via callback. |
| Start/stop/restart resource | POST /api/marketplace-resources/{uuid}/start|stop|restart/ | No direct chain message | Portal -> Chain | Treat as allocation state change; currently only pause/terminate/resize supported on-chain. |
| Check resource state | GET /api/marketplace-resources/{uuid}/ | QueryAllocationsByCustomer/Provider | Portal -> Chain | Chain has allocation queries; no resource-level query yet. |
| List resources | GET /api/marketplace-resources/ | QueryAllocationsByCustomer/Provider | Portal -> Chain | Map list to allocations; resource detail may remain off-chain. |
| Operation status | GET /api/marketplace-resources/{uuid}/actions/{operation_id}/ | No direct chain query | Portal -> Chain | Off-chain operation status should be reflected via callbacks. |
| Terminated confirmation | Provider daemon action | MsgWaldurCallback (callback_type=terminated) | Waldur -> Chain | Explicit terminate callback. |

### Usage and settlement

| Waldur action | Waldur API endpoint (client) | Chain message / query | Direction | Notes |
| --- | --- | --- | --- | --- |
| Submit usage report | POST /api/marketplace-resources/{uuid}/submit_usage/ | MsgRecordUsage | Provider -> Chain | Use chain usage records as source of truth. |
| Create usage record | POST /api/marketplace-resources/{uuid}/usages/ | MsgRecordUsage | Provider -> Chain | Map to per-period records. |
| List usage records | GET /api/marketplace-resources/{uuid}/usages/ | (No query in settlement v1) | Portal -> Chain | Settlement module lacks usage query; portal must index from events. |
| Update resource limits | POST /api/marketplace-resources/{uuid}/update_limits/ | MsgResizeAllocation (if limits map to resources) | Portal -> Chain | Otherwise off-chain only. |
| Settle order | Waldur billing job | MsgSettleOrder | Provider -> Chain | Chain settlement finalizes payouts. |

### HPC / SLURM actions

| Waldur action | Waldur API endpoint (client) | Chain message / query | Direction | Notes |
| --- | --- | --- | --- | --- |
| Create SLURM allocation | POST /api/slurm-allocations/ (implicit) | MsgSubmitJob or allocation flow | Chain -> Waldur | Allocation is created by provider daemon during job/offer execution. |
| List SLURM allocations | GET /api/slurm-allocations/ | QueryJobs / QueryOfferings | Portal -> Chain | Chain does not expose SLURM allocation list directly. |
| Update SLURM allocation limits | POST /api/slurm-allocations/{uuid}/set_limits/ | MsgResizeAllocation (marketplace) | Portal -> Chain | Map limits changes to allocation resize where applicable. |
| Terminate SLURM allocation | DELETE /api/slurm-allocations/{uuid}/ | MsgTerminateAllocation | Portal -> Chain | Provider reports termination via callback. |
| Submit job | POST /api/slurm-jobs/ (implicit) | MsgSubmitJob | Customer -> Chain | Portal uses chain message; provider daemon submits to SLURM. |
| Cancel job | DELETE /api/slurm-jobs/{uuid}/ | MsgCancelJob | Customer -> Chain | Chain cancellation triggers SLURM cancel. |
| Job complete/status | GET /api/slurm-jobs/{uuid}/ | MsgReportJobStatus | Waldur -> Chain | Provider daemon reports status with signed payload. |

### Support desk actions

| Waldur action | Waldur API endpoint (client) | Chain message / query | Direction | Notes |
| --- | --- | --- | --- | --- |
| Create ticket | POST /api/support-issues/ | No on-chain support module | Portal -> Waldur | Gap: chain does not model support tickets. |
| Update ticket | PATCH /api/support-issues/{uuid}/ | No on-chain support module | Portal -> Waldur | Off-chain only. |
| Change ticket state | POST /api/support-issues/{uuid}/{action}/ | No on-chain support module | Portal -> Waldur | Off-chain only. |
| Add comment | POST /api/support-issues/{uuid}/comment/ | No on-chain support module | Portal -> Waldur | Off-chain only. |
| List tickets | GET /api/support-issues/ | No on-chain support module | Portal -> Waldur | Portal sync UI shows Waldur status. |
| Delete ticket | DELETE /api/support-issues/{uuid}/ | No on-chain support module | Portal -> Waldur | Off-chain only. |

### Cloud provider actions (OpenStack/AWS/Azure)

| Waldur action | Waldur API endpoint (client) | Chain message / query | Direction | Notes |
| --- | --- | --- | --- | --- |
| List instances | GET /api/openstack-instances/, /api/aws-instances/, /api/azure-virtualmachines/ | No chain equivalent | Portal -> Waldur | Infrastructure view is provider-local. |
| Get instance | GET /api/openstack-instances/{uuid}/, /api/aws-instances/{uuid}/, /api/azure-virtualmachines/{uuid}/ | No chain equivalent | Portal -> Waldur | Off-chain only. |
| Start/stop/restart instance | POST /api/openstack-instances/{uuid}/start|stop|restart/, etc. | No chain equivalent | Portal -> Waldur | Should be reflected on-chain via MsgWaldurCallback if it affects allocation state. |
| Destroy/unlink instance | DELETE /api/openstack-instances/{uuid}/, /api/aws-instances/{uuid}/, /api/azure-virtualmachines/{uuid}/ | MsgTerminateAllocation | Portal -> Chain | Termination should be requested on-chain, then executed off-chain. |
| List volumes | GET /api/openstack-volumes/, /api/aws-volumes/ | No chain equivalent | Portal -> Waldur | Off-chain only. |
| Destroy volume | DELETE /api/openstack-volumes/{uuid}/, /api/aws-volumes/{uuid}/ | No chain equivalent | Portal -> Waldur | Off-chain only. |
| List tenants | GET /api/openstack-tenants/ | No chain equivalent | Portal -> Waldur | Off-chain only. |
| List Azure sizes/locations | GET /api/azure-sizes/, /api/azure-locations/ | No chain equivalent | Portal -> Waldur | Off-chain only. |

### Object storage and artifacts

| Waldur action | Waldur API endpoint (client) | Chain message / query | Direction | Notes |
| --- | --- | --- | --- | --- |
| Upload object | PUT /api/object-storage/buckets/{bucket}/objects/{key} | No chain message | Portal -> Waldur | Store encrypted payloads; chain stores pointers (`encrypted_config_ref`, `encrypted_inputs_pointer`). |
| Stream upload | PUT /api/object-storage/buckets/{bucket}/objects/{key}/stream | No chain message | Portal -> Waldur | Same as above. |
| Download object | GET /api/object-storage/buckets/{bucket}/objects/{key} | No chain message | Portal -> Waldur | Off-chain retrieval. |
| Head object | HEAD /api/object-storage/buckets/{bucket}/objects/{key} | No chain message | Portal -> Waldur | Off-chain retrieval. |
| Delete object | DELETE /api/object-storage/buckets/{bucket}/objects/{key} | No chain message | Portal -> Waldur | Off-chain cleanup. |
| List objects | GET /api/object-storage/buckets/{bucket}/objects | No chain message | Portal -> Waldur | Off-chain only. |
| Quota | GET /api/object-storage/buckets/{bucket}/quota | No chain message | Portal -> Waldur | Off-chain only. |
| Health check | GET /api/object-storage/health | No chain message | Provider -> Waldur | Operational. |

## Gap analysis (missing or unclear on-chain coverage)

1. Marketplace order creation: no `MsgCreateOrder` in marketplace v1; clarify whether `MsgCreateDeployment` (deployment v1beta5) is the canonical order creation flow for portal.
2. Marketplace offering queries: non-HPC offerings lack a public query endpoint; only HPC `Query/Offerings` exists.
3. Allocation lifecycle: no `MsgResumeAllocation` or start/stop/restart equivalents.
4. Usage queries: settlement module has no gRPC query for usage records; portal must index events or add queries.
5. Support desk: no on-chain support ticket module or messages; Waldur remains off-chain only.
6. Resource detail queries: chain exposes allocation list but not resource metadata or backend IDs; requires callbacks + off-chain storage.

## Migration guide for Waldur-only users

1. Register provider on-chain: create provider account and publish offerings via `MsgCreateOffering` (marketplace or hpc).
2. Map offerings: add Waldur offering UUIDs to `config/waldur-offering-map.json` with on-chain offering IDs.
3. Run provider-daemon with Waldur enabled and chain submit flags (see `_docs/provider-daemon-waldur-integration.md`).
4. Switch order flow to chain-first: portal submits chain order/bid messages; provider-daemon creates/approves Waldur orders.
5. Provisioning callbacks: ensure Waldur callbacks are signed and sent as `MsgWaldurCallback` to update allocation state.
6. Usage reporting: provider-daemon submits usage to chain with `MsgRecordUsage`; settlement uses `MsgSettleOrder`.
7. Decommission Waldur-only actions: keep Waldur UI for infrastructure operations, but do not treat Waldur as source of truth.

## Related docs

- _docs/provider-daemon-waldur-integration.md
- docs/documentation/portal-chain-gap-analysis.md
- docs/usage-reporting-settlement.md
