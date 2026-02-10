// Package nitro provides AWS Nitro Enclave integration for VirtEngine TEE.
//
// This file implements attestation document handling for AWS Nitro Enclaves.
// The attestation document follows the COSE_Sign1 format as specified by AWS.
//
// Document structure:
// - Protected header: Algorithm identifier (ES384)
// - Unprotected header: Empty
// - Payload: CBOR-encoded attestation data
// - Signature: ECDSA signature over protected header and payload
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package nitro

import (
	"bytes"
	"crypto/sha512"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// Document version
	AttestationDocVersion = 1

	// Maximum sizes
	MaxModuleIDSize  = 64
	MaxUserDataSize  = 1024
	MaxNonceSize     = 64
	MaxPublicKeySize = 1024
	MaxCertSize      = 4096
	MaxCABundleSize  = 16384

	// Digest algorithm
	DigestAlgorithmSHA384 = "SHA384"

	// COSE algorithm identifier for ES384
	COSEAlgorithmES384 = -35

	// CBOR major types
	CBORMajorTypeUnsigned   = 0
	CBORMajorTypeNegative   = 1
	CBORMajorTypeByteString = 2
	CBORMajorTypeTextString = 3
	CBORMajorTypeArray      = 4
	CBORMajorTypeMap        = 5
	CBORMajorTypeTag        = 6
	CBORMajorTypeFloat      = 7

	// CBOR additional info
	CBORAdditionalInfo1Byte  = 24
	CBORAdditionalInfo2Bytes = 25
	CBORAdditionalInfo4Bytes = 26
	CBORAdditionalInfo8Bytes = 27

	// COSE_Sign1 tag
	COSESign1Tag = 18
)

// =============================================================================
// Errors
// =============================================================================

var (
	// ErrInvalidDocument is returned when the attestation document is malformed
	ErrInvalidDocument = errors.New("invalid attestation document")

	// ErrInvalidSignature is returned when document signature is invalid
	ErrInvalidSignature = errors.New("invalid document signature")

	// ErrInvalidCertificate is returned when certificate is invalid
	ErrInvalidCertificate = errors.New("invalid certificate")

	// ErrDocumentExpired is returned when document timestamp is too old
	ErrDocumentExpired = errors.New("attestation document expired")

	// ErrInvalidPCR is returned when PCR value is invalid
	ErrInvalidPCR = errors.New("invalid PCR value")

	// ErrNonceMismatch is returned when nonce doesn't match expected
	ErrNonceMismatch = errors.New("nonce mismatch")

	// ErrInvalidPayload is returned when payload structure is invalid
	ErrInvalidPayload = errors.New("invalid payload structure")

	// ErrUnsupportedAlgorithm is returned for unsupported algorithms
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")

	// ErrCBORDecodeError is returned when CBOR decoding fails
	ErrCBORDecodeError = errors.New("CBOR decode error")
)

// =============================================================================
// Attestation Document Types
// =============================================================================

// AttestationDocument represents a complete AWS Nitro attestation document
// in COSE_Sign1 format
type AttestationDocument struct {
	// Protected is the protected header (CBOR-encoded)
	Protected []byte

	// Unprotected is the unprotected header (empty for Nitro)
	Unprotected map[interface{}]interface{}

	// Payload contains the attestation data
	Payload *DocumentPayload

	// Signature is the ECDSA signature
	Signature []byte

	// RawPayload is the raw CBOR-encoded payload (for signature verification)
	RawPayload []byte

	// RawDocument is the complete raw document
	RawDocument []byte
}

