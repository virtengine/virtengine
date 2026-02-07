# Getting Started (Rust)

## Install

```bash
cargo add virtengine-sdk
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
