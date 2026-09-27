# Supply Chain Security Policy

This document describes the controls that are actually enforced by `.github/workflows/security.yaml`, `.github/workflows/supply-chain.yaml`, `.github/workflows/license-compliance.yaml`, and `.github/workflows/pr-security-check.yaml`. It intentionally documents the current bar only and avoids aspirational claims.

## Effective Controls

- The four release-security workflows are self-validating. Each one runs `.github/scripts/validate_security_policies.py` plus scoped unit, integration, and e2e tests before any scanner result is trusted.
- The targeted workflows reject `@latest`, `continue-on-error: true`, and silent `|| true` bypasses. Scanner failures are treated as gate failures rather than advisory warnings.
- The targeted workflows use GitHub-native identity only: `GITHUB_TOKEN` for GitHub APIs and OIDC keyless signing for Sigstore and provenance jobs. They do not consume long-lived signing or publish secrets.
- Secret scanning uses `.gitleaks.toml` with narrow fixture-only allowlists. Markdown, docs, and public source surfaces remain scanned instead of being globally exempted.
- Vulnerability exceptions are governed by `.vulnerability-allowlist.yaml`. The current active exception count is `0`, and the file does not auto-suppress scanner findings.

## Scanner And Tool Versions

The enforced workflows pin the toolchain they install:

- `govulncheck` `v1.1.4`
- `gosec` `v2.25.0`
- `gitleaks` `v8.30.1`
- `pip-audit` `2.10.0`
- `pip-licenses` `5.5.5`
- `go-licenses` `v1.6.0`
- `license-checker-rseidelsohn` `4.4.2`
- `syft` `v1.42.4`
- `trivy` `v0.69.3`
- `cosign` `v3.0.6`
- `slsa-verifier` `v2.7.1`
- `actionlint` `v1.7.12`

## Release-Critical Tag Bar

For `v*` tags, `.github/workflows/supply-chain.yaml` is the release-critical verification gate. A tagged build is not considered publishable unless all of the following succeed:

1. Dependency verification:
   - `go mod verify`
   - `go mod tidy` and `go mod vendor` must be no-diff
   - `pnpm install --frozen-lockfile --ignore-scripts` must hold for the root workspace and `sdk/portal`
   - `npm ci --ignore-scripts` must hold for `sdk/ts`
2. Lockfile integrity:
   - required lockfiles must exist
   - pull requests cannot change a lockfile without the owning manifest
3. Attack detection:
   - `scripts/supply-chain/detect-supply-chain-attacks.sh --all --json`
4. Dependency risk assessment:
   - `go run ./scripts/supply-chain/assess-dependencies.go --report --json`
   - high-risk or below-threshold packages fail the workflow
5. SBOM generation:
   - `scripts/supply-chain/generate-sbom.sh --format all`
   - CycloneDX, SPDX, Syft, and Go-module inventories are uploaded with SHA-256 sidecars
6. Deterministic release-subject build:
   - the `virtengine` binary is built twice with `-trimpath -buildvcs=false`
   - the SHA-256 output must be identical across both runs
7. Keyless signing:
   - `cosign sign-blob` signs the binary, checksum file, and each SBOM JSON file
8. Signature verification:
   - `cosign verify-blob` must succeed for every signed artifact
   - verification is bound to the exact GitHub Actions workflow identity for `supply-chain.yaml`
9. Provenance generation:
   - SLSA provenance is generated from the release subjects in the same tag workflow
10. Provenance verification:
    - `slsa-verifier verify-artifact` must succeed for the binary, checksum file, and SBOM JSON artifacts against the emitted provenance

## Security And PR Bar

`.github/workflows/security.yaml` is the broad scheduled and post-merge gate. It runs:

- CodeQL for Go
- `govulncheck`
- `gosec`
- `pip-audit` across repository Python requirement files
- `pnpm audit` for the root workspace and `sdk/portal`
- `npm audit` for `sdk/ts`
- Trivy scans for the release-critical container images
- SBOM generation using the same pinned Syft flow used by the tag workflow
- full-repo gitleaks with SARIF upload

`.github/workflows/pr-security-check.yaml` is the pull-request gate. It keeps the fast path targeted to changed files while remaining fail-closed:

- policy validation always runs
- gitleaks always runs on the PR diff
- `govulncheck` runs when Go surfaces change
- `gosec` runs against changed non-test Go source files
- dependency review runs when dependency surfaces change
- security-focused `golangci-lint` runs when security-sensitive packages change

## License And Exception Policy

`.github/workflows/license-compliance.yaml` is the enforced license gate. It blocks disallowed license families across:

- Go modules
- the root pnpm workspace packages
- `sdk/portal`
- `sdk/ts`
- Python requirements under `ml/**`

The blocked families are:

- `GPL`
- `AGPL`
- `LGPL`
- `SSPL`
- `BUSL`

Allowed-license evidence is uploaded as workflow artifacts together with an SPDX SBOM.

`.vulnerability-allowlist.yaml` is governance input, not a suppression switch. Any future exception entry must include:

- identifier
- affected package or image
- reviewer
- review date
- expiry date
- issue or advisory reference
- compensating controls

Expired entries are invalid, and the declared `active_exception_count` must match the file contents.

## Local Validation

The local validation path for this policy surface is:

```bash
python .github/scripts/validate_security_policies.py
python -m unittest discover -s .github/tests -p "test_security_policy*.py"
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 \
  .github/workflows/security.yaml \
  .github/workflows/supply-chain.yaml \
  .github/workflows/license-compliance.yaml \
  .github/workflows/pr-security-check.yaml
bash scripts/supply-chain/detect-supply-chain-attacks.sh --all --json
```

