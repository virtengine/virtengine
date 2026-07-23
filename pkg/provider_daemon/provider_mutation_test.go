// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"encoding/hex"
	"errors"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/stretchr/testify/require"

	hpcv1 "github.com/virtengine/virtengine/sdk/go/node/hpc/v1"
	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
	marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"
	marketplacev1 "github.com/virtengine/virtengine/sdk/go/node/marketplace/v1"
	providerv1beta4 "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
	supportv1 "github.com/virtengine/virtengine/sdk/go/node/support/v1"
	depositv1 "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1"
)

type mutationChainFake struct {
	mu             sync.Mutex
	accountNumber  uint64
	sequence       uint64
	height         int64
	blockHash      string
	estimatedGas   uint64
	broadcastErrs  []error
	broadcasts     [][]byte
	confirmed      map[string]ProviderTxConfirmation
	reconciled     ProviderMutationReconciliation
	reconcileCalls int
	confirmMissing bool
	confirmMutator func(ProviderTxConfirmation) ProviderTxConfirmation
}

func newMutationChainFake() *mutationChainFake {
	return &mutationChainFake{
		accountNumber: 7,
		sequence:      11,
		height:        102,
		blockHash:     "AABBCC",
		estimatedGas:  100000,
		confirmed:     make(map[string]ProviderTxConfirmation),
	}
}

func (f *mutationChainFake) ResolveAccountSequence(context.Context, string) (uint64, uint64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.accountNumber, f.sequence, nil
}

func (f *mutationChainFake) EstimateGas(context.Context, []byte) (uint64, error) {
	return f.estimatedGas, nil
}

func (f *mutationChainFake) BroadcastTx(_ context.Context, tx []byte) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.broadcasts = append(f.broadcasts, append([]byte(nil), tx...))
	hash := strings.ToUpper(hex.EncodeToString(tmtypes.Tx(tx).Hash()))
	if len(f.broadcastErrs) > 0 {
		err := f.broadcastErrs[0]
		f.broadcastErrs = f.broadcastErrs[1:]
		return hash, err
	}
	f.confirmed[hash] = ProviderTxConfirmation{Found: true, TxHash: hash, Height: 100, BlockHash: f.blockHash}
	return hash, nil
}

func (f *mutationChainFake) ConfirmTx(_ context.Context, hash string) (ProviderTxConfirmation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.confirmMissing {
		return ProviderTxConfirmation{}, nil
	}
	confirmation := f.confirmed[hash]
	if f.confirmMutator != nil && confirmation.Found {
		confirmation = f.confirmMutator(confirmation)
	}
	return confirmation, nil
}

func (f *mutationChainFake) LatestHeight(context.Context) (int64, error) { return f.height, nil }
func (f *mutationChainFake) BlockHash(context.Context, int64) (string, error) {
	return f.blockHash, nil
}

func (f *mutationChainFake) ReconcileMutation(context.Context, *ProviderMutationEnvelope, sdk.Msg) (ProviderMutationReconciliation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reconcileCalls++
	return f.reconciled, nil
}

func newMutationSubmitterForTest(t *testing.T, chain *mutationChainFake, queuePath string) (*ProviderMutationSubmitter, *KeyManager) {
	t.Helper()
	address := sdk.AccAddress(make([]byte, 20)).String()
	keyConfig := DefaultKeyManagerConfig()
	keyConfig.StorageType = KeyStorageTypeMemory
	keyManager, err := NewKeyManager(keyConfig)
	require.NoError(t, err)
	require.NoError(t, keyManager.Unlock(""))
	_, err = keyManager.GenerateKey(address)
	require.NoError(t, err)
	cfg := DefaultProviderMutationSubmitterConfig()
	cfg.ChainID = "virtengine-task85a"
	cfg.ProviderAddress = address
	cfg.QueueStatePath = queuePath
	cfg.Chain = chain
	cfg.PollInterval = time.Millisecond
	cfg.ConfirmationTimeout = 100 * time.Millisecond
	cfg.FinalityBlocks = 2
	cfg.RetryBackoff = time.Millisecond
	cfg.MaxRetryBackoff = 2 * time.Millisecond
	submitter, err := NewProviderMutationSubmitter(cfg, keyManager)
	require.NoError(t, err)
	require.NoError(t, submitter.Start(context.Background()))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		require.NoError(t, submitter.Stop(ctx))
	})
	return submitter, keyManager
}

