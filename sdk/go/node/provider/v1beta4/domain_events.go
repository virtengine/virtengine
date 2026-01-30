package v1beta4

// EventProviderDomainVerified defines an SDK message for provider domain verified event
type EventProviderDomainVerified struct {
	Owner  string `json:"owner" yaml:"owner"`
	Domain string `json:"domain" yaml:"domain"`
}

// EventProviderDomainVerificationStarted defines an SDK message for when domain verification begins
type EventProviderDomainVerificationStarted struct {
	Owner  string `json:"owner" yaml:"owner"`
	Domain string `json:"domain" yaml:"domain"`
	Token  string `json:"token" yaml:"token"`
}
