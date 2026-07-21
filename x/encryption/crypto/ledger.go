package crypto

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ledgercosmos "github.com/cosmos/ledger-cosmos-go"
	"github.com/zondax/hid"
)

// ============================================================================
// VE-926: Ledger Hardware Wallet Integration
// ============================================================================
//
// This module implements Ledger hardware wallet support for VirtEngine.
// SECURITY CRITICAL: Private keys NEVER leave the device.
//
// Supported devices:
// - Ledger Nano S
// - Ledger Nano X
// - Ledger Nano S Plus
//
// Features:
// - BIP-44 address derivation (m/44'/118'/0'/0/x)
// - SECP256K1 signature generation
// - Transaction signing with on-device confirmation
// - Secure device communication
//
// Reference: Cosmos Ledger App specification

// Error format string constant
const errFmtWrapped = "%w: %v"

// Error definitions for Ledger operations
var (
	// ErrLedgerNotConnected indicates no Ledger device is connected
	ErrLedgerNotConnected = errors.New("ledger device not connected")

	// ErrLedgerAppNotOpen indicates the Cosmos app is not open on the device
	ErrLedgerAppNotOpen = errors.New("cosmos app not open on ledger device")

	// ErrLedgerUserRejected indicates the user rejected the operation on device
	ErrLedgerUserRejected = errors.New("user rejected operation on ledger device")

	// ErrLedgerCommunicationFailed indicates communication with the device failed
	ErrLedgerCommunicationFailed = errors.New("ledger communication failed")

	// ErrLedgerInvalidResponse indicates an invalid response from the device
	ErrLedgerInvalidResponse = errors.New("invalid response from ledger device")

	// ErrLedgerTimeout indicates a timeout waiting for device response
	ErrLedgerTimeout = errors.New("ledger device timeout")

	// ErrLedgerDeviceLocked indicates the device is locked
	ErrLedgerDeviceLocked = errors.New("ledger device is locked")

	// ErrLedgerInvalidPath indicates an invalid derivation path
	ErrLedgerInvalidPath = errors.New("invalid derivation path")

	// ErrLedgerTransactionTooLarge indicates the transaction is too large for the device
	ErrLedgerTransactionTooLarge = errors.New("transaction too large for ledger device")
)

// Ledger-related constants
const (
	// DefaultLedgerHDPath is the default Cosmos HD path for Ledger
	DefaultLedgerHDPath = "m/44'/118'/0'/0/0"

	// LedgerCosmosAppName is the name of the Cosmos app on Ledger
	LedgerCosmosAppName = "Cosmos"

	// LedgerCoinType is the coin type for Cosmos (SLIP-44)
	LedgerCoinType uint32 = 118

	// LedgerMaxMessageSize is the maximum transaction size the Ledger can handle
	LedgerMaxMessageSize = 1024 * 10 // 10 KB

	// LedgerDefaultTimeout is the default timeout for Ledger operations
	LedgerDefaultTimeout = 60 * time.Second

	// LedgerConnectionRetries is the number of retries for connection
	LedgerConnectionRetries = 3

	// LedgerRetryDelay is the delay between connection retries
	LedgerRetryDelay = 500 * time.Millisecond

	ledgerVendorID         = 0x2c97
	bip32HardenedOffset    = 0x80000000
	ledgerSignModeLegacyAM = byte(0)
)

type ledgerAppClient interface {
	Close() error
	GetVersion() (*ledgercosmos.VersionInfo, error)
	GetPublicKeySECP256K1(path []uint32) ([]byte, error)
	GetAddressPubKeySECP256K1(path []uint32, hrp string) ([]byte, string, error)
	SignSECP256K1(path []uint32, transaction []byte, p2 byte) ([]byte, error)
}

var (
	findLedgerCosmosUserApp = func() (ledgerAppClient, error) {
		return ledgercosmos.FindLedgerCosmosUserApp()
	}
	enumerateLedgerHIDDevices = hid.Enumerate
)

// LedgerDeviceType represents the type of Ledger device
type LedgerDeviceType string

const (
	// LedgerNanoS represents Ledger Nano S
	LedgerNanoS LedgerDeviceType = "nano_s"

	// LedgerNanoX represents Ledger Nano X (Bluetooth capable)
	LedgerNanoX LedgerDeviceType = "nano_x"

	// LedgerNanoSPlus represents Ledger Nano S Plus
	LedgerNanoSPlus LedgerDeviceType = "nano_s_plus"

	// LedgerUnknown represents an unknown Ledger device
	LedgerUnknown LedgerDeviceType = "unknown"
)

