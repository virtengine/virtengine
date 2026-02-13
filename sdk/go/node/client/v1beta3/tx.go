package v1beta3

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/boz/go-lifecycle"
	"github.com/edwingeng/deque/v2"

	"github.com/cometbft/cometbft/mempool"
	cbcoretypes "github.com/cometbft/cometbft/rpc/core/types"
	cbtypes "github.com/cometbft/cometbft/types"

	cerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ttx "github.com/cosmos/cosmos-sdk/types/tx"

	nutils "github.com/virtengine/virtengine/sdk/go/node/utils"
	"github.com/virtengine/virtengine/sdk/go/util/ctxlog"

	cltypes "github.com/virtengine/virtengine/sdk/go/node/client/types"
)

var (
	ErrNotRunning       = errors.New("tx client: not running")
	ErrSyncTimedOut     = errors.New("tx client: timed-out waiting for sequence sync")
	ErrNodeCatchingUp   = errors.New("tx client: cannot sync from catching up node")
	ErrSimulateOffline  = errors.New("tx client: cannot simulate tx in offline mode")
	ErrBroadcastOffline = errors.New("tx client: cannot broadcast tx in offline mode")
	ErrTxCanceledByUser = errors.New("tx client: transaction declined by user input")
)

const (
	// BroadcastSync defines a tx broadcasting mode where the client waits for
	// a CheckTx execution response only.
	BroadcastSync = "sync"
	// BroadcastAsync defines a tx broadcasting mode where the client returns
	// immediately.
	BroadcastAsync = "async"

	BroadcastBlock = "block"

	BroadcastDefaultTimeout    = 30 * time.Second
	BroadcastBlockRetryTimeout = 300 * time.Second
	broadcastBlockRetryPeriod  = time.Second
	sequenceSyncTimeout        = 30 * time.Second

	// sadface.

	// Only way to detect the timeout error.
	// https://github.com/tendermint/tendermint/blob/46e06c97320bc61c4d98d3018f59d47ec69863c9/rpc/core/mempool.go#L124
	timeoutErrorMessage = "timed out waiting for tx to be included in a block"

	// Only way to check for tx not found error.
	// https://github.com/tendermint/tendermint/blob/46e06c97320bc61c4d98d3018f59d47ec69863c9/rpc/core/tx.go#L31-L33
	notFoundErrorMessageSuffix = ") not found"
)

var (
	sequenceMismatchRegexp = regexp.MustCompile(`(account sequence mismatch, expected \d+, got \d+)`)
)

var _ TxClient = (*serialBroadcaster)(nil)

type ConfirmFn func(string) (bool, error)

// BroadcastOptions defines the options allowed to configure a transaction broadcast.
type BroadcastOptions struct {
	generateOnly     *bool
	timeoutHeight    *uint64
	gasAdjustment    *float64
	gas              *cltypes.GasSetting
	gasPrices        *string
	fees             *string
	note             *string
	broadcastTimeout time.Duration
	resultAsError    bool
	skipConfirm      *bool
	confirmFn        ConfirmFn
	broadcastMode    *string
}

// BroadcastOption is a function that takes as first argument a pointer to BroadcastOptions and returns an error
// if the option cannot be configured. A number of BroadcastOption functions are available in this package.
type BroadcastOption func(*BroadcastOptions) error

// WithGasAdjustment returns a BroadcastOption that sets the gas adjustment configuration for the transaction.
func WithGasAdjustment(val float64) BroadcastOption {
	return func(options *BroadcastOptions) error {
		options.gasAdjustment = new(float64)
		*options.gasAdjustment = val
		return nil
	}
}

// WithNote returns a BroadcastOption that sets the note configuration for the transaction.
func WithNote(val string) BroadcastOption {
	return func(options *BroadcastOptions) error {
		options.note = new(string)
		*options.note = val
		return nil
	}
}

