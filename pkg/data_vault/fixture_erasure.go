package data_vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

// FixtureErasureState is a durable fixture-only erasure checkpoint.
type FixtureErasureState string

const (
	FixtureErasureIntent         FixtureErasureState = "intent"
	FixtureErasureStorageDeleted FixtureErasureState = "storage_deleted"
	FixtureErasureKeyDestroyed   FixtureErasureState = "key_destroyed"
	FixtureErasureComplete       FixtureErasureState = "complete"
)

// FixtureErasureAuthorization binds a fixture authorization object to exact durable state.
// Production must replace this digest fixture with authenticated policy authorization.
type FixtureErasureAuthorization struct {
	Digest           string   `json:"digest"`
	ObjectRef        string   `json:"object_ref"`
	BlobIDs          []BlobID `json:"blob_ids"`
	Scope            Scope    `json:"scope"`
	KeyID            string   `json:"key_id"`
	ArtifactRevision uint64   `json:"artifact_revision"`
	KeyRevision      uint64   `json:"key_revision"`
}

type FixtureErasureTarget struct {
	BlobID     BlobID `json:"blob_id"`
	BackendRef string `json:"backend_ref"`
	Owner      string `json:"owner"`
}

// FixtureErasureTombstone persists every irreversible transition for crash resume.
type FixtureErasureTombstone struct {
	ID                  string                       `json:"id"`
	AuthorizationDigest string                       `json:"authorization_digest"`
	AuthorizationRef    string                       `json:"authorization_ref"`
	Scope               Scope                        `json:"scope"`
	KeyID               string                       `json:"key_id"`
	ArtifactRevision    uint64                       `json:"artifact_revision"`
	KeyRevision         uint64                       `json:"key_revision"`
	Targets             []FixtureErasureTarget       `json:"targets"`
	State               FixtureErasureState          `json:"state"`
	StorageReceipt      string                       `json:"storage_receipt,omitempty"`
	KeyReceipt          contracts.DestructionReceipt `json:"key_receipt"`
	UpdatedAt           time.Time                    `json:"updated_at"`
}

type fixtureDestructiveKeyManager interface {
	GetKey(scope keys.Scope, keyID string) (*keys.KeyInfo, error)
	DestroyKey(scope keys.Scope, keyID string) (string, error)
	Revision() uint64
}

// FixtureErasureAuthorizationVerifier authenticates the external authorization object.
// Fixture implementations must fail closed; production requires its policy authority instead.
type FixtureErasureAuthorizationVerifier interface {
	VerifyFixtureErasureAuthorization(context.Context, FixtureErasureAuthorization) error
}

// FixtureErasureCoordinator owns hold checks, artifact deletion, and key destruction.
// It is fixture-only; production requires a transactional policy service and non-exportable KMS operations.
type FixtureErasureCoordinator struct {
	artifacts *FixtureFileArtifactStore
	keys      fixtureDestructiveKeyManager
	verifier  FixtureErasureAuthorizationVerifier
}

func NewFixtureErasureCoordinator(artifacts *FixtureFileArtifactStore, keyManager fixtureDestructiveKeyManager, verifier FixtureErasureAuthorizationVerifier) (*FixtureErasureCoordinator, error) {
	if artifacts == nil || keyManager == nil || verifier == nil {
		return nil, errors.New("fixture artifact store, persistent key manager, and authorization verifier are required")
	}
	coordinator := &FixtureErasureCoordinator{artifacts: artifacts, keys: keyManager, verifier: verifier}
	if err := coordinator.ResumePending(context.Background()); err != nil {
		return nil, fmt.Errorf("resume fixture erasure: %w", err)
	}
	return coordinator, nil
}

// ResumePending completes already-authorized durable erasure intents.
func (c *FixtureErasureCoordinator) ResumePending(_ context.Context) error {
	c.artifacts.mu.RLock()
	ids := make([]string, 0, len(c.artifacts.index.ErasureIntents))
	for id, tombstone := range c.artifacts.index.ErasureIntents {
		if tombstone != nil && tombstone.State != FixtureErasureComplete {
			ids = append(ids, id)
		}
	}
	c.artifacts.mu.RUnlock()
	sort.Strings(ids)
	for _, id := range ids {
		if err := c.resumeStorageDeletion(id); err != nil {
			return err
		}
		if err := c.resumeKeyDestruction(id); err != nil {
			return err
		}
		if _, err := c.complete(id); err != nil {
			return err
		}
	}
	return nil
}

