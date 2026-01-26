# VirtEngine Deployment Playbooks

This directory contains Ansible playbooks for deploying and managing VirtEngine components.

## Playbook Overview

| Playbook | Description | Tags |
|----------|-------------|------|
| `deploy_virtengine_node.yml` | Deploy VirtEngine blockchain node | install, configure, service |
| `deploy_provider_daemon.yml` | Deploy Provider Daemon | install, configure, service |
| `deploy_portal.yml` | Deploy VirtEngine Portal | install, configure, nginx |
| `configure_ssl.yml` | Configure SSL/TLS certificates | ssl, nginx |
| `maintenance.yml` | System maintenance tasks | update, backup, cleanup |

## Usage

### Basic Execution

```bash
ansible-playbook -i inventory.ini deploy_virtengine_node.yml
```

### With Vault Encrypted Variables

```bash
ansible-playbook -i inventory.ini deploy_virtengine_node.yml --ask-vault-pass
```

### Using Specific Tags

```bash
ansible-playbook -i inventory.ini deploy_virtengine_node.yml --tags "install,configure"
```

### Check Mode (Dry Run)

```bash
ansible-playbook -i inventory.ini deploy_virtengine_node.yml --check
```

## Required Variables

### deploy_virtengine_node.yml
- `virtengine_version`: Version to deploy
- `chain_id`: Chain ID for the network
- `moniker`: Node moniker
- `seeds`: Seed node addresses

### deploy_provider_daemon.yml
- `provider_address`: Provider's blockchain address
- `provider_key_file`: Path to provider key file (vault encrypted)
- `virtengine_rpc`: VirtEngine RPC endpoint

### deploy_portal.yml
- `portal_domain`: Domain for the portal
- `api_endpoint`: Backend API endpoint
- `ssl_certificate`: Path to SSL certificate
- `ssl_private_key`: Path to SSL private key (vault encrypted)

## Security Notes

1. **Never commit vault passwords** to version control
2. Use vault-encrypted variables for all secrets
3. Restrict SSH key access to deployment hosts
4. Use `become` only when necessary
5. Review playbooks before execution in production
