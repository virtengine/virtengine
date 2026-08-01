package data_vault

import "context"

// ConsentRequest captures consent inputs for data vault reads.
type ConsentRequest struct {
	Requester string
	Owner     string
	Scope     Scope
	OrgID     string
	Purpose   string
	Reason    string
	Metadata  map[string]string
}

// ConsentResolver validates consent for data access.
type ConsentResolver interface {
	HasConsent(ctx context.Context, req ConsentRequest) (bool, error)
}

// DenyAllConsentResolver denies access when no consent integration is configured.
type DenyAllConsentResolver struct{}

// HasConsent returns false for all requests.
func (DenyAllConsentResolver) HasConsent(_ context.Context, _ ConsentRequest) (bool, error) {
	return false, nil
}

// AllowAllConsentResolver permits all access in explicitly configured development fixtures.
// Deprecated: use only for development fixtures; production must use a real resolver.
type AllowAllConsentResolver struct{}

// HasConsent returns true for all requests.
func (AllowAllConsentResolver) HasConsent(_ context.Context, _ ConsentRequest) (bool, error) {
	return true, nil
}
