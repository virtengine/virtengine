package fundauth

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const canonicalMagic = "VE-FUND-AUTH\x00"

const maxCanonicalTextLength = 4096

var (
	denomPattern       = regexp.MustCompile(`^[a-z][a-z0-9]{0,63}(?:[./_-][a-z0-9]+)*$`)
	canonicalIDPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:[._:-][a-z0-9]+)*$`)
)

type canonicalWriter struct {
	bytes.Buffer
}

func (writer *canonicalWriter) text(value, field string, required bool) error {
	if required && value == "" {
		return fmt.Errorf("%w: empty %s", ErrInvalidAuthorization, field)
	}
	if len(value) > maxCanonicalTextLength || len(value) > math.MaxUint32 {
		return fmt.Errorf("%w: %s too long", ErrInvalidAuthorization, field)
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%w: non-canonical %s", ErrInvalidAuthorization, field)
	}
	for _, character := range []byte(value) {
		if character < 0x21 || character > 0x7e {
			return fmt.Errorf("%w: invalid %s", ErrInvalidAuthorization, field)
		}
	}
	_ = binary.Write(&writer.Buffer, binary.BigEndian, uint32(len(value)))
	_, _ = writer.WriteString(value)
	return nil
}

func (writer *canonicalWriter) digest(value, field string, required, allowZero bool) error {
	digest, err := parseDigest(value, field, required, allowZero)
	if err != nil {
		return err
	}
	_, _ = writer.Write(digest[:])
	return nil
}

func parseRequiredDigest(value, field string) (Digest, error) {
	return parseDigest(value, field, true, false)
}

func parseDigest(value, field string, required, allowZero bool) (Digest, error) {
	var result Digest
	if value == "" && !required {
		return result, nil
	}
	if len(value) != hex.EncodedLen(len(result)) || strings.ToLower(value) != value {
		return result, fmt.Errorf("%w: %s must be lowercase SHA-256 hex", ErrInvalidAuthorization, field)
	}
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return result, fmt.Errorf("%w: invalid %s digest", ErrInvalidAuthorization, field)
	}
	copy(result[:], decoded)
	if !allowZero && result == (Digest{}) {
		return Digest{}, fmt.Errorf("%w: zero %s digest", ErrInvalidAuthorization, field)
	}
	return result, nil
}

func CanonicalSignBytes(auth FundAuthorization) ([]byte, Digest, error) {
	var writer canonicalWriter
	_, _ = writer.WriteString(canonicalMagic)
	if auth.Domain != AuthorizationDomain || auth.Version != AuthorizationVersion {
		return nil, Digest{}, fmt.Errorf("%w: domain or version", ErrInvalidAuthorization)
	}
	if err := writer.text(auth.Domain, "domain", true); err != nil {
		return nil, Digest{}, err
	}
	_ = binary.Write(&writer.Buffer, binary.BigEndian, auth.Version)
	for _, field := range []struct{ value, name string }{
		{auth.ChainID, "chain ID"}, {auth.AccountID, "account ID"}, {auth.SignerKeyID, "signer key ID"},
		{auth.SourceID, "source ID"}, {auth.TypeURL, "type URL"},
	} {
		if err := writer.text(field.value, field.name, field.name != "type URL"); err != nil {
			return nil, Digest{}, err
		}
	}
	if !canonicalIDPattern.MatchString(auth.AccountID) || !canonicalIDPattern.MatchString(auth.SignerKeyID) {
		return nil, Digest{}, fmt.Errorf("%w: non-canonical account or key ID", ErrInvalidAuthorization)
	}
	if auth.SignerKeyEpoch == 0 {
		return nil, Digest{}, fmt.Errorf("%w: zero signer key epoch", ErrInvalidAuthorization)
	}
	_ = binary.Write(&writer.Buffer, binary.BigEndian, auth.SignerKeyEpoch)
	if !auth.Phase.valid() || !auth.Effect.valid() {
		return nil, Digest{}, fmt.Errorf("%w: phase or effect", ErrInvalidAuthorization)
	}
	_ = writer.WriteByte(byte(auth.Phase))
	_ = writer.WriteByte(byte(auth.Effect))
	if err := writer.digest(auth.MessageDigestHex, "message", true, false); err != nil {
		return nil, Digest{}, err
	}
	if uint64(len(auth.Amounts)) > math.MaxUint32 {
		return nil, Digest{}, fmt.Errorf("%w: amount count", ErrInvalidAuthorization)
	}
	if len(auth.Amounts) == 0 && (auth.Phase != PhaseControl || auth.Effect != EffectRecoveryControl) {
		return nil, Digest{}, fmt.Errorf("%w: amounts required outside recovery controls", ErrInvalidAuthorization)
	}
	if len(auth.Amounts) != 0 && auth.Phase == PhaseControl && auth.Effect == EffectRecoveryControl {
		return nil, Digest{}, fmt.Errorf("%w: recovery control amounts", ErrInvalidAuthorization)
	}
	_ = binary.Write(&writer.Buffer, binary.BigEndian, uint32(len(auth.Amounts)))
	previousDenom := ""
	for _, amount := range auth.Amounts {
		if !denomPattern.MatchString(amount.Denom) || (previousDenom != "" && amount.Denom <= previousDenom) {
			return nil, Digest{}, fmt.Errorf("%w: non-canonical or duplicate amount denom", ErrInvalidAuthorization)
		}
		if amount.MinorUnits == "" || (len(amount.MinorUnits) > 1 && amount.MinorUnits[0] == '0') {
			return nil, Digest{}, fmt.Errorf("%w: non-canonical amount", ErrInvalidAuthorization)
		}
		value, err := strconv.ParseUint(amount.MinorUnits, 10, 64)
		if err != nil || value == 0 {
			return nil, Digest{}, fmt.Errorf("%w: amount overflow or syntax", ErrInvalidAuthorization)
		}
		if err := writer.text(amount.Denom, "denom", true); err != nil {
			return nil, Digest{}, err
		}
		_ = binary.Write(&writer.Buffer, binary.BigEndian, value)
		previousDenom = amount.Denom
	}
	if len(auth.Parties) == 0 || uint64(len(auth.Parties)) > math.MaxUint32 {
		return nil, Digest{}, fmt.Errorf("%w: party count", ErrInvalidAuthorization)
	}
	_ = binary.Write(&writer.Buffer, binary.BigEndian, uint32(len(auth.Parties)))
	var previous PartyBinding
	for index, party := range auth.Parties {
		if !party.Role.valid() {
			return nil, Digest{}, fmt.Errorf("%w: party role", ErrInvalidAuthorization)
		}
		if index > 0 && (party.Role < previous.Role || (party.Role == previous.Role && party.AccountID <= previous.AccountID)) {
			return nil, Digest{}, fmt.Errorf("%w: non-canonical or duplicate party", ErrInvalidAuthorization)
		}
		_ = writer.WriteByte(byte(party.Role))
		if err := writer.text(party.AccountID, "party account ID", true); err != nil {
			return nil, Digest{}, err
		}
		if !canonicalIDPattern.MatchString(party.AccountID) {
			return nil, Digest{}, fmt.Errorf("%w: non-canonical party account ID", ErrInvalidAuthorization)
		}
		previous = party
	}
	if auth.MFAMode != MFAPossessionOnlyPolicyApproved && auth.MFAMode != MFAEvidenceRequired {
		return nil, Digest{}, fmt.Errorf("%w: MFA mode", ErrInvalidAuthorization)
	}
	_ = writer.WriteByte(byte(auth.MFAMode))
	if auth.EligibilityMode != EligibilityNotRequired && auth.EligibilityMode != EligibilityEvidenceRequired {
		return nil, Digest{}, fmt.Errorf("%w: eligibility mode", ErrInvalidAuthorization)
	}
	_ = writer.WriteByte(byte(auth.EligibilityMode))
	for _, field := range []struct {
		value, name         string
		required, allowZero bool
	}{
		{auth.CaseDigestHex, "case", false, false}, {auth.OrderDigestHex, "order", false, false},
		{auth.ReferenceDigestHex, "reference", false, false}, {auth.MFADigestHex, "MFA", auth.MFAMode == MFAEvidenceRequired, false},
		{auth.EligibilityDigestHex, "eligibility", auth.EligibilityMode == EligibilityEvidenceRequired, false}, {auth.PolicyDigestHex, "policy", true, false},
		{auth.NonceDigestHex, "nonce", true, false},
	} {
		if err := writer.digest(field.value, field.name, field.required, field.allowZero); err != nil {
			return nil, Digest{}, err
		}
	}
	if auth.MFAMode == MFAPossessionOnlyPolicyApproved && auth.MFADigestHex != "" {
		return nil, Digest{}, fmt.Errorf("%w: possession-only MFA evidence", ErrInvalidAuthorization)
	}
	if auth.EligibilityMode == EligibilityNotRequired && auth.EligibilityDigestHex != "" {
		return nil, Digest{}, fmt.Errorf("%w: unexpected eligibility evidence", ErrInvalidAuthorization)
	}
	if auth.IssuedAtBlock == 0 || auth.IssuedAtUnix == 0 || auth.LowerBlock == 0 || auth.IssuedAtBlock != auth.LowerBlock || auth.UpperBlock < auth.LowerBlock || auth.Expiry.Value == 0 {
		return nil, Digest{}, fmt.Errorf("%w: bounds or expiry", ErrInvalidAuthorization)
	}
	_ = binary.Write(&writer.Buffer, binary.BigEndian, auth.IssuedAtBlock)
	_ = binary.Write(&writer.Buffer, binary.BigEndian, auth.IssuedAtUnix)
	_ = binary.Write(&writer.Buffer, binary.BigEndian, auth.LowerBlock)
	_ = binary.Write(&writer.Buffer, binary.BigEndian, auth.UpperBlock)
	if auth.Expiry.Kind != ExpiryAtBlock && auth.Expiry.Kind != ExpiryAtUnixTime {
		return nil, Digest{}, fmt.Errorf("%w: expiry kind", ErrInvalidAuthorization)
	}
	if auth.Expiry.Kind == ExpiryAtBlock && (auth.Expiry.Value < auth.LowerBlock || auth.Expiry.Value > auth.UpperBlock) {
		return nil, Digest{}, fmt.Errorf("%w: block expiry outside bounds", ErrInvalidAuthorization)
	}
	_ = writer.WriteByte(byte(auth.Expiry.Kind))
	_ = binary.Write(&writer.Buffer, binary.BigEndian, auth.Expiry.Value)
	result := append([]byte(nil), writer.Bytes()...)
	return result, sha256.Sum256(result), nil
}

func AuthorizationDigest(auth FundAuthorization) (Digest, error) {
	_, digest, err := CanonicalSignBytes(auth)
	return digest, err
}

func (role PartyRole) valid() bool { return role >= PartyRoleSender && role <= PartyRoleTreasury }