// DocumentPayload represents the payload of an attestation document
type DocumentPayload struct {
	// ModuleID is the enclave image ID
	ModuleID string `cbor:"module_id"`

	// Digest is the digest algorithm used (SHA384)
	Digest string `cbor:"digest"`

	// Timestamp is Unix timestamp in milliseconds
	Timestamp uint64 `cbor:"timestamp"`

	// PCRs contains Platform Configuration Register values
	// Map of PCR index (0-15) to 48-byte SHA-384 digest
	PCRs map[int][]byte `cbor:"pcrs"`

	// Certificate is the DER-encoded attestation certificate
	Certificate []byte `cbor:"certificate"`

	// CABundle is the certificate chain (array of DER-encoded certs)
	CABundle [][]byte `cbor:"cabundle"`

	// PublicKey is optional user-provided public key
	PublicKey []byte `cbor:"public_key,omitempty"`

	// UserData is optional user-provided data (up to 1KB)
	UserData []byte `cbor:"user_data,omitempty"`

	// Nonce is optional nonce for freshness
	Nonce []byte `cbor:"nonce,omitempty"`
}

// =============================================================================
// Document Parsing
// =============================================================================

// ParseDocument parses a CBOR/COSE-encoded attestation document
//
// The document is expected to be in COSE_Sign1 format:
// Tag(18, [protected, unprotected, payload, signature])
func ParseDocument(data []byte) (*AttestationDocument, error) {
	if len(data) < 10 {
		return nil, ErrInvalidDocument
	}

	doc := &AttestationDocument{
		RawDocument: data,
	}

	// Parse COSE_Sign1 structure
	reader := newCBORReader(data)

	// Check for COSE_Sign1 tag (18)
	tag, err := reader.readTag()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read tag: %v", ErrInvalidDocument, err)
	}
	if tag != COSESign1Tag {
		return nil, fmt.Errorf("%w: expected COSE_Sign1 tag (18), got %d", ErrInvalidDocument, tag)
	}

	// Read array header (should be 4 elements)
	arrayLen, err := reader.readArrayHeader()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read array header: %v", ErrInvalidDocument, err)
	}
	if arrayLen != 4 {
		return nil, fmt.Errorf("%w: expected 4 elements, got %d", ErrInvalidDocument, arrayLen)
	}

	// Read protected header (bstr)
	doc.Protected, err = reader.readByteString()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read protected header: %v", ErrInvalidDocument, err)
	}

	// Read unprotected header (map - should be empty)
	doc.Unprotected, err = reader.readMap()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read unprotected header: %v", ErrInvalidDocument, err)
	}

	// Read payload (bstr containing CBOR-encoded payload)
	doc.RawPayload, err = reader.readByteString()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read payload: %v", ErrInvalidDocument, err)
	}

	// Read signature (bstr)
	doc.Signature, err = reader.readByteString()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read signature: %v", ErrInvalidDocument, err)
	}

	// Parse the payload
	doc.Payload, err = parsePayload(doc.RawPayload)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}

	return doc, nil
}

// parsePayload parses the CBOR-encoded document payload
func parsePayload(data []byte) (*DocumentPayload, error) {
	if len(data) == 0 {
		return nil, errors.New("empty payload")
	}

	reader := newCBORReader(data)
	payload := &DocumentPayload{
		PCRs: make(map[int][]byte),
	}

	// Read map header
	mapLen, err := reader.readMapHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to read payload map: %v", err)
	}

	for i := 0; i < mapLen; i++ {
		// Read key (text string)
		key, err := reader.readTextString()
		if err != nil {
			return nil, fmt.Errorf("failed to read key: %v", err)
		}

		switch key {
		case "module_id":
			payload.ModuleID, err = reader.readTextString()
		case "digest":
			payload.Digest, err = reader.readTextString()
		case "timestamp":
			payload.Timestamp, err = reader.readUint64()
		case "pcrs":
			payload.PCRs, err = reader.readPCRMap()
		case "certificate":
			payload.Certificate, err = reader.readByteString()
		case "cabundle":
			payload.CABundle, err = reader.readByteStringArray()
		case "public_key":
			payload.PublicKey, err = reader.readByteStringOrNil()
		case "user_data":
			payload.UserData, err = reader.readByteStringOrNil()
		case "nonce":
			payload.Nonce, err = reader.readByteStringOrNil()
		default:
			// Skip unknown fields
			err = reader.skipValue()
		}

		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %v", key, err)
		}
	}

	return payload, nil
}

