# Thread 5 Master Prompt: Platform Security

Paste the prompt below into a new VS Code thread rooted at the `virtengine-t5`
worktree.

```text
You are persistent implementation Thread T5 for the VirtEngine three-day
prototype campaign. Work continuously through this queue without asking for
another prompt. Commit and push each green checkpoint, update the handoff, and
immediately continue. Use subagents for MFA, vault, organizations, federation,
cloud profiles, privileged policy, fixtures, and adversarial review when paths
are disjoint.

Repository/worktree: ../virtengine-t5
Required branch: ve/prototype-t5-platform
Owned roadmap tasks: 89A, 89B, 89C, 89D, 90A, 91A
Prototype objective: versioned contracts, deterministic fixtures, and fail-closed
feature gates that let downstream work proceed safely. Since 88D and later
predecessors are incomplete, these features remain fixture_only or disabled and
must not be wired, advertised, or marked engineering-complete.

Mandatory rules:
- Read applicable AGENTS.md files, the DAG, and the five-thread plan first.
- Never push to stable or integration branches.
- Use typed state disabled -> fixture_only -> sandbox -> production. This queue
  may not transition beyond fixture_only.
- Missing 88D/dependency digests, profiles, replay state, custody, or policy
  always deny mutation and readiness.
- Never fall back to memory vaults, process-local keys/nonces, local organization
  authority, uncertified cloud profiles, or unilateral privileged control.
- Avoid shared startup, app wiring, generated protobufs, portal stores, and
  release manifests until T4 grants an integration window.
- Prefer isolated contract packages and golden fixtures.
- Commit green checkpoints with Campaign, Thread: T5, Checkpoint, and
  Handoff-From trailers. Maintain
  _docs/ralph/handoffs/prototype-t5/HANDOFF.yaml.

Queue:
T5-01 (4h): Define 89A factor profile, enrollment state, canonical challenge,
recovery policy, compromise hold, participant hook, and supersession event
contracts. Add wrong chain/account/action, replay, expiry, absent prior policy,
and unknown participant-version fixtures.
T5-02 (4h): Define 89B vault/KMS profile, consent decision, legal-hold authority,
wrapped-key metadata, restore manifest, rotation state, and erasure tombstone.
Add restart/restore/hold/erasure validators using backend-neutral interfaces;
never select the in-memory backend for production state.
T5-03 (4h): Define isolated 89C organization authority/privacy contracts:
identity, privacy mode, invitation/member lifecycle, threshold policy, delegated
budget, ownership, projection cursor, recovery hooks, and billing lineage. Add
last-admin, replay, budget-conservation, leakage, and unknown-hook tests. Do not
wire an organization keeper or portal mutation.
T5-04 (4h): Define 89D canonical service record, signed discovery document,
client trust profile, request sign bytes, route policy, key epochs, and atomic
nonce-consume interface. Add Go/TypeScript vectors for stale/rollback discovery,
cross-provider replay, origin policy, and unsupported trust profiles. No live
route integration.
T5-05 (4h): Define 90A cloud profile and backend-neutral desired-resource graph
fixtures with exact backend/version, immutable lineage, capabilities, quotas,
cost ceiling, idempotency, cleanup graph, and certification state. Include
Kubernetes plus one named VM/IaaS fixture row marked fixture_only. Reject
unsupported, conflicting, uncertified, and ownership-unsafe cleanup paths.
T5-06 (4h): Define 91A role/account lifecycle, policy registry extension, scoped
emergency action, threshold approval, MFA requirement, legal-hold migration, and
append-only audit contracts. Test stale policy, self-approval, missing action,
scope expansion, quorum loss, and audit-chain tampering. Do not activate routes.
T5-07 (2h): Run cross-contract compatibility and negative-enablement tests,
document dependency digests required from T1/T3/T4, and publish the green
fixture-only checkpoint.

Focused gates:
- go test -count=1 ./x/mfa/types -run 'Recovery|FactorProfile|Challenge'
- go test -count=1 ./pkg/data_vault/...
- go test -count=1 ./pkg/organization/contracts/... when created
- go test -count=1 ./pkg/provider_daemon/federation/... when created
- go test -count=1 ./pkg/provider_daemon/backendprofile/... when created
- go test -count=1 ./x/roles/types -run 'Policy|Role|Account|Audit'
- git diff --check
Run narrower tests continuously and applicable preflight before task-level
checkpoints.

Fallback queue:
- Expand canonical-byte, malformed-input, unknown-version, downgrade, replay,
  expiry, privacy, and transition-table tests for every T5 contract.
- Add feature-gate tests proving no fixture_only profile is registered,
  advertised, ready, or mutable.
- Produce cross-language vectors where portal/provider consumers will need them,
  without editing shared generated artifacts.
- Audit T5-owned current defaults for memory/process-local fallbacks and add
  fail-closed interfaces or tests in isolated packages.

When exhausted, audit only T5-owned contract/default surfaces, add numbered T5-X
checkpoints, and continue while dependency-independent work remains. Never claim
real account recovery, durable KMS custody, organization authority, provider
federation, cloud certification, privileged governance, GA, or production
readiness from these fixtures.
```
