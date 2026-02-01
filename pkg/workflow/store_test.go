// Package workflow provides persistence tests for workflow storage
package workflow

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestMemoryWorkflowStore(t *testing.T) {
	ctx := context.Background()

	t.Run("save and load state", func(t *testing.T) {
		store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
		defer store.Close()

		state := &WorkflowState{
			ID:             "test-workflow-1",
			Name:           "test-workflow",
			Status:         WorkflowStatusRunning,
			CurrentStep:    "step1",
			Data:           map[string]interface{}{"key": "value"},
			StartedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CompletedSteps: []string{"step0"},
			Metadata:       map[string]string{"env": "test"},
		}

		err := store.SaveState(ctx, state.ID, state)
		if err != nil {
			t.Fatalf("failed to save state: %v", err)
		}

		loaded, err := store.LoadState(ctx, state.ID)
		if err != nil {
			t.Fatalf("failed to load state: %v", err)
		}

		if loaded == nil {
			t.Fatal("loaded state is nil")
		}
		if loaded.ID != state.ID {
			t.Errorf("expected ID %s, got %s", state.ID, loaded.ID)
		}
		if loaded.Status != state.Status {
			t.Errorf("expected status %s, got %s", state.Status, loaded.Status)
		}
		if loaded.Data["key"] != "value" {
			t.Errorf("expected data key=value, got %v", loaded.Data)
		}
		if len(loaded.CompletedSteps) != 1 || loaded.CompletedSteps[0] != "step0" {
			t.Errorf("expected completed steps [step0], got %v", loaded.CompletedSteps)
		}
	})

	t.Run("load non-existent state returns nil", func(t *testing.T) {
		store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
		defer store.Close()

		loaded, err := store.LoadState(ctx, "non-existent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if loaded != nil {
			t.Error("expected nil for non-existent state")
		}
	})

	t.Run("delete state", func(t *testing.T) {
		store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
		defer store.Close()

		state := &WorkflowState{
			ID:        "test-workflow-2",
			Name:      "test-workflow",
			Status:    WorkflowStatusRunning,
			StartedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_ = store.SaveState(ctx, state.ID, state)

		err := store.DeleteState(ctx, state.ID)
		if err != nil {
			t.Fatalf("failed to delete state: %v", err)
		}

		loaded, _ := store.LoadState(ctx, state.ID)
		if loaded != nil {
			t.Error("state should be deleted")
		}
	})

	t.Run("list states with filter", func(t *testing.T) {
		store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
		defer store.Close()

		// Create multiple states
		now := time.Now()
		states := []*WorkflowState{
			{ID: "wf-1", Name: "workflow-a", Status: WorkflowStatusRunning, StartedAt: now.Add(-3 * time.Hour), UpdatedAt: now},
			{ID: "wf-2", Name: "workflow-b", Status: WorkflowStatusCompleted, StartedAt: now.Add(-2 * time.Hour), UpdatedAt: now},
			{ID: "wf-3", Name: "workflow-a", Status: WorkflowStatusRunning, StartedAt: now.Add(-1 * time.Hour), UpdatedAt: now},
			{ID: "wf-4", Name: "workflow-c", Status: WorkflowStatusFailed, StartedAt: now, UpdatedAt: now},
		}

		for _, s := range states {
			_ = store.SaveState(ctx, s.ID, s)
		}

		// Filter by status
		running, err := store.ListStates(ctx, StateFilter{Status: WorkflowStatusRunning})
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(running) != 2 {
			t.Errorf("expected 2 running workflows, got %d", len(running))
		}

		// Filter by name
		workflowA, err := store.ListStates(ctx, StateFilter{Name: "workflow-a"})
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(workflowA) != 2 {
			t.Errorf("expected 2 workflow-a, got %d", len(workflowA))
		}

		// Pagination
		limited, err := store.ListStates(ctx, StateFilter{Limit: 2})
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(limited) != 2 {
			t.Errorf("expected 2 with limit, got %d", len(limited))
		}

		offset, err := store.ListStates(ctx, StateFilter{Offset: 2, Limit: 2})
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(offset) != 2 {
			t.Errorf("expected 2 with offset, got %d", len(offset))
		}
	})

	t.Run("checkpoint operations", func(t *testing.T) {
		store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
		defer store.Close()

		workflowID := "wf-checkpoints"

		// Save checkpoints
		cp1 := &Checkpoint{
			ID:         "wf-checkpoints-step1",
			WorkflowID: workflowID,
			Step:       "step1",
			Status:     CheckpointCompleted,
			Data:       "step1-data",
			CreatedAt:  time.Now().Add(-2 * time.Minute),
			UpdatedAt:  time.Now().Add(-2 * time.Minute),
		}
		cp2 := &Checkpoint{
			ID:         "wf-checkpoints-step2",
			WorkflowID: workflowID,
			Step:       "step2",
			Status:     CheckpointCompleted,
			Data:       "step2-data",
			CreatedAt:  time.Now().Add(-1 * time.Minute),
			UpdatedAt:  time.Now().Add(-1 * time.Minute),
		}
		cp3 := &Checkpoint{
			ID:         "wf-checkpoints-step3",
			WorkflowID: workflowID,
			Step:       "step3",
			Status:     CheckpointPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		_ = store.SaveCheckpoint(ctx, workflowID, cp1)
		_ = store.SaveCheckpoint(ctx, workflowID, cp2)
		_ = store.SaveCheckpoint(ctx, workflowID, cp3)

		// Load specific checkpoint
		loaded, err := store.LoadCheckpoint(ctx, workflowID, "step2")
		if err != nil {
			t.Fatalf("failed to load checkpoint: %v", err)
		}
		if loaded == nil || loaded.Step != "step2" {
			t.Error("expected step2 checkpoint")
		}

		// Load latest checkpoint
		latest, err := store.LoadLatestCheckpoint(ctx, workflowID)
		if err != nil {
			t.Fatalf("failed to load latest: %v", err)
		}
		if latest == nil || latest.Step != "step3" {
			t.Errorf("expected step3 as latest, got %v", latest)
		}

		// List checkpoints
		all, err := store.ListCheckpoints(ctx, workflowID)
		if err != nil {
			t.Fatalf("failed to list: %v", err)
		}
		if len(all) != 3 {
			t.Errorf("expected 3 checkpoints, got %d", len(all))
		}

		// Delete checkpoints
		err = store.DeleteCheckpoints(ctx, workflowID)
		if err != nil {
			t.Fatalf("failed to delete: %v", err)
		}

		after, _ := store.ListCheckpoints(ctx, workflowID)
		if len(after) != 0 {
			t.Error("expected 0 checkpoints after delete")
		}
	})

	t.Run("history operations", func(t *testing.T) {
		store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
		defer store.Close()

		workflowID := "wf-history"

		// Append events
		events := []*HistoryEvent{
			{ID: "e1", WorkflowID: workflowID, EventType: HistoryEventWorkflowStarted, Timestamp: time.Now().Add(-3 * time.Minute)},
			{ID: "e2", WorkflowID: workflowID, EventType: HistoryEventStepStarted, Step: "step1", Timestamp: time.Now().Add(-2 * time.Minute)},
			{ID: "e3", WorkflowID: workflowID, EventType: HistoryEventStepCompleted, Step: "step1", Timestamp: time.Now().Add(-1 * time.Minute)},
		}

		for _, e := range events {
			err := store.AppendHistory(ctx, workflowID, e)
			if err != nil {
				t.Fatalf("failed to append history: %v", err)
			}
		}

		// Get history
		history, err := store.GetHistory(ctx, workflowID)
		if err != nil {
			t.Fatalf("failed to get history: %v", err)
		}
		if len(history) != 3 {
			t.Errorf("expected 3 history events, got %d", len(history))
		}

		// Delete history
		err = store.DeleteHistory(ctx, workflowID)
		if err != nil {
			t.Fatalf("failed to delete history: %v", err)
		}

		after, _ := store.GetHistory(ctx, workflowID)
		if len(after) != 0 {
			t.Error("expected 0 history events after delete")
		}
	})

	t.Run("max history limit", func(t *testing.T) {
		config := DefaultWorkflowStoreConfig()
		config.MaxHistoryPerWorkflow = 5
		store := NewMemoryWorkflowStore(config)
		defer store.Close()

		workflowID := "wf-max-history"

		// Append more events than the limit
		for i := 0; i < 10; i++ {
			event := &HistoryEvent{
				ID:         "e" + string(rune('0'+i)),
				WorkflowID: workflowID,
				EventType:  HistoryEventStepCompleted,
				Timestamp:  time.Now(),
			}
			_ = store.AppendHistory(ctx, workflowID, event)
		}

		history, _ := store.GetHistory(ctx, workflowID)
		if len(history) != 5 {
			t.Errorf("expected 5 history events (max), got %d", len(history))
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
		defer store.Close()

		var wg sync.WaitGroup
		numWorkers := 10
		opsPerWorker := 100

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < opsPerWorker; j++ {
					state := &WorkflowState{
						ID:        "wf-concurrent-" + string(rune('0'+workerID)) + "-" + string(rune('0'+j%10)),
						Name:      "concurrent-test",
						Status:    WorkflowStatusRunning,
						StartedAt: time.Now(),
						UpdatedAt: time.Now(),
					}
					_ = store.SaveState(ctx, state.ID, state)
					_, _ = store.LoadState(ctx, state.ID)
					_, _ = store.ListStates(ctx, StateFilter{})
				}
			}(i)
		}

		wg.Wait()
		// If we got here without panicking, concurrent access is safe
	})

	t.Run("closed store returns error", func(t *testing.T) {
		store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
		store.Close()

		_, err := store.LoadState(ctx, "test")
		if err == nil {
			t.Error("expected error from closed store")
		}
	})
}

func TestWorkflowStoreFactory(t *testing.T) {
	ctx := context.Background()

	t.Run("creates memory store", func(t *testing.T) {
		config := WorkflowStoreConfig{Type: "memory"}
		store, err := NewWorkflowStore(ctx, config)
		if err != nil {
			t.Fatalf("failed to create memory store: %v", err)
		}
		defer store.Close()

		// Verify it works
		err = store.SaveState(ctx, "test", &WorkflowState{
			ID:        "test",
			Name:      "test",
			Status:    WorkflowStatusPending,
			StartedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("failed to save state: %v", err)
		}
	})

	t.Run("validates config", func(t *testing.T) {
		config := WorkflowStoreConfig{Type: "invalid"}
		_, err := NewWorkflowStore(ctx, config)
		if err == nil {
			t.Error("expected error for invalid type")
		}
	})

	t.Run("redis requires URL", func(t *testing.T) {
		config := WorkflowStoreConfig{Type: "redis"}
		_, err := NewWorkflowStore(ctx, config)
		if err == nil {
			t.Error("expected error when redis_url is missing")
		}
	})
}

func TestWorkflowStateMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal state", func(t *testing.T) {
		now := time.Now()
		completed := now.Add(time.Hour)
		state := &WorkflowState{
			ID:             "test-marshal",
			Name:           "test-workflow",
			Status:         WorkflowStatusCompleted,
			CurrentStep:    "step3",
			Data:           map[string]interface{}{"key": "value", "count": float64(42)},
			Error:          "",
			RetryCount:     2,
			MaxRetries:     5,
			StartedAt:      now,
			UpdatedAt:      now.Add(30 * time.Minute),
			CompletedAt:    &completed,
			CompletedSteps: []string{"step1", "step2", "step3"},
			Metadata:       map[string]string{"env": "test"},
		}

		data, err := MarshalWorkflowState(state)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		restored, err := UnmarshalWorkflowState(data)
		if err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if restored.ID != state.ID {
			t.Errorf("ID mismatch: %s != %s", restored.ID, state.ID)
		}
		if restored.Status != state.Status {
			t.Errorf("Status mismatch: %s != %s", restored.Status, state.Status)
		}
		if len(restored.CompletedSteps) != len(state.CompletedSteps) {
			t.Errorf("CompletedSteps length mismatch: %d != %d", len(restored.CompletedSteps), len(state.CompletedSteps))
		}
	})

	t.Run("marshal and unmarshal checkpoint", func(t *testing.T) {
		cp := &Checkpoint{
			ID:         "cp-1",
			WorkflowID: "wf-1",
			Step:       "step1",
			Data:       map[string]interface{}{"result": "success"},
			Status:     CheckpointCompleted,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		data, err := MarshalCheckpoint(cp)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		restored, err := UnmarshalCheckpoint(data)
		if err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if restored.ID != cp.ID {
			t.Errorf("ID mismatch")
		}
		if restored.Status != cp.Status {
			t.Errorf("Status mismatch")
		}
	})

	t.Run("marshal and unmarshal history event", func(t *testing.T) {
		event := &HistoryEvent{
			ID:          "evt-1",
			WorkflowID:  "wf-1",
			EventType:   HistoryEventStepCompleted,
			Step:        "step1",
			Status:      WorkflowStatusRunning,
			Timestamp:   time.Now(),
			DurationMs:  1500,
			Description: "Step completed successfully",
		}

		data, err := MarshalHistoryEvent(event)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		restored, err := UnmarshalHistoryEvent(data)
		if err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if restored.ID != event.ID {
			t.Errorf("ID mismatch")
		}
		if restored.EventType != event.EventType {
			t.Errorf("EventType mismatch")
		}
	})
}

