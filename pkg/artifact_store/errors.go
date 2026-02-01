package artifact_store

import (
	"cosmossdk.io/errors"
)

// Error codes for artifact store module
var (
	// ErrInvalidContentAddress is returned when a content address is invalid
	ErrInvalidContentAddress = errors.Register("artifact_store", 2500, "invalid content address")

	// ErrInvalidChunkManifest is returned when a chunk manifest is invalid
	ErrInvalidChunkManifest = errors.Register("artifact_store", 2501, "invalid chunk manifest")

	// ErrInvalidArtifactReference is returned when an artifact reference is invalid
	ErrInvalidArtifactReference = errors.Register("artifact_store", 2502, "invalid artifact reference")

	// ErrArtifactNotFound is returned when an artifact is not found
	ErrArtifactNotFound = errors.Register("artifact_store", 2503, "artifact not found")

	// ErrChunkNotFound is returned when a chunk is not found
	ErrChunkNotFound = errors.Register("artifact_store", 2504, "chunk not found")

	// ErrBackendUnavailable is returned when the backend is unavailable
	ErrBackendUnavailable = errors.Register("artifact_store", 2505, "backend unavailable")

	// ErrAuthenticationFailed is returned when authentication fails
	ErrAuthenticationFailed = errors.Register("artifact_store", 2506, "authentication failed")

	// ErrStorageLimitExceeded is returned when storage limits are exceeded
	ErrStorageLimitExceeded = errors.Register("artifact_store", 2507, "storage limit exceeded")

	// ErrRetentionExpired is returned when the artifact retention has expired
	ErrRetentionExpired = errors.Register("artifact_store", 2508, "retention expired")

	// ErrHashMismatch is returned when the content hash doesn't match
	ErrHashMismatch = errors.Register("artifact_store", 2509, "content hash mismatch")

	// ErrInvalidEncryption is returned when encryption is invalid
	ErrInvalidEncryption = errors.Register("artifact_store", 2510, "invalid encryption")

	// ErrChunkReassemblyFailed is returned when chunk reassembly fails
	ErrChunkReassemblyFailed = errors.Register("artifact_store", 2511, "chunk reassembly failed")

	// ErrBackendNotSupported is returned when the backend type is not supported
	ErrBackendNotSupported = errors.Register("artifact_store", 2512, "backend not supported")

	// ErrInvalidInput is returned when input parameters are invalid
	ErrInvalidInput = errors.Register("artifact_store", 2513, "invalid input")

	// ErrAlreadyExists is returned when an artifact already exists
	ErrAlreadyExists = errors.Register("artifact_store", 2514, "artifact already exists")

	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.Register("artifact_store", 2515, "resource not found")

	// ErrInvalidState is returned when an operation is invalid for the current state
	ErrInvalidState = errors.Register("artifact_store", 2516, "invalid state")
)

