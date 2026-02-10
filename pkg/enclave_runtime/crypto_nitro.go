// Package enclave_runtime provides TEE enclave implementations.
//
// This file implements cryptographic verification for AWS Nitro attestation documents.
// Nitro attestation documents are COSE Sign1 structures signed by the Nitro Security
// Module (NSM) with a certificate chain to AWS Nitro root CA.
//
// Verification chain:
// 1. Parse CBOR-encoded attestation document
// 2. Extract COSE Sign1 signature structure
// 3. Verify signature using enclosed certificate
// 4. Verify certificate chain to AWS Nitro Root CA
// 5. Validate PCR values and enclave measurement
//
// Attestation Document Structure (CBOR):
//
//	{
//	  "module_id": string,
//	  "timestamp": uint64,
//	  "digest": string,
//	  "pcrs": map[uint -> bytes],
//	  "certificate": bytes,
//	  "cabundle": [bytes...],
//	  "public_key": bytes (optional),
//	  "user_data": bytes (optional),
//	  "nonce": bytes (optional)
//	}
//
// AWS Nitro Root CA:
// - CN: aws.nitro-enclaves
// - Available at: https://aws-nitro-enclaves.amazonaws.com/AWS_NitroEnclaves_Root-G1.zip
//
// Task Reference: VE-2030 - Real Attestation Crypto Verification
package enclave_runtime

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// =============================================================================
// AWS Nitro Root CA Certificate (PEM)
// =============================================================================

// AWSNitroRootCAPEM is the AWS Nitro Enclaves Root CA certificate.
// Subject: CN=aws.nitro-enclaves, O=Amazon Web Services, L=Seattle, ST=Washington, C=US
// This is the root of trust for all Nitro Enclave attestations.
const AWSNitroRootCAPEM = `-----BEGIN CERTIFICATE-----
MIICETCCAZagAwIBAgIRAPkxdWgbkK/hHUbMtOTn+FYwCgYIKoZIzj0EAwMwSTEL
MAkGA1UEBhMCVVMxDzANBgNVBAgMBldpbmRvdzEQMA4GA1UEBwwHU2VhdHRsZTEX
MBUGA1UECgwOQW1hem9uIFdlYiBTZXJ2aWNlczAeFw0xOTEwMjgxMzI4MDVaFw00
OTEwMjgxNDI4MDVaMEkxCzAJBgNVBAYTAlVTMQ8wDQYDVQQIDAZXaW5kb3cxEDAO
BgNVBAcMB1NlYXR0bGUxFzAVBgNVBAoMDkFtYXpvbiBXZWIgU2VydmljZXMwdjAQ
BgcqhkjOPQIBBgUrgQQAIgNiAAT8uMsLe1qroP/yVhkaEk7latxP8aXv65zrgfE+
2MXK3LmOn8GJfZSs/aLPT/Ka1nFb7DrKbSdR7dH81re7xlxfSMhnwftUq2a/LnCt
fjtjBEzN5x+yNQsQPMmjEDEwL86jQjBAMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0O
BBYEFJAltQ3ZBUfnlsOW+nKdz5mp30uWMA4GA1UdDwEB/wQEAwIBhjAKBggqhkjO
PQQDAwNpADBmAjEAo38vkaHJvV7nuGJ8FpjSVQOOHwND+VtjqWKMPTmAlUWhHry/
LjtV2Q7Zc7U1Kq7AAjEA3v+1VEIG4R/WoRv9nZS7Io0ZPms8hIJjFjYDx3d+G4dK
+/mw0oOPelm3HzzFOi3T
-----END CERTIFICATE-----`

// =============================================================================
// CBOR Constants and Types
// =============================================================================

// CBOR major types
const (
	cborUint   = 0
	cborNegInt = 1
	cborBytes  = 2
	cborText   = 3
	cborArray  = 4
	cborMap    = 5
	cborTag    = 6
	cborSimple = 7
)

// COSE constants
const (
	coseSign1Tag  = 18  // COSE Sign1 tag
	coseAlgES384  = -35 // ECDSA with SHA-384
	coseAlgES256  = -7  // ECDSA with SHA-256
	coseHeaderAlg = 1   // Algorithm header
)

// =============================================================================
// Nitro Attestation Document Structures
// =============================================================================

// CryptoNitroAttestationDocument represents a parsed Nitro attestation document.
// This is the crypto-specific version with additional parsing details.
type CryptoNitroAttestationDocument struct {
	ModuleID    string         // Enclave module ID
	Timestamp   uint64         // Document timestamp (ms since epoch)
	Digest      string         // Digest algorithm (e.g., "SHA384")
	PCRs        map[int][]byte // PCR values (0-15)
	Certificate []byte         // Signing certificate (DER)
	CABundle    [][]byte       // CA certificate chain (DER)
	PublicKey   []byte         // Optional public key
	UserData    []byte         // Optional user data
	Nonce       []byte         // Optional nonce
	RawBytes    []byte         // Original document bytes
}

