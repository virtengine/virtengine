# VEID AI, Privacy-Preserving Biometric Uniqueness, and Private Claims Architecture

**Status:** Prototype architecture and implementation backlog
**Prepared:** 2026-08-01
**Owners:** Threads T1-T5
**Production status:** Disabled until every mandatory security and certification gate passes

## 1. Purpose

This document defines the intended model portfolio, decision policy, encrypted
evidence lifecycle, face-authentication boundary, and privacy-preserving
biometric uniqueness service for VEID.

The following are separate security decisions and must never be represented by
one score or one reusable biometric template:

1. identity verification: whether authenticated evidence supports the claimed
   identity;
2. biometric uniqueness: whether the same biometric subject has already enrolled
   in the governed eligibility program;
3. authentication: whether the present user authorizes this session or exact
   transaction;
4. eligibility: whether current policy permits a specific action;
5. fund authorization: whether a one-time, transaction-bound authorization
   permits a value-moving message.

Biometric uniqueness may gate initial mint eligibility. It must not be used as
the sole factor for wallet recovery, fund release, payout, withdrawal, or other
value movement.

## 2. Current Safety Findings

The current repository must be treated as prototype-only until these findings
are closed:

- multi-recipient envelope encryption derives data protection through an invalid
  all-zero peer-key path, and envelope authentication is a forgeable hash rather
  than a sender signature;
- active OTP verification accepts arbitrary non-empty TOTP, SMS, and email
  responses;
- the current `biometric_verified` proof is not bound to a governed issuer,
  credential, subject, challenge, or uniqueness decision;
- derived biometric hashes can be updated without a complete governed validator
  authorization path;
- the current truncated SHA-256 "LSH" is neither locality-sensitive nor global
  and cannot perform fuzzy cross-account deduplication;
- some face, liveness, GAN, anomaly, U-Net, and trust-score paths are deterministic
  placeholders, random/untrained models, or silently empty;
- Python training and Go serving use incompatible feature positions;
- raw encrypted evidence can remain in consensus state while erasure records do
  not invoke durable storage or key destruction;
- provider vault storage and keys are process-local, and missing consent can
  resolve to allow-all;
- settlement, escrow, reward, and payout authorization is not comprehensively
  bound to MFA or a one-time fund authorization;
- identity verification plus uniqueness does not currently gate actual token
  minting or fund unlock.

No prototype milestone may describe these paths as anonymous, production-ready,
compliant, or cryptographically erased until executable evidence proves it.

## 3. Model Portfolio

Model source code, framework, weights, and training data each require separate
license and provenance approval. Runtime model download and trust-on-first-use
weight registration are prohibited.

| Stage | Prototype baseline | Production candidate | Decision use |
| --- | --- | --- | --- |
| Capture quality | OpenCV deterministic quality/orientation/perspective checks | Same pinned preprocessing graph | Hard input gate only |
| Document routing | Governed rules over declared country/type/version | Custom trained MobileNetV3/ConvNeXt-class classifier exported to pinned CPU runtime | Select exact document template and validators |
| Machine-readable data | MRZ/PDF417/barcode parsers with checksum validation | Same, versioned by schema | Authoritative where cryptographically/checksum valid |
| OCR | Pinned Tesseract baseline after deterministic preprocessing | Governed OCR model such as PaddleOCR/TrOCR class only after weight/license/evaluation approval | Extract candidate fields; never establish authenticity alone |
| Document authenticity | Template, checksum, cross-field, image-quality rules | Custom tamper/portrait-substitution/security-feature models plus calibrated rules | Separate authenticity decision |
| Face detection/alignment | Existing MTCNN/RetinaFace wrappers in test profiles | SCRFD/RetinaFace-class detector with owned or explicitly licensed weights | Locate and normalize face only |
| Face comparison | Existing DeepFace adapters in test profiles | ArcFace-class embedding model with owned or explicitly licensed weights | Calibrated 1:1 selfie-to-document comparison |
| Passive liveness | Existing heuristic signals for test characterization | Independently evaluated MobileNetV3/CDCN-class PAD model with owned/licensed weights | Hard PAD gate |
| Active liveness | Randomized blink/head-turn/smile challenge | Signed randomized challenge with server nonce and anti-replay | Independent intent/presence gate |
| Age | Authenticated document DOB and evaluation date | Same | Authoritative age claim |
| Facial age estimation | Disabled | Optional advisory model after subgroup evaluation | Fraud/manual-review signal only; never eligibility |
| Fraud fusion | Explicit deterministic rules | Calibrated monotonic logistic or GBDT model over authenticated stage outputs | Risk/review decision, separate from identity |
| Trust/eligibility | Versioned policy evaluator | Same | Action-specific decision, not opaque neural score |
| LLM | Disabled in decision path | Schema-constrained, redacted case summarization or unsupported-document triage | Human assistance only; no biometric, eligibility, mint, or fund decision |

