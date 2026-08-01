package data_vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/artifact_store"
	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
)

const fixtureArtifactIndexVersion uint32 = 1

var (
	ErrStaleArtifactRevision = errors.New("stale artifact index revision")
	fixtureArtifactLocks     sync.Map
)

type fixtureArtifactIndex struct {
	Artifacts      map[string]*artifact_store.ArtifactReference `json:"artifacts"`
	BlobMetadata   map[BlobID]*BlobMetadata                     `json:"blob_metadata"`
	LegalHolds     map[string]contracts.LegalHoldAuthority      `json:"legal_holds"`
	ErasureIntents map[string]*FixtureErasureTombstone          `json:"erasure_intents"`
	Mutations      map[string]*fixtureArtifactMutation          `json:"mutations"`
}

type fixtureArtifactMutation struct {
	ID        string                            `json:"id"`
	Operation string                            `json:"operation"`
	RefID     string                            `json:"ref_id"`
	BlobID    BlobID                            `json:"blob_id"`
	Owner     string                            `json:"owner"`
	Data      []byte                            `json:"data,omitempty"`
	Reference *artifact_store.ArtifactReference `json:"reference,omitempty"`
	Metadata  *BlobMetadata                     `json:"metadata,omitempty"`
}

type fixtureArtifactIndexEnvelope struct {
	Version   uint32          `json:"version"`
	Namespace string          `json:"namespace"`
	Revision  uint64          `json:"revision"`
	Payload   json.RawMessage `json:"payload"`
	Checksum  string          `json:"checksum"`
}

// FixtureFileArtifactStore is a fixture-only durable ArtifactStore.
// It is intentionally not production-ready and rejects the production profile.
type FixtureFileArtifactStore struct {
	root          string
	index         fixtureArtifactIndex
	revision      uint64
	anchor        RevisionAnchor
	namespace     string
	lockFile      *os.File
	mu            sync.RWMutex
	persistCalls  int
	failPersistAt int
}

// NewFixtureFileArtifactStore opens or creates a fixture filesystem backend.
func NewFixtureFileArtifactStore(root, profile string, anchors ...RevisionAnchor) (*FixtureFileArtifactStore, error) {
	return newFixtureFileArtifactStore(root, profile, FixtureSecurityOptions{}, anchors...)
}

// NewFixtureFileArtifactStoreWithSecurity allows an explicit fixture-only Windows ACL override.
func NewFixtureFileArtifactStoreWithSecurity(root, profile string, options FixtureSecurityOptions, anchors ...RevisionAnchor) (*FixtureFileArtifactStore, error) {
	return newFixtureFileArtifactStore(root, profile, options, anchors...)
}

func newFixtureFileArtifactStore(root, profile string, options FixtureSecurityOptions, anchors ...RevisionAnchor) (*FixtureFileArtifactStore, error) {
	if profile == "production" {
		return nil, errors.New("fixture artifact store is forbidden in production")
	}
	if profile != "fixture" && profile != "development" {
		return nil, errors.New("fixture artifact store requires fixture or development profile")
	}
	if root == "" {
		return nil, errors.New("fixture artifact root is required")
	}
	if len(anchors) != 1 || anchors[0] == nil || anchors[0].Replayable() {
		return nil, errors.New("fixture artifact store requires one non-replayable revision anchor")
	}
	cleanRoot := filepath.Clean(root)
	abs, err := filepath.Abs(cleanRoot)
	if err != nil {
		return nil, err
	}
	if err := rejectFixtureSymlink(cleanRoot); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cleanRoot, 0o700); err != nil {
		return nil, err
	}
	if err := enforceFixturePathSecurity(cleanRoot, true, options); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(cleanRoot, ".fixture-artifact.lock")
	if err := rejectFixtureSymlink(lockPath); err != nil {
		return nil, err
	}
	lockFile, err := tryLockFile(lockPath)
	if err != nil {
		return nil, err
	}
	if err := enforceFixturePathSecurity(lockPath, false, options); err != nil {
		_ = unlockFile(lockFile)
		return nil, err
	}
	store := &FixtureFileArtifactStore{
		root: cleanRoot, anchor: anchors[0], namespace: "vault-artifacts:" + filepath.ToSlash(abs), lockFile: lockFile,
	}
	if err := os.MkdirAll(store.objectsDir(), 0o700); err != nil {
		_ = store.Close()
		return nil, err
	}
	if err := enforceFixturePathSecurity(store.objectsDir(), true, options); err != nil {
		_ = store.Close()
		return nil, err
	}
	if err := store.reload(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			_ = store.Close()
			return nil, err
		}
		store.index = newFixtureArtifactIndex()
	}
	if err := enforceFixturePathSecurity(store.indexPath(), false, options); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = store.Close()
		return nil, err
	}
	for refID := range store.index.Artifacts {
		if _, err := safeFixtureRef(refID); err != nil {
			_ = store.Close()
			return nil, err
		}
		if err := enforceFixturePathSecurity(store.objectPath(refID), false, options); err != nil {
			_ = store.Close()
			return nil, err
		}
	}
	anchored, err := store.anchor.Current(store.namespace)
	if err != nil || anchored != store.revision {
		_ = store.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%w: artifact index has %d, anchor has %d", ErrRevisionRollback, store.revision, anchored)
	}
	if err := store.recoverMutations(); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("recover artifact mutations: %w", err)
	}
	if err := store.recoverErasureStorage(); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("recover erasure storage: %w", err)
	}
	return store, nil
}

