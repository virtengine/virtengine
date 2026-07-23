package offramp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	WebhookAlgorithmHMACSHA256 = "HMAC-SHA256"
	defaultWebhookBodyLimit    = int64(1 << 20)
	defaultWebhookClockSkew    = 5 * time.Minute
)

// WebhookSecretResolver resolves one pinned signing key ID and version.
type WebhookSecretResolver interface {
	ResolveWebhookSecret(ctx context.Context, secretRef string) ([]byte, error)
}

// WebhookReplayRepository atomically records event identity and payload digest.
// Implementations must be durable in production.
type WebhookReplayRepository interface {
	RecordWebhookEvent(ctx context.Context, provider string, eventID string, payloadDigest string, receivedAt time.Time) (WebhookReplayResult, error)
}

// DurableWebhookReplayRepository identifies replay stores safe for production.
type DurableWebhookReplayRepository interface {
	WebhookReplayRepository
	Durable() bool
}

// WebhookReplayResult describes the atomic replay decision.
type WebhookReplayResult string

const (
	WebhookReplayNew         WebhookReplayResult = "new"
	WebhookReplayExact       WebhookReplayResult = "duplicate_exact"
	WebhookReplayConflicting WebhookReplayResult = "duplicate_conflicting"
)

// WebhookBinding binds provider callbacks to a locally initiated payout.
type WebhookBinding struct {
	Provider       string
	PayoutID       string
	QuoteID        string
	CorrelationID  string
	ReservationDay string
}

// WebhookBindingRepository resolves immutable payout bindings.
type WebhookBindingRepository interface {
	LookupWebhookBinding(ctx context.Context, provider string, payoutID string) (WebhookBinding, error)
}

// DurableWebhookBindingRepository identifies restart-safe payout bindings.
type DurableWebhookBindingRepository interface {
	WebhookBindingRepository
	Durable() bool
}

// WebhookHeaders contains parsed transport headers.
type WebhookHeaders struct {
	Timestamp  string
	Signature  string
	KeyID      string
	KeyVersion string
	APIVersion string
}

