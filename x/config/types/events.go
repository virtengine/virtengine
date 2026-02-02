package types

import (
	configv1 "github.com/virtengine/virtengine/sdk/go/node/config/v1"
)

// Event types for the config module
const (
	EventTypeClientRegistered  = "client_registered"
	EventTypeClientUpdated     = "client_updated"
	EventTypeClientSuspended   = "client_suspended"
	EventTypeClientRevoked     = "client_revoked"
	EventTypeClientReactivated = "client_reactivated"
	EventTypeSignatureVerified = "signature_verified"

	AttributeKeyClientID     = "client_id"
	AttributeKeyClientName   = "client_name"
	AttributeKeyStatus       = "status"
	AttributeKeyUpdatedBy    = "updated_by"
	AttributeKeyReason       = "reason"
	AttributeKeyMinVersion   = "min_version"
	AttributeKeyMaxVersion   = "max_version"
	AttributeKeyAllowedScope = "allowed_scope"
	AttributeKeyKeyType      = "key_type"
	AttributeKeyTimestamp    = "timestamp"
)

// Type aliases to generated protobuf event types
type (
	EventClientRegistered  = configv1.EventClientRegistered
	EventClientUpdated     = configv1.EventClientUpdated
	EventClientSuspended   = configv1.EventClientSuspended
	EventClientRevoked     = configv1.EventClientRevoked
	EventClientReactivated = configv1.EventClientReactivated
	EventSignatureVerified = configv1.EventSignatureVerified
)
