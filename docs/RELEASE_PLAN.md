# VirtEngine First Release — Readiness Assessment & Plan

Status: **PLAN (not executed)**. Nothing in this document authorizes a tag, a publish,
or a deletion. Every step that cuts a tag or publishes a release requires explicit
human approval.

**Update 2026-09-19 — the pre-tag fix pass has landed** (branch `wt/t_c69f37a6`, commit
`fix(ci): repair the pre-tag release pipeline defects`). Defects 1, 2, 4, 5, 6, 7 and 9 are
fixed, Defect 8 is decided and implemented, and Defect 3 still needs human approval.
Section 5 carries the per-defect status; section 10 carries the updated blockers. The
changelog generator now runs green end to end in CI (`workflow_dispatch` dry run
`35356251383`, conclusion `success`). **Nothing was tagged, published or deleted.**

Owner: release-captain
Date: 2026-09-18
Audience: project-steward, secops, test-guard, docs-scribe, and the human who approves
the first tag.

Every fact below was produced by a command run against this checkout on 2026-09-18.
Commands to reproduce the whole assessment are in section 11. Where a claim could not be
verified it is marked **UNVERIFIED** rather than asserted.

---

## 1. Executive summary

`virtengine/virtengine` has a mature-looking release pipeline — tag-driven CI, a manual
publish workflow, changelog generation, SBOM + SLSA provenance, cosign signing, and
upgrade tests — and it has **never published a release**. The newest release record is a
**stale 2021 draft with an empty `tag_name`**, which can never be published.

This document answers the four questions the release task asks, with evidence:

1. What `release.yaml`, `make release` and `make gen-changelog` actually do, and whether
   the path works end to end → **section 3 and section 4**.
2. What blocks a first real release → **section 5**.
3. Whether there is a versioning scheme and a release branch model → **sections 6 and 7**.
4. What a first release requires (changelog, notes, artifacts, checksums,
   provenance/SBOM) → **section 8**.

**Conclusion: the release path is not green, and the failures are all pre-flight.**
Four of them are release engineering's to fix, and three of those are one-line-to-small
fixes — but none of them can be validated end to end on this host, because Docker is not
installed and the two self-hosted runners the publish workflow requires **do not exist**
(0 registered runners, verified). The security and CI gates are owned elsewhere.

Nothing here should be read as "the release is close". It is *well instrumented* and
*never exercised*.

---

## 2. Current verified release state

| Fact | Verified value | Evidence (command) |
|---|---|---|
| GitHub releases | exactly one: `0.1.0` — Draft, created 2021-10-08 | `gh release list -R virtengine/virtengine` |
| Draft release id | `51014212`, `tag_name: ""`, `draft: true`, `published_at: null`, `target_commitish: "main"` | `gh api repos/virtengine/virtengine/releases` |
| Published releases | **none, ever** | same |
| Remote tags | 32, of which **1** is semver: `v0.1.0` (`0744259a…`) | `git ls-remote --tags origin` |
| Non-release tags | 31 `checkpoint/*` tags (prototype/T-line checkpoints) | same |
| Commit `v0.1.0` points at | `2026-01-26T13:23:19+11:00` | `git log -1 --format=%cI v0.1.0` |
| Remote heads | 70; of the release-model branch names only `main` exists | `git ls-remote --heads origin` |
| `mainnet/main`, `develop`, `release/**` | **do not exist** on origin | same |
| Self-hosted runners | **0 registered** | `gh api repos/virtengine/virtengine/actions/runners` → `{"total_count": 0}` |
| Latest `security.yaml` run | `35318730864`, scheduled, 2026-09-18, **failure** | `gh run list --workflow=security.yaml` |
| Latest `ci.yaml` run on `main` | `33473526555`, push, 2026-09-01, **failure** | `gh run list --workflow=ci.yaml --branch main` |
| Docker on this host | **not installed** (`docker: command not found`) | `docker --version` |
| Go on this host | `go1.25.8 windows/amd64` | `go version` |

The 2021 draft is a **phantom**: its `tag_name` is empty, so GitHub cannot publish it and
GoReleaser cannot map it to a tag. It is a lie in the UI (see Defect 3).

---

## 3. What the release tooling actually does

### 3.1 `.github/workflows/release.yaml` — the manual publish workflow

Trigger: `workflow_dispatch` only, inputs `release_tag` (must start with `v`) and
`environment` (default `staging`).

1. `validate-release-context` refuses to run unless `GITHUB_REF` equals
   `refs/tags/<release_tag>` — the workflow **must be dispatched from the tag ref**.
2. Six reusable gate workflows must pass: `compatibility`, `post-deploy-smoke`,
   `staging-e2e`, `veid-e2e`, `ml-determinism`, `veid-conformance`.
3. `test-network-upgrade-on-release` runs `tests/upgrade` on a runner labelled
   `self-hosted, gh-runner-test`, only when `script/upgrades.sh test-required` says so.
4. `publish` runs on a runner labelled `self-hosted, core-e2e`, checks out the tag, logs
   in to GHCR, and runs `make release` with `GORELEASER_RELEASE=true`,
   `GORELEASER_MOUNT_CONFIG=true`, `GITHUB_TOKEN` (and `RELEASE_TAG` via `$GITHUB_ENV`).
