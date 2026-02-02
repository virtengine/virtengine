// Package edugain provides EduGAIN federation integration.
//
// VE-2005: XML Encryption decryption for encrypted SAML assertions
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
	"hash"
	"strings"

	"github.com/beevik/etree"
)

// ============================================================================
// XML Encryption Constants
// ============================================================================

// Encryption algorithm URIs
const (
	// EncryptionAlgorithmAES128CBC is AES-128-CBC encryption
	EncryptionAlgorithmAES128CBC = "http://www.w3.org/2001/04/xmlenc#aes128-cbc"

	// EncryptionAlgorithmAES192CBC is AES-192-CBC encryption
	EncryptionAlgorithmAES192CBC = "http://www.w3.org/2001/04/xmlenc#aes192-cbc"

	// EncryptionAlgorithmAES256CBC is AES-256-CBC encryption
	EncryptionAlgorithmAES256CBC = "http://www.w3.org/2001/04/xmlenc#aes256-cbc"

	// EncryptionAlgorithmAES128GCM is AES-128-GCM encryption
	EncryptionAlgorithmAES128GCM = "http://www.w3.org/2009/xmlenc11#aes128-gcm"

	// EncryptionAlgorithmAES256GCM is AES-256-GCM encryption
	EncryptionAlgorithmAES256GCM = "http://www.w3.org/2009/xmlenc11#aes256-gcm"

	// KeyTransportAlgorithmRSAOAEP is RSA-OAEP key transport
	KeyTransportAlgorithmRSAOAEP = "http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p"

	// KeyTransportAlgorithmRSAOAEPSHA256 is RSA-OAEP with SHA-256
	KeyTransportAlgorithmRSAOAEPSHA256 = "http://www.w3.org/2009/xmlenc11#rsa-oaep"

	// KeyTransportAlgorithmRSA15 is RSA 1.5 (WEAK - not recommended)
	KeyTransportAlgorithmRSA15 = "http://www.w3.org/2001/04/xmlenc#rsa-1_5"
)

// XML Encryption namespace constants
const (
	xmlEncNS   = "http://www.w3.org/2001/04/xmlenc#"
	xmlEncNS11 = "http://www.w3.org/2009/xmlenc11#"
	xmlDSigNS  = "http://www.w3.org/2000/09/xmldsig#"
	samlNS     = "urn:oasis:names:tc:SAML:2.0:assertion"
)

// ============================================================================
// Assertion Decryptor
// ============================================================================

// AssertionDecryptor handles decryption of encrypted SAML assertions
type AssertionDecryptor struct {
	privateKey *rsa.PrivateKey
	certCache  *CertificateCache
}

// NewAssertionDecryptor creates a new assertion decryptor
func NewAssertionDecryptor(privateKeyPEM []byte) (*AssertionDecryptor, error) {
	key, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &AssertionDecryptor{
		privateKey: key,
		certCache:  GetGlobalCertificateCache(),
	}, nil
}

