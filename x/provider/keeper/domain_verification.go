package keeper

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
)

// DomainVerificationStatus represents the verification state of a provider's domain
type DomainVerificationStatus string

const (
	DomainVerificationPending  DomainVerificationStatus = "pending"
	DomainVerificationVerified DomainVerificationStatus = "verified"
	DomainVerificationFailed   DomainVerificationStatus = "failed"
	DomainVerificationExpired  DomainVerificationStatus = "expired"
	DomainVerificationRevoked  DomainVerificationStatus = "revoked"

	// VerificationTokenLength is the length of the verification token in bytes
	VerificationTokenLength = 32

	// TokenExpirationDays is the number of days before a verification token expires
	TokenExpirationDays = 7

	// VerificationRenewalDays is the number of days before expiry to allow renewal
	VerificationRenewalDays = 30

	// DNSVerificationPrefix is the subdomain prefix for verification TXT records
	DNSVerificationPrefix = "_virtengine-verification"

	// HTTPWellKnownPath is the path for HTTP well-known verification
	// #nosec G101 -- This is a well-known URL path constant, not a credential
	HTTPWellKnownPath = "/.well-known/virtengine-verification"
)

// VerificationMethodType represents the method used for domain verification
type VerificationMethodType string

const (
	VerificationMethodUnknown       VerificationMethodType = "unknown"
	VerificationMethodDNSTXT        VerificationMethodType = "dns_txt"
	VerificationMethodDNSCNAME      VerificationMethodType = "dns_cname"
	VerificationMethodHTTPWellKnown VerificationMethodType = "http_well_known"
)

// DomainVerificationRecord stores domain verification data for a provider
type DomainVerificationRecord struct {
	ProviderAddress string                   `json:"provider_address"`
	Domain          string                   `json:"domain"`
	Token           string                   `json:"token"`
	Method          VerificationMethodType   `json:"method"`
	Proof           string                   `json:"proof,omitempty"`
	Status          DomainVerificationStatus `json:"status"`
	GeneratedAt     int64                    `json:"generated_at"`
	VerifiedAt      int64                    `json:"verified_at,omitempty"`
	ExpiresAt       int64                    `json:"expires_at"`
	RenewalAt       int64                    `json:"renewal_at,omitempty"`
}

// RequestDomainVerification requests domain verification with specified method (replaces GenerateDomainVerificationToken)
// TODO: Replace int32 with types.VerificationMethod after proto generation
func (k Keeper) RequestDomainVerification(ctx sdk.Context, providerAddr sdk.AccAddress, domain string, method int32) (*DomainVerificationRecord, string, error) {
	if err := validateDomain(domain); err != nil {
		return nil, "", types.ErrInvalidDomain.Wrapf("invalid domain: %v", err)
	}

	var methodType VerificationMethodType
	var verificationTarget string

	// Temporary enum values until proto generation
	const (
		VERIFICATION_METHOD_DNS_TXT         = 1
		VERIFICATION_METHOD_DNS_CNAME       = 2
		VERIFICATION_METHOD_HTTP_WELL_KNOWN = 3
	)

	switch method {
	case VERIFICATION_METHOD_DNS_TXT:
		methodType = VerificationMethodDNSTXT
		verificationTarget = fmt.Sprintf("%s.%s", DNSVerificationPrefix, domain)
	case VERIFICATION_METHOD_DNS_CNAME:
		methodType = VerificationMethodDNSCNAME
		verificationTarget = fmt.Sprintf("%s.%s", DNSVerificationPrefix, domain)
	case VERIFICATION_METHOD_HTTP_WELL_KNOWN:
		methodType = VerificationMethodHTTPWellKnown
		verificationTarget = fmt.Sprintf("https://%s%s", domain, HTTPWellKnownPath)
	default:
		return nil, "", types.ErrInvalidDomain.Wrap("unsupported verification method")
	}

	tokenBytes := make([]byte, VerificationTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, "", types.ErrInternal.Wrapf("failed to generate token: %v", err)
	}
	token := hex.EncodeToString(tokenBytes)

	now := ctx.BlockTime().Unix()
	expiresAt := ctx.BlockTime().Add(TokenExpirationDays * 24 * time.Hour).Unix()

	record := &DomainVerificationRecord{
		ProviderAddress: providerAddr.String(),
		Domain:          domain,
		Token:           token,
		Method:          methodType,
		Status:          DomainVerificationPending,
		GeneratedAt:     now,
		ExpiresAt:       expiresAt,
	}

	if err := k.setDomainVerificationRecord(ctx, record); err != nil {
		return nil, "", err
	}

	return record, verificationTarget, nil
}

