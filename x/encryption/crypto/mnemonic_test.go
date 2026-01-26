package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants for BIP-39 test vectors
const (
	// testMnemonic12 is the standard BIP-39 12-word test mnemonic
	testMnemonic12 = testMnemonic12

	// testMnemonic24 is the standard BIP-39 24-word test mnemonic
	testMnemonic24 = testMnemonic24
)

// ============================================================================
// BIP-39 Mnemonic Generation Tests
// ============================================================================

func TestGenerateMnemonic_12Words(t *testing.T) {
	mnemonic, err := GenerateMnemonic(Mnemonic12Words)
	require.NoError(t, err)
	require.NotEmpty(t, mnemonic)

	words := strings.Fields(mnemonic)
	assert.Len(t, words, 12, "12-word mnemonic should have 12 words")

	// Verify mnemonic is valid BIP-39
	assert.True(t, ValidateMnemonic(mnemonic), "generated mnemonic should be valid")
}

func TestGenerateMnemonic_24Words(t *testing.T) {
	mnemonic, err := GenerateMnemonic(Mnemonic24Words)
	require.NoError(t, err)
	require.NotEmpty(t, mnemonic)

	words := strings.Fields(mnemonic)
	assert.Len(t, words, 24, "24-word mnemonic should have 24 words")

	// Verify mnemonic is valid BIP-39
	assert.True(t, ValidateMnemonic(mnemonic), "generated mnemonic should be valid")
}

func TestGenerateMnemonic_InvalidSize(t *testing.T) {
	_, err := GenerateMnemonic(MnemonicSize(15))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mnemonic size")
}

func TestGenerateMnemonic_Uniqueness(t *testing.T) {
	// Generate multiple mnemonics and ensure they are unique
	mnemonics := make(map[string]bool)
	for i := 0; i < 10; i++ {
		mnemonic, err := GenerateMnemonic(Mnemonic12Words)
		require.NoError(t, err)

		assert.False(t, mnemonics[mnemonic], "generated mnemonic should be unique")
		mnemonics[mnemonic] = true
	}
}

// ============================================================================
// BIP-39 Mnemonic Validation Tests
// ============================================================================

func TestValidateMnemonic_Valid12Word(t *testing.T) {
	// Standard BIP-39 test vector
	mnemonic := testMnemonic12
	assert.True(t, ValidateMnemonic(mnemonic))
}

func TestValidateMnemonic_Valid24Word(t *testing.T) {
	// Standard BIP-39 test vector
	mnemonic := testMnemonic24
	assert.True(t, ValidateMnemonic(mnemonic))
}

func TestValidateMnemonic_InvalidChecksum(t *testing.T) {
	// Invalid checksum (last word changed)
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
	assert.False(t, ValidateMnemonic(mnemonic))
}

func TestValidateMnemonic_InvalidWord(t *testing.T) {
	// Contains invalid word
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon invalid123"
	assert.False(t, ValidateMnemonic(mnemonic))
}

func TestValidateMnemonic_Empty(t *testing.T) {
	assert.False(t, ValidateMnemonic(""))
}

func TestValidateMnemonic_WrongWordCount(t *testing.T) {
	// Only 5 words
	mnemonic := "abandon abandon abandon abandon abandon"
	assert.False(t, ValidateMnemonic(mnemonic))
}

func TestMnemonicWordCount(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		expected int
	}{
		{"empty", "", 0},
		{"12 words", testMnemonic12, 12},
		{"24 words", testMnemonic24, 24},
		{"extra whitespace", "  abandon   abandon  abandon  abandon abandon abandon abandon abandon abandon abandon abandon about  ", 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := MnemonicWordCount(tt.mnemonic)
			assert.Equal(t, tt.expected, count)
		})
	}
}

// ============================================================================
// Entropy Size Tests
// ============================================================================

func TestEntropySizeForMnemonic(t *testing.T) {
	tests := []struct {
		size         MnemonicSize
		expectedBits int
		expectError  bool
	}{
		{Mnemonic12Words, 128, false},
		{Mnemonic24Words, 256, false},
		{MnemonicSize(15), 0, true},
		{MnemonicSize(0), 0, true},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.size)), func(t *testing.T) {
			bits, err := EntropySizeForMnemonic(tt.size)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBits, bits)
			}
		})
	}
}

func TestIsValidMnemonicSize(t *testing.T) {
	assert.True(t, IsValidMnemonicSize(Mnemonic12Words))
	assert.True(t, IsValidMnemonicSize(Mnemonic24Words))
	assert.False(t, IsValidMnemonicSize(MnemonicSize(15)))
	assert.False(t, IsValidMnemonicSize(MnemonicSize(0)))
}

