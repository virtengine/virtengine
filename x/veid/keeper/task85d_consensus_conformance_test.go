package keeper

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	coreheader "cosmossdk.io/core/header"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtcrypto "github.com/cometbft/cometbft/crypto/ed25519"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"github.com/virtengine/virtengine/x/veid/types"
)

const (
	task85DConsensusHelperProcessEnv = "VEID_TASK85D_CONSENSUS_HELPER_PROCESS"
	task85DConsensusHelperProfile    = "task85d-consensus-child-receipt-signer"
	task85DConsensusHelperMaxBytes   = 64 * 1024
	task85DConsensusHelperTimeout    = 10 * time.Second
)

type task85DConsensusHelperRequest struct {
	Action   string                         `json:"action"`
	Profile  string                         `json:"profile"`
	Sequence uint64                         `json:"sequence"`
	Receipt  task85DConsensusReceiptRequest `json:"receipt,omitempty"`
}

type task85DConsensusReceiptRequest struct {
	ChainID               string   `json:"chain_id"`
	AccountAddress        string   `json:"account_address"`
	RequestID             string   `json:"request_id"`
	ScopeIDs              []string `json:"scope_ids"`
	Nonce                 string   `json:"nonce"`
	InputDigest           string   `json:"input_digest"`
	FeatureDigest         string   `json:"feature_digest"`
	SchemaDigest          string   `json:"schema_digest"`
	EvidenceLineageDigest string   `json:"evidence_lineage_digest"`
	PipelineVersion       string   `json:"pipeline_version"`
	ModelManifestDigest   string   `json:"model_manifest_digest"`
	ModelDigest           string   `json:"model_digest"`
	RuntimeImageDigest    string   `json:"runtime_image_digest"`
	RuntimeDigest         string   `json:"runtime_digest"`
	ConfigDigest          string   `json:"config_digest"`
	Score                 uint32   `json:"score"`
	Status                string   `json:"status"`
	ConfidenceMillionths  uint32   `json:"confidence_millionths"`
	ReasonCodes           []string `json:"reason_codes"`
	IssuedHeight          int64    `json:"issued_height"`
	IssuedUnix            int64    `json:"issued_unix"`
	ExpiresHeight         int64    `json:"expires_height"`
	ExpiresUnix           int64    `json:"expires_unix"`
}

type task85DConsensusSignerRegistration struct {
	SignerID    string `json:"signer_id"`
	KeyID       string `json:"key_id"`
	Fingerprint string `json:"fingerprint"`
	Sequence    uint64 `json:"sequence"`
	PublicKey   []byte `json:"public_key"`
}

type task85DConsensusReceiptResponse struct {
	ReceiptBytes  []byte `json:"receipt_bytes"`
	ReceiptDigest string `json:"receipt_digest"`
}

func TestTask85DConsensusReceiptHelperProcess(t *testing.T) {
	if os.Getenv(task85DConsensusHelperProcessEnv) != "1" {
		return
	}
	req, err := task85DConsensusDecodeHelperRequest(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode helper request: %v\n", err)
		os.Exit(2)
	}
	var resp any
	switch req.Action {
	case "register":
		resp = task85DConsensusRegisterHelper(req)
	case "issue":
		resp, err = task85DConsensusIssueHelper(req)
	default:
		err = fmt.Errorf("unsupported consensus helper action %q", req.Action)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(3)
	}
	if err := task85DConsensusEncodeHelperResponse(os.Stdout, resp); err != nil {
		fmt.Fprintf(os.Stderr, "encode helper response: %v\n", err)
		os.Exit(4)
	}
	os.Exit(0)
}