func validRegistryMessages(address string) map[ProviderMutationKind]sdk.Msg {
	ownerBytes := make([]byte, 20)
	ownerBytes[19] = 1
	owner := sdk.AccAddress(ownerBytes).String()
	return map[ProviderMutationKind]sdk.Msg{
		MutationMarketCreateBid:           &marketv1beta5.MsgCreateBid{ID: marketv1.BidID{Owner: owner, Provider: address, DSeq: 1, GSeq: 1, OSeq: 1}, Price: sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyOneDec()), Deposit: depositv1.Deposit{Amount: sdk.NewInt64Coin("uvirt", 1), Sources: depositv1.Sources{depositv1.SourceBalance}}},
		MutationMarketCloseBid:            &marketv1beta5.MsgCloseBid{ID: marketv1.BidID{Owner: owner, Provider: address, DSeq: 1, GSeq: 1, OSeq: 1}, Reason: marketv1.LeaseClosedReasonUnspecified},
		MutationMarketWithdrawLease:       &marketv1beta5.MsgWithdrawLease{ID: marketv1.LeaseID{Owner: owner, Provider: address, DSeq: 1, GSeq: 1, OSeq: 1}},
		MutationHPCRegisterCluster:        &hpcv1.MsgRegisterCluster{ProviderAddress: address, Name: "cluster", Region: "eu", TotalNodes: 1},
		MutationHPCUpdateCluster:          &hpcv1.MsgUpdateCluster{ProviderAddress: address, ClusterId: "cluster-1"},
		MutationHPCDeregisterCluster:      &hpcv1.MsgDeregisterCluster{ProviderAddress: address, ClusterId: "cluster-1"},
		MutationHPCCreateOffering:         &hpcv1.MsgCreateOffering{ProviderAddress: address, ClusterId: "cluster-1", Name: "compute"},
		MutationHPCUpdateOffering:         &hpcv1.MsgUpdateOffering{ProviderAddress: address, OfferingId: "offering-1"},
		MutationHPCReportJobStatus:        &hpcv1.MsgReportJobStatus{ProviderAddress: address, JobId: "job-1", State: hpcv1.JobStateRunning, SignedTimestamp: 1},
		MutationHPCUpdateNodeMetadata:     &hpcv1.MsgUpdateNodeMetadata{ProviderAddress: address, NodeId: "node-1", ClusterId: "cluster-1", LastSequenceNumber: 1},
		MutationResourcesHeartbeat:        &resourcesv1.MsgProviderHeartbeat{ProviderAddress: address, InventoryId: "inventory-1", ResourceClass: resourcesv1.ResourceClass_RESOURCE_CLASS_COMPUTE, Sequence: 1},
		MutationResourcesActivate:         &resourcesv1.MsgActivateAllocation{ProviderAddress: address, AllocationId: "allocation-1"},
		MutationResourcesRelease:          &resourcesv1.MsgReleaseAllocation{RequesterAddress: address, AllocationId: "allocation-1"},
		MutationSettlementSettleOrder:     &settlementv1.MsgSettleOrder{Sender: address, OrderId: "order-1", UsageRecordIds: []string{"usage-1"}},
		MutationSettlementRecordUsage:     &settlementv1.MsgRecordUsage{Sender: address, OrderId: "order-1", LeaseId: "lease-1", AllocationId: "allocation-1", UsageUnits: 1, UsageType: "compute", PeriodStart: 1, PeriodEnd: 2, UnitPrice: sdk.NewDecCoinFromDec("uvirt", sdkmath.LegacyOneDec()), ChainId: "virtengine-task85a", PricingVersion: 1, FormulaVersion: 1, ModelVersion: 1, StreamSequence: 1, Nonce: make([]byte, 32), IdempotencyKey: bytesOf(1, 32), ProviderKeyEpoch: 1, ProviderKeyId: "key-1", IssuedAtHeight: 1, ExpiresAtHeight: 10, IssuedAtUnix: 1, ExpiresAtUnix: 10, SignatureVersion: 1, Signature: bytesOf(2, 64)},
		MutationSettlementFiatObservation: &settlementv1.MsgRecordFiatConversionObservation{Sender: address, ConversionId: "conversion-1", ObservationSequence: 1, IdempotencyKey: bytesOf(7, 32), Stage: settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED, DexProfileId: "dex-1", DexProfileDigest: bytesOf(8, 32), PayoutProfileId: "payout-1", PayoutProfileDigest: bytesOf(9, 32), QuoteDigest: bytesOf(10, 32), QuoteExpiry: 20, MinimumStableOutput: sdk.NewInt64Coin("uusdc", 1), Status: "accepted", ObservedAt: 10, EvidenceHash: bytesOf(11, 32), ComplianceDecisionHash: bytesOf(12, 32)},
		MutationProviderCreate:            &providerv1beta4.MsgCreateProvider{Owner: address, HostURI: "https://provider.example.com"},
		MutationProviderUpdate:            &providerv1beta4.MsgUpdateProvider{Owner: address, HostURI: "https://provider.example.com"},
		MutationProviderDelete:            &providerv1beta4.MsgDeleteProvider{Owner: address},
		MutationProviderGenerateDomain:    &providerv1beta4.MsgGenerateDomainVerificationToken{Owner: address, Domain: "provider.example.com"},
		MutationProviderVerifyDomain:      &providerv1beta4.MsgVerifyProviderDomain{Owner: address},
		MutationProviderRequestDomain:     &providerv1beta4.MsgRequestDomainVerification{Owner: address, Domain: "provider.example.com", Method: providerv1beta4.VerificationMethod_VERIFICATION_METHOD_DNS_TXT},
		MutationProviderConfirmDomain:     &providerv1beta4.MsgConfirmDomainVerification{Owner: address, Proof: "dns_txt:proof"},
		MutationProviderRevokeDomain:      &providerv1beta4.MsgRevokeDomainVerification{Owner: address},
		MutationProviderSetSigningKey:     &providerv1beta4.MsgSetProviderSigningKey{Owner: address, PublicKey: bytesOf(3, 32), KeyType: providerv1beta4.PublicKeyTypeEd25519},
		MutationProviderRotateKey:         &providerv1beta4.MsgRotateProviderSigningKey{Owner: address, NewPublicKey: bytesOf(4, 32), NewKeyType: providerv1beta4.PublicKeyTypeEd25519, RotationProof: bytesOf(5, 64), SignatureVersion: providerv1beta4.ProviderKeyRotationSignatureVersionV1},
		MutationProviderRevokeKey:         &providerv1beta4.MsgRevokeProviderSigningKey{Owner: address, KeyId: "key-1"},
		MutationMarketplaceCallback:       &marketplacev1.MsgWaldurCallback{Sender: address, CallbackType: "update", ResourceId: "resource-1", Status: "done"},
		MutationSupportUpdateRequest:      &supportv1.MsgUpdateSupportRequest{Sender: address, TicketId: "ticket-1", Status: "open"},
		MutationSupportAddResponse:        &supportv1.MsgAddSupportResponse{Sender: address, TicketId: "ticket-1", Payload: supportv1.EncryptedSupportPayload{EnvelopeRef: "vault://response", EnvelopeHash: bytesOf(6, 32), PayloadSize: 1}},
		MutationSupportRegisterExternal:   &supportv1.MsgRegisterExternalTicket{Sender: address, ResourceId: "ticket-1", ResourceType: "support_request", ExternalSystem: "waldur", ExternalTicketId: "ext-1"},
		MutationSupportUpdateExternal:     &supportv1.MsgUpdateExternalTicket{Sender: address, ResourceId: "ticket-1", ResourceType: "support_request", ExternalTicketId: "ext-2", ExternalUrl: "https://example.com/ext-2"},
	}
}

