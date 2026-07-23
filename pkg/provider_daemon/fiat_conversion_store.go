// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/dex"
	settlementv1 "github.com/virtengine/virtengine/sdk/go/node/settlement/v1"
)

const fiatConversionStoreSchemaVersion uint32 = 2

var (
	ErrFiatConversionWorkNotFound = errors.New("fiat conversion work item not found")
	ErrFiatConversionLeaseLost    = errors.New("fiat conversion lease lost")
)

// FiatConversionWorkState is the durable off-chain execution state machine.
type FiatConversionWorkState string

const (
	FiatWorkClaimed         FiatConversionWorkState = "claimed"
	FiatWorkQuoting         FiatConversionWorkState = "quoting"
	FiatWorkQuoteReported   FiatConversionWorkState = "quote-reported"
	FiatWorkSigning         FiatConversionWorkState = "signing"
	FiatWorkSwapBroadcast   FiatConversionWorkState = "swap-broadcast"
	FiatWorkAmbiguous       FiatConversionWorkState = "ambiguous"
	FiatWorkSwapFinalized   FiatConversionWorkState = "swap-finalized"
	FiatWorkPayoutQuote     FiatConversionWorkState = "payout-quote"
	FiatWorkPayoutSubmitted FiatConversionWorkState = "payout-submitted"
	FiatWorkPayoutAmbiguous FiatConversionWorkState = "payout-ambiguous"
	FiatWorkCompleted       FiatConversionWorkState = "completed"
	FiatWorkFailed          FiatConversionWorkState = "failed"
	FiatWorkManualReview    FiatConversionWorkState = "manual-review"
)

func (s FiatConversionWorkState) terminal() bool {
	return s == FiatWorkCompleted || s == FiatWorkFailed || s == FiatWorkManualReview
}

func validFiatWorkTransition(from, to FiatConversionWorkState) bool {
	if from == to {
		return true
	}
	if from.terminal() {
		return false
	}
	if to == FiatWorkFailed || to == FiatWorkManualReview {
		return true
	}
	if to == FiatWorkClaimed {
		return from == FiatWorkQuoting || from == FiatWorkQuoteReported
	}
	rank := map[FiatConversionWorkState]int{
		FiatWorkClaimed: 1, FiatWorkQuoting: 2, FiatWorkQuoteReported: 3, FiatWorkSigning: 4,
		FiatWorkSwapBroadcast: 5, FiatWorkAmbiguous: 5, FiatWorkSwapFinalized: 6,
		FiatWorkPayoutQuote: 7, FiatWorkPayoutSubmitted: 8, FiatWorkPayoutAmbiguous: 8, FiatWorkCompleted: 9,
	}
	fromRank, fromOK := rank[from]
	toRank, toOK := rank[to]
	return fromOK && toOK && toRank >= fromRank
}

// FiatConversionAttempt is bounded and contains no payload or destination.
type FiatConversionAttempt struct {
	Number         uint32    `json:"number"`
	Stage          string    `json:"stage"`
	StartedAt      time.Time `json:"started_at"`
	FinishedAt     time.Time `json:"finished_at,omitempty"`
	Classification string    `json:"classification,omitempty"`
	Outcome        string    `json:"outcome,omitempty"`
}

// FiatConversionIntentSnapshot is a privacy-safe immutable intent copy.
// DestinationRef is deliberately omitted; only its already-public hash remains.
type FiatConversionIntentSnapshot struct {
	ConversionID           string                 `json:"conversion_id"`
	InvoiceID              string                 `json:"invoice_id,omitempty"`
	SettlementID           string                 `json:"settlement_id"`
	PayoutID               string                 `json:"payout_id"`
	Provider               string                 `json:"provider"`
	Customer               string                 `json:"customer"`
	CryptoToken            settlementv1.TokenSpec `json:"crypto_token"`
	StableToken            settlementv1.TokenSpec `json:"stable_token"`
	CryptoAmount           string                 `json:"crypto_amount"`
	FiatCurrency           string                 `json:"fiat_currency"`
	PaymentMethod          string                 `json:"payment_method"`
	DestinationHash        string                 `json:"destination_hash"`
	DestinationRegion      string                 `json:"destination_region"`
	SlippageToleranceExact string                 `json:"slippage_tolerance_exact"`
	RequestDigest          string                 `json:"request_digest"`
	ComplianceDigest       string                 `json:"compliance_digest"`
}

