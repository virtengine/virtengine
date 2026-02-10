# Mainnet Configuration

This directory contains mainnet configuration inputs used during the genesis
ceremony and validation workflows. These files are treated as the source of
truth for production values and are referenced by `scripts/mainnet` tooling.

## Files

- `genesis-params.json` — canonical chain + module parameter values for
  mainnet genesis (staking, mint, gov, slashing, VEID, MFA, encryption, HPC).
- `genesis-allocations.json` — initial account allocations and vesting
  schedules. **Replace placeholder addresses before use.**
- `gentx-constraints.json` — validator gentx validation rules (commission,
  min self-delegation, bond denom, etc.).
- `genesis-checks.json` — explicit validation assertions for genesis files.

## Usage

1. Populate allocations + validators.
2. Run `scripts/mainnet/genesis-ceremony.sh` to build the final
   `genesis.json`.
3. Run `scripts/mainnet/genesis-validate.sh` to validate the output.

All values are in their on-chain base denom (`uve`) unless noted otherwise.