package gdpr

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// IdentityExport captures a minimal identity snapshot for export.
type IdentityExport struct {
	Address    string    `json:"address"`
	TrustScore float64   `json:"trust_score,omitempty"`
	TierLevel  string    `json:"tier_level,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

// ConsentExport captures a consent snapshot for export.
type ConsentExport struct {
	ID          string     `json:"id"`
	ScopeID     string     `json:"scope_id,omitempty"`
	Purpose     string     `json:"purpose,omitempty"`
	Status      string     `json:"status,omitempty"`
	GrantedAt   *time.Time `json:"granted_at,omitempty"`
	WithdrawnAt *time.Time `json:"withdrawn_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Version     string     `json:"version,omitempty"`
}

// TransactionExport represents a transaction entry in export output.
type TransactionExport struct {
	TxHash    string    `json:"tx_hash"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details,omitempty"`
}

// EscrowExport represents an escrow record in export output.
type EscrowExport struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Amount    string    `json:"amount,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// SupportAuditExport captures audit trail entries for support tickets.
type SupportAuditExport struct {
	Action      string    `json:"action"`
	PerformedBy string    `json:"performed_by"`
	Details     string    `json:"details,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	BlockHeight int64     `json:"block_height,omitempty"`
}

// SupportRetentionExport captures retention policy metadata.
type SupportRetentionExport struct {
	ArchiveAfterSeconds int64     `json:"archive_after_seconds,omitempty"`
	PurgeAfterSeconds   int64     `json:"purge_after_seconds,omitempty"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
	CreatedAtBlock      int64     `json:"created_at_block,omitempty"`
}

// SupportExport captures support ticket data for export.
type SupportExport struct {
	TicketID       string                  `json:"ticket_id"`
	TicketNumber   string                  `json:"ticket_number,omitempty"`
	Submitter      string                  `json:"submitter,omitempty"`
	Category       string                  `json:"category,omitempty"`
	Priority       string                  `json:"priority,omitempty"`
	Status         string                  `json:"status,omitempty"`
	CreatedAt      time.Time               `json:"created_at,omitempty"`
	UpdatedAt      time.Time               `json:"updated_at,omitempty"`
	ArchivedAt     *time.Time              `json:"archived_at,omitempty"`
	PurgedAt       *time.Time              `json:"purged_at,omitempty"`
	PublicMetadata map[string]string       `json:"public_metadata,omitempty"`
	Retention      *SupportRetentionExport `json:"retention,omitempty"`
	AuditTrail     []SupportAuditExport    `json:"audit_trail,omitempty"`
}

func (s *DataRightsService) processExport(ctx context.Context, req *DataExportRequest) {
	s.mu.Lock()
	req.Status = ExportProcessing
	_ = s.store.SaveExportRequest(ctx, req)
	s.mu.Unlock()

	data := s.buildExportPayload(ctx, req.DataSubject)
	payload, err := s.serializeExport(req.Format, data)
	if err != nil {
		s.failExport(ctx, req, err)
		return
	}

	if err := s.store.SaveExportData(ctx, req.ID, payload); err != nil {
		s.failExport(ctx, req, err)
		return
	}

	expiresAt := s.clock().Add(7 * 24 * time.Hour)
	req.Status = ExportReady
	req.DownloadURL = fmt.Sprintf("export://%s", req.ID)
	req.ExpiresAt = &expiresAt
	req.Error = ""

	if err := s.store.SaveExportRequest(ctx, req); err != nil {
		return
	}

	_ = s.audit.Log(ctx, AuditEvent{
		Action:      AuditExportReady,
		RequestID:   req.ID,
		DataSubject: req.DataSubject,
		Timestamp:   s.clock(),
	})

	_ = s.notifier.NotifyExportReady(ctx, req)
}

func (s *DataRightsService) failExport(ctx context.Context, req *DataExportRequest, err error) {
	req.Status = ExportFailed
	req.Error = err.Error()
	_ = s.store.SaveExportRequest(ctx, req)
	_ = s.audit.Log(ctx, AuditEvent{
		Action:      AuditExportFailed,
		RequestID:   req.ID,
		DataSubject: req.DataSubject,
		Timestamp:   s.clock(),
		Details: map[string]string{
			"error": err.Error(),
		},
	})
}

func (s *DataRightsService) buildExportPayload(ctx context.Context, dataSubject string) *UserDataExport {
	payload := &UserDataExport{
		ExportedAt:  s.clock(),
		DataSubject: dataSubject,
	}

	if s.identity != nil {
		identity, err := s.identity.GetIdentity(ctx, dataSubject)
		if err == nil {
			payload.Identity = identity
		}
	}

	if s.consents != nil {
		consents, err := s.consents.ListConsents(ctx, dataSubject)
		if err == nil {
			payload.Consents = consents
		}
	}

	if s.escrow != nil {
		escrows, err := s.escrow.GetUserEscrows(ctx, dataSubject)
		if err == nil {
			payload.EscrowRecords = escrows
		}
	}

	if s.support != nil {
		supportRecords, err := s.support.ListSupportRequests(ctx, dataSubject)
		if err == nil {
			payload.SupportRequests = supportRecords
		}
	}

	payload.Transactions = []TransactionExport{}
	if payload.SupportRequests == nil {
		payload.SupportRequests = []SupportExport{}
	}
	return payload
}

func (s *DataRightsService) serializeExport(format ExportFormat, payload *UserDataExport) ([]byte, error) {
	switch format {
	case FormatCSV:
		return toCSV(payload)
	default:
		return json.MarshalIndent(payload, "", "  ")
	}
}

func toCSV(payload *UserDataExport) ([]byte, error) {
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)

	if err := writer.Write([]string{"section", "field", "value"}); err != nil {
		return nil, err
	}

	write := func(section, field, value string) {
		_ = writer.Write([]string{section, field, value})
	}

	write("meta", "exported_at", payload.ExportedAt.Format(time.RFC3339))
	write("meta", "data_subject", payload.DataSubject)

	if payload.Identity != nil {
		write("identity", "address", payload.Identity.Address)
		write("identity", "trust_score", fmt.Sprintf("%0.2f", payload.Identity.TrustScore))
		write("identity", "tier_level", payload.Identity.TierLevel)
		write("identity", "created_at", payload.Identity.CreatedAt.Format(time.RFC3339))
	}

	for _, consent := range payload.Consents {
		write("consent", "id", consent.ID)
		write("consent", "scope_id", consent.ScopeID)
		write("consent", "purpose", consent.Purpose)
		write("consent", "status", consent.Status)
	}

	for _, escrow := range payload.EscrowRecords {
		write("escrow", "id", escrow.ID)
		write("escrow", "status", escrow.Status)
		write("escrow", "amount", escrow.Amount)
	}

	for _, support := range payload.SupportRequests {
		section := fmt.Sprintf("support:%s", support.TicketID)
		write(section, "ticket_id", support.TicketID)
		write(section, "ticket_number", support.TicketNumber)
		write(section, "submitter", support.Submitter)
		write(section, "category", support.Category)
		write(section, "priority", support.Priority)
		write(section, "status", support.Status)
		if !support.CreatedAt.IsZero() {
			write(section, "created_at", support.CreatedAt.Format(time.RFC3339))
		}
		if !support.UpdatedAt.IsZero() {
			write(section, "updated_at", support.UpdatedAt.Format(time.RFC3339))
		}
		if support.ArchivedAt != nil {
			write(section, "archived_at", support.ArchivedAt.Format(time.RFC3339))
		}
		if support.PurgedAt != nil {
			write(section, "purged_at", support.PurgedAt.Format(time.RFC3339))
		}
		for key, value := range support.PublicMetadata {
			write(section, fmt.Sprintf("public_metadata.%s", key), value)
		}
		if support.Retention != nil {
			write(section, "retention.archive_after_seconds", fmt.Sprintf("%d", support.Retention.ArchiveAfterSeconds))
			write(section, "retention.purge_after_seconds", fmt.Sprintf("%d", support.Retention.PurgeAfterSeconds))
			if !support.Retention.CreatedAt.IsZero() {
				write(section, "retention.created_at", support.Retention.CreatedAt.Format(time.RFC3339))
			}
		}
		for i, entry := range support.AuditTrail {
			auditSection := fmt.Sprintf("support_audit:%s:%d", support.TicketID, i)
			write(auditSection, "action", entry.Action)
			write(auditSection, "performed_by", entry.PerformedBy)
			write(auditSection, "details", entry.Details)
			if !entry.Timestamp.IsZero() {
				write(auditSection, "timestamp", entry.Timestamp.Format(time.RFC3339))
			}
			if entry.BlockHeight != 0 {
				write(auditSection, "block_height", fmt.Sprintf("%d", entry.BlockHeight))
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return []byte(strings.TrimSpace(buffer.String())), nil
}