// WithGas returns a BroadcastOption that sets the gas setting configuration for the transaction.
func WithGas(val cltypes.GasSetting) BroadcastOption {
	return func(options *BroadcastOptions) error {
		options.gas = new(cltypes.GasSetting)
		*options.gas = val
		return nil
	}
}

// WithGasPrices returns a BroadcastOption that sets the gas price configuration for the transaction.
// Gas price is a string of the amount. E.g. "0.25uve".
func WithGasPrices(val string) BroadcastOption {
	return func(options *BroadcastOptions) error {
		options.gasPrices = new(string)
		*options.gasPrices = val
		return nil
	}
}

// WithFees returns a BroadcastOption that sets the fees configuration for the transaction.
func WithFees(val string) BroadcastOption {
	return func(options *BroadcastOptions) error {
		options.fees = new(string)
		*options.fees = val
		return nil
	}
}

// WithTimeoutHeight returns a BroadcastOption that sets the timeout height configuration for the transaction.
func WithTimeoutHeight(val uint64) BroadcastOption {
	return func(options *BroadcastOptions) error {
		options.timeoutHeight = new(uint64)
		*options.timeoutHeight = val
		return nil
	}
}

// WithBroadcastTimeout returns a BroadcastOption that sets the timeout configuration for the transaction.
func WithBroadcastTimeout(val time.Duration) BroadcastOption {
	return func(options *BroadcastOptions) error {
		options.broadcastTimeout = val
		return nil
	}
}

// WithResultCodeAsError returns a BroadcastOption that enables the result code as error configuration for the transaction.
func WithResultCodeAsError() BroadcastOption {
	return func(opts *BroadcastOptions) error {
		opts.resultAsError = true
		return nil
	}
}

// WithSkipConfirm returns a BroadcastOption that sets whether to skip or not the confirmation for the transaction.
func WithSkipConfirm(val bool) BroadcastOption {
	return func(opts *BroadcastOptions) error {
		opts.skipConfirm = new(bool)
		*opts.skipConfirm = val
		return nil
	}
}

// WithConfirmFn returns a BroadcastOption that sets the ConfirmFn function configuration for the transaction.
func WithConfirmFn(val ConfirmFn) BroadcastOption {
	return func(opts *BroadcastOptions) error {
		opts.confirmFn = val
		return nil
	}
}

// WithBroadcastMode returns a BroadcastOption that sets the broadcast for particular tx
func WithBroadcastMode(val string) BroadcastOption {
	return func(opts *BroadcastOptions) error {
		opts.broadcastMode = new(string)
		*opts.broadcastMode = val
		return nil
	}
}

// WithGenerateOnly returns a BroadcastOption that sets transaction generator to generate only mode
func WithGenerateOnly(val bool) BroadcastOption {
	return func(opts *BroadcastOptions) error {
		opts.generateOnly = new(bool)
		*opts.generateOnly = val

		return nil
	}
}

type broadcastResp struct {
	resp interface{}
	err  error
}

type broadcastReq struct {
	ctx        context.Context
	responsech chan<- broadcastResp
	data       interface{}
	opts       *BroadcastOptions
}

type seqResp struct {
	seq uint64
	err error
}

type seqReq struct {
	curr uint64
	ch   chan<- seqResp
}

type broadcast struct {
	donech chan<- error
	respch chan<- broadcastResp
	ctx    context.Context
	data   interface{}
	opts   *BroadcastOptions
}

type serialBroadcaster struct {
	ctx         context.Context
	cctx        sdkclient.Context
	svcClient   ttx.ServiceClient
	info        *keyring.Record
	reqch       chan broadcastReq
	broadcastch chan broadcast
	seqreqch    chan seqReq
	lc          lifecycle.Lifecycle
	nd          *node
	log         log.Logger
}

