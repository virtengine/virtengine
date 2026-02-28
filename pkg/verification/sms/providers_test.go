package sms

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTwilioProviderWrapsRealGateway(t *testing.T) {
	t.Parallel()

	provider, err := NewTwilioProvider(ProviderConfig{
		AccountSID: "AC123",
		AuthToken:  "secret",
		FromNumber: "+14155551234",
	}, zerolog.Nop())
	require.NoError(t, err)

	require.NotNil(t, provider.gateway)
	_, ok := provider.gateway.(*TwilioGateway)
	assert.True(t, ok)
}

func TestCreateSMSProviderUsesConfiguredProviderType(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	config.PrimaryProvider = "primary"
	config.ProviderConfigs = map[string]ProviderConfig{
		"primary": {
			Type:       providerTwilio,
			AccountSID: "AC123",
			AuthToken:  "secret",
			FromNumber: "+14155551234",
		},
	}

	provider, err := createSMSProvider(config, zerolog.Nop())
	require.NoError(t, err)

	_, ok := provider.(*TwilioProvider)
	assert.True(t, ok)
}

func TestCreateSMSProviderRequiresExplicitPrimaryProviderConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	config.ProviderConfigs = nil

	provider, err := createSMSProvider(config, zerolog.Nop())
	require.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), `primary_provider "twilio" requires a matching provider_configs entry`)
}

func TestSNSProviderSendUsesPublishAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "Action=Publish")
		require.Contains(t, string(body), "PhoneNumber=%2B14155551234")
		require.Contains(t, string(body), "Message=Code+123456")

		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(`<PublishResponse><PublishResult><MessageId>sns-message-1</MessageId></PublishResult></PublishResponse>`))
	}))
	defer server.Close()

	provider, err := NewSNSProvider(ProviderConfig{
		Region:    "ap-southeast-2",
		APIKey:    "AKID",
		APISecret: "SECRET",
		SenderID:  "VirtEngine",
	}, zerolog.Nop())
	require.NoError(t, err)

	provider.endpoint = server.URL
	provider.httpClient = server.Client()

	result, err := provider.Send(context.Background(), &SMSMessage{
		To:   "+14155551234",
		Body: "Code 123456",
	})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "sns-message-1", result.MessageID)
	assert.Equal(t, providerSNS, result.Provider)
}
