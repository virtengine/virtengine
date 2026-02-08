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

// AllowAllConsentResolver permits all access when consent enforcement is disabled.
type AllowAllConsentResolver struct{}

// HasConsent returns true for all requests.
func (AllowAllConsentResolver) HasConsent(_ context.Context, _ ConsentRequest) (bool, error) {
	return true, nil
}
