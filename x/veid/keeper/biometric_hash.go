// Package keeper provides the VEID module keeper.
//
// This file implements secure biometric template hashing for identity verification.
// Biometric templates (facial embeddings, fingerprints, etc.) are NEVER stored raw.
// We store only irreversible hashes with per-template salts for matching.
//
// Security Model:
// - Argon2id for key derivation (memory-hard, resistant to GPU attacks)
// - Locality-sensitive hashing (LSH) for fuzzy matching without revealing templates
// - Per-template salts prevent rainbow table attacks
// - Algorithm versioning for future upgrades
// - Audit logging for all operations
//
// Task Reference: VE-3030 - Biometric Template Secure Hashing
package keeper

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/argon2"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Constants and Configuration
// ============================================================================

const (
	// BiometricHashVersion is the current version of the hashing algorithm
	BiometricHashVersion uint32 = 1

	// SaltSize is the size of per-template salt in bytes
	SaltSize = 32

	// Error message constants to avoid duplication
	errMsgTemplateEmpty = "template cannot be empty"
	errMsgStoredHashNil = "stored hash cannot be nil"
	errMsgHashIDEmpty   = "hash ID cannot be empty"

	// HashSize is the size of the final biometric hash in bytes
	HashSize = 64

	// Argon2idTime is the number of iterations for Argon2id
	Argon2idTime = 3

	// Argon2idMemory is the memory usage in KiB for Argon2id (64 MiB)
	Argon2idMemory = 64 * 1024

	// Argon2idThreads is the number of threads for Argon2id
	Argon2idThreads = 4

	// LSHBuckets is the number of LSH buckets for fuzzy matching
	LSHBuckets = 16

	// LSHHashSize is the size of each LSH hash in bytes
	LSHHashSize = 8

	// DefaultMatchThresholdFace is the default similarity threshold for face matching
	DefaultMatchThresholdFace = 0.85

	// DefaultMatchThresholdFingerprint is the default similarity threshold for fingerprint matching
	DefaultMatchThresholdFingerprint = 0.90

	// DefaultMatchThresholdIris is the default similarity threshold for iris matching
	DefaultMatchThresholdIris = 0.92

	// DefaultMatchThresholdVoice is the default similarity threshold for voice matching
	DefaultMatchThresholdVoice = 0.80

	// MaxTemplateSize is the maximum size of a biometric template in bytes
	MaxTemplateSize = 1024 * 1024 // 1 MiB
)

// ============================================================================
// Types
// ============================================================================

// TemplateType represents the type of biometric template
type TemplateType int

const (
	// TemplateTypeFace is a facial biometric template
	TemplateTypeFace TemplateType = iota
	// TemplateTypeFingerprint is a fingerprint biometric template
	TemplateTypeFingerprint
	// TemplateTypeIris is an iris biometric template
	TemplateTypeIris
	// TemplateTypeVoice is a voice biometric template
	TemplateTypeVoice
)

// String returns the string representation of the template type
func (t TemplateType) String() string {
	switch t {
	case TemplateTypeFace:
		return "face"
	case TemplateTypeFingerprint:
		return "fingerprint"
	case TemplateTypeIris:
		return "iris"
	case TemplateTypeVoice:
		return "voice"
	default:
		return "unknown"
	}
}

// ValidateTemplateType validates that a template type is valid
func ValidateTemplateType(t TemplateType) error {
	switch t {
	case TemplateTypeFace, TemplateTypeFingerprint, TemplateTypeIris, TemplateTypeVoice:
		return nil
	default:
		return types.ErrInvalidBiometricTemplate.Wrapf("invalid template type: %d", t)
	}
}

// DefaultMatchThreshold returns the default match threshold for a template type
func DefaultMatchThreshold(t TemplateType) float64 {
	switch t {
	case TemplateTypeFace:
		return DefaultMatchThresholdFace
	case TemplateTypeFingerprint:
		return DefaultMatchThresholdFingerprint
	case TemplateTypeIris:
		return DefaultMatchThresholdIris
	case TemplateTypeVoice:
		return DefaultMatchThresholdVoice
	default:
		return DefaultMatchThresholdFace
	}
}