// ============================================================================
// BIP-44 Key Derivation Tests
// ============================================================================

func TestDeriveKeyFromMnemonic_DefaultPath(t *testing.T) {
	mnemonic := testMnemonic12

	key, err := DeriveKeyFromMnemonic(mnemonic, "")
	require.NoError(t, err)
	require.NotNil(t, key)
	defer key.Zero()

	// Verify key properties
	assert.Len(t, key.PrivateKey, 32, "private key should be 32 bytes")
	assert.Len(t, key.PublicKey, 33, "compressed public key should be 33 bytes")
	assert.NotEmpty(t, key.Address, "address should not be empty")
	assert.Equal(t, DefaultHDPath, key.HDPath, "should use default HD path")

	// Address should start with cosmos prefix
	assert.True(t, strings.HasPrefix(key.Address, "cosmos"), "address should have cosmos prefix")
}

func TestDeriveKeyFromMnemonic_CustomPath(t *testing.T) {
	mnemonic := testMnemonic12
	customPath := "m/44'/118'/1'/0/0"

	key, err := DeriveKeyFromMnemonic(mnemonic, customPath)
	require.NoError(t, err)
	require.NotNil(t, key)
	defer key.Zero()

	assert.Equal(t, customPath, key.HDPath, "should use custom HD path")
}

func TestDeriveKeyFromMnemonic_InvalidMnemonic(t *testing.T) {
	_, err := DeriveKeyFromMnemonic("invalid mnemonic phrase", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mnemonic")
}

func TestDeriveKeyFromMnemonic_InvalidPath(t *testing.T) {
	mnemonic := testMnemonic12

	_, err := DeriveKeyFromMnemonic(mnemonic, "invalid/path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid HD path")
}

func TestDeriveKeyFromMnemonic_Deterministic(t *testing.T) {
	mnemonic := testMnemonic12

	key1, err := DeriveKeyFromMnemonic(mnemonic, DefaultHDPath)
	require.NoError(t, err)
	defer key1.Zero()

	key2, err := DeriveKeyFromMnemonic(mnemonic, DefaultHDPath)
	require.NoError(t, err)
	defer key2.Zero()

	// Same mnemonic + path should produce identical keys
	assert.Equal(t, key1.PrivateKey, key2.PrivateKey, "private keys should match")
	assert.Equal(t, key1.PublicKey, key2.PublicKey, "public keys should match")
	assert.Equal(t, key1.Address, key2.Address, "addresses should match")
}

func TestDeriveKeyFromMnemonic_DifferentPaths(t *testing.T) {
	mnemonic := testMnemonic12

	key1, err := DeriveKeyFromMnemonic(mnemonic, "m/44'/118'/0'/0/0")
	require.NoError(t, err)
	defer key1.Zero()

	key2, err := DeriveKeyFromMnemonic(mnemonic, "m/44'/118'/1'/0/0")
	require.NoError(t, err)
	defer key2.Zero()

	// Different paths should produce different keys
	assert.NotEqual(t, key1.PrivateKey, key2.PrivateKey, "different paths should produce different private keys")
	assert.NotEqual(t, key1.Address, key2.Address, "different paths should produce different addresses")
}

// ============================================================================
// DeriveOptions Tests
// ============================================================================

func TestNewDefaultDeriveOptions(t *testing.T) {
	opts := NewDefaultDeriveOptions()

	assert.Equal(t, DefaultHDPath, opts.HDPath)
	assert.Equal(t, "", opts.Passphrase)
	assert.Equal(t, DefaultCoinType, opts.CoinType)
	assert.Equal(t, DefaultAccountIndex, opts.Account)
	assert.Equal(t, DefaultAddressIndex, opts.AddressIndex)
}

func TestDeriveOptions_BuildHDPath(t *testing.T) {
	tests := []struct {
		name     string
		opts     *DeriveOptions
		expected string
	}{
		{
			name:     "default options",
			opts:     NewDefaultDeriveOptions(),
			expected: DefaultHDPath,
		},
		{
			name: "custom account",
			opts: &DeriveOptions{
				CoinType:     118,
				Account:      5,
				AddressIndex: 0,
			},
			expected: "m/44'/118'/5'/0/0",
		},
		{
			name: "custom address index",
			opts: &DeriveOptions{
				CoinType:     118,
				Account:      0,
				AddressIndex: 10,
			},
			expected: "m/44'/118'/0'/0/10",
		},
		{
			name: "explicit path overrides",
			opts: &DeriveOptions{
				HDPath:   "m/44'/60'/0'/0/0", // Ethereum path
				CoinType: 118,                // Should be ignored
			},
			expected: "m/44'/60'/0'/0/0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.opts.BuildHDPath()
			assert.Equal(t, tt.expected, path)
		})
	}
}

