// Package nitro provides AWS Nitro Enclave integration for VirtEngine TEE.
//
// This file implements the Nitro Security Module (NSM) device interaction.
// NSM is the hardware security module inside Nitro Enclaves that provides:
// - Attestation document generation
// - Platform Configuration Register (PCR) operations
// - Secure random number generation
//
// Device: /dev/nsm
// Communication: ioctl calls with CBOR-encoded request/response
//
// Build tags:
// - Default: Simulation mode
// - nitro_hardware: Real NSM device operations
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package nitro

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// =============================================================================
// NSM Constants
// =============================================================================

const (
	// NSM device path
	NSMDevPath = "/dev/nsm"

	// NSM ioctl commands
	NSMIocMsgSend = 0xC0187700

	// NSM message types (request opcodes)
	NSMMsgGetAttestation = 0x01
	NSMMsgDescribePCR    = 0x02
	NSMMsgExtendPCR      = 0x03
	NSMMsgLockPCR        = 0x04
	NSMMsgLockPCRs       = 0x05
	NSMMsgDescribeNSM    = 0x06
	NSMMsgGetRandom      = 0x07

	// NSM response codes
	NSMResponseSuccess        = 0x00
	NSMResponseInvalidArg     = 0x01
	NSMResponseInvalidIndex   = 0x02
	NSMResponseReadOnly       = 0x03
	NSMResponseInvalidOp      = 0x04
	NSMResponseBufferTooSmall = 0x05
	NSMResponseInputTooLarge  = 0x06
	NSMResponseInternalError  = 0x07

	// Maximum sizes
	NSMMaxUserDataSize  = 1024
	NSMMaxNonceSize     = 64
	NSMMaxPublicKeySize = 1024
	NSMMaxRandomBytes   = 256

	// PCR limits
	NSMMinPCRIndex = 0
	NSMMaxPCRIndex = 15
)

// =============================================================================
// NSM Errors
// =============================================================================

var (
	// ErrNSMDeviceNotAvailable is returned when /dev/nsm is not accessible
	ErrNSMDeviceNotAvailable = errors.New("NSM device not available")

	// ErrNSMDeviceNotOpen is returned when device is not open
	ErrNSMDeviceNotOpen = errors.New("NSM device not open")

	// ErrNSMInvalidArgument is returned for invalid arguments
	ErrNSMInvalidArgument = errors.New("invalid argument")

	// ErrNSMInvalidPCRIndex is returned for out-of-range PCR index
	ErrNSMInvalidPCRIndex = errors.New("invalid PCR index")

	// ErrNSMPCRLocked is returned when trying to modify a locked PCR
	ErrNSMPCRLocked = errors.New("PCR is locked")

	// ErrNSMBufferTooSmall is returned when response buffer is too small
	ErrNSMBufferTooSmall = errors.New("buffer too small")

	// ErrNSMInputTooLarge is returned when input exceeds limits
	ErrNSMInputTooLarge = errors.New("input too large")

	// ErrNSMInternalError is returned for internal NSM errors
	ErrNSMInternalError = errors.New("NSM internal error")

	// ErrNSMOperationFailed is returned for generic operation failures
	ErrNSMOperationFailed = errors.New("NSM operation failed")
)

// =============================================================================
// NSM Device Types
// =============================================================================

// NSMDevice represents a connection to the Nitro Security Module
type NSMDevice struct {
	mu sync.Mutex

	// Device state
	fd        *os.File
	simulated bool
	isOpen    bool

	// Simulated state
	simulatedPCRs     [16][]byte
	simulatedLocked   [16]bool
	simulatedModuleID string
}

// NSMInfo contains information about the NSM
type NSMInfo struct {
	// ModuleID is the enclave module identifier
	ModuleID string

	// Version is the NSM API version
	Version string

	// MaxPCRs is the number of available PCRs
	MaxPCRs int

	// Locked indicates which PCRs are locked
	Locked []int

	// Digest is the hash algorithm used
	Digest string
}

// PCRDescription describes a single PCR
type PCRDescription struct {
	// Index is the PCR index (0-15)
	Index int

	// Value is the current PCR value (48 bytes for SHA-384)
	Value []byte

	// Locked indicates if the PCR is locked
	Locked bool
}

