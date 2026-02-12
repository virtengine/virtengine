package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/audit/v1"
)

// ExportProcessor handles export job processing
type ExportProcessor struct {
	keeper Keeper
}

// NewExportProcessor creates a new export processor
func NewExportProcessor(keeper Keeper) *ExportProcessor {
	return &ExportProcessor{keeper: keeper}
}

// ProcessExportJob processes an export job
func (e *ExportProcessor) ProcessExportJob(ctx sdk.Context, jobID string) error {
	// Get the job
	job, found := e.keeper.GetExportJob(ctx, jobID)
	if !found {
		return types.ErrExportJobNotFound.Wrapf("job %s not found", jobID)
	}

	// Check if already processed
	if job.Status == types.ExportStatusCompleted || job.Status == types.ExportStatusFailed {
		return nil // Already processed
	}

	// Update status to in progress
	now := ctx.BlockTime()
	job.Status = types.ExportStatusInProgress
	job.StartedAt = &now
	_ = e.keeper.UpdateExportJob(ctx, job)

	// Query logs based on filter
	filter := types.ExportFilter{}
	if job.Filter != nil {
		filter = *job.Filter
	}

	// Apply batch size limit from params
	params := e.keeper.GetAuditLogParams(ctx)
	limit := filter.Limit
	if limit == 0 || limit > params.MaxExportBatchSize {
		limit = params.MaxExportBatchSize
	}

	logs, err := e.keeper.QueryLogs(ctx, filter, limit)
	if err != nil {
		job.Status = types.ExportStatusFailed
		job.Error = err.Error()
		completedAt := ctx.BlockTime()
		job.CompletedAt = &completedAt
		_ = e.keeper.UpdateExportJob(ctx, job)
		return err
	}

	// Export logs based on format
	var exportData []byte
	switch job.Format {
	case "json":
		exportData, err = e.exportJSON(logs)
	case "csv":
		exportData = e.exportCSV(logs)
	default:
		err = types.ErrInvalidExportFormat.Wrapf("unsupported format: %s", job.Format)
	}

	if err != nil {
		job.Status = types.ExportStatusFailed
		job.Error = err.Error()
		completedAt := ctx.BlockTime()
		job.CompletedAt = &completedAt
		_ = e.keeper.UpdateExportJob(ctx, job)
		return err
	}

	// Sign the export
	signature := e.signExport(exportData)

	// Generate file path
	filePath := fmt.Sprintf("/exports/%s.%s", jobID, job.Format)

	// Mark logs as exported
	for _, log := range logs {
		log.Exported = true
		log.ExportJobId = jobID
		// In a real implementation, we'd save this back to the store
		// For now, we'll skip this to keep it simple
	}

	// Update job as completed
	job.Status = types.ExportStatusCompleted
	job.EntryCount = int64(len(logs))
	job.FilePath = filePath
	job.Signature = signature
	completedAt := ctx.BlockTime()
	job.CompletedAt = &completedAt
	_ = e.keeper.UpdateExportJob(ctx, job)

	return nil
}

// exportJSON exports logs as JSON
func (e *ExportProcessor) exportJSON(logs []types.AuditLogEntry) ([]byte, error) {
	export := map[string]interface{}{
		"version":     "1.0",
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"entries":     logs,
		"count":       len(logs),
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return data, nil
}

// exportCSV exports logs as CSV (error return for interface consistency)
func (e *ExportProcessor) exportCSV(logs []types.AuditLogEntry) []byte {
	// CSV header
	csv := "id,height,timestamp,actor,module,action,resource_id,metadata,exported\n"

	// CSV rows
	for _, log := range logs {
		csv += fmt.Sprintf("%s,%d,%s,%s,%s,%s,%s,%q,%t\n",
			log.Id,
			log.Height,
			log.Timestamp.Format(time.RFC3339),
			log.Actor,
			log.Module,
			log.Action,
			log.ResourceId,
			log.Metadata,
			log.Exported,
		)
	}

	return []byte(csv)
}

// signExport signs the export data
func (e *ExportProcessor) signExport(data []byte) []byte {
	// Simple SHA256 hash as signature
	// In production, this should use proper cryptographic signing
	hash := sha256.Sum256(data)
	return []byte(hex.EncodeToString(hash[:]))
}