5. `notify-homebrew` dispatches the Homebrew tap only when the tag classifies as mainnet
   and non-prerelease, and only if `GORELEASER_ACCESS_TOKEN` is set.

### 3.2 `make release` (`make/releasing.mk:122`)

Depends on `gen-changelog`, then runs GoReleaser in a Docker container
(`ghcr.io/goreleaser/goreleaser-cross:<gotoolchain>`) against `.goreleaser.yaml`:

- command `release --clean --release-notes=…/.cache/changelog.md`;
- `GORELEASER_SKIP` gains `publish` unless `GORELEASER_RELEASE=true`
  (`make/releasing.mk:24-30`), and `GITHUB_TOKEN` is blanked in that case.

So `make release` on its own **does not publish**; only the workflow (or an explicit
`GORELEASER_RELEASE=true GITHUB_TOKEN=… make release`) does. Runtime requirement: Docker
plus the goreleaser-cross image — unavailable on this host, so the command could not be
exercised here (**UNVERIFIED end to end**).

### 3.3 `make gen-changelog` (`make/releasing.mk:117` + `script/genchangelog.sh`)

1. Materialises `git-chglog` `v0.15.1` into the devcache bin (`make/setup-cache.mk:11`).
2. Runs `script/genchangelog.sh "<RELEASE_TAG>" .cache/changelog.md`, which
   classifies the tag as mainnet (even minor) or testnet (odd minor) via
   `script/mainnet-from-tag.sh`, builds a semver `--tag-filter-pattern`, and calls
   `git-chglog --config .chglog/config.yaml --tag-filter-pattern=…`.

`RELEASE_TAG` defaults to `git describe --tags --abbrev=0` (`make/init.mk:129`).

**This target is broken twice over on this repository today** — see Defects 2 and 4. It
did run successfully when given a real existing tag (`RELEASE_TAG=v0.1.0`), which is what
separates "broken by default" from "impossible".

### 3.4 `.github/workflows/ci.yaml` `release` job (`ci.yaml:1075`)

On `push` of a `v*` tag, after every gate (`ci-summary`, `test-python`, `test-portal`,
`sims`, `network-upgrade`, compatibility/smoke/staging-e2e/veid-e2e/ml-determinism/
veid-conformance), it downloads all artifacts, writes `checksums.txt`, and creates a
**Draft** release with `draft: true`, `generate_release_notes: true`, prerelease
auto-detected from `-rc`/`-beta`/`-alpha`.

Because the job `needs:` all of those gates and `ci.yaml` is red on `main`, **no draft is
produced on tag push today.**

### 3.5 `.github/workflows/changelog.yaml`

On `push` of a `v*` tag: `git-chglog --output CHANGELOG.md`, then
`git-chglog <tag> > RELEASE_NOTES.md`, uploads both, opens a PR to `main`, validates
commit-message conventions, and overwrites the release body with `RELEASE_NOTES.md`.
**Its generation step fails** — see Defect 2.

### 3.6 `.github/workflows/supply-chain.yaml`

On `v*` tags: SBOM bundle (Syft), a deterministic rebuilt binary compared byte-for-byte,
cosign keyless `sign-blob` over binary/checksum/SBOM, verification of those signatures,
SLSA provenance via `slsa-github-generator`, and `slsa-verifier verify-artifact`, then a
summary job that fails on any gate. This is the provenance/SBOM machinery the mission
asks for; it is wired and its jobs are `if: startsWith(github.ref, 'refs/tags/v')`.

### 3.7 Security-workflow SBOM vs `make sbom`

`security.yaml:388` generates SBOMs into **`.cache/sbom`**; `supply-chain.yaml:301`
generates them into **`.cache/supply-chain/sbom`**; `make sbom`
(`make/supply-chain.mk:30`) defaults to **`.cache/sbom`**
(`scripts/supply-chain/generate-sbom.sh:36`). `.goreleaser.yaml:199-202` attaches
`.cache/sbom/*.json|*.sig|*.cert` to the release. Only `make sbom` (or the security
workflow) populates the path GoReleaser reads, and `release.yaml`'s `publish` job runs
neither — see Defect 6.

---

## 4. Does the path work end to end?

| Step | Command | Result |
|---|---|---|
| Install changelog tool | `make gen-changelog RELEASE_TAG=v0.1.0` | **PASS** — installs `git-chglog v0.15.1`, writes `.cache/changelog.md` (48 bytes: header only, since `v0.1.0` is the oldest tag) |
| Generate notes for a real existing tag | same | **PASS** |
| Generate notes for a to-be-cut tag | `./script/genchangelog.sh v0.3.0 …` | **FAIL** — `ERROR commits corresponding to "v0.3.0" was not found`, leaves a 0-byte file |
| Generate notes with the default tag | `make gen-changelog` (no `RELEASE_TAG`) | **FAIL** — `git describe --tags --abbrev=0` returns `checkpoint/stable-virtengine-beta/consolidated-v2`, which is not semver; git-chglog errors, 0-byte file |
| Preview notes for an uncommitted tag | `git-chglog --config .chglog/config.yaml --next-tag v0.3.0` | **PASS** — 3082 lines, compare range `v0.1.0...v0.3.0`, 2864 entries |
| Build/publish artifacts | `make release` | **BLOCKED** — needs Docker (absent) and a `v*` tag; additionally mis-targeted (Defect 1) |
| Publish workflow runners | `gh api …/actions/runners` | **BLOCKED** — 0 self-hosted runners; `publish` and the upgrade gate would queue forever |
| Draft on tag push | `ci.yaml` `release` job | **BLOCKED** — `needs:` a red CI |
| Publish gate | `release.yaml` → `make release` | **BLOCKED** — runners, Defect 1, and the six gate workflows |