// Close releases the fixture adapter's cross-process lease.
func (s *FixtureFileArtifactStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := unlockFile(s.lockFile)
	s.lockFile = nil
	return err
}

func newFixtureArtifactIndex() fixtureArtifactIndex {
	return fixtureArtifactIndex{
		Artifacts:      make(map[string]*artifact_store.ArtifactReference),
		BlobMetadata:   make(map[BlobID]*BlobMetadata),
		LegalHolds:     make(map[string]contracts.LegalHoldAuthority),
		ErasureIntents: make(map[string]*FixtureErasureTombstone),
		Mutations:      make(map[string]*fixtureArtifactMutation),
	}
}

func (s *FixtureFileArtifactStore) PutVaultBlob(_ context.Context, req *artifact_store.PutRequest, blobID BlobID, metadataFactory func(*artifact_store.PutResponse) *BlobMetadata) (*artifact_store.PutResponse, error) {
	if req == nil || blobID == "" || metadataFactory == nil {
		return nil, artifact_store.ErrInvalidInput.Wrap("transactional put inputs required")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	digest := sha256.Sum256(req.Data)
	refID := hex.EncodeToString(digest[:])
	address := artifact_store.NewContentAddressFromHash(append([]byte(nil), digest[:]...), uint64(len(req.Data)), artifact_store.BackendIPFS, refID)
	reference := artifact_store.NewArtifactReference(refID, address, cloneEncryptionMetadata(req.EncryptionMetadata), req.Owner, req.ArtifactType, 0)
	reference.RetentionTag = cloneRetentionTag(req.RetentionTag)
	reference.Metadata = cloneStringMap(req.Metadata)
	response := &artifact_store.PutResponse{ContentAddress: cloneContentAddress(address), ArtifactReference: cloneArtifactReference(reference)}
	metadata := metadataFactory(response)
	if metadata == nil || metadata.ID != blobID || metadata.BackendRef != refID {
		return nil, errors.New("transactional blob metadata does not match artifact")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return nil, err
	}
	if previous := s.index.Artifacts[refID]; previous != nil && previous.AccountAddress != req.Owner {
		return nil, artifact_store.ErrInvalidInput.Wrap("identical content is already owned by another account")
	}
	mutationID := "put:" + string(blobID)
	s.index.Mutations[mutationID] = &fixtureArtifactMutation{
		ID: mutationID, Operation: "put", RefID: refID, BlobID: blobID, Owner: req.Owner,
		Data: append([]byte(nil), req.Data...), Reference: cloneArtifactReference(reference), Metadata: cloneBlobMetadataValue(metadata),
	}
	if err := s.persist(); err != nil {
		delete(s.index.Mutations, mutationID)
		return nil, err
	}
	if err := writeExclusiveOrVerify(s.objectPath(refID), req.Data); err != nil {
		delete(s.index.Mutations, mutationID)
		if persistErr := s.persist(); persistErr != nil {
			return nil, errors.Join(err, persistErr)
		}
		return nil, err
	}
	s.index.Artifacts[refID] = reference
	s.index.BlobMetadata[blobID] = cloneBlobMetadataValue(metadata)
	delete(s.index.Mutations, mutationID)
	if err := s.persist(); err != nil {
		delete(s.index.Artifacts, refID)
		delete(s.index.BlobMetadata, blobID)
		return response, &ReconciliationRequiredError{OperationID: mutationID, Operation: "store", Cause: err}
	}
	return response, nil
}

func (s *FixtureFileArtifactStore) DeleteVaultBlob(_ context.Context, req *artifact_store.DeleteRequest, blobID BlobID) error {
	if req == nil || blobID == "" {
		return artifact_store.ErrInvalidInput.Wrap("transactional delete inputs required")
	}
	if err := req.Validate(); err != nil {
		return err
	}
	refID, err := safeFixtureRef(req.ContentAddress.BackendRef)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return err
	}
	reference := s.index.Artifacts[refID]
	if reference == nil {
		return artifact_store.ErrArtifactNotFound
	}
	if reference.AccountAddress != req.RequestingAccount {
		return artifact_store.ErrInvalidInput.Wrap("artifact owner mismatch")
	}
	if holdActive(s.index.LegalHolds, refID) {
		return errors.New("active legal hold blocks deletion")
	}
	mutationID := "delete:" + string(blobID)
	s.index.Mutations[mutationID] = &fixtureArtifactMutation{ID: mutationID, Operation: "delete", RefID: refID, BlobID: blobID, Owner: req.RequestingAccount}
	if err := s.persist(); err != nil {
		delete(s.index.Mutations, mutationID)
		return err
	}
	if err := os.Remove(s.objectPath(refID)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.applyDeleteMutation(s.index.Mutations[mutationID])
	delete(s.index.Mutations, mutationID)
	if err := s.persist(); err != nil {
		delete(s.index.Artifacts, refID)
		delete(s.index.BlobMetadata, blobID)
		delete(s.index.LegalHolds, refID)
		return &ReconciliationRequiredError{OperationID: mutationID, Operation: "delete", Cause: err}
	}
	return nil
}

func (s *FixtureFileArtifactStore) applyDeleteMutation(mutation *fixtureArtifactMutation) {
	delete(s.index.Artifacts, mutation.RefID)
	delete(s.index.BlobMetadata, mutation.BlobID)
	delete(s.index.LegalHolds, mutation.RefID)
}

func (s *FixtureFileArtifactStore) recoverMutations() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.index.Mutations) == 0 {
		return nil
	}
	ids := make([]string, 0, len(s.index.Mutations))
	for id := range s.index.Mutations {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		mutation := s.index.Mutations[id]
		if mutation == nil {
			return errors.New("nil artifact mutation")
		}
		if _, err := safeFixtureRef(mutation.RefID); err != nil {
			return err
		}
		switch mutation.Operation {
		case "put":
			if err := writeExclusiveOrVerify(s.objectPath(mutation.RefID), mutation.Data); err != nil {
				return err
			}
			s.index.Artifacts[mutation.RefID] = cloneArtifactReference(mutation.Reference)
			s.index.BlobMetadata[mutation.BlobID] = cloneBlobMetadataValue(mutation.Metadata)
		case "delete":
			if err := os.Remove(s.objectPath(mutation.RefID)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			s.applyDeleteMutation(mutation)
		default:
			return fmt.Errorf("unknown artifact mutation %q", mutation.Operation)
		}
		delete(s.index.Mutations, id)
	}
	return s.persist()
}

