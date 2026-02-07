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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/bech32"
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

const (
	DefaultMaxTimestampAge   = 5 * time.Minute
	DefaultFutureTimeDrift   = time.Minute
	DefaultOwnershipCacheTTL = 15 * time.Minute
)

// SignedRequest contains parsed authentication data.
type SignedRequest struct {
	Address   string
	Timestamp time.Time
	Nonce     string
	Signature []byte
	PubKey    *secp256k1.PubKey
	BodyHash  string
}

// RequestData is the canonical format of signed data.
type RequestData struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	BodyHash  string `json:"body_hash"`
}

// VerifierConfig configures wallet signature verification.
type VerifierConfig struct {
	ChainID           string
	NonceStore        NonceStore
	ChainQuerier      ChainQuerier
	MaxTimestampAge   time.Duration
	FutureTimeDrift   time.Duration
	OwnershipCacheTTL time.Duration
	Clock             func() time.Time
}

// Verifier verifies wallet-signed requests and lease ownership.
type Verifier struct {
	chainID           string
	nonceStore        NonceStore
	chainQuery        ChainQuerier
	maxTimestampAge   time.Duration
	futureTimeDrift   time.Duration
	ownershipCacheTTL time.Duration
	clock             func() time.Time

	cacheMu sync.RWMutex
	cache   map[string]ownershipCacheEntry
}

type ownershipCacheEntry struct {
	owner     string
	expiresAt time.Time
}

// NewVerifier constructs a verifier with defaults.
func NewVerifier(cfg VerifierConfig) *Verifier {
	nonceStore := cfg.NonceStore
	if nonceStore == nil {
		nonceStore = NewInMemoryNonceStore()
	}
	chainQuery := cfg.ChainQuerier
	if chainQuery == nil {
		chainQuery = NoopChainQuerier{}
	}
	maxAge := cfg.MaxTimestampAge
	if maxAge == 0 {
		maxAge = DefaultMaxTimestampAge
	}
	futureDrift := cfg.FutureTimeDrift
	if futureDrift == 0 {
		futureDrift = DefaultFutureTimeDrift
	}
	ownershipTTL := cfg.OwnershipCacheTTL
	if ownershipTTL == 0 {
		ownershipTTL = DefaultOwnershipCacheTTL
	}
	clock := cfg.Clock
	if clock == nil {
		clock = time.Now
	}

	return &Verifier{
		chainID:           cfg.ChainID,
		nonceStore:        nonceStore,
		chainQuery:        chainQuery,
		maxTimestampAge:   maxAge,
		futureTimeDrift:   futureDrift,
		ownershipCacheTTL: ownershipTTL,
		clock:             clock,
		cache:             make(map[string]ownershipCacheEntry),
	}
}

// HasWalletAuth checks if a request contains wallet auth headers or query params.
func HasWalletAuth(r *http.Request) bool {
	if r == nil {
		return false
	}
	return readAuthValue(r, HeaderAddress, QueryAddress) != ""
}

// Verify verifies the signature and returns the signer address.
func (v *Verifier) Verify(r *http.Request) (*SignedRequest, error) {
	if v == nil {
		return nil, errors.New("wallet verifier not configured")
	}
	if v.chainID == "" {
		return nil, errors.New("wallet auth chain id is required")
	}

	address := readAuthValue(r, HeaderAddress, QueryAddress)
	timestampStr := readAuthValue(r, HeaderTimestamp, QueryTimestamp)
	nonce := readAuthValue(r, HeaderNonce, QueryNonce)
	signatureB64 := readAuthValue(r, HeaderSignature, QuerySignature)
	pubKeyB64 := readAuthValue(r, HeaderPubKey, QueryPubKey)

	if address == "" || timestampStr == "" || nonce == "" || signatureB64 == "" || pubKeyB64 == "" {
		return nil, errors.New("missing wallet authentication headers")
	}

	timestampMs, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid timestamp")
	}
	timestamp := time.UnixMilli(timestampMs).UTC()

	now := v.clock()
	if now.Sub(timestamp) > v.maxTimestampAge {
		return nil, errors.New("request timestamp too old")
	}
	if timestamp.After(now.Add(v.futureTimeDrift)) {
		return nil, errors.New("request timestamp in future")
	}

	nonceKey := fmt.Sprintf("%s:%s", address, nonce)
	if v.nonceStore.HasSeen(nonceKey) {
		return nil, errors.New("nonce already used")
	}

	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, errors.New("invalid signature encoding")
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyB64)
	if err != nil {
		return nil, errors.New("invalid public key encoding")
	}
	pubKey := &secp256k1.PubKey{Key: pubKeyBytes}

	_, addrBytes, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return nil, errors.New("invalid address")
	}
	if !bytes.Equal(addrBytes, pubKey.Address()) {
		return nil, errors.New("address does not match public key")
	}

	bodyHash, err := hashRequestBody(r)
	if err != nil {
		return nil, fmt.Errorf("failed to hash body: %w", err)
	}

	requestData := RequestData{
		Method:    strings.ToUpper(r.Method),
		Path:      r.URL.Path,
		Timestamp: timestampMs,
		Nonce:     nonce,
		BodyHash:  bodyHash,
	}

	dataToSign := serializeRequestData(requestData)
	signDoc := buildSignDoc(v.chainID, address, base64.StdEncoding.EncodeToString([]byte(dataToSign)))
	signBytes, err := canonicalJSON(signDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize sign doc: %w", err)
	}

	if !pubKey.VerifySignature([]byte(signBytes), signature) {
		return nil, errors.New("signature verification failed")
	}

	v.nonceStore.MarkSeen(nonceKey, timestamp.Add(v.maxTimestampAge))

	return &SignedRequest{
		Address:   address,
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: signature,
		PubKey:    pubKey,
		BodyHash:  bodyHash,
	}, nil
}

