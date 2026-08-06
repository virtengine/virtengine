# Provider Daemon Training: Waldur Track

## Audience
Provider operators who run VirtEngine offerings through a Waldur control plane.

## Training Outcome
At handoff, an operator must be able to:
- publish a Waldur offering with a stable `backend_id`
- verify provider and customer visibility of that offering
- approve and reconcile Waldur orders with VirtEngine identifiers
- operate resource lifecycle actions safely
- submit usage and confirm invoice-backed settlement evidence
- diagnose duplicate backend IDs, missing resources, and partial usage failures

## Lab 1: Provider Onboarding

### Goal
Create a provider-managed offering that the bridge can reconcile deterministically.

### Steps
1. Create a Waldur offering with:
   - provider customer UUID
   - offering type such as `VirtEngine.Compute`
   - `shared=true`
   - `billable=true`
   - a unique `backend_id`
2. Query the offering back with backend ID lookup.
3. Confirm the customer-facing offering list returns the same UUID.

### Pass Criteria
- one offering is returned for the chosen backend ID
- zero duplicate backend matches exist
- the offering is visible through the public-offerings list path

## Lab 2: Provider Fulfillment

### Goal
Drive the real provider-side order flow the bridge expects.

### Steps
1. Create a Waldur order for the published offering.
2. Approve the order as provider.
3. Set the order backend ID to the canonical VirtEngine order or allocation identifier.
4. Verify the order state and backend ID persisted.
5. Verify the resulting resource carries the same backend ID.

### Pass Criteria
- order state is provider-approved
- order backend ID equals the VirtEngine identifier
- the resource is queryable by UUID and returns a valid lifecycle state

## Lab 3: Lifecycle Operations

### Goal
Operate a reconciled resource safely.

### Steps
1. Stop the resource.
2. Start the resource.
3. Restart the resource.
4. Resize the resource.
5. Terminate the resource.
6. Attempt one invalid action against the terminated resource.

### Pass Criteria
- valid actions succeed
- terminated resources reject `start`
- lifecycle behavior is checked with resource UUIDs, not manual dashboard assumptions

## Lab 4: Usage and Settlement

### Goal
Produce an auditable settlement chain from usage to invoice.

### Steps
1. Build canonical line items for CPU and memory usage.
2. Convert them with `UsageReportFromLineItems`.
3. Submit the usage report to the resource.
4. Verify Waldur recorded the usage.
5. Create or verify the corresponding invoice.
6. Confirm the invoice reaches paid state.

### Pass Criteria
- component names are canonical and deterministic
- the usage report includes the authoritative `backend_id`
- the invoice total matches the submitted usage quantities and prices
- the invoice ends in `paid`

## Lab 5: Failure Drills

### Goal
Prove the operator can recover from the real bridge failure modes.

### Drill A: Duplicate backend ID
- Create or simulate two offerings with the same backend ID.
- Verify backend lookup returns a conflict instead of choosing one silently.

### Drill B: Partial bulk usage failure
- Submit bulk usage for one valid resource and one invalid resource UUID.
- Confirm the valid resource succeeds and the invalid resource is surfaced explicitly as failed.

### Drill C: Create-order recovery
- Force a Waldur-side create-order failure.
- Retry after clearing the upstream failure.
- Confirm the recovered order returns a real UUID.

## Validation Command

```bash
go test -tags e2e.integration ./tests/e2e/waldur/...
```

## Evidence to Capture
- offering UUID and backend ID
- order UUID and backend ID
- resource UUID and lifecycle state
- usage submission timestamp and component payload
- invoice UUID, total, and paid timestamp
- any explicit conflict or partial-failure error text from the drill labs

## Stop Rules
- Stop immediately if `GetOfferingByBackendID` returns more than one offering.
- Stop if an operator cannot prove which VirtEngine identifier is stored as the Waldur backend ID.
- Stop if usage is submitted without a resource UUID or canonical component types.