func (s *FixtureFileArtifactStore) recoverErasureStorage() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed := false
	for _, tombstone := range s.index.ErasureIntents {
		if tombstone == nil || tombstone.State != FixtureErasureIntent {
			continue
		}
		for _, target := range tombstone.Targets {
			if _, err := safeFixtureRef(target.BackendRef); err != nil {
				return err
			}
			if holdActive(s.index.LegalHolds, target.BackendRef) {
				return fmt.Errorf("active legal hold blocks erasure of blob %s", target.BlobID)
			}
			if err := os.Remove(s.objectPath(target.BackendRef)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			delete(s.index.Artifacts, target.BackendRef)
			delete(s.index.LegalHolds, target.BackendRef)
			delete(s.index.BlobMetadata, target.BlobID)
		}
		tombstone.StorageReceipt = fixtureStorageReceipt(tombstone.ID, tombstone.Targets)
		tombstone.State = FixtureErasureStorageDeleted
		tombstone.UpdatedAt = time.Now().UTC()
		changed = true
	}
	if !changed {
		return nil
	}
	return s.persist()
}

func (s *FixtureFileArtifactStore) Put(_ context.Context, req *artifact_store.PutRequest) (*artifact_store.PutResponse, error) {
	if req == nil {
		return nil, artifact_store.ErrInvalidInput.Wrap("request required")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	digest := sha256.Sum256(req.Data)
	refID := hex.EncodeToString(digest[:])
	address := artifact_store.NewContentAddressFromHash(append([]byte(nil), digest[:]...), uint64(len(req.Data)), artifact_store.BackendIPFS, refID)
	reference := artifact_store.NewArtifactReference(refID, address, cloneEncryptionMetadata(req.EncryptionMetadata), req.Owner, req.ArtifactType, 0)
	reference.RetentionTag = cloneRetentionTag(req.RetentionTag)
	reference.Metadata = cloneStringMap(req.Metadata)

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return nil, err
	}
	previous := cloneArtifactReference(s.index.Artifacts[refID])
	if previous != nil && previous.AccountAddress != req.Owner {
		return nil, artifact_store.ErrInvalidInput.Wrap("identical content is already owned by another account")
	}
	if err := writeExclusiveOrVerify(s.objectPath(refID), req.Data); err != nil {
		return nil, err
	}
	s.index.Artifacts[refID] = reference
	if err := s.persist(); err != nil {
		if previous == nil {
			delete(s.index.Artifacts, refID)
		} else {
			s.index.Artifacts[refID] = previous
		}
		return nil, err
	}
	return &artifact_store.PutResponse{ContentAddress: cloneContentAddress(address), ArtifactReference: cloneArtifactReference(reference)}, nil
}

