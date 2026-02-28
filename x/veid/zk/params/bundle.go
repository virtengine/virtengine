package params

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	// ParamsDirEnv allows operators and tests to stage a verified parameter directory
	// without rebuilding the binary.
	ParamsDirEnv = "VEID_ZK_PARAMS_DIR"

	metadataFileName       = "params_metadata.json"
	metadataDigestFileName = "params_metadata.json.sha256"
	metadataSchemaVersion  = "virtengine.veid.zkparams/v1"
	artifactFormatVersion  = "virtengine.trusted_setup.artifact_manifest/v1"
)

var (
	requiredCircuits = map[string]struct {
		CircuitName string
		FileName    string
	}{
		"age": {
			CircuitName: "age-range",
			FileName:    "age_vk.bin",
		},
		"residency": {
			CircuitName: "residency",
			FileName:    "residency_vk.bin",
		},
		"score": {
			CircuitName: "score-range",
			FileName:    "score_vk.bin",
		},
	}
	placeholderPattern = regexp.MustCompile(`(?i)(placeholder|pending|tbd|not published yet|single-party|dev-setup)`)
)

type artifactSet struct {
	metadata *Metadata
	keys     map[string][]byte
}

type circuitArtifact struct {
	metadata *CircuitMetadata
	keyBytes []byte
}

type artifactSource struct {
	dir string
}

func resolveArtifactSource() artifactSource {
	return artifactSource{dir: strings.TrimSpace(os.Getenv(ParamsDirEnv))}
}

func (s artifactSource) description() string {
	if s.dir == "" {
		return "embedded VEID ZK params"
	}
	return fmt.Sprintf("VEID ZK params directory %q", s.dir)
}

func (s artifactSource) readFile(name string) ([]byte, error) {
	if s.dir != "" {
		path := filepath.Join(s.dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s from %s: %w", name, s.description(), err)
		}
		return data, nil
	}

	switch name {
	case "age_vk.bin":
		return slices.Clone(ageVKBytes), nil
	case "age_vk.bin.sha256":
		return slices.Clone(ageVKChecksum), nil
	case "residency_vk.bin":
		return slices.Clone(residencyVKBytes), nil
	case "residency_vk.bin.sha256":
		return slices.Clone(residencyVKChecksum), nil
	case "score_vk.bin":
		return slices.Clone(scoreVKBytes), nil
	case "score_vk.bin.sha256":
		return slices.Clone(scoreVKChecksum), nil
	case metadataFileName:
		return slices.Clone(metadataBytes), nil
	case metadataDigestFileName:
		return slices.Clone(metadataChecksum), nil
	default:
		return nil, fmt.Errorf("unknown artifact %q", name)
	}
}

func loadArtifact(name string) (*circuitArtifact, error) {
	artifactSet, err := loadArtifactSet(resolveArtifactSource())
	if err != nil {
		return nil, err
	}
	circuitMeta, ok := artifactSet.metadata.Circuits[name]
	if !ok {
		return nil, fmt.Errorf("unsupported VEID circuit %q", name)
	}
	keyBytes, ok := artifactSet.keys[name]
	if !ok {
		return nil, fmt.Errorf("verified key bytes missing for circuit %q", name)
	}
	copiedMeta := circuitMeta
	return &circuitArtifact{
		metadata: &copiedMeta,
		keyBytes: slices.Clone(keyBytes),
	}, nil
}

