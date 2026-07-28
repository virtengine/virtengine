// Copyright 2024-2026 VirtEngine Authors
// SPDX-License-Identifier: Apache-2.0

package veid

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	coreheader "cosmossdk.io/core/header"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

const (
	task85DHelperProcessEnv = "VEID_TASK85D_HELPER_PROCESS"
	task85DWebHelper        = "web-issuer"
	task85DInferenceHelper  = "inference-runtime"
	task85DHelperMaxBytes   = 64 * 1024
	task85DHelperTimeout    = 10 * time.Second

	task85DChainID = "virtengine-integration-1"

	task85DWebProfile       = "web-sandbox"
	task85DInferenceProfile = "engineering-only deterministic sandbox runtime"
	task85DInferenceGolden  = `{"account_age_days":365,"email_verified":true,"sms_mobile_verified":true,"social_verified":true,"org_domain_verified":true,"risk_signals":[]}`

	task85DKindSSO    = "sso"
	task85DKindEmail  = "email"
	task85DKindSMS    = "sms"
	task85DKindSocial = "social"
)

type task85DHelperRequest struct {
	Helper        string            `json:"helper"`
	Action        string            `json:"action"`
	Profile       string            `json:"profile"`
	Sequence      uint64            `json:"sequence"`
	ChainID       string            `json:"chain_id,omitempty"`
	Account       string            `json:"account,omitempty"`
	Kind          string            `json:"kind,omitempty"`
	EvidenceID    string            `json:"evidence_id,omitempty"`
	NonceHex      string            `json:"nonce_hex,omitempty"`
	IssuedUnix    int64             `json:"issued_unix,omitempty"`
	ExpiresUnix   int64             `json:"expires_unix,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	PayloadDigest string            `json:"payload_digest,omitempty"`
	GoldenInput   string            `json:"golden_input,omitempty"`
}

type task85DWebIssuerResponse struct {
	Issuer          veidtypes.AttestationIssuer `json:"issuer"`
	PublicKey       []byte                      `json:"public_key"`
	AttestationData []byte                      `json:"attestation_data,omitempty"`
	EvidenceHash    string                      `json:"evidence_hash,omitempty"`

	EmailHash        string `json:"email_hash,omitempty"`
	DomainHash       string `json:"domain_hash,omitempty"`
	PhoneHash        string `json:"phone_hash,omitempty"`
	PhoneHashSalt    string `json:"phone_hash_salt,omitempty"`
	CountryCodeHash  string `json:"country_code_hash,omitempty"`
	CarrierType      string `json:"carrier_type,omitempty"`
	ProfileNameHash  string `json:"profile_name_hash,omitempty"`
	SocialEmailHash  string `json:"social_email_hash,omitempty"`
	UsernameHash     string `json:"username_hash,omitempty"`
	OrgHash          string `json:"org_hash,omitempty"`
	FriendCountRange string `json:"friend_count_range,omitempty"`
	AccountCreatedAt int64  `json:"account_created_at,omitempty"`
	AccountAgeDays   uint32 `json:"account_age_days,omitempty"`
	IsOrganizational bool   `json:"is_organizational,omitempty"`
	IsVoip           bool   `json:"is_voip,omitempty"`
	IsVerified       bool   `json:"is_verified,omitempty"`
	VerifiedAt       int64  `json:"verified_at,omitempty"`
	ExpiresAt        int64  `json:"expires_at,omitempty"`
	ValidatorAddress string `json:"validator_address,omitempty"`
	OIDCIssuer       string `json:"oidc_issuer,omitempty"`
	OIDCNonce        string `json:"oidc_nonce,omitempty"`
	Provider         string `json:"provider,omitempty"`
	LinkedAccount    string `json:"linked_account,omitempty"`
	SubjectHash      string `json:"subject_hash,omitempty"`
	SSOEmailHash     string `json:"sso_email_hash,omitempty"`
	SSOEmailDomain   string `json:"sso_email_domain_hash,omitempty"`
}

type task85DInferenceReceiptResponse struct {
	RuntimeName         string   `json:"runtime_name"`
	PublicKey           []byte   `json:"public_key"`
	ReceiptBytes        []byte   `json:"receipt_bytes"`
	Score               uint32   `json:"score"`
	GoldenInput         string   `json:"golden_input"`
	InputDigest         string   `json:"input_digest"`
	FeatureDigest       string   `json:"feature_digest"`
	OutputDigest        string   `json:"output_digest"`
	EvidenceLineage     string   `json:"evidence_lineage_digest"`
	ModelManifestDigest string   `json:"model_manifest_digest"`
	ModelDigest         string   `json:"model_digest"`
	RuntimeImageDigest  string   `json:"runtime_image_digest"`
	RuntimeDigest       string   `json:"runtime_digest"`
	SchemaDigest        string   `json:"schema_digest"`
	ConfigDigest        string   `json:"config_digest"`
	ReasonCodes         []string `json:"reason_codes"`
}

type task85DInferenceGoldenInput struct {
	AccountAgeDays    uint32   `json:"account_age_days"`
	EmailVerified     bool     `json:"email_verified"`
	SMSMobileVerified bool     `json:"sms_mobile_verified"`
	SocialVerified    bool     `json:"social_verified"`
	OrgDomainVerified bool     `json:"org_domain_verified"`
	RiskSignals       []string `json:"risk_signals"`
}

type task85DStoredWebEvidence struct {
	Resp        task85DWebIssuerResponse
	Hash        string
	Backend     string
	Ref         string
	Metadata    map[string]string
	Attestation []byte
}

func TestTask85DHelperProcess(t *testing.T) {
	if os.Getenv(task85DHelperProcessEnv) != "1" {
		return
	}
	req, err := task85DDecodeHelperRequest(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode helper request: %v\n", err)
		os.Exit(2)
	}
	var (
		resp any
	)
	switch req.Helper {
	case task85DWebHelper:
		resp, err = task85DRunWebIssuerHelper(req)
	case task85DInferenceHelper:
		resp, err = task85DRunInferenceRuntimeHelper(req)
	default:
		err = fmt.Errorf("unknown helper %q", req.Helper)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(3)
	}
	if err := task85DEncodeHelperResponse(os.Stdout, resp); err != nil {
		fmt.Fprintf(os.Stderr, "encode helper response: %v\n", err)
		os.Exit(4)
	}
	os.Exit(0)
}

func TestTask85DHelperProtocolRejectsMalformedRequests(t *testing.T) {
	validWeb := `{"helper":"web-issuer","action":"register","profile":"web-sandbox","sequence":1}`
	validInference := `{"helper":"inference-runtime","action":"issue","profile":"engineering-only deterministic sandbox runtime","sequence":1,"chain_id":"virtengine-integration-1","account":"` + task85DAccountAddress("protocol").String() + `","evidence_id":"protocol","golden_input":` + strconv.Quote(task85DInferenceGolden) + `}`
	tests := []struct {
		name    string
		input   []byte
		wantErr string
	}{
		{name: "unknown-field", input: []byte(`{"helper":"web-issuer","action":"register","profile":"web-sandbox","unexpected":true}`), wantErr: "unknown field"},
		{name: "trailing-json", input: []byte(validWeb + ` {}`), wantErr: "trailing JSON"},
		{name: "oversized-input", input: bytes.Repeat([]byte("x"), task85DHelperMaxBytes+1), wantErr: "exceeds"},
		{name: "unsupported-action", input: []byte(`{"helper":"web-issuer","action":"delete","profile":"web-sandbox"}`), wantErr: "unsupported"},
		{name: "unsupported-profile", input: []byte(strings.Replace(validInference, task85DInferenceProfile, "unsupported-runtime", 1)), wantErr: "unsupported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr, err := task85DRunHelperRaw(t, tt.input)
			require.Error(t, err)
			require.Contains(t, stderr, tt.wantErr)
		})
	}
}

func task85DDecodeHelperRequest(r io.Reader) (task85DHelperRequest, error) {
	input, err := io.ReadAll(io.LimitReader(r, task85DHelperMaxBytes+1))
	if err != nil {
		return task85DHelperRequest{}, err
	}
	if len(input) > task85DHelperMaxBytes {
		return task85DHelperRequest{}, fmt.Errorf("helper request exceeds %d bytes", task85DHelperMaxBytes)
	}
	dec := json.NewDecoder(bytes.NewReader(input))
	dec.DisallowUnknownFields()
	var req task85DHelperRequest
	if err := dec.Decode(&req); err != nil {
		return task85DHelperRequest{}, err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return task85DHelperRequest{}, fmt.Errorf("trailing JSON after helper request")
		}
		return task85DHelperRequest{}, fmt.Errorf("trailing JSON after helper request: %w", err)
	}
	if req.Helper == "" {
		return task85DHelperRequest{}, fmt.Errorf("helper is required")
	}
	if req.Action == "" {
		return task85DHelperRequest{}, fmt.Errorf("action is required")
	}
	if req.Profile == "" {
		return task85DHelperRequest{}, fmt.Errorf("profile is required")
	}
	return req, nil
}

func task85DEncodeHelperResponse(w io.Writer, resp any) error {
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(resp); err != nil {
		return err
	}
	if out.Len() > task85DHelperMaxBytes {
		return fmt.Errorf("helper response exceeds %d bytes", task85DHelperMaxBytes)
	}
	_, err := w.Write(out.Bytes())
	return err
}

func task85DRunHelperRaw(t *testing.T, reqBytes []byte) (string, string, error) {
	t.Helper()
	require.LessOrEqual(t, len(reqBytes), task85DHelperMaxBytes+1)
	ctx, cancel := context.WithTimeout(context.Background(), task85DHelperTimeout)
	defer cancel()
	// #nosec G204 -- this intentionally re-execs the current Go test binary as a bounded helper process.
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestTask85DHelperProcess$")
	cmd.Env = append(os.Environ(), task85DHelperProcessEnv+"=1")
	cmd.Stdin = bytes.NewReader(reqBytes)
	dir := t.TempDir()
	stdoutPath := filepath.Join(dir, "stdout.json")
	stderrPath := filepath.Join(dir, "stderr.txt")
	stdoutFile, err := os.Create(stdoutPath)
	require.NoError(t, err)
	stderrFile, err := os.Create(stderrPath)
	require.NoError(t, err)
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	runErr := cmd.Run()
	require.NoError(t, stdoutFile.Close())
	require.NoError(t, stderrFile.Close())
	stdout := task85DReadBoundedFile(t, stdoutPath)
	stderr := task85DReadBoundedFile(t, stderrPath)
	if ctx.Err() != nil {
		return string(stdout), string(stderr), ctx.Err()
	}
	return string(stdout), string(stderr), runErr
}

func task85DReadBoundedFile(t *testing.T, path string) []byte {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, task85DHelperMaxBytes+1))
	require.NoError(t, err)
	require.LessOrEqual(t, len(data), task85DHelperMaxBytes)
	return data
}

func task85DDecodeHelperResponse[T any](t *testing.T, stdout string, target *T) {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(stdout))
	dec.DisallowUnknownFields()
	require.NoError(t, dec.Decode(target))
	require.ErrorIs(t, dec.Decode(&struct{}{}), io.EOF)
}

func TestTask85DWebEvidenceIssuerProcessBoundaryConformance(t *testing.T) {
	env := setupVEIDTestEnv(t)
	env.ctx = task85DChainContext(env.ctx)
	ctx := env.ctx
	account, accountPriv := task85DCreateWallet(t, ctx, env.msgServer, env.app.Keepers.VirtEngine.VEID, "task85d-account")
	task85DSeedSocialName(t, ctx, env.app.Keepers.VirtEngine.VEID, account)

	reg := task85DWebIssuerRegister(t, task85DWebProfile, 1)
	task85DRegisterWebSigner(t, ctx, env.app.Keepers.VirtEngine.VEID, reg, veidtypes.SignerKeyStateActive)

	srv := env.msgServer
	issued := make(map[string]task85DStoredWebEvidence)
	submitters := map[string]func(string) (any, error){
		task85DKindSSO: func(id string) (any, error) {
			resp := task85DWebIssuerIssue(t, task85DWebIssueConfig{
				Profile: task85DWebProfile, Sequence: 1, ChainID: ctx.ChainID(), Account: account.String(),
				Kind: task85DKindSSO, EvidenceID: id, IssuedAt: ctx.BlockTime(), ExpiresAt: ctx.BlockTime().Add(time.Hour),
				Metadata: map[string]string{"source": "task85d-sso-sandbox"},
			})
			msg := task85DSSOMsg(t, ctx, resp, account.String(), accountPriv, id, map[string]string{"source": "task85d-sso-sandbox"})
			issued[id] = task85DStoredWebEvidence{Resp: resp, Hash: msg.EvidenceHash, Backend: msg.EvidenceStorageBackend, Ref: msg.EvidenceStorageRef, Metadata: msg.EvidenceMetadata, Attestation: msg.AttestationData}
			return srv.SubmitSSOVerificationProof(ctx, msg)
		},
		task85DKindEmail: func(id string) (any, error) {
			resp := task85DWebIssuerIssue(t, task85DWebIssueConfig{
				Profile: task85DWebProfile, Sequence: 1, ChainID: ctx.ChainID(), Account: account.String(),
				Kind: task85DKindEmail, EvidenceID: id, IssuedAt: ctx.BlockTime(), ExpiresAt: ctx.BlockTime().Add(time.Hour),
				Metadata: map[string]string{"source": "task85d-email-sandbox"},
			})
			msg := task85DEmailMsg(t, ctx, resp, account.String(), accountPriv, id, map[string]string{"source": "task85d-email-sandbox"})
			issued[id] = task85DStoredWebEvidence{Resp: resp, Hash: msg.EvidenceHash, Backend: msg.EvidenceStorageBackend, Ref: msg.EvidenceStorageRef, Metadata: msg.EvidenceMetadata, Attestation: msg.AttestationData}
			return srv.SubmitEmailVerificationProof(ctx, msg)
		},
		task85DKindSMS: func(id string) (any, error) {
			resp := task85DWebIssuerIssue(t, task85DWebIssueConfig{
				Profile: task85DWebProfile, Sequence: 1, ChainID: ctx.ChainID(), Account: account.String(),
				Kind: task85DKindSMS, EvidenceID: id, IssuedAt: ctx.BlockTime(), ExpiresAt: ctx.BlockTime().Add(time.Hour),
				Metadata: map[string]string{"source": "task85d-sms-sandbox"},
			})
			msg := task85DSMSMsg(t, ctx, resp, account.String(), accountPriv, id, map[string]string{"source": "task85d-sms-sandbox"})
			issued[id] = task85DStoredWebEvidence{Resp: resp, Hash: msg.EvidenceHash, Backend: msg.EvidenceStorageBackend, Ref: msg.EvidenceStorageRef, Metadata: msg.EvidenceMetadata, Attestation: msg.AttestationData}
			return srv.SubmitSMSVerificationProof(ctx, msg)
		},
		task85DKindSocial: func(id string) (any, error) {
			payload := task85DEncryptedPayload(t)
			resp := task85DWebIssuerIssue(t, task85DWebIssueConfig{
				Profile: task85DWebProfile, Sequence: 1, ChainID: ctx.ChainID(), Account: account.String(),
				Kind: task85DKindSocial, EvidenceID: id, IssuedAt: ctx.BlockTime(), ExpiresAt: ctx.BlockTime().Add(time.Hour),
				Metadata: map[string]string{"source": "task85d-social-sandbox"}, PayloadDigest: task85DEncryptedPayloadDigest(t, payload),
			})
			msg := task85DSocialMsg(t, ctx, resp, account.String(), accountPriv, id, payload, map[string]string{"source": "task85d-social-sandbox"})
			issued[id] = task85DStoredWebEvidence{Resp: resp, Hash: msg.EvidenceHash, Backend: msg.EvidenceStorageBackend, Ref: msg.EvidenceStorageRef, Metadata: msg.EvidenceMetadata, Attestation: msg.AttestationData}
			return srv.SubmitSocialMediaScope(ctx, msg)
		},
	}

	for _, kind := range []string{task85DKindSSO, task85DKindEmail, task85DKindSMS, task85DKindSocial} {
		submit := submitters[kind]
		for round := 1; round <= 3; round++ {
			if kind == task85DKindSSO && round > 1 {
				continue
			}
			t.Run(fmt.Sprintf("%s-round-%d", kind, round), func(t *testing.T) {
				id := fmt.Sprintf("task85d-%s-success-%d", kind, round)
				first, err := submit(id)
				require.NoError(t, err)
				beforeRetry := task85DVEIDStoreDigest(t, ctx, env.app.Keepers.VirtEngine.VEID)
				second, err := submit(id)
				require.NoError(t, err)
				require.Equal(t, first, second)
				require.Equal(t, beforeRetry, task85DVEIDStoreDigest(t, ctx, env.app.Keepers.VirtEngine.VEID))
			})
		}
	}

	for _, kind := range []string{task85DKindSSO, task85DKindEmail, task85DKindSMS, task85DKindSocial} {
		t.Run(kind+"-final-idempotent-retry", func(t *testing.T) {
			submit := submitters[kind]
			round := 3
			if kind == task85DKindSSO {
				round = 1
			}
			id := fmt.Sprintf("task85d-%s-success-%d", kind, round)
			first, err := submit(id)
			require.NoError(t, err)
			beforeRetry := task85DVEIDStoreDigest(t, ctx, env.app.Keepers.VirtEngine.VEID)
			second, err := submit(id)
			require.NoError(t, err)
			require.Equal(t, first, second)
			require.Equal(t, beforeRetry, task85DVEIDStoreDigest(t, ctx, env.app.Keepers.VirtEngine.VEID))
		})
	}

	task85DAssertStoredWebEvidence(t, ctx, env.app.Keepers.VirtEngine.VEID, account.String(), issued)
	score, status, found := env.app.Keepers.VirtEngine.VEID.GetScore(ctx, account.String())
	require.True(t, found)
	require.Greater(t, score, uint32(0))
	require.Equal(t, veidtypes.AccountStatusVerified, status)
	history := env.app.Keepers.VirtEngine.VEID.GetScoreHistory(ctx, account.String())
	require.NotEmpty(t, history)
	require.Equal(t, score, history[0].Score)
	require.Equal(t, status, history[0].Status)
	require.Equal(t, "web-scope-v1", history[0].ModelVersion)
	require.Equal(t, "web-scope evidence update", history[0].Reason)
	wallet, found := env.app.Keepers.VirtEngine.VEID.GetWallet(ctx, account)
	require.True(t, found)
	require.Equal(t, score, wallet.CurrentScore)
	require.Equal(t, status, wallet.ScoreStatus)
	require.Equal(t, veidtypes.TierFromScore(score, status), wallet.Tier)
	require.NotEmpty(t, wallet.VerificationHistory)
	require.Equal(t, score, wallet.VerificationHistory[len(wallet.VerificationHistory)-1].NewScore)
	require.Equal(t, status, wallet.VerificationHistory[len(wallet.VerificationHistory)-1].NewStatus)
}

func task85DAssertStoredWebEvidence(t *testing.T, ctx sdk.Context, k keeper.Keeper, account string, issued map[string]task85DStoredWebEvidence) {
	t.Helper()
	assertPointer := func(expected task85DStoredWebEvidence, hash, backend, ref string, metadata map[string]string) {
		t.Helper()
		require.Equal(t, expected.Hash, hash)
		require.Equal(t, expected.Backend, backend)
		require.Equal(t, expected.Ref, ref)
		require.Equal(t, expected.Metadata, metadata)
		require.Equal(t, task85DSHA256HexBytes(expected.Attestation), hash)
	}
	assertIssuerLineage := func(attIssuer veidtypes.AttestationIssuer) {
		t.Helper()
		require.Equal(t, "did:virtengine:issuer:"+task85DWebProfile, attIssuer.ID)
		require.Equal(t, "did:virtengine:issuer:"+task85DWebProfile+":1", attIssuer.KeyID)
		require.Equal(t, issued["task85d-sso-success-1"].Resp.Issuer.KeyFingerprint, attIssuer.KeyFingerprint)
	}

	ssoExpected := issued["task85d-sso-success-1"]
	var ssoAtt veidtypes.SSOAttestation
	require.NoError(t, json.Unmarshal(ssoExpected.Attestation, &ssoAtt))
	linkage, found := k.GetSSOLinkage(ctx, "task85d-sso-success-1")
	require.True(t, found)
	require.Equal(t, veidtypes.SSOStatusVerified, linkage.Status)
	require.Equal(t, account, ssoAtt.LinkedAccountAddress)
	require.Equal(t, ssoExpected.Resp.Provider, string(linkage.Provider))
	require.Equal(t, ssoExpected.Resp.OIDCIssuer, linkage.Issuer)
	require.Equal(t, ssoExpected.Resp.SubjectHash, linkage.SubjectHash)
	require.Equal(t, ssoExpected.Resp.SSOEmailDomain, linkage.EmailDomainHash)
	assertPointer(ssoExpected, linkage.EvidenceHash, linkage.EvidenceStorageBackend, linkage.EvidenceStorageRef, linkage.EvidenceMetadata)
	assertIssuerLineage(ssoAtt.Issuer)
	require.Equal(t, "task85d-sso-success-1", k.GetSSOLinkageByAccountAndProvider(ctx, account, veidtypes.SSOProviderGoogle))

	emailExpected := issued["task85d-email-success-1"]
	emailAtt := task85DVerificationAttestation(t, emailExpected.Attestation)
	emailRecord, found := k.GetEmailVerificationRecord(ctx, "task85d-email-success-1")
	require.True(t, found)
	require.Equal(t, veidtypes.EmailStatusVerified, emailRecord.Status)
	require.Equal(t, account, emailRecord.AccountAddress)
	require.Equal(t, emailExpected.Resp.EmailHash, emailRecord.EmailHash)
	require.Equal(t, emailExpected.Resp.DomainHash, emailRecord.DomainHash)
	require.Equal(t, emailExpected.Resp.IsOrganizational, emailRecord.IsOrganizational)
	assertPointer(emailExpected, emailRecord.EvidenceHash, emailRecord.EvidenceStorageBackend, emailRecord.EvidenceStorageRef, emailRecord.EvidenceMetadata)
	assertIssuerLineage(emailAtt.Issuer)

	smsExpected := issued["task85d-sms-success-1"]
	smsAtt := task85DVerificationAttestation(t, smsExpected.Attestation)
	smsRecord, found := k.GetSMSVerificationRecord(ctx, "task85d-sms-success-1")
	require.True(t, found)
	require.Equal(t, veidtypes.SMSStatusVerified, smsRecord.Status)
	require.Equal(t, account, smsRecord.AccountAddress)
	require.Equal(t, smsExpected.Resp.PhoneHash, smsRecord.PhoneHash.Hash)
	require.Equal(t, smsExpected.Resp.PhoneHashSalt, smsRecord.PhoneHash.Salt)
	require.Equal(t, smsExpected.Resp.CountryCodeHash, smsRecord.PhoneHash.CountryCodeHash)
	require.Equal(t, smsExpected.Resp.CarrierType, smsRecord.CarrierType)
	require.Equal(t, smsExpected.Resp.ValidatorAddress, smsRecord.ValidatorAddress)
	assertPointer(smsExpected, smsRecord.EvidenceHash, smsRecord.EvidenceStorageBackend, smsRecord.EvidenceStorageRef, smsRecord.EvidenceMetadata)
	assertIssuerLineage(smsAtt.Issuer)

	socialExpected := issued["task85d-social-success-1"]
	socialAtt := task85DVerificationAttestation(t, socialExpected.Attestation)
	socialScope, found := k.GetSocialMediaScope(ctx, "task85d-social-success-1")
	require.True(t, found)
	require.Equal(t, veidtypes.SocialMediaStatusVerified, socialScope.Status)
	require.Equal(t, account, socialScope.AccountAddress)
	require.Equal(t, veidtypes.SocialMediaProviderGoogle, socialScope.Provider)
	require.Equal(t, socialExpected.Resp.ProfileNameHash, socialScope.ProfileNameHash)
	require.Equal(t, socialExpected.Resp.SocialEmailHash, socialScope.EmailHash)
	require.Equal(t, socialExpected.Resp.UsernameHash, socialScope.UsernameHash)
	require.Equal(t, socialExpected.Resp.OrgHash, socialScope.OrgHash)
	require.Equal(t, socialExpected.Resp.IsVerified, socialScope.IsVerified)
	require.Equal(t, socialExpected.Resp.FriendCountRange, socialScope.FriendCountRange)
	require.Equal(t, socialExpected.Resp.AccountAgeDays, socialScope.AccountAgeDays)
	assertPointer(socialExpected, socialScope.EvidenceHash, socialScope.EvidenceStorageBackend, socialScope.EvidenceStorageRef, socialScope.EvidenceMetadata)
	assertIssuerLineage(socialAtt.Issuer)
}

func TestTask85DWebEvidenceProcessBoundaryNegativeCasesDoNotMutate(t *testing.T) {
	tests := []struct {
		name  string
		build func(t *testing.T, env veidTestEnv, account sdk.AccAddress, accountPriv ed25519.PrivateKey) (*veidtypes.MsgSubmitEmailVerificationProof, sdk.Context)
		setup func(t *testing.T, env veidTestEnv, ctx sdk.Context, reg task85DWebIssuerResponse)
	}{
		{
			name: "wrong-account",
			build: func(t *testing.T, env veidTestEnv, account sdk.AccAddress, accountPriv ed25519.PrivateKey) (*veidtypes.MsgSubmitEmailVerificationProof, sdk.Context) {
				other := task85DAccountAddress("wrong-account-other")
				resp := task85DWebIssuerIssue(t, task85DDefaultEmailIssue(ctxFromEnv(env), other.String(), "email-wrong-account"))
				msg := task85DEmailMsg(t, env.ctx, resp, account.String(), accountPriv, "email-wrong-account", map[string]string{"source": "task85d-email-sandbox"})
				return msg, env.ctx
			},
		},
		{
			name: "wrong-chain",
			build: func(t *testing.T, env veidTestEnv, account sdk.AccAddress, accountPriv ed25519.PrivateKey) (*veidtypes.MsgSubmitEmailVerificationProof, sdk.Context) {
				cfg := task85DDefaultEmailIssue(ctxFromEnv(env), account.String(), "email-wrong-chain")
				cfg.ChainID = "virtengine-wrong-chain"
				resp := task85DWebIssuerIssue(t, cfg)
				msg := task85DEmailMsg(t, env.ctx, resp, account.String(), accountPriv, "email-wrong-chain", map[string]string{"source": "task85d-email-sandbox"})
				return msg, env.ctx
			},
		},
		{
			name: "wrong-scope",
			build: func(t *testing.T, env veidTestEnv, account sdk.AccAddress, accountPriv ed25519.PrivateKey) (*veidtypes.MsgSubmitEmailVerificationProof, sdk.Context) {
				resp := task85DWebIssuerIssue(t, task85DDefaultEmailIssue(ctxFromEnv(env), account.String(), "email-original-scope"))
				msg := task85DEmailMsg(t, env.ctx, resp, account.String(), accountPriv, "email-original-scope", map[string]string{"source": "task85d-email-sandbox"})
				msg.VerificationId = "email-changed-scope"
				return msg, env.ctx
			},
		},
		{
			name: "wrong-type",
			build: func(t *testing.T, env veidTestEnv, account sdk.AccAddress, accountPriv ed25519.PrivateKey) (*veidtypes.MsgSubmitEmailVerificationProof, sdk.Context) {
				cfg := task85DDefaultEmailIssue(ctxFromEnv(env), account.String(), "email-wrong-type")
				cfg.Kind = task85DKindSMS
				resp := task85DWebIssuerIssue(t, cfg)
				msg := task85DEmailMsg(t, env.ctx, resp, account.String(), accountPriv, "email-wrong-type", map[string]string{"source": "task85d-email-sandbox"})
				return msg, env.ctx
			},
		},
		{
			name: "changed-payload",
			build: func(t *testing.T, env veidTestEnv, account sdk.AccAddress, accountPriv ed25519.PrivateKey) (*veidtypes.MsgSubmitEmailVerificationProof, sdk.Context) {
				resp := task85DWebIssuerIssue(t, task85DDefaultEmailIssue(ctxFromEnv(env), account.String(), "email-changed-payload"))
				msg := task85DEmailMsg(t, env.ctx, resp, account.String(), accountPriv, "email-changed-payload", map[string]string{"source": "task85d-email-sandbox"})
				msg.EmailHash = task85DSHA256Hex("attacker@example.com")
				return msg, env.ctx
			},
		},
		{
			name: "stale-evidence",
			build: func(t *testing.T, env veidTestEnv, account sdk.AccAddress, accountPriv ed25519.PrivateKey) (*veidtypes.MsgSubmitEmailVerificationProof, sdk.Context) {
				cfg := task85DDefaultEmailIssue(ctxFromEnv(env), account.String(), "email-stale")
				cfg.IssuedAt = env.ctx.BlockTime().Add(-25 * time.Hour)
				cfg.ExpiresAt = env.ctx.BlockTime().Add(time.Hour)
				resp := task85DWebIssuerIssue(t, cfg)
				msg := task85DEmailMsg(t, env.ctx, resp, account.String(), accountPriv, "email-stale", map[string]string{"source": "task85d-email-sandbox"})
				return msg, env.ctx
			},
		},
		{
			name: "expired-evidence",
			build: func(t *testing.T, env veidTestEnv, account sdk.AccAddress, accountPriv ed25519.PrivateKey) (*veidtypes.MsgSubmitEmailVerificationProof, sdk.Context) {
				cfg := task85DDefaultEmailIssue(ctxFromEnv(env), account.String(), "email-expired")
				cfg.IssuedAt = env.ctx.BlockTime().Add(-2 * time.Hour)
				cfg.ExpiresAt = env.ctx.BlockTime().Add(-time.Hour)
				resp := task85DWebIssuerIssue(t, cfg)
				msg := task85DEmailMsg(t, env.ctx, resp, account.String(), accountPriv, "email-expired", map[string]string{"source": "task85d-email-sandbox"})
				return msg, env.ctx
			},
		},
		{
			name: "revoked-key",
			setup: func(t *testing.T, env veidTestEnv, ctx sdk.Context, reg task85DWebIssuerResponse) {
				task85DRegisterWebSigner(t, ctx, env.app.Keepers.VirtEngine.VEID, reg, veidtypes.SignerKeyStateRevoked)
			},
			build: func(t *testing.T, env veidTestEnv, account sdk.AccAddress, accountPriv ed25519.PrivateKey) (*veidtypes.MsgSubmitEmailVerificationProof, sdk.Context) {
				resp := task85DWebIssuerIssue(t, task85DDefaultEmailIssue(ctxFromEnv(env), account.String(), "email-revoked-key"))
				msg := task85DEmailMsg(t, env.ctx, resp, account.String(), accountPriv, "email-revoked-key", map[string]string{"source": "task85d-email-sandbox"})
				return msg, env.ctx
			},
		},
		{
			name: "rotated-key-not-registered",
			build: func(t *testing.T, env veidTestEnv, account sdk.AccAddress, accountPriv ed25519.PrivateKey) (*veidtypes.MsgSubmitEmailVerificationProof, sdk.Context) {
				cfg := task85DDefaultEmailIssue(ctxFromEnv(env), account.String(), "email-rotated-unregistered")
				cfg.Sequence = 2
				resp := task85DWebIssuerIssue(t, cfg)
				msg := task85DEmailMsg(t, env.ctx, resp, account.String(), accountPriv, "email-rotated-unregistered", map[string]string{"source": "task85d-email-sandbox"})
				return msg, env.ctx
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupVEIDTestEnv(t)
			env.ctx = task85DChainContext(env.ctx)
			ctx := env.ctx
			account, accountPriv := task85DCreateWallet(t, ctx, env.msgServer, env.app.Keepers.VirtEngine.VEID, "task85d-negative-"+tt.name)
			reg := task85DWebIssuerRegister(t, "web-sandbox", 1)
			if tt.setup != nil {
				tt.setup(t, env, ctx, reg)
			} else {
				task85DRegisterWebSigner(t, ctx, env.app.Keepers.VirtEngine.VEID, reg, veidtypes.SignerKeyStateActive)
			}
			msg, submitCtx := tt.build(t, env, account, accountPriv)
			before := task85DVEIDStoreDigest(t, submitCtx, env.app.Keepers.VirtEngine.VEID)
			_, err := env.msgServer.SubmitEmailVerificationProof(submitCtx, msg)
			require.Error(t, err)
			require.Equal(t, before, task85DVEIDStoreDigest(t, submitCtx, env.app.Keepers.VirtEngine.VEID))
		})
	}
}

func TestTask85DWebEvidenceNonceReplayDoesNotMutateSecondSubmission(t *testing.T) {
	env := setupVEIDTestEnv(t)
	env.ctx = task85DChainContext(env.ctx)
	ctx := env.ctx
	account, accountPriv := task85DCreateWallet(t, ctx, env.msgServer, env.app.Keepers.VirtEngine.VEID, "task85d-replay")
	reg := task85DWebIssuerRegister(t, "web-sandbox", 1)
	task85DRegisterWebSigner(t, ctx, env.app.Keepers.VirtEngine.VEID, reg, veidtypes.SignerKeyStateActive)

	nonce := task85DNonceHex("shared-replay-nonce")
	firstCfg := task85DDefaultEmailIssue(ctx, account.String(), "email-replay-original")
	firstCfg.NonceHex = nonce
	first := task85DWebIssuerIssue(t, firstCfg)
	firstMsg := task85DEmailMsg(t, ctx, first, account.String(), accountPriv, "email-replay-original", map[string]string{"source": "task85d-email-sandbox"})
	_, err := env.msgServer.SubmitEmailVerificationProof(ctx, firstMsg)
	require.NoError(t, err)

	secondCfg := task85DDefaultEmailIssue(ctx, account.String(), "email-replay-second")
	secondCfg.NonceHex = nonce
	second := task85DWebIssuerIssue(t, secondCfg)
	secondMsg := task85DEmailMsg(t, ctx, second, account.String(), accountPriv, "email-replay-second", map[string]string{"source": "task85d-email-sandbox"})
	before := task85DVEIDStoreDigest(t, ctx, env.app.Keepers.VirtEngine.VEID)
	_, err = env.msgServer.SubmitEmailVerificationProof(ctx, secondMsg)
	require.Error(t, err)
	require.Equal(t, before, task85DVEIDStoreDigest(t, ctx, env.app.Keepers.VirtEngine.VEID))
	_, found := env.app.Keepers.VirtEngine.VEID.GetEmailVerificationRecord(ctx, "email-replay-second")
	require.False(t, found)
}

func TestTask85DInferenceReceiptSeparateProcessDeterministicSandbox(t *testing.T) {
	account := task85DAccountAddress("task85d-inference-account").String()
	req := task85DHelperRequest{
		Helper:      task85DInferenceHelper,
		Action:      "issue",
		Profile:     task85DInferenceProfile,
		Sequence:    1,
		ChainID:     task85DChainID,
		Account:     account,
		EvidenceID:  "task85d-inference-golden-request",
		GoldenInput: task85DInferenceGolden,
	}

	first := task85DRunInferenceRuntime(t, req)
	second := task85DRunInferenceRuntime(t, req)
	require.Equal(t, task85DInferenceProfile, first.RuntimeName)
	require.NotContains(t, first.RuntimeName, "production")
	require.Equal(t, first.PublicKey, second.PublicKey)
	require.Equal(t, first.ReceiptBytes, second.ReceiptBytes)
	require.Equal(t, first.Score, second.Score)
	require.Equal(t, first.GoldenInput, second.GoldenInput)
	require.Equal(t, first.InputDigest, second.InputDigest)
	require.Equal(t, first.FeatureDigest, second.FeatureDigest)
	require.Equal(t, first.OutputDigest, second.OutputDigest)
	require.Equal(t, first.EvidenceLineage, second.EvidenceLineage)
	require.Equal(t, first.ModelManifestDigest, second.ModelManifestDigest)
	require.Equal(t, first.ModelDigest, second.ModelDigest)
	require.Equal(t, first.RuntimeImageDigest, second.RuntimeImageDigest)
	require.Equal(t, first.RuntimeDigest, second.RuntimeDigest)
	require.Equal(t, first.SchemaDigest, second.SchemaDigest)
	require.Equal(t, first.ConfigDigest, second.ConfigDigest)

	decoded, err := veidtypes.DecodeCanonicalSignedInferenceReceipt(first.ReceiptBytes)
	require.NoError(t, err)
	require.NoError(t, decoded.VerifySignature(first.PublicKey))
	canonical, err := decoded.CanonicalSignedBytes()
	require.NoError(t, err)
	require.Equal(t, first.ReceiptBytes, canonical)
	require.Equal(t, task85DChainID, decoded.ChainID)
	require.Equal(t, account, decoded.AccountAddress)
	require.Equal(t, req.EvidenceID, decoded.RequestID)
	require.Equal(t, uint32(91), decoded.Score)
	require.Equal(t, first.GoldenInput, req.GoldenInput)
	require.Equal(t, first.InputDigest, hex.EncodeToString(decoded.InputDigest))
	require.Equal(t, first.FeatureDigest, hex.EncodeToString(decoded.FeatureDigest))
	require.Equal(t, first.EvidenceLineage, hex.EncodeToString(decoded.EvidenceLineageDigest))
	require.Equal(t, veidtypes.VerificationResultStatusSuccess, decoded.Status)
	require.Equal(t, []veidtypes.ReasonCode{veidtypes.ReasonCodeSuccess}, decoded.ReasonCodes)
	require.Equal(t, []string{string(veidtypes.ReasonCodeSuccess)}, first.ReasonCodes)
	require.True(t, decoded.DeterminismProfile.IsCanonical())
	require.True(t, decoded.DeterminismProfile.ForceCPU)
	require.Equal(t, veidtypes.InferenceReceiptRequiredRandomSeed, decoded.DeterminismProfile.RandomSeed)
	require.True(t, decoded.DeterminismProfile.DeterministicOps)
	require.Equal(t, int32(1), decoded.DeterminismProfile.InterOpThreads)
	require.Equal(t, int32(1), decoded.DeterminismProfile.IntraOpThreads)
	require.True(t, decoded.DeterminismProfile.DisableGPU)
	require.Equal(t, first.ModelManifestDigest, hex.EncodeToString(decoded.ModelManifestDigest))
	require.Equal(t, first.ModelDigest, hex.EncodeToString(decoded.ModelDigest))
	require.Equal(t, first.RuntimeImageDigest, hex.EncodeToString(decoded.RuntimeImageDigest))
	require.Equal(t, first.RuntimeDigest, hex.EncodeToString(decoded.RuntimeDigest))
	require.Equal(t, first.SchemaDigest, hex.EncodeToString(decoded.SchemaDigest))
	require.Equal(t, first.ConfigDigest, hex.EncodeToString(decoded.ConfigDigest))
}

func TestTask85DInferenceReceiptChangedInputChangesDeterministicReceipt(t *testing.T) {
	account := task85DAccountAddress("task85d-inference-account").String()
	req := task85DHelperRequest{
		Helper:      task85DInferenceHelper,
		Action:      "issue",
		Profile:     task85DInferenceProfile,
		Sequence:    1,
		ChainID:     task85DChainID,
		Account:     account,
		EvidenceID:  "task85d-inference-golden-request",
		GoldenInput: task85DInferenceGolden,
	}
	changedReq := req
	changedReq.GoldenInput = `{"account_age_days":365,"email_verified":true,"sms_mobile_verified":true,"social_verified":false,"org_domain_verified":true,"risk_signals":[]}`

	originalA := task85DRunInferenceRuntime(t, req)
	originalB := task85DRunInferenceRuntime(t, req)
	changedA := task85DRunInferenceRuntime(t, changedReq)
	changedB := task85DRunInferenceRuntime(t, changedReq)

	require.Equal(t, originalA.ReceiptBytes, originalB.ReceiptBytes)
	require.Equal(t, changedA.ReceiptBytes, changedB.ReceiptBytes)
	require.NotEqual(t, originalA.InputDigest, changedA.InputDigest)
	require.NotEqual(t, originalA.FeatureDigest, changedA.FeatureDigest)
	require.NotEqual(t, originalA.OutputDigest, changedA.OutputDigest)
	require.NotEqual(t, originalA.EvidenceLineage, changedA.EvidenceLineage)
	require.NotEqual(t, originalA.Score, changedA.Score)
	require.NotEqual(t, originalA.ReceiptBytes, changedA.ReceiptBytes)
}

type task85DWebIssueConfig struct {
	Profile       string
	Sequence      uint64
	ChainID       string
	Account       string
	Kind          string
	EvidenceID    string
	NonceHex      string
	IssuedAt      time.Time
	ExpiresAt     time.Time
	Metadata      map[string]string
	PayloadDigest string
}

func task85DDefaultEmailIssue(ctx sdk.Context, account string, evidenceID string) task85DWebIssueConfig {
	return task85DWebIssueConfig{
		Profile:    "web-sandbox",
		Sequence:   1,
		ChainID:    ctx.ChainID(),
		Account:    account,
		Kind:       task85DKindEmail,
		EvidenceID: evidenceID,
		IssuedAt:   ctx.BlockTime(),
		ExpiresAt:  ctx.BlockTime().Add(time.Hour),
		Metadata:   map[string]string{"source": "task85d-email-sandbox"},
	}
}

func ctxFromEnv(env veidTestEnv) sdk.Context {
	return env.ctx
}

func task85DChainContext(ctx sdk.Context) sdk.Context {
	return ctx.WithChainID(task85DChainID).WithHeaderInfo(coreheader.Info{
		ChainID: task85DChainID,
		Height:  ctx.BlockHeight(),
		Time:    ctx.BlockTime(),
	})
}

func task85DRunWebIssuerHelper(req task85DHelperRequest) (task85DWebIssuerResponse, error) {
	if req.Profile == "" {
		return task85DWebIssuerResponse{}, fmt.Errorf("profile is required")
	}
	if req.Profile != task85DWebProfile {
		return task85DWebIssuerResponse{}, fmt.Errorf("unsupported web issuer profile %q", req.Profile)
	}
	if req.Sequence == 0 {
		req.Sequence = 1
	}
	pub, priv := task85DDeterministicEd25519Key("task85d-web-issuer:" + req.Profile + ":" + strconv.FormatUint(req.Sequence, 10))
	issuer := veidtypes.AttestationIssuer{
		ID:             "did:virtengine:issuer:" + req.Profile,
		KeyID:          "did:virtengine:issuer:" + req.Profile + ":" + strconv.FormatUint(req.Sequence, 10),
		KeyFingerprint: veidtypes.ComputeKeyFingerprint(pub),
	}
	resp := task85DWebIssuerResponse{
		Issuer:    issuer,
		PublicKey: append([]byte(nil), pub...),
	}
	if req.Action == "register" {
		return resp, nil
	}
	if req.Action != "issue" {
		return task85DWebIssuerResponse{}, fmt.Errorf("unsupported web issuer action %q", req.Action)
	}
	if req.ChainID == "" || req.Account == "" || req.Kind == "" || req.EvidenceID == "" || req.NonceHex == "" {
		return task85DWebIssuerResponse{}, fmt.Errorf("chain_id, account, kind, evidence_id, and nonce_hex are required")
	}
	if req.IssuedUnix <= 0 || req.ExpiresUnix <= 0 {
		return task85DWebIssuerResponse{}, fmt.Errorf("issued_unix and expires_unix are required")
	}
	issuedAt := time.Unix(req.IssuedUnix, 0).UTC()
	expiresAt := time.Unix(req.ExpiresUnix, 0).UTC()
	nonce, err := hex.DecodeString(req.NonceHex)
	if err != nil {
		return task85DWebIssuerResponse{}, err
	}
	switch req.Kind {
	case task85DKindSSO:
		att := veidtypes.NewSSOAttestation(issuer, veidtypes.NewAttestationSubject(req.Account), "https://accounts.google.com", "google-subject-"+req.EvidenceID, veidtypes.SSOProviderGoogle, "oidc-"+req.EvidenceID, nonce, issuedAt, expiresAt.Sub(issuedAt))
		att.SetEmail("jane@example.com", "example.com", true)
		canonical, err := att.CanonicalBytes()
		if err != nil {
			return task85DWebIssuerResponse{}, err
		}
		evidence := veidtypes.NewWebEvidenceContext(veidtypes.WebEvidenceContextConfig{
			ChainID:             req.ChainID,
			AccountAddress:      req.Account,
			EvidenceType:        veidtypes.AttestationTypeSSOVerification,
			Action:              veidtypes.WebEvidenceActionSubmitSSO,
			ScopeID:             req.EvidenceID,
			AttestationDigest:   veidtypes.WebEvidenceDigestHex(canonical),
			Issuer:              issuer,
			IssuerAlgorithm:     veidtypes.ProofTypeEd25519,
			Nonce:               att.Nonce,
			Challenge:           att.OIDCNonce,
			IssuedAt:            issuedAt,
			ExpiresAt:           expiresAt,
			ServiceMetadataHash: task85DServiceMetadataHash(req.Metadata),
			CallerFields: map[string]string{
				"linkage_id":             req.EvidenceID,
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
		if err := evidence.ApplyToAttestation(&att.VerificationAttestation); err != nil {
			return task85DWebIssuerResponse{}, err
		}
		signBytes, err := evidence.IssuerSignBytes()
		if err != nil {
			return task85DWebIssuerResponse{}, err
		}
		att.SetProof(veidtypes.NewAttestationProof(veidtypes.ProofTypeEd25519, issuedAt, issuer.ID+"#"+issuer.KeyID, ed25519.Sign(priv, signBytes), att.Nonce))
		bz, err := json.Marshal(att)
		if err != nil {
			return task85DWebIssuerResponse{}, err
		}
		resp.AttestationData = bz
		resp.EvidenceHash = task85DSHA256HexBytes(bz)
		resp.OIDCIssuer = att.OIDCIssuer
		resp.OIDCNonce = att.OIDCNonce
		resp.Provider = string(att.ProviderType)
		resp.LinkedAccount = att.LinkedAccountAddress
		resp.SubjectHash = att.SubjectHash
		resp.SSOEmailHash = att.EmailHash
		resp.SSOEmailDomain = att.EmailDomainHash
	case task85DKindEmail, task85DKindSMS, task85DKindSocial:
		attType, action, callerFields, fields, err := task85DWebIssuerVerificationFields(req, issuedAt, expiresAt)
		if err != nil {
			return task85DWebIssuerResponse{}, err
		}
		att := veidtypes.NewVerificationAttestation(issuer, veidtypes.NewAttestationSubject(req.Account), attType, nonce, issuedAt, expiresAt.Sub(issuedAt), 100, 100)
		digest, err := veidtypes.WebEvidenceAttestationDigestHex(att)
		if err != nil {
			return task85DWebIssuerResponse{}, err
		}
		evidence := veidtypes.NewWebEvidenceContext(veidtypes.WebEvidenceContextConfig{
			ChainID:             req.ChainID,
			AccountAddress:      req.Account,
			EvidenceType:        attType,
			Action:              action,
			ScopeID:             req.EvidenceID,
			AttestationDigest:   digest,
			Issuer:              issuer,
			IssuerAlgorithm:     veidtypes.ProofTypeEd25519,
			Nonce:               att.Nonce,
			Challenge:           req.EvidenceID,
			IssuedAt:            issuedAt,
			ExpiresAt:           expiresAt,
			ServiceMetadataHash: task85DServiceMetadataHash(req.Metadata),
			CallerFields:        callerFields,
		})
		if err := evidence.ApplyToAttestation(att); err != nil {
			return task85DWebIssuerResponse{}, err
		}
		signBytes, err := evidence.IssuerSignBytes()
		if err != nil {
			return task85DWebIssuerResponse{}, err
		}
		att.SetProof(veidtypes.NewAttestationProof(veidtypes.ProofTypeEd25519, issuedAt, issuer.ID+"#"+issuer.KeyID, ed25519.Sign(priv, signBytes), att.Nonce))
		bz, err := att.ToJSON()
		if err != nil {
			return task85DWebIssuerResponse{}, err
		}
		resp.AttestationData = bz
		resp.EvidenceHash = task85DSHA256HexBytes(bz)
		fields(&resp)
	default:
		return task85DWebIssuerResponse{}, fmt.Errorf("unsupported web evidence kind %q", req.Kind)
	}
	resp.VerifiedAt = issuedAt.Unix()
	resp.ExpiresAt = expiresAt.Unix()
	return resp, nil
}

func task85DRunInferenceRuntimeHelper(req task85DHelperRequest) (task85DInferenceReceiptResponse, error) {
	if req.Action != "issue" {
		return task85DInferenceReceiptResponse{}, fmt.Errorf("unsupported inference runtime action %q", req.Action)
	}
	if req.Profile != task85DInferenceProfile {
		return task85DInferenceReceiptResponse{}, fmt.Errorf("unsupported inference runtime profile %q", req.Profile)
	}
	if req.Sequence == 0 {
		req.Sequence = 1
	}
	if req.ChainID == "" || req.Account == "" || req.EvidenceID == "" {
		return task85DInferenceReceiptResponse{}, fmt.Errorf("chain_id, account, and evidence_id are required")
	}
	if req.GoldenInput == "" {
		return task85DInferenceReceiptResponse{}, fmt.Errorf("golden_input is required")
	}
	runtime.GOMAXPROCS(1)
	golden, canonicalInput, err := task85DParseInferenceGoldenInput(req.GoldenInput)
	if err != nil {
		return task85DInferenceReceiptResponse{}, err
	}
	inputDigest := task85DDigestBytes(canonicalInput)
	featureBytes, featureDigest := task85DInferenceFeatureVector(golden, inputDigest)
	score := task85DInferenceScore(golden)
	outputBytes, outputDigest := task85DInferenceOutput(score)
	modelManifestDigest := task85DDigestBytes("task85d deterministic sandbox model manifest v1")
	modelDigest := task85DDigestBytes("task85d deterministic sandbox model weights v1")
	runtimeImageDigest := task85DDigestBytes("task85d deterministic sandbox runtime image v1")
	runtimeDigest := task85DDigestBytes(task85DInferenceProfile + "|gomaxprocs=1|force_cpu=true")
	schemaDigest := task85DDigestBytes("task85d inference schema v1")
	configDigest := veidtypes.CanonicalInferenceDeterminismConfigDigest()
	lineageBytes := task85DJoin([]string{
		"config=" + hex.EncodeToString(configDigest),
		"feature=" + hex.EncodeToString(featureDigest),
		"input=" + hex.EncodeToString(inputDigest),
		"model=" + hex.EncodeToString(modelDigest),
		"model_manifest=" + hex.EncodeToString(modelManifestDigest),
		"output=" + hex.EncodeToString(outputDigest),
		"runtime=" + hex.EncodeToString(runtimeDigest),
		"runtime_image=" + hex.EncodeToString(runtimeImageDigest),
		"schema=" + hex.EncodeToString(schemaDigest),
	})
	evidenceLineageDigest := task85DDigestBytes(lineageBytes)
	_ = featureBytes
	_ = outputBytes

	pub, priv := task85DDeterministicEd25519Key("task85d-inference-signer:" + req.Profile + ":" + strconv.FormatUint(req.Sequence, 10))
	runtimeName := task85DInferenceProfile
	issuedAt := time.Unix(1710000000, 0).UTC()
	receipt := veidtypes.InferenceReceipt{
		Domain:                veidtypes.InferenceReceiptDomain,
		Version:               veidtypes.InferenceReceiptVersion,
		ChainID:               req.ChainID,
		AccountAddress:        req.Account,
		RequestID:             req.EvidenceID,
		ScopeIDs:              veidtypes.CanonicalInferenceReceiptScopeIDs([]string{"task85d-email-scope", "task85d-sms-scope", "task85d-social-scope"}),
		Nonce:                 task85DNonceHex("inference:" + req.EvidenceID),
		InputDigest:           inputDigest,
		FeatureDigest:         featureDigest,
		SchemaDigest:          schemaDigest,
		EvidenceLineageDigest: evidenceLineageDigest,
		PipelineVersion:       "task85d-engineering-sandbox-pipeline-v1",
		ModelManifestDigest:   modelManifestDigest,
		ModelDigest:           modelDigest,
		RuntimeImageDigest:    runtimeImageDigest,
		RuntimeDigest:         runtimeDigest,
		ConfigDigest:          configDigest,
		DeterminismProfile:    veidtypes.CanonicalInferenceDeterminismProfile(),
		Score:                 score,
		Status:                veidtypes.VerificationResultStatusSuccess,
		ConfidenceMillionths:  990000,
		ReasonCodes:           []veidtypes.ReasonCode{veidtypes.ReasonCodeSuccess},
		IssuedHeight:          85,
		IssuedAt:              issuedAt,
		ExpiresHeight:         185,
		ExpiresAt:             issuedAt.Add(time.Hour),
		SignerKeyID:           "did:virtengine:inference-signer:" + req.Profile,
		SignerFingerprint:     veidtypes.ComputeKeyFingerprint(pub),
		SignerSequence:        req.Sequence,
	}
	if err := receipt.Sign(priv); err != nil {
		return task85DInferenceReceiptResponse{}, err
	}
	receiptBytes, err := receipt.CanonicalSignedBytes()
	if err != nil {
		return task85DInferenceReceiptResponse{}, err
	}
	return task85DInferenceReceiptResponse{
		RuntimeName:         runtimeName,
		PublicKey:           append([]byte(nil), pub...),
		ReceiptBytes:        receiptBytes,
		Score:               receipt.Score,
		GoldenInput:         canonicalInput,
		InputDigest:         hex.EncodeToString(receipt.InputDigest),
		FeatureDigest:       hex.EncodeToString(receipt.FeatureDigest),
		OutputDigest:        hex.EncodeToString(outputDigest),
		EvidenceLineage:     hex.EncodeToString(receipt.EvidenceLineageDigest),
		ModelManifestDigest: hex.EncodeToString(receipt.ModelManifestDigest),
		ModelDigest:         hex.EncodeToString(receipt.ModelDigest),
		RuntimeImageDigest:  hex.EncodeToString(receipt.RuntimeImageDigest),
		RuntimeDigest:       hex.EncodeToString(receipt.RuntimeDigest),
		SchemaDigest:        hex.EncodeToString(receipt.SchemaDigest),
		ConfigDigest:        hex.EncodeToString(receipt.ConfigDigest),
		ReasonCodes:         []string{string(veidtypes.ReasonCodeSuccess)},
	}, nil
}

func task85DParseInferenceGoldenInput(input string) (task85DInferenceGoldenInput, string, error) {
	if len(input) > 8*1024 {
		return task85DInferenceGoldenInput{}, "", fmt.Errorf("golden_input exceeds 8192 bytes")
	}
	dec := json.NewDecoder(strings.NewReader(input))
	dec.DisallowUnknownFields()
	var golden task85DInferenceGoldenInput
	if err := dec.Decode(&golden); err != nil {
		return task85DInferenceGoldenInput{}, "", fmt.Errorf("decode golden_input: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return task85DInferenceGoldenInput{}, "", fmt.Errorf("trailing JSON after golden_input")
	}
	sort.Strings(golden.RiskSignals)
	if golden.AccountAgeDays == 0 {
		return task85DInferenceGoldenInput{}, "", fmt.Errorf("account_age_days is required")
	}
	canonical, err := json.Marshal(golden)
	if err != nil {
		return task85DInferenceGoldenInput{}, "", err
	}
	return golden, string(canonical), nil
}

func task85DInferenceFeatureVector(golden task85DInferenceGoldenInput, inputDigest []byte) ([]byte, []byte) {
	features := struct {
		InputDigest       string   `json:"input_digest"`
		AccountAgeBucket  string   `json:"account_age_bucket"`
		EmailVerified     bool     `json:"email_verified"`
		SMSMobileVerified bool     `json:"sms_mobile_verified"`
		SocialVerified    bool     `json:"social_verified"`
		OrgDomainVerified bool     `json:"org_domain_verified"`
		RiskSignals       []string `json:"risk_signals"`
	}{
		InputDigest:       hex.EncodeToString(inputDigest),
		AccountAgeBucket:  "365d-plus",
		EmailVerified:     golden.EmailVerified,
		SMSMobileVerified: golden.SMSMobileVerified,
		SocialVerified:    golden.SocialVerified,
		OrgDomainVerified: golden.OrgDomainVerified,
		RiskSignals:       append([]string(nil), golden.RiskSignals...),
	}
	if golden.AccountAgeDays < 365 {
		features.AccountAgeBucket = "under-365d"
	}
	bz, err := json.Marshal(features)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(bz)
	return bz, digest[:]
}

func task85DInferenceScore(golden task85DInferenceGoldenInput) uint32 {
	score := uint32(35)
	if golden.EmailVerified {
		score += 16
	}
	if golden.SMSMobileVerified {
		score += 14
	}
	if golden.SocialVerified {
		score += 15
	}
	if golden.OrgDomainVerified {
		score += 6
	}
	if golden.AccountAgeDays >= 365 {
		score += 5
	}
	riskCount := uint32(len(golden.RiskSignals)) //nolint:gosec // golden_input is bounded to 8 KiB before parsing.
	penalty := riskCount * 4
	if penalty >= score {
		return 0
	}
	return score - penalty
}

func task85DInferenceOutput(score uint32) ([]byte, []byte) {
	output := struct {
		Score       uint32   `json:"score"`
		Status      string   `json:"status"`
		ReasonCodes []string `json:"reason_codes"`
	}{
		Score:       score,
		Status:      string(veidtypes.VerificationResultStatusSuccess),
		ReasonCodes: []string{string(veidtypes.ReasonCodeSuccess)},
	}
	bz, err := json.Marshal(output)
	if err != nil {
		panic(err)
	}
	digest := sha256.Sum256(bz)
	return bz, digest[:]
}

func task85DWebIssuerVerificationFields(req task85DHelperRequest, issuedAt time.Time, expiresAt time.Time) (veidtypes.AttestationType, string, map[string]string, func(*task85DWebIssuerResponse), error) {
	switch req.Kind {
	case task85DKindEmail:
		emailHash := task85DSHA256Hex("jane@example.com")
		domainHash := task85DSHA256Hex("example.com")
		callerFields := map[string]string{
			"verification_id":   req.EvidenceID,
			"email_hash":        emailHash,
			"domain_hash":       domainHash,
			"nonce":             req.NonceHex,
			"is_organizational": "true",
			"verified_at_unix":  strconv.FormatInt(issuedAt.Unix(), 10),
			"expires_at_unix":   strconv.FormatInt(expiresAt.Unix(), 10),
		}
		return veidtypes.AttestationTypeEmailVerification, veidtypes.WebEvidenceActionSubmitEmail, callerFields, func(resp *task85DWebIssuerResponse) {
			resp.EmailHash = emailHash
			resp.DomainHash = domainHash
			resp.IsOrganizational = true
		}, nil
	case task85DKindSMS:
		phoneHash := task85DSHA256Hex("+15555550123")
		phoneSalt := hex.EncodeToString(bytes.Repeat([]byte{0x11}, 32))
		countryHash := task85DSHA256Hex("US")
		callerFields := map[string]string{
			"verification_id":   req.EvidenceID,
			"phone_hash":        phoneHash,
			"phone_hash_salt":   phoneSalt,
			"country_code_hash": countryHash,
			"is_voip":           "false",
			"carrier_type":      "mobile",
			"validator_address": "validator-1",
			"verified_at_unix":  strconv.FormatInt(issuedAt.Unix(), 10),
			"expires_at_unix":   strconv.FormatInt(expiresAt.Unix(), 10),
		}
		return veidtypes.AttestationTypeSMSVerification, veidtypes.WebEvidenceActionSubmitSMS, callerFields, func(resp *task85DWebIssuerResponse) {
			resp.PhoneHash = phoneHash
			resp.PhoneHashSalt = phoneSalt
			resp.CountryCodeHash = countryHash
			resp.CarrierType = "mobile"
			resp.ValidatorAddress = "validator-1"
		}, nil
	case task85DKindSocial:
		nameHash := veidtypes.HashSocialMediaField("Jane Doe")
		emailHash := veidtypes.HashSocialMediaField("jane@example.com")
		usernameHash := veidtypes.HashSocialMediaField("jane")
		orgHash := veidtypes.HashSocialMediaField("Example Org")
		accountCreatedAt := issuedAt.Add(-365 * 24 * time.Hour).Unix()
		if req.PayloadDigest == "" {
			return "", "", nil, nil, fmt.Errorf("payload_digest is required")
		}
		callerFields := map[string]string{
			"scope_id":                 req.EvidenceID,
			"provider":                 string(veidtypes.SocialMediaProviderGoogle),
			"profile_name_hash":        nameHash,
			"email_hash":               emailHash,
			"username_hash":            usernameHash,
			"org_hash":                 orgHash,
			"account_created_at_unix":  strconv.FormatInt(accountCreatedAt, 10),
			"account_age_days":         "365",
			"is_verified":              "true",
			"friend_count_range":       "500-999",
			"encrypted_payload_digest": req.PayloadDigest,
		}
		return veidtypes.AttestationTypeSocialMediaVerification, veidtypes.WebEvidenceActionSubmitSocial, callerFields, func(resp *task85DWebIssuerResponse) {
			resp.ProfileNameHash = nameHash
			resp.SocialEmailHash = emailHash
			resp.UsernameHash = usernameHash
			resp.OrgHash = orgHash
			resp.AccountCreatedAt = accountCreatedAt
			resp.AccountAgeDays = 365
			resp.IsVerified = true
			resp.FriendCountRange = "500-999"
			resp.Provider = string(veidtypes.SocialMediaProviderGoogle)
		}, nil
	default:
		return "", "", nil, nil, fmt.Errorf("unsupported verification kind %q", req.Kind)
	}
}

func task85DWebIssuerRegister(t *testing.T, profile string, sequence uint64) task85DWebIssuerResponse {
	t.Helper()
	return task85DRunWebIssuer(t, task85DHelperRequest{
		Helper:   task85DWebHelper,
		Action:   "register",
		Profile:  profile,
		Sequence: sequence,
	})
}

func task85DWebIssuerIssue(t *testing.T, cfg task85DWebIssueConfig) task85DWebIssuerResponse {
	t.Helper()
	if cfg.NonceHex == "" {
		cfg.NonceHex = task85DNonceHex(cfg.EvidenceID)
	}
	return task85DRunWebIssuer(t, task85DHelperRequest{
		Helper:        task85DWebHelper,
		Action:        "issue",
		Profile:       cfg.Profile,
		Sequence:      cfg.Sequence,
		ChainID:       cfg.ChainID,
		Account:       cfg.Account,
		Kind:          cfg.Kind,
		EvidenceID:    cfg.EvidenceID,
		NonceHex:      cfg.NonceHex,
		IssuedUnix:    cfg.IssuedAt.Unix(),
		ExpiresUnix:   cfg.ExpiresAt.Unix(),
		Metadata:      cfg.Metadata,
		PayloadDigest: cfg.PayloadDigest,
	})
}

func task85DRunWebIssuer(t *testing.T, req task85DHelperRequest) task85DWebIssuerResponse {
	t.Helper()
	reqBytes, err := json.Marshal(req)
	require.NoError(t, err)
	require.LessOrEqual(t, len(reqBytes), task85DHelperMaxBytes)
	stdout, stderr, err := task85DRunHelperRaw(t, append(reqBytes, '\n'))
	require.NoErrorf(t, err, "helper stderr: %s", stderr)
	var resp task85DWebIssuerResponse
	task85DDecodeHelperResponse(t, stdout, &resp)
	return resp
}

func task85DRunInferenceRuntime(t *testing.T, req task85DHelperRequest) task85DInferenceReceiptResponse {
	t.Helper()
	reqBytes, err := json.Marshal(req)
	require.NoError(t, err)
	require.LessOrEqual(t, len(reqBytes), task85DHelperMaxBytes)
	stdout, stderr, err := task85DRunHelperRaw(t, append(reqBytes, '\n'))
	require.NoErrorf(t, err, "helper stderr: %s", stderr)
	var resp task85DInferenceReceiptResponse
	task85DDecodeHelperResponse(t, stdout, &resp)
	return resp
}

func task85DRegisterWebSigner(t *testing.T, ctx sdk.Context, k keeper.Keeper, reg task85DWebIssuerResponse, state veidtypes.SignerKeyState) {
	t.Helper()
	now := ctx.BlockTime().UTC()
	key := veidtypes.NewSignerKeyInfo(reg.Issuer.ID, reg.PublicKey, veidtypes.ProofTypeEd25519, 1, now.Add(-time.Hour))
	key.KeyID = reg.Issuer.KeyID
	key.Fingerprint = reg.Issuer.KeyFingerprint
	key.State = state
	activatedAt := now.Add(-time.Minute)
	expiresAt := now.Add(24 * time.Hour)
	key.ActivatedAt = &activatedAt
	key.ExpiresAt = &expiresAt
	key.Metadata[veidtypes.SignerKeyMetadataEvidenceTypes] = task85DJoin([]string{
		string(veidtypes.AttestationTypeSSOVerification),
		string(veidtypes.AttestationTypeEmailVerification),
		string(veidtypes.AttestationTypeSMSVerification),
		string(veidtypes.AttestationTypeSocialMediaVerification),
	})
	if state == veidtypes.SignerKeyStateRevoked {
		revokedAt := now.Add(-time.Second)
		key.RevokedAt = &revokedAt
		key.RevocationReason = veidtypes.RevocationReasonAdministrative
	}
	require.NoError(t, k.RegisterSignerKey(ctx, k.GetAuthority(), key))
}

func task85DCreateWallet(t *testing.T, ctx sdk.Context, msgServer veidtypes.MsgServer, k keeper.Keeper, seed string) (sdk.AccAddress, ed25519.PrivateKey) {
	t.Helper()
	account := task85DAccountAddress(seed)
	pub, priv := task85DDeterministicEd25519Key("task85d-wallet:" + seed)
	walletID := keeper.GenerateWalletID(account.String())
	bindingSig := ed25519.Sign(priv, veidtypes.GetWalletBindingMessage(walletID, account.String()))
	resp, err := msgServer.CreateIdentityWallet(ctx, &veidtypes.MsgCreateIdentityWallet{
		Sender:           account.String(),
		BindingSignature: bindingSig,
		BindingPubKey:    pub,
	})
	require.NoError(t, err)
	require.Equal(t, walletID, resp.WalletId)
	_, found := k.GetWallet(ctx, account)
	require.True(t, found)
	return account, priv
}

func task85DSeedSocialName(t *testing.T, ctx sdk.Context, k keeper.Keeper, account sdk.AccAddress) {
	t.Helper()
	wallet, found := k.GetWallet(ctx, account)
	require.True(t, found)
	nameHashBytes, err := hex.DecodeString(veidtypes.HashSocialMediaField("Jane Doe"))
	require.NoError(t, err)
	wallet.DerivedFeatures.DocFieldHashes[veidtypes.DocFieldNameHash] = nameHashBytes
	wallet.DerivedFeatures.LastComputedAt = ctx.BlockTime()
	require.NoError(t, k.SetWallet(ctx, wallet))
}

func task85DSSOMsg(t *testing.T, ctx sdk.Context, resp task85DWebIssuerResponse, account string, accountPriv ed25519.PrivateKey, linkageID string, metadata map[string]string) *veidtypes.MsgSubmitSSOVerificationProof {
	t.Helper()
	var att veidtypes.SSOAttestation
	require.NoError(t, json.Unmarshal(resp.AttestationData, &att))
	evidence := task85DSSOEvidence(t, ctx.ChainID(), account, linkageID, metadata, &att)
	att.SetLinkageSignature(task85DAccountSignature(t, accountPriv, evidence))
	attestationData, err := json.Marshal(att)
	require.NoError(t, err)
	return &veidtypes.MsgSubmitSSOVerificationProof{
		AccountAddress:         account,
		LinkageId:              linkageID,
		AttestationData:        attestationData,
		EvidenceHash:           task85DSHA256HexBytes(attestationData),
		EvidenceStorageBackend: string(veidtypes.StorageBackendWaldur),
		EvidenceStorageRef:     "vault://task85d/sso/" + linkageID,
		EvidenceMetadata:       metadata,
	}
}

func task85DEmailMsg(t *testing.T, ctx sdk.Context, resp task85DWebIssuerResponse, account string, accountPriv ed25519.PrivateKey, verificationID string, metadata map[string]string) *veidtypes.MsgSubmitEmailVerificationProof {
	t.Helper()
	att := task85DVerificationAttestation(t, resp.AttestationData)
	evidence := task85DVerificationEvidence(t, ctx.ChainID(), account, verificationID, verificationID, metadata, att, veidtypes.AttestationTypeEmailVerification, veidtypes.WebEvidenceActionSubmitEmail, map[string]string{
		"verification_id":   verificationID,
		"email_hash":        resp.EmailHash,
		"domain_hash":       resp.DomainHash,
		"nonce":             att.Nonce,
		"is_organizational": strconv.FormatBool(resp.IsOrganizational),
		"verified_at_unix":  strconv.FormatInt(att.IssuedAt.Unix(), 10),
		"expires_at_unix":   strconv.FormatInt(att.ExpiresAt.Unix(), 10),
	})
	return &veidtypes.MsgSubmitEmailVerificationProof{
		AccountAddress:         account,
		VerificationId:         verificationID,
		EmailHash:              resp.EmailHash,
		DomainHash:             resp.DomainHash,
		Nonce:                  att.Nonce,
		VerifiedAt:             att.IssuedAt.Unix(),
		ExpiresAt:              att.ExpiresAt.Unix(),
		AttestationData:        resp.AttestationData,
		AccountSignature:       task85DAccountSignature(t, accountPriv, evidence),
		IsOrganizational:       resp.IsOrganizational,
		EvidenceHash:           task85DSHA256HexBytes(resp.AttestationData),
		EvidenceStorageBackend: string(veidtypes.StorageBackendWaldur),
		EvidenceStorageRef:     "vault://task85d/email/" + verificationID,
		EvidenceMetadata:       metadata,
	}
}

func task85DSMSMsg(t *testing.T, ctx sdk.Context, resp task85DWebIssuerResponse, account string, accountPriv ed25519.PrivateKey, verificationID string, metadata map[string]string) *veidtypes.MsgSubmitSMSVerificationProof {
	t.Helper()
	att := task85DVerificationAttestation(t, resp.AttestationData)
	evidence := task85DVerificationEvidence(t, ctx.ChainID(), account, verificationID, verificationID, metadata, att, veidtypes.AttestationTypeSMSVerification, veidtypes.WebEvidenceActionSubmitSMS, map[string]string{
		"verification_id":   verificationID,
		"phone_hash":        resp.PhoneHash,
		"phone_hash_salt":   resp.PhoneHashSalt,
		"country_code_hash": resp.CountryCodeHash,
		"is_voip":           strconv.FormatBool(resp.IsVoip),
		"carrier_type":      resp.CarrierType,
		"validator_address": resp.ValidatorAddress,
		"verified_at_unix":  strconv.FormatInt(att.IssuedAt.Unix(), 10),
		"expires_at_unix":   strconv.FormatInt(att.ExpiresAt.Unix(), 10),
	})
	return &veidtypes.MsgSubmitSMSVerificationProof{
		AccountAddress:         account,
		VerificationId:         verificationID,
		PhoneHash:              resp.PhoneHash,
		PhoneHashSalt:          resp.PhoneHashSalt,
		CountryCodeHash:        resp.CountryCodeHash,
		VerifiedAt:             att.IssuedAt.Unix(),
		ExpiresAt:              att.ExpiresAt.Unix(),
		IsVoip:                 resp.IsVoip,
		CarrierType:            resp.CarrierType,
		ValidatorAddress:       resp.ValidatorAddress,
		AttestationData:        resp.AttestationData,
		AccountSignature:       task85DAccountSignature(t, accountPriv, evidence),
		EvidenceHash:           task85DSHA256HexBytes(resp.AttestationData),
		EvidenceStorageBackend: string(veidtypes.StorageBackendWaldur),
		EvidenceStorageRef:     "vault://task85d/sms/" + verificationID,
		EvidenceMetadata:       metadata,
	}
}

func task85DSocialMsg(t *testing.T, ctx sdk.Context, resp task85DWebIssuerResponse, account string, accountPriv ed25519.PrivateKey, scopeID string, payload veidtypes.EncryptedPayloadEnvelopePB, metadata map[string]string) *veidtypes.MsgSubmitSocialMediaScope {
	t.Helper()
	att := task85DVerificationAttestation(t, resp.AttestationData)
	evidence := task85DVerificationEvidence(t, ctx.ChainID(), account, scopeID, scopeID, metadata, att, veidtypes.AttestationTypeSocialMediaVerification, veidtypes.WebEvidenceActionSubmitSocial, map[string]string{
		"scope_id":                 scopeID,
		"provider":                 resp.Provider,
		"profile_name_hash":        resp.ProfileNameHash,
		"email_hash":               resp.SocialEmailHash,
		"username_hash":            resp.UsernameHash,
		"org_hash":                 resp.OrgHash,
		"account_created_at_unix":  strconv.FormatInt(resp.AccountCreatedAt, 10),
		"account_age_days":         strconv.FormatUint(uint64(resp.AccountAgeDays), 10),
		"is_verified":              strconv.FormatBool(resp.IsVerified),
		"friend_count_range":       resp.FriendCountRange,
		"encrypted_payload_digest": task85DEncryptedPayloadDigest(t, payload),
	})
	return &veidtypes.MsgSubmitSocialMediaScope{
		AccountAddress:         account,
		ScopeId:                scopeID,
		Provider:               veidtypes.SocialMediaProviderToProto(veidtypes.SocialMediaProviderType(resp.Provider)),
		ProfileNameHash:        resp.ProfileNameHash,
		EmailHash:              resp.SocialEmailHash,
		UsernameHash:           resp.UsernameHash,
		OrgHash:                resp.OrgHash,
		AccountCreatedAt:       resp.AccountCreatedAt,
		AccountAgeDays:         resp.AccountAgeDays,
		IsVerified:             resp.IsVerified,
		FriendCountRange:       resp.FriendCountRange,
		AttestationData:        resp.AttestationData,
		AccountSignature:       task85DAccountSignature(t, accountPriv, evidence),
		EncryptedPayload:       payload,
		EvidenceHash:           task85DSHA256HexBytes(resp.AttestationData),
		EvidenceStorageBackend: string(veidtypes.StorageBackendWaldur),
		EvidenceStorageRef:     "vault://task85d/social/" + scopeID,
		EvidenceMetadata:       metadata,
	}
}

func task85DSSOEvidence(t *testing.T, chainID string, account string, linkageID string, metadata map[string]string, att *veidtypes.SSOAttestation) veidtypes.WebEvidenceContext {
	t.Helper()
	canonical, err := att.CanonicalBytes()
	require.NoError(t, err)
	return veidtypes.NewWebEvidenceContext(veidtypes.WebEvidenceContextConfig{
		ChainID:             chainID,
		AccountAddress:      account,
		EvidenceType:        veidtypes.AttestationTypeSSOVerification,
		Action:              veidtypes.WebEvidenceActionSubmitSSO,
		ScopeID:             linkageID,
		AttestationDigest:   veidtypes.WebEvidenceDigestHex(canonical),
		Issuer:              att.Issuer,
		IssuerAlgorithm:     att.Proof.Type,
		Nonce:               att.Nonce,
		Challenge:           att.OIDCNonce,
		IssuedAt:            att.IssuedAt,
		ExpiresAt:           att.ExpiresAt,
		ServiceMetadataHash: task85DServiceMetadataHashForTest(t, metadata),
		CallerFields: map[string]string{
			"linkage_id":             linkageID,
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
}

func task85DVerificationEvidence(t *testing.T, chainID string, account string, scopeID string, challenge string, metadata map[string]string, att *veidtypes.VerificationAttestation, attType veidtypes.AttestationType, action string, callerFields map[string]string) veidtypes.WebEvidenceContext {
	t.Helper()
	digest, err := veidtypes.WebEvidenceAttestationDigestHex(att)
	require.NoError(t, err)
	return veidtypes.NewWebEvidenceContext(veidtypes.WebEvidenceContextConfig{
		ChainID:             chainID,
		AccountAddress:      account,
		EvidenceType:        attType,
		Action:              action,
		ScopeID:             scopeID,
		AttestationDigest:   digest,
		Issuer:              att.Issuer,
		IssuerAlgorithm:     att.Proof.Type,
		Nonce:               att.Nonce,
		Challenge:           challenge,
		IssuedAt:            att.IssuedAt,
		ExpiresAt:           att.ExpiresAt,
		ServiceMetadataHash: task85DServiceMetadataHashForTest(t, metadata),
		CallerFields:        callerFields,
	})
}

func task85DVerificationAttestation(t *testing.T, data []byte) *veidtypes.VerificationAttestation {
	t.Helper()
	att, err := veidtypes.AttestationFromJSON(data)
	require.NoError(t, err)
	return att
}

func task85DAccountSignature(t *testing.T, priv ed25519.PrivateKey, evidence veidtypes.WebEvidenceContext) []byte {
	t.Helper()
	signBytes, err := evidence.AccountAuthorizationBytes()
	require.NoError(t, err)
	digest := sha256.Sum256(signBytes)
	return ed25519.Sign(priv, digest[:])
}

func task85DEncryptedPayload(t *testing.T) veidtypes.EncryptedPayloadEnvelopePB {
	t.Helper()
	alg := encryptiontypes.DefaultAlgorithm()
	info, err := encryptiontypes.GetAlgorithmInfo(alg)
	require.NoError(t, err)
	recipientPubKey := bytes.Repeat([]byte{0x22}, encryptiontypes.X25519PublicKeySize)
	return veidtypes.EncryptedPayloadEnvelopePB{
		Version:             encryptiontypes.EnvelopeVersion,
		AlgorithmId:         alg,
		AlgorithmVersion:    info.Version,
		RecipientKeyIds:     []string{encryptiontypes.ComputeKeyFingerprint(recipientPubKey)},
		RecipientPublicKeys: [][]byte{recipientPubKey},
		EncryptedKeys:       [][]byte{[]byte("task85d-encrypted-key")},
		Nonce:               bytes.Repeat([]byte{0x33}, info.NonceSize),
		Ciphertext:          []byte("task85d-social-ciphertext"),
		SenderPubKey:        bytes.Repeat([]byte{0x44}, encryptiontypes.X25519PublicKeySize),
		SenderSignature:     []byte("task85d-sender-signature"),
	}
}

func task85DEncryptedPayloadDigest(t *testing.T, payload veidtypes.EncryptedPayloadEnvelopePB) string {
	t.Helper()
	converted := encryptiontypes.EncryptedPayloadEnvelope{
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
	}
	bz, err := json.Marshal(converted)
	require.NoError(t, err)
	return task85DSHA256HexBytes(bz)
}

func task85DVEIDStoreDigest(t *testing.T, ctx sdk.Context, k keeper.Keeper) string {
	t.Helper()
	store := ctx.KVStore(k.StoreKey())
	iter := store.Iterator(nil, nil)
	defer iter.Close()
	h := sha256.New()
	for ; iter.Valid(); iter.Next() {
		writeLenPrefixed(h, iter.Key())
		writeLenPrefixed(h, iter.Value())
	}
	return hex.EncodeToString(h.Sum(nil))
}

func writeLenPrefixed(w io.Writer, value []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = w.Write(length[:])
	_, _ = w.Write(value)
}

func task85DAccountAddress(seed string) sdk.AccAddress {
	sum := sha256.Sum256([]byte(seed))
	return sdk.AccAddress(sum[:20])
}

func task85DDeterministicEd25519Key(seed string) (ed25519.PublicKey, ed25519.PrivateKey) {
	sum := sha256.Sum256([]byte(seed))
	priv := ed25519.NewKeyFromSeed(sum[:])
	return priv.Public().(ed25519.PublicKey), priv
}

func task85DNonceHex(seed string) string {
	sum := sha256.Sum256([]byte("task85d-nonce:" + seed))
	return hex.EncodeToString(sum[:])
}

func task85DSHA256Hex(value string) string {
	return task85DSHA256HexBytes([]byte(value))
}

func task85DDigestBytes(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

func task85DSHA256HexBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func task85DServiceMetadataHash(metadata map[string]string) string {
	hash, err := veidtypes.WebEvidenceServiceMetadataHash(metadata)
	if err != nil {
		panic(err)
	}
	return hash
}

func task85DServiceMetadataHashForTest(t *testing.T, metadata map[string]string) string {
	t.Helper()
	hash, err := veidtypes.WebEvidenceServiceMetadataHash(metadata)
	require.NoError(t, err)
	return hash
}

func task85DJoin(values []string) string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	var buf bytes.Buffer
	for i, value := range out {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(value)
	}
	return buf.String()
}
