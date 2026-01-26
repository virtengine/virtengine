package crypto

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// VE-926: Ledger Hardware Wallet Tests
// ============================================================================

// ============================================================================
// LedgerWallet Tests
// ============================================================================

func TestNewLedgerWallet_DefaultConfig(t *testing.T) {
	wallet := NewLedgerWallet(nil)
	require.NotNil(t, wallet)
	require.NotNil(t, wallet.config)
	assert.Equal(t, DefaultLedgerHDPath, wallet.config.DefaultHDPath)
	assert.Equal(t, LedgerConnectionUSB, wallet.config.ConnectionType)
	assert.True(t, wallet.config.RequireConfirmation)
}

func TestNewLedgerWallet_CustomConfig(t *testing.T) {
	config := &LedgerWalletConfig{
		DefaultHDPath:       "m/44'/118'/1'/0/0",
		ConnectionType:      LedgerConnectionBluetooth,
		RequireConfirmation: true,
		Timeout:             30 * time.Second,
		RetryCount:          5,
		HRPPrefix:           "virtengine",
	}

	wallet := NewLedgerWallet(config)
	require.NotNil(t, wallet)
	assert.Equal(t, "m/44'/118'/1'/0/0", wallet.config.DefaultHDPath)
	assert.Equal(t, LedgerConnectionBluetooth, wallet.config.ConnectionType)
	assert.Equal(t, 5, wallet.config.RetryCount)
	assert.Equal(t, "virtengine", wallet.config.HRPPrefix)
}

func TestDefaultLedgerWalletConfig(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	require.NotNil(t, config)

	assert.Equal(t, DefaultLedgerHDPath, config.DefaultHDPath)
	assert.Equal(t, LedgerConnectionUSB, config.ConnectionType)
	assert.True(t, config.RequireConfirmation)
	assert.Equal(t, LedgerDefaultTimeout, config.Timeout)
	assert.Equal(t, LedgerConnectionRetries, config.RetryCount)
}

// ============================================================================
// Connection Tests
// ============================================================================

func TestLedgerWallet_Connect_Success(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, wallet.IsConnected())
}

func TestLedgerWallet_Connect_Failure(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	mockDevice.SetShouldFail(true, ErrLedgerNotConnected)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.Error(t, err)
	assert.False(t, wallet.IsConnected())
}

func TestLedgerWallet_Disconnect(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, wallet.IsConnected())

	err = wallet.Disconnect()
	require.NoError(t, err)
	assert.False(t, wallet.IsConnected())
}

func TestLedgerWallet_IsConnected_NotConnected(t *testing.T) {
	wallet := NewLedgerWallet(nil)
	assert.False(t, wallet.IsConnected())
}

// ============================================================================
// Device Info Tests
// ============================================================================

func TestLedgerWallet_GetDeviceInfo_Success(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	mockDevice.deviceInfo = &LedgerDeviceInfo{
		DeviceType:      LedgerNanoX,
		ConnectionType:  LedgerConnectionUSB,
		AppName:         LedgerCosmosAppName,
		AppVersion:      "2.34.0",
		FirmwareVersion: "2.1.0",
		IsConnected:     true,
		IsLocked:        false,
	}
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	info, err := wallet.GetDeviceInfo(ctx)
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, LedgerNanoX, info.DeviceType)
	assert.Equal(t, LedgerCosmosAppName, info.AppName)
	assert.Equal(t, "2.34.0", info.AppVersion)
	assert.True(t, info.IsConnected)
	assert.False(t, info.IsLocked)
}

func TestLedgerWallet_GetDeviceInfo_NotConnected(t *testing.T) {
	wallet := NewLedgerWallet(nil)

	ctx := context.Background()
	info, err := wallet.GetDeviceInfo(ctx)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLedgerNotConnected))
	assert.Nil(t, info)
}

// ============================================================================
// Address Derivation Tests
// ============================================================================

func TestLedgerWallet_GetAddress_DefaultPath(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	mockDevice.SetAddress(DefaultLedgerHDPath, &LedgerAddress{
		Address:   "cosmos1abc123def456",
		PublicKey: make([]byte, 33),
		HDPath:    DefaultLedgerHDPath,
	})
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	addr, err := wallet.GetAddress(ctx, "", false)
	require.NoError(t, err)
	require.NotNil(t, addr)

	assert.Equal(t, "cosmos1abc123def456", addr.Address)
	assert.Equal(t, DefaultLedgerHDPath, addr.HDPath)
	assert.False(t, addr.DeviceVerified)
}