func TestTask85DConsensusHelperProtocolRejectsMalformedRequests(t *testing.T) {
	valid := `{"action":"register","profile":"task85d-consensus-child-receipt-signer","sequence":1}`
	tests := []struct {
		name    string
		input   []byte
		wantErr string
	}{
		{name: "unknown-field", input: []byte(`{"action":"register","profile":"task85d-consensus-child-receipt-signer","unknown":true}`), wantErr: "unknown field"},
		{name: "trailing-json", input: []byte(valid + ` {}`), wantErr: "trailing JSON"},
		{name: "oversized-input", input: bytes.Repeat([]byte("x"), task85DConsensusHelperMaxBytes+1), wantErr: "exceeds"},
		{name: "unsupported-action", input: []byte(`{"action":"destroy","profile":"task85d-consensus-child-receipt-signer"}`), wantErr: "unsupported"},
		{name: "unsupported-profile", input: []byte(`{"action":"register","profile":"bad-profile","sequence":1}`), wantErr: "unsupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr, err := task85DConsensusRunHelperRaw(t, tt.input)
			require.Error(t, err)
			require.Contains(t, stderr, tt.wantErr)
		})
	}
}

func TestTask85DFourValidatorConsensusConformanceReceiptBytes(t *testing.T) {
	reg := task85DConsensusRegister(t, 1)
	nodes := make([]task85DConsensusNode, 4)
	for i := range nodes {
		nodes[i] = setupTask85DConsensusNode(t, reg)
	}

	expected := nodes[0].expected
	quorumReceiptBytes := task85DConsensusIssueReceipt(t, task85DConsensusIssueRequest(t, nodes[0], reg, "task85d-quorum-receipt", 91))
	dissentReceiptBytes := task85DConsensusIssueReceipt(t, task85DConsensusIssueRequest(t, nodes[0], reg, "task85d-valid-dissent-receipt", 88))
	quorumReceipt, err := types.DecodeCanonicalSignedInferenceReceipt(quorumReceiptBytes)
	require.NoError(t, err)
	dissentReceipt, err := types.DecodeCanonicalSignedInferenceReceipt(dissentReceiptBytes)
	require.NoError(t, err)
	quorumResult := task85DConsensusCarrierFromReceipt(t, quorumReceipt, quorumReceiptBytes)
	dissentResult := task85DConsensusCarrierFromReceipt(t, dissentReceipt, dissentReceiptBytes)
	require.NotEqual(t, hex.EncodeToString(quorumResult.ReceiptDigest), hex.EncodeToString(dissentResult.ReceiptDigest))

	quorumBundles := make([]*veidv1.VEIDVoteExtension, 3)
	for i := 1; i <= 3; i++ {
		quorumBundles[i-1] = task85DConsensusStageExtendAndVerify(t, nodes[i], quorumReceiptBytes, expected)
	}
	dissentBundle := task85DConsensusStageExtendAndVerify(t, nodes[0], dissentReceiptBytes, expected)

	validatorSpecs := []consensusVoteSpec{
		{key: cmtcrypto.GenPrivKey(), power: 40, bundle: quorumBundles[0]},
		{key: cmtcrypto.GenPrivKey(), power: 30, bundle: quorumBundles[1]},
		{key: cmtcrypto.GenPrivKey(), power: 20, bundle: quorumBundles[2]},
		{key: cmtcrypto.GenPrivKey(), power: 10, bundle: dissentBundle},
	}
	staking := installConsensusValidators(t, validatorSpecs)
	commit := signedExtendedCommit(t, nodes[0].ctx, 2, validatorSpecs)
	aggregate, err := AggregateVoteExtensions(commit, expected)
	require.NoError(t, err)
	require.Equal(t, int64(100), aggregate.TotalVotingPower)
	require.Equal(t, int64(67), aggregate.QuorumVotingPower)
	require.Len(t, aggregate.Results, 1)
	require.Equal(t, int64(90), aggregate.Results[0].VotingPower)
	require.Equal(t, quorumResult.ReceiptDigest, aggregate.Results[0].Result.ReceiptDigest)
	require.NotEqual(t, dissentResult.ReceiptDigest, aggregate.Results[0].Result.ReceiptDigest)

	commitBytes, err := proto.Marshal(&commit)
	require.NoError(t, err)
	msg := &veidv1.MsgSubmitConsensusVerification{
		Version:        VoteExtensionVersion,
		ChainId:        nodes[0].ctx.ChainID(),
		Height:         nodes[0].ctx.BlockHeight(),
		ExtendedCommit: commitBytes,
		Aggregate:      aggregate,
	}

	var expectedDigest string
	for i, node := range nodes {
		finalizer := NewKeeper(node.keeper.cdc, node.keeper.skey, node.keeper.authority)
		finalizer.SetConsensusSystemTxAuthorizer(func(callCtx sdk.Context) bool {
			return string(callCtx.TxBytes()) == "task85d-authorized-system-tx"
		})
		finalizer.SetStakingKeeper(staking)
		finalizer.SetConsensusValidatorStore(staking)
		require.Empty(t, finalizer.GetBlockVerificationResults(node.ctx, expected.Height))
		ctx := node.ctx.WithCometInfo(finalCometInfo{commit: commit})
		server := NewMsgServerImpl(finalizer)
		response, err := server.SubmitConsensusVerification(ctx, msg)
		require.NoError(t, err)
		require.Equal(t, uint32(1), response.AppliedResults)

		storedResult, found := finalizer.GetVerificationResult(ctx, node.request.RequestID)
		require.True(t, found)
		require.Equal(t, types.VerificationResultStatusSuccess, storedResult.Status)
		require.Equal(t, uint32(91), storedResult.Score)
		require.Equal(t, hex.EncodeToString(quorumResult.ReceiptDigest), storedResult.Metadata[types.VerificationResultMetadataReceiptDigest])
		require.NotEqual(t, hex.EncodeToString(dissentResult.ReceiptDigest), storedResult.Metadata[types.VerificationResultMetadataReceiptDigest])

		score, status, found := finalizer.GetScore(ctx, node.account.String())
		require.True(t, found)
		require.Equal(t, uint32(91), score)
		require.Equal(t, types.AccountStatusVerified, status)

		stateDigest := task85DConsensusStoreDigest(t, ctx, finalizer)
		if i == 0 {
			expectedDigest = stateDigest
			continue
		}
		require.Equal(t, expectedDigest, stateDigest)
	}
}

