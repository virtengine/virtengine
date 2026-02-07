package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type mockLeaseQuery struct {
	lease *Lease
	err   error
	calls int
}

func (m *mockLeaseQuery) GetLease(_ context.Context, _ string) (*Lease, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.lease, nil
}

func TestVerifierVerifySuccess(t *testing.T) {
	const chainID = "virtengine-1"
	now := time.Now().UTC()
	verifier := NewVerifier(VerifierConfig{
		ChainID: chainID,
		Clock:   func() time.Time { return now },
	})

	priv := secp256k1.GenPrivKey()
	address := sdk.AccAddress(priv.PubKey().Address()).String()

	body := []byte(`{"action":"restart","metadata":{"reason":"test","count":1}}`)
	nonce := "nonce-123"

	req := buildSignedRequest(t, priv, chainID, address, "POST", "/api/v1/deployments/lease1/actions", body, nonce, now)

	signed, err := verifier.Verify(req)
	require.NoError(t, err)
	require.Equal(t, address, signed.Address)
	require.Equal(t, nonce, signed.Nonce)
	require.Equal(t, now.Truncate(time.Millisecond), signed.Timestamp)
	require.NotEmpty(t, signed.Signature)
	require.NotNil(t, signed.PubKey)

	_, err = verifier.Verify(req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "nonce already used")
}

func TestVerifierTimestampValidation(t *testing.T) {
	const chainID = "virtengine-1"
	now := time.Now().UTC()
	verifier := NewVerifier(VerifierConfig{
		ChainID: chainID,
		Clock:   func() time.Time { return now },
	})

	priv := secp256k1.GenPrivKey()
	address := sdk.AccAddress(priv.PubKey().Address()).String()
	body := []byte(`{"action":"restart"}`)

	oldTimestamp := now.Add(-DefaultMaxTimestampAge - time.Minute)
	req := buildSignedRequest(t, priv, chainID, address, "POST", "/api/v1/deployments/lease1/actions", body, "nonce-old", oldTimestamp)
	_, err := verifier.Verify(req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "timestamp too old")
}

func TestVerifierLeaseOwnershipCache(t *testing.T) {
	const chainID = "virtengine-1"
	now := time.Unix(1_700_000_000, 0).UTC()
	clock := now
	leaseQuery := &mockLeaseQuery{
		lease: &Lease{ID: "lease1", Owner: "owner1"},
	}

	verifier := NewVerifier(VerifierConfig{
		ChainID:           chainID,
		ChainQuerier:      leaseQuery,
		OwnershipCacheTTL: 15 * time.Minute,
		Clock:             func() time.Time { return clock },
	})

	err := verifier.VerifyLeaseOwnership(context.Background(), "owner1", "lease1")
	require.NoError(t, err)
	require.Equal(t, 1, leaseQuery.calls)

	err = verifier.VerifyLeaseOwnership(context.Background(), "owner1", "lease1")
	require.NoError(t, err)
	require.Equal(t, 1, leaseQuery.calls)

	err = verifier.VerifyLeaseOwnership(context.Background(), "other", "lease1")
	require.Error(t, err)
	require.Equal(t, 1, leaseQuery.calls)

	clock = clock.Add(16 * time.Minute)
	err = verifier.VerifyLeaseOwnership(context.Background(), "owner1", "lease1")
	require.NoError(t, err)
	require.Equal(t, 2, leaseQuery.calls)
}

func buildSignedRequest(
	t *testing.T,
	priv *secp256k1.PrivKey,
	chainID string,
	address string,
	method string,
	path string,
	body []byte,
	nonce string,
	timestamp time.Time,
) *http.Request {
	t.Helper()

	bodyHash := hashBodyForTest(body)

	requestData := RequestData{
		Method:    method,
		Path:      path,
		Timestamp: timestamp.UnixMilli(),
		Nonce:     nonce,
		BodyHash:  bodyHash,
	}

	dataToSign := serializeRequestData(requestData)
	signDoc := buildSignDoc(chainID, address, base64.StdEncoding.EncodeToString([]byte(dataToSign)))
	signBytes, err := canonicalJSON(signDoc)
	require.NoError(t, err)

	signature, err := priv.Sign([]byte(signBytes))
	require.NoError(t, err)

	pubKey := priv.PubKey().Bytes()

	req := httptest.NewRequest(method, "http://example.com"+path, bytes.NewReader(body))
	req.Header.Set(HeaderAddress, address)
	req.Header.Set(HeaderTimestamp, strconv.FormatInt(timestamp.UnixMilli(), 10))
	req.Header.Set(HeaderNonce, nonce)
	req.Header.Set(HeaderSignature, base64.StdEncoding.EncodeToString(signature))
	req.Header.Set(HeaderPubKey, base64.StdEncoding.EncodeToString(pubKey))
	req.Header.Set("Content-Type", "application/json")

	return req
}

func hashBodyForTest(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if canonical, ok := canonicalizeJSON(body); ok {
		sum := sha256.Sum256([]byte(canonical))
		return hex.EncodeToString(sum[:])
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