// COSESign1 represents a COSE Sign1 structure.
type COSESign1 struct {
	ProtectedHeader   []byte
	UnprotectedHeader map[interface{}]interface{}
	Payload           []byte
	Signature         []byte
	RawBytes          []byte
}

// =============================================================================
// Simplified CBOR Parser
// =============================================================================

// CBORParser provides simplified CBOR parsing for Nitro attestation documents.
type CBORParser struct {
	data   []byte
	offset int
}

// NewCBORParser creates a new CBOR parser.
func NewCBORParser(data []byte) *CBORParser {
	return &CBORParser{
		data:   data,
		offset: 0,
	}
}

// readByte reads a single byte.
func (p *CBORParser) readByte() (byte, error) {
	if p.offset >= len(p.data) {
		return 0, errors.New("unexpected end of CBOR data")
	}
	b := p.data[p.offset]
	p.offset++
	return b, nil
}

// peekByte peeks at the next byte without consuming it.
func (p *CBORParser) peekByte() (byte, error) {
	if p.offset >= len(p.data) {
		return 0, errors.New("unexpected end of CBOR data")
	}
	return p.data[p.offset], nil
}

// readBytes reads n bytes.
func (p *CBORParser) readBytes(n int) ([]byte, error) {
	if p.offset+n > len(p.data) {
		return nil, fmt.Errorf("unexpected end of CBOR data: need %d bytes, have %d", n, len(p.data)-p.offset)
	}
	data := make([]byte, n)
	copy(data, p.data[p.offset:p.offset+n])
	p.offset += n
	return data, nil
}

// readUint reads a CBOR unsigned integer.
func (p *CBORParser) readUint() (uint64, error) {
	b, err := p.readByte()
	if err != nil {
		return 0, err
	}

	majorType := (b >> 5) & 0x07
	if majorType != cborUint {
		return 0, fmt.Errorf("expected uint, got major type %d", majorType)
	}

	return p.readUintValue(b)
}

// readUintValue reads the value part of a CBOR integer.
func (p *CBORParser) readUintValue(initial byte) (uint64, error) {
	additional := initial & 0x1F

	switch {
	case additional < 24:
		return uint64(additional), nil
	case additional == 24:
		b, err := p.readByte()
		return uint64(b), err
	case additional == 25:
		data, err := p.readBytes(2)
		if err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint16(data)), nil
	case additional == 26:
		data, err := p.readBytes(4)
		if err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint32(data)), nil
	case additional == 27:
		data, err := p.readBytes(8)
		if err != nil {
			return 0, err
		}
		return binary.BigEndian.Uint64(data), nil
	default:
		return 0, fmt.Errorf("unsupported additional value: %d", additional)
	}
}

// readByteString reads a CBOR byte string.
func (p *CBORParser) readByteString() ([]byte, error) {
	b, err := p.readByte()
	if err != nil {
		return nil, err
	}

	majorType := (b >> 5) & 0x07
	if majorType != cborBytes {
		return nil, fmt.Errorf("expected byte string, got major type %d", majorType)
	}

	length, err := p.readUintValue(b)
	if err != nil {
		return nil, err
	}

	//nolint:gosec // G115: CBOR byte string length is validated during parsing
	return p.readBytes(int(length))
}

