# Run from PowerShell in sdk directory
# This generates protos without the slow vendor scanning

cd C:\Users\jON\Documents\source\repos\virtengine-gh\virtengine\sdk

# Use buf with input directly specified to avoid workspace scanning
$env:PATH = "$PWD\.cache\bin;$env:PATH"

# Generate node protos
buf generate `
  --template buf.gen.gogo.yaml `
  --path proto/node/virtengine/audit/v1 `
  --path proto/node/virtengine/base `
  --path proto/node/virtengine/oracle/v1 `
  --path proto/node/virtengine/take/v1 `
  --path proto/node/virtengine/wasm/v1 `
  --path proto/node/virtengine/cert/v1beta4 `
  --path proto/node/virtengine/deployment/v1beta4 `
  --path proto/node/virtengine/escrow/v1 `
  --path proto/node/virtengine/market/v1beta5 `
  --path proto/node/virtengine/marketplace/v1 `
  --path proto/node/virtengine/provider/v1beta4

# Generate provider protos  
buf generate `
  --template buf.gen.gogo.yaml `
  --path proto/provider
