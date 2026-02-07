# Security Policy

VirtEngine is committed to ensuring the security of our decentralized cloud computing platform and the safety of our users. This document outlines our security policies, vulnerability reporting procedures, and best practices for contributors.

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          | Notes                                |
| ------- | ------------------ | ------------------------------------ |
| 0.9.x   | :white_check_mark: | Current development (main branch)    |
| 0.8.x   | :white_check_mark: | Stable release (mainnet/main branch) |
| 0.7.x   | :warning:          | Critical fixes only                  |
| < 0.7   | :x:                | No longer supported                  |

We strongly recommend running the latest stable release to ensure you have all security patches.

## Reporting a Vulnerability

We take all security vulnerabilities seriously. If you discover a security issue, please report it responsibly.

### Responsible Disclosure Process

1. **Do NOT** disclose the vulnerability publicly until we have addressed it
2. **Do NOT** exploit the vulnerability beyond what is necessary to demonstrate the issue
3. **Do** provide detailed information to help us understand and reproduce the issue
4. **Do** allow reasonable time for us to address the issue before public disclosure

### How to Report

**Email:** [security@virtengine.io](mailto:security@virtengine.io)

When reporting a vulnerability, please include:

- **Description:** A clear and concise description of the vulnerability
- **Impact:** Potential impact and severity assessment
- **Reproduction Steps:** Detailed steps to reproduce the issue
- **Affected Components:** Module, file, or function where the vulnerability exists
- **Proof of Concept:** Code, screenshots, or logs demonstrating the vulnerability (if applicable)
- **Suggested Fix:** Any recommendations for remediation (optional)
- **Your Contact Information:** For follow-up questions

You will receive an acknowledgment within 48 hours of your report.

## Security Contact Information

| Contact Method | Details                                                 |
| -------------- | ------------------------------------------------------- |
| Email          | [security@virtengine.io](mailto:security@virtengine.io) |
| PGP Key        | Available upon request                                  |
| Response Time  | Initial acknowledgment within 48 hours                  |

For urgent matters, please include `[URGENT]` in your email subject line.

## Severity Definitions and Response Times

We classify vulnerabilities according to their potential impact:

### Critical Severity

**Response Time: 24 hours**

- Remote code execution on validators or provider nodes
- Consensus manipulation or chain halting attacks
- Private key extraction or unauthorized fund transfers
- Complete bypass of identity verification (VEID)
- Unauthorized access to encrypted user data

### High Severity

**Response Time: 7 days**

- Denial of service attacks on core infrastructure
- Privilege escalation within modules
- Bypass of MFA or authentication mechanisms
- Significant information disclosure of sensitive data
- Cross-module state corruption

### Medium Severity

**Response Time: 30 days**

- Limited information disclosure
- Non-critical authentication bypasses
- Resource exhaustion vulnerabilities
- Logic errors with limited impact
- Partial bypass of security controls

### Low Severity

**Response Time: 90 days**

- Minor information leaks (non-sensitive)
- Best practice violations
- Theoretical vulnerabilities with no practical exploit
- Issues requiring significant user interaction
- Documentation security improvements

## Public Disclosure Policy

We follow a **90-day disclosure timeline**:

1. **Day 0:** Vulnerability reported and acknowledged
2. **Day 1-7:** Initial triage and severity assessment
3. **Day 7-60:** Development and testing of fix
4. **Day 60-75:** Coordinated disclosure preparation
5. **Day 75-90:** Staged rollout to mainnet
6. **Day 90:** Public disclosure (or earlier if fix is deployed)

We may extend this timeline for complex issues requiring extensive changes. We will keep reporters informed of progress throughout the process.

## External Security Audit

VirtEngine completes regular third-party security audits for consensus-critical
cryptographic and identity components. The latest public summary is available at:

- `_docs/audits/security-audit-report-2026-02-06.md`

### Credit and Recognition

Security researchers who report valid vulnerabilities will be:

- Credited in our security advisories (unless anonymity is requested)
- Listed in our Security Hall of Fame (coming soon)
- Eligible for bug bounty rewards (see below)

## Security Best Practices for Contributors

All contributors must follow these security guidelines:

### Code Security

- **Never commit secrets:** API keys, private keys, passwords, or tokens must never be committed
- **Use environment variables:** Store sensitive configuration in environment variables
- **Validate all inputs:** Sanitize and validate user inputs in keepers and message handlers
- **Follow the principle of least privilege:** Request only necessary permissions
- **Use secure defaults:** Default configurations should be secure

### Cosmos SDK Specific

- **Authority validation:** Always validate `msg.Authority` against the module's authority
- **Context deadlines:** Set appropriate deadlines for keeper methods, especially for ML inference
- **Deterministic operations:** Ensure all consensus-critical code is deterministic
- **Store key isolation:** Use proper store key prefixes to prevent cross-module access

### Cryptographic Guidelines

- **Use approved algorithms:** X25519-XSalsa20-Poly1305 for encryption envelopes
- **Never roll your own crypto:** Use established libraries
- **Proper key management:** Clear sensitive data from memory after use
- **Signature validation:** Verify all required signatures (client, user, salt binding)

### Dependency Management

- **Pin dependencies:** Use exact versions in `go.mod`
- **Review updates:** Audit dependency updates for security implications
- **Monitor advisories:** Stay informed about vulnerabilities in dependencies

## Supply Chain Security

VirtEngine implements comprehensive supply chain security measures. For detailed information, see [SUPPLY_CHAIN_SECURITY.md](SUPPLY_CHAIN_SECURITY.md).

