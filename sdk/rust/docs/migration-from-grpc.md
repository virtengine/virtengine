# Migration from raw gRPC (Rust)

The SDK wraps tonic-generated gRPC clients and provides helpers for
transaction signing. Replace direct use of `QueryClient` types with
`VirtEngineClient` and module helpers (`VEIDModule`, `MarketModule`, `EscrowModule`).
