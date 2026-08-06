package app

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"

	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	utypes "github.com/virtengine/virtengine/upgrades/types"
)

// ProposalRejectionCode is a stable, consensus-policy rejection identifier.
type ProposalRejectionCode string

const (
	ProposalRejectEmptyTransaction        ProposalRejectionCode = "VE-CONS-0001"
	ProposalRejectMalformedTransaction    ProposalRejectionCode = "VE-CONS-0002"
	ProposalRejectNonCanonicalTransaction ProposalRejectionCode = "VE-CONS-0003"
	ProposalRejectDuplicateTransaction    ProposalRejectionCode = "VE-CONS-0004"
	ProposalRejectTransactionTooLarge     ProposalRejectionCode = "VE-CONS-0005"
	ProposalRejectProposalTooLarge        ProposalRejectionCode = "VE-CONS-0006"
	ProposalRejectTooManyTransactions     ProposalRejectionCode = "VE-CONS-0007"
	ProposalRejectGasLimit                ProposalRejectionCode = "VE-CONS-0008"
	ProposalRejectAnteHandler             ProposalRejectionCode = "VE-CONS-0009"
	ProposalRejectActivationUnavailable   ProposalRejectionCode = "VE-CONS-0010"
	ProposalRejectSystemTransaction       ProposalRejectionCode = "VE-CONS-0011"
)

// ProposalValidationError reports a stable rejection code and transaction index.
// Detail is diagnostic only and must never be parsed as consensus input.
type ProposalValidationError struct {
	Code    ProposalRejectionCode
	TxIndex int
	Detail  string
}

func (e *ProposalValidationError) Error() string {
	if e.TxIndex >= 0 {
		return fmt.Sprintf("%s: transaction %d: %s", e.Code, e.TxIndex, e.Detail)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Detail)
}

// ProposalErrorCode extracts a stable rejection code from an error.
func ProposalErrorCode(err error) ProposalRejectionCode {
	var validationErr *ProposalValidationError
	if errors.As(err, &validationErr) {
		return validationErr.Code
	}
	return ""
}

// ProposalLimits bounds transaction decoding, memory, and ante-handler work.
type ProposalLimits struct {
	MaxTxCount          int
	MaxCandidateTxCount int
	MaxTxBytes          int
	MaxProposalBytes    int64
}

// DefaultProposalLimits returns conservative, deterministic admission limits.
func DefaultProposalLimits() ProposalLimits {
	return ProposalLimits{
		MaxTxCount:          5_000,
		MaxCandidateTxCount: 10_000,
		MaxTxBytes:          1 << 20, // 1 MiB per ordinary Cosmos SDK transaction.
		MaxProposalBytes:    16 << 20,
	}
}

type proposalTxVerifier interface {
	PrepareProposalVerifyTx(tx sdk.Tx) ([]byte, error)
	ProcessProposalVerifyTx(txBz []byte) (sdk.Tx, error)
	TxDecode(txBz []byte) (sdk.Tx, error)
	TxEncode(tx sdk.Tx) ([]byte, error)
}

type proposalGasTx interface {
	GetGas() uint64
}

// CanonicalProposalValidator shares byte-level admission rules between prepare
// and process. It only decodes ordinary Cosmos SDK transactions with the app's
// configured decoder; Task 84A deliberately defines no synthetic system format.
type CanonicalProposalValidator struct {
	verifier proposalTxVerifier
	limits   ProposalLimits
}

// NewCanonicalProposalValidator creates a bounded canonical validator.
func NewCanonicalProposalValidator(verifier proposalTxVerifier, limits ProposalLimits) *CanonicalProposalValidator {
	if verifier == nil {
		panic("proposal transaction verifier is required")
	}
	return &CanonicalProposalValidator{verifier: verifier, limits: normalizeProposalLimits(limits)}
}

func normalizeProposalLimits(limits ProposalLimits) ProposalLimits {
	defaults := DefaultProposalLimits()
	if limits.MaxTxCount <= 0 {
		limits.MaxTxCount = defaults.MaxTxCount
	}
	if limits.MaxCandidateTxCount <= 0 {
		limits.MaxCandidateTxCount = defaults.MaxCandidateTxCount
	}
	if limits.MaxCandidateTxCount < limits.MaxTxCount {
		limits.MaxCandidateTxCount = limits.MaxTxCount
	}
	if limits.MaxTxBytes <= 0 {
		limits.MaxTxBytes = defaults.MaxTxBytes
	}
	if limits.MaxProposalBytes <= 0 {
		limits.MaxProposalBytes = defaults.MaxProposalBytes
	}
	return limits
}

