package crypto

import (
	"crypto/sha512"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
)

// ============================================================================
// BIP-39 Mnemonic Seed Generation
// ============================================================================
//
// SECURITY CRITICAL: This module handles sensitive cryptographic material.
// - Mnemonics and private keys are NEVER logged
// - All sensitive data should be zeroed after use
// - Caller is responsible for secure storage of generated mnemonics

// Error messages (constants to avoid duplication)
const (
	errMsgInvalidMnemonic    = "invalid mnemonic"
	errMsgFailedGenerateSeed = "failed to generate seed: %w"
)

// MnemonicSize represents the valid mnemonic word counts per BIP-39
type MnemonicSize int

const (
	// Mnemonic12Words generates a 12-word mnemonic (128-bit entropy)
	Mnemonic12Words MnemonicSize = 12

	// Mnemonic24Words generates a 24-word mnemonic (256-bit entropy)
	Mnemonic24Words MnemonicSize = 24
)

// EntropySizeForMnemonic returns the required entropy size in bits for a given mnemonic size
func EntropySizeForMnemonic(size MnemonicSize) (int, error) {
	switch size {
	case Mnemonic12Words:
		return 128, nil
	case Mnemonic24Words:
		return 256, nil
	default:
		return 0, fmt.Errorf("invalid mnemonic size: must be 12 or 24 words, got %d", size)
	}
}

// IsValidMnemonicSize checks if the mnemonic size is valid
func IsValidMnemonicSize(size MnemonicSize) bool {
	return size == Mnemonic12Words || size == Mnemonic24Words
}

// Cosmos SDK / BIP-44 derivation path constants
const (
	// DefaultCoinType is the Cosmos coin type (118) per SLIP-44
	DefaultCoinType uint32 = sdk.CoinType

	// DefaultAccountIndex is the default account index
	DefaultAccountIndex uint32 = 0

	// DefaultAddressIndex is the default address index
	DefaultAddressIndex uint32 = 0

	// DefaultHDPath is the standard Cosmos SDK HD derivation path
	// m/44'/118'/0'/0/0
	DefaultHDPath = "m/44'/118'/0'/0/0"
)

// DerivedKey represents a derived key pair from a mnemonic
type DerivedKey struct {
	// PrivateKey is the derived secp256k1 private key (32 bytes)
	// SECURITY: Must be zeroed after use
	PrivateKey []byte

	// PublicKey is the compressed secp256k1 public key (33 bytes)
	PublicKey []byte

	// Address is the Cosmos SDK bech32 address
	Address string

	// HDPath is the derivation path used
	HDPath string
}

// Zero securely zeros the private key material
func (dk *DerivedKey) Zero() {
	if dk.PrivateKey != nil {
		ZeroBytes(dk.PrivateKey)
		dk.PrivateKey = nil
	}
}

// GenerateMnemonic generates a new BIP-39 mnemonic seed phrase
// SECURITY: The returned mnemonic must be stored securely and never logged
func GenerateMnemonic(size MnemonicSize) (string, error) {
	entropyBits, err := EntropySizeForMnemonic(size)
	if err != nil {
		return "", err
	}

	// Generate entropy
	entropy, err := bip39.NewEntropy(entropyBits)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	// Generate mnemonic from entropy
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		// Zero entropy before returning
		ZeroBytes(entropy)
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	// Zero entropy after use
	ZeroBytes(entropy)

	return mnemonic, nil
}

// ValidateMnemonic validates that a mnemonic is a valid BIP-39 mnemonic
func ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// MnemonicWordCount returns the number of words in a mnemonic
func MnemonicWordCount(mnemonic string) int {
	if mnemonic == "" {
		return 0
	}

	words := splitMnemonicWords(mnemonic)
	return len(words)
}