func loadArtifactSet(source artifactSource) (*artifactSet, error) {
	rawMetadata, err := source.readFile(metadataFileName)
	if err != nil {
		return nil, err
	}
	rawMetadataDigest, err := source.readFile(metadataDigestFileName)
	if err != nil {
		return nil, err
	}
	if err := verifyDigestRecord(metadataFileName, rawMetadata, rawMetadataDigest); err != nil {
		return nil, fmt.Errorf("%s: %w", source.description(), err)
	}
	if placeholderPattern.Match(rawMetadata) {
		return nil, fmt.Errorf("%s: params metadata contains placeholder content", source.description())
	}

	var meta Metadata
	if err := json.Unmarshal(rawMetadata, &meta); err != nil {
		return nil, fmt.Errorf("%s: parse params metadata: %w", source.description(), err)
	}
	if err := validateMetadata(&meta); err != nil {
		return nil, fmt.Errorf("%s: %w", source.description(), err)
	}

	keys := make(map[string][]byte, len(requiredCircuits))
	for circuitKey, required := range requiredCircuits {
		circuitMeta := meta.Circuits[circuitKey]
		keyBytes, err := source.readFile(required.FileName)
		if err != nil {
			return nil, err
		}
		digestBytes, err := source.readFile(required.FileName + ".sha256")
		if err != nil {
			return nil, err
		}
		if err := verifyDigestRecord(required.FileName, keyBytes, digestBytes); err != nil {
			return nil, fmt.Errorf("%s: %w", source.description(), err)
		}
		if actualHash := hashBytes(keyBytes); actualHash != circuitMeta.VerificationKeyHash {
			return nil, fmt.Errorf("%s: verification key hash mismatch for %s: expected %s got %s", source.description(), circuitKey, circuitMeta.VerificationKeyHash, actualHash)
		}
		if actualSize := int64(len(keyBytes)); actualSize != circuitMeta.VerificationKeySizeBytes {
			return nil, fmt.Errorf("%s: verification key size mismatch for %s: expected %d got %d", source.description(), circuitKey, circuitMeta.VerificationKeySizeBytes, actualSize)
		}
		if _, err := decodeVerifyingKey(keyBytes); err != nil {
			return nil, fmt.Errorf("%s: invalid verifying key for %s: %w", source.description(), circuitKey, err)
		}
		keys[circuitKey] = keyBytes
	}

	return &artifactSet{
		metadata: &meta,
		keys:     keys,
	}, nil
}

func validateMetadata(meta *Metadata) error {
	if meta == nil {
		return fmt.Errorf("params metadata is required")
	}
	if meta.SchemaVersion != metadataSchemaVersion {
		return fmt.Errorf("unsupported params metadata schema %q", meta.SchemaVersion)
	}
	if meta.ArtifactFormat != artifactFormatVersion {
		return fmt.Errorf("unsupported trusted setup artifact format %q", meta.ArtifactFormat)
	}
	if err := validateHumanString("parameter_set_id", meta.ParameterSetID); err != nil {
		return err
	}
	if _, err := time.Parse(time.RFC3339, meta.GeneratedAt); err != nil {
		return fmt.Errorf("generated_at must be RFC3339: %w", err)
	}
	if len(meta.Circuits) != len(requiredCircuits) {
		return fmt.Errorf("expected %d circuits in params metadata, got %d", len(requiredCircuits), len(meta.Circuits))
	}
	for _, note := range meta.Notes {
		if err := validateHumanString("note", note); err != nil {
			return err
		}
	}

	for circuitKey, required := range requiredCircuits {
		circuitMeta, ok := meta.Circuits[circuitKey]
		if !ok {
			return fmt.Errorf("params metadata missing circuit %q", circuitKey)
		}
		if circuitMeta.CircuitName != required.CircuitName {
			return fmt.Errorf("circuit %s expected circuit_name %q, got %q", circuitKey, required.CircuitName, circuitMeta.CircuitName)
		}
		if circuitMeta.VerificationKeyFile != required.FileName {
			return fmt.Errorf("circuit %s expected verification_key_file %q, got %q", circuitKey, required.FileName, circuitMeta.VerificationKeyFile)
		}
		if err := validateCircuitMetadata(circuitKey, &circuitMeta); err != nil {
			return err
		}
	}
	return nil
}