func TestLedgerWallet_GetAddress_CustomPath(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	customPath := "m/44'/118'/1'/0/5"
	mockDevice.SetAddress(customPath, &LedgerAddress{
		Address:   "cosmos1customaddr",
		PublicKey: make([]byte, 33),
		HDPath:    customPath,
	})
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	addr, err := wallet.GetAddress(ctx, customPath, false)
	require.NoError(t, err)
	require.NotNil(t, addr)

	assert.Equal(t, "cosmos1customaddr", addr.Address)
	assert.Equal(t, customPath, addr.HDPath)
}

func TestLedgerWallet_GetAddress_InvalidPath(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	_, err = wallet.GetAddress(ctx, "invalid/path", false)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLedgerInvalidPath))
}

func TestLedgerWallet_GetAddress_NotConnected(t *testing.T) {
	wallet := NewLedgerWallet(nil)

	ctx := context.Background()
	_, err := wallet.GetAddress(ctx, "", false)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLedgerNotConnected))
}

func TestLedgerWallet_GetAddress_WithDisplay(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	addr, err := wallet.GetAddress(ctx, DefaultLedgerHDPath, true)
	require.NoError(t, err)
	require.NotNil(t, addr)

	assert.True(t, addr.DeviceVerified)
}

func TestLedgerWallet_GetAddress_UserRejected(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	mockDevice.SetUserReject(true)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	_, err = wallet.GetAddress(ctx, DefaultLedgerHDPath, true)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLedgerUserRejected))
}

func TestLedgerWallet_GetAddress_Caching(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	callCount := 0
	mockDevice := NewMockLedgerDevice(config)
	mockDevice.OnGetAddress = func(hdPath string, display bool) (*LedgerAddress, error) {
		callCount++
		return &LedgerAddress{
			Address:        "cosmos1cached",
			PublicKey:      make([]byte, 33),
			HDPath:         hdPath,
			DeviceVerified: display,
		}, nil
	}
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	// First call - should hit device
	addr1, err := wallet.GetAddress(ctx, DefaultLedgerHDPath, false)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second call - should use cache
	addr2, err := wallet.GetAddress(ctx, DefaultLedgerHDPath, false)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount) // Still 1, used cache

	assert.Equal(t, addr1.Address, addr2.Address)
}

func TestLedgerWallet_ClearCache(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	callCount := 0
	mockDevice := NewMockLedgerDevice(config)
	mockDevice.OnGetAddress = func(hdPath string, display bool) (*LedgerAddress, error) {
		callCount++
		return &LedgerAddress{
			Address:   "cosmos1test",
			PublicKey: make([]byte, 33),
			HDPath:    hdPath,
		}, nil
	}
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	// First call
	_, err = wallet.GetAddress(ctx, DefaultLedgerHDPath, false)
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Clear cache
	wallet.ClearCache()

	// Next call should hit device again
	_, err = wallet.GetAddress(ctx, DefaultLedgerHDPath, false)
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

// ============================================================================
// Multiple Address Derivation Tests
// ============================================================================

func TestLedgerWallet_DeriveAddresses(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	addresses, err := wallet.DeriveAddresses(ctx, 0, 0, 5)
	require.NoError(t, err)
	assert.Len(t, addresses, 5)

	// Verify each address has a unique path
	paths := make(map[string]bool)
	for _, addr := range addresses {
		assert.NotEmpty(t, addr.HDPath)
		assert.False(t, paths[addr.HDPath], "paths should be unique")
		paths[addr.HDPath] = true
	}
}

func TestLedgerWallet_DeriveAddresses_ZeroCount(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	_, err = wallet.DeriveAddresses(ctx, 0, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count must be greater than 0")
}

// ============================================================================
// Transaction Signing Tests
// ============================================================================

func TestLedgerWallet_SignTransaction_Success(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	message := []byte(`{"account_number":"0","chain_id":"cosmoshub-4","fee":{"amount":[],"gas":"200000"},"memo":"","msgs":[],"sequence":"0"}`)

	sig, err := wallet.SignTransaction(ctx, DefaultLedgerHDPath, message)
	require.NoError(t, err)
	require.NotNil(t, sig)

	assert.NotEmpty(t, sig.Signature)
	assert.NotEmpty(t, sig.PublicKey)
	assert.Equal(t, DefaultLedgerHDPath, sig.HDPath)
}

func TestLedgerWallet_SignTransaction_NotConnected(t *testing.T) {
	wallet := NewLedgerWallet(nil)

	ctx := context.Background()
	message := []byte("test message")

	_, err := wallet.SignTransaction(ctx, "", message)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLedgerNotConnected))
}

