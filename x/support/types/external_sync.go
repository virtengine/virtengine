package types

import (
	"fmt"
	"time"
)

// External sync store key prefixes
var (
	// PrefixExternalTicketRef is the prefix for external ticket references
	// Key: PrefixExternalTicketRef | ticket_id | service_desk_type -> ExternalTicketRef
	PrefixExternalTicketRef = []byte{0x10}

	// PrefixSyncRecord is the prefix for sync records
	// Key: PrefixSyncRecord | ticket_id -> TicketSyncRecord
	PrefixSyncRecord = []byte{0x11}

	// PrefixExternalCallback is the prefix for external callback nonce tracking
	// Key: PrefixExternalCallback | nonce -> timestamp
	PrefixExternalCallback = []byte{0x12}

	// PrefixSyncAudit is the prefix for sync audit log
	// Key: PrefixSyncAudit | timestamp | sequence -> SyncAuditEntry
	PrefixSyncAudit = []byte{0x13}
)

// ServiceDeskType identifies the external service desk system
type ServiceDeskType string

const (
	// ServiceDeskJira represents Jira Service Desk
	ServiceDeskJira ServiceDeskType = "jira"

	// ServiceDeskWaldur represents Waldur service desk
	ServiceDeskWaldur ServiceDeskType = "waldur"
)

// IsValid checks if the service desk type is valid
func (t ServiceDeskType) IsValid() bool {
	return t == ServiceDeskJira || t == ServiceDeskWaldur
}

// ExternalSyncStatus represents the synchronization status
type ExternalSyncStatus string

const (
	// ExternalSyncPending indicates sync is pending
	ExternalSyncPending ExternalSyncStatus = "pending"

	// ExternalSyncSynced indicates successful sync
	ExternalSyncSynced ExternalSyncStatus = "synced"

	// ExternalSyncFailed indicates sync failed
	ExternalSyncFailed ExternalSyncStatus = "failed"

	// ExternalSyncConflict indicates a conflict was detected
	ExternalSyncConflict ExternalSyncStatus = "conflict"
)