// FiatConversionConfirmation persists finality evidence needed after restart.
type FiatConversionConfirmation struct {
	Height        int64  `json:"height"`
	BlockHash     string `json:"block_hash"`
	Confirmations uint32 `json:"confirmations"`
	FinalityHash  string `json:"finality_hash"`
	StableAmount  string `json:"stable_amount"`
}

// FiatConversionWorkItem persists hashes and external identifiers, never
// credentials, raw execution payloads, TxRaw bytes, beneficiary tokens or PII.
type FiatConversionWorkItem struct {
	SchemaVersion uint32                       `json:"schema_version"`
	Intent        FiatConversionIntentSnapshot `json:"intent"`
	State         FiatConversionWorkState      `json:"state"`

	DEXProfileID        string `json:"dex_profile_id"`
	DEXProfileDigest    string `json:"dex_profile_digest"`
	PayoutProfileID     string `json:"payout_profile_id"`
	PayoutProfileDigest string `json:"payout_profile_digest"`

	DEXQuote         *dex.SwapQuote             `json:"dex_quote,omitempty"`
	QuoteDigest      string                     `json:"quote_digest,omitempty"`
	PayloadHash      string                     `json:"payload_hash,omitempty"`
	SignedTxHash     string                     `json:"signed_tx_hash,omitempty"`
	SwapTxHash       string                     `json:"swap_tx_hash,omitempty"`
	SwapConfirmation FiatConversionConfirmation `json:"swap_confirmation,omitempty"`

	PayoutQuote         FiatPayoutQuoteSnapshot `json:"payout_quote,omitempty"`
	PayoutQuoteDigest   string                  `json:"payout_quote_digest,omitempty"`
	PayoutID            string                  `json:"payout_id,omitempty"`
	PayoutStatus        string                  `json:"payout_status,omitempty"`
	PayoutReferenceHash string                  `json:"payout_reference_hash,omitempty"`
	PayoutFinalityHash  string                  `json:"payout_finality_hash,omitempty"`

	ObservationSequence    uint64                                           `json:"observation_sequence"`
	ObservationDigest      string                                           `json:"observation_digest,omitempty"`
	PendingObservation     *settlementv1.MsgRecordFiatConversionObservation `json:"pending_observation,omitempty"`
	PendingNextState       FiatConversionWorkState                          `json:"pending_next_state,omitempty"`
	PendingWebhookProvider string                                           `json:"pending_webhook_provider,omitempty"`
	PendingWebhookEventID  string                                           `json:"pending_webhook_event_id,omitempty"`
	Attempts               []FiatConversionAttempt                          `json:"attempts,omitempty"`
	AttemptCount           uint32                                           `json:"attempt_count"`
	NextRetryAt            time.Time                                        `json:"next_retry_at,omitempty"`
	CreatedAt              time.Time                                        `json:"created_at"`
	UpdatedAt              time.Time                                        `json:"updated_at"`
	FinalityStartedAt      time.Time                                        `json:"finality_started_at,omitempty"`
	TerminalResult         string                                           `json:"terminal_result,omitempty"`
	FailureCode            string                                           `json:"failure_code,omitempty"`
	LeaseToken             uint64                                           `json:"lease_token"`
}