// WebhookEvent is the verified, replay-safe callback payload.
type WebhookEvent struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	Provider      string    `json:"provider"`
	APIVersion    string    `json:"api_version"`
	PayoutID      string    `json:"payout_id"`
	QuoteID       string    `json:"quote_id"`
	CorrelationID string    `json:"correlation_id"`
	Status        Status    `json:"status"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// VerifiedWebhook reports whether an exact duplicate was safely ignored.
type VerifiedWebhook struct {
	Event      WebhookEvent
	Duplicate  bool
	KeyID      string
	KeyVersion string
}

// WebhookVerifierConfig configures one profile-pinned verifier.
type WebhookVerifierConfig struct {
	Profile      PayoutProfile
	Secrets      WebhookSecretResolver
	Replay       WebhookReplayRepository
	Bindings     WebhookBindingRepository
	Authorizer   ProfileAuthorizer
	Clock        func() time.Time
	MaxClockSkew time.Duration
	MaxBodyBytes int64
}

// WebhookVerifier authenticates partner callbacks before state reconciliation.
type WebhookVerifier struct {
	profile      PayoutProfile
	secrets      WebhookSecretResolver
	replay       WebhookReplayRepository
	bindings     WebhookBindingRepository
	authorizer   ProfileAuthorizer
	now          func() time.Time
	maxClockSkew time.Duration
	maxBodyBytes int64
}

// NewWebhookVerifier validates a webhook profile and its durable dependencies.
func NewWebhookVerifier(cfg WebhookVerifierConfig) (*WebhookVerifier, error) {
	if err := cfg.Profile.Validate(); err != nil {
		return nil, err
	}
	if cfg.Secrets == nil || cfg.Replay == nil || cfg.Bindings == nil {
		return nil, fmt.Errorf("%w: webhook secret, replay, and binding repositories are required", ErrInvalidRequest)
	}
	if cfg.Profile.Environment == EnvironmentProduction {
		if err := cfg.Profile.ValidateForExecution(ExecutionModeProduction, false); err != nil {
			return nil, err
		}
		if cfg.Authorizer == nil || cfg.Authorizer.AuthorizePayoutProfile(cfg.Profile) != nil {
			return nil, fmt.Errorf("%w: trusted webhook profile authorization is required", ErrProfileNotExecutable)
		}
		durable, ok := cfg.Replay.(DurableWebhookReplayRepository)
		if !ok || !durable.Durable() {
			return nil, fmt.Errorf("%w: durable webhook replay repository is required", ErrProfileNotExecutable)
		}
		durableBindings, ok := cfg.Bindings.(DurableWebhookBindingRepository)
		if !ok || !durableBindings.Durable() {
			return nil, fmt.Errorf("%w: durable webhook binding repository is required", ErrProfileNotExecutable)
		}
	}
	now := cfg.Clock
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	skew := cfg.MaxClockSkew
	if skew <= 0 {
		skew = defaultWebhookClockSkew
	}
	if skew > 15*time.Minute {
		return nil, fmt.Errorf("%w: webhook clock skew exceeds safety bound", ErrInvalidRequest)
	}
	bodyLimit := cfg.MaxBodyBytes
	if bodyLimit <= 0 {
		bodyLimit = defaultWebhookBodyLimit
	}
	if bodyLimit > 8<<20 {
		return nil, fmt.Errorf("%w: webhook body limit exceeds safety bound", ErrInvalidRequest)
	}
	return &WebhookVerifier{
		profile: clonePayoutProfile(cfg.Profile), secrets: cfg.Secrets, replay: cfg.Replay, bindings: cfg.Bindings,
		authorizer: cfg.Authorizer, now: now, maxClockSkew: skew, maxBodyBytes: bodyLimit,
	}, nil
}

// Verify authenticates timestamp plus raw body canonical bytes, then atomically
// claims the event ID. Signature verification happens before JSON parsing.
func (v *WebhookVerifier) Verify(ctx context.Context, headers WebhookHeaders, body io.Reader) (VerifiedWebhook, error) {
	if v.profile.Environment == EnvironmentProduction && (v.authorizer == nil || v.authorizer.AuthorizePayoutProfile(v.profile) != nil) {
		return VerifiedWebhook{}, ErrProfileNotExecutable
	}
	if body == nil || headers.APIVersion != v.profile.Webhook.Version || headers.KeyID == "" || headers.KeyVersion == "" {
		return VerifiedWebhook{}, ErrWebhookInvalid
	}
	timestampSeconds, err := strconv.ParseInt(headers.Timestamp, 10, 64)
	if err != nil || strconv.FormatInt(timestampSeconds, 10) != headers.Timestamp {
		return VerifiedWebhook{}, ErrWebhookInvalid
	}
	signedAt := time.Unix(timestampSeconds, 0).UTC()
	now := v.now().UTC()
	delta := now.Sub(signedAt)
	if delta < -v.maxClockSkew || delta > v.maxClockSkew {
		return VerifiedWebhook{}, ErrWebhookInvalid
	}
	limited := io.LimitReader(body, v.maxBodyBytes+1)
	raw, err := io.ReadAll(limited)
	if err != nil || int64(len(raw)) > v.maxBodyBytes || len(raw) == 0 {
		return VerifiedWebhook{}, ErrWebhookInvalid
	}
	keyRef, ok := v.webhookKey(headers.KeyID, headers.KeyVersion)
	if !ok {
		return VerifiedWebhook{}, ErrWebhookInvalid
	}
	secret, err := v.secrets.ResolveWebhookSecret(ctx, keyRef.SecretRef)
	if err != nil || len(secret) < 32 {
		return VerifiedWebhook{}, ErrWebhookInvalid
	}
	canonical := make([]byte, 0, len(headers.Timestamp)+1+len(raw))
	canonical = append(canonical, headers.Timestamp...)
	canonical = append(canonical, '.')
	canonical = append(canonical, raw...)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(canonical)
	expected := mac.Sum(nil)
	for i := range secret {
		secret[i] = 0
	}
	provided, err := decodeWebhookSignature(headers.Signature)
	if err != nil || !hmac.Equal(expected, provided) {
		return VerifiedWebhook{}, ErrWebhookInvalid
	}
	var event WebhookEvent
	if err := decodeStrictJSON(raw, &event); err != nil {
		return VerifiedWebhook{}, ErrWebhookInvalid
	}
	if !profileTokenPattern.MatchString(event.EventID) || strings.TrimSpace(event.EventType) == "" ||
		event.Provider != v.profile.Provider || event.APIVersion != v.profile.Webhook.Version ||
		!profileTokenPattern.MatchString(event.PayoutID) || !profileTokenPattern.MatchString(event.QuoteID) ||
		!profileTokenPattern.MatchString(event.CorrelationID) || !validStatus(event.Status) || event.OccurredAt.IsZero() ||
		event.OccurredAt.Before(now.Add(-v.maxClockSkew)) || event.OccurredAt.After(now.Add(v.maxClockSkew)) {
		return VerifiedWebhook{}, ErrWebhookInvalid
	}
	binding, err := v.bindings.LookupWebhookBinding(ctx, event.Provider, event.PayoutID)
	if err != nil || binding.Provider != event.Provider || binding.PayoutID != event.PayoutID ||
		binding.QuoteID != event.QuoteID || binding.CorrelationID != event.CorrelationID {
		return VerifiedWebhook{}, ErrWebhookInvalid
	}
	digest := sha256.Sum256(raw)
	replayResult, err := v.replay.RecordWebhookEvent(ctx, event.Provider, event.EventID, hex.EncodeToString(digest[:]), now)
	if err != nil {
		return VerifiedWebhook{}, fmt.Errorf("%w: replay repository unavailable", ErrAdapterUnavailable)
	}
	switch replayResult {
	case WebhookReplayNew:
		return VerifiedWebhook{Event: event, KeyID: headers.KeyID, KeyVersion: headers.KeyVersion}, nil
	case WebhookReplayExact:
		return VerifiedWebhook{Event: event, Duplicate: true, KeyID: headers.KeyID, KeyVersion: headers.KeyVersion}, nil
	case WebhookReplayConflicting:
		return VerifiedWebhook{}, ErrWebhookReplay
	default:
		return VerifiedWebhook{}, fmt.Errorf("%w: invalid replay repository result", ErrAdapterUnavailable)
	}
}

func (v *WebhookVerifier) webhookKey(keyID string, version string) (WebhookKeyReference, bool) {
	for _, key := range v.profile.Webhook.Keys {
		if key.KeyID == keyID && key.Version == version {
			return key, true
		}
	}
	return WebhookKeyReference{}, false
}

func decodeWebhookSignature(raw string) ([]byte, error) {
	raw = strings.TrimPrefix(raw, "sha256=")
	if len(raw) != sha256.Size*2 || strings.ToLower(raw) != raw {
		return nil, errors.New("invalid signature encoding")
	}
	return hex.DecodeString(raw)
}

// MemoryWebhookReplayRepository is a concurrency-safe test and development
// repository. Production callers must provide durable storage.
type MemoryWebhookReplayRepository struct {
	mu     sync.Mutex
	events map[string]string
}

// NewMemoryWebhookReplayRepository creates an empty replay repository.
func NewMemoryWebhookReplayRepository() *MemoryWebhookReplayRepository {
	return &MemoryWebhookReplayRepository{events: make(map[string]string)}
}

func (r *MemoryWebhookReplayRepository) RecordWebhookEvent(_ context.Context, provider string, eventID string, payloadDigest string, _ time.Time) (WebhookReplayResult, error) {
	if provider == "" || eventID == "" || payloadDigest == "" {
		return "", ErrInvalidRequest
	}
	key := provider + "|" + eventID
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.events[key]
	if !ok {
		r.events[key] = payloadDigest
		return WebhookReplayNew, nil
	}
	if hmac.Equal([]byte(existing), []byte(payloadDigest)) {
		return WebhookReplayExact, nil
	}
	return WebhookReplayConflicting, nil
}

func (*MemoryWebhookReplayRepository) Durable() bool { return false }

// SignWebhookForTesting signs canonical webhook bytes for deterministic tests.
// It is deliberately unexported from production protocol flow.
func signWebhookForTesting(timestamp string, body []byte, secret []byte) string {
	canonical := bytes.Join([][]byte{[]byte(timestamp), body}, []byte("."))
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(canonical)
	return hex.EncodeToString(mac.Sum(nil))
}

var _ WebhookReplayRepository = (*MemoryWebhookReplayRepository)(nil)