// =============================================================================
// NSMDevice Constructor
// =============================================================================

// NewNSMDevice creates a new NSM device handle
//
// The device is not opened automatically; call Open() to connect.
func NewNSMDevice() *NSMDevice {
	return &NSMDevice{
		simulatedModuleID: "sim-enclave-" + randomHex(8),
	}
}

// randomHex generates a random hex string
func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// =============================================================================
// Device Lifecycle
// =============================================================================

// Open opens the NSM device
//
// If /dev/nsm is not available, the device operates in simulation mode.
func (d *NSMDevice) Open() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.isOpen {
		return nil
	}

	// Try to open real NSM device
	fd, err := os.OpenFile(NSMDevPath, os.O_RDWR, 0)
	if err != nil {
		// Fall back to simulation mode
		d.simulated = true
		d.initSimulatedPCRs()
		d.isOpen = true
		return nil
	}

	d.fd = fd
	d.simulated = false
	d.isOpen = true
	return nil
}

// OpenWithMode opens the NSM device with explicit mode
func (d *NSMDevice) OpenWithMode(requireHardware bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.isOpen {
		return nil
	}

	fd, err := os.OpenFile(NSMDevPath, os.O_RDWR, 0)
	if err != nil {
		if requireHardware {
			return fmt.Errorf("%w: %v", ErrNSMDeviceNotAvailable, err)
		}
		d.simulated = true
		d.initSimulatedPCRs()
		d.isOpen = true
		return nil
	}

	d.fd = fd
	d.simulated = false
	d.isOpen = true
	return nil
}

// Close closes the NSM device
func (d *NSMDevice) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return nil
	}

	if d.fd != nil {
		if err := d.fd.Close(); err != nil {
			return err
		}
		d.fd = nil
	}

	d.isOpen = false
	return nil
}

// IsOpen returns true if the device is open
func (d *NSMDevice) IsOpen() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.isOpen
}

// IsSimulated returns true if operating in simulation mode
func (d *NSMDevice) IsSimulated() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.simulated
}

// initSimulatedPCRs initializes simulated PCR values
func (d *NSMDevice) initSimulatedPCRs() {
	for i := 0; i < 16; i++ {
		d.simulatedPCRs[i] = make([]byte, PCRDigestSize)
		// PCR0-2 get simulated measurements
		if i <= 2 {
			h := sha512.Sum384([]byte(fmt.Sprintf("simulated-pcr-%d-%s", i, d.simulatedModuleID)))
			copy(d.simulatedPCRs[i], h[:])
		}
	}
}

// =============================================================================
// Attestation Operations
// =============================================================================

// GetAttestation requests an attestation document from the NSM
//
// Parameters:
//   - userData: Optional user data to include (up to 1KB)
//   - nonce: Optional nonce for freshness (up to 64 bytes)
//   - publicKey: Optional public key to bind to attestation
//
// Returns the CBOR-encoded COSE_Sign1 attestation document
func (d *NSMDevice) GetAttestation(userData, nonce, publicKey []byte) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return nil, ErrNSMDeviceNotOpen
	}

	// Validate inputs
	if len(userData) > NSMMaxUserDataSize {
		return nil, fmt.Errorf("%w: user_data exceeds %d bytes", ErrNSMInputTooLarge, NSMMaxUserDataSize)
	}
	if len(nonce) > NSMMaxNonceSize {
		return nil, fmt.Errorf("%w: nonce exceeds %d bytes", ErrNSMInputTooLarge, NSMMaxNonceSize)
	}
	if len(publicKey) > NSMMaxPublicKeySize {
		return nil, fmt.Errorf("%w: public_key exceeds %d bytes", ErrNSMInputTooLarge, NSMMaxPublicKeySize)
	}

	if d.simulated {
		return d.simulateAttestation(userData, nonce, publicKey)
	}

	return d.hardwareGetAttestation(userData, nonce, publicKey)
}

// hardwareGetAttestation gets attestation from real NSM
func (d *NSMDevice) hardwareGetAttestation(userData, nonce, publicKey []byte) ([]byte, error) {
	// Build attestation request
	request := d.buildAttestationRequest(userData, nonce, publicKey)

	// Send ioctl request
	response, err := d.sendNSMRequest(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNSMOperationFailed, err)
	}

	return response, nil
}

