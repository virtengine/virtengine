# VirtEngine Rust SDK

Rust SDK for interacting with the VirtEngine chain via gRPC.

## Install

> **UNVERIFIED — not published to crates.io.** `cargo add virtengine-sdk` does not
> work: `https://crates.io/api/v1/crates/virtengine-sdk` returns
> `{"errors":[{"detail":"crate \`virtengine-sdk\` does not exist"}]}`, and
> `https://docs.rs/virtengine-sdk` returns 404. The crate is published by
> `.github/workflows/sdk-publish.yaml` on the `release: published` event, and
> `gh release list -R virtengine/virtengine` shows only `0.1.0` — a **draft** from
> 2021 — so that workflow has never run. Depend on the path from a checkout
> (see task `t_9d3b0f86`):

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
