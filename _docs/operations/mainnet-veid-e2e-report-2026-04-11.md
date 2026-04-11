# Mainnet VEID E2E Report - 2026-04-11

Last updated: 2026-04-11
Owner: VEID Lead

## Scope
Repository-backed launch rehearsal evidence for the VEID onboarding and
verification prerequisite in checklist item `6B`.

## Launch-grade command

```bash
go test -tags="e2e.integration" ./tests/e2e -run "^TestVEIDE2E$" -count=1
```

## Evidence artifact
- Raw log: `output/mainnet-launch/2026-04-11/veid-launch-suite.log`
- Completed (UTC): 2026-04-11 05:53:45
- SHA-256: `4c05f0c02f382f31ca458fb7b889bb0451a305b7f5eb0a62504b0e993374c43a`

## Covered flows
- complete onboarding flow
- email verification flow
- expired OTP rejection
- SMS verification flow
- VoIP blocking
- SSO verification flow
- attestation recording
- ML scoring and tier transitions
- marketplace VEID gating
- invalid client signature rejection
- max attempts exceeded handling
- scope revocation

## Result
- Status: `PASS`
- Launch relevance: satisfies the VEID execution-evidence prerequisite for the
  rehearsal bundle and launch packet.

## Determinism note
- The SMS launch flow now evaluates record activity against the suite's fixed
  timestamp fixture instead of wall-clock time, removing expiry drift from the
  launch evidence path.

## Source references
- `tests/e2e/veid_e2e_test.go`
- `output/mainnet-launch/2026-04-11/veid-launch-suite.log`
