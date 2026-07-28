package keeper_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

const (
	webEvidenceAuthority = "authority"
	webEvidenceChainID   = "ve-test-1"
)

type webEvidenceFixture struct {
	ctx            sdk.Context
	stateStore     store.CommitMultiStore
	keeper         keeper.Keeper
	account        sdk.AccAddress
	accountAddress string
	walletPriv     ed25519.PrivateKey
	issuerPriv     ed25519.PrivateKey
	issuer         types.AttestationIssuer
	signerKey      *types.SignerKeyInfo
}

func TestSubmitWebEvidenceProofsRequireIssuerAndAccountCrypto(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)
	require.NotEqual(t, f.walletPriv.Public().(ed25519.PublicKey), f.signerKey.PublicKey)

	srv := keeper.NewMsgServerImpl(f.keeper)

	ssoMsg := f.validSSOMsg(t, "sso-linkage-1")
	ssoResp, err := srv.SubmitSSOVerificationProof(f.ctx, ssoMsg)
	require.NoError(t, err)
	require.Equal(t, "sso-linkage-1", ssoResp.LinkageId)

	emailMsg := f.validEmailMsg(t, "email-proof-1", nil)
	emailResp, err := srv.SubmitEmailVerificationProof(f.ctx, emailMsg)
	require.NoError(t, err)
	require.Equal(t, "email-proof-1", emailResp.VerificationId)

	smsMsg := f.validSMSMsg(t, "sms-proof-1")
	smsResp, err := srv.SubmitSMSVerificationProof(f.ctx, smsMsg)
	require.NoError(t, err)
	require.Equal(t, "sms-proof-1", smsResp.VerificationId)

	socialMsg := f.validSocialMsg(t, "social-scope-1")
	socialResp, err := srv.SubmitSocialMediaScope(f.ctx, socialMsg)
	require.NoError(t, err)
	require.Equal(t, "social-scope-1", socialResp.ScopeId)

	_, found := f.keeper.GetSSOLinkage(f.ctx, "sso-linkage-1")
	require.True(t, found)
	_, found = f.keeper.GetEmailVerificationRecord(f.ctx, "email-proof-1")
	require.True(t, found)
	_, found = f.keeper.GetSMSVerificationRecord(f.ctx, "sms-proof-1")
	require.True(t, found)
	_, found = f.keeper.GetSocialMediaScope(f.ctx, "social-scope-1")
	require.True(t, found)

	score, _, scoreFound := f.keeper.GetScore(f.ctx, f.accountAddress)
	require.True(t, scoreFound)
	require.NotZero(t, score)
}

func TestRegisterSignerKeyRejectsUnauthorizedAndCollisions(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)

	updated := cloneSignerKey(f.signerKey)
	updated.State = types.SignerKeyStateRotating
	updated.SuccessorKeyID = updated.SignerID + ":2"
	require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, updated))
	byID, found := f.keeper.GetSignerKey(f.ctx, updated.KeyID)
	require.True(t, found)
	require.Equal(t, types.SignerKeyStateRotating, byID.State)
	byFingerprint, found := f.keeper.GetSignerKeyByFingerprint(f.ctx, updated.Fingerprint)
	require.True(t, found)
	require.Equal(t, updated.KeyID, byFingerprint.KeyID)

	unauthorized := cloneSignerKey(updated)
	unauthorized.State = types.SignerKeyStateActive
	require.Error(t, f.keeper.RegisterSignerKey(f.ctx, "not-authority", unauthorized))
	byID, found = f.keeper.GetSignerKey(f.ctx, updated.KeyID)
	require.True(t, found)
	require.Equal(t, types.SignerKeyStateRotating, byID.State)

	changedPubKey := cloneSignerKey(updated)
	newPub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	changedPubKey.PublicKey = newPub
	changedPubKey.Fingerprint = types.ComputeKeyFingerprint(newPub)
	require.Error(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, changedPubKey))
	_, found = f.keeper.GetSignerKeyByFingerprint(f.ctx, changedPubKey.Fingerprint)
	require.False(t, found)

	fingerprintCollision := types.NewSignerKeyInfo("did:virtengine:issuer:other", updated.PublicKey, types.ProofTypeEd25519, 1, f.ctx.BlockTime())
	fingerprintCollision.State = types.SignerKeyStateActive
	activatedAt := f.ctx.BlockTime().Add(-time.Minute)
	fingerprintCollision.ActivatedAt = &activatedAt
	fingerprintCollision.Metadata[types.SignerKeyMetadataEvidenceTypes] = string(types.AttestationTypeEmailVerification)
	require.Error(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, fingerprintCollision))
}

func TestRegisterSignerKeyAllowsOnlyMonotonicLifecycleTransitions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*types.SignerKeyInfo, time.Time)
	}{
		{
			name: "active to rotating",
			mutate: func(key *types.SignerKeyInfo, _ time.Time) {
				key.State = types.SignerKeyStateRotating
				key.SuccessorKeyID = key.SignerID + ":2"
			},
		},
		{
			name: "active to revoked",
			mutate: func(key *types.SignerKeyInfo, now time.Time) {
				key.State = types.SignerKeyStateRevoked
				revokedAt := now.Add(-time.Minute)
				key.RevokedAt = &revokedAt
				key.RevocationReason = types.RevocationReasonAdministrative
			},
		},
		{
			name: "active to expired",
			mutate: func(key *types.SignerKeyInfo, now time.Time) {
				key.State = types.SignerKeyStateExpired
				expiresAt := now
				key.ExpiresAt = &expiresAt
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newWebEvidenceFixture(t)
			defer CloseStoreIfNeeded(f.stateStore)
			updated := cloneSignerKey(f.signerKey)
			tt.mutate(updated, f.ctx.BlockTime())
			require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, updated))
			stored, found := f.keeper.GetSignerKey(f.ctx, updated.KeyID)
			require.True(t, found)
			require.Equal(t, updated.State, stored.State)
		})
	}

	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)
	pending := f.newPendingSignerKey(t, 2)
	require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, pending))
	activated := cloneSignerKey(pending)
	activated.State = types.SignerKeyStateActive
	activatedAt := f.ctx.BlockTime()
	expiresAt := f.ctx.BlockTime().Add(2 * time.Hour)
	activated.ActivatedAt = &activatedAt
	activated.ExpiresAt = &expiresAt
	require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, activated))
}

func TestRegisterSignerKeyRejectsDowngradeResurrectionAndNonMonotonicLifecycle(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T, *webEvidenceFixture) *types.SignerKeyInfo
		next  func(*types.SignerKeyInfo, time.Time) *types.SignerKeyInfo
	}{
		{
			name: "active to pending",
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.State = types.SignerKeyStatePending
				return updated
			},
		},
		{
			name: "pending to rotating",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				pending := f.newPendingSignerKey(t, 2)
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, pending))
				return pending
			},
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.State = types.SignerKeyStateRotating
				activatedAt := key.CreatedAt
				updated.ActivatedAt = &activatedAt
				updated.SuccessorKeyID = updated.SignerID + ":3"
				return updated
			},
		},
		{
			name: "pending to revoked",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				pending := f.newPendingSignerKey(t, 2)
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, pending))
				return pending
			},
			next: func(key *types.SignerKeyInfo, now time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.State = types.SignerKeyStateRevoked
				revokedAt := now.Add(-time.Minute)
				updated.RevokedAt = &revokedAt
				updated.RevocationReason = types.RevocationReasonAdministrative
				return updated
			},
		},
		{
			name: "pending to expired",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				pending := f.newPendingSignerKey(t, 2)
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, pending))
				return pending
			},
			next: func(key *types.SignerKeyInfo, now time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.State = types.SignerKeyStateExpired
				expiresAt := now.Add(-time.Minute)
				updated.ExpiresAt = &expiresAt
				return updated
			},
		},
		{
			name: "rotating to active",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				rotating := cloneSignerKey(f.signerKey)
				rotating.State = types.SignerKeyStateRotating
				rotating.SuccessorKeyID = rotating.SignerID + ":2"
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, rotating))
				return rotating
			},
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.State = types.SignerKeyStateActive
				updated.SuccessorKeyID = ""
				return updated
			},
		},
		{
			name: "revoked resurrection",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				revoked := cloneSignerKey(f.signerKey)
				revoked.State = types.SignerKeyStateRevoked
				revokedAt := f.ctx.BlockTime().Add(-time.Minute)
				revoked.RevokedAt = &revokedAt
				revoked.RevocationReason = types.RevocationReasonAdministrative
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, revoked))
				return revoked
			},
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.State = types.SignerKeyStateActive
				return updated
			},
		},
		{
			name: "expired resurrection",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				expired := cloneSignerKey(f.signerKey)
				expired.State = types.SignerKeyStateExpired
				expiresAt := f.ctx.BlockTime()
				expired.ExpiresAt = &expiresAt
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, expired))
				return expired
			},
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.State = types.SignerKeyStateActive
				return updated
			},
		},
		{
			name: "activation timestamp backwards",
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				activatedAt := key.ActivatedAt.Add(-time.Second)
				updated.ActivatedAt = &activatedAt
				return updated
			},
		},
		{
			name: "activation timestamp later",
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				activatedAt := key.ActivatedAt.Add(time.Second)
				updated.ActivatedAt = &activatedAt
				return updated
			},
		},
		{
			name: "activation height backwards",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				withHeight := cloneSignerKey(f.signerKey)
				withHeight.Metadata[types.SignerKeyMetadataActivationHeight] = "100"
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, withHeight))
				return withHeight
			},
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.Metadata[types.SignerKeyMetadataActivationHeight] = "99"
				return updated
			},
		},
		{
			name: "activation height later",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				withHeight := cloneSignerKey(f.signerKey)
				withHeight.Metadata[types.SignerKeyMetadataActivationHeight] = "100"
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, withHeight))
				return withHeight
			},
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.Metadata[types.SignerKeyMetadataActivationHeight] = "101"
				return updated
			},
		},
		{
			name: "expiry height moves later",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				withHeight := cloneSignerKey(f.signerKey)
				withHeight.Metadata[types.SignerKeyMetadataExpiryHeight] = "200"
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, withHeight))
				return withHeight
			},
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.Metadata[types.SignerKeyMetadataExpiryHeight] = "201"
				return updated
			},
		},
		{
			name: "same-state evidence policy broadening",
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.Metadata[types.SignerKeyMetadataEvidenceTypes] = updated.Metadata[types.SignerKeyMetadataEvidenceTypes] + "," + string(types.AttestationTypeDocumentVerification)
				return updated
			},
		},
		{
			name: "same-state service metadata policy retarget",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				withPolicy := cloneSignerKey(f.signerKey)
				withPolicy.Metadata[types.SignerKeyMetadataServiceMetadataHash] = sha256Hex("service-policy-1")
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, withPolicy))
				return withPolicy
			},
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.Metadata[types.SignerKeyMetadataServiceMetadataHash] = sha256Hex("service-policy-2")
				return updated
			},
		},
		{
			name: "same-state successor retarget",
			setup: func(t *testing.T, f *webEvidenceFixture) *types.SignerKeyInfo {
				t.Helper()
				rotating := cloneSignerKey(f.signerKey)
				rotating.State = types.SignerKeyStateRotating
				rotating.SuccessorKeyID = rotating.SignerID + ":2"
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, rotating))
				return rotating
			},
			next: func(key *types.SignerKeyInfo, _ time.Time) *types.SignerKeyInfo {
				updated := cloneSignerKey(key)
				updated.SuccessorKeyID = updated.SignerID + ":3"
				return updated
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newWebEvidenceFixture(t)
			defer CloseStoreIfNeeded(f.stateStore)
			current := f.signerKey
			if tt.setup != nil {
				current = tt.setup(t, f)
			}
			require.Error(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, tt.next(current, f.ctx.BlockTime())))
		})
	}
}

