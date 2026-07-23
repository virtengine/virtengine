// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

var (
	ErrFiatOrchestratorBlocked = errors.New("fiat conversion orchestrator externally blocked")
	ErrFiatIntentChanged       = errors.New("fiat conversion intent changed after claim")
	ErrFiatObservationPending  = errors.New("fiat conversion observation pending confirmation")
	ErrSwapOutcomeAmbiguous    = errors.New("DEX swap outcome ambiguous")
	ErrFiatGovernanceDisabled  = errors.New("fiat conversion disabled by governance")
	ErrFiatFinancialHold       = errors.New("fiat conversion lineage has an active financial hold")
	ErrFiatCurrentProfile      = errors.New("fiat conversion current profile changed")
	ErrFiatCurrentCompliance   = errors.New("fiat conversion current compliance authorization changed")
)

const fiatObservationStatusSubmitted = "submitted" //nolint:goconst // Fiat settlement wire status, not an HPC lifecycle event.

// FiatDestinationResolver decrypts or token-resolves a beneficiary only for an
// individual provider call. The returned string is never stored by the worker.
type FiatDestinationResolver interface {
	ResolveDestination(ctx context.Context, conversion *settlementv1.FiatConversionRecord) (ResolvedFiatDestination, error)
	ProductionReady(ctx context.Context) error
}

type ResolvedFiatDestination struct {
	Reference string
	Digest    []byte
}

// FiatComplianceResolver independently resolves the current off-ramp decision
// and binds it to the chain commitment. No identity evidence crosses this seam.
type FiatComplianceResolver interface {
	ResolveCompliance(ctx context.Context, conversion *settlementv1.FiatConversionRecord) (ResolvedFiatCompliance, error)
	ProductionReady(ctx context.Context) error
}

type ResolvedFiatCompliance struct {
	Decision offramp.ComplianceDecision
	Digest   []byte
}

// FiatObservationSubmitter is the mandatory Task 85A durable queue seam.
type FiatObservationSubmitter interface {
	SubmitFiatConversionObservation(ctx context.Context, msg *settlementv1.MsgRecordFiatConversionObservation) (ProviderMutationResult, error)
	Readiness(ctx context.Context) ProviderMutationReadiness
}

// FiatDEX executes and reconciles a signed exact quote. ReconcileSwap must
// return Found=false rather than treating an unavailable query as success.
type FiatDEX interface {
	GetSwapQuote(ctx context.Context, request dex.SwapRequest) (dex.SwapQuote, error)
	ExecuteSwap(ctx context.Context, quote dex.SwapQuote, signedEnvelope []byte) (dex.SwapResult, error)
	ReconcileSwap(ctx context.Context, quote dex.SwapQuote, txHash, correlationID string) (DEXSwapReconciliation, error)
}

type DEXSwapReconciliation struct {
	Found           bool
	Final           bool
	Reorged         bool
	TerminalFailure bool
	FailureCode     string
	TxHash          string
	Height          int64
	BlockHash       []byte
	Confirmations   uint32
	FinalityHash    []byte
	OutputAmount    sdkmath.Int
}

// AdapterFiatDEX adapts an existing DEX adapter plus an explicit finality query.
type AdapterFiatDEX struct {
	Adapter    dex.Adapter
	Reconciler interface {
		ReconcileSwap(context.Context, dex.SwapQuote, string, string) (DEXSwapReconciliation, error)
	}
}

func (a AdapterFiatDEX) GetSwapQuote(ctx context.Context, request dex.SwapRequest) (dex.SwapQuote, error) {
	if a.Adapter == nil {
		return dex.SwapQuote{}, dex.ErrProviderUnavailable
	}
	return a.Adapter.GetSwapQuote(ctx, request)
}
func (a AdapterFiatDEX) ExecuteSwap(ctx context.Context, quote dex.SwapQuote, signed []byte) (dex.SwapResult, error) {
	if a.Adapter == nil {
		return dex.SwapResult{}, dex.ErrProviderUnavailable
	}
	return a.Adapter.ExecuteSwap(ctx, quote, signed)
}
func (a AdapterFiatDEX) ReconcileSwap(ctx context.Context, quote dex.SwapQuote, txHash, correlationID string) (DEXSwapReconciliation, error) {
	if a.Reconciler == nil {
		return DEXSwapReconciliation{}, ErrFiatConversionQueryUnavailable
	}
	return a.Reconciler.ReconcileSwap(ctx, quote, txHash, correlationID)
}

// WebhookEventStore is the durable verified callback boundary.
type WebhookEventStore interface {
	PutVerifiedWebhookEvent(context.Context, offramp.WebhookEvent) error
	VerifiedWebhookEvents(context.Context, string) ([]offramp.WebhookEvent, error)
	ConsumeVerifiedWebhookEvent(context.Context, string, string, time.Time) error
	Durable() bool
}

// WebhookBindingStore persists immutable callback binding before payout state
// can be acknowledged.
type WebhookBindingStore interface {
	PutWebhookBinding(context.Context, offramp.WebhookBinding) error
	Durable() bool
}

// FiatConversionOrchestratorConfig is explicit dependency injection. The
// command defaults this subsystem to disabled and cannot synthesize custody or
// partner dependencies.
type FiatConversionOrchestratorConfig struct {
	Enabled                    bool
	Production                 bool
	EngineeringExternalBlocked bool
	ProviderAddress            string
	Store                      FiatConversionStore
	Lease                      FiatConversionLease
	LeaseTTL                   time.Duration
	PollInterval               time.Duration
	RetryBackoff               time.Duration
	MaxRetryBackoff            time.Duration
	MaxAttempts                uint32
	Now                        func() time.Time
	Profiles                   *TrustedFiatProfiles
	Query                      FiatConversionQuery
	Submitter                  FiatObservationSubmitter
	DEX                        FiatDEX
	Custody                    DEXCustodySigner
	Offramp                    offramp.Bridge
	Destination                FiatDestinationResolver
	Compliance                 FiatComplianceResolver
	WebhookEvents              WebhookEventStore
	WebhookBindings            WebhookBindingStore
}

func DefaultFiatConversionOrchestratorConfig() FiatConversionOrchestratorConfig {
	return FiatConversionOrchestratorConfig{
		Enabled: false, Production: true, LeaseTTL: 30 * time.Second,
		PollInterval: 10 * time.Second, RetryBackoff: 2 * time.Second,
		MaxRetryBackoff: 5 * time.Minute, MaxAttempts: 12,
		Now: func() time.Time { return time.Now().UTC() },
	}
}

type FiatConversionMetrics struct {
	QueueDepth       int
	OldestIntentAge  time.Duration
	AmbiguousSwaps   int
	AmbiguousPayouts int
	ManualReview     int
	DeadLetters      int
	FinalityLatency  time.Duration
	LastSuccess      time.Time
	LastFailure      time.Time
}

type FiatConversionReadiness struct {
	Enabled           bool
	Ready             bool
	Started           bool
	StoreReady        bool
	LeaseHeld         bool
	ProfilesReady     bool
	QueryReady        bool
	SubmitterReady    bool
	DEXReady          bool
	CustodyReady      bool
	OfframpReady      bool
	ResolversReady    bool
	WebhookReady      bool
	ExternallyBlocked bool
	Reason            string
	Metrics           FiatConversionMetrics
}

// FiatConversionOrchestrator owns each external conversion side effect.
type FiatConversionOrchestrator struct {
	cfg         FiatConversionOrchestratorConfig
	mu          sync.RWMutex
	processMu   sync.Mutex
	running     bool
	storeReady  bool
	leaseName   string
	leaseToken  uint64
	stopCh      chan struct{}
	wakeCh      chan struct{}
	wg          sync.WaitGroup
	lastSuccess time.Time
	lastFailure time.Time
}

func NewFiatConversionOrchestrator(cfg FiatConversionOrchestratorConfig) (*FiatConversionOrchestrator, error) {
	defaults := DefaultFiatConversionOrchestratorConfig()
	if !cfg.Enabled {
		return &FiatConversionOrchestrator{cfg: cfg}, nil
	}
	if cfg.ProviderAddress == "" || cfg.Store == nil || !cfg.Store.Durable() || cfg.Query == nil || cfg.Submitter == nil || cfg.Profiles == nil || cfg.DEX == nil || cfg.Custody == nil || cfg.Offramp == nil || cfg.Destination == nil || cfg.Compliance == nil {
		return nil, fmt.Errorf("%w: mandatory conversion dependency missing", ErrFiatOrchestratorBlocked)
	}
	if cfg.Lease == nil {
		cfg.Lease = NewLocalFiatConversionLease()
	}
	if cfg.LeaseTTL <= 0 {
		cfg.LeaseTTL = defaults.LeaseTTL
	}
	if cfg.LeaseTTL < 3*time.Nanosecond {
		return nil, fmt.Errorf("%w: lease TTL must be at least 3ns", ErrFiatOrchestratorBlocked)
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaults.PollInterval
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = defaults.RetryBackoff
	}
	if cfg.MaxRetryBackoff <= 0 {
		cfg.MaxRetryBackoff = defaults.MaxRetryBackoff
	}
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = defaults.MaxAttempts
	}
	if cfg.MaxAttempts > 32 || cfg.RetryBackoff > cfg.MaxRetryBackoff {
		return nil, fmt.Errorf("%w: retry bounds invalid", ErrFiatOrchestratorBlocked)
	}
	if cfg.Now == nil {
		cfg.Now = defaults.Now
	}
	if fileStore, ok := cfg.Store.(*FileFiatConversionStore); ok {
		fileStore.setClock(cfg.Now)
	}
	if cfg.Production {
		if cfg.EngineeringExternalBlocked || !cfg.Profiles.DEXTrusted || !cfg.Profiles.PayoutTrusted || cfg.Profiles.DEX.State != dex.RouteCertifiedEnabled || cfg.Profiles.Payout.State != offramp.ProfileCertifiedEnabled || cfg.Custody.TestOnly() {
			return nil, fmt.Errorf("%w: production requires certified profiles and non-test custody", ErrFiatOrchestratorBlocked)
		}
		if cfg.WebhookEvents == nil || !cfg.WebhookEvents.Durable() || cfg.WebhookBindings == nil || !cfg.WebhookBindings.Durable() {
			return nil, fmt.Errorf("%w: production webhook repositories are not durable", ErrFiatOrchestratorBlocked)
		}
	}
	return &FiatConversionOrchestrator{cfg: cfg, leaseName: "fiat-conversion:" + cfg.ProviderAddress, stopCh: make(chan struct{}), wakeCh: make(chan struct{}, 1)}, nil
}