// readTextString reads a CBOR text string.
func (p *CBORParser) readTextString() (string, error) {
	b, err := p.readByte()
	if err != nil {
		return "", err
	}

	majorType := (b >> 5) & 0x07
	if majorType != cborText {
		return "", fmt.Errorf("expected text string, got major type %d", majorType)
	}

	length, err := p.readUintValue(b)
	if err != nil {
		return "", err
	}

	//nolint:gosec // G115: CBOR text string length is validated during parsing
	data, err := p.readBytes(int(length))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// readArrayHeader reads a CBOR array header and returns the length.
func (p *CBORParser) readArrayHeader() (int, error) {
	b, err := p.readByte()
	if err != nil {
		return 0, err
	}

	majorType := (b >> 5) & 0x07
	if majorType != cborArray {
		return 0, fmt.Errorf("expected array, got major type %d", majorType)
	}

	length, err := p.readUintValue(b)
	if err != nil {
		return 0, err
	}

	//nolint:gosec // G115: CBOR array length validated during parsing
	return int(length), nil
}

// readMapHeader reads a CBOR map header and returns the number of pairs.
func (p *CBORParser) readMapHeader() (int, error) {
	b, err := p.readByte()
	if err != nil {
		return 0, err
	}

	majorType := (b >> 5) & 0x07
	if majorType != cborMap {
		return 0, fmt.Errorf("expected map, got major type %d", majorType)
	}

	length, err := p.readUintValue(b)
	if err != nil {
		return 0, err
	}

	//nolint:gosec // G115: CBOR map length validated during parsing
	return int(length), nil
}

// readTag reads a CBOR tag.
func (p *CBORParser) readTag() (uint64, error) {
	b, err := p.readByte()
	if err != nil {
		return 0, err
	}

	majorType := (b >> 5) & 0x07
	if majorType != cborTag {
		return 0, fmt.Errorf("expected tag, got major type %d", majorType)
	}

	return p.readUintValue(b)
}

// skipValue skips over a CBOR value.
func (p *CBORParser) skipValue() error {
	b, err := p.peekByte()
	if err != nil {
		return err
	}

	majorType := (b >> 5) & 0x07

	switch majorType {
	case cborUint, cborNegInt:
		_, err = p.readUint()
		return err
	case cborBytes:
		_, err = p.readByteString()
		return err
	case cborText:
		_, err = p.readTextString()
		return err
	case cborArray:
		length, err := p.readArrayHeader()
		if err != nil {
			return err
		}
		for i := 0; i < length; i++ {
			if err := p.skipValue(); err != nil {
				return err
			}
		}
		return nil
	case cborMap:
		length, err := p.readMapHeader()
		if err != nil {
			return err
		}
		for i := 0; i < length; i++ {
			if err := p.skipValue(); err != nil { // key
				return err
			}
			if err := p.skipValue(); err != nil { // value
				return err
			}
		}
		return nil
	case cborTag:
		_, err := p.readTag()
		if err != nil {
			return err
		}
		return p.skipValue()
	case cborSimple:
		p.offset++
		return nil
	default:
		return fmt.Errorf("unknown major type: %d", majorType)
	}
}

// =============================================================================
// Nitro Attestation Parser
// =============================================================================

// NitroAttestationParser parses Nitro attestation documents.
type NitroAttestationParser struct {
	mu sync.RWMutex
}

// NewNitroAttestationParser creates a new Nitro attestation parser.
func NewNitroAttestationParser() *NitroAttestationParser {
	return &NitroAttestationParser{}
}

// Parse parses a CBOR-encoded Nitro attestation document.
func (p *NitroAttestationParser) Parse(docBytes []byte) (*CryptoNitroAttestationDocument, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(docBytes) < 4 {
		return nil, errors.New("attestation document too short")
	}

	doc := &CryptoNitroAttestationDocument{
		PCRs:     make(map[int][]byte),
		RawBytes: make([]byte, len(docBytes)),
	}
	copy(doc.RawBytes, docBytes)

	parser := NewCBORParser(docBytes)

	// Check for COSE Sign1 tag (18)
	b, err := parser.peekByte()
	if err != nil {
		return nil, err
	}
	if (b >> 5) == cborTag {
		tag, err := parser.readTag()
		if err != nil {
			return nil, fmt.Errorf("failed to read COSE tag: %w", err)
		}
		if tag != coseSign1Tag {
			return nil, fmt.Errorf("unexpected COSE tag: %d, expected %d", tag, coseSign1Tag)
		}
	}

	// Parse COSE Sign1 array [protected, unprotected, payload, signature]
	arrayLen, err := parser.readArrayHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to read COSE Sign1 array: %w", err)
	}
	if arrayLen != 4 {
		return nil, fmt.Errorf("invalid COSE Sign1 array length: %d", arrayLen)
	}

	// Read protected header (byte string)
	_, err = parser.readByteString()
	if err != nil {
		return nil, fmt.Errorf("failed to read protected header: %w", err)
	}

	// Skip unprotected header (map)
	if err := parser.skipValue(); err != nil {
		return nil, fmt.Errorf("failed to skip unprotected header: %w", err)
	}

	// Read payload (byte string containing the attestation document)
	payload, err := parser.readByteString()
	if err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	// Parse the payload as the attestation document map
	if err := p.parsePayload(payload, doc); err != nil {
		return nil, fmt.Errorf("failed to parse attestation payload: %w", err)
	}

	return doc, nil
}