// =============================================================================
// Document Serialization
// =============================================================================

// SerializeDocument serializes an attestation document to CBOR/COSE format
func SerializeDocument(doc *AttestationDocument) ([]byte, error) {
	if doc == nil {
		return nil, errors.New("nil document")
	}

	// If we have the raw document, return it
	if len(doc.RawDocument) > 0 {
		return doc.RawDocument, nil
	}

	writer := newCBORWriter()

	// Write COSE_Sign1 tag
	writer.writeTag(COSESign1Tag)

	// Write array header (4 elements)
	writer.writeArrayHeader(4)

	// Write protected header
	writer.writeByteString(doc.Protected)

	// Write unprotected header (empty map)
	writer.writeMapHeader(0)

	// Serialize payload if needed
	if len(doc.RawPayload) == 0 && doc.Payload != nil {
		doc.RawPayload = serializePayload(doc.Payload)
	}
	writer.writeByteString(doc.RawPayload)

	// Write signature
	writer.writeByteString(doc.Signature)

	return writer.bytes(), nil
}

// serializePayload serializes the document payload to CBOR
func serializePayload(payload *DocumentPayload) []byte {
	writer := newCBORWriter()

	// Count non-empty fields
	fieldCount := 5 // module_id, digest, timestamp, pcrs, certificate, cabundle
	if len(payload.CABundle) > 0 {
		fieldCount++
	}
	if len(payload.PublicKey) > 0 {
		fieldCount++
	}
	if len(payload.UserData) > 0 {
		fieldCount++
	}
	if len(payload.Nonce) > 0 {
		fieldCount++
	}

	writer.writeMapHeader(fieldCount)

	// module_id
	writer.writeTextString("module_id")
	writer.writeTextString(payload.ModuleID)

	// digest
	writer.writeTextString("digest")
	writer.writeTextString(payload.Digest)

	// timestamp
	writer.writeTextString("timestamp")
	writer.writeUint64(payload.Timestamp)

	// pcrs
	writer.writeTextString("pcrs")
	writer.writeMapHeader(len(payload.PCRs))
	for idx, val := range payload.PCRs {
		writer.writeInt(idx)
		writer.writeByteString(val)
	}

	// certificate
	writer.writeTextString("certificate")
	writer.writeByteString(payload.Certificate)

	// cabundle
	if len(payload.CABundle) > 0 {
		writer.writeTextString("cabundle")
		writer.writeArrayHeader(len(payload.CABundle))
		for _, cert := range payload.CABundle {
			writer.writeByteString(cert)
		}
	}

	// public_key
	if len(payload.PublicKey) > 0 {
		writer.writeTextString("public_key")
		writer.writeByteString(payload.PublicKey)
	}

	// user_data
	if len(payload.UserData) > 0 {
		writer.writeTextString("user_data")
		writer.writeByteString(payload.UserData)
	}

	// nonce
	if len(payload.Nonce) > 0 {
		writer.writeTextString("nonce")
		writer.writeByteString(payload.Nonce)
	}

	return writer.bytes()
}

// =============================================================================
// Document Validation
// =============================================================================

