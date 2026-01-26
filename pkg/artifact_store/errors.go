package artifact_store

import (
	"cosmossdk.io/errors"
)

// Error codes for artifact store module
var (
	// ErrInvalidContentAddress is returned when a content address is invalid
	ErrInvalidContentAddress = errors.Register("artifact_store", 1, "invalid content address")

	// ErrInvalidChunkManifest is returned when a chunk manifest is invalid
	ErrInvalidChunkManifest = errors.Register("artifact_store", 2, "invalid chunk manifest")

	// ErrInvalidArtifactReference is returned when an artifact reference is invalid
	ErrInvalidArtifactReference = errors.Register("artifact_store", 3, "invalid artifact reference")

	// ErrArtifactNotFound is returned when an artifact is not found
	ErrArtifactNotFound = errors.Register("artifact_store", 4, "artifact not found")

	// ErrChunkNotFound is returned when a chunk is not found
	ErrChunkNotFound = errors.Register("artifact_store", 5, "chunk not found")

	// ErrBackendUnavailable is returned when the backend is unavailable
	ErrBackendUnavailable = errors.Register("artifact_store", 6, "backend unavailable")

	// ErrAuthenticationFailed is returned when authentication fails
	ErrAuthenticationFailed = errors.Register("artifact_store", 7, "authentication failed")

	// ErrStorageLimitExceeded is returned when storage limits are exceeded
	ErrStorageLimitExceeded = errors.Register("artifact_store", 8, "storage limit exceeded")

	// ErrRetentionExpired is returned when the artifact retention has expired
	ErrRetentionExpired = errors.Register("artifact_store", 9, "retention expired")

	// ErrHashMismatch is returned when the content hash doesn't match
	ErrHashMismatch = errors.Register("artifact_store", 10, "content hash mismatch")

	// ErrInvalidEncryption is returned when encryption is invalid
	ErrInvalidEncryption = errors.Register("artifact_store", 11, "invalid encryption")

	// ErrChunkReassemblyFailed is returned when chunk reassembly fails
	ErrChunkReassemblyFailed = errors.Register("artifact_store", 12, "chunk reassembly failed")

	// ErrBackendNotSupported is returned when the backend type is not supported
	ErrBackendNotSupported = errors.Register("artifact_store", 13, "backend not supported")

	// ErrInvalidInput is returned when input parameters are invalid
	ErrInvalidInput = errors.Register("artifact_store", 14, "invalid input")

	// ErrAlreadyExists is returned when an artifact already exists
	ErrAlreadyExists = errors.Register("artifact_store", 15, "artifact already exists")
)