// DecryptAssertion decrypts an encrypted SAML assertion
func (d *AssertionDecryptor) DecryptAssertion(encryptedData []byte) ([]byte, error) {
	if len(encryptedData) == 0 {
		return nil, fmt.Errorf("empty encrypted data")
	}

	// Parse the encrypted data as XML
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(encryptedData); err != nil {
		return nil, fmt.Errorf("failed to parse encrypted XML: %w", err)
	}

	root := doc.Root()
	if root == nil {
		return nil, fmt.Errorf("encrypted XML has no root element")
	}

	// Find EncryptedData element
	encryptedDataEl := findEncryptedData(root)
	if encryptedDataEl == nil {
		return nil, fmt.Errorf("no EncryptedData element found")
	}

	// Extract encryption algorithm
	encryptionAlgorithm, err := getEncryptionAlgorithm(encryptedDataEl)
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption algorithm: %w", err)
	}

	// Find and decrypt the session key
	sessionKey, err := d.decryptSessionKey(encryptedDataEl)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt session key: %w", err)
	}

	// Get the cipher value (encrypted data)
	cipherValue, err := getCipherValue(encryptedDataEl)
	if err != nil {
		return nil, fmt.Errorf("failed to get cipher value: %w", err)
	}

	// Decrypt the data
	plaintext, err := decryptData(cipherValue, sessionKey, encryptionAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// decryptSessionKey decrypts the session key from the EncryptedKey element
func (d *AssertionDecryptor) decryptSessionKey(encryptedDataEl *etree.Element) ([]byte, error) {
	// Find KeyInfo element
	keyInfoEl := encryptedDataEl.FindElement("./KeyInfo")
	if keyInfoEl == nil {
		keyInfoEl = encryptedDataEl.FindElement(".//ds:KeyInfo")
	}
	if keyInfoEl == nil {
		return nil, fmt.Errorf("no KeyInfo element found")
	}

	// Find EncryptedKey element
	encryptedKeyEl := keyInfoEl.FindElement(".//EncryptedKey")
	if encryptedKeyEl == nil {
		encryptedKeyEl = keyInfoEl.FindElement(".//xenc:EncryptedKey")
	}
	if encryptedKeyEl == nil {
		return nil, fmt.Errorf("no EncryptedKey element found")
	}

	// Get key transport algorithm
	keyAlgorithm, err := getKeyTransportAlgorithm(encryptedKeyEl)
	if err != nil {
		return nil, fmt.Errorf("failed to get key transport algorithm: %w", err)
	}

	// Reject weak algorithms
	if keyAlgorithm == KeyTransportAlgorithmRSA15 {
		return nil, fmt.Errorf("RSA 1.5 key transport is not allowed (weak algorithm)")
	}

	// Reject legacy SHA-1 based OAEP (mgf1p uses SHA-1 which is cryptographically weak)
	if keyAlgorithm == KeyTransportAlgorithmRSAOAEP {
		return nil, fmt.Errorf("RSA-OAEP with SHA-1 (mgf1p) is not allowed (weak hash algorithm), use SHA-256 variant")
	}

	// Get encrypted key cipher value
	keyCipherValue, err := getCipherValue(encryptedKeyEl)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted key: %w", err)
	}

	// Decrypt the session key using RSA-OAEP
	sessionKey, err := d.rsaOAEPDecrypt(keyCipherValue, keyAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt session key: %w", err)
	}

	return sessionKey, nil
}

// rsaOAEPDecrypt decrypts data using RSA-OAEP
func (d *AssertionDecryptor) rsaOAEPDecrypt(ciphertext []byte, algorithm string) ([]byte, error) {
	if d.privateKey == nil {
		return nil, fmt.Errorf("no private key available")
	}

	// Select hash function based on algorithm - only SHA-256 variants are supported
	var hashFunc hash.Hash
	switch algorithm {
	case KeyTransportAlgorithmRSAOAEPSHA256:
		hashFunc = sha256.New()
	default:
		// Reject unknown or weak algorithms - only SHA-256 is supported
		return nil, fmt.Errorf("unsupported key transport algorithm: %s (only SHA-256 based algorithms are supported)", algorithm)
	}

	// Decrypt using RSA-OAEP
	plaintext, err := rsa.DecryptOAEP(hashFunc, rand.Reader, d.privateKey, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA-OAEP decryption failed: %w", err)
	}

	return plaintext, nil
}

// ============================================================================
// Decryption Functions
// ============================================================================

// decryptData decrypts the cipher value using the specified algorithm
func decryptData(ciphertext, key []byte, algorithm string) ([]byte, error) {
	switch algorithm {
	case EncryptionAlgorithmAES128CBC, EncryptionAlgorithmAES192CBC, EncryptionAlgorithmAES256CBC:
		return decryptAESCBC(ciphertext, key)
	case EncryptionAlgorithmAES128GCM, EncryptionAlgorithmAES256GCM:
		return decryptAESGCM(ciphertext, key)
	default:
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", algorithm)
	}
}

// decryptAESCBC decrypts data using AES-CBC
func decryptAESCBC(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// First block is the IV
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of block size")
	}

	// Decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding
	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad: %w", err)
	}

	return plaintext, nil
}

// decryptAESGCM decrypts data using AES-GCM
func decryptAESGCM(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("GCM decryption failed: %w", err)
	}

	return plaintext, nil
}

// pkcs7Unpad removes PKCS7 padding
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding > aes.BlockSize || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}

	// Verify padding bytes
	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}

	return data[:len(data)-padding], nil
}

// ============================================================================
// XML Parsing Helpers
// ============================================================================

// findEncryptedData finds the EncryptedData element
func findEncryptedData(root *etree.Element) *etree.Element {
	// Direct child
	if root.Tag == "EncryptedData" {
		return root
	}

	// Try various paths
	paths := []string{
		".//EncryptedData",
		".//xenc:EncryptedData",
		".//{" + xmlEncNS + "}EncryptedData",
	}

	for _, path := range paths {
		if el := root.FindElement(path); el != nil {
			return el
		}
	}

	return nil
}