type task85DConsensusNode struct {
	keeper   Keeper
	ctx      sdk.Context
	account  sdk.AccAddress
	request  *types.VerificationRequest
	expected VoteExtensionExpectations
}

func setupTask85DConsensusNode(t *testing.T, reg task85DConsensusSignerRegistration) task85DConsensusNode {
	t.Helper()
	keeper, ctx, stateStore := setupInferenceReceiptKeeper(t)
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	require.NoError(t, keeper.SetParams(ctx, params))
	registerActiveInferencePipeline(t, keeper, ctx)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).
		WithBlockTime(ctx.BlockTime().Add(time.Second)).
		WithExecMode(sdk.ExecModeFinalize).
		WithTxBytes([]byte("task85d-authorized-system-tx")).
		WithHeaderInfo(coreheader.Info{ChainID: ctx.ChainID(), Height: ctx.BlockHeight() + 1, Time: ctx.BlockTime().Add(time.Second)}).
		WithConsensusParams(cmtproto.ConsensusParams{Abci: &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}})
	keeper.SetConsensusSystemTxAuthorizer(func(callCtx sdk.Context) bool {
		return string(callCtx.TxBytes()) == "task85d-authorized-system-tx"
	})

	account := sdk.AccAddress(testHash(0x44)[:20])
	_, err := keeper.CreateIdentityRecord(ctx, account)
	require.NoError(t, err)
	task85DConsensusRegisterSigner(t, keeper, ctx, reg)
	request := types.NewVerificationRequest("request-task85d-consensus", account.String(), []string{"scope-a"}, ctx.BlockTime(), ctx.BlockHeight()-1)
	backfillRequestInferenceProfileForTest(t, keeper, ctx, request)
	require.NoError(t, keeper.setVerificationRequest(ctx, request))
	keeper.addToPendingQueue(ctx, request)

	expected, err := keeper.VoteExtensionCommitments(ctx)
	require.NoError(t, err)
	expected.Height = ctx.BlockHeight() - 1
	expected.BlockHash = []byte("task85d-block-hash")

	return task85DConsensusNode{
		keeper:   keeper,
		ctx:      ctx,
		account:  account,
		request:  request,
		expected: expected,
	}
}