DeepFace wrapper licensing does not establish rights to its model weights or
training data. MS-Celeb-derived or otherwise unclear artifacts remain disabled.
Production selection requires an artifact manifest with SHA-256, source,
redistribution right, dataset lineage, model card, SBOM, feature schema,
evaluation report, and runtime image digest.

## 4. Decision and Scoring Policy

The inference worker emits independently calibrated, fixed-point stage results:

- `DocumentDecision`: supported type, checksums, field consistency,
  authenticity confidence, reason codes;
- `FaceMatchDecision`: detector/embedding profile, calibrated similarity band,
  threshold and quality;
- `LivenessDecision`: passive PAD, active challenge, assurance level and expiry;
- `UniquenessDecision`: unique, possible-match-review, duplicate-confirmed, or
  unavailable;
- `RiskDecision`: risk band, authenticated signals, holds and review reasons;
- `IdentityDecision`: verified, review, rejected, or unavailable;
- `EligibilityDecision`: action, result, dependency decisions, policy digest and
  expiry;
- `FundAuthorization`: exact message digest, amount, denomination, source,
  destination, case/order IDs, signer/MFA session, nonce and expiry.

Every decision binds chain ID, subject/account scope, evidence commitments,
model/runtime/schema digests, policy version, signer epoch, block bounds, reason
codes, and supersession state.

Hard gates cannot be outweighed by an average score. Initial mint eligibility
requires all of the following:

1. current verified identity decision;
2. current liveness decision at the required assurance level;
3. unique biometric decision, or completed independent adjudication;
4. no active fraud/compliance/account hold;
5. current account standing and policy version;
6. one atomic, idempotent eligibility consumption for the program nullifier.

Aggregate score is for explanation and prioritization only. A high face match
cannot compensate for failed liveness, unsupported documents, duplicate
biometrics, expired evidence, or an active hold.

Fund movement requires a separate one-time `FundAuthorization`. Local platform
biometrics should normally unlock a WebAuthn/passkey private key. Optional remote
face authentication is supplemental 1:1 verification and must be combined with
a possession factor. Face alone is never sufficient for recovery or release.

## 5. Evidence and Claim Lifecycle

1. Capture obtains explicit versioned consent bound to purpose, artifact types,
   verifier, retention, derived claims and intended relying parties.
2. The client encrypts raw evidence to a short-lived ingestion key and uploads
   it directly to an off-chain durable vault. Raw bytes or complete encrypted
   envelopes never enter consensus state.
3. Chain state receives only an opaque object reference commitment, evidence
   digest, policy/profile digests, key epoch and processing status.
4. An isolated inference worker decrypts authorized objects, executes the pinned
   model graph and emits signed stage receipts.
5. An issuer derives minimal claims such as DOB, jurisdiction, document validity,
   liveness level and decision epoch. Claims are encrypted to a user-controlled
   wallet/vault key; chain state stores status and randomized commitments only.
6. Terminal verification schedules deletion of raw images, OCR intermediates and
   temporary embeddings. The worker destroys per-object DEKs and records a
   durable backend/KMS deletion receipt. Legal hold is an explicit delayed state.
7. A provider/order creates a policy challenge. The wallet selects claims and
   produces an audience-, order-, purpose-, nonce-, expiry-, holder-, issuer-,
   status- and policy-bound presentation.
8. The provider verifies the presentation off-chain. Chain state records only a
   presentation digest, policy/status epochs, expiry and pass/fail commitment.
9. Consent withdrawal denies new access immediately, revokes grants and schedules
   deletion subject to the explicit retention/hold matrix.

Recovery backs up wrapped DEKs and metadata, never raw KEKs. Rotation and restore
are resumable, threshold-authorized and prove old-holder invalidation.

## 6. Privacy-Preserving Biometric Uniqueness

### 6.1 Required privacy property

There is no safe design where a public deterministic hash of a noisy face is
both fuzzy-matchable and non-linkable. Hashing an embedding is not anonymization.
The public chain must never receive an embedding, perceptual hash, fuzzy hash,
distance-search index, or globally reusable biometric identifier.

"Anonymous" means the uniqueness service cannot expose identity attributes or
raw templates to chain participants, and public outputs are scoped pseudonymous
nullifiers. A stable program nullifier is necessarily linkable within that one
eligibility program because enforcing one enrollment requires continuity.

### 6.2 Service architecture

1. A pinned face pipeline generates multiple quality-controlled embeddings in an
   isolated runtime after liveness and document-face comparison.