// Validate decodes and validates canonical transaction bytes, duplicate hashes,
// byte limits, transaction-count limits, and the consensus gas limit.
func (v *CanonicalProposalValidator) Validate(txs [][]byte, maxBytes int64, maxGas ...int64) ([]sdk.Tx, error) {
	if len(txs) > v.limits.MaxTxCount {
		return nil, proposalError(ProposalRejectTooManyTransactions, -1, "transaction count exceeds policy limit")
	}

	byteLimit := v.effectiveByteLimit(maxBytes)
	decoded := make([]sdk.Tx, 0, len(txs))
	seen := make(map[[sha256.Size]byte]struct{}, len(txs))
	var totalBytes int64
	var totalGas uint64
	gasLimit := int64(0)
	if len(maxGas) > 0 {
		gasLimit = maxGas[0]
	}

	for index, txBytes := range txs {
		tx, hash, err := v.inspectTransaction(txBytes, index)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[hash]; exists {
			return nil, proposalError(ProposalRejectDuplicateTransaction, index, "duplicate canonical transaction identifier")
		}
		seen[hash] = struct{}{}

		txSize := proposalTransactionSize(txBytes)
		if txSize > byteLimit-totalBytes {
			return nil, proposalError(ProposalRejectProposalTooLarge, index, "proposal byte limit exceeded")
		}
		totalBytes += txSize

		if gasLimit > 0 {
			gasTx, ok := tx.(proposalGasTx)
			if !ok {
				return nil, proposalError(ProposalRejectGasLimit, index, "transaction does not expose a gas limit")
			}
			gas := gasTx.GetGas()
			if gas > ^uint64(0)-totalGas {
				return nil, proposalError(ProposalRejectGasLimit, index, "transaction gas overflow")
			}
			totalGas += gas
			if totalGas > uint64(gasLimit) { //nolint:gosec // positivity checked before signed-to-unsigned conversion
				return nil, proposalError(ProposalRejectGasLimit, index, "proposal gas limit exceeded")
			}
		}

		decoded = append(decoded, tx)
	}

	return decoded, nil
}

func (v *CanonicalProposalValidator) inspectTransaction(txBytes []byte, index int) (sdk.Tx, [sha256.Size]byte, error) {
	if len(txBytes) == 0 {
		return nil, [sha256.Size]byte{}, proposalError(ProposalRejectEmptyTransaction, index, "empty transaction bytes")
	}
	if len(txBytes) > v.limits.MaxTxBytes {
		return nil, [sha256.Size]byte{}, proposalError(ProposalRejectTransactionTooLarge, index, "transaction byte limit exceeded")
	}

	tx, err := v.verifier.TxDecode(txBytes)
	if err != nil {
		return nil, [sha256.Size]byte{}, proposalError(ProposalRejectMalformedTransaction, index, "SDK transaction decoder rejected bytes")
	}
	canonical, err := v.verifier.TxEncode(tx)
	if err != nil {
		return nil, [sha256.Size]byte{}, proposalError(ProposalRejectMalformedTransaction, index, "SDK transaction encoder rejected transaction")
	}
	if !bytes.Equal(canonical, txBytes) {
		return nil, [sha256.Size]byte{}, proposalError(ProposalRejectNonCanonicalTransaction, index, "transaction bytes are not canonical SDK encoding")
	}

	return tx, sha256.Sum256(canonical), nil
}

func (v *CanonicalProposalValidator) effectiveByteLimit(requestLimit int64) int64 {
	if requestLimit <= 0 || requestLimit > v.limits.MaxProposalBytes {
		return v.limits.MaxProposalBytes
	}
	return requestLimit
}

func proposalError(code ProposalRejectionCode, index int, detail string) error {
	return &ProposalValidationError{Code: code, TxIndex: index, Detail: detail}
}

// ProposalActivation supplies the persisted activation decision.
type ProposalActivation interface {
	Active(ctx sdk.Context, proposalHeight int64) (bool, error)
}

type upgradeProposalActivation struct {
	keeper *upgradekeeper.Keeper
}

// NewUpgradeProposalActivation activates strict admission on the block after
// the registered software upgrade has committed its done marker.
func NewUpgradeProposalActivation(keeper *upgradekeeper.Keeper) ProposalActivation {
	return upgradeProposalActivation{keeper: keeper}
}

