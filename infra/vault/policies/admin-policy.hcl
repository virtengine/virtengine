# Admin policy for VirtEngine infrastructure management
# Use with caution - grants broad access

# Full access to VirtEngine secrets
path "secret/data/virtengine/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/metadata/virtengine/*" {
  capabilities = ["read", "list", "delete"]
}

# Transit key management
path "transit/keys/virtengine-*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "transit/encrypt/virtengine-*" {
  capabilities = ["update"]
}

path "transit/decrypt/virtengine-*" {
  capabilities = ["update"]
}

# PKI management
path "pki/roles/virtengine-*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "pki/issue/virtengine-*" {
  capabilities = ["create", "update"]
}

path "pki/cert/*" {
  capabilities = ["read"]
}

path "pki/revoke" {
  capabilities = ["create", "update"]
}

# Auth method configuration
path "auth/kubernetes/role/virtengine-*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Policy management (read-only)
path "sys/policies/acl/virtengine-*" {
  capabilities = ["read", "list"]
}

# Audit log access
path "sys/audit" {
  capabilities = ["read", "list"]
}
