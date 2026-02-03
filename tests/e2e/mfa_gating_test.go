//go:build e2e.integration

package e2e

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	aclient "github.com/virtengine/virtengine/sdk/go/node/client/discovery"
	cltypes "github.com/virtengine/virtengine/sdk/go/node/client/types"
	cclient "github.com/virtengine/virtengine/sdk/go/node/client/v1beta3"
	mfaquery "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
	mfatx "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
	rolespb "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
	rolesquery "github.com/virtengine/virtengine/sdk/go/node/roles/v1"

	"github.com/virtengine/virtengine/testutil"
	"github.com/virtengine/virtengine/testutil/network"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
)

type mfaGatingE2ETestSuite struct {
	*testutil.NetworkTestSuite
}

func TestIntegrationMFA(t *testing.T) {
	cfg := network.DefaultConfig(testutil.NewTestNetworkFixture, network.WithInterceptState(func(cdc codec.Codec, key string, raw json.RawMessage) json.RawMessage {
		if key != mfatypes.ModuleName {
			return raw
		}

		_ = cdc

		var state map[string]any
		if err := json.Unmarshal(raw, &state); err != nil {
			return raw
		}

		configs, ok := state["sensitive_tx_configs"].([]any)
		if !ok {
			configs, ok = state["sensitiveTxConfigs"].([]any)
		}
		if !ok {
			return raw
		}

		for _, cfgEntry := range configs {
			cfgMap, ok := cfgEntry.(map[string]any)
			if !ok {
				continue
			}

			txType := cfgMap["transaction_type"]
			if txType == nil {
				txType = cfgMap["transactionType"]
			}

			switch val := txType.(type) {
			case string:
				normalized := strings.ToLower(val)
				if strings.Contains(normalized, "validator") {
					cfgMap["enabled"] = false
				}
			case float64:
				if int(val) == int(mfaquery.SensitiveTxValidatorRegistration) {
					cfgMap["enabled"] = false
				}
			}
		}

		updated, err := json.Marshal(state)
		if err != nil {
			return raw
		}
		return updated
	}))
	cfg.NumValidators = 1

	mg := &mfaGatingE2ETestSuite{}
	mg.NetworkTestSuite = testutil.NewNetworkTestSuite(&cfg, mg)
	suite.Run(t, mg)
}

