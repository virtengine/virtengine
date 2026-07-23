// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

const (
	testSwapTxHash  = "6482380f6c3b8efab6d0efcc99eb3e928767a6748196a47480760cd6e32a1790"
	testCustodyAddr = "osmo1custody"
)

type fiatQueryFake struct {
	params           settlementv1.Params
	records          map[string]*settlementv1.FiatConversionRecord
	activeCaseID     string
	activeHoldCount  uint32
	err              error
	authorizationErr error
}

func (f *fiatQueryFake) Params(context.Context) (settlementv1.Params, error) {
	if f.err != nil {
		return settlementv1.Params{}, f.err
	}
	return f.params, nil
}
func (f *fiatQueryFake) GetConversion(_ context.Context, id string) (*settlementv1.FiatConversionRecord, error) {
	if f.err != nil {
		return nil, f.err
	}
	value := f.records[id]
	if value == nil {
		return nil, ErrFiatConversionQueryUnavailable
	}
	copyValue := *value
	return &copyValue, nil
}
func (f *fiatQueryFake) ListNonTerminalConversions(context.Context, string) ([]settlementv1.FiatConversionRecord, error) {
	if f.err != nil {
		return nil, f.err
	}
	result := make([]settlementv1.FiatConversionRecord, 0, len(f.records))
	for _, value := range f.records {
		result = append(result, *value)
	}
	return result, nil
}
func (f *fiatQueryFake) ExecutionAuthorization(_ context.Context, id string) (FiatExecutionAuthorization, error) {
	if f.authorizationErr != nil {
		return FiatExecutionAuthorization{}, f.authorizationErr
	}
	if f.err != nil {
		return FiatExecutionAuthorization{}, f.err
	}
	value := f.records[id]
	if value == nil {
		return FiatExecutionAuthorization{}, ErrFiatConversionQueryUnavailable
	}
	copyValue := *value
	return FiatExecutionAuthorization{
		Conversion: &copyValue, Params: f.params,
		ActiveCaseID: f.activeCaseID, ActiveHoldCount: f.activeHoldCount,
	}, nil
}

type observationSubmitterFake struct {
	submitted []*settlementv1.MsgRecordFiatConversionObservation
	ready     bool
	query     *fiatQueryFake
}

func (f *observationSubmitterFake) SubmitFiatConversionObservation(_ context.Context, msg *settlementv1.MsgRecordFiatConversionObservation) (ProviderMutationResult, error) {
	copyMsg := *msg
	f.submitted = append(f.submitted, &copyMsg)
	if f.query != nil {
		applyTestObservation(f.query.records[msg.ConversionId], msg)
	}
	return ProviderMutationResult{ID: "mutation", State: MutationStateConfirmed, Final: true}, nil
}
func (f *observationSubmitterFake) Readiness(context.Context) ProviderMutationReadiness {
	return ProviderMutationReadiness{Ready: f.ready}
}

type resolverFake struct {
	destination       string
	decision          offramp.ComplianceDecision
	destinationDigest []byte
	complianceDigest  []byte
	err               error
}

func (r resolverFake) ResolveDestination(context.Context, *settlementv1.FiatConversionRecord) (ResolvedFiatDestination, error) {
	return ResolvedFiatDestination{Reference: r.destination, Digest: append([]byte(nil), r.destinationDigest...)}, r.err
}
func (r resolverFake) ResolveCompliance(context.Context, *settlementv1.FiatConversionRecord) (ResolvedFiatCompliance, error) {
	return ResolvedFiatCompliance{Decision: r.decision, Digest: append([]byte(nil), r.complianceDigest...)}, r.err
}
func (r resolverFake) ProductionReady(context.Context) error { return r.err }

type custodyFake struct {
	testOnly bool
	err      error
}

func (c custodyFake) Address(context.Context, string) (string, error) { return testCustodyAddr, c.err }
func (c custodyFake) SignExecution(context.Context, dex.SwapQuote, []byte) ([]byte, error) {
	return []byte("txraw"), c.err
}
func (c custodyFake) RecoverSignedExecution(context.Context, dex.SwapQuote, []byte, string) ([]byte, error) {
	return []byte("txraw"), c.err
}
func (c custodyFake) VerifySignedExecution(context.Context, []byte, []byte) error { return c.err }
func (c custodyFake) ProductionReady(context.Context) error                       { return c.err }
func (c custodyFake) TestOnly() bool                                              { return c.testOnly }

type fiatDEXFake struct {
	quote          dex.SwapQuote
	reconciliation DEXSwapReconciliation
	err            error
}

type sideEffectDEXFake struct {
	quote          dex.SwapQuote
	reconciliation DEXSwapReconciliation
	quoteCalls     int
	executeCalls   int
	reconcileCalls int
}

func (f *sideEffectDEXFake) GetSwapQuote(context.Context, dex.SwapRequest) (dex.SwapQuote, error) {
	f.quoteCalls++
	return f.quote, nil
}
func (f *sideEffectDEXFake) ExecuteSwap(context.Context, dex.SwapQuote, []byte) (dex.SwapResult, error) {
	f.executeCalls++
	return dex.SwapResult{TxHash: testSwapTxHash, OutputAmount: sdkmath.NewInt(99)}, nil
}
func (f *sideEffectDEXFake) ReconcileSwap(context.Context, dex.SwapQuote, string, string) (DEXSwapReconciliation, error) {
	f.reconcileCalls++
	return f.reconciliation, nil
}

type sideEffectCustodyFake struct {
	signCalls    int
	recoverCalls int
	afterSign    func()
}

func (*sideEffectCustodyFake) Address(context.Context, string) (string, error) {
	return testCustodyAddr, nil
}
func (f *sideEffectCustodyFake) SignExecution(context.Context, dex.SwapQuote, []byte) ([]byte, error) {
	f.signCalls++
	if f.afterSign != nil {
		f.afterSign()
	}
	return []byte("txraw"), nil
}
func (f *sideEffectCustodyFake) RecoverSignedExecution(context.Context, dex.SwapQuote, []byte, string) ([]byte, error) {
	f.recoverCalls++
	return []byte("txraw"), nil
}
func (*sideEffectCustodyFake) VerifySignedExecution(context.Context, []byte, []byte) error {
	return nil
}
func (*sideEffectCustodyFake) ProductionReady(context.Context) error { return nil }
func (*sideEffectCustodyFake) TestOnly() bool                        { return false }

func (f fiatDEXFake) GetSwapQuote(context.Context, dex.SwapRequest) (dex.SwapQuote, error) {
	return f.quote, f.err
}
func (f fiatDEXFake) ExecuteSwap(context.Context, dex.SwapQuote, []byte) (dex.SwapResult, error) {
	return dex.SwapResult{TxHash: testSwapTxHash, OutputAmount: sdkmath.NewInt(99)}, f.err
}
func (f fiatDEXFake) ReconcileSwap(context.Context, dex.SwapQuote, string, string) (DEXSwapReconciliation, error) {
	return f.reconciliation, f.err
}

type bridgeFake struct{}

func (bridgeFake) RegisterAdapter(offramp.Adapter) error { return nil }
func (bridgeFake) GetQuote(context.Context, offramp.QuoteRequest) (offramp.Quote, error) {
	return offramp.Quote{}, errors.New("unused")
}
func (bridgeFake) InitiatePayout(context.Context, offramp.Quote, string, string, map[string]string) (offramp.PayoutResult, error) {
	return offramp.PayoutResult{}, errors.New("unused")
}
func (bridgeFake) GetStatus(context.Context, string) (offramp.PayoutResult, error) {
	return offramp.PayoutResult{}, errors.New("unused")
}
func (bridgeFake) FindPayoutByMetadata(context.Context, string, map[string]string) (offramp.PayoutResult, error) {
	return offramp.PayoutResult{}, errors.New("unused")
}
func (bridgeFake) Cancel(context.Context, string) error { return nil }
func (bridgeFake) ListProviders() []string              { return nil }

type scriptedBridge struct {
	now            time.Time
	quoteCalls     int
	payoutCalls    int
	statusCalls    int
	metadataCalls  int
	initiateErr    error
	metadataResult offramp.PayoutResult
}

