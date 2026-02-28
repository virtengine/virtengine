package hsm

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	hsmlib "github.com/virtengine/virtengine/pkg/keymanagement/hsm"
	pkcs11provider "github.com/virtengine/virtengine/pkg/keymanagement/hsm/pkcs11"
	vecrypto "github.com/virtengine/virtengine/x/encryption/crypto"
)

type ledgerStatusClient interface {
	Connect(ctx context.Context) error
	Disconnect() error
	GetDeviceInfo(ctx context.Context) (*vecrypto.LedgerDeviceInfo, error)
	GetAddress(ctx context.Context, hdPath string, display bool) (*vecrypto.LedgerAddress, error)
	GetPublicKey(ctx context.Context, hdPath string) ([]byte, error)
}

type pkcs11StatusClient interface {
	Connect(ctx context.Context) error
	Close() error
	ListKeys(ctx context.Context) ([]*hsmlib.KeyInfo, error)
}

var (
	loadStatusConfig      = readStatusConfig
	newLedgerStatusClient = func(config *vecrypto.LedgerWalletConfig) ledgerStatusClient {
		return vecrypto.NewLedgerWallet(config)
	}
	newPKCS11StatusClient = func(config hsmlib.PKCS11Config) (pkcs11StatusClient, error) {
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		return pkcs11provider.New(config, logger)
	}
	statusFileStat = os.Stat
)

// StatusCmd returns the hsm status command.
func StatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show HSM status and key inventory",
		Long: `Display the current HSM connection status, device information,
and a summary of stored keys.`,
		RunE: runStatus,
	}

	cmd.Flags().String(flagBackend, "", "Override HSM backend")
	cmd.Flags().String(flagConfig, "", "HSM config file path (default: $HOME/.virtengine/hsm.json)")

	return cmd
}

func runStatus(cmd *cobra.Command, _ []string) error {
	configPath, _ := cmd.Flags().GetString(flagConfig)

	cfg, resolvedPath, loadedFromFile, err := loadStatusConfig(configPath)
	if err != nil {
		return err
	}

	if cmd.Flags().Lookup(flagBackend).Changed {
		backend, _ := cmd.Flags().GetString(flagBackend)
		cfg.Backend = hsmlib.BackendType(backend)
	}

	switch cfg.Backend {
	case hsmlib.BackendLedger:
		if cfg.Ledger == nil {
			cfg.Ledger = &hsmlib.LedgerConfig{
				DerivationPath: vecrypto.DefaultLedgerHDPath,
				HRP:            vecrypto.DefaultLedgerWalletConfig().HRPPrefix,
			}
		}
	case hsmlib.BackendPKCS11, hsmlib.BackendSoftHSM:
		if cfg.PKCS11 == nil {
			cfg.PKCS11 = hsmlib.DefaultConfig().PKCS11
		}
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid HSM configuration: %w", err)
	}

	cmd.Printf("HSM Status\n")
	cmd.Printf("  Backend:     %s\n", cfg.Backend)
	cmd.Printf("  Config:      %s\n", resolvedPath)
	if loadedFromFile {
		cmd.Printf("  Source:      file\n")
	} else {
		cmd.Printf("  Source:      defaults\n")
	}

	switch cfg.Backend {
	case hsmlib.BackendLedger:
		return printLedgerStatus(cmd, cfg)
	case hsmlib.BackendPKCS11, hsmlib.BackendSoftHSM:
		return printPKCS11Status(cmd, cfg)
	default:
		return fmt.Errorf("status probe not implemented for backend %q", cfg.Backend)
	}
}