func (a upgradeProposalActivation) Active(ctx sdk.Context, proposalHeight int64) (bool, error) {
	if a.keeper == nil {
		return false, errors.New("upgrade keeper is not configured")
	}
	if ctx.MultiStore() == nil {
		return false, errors.New("proposal context has no multistore")
	}
	doneHeight, err := a.keeper.GetDoneHeight(ctx, utypes.ConsensusAdmissionUpgradeName)
	if err != nil {
		return false, err
	}
	return proposalActiveAfterDoneHeight(doneHeight, proposalHeight)
}

func proposalActiveAfterDoneHeight(doneHeight, proposalHeight int64) (bool, error) {
	if doneHeight < 0 || doneHeight > proposalHeight {
		return false, fmt.Errorf("invalid persisted upgrade done height %d for proposal height %d", doneHeight, proposalHeight)
	}
	return doneHeight > 0 && proposalHeight > doneHeight, nil
}

// ProposalHandler composes SDK default selection with strict canonical checks.
type ProposalHandler struct {
	verifier   proposalTxVerifier
	validator  *CanonicalProposalValidator
	sdkPrepare sdk.PrepareProposalHandler
	activation ProposalActivation
	systemTx   SystemTxCodec
	limits     ProposalLimits
	observer   func(string, error)
	skipAnte   bool
}

// ProposalHandlerOption configures test-only observational behavior.
type ProposalHandlerOption func(*ProposalHandler)

// WithProposalObserver installs an observer that receives diagnostic stages
// and errors. Its return value is ignored and cannot influence consensus.
func WithProposalObserver(observer func(string, error)) ProposalHandlerOption {
	return func(handler *ProposalHandler) {
		handler.observer = observer
	}
}

// WithCheckTxPrevalidatedProposals is test-only. It is safe only when the exact
// transaction batch was already validated through CheckTx.
func WithCheckTxPrevalidatedProposals() ProposalHandlerOption {
	return func(handler *ProposalHandler) {
		handler.skipAnte = true
	}
}

// NewProposalHandler creates the deterministic ABCI++ admission handler.
func NewProposalHandler(
	verifier proposalTxVerifier,
	sdkPrepare sdk.PrepareProposalHandler,
	activation ProposalActivation,
	limits ProposalLimits,
	systemTx ...SystemTxCodec,
) *ProposalHandler {
	if verifier == nil {
		panic("proposal transaction verifier is required")
	}
	if sdkPrepare == nil {
		panic("SDK prepare-proposal handler is required")
	}
	if activation == nil {
		panic("proposal activation source is required")
	}
	limits = normalizeProposalLimits(limits)
	var systemCodec SystemTxCodec
	if len(systemTx) > 0 {
		systemCodec = systemTx[0]
	}
	return &ProposalHandler{
		verifier:   verifier,
		validator:  NewCanonicalProposalValidator(verifier, limits),
		sdkPrepare: sdkPrepare,
		activation: activation,
		systemTx:   systemCodec,
		limits:     limits,
	}
}

func (h *ProposalHandler) observe(stage string, err error) {
	if h.observer != nil {
		h.observer(stage, err)
	}
}

// FailClosedPrepareProposal prevents BaseApp's compatibility error recovery from
// restoring the unvalidated CometBFT candidate list. A proposer that cannot
// determine or apply strict policy proposes an empty block instead.
func FailClosedPrepareProposal(handler sdk.PrepareProposalHandler) sdk.PrepareProposalHandler {
	if handler == nil {
		panic("prepare-proposal handler is required")
	}
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (response *abci.ResponsePrepareProposal, err error) {
		defer func() {
			if recover() != nil {
				response = &abci.ResponsePrepareProposal{Txs: nil}
				err = nil
			}
		}()
		response, err = handler(ctx, req)
		if err != nil || response == nil {
			return &abci.ResponsePrepareProposal{Txs: nil}, nil
		}
		return response, nil
	}
}

