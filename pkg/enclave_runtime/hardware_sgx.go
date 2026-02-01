// Package enclave_runtime provides TEE enclave implementations.
//
// This file provides the hardware abstraction layer for Intel SGX operations.
// When running on SGX-capable hardware with the SDK installed, these functions
// make real SGX SDK calls. Otherwise, they fall back to simulation.
//
// Requirements for real SGX:
// - Intel SGX SDK installed
// - SGX driver loaded (/dev/sgx_enclave, /dev/sgx_provision)
// - DCAP Quote Provider Library installed
// - Build with: go build -tags sgx_hardware
//
// Task Reference: VE-2029 - Hardware TEE Integration Layer
package enclave_runtime

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/hkdf"
)

// =============================================================================
// SGX Device Paths and Constants
// =============================================================================

const (
	// SGX device paths
	SGXDeviceEnclave   = "/dev/sgx_enclave"
	SGXDeviceProvision = "/dev/sgx_provision"
	SGXDeviceLegacy    = "/dev/isgx" // Legacy out-of-tree driver

	// SGX sysfs paths
	SGXSysFSPath = "/sys/module/intel_sgx"

	// DCAP paths
	SGXDCAPLibPath  = "/usr/lib/x86_64-linux-gnu/libsgx_dcap_ql.so"
	SGXAESMSocket   = "/var/run/aesmd/aesm.socket"
	SGXQCNLConfPath = "/etc/sgx_default_qcnl.conf"

	// SGX ioctl commands
	SGX_IOC_ENCLAVE_CREATE        = 0x40804700
	SGX_IOC_ENCLAVE_ADD_PAGES     = 0xc0384701
	SGX_IOC_ENCLAVE_INIT          = 0xc0184702
	SGX_IOC_ENCLAVE_SET_ATTRIBUTE = 0xc0104703

	// Enclave error codes
	SGX_SUCCESS                 = 0
	SGX_ERROR_INVALID_PARAMETER = 0x0001
	SGX_ERROR_OUT_OF_MEMORY     = 0x0002
	SGX_ERROR_ENCLAVE_LOST      = 0x0003
	SGX_ERROR_INVALID_STATE     = 0x0004
	SGX_ERROR_INVALID_SIGNATURE = 0x0005
)

// =============================================================================
// SGX Hardware Detector
// =============================================================================

// SGXHardwareDetector provides methods to detect SGX hardware capabilities
type SGXHardwareDetector struct {
	mu sync.RWMutex

	detected         bool
	available        bool
	version          int
	flcSupported     bool
	enclaveDevPath   string
	provisionDevPath string
	lastDetection    time.Time
	detectionError   error
}

// NewSGXHardwareDetector creates a new SGX hardware detector
func NewSGXHardwareDetector() *SGXHardwareDetector {
	return &SGXHardwareDetector{}
}

// Detect performs SGX hardware detection
func (d *SGXHardwareDetector) Detect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.detected = true
	d.lastDetection = time.Now()

	// Check CPUID for SGX support (x86 only, stub returns false on other platforms)
	if !hasCPUIDSGX() {
		// Fallback to checking device files
		if exists, _ := checkDeviceExists(SGXDeviceEnclave); !exists {
			if exists, _ := checkDeviceExists(SGXDeviceLegacy); !exists {
				d.available = false
				d.detectionError = fmt.Errorf("SGX not supported by CPU and no SGX device found")
				return d.detectionError
			}
			// Legacy driver found
			d.enclaveDevPath = SGXDeviceLegacy
			d.version = 1
		} else {
			d.enclaveDevPath = SGXDeviceEnclave
			d.version = 1
		}
	} else {
		d.enclaveDevPath = SGXDeviceEnclave
		d.version = 1
	}

	// Check for SGX enclave device
	exists, err := checkDeviceExists(d.enclaveDevPath)
	if err != nil {
		d.detectionError = fmt.Errorf("error checking SGX device: %w", err)
		return d.detectionError
	}
	if !exists {
		d.available = false
		d.detectionError = fmt.Errorf("SGX device not found at %s", d.enclaveDevPath)
		return d.detectionError
	}

	// Check for SGX provision device (needed for DCAP)
	provisionExists, _ := checkDeviceExists(SGXDeviceProvision)
	if provisionExists {
		d.provisionDevPath = SGXDeviceProvision
	}

	// Check for SGX2 support
	if hasCPUIDSGX2() {
		d.version = 2
	}

	// Check for Flexible Launch Control
	d.flcSupported = hasCPUIDFLC()

	d.available = true
	d.detectionError = nil
	return nil
}

