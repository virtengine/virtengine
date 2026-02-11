package gdpr

import (
	"context"
	"fmt"
)

func (s *DataRightsService) checkDeletionBlockers(ctx context.Context, dataSubject string) ([]string, error) {
	blockers := make([]string, 0)
	if s.escrow != nil {
		hasActive, err := s.escrow.HasActiveEscrow(ctx, dataSubject)
		if err != nil {
			return nil, err
		}
		if hasActive {
			blockers = append(blockers, "active_escrow")
		}
	}
	return blockers, nil
}

func (s *DataRightsService) executeDeletion(ctx context.Context, dataSubject string) error {
	if s.identity != nil {
		if err := s.identity.DeleteIdentity(ctx, dataSubject); err != nil {
			return fmt.Errorf("delete identity: %w", err)
		}
	}

	if s.consents != nil {
		if err := s.consents.DeleteConsentRecords(ctx, dataSubject); err != nil {
			return fmt.Errorf("delete consent records: %w", err)
		}
	}

	if s.support != nil {
		if err := s.support.DeleteSupportRequests(ctx, dataSubject); err != nil {
			return fmt.Errorf("delete support records: %w", err)
		}
	}

	return nil
}
