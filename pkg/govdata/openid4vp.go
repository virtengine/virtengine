package govdata

import (
	"context"
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrOpenID4VPVerification = errors.New("OpenID4VP verification failed")

// IssuerKeyResolver resolves a preconfigured issuer key by KID. Resolvers must
// obtain keys from trusted issuer metadata/JWKS; this package intentionally
// never follows jku/x5u URLs supplied by an untrusted presentation.
type IssuerKeyResolver interface {
	ResolveIssuerKey(ctx context.Context, issuer string, keyID string) (crypto.PublicKey, error)
}

// CredentialStatusResolver validates a credential's status reference with the
// appropriate trusted issuer service (including revocation/status-list logic).
type CredentialStatusResolver interface {
	ResolveCredentialStatus(ctx context.Context, issuer string, subject string, status json.RawMessage) (CredentialStatus, error)
}

// AssuranceMapper explicitly maps issuer-specific assurance values to VEID's
// neutral assurance levels. Unknown values must return ok=false.
type AssuranceMapper interface {
	MapAssurance(string) (AssuranceLevel, bool)
}

// AuthorizationReplayGuard atomically consumes a successfully verified
// authorization state. Implementations must retain the state through expiry
// and reject a second consume attempt.
type AuthorizationReplayGuard interface {
	ConsumeAuthorizationState(context.Context, string, time.Time) error
}

// OpenID4VPVerifierConfig pins the protocol issuer and policy. A production
// integration creates one verifier config per trusted issuer/profile.
type OpenID4VPVerifierConfig struct {
	ProviderID  string
	Issuer      string
	Leeway      time.Duration
	ReplayGuard AuthorizationReplayGuard
}

func (c OpenID4VPVerifierConfig) validate() error {
	if strings.TrimSpace(c.ProviderID) == "" || strings.TrimSpace(c.Issuer) == "" {
		return fmt.Errorf("%w: provider ID and issuer are required", ErrOpenID4VPVerification)
	}
	if c.ReplayGuard == nil {
		return fmt.Errorf("%w: authorization replay guard is required", ErrOpenID4VPVerification)
	}
	if c.Leeway < 0 || c.Leeway > 2*time.Minute {
		return fmt.Errorf("%w: leeway must be between zero and two minutes", ErrOpenID4VPVerification)
	}
	return nil
}

// OpenID4VPAuthorizationResponse is the callback payload for a presentation.
// VPToken must be a compact JWS; raw JSON claims are intentionally unsupported.
type OpenID4VPAuthorizationResponse struct {
	State       string
	VPToken     string
	Disclosures []string
	Error       string
}

func (r OpenID4VPAuthorizationResponse) validate(request DigitalIDAuthorizationRequest) error {
	if r.Error != "" {
		return fmt.Errorf("%w: wallet returned %q", ErrOpenID4VPVerification, r.Error)
	}
	if r.State == "" || r.State != request.State {
		return fmt.Errorf("%w: callback state mismatch", ErrOpenID4VPVerification)
	}
	if strings.Count(r.VPToken, ".") != 2 || len(r.VPToken) > 32*1024 {
		return fmt.Errorf("%w: vp_token must be a bounded compact JWS", ErrOpenID4VPVerification)
	}
	return nil
}

type openID4VPClaims struct {
	Assurance        string          `json:"assurance"`
	CredentialStatus json.RawMessage `json:"credential_status"`
	SD               []string        `json:"_sd"`
	SDAlgorithm      string          `json:"_sd_alg"`
	Nonce            string          `json:"nonce"`
	jwt.RegisteredClaims
}

// VerifyOpenID4VP verifies a signed OpenID4VP response and returns only claims
// authenticated by its issuer-signed JWT (and, when present, valid SD-JWT
// disclosures). It fails closed on every missing callback, signature, time,
// audience, nonce, assurance, or credential-status binding.
func VerifyOpenID4VP(
	ctx context.Context,
	request DigitalIDAuthorizationRequest,
	response OpenID4VPAuthorizationResponse,
	config OpenID4VPVerifierConfig,
	keys IssuerKeyResolver,
	status CredentialStatusResolver,
	assurance AssuranceMapper,
	now time.Time,
) (DigitalIDIdentity, error) {
	if err := request.Validate(now); err != nil {
		return DigitalIDIdentity{}, fmt.Errorf("%w: invalid authorization request: %v", ErrOpenID4VPVerification, err)
	}
	if err := response.validate(request); err != nil {
		return DigitalIDIdentity{}, err
	}
	if err := config.validate(); err != nil || keys == nil || status == nil || assurance == nil {
		if err != nil {
			return DigitalIDIdentity{}, err
		}
		return DigitalIDIdentity{}, fmt.Errorf("%w: key, status, and assurance resolvers are required", ErrOpenID4VPVerification)
	}

	claims := openID4VPClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg(), jwt.SigningMethodES256.Alg(), jwt.SigningMethodEdDSA.Alg()}), jwt.WithIssuer(config.Issuer), jwt.WithAudience(request.ClientID), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(config.Leeway), jwt.WithTimeFunc(func() time.Time { return now.UTC() }))
	_, err := parser.ParseWithClaims(response.VPToken, &claims, func(token *jwt.Token) (interface{}, error) {
		kid, _ := token.Header["kid"].(string)
		if strings.TrimSpace(kid) == "" {
			return nil, fmt.Errorf("missing signing key ID")
		}
		return keys.ResolveIssuerKey(ctx, config.Issuer, kid)
	})
	if err != nil {
		return DigitalIDIdentity{}, fmt.Errorf("%w: invalid signed vp_token: %v", ErrOpenID4VPVerification, err)
	}
	if claims.Subject == "" || claims.Nonce == "" || claims.Nonce != request.Nonce {
		return DigitalIDIdentity{}, fmt.Errorf("%w: subject or nonce binding mismatch", ErrOpenID4VPVerification)
	}
	if len(claims.CredentialStatus) == 0 || string(claims.CredentialStatus) == "null" {
		return DigitalIDIdentity{}, fmt.Errorf("%w: credential status reference is required", ErrOpenID4VPVerification)
	}
	level, ok := assurance.MapAssurance(claims.Assurance)
	if !ok || !level.Valid() {
		return DigitalIDIdentity{}, fmt.Errorf("%w: issuer assurance is not mapped", ErrOpenID4VPVerification)
	}
	credentialStatus, err := status.ResolveCredentialStatus(ctx, config.Issuer, claims.Subject, claims.CredentialStatus)
	if err != nil || credentialStatus != CredentialStatusActive {
		return DigitalIDIdentity{}, fmt.Errorf("%w: credential status is not active", ErrOpenID4VPVerification)
	}
	verifiedClaims, err := verifiedSDJWTClaims(response.Disclosures, claims.SD, claims.SDAlgorithm, request.RequestedClaims)
	if err != nil {
		return DigitalIDIdentity{}, err
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return DigitalIDIdentity{}, fmt.Errorf("%w: token issue and expiry times are required", ErrOpenID4VPVerification)
	}
	identity := DigitalIDIdentity{ProviderID: config.ProviderID, Subject: claims.Subject, Assurance: level, Claims: verifiedClaims, IssuedAt: claims.IssuedAt.Time.UTC(), ExpiresAt: claims.ExpiresAt.Time.UTC(), Status: credentialStatus}
	if err := ValidateIdentity(request, identity, now); err != nil {
		return DigitalIDIdentity{}, fmt.Errorf("%w: %v", ErrOpenID4VPVerification, err)
	}
	if err := config.ReplayGuard.ConsumeAuthorizationState(ctx, request.State, request.ExpiresAt.UTC()); err != nil {
		return DigitalIDIdentity{}, fmt.Errorf("%w: authorization state already consumed or unavailable: %v", ErrOpenID4VPVerification, err)
	}
	return identity, nil
}

