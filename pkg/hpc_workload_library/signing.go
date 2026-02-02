// Package hpc_workload_library provides template signing functionality.
//
// VE-5F: Template signing and versioning with artifact store references
package hpc_workload_library

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/virtengine/virtengine/pkg/artifact_store"
	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// SigningAlgorithm defines supported signing algorithms
type SigningAlgorithm string

const (
	// SigningAlgorithmEd25519 uses Ed25519 signatures
	SigningAlgorithmEd25519 SigningAlgorithm = "ed25519"

	// SigningAlgorithmSecp256k1 uses secp256k1 signatures (Cosmos SDK default)
	SigningAlgorithmSecp256k1 SigningAlgorithm = "secp256k1"
)

// TemplateSigner handles template signing operations
type TemplateSigner struct {
	algorithm  SigningAlgorithm
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewTemplateSigner creates a new template signer with Ed25519 key
func NewTemplateSigner(privateKey ed25519.PrivateKey) *TemplateSigner {
	return &TemplateSigner{
		algorithm:  SigningAlgorithmEd25519,
		privateKey: privateKey,
		publicKey:  privateKey.Public().(ed25519.PublicKey),
	}
}

// Sign signs a workload template and returns the signature
func (s *TemplateSigner) Sign(template *hpctypes.WorkloadTemplate) (*hpctypes.WorkloadSignature, error) {
	// Create a copy without signature for hashing
	templateCopy := *template
	templateCopy.Signature = hpctypes.WorkloadSignature{}

	// Serialize template to JSON for signing
	data, err := json.Marshal(templateCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template: %w", err)
	}

	// Compute content hash
	hash := sha256.Sum256(data)

	// Sign the hash
	signature := ed25519.Sign(s.privateKey, hash[:])

	return &hpctypes.WorkloadSignature{
		Algorithm:       string(s.algorithm),
		PublisherPubKey: hex.EncodeToString(s.publicKey),
		Signature:       hex.EncodeToString(signature),
		SignedAt:        time.Now().UTC(),
		ContentHash:     hex.EncodeToString(hash[:]),
	}, nil
}

// SignTemplate signs a template in place
func (s *TemplateSigner) SignTemplate(template *hpctypes.WorkloadTemplate) error {
	sig, err := s.Sign(template)
	if err != nil {
		return err
	}
	template.Signature = *sig
	return nil
}

// GetPublicKey returns the signer's public key
func (s *TemplateSigner) GetPublicKey() string {
	return hex.EncodeToString(s.publicKey)
}

// TemplateVerifier verifies template signatures
type TemplateVerifier struct{}

// NewTemplateVerifier creates a new template verifier
func NewTemplateVerifier() *TemplateVerifier {
	return &TemplateVerifier{}
}

// Verify verifies a template signature
func (v *TemplateVerifier) Verify(template *hpctypes.WorkloadTemplate) error {
	if template.Signature.Signature == "" {
		return fmt.Errorf("template has no signature")
	}

	// Decode public key
	pubKeyBytes, err := hex.DecodeString(template.Signature.PublisherPubKey)
	if err != nil {
		return fmt.Errorf("invalid public key encoding: %w", err)
	}

	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size")
	}

	// Decode signature
	sigBytes, err := hex.DecodeString(template.Signature.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}

	// Decode content hash
	expectedHash, err := hex.DecodeString(template.Signature.ContentHash)
	if err != nil {
		return fmt.Errorf("invalid content hash encoding: %w", err)
	}

	// Create a copy without signature for hashing
	templateCopy := *template
	templateCopy.Signature = hpctypes.WorkloadSignature{}

	// Serialize template to JSON
	data, err := json.Marshal(templateCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	// Compute content hash
	hash := sha256.Sum256(data)

	// Verify content hash matches
	if hex.EncodeToString(hash[:]) != hex.EncodeToString(expectedHash) {
		return fmt.Errorf("content hash mismatch: template has been modified")
	}

	// Verify signature
	if !ed25519.Verify(pubKeyBytes, hash[:], sigBytes) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// TemplateVersionManager manages template versions
type TemplateVersionManager struct {
	store artifact_store.ArtifactStore
}

// NewTemplateVersionManager creates a new version manager
func NewTemplateVersionManager(store artifact_store.ArtifactStore) *TemplateVersionManager {
	return &TemplateVersionManager{
		store: store,
	}
}

// PublishTemplate publishes a template to the artifact store
func (m *TemplateVersionManager) PublishTemplate(ctx context.Context, template *hpctypes.WorkloadTemplate, owner string) (*artifact_store.ContentAddress, error) {
	// Validate template before publishing
	if err := template.Validate(); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	// Serialize template
	data, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize template: %w", err)
	}

	// Compute content hash
	hash := sha256.Sum256(data)

	// Store in artifact store
	req := &artifact_store.PutRequest{
		Data:        data,
		ContentHash: hash[:],
		EncryptionMetadata: &artifact_store.EncryptionMetadata{
			AlgorithmID: "none", // Templates are public (not encrypted)
		},
		Owner:        owner,
		ArtifactType: "workload_template",
		Metadata: map[string]string{
			"template_id": template.TemplateID,
			"version":     template.Version,
			"type":        string(template.Type),
		},
	}

	resp, err := m.store.Put(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to store template: %w", err)
	}

	// Update template with artifact CID (use BackendRef as identifier)
	template.ArtifactCID = resp.ContentAddress.BackendRef

	return resp.ContentAddress, nil
}

// GetTemplate retrieves a template from the artifact store by its backend reference (CID)
func (m *TemplateVersionManager) GetTemplate(ctx context.Context, backendRef string) (*hpctypes.WorkloadTemplate, error) {
	// Create a minimal content address to retrieve by backend ref
	// The actual hash will be verified by the store
	addr := &artifact_store.ContentAddress{
		Version:    1,
		Algorithm:  "sha256",
		Backend:    artifact_store.BackendIPFS,
		BackendRef: backendRef,
	}

	resp, err := m.store.Get(ctx, &artifact_store.GetRequest{
		ContentAddress: addr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve template: %w", err)
	}

	var template hpctypes.WorkloadTemplate
	if err := json.Unmarshal(resp.Data, &template); err != nil {
		return nil, fmt.Errorf("failed to deserialize template: %w", err)
	}

	return &template, nil
}

// ListVersions lists all versions of a template
func (m *TemplateVersionManager) ListVersions(ctx context.Context, owner, templateID string) ([]*hpctypes.WorkloadTemplate, error) {
	resp, err := m.store.ListByOwner(ctx, owner, &artifact_store.Pagination{
		Limit: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	var templates []*hpctypes.WorkloadTemplate
	for _, ref := range resp.References {
		if ref.ArtifactType != "workload_template" {
			continue
		}

		// Get template and check ID
		template, err := m.GetTemplate(ctx, ref.ContentAddress.BackendRef)
		if err != nil {
			continue // Skip templates that can't be retrieved
		}

		if template.TemplateID == templateID {
			templates = append(templates, template)
		}
	}

	return templates, nil
}

// TemplateManifestBuilder builds complete workload manifests
type TemplateManifestBuilder struct {
	signer   *TemplateSigner
	verifier *TemplateVerifier
}

// NewTemplateManifestBuilder creates a new manifest builder
func NewTemplateManifestBuilder(signer *TemplateSigner) *TemplateManifestBuilder {
	return &TemplateManifestBuilder{
		signer:   signer,
		verifier: NewTemplateVerifier(),
	}
}

// BuildManifest creates a complete workload manifest
func (b *TemplateManifestBuilder) BuildManifest(template *hpctypes.WorkloadTemplate) (*hpctypes.WorkloadManifest, error) {
	// Sign template if signer is available
	if b.signer != nil {
		if err := b.signer.SignTemplate(template); err != nil {
			return nil, fmt.Errorf("failed to sign template: %w", err)
		}
	}

	// Serialize manifest
	data, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize template: %w", err)
	}

	// Compute checksum
	hash := sha256.Sum256(data)

	return &hpctypes.WorkloadManifest{
		SchemaVersion: "1.0.0",
		Template:      *template,
		Checksum:      hex.EncodeToString(hash[:]),
	}, nil
}

// VerifyManifest verifies a workload manifest
func (b *TemplateManifestBuilder) VerifyManifest(manifest *hpctypes.WorkloadManifest) error {
	// Verify checksum
	data, err := json.Marshal(manifest.Template)
	if err != nil {
		return fmt.Errorf("failed to serialize template: %w", err)
	}

	hash := sha256.Sum256(data)
	if hex.EncodeToString(hash[:]) != manifest.Checksum {
		return fmt.Errorf("manifest checksum mismatch")
	}

	// Verify template signature if present
	if manifest.Template.Signature.Signature != "" {
		if err := b.verifier.Verify(&manifest.Template); err != nil {
			return fmt.Errorf("signature verification failed: %w", err)
		}
	}

	return nil
}
