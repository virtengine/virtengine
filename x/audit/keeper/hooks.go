package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AuditHooks defines the interface for audit log hooks that other modules can use
type AuditHooks interface {
	// AfterScopeUploaded is called after a scope is uploaded in veid module
	AfterScopeUploaded(ctx sdk.Context, actor, scopeID string, scopeType string)

	// AfterScopeVerified is called after a scope is verified
	AfterScopeVerified(ctx sdk.Context, actor, scopeID string, status string, verifierAddr string)

	// AfterScoreUpdated is called after an identity score is updated
	AfterScoreUpdated(ctx sdk.Context, actor string, oldScore, newScore uint32, reason string)

	// AfterOrderCreated is called after an order is created in market module
	AfterOrderCreated(ctx sdk.Context, customerAddr, orderID string, resourceType string)

	// AfterBidMatched is called after a bid is matched
	AfterBidMatched(ctx sdk.Context, providerAddr, bidID, orderID string, price string)

	// AfterLeaseCreated is called after a lease is created
	AfterLeaseCreated(ctx sdk.Context, customerAddr, providerAddr, leaseID string)

	// AfterLeaseClosed is called after a lease is closed
	AfterLeaseClosed(ctx sdk.Context, leaseID string, reason string)

	// AfterPaymentSettled is called after a payment is settled in settlement module
	AfterPaymentSettled(ctx sdk.Context, payerAddr, payeeAddr, amount string, settlementID string)

	// AfterTicketCreated is called after a support ticket is created
	AfterTicketCreated(ctx sdk.Context, creatorAddr, ticketID, ticketType, priority string)

	// AfterTicketResolved is called after a support ticket is resolved
	AfterTicketResolved(ctx sdk.Context, ticketID, resolverAddr, resolution string)

	// AfterProviderRegistered is called after a provider is registered
	AfterProviderRegistered(ctx sdk.Context, providerAddr string, attributes map[string]string)

	// AfterProviderAttributesUpdated is called after provider attributes are updated
	AfterProviderAttributesUpdated(ctx sdk.Context, providerAddr string, updatedKeys []string)
}

// AuditLogHooks is the implementation of AuditHooks that writes to audit logs
type AuditLogHooks struct {
	k Keeper
}

// NewAuditLogHooks creates a new AuditLogHooks instance
func NewAuditLogHooks(k Keeper) AuditLogHooks {
	return AuditLogHooks{k: k}
}

// Ensure AuditLogHooks implements AuditHooks
var _ AuditHooks = AuditLogHooks{}

// AfterScopeUploaded implements AuditHooks
func (h AuditLogHooks) AfterScopeUploaded(ctx sdk.Context, actor, scopeID string, scopeType string) {
	metadata := map[string]interface{}{
		"scope_id":   scopeID,
		"scope_type": scopeType,
	}
	_ = h.k.AppendLog(ctx, actor, "veid", "scope_uploaded", scopeID, metadata)
}

// AfterScopeVerified implements AuditHooks
func (h AuditLogHooks) AfterScopeVerified(ctx sdk.Context, actor, scopeID string, status string, verifierAddr string) {
	metadata := map[string]interface{}{
		"scope_id":      scopeID,
		"status":        status,
		"verifier_addr": verifierAddr,
	}
	_ = h.k.AppendLog(ctx, actor, "veid", "scope_verified", scopeID, metadata)
}

// AfterScoreUpdated implements AuditHooks
func (h AuditLogHooks) AfterScoreUpdated(ctx sdk.Context, actor string, oldScore, newScore uint32, reason string) {
	metadata := map[string]interface{}{
		"old_score": oldScore,
		"new_score": newScore,
		"reason":    reason,
	}
	_ = h.k.AppendLog(ctx, actor, "veid", "score_updated", "", metadata)
}

// AfterOrderCreated implements AuditHooks
func (h AuditLogHooks) AfterOrderCreated(ctx sdk.Context, customerAddr, orderID string, resourceType string) {
	metadata := map[string]interface{}{
		"order_id":      orderID,
		"resource_type": resourceType,
	}
	_ = h.k.AppendLog(ctx, customerAddr, "market", "order_created", orderID, metadata)
}

// AfterBidMatched implements AuditHooks
func (h AuditLogHooks) AfterBidMatched(ctx sdk.Context, providerAddr, bidID, orderID string, price string) {
	metadata := map[string]interface{}{
		"bid_id":   bidID,
		"order_id": orderID,
		"price":    price,
	}
	_ = h.k.AppendLog(ctx, providerAddr, "market", "bid_matched", bidID, metadata)
}

// AfterLeaseCreated implements AuditHooks
func (h AuditLogHooks) AfterLeaseCreated(ctx sdk.Context, customerAddr, providerAddr, leaseID string) {
	metadata := map[string]interface{}{
		"lease_id":      leaseID,
		"customer_addr": customerAddr,
		"provider_addr": providerAddr,
	}
	_ = h.k.AppendLog(ctx, customerAddr, "market", "lease_created", leaseID, metadata)
}

// AfterLeaseClosed implements AuditHooks
func (h AuditLogHooks) AfterLeaseClosed(ctx sdk.Context, leaseID string, reason string) {
	metadata := map[string]interface{}{
		"lease_id": leaseID,
		"reason":   reason,
	}
	_ = h.k.AppendLog(ctx, "", "market", "lease_closed", leaseID, metadata)
}

// AfterPaymentSettled implements AuditHooks
func (h AuditLogHooks) AfterPaymentSettled(ctx sdk.Context, payerAddr, payeeAddr, amount string, settlementID string) {
	metadata := map[string]interface{}{
		"payer_addr":    payerAddr,
		"payee_addr":    payeeAddr,
		"amount":        amount,
		"settlement_id": settlementID,
	}
	_ = h.k.AppendLog(ctx, payerAddr, "settlement", "payment_settled", settlementID, metadata)
}

// AfterTicketCreated implements AuditHooks
func (h AuditLogHooks) AfterTicketCreated(ctx sdk.Context, creatorAddr, ticketID, ticketType, priority string) {
	metadata := map[string]interface{}{
		"ticket_id":   ticketID,
		"ticket_type": ticketType,
		"priority":    priority,
	}
	_ = h.k.AppendLog(ctx, creatorAddr, "support", "ticket_created", ticketID, metadata)
}

// AfterTicketResolved implements AuditHooks
func (h AuditLogHooks) AfterTicketResolved(ctx sdk.Context, ticketID, resolverAddr, resolution string) {
	metadata := map[string]interface{}{
		"ticket_id":     ticketID,
		"resolver_addr": resolverAddr,
		"resolution":    resolution,
	}
	_ = h.k.AppendLog(ctx, resolverAddr, "support", "ticket_resolved", ticketID, metadata)
}

// AfterProviderRegistered implements AuditHooks
func (h AuditLogHooks) AfterProviderRegistered(ctx sdk.Context, providerAddr string, attributes map[string]string) {
	metadata := map[string]interface{}{
		"provider_addr": providerAddr,
		"attributes":    attributes,
	}
	_ = h.k.AppendLog(ctx, providerAddr, "provider", "provider_registered", providerAddr, metadata)
}

// AfterProviderAttributesUpdated implements AuditHooks
func (h AuditLogHooks) AfterProviderAttributesUpdated(ctx sdk.Context, providerAddr string, updatedKeys []string) {
	metadata := map[string]interface{}{
		"provider_addr": providerAddr,
		"updated_keys":  updatedKeys,
	}
	_ = h.k.AppendLog(ctx, providerAddr, "provider", "provider_attributes_updated", providerAddr, metadata)
}
