package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"

	"github.com/virtengine/virtengine/tools/trusted-setup/transcript"
	"github.com/virtengine/virtengine/x/veid/zk/circuits"
)

type metadata struct {
	Version          string            `json:"version"`
	CeremonyHash     string            `json:"ceremony_hash"`
	ContributorCount int               `json:"contributor_count"`
	Contributors     []string          `json:"contributors"`
	CircuitHash      string            `json:"circuit_hash,omitempty"`
	CircuitHashes    map[string]string `json:"circuit_hashes,omitempty"`
	GeneratedAt      string            `json:"generated_at"`
	BeaconHash       string            `json:"beacon_hash,omitempty"`
	Notes            string            `json:"notes,omitempty"`
}

func main() {
	baseDir := filepath.Join("x", "veid", "zk", "params")
	if err := os.MkdirAll(baseDir, 0o750); err != nil {
		panic(err)
	}

	circuitHashes := map[string]string{
		"age":       hashCircuit(&circuits.AgeRangeCircuit{}),
		"residency": hashCircuit(&circuits.ResidencyCircuit{}),
		"score":     hashCircuit(&circuits.ScoreRangeCircuit{}),
	}
	if err := writeVK(filepath.Join(baseDir, "age_vk.bin"), &circuits.AgeRangeCircuit{}); err != nil {
		panic(err)
	}
	if err := writeVK(filepath.Join(baseDir, "residency_vk.bin"), &circuits.ResidencyCircuit{}); err != nil {
		panic(err)
	}
	if err := writeVK(filepath.Join(baseDir, "score_vk.bin"), &circuits.ScoreRangeCircuit{}); err != nil {
		panic(err)
	}

	meta := metadata{
		Version:          "dev",
		CeremonyHash:     "",
		ContributorCount: 1,
		Contributors:     []string{"dev-setup"},
		CircuitHashes:    circuitHashes,
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		Notes:            "dev setup (single-party)",
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "params_metadata.json"), data, 0o600); err != nil {
		panic(err)
	}
}

func writeVK(path string, circuit frontend.Circuit) error {
	r1csCompiled, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		return err
	}
	_, vk, err := groth16.Setup(r1csCompiled)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if _, err := vk.WriteTo(&buf); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o600)
}

func hashCircuit(circuit frontend.Circuit) string {
	r1csCompiled, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	if _, err := r1csCompiled.WriteTo(&buf); err != nil {
		panic(err)
	}
	return transcript.HashBytes(buf.Bytes())
}