func TestWorkflowRecovery(t *testing.T) {
	ctx := context.Background()

	t.Run("resumes from checkpoint", func(t *testing.T) {
		// Create store with pre-existing workflow state
		store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
		defer store.Close()

		// Save a workflow that was interrupted mid-execution
		state := &WorkflowState{
			ID:             "recovery-test-1",
			Name:           "recoverable-workflow",
			Status:         WorkflowStatusRunning,
			CurrentStep:    "step2",
			Data:           map[string]interface{}{"accumulated": "step1-result"},
			StartedAt:      time.Now().Add(-time.Hour),
			UpdatedAt:      time.Now().Add(-time.Minute),
			CompletedSteps: []string{"step1"},
		}
		_ = store.SaveState(ctx, state.ID, state)

		// Save checkpoint for completed step
		_ = store.SaveCheckpoint(ctx, state.ID, &Checkpoint{
			ID:         state.ID + "-step1",
			WorkflowID: state.ID,
			Step:       "step1",
			Status:     CheckpointCompleted,
			Data:       "step1-result",
			CreatedAt:  time.Now().Add(-30 * time.Minute),
			UpdatedAt:  time.Now().Add(-30 * time.Minute),
		})

		// List running workflows (simulating recovery)
		running, err := store.ListStates(ctx, StateFilter{Status: WorkflowStatusRunning})
		if err != nil {
			t.Fatalf("failed to list running: %v", err)
		}
		if len(running) != 1 {
			t.Fatalf("expected 1 running workflow, got %d", len(running))
		}

		// Load checkpoint
		checkpoint, err := store.LoadLatestCheckpoint(ctx, running[0].ID)
		if err != nil {
			t.Fatalf("failed to load checkpoint: %v", err)
		}
		if checkpoint == nil {
			t.Fatal("expected checkpoint")
		}
		if checkpoint.Step != "step1" {
			t.Errorf("expected checkpoint at step1, got %s", checkpoint.Step)
		}

		// Verify we know where to resume from
		if running[0].CurrentStep != "step2" {
			t.Errorf("expected to resume at step2, got %s", running[0].CurrentStep)
		}
		if len(running[0].CompletedSteps) != 1 || running[0].CompletedSteps[0] != "step1" {
			t.Errorf("expected step1 completed, got %v", running[0].CompletedSteps)
		}
	})

	t.Run("skips completed steps on recovery", func(t *testing.T) {
		runner := NewWorkflowRunner(NewInMemoryCheckpointStore())

		step1Executed := false
		step2Executed := false
		step3Executed := false

		runner.AddStep(WorkflowStep{
			Name: "step1",
			Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
				step1Executed = true
				return "step1-result", nil
			},
		})
		runner.AddStep(WorkflowStep{
			Name: "step2",
			Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
				step2Executed = true
				return "step2-result", nil
			},
		})
		runner.AddStep(WorkflowStep{
			Name: "step3",
			Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
				step3Executed = true
				return "step3-result", nil
			},
		})

		// Pre-populate checkpoints for steps 1 and 2
		store := runner.store.(*InMemoryCheckpointStore)
		_ = store.Save(ctx, &Checkpoint{
			ID:         "recovery-wf-step1",
			WorkflowID: "recovery-wf",
			Step:       "step1",
			Status:     CheckpointCompleted,
			Data:       "step1-data",
		})
		_ = store.Save(ctx, &Checkpoint{
			ID:         "recovery-wf-step2",
			WorkflowID: "recovery-wf",
			Step:       "step2",
			Status:     CheckpointCompleted,
			Data:       "step2-data",
		})

		// Run the workflow (should skip steps 1 and 2)
		err := runner.Run(ctx, "recovery-wf", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if step1Executed {
			t.Error("step1 should not have been executed")
		}
		if step2Executed {
			t.Error("step2 should not have been executed")
		}
		if !step3Executed {
			t.Error("step3 should have been executed")
		}
	})
}

