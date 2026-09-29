# GO-2022-0646 — `aws-sdk-go` S3 Crypto SDK CBC padding oracle: reachability assessment and accepted risk

**Advisory:** GO-2022-0646 · CVE-2020-8911 · GHSA-f5pg-7wfw-84q9
**Severity:** Moderate (`CVSS:3.1/AV:L/AC:H/PR:L/UI:N/S:C/C:H/I:N/A:N`, CWE-327) — attack vector
**Local**, requires write access to the target bucket and an endpoint that reveals decryption
failures
**Affected module:** `github.com/aws/aws-sdk-go` — the Go vulnerability DB registers the advisory
against the **whole module**; the vulnerable package it names is
`github.com/aws/aws-sdk-go/service/s3/s3crypto` with symbols `NewDecryptionClient` and
`NewEncryptionClient`
**Patched versions:** none in the Go vulnerability DB — the OSV range is a single
`{introduced: "0"}` event with **no `fixed` event** (GitHub's advisory database records the same
CVE as patched in `1.34.0`; this repo pins `v1.49.0`. See §3.4.)
**Assessed on:** `develop` @ `336ce39b`, 2026-09-29
**Toolchain:** Go 1.26.8, `govulncheck v1.1.4`, `GOWORK=off`, module cache `aws-sdk-go@v1.49.0`,
Go vulnerability DB read live on 2026-09-29
**Verdict:** accepted, time-boxed residual risk. Same affected package and symbols as
[GO-2022-0635](GO-2022-0635-S3CRYPTO-ASSESSMENT.md); the vulnerable package is compiled into
nothing, and no binary built from `./cmd/...` links the module. Tracked in
[issue #1079](https://github.com/virtengine/virtengine/issues/1079) and
`.vulnerability-allowlist.yaml`.

**RESOLVED 2026-09-29 — this exception has been deleted, not renewed.** The v1 module is no longer
a `require` of the root module and the two first-party signers now use the `aws-sdk-go-v2` SigV4
signer, so `govulncheck` no longer loads this advisory's OSV entry at all. Re-measured after the
change: three findings before (this one, GO-2022-0635, and the unrelated still-allowlisted
GO-2026-5932), one after (GO-2026-5932 only). Details and the A/B measurement are in §7.

---

## 1. What the advisory says

The v1 AWS S3 encryption client will encrypt with AES-CBC **without computing a MAC** over the
ciphertext. An attacker with write access to the bucket, plus an endpoint that reveals only whether
a decryption succeeded, can use CBC's malleability as a padding oracle and reconstruct plaintext in
about `128 × len(plaintext)` queries. The advisory's own note is that the fix is a new API
(`aws-sdk-go-v2`'s S3 encryption client with `KMS+context` key wrapping), not a patch to this
package — old files stay vulnerable until re-encrypted.

This is the CBC half of the same Google ISE finding as GO-2022-0635 (in-band key negotiation). Both
advisories name the identical package and symbols:

```json
"ecosystem_specific": {"imports": [{
  "path": "github.com/aws/aws-sdk-go/service/s3/s3crypto",
  "symbols": ["NewDecryptionClient", "NewEncryptionClient"]}]}
```

A version bump cannot clear either one: the Go vulnerability DB range for GO-2022-0646 is
`{"type": "SEMVER", "events": [{"introduced": "0"}]}` — no `fixed` event on any version of the v1
line.

## 2. Dependency route

`go.mod:20` declares `github.com/aws/aws-sdk-go v1.49.0` as a **direct** require of the root
module. Two first-party packages import it, both only for SigV4 request signing:

```
pkg/verification/email/providers.go:19-22    aws, aws/credentials, aws/session, aws/signer/v4
pkg/verification/sms/providers.go:18-21      aws, aws/credentials, aws/session, aws/signer/v4
```

(`go mod why -m github.com/aws/aws-sdk-go` prints only the `email` chain — it prints the shortest
path, not all of them. Three `_test.go` files also import the SDK's EC2/EKS service packages:
`infra/tests/infra_test.go`, `infra/terraform/tests/{eks,vpc}_test.go`.)

The earlier revision of the allowlist entry claimed the module arrived only through
`infra/terraform/tests` and that production S3 code used `aws-sdk-go-v2`; both were wrong and are
corrected in §7 of the sibling assessment.

## 3. Why the vulnerable symbols are on no reachable path

### 3.1 The S3 service packages are in no build in this repository

```
$ go list -deps ./... | grep -E '^github.com/aws/aws-sdk-go/service/s3'
(no output)
```

`grep -c s3crypto` over the dependency closure of each importer returns `0` for both:

```
$ go list -deps ./pkg/verification/email | grep -c s3crypto     -> 0
$ go list -deps ./pkg/verification/sms   | grep -c s3crypto     -> 0
```

There is no S3 client and no S3 encryption client in the compiled graph, so neither
`NewDecryptionClient` nor `NewEncryptionClient` exists in any package this repository builds.

### 3.2 The importers sign HTTP requests; they never touch the S3 encryption client

```go
// pkg/verification/sms/providers.go:882-900 (email equivalent at providers.go:852-870)
cfg := aws.NewConfig().WithRegion(region).WithHTTPClient(httpClient)
cfg = cfg.WithCredentials(credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""))
sess, err := session.NewSession(cfg)
return sess.Config.Credentials, v4.NewSigner(sess.Config.Credentials), nil
```

The email provider POSTs to `https://email.<region>.amazonaws.com/`; the SMS provider POSTs to
`https://sns.<region>.amazonaws.com/`. Both hand-roll the request and sign it with
`aws/signer/v4`. No `service/...` package of this module is imported by first-party non-test code.

### 3.3 No binary links the module

```
$ go list -deps ./cmd/... | grep -cE '^github.com/aws/aws-sdk-go/'   -> 0
```

Nothing imports `pkg/verification/email` or `pkg/verification/sms` (`grep -rlF` over all `*.go`
returns nothing for either import path), so the two leaf packages are in `go list ./...` — and
therefore in the scan graph — but in none of the shipped binaries, images, or daemons. A CBC
padding oracle needs a decrypting endpoint; there is no such endpoint in this codebase, and no
compiled code that could host one.

### 3.4 `govulncheck` reports it at module granularity, with zero affected vulnerabilities

```
$ govulncheck -show verbose ./pkg/verification/...
=== Module Results ===
Vulnerability #2: GO-2022-0646
    CBC padding oracle issue in AWS S3 Crypto SDK for golang in github.com/aws/aws-sdk-go
  Module: github.com/aws/aws-sdk-go
    Found in: github.com/aws/aws-sdk-go@v1.49.0
    Fixed in: N/A
...
Your code is affected by 0 vulnerabilities.
This scan also found 0 vulnerabilities in packages you import and 3
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
```

`Fixed in: N/A` is §1's point. The **Module Results** placement is the reason the finding has to be
dispositioned rather than fixed: the module is in the requirement graph, and `govulncheck` cannot
see a single reachable path, so it reports the module instead of a symbol. The sharded
`Go Vulnerability Scan` job shows the same entry on the `platform` shard.

The record conflict is left visible on purpose: GitHub's advisory database marks this CVE patched
in `1.34.0` (this repo pins `v1.49.0`), while the Go vulnerability DB that `govulncheck` consumes
carries no `fixed` event. The measurements in §3.1–§3.3, not either version claim, are the basis of
this assessment.

## 4. Decision

**Accepted residual risk, time-boxed.** Recorded in `.vulnerability-allowlist.yaml` with a 30-day
expiry, under the repository policy (`max_allowlist_age_days: 30`,
`require_compensating_controls: true`, `require_issue_reference: true`), tracked in
[issue #1079](https://github.com/virtengine/virtengine/issues/1079).

Compensating controls:

1. `service/s3/s3crypto` is absent from `go list -deps ./...`; the advisory's named symbols are
   compiled into nothing.
2. The module is required only for SigV4 request signing by two leaf packages, neither of which is
   imported by anything else; `go list -deps ./cmd/...` links zero v1 packages.
3. No S3 bucket, no S3 encryption client, no decrypting endpoint exists in this codebase — the
   oracle the attack needs has no host here.
4. Time-boxed to 30 days; `validate_security_policies.py` fails `Policy Validation` when the entry
   expires or exceeds the age limit.
5. Permanent fix tracked in #1079: swap the v1 `aws`/`credentials`/`session`/`signer/v4` usage in
   `pkg/verification/email` and `pkg/verification/sms` for the `aws-sdk-go-v2` equivalents
   (`github.com/aws/aws-sdk-go-v2` is already a dependency), then drop the v1 require and delete
   both s3crypto exceptions instead of renewing them.

## 5. Review trigger

Re-assess on or before the allowlist expiry (**2026-10-28**), or immediately if:

- first-party code starts importing any `github.com/aws/aws-sdk-go/service/...` package, or either
  verification package gains an importer;
- the Go vulnerability DB publishes a `fixed` event for GO-2022-0646;
- the v1 require is dropped (then this assessment is void and the entry is deleted).

## 6. Reproducing this assessment

```bash
export GOWORK=off

grep -n 'github.com/aws/aws-sdk-go ' go.mod
grep -rn 'github.com/aws/aws-sdk-go/' --include=*.go . | grep -v /vendor/
go list -deps ./... | grep -E '^github.com/aws/aws-sdk-go/service/s3'          # no output
go list -deps ./pkg/verification/email | grep -c s3crypto                      # 0
go list -deps ./pkg/verification/sms   | grep -c s3crypto                      # 0
go list -deps ./cmd/... | grep -cE '^github.com/aws/aws-sdk-go/'               # 0
govulncheck -show verbose ./pkg/verification/...
```

Any output from the `service/s3` or `s3crypto` greps, or a non-zero count from the `./cmd/...`
grep, voids this assessment and the exception must be withdrawn in favour of the `aws-sdk-go-v2`
port.

## 7. Resolution (2026-09-29)

The permanent fix recorded against issue #1079 landed as a **signer swap** in
`pkg/verification/email/providers.go` and `pkg/verification/sms/providers.go`: both now use
`aws-sdk-go-v2/aws/signer/v4` (`Signer.SignHTTP`) with credentials from
`aws-sdk-go-v2/config.LoadDefaultConfig` or `credentials.NewStaticCredentialsProvider`, and
`github.com/aws/aws-sdk-go v1` is no longer a require of the root module.

Re-measured on the change branch (base `develop` @ `d4f6c4f9`), go1.26.8, `GOWORK=off`:

- `govulncheck -format json ./pkg/verification/email ./pkg/verification/sms` reports **this
  advisory and GO-2022-0635 before** the change and **neither after it** — the remaining finding
  is the unrelated, still-allowlisted GO-2026-5932. The before-number came from running the same
  command against `develop` @ `d4f6c4f9` in a separate worktree.
- Both allowlist entries were **deleted**, `policy.active_exception_count` went 5 → 3, and
  `python .github/scripts/validate_security_policies.py` exits 0.
- Wire compatibility was checked, not assumed: signing a fixed request with the v1 and v2 signers
  and recomputing SigV4 independently reproduces both signatures exactly. The one difference is
  that v2 also signs `content-length` (set by `net/http` from `Request.ContentLength`).

`aws-sdk-go` v1 remains a transitive requirement of `wasmd` / `ibc-go/v10` / `go-kit`, so it is
still listed by `go list -m all`; `govulncheck` stops loading its advisories once no first-party
package imports it, which is now the case. The `infra/*/tests` modules keep their own v1 require
for `service/ec2` and `service/eks` and are outside the root scan.
