# Supply Chain Security Policy

This document outlines VirtEngine's supply chain security practices, policies, and tools for ensuring the integrity and security of our software dependencies and build processes.

## Table of Contents

- [Overview](#overview)
- [Dependency Management](#dependency-management)
- [SBOM (Software Bill of Materials)](#sbom-software-bill-of-materials)
- [Signed Releases](#signed-releases)
- [Build Provenance](#build-provenance)
- [Reproducible Builds](#reproducible-builds)
- [Vulnerability Scanning](#vulnerability-scanning)
- [Third-Party Risk Assessment](#third-party-risk-assessment)
- [Supply Chain Attack Detection](#supply-chain-attack-detection)
- [Verification Guide](#verification-guide)

## Overview

VirtEngine implements a defense-in-depth approach to supply chain security following industry best practices including:

- **SLSA Level 3** build provenance attestation
- **Sigstore** cosign for artifact signing
- **SBOM** generation in CycloneDX and SPDX formats
- **Automated vulnerability scanning** across all dependency ecosystems
- **Continuous monitoring** for supply chain attacks

## Dependency Management

### Pinning Policy

All dependencies MUST be pinned to exact versions:

#### Go Dependencies

```go
// ✅ GOOD - Exact version
require github.com/cosmos/cosmos-sdk v0.53.4

// ❌ BAD - Version range or latest
require github.com/cosmos/cosmos-sdk v0.53.x
```

**Requirements:**

- All direct dependencies in `go.mod` must use exact semantic versions
- `go.sum` must be committed and verified in CI
- `replace` directives must point to specific commits or tags
- Vendor directory is used for reproducibility: `go mod vendor`

#### Python Dependencies

```txt
# ✅ GOOD - Exact version with hash
tensorflow==2.16.1 --hash=sha256:abc123...

# ❌ BAD - Version range
tensorflow>=2.16.0
```

**Requirements:**

- Use `requirements.txt` with exact versions
- Include package hashes where possible
- Use `pip-compile` from pip-tools for lockfiles

#### npm Dependencies

```json
// ✅ GOOD - package-lock.json with exact versions
// ❌ BAD - package.json ranges without lockfile
```

**Requirements:**

- `package-lock.json` must be committed
- Use `npm ci` instead of `npm install` in CI
- Audit dependencies before adding new packages

### Lockfile Verification

All lockfiles are verified in CI to prevent unauthorized modifications:

| Ecosystem | Lockfile            | Verification Tool         |
| --------- | ------------------- | ------------------------- |
| Go        | `go.sum`            | `go mod verify`           |
| Python    | `requirements.txt`  | Hash verification         |
| npm       | `package-lock.json` | `npm ci --ignore-scripts` |

## SBOM (Software Bill of Materials)

VirtEngine generates SBOMs in multiple formats for each release:

### Formats

| Format    | File             | Use Case                           |
| --------- | ---------------- | ---------------------------------- |
| CycloneDX | `sbom.cdx.json`  | Vulnerability scanning, compliance |
| SPDX      | `sbom.spdx.json` | License compliance, auditing       |

### Generation

SBOMs are automatically generated during the release process:

```bash
# Generate CycloneDX SBOM
syft . -o cyclonedx-json > sbom.cdx.json

# Generate SPDX SBOM
syft . -o spdx-json > sbom.spdx.json
```

### Contents

The SBOM includes:

- All direct and transitive dependencies
- Version information and checksums
- License information
- Supplier/author information
- Component relationships

### Verification

To verify SBOM contents:

```bash
# Scan SBOM for vulnerabilities
grype sbom:sbom.cdx.json

# Validate SBOM format
sbom-scorecard score sbom.cdx.json
```

## Signed Releases

All release artifacts are signed using Sigstore cosign for cryptographic verification.

### What's Signed

| Artifact         | Signature File      | Certificate              |
| ---------------- | ------------------- | ------------------------ |
| Binary archives  | `*.sig`             | `*.sig.cert`             |
| Container images | In-registry         | Rekor transparency log   |
| SBOM files       | `*.sbom.sig`        | `*.sbom.sig.cert`        |
| Checksums        | `checksums.txt.sig` | `checksums.txt.sig.cert` |

### Signing Process

Releases are signed using keyless signing with Sigstore:

```bash
# Sign a release artifact
cosign sign-blob \
  --yes \
  --output-signature virtengine_v0.9.0_linux_amd64.zip.sig \
  --output-certificate virtengine_v0.9.0_linux_amd64.zip.sig.cert \
  virtengine_v0.9.0_linux_amd64.zip

# Sign container image
cosign sign --yes ghcr.io/virtengine/virtengine:v0.9.0
```

### Verification

Users can verify artifact signatures:

```bash
# Verify binary signature
cosign verify-blob \
  --signature virtengine_v0.9.0_linux_amd64.zip.sig \
  --certificate virtengine_v0.9.0_linux_amd64.zip.sig.cert \
  --certificate-identity-regexp ".*@virtengine.io" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  virtengine_v0.9.0_linux_amd64.zip

# Verify container image
cosign verify \
  --certificate-identity-regexp ".*@virtengine.io" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  ghcr.io/virtengine/virtengine:v0.9.0
```

## Build Provenance

VirtEngine provides SLSA Level 3 build provenance for all releases.

### What is Build Provenance?

Build provenance is cryptographically signed metadata about how an artifact was built, including:

- Source repository and commit
- Build platform and workflow
- Build parameters and environment
- Builder identity verification

### SLSA Levels

| Level   | Requirement                                 | VirtEngine Status |
| ------- | ------------------------------------------- | ----------------- |
| Level 1 | Documentation of build process              | ✅                |
| Level 2 | Tamper resistance of build service          | ✅                |
| Level 3 | Hardened builds, non-falsifiable provenance | ✅                |
| Level 4 | Two-person review, hermetic builds          | 🔄 In Progress    |

### Provenance Verification

```bash
# Download provenance attestation
curl -sL "https://github.com/virtengine/virtengine/releases/download/v0.9.0/virtengine_v0.9.0_linux_amd64.zip.intoto.jsonl" \
  -o provenance.json

# Verify with slsa-verifier
slsa-verifier verify-artifact \
  --provenance-path provenance.json \
  --source-uri github.com/virtengine/virtengine \
  virtengine_v0.9.0_linux_amd64.zip
```

## Reproducible Builds

VirtEngine builds are designed to be reproducible, allowing independent verification.

### Build Reproducibility

To verify a build is reproducible:

```bash
# Clone at specific tag
git clone --depth 1 --branch v0.9.0 https://github.com/virtengine/virtengine.git
cd virtengine

# Build with reproducible flags
make virtengine BUILD_FLAGS="-trimpath -buildvcs=false"

# Compare checksums
sha256sum .cache/bin/virtengine
```

### Reproducibility Guarantees

| Component        | Reproducible | Notes                             |
| ---------------- | ------------ | --------------------------------- |
| Go binaries      | ✅ Yes       | Using `-trimpath -buildvcs=false` |
| Container images | ✅ Yes       | Using `--reproducible` flag       |
| SBOM             | ✅ Yes       | Deterministic generation          |
| Documentation    | N/A          | Not applicable                    |

### Known Limitations

- CGO-enabled builds may have platform-specific variations
- Build timestamps are stripped for reproducibility
- Container image layer ordering is deterministic

## Vulnerability Scanning

### Continuous Scanning

Dependencies are continuously scanned for vulnerabilities:

| Scanner     | Ecosystem  | Schedule   | Severity Threshold |
| ----------- | ---------- | ---------- | ------------------ |
| govulncheck | Go         | Every push | All                |
| pip-audit   | Python     | Every push | High+              |
| npm audit   | npm        | Every push | High+              |
| Trivy       | Containers | Every push | Critical/High      |
| Dependabot  | All        | Daily      | All                |

### Vulnerability Response

| Severity | Response Time | Action                    |
| -------- | ------------- | ------------------------- |
| Critical | 24 hours      | Immediate patch, advisory |
| High     | 7 days        | Prioritized patch         |
| Medium   | 30 days       | Scheduled patch           |
| Low      | 90 days       | Backlog                   |

### False Positive Management

Known false positives are documented in `.vulnerability-allowlist.yaml`:

```yaml
vulnerabilities:
  - id: CVE-2023-XXXXX
    reason: "Not exploitable in our usage context"
    expires: 2024-12-31
    references:
      - https://github.com/virtengine/virtengine/issues/XXX
```

## Third-Party Risk Assessment

### Risk Scoring

All third-party dependencies are scored using OpenSSF Scorecard metrics:

| Metric            | Weight | Description                            |
| ----------------- | ------ | -------------------------------------- |
| Maintained        | 20%    | Recent commits, responsive maintainers |
| Security-Policy   | 15%    | Has security policy                    |
| Code-Review       | 15%    | Requires code review                   |
| Branch-Protection | 10%    | Protected branches                     |
| Vulnerabilities   | 20%    | Known vulnerability count              |
| License           | 10%    | Clear OSS license                      |
| CI/CD             | 10%    | Automated testing                      |

### Minimum Requirements

New dependencies must meet:

- Scorecard score ≥ 6.0
- Active maintenance (commits within 6 months)
- Clear license (Apache-2.0, MIT, BSD preferred)
- No critical vulnerabilities

### Dependency Review Process

1. **Propose** - Open issue with dependency justification
2. **Assess** - Run risk assessment tools
3. **Review** - Security team approval for score < 7.0
4. **Add** - Add with pinned version and document decision

### Assessment Command

```bash
# Run dependency risk assessment
./scripts/assess-dependencies.go

# Generate risk report
./scripts/assess-dependencies.go --report
```

## Supply Chain Attack Detection

### Attack Vectors Monitored

| Attack Type            | Detection Method        | Response       |
| ---------------------- | ----------------------- | -------------- |
| Dependency Confusion   | Namespace monitoring    | Block + Alert  |
| Typosquatting          | Package name analysis   | Block + Alert  |
| Compromised Maintainer | Scorecard monitoring    | Alert + Review |
| Malicious Update       | Checksum verification   | Block + Alert  |
| Build Compromise       | Provenance verification | Block + Alert  |

### Detection Tools

```bash
# Check for dependency confusion
./scripts/detect-supply-chain-attacks.sh --confusion

# Check for typosquatting
./scripts/detect-supply-chain-attacks.sh --typosquatting

# Full supply chain audit
./scripts/detect-supply-chain-attacks.sh --all
```

### Incident Response

1. **Detection** - Automated or manual discovery
2. **Containment** - Revert dependency, block builds
3. **Analysis** - Assess impact and exposure
4. **Remediation** - Update to safe version
5. **Disclosure** - Notify affected users if necessary
6. **Post-mortem** - Document and improve detection

## Verification Guide

### For Users

#### Verify Binary Releases

```bash
# 1. Download release and signatures
VERSION=v0.9.0
curl -sLO "https://github.com/virtengine/virtengine/releases/download/${VERSION}/virtengine_${VERSION}_linux_amd64.zip"
curl -sLO "https://github.com/virtengine/virtengine/releases/download/${VERSION}/virtengine_${VERSION}_linux_amd64.zip.sig"
curl -sLO "https://github.com/virtengine/virtengine/releases/download/${VERSION}/virtengine_${VERSION}_linux_amd64.zip.sig.cert"
curl -sLO "https://github.com/virtengine/virtengine/releases/download/${VERSION}/checksums.txt"

# 2. Verify checksum
sha256sum -c checksums.txt --ignore-missing

# 3. Verify signature
cosign verify-blob \
  --signature "virtengine_${VERSION}_linux_amd64.zip.sig" \
  --certificate "virtengine_${VERSION}_linux_amd64.zip.sig.cert" \
  --certificate-identity-regexp ".*@virtengine.io" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  "virtengine_${VERSION}_linux_amd64.zip"
```

#### Verify Container Images

```bash
# Verify container image signature
cosign verify \
  --certificate-identity-regexp ".*@virtengine.io" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  ghcr.io/virtengine/virtengine:v0.9.0

# Verify SBOM attestation
cosign verify-attestation \
  --type cyclonedx \
  --certificate-identity-regexp ".*@virtengine.io" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  ghcr.io/virtengine/virtengine:v0.9.0
```

#### Verify Build Provenance

```bash
# Install slsa-verifier
go install github.com/slsa-framework/slsa-verifier/v2/cli/slsa-verifier@latest

# Verify provenance
slsa-verifier verify-artifact \
  --provenance-path provenance.json \
  --source-uri github.com/virtengine/virtengine \
  virtengine_v0.9.0_linux_amd64.zip
```

### For Contributors

#### Before Committing

```bash
# Verify all dependencies
go mod verify
go mod tidy

# Check for vulnerabilities
govulncheck ./...

# Run supply chain checks
./scripts/verify-dependencies.sh
```

#### Adding New Dependencies

```bash
# 1. Assess the dependency
./scripts/assess-dependencies.go --package github.com/new/dependency

# 2. Add with exact version
go get github.com/new/dependency@v1.2.3

# 3. Update vendor
go mod vendor

# 4. Commit go.mod, go.sum, and vendor/
git add go.mod go.sum vendor/
git commit -s -m "chore(deps): add github.com/new/dependency v1.2.3"
```

## Tools Reference

| Tool          | Purpose                 | Installation                                                                       |
| ------------- | ----------------------- | ---------------------------------------------------------------------------------- |
| cosign        | Artifact signing        | `go install github.com/sigstore/cosign/v2/cmd/cosign@latest`                       |
| slsa-verifier | Provenance verification | `go install github.com/slsa-framework/slsa-verifier/v2/cli/slsa-verifier@latest`   |
| syft          | SBOM generation         | `curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh \| sh`  |
| grype         | Vulnerability scanning  | `curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh \| sh` |
| govulncheck   | Go vulnerability check  | `go install golang.org/x/vuln/cmd/govulncheck@latest`                              |
| scorecard     | Dependency scoring      | `go install github.com/ossf/scorecard/v4/cmd/scorecard@latest`                     |

## Contact

For supply chain security concerns, contact [security@virtengine.io](mailto:security@virtengine.io).

---

_Last updated: 2024_
_This document follows NIST SP 800-218 and SLSA guidelines._