func TestWorkflowStateTransitions(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
	defer store.Close()

	t.Run("valid state transitions", func(t *testing.T) {
		transitions := []struct {
			from WorkflowStatus
			to   WorkflowStatus
		}{
			{WorkflowStatusPending, WorkflowStatusRunning},
			{WorkflowStatusRunning, WorkflowStatusPaused},
			{WorkflowStatusPaused, WorkflowStatusRunning},
			{WorkflowStatusRunning, WorkflowStatusCompleted},
			{WorkflowStatusRunning, WorkflowStatusFailed},
			{WorkflowStatusRunning, WorkflowStatusCancelled},
		}

		for _, tr := range transitions {
			state := &WorkflowState{
				ID:        "state-transition-" + string(tr.from) + "-" + string(tr.to),
				Name:      "test",
				Status:    tr.from,
				StartedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_ = store.SaveState(ctx, state.ID, state)

			// Transition
			state.Status = tr.to
			state.UpdatedAt = time.Now()
			err := store.SaveState(ctx, state.ID, state)
			if err != nil {
				t.Errorf("failed to save transition %s -> %s: %v", tr.from, tr.to, err)
			}

			loaded, _ := store.LoadState(ctx, state.ID)
			if loaded.Status != tr.to {
				t.Errorf("expected status %s, got %s", tr.to, loaded.Status)
			}
		}
	})
}

func TestHistoryEventTypes(t *testing.T) {
	// Verify all event types have string values
	eventTypes := []HistoryEventType{
		HistoryEventWorkflowStarted,
		HistoryEventStepStarted,
		HistoryEventStepCompleted,
		HistoryEventStepFailed,
		HistoryEventStepRetried,
		HistoryEventStepSkipped,
		HistoryEventWorkflowPaused,
		HistoryEventWorkflowResumed,
		HistoryEventWorkflowCompleted,
		HistoryEventWorkflowFailed,
		HistoryEventWorkflowCancelled,
		HistoryEventCheckpointSaved,
		HistoryEventRecoveryStarted,
	}

	for _, et := range eventTypes {
		if string(et) == "" {
			t.Errorf("event type has empty string value: %v", et)
		}
	}
}

func TestDeepCopyIsolation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
	defer store.Close()

	t.Run("modifying loaded state doesn't affect stored state", func(t *testing.T) {
		original := &WorkflowState{
			ID:        "isolation-test",
			Name:      "test",
			Status:    WorkflowStatusRunning,
			Data:      map[string]interface{}{"key": "original"},
			StartedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = store.SaveState(ctx, original.ID, original)

		// Load and modify
		loaded, _ := store.LoadState(ctx, original.ID)
		loaded.Data["key"] = "modified"
		loaded.Status = WorkflowStatusCompleted

		// Load again - should be unchanged
		fresh, _ := store.LoadState(ctx, original.ID)
		if fresh.Data["key"] != "original" {
			t.Error("stored state was modified by external changes to loaded state")
		}
		if fresh.Status != WorkflowStatusRunning {
			t.Error("stored status was modified by external changes")
		}
	})

	t.Run("modifying input state doesn't affect stored state", func(t *testing.T) {
		state := &WorkflowState{
			ID:        "isolation-test-2",
			Name:      "test",
			Status:    WorkflowStatusRunning,
			Data:      map[string]interface{}{"key": "original"},
			StartedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = store.SaveState(ctx, state.ID, state)

		// Modify original after save
		state.Data["key"] = "modified"

		// Load - should have original value
		loaded, _ := store.LoadState(ctx, state.ID)
		if loaded.Data["key"] != "original" {
			t.Error("stored state was affected by changes to input state after save")
		}
	})
}

func TestStepRetryWithPersistence(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryCheckpointStore()
	runner := NewWorkflowRunner(store)

	attempts := 0
	runner.AddStep(WorkflowStep{
		Name:       "flaky-step",
		CanRetry:   true,
		MaxRetries: 3,
		Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("temporary failure")
			}
			return "success", nil
		},
	})

	err := runner.Run(ctx, "retry-test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	// Verify final checkpoint is marked complete
	cp, _ := store.Get(ctx, "retry-test", "flaky-step")
	if cp == nil {
		t.Fatal("checkpoint should exist")
	}
	if cp.Status != CheckpointCompleted {
		t.Errorf("expected checkpoint completed, got %s", cp.Status)
	}
}

func TestWorkflowDataPersistence(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryWorkflowStore(DefaultWorkflowStoreConfig())
	defer store.Close()

	t.Run("complex data types persist correctly", func(t *testing.T) {
		state := &WorkflowState{
			ID:     "complex-data",
			Name:   "test",
			Status: WorkflowStatusRunning,
			Data: map[string]interface{}{
				"string":  "value",
				"number":  float64(42),
				"float":   3.14,
				"bool":    true,
				"array":   []interface{}{"a", "b", "c"},
				"nested":  map[string]interface{}{"inner": "value"},
				"null":    nil,
			},
			StartedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_ = store.SaveState(ctx, state.ID, state)
		loaded, _ := store.LoadState(ctx, state.ID)

		// Verify complex types
		if loaded.Data["string"] != "value" {
			t.Error("string not preserved")
		}
		if loaded.Data["number"] != float64(42) {
			t.Error("number not preserved")
		}
		if loaded.Data["bool"] != true {
			t.Error("bool not preserved")
		}

		arr, ok := loaded.Data["array"].([]interface{})
		if !ok || len(arr) != 3 {
			t.Error("array not preserved")
		}

		nested, ok := loaded.Data["nested"].(map[string]interface{})
		if !ok || nested["inner"] != "value" {
			t.Error("nested object not preserved")
		}
	})
}