2. A purpose-specific cancellable transform creates uniqueness templates. Keys,
   transforms and templates are distinct from authentication and identity
   verification templates.
3. Templates are threshold-encrypted across independently operated uniqueness
   nodes. No single database, operator or application process holds plaintext
   global templates and the complete transform key.
4. A threshold-MPC, confidential-compute, or approved hybrid service performs
   global 1:N distance search. The exact construction must pass cryptographic
   review, template-inversion/linkability testing and malicious-node analysis;
   merely placing plaintext search in one enclave is insufficient.
5. Enrollment is atomic: search, adjudication state and insertion share one
   serialized/idempotent operation so concurrent duplicate enrollments cannot
   both pass.
6. A final governed no-match creates an internal random subject handle. An OPRF/threshold-PRF
   derives a stable program nullifier from program ID, subject handle and policy
  domain. Pairwise presentation nullifiers use separate relying-party domains.
  Pending, possible-match, rejected and unavailable outcomes use short-lived
  request-scoped opaque references and never receive a stable nullifier.
7. The service returns a threshold-signed `UniquenessReceipt` containing the
  stable scoped nullifier only for final unique enrollment, plus decision,
  threshold/profile/version, evidence/model/runtime commitments, signer set,
  freshness, reason codes and appeal reference.
8. Possible matches enter independent human adjudication. No fuzzy candidate is
   an automatic denial, account linkage, or adverse action.
9. Confirmed duplicates link only internal subject handles and governed case
   records. Identity claims remain in separately encrypted stores.
10. Template compromise triggers transform/key rotation and re-enrollment or
    governed migration. Biometrics are not secrets and cannot be reset like a
    password, so breach impact and recovery are first-class requirements.

For the prototype, implement contracts, deterministic simulators, concurrency
tests and fail-closed service interfaces only. Do not claim cryptographic
anonymity until an externally reviewed MPC/confidential-compute implementation
and real privacy attack evaluation exist.

## 7. Face Authentication

Preferred face-backed authentication uses device-local platform biometrics to
unlock a non-exportable WebAuthn/passkey. VirtEngine receives a challenge-bound
signature, not facial metrics.

If remote face authentication is later enabled:

- use a separate per-account 1:1 cancellable template and key domain;
- require fresh active and passive liveness plus device/session attestation;
- bind the result to chain ID, account, exact action/message digest, amount,
  recipient, nonce and short expiry;
- combine it with a possession factor;
- provide two non-biometric recovery factors and a delayed threshold recovery
  path with notification and cancellation;
- never query the global uniqueness index for routine authentication.

## 8. Evaluation and Release Gates

Required stage metrics include:

- OCR CER/WER and field-level accuracy by document/country/version;
- document routing confusion and unsupported-document rejection;
- face verification FMR/FNMR/EER and calibration by device, document and
  demographic/intersectional subgroup;
- PAD APCER/BPCER by attack instrument, deepfake/replay/mask coverage and active
  challenge failure;
- uniqueness 1:N false-match/false-non-match, candidate-set size, race behavior,
  adjudication outcomes and subgroup limits;
- fraud/eligibility calibration, reason stability and appeal overturn rates;
- template inversion, hill-climbing, membership inference, linkage, node
  collusion, enumeration and compromise-recovery tests.

Production activation also requires signed DPIA and jurisdiction assessment,
biometric consent, a non-biometric alternative, retention precedence matrix,
appeals, key ceremonies, incident runbooks, independent PAD/privacy/security
review, representative datasets, external model/weight license approval, and
exact-digest deployment evidence.

## 9. Ownership

- T1 owns canonical feature, stage-decision, eligibility and uniqueness receipt
  source schemas. T5 implements uniqueness custody/search/enrollment behind
  those interfaces. T2 and T3 are consumers only. T4 owns generated output.
- T5 owns canonical EvidenceObjectRef, retention and deletion-receipt source
  schemas and vault implementation. T2 owns capture/upload/claim clients. T3
  owns operational job/metric projections. T4 owns generated output and wiring.
- T2 owns passkey, user encrypted-claim and selective-presentation clients.
- T3 owns model, uniqueness, deletion and appeal operational evidence, not the
  authoritative domain records.
- T4 owns artifact/license manifests, cross-language parity, prohibited-model
  gates, generated contracts, cross-module fund-handler wiring and the exact-SHA
  prototype manifest.
- T5 owns cryptography repair, OTP, durable vault/KMS, consent enforcement,
  uniqueness service custody, confidential issuer-link continuity, canonical
  fund authorization, recovery and privileged policy.

The supplemental checkpoints in `_docs/thread-queues/` are the executable
backlog for this architecture.