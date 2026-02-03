package v1beta4

import "fmt"

// EventProviderDomainVerified defines an SDK message for provider domain verified event
type EventProviderDomainVerified struct {
	Owner  string `json:"owner" yaml:"owner"`
	Domain string `json:"domain" yaml:"domain"`
}

// ProtoMessage implements proto.Message
func (*EventProviderDomainVerified) ProtoMessage() {}

// Reset implements proto.Message
func (e *EventProviderDomainVerified) Reset() { *e = EventProviderDomainVerified{} }

// String implements proto.Message
func (e *EventProviderDomainVerified) String() string {
	return fmt.Sprintf("EventProviderDomainVerified{Owner: %s, Domain: %s}", e.Owner, e.Domain)
}

// EventProviderDomainVerificationStarted defines an SDK message for when domain verification begins
type EventProviderDomainVerificationStarted struct {
	Owner  string `json:"owner" yaml:"owner"`
	Domain string `json:"domain" yaml:"domain"`
	Token  string `json:"token" yaml:"token"`
}

// ProtoMessage implements proto.Message
func (*EventProviderDomainVerificationStarted) ProtoMessage() {}

// Reset implements proto.Message
func (e *EventProviderDomainVerificationStarted) Reset() {
	*e = EventProviderDomainVerificationStarted{}
}

// String implements proto.Message
func (e *EventProviderDomainVerificationStarted) String() string {
	return fmt.Sprintf("EventProviderDomainVerificationStarted{Owner: %s, Domain: %s}", e.Owner, e.Domain)
}
