// Package federation defines fixture-only provider federation contracts.
package federation

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"mime"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"
)

const Version1 uint32 = 1

const (
	serviceRecordDomain = "virtengine/federation/service-record/v1"
	discoveryDomain     = "virtengine/federation/discovery-document/v1"
	requestDomain       = "virtengine/federation/request/v1"
)

type FixtureState string

const (
	StateDisabled    FixtureState = "disabled"
	StateFixtureOnly FixtureState = "fixture_only"
	StateSandbox     FixtureState = "sandbox"
	StateProduction  FixtureState = "production"
)

type ServiceRecord struct {
	Version         uint32
	ProviderID      string
	ServiceID       string
	Revision        uint64
	DiscoveryDigest [32]byte
	ActiveKeyEpoch  uint64
	State           FixtureState
}

func (record ServiceRecord) Validate() error {
	if err := requireVersion(record.Version); err != nil {
		return err
	}
	if record.ProviderID == "" || record.ServiceID == "" || record.Revision == 0 || record.ActiveKeyEpoch == 0 || zeroDigest(record.DiscoveryDigest) {
		return errors.New("complete service record binding is required")
	}
	switch record.State {
	case StateDisabled, StateFixtureOnly:
		return nil
	case StateSandbox, StateProduction:
		return errors.New("federation may not exceed fixture_only")
	default:
		return errors.New("unknown federation state")
	}
}

func (record ServiceRecord) CanonicalBytes() ([]byte, error) {
	if err := record.Validate(); err != nil {
		return nil, err
	}
	encoder := newCanonicalEncoder(serviceRecordDomain)
	encoder.uint32(record.Version)
	encoder.text(record.ProviderID)
	encoder.text(record.ServiceID)
	encoder.uint64(record.Revision)
	encoder.fixed(record.DiscoveryDigest[:])
	encoder.uint64(record.ActiveKeyEpoch)
	encoder.text(string(record.State))
	return encoder.bytes(), nil
}

type Endpoint struct {
	Name string
	URL  string
}

type KeyEpoch struct {
	Epoch     uint64
	PublicKey []byte
	NotBefore time.Time
	NotAfter  time.Time
}

func (epoch KeyEpoch) validate() error {
	if epoch.Epoch == 0 || len(epoch.PublicKey) != ed25519.PublicKeySize || epoch.NotBefore.IsZero() || !epoch.NotBefore.Before(epoch.NotAfter) {
		return errors.New("invalid key epoch")
	}
	if !epoch.NotBefore.Equal(epoch.NotBefore.UTC().Truncate(time.Second)) || !epoch.NotAfter.Equal(epoch.NotAfter.UTC().Truncate(time.Second)) {
		return errors.New("key epoch timestamps must be UTC whole seconds")
	}
	return nil
}

func ValidateKeyEpochTransition(current, next KeyEpoch) error {
	if err := current.validate(); err != nil {
		return fmt.Errorf("current key epoch: %w", err)
	}
	if err := next.validate(); err != nil {
		return fmt.Errorf("next key epoch: %w", err)
	}
	if current.Epoch == ^uint64(0) || next.Epoch != current.Epoch+1 {
		return errors.New("key epoch must advance exactly once")
	}
	if bytes.Equal(current.PublicKey, next.PublicKey) {
		return errors.New("key epoch must rotate public key")
	}
	if next.NotBefore.Before(current.NotBefore) || next.NotBefore.After(current.NotAfter) {
		return errors.New("key epoch transition lacks validity continuity")
	}
	if !next.NotAfter.After(current.NotAfter) {
		return errors.New("key epoch expiry must advance")
	}
	return nil
}

type DiscoveryDocument struct {
	Version         uint32
	ProviderID      string
	ServiceID       string
	Revision        uint64
	Capabilities    []string
	Endpoints       []Endpoint
	KeyEpochs       []KeyEpoch
	IssuedAt        time.Time
	ExpiresAt       time.Time
	PreviousDigest  [32]byte
	SigningKeyEpoch uint64
}

type SignedDiscoveryDocument struct {
	Document  DiscoveryDocument
	Signature []byte
}

