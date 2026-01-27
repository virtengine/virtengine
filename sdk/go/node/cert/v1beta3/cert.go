package v1beta3

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ParseAndValidateCertificate(owner sdk.Address, crt, pub []byte) (*x509.Certificate, error) {
	blk, rest := pem.Decode(pub)
	if blk == nil || len(rest) > 0 {
		return nil, ErrInvalidPubkeyValue
	}

	if blk.Type != PemBlkTypeECPublicKey {
		return nil, fmt.Errorf("%w: invalid pem block type", ErrInvalidPubkeyValue)
	}

	pkRes, err := x509.ParsePKIXPublicKey(blk.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidPubkeyValue, err)
	}

	pk, valid := pkRes.(*ecdsa.PublicKey)
	if !valid {
		return nil, fmt.Errorf("%w: invalid public key type", ErrInvalidPubkeyValue)
	}

	blk, rest = pem.Decode(crt)
	if blk == nil || len(rest) > 0 {
		return nil, ErrInvalidCertificateValue
	}

	if blk.Type != PemBlkTypeCertificate {
		return nil, fmt.Errorf("%w: invalid pem block type", ErrInvalidCertificateValue)
	}

	cert, err := x509.ParseCertificate(blk.Bytes)
	if err != nil {
		return nil, err
	}

	// verify signature
	hash := sha256.New()
	hash.Write(cert.RawTBSCertificate)

	if !ecdsa.VerifyASN1(pk, hash.Sum(nil), cert.Signature) {
		return nil, fmt.Errorf("%w: invalid certificate signature", ErrInvalidCertificateValue)
	}

	cowner, err := sdk.AccAddressFromBech32(cert.Subject.CommonName)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCertificateValue, err)
	}

	if !owner.Equals(cowner) {
		return nil, fmt.Errorf("%w: CommonName does not match owner", ErrInvalidCertificateValue)
	}

	return cert, nil
}

func (m *CertificateID) String() string {
	return fmt.Sprintf("%s/%s", m.Owner, m.Serial)
}

func (m *CertificateID) Equals(val CertificateID) bool {
	return (m.Owner == val.Owner) && (m.Serial == val.Serial)
}

func (m Certificate) Validate(owner sdk.Address) error {
	if val, exists := Certificate_State_name[int32(m.State)]; !exists || val == "invalid" {
		return ErrInvalidState
	}

	_, err := ParseAndValidateCertificate(owner, m.Cert, m.Pubkey)
	if err != nil {
		return err
	}

	return nil
}

func (m Certificate) IsState(state Certificate_State) bool {
	return m.State == state
}