func (*scriptedBridge) RegisterAdapter(offramp.Adapter) error { return nil }
func (b *scriptedBridge) GetQuote(_ context.Context, request offramp.QuoteRequest) (offramp.Quote, error) {
	b.quoteCalls++
	return offramp.Quote{ID: "payout-quote", Request: request, FiatAmount: sdkmath.LegacyNewDec(99), ExchangeRate: sdkmath.LegacyOneDec(), Fee: sdkmath.NewInt(1), Provider: "partner", CreatedAt: b.now, ExpiresAt: b.now.Add(time.Minute)}, nil
}
func (b *scriptedBridge) InitiatePayout(_ context.Context, quote offramp.Quote, _ string, _ string, metadata map[string]string) (offramp.PayoutResult, error) {
	b.payoutCalls++
	if b.initiateErr != nil {
		return offramp.PayoutResult{}, b.initiateErr
	}
	return testPayoutResult(b.now, quote, metadata, offramp.StatusProcessing), nil
}
func (b *scriptedBridge) GetStatus(_ context.Context, _ string) (offramp.PayoutResult, error) {
	b.statusCalls++
	if b.metadataResult.ID != "" {
		return b.metadataResult, nil
	}
	quote := offramp.Quote{ID: "payout-quote", FiatAmount: sdkmath.LegacyNewDec(99), Fee: sdkmath.NewInt(1), Provider: "partner"}
	return testPayoutResult(b.now, quote, map[string]string{"idempotency_key": "id", "correlation_id": "correlation"}, offramp.StatusCompleted), nil
}
func (b *scriptedBridge) FindPayoutByMetadata(_ context.Context, _ string, _ map[string]string) (offramp.PayoutResult, error) {
	b.metadataCalls++
	if b.metadataResult.ID == "" {
		return offramp.PayoutResult{}, offramp.ErrPayoutNotFound
	}
	return b.metadataResult, nil
}
func (*scriptedBridge) Cancel(context.Context, string) error { return nil }
func (*scriptedBridge) ListProviders() []string              { return []string{"partner"} }

func testPayoutResult(now time.Time, quote offramp.Quote, metadata map[string]string, status offramp.Status) offramp.PayoutResult {
	result := offramp.PayoutResult{ID: "external-payout", QuoteID: quote.ID, Status: status, Provider: "partner", FiatAmount: sdkmath.LegacyNewDec(99), CryptoAmount: sdkmath.NewInt(99), Fee: sdkmath.NewInt(1), Reference: "safe-reference", Metadata: metadata, InitiatedAt: now, StatusUpdatedAt: now}
	if status == offramp.StatusCompleted {
		completed := now
		result.CompletedAt = &completed
	}
	return result
}

func applyTestObservation(record *settlementv1.FiatConversionRecord, msg *settlementv1.MsgRecordFiatConversionObservation) {
	if record == nil {
		return
	}
	raw, _ := proto.Marshal(msg)
	digest := sha256.Sum256(raw)
	record.ObservationSequence = msg.ObservationSequence
	record.LastObservationDigest = digest[:]
	switch msg.Stage {
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED:
		record.State, record.QuoteDigest, record.QuoteExpiry, record.MinimumStableOutput = "SWAP_PENDING", msg.QuoteDigest, msg.QuoteExpiry, msg.MinimumStableOutput
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED:
		record.State, record.SwapTxHash = "SWAP_SUBMITTED", msg.SwapTxHash
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_FINALIZED:
		record.State, record.StableAmount = "SWAP_SETTLED", msg.StableAmount
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_QUOTED:
		record.State, record.OffRampQuoteId, record.QuoteDigest, record.QuoteExpiry = fiatChainStatePayoutPending, msg.OffRampQuoteId, msg.QuoteDigest, msg.QuoteExpiry
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_SUBMITTED:
		record.State, record.OffRampId = "PAYOUT_SUBMITTED", msg.OffRampPayoutId
	case settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_COMPLETED:
		record.State = "PAYOUT_COMPLETED"
	}
}

func testTrustedProfiles() *TrustedFiatProfiles {
	dexProfile := dex.DEXRouteProfile{ID: "dex", State: dex.RouteCertifiedEnabled, Network: "osmosis", ChainID: dex.OsmosisChainIDMainnet, Environment: dex.EnvironmentMainnet, DEX: "osmosis", Version: dex.OsmosisAdapterVersion, AllowedPoolIDs: []string{"1"}, Tokens: []dex.RouteToken{{Symbol: "UVE", Denom: "uve", Decimals: 6}, {Symbol: "USDC", Denom: "uusdc", Decimals: 6}}, FinalityBlocks: 2, MaxObservationAge: time.Minute, MaxHeightLag: 2, MaxHops: 1, MinLiquidity: sdkmath.NewInt(1), MinReserve: sdkmath.NewInt(1), MaxAmount: sdkmath.NewInt(1000), MaxPriceImpact: sdkmath.LegacyMustNewDecFromStr("0.1"), MaxOracleDeviation: sdkmath.LegacyMustNewDecFromStr("0.1"), QuoteTTL: time.Minute, CustodyMode: "external", OracleSource: "oracle", Evidence: dex.RouteEvidence{EngineeringEvidence: "e", NetworkEvidence: "n", LiquidityEvidence: "l", OracleEvidence: "o", CustodyEvidence: "c", GovernanceEvidence: "g", EngineeringOwners: []string{"e"}, OperationsOwners: []string{"o"}, SecurityOwners: []string{"s"}}}
	payout := offramp.PayoutProfile{ID: "payout", State: offramp.ProfileCertifiedEnabled, Provider: "partner", APIVersion: "1", Environment: offramp.EnvironmentProduction, Corridors: []offramp.PayoutCorridor{{ID: "US-USD-ach", Jurisdiction: "US", Currency: "USD", Rail: "ach", MinimumAmount: sdkmath.LegacyOneDec(), MaximumAmount: sdkmath.LegacyNewDec(1000), DailyLimit: sdkmath.LegacyNewDec(1000), QuoteTTL: time.Minute, Finality: "webhook"}}, BeneficiaryRequirements: offramp.BeneficiaryRequirements{TokenizedReferenceRequired: true, ReferencePrefix: "token-", RequiredFields: []string{"beneficiary_reference"}, ProhibitedRawFields: []string{"account_number"}}, DecisionRequirements: offramp.DecisionRequirements{KYCRequired: true, SanctionsRequired: true}, CredentialSecretRefs: []offramp.SecretReference{{Purpose: "api", Ref: "vault://prod/api", Version: "1", Scope: "production"}}, Webhook: offramp.WebhookProfile{Version: "1", Algorithm: "HMAC-SHA256", Keys: []offramp.WebhookKeyReference{{KeyID: "key", Version: "1", SecretRef: "vault://prod/webhook"}}}, Evidence: offramp.ProfileEvidence{Contract: approval(), Legal: approval(), DPA: approval(), Compliance: approval(), Custody: approval(), Banking: approval(), WebhookRegistration: approval(), Corridor: approval()}, Owners: offramp.ProfileOwners{Engineering: "e", Operations: "o", Compliance: "c", Security: "s"}}
	dexDigest, _ := canonicalProfileDigest(dexProfile)
	payoutDigest, _ := canonicalProfileDigest(payout)
	return &TrustedFiatProfiles{DEX: dexProfile, Payout: payout, DEXDigest: dexDigest, PayoutDigest: payoutDigest, DEXTrusted: true, PayoutTrusted: true}
}
func approval() offramp.ApprovalEvidence {
	return offramp.ApprovalEvidence{Reference: "approved", Owner: "owner"}
}

func testChainIntent(profiles *TrustedFiatProfiles) *settlementv1.FiatConversionRecord {
	destinationDigest := sha256.Sum256([]byte("token-beneficiary"))
	return &settlementv1.FiatConversionRecord{ConversionId: "conversion-1", SettlementId: "settlement-1", PayoutId: "payout-chain-1", Provider: "provider", Customer: "customer", State: "CREATED", CryptoToken: settlementv1.TokenSpec{Symbol: "UVE", Denom: "uve", Decimals: 6, ChainId: dex.OsmosisChainIDMainnet}, StableToken: settlementv1.TokenSpec{Symbol: "USDC", Denom: "uusdc", Decimals: 6, ChainId: dex.OsmosisChainIDMainnet}, CryptoAmount: sdk.NewInt64Coin("uve", 100), FiatCurrency: "USD", PaymentMethod: "ach", DestinationHash: hex.EncodeToString(destinationDigest[:]), DestinationRegion: "US", RequestDigest: bytes.Repeat([]byte{1}, 32), ComplianceDecisionHash: bytes.Repeat([]byte{2}, 32), DexProfileId: profiles.DEX.ID, DexProfileDigest: profiles.DEXDigest[:], PayoutProfileId: profiles.Payout.ID, PayoutProfileDigest: profiles.PayoutDigest[:]}
}

