// Package servicedesk provides bi-directional integration between VirtEngine
// on-chain support tickets and external service desk systems (Jira/Waldur).
//
// VE-3E: Jira/Waldur Service Desk Sync
//
// This package implements:
//   - Mapping schema from on-chain ticket fields to Jira/Waldur fields
//   - Bridge service that listens to chain events and creates external tickets
//   - Bi-directional status sync with signed callbacks
//   - Attachment handling via artifact store with access controls
//   - Conflict resolution for concurrent updates
//   - Audit trail of external actions
//   - Retry and backoff for API failures
//
// CRITICAL: Never log API tokens, webhook secrets, or sensitive ticket content.
package servicedesk
