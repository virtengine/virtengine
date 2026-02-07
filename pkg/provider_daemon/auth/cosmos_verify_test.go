package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	marketv1 "github.com/virtengine/virtengine/sdk/go/node/market/v1"
)

type mockChainQuerier struct {
	lease marketv1.Lease
	err   error
}

func (m mockChainQuerier) GetLease(ctx context.Context, leaseID marketv1.LeaseID) (marketv1.Lease, error) {
	return m.lease, m.err
}

func TestVerifyWalletSignature(t *testing.T) {
	nonceStore := NewInMemoryNonceStore()
	now := time.Date(2026, 2, 7, 12, 0, 0, 0, time.UTC)
	chainID := "virtengine-1"

	verifier := NewVerifier(nonceStore, nil, VerifierOptions{
		ChainID: chainID,
		Now:     func() time.Time { return now },
	})

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey().(*secp256k1.PubKey)
	address := sdk.AccAddress(pubKey.Address()).String()

	body := []byte(`{"action":"start"}`)
	hash := sha256.Sum256(body)
	bodyHash := hex.EncodeToString(hash[:])
	timestamp := now.UnixMilli()

	requestData := RequestData{
		Method:    "POST",
		Path:      "/api/v1/deployments/owner/1/1/1/provider/actions",
		Timestamp: timestamp,
		Nonce:     "deadbeefcafefeed",
		BodyHash:  bodyHash,
	}

	requestJSON, err := serializeRequestData(requestData)
	if err != nil {
		t.Fatalf("serialize request: %v", err)
	}
	dataBase64 := base64.StdEncoding.EncodeToString([]byte(requestJSON))
	signDocBytes, err := buildADR036SignDocBytes(chainID, address, dataBase64)
	if err != nil {
		t.Fatalf("build sign doc: %v", err)
	}

	signature, err := privKey.Sign(signDocBytes)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://provider.example.com/api/v1/deployments/owner/1/1/1/provider/actions", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	req.Header.Set(HeaderAddress, address)
	req.Header.Set(HeaderTimestamp, strconvFormatInt(timestamp))
	req.Header.Set(HeaderNonce, requestData.Nonce)
	req.Header.Set(HeaderSignature, base64.StdEncoding.EncodeToString(signature))
	req.Header.Set(HeaderPubKey, base64.StdEncoding.EncodeToString(pubKey.Bytes()))

	signed, err := verifier.Verify(req)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if signed.Address != address {
		t.Fatalf("expected address %s got %s", address, signed.Address)
	}
}

func TestVerifyRejectsReplay(t *testing.T) {
	nonceStore := NewInMemoryNonceStore()
	now := time.Date(2026, 2, 7, 12, 0, 0, 0, time.UTC)
	chainID := "virtengine-1"
	verifier := NewVerifier(nonceStore, nil, VerifierOptions{
		ChainID: chainID,
		Now:     func() time.Time { return now },
	})

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey().(*secp256k1.PubKey)
	address := sdk.AccAddress(pubKey.Address()).String()
	timestamp := now.UnixMilli()

	requestData := RequestData{
		Method:    "GET",
		Path:      "/api/v1/deployments/owner/1/1/1/provider/logs",
		Timestamp: timestamp,
		Nonce:     "replaynonce",
		BodyHash:  "",
	}

	requestJSON, err := serializeRequestData(requestData)
	if err != nil {
		t.Fatalf("serialize request: %v", err)
	}
	dataBase64 := base64.StdEncoding.EncodeToString([]byte(requestJSON))
	signDocBytes, err := buildADR036SignDocBytes(chainID, address, dataBase64)
	if err != nil {
		t.Fatalf("build sign doc: %v", err)
	}

	signature, err := privKey.Sign(signDocBytes)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	makeRequest := func() *http.Request {
		req, err := http.NewRequest(http.MethodGet, "https://provider.example.com/api/v1/deployments/owner/1/1/1/provider/logs", nil)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		req.Header.Set(HeaderAddress, address)
		req.Header.Set(HeaderTimestamp, strconvFormatInt(timestamp))
		req.Header.Set(HeaderNonce, requestData.Nonce)
		req.Header.Set(HeaderSignature, base64.StdEncoding.EncodeToString(signature))
		req.Header.Set(HeaderPubKey, base64.StdEncoding.EncodeToString(pubKey.Bytes()))
		return req
	}

	if _, err := verifier.Verify(makeRequest()); err != nil {
		t.Fatalf("first verify failed: %v", err)
	}
	if _, err := verifier.Verify(makeRequest()); err == nil {
		t.Fatalf("expected replay to fail")
	}
}

func TestVerifyLeaseOwnership(t *testing.T) {
	ownerKey := secp256k1.GenPrivKey().PubKey().(*secp256k1.PubKey)
	providerKey := secp256k1.GenPrivKey().PubKey().(*secp256k1.PubKey)
	otherKey := secp256k1.GenPrivKey().PubKey().(*secp256k1.PubKey)
	owner := sdk.AccAddress(ownerKey.Address()).String()
	provider := sdk.AccAddress(providerKey.Address()).String()
	other := sdk.AccAddress(otherKey.Address()).String()
	lease := marketv1.Lease{
		ID: marketv1.LeaseID{
			Owner:    owner,
			DSeq:     1,
			GSeq:     1,
			OSeq:     1,
			Provider: provider,
		},
	}
	querier := mockChainQuerier{lease: lease}
	verifier := NewVerifier(NewInMemoryNonceStore(), querier, VerifierOptions{
		ChainID: "virtengine-1",
	})

	if err := verifier.VerifyLeaseOwnership(context.Background(), owner, fmt.Sprintf("lease/%s/1/1/1/%s", owner, provider)); err != nil {
		t.Fatalf("expected ownership ok: %v", err)
	}
	if err := verifier.VerifyLeaseOwnership(context.Background(), other, fmt.Sprintf("%s/1/1/1/%s", owner, provider)); err == nil {
		t.Fatalf("expected ownership mismatch")
	}
}

func strconvFormatInt(value int64) string {
	return fmt.Sprintf("%d", value)
}
