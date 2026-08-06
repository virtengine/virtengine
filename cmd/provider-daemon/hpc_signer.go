package main

import (
	"encoding/hex"
	"fmt"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
)

type hpcKeyManagerSigner struct {
	keyManager      *provider_daemon.KeyManager
	providerAddress string
}

func newHPCKeyManagerSigner(keyManager *provider_daemon.KeyManager, providerAddress string) provider_daemon.HPCSchedulerSigner {
	return &hpcKeyManagerSigner{
		keyManager:      keyManager,
		providerAddress: providerAddress,
	}
}

func (s *hpcKeyManagerSigner) Sign(data []byte) ([]byte, error) {
	sig, err := s.keyManager.Sign(data)
	if err != nil {
		return nil, err
	}

	decoded, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return nil, fmt.Errorf("failed to decode provider signature: %w", err)
	}

	return decoded, nil
}

func (s *hpcKeyManagerSigner) GetProviderAddress() string {
	return s.providerAddress
}
