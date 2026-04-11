# Mainnet Provider HPC E2E Report - 2026-04-11

Last updated: 2026-04-11
Owner: Provider Lead

## Scope
Repository-backed launch rehearsal evidence for checklist item `6C`:
provider onboarding and HPC marketplace execution.

## Launch-grade command

```bash
go test -tags="e2e.integration" ./tests/e2e -run "^TestHPCMarketplaceE2E$" -count=1
```

## Evidence artifact
- Raw log: `output/mainnet-launch/2026-04-11/provider-hpc-marketplace.log`
- Completed (UTC): 2026-04-11 05:54:02
- SHA-256: `01423f36bce70cca702001def194daf63dda3ee6b8e1400741206a0f6a79198a`

## Covered flows
- staging environment setup
- provider registration and offering publication
- order creation and allocation
- resource provisioning
- HPC job submission and scheduler execution
- usage metrics capture and reporting
- invoice generation and settlement
- payout and platform-fee verification
- state-transition and event validation
- negative scenarios including disputed settlement handling

## Result
- Status: `PASS`
- Launch relevance: satisfies the provider and marketplace execution-evidence
  prerequisite for the rehearsal bundle and launch packet.

## Execution note
- The archived evidence path is the deterministic HPC marketplace suite. It
  starts and tears down the chain harness cleanly and exercises the complete
  provider lifecycle required by the launch checklist.

## Source references
- `tests/e2e/hpc_marketplace_e2e_test.go`
- `output/mainnet-launch/2026-04-11/provider-hpc-marketplace.log`