func (s *FixtureFileArtifactStore) Get(_ context.Context, req *artifact_store.GetRequest) (*artifact_store.GetResponse, error) {
	if req == nil {
		return nil, artifact_store.ErrInvalidInput.Wrap("request required")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	refID, err := safeFixtureRef(req.ContentAddress.BackendRef)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	reference := s.index.Artifacts[refID]
	s.mu.RUnlock()
	if reference == nil {
		return nil, artifact_store.ErrArtifactNotFound
	}
	if req.RequestingAccount == "" || req.RequestingAccount != reference.AccountAddress {
		return nil, artifact_store.ErrInvalidInput.Wrap("artifact owner mismatch")
	}
	data, err := os.ReadFile(s.objectPath(refID))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, artifact_store.ErrArtifactNotFound
		}
		return nil, err
	}
	digest := sha256.Sum256(data)
	if !equalBytes(digest[:], req.ContentAddress.Hash) || hex.EncodeToString(digest[:]) != refID {
		return nil, artifact_store.ErrHashMismatch.Wrap("fixture artifact checksum mismatch")
	}
	return &artifact_store.GetResponse{Data: append([]byte(nil), data...), ContentAddress: cloneContentAddress(req.ContentAddress)}, nil
}

func (s *FixtureFileArtifactStore) Delete(_ context.Context, req *artifact_store.DeleteRequest) error {
	if req == nil {
		return artifact_store.ErrInvalidInput.Wrap("request required")
	}
	if err := req.Validate(); err != nil {
		return err
	}
	refID, err := safeFixtureRef(req.ContentAddress.BackendRef)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return err
	}
	reference := s.index.Artifacts[refID]
	if reference == nil {
		return artifact_store.ErrArtifactNotFound
	}
	if reference.AccountAddress != req.RequestingAccount {
		return artifact_store.ErrInvalidInput.Wrap("artifact owner mismatch")
	}
	if hold, ok := s.index.LegalHolds[refID]; ok && hold.State == contracts.HoldActive {
		return errors.New("active legal hold blocks deletion")
	}
	previousReference := cloneArtifactReference(reference)
	previousHold, hadHold := s.index.LegalHolds[refID]
	delete(s.index.Artifacts, refID)
	delete(s.index.LegalHolds, refID)
	if err := s.persist(); err != nil {
		s.index.Artifacts[refID] = previousReference
		if hadHold {
			s.index.LegalHolds[refID] = previousHold
		}
		return err
	}
	if err := os.Remove(s.objectPath(refID)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// SetLegalHold validates and durably records a hold for an artifact.
func (s *FixtureFileArtifactStore) SetLegalHold(address *artifact_store.ContentAddress, hold contracts.LegalHoldAuthority, verifier contracts.HoldAuthorityVerifier) error {
	if address == nil {
		return errors.New("content address required")
	}
	if err := contracts.ValidateHold(hold, verifier); err != nil {
		return err
	}
	refID, err := safeFixtureRef(address.BackendRef)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return err
	}
	if s.index.Artifacts[refID] == nil {
		return artifact_store.ErrArtifactNotFound
	}
	previous, existed := s.index.LegalHolds[refID]
	s.index.LegalHolds[refID] = hold
	if err := s.persist(); err != nil {
		if existed {
			s.index.LegalHolds[refID] = previous
		} else {
			delete(s.index.LegalHolds, refID)
		}
		return err
	}
	return nil
}