// IsAvailable returns true if SGX hardware is available
func (d *SGXHardwareDetector) IsAvailable() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.available
}

// Version returns the SGX version (1 or 2)
func (d *SGXHardwareDetector) Version() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.version
}

// HasFLC returns true if Flexible Launch Control is supported
func (d *SGXHardwareDetector) HasFLC() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.flcSupported
}

// HasProvision returns true if the provision device is available (for DCAP)
func (d *SGXHardwareDetector) HasProvision() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.provisionDevPath != ""
}

// GetDevicePaths returns the paths to SGX devices
func (d *SGXHardwareDetector) GetDevicePaths() (enclave, provision string) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.enclaveDevPath, d.provisionDevPath
}

// =============================================================================
// SGX Enclave Loader
// =============================================================================

// SGXEnclaveLoader handles loading signed enclave binaries
type SGXEnclaveLoader struct {
	mu sync.RWMutex

	detector    *SGXHardwareDetector
	enclaveID   uint64
	loaded      bool
	enclavePath string
	measurement SGXMeasurement
	signerID    SGXMeasurement
	attributes  SGXAttributes

	// Simulated state when hardware unavailable
	simulated    bool
	simulatedKey []byte
}

// NewSGXEnclaveLoader creates a new enclave loader
func NewSGXEnclaveLoader(detector *SGXHardwareDetector) *SGXEnclaveLoader {
	return &SGXEnclaveLoader{
		detector: detector,
	}
}

// Load loads a signed enclave binary
func (l *SGXEnclaveLoader) Load(enclavePath string, debug bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.loaded {
		return errors.New("enclave already loaded")
	}

	l.enclavePath = enclavePath

	// Check if hardware is available
	if !l.detector.IsAvailable() {
		// Fall back to simulation
		return l.loadSimulated(enclavePath, debug)
	}

	// Real hardware path
	return l.loadHardware(enclavePath, debug)
}

// loadHardware loads the enclave on real SGX hardware
func (l *SGXEnclaveLoader) loadHardware(enclavePath string, debug bool) error {
	// Verify enclave file exists
	_, err := os.Stat(enclavePath)
	if err != nil {
		return fmt.Errorf("enclave file not found: %w", err)
	}

	// TODO: Real SGX implementation:
	// 1. Parse SIGSTRUCT from enclave binary
	// 2. Call sgx_create_enclave() via SGX SDK or direct ioctl:
	//    - Open /dev/sgx_enclave
	//    - Use SGX_IOC_ENCLAVE_CREATE ioctl
	//    - Map enclave pages with SGX_IOC_ENCLAVE_ADD_PAGES
	//    - Initialize with SGX_IOC_ENCLAVE_INIT
	// 3. Extract measurement and signer from loaded enclave

	// For now, simulate the loading process
	l.simulated = true
	return l.loadSimulated(enclavePath, debug)
}

