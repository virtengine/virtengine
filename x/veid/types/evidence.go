package types

import (
	"encoding/hex"
	"fmt"
)

func isValidSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validateEvidencePointer(hash, backend, ref string, require bool) error {
	if hash == "" && backend == "" && ref == "" {
		if require {
			return fmt.Errorf("evidence_hash and evidence_storage_ref are required for verified records")
		}
		return nil
	}

	if hash == "" {
		return fmt.Errorf("evidence_hash cannot be empty")
	}
	if !isValidSHA256Hex(hash) {
		return fmt.Errorf("evidence_hash must be a valid SHA256 hex string")
	}
	if ref == "" {
		return fmt.Errorf("evidence_storage_ref cannot be empty")
	}
	if backend != "" && !IsValidStorageBackend(StorageBackend(backend)) {
		return fmt.Errorf("invalid evidence_storage_backend: %s", backend)
	}
	return nil
}
