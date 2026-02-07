package params

import (
	"bytes"
	"fmt"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
)

// LoadVerifyingKey loads a verifying key from disk for off-chain tooling.
func LoadVerifyingKey(path string) (groth16.VerifyingKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read verifying key: %w", err)
	}

	vk := groth16.NewVerifyingKey(ecc.BN254)
	if _, err := vk.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("parse verifying key: %w", err)
	}

	return vk, nil
}

// LoadProvingKey loads a proving key from disk for off-chain tooling.
func LoadProvingKey(path string) (groth16.ProvingKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read proving key: %w", err)
	}

	pk := groth16.NewProvingKey(ecc.BN254)
	if _, err := pk.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("parse proving key: %w", err)
	}

	return pk, nil
}