func task85DConsensusRegisterSigner(t *testing.T, keeper Keeper, ctx sdk.Context, reg task85DConsensusSignerRegistration) {
	t.Helper()
	now := ctx.BlockTime().UTC()
	key := types.NewSignerKeyInfo(reg.SignerID, reg.PublicKey, types.ProofTypeEd25519, reg.Sequence, now.Add(-time.Minute))
	key.KeyID = reg.KeyID
	key.Fingerprint = reg.Fingerprint
	key.State = types.SignerKeyStateActive
	activatedAt := now.Add(-time.Minute)
	expiresAt := now.Add(24 * time.Hour)
	key.ActivatedAt = &activatedAt
	key.ExpiresAt = &expiresAt
	key.Metadata[types.SignerKeyMetadataEvidenceTypes] = string(types.AttestationTypeInferenceReceipt)
	require.NoError(t, keeper.RegisterSignerKey(ctx, keeper.GetAuthority(), key))
}

func task85DConsensusIssueRequest(t *testing.T, node task85DConsensusNode, reg task85DConsensusSignerRegistration, nonce string, score uint32) task85DConsensusHelperRequest {
	t.Helper()
	snapshot, err := node.request.RequireInferenceProfileSnapshot()
	require.NoError(t, err)
	featureDigest := testHash(0x34)
	lineageDigest := testHash(byte(score))
	return task85DConsensusHelperRequest{
		Action:   "issue",
		Profile:  task85DConsensusHelperProfile,
		Sequence: reg.Sequence,
		Receipt: task85DConsensusReceiptRequest{
			ChainID:               node.ctx.ChainID(),
			AccountAddress:        node.request.AccountAddress,
			RequestID:             node.request.RequestID,
			ScopeIDs:              types.CanonicalInferenceReceiptScopeIDs(node.request.ScopeIDs),
			Nonce:                 nonce,
			InputDigest:           hex.EncodeToString(testHash(0x33)),
			FeatureDigest:         hex.EncodeToString(featureDigest),
			SchemaDigest:          hex.EncodeToString(snapshot.FeatureSchemaDigest),
			EvidenceLineageDigest: hex.EncodeToString(lineageDigest),
			PipelineVersion:       snapshot.PipelineVersion,
			ModelManifestDigest:   hex.EncodeToString(snapshot.ModelManifestDigest),
			ModelDigest:           hex.EncodeToString(snapshot.ModelDigest),
			RuntimeImageDigest:    hex.EncodeToString(snapshot.RuntimeImageDigest),
			RuntimeDigest:         hex.EncodeToString(snapshot.RuntimeDigest),
			ConfigDigest:          hex.EncodeToString(snapshot.DeterminismConfigDigest),
			Score:                 score,
			Status:                string(types.VerificationResultStatusSuccess),
			ConfidenceMillionths:  910000,
			ReasonCodes:           []string{string(types.ReasonCodeSuccess)},
			IssuedHeight:          node.expected.Height,
			IssuedUnix:            node.ctx.BlockTime().Unix(),
			ExpiresHeight:         node.expected.Height + 2,
			ExpiresUnix:           node.ctx.BlockTime().Add(2 * time.Minute).Unix(),
		},
	}
}

func task85DConsensusStageExtendAndVerify(t *testing.T, node task85DConsensusNode, receiptBytes []byte, expected VoteExtensionExpectations) *veidv1.VEIDVoteExtension {
	t.Helper()
	task85DConsensusStageAndExtend(t, node, receiptBytes, expected)
	resp, err := node.keeper.ExtendVote(node.ctx.WithExecMode(sdk.ExecModeVoteExtension), &abci.RequestExtendVote{
		Height: expected.Height,
		Hash:   expected.BlockHash,
	}, nil)
	require.NoError(t, err)
	verify, err := node.keeper.VerifyVoteExtension(node.ctx.WithExecMode(sdk.ExecModeVerifyVoteExtension), &abci.RequestVerifyVoteExtension{
		Height:           expected.Height,
		Hash:             expected.BlockHash,
		ValidatorAddress: cmtcrypto.GenPrivKey().PubKey().Address(),
		VoteExtension:    resp.VoteExtension,
	}, nil)
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_ACCEPT, verify.Status)
	bundle, err := UnmarshalVoteExtensionBundle(resp.VoteExtension)
	require.NoError(t, err)
	return bundle
}