// VerifyLeaseOwnership checks if the address owns the lease.
func (v *Verifier) VerifyLeaseOwnership(ctx context.Context, address, leaseID string) error {
	if v == nil {
		return errors.New("wallet verifier not configured")
	}
	if leaseID == "" {
		return errors.New("lease id required")
	}
	if address == "" {
		return errors.New("signer address required")
	}

	if owner, ok := v.cachedOwner(leaseID); ok {
		if owner != address {
			return errors.New("address does not own this lease")
		}
		return nil
	}

	lease, err := v.chainQuery.GetLease(ctx, leaseID)
	if err != nil {
		return err
	}
	if lease == nil || lease.Owner == "" {
		return ErrLeaseNotFound
	}
	v.storeOwner(leaseID, lease.Owner)
	if lease.Owner != address {
		return errors.New("address does not own this lease")
	}
	return nil
}

func (v *Verifier) cachedOwner(leaseID string) (string, bool) {
	v.cacheMu.RLock()
	entry, ok := v.cache[leaseID]
	v.cacheMu.RUnlock()
	if !ok {
		return "", false
	}
	if v.clock().After(entry.expiresAt) {
		return "", false
	}
	return entry.owner, true
}

func (v *Verifier) storeOwner(leaseID, owner string) {
	v.cacheMu.Lock()
	v.cache[leaseID] = ownershipCacheEntry{
		owner:     owner,
		expiresAt: v.clock().Add(v.ownershipCacheTTL),
	}
	v.cacheMu.Unlock()
}

func readAuthValue(r *http.Request, headerKey, queryKey string) string {
	if r == nil {
		return ""
	}
	if headerKey != "" {
		if val := strings.TrimSpace(r.Header.Get(headerKey)); val != "" {
			return val
		}
	}
	if queryKey != "" {
		if val := strings.TrimSpace(r.URL.Query().Get(queryKey)); val != "" {
			return val
		}
	}
	return ""
}

func serializeRequestData(data RequestData) string {
	return fmt.Sprintf(
		"{\"method\":\"%s\",\"path\":\"%s\",\"timestamp\":%d,\"nonce\":\"%s\",\"body_hash\":\"%s\"}",
		data.Method,
		data.Path,
		data.Timestamp,
		data.Nonce,
		data.BodyHash,
	)
}

func buildSignDoc(chainID, address, dataBase64 string) map[string]any {
	return map[string]any{
		"chain_id":       chainID,
		"account_number": "0",
		"sequence":       "0",
		"fee": map[string]any{
			"gas":    "0",
			"amount": []any{},
		},
		"msgs": []any{
			map[string]any{
				"type": "sign/MsgSignData",
				"value": map[string]any{
					"signer": address,
					"data":   dataBase64,
				},
			},
		},
		"memo": "",
	}
}

func hashRequestBody(r *http.Request) (string, error) {
	if r == nil || r.Body == nil {
		return "", nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))

	if len(body) == 0 {
		return "", nil
	}

	canonical, ok := canonicalizeJSON(body)
	if ok {
		sum := sha256.Sum256([]byte(canonical))
		return hex.EncodeToString(sum[:]), nil
	}

	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func canonicalizeJSON(body []byte) (string, bool) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return "", false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return "", false
	}
	canonical, err := canonicalJSON(payload)
	if err != nil {
		return "", false
	}
	return canonical, true
}

func canonicalJSON(value any) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "null", nil
	case string:
		return marshalPrimitive(typed)
	case bool:
		return marshalPrimitive(typed)
	case json.Number:
		return typed.String(), nil
	case float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		return marshalPrimitive(typed)
	case []any:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, item := range typed {
			if i > 0 {
				buf.WriteByte(',')
			}
			encoded, err := canonicalJSON(item)
			if err != nil {
				return "", err
			}
			buf.WriteString(encoded)
		}
		buf.WriteByte(']')
		return buf.String(), nil
	case map[string]any:
		return canonicalMap(typed)
	default:
		return marshalPrimitive(typed)
	}
}

func canonicalMap(value map[string]any) (string, error) {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		encodedKey, err := marshalPrimitive(key)
		if err != nil {
			return "", err
		}
		buf.WriteString(encodedKey)
		buf.WriteByte(':')
		encodedValue, err := canonicalJSON(value[key])
		if err != nil {
			return "", err
		}
		buf.WriteString(encodedValue)
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

func marshalPrimitive(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