// LedgerConnectionType represents the connection type
type LedgerConnectionType string

const (
	// LedgerConnectionUSB represents USB connection
	LedgerConnectionUSB LedgerConnectionType = "usb"

	// LedgerConnectionBluetooth represents Bluetooth connection (Nano X only)
	LedgerConnectionBluetooth LedgerConnectionType = "bluetooth"
)

// LedgerDeviceInfo contains information about a connected Ledger device
type LedgerDeviceInfo struct {
	// DeviceType is the type of Ledger device
	DeviceType LedgerDeviceType `json:"device_type"`

	// ConnectionType is how the device is connected
	ConnectionType LedgerConnectionType `json:"connection_type"`

	// AppName is the name of the currently open app
	AppName string `json:"app_name"`

	// AppVersion is the version of the currently open app
	AppVersion string `json:"app_version"`

	// FirmwareVersion is the device firmware version
	FirmwareVersion string `json:"firmware_version"`

	// SerialNumber is the device serial number (if available)
	SerialNumber string `json:"serial_number,omitempty"`

	// IsConnected indicates if the device is currently connected
	IsConnected bool `json:"is_connected"`

	// IsLocked indicates if the device is locked
	IsLocked bool `json:"is_locked"`
}

// LedgerAddress represents an address derived from Ledger
type LedgerAddress struct {
	// Address is the bech32 Cosmos address
	Address string `json:"address"`

	// PublicKey is the compressed secp256k1 public key (33 bytes)
	PublicKey []byte `json:"public_key"`

	// HDPath is the derivation path used
	HDPath string `json:"hd_path"`

	// DeviceVerified indicates if the address was verified on device display
	DeviceVerified bool `json:"device_verified"`
}

// LedgerSignature represents a signature from Ledger
type LedgerSignature struct {
	// Signature is the DER-encoded ECDSA signature
	Signature []byte `json:"signature"`

	// PublicKey is the public key that created the signature
	PublicKey []byte `json:"public_key"`

	// HDPath is the derivation path of the signing key
	HDPath string `json:"hd_path"`
}

// LedgerSignRequest contains parameters for signing a transaction
type LedgerSignRequest struct {
	// HDPath is the derivation path for the signing key
	HDPath string `json:"hd_path"`

	// Message is the message to sign (typically amino-encoded transaction)
	Message []byte `json:"message"`

	// RequireConfirmation requires user confirmation on device (always true for production)
	RequireConfirmation bool `json:"require_confirmation"`

	// Timeout is the timeout for the signing operation
	Timeout time.Duration `json:"timeout,omitempty"`
}

// LedgerConfig contains configuration for Ledger wallet operations
type LedgerWalletConfig struct {
	// DefaultHDPath is the default derivation path
	DefaultHDPath string `json:"default_hd_path"`

	// ConnectionType is the preferred connection type
	ConnectionType LedgerConnectionType `json:"connection_type"`

	// RequireConfirmation always require user confirmation (security)
	RequireConfirmation bool `json:"require_confirmation"`

	// Timeout is the default operation timeout
	Timeout time.Duration `json:"timeout"`

	// RetryCount is the number of retries for failed operations
	RetryCount int `json:"retry_count"`

	// HRPPrefix is the bech32 human-readable prefix
	HRPPrefix string `json:"hrp_prefix"`
}

// DefaultLedgerWalletConfig returns the default Ledger wallet configuration
func DefaultLedgerWalletConfig() *LedgerWalletConfig {
	return &LedgerWalletConfig{
		DefaultHDPath:       DefaultLedgerHDPath,
		ConnectionType:      LedgerConnectionUSB,
		RequireConfirmation: true, // Always require confirmation for security
		Timeout:             LedgerDefaultTimeout,
		RetryCount:          LedgerConnectionRetries,
		HRPPrefix:           sdk.Bech32MainPrefix, // "cosmos"
	}
}

// LedgerDevice is the interface for interacting with a Ledger device
// This interface allows for mock implementations in testing
type LedgerDevice interface {
	// Connect establishes connection to the device
	Connect(ctx context.Context) error

	// Disconnect closes the device connection
	Disconnect() error

	// IsConnected returns true if the device is connected
	IsConnected() bool

	// GetDeviceInfo returns information about the connected device
	GetDeviceInfo(ctx context.Context) (*LedgerDeviceInfo, error)

	// GetAddress derives an address at the specified HD path
	GetAddress(ctx context.Context, hdPath string, display bool) (*LedgerAddress, error)

	// SignTransaction signs a transaction message
	SignTransaction(ctx context.Context, req *LedgerSignRequest) (*LedgerSignature, error)

	// GetPublicKey retrieves the public key at the specified path
	GetPublicKey(ctx context.Context, hdPath string) ([]byte, error)
}

