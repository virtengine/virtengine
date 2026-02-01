// Package workflow provides workflow testing
package workflow

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStateMachine(t *testing.T) {
	t.Run("basic transitions", func(t *testing.T) {
		sm := NewStateMachine("test", "pending")
		sm.AddStates("pending", "processing", "completed", "failed")
		sm.AddTransition(Transition{
			From:  "pending",
			To:    "processing",
			Event: "start",
		})
		sm.AddTransition(Transition{
			From:  "processing",
			To:    "completed",
			Event: "complete",
		})
		sm.AddTransition(Transition{
			From:  "processing",
			To:    "failed",
			Event: "fail",
		})

		if sm.Current() != "pending" {
			t.Errorf("expected pending, got %s", sm.Current())
		}

		if !sm.Can("start") {
			t.Error("expected to be able to start")
		}

		if sm.Can("complete") {
			t.Error("should not be able to complete from pending")
		}

		err := sm.Fire(context.Background(), "start")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if sm.Current() != "processing" {
			t.Errorf("expected processing, got %s", sm.Current())
		}

		err = sm.Fire(context.Background(), "complete")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if sm.Current() != "completed" {
			t.Errorf("expected completed, got %s", sm.Current())
		}
	})

	t.Run("transition with guard", func(t *testing.T) {
		sm := NewStateMachine("test", "pending")
		sm.AddStates("pending", "approved", "rejected")

		type request struct {
			Amount int
		}
		sm.SetData(&request{Amount: 100})

		sm.AddTransition(Transition{
			From:  "pending",
			To:    "approved",
			Event: "approve",
			Guard: func(ctx context.Context, data interface{}) bool {
				req := data.(*request)
				return req.Amount <= 50
			},
		})

		if sm.Can("approve") {
			t.Error("guard should prevent approval")
		}

		sm.SetData(&request{Amount: 30})
		if !sm.Can("approve") {
			t.Error("should be able to approve with lower amount")
		}
	})

	t.Run("transition with action", func(t *testing.T) {
		executed := false
		sm := NewStateMachine("test", "pending")
		sm.AddStates("pending", "done")
		sm.AddTransition(Transition{
			From:  "pending",
			To:    "done",
			Event: "do",
			Action: func(ctx context.Context, data interface{}) error {
				executed = true
				return nil
			},
		})

		err := sm.Fire(context.Background(), "do")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !executed {
			t.Error("action was not executed")
		}
	})

	t.Run("action error prevents transition", func(t *testing.T) {
		sm := NewStateMachine("test", "pending")
		sm.AddStates("pending", "done")
		sm.AddTransition(Transition{
			From:  "pending",
			To:    "done",
			Event: "do",
			Action: func(ctx context.Context, data interface{}) error {
				return errors.New("action failed")
			},
		})

		err := sm.Fire(context.Background(), "do")
		if err == nil {
			t.Error("expected error")
		}

		if sm.Current() != "pending" {
			t.Error("state should not have changed")
		}
	})

	t.Run("history tracking", func(t *testing.T) {
		sm := NewStateMachine("test", "a")
		sm.AddStates("a", "b", "c")
		sm.AddTransition(Transition{From: "a", To: "b", Event: "go"})
		sm.AddTransition(Transition{From: "b", To: "c", Event: "go"})

		_ = sm.Fire(context.Background(), "go")
		_ = sm.Fire(context.Background(), "go")

		history := sm.History()
		if len(history) != 2 {
			t.Errorf("expected 2 history entries, got %d", len(history))
		}

		if history[0].From != "a" || history[0].To != "b" {
			t.Error("first transition incorrect")
		}

		if history[1].From != "b" || history[1].To != "c" {
			t.Error("second transition incorrect")
		}
	})

	t.Run("reset", func(t *testing.T) {
		sm := NewStateMachine("test", "a")
		sm.AddStates("a", "b")
		sm.AddTransition(Transition{From: "a", To: "b", Event: "go"})

		_ = sm.Fire(context.Background(), "go")
		if sm.Current() != "b" {
			t.Error("expected state b")
		}

		sm.Reset()
		if sm.Current() != "a" {
			t.Error("expected state a after reset")
		}

		if len(sm.History()) != 0 {
			t.Error("history should be cleared after reset")
		}
	})
}

func TestIdempotencyKey(t *testing.T) {
	t.Run("generates consistent keys", func(t *testing.T) {
		key1 := GenerateIdempotencyKey("user123", "action", 100)
		key2 := GenerateIdempotencyKey("user123", "action", 100)

		if key1 != key2 {
			t.Error("same inputs should generate same key")
		}
	})

	t.Run("different inputs generate different keys", func(t *testing.T) {
		key1 := GenerateIdempotencyKey("user123", "action", 100)
		key2 := GenerateIdempotencyKey("user456", "action", 100)

		if key1 == key2 {
			t.Error("different inputs should generate different keys")
		}
	})
}