// ValidateDocument performs structural validation on an attestation document
func ValidateDocument(doc *AttestationDocument) error {
	if doc == nil {
		return ErrInvalidDocument
	}

	if doc.Payload == nil {
		return fmt.Errorf("%w: missing payload", ErrInvalidDocument)
	}

	payload := doc.Payload

	// Validate module_id
	if payload.ModuleID == "" {
		return fmt.Errorf("%w: missing module_id", ErrInvalidDocument)
	}
	if len(payload.ModuleID) > MaxModuleIDSize {
		return fmt.Errorf("%w: module_id too long", ErrInvalidDocument)
	}

	// Validate digest algorithm
	if payload.Digest != DigestAlgorithmSHA384 {
		return fmt.Errorf("%w: digest must be SHA384", ErrInvalidDocument)
	}

	// Validate timestamp
	if payload.Timestamp == 0 {
		return fmt.Errorf("%w: missing timestamp", ErrInvalidDocument)
	}

	// Validate PCRs
	if len(payload.PCRs) == 0 {
		return fmt.Errorf("%w: missing PCRs", ErrInvalidDocument)
	}

	// PCR0 must be present (EIF measurement)
	pcr0, ok := payload.PCRs[PCRIndexEIF]
	if !ok || len(pcr0) == 0 {
		return fmt.Errorf("%w: missing PCR0 (EIF measurement)", ErrInvalidPCR)
	}

	// All PCRs must be correct size
	for idx, pcr := range payload.PCRs {
		if len(pcr) != PCRDigestSize {
			return fmt.Errorf("%w: PCR%d has invalid size %d", ErrInvalidPCR, idx, len(pcr))
		}
	}

	// Validate certificate
	if len(payload.Certificate) == 0 {
		return fmt.Errorf("%w: missing certificate", ErrInvalidDocument)
	}
	if len(payload.Certificate) > MaxCertSize {
		return fmt.Errorf("%w: certificate too large", ErrInvalidDocument)
	}

	// Validate optional fields
	if len(payload.UserData) > MaxUserDataSize {
		return fmt.Errorf("%w: user_data too large", ErrInvalidDocument)
	}
	if len(payload.Nonce) > MaxNonceSize {
		return fmt.Errorf("%w: nonce too large", ErrInvalidDocument)
	}
	if len(payload.PublicKey) > MaxPublicKeySize {
		return fmt.Errorf("%w: public_key too large", ErrInvalidDocument)
	}

	// Validate signature presence
	if len(doc.Signature) == 0 {
		return fmt.Errorf("%w: missing signature", ErrInvalidDocument)
	}

	return nil
}

// =============================================================================
// PCR Validation
// =============================================================================

// ValidatePCRs validates PCR values against expected values
func ValidatePCRs(actual map[int][]byte, expected map[int][]byte) error {
	for idx, expectedValue := range expected {
		actualValue, ok := actual[idx]
		if !ok {
			return fmt.Errorf("%w: PCR%d not present", ErrInvalidPCR, idx)
		}
		if !bytes.Equal(actualValue, expectedValue) {
			return fmt.Errorf("%w: PCR%d mismatch", ErrInvalidPCR, idx)
		}
	}
	return nil
}

// GetPCRDigest computes a combined digest of PCR0, PCR1, and PCR2
func GetPCRDigest(pcrs map[int][]byte) []byte {
	h := sha512.New384()
	if pcr0, ok := pcrs[PCRIndexEIF]; ok {
		h.Write(pcr0)
	}
	if pcr1, ok := pcrs[PCRIndexKernel]; ok {
		h.Write(pcr1)
	}
	if pcr2, ok := pcrs[PCRIndexApp]; ok {
		h.Write(pcr2)
	}
	return h.Sum(nil)
}

// =============================================================================
// Nonce and User Data
// =============================================================================

// ValidateNonce checks if the document nonce matches expected value
func ValidateNonce(doc *AttestationDocument, expected []byte) error {
	if doc == nil || doc.Payload == nil {
		return ErrInvalidDocument
	}
	if !bytes.Equal(doc.Payload.Nonce, expected) {
		return ErrNonceMismatch
	}
	return nil
}

// GetUserData extracts user data from the document
func GetUserData(doc *AttestationDocument) []byte {
	if doc == nil || doc.Payload == nil {
		return nil
	}
	return doc.Payload.UserData
}

// GetPublicKey extracts the public key from the document
func GetPublicKey(doc *AttestationDocument) []byte {
	if doc == nil || doc.Payload == nil {
		return nil
	}
	return doc.Payload.PublicKey
}

