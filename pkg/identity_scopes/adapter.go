package identity_scopes

import (
	"context"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// WebScopeChainClient defines the chain submission interface for web-scope proofs.
type WebScopeChainClient interface {
	SubmitSSOVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitSSOVerificationProof) error
	SubmitEmailVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitEmailVerificationProof) error
	SubmitSMSVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitSMSVerificationProof) error
}

// ProofResult is a generic submission result for web-scope proofs.
type ProofResult struct {
	VerificationID    string
	LinkageID         string
	Status            string
	ScoreContribution uint32
}