func TestRegisterSignerKeyRejectsImmutableIdentityAndSequenceCollisions(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)

	changedSequence := cloneSignerKey(f.signerKey)
	changedSequence.SequenceNumber++
	require.Error(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, changedSequence))

	changedCreatedAt := cloneSignerKey(f.signerKey)
	changedCreatedAt.CreatedAt = changedCreatedAt.CreatedAt.Add(time.Second)
	require.Error(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, changedCreatedAt))

	customSeq2 := f.newPendingSignerKey(t, 2)
	customSeq2.KeyID = "custom-" + customSeq2.Fingerprint[:16]
	require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, customSeq2))
	storedCustom, found := f.keeper.GetSignerKey(f.ctx, customSeq2.KeyID)
	require.True(t, found)
	require.Equal(t, customSeq2.KeyID, storedCustom.KeyID)
	byFingerprint, found := f.keeper.GetSignerKeyByFingerprint(f.ctx, customSeq2.Fingerprint)
	require.True(t, found)
	require.Equal(t, customSeq2.KeyID, byFingerprint.KeyID)

	duplicateSeq := f.newPendingSignerKey(t, 2)
	duplicateSeq.KeyID = duplicateSeq.SignerID + ":2-duplicate"
	require.Error(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, duplicateSeq))

	regression := f.newPendingSignerKey(t, 1)
	regression.KeyID = regression.SignerID + ":1-regression"
	require.Error(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, regression))
}

func TestRegisterSignerKeyRejectsSignerIndexCollision(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)

	colliding := f.newPendingSignerKey(t, 2)
	store := f.ctx.KVStore(f.keeper.StoreKey())
	store.Set(signerIndexKeyForTest(colliding.SignerID, colliding.KeyID), []byte{1})
	require.Error(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, colliding))
}

func TestRegisterSignerKeyRejectsIncoherentLifecycleAndPolicy(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*types.SignerKeyInfo, time.Time)
	}{
		{
			name: "active nil activation",
			mutate: func(key *types.SignerKeyInfo, _ time.Time) {
				key.State = types.SignerKeyStateActive
				key.ActivatedAt = nil
			},
		},
		{
			name: "rotating nil activation",
			mutate: func(key *types.SignerKeyInfo, _ time.Time) {
				key.State = types.SignerKeyStateRotating
				key.ActivatedAt = nil
			},
		},
		{
			name: "expiry before activation",
			mutate: func(key *types.SignerKeyInfo, now time.Time) {
				activatedAt := now.Add(-time.Minute)
				expiresAt := activatedAt
				key.ActivatedAt = &activatedAt
				key.ExpiresAt = &expiresAt
			},
		},
		{
			name: "revocation before activation",
			mutate: func(key *types.SignerKeyInfo, now time.Time) {
				activatedAt := now.Add(-time.Minute)
				revokedAt := activatedAt.Add(-time.Second)
				key.ActivatedAt = &activatedAt
				key.RevokedAt = &revokedAt
			},
		},
		{
			name: "missing evidence policy",
			mutate: func(key *types.SignerKeyInfo, _ time.Time) {
				delete(key.Metadata, types.SignerKeyMetadataEvidenceTypes)
			},
		},
		{
			name: "invalid evidence policy",
			mutate: func(key *types.SignerKeyInfo, _ time.Time) {
				key.Metadata[types.SignerKeyMetadataEvidenceTypes] = "email_verification,not_a_type"
			},
		},
		{
			name: "empty evidence policy item",
			mutate: func(key *types.SignerKeyInfo, _ time.Time) {
				key.Metadata[types.SignerKeyMetadataEvidenceTypes] = "email_verification,"
			},
		},
		{
			name: "expiry height before activation height",
			mutate: func(key *types.SignerKeyInfo, _ time.Time) {
				key.Metadata[types.SignerKeyMetadataActivationHeight] = "200"
				key.Metadata[types.SignerKeyMetadataExpiryHeight] = "200"
			},
		},
		{
			name: "revoked height before activation height",
			mutate: func(key *types.SignerKeyInfo, _ time.Time) {
				key.Metadata[types.SignerKeyMetadataActivationHeight] = "200"
				key.Metadata[types.SignerKeyMetadataRevokedHeight] = "199"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newWebEvidenceFixture(t)
			defer CloseStoreIfNeeded(f.stateStore)
			invalid := cloneSignerKey(f.signerKey)
			tt.mutate(invalid, f.ctx.BlockTime())
			require.Error(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, invalid))

			stored, found := f.keeper.GetSignerKey(f.ctx, f.signerKey.KeyID)
			require.True(t, found)
			require.Equal(t, f.signerKey.ActivatedAt, stored.ActivatedAt)
			require.Equal(t, f.signerKey.Metadata[types.SignerKeyMetadataEvidenceTypes], stored.Metadata[types.SignerKeyMetadataEvidenceTypes])
		})
	}
}

func TestSubmitEmailVerificationProofAcceptsRotatingSignerKey(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)

	rotating := cloneSignerKey(f.signerKey)
	rotating.State = types.SignerKeyStateRotating
	rotating.SuccessorKeyID = rotating.SignerID + ":2"
	require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, rotating))

	msg := f.validEmailMsg(t, "email-rotating-valid", nil)
	resp, err := keeper.NewMsgServerImpl(f.keeper).SubmitEmailVerificationProof(f.ctx, msg)
	require.NoError(t, err)
	require.Equal(t, "email-rotating-valid", resp.VerificationId)
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplay))
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
}

func TestSubmitEmailVerificationProofAllowsEmptyEvidenceMetadata(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)

	msg := f.validEmailMsgForAccountAtWithMetadata(
		t,
		webEvidenceChainID,
		f.accountAddress,
		f.walletPriv,
		"email-empty-service-metadata",
		nil,
		f.ctx.BlockTime(),
		f.ctx.BlockTime().Add(time.Hour),
		nil,
	)
	resp, err := keeper.NewMsgServerImpl(f.keeper).SubmitEmailVerificationProof(f.ctx, msg)
	require.NoError(t, err)
	require.Equal(t, "email-empty-service-metadata", resp.VerificationId)

	stored, found := f.keeper.GetEmailVerificationRecord(f.ctx, "email-empty-service-metadata")
	require.True(t, found)
	require.Empty(t, stored.EvidenceMetadata)
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplay))
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
}

func TestValidateSSOAttestationSubmissionRejectsNilSignerWithoutMutation(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)

	msg := f.validSSOMsg(t, "sso-linkage-nil-signer")
	var att types.SSOAttestation
	mustJSONUnmarshal(msg.AttestationData, &att)

	require.Error(t, f.keeper.ValidateSSOAttestationSubmission(f.ctx, &att, nil))
	require.False(t, f.keeper.IsSSONonceUsed(f.ctx, sha256Hex(att.OIDCNonce)))
	require.Zero(t, f.storePrefixCount(types.PrefixSSONonce))
	require.Zero(t, f.storePrefixCount(types.PrefixSecurityAudit))
}

