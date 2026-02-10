package identity_scopes

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

type mockChainClient struct {
	lastSSO    *veidtypes.MsgSubmitSSOVerificationProof
	lastEmail  *veidtypes.MsgSubmitEmailVerificationProof
	lastSMS    *veidtypes.MsgSubmitSMSVerificationProof
	lastSocial *veidtypes.MsgSubmitSocialMediaScope
}

func (m *mockChainClient) SubmitSSOVerificationProof(_ context.Context, msg *veidtypes.MsgSubmitSSOVerificationProof) error {
	m.lastSSO = msg
	return nil
}

func (m *mockChainClient) SubmitEmailVerificationProof(_ context.Context, msg *veidtypes.MsgSubmitEmailVerificationProof) error {
	m.lastEmail = msg
	return nil
}

func (m *mockChainClient) SubmitSMSVerificationProof(_ context.Context, msg *veidtypes.MsgSubmitSMSVerificationProof) error {
	m.lastSMS = msg
	return nil
}

func (m *mockChainClient) SubmitSocialMediaScope(_ context.Context, msg *veidtypes.MsgSubmitSocialMediaScope) error {
	m.lastSocial = msg
	return nil
}

func TestSSOAdapter_SubmitProof(t *testing.T) {
	chain := &mockChainClient{}
	adapter := NewGoogleSSOAdapter(chain)

	issuer := veidtypes.AttestationIssuer{
		ID:             "did:virtengine:signer:google",
		KeyFingerprint: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	subject := veidtypes.AttestationSubject{
		ID:             "did:virtengine:cosmos1abc",
		AccountAddress: "cosmos1abc",
	}
	att := veidtypes.NewSSOAttestation(
		issuer,
		subject,
		"https://accounts.google.com",
		"subject",
		veidtypes.SSOProviderGoogle,
		"nonce",
		[]byte("nonce-bytes-32---------------"),
		time.Now(),
		time.Hour,
	)

	_, err := adapter.SubmitProof(context.Background(), SSOProofRequest{
		AccountAddress:     "cosmos1abc",
		LinkageID:          "link-1",
		Attestation:        att,
		EvidenceStorageRef: "vault://evidence/1",
	})
	require.NoError(t, err)
	require.NotNil(t, chain.lastSSO)
	require.Equal(t, "link-1", chain.lastSSO.LinkageId)
	require.Equal(t, "cosmos1abc", chain.lastSSO.AccountAddress)
}

func TestEmailAdapter_SubmitProof(t *testing.T) {
	chain := &mockChainClient{}
	adapter := NewEmailOTPAdapter(chain)

	issuer := veidtypes.AttestationIssuer{
		ID:             "did:virtengine:signer:email",
		KeyFingerprint: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
	subject := veidtypes.AttestationSubject{
		ID:             "did:virtengine:cosmos1abc",
		AccountAddress: "cosmos1abc",
	}
	att := veidtypes.NewVerificationAttestation(issuer, subject, veidtypes.AttestationTypeEmailVerification, []byte("nonce-email-32---------------"), time.Now(), time.Hour, 100, 100)
	att.SetProof(veidtypes.AttestationProof{
		Type:               veidtypes.ProofTypeEd25519,
		Created:            time.Now(),
		VerificationMethod: "did:virtengine:signer:email#keys-1",
		ProofPurpose:       "assertionMethod",
		ProofValue:         base64.StdEncoding.EncodeToString([]byte("sig")),
		Nonce:              "nonce",
	})

	_, err := adapter.SubmitProof(context.Background(), EmailOTPProofRequest{
		AccountAddress:     "cosmos1abc",
		VerificationID:     "email-1",
		EmailHash:          "hash",
		Nonce:              "nonce",
		Attestation:        att,
		EvidenceStorageRef: "vault://email/1",
	})
	require.NoError(t, err)
	require.NotNil(t, chain.lastEmail)
	require.Equal(t, "email-1", chain.lastEmail.VerificationId)
}

func TestSMSAdapter_SubmitProof(t *testing.T) {
	chain := &mockChainClient{}
	adapter := NewSMSOTPAdapter(chain)

	issuer := veidtypes.AttestationIssuer{
		ID:             "did:virtengine:signer:sms",
		KeyFingerprint: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	}
	subject := veidtypes.AttestationSubject{
		ID:             "did:virtengine:cosmos1abc",
		AccountAddress: "cosmos1abc",
	}
	att := veidtypes.NewVerificationAttestation(issuer, subject, veidtypes.AttestationTypeSMSVerification, []byte("nonce-sms-32-----------------"), time.Now(), time.Hour, 100, 100)
	att.SetProof(veidtypes.AttestationProof{
		Type:               veidtypes.ProofTypeEd25519,
		Created:            time.Now(),
		VerificationMethod: "did:virtengine:signer:sms#keys-1",
		ProofPurpose:       "assertionMethod",
		ProofValue:         base64.StdEncoding.EncodeToString([]byte("sig")),
		Nonce:              "nonce",
	})

	_, err := adapter.SubmitProof(context.Background(), SMSOTPProofRequest{
		AccountAddress:     "cosmos1abc",
		VerificationID:     "sms-1",
		PhoneHash:          "hash",
		PhoneHashSalt:      "salt",
		Attestation:        att,
		EvidenceStorageRef: "vault://sms/1",
	})
	require.NoError(t, err)
	require.NotNil(t, chain.lastSMS)
	require.Equal(t, "sms-1", chain.lastSMS.VerificationId)
}

func TestSocialMediaAdapter_SubmitScope(t *testing.T) {
	chain := &mockChainClient{}
	adapter := NewGoogleSocialMediaAdapter(chain)

	issuer := veidtypes.AttestationIssuer{
		ID:             "did:virtengine:signer:social",
		KeyFingerprint: "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
	}
	subject := veidtypes.AttestationSubject{
		ID:             "did:virtengine:cosmos1abc",
		AccountAddress: "cosmos1abc",
	}
	att := veidtypes.NewVerificationAttestation(
		issuer,
		subject,
		veidtypes.AttestationTypeSocialMediaVerification,
		[]byte("nonce-social-32------------"),
		time.Now(),
		time.Hour,
		95,
		90,
	)
	att.SetProof(veidtypes.AttestationProof{
		Type:               veidtypes.ProofTypeEd25519,
		Created:            time.Now(),
		VerificationMethod: "did:virtengine:signer:social#keys-1",
		ProofPurpose:       "assertionMethod",
		ProofValue:         base64.StdEncoding.EncodeToString([]byte("sig")),
		Nonce:              "nonce",
	})

	payload := veidtypes.EncryptedPayloadEnvelope{
		Version:          1,
		AlgorithmId:      "x25519-xsalsa20-poly1305",
		AlgorithmVersion: 1,
		Nonce:            []byte("nonce-nonce-nonce-000000"),
		Ciphertext:       []byte("cipher"),
		SenderPubKey:     []byte("sender-pub-key"),
		SenderSignature:  []byte("sender-sig"),
	}

	_, err := adapter.SubmitScope(context.Background(), SocialMediaScopeRequest{
		AccountAddress:     "cosmos1abc",
		ScopeID:            "social-1",
		Provider:           veidtypes.SocialMediaProviderGoogle,
		ProfileName:        "Jane Doe",
		Email:              "jane@example.com",
		AccountAgeDays:     365,
		IsVerified:         true,
		Attestation:        att,
		AccountSignature:   []byte("account-sig"),
		EncryptedPayload:   payload,
		EvidenceStorageRef: "vault://social/1",
	})
	require.NoError(t, err)
	require.NotNil(t, chain.lastSocial)
	require.Equal(t, "social-1", chain.lastSocial.ScopeId)
	require.Equal(t, "cosmos1abc", chain.lastSocial.AccountAddress)
}