// ConfirmDomainVerification confirms domain verification with off-chain proof
func (k Keeper) ConfirmDomainVerification(ctx sdk.Context, providerAddr sdk.AccAddress, proof string) error {
	record, found := k.GetDomainVerificationRecord(ctx, providerAddr)
	if !found {
		return types.ErrDomainVerificationNotFound.Wrapf("no verification record for provider: %s", providerAddr.String())
	}

	if ctx.BlockTime().Unix() > record.ExpiresAt {
		record.Status = DomainVerificationExpired
		_ = k.setDomainVerificationRecord(ctx, record)
		return types.ErrDomainVerificationExpired.Wrap("verification token has expired")
	}

	if record.Status == DomainVerificationVerified {
		return types.ErrDomainVerificationFailed.Wrap("domain already verified")
	}

	if proof == "" {
		return types.ErrDomainVerificationFailed.Wrap("proof cannot be empty")
	}

	record.Status = DomainVerificationVerified
	record.VerifiedAt = ctx.BlockTime().Unix()
	record.Proof = proof
	record.RenewalAt = ctx.BlockTime().Add((TokenExpirationDays - VerificationRenewalDays) * 24 * time.Hour).Unix()

	if err := k.setDomainVerificationRecord(ctx, record); err != nil {
		return err
	}

	return nil
}

// RevokeDomainVerification revokes a provider's domain verification
func (k Keeper) RevokeDomainVerification(ctx sdk.Context, providerAddr sdk.AccAddress) error {
	record, found := k.GetDomainVerificationRecord(ctx, providerAddr)
	if !found {
		return types.ErrDomainVerificationNotFound.Wrapf("no verification record for provider: %s", providerAddr.String())
	}

	record.Status = DomainVerificationRevoked
	if err := k.setDomainVerificationRecord(ctx, record); err != nil {
		return err
	}

	return nil
}

// GenerateDomainVerificationToken generates a new verification token for a provider's domain (legacy - kept for compatibility)
func (k Keeper) GenerateDomainVerificationToken(ctx sdk.Context, providerAddr sdk.AccAddress, domain string) (*DomainVerificationRecord, error) {
	// Use DNS_TXT as default method (value 1)
	record, _, err := k.RequestDomainVerification(ctx, providerAddr, domain, 1)
	return record, err
}

// VerifyProviderDomain verifies a provider's domain (legacy - DNS check removed, kept for compatibility)
func (k Keeper) VerifyProviderDomain(ctx sdk.Context, providerAddr sdk.AccAddress) error {
	record, found := k.GetDomainVerificationRecord(ctx, providerAddr)
	if !found {
		return types.ErrDomainVerificationNotFound.Wrapf("no verification record for provider: %s", providerAddr.String())
	}

	if ctx.BlockTime().Unix() > record.ExpiresAt {
		record.Status = DomainVerificationExpired
		_ = k.setDomainVerificationRecord(ctx, record)
		return types.ErrDomainVerificationExpired.Wrap("verification token has expired")
	}

	if record.Status == DomainVerificationVerified {
		_ = ctx.EventManager().EmitTypedEvent(
			&types.EventProviderDomainVerified{
				Owner:  record.ProviderAddress,
				Domain: record.Domain,
			},
		)
		return nil
	}

	return types.ErrDomainVerificationFailed.Wrap("domain not verified - use ConfirmDomainVerification with off-chain proof")
}

// GetDomainVerificationRecord retrieves the domain verification record for a provider
func (k Keeper) GetDomainVerificationRecord(ctx sdk.Context, providerAddr sdk.AccAddress) (*DomainVerificationRecord, bool) {
	store := ctx.KVStore(k.skey)
	key := DomainVerificationKey(providerAddr)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var record DomainVerificationRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return nil, false
	}

	return &record, true
}

// setDomainVerificationRecord stores a domain verification record
func (k Keeper) setDomainVerificationRecord(ctx sdk.Context, record *DomainVerificationRecord) error {
	providerAddr, err := sdk.AccAddressFromBech32(record.ProviderAddress)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.skey)
	key := DomainVerificationKey(providerAddr)

	bz, err := json.Marshal(record)
	if err != nil {
		return types.ErrInternal.Wrapf("failed to marshal verification record: %v", err)
	}

	store.Set(key, bz)
	return nil
}

// DeleteDomainVerificationRecord removes a domain verification record
func (k Keeper) DeleteDomainVerificationRecord(ctx sdk.Context, providerAddr sdk.AccAddress) {
	store := ctx.KVStore(k.skey)
	key := DomainVerificationKey(providerAddr)
	store.Delete(key)
}

// IsDomainVerified checks if a provider's domain is verified
func (k Keeper) IsDomainVerified(ctx sdk.Context, providerAddr sdk.AccAddress) bool {
	record, found := k.GetDomainVerificationRecord(ctx, providerAddr)
	if !found {
		return false
	}

	// Check if verified and not expired
	return record.Status == DomainVerificationVerified &&
		ctx.BlockTime().Unix() <= record.ExpiresAt
}

// validateDomain performs basic domain validation
func validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Basic length check
	if len(domain) > 253 {
		return fmt.Errorf("domain too long")
	}

	// Check for valid characters and structure
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return fmt.Errorf("domain must have at least two parts")
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return fmt.Errorf("invalid domain part length")
		}

		// Each part must start and end with alphanumeric
		if !isAlphanumeric(part[0]) || !isAlphanumeric(part[len(part)-1]) {
			return fmt.Errorf("domain parts must start and end with alphanumeric characters")
		}
	}

	return nil
}

func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
