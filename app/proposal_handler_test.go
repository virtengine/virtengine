package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/require"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
)

type testProposalTx struct {
	bytes []byte
	gas   uint64
	msgs  []sdk.Msg
}

func (tx testProposalTx) GetMsgs() []sdk.Msg { return tx.msgs }

func (tx testProposalTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }

func (tx testProposalTx) GetGas() uint64 { return tx.gas }

type testProposalTxWithoutGas struct {
	bytes []byte
}

func (tx testProposalTxWithoutGas) GetMsgs() []sdk.Msg { return nil }

func (tx testProposalTxWithoutGas) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }

type testProposalVerifier struct{}

func (testProposalVerifier) TxDecode(txBytes []byte) (sdk.Tx, error) {
	if len(txBytes) == 0 || bytes.HasPrefix(txBytes, []byte("malformed")) || bytes.HasPrefix(txBytes, []byte("VE84A_SYSTEM")) {
		return nil, errors.New("cannot decode transaction")
	}
	gas := uint64(1)
	if bytes.HasPrefix(txBytes, []byte("gas-10")) {
		gas = 10
	}
	return testProposalTx{bytes: bytes.Clone(txBytes), gas: gas}, nil
}

func (testProposalVerifier) TxEncode(tx sdk.Tx) ([]byte, error) {
	testTx := tx.(testProposalTx)
	if bytes.HasPrefix(testTx.bytes, []byte("noncanonical:")) {
		return bytes.TrimPrefix(testTx.bytes, []byte("noncanonical:")), nil
	}
	return bytes.Clone(testTx.bytes), nil
}

func (v testProposalVerifier) PrepareProposalVerifyTx(tx sdk.Tx) ([]byte, error) {
	encoded, err := v.TxEncode(tx)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(encoded, []byte("ante-invalid")) {
		return nil, errors.New("ante rejected transaction")
	}
	return encoded, nil
}

func (v testProposalVerifier) ProcessProposalVerifyTx(txBytes []byte) (sdk.Tx, error) {
	tx, err := v.TxDecode(txBytes)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(txBytes, []byte("ante-invalid")) {
		return nil, errors.New("ante rejected transaction")
	}
	return tx, nil
}

type staticProposalActivation struct {
	active bool
	err    error
}

func (a staticProposalActivation) Active(_ sdk.Context, _ int64) (bool, error) {
	return a.active, a.err
}

func passthroughPrepareProposal(_ sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	return &abci.ResponsePrepareProposal{Txs: req.Txs}, nil
}

func testSystemCodec() staticSystemTxCodec {
	return staticSystemTxCodec{systemTx: []byte("system"), msg: &veidv1.MsgSubmitConsensusVerification{Height: 100}}
}

func TestProposalHandlerConstructorsRequireDependencies(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() { NewCanonicalProposalValidator(nil, DefaultProposalLimits()) })
	require.Panics(t, func() {
		NewProposalHandler(nil, passthroughPrepareProposal, staticProposalActivation{}, DefaultProposalLimits())
	})
	require.Panics(t, func() {
		NewProposalHandler(testProposalVerifier{}, nil, staticProposalActivation{}, DefaultProposalLimits())
	})
	require.Panics(t, func() {
		NewProposalHandler(testProposalVerifier{}, passthroughPrepareProposal, nil, DefaultProposalLimits())
	})
}

func TestNormalizeProposalLimitsPreservesExplicitBounds(t *testing.T) {
	t.Parallel()

	limits := normalizeProposalLimits(ProposalLimits{
		MaxTxCount:          3,
		MaxCandidateTxCount: 6,
		MaxTxBytes:          7,
		MaxProposalBytes:    8,
	})
	require.Equal(t, 3, limits.MaxTxCount)
	require.Equal(t, 6, limits.MaxCandidateTxCount)
	require.Equal(t, 7, limits.MaxTxBytes)
	require.Equal(t, int64(8), limits.MaxProposalBytes)

	limits = normalizeProposalLimits(ProposalLimits{MaxTxCount: 5, MaxCandidateTxCount: 2})
	require.Equal(t, 5, limits.MaxCandidateTxCount)
}

