package veid

import (
	"encoding/hex"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	mfakeeper "github.com/virtengine/virtengine/x/mfa/keeper"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"

	"github.com/virtengine/virtengine/tests/integration/veid/fixtures"
)

func TestVEIDRegistrationVerificationAuthorizationFlow(t *testing.T) {
	env := setupVEIDTestEnv(t)
	ctx := env.ctx
	customer := newTestAccount(t)

	scopeIDs := []string{"scope-id-001", "scope-selfie-001"}
	scopeTypes := []veidtypes.ScopeType{veidtypes.ScopeTypeIDDocument, veidtypes.ScopeTypeSelfie}

	for i, scopeID := range scopeIDs {
		uploadScope(t, ctx, env.msgServer, env.client, customer, scopeID, scopeTypes[i])
		requireEventsEmitted(t, ctx.EventManager().Events())
		ctx = advanceContext(env.app, ctx, 2, 2*time.Minute)
	}

	for _, scopeID := range scopeIDs {
		_, err := env.msgServer.RequestVerification(ctx, &veidtypes.MsgRequestVerification{
			Sender:  customer.String(),
			ScopeId: scopeID,
		})
		require.NoError(t, err)
		requireEventsEmitted(t, ctx.EventManager().Events())
		ctx = advanceContext(env.app, ctx, 1, time.Minute)

		_, err = env.msgServer.UpdateVerificationStatus(ctx, &veidtypes.MsgUpdateVerificationStatus{
			Sender:         env.validator.String(),
			AccountAddress: customer.String(),
			ScopeId:        scopeID,
			NewStatus:      veidtypes.VerificationStatusPBVerified,
			Reason:         "integration verification",
		})
		require.NoError(t, err)
		requireEventsEmitted(t, ctx.EventManager().Events())
		ctx = advanceContext(env.app, ctx, 1, time.Minute)
	}

	record, found := env.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(t, found)
	requireScopeStatus(t, record, scopeIDs[0], veidtypes.VerificationStatusVerified)
	requireScopeStatus(t, record, scopeIDs[1], veidtypes.VerificationStatusVerified)
	require.NotNil(t, record.LastVerifiedAt)

	fixture := fixtures.DeterministicMLScoreFixture()
	ctx = ctx.WithBlockHeight(fixture.BlockHeight).
		WithBlockTime(fixture.RequestTime).
		WithEventManager(sdk.NewEventManager())
	computedScore, modelVersion, reasonCodes, inputHash, err := env.app.Keepers.VirtEngine.VEID.ComputeIdentityScore(
		ctx,
		fixture.AccountAddress,
		fixture.Scopes,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, fixture.ExpectedScore, computedScore)
	require.Equal(t, fixture.ExpectedModel, modelVersion)
	require.Equal(t, fixture.ExpectedInputHash, inputHash)
	require.Equal(t, fixture.ExpectedInputHex, hex.EncodeToString(inputHash))
	require.Contains(t, reasonCodes, veidtypes.ReasonCodeSuccess)

	tierMsg := veidtypes.NewMsgUpdateScore(
		env.validator.String(),
		customer.String(),
		computedScore,
		modelVersion,
	)
	_, err = env.msgServer.UpdateScore(ctx, tierMsg)
	require.NoError(t, err)
	requireEventsEmitted(t, ctx.EventManager().Events())

	ctx = advanceContext(env.app, ctx, 1, time.Minute)

	updated, found := env.app.Keepers.VirtEngine.VEID.GetIdentityRecord(ctx, customer)
	require.True(t, found)
	require.Equal(t, veidtypes.ComputeTierFromScore(computedScore), updated.Tier)

	mfaHooks := mfakeeper.NewMFAGatingHooks(env.app.Keepers.VirtEngine.MFA)
	mfaMsgServer := mfakeeper.NewMsgServerWithContext(env.app.Keepers.VirtEngine.MFA)

	config := mfatypes.SensitiveTxConfig{
		TransactionType: mfatypes.SensitiveTxHighValueOrder,
		Enabled:         true,
		MinVEIDScore:    0,
		RequiredFactorCombinations: []mfatypes.FactorCombination{
			{Factors: []mfatypes.FactorType{mfatypes.FactorTypeEmail}},
		},
		SessionDuration:             10 * 60,
		IsSingleUse:                 false,
		AllowTrustedDeviceReduction: false,
		Description:                 "integration high value order",
	}
	require.NoError(t, env.app.Keepers.VirtEngine.MFA.SetSensitiveTxConfig(ctx, &config))

	enrollResp, err := mfaMsgServer.EnrollFactor(ctx, &mfatypes.MsgEnrollFactor{
		Sender:           customer.String(),
		FactorType:       mfatypes.FactorTypeEmail,
		PublicIdentifier: []byte("customer@example.com"),
		Label:            "primary-email",
	})
	require.NoError(t, err)
	require.NotEmpty(t, enrollResp.FactorID)
	requireEventWithAttributes(t, ctx.EventManager().Events(), map[string]string{
		"account_address": customer.String(),
		"factor_type":     mfatypes.FactorTypeEmail.String(),
	})

	policy := mfatypes.MFAPolicy{
		AccountAddress: customer.String(),
		RequiredFactors: []mfatypes.FactorCombination{
			{Factors: []mfatypes.FactorType{mfatypes.FactorTypeEmail}},
		},
		SessionDuration: 10 * 60,
		VEIDThreshold:   0,
		Enabled:         true,
	}
	_, err = mfaMsgServer.SetMFAPolicy(ctx, &mfatypes.MsgSetMFAPolicy{
		Sender: customer.String(),
		Policy: policy,
	})
	require.NoError(t, err)

	clientInfo := &mfatypes.ClientInfo{
		DeviceFingerprint: "device-1",
		RequestedAt:       ctx.BlockTime().Unix(),
	}

	challengeResp, err := mfaMsgServer.CreateChallenge(ctx, &mfatypes.MsgCreateChallenge{
		Sender:          customer.String(),
		FactorType:      mfatypes.FactorTypeEmail,
		TransactionType: mfatypes.SensitiveTxHighValueOrder,
		ClientInfo:      clientInfo,
	})
	require.NoError(t, err)

	verifyResp, err := mfaMsgServer.VerifyChallenge(ctx, &mfatypes.MsgVerifyChallenge{
		Sender:      customer.String(),
		ChallengeID: challengeResp.ChallengeID,
		Response: &mfatypes.ChallengeResponse{
			ChallengeID:  challengeResp.ChallengeID,
			FactorType:   mfatypes.FactorTypeEmail,
			ResponseData: []byte("verified"),
			ClientInfo:   clientInfo,
			Timestamp:    ctx.BlockTime().Unix(),
		},
	})
	require.NoError(t, err)
	require.True(t, verifyResp.Verified)
	require.NotEmpty(t, verifyResp.SessionID)
	requireEventWithAttributes(t, ctx.EventManager().Events(), map[string]string{
		"session_id": verifyResp.SessionID,
	})

	proof := &mfatypes.MFAProof{
		SessionID:       verifyResp.SessionID,
		VerifiedFactors: []mfatypes.FactorType{mfatypes.FactorTypeEmail},
		Timestamp:       ctx.BlockTime().Unix(),
	}
	require.NoError(t, mfaHooks.ValidateMFAProof(ctx, customer, mfatypes.SensitiveTxHighValueOrder, proof, clientInfo.DeviceFingerprint))
	requireEventWithAttributes(t, ctx.EventManager().Events(), map[string]string{
		"session_id":       verifyResp.SessionID,
		"transaction_type": mfatypes.SensitiveTxHighValueOrder.String(),
	})

	_, err = hex.DecodeString(fixture.ExpectedInputHex)
	require.NoError(t, err)
}