// LedgerWallet provides Ledger wallet functionality
type LedgerWallet struct {
	config *LedgerWalletConfig
	device LedgerDevice
	mu     sync.RWMutex

	// Cached addresses for performance
	addressCache map[string]*LedgerAddress
}

// NewLedgerWallet creates a new Ledger wallet with the given configuration
func NewLedgerWallet(config *LedgerWalletConfig) *LedgerWallet {
	if config == nil {
		config = DefaultLedgerWalletConfig()
	}

	return &LedgerWallet{
		config:       config,
		addressCache: make(map[string]*LedgerAddress),
	}
}

// SetDevice sets the Ledger device implementation
// This is used for injecting mock devices in testing
func (lw *LedgerWallet) SetDevice(device LedgerDevice) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	lw.device = device
}

// Connect establishes connection to the Ledger device
func (lw *LedgerWallet) Connect(ctx context.Context) error {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.device == nil {
		// Create real device implementation if not set (for production)
		lw.device = NewRealLedgerDevice(lw.config)
	}

	return lw.device.Connect(ctx)
}

// Disconnect closes the connection to the Ledger device
func (lw *LedgerWallet) Disconnect() error {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.device == nil {
		return nil
	}

	return lw.device.Disconnect()
}

// IsConnected returns true if the Ledger is connected
func (lw *LedgerWallet) IsConnected() bool {
	lw.mu.RLock()
	defer lw.mu.RUnlock()

	if lw.device == nil {
		return false
	}

	return lw.device.IsConnected()
}

// GetDeviceInfo returns information about the connected Ledger device
func (lw *LedgerWallet) GetDeviceInfo(ctx context.Context) (*LedgerDeviceInfo, error) {
	lw.mu.RLock()
	defer lw.mu.RUnlock()

	if lw.device == nil || !lw.device.IsConnected() {
		return nil, ErrLedgerNotConnected
	}

	return lw.device.GetDeviceInfo(ctx)
}

// GetAddress derives an address from the Ledger at the specified HD path
// If display is true, the address will be shown on the device for verification
func (lw *LedgerWallet) GetAddress(ctx context.Context, hdPath string, display bool) (*LedgerAddress, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.device == nil || !lw.device.IsConnected() {
		return nil, ErrLedgerNotConnected
	}

	// Use default path if not specified
	if hdPath == "" {
		hdPath = lw.config.DefaultHDPath
	}

	// Validate HD path
	if _, err := hd.NewParamsFromPath(hdPath); err != nil {
		return nil, fmt.Errorf(errFmtWrapped, ErrLedgerInvalidPath, err)
	}

	// Check cache if not displaying on device
	if !display {
		if cached, ok := lw.addressCache[hdPath]; ok {
			return cached, nil
		}
	}

	// Get address from device
	addr, err := lw.device.GetAddress(ctx, hdPath, display)
	if err != nil {
		return nil, err
	}

	// Cache the address
	lw.addressCache[hdPath] = addr

	return addr, nil
}

// DeriveAddresses derives multiple addresses from the Ledger
func (lw *LedgerWallet) DeriveAddresses(ctx context.Context, account, startIndex, count uint32) ([]*LedgerAddress, error) {
	if count == 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}

	addresses := make([]*LedgerAddress, 0, count)

	for i := uint32(0); i < count; i++ {
		hdPath := BuildHDPath(LedgerCoinType, account, startIndex+i)

		addr, err := lw.GetAddress(ctx, hdPath, false)
		if err != nil {
			return nil, fmt.Errorf("failed to derive address at index %d: %w", startIndex+i, err)
		}

		addresses = append(addresses, addr)
	}

	return addresses, nil
}