func (o *FiatConversionOrchestrator) Start(ctx context.Context) error {
	if o == nil || !o.cfg.Enabled {
		return nil
	}
	o.mu.Lock()
	if o.running {
		o.mu.Unlock()
		return nil
	}
	if err := o.cfg.Store.Open(ctx); err != nil {
		o.mu.Unlock()
		return err
	}
	o.storeReady = true
	token, err := o.cfg.Lease.Acquire(ctx, o.leaseName, o.cfg.LeaseTTL)
	if err != nil {
		_ = o.cfg.Store.Close()
		o.storeReady = false
		o.mu.Unlock()
		return err
	}
	o.leaseToken = token
	if err := o.recoverOnStart(ctx); err != nil {
		_ = o.cfg.Lease.Release(ctx, o.leaseName, token)
		_ = o.cfg.Store.Close()
		o.storeReady = false
		o.mu.Unlock()
		return err
	}
	o.running = true
	o.stopCh = make(chan struct{})
	o.wakeCh = make(chan struct{}, 1)
	o.wg.Add(1)
	o.mu.Unlock()
	go o.worker(ctx)
	return nil
}

func (o *FiatConversionOrchestrator) Stop(ctx context.Context) error {
	if o == nil || !o.cfg.Enabled {
		return nil
	}
	o.mu.Lock()
	if !o.running {
		o.mu.Unlock()
		return nil
	}
	o.running = false
	close(o.stopCh)
	token := o.leaseToken
	o.mu.Unlock()
	done := make(chan struct{})
	go func() { o.wg.Wait(); close(done) }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	_ = o.cfg.Lease.Release(context.Background(), o.leaseName, token)
	err := o.cfg.Store.Close()
	o.mu.Lock()
	o.storeReady = false
	o.mu.Unlock()
	return err
}

func (o *FiatConversionOrchestrator) Wake() {
	if o == nil || !o.cfg.Enabled {
		return
	}
	select {
	case o.wakeCh <- struct{}{}:
	default:
	}
}

func (o *FiatConversionOrchestrator) worker(ctx context.Context) {
	defer o.wg.Done()
	ticker := time.NewTicker(o.cfg.PollInterval)
	leaseInterval := o.cfg.LeaseTTL / 3
	if leaseInterval <= 0 {
		leaseInterval = time.Nanosecond
	}
	leaseTicker := time.NewTicker(leaseInterval)
	defer ticker.Stop()
	defer leaseTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stopCh:
			return
		case <-o.wakeCh:
			_ = o.ProcessDue(ctx, 64)
		case <-ticker.C:
			_ = o.Poll(ctx)
		case <-leaseTicker.C:
			if err := o.cfg.Lease.Renew(ctx, o.leaseName, o.leaseToken, o.cfg.LeaseTTL); err != nil {
				o.recordFailure()
				return
			}
		}
	}
}

// Poll imports all returned nonterminal intents, then processes due work. Query
// failure returns an error and does not alter local work or report success.
func (o *FiatConversionOrchestrator) Poll(ctx context.Context) error {
	if err := o.ensureLease(ctx); err != nil {
		return err
	}
	records, err := o.cfg.Query.ListNonTerminalConversions(ctx, o.cfg.ProviderAddress)
	if err != nil {
		o.recordFailure()
		return err
	}
	for i := range records {
		if _, err := o.Claim(ctx, &records[i]); err != nil {
			return err
		}
	}
	return o.ProcessDue(ctx, 64)
}

func (o *FiatConversionOrchestrator) Claim(ctx context.Context, record *settlementv1.FiatConversionRecord) (*FiatConversionWorkItem, error) {
	if err := o.ensureLease(ctx); err != nil {
		return nil, err
	}
	if record == nil || record.ConversionId == "" || record.Provider != o.cfg.ProviderAddress || !chainFiatConversionProcessable(*record) {
		return nil, errors.New("invalid fiat conversion intent")
	}
	snapshot := snapshotFiatIntent(record)
	now := o.cfg.Now()
	item := &FiatConversionWorkItem{SchemaVersion: fiatConversionStoreSchemaVersion, Intent: snapshot, State: FiatWorkClaimed, DEXProfileID: o.cfg.Profiles.DEX.ID, DEXProfileDigest: o.cfg.Profiles.DEXDigestHex(), PayoutProfileID: o.cfg.Profiles.Payout.ID, PayoutProfileDigest: o.cfg.Profiles.PayoutDigestHex(), ObservationSequence: record.ObservationSequence, ObservationDigest: hex.EncodeToString(record.LastObservationDigest), CreatedAt: now, UpdatedAt: now, NextRetryAt: now, LeaseToken: o.leaseToken}
	stored, existed, err := o.cfg.Store.PutIfAbsent(ctx, item)
	if err != nil {
		return nil, err
	}
	if existed && !fiatIntentEqual(stored.Intent, snapshot) {
		return nil, ErrFiatIntentChanged
	}
	return stored, nil
}

func (o *FiatConversionOrchestrator) ProcessDue(ctx context.Context, limit int) error {
	o.processMu.Lock()
	defer o.processMu.Unlock()
	if err := o.ensureLease(ctx); err != nil {
		return err
	}
	items, err := o.cfg.Store.List(ctx)
	if err != nil {
		return err
	}
	now := o.cfg.Now()
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].Intent.ConversionID < items[j].Intent.ConversionID
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	processed := 0
	for _, item := range items {
		if item.State.terminal() || item.NextRetryAt.After(now) {
			continue
		}
		if limit > 0 && processed >= limit {
			break
		}
		processed++
		if err := o.process(ctx, item.Intent.ConversionID); err != nil && isFiatTerminalError(err) {
			return err
		}
	}
	return nil
}

func (o *FiatConversionOrchestrator) process(ctx context.Context, id string) error {
	if err := o.ensureLease(ctx); err != nil {
		return err
	}
	item, err := o.cfg.Store.Get(ctx, id)
	if err != nil {
		return err
	}
	chain, err := o.cfg.Query.GetConversion(ctx, id)
	if err != nil {
		return o.retry(ctx, item, "QUERY_UNAVAILABLE", err)
	}
	if item.PendingObservation != nil {
		if chain == nil || !fiatIntentEqual(item.Intent, snapshotFiatIntent(chain)) {
			return o.manualReview(ctx, item, "PENDING_OBSERVATION_INTENT_MISMATCH", ErrFiatIntentChanged)
		}
		return o.reconcilePendingObservation(ctx, item, chain)
	}
	if fiatWorkRequiresReadOnlyReconciliation(item, chain) {
		if err := o.validateImmutableIntentAndProfiles(item, chain); err != nil {
			return o.manualReview(ctx, item, "RECONCILIATION_COMMITMENT_MISMATCH", err)
		}
		if chain.ObservationSequence < item.ObservationSequence {
			return o.deferWithoutAttempt(ctx, item, "CHAIN_OBSERVATION_LAG")
		}
		if chain.ObservationSequence != item.ObservationSequence || !bytes.Equal(chain.LastObservationDigest, mustHex32(item.ObservationDigest)) {
			return o.manualReview(ctx, item, "OBSERVATION_HEAD_DIVERGED", ErrFiatIntentChanged)
		}
		switch item.State {
		case FiatWorkSwapBroadcast, FiatWorkAmbiguous:
			return o.reconcileSwap(ctx, item, chain)
		case FiatWorkPayoutSubmitted, FiatWorkPayoutAmbiguous:
			return o.reconcilePayout(ctx, item)
		}
	}
	chain, err = o.validateIntentAndProfiles(ctx, item, chain)
	if err != nil {
		if errors.Is(err, ErrFiatConversionQueryUnavailable) {
			return o.retry(ctx, item, "AUTHORIZATION_QUERY_UNAVAILABLE", err)
		}
		if currentFiatAuthorizationError(err) {
			return o.deferExecutionBlocked(ctx, item, executionBlockCode(err), err)
		}
		return o.manualReview(ctx, item, "PROFILE_OR_INTENT_MISMATCH", err)
	}
	if chain.ObservationSequence < item.ObservationSequence {
		return o.deferWithoutAttempt(ctx, item, "CHAIN_OBSERVATION_LAG")
	}
	if chain.ObservationSequence != item.ObservationSequence || !bytes.Equal(chain.LastObservationDigest, mustHex32(item.ObservationDigest)) {
		return o.manualReview(ctx, item, "OBSERVATION_HEAD_DIVERGED", ErrFiatIntentChanged)
	}
	switch item.State {
	case FiatWorkClaimed, FiatWorkQuoting:
		return o.quoteSwap(ctx, item, chain)
	case FiatWorkQuoteReported:
		return o.signAndExecuteSwap(ctx, item)
	case FiatWorkSigning, FiatWorkSwapBroadcast, FiatWorkAmbiguous:
		return o.reconcileSwap(ctx, item, chain)
	case FiatWorkSwapFinalized:
		return o.quotePayout(ctx, item, chain)
	case FiatWorkPayoutQuote:
		return o.initiatePayout(ctx, item, chain)
	case FiatWorkPayoutSubmitted, FiatWorkPayoutAmbiguous:
		return o.reconcilePayout(ctx, item)
	default:
		return nil
	}
}

func (o *FiatConversionOrchestrator) quoteSwap(ctx context.Context, item *FiatConversionWorkItem, chain *settlementv1.FiatConversionRecord) error {
	resolvedDestination, err := o.cfg.Destination.ResolveDestination(ctx, chain)
	if err != nil {
		return o.retry(ctx, item, "DESTINATION_UNAVAILABLE", err)
	}
	if err := validateResolvedDestination(chain, resolvedDestination); err != nil {
		return o.manualReview(ctx, item, "DESTINATION_BINDING_MISMATCH", err)
	}
	clearString(&resolvedDestination.Reference)
	resolvedCompliance, err := o.cfg.Compliance.ResolveCompliance(ctx, chain)
	if err != nil {
		return o.deferExecutionBlocked(ctx, item, "COMPLIANCE_REVOKED", err)
	}
	if err := validateResolvedCompliance(chain, resolvedCompliance, o.cfg.Now()); err != nil {
		return o.deferExecutionBlocked(ctx, item, "COMPLIANCE_REVOKED", err)
	}
	address, err := o.cfg.Custody.Address(ctx, o.cfg.Profiles.DEX.ChainID)
	if err != nil {
		return o.retry(ctx, item, "CUSTODY_UNAVAILABLE", err)
	}
	slippage, err := slippageExact(chain.SlippageToleranceExact, chain.SlippageTolerance)
	if err != nil {
		return o.manualReview(ctx, item, "INVALID_SLIPPAGE", err)
	}
	deadline := o.cfg.Now().Add(o.cfg.Profiles.DEX.QuoteTTL)
	fromToken, err := tokenFromProto(chain.CryptoToken)
	if err != nil {
		return o.manualReview(ctx, item, "CRYPTO_TOKEN_INVALID", err)
	}
	toToken, err := tokenFromProto(chain.StableToken)
	if err != nil {
		return o.manualReview(ctx, item, "STABLE_TOKEN_INVALID", err)
	}
	request := dex.SwapRequest{FromToken: fromToken, ToToken: toToken, Amount: chain.CryptoAmount.Amount, Type: dex.SwapTypeExactIn, SlippageTolerance: chain.SlippageTolerance, SlippageToleranceExact: slippage, Deadline: deadline, Sender: address, Recipient: address, PreferredDEX: o.cfg.Profiles.DEX.ID}
	chain, err = o.refreshBeforeExternalBoundary(ctx, item, fiatBoundarySwapQuote)
	if err != nil {
		return err
	}
	request.Amount = chain.CryptoAmount.Amount
	quote, err := o.cfg.DEX.GetSwapQuote(ctx, request)
	if err != nil {
		return o.retry(ctx, item, "DEX_QUOTE_FAILED", err)
	}
	digest, err := dex.QuoteDigest(quote)
	if err != nil || quote.ID != digest || quote.QuoteDigest != digest || quote.ProfileID != o.cfg.Profiles.DEX.ID || quote.ChainID != o.cfg.Profiles.DEX.ChainID || quote.MinOutputAmount.IsNil() || !quote.MinOutputAmount.IsPositive() || quote.OutputAmount.LT(quote.MinOutputAmount) || quote.IsExpiredAt(o.cfg.Now()) {
		return o.manualReview(ctx, item, "DEX_QUOTE_INVALID", dex.ErrExecutionPayload)
	}
	quoteHash, err := hex.DecodeString(digest)
	if err != nil || len(quoteHash) != sha256.Size {
		return o.manualReview(ctx, item, "DEX_QUOTE_ENCODING_FAILED", dex.ErrExecutionPayload)
	}
	message := o.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_QUOTE_ACCEPTED)
	message.QuoteDigest = quoteHash
	message.QuoteExpiry = quote.ExpiresAt.Unix()
	message.MinimumStableOutput = sdk.NewCoin(chain.StableToken.Denom, quote.MinOutputAmount)
	message.Status = "accepted"
	message.EvidenceHash = hashEvidence("dex_quote", quoteHash)
	return o.submitObservation(ctx, item, message, FiatWorkQuoteReported, func(work *FiatConversionWorkItem) {
		work.State = FiatWorkQuoting
		work.DEXQuote = &quote
		work.QuoteDigest = digest
	})
}

