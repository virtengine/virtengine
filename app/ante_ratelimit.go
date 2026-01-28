package app

import (
	"fmt"
	"sync"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	apptypes "github.com/virtengine/virtengine/app/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// RateLimitDecorator implements chain-level rate limiting for transactions.
// It tracks per-account transaction counts and global VEID transaction counts
// per block using transient storage that resets each block.
type RateLimitDecorator struct {
	mu     sync.RWMutex
	store  *apptypes.TransientRateLimitStore
	logger log.Logger

	// Metrics tracking
	metrics *rateLimitMetrics
}

// rateLimitMetrics tracks rate limiting statistics
type rateLimitMetrics struct {
	mu             sync.Mutex
	totalBlocked   uint64
	accountBlocked uint64
	veidBlocked    uint64
	blockBlocked   uint64
}

// NewRateLimitDecorator creates a new rate limit decorator with the given parameters
func NewRateLimitDecorator(params apptypes.RateLimitParams, logger log.Logger) RateLimitDecorator {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	return RateLimitDecorator{
		store:   apptypes.NewTransientRateLimitStore(params),
		logger:  logger.With("module", "ante-ratelimit"),
		metrics: &rateLimitMetrics{},
	}
}

// AnteHandle implements sdk.AnteDecorator
func (rld RateLimitDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	// Skip rate limiting during simulation (gas estimation)
	if simulate {
		return next(ctx, tx, simulate)
	}

	// Get current parameters
	rld.mu.Lock()
	params := rld.store.GetParams()

	// Skip if rate limiting is disabled
	if !params.Enabled {
		rld.mu.Unlock()
		return next(ctx, tx, simulate)
	}

	// Reset counters for new block
	blockHeight := ctx.BlockHeight()
	rld.store.ResetForBlock(blockHeight)
	rld.mu.Unlock()

	// Check total block transaction limit first
	if err := rld.checkBlockLimit(ctx, params); err != nil {
		return ctx, err
	}

	// Get signers from the transaction
	signers, err := rld.getSigners(tx)
	if err != nil {
		// If we can't get signers, log and continue (don't block)
		rld.logger.Debug("failed to get signers for rate limiting", "error", err)
		return next(ctx, tx, simulate)
	}

	// Check per-account rate limits for each signer
	for _, signer := range signers {
		// Skip exempt addresses
		if params.IsExempt(signer) {
			continue
		}

		if err := rld.checkAccountLimit(ctx, signer, params); err != nil {
			return ctx, err
		}
	}

	// Check VEID transaction limits
	if rld.isVEIDTransaction(tx) {
		if err := rld.checkVEIDLimit(ctx, params); err != nil {
			return ctx, err
		}
	}

	// Increment counters for successful check
	rld.mu.Lock()
	for _, signer := range signers {
		if !params.IsExempt(signer) {
			rld.store.IncrementAccountTxCount(signer)
		}
	}
	if rld.isVEIDTransaction(tx) {
		rld.store.IncrementVEIDTxCount()
	}
	rld.store.IncrementTotalTxCount()
	rld.mu.Unlock()

	return next(ctx, tx, simulate)
}

// getSigners extracts signer addresses from a transaction
func (rld RateLimitDecorator) getSigners(tx sdk.Tx) ([]sdk.AccAddress, error) {
	sigTx, ok := tx.(signing.SigVerifiableTx)
	if !ok {
		return nil, nil // Not a signable tx, skip
	}

	signersBytes, err := sigTx.GetSigners()
	if err != nil {
		return nil, err
	}

	// Convert [][]byte to []sdk.AccAddress
	signers := make([]sdk.AccAddress, len(signersBytes))
	for i, signerBytes := range signersBytes {
		signers[i] = sdk.AccAddress(signerBytes)
	}

	return signers, nil
}

// checkBlockLimit checks if the total block transaction limit has been reached
func (rld RateLimitDecorator) checkBlockLimit(ctx sdk.Context, params apptypes.RateLimitParams) error {
	rld.mu.RLock()
	currentCount := rld.store.GetTotalTxCount()
	rld.mu.RUnlock()

	if currentCount >= params.MaxTotalTxPerBlock {
		rld.recordBlockedMetric("block")
		rld.emitRateLimitEvent(ctx, "", "block_limit", currentCount, params.MaxTotalTxPerBlock)

		rld.logger.Warn("block transaction limit exceeded",
			"current", currentCount,
			"limit", params.MaxTotalTxPerBlock,
			"height", ctx.BlockHeight())

		return apptypes.ErrBlockRateLimited.Wrapf(
			"block has %d transactions, limit is %d",
			currentCount, params.MaxTotalTxPerBlock)
	}

	return nil
}

// checkAccountLimit checks if an account has exceeded its per-block transaction limit
func (rld RateLimitDecorator) checkAccountLimit(ctx sdk.Context, addr sdk.AccAddress, params apptypes.RateLimitParams) error {
	rld.mu.RLock()
	currentCount := rld.store.GetAccountTxCount(addr)
	rld.mu.RUnlock()

	if currentCount >= params.MaxTxPerBlockPerAccount {
		rld.recordBlockedMetric("account")
		rld.emitRateLimitEvent(ctx, addr.String(), "account_limit", currentCount, params.MaxTxPerBlockPerAccount)

		rld.logger.Warn("account transaction limit exceeded",
			"account", addr.String(),
			"current", currentCount,
			"limit", params.MaxTxPerBlockPerAccount,
			"height", ctx.BlockHeight())

		return apptypes.ErrAccountRateLimited.Wrapf(
			"account %s has %d transactions in this block, limit is %d",
			addr.String(), currentCount, params.MaxTxPerBlockPerAccount)
	}

	return nil
}

// checkVEIDLimit checks if the global VEID transaction limit has been reached
func (rld RateLimitDecorator) checkVEIDLimit(ctx sdk.Context, params apptypes.RateLimitParams) error {
	rld.mu.RLock()
	currentCount := rld.store.GetVEIDTxCount()
	rld.mu.RUnlock()

	if currentCount >= params.MaxVEIDTxPerBlockGlobal {
		rld.recordBlockedMetric("veid")
		rld.emitRateLimitEvent(ctx, "", "veid_limit", currentCount, params.MaxVEIDTxPerBlockGlobal)

		rld.logger.Warn("VEID transaction limit exceeded",
			"current", currentCount,
			"limit", params.MaxVEIDTxPerBlockGlobal,
			"height", ctx.BlockHeight())

		return apptypes.ErrVEIDRateLimited.Wrapf(
			"block has %d VEID transactions, limit is %d",
			currentCount, params.MaxVEIDTxPerBlockGlobal)
	}

	return nil
}

// isVEIDTransaction checks if a transaction contains VEID-related messages
func (rld RateLimitDecorator) isVEIDTransaction(tx sdk.Tx) bool {
	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		switch msg.(type) {
		case *veidtypes.MsgUploadScope,
			*veidtypes.MsgRevokeScope,
			*veidtypes.MsgRequestVerification,
			*veidtypes.MsgUpdateVerificationStatus,
			*veidtypes.MsgUpdateScore:
			return true
		}

		// Also check by type URL for safety
		typeURL := sdk.MsgTypeURL(msg)
		if isVEIDTypeURL(typeURL) {
			return true
		}
	}
	return false
}