func (document DiscoveryDocument) Validate() error {
	if err := requireVersion(document.Version); err != nil {
		return err
	}
	if document.ProviderID == "" || document.ServiceID == "" || document.Revision == 0 || len(document.Capabilities) == 0 || len(document.Endpoints) == 0 || len(document.KeyEpochs) == 0 || document.SigningKeyEpoch == 0 {
		return errors.New("complete discovery binding is required")
	}
	if !document.IssuedAt.Equal(document.IssuedAt.UTC().Truncate(time.Second)) || !document.ExpiresAt.Equal(document.ExpiresAt.UTC().Truncate(time.Second)) || !document.IssuedAt.Before(document.ExpiresAt) {
		return errors.New("invalid discovery validity window")
	}
	if document.Revision == 1 {
		if !zeroDigest(document.PreviousDigest) {
			return errors.New("initial discovery document has previous digest")
		}
	} else if zeroDigest(document.PreviousDigest) {
		return errors.New("revised discovery document requires previous digest")
	}
	capabilities := make(map[string]struct{}, len(document.Capabilities))
	for _, capability := range document.Capabilities {
		if capability == "" {
			return errors.New("empty capability")
		}
		if _, exists := capabilities[capability]; exists {
			return errors.New("duplicate capability")
		}
		capabilities[capability] = struct{}{}
	}
	endpointNames := make(map[string]struct{}, len(document.Endpoints))
	endpointURLs := make(map[string]struct{}, len(document.Endpoints))
	for _, endpoint := range document.Endpoints {
		if endpoint.Name == "" {
			return errors.New("endpoint name is required")
		}
		parsed, err := url.Parse(endpoint.URL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return errors.New("endpoint must be an absolute HTTPS URL without credentials, query, or fragment")
		}
		if _, exists := endpointNames[endpoint.Name]; exists {
			return errors.New("duplicate endpoint")
		}
		if _, exists := endpointURLs[parsed.String()]; exists {
			return errors.New("duplicate endpoint")
		}
		endpointNames[endpoint.Name] = struct{}{}
		endpointURLs[parsed.String()] = struct{}{}
	}
	epochs := make(map[uint64]struct{}, len(document.KeyEpochs))
	for _, epoch := range document.KeyEpochs {
		if err := epoch.validate(); err != nil {
			return err
		}
		if _, exists := epochs[epoch.Epoch]; exists {
			return errors.New("duplicate key epoch")
		}
		epochs[epoch.Epoch] = struct{}{}
	}
	orderedEpochs := slices.Clone(document.KeyEpochs)
	slices.SortFunc(orderedEpochs, func(a, b KeyEpoch) int { return compareUint64(a.Epoch, b.Epoch) })
	for index := 1; index < len(orderedEpochs); index++ {
		if err := ValidateKeyEpochTransition(orderedEpochs[index-1], orderedEpochs[index]); err != nil {
			return err
		}
	}
	if _, exists := epochs[document.SigningKeyEpoch]; !exists {
		return errors.New("signing key epoch is absent")
	}
	return nil
}

func (document DiscoveryDocument) CanonicalBytes() ([]byte, error) {
	if err := document.Validate(); err != nil {
		return nil, err
	}
	capabilities := slices.Clone(document.Capabilities)
	slices.Sort(capabilities)
	endpoints := slices.Clone(document.Endpoints)
	slices.SortFunc(endpoints, func(a, b Endpoint) int {
		if result := strings.Compare(a.Name, b.Name); result != 0 {
			return result
		}
		return strings.Compare(a.URL, b.URL)
	})
	epochs := slices.Clone(document.KeyEpochs)
	slices.SortFunc(epochs, func(a, b KeyEpoch) int { return compareUint64(a.Epoch, b.Epoch) })
	encoder := newCanonicalEncoder(discoveryDomain)
	encoder.uint32(document.Version)
	encoder.text(document.ProviderID)
	encoder.text(document.ServiceID)
	encoder.uint64(document.Revision)
	encoder.uint32(uint32(len(capabilities)))
	for _, capability := range capabilities {
		encoder.text(capability)
	}
	encoder.uint32(uint32(len(endpoints)))
	for _, endpoint := range endpoints {
		encoder.text(endpoint.Name)
		encoder.text(endpoint.URL)
	}
	encoder.uint32(uint32(len(epochs)))
	for _, epoch := range epochs {
		encoder.uint64(epoch.Epoch)
		encoder.data(epoch.PublicKey)
		encoder.int64(epoch.NotBefore.Unix())
		encoder.int64(epoch.NotAfter.Unix())
	}
	encoder.int64(document.IssuedAt.Unix())
	encoder.int64(document.ExpiresAt.Unix())
	encoder.fixed(document.PreviousDigest[:])
	encoder.uint64(document.SigningKeyEpoch)
	return encoder.bytes(), nil
}

