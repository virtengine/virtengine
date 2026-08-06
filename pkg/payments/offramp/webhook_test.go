package offramp

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type staticBindingRepository struct {
	binding WebhookBinding
}

func (r staticBindingRepository) LookupWebhookBinding(_ context.Context, _ string, _ string) (WebhookBinding, error) {
	return r.binding, nil
}

func webhookFixture(t *testing.T) (*WebhookVerifier, *testClock, *staticSecretResolver, WebhookEvent) {
	t.Helper()
	clock := newTestClock(time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC))
	profile := blockedSandboxProfile()
	secrets := &staticSecretResolver{webhook: map[string][]byte{
		"vault://offramp/sandbox/webhook/1": bytes.Repeat([]byte{0x11}, 32),
		"vault://offramp/sandbox/webhook/2": bytes.Repeat([]byte{0x22}, 32),
	}}
	event := WebhookEvent{
		EventID: "event-1", EventType: "payout.status", Provider: profile.Provider,
		APIVersion: profile.Webhook.Version, PayoutID: "payout-1", QuoteID: "quote-1",
		CorrelationID: "correlation-1", Status: StatusCompleted, OccurredAt: clock.Now(),
	}
	verifier, err := NewWebhookVerifier(WebhookVerifierConfig{
		Profile: profile, Secrets: secrets, Replay: NewMemoryWebhookReplayRepository(),
		Bindings: staticBindingRepository{binding: WebhookBinding{
			Provider: event.Provider, PayoutID: event.PayoutID, QuoteID: event.QuoteID, CorrelationID: event.CorrelationID,
		}},
		Clock: clock.Now,
	})
	require.NoError(t, err)
	return verifier, clock, secrets, event
}

func signedWebhook(t *testing.T, clock *testClock, event WebhookEvent, secret []byte, keyVersion string) (WebhookHeaders, []byte) {
	t.Helper()
	body, err := json.Marshal(event)
	require.NoError(t, err)
	timestamp := strconv.FormatInt(clock.Now().Unix(), 10)
	return WebhookHeaders{
		Timestamp: timestamp, Signature: signWebhookForTesting(timestamp, body, secret),
		KeyID: "partner-key", KeyVersion: keyVersion, APIVersion: event.APIVersion,
	}, body
}

func TestWebhookVerifierAcceptsRotatedKeysAndExactDuplicates(t *testing.T) {
	t.Parallel()
	verifier, clock, secrets, event := webhookFixture(t)
	for _, version := range []string{"1", "2"} {
		event.EventID = "event-" + version
		secret := secrets.webhook["vault://offramp/sandbox/webhook/"+version]
		headers, body := signedWebhook(t, clock, event, secret, version)
		verified, err := verifier.Verify(context.Background(), headers, bytes.NewReader(body))
		require.NoError(t, err)
		require.False(t, verified.Duplicate)
		require.Equal(t, version, verified.KeyVersion)
		duplicate, err := verifier.Verify(context.Background(), headers, bytes.NewReader(body))
		require.NoError(t, err)
		require.True(t, duplicate.Duplicate)
	}
}

func TestWebhookVerifierRejectsInvalidSignatureTimestampKeyAndVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*WebhookHeaders, *testClock, *[]byte)
	}{
		{"signature", func(headers *WebhookHeaders, _ *testClock, _ *[]byte) { headers.Signature = strings.Repeat("0", 64) }},
		{"timestamp", func(headers *WebhookHeaders, _ *testClock, _ *[]byte) { headers.Timestamp = "not-a-timestamp" }},
		{"expired", func(headers *WebhookHeaders, clock *testClock, body *[]byte) {
			clock.Advance(10 * time.Minute)
			headers.Timestamp = strconv.FormatInt(clock.Now().Add(-10*time.Minute).Unix(), 10)
			headers.Signature = signWebhookForTesting(headers.Timestamp, *body, bytes.Repeat([]byte{0x11}, 32))
		}},
		{"key", func(headers *WebhookHeaders, _ *testClock, _ *[]byte) { headers.KeyID = "unknown" }},
		{"key version", func(headers *WebhookHeaders, _ *testClock, _ *[]byte) { headers.KeyVersion = "99" }},
		{"webhook version", func(headers *WebhookHeaders, _ *testClock, _ *[]byte) { headers.APIVersion = "wrong-version" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			verifier, clock, secrets, event := webhookFixture(t)
			headers, body := signedWebhook(t, clock, event, secrets.webhook["vault://offramp/sandbox/webhook/1"], "1")
			tc.edit(&headers, clock, &body)
			_, err := verifier.Verify(context.Background(), headers, bytes.NewReader(body))
			require.ErrorIs(t, err, ErrWebhookInvalid)
		})
	}
}