func (o *FiatConversionOrchestrator) signAndExecuteSwap(ctx context.Context, item *FiatConversionWorkItem) error {
	if item.DEXQuote == nil {
		return o.manualReview(ctx, item, "QUOTE_MISSING", ErrFiatIntentChanged)
	}
	quote := *item.DEXQuote
	if quote.IsExpiredAt(o.cfg.Now()) {
		return o.retryResetQuote(ctx, item, "QUOTE_EXPIRED")
	}
	payload, err := dex.BuildExecutionPayload(quote)
	if err != nil {
		return o.manualReview(ctx, item, "PAYLOAD_INVALID", err)
	}
	payloadHash := digestHex(payload)
	_, err = o.refreshBeforeExternalBoundary(ctx, item, fiatBoundarySwapSign)
	if err != nil {
		return err
	}
	updated, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State = FiatWorkSigning
		work.PayloadHash = payloadHash
		work.AttemptCount++
		work.Attempts = appendFiatAttempt(work.Attempts, FiatConversionAttempt{Number: work.AttemptCount, Stage: "swap", StartedAt: o.cfg.Now(), Outcome: "signing"})
		return nil
	})
	if err != nil {
		return err
	}
	txRaw, err := o.cfg.Custody.SignExecution(ctx, quote, payload)
	if err != nil {
		return o.retry(ctx, updated, "CUSTODY_SIGN_FAILED", err)
	}
	signedHash := digestHex(txRaw)
	envelope, err := dex.MarshalSignedExecutionEnvelope(payload, txRaw)
	if err != nil {
		return o.manualReview(ctx, updated, "SIGNED_ENVELOPE_INVALID", err)
	}
	_, err = o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.SignedTxHash = signedHash
		return nil
	})
	if err != nil {
		return err
	}
	chain, err := o.refreshBeforeExternalBoundary(ctx, updated, fiatBoundarySwapBroadcast)
	if err != nil {
		return err
	}
	updated, err = o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State = FiatWorkSwapBroadcast
		return nil
	})
	if err != nil {
		return err
	}
	if err := o.ensureLease(ctx); err != nil {
		return err
	}
	result, err := o.cfg.DEX.ExecuteSwap(ctx, quote, envelope)
	if err != nil {
		_, _ = o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
			work.State = FiatWorkAmbiguous
			work.NextRetryAt = o.cfg.Now()
			return nil
		})
		return &fiatRetryError{cause: fmt.Errorf("%w: %v", ErrSwapOutcomeAmbiguous, err)}
	}
	if result.TxHash == "" {
		_, _ = o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
			work.State = FiatWorkAmbiguous
			work.NextRetryAt = o.cfg.Now()
			return nil
		})
		return &fiatRetryError{cause: ErrSwapOutcomeAmbiguous}
	}
	if !strings.EqualFold(result.TxHash, signedHash) {
		return o.manualReview(ctx, updated, "SWAP_TX_HASH_BINDING_INVALID", ErrSwapOutcomeAmbiguous)
	}
	message := o.baseObservation(updated, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED)
	message.QuoteDigest = mustHex32(updated.QuoteDigest)
	message.MinimumStableOutput = chain.MinimumStableOutput
	message.SwapTxHash = result.TxHash
	message.Status = fiatObservationStatusSubmitted
	message.EvidenceHash = hashEvidence("swap_submitted", []byte(result.TxHash), mustHex32(updated.SignedTxHash))
	return o.submitObservation(ctx, updated, message, FiatWorkSwapBroadcast, func(work *FiatConversionWorkItem) { work.SwapTxHash = result.TxHash })
}

func (o *FiatConversionOrchestrator) reconcileSwap(ctx context.Context, item *FiatConversionWorkItem, chain *settlementv1.FiatConversionRecord) error {
	if item.DEXQuote == nil {
		return o.manualReview(ctx, item, "QUOTE_MISSING", ErrFiatIntentChanged)
	}
	txHash := item.SwapTxHash
	if txHash == "" {
		txHash = chain.SwapTxHash
	}
	status, err := o.cfg.DEX.ReconcileSwap(ctx, *item.DEXQuote, txHash, conversionCorrelation(item.Intent.ConversionID)+"-swap")
	if err != nil {
		return o.retry(ctx, item, "SWAP_RECONCILE_FAILED", err)
	}
	if status.Reorged {
		return o.manualReview(ctx, item, "SWAP_REORG", ErrSwapOutcomeAmbiguous)
	}
	if status.TerminalFailure {
		failureCode := boundedFiatFailureCode(status.FailureCode)
		if failureCode == "MANUAL_REVIEW" {
			return o.manualReview(ctx, item, "SWAP_FAILURE_EVIDENCE_INVALID", ErrSwapOutcomeAmbiguous)
		}
		message := o.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_FAILED)
		message.Status = string(FiatWorkFailed)
		message.FailureCode = failureCode
		message.EvidenceHash = hashEvidence("swap_terminal_failure", []byte(status.TxHash), []byte(failureCode))
		return o.submitObservation(ctx, item, message, FiatWorkFailed, nil)
	}
	if status.Found && item.SwapTxHash == "" {
		if status.TxHash == "" {
			return o.manualReview(ctx, item, "SWAP_RECOVERY_HASH_MISSING", ErrSwapOutcomeAmbiguous)
		}
		message := o.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED)
		message.QuoteDigest = mustHex32(item.QuoteDigest)
		message.MinimumStableOutput = chain.MinimumStableOutput
		message.SwapTxHash = status.TxHash
		message.Status = fiatObservationStatusSubmitted
		message.EvidenceHash = hashEvidence("swap_submitted_recovered", []byte(status.TxHash), mustHex32(item.SignedTxHash))
		return o.submitObservation(ctx, item, message, FiatWorkSwapBroadcast, func(work *FiatConversionWorkItem) { work.SwapTxHash = status.TxHash })
	}
	if !status.Found && txHash == "" && item.SignedTxHash != "" {
		payload, payloadErr := dex.BuildExecutionPayload(*item.DEXQuote)
		if payloadErr != nil || digestHex(payload) != item.PayloadHash {
			return o.manualReview(ctx, item, "SWAP_RECOVERY_PAYLOAD_INVALID", dex.ErrExecutionPayload)
		}
		_, guardErr := o.refreshBeforeExternalBoundary(ctx, item, fiatBoundarySwapSign)
		if guardErr != nil {
			return guardErr
		}
		txRaw, recoverErr := o.cfg.Custody.RecoverSignedExecution(ctx, *item.DEXQuote, payload, item.SignedTxHash)
		if recoverErr != nil {
			return o.manualReview(ctx, item, "SWAP_OUTCOME_UNKNOWN", ErrSwapOutcomeAmbiguous)
		}
		envelope, envelopeErr := dex.MarshalSignedExecutionEnvelope(payload, txRaw)
		if envelopeErr != nil {
			return o.manualReview(ctx, item, "SWAP_RECOVERY_ENVELOPE_INVALID", dex.ErrExecutionPayload)
		}
		chain, guardErr = o.refreshBeforeExternalBoundary(ctx, item, fiatBoundarySwapBroadcast)
		if guardErr != nil {
			return guardErr
		}
		if o.ensureLease(ctx) != nil {
			return ErrFiatConversionLeaseLost
		}
		result, executeErr := o.cfg.DEX.ExecuteSwap(ctx, *item.DEXQuote, envelope)
		if executeErr != nil || result.TxHash == "" {
			return o.manualReview(ctx, item, "SWAP_RECOVERY_AMBIGUOUS", ErrSwapOutcomeAmbiguous)
		}
		if !strings.EqualFold(result.TxHash, item.SignedTxHash) {
			return o.manualReview(ctx, item, "SWAP_RECOVERY_HASH_INVALID", ErrSwapOutcomeAmbiguous)
		}
		message := o.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_SUBMITTED)
		message.QuoteDigest = mustHex32(item.QuoteDigest)
		message.MinimumStableOutput = chain.MinimumStableOutput
		message.SwapTxHash = result.TxHash
		message.Status = fiatObservationStatusSubmitted
		message.EvidenceHash = hashEvidence("swap_submitted", []byte(result.TxHash), mustHex32(item.SignedTxHash))
		return o.submitObservation(ctx, item, message, FiatWorkSwapBroadcast, func(work *FiatConversionWorkItem) { work.SwapTxHash = result.TxHash })
	}
	if !status.Found || !status.Final {
		return o.retry(ctx, item, "SWAP_NOT_FINAL", ErrSwapOutcomeAmbiguous)
	}
	if status.TxHash == "" || (txHash != "" && !strings.EqualFold(status.TxHash, txHash)) || status.Height <= 0 || uint64(status.Confirmations) < o.cfg.Profiles.DEX.FinalityBlocks ||
		status.OutputAmount.IsNil() || status.OutputAmount.LT(item.DEXQuote.MinOutputAmount) || len(status.BlockHash) != sha256.Size || len(status.FinalityHash) != sha256.Size {
		return o.manualReview(ctx, item, "SWAP_FINALITY_INVALID", dex.ErrMinimumOutput)
	}
	expectedFinality, err := CanonicalDEXFinalityHash(item.DEXQuote.ChainID, status.TxHash, status.Height, status.BlockHash, status.Confirmations, status.OutputAmount)
	if err != nil || !bytes.Equal(expectedFinality, status.FinalityHash) {
		return o.manualReview(ctx, item, "SWAP_FINALITY_BINDING_INVALID", dex.ErrPoolStateEvidence)
	}
	message := o.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_SWAP_FINALIZED)
	message.QuoteDigest = mustHex32(item.QuoteDigest)
	message.MinimumStableOutput = chain.MinimumStableOutput
	message.SwapTxHash = status.TxHash
	message.SwapHeight = status.Height
	message.SwapBlockHash = status.BlockHash
	message.SwapFinalityConfirmations = status.Confirmations
	message.SwapFinalityHash = status.FinalityHash
	message.StableAmount = sdk.NewCoin(chain.StableToken.Denom, status.OutputAmount)
	message.Status = "finalized"
	message.EvidenceHash = hashEvidence("swap_finalized", status.BlockHash, status.FinalityHash, []byte(status.TxHash))
	return o.submitObservation(ctx, item, message, FiatWorkSwapFinalized, func(work *FiatConversionWorkItem) {
		work.SwapTxHash = status.TxHash
		work.SwapConfirmation = FiatConversionConfirmation{Height: status.Height, BlockHash: hex.EncodeToString(status.BlockHash), Confirmations: status.Confirmations, FinalityHash: hex.EncodeToString(status.FinalityHash), StableAmount: status.OutputAmount.String()}
		work.FinalityStartedAt = o.cfg.Now()
	})
}