// getEncryptionAlgorithm extracts the encryption algorithm from EncryptedData
func getEncryptionAlgorithm(encryptedData *etree.Element) (string, error) {
	encMethodEl := encryptedData.FindElement("./EncryptionMethod")
	if encMethodEl == nil {
		encMethodEl = encryptedData.FindElement(".//xenc:EncryptionMethod")
	}
	if encMethodEl == nil {
		return "", fmt.Errorf("no EncryptionMethod element found")
	}

	algorithm := encMethodEl.SelectAttrValue("Algorithm", "")
	if algorithm == "" {
		return "", fmt.Errorf("no Algorithm attribute found")
	}

	return algorithm, nil
}

// getKeyTransportAlgorithm extracts the key transport algorithm from EncryptedKey
func getKeyTransportAlgorithm(encryptedKey *etree.Element) (string, error) {
	encMethodEl := encryptedKey.FindElement("./EncryptionMethod")
	if encMethodEl == nil {
		encMethodEl = encryptedKey.FindElement(".//xenc:EncryptionMethod")
	}
	if encMethodEl == nil {
		return "", fmt.Errorf("no EncryptionMethod element found")
	}

	algorithm := encMethodEl.SelectAttrValue("Algorithm", "")
	if algorithm == "" {
		return "", fmt.Errorf("no Algorithm attribute found")
	}

	return algorithm, nil
}

// getCipherValue extracts the cipher value from an encrypted element
func getCipherValue(encryptedEl *etree.Element) ([]byte, error) {
	cipherDataEl := encryptedEl.FindElement("./CipherData")
	if cipherDataEl == nil {
		cipherDataEl = encryptedEl.FindElement(".//xenc:CipherData")
	}
	if cipherDataEl == nil {
		return nil, fmt.Errorf("no CipherData element found")
	}

	cipherValueEl := cipherDataEl.FindElement("./CipherValue")
	if cipherValueEl == nil {
		cipherValueEl = cipherDataEl.FindElement(".//xenc:CipherValue")
	}
	if cipherValueEl == nil {
		return nil, fmt.Errorf("no CipherValue element found")
	}

	// Decode base64
	cipherText := strings.TrimSpace(cipherValueEl.Text())
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cipher value: %w", err)
	}

	return data, nil
}

// parseRSAPrivateKey parses an RSA private key from PEM format
func parseRSAPrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	// Try to parse as PKCS1 first
	key, err := x509.ParsePKCS1PrivateKey(pemData)
	if err == nil {
		return key, nil
	}

	// Try PKCS8
	keyInterface, err := x509.ParsePKCS8PrivateKey(pemData)
	if err == nil {
		rsaKey, ok := keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not RSA")
		}
		return rsaKey, nil
	}

	// Try to decode PEM first
	block, _ := decodePEM(pemData)
	if block != nil {
		return parseRSAPrivateKey(block)
	}

	return nil, fmt.Errorf("failed to parse private key")
}

// decodePEM decodes a PEM block
func decodePEM(data []byte) ([]byte, []byte) {
	const pemHeader = "-----BEGIN"
	const pemFooter = "-----END"

	s := string(data)
	headerIdx := strings.Index(s, pemHeader)
	if headerIdx < 0 {
		return nil, data
	}

	// Find the end of the header line
	headerEnd := strings.Index(s[headerIdx:], "\n")
	if headerEnd < 0 {
		return nil, data
	}
	headerEnd += headerIdx

	// Find the footer
	footerIdx := strings.Index(s[headerEnd:], pemFooter)
	if footerIdx < 0 {
		return nil, data
	}
	footerIdx += headerEnd

	// Extract base64 content
	b64Content := strings.TrimSpace(s[headerEnd:footerIdx])
	b64Content = strings.ReplaceAll(b64Content, "\n", "")
	b64Content = strings.ReplaceAll(b64Content, "\r", "")

	decoded, err := base64.StdEncoding.DecodeString(b64Content)
	if err != nil {
		return nil, data
	}

	// Find end of footer line
	footerEnd := strings.Index(s[footerIdx:], "\n")
	if footerEnd < 0 {
		return decoded, nil
	}

	return decoded, []byte(s[footerIdx+footerEnd+1:])
}

// ============================================================================
// Updated decryptXMLEncryption function
// ============================================================================

// DecryptXMLEncryptionWithKey decrypts XML encryption data with a private key
func DecryptXMLEncryptionWithKey(encryptedData []byte, privateKeyPEM []byte) ([]byte, error) {
	decryptor, err := NewAssertionDecryptor(privateKeyPEM)
	if err != nil {
		return nil, err
	}
	return decryptor.DecryptAssertion(encryptedData)
}