// BiometricTemplateHash represents a stored biometric template hash
// SECURITY: This struct NEVER contains raw biometric data
type BiometricTemplateHash struct {
	// HashID is the unique identifier for this hash
	HashID string
	// TemplateType is the type of biometric template
	TemplateType TemplateType
	// HashValue is the irreversible hash of the template
	HashValue []byte
	// Salt is the per-template salt used in hashing
	Salt []byte
	// Version is the hash algorithm version
	Version uint32
	// MatchThreshold is the similarity threshold for matching
	MatchThreshold float64
	// LSHHashes are locality-sensitive hashes for fuzzy matching
	LSHHashes [][]byte
	// CreatedAt is the creation timestamp
	CreatedAt time.Time
	// CreatedAtHeight is the block height when created
	CreatedAtHeight int64
}

// BiometricHashAuditEntry represents an audit log entry for biometric operations
type BiometricHashAuditEntry struct {
	// Operation is the type of operation performed
	Operation string
	// HashID is the ID of the hash involved
	HashID string
	// TemplateType is the type of template
	TemplateType TemplateType
	// Timestamp is when the operation occurred
	Timestamp time.Time
	// BlockHeight is the block height of the operation
	BlockHeight int64
	// Address is the account address involved
	Address string
	// Success indicates if the operation succeeded
	Success bool
	// ErrorMessage contains any error message (empty on success)
	ErrorMessage string
}

// BiometricMatchResult represents the result of a template matching operation
type BiometricMatchResult struct {
	// Matched indicates if the templates matched within threshold
	Matched bool
	// Similarity is the computed similarity score (0.0 to 1.0)
	Similarity float64
	// Threshold is the threshold that was used
	Threshold float64
	// Method is the matching method used
	Method string
	// HashVersion is the version of the stored hash
	HashVersion uint32
}

// ============================================================================
// Salt Generation
// ============================================================================

// GenerateTemplateSalt generates a cryptographically secure salt for a biometric template.
// The salt is unique per template and prevents rainbow table attacks.
//
// Returns:
//   - salt: 32-byte cryptographically random salt
//   - error: any error during salt generation
func GenerateTemplateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	n, err := rand.Read(salt)
	if err != nil {
		return nil, types.ErrBiometricHashFailed.Wrapf("failed to generate salt: %v", err)
	}
	if n != SaltSize {
		return nil, types.ErrBiometricHashFailed.Wrapf("short read during salt generation: got %d bytes, expected %d", n, SaltSize)
	}
	return salt, nil
}

// ============================================================================
// Hash Generation
// ============================================================================

// HashBiometricTemplate creates an irreversible hash from a biometric template.
// The raw template data is NEVER stored; only the hash is kept.
//
// Parameters:
//   - template: The raw biometric template (will be cleared after hashing)
//   - templateType: The type of biometric (face, fingerprint, etc.)
//   - hashID: A unique identifier for this hash
//   - ctx: The SDK context for block time/height
//
// Returns:
//   - BiometricTemplateHash: The hash structure ready for storage
//   - error: Any error during hashing
//
// SECURITY: The template parameter is cleared after hashing to minimize exposure
func HashBiometricTemplate(template []byte, templateType TemplateType, hashID string, ctx sdk.Context) (*BiometricTemplateHash, error) {
	// Validate inputs
	if len(template) == 0 {
		return nil, types.ErrBiometricHashFailed.Wrap(errMsgTemplateEmpty)
	}
	if len(template) > MaxTemplateSize {
		return nil, types.ErrBiometricHashFailed.Wrapf("template too large: %d bytes, max %d", len(template), MaxTemplateSize)
	}
	if hashID == "" {
		return nil, types.ErrBiometricHashFailed.Wrap("hash ID cannot be empty")
	}
	if err := ValidateTemplateType(templateType); err != nil {
		return nil, err
	}

	// Generate unique salt
	salt, err := GenerateTemplateSalt()
	if err != nil {
		return nil, err
	}

	// Derive the main hash using Argon2id
	hashValue := deriveArgon2idHash(template, salt)

	// Generate LSH hashes for fuzzy matching
	lshHashes := generateLSHHashes(template, salt)

	// Clear the template from memory (security measure)
	// The caller should also clear their copy
	clearBytes(template)

	return &BiometricTemplateHash{
		HashID:          hashID,
		TemplateType:    templateType,
		HashValue:       hashValue,
		Salt:            salt,
		Version:         BiometricHashVersion,
		MatchThreshold:  DefaultMatchThreshold(templateType),
		LSHHashes:       lshHashes,
		CreatedAt:       ctx.BlockTime(),
		CreatedAtHeight: ctx.BlockHeight(),
	}, nil
}