// isVEIDTypeURL checks if a message type URL is a VEID message
func isVEIDTypeURL(typeURL string) bool {
	veidTypeURLs := []string{
		"/virtengine.veid.v1.MsgUploadScope",
		"/virtengine.veid.v1.MsgRevokeScope",
		"/virtengine.veid.v1.MsgRequestVerification",
		"/virtengine.veid.v1.MsgUpdateVerificationStatus",
		"/virtengine.veid.v1.MsgUpdateScore",
		"/virtengine.veid.v1.MsgRebindWallet",
	}
	for _, url := range veidTypeURLs {
		if typeURL == url {
			return true
		}
	}
	return false
}

// recordBlockedMetric records a blocked transaction metric
func (rld RateLimitDecorator) recordBlockedMetric(reason string) {
	rld.metrics.mu.Lock()
	defer rld.metrics.mu.Unlock()

	rld.metrics.totalBlocked++

	switch reason {
	case "account":
		rld.metrics.accountBlocked++
	case "veid":
		rld.metrics.veidBlocked++
	case "block":
		rld.metrics.blockBlocked++
	}

	// Emit telemetry metrics
	telemetry.IncrCounter(1, "ante", "ratelimit", "blocked", reason)
	telemetry.IncrCounter(1, "ante", "ratelimit", "blocked", "total")
}

// emitRateLimitEvent emits a rate limit event to the context
func (rld RateLimitDecorator) emitRateLimitEvent(ctx sdk.Context, account, reason string, current, limit uint64) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"rate_limit_exceeded",
			sdk.NewAttribute("account", account),
			sdk.NewAttribute("reason", reason),
			sdk.NewAttribute("current_count", fmt.Sprintf("%d", current)),
			sdk.NewAttribute("limit", fmt.Sprintf("%d", limit)),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

// GetMetrics returns the current rate limiting metrics
func (rld RateLimitDecorator) GetMetrics() apptypes.RateLimitMetrics {
	rld.metrics.mu.Lock()
	defer rld.metrics.mu.Unlock()

	return apptypes.RateLimitMetrics{
		TotalBlocked:   rld.metrics.totalBlocked,
		AccountBlocked: rld.metrics.accountBlocked,
		VEIDBlocked:    rld.metrics.veidBlocked,
		TotalTxBlocked: rld.metrics.blockBlocked,
	}
}

// UpdateParams updates the rate limit parameters
func (rld *RateLimitDecorator) UpdateParams(params apptypes.RateLimitParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	rld.mu.Lock()
	defer rld.mu.Unlock()
	rld.store.SetParams(params)

	rld.logger.Info("rate limit parameters updated",
		"enabled", params.Enabled,
		"max_tx_per_block_per_account", params.MaxTxPerBlockPerAccount,
		"max_veid_tx_per_block_global", params.MaxVEIDTxPerBlockGlobal,
		"max_total_tx_per_block", params.MaxTotalTxPerBlock,
		"exempt_addresses", len(params.ExemptAddresses))

	return nil
}

// DisableRateLimiting disables rate limiting (for testing or emergency)
func (rld *RateLimitDecorator) DisableRateLimiting() {
	rld.mu.Lock()
	defer rld.mu.Unlock()

	params := rld.store.GetParams()
	params.Enabled = false
	rld.store.SetParams(params)

	rld.logger.Warn("rate limiting disabled")
}

// EnableRateLimiting enables rate limiting
func (rld *RateLimitDecorator) EnableRateLimiting() {
	rld.mu.Lock()
	defer rld.mu.Unlock()

	params := rld.store.GetParams()
	params.Enabled = true
	rld.store.SetParams(params)

	rld.logger.Info("rate limiting enabled")
}