// SignTransaction signs a transaction using the Ledger device
// SECURITY: The transaction is displayed on the device for user verification
func (lw *LedgerWallet) SignTransaction(ctx context.Context, hdPath string, message []byte) (*LedgerSignature, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.device == nil || !lw.device.IsConnected() {
		return nil, ErrLedgerNotConnected
	}

	// Use default path if not specified
	if hdPath == "" {
		hdPath = lw.config.DefaultHDPath
	}

	// Validate HD path
	if _, err := hd.NewParamsFromPath(hdPath); err != nil {
		return nil, fmt.Errorf(errFmtWrapped, ErrLedgerInvalidPath, err)
	}

	// Validate message size
	if len(message) > LedgerMaxMessageSize {
		return nil, ErrLedgerTransactionTooLarge
	}

	if len(message) == 0 {
		return nil, fmt.Errorf("message cannot be empty")
	}

	// Create sign request
	req := &LedgerSignRequest{
		HDPath:              hdPath,
		Message:             message,
		RequireConfirmation: lw.config.RequireConfirmation,
		Timeout:             lw.config.Timeout,
	}

	return lw.device.SignTransaction(ctx, req)
}

// VerifyAddress prompts the user to verify an address on the device display
// This should be used before receiving funds to confirm the address is correct
func (lw *LedgerWallet) VerifyAddress(ctx context.Context, hdPath string) (*LedgerAddress, error) {
	return lw.GetAddress(ctx, hdPath, true)
}

// GetPublicKey retrieves the public key at the specified HD path
func (lw *LedgerWallet) GetPublicKey(ctx context.Context, hdPath string) ([]byte, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if lw.device == nil || !lw.device.IsConnected() {
		return nil, ErrLedgerNotConnected
	}

	// Use default path if not specified
	if hdPath == "" {
		hdPath = lw.config.DefaultHDPath
	}

	// Validate HD path
	if _, err := hd.NewParamsFromPath(hdPath); err != nil {
		return nil, fmt.Errorf(errFmtWrapped, ErrLedgerInvalidPath, err)
	}

	return lw.device.GetPublicKey(ctx, hdPath)
}

// ClearCache clears the address cache
func (lw *LedgerWallet) ClearCache() {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	lw.addressCache = make(map[string]*LedgerAddress)
}

// ============================================================================
// HD Path Utilities
// ============================================================================

// BuildHDPath constructs a BIP-44 HD path for Cosmos
// Path format: m/44'/coinType'/account'/0/addressIndex
func BuildHDPath(coinType, account, addressIndex uint32) string {
	return fmt.Sprintf("m/44'/%d'/%d'/0/%d", coinType, account, addressIndex)
}

// ParseHDPath parses an HD path into its components
type HDPathComponents struct {
	Purpose      uint32
	CoinType     uint32
	Account      uint32
	Change       uint32
	AddressIndex uint32
}

// ParseHDPath parses an HD path string into its components
func ParseHDPath(path string) (*HDPathComponents, error) {
	params, err := hd.NewParamsFromPath(path)
	if err != nil {
		return nil, fmt.Errorf("invalid HD path: %w", err)
	}

	// Convert bool Change to uint32 (0 for external, 1 for internal/change)
	var changeIndex uint32
	if params.Change {
		changeIndex = 1
	}

	// Extract components from params string
	return &HDPathComponents{
		Purpose:      44,
		CoinType:     params.CoinType,
		Account:      params.Account,
		Change:       changeIndex,
		AddressIndex: params.AddressIndex,
	}, nil
}

// ValidateHDPath validates an HD path string
func ValidateHDPath(path string) error {
	_, err := hd.NewParamsFromPath(path)
	if err != nil {
		return fmt.Errorf(errFmtWrapped, ErrLedgerInvalidPath, err)
	}
	return nil
}

// IsCosmosHDPath checks if the path uses the Cosmos coin type (118)
func IsCosmosHDPath(path string) bool {
	components, err := ParseHDPath(path)
	if err != nil {
		return false
	}
	return components.CoinType == LedgerCoinType
}

// ============================================================================
// Real Ledger Device Implementation
// ============================================================================

// RealLedgerDevice implements the LedgerDevice interface using actual HID communication
type RealLedgerDevice struct {
	config     *LedgerWalletConfig
	connected  bool
	deviceInfo *LedgerDeviceInfo
	app        ledgerAppClient
	mu         sync.RWMutex
}

// NewRealLedgerDevice creates a new real Ledger device instance
func NewRealLedgerDevice(config *LedgerWalletConfig) *RealLedgerDevice {
	if config == nil {
		config = DefaultLedgerWalletConfig()
	}

	return &RealLedgerDevice{
		config:    config,
		connected: false,
	}
}