func TestLedgerWallet_SignTransaction_InvalidPath(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	message := []byte("test message")
	_, err = wallet.SignTransaction(ctx, "invalid/path", message)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLedgerInvalidPath))
}

func TestLedgerWallet_SignTransaction_EmptyMessage(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	_, err = wallet.SignTransaction(ctx, "", []byte{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "message cannot be empty")
}

func TestLedgerWallet_SignTransaction_MessageTooLarge(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	// Create a message larger than the limit
	largeMessage := make([]byte, LedgerMaxMessageSize+1)
	_, err = wallet.SignTransaction(ctx, "", largeMessage)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLedgerTransactionTooLarge))
}

func TestLedgerWallet_SignTransaction_UserRejected(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	mockDevice.SetUserReject(true)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	message := []byte("test message")
	_, err = wallet.SignTransaction(ctx, "", message)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLedgerUserRejected))
}

func TestLedgerWallet_SignTransaction_DeviceError(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	// Configure to fail after connection
	mockDevice.SetShouldFail(true, ErrLedgerCommunicationFailed)

	message := []byte("test message")
	_, err = wallet.SignTransaction(ctx, "", message)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLedgerCommunicationFailed))
}

// ============================================================================
// Verify Address Tests
// ============================================================================

func TestLedgerWallet_VerifyAddress(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	addr, err := wallet.VerifyAddress(ctx, DefaultLedgerHDPath)
	require.NoError(t, err)
	require.NotNil(t, addr)

	assert.True(t, addr.DeviceVerified)
}

// ============================================================================
// Public Key Tests
// ============================================================================

func TestLedgerWallet_GetPublicKey_Success(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	expectedPubKey := make([]byte, 33)
	expectedPubKey[0] = 0x02

	mockDevice := NewMockLedgerDevice(config)
	mockDevice.SetAddress(DefaultLedgerHDPath, &LedgerAddress{
		Address:   "cosmos1test",
		PublicKey: expectedPubKey,
		HDPath:    DefaultLedgerHDPath,
	})
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	pubKey, err := wallet.GetPublicKey(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, pubKey)
	assert.Len(t, pubKey, 33)
}

func TestLedgerWallet_GetPublicKey_NotConnected(t *testing.T) {
	wallet := NewLedgerWallet(nil)

	ctx := context.Background()
	_, err := wallet.GetPublicKey(ctx, "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLedgerNotConnected))
}

// ============================================================================
// HD Path Utility Tests
// ============================================================================

func TestBuildHDPath(t *testing.T) {
	tests := []struct {
		coinType     uint32
		account      uint32
		addressIndex uint32
		expected     string
	}{
		{118, 0, 0, "m/44'/118'/0'/0/0"},
		{118, 0, 1, "m/44'/118'/0'/0/1"},
		{118, 1, 0, "m/44'/118'/1'/0/0"},
		{118, 5, 10, "m/44'/118'/5'/0/10"},
		{60, 0, 0, "m/44'/60'/0'/0/0"}, // Ethereum coin type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := BuildHDPath(tt.coinType, tt.account, tt.addressIndex)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseHDPath_Valid(t *testing.T) {
	components, err := ParseHDPath(DefaultLedgerHDPath)
	require.NoError(t, err)
	require.NotNil(t, components)

	assert.Equal(t, uint32(44), components.Purpose)
	assert.Equal(t, uint32(118), components.CoinType)
	assert.Equal(t, uint32(0), components.Account)
	assert.Equal(t, uint32(0), components.Change)
	assert.Equal(t, uint32(0), components.AddressIndex)
}

func TestParseHDPath_Invalid(t *testing.T) {
	_, err := ParseHDPath("invalid/path")
	require.Error(t, err)
}

func TestValidateHDPath(t *testing.T) {
	tests := []struct {
		path    string
		isValid bool
	}{
		{"m/44'/118'/0'/0/0", true},
		{"m/44'/118'/1'/0/5", true},
		{"m/44'/60'/0'/0/0", true},
		{"invalid/path", false},
		{"m/44/118/0/0/0", false}, // Missing hardened markers
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := ValidateHDPath(tt.path)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestIsCosmosHDPath(t *testing.T) {
	tests := []struct {
		path     string
		isCosmos bool
	}{
		{"m/44'/118'/0'/0/0", true},
		{"m/44'/118'/1'/0/5", true},
		{"m/44'/60'/0'/0/0", false},  // Ethereum
		{"m/44'/330'/0'/0/0", false}, // Terra
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := IsCosmosHDPath(tt.path)
			assert.Equal(t, tt.isCosmos, result)
		})
	}
}

// ============================================================================
// Device Type and Connection Type Tests
// ============================================================================

func TestLedgerDeviceTypes(t *testing.T) {
	assert.Equal(t, LedgerDeviceType("nano_s"), LedgerNanoS)
	assert.Equal(t, LedgerDeviceType("nano_x"), LedgerNanoX)
	assert.Equal(t, LedgerDeviceType("nano_s_plus"), LedgerNanoSPlus)
	assert.Equal(t, LedgerDeviceType("unknown"), LedgerUnknown)
}

func TestLedgerConnectionTypes(t *testing.T) {
	assert.Equal(t, LedgerConnectionType("usb"), LedgerConnectionUSB)
	assert.Equal(t, LedgerConnectionType("bluetooth"), LedgerConnectionBluetooth)
}

// ============================================================================
// Error Type Tests
// ============================================================================

func TestLedgerErrors(t *testing.T) {
	errors := []error{
		ErrLedgerNotConnected,
		ErrLedgerAppNotOpen,
		ErrLedgerUserRejected,
		ErrLedgerCommunicationFailed,
		ErrLedgerInvalidResponse,
		ErrLedgerTimeout,
		ErrLedgerDeviceLocked,
		ErrLedgerInvalidPath,
		ErrLedgerTransactionTooLarge,
	}

	for _, err := range errors {
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	}
}

// ============================================================================
// Mock Device Tests
// ============================================================================

func TestMockLedgerDevice_Connect(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	mock := NewMockLedgerDevice(config)

	ctx := context.Background()
	err := mock.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, mock.IsConnected())
}

func TestMockLedgerDevice_Connect_CustomCallback(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	mock := NewMockLedgerDevice(config)

	callbackCalled := false
	mock.OnConnect = func() error {
		callbackCalled = true
		return nil
	}

	ctx := context.Background()
	err := mock.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, callbackCalled)
}

