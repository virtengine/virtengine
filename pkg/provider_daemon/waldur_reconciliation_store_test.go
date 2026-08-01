// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFileReconciliationJobStoreRestartAndAtomicCompletion(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "reconciliation.json")
	store := openReconciliationStore(t, path)
	job := testReconciliationJob()

	stored, existed, err := store.PutJobIfAbsent(ctx, job)
	require.NoError(t, err)
	require.False(t, existed)
	require.Equal(t, job, stored)
	_, existed, err = store.PutJobIfAbsent(ctx, job)
	require.NoError(t, err)
	require.True(t, existed)

	firstAttempt, err := store.BeginAttempt(ctx, job.ID)
	require.NoError(t, err)
	require.Equal(t, uint32(1), firstAttempt.Number)
	require.NoError(t, store.Close())

	reopened := openReconciliationStore(t, path)
	secondAttempt, err := reopened.BeginAttempt(ctx, job.ID)
	require.NoError(t, err)
	require.Equal(t, uint32(2), secondAttempt.Number)

	result := testDurableReconciliationResult(job, secondAttempt.Number)
	intent := testReconciliationIntent(result)
	cursor := ReconciliationCursor{StreamID: "waldur/default", JobID: job.ID, ResultDigest: result.ResultDigest}
	require.NoError(t, reopened.CompleteAttempt(ctx, result, []ReconciliationActionIntent{intent}, cursor))
	require.NoError(t, reopened.CompleteAttempt(ctx, result, []ReconciliationActionIntent{intent}, cursor))
	require.NoError(t, reopened.Close())

	finalStore := openReconciliationStore(t, path)
	projection, err := finalStore.LoadProjection(ctx)
	require.NoError(t, err)
	require.Equal(t, job, projection.Jobs[job.ID])
	require.Len(t, projection.Attempts[job.ID], 2)
	require.Equal(t, result.ResultDigest, projection.Results[job.ID].ResultDigest)
	require.Equal(t, intent, projection.Intents[intent.ID])
	storedCursor := projection.Cursors[cursor.StreamID]
	require.Equal(t, cursor.JobID, storedCursor.JobID)
	require.Equal(t, cursor.ResultDigest, storedCursor.ResultDigest)
	require.Equal(t, uint64(4), storedCursor.LastCompletedJobSequence)

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var state reconciliationStoreState
	require.NoError(t, json.Unmarshal(raw, &state))
	require.Equal(t, uint64(6), state.TailSequence)
	require.Len(t, state.Events, 6)
}

func TestFileReconciliationJobStoreRejectsConflictingCompletion(t *testing.T) {
	ctx := context.Background()
	store := openReconciliationStore(t, filepath.Join(t.TempDir(), "reconciliation.json"))
	job := testReconciliationJob()
	_, _, err := store.PutJobIfAbsent(ctx, job)
	require.NoError(t, err)
	attempt, err := store.BeginAttempt(ctx, job.ID)
	require.NoError(t, err)
	result := testDurableReconciliationResult(job, attempt.Number)
	cursor := ReconciliationCursor{StreamID: "waldur/default", LastCompletedJobSequence: 1, JobID: job.ID, ResultDigest: result.ResultDigest}
	require.NoError(t, store.CompleteAttempt(ctx, result, nil, cursor))

	conflict := result
	conflict.Result.State = ReconciliationStateMatched
	conflict.Result.ReasonCode = ReconciliationReasonExactMatch
	conflict.ResultDigest, err = canonicalReconciliationDigest("result", conflict.Result)
	require.NoError(t, err)
	conflictingCursor := cursor
	conflictingCursor.ResultDigest = conflict.ResultDigest
	require.ErrorIs(t, store.CompleteAttempt(ctx, conflict, nil, conflictingCursor), ErrReconciliationConflict)
}

func TestFileReconciliationJobStoreRejectsCorruption(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "reconciliation.json")
	store := openReconciliationStore(t, path)
	_, _, err := store.PutJobIfAbsent(ctx, testReconciliationJob())
	require.NoError(t, err)
	require.NoError(t, store.Close())

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var state reconciliationStoreState
	require.NoError(t, json.Unmarshal(raw, &state))
	state.Events[0].Job.AllocationID = "tampered"
	tampered, err := json.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, tampered, 0o600))

	corrupt, err := NewFileReconciliationJobStore(path)
	require.NoError(t, err)
	require.ErrorContains(t, corrupt.Open(ctx), "digest")

	require.NoError(t, os.WriteFile(path, append(raw, []byte(` {}`)...), 0o600))
	trailing, err := NewFileReconciliationJobStore(path)
	require.NoError(t, err)
	require.ErrorContains(t, trailing.Open(ctx), "multiple JSON values")

	var unknown map[string]any
	require.NoError(t, json.Unmarshal(raw, &unknown))
	unknown["unexpected"] = true
	unknownRaw, err := json.Marshal(unknown)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, unknownRaw, 0o600))
	unknownStore, err := NewFileReconciliationJobStore(path)
	require.NoError(t, err)
	require.ErrorContains(t, unknownStore.Open(ctx), "unknown field")
}

