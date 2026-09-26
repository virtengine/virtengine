package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmoscrypto "github.com/cosmos/cosmos-sdk/crypto"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// The SDK's key armor used to be implemented on top of
// golang.org/x/crypto/openpgp/armor, which is unmaintained, unsafe by design and
// has no fixed version (GO-2026-5932). The cosmos-sdk fork now uses
// github.com/ProtonMail/go-crypto/openpgp/armor instead. This file pins the
// behaviour virtengine depends on, so the dependency swap cannot silently break
// key export/import for operators.
//
// The only difference between the two encoders is header line ordering: the
// retired one emitted headers in Go map-iteration order (unspecified, and
// therefore not stable between runs) while the replacement sorts them. Decoding
// is order-insensitive, so artifacts written by the old encoder still load.

// legacyArmoredPubKey is a fixture in the shape the retired encoder produced,
// with header lines in a non-sorted order. Decoding must not depend on order.
const legacyArmoredPubKey = `-----BEGIN TENDERMINT PUBLIC KEY-----
version: 0.0.1
type: secp256k1

Z29sZGVuLWJvZHk=
=Innj
-----END TENDERMINT PUBLIC KEY-----
`

// armoredPubKeyGolden is the exact current output for the same input, with the
// header lines sorted. If this changes, the on-disk key format changed and the
// change is a migration, not a refactor.
const armoredPubKeyGolden = `-----BEGIN TENDERMINT PUBLIC KEY-----
type: secp256k1
version: 0.0.1

Z29sZGVuLWJvZHk=
=Innj
-----END TENDERMINT PUBLIC KEY-----`

// TestLegacyKeyArmorStillDecodes proves artifacts written before the migration
// remain readable through the public SDK helpers.
func TestLegacyKeyArmorStillDecodes(t *testing.T) {
	blockType, header, data, err := cosmoscrypto.DecodeArmor(legacyArmoredPubKey)
	require.NoError(t, err)
	require.Equal(t, "TENDERMINT PUBLIC KEY", blockType)
	require.Equal(t, map[string]string{"version": "0.0.1", "type": "secp256k1"}, header)
	require.Equal(t, []byte("golden-body"), data)

	bz, algo, err := cosmoscrypto.UnarmorPubKeyBytes(legacyArmoredPubKey)
	require.NoError(t, err)
	require.Equal(t, []byte("golden-body"), bz)
	require.Equal(t, "secp256k1", algo)
}

// TestKeyArmorWireFormatIsStable pins the exact encoded bytes so a future
// dependency change cannot alter the format operators have on disk.
func TestKeyArmorWireFormatIsStable(t *testing.T) {
	require.Equal(t, armoredPubKeyGolden,
		cosmoscrypto.EncodeArmor("TENDERMINT PUBLIC KEY",
			map[string]string{"type": "secp256k1", "version": "0.0.1"},
			[]byte("golden-body")))

	// Encoding must also be deterministic: the retired encoder ranged over the
	// header map without sorting, so repeated calls could differ.
	headers := map[string]string{"type": "secp256k1", "version": "0.0.1", "kdf": "argon2"}
	first := cosmoscrypto.EncodeArmor("TENDERMINT PRIVATE KEY", headers, []byte("body"))
	for i := 0; i < 100; i++ {
		require.Equal(t, first,
			cosmoscrypto.EncodeArmor("TENDERMINT PRIVATE KEY", headers, []byte("body")),
			"armor encoding is not deterministic on iteration %d", i)
	}
}

// TestKeyringArmorExportImportRoundTrip exercises the exact path operators and
// integration tests rely on (keyring.ExportPrivKeyArmor ->
// UnarmorDecryptPrivKey), so the migration is verified end to end rather than
// only at the armor encoder.
func TestKeyringArmorExportImportRoundTrip(t *testing.T) {
	ir := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(ir)
	cdc := codec.NewProtoCodec(ir)

	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, t.TempDir(), nil, cdc)
	require.NoError(t, err)

	const (
		name       = "armor-roundtrip"
		passphrase = "passphrase"
	)

	_, mnemonic, err := kb.NewMnemonic(name, keyring.English, sdk.FullFundraiserPath, passphrase, hd.Secp256k1)
	require.NoError(t, err)
	require.NotEmpty(t, mnemonic)

	rec, err := kb.Key(name)
	require.NoError(t, err)
	want, err := rec.GetPubKey()
	require.NoError(t, err)

	armored, err := kb.ExportPrivKeyArmor(name, passphrase)
	require.NoError(t, err)
	require.Contains(t, armored, "-----BEGIN TENDERMINT PRIVATE KEY-----")

	privKey, algo, err := cosmoscrypto.UnarmorDecryptPrivKey(armored, passphrase)
	require.NoError(t, err)
	require.Equal(t, "secp256k1", algo)
	require.True(t, want.Equals(privKey.PubKey()), "exported armored key did not round-trip to the same public key")

	// A wrong passphrase must still be rejected.
	_, _, err = cosmoscrypto.UnarmorDecryptPrivKey(armored, "wrong-passphrase")
	require.Error(t, err)

	// Re-import the armored key under a new name and confirm it is the same key.
	require.NoError(t, kb.ImportPrivKey("armor-reimported", armored, passphrase))
	rec2, err := kb.Key("armor-reimported")
	require.NoError(t, err)
	got2, err := rec2.GetPubKey()
	require.NoError(t, err)
	require.True(t, want.Equals(got2), "re-imported key differs from the original")
}

