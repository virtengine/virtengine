# T4-16A Fund Route Inventory Evidence

Status: diagnostic-green, integration-blocked

The inventory binds 18 production Go files that invoke mint, burn, account to
module, module to account, or module to module bank primitives. Those files are
assigned to 20 registered-message, lifecycle, payout, reward, escrow, recovery,
and treasury route records across BME, delegation, escrow, HPC, market, oracle,
settlement, staking, and standard Cosmos bank/distribution modules.

Discovery fails on any added, removed, or unassigned mover file. Release
enforcement uses `--require-ready` and remains blocked until an immutable T5
FundAuthorization checkpoint is accepted and every route is wired and proven
atomic. A claimed checkpoint must match the accepted T4 ledger tag and payload;
a syntactically valid SHA alone cannot enable readiness. The accepted payload
must also contain the canonical authorization, keeper, policy, and registry
source paths. The inventory records the known non-atomic expired-escrow refund
path and incomplete standard bank/distribution route classification.

This checkpoint does not define or wire the producer-owned FundAuthorization
keeper. It provides the fail-closed T4 integration boundary only.