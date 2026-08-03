package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
)

const fixtureStateVersion uint32 = 1

var (
	ErrStateNotFound        = errors.New("key state not found")
	ErrStaleRevision        = errors.New("stale key state revision")
	errFixtureKeyStateInUse = errors.New("fixture key state is already open by another process or instance")
	fixtureStateLocks       sync.Map
)

// StatePersistence is backend-neutral durable custody for an opaque sealed key state.
type StatePersistence interface {
	Load() (revision uint64, state []byte, err error)
	CompareAndSwap(expectedRevision uint64, state []byte) (newRevision uint64, err error)
}

// FixtureFilePersistence is a fixture-only encrypted filesystem state adapter.
// It is not production-ready and rejects the production profile.
type FixtureFilePersistence struct {
	path        string
	wrappingKey [32]byte
	anchor      contracts.RevisionAnchor
	namespace   string
	lockFile    *os.File
	mu          sync.Mutex
}

type fixtureFileEnvelope struct {
	Version    uint32 `json:"version"`
	Namespace  string `json:"namespace"`
	Revision   uint64 `json:"revision"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
	Checksum   string `json:"checksum"`
}

// NewFixtureFilePersistence creates a fixture-only persistence adapter.
func NewFixtureFilePersistence(path string, wrappingKey []byte, profile string, anchors ...contracts.RevisionAnchor) (*FixtureFilePersistence, error) {
	return newFixtureFilePersistence(path, wrappingKey, profile, contracts.FixtureSecurityOptions{}, anchors...)
}

// NewFixtureFilePersistenceWithSecurity allows an explicit fixture-only Windows ACL override.
func NewFixtureFilePersistenceWithSecurity(path string, wrappingKey []byte, profile string, options contracts.FixtureSecurityOptions, anchors ...contracts.RevisionAnchor) (*FixtureFilePersistence, error) {
	return newFixtureFilePersistence(path, wrappingKey, profile, options, anchors...)
}

func newFixtureFilePersistence(path string, wrappingKey []byte, profile string, options contracts.FixtureSecurityOptions, anchors ...contracts.RevisionAnchor) (*FixtureFilePersistence, error) {
	if profile == "production" {
		return nil, errors.New("fixture key persistence is forbidden in production")
	}
	if profile != "fixture" && profile != "development" {
		return nil, errors.New("fixture key persistence requires fixture or development profile")
	}
	if path == "" || len(wrappingKey) != 32 {
		return nil, errors.New("key state path and 32-byte fixture wrapping key are required")
	}
	if len(anchors) != 1 || anchors[0] == nil || anchors[0].Replayable() {
		return nil, errors.New("fixture key persistence requires one non-replayable revision anchor")
	}
	clean := filepath.Clean(path)
	abs, err := filepath.Abs(clean)
	if err != nil {
		return nil, err
	}
	namespace := "vault-key-state:" + filepath.ToSlash(abs)
	if err := rejectSymlinkTarget(clean); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(clean), 0o700); err != nil {
		return nil, err
	}
	if err := enforceKeyPathSecurity(filepath.Dir(clean), true, options); err != nil {
		return nil, err
	}
	if err := rejectSymlinkTarget(filepath.Dir(clean)); err != nil {
		return nil, err
	}
	lockPath := clean + ".lock"
	if err := rejectSymlinkTarget(lockPath); err != nil {
		return nil, err
	}
	lockFile, err := tryLockFile(lockPath)
	if err != nil {
		return nil, err
	}
	if err := enforceKeyPathSecurity(lockPath, false, options); err != nil {
		_ = unlockFile(lockFile)
		return nil, err
	}
	var key [32]byte
	copy(key[:], wrappingKey)
	persistence := &FixtureFilePersistence{
		path: clean, wrappingKey: key, anchor: anchors[0], namespace: namespace, lockFile: lockFile,
	}
	if err := enforceKeyPathSecurity(clean, false, options); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = persistence.Close()
		return nil, err
	}
	if err := persistence.verifyAnchor(); err != nil {
		_ = persistence.Close()
		return nil, err
	}
	return persistence, nil
}

func (p *FixtureFilePersistence) verifyAnchor() error {
	anchored, err := p.anchor.Current(p.namespace)
	if err != nil {
		return err
	}
	revision := uint64(0)
	if _, err := os.Stat(p.path); err == nil {
		revision, _, err = p.loadFile()
		if err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if revision != anchored {
		return fmt.Errorf("%w: key state has %d, anchor has %d", contracts.ErrRevisionRollback, revision, anchored)
	}
	return nil
}

// Close releases the fixture adapter's cross-process lease.
func (p *FixtureFilePersistence) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	err := unlockFile(p.lockFile)
	p.lockFile = nil
	return err
}

func (p *FixtureFilePersistence) Load() (uint64, []byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.load()
}

func (p *FixtureFilePersistence) CompareAndSwap(expectedRevision uint64, state []byte) (uint64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	pathLock := fixtureStatePathLock(p.path)
	pathLock.Lock()
	defer pathLock.Unlock()

	currentRevision := uint64(0)
	if _, err := os.Stat(p.path); err == nil {
		revision, _, loadErr := p.load()
		if loadErr != nil {
			return 0, loadErr
		}
		currentRevision = revision
	} else if !errors.Is(err, os.ErrNotExist) {
		return 0, err
	}
	if currentRevision != expectedRevision {
		return 0, fmt.Errorf("%w: have %d, expected %d", ErrStaleRevision, currentRevision, expectedRevision)
	}

	newRevision := currentRevision + 1
	envelope, err := p.seal(newRevision, state)
	if err != nil {
		return 0, err
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		return 0, err
	}
	if err := atomicWriteFile(p.path, encoded, 0o600); err != nil {
		return 0, err
	}
	if err := p.anchor.CompareAndAdvance(p.namespace, currentRevision, newRevision); err != nil {
		return 0, fmt.Errorf("advance key state anchor: %w", err)
	}
	return newRevision, nil
}

func fixtureStatePathLock(path string) *sync.Mutex {
	lock, _ := fixtureStateLocks.LoadOrStore(path, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (p *FixtureFilePersistence) load() (uint64, []byte, error) {
	revision, state, err := p.loadFile()
	if err != nil {
		return 0, nil, err
	}
	anchored, err := p.anchor.Current(p.namespace)
	if err != nil {
		return 0, nil, err
	}
	if revision != anchored {
		return 0, nil, fmt.Errorf("%w: key state has %d, anchor has %d", contracts.ErrRevisionRollback, revision, anchored)
	}
	return revision, state, nil
}

func (p *FixtureFilePersistence) loadFile() (uint64, []byte, error) {
	encoded, err := os.ReadFile(p.path)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil, ErrStateNotFound
	}
	if err != nil {
		return 0, nil, err
	}
	var envelope fixtureFileEnvelope
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return 0, nil, fmt.Errorf("decode key state: %w", err)
	}
	if envelope.Version != fixtureStateVersion || envelope.Namespace != p.namespace || envelope.Revision == 0 {
		return 0, nil, errors.New("unsupported or invalid key state version")
	}
	digest := sha256.Sum256(envelope.Ciphertext)
	if hex.EncodeToString(digest[:]) != envelope.Checksum {
		return 0, nil, errors.New("key state checksum mismatch")
	}
	block, err := aes.NewCipher(p.wrappingKey[:])
	if err != nil {
		return 0, nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return 0, nil, err
	}
	state, err := aead.Open(nil, envelope.Nonce, envelope.Ciphertext, stateAAD(envelope.Namespace, envelope.Revision))
	if err != nil {
		return 0, nil, fmt.Errorf("authenticate key state: %w", err)
	}
	return envelope.Revision, state, nil
}

func (p *FixtureFilePersistence) seal(revision uint64, state []byte) (*fixtureFileEnvelope, error) {
	block, err := aes.NewCipher(p.wrappingKey[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := aead.Seal(nil, nonce, state, stateAAD(p.namespace, revision))
	digest := sha256.Sum256(ciphertext)
	return &fixtureFileEnvelope{
		Version: fixtureStateVersion, Namespace: p.namespace, Revision: revision, Nonce: nonce,
		Ciphertext: ciphertext, Checksum: hex.EncodeToString(digest[:]),
	}, nil
}

func stateAAD(namespace string, revision uint64) []byte {
	return []byte(fmt.Sprintf("virtengine-fixture-key-state:%d:%s:%d", fixtureStateVersion, namespace, revision))
}

func atomicWriteFile(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := rejectSymlinkTarget(path); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if err := rejectSymlinkTarget(dir); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, ".vault-state-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
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
	if err := os.Rename(tempName, path); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return nil
	}
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func rejectSymlinkTarget(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	for current := abs; ; current = filepath.Dir(current) {
		info, statErr := os.Lstat(current)
		if statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("fixture path ancestor must not be a symlink: %s", current)
			}
			if runtime.GOOS == "windows" {
				resolved, resolveErr := filepath.EvalSymlinks(current)
				if resolveErr != nil {
					return resolveErr
				}
				resolvedAbs, resolveErr := filepath.Abs(resolved)
				if resolveErr != nil || !strings.EqualFold(filepath.Clean(resolvedAbs), filepath.Clean(current)) {
					return fmt.Errorf("fixture path ancestor is a reparse point: %s", current)
				}
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return statErr
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return nil
}

func enforceKeyPathSecurity(path string, directory bool, options contracts.FixtureSecurityOptions) error {
	if err := rejectSymlinkTarget(path); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		if !options.UnsafeWindowsDevelopment {
			return errors.New("fixture key custody cannot enforce safe Windows ACLs; UnsafeWindowsDevelopment is required")
		}
		return nil
	}
	mode := os.FileMode(0o600)
	if directory {
		mode = 0o700
	}
	if err := os.Chmod(path, mode); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

type persistedManagerState struct {
	Keys             map[Scope]map[string]*KeyInfo `json:"keys"`
	ActiveKeys       map[Scope]string              `json:"active_keys"`
	RotationPolicies map[Scope]*RotationPolicy     `json:"rotation_policies"`
	RotationState    map[Scope]*KeyRotation        `json:"rotation_state"`
}

// PersistentKeyManager persists all key versions and rotation state after each mutation.
type PersistentKeyManager struct {
	manager     *KeyManager
	persistence StatePersistence
	revision    uint64
	mu          sync.Mutex
}

// NewPersistentKeyManager loads existing state. ErrStateNotFound is returned for a new store.
func NewPersistentKeyManager(persistence StatePersistence) (*PersistentKeyManager, error) {
	if persistence == nil {
		return nil, errors.New("key state persistence is required")
	}
	revision, state, err := persistence.Load()
	if err != nil {
		return nil, err
	}
	manager := NewKeyManager()
	if err := restoreManager(manager, state); err != nil {
		return nil, err
	}
	return &PersistentKeyManager{manager: manager, persistence: persistence, revision: revision}, nil
}

// NewUninitializedPersistentKeyManager creates an empty manager for explicit first initialization.
func NewUninitializedPersistentKeyManager(persistence StatePersistence) (*PersistentKeyManager, error) {
	if persistence == nil {
		return nil, errors.New("key state persistence is required")
	}
	if _, _, err := persistence.Load(); !errors.Is(err, ErrStateNotFound) {
		if err == nil {
			return nil, errors.New("key state already exists")
		}
		return nil, err
	}
	return &PersistentKeyManager{manager: NewKeyManager(), persistence: persistence}, nil
}

// Initialize explicitly creates initial fixture keys and persists them.
func (m *PersistentKeyManager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mutate(func() error { return m.manager.Initialize() })
}

func (m *PersistentKeyManager) GetActiveKey(scope Scope) (*KeyInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.manager.GetActiveKey(scope)
}
func (m *PersistentKeyManager) GetKey(scope Scope, keyID string) (*KeyInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.manager.GetKey(scope, keyID)
}
func (m *PersistentKeyManager) GetRotationStatus(scope Scope) (*KeyRotation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.manager.GetRotationStatus(scope)
}
func (m *PersistentKeyManager) ListKeys(scope Scope) ([]*KeyInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.manager.ListKeys(scope)
}

// Close releases resources owned by the persistence adapter.
func (m *PersistentKeyManager) Close() error {
	if closer, ok := m.persistence.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// Revision returns the durable state revision used to fence fixture erasure.
func (m *PersistentKeyManager) Revision() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.revision
}

func (m *PersistentKeyManager) RotateKey(scope Scope, overlap time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mutate(func() error { return m.manager.RotateKey(scope, overlap) })
}

func (m *PersistentKeyManager) CompleteRotation(scope Scope) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mutate(func() error { return m.manager.CompleteRotation(scope) })
}

// DestroyKey irreversibly removes a key and returns a non-secret destruction receipt digest.
func (m *PersistentKeyManager) DestroyKey(scope Scope, keyID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.mutate(func() error {
		m.manager.mu.Lock()
		defer m.manager.mu.Unlock()
		scopeKeys := m.manager.keys[scope]
		if scopeKeys == nil || scopeKeys[keyID] == nil {
			return fmt.Errorf("key %s not found for scope %s", keyID, scope)
		}
		delete(scopeKeys, keyID)
		if m.manager.activeKeys[scope] == keyID {
			delete(m.manager.activeKeys, scope)
		}
		return nil
	}); err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(fmt.Sprintf("destroyed:%s:%s:%d", scope, keyID, m.revision)))
	return hex.EncodeToString(digest[:]), nil
}

func (m *PersistentKeyManager) mutate(mutation func() error) error {
	before, err := snapshotManager(m.manager)
	if err != nil {
		return err
	}
	if err := mutation(); err != nil {
		_ = restoreManager(m.manager, before)
		return err
	}
	if err := m.persist(); err != nil {
		if restoreErr := restoreManager(m.manager, before); restoreErr != nil {
			return fmt.Errorf("persist key state: %v (rollback failed: %w)", err, restoreErr)
		}
		return err
	}
	return nil
}

func (m *PersistentKeyManager) persist() error {
	state, err := snapshotManager(m.manager)
	if err != nil {
		return err
	}
	newRevision, err := m.persistence.CompareAndSwap(m.revision, state)
	if err != nil {
		return err
	}
	m.revision = newRevision
	return nil
}

func snapshotManager(manager *KeyManager) ([]byte, error) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return json.Marshal(persistedManagerState{
		Keys: manager.keys, ActiveKeys: manager.activeKeys,
		RotationPolicies: manager.rotationPolicies, RotationState: manager.rotationState,
	})
}

func restoreManager(manager *KeyManager, encoded []byte) error {
	var state persistedManagerState
	if err := json.Unmarshal(encoded, &state); err != nil {
		return fmt.Errorf("decode persisted key manager: %w", err)
	}
	if state.Keys == nil || state.ActiveKeys == nil || state.RotationPolicies == nil || state.RotationState == nil {
		return errors.New("persisted key manager state is incomplete")
	}
	manager.keys = state.Keys
	manager.activeKeys = state.ActiveKeys
	manager.rotationPolicies = state.RotationPolicies
	manager.rotationState = state.RotationState
	return nil
}

// DestroyKeyForFixtureErasure is disabled because caller-provided holds cannot authorize destruction.
// Deprecated: use data_vault.FixtureErasureCoordinator, which reads durable holds itself.
func (m *PersistentKeyManager) DestroyKeyForFixtureErasure(scope Scope, keyID, authorizationDigest string, holds []contracts.LegalHoldAuthority, verifier contracts.HoldAuthorityVerifier) (contracts.DestructionReceipt, error) {
	return contracts.DestructionReceipt{}, errors.New("unsafe fixture erasure API disabled; use FixtureErasureCoordinator")
}