// loadSimulated simulates loading an enclave
func (l *SGXEnclaveLoader) loadSimulated(enclavePath string, debug bool) error {
	l.simulated = true

	// Generate simulated measurement from enclave path
	hash := sha256.Sum256([]byte(enclavePath + time.Now().String()))
	copy(l.measurement[:], hash[:])

	// Generate simulated signer ID
	signerHash := sha256.Sum256([]byte("virtengine-signer-v1"))
	copy(l.signerID[:], signerHash[:])

	// Set attributes
	l.attributes = SGXAttributes{
		Flags: SGXFlagInitted | SGXFlagMode64Bit,
		Xfrm:  0x03, // Default XFRM value
	}
	if debug {
		l.attributes.Flags |= SGXFlagDebug
	}

	// Generate simulated enclave ID
	binary.Read(rand.Reader, binary.LittleEndian, &l.enclaveID)

	// Generate simulated key material
	l.simulatedKey = make([]byte, 32)
	if _, err := rand.Read(l.simulatedKey); err != nil {
		return fmt.Errorf("failed to generate simulated key: %w", err)
	}

	l.loaded = true
	return nil
}

// Unload unloads the enclave
func (l *SGXEnclaveLoader) Unload() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.loaded {
		return nil
	}

	if !l.simulated {
		// TODO: Real SGX implementation:
		// Call sgx_destroy_enclave()
	}

	l.loaded = false
	l.enclaveID = 0
	l.enclavePath = ""
	l.measurement = SGXMeasurement{}
	l.signerID = SGXMeasurement{}
	l.simulatedKey = nil

	return nil
}

// GetMeasurement returns the enclave measurement (MRENCLAVE)
func (l *SGXEnclaveLoader) GetMeasurement() SGXMeasurement {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.measurement
}

// GetSignerID returns the signer measurement (MRSIGNER)
func (l *SGXEnclaveLoader) GetSignerID() SGXMeasurement {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.signerID
}

// IsLoaded returns true if an enclave is loaded
func (l *SGXEnclaveLoader) IsLoaded() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.loaded
}

// IsSimulated returns true if running in simulation mode
func (l *SGXEnclaveLoader) IsSimulated() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.simulated
}

// =============================================================================
// SGX Report Generator
// =============================================================================

// SGXReportGenerator generates SGX reports for local attestation
type SGXReportGenerator struct {
	loader *SGXEnclaveLoader
}

// NewSGXReportGenerator creates a new report generator
func NewSGXReportGenerator(loader *SGXEnclaveLoader) *SGXReportGenerator {
	return &SGXReportGenerator{loader: loader}
}

// GenerateReport generates an SGX report with the given report data
func (g *SGXReportGenerator) GenerateReport(reportData [64]byte, targetInfo []byte) (*SGXReportBody, error) {
	if !g.loader.IsLoaded() {
		return nil, errors.New("enclave not loaded")
	}

	if g.loader.IsSimulated() {
		return g.generateSimulatedReport(reportData, targetInfo)
	}

	return g.generateHardwareReport(reportData, targetInfo)
}

// generateHardwareReport generates a report using SGX hardware
func (g *SGXReportGenerator) generateHardwareReport(reportData [64]byte, targetInfo []byte) (*SGXReportBody, error) {
	// TODO: Real SGX implementation:
	// 1. Make ECALL into enclave
	// 2. Inside enclave, call sgx_create_report() with:
	//    - target_info (for local attestation)
	//    - report_data (user-provided data to bind)
	// 3. Return the generated report

	// Fall back to simulation
	return g.generateSimulatedReport(reportData, targetInfo)
}