func newSerialTx(ctx context.Context, cctx sdkclient.Context, nd *node, opts ...cltypes.ClientOption) (*serialBroadcaster, sdkclient.Context, error) {
	var err error

	if !cctx.GenerateOnly {
		if err = validateBroadcastMode(cctx.BroadcastMode); err != nil {
			return nil, cctx, err
		}
	}

	key := cctx.From
	if key == "" {
		key = cctx.FromName
	}

	var info *keyring.Record
	if key != "" {
		info, err = cctx.Keyring.Key(key)
		if err != nil {
			info, err = cctx.Keyring.KeyByAddress(cctx.GetFromAddress())
		}

		if err != nil {
			return nil, cctx, err
		}

		if cctx.FromAddress == nil {
			addr, err := info.GetAddress()
			if err != nil {
				return nil, cctx, err
			}

			cctx = cctx.WithFromAddress(addr)
		}

		if cctx.From == "" {
			cctx = cctx.WithFrom(info.Name)
		}

		if cctx.FromName == "" {
			cctx = cctx.WithFromName(info.Name)
		}
	}

	txf, err := cltypes.NewTxFactory(cctx, opts...)
	if err != nil {
		return nil, cctx, err
	}

	client := &serialBroadcaster{
		ctx:         ctx,
		cctx:        cctx,
		svcClient:   ttx.NewServiceClient(cctx),
		info:        info,
		lc:          lifecycle.New(),
		reqch:       make(chan broadcastReq, 1),
		broadcastch: make(chan broadcast, 1),
		seqreqch:    make(chan seqReq),
		nd:          nd,
		log:         ctxlog.Logger(ctx).With("cmp", "client/broadcaster"),
	}

	go client.lc.WatchContext(ctx)
	go client.run()
	go client.broadcaster(txf)

	if !client.cctx.Offline {
		go client.sequenceSync()
	}

	return client, cctx, nil
}

// BroadcastMsgs builds and broadcasts transaction. This transaction is composed of 1 or more messages. This allows several
// operations to be performed in a single transaction.
// A transaction broadcast can be configured with an arbitrary number of BroadcastOption.
// This method returns the response as an interface{} instance. If an error occurs when preparing the transaction,
// an error is returned.
// A transaction can fail with a given "transaction code" which will not be passed to the error value.
// This needs to be checked by the caller and handled accordingly.
func (c *serialBroadcaster) BroadcastMsgs(ctx context.Context, msgs []sdk.Msg, opts ...BroadcastOption) (interface{}, error) {
	bOpts := &BroadcastOptions{
		confirmFn:        defaultTxConfirm,
		broadcastTimeout: BroadcastDefaultTimeout,
	}

	for _, opt := range opts {
		if err := opt(bOpts); err != nil {
			return nil, err
		}
	}

	// Validate all msgs before generating or broadcasting the tx.
	// We were calling ValidateBasic separately in each CLI handler before.
	// Right now, we're factorizing that call inside this function.
	// ref: https://github.com/cosmos/cosmos-sdk/pull/9236#discussion_r623803504
	for _, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return nil, err
		}
	}

	responsech := make(chan broadcastResp, 1)
	request := broadcastReq{
		ctx:        ctx,
		responsech: responsech,
		data:       msgs,
		opts:       bOpts,
	}

	select {
	case c.reqch <- request:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.lc.ShuttingDown():
		return nil, ErrNotRunning
	}

	select {
	case resp := <-responsech:
		// if returned error is sdk error, it is likely to be wrapped response so discard it
		// as clients supposed to check Tx code, unless resp is nil, which is error during Tx preparation
		if !errors.As(resp.err, &cerrors.Error{}) || resp.resp == nil || bOpts.resultAsError {
			if bOpts.resultAsError {
				if txResp, valid := resp.resp.(*sdk.TxResponse); valid && txResp.Code != 0 && resp.err == nil {
					resp.err = cerrors.ABCIError(txResp.Codespace, txResp.Code, txResp.RawLog)
				}
			}

			return resp.resp, resp.err
		}
		return resp.resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.lc.ShuttingDown():
		return nil, ErrNotRunning
	}
}