func (document DiscoveryDocument) Digest() ([32]byte, error) {
	canonical, err := document.CanonicalBytes()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}

func (signed SignedDiscoveryDocument) Verify(record ServiceRecord, previous *DiscoveryDocument, now time.Time) error {
	if err := record.Validate(); err != nil {
		return err
	}
	canonical, err := signed.Document.CanonicalBytes()
	if err != nil {
		return err
	}
	if signed.Document.ProviderID != record.ProviderID || signed.Document.ServiceID != record.ServiceID || signed.Document.Revision != record.Revision || signed.Document.SigningKeyEpoch != record.ActiveKeyEpoch {
		return errors.New("discovery document does not match service record")
	}
	if now.Before(signed.Document.IssuedAt) || !now.Before(signed.Document.ExpiresAt) {
		return errors.New("discovery document is stale or not yet valid")
	}
	digest := sha256.Sum256(canonical)
	if digest != record.DiscoveryDigest {
		return errors.New("discovery digest mismatch")
	}
	if previous != nil {
		previousDigest, digestErr := previous.Digest()
		if digestErr != nil {
			return digestErr
		}
		if signed.Document.Revision != previous.Revision+1 || signed.Document.PreviousDigest != previousDigest || signed.Document.IssuedAt.Before(previous.IssuedAt) {
			return errors.New("discovery rollback or lineage mismatch")
		}
	}
	epoch, ok := findKeyEpoch(signed.Document.KeyEpochs, signed.Document.SigningKeyEpoch)
	if !ok || now.Before(epoch.NotBefore) || !now.Before(epoch.NotAfter) {
		return errors.New("discovery signing key epoch is inactive")
	}
	if len(signed.Signature) != ed25519.SignatureSize || !ed25519.Verify(ed25519.PublicKey(epoch.PublicKey), canonical, signed.Signature) {
		return errors.New("invalid discovery signature")
	}
	return nil
}

type TrustMode string

const (
	TrustNativeSPKI           TrustMode = "native_spki"
	TrustNativeCA             TrustMode = "native_ca"
	TrustGatewayProxy         TrustMode = "gateway_verification_proxy"
	TrustBrowserSameOrigin    TrustMode = "browser_same_origin"
	TrustBrowserAppContinuity TrustMode = "browser_application_continuity"
)

type ClientTrustProfile struct {
	Version                uint32
	Mode                   TrustMode
	SPKISHA256             [][32]byte
	CABundleSHA256         [32]byte
	GatewayOrigin          string
	GatewayVerificationKey []byte
	ApplicationOrigin      string
	ApplicationContinuity  [32]byte
	ClaimsBrowserPeerSPKI  bool
}

func (profile ClientTrustProfile) Validate() error {
	if err := requireVersion(profile.Version); err != nil {
		return err
	}
	if profile.ClaimsBrowserPeerSPKI && (profile.Mode == TrustBrowserSameOrigin || profile.Mode == TrustBrowserAppContinuity) {
		return errors.New("browser JavaScript cannot directly inspect peer SPKI")
	}
	switch profile.Mode {
	case TrustNativeSPKI:
		if len(profile.SPKISHA256) == 0 || !zeroDigest(profile.CABundleSHA256) || profile.GatewayOrigin != "" || len(profile.GatewayVerificationKey) != 0 || profile.ApplicationOrigin != "" || !zeroDigest(profile.ApplicationContinuity) {
			return errors.New("native SPKI profile requires pins")
		}
		seen := make(map[[32]byte]struct{}, len(profile.SPKISHA256))
		for _, pin := range profile.SPKISHA256 {
			if zeroDigest(pin) {
				return errors.New("native SPKI profile contains empty pin")
			}
			if _, exists := seen[pin]; exists {
				return errors.New("native SPKI profile contains duplicate pin")
			}
			seen[pin] = struct{}{}
		}
	case TrustNativeCA:
		if zeroDigest(profile.CABundleSHA256) || len(profile.SPKISHA256) != 0 || profile.GatewayOrigin != "" || len(profile.GatewayVerificationKey) != 0 || profile.ApplicationOrigin != "" || !zeroDigest(profile.ApplicationContinuity) {
			return errors.New("native CA profile requires bundle digest")
		}
	case TrustGatewayProxy:
		if !validOrigin(profile.GatewayOrigin) || len(profile.GatewayVerificationKey) != ed25519.PublicKeySize || len(profile.SPKISHA256) != 0 || !zeroDigest(profile.CABundleSHA256) || profile.ApplicationOrigin != "" || !zeroDigest(profile.ApplicationContinuity) {
			return errors.New("gateway profile requires HTTPS origin and verification key")
		}
	case TrustBrowserSameOrigin:
		if !validOrigin(profile.ApplicationOrigin) || len(profile.SPKISHA256) != 0 || !zeroDigest(profile.CABundleSHA256) || profile.GatewayOrigin != "" || len(profile.GatewayVerificationKey) != 0 || !zeroDigest(profile.ApplicationContinuity) {
			return errors.New("browser same-origin profile requires HTTPS application origin")
		}
	case TrustBrowserAppContinuity:
		if !validOrigin(profile.ApplicationOrigin) || zeroDigest(profile.ApplicationContinuity) || len(profile.SPKISHA256) != 0 || !zeroDigest(profile.CABundleSHA256) || profile.GatewayOrigin != "" || len(profile.GatewayVerificationKey) != 0 {
			return errors.New("browser continuity profile requires origin and application continuity digest")
		}
	default:
		return errors.New("unsupported trust profile")
	}
	return nil
}