// PrepareAuthorization snapshots the exact shared-key impact set for an external fixture authorizer.
func (c *FixtureErasureCoordinator) PrepareAuthorization(scope Scope, keyID, objectRef string) (FixtureErasureAuthorization, error) {
	if scope == "" || keyID == "" || objectRef == "" {
		return FixtureErasureAuthorization{}, errors.New("scope, key id, and authorization object reference are required")
	}
	c.artifacts.mu.RLock()
	defer c.artifacts.mu.RUnlock()
	blobIDs := make([]BlobID, 0)
	for blobID, metadata := range c.artifacts.index.BlobMetadata {
		if metadata != nil && metadata.Scope == scope && metadata.KeyID == keyID {
			blobIDs = append(blobIDs, blobID)
		}
	}
	sort.Slice(blobIDs, func(i, j int) bool { return blobIDs[i] < blobIDs[j] })
	if len(blobIDs) == 0 {
		return FixtureErasureAuthorization{}, errors.New("no durable blobs reference the key")
	}
	authorization := FixtureErasureAuthorization{
		ObjectRef: objectRef, BlobIDs: blobIDs, Scope: scope, KeyID: keyID,
		ArtifactRevision: c.artifacts.revision, KeyRevision: c.keys.Revision(),
	}
	authorization.Digest = fixtureAuthorizationDigest(authorization)
	return authorization, nil
}

// Erase validates durable holds and resumes the bound erasure intent to completion.
func (c *FixtureErasureCoordinator) Erase(ctx context.Context, authorization FixtureErasureAuthorization) (contracts.DestructionReceipt, error) {
	_ = ctx
	if authorization.Digest == "" || authorization.Digest != fixtureAuthorizationDigest(authorization) {
		return contracts.DestructionReceipt{}, errors.New("fixture erasure authorization digest mismatch")
	}
	if err := c.verifier.VerifyFixtureErasureAuthorization(ctx, authorization); err != nil {
		return contracts.DestructionReceipt{}, fmt.Errorf("verify fixture erasure authorization: %w", err)
	}
	tombstoneID := authorization.Digest

	c.artifacts.mu.Lock()
	if err := c.artifacts.ensureCurrent(); err != nil {
		c.artifacts.mu.Unlock()
		return contracts.DestructionReceipt{}, err
	}
	tombstone := c.artifacts.index.ErasureIntents[tombstoneID]
	if tombstone == nil {
		if c.artifacts.revision != authorization.ArtifactRevision || c.keys.Revision() != authorization.KeyRevision {
			c.artifacts.mu.Unlock()
			return contracts.DestructionReceipt{}, errors.New("fixture erasure authorization revision is stale")
		}
		targets, err := c.validateTargetsLocked(authorization)
		if err != nil {
			c.artifacts.mu.Unlock()
			return contracts.DestructionReceipt{}, err
		}
		tombstone = &FixtureErasureTombstone{
			ID: tombstoneID, AuthorizationDigest: authorization.Digest, AuthorizationRef: authorization.ObjectRef,
			Scope: authorization.Scope, KeyID: authorization.KeyID, ArtifactRevision: authorization.ArtifactRevision,
			KeyRevision: authorization.KeyRevision, Targets: targets, State: FixtureErasureIntent, UpdatedAt: time.Now().UTC(),
		}
		c.artifacts.index.ErasureIntents[tombstoneID] = tombstone
		if err := c.artifacts.persist(); err != nil {
			delete(c.artifacts.index.ErasureIntents, tombstoneID)
			c.artifacts.mu.Unlock()
			return contracts.DestructionReceipt{}, err
		}
	}
	c.artifacts.mu.Unlock()

	if err := c.resumeStorageDeletion(tombstoneID); err != nil {
		return contracts.DestructionReceipt{}, err
	}
	if err := c.resumeKeyDestruction(tombstoneID); err != nil {
		return contracts.DestructionReceipt{}, err
	}
	return c.complete(tombstoneID)
}

func (c *FixtureErasureCoordinator) validateTargetsLocked(authorization FixtureErasureAuthorization) ([]FixtureErasureTarget, error) {
	expected := make(map[BlobID]struct{}, len(authorization.BlobIDs))
	for _, blobID := range authorization.BlobIDs {
		if blobID == "" {
			return nil, errors.New("fixture erasure authorization contains an empty blob id")
		}
		expected[blobID] = struct{}{}
	}
	targets := make([]FixtureErasureTarget, 0, len(expected))
	for blobID, metadata := range c.artifacts.index.BlobMetadata {
		if metadata == nil || metadata.Scope != authorization.Scope || metadata.KeyID != authorization.KeyID {
			continue
		}
		if _, authorized := expected[blobID]; !authorized {
			return nil, fmt.Errorf("shared key destruction omitted affected blob %s", blobID)
		}
		if hold, held := c.artifacts.index.LegalHolds[metadata.BackendRef]; held && hold.State == contracts.HoldActive {
			return nil, fmt.Errorf("active legal hold blocks erasure of blob %s", blobID)
		}
		targets = append(targets, FixtureErasureTarget{BlobID: blobID, BackendRef: metadata.BackendRef, Owner: metadata.Owner})
		delete(expected, blobID)
	}
	if len(expected) != 0 || len(targets) == 0 {
		return nil, errors.New("authorized blob set does not exactly match durable key references")
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].BlobID < targets[j].BlobID })
	return targets, nil
}

