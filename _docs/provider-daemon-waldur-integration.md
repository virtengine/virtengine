# Provider Daemon Waldur Integration

## Purpose
This document is the operator-facing source of truth for the VirtEngine to Waldur bridge implemented in `pkg/waldur/**` and consumed by provider-side workflows. It describes the exact onboarding, order, lifecycle, usage, settlement, and recovery semantics that the automated Waldur E2E package now verifies.

## Owned Code Paths
- `pkg/waldur/client.go`
- `pkg/waldur/marketplace.go`
- `pkg/waldur/lifecycle.go`
- `pkg/waldur/usage.go`
- `pkg/waldur/usage_line_items.go`
- `cmd/virtengine/cmd/waldur/**`
- `tests/e2e/waldur/**`

## Bridge Contract

### Authentication and API base path
- `BaseURL` may point at the Waldur site root or an API root.
- The client negotiates `BaseURL`, `/api/`, and `/api/v1/` automatically before first use.
- `Token` may already include `Token ` or `Bearer `.
- `AuthScheme` is supported when the raw secret must be emitted as `Token <secret>` or `Bearer <secret>`.
- Connection probes and `GET /users/me/` are the readiness gate for the client.

### Provider offering identity
- Every provider offering published to Waldur must carry a stable `backend_id`.
- `backend_id` is the authoritative bridge key for reconciliation.
- The expected value is the VirtEngine provider-side offering identifier, for example `ve1providerabcd/compute-001`.
- `GetOfferingByBackendID` fails closed:
  - `ErrNotFound` when no offering matches.
  - `ErrConflict` when more than one offering shares the same backend ID.

### Order identity
- VirtEngine order or allocation identifiers must be copied into Waldur order metadata.
- The production bridge uses:
  - `ve_order_id`
  - `ve_customer_address`
  - `ve_provider_address`
  - region or placement attributes when relevant
- After order creation, the provider must approve the order and set the Waldur backend ID to the canonical VirtEngine order or allocation identifier.

## Provider Onboarding

### 1. Create the provider offering in Waldur
- Use `MarketplaceClient.CreateOffering`.
- Required inputs:
  - `name`
  - `customer_uuid`
  - `type`
  - `shared`
  - `billable`
- Recommended inputs:
  - `backend_id`
  - `description`
  - `components`
  - `attributes`

### 2. Verify backend reconciliation
- Immediately query the offering back with `GetOfferingByBackendID`.
- Treat duplicate matches as a release blocker, not an operator warning.

### 3. Validate customer visibility
- Use `ListOfferings` against `marketplace-public-offerings`.
- Confirm the returned offering UUID matches the provider offering UUID.

## Order and Resource Lifecycle

### Customer order path
1. Customer selects a public offering.
2. VirtEngine creates a corresponding Waldur order with provider and allocation metadata.
3. Provider approves the order with `ApproveOrderByProvider`.
4. Provider sets the authoritative VirtEngine identifier with `SetOrderBackendID`.
5. Waldur resource provisioning proceeds and the resulting resource carries:
   - `offering_uuid`
   - `project_uuid`
   - `backend_id`
   - provider-owned attributes such as IP addresses or adapter metadata

### Resource control path
- Resource lifecycle actions are executed through `LifecycleClient`:
  - `Start`
  - `Stop`
  - `Restart`
  - `Resize`
  - `Terminate`
- `ValidateLifecycleAction` is the preflight gate.
- Invalid transitions, such as starting a terminated resource, are treated as hard errors.

### Expected resource states
- `Creating`
- `OK`
- `Stopped`
- `Paused`
- `Updating`
- `Terminating`
- `Terminated`
- `Erred`

## Usage, Settlement, and Reconciliation

### Usage submission
- Canonical usage should be derived from line items with `UsageReportFromLineItems` when the upstream system already has normalized cost inputs.
- Direct reports may also be submitted with `SubmitUsageReport`.
- Supported component names in the bridge path are:
  - `cpu_hours`
  - `gpu_hours`
  - `ram_gb_hours`
  - `storage_gb_hours`
  - `network_gb`

### Required usage metadata
- `backend_id`
- provider identifier
- allocation or order identifiers when available
- optional currency or audit labels

### Bulk usage behavior
- `SubmitBulkUsage` is intentionally fail-open per resource and fail-closed per batch result.
- Successful resource submissions are preserved.
- Failed resources are returned in the error string and in per-resource response state with `state=failed`.

### Settlement evidence
- A valid settlement trail contains:
  - Waldur offering UUID
  - Waldur order UUID
  - Waldur resource UUID
  - VirtEngine backend ID
  - submitted usage payload or canonical line items
  - resulting invoice UUID and paid state when settlement is complete

## Failure Handling

### Hard failures
- `ErrUnauthorized`: invalid token or wrong auth scheme.
- `ErrForbidden`: token lacks provider permissions.
- `ErrNotFound`: missing offering, order, or resource.
- `ErrConflict`: duplicate backend ID reconciliation drift.
- `ErrServerError`: Waldur-side 5xx failure.

### Validation failures
- Missing resource UUID.
- Empty usage component list.
- Negative, `NaN`, or infinite usage amounts.
- usage period end before usage period start.

### Retry expectations
- Transport timeouts and 5xx responses are retryable inside the client.
- 401, 403, 404, and conflict conditions are not retried automatically.
- Operator recovery is:
  1. fix the external state
  2. rerun the exact bridge action
  3. confirm reconciliation by backend ID, not by human-readable name

## Operator Verification

### Automated proof
Run:

```bash
go test -tags e2e.integration ./tests/e2e/waldur/...
```

The suite proves:
- customer onboarding and auto-provisioned order flow
- provider offering update and backend reconciliation
- provider approval and backend ID persistence
- lifecycle actions on verified resources
- canonical usage-to-invoice settlement flow
- duplicate backend conflict rejection
- partial bulk usage failure reporting
- recovery after a Waldur order creation failure

### Manual spot checks
1. Create a provider offering with a unique `backend_id`.
2. Query it back by backend ID.
3. Create a Waldur order for the offering.
4. Approve the order and set its backend ID to the VirtEngine allocation or order ID.
5. Submit usage for the resulting resource.
6. Verify invoice creation and paid-state transition in Waldur.

## Release Gate
The Waldur bridge is not launch-ready unless all of the following are true:
- every production offering has a unique `backend_id`
- client auth and base URL settings resolve without manual path rewrites
- provider order approval and backend reconciliation succeed
- lifecycle actions operate only on verified resource UUIDs
- usage batches surface partial failures explicitly
- operator docs and training material match the passing E2E flow in `tests/e2e/waldur/**`
