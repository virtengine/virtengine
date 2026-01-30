# Provider Module API Reference

The Provider module manages provider registration, attributes, and lifecycle.

## Overview

The Provider module enables:
- Provider registration and updates
- Attribute management
- Provider discovery
- Audit integration

## Base URL

```
/virtengine/provider/v1beta4
```

## Authentication Requirements

| Operation Type | Authentication |
|----------------|----------------|
| Query endpoints | None required |
| Provider registration | Wallet signature + MFA |
| Provider updates | Provider key + MFA |

---

## Query Endpoints

### List Providers

Queries all registered providers.

```http
GET /virtengine/provider/v1beta4/providers/list
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `pagination.limit` | query | integer | No | Max results |
| `pagination.key` | query | string | No | Pagination key |

**Response:**

```json
{
  "providers": [
    {
      "owner": "virtengine1provider...",
      "host_uri": "https://provider.example.com:8443",
      "attributes": [
        { "key": "region", "value": "us-west-2" },
        { "key": "organization", "value": "Example Corp" }
      ],
      "info": {
        "email": "provider@example.com",
        "website": "https://example.com"
      }
    }
  ],
  "pagination": {
    "next_key": "base64...",
    "total": "50"
  }
}
```

**Example:**

```bash
curl "https://api.virtengine.com/virtengine/provider/v1beta4/providers/list"
```

---

### Get Provider

Queries a specific provider.

```http
GET /virtengine/provider/v1beta4/providers/info
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `owner` | query | string | Yes | Provider address |

**Response:**

```json
{
  "provider": {
    "owner": "virtengine1provider...",
    "host_uri": "https://provider.example.com:8443",
    "attributes": [
      { "key": "region", "value": "us-west-2" },
      { "key": "tier", "value": "enterprise" }
    ],
    "info": {
      "email": "provider@example.com",
      "website": "https://example.com"
    }
  }
}
```

---

## Transaction Messages

### Create Provider

Registers a new provider.

```protobuf
message MsgCreateProvider {
  string owner = 1;
  string host_uri = 2;
  repeated Attribute attributes = 3;
  ProviderInfo info = 4;
}
```

**Example (CLI):**

```bash
virtengine tx provider create \
  --host="https://provider.example.com:8443" \
  --attributes="region=us-west-2,tier=enterprise" \
  --from provider-key \
  --chain-id virtengine-1
```

---

### Update Provider

Updates provider information.

```protobuf
message MsgUpdateProvider {
  string owner = 1;
  string host_uri = 2;
  repeated Attribute attributes = 3;
  ProviderInfo info = 4;
}
```

---

### Delete Provider

Removes a provider registration.

```protobuf
message MsgDeleteProvider {
  string owner = 1;
}
```

---

## Provider Attributes

Common provider attributes:

| Key | Description | Example |
|-----|-------------|---------|
| `region` | Geographic region | `us-west-2` |
| `organization` | Organization name | `Example Corp` |
| `tier` | Service tier | `enterprise` |
| `cpu_arch` | CPU architecture | `amd64`, `arm64` |
| `gpu_vendor` | GPU vendor | `nvidia`, `amd` |
| `gpu_model` | GPU model | `a100`, `v100` |
| `storage_type` | Storage type | `ssd`, `nvme` |
| `network_speed` | Network speed | `10gbps` |

---

## Error Codes

| Code | Message | Category |
|------|---------|----------|
| provider:1800 | Invalid provider | validation |
| provider:1801 | Provider not found | not_found |
| provider:1802 | Provider exists | conflict |
| provider:1810 | Invalid host URI | validation |
| provider:1820 | Unauthorized | unauthorized |

---

## See Also

- [Market Module](./market.md) - Marketplace operations
- [Audit Module](./audit.md) - Provider auditing
- [Provider Daemon](./provider-daemon.md) - Provider services
