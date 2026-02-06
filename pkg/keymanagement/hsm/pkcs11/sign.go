package pkcs11

import (
	"context"
	"crypto"
	"io"

	"github.com/virtengine/virtengine/pkg/keymanagement/hsm"
)

// PKCS11Signer implements crypto.Signer by delegating to a PKCS#11 provider.
type PKCS11Signer struct {
	provider *Provider
	label    string
	info     *hsm.KeyInfo
}

// NewPKCS11Signer creates a signer for a specific key in the provider.
func NewPKCS11Signer(provider *Provider, label string, info *hsm.KeyInfo) *PKCS11Signer {
	return &PKCS11Signer{
		provider: provider,
		label:    label,
		info:     info,
	}
}

// Public returns the public key associated with the signer.
func (s *PKCS11Signer) Public() crypto.PublicKey {
	pk, err := s.provider.GetPublicKey(context.Background(), s.label)
	if err != nil {
		return nil
	}
	return pk
}

// Sign signs digest using the PKCS#11 provider.
func (s *PKCS11Signer) Sign(_ io.Reader, digest []byte, _ crypto.SignerOpts) ([]byte, error) {
	return s.provider.Sign(context.Background(), s.label, digest)
}

// Label returns the key label.
func (s *PKCS11Signer) Label() string { return s.label }

// KeyInfo returns metadata about the key.
func (s *PKCS11Signer) KeyInfo() *hsm.KeyInfo { return s.info }

// Ensure PKCS11Signer implements hsm.Signer.
var _ hsm.Signer = (*PKCS11Signer)(nil)