func bytesOf(value byte, count int) []byte {
	result := make([]byte, count)
	for i := range result {
		result[i] = value
	}
	return result
}

func TestProviderMutationRegistryContractEveryReachableKind(t *testing.T) {
	registry := NewProviderMutationRegistry()
	address := sdk.AccAddress(make([]byte, 20)).String()
	messages := validRegistryMessages(address)
	require.ElementsMatch(t, registry.Kinds(), mapKeys(messages))
	for kind, msg := range messages {
		t.Run(string(kind), func(t *testing.T) {
			envelope, err := registry.Encode("virtengine-task85a", kind, msg)
			require.NoError(t, err)
			require.Equal(t, sdk.MsgTypeURL(msg), envelope.TypeURL)
			decoded, err := registry.Decode(envelope)
			require.NoError(t, err)
			require.Equal(t, sdk.MsgTypeURL(msg), sdk.MsgTypeURL(decoded))
		})
	}
}

func TestProviderMutationSubmitterSignedTxContractEveryRegisteredKind(t *testing.T) {
	registry := NewProviderMutationRegistry()
	for _, kind := range registry.Kinds() {
		t.Run(string(kind), func(t *testing.T) {
			chain := newMutationChainFake()
			submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
			msg, ok := validRegistryMessages(submitter.cfg.ProviderAddress)[kind]
			require.Truef(t, ok, "missing valid message fixture for %s", kind)

			result, err := submitter.Submit(context.Background(), kind, msg)
			require.NoError(t, err)
			require.True(t, result.Final)

			chain.mu.Lock()
			require.Len(t, chain.broadcasts, 1)
			txBytes := append([]byte(nil), chain.broadcasts[0]...)
			chain.mu.Unlock()

			tx, err := submitter.encCfg.TxConfig.TxDecoder()(txBytes)
			require.NoError(t, err)
			require.Len(t, tx.GetMsgs(), 1)
			require.Equal(t, sdk.MsgTypeURL(msg), sdk.MsgTypeURL(tx.GetMsgs()[0]))
			signatures, err := tx.(interface {
				GetSignaturesV2() ([]signing.SignatureV2, error)
			}).GetSignaturesV2()
			require.NoError(t, err)
			require.Len(t, signatures, 1)
			require.NotNil(t, signatures[0].PubKey)
		})
	}
}