// splitMnemonicWords splits a mnemonic into words, handling various whitespace
func splitMnemonicWords(mnemonic string) []string {
	var words []string
	word := ""
	for _, r := range mnemonic {
		if r == ' ' || r == '\t' || r == '\n' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(r)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

// DeriveKeyFromMnemonic derives a key pair from a mnemonic using the specified HD path
// SECURITY: The caller must call Zero() on the returned DerivedKey when done
func DeriveKeyFromMnemonic(mnemonic string, hdPath string) (*DerivedKey, error) {
	if !ValidateMnemonic(mnemonic) {
		return nil, fmt.Errorf(errMsgInvalidMnemonic)
	}

	if hdPath == "" {
		hdPath = DefaultHDPath
	}

	// Validate HD path format
	if _, err := hd.NewParamsFromPath(hdPath); err != nil {
		return nil, fmt.Errorf("invalid HD path %q: %w", hdPath, err)
	}

	// Generate seed from mnemonic (no passphrase for standard Cosmos)
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, fmt.Errorf(errMsgFailedGenerateSeed, err)
	}
	defer ZeroBytes(seed)

	// Derive the master key and chain code using HMAC-SHA512
	master, chainCode := deriveMasterKey(seed)
	defer ZeroBytes(master)
	defer ZeroBytes(chainCode)

	// Parse HD path and derive child keys
	params, _ := hd.NewParamsFromPath(hdPath)

	// Derive private key using Cosmos SDK's derivation
	derivedPrivKey, err := hd.Secp256k1.Derive()(mnemonic, "", hdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	// Generate the private key object
	privKey := hd.Secp256k1.Generate()(derivedPrivKey)

	// Get public key
	pubKey := privKey.PubKey()

	// Generate address
	address := sdk.AccAddress(pubKey.Address()).String()

	return &DerivedKey{
		PrivateKey: derivedPrivKey,
		PublicKey:  pubKey.Bytes(),
		Address:    address,
		HDPath:     params.String(),
	}, nil
}

// deriveMasterKey derives the master key from seed using HMAC-SHA512
// This follows BIP-32 specification
func deriveMasterKey(seed []byte) ([]byte, []byte) {
	hmac := hmacSHA512([]byte("Bitcoin seed"), seed)
	return hmac[:32], hmac[32:]
}

// hmacSHA512 computes HMAC-SHA512
func hmacSHA512(key, data []byte) []byte {
	h := sha512.New()
	blockSize := h.BlockSize()

	// Key processing
	if len(key) > blockSize {
		h.Write(key)
		key = h.Sum(nil)
		h.Reset()
	}
	if len(key) < blockSize {
		padded := make([]byte, blockSize)
		copy(padded, key)
		key = padded
	}

	// Inner and outer padding
	ipad := make([]byte, blockSize)
	opad := make([]byte, blockSize)
	for i := 0; i < blockSize; i++ {
		ipad[i] = key[i] ^ 0x36
		opad[i] = key[i] ^ 0x5c
	}

	// Inner hash
	h.Write(ipad)
	h.Write(data)
	innerHash := h.Sum(nil)
	h.Reset()

	// Outer hash
	h.Write(opad)
	h.Write(innerHash)
	return h.Sum(nil)
}

// DeriveKeyWithOptions provides advanced key derivation with custom options
type DeriveOptions struct {
	// HDPath is the derivation path (default: m/44'/118'/0'/0/0)
	HDPath string

	// Passphrase is an optional BIP-39 passphrase
	// SECURITY: This adds additional protection to the seed derivation
	Passphrase string

	// CoinType overrides the coin type in the path (default: 118 for Cosmos)
	CoinType uint32

	// Account is the account index (default: 0)
	Account uint32

	// AddressIndex is the address index (default: 0)
	AddressIndex uint32
}

// NewDefaultDeriveOptions creates default derivation options for Cosmos
func NewDefaultDeriveOptions() *DeriveOptions {
	return &DeriveOptions{
		HDPath:       DefaultHDPath,
		Passphrase:   "",
		CoinType:     DefaultCoinType,
		Account:      DefaultAccountIndex,
		AddressIndex: DefaultAddressIndex,
	}
}

// BuildHDPath constructs an HD path from options
func (o *DeriveOptions) BuildHDPath() string {
	if o.HDPath != "" {
		return o.HDPath
	}
	return fmt.Sprintf("m/44'/%d'/%d'/0/%d", o.CoinType, o.Account, o.AddressIndex)
}

// DeriveKeyWithOpts derives a key with custom options
// SECURITY: The caller must call Zero() on the returned DerivedKey when done
func DeriveKeyWithOpts(mnemonic string, opts *DeriveOptions) (*DerivedKey, error) {
	if opts == nil {
		opts = NewDefaultDeriveOptions()
	}

	if !ValidateMnemonic(mnemonic) {
		return nil, fmt.Errorf(errMsgInvalidMnemonic)
	}

	hdPath := opts.BuildHDPath()

	// Validate HD path format
	if _, err := hd.NewParamsFromPath(hdPath); err != nil {
		return nil, fmt.Errorf("invalid HD path %q: %w", hdPath, err)
	}

	// Generate seed from mnemonic with optional passphrase
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, opts.Passphrase)
	if err != nil {
		return nil, fmt.Errorf(errMsgFailedGenerateSeed, err)
	}
	defer ZeroBytes(seed)

	// Derive private key using Cosmos SDK's derivation
	derivedPrivKey, err := hd.Secp256k1.Derive()(mnemonic, opts.Passphrase, hdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	// Generate the private key object
	privKey := hd.Secp256k1.Generate()(derivedPrivKey)

	// Get public key
	pubKey := privKey.PubKey()

	// Generate address
	address := sdk.AccAddress(pubKey.Address()).String()

	return &DerivedKey{
		PrivateKey: derivedPrivKey,
		PublicKey:  pubKey.Bytes(),
		Address:    address,
		HDPath:     hdPath,
	}, nil
}

// RecoverKeyFromMnemonic recovers a key from a mnemonic backup
// This is the primary backup/recovery function
// SECURITY: The caller must call Zero() on the returned DerivedKey when done
func RecoverKeyFromMnemonic(mnemonic string, hdPath string) (*DerivedKey, error) {
	return DeriveKeyFromMnemonic(mnemonic, hdPath)
}

// MnemonicToSeed converts a mnemonic to a seed for advanced use cases
// SECURITY: The returned seed must be zeroed after use
func MnemonicToSeed(mnemonic string, passphrase string) ([]byte, error) {
	if !ValidateMnemonic(mnemonic) {
		return nil, fmt.Errorf(errMsgInvalidMnemonic)
	}

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, passphrase)
	if err != nil {
		return nil, fmt.Errorf(errMsgFailedGenerateSeed, err)
	}

	return seed, nil
}

// EntropyToMnemonic converts raw entropy bytes to a mnemonic
// SECURITY: Entropy must be securely generated (128 or 256 bits)
func EntropyToMnemonic(entropy []byte) (string, error) {
	entropyBits := len(entropy) * 8
	if entropyBits != 128 && entropyBits != 256 {
		return "", fmt.Errorf("entropy must be 128 or 256 bits, got %d bits", entropyBits)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic from entropy: %w", err)
	}

	return mnemonic, nil
}

// MnemonicToEntropy converts a mnemonic back to its original entropy
// SECURITY: The returned entropy is sensitive material
func MnemonicToEntropy(mnemonic string) ([]byte, error) {
	if !ValidateMnemonic(mnemonic) {
		return nil, fmt.Errorf(errMsgInvalidMnemonic)
	}

	entropy, err := bip39.MnemonicToByteArray(mnemonic, true)
	if err != nil {
		return nil, fmt.Errorf("failed to extract entropy: %w", err)
	}

	return entropy, nil
}

// PrivateKeyToAddress converts a private key to a Cosmos address
func PrivateKeyToAddress(privateKey []byte) (string, error) {
	if len(privateKey) != 32 {
		return "", fmt.Errorf("invalid private key size: expected 32 bytes, got %d", len(privateKey))
	}

	privKey := &secp256k1.PrivKey{Key: privateKey}
	pubKey := privKey.PubKey()
	address := sdk.AccAddress(pubKey.Address()).String()

	return address, nil
}

// DeriveMultipleAccounts derives multiple accounts from a single mnemonic
// This is useful for HD wallet implementations
// SECURITY: All returned DerivedKeys must be zeroed after use
func DeriveMultipleAccounts(mnemonic string, startIndex, count uint32) ([]*DerivedKey, error) {
	if !ValidateMnemonic(mnemonic) {
		return nil, fmt.Errorf(errMsgInvalidMnemonic)
	}

	if count == 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}

	keys := make([]*DerivedKey, 0, count)

	for i := uint32(0); i < count; i++ {
		opts := &DeriveOptions{
			CoinType:     DefaultCoinType,
			Account:      startIndex + i,
			AddressIndex: 0,
		}

		key, err := DeriveKeyWithOpts(mnemonic, opts)
		if err != nil {
			// Zero all previously derived keys on error
			for _, k := range keys {
				k.Zero()
			}
			return nil, fmt.Errorf("failed to derive account %d: %w", startIndex+i, err)
		}

		keys = append(keys, key)
	}

	return keys, nil
}

// ChecksumMnemonic verifies that a mnemonic has a valid checksum
func ChecksumMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// NormalizeMnemonic normalizes a mnemonic by removing extra whitespace
func NormalizeMnemonic(mnemonic string) string {
	words := splitMnemonicWords(mnemonic)
	result := ""
	for i, word := range words {
		if i > 0 {
			result += " "
		}
		result += word
	}
	return result
}

// GetWordList returns the BIP-39 English word list
func GetWordList() []string {
	return bip39.GetWordList()
}

// IsValidWord checks if a word is in the BIP-39 word list
func IsValidWord(word string) bool {
	wordList := bip39.GetWordList()
	for _, w := range wordList {
		if w == word {
			return true
		}
	}
	return false
}

// ValidateMnemonicWords validates that all words in a mnemonic are valid
func ValidateMnemonicWords(mnemonic string) (bool, []string) {
	words := splitMnemonicWords(mnemonic)
	invalidWords := []string{}

	for _, word := range words {
		if !IsValidWord(word) {
			invalidWords = append(invalidWords, word)
		}
	}

	return len(invalidWords) == 0, invalidWords
}