// Connect establishes connection to the Ledger device via HID
func (d *RealLedgerDevice) Connect(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	attempts := d.config.RetryCount
	if attempts <= 0 {
		attempts = 1
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := d.tryConnect(ctx); err != nil {
			lastErr = err
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(LedgerRetryDelay):
				continue
			}
		}
		d.connected = true
		return nil
	}

	return fmt.Errorf(errFmtWrapped, ErrLedgerCommunicationFailed, lastErr)
}

// tryConnect attempts a single connection to the device
func (d *RealLedgerDevice) tryConnect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	app, err := findLedgerCosmosUserApp()
	if err != nil {
		return mapLedgerError(err)
	}

	version, err := app.GetVersion()
	if err != nil {
		_ = app.Close()
		return mapLedgerError(err)
	}

	deviceInfo, err := discoverPrimaryLedgerInfo()
	if err != nil {
		_ = app.Close()
		return err
	}

	if deviceInfo == nil {
		deviceInfo = &LedgerDeviceInfo{
			DeviceType:     LedgerUnknown,
			ConnectionType: d.config.ConnectionType,
			IsConnected:    true,
		}
	}

	deviceInfo.AppName = LedgerCosmosAppName
	deviceInfo.AppVersion = version.String()
	deviceInfo.IsConnected = true
	deviceInfo.IsLocked = false

	d.app = app
	d.deviceInfo = deviceInfo

	return nil
}

// Disconnect closes the device connection
func (d *RealLedgerDevice) Disconnect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var err error
	if d.app != nil {
		err = d.app.Close()
	}

	d.app = nil
	d.connected = false
	d.deviceInfo = nil

	if err != nil {
		return mapLedgerError(err)
	}

	return nil
}

// IsConnected returns true if the device is connected
func (d *RealLedgerDevice) IsConnected() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.connected
}

// GetDeviceInfo returns information about the connected device
func (d *RealLedgerDevice) GetDeviceInfo(ctx context.Context) (*LedgerDeviceInfo, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.connected {
		return nil, ErrLedgerNotConnected
	}

	if d.deviceInfo != nil {
		return cloneLedgerDeviceInfo(d.deviceInfo), nil
	}

	return &LedgerDeviceInfo{
		DeviceType:     LedgerUnknown,
		ConnectionType: d.config.ConnectionType,
		AppName:        LedgerCosmosAppName,
		IsConnected:    true,
		IsLocked:       false,
	}, nil
}

// GetAddress derives an address at the specified HD path
func (d *RealLedgerDevice) GetAddress(ctx context.Context, hdPath string, display bool) (*LedgerAddress, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected || d.app == nil {
		return nil, ErrLedgerNotConnected
	}

	path, err := ledgerBIP44Path(hdPath)
	if err != nil {
		return nil, err
	}

	var (
		pubKey  []byte
		address string
	)

	if display {
		pubKey, address, err = d.app.GetAddressPubKeySECP256K1(path, d.config.HRPPrefix)
		if err != nil {
			return nil, mapLedgerError(err)
		}
	} else {
		pubKey, err = d.app.GetPublicKeySECP256K1(path)
		if err != nil {
			return nil, mapLedgerError(err)
		}
		address, err = PublicKeyToAddress(pubKey, d.config.HRPPrefix)
		if err != nil {
			return nil, err
		}
	}

	return &LedgerAddress{
		Address:        address,
		PublicKey:      append([]byte(nil), pubKey...),
		HDPath:         hdPath,
		DeviceVerified: display,
	}, nil
}

// SignTransaction signs a transaction message
func (d *RealLedgerDevice) SignTransaction(ctx context.Context, req *LedgerSignRequest) (*LedgerSignature, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected || d.app == nil {
		return nil, ErrLedgerNotConnected
	}

	// Validate message size
	if len(req.Message) > LedgerMaxMessageSize {
		return nil, ErrLedgerTransactionTooLarge
	}

	// Set timeout
	timeout := req.Timeout
	if timeout == 0 {
		timeout = d.config.Timeout
	}

	// Create context with timeout
	signCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	path, err := ledgerBIP44Path(req.HDPath)
	if err != nil {
		return nil, err
	}

	select {
	case <-signCtx.Done():
		return nil, ErrLedgerTimeout
	default:
	}

	signature, err := d.app.SignSECP256K1(path, req.Message, ledgerSignModeLegacyAM)
	if err != nil {
		return nil, mapLedgerError(err)
	}

	publicKey, err := d.app.GetPublicKeySECP256K1(path)
	if err != nil {
		return nil, mapLedgerError(err)
	}

	return &LedgerSignature{
		Signature: append([]byte(nil), signature...),
		PublicKey: append([]byte(nil), publicKey...),
		HDPath:    req.HDPath,
	}, nil
}