func TestProviderMutationRegistryRejectsUnknownAndMismatchedTypes(t *testing.T) {
	registry := NewProviderMutationRegistry()
	address := sdk.AccAddress(make([]byte, 20)).String()
	_, err := registry.Encode("chain", "unknown.kind", &providerv1beta4.MsgDeleteProvider{Owner: address})
	require.ErrorIs(t, err, ErrUnknownProviderMutation)
	_, err = registry.Encode("chain", MutationProviderDelete, &providerv1beta4.MsgConfirmDomainVerification{Owner: address, Proof: "proof"})
	require.ErrorIs(t, err, ErrUnknownProviderMutation)
}

func TestProviderMutationRegistryDeterministicMapDigest(t *testing.T) {
	registry := NewProviderMutationRegistry()
	address := sdk.AccAddress(make([]byte, 20)).String()
	first := &supportv1.MsgUpdateSupportRequest{
		Sender:   address,
		TicketId: "ticket-1",
		Status:   "open",
		PublicMetadata: map[string]string{
			"b": "2",
			"a": "1",
		},
	}
	second := &supportv1.MsgUpdateSupportRequest{
		Sender:   address,
		TicketId: "ticket-1",
		Status:   "open",
		PublicMetadata: map[string]string{
			"a": "1",
			"b": "2",
		},
	}

	firstEnvelope, err := registry.Encode("chain", MutationSupportUpdateRequest, first)
	require.NoError(t, err)
	secondEnvelope, err := registry.Encode("chain", MutationSupportUpdateRequest, second)
	require.NoError(t, err)
	require.Equal(t, firstEnvelope.MessageDigest, secondEnvelope.MessageDigest)
	require.Equal(t, firstEnvelope.ID, secondEnvelope.ID)
}

func TestProviderMutationRegistryRejectsCustomerSignedLeaseMessages(t *testing.T) {
	registry := NewProviderMutationRegistry()
	address := sdk.AccAddress(make([]byte, 20)).String()
	ownerBytes := make([]byte, 20)
	ownerBytes[19] = 1
	owner := sdk.AccAddress(ownerBytes).String()

	_, err := registry.Encode("chain", MutationMarketCreateLease, &marketv1beta5.MsgCreateLease{
		BidID: marketv1.BidID{Owner: owner, Provider: address, DSeq: 1, GSeq: 1, OSeq: 1},
	})
	require.ErrorIs(t, err, ErrUnknownProviderMutation)
	_, err = registry.Encode("chain", MutationMarketCloseLease, &marketv1beta5.MsgCloseLease{
		ID: marketv1.LeaseID{Owner: owner, Provider: address, DSeq: 1, GSeq: 1, OSeq: 1},
	})
	require.ErrorIs(t, err, ErrUnknownProviderMutation)
}

