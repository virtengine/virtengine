package verify

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
	cs "github.com/consensys/gnark/constraint/bn254"

	"github.com/virtengine/virtengine/tools/trusted-setup/coordinator"
	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
)

// Result captures the verification output for a ceremony workspace.
type Result struct {
	Config           *coordinator.Config    `json:"config"`
	Transcript       *transcript.Transcript `json:"transcript"`
	TranscriptHash   string                 `json:"transcript_hash"`
	Phase1Valid      bool                   `json:"phase1_valid"`
	Phase2Valid      bool                   `json:"phase2_valid"`
	ProvingKeyHash   string                 `json:"proving_key_hash,omitempty"`
	VerifyingKeyHash string                 `json:"verifying_key_hash,omitempty"`
	Phase1Files      []string               `json:"phase1_files"`
	Phase2Files      []string               `json:"phase2_files"`
}

// Verify validates the ceremony transcript and contributions within a workspace.
func Verify(dir string) (*Result, error) {
	state := coordinator.State{BaseDir: dir}
	cfg, err := state.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	tr, err := loadTranscript(state)
	if err != nil {
		return nil, fmt.Errorf("load transcript: %w", err)
	}

	result := &Result{
		Config:     cfg,
		Transcript: tr,
	}
	result.TranscriptHash = hashTranscript(tr)

	phase1Files, err := state.Phase1ContributionPaths()
	if err != nil {
		return nil, fmt.Errorf("list phase1 contributions: %w", err)
	}
	result.Phase1Files = phase1Files

	phase2Files, err := state.Phase2ContributionPaths()
	if err != nil {
		return nil, fmt.Errorf("list phase2 contributions: %w", err)
	}
	result.Phase2Files = phase2Files

	phase1Valid, err := verifyPhase1(tr, phase1Files)
	if err != nil {
		return nil, err
	}
	result.Phase1Valid = phase1Valid

	phase2Valid, pkHash, vkHash, err := verifyPhase2(state, cfg, tr, phase2Files)
	if err != nil {
		return nil, err
	}
	result.Phase2Valid = phase2Valid
	result.ProvingKeyHash = pkHash
	result.VerifyingKeyHash = vkHash

	return result, nil
}