func task85DConsensusStageAndExtend(t *testing.T, node task85DConsensusNode, receiptBytes []byte, expected VoteExtensionExpectations) veidv1.VEIDVoteExtensionResult {
	t.Helper()
	receipt, err := types.DecodeCanonicalSignedInferenceReceipt(receiptBytes)
	require.NoError(t, err)
	key, err := node.keeper.resolveSignerKey(node.ctx, receipt.SignerKeyID, receipt.SignerFingerprint)
	require.NoError(t, err)
	require.NoError(t, receipt.VerifySignature(ed25519.PublicKey(key.PublicKey)))
	carrier := task85DConsensusCarrierFromReceipt(t, receipt, receiptBytes)
	replay, err := node.keeper.validateCarriedConsensusReceipt(node.ctx, carrier, node.request, expected.Height)
	require.NoError(t, err)
	require.False(t, replay.ExactReplay)
	result := task85DConsensusVerificationResult(t, receipt, replay, expected.Height)
	require.NoError(t, node.keeper.storeVerifiedBlockVerificationResult(node.ctx.WithExecMode(sdk.ExecModeVoteExtension), expected.Height, result, receiptBytes))
	return carrier
}

func task85DConsensusCarrierFromReceipt(t *testing.T, receipt types.InferenceReceipt, receiptBytes []byte) veidv1.VEIDVoteExtensionResult {
	t.Helper()
	receiptDigest, err := receipt.Digest()
	require.NoError(t, err)
	reasonCodes := make([]string, len(receipt.ReasonCodes))
	for i, reason := range receipt.ReasonCodes {
		reasonCodes[i] = string(reason)
	}
	carrier := veidv1.VEIDVoteExtensionResult{
		RequestId:      receipt.RequestID,
		AccountAddress: receipt.AccountAddress,
		Score:          receipt.Score,
		Status:         string(receipt.Status),
		ModelVersion:   receipt.PipelineVersion,
		InputHash:      bytes.Clone(receipt.InputDigest),
		ReasonCodes:    reasonCodes,
		ReceiptDigest:  receiptDigest,
		ReceiptBytes:   bytes.Clone(receiptBytes),
	}
	carrier.ResultHash = ComputeVoteExtensionResultHash(carrier)
	return carrier
}

func task85DConsensusVerificationResult(t *testing.T, receipt types.InferenceReceipt, replay inferenceReceiptReplayCheck, height int64) types.VerificationResult {
	t.Helper()
	result := types.NewVerificationResult(receipt.RequestID, receipt.AccountAddress, receipt.IssuedAt, height)
	result.Score = receipt.Score
	result.Status = receipt.Status
	result.ModelVersion = receipt.PipelineVersion
	result.InputHash = bytes.Clone(receipt.InputDigest)
	result.ReasonCodes = append([]types.ReasonCode(nil), receipt.ReasonCodes...)
	receiptDigest, err := receipt.Digest()
	require.NoError(t, err)
	result.Metadata[types.VerificationResultMetadataReceiptDigest] = hex.EncodeToString(receiptDigest)
	result.Metadata["receipt_context_digest"] = replay.ContextDigest
	result.Metadata["receipt_nonce_digest"] = replay.NonceDigest
	result.Metadata["runtime_digest"] = hex.EncodeToString(receipt.RuntimeDigest)
	result.Metadata["model_digest"] = hex.EncodeToString(receipt.ModelDigest)
	return *result
}