func verifiedSDJWTClaims(disclosures, digests []string, algorithm string, requested []string) (map[string]string, error) {
	if algorithm != "" && algorithm != "sha-256" {
		return nil, fmt.Errorf("%w: unsupported SD-JWT digest algorithm", ErrOpenID4VPVerification)
	}
	if len(digests) == 0 {
		return nil, fmt.Errorf("%w: signed SD-JWT disclosures are required", ErrOpenID4VPVerification)
	}
	allowed := make(map[string]struct{}, len(digests))
	for _, digest := range digests {
		decoded, err := base64.RawURLEncoding.DecodeString(digest)
		if err != nil || len(decoded) != sha256.Size {
			return nil, fmt.Errorf("%w: malformed SD-JWT digest", ErrOpenID4VPVerification)
		}
		allowed[digest] = struct{}{}
	}
	requestedSet := make(map[string]struct{}, len(requested))
	for _, name := range requested {
		requestedSet[name] = struct{}{}
	}
	claims := make(map[string]string, len(disclosures))
	for _, disclosure := range disclosures {
		digest := sha256.Sum256([]byte(disclosure))
		encodedDigest := base64.RawURLEncoding.EncodeToString(digest[:])
		if _, ok := allowed[encodedDigest]; !ok {
			return nil, fmt.Errorf("%w: disclosure is not committed by signed SD-JWT", ErrOpenID4VPVerification)
		}
		raw, err := base64.RawURLEncoding.DecodeString(disclosure)
		if err != nil {
			return nil, fmt.Errorf("%w: malformed disclosure", ErrOpenID4VPVerification)
		}
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil || len(values) != 3 {
			return nil, fmt.Errorf("%w: disclosure must contain salt, name, and value", ErrOpenID4VPVerification)
		}
		var salt, name, value string
		if json.Unmarshal(values[0], &salt) != nil || json.Unmarshal(values[1], &name) != nil || json.Unmarshal(values[2], &value) != nil || salt == "" || name == "" || value == "" {
			return nil, fmt.Errorf("%w: disclosure fields are invalid", ErrOpenID4VPVerification)
		}
		if _, exists := claims[name]; exists {
			return nil, fmt.Errorf("%w: duplicate disclosure claim", ErrOpenID4VPVerification)
		}
		if _, requested := requestedSet[name]; !requested {
			return nil, fmt.Errorf("%w: disclosure contains an unrequested claim", ErrOpenID4VPVerification)
		}
		claims[name] = value
	}
	for _, name := range requested {
		if strings.TrimSpace(claims[name]) == "" {
			return nil, fmt.Errorf("%w: required disclosed claim %q is missing", ErrOpenID4VPVerification, name)
		}
	}
	return claims, nil
}
