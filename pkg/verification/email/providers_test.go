package email

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSESProviderSendUsesAWSQueryAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "Action=SendEmail")
		require.Contains(t, string(body), "Destination.ToAddresses.member.1=user%40example.com")
		require.Contains(t, string(body), "Source=noreply%40virtengine.com")

		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(`<SendEmailResponse><SendEmailResult><MessageId>ses-message-1</MessageId></SendEmailResult></SendEmailResponse>`))
	}))
	defer server.Close()

	provider, err := NewSESProvider(SESConfig{
		Region:          "ap-southeast-2",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
	}, newTestLogger())
	require.NoError(t, err)

	provider.endpoint = server.URL
	provider.httpClient = server.Client()

	result, err := provider.Send(context.Background(), &Email{
		To:       "user@example.com",
		From:     "noreply@virtengine.com",
		Subject:  "Verify",
		TextBody: "Your code is 123456",
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "ses-message-1", result.MessageID)
	assert.Equal(t, "ses", result.Provider)
}

func TestSendGridProviderSendUsesHTTPAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "Bearer sg-key", r.Header.Get("Authorization"))
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		payload := string(body)
		require.Contains(t, payload, `"email":"user@example.com"`)
		require.Contains(t, payload, `"email":"noreply@virtengine.com"`)
		require.Contains(t, payload, `"subject":"Verify"`)

		w.Header().Set("X-Message-Id", "sg-message-1")
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	provider, err := NewSendGridProvider(SendGridConfig{
		APIKey:     "sg-key",
		FromDomain: "virtengine.com",
	}, newTestLogger())
	require.NoError(t, err)

	provider.baseURL = server.URL
	provider.httpClient = server.Client()

	result, err := provider.Send(context.Background(), &Email{
		To:       "user@example.com",
		From:     "noreply@virtengine.com",
		Subject:  "Verify",
		TextBody: strings.Repeat("A", 16),
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "sg-message-1", result.MessageID)
	assert.Equal(t, "sendgrid", result.Provider)
}
