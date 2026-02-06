# VirtEngine OpenAPI Specification

This directory contains OpenAPI/Swagger documentation for the VirtEngine API.

## Files

| File | Description |
|------|-------------|
| `virtengine-api.yaml` | Main OpenAPI 3.0 specification |
| `api/openapi/portal_api.yaml` | Provider portal OpenAPI 3.0 specification |
| `redoc.html` | Interactive documentation viewer |
| `swagger-config.json` | Swagger UI configuration |

## Generated Specifications

The main OpenAPI specification is auto-generated from proto files using `buf generate`:

```bash
# Generate swagger from proto files
cd sdk
buf generate --template buf.gen.swagger.yaml

# Merge generated files
cat .cache/tmp/swagger-gen/*.swagger.json | jq -s 'reduce .[] as $item ({}; . * $item)' > swagger-merged.json
```

The generated specification is available at:
- **Swagger UI**: `http://localhost:8000/portal` (via Kong gateway)
- **Static File**: `sdk/docs/swagger-ui/swagger.yaml`

## Provider Portal API

The provider portal OpenAPI spec lives at `api/openapi/portal_api.yaml`. Generated artifacts are:

- **TypeScript types**: `lib/portal/src/provider-api/generated/types.ts`
- **Go types**: `pkg/provider_daemon/api/generated/types.go`
- **Redoc HTML**: `docs/api/openapi/provider-portal.html`

Generate and validate with:

```bash
./scripts/generate-api-types.sh
```

## API Versioning in OpenAPI

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

## Custom Extensions

VirtEngine uses custom OpenAPI extensions:

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

## Interactive Documentation

### Using Redoc

Open `redoc.html` in a browser or serve it:

```bash
npx http-server -p 8080
# Open http://localhost:8080/redoc.html
```

### Using Swagger UI

The Swagger UI is available via the Kong API gateway:

```bash
# Start the gateway
docker-compose up -d kong swagger-ui

# Access at http://localhost:8000/portal
```

## Generating Client Libraries

Use OpenAPI Generator to create client SDKs:

```bash
# Install OpenAPI Generator
npm install -g @openapitools/openapi-generator-cli

# Generate TypeScript client
openapi-generator-cli generate \
  -i sdk/docs/swagger-ui/swagger.yaml \
  -g typescript-fetch \
  -o generated/typescript

# Generate Go client
openapi-generator-cli generate \
  -i sdk/docs/swagger-ui/swagger.yaml \
  -g go \
  -o generated/go

# Generate Python client
openapi-generator-cli generate \
  -i sdk/docs/swagger-ui/swagger.yaml \
  -g python \
  -o generated/python
```

## Validation

Validate the OpenAPI specification:

```bash
# Using openapi-cli
npx @redocly/openapi-cli lint sdk/docs/swagger-ui/swagger.yaml

# Using swagger-cli
npx swagger-cli validate sdk/docs/swagger-ui/swagger.yaml
```

## Adding Documentation

To add or improve documentation:

1. **Proto comments** - Add comments to proto files:

```protobuf
// Query defines the gRPC querier service for the market package.
// 
// This service provides read-only access to marketplace data including
// orders, bids, and leases.
service Query {
  // Orders queries orders with filters.
  //
  // Returns a paginated list of marketplace orders that match the
  // specified filter criteria.
  rpc Orders(QueryOrdersRequest) returns (QueryOrdersResponse) {
    option (google.api.http).get = "/virtengine/market/v2beta1/orders/list";
  }
}
```

2. **Regenerate** - Run buf generate to update swagger

3. **Manual additions** - For complex examples or descriptions not in proto, add to the merged specification

## See Also

- [API Reference](../reference/)
- [Getting Started](../guides/getting-started.md)
- [Buf Documentation](https://docs.buf.build/)