func TestFileReconciliationJobStoreSharedReplicasDeduplicateJob(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "reconciliation.json")
	first := openReconciliationStore(t, path)
	second := openReconciliationStore(t, path)
	job := testReconciliationJob()

	type putResult struct {
		existed bool
		err     error
	}
	results := make(chan putResult, 2)
	for _, store := range []*FileReconciliationJobStore{first, second} {
		go func(candidate *FileReconciliationJobStore) {
			_, existed, err := candidate.PutJobIfAbsent(ctx, job)
			results <- putResult{existed: existed, err: err}
		}(store)
	}
	winners := 0
	for range 2 {
		result := <-results
		require.NoError(t, result.err)
		if !result.existed {
			winners++
		}
	}
	require.Equal(t, 1, winners)
	projection, err := first.LoadProjection(ctx)
	require.NoError(t, err)
	require.Len(t, projection.Jobs, 1)
}

func TestFileReconciliationJobStoreFailAttemptPersists(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "reconciliation.json")
	store := openReconciliationStore(t, path)
	job := testReconciliationJob()
	_, _, err := store.PutJobIfAbsent(ctx, job)
	require.NoError(t, err)
	attempt, err := store.BeginAttempt(ctx, job.ID)
	require.NoError(t, err)
	require.NoError(t, store.FailAttempt(ctx, job.ID, attempt.Number, "independent_evidence_unavailable"))
	require.NoError(t, store.FailAttempt(ctx, job.ID, attempt.Number, "independent_evidence_unavailable"))
	require.ErrorIs(t, store.FailAttempt(ctx, job.ID, attempt.Number, "different_classification"), ErrReconciliationConflict)
	result := testDurableReconciliationResult(job, attempt.Number)
	cursor := ReconciliationCursor{StreamID: "waldur/default", JobID: job.ID, ResultDigest: result.ResultDigest}
	require.ErrorIs(t, store.CompleteAttempt(ctx, result, nil, cursor), ErrReconciliationConflict)
	require.NoError(t, store.Close())

	reopened := openReconciliationStore(t, path)
	projection, err := reopened.LoadProjection(ctx)
	require.NoError(t, err)
	require.Equal(t, "failed", projection.Attempts[job.ID][0].Outcome)
	require.Equal(t, "independent_evidence_unavailable", projection.Attempts[job.ID][0].Classification)
	require.Len(t, projection.Attempts[job.ID], 1)
}

func TestFileReconciliationJobStoreCompletionRetryMustMatchMetadata(t *testing.T) {
	ctx := context.Background()
	store := openReconciliationStore(t, filepath.Join(t.TempDir(), "reconciliation.json"))
	job := testReconciliationJob()
	_, _, err := store.PutJobIfAbsent(ctx, job)
	require.NoError(t, err)
	attempt, err := store.BeginAttempt(ctx, job.ID)
	require.NoError(t, err)
	result := testDurableReconciliationResult(job, attempt.Number)
	intent := testReconciliationIntent(result)
	cursor := ReconciliationCursor{StreamID: "waldur/default", JobID: job.ID, ResultDigest: result.ResultDigest}
	require.NoError(t, store.CompleteAttempt(ctx, result, []ReconciliationActionIntent{intent}, cursor))
	require.NoError(t, store.CompleteAttempt(ctx, result, []ReconciliationActionIntent{intent}, cursor))

	require.ErrorIs(t, store.CompleteAttempt(ctx, result, nil, cursor), ErrReconciliationConflict)
	wrongCursor := cursor
	wrongCursor.StreamID = "waldur/other"
	require.ErrorIs(t, store.CompleteAttempt(ctx, result, []ReconciliationActionIntent{intent}, wrongCursor), ErrReconciliationConflict)
	wrongIntent := intent
	wrongIntent.Severity = "critical"
	require.ErrorIs(t, store.CompleteAttempt(ctx, result, []ReconciliationActionIntent{wrongIntent}, cursor), ErrReconciliationConflict)

	projection, err := store.LoadProjection(ctx)
	require.NoError(t, err)
	require.Len(t, projection.Results, 1)
	require.Len(t, projection.Intents, 1)
	require.Len(t, projection.Cursors, 1)
}

func TestFileReconciliationJobStoreCapacityFailureIsAtomic(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "reconciliation.json")
	store := openReconciliationStore(t, path)
	store.maxEvents = 2
	job := testReconciliationJob()
	_, _, err := store.PutJobIfAbsent(ctx, job)
	require.NoError(t, err)
	attempt, err := store.BeginAttempt(ctx, job.ID)
	require.NoError(t, err)
	before, err := os.ReadFile(path)
	require.NoError(t, err)
	result := testDurableReconciliationResult(job, attempt.Number)
	intent := testReconciliationIntent(result)
	cursor := ReconciliationCursor{StreamID: "waldur/default", JobID: job.ID, ResultDigest: result.ResultDigest}
	require.ErrorIs(t, store.CompleteAttempt(ctx, result, []ReconciliationActionIntent{intent}, cursor), ErrReconciliationCapacityExceeded)
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after)
	projection, err := store.LoadProjection(ctx)
	require.NoError(t, err)
	require.Empty(t, projection.Results)
	require.Empty(t, projection.Intents)
	require.Empty(t, projection.Cursors)
	require.True(t, projection.Attempts[job.ID][0].FinishedAt.IsZero())
	require.NoError(t, store.Close())

	reopened := openReconciliationStore(t, path)
	projection, err = reopened.LoadProjection(ctx)
	require.NoError(t, err)
	require.Empty(t, projection.Results)
	require.True(t, projection.Attempts[job.ID][0].FinishedAt.IsZero())
}

