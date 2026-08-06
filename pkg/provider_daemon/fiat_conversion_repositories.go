// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"crypto/hmac"
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

	sdkmath "cosmossdk.io/math"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
)

const fiatRepositorySchemaVersion uint32 = 3

type persistedWebhookEvent struct {
	Event      offramp.WebhookEvent `json:"event"`
	Digest     string               `json:"digest"`
	ConsumedAt *time.Time           `json:"consumed_at,omitempty"`
}

type fiatRepositoryState struct {
	SchemaVersion  uint32                                    `json:"schema_version"`
	Payouts        map[string]offramp.PayoutResult           `json:"payouts"`
	Limits         map[string]string                         `json:"limits"`
	Reservations   map[string]string                         `json:"reservations"`
	WebhookEvents  map[string]string                         `json:"webhook_events"`
	Bindings       map[string]offramp.WebhookBinding         `json:"bindings"`
	VerifiedEvents map[string]persistedWebhookEvent          `json:"verified_events"`
	Initiations    map[string]offramp.PayoutInitiationRecord `json:"initiations"`
}

// FileFiatRepository implements all durable off-ramp repository contracts in
// one atomically replaced, exclusively locked state file.
type FileFiatRepository struct {
	path  string
	mu    sync.Mutex
	state fiatRepositoryState
	lock  *txSubmissionQueuePathLock
	open  bool
}

func NewFileFiatRepository(path string) (*FileFiatRepository, error) {
	absolute, err := filepath.Abs(path)
	if err != nil || path == "" {
		return nil, errors.New("fiat repository path is required")
	}
	if err := validateStatePath(absolute); err != nil {
		return nil, err
	}
	return &FileFiatRepository{path: filepath.Clean(absolute)}, nil
}

func (r *FileFiatRepository) Open(context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.open {
		return nil
	}
	lock, err := claimTxSubmissionQueuePath(r.path)
	if err != nil {
		return err
	}
	state := fiatRepositoryState{SchemaVersion: fiatRepositorySchemaVersion, Payouts: map[string]offramp.PayoutResult{}, Limits: map[string]string{}, Reservations: map[string]string{}, WebhookEvents: map[string]string{}, Bindings: map[string]offramp.WebhookBinding{}, VerifiedEvents: map[string]persistedWebhookEvent{}, Initiations: map[string]offramp.PayoutInitiationRecord{}}
	if raw, readErr := os.ReadFile(r.path); readErr == nil { // #nosec G304 -- validated path.
		var header struct {
			SchemaVersion uint32 `json:"schema_version"`
		}
		if err := decodeStrictFiatRepositoryJSON(raw, &header, false); err != nil {
			lock.release()
			return err
		}
		switch header.SchemaVersion {
		case 1:
			var legacy struct {
				SchemaVersion  uint32                            `json:"schema_version"`
				Payouts        map[string]offramp.PayoutResult   `json:"payouts"`
				Limits         map[string]string                 `json:"limits"`
				Reservations   map[string]string                 `json:"reservations"`
				WebhookEvents  map[string]string                 `json:"webhook_events"`
				Bindings       map[string]offramp.WebhookBinding `json:"bindings"`
				VerifiedEvents map[string]offramp.WebhookEvent   `json:"verified_events"`
			}
			if err := decodeStrictFiatRepositoryJSON(raw, &legacy, true); err != nil {
				lock.release()
				return err
			}
			state.Payouts, state.Limits, state.Reservations = legacy.Payouts, legacy.Limits, legacy.Reservations
			state.WebhookEvents, state.Bindings = legacy.WebhookEvents, legacy.Bindings
			for key, event := range legacy.VerifiedEvents {
				rawEvent, err := json.Marshal(event)
				if err != nil {
					lock.release()
					return err
				}
				digest := sha256.Sum256(rawEvent)
				state.VerifiedEvents[key] = persistedWebhookEvent{Event: event, Digest: hex.EncodeToString(digest[:])}
			}
		case 2:
			var legacy fiatRepositoryState
			if err := decodeStrictFiatRepositoryJSON(raw, &legacy, true); err != nil {
				lock.release()
				return err
			}
			state = legacy
			state.SchemaVersion = fiatRepositorySchemaVersion
			state.Initiations = map[string]offramp.PayoutInitiationRecord{}
		default:
			if err := decodeStrictFiatRepositoryJSON(raw, &state, true); err != nil {
				lock.release()
				return err
			}
			if state.SchemaVersion != fiatRepositorySchemaVersion {
				lock.release()
				return errors.New("unsupported fiat repository schema")
			}
		}
	} else if !os.IsNotExist(readErr) {
		lock.release()
		return readErr
	}
	if state.Payouts == nil {
		state.Payouts = map[string]offramp.PayoutResult{}
	}
	if state.Limits == nil {
		state.Limits = map[string]string{}
	}
	if state.Reservations == nil {
		state.Reservations = map[string]string{}
	}
	if state.WebhookEvents == nil {
		state.WebhookEvents = map[string]string{}
	}
	if state.Bindings == nil {
		state.Bindings = map[string]offramp.WebhookBinding{}
	}
	if state.VerifiedEvents == nil {
		state.VerifiedEvents = map[string]persistedWebhookEvent{}
	}
	if state.Initiations == nil {
		state.Initiations = map[string]offramp.PayoutInitiationRecord{}
	}
	r.state, r.lock, r.open = state, lock, true
	return nil
}