// GetTimestamp returns the document timestamp as time.Time
func GetTimestamp(doc *AttestationDocument) time.Time {
	if doc == nil || doc.Payload == nil {
		return time.Time{}
	}
	return time.UnixMilli(int64(doc.Payload.Timestamp)) //nolint:gosec // timestamp won't overflow int64 in practice
}

// =============================================================================
// Certificate Extraction
// =============================================================================

// GetCertificate parses and returns the attestation certificate
func GetCertificate(doc *AttestationDocument) (*x509.Certificate, error) {
	if doc == nil || doc.Payload == nil {
		return nil, ErrInvalidDocument
	}
	if len(doc.Payload.Certificate) == 0 {
		return nil, ErrInvalidCertificate
	}

	cert, err := x509.ParseCertificate(doc.Payload.Certificate)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCertificate, err)
	}

	return cert, nil
}

// GetCABundle parses and returns the CA certificate chain
func GetCABundle(doc *AttestationDocument) ([]*x509.Certificate, error) {
	if doc == nil || doc.Payload == nil {
		return nil, ErrInvalidDocument
	}

	certs := make([]*x509.Certificate, 0, len(doc.Payload.CABundle))
	for i, certData := range doc.Payload.CABundle {
		cert, err := x509.ParseCertificate(certData)
		if err != nil {
			return nil, fmt.Errorf("%w: CA cert %d: %v", ErrInvalidCertificate, i, err)
		}
		certs = append(certs, cert)
	}

	return certs, nil
}

// =============================================================================
// CBOR Reader (lightweight implementation)
// =============================================================================

type cborReader struct {
	data []byte
	pos  int
}

func newCBORReader(data []byte) *cborReader {
	return &cborReader{data: data, pos: 0}
}

func (r *cborReader) readByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, ErrCBORDecodeError
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

func (r *cborReader) readBytes(n int) ([]byte, error) {
	if r.pos+n > len(r.data) {
		return nil, ErrCBORDecodeError
	}
	result := r.data[r.pos : r.pos+n]
	r.pos += n
	return result, nil
}

func (r *cborReader) readHeader() (majorType int, info int, err error) {
	b, err := r.readByte()
	if err != nil {
		return 0, 0, err
	}
	majorType = int(b >> 5)
	info = int(b & 0x1f)
	return majorType, info, nil
}

func (r *cborReader) readLength(info int) (int, error) {
	if info < 24 {
		return info, nil
	}
	switch info {
	case CBORAdditionalInfo1Byte:
		b, err := r.readByte()
		return int(b), err
	case CBORAdditionalInfo2Bytes:
		data, err := r.readBytes(2)
		if err != nil {
			return 0, err
		}
		return int(binary.BigEndian.Uint16(data)), nil
	case CBORAdditionalInfo4Bytes:
		data, err := r.readBytes(4)
		if err != nil {
			return 0, err
		}
		return int(binary.BigEndian.Uint32(data)), nil
	case CBORAdditionalInfo8Bytes:
		data, err := r.readBytes(8)
		if err != nil {
			return 0, err
		}
		return int(binary.BigEndian.Uint64(data)), nil //nolint:gosec // CBOR length won't exceed int in practice
	default:
		return 0, ErrCBORDecodeError
	}
}

func (r *cborReader) readTag() (int, error) {
	major, info, err := r.readHeader()
	if err != nil {
		return 0, err
	}
	if major != CBORMajorTypeTag {
		return 0, fmt.Errorf("expected tag, got major type %d", major)
	}
	return r.readLength(info)
}

func (r *cborReader) readArrayHeader() (int, error) {
	major, info, err := r.readHeader()
	if err != nil {
		return 0, err
	}
	if major != CBORMajorTypeArray {
		return 0, fmt.Errorf("expected array, got major type %d", major)
	}
	return r.readLength(info)
}

func (r *cborReader) readMapHeader() (int, error) {
	major, info, err := r.readHeader()
	if err != nil {
		return 0, err
	}
	if major != CBORMajorTypeMap {
		return 0, fmt.Errorf("expected map, got major type %d", major)
	}
	return r.readLength(info)
}

