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
)

const (
	reconciliationStoreSchemaVersion uint32 = 1
	reconciliationStoreMaxEvents            = 100_000
)

var (
	ErrReconciliationJobNotFound      = errors.New("reconciliation job not found")
	ErrReconciliationConflict         = errors.New("reconciliation store conflict")
	ErrReconciliationUnavailable      = errors.New("reconciliation store unavailable")
	ErrReconciliationCapacityExceeded = errors.New("reconciliation store event capacity exceeded")
)

type ReconciliationEventType string

const (
	ReconciliationEventJobCreated     ReconciliationEventType = "job_created"
	ReconciliationEventAttemptStarted ReconciliationEventType = "attempt_started"
	ReconciliationEventAttemptFailed  ReconciliationEventType = "attempt_failed"
	ReconciliationEventResultRecorded ReconciliationEventType = "result_recorded"
	ReconciliationEventIntentRecorded ReconciliationEventType = "intent_recorded"
	ReconciliationEventCursorAdvanced ReconciliationEventType = "cursor_advanced"
)

type ReconciliationJob struct {
	ID           string    `json:"id"`
	AllocationID string    `json:"allocation_id"`
	ResourceUUID string    `json:"resource_uuid"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	CreatedAt    time.Time `json:"created_at"`
}

type ReconciliationAttempt struct {
	JobID          string    `json:"job_id"`
	Number         uint32    `json:"number"`
	StartedAt      time.Time `json:"started_at"`
	FinishedAt     time.Time `json:"finished_at,omitempty"`
	Outcome        string    `json:"outcome,omitempty"`
	Classification string    `json:"classification,omitempty"`
}

type ReconciliationEvidenceDigests struct {
	Algorithm         string `json:"algorithm"`
	SchemaVersion     string `json:"schema_version"`
	ProviderDigest    string `json:"provider_digest"`
	IndependentDigest string `json:"independent_digest"`
}

type DurableReconciliationResult struct {
	JobID         string                        `json:"job_id"`
	AttemptNumber uint32                        `json:"attempt_number"`
	Evidence      ReconciliationEvidenceDigests `json:"evidence"`
	Result        ReconciliationResult          `json:"result"`
	ResultDigest  string                        `json:"result_digest"`
	CompletedAt   time.Time                     `json:"completed_at"`
}

type ReconciliationActionIntent struct {
	ID               string    `json:"id"`
	JobID            string    `json:"job_id"`
	ResultDigest     string    `json:"result_digest"`
	Kind             string    `json:"kind"`
	AllocationID     string    `json:"allocation_id"`
	Severity         string    `json:"severity"`
	Status           string    `json:"status"`
	DeliveryAttempts uint32    `json:"delivery_attempts"`
	CreatedAt        time.Time `json:"created_at"`
}

type ReconciliationCursor struct {
	StreamID                 string `json:"stream_id"`
	LastCompletedJobSequence uint64 `json:"last_completed_job_sequence"`
	JobID                    string `json:"job_id"`
	ResultDigest             string `json:"result_digest"`
}

type ReconciliationEvent struct {
	Sequence       uint64                       `json:"sequence"`
	Type           ReconciliationEventType      `json:"type"`
	RecordedAt     time.Time                    `json:"recorded_at"`
	PreviousDigest string                       `json:"previous_digest,omitempty"`
	Digest         string                       `json:"digest"`
	Job            *ReconciliationJob           `json:"job,omitempty"`
	Attempt        *ReconciliationAttempt       `json:"attempt,omitempty"`
	Result         *DurableReconciliationResult `json:"result,omitempty"`
	Intent         *ReconciliationActionIntent  `json:"intent,omitempty"`
	Cursor         *ReconciliationCursor        `json:"cursor,omitempty"`
}

type reconciliationStoreState struct {
	SchemaVersion uint32                `json:"schema_version"`
	TailSequence  uint64                `json:"tail_sequence"`
	TailDigest    string                `json:"tail_digest,omitempty"`
	Events        []ReconciliationEvent `json:"events"`
}

type ReconciliationProjection struct {
	Jobs     map[string]ReconciliationJob
	Attempts map[string][]ReconciliationAttempt
	Results  map[string]DurableReconciliationResult
	Intents  map[string]ReconciliationActionIntent
	Cursors  map[string]ReconciliationCursor
}

type ReconciliationJobStore interface {
	Open(context.Context) error
	Close() error
	PutJobIfAbsent(context.Context, ReconciliationJob) (ReconciliationJob, bool, error)
	BeginAttempt(context.Context, string) (ReconciliationAttempt, error)
	FailAttempt(context.Context, string, uint32, string) error
	CompleteAttempt(context.Context, DurableReconciliationResult, []ReconciliationActionIntent, ReconciliationCursor) error
	LoadProjection(context.Context) (*ReconciliationProjection, error)
	PendingJobs(context.Context) ([]ReconciliationJob, error)
}

type FileReconciliationJobStore struct {
	path      string
	mu        sync.RWMutex
	state     reconciliationStoreState
	open      bool
	now       func() time.Time
	maxEvents int
}

func NewFileReconciliationJobStore(path string) (*FileReconciliationJobStore, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("reconciliation store path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err := validateStatePath(absolute); err != nil {
		return nil, err
	}
	return &FileReconciliationJobStore{
		path: filepath.Clean(absolute), now: func() time.Time { return time.Now().UTC() },
		maxEvents: reconciliationStoreMaxEvents,
	}, nil
}

func (s *FileReconciliationJobStore) Open(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.open {
		return nil
	}
	lock, err := claimTxSubmissionQueuePathWithRetry(ctx, s.path)
	if err != nil {
		return err
	}
	defer lock.release()
	state, err := s.loadStateLocked()
	if err != nil {
		return err
	}
	s.state, s.open = state, true
	return nil
}

func (s *FileReconciliationJobStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.open = false
	return nil
}

func (s *FileReconciliationJobStore) PutJobIfAbsent(ctx context.Context, job ReconciliationJob) (ReconciliationJob, bool, error) {
	if err := validateReconciliationJob(job); err != nil {
		return ReconciliationJob{}, false, err
	}
	var existing ReconciliationJob
	found := false
	err := s.update(ctx, func(state *reconciliationStoreState) error {
		projection, err := projectReconciliationState(*state)
		if err != nil {
			return err
		}
		if candidate, ok := projection.Jobs[job.ID]; ok {
			existing, found = candidate, true
			if candidate != job {
				return ErrReconciliationConflict
			}
			return nil
		}
		return appendReconciliationEvent(state, ReconciliationEvent{Type: ReconciliationEventJobCreated, RecordedAt: s.now(), Job: &job})
	})
	if err != nil {
		return ReconciliationJob{}, false, err
	}
	if found {
		return existing, true, nil
	}
	return job, false, nil
}

func (s *FileReconciliationJobStore) BeginAttempt(ctx context.Context, jobID string) (ReconciliationAttempt, error) {
	var attempt ReconciliationAttempt
	err := s.update(ctx, func(state *reconciliationStoreState) error {
		projection, err := projectReconciliationState(*state)
		if err != nil {
			return err
		}
		if _, ok := projection.Jobs[jobID]; !ok {
			return ErrReconciliationJobNotFound
		}
		attempt = ReconciliationAttempt{JobID: jobID, Number: uint32(len(projection.Attempts[jobID]) + 1), StartedAt: s.now()} //nolint:gosec // attempts are bounded by durable storage capacity
		return appendReconciliationEvent(state, ReconciliationEvent{Type: ReconciliationEventAttemptStarted, RecordedAt: s.now(), Attempt: &attempt})
	})
	return attempt, err
}

func (s *FileReconciliationJobStore) FailAttempt(ctx context.Context, jobID string, attemptNumber uint32, classification string) error {
	return s.update(ctx, func(state *reconciliationStoreState) error {
		projection, err := projectReconciliationState(*state)
		if err != nil {
			return err
		}
		attempt, err := findReconciliationAttempt(projection, jobID, attemptNumber)
		if err != nil {
			return err
		}
		if !attempt.FinishedAt.IsZero() {
			if attempt.Outcome == "failed" && attempt.Classification == classification {
				return nil
			}
			return ErrReconciliationConflict
		}
		attempt.FinishedAt, attempt.Outcome, attempt.Classification = s.now(), "failed", classification
		return appendReconciliationEvent(state, ReconciliationEvent{Type: ReconciliationEventAttemptFailed, RecordedAt: s.now(), Attempt: &attempt})
	})
}

func (s *FileReconciliationJobStore) CompleteAttempt(ctx context.Context, result DurableReconciliationResult, intents []ReconciliationActionIntent, cursor ReconciliationCursor) error {
	if err := validateDurableReconciliationResult(result); err != nil {
		return err
	}
	return s.update(ctx, func(state *reconciliationStoreState) error {
		projection, err := projectReconciliationState(*state)
		if err != nil {
			return err
		}
		attempt, err := findReconciliationAttempt(projection, result.JobID, result.AttemptNumber)
		if err != nil {
			return err
		}
		if !attempt.FinishedAt.IsZero() && attempt.Outcome != "completed" {
			return ErrReconciliationConflict
		}
		if existing, ok := projection.Results[result.JobID]; ok {
			if existing.ResultDigest != result.ResultDigest {
				return ErrReconciliationConflict
			}
			return validateReconciliationCompletionRetry(projection, result, intents, cursor)
		}
		if cursor.JobID != result.JobID || cursor.ResultDigest != result.ResultDigest || cursor.StreamID == "" {
			return errors.New("cursor must reference the completed result")
		}
		candidate := cloneReconciliationState(*state)
		if err := appendReconciliationEvent(&candidate, ReconciliationEvent{Type: ReconciliationEventResultRecorded, RecordedAt: s.now(), Result: &result}); err != nil {
			return err
		}
		cursor.LastCompletedJobSequence = candidate.TailSequence
		for i := range intents {
			intent := intents[i]
			if err := validateReconciliationIntent(intent, result); err != nil {
				return err
			}
			if _, exists := projection.Intents[intent.ID]; exists {
				continue
			}
			if err := appendReconciliationEvent(&candidate, ReconciliationEvent{Type: ReconciliationEventIntentRecorded, RecordedAt: s.now(), Intent: &intent}); err != nil {
				return err
			}
		}
		if err := appendReconciliationEvent(&candidate, ReconciliationEvent{Type: ReconciliationEventCursorAdvanced, RecordedAt: s.now(), Cursor: &cursor}); err != nil {
			return err
		}
		*state = candidate
		return nil
	})
}

func (s *FileReconciliationJobStore) LoadProjection(ctx context.Context) (*ReconciliationProjection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.open {
		return nil, ErrReconciliationUnavailable
	}
	lock, err := claimTxSubmissionQueuePathWithRetry(ctx, s.path)
	if err != nil {
		return nil, err
	}
	defer lock.release()
	state, err := s.loadStateLocked()
	if err != nil {
		return nil, err
	}
	projection, err := projectReconciliationState(state)
	return &projection, err
}

func (s *FileReconciliationJobStore) PendingJobs(ctx context.Context) ([]ReconciliationJob, error) {
	projection, err := s.LoadProjection(ctx)
	if err != nil {
		return nil, err
	}
	jobs := sortedReconciliationJobs(*projection)
	pending := jobs[:0]
	for _, job := range jobs {
		if _, completed := projection.Results[job.ID]; !completed {
			pending = append(pending, job)
		}
	}
	return pending, nil
}

func (s *FileReconciliationJobStore) update(ctx context.Context, mutate func(*reconciliationStoreState) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.open {
		return ErrReconciliationUnavailable
	}
	lock, err := claimTxSubmissionQueuePathWithRetry(ctx, s.path)
	if err != nil {
		return err
	}
	defer lock.release()
	state, err := s.loadStateLocked()
	if err != nil {
		return err
	}
	before := cloneReconciliationState(state)
	if err := mutate(&state); err != nil {
		return err
	}
	if state.TailSequence == before.TailSequence {
		s.state = state
		return nil
	}
	if len(state.Events) > s.maxEvents {
		return ErrReconciliationCapacityExceeded
	}
	if err := validateReconciliationState(state); err != nil {
		return err
	}
	s.state = state
	if err := s.saveLocked(); err != nil {
		s.state = before
		return err
	}
	return nil
}

func validateReconciliationCompletionRetry(
	projection ReconciliationProjection,
	result DurableReconciliationResult,
	intents []ReconciliationActionIntent,
	cursor ReconciliationCursor,
) error {
	storedCursor, ok := projection.Cursors[cursor.StreamID]
	if !ok || cursor.StreamID == "" || cursor.JobID != result.JobID || cursor.ResultDigest != result.ResultDigest ||
		storedCursor.JobID != cursor.JobID || storedCursor.ResultDigest != cursor.ResultDigest ||
		(cursor.LastCompletedJobSequence != 0 && cursor.LastCompletedJobSequence != storedCursor.LastCompletedJobSequence) {
		return ErrReconciliationConflict
	}
	if len(intents) == 0 {
		for _, stored := range projection.Intents {
			if stored.JobID == result.JobID {
				return ErrReconciliationConflict
			}
		}
		return nil
	}
	seen := make(map[string]struct{}, len(intents))
	for _, intent := range intents {
		stored, exists := projection.Intents[intent.ID]
		if !exists || stored != intent || intent.JobID != result.JobID {
			return ErrReconciliationConflict
		}
		if _, duplicate := seen[intent.ID]; duplicate {
			return ErrReconciliationConflict
		}
		seen[intent.ID] = struct{}{}
	}
	for _, stored := range projection.Intents {
		if stored.JobID == result.JobID {
			if _, exists := seen[stored.ID]; !exists {
				return ErrReconciliationConflict
			}
		}
	}
	return nil
}

func (s *FileReconciliationJobStore) loadStateLocked() (reconciliationStoreState, error) {
	state := reconciliationStoreState{SchemaVersion: reconciliationStoreSchemaVersion, Events: []ReconciliationEvent{}}
	raw, err := os.ReadFile(s.path) // #nosec G304 -- constructor validates path.
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return state, fmt.Errorf("decode reconciliation store: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return state, errors.New("decode reconciliation store: multiple JSON values")
	}
	if err := validateReconciliationState(state); err != nil {
		return state, err
	}
	return state, nil
}

func (s *FileReconciliationJobStore) saveLocked() error {
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".reconciliation-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
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
	if dir, err := os.Open(filepath.Dir(s.path)); err == nil { // #nosec G304 -- constructor validates path.
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

func appendReconciliationEvent(state *reconciliationStoreState, event ReconciliationEvent) error {
	event.Sequence = state.TailSequence + 1
	event.PreviousDigest = state.TailDigest
	event.Digest = ""
	digest, err := reconciliationEventDigest(event)
	if err != nil {
		return err
	}
	event.Digest = digest
	state.Events = append(state.Events, event)
	state.TailSequence, state.TailDigest = event.Sequence, event.Digest
	return nil
}

func reconciliationEventDigest(event ReconciliationEvent) (string, error) {
	event.Digest = ""
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(append([]byte("virtengine.reconciliation-event/v1:"), data...))
	return hex.EncodeToString(digest[:]), nil
}

func validateReconciliationState(state reconciliationStoreState) error {
	if state.SchemaVersion != reconciliationStoreSchemaVersion {
		return fmt.Errorf("unsupported reconciliation store schema %d", state.SchemaVersion)
	}
	previous := ""
	for index, event := range state.Events {
		if event.Sequence != uint64(index+1) || event.PreviousDigest != previous { //nolint:gosec // bounded by slice length
			return errors.New("invalid reconciliation event chain")
		}
		if err := validateReconciliationEvent(event); err != nil {
			return err
		}
		digest, err := reconciliationEventDigest(event)
		if err != nil || digest != event.Digest {
			return errors.New("invalid reconciliation event digest")
		}
		previous = event.Digest
	}
	if state.TailSequence != uint64(len(state.Events)) || state.TailDigest != previous { //nolint:gosec // bounded by slice length
		return errors.New("invalid reconciliation store tail")
	}
	_, err := projectReconciliationState(state)
	return err
}

func validateReconciliationEvent(event ReconciliationEvent) error {
	payloads := 0
	for _, present := range []bool{event.Job != nil, event.Attempt != nil, event.Result != nil, event.Intent != nil, event.Cursor != nil} {
		if present {
			payloads++
		}
	}
	if payloads != 1 || event.RecordedAt.IsZero() {
		return errors.New("invalid reconciliation event payload")
	}
	expected := map[ReconciliationEventType]bool{
		ReconciliationEventJobCreated:     event.Job != nil,
		ReconciliationEventAttemptStarted: event.Attempt != nil,
		ReconciliationEventAttemptFailed:  event.Attempt != nil,
		ReconciliationEventResultRecorded: event.Result != nil,
		ReconciliationEventIntentRecorded: event.Intent != nil,
		ReconciliationEventCursorAdvanced: event.Cursor != nil,
	}
	if !expected[event.Type] {
		return fmt.Errorf("invalid payload for event type %q", event.Type)
	}
	return nil
}

func projectReconciliationState(state reconciliationStoreState) (ReconciliationProjection, error) {
	projection := ReconciliationProjection{
		Jobs: make(map[string]ReconciliationJob), Attempts: make(map[string][]ReconciliationAttempt),
		Results: make(map[string]DurableReconciliationResult), Intents: make(map[string]ReconciliationActionIntent),
		Cursors: make(map[string]ReconciliationCursor),
	}
	resultSequences := make(map[string]uint64)
	for _, event := range state.Events {
		switch event.Type {
		case ReconciliationEventJobCreated:
			if err := validateReconciliationJob(*event.Job); err != nil {
				return projection, err
			}
			if _, exists := projection.Jobs[event.Job.ID]; exists {
				return projection, ErrReconciliationConflict
			}
			projection.Jobs[event.Job.ID] = *event.Job
		case ReconciliationEventAttemptStarted:
			if _, exists := projection.Jobs[event.Attempt.JobID]; !exists || event.Attempt.Number != uint32(len(projection.Attempts[event.Attempt.JobID])+1) { //nolint:gosec // bounded by event count
				return projection, errors.New("invalid reconciliation attempt sequence")
			}
			projection.Attempts[event.Attempt.JobID] = append(projection.Attempts[event.Attempt.JobID], *event.Attempt)
		case ReconciliationEventAttemptFailed:
			attempt, err := findReconciliationAttempt(projection, event.Attempt.JobID, event.Attempt.Number)
			if err != nil || !attempt.FinishedAt.IsZero() {
				return projection, errors.New("invalid failed reconciliation attempt")
			}
			projection.Attempts[event.Attempt.JobID][event.Attempt.Number-1] = *event.Attempt
		case ReconciliationEventResultRecorded:
			if err := validateReplayedDurableReconciliationResult(*event.Result); err != nil {
				return projection, err
			}
			if _, exists := projection.Results[event.Result.JobID]; exists {
				return projection, ErrReconciliationConflict
			}
			if _, exists := projection.Jobs[event.Result.JobID]; !exists {
				return projection, errors.New("reconciliation result references missing job")
			}
			if event.Result.Result.AllocationID != projection.Jobs[event.Result.JobID].AllocationID {
				return projection, errors.New("reconciliation result allocation mismatch")
			}
			attempt, err := findReconciliationAttempt(projection, event.Result.JobID, event.Result.AttemptNumber)
			if err != nil || !attempt.FinishedAt.IsZero() {
				return projection, errors.New("reconciliation result references invalid attempt")
			}
			attempt.FinishedAt, attempt.Outcome = event.Result.CompletedAt, "completed"
			projection.Attempts[event.Result.JobID][event.Result.AttemptNumber-1] = attempt
			projection.Results[event.Result.JobID] = *event.Result
			resultSequences[event.Result.JobID] = event.Sequence
		case ReconciliationEventIntentRecorded:
			if _, exists := projection.Intents[event.Intent.ID]; exists {
				return projection, ErrReconciliationConflict
			}
			result, exists := projection.Results[event.Intent.JobID]
			if !exists || validateReconciliationIntent(*event.Intent, result) != nil {
				return projection, errors.New("invalid reconciliation action intent")
			}
			projection.Intents[event.Intent.ID] = *event.Intent
		case ReconciliationEventCursorAdvanced:
			if event.Cursor.StreamID == "" || event.Cursor.JobID == "" || event.Cursor.ResultDigest == "" {
				return projection, errors.New("invalid reconciliation cursor")
			}
			result, exists := projection.Results[event.Cursor.JobID]
			if !exists || result.ResultDigest != event.Cursor.ResultDigest ||
				event.Cursor.LastCompletedJobSequence != resultSequences[event.Cursor.JobID] {
				return projection, errors.New("cursor references missing reconciliation result")
			}
			if previous, exists := projection.Cursors[event.Cursor.StreamID]; exists &&
				event.Cursor.LastCompletedJobSequence <= previous.LastCompletedJobSequence {
				return projection, errors.New("reconciliation cursor cannot regress")
			}
			projection.Cursors[event.Cursor.StreamID] = *event.Cursor
		}
	}
	return projection, nil
}

func findReconciliationAttempt(projection ReconciliationProjection, jobID string, number uint32) (ReconciliationAttempt, error) {
	attempts := projection.Attempts[jobID]
	if number == 0 || int(number) > len(attempts) {
		return ReconciliationAttempt{}, errors.New("reconciliation attempt not found")
	}
	return attempts[number-1], nil
}

func validateReconciliationJob(job ReconciliationJob) error {
	if job.ID == "" || job.AllocationID == "" || job.ResourceUUID == "" || job.CreatedAt.IsZero() || job.PeriodStart.IsZero() || !job.PeriodEnd.After(job.PeriodStart) {
		return errors.New("invalid reconciliation job")
	}
	return nil
}

func validateDurableReconciliationResult(result DurableReconciliationResult) error {
	if result.JobID == "" || result.AttemptNumber == 0 || result.ResultDigest == "" || result.CompletedAt.IsZero() || result.Evidence.Algorithm != "sha256" || result.Evidence.SchemaVersion == "" {
		return errors.New("invalid durable reconciliation result")
	}
	if !validReconciliationMetricState(result.Result.State) {
		return errors.New("unsupported durable reconciliation state")
	}
	if err := validateReconciliationResultSemantics(result.Result); err != nil {
		return err
	}
	for _, digest := range []string{result.ResultDigest, result.Evidence.ProviderDigest, result.Evidence.IndependentDigest} {
		if decoded, err := hex.DecodeString(digest); err != nil || len(decoded) != sha256.Size {
			return errors.New("invalid reconciliation digest")
		}
	}
	if !validReconciliationResultDigest(result.Result, result.ResultDigest) {
		return errors.New("reconciliation result digest mismatch")
	}
	expectedProvider, err := canonicalReconciliationDigest("provider-evidence", result.Result.ProviderMetrics)
	if err != nil || expectedProvider != result.Evidence.ProviderDigest {
		return errors.New("provider evidence digest mismatch")
	}
	expectedIndependent, err := canonicalReconciliationDigest("independent-evidence", result.Result.WaldurMetrics)
	if err != nil || expectedIndependent != result.Evidence.IndependentDigest {
		return errors.New("independent evidence digest mismatch")
	}
	return nil
}

func validateReplayedDurableReconciliationResult(result DurableReconciliationResult) error {
	if result.JobID == "" || result.AttemptNumber == 0 || result.CompletedAt.IsZero() ||
		result.Evidence.Algorithm != "sha256" || result.Evidence.SchemaVersion == "" {
		return errors.New("invalid replayed reconciliation result")
	}
	for _, digest := range []string{result.ResultDigest, result.Evidence.ProviderDigest, result.Evidence.IndependentDigest} {
		if decoded, err := hex.DecodeString(digest); err != nil || len(decoded) != sha256.Size {
			return errors.New("invalid reconciliation digest")
		}
	}
	if !validReconciliationResultDigest(result.Result, result.ResultDigest) {
		return errors.New("reconciliation result digest mismatch")
	}
	expectedProvider, err := canonicalReconciliationDigest("provider-evidence", result.Result.ProviderMetrics)
	if err != nil || expectedProvider != result.Evidence.ProviderDigest {
		return errors.New("provider evidence digest mismatch")
	}
	expectedIndependent, err := canonicalReconciliationDigest("independent-evidence", result.Result.WaldurMetrics)
	if err != nil || expectedIndependent != result.Evidence.IndependentDigest {
		return errors.New("independent evidence digest mismatch")
	}
	return validateReconciliationResultSemantics(result.Result)
}

func canonicalReconciliationResultDigest(result ReconciliationResult) (string, error) {
	canonical := struct {
		AllocationID       string                   `json:"allocation_id"`
		ReconciliationTime string                   `json:"reconciliation_time"`
		ProviderMetrics    ResourceMetrics          `json:"provider_metrics"`
		WaldurMetrics      *ResourceMetrics         `json:"waldur_metrics,omitempty"`
		Discrepancies      []MetricDiscrepancy      `json:"discrepancies,omitempty"`
		State              ReconciliationState      `json:"state"`
		ReasonCode         ReconciliationReasonCode `json:"reason_code"`
		Score              int                      `json:"score"`
	}{
		AllocationID: result.AllocationID, ReconciliationTime: result.ReconciliationTime.UTC().Format(time.RFC3339Nano),
		ProviderMetrics: result.ProviderMetrics, WaldurMetrics: result.WaldurMetrics,
		Discrepancies: result.Discrepancies, State: result.State, ReasonCode: result.ReasonCode, Score: result.Score,
	}
	return canonicalReconciliationDigest("result", canonical)
}

func validReconciliationResultDigest(result ReconciliationResult, digest string) bool {
	canonical, err := canonicalReconciliationResultDigest(result)
	if err == nil && canonical == digest {
		return true
	}
	if result.ReconciliationTime.Nanosecond() == 0 {
		for _, value := range []interface{}{result, &result} {
			expected, err := canonicalReconciliationDigest("result", value)
			if err == nil && expected == digest {
				return true
			}
		}
	}
	return false
}

func validateReconciliationIntent(intent ReconciliationActionIntent, result DurableReconciliationResult) error {
	if intent.ID == "" || intent.JobID != result.JobID || intent.ResultDigest != result.ResultDigest ||
		intent.AllocationID == "" || intent.AllocationID != result.Result.AllocationID || intent.CreatedAt.IsZero() || intent.Status != "pending" {
		return errors.New("invalid reconciliation action intent")
	}
	if intent.Kind != "alert_discrepancy" && intent.Kind != "auto_correct" {
		return errors.New("unsupported reconciliation action intent")
	}
	if intent.Severity != "critical" && intent.Severity != "warning" && intent.Severity != "high" && intent.Severity != "info" {
		return errors.New("unsupported reconciliation action intent severity")
	}
	return nil
}

func newReconciliationJob(allocationID, resourceUUID string, periodStart, periodEnd time.Time) ReconciliationJob {
	identity := struct {
		SchemaVersion string    `json:"schema_version"`
		AllocationID  string    `json:"allocation_id"`
		ResourceUUID  string    `json:"resource_uuid"`
		PeriodStart   time.Time `json:"period_start"`
		PeriodEnd     time.Time `json:"period_end"`
	}{"virtengine.reconciliation-job/v1", allocationID, resourceUUID, periodStart.UTC(), periodEnd.UTC()}
	digest, _ := canonicalReconciliationDigest("job", identity)
	return ReconciliationJob{
		ID: digest, AllocationID: allocationID, ResourceUUID: resourceUUID,
		PeriodStart: periodStart.UTC(), PeriodEnd: periodEnd.UTC(), CreatedAt: periodEnd.UTC(),
	}
}

func buildDurableReconciliationCompletion(job ReconciliationJob, attempt ReconciliationAttempt, result ReconciliationResult) (DurableReconciliationResult, []ReconciliationActionIntent, ReconciliationCursor, error) {
	result.ReconciliationTime = result.ReconciliationTime.UTC().Truncate(time.Second)
	providerDigest, err := canonicalReconciliationDigest("provider-evidence", result.ProviderMetrics)
	if err != nil {
		return DurableReconciliationResult{}, nil, ReconciliationCursor{}, err
	}
	independentDigest, err := canonicalReconciliationDigest("independent-evidence", result.WaldurMetrics)
	if err != nil {
		return DurableReconciliationResult{}, nil, ReconciliationCursor{}, err
	}
	resultDigest, err := canonicalReconciliationResultDigest(result)
	if err != nil {
		return DurableReconciliationResult{}, nil, ReconciliationCursor{}, err
	}
	durable := DurableReconciliationResult{
		JobID: job.ID, AttemptNumber: attempt.Number,
		Evidence: ReconciliationEvidenceDigests{
			Algorithm: "sha256", SchemaVersion: "virtengine.reconciliation-evidence/v1",
			ProviderDigest: providerDigest, IndependentDigest: independentDigest,
		},
		Result: result, ResultDigest: resultDigest, CompletedAt: job.PeriodEnd,
	}
	intents := []ReconciliationActionIntent{}
	if result.State != ReconciliationStateMatched {
		intentID, err := canonicalReconciliationDigest("intent", struct {
			ResultDigest string `json:"result_digest"`
			Kind         string `json:"kind"`
		}{resultDigest, "alert_discrepancy"})
		if err != nil {
			return DurableReconciliationResult{}, nil, ReconciliationCursor{}, err
		}
		intents = append(intents, ReconciliationActionIntent{
			ID: intentID, JobID: job.ID, ResultDigest: resultDigest, Kind: "alert_discrepancy",
			AllocationID: job.AllocationID, Severity: reconciliationIntentSeverity(result), Status: "pending", CreatedAt: job.PeriodEnd,
		})
	}
	cursor := ReconciliationCursor{
		StreamID: "waldur/default",
		JobID:    job.ID, ResultDigest: resultDigest,
	}
	return durable, intents, cursor, nil
}

func canonicalReconciliationDigest(domain string, value interface{}) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(append([]byte("virtengine.reconciliation/"+domain+"/v1:"), data...))
	return hex.EncodeToString(digest[:]), nil
}

func reconciliationIntentSeverity(result ReconciliationResult) string {
	switch result.State {
	case ReconciliationStateUnavailable, ReconciliationStateStale, ReconciliationStateUnresolved:
		return "high"
	case ReconciliationStateMismatched:
		if result.Score < 30 {
			return "critical"
		}
		return "warning"
	default:
		return "info"
	}
}

func cloneReconciliationState(state reconciliationStoreState) reconciliationStoreState {
	data, _ := json.Marshal(state)
	var clone reconciliationStoreState
	_ = json.Unmarshal(data, &clone)
	return clone
}

var _ ReconciliationJobStore = (*FileReconciliationJobStore)(nil)

func sortedReconciliationJobs(projection ReconciliationProjection) []ReconciliationJob {
	jobs := make([]ReconciliationJob, 0, len(projection.Jobs))
	for _, job := range projection.Jobs {
		jobs = append(jobs, job)
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].ID < jobs[j].ID })
	return jobs
}
