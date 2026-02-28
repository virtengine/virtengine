# VirtEngine Compatibility and Support Policy

This document describes the compatibility posture that can be supported by checked-in repository evidence today. It intentionally avoids promising release windows or support lifecycles that are not yet backed by a launched public network and published support commitments.

## Current Support Posture

As of `2026-04-11`:

- the repository contains active development on `main`;
- launch-readiness materials for mainnet exist in-repo;
- the checked-in mainnet decision is `GO` for the scheduled 2026-04-18 UTC launch window;
- no public stable mainnet support window or end-of-life calendar is approved in repository evidence.

Because of that, VirtEngine still does **not** publish an `N-2` support policy or a guaranteed production maintenance window for older minor releases, even though the scheduled launch window is now approved in repository evidence.

## What Is Currently Safe To Rely On

The following statements are supported by the repository state:

- semantic-versioned tags and prereleases are supported by release tooling;
- compatibility and upgrade testing workflows exist and should be treated as the authoritative validation gates for release candidates;
- module and API version strings checked into code define the current wire surface more reliably than older release-train documentation does.

## Repository API Surface

The repository currently exposes these primary module version families in code and generated APIs:

| Surface | Current in-tree version family |
| --- | --- |
| `veid` | `v1` |
| `mfa` | `v1` |
| `encryption` | `v1` |
| `market` | `v1beta5` |
| `deployment` | `v1beta5` |
| `provider` | `v1beta4` |
| `escrow` | `v1` |
| `audit` | `v1` |
| `cert` | `v1` |
| `hpc` | `v1` |

These version families describe the in-tree protocol surface. They do not, by themselves, guarantee long-term backward compatibility or a public deprecation window.

## Compatibility Commitment

Until the project records an approved public release and support policy, the compatibility commitment is intentionally narrow:

1. Tagged releases should preserve compatibility within the scope validated by the corresponding compatibility, upgrade, and security workflows.
2. Breaking changes must be called out in release notes and migration documentation before a tag is presented as a supported upgrade target.
3. Operators should validate exact client, node, and provider combinations against the target tag rather than assuming an open-ended maintenance promise.

## Clients and Operators

For the current pre-mainnet posture:

- use matching binaries and SDKs from the same validated release tag where possible;
- treat mixed-version environments as supported only when the relevant compatibility and upgrade tests cover them;
- do not assume older provider, portal, or SDK builds remain supported after a new tag unless release notes say so explicitly.

## Deprecation and Breaking Changes

Breaking changes still require the usual engineering controls:

- explicit release-note callouts;
- migration guidance when state, CLI, API, or operational procedures change materially;
- compatibility or upgrade tests proving the intended supported path.

What is not currently committed:

- a fixed number of minors kept in maintenance;
- a dated deprecation and sunset calendar;
- a public guarantee that deprecated APIs remain available for a specific number of months.

Those commitments should be added only after a released network line and support program are actually approved.

## Verification Sources

When deciding whether a release or environment combination is supported, use these sources in order:

1. the target tag and its release notes;
2. [VERIFICATION.md](../VERIFICATION.md);
3. [RELEASE.md](../RELEASE.md);
4. the relevant CI workflow results for compatibility, security, smoke, and upgrade coverage;
5. for mainnet specifically, the launch packet and the go/no-go record under `_docs/operations/`.

## What This Document Does Not Promise

This document does not promise:

- that the network is already live before the approved launch window begins;
- an `N-2` maintenance policy;
- automatic protocol negotiation behavior beyond what the live server and generated API surface implement;
- support for arbitrary cross-version client and server mixes.

## Related Documentation

- [README.md](../README.md)
- [RELEASE.md](../RELEASE.md)
- [VERIFICATION.md](../VERIFICATION.md)
- [_docs/version-control.md](../_docs/version-control.md)
