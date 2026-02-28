package hsm

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hsmlib "github.com/virtengine/virtengine/pkg/keymanagement/hsm"
	vecrypto "github.com/virtengine/virtengine/x/encryption/crypto"
)

type fakeLedgerStatusWallet struct {
	info    *vecrypto.LedgerDeviceInfo
	address *vecrypto.LedgerAddress
	pubKey  []byte
}

func (f *fakeLedgerStatusWallet) Connect(context.Context) error { return nil }
func (f *fakeLedgerStatusWallet) Disconnect() error             { return nil }
func (f *fakeLedgerStatusWallet) GetDeviceInfo(context.Context) (*vecrypto.LedgerDeviceInfo, error) {
	return f.info, nil
}
func (f *fakeLedgerStatusWallet) GetAddress(context.Context, string, bool) (*vecrypto.LedgerAddress, error) {
	return f.address, nil
}
func (f *fakeLedgerStatusWallet) GetPublicKey(context.Context, string) ([]byte, error) {
	return f.pubKey, nil
}

type fakePKCS11StatusClient struct {
	keys []*hsmlib.KeyInfo
}

func (f *fakePKCS11StatusClient) Connect(context.Context) error { return nil }
func (f *fakePKCS11StatusClient) Close() error                  { return nil }
func (f *fakePKCS11StatusClient) ListKeys(context.Context) ([]*hsmlib.KeyInfo, error) {
	return f.keys, nil
}

func TestStatusCmd_LedgerProbe(t *testing.T) {
	t.Cleanup(resetStatusTestHooks())

	configPath := writeStatusConfig(t, hsmlib.Config{
		Backend:           hsmlib.BackendLedger,
		Ledger:            &hsmlib.LedgerConfig{DerivationPath: "m/44'/118'/0'/0/0", HRP: "cosmos"},
		ConnectionTimeout: time.Second,
		OperationTimeout:  time.Second,
		MaxRetries:        1,
	})

	newLedgerStatusClient = func(config *vecrypto.LedgerWalletConfig) ledgerStatusClient {
		return &fakeLedgerStatusWallet{
			info: &vecrypto.LedgerDeviceInfo{
				DeviceType:      vecrypto.LedgerNanoX,
				ConnectionType:  vecrypto.LedgerConnectionUSB,
				AppName:         vecrypto.LedgerCosmosAppName,
				AppVersion:      "2.1.0",
				FirmwareVersion: "1.0.0",
				IsConnected:     true,
			},
			address: &vecrypto.LedgerAddress{
				Address:   "cosmos1ledgerstatus",
				PublicKey: bytes.Repeat([]byte{0x01}, 33),
				HDPath:    config.DefaultHDPath,
			},
			pubKey: bytes.Repeat([]byte{0x01}, 33),
		}
	}

	cmd := StatusCmd()
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"--config", configPath})

	require.NoError(t, cmd.Execute())
	assert.Contains(t, output.String(), "Backend:     ledger")
	assert.Contains(t, output.String(), "Connected:   yes")
	assert.Contains(t, output.String(), "Address:     cosmos1ledgerstatus")
	assert.NotContains(t, output.String(), "mock")
}

func TestStatusCmd_SoftHSMProbe(t *testing.T) {
	t.Cleanup(resetStatusTestHooks())

	tempDir := t.TempDir()
	libraryPath := filepath.Join(tempDir, "libsofthsm2.so")
	require.NoError(t, os.WriteFile(libraryPath, []byte("stub"), 0o600))

	configPath := writeStatusConfig(t, hsmlib.Config{
		Backend: hsmlib.BackendSoftHSM,
		PKCS11: &hsmlib.PKCS11Config{
			LibraryPath: libraryPath,
			SlotID:      7,
		},
		ConnectionTimeout: time.Second,
		OperationTimeout:  time.Second,
		MaxRetries:        1,
	})

	newPKCS11StatusClient = func(config hsmlib.PKCS11Config) (pkcs11StatusClient, error) {
		return &fakePKCS11StatusClient{
			keys: []*hsmlib.KeyInfo{{Label: "provider-key"}},
		}, nil
	}

	cmd := StatusCmd()
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"--config", configPath})

	require.NoError(t, cmd.Execute())
	assert.Contains(t, output.String(), "Backend:     softhsm")
	assert.Contains(t, output.String(), "Connected:   yes")
	assert.Contains(t, output.String(), "Mode:        software token")
	assert.Contains(t, output.String(), "Keys:        1")
	assert.NotContains(t, output.String(), "mock")
}

func writeStatusConfig(t *testing.T, config hsmlib.Config) string {
	t.Helper()

	data, err := json.Marshal(config)
	require.NoError(t, err)

	configPath := filepath.Join(t.TempDir(), "hsm.json")
	require.NoError(t, os.WriteFile(configPath, data, 0o600))
	return configPath
}

func resetStatusTestHooks() func() {
	originalLoad := loadStatusConfig
	originalLedger := newLedgerStatusClient
	originalPKCS11 := newPKCS11StatusClient
	originalStat := statusFileStat

	return func() {
		loadStatusConfig = originalLoad
		newLedgerStatusClient = originalLedger
		newPKCS11StatusClient = originalPKCS11
		statusFileStat = originalStat
	}
}
