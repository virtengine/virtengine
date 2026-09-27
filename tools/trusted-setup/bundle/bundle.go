package bundle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	PhaseRequestSchema     = "virtengine.trusted_setup.phase_request/v1"
	PhaseResponseSchema    = "virtengine.trusted_setup.phase_response/v1"
	ArtifactManifestSchema = "virtengine.trusted_setup.artifact_manifest/v1"
)

// ErrUnsafeBundlePath is returned when a file name carried inside a bundle would
// escape the directory the bundle is read from or written to.
var ErrUnsafeBundlePath = errors.New("unsafe bundle file name")

// SafeJoin joins a bundle-relative file name onto baseDir, rejecting any name
// that could escape it.
//
// File names recorded in phase request/response bundles are attacker-controlled:
// bundles are exchanged between ceremony participants, so a hostile participant
// can set `"input_file": "../../../../etc/passwd"` to make the tool read or
// overwrite files outside the bundle directory (CWE-22 path traversal, reported
// by gosec as G304/G703). Only bare file names are accepted.
func SafeJoin(baseDir, name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("%w: empty file name", ErrUnsafeBundlePath)
	}
	if strings.ContainsAny(trimmed, `/\`) {
		return "", fmt.Errorf("%w: %q must not contain a path separator", ErrUnsafeBundlePath, name)
	}
	if trimmed == "." || trimmed == ".." {
		return "", fmt.Errorf("%w: %q is not a file name", ErrUnsafeBundlePath, name)
	}

	joined := filepath.Join(baseDir, trimmed)

	// Defence in depth: the joined path must still resolve inside baseDir.
	rel, err := filepath.Rel(baseDir, joined)
	if err != nil {
		return "", fmt.Errorf("%w: %q: %v", ErrUnsafeBundlePath, name, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %q escapes %s", ErrUnsafeBundlePath, name, baseDir)
	}

	return joined, nil
}

type FileRecord struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
}

type PhaseRequest struct {
	SchemaVersion  string   `json:"schema_version"`
	CeremonyID     string   `json:"ceremony_id"`
	CircuitName    string   `json:"circuit_name"`
	Phase          string   `json:"phase"`
	TranscriptHash string   `json:"transcript_hash"`
	InputFile      string   `json:"input_file"`
	InputHash      string   `json:"input_hash"`
	Notes          []string `json:"notes,omitempty"`
}

type PhaseResponse struct {
	SchemaVersion string `json:"schema_version"`
	CeremonyID    string `json:"ceremony_id"`
	Phase         string `json:"phase"`
	RequestHash   string `json:"request_hash"`
	PayloadFile   string `json:"payload_file"`
	ParticipantID string `json:"participant_id"`
	PublicKey     string `json:"public_key"`
	Attestation   string `json:"attestation,omitempty"`
	InputHash     string `json:"input_hash"`
	OutputHash    string `json:"output_hash"`
	Signature     string `json:"signature"`
}

type ArtifactManifest struct {
	SchemaVersion        string       `json:"schema_version"`
	CeremonyID           string       `json:"ceremony_id"`
	CircuitName          string       `json:"circuit_name"`
	TranscriptHash       string       `json:"transcript_hash"`
	ParametersVersion    string       `json:"parameters_version"`
	ContributorCount     int          `json:"contributor_count"`
	Contributors         []string     `json:"contributors"`
	CoordinatorID        string       `json:"coordinator_id"`
	CoordinatorPublicKey string       `json:"coordinator_public_key"`
	SignatureAlgorithm   string       `json:"signature_algorithm"`
	VerificationReport   string       `json:"verification_report"`
	Files                []FileRecord `json:"files"`
}

func Marshal(v interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func WriteJSON(path string, v interface{}) error {
	data, err := Marshal(v)
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data, 0o600)
}

func ReadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path) // #nosec G304 -- the path is composed from the ceremony state directory (given once by the operator on the command line) plus fixed file names, so remote input cannot influence it
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func HashFile(path string) (string, int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}
	if info.IsDir() {
		return "", 0, fmt.Errorf("%s is a directory", path)
	}
	data, err := os.ReadFile(path) // #nosec G304 -- the path is composed from the ceremony state directory (given once by the operator on the command line) plus fixed file names, so remote input cannot influence it
	if err != nil {
		return "", 0, err
	}
	return HashBytes(data), info.Size(), nil
}

func BuildFileRecord(baseDir, path string) (FileRecord, error) {
	hash, size, err := HashFile(path)
	if err != nil {
		return FileRecord{}, err
	}
	rel, err := filepath.Rel(baseDir, path)
	if err != nil {
		return FileRecord{}, err
	}
	return FileRecord{
		Path:      filepath.ToSlash(rel),
		SHA256:    hash,
		SizeBytes: size,
	}, nil
}

func WriteDigestFile(path, target string) error {
	hash, _, err := HashFile(target)
	if err != nil {
		return err
	}
	label := filepath.Base(target)
	if rel, relErr := filepath.Rel(filepath.Dir(path), target); relErr == nil && !strings.HasPrefix(rel, "..") {
		label = filepath.ToSlash(rel)
	}
	return writeFileAtomic(path, []byte(fmt.Sprintf("%s  %s\n", hash, label)), 0o600)
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