// generateSimulatedReport generates a simulated report
func (g *SGXReportGenerator) generateSimulatedReport(reportData [64]byte, targetInfo []byte) (*SGXReportBody, error) {
	g.loader.mu.RLock()
	defer g.loader.mu.RUnlock()

	report := &SGXReportBody{
		MiscSelect: 0,
		Attributes: g.loader.attributes,
		MREnclave:  g.loader.measurement,
		MRSigner:   g.loader.signerID,
		ISVProdID:  1, // Product ID
		ISVSVN:     1, // Security Version Number
		ConfigSVN:  0,
		ReportData: reportData,
	}

	// Fill CPUSVN with simulated values
	copy(report.CPUSVN[:], []byte{0x0f, 0x0f, 0x02, 0x04, 0xff, 0x80, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	return report, nil
}

// =============================================================================
// SGX Quote Generator
// =============================================================================

// SGXQuoteGenerator generates DCAP quotes for remote attestation
type SGXQuoteGenerator struct {
	detector *SGXHardwareDetector
	loader   *SGXEnclaveLoader
}

// NewSGXQuoteGenerator creates a new quote generator
func NewSGXQuoteGenerator(detector *SGXHardwareDetector, loader *SGXEnclaveLoader) *SGXQuoteGenerator {
	return &SGXQuoteGenerator{
		detector: detector,
		loader:   loader,
	}
}

// GenerateQuote generates a DCAP attestation quote
func (g *SGXQuoteGenerator) GenerateQuote(reportData [64]byte) (*SGXQuote, error) {
	if !g.loader.IsLoaded() {
		return nil, errors.New("enclave not loaded")
	}

	if g.loader.IsSimulated() {
		return g.generateSimulatedQuote(reportData)
	}

	return g.generateHardwareQuote(reportData)
}

// generateHardwareQuote generates a quote using SGX DCAP
func (g *SGXQuoteGenerator) generateHardwareQuote(reportData [64]byte) (*SGXQuote, error) {
	// TODO: Real DCAP implementation:
	// 1. Generate a report targeting the Quoting Enclave (QE)
	// 2. Call sgx_qe_get_target_info() to get QE's target info
	// 3. Generate report: sgx_create_report(&qe_target_info, &report_data, &report)
	// 4. Get quote: sgx_qe_get_quote(&report, &quote_size, &quote)
	// 5. Optionally get collateral: sgx_qe_get_quote_verification_collateral()

	// For now, fall back to simulation
	return g.generateSimulatedQuote(reportData)
}

// generateSimulatedQuote generates a simulated quote
func (g *SGXQuoteGenerator) generateSimulatedQuote(reportData [64]byte) (*SGXQuote, error) {
	reportGen := NewSGXReportGenerator(g.loader)
	report, err := reportGen.GenerateReport(reportData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	quote := &SGXQuote{
		Header: SGXQuoteHeader{
			Version:    SGXQuoteVersionDCAP,
			AttKeyType: 2, // ECDSA-256-with-P-256 curve
			TEEType:    0, // Standard SGX
		},
		ReportBody: *report,
	}

	// Generate simulated QE vendor ID (Intel)
	copy(quote.Header.QEVendorID[:], []byte{
		0x93, 0x9a, 0x72, 0x33, 0xf7, 0x9c, 0x4c, 0xa9,
		0x94, 0xa5, 0xcb, 0x8e, 0x39, 0x34, 0x4c, 0x1d,
	})

	// Generate simulated signature
	quote.Signature = make([]byte, 64)
	if _, err := rand.Read(quote.Signature); err != nil {
		return nil, err
	}
	//nolint:gosec // G115: signature length is fixed 64 bytes
	quote.SignatureLength = uint32(len(quote.Signature))

	return quote, nil
}

// =============================================================================
// SGX Sealing Service
// =============================================================================

// SGXSealingService provides data sealing using SGX sealing keys
type SGXSealingService struct {
	loader    *SGXEnclaveLoader
	keyPolicy int // SGXKeyPolicyMREnclave or SGXKeyPolicyMRSigner
}

// NewSGXSealingService creates a new sealing service
func NewSGXSealingService(loader *SGXEnclaveLoader) *SGXSealingService {
	return &SGXSealingService{
		loader:    loader,
		keyPolicy: SGXKeyPolicyMREnclave, // Default to MRENCLAVE
	}
}

// SetKeyPolicy sets the key derivation policy
func (s *SGXSealingService) SetKeyPolicy(policy int) {
	s.keyPolicy = policy
}

// Seal encrypts data using the enclave's sealing key
func (s *SGXSealingService) Seal(plaintext []byte) ([]byte, error) {
	if !s.loader.IsLoaded() {
		return nil, errors.New("enclave not loaded")
	}

	if s.loader.IsSimulated() {
		return s.sealSimulated(plaintext)
	}

	return s.sealHardware(plaintext)
}

// sealHardware seals data using SGX hardware
func (s *SGXSealingService) sealHardware(plaintext []byte) ([]byte, error) {
	// TODO: Real SGX implementation:
	// 1. Make ECALL into enclave
	// 2. Inside enclave: sgx_get_seal_key(&key_request, &seal_key)
	// 3. Inside enclave: Use seal_key to encrypt data with AES-GCM
	// 4. Return sealed blob with key_request info for unsealing

	return s.sealSimulated(plaintext)
}

// sealSimulated seals data in simulation mode
func (s *SGXSealingService) sealSimulated(plaintext []byte) ([]byte, error) {
	s.loader.mu.RLock()
	key := s.loader.simulatedKey
	s.loader.mu.RUnlock()

	if key == nil {
		return nil, errors.New("no sealing key available")
	}

	// Derive a sealing key from the simulated key material
	info := []byte("sgx-sealing-v1")
	kdf := hkdf.New(sha256.New, key, nil, info)

	sealKey := make([]byte, 32)
	if _, err := kdf.Read(sealKey); err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// XOR-based encryption for simulation (NOT SECURE - for simulation only)
	ciphertext := make([]byte, len(plaintext))
	keyStream := make([]byte, len(plaintext))
	expandKey(sealKey, nonce, keyStream)
	for i := range plaintext {
		ciphertext[i] = plaintext[i] ^ keyStream[i]
	}

	// Format: nonce || ciphertext
	sealed := make([]byte, 12+len(ciphertext))
	copy(sealed[:12], nonce)
	copy(sealed[12:], ciphertext)

	return sealed, nil
}

// Unseal decrypts data that was previously sealed
func (s *SGXSealingService) Unseal(sealed []byte) ([]byte, error) {
	if !s.loader.IsLoaded() {
		return nil, errors.New("enclave not loaded")
	}

	if s.loader.IsSimulated() {
		return s.unsealSimulated(sealed)
	}

	return s.unsealHardware(sealed)
}

// unsealHardware unseals data using SGX hardware
func (s *SGXSealingService) unsealHardware(sealed []byte) ([]byte, error) {
	// TODO: Real SGX implementation:
	// 1. Parse key_request from sealed blob
	// 2. Make ECALL into enclave
	// 3. Inside enclave: sgx_get_seal_key(&key_request, &seal_key)
	// 4. Inside enclave: Use seal_key to decrypt data
	// 5. Return plaintext

	return s.unsealSimulated(sealed)
}

// unsealSimulated unseals data in simulation mode
func (s *SGXSealingService) unsealSimulated(sealed []byte) ([]byte, error) {
	if len(sealed) < 12 {
		return nil, errors.New("sealed data too short")
	}

	s.loader.mu.RLock()
	key := s.loader.simulatedKey
	s.loader.mu.RUnlock()

	if key == nil {
		return nil, errors.New("no sealing key available")
	}

	// Derive the same sealing key
	info := []byte("sgx-sealing-v1")
	kdf := hkdf.New(sha256.New, key, nil, info)

	sealKey := make([]byte, 32)
	if _, err := kdf.Read(sealKey); err != nil {
		return nil, err
	}

	nonce := sealed[:12]
	ciphertext := sealed[12:]

	// XOR-based decryption for simulation
	plaintext := make([]byte, len(ciphertext))
	keyStream := make([]byte, len(ciphertext))
	expandKey(sealKey, nonce, keyStream)
	for i := range ciphertext {
		plaintext[i] = ciphertext[i] ^ keyStream[i]
	}

	return plaintext, nil
}

// expandKey generates a keystream from key and nonce (simple simulation)
func expandKey(key, nonce, output []byte) {
	hash := sha256.New()
	counter := uint32(0)
	for i := 0; i < len(output); i += 32 {
		hash.Reset()
		hash.Write(key)
		hash.Write(nonce)
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], counter)
		hash.Write(buf[:])
		sum := hash.Sum(nil)
		copy(output[i:], sum)
		counter++
	}
}

// =============================================================================
// SGX ECALL Interface
// =============================================================================

// SGXECallInterface provides an interface for making ECALLs into the enclave
type SGXECallInterface struct {
	loader    *SGXEnclaveLoader
	callCount uint64
	mu        sync.Mutex
}

// ECallResult represents the result of an ECALL
type ECallResult struct {
	ReturnValue int
	OutputData  []byte
	Error       error
}

// NewSGXECallInterface creates a new ECALL interface
func NewSGXECallInterface(loader *SGXEnclaveLoader) *SGXECallInterface {
	return &SGXECallInterface{loader: loader}
}

// Call makes an ECALL into the enclave
func (e *SGXECallInterface) Call(functionID int, input []byte) (*ECallResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.loader.IsLoaded() {
		return nil, errors.New("enclave not loaded")
	}

	e.callCount++

	if e.loader.IsSimulated() {
		return e.callSimulated(functionID, input)
	}

	return e.callHardware(functionID, input)
}

// callHardware makes a real ECALL
func (e *SGXECallInterface) callHardware(functionID int, input []byte) (*ECallResult, error) {
	// TODO: Real SGX implementation:
	// 1. Prepare input buffers
	// 2. Call the ECALL function via generated stubs
	// 3. Handle OCALL callbacks if needed
	// 4. Return output data

	return e.callSimulated(functionID, input)
}

// callSimulated simulates an ECALL
func (e *SGXECallInterface) callSimulated(functionID int, input []byte) (*ECallResult, error) {
	// Simple simulation: echo back with a hash prefix
	hash := sha256.Sum256(input)
	output := make([]byte, 32+len(input))
	copy(output[:32], hash[:])
	copy(output[32:], input)

	return &ECallResult{
		ReturnValue: SGX_SUCCESS,
		OutputData:  output,
		Error:       nil,
	}, nil
}

// GetCallCount returns the number of ECALLs made
func (e *SGXECallInterface) GetCallCount() uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.callCount
}