// GetPublicKey retrieves the public key at the specified path
func (d *RealLedgerDevice) GetPublicKey(ctx context.Context, hdPath string) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected || d.app == nil {
		return nil, ErrLedgerNotConnected
	}

	path, err := ledgerBIP44Path(hdPath)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	pubKey, err := d.app.GetPublicKeySECP256K1(path)
	if err != nil {
		return nil, mapLedgerError(err)
	}

	return append([]byte(nil), pubKey...), nil
}

// ============================================================================
// Address Derivation Utilities
// ============================================================================

// PublicKeyToAddress converts a secp256k1 public key to a Cosmos address
func PublicKeyToAddress(pubKey []byte, hrpPrefix string) (string, error) {
	if len(pubKey) != 33 {
		return "", fmt.Errorf("invalid public key length: expected 33 bytes, got %d", len(pubKey))
	}

	// Create secp256k1 public key
	pk := &secp256k1.PubKey{Key: pubKey}

	// Get the address
	addr := sdk.AccAddress(pk.Address())

	// Convert to bech32 with prefix
	return sdk.Bech32ifyAddressBytes(hrpPrefix, addr)
}

// VerifySignature verifies a signature from the Ledger
func VerifySignature(pubKey, message, signature []byte) bool {
	if len(pubKey) != 33 {
		return false
	}

	pk := &secp256k1.PubKey{Key: pubKey}
	return pk.VerifySignature(message, signature)
}

// ============================================================================
// Mock Ledger Device for Testing
// ============================================================================

// MockLedgerDevice is a mock implementation of LedgerDevice for testing
type MockLedgerDevice struct {
	config     *LedgerWalletConfig
	connected  bool
	deviceInfo *LedgerDeviceInfo
	shouldFail bool
	failError  error
	userReject bool
	addresses  map[string]*LedgerAddress
	signatures map[string]*LedgerSignature
	mu         sync.RWMutex

	// Callbacks for testing behavior
	OnConnect    func() error
	OnGetAddress func(hdPath string, display bool) (*LedgerAddress, error)
	OnSign       func(req *LedgerSignRequest) (*LedgerSignature, error)
}

// NewMockLedgerDevice creates a new mock Ledger device
func NewMockLedgerDevice(config *LedgerWalletConfig) *MockLedgerDevice {
	return &MockLedgerDevice{
		config:     config,
		connected:  false,
		addresses:  make(map[string]*LedgerAddress),
		signatures: make(map[string]*LedgerSignature),
		deviceInfo: &LedgerDeviceInfo{
			DeviceType:      LedgerNanoS,
			ConnectionType:  LedgerConnectionUSB,
			AppName:         LedgerCosmosAppName,
			AppVersion:      "2.34.0",
			FirmwareVersion: "2.1.0",
			IsConnected:     true,
			IsLocked:        false,
		},
	}
}

// SetConnected sets the connection state
func (m *MockLedgerDevice) SetConnected(connected bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = connected
}

// SetShouldFail configures the mock to fail with the given error
func (m *MockLedgerDevice) SetShouldFail(shouldFail bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
	m.failError = err
}

// SetUserReject configures the mock to simulate user rejection
func (m *MockLedgerDevice) SetUserReject(reject bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userReject = reject
}

// SetAddress sets a mock address for a given path
func (m *MockLedgerDevice) SetAddress(hdPath string, addr *LedgerAddress) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addresses[hdPath] = addr
}

// SetSignature sets a mock signature for a given path
func (m *MockLedgerDevice) SetSignature(hdPath string, sig *LedgerSignature) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signatures[hdPath] = sig
}

// Connect implements LedgerDevice
func (m *MockLedgerDevice) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.OnConnect != nil {
		return m.OnConnect()
	}

	if m.shouldFail {
		return m.failError
	}

	m.connected = true
	return nil
}

// Disconnect implements LedgerDevice
func (m *MockLedgerDevice) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	return nil
}

// IsConnected implements LedgerDevice
func (m *MockLedgerDevice) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// GetDeviceInfo implements LedgerDevice
func (m *MockLedgerDevice) GetDeviceInfo(ctx context.Context) (*LedgerDeviceInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return nil, ErrLedgerNotConnected
	}

	if m.shouldFail {
		return nil, m.failError
	}

	return m.deviceInfo, nil
}