func task85DConsensusDecodeHelperRequest(r io.Reader) (task85DConsensusHelperRequest, error) {
	input, err := io.ReadAll(io.LimitReader(r, task85DConsensusHelperMaxBytes+1))
	if err != nil {
		return task85DConsensusHelperRequest{}, err
	}
	if len(input) > task85DConsensusHelperMaxBytes {
		return task85DConsensusHelperRequest{}, fmt.Errorf("helper request exceeds %d bytes", task85DConsensusHelperMaxBytes)
	}
	dec := json.NewDecoder(bytes.NewReader(input))
	dec.DisallowUnknownFields()
	var req task85DConsensusHelperRequest
	if err := dec.Decode(&req); err != nil {
		return task85DConsensusHelperRequest{}, err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return task85DConsensusHelperRequest{}, fmt.Errorf("trailing JSON after helper request")
		}
		return task85DConsensusHelperRequest{}, fmt.Errorf("trailing JSON after helper request: %w", err)
	}
	if req.Action == "" {
		return task85DConsensusHelperRequest{}, fmt.Errorf("action is required")
	}
	if req.Profile == "" {
		return task85DConsensusHelperRequest{}, fmt.Errorf("profile is required")
	}
	if req.Profile != task85DConsensusHelperProfile {
		return task85DConsensusHelperRequest{}, fmt.Errorf("unsupported consensus helper profile %q", req.Profile)
	}
	if req.Sequence == 0 {
		req.Sequence = 1
	}
	return req, nil
}

func task85DConsensusEncodeHelperResponse(w io.Writer, resp any) error {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(resp); err != nil {
		return err
	}
	if out.Len() > task85DConsensusHelperMaxBytes {
		return fmt.Errorf("helper response exceeds %d bytes", task85DConsensusHelperMaxBytes)
	}
	_, err := w.Write(out.Bytes())
	return err
}

func task85DConsensusRegisterHelper(req task85DConsensusHelperRequest) task85DConsensusSignerRegistration {
	pub, _ := task85DConsensusPrivateKey(req.Profile, req.Sequence)
	return task85DConsensusSignerRegistration{
		SignerID:    "did:virtengine:inference:" + req.Profile,
		KeyID:       "did:virtengine:inference:" + req.Profile + ":" + strconv.FormatUint(req.Sequence, 10),
		Fingerprint: types.ComputeKeyFingerprint(pub),
		Sequence:    req.Sequence,
		PublicKey:   append([]byte(nil), pub...),
	}
}

func task85DConsensusIssueHelper(req task85DConsensusHelperRequest) (task85DConsensusReceiptResponse, error) {
	reg := task85DConsensusRegisterHelper(req)
	receiptReq := req.Receipt
	if receiptReq.ChainID == "" || receiptReq.AccountAddress == "" || receiptReq.RequestID == "" || receiptReq.Nonce == "" || receiptReq.PipelineVersion == "" {
		return task85DConsensusReceiptResponse{}, fmt.Errorf("receipt chain_id, account_address, request_id, nonce, and pipeline_version are required")
	}
	pub, priv := task85DConsensusPrivateKey(req.Profile, req.Sequence)
	if !bytes.Equal(pub, reg.PublicKey) {
		return task85DConsensusReceiptResponse{}, fmt.Errorf("consensus helper key mismatch")
	}
	digests, err := task85DConsensusReceiptDigests(receiptReq)
	if err != nil {
		return task85DConsensusReceiptResponse{}, err
	}
	receipt := types.InferenceReceipt{
		Domain:                types.InferenceReceiptDomain,
		Version:               types.InferenceReceiptVersion,
		ChainID:               receiptReq.ChainID,
		AccountAddress:        receiptReq.AccountAddress,
		RequestID:             receiptReq.RequestID,
		ScopeIDs:              types.CanonicalInferenceReceiptScopeIDs(receiptReq.ScopeIDs),
		Nonce:                 receiptReq.Nonce,
		InputDigest:           digests.input,
		FeatureDigest:         digests.feature,
		SchemaDigest:          digests.schema,
		EvidenceLineageDigest: digests.lineage,
		PipelineVersion:       receiptReq.PipelineVersion,
		ModelManifestDigest:   digests.modelManifest,
		ModelDigest:           digests.model,
		RuntimeImageDigest:    digests.runtimeImage,
		RuntimeDigest:         digests.runtime,
		ConfigDigest:          digests.config,
		DeterminismProfile:    types.CanonicalInferenceDeterminismProfile(),
		Score:                 receiptReq.Score,
		Status:                types.VerificationResultStatus(receiptReq.Status),
		ConfidenceMillionths:  receiptReq.ConfidenceMillionths,
		IssuedHeight:          receiptReq.IssuedHeight,
		IssuedAt:              time.Unix(receiptReq.IssuedUnix, 0).UTC(),
		ExpiresHeight:         receiptReq.ExpiresHeight,
		ExpiresAt:             time.Unix(receiptReq.ExpiresUnix, 0).UTC(),
		SignerKeyID:           reg.KeyID,
		SignerFingerprint:     reg.Fingerprint,
		SignerSequence:        reg.Sequence,
	}
	for _, reason := range receiptReq.ReasonCodes {
		receipt.ReasonCodes = append(receipt.ReasonCodes, types.ReasonCode(reason))
	}
	if err := receipt.Sign(priv); err != nil {
		return task85DConsensusReceiptResponse{}, err
	}
	receiptBytes, err := receipt.CanonicalSignedBytes()
	if err != nil {
		return task85DConsensusReceiptResponse{}, err
	}
	digest, err := receipt.Digest()
	if err != nil {
		return task85DConsensusReceiptResponse{}, err
	}
	return task85DConsensusReceiptResponse{ReceiptBytes: receiptBytes, ReceiptDigest: hex.EncodeToString(digest)}, nil
}