func TestMockLedgerDevice_GetAddress_Generated(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	mock := NewMockLedgerDevice(config)
	mock.SetConnected(true)

	ctx := context.Background()
	addr, err := mock.GetAddress(ctx, DefaultLedgerHDPath, false)
	require.NoError(t, err)
	require.NotNil(t, addr)

	assert.Equal(t, DefaultLedgerHDPath, addr.HDPath)
	assert.Len(t, addr.PublicKey, 33)
}

func TestMockLedgerDevice_SignTransaction_Generated(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	mock := NewMockLedgerDevice(config)
	mock.SetConnected(true)

	ctx := context.Background()
	req := &LedgerSignRequest{
		HDPath:              DefaultLedgerHDPath,
		Message:             []byte("test transaction"),
		RequireConfirmation: true,
	}

	sig, err := mock.SignTransaction(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, sig)

	assert.NotEmpty(t, sig.Signature)
	assert.Equal(t, DefaultLedgerHDPath, sig.HDPath)
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestLedgerWallet_ConcurrentAccess(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	var wg sync.WaitGroup
	errChan := make(chan error, 10)

	// Spawn multiple goroutines accessing the wallet
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			path := BuildHDPath(118, 0, uint32(idx))
			_, err := wallet.GetAddress(ctx, path, false)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("concurrent access error: %v", err)
	}
}

func TestLedgerWallet_ConcurrentSignAndGetAddress(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	var wg sync.WaitGroup
	errChan := make(chan error, 20)

	// Concurrent address derivations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := BuildHDPath(118, 0, uint32(idx))
			_, err := wallet.GetAddress(ctx, path, false)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	// Concurrent signatures
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := BuildHDPath(118, 0, uint32(idx))
			msg := []byte("test message " + string(rune(idx)))
			_, err := wallet.SignTransaction(ctx, path, msg)
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("concurrent operation error: %v", err)
	}
}

// ============================================================================
// Device Discovery Tests
// ============================================================================

func TestDiscoverLedgerDevices(t *testing.T) {
	devices, err := DiscoverLedgerDevices()
	require.NoError(t, err)
	// In test environment, no real devices are expected
	assert.NotNil(t, devices)
}

// ============================================================================
// Context Cancellation Tests
// ============================================================================

func TestLedgerWallet_Connect_ContextCancelled(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	mockDevice.OnConnect = func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}
	wallet.SetDevice(mockDevice)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := wallet.Connect(ctx)
	// Should either succeed quickly or be cancelled
	if err != nil {
		assert.Equal(t, context.Canceled, err)
	}
}

// ============================================================================
// Address Verification Tests
// ============================================================================