func (o *FiatConversionOrchestrator) quotePayout(ctx context.Context, item *FiatConversionWorkItem, chain *settlementv1.FiatConversionRecord) error {
	resolvedDestination, err := o.cfg.Destination.ResolveDestination(ctx, chain)
	if err != nil {
		return o.retry(ctx, item, "DESTINATION_UNAVAILABLE", err)
	}
	if err := validateResolvedDestination(chain, resolvedDestination); err != nil {
		return o.manualReview(ctx, item, "DESTINATION_BINDING_MISMATCH", err)
	}
	destination := resolvedDestination.Reference
	defer clearString(&destination)
	resolvedCompliance, err := o.cfg.Compliance.ResolveCompliance(ctx, chain)
	if err != nil || validateResolvedCompliance(chain, resolvedCompliance, o.cfg.Now()) != nil {
		return o.deferExecutionBlocked(ctx, item, "COMPLIANCE_REVOKED", offramp.ErrComplianceRequired)
	}
	decision := resolvedCompliance.Decision
	amount, ok := sdkmath.NewIntFromString(item.SwapConfirmation.StableAmount)
	if !ok || !amount.IsPositive() {
		return o.manualReview(ctx, item, "STABLE_AMOUNT_INVALID", dex.ErrMinimumOutput)
	}
	correlation := conversionCorrelation(item.Intent.ConversionID)
	stableDecimals, err := checkedTokenDecimals(chain.StableToken.Decimals)
	if err != nil {
		return o.manualReview(ctx, item, "STABLE_TOKEN_INVALID", err)
	}
	chain, err = o.refreshBeforeExternalBoundary(ctx, item, fiatBoundaryPayoutQuote)
	if err != nil {
		return err
	}
	quote, err := o.cfg.Offramp.GetQuote(ctx, offramp.QuoteRequest{CryptoSymbol: chain.StableToken.Symbol, CryptoDenom: chain.StableToken.Denom, CryptoDecimals: stableDecimals, CryptoAmount: amount, FiatCurrency: chain.FiatCurrency, PaymentMethod: chain.PaymentMethod, Sender: chain.Provider, Destination: destination, BeneficiaryReference: destination, Jurisdiction: chain.DestinationRegion, CorrelationID: correlation, Compliance: decision})
	if err != nil {
		return o.retry(ctx, item, "PAYOUT_QUOTE_FAILED", err)
	}
	quoteHash, requestHash, corridorID, providerBinding, err := canonicalPayoutQuoteCommitments(quote, o.cfg.Profiles, item.Intent.ComplianceDigest)
	if err != nil || quote.ID == "" || quote.IsExpired(o.cfg.Now()) || quote.Provider != o.cfg.Profiles.Payout.Provider {
		return o.manualReview(ctx, item, "PAYOUT_QUOTE_INVALID", offramp.ErrInvalidRequest)
	}
	message := o.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_QUOTED)
	message.OffRampQuoteId = quote.ID
	message.QuoteDigest = quoteHash
	message.QuoteExpiry = quote.ExpiresAt.Unix()
	message.Status = "quoted"
	message.EvidenceHash = hashEvidence("payout_quoted", quoteHash, []byte(quote.ID))
	return o.submitObservation(ctx, item, message, FiatWorkPayoutQuote, func(work *FiatConversionWorkItem) {
		work.PayoutQuote = payoutQuoteSnapshot(quote, requestHash, corridorID, providerBinding, item.Intent.ComplianceDigest, work.PayoutProfileDigest)
		work.PayoutQuoteDigest = hex.EncodeToString(quoteHash)
	})
}

func (o *FiatConversionOrchestrator) initiatePayout(ctx context.Context, item *FiatConversionWorkItem, chain *settlementv1.FiatConversionRecord) error {
	if item.PayoutQuote.ID == "" {
		return o.manualReview(ctx, item, "PAYOUT_QUOTE_MISSING", ErrFiatIntentChanged)
	}
	if !o.cfg.Now().Before(item.PayoutQuote.ExpiresAt) {
		return o.retryResetPayoutQuote(ctx, item, chain, "PAYOUT_QUOTE_EXPIRED")
	}
	resolvedDestination, err := o.cfg.Destination.ResolveDestination(ctx, chain)
	if err != nil {
		return o.retry(ctx, item, "DESTINATION_UNAVAILABLE", err)
	}
	if err := validateResolvedDestination(chain, resolvedDestination); err != nil {
		return o.manualReview(ctx, item, "DESTINATION_BINDING_MISMATCH", err)
	}
	destination := resolvedDestination.Reference
	defer clearString(&destination)
	resolvedCompliance, err := o.cfg.Compliance.ResolveCompliance(ctx, chain)
	if err != nil || validateResolvedCompliance(chain, resolvedCompliance, o.cfg.Now()) != nil {
		return o.deferExecutionBlocked(ctx, item, "COMPLIANCE_REVOKED", offramp.ErrComplianceRequired)
	}
	decision := resolvedCompliance.Decision
	metadata := payoutMetadata(item)
	chain, err = o.refreshBeforeExternalBoundary(ctx, item, fiatBoundaryPayoutInitiate)
	if err != nil {
		return err
	}
	quote, err := reconstructPayoutQuote(item.PayoutQuote, chain, item, destination, decision, o.cfg.Profiles)
	if err != nil {
		return o.manualReview(ctx, item, "PAYOUT_QUOTE_INVALID", err)
	}
	item, err = o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State = FiatWorkPayoutAmbiguous
		work.NextRetryAt = o.cfg.Now()
		return nil
	})
	if err != nil {
		return err
	}
	if err := o.ensureLease(ctx); err != nil {
		return err
	}
	result, err := o.cfg.Offramp.InitiatePayout(ctx, quote, item.SwapTxHash, destination, metadata)
	if err != nil {
		if offramp.IsAmbiguous(err) {
			_, _ = o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
				work.State = FiatWorkPayoutAmbiguous
				work.NextRetryAt = o.cfg.Now()
				return nil
			})
			return &fiatRetryError{cause: err}
		}
		return o.retry(ctx, item, "PAYOUT_INIT_FAILED", err)
	}
	if result.ID == "" || result.QuoteID != item.PayoutQuote.ID {
		return o.manualReview(ctx, item, "PAYOUT_BINDING_INVALID", offramp.ErrProviderRejected)
	}
	item, err = o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.PayoutID = result.ID
		work.PayoutStatus = string(result.Status)
		work.PayoutReferenceHash = hex.EncodeToString(hashEvidence("payout_reference", []byte(result.Reference)))
		return nil
	})
	if err != nil {
		return err
	}
	if o.cfg.WebhookBindings != nil {
		if err := o.cfg.WebhookBindings.PutWebhookBinding(ctx, offramp.WebhookBinding{Provider: result.Provider, PayoutID: result.ID, QuoteID: result.QuoteID, CorrelationID: metadata["correlation_id"], ReservationDay: result.InitiatedAt.UTC().Format("2006-01-02")}); err != nil {
			return o.manualReview(ctx, item, "WEBHOOK_BINDING_STORE_FAILED", err)
		}
	}
	message := o.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_SUBMITTED)
	message.OffRampQuoteId = result.QuoteID
	message.OffRampPayoutId = result.ID
	message.QuoteDigest = mustHex32(item.PayoutQuoteDigest)
	message.QuoteExpiry = item.PayoutQuote.ExpiresAt.Unix()
	message.Status = string(result.Status)
	message.PrivacySafeReferenceHash = hashEvidence("payout_reference", []byte(result.Reference))
	message.EvidenceHash = hashEvidence("payout_submitted", []byte(result.ID), []byte(result.QuoteID))
	return o.submitObservation(ctx, item, message, FiatWorkPayoutSubmitted, func(work *FiatConversionWorkItem) {
		work.PayoutID = result.ID
		work.PayoutStatus = string(result.Status)
		work.PayoutReferenceHash = hex.EncodeToString(message.PrivacySafeReferenceHash)
	})
}