func (c *serialBroadcaster) BroadcastTx(ctx context.Context, tx sdk.Tx, opts ...BroadcastOption) (interface{}, error) {
	bOpts := &BroadcastOptions{
		confirmFn:        defaultTxConfirm,
		broadcastTimeout: BroadcastDefaultTimeout,
	}

	for _, opt := range opts {
		if err := opt(bOpts); err != nil {
			return nil, err
		}
	}

	responsech := make(chan broadcastResp, 1)
	request := broadcastReq{
		responsech: responsech,
		data:       tx,
		opts:       bOpts,
	}

	select {
	case c.reqch <- request:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.lc.ShuttingDown():
		return nil, ErrNotRunning
	}

	select {
	case resp := <-responsech:
		// if returned error is sdk error, it is likely to be wrapped response so discard it
		// as clients supposed to check Tx code, unless resp is nil, which is error during Tx preparation
		if !errors.As(resp.err, &cerrors.Error{}) || resp.resp == nil || bOpts.resultAsError {
			if bOpts.resultAsError {
				if txResp, valid := resp.resp.(*sdk.TxResponse); valid && txResp.Code != 0 && resp.err == nil {
					resp.err = cerrors.ABCIError(txResp.Codespace, txResp.Code, txResp.RawLog)
				}
			}

			return resp.resp, resp.err
		}
		return resp.resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.lc.ShuttingDown():
		return nil, ErrNotRunning
	}
}

func (c *serialBroadcaster) run() {
	defer c.lc.ShutdownCompleted()

	pending := deque.NewDeque[broadcastReq]()
	broadcastCh := c.broadcastch
	broadcastDoneCh := make(chan error, 1)

	tryBroadcast := func() {
		if pending.Len() == 0 {
			return
		}

		req := pending.Peek(0)

		select {
		case broadcastCh <- broadcast{
			donech: broadcastDoneCh,
			respch: req.responsech,
			ctx:    req.ctx,
			data:   req.data,
			opts:   req.opts,
		}:
			broadcastCh = nil
			_ = pending.PopFront()
		default:
		}
	}

loop:
	for {
		select {
		case err := <-c.lc.ShutdownRequest():
			c.lc.ShutdownInitiated(err)
			break loop
		case req := <-c.reqch:
			pending.PushBack(req)

			tryBroadcast()
		case err := <-broadcastDoneCh:
			broadcastCh = c.broadcastch

			if err != nil {
				c.log.Error("unable to broadcast messages", "error", err)
			}
			tryBroadcast()
		}
	}
}

func deriveTxfFromOptions(txf clienttx.Factory, opts *BroadcastOptions) clienttx.Factory {
	if opt := opts.note; opt != nil {
		txf = txf.WithMemo(*opt)
	}

	if opt := opts.gas; opt != nil {
		txf = txf.WithGas(opt.Gas).WithSimulateAndExecute(opt.Simulate)
	}

	if opt := opts.fees; opt != nil {
		txf = txf.WithFees(*opt)
	}

	if opt := opts.gasPrices; opt != nil {
		txf = txf.WithGasPrices(*opt)
	}

	if opt := opts.timeoutHeight; opt != nil {
		txf = txf.WithTimeoutHeight(*opt)
	}

	if opt := opts.gasAdjustment; opt != nil {
		txf = txf.WithGasAdjustment(*opt)
	}

	return txf
}

func deriveCctxFromOptions(cctx sdkclient.Context, opts *BroadcastOptions) sdkclient.Context {
	if opt := opts.broadcastMode; opt != nil {
		cctx = cctx.WithBroadcastMode(*opt)
	}

	if opt := opts.generateOnly; opt != nil {
		cctx = cctx.WithGenerateOnly(*opt)
	}

	return cctx
}

