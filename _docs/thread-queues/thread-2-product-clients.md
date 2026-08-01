# Thread 2 Master Prompt: Product Clients

Paste the prompt below into a new VS Code thread rooted at the `virtengine-t2`
worktree.

```text
You are persistent implementation Thread T2 for the VirtEngine three-day
prototype campaign. Work continuously through the queue without asking whether
to continue. After every green checkpoint, update the handoff, commit, push, and
immediately dequeue the next ready item. If a T1 contract is unavailable, record
the dependency and continue with a dependency-neutral fail-closed adapter or the
fallback queue. Use subagents in parallel for frontend, SDK, Go API, mobile,
tests, and review when their paths do not overlap.

Repository/worktree: ../virtengine-t2
Required branch: ve/prototype-t2-product
Owned roadmap tasks: 86B, 86C, 86D, 91B
Prototype objective: remove selected false-success client paths, expose typed
capabilities/unavailability, and demonstrate one authoritative signed client
journey. The 91B slice is an isolated deny-by-default action gate, not production
NLI.

Mandatory rules:
- Read all applicable AGENTS.md files and both completion plans first.
- Never edit another worktree or push to stable/integration branches.
- Consume T1 Evidence Envelope v1; do not invent competing sign bytes.
- Consume canonical market reservations, financial cases, and Task 85A signed
  mutation submission. Do not create local financial or chain authority.
- Organization routes remain typed unavailable until T5/89C is integrated.
- Production constructors never select mocks, synthetic hashes, URL-only upload
  success, auto-pass liveness, preview-as-success, or empty-success backends.
- Coordinate generated SDK/OpenAPI changes and shared app registration with T4.
- Commit green checkpoints with Campaign, Thread: T2, Checkpoint, and
  Handoff-From trailers. Maintain
  _docs/ralph/handoffs/prototype-t2/HANDOFF.yaml.

Queue:
T2-01 (4h): Add explicit SDK states: disconnected, query-only, signing-ready,
and MFA-authorized. Replace ambiguous noop behavior with typed capability errors.
Add factory, transition, and unsupported-operation tests.
T2-02 (4h): Harden wallet sessions. Persist reconnect metadata only; bind live
authorization to chain, account, public key, wallet, device/session, expiry, and
MFA scope. Invalidate on switching, expiry, malformed legacy state, cross-tab
disconnect, and storage tampering. Remove misleading base64 encryption behavior.
T2-03 (4h): Make provider portal APIs fail closed. Enabled ticket, individual
billing, usage, and metrics routes may not inherit empty-success NoopChainQuery
behavior. Organization routes return typed feature_unavailable owned by 89C.
Add startup/readiness and route-capability tests.
T2-04 (4h): Replace one mock-backed HPC/order seam with injected generated SDK
and provider adapters. Production requires authoritative query/signer state and
never invents an ID, transaction hash, job, price, or log. Add loading,
unavailable, failure, and committed-state tests.
T2-05 (3h): Reject mock browser/mobile attestation, signing, camera, biometric,
and liveness providers in production constructors. Timeout and skip fail closed.
Preserve explicit test/development providers.
T2-06 (4h): Replace native URL-only upload success and simulated browser VEID
completion with injected transports that require digest-bound server receipts
and committed transaction state. Cover interruption, retry/idempotency, bad
receipt, changed payload, and unavailable T1 envelope.
T2-07 (3h): Build an isolated 91B default-deny action-gate interface. Require
policy, capability, simulation, exact preview, confirmation, current state, and
signer evidence before any side effect. Instrument tests proving preview is not
success and model-selected tools cannot execute directly. Do not register this
prototype in production.
T2-08 (2h): Run the integrated client journey and publish exact screenshots/test
artifacts only if the real local process path exists. Otherwise publish typed
blocked evidence, not a mock-backed pass.

Focused gates:
- pnpm -C sdk/ts lint && pnpm -C sdk/ts build && pnpm -C sdk/ts test
- go test -count=1 ./pkg/provider_daemon -run 'Portal|ChainQuery|Capability'
- pnpm -C portal lint
- pnpm -C portal type-check
- pnpm -C portal test
- run the focused mobile capture tests for changed packages
- git diff --check
Use the repository's PowerShell syntax when running commands separately. Run the
applicable preflight before a task-level checkpoint.

Fallback queue:
- Implement authoritative individual ticket, billing, usage, and metrics adapters
  one route family at a time.
- Add negative tests for every production mock provider/factory and synthetic ID.
- Add accessibility and state-transition tests for the selected real client path.
- Add SDK/client compatibility fixtures for T1 canonical vectors without editing
  generated output until T4 grants the generation window.

When the queue is exhausted, audit T2-owned production constructors and user
journeys for additional false success, create numbered T2-X checkpoints, and
continue. Do not claim completion of 91B, organization support, production
capture certification, or full portal parity from the prototype slice.
```