func (o *FiatConversionOrchestrator) reconcilePayout(ctx context.Context, item *FiatConversionWorkItem) error {
	var result offramp.PayoutResult
	var err error
	recoveredWithoutSubmittedObservation := item.PayoutID == ""
	var webhook *offramp.WebhookEvent
	if item.PayoutID != "" && o.cfg.WebhookEvents != nil {
		events, eventErr := o.cfg.WebhookEvents.VerifiedWebhookEvents(ctx, item.PayoutID)
		if eventErr != nil {
			return o.retry(ctx, item, "WEBHOOK_QUERY_FAILED", eventErr)
		}
		for i := range events {
			if events[i].Provider != o.cfg.Profiles.Payout.Provider || events[i].PayoutID != item.PayoutID || events[i].QuoteID != item.PayoutQuote.ID || events[i].CorrelationID != payoutMetadata(item)["correlation_id"] {
				return o.manualReview(ctx, item, "WEBHOOK_BINDING_INVALID", offramp.ErrWebhookInvalid)
			}
			if webhook != nil && webhook.Status != events[i].Status {
				return o.manualReview(ctx, item, "WEBHOOK_STATUS_CONFLICT", offramp.ErrWebhookReplay)
			}
			copyEvent := events[i]
			webhook = &copyEvent
		}
	}
	if item.PayoutID == "" {
		result, err = o.cfg.Offramp.FindPayoutByMetadata(ctx, o.cfg.Profiles.Payout.Provider, payoutMetadata(item))
	} else {
		result, err = o.cfg.Offramp.GetStatus(ctx, item.PayoutID)
	}
	if err != nil {
		return o.retry(ctx, item, "PAYOUT_RECONCILE_FAILED", err)
	}
	if result.ID == "" {
		return o.retry(ctx, item, "PAYOUT_NOT_FOUND", offramp.ErrPayoutNotFound)
	}
	if err := validatePayoutResultBinding(item, result); err != nil {
		return o.manualReview(ctx, item, "PAYOUT_BINDING_INVALID", err)
	}
	if webhook != nil {
		if result.Status != webhook.Status {
			return o.retry(ctx, item, "WEBHOOK_PROVIDER_STATUS_DIVERGED", offramp.ErrProviderTemporary)
		}
	}
	if item.PayoutID == "" {
		item, _ = o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
			work.PayoutID = result.ID
			return nil
		})
	}
	if o.cfg.WebhookBindings != nil {
		bindingErr := o.cfg.WebhookBindings.PutWebhookBinding(ctx, offramp.WebhookBinding{Provider: result.Provider, PayoutID: result.ID, QuoteID: result.QuoteID, CorrelationID: payoutMetadata(item)["correlation_id"], ReservationDay: result.InitiatedAt.UTC().Format("2006-01-02")})
		if bindingErr != nil {
			return o.retry(ctx, item, "WEBHOOK_BINDING_STORE_FAILED", bindingErr)
		}
	}
	if result.Status == offramp.StatusFailed || result.Status == offramp.StatusCancelled {
		return o.manualReview(ctx, item, "PAYOUT_TERMINAL_FAILURE", offramp.ErrProviderRejected)
	}
	if recoveredWithoutSubmittedObservation {
		referenceHash := hashEvidence("payout_reference", []byte(result.Reference))
		message := o.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_SUBMITTED)
		message.OffRampQuoteId = result.QuoteID
		message.OffRampPayoutId = result.ID
		message.QuoteDigest = mustHex32(item.PayoutQuoteDigest)
		message.QuoteExpiry = item.PayoutQuote.ExpiresAt.Unix()
		message.Status = string(result.Status)
		message.PrivacySafeReferenceHash = referenceHash
		message.EvidenceHash = hashEvidence("payout_submitted_recovered", []byte(result.ID), []byte(result.QuoteID))
		return o.submitObservation(ctx, item, message, FiatWorkPayoutSubmitted, func(work *FiatConversionWorkItem) {
			work.PayoutID, work.PayoutStatus, work.PayoutReferenceHash = result.ID, string(result.Status), hex.EncodeToString(referenceHash)
		})
	}
	if result.Status != offramp.StatusCompleted {
		return o.retry(ctx, item, "PAYOUT_NOT_FINAL", offramp.ErrProviderTemporary)
	}
	if webhook == nil || result.CompletedAt == nil || result.FiatAmount.IsNil() || !result.FiatAmount.IsPositive() || result.QuoteID != item.PayoutQuote.ID || result.Provider != o.cfg.Profiles.Payout.Provider {
		return o.manualReview(ctx, item, "PAYOUT_FINALITY_INVALID", offramp.ErrProviderRejected)
	}
	referenceHash := hashEvidence("payout_reference", []byte(result.Reference))
	finalityHash := hashEvidence("payout_finality", []byte(result.ID), []byte(result.Status), []byte(result.FiatAmount.String()), []byte(result.CompletedAt.UTC().Format(time.RFC3339Nano)))
	message := o.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_PAYOUT_COMPLETED)
	message.OffRampQuoteId = item.PayoutQuote.ID
	message.OffRampPayoutId = result.ID
	message.QuoteDigest = mustHex32(item.PayoutQuoteDigest)
	message.Status = string(FiatWorkCompleted)
	message.PrivacySafeReferenceHash = referenceHash
	message.PayoutFinalityHash = finalityHash
	message.FiatAmount = result.FiatAmount.String()
	message.EvidenceHash = hashEvidence("payout_completed", referenceHash, finalityHash)
	return o.submitObservation(ctx, item, message, FiatWorkCompleted, func(work *FiatConversionWorkItem) {
		work.PayoutStatus = string(FiatWorkCompleted)
		work.PayoutReferenceHash = hex.EncodeToString(referenceHash)
		work.PayoutFinalityHash = hex.EncodeToString(finalityHash)
		work.TerminalResult = string(FiatWorkCompleted)
		work.PendingWebhookProvider = webhook.Provider
		work.PendingWebhookEventID = webhook.EventID
	})
}

func (o *FiatConversionOrchestrator) submitObservation(ctx context.Context, item *FiatConversionWorkItem, msg *settlementv1.MsgRecordFiatConversionObservation, next FiatConversionWorkState, mutate func(*FiatConversionWorkItem)) error {
	if mutate != nil {
		updated, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error { mutate(work); return nil })
		if err != nil {
			return err
		}
		item = updated
	}
	messageBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(messageBytes)
	updated, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		copyMsg := *msg
		work.PendingObservation = &copyMsg
		work.PendingNextState = next
		work.ObservationDigest = hex.EncodeToString(digest[:])
		return nil
	})
	if err != nil {
		return err
	}
	result, err := o.cfg.Submitter.SubmitFiatConversionObservation(ctx, msg)
	if err != nil {
		if result.ID != "" {
			return &fiatRetryError{cause: fmt.Errorf("%w: %v", ErrFiatObservationPending, err)}
		}
		return o.retry(ctx, updated, "OBSERVATION_SUBMIT_FAILED", err)
	}
	if !result.Final {
		return &fiatRetryError{cause: ErrFiatObservationPending}
	}
	return o.finalizePendingObservation(ctx, updated, next)
}

func (o *FiatConversionOrchestrator) reconcilePendingObservation(ctx context.Context, item *FiatConversionWorkItem, chain *settlementv1.FiatConversionRecord) error {
	msg := item.PendingObservation
	if msg == nil {
		return nil
	}
	digestBytes, err := proto.Marshal(msg)
	if err != nil {
		return o.manualReview(ctx, item, "OBSERVATION_MALFORMED", err)
	}
	digest := sha256.Sum256(digestBytes)
	if chain.ObservationSequence == msg.ObservationSequence && bytes.Equal(chain.LastObservationDigest, digest[:]) {
		return o.finalizePendingObservation(ctx, item, item.PendingNextState)
	}
	if chain.ObservationSequence >= msg.ObservationSequence {
		return o.manualReview(ctx, item, "OBSERVATION_CONFLICT", ErrFiatIntentChanged)
	}
	if chain.ObservationSequence < item.ObservationSequence {
		return o.deferWithoutAttempt(ctx, item, "CHAIN_OBSERVATION_LAG")
	}
	if chain.ObservationSequence+1 != msg.ObservationSequence {
		return o.manualReview(ctx, item, "OBSERVATION_SEQUENCE_DIVERGED", ErrFiatIntentChanged)
	}
	result, err := o.cfg.Submitter.SubmitFiatConversionObservation(ctx, msg)
	if err != nil || !result.Final {
		return &fiatRetryError{cause: ErrFiatObservationPending}
	}
	return o.finalizePendingObservation(ctx, item, item.PendingNextState)
}

func (o *FiatConversionOrchestrator) finalizePendingObservation(ctx context.Context, item *FiatConversionWorkItem, next FiatConversionWorkState) error {
	webhookProvider, webhookEventID := "", ""
	if current, getErr := o.cfg.Store.Get(ctx, item.Intent.ConversionID); getErr == nil {
		webhookProvider, webhookEventID = current.PendingWebhookProvider, current.PendingWebhookEventID
	}
	if next == FiatWorkCompleted && webhookEventID != "" {
		if o.cfg.WebhookEvents == nil || o.cfg.WebhookEvents.ConsumeVerifiedWebhookEvent(ctx, webhookProvider, webhookEventID, o.cfg.Now()) != nil {
			return ErrFiatObservationPending
		}
	}
	_, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		if work.PendingObservation != nil {
			work.ObservationSequence = work.PendingObservation.ObservationSequence
		}
		work.PendingObservation = nil
		work.PendingNextState = ""
		work.PendingWebhookProvider = ""
		work.PendingWebhookEventID = ""
		work.State = next
		work.NextRetryAt = o.cfg.Now()
		if next == FiatWorkCompleted {
			work.TerminalResult = string(FiatWorkCompleted)
		}
		return nil
	})
	if err == nil && next == FiatWorkCompleted {
		o.mu.Lock()
		o.lastSuccess = o.cfg.Now()
		o.mu.Unlock()
	}
	return err
}

func (o *FiatConversionOrchestrator) baseObservation(item *FiatConversionWorkItem, stage settlementv1.FiatConversionObservationStage) *settlementv1.MsgRecordFiatConversionObservation {
	sequence := item.ObservationSequence + 1
	idempotency := hashEvidence("fiat_observation", []byte(item.Intent.ConversionID), []byte(strconv.FormatUint(sequence, 10)), []byte(stage.String()))
	return &settlementv1.MsgRecordFiatConversionObservation{Sender: o.cfg.ProviderAddress, ConversionId: item.Intent.ConversionID, ObservationSequence: sequence, IdempotencyKey: idempotency, Stage: stage, DexProfileId: item.DEXProfileID, DexProfileDigest: mustHex32(item.DEXProfileDigest), PayoutProfileId: item.PayoutProfileID, PayoutProfileDigest: mustHex32(item.PayoutProfileDigest), ComplianceDecisionHash: mustHex32(item.Intent.ComplianceDigest), ObservedAt: o.cfg.Now().Unix(), Status: strings.ToLower(strings.TrimPrefix(stage.String(), "FIAT_CONVERSION_OBSERVATION_STAGE_"))}
}

func (o *FiatConversionOrchestrator) validateIntentAndProfiles(ctx context.Context, item *FiatConversionWorkItem, chain *settlementv1.FiatConversionRecord) (*settlementv1.FiatConversionRecord, error) {
	if chain == nil || !fiatIntentEqual(item.Intent, snapshotFiatIntent(chain)) {
		return nil, ErrFiatIntentChanged
	}
	authorization, err := o.cfg.Query.ExecutionAuthorization(ctx, item.Intent.ConversionID)
	if err != nil {
		return nil, err
	}
	chain = authorization.Conversion
	params := authorization.Params
	if chain == nil || !fiatIntentEqual(item.Intent, snapshotFiatIntent(chain)) {
		return nil, ErrFiatIntentChanged
	}
	if !params.FiatConversionEnabled {
		return nil, ErrFiatGovernanceDisabled
	}
	if authorization.ActiveHoldCount > 0 {
		return nil, fmt.Errorf("%w: case %s", ErrFiatFinancialHold, authorization.ActiveCaseID)
	}
	if err := o.validateImmutableIntentAndProfiles(item, chain); err != nil {
		return nil, err
	}
	certified := settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED
	if o.cfg.Production && (params.FiatConversionDexProfileState != certified || params.FiatConversionPayoutProfileState != certified) {
		return nil, ErrFiatCurrentProfile
	}
	if params.FiatConversionDexProfileId != item.DEXProfileID || params.FiatConversionPayoutProfileId != item.PayoutProfileID || !bytes.Equal(params.FiatConversionDexProfileDigest, mustHex32(item.DEXProfileDigest)) || !bytes.Equal(params.FiatConversionPayoutProfileDigest, mustHex32(item.PayoutProfileDigest)) {
		return nil, ErrFiatCurrentProfile
	}
	if chain.LegacyQuarantined || chainFiatConversionTerminal(chain.State) {
		return nil, ErrFiatIntentChanged
	}
	if err := o.cfg.Destination.ProductionReady(ctx); o.cfg.Production && err != nil {
		return nil, err
	}
	if err := o.cfg.Compliance.ProductionReady(ctx); o.cfg.Production && err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFiatCurrentCompliance, err)
	}
	return chain, nil
}