func newOrchestratorFixture(t *testing.T) (*FiatConversionOrchestrator, *fiatQueryFake, *observationSubmitterFake, *FileFiatRepository) {
	t.Helper()
	profiles := testTrustedProfiles()
	intent := testChainIntent(profiles)
	query := &fiatQueryFake{params: settlementv1.Params{FiatConversionEnabled: true, FiatConversionDexProfileId: profiles.DEX.ID, FiatConversionDexProfileDigest: profiles.DEXDigest[:], FiatConversionDexProfileState: settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED, FiatConversionPayoutProfileId: profiles.Payout.ID, FiatConversionPayoutProfileDigest: profiles.PayoutDigest[:], FiatConversionPayoutProfileState: settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED}, records: map[string]*settlementv1.FiatConversionRecord{intent.ConversionId: intent}}
	submitter := &observationSubmitterFake{ready: true, query: query}
	store, err := NewFileFiatConversionStore(filepath.Join(t.TempDir(), "orchestrator.json"))
	require.NoError(t, err)
	repository, err := NewFileFiatRepository(filepath.Join(t.TempDir(), "repository.json"))
	require.NoError(t, err)
	require.NoError(t, repository.Open(context.Background()))
	t.Cleanup(func() { _ = repository.Close() })
	now := time.Unix(1_700_000_000, 0).UTC()
	destinationDigest := sha256.Sum256([]byte("token-beneficiary"))
	orchestrator, err := NewFiatConversionOrchestrator(FiatConversionOrchestratorConfig{Enabled: true, Production: true, ProviderAddress: "provider", Store: store, Lease: NewLocalFiatConversionLease(), LeaseTTL: time.Minute, PollInterval: time.Hour, RetryBackoff: time.Millisecond, MaxRetryBackoff: time.Second, MaxAttempts: 3, Now: func() time.Time { return now }, Profiles: profiles, Query: query, Submitter: submitter, DEX: fiatDEXFake{}, Custody: custodyFake{}, Offramp: bridgeFake{}, Destination: resolverFake{destination: "token-beneficiary", destinationDigest: destinationDigest[:]}, Compliance: resolverFake{decision: offramp.ComplianceDecision{Reference: "decision", KYCDecision: "approved", SanctionsDecision: "approved", ValidUntil: now.Add(time.Hour)}, complianceDigest: bytes.Repeat([]byte{2}, 32)}, WebhookEvents: repository, WebhookBindings: repository})
	require.NoError(t, err)
	require.NoError(t, orchestrator.Start(context.Background()))
	t.Cleanup(func() { _ = orchestrator.Stop(context.Background()) })
	return orchestrator, query, submitter, repository
}

func TestFileFiatConversionStorePersistsPrivacySafeState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store, err := NewFileFiatConversionStore(path)
	require.NoError(t, err)
	require.NoError(t, store.Open(context.Background()))
	secret := "4111111111111111"
	item := minimalFiatWorkItem(FiatWorkClaimed)
	item.Intent.DestinationHash = digestHex([]byte(secret))
	_, _, err = store.PutIfAbsent(context.Background(), item)
	require.NoError(t, err)
	require.NoError(t, store.Close())
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NotContains(t, string(raw), secret)
	require.Contains(t, string(raw), item.Intent.DestinationHash)
	reopened, err := NewFileFiatConversionStore(path)
	require.NoError(t, err)
	require.NoError(t, reopened.Open(context.Background()))
	defer reopened.Close()
	loaded, err := reopened.Get(context.Background(), "conversion")
	require.NoError(t, err)
	require.Equal(t, FiatWorkClaimed, loaded.State)
}

func TestFileFiatConversionStoreRejectsTrailingJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store, err := NewFileFiatConversionStore(path)
	require.NoError(t, err)
	require.NoError(t, store.Open(context.Background()))
	_, _, err = store.PutIfAbsent(context.Background(), minimalFiatWorkItem(FiatWorkClaimed))
	require.NoError(t, err)
	require.NoError(t, store.Close())

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0) // #nosec G304 -- test-owned temporary path.
	require.NoError(t, err)
	_, err = file.WriteString("\n{}")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	reopened, err := NewFileFiatConversionStore(path)
	require.NoError(t, err)
	require.ErrorContains(t, reopened.Open(context.Background()), "multiple JSON values")
}

func TestFiatConversionProductionRejectsTestCustodyAndExternalBlocked(t *testing.T) {
	profiles := testTrustedProfiles()
	store, err := NewFileFiatConversionStore(filepath.Join(t.TempDir(), "state.json"))
	require.NoError(t, err)
	base := FiatConversionOrchestratorConfig{Enabled: true, Production: true, ProviderAddress: "provider", Store: store, Profiles: profiles, Query: &fiatQueryFake{}, Submitter: &observationSubmitterFake{}, DEX: fiatDEXFake{}, Custody: custodyFake{testOnly: true}, Offramp: bridgeFake{}, Destination: resolverFake{}, Compliance: resolverFake{}, WebhookEvents: &durableWebhookStub{}, WebhookBindings: &durableWebhookStub{}}
	_, err = NewFiatConversionOrchestrator(base)
	require.ErrorIs(t, err, ErrFiatOrchestratorBlocked)
	base.Custody = custodyFake{}
	profiles.DEX.State = dex.RouteEngineeringCompleteExternalBlocked
	_, err = NewFiatConversionOrchestrator(base)
	require.ErrorIs(t, err, ErrFiatOrchestratorBlocked)
}

type durableWebhookStub struct{}

func (*durableWebhookStub) Durable() bool { return true }
func (*durableWebhookStub) PutVerifiedWebhookEvent(context.Context, offramp.WebhookEvent) error {
	return nil
}
func (*durableWebhookStub) VerifiedWebhookEvents(context.Context, string) ([]offramp.WebhookEvent, error) {
	return nil, nil
}
func (*durableWebhookStub) ConsumeVerifiedWebhookEvent(context.Context, string, string, time.Time) error {
	return nil
}
func (*durableWebhookStub) PutWebhookBinding(context.Context, offramp.WebhookBinding) error {
	return nil
}

func TestFiatConversionProfileMismatchAndLeaseLossFailClosed(t *testing.T) {
	orchestrator, query, _, _ := newOrchestratorFixture(t)
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	query.params.FiatConversionDexProfileDigest = bytes.Repeat([]byte{9}, 32)
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	require.Error(t, err)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkFailed, stored.State)
	require.NoError(t, orchestrator.cfg.Lease.Release(context.Background(), orchestrator.leaseName, orchestrator.leaseToken))
	require.ErrorIs(t, orchestrator.ProcessDue(context.Background(), 1), ErrFiatConversionLeaseLost)
}

func TestFiatConversionRestartMarksSigningAmbiguous(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	profiles := testTrustedProfiles()
	store, err := NewFileFiatConversionStore(path)
	require.NoError(t, err)
	require.NoError(t, store.Open(context.Background()))
	_, _, err = store.PutIfAbsent(context.Background(), minimalFiatWorkItem(FiatWorkSigning))
	require.NoError(t, err)
	require.NoError(t, store.Close())
	query := &fiatQueryFake{records: map[string]*settlementv1.FiatConversionRecord{}}
	submitter := &observationSubmitterFake{ready: true}
	secondStore, err := NewFileFiatConversionStore(path)
	require.NoError(t, err)
	orchestrator, err := NewFiatConversionOrchestrator(FiatConversionOrchestratorConfig{Enabled: true, ProviderAddress: "provider", Store: secondStore, Profiles: profiles, Query: query, Submitter: submitter, DEX: fiatDEXFake{}, Custody: custodyFake{}, Offramp: bridgeFake{}, Destination: resolverFake{}, Compliance: resolverFake{}})
	require.NoError(t, err)
	require.NoError(t, orchestrator.Start(context.Background()))
	defer func() { require.NoError(t, orchestrator.Stop(context.Background())) }()
	loaded, err := secondStore.Get(context.Background(), "conversion")
	require.NoError(t, err)
	require.Equal(t, FiatWorkAmbiguous, loaded.State)
}

func TestFiatConversionReadinessAndMetrics(t *testing.T) {
	orchestrator, query, _, _ := newOrchestratorFixture(t)
	_, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	readiness := orchestrator.Readiness(context.Background())
	require.True(t, readiness.Ready)
	require.Equal(t, 1, readiness.Metrics.QueueDepth)
	_, err = orchestrator.updateOwned(context.Background(), "conversion-1", func(work *FiatConversionWorkItem) error { work.State = FiatWorkPayoutAmbiguous; return nil })
	require.NoError(t, err)
	require.Equal(t, 1, orchestrator.Metrics(context.Background()).AmbiguousPayouts)
}

