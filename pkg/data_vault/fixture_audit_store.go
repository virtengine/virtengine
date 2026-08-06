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
	"sync"
)

const fixtureAuditVersion uint32 = 1

var fixtureAuditLocks sync.Map

type fixtureAuditEnvelope struct {
	Version   uint32          `json:"version"`
	Namespace string          `json:"namespace"`
	Revision  uint64          `json:"revision"`
	Events    json.RawMessage `json:"events"`
	Checksum  string          `json:"checksum"`
}

// DurableAuditStore marks audit stores that survive process restart.
type DurableAuditStore interface {
	AuditStore
	Durable() bool
}

// FixtureFileAuditStore is a fixture-only durable audit store.
type FixtureFileAuditStore struct {
	path      string
	events    []*AuditEvent
	revision  uint64
	anchor    RevisionAnchor
	namespace string
	lockFile  *os.File
	mu        sync.RWMutex
}

// NewFixtureFileAuditStore opens or creates a fixture audit store.
func NewFixtureFileAuditStore(path, profile string, anchors ...RevisionAnchor) (*FixtureFileAuditStore, error) {
	return newFixtureFileAuditStore(path, profile, FixtureSecurityOptions{}, anchors...)
}

// NewFixtureFileAuditStoreWithSecurity allows an explicit fixture-only Windows ACL override.
func NewFixtureFileAuditStoreWithSecurity(path, profile string, options FixtureSecurityOptions, anchors ...RevisionAnchor) (*FixtureFileAuditStore, error) {
	return newFixtureFileAuditStore(path, profile, options, anchors...)
}

func newFixtureFileAuditStore(path, profile string, options FixtureSecurityOptions, anchors ...RevisionAnchor) (*FixtureFileAuditStore, error) {
	if profile == "production" {
		return nil, errors.New("fixture audit store is forbidden in production")
	}
	if profile != "fixture" && profile != "development" {
		return nil, errors.New("fixture audit store requires fixture or development profile")
	}
	if path == "" {
		return nil, errors.New("fixture audit path is required")
	}
	if len(anchors) != 1 || anchors[0] == nil || anchors[0].Replayable() {
		return nil, errors.New("fixture audit store requires one non-replayable revision anchor")
	}
	cleanPath := filepath.Clean(path)
	abs, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, err
	}
	if err := rejectFixtureSymlink(cleanPath); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0o700); err != nil {
		return nil, err
	}
	if err := enforceFixturePathSecurity(filepath.Dir(cleanPath), true, options); err != nil {
		return nil, err
	}
	if err := rejectFixtureSymlink(filepath.Dir(cleanPath)); err != nil {
		return nil, err
	}
	lockPath := cleanPath + ".lock"
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
	store := &FixtureFileAuditStore{
		path: cleanPath, events: make([]*AuditEvent, 0), anchor: anchors[0],
		namespace: "vault-audit:" + filepath.ToSlash(abs), lockFile: lockFile,
	}
	if err := store.reload(); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = store.Close()
		return nil, err
	}
	if err := enforceFixturePathSecurity(cleanPath, false, options); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = store.Close()
		return nil, err
	}
	anchored, err := store.anchor.Current(store.namespace)
	if err != nil || anchored != store.revision {
		_ = store.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%w: audit has %d, anchor has %d", ErrRevisionRollback, store.revision, anchored)
	}
	return store, nil
}

// Close releases the fixture adapter's cross-process lease.
func (s *FixtureFileAuditStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := unlockFile(s.lockFile)
	s.lockFile = nil
	return err
}

func (*FixtureFileAuditStore) Durable() bool { return true }

