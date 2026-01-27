All VirtEngine custom modules MUST use error codes starting from 100 to avoid conflicts with:
- Cosmos SDK core modules (codes 1-50)
- IBC-Go modules (codes 1-50 per module)
- CosmWasm modules (codes 1-50)

Recommendation: Use error code ranges:
- pkg/* modules: 100-199
- x/* modules: 100-199 (per module)