func validateCircuitMetadata(circuitKey string, meta *CircuitMetadata) error {
	if meta == nil {
		return fmt.Errorf("circuit metadata missing for %s", circuitKey)
	}
	requiredHumanFields := map[string]string{
		"parameters_version": meta.ParametersVersion,
		"ceremony_id":        meta.CeremonyID,
		"coordinator_id":     meta.CoordinatorID,
		"beacon":             meta.Beacon,
	}
	for field, value := range requiredHumanFields {
		if err := validateHumanString(field, value); err != nil {
			return fmt.Errorf("circuit %s: %w", circuitKey, err)
		}
	}
	requiredHashFields := map[string]string{
		"artifact_manifest_hash":   meta.ArtifactManifestHash,
		"verification_report_hash": meta.VerificationReportHash,
		"transcript_hash":          meta.TranscriptHash,
		"circuit_hash":             meta.CircuitHash,
		"verification_key_hash":    meta.VerificationKeyHash,
	}
	for field, value := range requiredHashFields {
		if err := validateSHA256(field, value); err != nil {
			return fmt.Errorf("circuit %s: %w", circuitKey, err)
		}
	}
	if meta.VerificationKeySizeBytes <= 0 {
		return fmt.Errorf("circuit %s: verification_key_size_bytes must be positive", circuitKey)
	}
	if meta.ContributorCount < 2 {
		return fmt.Errorf("circuit %s: contributor_count must be at least 2", circuitKey)
	}
	if len(meta.Contributors) != meta.ContributorCount {
		return fmt.Errorf("circuit %s: contributor_count %d does not match contributors list length %d", circuitKey, meta.ContributorCount, len(meta.Contributors))
	}
	seen := make(map[string]struct{}, len(meta.Contributors))
	for _, contributor := range meta.Contributors {
		if err := validateHumanString("contributor", contributor); err != nil {
			return fmt.Errorf("circuit %s: %w", circuitKey, err)
		}
		if _, exists := seen[contributor]; exists {
			return fmt.Errorf("circuit %s: duplicate contributor %q", circuitKey, contributor)
		}
		seen[contributor] = struct{}{}
	}
	if _, err := base64.StdEncoding.DecodeString(meta.CoordinatorPublicKey); err != nil {
		return fmt.Errorf("circuit %s: coordinator_public_key is not valid base64: %w", circuitKey, err)
	}
	return nil
}

func validateHumanString(field, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s cannot be empty", field)
	}
	if strings.EqualFold(trimmed, "dev") {
		return fmt.Errorf("%s cannot use placeholder value %q", field, trimmed)
	}
	if placeholderPattern.MatchString(trimmed) {
		return fmt.Errorf("%s contains placeholder content %q", field, trimmed)
	}
	return nil
}

func validateSHA256(field, value string) error {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) != 64 {
		return fmt.Errorf("%s must be a 64-character SHA-256 hex digest", field)
	}
	if _, err := hex.DecodeString(trimmed); err != nil {
		return fmt.Errorf("%s must be valid hex: %w", field, err)
	}
	return nil
}

func verifyDigestRecord(expectedFile string, data, digestBytes []byte) error {
	parts := strings.Fields(string(digestBytes))
	if len(parts) != 2 {
		return fmt.Errorf("invalid checksum format for %s", expectedFile)
	}
	if parts[1] != expectedFile {
		return fmt.Errorf("checksum label mismatch for %s: expected %s got %s", expectedFile, expectedFile, parts[1])
	}
	if err := validateSHA256(expectedFile+" checksum", parts[0]); err != nil {
		return err
	}
	actualHash := hashBytes(data)
	if parts[0] != actualHash {
		return fmt.Errorf("checksum mismatch for %s: expected %s got %s", expectedFile, parts[0], actualHash)
	}
	return nil
}

func verifyCompiledCircuitHash(meta *CircuitMetadata, compiled interface {
	WriteTo(io.Writer) (int64, error)
}) error {
	if meta == nil {
		return fmt.Errorf("circuit metadata is required")
	}
	actualHash, err := hashWriterTo(compiled)
	if err != nil {
		return fmt.Errorf("serialize compiled circuit for %s: %w", meta.CircuitName, err)
	}
	if actualHash != meta.CircuitHash {
		return fmt.Errorf("compiled circuit hash mismatch for %s: expected %s got %s", meta.CircuitName, meta.CircuitHash, actualHash)
	}
	return nil
}

func hashWriterTo(writer interface {
	WriteTo(io.Writer) (int64, error)
}) (string, error) {
	var buf bytes.Buffer
	if _, err := writer.WriteTo(&buf); err != nil {
		return "", err
	}
	return hashBytes(buf.Bytes()), nil
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