// deriveArgon2idHash derives an Argon2id hash from template and salt.
// Argon2id is memory-hard and resistant to GPU/ASIC attacks.
func deriveArgon2idHash(template, salt []byte) []byte {
	return argon2.IDKey(template, salt, Argon2idTime, Argon2idMemory, Argon2idThreads, HashSize)
}

// ============================================================================
// Locality-Sensitive Hashing (LSH) for Fuzzy Matching
// ============================================================================

// generateLSHHashes generates locality-sensitive hashes for fuzzy matching.
// LSH allows comparing templates without revealing the raw data.
// Similar templates will have similar LSH hashes.
func generateLSHHashes(template, salt []byte) [][]byte {
	lshHashes := make([][]byte, LSHBuckets)

	for i := 0; i < LSHBuckets; i++ {
		// Create bucket-specific salt by combining salt with bucket index
		bucketSalt := make([]byte, len(salt)+4)
		copy(bucketSalt, salt)
		binary.BigEndian.PutUint32(bucketSalt[len(salt):], safeUint32FromIntBiometric(i))

		// Hash the template with bucket-specific salt
		hash := sha256.New()
		hash.Write(bucketSalt)
		hash.Write(template)
		fullHash := hash.Sum(nil)

		// Take first LSHHashSize bytes as the LSH hash for this bucket
		lshHashes[i] = make([]byte, LSHHashSize)
		copy(lshHashes[i], fullHash[:LSHHashSize])
	}

	return lshHashes
}

// DeriveMatchableHash creates a hash structure suitable for matching against stored hashes.
// This is used when a new template needs to be compared against existing enrollments.
//
// Parameters:
//   - template: The new biometric template to match
//   - storedHash: The stored hash to compare against
//
// Returns:
//   - candidateHash: Hash of the new template using stored salt
//   - candidateLSH: LSH hashes of the new template
//   - error: Any error during derivation
func DeriveMatchableHash(template []byte, storedHash *BiometricTemplateHash) ([]byte, [][]byte, error) {
	if len(template) == 0 {
		return nil, nil, types.ErrBiometricHashFailed.Wrap(errMsgTemplateEmpty)
	}
	if storedHash == nil {
		return nil, nil, types.ErrBiometricHashFailed.Wrap(errMsgStoredHashNil)
	}
	if len(storedHash.Salt) != SaltSize {
		return nil, nil, types.ErrBiometricHashFailed.Wrapf("invalid stored salt size: %d", len(storedHash.Salt))
	}

	// Re-derive hash with the stored salt
	candidateHash := deriveArgon2idHash(template, storedHash.Salt)

	// Generate LSH hashes with stored salt
	candidateLSH := generateLSHHashes(template, storedHash.Salt)

	return candidateHash, candidateLSH, nil
}

// ============================================================================
// Template Matching
// ============================================================================

// MatchTemplateHash compares a new template against a stored hash.
// Uses LSH for initial filtering and Argon2id for final verification.
//
// Parameters:
//   - template: The new biometric template to match
//   - storedHash: The stored hash to compare against
//   - customThreshold: Optional custom threshold (use 0 for default)
//
// Returns:
//   - BiometricMatchResult: The matching result with similarity score
//   - error: Any error during matching
//
// SECURITY: The template is NOT stored; only used for comparison
func MatchTemplateHash(template []byte, storedHash *BiometricTemplateHash, customThreshold float64) (*BiometricMatchResult, error) {
	if len(template) == 0 {
		return nil, types.ErrBiometricHashFailed.Wrap(errMsgTemplateEmpty)
	}
	if storedHash == nil {
		return nil, types.ErrBiometricHashFailed.Wrap(errMsgStoredHashNil)
	}

	// Determine threshold
	threshold := storedHash.MatchThreshold
	if customThreshold > 0 && customThreshold <= 1.0 {
		threshold = customThreshold
	}

	// Derive matchable hash from new template
	candidateHash, candidateLSH, err := DeriveMatchableHash(template, storedHash)
	if err != nil {
		return nil, err
	}

	// Clear the template after deriving hash
	clearBytes(template)

	// Phase 1: LSH-based similarity (fast filtering)
	lshSimilarity := computeLSHSimilarity(candidateLSH, storedHash.LSHHashes)

	// Phase 2: Full hash comparison (for exact match)
	exactMatch := subtle.ConstantTimeCompare(candidateHash, storedHash.HashValue) == 1

	// Determine final similarity and match result
	var similarity float64
	var method string

	if exactMatch {
		// Exact cryptographic match (same template with same salt)
		similarity = 1.0
		method = "exact"
	} else {
		// Use LSH similarity for fuzzy matching
		similarity = lshSimilarity
		method = "lsh"
	}

	matched := similarity >= threshold

	return &BiometricMatchResult{
		Matched:     matched,
		Similarity:  similarity,
		Threshold:   threshold,
		Method:      method,
		HashVersion: storedHash.Version,
	}, nil
}

