# Mainnet Launch Communications Packet - 2026-04-11

Last updated: 2026-04-11
Owner: Product Lead

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
> Launch control remains in `HOLD` following the 2026-04-11 rehearsal bundle.
> The execution-evidence gate is complete, but final mainnet activation is
> still waiting on signed canonical treasury, community-pool, and team-vesting
> addresses so the final genesis artifact can be rebuilt and published.

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
> Validators and providers: the 2026-04-11 rehearsal bundle is complete and
> the launch evidence packet is current. The launch window remains on hold
> pending signed mainnet allocation addresses and final genesis publication.
> Do not start production chain processes until `LAUNCH-DEC-001` is moved from
> `HOLD` to `GO`.
