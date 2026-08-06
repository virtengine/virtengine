package params

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/zk/circuits"
)

func TestLoadArtifactSetAcceptsEmbeddedParams(t *testing.T) {
	artifactSet, err := loadArtifactSet(artifactSource{})
	require.NoError(t, err)
	require.Equal(t, "veid-zkparams-mpc-20260410", artifactSet.metadata.ParameterSetID)
	require.Len(t, artifactSet.keys, 3)
	require.Equal(t, 2, artifactSet.metadata.Circuits["age"].ContributorCount)
}

func TestLoadArtifactSetRejectsPlaceholderMetadata(t *testing.T) {
	dir := stageCurrentBundle(t)
	metaPath := filepath.Join(dir, metadataFileName)
	data, err := os.ReadFile(metaPath)
	require.NoError(t, err)

	updated := strings.Replace(string(data), "\"parameter_set_id\": \"veid-zkparams-mpc-20260410\"", "\"parameter_set_id\": \"dev\"", 1)
	require.NotEqual(t, string(data), updated)
	writeFile(t, metaPath, []byte(updated))
	writeDigest(t, filepath.Join(dir, metadataDigestFileName), metadataFileName, []byte(updated))

	_, err = loadArtifactSet(artifactSource{dir: dir})
	require.ErrorContains(t, err, "parameter_set_id cannot use placeholder value")
}

func TestLoadArtifactSetRejectsChecksumMismatch(t *testing.T) {
	dir := stageCurrentBundle(t)
	agePath := filepath.Join(dir, "age_vk.bin")
	data, err := os.ReadFile(agePath)
	require.NoError(t, err)

	data[0] ^= 0xff
	writeFile(t, agePath, data)

	_, err = loadArtifactSet(artifactSource{dir: dir})
	require.ErrorContains(t, err, "checksum mismatch for age_vk.bin")
}

func TestGetVerifiedVerifyingKeyRejectsCircuitDrift(t *testing.T) {
	residencyCS, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuits.ResidencyCircuit{})
	require.NoError(t, err)

	_, err = GetVerifiedVerifyingKey("age", residencyCS)
	require.ErrorContains(t, err, "compiled circuit hash mismatch")
}

func stageCurrentBundle(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	files := map[string][]byte{
		"age_vk.bin":              ageVKBytes,
		"age_vk.bin.sha256":       ageVKChecksum,
		"residency_vk.bin":        residencyVKBytes,
		"residency_vk.bin.sha256": residencyVKChecksum,
		"score_vk.bin":            scoreVKBytes,
		"score_vk.bin.sha256":     scoreVKChecksum,
		metadataFileName:          metadataBytes,
		metadataDigestFileName:    metadataChecksum,
	}
	for name, data := range files {
		writeFile(t, filepath.Join(dir, name), data)
	}
	return dir
}

func writeDigest(t *testing.T, path, label string, data []byte) {
	t.Helper()
	content := []byte(hashBytes(data) + "  " + label + "\n")
	writeFile(t, path, content)
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, data, 0o600))
}
