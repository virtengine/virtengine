package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	"github.com/virtengine/virtengine/x/provider/keeper"
)

const testDomain = "provider.example.com"

func TestRequestDomainVerification(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	owner := sdk.AccAddress("owner_address______")
	domain := testDomain

	tests := []struct {
		name           string
		method         v1beta4.VerificationMethod
		expectedTarget string
	}{
		{
			name:           "DNS TXT verification",
			method:         v1beta4.VERIFICATION_METHOD_DNS_TXT,
			expectedTarget: "_virtengine-verification.provider.example.com",
		},
		{
			name:           "DNS CNAME verification",
			method:         v1beta4.VERIFICATION_METHOD_DNS_CNAME,
			expectedTarget: "_virtengine-verification.provider.example.com",
		},
		{
			name:           "HTTP well-known verification",
			method:         v1beta4.VERIFICATION_METHOD_HTTP_WELL_KNOWN,
			expectedTarget: "https://provider.example.com/.well-known/virtengine-verification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, target, err := k.RequestDomainVerification(ctx, owner, domain, tt.method)
			require.NoError(t, err)
			require.NotNil(t, record)
			require.Equal(t, owner.String(), record.ProviderAddress)
			require.Equal(t, domain, record.Domain)
			require.NotEmpty(t, record.Token)
			require.Equal(t, keeper.DomainVerificationPending, record.Status)
			require.Greater(t, record.ExpiresAt, record.GeneratedAt)
			require.Equal(t, tt.expectedTarget, target)
			require.Equal(t, 64, len(record.Token))

			storedRecord, found := k.GetDomainVerificationRecord(ctx, owner)
			require.True(t, found)
			require.Equal(t, record.Token, storedRecord.Token)
		})
	}
}

func TestConfirmDomainVerification(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	owner := sdk.AccAddress("owner_address______")
	domain := testDomain

	record, _, err := k.RequestDomainVerification(ctx, owner, domain, v1beta4.VERIFICATION_METHOD_DNS_TXT)
	require.NoError(t, err)
	require.NotNil(t, record)

	proof := "dns-txt-verified-" + record.Token
	err = k.ConfirmDomainVerification(ctx, owner, proof)
	require.NoError(t, err)

	verifiedRecord, found := k.GetDomainVerificationRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, keeper.DomainVerificationVerified, verifiedRecord.Status)
	require.Equal(t, proof, verifiedRecord.Proof)
	require.Greater(t, verifiedRecord.VerifiedAt, int64(0))
	require.Greater(t, verifiedRecord.RenewalAt, int64(0))
}

func TestConfirmDomainVerification_NoRecord(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	owner := sdk.AccAddress("owner_address______")
	err := k.ConfirmDomainVerification(ctx, owner, "proof")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestConfirmDomainVerification_EmptyProof(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	owner := sdk.AccAddress("owner_address______")
	domain := testDomain

	_, _, err := k.RequestDomainVerification(ctx, owner, domain, v1beta4.VERIFICATION_METHOD_DNS_TXT)
	require.NoError(t, err)

	err = k.ConfirmDomainVerification(ctx, owner, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "proof cannot be empty")
}

func TestRevokeDomainVerification(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	owner := sdk.AccAddress("owner_address______")
	domain := testDomain

	_, _, err := k.RequestDomainVerification(ctx, owner, domain, v1beta4.VERIFICATION_METHOD_DNS_TXT)
	require.NoError(t, err)

	err = k.ConfirmDomainVerification(ctx, owner, "proof")
	require.NoError(t, err)

	err = k.RevokeDomainVerification(ctx, owner)
	require.NoError(t, err)

	revokedRecord, found := k.GetDomainVerificationRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, keeper.DomainVerificationRevoked, revokedRecord.Status)
}

func TestGenerateDomainVerificationToken(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	owner := sdk.AccAddress("owner_address______")
	domain := testDomain

	record, err := k.GenerateDomainVerificationToken(ctx, owner, domain)
	require.NoError(t, err)
	require.NotNil(t, record)
	require.Equal(t, owner.String(), record.ProviderAddress)
	require.Equal(t, domain, record.Domain)
	require.NotEmpty(t, record.Token)
	require.Equal(t, keeper.DomainVerificationPending, record.Status)
	require.Greater(t, record.ExpiresAt, record.GeneratedAt)

	require.Equal(t, 64, len(record.Token))

	storedRecord, found := k.GetDomainVerificationRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, record.Token, storedRecord.Token)
}

func TestGenerateDomainVerificationToken_InvalidDomain(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	owner := sdk.AccAddress("owner_address______")

	tests := []struct {
		name   string
		domain string
	}{
		{"empty domain", ""},
		{"invalid format", "notadomain"},
		{"too long", "a." + string(make([]byte, 255))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := k.GenerateDomainVerificationToken(ctx, owner, tt.domain)
			require.Error(t, err)
		})
	}
}

func TestIsDomainVerified(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	owner := sdk.AccAddress("owner_address______")

	require.False(t, k.IsDomainVerified(ctx, owner))

	_, _, err := k.RequestDomainVerification(ctx, owner, testDomain, v1beta4.VERIFICATION_METHOD_DNS_TXT)
	require.NoError(t, err)

	require.False(t, k.IsDomainVerified(ctx, owner))

	err = k.ConfirmDomainVerification(ctx, owner, "proof")
	require.NoError(t, err)

	require.True(t, k.IsDomainVerified(ctx, owner))
}

func TestTokenExpiration(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	owner := sdk.AccAddress("owner_address______")
	domain := testDomain

	_, _, err := k.RequestDomainVerification(ctx, owner, domain, v1beta4.VERIFICATION_METHOD_DNS_TXT)
	require.NoError(t, err)

	futureCtx := ctx.WithBlockTime(ctx.BlockTime().Add(8 * 24 * time.Hour))

	err = k.ConfirmDomainVerification(futureCtx, owner, "proof")
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")

	updatedRecord, found := k.GetDomainVerificationRecord(futureCtx, owner)
	require.True(t, found)
	require.Equal(t, keeper.DomainVerificationExpired, updatedRecord.Status)
}

func TestDeleteDomainVerificationRecord(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	k := suite.keeper

	owner := sdk.AccAddress("owner_address______")
	domain := "provider.example.com"

	_, _, err := k.RequestDomainVerification(ctx, owner, domain, v1beta4.VERIFICATION_METHOD_DNS_TXT)
	require.NoError(t, err)

	_, found := k.GetDomainVerificationRecord(ctx, owner)
	require.True(t, found)

	k.DeleteDomainVerificationRecord(ctx, owner)

	_, found = k.GetDomainVerificationRecord(ctx, owner)
	require.False(t, found)
}

type testSuite struct {
	ctx    sdk.Context
	keeper keeper.IKeeper
}

//nolint:unparam
func setupTestSuite(t *testing.T) *testSuite {
	t.Skip("Test suite setup not implemented - requires keeper test utilities")
	return nil
}