func (c *FixtureErasureCoordinator) resumeStorageDeletion(tombstoneID string) error {
	c.artifacts.mu.Lock()
	defer c.artifacts.mu.Unlock()
	tombstone := c.artifacts.index.ErasureIntents[tombstoneID]
	if tombstone == nil {
		return errors.New("fixture erasure tombstone missing")
	}
	if tombstone.State != FixtureErasureIntent {
		return nil
	}
	for _, target := range tombstone.Targets {
		if _, err := safeFixtureRef(target.BackendRef); err != nil {
			return err
		}
		if hold, held := c.artifacts.index.LegalHolds[target.BackendRef]; held && hold.State == contracts.HoldActive {
			return fmt.Errorf("active legal hold blocks erasure of blob %s", target.BlobID)
		}
		if err := os.Remove(c.artifacts.objectPath(target.BackendRef)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		delete(c.artifacts.index.Artifacts, target.BackendRef)
		delete(c.artifacts.index.LegalHolds, target.BackendRef)
		delete(c.artifacts.index.BlobMetadata, target.BlobID)
	}
	tombstone.StorageReceipt = fixtureStorageReceipt(tombstone.ID, tombstone.Targets)
	tombstone.State = FixtureErasureStorageDeleted
	tombstone.UpdatedAt = time.Now().UTC()
	if err := c.artifacts.persist(); err != nil {
		tombstone.State = FixtureErasureIntent
		tombstone.StorageReceipt = ""
		return &ReconciliationRequiredError{OperationID: tombstoneID, Operation: "erasure-storage", Cause: err}
	}
	return nil
}

func (c *FixtureErasureCoordinator) resumeKeyDestruction(tombstoneID string) error {
	c.artifacts.mu.RLock()
	tombstone := cloneFixtureErasureTombstone(c.artifacts.index.ErasureIntents[tombstoneID])
	c.artifacts.mu.RUnlock()
	if tombstone == nil {
		return errors.New("fixture erasure tombstone missing")
	}
	if tombstone.State != FixtureErasureStorageDeleted {
		return nil
	}
	if tombstone.StorageReceipt == "" {
		return errors.New("fixture erasure storage deletion receipt is missing")
	}
	digest, err := c.keys.DestroyKey(keys.Scope(tombstone.Scope), tombstone.KeyID)
	if err != nil {
		if _, getErr := c.keys.GetKey(keys.Scope(tombstone.Scope), tombstone.KeyID); getErr == nil {
			return err
		}
		digestBytes := sha256.Sum256([]byte("fixture-erasure-resumed:" + tombstone.ID + ":" + tombstone.KeyID))
		digest = hex.EncodeToString(digestBytes[:])
	}
	c.artifacts.mu.Lock()
	defer c.artifacts.mu.Unlock()
	tombstone = c.artifacts.index.ErasureIntents[tombstoneID]
	tombstone.KeyReceipt = contracts.DestructionReceipt{TargetID: tombstone.KeyID, Digest: digest}
	tombstone.State = FixtureErasureKeyDestroyed
	tombstone.UpdatedAt = time.Now().UTC()
	if err := c.artifacts.persist(); err != nil {
		tombstone.KeyReceipt = contracts.DestructionReceipt{}
		tombstone.State = FixtureErasureStorageDeleted
		return err
	}
	return nil
}

func (c *FixtureErasureCoordinator) complete(tombstoneID string) (contracts.DestructionReceipt, error) {
	c.artifacts.mu.Lock()
	defer c.artifacts.mu.Unlock()
	tombstone := c.artifacts.index.ErasureIntents[tombstoneID]
	if tombstone == nil {
		return contracts.DestructionReceipt{}, errors.New("fixture erasure tombstone missing")
	}
	if tombstone.State == FixtureErasureKeyDestroyed {
		tombstone.State = FixtureErasureComplete
		tombstone.UpdatedAt = time.Now().UTC()
		if err := c.artifacts.persist(); err != nil {
			tombstone.State = FixtureErasureKeyDestroyed
			return contracts.DestructionReceipt{}, err
		}
	}
	if tombstone.State != FixtureErasureComplete {
		return contracts.DestructionReceipt{}, fmt.Errorf("fixture erasure stopped in state %s", tombstone.State)
	}
	return tombstone.KeyReceipt, nil
}

func fixtureAuthorizationDigest(authorization FixtureErasureAuthorization) string {
	copy := authorization
	copy.Digest = ""
	copy.BlobIDs = append([]BlobID(nil), authorization.BlobIDs...)
	sort.Slice(copy.BlobIDs, func(i, j int) bool { return copy.BlobIDs[i] < copy.BlobIDs[j] })
	encoded, _ := json.Marshal(copy)
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func fixtureStorageReceipt(tombstoneID string, targets []FixtureErasureTarget) string {
	receiptPayload, _ := json.Marshal(targets)
	receiptDigest := sha256.Sum256(append([]byte("fixture-storage-deleted:"+tombstoneID+":"), receiptPayload...))
	return hex.EncodeToString(receiptDigest[:])
}

func cloneFixtureErasureTombstone(tombstone *FixtureErasureTombstone) *FixtureErasureTombstone {
	if tombstone == nil {
		return nil
	}
	copy := *tombstone
	copy.Targets = append([]FixtureErasureTarget(nil), tombstone.Targets...)
	return &copy
}