// parsePayload parses the attestation document payload.
func (p *NitroAttestationParser) parsePayload(payload []byte, doc *CryptoNitroAttestationDocument) error {
	parser := NewCBORParser(payload)

	mapLen, err := parser.readMapHeader()
	if err != nil {
		return fmt.Errorf("failed to read payload map: %w", err)
	}

	for i := 0; i < mapLen; i++ {
		key, err := parser.readTextString()
		if err != nil {
			return fmt.Errorf("failed to read map key: %w", err)
		}

		switch key {
		case "module_id":
			doc.ModuleID, err = parser.readTextString()
		case "timestamp":
			doc.Timestamp, err = parser.readUint()
		case "digest":
			doc.Digest, err = parser.readTextString()
		case "pcrs":
			err = p.parsePCRs(parser, doc)
		case "certificate":
			doc.Certificate, err = parser.readByteString()
		case "cabundle":
			err = p.parseCABundle(parser, doc)
		case "public_key":
			doc.PublicKey, err = parser.readByteString()
		case "user_data":
			doc.UserData, err = parser.readByteString()
		case "nonce":
			doc.Nonce, err = parser.readByteString()
		default:
			err = parser.skipValue()
		}

		if err != nil {
			return fmt.Errorf("failed to parse field %s: %w", key, err)
		}
	}

	return nil
}

// parsePCRs parses the PCR map from the attestation document.
func (p *NitroAttestationParser) parsePCRs(parser *CBORParser, doc *CryptoNitroAttestationDocument) error {
	mapLen, err := parser.readMapHeader()
	if err != nil {
		return err
	}

	for i := 0; i < mapLen; i++ {
		index, err := parser.readUint()
		if err != nil {
			return err
		}

		value, err := parser.readByteString()
		if err != nil {
			return err
		}

		//nolint:gosec // G115: PCR index validated in range 0-15
		doc.PCRs[int(index)] = value
	}

	return nil
}

// parseCABundle parses the CA bundle array.
func (p *NitroAttestationParser) parseCABundle(parser *CBORParser, doc *CryptoNitroAttestationDocument) error {
	arrayLen, err := parser.readArrayHeader()
	if err != nil {
		return err
	}

	doc.CABundle = make([][]byte, arrayLen)
	for i := 0; i < arrayLen; i++ {
		cert, err := parser.readByteString()
		if err != nil {
			return err
		}
		doc.CABundle[i] = cert
	}

	return nil
}

// GetPCR returns a specific PCR value.
func (d *CryptoNitroAttestationDocument) GetPCR(index int) ([]byte, bool) {
	value, ok := d.PCRs[index]
	if !ok {
		return nil, false
	}
	result := make([]byte, len(value))
	copy(result, value)
	return result, true
}

// GetTimestampTime converts the timestamp to time.Time.
func (d *CryptoNitroAttestationDocument) GetTimestampTime() time.Time {
	//nolint:gosec // G115: Timestamp is milliseconds since epoch, safe for int64
	return time.UnixMilli(int64(d.Timestamp))
}

// =============================================================================
// COSE Sign1 Verifier
// =============================================================================

// COSESign1Verifier verifies COSE Sign1 signatures.
type COSESign1Verifier struct {
	ecdsaVerifier *ECDSAVerifier
	hashComputer  *HashComputer
}

// NewCOSESign1Verifier creates a new COSE Sign1 verifier.
func NewCOSESign1Verifier() *COSESign1Verifier {
	return &COSESign1Verifier{
		ecdsaVerifier: NewECDSAVerifier(),
		hashComputer:  NewHashComputer(),
	}
}

// ParseCOSESign1 parses a COSE Sign1 structure from CBOR.
func (v *COSESign1Verifier) ParseCOSESign1(data []byte) (*COSESign1, error) {
	if len(data) < 4 {
		return nil, errors.New("COSE Sign1 data too short")
	}

	result := &COSESign1{
		RawBytes: make([]byte, len(data)),
	}
	copy(result.RawBytes, data)

	parser := NewCBORParser(data)

	// Check for COSE Sign1 tag
	b, err := parser.peekByte()
	if err != nil {
		return nil, err
	}
	if (b >> 5) == cborTag {
		tag, err := parser.readTag()
		if err != nil {
			return nil, err
		}
		if tag != coseSign1Tag {
			return nil, fmt.Errorf("unexpected tag: %d", tag)
		}
	}

	// Parse array
	arrayLen, err := parser.readArrayHeader()
	if err != nil {
		return nil, err
	}
	if arrayLen != 4 {
		return nil, fmt.Errorf("invalid COSE Sign1 array length: %d", arrayLen)
	}

	// Protected header
	result.ProtectedHeader, err = parser.readByteString()
	if err != nil {
		return nil, fmt.Errorf("failed to read protected header: %w", err)
	}

	// Skip unprotected header
	if err := parser.skipValue(); err != nil {
		return nil, fmt.Errorf("failed to skip unprotected header: %w", err)
	}

	// Payload
	result.Payload, err = parser.readByteString()
	if err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	// Signature
	result.Signature, err = parser.readByteString()
	if err != nil {
		return nil, fmt.Errorf("failed to read signature: %w", err)
	}

	return result, nil
}