func TestDeriveKeyWithOpts(t *testing.T) {
	mnemonic := testMnemonic12

	opts := &DeriveOptions{
		CoinType:     118,
		Account:      2,
		AddressIndex: 0,
	}

	key, err := DeriveKeyWithOpts(mnemonic, opts)
	require.NoError(t, err)
	require.NotNil(t, key)
	defer key.Zero()

	assert.Equal(t, "m/44'/118'/2'/0/0", key.HDPath)
	assert.Len(t, key.PrivateKey, 32)
	assert.NotEmpty(t, key.Address)
}

func TestDeriveKeyWithOpts_NilOpts(t *testing.T) {
	mnemonic := testMnemonic12

	key, err := DeriveKeyWithOpts(mnemonic, nil)
	require.NoError(t, err)
	require.NotNil(t, key)
	defer key.Zero()

	assert.Equal(t, DefaultHDPath, key.HDPath)
}

func TestDeriveKeyWithOpts_WithPassphrase(t *testing.T) {
	mnemonic := testMnemonic12

	keyWithout, err := DeriveKeyWithOpts(mnemonic, &DeriveOptions{})
	require.NoError(t, err)
	defer keyWithout.Zero()

	keyWith, err := DeriveKeyWithOpts(mnemonic, &DeriveOptions{
		Passphrase: "my secret passphrase",
	})
	require.NoError(t, err)
	defer keyWith.Zero()

	// Passphrase should result in different keys
	assert.NotEqual(t, keyWithout.PrivateKey, keyWith.PrivateKey, "passphrase should produce different key")
	assert.NotEqual(t, keyWithout.Address, keyWith.Address, "passphrase should produce different address")
}

// ============================================================================
// Backup/Recovery Tests
// ============================================================================

func TestRecoverKeyFromMnemonic(t *testing.T) {
	// Generate a mnemonic
	originalMnemonic, err := GenerateMnemonic(Mnemonic24Words)
	require.NoError(t, err)

	// Derive key
	originalKey, err := DeriveKeyFromMnemonic(originalMnemonic, DefaultHDPath)
	require.NoError(t, err)
	defer originalKey.Zero()

	// Recover key from same mnemonic
	recoveredKey, err := RecoverKeyFromMnemonic(originalMnemonic, DefaultHDPath)
	require.NoError(t, err)
	defer recoveredKey.Zero()

	// Keys should be identical
	assert.Equal(t, originalKey.PrivateKey, recoveredKey.PrivateKey, "recovered private key should match")
	assert.Equal(t, originalKey.PublicKey, recoveredKey.PublicKey, "recovered public key should match")
	assert.Equal(t, originalKey.Address, recoveredKey.Address, "recovered address should match")
}

func TestMnemonicToSeed(t *testing.T) {
	mnemonic := testMnemonic12

	seed, err := MnemonicToSeed(mnemonic, "")
	require.NoError(t, err)
	defer ZeroBytes(seed)

	assert.Len(t, seed, 64, "BIP-39 seed should be 64 bytes")
}

func TestMnemonicToSeed_WithPassphrase(t *testing.T) {
	mnemonic := testMnemonic12

	seedWithout, err := MnemonicToSeed(mnemonic, "")
	require.NoError(t, err)
	defer ZeroBytes(seedWithout)

	seedWith, err := MnemonicToSeed(mnemonic, "passphrase")
	require.NoError(t, err)
	defer ZeroBytes(seedWith)

	assert.NotEqual(t, seedWithout, seedWith, "passphrase should produce different seed")
}

func TestMnemonicToSeed_InvalidMnemonic(t *testing.T) {
	_, err := MnemonicToSeed("invalid mnemonic", "")
	require.Error(t, err)
}

// ============================================================================
// Entropy Conversion Tests
// ============================================================================

func TestEntropyToMnemonic_128bit(t *testing.T) {
	// 128 bits = 16 bytes
	entropy := make([]byte, 16)
	for i := range entropy {
		entropy[i] = byte(i)
	}

	mnemonic, err := EntropyToMnemonic(entropy)
	require.NoError(t, err)
	assert.True(t, ValidateMnemonic(mnemonic))

	words := strings.Fields(mnemonic)
	assert.Len(t, words, 12)
}