func (r *cborReader) readByteString() ([]byte, error) {
	major, info, err := r.readHeader()
	if err != nil {
		return nil, err
	}
	if major != CBORMajorTypeByteString {
		return nil, fmt.Errorf("expected byte string, got major type %d", major)
	}
	length, err := r.readLength(info)
	if err != nil {
		return nil, err
	}
	return r.readBytes(length)
}

func (r *cborReader) readByteStringOrNil() ([]byte, error) {
	if r.pos >= len(r.data) {
		return nil, nil
	}
	// Check for null (0xf6)
	if r.data[r.pos] == 0xf6 {
		r.pos++
		return nil, nil
	}
	return r.readByteString()
}

func (r *cborReader) readTextString() (string, error) {
	major, info, err := r.readHeader()
	if err != nil {
		return "", err
	}
	if major != CBORMajorTypeTextString {
		return "", fmt.Errorf("expected text string, got major type %d", major)
	}
	length, err := r.readLength(info)
	if err != nil {
		return "", err
	}
	data, err := r.readBytes(length)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *cborReader) readUint64() (uint64, error) {
	major, info, err := r.readHeader()
	if err != nil {
		return 0, err
	}
	if major != CBORMajorTypeUnsigned {
		return 0, fmt.Errorf("expected unsigned int, got major type %d", major)
	}
	length, err := r.readLength(info)
	if err != nil {
		return 0, err
	}
	return uint64(length), nil //nolint:gosec // length is non-negative from readLength
}

func (r *cborReader) readInt() (int, error) {
	major, info, err := r.readHeader()
	if err != nil {
		return 0, err
	}
	length, err := r.readLength(info)
	if err != nil {
		return 0, err
	}
	if major == CBORMajorTypeUnsigned {
		return length, nil
	}
	if major == CBORMajorTypeNegative {
		return -1 - length, nil
	}
	return 0, fmt.Errorf("expected integer, got major type %d", major)
}

func (r *cborReader) readMap() (map[interface{}]interface{}, error) {
	length, err := r.readMapHeader()
	if err != nil {
		return nil, err
	}
	result := make(map[interface{}]interface{}, length)
	for i := 0; i < length; i++ {
		key, err := r.readAny()
		if err != nil {
			return nil, err
		}
		val, err := r.readAny()
		if err != nil {
			return nil, err
		}
		result[key] = val
	}
	return result, nil
}

func (r *cborReader) readPCRMap() (map[int][]byte, error) {
	length, err := r.readMapHeader()
	if err != nil {
		return nil, err
	}
	result := make(map[int][]byte, length)
	for i := 0; i < length; i++ {
		idx, err := r.readInt()
		if err != nil {
			return nil, err
		}
		val, err := r.readByteString()
		if err != nil {
			return nil, err
		}
		result[idx] = val
	}
	return result, nil
}