func (s *mfaGatingE2ETestSuite) TestMFAGatingFlow() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	val := s.Network().Validators[0]
	cctx := val.ClientCtx.WithFromAddress(val.Address).WithFromName("node0")

	cl, err := aclient.DiscoverClient(
		ctx,
		cctx,
		cltypes.WithGas(cltypes.GasSetting{Simulate: true}),
		cltypes.WithGasAdjustment(1.5),
		cltypes.WithGasPrices("0.0025uve"),
	)
	s.Require().NoError(err)

	conn, err := grpc.DialContext(ctx, val.AppConfig.GRPC.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	defer func() {
		_ = conn.Close()
	}()

	mfaClient := mfaquery.NewQueryClient(conn)
	rolesClient := rolesquery.NewQueryClient(conn)

	sender := val.Address.String()
	target := s.WalletForTest().String()
	deviceFingerprint := "device-e2e-mfa-1"

	// Enroll a TOTP factor.
	_, err = cl.Tx().BroadcastMsgs(
		ctx,
		[]sdk.Msg{
			&mfatx.MsgEnrollFactor{
				Sender:     sender,
				FactorType: mfaquery.FactorTypeTOTP,
				Label:      "e2e-totp",
			},
		},
		cclient.WithBroadcastMode("block"),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	// Enable MFA policy with a single required factor to keep the flow deterministic.
	policy := mfaquery.MFAPolicy{
		AccountAddress: sender,
		RequiredFactors: []mfaquery.FactorCombination{
			{Factors: []mfaquery.FactorType{mfaquery.FactorTypeTOTP}},
		},
		SessionDuration: 15 * 60,
		Enabled:         true,
	}

	_, err = cl.Tx().BroadcastMsgs(
		ctx,
		[]sdk.Msg{
			&mfatx.MsgSetMFAPolicy{
				Sender: sender,
				Policy: policy,
			},
		},
		cclient.WithBroadcastMode("block"),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	// Attempt a sensitive action without MFA proof; expect failure.
	_, err = cl.Tx().BroadcastMsgs(
		ctx,
		[]sdk.Msg{
			&rolespb.MsgSetAccountState{
				Sender:  sender,
				Address: target,
				State:   "suspended",
				Reason:  "mfa gating test without proof",
			},
		},
		cclient.WithBroadcastMode("block"),
	)
	s.Require().Error(err)

	// Create an MFA challenge for account recovery.
	_, err = cl.Tx().BroadcastMsgs(
		ctx,
		[]sdk.Msg{
			&mfatx.MsgCreateChallenge{
				Sender:          sender,
				FactorType:      mfaquery.FactorTypeTOTP,
				TransactionType: mfaquery.SensitiveTxAccountRecovery,
			},
		},
		cclient.WithBroadcastMode("block"),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	pending, err := mfaClient.PendingChallenges(ctx, &mfaquery.QueryPendingChallengesRequest{
		Address: sender,
	})
	s.Require().NoError(err)
	s.Require().NotEmpty(pending.Challenges)

	var challenge *mfaquery.Challenge
	for i := range pending.Challenges {
		if pending.Challenges[i].TransactionType == mfaquery.SensitiveTxAccountRecovery &&
			pending.Challenges[i].FactorType == mfaquery.FactorTypeTOTP {
			challenge = &pending.Challenges[i]
			break
		}
	}
	s.Require().NotNil(challenge)
	s.Require().Equal(mfaquery.ChallengeStatusPending, challenge.Status)

	// Complete the challenge.
	_, err = cl.Tx().BroadcastMsgs(
		ctx,
		[]sdk.Msg{
			&mfatx.MsgVerifyChallenge{
				Sender:      sender,
				ChallengeId: challenge.ChallengeId,
				Response: mfatx.ChallengeResponse{
					ChallengeId:  challenge.ChallengeId,
					FactorType:   mfaquery.FactorTypeTOTP,
					ResponseData: []byte("e2e-ok"),
					ClientInfo: &mfatx.ClientInfo{
						DeviceFingerprint: deviceFingerprint,
						RequestedAt:       time.Now().Unix(),
					},
					Timestamp: time.Now().Unix(),
				},
			},
		},
		cclient.WithBroadcastMode("block"),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	verified, err := mfaClient.Challenge(ctx, &mfaquery.QueryChallengeRequest{
		ChallengeId: challenge.ChallengeId,
	})
	s.Require().NoError(err)
	s.Require().Equal(mfaquery.ChallengeStatusVerified, verified.Challenge.Status)
	s.Require().NotEmpty(verified.Challenge.SessionId)

	sessionResp, err := mfaClient.AuthorizationSession(ctx, &mfaquery.QueryAuthorizationSessionRequest{
		SessionId: verified.Challenge.SessionId,
	})
	s.Require().NoError(err)
	s.Require().Equal(mfaquery.SensitiveTxAccountRecovery, sessionResp.Session.TransactionType)

	// Execute the sensitive action with MFA proof.
	_, err = cl.Tx().BroadcastMsgs(
		ctx,
		[]sdk.Msg{
			&rolespb.MsgSetAccountState{
				Sender:  sender,
				Address: target,
				State:   "suspended",
				Reason:  "mfa gating test with proof",
				MfaProof: &mfatx.MFAProof{
					SessionId:       verified.Challenge.SessionId,
					VerifiedFactors: []mfaquery.FactorType{mfaquery.FactorTypeTOTP},
					Timestamp:       time.Now().Unix(),
				},
				DeviceFingerprint: deviceFingerprint,
			},
		},
		cclient.WithBroadcastMode("block"),
	)
	s.Require().NoError(err)
	s.Require().NoError(s.Network().WaitForNextBlock())

	accountState, err := rolesClient.AccountState(ctx, &rolesquery.QueryAccountStateRequest{
		Address: target,
	})
	s.Require().NoError(err)
	s.Require().Equal(rolesquery.AccountStateSuspended, accountState.AccountState.State)
}