To run the local secret scan with the enforced config:

```bash
gitleaks dir . --config .gitleaks.toml --redact
```

## Local Make Targets

The controls above run in CI. The same underlying scripts are exposed as root `make` targets.
Every target in this section was confirmed present from the repository root with
`make -n <target>`; the recipes below are transcribed from `make/supply-chain.mk` and
`make/mod.mk`. The build system prints the same list with `make help-supply-chain`.

These targets are local conveniences layered over the scripts. Where a target is not itself
enforced, that is stated below — the enforced gate is always the workflow in the section above.

### Verification

| Target | Recipe |
| --- | --- |
| `make supply-chain-verify` | `scripts/supply-chain/verify-dependencies.sh --all` |
| `make supply-chain-detect` | `scripts/supply-chain/detect-supply-chain-attacks.sh --all` |
| `make supply-chain-risk` | `go run assess-dependencies.go` from `scripts/supply-chain/` |
| `make supply-chain-risk-report` | as above with `--report --json`, then `--report`; JSON is written to `.cache/risk-assessment.json` |
| `make supply-chain-audit` | `supply-chain-verify` + `supply-chain-detect` + `supply-chain-risk` + `sbom` |

`vuln-check` is not part of `supply-chain-audit`; run it separately.

### SBOM

| Target | Recipe |
| --- | --- |
| `make sbom` | `scripts/supply-chain/generate-sbom.sh --format all` |
| `make sbom-verify` | `sbom`, then `generate-sbom.sh --format cyclonedx --verify` |

`--format all` emits CycloneDX, SPDX, Syft-native, and Go-module inventories. The script's
default output directory is `.cache/sbom`.

Two gaps between these targets and the enforced CI bar:

- `make sbom` does not pass `--output`. The tag workflow runs
  `generate-sbom.sh --format all --output .cache/supply-chain/sbom`
  (`.github/workflows/supply-chain.yaml:301`), so local and CI SBOMs land in different
  directories. Do not assume a local `make sbom` reproduces the release layout.
- `make sbom-verify` is not a gate. Its verifier returns early with a warning when `grype` is
  not installed, and its `grype` invocation is suffixed `|| true`; it prints a severity summary
  rather than failing the target. The enforced vulnerability check is the tag workflow.

### Dependencies

| Target | Recipe |
| --- | --- |
| `make deps-verify` | `go mod verify` |
| `make deps-tidy` | `go mod tidy` |
| `make deps-vendor` | `deps-tidy`, then `go mod vendor` |
| `make deps-update-check` | `go list -u -m all`, filtered to out-of-date modules |

### Vulnerabilities

| Target | Recipe |
| --- | --- |
| `make vuln-check` | installs `govulncheck@latest` if absent, then `govulncheck ./...` |
| `make vuln-check-json` | same, with `-format json` into `.cache/govulncheck-report.json` |

Both targets install `govulncheck@latest` when it is not on `PATH`, while the security and
PR workflows pin `govulncheck` `v1.1.4`
(`.github/workflows/security.yaml:23`, `.github/workflows/pr-security-check.yaml:24`). Local
results can therefore differ from the enforced scanner. Compare against the pinned version
before treating a local clean run as evidence.

### Signing

| Target | Recipe |
| --- | --- |
| `make sign-artifact ARTIFACT=<path>` | `cosign sign-blob`, writing `<path>.sig` and `<path>.pem` |
| `make verify-signature ARTIFACT=<path>` | `cosign verify-blob` against `<path>.sig` and `<path>.pem` |

Both targets declare `ARTIFACT` as required and stop with an error when it is unset. Both also
require `cosign` on `PATH` and fail if it is missing.

`make verify-signature` cannot validate artifacts produced by the tag workflow as written,
because the local and CI conventions differ on both the certificate filename and the identity:

- The target reads the certificate from `<path>.pem`; the workflow writes
  `${artifact}.sig.cert` (`.github/workflows/supply-chain.yaml:405`).
- The target verifies `--certificate-identity-regexp ".*@virtengine.com"`; the workflow
  verifies the exact workflow URL identity
  `https://github.com/<repo>/.github/workflows/supply-chain.yaml@<ref>`
  (`.github/workflows/supply-chain.yaml:445-455`). A workflow-URL identity does not match that
  regexp, and a local keyless identity will not match the workflow URL.

Use `make verify-signature` for locally-signed artifacts only. Verifying a release means either
running the workflow's own verification step or reproducing its
`--certificate-identity` / `--certificate` arguments explicitly, as in
[Release Artifact Verification](#release-artifact-verification) below.

## Release Artifact Verification

For a local verification of release-critical artifacts produced by the tag workflow, use:

```bash
cosign verify-blob \
  --signature release-subjects/virtengine.sig \
  --certificate release-subjects/virtengine.sig.cert \
  --certificate-identity "https://github.com/virtengine/virtengine/.github/workflows/supply-chain.yaml@refs/tags/vX.Y.Z" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  release-subjects/virtengine

slsa-verifier verify-artifact \
  --provenance-path provenance.intoto.jsonl \
  --source-uri github.com/virtengine/virtengine \
  release-subjects/virtengine
```

The same pattern applies to `release-subjects/virtengine.sha256` and the SBOM JSON artifacts. A release that cannot satisfy both the signature and provenance verification steps does not meet the current production bar.