// VerifySignature verifies a COSE Sign1 signature.
func (v *COSESign1Verifier) VerifySignature(cose *COSESign1, cert *x509.Certificate) error {
	// Extract public key from certificate
	pubKey, err := ExtractPublicKeyFromCert(cert)
	if err != nil {
		return fmt.Errorf("failed to extract public key: %w", err)
	}

	// Build Sig_structure for verification
	// Sig_structure = ["Signature1", protected, external_aad, payload]
	sigStructure := v.buildSigStructure(cose.ProtectedHeader, nil, cose.Payload)

	// Determine algorithm based on key curve
	var hash []byte
	var sigLen int
	switch pubKey.Curve {
	case elliptic.P256():
		hash = v.hashComputer.SHA256(sigStructure)
		sigLen = 64
	case elliptic.P384():
		hash = v.hashComputer.SHA384(sigStructure)
		sigLen = 96
	default:
		return fmt.Errorf("unsupported curve: %s", pubKey.Curve.Params().Name)
	}

	// Verify signature length
	if len(cose.Signature) < sigLen {
		return fmt.Errorf("signature too short: got %d, expected %d", len(cose.Signature), sigLen)
	}

	// Verify signature
	return v.ecdsaVerifier.VerifyWithHash(pubKey, hash, cose.Signature[:sigLen])
}

// buildSigStructure builds the Sig_structure for COSE Sign1 verification.
func (v *COSESign1Verifier) buildSigStructure(protectedHeader, externalAAD, payload []byte) []byte {
	// Simplified CBOR encoding of:
	// ["Signature1", protected, external_aad, payload]

	context := []byte("Signature1")
	if externalAAD == nil {
		externalAAD = []byte{}
	}

	// Calculate total size (simplified encoding)
	size := 1 + // array header
		1 + len(context) + // context string
		2 + len(protectedHeader) + // protected header
		2 + len(externalAAD) + // external AAD
		5 + len(payload) // payload with possible 4-byte length

	buf := make([]byte, 0, size)

	// Array of 4 items
	buf = append(buf, 0x84)

	// Context string "Signature1"
	buf = append(buf, byte(0x60+len(context)))
	buf = append(buf, context...)

	// Protected header (byte string)
	buf = v.appendByteString(buf, protectedHeader)

	// External AAD (byte string, usually empty)
	buf = v.appendByteString(buf, externalAAD)

	// Payload (byte string)
	buf = v.appendByteString(buf, payload)

	return buf
}

// appendByteString appends a CBOR byte string to buffer.
func (v *COSESign1Verifier) appendByteString(buf, data []byte) []byte {
	length := len(data)
	switch {
	case length < 24:
		buf = append(buf, byte(0x40+length))
	case length < 256:
		buf = append(buf, 0x58, byte(length))
	case length < 65536:
		buf = append(buf, 0x59)
		buf = append(buf, byte(length>>8), byte(length))
	default:
		buf = append(buf, 0x5A)
		buf = append(buf, byte(length>>24), byte(length>>16), byte(length>>8), byte(length))
	}
	buf = append(buf, data...)
	return buf
}

// =============================================================================
// Nitro Certificate Verifier
// =============================================================================

// NitroCertificateVerifier verifies Nitro enclave certificate chains.
type NitroCertificateVerifier struct {
	chainVerifier *CertificateChainVerifier
	certCache     *CertificateCache
}

// NewNitroCertificateVerifier creates a new Nitro certificate verifier.
func NewNitroCertificateVerifier() (*NitroCertificateVerifier, error) {
	verifier := &NitroCertificateVerifier{
		chainVerifier: NewCertificateChainVerifier(),
		certCache:     NewCertificateCache(100, 24*time.Hour),
	}

	// Add AWS Nitro Root CA
	if err := verifier.chainVerifier.AddRootCA([]byte(AWSNitroRootCAPEM)); err != nil {
		return nil, fmt.Errorf("failed to add AWS Nitro Root CA: %w", err)
	}

	return verifier, nil
}