func TestWebhookVerifierRejectsStaleOccurredAt(t *testing.T) {
	verifier, clock, secrets, event := webhookFixture(t)
	event.OccurredAt = clock.Now().Add(-10 * time.Minute)
	headers, body := signedWebhook(t, clock, event, secrets.webhook["vault://offramp/sandbox/webhook/1"], "1")
	_, err := verifier.Verify(context.Background(), headers, bytes.NewReader(body))
	require.ErrorIs(t, err, ErrWebhookInvalid)
}

func TestWebhookVerifierRejectsConflictingDuplicate(t *testing.T) {
	t.Parallel()
	verifier, clock, secrets, event := webhookFixture(t)
	secret := secrets.webhook["vault://offramp/sandbox/webhook/1"]
	headers, body := signedWebhook(t, clock, event, secret, "1")
	_, err := verifier.Verify(context.Background(), headers, bytes.NewReader(body))
	require.NoError(t, err)
	event.Status = StatusFailed
	headers, body = signedWebhook(t, clock, event, secret, "1")
	_, err = verifier.Verify(context.Background(), headers, bytes.NewReader(body))
	require.ErrorIs(t, err, ErrWebhookReplay)
}

func TestWebhookVerifierRejectsUnknownFieldsBindingAndOversize(t *testing.T) {
	t.Parallel()
	verifier, clock, secrets, event := webhookFixture(t)
	secret := secrets.webhook["vault://offramp/sandbox/webhook/1"]

	validBody, err := json.Marshal(event)
	require.NoError(t, err)
	unknownBody := bytes.TrimSuffix(validBody, []byte("}"))
	unknownBody = append(unknownBody, []byte(`,"beneficiary":"raw-bank-data"}`)...)
	timestamp := strconv.FormatInt(clock.Now().Unix(), 10)
	headers := WebhookHeaders{
		Timestamp: timestamp, Signature: signWebhookForTesting(timestamp, unknownBody, secret),
		KeyID: "partner-key", KeyVersion: "1", APIVersion: event.APIVersion,
	}
	_, err = verifier.Verify(context.Background(), headers, bytes.NewReader(unknownBody))
	require.ErrorIs(t, err, ErrWebhookInvalid)

	event.CorrelationID = "wrong-correlation"
	headers, validBody = signedWebhook(t, clock, event, secret, "1")
	_, err = verifier.Verify(context.Background(), headers, bytes.NewReader(validBody))
	require.ErrorIs(t, err, ErrWebhookInvalid)

	verifier.maxBodyBytes = 16
	event.CorrelationID = "correlation-1"
	headers, validBody = signedWebhook(t, clock, event, secret, "1")
	_, err = verifier.Verify(context.Background(), headers, bytes.NewReader(validBody))
	require.ErrorIs(t, err, ErrWebhookInvalid)
}

func TestWebhookVerifierFailsClosedWhenReplayRepositoryFails(t *testing.T) {
	t.Parallel()
	verifier, clock, secrets, event := webhookFixture(t)
	verifier.replay = failingReplayRepository{}
	headers, body := signedWebhook(t, clock, event, secrets.webhook["vault://offramp/sandbox/webhook/1"], "1")
	_, err := verifier.Verify(context.Background(), headers, bytes.NewReader(body))
	require.ErrorIs(t, err, ErrAdapterUnavailable)
}

func TestProductionWebhookVerifierRejectsInMemoryReplayRepository(t *testing.T) {
	t.Parallel()
	profile := certifiedProductionProfile()
	_, err := NewWebhookVerifier(WebhookVerifierConfig{
		Profile: profile,
		Secrets: &staticSecretResolver{webhook: map[string][]byte{
			"vault://offramp/production/webhook/1": bytes.Repeat([]byte{0x11}, 32),
			"vault://offramp/production/webhook/2": bytes.Repeat([]byte{0x22}, 32),
		}},
		Replay:   NewMemoryWebhookReplayRepository(),
		Bindings: staticBindingRepository{},
	})
	require.ErrorIs(t, err, ErrProfileNotExecutable)
}

type failingReplayRepository struct{}

func (failingReplayRepository) RecordWebhookEvent(context.Context, string, string, string, time.Time) (WebhookReplayResult, error) {
	return "", context.DeadlineExceeded
}
