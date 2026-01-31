// Package edugain provides EduGAIN federation integration.
//
// VE-2005: Tests for XML Encryption decryption
package edugain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Encryption Helper Functions for Tests
// ============================================================================

// generateTestRSAKeyPair generates a test RSA key pair
func generateTestRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// encryptSessionKeyRSAOAEP encrypts a session key with RSA-OAEP using SHA-256
func encryptSessionKeyRSAOAEP(sessionKey []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, sessionKey, nil)
}

// encryptDataAESCBC encrypts data with AES-CBC
func encryptDataAESCBC(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// PKCS7 padding
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	padded := make([]byte, len(plaintext)+padding)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	// Generate IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	// Encrypt
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(padded))
	mode.CryptBlocks(ciphertext, padded)

	// Prepend IV
	return append(iv, ciphertext...), nil
}

// encryptDataAESGCM encrypts data with AES-GCM
func encryptDataAESGCM(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return append(nonce, ciphertext...), nil
}

// createEncryptedAssertionXML creates a mock encrypted assertion for testing
func createEncryptedAssertionXML(
	encryptedKey []byte,
	encryptedData []byte,
	keyAlgorithm string,
	dataAlgorithm string,
) []byte {
	encKeyB64 := base64.StdEncoding.EncodeToString(encryptedKey)
	encDataB64 := base64.StdEncoding.EncodeToString(encryptedData)

	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<EncryptedAssertion xmlns="urn:oasis:names:tc:SAML:2.0:assertion">
    <EncryptedData xmlns="http://www.w3.org/2001/04/xmlenc#" Type="http://www.w3.org/2001/04/xmlenc#Element">
        <EncryptionMethod Algorithm="%s"/>
        <KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
            <EncryptedKey xmlns="http://www.w3.org/2001/04/xmlenc#">
                <EncryptionMethod Algorithm="%s"/>
                <CipherData>
                    <CipherValue>%s</CipherValue>
                </CipherData>
            </EncryptedKey>
        </KeyInfo>
        <CipherData>
            <CipherValue>%s</CipherValue>
        </CipherData>
    </EncryptedData>
</EncryptedAssertion>`,
		dataAlgorithm,
		keyAlgorithm,
		encKeyB64,
		encDataB64,
	))
}

// ============================================================================
// AES-CBC Decryption Tests
// ============================================================================

func TestDecryptAESCBC_Valid(t *testing.T) {
	key := make([]byte, 32) // 256-bit key
	rand.Read(key)

	plaintext := []byte("Test SAML assertion content that needs to be encrypted")

	// Encrypt
	ciphertext, err := encryptDataAESCBC(plaintext, key)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptAESCBC(ciphertext, key)
	require.NoError(t, err)

	assert.Equal(t, plaintext, decrypted)
}

func TestDecryptAESCBC_TooShort(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	// Ciphertext shorter than block size
	_, err := decryptAESCBC(make([]byte, 10), key)
	assert.Error(t, err)
}

func TestDecryptAESCBC_NotBlockAligned(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	// Ciphertext not aligned to block size (after removing IV)
	ciphertext := make([]byte, aes.BlockSize+10) // IV + non-aligned data
	rand.Read(ciphertext)

	_, err := decryptAESCBC(ciphertext, key)
	assert.Error(t, err)
}

func TestDecryptAESCBC_InvalidPadding(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	// Create ciphertext with invalid padding
	block, _ := aes.NewCipher(key)
	iv := make([]byte, aes.BlockSize)
	rand.Read(iv)

	// Create block with invalid padding (padding byte > block size)
	plainWithBadPadding := make([]byte, aes.BlockSize)
	plainWithBadPadding[aes.BlockSize-1] = 20 // Invalid: > 16

	mode := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, aes.BlockSize)
	mode.CryptBlocks(encrypted, plainWithBadPadding)

	ciphertext := append(iv, encrypted...)

	_, err := decryptAESCBC(ciphertext, key)
	assert.Error(t, err)
}

// ============================================================================
// AES-GCM Decryption Tests
// ============================================================================

func TestDecryptAESGCM_Valid(t *testing.T) {
	key := make([]byte, 32) // 256-bit key
	rand.Read(key)

	plaintext := []byte("Test SAML assertion content for GCM encryption")

	// Encrypt
	ciphertext, err := encryptDataAESGCM(plaintext, key)
	require.NoError(t, err)

	// Decrypt
	decrypted, err := decryptAESGCM(ciphertext, key)
	require.NoError(t, err)

	assert.Equal(t, plaintext, decrypted)
}

func TestDecryptAESGCM_TooShort(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	// Ciphertext shorter than nonce size
	_, err := decryptAESGCM(make([]byte, 5), key)
	assert.Error(t, err)
}

func TestDecryptAESGCM_TamperedData(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	plaintext := []byte("Original message")

	// Encrypt
	ciphertext, err := encryptDataAESGCM(plaintext, key)
	require.NoError(t, err)

	// Tamper with ciphertext
	if len(ciphertext) > 20 {
		ciphertext[20] ^= 0xFF
	}

	// Decrypt should fail due to authentication
	_, err = decryptAESGCM(ciphertext, key)
	assert.Error(t, err, "GCM should detect tampering")
}

// ============================================================================
// PKCS7 Unpadding Tests
// ============================================================================

func TestPKCS7Unpad_Valid(t *testing.T) {
	// Valid padding of 4
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 4, 4, 4, 4}
	result, err := pkcs7Unpad(data)
	require.NoError(t, err)
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, result)
}

func TestPKCS7Unpad_FullBlock(t *testing.T) {
	// Full block of padding (16)
	data := make([]byte, 16)
	for i := range data {
		data[i] = 16
	}
	result, err := pkcs7Unpad(data)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestPKCS7Unpad_Empty(t *testing.T) {
	_, err := pkcs7Unpad([]byte{})
	assert.Error(t, err)
}

func TestPKCS7Unpad_ZeroPadding(t *testing.T) {
	data := []byte{1, 2, 3, 0}
	_, err := pkcs7Unpad(data)
	assert.Error(t, err)
}

func TestPKCS7Unpad_PaddingTooLarge(t *testing.T) {
	data := []byte{1, 2, 3, 20}
	_, err := pkcs7Unpad(data)
	assert.Error(t, err)
}

func TestPKCS7Unpad_InconsistentPadding(t *testing.T) {
	// Padding says 3, but bytes don't match
	data := []byte{1, 2, 3, 4, 5, 3, 3, 2}
	_, err := pkcs7Unpad(data)
	assert.Error(t, err)
}

// ============================================================================
// RSA-OAEP Decryption Tests
// ============================================================================

func TestRSAOAEPDecrypt_Valid(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	// Encrypt session key with SHA-256
	encryptedKey, err := encryptSessionKeyRSAOAEP(sessionKey, publicKey)
	require.NoError(t, err)

	// Create decryptor with private key
	// We need to serialize the key first
	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	// Decrypt using SHA-256 algorithm
	decrypted, err := decryptor.rsaOAEPDecrypt(encryptedKey, KeyTransportAlgorithmRSAOAEPSHA256)
	require.NoError(t, err)

	assert.Equal(t, sessionKey, decrypted)
}

func TestRSAOAEPDecrypt_SHA256(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	// Encrypt with SHA-256
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, sessionKey, nil)
	require.NoError(t, err)

	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	// Decrypt with SHA-256 algorithm
	decrypted, err := decryptor.rsaOAEPDecrypt(encryptedKey, KeyTransportAlgorithmRSAOAEPSHA256)
	require.NoError(t, err)

	assert.Equal(t, sessionKey, decrypted)
}

func TestRSAOAEPDecrypt_WrongKey(t *testing.T) {
	privateKey1, _, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	_, publicKey2, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	// Encrypt with key 2
	encryptedKey, err := encryptSessionKeyRSAOAEP(sessionKey, publicKey2)
	require.NoError(t, err)

	// Try to decrypt with key 1
	keyDER := x509.MarshalPKCS1PrivateKey(privateKey1)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	_, err = decryptor.rsaOAEPDecrypt(encryptedKey, KeyTransportAlgorithmRSAOAEPSHA256)
	assert.Error(t, err, "should fail with wrong key")
}

// ============================================================================
// Full Decryption Flow Tests
// ============================================================================

func TestDecryptAssertion_AES256CBC(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	// Session key for AES-256
	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	// Encrypt session key with SHA-256
	encryptedKey, err := encryptSessionKeyRSAOAEP(sessionKey, publicKey)
	require.NoError(t, err)

	// Original assertion
	assertion := []byte(`<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">
		<saml:Subject>testuser@example.com</saml:Subject>
	</saml:Assertion>`)

	// Encrypt assertion
	encryptedData, err := encryptDataAESCBC(assertion, sessionKey)
	require.NoError(t, err)

	// Create encrypted XML with SHA-256 algorithm
	encryptedXML := createEncryptedAssertionXML(
		encryptedKey,
		encryptedData,
		KeyTransportAlgorithmRSAOAEPSHA256,
		EncryptionAlgorithmAES256CBC,
	)

	// Decrypt
	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	decrypted, err := decryptor.DecryptAssertion(encryptedXML)
	require.NoError(t, err)

	assert.Equal(t, assertion, decrypted)
}

func TestDecryptAssertion_AES256GCM(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	// Session key for AES-256
	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	// Encrypt session key with SHA-256
	encryptedKey, err := encryptSessionKeyRSAOAEP(sessionKey, publicKey)
	require.NoError(t, err)

	// Original assertion
	assertion := []byte(`<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">
		<saml:Subject>testuser@example.com</saml:Subject>
	</saml:Assertion>`)

	// Encrypt assertion with GCM
	encryptedData, err := encryptDataAESGCM(assertion, sessionKey)
	require.NoError(t, err)

	// Create encrypted XML with SHA-256 algorithm
	encryptedXML := createEncryptedAssertionXML(
		encryptedKey,
		encryptedData,
		KeyTransportAlgorithmRSAOAEPSHA256,
		EncryptionAlgorithmAES256GCM,
	)

	// Decrypt
	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	decrypted, err := decryptor.DecryptAssertion(encryptedXML)
	require.NoError(t, err)

	assert.Equal(t, assertion, decrypted)
}

func TestDecryptAssertion_EmptyData(t *testing.T) {
	privateKey, _, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	_, err = decryptor.DecryptAssertion([]byte{})
	assert.Error(t, err)
}

func TestDecryptAssertion_InvalidXML(t *testing.T) {
	privateKey, _, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	_, err = decryptor.DecryptAssertion([]byte("not xml"))
	assert.Error(t, err)
}

func TestDecryptAssertion_NoEncryptedData(t *testing.T) {
	privateKey, _, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	xml := []byte(`<saml:Assertion><saml:Subject>test</saml:Subject></saml:Assertion>`)
	_, err = decryptor.DecryptAssertion(xml)
	assert.Error(t, err)
}

// ============================================================================
// Algorithm Constant Tests
// ============================================================================

func TestEncryptionAlgorithmConstants(t *testing.T) {
	// Verify algorithm URIs are correct
	assert.Equal(t, "http://www.w3.org/2001/04/xmlenc#aes128-cbc", EncryptionAlgorithmAES128CBC)
	assert.Equal(t, "http://www.w3.org/2001/04/xmlenc#aes256-cbc", EncryptionAlgorithmAES256CBC)
	assert.Equal(t, "http://www.w3.org/2009/xmlenc11#aes128-gcm", EncryptionAlgorithmAES128GCM)
	assert.Equal(t, "http://www.w3.org/2009/xmlenc11#aes256-gcm", EncryptionAlgorithmAES256GCM)
	assert.Equal(t, "http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p", KeyTransportAlgorithmRSAOAEP)
	assert.Equal(t, "http://www.w3.org/2001/04/xmlenc#rsa-1_5", KeyTransportAlgorithmRSA15)
}

// ============================================================================
// Weak Algorithm Rejection Tests
// ============================================================================

func TestDecryptAssertion_RejectsRSA15(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	// Note: We can't actually encrypt with RSA 1.5 safely, but we test that
	// the algorithm is rejected by creating XML with that algorithm URI
	encryptedKey, err := encryptSessionKeyRSAOAEP(sessionKey, publicKey)
	require.NoError(t, err)

	encryptedData, err := encryptDataAESCBC([]byte("test"), sessionKey)
	require.NoError(t, err)

	// Create encrypted XML with RSA 1.5 algorithm (should be rejected)
	encryptedXML := createEncryptedAssertionXML(
		encryptedKey,
		encryptedData,
		KeyTransportAlgorithmRSA15, // Weak algorithm
		EncryptionAlgorithmAES256CBC,
	)

	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	_, err = decryptor.DecryptAssertion(encryptedXML)
	assert.Error(t, err, "RSA 1.5 should be rejected")
	assert.Contains(t, err.Error(), "RSA 1.5")
}

func TestDecryptAssertion_RejectsSHA1OAEP(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	// Encrypt session key (note: the actual encryption uses SHA-256 now,
	// but we test that the algorithm URI is rejected)
	encryptedKey, err := encryptSessionKeyRSAOAEP(sessionKey, publicKey)
	require.NoError(t, err)

	encryptedData, err := encryptDataAESCBC([]byte("test"), sessionKey)
	require.NoError(t, err)

	// Create encrypted XML with SHA-1 based OAEP algorithm (should be rejected)
	encryptedXML := createEncryptedAssertionXML(
		encryptedKey,
		encryptedData,
		KeyTransportAlgorithmRSAOAEP, // SHA-1 based - weak
		EncryptionAlgorithmAES256CBC,
	)

	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	_, err = decryptor.DecryptAssertion(encryptedXML)
	assert.Error(t, err, "SHA-1 based OAEP should be rejected")
	assert.Contains(t, err.Error(), "SHA-1")
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestDecryptAssertion_UnsupportedDataAlgorithm(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	encryptedKey, err := encryptSessionKeyRSAOAEP(sessionKey, publicKey)
	require.NoError(t, err)

	encryptedData := make([]byte, 100)
	rand.Read(encryptedData)

	// Create encrypted XML with unsupported algorithm
	encryptedXML := createEncryptedAssertionXML(
		encryptedKey,
		encryptedData,
		KeyTransportAlgorithmRSAOAEPSHA256,
		"http://example.com/unsupported-algorithm",
	)

	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	_, err = decryptor.DecryptAssertion(encryptedXML)
	assert.Error(t, err, "unsupported algorithm should fail")
}

func TestDecryptAssertion_TamperedEncryptedKey(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	encryptedKey, err := encryptSessionKeyRSAOAEP(sessionKey, publicKey)
	require.NoError(t, err)

	// Tamper with encrypted key
	if len(encryptedKey) > 10 {
		encryptedKey[10] ^= 0xFF
	}

	encryptedData, err := encryptDataAESCBC([]byte("test"), sessionKey)
	require.NoError(t, err)

	encryptedXML := createEncryptedAssertionXML(
		encryptedKey,
		encryptedData,
		KeyTransportAlgorithmRSAOAEPSHA256,
		EncryptionAlgorithmAES256CBC,
	)

	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	decryptor, err := NewAssertionDecryptor(keyDER)
	require.NoError(t, err)

	_, err = decryptor.DecryptAssertion(encryptedXML)
	assert.Error(t, err, "tampered encrypted key should fail")
}

// ============================================================================
// DecryptXMLEncryptionWithKey Tests
// ============================================================================

func TestDecryptXMLEncryptionWithKey_Valid(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKeyPair()
	require.NoError(t, err)

	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	encryptedKey, err := encryptSessionKeyRSAOAEP(sessionKey, publicKey)
	require.NoError(t, err)

	assertion := []byte("test assertion")
	encryptedData, err := encryptDataAESCBC(assertion, sessionKey)
	require.NoError(t, err)

	encryptedXML := createEncryptedAssertionXML(
		encryptedKey,
		encryptedData,
		KeyTransportAlgorithmRSAOAEPSHA256,
		EncryptionAlgorithmAES256CBC,
	)

	keyDER := x509.MarshalPKCS1PrivateKey(privateKey)

	decrypted, err := DecryptXMLEncryptionWithKey(encryptedXML, keyDER)
	require.NoError(t, err)

	assert.Equal(t, assertion, decrypted)
}

func TestNewAssertionDecryptor_InvalidKey(t *testing.T) {
	_, err := NewAssertionDecryptor([]byte("not a valid key"))
	assert.Error(t, err)
}

func TestNewAssertionDecryptor_EmptyKey(t *testing.T) {
	_, err := NewAssertionDecryptor([]byte{})
	assert.Error(t, err)
}
