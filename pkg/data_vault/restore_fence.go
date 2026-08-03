package data_vault

import (
	"context"
	"errors"

	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
)

type ErasureFence struct {
	TargetCommitment      string `json:"target_commitment"`
	BackupGenerationDigest string `json:"backup_generation_digest"`
	ErasureEpoch          uint64 `json:"erasure_epoch"`
	ErasedHeight          int64  `json:"erased_height"`
	ErasedUnix            int64  `json:"erased_unix"`
}

type TombstoneInventory struct {
	TargetCommitment string `json:"target_commitment"`
	BackupGenerationDigest string `json:"backup_generation_digest"`
	SnapshotManifestDigest string `json:"snapshot_manifest_digest"`
	SnapshotHeight   int64  `json:"snapshot_height"`
	SnapshotUnix     int64  `json:"snapshot_unix"`
	ErasureEpoch     uint64 `json:"erasure_epoch"`
	Tombstone        bool   `json:"tombstone"`
	AuditRecords     bool   `json:"audit_records"`
	ObjectMetadata   bool   `json:"object_metadata"`
	Ciphertext       bool   `json:"ciphertext"`
	WrappedKeys      bool   `json:"wrapped_keys"`
	KeyReferences    bool   `json:"key_references"`
	Undecryptable    bool   `json:"undecryptable"`
}

type RestoreInventoryAuthority interface {
	VerifyRestoreInventory(context.Context, ErasureFence, TombstoneInventory) error
}

func ValidateRestoreAgainstErasure(ctx context.Context, authority RestoreInventoryAuthority, fence ErasureFence, inventory TombstoneInventory) error {
	if authority == nil {
		return errors.New("restore inventory authority is required")
	}
	if fence.TargetCommitment == "" || inventory.TargetCommitment != fence.TargetCommitment || fence.ErasureEpoch == 0 ||
		!isSHA256(fence.BackupGenerationDigest) || inventory.BackupGenerationDigest != fence.BackupGenerationDigest ||
		!isSHA256(inventory.SnapshotManifestDigest) {
		return errors.New("restore inventory does not match erasure fence")
	}
	if err := authority.VerifyRestoreInventory(ctx, fence, inventory); err != nil {
		return errors.Join(errors.New("restore inventory provenance is not authentic"), err)
	}
	if inventory.SnapshotHeight < fence.ErasedHeight || inventory.SnapshotUnix < fence.ErasedUnix || inventory.ErasureEpoch < fence.ErasureEpoch {
		return errors.New("pre-erasure snapshot restore is prohibited")
	}
	if inventory.ObjectMetadata || inventory.Ciphertext || inventory.WrappedKeys || inventory.KeyReferences {
		return errors.New("post-erasure restore may contain tombstone and audit records only")
	}
	if !inventory.Tombstone || !inventory.Undecryptable {
		return errors.New("post-erasure restore must retain tombstone and undecryptability")
	}
	return nil
}

func NewErasureFence(ref contracts.EvidenceObjectRef, erasureEpoch uint64, backupGenerationDigest string) (ErasureFence, error) {
	if ref.State != contracts.RetentionDeleted || erasureEpoch == 0 || !isSHA256(backupGenerationDigest) {
		return ErasureFence{}, errors.New("resolved deletion and erasure epoch are required")
	}
	return ErasureFence{TargetCommitment: ref.ObjectCommitment, BackupGenerationDigest: backupGenerationDigest, ErasureEpoch: erasureEpoch, ErasedHeight: ref.UpdatedHeight, ErasedUnix: ref.UpdatedUnix}, nil
}
