package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
)

type stubLeaseQuerier struct {
	owner string
}

func (s stubLeaseQuerier) LeaseOwner(_ context.Context, _ marketv1.LeaseID) (string, error) {
	return s.owner, nil
}

func TestInMemoryNonceStore(t *testing.T) {
	store := NewInMemoryNonceStore()
	nonce := "nonce-1"
	store.MarkSeen(nonce, time.Now().Add(time.Minute))
	if !store.HasSeen(nonce) {
		t.Fatalf("expected nonce to be seen")
	}
	store.MarkSeen("expired", time.Now().Add(-time.Minute))
	if store.HasSeen("expired") {
		t.Fatalf("expected expired nonce to be cleared")
	}
	store.Stop()
}

func TestVerifySignedRequest(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	chainID := "virtengine-test"

	body := []byte(`{"b":2,"a":1}`)
	sorted, err := sdk.SortJSON(body)
	if err != nil {
		t.Fatalf("failed to sort json: %v", err)
	}
	bodyHash := sha256.Sum256(sorted)
	timestamp := time.Now().UTC()
	timestampMs := timestamp.UnixMilli()

	requestData := RequestData{
		Method:    "POST",
		Path:      "/api/v1/deployments/lease-1/logs",
		Timestamp: timestampMs,
		Nonce:     "nonce-abc",
		BodyHash:  hex.EncodeToString(bodyHash[:]),
	}

	payload, err := marshalRequestData(requestData)
	if err != nil {
		t.Fatalf("failed to marshal request data: %v", err)
	}
	payloadB64 := base64.StdEncoding.EncodeToString(payload)

	signBytes, err := buildSignDocBytes(chainID, address, payloadB64, "")
	if err != nil {
		t.Fatalf("failed to build sign doc: %v", err)
	}

	signature, err := privKey.Sign(signBytes)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	req := httptest.NewRequest("POST", "https://provider.example.com/api/v1/deployments/lease-1/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderAddress, address)
	req.Header.Set(HeaderTimestamp, strconv.FormatInt(timestampMs, 10))
	req.Header.Set(HeaderNonce, requestData.Nonce)
	req.Header.Set(HeaderSignature, base64.StdEncoding.EncodeToString(signature))
	req.Header.Set(HeaderPubKey, base64.StdEncoding.EncodeToString(privKey.PubKey().Bytes()))

	verifier := NewVerifier(VerifierConfig{
		ChainID:         chainID,
		NonceStore:      NewInMemoryNonceStore(),
		MaxTimestampAge: time.Minute * 10,
	})

	signed, err := verifier.Verify(req)
	if err != nil {
		t.Fatalf("expected verification success, got %v", err)
	}
	if signed.Address != address {
		t.Fatalf("expected address %s, got %s", address, signed.Address)
	}

	// replay should fail
	reqReplay := httptest.NewRequest("POST", "https://provider.example.com/api/v1/deployments/lease-1/logs", bytes.NewReader(body))
	reqReplay.Header = req.Header.Clone()
	if _, err := verifier.Verify(reqReplay); err == nil {
		t.Fatalf("expected replay to be rejected")
	}
}

func TestVerifyLeaseOwnership(t *testing.T) {
	ownerKey := secp256k1.GenPrivKey()
	providerKey := secp256k1.GenPrivKey()
	owner := sdk.AccAddress(ownerKey.PubKey().Address()).String()
	provider := sdk.AccAddress(providerKey.PubKey().Address()).String()

	verifier := NewVerifier(VerifierConfig{
		ChainID:      "virtengine-test",
		LeaseQuerier: stubLeaseQuerier{owner: owner},
	})

	leaseID := owner + "/1/1/1/" + provider
	if err := verifier.VerifyLeaseOwnership(context.Background(), owner, leaseID); err != nil {
		t.Fatalf("expected lease ownership to pass: %v", err)
	}
	if err := verifier.VerifyLeaseOwnership(context.Background(), "cosmos1other", leaseID); err == nil {
		t.Fatalf("expected lease ownership to fail")
	}
}
