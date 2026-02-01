// Package servicedesk provides bi-directional integration between VirtEngine
// on-chain support tickets and external service desk systems (Jira/Waldur).
//
// VE-12B: Jira/Waldur Service Desk Integration
//
// This package implements:
//   - Mapping schema from on-chain ticket fields to Jira/Waldur fields
//   - Bridge service that listens to chain events and creates external tickets
//   - Bi-directional status sync with signed callbacks
//   - Waldur Support Issue API integration for ticket management
//   - Jira Service Desk integration with full CRUD operations
//   - Attachment handling via artifact store with access controls
//   - Conflict resolution for concurrent updates
//   - Audit trail of external actions
//   - Retry and backoff for API failures
//   - Chain event listener for automatic synchronization
//
// Architecture:
//
//	On-Chain (x/support)     Service Desk Bridge      External Systems
//	     │                         │                        │
//	     ├──ticket created────────►├──────create issue─────►│ Jira
//	     ├──ticket updated────────►├──────update issue─────►│ Waldur
//	     ├──ticket closed─────────►├──────close issue──────►│
//	     │                         │                        │
//	     │◄──update status─────────┤◄─────webhook callback──┤
//	     │                         │                        │
//
// Configuration is loaded from environment variables or config file.
// See Config struct for all available options.
//
// CRITICAL: Never log API tokens, webhook secrets, or sensitive ticket content.
package servicedesk