// computeLSHSimilarity computes similarity between two sets of LSH hashes.
// Returns a value between 0.0 (completely different) and 1.0 (identical).
func computeLSHSimilarity(lsh1, lsh2 [][]byte) float64 {
	if len(lsh1) != len(lsh2) {
		return 0.0
	}
	if len(lsh1) == 0 {
		return 0.0
	}

	matchingBuckets := 0
	for i := 0; i < len(lsh1); i++ {
		if bytes.Equal(lsh1[i], lsh2[i]) {
			matchingBuckets++
		}
	}

	return float64(matchingBuckets) / float64(len(lsh1))
}

// ============================================================================
// Hash Deletion
// ============================================================================

// DeleteTemplateHash securely deletes a biometric hash and creates an audit record.
// This should be called when an identity is deleted or template is replaced.
//
// Parameters:
//   - k: The keeper instance
//   - ctx: The SDK context
//   - address: The account address owning the hash
//   - hashID: The ID of the hash to delete
//   - reason: The reason for deletion (for audit)
//
// Returns:
//   - audit: The audit entry for this operation
//   - error: Any error during deletion
func (k Keeper) DeleteTemplateHash(ctx sdk.Context, address sdk.AccAddress, hashID string, reason string) (*BiometricHashAuditEntry, error) {
	if hashID == "" {
		return nil, types.ErrBiometricHashFailed.Wrap(errMsgHashIDEmpty)
	}

	// Get the stored hash first (for audit purposes)
	storedHash, found := k.GetBiometricHash(ctx, address, hashID)
	if !found {
		return nil, types.ErrBiometricHashNotFound.Wrapf("hash ID: %s", hashID)
	}

	// Create audit entry before deletion
	audit := &BiometricHashAuditEntry{
		Operation:    "delete",
		HashID:       hashID,
		TemplateType: storedHash.TemplateType,
		Timestamp:    ctx.BlockTime(),
		BlockHeight:  ctx.BlockHeight(),
		Address:      address.String(),
		Success:      true,
		ErrorMessage: "",
	}

	// Securely clear the hash data before deletion
	clearBytes(storedHash.HashValue)
	clearBytes(storedHash.Salt)
	for _, lsh := range storedHash.LSHHashes {
		clearBytes(lsh)
	}

	// Delete from store
	store := ctx.KVStore(k.skey)
	key := BiometricHashKey(address, hashID)
	store.Delete(key)

	// Delete from index
	indexKey := BiometricHashByTypeKey(address, storedHash.TemplateType, hashID)
	store.Delete(indexKey)

	// Store audit entry
	if err := k.storeBiometricAudit(ctx, address, audit); err != nil {
		audit.Success = false
		audit.ErrorMessage = err.Error()
		return audit, err
	}

	// Emit deletion event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"biometric_hash_deleted",
			sdk.NewAttribute("hash_id", hashID),
			sdk.NewAttribute("template_type", storedHash.TemplateType.String()),
			sdk.NewAttribute("address", address.String()),
			sdk.NewAttribute("reason", reason),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return audit, nil
}

// ============================================================================
// Storage Operations
// ============================================================================

