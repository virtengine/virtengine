package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	_ "embed"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
)

//go:embed age_vk.bin
var ageVKBytes []byte

//go:embed age_vk.bin.sha256
var ageVKChecksum []byte

//go:embed residency_vk.bin
var residencyVKBytes []byte

//go:embed residency_vk.bin.sha256
var residencyVKChecksum []byte

//go:embed score_vk.bin
var scoreVKBytes []byte

//go:embed score_vk.bin.sha256
var scoreVKChecksum []byte

//go:embed params_metadata.json
var metadataBytes []byte

//go:embed params_metadata.json.sha256
var metadataChecksum []byte

// Metadata describes the trusted setup outputs embedded with the binary.
type Metadata struct {
	SchemaVersion  string                     `json:"schema_version"`
	ParameterSetID string                     `json:"parameter_set_id"`
	GeneratedAt    string                     `json:"generated_at"`
	ArtifactFormat string                     `json:"artifact_format"`
	Notes          []string                   `json:"notes,omitempty"`
	Circuits       map[string]CircuitMetadata `json:"circuits"`
}

// CircuitMetadata describes the ceremony outputs bound to a single circuit.
type CircuitMetadata struct {
	CircuitName              string   `json:"circuit_name"`
	ParametersVersion        string   `json:"parameters_version"`
	CeremonyID               string   `json:"ceremony_id"`
	ArtifactManifestHash     string   `json:"artifact_manifest_hash"`
	VerificationReportHash   string   `json:"verification_report_hash"`
	TranscriptHash           string   `json:"transcript_hash"`
	CircuitHash              string   `json:"circuit_hash"`
	VerificationKeyFile      string   `json:"verification_key_file"`
	VerificationKeyHash      string   `json:"verification_key_hash"`
	VerificationKeySizeBytes int64    `json:"verification_key_size_bytes"`
	ContributorCount         int      `json:"contributor_count"`
	Contributors             []string `json:"contributors"`
	CoordinatorID            string   `json:"coordinator_id"`
	CoordinatorPublicKey     string   `json:"coordinator_public_key"`
	Beacon                   string   `json:"beacon"`
}

// GetVerifyingKey returns the checksum-verified verifying key for the named circuit.
// Valid circuit names: age, residency, score.
func GetVerifyingKey(name string) (groth16.VerifyingKey, error) {
	artifact, err := loadArtifact(name)
	if err != nil {
		return nil, err
	}
	return decodeVerifyingKey(artifact.keyBytes)
}

// GetVerifiedVerifyingKey returns a checksum-verified verifying key that is also
// bound to the provided compiled circuit hash.
func GetVerifiedVerifyingKey(name string, compiled interface {
	WriteTo(io.Writer) (int64, error)
}) (groth16.VerifyingKey, error) {
	artifact, err := loadArtifact(name)
	if err != nil {
		return nil, err
	}
	if err := verifyCompiledCircuitHash(artifact.metadata, compiled); err != nil {
		return nil, err
	}
	return decodeVerifyingKey(artifact.keyBytes)
}

// GetMetadata returns validated metadata describing the embedded or overridden ceremony output.
func GetMetadata() (*Metadata, error) {
	artifactSet, err := loadArtifactSet(resolveArtifactSource())
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(artifactSet.metadata)
	if err != nil {
		return nil, err
	}
	var copied Metadata
	if err := json.Unmarshal(data, &copied); err != nil {
		return nil, err
	}
	return &copied, nil
}

func decodeVerifyingKey(data []byte) (groth16.VerifyingKey, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty verifying key data")
	}

	vk := groth16.NewVerifyingKey(ecc.BN254)
	if _, err := vk.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("parse verifying key: %w", err)
	}

	return vk, nil
}
