package email

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hostileErrorMessage is what a downstream verification service (or a wrapped
// request error) can produce. Interpolated into a hand-built JSON string literal
// it closes the "message" string and injects arbitrary fields.
const hostileErrorMessage = `bad "secret"</script>,\n"admin":true,"message":"x`

// stubVerificationService returns a fixed challenge and then fails verification
// with an attacker-influenced message, so the HTTP handler's JSON error branch is
// reached with hostile data.
type stubVerificationService struct {
	challenge *EmailChallenge
	verifyErr error
	message   string
}

func (s *stubVerificationService) InitiateVerification(context.Context, *InitiateRequest) (*InitiateResponse, error) {
	return nil, nil
}

func (s *stubVerificationService) VerifyChallenge(context.Context, *VerifyRequest) (*VerifyResponse, error) {
	return &VerifyResponse{ErrorCode: `BAD"CODE`, ErrorMessage: s.message}, s.verifyErr
}

func (s *stubVerificationService) ResendVerification(context.Context, *ResendRequest) (*ResendResponse, error) {
	return nil, nil
}

func (s *stubVerificationService) GetChallenge(context.Context, string) (*EmailChallenge, error) {
	return s.challenge, nil
}

func (s *stubVerificationService) CancelChallenge(context.Context, string, string) error { return nil }

func (s *stubVerificationService) ProcessWebhook(context.Context, string, []byte) error { return nil }

func (s *stubVerificationService) GetDeliveryStatus(context.Context, string) (*DeliveryResult, error) {
	return nil, nil
}

func (s *stubVerificationService) HealthCheck(context.Context) (*HealthStatus, error) {
	return &HealthStatus{}, nil
}

func (s *stubVerificationService) Close() error { return nil }

func newTestLinkService(t *testing.T, stub *stubVerificationService) *LinkVerificationService {
	t.Helper()
	svc, err := NewLinkVerificationService(
		LinkVerificationConfig{
			BaseURL:    "https://example.test/verify",
			VerifyPath: "/verify",
		},
		stub,
		nil,
		zerolog.Nop(),
	)
	require.NoError(t, err)
	return svc
}

// TestHTTPHandlerDoesNotInjectJSONFromErrorMessage drives a hostile downstream
// error message through the real HTTP handler and asserts the response body is
// valid JSON containing exactly the intended fields -- no injected "admin".
func TestHTTPHandlerDoesNotInjectJSONFromErrorMessage(t *testing.T) {
	stub := &stubVerificationService{
		challenge: &EmailChallenge{ChallengeID: "ch-1", AccountAddress: "ve1abc"},
		verifyErr: ErrInvalidRequest,
		message:   hostileErrorMessage,
	}
	svc := newTestLinkService(t, stub)

	req := httptest.NewRequest(http.MethodGet, "/verify?token=t&challenge=ch-1", nil)
	rec := httptest.NewRecorder()
	svc.HTTPHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body),
		"response body must be valid JSON, got %q", rec.Body.String())

	assert.NotContains(t, body, "admin", "a hostile error message injected a JSON field")
	assert.Equal(t, false, body["success"])
	assert.Equal(t, `BAD"CODE`, body["error"])
	assert.Equal(t, hostileErrorMessage, body["message"],
		"the hostile message should survive as data, not as structure")
	assert.Len(t, body, 3)
}

// TestHTTPHandlerSuccessBodyIsValidJSON pins the success branch shape too.
// The stub always fails verification, and the success branch is only reachable
// with a live downstream service, so this asserts the shared encoder the success
// branch uses with an attestation ID that needs escaping.
func TestHTTPHandlerSuccessBodyIsValidJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSONResponse(rec, struct {
		Success       bool   `json:"success"`
		Verified      bool   `json:"verified"`
		AttestationID string `json:"attestation_id"`
	}{true, true, `at"tack`})

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, true, body["success"])
	assert.Equal(t, true, body["verified"])
	assert.Equal(t, `at"tack`, body["attestation_id"])
	assert.Len(t, body, 3)
}

// TestWriteJSONResponseEscapesControlCharacters covers the raw newline case that
// made the old hand-built literal produce a body no JSON parser accepts.
func TestWriteJSONResponseEscapesControlCharacters(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSONResponse(rec, struct {
		Message string `json:"message"`
	}{"line1\nline2\ttab \\backslash"})

	raw := rec.Body.String()
	assert.False(t, strings.ContainsAny(raw, "\n\t"), "control characters must be escaped: %q", raw)
	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &body))
	assert.Equal(t, "line1\nline2\ttab \\backslash", body["message"])
}