// TestKeyArmorRoundTripAcrossLineWrapBoundaries covers the base64 line wrapper,
// which emits 48 raw bytes per line and therefore has off-by-one hazards.
func TestKeyArmorRoundTripAcrossLineWrapBoundaries(t *testing.T) {
	for _, n := range []int{0, 1, 2, 47, 48, 49, 63, 64, 95, 96, 97, 100, 1000, 4096} {
		body := make([]byte, n)
		for i := range body {
			body[i] = byte(i % 251)
		}

		armored := cosmoscrypto.EncodeArmor("MINT TEST", map[string]string{"type": "Info"}, body)

		blockType, header, got, err := cosmoscrypto.DecodeArmor(armored)
		require.NoError(t, err, "n=%d", n)
		require.Equal(t, "MINT TEST", blockType, "n=%d", n)
		require.Equal(t, map[string]string{"type": "Info"}, header, "n=%d", n)
		require.Equal(t, body, got, "n=%d", n)
	}
}

// TestKeyArmorPrivateKeySurvivesHeaderReordering mirrors how legacy keyring
// files were written: the retired encoder chose header order at random. Every
// ordering must still decrypt.
func TestKeyArmorPrivateKeySurvivesHeaderReordering(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	const passphrase = "passphrase"

	armored := cosmoscrypto.EncryptArmorPrivKey(privKey, passphrase, "secp256k1")

	lines := splitArmorLines(armored)
	require.Len(t, lines.headerIdx, 3, "expected kdf, salt and type headers")

	for _, perm := range permutations(3) {
		variant := reorderArmorHeaders(lines, perm)

		got, algo, err := cosmoscrypto.UnarmorDecryptPrivKey(variant, passphrase)
		require.NoError(t, err, "header order %v must remain decryptable", perm)
		require.True(t, privKey.Equals(got), "header order %v changed the decrypted key", perm)
		require.Equal(t, "secp256k1", algo)
	}
}

type armorLayout struct {
	lines     []string
	headerIdx []int
}

func splitArmorLines(armored string) armorLayout {
	lines := splitLines(armored)

	var idx []int
	for i, l := range lines {
		if len(l) > 2 && l[0] != '-' && containsHeaderSep(l) {
			idx = append(idx, i)
		}
	}
	return armorLayout{lines: lines, headerIdx: idx}
}

func containsHeaderSep(s string) bool {
	for i := 0; i+1 < len(s); i++ {
		if s[i] == ':' && s[i+1] == ' ' {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func reorderArmorHeaders(l armorLayout, order []int) string {
	lines := make([]string, len(l.lines))
	copy(lines, l.lines)

	orig := make([]string, len(l.headerIdx))
	for i, lineIdx := range l.headerIdx {
		orig[i] = lines[lineIdx]
	}
	for i, lineIdx := range l.headerIdx {
		lines[lineIdx] = orig[order[i]]
	}

	out := ""
	for i, line := range lines {
		if i > 0 {
			out += "\n"
		}
		out += line
	}
	return out
}

func permutations(n int) [][]int {
	var out [][]int
	var rec func(cur []int, used []bool)
	rec = func(cur []int, used []bool) {
		if len(cur) == n {
			cp := make([]int, n)
			copy(cp, cur)
			out = append(out, cp)
			return
		}
		for i := 0; i < n; i++ {
			if used[i] {
				continue
			}
			used[i] = true
			rec(append(cur, i), used)
			used[i] = false
		}
	}
	rec(nil, make([]bool, n))
	return out
}