func TestProviderMutationSubmitterBuildsDecodableSignedSDKTx(t *testing.T) {
	chain := newMutationChainFake()
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	address := submitter.cfg.ProviderAddress
	result, err := submitter.Submit(context.Background(), MutationHPCReportJobStatus, &hpcv1.MsgReportJobStatus{ProviderAddress: address, JobId: "job-1", State: hpcv1.JobStateRunning, SignedTimestamp: 1})
	require.NoError(t, err)
	require.True(t, result.Final)
	chain.mu.Lock()
	require.Len(t, chain.broadcasts, 1)
	txBytes := append([]byte(nil), chain.broadcasts[0]...)
	chain.mu.Unlock()
	tx, err := submitter.encCfg.TxConfig.TxDecoder()(txBytes)
	require.NoError(t, err)
	require.Len(t, tx.GetMsgs(), 1)
	require.IsType(t, &hpcv1.MsgReportJobStatus{}, tx.GetMsgs()[0])
	signatures, err := tx.(interface {
		GetSignaturesV2() ([]signing.SignatureV2, error)
	}).GetSignaturesV2()
	require.NoError(t, err)
	require.Len(t, signatures, 1)
	require.NotNil(t, signatures[0].PubKey)
}

func TestProviderMutationSubmitterUnavailableAndReadiness(t *testing.T) {
	cfg := DefaultProviderMutationSubmitterConfig()
	cfg.ChainID = "chain"
	cfg.ProviderAddress = sdk.AccAddress(make([]byte, 20)).String()
	_, err := NewProviderMutationSubmitter(cfg, nil)
	require.ErrorIs(t, err, ErrProviderMutationUnavailable)
	client := &rpcChainClient{}
	err = client.SubmitResourceHeartbeat(context.Background(), &resourcesv1.MsgProviderHeartbeat{})
	require.ErrorIs(t, err, ErrProviderMutationUnavailable)
	require.False(t, client.MutationReadiness(context.Background()).Ready)
}

func TestProviderMutationSubmitterResponseLostReconcilesBeforeRebuild(t *testing.T) {
	chain := newMutationChainFake()
	chain.broadcastErrs = []error{context.DeadlineExceeded}
	chain.reconciled = ProviderMutationReconciliation{Committed: true, TxHash: "reconciled", Height: 100, BlockHash: chain.blockHash, Reason: "logical_state_committed"}
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	address := submitter.cfg.ProviderAddress
	result, err := submitter.Submit(context.Background(), MutationProviderDelete, &providerv1beta4.MsgDeleteProvider{Owner: address})
	require.NoError(t, err)
	require.True(t, result.Final)
	chain.mu.Lock()
	defer chain.mu.Unlock()
	require.Len(t, chain.broadcasts, 1)
	require.Positive(t, chain.reconcileCalls)
}

func TestProviderMutationSubmitterMutableIdempotencyAllowsNextTerminalUpdate(t *testing.T) {
	chain := newMutationChainFake()
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	address := submitter.cfg.ProviderAddress

	first, err := submitter.Submit(context.Background(), MutationProviderUpdate, &providerv1beta4.MsgUpdateProvider{Owner: address, HostURI: "https://one.example.com"})
	require.NoError(t, err)
	require.True(t, first.Final)

	second, err := submitter.Submit(context.Background(), MutationProviderUpdate, &providerv1beta4.MsgUpdateProvider{Owner: address, HostURI: "https://two.example.com"})
	require.NoError(t, err)
	require.True(t, second.Final)
	require.False(t, second.Existed)
	require.NotEqual(t, first.ID, second.ID)

	chain.mu.Lock()
	require.Len(t, chain.broadcasts, 2)
	chain.mu.Unlock()
}

func TestProviderMutationSubmitterMutableIdempotencyConflictsWhileActive(t *testing.T) {
	chain := newMutationChainFake()
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	address := submitter.cfg.ProviderAddress
	pending, err := submitter.registry.Encode(submitter.cfg.ChainID, MutationProviderUpdate, &providerv1beta4.MsgUpdateProvider{Owner: address, HostURI: "https://one.example.com"})
	require.NoError(t, err)
	pending.LeaseToken = submitter.leaseToken
	_, _, err = submitter.store.PutIfAbsent(context.Background(), pending)
	require.NoError(t, err)

	_, err = submitter.Submit(context.Background(), MutationProviderUpdate, &providerv1beta4.MsgUpdateProvider{Owner: address, HostURI: "https://two.example.com"})
	require.ErrorIs(t, err, ErrProviderMutationConflict)
}