func (profile ClientTrustProfile) ValidateOrigin(requestOrigin, endpointURL string) error {
	if err := profile.Validate(); err != nil {
		return err
	}
	endpoint, err := url.Parse(endpointURL)
	if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" {
		return errors.New("invalid endpoint origin")
	}
	endpointOrigin := endpoint.Scheme + "://" + endpoint.Host
	switch profile.Mode {
	case TrustBrowserSameOrigin:
		if requestOrigin != profile.ApplicationOrigin || endpointOrigin != profile.ApplicationOrigin {
			return errors.New("browser same-origin policy mismatch")
		}
	case TrustBrowserAppContinuity:
		if requestOrigin != profile.ApplicationOrigin {
			return errors.New("browser application origin mismatch")
		}
	case TrustGatewayProxy:
		if requestOrigin != profile.GatewayOrigin {
			return errors.New("gateway origin mismatch")
		}
	case TrustNativeSPKI, TrustNativeCA:
		if requestOrigin != "" {
			return errors.New("native trust profile must not assert browser origin")
		}
	}
	return nil
}

type RequestBinding struct {
	Version         uint32
	ProviderID      string
	ServiceID       string
	Method          string
	Path            string
	Query           string
	BodySHA256      [32]byte
	ContentType     string
	Timestamp       time.Time
	Nonce           string
	SigningKeyEpoch uint64
	DiscoveryDigest [32]byte
	ChainID         string
}

func NewRequestBinding(providerID, serviceID, method, requestTarget, contentType string, body []byte, timestamp time.Time, nonce string, signingKeyEpoch uint64, discoveryDigest [32]byte, chainID string) (RequestBinding, error) {
	normalizedPath, normalizedQuery, err := normalizeTarget(requestTarget)
	if err != nil {
		return RequestBinding{}, err
	}
	normalizedContentType, err := normalizeContentType(contentType)
	if err != nil {
		return RequestBinding{}, err
	}
	return RequestBinding{
		Version: Version1, ProviderID: providerID, ServiceID: serviceID, Method: strings.ToUpper(method), Path: normalizedPath,
		Query: normalizedQuery, BodySHA256: sha256.Sum256(body), ContentType: normalizedContentType, Timestamp: timestamp,
		Nonce: nonce, SigningKeyEpoch: signingKeyEpoch, DiscoveryDigest: discoveryDigest, ChainID: chainID,
	}, nil
}

func (binding RequestBinding) Validate() error {
	if err := requireVersion(binding.Version); err != nil {
		return err
	}
	if binding.ProviderID == "" || binding.ServiceID == "" || binding.Method == "" || binding.Path == "" || binding.Nonce == "" || binding.SigningKeyEpoch == 0 || zeroDigest(binding.DiscoveryDigest) || binding.ChainID == "" || zeroDigest(binding.BodySHA256) {
		return errors.New("complete request binding is required")
	}
	if binding.Method != strings.ToUpper(binding.Method) || !binding.Timestamp.Equal(binding.Timestamp.UTC().Truncate(time.Second)) {
		return errors.New("request method or timestamp is not canonical")
	}
	normalizedPath, normalizedQuery, err := normalizeTarget(binding.Path + querySuffix(binding.Query))
	if err != nil || normalizedPath != binding.Path || normalizedQuery != binding.Query {
		return errors.New("request target is not canonical")
	}
	normalizedContentType, err := normalizeContentType(binding.ContentType)
	if err != nil || normalizedContentType != binding.ContentType {
		return errors.New("request content type is not canonical")
	}
	return nil
}