func TestSubmitSSOVerificationProofRejectsSubjectBindingMismatchWithoutMutation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*types.SSOAttestation, string)
	}{
		{
			name: "linked account mismatch",
			mutate: func(att *types.SSOAttestation, otherAccount string) {
				att.LinkedAccountAddress = otherAccount
			},
		},
		{
			name: "subject account mismatch",
			mutate: func(att *types.SSOAttestation, otherAccount string) {
				att.Subject = types.NewAttestationSubject(otherAccount)
			},
		},
		{
			name: "subject id mismatch",
			mutate: func(att *types.SSOAttestation, otherAccount string) {
				att.Subject.ID = "did:virtengine:" + otherAccount
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newWebEvidenceFixture(t)
			defer CloseStoreIfNeeded(f.stateStore)
			otherAccount, _ := f.createAdditionalWallet(t, 0x66)
			msg := f.validSSOMsg(t, "sso-subject-mismatch")
			var att types.SSOAttestation
			mustJSONUnmarshal(msg.AttestationData, &att)
			tt.mutate(&att, otherAccount.String())
			f.resignSSOMsg(t, msg, &att)

			_, err := keeper.NewMsgServerImpl(f.keeper).SubmitSSOVerificationProof(f.ctx, msg)
			require.Error(t, err)
			_, found := f.keeper.GetSSOLinkage(f.ctx, msg.LinkageId)
			require.False(t, found)
			require.Zero(t, f.storePrefixCount(types.PrefixSSONonce))
			require.Zero(t, f.storePrefixCount(types.PrefixWebEvidenceReplay))
			require.Zero(t, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
		})
	}
}

func TestSubmitWebEvidenceExactRetriesAreIdempotent(t *testing.T) {
	tests := []struct {
		name   string
		run    func(*testing.T, *webEvidenceFixture)
		prefix []byte
	}{
		{
			name: "sso",
			run: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				srv := keeper.NewMsgServerImpl(f.keeper)
				msg := f.validSSOMsg(t, "sso-idempotent")
				first, err := srv.SubmitSSOVerificationProof(f.ctx, msg)
				require.NoError(t, err)
				replayCount := f.storePrefixCount(types.PrefixWebEvidenceReplay)
				nonceCount := f.storePrefixCount(types.PrefixWebEvidenceReplayNonce)
				second, err := srv.SubmitSSOVerificationProof(f.ctx, msg)
				require.NoError(t, err)
				require.Equal(t, first, second)
				require.Equal(t, replayCount, f.storePrefixCount(types.PrefixWebEvidenceReplay))
				require.Equal(t, nonceCount, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
			},
		},
		{
			name: "email",
			run: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				srv := keeper.NewMsgServerImpl(f.keeper)
				msg := f.validEmailMsg(t, "email-idempotent", nil)
				first, err := srv.SubmitEmailVerificationProof(f.ctx, msg)
				require.NoError(t, err)
				replayCount := f.storePrefixCount(types.PrefixWebEvidenceReplay)
				nonceCount := f.storePrefixCount(types.PrefixWebEvidenceReplayNonce)
				second, err := srv.SubmitEmailVerificationProof(f.ctx, msg)
				require.NoError(t, err)
				require.Equal(t, first, second)
				require.Equal(t, replayCount, f.storePrefixCount(types.PrefixWebEvidenceReplay))
				require.Equal(t, nonceCount, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
			},
		},
		{
			name: "sms",
			run: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				srv := keeper.NewMsgServerImpl(f.keeper)
				msg := f.validSMSMsg(t, "sms-idempotent")
				first, err := srv.SubmitSMSVerificationProof(f.ctx, msg)
				require.NoError(t, err)
				replayCount := f.storePrefixCount(types.PrefixWebEvidenceReplay)
				nonceCount := f.storePrefixCount(types.PrefixWebEvidenceReplayNonce)
				second, err := srv.SubmitSMSVerificationProof(f.ctx, msg)
				require.NoError(t, err)
				require.Equal(t, first, second)
				require.Equal(t, replayCount, f.storePrefixCount(types.PrefixWebEvidenceReplay))
				require.Equal(t, nonceCount, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
			},
		},
		{
			name: "social",
			run: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				srv := keeper.NewMsgServerImpl(f.keeper)
				msg := f.validSocialMsg(t, "social-idempotent")
				first, err := srv.SubmitSocialMediaScope(f.ctx, msg)
				require.NoError(t, err)
				replayCount := f.storePrefixCount(types.PrefixWebEvidenceReplay)
				nonceCount := f.storePrefixCount(types.PrefixWebEvidenceReplayNonce)
				second, err := srv.SubmitSocialMediaScope(f.ctx, msg)
				require.NoError(t, err)
				require.Equal(t, first, second)
				require.Equal(t, replayCount, f.storePrefixCount(types.PrefixWebEvidenceReplay))
				require.Equal(t, nonceCount, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newWebEvidenceFixture(t)
			defer CloseStoreIfNeeded(f.stateStore)
			tt.run(t, f)
		})
	}
}

func TestSubmitEmailVerificationProofRejectsExactReplayWithChangedStoragePointer(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)
	srv := keeper.NewMsgServerImpl(f.keeper)

	msg := f.validEmailMsg(t, "email-storage-idempotency", nil)
	_, err := srv.SubmitEmailVerificationProof(f.ctx, msg)
	require.NoError(t, err)

	replayCount := f.storePrefixCount(types.PrefixWebEvidenceReplay)
	nonceCount := f.storePrefixCount(types.PrefixWebEvidenceReplayNonce)

	msg.EvidenceStorageRef = "vault://email/changed-storage-ref"
	_, err = srv.SubmitEmailVerificationProof(f.ctx, msg)
	require.Error(t, err)
	require.Equal(t, replayCount, f.storePrefixCount(types.PrefixWebEvidenceReplay))
	require.Equal(t, nonceCount, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))

	stored, found := f.keeper.GetEmailVerificationRecord(f.ctx, "email-storage-idempotency")
	require.True(t, found)
	require.NotEqual(t, msg.EvidenceStorageRef, stored.EvidenceStorageRef)
}

func TestSubmitEmailVerificationProofRejectsInvalidCryptoWithoutMutation(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*testing.T, *webEvidenceFixture)
		mutateMsg func(*types.MsgSubmitEmailVerificationProof)
	}{
		{
			name: "unknown signer key",
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				f.deleteSignerKeyForTest()
			},
		},
		{
			name: "revoked signer key",
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				revokedAt := f.ctx.BlockTime().Add(-time.Minute)
				f.signerKey.State = types.SignerKeyStateRevoked
				f.signerKey.RevokedAt = &revokedAt
				f.signerKey.RevocationReason = types.RevocationReasonAdministrative
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, f.signerKey))
			},
		},
		{
			name: "expired signer key",
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				f.signerKey.State = types.SignerKeyStateExpired
				f.forceSignerKeyForTest(t, f.signerKey)
			},
		},
		{
			name: "preactivation signer key time",
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				activatedAt := f.ctx.BlockTime().Add(time.Hour)
				f.signerKey.ActivatedAt = &activatedAt
				f.forceSignerKeyForTest(t, f.signerKey)
			},
		},
		{
			name: "preactivation signer key height",
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				f.signerKey.Metadata[types.SignerKeyMetadataActivationHeight] = "200"
				require.NoError(t, f.keeper.RegisterSignerKey(f.ctx, webEvidenceAuthority, f.signerKey))
			},
		},
		{
			name: "pending signer key",
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				f.signerKey.State = types.SignerKeyStatePending
				f.forceSignerKeyForTest(t, f.signerKey)
			},
		},
		{
			name: "wrong committed evidence type",
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				f.signerKey.Metadata[types.SignerKeyMetadataEvidenceTypes] = string(types.AttestationTypeSMSVerification)
				f.forceSignerKeyForTest(t, f.signerKey)
			},
		},
		{
			name: "missing committed evidence type policy",
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				delete(f.signerKey.Metadata, types.SignerKeyMetadataEvidenceTypes)
				f.forceSignerKeyForTest(t, f.signerKey)
			},
		},
		{
			name:      "issued before signer key activation",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {},
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				msg := f.validEmailMsgAt(t, "email-proof-1", nil, f.ctx.BlockTime().Add(-2*time.Minute), f.ctx.BlockTime().Add(time.Hour))
				*testEmailMsgOverride = *msg
			},
		},
		{
			name: "unknown unsigned metadata",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				att.Metadata["unsigned.extra"] = "value"
				msg.AttestationData = mustAttestationJSON(&att)
				msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
			},
		},
		{
			name: "metadata fingerprint tampering",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				att.Metadata[types.WebEvidenceMetadataIssuerFingerprint] = hex.EncodeToString(bytesOf('f', 32))
				msg.AttestationData = mustAttestationJSON(&att)
				msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
			},
		},
		{
			name: "algorithm tampering",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				att.Proof.Type = types.ProofTypeSecp256k1
				msg.AttestationData = mustAttestationJSON(&att)
				msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
			},
		},
		{
			name: "proof created timestamp tampering",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				att.Proof.Created = att.Proof.Created.Add(time.Minute)
				msg.AttestationData = mustAttestationJSON(&att)
				msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
			},
		},
		{
			name: "verification method tampering",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				att.Proof.VerificationMethod = "did:virtengine:issuer:web#other-key"
				msg.AttestationData = mustAttestationJSON(&att)
				msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
			},
		},
		{
			name: "issuer signature tampering",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				sig, err := base64.StdEncoding.DecodeString(att.Proof.ProofValue)
				if err != nil {
					panic(err)
				}
				sig[0] ^= 0x01
				att.Proof.ProofValue = base64.StdEncoding.EncodeToString(sig)
				msg.AttestationData = mustAttestationJSON(&att)
				msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
			},
		},
		{
			name: "missing account signature",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {
				msg.AccountSignature = nil
			},
		},
		{
			name: "invalid account signature",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {
				msg.AccountSignature[0] ^= 0x01
			},
		},
		{
			name: "missing wallet",
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				f.deleteWalletForTest()
			},
		},
		{
			name:      "future attestation",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {},
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				msg := f.validEmailMsgAt(t, "email-proof-1", nil, f.ctx.BlockTime().Add(time.Hour), f.ctx.BlockTime().Add(2*time.Hour))
				*testEmailMsgOverride = *msg
			},
		},
		{
			name:      "expired attestation",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {},
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				msg := f.validEmailMsgAt(t, "email-proof-1", nil, f.ctx.BlockTime().Add(-2*time.Hour), f.ctx.BlockTime().Add(-time.Hour))
				*testEmailMsgOverride = *msg
			},
		},
		{
			name:      "stale attestation",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {},
			setup: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				msg := f.validEmailMsgAt(t, "email-proof-1", nil, f.ctx.BlockTime().Add(-25*time.Hour), f.ctx.BlockTime().Add(time.Hour))
				*testEmailMsgOverride = *msg
			},
		},
		{
			name: "caller field tampering",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {
				msg.EmailHash = sha256Hex("changed-email")
			},
		},
		{
			name: "chain tampering",
			mutateMsg: func(msg *types.MsgSubmitEmailVerificationProof) {
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				att.Metadata[types.WebEvidenceMetadataChainID] = "other-chain"
				msg.AttestationData = mustAttestationJSON(&att)
				msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testEmailMsgOverride = &types.MsgSubmitEmailVerificationProof{}
			f := newWebEvidenceFixture(t)
			defer CloseStoreIfNeeded(f.stateStore)
			if tt.setup != nil {
				tt.setup(t, f)
			}
			msg := f.validEmailMsg(t, "email-proof-1", nil)
			if testEmailMsgOverride.AttestationData != nil {
				msg = testEmailMsgOverride
			}
			if tt.mutateMsg != nil {
				tt.mutateMsg(msg)
			}

			_, err := keeper.NewMsgServerImpl(f.keeper).SubmitEmailVerificationProof(f.ctx, msg)
			require.Error(t, err)
			_, found := f.keeper.GetEmailVerificationRecord(f.ctx, "email-proof-1")
			require.False(t, found)
			require.False(t, f.keeper.IsEmailNonceUsed(f.ctx, hashTestEvidence([]byte(msg.Nonce))))
			_, _, scoreFound := f.keeper.GetScore(f.ctx, f.accountAddress)
			require.False(t, scoreFound)
			require.Zero(t, f.storePrefixCount(types.PrefixWebEvidenceReplay))
			require.Zero(t, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
		})
	}
}

