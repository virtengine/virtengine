package gdpr

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// Public Types
// =============================================================================

// ExportStatus tracks the lifecycle of a data export request.
type ExportStatus string

const (
	ExportPending    ExportStatus = "pending"
	ExportProcessing ExportStatus = "processing"
	ExportReady      ExportStatus = "ready"
	ExportFailed     ExportStatus = "failed"
	ExportExpired    ExportStatus = "expired"
)

// ExportFormat defines the export data format.
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
)

// DataExportRequest represents a GDPR export request.
type DataExportRequest struct {
	ID          string       `json:"id"`
	DataSubject string       `json:"data_subject"`
	RequestedAt time.Time    `json:"requested_at"`
	Status      ExportStatus `json:"status"`
	Format      ExportFormat `json:"format"`
	DownloadURL string       `json:"download_url,omitempty"`
	ExpiresAt   *time.Time   `json:"expires_at,omitempty"`
	Error       string       `json:"error,omitempty"`
}

// DeletionStatus tracks the lifecycle of a deletion request.
type DeletionStatus string

const (
	DeletionPending    DeletionStatus = "pending"
	DeletionBlocked    DeletionStatus = "blocked"
	DeletionProcessing DeletionStatus = "processing"
	DeletionComplete   DeletionStatus = "complete"
	DeletionFailed     DeletionStatus = "failed"
)

