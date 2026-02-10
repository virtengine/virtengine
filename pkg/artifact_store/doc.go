// Package artifact_store provides a scalable storage strategy for identity artifacts
// that respects blockchain size constraints: encrypted off-chain storage with on-chain
// content-addressed references.
//
// The package implements VE-218 from the VirtEngine PRD, supporting:
//   - Artifact store interface with put/get by content hash
//   - Encryption envelope integration for secure artifact storage
//   - Retention tags for lifecycle management
//   - Multiple backend support: Waldur DB and IPFS
//   - Deterministic chunk reconstruction for fragmented artifacts
//
// Architecture:
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                     Artifact Store Interface                     │
//	├─────────────────────────────────────────────────────────────────┤
//	│  Put(artifact) → ContentAddress                                  │
//	│  Get(address) → Artifact                                         │
//	│  Delete(address) → error                                         │
//	│  Exists(address) → bool                                          │
//	└───────────────────────┬─────────────────────────────────────────┘
//	                        │
//	         ┌──────────────┴──────────────┐
//	         │                             │
//	         ▼                             ▼
//	┌─────────────────┐           ┌─────────────────┐
//	│  Waldur Backend │           │   IPFS Backend  │
//	│   (Encrypted    │           │  (Encrypted     │
//	│    at rest)     │           │   chunks + CID) │
//	└─────────────────┘           └─────────────────┘
//
// Security Considerations:
//   - Raw biometric images are NEVER stored on-chain
//   - All artifacts are encrypted before off-chain storage
//   - Only content-addressed hashes are stored on-chain
//   - Encryption envelopes ensure only authorized recipients can decrypt
//   - Retention tags enable compliance with data minimization requirements
//
// See _docs/architecture.md for full system architecture.
package artifact_store
