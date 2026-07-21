package app

import (
	"bytes"
	"errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
)

var errInvalidSystemTransaction = errors.New("invalid VEID consensus system transaction")

type authorizedSystemTxKey struct{}

func authorizeSystemTransaction(ctx sdk.Context, txBytes []byte) sdk.Context {
	return ctx.WithValue(authorizedSystemTxKey{}, bytes.Clone(txBytes))
}

func isAuthorizedSystemTransaction(ctx sdk.Context) bool {
	authorized, ok := ctx.Value(authorizedSystemTxKey{}).([]byte)
	return ok && len(authorized) > 0 && bytes.Equal(authorized, ctx.TxBytes())
}

func configureConsensusSystemTxAuthorization(keeper *veidkeeper.Keeper, validatorStore baseapp.ValidatorStore) {
	if keeper == nil {
		panic("VEID keeper is required")
	}
	if validatorStore == nil {
		panic("validator store is required")
	}
	keeper.SetConsensusSystemTxAuthorizer(isAuthorizedSystemTransaction)
	keeper.SetConsensusValidatorStore(validatorStore)
}

// SystemTxDecorator admits the no-ordinary-signer system transaction only in
// proposal verification and FinalizeBlock. CheckTx/ReCheckTx/Simulate reject it.
type SystemTxDecorator struct{}

// NewSystemTxDecorator creates the system transaction ante boundary.
func NewSystemTxDecorator() SystemTxDecorator { return SystemTxDecorator{} }

// AnteHandle enforces exclusive system transaction composition and execution mode.
func (SystemTxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	isSystem, invalid := classifySystemTransaction(tx)
	if invalid {
		return ctx, sdkerrors.ErrUnauthorized.Wrap(errInvalidSystemTransaction.Error())
	}
	if !isSystem {
		return next(ctx, tx, simulate)
	}
	if simulate || (ctx.ExecMode() != sdk.ExecModePrepareProposal &&
		ctx.ExecMode() != sdk.ExecModeProcessProposal &&
		ctx.ExecMode() != sdk.ExecModeFinalize) {
		return ctx, sdkerrors.ErrUnauthorized.Wrap("VEID consensus system transaction is proposal-only")
	}
	if ctx.ExecMode() == sdk.ExecModeFinalize && !isAuthorizedSystemTransaction(ctx) {
		return ctx, sdkerrors.ErrUnauthorized.Wrap("VEID consensus system transaction was not authorized by FinalizeBlock pre-validation")
	}
	return next(ctx, tx, simulate)
}

func classifySystemTransaction(tx sdk.Tx) (isSystem bool, invalid bool) {
	if tx == nil {
		return false, false
	}
	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		if _, ok := msg.(*veidv1.MsgSubmitConsensusVerification); ok {
			if len(msgs) != 1 {
				return false, true
			}
			return true, false
		}
	}
	return false, false
}