func TestSubmitWebEvidenceRejectsFieldTamperingWithoutMutation(t *testing.T) {
	tests := []struct {
		name   string
		run    func(*testing.T, *webEvidenceFixture) error
		assert func(*testing.T, *webEvidenceFixture)
	}{
		{
			name: "sso linkage id",
			run: func(t *testing.T, f *webEvidenceFixture) error {
				t.Helper()
				msg := f.validSSOMsg(t, "sso-linkage-original")
				msg.LinkageId = "sso-linkage-tampered"
				_, err := keeper.NewMsgServerImpl(f.keeper).SubmitSSOVerificationProof(f.ctx, msg)
				return err
			},
			assert: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				_, found := f.keeper.GetSSOLinkage(f.ctx, "sso-linkage-tampered")
				require.False(t, found)
				require.Zero(t, f.storePrefixCount(types.PrefixSSONonce))
			},
		},
		{
			name: "sms carrier and voip",
			run: func(t *testing.T, f *webEvidenceFixture) error {
				t.Helper()
				msg := f.validSMSMsg(t, "sms-tampered")
				msg.CarrierType = "voip"
				msg.IsVoip = true
				_, err := keeper.NewMsgServerImpl(f.keeper).SubmitSMSVerificationProof(f.ctx, msg)
				return err
			},
			assert: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				_, found := f.keeper.GetSMSVerificationRecord(f.ctx, "sms-tampered")
				require.False(t, found)
			},
		},
		{
			name: "social encrypted payload",
			run: func(t *testing.T, f *webEvidenceFixture) error {
				t.Helper()
				msg := f.validSocialMsg(t, "social-tampered")
				msg.EncryptedPayload.Ciphertext = []byte("changed-ciphertext")
				_, err := keeper.NewMsgServerImpl(f.keeper).SubmitSocialMediaScope(f.ctx, msg)
				return err
			},
			assert: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				_, found := f.keeper.GetSocialMediaScope(f.ctx, "social-tampered")
				require.False(t, found)
			},
		},
		{
			name: "email subject id",
			run: func(t *testing.T, f *webEvidenceFixture) error {
				t.Helper()
				msg := f.validEmailMsg(t, "email-subject-id-mismatch", nil)
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				att.Subject.ID = "did:virtengine:" + f.accountAddress + "-other"
				f.resignEmailMsgForTest(t, msg, &att)
				_, err := keeper.NewMsgServerImpl(f.keeper).SubmitEmailVerificationProof(f.ctx, msg)
				return err
			},
			assert: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				_, found := f.keeper.GetEmailVerificationRecord(f.ctx, "email-subject-id-mismatch")
				require.False(t, found)
			},
		},
		{
			name: "sms subject id",
			run: func(t *testing.T, f *webEvidenceFixture) error {
				t.Helper()
				msg := f.validSMSMsg(t, "sms-subject-id-mismatch")
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				att.Subject.ID = "did:virtengine:" + f.accountAddress + "-other"
				f.resignSMSMsgForTest(t, msg, &att)
				_, err := keeper.NewMsgServerImpl(f.keeper).SubmitSMSVerificationProof(f.ctx, msg)
				return err
			},
			assert: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				_, found := f.keeper.GetSMSVerificationRecord(f.ctx, "sms-subject-id-mismatch")
				require.False(t, found)
			},
		},
		{
			name: "social subject id",
			run: func(t *testing.T, f *webEvidenceFixture) error {
				t.Helper()
				msg := f.validSocialMsg(t, "social-subject-id-mismatch")
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				att.Subject.ID = "did:virtengine:" + f.accountAddress + "-other"
				f.resignSocialMsgForTest(t, msg, &att)
				_, err := keeper.NewMsgServerImpl(f.keeper).SubmitSocialMediaScope(f.ctx, msg)
				return err
			},
			assert: func(t *testing.T, f *webEvidenceFixture) {
				t.Helper()
				_, found := f.keeper.GetSocialMediaScope(f.ctx, "social-subject-id-mismatch")
				require.False(t, found)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newWebEvidenceFixture(t)
			defer CloseStoreIfNeeded(f.stateStore)
			require.Error(t, tt.run(t, f))
			tt.assert(t, f)
			_, _, scoreFound := f.keeper.GetScore(f.ctx, f.accountAddress)
			require.False(t, scoreFound)
			require.Zero(t, f.storePrefixCount(types.PrefixWebEvidenceReplay))
			require.Zero(t, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
		})
	}
}

func TestSubmitEmailVerificationProofRejectsInvalidRecordBeforeReplayWrite(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)

	msg := f.validEmailMsg(t, "email-invalid-record", nil)
	msg.EvidenceStorageRef = ""

	_, err := keeper.NewMsgServerImpl(f.keeper).SubmitEmailVerificationProof(f.ctx, msg)
	require.Error(t, err)
	_, found := f.keeper.GetEmailVerificationRecord(f.ctx, "email-invalid-record")
	require.False(t, found)
	require.False(t, f.keeper.IsEmailNonceUsed(f.ctx, hashTestEvidence([]byte(msg.Nonce))))
	_, _, scoreFound := f.keeper.GetScore(f.ctx, f.accountAddress)
	require.False(t, scoreFound)
	require.Zero(t, f.storePrefixCount(types.PrefixWebEvidenceReplay))
	require.Zero(t, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
}

func TestSubmitEmailVerificationProofRejectsReplayWithNoSecondMutation(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)
	srv := keeper.NewMsgServerImpl(f.keeper)

	nonce := nonceBytes("replay-nonce")
	first := f.validEmailMsg(t, "email-proof-1", nonce)
	_, err := srv.SubmitEmailVerificationProof(f.ctx, first)
	require.NoError(t, err)

	second := f.validEmailMsg(t, "email-proof-2", nonce)
	_, err = srv.SubmitEmailVerificationProof(f.ctx, second)
	require.Error(t, err)
	_, found := f.keeper.GetEmailVerificationRecord(f.ctx, "email-proof-2")
	require.False(t, found)
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplay))
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
}