func TestVEIDAuthorizationFailsOnInsufficientScore(t *testing.T) {
	env := setupVEIDTestEnv(t)
	ctx := env.ctx
	customer := newTestAccount(t)

	uploadScope(t, ctx, env.msgServer, env.client, customer, "scope-low-score", veidtypes.ScopeTypeSelfie)
	ctx = advanceContext(env.app, ctx, 1, time.Minute)

	_, err := env.msgServer.UpdateScore(ctx, veidtypes.NewMsgUpdateScore(
		env.validator.String(),
		customer.String(),
		40,
		"fixture-low-score",
	))
	require.NoError(t, err)
	ctx = advanceContext(env.app, ctx, 1, time.Minute)

	mfaMsgServer := mfakeeper.NewMsgServerWithContext(env.app.Keepers.VirtEngine.MFA)

	_, err = mfaMsgServer.EnrollFactor(ctx, &mfatypes.MsgEnrollFactor{
		Sender:     customer.String(),
		FactorType: mfatypes.FactorTypeVEID,
		Label:      "veid-score",
		Metadata: &mfatypes.FactorMetadata{
			VEIDThreshold: 70,
		},
	})
	require.NoError(t, err)

	policy := mfatypes.MFAPolicy{
		AccountAddress: customer.String(),
		RequiredFactors: []mfatypes.FactorCombination{
			{Factors: []mfatypes.FactorType{mfatypes.FactorTypeVEID}},
		},
		VEIDThreshold: 70,
		Enabled:       true,
	}
	_, err = mfaMsgServer.SetMFAPolicy(ctx, &mfatypes.MsgSetMFAPolicy{
		Sender: customer.String(),
		Policy: policy,
	})
	require.NoError(t, err)

	challengeResp, err := mfaMsgServer.CreateChallenge(ctx, &mfatypes.MsgCreateChallenge{
		Sender:          customer.String(),
		FactorType:      mfatypes.FactorTypeVEID,
		TransactionType: mfatypes.SensitiveTxHighValueOrder,
	})
	require.NoError(t, err)

	_, err = mfaMsgServer.VerifyChallenge(ctx, &mfatypes.MsgVerifyChallenge{
		Sender:      customer.String(),
		ChallengeID: challengeResp.ChallengeID,
		Response: &mfatypes.ChallengeResponse{
			ChallengeID:  challengeResp.ChallengeID,
			FactorType:   mfatypes.FactorTypeVEID,
			ResponseData: []byte("score-check"),
			Timestamp:    ctx.BlockTime().Unix(),
		},
	})
	require.Error(t, err)
	require.ErrorIs(t, err, mfatypes.ErrVEIDScoreInsufficient)
}

