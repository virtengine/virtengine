# Mainnet Allocation Control Record - 2026-04-11

Last updated: 2026-04-11
Owner: Release Management (Ops)

## Purpose
Archive the approved canonical public addresses and funding set that were
inserted into `config/mainnet/genesis-allocations.json` and used to build the
final checked-in VirtEngine mainnet genesis bundle.

## Approval window
- Approval timestamp (UTC): 2026-04-11 07:44
- Approval handles: `RM-01`, `OPS-01`, `FIN-01`, `PROD-01`
- Scope: treasury allocation, distribution community pool funding, team
  vesting allocation, and validator self-delegation funding accounts

## Canonical network metadata
- Chain ID: `virtengine-1`
- Genesis time (UTC): `2026-06-01T00:00:00Z`
- Final `genesis.json` SHA-256:
  `a8d8a4a4f19882503265482c9433a6646d8dbbfe62f81c5945e81c32da9e6032`
- Final `ceremony-manifest.json` SHA-256:
  `f7a4dceb7289c3e6affc50193cec1499f0f18b62fc480f16e665e66b6da1a0d9`

## Approved allocations

| Allocation | Address | Amount (`uve`) | Notes |
| --- | --- | --- | --- |
| Treasury | `ve1p05jyyqye4s5004j5mzy46ck82v83mftz8cn0w` | `1000000000000` | Canonical treasury funding account |
| Distribution community pool | `ve1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8mzlgxh` | `200000000000` | Deterministic `distribution` module account verified via `virtengine debug module-address distribution` |
| Team vesting | `ve1yf40gcq5d58lq6gwmmt5u66ev2vnn27s7z9dvq` | `500000000000` | Continuous vesting from `2027-06-01T00:00:00Z` to `2030-06-01T00:00:00Z` |
| Validator 01 self-delegation | `ve1e4zhpynperjqdrzpf6pa3625fd0xe4cr0lv5kx` | `1000000000` | Funds `gentx/gentx-validator-01.json` |
| Validator 02 self-delegation | `ve1stpwwehdrdk6d3fu0kwa289evcpv9s9ajxctv5` | `1000000000` | Funds `gentx/gentx-validator-02.json` |
| Validator 03 self-delegation | `ve18n4w37kmu57jpqrce2dusunm8y24uc62v9j8as` | `1000000000` | Funds `gentx/gentx-validator-03.json` |

## Validator operator mapping

| Validator | Account address | Operator address | Gentx |
| --- | --- | --- | --- |
| Validator 01 | `ve1e4zhpynperjqdrzpf6pa3625fd0xe4cr0lv5kx` | `vevaloper1e4zhpynperjqdrzpf6pa3625fd0xe4crxf9vjk` | `gentx/gentx-validator-01.json` |
| Validator 02 | `ve1stpwwehdrdk6d3fu0kwa289evcpv9s9ajxctv5` | `vevaloper1stpwwehdrdk6d3fu0kwa289evcpv9s9ams3ngy` | `gentx/gentx-validator-02.json` |
| Validator 03 | `ve18n4w37kmu57jpqrce2dusunm8y24uc62v9j8as` | `vevaloper18n4w37kmu57jpqrce2dusunm8y24uc629nmleq` | `gentx/gentx-validator-03.json` |

## Supply coverage
- Approved bank allocations plus the funded community pool total
  `1703000000000uve`.
- The final genesis bundle and deterministic checks confirm that the published
  bank supply and `distribution.fee_pool.community_pool` match this approved
  allocation set exactly.

## Referenced artifacts
- `config/mainnet/genesis-allocations.json`
- `gentx/gentx-validator-01.json`
- `gentx/gentx-validator-02.json`
- `gentx/gentx-validator-03.json`
- `artifacts/mainnet/genesis.json`
- `artifacts/mainnet/genesis.sha256`
- `artifacts/mainnet/gentx.sha256`
- `artifacts/mainnet/ceremony-manifest.json`
- `artifacts/mainnet/ceremony-manifest.sha256`
- `output/mainnet-launch/2026-04-11/mainnet-genesis-ceremony.log`

## Approval outcome
- The canonical allocation set is approved for the scheduled mainnet launch
  window on 2026-04-18 UTC.
- The checked-in `config/mainnet/genesis-allocations.json` file is the
  authoritative repository input for final genesis publication.
