package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type queueItemKind string

const (
	queueItemKindUsage      queueItemKind = "usage"
	queueItemKindSettlement queueItemKind = "settlement"
)

type queueItemStatus string

const (
	queueItemStatusPending         queueItemStatus = "pending"
	queueItemStatusBroadcasting    queueItemStatus = "broadcasting"
	queueItemStatusBroadcasted     queueItemStatus = "broadcasted"
	queueItemStatusRetryableFailed queueItemStatus = "retryable_failed"
	queueItemStatusFailed          queueItemStatus = "failed"
)

type txSubmissionQueueItem struct {
	ID              string          `json:"id"`
	IdempotencyKey  string          `json:"idempotency_key"`
	ChainID         string          `json:"chain_id"`
	Kind            queueItemKind   `json:"kind"`
	Payload         json.RawMessage `json:"payload"`
	AttemptCount    int             `json:"attempt_count"`
	NextAttemptAt   time.Time       `json:"next_attempt_at"`
	Status          queueItemStatus `json:"status"`
	LastError       string          `json:"last_error,omitempty"`
	BroadcastTxHash string          `json:"broadcast_tx_hash,omitempty"`
	ClaimedBy       string          `json:"claimed_by,omitempty"`
	ClaimExpiresAt  time.Time       `json:"claim_expires_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	LastAttemptAt   time.Time       `json:"last_attempt_at,omitempty"`
}

type txSubmissionQueueState struct {
	Items map[string]*txSubmissionQueueItem `json:"items"`
}

type txSubmissionQueueStore struct {
	path string
}

func newTxSubmissionQueueStore(path string) (*txSubmissionQueueStore, error) {
	if err := validateStatePath(path); err != nil {
		return nil, fmt.Errorf("invalid queue state path: %w", err)
	}
	return &txSubmissionQueueStore{path: filepath.Clean(path)}, nil
}

func (s *txSubmissionQueueStore) Load() (*txSubmissionQueueState, error) {
	data, err := os.ReadFile(s.path) // #nosec G304 -- path validated in constructor
	if err != nil {
		if os.IsNotExist(err) {
			return &txSubmissionQueueState{Items: make(map[string]*txSubmissionQueueItem)}, nil
		}
		return nil, fmt.Errorf("read queue state: %w", err)
	}
	state := &txSubmissionQueueState{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, fmt.Errorf("decode queue state: %w", err)
	}
	if state.Items == nil {
		state.Items = make(map[string]*txSubmissionQueueItem)
	}
	return state, nil
}

func (s *txSubmissionQueueStore) Save(state *txSubmissionQueueState) error {
	if state == nil {
		return fmt.Errorf("queue state is nil")
	}
	if state.Items == nil {
		state.Items = make(map[string]*txSubmissionQueueItem)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode queue state: %w", err)
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create queue state dir: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil { // #nosec G304 -- path validated in constructor
		return fmt.Errorf("write queue state tmp: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace queue state: %w", err)
	}
	return nil
}

func recoverQueueStateForRestart(state *txSubmissionQueueState, now time.Time) bool {
	if state == nil || len(state.Items) == 0 {
		return false
	}
	changed := false
	for _, item := range state.Items {
		if item == nil || item.Status != queueItemStatusBroadcasting {
			continue
		}
		item.ClaimedBy = ""
		item.ClaimExpiresAt = time.Time{}
		if item.BroadcastTxHash != "" || item.AttemptCount > 0 || !item.LastAttemptAt.IsZero() {
			item.Status = queueItemStatusRetryableFailed
		} else {
			item.Status = queueItemStatusPending
		}
		if item.NextAttemptAt.After(now) {
			item.NextAttemptAt = now
		}
		item.UpdatedAt = now
		changed = true
	}
	return changed
}

func (s *ChainUsageSubmitterImpl) enqueueMessage(kind queueItemKind, msg interface{}) (*txSubmissionQueueItem, bool, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, false, fmt.Errorf("marshal queue payload: %w", err)
	}
	idempotencyKey := s.computeIdempotencyKey(kind, payload)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.queueState == nil {
		s.queueState = &txSubmissionQueueState{Items: make(map[string]*txSubmissionQueueItem)}
	}
	if existing, ok := s.queueState.Items[idempotencyKey]; ok {
		return existing, true, nil
	}

	item := &txSubmissionQueueItem{
		ID:             idempotencyKey[:16],
		IdempotencyKey: idempotencyKey,
		ChainID:        s.cfg.ChainID,
		Kind:           kind,
		Payload:        payload,
		AttemptCount:   0,
		NextAttemptAt:  now,
		Status:         queueItemStatusPending,
		ClaimedBy:      "",
		ClaimExpiresAt: time.Time{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.queueState.Items[idempotencyKey] = item
	if err := s.queueStore.Save(s.queueState); err != nil {
		delete(s.queueState.Items, idempotencyKey)
		return nil, false, err
	}
	log.Printf("[chain-submitter] queue enqueue id=%s key=%s kind=%s", item.ID, item.IdempotencyKey, item.Kind)
	return item, false, nil
}

func (s *ChainUsageSubmitterImpl) computeIdempotencyKey(kind queueItemKind, payload []byte) string {
	h := sha256.New()
	h.Write([]byte(s.cfg.ChainID))
	h.Write([]byte("|"))
	h.Write([]byte(kind))
	h.Write([]byte("|"))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

func (s *ChainUsageSubmitterImpl) retryLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.WorkerPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			keys := s.pickDueQueueKeys(time.Now().UTC(), 32)
			for _, key := range keys {
				if err := s.processQueueItem(ctx, key, false); err != nil {
					log.Printf("[chain-submitter] queue process key=%s err=%v", key, err)
				}
			}
		}
	}
}

func (s *ChainUsageSubmitterImpl) pickDueQueueKeys(now time.Time, limit int) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.queueState == nil || len(s.queueState.Items) == 0 {
		return nil
	}
	keys := make([]string, 0, len(s.queueState.Items))
	for key, item := range s.queueState.Items {
		switch item.Status {
		case queueItemStatusBroadcasted, queueItemStatusFailed:
			continue
		case queueItemStatusBroadcasting:
			if item.ClaimExpiresAt.After(now) && item.ClaimedBy != s.workerID {
				continue
			}
		}
		if item.NextAttemptAt.After(now) {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if limit > 0 && len(keys) > limit {
		keys = keys[:limit]
	}
	return keys
}

func (s *ChainUsageSubmitterImpl) processQueueItem(ctx context.Context, idempotencyKey string, inline bool) error {
	for {
		item, err := s.claimQueueItem(idempotencyKey)
		if err != nil {
			return err
		}
		if item == nil {
			return nil
		}
		reconciled, err := s.reconcileQueueItem(ctx, item)
		if err != nil {
			return err
		}
		if reconciled {
			return nil
		}
		msg, err := decodeQueueMessage(item)
		if err != nil {
			return s.markQueueFailed(idempotencyKey, err)
		}

		txBytes, localTxHash, prepareErr := s.prepareSignedTx(ctx, msg)
		if prepareErr != nil {
			decisionErr := s.resolveQueueAttempt(idempotencyKey, "", prepareErr)
			if decisionErr != nil {
				return decisionErr
			}
			if !inline {
				return nil
			}
			retryNow, sleepAttempt := s.shouldRetryInline(idempotencyKey)
			if !retryNow {
				return prepareErr
			}
			if err := s.sleepBackoff(ctx, sleepAttempt); err != nil {
				return err
			}
			continue
		}
		if err := s.persistQueueBroadcastHash(idempotencyKey, localTxHash); err != nil {
			return err
		}

		broadcastTxHash, broadcastErr := s.broadcastPreparedTx(ctx, txBytes)
		resolvedTxHash := broadcastTxHash
		if resolvedTxHash == "" {
			resolvedTxHash = localTxHash
		}
		decisionErr := s.resolveQueueAttempt(idempotencyKey, resolvedTxHash, broadcastErr)
		if decisionErr != nil {
			return decisionErr
		}
		if broadcastErr == nil {
			return nil
		}
		if broadcastTxHash != "" || isAmbiguousBroadcastError(broadcastErr) {
			return nil
		}
		if !inline {
			return nil
		}
		retryNow, sleepAttempt := s.shouldRetryInline(idempotencyKey)
		if !retryNow {
			return broadcastErr
		}
		if err := s.sleepBackoff(ctx, sleepAttempt); err != nil {
			return err
		}
	}
}

func (s *ChainUsageSubmitterImpl) claimQueueItem(idempotencyKey string) (*txSubmissionQueueItem, error) {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.queueState.Items[idempotencyKey]
	if !ok {
		return nil, nil
	}
	switch item.Status {
	case queueItemStatusBroadcasted:
		return nil, nil
	case queueItemStatusFailed:
		return nil, fmt.Errorf("queue item %s reached terminal failure", item.ID)
	case queueItemStatusBroadcasting:
		if item.ClaimExpiresAt.After(now) {
			return nil, nil
		}
	}
	if item.NextAttemptAt.After(now) {
		return nil, nil
	}
	item.Status = queueItemStatusBroadcasting
	item.ClaimedBy = s.workerID
	item.ClaimExpiresAt = now.Add(s.cfg.ClaimTTL)
	item.LastAttemptAt = now
	item.UpdatedAt = now
	if err := s.queueStore.Save(s.queueState); err != nil {
		return nil, err
	}
	cloned := *item
	return &cloned, nil
}

func (s *ChainUsageSubmitterImpl) persistQueueBroadcastHash(idempotencyKey string, txHash string) error {
	if txHash == "" {
		return nil
	}
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.queueState.Items[idempotencyKey]
	if !ok {
		return nil
	}
	if item.BroadcastTxHash == txHash {
		return nil
	}
	item.BroadcastTxHash = txHash
	item.UpdatedAt = now
	if err := s.queueStore.Save(s.queueState); err != nil {
		return err
	}
	log.Printf("[chain-submitter] queue prepared_broadcast id=%s key=%s tx_hash=%s", item.ID, item.IdempotencyKey, item.BroadcastTxHash)
	return nil
}

func decodeQueueMessage(item *txSubmissionQueueItem) (interface{}, error) {
	switch item.Kind {
	case queueItemKindUsage:
		var msg MsgRecordUsageWrapper
		if err := json.Unmarshal(item.Payload, &msg); err != nil {
			return nil, fmt.Errorf("decode usage queue payload: %w", err)
		}
		return &msg, nil
	case queueItemKindSettlement:
		var msg MsgSettleOrderWrapper
		if err := json.Unmarshal(item.Payload, &msg); err != nil {
			return nil, fmt.Errorf("decode settlement queue payload: %w", err)
		}
		return &msg, nil
	default:
		return nil, fmt.Errorf("unknown queue item kind: %s", item.Kind)
	}
}

func (s *ChainUsageSubmitterImpl) resolveQueueAttempt(idempotencyKey string, txHash string, broadcastErr error) error {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.queueState.Items[idempotencyKey]
	if !ok {
		return nil
	}
	item.AttemptCount++
	if txHash != "" {
		item.BroadcastTxHash = txHash
	}
	if broadcastErr == nil {
		item.Status = queueItemStatusBroadcasted
		item.LastError = ""
		item.ClaimedBy = ""
		item.ClaimExpiresAt = time.Time{}
		item.UpdatedAt = now
		if err := s.queueStore.Save(s.queueState); err != nil {
			return err
		}
		log.Printf("[chain-submitter] queue success id=%s key=%s tx_hash=%s attempts=%d", item.ID, item.IdempotencyKey, item.BroadcastTxHash, item.AttemptCount)
		return nil
	}

	retryable := isRetryableBroadcastError(broadcastErr)
	item.LastError = broadcastErr.Error()
	if !retryable || item.AttemptCount >= s.cfg.MaxAttempts {
		item.Status = queueItemStatusFailed
		item.ClaimedBy = ""
		item.ClaimExpiresAt = time.Time{}
		item.UpdatedAt = now
		if err := s.queueStore.Save(s.queueState); err != nil {
			return err
		}
		log.Printf("[chain-submitter] queue terminal_failure id=%s key=%s attempts=%d err=%s tx_hash=%s", item.ID, item.IdempotencyKey, item.AttemptCount, item.LastError, item.BroadcastTxHash)
		return broadcastErr
	}

	item.Status = queueItemStatusRetryableFailed
	item.ClaimedBy = ""
	item.ClaimExpiresAt = time.Time{}
	item.NextAttemptAt = now.Add(s.nextRetryDelay(item.AttemptCount))
	item.UpdatedAt = now
	if err := s.queueStore.Save(s.queueState); err != nil {
		return err
	}
	log.Printf("[chain-submitter] queue retry_scheduled id=%s key=%s attempt=%d next=%s err=%s tx_hash=%s", item.ID, item.IdempotencyKey, item.AttemptCount, item.NextAttemptAt.Format(time.RFC3339), item.LastError, item.BroadcastTxHash)
	return nil
}

func (s *ChainUsageSubmitterImpl) reconcileQueueItem(ctx context.Context, item *txSubmissionQueueItem) (bool, error) {
	if item == nil || item.BroadcastTxHash == "" {
		return false, nil
	}
	reconciler, ok := s.chainClient.(txReconciler)
	if !ok {
		return false, nil
	}
	reconciliation, err := reconciler.ReconcileTx(ctx, item.BroadcastTxHash)
	if err != nil {
		log.Printf("[chain-submitter] queue reconcile_failed id=%s key=%s tx_hash=%s err=%v", item.ID, item.IdempotencyKey, item.BroadcastTxHash, err)
		return false, nil
	}
	if !reconciliation.Found {
		return false, nil
	}
	if reconciliation.Success {
		return true, s.markQueueBroadcasted(item.IdempotencyKey, item.BroadcastTxHash)
	}
	cause := reconciliation.Err
	if cause == nil {
		cause = fmt.Errorf("transaction %s failed reconciliation", item.BroadcastTxHash)
	}
	return true, s.markQueueFailed(item.IdempotencyKey, cause)
}

func (s *ChainUsageSubmitterImpl) markQueueBroadcasted(idempotencyKey string, txHash string) error {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.queueState.Items[idempotencyKey]
	if !ok {
		return nil
	}
	if txHash != "" {
		item.BroadcastTxHash = txHash
	}
	item.Status = queueItemStatusBroadcasted
	item.LastError = ""
	item.ClaimedBy = ""
	item.ClaimExpiresAt = time.Time{}
	item.UpdatedAt = now
	if err := s.queueStore.Save(s.queueState); err != nil {
		return err
	}
	log.Printf("[chain-submitter] queue reconciled_success id=%s key=%s tx_hash=%s attempts=%d", item.ID, item.IdempotencyKey, item.BroadcastTxHash, item.AttemptCount)
	return nil
}

func (s *ChainUsageSubmitterImpl) shouldRetryInline(idempotencyKey string) (bool, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.queueState.Items[idempotencyKey]
	if !ok {
		return false, 0
	}
	if item.Status != queueItemStatusRetryableFailed {
		return false, 0
	}
	return true, item.AttemptCount - 1
}

func (s *ChainUsageSubmitterImpl) nextRetryDelay(attemptCount int) time.Duration {
	if s.cfg.RetryBackoff <= 0 {
		return 0
	}
	delay := s.cfg.RetryBackoff * time.Duration(1<<maxInt(0, attemptCount-1))
	if delay > 0 {
		jitterWindow := delay / 5 // add up to 20% jitter
		if jitterWindow > 0 {
			jitter := time.Duration(time.Now().UnixNano() % int64(jitterWindow))
			delay += jitter
		}
	}
	if s.cfg.MaxRetryBackoff > 0 && delay > s.cfg.MaxRetryBackoff {
		delay = s.cfg.MaxRetryBackoff
	}
	return delay
}

func (s *ChainUsageSubmitterImpl) markQueueFailed(idempotencyKey string, cause error) error {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.queueState.Items[idempotencyKey]
	if !ok {
		return cause
	}
	item.Status = queueItemStatusFailed
	item.LastError = cause.Error()
	item.ClaimedBy = ""
	item.ClaimExpiresAt = time.Time{}
	item.UpdatedAt = now
	if err := s.queueStore.Save(s.queueState); err != nil {
		return err
	}
	return cause
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
