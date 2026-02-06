// Package testutil provides test helpers for HSM integration tests.
package testutil

import (
	"log/slog"
	"os"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
	"github.com/virtengine/virtengine/pkg/keymanagement/hsm/pkcs11"
)

// NewSoftHSMProvider creates a PKCS#11 provider configured for SoftHSM2
// testing. It uses the software fallback so no real PKCS#11 library is needed.
func NewSoftHSMProvider() (*pkcs11.Provider, error) {
	config := hsm.PKCS11Config{
		LibraryPath: "/usr/lib/softhsm/libsofthsm2.so",
		SlotID:      0,
		TokenLabel:  "test-token",
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return pkcs11.New(config, logger)
}

// NewTestManager creates an HSM Manager with a SoftHSM provider for testing.
func NewTestManager() (*hsm.Manager, *pkcs11.Provider, error) {
	provider, err := NewSoftHSMProvider()
	if err != nil {
		return nil, nil, err
	}

	config := hsm.DefaultConfig()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mgr, err := hsm.NewManager(config, logger)
	if err != nil {
		return nil, nil, err
	}

	mgr.SetProvider(provider)
	return mgr, provider, nil
}