// buildAttestationRequest builds CBOR-encoded attestation request
func (d *NSMDevice) buildAttestationRequest(userData, nonce, publicKey []byte) []byte {
	writer := newCBORWriter()

	// Build request map
	fieldCount := 0
	if len(userData) > 0 {
		fieldCount++
	}
	if len(nonce) > 0 {
		fieldCount++
	}
	if len(publicKey) > 0 {
		fieldCount++
	}

	writer.writeMapHeader(fieldCount)

	if len(userData) > 0 {
		writer.writeTextString("user_data")
		writer.writeByteString(userData)
	}
	if len(nonce) > 0 {
		writer.writeTextString("nonce")
		writer.writeByteString(nonce)
	}
	if len(publicKey) > 0 {
		writer.writeTextString("public_key")
		writer.writeByteString(publicKey)
	}

	return writer.bytes()
}

// simulateAttestation generates a simulated attestation document
func (d *NSMDevice) simulateAttestation(userData, nonce, publicKey []byte) ([]byte, error) {
	// Build simulated payload
	payload := &DocumentPayload{
		ModuleID:  d.simulatedModuleID,
		Digest:    DigestAlgorithmSHA384,
		Timestamp: uint64(time.Now().UnixMilli()),
		PCRs:      make(map[int][]byte),
		UserData:  userData,
		Nonce:     nonce,
		PublicKey: publicKey,
	}

	// Copy PCR values
	for i := 0; i < 16; i++ {
		if len(d.simulatedPCRs[i]) > 0 {
			payload.PCRs[i] = make([]byte, len(d.simulatedPCRs[i]))
			copy(payload.PCRs[i], d.simulatedPCRs[i])
		}
	}

	// Generate simulated certificate
	payload.Certificate = d.generateSimulatedCert()

	// Build document
	doc := &AttestationDocument{
		Payload: payload,
	}

	// Build protected header (algorithm = ES384)
	protectedWriter := newCBORWriter()
	protectedWriter.writeMapHeader(1)
	protectedWriter.writeInt(1) // alg key
	protectedWriter.writeInt(COSEAlgorithmES384)
	doc.Protected = protectedWriter.bytes()

	// Serialize payload
	doc.RawPayload = serializePayload(payload)

	// Generate simulated signature (not cryptographically valid)
	signatureData := sha512.Sum384(append(doc.Protected, doc.RawPayload...))
	doc.Signature = signatureData[:]

	return SerializeDocument(doc)
}

// generateSimulatedCert generates a simulated certificate
func (d *NSMDevice) generateSimulatedCert() []byte {
	// Return a minimal simulated certificate
	// In production, this would be a real X.509 certificate from AWS
	cert := []byte("SIMULATED_CERTIFICATE:" + d.simulatedModuleID)
	return cert
}

// =============================================================================
// PCR Operations
// =============================================================================

// DescribePCR returns the current value and status of a PCR
func (d *NSMDevice) DescribePCR(index int) (*PCRDescription, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return nil, ErrNSMDeviceNotOpen
	}

	if index < NSMMinPCRIndex || index > NSMMaxPCRIndex {
		return nil, ErrNSMInvalidPCRIndex
	}

	if d.simulated {
		return &PCRDescription{
			Index:  index,
			Value:  d.simulatedPCRs[index],
			Locked: d.simulatedLocked[index],
		}, nil
	}

	return d.hardwareDescribePCR(index)
}

// hardwareDescribePCR describes a PCR using real NSM
func (d *NSMDevice) hardwareDescribePCR(index int) (*PCRDescription, error) {
	// Build describe PCR request
	writer := newCBORWriter()
	writer.writeMapHeader(1)
	writer.writeTextString("index")
	writer.writeInt(index)
	request := writer.bytes()

	// Send request
	response, err := d.sendNSMRequestWithOpcode(NSMMsgDescribePCR, request)
	if err != nil {
		return nil, err
	}

	// Parse response
	reader := newCBORReader(response)
	mapLen, err := reader.readMapHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	desc := &PCRDescription{Index: index}
	for i := 0; i < mapLen; i++ {
		key, err := reader.readTextString()
		if err != nil {
			return nil, err
		}
		switch key {
		case "data":
			desc.Value, err = reader.readByteString()
		case "lock":
			locked, readErr := reader.readAny()
			if readErr == nil {
				if b, ok := locked.(bool); ok {
					desc.Locked = b
				}
			}
			err = readErr
		default:
			err = reader.skipValue()
		}
		if err != nil {
			return nil, err
		}
	}

	return desc, nil
}