type task85DConsensusReceiptDigestSet struct {
	input         []byte
	feature       []byte
	schema        []byte
	lineage       []byte
	modelManifest []byte
	model         []byte
	runtimeImage  []byte
	runtime       []byte
	config        []byte
}

func task85DConsensusReceiptDigestsFromHex(value string, name string) ([]byte, error) {
	bz, err := hex.DecodeString(value)
	if err != nil || len(bz) != sha256.Size {
		return nil, fmt.Errorf("invalid %s digest", name)
	}
	return bz, nil
}

func task85DConsensusReceiptDigests(req task85DConsensusReceiptRequest) (task85DConsensusReceiptDigestSet, error) {
	var out task85DConsensusReceiptDigestSet
	var err error
	if out.input, err = task85DConsensusReceiptDigestsFromHex(req.InputDigest, "input"); err != nil {
		return out, err
	}
	if out.feature, err = task85DConsensusReceiptDigestsFromHex(req.FeatureDigest, "feature"); err != nil {
		return out, err
	}
	if out.schema, err = task85DConsensusReceiptDigestsFromHex(req.SchemaDigest, "schema"); err != nil {
		return out, err
	}
	if out.lineage, err = task85DConsensusReceiptDigestsFromHex(req.EvidenceLineageDigest, "evidence_lineage"); err != nil {
		return out, err
	}
	if out.modelManifest, err = task85DConsensusReceiptDigestsFromHex(req.ModelManifestDigest, "model_manifest"); err != nil {
		return out, err
	}
	if out.model, err = task85DConsensusReceiptDigestsFromHex(req.ModelDigest, "model"); err != nil {
		return out, err
	}
	if out.runtimeImage, err = task85DConsensusReceiptDigestsFromHex(req.RuntimeImageDigest, "runtime_image"); err != nil {
		return out, err
	}
	if out.runtime, err = task85DConsensusReceiptDigestsFromHex(req.RuntimeDigest, "runtime"); err != nil {
		return out, err
	}
	if out.config, err = task85DConsensusReceiptDigestsFromHex(req.ConfigDigest, "config"); err != nil {
		return out, err
	}
	return out, nil
}

func task85DConsensusPrivateKey(profile string, sequence uint64) (ed25519.PublicKey, ed25519.PrivateKey) {
	sum := sha256.Sum256([]byte("task85d-consensus-receipt:" + profile + ":" + strconv.FormatUint(sequence, 10)))
	priv := ed25519.NewKeyFromSeed(sum[:])
	return priv.Public().(ed25519.PublicKey), priv
}