func TestFileReconciliationJobStoreExactCapacityAllowsIdempotentRetry(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "reconciliation.json")
	store := openReconciliationStore(t, path)
	store.maxEvents = 4
	job := testReconciliationJob()
	_, _, err := store.PutJobIfAbsent(ctx, job)
	require.NoError(t, err)
	attempt, err := store.BeginAttempt(ctx, job.ID)
	require.NoError(t, err)
	result := testDurableReconciliationResult(job, attempt.Number)
	cursor := ReconciliationCursor{StreamID: "waldur/default", JobID: job.ID, ResultDigest: result.ResultDigest}
	require.NoError(t, store.CompleteAttempt(ctx, result, nil, cursor))
	before, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, store.CompleteAttempt(ctx, result, nil, cursor))
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestFileReconciliationJobStoreSharedReplicasRespectCapacity(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "reconciliation.json")
	first := openReconciliationStore(t, path)
	second := openReconciliationStore(t, path)
	first.maxEvents, second.maxEvents = 1, 1
	firstJob := testReconciliationJob()
	secondJob := firstJob
	secondJob.ID, secondJob.AllocationID, secondJob.ResourceUUID = "job-2", "allocation-2", "resource-2"
	type result struct{ err error }
	results := make(chan result, 2)
	for _, candidate := range []struct {
		store *FileReconciliationJobStore
		job   ReconciliationJob
	}{{first, firstJob}, {second, secondJob}} {
		go func(store *FileReconciliationJobStore, job ReconciliationJob) {
			_, _, err := store.PutJobIfAbsent(ctx, job)
			results <- result{err: err}
		}(candidate.store, candidate.job)
	}
	succeeded, capacityFailures := 0, 0
	for range 2 {
		outcome := <-results
		switch {
		case outcome.err == nil:
			succeeded++
		case errors.Is(outcome.err, ErrReconciliationCapacityExceeded):
			capacityFailures++
		default:
			t.Fatalf("unexpected replica result: %v", outcome.err)
		}
	}
	require.Equal(t, 1, succeeded)
	require.Equal(t, 1, capacityFailures)
	projection, err := first.LoadProjection(ctx)
	require.NoError(t, err)
	require.Len(t, projection.Jobs, 1)
}

func openReconciliationStore(t *testing.T, path string) *FileReconciliationJobStore {
	t.Helper()
	store, err := NewFileReconciliationJobStore(path)
	require.NoError(t, err)
	require.NoError(t, store.Open(context.Background()))
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func testReconciliationJob() ReconciliationJob {
	now := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	return ReconciliationJob{
		ID: "job-1", AllocationID: "allocation-1", ResourceUUID: "resource-1",
		PeriodStart: now.Add(-time.Hour), PeriodEnd: now, CreatedAt: now,
	}
}

func testDurableReconciliationResult(job ReconciliationJob, attempt uint32) DurableReconciliationResult {
	result := ReconciliationResult{
		AllocationID: job.AllocationID, ReconciliationTime: job.PeriodEnd,
		State: ReconciliationStateMismatched, ReasonCode: ReconciliationReasonMetricThresholdExceeded,
	}
	providerDigest, _ := canonicalReconciliationDigest("provider-evidence", result.ProviderMetrics)
	independentDigest, _ := canonicalReconciliationDigest("independent-evidence", result.WaldurMetrics)
	resultDigest, _ := canonicalReconciliationDigest("result", result)
	return DurableReconciliationResult{
		JobID: job.ID, AttemptNumber: attempt,
		Evidence: ReconciliationEvidenceDigests{
			Algorithm: "sha256", SchemaVersion: "virtengine.reconciliation-evidence/v1",
			ProviderDigest: providerDigest, IndependentDigest: independentDigest,
		},
		Result: result, ResultDigest: resultDigest, CompletedAt: job.PeriodEnd,
	}
}

func testReconciliationIntent(result DurableReconciliationResult) ReconciliationActionIntent {
	digest := sha256.Sum256([]byte(result.ResultDigest + ":alert_discrepancy"))
	return ReconciliationActionIntent{
		ID: hex.EncodeToString(digest[:]), JobID: result.JobID, ResultDigest: result.ResultDigest,
		Kind: "alert_discrepancy", AllocationID: result.Result.AllocationID, Severity: "high", Status: "pending", CreatedAt: result.CompletedAt,
	}
}