// ExtendPCR extends a PCR with additional data
//
// The new PCR value is: SHA-384(old_value || data)
func (d *NSMDevice) ExtendPCR(index int, data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return ErrNSMDeviceNotOpen
	}

	if index < NSMMinPCRIndex || index > NSMMaxPCRIndex {
		return ErrNSMInvalidPCRIndex
	}

	if len(data) == 0 {
		return ErrNSMInvalidArgument
	}

	if d.simulated {
		if d.simulatedLocked[index] {
			return ErrNSMPCRLocked
		}
		// Extend: SHA-384(old || new)
		h := sha512.New384()
		h.Write(d.simulatedPCRs[index])
		h.Write(data)
		d.simulatedPCRs[index] = h.Sum(nil)
		return nil
	}

	return d.hardwareExtendPCR(index, data)
}

// hardwareExtendPCR extends a PCR using real NSM
func (d *NSMDevice) hardwareExtendPCR(index int, data []byte) error {
	writer := newCBORWriter()
	writer.writeMapHeader(2)
	writer.writeTextString("index")
	writer.writeInt(index)
	writer.writeTextString("data")
	writer.writeByteString(data)
	request := writer.bytes()

	_, err := d.sendNSMRequestWithOpcode(NSMMsgExtendPCR, request)
	if err != nil {
		return fmt.Errorf("extend PCR failed: %w", err)
	}
	return nil
}

// LockPCR locks a PCR to prevent further modifications
func (d *NSMDevice) LockPCR(index int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return ErrNSMDeviceNotOpen
	}

	if index < NSMMinPCRIndex || index > NSMMaxPCRIndex {
		return ErrNSMInvalidPCRIndex
	}

	if d.simulated {
		d.simulatedLocked[index] = true
		return nil
	}

	return d.hardwareLockPCR(index)
}

// hardwareLockPCR locks a PCR using real NSM
func (d *NSMDevice) hardwareLockPCR(index int) error {
	writer := newCBORWriter()
	writer.writeMapHeader(1)
	writer.writeTextString("index")
	writer.writeInt(index)
	request := writer.bytes()

	_, err := d.sendNSMRequestWithOpcode(NSMMsgLockPCR, request)
	if err != nil {
		return fmt.Errorf("lock PCR failed: %w", err)
	}
	return nil
}

// LockPCRs locks multiple PCRs at once
func (d *NSMDevice) LockPCRs(indices []int) error {
	for _, idx := range indices {
		if err := d.LockPCR(idx); err != nil {
			return fmt.Errorf("failed to lock PCR%d: %w", idx, err)
		}
	}
	return nil
}

// =============================================================================
// Random Number Generation
// =============================================================================

// GetRandomBytes returns cryptographically secure random bytes from NSM
//
// Maximum count is 256 bytes per call.
func (d *NSMDevice) GetRandomBytes(count int) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return nil, ErrNSMDeviceNotOpen
	}

	if count <= 0 || count > NSMMaxRandomBytes {
		return nil, fmt.Errorf("%w: count must be 1-%d", ErrNSMInvalidArgument, NSMMaxRandomBytes)
	}

	if d.simulated {
		result := make([]byte, count)
		_, err := rand.Read(result)
		return result, err
	}

	return d.hardwareGetRandom(count)
}

