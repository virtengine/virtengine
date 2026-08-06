package privileged

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

type PrivilegedAuditEntry struct {
	Sequence       uint64      `json:"sequence"`
	OccurredAt     int64       `json:"occurred_at"`
	ActorID        string      `json:"actor_id"`
	TargetID       string      `json:"target_id"`
	Action         string      `json:"action"`
	Scope          []ScopeAtom `json:"scope"`
	ActionDigest   [32]byte    `json:"action_digest"`
	EvidenceDigest [32]byte    `json:"evidence_digest"`
	PreviousHash   [32]byte    `json:"previous_hash"`
	Hash           [32]byte    `json:"hash"`
}

func (e PrivilegedAuditEntry) CanonicalHash() ([32]byte, error) {
	if e.Sequence == 0 || e.OccurredAt <= 0 || invalidExactValue(e.ActorID) || invalidExactValue(e.TargetID) || invalidExactValue(e.Action) {
		return [32]byte{}, fmt.Errorf("audit sequence, time, actor, target, and action are required")
	}
	if err := validateScope(e.Scope); err != nil {
		return [32]byte{}, err
	}
	exactDigest, err := ScopeDigest(e.Action, e.Scope)
	if err != nil {
		return [32]byte{}, err
	}
	if e.ActionDigest != exactDigest || e.EvidenceDigest == ([32]byte{}) {
		return [32]byte{}, fmt.Errorf("audit exact action or evidence digest mismatch")
	}
	var output bytes.Buffer
	_ = writeString(&output, "virtengine.roles.privileged/audit-entry/v1")
	writeUint64(&output, e.Sequence)
	writeInt64(&output, e.OccurredAt)
	_ = writeString(&output, e.ActorID)
	_ = writeString(&output, e.TargetID)
	_ = writeString(&output, e.Action)
	for _, atom := range sortedScope(e.Scope) {
		_ = writeString(&output, atom.ResourceType)
		_ = writeString(&output, atom.ResourceID)
		_ = writeString(&output, atom.Operation)
	}
	output.Write(e.ActionDigest[:])
	output.Write(e.EvidenceDigest[:])
	output.Write(e.PreviousHash[:])
	return sha256.Sum256(output.Bytes()), nil
}

func NewAuditEntry(sequence uint64, occurredAt int64, actorID, targetID, action string, scope []ScopeAtom, evidenceDigest, previousHash [32]byte) (PrivilegedAuditEntry, error) {
	actionDigest, err := ScopeDigest(action, scope)
	if err != nil {
		return PrivilegedAuditEntry{}, err
	}
	entry := PrivilegedAuditEntry{Sequence: sequence, OccurredAt: occurredAt, ActorID: actorID, TargetID: targetID, Action: action, Scope: append([]ScopeAtom(nil), scope...), ActionDigest: actionDigest, EvidenceDigest: evidenceDigest, PreviousHash: previousHash}
	entry.Hash, err = entry.CanonicalHash()
	return entry, err
}

type AuditCheckpoint struct {
	Sequence       uint64   `json:"sequence"`
	EntryHash      [32]byte `json:"entry_hash"`
	PreviousHash   [32]byte `json:"previous_hash"`
	CheckpointHash [32]byte `json:"checkpoint_hash"`
}

func (c AuditCheckpoint) CanonicalHash() ([32]byte, error) {
	if c.Sequence == 0 || c.EntryHash == ([32]byte{}) {
		return [32]byte{}, fmt.Errorf("checkpoint sequence and entry hash are required")
	}
	var output bytes.Buffer
	_ = writeString(&output, "virtengine.roles.privileged/audit-checkpoint/v1")
	writeUint64(&output, c.Sequence)
	output.Write(c.EntryHash[:])
	output.Write(c.PreviousHash[:])
	return sha256.Sum256(output.Bytes()), nil
}

func NewAuditCheckpoint(entry PrivilegedAuditEntry) (AuditCheckpoint, error) {
	checkpoint := AuditCheckpoint{Sequence: entry.Sequence, EntryHash: entry.Hash, PreviousHash: entry.PreviousHash}
	var err error
	checkpoint.CheckpointHash, err = checkpoint.CanonicalHash()
	return checkpoint, err
}

func VerifyAuditChain(entries []PrivilegedAuditEntry, checkpoints []AuditCheckpoint) error {
	if len(entries) == 0 {
		return fmt.Errorf("audit chain cannot be empty")
	}
	if len(checkpoints) == 0 {
		return fmt.Errorf("audit chain requires an externally retained checkpoint")
	}
	checkpointBySequence := make(map[uint64]AuditCheckpoint, len(checkpoints))
	for _, checkpoint := range checkpoints {
		if _, duplicate := checkpointBySequence[checkpoint.Sequence]; duplicate {
			return fmt.Errorf("duplicate audit checkpoint at sequence %d", checkpoint.Sequence)
		}
		checkpointBySequence[checkpoint.Sequence] = checkpoint
	}
	var previous [32]byte
	for index, entry := range entries {
		expectedSequence := uint64(index + 1)
		if entry.Sequence != expectedSequence {
			return fmt.Errorf("audit sequence gap or reorder at %d", expectedSequence)
		}
		if entry.PreviousHash != previous {
			return fmt.Errorf("audit predecessor mismatch at sequence %d", entry.Sequence)
		}
		hash, err := entry.CanonicalHash()
		if err != nil {
			return err
		}
		if entry.Hash != hash {
			return fmt.Errorf("audit entry mutation at sequence %d", entry.Sequence)
		}
		if checkpoint, exists := checkpointBySequence[entry.Sequence]; exists {
			checkpointHash, err := checkpoint.CanonicalHash()
			if err != nil {
				return err
			}
			if checkpoint.EntryHash != entry.Hash || checkpoint.PreviousHash != entry.PreviousHash || checkpoint.CheckpointHash != checkpointHash {
				return fmt.Errorf("audit checkpoint mismatch at sequence %d", entry.Sequence)
			}
		}
		previous = entry.Hash
	}
	for sequence := range checkpointBySequence {
		if sequence > uint64(len(entries)) {
			return fmt.Errorf("audit checkpoint references missing entry %d", sequence)
		}
	}
	if _, exists := checkpointBySequence[uint64(len(entries))]; !exists {
		return fmt.Errorf("audit chain tip is not checkpointed")
	}
	return nil
}