func (r *cborReader) readByteStringArray() ([][]byte, error) {
	length, err := r.readArrayHeader()
	if err != nil {
		return nil, err
	}
	result := make([][]byte, length)
	for i := 0; i < length; i++ {
		result[i], err = r.readByteString()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (r *cborReader) readAny() (interface{}, error) {
	if r.pos >= len(r.data) {
		return nil, ErrCBORDecodeError
	}
	major := int(r.data[r.pos] >> 5)

	switch major {
	case CBORMajorTypeUnsigned, CBORMajorTypeNegative:
		return r.readInt()
	case CBORMajorTypeByteString:
		return r.readByteString()
	case CBORMajorTypeTextString:
		return r.readTextString()
	case CBORMajorTypeArray:
		length, err := r.readArrayHeader()
		if err != nil {
			return nil, err
		}
		result := make([]interface{}, length)
		for i := 0; i < length; i++ {
			result[i], err = r.readAny()
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	case CBORMajorTypeMap:
		return r.readMap()
	case CBORMajorTypeFloat:
		// Handle null, true, false, float
		_, info, err := r.readHeader()
		if err != nil {
			return nil, err
		}
		switch info {
		case 20:
			return false, nil
		case 21:
			return true, nil
		case 22, 23:
			return nil, nil
		default:
			return nil, ErrCBORDecodeError
		}
	default:
		return nil, ErrCBORDecodeError
	}
}

func (r *cborReader) skipValue() error {
	_, err := r.readAny()
	return err
}

// =============================================================================
// CBOR Writer (lightweight implementation)
// =============================================================================

type cborWriter struct {
	buf bytes.Buffer
}

func newCBORWriter() *cborWriter {
	return &cborWriter{}
}

func (w *cborWriter) bytes() []byte {
	return w.buf.Bytes()
}

func (w *cborWriter) writeHeader(majorType int, value int) {
	if value < 24 {
		w.buf.WriteByte(byte(majorType<<5) | byte(value))
	} else if value <= 0xff {
		w.buf.WriteByte(byte(majorType<<5) | CBORAdditionalInfo1Byte)
		w.buf.WriteByte(byte(value))
	} else if value <= 0xffff {
		w.buf.WriteByte(byte(majorType<<5) | CBORAdditionalInfo2Bytes)
		var b [2]byte
		binary.BigEndian.PutUint16(b[:], uint16(value))
		w.buf.Write(b[:])
	} else if value <= 0xffffffff {
		w.buf.WriteByte(byte(majorType<<5) | CBORAdditionalInfo4Bytes)
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], uint32(value))
		w.buf.Write(b[:])
	} else {
		w.buf.WriteByte(byte(majorType<<5) | CBORAdditionalInfo8Bytes)
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], uint64(value))
		w.buf.Write(b[:])
	}
}

func (w *cborWriter) writeTag(tag int) {
	w.writeHeader(CBORMajorTypeTag, tag)
}

func (w *cborWriter) writeArrayHeader(length int) {
	w.writeHeader(CBORMajorTypeArray, length)
}

func (w *cborWriter) writeMapHeader(length int) {
	w.writeHeader(CBORMajorTypeMap, length)
}

func (w *cborWriter) writeByteString(data []byte) {
	w.writeHeader(CBORMajorTypeByteString, len(data))
	w.buf.Write(data)
}

func (w *cborWriter) writeTextString(s string) {
	w.writeHeader(CBORMajorTypeTextString, len(s))
	w.buf.WriteString(s)
}

func (w *cborWriter) writeInt(v int) {
	if v >= 0 {
		w.writeHeader(CBORMajorTypeUnsigned, v)
	} else {
		w.writeHeader(CBORMajorTypeNegative, -1-v)
	}
}

func (w *cborWriter) writeUint64(v uint64) {
	if v <= 0xffffffff {
		w.writeHeader(CBORMajorTypeUnsigned, int(v))
	} else {
		w.buf.WriteByte(byte(CBORMajorTypeUnsigned<<5) | CBORAdditionalInfo8Bytes)
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], v)
		w.buf.Write(b[:])
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// PCRMapToHex converts a PCR map to hex string map (for display/logging)
func PCRMapToHex(pcrs map[int][]byte) map[int]string {
	result := make(map[int]string, len(pcrs))
	for idx, val := range pcrs {
		result[idx] = hex.EncodeToString(val)
	}
	return result
}

// PCRMapFromHex converts a hex string map to PCR map
func PCRMapFromHex(pcrs map[int]string) (map[int][]byte, error) {
	result := make(map[int][]byte, len(pcrs))
	for idx, hexStr := range pcrs {
		val, err := hex.DecodeString(hexStr)
		if err != nil {
			return nil, fmt.Errorf("PCR%d: invalid hex: %w", idx, err)
		}
		if len(val) != PCRDigestSize {
			return nil, fmt.Errorf("PCR%d: invalid size %d", idx, len(val))
		}
		result[idx] = val
	}
	return result, nil
}
