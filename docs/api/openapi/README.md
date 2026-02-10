# VirtEngine OpenAPI Specification

This directory contains comprehensive OpenAPI/Swagger documentation for the VirtEngine API.

## 📁 Files

| File | Description |
|------|-------------|
| `virtengine-api.yaml` | **Main OpenAPI 3.0 specification** - Complete blockchain API |
| `portal_api.yaml` | Provider portal OpenAPI 3.0 specification |
| `swagger-ui.html` | Interactive Swagger UI documentation viewer |
| `redoc.html` | Beautiful Redoc documentation viewer |
| `swagger-config.json` | Swagger UI configuration |

## 🚀 Quick Start

### View Interactive Documentation

#### Option 1: Local HTTP Server

```bash
# Serve documentation
npx http-server docs/api/openapi -p 8080

# Open in browser
open http://localhost:8080/swagger-ui.html
open http://localhost:8080/redoc.html
```

#### Option 2: Via Docker

```bash
# Start Swagger UI with Docker
docker run -p 8080:8080 \
  -e SWAGGER_JSON=/docs/virtengine-api.yaml \
  -v $(pwd):/docs \
  swaggerapi/swagger-ui

# Access at http://localhost:8080
```

#### Option 3: Online Viewer

Upload `virtengine-api.yaml` to:
- https://editor.swagger.io
- https://redocly.github.io/redoc/

## 📖 Specifications Overview

### Main Blockchain API (`virtengine-api.yaml`)

Complete OpenAPI specification for VirtEngine blockchain modules:

**Modules covered:**
- ✅ VEID (Identity verification)
- ✅ MFA (Multi-factor authentication)
- ✅ Encryption (Public-key encryption)
- ✅ Market (Orders, bids, leases)
- ✅ Deployment (Deployment management)
- ✅ Provider (Provider registration)
- ✅ Escrow (Payment escrow)
- ✅ HPC (High-performance computing)
- ✅ Cert, Audit, Roles, Staking, Enclave, etc.

**Features:**
- Rate limit annotations (`x-rate-limit`)
- Authentication requirements (`x-auth-required`)
- Deprecation warnings (`x-deprecation`)
- Complete schema definitions
- Request/response examples

### Provider Portal API (`portal_api.yaml`)

Off-chain provider portal REST API specification:

**Endpoints:**
- Health and status
- Deployment management
- Metrics and monitoring
- Organization management
- Support tickets
- Billing and invoices
- Vault operations

## 🛠️ Generated Specifications

### From Proto Files

The main OpenAPI specification can be auto-generated from proto files using `buf generate`:

```bash
# Generate swagger from proto files
cd sdk
buf generate --template buf.gen.swagger.yaml

# Merge generated files
cat .cache/tmp/swagger-gen/*.swagger.json | \
  jq -s 'reduce .[] as $item ({}; . * $item)' > swagger-merged.json

# Convert to YAML
npx js-yaml swagger-merged.json > docs/api/openapi/generated.yaml
```

The generated specification is available at:
- **Swagger UI**: `http://localhost:8000/portal` (via Kong gateway)
- **Static File**: `sdk/docs/swagger-ui/swagger.yaml`

### Provider Portal Generation

The provider portal OpenAPI spec generates TypeScript and Go types:

```bash
# Generate types for all languages
./scripts/generate-api-types.sh

# Output locations:
# - TypeScript: lib/portal/src/provider-api/generated/types.ts
# - Go: pkg/provider_daemon/api/generated/types.go
```

## 🔧 API Versioning in OpenAPI

The API uses path-based versioning reflected in the OpenAPI paths:

```yaml
paths:
  /virtengine/veid/v1/identity/{account_address}:
    get:
      summary: Query identity record
      tags:
        - VEID
      # ...

  /virtengine/market/v2beta1/orders/list:
    get:
      summary: List marketplace orders
      tags:
        - Market
      # ...
```

## 🏷️ Custom Extensions

VirtEngine uses custom OpenAPI extensions for enhanced documentation:

### x-rate-limit

Defines rate limiting for endpoints:

```yaml
paths:
  /virtengine/veid/v1/scope/submit:
    post:
      x-rate-limit:
        requests-per-second: 5
        requests-per-minute: 100
        burst: 10
```

### x-auth-required

Specifies authentication requirements:

```yaml
paths:
  /virtengine/veid/v1/scope/submit:
    post:
      x-auth-required:
        type: wallet-signature
        mfa: required
```

### x-deprecation

Marks deprecated endpoints:

```yaml
paths:
  /virtengine/market/v1beta5/orders/list:
    get:
      x-deprecation:
        deprecated: true
        sunset: "2026-12-31"
        replacement: /virtengine/market/v2beta1/orders/list
```

## 🌐 Interactive Documentation

### Using Redoc

The enhanced Redoc viewer includes:
- Custom theme matching VirtEngine branding
- Code sample generation (cURL, JavaScript, Python, Go)
- Copy-to-clipboard for code samples
- Rate limit and auth callouts
- Deprecated endpoint warnings

**Features:**
```yaml
theme:
  colors:
    primary: '#3b82f6'  # VirtEngine blue
  typography:
    fontFamily: 'Roboto, sans-serif'
    headings:
      fontFamily: 'Montserrat, sans-serif'
```

### Using Swagger UI

