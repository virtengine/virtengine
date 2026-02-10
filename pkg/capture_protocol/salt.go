package capture_protocol

import (
	"crypto/sha256"
	"sync"
	"time"
)

// SaltValidator validates salt binding and freshness for the capture protocol.
// It ensures salts are properly formatted, fresh, correctly bound, and not replayed.
type SaltValidator struct {
	// Configuration
	minSaltLen   int
	maxSaltAge   time.Duration
	replayWindow time.Duration
	maxClockSkew time.Duration

	// Salt cache for replay detection
	usedSalts *saltCache

	// Time source (for testing)
	now func() time.Time
}

// SaltValidatorOption is a functional option for SaltValidator
type SaltValidatorOption func(*SaltValidator)

// WithMinSaltLength sets the minimum salt length
func WithMinSaltLength(length int) SaltValidatorOption {
	return func(sv *SaltValidator) {
		sv.minSaltLen = length
	}
}

// WithMaxSaltAge sets the maximum salt age
func WithMaxSaltAge(age time.Duration) SaltValidatorOption {
	return func(sv *SaltValidator) {
		sv.maxSaltAge = age
	}
}

// WithReplayWindow sets the replay detection window
func WithReplayWindow(window time.Duration) SaltValidatorOption {
	return func(sv *SaltValidator) {
		sv.replayWindow = window
	}
}

// WithMaxClockSkew sets the maximum allowed clock skew
func WithMaxClockSkew(skew time.Duration) SaltValidatorOption {
	return func(sv *SaltValidator) {
		sv.maxClockSkew = skew
	}
}

// WithTimeSource sets a custom time source (for testing)
func WithTimeSource(now func() time.Time) SaltValidatorOption {
	return func(sv *SaltValidator) {
		sv.now = now
	}
}

// WithSaltCacheSize sets the maximum size of the salt cache
func WithSaltCacheSize(size int) SaltValidatorOption {
	return func(sv *SaltValidator) {
		sv.usedSalts = newSaltCache(size, sv.replayWindow)
	}
}

// NewSaltValidator creates a new SaltValidator with the given options
func NewSaltValidator(opts ...SaltValidatorOption) *SaltValidator {
	sv := &SaltValidator{
		minSaltLen:   MinSaltLength,
		maxSaltAge:   DefaultMaxSaltAge,
		replayWindow: DefaultReplayWindow,
		maxClockSkew: DefaultMaxClockSkew,
		now:          time.Now,
	}

	for _, opt := range opts {
		opt(sv)
	}

	// Initialize cache if not set by options
	if sv.usedSalts == nil {
		sv.usedSalts = newSaltCacheWithTimeSource(100000, sv.replayWindow, sv.now)
	}

	return sv
}

// ValidateSalt performs complete salt validation including:
// - Salt length check
// - Binding hash verification
// - Timestamp freshness check
// - Replay detection
func (sv *SaltValidator) ValidateSalt(binding SaltBinding) error {
	// 1. Check salt length
	if err := sv.validateSaltLength(binding.Salt); err != nil {
		return err
	}

	// 2. Check binding hash
	if err := sv.validateBindingHash(binding); err != nil {
		return err
	}

	// 3. Check timestamp freshness
	if err := sv.validateSaltFreshness(binding.Timestamp); err != nil {
		return err
	}

	// 4. Check not replayed
	if err := sv.checkNotReplayed(binding.Salt); err != nil {
		return err
	}

	return nil
}

// ValidateSaltOnly validates just the salt bytes without binding
func (sv *SaltValidator) ValidateSaltOnly(salt []byte) error {
	return sv.validateSaltLength(salt)
}

// RecordUsedSalt records a salt as used to prevent replay
func (sv *SaltValidator) RecordUsedSalt(salt []byte) error {
	saltHash := sv.computeSaltHash(salt)
	sv.usedSalts.add(saltHash)
	return nil
}

// IsSaltUsed checks if a salt has been used before
func (sv *SaltValidator) IsSaltUsed(salt []byte) bool {
	saltHash := sv.computeSaltHash(salt)
	return sv.usedSalts.exists(saltHash)
}

// ClearExpiredSalts removes expired entries from the cache
func (sv *SaltValidator) ClearExpiredSalts() int {
	return sv.usedSalts.cleanup()
}

// validateSaltLength checks if salt meets minimum length requirements
func (sv *SaltValidator) validateSaltLength(salt []byte) error {
	if salt == nil {
		return ErrSaltEmpty
	}

	if len(salt) < sv.minSaltLen {
		return ErrSaltTooShort.WithDetails(
			"minimum_length", sv.minSaltLen,
			"actual_length", len(salt),
		)
	}

	if len(salt) > MaxSaltLength {
		return ErrSaltTooLong.WithDetails(
			"maximum_length", MaxSaltLength,
			"actual_length", len(salt),
		)
	}

	// Check for weak salts (all zeros or all same value)
	if isWeakSalt(salt) {
		return ErrSaltWeak
	}

	return nil
}