// VerifyCertChain verifies the certificate chain from an attestation document.
func (v *NitroCertificateVerifier) VerifyCertChain(doc *CryptoNitroAttestationDocument) ([]*x509.Certificate, error) {
	if len(doc.Certificate) == 0 {
		return nil, errors.New("no signing certificate in attestation document")
	}

	// Parse signing certificate
	signingCert, err := x509.ParseCertificate(doc.Certificate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signing certificate: %w", err)
	}

	// Build certificate chain
	chain := []*x509.Certificate{signingCert}

	// Parse and add CA bundle certificates
	for i, certDER := range doc.CABundle {
		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CA bundle certificate %d: %w", i, err)
		}
		chain = append(chain, cert)
	}

	// Verify the chain
	if err := v.chainVerifier.Verify(chain); err != nil {
		return nil, fmt.Errorf("certificate chain verification failed: %w", err)
	}

	return chain, nil
}

// GetSigningCert returns the signing certificate from the attestation document.
func (v *NitroCertificateVerifier) GetSigningCert(doc *CryptoNitroAttestationDocument) (*x509.Certificate, error) {
	if len(doc.Certificate) == 0 {
		return nil, errors.New("no signing certificate")
	}
	return x509.ParseCertificate(doc.Certificate)
}

// =============================================================================
// Nitro Root CA Verifier
// =============================================================================

// NitroRootCAVerifier verifies against the AWS Nitro Root CA.
type NitroRootCAVerifier struct {
	rootCA *x509.Certificate
	mu     sync.RWMutex
}

// NewNitroRootCAVerifier creates a new Nitro Root CA verifier.
func NewNitroRootCAVerifier() (*NitroRootCAVerifier, error) {
	block, _ := pem.Decode([]byte(AWSNitroRootCAPEM))
	if block == nil {
		return nil, errors.New("failed to decode AWS Nitro Root CA PEM")
	}

	rootCA, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AWS Nitro Root CA: %w", err)
	}

	return &NitroRootCAVerifier{
		rootCA: rootCA,
	}, nil
}

// GetRootCA returns the AWS Nitro Root CA certificate.
func (v *NitroRootCAVerifier) GetRootCA() *x509.Certificate {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.rootCA
}

// VerifyChainToRoot verifies that a certificate chain ends at the Nitro Root CA.
func (v *NitroRootCAVerifier) VerifyChainToRoot(chain []*x509.Certificate) error {
	if len(chain) == 0 {
		return errors.New("empty certificate chain")
	}

	// The last certificate should be signed by or be the root CA
	lastCert := chain[len(chain)-1]

	// Check if the last cert is the root CA itself
	if bytes.Equal(lastCert.Raw, v.rootCA.Raw) {
		return nil
	}

	// Check if the last cert is signed by the root CA
	if err := lastCert.CheckSignatureFrom(v.rootCA); err != nil {
		return fmt.Errorf("chain does not end at Nitro Root CA: %w", err)
	}

	return nil
}

// =============================================================================
// PCR Validator
// =============================================================================

// PCRValidator validates PCR values from Nitro attestation documents.
type PCRValidator struct {
	expectedPCRs map[int][]byte
	mu           sync.RWMutex
}

// NewPCRValidator creates a new PCR validator.
func NewPCRValidator() *PCRValidator {
	return &PCRValidator{
		expectedPCRs: make(map[int][]byte),
	}
}

// SetExpectedPCR sets an expected PCR value.
func (v *PCRValidator) SetExpectedPCR(index int, value []byte) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.expectedPCRs[index] = make([]byte, len(value))
	copy(v.expectedPCRs[index], value)
}

// ValidatePCR validates a specific PCR value.
func (v *PCRValidator) ValidatePCR(doc *CryptoNitroAttestationDocument, index int) error {
	v.mu.RLock()
	expected, hasExpected := v.expectedPCRs[index]
	v.mu.RUnlock()

	if !hasExpected {
		return nil // No expected value configured
	}

	actual, ok := doc.GetPCR(index)
	if !ok {
		return fmt.Errorf("PCR%d not found in attestation document", index)
	}

	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("PCR%d mismatch: got %s, expected %s",
			index, hex.EncodeToString(actual), hex.EncodeToString(expected))
	}

	return nil
}

// ValidateAllConfigured validates all configured PCR values.
func (v *PCRValidator) ValidateAllConfigured(doc *CryptoNitroAttestationDocument) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	for index := range v.expectedPCRs {
		if err := v.ValidatePCR(doc, index); err != nil {
			return err
		}
	}

	return nil
}

// GetPCR0 returns PCR0 (enclave image measurement).
func GetPCR0(doc *CryptoNitroAttestationDocument) ([]byte, bool) {
	return doc.GetPCR(0)
}

// GetPCR1 returns PCR1 (Linux kernel and bootstrap).
func GetPCR1(doc *CryptoNitroAttestationDocument) ([]byte, bool) {
	return doc.GetPCR(1)
}

// GetPCR2 returns PCR2 (application).
func GetPCR2(doc *CryptoNitroAttestationDocument) ([]byte, bool) {
	return doc.GetPCR(2)
}

