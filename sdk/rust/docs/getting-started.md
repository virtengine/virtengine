# Getting Started (Rust)

## Install

> **This SDK is source-only and will not be published to crates.io** (operator decision,
> 2026-09-29). `cargo add virtengine-sdk` fails — `https://crates.io/api/v1/crates/virtengine-sdk`
> reports `crate \`virtengine-sdk\` does not exist` — and will continue to, because the
> publish workflow (`.github/workflows/sdk-publish.yaml`) is disabled. See
> [the SDK README](../README.md) for the verified path-dependency form.

```toml
[dependencies]
virtengine-sdk = { path = "path/to/virtengine/sdk/rust" }
```

## Connect

```rust
use virtengine_sdk::VirtEngineClient;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut client = VirtEngineClient::connect("http://localhost:9090", "virtengine-1").await?;
    let response = client.veid.identity("ve1...").await?;
    println!("{:?}", response);
    Ok(())
}
```