// PrepareProposal preserves legacy pass-through behavior before activation. Once
// active it bounds candidate inspection, removes malformed/duplicate/invalid
// transactions, and delegates ordering/selection to the SDK default handler.
func (h *ProposalHandler) PrepareProposal(ctx sdk.Context, req *abci.RequestPrepareProposal) (resp *abci.ResponsePrepareProposal, err error) {
	if req == nil {
		return nil, proposalError(ProposalRejectMalformedTransaction, -1, "nil prepare-proposal request")
	}
	active, err := h.activation.Active(ctx, req.Height)
	if err != nil {
		return &abci.ResponsePrepareProposal{Txs: nil}, nil
	}
	if !active {
		return &abci.ResponsePrepareProposal{Txs: cloneTransactions(req.Txs)}, nil
	}
	if h.systemTx == nil {
		return &abci.ResponsePrepareProposal{Txs: nil}, nil
	}
	// The current BaseApp fork replaces PrepareProposal failures with req.Txs.
	// Contain active-policy failures here so unfiltered bytes cannot bypass the
	// application. Empty proposals are valid and preserve consensus liveness.
	defer func() {
		if recover() != nil {
			resp = &abci.ResponsePrepareProposal{Txs: nil}
			err = nil
		}
	}()

	totalByteLimit := h.validator.effectiveByteLimit(req.MaxTxBytes)
	systemTx, buildErr := h.systemTx.Build(ctx, req.LocalLastCommit)
	if buildErr != nil {
		h.observe("build-system-transaction", buildErr)
		return &abci.ResponsePrepareProposal{Txs: nil}, nil
	}
	ordinaryByteLimit := totalByteLimit - proposalTransactionSize(systemTx)
	if ordinaryByteLimit < 0 {
		return &abci.ResponsePrepareProposal{Txs: nil}, nil
	}
	candidates := h.filterCandidates(req.Txs, ordinaryByteLimit)
	filteredReq := *req
	filteredReq.Txs = candidates
	filteredReq.MaxTxBytes = ordinaryByteLimit

	selected, err := h.sdkPrepare(ctx, &filteredReq)
	if err != nil {
		h.observe("sdk-prepare", err)
		return &abci.ResponsePrepareProposal{Txs: nil}, nil
	}
	if selected == nil {
		return &abci.ResponsePrepareProposal{Txs: nil}, nil
	}

	allowed := make(map[[sha256.Size]byte]struct{}, len(candidates))
	for _, txBytes := range candidates {
		allowed[sha256.Sum256(txBytes)] = struct{}{}
	}
	verified := make([][]byte, 0, min(len(selected.Txs), h.limits.MaxTxCount))
	for _, txBytes := range selected.Txs {
		if len(verified) >= h.limits.MaxTxCount {
			break
		}
		if _, ok := allowed[sha256.Sum256(txBytes)]; !ok {
			h.observe("sdk-selected-unrequested-transaction", proposalError(ProposalRejectMalformedTransaction, -1, "SDK selected transaction outside bounded candidate set"))
			return &abci.ResponsePrepareProposal{Txs: nil}, nil
		}
		decoded, validateErr := h.validator.Validate([][]byte{txBytes}, filteredReq.MaxTxBytes)
		if validateErr != nil {
			continue
		}
		if !h.skipAnte {
			canonical, verifyErr := h.verifier.PrepareProposalVerifyTx(decoded[0])
			if verifyErr != nil || !bytes.Equal(canonical, txBytes) {
				continue
			}
		}
		verified = append(verified, bytes.Clone(txBytes))
	}

	if _, err := h.validator.Validate(verified, ordinaryByteLimit, proposalMaxGas(ctx)); err != nil {
		h.observe("validate-prepared-transactions", err)
		return &abci.ResponsePrepareProposal{Txs: nil}, nil
	}
	result := make([][]byte, 0, len(verified)+1)
	result = append(result, bytes.Clone(systemTx))
	result = append(result, verified...)
	if proposalTransactionsSize(result) > totalByteLimit {
		h.observe("validate-prepared-size", proposalError(ProposalRejectProposalTooLarge, -1, "prepared proposal byte limit exceeded"))
		return &abci.ResponsePrepareProposal{Txs: nil}, nil
	}
	return &abci.ResponsePrepareProposal{Txs: result}, nil
}

func (h *ProposalHandler) filterCandidates(txs [][]byte, maxBytes int64) [][]byte {
	byteLimit := h.validator.effectiveByteLimit(maxBytes)
	result := make([][]byte, 0, min(len(txs), h.limits.MaxCandidateTxCount))
	seen := make(map[[sha256.Size]byte]struct{}, min(len(txs), h.limits.MaxCandidateTxCount))
	var inspectedBytes int64

	candidateCount := min(len(txs), h.limits.MaxCandidateTxCount)
	for index := 0; index < candidateCount; index++ {
		txBytes := txs[index]
		if h.systemTx != nil {
			if _, isSystem, err := h.systemTx.Decode(txBytes); err != nil || isSystem {
				continue
			}
		}
		txSize := proposalTransactionSize(txBytes)
		if txSize > byteLimit-inspectedBytes {
			break
		}
		inspectedBytes += txSize
		_, hash, err := h.validator.inspectTransaction(txBytes, index)
		if err != nil {
			continue
		}
		if _, exists := seen[hash]; exists {
			continue
		}
		seen[hash] = struct{}{}
		result = append(result, bytes.Clone(txBytes))
	}
	return result
}

