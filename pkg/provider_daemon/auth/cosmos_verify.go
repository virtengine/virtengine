package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

const (
	HeaderAddress   = "X-VE-Address"
	HeaderTimestamp = "X-VE-Timestamp"
	HeaderNonce     = "X-VE-Nonce"
	HeaderSignature = "X-VE-Signature"
	HeaderPubKey    = "X-VE-PubKey"

	QueryAddress   = "ve_address"
	QueryTimestamp = "ve_ts"
	QueryNonce     = "ve_nonce"
	QuerySignature = "ve_sig"
	QueryPubKey    = "ve_pub"
)

type SignedRequest struct {
	Address   string
	Timestamp time.Time
	Nonce     string
	Signature []byte
	PubKey    *secp256k1.PubKey
	Request   RequestData
}

type RequestData struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	BodyHash  string `json:"body_hash"`
}

type VerifierOptions struct {
	MaxTimestampAge time.Duration
	MaxFutureSkew   time.Duration
	LeaseCacheTTL   time.Duration
	ChainID         string
	Now             func() time.Time
}

type Verifier struct {
	nonceStore      NonceStore
	chainQuery      ChainQuerier
	maxAge          time.Duration
	maxFutureSkew   time.Duration
	now             func() time.Time
	leaseOwnerCache *LeaseOwnerCache
	chainID         string
}

func NewVerifier(nonceStore NonceStore, chainQuery ChainQuerier, opts VerifierOptions) *Verifier {
	if nonceStore == nil {
		nonceStore = NewInMemoryNonceStore()
	}
	nowFn := opts.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	maxAge := opts.MaxTimestampAge
	if maxAge == 0 {
		maxAge = 5 * time.Minute
	}
	maxFuture := opts.MaxFutureSkew
	if maxFuture == 0 {
		maxFuture = time.Minute
	}
	cache := NewLeaseOwnerCache(opts.LeaseCacheTTL)
	return &Verifier{
		nonceStore:      nonceStore,
		chainQuery:      chainQuery,
		maxAge:          maxAge,
		maxFutureSkew:   maxFuture,
		now:             nowFn,
		leaseOwnerCache: cache,
		chainID:         opts.ChainID,
	}
}

func HasWalletHeaders(r *http.Request) bool {
	return headerOrQuery(r, HeaderAddress, QueryAddress) != "" ||
		headerOrQuery(r, HeaderSignature, QuerySignature) != "" ||
		headerOrQuery(r, HeaderPubKey, QueryPubKey) != ""
}

func (v *Verifier) Verify(r *http.Request) (*SignedRequest, error) {
	address := headerOrQuery(r, HeaderAddress, QueryAddress)
	timestampStr := headerOrQuery(r, HeaderTimestamp, QueryTimestamp)
	nonce := headerOrQuery(r, HeaderNonce, QueryNonce)
	signatureB64 := headerOrQuery(r, HeaderSignature, QuerySignature)
	pubKeyB64 := headerOrQuery(r, HeaderPubKey, QueryPubKey)

	if address == "" || timestampStr == "" || nonce == "" || signatureB64 == "" || pubKeyB64 == "" {
		return nil, errors.New("missing authentication headers")
	}

	timestampMs, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}
	timestamp := time.UnixMilli(timestampMs)

	now := v.now()
	if timestamp.Before(now.Add(-v.maxAge)) {
		return nil, errors.New("request timestamp too old")
	}
	if timestamp.After(now.Add(v.maxFutureSkew)) {
		return nil, errors.New("request timestamp in future")
	}

	if v.nonceStore.HasSeen(nonce) {
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

	derivedAddr := sdk.AccAddress(pubKey.Address()).String()
	if derivedAddr != address {
		return nil, errors.New("address does not match public key")
	}

	bodyHash, err := computeBodyHash(r)
	if err != nil {
		return nil, err
	}

	requestData := RequestData{
		Method:    strings.ToUpper(r.Method),
		Path:      r.URL.Path,
		Timestamp: timestampMs,
		Nonce:     nonce,
		BodyHash:  bodyHash,
	}

	requestJSON, err := serializeRequestData(requestData)
	if err != nil {
		return nil, err
	}
	dataBase64 := base64.StdEncoding.EncodeToString([]byte(requestJSON))
	if v.chainID == "" {
		return nil, errors.New("chain id not configured")
	}
	signDocBytes, err := buildADR036SignDocBytes(v.chainID, address, dataBase64)
	if err != nil {
		return nil, err
	}

	if !pubKey.VerifySignature(signDocBytes, signature) {
		return nil, errors.New("signature verification failed")
	}

	v.nonceStore.MarkSeen(nonce, timestamp.Add(v.maxAge))

	return &SignedRequest{
		Address:   address,
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: signature,
		PubKey:    pubKey,
		Request:   requestData,
	}, nil
}

func (v *Verifier) VerifyLeaseOwnership(ctx context.Context, address string, leaseID string) error {
	if address == "" {
		return errors.New("missing signer address")
	}
	if v.chainQuery == nil {
		return errors.New("chain query not configured")
	}
	if cachedOwner, ok := v.leaseOwnerCache.Get(leaseID); ok {
		if cachedOwner != address {
			return errors.New("address does not own this lease")
		}
		return nil
	}

	parsed, err := parseLeaseID(leaseID)
	if err != nil {
		return err
	}
	lease, err := v.chainQuery.GetLease(ctx, parsed)
	if err != nil {
		return fmt.Errorf("failed to query lease: %w", err)
	}

	owner := lease.ID.Owner
	if owner == "" {
		return errors.New("lease owner not found")
	}

	v.leaseOwnerCache.Set(leaseID, owner)

	if owner != address {
		return errors.New("address does not own this lease")
	}
	return nil
}

func computeBodyHash(r *http.Request) (string, error) {
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
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:]), nil
}

func serializeRequestData(data RequestData) (string, error) {
	method, err := quoteJSONString(data.Method)
	if err != nil {
		return "", err
	}
	path, err := quoteJSONString(data.Path)
	if err != nil {
		return "", err
	}
	nonce, err := quoteJSONString(data.Nonce)
	if err != nil {
		return "", err
	}
	bodyHash, err := quoteJSONString(data.BodyHash)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("{\"method\":%s,\"path\":%s,\"timestamp\":%d,\"nonce\":%s,\"body_hash\":%s}",
		method,
		path,
		data.Timestamp,
		nonce,
		bodyHash,
	), nil
}

func buildADR036SignDocBytes(chainID, address, dataBase64 string) ([]byte, error) {
	signDoc := map[string]any{
		"account_number": "0",
		"chain_id":       chainID,
		"fee": map[string]any{
			"amount": []any{},
			"gas":    "0",
		},
		"memo": "",
		"msgs": []any{
			map[string]any{
				"type": "sign/MsgSignData",
				"value": map[string]any{
					"data":   dataBase64,
					"signer": address,
				},
			},
		},
		"sequence": "0",
	}

	return canonicalJSON(signDoc)
}

func headerOrQuery(r *http.Request, header, query string) string {
	value := r.Header.Get(header)
	if value == "" {
		value = r.URL.Query().Get(query)
	}
	return value
}
