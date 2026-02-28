//go:build e2e.integration

package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/zk/params"
)

func TestKeeperStartupRequiresVerifiedZKParams(t *testing.T) {
	t.Run("accepts verified staged params", func(t *testing.T) {
		dir := stageRuntimeParams(t)
		t.Setenv(params.ParamsDirEnv, dir)

		require.NotPanics(t, func() {
			_ = newKeeperForTest()
		})
	})

	t.Run("rejects placeholder metadata", func(t *testing.T) {
		dir := stageRuntimeParams(t)
		metaPath := filepath.Join(dir, "params_metadata.json")
		metaBytes, err := os.ReadFile(metaPath)
		require.NoError(t, err)

		updated := strings.Replace(string(metaBytes), "\"parameter_set_id\": \"veid-zkparams-mpc-20260410\"", "\"parameter_set_id\": \"dev\"", 1)
		require.NotEqual(t, string(metaBytes), updated)
		writeRuntimeFile(t, metaPath, []byte(updated))
		writeRuntimeDigest(t, filepath.Join(dir, "params_metadata.json.sha256"), "params_metadata.json", []byte(updated))
		t.Setenv(params.ParamsDirEnv, dir)

		require.Panics(t, func() {
			_ = newKeeperForTest()
		})
	})

	t.Run("rejects corrupted staged key", func(t *testing.T) {
		dir := stageRuntimeParams(t)
		scorePath := filepath.Join(dir, "score_vk.bin")
		scoreBytes, err := os.ReadFile(scorePath)
		require.NoError(t, err)
		scoreBytes[0] ^= 0xff
		writeRuntimeFile(t, scorePath, scoreBytes)
		writeRuntimeDigest(t, filepath.Join(dir, "score_vk.bin.sha256"), "score_vk.bin", scoreBytes)
		t.Setenv(params.ParamsDirEnv, dir)

		require.Panics(t, func() {
			_ = newKeeperForTest()
		})
	})
}

func newKeeperForTest() Keeper {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	return NewKeeper(cdc, storetypes.NewKVStoreKey("veid"), "authority")
}

func stageRuntimeParams(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	srcDir := filepath.Join(filepath.Dir(file), "..", "zk", "params")
	dir := t.TempDir()
	for _, name := range []string{
		"age_vk.bin",
		"age_vk.bin.sha256",
		"residency_vk.bin",
		"residency_vk.bin.sha256",
		"score_vk.bin",
		"score_vk.bin.sha256",
		"params_metadata.json",
		"params_metadata.json.sha256",
	} {
		data, err := os.ReadFile(filepath.Join(srcDir, name))
		require.NoError(t, err)
		writeRuntimeFile(t, filepath.Join(dir, name), data)
	}
	return dir
}

func writeRuntimeDigest(t *testing.T, path, label string, data []byte) {
	t.Helper()
	content := []byte(runtimeHashBytes(data) + "  " + label + "\n")
	writeRuntimeFile(t, path, content)
}

func runtimeHashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func writeRuntimeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, data, 0o600))
}