// DeletionRequest represents a GDPR deletion request.
type DeletionRequest struct {
	ID          string         `json:"id"`
	DataSubject string         `json:"data_subject"`
	RequestedAt time.Time      `json:"requested_at"`
	Status      DeletionStatus `json:"status"`
	Blockers    []string       `json:"blockers,omitempty"`
	Error       string         `json:"error,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// UserDataExport contains the aggregated export payload.
type UserDataExport struct {
	ExportedAt    time.Time           `json:"exported_at"`
	DataSubject   string              `json:"data_subject"`
	Identity      *IdentityExport     `json:"identity,omitempty"`
	Consents      []ConsentExport     `json:"consents,omitempty"`
	Transactions  []TransactionExport `json:"transactions,omitempty"`
	EscrowRecords []EscrowExport      `json:"escrow_records,omitempty"`
}

// =============================================================================
// Interfaces
// =============================================================================

// ConsentProvider exposes consent history for export/deletion.
type ConsentProvider interface {
	ListConsents(ctx context.Context, dataSubject string) ([]ConsentExport, error)
	DeleteConsentRecords(ctx context.Context, dataSubject string) error
}

// IdentityProvider exposes VEID identity records for export/deletion.
type IdentityProvider interface {
	GetIdentity(ctx context.Context, dataSubject string) (*IdentityExport, error)
	DeleteIdentity(ctx context.Context, dataSubject string) error
}

// EscrowProvider exposes escrow data for export/deletion checks.
type EscrowProvider interface {
	GetUserEscrows(ctx context.Context, dataSubject string) ([]EscrowExport, error)
	HasActiveEscrow(ctx context.Context, dataSubject string) (bool, error)
}

// ExportStore persists export and deletion requests.
type ExportStore interface {
	SaveExportRequest(ctx context.Context, req *DataExportRequest) error
	SaveDeletionRequest(ctx context.Context, req *DeletionRequest) error
	SaveExportData(ctx context.Context, requestID string, data []byte) error
}

// Notifier delivers export-ready notifications.
type Notifier interface {
	NotifyExportReady(ctx context.Context, req *DataExportRequest) error
}

// AuditLogger captures audit trail events for GDPR actions.
type AuditLogger interface {
	Log(ctx context.Context, event AuditEvent) error
}

// =============================================================================
// DataRightsService
// =============================================================================

// DataRightsService orchestrates GDPR export and deletion workflows.
type DataRightsService struct {
	consents ConsentProvider
	identity IdentityProvider
	escrow   EscrowProvider
	store    ExportStore
	notifier Notifier
	audit    AuditLogger
	clock    func() time.Time
	mu       sync.Mutex
}

// NewDataRightsService constructs a DataRightsService.
func NewDataRightsService(store ExportStore, consents ConsentProvider, identity IdentityProvider, escrow EscrowProvider, notifier Notifier, audit AuditLogger) *DataRightsService {
	if notifier == nil {
		notifier = noopNotifier{}
	}
	if audit == nil {
		audit = noopAuditLogger{}
	}
	return &DataRightsService{
		consents: consents,
		identity: identity,
		escrow:   escrow,
		store:    store,
		notifier: notifier,
		audit:    audit,
		clock:    time.Now,
	}
}

// RequestDataExport initiates a data export request (Right to Access/Portability).
func (s *DataRightsService) RequestDataExport(ctx context.Context, dataSubject string, format ExportFormat) (*DataExportRequest, error) {
	if dataSubject == "" {
		return nil, errors.New("data subject required")
	}
	if format == "" {
		format = FormatJSON
	}

	req := &DataExportRequest{
		ID:          generateRequestID(dataSubject),
		DataSubject: dataSubject,
		RequestedAt: s.clock(),
		Status:      ExportPending,
		Format:      format,
	}

	if err := s.store.SaveExportRequest(ctx, req); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, AuditEvent{
		Action:      AuditExportRequested,
		RequestID:   req.ID,
		DataSubject: dataSubject,
		Timestamp:   req.RequestedAt,
	})

	go s.processExport(ctx, req)
	return req, nil
}

// RequestDataDeletion initiates a deletion request (Right to be Forgotten).
func (s *DataRightsService) RequestDataDeletion(ctx context.Context, dataSubject string) (*DeletionRequest, error) {
	if dataSubject == "" {
		return nil, errors.New("data subject required")
	}

	req := &DeletionRequest{
		ID:          generateRequestID(dataSubject),
		DataSubject: dataSubject,
		RequestedAt: s.clock(),
		Status:      DeletionPending,
	}

	blockers, err := s.checkDeletionBlockers(ctx, dataSubject)
	if err != nil {
		req.Status = DeletionFailed
		req.Error = err.Error()
		_ = s.store.SaveDeletionRequest(ctx, req)
		return req, err
	}

	if len(blockers) > 0 {
		req.Status = DeletionBlocked
		req.Blockers = blockers
		if err := s.store.SaveDeletionRequest(ctx, req); err != nil {
			return nil, err
		}
		_ = s.audit.Log(ctx, AuditEvent{
			Action:      AuditDeletionBlocked,
			RequestID:   req.ID,
			DataSubject: dataSubject,
			Timestamp:   req.RequestedAt,
			Details: map[string]string{
				"blockers": fmt.Sprintf("%v", blockers),
			},
		})
		return req, nil
	}

	req.Status = DeletionProcessing
	if err := s.store.SaveDeletionRequest(ctx, req); err != nil {
		return nil, err
	}

	if err := s.executeDeletion(ctx, dataSubject); err != nil {
		req.Status = DeletionFailed
		req.Error = err.Error()
		_ = s.store.SaveDeletionRequest(ctx, req)
		_ = s.audit.Log(ctx, AuditEvent{
			Action:      AuditDeletionFailed,
			RequestID:   req.ID,
			DataSubject: dataSubject,
			Timestamp:   s.clock(),
			Details: map[string]string{
				"error": err.Error(),
			},
		})
		return req, err
	}

	now := s.clock()
	req.Status = DeletionComplete
	req.CompletedAt = &now
	if err := s.store.SaveDeletionRequest(ctx, req); err != nil {
		return nil, err
	}
	_ = s.audit.Log(ctx, AuditEvent{
		Action:      AuditDeletionCompleted,
		RequestID:   req.ID,
		DataSubject: dataSubject,
		Timestamp:   now,
	})

	return req, nil
}

func generateRequestID(dataSubject string) string {
	return fmt.Sprintf("gdpr-%s-%d", dataSubject, time.Now().UnixNano())
}

type noopNotifier struct{}

func (noopNotifier) NotifyExportReady(_ context.Context, _ *DataExportRequest) error {
	return nil
}

type noopAuditLogger struct{}

func (noopAuditLogger) Log(_ context.Context, _ AuditEvent) error {
	return nil
}
