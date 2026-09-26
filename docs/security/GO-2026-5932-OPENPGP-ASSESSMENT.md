# GO-2026-5932 — `golang.org/x/crypto/openpgp` unmaintained: reachability assessment and accepted risk

**Advisory:** GO-2026-5932 (no CVE; the package is deprecated upstream as unmaintained and
unsafe by design)
**Affected module:** `golang.org/x/crypto` (subpackage `openpgp/...`)
**Patched versions:** none — the Go vuln DB records `Fixed in: N/A`, and the OSV record carries a
single `{introduced: "0"}` range with no `fixed` event.
**Assessed on:** `develop` @ `0a7b0f1748113fd9569bb5a1931feca903ad7459`, 2026-09-26
**Toolchain:** Go 1.26.8, `golang.org/x/crypto v0.56.0`
**Verdict:** govulncheck reports the advisory at **module granularity only, with no symbol
traces** — no package in this repository's build imports, links, or initialises
`golang.org/x/crypto/openpgp`, and no fixed version exists upstream. Dated governance record
with time-boxed residual risk; tracked in
[issue #954](https://github.com/virtengine/virtengine/issues/954) and
`.vulnerability-allowlist.yaml`. This record does not clear, suppress, or gate-route-around the
`Go Vulnerability Scan` finding — that job does not consume the allowlist.

---

## 1. What the advisory says

The `golang.org/x/crypto/openpgp` package (OpenPGP packet parsing, key handling, ASCII armor)
is unmaintained and considered unsafe by design, with known security issues. Upstream's answer is
deprecation, not a patch: there is no fixed version and none is expected.

The finding cannot be cleared by a version bump. This was tested, not assumed:

- `golang.org/x/crypto v0.57.0` (newest release at assessment time) **still ships**
  `openpgp/armor/armor.go` (verified against the module proxy source).
- The OSV record for GO-2026-5932 contains a single `{introduced: "0"}` range event and no
  `fixed` event, so no version in any line is marked fixed.
- The ecosystem_specific field names only import paths (`openpgp`, `openpgp/packet`,
  `openpgp/armor`, `openpgp/clearsign`, `openpgp/errors`, `openpgp/elgamal`, `openpgp/s2k`) and
  no symbols, so govulncheck reports at package/module granularity.

## 2. Dependency route

`golang.org/x/crypto` is a **direct** dependency (`go.mod:76`, `v0.56.0`), required for
`ssh` (slurm/moab adapters), `argon2`, and `chacha20poly1305`. The vulnerable *subpackage*,
however, has no route into the build:

- A repository-wide search of `*.go` for `golang.org/x/crypto/openpgp` returns **zero** files.
  No first-party code imports it.
- `go list -deps ./... | grep 'golang.org/x/crypto/openpgp'` returns **zero** packages. The
  subpackage is not linked and none of its `init` functions run. It is dead weight inside a
  required module.

## 3. Why the subpackage is unlinked: the fork already migrated

The historical route was `github.com/cosmos/cosmos-sdk/crypto` (`armor.go`), which upstream
still carries:

```go
"golang.org/x/crypto/openpgp/armor" //nolint:staticcheck //TODO: remove this dependency
```

(present in upstream `v0.50.11`, `v0.53.0`, `v0.54.3`, and `main`).

But this repository does not build upstream. `go.mod:112` pins the VirtEngine fork:

```
github.com/cosmos/cosmos-sdk => github.com/virtengine/cosmos-sdk v0.53.4-virtengine.2
```

and that fork's `crypto/armor.go` already uses the maintained replacement:

```go
"github.com/ProtonMail/go-crypto/openpgp/armor"
```

(`github.com/ProtonMail/go-crypto v1.4.1`, `go.mod:179`, indirect.) The fork is *ahead* of
upstream here, so no parent-stack upgrade is needed or available to change this — the migration
off the deprecated package already happened. The only `openpgp/armor` in the dep graph is the
ProtonMail one, which this advisory does not cover.

## 4. `govulncheck`'s own output agrees

`govulncheck -show verbose ./...` on the assessed commit lists GO-2026-5932 under
**Module Results** with no example traces:

```
Vulnerability #1: GO-2026-5932
    The golang.org/x/crypto/openpgp package is unmaintained, unsafe by design,
    and has known security issues
  Module: golang.org/x/crypto
    Found in: golang.org/x/crypto@v0.56.0
    Fixed in: N/A
```

with the summary line `Your code is affected by 1 vulnerability from 1 module` referring to the
unrelated call-level GO-2026-4740 (msgpack), not to this advisory.

**Why the scanner still lists it.** With no symbols in the advisory, govulncheck flags every
package in the affected module at package granularity: mere module requirement is enough for a
module-level listing. That is a statement about the dependency manifest, not about any reachable
call path — and it is the specific reason the listing cannot be reasoned away from the tool
output and has to be closed by a documented human decision.

**Gate impact, measured.** Module-level-only findings do not drive the gate: `govulncheck
./x/...` on the same commit reports module-level findings yet exits 0 (`Your code is affected
by 0 vulnerabilities`), and the `Go Vulnerability Scan` job fails only when the text-mode exit
is 3 (see `govulncheck_verdict.sh`). The current gate failure is driven by GO-2026-4740, which
is tracked separately.

## 5. Decision

**Accepted residual risk, time-boxed.** Recorded in `.vulnerability-allowlist.yaml` with a 30-day
expiry under the repository's existing policy (`max_allowlist_age_days: 30`,
`require_compensating_controls: true`).