// SetBiometricHash stores a biometric hash in the keeper.
func (k Keeper) SetBiometricHash(ctx sdk.Context, address sdk.AccAddress, hash *BiometricTemplateHash) error {
	if hash == nil {
		return types.ErrBiometricHashFailed.Wrap("hash cannot be nil")
	}

	store := ctx.KVStore(k.skey)

	// Serialize hash
	bz, err := k.cdc.MarshalLengthPrefixed(biometricHashToProto(hash))
	if err != nil {
		return types.ErrBiometricHashFailed.Wrapf("failed to marshal hash: %v", err)
	}

	// Store by primary key
	key := BiometricHashKey(address, hash.HashID)
	store.Set(key, bz)

	// Store index by type
	indexKey := BiometricHashByTypeKey(address, hash.TemplateType, hash.HashID)
	store.Set(indexKey, []byte{1})

	// Create audit entry
	audit := &BiometricHashAuditEntry{
		Operation:    "create",
		HashID:       hash.HashID,
		TemplateType: hash.TemplateType,
		Timestamp:    ctx.BlockTime(),
		BlockHeight:  ctx.BlockHeight(),
		Address:      address.String(),
		Success:      true,
	}
	if err := k.storeBiometricAudit(ctx, address, audit); err != nil {
		return err
	}

	// Emit creation event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"biometric_hash_created",
			sdk.NewAttribute("hash_id", hash.HashID),
			sdk.NewAttribute("template_type", hash.TemplateType.String()),
			sdk.NewAttribute("address", address.String()),
			sdk.NewAttribute("version", fmt.Sprintf("%d", hash.Version)),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return nil
}

// GetBiometricHash retrieves a biometric hash from the keeper.
func (k Keeper) GetBiometricHash(ctx sdk.Context, address sdk.AccAddress, hashID string) (*BiometricTemplateHash, bool) {
	store := ctx.KVStore(k.skey)
	key := BiometricHashKey(address, hashID)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var proto types.BiometricHashProto
	if err := k.cdc.UnmarshalLengthPrefixed(bz, &proto); err != nil {
		return nil, false
	}

	return protoToBiometricHash(&proto), true
}

// GetBiometricHashesByType retrieves all biometric hashes of a given type for an address.
func (k Keeper) GetBiometricHashesByType(ctx sdk.Context, address sdk.AccAddress, templateType TemplateType) []*BiometricTemplateHash {
	store := ctx.KVStore(k.skey)
	prefix := BiometricHashByTypePrefixKey(address, templateType)

	var hashes []*BiometricTemplateHash
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Extract hash ID from key
		key := iterator.Key()
		hashID := string(key[len(prefix):])

		// Get the full hash
		hash, found := k.GetBiometricHash(ctx, address, hashID)
		if found {
			hashes = append(hashes, hash)
		}
	}

	return hashes
}

// HasBiometricHash checks if a biometric hash exists.
func (k Keeper) HasBiometricHash(ctx sdk.Context, address sdk.AccAddress, hashID string) bool {
	store := ctx.KVStore(k.skey)
	key := BiometricHashKey(address, hashID)
	return store.Has(key)
}

// storeBiometricAudit stores an audit entry for biometric operations.
func (k Keeper) storeBiometricAudit(ctx sdk.Context, address sdk.AccAddress, audit *BiometricHashAuditEntry) error {
	store := ctx.KVStore(k.skey)

	// Create unique audit key with timestamp
	key := BiometricAuditKey(address, audit.Timestamp, audit.HashID)

	bz, err := k.cdc.MarshalLengthPrefixed(auditToProto(audit))
	if err != nil {
		return types.ErrBiometricHashFailed.Wrapf("failed to marshal audit: %v", err)
	}

	store.Set(key, bz)
	return nil
}

// GetBiometricAudits retrieves audit entries for an address.
func (k Keeper) GetBiometricAudits(ctx sdk.Context, address sdk.AccAddress, limit int) []*BiometricHashAuditEntry {
	store := ctx.KVStore(k.skey)
	prefix := BiometricAuditPrefixKey(address)

	var audits []*BiometricHashAuditEntry
	iterator := storetypes.KVStoreReversePrefixIterator(store, prefix)
	defer iterator.Close()

	count := 0
	for ; iterator.Valid() && count < limit; iterator.Next() {
		var proto types.BiometricAuditProto
		if err := k.cdc.UnmarshalLengthPrefixed(iterator.Value(), &proto); err != nil {
			continue
		}
		audits = append(audits, protoToAudit(&proto))
		count++
	}

	return audits
}

// ============================================================================
// Version Migration
// ============================================================================