func (binding RequestBinding) CanonicalBytes() ([]byte, error) {
	if err := binding.Validate(); err != nil {
		return nil, err
	}
	encoder := newCanonicalEncoder(requestDomain)
	encoder.uint32(binding.Version)
	encoder.text(binding.ProviderID)
	encoder.text(binding.ServiceID)
	encoder.text(binding.Method)
	encoder.text(binding.Path)
	encoder.text(binding.Query)
	encoder.fixed(binding.BodySHA256[:])
	encoder.text(binding.ContentType)
	encoder.int64(binding.Timestamp.Unix())
	encoder.text(binding.Nonce)
	encoder.uint64(binding.SigningKeyEpoch)
	encoder.fixed(binding.DiscoveryDigest[:])
	encoder.text(binding.ChainID)
	return encoder.bytes(), nil
}

func (binding RequestBinding) Digest() ([32]byte, error) {
	canonical, err := binding.CanonicalBytes()
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}

type AuthRequirement string

const (
	AuthEd25519 AuthRequirement = "ed25519"
	AuthNone    AuthRequirement = "none"
)

type RoutePolicy struct {
	Version            uint32
	ChainID            string
	ServiceID          string
	Method             string
	ExactPath          string
	Auth               AuthRequirement
	RequiredCapability string
	MaxBodyBytes       uint64
	MaxClockSkew       time.Duration
	MaxRequestAge      time.Duration
	ContentTypes       []string
}

func (policy RoutePolicy) Validate() error {
	if err := requireVersion(policy.Version); err != nil {
		return err
	}
	if policy.ChainID == "" || policy.ServiceID == "" || policy.Method == "" || policy.Method != strings.ToUpper(policy.Method) || policy.ExactPath == "" || policy.RequiredCapability == "" || policy.MaxBodyBytes == 0 || policy.MaxClockSkew < 0 || policy.MaxRequestAge <= 0 || len(policy.ContentTypes) == 0 {
		return errors.New("complete route policy is required")
	}
	if policy.Auth != AuthEd25519 {
		return errors.New("unsupported or unauthenticated route policy")
	}
	normalizedPath, query, err := normalizeTarget(policy.ExactPath)
	if err != nil || query != "" || normalizedPath != policy.ExactPath {
		return errors.New("route policy path is not exact canonical path")
	}
	seen := make(map[string]struct{}, len(policy.ContentTypes))
	for _, contentType := range policy.ContentTypes {
		normalized, normalizeErr := normalizeContentType(contentType)
		if normalizeErr != nil || normalized != contentType {
			return errors.New("route content type is not canonical")
		}
		if _, exists := seen[contentType]; exists {
			return errors.New("duplicate route content type")
		}
		seen[contentType] = struct{}{}
	}
	return nil
}

func (policy RoutePolicy) Authorize(binding RequestBinding, document DiscoveryDocument, bodyBytes uint64, now time.Time) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if err := binding.Validate(); err != nil {
		return err
	}
	if err := document.Validate(); err != nil {
		return err
	}
	if binding.ChainID != policy.ChainID || binding.ProviderID != document.ProviderID || binding.ServiceID != document.ServiceID || binding.ServiceID != policy.ServiceID {
		return errors.New("request provider or service mismatch")
	}
	if binding.Method != policy.Method || binding.Path != policy.ExactPath {
		return errors.New("request route or method mismatch")
	}
	if bodyBytes > policy.MaxBodyBytes || !slices.Contains(policy.ContentTypes, binding.ContentType) {
		return errors.New("request body or content type denied")
	}
	if !slices.Contains(document.Capabilities, policy.RequiredCapability) {
		return errors.New("required capability absent")
	}
	if now.Add(policy.MaxClockSkew).Before(binding.Timestamp) || !binding.Timestamp.Add(policy.MaxRequestAge).After(now) {
		return errors.New("request timestamp outside policy window")
	}
	discoveryDigest, err := document.Digest()
	if err != nil || binding.DiscoveryDigest != discoveryDigest {
		return errors.New("request discovery digest mismatch")
	}
	epoch, ok := findKeyEpoch(document.KeyEpochs, binding.SigningKeyEpoch)
	if !ok || binding.Timestamp.Before(epoch.NotBefore) || !binding.Timestamp.Before(epoch.NotAfter) {
		return errors.New("request key epoch inactive")
	}
	return nil
}