// ExternalTicketRef represents a reference to an external service desk ticket
type ExternalTicketRef struct {
	// TicketID is the on-chain ticket ID
	TicketID string `json:"ticket_id"`

	// ServiceDeskType is the external service desk type
	ServiceDeskType ServiceDeskType `json:"service_desk_type"`

	// ExternalID is the external ticket ID (e.g., JIRA-123)
	ExternalID string `json:"external_id"`

	// ExternalURL is the URL to the external ticket
	ExternalURL string `json:"external_url"`

	// ProjectKey is the external project key
	ProjectKey string `json:"project_key"`

	// SyncStatus is the current sync status
	SyncStatus ExternalSyncStatus `json:"sync_status"`

	// LastSyncAt is the last successful sync timestamp
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`

	// LastSyncError is the last sync error message
	LastSyncError string `json:"last_sync_error,omitempty"`

	// SyncVersion is used for optimistic locking
	SyncVersion int64 `json:"sync_version"`

	// CreatedAt is when the external ticket was created
	CreatedAt time.Time `json:"created_at"`

	// LastExternalUpdate is the last external update timestamp
	LastExternalUpdate *time.Time `json:"last_external_update,omitempty"`
}

// Validate validates the external ticket reference
func (r *ExternalTicketRef) Validate() error {
	if r.TicketID == "" {
		return ErrInvalidTicketID.Wrap("ticket_id is required")
	}
	if !r.ServiceDeskType.IsValid() {
		return ErrInvalidSyncConfig.Wrapf("invalid service desk type: %s", r.ServiceDeskType)
	}
	if r.ExternalID == "" {
		return ErrInvalidSyncConfig.Wrap("external_id is required")
	}
	return nil
}

// TicketSyncRecord tracks the complete sync state for a ticket
type TicketSyncRecord struct {
	// TicketID is the on-chain ticket ID
	TicketID string `json:"ticket_id"`

	// ExternalRefs are references to external tickets
	ExternalRefs []ExternalTicketRef `json:"external_refs"`

	// LastOnChainUpdate is the last on-chain update block height
	LastOnChainBlockHeight int64 `json:"last_on_chain_block_height"`

	// LastOnChainTxHash is the last on-chain transaction hash
	LastOnChainTxHash string `json:"last_on_chain_tx_hash"`

	// PendingSync indicates if there's a pending sync operation
	PendingSync bool `json:"pending_sync"`

	// PendingSyncDirection indicates the pending sync direction
	PendingSyncDirection string `json:"pending_sync_direction,omitempty"`

	// ConflictResolution is the conflict resolution strategy
	ConflictResolution string `json:"conflict_resolution"`

	// CreatedAt is when the sync record was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the sync record was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// GetExternalRef returns the external reference for the given service desk type
func (r *TicketSyncRecord) GetExternalRef(t ServiceDeskType) *ExternalTicketRef {
	for i := range r.ExternalRefs {
		if r.ExternalRefs[i].ServiceDeskType == t {
			return &r.ExternalRefs[i]
		}
	}
	return nil
}

// Validate validates the sync record
func (r *TicketSyncRecord) Validate() error {
	if r.TicketID == "" {
		return ErrInvalidTicketID.Wrap("ticket_id is required")
	}
	for _, ref := range r.ExternalRefs {
		if err := ref.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// SyncAuditEntry represents an entry in the sync audit log
type SyncAuditEntry struct {
	// ID is the unique entry ID
	ID string `json:"id"`

	// TicketID is the on-chain ticket ID
	TicketID string `json:"ticket_id"`

	// ServiceDeskType is the external service desk type
	ServiceDeskType ServiceDeskType `json:"service_desk_type"`

	// ExternalID is the external ticket ID
	ExternalID string `json:"external_id"`

	// EventType is the type of sync event
	EventType string `json:"event_type"`

	// Direction is the sync direction (inbound/outbound)
	Direction string `json:"direction"`

	// Status is the sync status
	Status ExternalSyncStatus `json:"status"`

	// Details contains additional event details
	Details string `json:"details,omitempty"`

	// Error contains error information
	Error string `json:"error,omitempty"`

	// BlockHeight is the on-chain block height
	BlockHeight int64 `json:"block_height"`

	// TxHash is the related transaction hash
	TxHash string `json:"tx_hash,omitempty"`

	// Timestamp is when the event occurred
	Timestamp time.Time `json:"timestamp"`
}

// Validate validates the sync audit entry
func (e *SyncAuditEntry) Validate() error {
	if e.TicketID == "" {
		return ErrInvalidTicketID.Wrap("ticket_id is required")
	}
	if e.EventType == "" {
		return ErrInvalidSyncConfig.Wrap("event_type is required")
	}
	return nil
}

// ExternalTicketRefKey returns the store key for an external ticket reference
func ExternalTicketRefKey(ticketID string, serviceDeskType ServiceDeskType) []byte {
	key := make([]byte, 0, len(PrefixExternalTicketRef)+len(ticketID)+len(serviceDeskType)+2)
	key = append(key, PrefixExternalTicketRef...)
	key = append(key, []byte(ticketID)...)
	key = append(key, '/')
	key = append(key, []byte(serviceDeskType)...)
	return key
}

// ExternalTicketRefPrefixKey returns the prefix for a ticket's external references
func ExternalTicketRefPrefixKey(ticketID string) []byte {
	key := make([]byte, 0, len(PrefixExternalTicketRef)+len(ticketID)+1)
	key = append(key, PrefixExternalTicketRef...)
	key = append(key, []byte(ticketID)...)
	key = append(key, '/')
	return key
}

// SyncRecordKey returns the store key for a sync record
func SyncRecordKey(ticketID string) []byte {
	key := make([]byte, 0, len(PrefixSyncRecord)+len(ticketID))
	key = append(key, PrefixSyncRecord...)
	key = append(key, []byte(ticketID)...)
	return key
}

// ExternalCallbackNonceKey returns the store key for callback nonce tracking
func ExternalCallbackNonceKey(nonce string) []byte {
	key := make([]byte, 0, len(PrefixExternalCallback)+len(nonce))
	key = append(key, PrefixExternalCallback...)
	key = append(key, []byte(nonce)...)
	return key
}

// Proto message interface stubs
func (*ExternalTicketRef) ProtoMessage()   {}
func (r *ExternalTicketRef) Reset()        { *r = ExternalTicketRef{} }
func (r *ExternalTicketRef) String() string { return fmt.Sprintf("%+v", *r) }

func (*TicketSyncRecord) ProtoMessage()    {}
func (r *TicketSyncRecord) Reset()         { *r = TicketSyncRecord{} }
func (r *TicketSyncRecord) String() string { return fmt.Sprintf("%+v", *r) }

func (*SyncAuditEntry) ProtoMessage()      {}
func (e *SyncAuditEntry) Reset()           { *e = SyncAuditEntry{} }
func (e *SyncAuditEntry) String() string   { return fmt.Sprintf("%+v", *e) }