func loadTranscript(state coordinator.State) (*transcript.Transcript, error) {
	data, err := state.LoadTranscript()
	if err != nil {
		return nil, err
	}
	var tr transcript.Transcript
	if err := json.Unmarshal(data, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func hashTranscript(tr *transcript.Transcript) string {
	hash, err := transcript.HashJSON(tr)
	if err != nil {
		return ""
	}
	return hash
}

func verifyPhase1(tr *transcript.Transcript, paths []string) (bool, error) {
	if len(paths) == 0 && len(tr.Phase1.Contributions) == 0 && tr.Phase1.InitialHash == "" {
		return false, nil
	}
	if len(paths) == 0 {
		return false, fmt.Errorf("phase1 contributions missing")
	}

	initialBytes, err := os.ReadFile(paths[0])
	if err != nil {
		return false, fmt.Errorf("read phase1 initial: %w", err)
	}
	initialHash := transcript.HashBytes(initialBytes)
	if tr.Phase1.InitialHash != "" && tr.Phase1.InitialHash != initialHash {
		return false, fmt.Errorf("phase1 initial hash mismatch: %s != %s", tr.Phase1.InitialHash, initialHash)
	}

	if len(tr.Phase1.Contributions) > 0 && len(tr.Phase1.Contributions) != len(paths)-1 {
		return false, fmt.Errorf("phase1 transcript count mismatch: transcript=%d files=%d", len(tr.Phase1.Contributions), len(paths)-1)
	}

	prev := new(mpcsetup.Phase1)
	if _, err := prev.ReadFrom(bytes.NewReader(initialBytes)); err != nil {
		return false, fmt.Errorf("parse phase1 initial: %w", err)
	}
	prevBytes := initialBytes

	for idx, path := range paths[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			return false, fmt.Errorf("read phase1 contribution: %w", err)
		}
		next := new(mpcsetup.Phase1)
		if _, err := next.ReadFrom(bytes.NewReader(data)); err != nil {
			return false, fmt.Errorf("parse phase1 contribution: %w", err)
		}
		if err := prev.Verify(next); err != nil {
			return false, fmt.Errorf("phase1 verify failed: %w", err)
		}

		if idx < len(tr.Phase1.Contributions) {
			record := tr.Phase1.Contributions[idx]
			inputHash := transcript.HashBytes(prevBytes)
			outputHash := transcript.HashBytes(data)
			if record.InputHash != "" && record.InputHash != inputHash {
				return false, fmt.Errorf("phase1 input hash mismatch for %s", record.File)
			}
			if record.OutputHash != "" && record.OutputHash != outputHash {
				return false, fmt.Errorf("phase1 output hash mismatch for %s", record.File)
			}
			if record.File != "" && filepath.Base(path) != record.File {
				return false, fmt.Errorf("phase1 file mismatch: %s != %s", record.File, filepath.Base(path))
			}
		}

		prev = next
		prevBytes = data
	}

	finalHash := transcript.HashBytes(prevBytes)
	if tr.Phase1.FinalHash != "" && tr.Phase1.FinalHash != finalHash {
		return false, fmt.Errorf("phase1 final hash mismatch: %s != %s", tr.Phase1.FinalHash, finalHash)
	}

	return true, nil
}

func verifyPhase2(state coordinator.State, cfg *coordinator.Config, tr *transcript.Transcript, paths []string) (bool, string, string, error) {
	if len(paths) == 0 && len(tr.Phase2.Contributions) == 0 && tr.Phase2.InitialHash == "" {
		return false, "", "", nil
	}
	if len(paths) == 0 {
		return false, "", "", fmt.Errorf("phase2 contributions missing")
	}

	initialBytes, err := os.ReadFile(paths[0])
	if err != nil {
		return false, "", "", fmt.Errorf("read phase2 initial: %w", err)
	}
	initialHash := transcript.HashBytes(initialBytes)
	if tr.Phase2.InitialHash != "" && tr.Phase2.InitialHash != initialHash {
		return false, "", "", fmt.Errorf("phase2 initial hash mismatch: %s != %s", tr.Phase2.InitialHash, initialHash)
	}

	if len(tr.Phase2.Contributions) > 0 && len(tr.Phase2.Contributions) != len(paths)-1 {
		return false, "", "", fmt.Errorf("phase2 transcript count mismatch: transcript=%d files=%d", len(tr.Phase2.Contributions), len(paths)-1)
	}

	prev := new(mpcsetup.Phase2)
	if _, err := prev.ReadFrom(bytes.NewReader(initialBytes)); err != nil {
		return false, "", "", fmt.Errorf("parse phase2 initial: %w", err)
	}
	prevBytes := initialBytes

	contribs := make([]*mpcsetup.Phase2, 0, len(paths)-1)
	for idx, path := range paths[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			return false, "", "", fmt.Errorf("read phase2 contribution: %w", err)
		}
		next := new(mpcsetup.Phase2)
		if _, err := next.ReadFrom(bytes.NewReader(data)); err != nil {
			return false, "", "", fmt.Errorf("parse phase2 contribution: %w", err)
		}
		if err := prev.Verify(next); err != nil {
			return false, "", "", fmt.Errorf("phase2 verify failed: %w", err)
		}

		if idx < len(tr.Phase2.Contributions) {
			record := tr.Phase2.Contributions[idx]
			inputHash := transcript.HashBytes(prevBytes)
			outputHash := transcript.HashBytes(data)
			if record.InputHash != "" && record.InputHash != inputHash {
				return false, "", "", fmt.Errorf("phase2 input hash mismatch for %s", record.File)
			}
			if record.OutputHash != "" && record.OutputHash != outputHash {
				return false, "", "", fmt.Errorf("phase2 output hash mismatch for %s", record.File)
			}
			if record.File != "" && filepath.Base(path) != record.File {
				return false, "", "", fmt.Errorf("phase2 file mismatch: %s != %s", record.File, filepath.Base(path))
			}
		}

		contribs = append(contribs, next)
		prev = next
		prevBytes = data
	}

	finalHash := transcript.HashBytes(prevBytes)
	if tr.Phase2.FinalHash != "" && tr.Phase2.FinalHash != finalHash {
		return false, "", "", fmt.Errorf("phase2 final hash mismatch: %s != %s", tr.Phase2.FinalHash, finalHash)
	}

	if len(contribs) == 0 {
		return true, "", "", nil
	}

	r1csBN, err := loadR1CS(state)
	if err != nil {
		return false, "", "", fmt.Errorf("load r1cs: %w", err)
	}
	commons, err := loadCommons(state)
	if err != nil {
		return false, "", "", fmt.Errorf("load commons: %w", err)
	}

	pk, vk, err := mpcsetup.VerifyPhase2(r1csBN, commons, decodeBeacon(cfg.Beacon), contribs...)
	if err != nil {
		return false, "", "", fmt.Errorf("verify phase2 parameters: %w", err)
	}

	pkHash, vkHash, err := hashKeys(pk, vk)
	if err != nil {
		return false, "", "", err
	}

	if tr.Final.ProvingKeyHash != "" && tr.Final.ProvingKeyHash != pkHash {
		return false, "", "", fmt.Errorf("proving key hash mismatch: %s != %s", tr.Final.ProvingKeyHash, pkHash)
	}
	if tr.Final.VerifyingKeyHash != "" && tr.Final.VerifyingKeyHash != vkHash {
		return false, "", "", fmt.Errorf("verifying key hash mismatch: %s != %s", tr.Final.VerifyingKeyHash, vkHash)
	}

	return true, pkHash, vkHash, nil
}

func loadR1CS(state coordinator.State) (*cs.R1CS, error) {
	path := filepath.Join(state.Phase2Dir(), "r1cs.bin")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r1cs := new(cs.R1CS)
	if _, err := r1cs.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return r1cs, nil
}

func loadCommons(state coordinator.State) (*mpcsetup.SrsCommons, error) {
	path := filepath.Join(state.Phase1Dir(), "commons.bin")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	commons := new(mpcsetup.SrsCommons)
	if _, err := commons.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return commons, nil
}

func decodeBeacon(beacon string) []byte {
	if beacon == "" {
		return nil
	}
	data, err := hex.DecodeString(beacon)
	if err != nil {
		return []byte(beacon)
	}
	return data
}

func hashKeys(pk groth16.ProvingKey, vk groth16.VerifyingKey) (string, string, error) {
	pkBytes, err := writeToBytes(pk)
	if err != nil {
		return "", "", fmt.Errorf("serialize proving key: %w", err)
	}
	vkBytes, err := writeToBytes(vk)
	if err != nil {
		return "", "", fmt.Errorf("serialize verifying key: %w", err)
	}
	return transcript.HashBytes(pkBytes), transcript.HashBytes(vkBytes), nil
}

func writeToBytes(writer interface {
	WriteTo(io.Writer) (int64, error)
}) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := writer.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
