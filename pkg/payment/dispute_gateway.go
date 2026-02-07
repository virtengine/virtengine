// Package payment provides payment gateway integration for Visa/Mastercard.
//
// PAY-003: Dispute lifecycle persistence and gateway actions
package payment

import "context"

// disputeGateway is the internal interface for dispute-related gateway actions.
// It is intentionally narrower than Gateway to avoid exposing dispute methods
// on the public Gateway interface.
type disputeGateway interface {
	GetDispute(ctx context.Context, disputeID string) (Dispute, error)
	ListDisputes(ctx context.Context, paymentIntentID string) ([]Dispute, error)
	SubmitDisputeEvidence(ctx context.Context, disputeID string, evidence DisputeEvidence) error
	AcceptDispute(ctx context.Context, disputeID string) error
}

func resolveDisputeGateway(gateway Gateway) (disputeGateway, error) {
	if gateway == nil {
		return nil, ErrGatewayNotConfigured
	}
	if dg, ok := gateway.(disputeGateway); ok {
		return dg, nil
	}
	return nil, ErrGatewayNotConfigured
}
