// Package provider_daemon implements the provider daemon for VirtEngine.
//
// VE-400 through VE-404: Provider Daemon Implementation
// VE-913: OpenStack Adapter (Gap Phase)
// VE-920: Ansible Automation with Waldur Integration (Gap Phase)
//
// This package provides:
//   - Key management and transaction signing (VE-400)
//   - Bid engine and configuration watcher (VE-401)
//   - Manifest parsing and validation (VE-402)
//   - Kubernetes orchestration adapter (VE-403)
//   - Usage metering and on-chain recording (VE-404)
//   - OpenStack orchestration adapter via Waldur (VE-913)
//   - Ansible playbook execution with vault secrets integration (VE-920)
//
// Ansible Adapter (VE-920):
//   - Playbook execution with inventory support
//   - Variable injection with Ansible Vault encryption/decryption
//   - Async execution with status tracking
//   - Deployment playbooks for VirtEngine components
//   - Secure handling of vault passwords (never logged)
//
// Security Properties:
//   - Provider keys support hardware/ledger/non-custodial storage
//   - All on-chain submissions are cryptographically signed
//   - Secrets are never logged or stored in plaintext
//   - Memory is scrubbed after handling sensitive data
//   - Vault passwords are cleared from memory after use
//   - Log redaction for sensitive data patterns
package provider_daemon