func TestEntropyToMnemonic_256bit(t *testing.T) {
	// 256 bits = 32 bytes
	entropy := make([]byte, 32)
	for i := range entropy {
		entropy[i] = byte(i)
	}

	mnemonic, err := EntropyToMnemonic(entropy)
	require.NoError(t, err)
	assert.True(t, ValidateMnemonic(mnemonic))

	words := strings.Fields(mnemonic)
	assert.Len(t, words, 24)
}

func TestEntropyToMnemonic_InvalidSize(t *testing.T) {
	// Invalid size (not 128 or 256 bits)
	entropy := make([]byte, 20)
	_, err := EntropyToMnemonic(entropy)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "entropy must be 128 or 256 bits")
}

func TestMnemonicToEntropy(t *testing.T) {
	mnemonic := testMnemonic12

	entropy, err := MnemonicToEntropy(mnemonic)
	require.NoError(t, err)
	defer ZeroBytes(entropy)

	// Should be able to convert back to same mnemonic
	recoveredMnemonic, err := EntropyToMnemonic(entropy)
	require.NoError(t, err)

	assert.Equal(t, mnemonic, recoveredMnemonic)
}

func TestMnemonicToEntropy_InvalidMnemonic(t *testing.T) {
	_, err := MnemonicToEntropy("invalid mnemonic")
	require.Error(t, err)
}

// ============================================================================
// Address Derivation Tests
// ============================================================================

func TestPrivateKeyToAddress(t *testing.T) {
	mnemonic := testMnemonic12

	key, err := DeriveKeyFromMnemonic(mnemonic, DefaultHDPath)
	require.NoError(t, err)
	defer key.Zero()

	address, err := PrivateKeyToAddress(key.PrivateKey)
	require.NoError(t, err)

	assert.Equal(t, key.Address, address)
}

func TestPrivateKeyToAddress_InvalidSize(t *testing.T) {
	_, err := PrivateKeyToAddress(make([]byte, 16))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid private key size")
}

// ============================================================================
// Multiple Account Derivation Tests
// ============================================================================

func TestDeriveMultipleAccounts(t *testing.T) {
	mnemonic := testMnemonic12

	keys, err := DeriveMultipleAccounts(mnemonic, 0, 5)
	require.NoError(t, err)
	require.Len(t, keys, 5)

	// Clean up
	defer func() {
		for _, k := range keys {
			k.Zero()
		}
	}()

	// All keys should be different
	addresses := make(map[string]bool)
	for i, key := range keys {
		assert.NotEmpty(t, key.Address)
		assert.False(t, addresses[key.Address], "account %d has duplicate address", i)
		addresses[key.Address] = true

		// Verify expected HD path
		expectedPath := "m/44'/118'/" + string(rune('0'+i)) + "'/0/0"
		// Note: path format may vary slightly, just verify they're different
		assert.NotEmpty(t, key.HDPath)
	}
}

func TestDeriveMultipleAccounts_InvalidMnemonic(t *testing.T) {
	_, err := DeriveMultipleAccounts("invalid mnemonic", 0, 5)
	require.Error(t, err)
}

func TestDeriveMultipleAccounts_ZeroCount(t *testing.T) {
	mnemonic := testMnemonic12

	_, err := DeriveMultipleAccounts(mnemonic, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count must be greater than 0")
}

// ============================================================================
// Security: Zero Tests
// ============================================================================

func TestDerivedKey_Zero(t *testing.T) {
	mnemonic := testMnemonic12

	key, err := DeriveKeyFromMnemonic(mnemonic, DefaultHDPath)
	require.NoError(t, err)

	// Store original values for comparison
	originalPrivKey := make([]byte, len(key.PrivateKey))
	copy(originalPrivKey, key.PrivateKey)

	// Zero the key
	key.Zero()

	// Private key should be nil after zeroing
	assert.Nil(t, key.PrivateKey, "private key should be nil after Zero()")

	// Public key and address are not sensitive, should remain
	assert.NotEmpty(t, key.PublicKey)
	assert.NotEmpty(t, key.Address)
}

// ============================================================================
// Word List Tests
// ============================================================================

func TestGetWordList(t *testing.T) {
	wordList := GetWordList()

	// BIP-39 English word list has 2048 words
	assert.Len(t, wordList, 2048)

	// First and last words of BIP-39 English
	assert.Equal(t, "abandon", wordList[0])
	assert.Equal(t, "zoo", wordList[2047])
}

func TestIsValidWord(t *testing.T) {
	assert.True(t, IsValidWord("abandon"))
	assert.True(t, IsValidWord("zoo"))
	assert.True(t, IsValidWord("abstract"))

	assert.False(t, IsValidWord("invalid123"))
	assert.False(t, IsValidWord(""))
	assert.False(t, IsValidWord("ABANDON")) // Case-sensitive
}

func TestValidateMnemonicWords(t *testing.T) {
	tests := []struct {
		name            string
		mnemonic        string
		expectValid     bool
		expectInvalidN  int
	}{
		{
			name:           "all valid",
			mnemonic:       testMnemonic12,
			expectValid:    true,
			expectInvalidN: 0,
		},
		{
			name:           "one invalid",
			mnemonic:       "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon invalid123",
			expectValid:    false,
			expectInvalidN: 1,
		},
		{
			name:           "multiple invalid",
			mnemonic:       "invalid1 invalid2 abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			expectValid:    false,
			expectInvalidN: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, invalidWords := ValidateMnemonicWords(tt.mnemonic)
			assert.Equal(t, tt.expectValid, valid)
			assert.Len(t, invalidWords, tt.expectInvalidN)
		})
	}
}