func TestSubmitWebEvidenceRejectsNonceReplayAcrossActionAndEnvironment(t *testing.T) {
	tests := []struct {
		name string
		run  func(*testing.T, *webEvidenceFixture, []byte) error
	}{
		{
			name: "action changes",
			run: func(t *testing.T, f *webEvidenceFixture, nonce []byte) error {
				t.Helper()
				msg := f.validSMSMsg(t, "sms-cross-action")
				var att types.VerificationAttestation
				mustJSONUnmarshal(msg.AttestationData, &att)
				att.Nonce = hex.EncodeToString(nonce)
				att.Proof.Nonce = att.Nonce
				issuedAt := f.ctx.BlockTime()
				expiresAt := issuedAt.Add(time.Hour)
				phoneHash := msg.PhoneHash
				phoneSalt := msg.PhoneHashSalt
				countryHash := msg.CountryCodeHash
				callerFields := map[string]string{
					"verification_id":   msg.VerificationId,
					"phone_hash":        phoneHash,
					"phone_hash_salt":   phoneSalt,
					"country_code_hash": countryHash,
					"is_voip":           "false",
					"carrier_type":      msg.CarrierType,
					"validator_address": msg.ValidatorAddress,
					"verified_at_unix":  strconv.FormatInt(issuedAt.Unix(), 10),
					"expires_at_unix":   strconv.FormatInt(expiresAt.Unix(), 10),
				}
				signed, evidence := f.signedVerificationAttestation(t, types.AttestationTypeSMSVerification, types.WebEvidenceActionSubmitSMS, msg.VerificationId, msg.VerificationId, nonce, issuedAt, expiresAt, msg.EvidenceMetadata, callerFields)
				msg.AttestationData = mustAttestationJSON(signed)
				msg.AccountSignature = f.signAccount(t, evidence)
				msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
				_, err := keeper.NewMsgServerImpl(f.keeper).SubmitSMSVerificationProof(f.ctx, msg)
				return err
			},
		},
		{
			name: "service metadata changes",
			run: func(t *testing.T, f *webEvidenceFixture, nonce []byte) error {
				t.Helper()
				msg := f.validEmailMsgForAccountAtWithMetadata(
					t,
					webEvidenceChainID,
					f.accountAddress,
					f.walletPriv,
					"email-cross-environment",
					nonce,
					f.ctx.BlockTime(),
					f.ctx.BlockTime().Add(time.Hour),
					map[string]string{"source": "different-email-provider"},
				)
				_, err := keeper.NewMsgServerImpl(f.keeper).SubmitEmailVerificationProof(f.ctx, msg)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newWebEvidenceFixture(t)
			defer CloseStoreIfNeeded(f.stateStore)
			srv := keeper.NewMsgServerImpl(f.keeper)
			nonce := nonceBytes("cross-action-environment-replay")
			first := f.validEmailMsg(t, "email-cross-original", nonce)
			_, err := srv.SubmitEmailVerificationProof(f.ctx, first)
			require.NoError(t, err)

			require.Error(t, tt.run(t, f, nonce))
			require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplay))
			require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
			_, _, scoreFound := f.keeper.GetScore(f.ctx, f.accountAddress)
			require.True(t, scoreFound)
		})
	}
}

func TestSubmitEmailVerificationProofRejectsCrossAccountAndCrossChainReplay(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)
	srv := keeper.NewMsgServerImpl(f.keeper)

	nonce := nonceBytes("cross-boundary-replay")
	first := f.validEmailMsg(t, "email-cross-original", nonce)
	_, err := srv.SubmitEmailVerificationProof(f.ctx, first)
	require.NoError(t, err)
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplay))

	otherAccount, otherPriv := f.createAdditionalWallet(t, 0x55)
	crossAccount := f.validEmailMsgForAccountAt(t, webEvidenceChainID, otherAccount.String(), otherPriv, "email-cross-account", nonce, f.ctx.BlockTime(), f.ctx.BlockTime().Add(time.Hour))
	_, err = srv.SubmitEmailVerificationProof(f.ctx, crossAccount)
	require.Error(t, err)
	_, found := f.keeper.GetEmailVerificationRecord(f.ctx, "email-cross-account")
	require.False(t, found)
	_, _, scoreFound := f.keeper.GetScore(f.ctx, otherAccount.String())
	require.False(t, scoreFound)
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplay))
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))

	otherCtx := sdk.NewContext(f.stateStore, cmtproto.Header{
		ChainID: "ve-other-1",
		Time:    f.ctx.BlockTime(),
		Height:  f.ctx.BlockHeight() + 1,
	}, false, log.NewNopLogger())
	crossChain := f.validEmailMsgForAccountAt(t, "ve-other-1", f.accountAddress, f.walletPriv, "email-cross-chain", nonce, otherCtx.BlockTime(), otherCtx.BlockTime().Add(time.Hour))
	_, err = srv.SubmitEmailVerificationProof(otherCtx, crossChain)
	require.Error(t, err)
	_, found = f.keeper.GetEmailVerificationRecord(otherCtx, "email-cross-chain")
	require.False(t, found)
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplay))
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
}

func TestSubmitSocialMediaScopeRejectsNonceReplayAcrossChangedScope(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)
	srv := keeper.NewMsgServerImpl(f.keeper)

	nonce := nonceBytes("social-replay-nonce")
	first := f.validSocialMsgWithNonce(t, "social-scope-1", nonce)
	_, err := srv.SubmitSocialMediaScope(f.ctx, first)
	require.NoError(t, err)

	second := f.validSocialMsgWithNonce(t, "social-scope-2", nonce)
	_, err = srv.SubmitSocialMediaScope(f.ctx, second)
	require.Error(t, err)
	_, found := f.keeper.GetSocialMediaScope(f.ctx, "social-scope-2")
	require.False(t, found)
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplay))
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
}

func TestSubmitSSOVerificationProofRejectsNonceReplayAcrossChangedChallenge(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)
	srv := keeper.NewMsgServerImpl(f.keeper)

	nonce := nonceBytes("sso-challenge-replay-nonce")
	first := f.validSSOMsgWithNonceAndOIDC(t, "sso-challenge-replay", nonce, "oidc-challenge-original")
	_, err := srv.SubmitSSOVerificationProof(f.ctx, first)
	require.NoError(t, err)

	second := f.validSSOMsgWithNonceAndOIDC(t, "sso-challenge-replay", nonce, "oidc-challenge-changed")
	_, err = srv.SubmitSSOVerificationProof(f.ctx, second)
	require.Error(t, err)
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplay))
	require.Equal(t, 1, f.storePrefixCount(types.PrefixWebEvidenceReplayNonce))
}

func TestSubmitSocialMediaScopeRejectsScoreFieldTampering(t *testing.T) {
	f := newWebEvidenceFixture(t)
	defer CloseStoreIfNeeded(f.stateStore)

	msg := f.validSocialMsg(t, "social-scope-1")
	msg.IsVerified = false

	_, err := keeper.NewMsgServerImpl(f.keeper).SubmitSocialMediaScope(f.ctx, msg)
	require.Error(t, err)
	_, found := f.keeper.GetSocialMediaScope(f.ctx, "social-scope-1")
	require.False(t, found)
	_, _, scoreFound := f.keeper.GetScore(f.ctx, f.accountAddress)
	require.False(t, scoreFound)
}

var testEmailMsgOverride = &types.MsgSubmitEmailVerificationProof{}

func newWebEvidenceFixture(t *testing.T) *webEvidenceFixture {
	t.Helper()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	ctx, stateStore := createContextWithStore(t, storeKey)
	k := keeper.NewKeeper(cdc, storeKey, webEvidenceAuthority)
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	account := sdk.AccAddress([]byte("web_evidence_acct001"))
	walletPub, walletPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	walletID := keeper.GenerateWalletID(account.String())
	bindingSig := ed25519.Sign(walletPriv, types.GetWalletBindingMessage(walletID, account.String()))
	_, err = k.CreateWallet(ctx, account, bindingSig, walletPub)
	require.NoError(t, err)

	issuerPub, issuerPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	now := ctx.BlockTime()
	signerKey := types.NewSignerKeyInfo("did:virtengine:issuer:web", issuerPub, types.ProofTypeEd25519, 1, now.Add(-time.Hour))
	activatedAt := now.Add(-time.Minute)
	expiresAt := now.Add(24 * time.Hour)
	signerKey.State = types.SignerKeyStateActive
	signerKey.ActivatedAt = &activatedAt
	signerKey.ExpiresAt = &expiresAt
	signerKey.Metadata[types.SignerKeyMetadataEvidenceTypes] = stringsJoin([]string{
		string(types.AttestationTypeSSOVerification),
		string(types.AttestationTypeEmailVerification),
		string(types.AttestationTypeSMSVerification),
		string(types.AttestationTypeSocialMediaVerification),
	})
	require.NoError(t, k.RegisterSignerKey(ctx, webEvidenceAuthority, signerKey))

	issuer := types.AttestationIssuer{
		ID:             signerKey.SignerID,
		KeyID:          signerKey.KeyID,
		KeyFingerprint: signerKey.Fingerprint,
	}
	nameHash := types.HashSocialMediaField("Jane Doe")
	wallet, found := k.GetWallet(ctx, account)
	require.True(t, found)
	nameHashBytes, err := hex.DecodeString(nameHash)
	require.NoError(t, err)
	wallet.DerivedFeatures.DocFieldHashes[types.DocFieldNameHash] = nameHashBytes
	wallet.DerivedFeatures.LastComputedAt = ctx.BlockTime()
	require.NoError(t, k.SetWallet(ctx, wallet))

	return &webEvidenceFixture{
		ctx:            ctx,
		stateStore:     stateStore,
		keeper:         k,
		account:        account,
		accountAddress: account.String(),
		walletPriv:     walletPriv,
		issuerPriv:     issuerPriv,
		issuer:         issuer,
		signerKey:      signerKey,
	}
}

func (f *webEvidenceFixture) validEmailMsg(t *testing.T, verificationID string, nonce []byte) *types.MsgSubmitEmailVerificationProof {
	t.Helper()
	return f.validEmailMsgAt(t, verificationID, nonce, f.ctx.BlockTime(), f.ctx.BlockTime().Add(time.Hour))
}

func (f *webEvidenceFixture) validEmailMsgAt(t *testing.T, verificationID string, nonce []byte, issuedAt time.Time, expiresAt time.Time) *types.MsgSubmitEmailVerificationProof {
	t.Helper()
	return f.validEmailMsgForAccountAt(t, webEvidenceChainID, f.accountAddress, f.walletPriv, verificationID, nonce, issuedAt, expiresAt)
}