func (o *FiatConversionOrchestrator) validateImmutableIntentAndProfiles(item *FiatConversionWorkItem, chain *settlementv1.FiatConversionRecord) error {
	if chain == nil || !fiatIntentEqual(item.Intent, snapshotFiatIntent(chain)) ||
		chain.DexProfileId != item.DEXProfileID || chain.PayoutProfileId != item.PayoutProfileID ||
		!bytes.Equal(chain.DexProfileDigest, mustHex32(item.DEXProfileDigest)) ||
		!bytes.Equal(chain.PayoutProfileDigest, mustHex32(item.PayoutProfileDigest)) {
		return ErrFiatProfileMismatch
	}
	if chain.LegacyQuarantined || chainFiatConversionTerminal(chain.State) {
		return ErrFiatIntentChanged
	}
	return nil
}

func fiatWorkRequiresReadOnlyReconciliation(item *FiatConversionWorkItem, chain *settlementv1.FiatConversionRecord) bool {
	if item == nil || chain == nil {
		return false
	}
	switch item.State {
	case FiatWorkSwapBroadcast, FiatWorkAmbiguous, FiatWorkPayoutSubmitted, FiatWorkPayoutAmbiguous:
		return true
	default:
		return false
	}
}

type fiatExternalBoundary string

const (
	fiatBoundarySwapQuote      fiatExternalBoundary = "swap_quote"
	fiatBoundarySwapSign       fiatExternalBoundary = "swap_sign"
	fiatBoundarySwapBroadcast  fiatExternalBoundary = "swap_broadcast"
	fiatBoundaryPayoutQuote    fiatExternalBoundary = "payout_quote"
	fiatBoundaryPayoutInitiate fiatExternalBoundary = "payout_initiate"
)

func (o *FiatConversionOrchestrator) refreshBeforeExternalBoundary(ctx context.Context, item *FiatConversionWorkItem, boundary fiatExternalBoundary) (*settlementv1.FiatConversionRecord, error) {
	authorization, err := o.cfg.Query.ExecutionAuthorization(ctx, item.Intent.ConversionID)
	if err != nil {
		return nil, o.retry(ctx, item, "AUTHORIZATION_QUERY_UNAVAILABLE", err)
	}
	chain, err := o.validateExecutionAuthorization(item, authorization, boundary)
	if err == nil {
		return chain, nil
	}
	if currentFiatAuthorizationError(err) {
		return nil, o.deferExecutionBlocked(ctx, item, executionBlockCode(err), err)
	}
	return nil, o.manualReview(ctx, item, "BOUNDARY_AUTHORIZATION_INVALID", err)
}

func (o *FiatConversionOrchestrator) validateExecutionAuthorization(item *FiatConversionWorkItem, authorization FiatExecutionAuthorization, boundary fiatExternalBoundary) (*settlementv1.FiatConversionRecord, error) {
	chain := authorization.Conversion
	if chain == nil || !fiatIntentEqual(item.Intent, snapshotFiatIntent(chain)) {
		return nil, ErrFiatIntentChanged
	}
	if !authorization.Params.FiatConversionEnabled {
		return nil, ErrFiatGovernanceDisabled
	}
	if authorization.ActiveHoldCount > 0 {
		return nil, fmt.Errorf("%w: case %s", ErrFiatFinancialHold, authorization.ActiveCaseID)
	}
	if chain.ObservationSequence != item.ObservationSequence || !bytes.Equal(chain.LastObservationDigest, mustHex32(item.ObservationDigest)) {
		return nil, ErrFiatIntentChanged
	}
	if _, err := o.validateIntentProfilesOnly(item, chain, authorization.Params); err != nil {
		return nil, err
	}
	switch boundary {
	case fiatBoundarySwapQuote:
		if !chainStateIs(chain.State, "CREATED", "SWAP_PENDING") || chain.SwapTxHash != "" {
			return nil, ErrFiatIntentChanged
		}
	case fiatBoundarySwapSign, fiatBoundarySwapBroadcast:
		if !chainStateIs(chain.State, fiatChainStateSwapPending) || item.DEXQuote == nil || chain.SwapTxHash != "" ||
			!bytes.Equal(chain.QuoteDigest, mustHex32(item.QuoteDigest)) || chain.QuoteExpiry != item.DEXQuote.ExpiresAt.Unix() ||
			!chain.MinimumStableOutput.IsEqual(sdk.NewCoin(chain.StableToken.Denom, item.DEXQuote.MinOutputAmount)) {
			return nil, ErrFiatIntentChanged
		}
	case fiatBoundaryPayoutQuote:
		if chainStateIs(chain.State, fiatChainStateSwapSettled) {
			if chain.SwapTxHash == "" {
				return nil, ErrFiatIntentChanged
			}
		} else if !chainStateIs(chain.State, fiatChainStatePayoutPending) || item.PayoutQuote.ID == "" ||
			chain.OffRampId != "" || chain.OffRampReference != "" || chain.OffRampQuoteId != item.PayoutQuote.ID ||
			!bytes.Equal(chain.QuoteDigest, mustHex32(item.PayoutQuoteDigest)) || chain.QuoteExpiry != item.PayoutQuote.ExpiresAt.Unix() ||
			o.cfg.Now().Unix() < chain.QuoteExpiry {
			return nil, ErrFiatIntentChanged
		}
	case fiatBoundaryPayoutInitiate:
		if !chainStateIs(chain.State, fiatChainStatePayoutPending) || chain.OffRampQuoteId != item.PayoutQuote.ID ||
			!bytes.Equal(chain.QuoteDigest, mustHex32(item.PayoutQuoteDigest)) || chain.QuoteExpiry != item.PayoutQuote.ExpiresAt.Unix() {
			return nil, ErrFiatIntentChanged
		}
	}
	return chain, nil
}

func (o *FiatConversionOrchestrator) validateIntentProfilesOnly(item *FiatConversionWorkItem, chain *settlementv1.FiatConversionRecord, params settlementv1.Params) (*settlementv1.FiatConversionRecord, error) {
	if err := o.validateImmutableIntentAndProfiles(item, chain); err != nil {
		return nil, err
	}
	certified := settlementv1.FiatConversionProfileState_FIAT_CONVERSION_PROFILE_STATE_CERTIFIED_ENABLED
	if o.cfg.Production && (params.FiatConversionDexProfileState != certified || params.FiatConversionPayoutProfileState != certified) {
		return nil, ErrFiatCurrentProfile
	}
	if params.FiatConversionDexProfileId != item.DEXProfileID || params.FiatConversionPayoutProfileId != item.PayoutProfileID ||
		!bytes.Equal(params.FiatConversionDexProfileDigest, mustHex32(item.DEXProfileDigest)) || !bytes.Equal(params.FiatConversionPayoutProfileDigest, mustHex32(item.PayoutProfileDigest)) {
		return nil, ErrFiatCurrentProfile
	}
	return chain, nil
}

func chainStateIs(state string, allowed ...string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(state))
	for _, candidate := range allowed {
		if normalized == candidate {
			return true
		}
	}
	return false
}

func executionBlockCode(err error) string {
	if errors.Is(err, ErrFiatGovernanceDisabled) {
		return "GOVERNANCE_DISABLED"
	}
	if errors.Is(err, ErrFiatCurrentProfile) {
		return "CURRENT_PROFILE_CHANGED"
	}
	if errors.Is(err, ErrFiatCurrentCompliance) {
		return "COMPLIANCE_REVOKED"
	}
	return "FINANCIAL_HOLD_ACTIVE"
}

func currentFiatAuthorizationError(err error) bool {
	return errors.Is(err, ErrFiatGovernanceDisabled) || errors.Is(err, ErrFiatFinancialHold) || errors.Is(err, ErrFiatCurrentProfile) || errors.Is(err, ErrFiatCurrentCompliance)
}

func (o *FiatConversionOrchestrator) deferExecutionBlocked(ctx context.Context, item *FiatConversionWorkItem, code string, cause error) error {
	_, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.FailureCode = boundedFiatFailureCode(code)
		work.NextRetryAt = o.cfg.Now().Add(o.cfg.RetryBackoff)
		return nil
	})
	o.recordFailure()
	if err != nil {
		return err
	}
	return &fiatRetryError{cause: cause}
}

func (o *FiatConversionOrchestrator) manualReviewWithoutObservation(ctx context.Context, item *FiatConversionWorkItem, code string, cause error) error {
	_, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State = FiatWorkManualReview
		work.FailureCode = boundedFiatFailureCode(code)
		work.TerminalResult = "manual_review"
		work.NextRetryAt = time.Time{}
		return nil
	})
	o.recordFailure()
	if err != nil {
		return err
	}
	return &fiatTerminalError{cause: cause}
}

func (o *FiatConversionOrchestrator) recoverOnStart(ctx context.Context) error {
	items, err := o.cfg.Store.List(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.State.terminal() {
			continue
		}
		_, err = o.cfg.Store.Update(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
			work.LeaseToken = o.leaseToken
			switch work.State {
			case FiatWorkSigning, FiatWorkSwapBroadcast:
				work.State = FiatWorkAmbiguous
			case FiatWorkPayoutSubmitted:
				work.State = FiatWorkPayoutAmbiguous
			}
			work.NextRetryAt = o.cfg.Now()
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *FiatConversionOrchestrator) retry(ctx context.Context, item *FiatConversionWorkItem, code string, cause error) error {
	if current, err := o.cfg.Store.Get(ctx, item.Intent.ConversionID); err == nil {
		item = current
	}
	if item.AttemptCount+1 >= o.cfg.MaxAttempts {
		return o.manualReview(ctx, item, "ATTEMPTS_EXHAUSTED", cause)
	}
	_, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.AttemptCount++
		work.NextRetryAt = o.cfg.Now().Add(o.retryDelay(work.AttemptCount))
		work.FailureCode = code
		work.Attempts = appendFiatAttempt(work.Attempts, FiatConversionAttempt{Number: work.AttemptCount, Stage: string(work.State), StartedAt: o.cfg.Now(), FinishedAt: o.cfg.Now(), Classification: code, Outcome: "retry"})
		return nil
	})
	o.recordFailure()
	if err != nil {
		return err
	}
	return &fiatRetryError{cause: cause}
}