// GetAddress implements LedgerDevice
func (m *MockLedgerDevice) GetAddress(ctx context.Context, hdPath string, display bool) (*LedgerAddress, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return nil, ErrLedgerNotConnected
	}

	if m.OnGetAddress != nil {
		return m.OnGetAddress(hdPath, display)
	}

	if m.shouldFail {
		return nil, m.failError
	}

	if m.userReject && display {
		return nil, ErrLedgerUserRejected
	}

	// Return cached address if available
	if addr, ok := m.addresses[hdPath]; ok {
		addr.DeviceVerified = display
		return addr, nil
	}

	// Generate a deterministic mock address based on path
	return m.generateMockAddress(hdPath, display)
}

// generateMockAddress creates a mock address for testing
//
//nolint:unparam // result 1 (error) reserved for future address generation failures
func (m *MockLedgerDevice) generateMockAddress(hdPath string, display bool) (*LedgerAddress, error) {
	// Use a deterministic "public key" based on the path
	// This is for testing only - real keys come from the device
	pathHash := sha256Hash([]byte(hdPath))
	mockPubKey := make([]byte, 33)
	mockPubKey[0] = 0x02 // Compressed secp256k1 prefix
	copy(mockPubKey[1:], pathHash[:32])

	// Generate mock address
	addr, err := PublicKeyToAddress(mockPubKey, m.config.HRPPrefix)
	if err != nil {
		// Fallback to a hardcoded address for testing
		addr = "cosmos1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	}

	return &LedgerAddress{
		Address:        addr,
		PublicKey:      mockPubKey,
		HDPath:         hdPath,
		DeviceVerified: display,
	}, nil
}

// sha256Hash computes SHA256 hash for mock key generation
func sha256Hash(data []byte) [32]byte {
	var result [32]byte
	// Simple hash for testing - not cryptographically secure
	for i, b := range data {
		result[i%32] ^= b
	}
	return result
}

// SignTransaction implements LedgerDevice
func (m *MockLedgerDevice) SignTransaction(ctx context.Context, req *LedgerSignRequest) (*LedgerSignature, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return nil, ErrLedgerNotConnected
	}

	if m.OnSign != nil {
		return m.OnSign(req)
	}

	if m.shouldFail {
		return nil, m.failError
	}

	if m.userReject {
		return nil, ErrLedgerUserRejected
	}

	// Check message size
	if len(req.Message) > LedgerMaxMessageSize {
		return nil, ErrLedgerTransactionTooLarge
	}

	// Return cached signature if available
	if sig, ok := m.signatures[req.HDPath]; ok {
		return sig, nil
	}

	// Generate mock signature
	return m.generateMockSignature(req)
}

// generateMockSignature creates a mock signature for testing
//
//nolint:unparam // result 1 (error) reserved for future signature failures
func (m *MockLedgerDevice) generateMockSignature(req *LedgerSignRequest) (*LedgerSignature, error) {
	// Generate deterministic mock signature based on message
	msgHash := sha256Hash(req.Message)

	// Create mock DER signature (this is NOT a valid signature)
	// Real signatures come from the device's secure element
	mockSig := make([]byte, 71)
	mockSig[0] = 0x30               // DER sequence
	mockSig[1] = 0x45               // Length
	mockSig[2] = 0x02               // Integer type
	mockSig[3] = 0x21               // r length
	copy(mockSig[4:36], msgHash[:]) // r value
	mockSig[36] = 0x02              // Integer type
	mockSig[37] = 0x20              // s length
	// Hash the HD path to get a deterministic 32-byte value for the s component
	pathHash := sha256Hash([]byte(req.HDPath))
	copy(mockSig[38:70], pathHash[:]) // s value (hash of path as placeholder)

	// Get the mock public key for this path
	addr, _ := m.generateMockAddress(req.HDPath, false)

	return &LedgerSignature{
		Signature: mockSig,
		PublicKey: addr.PublicKey,
		HDPath:    req.HDPath,
	}, nil
}

// GetPublicKey implements LedgerDevice
func (m *MockLedgerDevice) GetPublicKey(ctx context.Context, hdPath string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		return nil, ErrLedgerNotConnected
	}

	if m.shouldFail {
		return nil, m.failError
	}

	// Return public key from cached address if available
	if addr, ok := m.addresses[hdPath]; ok {
		return addr.PublicKey, nil
	}

	// Generate mock public key
	addr, err := m.generateMockAddress(hdPath, false)
	if err != nil {
		return nil, err
	}

	return addr.PublicKey, nil
}