func (f *webEvidenceFixture) validEmailMsgForAccountAt(
	t *testing.T,
	chainID string,
	accountAddress string,
	walletPriv ed25519.PrivateKey,
	verificationID string,
	nonce []byte,
	issuedAt time.Time,
	expiresAt time.Time,
) *types.MsgSubmitEmailVerificationProof {
	t.Helper()
	return f.validEmailMsgForAccountAtWithMetadata(t, chainID, accountAddress, walletPriv, verificationID, nonce, issuedAt, expiresAt, map[string]string{"source": "email-provider"})
}

func (f *webEvidenceFixture) validEmailMsgForAccountAtWithMetadata(
	t *testing.T,
	chainID string,
	accountAddress string,
	walletPriv ed25519.PrivateKey,
	verificationID string,
	nonce []byte,
	issuedAt time.Time,
	expiresAt time.Time,
	metadata map[string]string,
) *types.MsgSubmitEmailVerificationProof {
	t.Helper()
	if nonce == nil {
		nonce = nonceBytes(verificationID)
	}
	emailHash := sha256Hex("jane@example.com")
	domainHash := sha256Hex("example.com")
	callerFields := map[string]string{
		"verification_id":   verificationID,
		"email_hash":        emailHash,
		"domain_hash":       domainHash,
		"nonce":             hex.EncodeToString(nonce),
		"is_organizational": "true",
		"verified_at_unix":  strconv.FormatInt(issuedAt.Unix(), 10),
		"expires_at_unix":   strconv.FormatInt(expiresAt.Unix(), 10),
	}
	att, evidence := f.signedVerificationAttestationForAccount(t, chainID, accountAddress, types.AttestationTypeEmailVerification, types.WebEvidenceActionSubmitEmail, verificationID, verificationID, nonce, issuedAt, expiresAt, metadata, callerFields)
	accountSig := signAccountFor(t, walletPriv, evidence)
	attestationData := mustAttestationJSON(att)
	return &types.MsgSubmitEmailVerificationProof{
		AccountAddress:         accountAddress,
		VerificationId:         verificationID,
		EmailHash:              emailHash,
		DomainHash:             domainHash,
		Nonce:                  hex.EncodeToString(nonce),
		VerifiedAt:             issuedAt.Unix(),
		ExpiresAt:              expiresAt.Unix(),
		AttestationData:        attestationData,
		AccountSignature:       accountSig,
		IsOrganizational:       true,
		EvidenceHash:           hashTestEvidence(attestationData),
		EvidenceStorageBackend: string(types.StorageBackendWaldur),
		EvidenceStorageRef:     "vault://email/" + verificationID,
		EvidenceMetadata:       metadata,
	}
}

func (f *webEvidenceFixture) validSMSMsg(t *testing.T, verificationID string) *types.MsgSubmitSMSVerificationProof {
	t.Helper()
	issuedAt := f.ctx.BlockTime()
	expiresAt := issuedAt.Add(time.Hour)
	phoneHash := sha256Hex("+15555550123")
	phoneSalt := hex.EncodeToString(bytesOf(0x11, 32))
	countryHash := sha256Hex("US")
	metadata := map[string]string{"source": "sms-provider"}
	callerFields := map[string]string{
		"verification_id":   verificationID,
		"phone_hash":        phoneHash,
		"phone_hash_salt":   phoneSalt,
		"country_code_hash": countryHash,
		"is_voip":           "false",
		"carrier_type":      "mobile",
		"validator_address": "validator-1",
		"verified_at_unix":  strconv.FormatInt(issuedAt.Unix(), 10),
		"expires_at_unix":   strconv.FormatInt(expiresAt.Unix(), 10),
	}
	att, evidence := f.signedVerificationAttestation(t, types.AttestationTypeSMSVerification, types.WebEvidenceActionSubmitSMS, verificationID, verificationID, nonceBytes(verificationID), issuedAt, expiresAt, metadata, callerFields)
	accountSig := f.signAccount(t, evidence)
	attestationData := mustAttestationJSON(att)
	return &types.MsgSubmitSMSVerificationProof{
		AccountAddress:         f.accountAddress,
		VerificationId:         verificationID,
		PhoneHash:              phoneHash,
		PhoneHashSalt:          phoneSalt,
		CountryCodeHash:        countryHash,
		VerifiedAt:             issuedAt.Unix(),
		ExpiresAt:              expiresAt.Unix(),
		IsVoip:                 false,
		CarrierType:            "mobile",
		ValidatorAddress:       "validator-1",
		AttestationData:        attestationData,
		AccountSignature:       accountSig,
		EvidenceHash:           hashTestEvidence(attestationData),
		EvidenceStorageBackend: string(types.StorageBackendWaldur),
		EvidenceStorageRef:     "vault://sms/" + verificationID,
		EvidenceMetadata:       metadata,
	}
}

func (f *webEvidenceFixture) validSocialMsg(t *testing.T, scopeID string) *types.MsgSubmitSocialMediaScope {
	t.Helper()
	return f.validSocialMsgWithNonce(t, scopeID, nonceBytes(scopeID))
}

func (f *webEvidenceFixture) validSocialMsgWithNonce(t *testing.T, scopeID string, nonce []byte) *types.MsgSubmitSocialMediaScope {
	t.Helper()
	issuedAt := f.ctx.BlockTime()
	expiresAt := issuedAt.Add(time.Hour)
	payload := validEncryptedPayload(t)
	payloadDigest := encryptedPayloadDigestForTest(t, &payload)
	nameHash := types.HashSocialMediaField("Jane Doe")
	emailHash := types.HashSocialMediaField("jane@example.com")
	usernameHash := types.HashSocialMediaField("jane")
	orgHash := types.HashSocialMediaField("Example Org")
	accountCreatedAt := issuedAt.Add(-365 * 24 * time.Hour).Unix()
	metadata := map[string]string{"source": "social-provider"}
	callerFields := map[string]string{
		"scope_id":                 scopeID,
		"provider":                 string(types.SocialMediaProviderGoogle),
		"profile_name_hash":        nameHash,
		"email_hash":               emailHash,
		"username_hash":            usernameHash,
		"org_hash":                 orgHash,
		"account_created_at_unix":  strconv.FormatInt(accountCreatedAt, 10),
		"account_age_days":         "365",
		"is_verified":              "true",
		"friend_count_range":       "500-999",
		"encrypted_payload_digest": payloadDigest,
	}
	att, evidence := f.signedVerificationAttestation(t, types.AttestationTypeSocialMediaVerification, types.WebEvidenceActionSubmitSocial, scopeID, scopeID, nonce, issuedAt, expiresAt, metadata, callerFields)
	accountSig := f.signAccount(t, evidence)
	attestationData := mustAttestationJSON(att)
	return &types.MsgSubmitSocialMediaScope{
		AccountAddress:         f.accountAddress,
		ScopeId:                scopeID,
		Provider:               types.SocialMediaProviderToProto(types.SocialMediaProviderGoogle),
		ProfileNameHash:        nameHash,
		EmailHash:              emailHash,
		UsernameHash:           usernameHash,
		OrgHash:                orgHash,
		AccountCreatedAt:       accountCreatedAt,
		AccountAgeDays:         365,
		IsVerified:             true,
		FriendCountRange:       "500-999",
		AttestationData:        attestationData,
		AccountSignature:       accountSig,
		EncryptedPayload:       payload,
		EvidenceHash:           hashTestEvidence(attestationData),
		EvidenceStorageBackend: string(types.StorageBackendWaldur),
		EvidenceStorageRef:     "vault://social/" + scopeID,
		EvidenceMetadata:       metadata,
	}
}

func (f *webEvidenceFixture) validSSOMsg(t *testing.T, linkageID string) *types.MsgSubmitSSOVerificationProof {
	t.Helper()
	issuedAt := f.ctx.BlockTime()
	expiresAt := issuedAt.Add(time.Hour)
	subject := types.NewAttestationSubject(f.accountAddress)
	att := types.NewSSOAttestation(
		f.issuer,
		subject,
		"https://accounts.google.com",
		"google-subject-1",
		types.SSOProviderGoogle,
		"oidc-nonce-"+linkageID,
		nonceBytes(linkageID),
		issuedAt,
		time.Hour,
	)
	att.SetEmail("jane@example.com", "example.com", true)
	metadata := map[string]string{"source": "sso-provider"}
	canonical, err := att.CanonicalBytes()
	require.NoError(t, err)
	evidence := types.NewWebEvidenceContext(types.WebEvidenceContextConfig{
		ChainID:             webEvidenceChainID,
		AccountAddress:      f.accountAddress,
		EvidenceType:        types.AttestationTypeSSOVerification,
		Action:              types.WebEvidenceActionSubmitSSO,
		ScopeID:             linkageID,
		AttestationDigest:   types.WebEvidenceDigestHex(canonical),
		Issuer:              f.issuer,
		IssuerAlgorithm:     types.ProofTypeEd25519,
		Nonce:               att.Nonce,
		Challenge:           att.OIDCNonce,
		IssuedAt:            issuedAt,
		ExpiresAt:           expiresAt,
		ServiceMetadataHash: serviceMetadataHashForTest(t, metadata),
		CallerFields: map[string]string{
			"linkage_id":             linkageID,
			"provider":               string(att.ProviderType),
			"oidc_issuer":            att.OIDCIssuer,
			"subject_hash":           att.SubjectHash,
			"email_hash":             att.EmailHash,
			"email_domain_hash":      att.EmailDomainHash,
			"tenant_id_hash":         att.TenantIDHash,
			"oidc_nonce":             att.OIDCNonce,
			"email_verified":         "true",
			"linked_account_address": att.LinkedAccountAddress,
		},
	})
	require.NoError(t, evidence.ApplyToAttestation(&att.VerificationAttestation))
	signBytes, err := evidence.IssuerSignBytes()
	require.NoError(t, err)
	att.SetProof(types.NewAttestationProof(types.ProofTypeEd25519, issuedAt, f.issuer.ID+"#"+f.issuer.KeyID, ed25519.Sign(f.issuerPriv, signBytes), att.Nonce))
	att.SetLinkageSignature(f.signAccount(t, evidence))
	attestationData, err := json.Marshal(att)
	require.NoError(t, err)
	return &types.MsgSubmitSSOVerificationProof{
		AccountAddress:         f.accountAddress,
		LinkageId:              linkageID,
		AttestationData:        attestationData,
		EvidenceHash:           hashTestEvidence(attestationData),
		EvidenceStorageBackend: string(types.StorageBackendWaldur),
		EvidenceStorageRef:     "vault://sso/" + linkageID,
		EvidenceMetadata:       metadata,
	}
}