func TestVEIDAuthorizationFailsOnExpiredSession(t *testing.T) {
	env := setupVEIDTestEnv(t)
	ctx := env.ctx
	customer := newTestAccount(t)

	mfaMsgServer := mfakeeper.NewMsgServerWithContext(env.app.Keepers.VirtEngine.MFA)
	mfaHooks := mfakeeper.NewMFAGatingHooks(env.app.Keepers.VirtEngine.MFA)

	_, err := mfaMsgServer.EnrollFactor(ctx, &mfatypes.MsgEnrollFactor{
		Sender:           customer.String(),
		FactorType:       mfatypes.FactorTypeEmail,
		PublicIdentifier: []byte("customer@example.com"),
		Label:            "primary-email",
	})
	require.NoError(t, err)

	shortPolicy := mfatypes.MFAPolicy{
		AccountAddress: customer.String(),
		RequiredFactors: []mfatypes.FactorCombination{
			{Factors: []mfatypes.FactorType{mfatypes.FactorTypeEmail}},
		},
		SessionDuration: 2,
		Enabled:         true,
	}
	_, err = mfaMsgServer.SetMFAPolicy(ctx, &mfatypes.MsgSetMFAPolicy{
		Sender: customer.String(),
		Policy: shortPolicy,
	})
	require.NoError(t, err)

	clientInfo := &mfatypes.ClientInfo{
		DeviceFingerprint: "device-expired",
		RequestedAt:       ctx.BlockTime().Unix(),
	}

	challengeResp, err := mfaMsgServer.CreateChallenge(ctx, &mfatypes.MsgCreateChallenge{
		Sender:          customer.String(),
		FactorType:      mfatypes.FactorTypeEmail,
		TransactionType: mfatypes.SensitiveTxHighValueOrder,
		ClientInfo:      clientInfo,
	})
	require.NoError(t, err)

	verifyResp, err := mfaMsgServer.VerifyChallenge(ctx, &mfatypes.MsgVerifyChallenge{
		Sender:      customer.String(),
		ChallengeID: challengeResp.ChallengeID,
		Response: &mfatypes.ChallengeResponse{
			ChallengeID:  challengeResp.ChallengeID,
			FactorType:   mfatypes.FactorTypeEmail,
			ResponseData: []byte("verified"),
			ClientInfo:   clientInfo,
			Timestamp:    ctx.BlockTime().Unix(),
		},
	})
	require.NoError(t, err)

	proof := &mfatypes.MFAProof{
		SessionID:       verifyResp.SessionID,
		VerifiedFactors: []mfatypes.FactorType{mfatypes.FactorTypeEmail},
		Timestamp:       ctx.BlockTime().Unix(),
	}
	require.NoError(t, mfaHooks.ValidateMFAProof(ctx, customer, mfatypes.SensitiveTxHighValueOrder, proof, clientInfo.DeviceFingerprint))

	ctx = advanceContext(env.app, ctx, 1, 5*time.Second)

	err = mfaHooks.ValidateMFAProof(ctx, customer, mfatypes.SensitiveTxHighValueOrder, proof, clientInfo.DeviceFingerprint)
	require.Error(t, err)
	require.ErrorIs(t, err, mfatypes.ErrSessionExpired)

	env.app.Keepers.VirtEngine.MFA.CleanupExpiredSessions(ctx, customer)
	requireEventWithAttributes(t, ctx.EventManager().Events(), map[string]string{
		"session_id": verifyResp.SessionID,
	})
}