func (c *serialBroadcaster) syncSequence(f clienttx.Factory, resp interface{}, rErr error) (uint64, bool) {
	txResp, valid := resp.(*sdk.TxResponse)

	// due to cosmos-sdk not returning ABCI errors for /simulate call
	// exact error match does not work, and we have to improvise
	// use sdkerrors.ErrWrongSequence.Is(rErr) when /simulate call is fixed
	if (rErr != nil && (sdkerrors.ErrWrongSequence.Is(rErr) || sdkerrors.ErrInvalidSequence.Is(rErr))) || (valid && (txResp.Code == sdkerrors.ErrWrongSequence.ABCICode())) {
		// attempt to sync account sequence
		if rSeq, err := c.syncAccountSequence(f.Sequence()); err == nil {
			return rSeq, true
		}

		return f.Sequence(), true
	}

	return f.Sequence(), false
}

func (c *serialBroadcaster) broadcaster(ptxf clienttx.Factory) {
	for {
		select {
		case <-c.lc.ShuttingDown():
			return
		case req := <-c.broadcastch:
			var err error
			var resp interface{}

			cctx := deriveCctxFromOptions(c.cctx, req.opts)

			switch mType := req.data.(type) {
			case []sdk.Msg:
			done:
				for i := 0; i < 2; i++ {
					var seq uint64
					txf := deriveTxfFromOptions(ptxf, req.opts)

					resp, seq, err = c.buildAndBroadcastTx(req.ctx, cctx, txf, req.opts, mType)
					rSeq, synced := c.syncSequence(ptxf.WithSequence(seq), resp, err)
					ptxf = ptxf.WithSequence(rSeq)

					if !synced {
						break done
					}
				}
			case sdk.Tx:
				cctx := deriveCctxFromOptions(c.cctx, req.opts)
				resp, err = c.broadcastTx(req.ctx, cctx, req.opts.broadcastTimeout, mType)
			}

			req.respch <- broadcastResp{
				resp: resp,
				err:  err,
			}

			if c.info != nil {
				terr := &cerrors.Error{}
				if !cctx.GenerateOnly && errors.Is(err, terr) {
					rSeq, _ := c.syncSequence(ptxf, resp, err)
					ptxf = ptxf.WithSequence(rSeq)
				}
			}

			select {
			case <-c.lc.ShuttingDown():
				return
			case req.donech <- err:
			}
		}
	}
}

func (c *serialBroadcaster) buildAndBroadcastTx(
	ctx context.Context,
	cctx sdkclient.Context,
	txf clienttx.Factory,
	bOpts *BroadcastOptions,
	msgs []sdk.Msg,
) (interface{}, uint64, error) {
	var err error
	var res *sdk.TxResponse

	if txf.SimulateAndExecute() || cctx.Simulate {
		if cctx.Offline {
			return nil, txf.Sequence(), ErrSimulateOffline
		}

		simResp, adjusted, err := c.calculateGas(ctx, txf, msgs...)
		if errRes := CheckTendermintError(err, nil); errRes != nil {
			return errRes, txf.Sequence(), nil
		} else if err != nil {
			return nil, txf.Sequence(), err
		}

		// context Simulate differs from tx.Factory.simulate!
		// later is to calculate gas if one set to auto
		// if context has Simulate flag set, just bail out here with simulation result
		if cctx.Simulate {
			return simResp, txf.Sequence(), nil
		}

		txf = txf.WithGas(adjusted)
	}

	utx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, txf.Sequence(), err
	}

	if gAddr := cctx.GetFeeGranterAddress(); gAddr != nil {
		utx.SetFeeGranter(gAddr)
	}

	if cctx.GenerateOnly {
		txb, err := cctx.TxConfig.TxJSONEncoder()(utx.GetTx())
		if err != nil {
			return nil, txf.Sequence(), err
		}

		return txb, txf.Sequence(), nil
	}

	if !cctx.SkipConfirm {
		var out []byte
		if out, err = cctx.TxConfig.TxJSONEncoder()(utx.GetTx()); err != nil {
			return nil, txf.Sequence(), err
		}

		var shipIt bool
		if shipIt, err = bOpts.confirmFn(string(out)); err != nil {
			return nil, txf.Sequence(), err
		}

		if !shipIt {
			return nil, txf.Sequence(), ErrTxCanceledByUser
		}
	}

	if err = clienttx.Sign(ctx, txf, c.info.Name, utx, true); err != nil {
		return nil, txf.Sequence(), err
	}

	res, err = c.broadcastTx(ctx, cctx, bOpts.broadcastTimeout, utx.GetTx())
	if err != nil {
		return res, txf.Sequence(), err
	}

	txf = txf.WithSequence(txf.Sequence() + 1)

	if res.Code != 0 {
		return res, txf.Sequence(), cerrors.ABCIError(res.Codespace, res.Code, res.RawLog)
	}

	return res, txf.Sequence(), err
}

