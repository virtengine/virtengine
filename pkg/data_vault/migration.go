package data_vault

import (
	"context"
	"time"
)

// PlaintextArtifact describes an existing plaintext artifact to migrate.
type PlaintextArtifact struct {
	Scope           Scope
	Owner           string
	OrgID           string
	RetentionPolicy string
	ExpiresAt       *time.Time
	Tags            map[string]string
	Data            []byte
}

// MigrationResult captures migration output.
type MigrationResult struct {
	BlobID BlobID
	Error  error
}

// MigratePlaintext encrypts and uploads a plaintext artifact into the vault.
func MigratePlaintext(ctx context.Context, vault VaultService, artifact PlaintextArtifact) (*EncryptedBlob, error) {
	if vault == nil {
		return nil, NewVaultError("MigratePlaintext", ErrInvalidRequest, "vault required")
	}
	if len(artifact.Data) == 0 {
		return nil, NewVaultError("MigratePlaintext", ErrInvalidRequest, "artifact data required")
	}

	req := &UploadRequest{
		Scope:           artifact.Scope,
		Plaintext:       artifact.Data,
		Owner:           artifact.Owner,
		OrgID:           artifact.OrgID,
		RetentionPolicy: artifact.RetentionPolicy,
		ExpiresAt:       artifact.ExpiresAt,
		Tags:            artifact.Tags,
	}

	return vault.Upload(ctx, req)
}