// validateBindingHash verifies the salt binding hash is correct
func (sv *SaltValidator) validateBindingHash(binding SaltBinding) error {
	if len(binding.BindingHash) == 0 {
		return ErrBindingHashMissing
	}

	if !binding.VerifyBindingHash() {
		return ErrBindingHashMismatch
	}

	// Verify salt in binding matches
	if !constantTimeEqual(binding.Salt, binding.Salt) {
		return ErrBindingSaltMismatch
	}

	// Verify device ID is present
	if binding.DeviceID == "" {
		return ErrBindingDeviceIDMissing
	}

	// Verify session ID is present
	if binding.SessionID == "" {
		return ErrBindingSessionIDMissing
	}

	return nil
}

// validateSaltFreshness checks if the salt timestamp is within acceptable bounds
func (sv *SaltValidator) validateSaltFreshness(timestamp int64) error {
	now := sv.now()
	saltTime := time.Unix(timestamp, 0)

	// Check if salt is too old
	age := now.Sub(saltTime)
	if age > sv.maxSaltAge {
		return ErrSaltExpired.WithDetails(
			"max_age", sv.maxSaltAge.String(),
			"actual_age", age.String(),
		)
	}

	// Check if salt is from the future (clock skew)
	if age < -sv.maxClockSkew {
		return ErrSaltFromFuture.WithDetails(
			"max_skew", sv.maxClockSkew.String(),
			"time_ahead", (-age).String(),
		)
	}

	return nil
}

// checkNotReplayed verifies the salt has not been used before
func (sv *SaltValidator) checkNotReplayed(salt []byte) error {
	saltHash := sv.computeSaltHash(salt)

	if sv.usedSalts.exists(saltHash) {
		return ErrSaltReplayed
	}

	return nil
}

// computeSaltHash computes a hash of the salt for cache lookup
func (sv *SaltValidator) computeSaltHash(salt []byte) [32]byte {
	return sha256.Sum256(salt)
}

// isWeakSalt checks if a salt value is weak (all zeros or all same byte)
func isWeakSalt(salt []byte) bool {
	if len(salt) == 0 {
		return true
	}

	// Check for all zeros
	allZeros := true
	for _, b := range salt {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		return true
	}

	// Check for all same value
	first := salt[0]
	allSame := true
	for _, b := range salt {
		if b != first {
			allSame = false
			break
		}
	}
	return allSame
}

// saltCache is a thread-safe cache for tracking used salts
type saltCache struct {
	mu      sync.RWMutex
	entries map[[32]byte]time.Time
	maxSize int
	ttl     time.Duration
	now     func() time.Time
}

// newSaltCache creates a new salt cache
func newSaltCache(maxSize int, ttl time.Duration) *saltCache {
	return newSaltCacheWithTimeSource(maxSize, ttl, time.Now)
}

// newSaltCacheWithTimeSource creates a new salt cache with a custom time source
func newSaltCacheWithTimeSource(maxSize int, ttl time.Duration, now func() time.Time) *saltCache {
	return &saltCache{
		entries: make(map[[32]byte]time.Time),
		maxSize: maxSize,
		ttl:     ttl,
		now:     now,
	}
}

// add adds a salt hash to the cache
func (c *saltCache) add(hash [32]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if at capacity
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	c.entries[hash] = c.now()
}

// exists checks if a salt hash exists in the cache
func (c *saltCache) exists(hash [32]byte) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	timestamp, ok := c.entries[hash]
	if !ok {
		return false
	}

	// Check if entry has expired
	if c.now().Sub(timestamp) > c.ttl {
		return false
	}

	return true
}

// cleanup removes expired entries and returns count removed
func (c *saltCache) cleanup() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	removed := 0

	for hash, timestamp := range c.entries {
		if now.Sub(timestamp) > c.ttl {
			delete(c.entries, hash)
			removed++
		}
	}

	return removed
}

// evictOldest removes the oldest entry (caller must hold lock)
func (c *saltCache) evictOldest() {
	var oldestHash [32]byte
	var oldestTime time.Time
	first := true

	for hash, timestamp := range c.entries {
		if first || timestamp.Before(oldestTime) {
			oldestHash = hash
			oldestTime = timestamp
			first = false
		}
	}

	if !first {
		delete(c.entries, oldestHash)
	}
}

// size returns the current cache size
//
//nolint:unused // Reserved for diagnostics and testing
func (c *saltCache) size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// ComputeBindingHash is a helper function to compute a salt binding hash
func ComputeBindingHash(salt []byte, deviceID, sessionID string, timestamp int64) []byte {
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(deviceID))
	h.Write([]byte(sessionID))
	h.Write(int64ToBytes(timestamp))
	return h.Sum(nil)
}

// CreateSaltBinding creates a new salt binding with computed hash
func CreateSaltBinding(salt []byte, deviceID, sessionID string, timestamp int64) SaltBinding {
	return SaltBinding{
		Salt:        salt,
		DeviceID:    deviceID,
		SessionID:   sessionID,
		Timestamp:   timestamp,
		BindingHash: ComputeBindingHash(salt, deviceID, sessionID, timestamp),
	}
}