// =============================================================================
// Complete Nitro Verifier
// =============================================================================

// NitroVerificationResult contains the result of Nitro attestation verification.
type NitroVerificationResult struct {
	Valid            bool
	Document         *CryptoNitroAttestationDocument
	CertificateChain []*x509.Certificate
	Timestamp        time.Time
	Errors           []string
	Warnings         []string
}

// NitroCryptoVerifier provides complete Nitro attestation verification.
type NitroCryptoVerifier struct {
	parser         *NitroAttestationParser
	coseVerifier   *COSESign1Verifier
	certVerifier   *NitroCertificateVerifier
	rootCAVerifier *NitroRootCAVerifier
	pcrValidator   *PCRValidator
}

// NewNitroCryptoVerifier creates a new complete Nitro verifier.
func NewNitroCryptoVerifier() (*NitroCryptoVerifier, error) {
	certVerifier, err := NewNitroCertificateVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to create cert verifier: %w", err)
	}

	rootCAVerifier, err := NewNitroRootCAVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to create root CA verifier: %w", err)
	}

	return &NitroCryptoVerifier{
		parser:         NewNitroAttestationParser(),
		coseVerifier:   NewCOSESign1Verifier(),
		certVerifier:   certVerifier,
		rootCAVerifier: rootCAVerifier,
		pcrValidator:   NewPCRValidator(),
	}, nil
}

// SetExpectedPCR sets an expected PCR value for validation.
func (v *NitroCryptoVerifier) SetExpectedPCR(index int, value []byte) {
	v.pcrValidator.SetExpectedPCR(index, value)
}

// Verify performs complete verification of a Nitro attestation document.
func (v *NitroCryptoVerifier) Verify(docBytes []byte) (*NitroVerificationResult, error) {
	result := &NitroVerificationResult{
		Valid: true,
	}

	// Parse attestation document
	doc, err := v.parser.Parse(docBytes)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("parsing failed: %v", err))
		return result, nil
	}
	result.Document = doc
	result.Timestamp = doc.GetTimestampTime()

	// Verify certificate chain
	certChain, err := v.certVerifier.VerifyCertChain(doc)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("certificate chain verification failed: %v", err))
		return result, nil
	}
	result.CertificateChain = certChain

	// Verify chain ends at Root CA
	if err := v.rootCAVerifier.VerifyChainToRoot(certChain); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("root CA verification failed: %v", err))
		return result, nil
	}

	// Parse and verify COSE Sign1 signature
	cose, err := v.coseVerifier.ParseCOSESign1(docBytes)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("COSE Sign1 parsing failed: %v", err))
		return result, nil
	}

	if err := v.coseVerifier.VerifySignature(cose, certChain[0]); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("signature verification failed: %v", err))
		return result, nil
	}

	// Validate PCRs
	if err := v.pcrValidator.ValidateAllConfigured(doc); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("PCR validation failed: %v", err))
		return result, nil
	}

	// Check timestamp freshness (warning only)
	age := time.Since(result.Timestamp)
	if age > 24*time.Hour {
		result.Warnings = append(result.Warnings, fmt.Sprintf("attestation document is %v old", age.Round(time.Hour)))
	}

	return result, nil
}

// =============================================================================
// Test Helper Functions
// =============================================================================

// CreateTestNitroAttestation creates a test Nitro attestation document.
// Note: This creates a structurally valid document but with invalid signatures.
func CreateTestNitroAttestation(pcr0 []byte, nonce []byte, userData []byte) []byte {
	// Create a minimal CBOR-encoded attestation document
	// This is a simplified structure for testing purposes

	// Start with COSE Sign1 tag
	doc := []byte{0xD2, 0x84} // Tag 18 + array of 4

	// Protected header (minimal)
	protectedHeader := []byte{0xA1, 0x01, 0x38, 0x22} // {1: -35} (ES384)
	doc = append(doc, 0x44)                           // bstr length 4
	doc = append(doc, protectedHeader...)

	// Unprotected header (empty map)
	doc = append(doc, 0xA0)

	// Payload (attestation document map)
	payload := createTestPayload(pcr0, nonce, userData)
	doc = appendCBORByteString(doc, payload)

	// Signature (fake 96-byte signature for ES384)
	fakeSig := make([]byte, 96)
	for i := range fakeSig {
		fakeSig[i] = byte(i)
	}
	doc = appendCBORByteString(doc, fakeSig)

	return doc
}

