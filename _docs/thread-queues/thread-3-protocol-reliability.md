# Thread 3 Master Prompt: Protocol Reliability

Paste the prompt below into a new VS Code thread rooted at the `virtengine-t3`
worktree.

```text
You are persistent implementation Thread T3 for the VirtEngine three-day
prototype campaign. Continue through all ready queue items without asking for a
new prompt. Commit and push every green checkpoint, update the handoff, then
immediately start the next. Use subagents for IBC, reconciliation, observability,
benchmark contracts, race tests, and security review when ownership is disjoint.

Repository/worktree: ../virtengine-t3
Required branch: ve/prototype-t3-reliability
Owned roadmap tasks: 87B, 87C, 87D, 90B
Prototype objective: replay-safe IBC terminal transitions, durable and truthful
reconciliation, real producer telemetry for those paths, and an inactive signed
benchmark envelope. No external relayer or benchmark certification is claimed.

Mandatory rules:
- Read applicable AGENTS.md files and both plans before edits.
- Never push to stable or integration branches.
- Use integer/fixed-point canonical units; preserve conservation.
- Unavailable independent evidence is unavailable, never synchronized.
- Do not own identity keys, dispute policy, portal views, deployment topology,
  dashboards, alert routing, rewards, suspension, or benchmark activation.
- Preserve provider metrics port 9090.
- Avoid generated protobuf/OpenAPI/SDK changes in this queue unless T4 grants a
  single-owner window.
- Commit green checkpoints with Campaign, Thread: T3, Checkpoint, and
  Handoff-From trailers. Maintain
  _docs/ralph/handoffs/prototype-t3/HANDOFF.yaml.

Queue:
T3-01 (3h): Define versioned IBC transfer identity, pending/terminal states,
compensation reasons, and terminal markers. Add RED cases for duplicate,
conflicting, late, reordered, and exact-retry acknowledgments/timeouts.
T3-02 (5h): Implement atomic compare-and-transition acknowledgment/error/timeout
handling. Retain recoverable state until custody/refund action succeeds. Require
accounting/audit hooks. Add race and conservation tests.
T3-03 (3h): Replace reconciliation InSync booleans with matched, mismatched,
unavailable, stale, and unresolved states plus stable reason codes. Cover exact,
tolerance boundary, partial, malformed, stale, and unavailable inputs.
T3-04 (5h): Add durable append-only reconciliation jobs/results, cursors,
evidence digests, attempts, and action intents using repository atomic storage
patterns. Prove restart at each phase, duplicate replay convergence, corrupt
store rejection, and race safety.
T3-05 (4h): Connect bounded production metrics for T3 paths: pending exposure,
terminal outcomes, compensation failures, reconciliation state/freshness,
backlog, and action intent. Add duplicate-registration and bounded-label tests.
Correct observability compose mount mismatches without changing topology.
T3-06 (4h): Add an inactive canonical 90B benchmark/reliability envelope binding
provider, cluster, hardware, suite/image digest, runner/key epoch, timestamp,
nonce, and source freshness. Missing or stale source can never yield verified or
perfect reliability. Add golden, tamper, replay, and fixed-point tests.
T3-07 (2h): Run restart/replay/conservation/telemetry integration, document exact
handoffs to T4 and prerequisites from T5/90A, then publish the green checkpoint.

Focused gates:
- go test -race -count=1 ./x/settlement/ibc
- go test -race -count=1 ./pkg/provider_daemon -run 'Waldur|Reconcile|UsageReporter'
- go test -count=1 ./pkg/observability/...
- go test -count=1 ./pkg/benchmark_daemon ./x/benchmark/types
- docker compose -f docker-compose.observability.yaml config when Docker exists
- git diff --check
Run applicable preflight before publishing a task-level checkpoint.

Fallback queue:
- Expand IBC state-machine fuzz/property tests and exposure conservation cases.
- Add reconciliation store corruption, crash boundary, idempotency, and bounded
  retention tests.
- Add metric conformance tests proving each state transition changes exactly the
  expected series without high-cardinality labels.
- Inventory and replace T3-owned hard-coded benchmark reliability inputs with
  fail-closed source interfaces, without activating consumers.

When exhausted, audit only T3-owned state machines and producer call sites for
remaining prototype-critical fail-open or non-durable behavior, number new T3-X
checkpoints, and continue. Never claim two-chain certification, production
reconciliation authority, SLO completion, or 90B certification.
```