// ProcessProposal rejects any non-canonical, duplicate, over-budget, or
// ante-invalid proposal after activation. There is no fail-open error branch.
func (h *ProposalHandler) ProcessProposal(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	if req == nil {
		return rejectProposal(), nil
	}
	active, err := h.activation.Active(ctx, req.Height)
	if err != nil {
		return rejectProposal(), nil
	}
	if !active {
		return acceptProposal(), nil
	}
	if h.systemTx == nil || len(req.Txs) == 0 {
		return rejectProposal(), nil
	}
	systemMsg, isSystem, err := h.systemTx.Decode(req.Txs[0])
	if err != nil || !isSystem {
		return rejectProposal(), nil
	}
	if err := h.systemTx.Validate(ctx, systemMsg, req.ProposedLastCommit); err != nil {
		return rejectProposal(), nil
	}
	for index := 1; index < len(req.Txs); index++ {
		if _, duplicateSystem, decodeErr := h.systemTx.Decode(req.Txs[index]); decodeErr != nil || duplicateSystem {
			return rejectProposal(), nil
		}
	}

	ordinary := req.Txs[1:]
	ordinaryByteLimit := h.limits.MaxProposalBytes - proposalTransactionSize(req.Txs[0])
	decoded, err := h.validator.Validate(ordinary, ordinaryByteLimit, proposalMaxGas(ctx))
	if err != nil {
		return rejectProposal(), nil
	}
	if !h.skipAnte {
		for index, txBytes := range ordinary {
			verifiedTx, verifyErr := h.verifier.ProcessProposalVerifyTx(txBytes)
			if verifyErr != nil {
				return rejectProposal(), nil
			}
			// The canonical decoder and BaseApp verifier must agree on transaction type.
			if verifiedTx == nil || decoded[index] == nil {
				return rejectProposal(), nil
			}
		}
	}
	return acceptProposal(), nil
}

// AuthorizeFinalizeBlock independently revalidates the finalized proposal and
// marks only its exact index-zero system transaction for ante/message execution.
// This protects replay and block-sync paths where ProcessProposal is skipped.
func (h *ProposalHandler) AuthorizeFinalizeBlock(ctx sdk.Context, req *abci.RequestFinalizeBlock) (sdk.Context, error) {
	if req == nil {
		return ctx, proposalError(ProposalRejectSystemTransaction, -1, "nil finalize-block request")
	}
	active, err := h.activation.Active(ctx, req.Height)
	if err != nil {
		return ctx, proposalError(ProposalRejectActivationUnavailable, -1, "persisted activation state unavailable")
	}
	if !active {
		return ctx, nil
	}
	if h.systemTx == nil || len(req.Txs) == 0 {
		return ctx, proposalError(ProposalRejectSystemTransaction, -1, "missing index-zero system transaction")
	}
	msg, isSystem, err := h.systemTx.Decode(req.Txs[0])
	if err != nil || !isSystem {
		return ctx, proposalError(ProposalRejectSystemTransaction, 0, "invalid index-zero system transaction")
	}
	if err := h.systemTx.Validate(ctx, msg, req.DecidedLastCommit); err != nil {
		return ctx, proposalError(ProposalRejectSystemTransaction, 0, "system transaction evidence validation failed")
	}
	for index := 1; index < len(req.Txs); index++ {
		if _, duplicateSystem, decodeErr := h.systemTx.Decode(req.Txs[index]); decodeErr != nil || duplicateSystem {
			return ctx, proposalError(ProposalRejectSystemTransaction, index, "duplicate or malformed system transaction")
		}
	}
	return authorizeSystemTransaction(ctx, req.Txs[0]), nil
}

func proposalTransactionSize(txBytes []byte) int64 {
	return cmttypes.ComputeProtoSizeForTxs([]cmttypes.Tx{txBytes})
}

func proposalTransactionsSize(txs [][]byte) int64 {
	converted := make([]cmttypes.Tx, len(txs))
	for i := range txs {
		converted[i] = txs[i]
	}
	return cmttypes.ComputeProtoSizeForTxs(converted)
}

func proposalMaxGas(ctx sdk.Context) int64 {
	if block := ctx.ConsensusParams().Block; block != nil {
		return block.MaxGas
	}
	return 0
}

func cloneTransactions(txs [][]byte) [][]byte {
	cloned := make([][]byte, len(txs))
	for i, tx := range txs {
		cloned[i] = bytes.Clone(tx)
	}
	return cloned
}

func acceptProposal() *abci.ResponseProcessProposal {
	return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}
}

func rejectProposal() *abci.ResponseProcessProposal {
	return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}
}
