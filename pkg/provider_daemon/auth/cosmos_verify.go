package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VerifierConfig configures the request verifier.
type VerifierConfig struct {
	ChainID         string
	NonceStore      NonceStore
	LeaseQuerier    LeaseOwnerQuerier
	MaxTimestampAge time.Duration
	AllowedTimeSkew time.Duration
}

// Verifier verifies wallet-signed requests.
type Verifier struct {
	chainID         string
	nonceStore      NonceStore
	leaseQuerier    LeaseOwnerQuerier
	maxTimestampAge time.Duration
	allowedSkew     time.Duration
}

// NewVerifier creates a new verifier instance.
func NewVerifier(cfg VerifierConfig) *Verifier {
	maxAge := cfg.MaxTimestampAge
	if maxAge <= 0 {
		maxAge = MaxTimestampAge
	}
	allowedSkew := cfg.AllowedTimeSkew
	if allowedSkew <= 0 {
		allowedSkew = AllowedFutureSkew
	}

	return &Verifier{
		chainID:         cfg.ChainID,
		nonceStore:      cfg.NonceStore,
		leaseQuerier:    cfg.LeaseQuerier,
		maxTimestampAge: maxAge,
		allowedSkew:     allowedSkew,
	}
}

// HasSignature returns true if wallet auth headers/query are present.
func HasSignature(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.Header.Get(HeaderAddress) != "" || r.URL.Query().Get(QueryAddress) != "" {
		return true
	}
	if r.Header.Get(HeaderPubKey) != "" || r.URL.Query().Get(QueryPubKey) != "" {
		return true
	}
	return false
}

// Verify validates a signed request and returns the parsed details.
func (v *Verifier) Verify(r *http.Request) (*SignedRequest, error) {
	if r == nil {
		return nil, errors.New("request is nil")
	}

	address, timestampStr, nonce, signatureB64, pubKeyB64 := extractAuthFields(r)
	if address == "" || timestampStr == "" || nonce == "" || signatureB64 == "" || pubKeyB64 == "" {
		return nil, errors.New("missing authentication headers")
	}

	timestampVal, timestampTime, err := parseTimestamp(timestampStr)
	if err != nil {
		return nil, err
	}

	if err := v.validateTimestamp(timestampTime); err != nil {
		return nil, err
	}

	if v.nonceStore != nil && v.nonceStore.HasSeen(nonce) {
		return nil, errors.New("nonce already used")
	}

	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding: %w", err)
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyB64)
	if err != nil {
		return nil, fmt.Errorf("invalid public key encoding: %w", err)
	}
	pubKey := &secp256k1.PubKey{Key: pubKeyBytes}

	addrBytes, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}
	if !bytes.Equal(addrBytes, pubKey.Address()) {
		return nil, errors.New("address does not match public key")
	}

	bodyHash, err := readAndHashBody(r)
	if err != nil {
		return nil, err
	}

	requestData := RequestData{
		Method:    strings.ToUpper(r.Method),
		Path:      r.URL.Path,
		Timestamp: timestampVal,
		Nonce:     nonce,
		BodyHash:  bodyHash,
	}

	payload, err := marshalRequestData(requestData)
	if err != nil {
		return nil, err
	}
	payloadB64 := base64.StdEncoding.EncodeToString(payload)

	signBytes, err := buildSignDocBytes(v.chainID, address, payloadB64, "")
	if err != nil {
		return nil, err
	}

	if !pubKey.VerifySignature(signBytes, signature) {
		return nil, errors.New("signature verification failed")
	}

	if v.nonceStore != nil {
		v.nonceStore.MarkSeen(nonce, timestampTime.Add(v.maxTimestampAge))
	}

	return &SignedRequest{
		Address:   address,
		Timestamp: timestampTime,
		Nonce:     nonce,
		Signature: signature,
		PubKey:    pubKey,
		BodyHash:  bodyHash,
	}, nil
}