func task85DConsensusRegister(t *testing.T, sequence uint64) task85DConsensusSignerRegistration {
	t.Helper()
	req := task85DConsensusHelperRequest{Action: "register", Profile: task85DConsensusHelperProfile, Sequence: sequence}
	stdout := task85DConsensusRunHelper(t, req)
	var reg task85DConsensusSignerRegistration
	task85DConsensusDecodeHelperResponse(t, stdout, &reg)
	return reg
}

func task85DConsensusIssueReceipt(t *testing.T, req task85DConsensusHelperRequest) []byte {
	t.Helper()
	stdout := task85DConsensusRunHelper(t, req)
	var resp task85DConsensusReceiptResponse
	task85DConsensusDecodeHelperResponse(t, stdout, &resp)
	require.NotEmpty(t, resp.ReceiptBytes)
	decoded, err := types.DecodeCanonicalSignedInferenceReceipt(resp.ReceiptBytes)
	require.NoError(t, err)
	digest, err := decoded.Digest()
	require.NoError(t, err)
	require.Equal(t, resp.ReceiptDigest, hex.EncodeToString(digest))
	return resp.ReceiptBytes
}

func task85DConsensusRunHelper(t *testing.T, req task85DConsensusHelperRequest) string {
	t.Helper()
	reqBytes, err := json.Marshal(req)
	require.NoError(t, err)
	require.LessOrEqual(t, len(reqBytes), task85DConsensusHelperMaxBytes)
	stdout, stderr, err := task85DConsensusRunHelperRaw(t, append(reqBytes, '\n'))
	require.NoErrorf(t, err, "helper stderr: %s", stderr)
	return stdout
}

func task85DConsensusRunHelperRaw(t *testing.T, reqBytes []byte) (string, string, error) {
	t.Helper()
	require.LessOrEqual(t, len(reqBytes), task85DConsensusHelperMaxBytes+1)
	ctx, cancel := context.WithTimeout(context.Background(), task85DConsensusHelperTimeout)
	defer cancel()
	// #nosec G204 -- this intentionally re-execs the current Go test binary as a bounded helper process.
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestTask85DConsensusReceiptHelperProcess$")
	cmd.Env = append(os.Environ(), task85DConsensusHelperProcessEnv+"=1")
	cmd.Stdin = bytes.NewReader(reqBytes)
	dir := t.TempDir()
	stdoutPath := filepath.Join(dir, "stdout.json")
	stderrPath := filepath.Join(dir, "stderr.txt")
	stdoutFile, err := os.Create(stdoutPath)
	require.NoError(t, err)
	stderrFile, err := os.Create(stderrPath)
	require.NoError(t, err)
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	runErr := cmd.Run()
	require.NoError(t, stdoutFile.Close())
	require.NoError(t, stderrFile.Close())
	stdout := task85DConsensusReadBoundedFile(t, stdoutPath)
	stderr := task85DConsensusReadBoundedFile(t, stderrPath)
	if ctx.Err() != nil {
		return string(stdout), string(stderr), ctx.Err()
	}
	return string(stdout), string(stderr), runErr
}

func task85DConsensusReadBoundedFile(t *testing.T, path string) []byte {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, task85DConsensusHelperMaxBytes+1))
	require.NoError(t, err)
	require.LessOrEqual(t, len(data), task85DConsensusHelperMaxBytes)
	return data
}

func task85DConsensusDecodeHelperResponse[T any](t *testing.T, stdout string, target *T) {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	dec.DisallowUnknownFields()
	require.NoError(t, dec.Decode(target))
	require.ErrorIs(t, dec.Decode(&struct{}{}), io.EOF)
}

func task85DConsensusStoreDigest(t *testing.T, ctx sdk.Context, keeper Keeper) string {
	t.Helper()
	store := ctx.KVStore(keeper.skey)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	hash := sha256.New()
	for ; iterator.Valid(); iterator.Next() {
		hash.Write(iterator.Key())
		hash.Write([]byte{0})
		hash.Write(iterator.Value())
		hash.Write([]byte{0})
	}
	require.NoError(t, iterator.Error())
	return hex.EncodeToString(hash.Sum(nil))
}
