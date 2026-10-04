# x/fraud <-> x/roles sanction join — implementation plan (option 4, panel-RESOLVED)

Card: t_998bb053. Design decided and recorded in the card thread (Tier-1 panel, exit 0).
Branch: wt/t_998bb053-v2. Base: develop. Toolchain: Go 1.26.8, `GOFLAGS=-mod=mod`, `GOWORK=off`.

## Verified baseline (real output, this worktree)

    go build ./x/roles/... ./x/fraud/...   -> exit 0
    go test  ./x/roles/... ./x/fraud/...   -> all ok, TEST_EXIT=0
      ok x/roles  ok x/roles/keeper  ok x/roles/types  ok x/roles/types/privileged
      ok x/fraud/keeper  ok x/fraud/types   (x/fraud, x/fraud/client/cli: no test files)

`GOFLAGS=-mod=vendor` FAILS on this tree: vendor/modules.txt is not marked as
replaced for 9 modules. Use `-mod=mod`.

## Facts established by reading the source (do not re-derive)

- `x/fraud/keeper/keeper.go:92` `RolesKeeper` iface = HasRole/IsModerator/IsAdmin only.
- `app/types/app.go:652-658` ALREADY passes `app.Keepers.VirtEngine.Roles` into
  `fraudkeeper.NewKeeper`. Only the interface needs widening; no app wiring change.
- `applyResolution` (`x/fraud/keeper/keeper.go:737`) writes report + queue removal + audit.
- `ProposeResolution` / `ConfirmResolution` live in `x/fraud/keeper/pending_resolution.go`
  (`:89`, `:175`). `ConfirmResolution` already returns `ResolutionType`; the msg server
  (`x/fraud/keeper/msg_server.go:348`) already runs it in a **cache context** — the
  atomicity the panel wants is partly there already; extend it to cover the sanction.
- `ImposeSanction` (`x/roles/keeper/sanction.go:154`) and `ConfirmSanction` (`:266`) are
  the roles-side entry points. Both validate `reviewer != imposed_by` themselves.
- `Sanction` (`x/roles/types/sanction.go:430`) has **no proto twin** — JSON in store.
  `PendingResolution` (`x/fraud/types/pending_resolution.go:33`) likewise. This is the
  established precedent: the linked sanction id needs NO new proto message, only a
  `sanction_id` string field on the JSON `PendingResolution`.
- **Termination is intentionally indefinite**: `ExpiresAt` 0 == termination only
  (`x/roles/types/sanction.go:462-464`); `ImposeSanction` requires a future expiry for
  `SanctionKindSuspension` (`:205-215`) and NOT for `SanctionKindTermination` (`:217-219`).
  The card's premise "sanctions require expires_at" is wrong for termination.
- ConsensusVersion: x/roles = 2 (`x/roles/module.go:205`), x/fraud = 2 (`x/fraud/module.go:152`).
- Test to replace: `TestConfirmedFraudResolutionDoesNotImposeAnAccountSanction` at
  `x/fraud/keeper/msg_server_cosign_test.go:268` — it pins the UNJOINED behaviour and WILL
  go red once the join lands. Must be deleted/rewritten, never left silently red.
- Docker is unavailable on this host, so `scripts/proto-generate.sh` (docker build) cannot
  run. Local `buf` 1.47.2 + `~/go/bin/protoc-gen-gocosmos` reproduce protos byte-identically
  (normalize CRLF) — that is the known-good local path from prior work.

## Step 1 — widen the interface (no proto change)

`x/fraud/keeper/keeper.go`, add to `RolesKeeper` (keep it NARROW — this is the
separation-of-concerns guardrail the panel insisted on):

    // ProposeLinkedSanction / ConfirmLinkedSanction / CancelLinkedSanction are the
    // only roles writes x/fraud may perform, and only for a sanction it created.
    ProposeLinkedSanction(ctx, subject string, kind rolestypes.SanctionKind,
        durationSeconds int64, reasonCode, justification, notice, imposedBy, sourceReportID string) (string, error)
    ConfirmLinkedSanction(ctx, sanctionID, reviewer string) error
    CancelLinkedSanction(ctx, sanctionID, reason string) error

Implementing these on `x/roles/keeper` (NOT reusing ImposeSanction's caller-visible
signatures blindly): tag the sanction with `SourceModule: "fraud"`, `SourceRef: report_id`.

## Step 2 — roles-side source tagging + guard

- Add `SourceModule`/`SourceRef` to the `Sanction` Go struct (`x/roles/types/sanction.go:430`).
  JSON-absent => decodes as empty => standalone sanctions unaffected (panel requirement).
- `ConfirmSanction` must REFUSE a `SourceModule == "fraud"` sanction
  (`ErrSanctionFraudLinked`) so it cannot be confirmed independently of the report.
- Add a constructor that takes explicit terms; do not invent defaults.

## Step 3 — fraud-side join

- `PendingResolution` gains `SanctionID string \`json:"sanction_id,omitempty"\``.
- `ProposeResolution`: for Suspension/Termination, require terms; call
  `ProposeLinkedSanction`. A Warning/NoAction carrying terms is rejected.
- `ConfirmResolution`: confirm the linked sanction FIRST, then applyResolution, both
  inside the existing cache context so a failure commits neither.
- Lapse/replacement (`ExpirePendingResolutions`, `ProposeResolution` overwrite) must
  `CancelLinkedSanction` — a linked pending sanction must never remain confirmable.

## Step 4 — duration + fail-closed

- Suspension: moderator supplies `duration_seconds`; `expires_at = BlockTime + duration`.
- Termination: `expires_at = 0` (indefinite) — no invented default.
- Old-format messages (no terms) must FAIL CLOSED after activation, never acquire a default.

## Step 5 — migrations + consensus version

- Bump BOTH modules 2 -> 3 and register real migrations.
- Migration must be IDEMPOTENT and TESTED (standing rule).
- No backfill of account state. Existing resolved reports surface `LEGACY_NO_SANCTION`.
- Existing pending fraud resolutions lack terms -> deterministically lapse/cancel in migration.

## Step 6 — tests (DONE WHEN #2)

1. Replace `TestConfirmedFraudResolutionDoesNotImposeAnAccountSanction` with
   `TestConfirmedFraudResolutionImposesLinkedSanction` using a **real roles keeper**, not
   the existing mock: propose -> report STILL unresolved AND account STILL active;
   confirm with a distinct reviewer -> report resolved, sanction ACTIVE, both reviewers
   recorded, source recorded, account SUSPENDED.
2. Warning -> no sanction created, and terms rejected.
3. Termination -> expires_at == 0, indefinite.
4. Suspension expiry -> EndBlocker auto-clears and account restored.
5. Lapse -> linked sanction cancelled, not independently confirmable.
6. `MsgConfirmSanction` on a fraud-linked sanction -> error.
7. Atomicity: roles confirm fails -> report NOT resolved (no partial commit).
8. Migration idempotence + legacy no-backfill.
9. `go run ./scripts/consensusdeterminism` green.

## Gates before PR

    GOFLAGS=-mod=mod GOWORK=off go build ./x/roles/... ./x/fraud/...
    GOFLAGS=-mod=mod GOWORK=off go test  ./x/roles/... ./x/fraud/...
    go run ./scripts/consensusdeterminism
    git fetch origin develop && git rebase origin/develop   # branch is 86 behind
    gh pr create --base develop

CI on develop is red at baseline (Lint, Go Tests, Integration, contracts, Go Security
Scan, Security Summary) as of develop head 1c65bcaf — diff against that, do not assume
those reds are ours.