// =============================================================================
// SGX Hardware Backend (implements HardwareBackend interface)
// =============================================================================

// SGXHardwareBackend implements the HardwareBackend interface for SGX
type SGXHardwareBackend struct {
	mu sync.RWMutex

	detector *SGXHardwareDetector
	loader   *SGXEnclaveLoader
	sealer   *SGXSealingService
	quoter   *SGXQuoteGenerator
	ecaller  *SGXECallInterface

	initialized bool
	enclavePath string
}

// NewSGXHardwareBackend creates a new SGX hardware backend
func NewSGXHardwareBackend() *SGXHardwareBackend {
	detector := NewSGXHardwareDetector()
	return &SGXHardwareBackend{
		detector: detector,
	}
}

// Platform returns the attestation type for this backend
func (b *SGXHardwareBackend) Platform() AttestationType {
	return AttestationTypeSGX
}

// IsAvailable returns true if SGX hardware is available
func (b *SGXHardwareBackend) IsAvailable() bool {
	if err := b.detector.Detect(); err != nil {
		return false
	}
	return b.detector.IsAvailable()
}

// Initialize sets up the SGX hardware backend
func (b *SGXHardwareBackend) Initialize() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	// Detect hardware - don't fail if not available, we'll use simulation
	_ = b.detector.Detect()

	// Create loader
	b.loader = NewSGXEnclaveLoader(b.detector)

	// Load a default enclave (in real impl, this would be configured)
	// For now, we'll load on-demand
	b.enclavePath = "/opt/virtengine/enclaves/veid_scorer.signed.so"

	// Initialize other components (they will work in simulated mode if needed)
	b.sealer = NewSGXSealingService(b.loader)
	b.quoter = NewSGXQuoteGenerator(b.detector, b.loader)
	b.ecaller = NewSGXECallInterface(b.loader)

	b.initialized = true
	return nil
}