func TestFiatConversionFullHappyPathProducesExactlySixObservations(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	now := orchestrator.cfg.Now()
	quote := testDEXQuote(t, orchestrator.cfg.Profiles, now)
	blockHash := bytes.Repeat([]byte{3}, 32)
	finalityHash, err := CanonicalDEXFinalityHash(quote.ChainID, testSwapTxHash, 100, blockHash, 3, sdkmath.NewInt(99))
	require.NoError(t, err)
	orchestrator.cfg.DEX = fiatDEXFake{quote: quote, reconciliation: DEXSwapReconciliation{Found: true, Final: true, TxHash: testSwapTxHash, Height: 100, BlockHash: blockHash, Confirmations: 3, FinalityHash: finalityHash, OutputAmount: sdkmath.NewInt(99)}}
	bridge := &scriptedBridge{now: now}
	orchestrator.cfg.Offramp = bridge

	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	bridge.metadataResult = testPayoutResult(now, offramp.Quote{ID: "payout-quote", Provider: "partner"}, payoutMetadata(item), offramp.StatusCompleted)
	for index := 0; index < 8; index++ {
		err = orchestrator.process(context.Background(), item.Intent.ConversionID)
		if err != nil {
			var retry *fiatRetryError
			require.ErrorAs(t, err, &retry)
		}
		stored, getErr := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
		require.NoError(t, getErr)
		if stored.State == FiatWorkPayoutSubmitted {
			event := offramp.WebhookEvent{EventID: "happy-completed", EventType: "payout.status", Provider: "partner", APIVersion: orchestrator.cfg.Profiles.Payout.Webhook.Version, PayoutID: stored.PayoutID, QuoteID: stored.PayoutQuote.ID, CorrelationID: payoutMetadata(stored)["correlation_id"], Status: offramp.StatusCompleted, OccurredAt: now}
			require.NoError(t, orchestrator.cfg.WebhookEvents.PutVerifiedWebhookEvent(context.Background(), event))
		}
		if stored.State == FiatWorkCompleted {
			break
		}
	}
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkCompleted, stored.State)
	require.Len(t, submitter.submitted, 6)
	stages := []settlementv1.FiatConversionObservationStage{
		settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED,   // 1
		settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED,   // 2
		settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_FINALIZED,   // 3
		settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_QUOTED,    // 4
		settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_SUBMITTED, // 5
		settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_COMPLETED, // 6
	}
	for index, stage := range stages {
		require.Equal(t, uint64(index)+1, submitter.submitted[index].ObservationSequence) //nolint:gosec // six fixed stages.
		require.Equal(t, stage, submitter.submitted[index].Stage)
	}
	require.Equal(t, 1, bridge.quoteCalls)
	require.Equal(t, 1, bridge.payoutCalls)
	require.Equal(t, 1, bridge.statusCalls)
	raw, err := os.ReadFile(orchestrator.cfg.Store.(*FileFiatConversionStore).path)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "token-beneficiary")
}

func TestFiatConversionPayoutAmbiguityRecoversByMetadataWithoutSecondInitiation(t *testing.T) {
	orchestrator, query, _, _ := newOrchestratorFixture(t)
	now := orchestrator.cfg.Now()
	bridge := &scriptedBridge{now: now, initiateErr: &offramp.ProviderError{Kind: offramp.ErrorKindAmbiguous, Retryable: true, Ambiguous: true}}
	orchestrator.cfg.Offramp = bridge
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	bridge.metadataResult = testPayoutResult(now, offramp.Quote{ID: "payout-quote", Provider: "partner"}, payoutMetadata(item), offramp.StatusProcessing)
	_, err = orchestrator.updateOwned(context.Background(), item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State, work.SwapTxHash, work.ObservationSequence = FiatWorkPayoutQuote, testSwapTxHash, 4
		work.SwapConfirmation.StableAmount = "99"
		decision := orchestrator.cfg.Compliance.(resolverFake).decision
		quote := offramp.Quote{ID: "payout-quote", Provider: "partner", FiatAmount: sdkmath.LegacyNewDec(99), ExchangeRate: sdkmath.LegacyOneDec(), Fee: sdkmath.NewInt(1), CreatedAt: now, ExpiresAt: now.Add(time.Minute), Request: offramp.QuoteRequest{CryptoSymbol: "USDC", CryptoDenom: "uusdc", CryptoDecimals: 6, CryptoAmount: sdkmath.NewInt(99), FiatCurrency: "USD", PaymentMethod: "ach", Sender: "provider", Destination: "token-beneficiary", BeneficiaryReference: "token-beneficiary", Jurisdiction: "US", CorrelationID: conversionCorrelation(work.Intent.ConversionID), Compliance: decision}}
		quoteHash, requestHash, corridorID, providerBinding, commitmentErr := canonicalPayoutQuoteCommitments(quote, orchestrator.cfg.Profiles, work.Intent.ComplianceDigest)
		require.NoError(t, commitmentErr)
		work.PayoutQuote = payoutQuoteSnapshot(quote, requestHash, corridorID, providerBinding, work.Intent.ComplianceDigest, work.PayoutProfileDigest)
		work.PayoutQuoteDigest = hex.EncodeToString(quoteHash)
		return nil
	})
	require.NoError(t, err)
	query.records["conversion-1"].State = fiatChainStatePayoutPending
	query.records["conversion-1"].ObservationSequence = 4
	query.records["conversion-1"].OffRampQuoteId = "payout-quote"
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	query.records["conversion-1"].QuoteDigest = mustHex32(stored.PayoutQuoteDigest)
	query.records["conversion-1"].QuoteExpiry = now.Add(time.Minute).Unix()
	bridge.metadataResult.CryptoAmount = sdkmath.NewInt(99)
	bridge.metadataResult.Fee = sdkmath.NewInt(1)
	bridge.metadataResult.Provider = stored.PayoutQuote.Provider
	bridge.metadataResult.QuoteID = stored.PayoutQuote.ID
	bridge.metadataResult.Metadata = payoutMetadata(stored)
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	require.Error(t, err)
	stored, err = orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkPayoutAmbiguous, stored.State)
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, 1, bridge.payoutCalls)
	require.Equal(t, 1, bridge.metadataCalls)
}

func TestFiatConversionQueryUnavailableDoesNotClaimOrSucceed(t *testing.T) {
	orchestrator, query, _, _ := newOrchestratorFixture(t)
	query.err = ErrFiatConversionQueryUnavailable
	err := orchestrator.Poll(context.Background())
	require.ErrorIs(t, err, ErrFiatConversionQueryUnavailable)
	require.Zero(t, orchestrator.Metrics(context.Background()).QueueDepth)
}

func TestLoadTrustedFiatProfilesCanonicalDigestAndUnknownField(t *testing.T) {
	profiles := testTrustedProfiles()
	temporary := t.TempDir()
	dexPath, payoutPath := filepath.Join(temporary, "dex.json"), filepath.Join(temporary, "payout.json")
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	authority, err := NewEd25519FiatProfileAuthority("test-authority", publicKey)
	require.NoError(t, err)
	dexSignature := base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, fiatProfileAuthorizationPayload("dex", "test-authority", 1, profiles.DEXDigest)))
	payoutSignature := base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, fiatProfileAuthorizationPayload("payout", "test-authority", 1, profiles.PayoutDigest)))
	dexBytes, err := json.Marshal(VersionedDEXProfileFile{SchemaVersion: 1, Profile: profiles.DEX, AuthorityID: "test-authority", Signature: dexSignature})
	require.NoError(t, err)
	payoutBytes, err := json.Marshal(VersionedPayoutProfileFile{SchemaVersion: 1, Profile: profiles.Payout, AuthorityID: "test-authority", Signature: payoutSignature})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(dexPath, dexBytes, 0o600))
	require.NoError(t, os.WriteFile(payoutPath, payoutBytes, 0o600))
	_, err = LoadTrustedFiatProfiles(dexPath, payoutPath)
	require.Error(t, err, "certified files cannot self-authorize")
	loaded, err := LoadTrustedFiatProfilesWithAuthority(dexPath, payoutPath, authority)
	require.NoError(t, err)
	require.Equal(t, profiles.DEXDigest, loaded.DEXDigest)
	require.Equal(t, profiles.PayoutDigest, loaded.PayoutDigest)
	tampered := append(bytes.TrimSuffix(dexBytes, []byte("}")), []byte(`,"unknown":true}`)...)
	require.NoError(t, os.WriteFile(dexPath, tampered, 0o600))
	_, err = LoadTrustedFiatProfiles(dexPath, payoutPath)
	require.Error(t, err)
}