Compensating controls:

1. No package in `go list -deps ./...` imports or links `golang.org/x/crypto/openpgp`; the
   subpackage's code never executes in any shipped binary. The module stays required for
   `ssh`/`argon2`/`chacha20poly1305` — this control narrows the finding to manifest-only, it
   does not remove the listing.
2. The one historical call path (cosmos-sdk ASCII-armor key encoding) now resolves to the
   maintained ProtonMail fork via our pinned `virtengine/cosmos-sdk v0.53.4-virtengine.2`;
   even the armor framing use never touches OpenPGP packet crypto.
3. The exception is time-boxed to 30 days and re-reviewed at PR time; expiry is enforced by
   `.github/scripts/validate_security_policies.py`, which fails the `Policy Validation` job if
   the entry lapses or the age limit is exceeded.

**This entry does not turn `Go Vulnerability Scan` green.** `.vulnerability-allowlist.yaml` is
validated but is not consumed by any scanner in this repository. That job additionally fails on
the unrelated call-level GO-2026-4740, tracked separately.

## 6. Review trigger

Re-assess on or before the allowlist expiry (**2026-10-26**), or immediately if any of these change:

- a patched `golang.org/x/crypto` release appears (then prefer the bump and delete this exception);
- the `virtengine/cosmos-sdk` fork regresses `crypto/armor.go` to `x/crypto/openpgp`;
- first-party code imports `golang.org/x/crypto/openpgp` for anything beyond armor framing —
  and armor framing itself must stay on the ProtonMail fork;
- govulncheck promotes this advisory from Module Results to Symbol Results with a trace into
  first-party code.

## 7. Reproducing this assessment

```bash
# 1. Confirm no first-party import
grep -rn 'golang.org/x/crypto/openpgp' --include='*.go' . | grep -v vendor   # no output
go mod why golang.org/x/crypto/openpgp   # "main module does not need package ..."

# 2. Confirm nothing in the build links it
go list -deps ./... | grep 'golang.org/x/crypto/openpgp'   # no output

# 3. Confirm the fork migration (the reason step 2 is empty)
grep -n 'openpgp' "$(go env GOMODCACHE)"/github.com/virtengine/cosmos-sdk@v0.53.4-virtengine.2/crypto/armor.go
# -> "github.com/ProtonMail/go-crypto/openpgp/armor"

# 4. Confirm the advisory is module-level only with no fix
govulncheck -show verbose ./... 2>&1 | grep -A 8 'GO-2026-5932'   # Module Results, Fixed in: N/A
curl -sS https://vuln.go.dev/ID/GO-2026-5932.json   # single {introduced: 0}, no fixed event
```

If step 1 or 2 ever produces output, this assessment is void and the exception must be withdrawn.
