# Compute-credit naming and economic dependency audit

Reviewed 3 October 2026. This records a source/API rename and implementation
audit, not a deployment, executed conversion, or live asset migration.

## Adopted source convention

| Concept | Name | Minimal unit / interface |
| --- | --- | --- |
| Native monetary asset | VE | `uve` |
| Service accounting credit | VirtEngine Compute Credit (VCC) | `uvcc` |
| Conversion implementation module | Burn/Mint Engine | `x/bme` |
| Credit conversion requests | Mint / burn VCC | `MsgMintVCC`, `MsgBurnVCC`, RPCs `MintVCC`, `BurnVCC` |
| CLI | Mint / burn VCC | `mint-vcc`, `burn-vcc` |
| Native oracle feed | Explicitly configured VE feed | `native_price_feed_id` |

VCC describes purpose and ownership of the design. It is not Akash Compute
Token (ACT), an external-network dependency, or a newly established tradable
asset with guaranteed stable value. The six-decimal client display convention
does not create collateral or a peg. Burn/mint mechanics are generic; the term
“equilibrium” should be reserved for a model and outcome actually evaluated.

## Affected surfaces

- Canonical `sdk/proto/node/virtengine/bme/v1` messages/RPCs/comments and escrow
  deposit documentation; `vault_native` replaces the inherited asset-specific
  vault response field while retaining protobuf field number 1.
- BME message handlers and tests, SDK Go message and gRPC descriptors, CLI,
  TypeScript generated messages/services/index/SDK factory, and wallet currencies.
  Wallet fees and stake now display native `VE/uve`, matching the native Go paths.
- Funding-authorization registry entries and canonical inventory fixtures now use
  the VCC message URLs. Their registry and inventory digests change with this
  interface update; the declared registry does not establish live BME enforcement.
- Descriptor set, inventory, protobuf OpenAPI, generated reference documentation,
  and experimental Python/Rust outputs. These latter SDKs remain unsupported
  release contracts under `sdk/generation/toolchain.json`.
- Oracle config formerly named an AKT/USD price feed. It now names the native
  VE feed and supplies no hard-coded default identifier. An unconfigured Pyth
  config fails validation; default module params already contain no configured
  feed contracts. This does not assert that a VE market feed currently exists.
- Market liquidity reward defaults and benchmarks use `uve`, not `uakt`.
  The reward denomination correction does not validate reward financing or
  primary-issuance policy integration.
- SDK gas pricing and the email chain client's default gas price use `uve`.
  Natural-language amount extraction recognizes `uvcc` service units alongside
  native `uve`, without treating them as equivalent monetary values.
- DET.io, VirtEngine's public site, and protocol docs use VCC for service-credit
  nomenclature. Legal jurisdiction abbreviations, document classifications,
  unrelated instruction prose and historical fork provenance are not tokens.

## Execution and economics findings

`x/bme/keeper/msg_server.go` conversion handlers return pending and do not
execute oracle-priced burns/mints, complete collateral checks, or enqueue a
durable conversion. Generic bank helpers in `fee_collection.go` exist, but
cannot be treated as completed conversion integration.

`SettleBilling` transfers payment to providers and burns a configurable fee.
`DefaultSettleSpreadBps` is zero. The checked path therefore does not establish
permanent destruction of service payments. Reconcile configurable nonzero fee
settings with the agreed full-provider-payout policy before adoption.

Primary identity-led issuance is a separate proposed policy. Service-credit
minting, native reissuance, refunds and reward-pool funding must be individually
accounted for. Track native gross burns minus native reissuance over a defined
interval, credit liabilities, redemption costs, reserves and provider payments.
Do not infer an irreversible monetary sink from a gross burn or a credit burn.

## Compatibility requirements

Old ACT/vACT message URLs, RPC names and CLI commands are not maintained as
active aliases. Renamed message type URLs and methods break old clients and
signed transactions, even though protobuf field numbers are preserved. The
oracle field rename also changes JSON/config field names. Rebuild clients and
signed requests for a coordinated version upgrade. Authorizations or proofs bound
to the previous funding-registry digest require coordinated regeneration too.

`uact` and `uvact` balances are not automatically `uvcc` balances. No live state
has been rewritten. Fresh development genesis can use the new naming. Before
an existing network adopts it, inventory bank supply, metadata and balances;
escrow deposits and payments; BME balances, ledger IDs, pending records and
remint-credit accounting; configured limits and oracle contract parameters;
wallet/client metadata; and signed/off-chain records. Define one audited
migration with conservation assertions, collision handling and rollback, or
retain the existing asset identifier until that upgrade is authorized.

## Research implications

The attached economic insight motivates testing five jointly relevant
conditions: use of the governed resource market; recurring net consumption;
human-rooted primary issuance; control of future eligibility/benefits; and
protection against monetary-rule capture. They are hypotheses, not measured
probabilities or a validated product formula. Identity uniqueness does not
prevent coercion or key/reward capture. The Foundation's asset lock does not
automatically protect stake-weighted chain governance.

DET.io now teaches these distinctions through interactive net-flow, entitlement
and condition experiments at `/learn/economic-invariants`, with implementation
evidence and possible post-labour benefits in the dedicated research synthesis.

## Validation scope

Targeted BME and SDK Go tests, descriptor-backed credit-message decoding,
CLI command discovery, TypeScript SDK build and contract parity, source scans,
site builds/link/accessibility checks, and browser interaction checks cover this
rename. No live conversion, price feed, production migration or post-labour
economic outcome is claimed. Native Windows code generation used installed
plugins and the repository templates because the Docker entry point was
unavailable; this is not certification of the pinned Linux release generator.

## Sources

- [Akash's ACT/BME description](https://akash.network/blog/what-burn-mint-equilibrium-means-for-akash/)
- `sdk/proto/node/virtengine/bme/v1/{msgs,service,types}.proto`
- `x/bme/keeper/{msg_server,fee_collection,settlement}.go`
- `sdk/go/node/bme/v1/params.go`
- `sdk/proto/node/virtengine/oracle/v1/params.proto`
- `sdk/go/node/oracle/v1/params.go`
- [Personhood reward capture case study](https://papers.ssrn.com/sol3/papers.cfm?abstract_id=4749892)
- [Work and distribution scenario research](https://www.nber.org/papers/w30172)
