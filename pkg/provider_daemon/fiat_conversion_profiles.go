// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package provider_daemon

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
)

const fiatProfileFileSchemaVersion uint32 = 1

var (
	ErrFiatProfileInvalid  = errors.New("fiat conversion profile invalid")
	ErrFiatProfileMismatch = errors.New("fiat conversion profile commitment mismatch")
)

// VersionedDEXProfileFile is the only accepted local DEX profile file shape.
// The profile digest is SHA-256 over canonical json.Marshal(profile) bytes;
// certified rows additionally require an independent authority signature.
type VersionedDEXProfileFile struct {
	SchemaVersion uint32              `json:"schema_version"`
	Profile       dex.DEXRouteProfile `json:"profile"`
	AuthorityID   string              `json:"authority_id,omitempty"`
	Signature     string              `json:"signature,omitempty"`
}

// VersionedPayoutProfileFile is the only accepted local payout profile shape.
type VersionedPayoutProfileFile struct {
	SchemaVersion uint32                `json:"schema_version"`
	Profile       offramp.PayoutProfile `json:"profile"`
	AuthorityID   string                `json:"authority_id,omitempty"`
	Signature     string                `json:"signature,omitempty"`
}

// TrustedFiatProfiles contains immutable local profile snapshots and their
// canonical digests. It also implements both package trust-authority seams.
type TrustedFiatProfiles struct {
	DEX           dex.DEXRouteProfile
	Payout        offramp.PayoutProfile
	DEXDigest     [sha256.Size]byte
	PayoutDigest  [sha256.Size]byte
	DEXTrusted    bool
	PayoutTrusted bool
}

// FiatProfileAuthority verifies deployment profile signatures against a trust
// root configured independently from the profile files themselves.
type FiatProfileAuthority interface {
	AuthorizeFiatProfile(kind, authorityID string, schemaVersion uint32, digest [sha256.Size]byte, signature string) error
}

// Ed25519FiatProfileAuthority is a file-independent public-key trust root.
type Ed25519FiatProfileAuthority struct {
	AuthorityID string
	PublicKey   ed25519.PublicKey
}

func NewEd25519FiatProfileAuthority(authorityID string, publicKey []byte) (*Ed25519FiatProfileAuthority, error) {
	if strings.TrimSpace(authorityID) == "" || len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("%w: invalid profile authority", ErrFiatProfileInvalid)
	}
	return &Ed25519FiatProfileAuthority{AuthorityID: authorityID, PublicKey: append(ed25519.PublicKey(nil), publicKey...)}, nil
}

func (a *Ed25519FiatProfileAuthority) AuthorizeFiatProfile(kind, authorityID string, schemaVersion uint32, digest [sha256.Size]byte, signature string) error {
	if a == nil || authorityID != a.AuthorityID || (kind != "dex" && kind != "payout") {
		return ErrFiatProfileMismatch
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(signature)
	if err != nil || len(decoded) != ed25519.SignatureSize {
		return ErrFiatProfileMismatch
	}
	payload := fiatProfileAuthorizationPayload(kind, authorityID, schemaVersion, digest)
	if !ed25519.Verify(a.PublicKey, payload, decoded) {
		return ErrFiatProfileMismatch
	}
	return nil
}

// LoadTrustedFiatProfiles strictly loads versioned profile JSON. Unknown fields
// and trailing JSON are rejected so the on-chain digest cannot hide ignored data.
func LoadTrustedFiatProfiles(dexPath, payoutPath string) (*TrustedFiatProfiles, error) {
	return LoadTrustedFiatProfilesWithAuthority(dexPath, payoutPath, nil)
}

// LoadTrustedFiatProfilesWithAuthority requires an independent trust root for
// every certified production row. Unsigned external-blocked engineering rows
// may be loaded for conformance, but can never authorize production execution.
func LoadTrustedFiatProfilesWithAuthority(dexPath, payoutPath string, authority FiatProfileAuthority) (*TrustedFiatProfiles, error) {
	if strings.TrimSpace(dexPath) == "" || strings.TrimSpace(payoutPath) == "" {
		return nil, fmt.Errorf("%w: DEX and payout profile paths are required", ErrFiatProfileInvalid)
	}
	var dexFile VersionedDEXProfileFile
	if err := readStrictProfileJSON(dexPath, &dexFile); err != nil {
		return nil, fmt.Errorf("load DEX profile: %w", err)
	}
	var payoutFile VersionedPayoutProfileFile
	if err := readStrictProfileJSON(payoutPath, &payoutFile); err != nil {
		return nil, fmt.Errorf("load payout profile: %w", err)
	}
	if dexFile.SchemaVersion != fiatProfileFileSchemaVersion || payoutFile.SchemaVersion != fiatProfileFileSchemaVersion {
		return nil, fmt.Errorf("%w: unsupported profile file schema", ErrFiatProfileInvalid)
	}
	dexValidationMode := dex.RouteValidationEngineering
	if dexFile.Profile.State == dex.RouteCertifiedEnabled {
		dexValidationMode = dex.RouteValidationRuntime
	}
	if err := dexFile.Profile.Validate(dexValidationMode); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFiatProfileInvalid, err)
	}
	if err := payoutFile.Profile.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFiatProfileInvalid, err)
	}
	dexDigest, err := canonicalProfileDigest(dexFile.Profile)
	if err != nil {
		return nil, err
	}
	payoutDigest, err := canonicalProfileDigest(payoutFile.Profile)
	if err != nil {
		return nil, err
	}
	dexTrusted := false
	if dexFile.Profile.State == dex.RouteCertifiedEnabled {
		if authority == nil || authority.AuthorizeFiatProfile("dex", dexFile.AuthorityID, dexFile.SchemaVersion, dexDigest, dexFile.Signature) != nil {
			return nil, fmt.Errorf("%w: certified DEX profile lacks trusted authority", ErrFiatProfileInvalid)
		}
		dexTrusted = true
	}
	payoutTrusted := false
	if payoutFile.Profile.State == offramp.ProfileCertifiedEnabled {
		if authority == nil || authority.AuthorizeFiatProfile("payout", payoutFile.AuthorityID, payoutFile.SchemaVersion, payoutDigest, payoutFile.Signature) != nil {
			return nil, fmt.Errorf("%w: certified payout profile lacks trusted authority", ErrFiatProfileInvalid)
		}
		payoutTrusted = true
	}
	return &TrustedFiatProfiles{
		DEX: dexFile.Profile, Payout: payoutFile.Profile,
		DEXDigest: dexDigest, PayoutDigest: payoutDigest, DEXTrusted: dexTrusted, PayoutTrusted: payoutTrusted,
	}, nil
}

