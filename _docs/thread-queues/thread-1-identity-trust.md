# Thread 1 Master Prompt: Identity Trust

Paste the prompt below into a new VS Code thread rooted at the `virtengine-t1`
worktree.

```text
You are persistent implementation Thread T1 for the VirtEngine three-day
prototype campaign. Work continuously through this queue. Do not stop after one
checkpoint and do not ask whether to continue. After each green commit, update
the handoff ledger, push the thread branch, and immediately start the next
highest-priority dependency-satisfied checkpoint. If blocked, record the exact
blocker and pull the first fallback item. You may launch subagents in parallel
for research, tests, security review, and non-overlapping implementation, but you
own their integration and verification.

Repository/worktree: ../virtengine-t1
Required branch: ve/prototype-t1-identity
Baseline: the campaign baseline recorded in
_docs/five-agent-protocol-completion-plan.md
Owned roadmap tasks: 85D, 87A, 90C, 90D
Prototype objective: authenticated VEID evidence plus a deterministic signed
inference-receipt seam, with future TEE/model/credential contracts that remain
fail closed until their dependencies exist.

Mandatory rules:
- Read AGENTS.md, applicable nested AGENTS.md files, the continuation plan, and
  the five-thread plan before editing.
- Never edit another thread's worktree or branch. Never push to
  stable-virtengine-beta or ve/prototype-integration.
- Keep Task 84A consensus callbacks free of inference and network I/O.
- Do not implement Task 90C lifecycle mutation inside 85D. Consume only the
  exact runtime-eligible profile.
- Do not extend keeper-held credential private-key issuance as the 90D design.
- Coordinate generated protobuf/OpenAPI/SDK files and app wiring through T4.
- Commit only green checkpoints using Conventional Commits and trailers:
  Campaign: three-day-prototype; Thread: T1; Checkpoint: <ID>;
  Handoff-From: <SHA>.
- Maintain _docs/ralph/handoffs/prototype-t1/HANDOFF.yaml with commands, exit
  codes, hashes, changed paths, blockers, and next checkpoint.

Queue, in order when dependencies permit:
T1-01 (2h): Inventory every VEID attestation/evidence class, handler, signer,
replay key, mutation point, score contribution, and credential consumer. Add a
machine-readable or table-driven trust inventory. Unknown production evidence
must be classified fail closed.
T1-02 (4h): Freeze Evidence Envelope v1 using the existing canonical web evidence
contract. Cover domain/version, chain/account/scope, issuer epoch, account
signature, payload digest, nonce/challenge, block bounds, privacy-safe storage,
and exact retry. Add golden and tamper vectors.
T1-03 (3h): Freeze Runtime Policy Reader v1 over VEID pipeline/profile and
veidregistry state. Add bootstrap-active, disabled, unknown, mismatched, and
superseded compatibility fixtures.
T1-04 (4h): Complete one vertical evidence path: SSO plus one non-web evidence
class through issuer resolution, account authorization, replay consumption,
mutation, and score lineage. Cover revoked keys, wrong account/chain/scope,
changed payload, expiry, replay, and exact retry.
T1-05 (3h): Define Workload Binding v1 for 87A: chain, workload identity, signer
key, model/runtime digest, nonce, profile, activation/expiry, collateral
reference, and debug-mode rejection. Provide verifier adapter tests only; do not
claim real-hardware certification.
T1-06 (4h): Add the deterministic signed-receipt producer interface and in-process
sidecar vector. Bind every receipt field expected by keeper verification. Tamper
each field and prove rejection. Do not add network access to consensus.
T1-07 (3h): Add 90C lifecycle compatibility fixtures. The 85D reader accepts only
the exact runtime-eligible state and rejects stale or unknown states. Do not add
promotion, canary, pause, rollback, training, or connector workflows.
T1-08 (3h): Define a read-only 90D credential-decision envelope containing only
subject/key epoch, accepted evidence and receipt digests, policy/score epochs,
consent-purpose reference, expiry, and status. Reject fabricated age/residency
and raw private keys/plaintext claims.
T1-09 (2h): Run the integrated T1 contract gate, freeze canonical vectors, reserve
store/protobuf identifiers, document handoffs to T2/T4/T5, and publish the green
checkpoint SHA.

Supplemental AI and biometric queue:
T1-10 (4h): Freeze one canonical feature schema shared byte-for-byte by Python
training and Go serving. Repair shifted/missing face/document fields and add
cross-language golden vectors. Silent zero-padding or extraction failure is an
error in production profiles.
T1-11 (3h): Add a governed model-portfolio manifest for document routing, OCR,
authenticity, face detection/comparison, liveness, fraud fusion and policy
evaluation. Every stage binds artifact, preprocessing, schema, license,
evaluation and runtime digests; unknown or placeholder stages fail closed.
T1-12 (4h): Mark the untrained GAN, anomaly, U-Net fallback, synthetic age and
hash-derived face/liveness implementations non-production. Add constructor and
policy tests proving no production profile can select them.
T1-13 (5h): Define separate fixed-point Document, FaceMatch, Liveness, Risk,
Identity, Uniqueness and Eligibility decisions. Replace aggregate-score pass
semantics with hard gates and stable reason codes. A high average cannot override
failed liveness, duplicate biometrics, expired evidence or active holds.
T1-14 (5h): Define versioned UniquenessRequest and threshold-signed
UniquenessReceipt contracts with final program-scoped nullifier,
possible-match-review, confirmed duplicate, unavailable, appeal and supersession
states. Pending/non-unique outcomes receive request-scoped opaque references and
never a stable nullifier. Retire truncated SHA-256 LSH from production policy.
T1-15 (4h): Build a deterministic uniqueness-service simulator for contract,
concurrency and policy tests. Prove atomic search-plus-insert, exact retry,
simultaneous duplicate enrollment, threshold/profile mismatch and no biometric
template in chain-facing receipts. Label it non-cryptographic fixture_only.
T1-16 (4h): Define PolicyBundle v1 and an action-specific eligibility evaluator.
Initial mint eligibility requires current identity, liveness and uniqueness,
good account standing and no risk hold. Add complete dependency/reason lineage.
T1-17 (4h): Bind credential and privacy proofs to governed issuer credential,
subject key, policy, verifier challenge, expiry, status and non-reusable
nullifier. Add transferred-proof and self-constructed biometric-proof negatives.
T1-18 (5h): Implement the authenticated worker boundary that builds signed stage
and aggregate receipts from the pinned model graph. Keep mTLS/workload identity,
model/runtime/schema pinning, deterministic result normalization and fail-closed
outage behavior explicit.

Focused gates:
- go test -count=1 ./x/veid/types ./x/veid/keeper ./x/veidregistry/...
- go test -race -count=1 ./x/veid/keeper ./pkg/inference ./cmd/inference-sidecar
- go test -count=1 ./pkg/enclave_runtime ./x/enclave/keeper -run 'Attestation|Heartbeat'
- go test -count=1 ./app -run 'VoteExtension|Proposal|Inference'
- python .github/tests/test_inference_deployment_policy.py
- git diff --check
Run narrower tests after every edit and the applicable repository preflight
before publishing a task-level checkpoint.

Fallback queue:
- Convert remaining structural-only evidence classes to table-driven negative
  tests without changing shared generated contracts.
- Add replay-domain collision, signer rotation, block-boundary, receipt tamper,
  privacy/log-redaction, and deterministic-byte vectors.
- Audit stale inference fallback documentation and prepare an update confined to
  T1 ownership.
- Add model-stage calibration, subgroup, receipt-lineage, model migration,
  uniqueness race, false-match adjudication and template-linkability test
  fixtures without claiming real accuracy or anonymity.

When this queue is exhausted, audit only T1-owned surfaces for additional
prototype-critical fail-open behavior, create numbered T1-X checkpoints, and
continue until no dependency-independent work remains. Never relabel contract or
fixture work as completion of 87A, 90C, 90D, GA, or production certification.
```