func (s *FixtureFileArtifactStore) Exists(_ context.Context, address *artifact_store.ContentAddress) (bool, error) {
	if address == nil {
		return false, artifact_store.ErrInvalidInput.Wrap("content address required")
	}
	refID, err := safeFixtureRef(address.BackendRef)
	if err != nil {
		return false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.index.Artifacts[refID]
	return ok, nil
}

func (s *FixtureFileArtifactStore) GetChunk(context.Context, *artifact_store.ContentAddress, uint32) (*artifact_store.ChunkData, error) {
	return nil, artifact_store.ErrBackendNotSupported
}

func (s *FixtureFileArtifactStore) ListByOwner(_ context.Context, owner string, pagination *artifact_store.Pagination) (*artifact_store.ListResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	refs := make([]*artifact_store.ArtifactReference, 0)
	for _, reference := range s.index.Artifacts {
		if reference.AccountAddress == owner {
			refs = append(refs, cloneArtifactReference(reference))
		}
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].ReferenceID < refs[j].ReferenceID })
	total := uint64(len(refs))
	offset, limit := uint64(0), total
	if pagination != nil {
		offset = pagination.Offset
		if pagination.Limit > 0 {
			limit = pagination.Limit
		}
	}
	if offset >= total {
		return &artifact_store.ListResponse{References: []*artifact_store.ArtifactReference{}, Total: total}, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return &artifact_store.ListResponse{References: refs[offset:end], Total: total, HasMore: end < total}, nil
}

func (s *FixtureFileArtifactStore) UpdateRetention(_ context.Context, address *artifact_store.ContentAddress, tag *artifact_store.RetentionTag) error {
	if address == nil {
		return artifact_store.ErrInvalidInput.Wrap("content address required")
	}
	refID, err := safeFixtureRef(address.BackendRef)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return err
	}
	if s.index.Artifacts[refID] == nil {
		return artifact_store.ErrArtifactNotFound
	}
	previous := cloneRetentionTag(s.index.Artifacts[refID].RetentionTag)
	s.index.Artifacts[refID].RetentionTag = cloneRetentionTag(tag)
	if err := s.persist(); err != nil {
		s.index.Artifacts[refID].RetentionTag = previous
		return err
	}
	return nil
}

func (s *FixtureFileArtifactStore) PurgeExpired(_ context.Context, currentBlock int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return 0, err
	}
	removed := make([]string, 0)
	for refID, reference := range s.index.Artifacts {
		tag := reference.RetentionTag
		if tag == nil || (holdActive(s.index.LegalHolds, refID)) {
			continue
		}
		if (tag.ExpiresAt != nil && time.Now().UTC().After(*tag.ExpiresAt)) || (tag.ExpiresAtBlock != nil && currentBlock >= *tag.ExpiresAtBlock) {
			delete(s.index.Artifacts, refID)
			removed = append(removed, refID)
		}
	}
	if len(removed) == 0 {
		return 0, nil
	}
	if err := s.persist(); err != nil {
		return 0, err
	}
	for _, refID := range removed {
		_ = os.Remove(s.objectPath(refID))
	}
	return len(removed), nil
}

func (s *FixtureFileArtifactStore) GetMetrics(context.Context) (*artifact_store.StorageMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var totalBytes uint64
	for _, reference := range s.index.Artifacts {
		totalBytes += reference.ContentAddress.Size
	}
	return &artifact_store.StorageMetrics{
		TotalArtifacts: uint64(len(s.index.Artifacts)), TotalBytes: totalBytes,
		BackendType:   artifact_store.BackendIPFS,
		BackendStatus: map[string]string{"mode": "fixture-only", "production_ready": "false"},
	}, nil
}