func TestFiatConversionComplianceRevocationEntersManualReview(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	orchestrator.cfg.Compliance = resolverFake{err: offramp.ErrComplianceRequired}
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	require.Error(t, err)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkFailed, stored.State)
	require.Len(t, submitter.submitted, 1)
	require.Equal(t, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_FAILED, submitter.submitted[0].Stage)
}

func TestFiatConversionDestinationUnavailableEntersManualReviewWithoutObservation(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	orchestrator.cfg.Destination = resolverFake{err: errors.New("destination vault unavailable")}
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	require.Error(t, err)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkFailed, stored.State)
	require.Len(t, submitter.submitted, 1)
	require.Equal(t, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_FAILED, submitter.submitted[0].Stage)
}

func TestFiatConversionSwapReorgEntersManualReview(t *testing.T) {
	orchestrator, query, _, _ := newOrchestratorFixture(t)
	quote := testDEXQuote(t, orchestrator.cfg.Profiles, orchestrator.cfg.Now())
	orchestrator.cfg.DEX = fiatDEXFake{quote: quote, reconciliation: DEXSwapReconciliation{Found: true, Reorged: true}}
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	_, err = orchestrator.updateOwned(context.Background(), item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State, work.DEXQuote, work.SwapTxHash, work.QuoteDigest = FiatWorkAmbiguous, &quote, testSwapTxHash, quote.QuoteDigest
		return nil
	})
	require.NoError(t, err)
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	require.Error(t, err)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkManualReview, stored.State)
}

func TestFiatConversionExpiredDEXQuoteReturnsToClaimedWithoutSigning(t *testing.T) {
	orchestrator, query, _, _ := newOrchestratorFixture(t)
	quote := testDEXQuote(t, orchestrator.cfg.Profiles, orchestrator.cfg.Now())
	quote.ExpiresAt = orchestrator.cfg.Now()
	quote.ID, _ = dex.QuoteDigest(quote)
	quote.QuoteDigest = quote.ID
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	_, err = orchestrator.updateOwned(context.Background(), item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State, work.DEXQuote, work.QuoteDigest = FiatWorkQuoteReported, &quote, quote.QuoteDigest
		return nil
	})
	require.NoError(t, err)
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	var retry *fiatRetryError
	require.ErrorAs(t, err, &retry)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkClaimed, stored.State)
	require.Empty(t, stored.SignedTxHash)
}

func TestFiatConversionGovernanceDisabledAfterQuoteStopsSigningAndBroadcast(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	quote := testDEXQuote(t, orchestrator.cfg.Profiles, orchestrator.cfg.Now())
	dexBoundary := &sideEffectDEXFake{quote: quote}
	custodyBoundary := &sideEffectCustodyFake{}
	orchestrator.cfg.DEX = dexBoundary
	orchestrator.cfg.Custody = custodyBoundary
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	require.NoError(t, orchestrator.process(context.Background(), item.Intent.ConversionID))
	require.Len(t, submitter.submitted, 1)
	query.params.FiatConversionEnabled = false
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	var terminal *fiatTerminalError
	require.ErrorAs(t, err, &terminal)
	require.Zero(t, custodyBoundary.signCalls)
	require.Zero(t, dexBoundary.executeCalls)
	require.Len(t, submitter.submitted, 1, "a governance stop must not race a chain observation")
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkManualReview, stored.State)
	require.Equal(t, "GOVERNANCE_DISABLED", stored.FailureCode)
}

func TestFiatConversionFinancialHoldAfterQuoteStopsSigningAndBroadcast(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	quote := testDEXQuote(t, orchestrator.cfg.Profiles, orchestrator.cfg.Now())
	dexBoundary := &sideEffectDEXFake{quote: quote}
	custodyBoundary := &sideEffectCustodyFake{}
	orchestrator.cfg.DEX = dexBoundary
	orchestrator.cfg.Custody = custodyBoundary
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	require.NoError(t, orchestrator.process(context.Background(), item.Intent.ConversionID))
	query.activeCaseID, query.activeHoldCount = "case-after-quote", 1
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	var terminal *fiatTerminalError
	require.ErrorAs(t, err, &terminal)
	require.Zero(t, custodyBoundary.signCalls)
	require.Zero(t, dexBoundary.executeCalls)
	require.Len(t, submitter.submitted, 1)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkManualReview, stored.State)
	require.Equal(t, "FINANCIAL_HOLD_ACTIVE", stored.FailureCode)
}

func TestFiatConversionFinancialHoldAfterSigningStopsBroadcast(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	quote := testDEXQuote(t, orchestrator.cfg.Profiles, orchestrator.cfg.Now())
	dexBoundary := &sideEffectDEXFake{quote: quote}
	custodyBoundary := &sideEffectCustodyFake{afterSign: func() {
		query.activeCaseID, query.activeHoldCount = "case-after-signing", 1
	}}
	orchestrator.cfg.DEX = dexBoundary
	orchestrator.cfg.Custody = custodyBoundary
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	require.NoError(t, orchestrator.process(context.Background(), item.Intent.ConversionID))
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	var terminal *fiatTerminalError
	require.ErrorAs(t, err, &terminal)
	require.Equal(t, 1, custodyBoundary.signCalls)
	require.Zero(t, dexBoundary.executeCalls, "fresh hold read must stop target-chain broadcast")
	require.Len(t, submitter.submitted, 1)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkManualReview, stored.State)
	require.NotEmpty(t, stored.SignedTxHash)
	require.Equal(t, "FINANCIAL_HOLD_ACTIVE", stored.FailureCode)
}

func TestFiatConversionAuthorizationQueryFailureAfterQuoteHasNoSideEffectAndRetries(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	quote := testDEXQuote(t, orchestrator.cfg.Profiles, orchestrator.cfg.Now())
	dexBoundary := &sideEffectDEXFake{quote: quote}
	custodyBoundary := &sideEffectCustodyFake{}
	orchestrator.cfg.DEX = dexBoundary
	orchestrator.cfg.Custody = custodyBoundary
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	require.NoError(t, orchestrator.process(context.Background(), item.Intent.ConversionID))
	query.authorizationErr = ErrFiatConversionQueryUnavailable
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	var retry *fiatRetryError
	require.ErrorAs(t, err, &retry)
	require.Zero(t, custodyBoundary.signCalls)
	require.Zero(t, dexBoundary.executeCalls)
	require.Len(t, submitter.submitted, 1)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkQuoteReported, stored.State)
	require.Equal(t, "AUTHORIZATION_QUERY_UNAVAILABLE", stored.FailureCode)
}

func TestFiatConversionFinancialHoldAfterSwapStopsPayoutInitiation(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	now := orchestrator.cfg.Now()
	bridge := &scriptedBridge{now: now}
	orchestrator.cfg.Offramp = bridge
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	query.records["conversion-1"].State = fiatChainStateSwapSettled
	query.records["conversion-1"].SwapTxHash = testSwapTxHash
	item, err = orchestrator.updateOwned(context.Background(), item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State = FiatWorkSwapFinalized
		work.SwapTxHash = testSwapTxHash
		work.SwapConfirmation.StableAmount = "99"
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, orchestrator.process(context.Background(), item.Intent.ConversionID))
	require.Equal(t, 1, bridge.quoteCalls)
	require.Zero(t, bridge.payoutCalls)
	require.Len(t, submitter.submitted, 1)
	query.activeCaseID, query.activeHoldCount = "case-after-swap", 1
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	var terminal *fiatTerminalError
	require.ErrorAs(t, err, &terminal)
	require.Zero(t, bridge.payoutCalls)
	require.Len(t, submitter.submitted, 1)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkManualReview, stored.State)
}