func TestPublicKeyToAddress(t *testing.T) {
	// Create a valid compressed public key (33 bytes, starts with 0x02 or 0x03)
	pubKey := make([]byte, 33)
	pubKey[0] = 0x02
	for i := 1; i < 33; i++ {
		pubKey[i] = byte(i)
	}

	addr, err := PublicKeyToAddress(pubKey, "cosmos")
	require.NoError(t, err)
	assert.NotEmpty(t, addr)
	assert.Contains(t, addr, "cosmos1")
}

func TestPublicKeyToAddress_InvalidLength(t *testing.T) {
	shortKey := make([]byte, 32) // Should be 33
	_, err := PublicKeyToAddress(shortKey, "cosmos")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key length")
}

// ============================================================================
// Sign Request Tests
// ============================================================================

func TestLedgerSignRequest(t *testing.T) {
	req := &LedgerSignRequest{
		HDPath:              "m/44'/118'/0'/0/0",
		Message:             []byte("test transaction"),
		RequireConfirmation: true,
		Timeout:             30 * time.Second,
	}

	assert.Equal(t, "m/44'/118'/0'/0/0", req.HDPath)
	assert.Equal(t, []byte("test transaction"), req.Message)
	assert.True(t, req.RequireConfirmation)
	assert.Equal(t, 30*time.Second, req.Timeout)
}

// ============================================================================
// Constants Tests
// ============================================================================

func TestLedgerConstants(t *testing.T) {
	assert.Equal(t, "m/44'/118'/0'/0/0", DefaultLedgerHDPath)
	assert.Equal(t, "Cosmos", LedgerCosmosAppName)
	assert.Equal(t, uint32(118), LedgerCoinType)
	assert.Equal(t, 10*1024, LedgerMaxMessageSize)
	assert.Equal(t, 60*time.Second, LedgerDefaultTimeout)
	assert.Equal(t, 3, LedgerConnectionRetries)
	assert.Equal(t, 500*time.Millisecond, LedgerRetryDelay)
}

// ============================================================================
// Integration-Style Tests (with mock)
// ============================================================================

func TestLedgerWallet_FullWorkflow(t *testing.T) {
	// Test a complete workflow: connect, get address, sign, disconnect
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()

	// 1. Connect
	err := wallet.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, wallet.IsConnected())

	// 2. Get device info
	info, err := wallet.GetDeviceInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, LedgerCosmosAppName, info.AppName)

	// 3. Derive address
	addr, err := wallet.GetAddress(ctx, DefaultLedgerHDPath, false)
	require.NoError(t, err)
	assert.NotEmpty(t, addr.Address)

	// 4. Verify address on device
	verifiedAddr, err := wallet.VerifyAddress(ctx, DefaultLedgerHDPath)
	require.NoError(t, err)
	assert.True(t, verifiedAddr.DeviceVerified)

	// 5. Sign a transaction
	txBytes := []byte(`{"account_number":"1","chain_id":"test-chain","fee":{"amount":[{"amount":"1000","denom":"uatom"}],"gas":"100000"},"memo":"test","msgs":[{"type":"cosmos-sdk/MsgSend","value":{"amount":[{"amount":"1000000","denom":"uatom"}],"from_address":"cosmos1...","to_address":"cosmos1..."}}],"sequence":"0"}`)
	sig, err := wallet.SignTransaction(ctx, DefaultLedgerHDPath, txBytes)
	require.NoError(t, err)
	assert.NotEmpty(t, sig.Signature)

	// 6. Disconnect
	err = wallet.Disconnect()
	require.NoError(t, err)
	assert.False(t, wallet.IsConnected())
}

func TestLedgerWallet_MultiAccountDerivation(t *testing.T) {
	config := DefaultLedgerWalletConfig()
	wallet := NewLedgerWallet(config)

	mockDevice := NewMockLedgerDevice(config)
	wallet.SetDevice(mockDevice)

	ctx := context.Background()
	err := wallet.Connect(ctx)
	require.NoError(t, err)

	// Derive addresses from multiple accounts
	accounts := make(map[string]*LedgerAddress)

	for account := uint32(0); account < 3; account++ {
		for index := uint32(0); index < 3; index++ {
			path := BuildHDPath(LedgerCoinType, account, index)
			addr, err := wallet.GetAddress(ctx, path, false)
			require.NoError(t, err)

			// Ensure each address is unique
			assert.NotContains(t, accounts, addr.Address)
			accounts[addr.Address] = addr
		}
	}

	assert.Len(t, accounts, 9) // 3 accounts * 3 indices
}