// FiatPayoutQuoteSnapshot contains only economic terms and immutable external
// IDs. The original QuoteRequest is intentionally excluded because it contains
// destination and compliance references.
type FiatPayoutQuoteSnapshot struct {
	ID               string    `json:"id"`
	FiatAmount       string    `json:"fiat_amount"`
	ExchangeRate     string    `json:"exchange_rate"`
	Fee              string    `json:"fee"`
	Provider         string    `json:"provider"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	RequestDigest    string    `json:"request_digest"`
	CorridorID       string    `json:"corridor_id"`
	ProviderBinding  string    `json:"provider_binding"`
	ComplianceDigest string    `json:"compliance_digest"`
	ProfileDigest    string    `json:"profile_digest"`
}

// FiatConversionStore is the Task 85C-extensible durable state boundary.
type FiatConversionStore interface {
	Open(context.Context) error
	Close() error
	PutIfAbsent(context.Context, *FiatConversionWorkItem) (*FiatConversionWorkItem, bool, error)
	Get(context.Context, string) (*FiatConversionWorkItem, error)
	Update(context.Context, string, func(*FiatConversionWorkItem) error) (*FiatConversionWorkItem, error)
	List(context.Context) ([]*FiatConversionWorkItem, error)
	Durable() bool
}

type fiatConversionFileState struct {
	SchemaVersion uint32                             `json:"schema_version"`
	Items         map[string]*FiatConversionWorkItem `json:"items"`
}

// FileFiatConversionStore uses an exclusive process lock and atomic replacement.
type FileFiatConversionStore struct {
	path  string
	mu    sync.RWMutex
	state fiatConversionFileState
	lock  *txSubmissionQueuePathLock
	open  bool
	now   func() time.Time
}

func NewFileFiatConversionStore(path string) (*FileFiatConversionStore, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("fiat conversion store path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err := validateStatePath(absolute); err != nil {
		return nil, err
	}
	return &FileFiatConversionStore{path: filepath.Clean(absolute), now: func() time.Time { return time.Now().UTC() }}, nil
}

func (s *FileFiatConversionStore) setClock(now func() time.Time) {
	if s != nil && now != nil {
		s.mu.Lock()
		s.now = now
		s.mu.Unlock()
	}
}

func (s *FileFiatConversionStore) Open(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.open {
		return nil
	}
	lock, err := claimTxSubmissionQueuePath(s.path)
	if err != nil {
		return err
	}
	state := fiatConversionFileState{SchemaVersion: fiatConversionStoreSchemaVersion, Items: make(map[string]*FiatConversionWorkItem)}
	raw, readErr := os.ReadFile(s.path) // #nosec G304 -- constructor validates path.
	if readErr == nil {
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&state); err != nil {
			lock.release()
			return fmt.Errorf("decode fiat conversion store: %w", err)
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			lock.release()
			return errors.New("decode fiat conversion store: multiple JSON values")
		}
		if (state.SchemaVersion != 1 && state.SchemaVersion != fiatConversionStoreSchemaVersion) || state.Items == nil {
			lock.release()
			return fmt.Errorf("unsupported fiat conversion store schema")
		}
		legacy := state.SchemaVersion == 1
		state.SchemaVersion = fiatConversionStoreSchemaVersion
		for id, item := range state.Items {
			if item == nil || item.Intent.ConversionID != id {
				lock.release()
				return fmt.Errorf("invalid fiat conversion work item %q", id)
			}
			if legacy {
				item.SchemaVersion = fiatConversionStoreSchemaVersion
			}
			if err := validateFiatWorkItem(item); err != nil {
				lock.release()
				return fmt.Errorf("invalid fiat conversion work item %q: %w", id, err)
			}
		}
	} else if !os.IsNotExist(readErr) {
		lock.release()
		return readErr
	}
	s.state, s.lock, s.open = state, lock, true
	return nil
}

func (s *FileFiatConversionStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lock != nil {
		s.lock.release()
	}
	s.lock, s.open = nil, false
	return nil
}

func cloneFiatWork(item *FiatConversionWorkItem) *FiatConversionWorkItem {
	if item == nil {
		return nil
	}
	raw, err := json.Marshal(item)
	if err != nil {
		return nil
	}
	var result FiatConversionWorkItem
	if json.Unmarshal(raw, &result) != nil {
		return nil
	}
	return &result
}

func (s *FileFiatConversionStore) PutIfAbsent(_ context.Context, item *FiatConversionWorkItem) (*FiatConversionWorkItem, bool, error) {
	if item == nil || item.Intent.ConversionID == "" {
		return nil, false, errors.New("conversion identity is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.open {
		return nil, false, ErrProviderMutationUnavailable
	}
	if existing, ok := s.state.Items[item.Intent.ConversionID]; ok {
		return cloneFiatWork(existing), true, nil
	}
	item.SchemaVersion = fiatConversionStoreSchemaVersion
	if err := validateFiatWorkItem(item); err != nil {
		return nil, false, err
	}
	s.state.Items[item.Intent.ConversionID] = cloneFiatWork(item)
	if err := s.saveLocked(); err != nil {
		delete(s.state.Items, item.Intent.ConversionID)
		return nil, false, err
	}
	return cloneFiatWork(item), false, nil
}

func (s *FileFiatConversionStore) Get(_ context.Context, id string) (*FiatConversionWorkItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.open {
		return nil, ErrProviderMutationUnavailable
	}
	item, ok := s.state.Items[id]
	if !ok {
		return nil, ErrFiatConversionWorkNotFound
	}
	return cloneFiatWork(item), nil
}

func (s *FileFiatConversionStore) Update(_ context.Context, id string, fn func(*FiatConversionWorkItem) error) (*FiatConversionWorkItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.open {
		return nil, ErrProviderMutationUnavailable
	}
	current, ok := s.state.Items[id]
	if !ok {
		return nil, ErrFiatConversionWorkNotFound
	}
	candidate := cloneFiatWork(current)
	if candidate == nil {
		return nil, errors.New("fiat conversion work item cannot be cloned")
	}
	if err := fn(candidate); err != nil {
		return nil, err
	}
	if !validFiatWorkTransition(current.State, candidate.State) {
		return nil, fmt.Errorf("invalid fiat conversion work transition %s -> %s", current.State, candidate.State)
	}
	candidate.SchemaVersion = fiatConversionStoreSchemaVersion
	candidate.UpdatedAt = s.now().UTC()
	if err := validateFiatWorkItem(candidate); err != nil {
		return nil, err
	}
	s.state.Items[id] = candidate
	if err := s.saveLocked(); err != nil {
		s.state.Items[id] = current
		return nil, err
	}
	return cloneFiatWork(candidate), nil
}

func (s *FileFiatConversionStore) List(context.Context) ([]*FiatConversionWorkItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.open {
		return nil, ErrProviderMutationUnavailable
	}
	result := make([]*FiatConversionWorkItem, 0, len(s.state.Items))
	for _, item := range s.state.Items {
		result = append(result, cloneFiatWork(item))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Intent.ConversionID < result[j].Intent.ConversionID })
	return result, nil
}

func (*FileFiatConversionStore) Durable() bool { return true }

func (s *FileFiatConversionStore) saveLocked() error {
	raw, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".fiat-conversion-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := atomicReplaceFile(tmpPath, s.path); err != nil {
		return err
	}
	if directory, err := os.Open(filepath.Dir(s.path)); err == nil { // #nosec G304 -- validated path.
		_ = directory.Sync()
		_ = directory.Close()
	}
	return nil
}

// LocalFiatConversionLease is Task 85B's single-process implementation. A
// distributed implementation can replace it in Task 85C without state changes.
type FiatConversionLease interface {
	Acquire(context.Context, string, time.Duration) (uint64, error)
	Renew(context.Context, string, uint64, time.Duration) error
	Release(context.Context, string, uint64) error
	Held(context.Context, string, uint64) bool
}

type LocalFiatConversionLease struct{ inner *LocalSubmitterLease }

func NewLocalFiatConversionLease() *LocalFiatConversionLease {
	return &LocalFiatConversionLease{inner: NewLocalSubmitterLease()}
}

func (l *LocalFiatConversionLease) Acquire(ctx context.Context, name string, ttl time.Duration) (uint64, error) {
	return l.inner.Acquire(ctx, name, ttl)
}
func (l *LocalFiatConversionLease) Renew(ctx context.Context, name string, token uint64, ttl time.Duration) error {
	return l.inner.Renew(ctx, name, token, ttl)
}
func (l *LocalFiatConversionLease) Release(ctx context.Context, name string, token uint64) error {
	return l.inner.Release(ctx, name, token)
}
func (l *LocalFiatConversionLease) Held(ctx context.Context, name string, token uint64) bool {
	return l.inner.Held(ctx, name, token)
}

func appendFiatAttempt(attempts []FiatConversionAttempt, attempt FiatConversionAttempt) []FiatConversionAttempt {
	const maximum = 32
	attempts = append(attempts, attempt)
	if len(attempts) > maximum {
		attempts = append([]FiatConversionAttempt(nil), attempts[len(attempts)-maximum:]...)
	}
	return attempts
}

func digestHex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func validateFiatWorkItem(item *FiatConversionWorkItem) error {
	if item == nil || item.SchemaVersion != fiatConversionStoreSchemaVersion || strings.TrimSpace(item.Intent.ConversionID) == "" ||
		len(item.Intent.RequestDigest) != sha256.Size*2 || len(item.Intent.ComplianceDigest) != sha256.Size*2 ||
		len(item.DEXProfileDigest) != sha256.Size*2 || len(item.PayoutProfileDigest) != sha256.Size*2 ||
		item.CreatedAt.IsZero() || item.UpdatedAt.IsZero() || item.UpdatedAt.Before(item.CreatedAt) {
		return errors.New("incomplete work item identity or commitments")
	}
	if item.PendingObservation == nil != (item.PendingNextState == "") {
		return errors.New("pending observation and next state must be stored together")
	}
	if (item.PendingWebhookProvider == "") != (item.PendingWebhookEventID == "") {
		return errors.New("pending webhook identity is detached from completion observation")
	}
	if item.PendingWebhookEventID != "" && item.State != FiatWorkPayoutSubmitted && item.State != FiatWorkPayoutAmbiguous && item.State != FiatWorkCompleted {
		return errors.New("pending webhook identity is invalid for work state")
	}
	if item.PendingObservation != nil && (item.PendingObservation.ConversionId != item.Intent.ConversionID || item.PendingObservation.ObservationSequence != item.ObservationSequence+1) {
		return errors.New("pending observation sequence or conversion mismatch")
	}
	if item.AttemptCount > 1_000_000 || len(item.Attempts) > 32 {
		return errors.New("attempt history exceeds safety bound")
	}
	if item.DEXQuote != nil {
		digest, err := dex.QuoteDigest(*item.DEXQuote)
		if err != nil || item.DEXQuote.ID != digest || item.DEXQuote.QuoteDigest != digest || len(item.QuoteDigest) != sha256.Size*2 {
			return fmt.Errorf("invalid canonical DEX quote commitment id=%q stored=%q computed=%q", item.DEXQuote.ID, item.QuoteDigest, digest)
		}
	}
	if item.PayoutQuote.ID != "" {
		for _, digest := range []string{item.PayoutQuoteDigest, item.PayoutQuote.RequestDigest, item.PayoutQuote.ProviderBinding, item.PayoutQuote.ComplianceDigest, item.PayoutQuote.ProfileDigest} {
			if len(digest) != sha256.Size*2 {
				return errors.New("invalid payout quote commitment")
			}
		}
		if item.PayoutQuote.CorridorID == "" || item.PayoutQuote.Provider == "" || !item.PayoutQuote.ExpiresAt.After(item.PayoutQuote.CreatedAt) {
			return errors.New("invalid payout quote identity")
		}
	}
	return nil
}

var _ FiatConversionStore = (*FileFiatConversionStore)(nil)
var _ FiatConversionLease = (*LocalFiatConversionLease)(nil)
