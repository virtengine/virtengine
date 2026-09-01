# Mainnet Launch Communications Packet - 2026-04-11

> **Historical record — do not publish these drafts:** The April 2026 MainNet
> window did not proceed. Current public messaging must state TestNet in
> January 2027 and MainNet in March 2027, subject to separate confirmation. See
> `network-launch-schedule.md`.

Last updated: 2026-04-11
Owner: Product Lead

## Current control state
- Launch decision: `GO`
- Scope: approved for the scheduled UTC launch windows below
- Public posture: scheduled / approved, not already live before the launch
  window begins

## Approved control windows
- Primary launch window (UTC): 2026-04-18 05:00 to 2026-04-18 07:00
- Backup launch window (UTC): 2026-04-19 05:00 to 2026-04-19 07:00
- Freeze window (UTC): 2026-04-17 12:00 to 2026-04-19 12:00

## Channels and cadence
- Internal incident channel: `#ve-mainnet-launch`
- War room bridge owner: `PROD-01`
- Internal launch updates: every 10 minutes from `T-30` to `T+30`, then every
  15 minutes until `T+120`
- External status-page updates: every 15 minutes during launch or rollback

## Draft status-page scheduled-maintenance notice
> VirtEngine mainnet launch operations are scheduled for 2026-04-18 between
> 05:00 UTC and 07:00 UTC. Customer-facing verification, provider onboarding,
> and settlement smoke checks will run during this window. We will post updates
> every 15 minutes and immediately if the launch is placed on hold or rolled
> back.

## Draft hold notice
> VirtEngine mainnet launch has been placed on temporary hold before chain
> activation. The team is investigating a pre-launch control issue. Production
> chain processes must remain stopped until a new `GO` notice is issued.

## Draft go-live notice
> VirtEngine mainnet launch has started at the approved UTC launch window. Core
> verification, provider marketplace, finance reconciliation, and DR smoke
> checks are in progress. The next update will be posted in 15 minutes.

## Draft rollback notice
> VirtEngine mainnet launch has been rolled back within the approved launch
> window. Traffic and operational controls have returned to the previously
> stable state while the team investigates the triggering condition. The next
> update will be posted in 15 minutes.

## Draft partner notice
> Validators and providers: the final genesis bundle and launch evidence packet
> are published in the repository, and `LAUNCH-DEC-001` is `GO` for the
> approved UTC launch window on 2026-04-18. Do not start production chain
> processes before the coordinator green-light call inside that window.
