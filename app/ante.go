package app

import (
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	apptypes "github.com/virtengine/virtengine/app/types"
	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
)

// HandlerOptions extends the SDK's AnteHandler options
type HandlerOptions struct {
	ante.HandlerOptions
	CDC             codec.BinaryCodec
	GovKeeper       *govkeeper.Keeper
	MFAGatingKeeper *mfakeeper.Keeper
	VEIDKeeper      *veidkeeper.Keeper
	RateLimitParams apptypes.RateLimitParams
	Logger          log.Logger
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.ErrLogic.Wrap("sign mode handler is required for ante builder")
	}

	if options.SigGasConsumer == nil {
		return nil, sdkerrors.ErrLogic.Wrap("sig gas consumer handler is required for ante builder")
	}

	if options.GovKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("virtengine governance keeper is required for ante builder")
	}

	if options.MFAGatingKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("virtengine MFA keeper is required for ante builder")
	}

	if options.VEIDKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("virtengine VEID keeper is required for ante builder")
	}

	if options.FeegrantKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("virtengine feegrant keeper is required for ante builder")
	}

	// Get rate limit params, using defaults if not provided
	rateLimitParams := options.RateLimitParams
	if rateLimitParams.MaxTxPerBlockPerAccount == 0 {
		rateLimitParams = apptypes.DefaultRateLimitParams()
	}

	// Get logger for rate limiting
	rateLimitLogger := options.Logger
	if rateLimitLogger == nil {
		rateLimitLogger = log.NewNopLogger()
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(),                         // outermost AnteDecorator. SetUpContext must be called first
		NewRateLimitDecorator(rateLimitParams, rateLimitLogger), // Rate limiting early to block spam before expensive ops
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, nil),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		NewVEIDDecorator(*options.VEIDKeeper, *options.MFAGatingKeeper, options.GovKeeper),
		NewMFAGatingDecorator(*options.MFAGatingKeeper),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