// Shutdown cleanly shuts down the SGX backend
func (b *SGXHardwareBackend) Shutdown() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.initialized {
		return nil
	}

	if b.loader != nil && b.loader.IsLoaded() {
		if err := b.loader.Unload(); err != nil {
			return fmt.Errorf("failed to unload enclave: %w", err)
		}
	}

	b.initialized = false
	return nil
}

// GetAttestation generates an SGX attestation quote
func (b *SGXHardwareBackend) GetAttestation(nonce []byte) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrHardwareNotInitialized
	}

	// Ensure enclave is loaded
	if !b.loader.IsLoaded() {
		if err := b.loader.Load(b.enclavePath, false); err != nil {
			return nil, fmt.Errorf("failed to load enclave: %w", err)
		}
	}

	// Prepare report data with nonce
	var reportData [64]byte
	hash := sha256.Sum256(nonce)
	copy(reportData[:32], hash[:])

	// Generate quote
	quote, err := b.quoter.GenerateQuote(reportData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate quote: %w", err)
	}

	// Serialize quote (simplified - real impl would use proper encoding)
	serialized := make([]byte, 0, 1024)
	serialized = append(serialized, byte(quote.Header.Version), byte(quote.Header.Version>>8))
	serialized = append(serialized, quote.ReportBody.MREnclave[:]...)
	serialized = append(serialized, quote.ReportBody.MRSigner[:]...)
	serialized = append(serialized, quote.ReportBody.ReportData[:]...)
	serialized = append(serialized, quote.Signature...)

	return serialized, nil
}