func TestIdempotentHandler(t *testing.T) {
	t.Run("executes handler once", func(t *testing.T) {
		store := NewInMemoryIdempotencyStore()
		handler := NewIdempotentHandler(store, time.Hour)

		execCount := 0
		fn := func(ctx context.Context) (interface{}, error) {
			execCount++
			return "result", nil
		}

		key := GenerateIdempotencyKey("test")

		result1, err := handler.Execute(context.Background(), key, fn)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		result2, err := handler.Execute(context.Background(), key, fn)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if execCount != 1 {
			t.Errorf("expected 1 execution, got %d", execCount)
		}

		if result1 != result2 {
			t.Error("results should be equal")
		}
	})

	t.Run("calls onReplay for cached results", func(t *testing.T) {
		store := NewInMemoryIdempotencyStore()
		handler := NewIdempotentHandler(store, time.Hour)

		replayCount := 0
		handler.OnReplay(func(key IdempotencyKey, result interface{}) {
			replayCount++
		})

		key := GenerateIdempotencyKey("test")
		fn := func(ctx context.Context) (interface{}, error) {
			return "result", nil
		}

		_, _ = handler.Execute(context.Background(), key, fn)
		_, _ = handler.Execute(context.Background(), key, fn)
		_, _ = handler.Execute(context.Background(), key, fn)

		if replayCount != 2 {
			t.Errorf("expected 2 replays, got %d", replayCount)
		}
	})

	t.Run("does not store failed results", func(t *testing.T) {
		store := NewInMemoryIdempotencyStore()
		handler := NewIdempotentHandler(store, time.Hour)

		execCount := 0
		fn := func(ctx context.Context) (interface{}, error) {
			execCount++
			return nil, errors.New("failed")
		}

		key := GenerateIdempotencyKey("test")

		_, _ = handler.Execute(context.Background(), key, fn)
		_, _ = handler.Execute(context.Background(), key, fn)

		if execCount != 2 {
			t.Errorf("expected 2 executions for failed handler, got %d", execCount)
		}
	})
}

func TestWorkflowRunner(t *testing.T) {
	t.Run("executes all steps", func(t *testing.T) {
		store := NewInMemoryCheckpointStore()
		runner := NewWorkflowRunner(store)

		steps := []string{}
		runner.AddStep(WorkflowStep{
			Name: "step1",
			Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
				steps = append(steps, "step1")
				return "step1-result", nil
			},
		})
		runner.AddStep(WorkflowStep{
			Name: "step2",
			Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
				steps = append(steps, "step2")
				return "step2-result", nil
			},
		})
		runner.AddStep(WorkflowStep{
			Name: "step3",
			Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
				steps = append(steps, "step3")
				return "step3-result", nil
			},
		})

		err := runner.Run(context.Background(), "workflow-1", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(steps) != 3 {
			t.Errorf("expected 3 steps, got %d", len(steps))
		}
	})

	t.Run("resumes from checkpoint", func(t *testing.T) {
		store := NewInMemoryCheckpointStore()
		runner := NewWorkflowRunner(store)

		// Pre-create checkpoints for first two steps
		_ = store.Save(context.Background(), &Checkpoint{
			ID:         "workflow-2-step1",
			WorkflowID: "workflow-2",
			Step:       "step1",
			Status:     CheckpointCompleted,
			Data:       "step1-data",
		})
		_ = store.Save(context.Background(), &Checkpoint{
			ID:         "workflow-2-step2",
			WorkflowID: "workflow-2",
			Step:       "step2",
			Status:     CheckpointCompleted,
			Data:       "step2-data",
		})

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

		err := runner.Run(context.Background(), "workflow-2", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if step1Executed || step2Executed {
			t.Error("completed steps should not be re-executed")
		}

		if !step3Executed {
			t.Error("remaining step should be executed")
		}
	})

	t.Run("retries on failure", func(t *testing.T) {
		store := NewInMemoryCheckpointStore()
		runner := NewWorkflowRunner(store)

		attempts := 0
		runner.AddStep(WorkflowStep{
			Name:       "flaky",
			CanRetry:   true,
			MaxRetries: 3,
			Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
				attempts++
				if attempts < 3 {
					return nil, errors.New("temporary error")
				}
				return "success", nil
			},
		})

		err := runner.Run(context.Background(), "workflow-3", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("compensates on error", func(t *testing.T) {
		store := NewInMemoryCheckpointStore()
		runner := NewWorkflowRunner(store)

		compensated := []string{}

		runner.AddStep(WorkflowStep{
			Name: "step1",
			Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
				return "step1", nil
			},
			Compensate: func(ctx context.Context, data interface{}) error {
				compensated = append(compensated, "step1")
				return nil
			},
		})
		runner.AddStep(WorkflowStep{
			Name: "step2",
			Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
				return "step2", nil
			},
			Compensate: func(ctx context.Context, data interface{}) error {
				compensated = append(compensated, "step2")
				return nil
			},
		})
		runner.AddStep(WorkflowStep{
			Name: "step3",
			Execute: func(ctx context.Context, data interface{}) (interface{}, error) {
				return nil, errors.New("step3 failed")
			},
		})

		runner.OnError(func(ctx context.Context, step string, err error) error {
			return err // Return error to trigger compensation
		})

		err := runner.Run(context.Background(), "workflow-4", nil)
		if err == nil {
			t.Error("expected error")
		}

		// Compensations happen in reverse order
		if len(compensated) != 2 {
			t.Errorf("expected 2 compensations, got %d", len(compensated))
		}

		if compensated[0] != "step2" || compensated[1] != "step1" {
			t.Error("compensations should happen in reverse order")
		}
	})
}

