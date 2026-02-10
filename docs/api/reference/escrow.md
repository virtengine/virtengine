# Escrow Module API Reference

The Escrow module manages payment accounts and withdrawals for deployments and leases.

## Overview

The Escrow module enables:
- Escrow account management
- Payment tracking
- Provider withdrawals
- Balance monitoring

## Base URL

```
/virtengine/escrow/v1
```

---

## Query Endpoints

### List Accounts

```http
GET /virtengine/escrow/v1/accounts/list
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `scope` | query | string | Yes | Scope (e.g., `deployment`) |
| `owner` | query | string | No | Filter by owner |
| `state` | query | string | No | Filter by state |

**Response:**

```json
{
  "accounts": [
    {
      "id": {
        "scope": "deployment",
        "xid": "virtengine1abc.../12345"
      },
      "owner": "virtengine1abc...",
      "state": "open",
      "balance": {
        "denom": "uvirt",
        "amount": "10000000"
      },
      "transferred": {
        "denom": "uvirt",
        "amount": "500000"
      },
      "settled_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### Get Account

```http
GET /virtengine/escrow/v1/accounts/info
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `id.scope` | query | string | Yes | Account scope |
| `id.xid` | query | string | Yes | Account external ID |

---

### List Payments

```http
GET /virtengine/escrow/v1/payments/list
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `scope` | query | string | Yes | Payment scope |
| `owner` | query | string | No | Filter by owner |

**Response:**

```json
{
  "payments": [
    {
      "account_id": {
        "scope": "deployment",
        "xid": "virtengine1abc.../12345"
      },
      "payment_id": "virtengine1provider.../1/1",
      "owner": "virtengine1abc...",
      "state": "open",
      "rate": {
        "denom": "uvirt",
        "amount": "950"
      },
      "balance": {
        "denom": "uvirt",
        "amount": "5000000"
      },
      "withdrawn": {
        "denom": "uvirt",
        "amount": "100000"
      }
    }
  ]
}
```

---

## Account States

| State | Description |
|-------|-------------|
| `open` | Account is active |
| `closed` | Account is closed |
| `overdrawn` | Account has insufficient funds |

---

## Error Codes

| Code | Message | Category |
|------|---------|----------|
| escrow:1500 | Invalid account | validation |
| escrow:1501 | Account not found | not_found |
| escrow:1510 | Insufficient funds | validation |
| escrow:1520 | Unauthorized | unauthorized |

---

## See Also

- [Market Module](./market.md) - Lease management
- [Deployment Module](./deployment.md) - Deployment funding
