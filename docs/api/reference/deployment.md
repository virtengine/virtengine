# Deployment Module API Reference

The Deployment module manages compute deployments on the VirtEngine network.

## Overview

The Deployment module enables:
- Deployment creation and management
- Group resource specification
- Deployment lifecycle management
- Escrow account integration

## Base URL

```
/virtengine/deployment/v1beta5
```

---

## Query Endpoints

### List Deployments

```http
GET /virtengine/deployment/v1beta5/deployments/list
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `filters.owner` | query | string | No | Filter by owner |
| `filters.state` | query | string | No | Filter by state |
| `pagination.limit` | query | integer | No | Max results |

**Response:**

```json
{
  "deployments": [
    {
      "deployment_id": {
        "owner": "virtengine1abc...",
        "dseq": "12345"
      },
      "state": "active",
      "version": "v1.0.0",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "next_key": "base64...",
    "total": "10"
  }
}
```

---

### Get Deployment

```http
GET /virtengine/deployment/v1beta5/deployments/info
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `id.owner` | query | string | Yes | Owner address |
| `id.dseq` | query | string | Yes | Deployment sequence |

---

### Get Group

```http
GET /virtengine/deployment/v1beta5/groups/info
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `id.owner` | query | string | Yes | Owner address |
| `id.dseq` | query | string | Yes | Deployment sequence |
| `id.gseq` | query | integer | Yes | Group sequence |

---

## Transaction Messages

### Create Deployment

```protobuf
message MsgCreateDeployment {
  DeploymentID id = 1;
  repeated GroupSpec groups = 2;
  bytes version = 3;
  cosmos.base.v1beta1.Coin deposit = 4;
  string depositor = 5;
}
```

**Example (CLI):**

```bash
virtengine tx deployment create deployment.yaml \
  --deposit=5000000uvirt \
  --from mykey \
  --chain-id virtengine-1
```

---

### Update Deployment

```protobuf
message MsgUpdateDeployment {
  DeploymentID id = 1;
  repeated GroupSpec groups = 2;
  bytes version = 3;
}
```

---

### Close Deployment

```protobuf
message MsgCloseDeployment {
  DeploymentID id = 1;
}
```

---

### Deposit to Deployment

```protobuf
message MsgDepositDeployment {
  DeploymentID id = 1;
  cosmos.base.v1beta1.Coin amount = 2;
  string depositor = 3;
}
```

---

## Deployment States

| State | Description |
|-------|-------------|
| `active` | Deployment is active |
| `closed` | Deployment is closed |

## Group States

| State | Description |
|-------|-------------|
| `open` | Group is accepting bids |
| `paused` | Group is paused |
| `insufficient_funds` | Escrow depleted |
| `closed` | Group is closed |

---

## Error Codes

| Code | Message | Category |
|------|---------|----------|
| deployment:1900 | Invalid deployment | validation |
| deployment:1901 | Deployment not found | not_found |
| deployment:1902 | Deployment exists | conflict |
| deployment:1910 | Invalid group spec | validation |

---

## See Also

- [Market Module](./market.md) - Order and lease management
- [Escrow Module](./escrow.md) - Payment handling