type AtomicNonceStore interface {
	WithNonce(ctx context.Context, scope, nonce string, protected func() error) error
}

func VerifyAndConsume(ctx context.Context, store AtomicNonceStore, policy RoutePolicy, binding RequestBinding, document DiscoveryDocument, body []byte, now time.Time, signature []byte, protected func() error) error {
	if store == nil || protected == nil {
		return errors.New("atomic nonce store and protected operation are required")
	}
	if uint64(len(body)) > policy.MaxBodyBytes || sha256.Sum256(body) != binding.BodySHA256 {
		return errors.New("request body mismatch")
	}
	if err := policy.Authorize(binding, document, uint64(len(body)), now); err != nil {
		return err
	}
	canonical, err := binding.CanonicalBytes()
	if err != nil {
		return err
	}
	epoch, ok := findKeyEpoch(document.KeyEpochs, binding.SigningKeyEpoch)
	if !ok || len(signature) != ed25519.SignatureSize || !ed25519.Verify(ed25519.PublicKey(epoch.PublicKey), canonical, signature) {
		return errors.New("invalid request signature")
	}
	scope := binding.ChainID + "\x00" + binding.ProviderID + "\x00" + binding.ServiceID + "\x00" + fmt.Sprintf("%d", binding.SigningKeyEpoch)
	return store.WithNonce(ctx, scope, binding.Nonce, protected)
}

type canonicalEncoder struct{ buffer bytes.Buffer }

func newCanonicalEncoder(domain string) *canonicalEncoder {
	encoder := &canonicalEncoder{}
	encoder.text(domain)
	return encoder
}

func (encoder *canonicalEncoder) uint32(value uint32) {
	_ = binary.Write(&encoder.buffer, binary.BigEndian, value)
}
func (encoder *canonicalEncoder) uint64(value uint64) {
	_ = binary.Write(&encoder.buffer, binary.BigEndian, value)
}
func (encoder *canonicalEncoder) int64(value int64) {
	_ = binary.Write(&encoder.buffer, binary.BigEndian, value)
}
func (encoder *canonicalEncoder) fixed(value []byte) { _, _ = encoder.buffer.Write(value) }
func (encoder *canonicalEncoder) data(value []byte) {
	encoder.uint32(uint32(len(value)))
	encoder.fixed(value)
}
func (encoder *canonicalEncoder) text(value string) { encoder.data([]byte(value)) }
func (encoder *canonicalEncoder) bytes() []byte     { return slices.Clone(encoder.buffer.Bytes()) }

func requireVersion(version uint32) error {
	if version != Version1 {
		return fmt.Errorf("unsupported version %d", version)
	}
	return nil
}

func zeroDigest(digest [32]byte) bool { return digest == [32]byte{} }

func findKeyEpoch(epochs []KeyEpoch, number uint64) (KeyEpoch, bool) {
	for _, epoch := range epochs {
		if epoch.Epoch == number {
			return epoch, true
		}
	}
	return KeyEpoch{}, false
}

func compareUint64(a, b uint64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func validOrigin(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil && parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == "" && raw == parsed.Scheme+"://"+parsed.Host
}

func normalizeTarget(target string) (string, string, error) {
	parsed, err := url.ParseRequestURI(target)
	if err != nil || !strings.HasPrefix(parsed.Path, "/") || parsed.Fragment != "" {
		return "", "", errors.New("invalid request target")
	}
	normalizedPath := path.Clean(parsed.Path)
	if strings.HasSuffix(parsed.Path, "/") && normalizedPath != "/" {
		normalizedPath += "/"
	}
	if normalizedPath != parsed.Path || strings.Contains(parsed.RawPath, "%2f") || strings.Contains(parsed.RawPath, "%2F") {
		return "", "", errors.New("ambiguous request path")
	}
	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return "", "", errors.New("invalid request query")
	}
	return normalizedPath, query.Encode(), nil
}

func normalizeContentType(value string) (string, error) {
	mediaType, parameters, err := mime.ParseMediaType(value)
	if err != nil || mediaType == "" {
		return "", errors.New("invalid content type")
	}
	return mime.FormatMediaType(strings.ToLower(mediaType), parameters), nil
}

func querySuffix(query string) string {
	if query == "" {
		return ""
	}
	return "?" + query
}