func readStrictProfileJSON(path string, target any) error {
	file, err := os.Open(path) // #nosec G304 -- operator-supplied configuration path.
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, 8<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func canonicalProfileDigest(profile any) ([sha256.Size]byte, error) {
	encoded, err := json.Marshal(profile)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("canonical profile JSON: %w", err)
	}
	return sha256.Sum256(encoded), nil
}

// AuthorizeDEXRoute accepts only the exact loaded profile bytes.
func (p *TrustedFiatProfiles) AuthorizeDEXRoute(profile dex.DEXRouteProfile) error {
	if p == nil || !p.DEXTrusted || profile.ID != p.DEX.ID {
		return ErrFiatProfileMismatch
	}
	digest, err := canonicalProfileDigest(profile)
	if err != nil || !bytes.Equal(digest[:], p.DEXDigest[:]) {
		return ErrFiatProfileMismatch
	}
	return nil
}

// AuthorizePayoutProfile accepts only the exact loaded profile bytes.
func (p *TrustedFiatProfiles) AuthorizePayoutProfile(profile offramp.PayoutProfile) error {
	if p == nil || !p.PayoutTrusted || profile.ID != p.Payout.ID {
		return ErrFiatProfileMismatch
	}
	digest, err := canonicalProfileDigest(profile)
	if err != nil || !bytes.Equal(digest[:], p.PayoutDigest[:]) {
		return ErrFiatProfileMismatch
	}
	return nil
}

func (p *TrustedFiatProfiles) DEXDigestHex() string {
	if p == nil {
		return ""
	}
	return hex.EncodeToString(p.DEXDigest[:])
}

func (p *TrustedFiatProfiles) PayoutDigestHex() string {
	if p == nil {
		return ""
	}
	return hex.EncodeToString(p.PayoutDigest[:])
}

func fiatProfileAuthorizationPayload(kind, authorityID string, schemaVersion uint32, digest [sha256.Size]byte) []byte {
	result := make([]byte, 0, 128)
	for _, value := range [][]byte{[]byte("virtengine/provider-daemon/fiat-profile/v1"), []byte(kind), []byte(authorityID)} {
		length := make([]byte, 4)
		binary.BigEndian.PutUint32(length, uint32(len(value))) //nolint:gosec // bounded configuration fields.
		result = append(result, length...)
		result = append(result, value...)
	}
	version := make([]byte, 4)
	binary.BigEndian.PutUint32(version, schemaVersion)
	result = append(result, version...)
	result = append(result, digest[:]...)
	return result
}

var _ dex.RouteProfileAuthorizer = (*TrustedFiatProfiles)(nil)
var _ offramp.ProfileAuthorizer = (*TrustedFiatProfiles)(nil)
var _ FiatProfileAuthority = (*Ed25519FiatProfileAuthority)(nil)
