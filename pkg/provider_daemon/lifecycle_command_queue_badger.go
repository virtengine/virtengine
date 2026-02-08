// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-34E: Badger-backed lifecycle command queue storage.
package provider_daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
)

const (
	lifecycleQueueSchemaVersion      = 1
	lifecycleQueueCmdPrefix          = "lifecyclequeue/cmd/"
	lifecycleQueueIdempotencyPrefix  = "lifecyclequeue/idem/"
	lifecycleQueueReadyPrefix        = "lifecyclequeue/ready/"
	lifecycleQueueAllocationPrefix   = "lifecyclequeue/alloc/"
	lifecycleQueueDesiredStatePrefix = "lifecyclequeue/desired/"
	lifecycleQueueSchemaKey          = "lifecyclequeue/meta/schema_version"
)

// BadgerLifecycleCommandStore implements LifecycleCommandStore using Badger.
type BadgerLifecycleCommandStore struct {
	db *badger.DB
}

// OpenBadgerLifecycleCommandStore opens or creates the Badger-backed store.
func OpenBadgerLifecycleCommandStore(path string, inMemory bool) (*BadgerLifecycleCommandStore, error) {
	if path == "" {
		path = "data/lifecycle_queue"
	}

	opts := badger.DefaultOptions(path).WithLogger(nil)
	if inMemory {
		opts = badger.DefaultOptions("").WithLogger(nil).WithInMemory(true)
	} else {
		opts = opts.WithValueDir(filepath.Join(path, "value"))
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	store := &BadgerLifecycleCommandStore{db: db}
	if err := store.ensureSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *BadgerLifecycleCommandStore) ensureSchema() error {
	return s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lifecycleQueueSchemaKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return txn.Set([]byte(lifecycleQueueSchemaKey), []byte(fmt.Sprintf("%d", lifecycleQueueSchemaVersion)))
			}
			return err
		}

		return item.Value(func(val []byte) error {
			version := strings.TrimSpace(string(val))
			if version == fmt.Sprintf("%d", lifecycleQueueSchemaVersion) {
				return nil
			}
			return fmt.Errorf("unsupported lifecycle queue schema version %s", version)
		})
	})
}

// Enqueue stores a new command, enforcing idempotency by IdempotencyKey.
func (s *BadgerLifecycleCommandStore) Enqueue(ctx context.Context, cmd *LifecycleCommand) (*LifecycleCommand, bool, error) {
	if cmd == nil {
		return nil, false, fmt.Errorf("command is nil")
	}

	var stored *LifecycleCommand
	existingFound := false
	err := s.db.Update(func(txn *badger.Txn) error {
		if cmd.IdempotencyKey != "" {
			item, err := txn.Get(idempotencyKey(cmd.IdempotencyKey))
			if err == nil {
				return item.Value(func(val []byte) error {
					existingID := string(val)
					existing, err := readCommand(txn, existingID)
					if err != nil {
						return err
					}
					stored = existing
					existingFound = true
					return nil
				})
			}
			if err != nil && err != badger.ErrKeyNotFound {
				return err
			}
		}
		if existingFound {
			return nil
		}

		data, err := json.Marshal(cmd)
		if err != nil {
			return err
		}
		if err := txn.Set(cmdKey(cmd.ID), data); err != nil {
			return err
		}

		if cmd.IdempotencyKey != "" {
			if err := txn.Set(idempotencyKey(cmd.IdempotencyKey), []byte(cmd.ID)); err != nil {
				return err
			}
		}

		if err := txn.Set(allocationKey(cmd.AllocationID, cmd.ID), []byte(cmd.ID)); err != nil {
			return err
		}

		if cmd.Status == LifecycleCommandStatusPending && cmd.NextAttemptAt != nil {
			if err := txn.Set(readyKey(*cmd.NextAttemptAt, cmd.ID), []byte(cmd.ID)); err != nil {
				return err
			}
		}

		stored = cmd
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	if existingFound && stored != nil {
		return stored, true, nil
	}
	return stored, false, nil
}

// Get loads a command by ID.
func (s *BadgerLifecycleCommandStore) Get(ctx context.Context, id string) (*LifecycleCommand, error) {
	var cmd *LifecycleCommand
	err := s.db.View(func(txn *badger.Txn) error {
		var err error
		cmd, err = readCommand(txn, id)
		return err
	})
	return cmd, err
}

// Update replaces a command and updates indexes.
func (s *BadgerLifecycleCommandStore) Update(ctx context.Context, cmd *LifecycleCommand) error {
	if cmd == nil {
		return fmt.Errorf("command is nil")
	}

	return s.db.Update(func(txn *badger.Txn) error {
		existing, err := readCommand(txn, cmd.ID)
		if err != nil {
			return err
		}

		if existing != nil && existing.NextAttemptAt != nil {
			_ = txn.Delete(readyKey(*existing.NextAttemptAt, existing.ID))
		}

		data, err := json.Marshal(cmd)
		if err != nil {
			return err
		}
		if err := txn.Set(cmdKey(cmd.ID), data); err != nil {
			return err
		}

		if cmd.Status == LifecycleCommandStatusPending && cmd.NextAttemptAt != nil {
			if err := txn.Set(readyKey(*cmd.NextAttemptAt, cmd.ID), []byte(cmd.ID)); err != nil {
				return err
			}
		}

		return nil
	})
}

// ClaimNextReady picks the next ready command and marks it executing.
func (s *BadgerLifecycleCommandStore) ClaimNextReady(ctx context.Context, now time.Time, workerID string) (*LifecycleCommand, error) {
	var claimed *LifecycleCommand
	err := s.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(lifecycleQueueReadyPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.KeyCopy(nil)
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			cmdID := string(val)
			cmd, err := readCommand(txn, cmdID)
			if err != nil {
				return err
			}
			if cmd == nil || cmd.NextAttemptAt == nil {
				_ = txn.Delete(key)
				continue
			}
			if cmd.NextAttemptAt.After(now) {
				return nil
			}

			_ = txn.Delete(key)

			cmd.Status = LifecycleCommandStatusExecuting
			cmd.AttemptCount++
			cmd.LastAttemptAt = &now
			cmd.UpdatedAt = now

			data, err := json.Marshal(cmd)
			if err != nil {
				return err
			}
			if err := txn.Set(cmdKey(cmd.ID), data); err != nil {
				return err
			}

			claimed = cmd
			return nil
		}

		return nil
	})
	return claimed, err
}