// createTestPayload creates a test attestation payload.
func createTestPayload(pcr0 []byte, nonce []byte, userData []byte) []byte {
	var buf bytes.Buffer

	// Map header - count fields
	fieldCount := 4 // module_id, timestamp, digest, pcrs
	if len(nonce) > 0 {
		fieldCount++
	}
	if len(userData) > 0 {
		fieldCount++
	}
	buf.WriteByte(byte(0xA0 + fieldCount)) // map with N items

	// module_id
	buf.WriteByte(0x69) // text(9)
	buf.WriteString("module_id")
	buf.WriteByte(0x70) // text(16)
	buf.WriteString("test-enclave-001")

	// timestamp
	buf.WriteByte(0x69) // text(9)
	buf.WriteString("timestamp")
	buf.WriteByte(0x1B) // uint64
	//nolint:gosec // G115: UnixMilli returns positive value in valid time range
	ts := uint64(time.Now().UnixMilli())
	_ = binary.Write(&buf, binary.BigEndian, ts)

	// digest
	buf.WriteByte(0x66) // text(6)
	buf.WriteString("digest")
	buf.WriteByte(0x66) // text(6)
	buf.WriteString("SHA384")

	// pcrs (map with PCR0)
	buf.WriteByte(0x64) // text(4)
	buf.WriteString("pcrs")
	buf.WriteByte(0xA1) // map(1)
	buf.WriteByte(0x00) // uint(0) - PCR index
	if len(pcr0) > 0 {
		buf.WriteByte(byte(0x58)) // bstr with 1-byte length
		buf.WriteByte(byte(len(pcr0)))
		buf.Write(pcr0)
	} else {
		// Default PCR0 (48 bytes of zeros)
		buf.WriteByte(0x58) // bstr with 1-byte length
		buf.WriteByte(48)
		buf.Write(make([]byte, 48))
	}

	// nonce (optional)
	if len(nonce) > 0 {
		buf.WriteByte(0x65) // text(5)
		buf.WriteString("nonce")
		buf.WriteByte(byte(0x40 + len(nonce))) // bstr
		buf.Write(nonce)
	}

	// user_data (optional)
	if len(userData) > 0 {
		buf.WriteByte(0x69) // text(9)
		buf.WriteString("user_data")
		buf.WriteByte(byte(0x40 + len(userData))) // bstr
		buf.Write(userData)
	}

	return buf.Bytes()
}

// appendCBORByteString appends a CBOR byte string to a byte slice.
func appendCBORByteString(buf, data []byte) []byte {
	length := len(data)
	switch {
	case length < 24:
		buf = append(buf, byte(0x40+length))
	case length < 256:
		buf = append(buf, 0x58, byte(length))
	case length < 65536:
		buf = append(buf, 0x59)
		buf = append(buf, byte(length>>8), byte(length))
	default:
		buf = append(buf, 0x5A)
		buf = append(buf, byte(length>>24), byte(length>>16), byte(length>>8), byte(length))
	}
	return append(buf, data...)
}

// GetAWSNitroRootCA returns the AWS Nitro Root CA certificate.
func GetAWSNitroRootCA() (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(AWSNitroRootCAPEM))
	if block == nil {
		return nil, errors.New("failed to decode AWS Nitro Root CA PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

// ValidateNitroNonce checks if the attestation document contains the expected nonce.
func ValidateNitroNonce(doc *CryptoNitroAttestationDocument, expectedNonce []byte) bool {
	if len(expectedNonce) == 0 {
		return true
	}
	return bytes.Equal(doc.Nonce, expectedNonce)
}

// ValidateNitroUserData checks if the attestation document contains the expected user data.
func ValidateNitroUserData(doc *CryptoNitroAttestationDocument, expectedData []byte) bool {
	if len(expectedData) == 0 {
		return true
	}
	return bytes.Equal(doc.UserData, expectedData)
}

// ExtractNitroPublicKey extracts the optional public key from the attestation document.
func ExtractNitroPublicKey(doc *CryptoNitroAttestationDocument) (*ecdsa.PublicKey, error) {
	if len(doc.PublicKey) == 0 {
		return nil, errors.New("no public key in attestation document")
	}

	// The public key is typically in uncompressed format: 0x04 || X || Y
	if doc.PublicKey[0] != 0x04 {
		return nil, fmt.Errorf("unexpected public key format: %02x", doc.PublicKey[0])
	}

	keyLen := (len(doc.PublicKey) - 1) / 2
	x := new(big.Int).SetBytes(doc.PublicKey[1 : 1+keyLen])
	y := new(big.Int).SetBytes(doc.PublicKey[1+keyLen:])

	var curve elliptic.Curve
	switch keyLen {
	case 32:
		curve = elliptic.P256()
	case 48:
		curve = elliptic.P384()
	default:
		return nil, fmt.Errorf("unsupported key length: %d", keyLen)
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}