func TestFiatConversionExpiredAcceptedQuoteAppendsReplacementBeforeSigning(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	now := orchestrator.cfg.Now()
	oldQuote := testDEXQuote(t, orchestrator.cfg.Profiles, now.Add(-2*time.Minute))
	oldQuote.ExpiresAt = now.Add(-time.Minute)
	oldQuote.ID, _ = dex.QuoteDigest(oldQuote)
	oldQuote.QuoteDigest = oldQuote.ID
	newQuote := testDEXQuote(t, orchestrator.cfg.Profiles, now)
	dexBoundary := &sideEffectDEXFake{quote: newQuote}
	custodyBoundary := &sideEffectCustodyFake{}
	orchestrator.cfg.DEX = dexBoundary
	orchestrator.cfg.Custody = custodyBoundary
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	oldDigest := mustHex32(oldQuote.QuoteDigest)
	query.records["conversion-1"].State = "SWAP_PENDING"
	query.records["conversion-1"].ObservationSequence = 1
	query.records["conversion-1"].LastObservationDigest = bytes.Repeat([]byte{0x45}, 32)
	query.records["conversion-1"].QuoteDigest = oldDigest
	query.records["conversion-1"].QuoteExpiry = oldQuote.ExpiresAt.Unix()
	query.records["conversion-1"].MinimumStableOutput = sdk.NewCoin("uusdc", oldQuote.MinOutputAmount)
	item, err = orchestrator.updateOwned(context.Background(), item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State = FiatWorkQuoteReported
		work.DEXQuote = &oldQuote
		work.QuoteDigest = oldQuote.QuoteDigest
		work.ObservationSequence = 1
		work.ObservationDigest = hex.EncodeToString(query.records["conversion-1"].LastObservationDigest)
		return nil
	})
	require.NoError(t, err)
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	var retry *fiatRetryError
	require.ErrorAs(t, err, &retry)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkClaimed, stored.State)
	require.Empty(t, stored.SignedTxHash)
	require.Zero(t, custodyBoundary.signCalls)
	require.Zero(t, dexBoundary.executeCalls)
	_, err = orchestrator.updateOwned(context.Background(), item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.NextRetryAt = now
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, orchestrator.process(context.Background(), item.Intent.ConversionID))
	require.Len(t, submitter.submitted, 1)
	require.Equal(t, uint64(2), submitter.submitted[0].ObservationSequence)
	require.Equal(t, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED, submitter.submitted[0].Stage)
	require.NotEqual(t, oldDigest, submitter.submitted[0].QuoteDigest)
	require.Zero(t, custodyBoundary.signCalls)
	require.Zero(t, dexBoundary.executeCalls)
}

func TestFiatConversionRestartRecoveryAtEveryMajorStage(t *testing.T) {
	tests := []struct {
		state FiatConversionWorkState
		want  FiatConversionWorkState
	}{
		{FiatWorkClaimed, FiatWorkClaimed}, {FiatWorkQuoting, FiatWorkQuoting},
		{FiatWorkQuoteReported, FiatWorkQuoteReported}, {FiatWorkSigning, FiatWorkAmbiguous},
		{FiatWorkSwapBroadcast, FiatWorkAmbiguous}, {FiatWorkAmbiguous, FiatWorkAmbiguous},
		{FiatWorkSwapFinalized, FiatWorkSwapFinalized},
		{FiatWorkPayoutQuote, FiatWorkPayoutAmbiguous}, {FiatWorkPayoutSubmitted, FiatWorkPayoutAmbiguous}, {FiatWorkPayoutAmbiguous, FiatWorkPayoutAmbiguous},
	}
	for _, test := range tests {
		t.Run(string(test.state), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "state.json")
			store, err := NewFileFiatConversionStore(path)
			require.NoError(t, err)
			require.NoError(t, store.Open(context.Background()))
			_, _, err = store.PutIfAbsent(context.Background(), minimalFiatWorkItem(test.state))
			require.NoError(t, err)
			require.NoError(t, store.Close())
			reopened, err := NewFileFiatConversionStore(path)
			require.NoError(t, err)
			orchestrator, err := NewFiatConversionOrchestrator(FiatConversionOrchestratorConfig{Enabled: true, ProviderAddress: "provider", Store: reopened, Profiles: testTrustedProfiles(), Query: &fiatQueryFake{records: map[string]*settlementv1.FiatConversionRecord{}}, Submitter: &observationSubmitterFake{ready: true}, DEX: fiatDEXFake{}, Custody: custodyFake{}, Offramp: bridgeFake{}, Destination: resolverFake{}, Compliance: resolverFake{}})
			require.NoError(t, err)
			require.NoError(t, orchestrator.Start(context.Background()))
			loaded, err := reopened.Get(context.Background(), "conversion")
			require.NoError(t, err)
			require.Equal(t, test.want, loaded.State)
			require.NoError(t, orchestrator.Stop(context.Background()))
		})
	}
}

func TestFiatConversionFinalityBelowMinimumEntersManualReview(t *testing.T) {
	orchestrator, query, _, _ := newOrchestratorFixture(t)
	quote := testDEXQuote(t, orchestrator.cfg.Profiles, orchestrator.cfg.Now())
	blockHash := bytes.Repeat([]byte{3}, 32)
	output := quote.MinOutputAmount.SubRaw(1)
	finalityHash, hashErr := CanonicalDEXFinalityHash(quote.ChainID, testSwapTxHash, 100, blockHash, 3, output)
	require.NoError(t, hashErr)
	orchestrator.cfg.DEX = fiatDEXFake{quote: quote, reconciliation: DEXSwapReconciliation{Found: true, Final: true, TxHash: testSwapTxHash, Height: 100, BlockHash: blockHash, Confirmations: 3, FinalityHash: finalityHash, OutputAmount: output}}
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	_, err = orchestrator.updateOwned(context.Background(), item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State, work.DEXQuote, work.SwapTxHash, work.QuoteDigest = FiatWorkAmbiguous, &quote, testSwapTxHash, quote.QuoteDigest
		return nil
	})
	require.NoError(t, err)
	err = orchestrator.process(context.Background(), item.Intent.ConversionID)
	require.Error(t, err)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkManualReview, stored.State)
}

func TestFiatConversionReadinessFailsWhenMutationSubmitterUnavailable(t *testing.T) {
	orchestrator, _, submitter, _ := newOrchestratorFixture(t)
	submitter.ready = false
	readiness := orchestrator.Readiness(context.Background())
	require.False(t, readiness.Ready)
	require.Equal(t, "mutation submitter unavailable", readiness.Reason)
}

type custodyBackendFake struct{ calls int }

func (b *custodyBackendFake) Address(context.Context, string) (string, error) {
	return "osmo1custody", nil
}
func (b *custodyBackendFake) SignTxRaw(context.Context, DEXCustodySignRequest) ([]byte, error) {
	b.calls++
	return []byte("target-chain-txraw"), nil
}
func (*custodyBackendFake) VerifyTxRaw(context.Context, []byte, []byte) error { return nil }
func (*custodyBackendFake) RecoverTxRaw(context.Context, DEXCustodySignRequest, string) ([]byte, error) {
	return []byte("target-chain-txraw"), nil
}
func (*custodyBackendFake) ProductionReady(context.Context) error { return nil }
func (*custodyBackendFake) TestOnly() bool                        { return false }

func TestKeyManagerDEXCustodySignerRejectsTamperedExecutionPayload(t *testing.T) {
	keyConfig := DefaultKeyManagerConfig()
	keyConfig.StorageType = KeyStorageTypeMemory
	manager, err := NewKeyManager(keyConfig)
	require.NoError(t, err)
	require.NoError(t, manager.Unlock(""))
	_, err = manager.GenerateKey("provider")
	require.NoError(t, err)
	backend := &custodyBackendFake{}
	signer, err := NewKeyManagerDEXCustodySigner(manager, backend)
	require.NoError(t, err)
	quote := testDEXQuote(t, testTrustedProfiles(), time.Unix(1_700_000_000, 0).UTC())
	payload, err := dex.BuildExecutionPayload(quote)
	require.NoError(t, err)
	payload[len(payload)-1] ^= 1
	_, err = signer.SignExecution(context.Background(), quote, payload)
	require.ErrorIs(t, err, dex.ErrExecutionPayload)
	require.Zero(t, backend.calls)
}

