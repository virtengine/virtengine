# Getting Started (Rust)

## Install

> **UNVERIFIED — not published to crates.io.** `cargo add virtengine-sdk` fails:
> `https://crates.io/api/v1/crates/virtengine-sdk` reports
> `crate \`virtengine-sdk\` does not exist`; the publish workflow
> (`.github/workflows/sdk-publish.yaml`) fires only on `release: published` and
> the repository has never published a release. See
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