func (o *FiatConversionOrchestrator) deferWithoutAttempt(ctx context.Context, item *FiatConversionWorkItem, code string) error {
	_, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.NextRetryAt = o.cfg.Now().Add(o.cfg.RetryBackoff)
		work.FailureCode = code
		return nil
	})
	if err != nil {
		return err
	}
	return &fiatRetryError{cause: ErrFiatObservationPending}
}

func (o *FiatConversionOrchestrator) retryResetQuote(ctx context.Context, item *FiatConversionWorkItem, code string) error {
	if item.AttemptCount+1 >= o.cfg.MaxAttempts {
		return o.manualReview(ctx, item, "ATTEMPTS_EXHAUSTED", dex.ErrQuoteExpired)
	}
	_, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.AttemptCount++
		work.State = FiatWorkClaimed
		work.DEXQuote = nil
		work.QuoteDigest = ""
		work.PayloadHash = ""
		work.SignedTxHash = ""
		work.NextRetryAt = o.cfg.Now().Add(o.retryDelay(work.AttemptCount))
		work.FailureCode = code
		return nil
	})
	if err != nil {
		return err
	}
	return &fiatRetryError{cause: dex.ErrQuoteExpired}
}

func (o *FiatConversionOrchestrator) retryResetPayoutQuote(ctx context.Context, item *FiatConversionWorkItem, chain *settlementv1.FiatConversionRecord, code string) error {
	if item.PayoutID != "" || item.PayoutReferenceHash != "" || chain == nil || !chainStateIs(chain.State, fiatChainStatePayoutPending) ||
		chain.OffRampId != "" || chain.OffRampReference != "" || chain.OffRampQuoteId != item.PayoutQuote.ID ||
		!bytes.Equal(chain.QuoteDigest, mustHex32(item.PayoutQuoteDigest)) || chain.QuoteExpiry != item.PayoutQuote.ExpiresAt.Unix() ||
		o.cfg.Now().Unix() < chain.QuoteExpiry {
		return o.manualReview(ctx, item, "PAYOUT_REQUOTE_UNSAFE", ErrFiatIntentChanged)
	}
	if item.AttemptCount+1 >= o.cfg.MaxAttempts {
		return o.manualReview(ctx, item, "ATTEMPTS_EXHAUSTED", offramp.ErrQuoteExpired)
	}
	_, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.AttemptCount++
		work.State = FiatWorkSwapFinalized
		work.NextRetryAt = o.cfg.Now().Add(o.retryDelay(work.AttemptCount))
		work.FailureCode = code
		work.Attempts = appendFiatAttempt(work.Attempts, FiatConversionAttempt{Number: work.AttemptCount, Stage: "payout_quote", StartedAt: o.cfg.Now(), FinishedAt: o.cfg.Now(), Classification: code, Outcome: "requote"})
		return nil
	})
	if err != nil {
		return err
	}
	return &fiatRetryError{cause: offramp.ErrQuoteExpired}
}

func (o *FiatConversionOrchestrator) manualReview(ctx context.Context, item *FiatConversionWorkItem, code string, cause error) error {
	if safeFailureObservation(item, code) && item.PendingObservation == nil {
		msg := o.baseObservation(item, settlementv1.FiatConversionObservationStage_FIAT_CONVERSION_OBSERVATION_STAGE_FAILED)
		msg.Status = string(FiatWorkFailed)
		msg.FailureCode = boundedFiatFailureCode(code)
		msg.EvidenceHash = hashEvidence("fiat_failure", []byte(msg.FailureCode), []byte(item.Intent.ConversionID))
		messageBytes, marshalErr := proto.Marshal(msg)
		if marshalErr == nil {
			digest := sha256.Sum256(messageBytes)
			updated, updateErr := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
				work.PendingObservation = msg
				work.PendingNextState = FiatWorkFailed
				work.FailureCode = msg.FailureCode
				work.TerminalResult = string(FiatWorkFailed)
				work.ObservationDigest = hex.EncodeToString(digest[:])
				work.NextRetryAt = o.cfg.Now().Add(o.cfg.MaxRetryBackoff)
				return nil
			})
			if updateErr == nil {
				result, submitErr := o.cfg.Submitter.SubmitFiatConversionObservation(ctx, msg)
				if submitErr == nil && result.Final {
					_ = o.finalizePendingObservation(ctx, updated, FiatWorkFailed)
				}
				o.recordFailure()
				return &fiatTerminalError{cause: cause}
			}
		}
	}
	_, err := o.updateOwned(ctx, item.Intent.ConversionID, func(work *FiatConversionWorkItem) error {
		work.State = FiatWorkManualReview
		work.DEXQuote = nil
		work.QuoteDigest = ""
		work.FailureCode = code
		work.TerminalResult = "manual_review"
		work.NextRetryAt = time.Time{}
		return nil
	})
	o.recordFailure()
	if err != nil {
		return err
	}
	return &fiatTerminalError{cause: cause}
}

func (o *FiatConversionOrchestrator) updateOwned(ctx context.Context, id string, fn func(*FiatConversionWorkItem) error) (*FiatConversionWorkItem, error) {
	if err := o.ensureLease(ctx); err != nil {
		return nil, err
	}
	return o.cfg.Store.Update(ctx, id, func(work *FiatConversionWorkItem) error {
		if work.LeaseToken != 0 && work.LeaseToken != o.leaseToken {
			return ErrFiatConversionLeaseLost
		}
		work.LeaseToken = o.leaseToken
		return fn(work)
	})
}
func (o *FiatConversionOrchestrator) ensureLease(ctx context.Context) error {
	if o == nil || !o.cfg.Enabled {
		return ErrFiatOrchestratorBlocked
	}
	o.mu.RLock()
	running := o.running
	token := o.leaseToken
	o.mu.RUnlock()
	if !running || !o.cfg.Lease.Held(ctx, o.leaseName, token) {
		return ErrFiatConversionLeaseLost
	}
	return nil
}
func (o *FiatConversionOrchestrator) retryDelay(attempt uint32) time.Duration {
	delay := o.cfg.RetryBackoff
	for i := uint32(1); i < attempt && delay < o.cfg.MaxRetryBackoff; i++ {
		if delay > o.cfg.MaxRetryBackoff/2 {
			delay = o.cfg.MaxRetryBackoff
			break
		}
		delay *= 2
	}
	if delay > o.cfg.MaxRetryBackoff {
		delay = o.cfg.MaxRetryBackoff
	}
	return delay
}
func (o *FiatConversionOrchestrator) recordFailure() {
	o.mu.Lock()
	o.lastFailure = o.cfg.Now()
	o.mu.Unlock()
}

func (o *FiatConversionOrchestrator) Metrics(ctx context.Context) FiatConversionMetrics {
	if o == nil || !o.cfg.Enabled {
		return FiatConversionMetrics{}
	}
	items, err := o.cfg.Store.List(ctx)
	if err != nil {
		return FiatConversionMetrics{}
	}
	now := o.cfg.Now()
	metrics := FiatConversionMetrics{}
	for _, item := range items {
		switch item.State {
		case FiatWorkCompleted:
			if !item.FinalityStartedAt.IsZero() {
				latency := item.UpdatedAt.Sub(item.FinalityStartedAt)
				if latency > metrics.FinalityLatency {
					metrics.FinalityLatency = latency
				}
			}
		case FiatWorkFailed:
			metrics.DeadLetters++
		case FiatWorkManualReview:
			metrics.ManualReview++
		default:
			metrics.QueueDepth++
			age := now.Sub(item.CreatedAt)
			if age > metrics.OldestIntentAge {
				metrics.OldestIntentAge = age
			}
		}
		if item.State == FiatWorkAmbiguous {
			metrics.AmbiguousSwaps++
		}
		if item.State == FiatWorkPayoutAmbiguous {
			metrics.AmbiguousPayouts++
		}
	}
	o.mu.RLock()
	metrics.LastSuccess = o.lastSuccess
	metrics.LastFailure = o.lastFailure
	o.mu.RUnlock()
	return metrics
}

func (o *FiatConversionOrchestrator) Readiness(ctx context.Context) FiatConversionReadiness {
	result := FiatConversionReadiness{Enabled: o != nil && o.cfg.Enabled}
	if !result.Enabled {
		result.Reason = "disabled"
		return result
	}
	o.mu.RLock()
	result.Started = o.running
	result.StoreReady = o.storeReady
	token := o.leaseToken
	o.mu.RUnlock()
	result.LeaseHeld = result.Started && o.cfg.Lease.Held(ctx, o.leaseName, token)
	result.ProfilesReady = o.cfg.Profiles != nil
	result.QueryReady = o.cfg.Query != nil
	result.SubmitterReady = o.cfg.Submitter != nil && o.cfg.Submitter.Readiness(ctx).Ready
	result.DEXReady = o.cfg.DEX != nil
	result.CustodyReady = o.cfg.Custody != nil && o.cfg.Custody.ProductionReady(ctx) == nil
	result.OfframpReady = o.cfg.Offramp != nil
	result.ResolversReady = o.cfg.Destination != nil && o.cfg.Destination.ProductionReady(ctx) == nil && o.cfg.Compliance != nil && o.cfg.Compliance.ProductionReady(ctx) == nil
	result.WebhookReady = !o.cfg.Production || (o.cfg.WebhookEvents != nil && o.cfg.WebhookEvents.Durable() && o.cfg.WebhookBindings != nil && o.cfg.WebhookBindings.Durable())
	if o.cfg.Profiles != nil {
		result.ExternallyBlocked = o.cfg.Profiles.DEX.State == dex.RouteEngineeringCompleteExternalBlocked || o.cfg.Profiles.Payout.State == offramp.ProfileEngineeringCompleteExternalBlocked
	}
	result.Metrics = o.Metrics(ctx)
	result.Ready = result.Started && result.StoreReady && result.LeaseHeld && result.ProfilesReady && result.QueryReady && result.SubmitterReady && result.DEXReady && result.CustodyReady && result.OfframpReady && result.ResolversReady && result.WebhookReady && !result.ExternallyBlocked
	switch {
	case !result.Started:
		result.Reason = "not started"
	case !result.StoreReady:
		result.Reason = "store unavailable"
	case !result.LeaseHeld:
		result.Reason = "lease not held"
	case result.ExternallyBlocked:
		result.Reason = "profiles externally blocked"
	case !result.CustodyReady:
		result.Reason = "DEX custody signer unavailable"
	case !result.SubmitterReady:
		result.Reason = "mutation submitter unavailable"
	case !result.QueryReady:
		result.Reason = "query unavailable"
	case !result.ResolversReady:
		result.Reason = "destination/compliance resolver unavailable"
	case !result.WebhookReady:
		result.Reason = "webhook durable stores unavailable"
	}
	return result
}

type fiatRetryError struct{ cause error }

