// Package data_vault provides a unified, off-chain encrypted data vault service
// for all sensitive payloads referenced by on-chain records.
//
// The vault implements:
// - Envelope encryption using X25519-XSalsa20-Poly1305
// - Key rotation with backward-compatible decryption
// - Scope-based access control (VEID, support, market, audit)
// - Audit trails for all decrypt operations
//
// Architecture:
// - VaultService: Main service interface for encrypted blob CRUD
// - EncryptedBlobStore: Storage layer wrapping artifact_store with encryption
// - KeyManager: DEK/KEK hierarchy with rotation support
// - AccessControl: Wallet-scoped auth + role/org permissions
// - AuditLogger: Immutable audit trail for compliance
//
// Security Properties:
// - All sensitive data encrypted at rest with unique DEKs
// - DEKs wrapped with KEKs (Key Encryption Keys)
// - On-chain references contain only blob IDs + content hashes
// - Cross-org access strictly denied
// - Every decrypt operation logged with full context
//
// Task Reference: 32A - Encrypted data vault + key rotation
package data_vault