func decodeStrictFiatRepositoryJSON(raw []byte, target any, disallowUnknown bool) error {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	if disallowUnknown {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("decode fiat repository: multiple JSON values")
	}
	return nil
}

func (r *FileFiatRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.lock != nil {
		r.lock.release()
	}
	r.open = false
	return nil
}
func (*FileFiatRepository) Durable() bool { return true }

func (r *FileFiatRepository) GetPayout(_ context.Context, id string) (offramp.PayoutResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.state.Payouts[id]
	if !ok {
		return offramp.PayoutResult{}, offramp.ErrPayoutNotFound
	}
	return cloneOfframpPayout(value), nil
}
func (r *FileFiatRepository) FindPayout(_ context.Context, provider string, metadata map[string]string) (offramp.PayoutResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, value := range r.state.Payouts {
		if provider != "" && value.Provider != provider {
			continue
		}
		if safeMetadataMatch(value.Metadata, metadata) {
			return cloneOfframpPayout(value), nil
		}
	}
	return offramp.PayoutResult{}, offramp.ErrPayoutNotFound
}
func (r *FileFiatRepository) PutPayout(_ context.Context, value offramp.PayoutResult) error {
	safeValue, err := privacySafePersistedPayout(value)
	if err != nil {
		return offramp.ErrInvalidRequest
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	before := r.state.Payouts[safeValue.ID]
	if before.ID != "" {
		if err := validatePersistedPayoutUpdate(before, safeValue); err != nil {
			return err
		}
	}
	r.state.Payouts[safeValue.ID] = cloneOfframpPayout(safeValue)
	if err := r.saveLocked(); err != nil {
		if before.ID == "" {
			delete(r.state.Payouts, safeValue.ID)
		} else {
			r.state.Payouts[safeValue.ID] = before
		}
		return err
	}
	return nil
}

func (r *FileFiatRepository) GetPayoutInitiation(_ context.Context, provider string, metadata map[string]string) (offramp.PayoutInitiationRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.state.Initiations[fiatRepositoryMetadataKey(provider, metadata)]
	if !ok {
		return offramp.PayoutInitiationRecord{}, offramp.ErrPayoutNotFound
	}
	return clonePayoutInitiation(value), nil
}

func (r *FileFiatRepository) PutPayoutInitiation(_ context.Context, value offramp.PayoutInitiationRecord) error {
	if value.Provider == "" || value.QuoteID == "" || value.OperationBinding == "" || value.RequestBinding == "" ||
		value.FiatAmount == "" || value.CryptoAmount == "" || value.Fee == "" || value.DailyReservationKey == "" ||
		value.DailyReservationOperationID == "" || value.PreparedAt.IsZero() || !safeOperationalMetadata(value.Metadata) {
		return offramp.ErrInvalidRequest
	}
	key := fiatRepositoryMetadataKey(value.Provider, value.Metadata)
	r.mu.Lock()
	defer r.mu.Unlock()
	before, existed := r.state.Initiations[key]
	if existed && !validFiatInitiationUpdate(before, value) {
		return offramp.ErrProviderRejected
	}
	r.state.Initiations[key] = clonePayoutInitiation(value)
	if err := r.saveLocked(); err != nil {
		if existed {
			r.state.Initiations[key] = before
		} else {
			delete(r.state.Initiations, key)
		}
		return err
	}
	return nil
}

func (r *FileFiatRepository) ReserveDailyAmount(_ context.Context, key, operationID string, amount, limit sdkmath.LegacyDec) (bool, error) {
	if key == "" || operationID == "" || amount.IsNil() || !amount.IsPositive() || limit.IsNil() || !limit.IsPositive() {
		return false, offramp.ErrInvalidRequest
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	reservationKey := key + "|" + operationID
	if existing, ok := r.state.Reservations[reservationKey]; ok {
		return existing == amount.String(), nil
	}
	total := sdkmath.LegacyZeroDec()
	if encoded := r.state.Limits[key]; encoded != "" {
		parsed, err := sdkmath.LegacyNewDecFromStr(encoded)
		if err != nil {
			return false, err
		}
		total = parsed
	}
	if total.Add(amount).GT(limit) {
		return false, nil
	}
	r.state.Limits[key], r.state.Reservations[reservationKey] = total.Add(amount).String(), amount.String()
	if err := r.saveLocked(); err != nil {
		delete(r.state.Reservations, reservationKey)
		r.state.Limits[key] = total.String()
		return false, err
	}
	return true, nil
}
func (r *FileFiatRepository) ReleaseDailyAmount(_ context.Context, key, operationID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	reservationKey := key + "|" + operationID
	encoded, ok := r.state.Reservations[reservationKey]
	if !ok {
		return nil
	}
	amount, err := sdkmath.LegacyNewDecFromStr(encoded)
	if err != nil {
		return err
	}
	total, err := sdkmath.LegacyNewDecFromStr(r.state.Limits[key])
	if err != nil {
		return err
	}
	remaining := total.Sub(amount)
	if remaining.IsNegative() {
		return errors.New("daily limit underflow")
	}
	delete(r.state.Reservations, reservationKey)
	r.state.Limits[key] = remaining.String()
	if err := r.saveLocked(); err != nil {
		r.state.Reservations[reservationKey] = encoded
		r.state.Limits[key] = total.String()
		return err
	}
	return nil
}

func (r *FileFiatRepository) RecordWebhookEvent(_ context.Context, provider, eventID, payloadDigest string, _ time.Time) (offramp.WebhookReplayResult, error) {
	if provider == "" || eventID == "" || payloadDigest == "" {
		return "", offramp.ErrInvalidRequest
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := provider + "|" + eventID
	existing, ok := r.state.WebhookEvents[key]
	if ok {
		if hmac.Equal([]byte(existing), []byte(payloadDigest)) {
			return offramp.WebhookReplayExact, nil
		}
		return offramp.WebhookReplayConflicting, nil
	}
	r.state.WebhookEvents[key] = payloadDigest
	if err := r.saveLocked(); err != nil {
		delete(r.state.WebhookEvents, key)
		return "", err
	}
	return offramp.WebhookReplayNew, nil
}
func (r *FileFiatRepository) LookupWebhookBinding(_ context.Context, provider, payoutID string) (offramp.WebhookBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.state.Bindings[provider+"|"+payoutID]
	if !ok {
		return offramp.WebhookBinding{}, offramp.ErrPayoutNotFound
	}
	return value, nil
}
func (r *FileFiatRepository) PutWebhookBinding(_ context.Context, value offramp.WebhookBinding) error {
	if value.Provider == "" || value.PayoutID == "" || value.QuoteID == "" || value.CorrelationID == "" {
		return offramp.ErrInvalidRequest
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := value.Provider + "|" + value.PayoutID
	if existing, ok := r.state.Bindings[key]; ok && existing != value {
		return offramp.ErrWebhookReplay
	}
	previous, existed := r.state.Bindings[key]
	r.state.Bindings[key] = value
	if err := r.saveLocked(); err != nil {
		if existed {
			r.state.Bindings[key] = previous
		} else {
			delete(r.state.Bindings, key)
		}
		return err
	}
	return nil
}

// PutVerifiedWebhookEvent durably stores the already authenticated event before
// the HTTP handler acknowledges it or wakes the orchestrator.
func (r *FileFiatRepository) PutVerifiedWebhookEvent(_ context.Context, event offramp.WebhookEvent) error {
	if event.Provider == "" || event.EventID == "" || event.PayoutID == "" {
		return offramp.ErrWebhookInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := event.Provider + "|" + event.EventID
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(eventBytes)
	digestHex := hex.EncodeToString(digest[:])
	if existing, ok := r.state.VerifiedEvents[key]; ok {
		if !hmac.Equal([]byte(existing.Digest), []byte(digestHex)) {
			return offramp.ErrWebhookReplay
		}
		return nil
	}
	r.state.VerifiedEvents[key] = persistedWebhookEvent{Event: event, Digest: digestHex}
	if err := r.saveLocked(); err != nil {
		delete(r.state.VerifiedEvents, key)
		return err
	}
	return nil
}

func (r *FileFiatRepository) VerifiedWebhookEvents(_ context.Context, payoutID string) ([]offramp.WebhookEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]offramp.WebhookEvent, 0)
	for _, stored := range r.state.VerifiedEvents {
		if stored.Event.PayoutID == payoutID && stored.ConsumedAt == nil {
			result = append(result, stored.Event)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].OccurredAt.Equal(result[j].OccurredAt) {
			return result[i].EventID < result[j].EventID
		}
		return result[i].OccurredAt.Before(result[j].OccurredAt)
	})
	return result, nil
}

// ConsumeVerifiedWebhookEvent atomically marks one event consumed. An exact
// retry is a no-op; conflicting provider/event identity is rejected.
func (r *FileFiatRepository) ConsumeVerifiedWebhookEvent(_ context.Context, provider, eventID string, consumedAt time.Time) error {
	if provider == "" || eventID == "" || consumedAt.IsZero() {
		return offramp.ErrWebhookInvalid
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := provider + "|" + eventID
	stored, ok := r.state.VerifiedEvents[key]
	if !ok {
		return offramp.ErrWebhookInvalid
	}
	if stored.ConsumedAt != nil {
		return nil
	}
	value := consumedAt.UTC()
	stored.ConsumedAt = &value
	r.state.VerifiedEvents[key] = stored
	if err := r.saveLocked(); err != nil {
		stored.ConsumedAt = nil
		r.state.VerifiedEvents[key] = stored
		return err
	}
	return nil
}

func (r *FileFiatRepository) saveLocked() error {
	raw, err := json.MarshalIndent(r.state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(r.path), ".fiat-repository-*.tmp")
	if err != nil {
		return err
	}
	path := tmp.Name()
	defer func() { _ = os.Remove(path) }()
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
	return atomicReplaceFile(path, r.path)
}

func cloneOfframpPayout(value offramp.PayoutResult) offramp.PayoutResult {
	raw, err := json.Marshal(value)
	if err != nil {
		return offramp.PayoutResult{}
	}
	var copyValue offramp.PayoutResult
	if err := json.Unmarshal(raw, &copyValue); err != nil {
		return offramp.PayoutResult{}
	}
	return copyValue
}
func clonePayoutInitiation(value offramp.PayoutInitiationRecord) offramp.PayoutInitiationRecord {
	raw, err := json.Marshal(value)
	if err != nil {
		return offramp.PayoutInitiationRecord{}
	}
	var copyValue offramp.PayoutInitiationRecord
	if err := json.Unmarshal(raw, &copyValue); err != nil {
		return offramp.PayoutInitiationRecord{}
	}
	return copyValue
}

func fiatRepositoryMetadataKey(provider string, metadata map[string]string) string {
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := []string{provider}
	for _, key := range keys {
		parts = append(parts, key+"="+metadata[key])
	}
	return strings.Join(parts, "|")
}

func safeOperationalMetadata(metadata map[string]string) bool {
	if len(metadata) == 0 || len(metadata) > 8 {
		return false
	}
	allowed := map[string]bool{"idempotency_key": true, "correlation_id": true, "conversion_id": true}
	for key, value := range metadata {
		if !allowed[key] || value == "" || len(value) > 256 || strings.ContainsAny(value, "\x00\r\n") {
			return false
		}
	}
	return metadata["idempotency_key"] != "" && metadata["correlation_id"] != ""
}

func validFiatInitiationUpdate(current, updated offramp.PayoutInitiationRecord) bool {
	if current.Provider != updated.Provider || current.QuoteID != updated.QuoteID || current.OperationBinding != updated.OperationBinding ||
		current.RequestBinding != updated.RequestBinding || current.FiatAmount != updated.FiatAmount || current.CryptoAmount != updated.CryptoAmount ||
		current.Fee != updated.Fee || current.DailyReservationKey != updated.DailyReservationKey ||
		current.DailyReservationOperationID != updated.DailyReservationOperationID || !current.PreparedAt.Equal(updated.PreparedAt) ||
		!safeMetadataMatch(current.Metadata, updated.Metadata) || !safeMetadataMatch(updated.Metadata, current.Metadata) {
		return false
	}
	if current.PayoutID != "" && current.PayoutID != updated.PayoutID {
		return false
	}
	rank := map[offramp.PayoutInitiationState]int{
		offramp.PayoutInitiationPrepared: 1, offramp.PayoutInitiationAmbiguous: 2, offramp.PayoutInitiationAccepted: 3,
		offramp.PayoutInitiationNoPayout: 3, offramp.PayoutInitiationTerminalFailed: 3, offramp.PayoutInitiationTerminalCancelled: 3,
	}
	return current.State == updated.State || rank[updated.State] >= rank[current.State]
}
func safeMetadataMatch(left, right map[string]string) bool {
	if len(right) == 0 {
		return false
	}
	for key, value := range right {
		if left[key] != value {
			return false
		}
	}
	return true
}

func validatePersistedPayoutUpdate(current, updated offramp.PayoutResult) error {
	if current.Provider != updated.Provider || current.QuoteID != updated.QuoteID || current.Reference != updated.Reference ||
		!current.FiatAmount.Equal(updated.FiatAmount) || !current.CryptoAmount.Equal(updated.CryptoAmount) || !current.Fee.Equal(updated.Fee) ||
		current.DailyReservationKey != updated.DailyReservationKey || current.DailyReservationOperationID != updated.DailyReservationOperationID ||
		updated.StatusUpdatedAt.Before(current.StatusUpdatedAt) || !fiatPayoutStatusCanAdvance(current.Status, updated.Status) {
		return offramp.ErrProviderRejected
	}
	if current.Status.IsTerminal() && current.Status != updated.Status {
		return offramp.ErrProviderRejected
	}
	return nil
}

func privacySafePersistedPayout(value offramp.PayoutResult) (offramp.PayoutResult, error) {
	for _, id := range []string{value.ID, value.QuoteID, value.Provider, value.Reference, value.FailureCode} {
		if len(id) > 256 || strings.ContainsRune(id, '\x00') || strings.ContainsAny(id, "\r\n") {
			return offramp.PayoutResult{}, offramp.ErrInvalidRequest
		}
	}
	if value.ID == "" || value.QuoteID == "" || value.Provider == "" || value.Reference == "" || len(value.Metadata) == 0 || len(value.Metadata) > 8 {
		return offramp.PayoutResult{}, offramp.ErrInvalidRequest
	}
	if len(value.DailyReservationKey) > 512 || len(value.DailyReservationOperationID) > 256 ||
		strings.ContainsAny(value.DailyReservationKey, "\x00\r\n") || strings.ContainsAny(value.DailyReservationOperationID, "\x00\r\n") ||
		(value.DailyReservationKey == "") != (value.DailyReservationOperationID == "") {
		return offramp.PayoutResult{}, offramp.ErrInvalidRequest
	}
	allowedMetadata := map[string]bool{"idempotency_key": true, "correlation_id": true, "conversion_id": true}
	for key, entry := range value.Metadata {
		lower := strings.ToLower(key)
		if !allowedMetadata[key] || len(entry) == 0 || len(entry) > 256 || strings.ContainsAny(entry, "\x00\r\n") ||
			strings.Contains(lower, "account") || strings.Contains(lower, "bank") || strings.Contains(lower, "beneficiary") || strings.Contains(lower, "secret") || strings.Contains(lower, "credential") {
			return offramp.PayoutResult{}, offramp.ErrInvalidRequest
		}
	}
	value.FailureReason = ""
	value.AuditFields = nil
	return cloneOfframpPayout(value), nil
}

func fiatPayoutStatusCanAdvance(current, updated offramp.Status) bool {
	if current == updated {
		return true
	}
	switch current {
	case offramp.StatusPending:
		return updated == offramp.StatusProcessing || updated.IsTerminal()
	case offramp.StatusProcessing:
		return updated.IsTerminal()
	default:
		return false
	}
}

var _ offramp.PayoutRepository = (*FileFiatRepository)(nil)
var _ offramp.PayoutInitiationRepository = (*FileFiatRepository)(nil)
var _ offramp.DailyLimitRepository = (*FileFiatRepository)(nil)
var _ offramp.DurableWebhookReplayRepository = (*FileFiatRepository)(nil)
var _ offramp.DurableWebhookBindingRepository = (*FileFiatRepository)(nil)
var _ = fmt.Sprintf