func TestCanonicalProposalValidatorStableRejectionCodes(t *testing.T) {
	t.Parallel()

	limits := ProposalLimits{MaxTxCount: 2, MaxCandidateTxCount: 3, MaxTxBytes: 16, MaxProposalBytes: 24}
	validator := NewCanonicalProposalValidator(testProposalVerifier{}, limits)

	tests := []struct {
		name string
		txs  [][]byte
		code ProposalRejectionCode
	}{
		{name: "empty", txs: [][]byte{nil}, code: ProposalRejectEmptyTransaction},
		{name: "malformed", txs: [][]byte{[]byte("malformed")}, code: ProposalRejectMalformedTransaction},
		{name: "non canonical", txs: [][]byte{[]byte("noncanonical:tx")}, code: ProposalRejectNonCanonicalTransaction},
		{name: "duplicate", txs: [][]byte{[]byte("one"), []byte("one")}, code: ProposalRejectDuplicateTransaction},
		{name: "tx too large", txs: [][]byte{bytes.Repeat([]byte{'x'}, 17)}, code: ProposalRejectTransactionTooLarge},
		{name: "too many", txs: [][]byte{[]byte("one"), []byte("two"), []byte("three")}, code: ProposalRejectTooManyTransactions},
		{name: "proposal too large", txs: [][]byte{bytes.Repeat([]byte{'a'}, 13), bytes.Repeat([]byte{'b'}, 13)}, code: ProposalRejectProposalTooLarge},
		{name: "reserved system carrier at zero", txs: [][]byte{[]byte("VE84A_SYSTEM:v1")}, code: ProposalRejectMalformedTransaction},
		{name: "reserved system carrier at nonzero", txs: [][]byte{[]byte("ordinary"), []byte("VE84A_SYSTEM:v1")}, code: ProposalRejectMalformedTransaction},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validator.Validate(tc.txs, 0)
			require.Error(t, err)
			require.Equal(t, tc.code, ProposalErrorCode(err))
		})
	}
}

func TestProposalHandlerPrepareIsBoundedAndDeterministic(t *testing.T) {
	t.Parallel()

	limits := ProposalLimits{MaxTxCount: 3, MaxCandidateTxCount: 6, MaxTxBytes: 32, MaxProposalBytes: 64}
	handler := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{active: true},
		limits,
		testSystemCodec(),
	)
	req := &abci.RequestPrepareProposal{
		MaxTxBytes: 64,
		Height:     101,
		Txs: [][]byte{
			[]byte("one"),
			[]byte("malformed"),
			[]byte("one"),
			[]byte("ante-invalid"),
			[]byte("two"),
			[]byte("three"),
		},
	}

	first, err := handler.PrepareProposal(sdk.Context{}, req)
	require.NoError(t, err)
	second, err := handler.PrepareProposal(sdk.Context{}, req)
	require.NoError(t, err)
	require.Equal(t, [][]byte{[]byte("system"), []byte("one"), []byte("two"), []byte("three")}, first.Txs)
	require.Equal(t, first.Txs, second.Txs)
}

func TestProposalHandlerPrepareDoesNotMutateRequestOrCandidateBytes(t *testing.T) {
	t.Parallel()

	handler := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{active: true},
		DefaultProposalLimits(),
		testSystemCodec(),
	)
	req := &abci.RequestPrepareProposal{
		Height:     101,
		MaxTxBytes: 64,
		Txs:        [][]byte{[]byte("one"), []byte("two")},
	}
	original := cloneTransactions(req.Txs)

	prepared, err := handler.PrepareProposal(sdk.Context{}, req)
	require.NoError(t, err)
	require.Equal(t, original, req.Txs)
	prepared.Txs[0][0] = 'X'
	require.Equal(t, original, req.Txs)
}

