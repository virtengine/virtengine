package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

func TestConsentRecordLifecycle(t *testing.T) {
	ts := setupWalletTest(t)

	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	scopeID := "scope-consent-record"
	grantUpdate := types.ConsentUpdateRequest{
		ScopeID:      scopeID,
		GrantConsent: true,
		Purpose:      string(types.PurposeBiometricProcessing),
	}
	grantMsg := []byte("VEID_CONSENT_UPDATE:" + ts.address.String() + ":" + scopeID + ":grant")
	grantSig := ts.signMessage(grantMsg)
	err = ts.keeper.UpdateConsent(ts.ctx, ts.address, grantUpdate, grantSig)
	require.NoError(t, err)

	record, found := ts.keeper.GetConsentRecordBySubjectScope(ts.ctx, ts.address, scopeID)
	require.True(t, found)
	require.Equal(t, types.ConsentStatusActive, record.Status)
	require.Equal(t, types.PurposeBiometricProcessing, record.Purpose)
	require.NotEmpty(t, record.ConsentHash)
	require.NotEmpty(t, record.SignatureHash)

	events := ts.keeper.GetConsentEventsBySubject(ts.ctx, ts.address)
	require.Len(t, events, 1)
	require.Equal(t, types.ConsentEventGranted, events[0].EventType)

	revokeUpdate := types.ConsentUpdateRequest{
		ScopeID:      scopeID,
		GrantConsent: false,
	}
	revokeMsg := []byte("VEID_CONSENT_UPDATE:" + ts.address.String() + ":" + scopeID + ":revoke")
	revokeSig := ts.signMessage(revokeMsg)
	err = ts.keeper.UpdateConsent(ts.ctx, ts.address, revokeUpdate, revokeSig)
	require.NoError(t, err)

	record, found = ts.keeper.GetConsentRecordBySubjectScope(ts.ctx, ts.address, scopeID)
	require.True(t, found)
	require.Equal(t, types.ConsentStatusWithdrawn, record.Status)

	events = ts.keeper.GetConsentEventsBySubject(ts.ctx, ts.address)
	require.Len(t, events, 2)
	require.Equal(t, types.ConsentEventRevoked, events[1].EventType)
}

func TestAddScopeRequiresConsent(t *testing.T) {
	ts := setupWalletTest(t)

	walletID := GenerateWalletID(ts.address.String())
	bindingSignature := ts.signWalletBinding(walletID)
	_, err := ts.keeper.CreateWallet(ts.ctx, ts.address, bindingSignature, ts.pubKey)
	require.NoError(t, err)

	scopeID := "scope-sensitive"
	envelopeHash := sha256.Sum256([]byte("sensitive data"))
	scopeRef := types.ScopeReference{
		ScopeID:      scopeID,
		ScopeType:    types.ScopeTypeIDDocument,
		EnvelopeHash: envelopeHash[:],
		AddedAt:      time.Now(),
		Status:       types.ScopeRefStatusPending,
	}

	addScopeMsg := types.GetAddScopeSigningMessage(ts.address.String(), scopeID)
	addScopeSig := ts.signMessage(addScopeMsg)
	err = ts.keeper.AddScopeToWallet(ts.ctx, ts.address, scopeRef, addScopeSig)
	require.Error(t, err)
	require.ErrorContains(t, err, "consent")

	grantUpdate := types.ConsentUpdateRequest{
		ScopeID:      scopeID,
		GrantConsent: true,
		Purpose:      "KYC verification",
	}
	grantMsg := []byte("VEID_CONSENT_UPDATE:" + ts.address.String() + ":" + scopeID + ":grant")
	grantSig := ts.signMessage(grantMsg)
	err = ts.keeper.UpdateConsent(ts.ctx, ts.address, grantUpdate, grantSig)
	require.NoError(t, err)

	err = ts.keeper.AddScopeToWallet(ts.ctx, ts.address, scopeRef, addScopeSig)
	require.NoError(t, err)
}