func (s *FixtureFileArtifactStore) Health(context.Context) error { return s.reloadReadOnly() }
func (s *FixtureFileArtifactStore) Backend() artifact_store.BackendType {
	return artifact_store.BackendIPFS
}

func (s *FixtureFileArtifactStore) LoadVaultMetadata() (map[BlobID]*BlobMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneBlobMetadata(s.index.BlobMetadata), nil
}

func (s *FixtureFileArtifactStore) SaveVaultMetadata(metadata map[BlobID]*BlobMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return err
	}
	previous := s.index.BlobMetadata
	s.index.BlobMetadata = cloneBlobMetadata(metadata)
	if err := s.persist(); err != nil {
		s.index.BlobMetadata = previous
		return err
	}
	return nil
}

func (s *FixtureFileArtifactStore) reloadReadOnly() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reload()
}

func (s *FixtureFileArtifactStore) ensureCurrent() error {
	revision, _, err := readFixtureArtifactIndex(s.indexPath())
	if errors.Is(err, os.ErrNotExist) && s.revision == 0 {
		return nil
	}
	if err != nil {
		return err
	}
	if revision != s.revision {
		return fmt.Errorf("%w: have %d, expected %d", ErrStaleArtifactRevision, revision, s.revision)
	}
	return nil
}

func (s *FixtureFileArtifactStore) reload() error {
	revision, index, err := readFixtureArtifactIndex(s.indexPath())
	if err != nil {
		return err
	}
	encoded, err := os.ReadFile(s.indexPath())
	if err != nil {
		return err
	}
	var envelope fixtureArtifactIndexEnvelope
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return err
	}
	if envelope.Namespace != s.namespace {
		return errors.New("fixture artifact namespace mismatch")
	}
	s.revision, s.index = revision, index
	return nil
}

func (s *FixtureFileArtifactStore) persist() error {
	s.persistCalls++
	if s.failPersistAt > 0 && s.persistCalls == s.failPersistAt {
		return errors.New("injected fixture artifact persistence failure")
	}
	pathLock := fixtureArtifactPathLock(s.indexPath())
	pathLock.Lock()
	defer pathLock.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return err
	}
	payload, err := json.Marshal(s.index)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(payload)
	envelope := fixtureArtifactIndexEnvelope{
		Version: fixtureArtifactIndexVersion, Namespace: s.namespace, Revision: s.revision + 1,
		Payload: payload, Checksum: hex.EncodeToString(digest[:]),
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	if err := atomicWriteFixtureFile(s.indexPath(), encoded, 0o600); err != nil {
		return err
	}
	if err := s.anchor.CompareAndAdvance(s.namespace, s.revision, s.revision+1); err != nil {
		return fmt.Errorf("advance artifact revision anchor: %w", err)
	}
	s.revision++
	return nil
}

func fixtureArtifactPathLock(path string) *sync.Mutex {
	lock, _ := fixtureArtifactLocks.LoadOrStore(path, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func readFixtureArtifactIndex(path string) (uint64, fixtureArtifactIndex, error) {
	encoded, err := os.ReadFile(path)
	if err != nil {
		return 0, fixtureArtifactIndex{}, err
	}
	var envelope fixtureArtifactIndexEnvelope
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return 0, fixtureArtifactIndex{}, fmt.Errorf("decode fixture artifact index: %w", err)
	}
	digest := sha256.Sum256(envelope.Payload)
	if envelope.Version != fixtureArtifactIndexVersion || envelope.Revision == 0 || hex.EncodeToString(digest[:]) != envelope.Checksum {
		return 0, fixtureArtifactIndex{}, errors.New("fixture artifact index checksum or version invalid")
	}
	var index fixtureArtifactIndex
	if err := json.Unmarshal(envelope.Payload, &index); err != nil {
		return 0, fixtureArtifactIndex{}, err
	}
	if index.Artifacts == nil || index.BlobMetadata == nil || index.LegalHolds == nil {
		return 0, fixtureArtifactIndex{}, errors.New("fixture artifact index is incomplete")
	}
	if index.ErasureIntents == nil {
		index.ErasureIntents = make(map[string]*FixtureErasureTombstone)
	}
	if index.Mutations == nil {
		index.Mutations = make(map[string]*fixtureArtifactMutation)
	}
	return envelope.Revision, index, nil
}

// Revision returns the durable fixture index revision.
func (s *FixtureFileArtifactStore) Revision() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.revision
}