func (s *FixtureFileAuditStore) Append(_ context.Context, event *AuditEvent) error {
	if event == nil {
		return errors.New("audit event required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return err
	}
	if err := validateAuditAppend(s.events, event); err != nil {
		return err
	}
	s.events = append(s.events, cloneAuditEvent(event))
	if err := s.persist(); err != nil {
		s.events = s.events[:len(s.events)-1]
		return err
	}
	return nil
}

func (s *FixtureFileAuditStore) Query(_ context.Context, filter AuditFilter) ([]*AuditEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*AuditEvent, 0, len(s.events))
	for _, event := range s.events {
		if filter.BlobID != "" && event.BlobID != filter.BlobID ||
			filter.Scope != "" && event.Scope != filter.Scope ||
			filter.Requester != "" && event.Requester != filter.Requester ||
			filter.OrgID != "" && event.OrgID != filter.OrgID ||
			filter.StartTime != nil && event.Timestamp.Unix() < *filter.StartTime ||
			filter.EndTime != nil && event.Timestamp.Unix() > *filter.EndTime {
			continue
		}
		result = append(result, cloneAuditEvent(event))
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result, nil
}

func (s *FixtureFileAuditStore) ensureCurrent() error {
	revision, _, err := readFixtureAudit(s.path)
	if errors.Is(err, os.ErrNotExist) && s.revision == 0 {
		return nil
	}
	if err != nil {
		return err
	}
	if revision != s.revision {
		return fmt.Errorf("stale audit revision: have %d, expected %d", revision, s.revision)
	}
	return nil
}

func (s *FixtureFileAuditStore) reload() error {
	revision, events, err := readFixtureAudit(s.path)
	if err != nil {
		return err
	}
	encoded, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var envelope fixtureAuditEnvelope
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return err
	}
	if envelope.Namespace != s.namespace {
		return errors.New("fixture audit namespace mismatch")
	}
	s.revision, s.events = revision, events
	return nil
}

func (s *FixtureFileAuditStore) persist() error {
	pathLock := fixtureAuditPathLock(s.path)
	pathLock.Lock()
	defer pathLock.Unlock()
	if err := s.ensureCurrent(); err != nil {
		return err
	}
	events, err := json.Marshal(s.events)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(events)
	envelope := fixtureAuditEnvelope{
		Version: fixtureAuditVersion, Namespace: s.namespace, Revision: s.revision + 1,
		Events: events, Checksum: hex.EncodeToString(digest[:]),
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	if err := atomicWriteFixtureFile(s.path, encoded, 0o600); err != nil {
		return err
	}
	if err := s.anchor.CompareAndAdvance(s.namespace, s.revision, s.revision+1); err != nil {
		return fmt.Errorf("advance audit revision anchor: %w", err)
	}
	s.revision++
	return nil
}

func fixtureAuditPathLock(path string) *sync.Mutex {
	lock, _ := fixtureAuditLocks.LoadOrStore(path, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func readFixtureAudit(path string) (uint64, []*AuditEvent, error) {
	encoded, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, err
	}
	var envelope fixtureAuditEnvelope
	if err := json.Unmarshal(encoded, &envelope); err != nil {
		return 0, nil, err
	}
	digest := sha256.Sum256(envelope.Events)
	if envelope.Version != fixtureAuditVersion || envelope.Revision == 0 || hex.EncodeToString(digest[:]) != envelope.Checksum {
		return 0, nil, errors.New("fixture audit checksum or version invalid")
	}
	var events []*AuditEvent
	if err := json.Unmarshal(envelope.Events, &events); err != nil {
		return 0, nil, err
	}
	if envelope.Revision != uint64(len(events)) {
		return 0, nil, errors.New("fixture audit revision does not match event count")
	}
	if err := validateAuditEvents(events); err != nil {
		return 0, nil, err
	}
	return envelope.Revision, events, nil
}

func validateAuditAppend(events []*AuditEvent, event *AuditEvent) error {
	expected := ""
	if len(events) > 0 {
		expected = events[len(events)-1].Hash
	}
	if event.PreviousHash != expected {
		return errors.New("audit append predecessor mismatch")
	}
	if event.Hash == "" || event.Hash != computeAuditHash(event) {
		return errors.New("audit append hash mismatch")
	}
	return nil
}
