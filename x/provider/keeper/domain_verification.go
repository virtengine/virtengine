package keeper

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
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

	// VerificationTokenLength is the length of the verification token in bytes
	VerificationTokenLength = 32

	// TokenExpirationDays is the number of days before a verification token expires
	TokenExpirationDays = 7

	// DNSVerificationPrefix is the subdomain prefix for verification TXT records
	DNSVerificationPrefix = "_virtengine-verification"
)

// DomainVerificationRecord stores domain verification data for a provider
type DomainVerificationRecord struct {
	ProviderAddress string                   `json:"provider_address"`
	Domain          string                   `json:"domain"`
	Token           string                   `json:"token"`
	Status          DomainVerificationStatus `json:"status"`
	GeneratedAt     int64                    `json:"generated_at"`
	VerifiedAt      int64                    `json:"verified_at,omitempty"`
	ExpiresAt       int64                    `json:"expires_at"`
}

// GenerateDomainVerificationToken generates a new verification token for a provider's domain
func (k Keeper) GenerateDomainVerificationToken(ctx sdk.Context, providerAddr sdk.AccAddress, domain string) (*DomainVerificationRecord, error) {
	// Validate domain format
	if err := validateDomain(domain); err != nil {
		return nil, types.ErrInvalidDomain.Wrapf("invalid domain: %v", err)
	}

	// Generate random token
	tokenBytes := make([]byte, VerificationTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, types.ErrInternal.Wrapf("failed to generate token: %v", err)
	}
	token := hex.EncodeToString(tokenBytes)

	now := ctx.BlockTime().Unix()
	expiresAt := ctx.BlockTime().Add(TokenExpirationDays * 24 * time.Hour).Unix()

	record := &DomainVerificationRecord{
		ProviderAddress: providerAddr.String(),
		Domain:          domain,
		Token:           token,
		Status:          DomainVerificationPending,
		GeneratedAt:     now,
		ExpiresAt:       expiresAt,
	}

	// Store the record
	if err := k.setDomainVerificationRecord(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// VerifyProviderDomain verifies a provider's domain by checking DNS TXT records
func (k Keeper) VerifyProviderDomain(ctx sdk.Context, providerAddr sdk.AccAddress) error {
	record, found := k.GetDomainVerificationRecord(ctx, providerAddr)
	if !found {
		return types.ErrDomainVerificationNotFound.Wrapf("no verification record for provider: %s", providerAddr.String())
	}

	// Check if token is expired
	if ctx.BlockTime().Unix() > record.ExpiresAt {
		record.Status = DomainVerificationExpired
		_ = k.setDomainVerificationRecord(ctx, record)
		return types.ErrDomainVerificationExpired.Wrap("verification token has expired")
	}

	// Perform DNS TXT record lookup
	expectedRecord := fmt.Sprintf("%s.%s", DNSVerificationPrefix, record.Domain)

	// Query DNS TXT records
	txtRecords, err := net.LookupTXT(expectedRecord)
	if err != nil {
		record.Status = DomainVerificationFailed
		_ = k.setDomainVerificationRecord(ctx, record)
		return types.ErrDomainVerificationFailed.Wrapf("DNS lookup failed: %v", err)
	}

	// Check if any TXT record matches the token
	found = false
	for _, txt := range txtRecords {
		if strings.TrimSpace(txt) == record.Token {
			found = true
			break
		}
	}

	if !found {
		record.Status = DomainVerificationFailed
		_ = k.setDomainVerificationRecord(ctx, record)
		return types.ErrDomainVerificationFailed.Wrap("verification token not found in DNS TXT records")
	}

	// Mark as verified
	record.Status = DomainVerificationVerified
	record.VerifiedAt = ctx.BlockTime().Unix()

	if err := k.setDomainVerificationRecord(ctx, record); err != nil {
		return err
	}

	// Emit verification success event
	_ = ctx.EventManager().EmitTypedEvent(
		&types.EventProviderDomainVerified{
			Owner:  record.ProviderAddress,
			Domain: record.Domain,
		},
	)

	return nil
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
	if err := k.cdc.UnmarshalJSON(bz, &record); err != nil {
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

	bz, err := k.cdc.MarshalJSON(record)
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
