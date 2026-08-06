package keys

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/data_vault/contracts"
)

func newTestFixturePersistence(path string, wrappingKey []byte, profile string, anchor contracts.RevisionAnchor) (*FixtureFilePersistence, error) {
	return NewFixtureFilePersistenceWithSecurity(path, wrappingKey, profile, contracts.FixtureSecurityOptions{
		UnsafeWindowsDevelopment: true,
	}, anchor)
}

func TestPersistentKeyManagerRestartRotationAndStaleWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.state")
	wrappingKey := []byte("0123456789abcdef0123456789abcdef")
	anchor := contracts.NewProcessRevisionAnchor()
	persistence, err := newTestFixturePersistence(path, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	manager, err := NewUninitializedPersistentKeyManager(persistence)
	require.NoError(t, err)
	require.NoError(t, manager.Initialize())
	original, err := manager.GetActiveKey(ScopeSupport)
	require.NoError(t, err)
	_, err = newTestFixturePersistence(path, wrappingKey, "fixture", anchor)
	require.ErrorIs(t, err, errFixtureKeyStateInUse)
	require.NoError(t, manager.Close())

	reloadedPersistence, err := newTestFixturePersistence(path, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	reloaded, err := NewPersistentKeyManager(reloadedPersistence)
	require.NoError(t, err)
	restored, err := reloaded.GetActiveKey(ScopeSupport)
	require.NoError(t, err)
	require.Equal(t, original.ID, restored.ID)
	require.Equal(t, original.Version, restored.Version)
	require.Equal(t, original.PublicKey, restored.PublicKey)
	require.Equal(t, original.PrivateKey, restored.PrivateKey)
	require.True(t, original.CreatedAt.Equal(restored.CreatedAt))

	require.NoError(t, reloaded.RotateKey(ScopeSupport, time.Hour))
	require.NoError(t, reloaded.Close())

	afterRotationPersistence, err := newTestFixturePersistence(path, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	afterRotation, err := NewPersistentKeyManager(afterRotationPersistence)
	require.NoError(t, err)
	rotation, err := afterRotation.GetRotationStatus(ScopeSupport)
	require.NoError(t, err)
	require.Equal(t, RotationStatusInProgress, rotation.Status)
	active, err := afterRotation.GetActiveKey(ScopeSupport)
	require.NoError(t, err)
	require.Equal(t, rotation.NewKeyID, active.ID)

	onDisk, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NotContains(t, string(onDisk), original.ID)
	require.NoError(t, afterRotation.Close())
}

type staleStatePersistence struct {
	revision uint64
	state    []byte
	missing  bool
}

type blockingStalePersistence struct {
	staleStatePersistence
	entered chan struct{}
	release chan struct{}
}

func (p *blockingStalePersistence) CompareAndSwap(uint64, []byte) (uint64, error) {
	close(p.entered)
	<-p.release
	return 0, ErrStaleRevision
}

func (p *staleStatePersistence) Load() (uint64, []byte, error) {
	if p.missing {
		return 0, nil, ErrStateNotFound
	}
	return p.revision, append([]byte(nil), p.state...), nil
}

func (*staleStatePersistence) CompareAndSwap(uint64, []byte) (uint64, error) {
	return 0, ErrStaleRevision
}

func TestPersistentKeyManagerMutationRollbackOnCASFailure(t *testing.T) {
	initializePersistence := &staleStatePersistence{missing: true}
	initializeManager, err := NewUninitializedPersistentKeyManager(initializePersistence)
	require.NoError(t, err)
	require.ErrorIs(t, initializeManager.Initialize(), ErrStaleRevision)
	_, err = initializeManager.GetActiveKey(ScopeSupport)
	require.Error(t, err)

	base := NewKeyManager()
	require.NoError(t, base.Initialize())
	initializedState, err := snapshotManager(base)
	require.NoError(t, err)

	rotateManager, err := NewPersistentKeyManager(&staleStatePersistence{revision: 1, state: initializedState})
	require.NoError(t, err)
	activeBefore, err := rotateManager.GetActiveKey(ScopeSupport)
	require.NoError(t, err)
	require.ErrorIs(t, rotateManager.RotateKey(ScopeSupport, time.Hour), ErrStaleRevision)
	activeAfter, err := rotateManager.GetActiveKey(ScopeSupport)
	require.NoError(t, err)
	require.Equal(t, activeBefore.ID, activeAfter.ID)
	_, err = rotateManager.GetRotationStatus(ScopeSupport)
	require.Error(t, err)

	require.NoError(t, base.RotateKey(ScopeSupport, time.Hour))
	rotatedState, err := snapshotManager(base)
	require.NoError(t, err)
	completeManager, err := NewPersistentKeyManager(&staleStatePersistence{revision: 2, state: rotatedState})
	require.NoError(t, err)
	rotationBefore, err := completeManager.GetRotationStatus(ScopeSupport)
	require.NoError(t, err)
	require.ErrorIs(t, completeManager.CompleteRotation(ScopeSupport), ErrStaleRevision)
	rotationAfter, err := completeManager.GetRotationStatus(ScopeSupport)
	require.NoError(t, err)
	require.Equal(t, rotationBefore.Status, rotationAfter.Status)

	destroyManager, err := NewPersistentKeyManager(&staleStatePersistence{revision: 1, state: initializedState})
	require.NoError(t, err)
	key, err := destroyManager.GetActiveKey(ScopeSupport)
	require.NoError(t, err)
	_, err = destroyManager.DestroyKey(ScopeSupport, key.ID)
	require.ErrorIs(t, err, ErrStaleRevision)
	_, err = destroyManager.GetKey(ScopeSupport, key.ID)
	require.NoError(t, err)
}

func TestPersistentKeyManagerDoesNotExposeMutationBeforeCASCompletes(t *testing.T) {
	base := NewKeyManager()
	require.NoError(t, base.Initialize())
	state, err := snapshotManager(base)
	require.NoError(t, err)
	persistence := &blockingStalePersistence{
		staleStatePersistence: staleStatePersistence{revision: 1, state: state},
		entered:               make(chan struct{}),
		release:               make(chan struct{}),
	}
	manager, err := NewPersistentKeyManager(persistence)
	require.NoError(t, err)
	before, err := manager.GetActiveKey(ScopeSupport)
	require.NoError(t, err)
	mutationResult := make(chan error, 1)
	go func() { mutationResult <- manager.RotateKey(ScopeSupport, time.Hour) }()
	<-persistence.entered
	readResult := make(chan *KeyInfo, 1)
	go func() {
		key, _ := manager.GetActiveKey(ScopeSupport)
		readResult <- key
	}()
	select {
	case key := <-readResult:
		t.Fatalf("read exposed key %q before CAS completed", key.ID)
	case <-time.After(50 * time.Millisecond):
	}
	close(persistence.release)
	require.ErrorIs(t, <-mutationResult, ErrStaleRevision)
	after := <-readResult
	require.Equal(t, before.ID, after.ID)
}

func TestPersistentKeyManagerLoadDoesNotRegenerate(t *testing.T) {
	persistence, err := newTestFixturePersistence(filepath.Join(t.TempDir(), "missing"), make([]byte, 32), "fixture", contracts.NewProcessRevisionAnchor())
	require.NoError(t, err)
	defer persistence.Close()
	_, err = NewPersistentKeyManager(persistence)
	require.ErrorIs(t, err, ErrStateNotFound)
	_, err = NewFixtureFilePersistence("keys", make([]byte, 32), "production")
	require.Error(t, err)
}

func TestFixtureFilePersistenceRejectsSymlinkTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	require.NoError(t, os.WriteFile(target, nil, 0o600))
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	_, err := newTestFixturePersistence(link, make([]byte, 32), "fixture", contracts.NewProcessRevisionAnchor())
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrStateNotFound))
}

func TestFixtureFilePersistenceRejectsOlderValidStateReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.state")
	wrappingKey := []byte("0123456789abcdef0123456789abcdef")
	anchor := contracts.NewProcessRevisionAnchor()
	persistence, err := newTestFixturePersistence(path, wrappingKey, "fixture", anchor)
	require.NoError(t, err)
	manager, err := NewUninitializedPersistentKeyManager(persistence)
	require.NoError(t, err)
	require.NoError(t, manager.Initialize())
	older, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, manager.RotateKey(ScopeSupport, time.Hour))
	require.NoError(t, manager.Close())
	require.NoError(t, os.WriteFile(path, older, 0o600))
	_, err = newTestFixturePersistence(path, wrappingKey, "fixture", anchor)
	require.ErrorIs(t, err, contracts.ErrRevisionRollback)
}
