package sms

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/identity_scopes"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type fakeSMSChainBackend struct {
	submittedSMSMsg *veidtypes.MsgSubmitSMSVerificationProof
	queryResponse   *veidv1.QuerySMSVerificationResponse
	submitErr       error
	queryErr        error
	closed          bool
}

func (f *fakeSMSChainBackend) SubmitSSOVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitSSOVerificationProof) error {
	return nil
}

func (f *fakeSMSChainBackend) SubmitEmailVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitEmailVerificationProof) error {
	return nil
}

func (f *fakeSMSChainBackend) SubmitSMSVerificationProof(ctx context.Context, msg *veidtypes.MsgSubmitSMSVerificationProof) error {
	cloned := *msg
	cloned.EvidenceMetadata = cloneSMSStringMap(msg.EvidenceMetadata)
	f.submittedSMSMsg = &cloned
	if f.submitErr != nil {
		return f.submitErr
	}
	if f.queryResponse == nil {
		now := time.Now().Unix()
		expiresAt := msg.ExpiresAt
		if expiresAt == 0 {
			expiresAt = time.Now().Add(365 * 24 * time.Hour).Unix()
		}
		f.queryResponse = &veidv1.QuerySMSVerificationResponse{
			Record: &veidv1.SMSVerificationRecord{
				VerificationId:         msg.VerificationId,
				AccountAddress:         msg.AccountAddress,
				PhoneHash:              &veidv1.PhoneNumberHash{Hash: msg.PhoneHash, Salt: msg.PhoneHashSalt, CountryCodeHash: msg.CountryCodeHash, CreatedAt: now},
				Status:                 string(veidtypes.SMSStatusVerified),
				VerifiedAt:             msg.VerifiedAt,
				ExpiresAt:              expiresAt,
				CreatedAt:              now,
				UpdatedAt:              now,
				IsVoip:                 msg.IsVoip,
				CarrierType:            msg.CarrierType,
				ValidatorAddress:       msg.ValidatorAddress,
				AccountSignature:       append([]byte(nil), msg.AccountSignature...),
				EvidenceStorageBackend: msg.EvidenceStorageBackend,
				EvidenceStorageRef:     msg.EvidenceStorageRef,
				EvidenceMetadata:       cloneSMSStringMap(msg.EvidenceMetadata),
			},
		}
	}
	return nil
}

func (f *fakeSMSChainBackend) SubmitSocialMediaScope(ctx context.Context, msg *veidtypes.MsgSubmitSocialMediaScope) error {
	return nil
}

func (f *fakeSMSChainBackend) QuerySMSVerification(ctx context.Context, req *veidv1.QuerySMSVerificationRequest) (*veidv1.QuerySMSVerificationResponse, error) {
	if f.queryErr != nil {
		return nil, f.queryErr
	}
	return f.queryResponse, nil
}

func (f *fakeSMSChainBackend) Close() error {
	f.closed = true
	return nil
}

func newTestSMSChainIntegrator(config ChainIntegrationConfig, backend *fakeSMSChainBackend) *DefaultChainIntegrator {
	if config.VerificationExpiryDays == 0 {
		config.VerificationExpiryDays = 365
	}
	return &DefaultChainIntegrator{
		config:       config,
		logger:       zerolog.Nop(),
		records:      make(map[string]*OnChainVerificationRecord),
		accountIndex: make(map[string][]string),
		pendingBatch: make([]*RecordVerificationRequest, 0),
		backend:      backend,
		adapter:      identity_scopes.NewSMSOTPAdapter(backend),
	}
}

func newTestSMSVerificationAttestation(accountAddress, attestationID, validatorAddress string, issuedAt time.Time) *veidtypes.VerificationAttestation {
	return &veidtypes.VerificationAttestation{
		ID:        attestationID,
		Type:      veidtypes.AttestationTypeSMSVerification,
		Issuer:    veidtypes.AttestationIssuer{ValidatorAddress: validatorAddress},
		Subject:   veidtypes.AttestationSubject{AccountAddress: accountAddress},
		IssuedAt:  issuedAt,
		ExpiresAt: issuedAt.Add(365 * 24 * time.Hour),
		Metadata:  make(map[string]string),
	}
}

func cloneSMSStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(in))
	for key, value := range in {
		cloned[key] = value
	}
	return cloned
}

func TestRecordVerification_SubmitsRealProof(t *testing.T) {
	t.Parallel()

	backend := &fakeSMSChainBackend{}
	integrator := newTestSMSChainIntegrator(ChainIntegrationConfig{
		VerificationExpiryDays: 365,
		MaxRetries:             0,
		RetryDelay:             time.Millisecond,
	}, backend)

	issuedAt := time.Unix(1710000000, 0).UTC()
	attestation := newTestSMSVerificationAttestation("virtengine1account", "veid:attestation:sms:1", "virtenginevaloper1validator", issuedAt)

	resp, err := integrator.RecordVerification(context.Background(), RecordVerificationRequest{
		AccountAddress:         "virtengine1account",
		ChallengeID:            "challenge-1",
		VerificationID:         "verification-1",
		PhoneHash:              "phone-hash-1",
		PhoneHashSalt:          "phone-salt-1",
		CountryCodeHash:        "country-hash-1",
		CountryCode:            "US",
		CarrierType:            CarrierTypeMobile,
		IsVoIP:                 false,
		RiskScore:              9,
		VerifiedAt:             issuedAt,
		ValidatorAddress:       "virtenginevaloper1validator",
		AccountSignature:       []byte("account-signature"),
		Attestation:            attestation,
		EvidenceStorageRef:     "vault://sms/verification-1",
		EvidenceStorageBackend: "s3",
		EvidenceMetadata: map[string]string{
			"attestation_id": attestation.ID,
			"trace_id":       "trace-123",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, backend.submittedSMSMsg)
	require.NotNil(t, resp)

	assert.Equal(t, "verification-1", resp.VerificationID)
	assert.Equal(t, "phone-hash-1", backend.submittedSMSMsg.PhoneHash)
	assert.Equal(t, "phone-salt-1", backend.submittedSMSMsg.PhoneHashSalt)
	assert.Equal(t, "country-hash-1", backend.submittedSMSMsg.CountryCodeHash)
	assert.Equal(t, []byte("account-signature"), backend.submittedSMSMsg.AccountSignature)
	assert.Equal(t, "vault://sms/verification-1", backend.submittedSMSMsg.EvidenceStorageRef)
	assert.Equal(t, "trace-123", backend.submittedSMSMsg.EvidenceMetadata["trace_id"])

	record, err := integrator.GetVerificationRecord(context.Background(), "virtengine1account", "verification-1")
	require.NoError(t, err)
	require.NotNil(t, record)
	assert.Equal(t, "virtengine1account", record.AccountAddress)
	assert.Equal(t, "country-hash-1", record.CountryCodeHash)
	assert.Equal(t, attestation.ID, record.AttestationID)
}

func TestGetVerificationRecord_QueriesBackend(t *testing.T) {
	t.Parallel()

	backend := &fakeSMSChainBackend{
		queryResponse: &veidv1.QuerySMSVerificationResponse{
			Record: &veidv1.SMSVerificationRecord{
				VerificationId:   "verification-lookup",
				AccountAddress:   "virtengine1lookup",
				PhoneHash:        &veidv1.PhoneNumberHash{Hash: "phone-hash-lookup", CountryCodeHash: "country-hash-lookup"},
				Status:           string(veidtypes.SMSStatusVerified),
				VerifiedAt:       time.Unix(1710000100, 0).UTC().Unix(),
				ExpiresAt:        time.Unix(1710864100, 0).UTC().Unix(),
				CreatedAt:        time.Unix(1710000000, 0).UTC().Unix(),
				UpdatedAt:        time.Unix(1710000200, 0).UTC().Unix(),
				IsVoip:           true,
				CarrierType:      string(CarrierTypeVoIP),
				ValidatorAddress: "virtenginevaloper1lookup",
				EvidenceMetadata: map[string]string{"attestation_id": "veid:attestation:sms:lookup"},
			},
		},
	}
	integrator := newTestSMSChainIntegrator(ChainIntegrationConfig{VerificationExpiryDays: 365}, backend)

	record, err := integrator.GetVerificationRecord(context.Background(), "virtengine1lookup", "verification-lookup")
	require.NoError(t, err)
	require.NotNil(t, record)

	assert.Equal(t, "verification-lookup", record.VerificationID)
	assert.Equal(t, "virtengine1lookup", record.AccountAddress)
	assert.Equal(t, CarrierTypeVoIP, record.CarrierType)
	assert.True(t, record.IsVoIP)
	assert.Equal(t, "veid:attestation:sms:lookup", record.AttestationID)

	cached, ok := integrator.records["verification-lookup"]
	require.True(t, ok)
	assert.Equal(t, record.VerificationID, cached.VerificationID)
}

func TestUpdateVerificationStatus_FailsClosedWhenLive(t *testing.T) {
	t.Parallel()

	integrator := newTestSMSChainIntegrator(ChainIntegrationConfig{OfflineMode: false}, &fakeSMSChainBackend{})
	err := integrator.UpdateVerificationStatus(context.Background(), UpdateStatusRequest{
		AccountAddress: "virtengine1account",
		VerificationID: "verification-1",
		NewStatus:      StatusRevoked,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestRevokeVerification_FailsClosedWhenLive(t *testing.T) {
	t.Parallel()

	integrator := newTestSMSChainIntegrator(ChainIntegrationConfig{OfflineMode: false}, &fakeSMSChainBackend{})
	err := integrator.RevokeVerification(context.Background(), RevokeVerificationRequest{
		AccountAddress: "virtengine1account",
		VerificationID: "verification-1",
		Reason:         "compromised",
		RevokedBy:      "validator",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestSubmitAttestation_UsesMetadata(t *testing.T) {
	t.Parallel()

	backend := &fakeSMSChainBackend{}
	integrator := newTestSMSChainIntegrator(ChainIntegrationConfig{
		VerificationExpiryDays: 365,
		MaxRetries:             0,
		RetryDelay:             time.Millisecond,
	}, backend)

	issuedAt := time.Unix(1710000300, 0).UTC()
	attestation := newTestSMSVerificationAttestation("virtengine1account", "veid:attestation:sms:2", "virtenginevaloper1validator", issuedAt)
	attestation.Metadata = map[string]string{
		"sms_verification_id":       "verification-2",
		"sms_challenge_id":          "challenge-2",
		"sms_phone_hash":            "phone-hash-2",
		"sms_phone_hash_salt":       "phone-salt-2",
		"sms_country_code_hash":     "country-hash-2",
		"sms_account_signature_b64": base64.StdEncoding.EncodeToString([]byte("sig-2")),
		"evidence_storage_ref":      "vault://sms/verification-2",
		"evidence_storage_backend":  "s3",
		"country_code":              "US",
		"carrier_type":              string(CarrierTypeVoIP),
		"is_voip":                   "true",
		"risk_score":                "77",
		"attestation_id":            attestation.ID,
	}

	err := integrator.SubmitAttestation(context.Background(), attestation)
	require.NoError(t, err)
	require.NotNil(t, backend.submittedSMSMsg)

	assert.Equal(t, "verification-2", backend.submittedSMSMsg.VerificationId)
	assert.Equal(t, "phone-hash-2", backend.submittedSMSMsg.PhoneHash)
	assert.Equal(t, "phone-salt-2", backend.submittedSMSMsg.PhoneHashSalt)
	assert.Equal(t, "country-hash-2", backend.submittedSMSMsg.CountryCodeHash)
	assert.Equal(t, []byte("sig-2"), backend.submittedSMSMsg.AccountSignature)
	assert.Equal(t, "vault://sms/verification-2", backend.submittedSMSMsg.EvidenceStorageRef)
}