func (c *serialBroadcaster) sequenceSync() {
	for {
		select {
		case <-c.lc.ShuttingDown():
			return
		case req := <-c.seqreqch:
			// reply back with current value if any error to occur
			seq := seqResp{
				seq: req.curr,
			}

			ndStatus, err := c.nd.SyncInfo(c.ctx)
			if err != nil {
				c.log.Error("cannot obtain node status to sync account sequence", "err", err)
				seq.err = err
			}

			if err == nil && ndStatus.CatchingUp {
				c.log.Error("cannot sync account sequence from node that is catching up")
				err = ErrNodeCatchingUp
			}

			if err == nil {
				addr, _ := c.info.GetAddress()
				// query sequence number
				if _, seq.seq, err = c.cctx.AccountRetriever.GetAccountNumberSequence(c.cctx, addr); err != nil {
					c.log.Error("error requesting account", "err", err)
					seq.err = err
				}
			}

			select {
			case req.ch <- seq:
			case <-c.lc.ShuttingDown():
			}
		}
	}
}

func (c *serialBroadcaster) syncAccountSequence(lSeq uint64) (uint64, error) {
	ch := make(chan seqResp, 1)

	c.seqreqch <- seqReq{
		curr: lSeq,
		ch:   ch,
	}

	ctx, cancel := context.WithTimeout(c.ctx, sequenceSyncTimeout)
	defer cancel()

	select {
	case rSeq := <-ch:
		return rSeq.seq, rSeq.err
	case <-ctx.Done():
		return lSeq, ErrSyncTimedOut
	case <-c.lc.ShuttingDown():
		return lSeq, ErrNotRunning
	}
}

// broadcastTxb broadcasts fully built transaction in sync/async or block modes
// based on the context parameters. The result of the broadcast is parsed into
// an intermediate structure which is logged if the context has a logger
// defined.
func (c *serialBroadcaster) broadcastTx(ctx context.Context, cctx sdkclient.Context, timeout time.Duration, tx sdk.Tx) (*sdk.TxResponse, error) {
	txb, err := cctx.TxConfig.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	hash := cbtypes.Tx(txb).Hash()

	node, err := cctx.GetNode()
	if err != nil {
		return nil, err
	}

	var resp *sdk.TxResponse

	// broadcast mode has been validated
	switch cctx.BroadcastMode {
	case BroadcastBlock:
		var res *cbcoretypes.ResultBroadcastTxCommit
		res, err = node.BroadcastTxCommit(context.Background(), txb)
		if errRes := CheckTendermintError(err, txb); errRes != nil {
			return errRes, nil
		}

		resp = NewResponseFormatBroadcastTxCommit(res)
	case BroadcastSync:
		var res *cbcoretypes.ResultBroadcastTx
		res, err = node.BroadcastTxSync(context.Background(), txb)
		if errRes := CheckTendermintError(err, txb); errRes != nil {
			return errRes, nil
		}

		resp = sdk.NewResponseFormatBroadcastTx(res)
	case BroadcastAsync:
		var res *cbcoretypes.ResultBroadcastTx
		res, err = node.BroadcastTxAsync(context.Background(), txb)
		if errRes := CheckTendermintError(err, txb); errRes != nil {
			return errRes, nil
		}

		resp = sdk.NewResponseFormatBroadcastTx(res)
	}

	if err == nil {
		// good job
		return resp, nil
	} else if !strings.HasSuffix(err.Error(), timeoutErrorMessage) {
		return resp, err
	}

	// timeout error, continue on to retry
	// loop
	lctx, cancel := context.WithTimeout(c.ctx, timeout)
	defer cancel()

	for lctx.Err() == nil {
		// wait up to one second
		select {
		case <-lctx.Done():
			return resp, err
		case <-time.After(broadcastBlockRetryPeriod):
		}

		// check transaction
		// https://github.com/cosmos/cosmos-sdk/pull/8734
		res, err := nutils.QueryTx(ctx, cctx, hash)
		if err == nil {
			return res, nil
		}

		// if it's not a "not found" error, return
		if !strings.HasSuffix(err.Error(), notFoundErrorMessageSuffix) {
			return res, err
		}
	}

	return resp, lctx.Err()
}