func (s *FixtureFileArtifactStore) indexPath() string  { return filepath.Join(s.root, "index.json") }
func (s *FixtureFileArtifactStore) objectsDir() string { return filepath.Join(s.root, "objects") }
func (s *FixtureFileArtifactStore) objectPath(refID string) string {
	return filepath.Join(s.objectsDir(), refID)
}

func safeFixtureRef(value string) (string, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size || value != hex.EncodeToString(decoded) {
		return "", artifact_store.ErrInvalidInput.Wrap("invalid fixture content reference")
	}
	return value, nil
}

func writeExclusiveOrVerify(path string, data []byte) error {
	existing, err := os.ReadFile(path)
	if err == nil {
		if !equalBytes(existing, data) {
			return errors.New("content address collision")
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return atomicWriteFixtureFile(path, data, 0o600)
}

func atomicWriteFixtureFile(path string, data []byte, mode os.FileMode) error {
	if err := rejectFixtureSymlink(path); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	if err := rejectFixtureSymlink(filepath.Dir(path)); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".artifact-*")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if err := temp.Chmod(mode); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, path); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return nil
	}
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func cloneBlobMetadata(source map[BlobID]*BlobMetadata) map[BlobID]*BlobMetadata {
	result := make(map[BlobID]*BlobMetadata, len(source))
	for id, metadata := range source {
		if metadata == nil {
			result[id] = nil
			continue
		}
		copy := *metadata
		copy.ContentHash = append([]byte(nil), metadata.ContentHash...)
		copy.ContentAddressHash = append([]byte(nil), metadata.ContentAddressHash...)
		copy.Tags = cloneStringMap(metadata.Tags)
		if metadata.ExpiresAt != nil {
			expiresAt := *metadata.ExpiresAt
			copy.ExpiresAt = &expiresAt
		}
		result[id] = &copy
	}
	return result
}

func cloneArtifactReference(reference *artifact_store.ArtifactReference) *artifact_store.ArtifactReference {
	if reference == nil {
		return nil
	}
	cloned := *reference
	cloned.ContentAddress = cloneContentAddress(reference.ContentAddress)
	cloned.EncryptionMetadata = cloneEncryptionMetadata(reference.EncryptionMetadata)
	cloned.RetentionTag = cloneRetentionTag(reference.RetentionTag)
	cloned.Metadata = cloneStringMap(reference.Metadata)
	if reference.ChunkManifest != nil {
		manifest := *reference.ChunkManifest
		manifest.RootHash = append([]byte(nil), reference.ChunkManifest.RootHash...)
		manifest.Chunks = append([]artifact_store.ChunkInfo(nil), reference.ChunkManifest.Chunks...)
		for index := range manifest.Chunks {
			manifest.Chunks[index].Hash = append([]byte(nil), manifest.Chunks[index].Hash...)
		}
		cloned.ChunkManifest = &manifest
	}
	return &cloned
}

func cloneContentAddress(address *artifact_store.ContentAddress) *artifact_store.ContentAddress {
	if address == nil {
		return nil
	}
	cloned := *address
	cloned.Hash = append([]byte(nil), address.Hash...)
	return &cloned
}

func cloneEncryptionMetadata(metadata *artifact_store.EncryptionMetadata) *artifact_store.EncryptionMetadata {
	if metadata == nil {
		return nil
	}
	cloned := *metadata
	cloned.RecipientKeyIDs = append([]string(nil), metadata.RecipientKeyIDs...)
	cloned.EnvelopeHash = append([]byte(nil), metadata.EnvelopeHash...)
	return &cloned
}

func cloneRetentionTag(tag *artifact_store.RetentionTag) *artifact_store.RetentionTag {
	if tag == nil {
		return nil
	}
	cloned := *tag
	if tag.ExpiresAt != nil {
		expiresAt := *tag.ExpiresAt
		cloned.ExpiresAt = &expiresAt
	}
	if tag.ExpiresAtBlock != nil {
		expiresAtBlock := *tag.ExpiresAtBlock
		cloned.ExpiresAtBlock = &expiresAtBlock
	}
	return &cloned
}

func holdActive(holds map[string]contracts.LegalHoldAuthority, refID string) bool {
	hold, ok := holds[refID]
	return ok && hold.State == contracts.HoldActive
}

func equalBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