func printLedgerStatus(cmd *cobra.Command, cfg *hsmlib.Config) error {
	ledgerCfg := vecrypto.DefaultLedgerWalletConfig()
	if cfg.ConnectionTimeout > 0 {
		ledgerCfg.Timeout = cfg.ConnectionTimeout
	}
	if cfg.MaxRetries > 0 {
		ledgerCfg.RetryCount = cfg.MaxRetries
	}
	if cfg.Ledger != nil {
		if cfg.Ledger.DerivationPath != "" {
			ledgerCfg.DefaultHDPath = cfg.Ledger.DerivationPath
		}
		if cfg.Ledger.HRP != "" {
			ledgerCfg.HRPPrefix = cfg.Ledger.HRP
		}
	}

	client := newLedgerStatusClient(ledgerCfg)
	ctx, cancel := context.WithTimeout(cmd.Context(), ledgerCfg.Timeout)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("ledger probe failed: %w", err)
	}
	defer func() { _ = client.Disconnect() }()

	info, err := client.GetDeviceInfo(ctx)
	if err != nil {
		return fmt.Errorf("ledger device info failed: %w", err)
	}

	address, err := client.GetAddress(ctx, ledgerCfg.DefaultHDPath, false)
	if err != nil {
		return fmt.Errorf("ledger address derivation failed: %w", err)
	}

	publicKey := address.PublicKey
	if len(publicKey) == 0 {
		publicKey, err = client.GetPublicKey(ctx, ledgerCfg.DefaultHDPath)
		if err != nil {
			return fmt.Errorf("ledger public key lookup failed: %w", err)
		}
	}

	cmd.Printf("  Connected:   yes\n")
	cmd.Printf("  Device:      %s\n", info.DeviceType)
	cmd.Printf("  Transport:   %s\n", info.ConnectionType)
	cmd.Printf("  App:         %s %s\n", info.AppName, info.AppVersion)
	if info.SerialNumber != "" {
		cmd.Printf("  Serial:      %s\n", info.SerialNumber)
	}
	if info.FirmwareVersion != "" {
		cmd.Printf("  Firmware:    %s\n", info.FirmwareVersion)
	}
	cmd.Printf("  Path:        %s\n", ledgerCfg.DefaultHDPath)
	cmd.Printf("  Address:     %s\n", address.Address)
	cmd.Printf("  Public Key:  %s\n", hex.EncodeToString(publicKey))
	cmd.Printf("  Keys:        1\n")

	return nil
}

func printPKCS11Status(cmd *cobra.Command, cfg *hsmlib.Config) error {
	if cfg.PKCS11 == nil {
		return errors.New("pkcs11 config is required")
	}

	if _, err := statusFileStat(cfg.PKCS11.LibraryPath); err != nil {
		return fmt.Errorf("PKCS#11 library unavailable at %s: %w", cfg.PKCS11.LibraryPath, err)
	}

	client, err := newPKCS11StatusClient(*cfg.PKCS11)
	if err != nil {
		return fmt.Errorf("PKCS#11 probe setup failed: %w", err)
	}
	defer func() { _ = client.Close() }()

	timeout := cfg.ConnectionTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("PKCS#11 connection failed: %w", err)
	}

	keys, err := client.ListKeys(ctx)
	if err != nil {
		return fmt.Errorf("PKCS#11 key listing failed: %w", err)
	}

	cmd.Printf("  Connected:   yes\n")
	cmd.Printf("  Library:     %s\n", cfg.PKCS11.LibraryPath)
	cmd.Printf("  Slot:        %d\n", cfg.PKCS11.SlotID)
	if cfg.Backend == hsmlib.BackendSoftHSM {
		cmd.Printf("  Mode:        software token\n")
	}
	cmd.Printf("  Keys:        %d\n", len(keys))

	return nil
}

func readStatusConfig(configPath string) (*hsmlib.Config, string, bool, error) {
	resolvedPath, err := resolveStatusConfigPath(configPath)
	if err != nil {
		return nil, "", false, err
	}

	cfg := hsmlib.DefaultConfig()
	data, err := os.ReadFile(resolvedPath)
	if errors.Is(err, os.ErrNotExist) {
		return &cfg, resolvedPath, false, nil
	}
	if err != nil {
		return nil, "", false, fmt.Errorf("cannot read HSM config %s: %w", resolvedPath, err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, "", false, fmt.Errorf("cannot parse HSM config %s: %w", resolvedPath, err)
	}

	return &cfg, resolvedPath, true, nil
}

func resolveStatusConfigPath(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	return filepath.Join(home, ".virtengine", "hsm.json"), nil
}