func TestProposalHandlerBoundsCandidateDecodeBytes(t *testing.T) {
	t.Parallel()

	decoded := 0
	verifier := proposalVerifierFunc{
		decode: func(txBytes []byte) (sdk.Tx, error) {
			decoded++
			return testProposalTx{bytes: bytes.Clone(txBytes), gas: 1}, nil
		},
		encode: func(tx sdk.Tx) ([]byte, error) { return bytes.Clone(tx.(testProposalTx).bytes), nil },
	}
	handler := NewProposalHandler(
		verifier,
		passthroughPrepareProposal,
		staticProposalActivation{active: true},
		ProposalLimits{MaxTxCount: 3, MaxCandidateTxCount: 3, MaxTxBytes: 32, MaxProposalBytes: 64},
		testSystemCodec(),
	)
	prepared, err := handler.PrepareProposal(sdk.Context{}, &abci.RequestPrepareProposal{
		Height:     101,
		MaxTxBytes: proposalTransactionsSize([][]byte{[]byte("system"), []byte("one")}),
		Txs:        [][]byte{[]byte("one"), []byte("two"), []byte("three")},
	})
	require.NoError(t, err)
	require.Equal(t, [][]byte{[]byte("system"), []byte("one")}, prepared.Txs)
	require.Equal(t, 3, decoded, "one canonical inspection, SDK decode, and ante decode are expected only for the admitted candidate")
}

func TestProposalHandlerUsesCometTransactionWireSize(t *testing.T) {
	t.Parallel()

	validator := NewCanonicalProposalValidator(testProposalVerifier{}, ProposalLimits{
		MaxTxCount:          2,
		MaxCandidateTxCount: 2,
		MaxTxBytes:          8,
		MaxProposalBytes:    8,
	})

	_, err := validator.Validate([][]byte{[]byte("1234")}, 4)
	require.Error(t, err)
	require.Equal(t, ProposalRejectProposalTooLarge, ProposalErrorCode(err))
}

func TestProposalHandlerGasAccountingRejectsOverflowAndMissingGas(t *testing.T) {
	t.Parallel()

	validator := NewCanonicalProposalValidator(testProposalVerifier{}, DefaultProposalLimits())
	_, err := validator.Validate([][]byte{[]byte("ordinary")}, 0, 0)
	require.NoError(t, err, "non-positive consensus MaxGas is unlimited")

	overflowVerifier := proposalVerifierFunc{
		decode: func(txBytes []byte) (sdk.Tx, error) {
			gas := ^uint64(0)
			if bytes.Equal(txBytes, []byte("second")) {
				gas = 1
			}
			return testProposalTx{bytes: bytes.Clone(txBytes), gas: gas}, nil
		},
		encode: func(tx sdk.Tx) ([]byte, error) { return bytes.Clone(tx.(testProposalTx).bytes), nil },
	}
	overflowValidator := NewCanonicalProposalValidator(overflowVerifier, DefaultProposalLimits())
	_, err = overflowValidator.Validate([][]byte{[]byte("first"), []byte("second")}, 0, 1)
	require.Error(t, err)
	require.Equal(t, ProposalRejectGasLimit, ProposalErrorCode(err))

	missingGasVerifier := proposalVerifierFunc{
		decode: func(txBytes []byte) (sdk.Tx, error) {
			return testProposalTxWithoutGas{bytes: bytes.Clone(txBytes)}, nil
		},
		encode: func(tx sdk.Tx) ([]byte, error) { return bytes.Clone(tx.(testProposalTxWithoutGas).bytes), nil },
	}
	missingGasValidator := NewCanonicalProposalValidator(missingGasVerifier, DefaultProposalLimits())
	_, err = missingGasValidator.Validate([][]byte{[]byte("ordinary")}, 0, 1)
	require.Error(t, err)
	require.Equal(t, ProposalRejectGasLimit, ProposalErrorCode(err))
}

type proposalVerifierFunc struct {
	decode func([]byte) (sdk.Tx, error)
	encode func(sdk.Tx) ([]byte, error)
}

func (v proposalVerifierFunc) TxDecode(txBytes []byte) (sdk.Tx, error) { return v.decode(txBytes) }

func (v proposalVerifierFunc) TxEncode(tx sdk.Tx) ([]byte, error) { return v.encode(tx) }