// ============================================================================
// Device Discovery
// ============================================================================

// DiscoverLedgerDevices searches for connected Ledger devices
func DiscoverLedgerDevices() ([]*LedgerDeviceInfo, error) {
	hidDevices := enumerateLedgerHIDDevices(ledgerVendorID, 0)
	devices := make([]*LedgerDeviceInfo, 0, len(hidDevices))

	for _, info := range hidDevices {
		devices = append(devices, &LedgerDeviceInfo{
			DeviceType:      ledgerDeviceTypeFromProduct(info.Product),
			ConnectionType:  ledgerConnectionTypeFromHID(info),
			SerialNumber:    info.Serial,
			FirmwareVersion: formatLedgerRelease(info.Release),
			IsConnected:     true,
			IsLocked:        false,
		})
	}

	return devices, nil
}

// WaitForDevice waits for a Ledger device to be connected
func WaitForDevice(ctx context.Context, timeout time.Duration) (*LedgerDeviceInfo, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		devices, err := DiscoverLedgerDevices()
		if err != nil {
			return nil, err
		}

		if len(devices) > 0 {
			return devices[0], nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(500 * time.Millisecond):
			continue
		}
	}

	return nil, ErrLedgerNotConnected
}

func ledgerBIP44Path(hdPath string) ([]uint32, error) {
	params, err := hd.NewParamsFromPath(hdPath)
	if err != nil {
		return nil, fmt.Errorf(errFmtWrapped, ErrLedgerInvalidPath, err)
	}

	change := uint32(0)
	if params.Change {
		change = 1
	}

	return []uint32{
		params.Purpose + bip32HardenedOffset,
		params.CoinType + bip32HardenedOffset,
		params.Account + bip32HardenedOffset,
		change,
		params.AddressIndex,
	}, nil
}

func mapLedgerError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return ErrLedgerTimeout
	}

	lower := strings.ToLower(err.Error())
	switch {
	case strings.Contains(lower, "couldn't find ledger device"),
		strings.Contains(lower, "device not connected"),
		strings.Contains(lower, "hid error"),
		strings.Contains(lower, "no device"):
		return fmt.Errorf(errFmtWrapped, ErrLedgerNotConnected, err)
	case strings.Contains(lower, "cla not supported"),
		strings.Contains(lower, "class not supported"),
		strings.Contains(lower, "app is not open"),
		strings.Contains(lower, "cosmos app"):
		return fmt.Errorf(errFmtWrapped, ErrLedgerAppNotOpen, err)
	case strings.Contains(lower, "rejected"),
		strings.Contains(lower, "denied"),
		strings.Contains(lower, "refused"):
		return fmt.Errorf(errFmtWrapped, ErrLedgerUserRejected, err)
	case strings.Contains(lower, "locked"):
		return fmt.Errorf(errFmtWrapped, ErrLedgerDeviceLocked, err)
	case strings.Contains(lower, "timeout"):
		return fmt.Errorf(errFmtWrapped, ErrLedgerTimeout, err)
	default:
		return fmt.Errorf(errFmtWrapped, ErrLedgerCommunicationFailed, err)
	}
}

func cloneLedgerDeviceInfo(info *LedgerDeviceInfo) *LedgerDeviceInfo {
	if info == nil {
		return nil
	}

	cloned := *info
	return &cloned
}

func discoverPrimaryLedgerInfo() (*LedgerDeviceInfo, error) {
	devices, err := DiscoverLedgerDevices()
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, nil
	}
	return cloneLedgerDeviceInfo(devices[0]), nil
}

func ledgerDeviceTypeFromProduct(product string) LedgerDeviceType {
	lower := strings.ToLower(product)
	switch {
	case strings.Contains(lower, "nano x"):
		return LedgerNanoX
	case strings.Contains(lower, "nano s plus"):
		return LedgerNanoSPlus
	case strings.Contains(lower, "nano s"):
		return LedgerNanoS
	default:
		return LedgerUnknown
	}
}

func ledgerConnectionTypeFromHID(info hid.DeviceInfo) LedgerConnectionType {
	switch strings.ToLower(info.BusType.String()) {
	case string(LedgerConnectionBluetooth):
		return LedgerConnectionBluetooth
	default:
		return LedgerConnectionUSB
	}
}

func formatLedgerRelease(release uint16) string {
	if release == 0 {
		return ""
	}
	major := release >> 8
	minor := release & 0x00ff
	return fmt.Sprintf("%d.%d", major, minor)
}
