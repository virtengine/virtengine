package auth

import (
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

const (
	// MaxTimestampAge is the maximum age for signed requests.
	MaxTimestampAge = 5 * time.Minute
	// AllowedFutureSkew allows small client clock drift.
	AllowedFutureSkew = 1 * time.Minute

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

// SignedRequest contains parsed authentication details.
type SignedRequest struct {
	Address   string
	Timestamp time.Time
	Nonce     string
	Signature []byte
	PubKey    *secp256k1.PubKey
	BodyHash  string
}

// RequestData is the canonical format for signed data.
type RequestData struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	BodyHash  string `json:"body_hash"`
}

// AuthMethod represents the auth mechanism used.
type AuthMethod string

const (
	AuthMethodWallet   AuthMethod = "wallet"
	AuthMethodHMAC     AuthMethod = "hmac"
	AuthMethodInsecure AuthMethod = "insecure"
)

// AuthResult represents an authenticated principal.
type AuthResult struct {
	Principal string
	Method    AuthMethod
}