func (f *webEvidenceFixture) validSSOMsgWithNonceAndOIDC(t *testing.T, linkageID string, nonce []byte, oidcNonce string) *types.MsgSubmitSSOVerificationProof {
	t.Helper()
	msg := f.validSSOMsg(t, linkageID)
	var att types.SSOAttestation
	mustJSONUnmarshal(msg.AttestationData, &att)
	att.Nonce = hex.EncodeToString(nonce)
	att.Proof.Nonce = att.Nonce
	att.OIDCNonce = oidcNonce
	f.resignSSOMsg(t, msg, &att)
	return msg
}

func (f *webEvidenceFixture) resignSSOMsg(t *testing.T, msg *types.MsgSubmitSSOVerificationProof, att *types.SSOAttestation) {
	t.Helper()
	canonical, err := att.CanonicalBytes()
	require.NoError(t, err)
	evidence := types.NewWebEvidenceContext(types.WebEvidenceContextConfig{
		ChainID:             webEvidenceChainID,
		AccountAddress:      msg.AccountAddress,
		EvidenceType:        types.AttestationTypeSSOVerification,
		Action:              types.WebEvidenceActionSubmitSSO,
		ScopeID:             msg.LinkageId,
		AttestationDigest:   types.WebEvidenceDigestHex(canonical),
		Issuer:              f.issuer,
		IssuerAlgorithm:     types.ProofTypeEd25519,
		Nonce:               att.Nonce,
		Challenge:           att.OIDCNonce,
		IssuedAt:            att.IssuedAt,
		ExpiresAt:           att.ExpiresAt,
		ServiceMetadataHash: serviceMetadataHashForTest(t, msg.EvidenceMetadata),
		CallerFields: map[string]string{
			"linkage_id":             msg.LinkageId,
			"provider":               string(att.ProviderType),
			"oidc_issuer":            att.OIDCIssuer,
			"subject_hash":           att.SubjectHash,
			"email_hash":             att.EmailHash,
			"email_domain_hash":      att.EmailDomainHash,
			"tenant_id_hash":         att.TenantIDHash,
			"oidc_nonce":             att.OIDCNonce,
			"email_verified":         strconv.FormatBool(att.EmailVerified),
			"linked_account_address": att.LinkedAccountAddress,
		},
	})
	require.NoError(t, evidence.ApplyToAttestation(&att.VerificationAttestation))
	signBytes, err := evidence.IssuerSignBytes()
	require.NoError(t, err)
	att.SetProof(types.NewAttestationProof(types.ProofTypeEd25519, att.IssuedAt, f.issuer.ID+"#"+f.issuer.KeyID, ed25519.Sign(f.issuerPriv, signBytes), att.Nonce))
	att.SetLinkageSignature(f.signAccount(t, evidence))
	attestationData, err := json.Marshal(att)
	require.NoError(t, err)
	msg.AttestationData = attestationData
	msg.EvidenceHash = hashTestEvidence(attestationData)
}

func (f *webEvidenceFixture) resignEmailMsgForTest(t *testing.T, msg *types.MsgSubmitEmailVerificationProof, att *types.VerificationAttestation) {
	t.Helper()
	evidence := f.webEvidenceForVerificationAttestation(t, att, types.AttestationTypeEmailVerification, types.WebEvidenceActionSubmitEmail, msg.VerificationId, msg.VerificationId, msg.EvidenceMetadata, map[string]string{
		"verification_id":   msg.VerificationId,
		"email_hash":        msg.EmailHash,
		"domain_hash":       msg.DomainHash,
		"nonce":             msg.Nonce,
		"is_organizational": strconv.FormatBool(msg.IsOrganizational),
		"verified_at_unix":  strconv.FormatInt(att.IssuedAt.Unix(), 10),
		"expires_at_unix":   strconv.FormatInt(att.ExpiresAt.Unix(), 10),
	})
	f.resignVerificationMsgForTest(t, att, evidence)
	msg.AttestationData = mustAttestationJSON(att)
	msg.AccountSignature = f.signAccount(t, evidence)
	msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
}

func (f *webEvidenceFixture) resignSMSMsgForTest(t *testing.T, msg *types.MsgSubmitSMSVerificationProof, att *types.VerificationAttestation) {
	t.Helper()
	evidence := f.webEvidenceForVerificationAttestation(t, att, types.AttestationTypeSMSVerification, types.WebEvidenceActionSubmitSMS, msg.VerificationId, msg.VerificationId, msg.EvidenceMetadata, map[string]string{
		"verification_id":   msg.VerificationId,
		"phone_hash":        msg.PhoneHash,
		"phone_hash_salt":   msg.PhoneHashSalt,
		"country_code_hash": msg.CountryCodeHash,
		"is_voip":           strconv.FormatBool(msg.IsVoip),
		"carrier_type":      msg.CarrierType,
		"validator_address": msg.ValidatorAddress,
		"verified_at_unix":  strconv.FormatInt(att.IssuedAt.Unix(), 10),
		"expires_at_unix":   strconv.FormatInt(att.ExpiresAt.Unix(), 10),
	})
	f.resignVerificationMsgForTest(t, att, evidence)
	msg.AttestationData = mustAttestationJSON(att)
	msg.AccountSignature = f.signAccount(t, evidence)
	msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
}

func (f *webEvidenceFixture) resignSocialMsgForTest(t *testing.T, msg *types.MsgSubmitSocialMediaScope, att *types.VerificationAttestation) {
	t.Helper()
	payloadDigest := encryptedPayloadDigestForTest(t, &msg.EncryptedPayload)
	provider := types.SocialMediaProviderFromProto(msg.Provider)
	evidence := f.webEvidenceForVerificationAttestation(t, att, types.AttestationTypeSocialMediaVerification, types.WebEvidenceActionSubmitSocial, msg.ScopeId, msg.ScopeId, msg.EvidenceMetadata, map[string]string{
		"scope_id":                 msg.ScopeId,
		"provider":                 string(provider),
		"profile_name_hash":        msg.ProfileNameHash,
		"email_hash":               msg.EmailHash,
		"username_hash":            msg.UsernameHash,
		"org_hash":                 msg.OrgHash,
		"account_created_at_unix":  strconv.FormatInt(msg.AccountCreatedAt, 10),
		"account_age_days":         strconv.FormatUint(uint64(msg.AccountAgeDays), 10),
		"is_verified":              strconv.FormatBool(msg.IsVerified),
		"friend_count_range":       msg.FriendCountRange,
		"encrypted_payload_digest": payloadDigest,
	})
	f.resignVerificationMsgForTest(t, att, evidence)
	msg.AttestationData = mustAttestationJSON(att)
	msg.AccountSignature = f.signAccount(t, evidence)
	msg.EvidenceHash = hashTestEvidence(msg.AttestationData)
}

func (f *webEvidenceFixture) webEvidenceForVerificationAttestation(
	t *testing.T,
	att *types.VerificationAttestation,
	attType types.AttestationType,
	action string,
	scopeID string,
	challenge string,
	metadata map[string]string,
	callerFields map[string]string,
) types.WebEvidenceContext {
	t.Helper()
	digest, err := types.WebEvidenceAttestationDigestHex(att)
	require.NoError(t, err)
	return types.NewWebEvidenceContext(types.WebEvidenceContextConfig{
		ChainID:             webEvidenceChainID,
		AccountAddress:      att.Subject.AccountAddress,
		EvidenceType:        attType,
		Action:              action,
		ScopeID:             scopeID,
		AttestationDigest:   digest,
		Issuer:              f.issuer,
		IssuerAlgorithm:     types.ProofTypeEd25519,
		Nonce:               att.Nonce,
		Challenge:           challenge,
		IssuedAt:            att.IssuedAt,
		ExpiresAt:           att.ExpiresAt,
		ServiceMetadataHash: serviceMetadataHashForTest(t, metadata),
		CallerFields:        callerFields,
	})
}

func (f *webEvidenceFixture) resignVerificationMsgForTest(t *testing.T, att *types.VerificationAttestation, evidence types.WebEvidenceContext) {
	t.Helper()
	require.NoError(t, evidence.ApplyToAttestation(att))
	signBytes, err := evidence.IssuerSignBytes()
	require.NoError(t, err)
	att.SetProof(types.NewAttestationProof(types.ProofTypeEd25519, att.IssuedAt, f.issuer.ID+"#"+f.issuer.KeyID, ed25519.Sign(f.issuerPriv, signBytes), att.Nonce))
}

func (f *webEvidenceFixture) signedVerificationAttestation(
	t *testing.T,
	attType types.AttestationType,
	action string,
	scopeID string,
	challenge string,
	nonce []byte,
	issuedAt time.Time,
	expiresAt time.Time,
	metadata map[string]string,
	callerFields map[string]string,
) (*types.VerificationAttestation, types.WebEvidenceContext) {
	t.Helper()
	return f.signedVerificationAttestationForAccount(t, webEvidenceChainID, f.accountAddress, attType, action, scopeID, challenge, nonce, issuedAt, expiresAt, metadata, callerFields)
}

