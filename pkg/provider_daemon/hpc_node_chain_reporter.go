package provider_daemon

import (
	"context"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
)

// HPCNodeChainReporter submits node metadata updates to the chain.
type HPCNodeChainReporter interface {
	SubmitNodeMetadata(ctx context.Context, msg *hpcv1.MsgUpdateNodeMetadata) error
}