func testDEXQuote(t *testing.T, profiles *TrustedFiatProfiles, now time.Time) dex.SwapQuote {
	t.Helper()
	request := dex.SwapRequest{FromToken: dex.Token{Symbol: "UVE", Denom: "uve", Decimals: 6, ChainID: profiles.DEX.ChainID}, ToToken: dex.Token{Symbol: "USDC", Denom: "uusdc", Decimals: 6, ChainID: profiles.DEX.ChainID}, Amount: sdkmath.NewInt(100), Type: dex.SwapTypeExactIn, SlippageTolerance: .01, SlippageToleranceExact: sdkmath.LegacyMustNewDecFromStr("0.01"), Deadline: now.Add(time.Minute), Sender: testCustodyAddr, Recipient: testCustodyAddr}
	evidence := dex.PoolStateEvidence{ChainID: profiles.DEX.ChainID, ProfileID: profiles.DEX.ID, PoolID: "1", Height: 90, BlockHash: strings.Repeat("a", 64), ObservedAt: now, FromDenom: "uve", ToDenom: "uusdc", FromDecimals: 6, ToDecimals: 6, ReserveIn: sdkmath.NewInt(1000), ReserveOut: sdkmath.NewInt(1000), SwapFee: sdkmath.LegacyMustNewDecFromStr("0.003"), StateDigest: strings.Repeat("b", 64)}
	quote := dex.SwapQuote{Request: request, Route: dex.SwapRoute{Hops: []dex.SwapHop{{PoolID: "1", DEX: "osmosis", FromToken: request.FromToken, ToToken: request.ToToken, AmountIn: sdkmath.NewInt(100), AmountOut: sdkmath.NewInt(99), Fee: evidence.SwapFee}}}, InputAmount: sdkmath.NewInt(100), OutputAmount: sdkmath.NewInt(99), MinOutputAmount: sdkmath.NewInt(98), Rate: sdkmath.LegacyMustNewDecFromStr("0.99"), PriceImpactExact: sdkmath.LegacyMustNewDecFromStr("0.01"), ProfileID: profiles.DEX.ID, ChainID: profiles.DEX.ChainID, DEX: "osmosis", DEXVersion: profiles.DEX.Version, PoolStateEvidence: []dex.PoolStateEvidence{evidence}, OraclePrice: sdkmath.LegacyOneDec(), OracleDeviation: sdkmath.LegacyMustNewDecFromStr("0.01"), ObservationHeight: 90, ObservationBlockHash: evidence.BlockHash, ExpiresAt: now.Add(time.Minute), CreatedAt: now}
	var err error
	quote.ID, err = dex.QuoteDigest(quote)
	require.NoError(t, err)
	quote.QuoteDigest = quote.ID
	return quote
}

func minimalFiatWorkItem(state FiatConversionWorkState) *FiatConversionWorkItem {
	now := time.Now().UTC()
	digest := strings.Repeat("01", sha256.Size)
	return &FiatConversionWorkItem{
		SchemaVersion: fiatConversionStoreSchemaVersion,
		Intent:        FiatConversionIntentSnapshot{ConversionID: "conversion", Provider: "provider", RequestDigest: digest, ComplianceDigest: digest},
		State:         state, DEXProfileDigest: digest, PayoutProfileDigest: digest, CreatedAt: now, UpdatedAt: now,
	}
}

type webhookSecrets struct{ secret []byte }

func (s webhookSecrets) ResolveWebhookSecret(context.Context, string) ([]byte, error) {
	return append([]byte(nil), s.secret...), nil
}

type webhookProfileAuthority struct{}

func (webhookProfileAuthority) AuthorizePayoutProfile(offramp.PayoutProfile) error { return nil }

func TestFiatWebhookHandlerVerifiesPersistsAndRejectsConflict(t *testing.T) {
	orchestrator, _, _, repository := newOrchestratorFixture(t)
	profile := orchestrator.cfg.Profiles.Payout
	secret := bytes.Repeat([]byte{7}, 32)
	binding := offramp.WebhookBinding{Provider: profile.Provider, PayoutID: "payout-1", QuoteID: "quote-1", CorrelationID: "correlation-1", ReservationDay: "2026-07-23"}
	require.NoError(t, repository.PutWebhookBinding(context.Background(), binding))
	now := time.Unix(1_700_000_000, 0).UTC()
	verifier, err := offramp.NewWebhookVerifier(offramp.WebhookVerifierConfig{Profile: profile, Secrets: webhookSecrets{secret}, Replay: repository, Bindings: repository, Authorizer: webhookProfileAuthority{}, Clock: func() time.Time { return now }})
	require.NoError(t, err)
	handler, err := NewFiatWebhookHandler(FiatWebhookHandlerConfig{Verifier: verifier, Events: repository, Orchestrator: orchestrator, MaxBodyBytes: 4096, Timeout: time.Second})
	require.NoError(t, err)
	event := offramp.WebhookEvent{EventID: "event-1", EventType: "payout.status", Provider: profile.Provider, APIVersion: profile.Webhook.Version, PayoutID: binding.PayoutID, QuoteID: binding.QuoteID, CorrelationID: binding.CorrelationID, Status: offramp.StatusCompleted, OccurredAt: now}
	body, err := json.Marshal(event)
	require.NoError(t, err)
	request := signedWebhookRequest(t, body, secret, now, profile.Webhook.Version)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	require.Equal(t, http.StatusAccepted, response.Code)
	events, err := repository.VerifiedWebhookEvents(context.Background(), binding.PayoutID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	event.Status = offramp.StatusFailed
	body, err = json.Marshal(event)
	require.NoError(t, err)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, signedWebhookRequest(t, body, secret, now, profile.Webhook.Version))
	require.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestFiatPendingObservationCrashAfterChainCommitConvergesWithoutResubmit(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	msg := orchestrator.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_FAILED)
	msg.Status, msg.FailureCode, msg.EvidenceHash = string(FiatWorkFailed), "INJECTED_SAFE_FAILURE", bytes.Repeat([]byte{9}, 32)
	raw, err := proto.Marshal(msg)
	require.NoError(t, err)
	digest := sha256.Sum256(raw)
	item, err = orchestrator.updateOwned(context.Background(), item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.PendingObservation, work.PendingNextState, work.ObservationDigest = msg, FiatWorkFailed, hex.EncodeToString(digest[:])
		return nil
	})
	require.NoError(t, err)
	applyTestObservation(query.records[item.Intent.ConversionID], msg)
	require.NoError(t, orchestrator.reconcilePendingObservation(context.Background(), item, query.records[item.Intent.ConversionID]))
	require.Empty(t, submitter.submitted)
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkFailed, stored.State)
	require.Nil(t, stored.PendingObservation)
}

func TestFiatPendingObservationConflictingHigherSequenceNeverResubmits(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	msg := orchestrator.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_FAILED)
	msg.Status, msg.FailureCode, msg.EvidenceHash = string(FiatWorkFailed), "INJECTED_SAFE_FAILURE", bytes.Repeat([]byte{8}, 32)
	raw, err := proto.Marshal(msg)
	require.NoError(t, err)
	digest := sha256.Sum256(raw)
	item, err = orchestrator.updateOwned(context.Background(), item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.PendingObservation, work.PendingNextState, work.ObservationDigest = msg, FiatWorkFailed, hex.EncodeToString(digest[:])
		return nil
	})
	require.NoError(t, err)
	query.records[item.Intent.ConversionID].ObservationSequence = msg.ObservationSequence + 1
	query.records[item.Intent.ConversionID].LastObservationDigest = bytes.Repeat([]byte{0xAA}, 32)
	err = orchestrator.reconcilePendingObservation(context.Background(), item, query.records[item.Intent.ConversionID])
	var terminal *fiatTerminalError
	require.ErrorAs(t, err, &terminal)
	require.Empty(t, submitter.submitted)
}

func TestFiatSwapUnknownHashWithoutDeterministicRecoveryStaysManual(t *testing.T) {
	orchestrator, query, submitter, _ := newOrchestratorFixture(t)
	quote := testDEXQuote(t, orchestrator.cfg.Profiles, orchestrator.cfg.Now())
	orchestrator.cfg.DEX = fiatDEXFake{reconciliation: DEXSwapReconciliation{Found: false}}
	orchestrator.cfg.Custody = custodyFake{err: ErrDEXCustodyUnavailable}
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	item, err = orchestrator.updateOwned(context.Background(), item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State, work.DEXQuote, work.QuoteDigest = FiatWorkAmbiguous, &quote, quote.QuoteDigest
		work.PayloadHash, work.SignedTxHash = digestHex(mustExecutionPayload(t, quote)), strings.Repeat("ab", 32)
		return nil
	})
	require.NoError(t, err)
	err = orchestrator.reconcileSwap(context.Background(), item, query.records[item.Intent.ConversionID])
	var terminal *fiatTerminalError
	require.ErrorAs(t, err, &terminal)
	require.Empty(t, submitter.submitted, "uncertain external outcome must not be overwritten")
	stored, err := orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	require.NoError(t, err)
	require.Equal(t, FiatWorkManualReview, stored.State)
}