// hardwareGetRandom gets random bytes from real NSM
func (d *NSMDevice) hardwareGetRandom(count int) ([]byte, error) {
	writer := newCBORWriter()
	writer.writeMapHeader(1)
	writer.writeTextString("size")
	writer.writeInt(count)
	request := writer.bytes()

	response, err := d.sendNSMRequestWithOpcode(NSMMsgGetRandom, request)
	if err != nil {
		return nil, fmt.Errorf("get random failed: %w", err)
	}

	// Parse response
	reader := newCBORReader(response)
	mapLen, err := reader.readMapHeader()
	if err != nil {
		return nil, err
	}

	for i := 0; i < mapLen; i++ {
		key, err := reader.readTextString()
		if err != nil {
			return nil, err
		}
		if key == "random" {
			return reader.readByteString()
		}
		if err := reader.skipValue(); err != nil {
			return nil, err
		}
	}

	return nil, errors.New("random data not found in response")
}

// =============================================================================
// NSM Information
// =============================================================================

// DescribeNSM returns information about the NSM
func (d *NSMDevice) DescribeNSM() (*NSMInfo, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return nil, ErrNSMDeviceNotOpen
	}

	if d.simulated {
		locked := make([]int, 0)
		for i, l := range d.simulatedLocked {
			if l {
				locked = append(locked, i)
			}
		}
		return &NSMInfo{
			ModuleID: d.simulatedModuleID,
			Version:  "1.0.0-sim",
			MaxPCRs:  16,
			Locked:   locked,
			Digest:   DigestAlgorithmSHA384,
		}, nil
	}

	return d.hardwareDescribeNSM()
}

// hardwareDescribeNSM describes NSM using real device
func (d *NSMDevice) hardwareDescribeNSM() (*NSMInfo, error) {
	response, err := d.sendNSMRequestWithOpcode(NSMMsgDescribeNSM, nil)
	if err != nil {
		return nil, err
	}

	reader := newCBORReader(response)
	mapLen, err := reader.readMapHeader()
	if err != nil {
		return nil, err
	}

	info := &NSMInfo{}
	for i := 0; i < mapLen; i++ {
		key, err := reader.readTextString()
		if err != nil {
			return nil, err
		}
		switch key {
		case "module_id":
			info.ModuleID, err = reader.readTextString()
		case "version_major", "version_minor", "version_patch":
			_, err = reader.readInt()
		case "max_pcrs":
			var n int
			n, err = reader.readInt()
			info.MaxPCRs = n
		case "locked_pcrs":
			var arr []interface{}
			length, readErr := reader.readArrayHeader()
			if readErr == nil {
				arr = make([]interface{}, length)
				for j := 0; j < length; j++ {
					arr[j], readErr = reader.readInt()
					if readErr != nil {
						break
					}
				}
				for _, v := range arr {
					if idx, ok := v.(int); ok {
						info.Locked = append(info.Locked, idx)
					}
				}
			}
			err = readErr
		case "digest":
			info.Digest, err = reader.readTextString()
		default:
			err = reader.skipValue()
		}
		if err != nil {
			return nil, err
		}
	}

	return info, nil
}

// =============================================================================
// Low-Level NSM Communication
// =============================================================================

// sendNSMRequest sends a request to the NSM device
func (d *NSMDevice) sendNSMRequest(request []byte) ([]byte, error) {
	return d.sendNSMRequestWithOpcode(NSMMsgGetAttestation, request)
}

// sendNSMRequestWithOpcode sends a request with specific opcode
func (d *NSMDevice) sendNSMRequestWithOpcode(opcode int, request []byte) ([]byte, error) {
	if d.fd == nil {
		return nil, ErrNSMDeviceNotOpen
	}

	// Build full request with opcode header
	fullRequest := buildNSMMessage(opcode, request)

	// Allocate response buffer
	responseBuffer := make([]byte, 16384) // 16KB should be enough for attestation

	// Perform ioctl
	n, err := nsmIoctl(d.fd, fullRequest, responseBuffer)
	if err != nil {
		return nil, err
	}

	// Parse response to check for errors
	if n > 0 {
		if err := checkNSMResponse(responseBuffer[:n]); err != nil {
			return nil, err
		}
		return extractNSMPayload(responseBuffer[:n])
	}

	return nil, ErrNSMOperationFailed
}

// buildNSMMessage builds an NSM message with opcode
func buildNSMMessage(opcode int, payload []byte) []byte {
	writer := newCBORWriter()
	writer.writeMapHeader(2)
	writer.writeTextString("opcode")
	writer.writeInt(opcode)
	if len(payload) > 0 {
		writer.writeTextString("request")
		writer.writeByteString(payload)
	} else {
		writer.writeTextString("request")
		writer.writeMapHeader(0)
	}
	return writer.bytes()
}

