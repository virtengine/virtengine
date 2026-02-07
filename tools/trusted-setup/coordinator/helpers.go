package coordinator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/consensys/gnark/backend/groth16/bn254/mpcsetup"
	cs "github.com/consensys/gnark/constraint/bn254"

	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
)

func loadTranscript(state State) (*transcript.Transcript, error) {
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

func saveTranscript(state State, tr *transcript.Transcript) error {
	data, err := tr.Marshal()
	if err != nil {
		return err
	}
	return state.SaveTranscript(data)
}

func hashTranscript(tr *transcript.Transcript) string {
	hash, err := transcript.HashJSON(tr)
	if err != nil {
		return ""
	}
	return hash
}

func loadLatestPhase1(state State) (*mpcsetup.Phase1, []byte, error) {
	paths, err := state.Phase1ContributionPaths()
	if err != nil {
		return nil, nil, err
	}
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("no phase1 contributions found")
	}
	latest := paths[len(paths)-1]
	data, err := os.ReadFile(latest)
	if err != nil {
		return nil, nil, err
	}
	phase := new(mpcsetup.Phase1)
	if _, err := phase.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, nil, err
	}
	return phase, data, nil
}

func loadLatestPhase2(state State) (*mpcsetup.Phase2, []byte, error) {
	paths, err := state.Phase2ContributionPaths()
	if err != nil {
		return nil, nil, err
	}
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("no phase2 contributions found")
	}
	latest := paths[len(paths)-1]
	data, err := os.ReadFile(latest)
	if err != nil {
		return nil, nil, err
	}
	phase := new(mpcsetup.Phase2)
	if _, err := phase.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, nil, err
	}
	return phase, data, nil
}

func loadPhase1Contributions(state State) ([]*mpcsetup.Phase1, []string, error) {
	paths, err := state.Phase1ContributionPaths()
	if err != nil {
		return nil, nil, err
	}
	if len(paths) < 2 {
		return nil, paths, fmt.Errorf("phase1 requires at least one participant contribution")
	}
	var contribs []*mpcsetup.Phase1
	for _, path := range paths[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		phase := new(mpcsetup.Phase1)
		if _, err := phase.ReadFrom(bytes.NewReader(data)); err != nil {
			return nil, nil, err
		}
		contribs = append(contribs, phase)
	}
	return contribs, paths[1:], nil
}

func loadPhase2Contributions(state State) ([]*mpcsetup.Phase2, []string, error) {
	paths, err := state.Phase2ContributionPaths()
	if err != nil {
		return nil, nil, err
	}
	if len(paths) < 2 {
		return nil, paths, fmt.Errorf("phase2 requires at least one participant contribution")
	}
	var contribs []*mpcsetup.Phase2
	for _, path := range paths[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		phase := new(mpcsetup.Phase2)
		if _, err := phase.ReadFrom(bytes.NewReader(data)); err != nil {
			return nil, nil, err
		}
		contribs = append(contribs, phase)
	}
	return contribs, paths[1:], nil
}

func loadR1CS(state State) (*cs.R1CS, error) {
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

func loadCommons(state State) (*mpcsetup.SrsCommons, error) {
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

func writeToBytes(writer interface {
	WriteTo(io.Writer) (int64, error)
}) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := writer.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
