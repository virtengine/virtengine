package types

import (
sdk "github.com/cosmos/cosmos-sdk/types"

configv1 "github.com/virtengine/virtengine/sdk/go/node/config/v1"
)

// Type aliases to generated protobuf types
type (
MsgRegisterApprovedClient           = configv1.MsgRegisterApprovedClient
MsgRegisterApprovedClientResponse   = configv1.MsgRegisterApprovedClientResponse
MsgUpdateApprovedClient             = configv1.MsgUpdateApprovedClient
MsgUpdateApprovedClientResponse     = configv1.MsgUpdateApprovedClientResponse
MsgSuspendApprovedClient            = configv1.MsgSuspendApprovedClient
MsgSuspendApprovedClientResponse    = configv1.MsgSuspendApprovedClientResponse
MsgRevokeApprovedClient             = configv1.MsgRevokeApprovedClient
MsgRevokeApprovedClientResponse     = configv1.MsgRevokeApprovedClientResponse
MsgReactivateApprovedClient         = configv1.MsgReactivateApprovedClient
MsgReactivateApprovedClientResponse = configv1.MsgReactivateApprovedClientResponse
MsgUpdateParams                     = configv1.MsgUpdateParams
MsgUpdateParamsResponse             = configv1.MsgUpdateParamsResponse
)

// Message type constants
const (
TypeMsgRegisterApprovedClient   = "register_approved_client"
TypeMsgUpdateApprovedClient     = "update_approved_client"
TypeMsgSuspendApprovedClient    = "suspend_approved_client"
TypeMsgRevokeApprovedClient     = "revoke_approved_client"
TypeMsgReactivateApprovedClient = "reactivate_approved_client"
TypeMsgUpdateParams             = "update_params"
)

var (
_ sdk.Msg = &MsgRegisterApprovedClient{}
_ sdk.Msg = &MsgUpdateApprovedClient{}
_ sdk.Msg = &MsgSuspendApprovedClient{}
_ sdk.Msg = &MsgRevokeApprovedClient{}
_ sdk.Msg = &MsgReactivateApprovedClient{}
_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgRegisterApprovedClient creates a new MsgRegisterApprovedClient
func NewMsgRegisterApprovedClient(
authority string,
clientID string,
publicKey string,
name string,
description string,
versionConstraint string,
allowedScopes []string,
) *MsgRegisterApprovedClient {
return &MsgRegisterApprovedClient{
Authority:         authority,
ClientId:          clientID,
PublicKey:         publicKey,
Name:              name,
Description:       description,
VersionConstraint: versionConstraint,
AllowedScopes:     allowedScopes,
}
}

// NewMsgUpdateApprovedClient creates a new MsgUpdateApprovedClient
func NewMsgUpdateApprovedClient(
authority string,
clientID string,
publicKey string,
versionConstraint string,
allowedScopes []string,
) *MsgUpdateApprovedClient {
return &MsgUpdateApprovedClient{
Authority:         authority,
ClientId:          clientID,
PublicKey:         publicKey,
VersionConstraint: versionConstraint,
AllowedScopes:     allowedScopes,
}
}

// NewMsgSuspendApprovedClient creates a new MsgSuspendApprovedClient
func NewMsgSuspendApprovedClient(authority string, clientID string, reason string) *MsgSuspendApprovedClient {
return &MsgSuspendApprovedClient{
Authority: authority,
ClientId:  clientID,
Reason:    reason,
}
}

// NewMsgRevokeApprovedClient creates a new MsgRevokeApprovedClient
func NewMsgRevokeApprovedClient(authority string, clientID string, reason string) *MsgRevokeApprovedClient {
return &MsgRevokeApprovedClient{
Authority: authority,
ClientId:  clientID,
Reason:    reason,
}
}

// NewMsgReactivateApprovedClient creates a new MsgReactivateApprovedClient
func NewMsgReactivateApprovedClient(authority string, clientID string) *MsgReactivateApprovedClient {
return &MsgReactivateApprovedClient{
Authority: authority,
ClientId:  clientID,
}
}