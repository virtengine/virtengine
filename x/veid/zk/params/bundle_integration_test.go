//go:build integration

package params

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/zk/circuits"
)

func TestGetVerifiedVerifyingKeySupportsStagedParamDirectory(t *testing.T) {
	dir := stageCurrentBundle(t)
	t.Setenv(ParamsDirEnv, dir)

	ageCS, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuits.AgeRangeCircuit{})
	require.NoError(t, err)

	vk, err := GetVerifiedVerifyingKey("age", ageCS)
	require.NoError(t, err)
	require.NotNil(t, vk)
}

func TestGetVerifiedVerifyingKeyRejectsStagedBundleWithHashDrift(t *testing.T) {
	dir := stageCurrentBundle(t)
	t.Setenv(ParamsDirEnv, dir)

	scorePath := filepath.Join(dir, "score_vk.bin")
	scoreBytes, err := os.ReadFile(scorePath)
	require.NoError(t, err)
	scoreBytes[len(scoreBytes)-1] ^= 0xff
	writeFile(t, scorePath, scoreBytes)
	writeDigest(t, filepath.Join(dir, "score_vk.bin.sha256"), "score_vk.bin", scoreBytes)

	scoreCS, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuits.ScoreRangeCircuit{})
	require.NoError(t, err)

	_, err = GetVerifiedVerifyingKey("score", scoreCS)
	require.ErrorContains(t, err, "verification key hash mismatch")
}
