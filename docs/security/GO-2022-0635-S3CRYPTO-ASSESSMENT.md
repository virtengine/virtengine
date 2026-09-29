# GO-2022-0635 — `aws-sdk-go` S3 Crypto SDK in-band key negotiation: reachability assessment and accepted risk

**Advisory:** GO-2022-0635 · CVE-2020-8912 · GHSA-7f33-f4f5-xwgw
**Severity:** Low (`CVSS:3.1/AV:L/AC:H/PR:L/UI:N/S:U/C:L/I:N/A:N`) — attack vector **Local**, requires
write access to the target bucket and an endpoint that reveals decryption failures
**Affected module:** `github.com/aws/aws-sdk-go` — the Go vulnerability DB registers the advisory
against the **whole module**; the vulnerable package it names is
`github.com/aws/aws-sdk-go/service/s3/s3crypto` with symbols `NewDecryptionClient` and
`NewEncryptionClient`
**Patched versions:** none in the Go vulnerability DB — the OSV range is a single
`{introduced: "0"}` event with **no `fixed` event**, so `govulncheck` cannot clear it with any
version bump (GitHub's advisory database records the same CVE as patched in `1.34.0`; this repo
pins `v1.49.0`. See §3.5 — the discrepancy is not load-bearing, because the package is not
compiled into anything here.)
**Assessed on:** `develop` @ `336ce39b`, 2026-09-29
**Toolchain:** Go 1.26.8, `govulncheck v1.1.4`, `GOWORK=off`, module cache `aws-sdk-go@v1.49.0`,
Go vulnerability DB read live on 2026-09-29
**Verdict:** accepted, time-boxed residual risk. No first-party code imports
`service/s3/s3crypto`, no package built from this module contains that package, and no binary built
from `./cmd/...` links the `aws-sdk-go` v1 module at all. Tracked in
[issue #1079](https://github.com/virtengine/virtengine/issues/1079) and
`.vulnerability-allowlist.yaml`. Sibling advisory on the same package and symbols:
[GO-2022-0646](GO-2022-0646-S3CRYPTO-ASSESSMENT.md).

**RESOLVED 2026-09-29 — this exception has been deleted, not renewed.** The permanent fix this
assessment named as its exit condition landed: `github.com/aws/aws-sdk-go` v1 is no longer a
`require` of the root module, and `pkg/verification/email/providers.go` /
`pkg/verification/sms/providers.go` sign their SES and SNS query requests with
`github.com/aws/aws-sdk-go-v2/aws/signer/v4` (`aws-sdk-go-v2/config` and
`aws-sdk-go-v2/credentials` supply the credential chain). Re-verified with `govulncheck v1.1.4` on
go1.26.8 after the change: the two module-granularity findings are gone from the report. Details
and the A/B measurement are in §7.

---

## 1. What the advisory says

The v1 AWS S3 encryption client does not authenticate the algorithm parameters of a data
encryption key ("in-band key negotiation"). An attacker with write access to the bucket can
rewrite an object's `X-Amz-Meta-X-Amz-Cek-Alg` header and switch the content cipher — for example
from AES-GCM to AES-CTR. Combined with an endpoint (oracle) that reveals only whether decryption
succeeded, that lets the attacker recover the AES-GCM authentication key algebraically and then
reconstruct plaintext. Both CVEs come from the same Google ISE finding (Sophie Schmieg) and share
the same affected package and symbols; GO-2022-0646 is the CBC padding-oracle half.

This is not a "bump the dependency" finding in this repository:

- the Go vulnerability DB's OSV range for `GO-2022-0635` is
  `{"type": "SEMVER", "events": [{"introduced": "0"}]}` — there is **no `fixed` event**, for any
  version, in any major line, so no `go get` can make `govulncheck` stop reporting it;
- the remediation AWS shipped is a **new API** (`aws-sdk-go-v2`'s S3 encryption client, using
  `KMS+context` key wrapping), not a patch to the v1 package;
- `service/s3/s3crypto` at `v1.49.0` still contains `NewDecryptionClient`
  (`decryption_client.go:52`) and `NewEncryptionClient` (`encryption_client.go:60`).

## 2. Dependency route

`go.mod:20` declares it a **direct** require of the root module:

```
github.com/aws/aws-sdk-go v1.49.0
```

`go mod why -m github.com/aws/aws-sdk-go` prints one chain (it prints the shortest chain, not all
of them):

```
# github.com/aws/aws-sdk-go
github.com/virtengine/virtengine/pkg/verification/email
github.com/aws/aws-sdk-go/aws
```

**Correction (2026-09-29).** An earlier revision of the allowlist entry — and the 2026-09-28
audit-log note — said the module "enters only through `infra/terraform/tests`" and that
production S3 operations use `aws-sdk-go-v2`. Both statements were wrong, and a third, subtler
one replaced them: that the module "reaches the scan only through `pkg/verification/email`". A
repository-wide search of `*.go` finds **two** first-party packages importing the v1 module:

```
pkg/verification/email/providers.go:19-22    aws, aws/credentials, aws/session, aws/signer/v4
pkg/verification/sms/providers.go:18-21      aws, aws/credentials, aws/session, aws/signer/v4
```

plus three `_test.go` files that import the SDK's EC2/EKS *service* packages
(`infra/tests/infra_test.go`, `infra/terraform/tests/{eks,vpc}_test.go`). `go mod why -m` reports
only the `email` chain because it prints the shortest path; the `sms` chain is the same shape.

## 3. Why the vulnerable symbols are on no reachable path

### 3.1 The S3 service packages are not in the build at all

```
$ go list -deps ./... | grep -E '^github.com/aws/aws-sdk-go/service/s3'
(no output)
```

Not `s3crypto`, and not even `service/s3`. The 41 `github.com/aws/aws-sdk-go/*` packages that
`go list -deps ./...` (3015 packages) does contain are the config/credentials/session/signer and
SSO/STS machinery pulled in by the two packages below. There is no S3 client, no S3 encryption
client, and no code that could call `NewDecryptionClient`/`NewEncryptionClient`.

### 3.2 The two importers use only the SigV4 signer, not any service package

Both `pkg/verification/email` and `pkg/verification/sms` hand-roll their HTTP requests to
`https://email.<region>.amazonaws.com/` and `https://sns.<region>.amazonaws.com/` and sign them
with `aws/signer/v4`:

```go
// pkg/verification/email/providers.go:852-870  (pkg/verification/sms/providers.go:882-900 is the same)
cfg := aws.NewConfig().WithRegion(region).WithHTTPClient(httpClient)
cfg = cfg.WithCredentials(credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""))
sess, err := session.NewSession(cfg)
return sess.Config.Credentials, v4.NewSigner(sess.Config.Credentials), nil
```

That is the entire reason the module is required: request signing. The vulnerable package is not
imported by either, and no `service/...` package from this module is imported anywhere in
first-party code outside the three test files named in §2.

### 3.3 The vulnerable package is compiled into nothing

```
$ go list -deps ./pkg/verification/email | grep -c s3crypto     -> 0
$ go list -deps ./pkg/verification/sms   | grep -c s3crypto     -> 0
```

Consequently the advisory's named symbols do not exist in any compiled package.

### 3.4 No binary links the module

```
$ go list -deps ./cmd/... | grep -cE '^github.com/aws/aws-sdk-go/'   -> 0
```

Nothing imports `pkg/verification/email` or `pkg/verification/sms` either (`grep -rlF` over all
`*.go` returns nothing for both import paths, and `go list -deps ./cmd/...` contains neither), so
they are leaf packages: they are in `go list ./...` and therefore in the scan graph, but they are
not part of any shipped binary, image, or daemon.

### 3.5 `govulncheck`'s own output agrees, and says why the finding persists

```
$ govulncheck -show verbose ./pkg/verification/...
=== Module Results ===
Vulnerability #2: GO-2022-0646 ... Module: github.com/aws/aws-sdk-go
    Found in: github.com/aws/aws-sdk-go@v1.49.0
    Fixed in: N/A
Vulnerability #3: GO-2022-0635 ... Module: github.com/aws/aws-sdk-go
    Found in: github.com/aws/aws-sdk-go@v1.49.0
    Fixed in: N/A

Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 3
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
```

Two things are visible here. `Fixed in: N/A` is the §1 point: version resolution is not available
to us. And the finding is reported under **Module Results** — not symbol results — with **0
affected vulnerabilities**. `govulncheck` reports a module-level entry when the advisory's
declared package is a required module but nothing in the scanned build imports it: the module is
in the *requirement* graph, not in the *call* graph. The same shape appears in the sharded
`Go Vulnerability Scan` job, where the `platform` shard carries GO-2022-0635/0646 as
module-granularity findings.

Note the record conflict deliberately left visible: GitHub's advisory database marks the same CVEs
patched in `1.34.0` and this repo pins `v1.49.0` (past it), while the Go vulnerability DB — the
source `govulncheck` uses — records no `fixed` event at all. The reachability measurements above,
not either version claim, are what this assessment rests on.

## 4. Decision

**Accepted residual risk, time-boxed.** Recorded in `.vulnerability-allowlist.yaml` with a 30-day
expiry (`max_allowlist_age_days: 30`, `require_compensating_controls: true`,
`require_issue_reference: true`), tracked in [issue #1079](https://github.com/virtengine/virtengine/issues/1079).

Compensating controls:

1. `service/s3/s3crypto` — and every other `service/s3*` package of the v1 module — is absent from
   `go list -deps ./...`; the advisory's named symbols are compiled into nothing.
2. The module is required only for SigV4 request signing, used by two leaf packages that nothing
   else imports; `go list -deps ./cmd/...` links zero v1 packages, so no shipped binary contains
   the module at all.
3. The advisory requires write access to an S3 bucket and a decryption-failure oracle. VirtEngine
   holds no S3 bucket in this path and uses no S3 encryption client, so the precondition cannot be
   established from this codebase.
4. Time-boxed to 30 days; `validate_security_policies.py` fails the `Policy Validation` job when
   the entry expires or exceeds the age limit.
5. The permanent fix is tracked: replace the v1 `aws`/`credentials`/`session`/`signer/v4` usage in
   `pkg/verification/email` and `pkg/verification/sms` with the `aws-sdk-go-v2` equivalents
   (`github.com/aws/aws-sdk-go-v2` is already a dependency), then drop the v1 require and delete
   this exception rather than renewing it.

**This entry does not turn `Go Vulnerability Scan` green by itself.**
`.vulnerability-allowlist.yaml` is validated by policy but is not consumed by any scanner in this
repository; the shard gate classifies findings and the allowlist is the governance record for
findings that are dispositioned rather than fixed.

## 5. Review trigger

Re-assess on or before the allowlist expiry (**2026-10-28**), or immediately if any of these change:

- first-party code starts importing any `github.com/aws/aws-sdk-go/service/...` package (S3, SES,
  SNS), or `pkg/verification/email` / `pkg/verification/sms` gains an importer;
- the Go vulnerability DB publishes a `fixed` event for GO-2022-0635 (then prefer the bump and
  delete both this and the GO-2022-0646 entry);
- the v1 require is dropped (then this assessment is void and the entry is deleted).

## 6. Reproducing this assessment

```bash
export GOWORK=off

# 1. It is a direct require of the root module
grep -n 'github.com/aws/aws-sdk-go ' go.mod

# 2. Who actually imports it (expect email + sms + three _test.go files)
grep -rn 'github.com/aws/aws-sdk-go/' --include=*.go . | grep -v /vendor/

# 3. The vulnerable package is in no build
go list -deps ./... | grep -E '^github.com/aws/aws-sdk-go/service/s3'          # no output
go list -deps ./pkg/verification/email | grep -c s3crypto                      # 0
go list -deps ./pkg/verification/sms   | grep -c s3crypto                      # 0

# 4. No shipped binary links the module
go list -deps ./cmd/... | grep -cE '^github.com/aws/aws-sdk-go/'               # 0

# 5. The scanner's shape: module-level, not called
govulncheck -show verbose ./pkg/verification/...
```

If step 3 or step 4 ever produces output, this assessment is void and the exception must be
withdrawn in favour of the `aws-sdk-go-v2` port.

## 7. Resolution (2026-09-29)

The exit condition above was met by a **signer swap**, not an SDK client port: both packages
hand-roll their AWS query-API calls over `net/http` and only ever needed SigV4 signing.

| | before | after |
| --- | --- | --- |
| signer | `aws-sdk-go/aws/signer/v4` — `Signer.Sign(req, body, service, region, t)` | `aws-sdk-go-v2/aws/signer/v4` — `Signer.SignHTTP(ctx, creds, req, payloadHashHex, service, region, t)` |
| credentials | `session.NewSession` + `credentials.NewStaticCredentials` | `aws-sdk-go-v2/config.LoadDefaultConfig` default chain, or `credentials.NewStaticCredentialsProvider` |
| `go.mod` | `github.com/aws/aws-sdk-go v1.49.0` (direct require) | no v1 require at all |

Measured on the change branch (base `develop` @ `d4f6c4f9`), go1.26.8, `GOWORK=off`:

- `grep -n 'aws-sdk-go v1' go.mod` → no match, and
  `go list -deps ./... | grep -c '^github.com/aws/aws-sdk-go/'` → 0.
- `govulncheck -format json ./pkg/verification/email ./pkg/verification/sms`:
  **3 findings before** (this advisory, GO-2022-0646, and the unrelated GO-2026-5932) and
  **1 after** (GO-2026-5932 only — still allowlisted, still a separate residual risk). The
  pre-change number was measured by running the identical command against `develop` @ `d4f6c4f9`
  in a separate worktree, so the A/B is exact rather than inferred.
- `python .github/scripts/filter_allowlist.py <report> .vulnerability-allowlist.yaml` exits 0 with
  the post-change report and 3 with the pre-change report against the same (edited) allowlist —
  i.e. deleting the exception is safe *because* of this change, not independently of it.
- The two signers are wire-compatible. Signing the same request with both at a fixed signing time
  and recomputing SigV4 from first principles (canonical request → string to sign → HMAC chain)
  reproduces both signatures byte for byte. The only difference is the signed-header set: v2 also
  signs `content-length`, whose value `net/http` sets from `Request.ContentLength`; the
  `pkg/verification/{email,sms}` tests now assert the length the signer covered equals the length
  the server received.
- The allowlist entries for GO-2022-0635 and GO-2022-0646 were **deleted** (not renewed),
  `policy.active_exception_count` went 5 → 3, and
  `python .github/scripts/validate_security_policies.py` exits 0.

`aws-sdk-go` v1 is still in the module graph as a requirement of `wasmd`, `ibc-go/v10` and
`go-kit`, so it continues to appear in `go list -m all` — but `govulncheck` no longer loads its
advisories, because no first-party package imports it. That is exactly the condition the
module-granularity finding was always waiting on. The separate `infra/terraform/tests` and
`infra/tests` modules still require v1 for `service/ec2` and `service/eks`; they are distinct
modules, outside the root scan, and import no `s3crypto`.

## 8. Correction log

| date | change |
| --- | --- |
| 2026-09-28 | entry created; claimed the module entered only through `infra/terraform/tests` and that production S3 used `aws-sdk-go-v2` — both wrong |
| 2026-09-29 | root cause re-derived and corrected (direct require, pulled in by `pkg/verification/email`); placeholder `issues/NEW` reference replaced with issue #1079; this assessment written |
| 2026-09-29 | second correction: the module is pulled in by **two** first-party packages (`email` **and** `sms`), and the fix route is a SigV4-signer swap, not an SDK "SES client" port — issue #1079 updated accordingly |
| 2026-09-29 | **resolved**: the v1 require is dropped from `go.mod` and both signers moved to `aws-sdk-go-v2`; `govulncheck` no longer reports this advisory (3 findings → 1 on the same scan). The allowlist entry was deleted, not renewed. See §7 |