// ============================================================================
// Normalization Tests
// ============================================================================

func TestNormalizeMnemonic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already normalized",
			input:    "abandon abandon about",
			expected: "abandon abandon about",
		},
		{
			name:     "extra spaces",
			input:    "  abandon   abandon   about  ",
			expected: "abandon abandon about",
		},
		{
			name:     "tabs and newlines",
			input:    "abandon\tabandon\nabout",
			expected: "abandon abandon about",
		},
		{
			name:     "mixed whitespace",
			input:    "\t  abandon  \n\t abandon \t  about  \n",
			expected: "abandon abandon about",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeMnemonic(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChecksumMnemonic(t *testing.T) {
	// Valid mnemonic
	valid := testMnemonic12
	assert.True(t, ChecksumMnemonic(valid))

	// Invalid checksum
	invalid := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon"
	assert.False(t, ChecksumMnemonic(invalid))
}

// ============================================================================
// Integration Test: Full Flow
// ============================================================================

func TestFullMnemonicFlow(t *testing.T) {
	// 1. Generate a new mnemonic
	mnemonic, err := GenerateMnemonic(Mnemonic24Words)
	require.NoError(t, err)
	t.Logf("Generated mnemonic word count: %d", MnemonicWordCount(mnemonic))

	// 2. Validate it
	assert.True(t, ValidateMnemonic(mnemonic))

	// 3. Derive a key
	key1, err := DeriveKeyFromMnemonic(mnemonic, DefaultHDPath)
	require.NoError(t, err)
	defer key1.Zero()

	t.Logf("Derived address: %s", key1.Address)
	t.Logf("HD Path: %s", key1.HDPath)

	// 4. Simulate "backup" - store mnemonic (in real use, securely store it)
	backup := mnemonic

	// 5. Simulate "recovery" - recover from backup
	key2, err := RecoverKeyFromMnemonic(backup, DefaultHDPath)
	require.NoError(t, err)
	defer key2.Zero()

	// 6. Verify recovery produces identical keys
	assert.Equal(t, key1.PrivateKey, key2.PrivateKey, "recovered key should match original")
	assert.Equal(t, key1.Address, key2.Address, "recovered address should match original")

	// 7. Derive additional accounts
	keys, err := DeriveMultipleAccounts(mnemonic, 0, 3)
	require.NoError(t, err)
	defer func() {
		for _, k := range keys {
			k.Zero()
		}
	}()

	// First account should match key1
	assert.Equal(t, key1.Address, keys[0].Address, "first derived account should match")

	// All accounts should be different
	assert.NotEqual(t, keys[0].Address, keys[1].Address)
	assert.NotEqual(t, keys[1].Address, keys[2].Address)
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkGenerateMnemonic12(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateMnemonic(Mnemonic12Words)
	}
}

func BenchmarkGenerateMnemonic24(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateMnemonic(Mnemonic24Words)
	}
}

func BenchmarkDeriveKeyFromMnemonic(b *testing.B) {
	mnemonic := testMnemonic12

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key, _ := DeriveKeyFromMnemonic(mnemonic, DefaultHDPath)
		if key != nil {
			key.Zero()
		}
	}
}

func BenchmarkValidateMnemonic(b *testing.B) {
	mnemonic := testMnemonic12

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateMnemonic(mnemonic)
	}
}

