# VirtEngine Bug Bounty Program

## Overview
VirtEngine operates a responsible disclosure program to encourage security
researchers to report vulnerabilities in a safe and coordinated manner.

## In-Scope Targets
- VirtEngine blockchain modules (x/*)
- Provider daemon (pkg/provider_daemon/)
- Cryptographic implementations (pkg/veid/crypto/, pkg/keymanagement/)
- ML inference pipeline (pkg/inference/)
- Smart contracts, state machines, and consensus-critical logic

## Out-of-Scope Targets
- Frontend applications unless they lead to backend compromise
- Third-party dependencies (report to upstream maintainers)
- Social engineering, phishing, or physical attacks
- Denial-of-service attacks (unless coordinated)
- Automated scanning without prior approval

## Severity and Rewards

| Severity | Example Impact | Reward Range |
| --- | --- | --- |
| Critical | Fund loss, consensus break, total DoS | ,000 - ,000 |
| High | Significant financial impact, data breach | ,000 - ,000 |
| Medium | Limited impact, requires user action | ,000 - ,000 |
| Low | Minor issues, unlikely exploitation |  - ,000 |

## Examples

### Critical
- Unauthorized fund withdrawal
- Consensus manipulation
- Identity verification bypass
- Encryption key extraction

### High
- Escrow fund locking
- Provider reputation manipulation
- Partial DoS of network
- Sensitive data exposure

## Rules of Engagement
1. Do not publicly disclose vulnerabilities before a fix is released.
2. The first valid reporter is eligible for the bounty.
3. Provide proof-of-concept or clear reproduction steps.
4. Avoid impacting production environments without approval.
5. Use good faith testing and minimize data exposure.

## Submission Process
- Email: security@virtengine.io
- PGP Key: available upon request
- Expected response: 24 hours
- Expected triage: 72 hours

## Disclosure Timeline
- Coordinated disclosure target: 90 days from report date.
- Extensions possible for complex fixes or upstream dependencies.

## Legal Safe Harbor
VirtEngine will not pursue legal action against researchers who:
- Follow this program in good faith.
- Avoid privacy violations and data destruction.
- Report findings promptly and refrain from public disclosure.