func TestProviderMutationSubmitterRestartRecoversBuiltItem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.json")
	chain := newMutationChainFake()
	submitter, keyManager := newMutationSubmitterForTest(t, chain, path)
	envelope, err := submitter.registry.Encode(submitter.cfg.ChainID, MutationProviderDelete, &providerv1beta4.MsgDeleteProvider{Owner: submitter.cfg.ProviderAddress})
	require.NoError(t, err)
	envelope.State = MutationStateBuilt
	envelope.TxHash = "unknown-hash"
	envelope.NextAttemptAt = time.Now().UTC().Add(time.Hour)
	_, _, err = submitter.store.PutIfAbsent(context.Background(), envelope)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	require.NoError(t, submitter.Stop(ctx))
	cancel()

	cfg := submitter.cfg
	cfg.PollInterval = time.Hour
	second, err := NewProviderMutationSubmitter(cfg, keyManager)
	require.NoError(t, err)
	require.NoError(t, second.Start(context.Background()))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = second.Stop(ctx)
	})
	recovered, err := second.store.Get(context.Background(), envelope.ID)
	require.NoError(t, err)
	require.Equal(t, MutationStateAmbiguous, recovered.State)
	require.Equal(t, "restart_reconciliation_required", recovered.ReconciliationState)
}

func TestFileProviderMutationStorePersistsAndReopens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.json")
	store, err := NewFileProviderMutationStore(path)
	require.NoError(t, err)
	require.NoError(t, store.Open(context.Background()))
	envelope := &ProviderMutationEnvelope{
		SchemaVersion:  providerMutationSchemaVersion,
		ID:             "mutation-1",
		Kind:           MutationProviderDelete,
		TypeURL:        sdk.MsgTypeURL(&providerv1beta4.MsgDeleteProvider{}),
		MessageDigest:  strings.Repeat("a", 64),
		Signer:         "provider",
		IdempotencyKey: "idem-1",
		State:          MutationStatePending,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		NextAttemptAt:  time.Now().UTC(),
	}
	_, existed, err := store.PutIfAbsent(context.Background(), envelope)
	require.NoError(t, err)
	require.False(t, existed)
	require.NoError(t, store.Close())

	reopened, err := NewFileProviderMutationStore(path)
	require.NoError(t, err)
	require.NoError(t, reopened.Open(context.Background()))
	t.Cleanup(func() { _ = reopened.Close() })
	stored, err := reopened.Get(context.Background(), envelope.ID)
	require.NoError(t, err)
	require.Equal(t, envelope.IdempotencyKey, stored.IdempotencyKey)
}

func TestProviderMutationConfirmationEvidenceRequired(t *testing.T) {
	tests := []struct {
		name    string
		mutator func(ProviderTxConfirmation) ProviderTxConfirmation
	}{
		{
			name: "missing height",
			mutator: func(c ProviderTxConfirmation) ProviderTxConfirmation {
				c.Height = 0
				return c
			},
		},
		{
			name: "missing block hash",
			mutator: func(c ProviderTxConfirmation) ProviderTxConfirmation {
				c.BlockHash = ""
				return c
			},
		},
		{
			name: "wrong tx hash",
			mutator: func(c ProviderTxConfirmation) ProviderTxConfirmation {
				c.TxHash = "WRONG"
				return c
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chain := newMutationChainFake()
			chain.confirmMutator = test.mutator
			submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
			result, err := submitter.Submit(context.Background(), MutationProviderDelete, &providerv1beta4.MsgDeleteProvider{Owner: submitter.cfg.ProviderAddress})
			require.ErrorIs(t, err, ErrProviderMutationEvidence)
			status, statusErr := submitter.Status(context.Background(), result.ID)
			require.NoError(t, statusErr)
			require.Equal(t, MutationStateAmbiguous, status.State)
		})
	}
}

func TestProviderMutationSubmitterLeaseLossDuringConfirmationFailsClosed(t *testing.T) {
	chain := newMutationChainFake()
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	chain.confirmMutator = func(c ProviderTxConfirmation) ProviderTxConfirmation {
		require.NoError(t, submitter.lease.Release(context.Background(), submitter.leaseName, submitter.leaseToken))
		return c
	}

	result, err := submitter.Submit(context.Background(), MutationProviderDelete, &providerv1beta4.MsgDeleteProvider{Owner: submitter.cfg.ProviderAddress})
	require.ErrorIs(t, err, ErrProviderMutationNotReady)
	status, statusErr := submitter.Status(context.Background(), result.ID)
	require.NoError(t, statusErr)
	require.Equal(t, MutationStateAmbiguous, status.State)
	require.False(t, submitter.Readiness(context.Background()).Ready)
}

