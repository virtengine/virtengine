# VirtEngine Rust SDK

Rust SDK for interacting with the VirtEngine chain via gRPC.

## Install

> **This SDK is source-only and will not be published to crates.io** (operator decision,
> 2026-09-29, task `t_9d3b0f86`). `cargo add virtengine-sdk` therefore does not and will
> never work: `https://crates.io/api/v1/crates/virtengine-sdk` returns
> `{"errors":[{"detail":"crate \`virtengine-sdk\` does not exist"}]}`, `https://docs.rs/virtengine-sdk`
> returns 404, and `.github/workflows/sdk-publish.yaml` is disabled (it no longer triggers
> on a release event, so publishing can never be an unattended side effect of tagging).
> Depend on the path from a checkout:

```toml
[dependencies]
virtengine-sdk = { path = "path/to/virtengine/sdk/rust" }
```

The crate name and version are declared in
[`sdk/rust/Cargo.toml`](./Cargo.toml) (`name = "virtengine-sdk"`,
`version = "0.1.0"`). It is a member of the `sdk/` Cargo workspace
(`sdk/Cargo.toml`).

## Quick start

```rust
use virtengine_sdk::VirtEngineClient;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut client = VirtEngineClient::connect("http://localhost:9090", "virtengine-1").await?;
    let identity = client.veid.identity("ve1...").await?;
    println!("{:?}", identity);
    Ok(())
}
```

## Development

```bash
cargo test
```

## Events

Use `EventSubscriber` for Tendermint WebSocket subscriptions.