// ListByStatus returns commands with the provided statuses.
func (s *BadgerLifecycleCommandStore) ListByStatus(ctx context.Context, statuses ...LifecycleCommandStatus) ([]*LifecycleCommand, error) {
	allowed := map[LifecycleCommandStatus]bool{}
	for _, status := range statuses {
		allowed[status] = true
	}

	var commands []*LifecycleCommand
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(lifecycleQueueCmdPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var cmd LifecycleCommand
			if err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &cmd)
			}); err != nil {
				return err
			}
			if len(allowed) == 0 || allowed[cmd.Status] {
				copyCmd := cmd
				commands = append(commands, &copyCmd)
			}
		}
		return nil
	})
	return commands, err
}

// ListByAllocation returns commands for a specific allocation.
func (s *BadgerLifecycleCommandStore) ListByAllocation(ctx context.Context, allocationID string, statuses ...LifecycleCommandStatus) ([]*LifecycleCommand, error) {
	if allocationID == "" {
		return nil, fmt.Errorf("allocation ID is required")
	}
	allowed := map[LifecycleCommandStatus]bool{}
	for _, status := range statuses {
		allowed[status] = true
	}

	var commands []*LifecycleCommand
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(lifecycleQueueAllocationPrefix + allocationID + "/")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			cmdID := string(val)
			cmd, err := readCommand(txn, cmdID)
			if err != nil {
				return err
			}
			if cmd == nil {
				continue
			}
			if len(allowed) == 0 || allowed[cmd.Status] {
				commands = append(commands, cmd)
			}
		}
		return nil
	})
	return commands, err
}

// ListDesiredStates returns all desired state records.
func (s *BadgerLifecycleCommandStore) ListDesiredStates(ctx context.Context) ([]*LifecycleDesiredState, error) {
	var states []*LifecycleDesiredState
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(lifecycleQueueDesiredStatePrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var state LifecycleDesiredState
			if err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &state)
			}); err != nil {
				return err
			}
			copyState := state
			states = append(states, &copyState)
		}
		return nil
	})
	return states, err
}

// GetDesiredState fetches a desired state record for an allocation.
func (s *BadgerLifecycleCommandStore) GetDesiredState(ctx context.Context, allocationID string) (*LifecycleDesiredState, error) {
	var state *LifecycleDesiredState
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(desiredStateKey(allocationID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			var decoded LifecycleDesiredState
			if err := json.Unmarshal(val, &decoded); err != nil {
				return err
			}
			state = &decoded
			return nil
		})
	})
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	return state, err
}

// SetDesiredState stores the desired state record.
func (s *BadgerLifecycleCommandStore) SetDesiredState(ctx context.Context, state *LifecycleDesiredState) error {
	if state == nil {
		return fmt.Errorf("desired state is nil")
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(desiredStateKey(state.AllocationID), data)
	})
}

// Close closes the underlying Badger DB.
func (s *BadgerLifecycleCommandStore) Close() error {
	return s.db.Close()
}

func cmdKey(id string) []byte {
	return []byte(lifecycleQueueCmdPrefix + id)
}

func idempotencyKey(id string) []byte {
	return []byte(lifecycleQueueIdempotencyPrefix + id)
}

func readyKey(at time.Time, id string) []byte {
	return []byte(fmt.Sprintf("%s%020d/%s", lifecycleQueueReadyPrefix, at.UnixNano(), id))
}

func allocationKey(allocationID, id string) []byte {
	return []byte(lifecycleQueueAllocationPrefix + allocationID + "/" + id)
}

func desiredStateKey(allocationID string) []byte {
	return []byte(lifecycleQueueDesiredStatePrefix + allocationID)
}

func readCommand(txn *badger.Txn, id string) (*LifecycleCommand, error) {
	item, err := txn.Get(cmdKey(id))
	if err != nil {
		return nil, err
	}
	var cmd LifecycleCommand
	if err := item.Value(func(val []byte) error {
		return json.Unmarshal(val, &cmd)
	}); err != nil {
		return nil, err
	}
	return &cmd, nil
}