func (v proposalVerifierFunc) PrepareProposalVerifyTx(tx sdk.Tx) ([]byte, error) {
	return v.encode(tx)
}

func (v proposalVerifierFunc) ProcessProposalVerifyTx(txBytes []byte) (sdk.Tx, error) {
	return v.decode(txBytes)
}

func TestProposalHandlerProcessRejectsInvalidAndOverBudget(t *testing.T) {
	t.Parallel()

	limits := ProposalLimits{MaxTxCount: 3, MaxCandidateTxCount: 4, MaxTxBytes: 32, MaxProposalBytes: 64}
	handler := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{active: true},
		limits,
		testSystemCodec(),
	)

	tests := []struct {
		name   string
		txs    [][]byte
		maxGas int64
	}{
		{name: "malformed", txs: [][]byte{[]byte("malformed")}},
		{name: "duplicate", txs: [][]byte{[]byte("one"), []byte("one")}},
		{name: "ante invalid", txs: [][]byte{[]byte("ante-invalid")}},
		{name: "incorrect system position", txs: [][]byte{[]byte("one"), []byte("VE84A_SYSTEM:v1")}},
		{name: "gas over budget", txs: [][]byte{[]byte("gas-10")}, maxGas: 9},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := sdk.Context{}
			if tc.maxGas > 0 {
				ctx = ctx.WithConsensusParams(cmtproto.ConsensusParams{Block: &cmtproto.BlockParams{MaxGas: tc.maxGas, MaxBytes: 64}})
			}
			response, err := handler.ProcessProposal(ctx, &abci.RequestProcessProposal{Height: 101, Txs: tc.txs})
			require.NoError(t, err)
			require.Equal(t, abci.ResponseProcessProposal_REJECT, response.Status)
		})
	}
}

func TestProposalHandlerProcessRepeatedAndPermutedOutcomes(t *testing.T) {
	t.Parallel()

	handler := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{active: true},
		DefaultProposalLimits(),
		testSystemCodec(),
	)
	permutations := [][][]byte{
		{[]byte("system"), []byte("one"), []byte("two"), []byte("three")},
		{[]byte("system"), []byte("three"), []byte("one"), []byte("two")},
		{[]byte("system"), []byte("two"), []byte("three"), []byte("one")},
	}

	for _, txs := range permutations {
		for range 10 {
			response, err := handler.ProcessProposal(sdk.Context{}, &abci.RequestProcessProposal{Height: 101, Txs: txs})
			require.NoError(t, err)
			require.Equal(t, abci.ResponseProcessProposal_ACCEPT, response.Status)
		}
	}
}

func TestProposalHandlerUpgradeCompatibilityAndFailClosedActivation(t *testing.T) {
	t.Parallel()

	legacy := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{active: false},
		DefaultProposalLimits(),
		testSystemCodec(),
	)
	legacyReq := &abci.RequestPrepareProposal{Height: 100, Txs: [][]byte{[]byte("malformed")}}
	prepared, err := legacy.PrepareProposal(sdk.Context{}, legacyReq)
	require.NoError(t, err)
	require.Equal(t, legacyReq.Txs, prepared.Txs)
	processed, err := legacy.ProcessProposal(sdk.Context{}, &abci.RequestProcessProposal{Height: 100, Txs: legacyReq.Txs})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseProcessProposal_ACCEPT, processed.Status)

	failing := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{err: errors.New("upgrade store unavailable")},
		DefaultProposalLimits(),
		testSystemCodec(),
	)
	processed, err = failing.ProcessProposal(sdk.Context{}, &abci.RequestProcessProposal{Height: 101, Txs: [][]byte{[]byte("one")}})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseProcessProposal_REJECT, processed.Status)

	prepared, err = failing.PrepareProposal(sdk.Context{}, &abci.RequestPrepareProposal{Height: 101, Txs: [][]byte{[]byte("one")}})
	require.NoError(t, err)
	require.Empty(t, prepared.Txs)
}

func TestProposalActivationDoneHeightBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		doneHeight     int64
		proposalHeight int64
		active         bool
		wantErr        bool
	}{
		{name: "new network marker absent", doneHeight: 0, proposalHeight: 1},
		{name: "upgrade block compatibility", doneHeight: 100, proposalHeight: 100},
		{name: "first strict block", doneHeight: 100, proposalHeight: 101, active: true},
		{name: "later strict block", doneHeight: 100, proposalHeight: 500, active: true},
		{name: "negative marker", doneHeight: -1, proposalHeight: 101, wantErr: true},
		{name: "future marker", doneHeight: 102, proposalHeight: 101, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			active, err := proposalActiveAfterDoneHeight(tc.doneHeight, tc.proposalHeight)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.active, active)
		})
	}
}

func TestProposalHandlerNilRequestsAndSDKResponsesFailClosed(t *testing.T) {
	t.Parallel()

	handler := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{active: true},
		DefaultProposalLimits(),
		testSystemCodec(),
	)
	_, err := handler.PrepareProposal(sdk.Context{}, nil)
	require.Error(t, err)
	processed, err := handler.ProcessProposal(sdk.Context{}, nil)
	require.NoError(t, err)
	require.Equal(t, abci.ResponseProcessProposal_REJECT, processed.Status)

	nilSDKResponse := NewProposalHandler(
		testProposalVerifier{},
		func(sdk.Context, *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			return nil, nil
		},
		staticProposalActivation{active: true},
		DefaultProposalLimits(),
		testSystemCodec(),
	)
	prepared, err := nilSDKResponse.PrepareProposal(sdk.Context{}, &abci.RequestPrepareProposal{Height: 101})
	require.NoError(t, err)
	require.Empty(t, prepared.Txs)
}

func TestProposalHandlerRejectsSDKInjectedTransactions(t *testing.T) {
	t.Parallel()

	handler := NewProposalHandler(
		testProposalVerifier{},
		func(_ sdk.Context, _ *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
			return &abci.ResponsePrepareProposal{Txs: [][]byte{[]byte("injected")}}, nil
		},
		staticProposalActivation{active: true},
		DefaultProposalLimits(),
		testSystemCodec(),
	)

	prepared, err := handler.PrepareProposal(sdk.Context{}, &abci.RequestPrepareProposal{
		Height:     101,
		MaxTxBytes: 64,
		Txs:        [][]byte{[]byte("candidate")},
	})
	require.NoError(t, err)
	require.Empty(t, prepared.Txs)
}

func TestProposalHandlerContainsSDKPrepareFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare sdk.PrepareProposalHandler
	}{
		{
			name: "error",
			prepare: func(sdk.Context, *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
				return nil, errors.New("prepare failed")
			},
		},
		{
			name: "panic",
			prepare: func(sdk.Context, *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
				panic("prepare failed")
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewProposalHandler(
				testProposalVerifier{},
				tc.prepare,
				staticProposalActivation{active: true},
				DefaultProposalLimits(),
				testSystemCodec(),
			)
			prepared, err := handler.PrepareProposal(sdk.Context{}, &abci.RequestPrepareProposal{
				Height:     101,
				MaxTxBytes: 64,
				Txs:        [][]byte{[]byte("candidate")},
			})
			require.NoError(t, err)
			require.Empty(t, prepared.Txs)
		})
	}
}

func TestFailClosedPrepareProposalNeverRestoresCandidates(t *testing.T) {
	t.Parallel()

	candidates := [][]byte{[]byte("unvalidated")}
	tests := []struct {
		name    string
		handler sdk.PrepareProposalHandler
	}{
		{
			name: "error",
			handler: func(sdk.Context, *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
				return nil, errors.New("activation store unavailable")
			},
		},
		{
			name: "nil response",
			handler: func(sdk.Context, *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
				return nil, nil
			},
		},
		{
			name: "panic",
			handler: func(sdk.Context, *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
				panic("unexpected proposal panic")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := FailClosedPrepareProposal(tc.handler)(sdk.Context{}, &abci.RequestPrepareProposal{Txs: candidates})
			require.NoError(t, err)
			require.NotNil(t, response)
			require.Empty(t, response.Txs)
			require.Equal(t, [][]byte{[]byte("unvalidated")}, candidates)
		})
	}
}