### Key Features

| Feature                    | Description                                    | Status    |
| -------------------------- | ---------------------------------------------- | --------- |
| **Dependency Pinning**     | All dependencies use exact versions            | ✅ Active |
| **SBOM Generation**        | CycloneDX and SPDX formats                     | ✅ Active |
| **Signed Releases**        | Sigstore cosign keyless signing                | ✅ Active |
| **Build Provenance**       | SLSA Level 3 attestation                       | ✅ Active |
| **Vulnerability Scanning** | Automated via Dependabot, govulncheck, Trivy   | ✅ Active |
| **Attack Detection**       | Typosquatting, dependency confusion monitoring | ✅ Active |

### Verifying Releases

All release artifacts are signed with Sigstore. Verify signatures with:

```bash
# Install cosign
go install github.com/sigstore/cosign/v2/cmd/cosign@latest

# Verify binary signature
cosign verify-blob \
  --signature virtengine_v0.9.0_linux_amd64.zip.sig \
  --certificate virtengine_v0.9.0_linux_amd64.zip.pem \
  --certificate-identity-regexp ".*@virtengine.io" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  virtengine_v0.9.0_linux_amd64.zip

# Verify container image
cosign verify \
  --certificate-identity-regexp ".*@virtengine.io" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  ghcr.io/virtengine/virtengine:v0.9.0
```

### Supply Chain Tools

```bash
# Verify dependencies
./scripts/supply-chain/verify-dependencies.sh

# Detect supply chain attacks
./scripts/supply-chain/detect-supply-chain-attacks.sh

# Assess dependency risk
go run ./scripts/supply-chain/assess-dependencies.go

# Generate SBOM
./scripts/supply-chain/generate-sbom.sh
```

## Security Scanning in CI

Our continuous integration pipeline includes comprehensive security scanning:

### Static Analysis Tools

| Tool            | Purpose                                           | Runs On  |
| --------------- | ------------------------------------------------- | -------- |
| **gosec**       | Go security linter for common vulnerabilities     | Every PR |
| **gitleaks**    | Secrets and credential detection                  | Every PR |
| **govulncheck** | Go vulnerability database checking                | Every PR |
| **CodeQL**      | Semantic code analysis (GitHub Advanced Security) | Every PR |

### Additional Checks

- **golangci-lint:** Includes security-focused linters
- **Dependency scanning:** Automated alerts for vulnerable dependencies
- **Container scanning:** Security analysis of Docker images
- **SAST/DAST:** Static and dynamic application security testing

### Running Security Scans Locally

```bash
# Run gosec
make lint-go  # Includes gosec via golangci-lint

# Run govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Run gitleaks
gitleaks detect --source .
```

## Bug Bounty Program

> **Note:** Our formal bug bounty program is currently under development. The information below represents our planned program structure.

### Rewards (Planned)

| Severity | Reward Range     |
| -------- | ---------------- |
| Critical | $5,000 - $50,000 |
| High     | $2,000 - $10,000 |
| Medium   | $500 - $2,000    |
| Low      | $100 - $500      |

Actual rewards will depend on:

- Impact and exploitability
- Quality of the report
- Novelty of the vulnerability

### Program Status

We are actively developing our bug bounty program. In the meantime:

- Report vulnerabilities to [security@virtengine.io](mailto:security@virtengine.io)
- Good-faith reporters will be recognized and may receive discretionary rewards
- Sign up for announcements at [virtengine.io](https://virtengine.io) for program launch updates

## Scope

### In Scope

The following components are eligible for security reporting:

#### Core Blockchain

- `x/veid` - Identity verification module
- `x/mfa` - Multi-factor authentication module
- `x/encryption` - Encryption module
- `x/market` - Marketplace module
- `x/escrow` - Escrow module
- `x/roles` - Roles module
- `x/hpc` - HPC module
- `app/` - Application wiring and ante handlers

#### Infrastructure

- `pkg/provider_daemon` - Provider daemon and adapters
- `pkg/inference` - ML inference engine
- `cmd/` - CLI binaries

#### Cryptographic Components

- Encryption envelope implementation
- Key management systems
- Signature verification

#### Smart Contracts and State Machine

- Genesis state validation
- Message handlers and keeper methods
- State transitions and invariants

### Out of Scope

The following are **not** eligible for bounty rewards:

- Third-party services and dependencies (report to upstream maintainers)
- Social engineering attacks
- Physical security attacks
- Denial of service via resource exhaustion without amplification
- Issues in test or development environments only
- Previously reported vulnerabilities
- Vulnerabilities requiring unlikely user interaction
- Issues in deprecated or unsupported versions
- Theoretical vulnerabilities without proof of concept

### Safe Harbor

We will not pursue legal action against researchers who:

- Act in good faith and follow this policy
- Avoid privacy violations and data destruction
- Do not disrupt our services or users
- Report findings promptly and confidentially

## Security Advisories

Security advisories are published via:

- [GitHub Security Advisories](https://github.com/virtengine/virtengine/security/advisories)
- Our security mailing list (subscribe at [virtengine.io](https://virtengine.io))
- Release notes and CHANGELOG.md

## Questions

For questions about this security policy, contact [security@virtengine.io](mailto:security@virtengine.io).

---

_Last updated: 2024_

_This security policy follows [GitHub's security policy best practices](https://docs.github.com/en/code-security/getting-started/adding-a-security-policy-to-your-repository)._