// calculateGas simulates the execution of a transaction and returns the
// simulation response obtained by the query and the adjusted gas amount.
func (c *serialBroadcaster) calculateGas(
	ctx context.Context,
	txf clienttx.Factory,
	msgs ...sdk.Msg,
) (*ttx.SimulateResponse, uint64, error) {
	txBytes, err := txf.BuildSimTx(msgs...)
	if err != nil {
		return nil, 0, err
	}

	simRes, err := c.svcClient.Simulate(ctx, &ttx.SimulateRequest{
		TxBytes: txBytes,
	})
	if err != nil {
		return nil, 0, err
	}

	return simRes, uint64(txf.GasAdjustment() * float64(simRes.GasInfo.GasUsed)), nil
}

func defaultTxConfirm(txn string) (bool, error) {
	_, _ = fmt.Printf("%s\n\n", txn)

	buf := bufio.NewReader(os.Stdin)

	return input.GetConfirmation("confirm transaction before signing and broadcasting", buf, os.Stdin)
}

func validateBroadcastMode(val string) error {
	switch val {
	case BroadcastAsync:
		fallthrough
	case BroadcastSync:
		fallthrough
	case BroadcastBlock:
		return nil
	}

	return fmt.Errorf("invalid broadcast mode \"%s\". expected %s|%s|%s",
		val,
		BroadcastAsync,
		BroadcastSync,
		BroadcastBlock,
	)
}

// CheckTendermintError checks if the error returned from BroadcastTx is a
// Tendermint error that is returned before the tx is submitted due to
// precondition checks that failed. If an Tendermint error is detected, this
// function returns the correct code back in TxResponse.
//
// TODO: Avoid brittle string matching in favor of error matching. This requires
// a change to Tendermint's RPCError type to allow retrieval or matching against
// a concrete error type.
func CheckTendermintError(err error, tx cbtypes.Tx) *sdk.TxResponse {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	var txHash string

	if tx != nil {
		txHash = fmt.Sprintf("%X", tx.Hash())
	}

	switch {
	case sequenceMismatchRegexp.MatchString(err.Error()):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrWrongSequence.ABCICode(),
			Codespace: sdkerrors.ErrWrongSequence.Codespace(),
			TxHash:    txHash,
		}
	case strings.Contains(errStr, strings.ToLower(mempool.ErrTxInCache.Error())):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrTxInMempoolCache.ABCICode(),
			Codespace: sdkerrors.ErrTxInMempoolCache.Codespace(),
			TxHash:    txHash,
		}
	case strings.Contains(errStr, "mempool is full"):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrMempoolIsFull.ABCICode(),
			Codespace: sdkerrors.ErrMempoolIsFull.Codespace(),
			TxHash:    txHash,
		}
	case strings.Contains(errStr, "tx too large"):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrTxTooLarge.ABCICode(),
			Codespace: sdkerrors.ErrTxTooLarge.Codespace(),
			TxHash:    txHash,
		}
	}

	return nil
}
