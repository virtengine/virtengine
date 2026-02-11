package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/pkg/gdpr"
)

// GDPRSupportProvider implements gdpr.SupportProvider using the support keeper.
type GDPRSupportProvider struct {
	keeper Keeper
}

// NewGDPRSupportProvider creates a GDPR support provider wrapper.
func NewGDPRSupportProvider(k Keeper) GDPRSupportProvider {
	return GDPRSupportProvider{keeper: k}
}

// ListSupportRequests returns support tickets for a data subject.
func (p GDPRSupportProvider) ListSupportRequests(ctx context.Context, dataSubject string) ([]gdpr.SupportExport, error) {
	sdkCtx, err := unwrapSDKContext(ctx)
	if err != nil {
		return nil, err
	}
	addr, err := sdk.AccAddressFromBech32(dataSubject)
	if err != nil {
		return nil, err
	}

	requests := p.keeper.GetSupportRequestsBySubmitter(sdkCtx, addr)
	exports := make([]gdpr.SupportExport, 0, len(requests))
	for _, req := range requests {
		export := gdpr.SupportExport{
			TicketID:       req.ID.String(),
			TicketNumber:   req.TicketNumber,
			Submitter:      req.SubmitterAddress,
			Category:       string(req.Category),
			Priority:       string(req.Priority),
			Status:         req.Status.String(),
			CreatedAt:      req.CreatedAt,
			UpdatedAt:      req.UpdatedAt,
			ArchivedAt:     req.ArchivedAt,
			PurgedAt:       req.PurgedAt,
			PublicMetadata: req.PublicMetadata,
			AuditTrail:     make([]gdpr.SupportAuditExport, 0, len(req.AuditTrail)),
		}
		if req.RetentionPolicy != nil {
			export.Retention = &gdpr.SupportRetentionExport{
				ArchiveAfterSeconds: req.RetentionPolicy.ArchiveAfterSeconds,
				PurgeAfterSeconds:   req.RetentionPolicy.PurgeAfterSeconds,
				CreatedAt:           req.RetentionPolicy.CreatedAt,
				CreatedAtBlock:      req.RetentionPolicy.CreatedAtBlock,
			}
		}
		for _, entry := range req.AuditTrail {
			export.AuditTrail = append(export.AuditTrail, gdpr.SupportAuditExport{
				Action:      string(entry.Action),
				PerformedBy: entry.PerformedBy,
				Details:     entry.Details,
				Timestamp:   entry.Timestamp,
				BlockHeight: entry.BlockHeight,
			})
		}
		exports = append(exports, export)
	}

	return exports, nil
}

// DeleteSupportRequests purges payloads for the data subject's support requests.
func (p GDPRSupportProvider) DeleteSupportRequests(ctx context.Context, dataSubject string) error {
	sdkCtx, err := unwrapSDKContext(ctx)
	if err != nil {
		return err
	}
	addr, err := sdk.AccAddressFromBech32(dataSubject)
	if err != nil {
		return err
	}

	requests := p.keeper.GetSupportRequestsBySubmitter(sdkCtx, addr)
	for _, req := range requests {
		if req.Purged {
			continue
		}
		if err := p.keeper.PurgeSupportRequestPayload(sdkCtx, req.ID, "gdpr deletion", "system"); err != nil {
			return fmt.Errorf("purge support request %s: %w", req.ID.String(), err)
		}
	}
	return nil
}

func unwrapSDKContext(ctx context.Context) (sdkCtx sdk.Context, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unsupported context type: %v", r)
		}
	}()
	sdkCtx = sdk.UnwrapSDKContext(ctx)
	return sdkCtx, err
}
