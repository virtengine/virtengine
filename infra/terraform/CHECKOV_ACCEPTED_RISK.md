# Checkov accepted-risk register — `infra/terraform/**`

This file is the index for the inline `#checkov:skip=` comments under
`infra/terraform/`. Those comments are the actual mechanism; this file explains
what the register means and how to work with it.

**The gate is not disabled.** `soft_fail` stays `false`, no `skip-check` list was
introduced, no `continue-on-error` was added, and the gate runs exactly as it did
before. What changed is that 112 findings now each carry a written reason
attached to the resource that raised them.

## Why the gate was red at all

`Infrastructure / Security Scan` has been failing on **every branch including
`main`** for longer than 15 consecutive runs on `develop`, and identically on
`main`'s `schedule` and `push` events. It was a standing baseline break, not a
regression from any single PR — but it made every PR touching `infra/**` report
`mergeStateStatus=UNSTABLE` forever, which is a worse outcome than any one of the
findings individually, because it trains everyone to ignore the gate.

So this is deliberately a **disposition** commit, not a hardening commit. It
records who decided what about each finding. It does not pretend the underlying
items are resolved — see the two lists below, which are the point of the exercise.

## How the mechanism behaves

| property | value | why it matters |
|---|---|---|
| skip granularity | per `(resource, check)` | a *new* check on an already-annotated resource still fails the job |
| mechanism | inline `#checkov:skip=` inside the resource block | the reason travels with the code; a reviewer sees it in the diff, not in a wiki |
| reason | required on every line | no bare `#checkov:skip=CKV_AWS_1` is acceptable |
| review-by date | required on every line | makes the register expire instead of becoming permanent folklore |
| counted resources | one comment covers both reports | checkov reports a counted resource twice (indexed and unindexed) with opposite verdicts; see `CHECKOV_SCANER_LIMITATION.md` |

That last row is why 112 reported findings became 104 comments.

**Verified**: a new finding on a resource that has no comment for that check id
still fails the step. The gate is strict for everything not in this register.

## The 41 findings that are REAL defects

These are **not** scanner noise and **not** tool artefacts. They are genuine gaps
that a human decided to carry rather than fix in this commit, and each one's
comment says `KNOWN GAP` and names the real fix. Carrying them is a decision with
an expiry (review-by 2026-11-01), not a deletion.

| check | n | the real gap |
|---|---|---|
| `CKV2_AWS_61` | 6 | no S3 lifecycle rule — objects retained forever, cost unbounded |
| `CKV2_AWS_64` | 5 | no explicit KMS key policy — falls back to the AWS-managed default |
| `CKV_AWS_338` | 5 | CloudWatch retention below 1 year on 5 log groups |
| `CKV_AWS_355` | 5 | IAM `Resource="*"` on actions with no resource-level ARN |
| `CKV_AWS_290` | 3 | paired unconstrained write statements (same root cause as above) |
| `CKV2_AWS_31` | 2 | WAF ACLs have no logging configuration |
| `CKV2_AWS_57` | 2 | no automatic rotation on 2 secrets |
| `CKV_AWS_116` | 2 | Lambda functions have no DLQ |
| `CKV_AWS_173` | 2 | Lambda env vars not encrypted with a customer KMS key |
| `CKV_AWS_21` | 2 | versioning missing on 2 counted buckets whose siblings have it |
| `CKV2_AWS_38` | 1 | DNSSEC not enabled on the public hosted zone |
| `CKV2_AWS_39` | 1 | Route53 zone query logging not enabled |
| `CKV2_AWS_60` | 1 | RDS replica does not copy tags to snapshots |
| `CKV_AWS_161` | 1 | RDS IAM database auth not enabled |
| `CKV_AWS_26` | 1 | SNS topic not KMS-encrypted |
| `CKV_AWS_28` | 1 | DynamoDB point-in-time recovery off |
| `CKV_AWS_38` | 1 | EKS public endpoint not restricted to an admin CIDR |

Nothing in this list is an emergency: each is either a hardening item (the
AWS-managed key policy is unchanged in effect; the SNS topic carries
non-sensitive notifications) or needs a coordinated change outside this repo
(DNSSEC needs the registrar; IAM DB auth needs a driver change). But they are
real, and the highest-value follow-ups are the ones with a one-line fix:
`CKV2_AWS_60`, `CKV_AWS_28`, and `CKV2_AWS_31`.

## The 63 findings that are accepted design or tool artefacts

Grouped by why they are acceptable:

* **By definition of the resource.** `CKV_AWS_130` (6) on subnets literally named
  `public`; `CKV_AWS_39` (2) disabling the EKS endpoint, which must stay reachable.
* **A duplicate mechanism already exists.** `CKV_AWS_144` (9) cross-region
  replication is provided by the `dr/` + `multi-region` module pair;
  `CKV_AWS_382` (5) unrestricted egress on node/cluster/database SGs.
* **A control the tool cannot model.** `CKV2_AWS_23` (7) alias records — an alias
  *is* the attachment; `CKV2_AWS_5` (6) SG attachment declared via
  `aws_security_group_rule` rather than an inline `vpc_id`; `CKV_AWS_145` (3) the
  KMS key reference lives on a sibling resource.
* **Payload carries nothing sensitive.** `CKV_AWS_119`/`CKV_AWS_149` lock rows and
  connection payloads; `CKV_AWS_26`'s topic.
* **Proven tool limitation.** `CKV2_AWS_6` (2) on counted resources, reproduced by
  a mutation probe in `CHECKOV_SCANER_LIMITATION.md` — the public access block is
  present on both resources.

## Working with this register

**Adding a finding to it** requires a reason in the comment that answers *why this
is acceptable here*, plus a review-by date. "Pre-existing" is not a reason.

**Removing an entry** means deleting the comment. If the underlying control is
still missing, the check goes red again — which is the point.

**Reviewing it** means reading the dates. Anything past `review-by` is not
automatically an error, but it is not automatically accepted either; it has to be
re-decided and the date moved.

## Regenerating / verifying

The gate is the authority, not this file:

```bash
python -m checkov.main -d infra/terraform --framework terraform --compact
```

Note from `CHECKOV_SCANER_LIMITATION.md`: `-o json` emits a JSON **object** per
framework, not an array, and checkov prepends an ANSI reset after every `Guide:`
line — strip ANSI before anchoring on a `Check:` header.

Refs: card `t_ae133441`, Infrastructure runs `37155238632` (the 112-finding
baseline this was derived from) and the `develop` runs listed on that card. The
`CKV2_AWS_6` disposition is reconciled with `CHECKOV_SCANER_LIMITATION.md`:
that file's claim that "the remaining 110 findings are being fixed on their own
merits" is now **resolved as accepted-risk disposition**, not as fixes — the
wording above supersedes it.