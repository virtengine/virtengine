package params

import (
	"bytes"
	"encoding/json"
	"fmt"

	_ "embed"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
)

//go:embed age_vk.bin
var ageVKBytes []byte

//go:embed residency_vk.bin
var residencyVKBytes []byte

//go:embed score_vk.bin
var scoreVKBytes []byte

//go:embed params_metadata.json
var metadataBytes []byte

// Metadata describes the ceremony output embedded with the binary.
type Metadata struct {
	Version          string            `json:"version"`
	CeremonyHash     string            `json:"ceremony_hash"`
	ContributorCount int               `json:"contributor_count"`
	Contributors     []string          `json:"contributors"`
	CircuitHashes    map[string]string `json:"circuit_hashes"`
	GeneratedAt      string            `json:"generated_at"`
	BeaconHash       string            `json:"beacon_hash"`
	Notes            string            `json:"notes"`
}

var (
	cachedAgeVK       groth16.VerifyingKey
	cachedResidencyVK groth16.VerifyingKey
	cachedScoreVK     groth16.VerifyingKey
	cachedMetadata    *Metadata
)

// GetVerifyingKey returns the embedded verifying key for the named circuit.
// Valid circuit names: age, residency, score.
func GetVerifyingKey(name string) (groth16.VerifyingKey, error) {
	switch name {
	case "age":
		if cachedAgeVK != nil {
			return cachedAgeVK, nil
		}
		vk, err := decodeVerifyingKey(ageVKBytes)
		if err != nil {
			return nil, err
		}
		cachedAgeVK = vk
		return cachedAgeVK, nil
	case "residency":
		if cachedResidencyVK != nil {
			return cachedResidencyVK, nil
		}
		vk, err := decodeVerifyingKey(residencyVKBytes)
		if err != nil {
			return nil, err
		}
		cachedResidencyVK = vk
		return cachedResidencyVK, nil
	case "score":
		if cachedScoreVK != nil {
			return cachedScoreVK, nil
		}
		vk, err := decodeVerifyingKey(scoreVKBytes)
		if err != nil {
			return nil, err
		}
		cachedScoreVK = vk
		return cachedScoreVK, nil
	default:
		return nil, fmt.Errorf("unknown circuit name: %s", name)
	}
}

// GetMetadata returns metadata describing the embedded ceremony output.
func GetMetadata() (*Metadata, error) {
	if cachedMetadata != nil {
		return cachedMetadata, nil
	}

	var meta Metadata
	if err := json.Unmarshal(metadataBytes, &meta); err != nil {
		return nil, fmt.Errorf("parse ceremony metadata: %w", err)
	}

	cachedMetadata = &meta
	return cachedMetadata, nil
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
