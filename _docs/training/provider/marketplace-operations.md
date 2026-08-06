# Marketplace Operations: Waldur Provider Path

## Scope
This module covers day-2 operations for providers whose marketplace fulfillment path runs through Waldur.

## Daily Checklist
- verify every active offering still resolves uniquely by `backend_id`
- verify recent orders have a persisted backend ID after provider approval
- verify active resources report a valid Waldur lifecycle state
- verify new usage reports were accepted for billable resources
- verify no invoice expected for settlement remains stuck before `paid`

## Offering Operations

### Publish
- create provider offering
- assign unique `backend_id`
- verify customer visibility through the public-offerings list

### Update
- update description, pricing components, and relevant attributes
- re-query by backend ID after every change

### Deprecate
- stop routing new customer demand to the offering
- keep historical backend ID evidence for existing orders and invoices

## Order Operations

### Expected provider actions
1. create or receive a Waldur order for the chosen offering
2. approve the order as provider
3. set backend ID to the VirtEngine order or allocation identifier
4. confirm the resource, once provisioned, carries the same backend ID

### Escalate when
- the order cannot be approved
- the backend ID cannot be persisted
- the resource UUID exists but the resource state cannot be queried

## Settlement Operations

### Required fields for every billable report
- resource UUID
- period start
- period end
- canonical component list
- authoritative `backend_id`

### Bulk usage rule
- partial success is acceptable only if the failed resources are surfaced explicitly and reconciled before closeout
- silent dropping of failed resources is not acceptable

### Invoice review
- compare invoice total against submitted usage quantities and offering prices
- confirm invoice state reaches `paid`
- archive offering UUID, order UUID, resource UUID, invoice UUID, and backend ID together

## Failure Playbooks

### Duplicate backend ID
- symptom: backend lookup returns conflict
- action: stop publication and remove the duplicate before continuing

### Missing resource
- symptom: usage or lifecycle action returns not found
- action: reconcile the resource UUID from the approved order before retrying

### Waldur 5xx on order creation
- symptom: create-order returns server error
- action: confirm upstream health, clear the failure source, rerun create-order, verify a new order UUID is returned

### Partial bulk usage failure
- symptom: one or more resources return `state=failed`
- action: capture the failed resource UUIDs, resubmit only after the resource mapping issue is corrected

## Verification Command

```bash
go test -tags e2e.integration ./tests/e2e/waldur/...
```

## Evidence Bundle
- screenshot or API output showing offering UUID plus backend ID
- order approval evidence
- order backend ID evidence
- resource UUID and lifecycle state evidence
- usage submission evidence
- invoice paid-state evidence