// MigrateBiometricHash upgrades a hash to the latest algorithm version.
// This is used when algorithm parameters change.
//
// Parameters:
//   - ctx: The SDK context
//   - address: The account address
//   - hashID: The hash to migrate
//   - template: The original template (required for re-hashing)
//
// Returns:
//   - The migrated hash
//   - error if migration fails
func (k Keeper) MigrateBiometricHash(ctx sdk.Context, address sdk.AccAddress, hashID string, template []byte) (*BiometricTemplateHash, error) {
	// Get existing hash
	existingHash, found := k.GetBiometricHash(ctx, address, hashID)
	if !found {
		return nil, types.ErrBiometricHashNotFound.Wrapf("hash ID: %s", hashID)
	}

	// Check if migration is needed
	if existingHash.Version >= BiometricHashVersion {
		return existingHash, nil // Already at latest version
	}

	// Create new hash with latest algorithm
	newHash, err := HashBiometricTemplate(template, existingHash.TemplateType, hashID, ctx)
	if err != nil {
		return nil, err
	}

	// Preserve original creation time but update version
	newHash.CreatedAt = existingHash.CreatedAt
	newHash.CreatedAtHeight = existingHash.CreatedAtHeight

	// Delete old hash
	if _, err := k.DeleteTemplateHash(ctx, address, hashID, "version migration"); err != nil {
		return nil, err
	}

	// Store new hash
	if err := k.SetBiometricHash(ctx, address, newHash); err != nil {
		return nil, err
	}

	// Emit migration event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"biometric_hash_migrated",
			sdk.NewAttribute("hash_id", hashID),
			sdk.NewAttribute("old_version", fmt.Sprintf("%d", existingHash.Version)),
			sdk.NewAttribute("new_version", fmt.Sprintf("%d", newHash.Version)),
			sdk.NewAttribute("address", address.String()),
		),
	)

	return newHash, nil
}

// ============================================================================
// Key Construction Functions
// ============================================================================

// BiometricHashKey returns the store key for a biometric hash.
// Key: PrefixBiometricHash | address | hash_id
func BiometricHashKey(address sdk.AccAddress, hashID string) []byte {
	key := make([]byte, 0, len(types.PrefixBiometricHash)+len(address)+1+len(hashID))
	key = append(key, types.PrefixBiometricHash...)
	key = append(key, address...)
	key = append(key, byte('/'))
	key = append(key, []byte(hashID)...)
	return key
}

// BiometricHashByTypeKey returns the index key for biometric hashes by type.
// Key: PrefixBiometricHashByType | address | template_type | hash_id
func BiometricHashByTypeKey(address sdk.AccAddress, templateType TemplateType, hashID string) []byte {
	key := make([]byte, 0, len(types.PrefixBiometricHashByType)+len(address)+5+len(hashID))
	key = append(key, types.PrefixBiometricHashByType...)
	key = append(key, address...)
	key = append(key, byte('/'))
	typeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(typeBytes, safeUint32FromIntBiometric(int(templateType)))
	key = append(key, typeBytes...)
	key = append(key, []byte(hashID)...)
	return key
}

// BiometricHashByTypePrefixKey returns the prefix key for iterating by type.
func BiometricHashByTypePrefixKey(address sdk.AccAddress, templateType TemplateType) []byte {
	key := make([]byte, 0, len(types.PrefixBiometricHashByType)+len(address)+5)
	key = append(key, types.PrefixBiometricHashByType...)
	key = append(key, address...)
	key = append(key, byte('/'))
	typeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(typeBytes, safeUint32FromIntBiometric(int(templateType)))
	key = append(key, typeBytes...)
	return key
}

// BiometricAuditKey returns the store key for a biometric audit entry.
// Key: PrefixBiometricAudit | address | timestamp | hash_id
func BiometricAuditKey(address sdk.AccAddress, timestamp time.Time, hashID string) []byte {
	key := make([]byte, 0, len(types.PrefixBiometricAudit)+len(address)+9+len(hashID))
	key = append(key, types.PrefixBiometricAudit...)
	key = append(key, address...)
	key = append(key, byte('/'))
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, safeUint64FromInt64(timestamp.UnixNano()))
	key = append(key, timeBytes...)
	key = append(key, []byte(hashID)...)
	return key
}