Short answer: **the install and template layers work; every path that produces or
publishes release notes or artifacts does not, today, in this repository.**

---

## 5. Defects (evidence, fix, owner)

**Status column added 2026-09-19 by the fix pass (branch `wt/t_c69f37a6`; the pre-tag fixes
landed as commit "fix(ci): repair the pre-tag release pipeline defects"). No tag, release or
publication was created. Where a status says FIXED, the command that proved it is named in
the notes below.**

| # | Defect | Status 2026-09-19 |
|---|--------|-------------------|
| 1 | GoReleaser publishes to `virtengine/node` | FIXED — `project_name`/`release.github.name` = `virtengine` (both configs) |
| 2 | `changelog.yaml` calls `git-chglog` with no `--config` | FIXED — `--config` + semver tag filter; pin aligned to 0.15.1; `workflow_dispatch` dry run green |
| 3 | Phantom 2021 draft release (empty `tag_name`) | OPEN — deleting a release record requires human approval |
| 4 | Default `RELEASE_TAG` is a `checkpoint/*` tag | FIXED — defaults to newest `v*` semver tag; `release-tag-check` fails loudly |
| 5 | Release notes cannot be previewed before the tag exists | FIXED — `--next-tag` mode + `make gen-changelog-preview` |
| 6 | The release ships no SBOM/signature extras | FIXED — publish generates + signs them and hard-fails when empty |
| 7 | Three writers to the same release notes | FIXED — GoReleaser is the single writer (decision implemented) |
| 8 | The publish path has no runner | DECIDED — both jobs repointed to `ubuntu-latest`; revert when runners are registered |
| 9 | Local verification cannot validate CI signatures | FIXED — identity regexp accepts workflow-URL and email; certificates standardised on `.pem` |
| 10 | Security gates red (12/13) | OPEN — owner `secops` |
| 11 | CI red on `main` | OPEN — owner `test-guard` |
| 12 | devcache paths on a Windows host | OPEN — low priority, documented |

### Fix-pass evidence (2026-09-19)