func TestFiatPayoutQuoteSnapshotTamperDetectedBeforeSideEffect(t *testing.T) {
	orchestrator, query, _, _ := newOrchestratorFixture(t)
	now := orchestrator.cfg.Now()
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	decision := orchestrator.cfg.Compliance.(resolverFake).decision
	quote := offramp.Quote{ID: "tamper-quote", Provider: "partner", FiatAmount: sdkmath.LegacyNewDec(99), ExchangeRate: sdkmath.LegacyOneDec(), Fee: sdkmath.NewInt(1), CreatedAt: now, ExpiresAt: now.Add(time.Minute), Request: offramp.QuoteRequest{CryptoSymbol: "USDC", CryptoDenom: "uusdc", CryptoDecimals: 6, CryptoAmount: sdkmath.NewInt(99), FiatCurrency: "USD", PaymentMethod: "ach", Sender: "provider", Destination: "token-beneficiary", BeneficiaryReference: "token-beneficiary", Jurisdiction: "US", CorrelationID: conversionCorrelation(item.Intent.ConversionID), Compliance: decision}}
	quoteHash, requestHash, corridorID, providerBinding, err := canonicalPayoutQuoteCommitments(quote, orchestrator.cfg.Profiles, item.Intent.ComplianceDigest)
	require.NoError(t, err)
	item.PayoutQuote = payoutQuoteSnapshot(quote, requestHash, corridorID, providerBinding, item.Intent.ComplianceDigest, item.PayoutProfileDigest)
	item.PayoutQuote.FiatAmount = sdkmath.LegacyNewDec(100).String()
	item.PayoutQuoteDigest = hex.EncodeToString(quoteHash)
	item.SwapConfirmation.StableAmount = "99"
	_, err = reconstructPayoutQuote(item.PayoutQuote, query.records[item.Intent.ConversionID], item, "token-beneficiary", decision, orchestrator.cfg.Profiles)
	require.ErrorIs(t, err, offramp.ErrInvalidRequest)
}

func TestFiatWebhookCompletionConsumedExactlyOnce(t *testing.T) {
	repository, err := NewFileFiatRepository(filepath.Join(t.TempDir(), "repository.json"))
	require.NoError(t, err)
	require.NoError(t, repository.Open(context.Background()))
	defer repository.Close()
	event := offramp.WebhookEvent{EventID: "consume-1", EventType: "payout.status", Provider: "partner", APIVersion: "1", PayoutID: "payout-1", QuoteID: "quote-1", CorrelationID: "correlation-1", Status: offramp.StatusCompleted, OccurredAt: time.Now().UTC()}
	require.NoError(t, repository.PutVerifiedWebhookEvent(context.Background(), event))
	require.NoError(t, repository.ConsumeVerifiedWebhookEvent(context.Background(), event.Provider, event.EventID, time.Now().UTC()))
	require.NoError(t, repository.ConsumeVerifiedWebhookEvent(context.Background(), event.Provider, event.EventID, time.Now().UTC()))
	events, err := repository.VerifiedWebhookEvents(context.Background(), event.PayoutID)
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestFileFiatRepositoryRejectsTrailingJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "repository.json")
	repository, err := NewFileFiatRepository(path)
	require.NoError(t, err)
	require.NoError(t, repository.Open(context.Background()))
	require.NoError(t, repository.PutWebhookBinding(context.Background(), offramp.WebhookBinding{Provider: "partner", PayoutID: "payout-1", QuoteID: "quote-1", CorrelationID: "correlation-1", ReservationDay: "2026-07-23"}))
	require.NoError(t, repository.Close())

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0) // #nosec G304 -- test-owned temporary path.
	require.NoError(t, err)
	_, err = file.WriteString("\n{}")
	require.NoError(t, err)
	require.NoError(t, file.Close())

	reopened, err := NewFileFiatRepository(path)
	require.NoError(t, err)
	require.ErrorContains(t, reopened.Open(context.Background()), "multiple JSON values")
}

func TestFiatWebhookHandlerRejectsSmugglingAndDuplicateHeaders(t *testing.T) {
	orchestrator, _, _, repository := newOrchestratorFixture(t)
	profile := orchestrator.cfg.Profiles.Payout
	secret := bytes.Repeat([]byte{7}, 32)
	binding := offramp.WebhookBinding{Provider: profile.Provider, PayoutID: "payout-smuggle", QuoteID: "quote-smuggle", CorrelationID: "correlation-smuggle", ReservationDay: "2026-07-23"}
	require.NoError(t, repository.PutWebhookBinding(context.Background(), binding))
	now := orchestrator.cfg.Now()
	verifier, err := offramp.NewWebhookVerifier(offramp.WebhookVerifierConfig{Profile: profile, Secrets: webhookSecrets{secret}, Replay: repository, Bindings: repository, Authorizer: webhookProfileAuthority{}, Clock: func() time.Time { return now }})
	require.NoError(t, err)
	handler, err := NewFiatWebhookHandler(FiatWebhookHandlerConfig{Verifier: verifier, Events: repository, Orchestrator: orchestrator, MaxBodyBytes: 4096})
	require.NoError(t, err)
	event := offramp.WebhookEvent{EventID: "smuggle-1", EventType: "payout.status", Provider: profile.Provider, APIVersion: profile.Webhook.Version, PayoutID: binding.PayoutID, QuoteID: binding.QuoteID, CorrelationID: binding.CorrelationID, Status: offramp.StatusProcessing, OccurredAt: now}
	body, err := json.Marshal(event)
	require.NoError(t, err)
	request := signedWebhookRequest(t, body, secret, now, profile.Webhook.Version)
	request.Header.Add("X-Offramp-Signature", request.Header.Get("X-Offramp-Signature"))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	require.Equal(t, http.StatusUnauthorized, response.Code)
	request = signedWebhookRequest(t, body, secret, now, profile.Webhook.Version)
	request.TransferEncoding = []string{"chunked"}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	require.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestFiatRetryCountStopsExactlyAtMaximum(t *testing.T) {
	orchestrator, query, _, _ := newOrchestratorFixture(t)
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	orchestrator.cfg.MaxAttempts = 3
	for expected := uint32(1); expected <= orchestrator.cfg.MaxAttempts; expected++ {
		err = orchestrator.retry(context.Background(), item, "TRANSIENT", context.DeadlineExceeded)
		if expected < orchestrator.cfg.MaxAttempts {
			var retry *fiatRetryError
			require.ErrorAs(t, err, &retry)
		} else {
			var terminal *fiatTerminalError
			require.ErrorAs(t, err, &terminal)
		}
		item, _ = orchestrator.cfg.Store.Get(context.Background(), item.Intent.ConversionID)
	}
	require.Equal(t, uint32(2), item.AttemptCount, "terminal safe-failure observation does not invent an external retry")
	require.Equal(t, FiatWorkFailed, item.State)
}

func TestFiatOrchestratorStartStopRestartNoLeak(t *testing.T) {
	orchestrator, _, _, _ := newOrchestratorFixture(t)
	require.NoError(t, orchestrator.Stop(context.Background()))
	require.NoError(t, orchestrator.Stop(context.Background()))
	require.NoError(t, orchestrator.Start(context.Background()))
	require.NoError(t, orchestrator.Start(context.Background()))
	require.NoError(t, orchestrator.Stop(context.Background()))
}

func TestFiatStateFilesContainNoPIIOrSecrets(t *testing.T) {
	orchestrator, query, _, repository := newOrchestratorFixture(t)
	item, err := orchestrator.Claim(context.Background(), query.records["conversion-1"])
	require.NoError(t, err)
	secretValues := []string{"4111111111111111", "DE89370400440532013000", "routing-021000021", "api-secret-value", "beneficiary@example.com"}
	payout := testPayoutResult(orchestrator.cfg.Now(), offramp.Quote{ID: "privacy-quote", Provider: "partner"}, payoutMetadata(item), offramp.StatusProcessing)
	payout.Reference = "safe-reference"
	require.NoError(t, repository.PutPayout(context.Background(), payout))
	for _, path := range []string{orchestrator.cfg.Store.(*FileFiatConversionStore).path, repository.path} {
		raw, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		for _, secret := range secretValues {
			require.NotContains(t, string(raw), secret)
		}
		require.NotContains(t, strings.ToLower(string(raw)), "account_number")
		require.NotContains(t, strings.ToLower(string(raw)), "beneficiary_reference")
	}
}

func mustExecutionPayload(t *testing.T, quote dex.SwapQuote) []byte {
	t.Helper()
	payload, err := dex.BuildExecutionPayload(quote)
	require.NoError(t, err)
	return payload
}

func signedWebhookRequest(t *testing.T, body, secret []byte, now time.Time, version string) *http.Request {
	t.Helper()
	timestamp := strconv.FormatInt(now.Unix(), 10)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(append(append([]byte(timestamp), '.'), body...))
	request := httptest.NewRequest(http.MethodPost, "http://localhost/internal", bytes.NewReader(body))
	request.Header.Set("X-Offramp-Timestamp", timestamp)
	request.Header.Set("X-Offramp-Signature", hex.EncodeToString(mac.Sum(nil)))
	request.Header.Set("X-Offramp-Key-ID", "key")
	request.Header.Set("X-Offramp-Key-Version", "1")
	request.Header.Set("X-Offramp-API-Version", version)
	request.Header.Set("Content-Type", "application/json")
	return request
}