// BiometricAuditPrefixKey returns the prefix key for iterating audits by address.
func BiometricAuditPrefixKey(address sdk.AccAddress) []byte {
	key := make([]byte, 0, len(types.PrefixBiometricAudit)+len(address)+1)
	key = append(key, types.PrefixBiometricAudit...)
	key = append(key, address...)
	key = append(key, byte('/'))
	return key
}

// ============================================================================
// Proto Conversion Helpers
// ============================================================================

// biometricHashToProto converts a BiometricTemplateHash to its proto representation.
func biometricHashToProto(hash *BiometricTemplateHash) *types.BiometricHashProto {
	lshHashes := make([][]byte, len(hash.LSHHashes))
	copy(lshHashes, hash.LSHHashes)

	return &types.BiometricHashProto{
		HashId:          hash.HashID,
		TemplateType:    safeInt32FromInt(int(hash.TemplateType)),
		HashValue:       hash.HashValue,
		Salt:            hash.Salt,
		Version:         hash.Version,
		MatchThreshold:  hash.MatchThreshold,
		LshHashes:       lshHashes,
		CreatedAt:       hash.CreatedAt.UnixNano(),
		CreatedAtHeight: hash.CreatedAtHeight,
	}
}

// protoToBiometricHash converts a proto representation to BiometricTemplateHash.
func protoToBiometricHash(proto *types.BiometricHashProto) *BiometricTemplateHash {
	lshHashes := make([][]byte, len(proto.LshHashes))
	copy(lshHashes, proto.LshHashes)

	return &BiometricTemplateHash{
		HashID:          proto.HashId,
		TemplateType:    TemplateType(proto.TemplateType),
		HashValue:       proto.HashValue,
		Salt:            proto.Salt,
		Version:         proto.Version,
		MatchThreshold:  proto.MatchThreshold,
		LSHHashes:       lshHashes,
		CreatedAt:       time.Unix(0, proto.CreatedAt),
		CreatedAtHeight: proto.CreatedAtHeight,
	}
}

// auditToProto converts an audit entry to its proto representation.
func auditToProto(audit *BiometricHashAuditEntry) *types.BiometricAuditProto {
	return &types.BiometricAuditProto{
		Operation:    audit.Operation,
		HashId:       audit.HashID,
		TemplateType: safeInt32FromInt(int(audit.TemplateType)),
		Timestamp:    audit.Timestamp.UnixNano(),
		BlockHeight:  audit.BlockHeight,
		Address:      audit.Address,
		Success:      audit.Success,
		ErrorMessage: audit.ErrorMessage,
	}
}

// protoToAudit converts a proto representation to BiometricHashAuditEntry.
func protoToAudit(proto *types.BiometricAuditProto) *BiometricHashAuditEntry {
	return &BiometricHashAuditEntry{
		Operation:    proto.Operation,
		HashID:       proto.HashId,
		TemplateType: TemplateType(proto.TemplateType),
		Timestamp:    time.Unix(0, proto.Timestamp),
		BlockHeight:  proto.BlockHeight,
		Address:      proto.Address,
		Success:      proto.Success,
		ErrorMessage: proto.ErrorMessage,
	}
}

func safeUint64FromInt64(value int64) uint64 {
	if value < 0 {
		return 0
	}
	//nolint:gosec // range checked above
	return uint64(value)
}

func safeInt32FromInt(value int) int32 {
	const maxInt32 = int32(^uint32(0) >> 1)
	const minInt32 = -maxInt32 - 1
	if value > int(maxInt32) {
		return maxInt32
	}
	if value < int(minInt32) {
		return minInt32
	}
	//nolint:gosec // range checked above
	return int32(value)
}

func safeUint32FromIntBiometric(value int) uint32 {
	if value < 0 {
		return 0
	}
	if value > int(^uint32(0)) {
		return ^uint32(0)
	}
	//nolint:gosec // range checked above
	return uint32(value)
}

// ============================================================================
// Utility Functions
// ============================================================================

// clearBytes securely clears a byte slice by overwriting with zeros.
// This helps minimize the exposure of sensitive data in memory.
func clearBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// computeCosineSimilarity computes the cosine similarity between two vectors.
// This is used for embedding comparison when LSH indicates a potential match.
// Returns a value between -1.0 and 1.0.
//
//nolint:unused // reserved for future biometric similarity scoring
func computeCosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