func (f *webEvidenceFixture) signedVerificationAttestationForAccount(
	t *testing.T,
	chainID string,
	accountAddress string,
	attType types.AttestationType,
	action string,
	scopeID string,
	challenge string,
	nonce []byte,
	issuedAt time.Time,
	expiresAt time.Time,
	metadata map[string]string,
	callerFields map[string]string,
) (*types.VerificationAttestation, types.WebEvidenceContext) {
	t.Helper()
	att := types.NewVerificationAttestation(
		f.issuer,
		types.NewAttestationSubject(accountAddress),
		attType,
		nonce,
		issuedAt,
		expiresAt.Sub(issuedAt),
		100,
		100,
	)
	digest, err := types.WebEvidenceAttestationDigestHex(att)
	require.NoError(t, err)
	evidence := types.NewWebEvidenceContext(types.WebEvidenceContextConfig{
		ChainID:             chainID,
		AccountAddress:      accountAddress,
		EvidenceType:        attType,
		Action:              action,
		ScopeID:             scopeID,
		AttestationDigest:   digest,
		Issuer:              f.issuer,
		IssuerAlgorithm:     types.ProofTypeEd25519,
		Nonce:               att.Nonce,
		Challenge:           challenge,
		IssuedAt:            issuedAt,
		ExpiresAt:           expiresAt,
		ServiceMetadataHash: serviceMetadataHashForTest(t, metadata),
		CallerFields:        callerFields,
	})
	require.NoError(t, evidence.ApplyToAttestation(att))
	signBytes, err := evidence.IssuerSignBytes()
	require.NoError(t, err)
	att.SetProof(types.NewAttestationProof(types.ProofTypeEd25519, issuedAt, f.issuer.ID+"#"+f.issuer.KeyID, ed25519.Sign(f.issuerPriv, signBytes), att.Nonce))
	return att, evidence
}

func (f *webEvidenceFixture) signAccount(t *testing.T, evidence types.WebEvidenceContext) []byte {
	t.Helper()
	return signAccountFor(t, f.walletPriv, evidence)
}

func signAccountFor(t *testing.T, walletPriv ed25519.PrivateKey, evidence types.WebEvidenceContext) []byte {
	t.Helper()
	signBytes, err := evidence.AccountAuthorizationBytes()
	require.NoError(t, err)
	digest := sha256.Sum256(signBytes)
	return ed25519.Sign(walletPriv, digest[:])
}

func (f *webEvidenceFixture) createAdditionalWallet(t *testing.T, seed byte) (sdk.AccAddress, ed25519.PrivateKey) {
	t.Helper()
	account := sdk.AccAddress(bytesOf(seed, 20))
	walletPub, walletPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	walletID := keeper.GenerateWalletID(account.String())
	bindingSig := ed25519.Sign(walletPriv, types.GetWalletBindingMessage(walletID, account.String()))
	_, err = f.keeper.CreateWallet(f.ctx, account, bindingSig, walletPub)
	require.NoError(t, err)
	return account, walletPriv
}

func (f *webEvidenceFixture) storePrefixCount(prefix []byte) int {
	store := f.ctx.KVStore(f.keeper.StoreKey())
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	count := 0
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
}

func (f *webEvidenceFixture) deleteSignerKeyForTest() {
	store := f.ctx.KVStore(f.keeper.StoreKey())
	store.Delete(append(types.PrefixSignerKey, []byte(f.signerKey.KeyID)...))
	store.Delete(append(types.PrefixSignerKeyByFingerprint, []byte(f.signerKey.Fingerprint)...))
}

func (f *webEvidenceFixture) forceSignerKeyForTest(t *testing.T, key *types.SignerKeyInfo) {
	t.Helper()
	bz, err := json.Marshal(key)
	require.NoError(t, err)
	store := f.ctx.KVStore(f.keeper.StoreKey())
	store.Set(append(types.PrefixSignerKey, []byte(key.KeyID)...), bz)
	store.Set(append(types.PrefixSignerKeyByFingerprint, []byte(key.Fingerprint)...), []byte(key.KeyID))
}

func (f *webEvidenceFixture) deleteWalletForTest() {
	store := f.ctx.KVStore(f.keeper.StoreKey())
	store.Delete(types.IdentityWalletKey(f.account.Bytes()))
}

func createContextWithStore(t *testing.T, storeKey *storetypes.KVStoreKey) (sdk.Context, store.CommitMultiStore) {
	t.Helper()
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		ChainID: webEvidenceChainID,
		Time:    time.Unix(1710000000, 0).UTC(),
		Height:  100,
	}, false, log.NewNopLogger())
	return ctx, stateStore
}

func validEncryptedPayload(t *testing.T) types.EncryptedPayloadEnvelope {
	t.Helper()
	alg := encryptiontypes.DefaultAlgorithm()
	algInfo, err := encryptiontypes.GetAlgorithmInfo(alg)
	require.NoError(t, err)
	recipientPubKey := bytesOf(0x22, encryptiontypes.X25519PublicKeySize)
	recipientKeyID := encryptiontypes.ComputeKeyFingerprint(recipientPubKey)
	return types.EncryptedPayloadEnvelope{
		Version:             encryptiontypes.EnvelopeVersion,
		AlgorithmId:         alg,
		AlgorithmVersion:    algInfo.Version,
		RecipientKeyIds:     []string{recipientKeyID},
		RecipientPublicKeys: [][]byte{recipientPubKey},
		EncryptedKeys:       [][]byte{[]byte("encrypted-key")},
		Nonce:               bytesOf(0x33, algInfo.NonceSize),
		Ciphertext:          []byte("ciphertext"),
		SenderPubKey:        bytesOf(0x44, encryptiontypes.X25519PublicKeySize),
		SenderSignature:     []byte("sender-sig"),
	}
}

func serviceMetadataHashForTest(t *testing.T, metadata map[string]string) string {
	t.Helper()
	hash, err := types.WebEvidenceServiceMetadataHash(metadata)
	require.NoError(t, err)
	return hash
}

func encryptedPayloadDigestForTest(t *testing.T, payload *types.EncryptedPayloadEnvelope) string {
	t.Helper()
	bz, err := json.Marshal(encryptiontypes.EncryptedPayloadEnvelope{
		Version:             payload.Version,
		AlgorithmID:         payload.AlgorithmId,
		AlgorithmVersion:    payload.AlgorithmVersion,
		RecipientKeyIDs:     payload.RecipientKeyIds,
		RecipientPublicKeys: payload.RecipientPublicKeys,
		EncryptedKeys:       payload.EncryptedKeys,
		Nonce:               payload.Nonce,
		Ciphertext:          payload.Ciphertext,
		SenderSignature:     payload.SenderSignature,
		SenderPubKey:        payload.SenderPubKey,
		Metadata:            payload.Metadata,
	})
	require.NoError(t, err)
	return sha256HexBytes(bz)
}

func mustAttestationJSON(att *types.VerificationAttestation) []byte {
	bz, err := att.ToJSON()
	if err != nil {
		panic(err)
	}
	return bz
}

func mustJSONUnmarshal(bz []byte, out interface{}) {
	if err := json.Unmarshal(bz, out); err != nil {
		panic(err)
	}
}

func nonceBytes(seed string) []byte {
	sum := sha256.Sum256([]byte(seed))
	return sum[:]
}

func sha256Hex(value string) string {
	return sha256HexBytes([]byte(value))
}

func sha256HexBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func hashTestEvidence(data []byte) string {
	return sha256HexBytes(data)
}

func bytesOf(value byte, size int) []byte {
	out := make([]byte, size)
	for i := range out {
		out[i] = value
	}
	return out
}

func stringsJoin(values []string) string {
	out := ""
	for i, value := range values {
		if i > 0 {
			out += ","
		}
		out += value
	}
	return out
}

func cloneSignerKey(key *types.SignerKeyInfo) *types.SignerKeyInfo {
	clone := *key
	clone.PublicKey = append([]byte(nil), key.PublicKey...)
	if key.ActivatedAt != nil {
		activatedAt := *key.ActivatedAt
		clone.ActivatedAt = &activatedAt
	}
	if key.ExpiresAt != nil {
		expiresAt := *key.ExpiresAt
		clone.ExpiresAt = &expiresAt
	}
	if key.RevokedAt != nil {
		revokedAt := *key.RevokedAt
		clone.RevokedAt = &revokedAt
	}
	clone.Metadata = make(map[string]string, len(key.Metadata))
	for k, v := range key.Metadata {
		clone.Metadata[k] = v
	}
	return &clone
}

func (f *webEvidenceFixture) newPendingSignerKey(t *testing.T, sequence uint64) *types.SignerKeyInfo {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	key := types.NewSignerKeyInfo(f.signerKey.SignerID, pub, types.ProofTypeEd25519, sequence, f.ctx.BlockTime())
	key.Metadata[types.SignerKeyMetadataEvidenceTypes] = stringsJoin([]string{
		string(types.AttestationTypeSSOVerification),
		string(types.AttestationTypeEmailVerification),
		string(types.AttestationTypeSMSVerification),
		string(types.AttestationTypeSocialMediaVerification),
	})
	return key
}

func signerIndexKeyForTest(signerID string, keyID string) []byte {
	key := make([]byte, 0, len(types.PrefixSignerKeyBySigner)+len(signerID)+1+len(keyID))
	key = append(key, types.PrefixSignerKeyBySigner...)
	key = append(key, []byte(signerID)...)
	key = append(key, byte('/'))
	key = append(key, []byte(keyID)...)
	return key
}
