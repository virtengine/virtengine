# VirtEngine Vault Policy
# Grants access to secrets for VirtEngine services

# Read-only access to common secrets
path "secret/data/virtengine/common/*" {
  capabilities = ["read"]
}

# Environment-specific secrets (dev, staging, prod)
path "secret/data/virtengine/{{identity.entity.metadata.environment}}/*" {
  capabilities = ["read"]
}

# Node-specific secrets
path "secret/data/virtengine/{{identity.entity.metadata.environment}}/node/*" {
  capabilities = ["read"]
}

# Provider-specific secrets
path "secret/data/virtengine/{{identity.entity.metadata.environment}}/provider/*" {
  capabilities = ["read"]
}

# Validator key access (restricted)
path "secret/data/virtengine/{{identity.entity.metadata.environment}}/validator/*" {
  capabilities = ["read"]
  required_parameters = ["version"]
  allowed_parameters = {
    "version" = ["1", "2", "3"]
  }
}

# Transit engine for encryption operations
path "transit/encrypt/virtengine-{{identity.entity.metadata.environment}}" {
  capabilities = ["update"]
}

path "transit/decrypt/virtengine-{{identity.entity.metadata.environment}}" {
  capabilities = ["update"]
}

# PKI for certificate generation
path "pki/issue/virtengine-{{identity.entity.metadata.environment}}" {
  capabilities = ["create", "update"]
}

path "pki/cert/*" {
  capabilities = ["read"]
}