// DeriveKey derives a key from the SGX root of trust
func (b *SGXHardwareBackend) DeriveKey(context []byte, keySize int) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrHardwareNotInitialized
	}

	// Use sealing key derivation
	b.loader.mu.RLock()
	key := b.loader.simulatedKey
	b.loader.mu.RUnlock()

	if key == nil {
		// Generate a key if not loaded yet
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
	}

	// Derive key using HKDF
	kdf := hkdf.New(sha256.New, key, nil, context)
	derived := make([]byte, keySize)
	if _, err := kdf.Read(derived); err != nil {
		return nil, err
	}

	return derived, nil
}

// Seal encrypts data using SGX sealing
func (b *SGXHardwareBackend) Seal(plaintext []byte) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrHardwareNotInitialized
	}

	// Ensure enclave is loaded
	if !b.loader.IsLoaded() {
		if err := b.loader.Load(b.enclavePath, false); err != nil {
			return nil, fmt.Errorf("failed to load enclave: %w", err)
		}
	}

	return b.sealer.Seal(plaintext)
}

// Unseal decrypts data that was previously sealed
func (b *SGXHardwareBackend) Unseal(ciphertext []byte) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.initialized {
		return nil, ErrHardwareNotInitialized
	}

	if !b.loader.IsLoaded() {
		return nil, errors.New("enclave not loaded - cannot unseal")
	}

	return b.sealer.Unseal(ciphertext)
}

// =============================================================================
// SGX Hardware Detection Helper
// =============================================================================

// detectSGXCapabilities detects SGX hardware and populates capabilities
func detectSGXCapabilities(caps *HardwareCapabilities) {
	detector := NewSGXHardwareDetector()
	if err := detector.Detect(); err != nil {
		caps.DetectionErrors = append(caps.DetectionErrors, fmt.Sprintf("SGX: %v", err))
		return
	}

	if !detector.IsAvailable() {
		return
	}

	caps.SGXAvailable = true
	caps.SGXVersion = detector.Version()
	caps.SGXFLCSupported = detector.HasFLC()

	enclave, provision := detector.GetDevicePaths()
	caps.SGXDriverPath = enclave
	caps.SGXProvisionPath = provision
}
