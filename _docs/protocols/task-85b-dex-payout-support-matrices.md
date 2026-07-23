# Task 85B DEX Route and Payout Corridor Support Matrices

**Matrix version:** 1.0
**Effective date:** 2026-07-23
**Overall status:** `engineering_complete_external_blocked`

These matrices describe the checked-in engineering capability. They are not executable profile files. `TBD_EXTERNAL` means the required value or evidence is not supplied in the repository. A row containing `TBD_EXTERNAL` must not be converted into runtime configuration.

No row in this version is production-enabled. The accepted production-row states in this artifact are `engineering_complete_external_blocked`, `unsupported`, and `paused` only.

## DEX route support matrix

| Row | Environment | Network / chain ID | DEX / exact adapter version | Pool or module identity | Tokens / decimals | Finality and route limits | Oracle and custody | State |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| DEX-OSMOSIS-MAINNET-V1 | Production | Osmosis / `osmosis-1` | Osmosis / `gamm-v1beta1-cp-equal-weight-v1` | Pool type `/osmosis.gamm.v1beta1.Pool`; allowed pool IDs `TBD_EXTERNAL` | Exact denoms, IBC traces and decimals `TBD_EXTERNAL` | Finality blocks, maximum observation age/height lag, hops, minimum liquidity/reserves, maximum amount/impact/deviation and quote TTL `TBD_EXTERNAL` | Authenticated oracle source and approved external target-chain custody/HSM signer `TBD_EXTERNAL` | `engineering_complete_external_blocked` |
| DEX-OSMOSIS-TESTNET-V1 | Engineering test | Osmosis testnet / `osmo-test-5` | Osmosis / `gamm-v1beta1-cp-equal-weight-v1` | Pool IDs `TBD_EXTERNAL`; no checked-in executable profile | Token denoms/decimals `TBD_EXTERNAL` | Profile limits and retained testnet execution/finality evidence `TBD_EXTERNAL` | Test oracle/custody evidence `TBD_EXTERNAL` | `engineering_complete_external_blocked` |
| DEX-UNISWAP-V2 | Production | Network/chain `TBD_EXTERNAL` | Uniswap V2 production adapter not implemented | Contracts/pools `TBD_EXTERNAL` | `TBD_EXTERNAL` | No production quote/finality implementation | `TBD_EXTERNAL` | `unsupported` |
| DEX-CURVE | Production | Network/chain `TBD_EXTERNAL` | Curve production adapter not implemented | Contracts/pools `TBD_EXTERNAL` | `TBD_EXTERNAL` | No production quote/finality implementation | `TBD_EXTERNAL` | `unsupported` |

### Exact implemented Osmosis capability

The production-capable code supports only two-asset, equal-weight constant-product GAMM pools. It verifies exact reserve direction by denom, local constant-product exact-in math, fees, minimum output, price impact, oracle deviation, pool state at a known height/block, quote digest/expiry, canonical payload binding, actual Cosmos `TxRaw`, confirmed output and profile-defined finality. Production service registration requires an independently authorized exact profile.

Weighted pools with unequal weights, concentrated-liquidity pools, stable-swap pools, exact-out production execution, Uniswap and Curve are not supported by this matrix version.

### Missing DEX production evidence

The mainnet row cannot advance until all of the following are supplied and reviewed together:

- real allowlisted pool/module IDs and exact token denominations, decimals and IBC traces;
- retained pool-state, liquidity, reserve, route-size, fee and price-impact evidence;
- authenticated oracle source, freshness/deviation policy and incident owner;
- funded governed target-chain account, custody/HSM integration and key ceremony;
- RPC/archive/finality access and reorg reconciliation evidence;
- governance approval for exact profile bytes/digest and named engineering, operations and security owners; and
- independently reconciled low-value production swap evidence.

## Payout corridor support matrix

| Row | Environment | Provider / exact API version | Jurisdiction / currency / rail | Beneficiary and compliance | Webhook / limits / finality | Contract, credentials and owners | State |
| --- | --- | --- | --- | --- | --- | --- | --- |
| PAYOUT-PRODUCTION-JSON-V1 | Production | Contracted JSON partner `TBD_EXTERNAL`; API version `TBD_EXTERNAL` | Jurisdiction, currency, rail and corridor ID `TBD_EXTERNAL` | Tokenized beneficiary reference required by framework; exact required/prohibited fields, KYC decision, sanctions decision and approval source `TBD_EXTERNAL` | HMAC-SHA256 supported; exact webhook version/key IDs/registration, amount/daily limits, quote TTL and provider finality definition `TBD_EXTERNAL` | Executed contract, production secret-manager references, bank/custody accounts, legal/compliance/DPA approvals and engineering/operations/compliance/security owners `TBD_EXTERNAL` | `engineering_complete_external_blocked` |
| PAYOUT-SANDBOX-JSON-V1 | Engineering sandbox | Sandbox partner and exact API version `TBD_EXTERNAL` | Test jurisdiction/currency/rail `TBD_EXTERNAL` | Same strict profile schema; sandbox decisions/beneficiary tokens `TBD_EXTERNAL` | Sandbox credential, webhook and conformance evidence `TBD_EXTERNAL` | No checked-in sandbox profile or credentials | `engineering_complete_external_blocked` |
| PAYOUT-MOCK | Test only | In-memory mock | Synthetic test data only | No production standing | No production finality | Not eligible for any release floor | `unsupported` |

The generic HTTP adapter is a framework, not evidence that any named vendor or corridor is supported. Legacy references to PayPal or ACH do not constitute a configured provider, contract, API version, banking rail approval or production profile.

### Exact implemented payout capability

The strict adapter enforces an exact profile/API version, fixed HTTPS origin and allowlisted paths, bounded strict JSON, secret references resolved per request, request/response economic binding, correlation and idempotency keys, tokenized beneficiary references, independently resolved KYC/sanctions decisions, durable production daily-limit accounting, metadata recovery, monotonic status reconciliation and cancellation rules.

The webhook path authenticates timestamp plus raw body with profile-pinned HMAC-SHA256 key ID/version, verifies payout/quote/correlation binding, atomically detects exact/conflicting replay, persists verified events before acknowledgement and requires provider status agreement before completion.

### Missing payout production evidence

The production row cannot advance until all of the following are supplied and reviewed together:

- selected licensed provider, executed contract and exact API/webhook versions;
- production credentials in an approved secret manager and rotation/revocation rehearsal;
- approved bank/custody accounts and beneficiary-tokenization service;
- jurisdiction, currency, rail, amount limits, finality definition and corridor approval;
- KYC/sanctions decision authority, legal/compliance approval and DPA;
- registered webhook endpoint, signing-key ceremony/rotation and durable store ownership;
- named engineering, operations, compliance and security owners; and
- independently reconciled low-value production payout evidence, including destination receipt.

## Profile non-executability

No production profile JSON is checked in. The documented `TBD_EXTERNAL` rows intentionally fail runtime profile validation. External-blocked engineering rows require explicit engineering mode and can never satisfy production startup. Production requires independently signed exact profile files, matching on-chain IDs/digests, certified governance state, reviewed injected external backends and complete ledger evidence.