func TestProposalHandlerFourProcessDeterminism(t *testing.T) {
	t.Parallel()

	zones := []string{"UTC", "Pacific/Auckland", "America/Los_Angeles", "Asia/Kolkata"}
	outputs := make([][]byte, len(zones))
	for i, zone := range zones {
		outputPath := filepath.Join(t.TempDir(), "decision.json")
		//nolint:gosec // G204: fixed test-binary path and fixed test selector; no user-controlled command or argument.
		command := exec.Command(os.Args[0], "-test.run=^TestProposalHandlerProcessBoundaryHelper$")
		command.Env = append(os.Environ(),
			"VE_TEST_PROPOSAL_HELPER=1",
			"VE_TEST_PROPOSAL_OUTPUT="+outputPath,
			"TZ="+zone,
		)
		combined, err := command.CombinedOutput()
		require.NoError(t, err, "process %d (%s) failed: %s", i, zone, combined)
		outputs[i], err = os.ReadFile(outputPath)
		require.NoError(t, err)
	}

	for i := 1; i < len(outputs); i++ {
		require.Equal(t, outputs[0], outputs[i], "process %d diverged", i)
	}
}

func TestProposalHandlerProcessBoundaryHelper(t *testing.T) {
	if os.Getenv("VE_TEST_PROPOSAL_HELPER") != "1" {
		t.Skip("helper subprocess only")
	}

	handler := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{active: true},
		DefaultProposalLimits(),
		testSystemCodec(),
	)
	request := &abci.RequestPrepareProposal{
		Height:     101,
		MaxTxBytes: 1024,
		Txs: [][]byte{
			[]byte("ordinary-a"),
			[]byte("malformed"),
			[]byte("ordinary-b"),
			[]byte("ordinary-a"),
		},
	}
	prepared, err := handler.PrepareProposal(sdk.Context{}, request)
	require.NoError(t, err)
	processed, err := handler.ProcessProposal(sdk.Context{}, &abci.RequestProcessProposal{Height: 101, Txs: prepared.Txs})
	require.NoError(t, err)

	decision, err := json.Marshal(struct {
		Txs    [][]byte                                    `json:"txs"`
		Status abci.ResponseProcessProposal_ProposalStatus `json:"status"`
	}{Txs: prepared.Txs, Status: processed.Status})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(os.Getenv("VE_TEST_PROPOSAL_OUTPUT"), decision, 0o600))
}

func FuzzCanonicalProposalValidator(f *testing.F) {
	validator := NewCanonicalProposalValidator(testProposalVerifier{}, DefaultProposalLimits())
	f.Add([]byte("ordinary"), []byte("second"))
	f.Add([]byte("malformed"), []byte("ordinary"))
	f.Fuzz(func(t *testing.T, first, second []byte) {
		txs := [][]byte{bytes.Clone(first), bytes.Clone(second)}
		_, firstErr := validator.Validate(txs, 0)
		_, secondErr := validator.Validate(txs, 0)
		require.Equal(t, ProposalErrorCode(firstErr), ProposalErrorCode(secondErr))
	})
}

func BenchmarkProposalHandlerProcessProposal(b *testing.B) {
	handler := NewProposalHandler(
		testProposalVerifier{},
		passthroughPrepareProposal,
		staticProposalActivation{active: true},
		DefaultProposalLimits(),
		testSystemCodec(),
	)
	txs := make([][]byte, 501)
	txs[0] = []byte("system")
	for i := 1; i < len(txs); i++ {
		txs[i] = []byte{byte(i >> 8), byte(i), 't', 'x'}
	}
	req := &abci.RequestProcessProposal{Height: 101, Txs: txs}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		response, err := handler.ProcessProposal(sdk.Context{}, req)
		if err != nil || response.Status != abci.ResponseProcessProposal_ACCEPT {
			b.Fatalf("unexpected response: response=%v err=%v", response, err)
		}
	}
}
