# VirtEngine Network Launch Schedule

Last updated: 2026-09-01

## Launch windows

| Network | Launch window | Purpose |
| --- | --- | --- |
| **TestNet** | **January 2027** | Public pre-production validation, integration testing, and operator rehearsal. |
| **MainNet** | **March 2027** | Production network launch after TestNet exit criteria and a fresh go/no-go approval are complete. |

These are month-level launch windows, not guaranteed launch dates. Exact UTC
dates and times will be published through the formal release and launch
process. Neither network should be described as live until its launch is
officially confirmed.

## TestNet and MainNet are different networks

**TestNet** is the public proving environment. It gives validators, providers,
integrators, and application teams a shared network on which to validate
consensus, upgrades, VEID workflows, marketplace and settlement flows,
monitoring, incident response, recovery, and operational coordination.
TestNet state may be reset, TestNet tokens have no production value, and
TestNet data or balances are not guaranteed to carry into MainNet.

**MainNet** is the production network. It starts from the approved canonical
genesis bundle and production economic parameters, and it is intended for
persistent state and production activity. MainNet therefore requires a
separate release decision, final artifact verification, and coordinated
validator activation. A successful TestNet launch does not automatically
authorize MainNet launch.

## Why the launches are separated

The gap is an explicit safety and quality gate:

1. January provides real multi-operator TestNet evidence under public network
   conditions.
2. February provides time to diagnose findings, remediate defects, repeat
   failed tests, complete security and operational reviews, and freeze the
   production release and genesis inputs.
3. March is the MainNet window only after the TestNet exit criteria are met and
   release management records a fresh `GO` decision.

Separating the launches prevents rehearsal evidence from being mistaken for
production approval and reduces the risk of carrying unresolved consensus,
security, economic, integration, or operational issues into persistent
MainNet state.

## MainNet entry criteria

Before the March 2027 MainNet window can be activated, the launch package must
confirm at least:

- stable TestNet consensus and acceptable reliability over the agreed
  observation period;
- successful validator and provider onboarding, upgrade, rollback, backup,
  and recovery exercises;
- successful end-to-end VEID, marketplace, settlement, and governance flows;
- closure and re-test of launch-blocking security, correctness, and
  performance findings;
- final release artifacts, canonical genesis inputs, and published hashes;
- a fresh MainNet readiness review and explicit go/no-go record.

Current release and verification posture is documented in
[RELEASE.md](../RELEASE.md), [VERIFICATION.md](../VERIFICATION.md), and the
internal [network launch schedule](../_docs/operations/network-launch-schedule.md).
