# Mainnet Configuration

This directory contains mainnet configuration inputs used during the genesis
ceremony and validation workflows. These files are treated as the source of
truth for production values and are referenced by `scripts/mainnet` tooling.

## Files

- `genesis-params.json` — canonical chain + module parameter values for
  mainnet genesis (staking, mint, gov, slashing, VEID, MFA, encryption, HPC).
- `genesis-allocations.json` — initial account allocations and vesting
  schedules. The file now contains the approved treasury, team-vesting, and
  validator self-delegation accounts plus the deterministic distribution module
  account funding required for the initial community pool.
- `gentx-constraints.json` — validator gentx validation rules (commission,
  min self-delegation, bond denom, etc.).
- `genesis-checks.json` — explicit validation assertions for genesis files.

## Usage

1. Review or update allocations + validators.
2. Run `scripts/mainnet/genesis-ceremony.sh` to build the final
   `genesis.json`.
3. Run `scripts/mainnet/genesis-validate.sh` to validate the output.
4. Keep `community_pool` aligned with the `distribution` module account address
   returned by `virtengine debug module-address distribution`.
5. Refresh `_docs/operations/mainnet-launch-packet.md` after any input change so
   the published hashes continue to match the checked-in artifact bundle.

All values are in their on-chain base denom (`uve`) unless noted otherwise.
