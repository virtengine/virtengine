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
- Follow `_docs/prototype-thread-intake-runbook.md`. Before publication, merge
  the open frozen epoch base, rerun gates, commit `payload_head`, retain literal
  outputs, commit the handoff/evidence descendant, push with upstream, and push
  an annotated `checkpoint/prototype-t3/<id>` tag targeting the handoff commit.
  Record planning SHA, epoch/base, payload and prior accepted payload exactly.
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
tolerance boundary, partial, malformed, stale, and unavailable inputs. This is
T3-03A classifier groundwork until T3-04 durability and mutation gating pass.
T3-04 (5h): Add durable append-only reconciliation jobs/results, cursors,
evidence digests, attempts, and action intents using repository atomic storage
patterns. Prove restart at each phase, duplicate replay convergence, corrupt
store rejection, and race safety. Every non-matched state must durably preserve
the action intent and block usage/settlement mutation until authoritative
resolution; add a table-driven state/action matrix.
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

Supplemental AI operations queue:
T3-08 (4h): Define bounded operational metrics for every model stage and decision:
volume, unavailable/error, latency, freshness, reason-code counts and profile
digest. Prohibit account, document, biometric, nullifier and evidence IDs in
labels or logs.
T3-09 (4h): Add model-quality evidence schemas for OCR CER/WER, document
confusion, face FMR/FNMR/EER, PAD APCER/BPCER, fraud calibration and subgroup
limits. Synthetic fixtures validate computation but never certify production.
T3-10 (4h): Add uniqueness operations evidence: candidate rate, reviewed and
confirmed matches, false-match/false-non-match outcomes, atomic enrollment
conflicts, threshold-node availability, appeal backlog and subgroup monitoring.
All outputs are aggregate and anti-enumeration bounded.
T3-11 (4h): Add durable raw-evidence deletion and key-destruction job states,
retry/reconciliation metrics, legal-hold conflicts and backend/KMS receipt
validation. Consume T5's authoritative retention/deletion receipt; T3 owns only
operational jobs and projections. A claim without storage/KMS evidence is unresolved.
T3-12 (4h): Add model/profile drift, canary, pause and rollback telemetry contract
fixtures for T1/90C. Unknown profile or missing baseline is unavailable, never
healthy. Do not activate lifecycle mutations.
T3-13 (3h): Add appeal and manual-review queue SLIs: age, independent reviewer,
recusal, reason delivery, restoration and subgroup overturn rates. Keep identity
details out of metrics.

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
- Add privacy-safe model incident, uniqueness-service outage, mass false-match,
  deletion backlog and key-compromise alert fixtures for T4 integration.

When exhausted, audit only T3-owned state machines and producer call sites for
remaining prototype-critical fail-open or non-durable behavior, number new T3-X
checkpoints, and continue. Never claim two-chain certification, production
reconciliation authority, SLO completion, or 90B certification.
```