// VerifyLeaseOwnership checks if the address owns the lease.
func (v *Verifier) VerifyLeaseOwnership(ctx context.Context, address string, leaseID string) error {
	if v.leaseQuerier == nil {
		return errors.New("lease ownership query not configured")
	}
	lease, err := ParseLeaseID(leaseID)
	if err != nil {
		return fmt.Errorf("invalid lease id: %w", err)
	}
	owner, err := v.leaseQuerier.LeaseOwner(ctx, lease)
	if err != nil {
		return fmt.Errorf("failed to query lease: %w", err)
	}
	if owner != address {
		return errors.New("address does not own this lease")
	}
	return nil
}

func extractAuthFields(r *http.Request) (string, string, string, string, string) {
	address := firstNonEmpty(r.Header.Get(HeaderAddress), r.URL.Query().Get(QueryAddress))
	timestamp := firstNonEmpty(r.Header.Get(HeaderTimestamp), r.URL.Query().Get(QueryTimestamp))
	nonce := firstNonEmpty(r.Header.Get(HeaderNonce), r.URL.Query().Get(QueryNonce))
	signature := firstNonEmpty(r.Header.Get(HeaderSignature), r.URL.Query().Get(QuerySignature))
	pubKey := firstNonEmpty(r.Header.Get(HeaderPubKey), r.URL.Query().Get(QueryPubKey))
	return address, timestamp, nonce, signature, pubKey
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseTimestamp(raw string) (int64, time.Time, error) {
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("invalid timestamp: %w", err)
	}
	if val == 0 {
		return 0, time.Time{}, errors.New("invalid timestamp")
	}
	if val < 1_000_000_000_000 {
		// treat as seconds
		return val, time.Unix(val, 0), nil
	}
	return val, time.UnixMilli(val), nil
}

func (v *Verifier) validateTimestamp(ts time.Time) error {
	now := time.Now()
	if now.Sub(ts) > v.maxTimestampAge {
		return errors.New("request timestamp too old")
	}
	if ts.After(now.Add(v.allowedSkew)) {
		return errors.New("request timestamp in future")
	}
	return nil
}

func readAndHashBody(r *http.Request) (string, error) {
	if r.Body == nil {
		return "", nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return "", nil
	}

	normalized := body
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if sorted, err := sdk.SortJSON(body); err == nil {
			normalized = sorted
		}
	}

	hash := sha256.Sum256(normalized)
	return hex.EncodeToString(hash[:]), nil
}

func marshalRequestData(data RequestData) ([]byte, error) {
	return json.Marshal(data)
}

type aminoFee struct {
	Gas    string      `json:"gas"`
	Amount []aminoCoin `json:"amount"`
}

type aminoCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type aminoMsg struct {
	Type  string      `json:"type"`
	Value aminoMsgVal `json:"value"`
}

type aminoMsgVal struct {
	Signer string `json:"signer"`
	Data   string `json:"data"`
}

type aminoSignDoc struct {
	ChainID       string     `json:"chain_id"`
	AccountNumber string     `json:"account_number"`
	Sequence      string     `json:"sequence"`
	Fee           aminoFee   `json:"fee"`
	Msgs          []aminoMsg `json:"msgs"`
	Memo          string     `json:"memo"`
}

func buildSignDocBytes(chainID, signer, dataB64, memo string) ([]byte, error) {
	if chainID == "" {
		return nil, errors.New("chain id is required")
	}

	signDoc := aminoSignDoc{
		ChainID:       chainID,
		AccountNumber: "0",
		Sequence:      "0",
		Fee: aminoFee{
			Gas:    "0",
			Amount: []aminoCoin{},
		},
		Msgs: []aminoMsg{
			{
				Type: "sign/MsgSignData",
				Value: aminoMsgVal{
					Signer: signer,
					Data:   dataB64,
				},
			},
		},
		Memo: memo,
	}

	payload, err := json.Marshal(signDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sign doc: %w", err)
	}
	return sdk.SortJSON(payload)
}