func (e *fiatRetryError) Error() string { return fmt.Sprintf("fiat conversion retry: %v", e.cause) }
func (e *fiatRetryError) Unwrap() error { return e.cause }

type fiatTerminalError struct{ cause error }

func (e *fiatTerminalError) Error() string {
	return fmt.Sprintf("fiat conversion manual review: %v", e.cause)
}
func (e *fiatTerminalError) Unwrap() error { return e.cause }
func isFiatTerminalError(err error) bool {
	var terminal *fiatTerminalError
	return errors.As(err, &terminal)
}

func snapshotFiatIntent(record *settlementv1.FiatConversionRecord) FiatConversionIntentSnapshot {
	return FiatConversionIntentSnapshot{ConversionID: record.ConversionId, InvoiceID: record.InvoiceId, SettlementID: record.SettlementId, PayoutID: record.PayoutId, Provider: record.Provider, Customer: record.Customer, CryptoToken: record.CryptoToken, StableToken: record.StableToken, CryptoAmount: record.CryptoAmount.String(), FiatCurrency: record.FiatCurrency, PaymentMethod: record.PaymentMethod, DestinationHash: record.DestinationHash, DestinationRegion: record.DestinationRegion, SlippageToleranceExact: record.SlippageToleranceExact, RequestDigest: hex.EncodeToString(record.RequestDigest), ComplianceDigest: hex.EncodeToString(record.ComplianceDecisionHash)}
}
func fiatIntentEqual(left, right FiatConversionIntentSnapshot) bool {
	return left == right
}
func tokenFromProto(value settlementv1.TokenSpec) (dex.Token, error) {
	decimals, err := checkedTokenDecimals(value.Decimals)
	if err != nil {
		return dex.Token{}, err
	}
	return dex.Token{Symbol: value.Symbol, Denom: value.Denom, Decimals: decimals, ChainID: value.ChainId, IsNative: true}, nil
}

func checkedTokenDecimals(value uint32) (uint8, error) {
	if value > 18 {
		return 0, dex.ErrTokenDecimals
	}
	return uint8(value), nil //nolint:gosec // protocol bounded to 18 above.
}
func slippageExact(exact string, legacy float64) (sdkmath.LegacyDec, error) {
	if strings.TrimSpace(exact) == "" {
		exact = strconv.FormatFloat(legacy, 'f', -1, 64)
	}
	return sdkmath.LegacyNewDecFromStr(exact)
}
func canonicalJSONHash(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(raw)
	return digest[:], nil
}
func hashEvidence(domain string, parts ...[]byte) []byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte(domain))
	for _, part := range parts {
		length := []byte{byte(len(part) >> 24), byte(len(part) >> 16), byte(len(part) >> 8), byte(len(part))}
		_, _ = hash.Write(length)
		_, _ = hash.Write(part)
	}
	return hash.Sum(nil)
}
func mustHex32(value string) []byte {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size {
		return nil
	}
	return decoded
}
func conversionCorrelation(id string) string {
	return "fiat-" + hex.EncodeToString(hashEvidence("correlation", []byte(id))[:16])
}
func payoutMetadata(item *FiatConversionWorkItem) map[string]string {
	return map[string]string{"idempotency_key": conversionCorrelation(item.Intent.ConversionID) + "-payout", "correlation_id": conversionCorrelation(item.Intent.ConversionID), "conversion_id": item.Intent.ConversionID}
}
func validateResolvedCompliance(chain *settlementv1.FiatConversionRecord, resolved ResolvedFiatCompliance, now time.Time) error {
	decision := resolved.Decision
	if decision.Revoked || decision.Reference == "" || decision.KYCDecision != "approved" || decision.SanctionsDecision != "approved" || !decision.ValidUntil.After(now) || len(chain.ComplianceDecisionHash) != sha256.Size || !bytes.Equal(chain.ComplianceDecisionHash, resolved.Digest) {
		return offramp.ErrComplianceRequired
	}
	return nil
}

func validateResolvedDestination(chain *settlementv1.FiatConversionRecord, resolved ResolvedFiatDestination) error {
	if strings.TrimSpace(resolved.Reference) == "" || len(resolved.Digest) != sha256.Size ||
		!strings.EqualFold(chain.DestinationHash, hex.EncodeToString(resolved.Digest)) {
		return errors.New("resolved destination does not match on-chain commitment")
	}
	return nil
}
func clearString(value *string) {
	if value != nil {
		*value = ""
	}
}

func payoutQuoteSnapshot(quote offramp.Quote, requestDigest, corridorID, providerBinding, complianceDigest, profileDigest string) FiatPayoutQuoteSnapshot {
	return FiatPayoutQuoteSnapshot{
		ID: quote.ID, FiatAmount: quote.FiatAmount.String(), ExchangeRate: quote.ExchangeRate.String(),
		Fee: quote.Fee.String(), Provider: quote.Provider, CreatedAt: quote.CreatedAt, ExpiresAt: quote.ExpiresAt,
		RequestDigest: requestDigest, CorridorID: corridorID, ProviderBinding: providerBinding,
		ComplianceDigest: complianceDigest, ProfileDigest: profileDigest,
	}
}

func reconstructPayoutQuote(snapshot FiatPayoutQuoteSnapshot, chain *settlementv1.FiatConversionRecord, item *FiatConversionWorkItem, destination string, decision offramp.ComplianceDecision, profiles *TrustedFiatProfiles) (offramp.Quote, error) {
	fiatAmount, err := sdkmath.LegacyNewDecFromStr(snapshot.FiatAmount)
	if err != nil || !fiatAmount.IsPositive() {
		return offramp.Quote{}, offramp.ErrInvalidRequest
	}
	exchangeRate, err := sdkmath.LegacyNewDecFromStr(snapshot.ExchangeRate)
	if err != nil || !exchangeRate.IsPositive() {
		return offramp.Quote{}, offramp.ErrInvalidRequest
	}
	fee, ok := sdkmath.NewIntFromString(snapshot.Fee)
	if !ok || fee.IsNegative() {
		return offramp.Quote{}, offramp.ErrInvalidRequest
	}
	amount, ok := sdkmath.NewIntFromString(item.SwapConfirmation.StableAmount)
	if !ok || !amount.IsPositive() {
		return offramp.Quote{}, offramp.ErrInvalidRequest
	}
	decimals, err := checkedTokenDecimals(chain.StableToken.Decimals)
	if err != nil {
		return offramp.Quote{}, err
	}
	quote := offramp.Quote{
		ID: snapshot.ID, Provider: snapshot.Provider, FiatAmount: fiatAmount, ExchangeRate: exchangeRate,
		Fee: fee, CreatedAt: snapshot.CreatedAt, ExpiresAt: snapshot.ExpiresAt,
		Request: offramp.QuoteRequest{
			CryptoSymbol: chain.StableToken.Symbol, CryptoDenom: chain.StableToken.Denom,
			CryptoDecimals: decimals, CryptoAmount: amount,
			FiatCurrency: chain.FiatCurrency, PaymentMethod: chain.PaymentMethod, Sender: chain.Provider,
			Destination: destination, BeneficiaryReference: destination, Jurisdiction: chain.DestinationRegion,
			CorrelationID: conversionCorrelation(item.Intent.ConversionID), Compliance: decision,
		},
	}
	quoteHash, requestHash, corridorID, providerBinding, err := canonicalPayoutQuoteCommitments(quote, profiles, item.Intent.ComplianceDigest)
	if err != nil || !bytes.Equal(quoteHash, mustHex32(item.PayoutQuoteDigest)) || requestHash != snapshot.RequestDigest ||
		corridorID != snapshot.CorridorID || providerBinding != snapshot.ProviderBinding || snapshot.ComplianceDigest != item.Intent.ComplianceDigest ||
		snapshot.ProfileDigest != item.PayoutProfileDigest {
		return offramp.Quote{}, offramp.ErrInvalidRequest
	}
	return quote, nil
}

func canonicalPayoutQuoteCommitments(quote offramp.Quote, profiles *TrustedFiatProfiles, complianceDigest string) ([]byte, string, string, string, error) {
	if profiles == nil || quote.Provider != profiles.Payout.Provider || len(complianceDigest) != sha256.Size*2 {
		return nil, "", "", "", offramp.ErrInvalidRequest
	}
	corridor, err := profiles.Payout.Corridor(quote.Request.Jurisdiction, quote.Request.FiatCurrency, quote.Request.PaymentMethod)
	if err != nil {
		return nil, "", "", "", err
	}
	quoteHash, err := canonicalJSONHash(quote)
	if err != nil {
		return nil, "", "", "", err
	}
	requestHash, err := canonicalJSONHash(quote.Request)
	if err != nil {
		return nil, "", "", "", err
	}
	providerView := struct {
		ID, Provider, FiatAmount, ExchangeRate, Fee, CreatedAt, ExpiresAt string
	}{quote.ID, quote.Provider, quote.FiatAmount.String(), quote.ExchangeRate.String(), quote.Fee.String(), quote.CreatedAt.UTC().Format(time.RFC3339Nano), quote.ExpiresAt.UTC().Format(time.RFC3339Nano)}
	providerHash, err := canonicalJSONHash(providerView)
	if err != nil {
		return nil, "", "", "", err
	}
	return quoteHash, hex.EncodeToString(requestHash), corridor.ID, hex.EncodeToString(providerHash), nil
}

func validatePayoutResultBinding(item *FiatConversionWorkItem, result offramp.PayoutResult) error {
	if item == nil || result.ID == "" || result.Provider != item.PayoutQuote.Provider || result.QuoteID != item.PayoutQuote.ID ||
		result.FiatAmount.String() != item.PayoutQuote.FiatAmount || result.CryptoAmount.String() != item.SwapConfirmation.StableAmount ||
		result.Fee.String() != item.PayoutQuote.Fee || result.Reference == "" || result.StatusUpdatedAt.Before(result.InitiatedAt) ||
		(item.PayoutID != "" && result.ID != item.PayoutID) || !safeMetadataMatch(result.Metadata, payoutMetadata(item)) {
		return offramp.ErrProviderRejected
	}
	return nil
}

func safeFailureObservation(item *FiatConversionWorkItem, code string) bool {
	if item == nil || item.PendingObservation != nil {
		return false
	}
	if code == "PAYOUT_TERMINAL_FAILURE" && item.PayoutID != "" {
		return true
	}
	switch item.State {
	case FiatWorkClaimed, FiatWorkQuoting, FiatWorkQuoteReported, FiatWorkSigning, FiatWorkSwapFinalized, FiatWorkPayoutQuote:
		return true
	default:
		return false
	}
}

func boundedFiatFailureCode(value string) string {
	value = strings.ToUpper(value)
	var result strings.Builder
	for _, char := range value {
		if (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' {
			result.WriteRune(char)
		}
		if result.Len() >= 64 {
			break
		}
	}
	if result.Len() == 0 {
		return "MANUAL_REVIEW"
	}
	return result.String()
}

var _ FiatObservationSubmitter = (*ProviderMutationSubmitter)(nil)
var _ FiatDEX = AdapterFiatDEX{}