Interactive API testing interface with:
- Try-it-out functionality
- Request/response validation
- Authentication testing
- Deep linking to specific operations

## 🔨 Generating Client Libraries

Use OpenAPI Generator to create client SDKs:

```bash
# Install OpenAPI Generator
npm install -g @openapitools/openapi-generator-cli

# Generate TypeScript client
openapi-generator-cli generate \
  -i api/openapi/virtengine-api.yaml \
  -g typescript-fetch \
  -o generated/typescript \
  --additional-properties=npmName=@virtengine/api-client

# Generate Go client
openapi-generator-cli generate \
  -i api/openapi/virtengine-api.yaml \
  -g go \
  -o generated/go \
  --git-repo-id=virtengine-api-client-go \
  --git-user-id=virtengine

# Generate Python client
openapi-generator-cli generate \
  -i api/openapi/virtengine-api.yaml \
  -g python \
  -o generated/python \
  --additional-properties=packageName=virtengine_api

# Generate Java client
openapi-generator-cli generate \
  -i api/openapi/virtengine-api.yaml \
  -g java \
  -o generated/java \
  --additional-properties=groupId=com.virtengine,artifactId=virtengine-api-client
```

### Available Generators

Run `openapi-generator-cli list` to see all available generators:

**Popular languages:**
- `typescript-fetch`, `typescript-axios`, `typescript-node`
- `go`, `go-gin-server`
- `python`, `python-flask`
- `java`, `kotlin`
- `rust`, `ruby`, `php`, `swift`
- `csharp`, `dart`

## ✅ Validation

Validate the OpenAPI specification:

```bash
# Using Redocly CLI (recommended)
npx @redocly/openapi-cli lint api/openapi/virtengine-api.yaml

# Using Swagger CLI
npx swagger-cli validate api/openapi/virtengine-api.yaml

# Using Spectral
npx @stoplight/spectral-cli lint api/openapi/virtengine-api.yaml
```

### Validation Rules

The specs follow:
- OpenAPI 3.0.3 specification
- RESTful API design principles
- Consistent naming conventions
- Complete schema definitions
- Proper use of HTTP status codes

## 📝 Adding Documentation

To add or improve documentation:

### 1. Update Proto Comments

Add detailed comments to proto files:

```protobuf
// Query defines the gRPC querier service for the market package.
// 
// This service provides read-only access to marketplace data including
// orders, bids, and leases. All query endpoints are publicly accessible
// without authentication.
service Query {
  // Orders queries orders with filters.
  //
  // Returns a paginated list of marketplace orders that match the
  // specified filter criteria. Orders can be filtered by owner, state,
  // and other attributes.
  //
  // Rate limit: 20 requests/second
  // Burst capacity: 50 requests
  rpc Orders(QueryOrdersRequest) returns (QueryOrdersResponse) {
    option (google.api.http).get = "/virtengine/market/v2beta1/orders/list";
  }
}
```

### 2. Regenerate from Proto

```bash
cd sdk
buf generate --template buf.gen.swagger.yaml
```

### 3. Manual Additions

For complex examples or descriptions not suitable for proto comments, edit the OpenAPI YAML directly:

```yaml
paths:
  /virtengine/market/v2beta1/orders/list:
    get:
      summary: List marketplace orders
      description: |
        Returns a paginated list of orders matching the specified filters.
        
        **Common use cases:**
        - List all open orders for resource discovery
        - Monitor orders from a specific tenant
        - Track order state transitions
        
        **Performance notes:**
        - Results are cached for 2 seconds at CDN edge
        - Use pagination for large result sets
        - Filter by owner when possible to reduce query time
      # ... rest of definition
```

### 4. Add Examples

Include request/response examples:

```yaml
responses:
  '200':
    description: List of orders
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/OrdersResponse'
        examples:
          open-orders:
            summary: Open orders
            value:
              orders:
                - orderId: { owner: "virtengine1...", dseq: "12345" }
                  state: "open"
                  # ...
```

## 🧪 Testing

### Manual Testing with Swagger UI

1. Open `swagger-ui.html`
2. Click "Authorize" and enter API key
3. Select an endpoint
4. Click "Try it out"
5. Fill in parameters
6. Click "Execute"

### Automated Testing

```bash
# Install Newman (Postman CLI)
npm install -g newman

# Convert OpenAPI to Postman collection
npx openapi-to-postman api/openapi/virtengine-api.yaml \
  -o tests/api/virtengine-api.postman.json

# Run tests
newman run tests/api/virtengine-api.postman.json \
  --environment tests/api/testnet.postman_environment.json
```

## 📚 See Also

- [API Reference Documentation](../reference/)
- [Getting Started Guide](../guides/getting-started.md)
- [Authentication Guide](../guides/authentication.md)
- [Rate Limits Guide](../guides/rate-limits.md)
- [Buf Documentation](https://docs.buf.build/)
- [OpenAPI Specification](https://spec.openapis.org/oas/v3.0.3)

## 🤝 Contributing

Improvements to the API documentation are welcome!

**Guidelines:**
1. Keep descriptions clear and concise
2. Include practical examples
3. Document rate limits and authentication
4. Note deprecations with replacement paths
5. Validate before committing

```bash
# Before committing
npx @redocly/openapi-cli lint api/openapi/virtengine-api.yaml
```

---

**Last updated**: February 2026