// checkNSMResponse checks the response for errors
func checkNSMResponse(response []byte) error {
	if len(response) == 0 {
		return ErrNSMOperationFailed
	}

	reader := newCBORReader(response)
	mapLen, err := reader.readMapHeader()
	if err != nil {
		return err
	}

	for i := 0; i < mapLen; i++ {
		key, err := reader.readTextString()
		if err != nil {
			return err
		}
		if key == "error" {
			errCode, err := reader.readInt()
			if err != nil {
				return err
			}
			return nsmErrorFromCode(errCode)
		}
		if err := reader.skipValue(); err != nil {
			return err
		}
	}

	return nil
}

// extractNSMPayload extracts the payload from NSM response
func extractNSMPayload(response []byte) ([]byte, error) {
	reader := newCBORReader(response)
	mapLen, err := reader.readMapHeader()
	if err != nil {
		return nil, err
	}

	for i := 0; i < mapLen; i++ {
		key, err := reader.readTextString()
		if err != nil {
			return nil, err
		}
		if key == "response" || key == "document" {
			return reader.readByteString()
		}
		if err := reader.skipValue(); err != nil {
			return nil, err
		}
	}

	return nil, errors.New("payload not found in response")
}

// nsmErrorFromCode converts NSM error code to Go error
func nsmErrorFromCode(code int) error {
	switch code {
	case NSMResponseSuccess:
		return nil
	case NSMResponseInvalidArg:
		return ErrNSMInvalidArgument
	case NSMResponseInvalidIndex:
		return ErrNSMInvalidPCRIndex
	case NSMResponseReadOnly:
		return ErrNSMPCRLocked
	case NSMResponseBufferTooSmall:
		return ErrNSMBufferTooSmall
	case NSMResponseInputTooLarge:
		return ErrNSMInputTooLarge
	case NSMResponseInternalError:
		return ErrNSMInternalError
	default:
		return fmt.Errorf("%w: code %d", ErrNSMOperationFailed, code)
	}
}

// nsmIoctl performs an ioctl call to the NSM device
//
// This is a placeholder - the actual implementation uses syscall.Syscall
// which requires platform-specific code.
func nsmIoctl(fd *os.File, request, response []byte) (int, error) {
	// In a real implementation, this would use:
	// syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(), NSMIocMsgSend, uintptr(unsafe.Pointer(&msg)))

	// For now, return an error indicating real hardware is needed
	return 0, errors.New("real NSM ioctl requires nitro_hardware build tag")
}

// =============================================================================
// Thread-Safe Wrappers
// =============================================================================

// NSMSession provides a thread-safe session wrapper for NSM operations
type NSMSession struct {
	device *NSMDevice
}

// NewNSMSession creates a new NSM session
func NewNSMSession() (*NSMSession, error) {
	device := NewNSMDevice()
	if err := device.Open(); err != nil {
		return nil, err
	}
	return &NSMSession{device: device}, nil
}

// Close closes the NSM session
func (s *NSMSession) Close() error {
	return s.device.Close()
}

// GetAttestation requests an attestation document
func (s *NSMSession) GetAttestation(userData, nonce, publicKey []byte) ([]byte, error) {
	return s.device.GetAttestation(userData, nonce, publicKey)
}

// GetRandomBytes returns random bytes
func (s *NSMSession) GetRandomBytes(count int) ([]byte, error) {
	return s.device.GetRandomBytes(count)
}

// DescribePCR returns PCR information
func (s *NSMSession) DescribePCR(index int) (*PCRDescription, error) {
	return s.device.DescribePCR(index)
}

// ExtendPCR extends a PCR
func (s *NSMSession) ExtendPCR(index int, data []byte) error {
	return s.device.ExtendPCR(index, data)
}

// LockPCR locks a PCR
func (s *NSMSession) LockPCR(index int) error {
	return s.device.LockPCR(index)
}

// IsSimulated returns true if running in simulation
func (s *NSMSession) IsSimulated() bool {
	return s.device.IsSimulated()
}
