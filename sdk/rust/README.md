# VirtEngine Rust SDK

Rust SDK for interacting with the VirtEngine chain via gRPC.

## Install

```bash
cargo add virtengine-sdk
```

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