func TestProviderMutationSubmitterDeadLettersTerminalFailure(t *testing.T) {
	chain := newMutationChainFake()
	chain.broadcastErrs = []error{errors.New("unauthorized signature")}
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	result, err := submitter.Submit(context.Background(), MutationProviderDelete, &providerv1beta4.MsgDeleteProvider{Owner: submitter.cfg.ProviderAddress})
	require.Error(t, err)
	status, statusErr := submitter.Status(context.Background(), result.ID)
	require.NoError(t, statusErr)
	require.Equal(t, MutationStateDeadLetter, status.State)
	require.Equal(t, 1, submitter.Metrics(context.Background()).DeadLetters)
}

func TestProviderMutationFailureClassificationMatrix(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		classification ProviderMutationClassification
		retryable      bool
		ambiguous      bool
	}{
		{name: "sequence", err: ErrSequenceMismatch, classification: MutationClassSequenceMismatch, retryable: true},
		{name: "out of gas", err: errors.New("out of gas in location"), classification: MutationClassOutOfGas, retryable: true},
		{name: "mempool", err: errors.New("mempool is full"), classification: MutationClassMempoolReject, retryable: true},
		{name: "timeout", err: context.DeadlineExceeded, classification: MutationClassTimeout, retryable: true, ambiguous: true},
		{name: "unavailable", err: errors.New("connection unavailable"), classification: MutationClassUnavailable, retryable: true, ambiguous: true},
		{name: "signature", err: errors.New("signature unauthorized"), classification: MutationClassUnauthorized},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			classification, retryable, ambiguous := classifyProviderMutationError(test.err)
			require.Equal(t, test.classification, classification)
			require.Equal(t, test.retryable, retryable)
			require.Equal(t, test.ambiguous, ambiguous)
		})
	}
}

func TestProviderMutationGasAdjustmentChecked(t *testing.T) {
	adjusted, err := checkedAdjustedGasLimit(100, 1.25)
	require.NoError(t, err)
	require.Equal(t, uint64(125), adjusted)
	adjusted, err = checkedAdjustedGasLimit(101, 1.25)
	require.NoError(t, err)
	require.Equal(t, uint64(127), adjusted)
	_, err = checkedAdjustedGasLimit(^uint64(0), 2)
	require.Error(t, err)
	_, err = checkedAdjustedGasLimit(100, math.NaN())
	require.Error(t, err)
}

func TestProviderMutationConfirmationTimeoutRemainsAmbiguous(t *testing.T) {
	chain := newMutationChainFake()
	chain.confirmMissing = true
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	submitter.cfg.ConfirmationTimeout = 5 * time.Millisecond
	result, err := submitter.Submit(context.Background(), MutationProviderDelete, &providerv1beta4.MsgDeleteProvider{Owner: submitter.cfg.ProviderAddress})
	require.Error(t, err)
	status, statusErr := submitter.Status(context.Background(), result.ID)
	require.NoError(t, statusErr)
	require.Equal(t, MutationStateAmbiguous, status.State)
}

func TestProviderMutationReorgReturnsExplicitRetry(t *testing.T) {
	chain := newMutationChainFake()
	submitter, _ := newMutationSubmitterForTest(t, chain, filepath.Join(t.TempDir(), "queue.json"))
	envelope, err := submitter.registry.Encode(submitter.cfg.ChainID, MutationProviderDelete, &providerv1beta4.MsgDeleteProvider{Owner: submitter.cfg.ProviderAddress})
	require.NoError(t, err)
	envelope.State = MutationStateIncluded
	envelope.ConfirmationHeight = 100
	envelope.ConfirmationBlockHash = "OLD-BLOCK"
	envelope.TxHash = "hash"
	_, _, err = submitter.store.PutIfAbsent(context.Background(), envelope)
	require.NoError(t, err)
	err = submitter.awaitFinality(context.Background(), envelope.ID)
	require.ErrorIs(t, err, ErrProviderMutationReorg)
	stored, getErr := submitter.store.Get(context.Background(), envelope.ID)
	require.NoError(t, getErr)
	require.Equal(t, MutationStateAmbiguous, stored.State)
	require.Equal(t, "reorg_detected", stored.ReconciliationState)
}

func mapKeys(values map[ProviderMutationKind]sdk.Msg) []ProviderMutationKind {
	result := make([]ProviderMutationKind, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return result
}