- **changelog.yaml `workflow_dispatch` dry run** (D2/D5 end-to-end): run `35356251383` on `wt/t_c69f37a6` — conclusion **success** (https://github.com/virtengine/virtengine/actions/runs/35356251383). Dispatched as `version=v0.3.0` (a tag that does not exist), `dry_run=true`.
- **D1** `.goreleaser.yaml` and `.goreleaser-test-bins.yaml` both carried
  `project_name: node`; `release.github.name` was `node`. Now `virtengine`. The
  `{{ .ProjectName }}` consumers (SBOM document template, docker image labels) follow.
  `checksum.name_template` was already `virtengine_{{ .Version }}_checksums.txt`.
- **D2** The workflow now passes `--config .chglog/config.yaml` and
  `--tag-filter-pattern` on both calls, generates the version notes through
  `script/genchangelog.sh` (one generator), and pins git-chglog `0.15.1` to match
  `make/init.mk:101`. `workflow_dispatch` dry run: see below.
- **D4** `RELEASE_TAG ?=` now resolves the newest `v*` semver tag
  (`git tag --list "v[0-9]*" --sort=-v:refname`), with a `v0.0.0` fallback when none
  exists. `make release-tag-check` rejects a non-semver value. Reproduced:
  `make gen-changelog` → exit 0 with real notes (was exit 1 + 0-byte file);
  `make release-tag-check RELEASE_TAG=checkpoint/stable-virtengine-beta/consolidated-v2`
  → exit 1 with an explanatory message.
- **D5** `script/genchangelog.sh [--next-tag] <tag> <out>` plus
  `make gen-changelog-preview`. Measured locally for the not-yet-cut `v0.3.0`:
  3084 lines of notes, exit 0. Without `--next-tag` the script still refuses a
  non-existent tag (`ERROR commits corresponding to "v0.3.0" was not found`, exit 1),
  so existing-tag behaviour is unchanged.
- **D6** The `publish` job now installs syft/cosign, runs `make sbom`, signs every
  `.cache/sbom/*.json` with cosign keyless (`.sig` + `.pem`), and fails the step when
  that produces no documents. The GoReleaser `extra_files` certificate glob moved
  `*.cert` → `*.pem` so it matches what the producers actually write. Empty-glob
  behaviour: **silent omission** (measured, above).
- **D7** `ci.yaml` `generate_release_notes: false`; `changelog.yaml`'s
  `action-gh-release` body step removed (its notes are still uploaded as a workflow
  artifact). GoReleaser with `--release-notes=.cache/changelog.md` and
  `mode: replace` is now the only writer of the release body.
- **D8** Re-verified before the change: `gh api
  repos/virtengine/virtengine/actions/runners` → `total_count: 0`. Both
  `test-network-upgrade-on-release` and `publish` now run on `ubuntu-latest`
  (docker/buildx comes from the existing QEMU/Buildx steps). If dedicated runners are
  registered later, restore the `[self-hosted, …]` labels.
- **D9** `CERTIFICATE_IDENTITY_REGEXP` now accepts both
  `…@virtengine.com` (locally signed) and
  `https://github.com/<owner>/<repo>/.github/workflows/<workflow>@<ref>` (CI keyless).
  Verified against four identities: both CI workflow URLs match, `release@virtengine.com`
  matches, `attacker@notvirtengine.com` and an unrelated URL do not. Certificate files
  are `.pem` everywhere (goreleaser, `make sign-artifact`, `generate-sbom.sh`,
  `supply-chain.yaml`).


### Defect 1 — GoReleaser publishes to the wrong repository (release-captain, blocker) — **FIXED 2026-09-19**

`.goreleaser.yaml:3` declares `project_name: node`, and `.goreleaser.yaml:192-195`:

```yaml
release:
  github:
    owner: virtengine
    name: node
```

`.goreleaser-test-bins.yaml:2` carries the same `project_name: node`. `node` is leftover
from the pre-fork lineage; `.github/repo` (used as the module mount) correctly says
`github.com/virtengine/virtengine`. With `GORELEASER_RELEASE=true`, GoReleaser would try
to create the release on `virtengine/node` — a repository that does not exist.

**Fix:** set `project_name: virtengine` and `release.github.name: virtengine` (or drop
`release.github` and let GoReleaser derive it from `origin`). Also review the archive/
`ProjectName`-derived names before the change lands: `checksum.name_template` is already
`virtengine_{{ .Version }}_checksums.txt`, the SBOM template is `sbom_{{ .ProjectName }}_…`,
and image labels use `{{ .ProjectName }}`, so this changes observable artifact naming.

### Defect 2 — `changelog.yaml` invokes `git-chglog` without a config (release-captain, blocker) — **FIXED 2026-09-19**

`changelog.yaml:69` runs `git-chglog --output CHANGELOG.md` and `changelog.yaml:72` runs
`git-chglog ${{ version }}`. Neither passes `--config`. git-chglog's built-in default is
`.chglog/config.yml`; this repository ships **`.chglog/config.yaml`**. Reproduced exactly:

```
$ git-chglog --output .cache/cl-default.md v0.1.0
 ERROR  open .chglog\config.yml: The system cannot find the file specified.
```

`run:` steps execute under `set -e`, so the `generate-changelog` job fails before it can
upload `CHANGELOG.md`/`RELEASE_NOTES.md`, open the changelog PR, or update the release
body. The workflow has never succeeded on a tag.

**Fix:** pass `--config .chglog/config.yaml` (and `--tag-filter-pattern` per
`script/genchangelog.sh`) in both calls, or rename the config to `.chglog/config.yml`.
Also bump the pinned `GIT_CHGLOG_VERSION 0.15.4` (`changelog.yaml:33`) into agreement with
`make/init.mk:101` (`v0.15.1`) — two different generators producing the same artifact is
drift by construction.

### Defect 3 — Phantom 2021 draft release (release-captain) — **OPEN (needs human approval to delete)**

Release id `51014212`: `name: "0.1.0"`, `tag_name: ""`, `draft: true`, `published_at:
null`, created 2021-10-08. An empty `tag_name` means it can never be published; it exists
only to mislead anyone reading the releases page.

**Fix (needs approval — deleting a release record is externally visible):** delete the
draft, or repurpose it for the real first tag at tag time. Do not leave it implying an
imminent release.

### Defect 4 — the default `RELEASE_TAG` is not a release tag (release-captain, blocker) — **FIXED 2026-09-19**

`make/init.mk:129`: `RELEASE_TAG ?= $(shell git describe --tags --abbrev=0 …)`. On this
repository the newest tag is `checkpoint/stable-virtengine-beta/consolidated-v2` (31 of
the 32 remote tags are `checkpoint/*`). Reproduced:

```
$ make gen-changelog
version checkpoint/stable-virtengine-beta/consolidated-v2 does not match the semver scheme…
 ERROR  commits corresponding to "checkpoint/stable-virtengine-beta/consolidated-v2" was not found
make: *** [make/releasing.mk:120: gen-changelog] Error 1
$ wc -l .cache/changelog.md → 0
```

Consequences: (a) `make release` as documented in `RELEASE.md` step 3 (a manual run) fails
before Docker is ever consulted; (b) if it ever did proceed, the notes source would be a
non-release checkpoint tag. In `release.yaml` the workflow happens to mask this by
exporting `RELEASE_TAG` into `$GITHUB_ENV`, but nothing in the repo documents that
requirement.

**Fix:** default `RELEASE_TAG` to the newest **semver** tag
(`git describe --tags --match 'v[0-9]*' --abbrev=0`) and make `gen-changelog` fail loudly
with a non-zero exit and a message when `RELEASE_TAG` is not a validated semver tag
(`./script/semver.sh validate`).

### Defect 5 — release notes cannot be previewed before the tag exists (release-captain) — **FIXED 2026-09-19**

git-chglog only knows tags that exist. `genchangelog.sh` therefore cannot generate the
notes for a tag that has not been cut yet — verified: it errors and leaves a 0-byte file,
which is exactly the file `make release` then passes to
`--release-notes=…/.cache/changelog.md`. In practice the workflow sets the tag before
running, so the failure is silent today; the moment anyone follows `RELEASE.md` and runs
`make release` (or reviews notes pre-tag) they get an empty notes file.

**Fix:** add a preview mode to `genchangelog.sh` that uses `git-chglog --next-tag <tag>`
when the tag does not exist. Verified working today:

```
$ git-chglog --config .chglog/config.yaml --tag-filter-pattern='^[v|V]?(0|[1-9][0-9]*)\.(\d*[13579])\.(0|[1-9][0-9]*)$' --next-tag v0.3.0
→ 3082 lines, ## [v0.3.0](…/compare/v0.1.0...v0.3.0), 2864 entries
$ git-chglog --config .chglog/config.yaml --next-tag v0.4.0   # even/minor, unfiltered base
→ 3076 lines
$ git-chglog --config .chglog/config.yaml --next-tag v0.3.0   # no filter pattern
→ 5216 lines, compare link starts at a checkpoint tag
```

Note the third result: without a `--tag-filter-pattern`, the compare range anchors on a
`checkpoint/*` tag. Filtering is not cosmetic here.

### Defect 6 — the release ships no SBOM/signature extras (release-captain) — **FIXED 2026-09-19 (empty-glob behaviour now measured)**

`.goreleaser.yaml:199-202` attaches `.cache/sbom/*.json|*.sig|*.cert` as `extra_files`.
`release.yaml`'s `publish` job runs only `make release`; it never runs `make sbom` or
`make sign-artifact`, and a fresh checkout has no `.cache/sbom`. So the SBOMs and cosign
signatures that the security/supply-chain workflows produce are **not** the ones attached
to the GitHub release — the glob evaluates against an empty directory.
Whether GoReleaser tolerates a glob that matches nothing or aborts was **UNVERIFIED**
in this assessment (no Docker here). **Measured 2026-09-19:** it silently omits. A
wildcard pattern whose static prefix is missing returns `(empty, nil)` from
`fileglob.Glob` (v1.4.1, called by GoReleaser's `internal/extrafiles.Find`), so the
release would publish with no SBOM and no signatures and report success. Reproduced
with a 5-case harness against the real library; only a wildcard-free literal path that
is missing returns an error. The fix below therefore adds an explicit emptiness guard
rather than relying on GoReleaser to notice.

**Fix:** decide one producer for release-time SBOMs/signatures and wire it into the
`publish` job before `make release` (e.g. `make sbom`), then dry-run with
`GORELEASER_RELEASE=false`. Verify the empty-glob behaviour at the same time.

### Defect 7 — three writers to the same release notes (release-captain) — **FIXED 2026-09-19 (decision implemented)**

`ci.yaml:1123` sets `generate_release_notes: true`; `changelog.yaml:144-149` overwrites
the body with `RELEASE_NOTES.md` (`append_body: false`); `make release` passes
`--release-notes=.cache/changelog.md` to GoReleaser, whose release settings are
`mode: replace`, `draft: false` (`.goreleaser.yaml:196-198`). Last writer wins, and the
publish step is a *replace*, so the manually reviewed draft body is destroyed at publish
time.

**DECIDED (no human available):** `git-chglog` output via `make gen-changelog` is the
single source of truth for release notes. Follow-ups: disable `generate_release_notes` in
`ci.yaml`, and turn `changelog.yaml`'s body step into a no-op (or restrict it to
`CHANGELOG.md` + the PR) rather than a competing writer. This is reversible — it is
workflow configuration, not a history rewrite.

### Defect 8 — the publish path has no runner (release-captain / project-steward, blocker) — **DECIDED 2026-09-19 (repointed to GitHub-hosted)**

`release.yaml:144-148` requires `[self-hosted, core-e2e]`; `release.yaml:90-94` requires
`[self-hosted, gh-runner-test]`. The API reports **0 registered runners** for this
repository. Both jobs will queue indefinitely; `notify-homebrew` can never be reached.

**Fix:** register runners with those labels (Docker + Buildx + Go 1.25.x, since
`make gen-changelog` runs on the runner host and `make release` needs the Docker socket),
or repoint `publish` at a GitHub-hosted runner with docker/buildx. Verify with a
`workflow_dispatch` dry run **before** the first tag.

### Defect 9 — local signature verification cannot verify CI-signed artifacts (release-captain) — **FIXED 2026-09-19**

`make/supply-chain.mk:100-105` verifies with
`--certificate-identity-regexp ".*@virtengine.com"`, but `supply-chain.yaml:445` signs
keyless with a workflow identity
(`https://github.com/<repo>/.github/workflows/supply-chain.yaml@<ref>`). An email-regexp
identity will not match a workflow-URL identity, so `make verify-signature ARTIFACT=…`
cannot validate what CI produces. Naming also diverges: `make sign-artifact` writes
`<artifact>.pem` while the workflow writes `<artifact>.sig.cert`.

**Fix:** align the local target's identity regexp with the workflow identity
(or accept `--certificate-identity-regexp 'https://github.com/virtengine/virtengine/.github/workflows/.*'`)
and standardise the certificate filename.

### Defect 10 — the security gates are red (secops, blocker) — **OPEN (owner: secops)**

Latest `security.yaml` run `35318730864` (2026-09-18, scheduled): **12 of 13 jobs fail**.

Failing: `CodeQL SAST`, `JavaScript Vulnerability Scan`, `Secret Scanning`,
`Policy Validation`, `gosec Security Scan`, `Go Vulnerability Scan`,
`Python Vulnerability Scan`, `Container Vulnerability Scan` ×4
(`virtengine`, `provider-daemon`, `veid-pipeline`, `test-runner`), `Security Summary`.
Passing: `Generate SBOM`.

(The estate record says "14 of 14"; the workflow currently exposes 13 jobs. Reality wins —
the count is corrected here.)

**Fix:** owned by secops (`make vuln-check`, `make supply-chain-audit`, `gosec ./...`).
Release-captain owns the dependency: no tag while this is red.

### Defect 11 — CI is red on `main` (test-guard, blocker) — **OPEN (owner: test-guard)**

Run `33473526555` (2026-09-01, push to `main`) failed these jobs:
`Lint`, `Go Vet`, `Go Tests`, `Integration Tests`, `Windows Native Build and Unit Tests`,
`DR Backup/Restore Smoke Test`, `CI Quality Gates`. The tag-triggered draft job `needs:`
that chain, so a tag push today yields no draft, no checksums, no release.

**Fix:** owned by test-guard (green baseline). Release-captain owns the dependency.

### Defect 12 — devcache paths on a Windows host (release-captain, low) — **OPEN (low, documented)**

Without direnv, `VE_DEVCACHE*` are unset (`make/init.mk` has a Windows fallback for
`VE_ROOT`, none for the devcache family), so `make gen-changelog` runs
`mkdir -p /git-chglog/` and dies with "Permission denied"; and Go rejects MSYS-style
`GOBIN` ("cannot install, GOBIN must be an absolute path") so the devcache must be a
native `C:/…` path. Recorded because it cost real time and will bite the next Windows
contributor.

**Fix (low priority):** derive the devcache defaults on Windows when they are unset, or
document the required exports in `RELEASE.md`.

---

## 6. Branching model (documented, with evidence)

Two descriptions exist in the tree and they disagree; resolved here.

- **Legacy/intended model** — repo `AGENTS.md`: `main` = active development
  (**odd** minor versions, e.g. `v0.9.x`); `mainnet/main` = stable releases (**even**
  minor versions, e.g. `v0.8.x`). `ci.yaml:31-36` triggers on `main`, `mainnet/main`,
  `develop`, and `release/**`; `security.yaml` and `supply-chain.yaml` list the same set.
- **Actual state (verified):** origin has 70 heads and only `main` among those names.
  There is no `mainnet/main`, no `develop`, no `release/**`.
- **Current documentation** — `RELEASE.md` ("this checkout does not use `mainnet/main` as
  the authoritative public release branch") and `_docs/version-control.md:40-48`
  ("treat them as compatibility or migration remnants") both agree, and both instruct that
  the policy documents be updated together if a stable line is ever restored.

**DECIDED (no human available):** cut the first release from `main`. Do **not**
materialise `mainnet/main` as part of this release; reactivating a stable line is a
launch-window decision (TestNet Jan 2027 / MainNet Mar 2027 per `RELEASE.md`) and would
need `RELEASE.md`, `README.md`, `docs/COMPATIBILITY.md`, `VERIFICATION.md` and
`_docs/version-control.md` updated in the same change. Reversible, conservative, and
consistent with the checked-in docs.

---

## 7. Versioning scheme

- Format: `vMAJOR.MINOR.PATCH[-prerelease][+build]`, validated by `script/semver.sh`.
- Minor-parity convention, enforced by `script/mainnet-from-tag.sh` and
  `script/genchangelog.sh`: **even minor = mainnet, odd minor = testnet/development**.
  Verified by running the helpers:

  ```
  v0.1.0       → valid, mainnet:no  (odd ⇒ testnet)  prerelease:no
  v0.3.0       → valid, mainnet:no                    prerelease:no
  v0.3.0-rc.1  → valid, mainnet:no                    prerelease:yes
  v0.4.0       → valid, mainnet:yes                   prerelease:no
  ```

- `v0.1.0` already exists as a tag, so it cannot be reused.
- The first release on the development line therefore needs an **odd** minor. Shape of the
  first candidate: **`v0.3.0-rc.1`** (release candidate, exercised through
  `tests/upgrade`), then **`v0.3.0`**. An even-minor "mainnet" first release would imply a
  support commitment that `RELEASE.md` explicitly says does not exist yet.
- The exact version string is a **human decision at tag time**; this plan fixes only the
  scheme and the shape.

Scale of the first release's notes: **3295 commits** since `v0.1.0` (235 days). By
conventional-commit type (subject prefixes, `git log v0.1.0..HEAD`): `fix` 762, `docs`
641, `feat` 554, `chore` 203, `test` 85, `refactor` 29, `ci` 28, `style` 7, `build` 5 —
plus non-conventional subjects that leak into the changelog (`Initial plan` ×34, variant
`merge`/`Merge …` subjects). The filtered preview contains **2864 entries** and group
headings the template never intended (`Blockers`, `Error`, `Upd`, `Polish`, `Summary`,
`Merge`, `Release`), because `.chglog/config.yaml` only maps `feat`, `fix`, `perf`,
`refactor`. See the checklist step 6 — the notes need a curation pass and probably a
`title_maps` extension before they represent a release.

---

## 8. What the first release will contain

From `.goreleaser.yaml` (and `.github/workflows/supply-chain.yaml` for provenance):

- binaries: `virtengine` for darwin/amd64, darwin/arm64 (folded into a universal binary),
  linux/amd64, linux/arm64;
- zip archives, with and without the version in the name
  (`virtengine_<ver>_<os>_<arch>.zip`, `virtengine_<os>_<arch>.zip`);
- `virtengine_<version>_checksums.txt` (sha256, cosign-signed via `signs: artifacts: checksum`);
- deb and rpm packages (nfpms);
- container images `ghcr.io/virtengine/virtengine` (`<shortcommit>`, `<version>`,
  `latest`/`stable`, plus multi-arch manifests) — note `docker_manifests` always publishes
  a `:latest` tag, including for a testnet/rc tag, which is worth a deliberate decision;
- cosign keyless signatures (`sign-blob`) and per-archive CycloneDX SBOMs
  (`sboms:` via syft) attached by GoReleaser — the release-time extras (Defect 6) are now
  produced by the `publish` job itself;
- SBOM bundle + signature verification + SLSA provenance from `supply-chain.yaml`;
- release notes from `git-chglog` (`make gen-changelog` → `--release-notes`).

Note the mismatch: `ci.yaml`'s draft bundle collects `*.tar.gz`/`*.zip` artifacts from the
CI jobs and errors if there are zero archives (`ci.yaml:1108-1111`), i.e. the draft path
expects tar.gz artifacts that GoReleaser (zip-only configuration) does not produce.

---

## 9. First-release checklist (sequenced, gated)

Anything that tags, publishes or deletes requires explicit human approval. Ownership in
parentheses.

1. **(secops)** Green `security.yaml` on `main` — currently 12/13 red (Defect 10).
2. **(test-guard)** Green `ci.yaml` on `main` — currently 7 jobs red (Defect 11).
3. **(release-captain)** ~~Fix Defect 1 (GoReleaser release target) — one small PR, plus a
   naming review.~~ **DONE 2026-09-19** — `project_name`/`release.github.name` = `virtengine`
   in `.goreleaser.yaml` and `.goreleaser-test-bins.yaml`; `{{ .ProjectName }}` consumers
   follow. Not executed: a GoReleaser run (no Docker/goreleaser on this host, and a
   pre-tag run has no tag to release).
4. **(release-captain)** ~~Fix Defect 4 (semver-only default `RELEASE_TAG`) and Defect 5
   (`--next-tag` preview mode) so notes are reviewable **before** the tag exists.~~
   **DONE 2026-09-19** — default is the newest `v*` semver tag, `release-tag-check` fails
   loudly on a non-semver value, and `--next-tag` preview mode exists
   (`make gen-changelog-preview`).
5. **(release-captain)** ~~Fix Defect 2 (`--config .chglog/config.yaml` + version pin) so
   `changelog.yaml` can run at all.~~ **DONE 2026-09-19** — dry run `35356251383`
   (`workflow_dispatch`, `version=v0.3.0`, `dry_run=true`) → **success**.
6. **(release-captain)** ~~Decide and wire Defect 6 (release-time SBOM/signature extras) and
   Defect 7 (single notes writer; disable the competing writers).~~ **DONE 2026-09-19** —
   the publish job generates and signs the SBOM extras and fails if it produces none;
   GoReleaser is the only release-body writer. Unverified remaining: the GoReleaser
   `extra_files` attachment itself needs a real publish run.
7. **(release-captain / project-steward)** ~~Register or repoint the runners (Defect 8) and
   prove it with a dispatch dry run.~~ **DECIDED + IMPLEMENTED 2026-09-19** — no runners
   exist, so both release jobs were repointed to `ubuntu-latest`; a dispatch dry run of
   `changelog.yaml` (GitHub-hosted) is green. The publish job itself still cannot be
   exercised before a tag exists.
8. **(release-captain)** Curate the notes: `git-chglog … --next-tag v0.3.0` preview,
   extend `title_maps`, and confirm the entry count/ordering is acceptable for a release.
9. **(human)** Approve the version string and the tag commit.
10. **(release-captain, after approval)** Tag `v0.3.0-rc.1` from the approved commit on
    `main` and push the tag — never force-push, never rewrite an existing tag. Expect
    `ci.yaml` (draft), `changelog.yaml`, `supply-chain.yaml` to run.
11. **(human)** Review the draft, then approve publishing via `release.yaml`, dispatched
    **from the tag ref**, environment `staging`.
12. **(release-captain)** Verify with `gh release view v0.3.0-rc.1` that the release is
    published with notes, archives, checksums, SBOM and signatures, and verify a
    signature with `cosign verify-blob` (after Defect 9). Never describe a release as "out"
    without that read-back.

Steps 3–7 are PR-sized, independently reviewable, and none of them touch history.

---

## 10. Blockers and dependencies

| Blocker | Owner | Status (2026-09-19) |
|---|---|---|
| `security.yaml` 12/13 red | secops | **open** (Defect 10) |
| `ci.yaml` red on `main` (7 jobs) | test-guard | **open** (Defect 11) |
| GoReleaser release target `virtengine/node` | release-captain | **fixed** (Defect 1) |
| `changelog.yaml` missing `--config` | release-captain | **fixed** — dry run green (Defect 2) |
| Phantom 2021 draft release | release-captain (+human approval to delete) | **open — needs approval** (Defect 3) |
| Non-semver default `RELEASE_TAG` | release-captain | **fixed** (Defect 4) |
| No pre-tag notes preview | release-captain | **fixed** (Defect 5) |
| SBOM/signature extras never attached | release-captain | **fixed** (Defect 6) |
| Three competing notes writers | release-captain | **fixed** (Defect 7) |
| 0 self-hosted runners (`core-e2e`, `gh-runner-test`) | release-captain / project-steward | **decided:** jobs repointed to GitHub-hosted; revisit if runners are registered (Defect 8) |
| Local `verify-signature` identity mismatch | release-captain | **fixed** (Defect 9) |

Explicitly **not** a blocker: the changelog generator itself. Given a real tag it works —
`make gen-changelog RELEASE_TAG=v0.1.0` completed and wrote `.cache/changelog.md`.

---

## 11. Reproducing this assessment

```bash
# release records and tags
gh release list -R virtengine/virtengine --limit 20
gh api repos/virtengine/virtengine/releases \
  --jq '.[] | {id,name,tag_name,draft,published_at,target_commitish}'
git ls-remote --tags origin | grep -E 'refs/tags/v[0-9]'
git ls-remote --heads origin | grep -E 'mainnet|develop|release/|refs/heads/main$'

# runners (publish path prerequisite)
gh api repos/virtengine/virtengine/actions/runners --jq '{total_count}'

# gate state
gh run list -R virtengine/virtengine --workflow=security.yaml --limit 3
gh run view 35318730864 -R virtengine/virtengine --json jobs \
  --jq '.jobs[] | "\(.conclusion)\t\(.name)"'
gh run list -R virtengine/virtengine --workflow=ci.yaml --branch main --limit 3
gh run view 33473526555 -R virtengine/virtengine --json jobs \
  --jq '[.jobs[] | select(.conclusion=="failure") | .name]'

# release tooling (needs the devcache exported; git-chglog installs itself)
export VE_ROOT=<repo>            # NOTE: Go requires a native path, e.g. C:/…
export VE_DEVCACHE=$VE_ROOT/.cache
export VE_DEVCACHE_BIN=$VE_DEVCACHE/bin
export VE_DEVCACHE_VERSIONS=$VE_DEVCACHE/versions
export VE_DEVCACHE_INCLUDE=$VE_DEVCACHE/include
export VE_DEVCACHE_NODE_MODULES=$VE_DEVCACHE/node_modules
export VE_RUN=$VE_DEVCACHE/run VE_RUN_BIN=$VE_RUN/bin

make gen-changelog RELEASE_TAG=v0.1.0      # PASS (header only: oldest tag)
make gen-changelog                          # FAIL: default tag is a checkpoint tag
./script/genchangelog.sh v0.3.0 .cache/changelog-v030.md   # FAIL: tag does not exist
git-chglog --output .cache/x.md v0.1.0      # FAIL: .chglog/config.yml not found
git-chglog --config .chglog/config.yaml --next-tag v0.3.0  # PASS: preview mode

# tag classification
for t in v0.1.0 v0.3.0 v0.3.0-rc.1 v0.4.0; do
  ./script/semver.sh validate "$t"; ./script/mainnet-from-tag.sh "$t"; done

# scale of the first release
git rev-list --count v0.1.0..HEAD
git log v0.1.0..HEAD --format=%s | sed -E 's/^([a-zA-Z]+)(\(.*\))?:.*/\1/' | sort | uniq -c | sort -rn

# artifact/publish configuration
grep -n -E '^(project_name|release):' -A6 .goreleaser.yaml
```

---

## 12. Non-negotiables

- **Never create, tag, or publish a release without explicit human approval.**
- Never rewrite or delete an existing tag; never force-push a release branch.
- Never write release notes for changes not verified in the commit range.
- Never claim a release "is out" without `gh release view` confirmation.
- Keep `Draft` status meaningful: a draft is either finished or explained.

---

## Related documentation

- `RELEASE.md` — current release process and launch windows
- `_docs/version-control.md` — branch/tag utilities and working model
- `docs/NETWORK_LAUNCH_SCHEDULE.md` — environment boundary and promotion criteria
- `docs/COMPATIBILITY.md`, `VERIFICATION.md`
- `.github/workflows/{release,ci,changelog,supply-chain,security}.yaml`
- `.goreleaser.yaml`, `make/releasing.mk`, `make/supply-chain.mk`, `script/genchangelog.sh`
